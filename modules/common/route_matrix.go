package common

import (
	"fmt"
	"math"
	"math/rand"
	"net"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"

	// todo: we want to remove this
	"github.com/networknext/backend/modules-old/analytics"
)

const (
	RouteMatrixVersion_Min   = 7
	RouteMatrixVersion_Max   = 7
	RouteMatrixVersion_Write = 7

	MaxDatabaseBinWrapperSize = 100000000 // todo: 100mb seems overkill :)
)

type RouteMatrix struct {
	RelayIds           []uint64
	RelayIdToIndex     map[uint64]int32
	RelayAddresses     []net.UDPAddr // external IPs only
	RelayNames         []string
	RelayLatitudes     []float32
	RelayLongitudes    []float32
	RelayDatacenterIds []uint64
	DestRelays         []bool
	RouteEntries       []core.RouteEntry
	BinFileBytes       int32
	BinFileData        []byte
	CreatedAt          uint64
	Version            uint32
	FullRelayIds       []uint64
	FullRelayIndexSet  map[int32]bool // todo: this should probably just be an array of bools? why do we need a map?
}

func (m *RouteMatrix) Serialize(stream encoding.Stream) error {

	stream.SerializeUint32(&m.Version)

	numRelays := uint32(len(m.RelayIds))

	stream.SerializeUint32(&numRelays)

	if stream.IsReading() {
		m.RelayIdToIndex = make(map[uint64]int32)
		m.RelayIds = make([]uint64, numRelays)
		m.RelayAddresses = make([]net.UDPAddr, numRelays)
		m.RelayNames = make([]string, numRelays)
		m.RelayLatitudes = make([]float32, numRelays)
		m.RelayLongitudes = make([]float32, numRelays)
		m.RelayDatacenterIds = make([]uint64, numRelays)
	}

	for i := uint32(0); i < numRelays; i++ {
		stream.SerializeUint64(&m.RelayIds[i])
		stream.SerializeAddress(&m.RelayAddresses[i])
		stream.SerializeString(&m.RelayNames[i], MaxRelayNameLength)
		stream.SerializeFloat32(&m.RelayLatitudes[i])
		stream.SerializeFloat32(&m.RelayLongitudes[i])
		stream.SerializeUint64(&m.RelayDatacenterIds[i])

		if stream.IsReading() {
			m.RelayIdToIndex[m.RelayIds[i]] = int32(i)
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

	stream.SerializeInteger(&m.BinFileBytes, 0, MaxDatabaseBinWrapperSize) // todo: eventually just serialize an unbounded uint32
	if m.BinFileBytes > 0 {
		if stream.IsReading() {
			m.BinFileData = make([]byte, m.BinFileBytes)
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

	if m.Version >= 3 && m.Version < 7 {

		numRelayEntries := uint32(0)

		stream.SerializeUint32(&numRelayEntries)

		// todo: we don't want to use the old "analytics" module here

		var relayStats []analytics.RelayStatsEntry

		if stream.IsReading() {
			relayStats = make([]analytics.RelayStatsEntry, numRelayEntries)
		}

		for i := uint32(0); i < numRelayEntries; i++ {

			entry := &relayStats[i]

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

		numPingEntries := uint32(0)

		stream.SerializeUint32(&numPingEntries)

		var pingStats []analytics.PingStatsEntry

		if stream.IsReading() {
			pingStats = make([]analytics.PingStatsEntry, numPingEntries)
		}

		for i := uint32(0); i < numPingEntries; i++ {

			entry := &pingStats[i]

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

		numFullRelayIds := uint32(len(m.FullRelayIds))

		stream.SerializeUint32(&numFullRelayIds)

		if stream.IsReading() {
			m.FullRelayIds = make([]uint64, numFullRelayIds)
			m.FullRelayIndexSet = make(map[int32]bool)
		}

		for i := uint32(0); i < numFullRelayIds; i++ {
			stream.SerializeUint64(&m.FullRelayIds[i])
			if stream.IsReading() {
				relayIndex, _ := m.RelayIdToIndex[m.FullRelayIds[i]]
				m.FullRelayIndexSet[relayIndex] = true
			}
		}
	}

	if m.Version == 6 {

		// dummy vars because we don't support this feature anymore
		var InternalAddressClientRoutableRelayIDs []uint64
		var InternalAddressClientRoutableRelayAddresses []net.UDPAddr // internal IPs only
		var InternalAddressClientRoutableRelayAddrMap map[uint64]net.UDPAddr
		var DestFirstRelayIDs []uint64
		var DestFirstRelayIDsSet map[uint64]bool

		numInternalAddressClientRoutableRelayIDs := uint32(len(InternalAddressClientRoutableRelayIDs))
		stream.SerializeUint32(&numInternalAddressClientRoutableRelayIDs)

		if stream.IsReading() {
			InternalAddressClientRoutableRelayIDs = make([]uint64, numInternalAddressClientRoutableRelayIDs)
			InternalAddressClientRoutableRelayAddresses = make([]net.UDPAddr, numInternalAddressClientRoutableRelayIDs)
			InternalAddressClientRoutableRelayAddrMap = make(map[uint64]net.UDPAddr)
		}

		for i := uint32(0); i < numInternalAddressClientRoutableRelayIDs; i++ {
			stream.SerializeUint64(&InternalAddressClientRoutableRelayIDs[i])
			stream.SerializeAddress(&InternalAddressClientRoutableRelayAddresses[i])

			if stream.IsReading() {
				InternalAddressClientRoutableRelayAddrMap[InternalAddressClientRoutableRelayIDs[i]] = InternalAddressClientRoutableRelayAddresses[i]
			}
		}

		numDestFirstRelayIDs := uint32(len(DestFirstRelayIDs))

		stream.SerializeUint32(&numDestFirstRelayIDs)

		if stream.IsReading() {
			DestFirstRelayIDs = make([]uint64, numDestFirstRelayIDs)
			DestFirstRelayIDsSet = make(map[uint64]bool)
		}

		for i := uint32(0); i < numDestFirstRelayIDs; i++ {
			stream.SerializeUint64(&DestFirstRelayIDs[i])

			if stream.IsReading() {
				DestFirstRelayIDsSet[DestFirstRelayIDs[i]] = true
			}
		}
	}

	return stream.Error()
}

// todo: there should be unit tests for route matrix read and write

// todo: tests should include writing with the new codebase, and reading with the old codebase

func (m *RouteMatrix) Write(bufferSize int) ([]byte, error) {
	buffer := make([]byte, bufferSize)
	ws := encoding.CreateWriteStream(buffer)
	if err := m.Serialize(ws); err != nil {
		return nil, fmt.Errorf("failed to serialize route matrix: %v", err)
	}
	ws.Flush()
	return buffer[:ws.GetBytesProcessed()], nil
}

func (m *RouteMatrix) Read(buffer []byte) error {
	readStream := encoding.CreateReadStream(buffer)
	return m.Serialize(readStream)
}

type RouteMatrixAnalysis struct {
	TotalRoutes             int
	AverageNumRoutes        float32
	AverageRouteLength      float32
	NoRoutePercent          float32
	OneRoutePercent         float32
	NoDirectRoutePercent    float32
	RTTBucket_NoImprovement float32
	RTTBucket_0_5ms         float32
	RTTBucket_5_10ms        float32
	RTTBucket_10_15ms       float32
	RTTBucket_15_20ms       float32
	RTTBucket_20_25ms       float32
	RTTBucket_25_30ms       float32
	RTTBucket_30_35ms       float32
	RTTBucket_35_40ms       float32
	RTTBucket_40_45ms       float32
	RTTBucket_45_50ms       float32
	RTTBucket_50ms_Plus     float32
}

func (m *RouteMatrix) Analyze() RouteMatrixAnalysis {

	analysis := RouteMatrixAnalysis{}

	src := m.RelayIds
	dest := m.RelayIds

	numRelayPairs := 0.0
	numRelayPairsNoDirectRoute := 0.0
	numRelayPairsWithoutImprovement := 0.0

	buckets := make([]int, 11)

	for i := range src {
		for j := range dest {
			if j < i {
				if !m.DestRelays[i] && !m.DestRelays[j] {
					continue
				}
				abFlatIndex := TriMatrixIndex(i, j)
				numRelayPairs++
				if len(m.RouteEntries[abFlatIndex].RouteCost) > 0 {
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
						numRelayPairsWithoutImprovement++
					}
				} else {
					numRelayPairsNoDirectRoute++
				}
			}
		}
	}

	analysis.NoDirectRoutePercent = float32(numRelayPairsNoDirectRoute / numRelayPairs)

	analysis.RTTBucket_NoImprovement = float32(numRelayPairsWithoutImprovement / numRelayPairs * 100.0)
	analysis.RTTBucket_0_5ms = float32(float64(buckets[0]) / numRelayPairs * 100.0)
	analysis.RTTBucket_5_10ms = float32(float64(buckets[1]) / numRelayPairs * 100.0)
	analysis.RTTBucket_10_15ms = float32(float64(buckets[2]) / numRelayPairs * 100.0)
	analysis.RTTBucket_15_20ms = float32(float64(buckets[3]) / numRelayPairs * 100.0)
	analysis.RTTBucket_20_25ms = float32(float64(buckets[4]) / numRelayPairs * 100.0)
	analysis.RTTBucket_25_30ms = float32(float64(buckets[5]) / numRelayPairs * 100.0)
	analysis.RTTBucket_30_35ms = float32(float64(buckets[6]) / numRelayPairs * 100.0)
	analysis.RTTBucket_35_40ms = float32(float64(buckets[7]) / numRelayPairs * 100.0)
	analysis.RTTBucket_40_45ms = float32(float64(buckets[8]) / numRelayPairs * 100.0)
	analysis.RTTBucket_45_50ms = float32(float64(buckets[9]) / numRelayPairs * 100.0)
	analysis.RTTBucket_50ms_Plus = float32(float64(buckets[10]) / numRelayPairs * 100.0)

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

	analysis.TotalRoutes = int(totalRoutes)
	analysis.AverageNumRoutes = float32(averageNumRoutes)
	analysis.AverageRouteLength = float32(averageRouteLength)

	analysis.NoRoutePercent = float32(relayPairsWithNoRoutes / relayPairs)
	analysis.OneRoutePercent = float32(relayPairsWithOneRoute / relayPairs)

	return analysis
}

// todo: this function needs to be unit tested
func (routeMatrix *RouteMatrix) GetDatacenterRelays(datacenterId uint64) []uint64 {
	// todo: would be good if this wasn't O(n) where n is the number of relays
	relayIds := make([]uint64, 0, 8)
	for i := range routeMatrix.RelayDatacenterIds {
		if routeMatrix.RelayDatacenterIds[i] == datacenterId {
			relayIds = append(relayIds, routeMatrix.RelayIds[i])
		}
	}
	return relayIds
}

func GenerateRandomRouteMatrix() RouteMatrix {

	routeMatrix := RouteMatrix{
		Version: uint32(RandomInt(RouteMatrixVersion_Min, RouteMatrixVersion_Max)),
	}

	numRelays := RandomInt(0, 64)

	routeMatrix.RelayIds = make([]uint64, numRelays)
	routeMatrix.RelayAddresses = make([]net.UDPAddr, numRelays)
	routeMatrix.RelayNames = make([]string, numRelays)
	routeMatrix.RelayLatitudes = make([]float32, numRelays)
	routeMatrix.RelayLongitudes = make([]float32, numRelays)
	routeMatrix.RelayDatacenterIds = make([]uint64, numRelays)
	routeMatrix.DestRelays = make([]bool, numRelays)

	for i := 0; i < numRelays; i++ {
		routeMatrix.RelayIds[i] = rand.Uint64()
		routeMatrix.RelayAddresses[i] = RandomAddress()
		routeMatrix.RelayNames[i] = RandomString(MaxRelayNameLength)
		routeMatrix.RelayLatitudes[i] = rand.Float32()
		routeMatrix.RelayLongitudes[i] = rand.Float32()
		routeMatrix.RelayDatacenterIds[i] = rand.Uint64()
		routeMatrix.DestRelays[i] = RandomBool()
	}

	routeMatrix.RelayIdToIndex = make(map[uint64]int32)
	for i := range routeMatrix.RelayIds {
		routeMatrix.RelayIdToIndex[routeMatrix.RelayIds[i]] = int32(i)
	}

	routeMatrix.BinFileBytes = int32(RandomInt(100, 10000))
	routeMatrix.BinFileData = make([]byte, routeMatrix.BinFileBytes)
	RandomBytes(routeMatrix.BinFileData)

	routeMatrix.CreatedAt = rand.Uint64()
	routeMatrix.Version = uint32(RandomInt(RouteMatrixVersion_Min, RouteMatrixVersion_Max))

	numFullRelays := RandomInt(0, numRelays)

	routeMatrix.FullRelayIds = make([]uint64, numFullRelays)
	routeMatrix.FullRelayIndexSet = make(map[int32]bool)
	for i := range routeMatrix.FullRelayIds {
		routeMatrix.FullRelayIds[i] = routeMatrix.RelayIds[i]
		routeMatrix.FullRelayIndexSet[int32(i)] = true
	}

	numEntries := RandomInt(0, 1000)

	routeMatrix.RouteEntries = make([]core.RouteEntry, numEntries)

	for i := range routeMatrix.RouteEntries {
		routeMatrix.RouteEntries[i].DirectCost = int32(RandomInt(1, 1000))
		routeMatrix.RouteEntries[i].NumRoutes = int32(RandomInt(0, core.MaxRoutesPerEntry))
		for j := 0; j < int(routeMatrix.RouteEntries[i].NumRoutes); j++ {
			routeMatrix.RouteEntries[i].RouteCost[j] = int32(RandomInt(1, 1000))
			routeMatrix.RouteEntries[i].RouteNumRelays[j] = int32(RandomInt(1, core.MaxRelaysPerRoute))
			for k := 0; k < int(routeMatrix.RouteEntries[i].RouteNumRelays[j]); k++ {
				routeMatrix.RouteEntries[i].RouteRelays[j][k] = int32(k)
			}
			routeMatrix.RouteEntries[i].RouteHash[j] = core.RouteHash(routeMatrix.RouteEntries[i].RouteRelays[j][:]...)
		}
	}

	return routeMatrix
}
