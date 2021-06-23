package routing_test

import (
	"net"
	"os"
	"testing"

	"github.com/networknext/backend/modules/analytics"
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
}

func TestRouteMatrixSerializeWithTimestampBackwardsComp(t *testing.T) {
	original := getRouteMatrix(t)
	expected := original
	original.CreatedAt = 19803

	buffer := make([]byte, 10000)

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

func TestRouteMatrixGetJsonAnalysis(t *testing.T) {

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
	expected := getRouteMatrix(t)
	// Fill in all the info up to v4
	expected.BinFileBytes = 0
	expected.BinFileData = []byte{}
	expected.CreatedAt = uint64(0)
	expected.Version = routing.RouteMatrixSerializeVersion
	expected.DestRelays = []bool{false, false, false, false}
	expected.PingStats = []analytics.PingStatsEntry{}
	expected.RelayStats = []analytics.RelayStatsEntry{}
	expected.FullRelayIDs = []uint64{1}

	buffer := make([]byte, 1000)

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
