package common

import (
	// "bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	// "sort"
	// "sync"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
)

const RouteMatrixSerializeVersion = 7

type RouteMatrix struct {
	RelayIds           []uint64
	RelayIdToIndex     map[uint64]int32
	RelayAddresses     []net.UDPAddr // external IPs only
	RelayNames         []string
	RelayLatitudes     []float32
	RelayLongitudes    []float32
	RelayDatacenterIds []uint64
	RouteEntries       []core.RouteEntry
	BinFileBytes       int32
	BinFileData        []byte
	CreatedAt          uint64
	Version            uint32
	DestRelays         []bool

	// todo: how is this stored
	FullRelayIds []uint64
	// todo: what is this?

	// todo: what the fuck? this should just be an array...
	FullRelayIndexSet map[int32]bool

	// todo: review below
	PingStats  []analytics.PingStatsEntry
	RelayStats []analytics.RelayStatsEntry
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
		stream.SerializeString(&m.RelayNames[i], routing.MaxRelayNameLength)
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

		stream.SerializeInteger(&entry.DirectCost, -1, routing.InvalidRouteValue)
		stream.SerializeInteger(&entry.NumRoutes, 0, math.MaxInt32)

		for i := 0; i < int(entry.NumRoutes); i++ {
			stream.SerializeInteger(&entry.RouteCost[i], -1, routing.InvalidRouteValue)
			stream.SerializeInteger(&entry.RouteNumRelays[i], 0, core.MaxRelaysPerRoute)
			stream.SerializeUint32(&entry.RouteHash[i])
			for j := 0; j < int(entry.RouteNumRelays[i]); j++ {
				stream.SerializeInteger(&entry.RouteRelays[i][j], 0, math.MaxInt32)
			}
		}
	}

	stream.SerializeInteger(&m.BinFileBytes, 0, routing.MaxDatabaseBinWrapperSize)
	if m.BinFileBytes > 0 {
		if stream.IsReading() {
			m.BinFileData = make([]byte, routing.MaxDatabaseBinWrapperSize)
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

func (m *RouteMatrix) ReadFrom(reader io.Reader) (int64, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return 0, err
	}
	readStream := encoding.CreateReadStream(data)
	err = m.Serialize(readStream)
	return int64(readStream.GetBytesProcessed()), err
}

func (m *RouteMatrix) Write(bufferSize int) ([]byte, error) {
	buffer := make([]byte, bufferSize)
	ws, err := encoding.CreateWriteStream(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create write stream for route matrix: %v", err)
	}
	if err := m.Serialize(ws); err != nil {
		return nil, fmt.Errorf("failed to serialize route matrix: %v", err)
	}
	ws.Flush()
	return buffer[:ws.GetBytesProcessed()], nil
}

type RouteMatrixAnalysis struct {
	NumRelays     int
	NumDestRelays int
	NumRoutes     int
	NumRelayPairs int
	NumValidRelayPairs int
	numValidRelayPairsWithoutImprovement int
}

func (m *RouteMatrix) Analyze() *RouteMatrixAnalysis {

	// analyze relay pairs

	src := m.RelayIds
	dest := m.RelayIds

	numRelays := len(m.RelayIds)
	numRelayPairs := 0
	numValidRelayPairs := 0
	numValidRelayPairsWithoutImprovement := 0

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

	// count number of dest relays

	numDestRelays := 0
	for i := range m.DestRelays {
		if m.DestRelays[i] {
			numDestRelays++
		}
	}

	// return analysis to caller

	analysis := &RouteMatrixAnalysis{}

	analysis.NumRelays = numRelays
	analysis.NumDestRelays = numDestRelays
	analysis.NumRelayPairs = numRelayPairs
	analysis.NumValidRelayPairs = numValidRelayPairs
	analysis.numValidRelayPairsWithoutImprovement = numValidRelayPairsWithoutImprovement

	return analysis

	/*
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


		averageNumRoutes := float64(totalRoutes) / float64(numRelayPairs)
		averageRouteLength := float64(totalRouteLength) / float64(totalRoutes)

		fmt.Printf("\n%s Summary:\n\n", "Route")
		fmt.Printf("    %d relays\n", len(m.RelayIds))
		fmt.Printf("    %d total routes\n", totalRoutes)
		fmt.Printf("    %d relay pairs\n", relayPairs)
		fmt.Printf("    %d destination relays\n", numDestRelays)
		fmt.Printf("    %.1f routes per relay pair on average (%d max)\n", averageNumRoutes, maxRoutesPerRelayPair)
		fmt.Printf("    %.1f relays per route on average (%d max)\n", averageRouteLength, maxRouteLength)
		fmt.Printf("    %.1f%% of relay pairs have only one route\n", float64(relayPairsWithOneRoute)/float64(numRelayPairs)*100)
		fmt.Printf("    %.1f%% of relay pairs have no route\n", float64(relayPairsWithNoRoutes)/float64(numRelayPairs)*100)
	*/

	// todo
	return nil
}
