package routing_test

import (
	"net"
	"testing"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
	"github.com/stretchr/testify/assert"
)

func getRouteMatrix(t *testing.T) routing.RouteMatrix {
	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	expected := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						0, 1,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1)},
			},
		},
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

func TestRouteMatrixNoNearRelays(t *testing.T) {
	routeMatrix := routing.RouteMatrix{}

	nearRelays, err := routeMatrix.GetNearRelays(0, 0, 0, 0, 0, core.MaxNearRelays)
	assert.EqualError(t, err, "no near relays")
	assert.Empty(t, nearRelays)
}

func TestRouteMatrixGetNearRelaysSuccess(t *testing.T) {
	routeMatrix := getRouteMatrix(t)

	expected := routeMatrix.RelayIDs

	actual, err := routeMatrix.GetNearRelays(80, 0, 0, 0, 0, core.MaxNearRelays)
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestRouteMatrixGetNearRelaysSuccessWithMax(t *testing.T) {
	routeMatrix := getRouteMatrix(t)

	expected := routeMatrix.RelayIDs[:1]

	actual, err := routeMatrix.GetNearRelays(80, 0, 0, 0, 0, 1)
	assert.NoError(t, err)

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

	expected := []uint64{1, 2}
	actual := routeMatrix.GetDatacenterRelayIDs(10)
	assert.Equal(t, expected, actual)
}
