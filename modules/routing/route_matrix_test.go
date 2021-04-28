package routing_test

import (
	"net"
	"testing"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
	"github.com/stretchr/testify/assert"
)

const LosAngelesLatitude = 33.9909
const LosAngelesLongitude = -118.2144

const TokyoLatitude = 35.6762
const TokyoLongitude = 139.6503

const PapuaNewGuineaLatitude = -6.3150
const PapuaNewGuineaLongitude = 143.9555

const HonoluluLatitude = 21.3069
const HonoluluLongitude = -157.8583

func getRouteMatrix(t *testing.T) routing.RouteMatrix {
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
		RelayDatacenterIDs: []uint64{10, 10, 10, 10},
		RouteEntries:       []core.RouteEntry{},
	}

	return expected
}

func TestRouteMatrixSerialize(t *testing.T) {
	expected := getRouteMatrix(t)

	ws, err := encoding.CreateWriteStream(10000)
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
}

func TestRouteMatrixSerializeWithTimestampBackwardsComp(t *testing.T) {
	original := getRouteMatrix(t)
	expected := original
	original.CreatedAt = 19803

	ws, err := encoding.CreateWriteStream(10000)
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

// todo: GetNearRelays tests should be extended to also check the second value returned, array of relay addresses is correct

func TestRouteMatrixNoNearRelays(t *testing.T) {
	routeMatrix := routing.RouteMatrix{}

	nearRelayIDs, nearRelayAddresses := routeMatrix.GetNearRelays(0, 0, 0, 0, 0, core.MaxNearRelays)

	assert.Empty(t, nearRelayIDs)
	assert.Empty(t, nearRelayAddresses)
}

func TestRouteMatrixGetNearRelays(t *testing.T) {
	routeMatrix := getRouteMatrix(t)

	expected := []uint64{1, 4, 3}

	actual, _ := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, core.MaxNearRelays)

	assert.Equal(t, expected, actual)
}

func TestRouteMatrixGetNearRelaysWithMax(t *testing.T) {
	routeMatrix := getRouteMatrix(t)

	expected := routeMatrix.RelayIDs[:1]

	actual, _ := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, 1)

	assert.Equal(t, expected, actual)
}

func TestRouteMatrixGetNearRelaysNoNearRelaysAroundSource(t *testing.T) {
	routeMatrix := getRouteMatrix(t)

	// Zero out the Tokyo relay lat/long so that it is far enough away that it
	// won't be picked up by the speed of light check
	routeMatrix.RelayLatitudes[0] = 0
	routeMatrix.RelayLongitudes[0] = 0

	expected := []uint64{4, 3}

	actual, _ := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, core.MaxNearRelays)

	assert.Equal(t, expected, actual)
}

func TestRouteMatrixGetNearRelaysNoNearRelaysAroundDest(t *testing.T) {
	routeMatrix := getRouteMatrix(t)

	// Zero out the Los Angeles relay lat/long so that it is far enough away that it
	// won't be picked up by the speed of light check
	routeMatrix.RelayLatitudes[len(routeMatrix.RelayLatitudes)-1] = 0
	routeMatrix.RelayLongitudes[len(routeMatrix.RelayLatitudes)-1] = 0

	expected := []uint64{1, 3}

	actual, _ := routeMatrix.GetNearRelays(30, TokyoLatitude, TokyoLongitude, LosAngelesLatitude, LosAngelesLongitude, core.MaxNearRelays)

	assert.Equal(t, expected, actual)
}

func TestRouteMatrixGetDatacenterIDsEmpty(t *testing.T) {
	routeMatrix := getRouteMatrix(t)

	expected := []uint64{}
	actual := routeMatrix.GetDatacenterRelayIDs(0)
	assert.Equal(t, expected, actual)
}

func TestRouteMatrixGetDatacenterIDsSuccess(t *testing.T) {
	routeMatrix := getRouteMatrix(t)

	expected := routeMatrix.RelayIDs
	actual := routeMatrix.GetDatacenterRelayIDs(10)
	assert.Equal(t, expected, actual)
}
