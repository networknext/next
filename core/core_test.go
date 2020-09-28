package core

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

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
	assert.Equal(t, 0, TriMatrixIndex(0, 1))
	assert.Equal(t, 1, TriMatrixIndex(0, 2))
	assert.Equal(t, 2, TriMatrixIndex(1, 2))
	assert.Equal(t, 3, TriMatrixIndex(0, 3))
	assert.Equal(t, 4, TriMatrixIndex(1, 3))
	assert.Equal(t, 5, TriMatrixIndex(2, 3))
	assert.Equal(t, 0, TriMatrixIndex(1, 0))
	assert.Equal(t, 1, TriMatrixIndex(2, 0))
	assert.Equal(t, 2, TriMatrixIndex(2, 1))
	assert.Equal(t, 3, TriMatrixIndex(3, 0))
	assert.Equal(t, 4, TriMatrixIndex(3, 1))
	assert.Equal(t, 5, TriMatrixIndex(3, 2))
}

func CheckNilAddress(t *testing.T) {
	var address *net.UDPAddr
	buffer := make([]uint8, NEXT_ADDRESS_BYTES)
	WriteAddress(buffer, address)
	readAddress := ReadAddress(buffer)
	assert.True(t, readAddress == nil)
}

func CheckIPv4Address(t *testing.T, addressString string, expected string) {
	address := ParseAddress(addressString)
	buffer := make([]uint8, NEXT_ADDRESS_BYTES)
	WriteAddress(buffer, address)
	readAddress := ReadAddress(buffer)
	readAddressString := readAddress.String()
	assert.Equal(t, expected, readAddressString)
}

func CheckIPv6Address(t *testing.T, addressString string, expected string) {
	address := ParseAddress(addressString)
	buffer := make([]uint8, NEXT_ADDRESS_BYTES)
	WriteAddress(buffer, address)
	readAddress := ReadAddress(buffer)
	assert.Equal(t, readAddress.IP, address.IP)
	assert.Equal(t, readAddress.Port, address.Port)
}

func TestAddress(t *testing.T) {
	CheckNilAddress(t)
	CheckIPv4Address(t, "127.0.0.1", "127.0.0.1:0")
	CheckIPv4Address(t, "127.0.0.1:40000", "127.0.0.1:40000")
	CheckIPv4Address(t, "1.2.3.4:50000", "1.2.3.4:50000")
	CheckIPv6Address(t, "[::C0A8:1]:80", "[::C0A8:1]:80")
	CheckIPv6Address(t, "[::1]:80", "[::1]:80")
}

func TestRouteManager(t *testing.T) {

	t.Parallel()

	routeManager := RouteManager{}
	routeManager.RelayDatacenter = make([]uint64, 256)
	for i := range routeManager.RelayDatacenter {
		routeManager.RelayDatacenter[i] = uint64(i)
	}
	routeManager.RelayDatacenter[255] = 254

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

	routeManager.AddRoute(200, 4, 4, 5, 6) // verify loops get filtered out
	assert.Equal(t, 2, routeManager.NumRoutes)

	routeManager.AddRoute(200, 4, 5, 254, 255) // verify routes with multiple relays in same datacenter get filtered out
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

func Analyze(numRelays int, routes []RouteEntry) []int {

	buckets := make([]int, 8)

	for i := 0; i < numRelays; i++ {
		for j := 0; j < numRelays; j++ {
			if j < i {
				abFlatIndex := TriMatrixIndex(i, j)
				if routes[abFlatIndex].DirectCost > 0 {
					improvement := routes[abFlatIndex].DirectCost - routes[abFlatIndex].RouteCost[0]
					if improvement == 0 {
						buckets[1]++
					} else if improvement <= 10 {
						buckets[2]++
					} else if improvement <= 20 {
						buckets[3]++
					} else if improvement <= 30 {
						buckets[4]++
					} else if improvement <= 40 {
						buckets[5]++
					} else if improvement <= 50 {
						buckets[6]++
					} else {
						buckets[7]++
					}
				} else {
					if routes[abFlatIndex].NumRoutes > 0 {
						buckets[0]++
					} else {
						buckets[1]++
					}
				}
			}
		}
	}

	return buckets

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
			index := TriMatrixIndex(i, j)
			cost[index] = int32(costValues[i+j*numRelays])
		}
	}

	costThreshold := int32(5)

	relayDatacenters := make([]uint64, 1024)
	for i := range relayDatacenters {
		relayDatacenters[i] = uint64(i)
	}

	numSegments := numRelays

	routes := Optimize(numRelays, numSegments, cost, costThreshold, relayDatacenters)

	buckets := Analyze(numRelays, routes)

	// t.Log(fmt.Sprintf("buckets = %v\n", buckets))

	expectedBuckets := []int{17815, 15021, 3748, 3390, 1589, 846, 514, 1628}

	assert.Equal(t, expectedBuckets, buckets)

	for index := 0; index < entryCount; index++ {
		go func(i int) {
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
		}(index)
	}
}

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

func (env *TestEnvironment) AddRelay(relayName string, relayAddress string) {
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
	relayDatacenters := make([]uint64, len(env.relays))
	for i := range relayDatacenters {
		relayDatacenters[i] = uint64(i)
	}
	return relayDatacenters
}

func (env *TestEnvironment) GetRelayIdToIndex() map[uint64]int32 {
	relayIdToIndex := make(map[uint64]int32)
	for i := range env.relayArray {
		relayHash := RelayHash64(env.relayArray[i].name)
		relayIdToIndex[relayHash] = int32(i)
	}
	return relayIdToIndex
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

func (env *TestEnvironment) GetBestRouteCost(routeMatrix []RouteEntry, sourceRelays []string, sourceRelayCost []int32, destRelays []string) int32 {
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

func (env *TestEnvironment) GetCurrentRouteCost(routeMatrix []RouteEntry, routeRelays []string, sourceRelays []string, sourceRelayCost []int32, destRelays []string) int32 {
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

func (env *TestEnvironment) GetBestRoutes(routeMatrix []RouteEntry, sourceRelays []string, sourceRelayCost []int32, destRelays []string, maxCost int32) []TestRouteData {
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

func (env *TestEnvironment) GetRandomBestRoute(routeMatrix []RouteEntry, sourceRelays []string, sourceRelayCost []int32, destRelays []string, maxCost int32) *TestRouteData {
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
	var bestRouteCost int32
	var bestRouteNumRelays int32
	var bestRouteRelays [MaxRelaysPerRoute]int32
	GetRandomBestRoute(routeMatrix, sourceRelayIndex, sourceRelayCost, destRelayIndex, maxCost, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays)
	if bestRouteNumRelays == 0 {
		return nil
	}
	var route TestRouteData
	route.cost = bestRouteCost
	route.relays = make([]string, bestRouteNumRelays)
	for j := 0; j < int(bestRouteNumRelays); j++ {
		relayIndex := bestRouteRelays[j]
		route.relays[j] = env.relayArray[relayIndex].name
	}
	return &route
}

func (env *TestEnvironment) ReframeRouteHash(route []uint64) (int32, [MaxRelaysPerRoute]int32) {
	relayIdToIndex := make(map[uint64]int32)
	for _, v := range env.relays {
		id := RelayHash64(v.name)
		relayIdToIndex[id] = int32(v.index)
	}
	reframedRoute := [MaxRelaysPerRoute]int32{}
	result := ReframeRoute(relayIdToIndex, route, &reframedRoute)
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

func (env *TestEnvironment) GetBestRoute_Initial(routeMatrix []RouteEntry, sourceRelayNames []string, sourceRelayCost []int32, destRelayNames []string, maxCost int32) (int32, []string) {

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

func (env *TestEnvironment) GetBestRoute_Update(routeMatrix []RouteEntry, sourceRelayNames []string, sourceRelayCost []int32, destRelayNames []string, maxCost int32, costThreshold int32, currentRouteRelayNames []string) (int32, []string) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	sourceIndex := env.GetRelayIndex("losangeles")
	destIndex := env.GetRelayIndex("chicago")

	assert.True(t, sourceIndex != -1)
	assert.True(t, destIndex != -1)

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	// verify the optimizer finds the indirect 3 hop route when the direct route does not exist

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	// verify the optimizer finds the indirect 4 hop route when the direct route does not exist

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	// verify the optimizer finds the indirect 5 hop route when the direct route does not exist

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	// verify the optimizer finds the 3 hop route that is faster than direct

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	// verify the optimizer finds the 4 hop route that is faster than direct

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	// verify the optimizer finds the 5 hop route that is faster than direct

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	// all routes are slower than direct. verify that we only have the direct route between losangeles and chicago

	assert.Equal(t, 1, len(routes))
	if len(routes) == 1 {
		assert.Equal(t, int32(10), routes[0].cost)
		assert.Equal(t, []string{"losangeles", "chicago"}, routes[0].relays)
	}
}

func TestEncrypt(t *testing.T) {

	senderPublicKey := [...]byte{0x6f, 0x58, 0xb4, 0xd7, 0x3d, 0xdc, 0x73, 0x06, 0xb8, 0x97, 0x3d, 0x22, 0x4d, 0xe6, 0xf1, 0xfd, 0x2a, 0xf0, 0x26, 0x7e, 0x8b, 0x1d, 0x93, 0x73, 0xd0, 0x40, 0xa9, 0x8b, 0x86, 0x75, 0xcd, 0x43}
	senderPrivateKey := [...]byte{0x2a, 0xad, 0xd5, 0x43, 0x4e, 0x52, 0xbf, 0x62, 0x0b, 0x76, 0x24, 0x18, 0xe1, 0x26, 0xfb, 0xda, 0x89, 0x95, 0x32, 0xde, 0x1d, 0x39, 0x7f, 0xcd, 0x7b, 0x7a, 0xd5, 0x96, 0x3b, 0x0d, 0x23, 0xe5}

	receiverPublicKey := [...]byte{0x6f, 0x58, 0xb4, 0xd7, 0x3d, 0xdc, 0x73, 0x06, 0xb8, 0x97, 0x3d, 0x22, 0x4d, 0xe6, 0xf1, 0xfd, 0x2a, 0xf0, 0x26, 0x7e, 0x8b, 0x1d, 0x93, 0x73, 0xd0, 0x40, 0xa9, 0x8b, 0x86, 0x75, 0xcd, 0x43}
	receiverPrivateKey := [...]byte{0x2a, 0xad, 0xd5, 0x43, 0x4e, 0x52, 0xbf, 0x62, 0x0b, 0x76, 0x24, 0x18, 0xe1, 0x26, 0xfb, 0xda, 0x89, 0x95, 0x32, 0xde, 0x1d, 0x39, 0x7f, 0xcd, 0x7b, 0x7a, 0xd5, 0x96, 0x3b, 0x0d, 0x23, 0xe5}

	// encrypt random data and verify we can decrypt it

	nonce := make([]byte, NonceBytes)
	RandomBytes(nonce)

	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(data[i])
	}

	encryptedData := make([]byte, 256+MacBytes)

	encryptedBytes := Encrypt(senderPrivateKey[:], receiverPublicKey[:], nonce, encryptedData, len(data))

	assert.Equal(t, 256+MacBytes, encryptedBytes)

	err := Decrypt(senderPublicKey[:], receiverPrivateKey[:], nonce, encryptedData, encryptedBytes)

	assert.NoError(t, err)

	// decryption should fail with garbage data

	garbageData := make([]byte, 256+MacBytes)
	RandomBytes(garbageData[:])

	err = Decrypt(senderPublicKey[:], receiverPrivateKey[:], nonce, garbageData, encryptedBytes)

	assert.Error(t, err)

	// decryption should fail with the wrong receiver private key

	RandomBytes(receiverPrivateKey[:])

	err = Decrypt(senderPublicKey[:], receiverPrivateKey[:], nonce, encryptedData, encryptedBytes)

	assert.Error(t, err)

}

func TestRouteToken(t *testing.T) {

	t.Parallel()

	relayPublicKey := [...]byte{0x71, 0x16, 0xce, 0xc5, 0x16, 0x1a, 0xda, 0xc7, 0xa5, 0x89, 0xb2, 0x51, 0x2b, 0x67, 0x4f, 0x8f, 0x98, 0x21, 0xad, 0x8f, 0xe6, 0x2d, 0x39, 0xca, 0xe3, 0x9b, 0xec, 0xdf, 0x3e, 0xfc, 0x2c, 0x24}
	relayPrivateKey := [...]byte{0xb6, 0x7d, 0x01, 0x0d, 0xaf, 0xba, 0xd1, 0x40, 0x75, 0x99, 0x08, 0x15, 0x0d, 0x3a, 0xce, 0x7b, 0x82, 0x28, 0x01, 0x5f, 0x7d, 0xa0, 0x75, 0xb6, 0xc1, 0x15, 0x56, 0x33, 0xe1, 0x01, 0x99, 0xd6}
	masterPublicKey := [...]byte{0x6f, 0x58, 0xb4, 0xd7, 0x3d, 0xdc, 0x73, 0x06, 0xb8, 0x97, 0x3d, 0x22, 0x4d, 0xe6, 0xf1, 0xfd, 0x2a, 0xf0, 0x26, 0x7e, 0x8b, 0x1d, 0x93, 0x73, 0xd0, 0x40, 0xa9, 0x8b, 0x86, 0x75, 0xcd, 0x43}
	masterPrivateKey := [...]byte{0x2a, 0xad, 0xd5, 0x43, 0x4e, 0x52, 0xbf, 0x62, 0x0b, 0x76, 0x24, 0x18, 0xe1, 0x26, 0xfb, 0xda, 0x89, 0x95, 0x32, 0xde, 0x1d, 0x39, 0x7f, 0xcd, 0x7b, 0x7a, 0xd5, 0x96, 0x3b, 0x0d, 0x23, 0xe5}

	routeToken := RouteToken{}
	routeToken.ExpireTimestamp = uint64(time.Now().Unix() + 10)
	routeToken.SessionId = 0x123131231313131
	routeToken.SessionVersion = 100
	routeToken.KbpsUp = 256
	routeToken.KbpsDown = 512
	routeToken.NextAddress = ParseAddress("127.0.0.1:40000")
	RandomBytes(routeToken.PrivateKey[:])

	// write the token to a buffer and read it back in

	buffer := make([]byte, NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)

	WriteRouteToken(&routeToken, buffer[:])

	var readRouteToken RouteToken
	err := ReadRouteToken(&readRouteToken, buffer)

	assert.NoError(t, err)
	assert.Equal(t, routeToken, readRouteToken)

	// can't read a token if the buffer is too small

	err = ReadRouteToken(&readRouteToken, buffer[:10])

	assert.Error(t, err)

	// write an encrypted route token and read it back

	WriteEncryptedRouteToken(&routeToken, buffer, masterPrivateKey[:], relayPublicKey[:])

	err = ReadEncryptedRouteToken(&readRouteToken, buffer, masterPublicKey[:], relayPrivateKey[:])

	assert.NoError(t, err)
	assert.Equal(t, routeToken, readRouteToken)

	// can't read an encrypted route token if the buffer is too small

	err = ReadEncryptedRouteToken(&readRouteToken, buffer[:10], masterPublicKey[:], relayPrivateKey[:])

	assert.Error(t, err)

	// can't read an encrypted route token if the buffer is garbage

	buffer = make([]byte, NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)

	err = ReadEncryptedRouteToken(&readRouteToken, buffer, masterPublicKey[:], relayPrivateKey[:])

	assert.Error(t, err)
}

func TestRouteTokens(t *testing.T) {

	t.Parallel()

	relayPublicKey := [...]byte{0x71, 0x16, 0xce, 0xc5, 0x16, 0x1a, 0xda, 0xc7, 0xa5, 0x89, 0xb2, 0x51, 0x2b, 0x67, 0x4f, 0x8f, 0x98, 0x21, 0xad, 0x8f, 0xe6, 0x2d, 0x39, 0xca, 0xe3, 0x9b, 0xec, 0xdf, 0x3e, 0xfc, 0x2c, 0x24}
	relayPrivateKey := [...]byte{0xb6, 0x7d, 0x01, 0x0d, 0xaf, 0xba, 0xd1, 0x40, 0x75, 0x99, 0x08, 0x15, 0x0d, 0x3a, 0xce, 0x7b, 0x82, 0x28, 0x01, 0x5f, 0x7d, 0xa0, 0x75, 0xb6, 0xc1, 0x15, 0x56, 0x33, 0xe1, 0x01, 0x99, 0xd6}
	masterPublicKey := [...]byte{0x6f, 0x58, 0xb4, 0xd7, 0x3d, 0xdc, 0x73, 0x06, 0xb8, 0x97, 0x3d, 0x22, 0x4d, 0xe6, 0xf1, 0xfd, 0x2a, 0xf0, 0x26, 0x7e, 0x8b, 0x1d, 0x93, 0x73, 0xd0, 0x40, 0xa9, 0x8b, 0x86, 0x75, 0xcd, 0x43}
	masterPrivateKey := [...]byte{0x2a, 0xad, 0xd5, 0x43, 0x4e, 0x52, 0xbf, 0x62, 0x0b, 0x76, 0x24, 0x18, 0xe1, 0x26, 0xfb, 0xda, 0x89, 0x95, 0x32, 0xde, 0x1d, 0x39, 0x7f, 0xcd, 0x7b, 0x7a, 0xd5, 0x96, 0x3b, 0x0d, 0x23, 0xe5}

	// write a bunch of tokens to a buffer

	addresses := make([]*net.UDPAddr, NEXT_MAX_NODES)
	for i := range addresses {
		addresses[i] = ParseAddress(fmt.Sprintf("127.0.0.1:%d", 40000+i))
	}

	publicKeys := make([][]byte, NEXT_MAX_NODES)
	for i := range publicKeys {
		publicKeys[i] = make([]byte, PublicKeyBytes)
		copy(publicKeys[i], relayPublicKey[:])
	}

	sessionId := uint64(0x123131231313131)
	sessionVersion := byte(100)
	kbpsUp := uint32(256)
	kbpsDown := uint32(256)
	expireTimestamp := uint64(time.Now().Unix() + 10)

	tokenData := make([]byte, NEXT_MAX_NODES*NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)

	WriteRouteTokens(tokenData, expireTimestamp, sessionId, sessionVersion, kbpsUp, kbpsDown, NEXT_MAX_NODES, addresses, publicKeys, masterPrivateKey)

	// read each token back individually and verify the token data matches what was written

	for i := 0; i < NEXT_MAX_NODES; i++ {
		var routeToken RouteToken
		err := ReadEncryptedRouteToken(&routeToken, tokenData[i*NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], masterPublicKey[:], relayPrivateKey[:])
		assert.NoError(t, err)
		assert.Equal(t, sessionId, routeToken.SessionId)
		assert.Equal(t, sessionVersion, routeToken.SessionVersion)
		assert.Equal(t, kbpsUp, routeToken.KbpsUp)
		assert.Equal(t, kbpsDown, routeToken.KbpsDown)
		assert.Equal(t, expireTimestamp, routeToken.ExpireTimestamp)
		if i != NEXT_MAX_NODES-1 {
			assert.Equal(t, addresses[i+1].String(), routeToken.NextAddress.String())
		} else {
			assert.True(t, routeToken.NextAddress == nil)
		}
		assert.Equal(t, publicKeys[i], relayPublicKey[:])
	}
}

func TestContinueToken(t *testing.T) {

	t.Parallel()

	relayPublicKey := [...]byte{0x71, 0x16, 0xce, 0xc5, 0x16, 0x1a, 0xda, 0xc7, 0xa5, 0x89, 0xb2, 0x51, 0x2b, 0x67, 0x4f, 0x8f, 0x98, 0x21, 0xad, 0x8f, 0xe6, 0x2d, 0x39, 0xca, 0xe3, 0x9b, 0xec, 0xdf, 0x3e, 0xfc, 0x2c, 0x24}
	relayPrivateKey := [...]byte{0xb6, 0x7d, 0x01, 0x0d, 0xaf, 0xba, 0xd1, 0x40, 0x75, 0x99, 0x08, 0x15, 0x0d, 0x3a, 0xce, 0x7b, 0x82, 0x28, 0x01, 0x5f, 0x7d, 0xa0, 0x75, 0xb6, 0xc1, 0x15, 0x56, 0x33, 0xe1, 0x01, 0x99, 0xd6}
	masterPublicKey := [...]byte{0x6f, 0x58, 0xb4, 0xd7, 0x3d, 0xdc, 0x73, 0x06, 0xb8, 0x97, 0x3d, 0x22, 0x4d, 0xe6, 0xf1, 0xfd, 0x2a, 0xf0, 0x26, 0x7e, 0x8b, 0x1d, 0x93, 0x73, 0xd0, 0x40, 0xa9, 0x8b, 0x86, 0x75, 0xcd, 0x43}
	masterPrivateKey := [...]byte{0x2a, 0xad, 0xd5, 0x43, 0x4e, 0x52, 0xbf, 0x62, 0x0b, 0x76, 0x24, 0x18, 0xe1, 0x26, 0xfb, 0xda, 0x89, 0x95, 0x32, 0xde, 0x1d, 0x39, 0x7f, 0xcd, 0x7b, 0x7a, 0xd5, 0x96, 0x3b, 0x0d, 0x23, 0xe5}

	// write a continue token and verify we can read it back

	continueToken := ContinueToken{}
	continueToken.ExpireTimestamp = uint64(time.Now().Unix() + 10)
	continueToken.SessionId = 0x123131231313131
	continueToken.SessionVersion = 100

	buffer := make([]byte, NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES)

	WriteContinueToken(&continueToken, buffer[:])

	var readContinueToken ContinueToken

	err := ReadContinueToken(&readContinueToken, buffer)

	assert.NoError(t, err)
	assert.Equal(t, continueToken, readContinueToken)

	// read continue token should fail when the buffer is too small

	err = ReadContinueToken(&readContinueToken, buffer[:10])

	assert.Error(t, err)

	// write an encrypted continue token and verify we can decrypt and read it back

	WriteEncryptedContinueToken(&continueToken, buffer, masterPrivateKey[:], relayPublicKey[:])

	err = ReadEncryptedContinueToken(&continueToken, buffer, masterPublicKey[:], relayPrivateKey[:])

	assert.NoError(t, err)
	assert.Equal(t, continueToken, readContinueToken)

	// read encrypted continue token should fail when buffer is too small

	err = ReadEncryptedContinueToken(&continueToken, buffer[:10], masterPublicKey[:], relayPrivateKey[:])

	assert.Error(t, err)

	// read encrypted continue token should fail on garbage data

	garbageData := make([]byte, NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES)
	RandomBytes(garbageData)

	err = ReadEncryptedContinueToken(&continueToken, garbageData, masterPublicKey[:], relayPrivateKey[:])

	assert.Error(t, err)
}

func TestContinueTokens(t *testing.T) {

	t.Parallel()

	relayPublicKey := [...]byte{0x71, 0x16, 0xce, 0xc5, 0x16, 0x1a, 0xda, 0xc7, 0xa5, 0x89, 0xb2, 0x51, 0x2b, 0x67, 0x4f, 0x8f, 0x98, 0x21, 0xad, 0x8f, 0xe6, 0x2d, 0x39, 0xca, 0xe3, 0x9b, 0xec, 0xdf, 0x3e, 0xfc, 0x2c, 0x24}
	relayPrivateKey := [...]byte{0xb6, 0x7d, 0x01, 0x0d, 0xaf, 0xba, 0xd1, 0x40, 0x75, 0x99, 0x08, 0x15, 0x0d, 0x3a, 0xce, 0x7b, 0x82, 0x28, 0x01, 0x5f, 0x7d, 0xa0, 0x75, 0xb6, 0xc1, 0x15, 0x56, 0x33, 0xe1, 0x01, 0x99, 0xd6}
	masterPublicKey := [...]byte{0x6f, 0x58, 0xb4, 0xd7, 0x3d, 0xdc, 0x73, 0x06, 0xb8, 0x97, 0x3d, 0x22, 0x4d, 0xe6, 0xf1, 0xfd, 0x2a, 0xf0, 0x26, 0x7e, 0x8b, 0x1d, 0x93, 0x73, 0xd0, 0x40, 0xa9, 0x8b, 0x86, 0x75, 0xcd, 0x43}
	masterPrivateKey := [...]byte{0x2a, 0xad, 0xd5, 0x43, 0x4e, 0x52, 0xbf, 0x62, 0x0b, 0x76, 0x24, 0x18, 0xe1, 0x26, 0xfb, 0xda, 0x89, 0x95, 0x32, 0xde, 0x1d, 0x39, 0x7f, 0xcd, 0x7b, 0x7a, 0xd5, 0x96, 0x3b, 0x0d, 0x23, 0xe5}

	// write a bunch of tokens to a buffer

	publicKeys := make([][]byte, NEXT_MAX_NODES)
	for i := range publicKeys {
		publicKeys[i] = make([]byte, PublicKeyBytes)
		copy(publicKeys[i], relayPublicKey[:])
	}

	sessionId := uint64(0x123131231313131)
	sessionVersion := byte(100)
	expireTimestamp := uint64(time.Now().Unix() + 10)

	tokenData := make([]byte, NEXT_MAX_NODES*NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES)

	WriteContinueTokens(tokenData, expireTimestamp, sessionId, sessionVersion, NEXT_MAX_NODES, publicKeys, masterPrivateKey)

	// read each token back individually and verify the token data matches what was written

	for i := 0; i < NEXT_MAX_NODES; i++ {
		var routeToken ContinueToken
		err := ReadEncryptedContinueToken(&routeToken, tokenData[i*NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES:], masterPublicKey[:], relayPrivateKey[:])
		assert.NoError(t, err)
		assert.Equal(t, sessionId, routeToken.SessionId)
		assert.Equal(t, sessionVersion, routeToken.SessionVersion)
		assert.Equal(t, expireTimestamp, routeToken.ExpireTimestamp)
	}
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

	numSegments := 64

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	sourceRelays := []string{"losangeles.a", "losangeles.b", "chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{10, 5, 100, 100}

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	sourceRelays := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{10, 5}

	destRelays := []string{"chicago.a", "chicago.b"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(math.MaxInt32), bestRouteCost)
}

func TestCurrentRouteCost_Simple(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	routeRelays := []string{"losangeles", "a", "b", "chicago"}

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	currentRouteCost := env.GetCurrentRouteCost(routeMatrix, routeRelays, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(40), currentRouteCost)
}

func TestCurrentRouteCost_Reverse(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	routeRelays := []string{"chicago", "b", "a", "losangeles"}

	sourceRelays := []string{"chicago"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"losangeles"}

	currentRouteCost := env.GetCurrentRouteCost(routeMatrix, routeRelays, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(40), currentRouteCost)
}

func TestGetBestRoutes_Simple(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

func TestGetBestRoutes_Reverse(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

func TestGetBestRoutes_Complex(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

func TestGetBestRoutes_NoRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	maxCost := int32(1000)

	bestRoutes := env.GetBestRoutes(routeMatrix, sourceRelays, sourceRelayCosts, destRelays, maxCost)

	assert.Equal(t, 0, len(bestRoutes))
}

func TestGetRandomBestRoute_Simple(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(20)

	bestRoute := env.GetRandomBestRoute(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute != nil)
	assert.True(t, bestRoute.cost > 0)
	assert.True(t, bestRoute.cost <= maxCost)
	assert.True(t, bestRoute.cost == 12 || bestRoute.cost == 17)

	if bestRoute.cost == 12 {
		assert.Equal(t, []string{"losangeles.b", "b", "chicago.b"}, bestRoute.relays)
	}

	if bestRoute.cost == 17 {
		assert.Equal(t, []string{"losangeles.b", "b", "chicago.a"}, bestRoute.relays)
	}
}

func TestGetRandomBestRoute_Reverse(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	sourceRelayNames := []string{"chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"losangeles.a", "losangeles.b"}

	maxCost := int32(17)

	bestRoute := env.GetRandomBestRoute(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute != nil)
	assert.True(t, bestRoute.cost > 0)
	assert.True(t, bestRoute.cost <= maxCost)
	assert.True(t, bestRoute.cost == 12 || bestRoute.cost == 17)

	if bestRoute.cost == 12 {
		assert.Equal(t, []string{"chicago.b", "b", "losangeles.b"}, bestRoute.relays)
	}

	if bestRoute.cost == 17 {
		assert.Equal(t, []string{"chicago.a", "b", "losangeles.b"}, bestRoute.relays)
	}
}

func TestGetRandomBestRoute_NoRoute(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	sourceRelayNames := []string{"chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"losangeles.a", "losangeles.b"}

	maxCost := int32(20)

	bestRoute := env.GetRandomBestRoute(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute == nil)
}

func TestReframeRoute_Simple(t *testing.T) {

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

func TestReframeRoute_RelayNoLongerExists(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles.b", "10.0.0.2")
	env.AddRelay("chicago.a", "10.0.0.3")
	env.AddRelay("chicago.b", "10.0.0.4")
	env.AddRelay("a", "10.0.0.5")
	env.AddRelay("b", "10.0.0.6")

	currentRoute := make([]string, 3)
	currentRoute[0] = "losangeles.a"
	currentRoute[1] = "a"
	currentRoute[2] = "chicago.b"

	numRouteRelays, _ := env.ReframeRoute(currentRoute)

	assert.Equal(t, int32(0), numRouteRelays)
}

func TestReframeRelays(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles.a", "10.0.0.1")
	env.AddRelay("losangeles.b", "10.0.0.2")
	env.AddRelay("chicago.a", "10.0.0.3")
	env.AddRelay("chicago.b", "10.0.0.4")
	env.AddRelay("a", "10.0.0.5")
	env.AddRelay("b", "10.0.0.6")

	relayIdToIndex := env.GetRelayIdToIndex()

	assert.Equal(t, int32(0), relayIdToIndex[RelayHash64("losangeles.a")])
	assert.Equal(t, int32(1), relayIdToIndex[RelayHash64("losangeles.b")])
	assert.Equal(t, int32(2), relayIdToIndex[RelayHash64("chicago.a")])
	assert.Equal(t, int32(3), relayIdToIndex[RelayHash64("chicago.b")])
	assert.Equal(t, int32(4), relayIdToIndex[RelayHash64("a")])
	assert.Equal(t, int32(5), relayIdToIndex[RelayHash64("b")])

	sourceRelayIds := []uint64{
		RelayHash64("losangeles.a"),
		RelayHash64("losangeles.b"),
		RelayHash64("a"),
		RelayHash64("b"),
		RelayHash64("idontexist"),
	}

	sourceRelayLatency := []int32{
		10,
		0,
		100,
		-1,
		1,
	}

	sourceRelayPacketLoss := []float32{
		0.0,
		0.0,
		100.0,
		0.0,
		0.0,
	}

	destRelayIds := []uint64{
		RelayHash64("idontexist"),
		RelayHash64("chicago.a"),
		RelayHash64("chicago.b"),
	}

	numSourceRelays := int32(0)
	sourceRelays := [32]int32{}

	numDestRelays := int32(0)
	destRelays := [32]int32{}

	ReframeRelays(relayIdToIndex, sourceRelayIds, sourceRelayLatency, sourceRelayPacketLoss, destRelayIds, &numSourceRelays, sourceRelays[:], &numDestRelays, destRelays[:])

	assert.Equal(t, int32(1), numSourceRelays)
	assert.Equal(t, int32(0), sourceRelays[0])

	assert.Equal(t, int32(2), numDestRelays)
	assert.Equal(t, int32(2), destRelays[0])
	assert.Equal(t, int32(3), destRelays[1])
}

func TestEarlyOutDirect(t *testing.T) {

	routeShader := NewRouteShader()
	routeState := RouteState{}
	assert.False(t, EarlyOutDirect(&routeShader, &routeState))

	routeState = RouteState{Veto: true}
	assert.True(t, EarlyOutDirect(&routeShader, &routeState))

	routeState = RouteState{Banned: true}
	assert.True(t, EarlyOutDirect(&routeShader, &routeState))

	routeState = RouteState{Disabled: true}
	assert.True(t, EarlyOutDirect(&routeShader, &routeState))

	routeState = RouteState{NotSelected: true}
	assert.True(t, EarlyOutDirect(&routeShader, &routeState))

	routeState = RouteState{B: true}
	assert.True(t, EarlyOutDirect(&routeShader, &routeState))

	routeShader = NewRouteShader()
	routeShader.DisableNetworkNext = true
	routeState = RouteState{}
	assert.True(t, EarlyOutDirect(&routeShader, &routeState))
	assert.True(t, routeState.Disabled)

	routeShader = NewRouteShader()
	routeShader.SelectionPercent = 0
	routeState = RouteState{}
	assert.True(t, EarlyOutDirect(&routeShader, &routeState))
	assert.True(t, routeState.NotSelected)

	routeShader = NewRouteShader()
	routeShader.SelectionPercent = 0
	routeState = RouteState{}
	assert.True(t, EarlyOutDirect(&routeShader, &routeState))
	assert.True(t, routeState.NotSelected)

	routeShader = NewRouteShader()
	routeShader.ABTest = true
	routeState = RouteState{}
	routeState.UserID = 0
	assert.False(t, EarlyOutDirect(&routeShader, &routeState))
	assert.True(t, routeState.ABTest)
	assert.True(t, routeState.A)
	assert.False(t, routeState.B)

	routeShader = NewRouteShader()
	routeShader.ABTest = true
	routeState = RouteState{}
	routeState.UserID = 1
	assert.True(t, EarlyOutDirect(&routeShader, &routeState))
	assert.True(t, routeState.ABTest)
	assert.False(t, routeState.A)
	assert.True(t, routeState.B)

	routeShader = NewRouteShader()
	routeShader.BannedUsers[1000] = true
	routeState = RouteState{}
	assert.False(t, EarlyOutDirect(&routeShader, &routeState))

	routeShader = NewRouteShader()
	routeShader.BannedUsers[1000] = true
	routeState = RouteState{}
	routeState.UserID = 1000
	assert.True(t, EarlyOutDirect(&routeShader, &routeState))
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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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

// -------------------------------------------------------------------------------

func TestTakeNetworkNext_EarlyOutDirect_Veto(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeState.UserID = 100
	routeState.Veto = true

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestTakeNetworkNext_EarlyOutDirect_Banned(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeState.UserID = 100
	routeState.Banned = true

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Banned = true

	assert.Equal(t, expectedRouteState, routeState)

}

func TestTakeNetworkNext_EarlyOutDirect_Disabled(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeState.UserID = 100
	routeState.Disabled = true

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Disabled = true

	assert.Equal(t, expectedRouteState, routeState)

}

func TestTakeNetworkNext_EarlyOutDirect_NotSelected(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeState.UserID = 100
	routeState.NotSelected = true

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.NotSelected = true

	assert.Equal(t, expectedRouteState, routeState)

}

func TestTakeNetworkNext_EarlyOutDirect_B(t *testing.T) {

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeState.UserID = 100
	routeState.B = true

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.B = true

	assert.Equal(t, expectedRouteState, routeState)

}

// -------------------------------------------------------------------------------

func TestTakeNetworkNext_ReduceLatency_Simple(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeShader.AcceptableLatency = 50

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeShader.LatencyThreshold = 20

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100

	assert.Equal(t, expectedRouteState, routeState)
}

// -----------------------------------------------------------------------------

func TestTakeNetworkNext_ReducePacketLoss_Simple(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

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

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

// -----------------------------------------------------------------------------

func TestTakeNetworkNext_ReduceLatency_Multipath(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeShader.Multipath = true

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestTakeNetworkNext_ReducePacketLoss_Multipath(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeShader.Multipath = true

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestTakeNetworkNext_ReducePacketLossAndLatency_Multipath(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeShader.Multipath = true

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestTakeNetworkNext_ReducePacketLossAndLatency_MultipathVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

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
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeShader.Multipath = true

	routeState.UserID = 100

	multipathVetoUsers[routeState.UserID] = true

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = false
	expectedRouteState.ReduceLatency = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

// -----------------------------------------------------------------------------

func TestTakeNetworkNext_ProMode(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(20)
	directPacketLoss := float32(0.0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeShader.ProMode = true

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ProMode = true
	expectedRouteState.Multipath = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestTakeNetworkNext_ProMode_MultipathVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(20)
	directPacketLoss := float32(0.0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()

	routeShader.ProMode = true

	routeState.UserID = 100

	multipathVetoUsers[routeState.UserID] = true

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100

	assert.Equal(t, expectedRouteState, routeState)
}

// -----------------------------------------------------------------------------

func TestStayOnNetworkNext_EarlyOut_Veto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true
	routeState.Veto = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_EarlyOut_Banned(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true
	routeState.Banned = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Banned = true
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, routeState)
}

// -----------------------------------------------------------------------------

func TestStayOnNetworkNext_ReduceLatency_Simple(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_ReduceLatency_SlightlyWorse(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(15)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_ReduceLatency_RTTVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(5)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.ReduceLatency = true
	expectedRouteState.LatencyWorse = true
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_ReduceLatency_NoRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(5)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.ReduceLatency = true
	expectedRouteState.NoRoute = true
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_ReduceLatency_SwitchToNewRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")

	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
	assert.Equal(t, int32(12), routeCost)
	assert.Equal(t, int32(3), routeNumRelays)
}

func TestStayOnNetworkNext_ReduceLatency_SwitchToBetterRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")

	env.SetCost("losangeles", "chicago", 20)
	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
	assert.Equal(t, int32(12), routeCost)
	assert.Equal(t, int32(3), routeNumRelays)
}

// -----------------------------------------------------------------------------

func TestStayOnNetworkNext_ReducePacketLoss_LatencyTradeOff(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(10)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReducePacketLoss = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_ReducePacketLoss_RTTVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(5)

	nextLatency := int32(30)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReducePacketLoss = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.LatencyWorse = true
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_ReducePacketLoss_NoRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(10)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReducePacketLoss = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

// -----------------------------------------------------------------------------

func TestStayOnNetworkNext_MultipathOverload(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(550)

	nextLatency := int32(30)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.Multipath = true
	routeState.ReducePacketLoss = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.MultipathOverload = true
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_Multipath_LatencyTradeOff(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 20)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(10)

	nextLatency := int32(30)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.Multipath = true
	routeState.ReducePacketLoss = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_Multipath_RTTVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 20)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(10)

	nextLatency := int32(50)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()

	routeState.Next = true
	routeState.UserID = 100
	routeState.Multipath = true
	routeState.ReducePacketLoss = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.LatencyWorse = true
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, routeState)
}

// -----------------------------------------------------------------------------

func TestStayOnNetworkNext_TryBeforeYouBuy_NewRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")

	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	directPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	multipathVetoUsers := map[uint64]bool{}
	internal := NewInternalConfig()
	internal.TryBeforeYouBuy = true

	routeState.UserID = 100

	result := MakeRouteDecision_TakeNetworkNext(routeMatrix, &routeShader, &routeState, multipathVetoUsers, &internal, directLatency, directPacketLoss, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.CommitPending = true

	assert.Equal(t, expectedRouteState, routeState)
	assert.Equal(t, int32(12), routeCost)
	assert.Equal(t, int32(3), routeNumRelays)
}

func TestStayOnNetworkNext_TryBeforeYouBuy_SwitchToNewRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")

	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()
	internal.TryBeforeYouBuy = true

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true
	routeState.CommitPending = true
	routeState.CommitCounter = 2

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.CommitPending = true

	assert.Equal(t, expectedRouteState, routeState)
	assert.Equal(t, int32(12), routeCost)
	assert.Equal(t, int32(3), routeNumRelays)
}

func TestStayOnNetworkNext_TryBeforeYouBuy_Multipath(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")

	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()
	internal.TryBeforeYouBuy = true

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true
	routeState.Multipath = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Multipath = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
	assert.Equal(t, int32(12), routeCost)
	assert.Equal(t, int32(3), routeNumRelays)
}

func TestStayOnNetworkNext_TryBeforeYouBuy_Improvement(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()
	internal.TryBeforeYouBuy = true

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true
	routeState.CommitPending = true
	routeState.CommitCounter = 2

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_TryBeforeYouBuy_KeepWatching(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(35)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()
	internal.TryBeforeYouBuy = true

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true
	routeState.CommitPending = true
	routeState.CommitCounter = 2

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.CommitPending = true
	expectedRouteState.CommitCounter = 3

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_TryBeforeYouBuy_CommitVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(35)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()
	internal.TryBeforeYouBuy = true

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true
	routeState.CommitPending = true
	routeState.CommitCounter = 3

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Veto = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.CommitVeto = true

	assert.Equal(t, expectedRouteState, routeState)
}

func TestStayOnNetworkNext_TryBeforeYouBuy_AlreadyCommitted(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	directLatency := int32(30)

	nextLatency := int32(20)

	directPacketLoss := float32(0)

	nextPacketLoss := float32(0)

	sourceRelays := []int32{0}
	sourceRelayCosts := []int32{10}

	destRelays := []int32{1}

	routeCost := int32(0)
	routeNumRelays := int32(0)
	routeRelays := [MaxRelaysPerRoute]int32{}

	routeShader := NewRouteShader()
	routeState := RouteState{}
	internal := NewInternalConfig()
	internal.TryBeforeYouBuy = true

	routeState.Next = true
	routeState.UserID = 100
	routeState.ReduceLatency = true
	routeState.Committed = true

	currentRouteNumRelays := int32(2)
	currentRouteRelays := [MaxRelaysPerRoute]int32{0, 1}

	result := MakeRouteDecision_StayOnNetworkNext(routeMatrix, &routeShader, &routeState, &internal, directLatency, nextLatency, directPacketLoss, nextPacketLoss, currentRouteNumRelays, currentRouteRelays, sourceRelays, sourceRelayCosts, destRelays, &routeCost, &routeNumRelays, routeRelays[:])

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, routeState)
}

// -----------------------------------------------------------------------------
