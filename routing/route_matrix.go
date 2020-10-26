package routing

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"sort"
	"sync"

	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/modules/core"
)

type RouteMatrix struct {
	RelayIDsToIndices  map[uint64]int32
	RelayIDs           []uint64
	RelayAddresses     []net.UDPAddr
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

		for i := 0; i < MaxRoutesPerRelayPair; i++ {
			stream.SerializeInteger(&entry.RouteCost[i], -1, InvalidRouteValue)
			stream.SerializeInteger(&entry.RouteNumRelays[i], 0, MaxRelays)
			stream.SerializeUint32(&entry.RouteHash[i])

			for j := 0; j < MaxRelays; j++ {
				stream.SerializeInteger(&entry.RouteRelays[i][j], 0, math.MaxInt32)
			}
		}
	}

	return stream.Error()
}

type NearRelayData struct {
	ID          uint64
	Addr        *net.UDPAddr
	Name        string
	Distance    int
	ClientStats Stats
}

func (m *RouteMatrix) GetNearRelays(latitude float64, longitude float64, maxNearRelays int) ([]NearRelayData, error) {
	nearRelayData := make([]NearRelayData, len(m.RelayIDs))

	// IMPORTANT: Truncate the lat/long values to nearest integer.
	// This fixes numerical instabilities that can happen in the haversine function
	// when two relays are really close together, they can get sorted differently in
	// subsequent passes otherwise.

	lat1 := float64(int64(latitude))
	long1 := float64(int64(longitude))

	for i, relayID := range m.RelayIDs {
		nearRelayData[i].ID = relayID
		nearRelayData[i].Addr = &m.RelayAddresses[i]
		nearRelayData[i].Name = m.RelayNames[i]
		lat2 := float64(m.RelayLatitudes[i])
		long2 := float64(m.RelayLongitudes[i])
		nearRelayData[i].Distance = int(core.HaversineDistance(lat1, long1, lat2, long2))
	}

	// IMPORTANT: Sort near relays by distance using a *stable sort*
	// This is necessary to ensure that relays are always sorted in the same order,
	// even when some relays have the same integer distance from the client. Without this
	// the set of near relays passed down to the SDK can be different from one slice to the next!

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].Distance < nearRelayData[j].Distance })

	if len(nearRelayData) > maxNearRelays {
		nearRelayData = nearRelayData[:maxNearRelays]
	}

	if len(nearRelayData) == 0 {
		return nil, errors.New("no near relays")
	}

	return nearRelayData, nil
}

func (m *RouteMatrix) GetDatacenterRelayIDs(datacenterID uint64) []uint64 {
	relayIDs := make([]uint64, 0)

	for i := 0; i < len(m.RelayDatacenterIDs); i++ {
		if m.RelayDatacenterIDs[i] == datacenterID {
			relayIDs = append(relayIDs, m.RelayIDs[i])
		}
	}

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

	err = m.Serialize(writeStream)
	return int64(writeStream.GetBytesProcessed()), err
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
	averageRouteLength := 0.0

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
				for k := 0; k < int(m.RouteEntries[ijFlatIndex].NumRoutes); k++ {
					numRelays := m.RouteEntries[ijFlatIndex].RouteNumRelays[k]
					averageRouteLength += float64(numRelays)
					if numRelays > maxRouteLength {
						maxRouteLength = numRelays
					}
				}
			}
		}
	}

	averageNumRoutes := float64(totalRoutes) / float64(numRelayPairs)
	averageRouteLength = averageRouteLength / float64(totalRoutes)

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
