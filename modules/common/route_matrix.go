package common

import (
	"fmt"
	"math"
	"net"

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
	FullRelayIds 	   []uint64
	FullRelayIndexSet  map[int32]bool   // this should just be an array of bools?
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

	if m.Version >= 3 && m.Version < 7 {

		numRelayEntries := uint32(0)

		stream.SerializeUint32(&numRelayEntries)

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

func (m *RouteMatrix) Read(buffer []byte) error {
	readStream := encoding.CreateReadStream(buffer)
	return m.Serialize(readStream)
}
