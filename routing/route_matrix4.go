package routing

import (
	"errors"
	"math"
	"net"
	"sort"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/encoding"
)

type RouteMatrix4 struct {
	RelayIDs           []uint64
	RelayAddresses     []net.UDPAddr
	RelayNames         []string
	RelayLatitudes     []float32
	RelayLongitudes    []float32
	RelayDatacenterIDs []uint64
	RouteEntries       []core.RouteEntry
}

func (m *RouteMatrix4) Serialize(stream encoding.Stream) error {
	numRelays := uint32(len(m.RelayIDs))
	stream.SerializeUint32(&numRelays)

	if stream.IsReading() {
		m.RelayIDs = make([]uint64, numRelays)
		m.RelayAddresses = make([]net.UDPAddr, numRelays)
		m.RelayNames = make([]string, numRelays)
		m.RelayLatitudes = make([]float32, numRelays)
		m.RelayLongitudes = make([]float32, numRelays)
		m.RelayDatacenterIDs = make([]uint64, numRelays)
	}

	for i := 0; i < len(m.RelayIDs); i++ {
		stream.SerializeUint64(&m.RelayIDs[i])
		stream.SerializeAddress(&m.RelayAddresses[i])
		stream.SerializeString(&m.RelayNames[i], math.MaxInt8)
		stream.SerializeFloat32(&m.RelayLatitudes[i])
		stream.SerializeFloat32(&m.RelayLongitudes[i])
		stream.SerializeUint64(&m.RelayDatacenterIDs[i])
	}

	numEntries := uint32(len(m.RouteEntries))
	stream.SerializeUint32(&numEntries)

	if stream.IsReading() {
		m.RouteEntries = make([]core.RouteEntry, numEntries)
	}

	for i := uint32(0); i < numEntries; i++ {
		entry := &m.RouteEntries[i]

		stream.SerializeInteger(&entry.DirectCost, 0, 10000)
		stream.SerializeInteger(&entry.NumRoutes, 0, math.MaxInt32)

		for i := 0; i < MaxRoutesPerRelayPair; i++ {
			stream.SerializeInteger(&entry.RouteCost[i], 0, 10000)
			stream.SerializeInteger(&entry.RouteNumRelays[i], 0, MaxRelays)
			stream.SerializeUint32(&entry.RouteHash[i])

			for j := 0; j < MaxRelays; j++ {
				stream.SerializeInteger(&entry.RouteRelays[i][j], 0, math.MaxInt32)
			}
		}
	}

	return stream.Error()
}

func (m *RouteMatrix4) GetNearRelays(latitude float64, longitude float64, maxNearRelays int) ([]NearRelayData, error) {
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
		nearRelayData[i].distance = int(core.HaversineDistance(lat1, long1, lat2, long2))
	}

	// IMPORTANT: Sort near relays by distance using a *stable sort*
	// This is necessary to ensure that relays are always sorted in the same order,
	// even when some relays have the same integer distance from the client. Without this
	// the set of near relays passed down to the SDK can be different from one slice to the next!

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].distance < nearRelayData[j].distance })

	if len(nearRelayData) > maxNearRelays {
		nearRelayData = nearRelayData[:maxNearRelays]
	}

	if len(nearRelayData) == 0 {
		return nil, errors.New("no near relays")
	}

	return nearRelayData, nil
}

func (m *RouteMatrix4) GetDatacenterRelayIDs(datacenterID uint64) []uint64 {
	relayIDs := make([]uint64, 0)

	for i := 0; i < len(m.RelayDatacenterIDs); i++ {
		if m.RelayDatacenterIDs[i] == datacenterID {
			relayIDs = append(relayIDs, m.RelayIDs[i])
		}
	}

	return relayIDs
}

// No-op, just for interface compatibility for now
func (m *RouteMatrix4) GetAcceptableRoutes(nearIDs []NearRelayData, destIDs []uint64, prevRouteHash uint64, rttEpsilon int32) ([]Route, error) {
	return nil, nil
}
