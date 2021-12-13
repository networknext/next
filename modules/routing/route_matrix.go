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

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
)

const RouteMatrixSerializeVersion = 5

type RouteMatrix struct {
	RelayIDsToIndices   map[uint64]int32
	RelayIDs            []uint64
	RelayAddresses      []net.UDPAddr // external IPs only
	RelayNames          []string
	RelayLatitudes      []float32
	RelayLongitudes     []float32
	RelayDatacenterIDs  []uint64
	RouteEntries        []core.RouteEntry
	BinFileBytes        int32
	BinFileData         []byte
	CreatedAt           uint64
	Version             uint32
	DestRelays          []bool
	PingStats           []analytics.PingStatsEntry
	RelayStats          []analytics.RelayStatsEntry
	FullRelayIDs        []uint64
	FullRelayIndicesSet map[int32]bool

	cachedResponse      []byte
	cachedResponseMutex sync.RWMutex

	cachedAnalysis      []byte
	cachedAnalysisMutex sync.RWMutex
}

func (m *RouteMatrix) Serialize(stream encoding.Stream) error {

	stream.SerializeUint32(&m.Version)

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

		for i := 0; i < int(entry.NumRoutes); i++ {
			stream.SerializeInteger(&entry.RouteCost[i], -1, InvalidRouteValue)
			stream.SerializeInteger(&entry.RouteNumRelays[i], 0, core.MaxRelaysPerRoute)
			stream.SerializeUint32(&entry.RouteHash[i])
			for j := 0; j < int(entry.RouteNumRelays[i]); j++ {
				stream.SerializeInteger(&entry.RouteRelays[i][j], 0, math.MaxInt32)
			}
		}
	}

	stream.SerializeInteger(&m.BinFileBytes, 0, MaxDatabaseBinWrapperSize)
	if m.BinFileBytes > 0 {
		if stream.IsReading() {
			m.BinFileData = make([]byte, MaxDatabaseBinWrapperSize)
		}
		binFileData := m.BinFileData[:m.BinFileBytes]
		stream.SerializeBytes(binFileData)
	}

	stream.SerializeUint64(&m.CreatedAt)

	if m.Version >= 2 {
		if stream.IsReading() {
			m.DestRelays = make([]bool, numRelays)
		}
		for i := range m.DestRelays {
			stream.SerializeBool(&m.DestRelays[i])
		}
	}

	if m.Version >= 3 {

		numRelayEntries := uint32(len(m.RelayStats))
		stream.SerializeUint32(&numRelayEntries)

		if stream.IsReading() {
			m.RelayStats = make([]analytics.RelayStatsEntry, numRelayEntries)
		}

		for i := uint32(0); i < numRelayEntries; i++ {
			entry := &m.RelayStats[i]

			stream.SerializeUint64(&entry.Timestamp)
			stream.SerializeUint64(&entry.ID)
			stream.SerializeUint32(&entry.NumSessions)
			stream.SerializeUint32(&entry.MaxSessions)
			stream.SerializeUint32(&entry.NumRoutable)
			stream.SerializeUint32(&entry.NumUnroutable)

			if m.Version >= 4 {
				stream.SerializeBool(&entry.Full)
			}

			if m.Version >= 5 {
				stream.SerializeFloat32(&entry.CPUUsage)

				stream.SerializeFloat32(&entry.BandwidthSentPercent)
				stream.SerializeFloat32(&entry.BandwidthReceivedPercent)

				stream.SerializeFloat32(&entry.EnvelopeSentPercent)
				stream.SerializeFloat32(&entry.EnvelopeReceivedPercent)

				stream.SerializeFloat32(&entry.BandwidthSentMbps)
				stream.SerializeFloat32(&entry.BandwidthReceivedMbps)

				stream.SerializeFloat32(&entry.EnvelopeSentMbps)
				stream.SerializeFloat32(&entry.EnvelopeReceivedMbps)
			}
		}

		numPingEntries := uint32(len(m.PingStats))
		stream.SerializeUint32(&numPingEntries)

		if stream.IsReading() {
			m.PingStats = make([]analytics.PingStatsEntry, numPingEntries)
		}

		for i := uint32(0); i < numPingEntries; i++ {
			entry := &m.PingStats[i]

			stream.SerializeUint64(&entry.Timestamp)
			stream.SerializeUint64(&entry.RelayA)
			stream.SerializeUint64(&entry.RelayB)
			stream.SerializeFloat32(&entry.RTT)
			stream.SerializeFloat32(&entry.Jitter)
			stream.SerializeFloat32(&entry.PacketLoss)
			stream.SerializeBool(&entry.Routable)
			stream.SerializeString(&entry.InstanceID, 64)
			stream.SerializeBool(&entry.Debug)
		}
	}

	if m.Version >= 4 {

		numFullRelayIDs := uint32(len(m.FullRelayIDs))
		stream.SerializeUint32(&numFullRelayIDs)

		if stream.IsReading() {
			m.FullRelayIDs = make([]uint64, numFullRelayIDs)
			m.FullRelayIndicesSet = make(map[int32]bool)
		}

		for i := uint32(0); i < numFullRelayIDs; i++ {
			stream.SerializeUint64(&m.FullRelayIDs[i])

			if stream.IsReading() {
				relayIndex, _ := m.RelayIDsToIndices[m.FullRelayIDs[i]]
				m.FullRelayIndicesSet[relayIndex] = true
			}
		}
	}

	return stream.Error()
}

func (m *RouteMatrix) GetNearRelays(directLatency float32, source_latitude float32, source_longitude float32, dest_latitude float32, dest_longitude float32, maxNearRelays int) ([]uint64, []net.UDPAddr) {

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
	nearRelayAddresses := make([]net.UDPAddr, 0, maxNearRelays)

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
		nearRelayAddresses = append(nearRelayAddresses, nearRelayData[i].Addr)
		nearRelayIDMap[nearRelayData[i].ID] = struct{}{}
	}

	// If we already have enough relays, stop and return them

	if len(nearRelayIDs) == maxNearRelays {
		return nearRelayIDs, nearRelayAddresses
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

		// don't add the same relay twice
		if _, ok := nearRelayIDMap[nearRelayData[i].ID]; ok {
			continue
		}

		nearRelayLatency := 3.0 / 2.0 * float32(core.SpeedOfLightTimeMilliseconds(sourceLatitude, sourceLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude, destLatitude, destLongitude))
		if nearRelayLatency > directLatency+latencyThreshold {
			continue
		}

		nearRelayIDs = append(nearRelayIDs, nearRelayData[i].ID)
		nearRelayAddresses = append(nearRelayAddresses, nearRelayData[i].Addr)
	}

	return nearRelayIDs, nearRelayAddresses
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
	buffer := make([]byte, bufferSize)
	writeStream, err := encoding.CreateWriteStream(buffer)
	if err != nil {
		return 0, err
	}
	if err = m.Serialize(writeStream); err != nil {
		return int64(writeStream.GetBytesProcessed()), err
	}
	writeStream.Flush()
	n, err := writer.Write(buffer[:writeStream.GetBytesProcessed()])
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
				if !m.DestRelays[i] && !m.DestRelays[j] {
					continue
				}
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
	relayPairs := 0
	relayPairsWithNoRoutes := 0
	relayPairsWithOneRoute := 0
	totalRouteLength := uint64(0)

	for i := range src {
		for j := range dest {
			if j < i {
				if !m.DestRelays[i] && !m.DestRelays[j] {
					continue
				}
				relayPairs++
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

	numDestRelays := 0
	for i := range m.DestRelays {
		if m.DestRelays[i] {
			numDestRelays++
		}
	}

	averageNumRoutes := float64(totalRoutes) / float64(numRelayPairs)
	averageRouteLength := float64(totalRouteLength) / float64(totalRoutes)

	fmt.Fprintf(writer, "\n%s Summary:\n\n", "Route")
	fmt.Fprintf(writer, "    %d relays\n", len(m.RelayIDs))
	fmt.Fprintf(writer, "    %d total routes\n", totalRoutes)
	fmt.Fprintf(writer, "    %d relay pairs\n", relayPairs)
	fmt.Fprintf(writer, "    %d destination relays\n", numDestRelays)
	fmt.Fprintf(writer, "    %.1f routes per relay pair on average (%d max)\n", averageNumRoutes, maxRoutesPerRelayPair)
	fmt.Fprintf(writer, "    %.1f relays per route on average (%d max)\n", averageRouteLength, maxRouteLength)
	fmt.Fprintf(writer, "    %.1f%% of relay pairs have only one route\n", float64(relayPairsWithOneRoute)/float64(numRelayPairs)*100)
	fmt.Fprintf(writer, "    %.1f%% of relay pairs have no route\n", float64(relayPairsWithNoRoutes)/float64(numRelayPairs)*100)
}

// JsonMatrixAnalysis assembles the analysis into a form that can be
// easily marshaled in the sender (json on the wire)
type JsonMatrixAnalysis struct {
	// RTT Improvement
	RttImprovementNone     int `json:"rttImprovementNone"`
	RttImprovement0_5ms    int `json:"rttImprovement0_5ms"`
	RttImprovement5_10ms   int `json:"rttImprovement5_10ms"`
	RttImprovement10_15ms  int `json:"rttImprovement10_15ms"`
	RttImprovement15_20ms  int `json:"rttImprovement15_20ms"`
	RttImprovement20_25ms  int `json:"rttImprovement20_25ms"`
	RttImprovement25_30ms  int `json:"rttImprovement25_30ms"`
	RttImprovement30_35ms  int `json:"rttImprovement30_35ms"`
	RttImprovement35_40ms  int `json:"rttImprovement35_40ms"`
	RttImprovement40_45ms  int `json:"rttImprovement40_45ms"`
	RttImprovement45_50ms  int `json:"rttImprovement45_50ms"`
	RttImprovement50plusms int `json:"rttImprovement50plusms"`

	// Route Summary
	RelayCount                    int     `json:"relayCount"`
	TotalRoutes                   int     `json:"totalRoutes"`
	RelayPairs                    int     `json:"relayPairs"`
	DestinationRelays             int     `json:"destinationRelays"`
	AvgRoutesPerRelayPair         float64 `json:"avgRoutesPerRelayPair"`
	MaxRoutesPerRelayPair         int     `json:"maxRoutesPerRelayPair"`
	AvgRelaysPerRoute             float64 `json:"avgRelaysPerRoute"`
	MaxRelaysPerRoute             int     `json:"maxRelaysPerRoute"`
	RelayPairsWithOneRoutePercent float64 `json:"relayPairsWithOneRoutePercent"`
	RelayPairsWIthNoRoutesPercent float64 `json:"relayPairsWIthNoRoutesPercent"`
}

func (jma *JsonMatrixAnalysis) String() string {

	var jmaString string

	jmaString = "RttImprovementNone: " + fmt.Sprintf("%d", jma.RttImprovementNone) + "\n"
	jmaString += "RttImprovement0_5ms: " + fmt.Sprintf("%d", jma.RttImprovement0_5ms) + "\n"
	jmaString += "RttImprovement5_10ms: " + fmt.Sprintf("%d", jma.RttImprovement5_10ms) + "\n"
	jmaString += "RttImprovement10_15ms: " + fmt.Sprintf("%d", jma.RttImprovement10_15ms) + "\n"
	jmaString += "RttImprovement15_20ms: " + fmt.Sprintf("%d", jma.RttImprovement15_20ms) + "\n"
	jmaString += "RttImprovement20_25ms: " + fmt.Sprintf("%d", jma.RttImprovement20_25ms) + "\n"
	jmaString += "RttImprovement25_30ms: " + fmt.Sprintf("%d", jma.RttImprovement25_30ms) + "\n"
	jmaString += "RttImprovement30_35ms: " + fmt.Sprintf("%d", jma.RttImprovement30_35ms) + "\n"
	jmaString += "RttImprovement35_40ms: " + fmt.Sprintf("%d", jma.RttImprovement35_40ms) + "\n"
	jmaString += "RttImprovement40_45ms: " + fmt.Sprintf("%d", jma.RttImprovement40_45ms) + "\n"
	jmaString += "RttImprovement45_50ms: " + fmt.Sprintf("%d", jma.RttImprovement45_50ms) + "\n"
	jmaString += "RttImprovement50plusms: " + fmt.Sprintf("%d", jma.RttImprovement50plusms) + "\n"
	jmaString += "RelayCount: " + fmt.Sprintf("%d", jma.RelayCount) + "\n"
	jmaString += "TotalRoutes: " + fmt.Sprintf("%d", jma.TotalRoutes) + "\n"
	jmaString += "RelayPairs: " + fmt.Sprintf("%d", jma.RelayPairs) + "\n"
	jmaString += "DestinationRelays: " + fmt.Sprintf("%d", jma.DestinationRelays) + "\n"
	jmaString += "AvgRoutesPerRelayPair: " + fmt.Sprintf("%f", jma.AvgRoutesPerRelayPair) + "\n"
	jmaString += "MaxRoutesPerRelayPair: " + fmt.Sprintf("%d", jma.MaxRoutesPerRelayPair) + "\n"
	jmaString += "AvgRelaysPerRoute: " + fmt.Sprintf("%f", jma.AvgRelaysPerRoute) + "\n"
	jmaString += "MaxRelaysPerRoute: " + fmt.Sprintf("%d", jma.MaxRelaysPerRoute) + "\n"
	jmaString += "RelayPairsWithOneRoutePercent: " + fmt.Sprintf("%f", jma.RelayPairsWithOneRoutePercent) + "\n"
	jmaString += "RelayPairsWIthNoRoutesPercent: " + fmt.Sprintf("%f", jma.RelayPairsWIthNoRoutesPercent)

	return jmaString
}

// GetJsonAnalysis returns a JsonMatrixAnalysis of the route matrix for ease
// of transmission on the wire
func (m *RouteMatrix) GetJsonAnalysis() JsonMatrixAnalysis {

	var jsonMatrixAnalysis JsonMatrixAnalysis

	src := m.RelayIDs
	dest := m.RelayIDs

	numRelayPairs := 0.0
	numValidRelayPairs := 0.0

	numValidRelayPairsWithoutImprovement := 0.0

	buckets := make([]int, 11)

	for i := range src {
		for j := range dest {
			if j < i {
				if !m.DestRelays[i] && !m.DestRelays[j] {
					continue
				}
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

	jsonMatrixAnalysis.RttImprovementNone = int(numValidRelayPairsWithoutImprovement)
	jsonMatrixAnalysis.RttImprovement0_5ms = buckets[0]
	jsonMatrixAnalysis.RttImprovement5_10ms = buckets[1]
	jsonMatrixAnalysis.RttImprovement10_15ms = buckets[2]
	jsonMatrixAnalysis.RttImprovement15_20ms = buckets[3]
	jsonMatrixAnalysis.RttImprovement20_25ms = buckets[4]
	jsonMatrixAnalysis.RttImprovement25_30ms = buckets[5]
	jsonMatrixAnalysis.RttImprovement30_35ms = buckets[6]
	jsonMatrixAnalysis.RttImprovement35_40ms = buckets[7]
	jsonMatrixAnalysis.RttImprovement40_45ms = buckets[8]
	jsonMatrixAnalysis.RttImprovement45_50ms = buckets[9]
	jsonMatrixAnalysis.RttImprovement50plusms = buckets[10]

	totalRoutes := uint64(0)
	maxRouteLength := int32(0)
	maxRoutesPerRelayPair := int32(0)
	relayPairs := 0
	relayPairsWithNoRoutes := 0
	relayPairsWithOneRoute := 0
	totalRouteLength := uint64(0)

	for i := range src {
		for j := range dest {
			if j < i {
				if !m.DestRelays[i] && !m.DestRelays[j] {
					continue
				}
				relayPairs++
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

	numDestRelays := 0
	for i := range m.DestRelays {
		if m.DestRelays[i] {
			numDestRelays++
		}
	}

	averageNumRoutes := float64(totalRoutes) / float64(numRelayPairs)
	averageRouteLength := float64(totalRouteLength) / float64(totalRoutes)

	jsonMatrixAnalysis.RelayCount = len(m.RelayIDs)
	jsonMatrixAnalysis.TotalRoutes = int(totalRoutes)
	jsonMatrixAnalysis.RelayPairs = relayPairs
	jsonMatrixAnalysis.DestinationRelays = numDestRelays
	jsonMatrixAnalysis.AvgRoutesPerRelayPair = averageNumRoutes
	jsonMatrixAnalysis.MaxRoutesPerRelayPair = int(maxRoutesPerRelayPair)
	jsonMatrixAnalysis.AvgRelaysPerRoute = averageRouteLength
	jsonMatrixAnalysis.MaxRelaysPerRoute = int(maxRouteLength)
	jsonMatrixAnalysis.RelayPairsWithOneRoutePercent = float64(relayPairsWithOneRoute) / float64(numRelayPairs) * 100
	jsonMatrixAnalysis.RelayPairsWIthNoRoutesPercent = float64(relayPairsWithNoRoutes) / float64(numRelayPairs) * 100

	return jsonMatrixAnalysis

}

func (m *RouteMatrix) GetResponseData() []byte {
	m.cachedResponseMutex.RLock()
	response := m.cachedResponse
	m.cachedResponseMutex.RUnlock()
	return response
}

func (m *RouteMatrix) WriteResponseData(bufferSize int) error {
	buffer := make([]byte, bufferSize)
	stream, err := encoding.CreateWriteStream(buffer)
	if err != nil {
		return fmt.Errorf("failed to create write stream in route matrix WriteResponseData(): %v", err)
	}

	if err := m.Serialize(stream); err != nil {
		return fmt.Errorf("failed to serialize route matrix in WriteResponseData(): %v", err)
	}
	stream.Flush()
	m.cachedResponseMutex.Lock()
	m.cachedResponse = buffer[:stream.GetBytesProcessed()]
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
