
package main

import (
    "testing"
    "math"
    "io/ioutil"
    "strings"
    "strconv"
    "github.com/stretchr/testify/assert"
)

func TestProtocolVersionAtLeast(t *testing.T) {
    t.Parallel()
    assert.True(t, ProtocolVersionAtLeast(3, 0, 0, 3, 0, 0))
    assert.True(t, ProtocolVersionAtLeast(4, 0, 0, 3, 0, 0))
    assert.True(t, ProtocolVersionAtLeast(3, 1, 0, 3, 0, 0))
    assert.True(t, ProtocolVersionAtLeast(3, 0, 1, 3, 0, 0))
    assert.True(t, ProtocolVersionAtLeast(3, 4, 5, 3, 4, 5))
    assert.True(t, ProtocolVersionAtLeast(4, 0, 0, 3, 4, 5))
    assert.True(t, ProtocolVersionAtLeast(3, 5, 0, 3, 4, 5))
    assert.True(t, ProtocolVersionAtLeast(3, 4, 6, 3, 4, 5))
    assert.True(t, ProtocolVersionAtLeast(3, 1, 0, 3, 1, 0))
    assert.False(t, ProtocolVersionAtLeast(3, 0, 99, 3, 1, 1))
    assert.False(t, ProtocolVersionAtLeast(3, 1, 0, 3, 1, 1))
    assert.False(t, ProtocolVersionAtLeast(2, 0, 0, 3, 1, 1))
    assert.False(t, ProtocolVersionAtLeast(3, 0, 5, 3, 1, 0))
}

func TestHaversineDistance(t *testing.T) {
    t.Parallel()
    losangelesLatitude := 34.0522
    losangelesLongitude := -118.2437
    bostonLatitude := 42.3601
    bostonLongitude := -71.0589
    distance := HaversineDistance(losangelesLatitude, losangelesLongitude, bostonLatitude, bostonLongitude)
    assert.Equal(t, 4169.607203810275, distance)
}

func TestTriMatrixLength(t *testing.T) {
    t.Parallel()
    assert.Equal(t, 0, TriMatrixLength(0))
    assert.Equal(t, 0, TriMatrixLength(1))
    assert.Equal(t, 1, TriMatrixLength(2))
    assert.Equal(t, 3, TriMatrixLength(3))
    assert.Equal(t, 6, TriMatrixLength(4))
    assert.Equal(t, 10, TriMatrixLength(5))
    assert.Equal(t, 15, TriMatrixLength(6))
}

func TestTriMatrixIndex(t *testing.T) {
    t.Parallel()
    assert.Equal(t, 0, TriMatrixIndex(0,1))
    assert.Equal(t, 1, TriMatrixIndex(0,2))
    assert.Equal(t, 2, TriMatrixIndex(1,2))
    assert.Equal(t, 3, TriMatrixIndex(0,3))
    assert.Equal(t, 4, TriMatrixIndex(1,3))
    assert.Equal(t, 5, TriMatrixIndex(2,3))
    assert.Equal(t, 0, TriMatrixIndex(1,0))
    assert.Equal(t, 1, TriMatrixIndex(2,0))
    assert.Equal(t, 2, TriMatrixIndex(2,1))
    assert.Equal(t, 3, TriMatrixIndex(3,0))
    assert.Equal(t, 4, TriMatrixIndex(3,1))
    assert.Equal(t, 5, TriMatrixIndex(3,2))
}

func TestRouteManager(t *testing.T) {

    t.Parallel()

    routeManager := RouteManager{}

    assert.Equal(t, 0, routeManager.NumRoutes)

    routeManager.AddRoute(100, 1, 2, 3)
    assert.Equal(t, 1, routeManager.NumRoutes)
    assert.Equal(t, int32(100), routeManager.RouteCost[0])
    assert.Equal(t, int32(3), routeManager.RouteNumRelays[0])
    assert.Equal(t, int32(1), routeManager.RouteRelays[0][0])
    assert.Equal(t, int32(2), routeManager.RouteRelays[0][1])
    assert.Equal(t, int32(3), routeManager.RouteRelays[0][2])

    routeManager.AddRoute(200, 4, 5, 6)
    assert.Equal(t, 2, routeManager.NumRoutes)

    routeManager.AddRoute(100, 4, 5, 6)
    assert.Equal(t, 2, routeManager.NumRoutes)

    routeManager.AddRoute(190, 5, 6, 7, 8, 9)
    assert.Equal(t, 3, routeManager.NumRoutes)

    routeManager.AddRoute(180, 6, 7, 8)
    assert.Equal(t, 4, routeManager.NumRoutes)

    routeManager.AddRoute(175, 8, 9)
    assert.Equal(t, 5, routeManager.NumRoutes)

    routeManager.AddRoute(160, 9, 10, 11)
    assert.Equal(t, 6, routeManager.NumRoutes)

    routeManager.AddRoute(165, 10, 11, 12, 13, 14)
    assert.Equal(t, 7, routeManager.NumRoutes)

    routeManager.AddRoute(150, 11, 12)
    assert.Equal(t, 8, routeManager.NumRoutes)

    for i := 0; i < routeManager.NumRoutes-1; i++ {
        assert.True(t, routeManager.RouteCost[i] <= routeManager.RouteCost[i+1])
    }

    routeManager.AddRoute(1000, 12, 13, 14)
    assert.Equal(t, routeManager.NumRoutes, 8)
    for i := 0; i < routeManager.NumRoutes; i++ {
        assert.True(t, routeManager.RouteCost[i] != 1000)
    }

    routeManager.AddRoute(177, 13, 14, 15, 16, 17)
    assert.Equal(t, routeManager.NumRoutes, 8)
    for i := 0; i < routeManager.NumRoutes-1; i++ {
        assert.True(t, routeManager.RouteCost[i] <= routeManager.RouteCost[i+1])
    }
    found := false
    for i := 0; i < routeManager.NumRoutes; i++ {
        if routeManager.RouteCost[i] == 177 {
            found = true
        }
    }
    assert.True(t, found)

    assert.Equal(t, int32(100), routeManager.RouteCost[0])
    assert.Equal(t, int32(3), routeManager.RouteNumRelays[0])
    assert.Equal(t, int32(1), routeManager.RouteRelays[0][0])
    assert.Equal(t, int32(2), routeManager.RouteRelays[0][1])
    assert.Equal(t, int32(3), routeManager.RouteRelays[0][2])
    assert.Equal(t, RouteHash(1, 2, 3), routeManager.RouteHash[0])

    assert.Equal(t, int32(150), routeManager.RouteCost[1])
    assert.Equal(t, int32(2), routeManager.RouteNumRelays[1])
    assert.Equal(t, int32(11), routeManager.RouteRelays[1][0])
    assert.Equal(t, int32(12), routeManager.RouteRelays[1][1])
    assert.Equal(t, RouteHash(11, 12), routeManager.RouteHash[1])

    assert.Equal(t, int32(160), routeManager.RouteCost[2])
    assert.Equal(t, int32(3), routeManager.RouteNumRelays[2])
    assert.Equal(t, int32(9), routeManager.RouteRelays[2][0])
    assert.Equal(t, int32(10), routeManager.RouteRelays[2][1])
    assert.Equal(t, int32(11), routeManager.RouteRelays[2][2])
    assert.Equal(t, RouteHash(9, 10, 11), routeManager.RouteHash[2])

    assert.Equal(t, int32(165), routeManager.RouteCost[3])
    assert.Equal(t, int32(5), routeManager.RouteNumRelays[3])
    assert.Equal(t, int32(10), routeManager.RouteRelays[3][0])
    assert.Equal(t, int32(11), routeManager.RouteRelays[3][1])
    assert.Equal(t, int32(12), routeManager.RouteRelays[3][2])
    assert.Equal(t, int32(13), routeManager.RouteRelays[3][3])
    assert.Equal(t, int32(14), routeManager.RouteRelays[3][4])
    assert.Equal(t, RouteHash(10, 11, 12, 13, 14), routeManager.RouteHash[3])

    assert.Equal(t, int32(175), routeManager.RouteCost[4])
    assert.Equal(t, int32(2), routeManager.RouteNumRelays[4])
    assert.Equal(t, int32(8), routeManager.RouteRelays[4][0])
    assert.Equal(t, int32(9), routeManager.RouteRelays[4][1])
    assert.Equal(t, RouteHash(8, 9), routeManager.RouteHash[4])

    assert.Equal(t, int32(177), routeManager.RouteCost[5])
    assert.Equal(t, int32(5), routeManager.RouteNumRelays[5])
    assert.Equal(t, int32(13), routeManager.RouteRelays[5][0])
    assert.Equal(t, int32(14), routeManager.RouteRelays[5][1])
    assert.Equal(t, int32(15), routeManager.RouteRelays[5][2])
    assert.Equal(t, int32(16), routeManager.RouteRelays[5][3])
    assert.Equal(t, int32(17), routeManager.RouteRelays[5][4])
    assert.Equal(t, RouteHash(13, 14, 15, 16, 17), routeManager.RouteHash[5])

    assert.Equal(t, int32(180), routeManager.RouteCost[6])
    assert.Equal(t, int32(3), routeManager.RouteNumRelays[6])
    assert.Equal(t, int32(6), routeManager.RouteRelays[6][0])
    assert.Equal(t, int32(7), routeManager.RouteRelays[6][1])
    assert.Equal(t, int32(8), routeManager.RouteRelays[6][2])
    assert.Equal(t, RouteHash(6, 7, 8), routeManager.RouteHash[6])

    assert.Equal(t, int32(190), routeManager.RouteCost[7])
    assert.Equal(t, int32(5), routeManager.RouteNumRelays[7])
    assert.Equal(t, int32(5), routeManager.RouteRelays[7][0])
    assert.Equal(t, int32(6), routeManager.RouteRelays[7][1])
    assert.Equal(t, int32(7), routeManager.RouteRelays[7][2])
    assert.Equal(t, int32(8), routeManager.RouteRelays[7][3])
    assert.Equal(t, int32(9), routeManager.RouteRelays[7][4])
    assert.Equal(t, RouteHash(5, 6, 7, 8, 9), routeManager.RouteHash[7])

    // todo: test for ignoring loops

    // todo: test for ignoring cost < 0

    // todo: test for ignoring routes with same datacenter
}

func TestOptimize(t *testing.T) {

    t.Parallel()

	costData, err := ioutil.ReadFile("cost.txt")

    assert.NoError(t, err)

    costStrings := strings.Split(string(costData), ",")

    costValues := make([]int, len(costStrings))
    
    for i := range costStrings {
    	costValues[i], err = strconv.Atoi(costStrings[i])
        assert.NoError(t, err)
    }

    numRelays := int(math.Sqrt(float64(len(costValues))))

    entryCount := TriMatrixLength(numRelays)

    cost := make([]int32, entryCount)

    for i := 0; i < numRelays; i++ {
    	for j := 0; j < numRelays; j++ {
    		if i == j {
    			continue
    		}
    		index := TriMatrixIndex(i,j)
    		cost[index] = int32(costValues[i+j*numRelays])
    	}
    }

    routes := Optimize(numRelays, cost)

    buckets := Analyze(numRelays, routes)

    assert.Equal(t, 36558, buckets[0])
    assert.Equal(t, 3309, buckets[1])
    assert.Equal(t, 1590, buckets[2])
    assert.Equal(t, 855, buckets[3])
    assert.Equal(t, 515, buckets[4])
    assert.Equal(t, 1724, buckets[5])

    for i := 0; i < entryCount; i++ {
        assert.True(t, routes[i].NumRoutes >= 0)
        assert.True(t, routes[i].NumRoutes <= MaxRoutesPerRelayPair)
        for j := 0; j < int(routes[i].NumRoutes); j++ {
            assert.True(t, routes[i].DirectCost == -1 || routes[i].DirectCost >= routes[i].RouteCost[j])
            assert.True(t, routes[i].RouteNumRelays[j] >= 0)
            assert.True(t, routes[i].RouteNumRelays[j] <= MaxRelays)
            relays := make(map[int32]bool, 0)
            for k := 0; k < int(routes[i].RouteNumRelays[j]); k++ {
                _, found := relays[routes[i].RouteRelays[j][k]]
                assert.False(t, found)
                relays[routes[i].RouteRelays[j][k]] = true
            }
        }
    }
}
