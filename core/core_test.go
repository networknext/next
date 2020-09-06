
package main

import (
    "testing"
    "math"
    "io/ioutil"
    "strings"
    "strconv"
    "github.com/stretchr/testify/assert"
)

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
