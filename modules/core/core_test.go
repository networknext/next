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
	"math/rand"

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

	// verify loops get filtered out

	routeManager.AddRoute(200, 4, 4, 5, 6)
	assert.Equal(t, 2, routeManager.NumRoutes)

	// verify routes with multiple relays in same datacenter get filtered out

	routeManager.AddRoute(200, 4, 5, 254, 255)
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

	// fill up lots of extra routes to get to max routes

	numFillers := MaxRoutesPerEntry - routeManager.NumRoutes

	for i := 0; i < numFillers; i++ {
		routeManager.AddRoute(int32(1000+i), int32(100+i), int32(100+i+1), int32(100+i+2))
		assert.Equal(t, 8+i+1, routeManager.NumRoutes)
	}

	assert.Equal(t, MaxRoutesPerEntry, routeManager.NumRoutes)

	// make sure we can't add worse routes once we are at max routes

	routeManager.AddRoute(10000, 12, 13, 14)
	assert.Equal(t, routeManager.NumRoutes, MaxRoutesPerEntry)
	for i := 0; i < routeManager.NumRoutes; i++ {
		assert.True(t, routeManager.RouteCost[i] != 10000)
	}

	// make sure we can add better routes while at max routes

	routeManager.AddRoute(177, 13, 14, 15, 16, 17)
	assert.Equal(t, routeManager.NumRoutes, MaxRoutesPerEntry)
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

	// check all the best routes are sorted and they have correct data

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

type TestRelayData struct {
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

func (env *TestEnvironment) GetRelayIds() []uint64 {
	relayIds := make([]uint64, len(env.relayArray))
	for i := range env.relayArray {
		relayIds[i] = RelayHash64(env.relayArray[i].name)
	}
	return relayIds
}

func (env *TestEnvironment) GetRelayNames() []string {
	relayNames := make([]string, len(env.relayArray))
	for i := range env.relayArray {
		relayNames[i] = env.relayArray[i].name
	}
	return relayNames
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

func (env *TestEnvironment) GetBestRouteCost(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, sourceRelays []string, sourceRelayCost []int32, destRelays []string) int32 {
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
	return GetBestRouteCost(routeMatrix, fullRelaySet, sourceRelayIndex, sourceRelayCost, destRelayIndex)
}

func (env *TestEnvironment) RouteExists(routeMatrix []RouteEntry, routeRelays []string) bool {
	routeRelayIndex := [MaxRelaysPerRoute]int32{}
	for i := range routeRelays {
		routeRelayIndex[i] = int32(env.GetRelayIndex(routeRelays[i]))
		if routeRelayIndex[i] == -1 {
			panic("bad route relay name")
		}
	}
	debug := ""
	return RouteExists(routeMatrix, int32(len(routeRelays)), routeRelayIndex, &debug)
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
	debug := ""
	return GetCurrentRouteCost(routeMatrix, int32(len(routeRelays)), routeRelayIndex, sourceRelayIndex, sourceRelayCost, destRelayIndex, &debug)
}

func (env *TestEnvironment) GetBestRoutes(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, sourceRelays []string, sourceRelayCost []int32, destRelays []string, maxCost int32) []TestRouteData {
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
	routeDiversity := int32(0)
	bestRoutes := make([]BestRoute, 1024)
	GetBestRoutes(routeMatrix, fullRelaySet, sourceRelayIndex, sourceRelayCost, destRelayIndex, maxCost, bestRoutes, &numBestRoutes, &routeDiversity)
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

func (env *TestEnvironment) GetRandomBestRoute(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, sourceRelays []string, sourceRelayCost []int32, destRelays []string, maxCost int32) *TestRouteData {
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
	debug := ""
	selectThreshold := int32(2)
	GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelayIndex, sourceRelayCost, destRelayIndex, maxCost, selectThreshold, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, &debug)
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
	routeState := RouteState{}
	reframedRoute := [MaxRelaysPerRoute]int32{}
	if ReframeRoute(&routeState, relayIdToIndex, route, &reframedRoute) {
		return int32(len(route)), reframedRoute
	}
	return 0, reframedRoute
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

func (env *TestEnvironment) GetBestRoute_Initial(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, sourceRelayNames []string, sourceRelayCost []int32, destRelayNames []string, maxCost int32) (int32, int32, []string) {

	sourceRelays, destRelays := env.ReframeRelays(sourceRelayNames, destRelayNames)

	bestRouteCost := int32(0)
	bestRouteNumRelays := int32(0)
	bestRouteRelays := [MaxRelaysPerRoute]int32{}

	debug := ""
	selectThreshold := int32(2)
	hasRoute, routeDiversity := GetBestRoute_Initial(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, &debug)
	if !hasRoute {
		return 0, 0, []string{}
	}

	bestRouteRelayNames := make([]string, bestRouteNumRelays)

	for i := 0; i < int(bestRouteNumRelays); i++ {
		routeData := env.relayArray[bestRouteRelays[i]]
		bestRouteRelayNames[i] = routeData.name
	}

	return bestRouteCost, routeDiversity, bestRouteRelayNames
}

func (env *TestEnvironment) GetBestRoute_Update(routeMatrix []RouteEntry, fullRelaySet map[int32]bool, sourceRelayNames []string, sourceRelayCost []int32, destRelayNames []string, maxCost int32, selectThreshold int32, switchThreshold int32, currentRouteRelayNames []string) (int32, []string) {

	sourceRelays, destRelays := env.ReframeRelays(sourceRelayNames, destRelayNames)

	currentRouteNumRelays, currentRouteRelays := env.ReframeRoute(currentRouteRelayNames)
	if currentRouteNumRelays == 0 {
		panic("current route has no relays")
	}

	bestRouteCost := int32(0)
	bestRouteNumRelays := int32(0)
	bestRouteRelays := [MaxRelaysPerRoute]int32{}

	debug := ""
	GetBestRoute_Update(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, switchThreshold, currentRouteNumRelays, currentRouteRelays, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, &debug)

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

	fullRelaySet := make(map[int32]bool)

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(40+CostBias), bestRouteCost)
}

func TestBestRouteCostSimple_RelayFull(t *testing.T) {

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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("b"))] = true

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(70+CostBias), bestRouteCost)
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

	fullRelaySet := make(map[int32]bool)

	sourceRelays := []string{"losangeles.a", "losangeles.b", "chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{10, 5, 100, 100}

	destRelays := []string{"chicago.a", "chicago.b"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(15+CostBias), bestRouteCost)
}

func TestBestRouteCostComplex_RelayFull(t *testing.T) {

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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("b"))] = true

	sourceRelays := []string{"losangeles.a", "losangeles.b", "chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{10, 5, 100, 100}

	destRelays := []string{"chicago.a", "chicago.b"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(35+CostBias), bestRouteCost)
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
	fullRelaySet := make(map[int32]bool)

	sourceRelays := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{10, 5}

	destRelays := []string{"chicago.a", "chicago.b"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(math.MaxInt32), bestRouteCost)
}

func TestBestRouteCostNoRoute_RelayFull_Source(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles.a", "10.0.0.1")
	env.AddRelay("losangeles.b", "10.0.0.2")
	env.AddRelay("chicago.a", "10.0.0.3")
	env.AddRelay("chicago.b", "10.0.0.4")
	env.AddRelay("a", "10.0.0.5")
	env.AddRelay("b", "10.0.0.6")

	env.SetCost("losangeles.a", "a", 10)
	env.SetCost("a", "b", 10)
	env.SetCost("b", "chicago.a", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("losangeles.a"))] = true

	sourceRelays := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{10, 5}

	destRelays := []string{"chicago.a", "chicago.b"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(math.MaxInt32), bestRouteCost)
}

func TestBestRouteCostNoRoute_RelayFull_Hop(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles.a", "10.0.0.1")
	env.AddRelay("losangeles.b", "10.0.0.2")
	env.AddRelay("chicago.a", "10.0.0.3")
	env.AddRelay("chicago.b", "10.0.0.4")
	env.AddRelay("a", "10.0.0.5")
	env.AddRelay("b", "10.0.0.6")

	env.SetCost("losangeles.a", "a", 10)
	env.SetCost("a", "b", 10)
	env.SetCost("b", "chicago.a", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("b"))] = true

	sourceRelays := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{10, 5}

	destRelays := []string{"chicago.a", "chicago.b"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(math.MaxInt32), bestRouteCost)
}

func TestBestRouteCostNoRoute_RelayFull_Dest(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles.a", "10.0.0.1")
	env.AddRelay("losangeles.b", "10.0.0.2")
	env.AddRelay("chicago.a", "10.0.0.3")
	env.AddRelay("chicago.b", "10.0.0.4")
	env.AddRelay("a", "10.0.0.5")
	env.AddRelay("b", "10.0.0.6")

	env.SetCost("losangeles.a", "a", 10)
	env.SetCost("a", "b", 10)
	env.SetCost("b", "chicago.a", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("chicago.a"))] = true

	sourceRelays := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{10, 5}

	destRelays := []string{"chicago.a", "chicago.b"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays)

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

	currentRouteExists := env.RouteExists(routeMatrix, routeRelays)

	assert.Equal(t, true, currentRouteExists)

	currentRouteCost := env.GetCurrentRouteCost(routeMatrix, routeRelays, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(40+CostBias), currentRouteCost)
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

	currentRouteExists := env.RouteExists(routeMatrix, routeRelays)

	assert.Equal(t, true, currentRouteExists)

	currentRouteCost := env.GetCurrentRouteCost(routeMatrix, routeRelays, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(40+CostBias), currentRouteCost)
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

	fullRelaySet := make(map[int32]bool)

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	maxCost := int32(1000)

	bestRoutes := env.GetBestRoutes(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays, maxCost)

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

func TestGetBestRoutes_Simple_RelayFull(t *testing.T) {

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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("b"))] = true

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	maxCost := int32(1000)

	bestRoutes := env.GetBestRoutes(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays, maxCost)

	sort.Slice(bestRoutes, func(i int, j int) bool { return bestRoutes[i].cost < bestRoutes[j].cost })

	expectedBestRoutes := make([]TestRouteData, 2)

	expectedBestRoutes[0].cost = 70
	expectedBestRoutes[0].relays = []string{"losangeles", "a", "chicago"}

	expectedBestRoutes[1].cost = 110
	expectedBestRoutes[1].relays = []string{"losangeles", "chicago"}

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

	fullRelaySet := make(map[int32]bool)

	sourceRelays := []string{"chicago"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"losangeles"}

	maxCost := int32(1000)

	bestRoutes := env.GetBestRoutes(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays, maxCost)

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

func TestGetBestRoutes_Reverse_RelayFull(t *testing.T) {

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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("b"))] = true

	sourceRelays := []string{"chicago"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"losangeles"}

	maxCost := int32(1000)

	bestRoutes := env.GetBestRoutes(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays, maxCost)

	sort.Slice(bestRoutes, func(i int, j int) bool { return bestRoutes[i].cost < bestRoutes[j].cost })

	expectedBestRoutes := make([]TestRouteData, 2)

	expectedBestRoutes[0].cost = 70
	expectedBestRoutes[0].relays = []string{"chicago", "a", "losangeles"}

	expectedBestRoutes[1].cost = 110
	expectedBestRoutes[1].relays = []string{"chicago", "losangeles"}

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

	fullRelaySet := make(map[int32]bool)

	sourceRelays := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 3}

	destRelays := []string{"chicago.a", "chicago.b"}

	maxCost := int32(30)

	bestRoutes := env.GetBestRoutes(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays, maxCost)

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

func TestGetBestRoutes_Complex_RelayFull(t *testing.T) {

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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("b"))] = true

	sourceRelays := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 3}

	destRelays := []string{"chicago.a", "chicago.b"}

	maxCost := int32(30)

	bestRoutes := env.GetBestRoutes(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays, maxCost)

	sort.Slice(bestRoutes, func(i int, j int) bool { return bestRoutes[i].cost < bestRoutes[j].cost })

	expectedBestRoutes := make([]TestRouteData, 2)

	expectedBestRoutes[0].cost = 6
	expectedBestRoutes[0].relays = []string{"losangeles.a", "chicago.a"}

	expectedBestRoutes[1].cost = 24
	expectedBestRoutes[1].relays = []string{"losangeles.b", "a", "losangeles.a", "chicago.a"}

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

	fullRelaySet := make(map[int32]bool)

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	maxCost := int32(1000)

	bestRoutes := env.GetBestRoutes(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays, maxCost)

	assert.Equal(t, 0, len(bestRoutes))
}

func TestGetBestRoutes_NoRoute_RelayFull_Source(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "a", 10)
	env.SetCost("a", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("losangeles"))] = true

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	maxCost := int32(1000)

	bestRoutes := env.GetBestRoutes(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays, maxCost)

	assert.Equal(t, 0, len(bestRoutes))
}

func TestGetBestRoutes_NoRoute_RelayFull_Hop(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "a", 10)
	env.SetCost("a", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("a"))] = true

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	maxCost := int32(1000)

	bestRoutes := env.GetBestRoutes(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays, maxCost)

	assert.Equal(t, 0, len(bestRoutes))
}

func TestGetBestRoutes_NoRoute_RelayFull_Dest(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "a", 10)
	env.SetCost("a", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("chicago"))] = true

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	maxCost := int32(1000)

	bestRoutes := env.GetBestRoutes(routeMatrix, fullRelaySet, sourceRelays, sourceRelayCosts, destRelays, maxCost)

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

	fullRelaySet := make(map[int32]bool)

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(20)

	bestRoute := env.GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute != nil)
	assert.True(t, bestRoute.cost > 0)
	assert.True(t, bestRoute.cost <= maxCost)
	assert.True(t, bestRoute.cost == 12+CostBias || bestRoute.cost == 17+CostBias)

	if bestRoute.cost == 12 {
		assert.Equal(t, []string{"losangeles.b", "b", "chicago.b"}, bestRoute.relays)
	}

	if bestRoute.cost == 17 {
		assert.Equal(t, []string{"losangeles.b", "b", "chicago.a"}, bestRoute.relays)
	}
}

func TestGetRandomBestRoute_Simple_RelayFull(t *testing.T) {

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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("chicago.b"))] = true

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(20)

	bestRoute := env.GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute != nil)
	assert.True(t, bestRoute.cost > 0)
	assert.True(t, bestRoute.cost <= maxCost)
	assert.True(t, bestRoute.cost == 17+CostBias)

	assert.Equal(t, []string{"losangeles.b", "b", "chicago.a"}, bestRoute.relays)
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

	fullRelaySet := make(map[int32]bool)

	sourceRelayNames := []string{"chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"losangeles.a", "losangeles.b"}

	maxCost := int32(17)

	bestRoute := env.GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute != nil)
	assert.True(t, bestRoute.cost > 0)
	assert.True(t, bestRoute.cost <= maxCost)
	assert.True(t, bestRoute.cost == 12+CostBias || bestRoute.cost == 17+CostBias)

	if bestRoute.cost == 12 {
		assert.Equal(t, []string{"chicago.b", "b", "losangeles.b"}, bestRoute.relays)
	}

	if bestRoute.cost == 17 {
		assert.Equal(t, []string{"chicago.a", "b", "losangeles.b"}, bestRoute.relays)
	}
}

func TestGetRandomBestRoute_Reverse_RelayFull(t *testing.T) {

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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("chicago.a"))] = true

	sourceRelayNames := []string{"chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"losangeles.a", "losangeles.b"}

	maxCost := int32(17)

	bestRoute := env.GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute != nil)
	assert.True(t, bestRoute.cost > 0)
	assert.True(t, bestRoute.cost <= maxCost)
	assert.True(t, bestRoute.cost == 12+CostBias)

	assert.Equal(t, []string{"chicago.b", "b", "losangeles.b"}, bestRoute.relays)
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

	fullRelaySet := make(map[int32]bool)

	sourceRelayNames := []string{"chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"losangeles.a", "losangeles.b"}

	maxCost := int32(20)

	bestRoute := env.GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute == nil)
}

func TestGetRandomBestRoute_NoRoute_RelayFull_Source(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles.a", "10.0.0.1")
	env.AddRelay("losangeles.b", "10.0.0.2")
	env.AddRelay("chicago.a", "10.0.0.3")
	env.AddRelay("chicago.b", "10.0.0.4")
	env.AddRelay("a", "10.0.0.5")
	env.AddRelay("b", "10.0.0.6")

	env.SetCost("chicago.a", "a", 10)
	env.SetCost("a", "losangeles.a", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("chicago.a"))] = true

	sourceRelayNames := []string{"chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"losangeles.a", "losangeles.b"}

	maxCost := int32(20)

	bestRoute := env.GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute == nil)
}

func TestGetRandomBestRoute_NoRoute_RelayFull_Hop(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles.a", "10.0.0.1")
	env.AddRelay("losangeles.b", "10.0.0.2")
	env.AddRelay("chicago.a", "10.0.0.3")
	env.AddRelay("chicago.b", "10.0.0.4")
	env.AddRelay("a", "10.0.0.5")
	env.AddRelay("b", "10.0.0.6")

	env.SetCost("chicago.a", "a", 10)
	env.SetCost("a", "losangeles.a", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("a"))] = true

	sourceRelayNames := []string{"chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"losangeles.a", "losangeles.b"}

	maxCost := int32(20)

	bestRoute := env.GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute == nil)
}

func TestGetRandomBestRoute_NoRoute_RelayFull_Dest(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles.a", "10.0.0.1")
	env.AddRelay("losangeles.b", "10.0.0.2")
	env.AddRelay("chicago.a", "10.0.0.3")
	env.AddRelay("chicago.b", "10.0.0.4")
	env.AddRelay("a", "10.0.0.5")
	env.AddRelay("b", "10.0.0.6")

	env.SetCost("chicago.a", "a", 10)
	env.SetCost("a", "losangeles.a", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("losangeles.a"))] = true

	sourceRelayNames := []string{"chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"losangeles.a", "losangeles.b"}

	maxCost := int32(20)

	bestRoute := env.GetRandomBestRoute(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

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

func TestEarlyOutDirect(t *testing.T) {

	routeShader := NewRouteShader()
	routeState := RouteState{}
	assert.False(t, EarlyOutDirect(&routeShader, &routeState))

	routeState = RouteState{Veto: true}
	assert.True(t, EarlyOutDirect(&routeShader, &routeState))

	routeState = RouteState{LocationVeto: true}
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

	fullRelaySet := make(map[int32]bool)

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{5}

	destRelayNames := []string{"chicago"}

	maxCost := int32(40)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.Equal(t, int32(35+CostBias), bestRouteCost)
	assert.Equal(t, int32(1), routeDiversity)
	assert.Equal(t, []string{"losangeles", "a", "b", "chicago"}, bestRouteRelays)
}

func TestGetBestRoute_Initial_Simple_RelayFull(t *testing.T) {

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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("b"))] = true

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{5}

	destRelayNames := []string{"chicago"}

	maxCost := int32(75)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.Equal(t, int32(65+CostBias), bestRouteCost)
	assert.Equal(t, int32(1), routeDiversity)
	assert.Equal(t, []string{"losangeles", "a", "chicago"}, bestRouteRelays)
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

	fullRelaySet := make(map[int32]bool)

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(20)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRouteCost > 0)
	assert.True(t, bestRouteCost <= maxCost)
	assert.True(t, bestRouteCost == 12+CostBias || bestRouteCost == 17+CostBias)

	if bestRouteCost == 12 {
		assert.Equal(t, []string{"losangeles.b", "b", "chicago.b"}, bestRouteRelays)
	}

	if bestRouteCost == 17 {
		assert.Equal(t, []string{"losangeles.b", "b", "chicago.a"}, bestRouteRelays)
	}

	assert.Equal(t, int32(1), routeDiversity)
}

func TestGetBestRoute_Initial_Complex_RelayFull(t *testing.T) {

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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("chicago.b"))] = true

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(20)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRouteCost > 0)
	assert.True(t, bestRouteCost <= maxCost)
	assert.True(t, bestRouteCost == 17+CostBias)

	assert.Equal(t, []string{"losangeles.b", "b", "chicago.a"}, bestRouteRelays)

	assert.Equal(t, int32(1), routeDiversity)
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

	fullRelaySet := make(map[int32]bool)

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(1)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRouteCost == 0)
	assert.True(t, routeDiversity == int32(0))
	assert.Equal(t, 0, len(bestRouteRelays))
}

func TestGetBestRoute_Initial_NoRoute_RelayFull_Source(t *testing.T) {

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

	env.SetCost("a", "chicago.a", 10)
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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("losangeles.a"))] = true
	fullRelaySet[int32(env.GetRelayIndex("losangeles.b"))] = true

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(30)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRouteCost == 0)
	assert.True(t, routeDiversity == int32(0))
	assert.Equal(t, 0, len(bestRouteRelays))
}

func TestGetBestRoute_Initial_NoRoute_RelayFull_Hop(t *testing.T) {

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

	env.SetCost("a", "chicago.a", 10)
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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("a"))] = true
	fullRelaySet[int32(env.GetRelayIndex("b"))] = true

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(30)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRouteCost == 0)
	assert.True(t, routeDiversity == int32(0))
	assert.Equal(t, 0, len(bestRouteRelays))
}

func TestGetBestRoute_Initial_NoRoute_RelayFull_Dest(t *testing.T) {

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

	env.SetCost("a", "chicago.a", 10)
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

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("chicago.a"))] = true
	fullRelaySet[int32(env.GetRelayIndex("chicago.b"))] = true

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(30)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRouteCost == 0)
	assert.True(t, routeDiversity == int32(0))
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

	fullRelaySet := make(map[int32]bool)

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(-1)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.Equal(t, int32(0), bestRouteCost)
	assert.Equal(t, int32(0), routeDiversity)
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

	fullRelaySet := make(map[int32]bool)

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelayNames := []string{"chicago"}

	maxCost := int32(1000)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

	assert.Equal(t, int32(40+CostBias), bestRouteCost)
	assert.Equal(t, []string{"losangeles", "a", "b", "chicago"}, bestRouteRelays)
}

func TestGetBestRoute_Update_Simple_RelayFull(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "chicago", 100)
	env.SetCost("losangeles", "a", 10)
	env.SetCost("losangeles", "b", 1)
	env.SetCost("a", "chicago", 50)
	env.SetCost("a", "b", 10)
	env.SetCost("b", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("b"))] = true

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelayNames := []string{"chicago"}

	maxCost := int32(1000)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

	assert.Equal(t, int32(40+CostBias), bestRouteCost)
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

	fullRelaySet := make(map[int32]bool)

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{1}

	destRelayNames := []string{"chicago"}

	maxCost := int32(5)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

	assert.Equal(t, int32(2+CostBias), bestRouteCost)
	assert.Equal(t, []string{"losangeles", "chicago"}, bestRouteRelays)
}

func TestGetBestRoute_Update_BetterRoute_RelayFull(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "chicago", 3)
	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 2)
	env.SetCost("a", "b", 10)
	env.SetCost("b", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("a"))] = true

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{1}

	destRelayNames := []string{"chicago"}

	maxCost := int32(15)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

	assert.Equal(t, int32(4+CostBias), bestRouteCost)
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

	fullRelaySet := make(map[int32]bool)

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{1}

	destRelayNames := []string{"chicago"}

	maxCost := int32(5)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

	assert.Equal(t, int32(0), bestRouteCost)
	assert.Equal(t, []string{}, bestRouteRelays)
}

func TestGetBestRoute_Update_NoRoute_RelayFull_Source(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("losangeles"))] = true

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{1}

	destRelayNames := []string{"chicago"}

	maxCost := int32(10)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

	assert.Equal(t, int32(0), bestRouteCost)
	assert.Equal(t, []string{}, bestRouteRelays)
}

func TestGetBestRoute_Update_NoRoute_RelayFull_Hop(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("a"))] = true

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{1}

	destRelayNames := []string{"chicago"}

	maxCost := int32(10)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

	assert.Equal(t, int32(0), bestRouteCost)
	assert.Equal(t, []string{}, bestRouteRelays)
}

func TestGetBestRoute_Update_NoRoute_RelayFull_Dest(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := Optimize(numRelays, numSegments, costMatrix, 5, relayDatacenters)

	fullRelaySet := make(map[int32]bool)
	fullRelaySet[int32(env.GetRelayIndex("chicago"))] = true

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{1}

	destRelayNames := []string{"chicago"}

	maxCost := int32(10)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

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

	fullRelaySet := make(map[int32]bool)

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{1}

	destRelayNames := []string{"chicago"}

	maxCost := int32(-1)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, fullRelaySet, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

	assert.Equal(t, int32(0), bestRouteCost)
	assert.Equal(t, []string{}, bestRouteRelays)
}

// -------------------------------------------------------------------------------

type TestData struct {
	numRelays        int
	relayNames       []string
	relayDatacenters []uint64
	costMatrix       []int32
	routeMatrix      []RouteEntry
	fullRelaySet     map[int32]bool

	directLatency    int32
	directPacketLoss float32

	sourceRelays     []int32
	sourceRelayCosts []int32

	destRelays []int32

	routeCost      int32
	routeNumRelays int32
	routeRelays    [MaxRelaysPerRoute]int32

	internal           InternalConfig
	routeShader        RouteShader
	routeState         RouteState
	multipathVetoUsers map[uint64]bool

	debug string

	routeDiversity int32

	nextLatency           int32
	nextPacketLoss        float32
	predictedLatency      int32
	currentRouteNumRelays int32
	currentRouteRelays    [MaxRelaysPerRoute]int32

	sliceNumber int32

	realPacketLoss float32
}

func NewTestData(env *TestEnvironment) *TestData {

	test := &TestData{}

	test.costMatrix, test.numRelays = env.GetCostMatrix()

	test.fullRelaySet = map[int32]bool{}

	test.relayNames = env.GetRelayNames()

	test.relayDatacenters = env.GetRelayDatacenters()

	numSegments := test.numRelays
	costThreshold := int32(5)
	test.routeMatrix = Optimize(test.numRelays, numSegments, test.costMatrix, costThreshold, test.relayDatacenters)

	test.routeShader = NewRouteShader()

	test.multipathVetoUsers = map[uint64]bool{}

	test.internal = NewInternalConfig()

	return test
}

func (test *TestData) TakeNetworkNext() bool {
	return MakeRouteDecision_TakeNetworkNext(test.routeMatrix,
		test.fullRelaySet,
		&test.routeShader,
		&test.routeState,
		test.multipathVetoUsers,
		&test.internal,
		test.directLatency,
		test.directPacketLoss,
		test.sourceRelays,
		test.sourceRelayCosts,
		test.destRelays,
		&test.routeCost,
		&test.routeNumRelays,
		test.routeRelays[:],
		&test.routeDiversity,
		&test.debug,
		test.sliceNumber,
	)
}

func (test *TestData) StayOnNetworkNext() (bool, bool) {
	return MakeRouteDecision_StayOnNetworkNext(test.routeMatrix,
		test.fullRelaySet,
		test.relayNames,
		&test.routeShader,
		&test.routeState,
		&test.internal,
		test.directLatency,
		test.nextLatency,
		test.predictedLatency,
		test.directPacketLoss,
		test.nextPacketLoss,
		test.currentRouteNumRelays,
		test.currentRouteRelays,
		test.sourceRelays,
		test.sourceRelayCosts,
		test.destRelays,
		&test.routeCost,
		&test.routeNumRelays,
		test.routeRelays[:],
		&test.debug,
	)
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

	test := NewTestData(env)

	test.directLatency = 50

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100
	test.routeState.Veto = true

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
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

	test := NewTestData(env)

	test.directLatency = 50

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100
	test.routeState.Banned = true

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Banned = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
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

	test := NewTestData(env)

	test.directLatency = 50

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100
	test.routeState.Disabled = true

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Disabled = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
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

	test := NewTestData(env)

	test.directLatency = 50

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100
	test.routeState.NotSelected = true

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.NotSelected = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
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

	test := NewTestData(env)

	test.directLatency = 50

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100
	test.routeState.B = true

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.B = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
}

// -------------------------------------------------------------------------------

func TestTakeNetworkNext_ReduceLatency_Simple(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = 50

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReduceLatency_RouteDiversity(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles.a", "10.0.0.1")
	env.AddRelay("losangeles.b", "10.0.0.2")
	env.AddRelay("losangeles.c", "10.0.0.3")
	env.AddRelay("losangeles.d", "10.0.0.4")
	env.AddRelay("losangeles.e", "10.0.0.5")
	env.AddRelay("chicago", "10.0.0.6")

	env.SetCost("losangeles.a", "chicago", 10)
	env.SetCost("losangeles.b", "chicago", 10)
	env.SetCost("losangeles.c", "chicago", 10)
	env.SetCost("losangeles.d", "chicago", 10)
	env.SetCost("losangeles.e", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = 50

	test.sourceRelays = []int32{0, 1, 2, 3, 4}
	test.sourceRelayCosts = []int32{10, 10, 10, 10, 10}

	test.destRelays = []int32{5}

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(5), test.routeDiversity)
}

func TestTakeNetworkNext_ReduceLatency_LackOfDiversity(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles.a", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles.a", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = 50

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.internal.RouteDiversity = 5

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.LackOfDiversity = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
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

	test := NewTestData(env)

	test.directLatency = 50

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.AcceptableLatency = 50

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
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

	test := NewTestData(env)

	test.directLatency = 50

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.LatencyThreshold = 20

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
}

func TestTakeNetworkNext_ReduceLatency_MaxRTT(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 500)

	test := NewTestData(env)

	test.directLatency = int32(1000)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

// -----------------------------------------------------------------------------

func TestTakeNetworkNext_ReducePacketLoss_Simple(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(20)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.AcceptableLatency = 100
	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLoss_TradeLatency(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(10)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.AcceptableLatency = 25
	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLoss_DontTradeTooMuchLatency(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 100)

	test := NewTestData(env)

	test.directLatency = int32(10)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLoss_ReducePacketLossAndLatency(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(100)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLoss_MaxRTT(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 500)

	test := NewTestData(env)

	test.directLatency = int32(1000)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.AcceptableLatency = 100

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLoss_BeforeMinSliceNumber(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(20)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.AcceptableLatency = 100
	test.routeState.UserID = 100

	test.internal.ReducePacketLossMinSliceNumber = 3
	test.sliceNumber = 0

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.ReducePacketLoss = false
	expectedRouteState.Committed = false

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLoss_EqualMinSliceNumber(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(20)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.AcceptableLatency = 100
	test.routeState.UserID = 100

	test.internal.ReducePacketLossMinSliceNumber = 3
	test.sliceNumber = 3

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLoss_AfterMinSliceNumber(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(20)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.AcceptableLatency = 100
	test.routeState.UserID = 100

	test.internal.ReducePacketLossMinSliceNumber = 3
	test.sliceNumber = 4

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLoss_PLBelowSustained(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	// Won't go next because of latency
	test.directLatency = int32(20)
	test.routeShader.AcceptableLatency = 100

	// Won't go next because of packet Loss
	test.directPacketLoss = float32(5.0)
	test.routeShader.AcceptablePacketLoss = float32(20)

	// Will go next after 3 slices of sustained packet loss
	test.routeShader.PacketLossSustained = float32(2.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(1), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(2), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.True(t, result)
	assert.Equal(t, int32(3), test.routeState.PLSustainedCounter)
}

func TestTakeNetworkNext_ReducePacketLoss_PLEqualSustained(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	// Won't go next because of latency
	test.directLatency = int32(20)
	test.routeShader.AcceptableLatency = 100

	// Won't go next because of packet Loss
	test.directPacketLoss = float32(5.0)
	test.routeShader.AcceptablePacketLoss = float32(20)

	// Will go next after 3 slices of sustained packet loss
	test.routeShader.PacketLossSustained = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(1), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(2), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.True(t, result)
	assert.Equal(t, int32(3), test.routeState.PLSustainedCounter)
}

func TestTakeNetworkNext_ReducePacketLoss_PLAboveSustained(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	// Won't go next because of latency
	test.directLatency = int32(20)
	test.routeShader.AcceptableLatency = 100

	// Won't go next because of packet Loss
	test.directPacketLoss = float32(5.0)
	test.routeShader.AcceptablePacketLoss = float32(20)

	// Won't go next after 3 slices of sustained packet loss
	test.routeShader.PacketLossSustained = float32(10.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(0), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(0), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(0), test.routeState.PLSustainedCounter)
}

func TestTakeNetworkNext_ReducePacketLoss_SustainedCount_ResetCount(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	// Won't go next because of latency
	test.directLatency = int32(20)
	test.routeShader.AcceptableLatency = 100

	// Won't go next because of packet Loss
	test.directPacketLoss = float32(5.0)
	test.routeShader.AcceptablePacketLoss = float32(20)

	// Will go next after 3 slices of sustained packet loss
	test.routeShader.PacketLossSustained = float32(2.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(1), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(2), test.routeState.PLSustainedCounter)

	test.directPacketLoss = 1

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(0), test.routeState.PLSustainedCounter)
}

func TestTakeNetworkNext_ReducePacketLoss_SustainedCount_Mix_Next(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	// Won't go next because of latency
	test.directLatency = int32(20)
	test.routeShader.AcceptableLatency = 100

	// Won't go next because of packet Loss
	test.directPacketLoss = float32(5.0)
	test.routeShader.AcceptablePacketLoss = float32(20)

	// Will go next after 3 slices of sustained packet loss
	test.routeShader.PacketLossSustained = float32(2.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(1), test.routeState.PLSustainedCounter)

	test.directPacketLoss = 1

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(0), test.routeState.PLSustainedCounter)

	test.directPacketLoss = 5

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(1), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(2), test.routeState.PLSustainedCounter)

	test.directPacketLoss = 1

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(0), test.routeState.PLSustainedCounter)

	test.directPacketLoss = 5

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(1), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, int32(2), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.True(t, result)
	assert.Equal(t, int32(3), test.routeState.PLSustainedCounter)
}

// -----------------------------------------------------------------------------

func TestTakeNetworkNext_ReduceLatency_Multipath(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(50)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.internal.MultipathThreshold = 100

	test.routeShader.Multipath = true

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReduceLatency_MultipathThreshold(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(100)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.internal.MultipathThreshold = 20

	test.routeShader.Multipath = true

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = false
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLoss_Multipath(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(20)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.Multipath = true

	test.routeShader.AcceptableLatency = 25
	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLossAndLatency_Multipath(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(100)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.internal.MultipathThreshold = 100

	test.routeShader.Multipath = true

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLossAndLatency_MultipathVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(100)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.Multipath = true

	test.routeState.UserID = 100

	test.multipathVetoUsers[test.routeState.UserID] = true

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = false
	expectedRouteState.MultipathRestricted = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
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

	test := NewTestData(env)

	test.directLatency = int32(20)
	test.directPacketLoss = float32(0.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.ProMode = true

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ProMode = true
	expectedRouteState.Multipath = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ProMode_MultipathVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(20)
	test.directPacketLoss = float32(0.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.ProMode = true

	test.routeState.UserID = 100

	test.multipathVetoUsers[test.routeState.UserID] = true

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.MultipathRestricted = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
}

// -----------------------------------------------------------------------------

func TestStayOnNetworkNext_EarlyOut_Veto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(20)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Veto = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Veto = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_EarlyOut_Banned(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(20)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = true
	test.routeState.Banned = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true
	expectedRouteState.Banned = true
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

// -----------------------------------------------------------------------------

func TestStayOnNetworkNext_ReduceLatency_Simple(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(20)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ReduceLatency_SlightlyWorse(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(15)

	test.nextLatency = int32(20)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ReduceLatency_RTTVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(5)

	test.nextLatency = int32(20)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeCost = int32(0)
	test.routeNumRelays = int32(0)
	test.routeRelays = [MaxRelaysPerRoute]int32{}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.ReduceLatency = true
	expectedRouteState.LatencyWorse = true
	expectedRouteState.Committed = true
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ReduceLatency_NoRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(5)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true
	expectedRouteState.NoRoute = true
	expectedRouteState.Veto = true
	expectedRouteState.RouteLost = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ReduceLatency_MispredictVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(20)

	test.predictedLatency = int32(1)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	// first slice mispredicting is fine

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true
	expectedRouteState.MispredictCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)

	// first slice mispredicting is fine

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true
	expectedRouteState.MispredictCounter = 2

	assert.Equal(t, expectedRouteState, test.routeState)

	// third slice mispredicting is veto

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true
	expectedRouteState.Mispredict = true
	expectedRouteState.Veto = true
	expectedRouteState.MispredictCounter = 3

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ReduceLatency_MispredictRecover(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(20)

	test.predictedLatency = int32(1)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	// first slice mispredicting is fine

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true
	expectedRouteState.MispredictCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)

	// check that we recover when no longer mispredicting

	test.predictedLatency = int32(100)

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true
	expectedRouteState.MispredictCounter = 0

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ReduceLatency_SwitchToNewRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")

	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)
	env.SetCost("losangeles", "chicago", 300)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(20)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.True(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(12+CostBias), test.routeCost)
	assert.Equal(t, int32(3), test.routeNumRelays)
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

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(20)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.True(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(12+CostBias), test.routeCost)
	assert.Equal(t, int32(3), test.routeNumRelays)
}

func TestStayOnNetworkNext_ReduceLatency_MaxRTT(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 500)

	test := NewTestData(env)

	test.directLatency = int32(1000)

	test.nextLatency = int32(20)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.Veto = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true
	expectedRouteState.NextLatencyTooHigh = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

// -----------------------------------------------------------------------------

func TestStayOnNetworkNext_ReducePacketLoss_LatencyTradeOff(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(10)

	test.nextLatency = int32(20)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReducePacketLoss = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ReducePacketLoss_RTTVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(5)

	test.nextLatency = int32(40)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReducePacketLoss = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.LatencyWorse = true
	expectedRouteState.Veto = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ReducePacketLoss_NoRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(10)

	test.nextLatency = int32(20)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReducePacketLoss = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ReducePacketLoss_MaxRTT(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 500)

	test := NewTestData(env)

	test.directLatency = int32(1000)

	test.nextLatency = int32(30)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReducePacketLoss = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.NextLatencyTooHigh = true
	expectedRouteState.Veto = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

// -----------------------------------------------------------------------------

func TestStayOnNetworkNext_MultipathOverload(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(550)

	test.nextLatency = int32(30)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.Multipath = true
	test.routeState.ReducePacketLoss = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.Multipath = true
	expectedRouteState.Committed = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.MultipathOverload = true
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_Multipath_LatencyTradeOff(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 20)

	test := NewTestData(env)

	test.directLatency = int32(10)

	test.nextLatency = int32(30)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.Multipath = true
	test.routeState.ReducePacketLoss = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_Multipath_RTTVeto(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 20)

	test := NewTestData(env)

	test.directLatency = int32(10)

	test.nextLatency = int32(50)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.Multipath = true
	test.routeState.ReducePacketLoss = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	// first latency worse is fine

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true
	expectedRouteState.LatencyWorseCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)

	// second latency worse is fine

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true
	expectedRouteState.LatencyWorseCounter = 2

	assert.Equal(t, expectedRouteState, test.routeState)

	// third latency worse is veto

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = false
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.LatencyWorse = true
	expectedRouteState.Committed = true
	expectedRouteState.Veto = true
	expectedRouteState.LatencyWorseCounter = 3

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_Multipath_RTTVeto_Recover(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 20)

	test := NewTestData(env)

	test.directLatency = int32(10)

	test.nextLatency = int32(50)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.Multipath = true
	test.routeState.ReducePacketLoss = true
	test.routeState.Committed = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	// first latency worse is fine

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true
	expectedRouteState.LatencyWorseCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)

	// now latency is not worse, we should recover

	test.nextLatency = int32(5)

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true
	expectedRouteState.LatencyWorseCounter = 0

	assert.Equal(t, expectedRouteState, test.routeState)
}

// -----------------------------------------------------------------------------

func TestTakeNetworkNext_ForceNext(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 40)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.directPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.ReduceLatency = false

	test.internal.ForceNext = true

	test.routeState.Next = false
	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ForcedNext = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(50+CostBias), test.routeCost)
	assert.Equal(t, int32(2), test.routeNumRelays)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ForceNext_NoRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.routeShader.ReduceLatency = false

	test.internal.ForceNext = true

	test.routeState.Next = false
	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.ForcedNext = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
}

func TestStayOnNetworkNext_ForceNext(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 40)

	test := NewTestData(env)

	test.routeShader.ReduceLatency = false

	test.directLatency = int32(30)

	test.nextLatency = int32(60)

	test.predictedLatency = int32(0)

	test.nextPacketLoss = float32(5)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.ReduceLatency = false

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ForcedNext = true
	test.routeState.Committed = true

	test.internal.ForceNext = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, routeSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, routeSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ForcedNext = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(50+CostBias), test.routeCost)
	assert.Equal(t, int32(2), test.routeNumRelays)
}

func TestStayOnNetworkNext_ForceNext_NoRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(60)

	test.nextPacketLoss = float32(5)

	test.sourceRelays = []int32{}
	test.sourceRelayCosts = []int32{}

	test.destRelays = []int32{}

	test.routeShader.ReduceLatency = false

	test.internal.ForceNext = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ForcedNext = true
	test.routeState.Committed = true

	result, routeSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, routeSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.ForcedNext = true
	expectedRouteState.Veto = true
	expectedRouteState.NoRoute = true
	expectedRouteState.Committed = true
	expectedRouteState.RouteLost = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ForceNext_RouteSwitched(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")

	env.SetCost("losangeles", "chicago", 400)
	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(1)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{1}

	test.destRelays = []int32{1}

	test.routeShader.ReduceLatency = false

	test.internal.ForceNext = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ForcedNext = true
	test.routeState.Committed = true

	result, routeSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.True(t, routeSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ForcedNext = true
	expectedRouteState.Committed = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(3+CostBias), test.routeCost)
	assert.Equal(t, int32(3), test.routeNumRelays)
}

// -----------------------------------------------------------------------------

func TestTakeNetworkNext_Uncommitted(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")

	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.internal.Uncommitted = true

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(12+CostBias), test.routeCost)
	assert.Equal(t, int32(3), test.routeNumRelays)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestStayOnNetworkNext_Uncommitted(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(20)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.internal.Uncommitted = true

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

// -----------------------------------------------------------------------------

func TestTakeNetworkNext_TryBeforeYouBuy(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(50)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100

	test.internal.TryBeforeYouBuy = true

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = false

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestStayOnNetworkNext_TryBeforeYouBuy_Commit(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(20)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = false

	test.internal.TryBeforeYouBuy = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true
	expectedRouteState.CommitCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_TryBeforeYouBuy_LatencyWorse(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(100)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = false

	test.internal.TryBeforeYouBuy = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	// don't commit first slice

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = false
	expectedRouteState.CommitCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)

	// don't commit second slice

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState.Committed = false
	expectedRouteState.CommitCounter = 2

	assert.Equal(t, expectedRouteState, test.routeState)

	// don't commit third slice

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState.Committed = false
	expectedRouteState.CommitCounter = 3

	assert.Equal(t, expectedRouteState, test.routeState)

	// abort fourth slice

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState.Next = false
	expectedRouteState.Veto = true
	expectedRouteState.Committed = false
	expectedRouteState.CommitCounter = 4
	expectedRouteState.CommitVeto = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_TryBeforeYouBuy_PacketLossWorse(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(10)

	test.nextPacketLoss = float32(1)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReduceLatency = true
	test.routeState.Committed = false

	test.internal.TryBeforeYouBuy = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	// first slice don't commit

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = false
	expectedRouteState.CommitCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)

	// second slice don't commit

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState.Committed = false
	expectedRouteState.CommitCounter = 2

	assert.Equal(t, expectedRouteState, test.routeState)

	// third slice don't commit

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState.Committed = false
	expectedRouteState.CommitCounter = 3

	assert.Equal(t, expectedRouteState, test.routeState)

	// abort on fourth slice

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState.Next = false
	expectedRouteState.Veto = true
	expectedRouteState.Committed = false
	expectedRouteState.CommitCounter = 4
	expectedRouteState.CommitVeto = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_TryBeforeYouBuy_ReducePacketLoss_LatencyTolerance(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(40)

	test.directPacketLoss = float32(1)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.UserID = 100
	test.routeState.ReducePacketLoss = true
	test.routeState.Committed = false

	test.internal.TryBeforeYouBuy = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [MaxRelaysPerRoute]int32{0, 1}

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.Committed = true
	expectedRouteState.CommitCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestTakeNetworkNext_TryBeforeYouBuy_Multipath(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(50)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.Multipath = true

	test.routeState.UserID = 100

	test.internal.TryBeforeYouBuy = true
	test.internal.MultipathThreshold = 100

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Committed = true
	expectedRouteState.Multipath = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

// -------------------------------------------------------------

// This test makes sure we return predicted RTT (routeCost) even if
// the network next route is worse than direct. This gives us better visibility
// in the admin portal and in bigquery for direct routes, so we can see why
// they chose not to take network next (eg. best next route had high RTT).

func TestPredictedRTT(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	test := NewTestData(env)

	test.directLatency = int32(1)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.UserID = 100

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := RouteState{}
	expectedRouteState.UserID = 100

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(20+CostBias), test.routeCost)
	assert.Equal(t, int32(0), test.routeDiversity)
}

// -------------------------------------------------------------

func TestReframeRelays_NearRelayFilter(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("a", "10.0.0.1")
	env.AddRelay("b", "10.0.0.2")
	env.AddRelay("c", "10.0.0.3")
	env.AddRelay("d", "10.0.0.4")
	env.AddRelay("e", "10.0.0.5")

	relayIds := env.GetRelayIds()

	relayIdToIndex := env.GetRelayIdToIndex()

	// start with a clean slate

	routeShader := NewRouteShader()

	routeState := RouteState{}

	// next, pass in some near relay ids with initial rtt and jitter values

	directLatency := int32(100)
	directJitter := int32(5)
	directPacketLoss := int32(0)

	nextPacketLoss := int32(0)

	sourceRelayIds := relayIds
	sourceRelayLatency := []int32{50, 50, 50, 50, 50}
	sourceRelayJitter := []int32{5, 5, 5, 5, 5}
	sourceRelayPacketLoss := []int32{0, 0, 0, 0, 0}

	destRelayIds := relayIds

	out_sourceRelayLatency := []int32{0, 0, 0, 0, 0}
	out_sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	out_numDestRelays := int32(0)
	out_destRelays := []int32{0, 0, 0, 0, 0}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 0, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{50, 50, 50, 50, 50}, out_sourceRelayLatency)
	assert.Equal(t, []int32{5, 5, 5, 5, 5}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(5), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{50, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{5, 5, 5, 5, 5}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// now pass in some higher values and make sure they get picked up

	directJitter = 10
	sourceRelayLatency = []int32{100, 100, 100, 100, 100}
	sourceRelayJitter = []int32{10, 10, 10, 10, 10}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 1, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{100, 100, 100, 100, 100}, out_sourceRelayLatency)
	assert.Equal(t, []int32{10, 10, 10, 10, 10}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(10), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{100, 100, 100, 100, 100}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{10, 10, 10, 10, 10}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// pass in some lower values and make sure they get ignored

	directJitter = 9
	sourceRelayLatency = []int32{99, 99, 99, 99, 99}
	sourceRelayJitter = []int32{9, 9, 9, 9, 9}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 2, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{100, 100, 100, 100, 100}, out_sourceRelayLatency)
	assert.Equal(t, []int32{10, 10, 10, 10, 10}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(10), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{100, 100, 100, 100, 100}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{10, 10, 10, 10, 10}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// filter out the first source relay permanently by giving it high packet loss

	sourceRelayPacketLoss = []int32{50, 0, 0, 0, 0}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 3, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{255, 100, 100, 100, 100}, out_sourceRelayLatency)
	assert.Equal(t, []int32{10, 10, 10, 10, 10}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(10), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{255, 100, 100, 100, 100}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{10, 10, 10, 10, 10}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// filter out the second source relay permanently by having it report 0 RTT (impossible)

	sourceRelayLatency = []int32{99, 0, 99, 99, 99}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 4, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{255, 255, 100, 100, 100}, out_sourceRelayLatency)
	assert.Equal(t, []int32{10, 10, 10, 10, 10}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(10), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{255, 255, 100, 100, 100}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{10, 10, 10, 10, 10}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// filter out the third relay permanently by removing it from the set of relays

	delete(relayIdToIndex, relayIds[2])

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 5, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{255, 255, 255, 100, 100}, out_sourceRelayLatency)
	assert.Equal(t, []int32{10, 10, 10, 10, 10}, out_sourceRelayJitter)
	assert.Equal(t, int32(4), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 3, 4}, out_destRelays[:out_numDestRelays])

	assert.Equal(t, int32(10), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{255, 255, 255, 100, 100}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{10, 10, 10, 10, 10}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// temporary exclude source relay with higher jitter than direct

	sourceRelayJitter = []int32{100, 100, 100, 100, 100}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 6, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{255, 255, 255, 255, 255}, out_sourceRelayLatency)
	assert.Equal(t, []int32{10, 10, 10, 100, 100}, out_sourceRelayJitter)
	assert.Equal(t, int32(4), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 3, 4}, out_destRelays[:out_numDestRelays])

	assert.Equal(t, int32(10), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{255, 255, 255, 100, 100}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{10, 10, 10, 100, 100}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// increase direct jitter above source relays and verify they recover

	directJitter = int32(110)

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 7, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{255, 255, 255, 100, 100}, out_sourceRelayLatency)
	assert.Equal(t, []int32{10, 10, 10, 100, 100}, out_sourceRelayJitter)
	assert.Equal(t, int32(4), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 3, 4}, out_destRelays[:out_numDestRelays])

	assert.Equal(t, int32(110), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{255, 255, 255, 100, 100}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{10, 10, 10, 100, 100}, routeState.NearRelayJitter[:routeState.NumNearRelays])
}

func TestReframeRelays_ReduceJitter_Simple(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("a", "10.0.0.1")
	env.AddRelay("b", "10.0.0.2")
	env.AddRelay("c", "10.0.0.3")
	env.AddRelay("d", "10.0.0.4")
	env.AddRelay("e", "10.0.0.5")

	relayIds := env.GetRelayIds()

	relayIdToIndex := env.GetRelayIdToIndex()

	// start with a clean slate

	routeShader := NewRouteShader()

	routeState := RouteState{}

	// next, pass in some near relay ids with initial rtt and jitter values

	directLatency := int32(100)
	directJitter := int32(5)
	directPacketLoss := int32(0)

	nextPacketLoss := int32(0)

	sourceRelayIds := relayIds
	sourceRelayLatency := []int32{50, 50, 50, 50, 50}
	sourceRelayJitter := []int32{5, 5, 5, 5, 5}
	sourceRelayPacketLoss := []int32{0, 0, 0, 0, 0}

	destRelayIds := relayIds

	out_sourceRelayLatency := []int32{0, 0, 0, 0, 0}
	out_sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	out_numDestRelays := int32(0)
	out_destRelays := []int32{0, 0, 0, 0, 0}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 0, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{50, 50, 50, 50, 50}, out_sourceRelayLatency)
	assert.Equal(t, []int32{5, 5, 5, 5, 5}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(5), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{50, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{5, 5, 5, 5, 5}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// now increase the near relay jitter so it is above the direct jitter + threshold and verify these relays are excluded temporarily

	highJitter := directJitter + JitterThreshold + 1

	sourceRelayJitter = []int32{highJitter, highJitter, highJitter, highJitter, highJitter}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 1, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{255, 255, 255, 255, 255}, out_sourceRelayLatency)
	assert.Equal(t, []int32{highJitter, highJitter, highJitter, highJitter, highJitter}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(5), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{50, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{highJitter, highJitter, highJitter, highJitter, highJitter}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// increase direct jitter and all near relays should recover and be routable again

	directJitter = int32(30)

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 2, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{50, 50, 50, 50, 50}, out_sourceRelayLatency)
	assert.Equal(t, []int32{highJitter, highJitter, highJitter, highJitter, highJitter}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(30), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{50, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{highJitter, highJitter, highJitter, highJitter, highJitter}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// now blow out direct jitter and a few near relays. make sure the high jitter near relays
	// get excluded because they have higher than average jitter across all near relays.

	directJitter = int32(100)

	sourceRelayJitter = []int32{highJitter, highJitter, 90, 90, 90}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 3, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{50, 50, 255, 255, 255}, out_sourceRelayLatency)
	assert.Equal(t, []int32{highJitter, highJitter, 90, 90, 90}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(100), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{50, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{highJitter, highJitter, 90, 90, 90}, routeState.NearRelayJitter[:routeState.NumNearRelays])
}

func TestReframeRelays_ReduceJitter_Threshold(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("a", "10.0.0.1")
	env.AddRelay("b", "10.0.0.2")
	env.AddRelay("c", "10.0.0.3")
	env.AddRelay("d", "10.0.0.4")
	env.AddRelay("e", "10.0.0.5")

	relayIds := env.GetRelayIds()

	relayIdToIndex := env.GetRelayIdToIndex()

	// start with a clean slate

	routeShader := NewRouteShader()

	routeState := RouteState{}

	// pass in zero jitter for direct, and verify 1-5ms jitter (jitter threshold) are still routable
	// this means we consider any jitter <= jitter threshold basically equivalent, so we don't
	// overconstrain the system and reject routes that are otherwise good.

	directLatency := int32(100)
	directJitter := int32(0)
	directPacketLoss := int32(0)

	nextPacketLoss := int32(0)

	sourceRelayIds := relayIds
	sourceRelayLatency := []int32{50, 50, 50, 50, 50}
	sourceRelayJitter := []int32{1, 2, 3, 4, 5}
	sourceRelayPacketLoss := []int32{0, 0, 0, 0, 0}

	destRelayIds := relayIds

	out_sourceRelayLatency := []int32{0, 0, 0, 0, 0}
	out_sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	out_numDestRelays := int32(0)
	out_destRelays := []int32{0, 0, 0, 0, 0}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 0, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{50, 50, 50, 50, 50}, out_sourceRelayLatency)
	assert.Equal(t, []int32{1, 2, 3, 4, 5}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(0), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{50, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{1, 2, 3, 4, 5}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// now increase the near relay jitter above the threshold and verify the relays get excluded

	sourceRelayJitter = []int32{JitterThreshold + 1, JitterThreshold + 1, JitterThreshold + 1, JitterThreshold + 1, JitterThreshold + 1}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 1, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{255, 255, 255, 255, 255}, out_sourceRelayLatency)
	assert.Equal(t, []int32{JitterThreshold + 1, JitterThreshold + 1, JitterThreshold + 1, JitterThreshold + 1, JitterThreshold + 1}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(0), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{50, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{JitterThreshold + 1, JitterThreshold + 1, JitterThreshold + 1, JitterThreshold + 1, JitterThreshold + 1}, routeState.NearRelayJitter[:routeState.NumNearRelays])

}

func TestReframeRelays_ReducePacketLoss_Sporadic(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("a", "10.0.0.1")
	env.AddRelay("b", "10.0.0.2")
	env.AddRelay("c", "10.0.0.3")
	env.AddRelay("d", "10.0.0.4")
	env.AddRelay("e", "10.0.0.5")

	relayIds := env.GetRelayIds()

	relayIdToIndex := env.GetRelayIdToIndex()

	// start with a clean slate

	routeShader := NewRouteShader()

	routeState := RouteState{}

	// pass in some near relay ids with a near relay that LOOKS attractive (lowest latency)
	// but has packet loss, and verify it gets excluded temporarily.

	directLatency := int32(100)
	directJitter := int32(10)
	directPacketLoss := int32(0)

	nextPacketLoss := int32(0)

	sourceRelayIds := relayIds
	sourceRelayLatency := []int32{10, 50, 50, 50, 50}
	sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	sourceRelayPacketLoss := []int32{1, 0, 0, 0, 0}

	destRelayIds := relayIds

	out_sourceRelayLatency := []int32{0, 0, 0, 0, 0}
	out_sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	out_numDestRelays := int32(0)
	out_destRelays := []int32{0, 0, 0, 0, 0}

	sliceNumber := int32(0)

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, sliceNumber, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	sliceNumber++

	assert.Equal(t, []int32{255, 50, 50, 50, 50}, out_sourceRelayLatency)
	assert.Equal(t, []int32{0, 0, 0, 0, 0}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(10), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{10, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{0, 0, 0, 0, 0}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// verify it remains excluded from now on, as long as direct has no packet loss

	sourceRelayPacketLoss = []int32{0, 0, 0, 0, 0}

	for i := 0; i < 64; i++ {

		ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, sliceNumber, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

		sliceNumber++

		assert.Equal(t, []int32{255, 50, 50, 50, 50}, out_sourceRelayLatency)
		assert.Equal(t, []int32{0, 0, 0, 0, 0}, out_sourceRelayJitter)
		assert.Equal(t, int32(5), out_numDestRelays)
		assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

		assert.Equal(t, int32(10), routeState.DirectJitter)
		assert.Equal(t, int32(5), routeState.NumNearRelays)
		assert.Equal(t, []int32{10, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
		assert.Equal(t, []int32{0, 0, 0, 0, 0}, routeState.NearRelayJitter[:routeState.NumNearRelays])
	}

}

func TestReframeRelays_ReducePacketLoss_Continuous(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("a", "10.0.0.1")
	env.AddRelay("b", "10.0.0.2")
	env.AddRelay("c", "10.0.0.3")
	env.AddRelay("d", "10.0.0.4")
	env.AddRelay("e", "10.0.0.5")

	relayIds := env.GetRelayIds()

	relayIdToIndex := env.GetRelayIdToIndex()

	// start with a clean slate

	routeShader := NewRouteShader()

	routeState := RouteState{}

	// pass in some near relay ids with a near relay that LOOKS attractive (lowest latency)
	// but has packet loss higher than direct, and verify it gets excluded temporarily.

	directLatency := int32(100)
	directJitter := int32(10)
	directPacketLoss := int32(1)

	nextPacketLoss := int32(0)

	sourceRelayIds := relayIds
	sourceRelayLatency := []int32{10, 50, 50, 50, 50}
	sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	sourceRelayPacketLoss := []int32{2, 0, 0, 0, 0}

	destRelayIds := relayIds

	out_sourceRelayLatency := []int32{0, 0, 0, 0, 0}
	out_sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	out_numDestRelays := int32(0)
	out_destRelays := []int32{0, 0, 0, 0, 0}

	sliceNumber := int32(0)

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, sliceNumber, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	sliceNumber++

	assert.Equal(t, []int32{255, 50, 50, 50, 50}, out_sourceRelayLatency)
	assert.Equal(t, []int32{0, 0, 0, 0, 0}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(10), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{10, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{0, 0, 0, 0, 0}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// verify it remains excluded while packet loss continues

	for i := 0; i < 7; i++ {

		ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, sliceNumber, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

		sliceNumber++

		assert.Equal(t, []int32{255, 50, 50, 50, 50}, out_sourceRelayLatency)
		assert.Equal(t, []int32{0, 0, 0, 0, 0}, out_sourceRelayJitter)
		assert.Equal(t, int32(5), out_numDestRelays)
		assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

		assert.Equal(t, int32(10), routeState.DirectJitter)
		assert.Equal(t, int32(5), routeState.NumNearRelays)
		assert.Equal(t, []int32{10, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
		assert.Equal(t, []int32{0, 0, 0, 0, 0}, routeState.NearRelayJitter[:routeState.NumNearRelays])
	}

	// stop the packet loss and it should recover after three slices (30 seconds)

	sourceRelayPacketLoss = []int32{0, 0, 0, 0, 0}

	for i := 0; i < 3; i++ {

		ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, sliceNumber, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

		sliceNumber++

		assert.Equal(t, []int32{255, 50, 50, 50, 50}, out_sourceRelayLatency)
		assert.Equal(t, []int32{0, 0, 0, 0, 0}, out_sourceRelayJitter)
		assert.Equal(t, int32(5), out_numDestRelays)
		assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

		assert.Equal(t, int32(10), routeState.DirectJitter)
		assert.Equal(t, int32(5), routeState.NumNearRelays)
		assert.Equal(t, []int32{10, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
		assert.Equal(t, []int32{0, 0, 0, 0, 0}, routeState.NearRelayJitter[:routeState.NumNearRelays])
	}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, sliceNumber, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	sliceNumber++

	assert.Equal(t, []int32{10, 50, 50, 50, 50}, out_sourceRelayLatency)
	assert.Equal(t, []int32{0, 0, 0, 0, 0}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(10), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{10, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{0, 0, 0, 0, 0}, routeState.NearRelayJitter[:routeState.NumNearRelays])
}

func TestReframeRelays_ReducePacketLoss_NotWorse(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("a", "10.0.0.1")
	env.AddRelay("b", "10.0.0.2")
	env.AddRelay("c", "10.0.0.3")
	env.AddRelay("d", "10.0.0.4")
	env.AddRelay("e", "10.0.0.5")

	relayIds := env.GetRelayIds()

	relayIdToIndex := env.GetRelayIdToIndex()

	// start with a clean slate

	routeShader := NewRouteShader()

	routeState := RouteState{}

	// give direct some packet loss on first slice and the same amount of PL on the first near relay
	// the near relay should not be excluded because it was at no point ever "worse" packet loss than direct

	directLatency := int32(100)
	directJitter := int32(10)
	directPacketLoss := int32(1)

	nextPacketLoss := int32(1)

	sourceRelayIds := relayIds
	sourceRelayLatency := []int32{10, 50, 50, 50, 50}
	sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	sourceRelayPacketLoss := []int32{1, 0, 0, 0, 0}

	destRelayIds := relayIds

	out_sourceRelayLatency := []int32{0, 0, 0, 0, 0}
	out_sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	out_numDestRelays := int32(0)
	out_destRelays := []int32{0, 0, 0, 0, 0}

	ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 0, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

	assert.Equal(t, []int32{10, 50, 50, 50, 50}, out_sourceRelayLatency)
	assert.Equal(t, []int32{0, 0, 0, 0, 0}, out_sourceRelayJitter)
	assert.Equal(t, int32(5), out_numDestRelays)
	assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

	assert.Equal(t, int32(10), routeState.DirectJitter)
	assert.Equal(t, int32(5), routeState.NumNearRelays)
	assert.Equal(t, []int32{10, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
	assert.Equal(t, []int32{0, 0, 0, 0, 0}, routeState.NearRelayJitter[:routeState.NumNearRelays])

	// verify it remains routable for 7 more slices (history size is 8) after packet loss stops

	directPacketLoss = 0.0
	nextPacketLoss = 0.0

	sourceRelayPacketLoss = []int32{0, 0, 0, 0, 0}

	for i := 0; i < 7; i++ {

		ReframeRelays(&routeShader, &routeState, relayIdToIndex, directLatency, directJitter, directPacketLoss, nextPacketLoss, 1, sourceRelayIds, sourceRelayLatency, sourceRelayJitter, sourceRelayPacketLoss, destRelayIds, out_sourceRelayLatency, out_sourceRelayJitter, &out_numDestRelays, out_destRelays)

		assert.Equal(t, []int32{10, 50, 50, 50, 50}, out_sourceRelayLatency)
		assert.Equal(t, []int32{0, 0, 0, 0, 0}, out_sourceRelayJitter)
		assert.Equal(t, int32(5), out_numDestRelays)
		assert.Equal(t, []int32{0, 1, 2, 3, 4}, out_destRelays)

		assert.Equal(t, int32(10), routeState.DirectJitter)
		assert.Equal(t, int32(5), routeState.NumNearRelays)
		assert.Equal(t, []int32{10, 50, 50, 50, 50}, routeState.NearRelayRTT[:routeState.NumNearRelays])
		assert.Equal(t, []int32{0, 0, 0, 0, 0}, routeState.NearRelayJitter[:routeState.NumNearRelays])
	}
}

// -------------------------------------------------------------

func randomBytes(buffer []byte) {
	for i := 0; i < len(buffer); i++ {
		buffer[i] = byte(rand.Intn(256))
	}
}

func TestPittle(t *testing.T) {
	rand.Seed(42)
	var output [256]byte
    for i := 0; i < 10000; i++ {
    	var fromAddress [4]byte
    	var toAddress [4]byte
    	randomBytes(fromAddress[:])
    	randomBytes(toAddress[:])
    	fromPort := uint16(i+1000000)
    	toPort := uint16(i+5000)
    	packetLength := 1 + (i % 1500)
    	GeneratePittle(output[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
    	assert.NotEqual(t, output[0], 0)
    	assert.NotEqual(t, output[1], 0)
    }
}

func TestChonkle(t *testing.T) {
	rand.Seed(42)
	var output [1500]byte
	output[0] = 1
	for i := 0; i < 10000; i++ {
		var magic [8]byte
		var fromAddress [4]byte
		var toAddress [4]byte
    	randomBytes(magic[:])
    	randomBytes(fromAddress[:])
    	randomBytes(toAddress[:])
    	fromPort := uint16(i+1000000)
    	toPort := uint16(i+5000)
    	packetLength := 18 + ( i % ( len(output) - 18 ) )
    	GenerateChonkle(output[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
    	assert.Equal(t, true, BasicPacketFilter(output[:], packetLength))
	}
}

func TestABI(t *testing.T) {

	var output [1024]byte
	
	magic := [8]byte{1,2,3,4,5,6,7,8}
	fromAddress := [4]byte{1,2,3,4}
	toAddress := [4]byte{4,3,2,1}
	fromPort := uint16(1000)
	toPort := uint16(5000)
    packetLength := 1000
    
    GeneratePittle(output[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	assert.Equal(t, output[0], uint8(71) )
	assert.Equal(t, output[1], uint8(201) )

    GenerateChonkle(output[:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
	
	assert.Equal(t, output[0], uint8(45) )
	assert.Equal(t, output[1], uint8(203) )
	assert.Equal(t, output[2], uint8(67) )
	assert.Equal(t, output[3], uint8(96) )
	assert.Equal(t, output[4], uint8(78) )
	assert.Equal(t, output[5], uint8(180) )
	assert.Equal(t, output[6], uint8(127) )
	assert.Equal(t, output[7], uint8(7) )
}

func TestPittleAndChonkle(t *testing.T) {
	rand.Seed(42)
	var output [1500]byte
	output[0] = 1
	for i := 0; i < 10000; i++ {
		var magic [8]byte
		var fromAddress [4]byte
		var toAddress [4]byte
    	randomBytes(magic[:])
    	randomBytes(fromAddress[:])
    	randomBytes(toAddress[:])
    	fromPort := uint16(i+1000000)
    	toPort := uint16(i+5000)
    	packetLength := 18 + ( i % ( len(output) - 18 ) )
    	GenerateChonkle(output[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
    	GeneratePittle(output[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
    	assert.Equal(t, true, BasicPacketFilter(output[:], packetLength))
    	assert.Equal(t, true, AdvancedPacketFilter(output[:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength))
	}
}

func TestBasicPacketFilter(t *testing.T) {
	rand.Seed(42)
	var output [256]byte
	pass := 0
	iterations := 10000
	for i := 0; i < iterations; i++ {
		randomBytes(output[:])
		packetLength := i % len(output)
    	assert.Equal(t, false, BasicPacketFilter(output[:], packetLength))
	}
   	assert.Equal(t, 0, pass)
}

func TestAdvancedBasicPacketFilter(t *testing.T) {
	rand.Seed(42)
	var output [1500]byte
	iterations := 10000
	for i := 0; i < iterations; i++ {
		var magic [8]byte
		var fromAddress [4]byte
		var toAddress [4]byte
    	randomBytes(magic[:])
    	randomBytes(fromAddress[:])
    	randomBytes(toAddress[:])
    	fromPort := uint16(i+1000000)
    	toPort := uint16(i+5000)
		randomBytes(output[:])
		packetLength := i % len(output)
    	assert.Equal(t, false, BasicPacketFilter(output[:], packetLength))
    	assert.Equal(t, false, AdvancedPacketFilter(output[:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength))
	}
}
