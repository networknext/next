package routing_test

import (
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
	"github.com/stretchr/testify/assert"
)

const (
	LosAngelesLatitude  = 33.9909
	LosAngelesLongitude = -118.2144

	TokyoLatitude  = 35.6762
	TokyoLongitude = 139.6503

	PapuaNewGuineaLatitude  = -6.3150
	PapuaNewGuineaLongitude = 143.9555

	HonoluluLatitude  = 21.3069
	HonoluluLongitude = -157.8583
)

func getRouteMatrix(t *testing.T, version uint32) routing.RouteMatrix {
	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)
	relayAddr3, err := net.ResolveUDPAddr("udp", "127.0.0.1:10002")
	assert.NoError(t, err)
	relayAddr4, err := net.ResolveUDPAddr("udp", "127.0.0.1:10003")
	assert.NoError(t, err)

	expected := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1, 3: 2, 4: 3},
		RelayIDs:           []uint64{1, 2, 3, 4},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2, *relayAddr3, *relayAddr4},
		RelayNames:         []string{"test.relay.1", "test.relay.2", "test.relay.3", "test.relay.4"},
		RelayLatitudes:     []float32{TokyoLatitude, PapuaNewGuineaLatitude, HonoluluLatitude, LosAngelesLatitude},
		RelayLongitudes:    []float32{TokyoLongitude, PapuaNewGuineaLongitude, HonoluluLongitude, LosAngelesLongitude},
		RelayDatacenterIDs: []uint64{11, 12, 13, 14},
		RouteEntries:       []core.RouteEntry{},
	}

	if version >= 2 {
		expected.Version = 2

		expected.DestRelays = []bool{false, false, false, true}
	}

	if version >= 3 {
		expected.Version = 3

		pingStat := analytics.PingStatsEntry{
			Timestamp:  uint64(time.Now().Unix()),
			RelayA:     uint64(0),
			RelayB:     uint64(1),
			RTT:        float32(rand.Int63n(255)),
			Jitter:     float32(rand.Int63n(255)),
			PacketLoss: float32(rand.Int63n(100)),
			Routable:   true,
			InstanceID: "test_ping_stat",
			Debug:      false,
		}

		expected.PingStats = []analytics.PingStatsEntry{pingStat}

		relayStat := analytics.RelayStatsEntry{
			Timestamp:     uint64(time.Now().Unix()),
			ID:            uint64(0),
			NumSessions:   uint32(1),
			MaxSessions:   uint32(1),
			NumRoutable:   uint32(4),
			NumUnroutable: uint32(0),
		}

		expected.RelayStats = []analytics.RelayStatsEntry{relayStat}
	}

	if version >= 4 {
		expected.Version = 4

		expected.RelayStats[0].Full = true

		expected.FullRelayIDs = []uint64{1}
	}

	if version >= 5 {
		expected.Version = 5

		expected.RelayStats[0].CPUUsage = float32(50)
		expected.RelayStats[0].BandwidthSentPercent = float32(51)
		expected.RelayStats[0].BandwidthReceivedPercent = float32(49)
		expected.RelayStats[0].EnvelopeSentPercent = float32(75)
		expected.RelayStats[0].EnvelopeReceivedPercent = float32(70)
		expected.RelayStats[0].BandwidthSentMbps = float32(510)
		expected.RelayStats[0].BandwidthReceivedMbps = float32(490)
		expected.RelayStats[0].EnvelopeSentMbps = float32(680)
		expected.RelayStats[0].EnvelopeReceivedMbps = float32(700)
	}

	if version >= 6 {
		expected.Version = 6

		internalAddr2, err := net.ResolveUDPAddr("udp", "128.0.0.1:10001")
		assert.NoError(t, err)

		expected.InternalAddressClientRoutableRelayIDs = []uint64{2}
		expected.InternalAddressClientRoutableRelayAddresses = []net.UDPAddr{*internalAddr2}

		expected.DestFirstRelayIDs = []uint64{2}
	}

	return expected
}

func TestRouteMatrixSerialize(t *testing.T) {
	t.Parallel()

	t.Run("serialize v0 and v1", func(t *testing.T) {
		expected := getRouteMatrix(t, 0)

		buffer := make([]byte, 10000)

		ws, err := encoding.CreateWriteStream(buffer)
		assert.NoError(t, err)
		err = expected.Serialize(ws)
		assert.NoError(t, err)

		ws.Flush()
		data := ws.GetData()[:ws.GetBytesProcessed()]

		var actual routing.RouteMatrix
		rs := encoding.CreateReadStream(data)
		err = actual.Serialize(rs)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual)
	})

	t.Run("serialize v2", func(t *testing.T) {
		expected := getRouteMatrix(t, 2)

		buffer := make([]byte, 10000)

		ws, err := encoding.CreateWriteStream(buffer)
		assert.NoError(t, err)
		err = expected.Serialize(ws)
		assert.NoError(t, err)

		ws.Flush()
		data := ws.GetData()[:ws.GetBytesProcessed()]

		var actual routing.RouteMatrix
		rs := encoding.CreateReadStream(data)
		err = actual.Serialize(rs)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual)
	})

	t.Run("serialize v3", func(t *testing.T) {
		expected := getRouteMatrix(t, 3)

		buffer := make([]byte, 10000)

		ws, err := encoding.CreateWriteStream(buffer)
		assert.NoError(t, err)
		err = expected.Serialize(ws)
		assert.NoError(t, err)

		ws.Flush()
		data := ws.GetData()[:ws.GetBytesProcessed()]

		var actual routing.RouteMatrix
		rs := encoding.CreateReadStream(data)
		err = actual.Serialize(rs)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual)
	})

	t.Run("serialize v4", func(t *testing.T) {
		expected := getRouteMatrix(t, 4)

		buffer := make([]byte, 10000)

		ws, err := encoding.CreateWriteStream(buffer)
		assert.NoError(t, err)
		err = expected.Serialize(ws)
		assert.NoError(t, err)

		ws.Flush()
		data := ws.GetData()[:ws.GetBytesProcessed()]

		var actual routing.RouteMatrix
		rs := encoding.CreateReadStream(data)
		err = actual.Serialize(rs)
		assert.NoError(t, err)

		assert.NotEmpty(t, actual.FullRelayIndicesSet)
		actual.FullRelayIndicesSet = nil

		assert.Equal(t, expected, actual)
	})

	t.Run("serialize v5", func(t *testing.T) {
		expected := getRouteMatrix(t, 5)

		buffer := make([]byte, 10000)

		ws, err := encoding.CreateWriteStream(buffer)
		assert.NoError(t, err)
		err = expected.Serialize(ws)
		assert.NoError(t, err)

		ws.Flush()
		data := ws.GetData()[:ws.GetBytesProcessed()]

		var actual routing.RouteMatrix
		rs := encoding.CreateReadStream(data)
		err = actual.Serialize(rs)
		assert.NoError(t, err)

		assert.NotEmpty(t, actual.FullRelayIndicesSet)
		actual.FullRelayIndicesSet = nil

		assert.Equal(t, expected, actual)
	})

	t.Run("serialize v6", func(t *testing.T) {
		expected := getRouteMatrix(t, 6)

		buffer := make([]byte, 10000)

		ws, err := encoding.CreateWriteStream(buffer)
		assert.NoError(t, err)
		err = expected.Serialize(ws)
		assert.NoError(t, err)

		ws.Flush()
		data := ws.GetData()[:ws.GetBytesProcessed()]

		var actual routing.RouteMatrix
		rs := encoding.CreateReadStream(data)
		err = actual.Serialize(rs)
		assert.NoError(t, err)

		assert.NotEmpty(t, actual.FullRelayIndicesSet)
		actual.FullRelayIndicesSet = nil

		assert.NotEmpty(t, actual.InternalAddressClientRoutableRelayAddrMap)
		actual.InternalAddressClientRoutableRelayAddrMap = nil

		assert.NotEmpty(t, actual.DestFirstRelayIDsSet)
		actual.DestFirstRelayIDsSet = nil

		assert.Equal(t, expected, actual)
	})
}

func TestRouteMatrixSerializeWithTimestampBackwardsComp(t *testing.T) {
	original := getRouteMatrix(t, 0)
	original.CreatedAt = 0
	expected := original
	original.CreatedAt = 19803

	buffer := make([]byte, 20000)

	ws, err := encoding.CreateWriteStream(buffer)
	assert.NoError(t, err)
	err = original.Serialize(ws)
	assert.NoError(t, err)

	ws.Flush()
	data := ws.GetData()[:ws.GetBytesProcessed()]

	var actual routing.RouteMatrix
	rs := encoding.CreateReadStream(data)
	err = actual.Serialize(rs)
	assert.NoError(t, err)

	assert.Equal(t, expected.BinFileBytes, actual.BinFileBytes)
	assert.Equal(t, expected.BinFileData, actual.BinFileData)
	assert.Equal(t, expected.RelayAddresses, actual.RelayAddresses)
	assert.Equal(t, expected.RelayDatacenterIDs, actual.RelayDatacenterIDs)
	assert.Equal(t, expected.RelayIDs, actual.RelayIDs)
	assert.Equal(t, expected.RelayIDsToIndices, actual.RelayIDsToIndices)
	assert.Equal(t, expected.RelayLatitudes, actual.RelayLatitudes)
	assert.Equal(t, expected.RelayLongitudes, actual.RelayLongitudes)
	assert.Equal(t, expected.RelayNames, actual.RelayNames)
	assert.Equal(t, expected.RouteEntries, actual.RouteEntries)
	assert.NotEqual(t, expected.CreatedAt, actual.CreatedAt)
	assert.Equal(t, actual.CreatedAt, uint64(19803))
	assert.Equal(t, expected.CreatedAt, uint64(0))
}

func TestRouteMatrix_GetNearRelays_NoNearRelays(t *testing.T) {
	t.Parallel()

	routeMatrix := routing.RouteMatrix{}

	nearRelayIDs, nearRelayAddresses := routeMatrix.GetNearRelays(0, 0, 0, 0, 0, core.MaxNearRelays, 0)

	assert.Empty(t, nearRelayIDs)
	assert.Empty(t, nearRelayAddresses)
}

func TestRouteMatrix_GetNearRelays_Success(t *testing.T) {
	t.Parallel()

	routeMatrix := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)

	expected := []uint64{1, 4, 3}

	actual, actualRelayAddrs := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, core.MaxNearRelays, 5)

	assert.Equal(t, expected, actual)

	// All relay addresses should be external
	expectedRelayAddrs := []net.UDPAddr{
		routeMatrix.RelayAddresses[0],
		routeMatrix.RelayAddresses[3],
		routeMatrix.RelayAddresses[2],
	}

	assert.Equal(t, expectedRelayAddrs, actualRelayAddrs)
}

func TestRouteMatrix_GetNearRelays_Success_InternalAddrClientRoutable(t *testing.T) {
	t.Parallel()

	routeMatrix := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)
	// Add in internal address for relay 1
	internalAddr1, err := net.ResolveUDPAddr("udp", "128.0.0.1:10000")
	assert.NoError(t, err)

	routeMatrix.InternalAddressClientRoutableRelayIDs = []uint64{1}
	routeMatrix.InternalAddressClientRoutableRelayAddresses = []net.UDPAddr{*internalAddr1}
	routeMatrix.InternalAddressClientRoutableRelayAddrMap = make(map[uint64]net.UDPAddr)
	routeMatrix.InternalAddressClientRoutableRelayAddrMap[uint64(1)] = *internalAddr1

	expected := []uint64{1, 4, 3}

	actual, actualRelayAddrs := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, core.MaxNearRelays, 5)

	assert.Equal(t, expected, actual)

	// Relay 1's address should be internal
	expectedRelayAddrs := []net.UDPAddr{
		routeMatrix.InternalAddressClientRoutableRelayAddresses[0],
		routeMatrix.RelayAddresses[3],
		routeMatrix.RelayAddresses[2],
	}

	assert.Equal(t, expectedRelayAddrs, actualRelayAddrs)
}

func TestRouteMatrix_GetNearRelays_WithMax(t *testing.T) {
	t.Parallel()

	routeMatrix := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)

	expected := routeMatrix.RelayIDs[:1]

	actual, actualRelayAddrs := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, 1, 5)

	assert.Equal(t, expected, actual)

	// All relay addresses should be external
	expectedRelayAddrs := []net.UDPAddr{
		routeMatrix.RelayAddresses[0],
	}

	assert.Equal(t, expectedRelayAddrs, actualRelayAddrs)
}

func TestRouteMatrix_GetNearRelays_WithMax_InternalAddrClientRoutable(t *testing.T) {
	t.Parallel()

	routeMatrix := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)
	// Add in internal address for relay 1
	internalAddr1, err := net.ResolveUDPAddr("udp", "128.0.0.1:10000")
	assert.NoError(t, err)

	routeMatrix.InternalAddressClientRoutableRelayIDs = []uint64{1}
	routeMatrix.InternalAddressClientRoutableRelayAddresses = []net.UDPAddr{*internalAddr1}
	routeMatrix.InternalAddressClientRoutableRelayAddrMap = make(map[uint64]net.UDPAddr)
	routeMatrix.InternalAddressClientRoutableRelayAddrMap[uint64(1)] = *internalAddr1

	expected := routeMatrix.RelayIDs[:1]

	actual, actualRelayAddrs := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, 1, 5)

	assert.Equal(t, expected, actual)

	// Relay 1's address should be internal
	expectedRelayAddrs := []net.UDPAddr{
		routeMatrix.InternalAddressClientRoutableRelayAddresses[0],
	}

	assert.Equal(t, expectedRelayAddrs, actualRelayAddrs)
}

func TestRouteMatrix_GetNearRelays_NoNearRelaysAroundSource(t *testing.T) {
	t.Parallel()

	routeMatrix := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)

	// Zero out the Tokyo relay lat/long so that it is far enough away that it
	// won't be picked up by the speed of light check
	routeMatrix.RelayLatitudes[0] = 0
	routeMatrix.RelayLongitudes[0] = 0

	expected := []uint64{4, 3}

	actual, actualRelayAddrs := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, core.MaxNearRelays, 5)

	assert.Equal(t, expected, actual)

	// All relay addresses should be external
	expectedRelayAddrs := []net.UDPAddr{
		routeMatrix.RelayAddresses[3],
		routeMatrix.RelayAddresses[2],
	}

	assert.Equal(t, expectedRelayAddrs, actualRelayAddrs)
}

func TestRouteMatrix_GetNearRelays_NoNearRelaysAroundSource_InternalAddrClientRoutable(t *testing.T) {
	t.Parallel()

	routeMatrix := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)

	// Add in internal address for relay 4
	internalAddr4, err := net.ResolveUDPAddr("udp", "128.0.0.1:10003")
	assert.NoError(t, err)

	routeMatrix.InternalAddressClientRoutableRelayIDs = []uint64{4}
	routeMatrix.InternalAddressClientRoutableRelayAddresses = []net.UDPAddr{*internalAddr4}
	routeMatrix.InternalAddressClientRoutableRelayAddrMap = make(map[uint64]net.UDPAddr)
	routeMatrix.InternalAddressClientRoutableRelayAddrMap[uint64(4)] = *internalAddr4

	// Zero out the Tokyo relay lat/long so that it is far enough away that it
	// won't be picked up by the speed of light check
	routeMatrix.RelayLatitudes[0] = 0
	routeMatrix.RelayLongitudes[0] = 0

	expected := []uint64{4, 3}

	actual, actualRelayAddrs := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, core.MaxNearRelays, 5)

	assert.Equal(t, expected, actual)

	// Relay 4's address should be internal
	expectedRelayAddrs := []net.UDPAddr{
		routeMatrix.InternalAddressClientRoutableRelayAddresses[0],
		routeMatrix.RelayAddresses[2],
	}

	assert.Equal(t, expectedRelayAddrs, actualRelayAddrs)
}

func TestRouteMatrix_GetNearRelays_NoNearRelaysAroundDest(t *testing.T) {
	t.Parallel()

	routeMatrix := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)

	// Zero out the Los Angeles relay lat/long so that it is far enough away that it
	// won't be picked up by the speed of light check
	routeMatrix.RelayLatitudes[len(routeMatrix.RelayLatitudes)-1] = 0
	routeMatrix.RelayLongitudes[len(routeMatrix.RelayLatitudes)-1] = 0

	expected := []uint64{1, 3}

	actual, actualRelayAddrs := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, core.MaxNearRelays, 5)

	assert.Equal(t, expected, actual)

	// All relay addresses should be external
	expectedRelayAddrs := []net.UDPAddr{
		routeMatrix.RelayAddresses[0],
		routeMatrix.RelayAddresses[2],
	}

	assert.Equal(t, expectedRelayAddrs, actualRelayAddrs)
}

func TestRouteMatrix_GetNearRelays_NoNearRelaysAroundDest_InternalAddrClientRoutable(t *testing.T) {
	t.Parallel()

	routeMatrix := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)

	// Add in internal address for relay 1
	internalAddr1, err := net.ResolveUDPAddr("udp", "128.0.0.1:10000")
	assert.NoError(t, err)

	routeMatrix.InternalAddressClientRoutableRelayIDs = []uint64{1}
	routeMatrix.InternalAddressClientRoutableRelayAddresses = []net.UDPAddr{*internalAddr1}
	routeMatrix.InternalAddressClientRoutableRelayAddrMap = make(map[uint64]net.UDPAddr)
	routeMatrix.InternalAddressClientRoutableRelayAddrMap[uint64(1)] = *internalAddr1

	// Zero out the Los Angeles relay lat/long so that it is far enough away that it
	// won't be picked up by the speed of light check
	routeMatrix.RelayLatitudes[len(routeMatrix.RelayLatitudes)-1] = 0
	routeMatrix.RelayLongitudes[len(routeMatrix.RelayLatitudes)-1] = 0

	expected := []uint64{1, 3}

	actual, actualRelayAddrs := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, core.MaxNearRelays, 5)

	assert.Equal(t, expected, actual)

	// Relay 1's address should be internal
	expectedRelayAddrs := []net.UDPAddr{
		routeMatrix.InternalAddressClientRoutableRelayAddresses[0],
		routeMatrix.RelayAddresses[2],
	}

	assert.Equal(t, expectedRelayAddrs, actualRelayAddrs)
}

func TestRouteMatrix_GetNearRelays_DestFirst(t *testing.T) {
	t.Parallel()

	// Serialize the route matrix so we get the dest first relay set
	expected := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)

	// Remove the InternalAddressClientRoutableRelayIDs to ensure external addresses are used
	expected.InternalAddressClientRoutableRelayIDs = []uint64{}
	expected.InternalAddressClientRoutableRelayAddresses = []net.UDPAddr{}

	buffer := make([]byte, 20000)

	ws, err := encoding.CreateWriteStream(buffer)
	assert.NoError(t, err)
	err = expected.Serialize(ws)
	assert.NoError(t, err)

	ws.Flush()
	data := ws.GetData()[:ws.GetBytesProcessed()]

	var actual routing.RouteMatrix
	rs := encoding.CreateReadStream(data)
	err = actual.Serialize(rs)
	assert.NoError(t, err)

	assert.Equal(t, expected.DestFirstRelayIDs, actual.DestFirstRelayIDs)
	assert.Equal(t, 1, len(actual.DestFirstRelayIDsSet))

	// Datacenter ID 12 maps to Relay 2
	// Even though Relay 2 does not have the same lat/long as the destination,
	// it is marked as a dest first relay and its datacenter ID is passed in
	expectedRelays := []uint64{2, 1, 4, 3}

	actualRelays, actualRelayAddrs := actual.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, core.MaxNearRelays, 12)

	assert.Equal(t, expectedRelays, actualRelays)

	// All relay addresses should be external
	expectedRelayAddrs := []net.UDPAddr{
		expected.RelayAddresses[1],
		expected.RelayAddresses[0],
		expected.RelayAddresses[3],
		expected.RelayAddresses[2],
	}

	assert.Equal(t, expectedRelayAddrs, actualRelayAddrs)
}

func TestRouteMatrix_GetNearRelays_DestFirst_InternalAddrClientRoutable(t *testing.T) {
	t.Parallel()

	// Serialize the route matrix so we get the dest first relay set
	expected := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)

	buffer := make([]byte, 20000)

	ws, err := encoding.CreateWriteStream(buffer)
	assert.NoError(t, err)
	err = expected.Serialize(ws)
	assert.NoError(t, err)

	ws.Flush()
	data := ws.GetData()[:ws.GetBytesProcessed()]

	var actual routing.RouteMatrix
	rs := encoding.CreateReadStream(data)
	err = actual.Serialize(rs)
	assert.NoError(t, err)

	assert.Equal(t, expected.DestFirstRelayIDs, actual.DestFirstRelayIDs)
	assert.Equal(t, 1, len(actual.DestFirstRelayIDsSet))

	// Datacenter ID 12 maps to Relay 2
	// Even though Relay 2 does not have the same lat/long as the destination,
	// it is marked as a dest first relay and its datacenter ID is passed in
	expectedRelays := []uint64{2, 1, 4, 3}

	actualRelays, actualRelayAddrs := actual.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, core.MaxNearRelays, 12)
	assert.Equal(t, expectedRelays, actualRelays)

	// Relay 2's address should be the internal
	expectedRelayAddrs := []net.UDPAddr{
		expected.InternalAddressClientRoutableRelayAddresses[0],
		expected.RelayAddresses[0],
		expected.RelayAddresses[3],
		expected.RelayAddresses[2],
	}
	assert.Equal(t, expectedRelayAddrs, actualRelayAddrs)
}

func TestRouteMatrixGetDatacenterIDsEmpty(t *testing.T) {
	t.Parallel()

	routeMatrix := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)

	expected := []uint64{}
	actual := routeMatrix.GetDatacenterRelayIDs(0)
	assert.Equal(t, expected, actual)
}

func TestRouteMatrixGetDatacenterIDsSuccess(t *testing.T) {
	t.Parallel()

	routeMatrix := getRouteMatrix(t, routing.RouteMatrixSerializeVersion)

	expected := routeMatrix.RelayIDs[0]
	actual := routeMatrix.GetDatacenterRelayIDs(11)
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expected, actual[0])

	expected = routeMatrix.RelayIDs[1]
	actual = routeMatrix.GetDatacenterRelayIDs(12)
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expected, actual[0])

	expected = routeMatrix.RelayIDs[2]
	actual = routeMatrix.GetDatacenterRelayIDs(13)
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expected, actual[0])

	expected = routeMatrix.RelayIDs[3]
	actual = routeMatrix.GetDatacenterRelayIDs(14)
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expected, actual[0])
}

func TestRouteMatrixGetJsonAnalysis(t *testing.T) {
	t.Parallel()

	fileName := "../../testdata/optimize.bin.prod-5_may"
	file, err := os.Open(fileName)
	assert.NoError(t, err)
	defer file.Close()

	var routeMatrix routing.RouteMatrix
	_, err = routeMatrix.ReadFrom(file)
	assert.NoError(t, err)

	// routeMatrix.WriteAnalysisTo(os.Stdout)
	// the line above prints:
	// 	RTT Improvement:
	//     None: 9706 (56.81%)
	//     0-5ms: 2605 (15.25%)
	//     5-10ms: 1898 (11.11%)
	//     10-15ms: 1110 (6.50%)
	//     15-20ms: 647 (3.79%)
	//     20-25ms: 369 (2.16%)
	//     25-30ms: 230 (1.35%)
	//     30-35ms: 108 (0.63%)
	//     35-40ms: 58 (0.34%)
	//     40-45ms: 39 (0.23%)
	//     45-50ms: 48 (0.28%)
	//     50ms+: 267 (1.56%)
	// Route Summary:
	//     289 relays
	//     171333 total routes
	//     17085 relay pairs
	//     67 destination relays
	//     10.0 routes per relay pair on average (16 max)
	//     3.6 relays per route on average (5 max)
	//     14.9% of relay pairs have only one route
	//     9.1% of relay pairs have no route

	jsonMatrixAnalysis := routeMatrix.GetJsonAnalysis()

	assert.Equal(t, 9706, jsonMatrixAnalysis.RttImprovementNone)
	assert.Equal(t, 2605, jsonMatrixAnalysis.RttImprovement0_5ms)
	assert.Equal(t, 1898, jsonMatrixAnalysis.RttImprovement5_10ms)
	assert.Equal(t, 1110, jsonMatrixAnalysis.RttImprovement10_15ms)
	assert.Equal(t, 647, jsonMatrixAnalysis.RttImprovement15_20ms)
	assert.Equal(t, 369, jsonMatrixAnalysis.RttImprovement20_25ms)
	assert.Equal(t, 230, jsonMatrixAnalysis.RttImprovement25_30ms)
	assert.Equal(t, 108, jsonMatrixAnalysis.RttImprovement30_35ms)
	assert.Equal(t, 58, jsonMatrixAnalysis.RttImprovement35_40ms)
	assert.Equal(t, 39, jsonMatrixAnalysis.RttImprovement40_45ms)
	assert.Equal(t, 48, jsonMatrixAnalysis.RttImprovement45_50ms)
	assert.Equal(t, 267, jsonMatrixAnalysis.RttImprovement50plusms)

	assert.Equal(t, 289, jsonMatrixAnalysis.RelayCount)
	assert.Equal(t, 171333, jsonMatrixAnalysis.TotalRoutes)
	assert.Equal(t, 17085, jsonMatrixAnalysis.RelayPairs)
	assert.Equal(t, 67, jsonMatrixAnalysis.DestinationRelays)
	assert.Equal(t, 10.02827041264267, jsonMatrixAnalysis.AvgRoutesPerRelayPair)
	assert.Equal(t, 16, jsonMatrixAnalysis.MaxRoutesPerRelayPair)
	assert.Equal(t, 3.619034278276806, jsonMatrixAnalysis.AvgRelaysPerRoute)
	assert.Equal(t, 5, jsonMatrixAnalysis.MaxRelaysPerRoute)
	assert.Equal(t, 14.937079309335674, jsonMatrixAnalysis.RelayPairsWithOneRoutePercent)
	assert.Equal(t, 9.089844893181153, jsonMatrixAnalysis.RelayPairsWIthNoRoutesPercent)

}

func TestRouteMatrixRelayFull(t *testing.T) {
	t.Parallel()

	expected := getRouteMatrix(t, 4)

	buffer := make([]byte, 20000)

	ws, err := encoding.CreateWriteStream(buffer)
	assert.NoError(t, err)
	err = expected.Serialize(ws)
	assert.NoError(t, err)

	ws.Flush()
	data := ws.GetData()[:ws.GetBytesProcessed()]

	var actual routing.RouteMatrix
	rs := encoding.CreateReadStream(data)
	err = actual.Serialize(rs)
	assert.NoError(t, err)

	assert.Equal(t, expected.FullRelayIDs, actual.FullRelayIDs)
	assert.Equal(t, 1, len(actual.FullRelayIndicesSet))

	relayIndex, exists := actual.RelayIDsToIndices[uint64(1)]
	assert.True(t, exists)
	assert.Equal(t, int32(0), relayIndex)

	val, ok := actual.FullRelayIndicesSet[relayIndex]
	assert.True(t, ok)
	assert.NotEmpty(t, val)

}

func TestRouteMatrixRelayStatsBandwidth(t *testing.T) {
	t.Parallel()

	expected := getRouteMatrix(t, 5)

	buffer := make([]byte, 20000)

	ws, err := encoding.CreateWriteStream(buffer)
	assert.NoError(t, err)
	err = expected.Serialize(ws)
	assert.NoError(t, err)

	ws.Flush()
	data := ws.GetData()[:ws.GetBytesProcessed()]

	var actual routing.RouteMatrix
	rs := encoding.CreateReadStream(data)
	err = actual.Serialize(rs)
	assert.NoError(t, err)

	assert.Equal(t, float32(50), actual.RelayStats[0].CPUUsage)
	assert.Equal(t, float32(51), actual.RelayStats[0].BandwidthSentPercent)
	assert.Equal(t, float32(49), actual.RelayStats[0].BandwidthReceivedPercent)
	assert.Equal(t, float32(75), actual.RelayStats[0].EnvelopeSentPercent)
	assert.Equal(t, float32(70), actual.RelayStats[0].EnvelopeReceivedPercent)
	assert.Equal(t, float32(510), actual.RelayStats[0].BandwidthSentMbps)
	assert.Equal(t, float32(490), actual.RelayStats[0].BandwidthReceivedMbps)
	assert.Equal(t, float32(680), actual.RelayStats[0].EnvelopeSentMbps)
	assert.Equal(t, float32(700), actual.RelayStats[0].EnvelopeReceivedMbps)
}
