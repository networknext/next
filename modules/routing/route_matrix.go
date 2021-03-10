package routing

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"sort"
	"sync"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
)

type RouteMatrix struct {
	RelayIDsToIndices  map[uint64]int32
	RelayIDs           []uint64
	RelayAddresses     []net.UDPAddr // external IPs only
	RelayNames         []string
	RelayLatitudes     []float32
	RelayLongitudes    []float32
	RelayDatacenterIDs []uint64
	RouteEntries       []core.RouteEntry

	cachedResponse      []byte
	cachedResponseMutex sync.RWMutex

	cachedAnalysis      []byte
	cachedAnalysisMutex sync.RWMutex
}

func (m *RouteMatrix) Serialize(stream encoding.Stream) error {
	numRelays := uint32(len(m.RelayIDs))
	stream.SerializeUint32(&numRelays)

	if stream.IsReading() {
		m.RelayIDsToIndices = make(map[uint64]int32)
		m.RelayIDs = make([]uint64, numRelays)
		m.RelayAddresses = make([]net.UDPAddr, numRelays)
		m.RelayNames = make([]string, numRelays)
		m.RelayLatitudes = make([]float32, numRelays)
		m.RelayLongitudes = make([]float32, numRelays)
		m.RelayDatacenterIDs = make([]uint64, numRelays)
	}

	for i := uint32(0); i < numRelays; i++ {
		stream.SerializeUint64(&m.RelayIDs[i])
		stream.SerializeAddress(&m.RelayAddresses[i])
		stream.SerializeString(&m.RelayNames[i], MaxRelayNameLength)
		stream.SerializeFloat32(&m.RelayLatitudes[i])
		stream.SerializeFloat32(&m.RelayLongitudes[i])
		stream.SerializeUint64(&m.RelayDatacenterIDs[i])

		if stream.IsReading() {
			m.RelayIDsToIndices[m.RelayIDs[i]] = int32(i)
		}
	}

	numEntries := uint32(len(m.RouteEntries))
	stream.SerializeUint32(&numEntries)

	if stream.IsReading() {
		m.RouteEntries = make([]core.RouteEntry, numEntries)
	}

	for i := uint32(0); i < numEntries; i++ {
		entry := &m.RouteEntries[i]

		stream.SerializeInteger(&entry.DirectCost, -1, InvalidRouteValue)
		stream.SerializeInteger(&entry.NumRoutes, 0, math.MaxInt32)

		for i := 0; i < core.MaxRoutesPerEntry; i++ {
			stream.SerializeInteger(&entry.RouteCost[i], -1, InvalidRouteValue)
			stream.SerializeInteger(&entry.RouteNumRelays[i], 0, core.MaxRelaysPerRoute)
			stream.SerializeUint32(&entry.RouteHash[i])

			for j := 0; j < core.MaxRelaysPerRoute; j++ {
				stream.SerializeInteger(&entry.RouteRelays[i][j], 0, math.MaxInt32)
			}
		}
	}

	return stream.Error()
}

func (m *RouteMatrix) GetNearRelays(directLatency float32, source_latitude float32, source_longitude float32, dest_latitude float32, dest_longitude float32, maxNearRelays int) []uint64 {

	// Quantize to integer values so we don't have noise in low bits

	sourceLatitude := float64(int64(source_latitude))
	sourceLongitude := float64(int64(source_longitude))

	destLatitude := float64(int64(dest_latitude))
	destLongitude := float64(int64(dest_longitude))

	// If direct latency is 0, we don't know it yet. Approximate it via speed of light * 2

	if directLatency <= 0.0 {
		directDistanceKilometers := core.HaversineDistance(sourceLatitude, sourceLongitude, destLatitude, destLongitude)
		directLatency = float32(directDistanceKilometers/299792.458*1000.0) * 2
	}

	// Work with the near relays as an array of structs first for easier sorting

	type NearRelayData struct {
		ID        uint64
		Addr      net.UDPAddr
		Name      string
		Distance  int
		Latitude  float64
		Longitude float64
	}

	nearRelayData := make([]NearRelayData, len(m.RelayIDs))

	for i, relayID := range m.RelayIDs {
		nearRelayData[i].ID = relayID
		nearRelayData[i].Addr = m.RelayAddresses[i]
		nearRelayData[i].Name = m.RelayNames[i]
		nearRelayData[i].Latitude = float64(int64(m.RelayLatitudes[i]))
		nearRelayData[i].Longitude = float64(int64(m.RelayLongitudes[i]))
		nearRelayData[i].Distance = int(core.HaversineDistance(sourceLatitude, sourceLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude))
	}

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].Distance < nearRelayData[j].Distance })

	// Exclude any near relays whose 2/3rds speed of light * 2/3rds route through them is greater than direct + threshold
	// or who are further than x kilometers away from the player's location

	distanceThreshold := 2500

	latencyThreshold := float32(30.0)

	nearRelayIDs := make([]uint64, 0, maxNearRelays)
	nearRelayIDMap := map[uint64]struct{}{}

	for i := 0; i < len(nearRelayData); i++ {
		if len(nearRelayIDs) == maxNearRelays {
			break
		}

		if nearRelayData[i].Distance > distanceThreshold {
			break
		}

		nearRelayLatency := 3.0 / 2.0 * float32(core.SpeedOfLightTimeMilliseconds(sourceLatitude, sourceLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude, destLatitude, destLongitude))
		if nearRelayLatency > directLatency+latencyThreshold {
			continue
		}

		nearRelayIDs = append(nearRelayIDs, nearRelayData[i].ID)

		// Store the near relay ID in a map so that we don't reinsert it later
		nearRelayIDMap[nearRelayData[i].ID] = struct{}{}
	}

	// If we already have enough relays, stop and return them

	if len(nearRelayIDs) == maxNearRelays {
		return nearRelayIDs
	}

	// We need more relays. Look for near relays around the *destination*
	// Paradoxically, this can really help, especially for cases like South America <-> Miami

	for i := range m.RelayIDs {
		nearRelayData[i].Distance = int(core.HaversineDistance(destLatitude, destLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude))
	}

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].Distance < nearRelayData[j].Distance })

	for i := 0; i < len(nearRelayData); i++ {
		if len(nearRelayIDs) == maxNearRelays {
			break
		}

		// Don't add the relay if we've already added it
		if _, ok := nearRelayIDMap[nearRelayData[i].ID]; ok {
			continue
		}

		nearRelayLatency := 3.0 / 2.0 * float32(core.SpeedOfLightTimeMilliseconds(sourceLatitude, sourceLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude, destLatitude, destLongitude))
		if nearRelayLatency > directLatency+latencyThreshold {
			continue
		}

		nearRelayIDs = append(nearRelayIDs, nearRelayData[i].ID)
	}

	return nearRelayIDs
}

func (m *RouteMatrix) GetDatacenterRelayIDs(datacenterID uint64) []uint64 {
	relayIDs := make([]uint64, 0)
	fmt.Println("DC ID:")
	fmt.Println(fmt.Sprintf("%016x\n", datacenterID))
	fmt.Println("------------------------------")
	fmt.Println("Route Matrix Relays:")
	for i := 0; i < len(m.RelayDatacenterIDs); i++ {
		fmt.Println(m.RelayNames[i])
		fmt.Println(fmt.Sprintf("%016x\n", m.RelayDatacenterIDs[i]))
		if m.RelayDatacenterIDs[i] == datacenterID {
			relayIDs = append(relayIDs, m.RelayIDs[i])
		}
	}
	fmt.Println("------------------------------\n")

	return relayIDs
}

func (m *RouteMatrix) ReadFrom(reader io.Reader) (int64, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return 0, err
	}

	readStream := encoding.CreateReadStream(data)
	err = m.Serialize(readStream)
	return int64(readStream.GetBytesProcessed()), err
}

func (m *RouteMatrix) WriteTo(writer io.Writer, bufferSize int) (int64, error) {
	writeStream, err := encoding.CreateWriteStream(bufferSize)
	if err != nil {
		return 0, err
	}

	if err = m.Serialize(writeStream); err != nil {
		return int64(writeStream.GetBytesProcessed()), err
	}

	n, err := writer.Write(writeStream.GetData()[:writeStream.GetBytesProcessed()])
	return int64(n), err
}

func (m *RouteMatrix) WriteAnalysisTo(writer io.Writer) {
	src := m.RelayIDs
	dest := m.RelayIDs

	numRelayPairs := 0.0
	numValidRelayPairs := 0.0

	numValidRelayPairsWithoutImprovement := 0.0

	buckets := make([]int, 11)

	for i := range src {
		for j := range dest {
			if j < i {
				numRelayPairs++
				abFlatIndex := TriMatrixIndex(i, j)
				if len(m.RouteEntries[abFlatIndex].RouteCost) > 0 {
					numValidRelayPairs++
					improvement := m.RouteEntries[abFlatIndex].DirectCost - m.RouteEntries[abFlatIndex].RouteCost[0]
					if improvement > 0.0 {
						if improvement <= 5 {
							buckets[0]++
						} else if improvement <= 10 {
							buckets[1]++
						} else if improvement <= 15 {
							buckets[2]++
						} else if improvement <= 20 {
							buckets[3]++
						} else if improvement <= 25 {
							buckets[4]++
						} else if improvement <= 30 {
							buckets[5]++
						} else if improvement <= 35 {
							buckets[6]++
						} else if improvement <= 40 {
							buckets[7]++
						} else if improvement <= 45 {
							buckets[8]++
						} else if improvement <= 50 {
							buckets[9]++
						} else {
							buckets[10]++
						}
					} else {
						numValidRelayPairsWithoutImprovement++
					}
				}
			}
		}
	}

	fmt.Fprintf(writer, "%s Improvement:\n\n", "RTT")
	fmt.Fprintf(writer, "    None: %d (%.2f%%)\n", int(numValidRelayPairsWithoutImprovement), numValidRelayPairsWithoutImprovement/numValidRelayPairs*100.0)

	for i := range buckets {
		if i != len(buckets)-1 {
			fmt.Fprintf(writer, "    %d-%d%s: %d (%.2f%%)\n", i*5, (i+1)*5, "ms", buckets[i], float64(buckets[i])/numValidRelayPairs*100.0)
		} else {
			fmt.Fprintf(writer, "    %d%s+: %d (%.2f%%)\n", i*5, "ms", buckets[i], float64(buckets[i])/numValidRelayPairs*100.0)
		}
	}

	totalRoutes := uint64(0)
	maxRouteLength := int32(0)
	maxRoutesPerRelayPair := int32(0)
	relayPairsWithNoRoutes := 0
	relayPairsWithOneRoute := 0
	totalRouteLength := uint64(0)

	for i := range src {
		for j := range dest {
			if j < i {
				ijFlatIndex := TriMatrixIndex(i, j)
				n := m.RouteEntries[ijFlatIndex].NumRoutes
				if n > maxRoutesPerRelayPair {
					maxRoutesPerRelayPair = n
				}
				totalRoutes += uint64(n)
				if n == 0 {
					relayPairsWithNoRoutes++
				}
				if n == 1 {
					relayPairsWithOneRoute++
				}
				for k := 0; k < int(n); k++ {
					numRelays := m.RouteEntries[ijFlatIndex].RouteNumRelays[k]
					totalRouteLength += uint64(numRelays)
					if numRelays > maxRouteLength {
						maxRouteLength = numRelays
					}
				}
			}
		}
	}

	averageNumRoutes := float64(totalRoutes) / float64(numRelayPairs)
	averageRouteLength := float64(totalRouteLength) / float64(totalRoutes)

	fmt.Fprintf(writer, "\n%s Summary:\n\n", "Route")
	fmt.Fprintf(writer, "    %.1f routes per relay pair on average (%d max)\n", averageNumRoutes, maxRoutesPerRelayPair)
	fmt.Fprintf(writer, "    %.1f relays per route on average (%d max)\n", averageRouteLength, maxRouteLength)
	fmt.Fprintf(writer, "    %.1f%% of relay pairs have only one route\n", float64(relayPairsWithOneRoute)/float64(numRelayPairs)*100)
	fmt.Fprintf(writer, "    %.1f%% of relay pairs have no route\n", float64(relayPairsWithNoRoutes)/float64(numRelayPairs)*100)
}

func (m *RouteMatrix) GetResponseData() []byte {
	m.cachedResponseMutex.RLock()
	response := m.cachedResponse
	m.cachedResponseMutex.RUnlock()
	return response
}

func (m *RouteMatrix) WriteResponseData(bufferSize int) error {
	ws, err := encoding.CreateWriteStream(bufferSize)
	if err != nil {
		return fmt.Errorf("failed to create write stream in route matrix WriteResponseData(): %v", err)
	}

	if err := m.Serialize(ws); err != nil {
		return fmt.Errorf("failed to serialize route matrix in WriteResponseData(): %v", err)
	}

	ws.Flush()

	m.cachedResponseMutex.Lock()
	m.cachedResponse = ws.GetData()[:ws.GetBytesProcessed()]
	m.cachedResponseMutex.Unlock()
	return nil
}

func (m *RouteMatrix) GetAnalysisData() []byte {
	m.cachedAnalysisMutex.RLock()
	analysis := m.cachedAnalysis
	m.cachedAnalysisMutex.RUnlock()
	return analysis
}

func (m *RouteMatrix) WriteAnalysisData() {
	var buffer bytes.Buffer
	m.WriteAnalysisTo(&buffer)

	m.cachedAnalysisMutex.Lock()
	m.cachedAnalysis = buffer.Bytes()
	m.cachedAnalysisMutex.Unlock()
}
