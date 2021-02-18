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

const RouteMatrixV1 = 1

type RouteMatrix struct {
	Version            uint32
	CreatedAt          uint64
	RelayIDsToIndices  map[uint64]int32
	RelayIDs           []uint64
	RelayAddresses     []net.UDPAddr // external IPs only
	RelayNames         []string
	RelayLatitudes     []float32
	RelayLongitudes    []float32
	RelayDatacenterIDs []uint64
	RouteEntries       []core.RouteEntry

	cachedResponse       []byte
	cachedResponseBinary []byte
	cachedResponseMutex  sync.RWMutex

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

func (m *RouteMatrix) decodeBinaryV1(index int, data []byte) error {
	if !encoding.ReadUint64(data, &index, &m.CreatedAt) {
		return fmt.Errorf("unable to decode created at")
	}

	var numRelays uint32
	if !encoding.ReadUint32(data, &index, &numRelays) {
		return fmt.Errorf("unable to decode number of relays")
	}

	m.RelayIDsToIndices = make(map[uint64]int32)
	m.RelayIDs = make([]uint64, numRelays)
	m.RelayAddresses = make([]net.UDPAddr, numRelays)
	m.RelayNames = make([]string, numRelays)
	m.RelayLatitudes = make([]float32, numRelays)
	m.RelayLongitudes = make([]float32, numRelays)
	m.RelayDatacenterIDs = make([]uint64, numRelays)

	fmt.Println(numRelays)
	for i := uint32(0); i < numRelays; i++ {
		if !encoding.ReadUint64(data, &index, &m.RelayIDs[i]) {
			return fmt.Errorf("unable to decode relay ID")
		}

		var addrSize uint8
		if !encoding.ReadUint8(data, &index, &addrSize) {
			return fmt.Errorf("unable to decode relay address size")
		}
		addrBytes := make([]byte, addrSize)
		if !encoding.ReadBytes(data, &index, &addrBytes, uint32(addrSize)) {
			return fmt.Errorf("unable to retrieve relay address bytes from data")
		}
		m.RelayAddresses[i] = *encoding.ReadAddress(addrBytes)

		if !encoding.ReadString(data, &index, &m.RelayNames[i], MaxRelayNameLength) {
			return fmt.Errorf("unable to decode relay name")
		}

		if !encoding.ReadFloat32(data, &index, &m.RelayLatitudes[i]) {
			return fmt.Errorf("unable to decode relay latitude")
		}

		if !encoding.ReadFloat32(data, &index, &m.RelayLongitudes[i]) {
			return fmt.Errorf("unable to decode relay longitude")
		}

		if !encoding.ReadUint64(data, &index, &m.RelayDatacenterIDs[i]) {
			return fmt.Errorf("unable to decode relay datacenter ID")
		}
		m.RelayIDsToIndices[m.RelayIDs[i]] = int32(i)
	}

	var numEntries uint32
	if !encoding.ReadUint32(data, &index, &numEntries) {
		return fmt.Errorf("unable to decode number of route entries")
	}
	m.RouteEntries = make([]core.RouteEntry, numEntries)

	for i := uint32(0); i < numEntries; i++ {
		entry := &m.RouteEntries[i]

		if !encoding.ReadInt32(data, &index, &entry.DirectCost) {
			return fmt.Errorf("unable to decode route direct cost")
		}
		if entry.DirectCost < -1 || entry.DirectCost > InvalidRouteValue {
			return fmt.Errorf("entry direct cost is invalid")
		}

		if !encoding.ReadInt32(data, &index, &entry.NumRoutes) {
			return fmt.Errorf("unable to decode route direct cost")
		}

		if entry.NumRoutes < 0 {
			return fmt.Errorf("entry numRoutes is invalid")
		}

		for i := 0; i < core.MaxRoutesPerEntry; i++ {
			if !encoding.ReadInt32(data, &index, &entry.RouteCost[i]) {
				return fmt.Errorf("unable to decode route cost")
			}
			if entry.RouteCost[i] < -1 || entry.RouteCost[i] > InvalidRouteValue {
				return fmt.Errorf("entry route cost is invalid")
			}

			if !encoding.ReadInt32(data, &index, &entry.RouteNumRelays[i]) {
				return fmt.Errorf("unable to decode ")
			}
			if entry.RouteNumRelays[i] < 0 || entry.RouteNumRelays[i] > int32(core.MaxRelaysPerRoute) {
				return fmt.Errorf("entry routeNumRoutes is invalid")
			}

			if !encoding.ReadUint32(data, &index, &entry.RouteHash[i]) {
				return fmt.Errorf("unable to decode route direct cost")
			}

			for j := int32(0); j < entry.RouteNumRelays[i]; j++ {
				encoding.ReadInt32(data, &index, &entry.RouteRelays[i][j])
			}
		}
	}

	return nil
}

func (m *RouteMatrix) encodeBinaryV1(data []byte) error {
	index := 0
	encoding.WriteUint32(data, &index, RouteMatrixV1)
	encoding.WriteUint64(data, &index, m.CreatedAt)

	numRelays := uint32(len(m.RelayIDs))
	encoding.WriteUint32(data, &index, numRelays)

	for i := uint32(0); i < numRelays; i++ {
		encoding.WriteUint64(data, &index, m.RelayIDs[i])
		addrBytes := make([]byte, 50)
		encoding.WriteAddress(addrBytes, &m.RelayAddresses[i])
		encoding.WriteUint8(data, &index, uint8(len(addrBytes)))
		encoding.WriteBytes(data, &index, addrBytes, len(addrBytes))
		encoding.WriteString(data, &index, m.RelayNames[i], MaxRelayNameLength)
		encoding.WriteFloat32(data, &index, m.RelayLatitudes[i])
		encoding.WriteFloat32(data, &index, m.RelayLongitudes[i])
		encoding.WriteUint64(data, &index, m.RelayDatacenterIDs[i])
	}

	numEntries := uint32(len(m.RouteEntries))
	encoding.WriteUint32(data, &index, numEntries)

	for i := uint32(0); i < numEntries; i++ {
		entry := m.RouteEntries[i]

		if entry.DirectCost < -1 || entry.DirectCost > InvalidRouteValue {
			return fmt.Errorf("entry direct cost is invalid")
		}
		encoding.WriteInt32(data, &index, entry.DirectCost)

		if entry.NumRoutes < 0 {
			return fmt.Errorf("entry numRoutes is invalid")
		}
		encoding.WriteInt32(data, &index, entry.NumRoutes)

		for i := 0; i < core.MaxRoutesPerEntry; i++ {
			if entry.RouteCost[i] < -1 || entry.RouteCost[i] > InvalidRouteValue {
				return fmt.Errorf("entry route cost is invalid")
			}
			encoding.WriteInt32(data, &index, entry.RouteCost[i])

			if entry.RouteNumRelays[i] < 0 || entry.RouteNumRelays[i] > int32(core.MaxRelaysPerRoute) {
				return fmt.Errorf("entry routeNumRoutes is invalid")
			}
			encoding.WriteInt32(data, &index, entry.RouteNumRelays[i])
			encoding.WriteUint32(data, &index, entry.RouteHash[i])

			for j := int32(0); j < entry.RouteNumRelays[i]; j++ {
				encoding.WriteInt32(data, &index, entry.RouteRelays[i][j])
			}
		}
	}

	return nil
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

func (m *RouteMatrix) ReadFromBinary(data []byte) error {
	index := 0
	encoding.ReadUint32(data, &index, &m.Version)

	switch m.Version {
	case RouteMatrixV1:
		return m.decodeBinaryV1(index, data)
	default:
		return fmt.Errorf("unsuported Route Matrix version: %v", m.Version)
	}
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

func (m *RouteMatrix) WriteToBinary(data []byte) error {
	switch m.Version {
	case RouteMatrixV1:
		return m.encodeBinaryV1(data)
	default:
		return fmt.Errorf("unsuported Route Matrix version: %v", m.Version)
	}
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

func (m *RouteMatrix) GetResponseDataBinary() []byte {
	m.cachedResponseMutex.RLock()
	response := m.cachedResponseBinary
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

func (m *RouteMatrix) WriteResponseDataBinary(bufferSize int) error {
	data := make([]byte, bufferSize)
	if err := m.WriteToBinary(data); err != nil {
		return fmt.Errorf("failed to encode route matrix in WriteResponseData(): %v", err)
	}

	m.cachedResponseMutex.Lock()
	m.cachedResponseBinary = data
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
