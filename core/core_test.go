
package main

import (
    "os"
    "fmt"
    "sort"
    "math"
    "time"
    "testing"
    "net"
    // "io/ioutil"
    // "strings"
    // "strconv"
    "hash/fnv"
    "github.com/stretchr/testify/assert"
)

func FuckOffGolang() {
    fmt.Fprintf(os.Stdout, "I'm sick of adding and removing the fmt and os imports as I work")
}

func RelayHash64(name string) uint64 {
    hash := fnv.New64a()
    hash.Write([]byte(name))
    return hash.Sum64()
}

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

func CheckAddress(t *testing.T, addressString string, expected string) {
    address := ParseAddress(addressString)
    buffer := make([]uint8, NEXT_ADDRESS_BYTES)
    WriteAddress(buffer, address)
    readAddress := ReadAddress(buffer)
    readAddressString := readAddress.String()
    assert.Equal(t, expected, readAddressString)
}

func TestAddress(t *testing.T) {
    CheckAddress(t, "127.0.0.1", "127.0.0.1:0")
    CheckAddress(t, "127.0.0.1:40000", "127.0.0.1:40000")
    CheckAddress(t, "1.2.3.4:50000", "1.2.3.4:50000")
}

func TestRouteManager(t *testing.T) {

    t.Parallel()

    routeManager := RouteManager{}
    routeManager.RelayDatacenter = make([]uint64, 256)
    for i := range routeManager.RelayDatacenter {
        routeManager.RelayDatacenter[i] = uint64(i)
    }

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
}

// todo: test for route manager ignoring loops

// todo: test for route manager ignoring routes with multiple relays from the same datacenter

// todo: disabled because it's slow
/*
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

    costThreshold := int32(5)

    relayDatacenters := make([]uint64, 1024)
    for i := range relayDatacenters {
        relayDatacenters[i] = uint64(i)
    }

    routes := Optimize(numRelays, cost, costThreshold, relayDatacenters)

    buckets := Analyze(numRelays, routes)

    assert.Equal(t, 36558, buckets[0])
    assert.Equal(t, 3309, buckets[1])
    assert.Equal(t, 1590, buckets[2])
    assert.Equal(t, 855, buckets[3])
    assert.Equal(t, 515, buckets[4])
    assert.Equal(t, 1724, buckets[5])

    // todo: go wide on checks here, they are very slow (1.6 second...)
    for i := 0; i < entryCount; i++ {
        assert.True(t, routes[i].NumRoutes >= 0)
        assert.True(t, routes[i].NumRoutes <= MaxRoutesPerEntry)
        for j := 0; j < int(routes[i].NumRoutes); j++ {
            assert.True(t, routes[i].DirectCost == -1 || routes[i].DirectCost >= routes[i].RouteCost[j])
            assert.True(t, routes[i].RouteNumRelays[j] >= 0)
            assert.True(t, routes[i].RouteNumRelays[j] <= MaxRelaysPerRoute)
            relays := make(map[int32]bool, 0)
            for k := 0; k < int(routes[i].RouteNumRelays[j]); k++ {
                _, found := relays[routes[i].RouteRelays[j][k]]
                assert.False(t, found)
                relays[routes[i].RouteRelays[j][k]] = true
            }
        }
    }
}
*/

func GetTestRelayId(name string) uint32 {
    hash := fnv.New32a()
    hash.Write([]byte(name))
    return hash.Sum32()
}

type TestRelayData struct {
    id         uint32
    name       string
    address    *net.UDPAddr
    publicKey  []byte
    privateKey []byte
    index      int
}

type TestEnvironment struct {
    relayArray []*TestRelayData
    relays     map[string]*TestRelayData
    cost       [][]int32
}

func NewTestEnvironment() *TestEnvironment {
    env := &TestEnvironment{}
    env.relays = make(map[string]*TestRelayData)
    return env
}

func (env *TestEnvironment) Clear() {
    numRelays := len(env.relays)
    env.cost = make([][]int32, numRelays)
    for i := 0; i < numRelays; i++ {
        env.cost[i] = make([]int32, numRelays)
        for j := 0; j < numRelays; j++ {
            env.cost[i][j] = -1
        }
    }
}

func (env *TestEnvironment) AddRelay(relayName string, relayAddress string) {       // todo: add datacenter string per-relay
    relay := &TestRelayData{}
    relay.id = GetTestRelayId(relayName)
    relay.name = relayName
    relay.address = ParseAddress(relayAddress)
    var err error
    relay.publicKey, relay.privateKey, err = GenerateRelayKeyPair()
    if err != nil {
        panic(err)
    }
    relay.index = len(env.relayArray)
    env.relays[relayName] = relay
    env.relayArray = append(env.relayArray, relay)
    env.Clear()
}

func (env *TestEnvironment) GetRelayDatacenters() []uint64 {
    // todo: get datacenter mapping properly
    relayDatacenters := make([]uint64, len(env.relays))
    for i := range relayDatacenters {
        relayDatacenters[i] = uint64(i)
    }    
    return relayDatacenters
}

func (env *TestEnvironment) SetCost(sourceRelayName string, destRelayName string, cost int32) {
    i := env.relays[sourceRelayName].index
    j := env.relays[destRelayName].index
    if j > i {
        i, j = j, i
    }
    env.cost[i][j] = cost
}

func (env *TestEnvironment) GetRelayIndex(relayName string) int {
    relayData := env.GetRelayData(relayName)
    if relayData != nil {
        return relayData.index
    }
    return -1
}

func (env *TestEnvironment) GetRelayData(relayName string) *TestRelayData {
    return env.relays[relayName]
}

func (env *TestEnvironment) GetCostMatrix() ([]int32, int) {
    numRelays := len(env.relays)
    entryCount := TriMatrixLength(numRelays)
    costMatrix := make([]int32, entryCount)
    for i := 0; i < numRelays; i++ {
        for j := 0; j < i; j++ {
            index := TriMatrixIndex(i, j)
            costMatrix[index] = env.cost[i][j]
        }
    }
    return costMatrix, numRelays
}

type TestRouteData struct {
    cost   int32
    relays []string
}

func (env *TestEnvironment) GetRoutes(routeMatrix []RouteEntry, sourceRelayName string, destRelayName string) []TestRouteData {
    sourceRelay := env.relays[sourceRelayName]
    destRelay := env.relays[destRelayName]
    i := sourceRelay.index
    j := destRelay.index
    if i == j {
        return nil
    }
    index := TriMatrixIndex(i, j)
    entry := routeMatrix[index]
    testRouteData := make([]TestRouteData, entry.NumRoutes)
    for k := 0; k < int(entry.NumRoutes); k++ {
        testRouteData[k].cost = entry.RouteCost[k]
        testRouteData[k].relays = make([]string, entry.RouteNumRelays[k])
        if j < i {
            for l := 0; l < int(entry.RouteNumRelays[k]); l++ {
                relayIndex := entry.RouteRelays[k][l]
                testRouteData[k].relays[l] = env.relayArray[relayIndex].name
            }
        } else {
            for l := 0; l < int(entry.RouteNumRelays[k]); l++ {
                relayIndex := entry.RouteRelays[k][int(entry.RouteNumRelays[k])-1-l]
                testRouteData[k].relays[l] = env.relayArray[relayIndex].name
            }
        }
    }
    return testRouteData
}

func (env *TestEnvironment) GetBestRouteCost(routeMatrix []RouteEntry, sourceRelays []string, sourceRelayCost[] int32, destRelays []string) int32 {
    sourceRelayIndex := make([]int32, len(sourceRelays))
    for i := range sourceRelays {
        sourceRelayIndex[i] = int32(env.GetRelayIndex(sourceRelays[i]))
        if sourceRelayIndex[i] == -1 {
            panic("bad source relay name")
        }
    }
    destRelayIndex := make([]int32, len(destRelays))
    for i := range destRelays {
        destRelayIndex[i] = int32(env.GetRelayIndex(destRelays[i]))
        if destRelayIndex[i] == -1 {
            panic("bad dest relay name")
        }
    }
    return GetBestRouteCost(routeMatrix, sourceRelayIndex, sourceRelayCost, destRelayIndex)
}

func (env *TestEnvironment) GetCurrentRouteCost(routeMatrix []RouteEntry, routeRelays []string, sourceRelays []string, sourceRelayCost[] int32, destRelays []string) int32 {
    routeRelayIndex := [MaxRelaysPerRoute]int32{}
    for i := range routeRelays {
        routeRelayIndex[i] = int32(env.GetRelayIndex(routeRelays[i]))
        if routeRelayIndex[i] == -1 {
            panic("bad route relay name")
        }
    }
    sourceRelayIndex := make([]int32, len(sourceRelays))
    for i := range sourceRelays {
        sourceRelayIndex[i] = int32(env.GetRelayIndex(sourceRelays[i]))
        if sourceRelayIndex[i] == -1 {
            panic("bad source relay name")
        }
    }
    destRelayIndex := make([]int32, len(destRelays))
    for i := range destRelays {
        destRelayIndex[i] = int32(env.GetRelayIndex(destRelays[i]))
        if destRelayIndex[i] == -1 {
            panic("bad dest relay name")
        }
    }
    return GetCurrentRouteCost(routeMatrix, int32(len(routeRelays)), routeRelayIndex, sourceRelayIndex, sourceRelayCost, destRelayIndex)
}

func (env *TestEnvironment) GetBestRoutes(routeMatrix []RouteEntry, sourceRelays []string, sourceRelayCost[] int32, destRelays []string, maxCost int32) []TestRouteData {
    sourceRelayIndex := make([]int32, len(sourceRelays))
    for i := range sourceRelays {
        sourceRelayIndex[i] = int32(env.GetRelayIndex(sourceRelays[i]))
        if sourceRelayIndex[i] == -1 {
            panic("bad source relay name")
        }
    }
    destRelayIndex := make([]int32, len(destRelays))
    for i := range destRelays {
        destRelayIndex[i] = int32(env.GetRelayIndex(destRelays[i]))
        if destRelayIndex[i] == -1 {
            panic("bad dest relay name")
        }
    }
    numBestRoutes := 0
    bestRoutes := make([]BestRoute, 1024)
    GetBestRoutes(routeMatrix, sourceRelayIndex, sourceRelayCost, destRelayIndex, maxCost, bestRoutes, &numBestRoutes)
    routes := make([]TestRouteData, numBestRoutes)
    for i := 0; i < numBestRoutes; i++ {
        routes[i].cost = bestRoutes[i].Cost
        routes[i].relays = make([]string, bestRoutes[i].NumRelays)
        if bestRoutes[i].NeedToReverse {
            for j := 0; j < int(bestRoutes[i].NumRelays); j++ {
                relayIndex := bestRoutes[i].Relays[int(bestRoutes[i].NumRelays)-1-j]
                routes[i].relays[j] = env.relayArray[relayIndex].name
            }
        } else {
            for j := 0; j < int(bestRoutes[i].NumRelays); j++ {
                relayIndex := bestRoutes[i].Relays[j]
                routes[i].relays[j] = env.relayArray[relayIndex].name
            }
        }
    }
    return routes
}

func (env *TestEnvironment) ReframeRouteHash(route []uint64) (int32, [MaxRelaysPerRoute]int32) {
    relayIdToIndex := make(map[uint64]int32)
    for _, v := range env.relays {
        id := RelayHash64(v.name)
        relayIdToIndex[id] = int32(v.index)
    }
    reframedRoute := [MaxRelaysPerRoute]int32{}
    result := ReframeRoute(route, relayIdToIndex, &reframedRoute)
    if !result {
        return 0, reframedRoute
    } else {
        return int32(len(route)), reframedRoute
    }
}

func (env *TestEnvironment) ReframeRoute(routeRelayNames []string) (int32, [MaxRelaysPerRoute]int32) {
    route := make([]uint64, len(routeRelayNames))
    for i := range routeRelayNames {
        route[i] = RelayHash64(routeRelayNames[i])
    }
    return env.ReframeRouteHash(route)
}

func (env *TestEnvironment) ReframeRelays(sourceRelayNames []string, destRelayNames []string) ([]int32, []int32) {
    sourceRelays := make([]int32, len(sourceRelayNames))
    for i := range sourceRelayNames {
        relayData, ok := env.relays[sourceRelayNames[i]]
        if !ok {
            panic("source relay does not exist")
        }
        sourceRelays[i] = int32(relayData.index)
    }
    destRelays := make([]int32, len(destRelayNames))
    for i := range destRelayNames {
        relayData, ok := env.relays[destRelayNames[i]]
        if !ok {
            panic("dest relay does not exist")
        }
        destRelays[i] = int32(relayData.index)
    }
    return sourceRelays, destRelays
}

func (env *TestEnvironment) GetBestRoute_Initial(routeMatrix []RouteEntry, sourceRelayNames []string, sourceRelayCost[] int32, destRelayNames []string, maxCost int32) (int32, []string) {

    sourceRelays, destRelays := env.ReframeRelays(sourceRelayNames, destRelayNames)

    bestRouteCost := int32(0)
    bestRouteNumRelays := int32(0)
    bestRouteRelays := [MaxRelaysPerRoute]int32{}

    result := GetBestRoute_Initial(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays)
    if !result {
        return 0, []string{}
    }

    bestRouteRelayNames := make([]string, bestRouteNumRelays)

    for i := 0; i < int(bestRouteNumRelays); i++ {
        routeData := env.relayArray[bestRouteRelays[i]]
        bestRouteRelayNames[i] = routeData.name
    }

    return bestRouteCost, bestRouteRelayNames
}

func (env *TestEnvironment) GetBestRoute_Update(routeMatrix []RouteEntry, sourceRelayNames []string, sourceRelayCost[] int32, destRelayNames []string, maxCost int32, costThreshold int32, currentRouteRelayNames []string) (int32, []string) {

    sourceRelays, destRelays := env.ReframeRelays(sourceRelayNames, destRelayNames)

    currentRouteNumRelays, currentRouteRelays := env.ReframeRoute(currentRouteRelayNames)
    if currentRouteNumRelays == 0 {
        panic("current route has no relays")
    }
     
    bestRouteCost := int32(0)
    bestRouteNumRelays := int32(0)
    bestRouteRelays := [MaxRelaysPerRoute]int32{}

    GetBestRoute_Update(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, costThreshold, currentRouteNumRelays, currentRouteRelays, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays)

    if bestRouteNumRelays == 0 {
        return 0, []string{}
    }
    
    bestRouteRelayNames := make([]string, bestRouteNumRelays)
    for i := 0; i < int(bestRouteNumRelays); i++ {
        routeData := env.relayArray[bestRouteRelays[i]]
        bestRouteRelayNames[i] = routeData.name
    }

    return bestRouteCost, bestRouteRelayNames
}

func TestTheTestEnvironment(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")
    env.AddRelay("c", "10.0.0.5")
    env.AddRelay("d", "10.0.0.6")
    env.AddRelay("e", "10.0.0.7")

    env.SetCost("losangeles", "chicago", 100)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceIndex := env.GetRelayIndex("losangeles")
    destIndex := env.GetRelayIndex("chicago")

    assert.True(t, sourceIndex != -1 )
    assert.True(t, destIndex != -1 )

    routeIndex := TriMatrixIndex(sourceIndex, destIndex)
    
    assert.Equal(t, int32(1), routeMatrix[routeIndex].NumRoutes)
}

func TestIndirectRoute3(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")
    env.AddRelay("c", "10.0.0.5")
    env.AddRelay("d", "10.0.0.6")
    env.AddRelay("e", "10.0.0.7")

    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

    assert.Equal(t, 1, len(routes))
    if len(routes) == 1 {
        assert.Equal(t, int32(20), routes[0].cost)
        assert.Equal(t, 3, len(routes[0].relays))
        if len(routes[0].relays) == 3 {
            assert.Equal(t, []string{"losangeles", "a", "chicago"}, routes[0].relays)
        }
    }
}

func TestIndirectRoute4(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")
    env.AddRelay("c", "10.0.0.5")
    env.AddRelay("d", "10.0.0.6")
    env.AddRelay("e", "10.0.0.7")

    env.SetCost("losangeles", "a", 10)
    env.SetCost("losangeles", "b", 100)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

    assert.True(t, len(routes) >= 1)
    if len(routes) >= 1 {
        assert.Equal(t, int32(30), routes[0].cost)
        assert.Equal(t, []string{"losangeles", "a", "b", "chicago"}, routes[0].relays)
    }
}

func TestIndirectRoute5(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")
    env.AddRelay("c", "10.0.0.5")
    env.AddRelay("d", "10.0.0.6")
    env.AddRelay("e", "10.0.0.7")

    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "c", 10)
    env.SetCost("c", "chicago", 10)

    env.SetCost("losangeles", "b", 100)
    env.SetCost("b", "chicago", 100)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

    assert.True(t, len(routes) >= 1)
    if len(routes) >= 1 {
        assert.Equal(t, int32(40), routes[0].cost)
        assert.Equal(t, []string{"losangeles", "a", "b", "c", "chicago"}, routes[0].relays)
    }
}

func TestFasterRoute3(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

    assert.Equal(t, 2, len(routes))
    if len(routes) == 2 {
        assert.Equal(t, int32(20), routes[0].cost)
        assert.Equal(t, []string{"losangeles", "a", "chicago"}, routes[0].relays)
        assert.Equal(t, int32(100), routes[1].cost)
        assert.Equal(t, []string{"losangeles", "chicago"}, routes[1].relays)
    }
}

func TestFasterRoute4(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

    assert.Equal(t, 3, len(routes))
    if len(routes) == 3 {
        assert.Equal(t, int32(30), routes[0].cost)
        assert.Equal(t, []string{"losangeles", "a", "b", "chicago"}, routes[0].relays)
    }
}

func TestFasterRoute5(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")
    env.AddRelay("c", "10.0.0.5")

    env.SetCost("losangeles", "chicago", 1000)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("losangeles", "b", 100)
    env.SetCost("losangeles", "c", 100)
    env.SetCost("a", "chicago", 100)
    env.SetCost("b", "chicago", 100)
    env.SetCost("c", "chicago", 10)
    env.SetCost("a", "b", 10)
    env.SetCost("a", "c", 100)
    env.SetCost("b", "c", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

    assert.Equal(t, 7, len(routes))
    if len(routes) == 7 {
        assert.Equal(t, int32(40), routes[0].cost)
        assert.Equal(t, []string{"losangeles", "a", "b", "c", "chicago"}, routes[0].relays)
    }
}

func TestSlowerRoute(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")
    env.AddRelay("c", "10.0.0.5")

    env.SetCost("losangeles", "chicago", 10)
    env.SetCost("losangeles", "a", 100)
    env.SetCost("a", "chicago", 100)
    env.SetCost("b", "chicago", 100)
    env.SetCost("c", "chicago", 100)
    env.SetCost("a", "b", 100)
    env.SetCost("a", "c", 100)
    env.SetCost("b", "c", 100)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

    assert.Equal(t, 1, len(routes))
    if len(routes) == 1 {
        assert.Equal(t, int32(10), routes[0].cost)
        assert.Equal(t, []string{"losangeles", "chicago"}, routes[0].relays)
    }
}

func TestRouteToken(t *testing.T) {

    t.Parallel()

    relayPublicKey := [...]byte{0x71, 0x16, 0xce, 0xc5, 0x16, 0x1a, 0xda, 0xc7, 0xa5, 0x89, 0xb2, 0x51, 0x2b, 0x67, 0x4f, 0x8f, 0x98, 0x21, 0xad, 0x8f, 0xe6, 0x2d, 0x39, 0xca, 0xe3, 0x9b, 0xec, 0xdf, 0x3e, 0xfc, 0x2c, 0x24}
    relayPrivateKey := [...]byte{0xb6, 0x7d, 0x01, 0x0d, 0xaf, 0xba, 0xd1, 0x40, 0x75, 0x99, 0x08, 0x15, 0x0d, 0x3a, 0xce, 0x7b, 0x82, 0x28, 0x01, 0x5f, 0x7d, 0xa0, 0x75, 0xb6, 0xc1, 0x15, 0x56, 0x33, 0xe1, 0x01, 0x99, 0xd6}
    masterPublicKey := [...]byte{0x6f, 0x58, 0xb4, 0xd7, 0x3d, 0xdc, 0x73, 0x06, 0xb8, 0x97, 0x3d, 0x22, 0x4d, 0xe6, 0xf1, 0xfd, 0x2a, 0xf0, 0x26, 0x7e, 0x8b, 0x1d, 0x93, 0x73, 0xd0, 0x40, 0xa9, 0x8b, 0x86, 0x75, 0xcd, 0x43}
    masterPrivateKey := [...]byte{0x2a, 0xad, 0xd5, 0x43, 0x4e, 0x52, 0xbf, 0x62, 0x0b, 0x76, 0x24, 0x18, 0xe1, 0x26, 0xfb, 0xda, 0x89, 0x95, 0x32, 0xde, 0x1d, 0x39, 0x7f, 0xcd, 0x7b, 0x7a, 0xd5, 0x96, 0x3b, 0x0d, 0x23, 0xe5}

    routeToken := &RouteToken{}
    routeToken.expireTimestamp = uint64(time.Now().Unix() + 10)
    routeToken.sessionId = 0x123131231313131
    routeToken.sessionVersion = 100
    routeToken.kbpsUp = 256
    routeToken.kbpsDown = 512
    routeToken.nextAddress = ParseAddress("127.0.0.1:40000")
    routeToken.privateKey = RandomBytes(NEXT_PRIVATE_KEY_BYTES)

    buffer := make([]byte, NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)

    WriteRouteToken(routeToken, buffer[:])

    readRouteToken, err := ReadRouteToken(buffer)

    assert.NoError(t, err)
    assert.Equal(t, routeToken, readRouteToken)

    err = WriteEncryptedRouteToken(buffer, routeToken, masterPrivateKey[:], relayPublicKey[:])

    assert.NoError(t, err)

    readRouteToken, err = ReadEncryptedRouteToken(buffer, masterPublicKey[:], relayPrivateKey[:])

    assert.NoError(t, err)
    assert.Equal(t, routeToken, readRouteToken)

    // todo: test "WriteRouteTokens"
}

func TestContinueToken(t *testing.T) {

    t.Parallel()

    relayPublicKey := [...]byte{0x71, 0x16, 0xce, 0xc5, 0x16, 0x1a, 0xda, 0xc7, 0xa5, 0x89, 0xb2, 0x51, 0x2b, 0x67, 0x4f, 0x8f, 0x98, 0x21, 0xad, 0x8f, 0xe6, 0x2d, 0x39, 0xca, 0xe3, 0x9b, 0xec, 0xdf, 0x3e, 0xfc, 0x2c, 0x24}
    relayPrivateKey := [...]byte{0xb6, 0x7d, 0x01, 0x0d, 0xaf, 0xba, 0xd1, 0x40, 0x75, 0x99, 0x08, 0x15, 0x0d, 0x3a, 0xce, 0x7b, 0x82, 0x28, 0x01, 0x5f, 0x7d, 0xa0, 0x75, 0xb6, 0xc1, 0x15, 0x56, 0x33, 0xe1, 0x01, 0x99, 0xd6}
    masterPublicKey := [...]byte{0x6f, 0x58, 0xb4, 0xd7, 0x3d, 0xdc, 0x73, 0x06, 0xb8, 0x97, 0x3d, 0x22, 0x4d, 0xe6, 0xf1, 0xfd, 0x2a, 0xf0, 0x26, 0x7e, 0x8b, 0x1d, 0x93, 0x73, 0xd0, 0x40, 0xa9, 0x8b, 0x86, 0x75, 0xcd, 0x43}
    masterPrivateKey := [...]byte{0x2a, 0xad, 0xd5, 0x43, 0x4e, 0x52, 0xbf, 0x62, 0x0b, 0x76, 0x24, 0x18, 0xe1, 0x26, 0xfb, 0xda, 0x89, 0x95, 0x32, 0xde, 0x1d, 0x39, 0x7f, 0xcd, 0x7b, 0x7a, 0xd5, 0x96, 0x3b, 0x0d, 0x23, 0xe5}

    continueToken := &ContinueToken{}
    continueToken.expireTimestamp = uint64(time.Now().Unix() + 10)
    continueToken.sessionId = 0x123131231313131
    continueToken.sessionVersion = 100

    buffer := make([]byte, NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES)

    WriteContinueToken(continueToken, buffer[:])

    readContinueToken, err := ReadContinueToken(buffer)

    assert.NoError(t, err)
    assert.Equal(t, continueToken, readContinueToken)

    err = WriteEncryptedContinueToken(buffer, continueToken, masterPrivateKey[:], relayPublicKey[:])

    assert.NoError(t, err)

    readContinueToken, err = ReadEncryptedContinueToken(buffer, masterPublicKey[:], relayPrivateKey[:])

    assert.NoError(t, err)
    assert.Equal(t, continueToken, readContinueToken)

    // todo: test "WriteContinueTokens"
}

func TestBestRouteCostSimple(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelays := []string{"losangeles"}
    sourceRelayCosts := []int32{10}

    destRelays := []string{"chicago"}

    bestRouteCost := env.GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCosts, destRelays)

    assert.Equal(t, int32(40), bestRouteCost)
}

func TestBestRouteCostComplex(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles.a", "10.0.0.1")
    env.AddRelay("losangeles.b", "10.0.0.2")
    env.AddRelay("chicago.a", "10.0.0.3")
    env.AddRelay("chicago.b", "10.0.0.4")
    env.AddRelay("a", "10.0.0.5")
    env.AddRelay("b", "10.0.0.6")

    env.SetCost("losangeles.a", "chicago.a", 100)
    env.SetCost("losangeles.a", "chicago.b", 150)
    env.SetCost("losangeles.a", "a", 10)

    env.SetCost("a", "chicago.a", 50)
    env.SetCost("a", "chicago.b", 20)
    env.SetCost("a", "b", 10)

    env.SetCost("b", "chicago.a", 10)
    env.SetCost("b", "chicago.b", 5)

    env.SetCost("losangeles.b", "chicago.a", 75)
    env.SetCost("losangeles.b", "chicago.b", 110)
    env.SetCost("losangeles.b", "a", 10)
    env.SetCost("losangeles.b", "b", 5)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelays := []string{"losangeles.a", "losangeles.b"}
    sourceRelayCosts := []int32{10, 5}

    destRelays := []string{"chicago.a", "chicago.b"}

    bestRouteCost := env.GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCosts, destRelays)

    assert.Equal(t, int32(15), bestRouteCost)
}

func TestBestRouteCostNoRoute(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles.a", "10.0.0.1")
    env.AddRelay("losangeles.b", "10.0.0.2")
    env.AddRelay("chicago.a", "10.0.0.3")
    env.AddRelay("chicago.b", "10.0.0.4")
    env.AddRelay("a", "10.0.0.5")
    env.AddRelay("b", "10.0.0.6")

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelays := []string{"losangeles.a", "losangeles.b"}
    sourceRelayCosts := []int32{10, 5}

    destRelays := []string{"chicago.a", "chicago.b"}

    bestRouteCost := env.GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCosts, destRelays)

    assert.Equal(t, int32(math.MaxInt32), bestRouteCost)
}

func TestCurrentRouteCost(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    routeRelays := []string{"losangeles", "a", "b", "chicago"}

    sourceRelays := []string{"losangeles"}
    sourceRelayCosts := []int32{10}

    destRelays := []string{"chicago"}

    currentRouteCost := env.GetCurrentRouteCost(routeMatrix, routeRelays, sourceRelays, sourceRelayCosts, destRelays)

    assert.Equal(t, int32(40), currentRouteCost)
}

func TestCurrentRouteCostReverse(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    routeRelays := []string{"chicago", "b", "a", "losangeles"}

    sourceRelays := []string{"chicago"}
    sourceRelayCosts := []int32{10}

    destRelays := []string{"losangeles"}

    currentRouteCost := env.GetCurrentRouteCost(routeMatrix, routeRelays, sourceRelays, sourceRelayCosts, destRelays)

    assert.Equal(t, int32(40), currentRouteCost)
}

func TestGetBestRoutesSimple(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelays := []string{"losangeles"}
    sourceRelayCosts := []int32{10}

    destRelays := []string{"chicago"}

    maxCost := int32(1000)

    bestRoutes := env.GetBestRoutes(routeMatrix, sourceRelays, sourceRelayCosts, destRelays, maxCost)

    sort.Slice(bestRoutes, func(i int, j int) bool { return bestRoutes[i].cost < bestRoutes[j].cost })

    expectedBestRoutes := make([]TestRouteData, 3)

    expectedBestRoutes[0].cost = 40
    expectedBestRoutes[0].relays = []string{"losangeles", "a", "b", "chicago"}

    expectedBestRoutes[1].cost = 70
    expectedBestRoutes[1].relays = []string{"losangeles", "a", "chicago"}

    expectedBestRoutes[2].cost = 110
    expectedBestRoutes[2].relays = []string{"losangeles", "chicago"}

    assert.Equal(t, expectedBestRoutes, bestRoutes)
}

func TestGetBestRoutesReverse(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelays := []string{"chicago"}
    sourceRelayCosts := []int32{10}

    destRelays := []string{"losangeles"}

    maxCost := int32(1000)

    bestRoutes := env.GetBestRoutes(routeMatrix, sourceRelays, sourceRelayCosts, destRelays, maxCost)

    sort.Slice(bestRoutes, func(i int, j int) bool { return bestRoutes[i].cost < bestRoutes[j].cost })

    expectedBestRoutes := make([]TestRouteData, 3)

    expectedBestRoutes[0].cost = 40
    expectedBestRoutes[0].relays = []string{"chicago", "b", "a", "losangeles"}

    expectedBestRoutes[1].cost = 70
    expectedBestRoutes[1].relays = []string{"chicago", "a", "losangeles"}

    expectedBestRoutes[2].cost = 110
    expectedBestRoutes[2].relays = []string{"chicago", "losangeles"}

    assert.Equal(t, expectedBestRoutes, bestRoutes)
}

func TestGetBestRoutesComplex(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles.a", "10.0.0.1")
    env.AddRelay("losangeles.b", "10.0.0.2")
    env.AddRelay("chicago.a", "10.0.0.3")
    env.AddRelay("chicago.b", "10.0.0.4")
    env.AddRelay("a", "10.0.0.5")
    env.AddRelay("b", "10.0.0.6")

    env.SetCost("losangeles.a", "chicago.a", 1)
    env.SetCost("losangeles.a", "chicago.b", 150)
    env.SetCost("losangeles.a", "a", 10)

    env.SetCost("a", "chicago.a", 50)
    env.SetCost("a", "chicago.b", 20)
    env.SetCost("a", "b", 10)

    env.SetCost("b", "chicago.a", 10)
    env.SetCost("b", "chicago.b", 5)

    env.SetCost("losangeles.b", "chicago.a", 75)
    env.SetCost("losangeles.b", "chicago.b", 110)
    env.SetCost("losangeles.b", "a", 10)
    env.SetCost("losangeles.b", "b", 5)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelays := []string{"losangeles.a", "losangeles.b"}
    sourceRelayCosts := []int32{5, 3}

    destRelays := []string{"chicago.a", "chicago.b"}

    maxCost := int32(30)

    bestRoutes := env.GetBestRoutes(routeMatrix, sourceRelays, sourceRelayCosts, destRelays, maxCost)

    sort.Slice(bestRoutes, func(i int, j int) bool { return bestRoutes[i].cost < bestRoutes[j].cost })

    expectedBestRoutes := make([]TestRouteData, 6)

    expectedBestRoutes[0].cost = 6
    expectedBestRoutes[0].relays = []string{"losangeles.a", "chicago.a"}

    expectedBestRoutes[1].cost = 13
    expectedBestRoutes[1].relays = []string{"losangeles.b", "b", "chicago.b"}

    expectedBestRoutes[2].cost = 18
    expectedBestRoutes[2].relays = []string{"losangeles.b", "b", "chicago.a"}

    expectedBestRoutes[3].cost = 24
    expectedBestRoutes[3].relays = []string{"losangeles.b", "a", "losangeles.a", "chicago.a"}

    expectedBestRoutes[4].cost = 28
    expectedBestRoutes[4].relays = []string{"losangeles.b", "a", "b", "chicago.b"}

    expectedBestRoutes[5].cost = 30
    expectedBestRoutes[5].relays = []string{"losangeles.a", "a", "b", "chicago.b"}

    assert.Equal(t, expectedBestRoutes, bestRoutes)
}

func TestReframeRoute(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles.a", "10.0.0.1")
    env.AddRelay("losangeles.b", "10.0.0.2")
    env.AddRelay("chicago.a", "10.0.0.3")
    env.AddRelay("chicago.b", "10.0.0.4")
    env.AddRelay("a", "10.0.0.5")
    env.AddRelay("b", "10.0.0.6")

    currentRoute := make([]string, 3)
    currentRoute[0] = "losangeles.a"
    currentRoute[1] = "a"
    currentRoute[2] = "chicago.b"

    numRouteRelays, routeRelays := env.ReframeRoute(currentRoute)

    assert.Equal(t, int32(3), numRouteRelays)
    assert.Equal(t, int32(0), routeRelays[0])
    assert.Equal(t, int32(4), routeRelays[1])
    assert.Equal(t, int32(3), routeRelays[2])
}

// todo: test for reframe source relays

// todo: make sure that if a relay has 100% PL that it is *excluded* from being used as a source

func TestEarlyOutDirect(t *testing.T) {

    routeShader := NewRouteShader()
    routeState := RouteState{}
    customer := NewCustomerConfig()
    assert.False(t, EarlyOutDirect(&routeShader, &routeState, &customer))

    routeState = RouteState{Veto: true}
    assert.True(t, EarlyOutDirect(&routeShader, &routeState, &customer))

    routeState = RouteState{Banned: true}
    assert.True(t, EarlyOutDirect(&routeShader, &routeState, &customer))

    routeState = RouteState{Disabled: true}
    assert.True(t, EarlyOutDirect(&routeShader, &routeState, &customer))

    routeState = RouteState{NotSelected: true}
    assert.True(t, EarlyOutDirect(&routeShader, &routeState, &customer))

    routeState = RouteState{B: true}
    assert.True(t, EarlyOutDirect(&routeShader, &routeState, &customer))

    routeShader = NewRouteShader()
    routeShader.DisableNetworkNext = true
    routeState = RouteState{}
    assert.True(t, EarlyOutDirect(&routeShader, &routeState, &customer))
    assert.True(t, routeState.Disabled)

    routeShader = NewRouteShader()
    routeShader.SelectionPercent = 0
    routeState = RouteState{}
    assert.True(t, EarlyOutDirect(&routeShader, &routeState, &customer))
    assert.True(t, routeState.NotSelected)

    routeShader = NewRouteShader()
    routeShader.SelectionPercent = 0
    routeState = RouteState{}
    assert.True(t, EarlyOutDirect(&routeShader, &routeState, &customer))
    assert.True(t, routeState.NotSelected)

    routeShader = NewRouteShader()
    routeShader.ABTest = true
    routeState = RouteState{}
    routeState.UserID = 0
    assert.False(t, EarlyOutDirect(&routeShader, &routeState, &customer))
    assert.True(t, routeState.ABTest)
    assert.True(t, routeState.A)
    assert.False(t, routeState.B)

    routeShader = NewRouteShader()
    routeShader.ABTest = true
    routeState = RouteState{}
    routeState.UserID = 1
    assert.True(t, EarlyOutDirect(&routeShader, &routeState, &customer))
    assert.True(t, routeState.ABTest)
    assert.False(t, routeState.A)
    assert.True(t, routeState.B)

    routeShader = NewRouteShader()
    routeState = RouteState{}
    customer = NewCustomerConfig()
    customer.BannedUsers[1000] = true
    assert.False(t, EarlyOutDirect(&routeShader, &routeState, &customer))

    routeShader = NewRouteShader()
    routeState = RouteState{}
    routeState.UserID = 1000
    customer = NewCustomerConfig()
    customer.BannedUsers[1000] = true
    assert.True(t, EarlyOutDirect(&routeShader, &routeState, &customer))
    assert.True(t, routeState.Banned)
}

func TestGetBestRoute_Initial_Simple(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelayNames := []string{"losangeles"}
    sourceRelayCosts := []int32{10}

    destRelayNames := []string{"chicago"}

    maxCost := int32(40)

    bestRouteCost, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

    assert.Equal(t, int32(40), bestRouteCost)
    assert.Equal(t, []string{"losangeles", "a", "b", "chicago"}, bestRouteRelays)
}

func TestGetBestRoute_Initial_Complex(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles.a", "10.0.0.1")
    env.AddRelay("losangeles.b", "10.0.0.2")
    env.AddRelay("chicago.a", "10.0.0.3")
    env.AddRelay("chicago.b", "10.0.0.4")
    env.AddRelay("a", "10.0.0.5")
    env.AddRelay("b", "10.0.0.6")

    env.SetCost("losangeles.a", "chicago.a", 100)
    env.SetCost("losangeles.a", "chicago.b", 150)
    env.SetCost("losangeles.a", "a", 10)

    env.SetCost("a", "chicago.a", 50)
    env.SetCost("a", "chicago.b", 20)
    env.SetCost("a", "b", 10)

    env.SetCost("b", "chicago.a", 10)
    env.SetCost("b", "chicago.b", 5)

    env.SetCost("losangeles.b", "chicago.a", 75)
    env.SetCost("losangeles.b", "chicago.b", 110)
    env.SetCost("losangeles.b", "a", 10)
    env.SetCost("losangeles.b", "b", 5)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
    sourceRelayCosts := []int32{5, 2}

    destRelayNames := []string{"chicago.a", "chicago.b"}

    maxCost := int32(20)

    bestRouteCost, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

    assert.True(t, bestRouteCost > 0)
    assert.True(t, bestRouteCost <= maxCost)
    assert.True(t, bestRouteCost == 12 || bestRouteCost == 17)

    if bestRouteCost == 12 {
        assert.Equal(t, []string{"losangeles.b", "b", "chicago.b"}, bestRouteRelays)
    }

    if bestRouteCost == 17 {
        assert.Equal(t, []string{"losangeles.b", "b", "chicago.a"}, bestRouteRelays)
    }
}

func TestGetBestRoute_Initial_NoRoute(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles.a", "10.0.0.1")
    env.AddRelay("losangeles.b", "10.0.0.2")
    env.AddRelay("chicago.a", "10.0.0.3")
    env.AddRelay("chicago.b", "10.0.0.4")
    env.AddRelay("a", "10.0.0.5")
    env.AddRelay("b", "10.0.0.6")

    env.SetCost("losangeles.a", "chicago.a", 100)
    env.SetCost("losangeles.a", "chicago.b", 150)
    env.SetCost("losangeles.a", "a", 10)

    env.SetCost("a", "chicago.a", 50)
    env.SetCost("a", "chicago.b", 20)
    env.SetCost("a", "b", 10)

    env.SetCost("b", "chicago.a", 10)
    env.SetCost("b", "chicago.b", 5)

    env.SetCost("losangeles.b", "chicago.a", 75)
    env.SetCost("losangeles.b", "chicago.b", 110)
    env.SetCost("losangeles.b", "a", 10)
    env.SetCost("losangeles.b", "b", 5)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
    sourceRelayCosts := []int32{5, 2}

    destRelayNames := []string{"chicago.a", "chicago.b"}

    maxCost := int32(1)

    bestRouteCost, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

    assert.True(t, bestRouteCost == 0)
    assert.Equal(t, 0, len(bestRouteRelays))
}

func TestGetBestRoute_Initial_NegativeMaxCost(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles.a", "10.0.0.1")
    env.AddRelay("losangeles.b", "10.0.0.2")
    env.AddRelay("chicago.a", "10.0.0.3")
    env.AddRelay("chicago.b", "10.0.0.4")
    env.AddRelay("a", "10.0.0.5")
    env.AddRelay("b", "10.0.0.6")

    env.SetCost("losangeles.a", "chicago.a", 100)
    env.SetCost("losangeles.a", "chicago.b", 150)
    env.SetCost("losangeles.a", "a", 10)

    env.SetCost("a", "chicago.a", 50)
    env.SetCost("a", "chicago.b", 20)
    env.SetCost("a", "b", 10)

    env.SetCost("b", "chicago.a", 10)
    env.SetCost("b", "chicago.b", 5)

    env.SetCost("losangeles.b", "chicago.a", 75)
    env.SetCost("losangeles.b", "chicago.b", 110)
    env.SetCost("losangeles.b", "a", 10)
    env.SetCost("losangeles.b", "b", 5)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
    sourceRelayCosts := []int32{5, 2}

    destRelayNames := []string{"chicago.a", "chicago.b"}

    maxCost := int32(-1)

    bestRouteCost, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

    assert.True(t, bestRouteCost == 0)
    assert.Equal(t, 0, len(bestRouteRelays))
}

func TestGetBestRoute_Update_Simple(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelayNames := []string{"losangeles"}
    sourceRelayCosts := []int32{10}

    destRelayNames := []string{"chicago"}

    maxCost := int32(1000)

    costThreshold := int32(5)

    currentRoute := []string{"losangeles", "a", "b", "chicago"}

    bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, costThreshold, currentRoute)

    assert.Equal(t, int32(40), bestRouteCost)
    assert.Equal(t, []string{"losangeles", "a", "b", "chicago"}, bestRouteRelays)
}

func TestGetBestRoute_Update_BetterRoute(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 1)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelayNames := []string{"losangeles"}
    sourceRelayCosts := []int32{1}

    destRelayNames := []string{"chicago"}

    maxCost := int32(5)

    costThreshold := int32(5)

    currentRoute := []string{"losangeles", "a", "b", "chicago"}

    bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, costThreshold, currentRoute)

    assert.Equal(t, int32(2), bestRouteCost)
    assert.Equal(t, []string{"losangeles", "chicago"}, bestRouteRelays)
}

func TestGetBestRoute_Update_NoRoute(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelayNames := []string{"losangeles"}
    sourceRelayCosts := []int32{1}

    destRelayNames := []string{"chicago"}

    maxCost := int32(5)

    costThreshold := int32(5)

    currentRoute := []string{"losangeles", "a", "b", "chicago"}

    bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, costThreshold, currentRoute)

    assert.Equal(t, int32(0), bestRouteCost)
    assert.Equal(t, []string{}, bestRouteRelays)
}

func TestGetBestRoute_Update_NegativeMaxCost(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    sourceRelayNames := []string{"losangeles"}
    sourceRelayCosts := []int32{1}

    destRelayNames := []string{"chicago"}

    maxCost := int32(-1)

    costThreshold := int32(5)

    currentRoute := []string{"losangeles", "a", "b", "chicago"}

    bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, costThreshold, currentRoute)

    assert.Equal(t, int32(0), bestRouteCost)
    assert.Equal(t, []string{}, bestRouteRelays)
}

func TestTakeNetworkNext_ReduceLatency_Simple(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    directLatency := int32(50)
    directPacketLoss := float32(0.0)

    sourceRelays := []int32{0}
    sourceRelayCosts := []int32{10}

    destRelays := []int32{1}

    routeCost := int32(0)
    routeNumRelays := int32(0)
    routeRelays := [MaxRelaysPerRoute]int32{}

    routeShader := NewRouteShader()
    routeState := RouteState{}
    customer := CustomerConfig{}
    internal := NewInternalConfig()

    routeState.UserID = 100

    result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, &customer, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

    assert.True(t, result)

    expectedRouteState := RouteState{}
    expectedRouteState.UserID = 100
    expectedRouteState.Next = true
    expectedRouteState.ReduceLatency = true

    assert.Equal(t, expectedRouteState, routeState)
}

func TestTakeNetworkNext_ReduceLatency_LatencyIsAcceptable(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    directLatency := int32(50)
    directPacketLoss := float32(0.0)

    sourceRelays := []int32{0}
    sourceRelayCosts := []int32{10}

    destRelays := []int32{1}

    routeCost := int32(0)
    routeNumRelays := int32(0)
    routeRelays := [MaxRelaysPerRoute]int32{}

    routeShader := NewRouteShader()
    routeState := RouteState{}
    customer := CustomerConfig{}
    internal := NewInternalConfig()

    routeShader.AcceptableLatency = 50

    routeState.UserID = 100

    result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, &customer, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

    assert.False(t, result)

    expectedRouteState := RouteState{}
    expectedRouteState.UserID = 100

    assert.Equal(t, expectedRouteState, routeState)
}

func TestTakeNetworkNext_ReduceLatency_NotEnoughReduction(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")
    env.AddRelay("a", "10.0.0.3")
    env.AddRelay("b", "10.0.0.4")

    env.SetCost("losangeles", "chicago", 100)
    env.SetCost("losangeles", "a", 10)
    env.SetCost("a", "chicago", 50)
    env.SetCost("a", "b", 10)
    env.SetCost("b", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    directLatency := int32(50)
    directPacketLoss := float32(0.0)

    sourceRelays := []int32{0}
    sourceRelayCosts := []int32{10}

    destRelays := []int32{1}

    routeCost := int32(0)
    routeNumRelays := int32(0)
    routeRelays := [MaxRelaysPerRoute]int32{}

    routeShader := NewRouteShader()
    routeState := RouteState{}
    customer := CustomerConfig{}
    internal := NewInternalConfig()

    routeShader.LatencyThreshold = 20

    routeState.UserID = 100

    result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, &customer, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

    assert.False(t, result)

    expectedRouteState := RouteState{}
    expectedRouteState.UserID = 100

    assert.Equal(t, expectedRouteState, routeState)
}

// todo: all the boring things, eg. don't take network next if disabled, AB test, selection %, banned etc.

func TestTakeNetworkNext_ReducePacketLoss_Simple(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")

    env.SetCost("losangeles", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    directLatency := int32(20)
    directPacketLoss := float32(5.0)

    sourceRelays := []int32{0}
    sourceRelayCosts := []int32{10}

    destRelays := []int32{1}

    routeCost := int32(0)
    routeNumRelays := int32(0)
    routeRelays := [MaxRelaysPerRoute]int32{}

    routeShader := NewRouteShader()
    routeState := RouteState{}
    customer := CustomerConfig{}
    internal := NewInternalConfig()

    routeState.UserID = 100

    result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, &customer, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

    assert.True(t, result)

    expectedRouteState := RouteState{}
    expectedRouteState.UserID = 100
    expectedRouteState.Next = true
    expectedRouteState.ReducePacketLoss = true

    assert.Equal(t, expectedRouteState, routeState)
}

func TestTakeNetworkNext_ReducePacketLoss_TradeLatency(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")

    env.SetCost("losangeles", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    directLatency := int32(10)
    directPacketLoss := float32(5.0)

    sourceRelays := []int32{0}
    sourceRelayCosts := []int32{10}

    destRelays := []int32{1}

    routeCost := int32(0)
    routeNumRelays := int32(0)
    routeRelays := [MaxRelaysPerRoute]int32{}

    routeShader := NewRouteShader()
    routeState := RouteState{}
    customer := CustomerConfig{}
    internal := NewInternalConfig()

    routeState.UserID = 100

    result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, &customer, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

    assert.True(t, result)

    expectedRouteState := RouteState{}
    expectedRouteState.UserID = 100
    expectedRouteState.Next = true
    expectedRouteState.ReducePacketLoss = true

    assert.Equal(t, expectedRouteState, routeState)
}

func TestTakeNetworkNext_ReducePacketLoss_DontTradeTooMuchLatency(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")

    env.SetCost("losangeles", "chicago", 100)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    directLatency := int32(10)
    directPacketLoss := float32(5.0)

    sourceRelays := []int32{0}
    sourceRelayCosts := []int32{10}

    destRelays := []int32{1}

    routeCost := int32(0)
    routeNumRelays := int32(0)
    routeRelays := [MaxRelaysPerRoute]int32{}

    routeShader := NewRouteShader()
    routeState := RouteState{}
    customer := CustomerConfig{}
    internal := NewInternalConfig()

    routeState.UserID = 100

    result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, &customer, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

    assert.False(t, result)

    expectedRouteState := RouteState{}
    expectedRouteState.UserID = 100

    assert.Equal(t, expectedRouteState, routeState)
}

func TestTakeNetworkNext_ReducePacketLoss_ReducePacketLossAndLatency(t *testing.T) {

    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")

    env.SetCost("losangeles", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    directLatency := int32(100)
    directPacketLoss := float32(5.0)

    sourceRelays := []int32{0}
    sourceRelayCosts := []int32{10}

    destRelays := []int32{1}

    routeCost := int32(0)
    routeNumRelays := int32(0)
    routeRelays := [MaxRelaysPerRoute]int32{}

    routeShader := NewRouteShader()
    routeState := RouteState{}
    customer := CustomerConfig{}
    internal := NewInternalConfig()

    routeState.UserID = 100

    result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, &customer, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

    assert.True(t, result)

    expectedRouteState := RouteState{}
    expectedRouteState.UserID = 100
    expectedRouteState.Next = true
    expectedRouteState.ReduceLatency = true
    expectedRouteState.ReducePacketLoss = true

    assert.Equal(t, expectedRouteState, routeState)
}

// todo: all the complex stuff like pro mode, multipath, etc.

func TestStayOnNetworkNext_ReduceLatency_Simple(t *testing.T) {

    // test that we stay on network next if we are still reducing latency

    /*
    t.Parallel()

    env := NewTestEnvironment()

    env.AddRelay("losangeles", "10.0.0.1")
    env.AddRelay("chicago", "10.0.0.2")

    env.SetCost("losangeles", "chicago", 10)

    costMatrix, numRelays := env.GetCostMatrix()

    relayDatacenters := env.GetRelayDatacenters()

    routeMatrix := Optimize(numRelays, costMatrix, 5, relayDatacenters)

    directLatency := int32(20)
    directPacketLoss := float32(5.0)

    sourceRelays := []int32{0}
    sourceRelayCosts := []int32{10}

    destRelays := []int32{1}

    routeCost := int32(0)
    routeNumRelays := int32(0)
    routeRelays := [MaxRelaysPerRoute]int32{}

    routeShader := NewRouteShader()
    routeState := RouteState{}
    customer := CustomerConfig{}
    internal := NewInternalConfig()

    routeState.UserID = 100

    result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, &customer, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

    assert.True(t, result)

    expectedRouteState := RouteState{}
    expectedRouteState.UserID = 100
    expectedRouteState.Next = true
    expectedRouteState.ReducePacketLoss = true

    assert.Equal(t, expectedRouteState, routeState)
    */
}

// todo: test that we can make rtt slightly worse, but not get kicked off

// todo: test rtt veto

// todo: test that we will change to a better route if one is available

// todo: if our current routes goes away, test that we change to another route

// todo: if all routes go away, test that we get kicked off network next with "no route" and "veto" set

// -----------------------------------------------

// todo: test that we can make rtt moderately worse, but not get kicked off

// todo: test rtt veto (PL)

// -----------------------------------------------

// todo: test various multipath things

// todo: test multipath overload

// todo: test all the boring stuff, eg. network next disabled, selection % etc.
