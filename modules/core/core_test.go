package core_test

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"net"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
)

func RelayHash64(name string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(name))
	return hash.Sum64()
}

func TestProtocolVersionAtLeast(t *testing.T) {
	t.Parallel()
	assert.True(t, core.ProtocolVersionAtLeast(3, 0, 0, 3, 0, 0))
	assert.True(t, core.ProtocolVersionAtLeast(4, 0, 0, 3, 0, 0))
	assert.True(t, core.ProtocolVersionAtLeast(3, 1, 0, 3, 0, 0))
	assert.True(t, core.ProtocolVersionAtLeast(3, 0, 1, 3, 0, 0))
	assert.True(t, core.ProtocolVersionAtLeast(3, 4, 5, 3, 4, 5))
	assert.True(t, core.ProtocolVersionAtLeast(4, 0, 0, 3, 4, 5))
	assert.True(t, core.ProtocolVersionAtLeast(3, 5, 0, 3, 4, 5))
	assert.True(t, core.ProtocolVersionAtLeast(3, 4, 6, 3, 4, 5))
	assert.True(t, core.ProtocolVersionAtLeast(3, 1, 0, 3, 1, 0))
	assert.False(t, core.ProtocolVersionAtLeast(3, 0, 99, 3, 1, 1))
	assert.False(t, core.ProtocolVersionAtLeast(3, 1, 0, 3, 1, 1))
	assert.False(t, core.ProtocolVersionAtLeast(2, 0, 0, 3, 1, 1))
	assert.False(t, core.ProtocolVersionAtLeast(3, 0, 5, 3, 1, 0))
}

func TestHaversineDistance(t *testing.T) {
	t.Parallel()
	losangelesLatitude := 34.0522
	losangelesLongitude := -118.2437
	bostonLatitude := 42.3601
	bostonLongitude := -71.0589
	distance := core.HaversineDistance(losangelesLatitude, losangelesLongitude, bostonLatitude, bostonLongitude)
	assert.Equal(t, 4169.607203810275, distance)
}

func TestTriMatrixLength(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 0, core.TriMatrixLength(0))
	assert.Equal(t, 0, core.TriMatrixLength(1))
	assert.Equal(t, 1, core.TriMatrixLength(2))
	assert.Equal(t, 3, core.TriMatrixLength(3))
	assert.Equal(t, 6, core.TriMatrixLength(4))
	assert.Equal(t, 10, core.TriMatrixLength(5))
	assert.Equal(t, 15, core.TriMatrixLength(6))
}

func TestTriMatrixIndex(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 0, core.TriMatrixIndex(0, 1))
	assert.Equal(t, 1, core.TriMatrixIndex(0, 2))
	assert.Equal(t, 2, core.TriMatrixIndex(1, 2))
	assert.Equal(t, 3, core.TriMatrixIndex(0, 3))
	assert.Equal(t, 4, core.TriMatrixIndex(1, 3))
	assert.Equal(t, 5, core.TriMatrixIndex(2, 3))
	assert.Equal(t, 0, core.TriMatrixIndex(1, 0))
	assert.Equal(t, 1, core.TriMatrixIndex(2, 0))
	assert.Equal(t, 2, core.TriMatrixIndex(2, 1))
	assert.Equal(t, 3, core.TriMatrixIndex(3, 0))
	assert.Equal(t, 4, core.TriMatrixIndex(3, 1))
	assert.Equal(t, 5, core.TriMatrixIndex(3, 2))
}

func CheckNilAddress(t *testing.T) {
	var address *net.UDPAddr
	buffer := make([]uint8, constants.NEXT_ADDRESS_BYTES)
	core.WriteAddress(buffer, address)
	readAddress := core.ReadAddress(buffer)
	assert.Equal(t, readAddress, net.UDPAddr{})
}

func CheckIPv4Address(t *testing.T, addressString string, expected string) {
	address := core.ParseAddress(addressString)
	buffer := make([]uint8, constants.NEXT_ADDRESS_BYTES)
	core.WriteAddress(buffer, &address)
	readAddress := core.ReadAddress(buffer)
	readAddressString := readAddress.String()
	assert.Equal(t, expected, readAddressString)
}

func CheckIPv6Address(t *testing.T, addressString string, expected string) {
	address := core.ParseAddress(addressString)
	buffer := make([]uint8, constants.NEXT_ADDRESS_BYTES)
	core.WriteAddress(buffer, &address)
	readAddress := core.ReadAddress(buffer)
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

	routeManager := core.RouteManager{}
	routeManager.RelayDatacenter = make([]uint64, 256)
	for i := range routeManager.RelayDatacenter {
		routeManager.RelayDatacenter[i] = uint64(i)
	}
	routeManager.RelayDatacenter[255] = 254

	assert.Equal(t, 0, routeManager.NumRoutes)

	routeManager.AddRoute(10, 1, 2, 3)
	assert.Equal(t, 1, routeManager.NumRoutes)
	assert.Equal(t, int32(10), routeManager.RouteCost[0])
	assert.Equal(t, int32(3), routeManager.RouteNumRelays[0])
	assert.Equal(t, int32(1), routeManager.RouteRelays[0][0])
	assert.Equal(t, int32(2), routeManager.RouteRelays[0][1])
	assert.Equal(t, int32(3), routeManager.RouteRelays[0][2])

	routeManager.AddRoute(20, 4, 5, 6)
	assert.Equal(t, 2, routeManager.NumRoutes)

	// verify adding the same route twice gets filtered out

	routeManager.AddRoute(20, 4, 5, 6)
	assert.Equal(t, 2, routeManager.NumRoutes)

	// verify loops get filtered out

	routeManager.AddRoute(20, 4, 4, 5, 6)
	assert.Equal(t, 2, routeManager.NumRoutes)

	// verify routes with cost >= 255 get filtered out

	routeManager.AddRoute(255, 4, 5, 6)
	assert.Equal(t, 2, routeManager.NumRoutes)

	routeManager.AddRoute(1000, 4, 5, 6)
	assert.Equal(t, 2, routeManager.NumRoutes)

	// add some more routes

	routeManager.AddRoute(21, 4, 5, 254, 255)
	assert.Equal(t, 3, routeManager.NumRoutes)

	routeManager.AddRoute(14, 5, 6, 7, 8, 9)
	assert.Equal(t, 4, routeManager.NumRoutes)

	routeManager.AddRoute(18, 6, 7, 8)
	assert.Equal(t, 5, routeManager.NumRoutes)

	routeManager.AddRoute(17, 8, 9)
	assert.Equal(t, 6, routeManager.NumRoutes)

	routeManager.AddRoute(16, 9, 10, 11)
	assert.Equal(t, 7, routeManager.NumRoutes)

	routeManager.AddRoute(19, 10, 11, 12, 13, 14)
	assert.Equal(t, 8, routeManager.NumRoutes)

	routeManager.AddRoute(15, 11, 12)
	assert.Equal(t, 9, routeManager.NumRoutes)

	for i := 0; i < routeManager.NumRoutes-1; i++ {
		assert.True(t, routeManager.RouteCost[i] <= routeManager.RouteCost[i+1])
	}

	// fill up lots of extra routes to get to max routes

	numFillers := constants.MaxRoutesPerEntry - routeManager.NumRoutes

	for i := 0; i < numFillers; i++ {
		routeManager.AddRoute(int32(50+i), int32(100+i), int32(100+i+1), int32(100+i+2))
		assert.Equal(t, 9+i+1, routeManager.NumRoutes)
	}

	assert.Equal(t, constants.MaxRoutesPerEntry, routeManager.NumRoutes)

	// make sure we can't add worse routes once we are at max routes

	routeManager.AddRoute(250, 12, 13, 14)
	assert.Equal(t, routeManager.NumRoutes, constants.MaxRoutesPerEntry)
	for i := 0; i < routeManager.NumRoutes; i++ {
		assert.True(t, routeManager.RouteCost[i] != 250)
	}

	// make sure we can add better routes while at max routes

	routeManager.AddRoute(5, 13, 14, 15, 16, 17)
	assert.Equal(t, routeManager.NumRoutes, constants.MaxRoutesPerEntry)
	for i := 0; i < routeManager.NumRoutes-1; i++ {
		assert.True(t, routeManager.RouteCost[i] <= routeManager.RouteCost[i+1])
	}
	found := false
	for i := 0; i < routeManager.NumRoutes; i++ {
		if routeManager.RouteCost[i] == 5 {
			found = true
		}
	}
	assert.True(t, found)

	// check all the best routes are sorted and they have correct data

	assert.Equal(t, int32(5), routeManager.RouteCost[0])
	assert.Equal(t, int32(5), routeManager.RouteNumRelays[0])
	assert.Equal(t, int32(13), routeManager.RouteRelays[0][0])
	assert.Equal(t, int32(14), routeManager.RouteRelays[0][1])
	assert.Equal(t, int32(15), routeManager.RouteRelays[0][2])
	assert.Equal(t, int32(16), routeManager.RouteRelays[0][3])
	assert.Equal(t, int32(17), routeManager.RouteRelays[0][4])
	assert.Equal(t, core.RouteHash(13, 14, 15, 16, 17), routeManager.RouteHash[0])

	assert.Equal(t, int32(10), routeManager.RouteCost[1])
	assert.Equal(t, int32(3), routeManager.RouteNumRelays[1])
	assert.Equal(t, int32(1), routeManager.RouteRelays[1][0])
	assert.Equal(t, int32(2), routeManager.RouteRelays[1][1])
	assert.Equal(t, int32(3), routeManager.RouteRelays[1][2])
	assert.Equal(t, core.RouteHash(1, 2, 3), routeManager.RouteHash[1])

	assert.Equal(t, int32(14), routeManager.RouteCost[2])
	assert.Equal(t, int32(5), routeManager.RouteNumRelays[2])
	assert.Equal(t, int32(5), routeManager.RouteRelays[2][0])
	assert.Equal(t, int32(6), routeManager.RouteRelays[2][1])
	assert.Equal(t, int32(7), routeManager.RouteRelays[2][2])
	assert.Equal(t, int32(8), routeManager.RouteRelays[2][3])
	assert.Equal(t, int32(9), routeManager.RouteRelays[2][4])
	assert.Equal(t, core.RouteHash(5, 6, 7, 8, 9), routeManager.RouteHash[2])

	assert.Equal(t, int32(15), routeManager.RouteCost[3])
	assert.Equal(t, int32(2), routeManager.RouteNumRelays[3])
	assert.Equal(t, int32(11), routeManager.RouteRelays[3][0])
	assert.Equal(t, int32(12), routeManager.RouteRelays[3][1])
	assert.Equal(t, core.RouteHash(11, 12), routeManager.RouteHash[3])

	assert.Equal(t, int32(16), routeManager.RouteCost[4])
	assert.Equal(t, int32(3), routeManager.RouteNumRelays[4])
	assert.Equal(t, int32(9), routeManager.RouteRelays[4][0])
	assert.Equal(t, int32(10), routeManager.RouteRelays[4][1])
	assert.Equal(t, int32(11), routeManager.RouteRelays[4][2])
	assert.Equal(t, core.RouteHash(9, 10, 11), routeManager.RouteHash[4])

	assert.Equal(t, int32(17), routeManager.RouteCost[5])
	assert.Equal(t, int32(2), routeManager.RouteNumRelays[5])
	assert.Equal(t, int32(8), routeManager.RouteRelays[5][0])
	assert.Equal(t, int32(9), routeManager.RouteRelays[5][1])
	assert.Equal(t, core.RouteHash(8, 9), routeManager.RouteHash[5])

	assert.Equal(t, int32(18), routeManager.RouteCost[6])
	assert.Equal(t, int32(3), routeManager.RouteNumRelays[6])
	assert.Equal(t, int32(6), routeManager.RouteRelays[6][0])
	assert.Equal(t, int32(7), routeManager.RouteRelays[6][1])
	assert.Equal(t, int32(8), routeManager.RouteRelays[6][2])
	assert.Equal(t, core.RouteHash(6, 7, 8), routeManager.RouteHash[6])

	assert.Equal(t, int32(19), routeManager.RouteCost[7])
	assert.Equal(t, int32(5), routeManager.RouteNumRelays[7])
	assert.Equal(t, int32(10), routeManager.RouteRelays[7][0])
	assert.Equal(t, int32(11), routeManager.RouteRelays[7][1])
	assert.Equal(t, int32(12), routeManager.RouteRelays[7][2])
	assert.Equal(t, int32(13), routeManager.RouteRelays[7][3])
	assert.Equal(t, int32(14), routeManager.RouteRelays[7][4])
	assert.Equal(t, core.RouteHash(10, 11, 12, 13, 14), routeManager.RouteHash[7])

	assert.Equal(t, int32(20), routeManager.RouteCost[8])
	assert.Equal(t, int32(3), routeManager.RouteNumRelays[8])
	assert.Equal(t, int32(4), routeManager.RouteRelays[8][0])
	assert.Equal(t, int32(5), routeManager.RouteRelays[8][1])
	assert.Equal(t, int32(6), routeManager.RouteRelays[8][2])
	assert.Equal(t, core.RouteHash(4, 5, 6), routeManager.RouteHash[8])
}

func Analyze(numRelays int, routes []core.RouteEntry) []int {

	buckets := make([]int, 8)

	for i := 0; i < numRelays; i++ {
		for j := 0; j < numRelays; j++ {
			if j < i {
				abFlatIndex := core.TriMatrixIndex(i, j)
				if routes[abFlatIndex].DirectCost != 255 {
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

type TestRelayData struct {
	name       string
	address    net.UDPAddr
	publicKey  []byte
	privateKey []byte
	index      int
}

type TestEnvironment struct {
	relayArray []*TestRelayData
	relays     map[string]*TestRelayData
	cost       [][]uint8
}

func NewTestEnvironment() *TestEnvironment {
	env := &TestEnvironment{}
	env.relays = make(map[string]*TestRelayData)
	return env
}

func (env *TestEnvironment) Clear() {
	numRelays := len(env.relays)
	env.cost = make([][]uint8, numRelays)
	for i := 0; i < numRelays; i++ {
		env.cost[i] = make([]uint8, numRelays)
		for j := 0; j < numRelays; j++ {
			env.cost[i][j] = 255
		}
	}
}

func GenerateRelayKeyPair() ([]byte, []byte, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	return publicKey, privateKey, err
}

func (env *TestEnvironment) AddRelay(relayName string, relayAddress string) {
	relay := &TestRelayData{}
	relay.name = relayName
	relay.address = core.ParseAddress(relayAddress)
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

func (env *TestEnvironment) SetCost(sourceRelayName string, destRelayName string, cost uint8) {
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

func (env *TestEnvironment) GetCostMatrix() ([]uint8, int) {
	numRelays := len(env.relays)
	entryCount := core.TriMatrixLength(numRelays)
	costMatrix := make([]uint8, entryCount)
	for i := 0; i < numRelays; i++ {
		for j := 0; j < i; j++ {
			index := core.TriMatrixIndex(i, j)
			costMatrix[index] = env.cost[i][j]
		}
	}
	return costMatrix, numRelays
}

type TestRouteData struct {
	cost   int32
	relays []string
}

func (env *TestEnvironment) GetRoutes(routeMatrix []core.RouteEntry, sourceRelayName string, destRelayName string) []TestRouteData {
	sourceRelay := env.relays[sourceRelayName]
	destRelay := env.relays[destRelayName]
	i := sourceRelay.index
	j := destRelay.index
	if i == j {
		return nil
	}
	index := core.TriMatrixIndex(i, j)
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

func (env *TestEnvironment) GetBestRouteCost(routeMatrix []core.RouteEntry, sourceRelays []string, sourceRelayCost []int32, destRelays []string) int32 {
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
	return core.GetBestRouteCost(routeMatrix, sourceRelayIndex, sourceRelayCost, destRelayIndex)
}

func (env *TestEnvironment) RouteExists(routeMatrix []core.RouteEntry, routeRelays []string) bool {
	routeRelayIndex := [constants.MaxRouteRelays]int32{}
	for i := range routeRelays {
		routeRelayIndex[i] = int32(env.GetRelayIndex(routeRelays[i]))
		if routeRelayIndex[i] == -1 {
			panic("bad route relay name")
		}
	}
	debug := ""
	return core.RouteExists(routeMatrix, int32(len(routeRelays)), routeRelayIndex, &debug)
}

func (env *TestEnvironment) GetCurrentRouteCost(routeMatrix []core.RouteEntry, routeRelays []string, sourceRelays []string, sourceRelayCost []int32, destRelays []string) int32 {
	routeRelayIndex := [constants.MaxRouteRelays]int32{}
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
	return core.GetCurrentRouteCost(routeMatrix, int32(len(routeRelays)), routeRelayIndex, sourceRelayIndex, sourceRelayCost, destRelayIndex, &debug)
}

func (env *TestEnvironment) GetBestRoutes(routeMatrix []core.RouteEntry, sourceRelays []string, sourceRelayCost []int32, destRelays []string, maxCost int32) []TestRouteData {
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
	bestRoutes := make([]core.BestRoute, 1024)
	core.GetBestRoutes(routeMatrix, sourceRelayIndex, sourceRelayCost, destRelayIndex, maxCost, bestRoutes, &numBestRoutes, &routeDiversity)
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

func (env *TestEnvironment) GetRandomBestRoute(routeMatrix []core.RouteEntry, sourceRelays []string, sourceRelayCost []int32, destRelays []string, maxCost int32) *TestRouteData {
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
	var bestRouteRelays [constants.MaxRouteRelays]int32
	debug := ""
	selectThreshold := int32(2)
	core.GetRandomBestRoute(routeMatrix, sourceRelayIndex, sourceRelayCost, destRelayIndex, maxCost, selectThreshold, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, &debug)
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

func (env *TestEnvironment) ReframeRouteHash(route []uint64) (int32, [constants.MaxRouteRelays]int32) {
	relayIdToIndex := make(map[uint64]int32)
	for _, v := range env.relays {
		id := RelayHash64(v.name)
		relayIdToIndex[id] = int32(v.index)
	}
	reframedRoute := [constants.MaxRouteRelays]int32{}
	if core.ReframeRoute(relayIdToIndex, route, &reframedRoute) {
		return int32(len(route)), reframedRoute
	}
	return 0, reframedRoute
}

func (env *TestEnvironment) ReframeRoute(routeRelayNames []string) (int32, [constants.MaxRouteRelays]int32) {
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

func (env *TestEnvironment) GetBestRoute_Initial(routeMatrix []core.RouteEntry, sourceRelayNames []string, sourceRelayCost []int32, destRelayNames []string, maxCost int32) (int32, int32, []string) {

	sourceRelays, destRelays := env.ReframeRelays(sourceRelayNames, destRelayNames)

	bestRouteCost := int32(0)
	bestRouteNumRelays := int32(0)
	bestRouteRelays := [constants.MaxRouteRelays]int32{}

	debug := ""
	selectThreshold := int32(2)
	hasRoute, routeDiversity := core.GetBestRoute_Initial(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, &debug)
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

func (env *TestEnvironment) GetBestRoute_Update(routeMatrix []core.RouteEntry, sourceRelayNames []string, sourceRelayCost []int32, destRelayNames []string, maxCost int32, selectThreshold int32, switchThreshold int32, currentRouteRelayNames []string) (int32, []string) {

	sourceRelays, destRelays := env.ReframeRelays(sourceRelayNames, destRelayNames)

	currentRouteNumRelays, currentRouteRelays := env.ReframeRoute(currentRouteRelayNames)
	if currentRouteNumRelays == 0 {
		panic("current route has no relays")
	}

	bestRouteCost := int32(0)
	bestRouteNumRelays := int32(0)
	bestRouteRelays := [constants.MaxRouteRelays]int32{}

	debug := ""
	core.GetBestRoute_Update(routeMatrix, sourceRelays, sourceRelayCost, destRelays, maxCost, selectThreshold, switchThreshold, currentRouteNumRelays, currentRouteRelays, &bestRouteCost, &bestRouteNumRelays, &bestRouteRelays, &debug)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceIndex := env.GetRelayIndex("losangeles")
	destIndex := env.GetRelayIndex("chicago")

	assert.True(t, sourceIndex != -1)
	assert.True(t, destIndex != -1)

	routeIndex := core.TriMatrixIndex(sourceIndex, destIndex)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	// verify the optimizer finds the indirect 3 relay route when the direct route does not exist

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

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

	env.SetCost("losangeles", "chicago", 250)
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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	// all routes are slower than direct. verify that we only have the direct route between losangeles and chicago

	assert.Equal(t, 1, len(routes))
	if len(routes) == 1 {
		assert.Equal(t, int32(10), routes[0].cost)
		assert.Equal(t, []string{"losangeles", "chicago"}, routes[0].relays)
	}
}

func TestRouteToken(t *testing.T) {

	t.Parallel()

	relayPublicKey, relayPrivateKey := crypto.Box_KeyPair()

	masterPublicKey, masterPrivateKey := crypto.Box_KeyPair()

	routeToken := core.RouteToken{}
	routeToken.ExpireTimestamp = uint64(time.Now().Unix() + 10)
	routeToken.SessionId = 0x123131231313131
	routeToken.SessionVersion = 100
	routeToken.KbpsUp = 256
	routeToken.KbpsDown = 512
	routeToken.NextAddress = core.ParseAddress("127.0.0.1:40000")
	routeToken.PrevAddress = core.ParseAddress("127.0.0.1:50000")
	routeToken.NextInternal = 1
	routeToken.PrevInternal = 1
	core.RandomBytes(routeToken.PrivateKey[:])

	// write the token to a buffer and read it back in

	buffer := make([]byte, constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)

	core.WriteRouteToken(&routeToken, buffer[:])

	var readRouteToken core.RouteToken
	err := core.ReadRouteToken(&readRouteToken, buffer)

	assert.NoError(t, err)
	assert.Equal(t, routeToken, readRouteToken)

	// can't read a token if the buffer is too small

	err = core.ReadRouteToken(&readRouteToken, buffer[:10])

	assert.Error(t, err)

	// write an encrypted route token and read it back

	core.WriteEncryptedRouteToken(&routeToken, buffer, masterPrivateKey[:], relayPublicKey[:])

	err = core.ReadEncryptedRouteToken(&readRouteToken, buffer, masterPublicKey[:], relayPrivateKey[:])

	assert.NoError(t, err)
	assert.Equal(t, routeToken, readRouteToken)

	// can't read an encrypted route token if the buffer is too small

	err = core.ReadEncryptedRouteToken(&readRouteToken, buffer[:10], masterPublicKey[:], relayPrivateKey[:])

	assert.Error(t, err)

	// can't read an encrypted route token if the buffer is garbage

	buffer = make([]byte, constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)

	err = core.ReadEncryptedRouteToken(&readRouteToken, buffer, masterPublicKey[:], relayPrivateKey[:])

	assert.Error(t, err)
}

func TestRouteTokens_PublicAddresses(t *testing.T) {

	t.Parallel()

	relayPublicKey, relayPrivateKey := crypto.Box_KeyPair()

	masterPublicKey, masterPrivateKey := crypto.Box_KeyPair()

	// write a bunch of tokens to a buffer

	publicAddresses := make([]net.UDPAddr, constants.NEXT_MAX_NODES)
	for i := range publicAddresses {
		publicAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 40000+i))
	}

	hasInternalAddresses := make([]bool, constants.NEXT_MAX_NODES)
	internalAddresses := make([]net.UDPAddr, constants.NEXT_MAX_NODES)
	internalGroups := make([]uint64, constants.NEXT_MAX_NODES)
	sellers := make([]int, constants.NEXT_MAX_NODES)

	publicKeys := make([][]byte, constants.NEXT_MAX_NODES)
	for i := range publicKeys {
		publicKeys[i] = make([]byte, crypto.Box_PublicKeySize)
		copy(publicKeys[i], relayPublicKey[:])
	}

	sessionId := uint64(0x123131231313131)
	sessionVersion := byte(100)
	kbpsUp := uint32(256)
	kbpsDown := uint32(256)
	expireTimestamp := uint64(time.Now().Unix() + 10)

	tokenData := make([]byte, constants.NEXT_MAX_NODES*constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)

	core.WriteRouteTokens(tokenData, expireTimestamp, sessionId, sessionVersion, kbpsUp, kbpsDown, constants.NEXT_MAX_NODES, publicAddresses, hasInternalAddresses, internalAddresses, internalGroups, sellers, publicKeys, masterPrivateKey)

	// read each token back individually and verify the token data matches what was written

	for i := 0; i < constants.NEXT_MAX_NODES; i++ {
		var routeToken core.RouteToken
		err := core.ReadEncryptedRouteToken(&routeToken, tokenData[i*constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], masterPublicKey[:], relayPrivateKey[:])
		assert.NoError(t, err)
		assert.Equal(t, sessionId, routeToken.SessionId)
		assert.Equal(t, sessionVersion, routeToken.SessionVersion)
		assert.Equal(t, kbpsUp, routeToken.KbpsUp)
		assert.Equal(t, kbpsDown, routeToken.KbpsDown)
		assert.Equal(t, expireTimestamp, routeToken.ExpireTimestamp)
		if i != 0 {
			assert.Equal(t, publicAddresses[i-1].String(), routeToken.PrevAddress.String())
		}
		if i != constants.NEXT_MAX_NODES-1 {
			assert.Equal(t, publicAddresses[i+1].String(), routeToken.NextAddress.String())
		}
		assert.Equal(t, routeToken.NextInternal, uint8(0))
		assert.Equal(t, routeToken.PrevInternal, uint8(0))
		assert.Equal(t, publicKeys[i], relayPublicKey[:])
	}
}

func TestRouteTokens_InternalAddresses(t *testing.T) {

	t.Parallel()

	relayPublicKey, relayPrivateKey := crypto.Box_KeyPair()

	masterPublicKey, masterPrivateKey := crypto.Box_KeyPair()

	// write some tokens with some that should communicate over internal addresses

	publicAddresses := make([]net.UDPAddr, constants.NEXT_MAX_NODES)
	for i := range publicAddresses {
		publicAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 40000+i))
	}

	hasInternalAddresses := make([]bool, constants.NEXT_MAX_NODES)
	internalAddresses := make([]net.UDPAddr, constants.NEXT_MAX_NODES)
	internalGroups := make([]uint64, constants.NEXT_MAX_NODES)
	sellers := make([]int, constants.NEXT_MAX_NODES)

	hasInternalAddresses[2] = true
	hasInternalAddresses[3] = true

	internalAddresses[2] = core.ParseAddress("10.0.0.1:40000")
	internalAddresses[3] = core.ParseAddress("10.0.0.2:40000")

	internalGroups[2] = 0x12345
	internalGroups[3] = 0x12345

	publicKeys := make([][]byte, constants.NEXT_MAX_NODES)
	for i := range publicKeys {
		publicKeys[i] = make([]byte, crypto.Box_PublicKeySize)
		copy(publicKeys[i], relayPublicKey[:])
	}

	sessionId := uint64(0x123131231313131)
	sessionVersion := byte(100)
	kbpsUp := uint32(256)
	kbpsDown := uint32(256)
	expireTimestamp := uint64(time.Now().Unix() + 10)

	tokenData := make([]byte, constants.NEXT_MAX_NODES*constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)

	core.WriteRouteTokens(tokenData, expireTimestamp, sessionId, sessionVersion, kbpsUp, kbpsDown, constants.NEXT_MAX_NODES, publicAddresses, hasInternalAddresses, internalAddresses, internalGroups, sellers, publicKeys, masterPrivateKey)

	// read each token back individually and verify the token data matches what was written

	for i := 0; i < constants.NEXT_MAX_NODES; i++ {
		var routeToken core.RouteToken
		err := core.ReadEncryptedRouteToken(&routeToken, tokenData[i*constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], masterPublicKey[:], relayPrivateKey[:])
		assert.NoError(t, err)
		assert.Equal(t, sessionId, routeToken.SessionId)
		assert.Equal(t, sessionVersion, routeToken.SessionVersion)
		assert.Equal(t, kbpsUp, routeToken.KbpsUp)
		assert.Equal(t, kbpsDown, routeToken.KbpsDown)
		assert.Equal(t, expireTimestamp, routeToken.ExpireTimestamp)
		if i == 2 {
			assert.Equal(t, routeToken.PrevInternal, uint8(0))
			assert.Equal(t, routeToken.NextInternal, uint8(1))
		} else if i == 3 {
			assert.Equal(t, routeToken.PrevInternal, uint8(1))
			assert.Equal(t, routeToken.NextInternal, uint8(0))
		} else {
			assert.Equal(t, routeToken.NextInternal, uint8(0))
			assert.Equal(t, routeToken.PrevInternal, uint8(0))
		}
		if i != 0 {
			if i == 3 {
				assert.Equal(t, "10.0.0.1:40000", routeToken.PrevAddress.String())
			} else {
				assert.Equal(t, publicAddresses[i-1].String(), routeToken.PrevAddress.String())
			}
		}
		if i != constants.NEXT_MAX_NODES-1 {
			if i == 2 {
				assert.Equal(t, "10.0.0.2:40000", routeToken.NextAddress.String())
			} else {
				assert.Equal(t, publicAddresses[i+1].String(), routeToken.NextAddress.String())
			}
		}
		assert.Equal(t, publicKeys[i], relayPublicKey[:])
	}
}

func TestRouteTokens_DifferentSellers(t *testing.T) {

	t.Parallel()

	relayPublicKey, relayPrivateKey := crypto.Box_KeyPair()

	masterPublicKey, masterPrivateKey := crypto.Box_KeyPair()

	// setup some relays with internal addresses, but give them different sellers. they should not use the internal addresses

	publicAddresses := make([]net.UDPAddr, constants.NEXT_MAX_NODES)
	for i := range publicAddresses {
		publicAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 40000+i))
	}

	hasInternalAddresses := make([]bool, constants.NEXT_MAX_NODES)
	internalAddresses := make([]net.UDPAddr, constants.NEXT_MAX_NODES)
	internalGroups := make([]uint64, constants.NEXT_MAX_NODES)
	sellers := make([]int, constants.NEXT_MAX_NODES)

	hasInternalAddresses[2] = true
	hasInternalAddresses[3] = true

	internalAddresses[2] = core.ParseAddress("10.0.0.1:40000")
	internalAddresses[3] = core.ParseAddress("10.0.0.2:40000")

	internalGroups[2] = 0x12345
	internalGroups[3] = 0x12345

	sellers[2] = 1
	sellers[3] = 2

	publicKeys := make([][]byte, constants.NEXT_MAX_NODES)
	for i := range publicKeys {
		publicKeys[i] = make([]byte, crypto.Box_PublicKeySize)
		copy(publicKeys[i], relayPublicKey[:])
	}

	sessionId := uint64(0x123131231313131)
	sessionVersion := byte(100)
	kbpsUp := uint32(256)
	kbpsDown := uint32(256)
	expireTimestamp := uint64(time.Now().Unix() + 10)

	tokenData := make([]byte, constants.NEXT_MAX_NODES*constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)

	core.WriteRouteTokens(tokenData, expireTimestamp, sessionId, sessionVersion, kbpsUp, kbpsDown, constants.NEXT_MAX_NODES, publicAddresses, hasInternalAddresses, internalAddresses, internalGroups, sellers, publicKeys, masterPrivateKey)

	// read each token back individually and verify the token data matches what was written

	for i := 0; i < constants.NEXT_MAX_NODES; i++ {
		var routeToken core.RouteToken
		err := core.ReadEncryptedRouteToken(&routeToken, tokenData[i*constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], masterPublicKey[:], relayPrivateKey[:])
		assert.NoError(t, err)
		assert.Equal(t, sessionId, routeToken.SessionId)
		assert.Equal(t, sessionVersion, routeToken.SessionVersion)
		assert.Equal(t, kbpsUp, routeToken.KbpsUp)
		assert.Equal(t, kbpsDown, routeToken.KbpsDown)
		assert.Equal(t, expireTimestamp, routeToken.ExpireTimestamp)
		if i != 0 {
			assert.Equal(t, publicAddresses[i-1].String(), routeToken.PrevAddress.String())
		}
		if i != constants.NEXT_MAX_NODES-1 {
			assert.Equal(t, publicAddresses[i+1].String(), routeToken.NextAddress.String())
		}
		assert.Equal(t, routeToken.NextInternal, uint8(0))
		assert.Equal(t, routeToken.PrevInternal, uint8(0))
		assert.Equal(t, publicKeys[i], relayPublicKey[:])
	}
}

func TestRouteTokens_DifferentGroups(t *testing.T) {

	t.Parallel()

	relayPublicKey, relayPrivateKey := crypto.Box_KeyPair()

	masterPublicKey, masterPrivateKey := crypto.Box_KeyPair()

	// setup some relays with internal addresses, but give them different internal groups in the same seller. they should not use internal addresses

	publicAddresses := make([]net.UDPAddr, constants.NEXT_MAX_NODES)
	for i := range publicAddresses {
		publicAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 40000+i))
	}

	hasInternalAddresses := make([]bool, constants.NEXT_MAX_NODES)
	internalAddresses := make([]net.UDPAddr, constants.NEXT_MAX_NODES)
	internalGroups := make([]uint64, constants.NEXT_MAX_NODES)
	sellers := make([]int, constants.NEXT_MAX_NODES)

	hasInternalAddresses[2] = true
	hasInternalAddresses[3] = true

	internalAddresses[2] = core.ParseAddress("10.0.0.1:40000")
	internalAddresses[3] = core.ParseAddress("10.0.0.2:40000")

	internalGroups[2] = 0x12345
	internalGroups[3] = 0x22334

	sellers[2] = 1
	sellers[3] = 1

	publicKeys := make([][]byte, constants.NEXT_MAX_NODES)
	for i := range publicKeys {
		publicKeys[i] = make([]byte, crypto.Box_PublicKeySize)
		copy(publicKeys[i], relayPublicKey[:])
	}

	sessionId := uint64(0x123131231313131)
	sessionVersion := byte(100)
	kbpsUp := uint32(256)
	kbpsDown := uint32(256)
	expireTimestamp := uint64(time.Now().Unix() + 10)

	tokenData := make([]byte, constants.NEXT_MAX_NODES*constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)

	core.WriteRouteTokens(tokenData, expireTimestamp, sessionId, sessionVersion, kbpsUp, kbpsDown, constants.NEXT_MAX_NODES, publicAddresses, hasInternalAddresses, internalAddresses, internalGroups, sellers, publicKeys, masterPrivateKey)

	// read each token back individually and verify the token data matches what was written

	for i := 0; i < constants.NEXT_MAX_NODES; i++ {
		var routeToken core.RouteToken
		err := core.ReadEncryptedRouteToken(&routeToken, tokenData[i*constants.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], masterPublicKey[:], relayPrivateKey[:])
		assert.NoError(t, err)
		assert.Equal(t, sessionId, routeToken.SessionId)
		assert.Equal(t, sessionVersion, routeToken.SessionVersion)
		assert.Equal(t, kbpsUp, routeToken.KbpsUp)
		assert.Equal(t, kbpsDown, routeToken.KbpsDown)
		assert.Equal(t, expireTimestamp, routeToken.ExpireTimestamp)
		if i != 0 {
			assert.Equal(t, publicAddresses[i-1].String(), routeToken.PrevAddress.String())
		}
		if i != constants.NEXT_MAX_NODES-1 {
			assert.Equal(t, publicAddresses[i+1].String(), routeToken.NextAddress.String())
		}
		assert.Equal(t, routeToken.NextInternal, uint8(0))
		assert.Equal(t, routeToken.PrevInternal, uint8(0))
		assert.Equal(t, publicKeys[i], relayPublicKey[:])
	}
}

func TestContinueToken(t *testing.T) {

	t.Parallel()

	relayPublicKey, relayPrivateKey := crypto.Box_KeyPair()

	masterPublicKey, masterPrivateKey := crypto.Box_KeyPair()

	// write a continue token and verify we can read it back

	continueToken := core.ContinueToken{}
	continueToken.ExpireTimestamp = uint64(time.Now().Unix() + 10)
	continueToken.SessionId = 0x123131231313131
	continueToken.SessionVersion = 100

	buffer := make([]byte, constants.NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES)

	core.WriteContinueToken(&continueToken, buffer[:])

	var readContinueToken core.ContinueToken

	err := core.ReadContinueToken(&readContinueToken, buffer)

	assert.NoError(t, err)
	assert.Equal(t, continueToken, readContinueToken)

	// read continue token should fail when the buffer is too small

	err = core.ReadContinueToken(&readContinueToken, buffer[:10])

	assert.Error(t, err)

	// write an encrypted continue token and verify we can decrypt and read it back

	core.WriteEncryptedContinueToken(&continueToken, buffer, masterPrivateKey[:], relayPublicKey[:])

	err = core.ReadEncryptedContinueToken(&continueToken, buffer, masterPublicKey[:], relayPrivateKey[:])

	assert.NoError(t, err)
	assert.Equal(t, continueToken, readContinueToken)

	// read encrypted continue token should fail when buffer is too small

	err = core.ReadEncryptedContinueToken(&continueToken, buffer[:10], masterPublicKey[:], relayPrivateKey[:])

	assert.Error(t, err)

	// read encrypted continue token should fail on garbage data

	garbageData := make([]byte, constants.NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES)
	core.RandomBytes(garbageData)

	err = core.ReadEncryptedContinueToken(&continueToken, garbageData, masterPublicKey[:], relayPrivateKey[:])

	assert.Error(t, err)
}

func TestContinueTokens(t *testing.T) {

	t.Parallel()

	relayPublicKey, relayPrivateKey := crypto.Box_KeyPair()

	masterPublicKey, masterPrivateKey := crypto.Box_KeyPair()

	// write a bunch of tokens to a buffer

	publicKeys := make([][]byte, constants.NEXT_MAX_NODES)
	for i := range publicKeys {
		publicKeys[i] = make([]byte, crypto.Box_PublicKeySize)
		copy(publicKeys[i], relayPublicKey[:])
	}

	sessionId := uint64(0x123131231313131)
	sessionVersion := byte(100)
	expireTimestamp := uint64(time.Now().Unix() + 10)

	tokenData := make([]byte, constants.NEXT_MAX_NODES*constants.NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES)

	core.WriteContinueTokens(tokenData, expireTimestamp, sessionId, sessionVersion, constants.NEXT_MAX_NODES, publicKeys, masterPrivateKey)

	// read each token back individually and verify the token data matches what was written

	for i := 0; i < constants.NEXT_MAX_NODES; i++ {
		var routeToken core.ContinueToken
		err := core.ReadEncryptedContinueToken(&routeToken, tokenData[i*constants.NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES:], masterPublicKey[:], relayPrivateKey[:])
		assert.NoError(t, err)
		assert.Equal(t, sessionId, routeToken.SessionId)
		assert.Equal(t, sessionVersion, routeToken.SessionVersion)
		assert.Equal(t, expireTimestamp, routeToken.ExpireTimestamp)
	}
}

func TestBestRouteCostReallySimple(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetCost("losangeles", "chicago", 100)
	env.SetCost("losangeles", "a", 10)
	env.SetCost("a", "chicago", 10)

	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(30+constants.CostBias), bestRouteCost)
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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(40+constants.CostBias), bestRouteCost)
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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelays := []string{"losangeles.a", "losangeles.b", "chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{10, 5, 100, 100}

	destRelays := []string{"chicago.a", "chicago.b"}

	bestRouteCost := env.GetBestRouteCost(routeMatrix, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(15+constants.CostBias), bestRouteCost)
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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	routeRelays := []string{"losangeles", "a", "b", "chicago"}

	sourceRelays := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"chicago"}

	currentRouteExists := env.RouteExists(routeMatrix, routeRelays)

	assert.Equal(t, true, currentRouteExists)

	currentRouteCost := env.GetCurrentRouteCost(routeMatrix, routeRelays, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(40+constants.CostBias), currentRouteCost)
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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	routeRelays := []string{"chicago", "b", "a", "losangeles"}

	sourceRelays := []string{"chicago"}
	sourceRelayCosts := []int32{10}

	destRelays := []string{"losangeles"}

	currentRouteExists := env.RouteExists(routeMatrix, routeRelays)

	assert.Equal(t, true, currentRouteExists)

	currentRouteCost := env.GetCurrentRouteCost(routeMatrix, routeRelays, sourceRelays, sourceRelayCosts, destRelays)

	assert.Equal(t, int32(40+constants.CostBias), currentRouteCost)
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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(20)

	bestRoute := env.GetRandomBestRoute(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute != nil)
	assert.True(t, bestRoute.cost > 0)
	assert.True(t, bestRoute.cost <= maxCost)
	assert.True(t, bestRoute.cost == 12+constants.CostBias || bestRoute.cost == 17+constants.CostBias)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelayNames := []string{"chicago.a", "chicago.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"losangeles.a", "losangeles.b"}

	maxCost := int32(17)

	bestRoute := env.GetRandomBestRoute(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRoute != nil)
	assert.True(t, bestRoute.cost > 0)
	assert.True(t, bestRoute.cost <= maxCost)
	assert.True(t, bestRoute.cost == 12+constants.CostBias || bestRoute.cost == 17+constants.CostBias)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

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

func TestReframeDestRelays(t *testing.T) {

	t.Parallel()

	relayIdToIndex := make(map[uint64]int32)
	relayIdToIndex[1] = 0
	relayIdToIndex[2] = 1
	relayIdToIndex[3] = 2
	relayIdToIndex[4] = 3
	relayIdToIndex[5] = 4
	relayIdToIndex[6] = 5

	inputDestRelayIds := [...]uint64{4, 5, 6, 7}

	outputNumDestRelays := 0
	outputDestRelays := make([]int32, len(inputDestRelayIds))

	core.ReframeDestRelays(relayIdToIndex, inputDestRelayIds[:], &outputNumDestRelays, outputDestRelays[:])

	assert.Equal(t, outputNumDestRelays, 3)
	assert.Equal(t, outputDestRelays[0], int32(3))
	assert.Equal(t, outputDestRelays[1], int32(4))
	assert.Equal(t, outputDestRelays[2], int32(5))
}

func TestReframeSourceRelays(t *testing.T) {

	t.Parallel()

	relayIdToIndex := make(map[uint64]int32)
	relayIdToIndex[1] = 0
	relayIdToIndex[2] = 1
	relayIdToIndex[3] = 2
	relayIdToIndex[4] = 3
	relayIdToIndex[5] = 4
	relayIdToIndex[6] = 5
	relayIdToIndex[7] = 6
	relayIdToIndex[8] = 7
	relayIdToIndex[9] = 8

	inputSourceRelayIds := [...]uint64{4, 5, 6, 7, 10}
	inputSourceRelayLatency := [...]int32{100, 10, 0, 300, 10}

	outputSourceRelays := make([]int32, len(inputSourceRelayIds))
	outputSourceRelayLatency := make([]int32, len(inputSourceRelayIds))

	core.ReframeSourceRelays(relayIdToIndex, inputSourceRelayIds[:], inputSourceRelayLatency[:], outputSourceRelays[:], outputSourceRelayLatency[:])

	assert.Equal(t, outputSourceRelays[0], int32(3))
	assert.Equal(t, outputSourceRelays[1], int32(4))
	assert.Equal(t, outputSourceRelays[2], int32(-1))
	assert.Equal(t, outputSourceRelays[3], int32(-1))
	assert.Equal(t, outputSourceRelays[4], int32(-1))

	assert.Equal(t, outputSourceRelayLatency[0], int32(100))
	assert.Equal(t, outputSourceRelayLatency[1], int32(10))
	assert.Equal(t, outputSourceRelayLatency[2], int32(255))
	assert.Equal(t, outputSourceRelayLatency[3], int32(255))
	assert.Equal(t, outputSourceRelayLatency[4], int32(255))
}

func TestEarlyOutDirect(t *testing.T) {

	var debug string

	userId := uint64(100)

	routeShader := core.NewRouteShader()
	routeState := core.RouteState{}
	assert.False(t, core.EarlyOutDirect(userId, &routeShader, &routeState, &debug))

	routeState = core.RouteState{Veto: true}
	assert.True(t, core.EarlyOutDirect(userId, &routeShader, &routeState, &debug))

	routeState = core.RouteState{LocationVeto: true}
	assert.True(t, core.EarlyOutDirect(userId, &routeShader, &routeState, &debug))

	routeState = core.RouteState{Disabled: true}
	assert.True(t, core.EarlyOutDirect(userId, &routeShader, &routeState, &debug))

	routeState = core.RouteState{NotSelected: true}
	assert.True(t, core.EarlyOutDirect(userId, &routeShader, &routeState, &debug))

	routeState = core.RouteState{B: true}
	assert.True(t, core.EarlyOutDirect(userId, &routeShader, &routeState, &debug))

	routeShader = core.NewRouteShader()
	routeShader.DisableNetworkNext = true
	routeState = core.RouteState{}
	assert.True(t, core.EarlyOutDirect(userId, &routeShader, &routeState, &debug))
	assert.True(t, routeState.Disabled)

	routeShader = core.NewRouteShader()
	routeShader.AnalysisOnly = true
	routeState = core.RouteState{}
	assert.True(t, core.EarlyOutDirect(userId, &routeShader, &routeState, &debug))
	assert.True(t, routeState.Disabled)

	routeShader = core.NewRouteShader()
	routeShader.SelectionPercent = 0
	routeState = core.RouteState{}
	assert.True(t, core.EarlyOutDirect(userId, &routeShader, &routeState, &debug))
	assert.True(t, routeState.NotSelected)

	routeShader = core.NewRouteShader()
	routeShader.SelectionPercent = 0
	routeState = core.RouteState{}
	assert.True(t, core.EarlyOutDirect(userId, &routeShader, &routeState, &debug))
	assert.True(t, routeState.NotSelected)

	routeShader = core.NewRouteShader()
	routeShader.ABTest = true
	routeState = core.RouteState{}
	assert.False(t, core.EarlyOutDirect(0, &routeShader, &routeState, &debug))
	assert.True(t, routeState.ABTest)
	assert.True(t, routeState.A)
	assert.False(t, routeState.B)

	routeShader = core.NewRouteShader()
	routeShader.ABTest = true
	routeState = core.RouteState{}
	assert.True(t, core.EarlyOutDirect(1, &routeShader, &routeState, &debug))
	assert.True(t, routeState.ABTest)
	assert.False(t, routeState.A)
	assert.True(t, routeState.B)
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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{5}

	destRelayNames := []string{"chicago"}

	maxCost := int32(40)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.Equal(t, int32(35+constants.CostBias), bestRouteCost)
	assert.Equal(t, int32(1), routeDiversity)
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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(20)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

	assert.True(t, bestRouteCost > 0)
	assert.True(t, bestRouteCost <= maxCost)
	assert.True(t, bestRouteCost == 12+constants.CostBias || bestRouteCost == 17+constants.CostBias)

	if bestRouteCost == 12 {
		assert.Equal(t, []string{"losangeles.b", "b", "chicago.b"}, bestRouteRelays)
	}

	if bestRouteCost == 17 {
		assert.Equal(t, []string{"losangeles.b", "b", "chicago.a"}, bestRouteRelays)
	}

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(1)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelayNames := []string{"losangeles.a", "losangeles.b"}
	sourceRelayCosts := []int32{5, 2}

	destRelayNames := []string{"chicago.a", "chicago.b"}

	maxCost := int32(-1)

	bestRouteCost, routeDiversity, bestRouteRelays := env.GetBestRoute_Initial(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{10}

	destRelayNames := []string{"chicago"}

	maxCost := int32(1000)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

	assert.Equal(t, int32(40+constants.CostBias), bestRouteCost)
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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{1}

	destRelayNames := []string{"chicago"}

	maxCost := int32(5)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

	assert.Equal(t, int32(2+constants.CostBias), bestRouteCost)
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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{1}

	destRelayNames := []string{"chicago"}

	maxCost := int32(5)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

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

	routeMatrix := core.Optimize(numRelays, numSegments, costMatrix, relayDatacenters)

	sourceRelayNames := []string{"losangeles"}
	sourceRelayCosts := []int32{1}

	destRelayNames := []string{"chicago"}

	maxCost := int32(-1)

	selectThreshold := int32(2)
	switchThreshold := int32(5)

	currentRoute := []string{"losangeles", "a", "b", "chicago"}

	bestRouteCost, bestRouteRelays := env.GetBestRoute_Update(routeMatrix, sourceRelayNames, sourceRelayCosts, destRelayNames, maxCost, selectThreshold, switchThreshold, currentRoute)

	assert.Equal(t, int32(0), bestRouteCost)
	assert.Equal(t, []string{}, bestRouteRelays)
}

// -------------------------------------------------------------------------------

type TestData struct {
	numRelays        int
	relayNames       []string
	relayDatacenters []uint64
	costMatrix       []uint8
	routeMatrix      []core.RouteEntry

	directLatency    int32
	directPacketLoss float32

	sourceRelays     []int32
	sourceRelayCosts []int32

	destRelays []int32

	routeCost      int32
	routeNumRelays int32
	routeRelays    [constants.MaxRouteRelays]int32

	routeShader        core.RouteShader
	routeState         core.RouteState
	multipathVetoUsers map[uint64]bool

	debug string

	routeDiversity int32

	nextLatency           int32
	nextPacketLoss        float32
	predictedLatency      int32
	currentRouteNumRelays int32
	currentRouteRelays    [constants.MaxRouteRelays]int32

	sliceNumber int32

	realPacketLoss float32

	userId uint64
}

func NewTestData(env *TestEnvironment) *TestData {

	test := &TestData{}

	test.costMatrix, test.numRelays = env.GetCostMatrix()

	test.relayNames = env.GetRelayNames()

	test.relayDatacenters = env.GetRelayDatacenters()

	numSegments := test.numRelays
	test.routeMatrix = core.Optimize(test.numRelays, numSegments, test.costMatrix, test.relayDatacenters)
	test.routeShader = core.NewRouteShader()

	test.multipathVetoUsers = map[uint64]bool{}

	test.userId = 100

	return test
}

func (test *TestData) TakeNetworkNext() bool {
	return core.MakeRouteDecision_TakeNetworkNext(test.userId,
		test.routeMatrix,
		&test.routeShader,
		&test.routeState,
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
	return core.MakeRouteDecision_StayOnNetworkNext(test.userId,
		test.routeMatrix,
		test.relayNames,
		&test.routeShader,
		&test.routeState,
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

	test.routeState.Veto = true

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Veto = true

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

	test.routeState.Disabled = true

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := core.RouteState{}
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

	test.routeState.NotSelected = true

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := core.RouteState{}
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

	test.routeState.B = true

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := core.RouteState{}
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

	test.routeShader.Multipath = false
	test.routeShader.AcceptablePacketLossSustained = float32(10.0)

	test.sliceNumber = 1

	test.destRelays = []int32{1}

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true

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

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.Multipath = true

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

	test.routeShader.RouteDiversity = 5

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := core.RouteState{}
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

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := core.RouteState{}

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

	test.routeShader.LatencyReductionThreshold = 20

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := core.RouteState{}

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(0), test.routeDiversity)
}

func TestTakeNetworkNext_ReduceLatency_MaxRTT(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 251)

	test := NewTestData(env)

	test.directLatency = int32(252)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{1}

	test.destRelays = []int32{1}

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := core.RouteState{}

	assert.Equal(t, expectedRouteState, test.routeState)
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
	test.routeShader.Multipath = false
	test.routeShader.AcceptablePacketLossSustained = float32(10.0)

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true

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
	test.routeShader.Multipath = false
	test.routeShader.AcceptablePacketLossSustained = float32(10.0)

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true

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

	test.routeShader.Multipath = false
	test.routeShader.AcceptablePacketLossSustained = float32(10.0)

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := core.RouteState{}

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

	test.routeShader.Multipath = false
	test.routeShader.AcceptablePacketLossSustained = float32(10.0)

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.ReducePacketLoss = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ReducePacketLoss_MaxRTT(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 250)

	test := NewTestData(env)

	test.directLatency = int32(251)
	test.directPacketLoss = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.AcceptableLatency = 100

	test.routeShader.AcceptablePacketLossSustained = float32(10.0)

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := core.RouteState{}

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
	test.routeShader.AcceptablePacketLossInstant = float32(20)

	// Will go next after 3 slices of sustained packet loss
	test.routeShader.AcceptablePacketLossSustained = float32(2.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(1), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(2), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.True(t, result)
	assert.Equal(t, uint32(3), test.routeState.PLSustainedCounter)
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
	test.routeShader.AcceptablePacketLossInstant = float32(20)

	// Will go next after 3 slices of sustained packet loss
	test.routeShader.AcceptablePacketLossSustained = float32(5.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(1), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(2), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.True(t, result)
	assert.Equal(t, uint32(3), test.routeState.PLSustainedCounter)
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
	test.routeShader.AcceptablePacketLossInstant = float32(20)

	// Won't go next after 3 slices of sustained packet loss
	test.routeShader.AcceptablePacketLossSustained = float32(10.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(0), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(0), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(0), test.routeState.PLSustainedCounter)
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
	test.routeShader.AcceptablePacketLossInstant = float32(20)

	// Will go next after 3 slices of sustained packet loss
	test.routeShader.AcceptablePacketLossSustained = float32(2.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(1), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(2), test.routeState.PLSustainedCounter)

	test.directPacketLoss = 1

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(0), test.routeState.PLSustainedCounter)
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
	test.routeShader.AcceptablePacketLossInstant = float32(20)

	// Will go next after 3 slices of sustained packet loss
	test.routeShader.AcceptablePacketLossSustained = float32(2.0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(1), test.routeState.PLSustainedCounter)

	test.directPacketLoss = 1

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(0), test.routeState.PLSustainedCounter)

	test.directPacketLoss = 5

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(1), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(2), test.routeState.PLSustainedCounter)

	test.directPacketLoss = 1

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(0), test.routeState.PLSustainedCounter)

	test.directPacketLoss = 5

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(1), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.False(t, result)
	assert.Equal(t, uint32(2), test.routeState.PLSustainedCounter)

	result = test.TakeNetworkNext()

	assert.True(t, result)
	assert.Equal(t, uint32(3), test.routeState.PLSustainedCounter)
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

	test.routeShader.Multipath = true

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReduceLatency = true

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

	test.routeShader.AcceptablePacketLossSustained = float32(10.0)

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true

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

	test.routeShader.Multipath = true
	test.routeShader.AcceptablePacketLossSustained = float32(10.0)

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.ReducePacketLoss = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(1), test.routeDiversity)
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
	test.routeState.ReduceLatency = true
	test.routeState.Veto = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.ReduceLatency = true
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
	test.routeState.ReduceLatency = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true

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
	test.routeState.ReduceLatency = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true

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
	test.routeRelays = [constants.MaxRouteRelays]int32{}

	test.routeState.Next = true
	test.routeState.ReduceLatency = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = false
	expectedRouteState.ReduceLatency = true
	expectedRouteState.LatencyWorse = true
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
	test.routeState.ReduceLatency = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = false
	expectedRouteState.ReduceLatency = true
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
	test.routeState.ReduceLatency = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	// first slice mispredicting is fine

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.MispredictCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)

	// first slice mispredicting is fine

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.MispredictCounter = 2

	assert.Equal(t, expectedRouteState, test.routeState)

	// third slice mispredicting is veto

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = core.RouteState{}
	expectedRouteState.Next = false
	expectedRouteState.ReduceLatency = true
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
	test.routeState.ReduceLatency = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	// first slice mispredicting is fine

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
	expectedRouteState.MispredictCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)

	// check that we recover when no longer mispredicting

	test.predictedLatency = int32(100)

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true
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
	env.SetCost("losangeles", "chicago", 250)

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
	test.routeState.ReduceLatency = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.True(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(12+constants.CostBias), test.routeCost)
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
	test.routeState.ReduceLatency = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.True(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReduceLatency = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(12+constants.CostBias), test.routeCost)
	assert.Equal(t, int32(3), test.routeNumRelays)
}

func TestStayOnNetworkNext_ReduceLatency_MaxRTT(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 250)

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
	test.routeState.ReduceLatency = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = false
	expectedRouteState.Veto = true
	expectedRouteState.ReduceLatency = true
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
	test.routeState.ReducePacketLoss = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true

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
	test.routeState.ReducePacketLoss = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = false
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.LatencyWorse = true
	expectedRouteState.Veto = true

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
	test.routeState.ReducePacketLoss = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ReducePacketLoss = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ReducePacketLoss_MaxRTT(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 250)

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
	test.routeState.ReducePacketLoss = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = false
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.NextLatencyTooHigh = true
	expectedRouteState.Veto = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

// -----------------------------------------------------------------------------

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
	test.routeState.Multipath = true
	test.routeState.ReducePacketLoss = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true

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
	test.routeState.Multipath = true
	test.routeState.ReducePacketLoss = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	// first latency worse is fine

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.LatencyWorseCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)

	// second latency worse is fine

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.LatencyWorseCounter = 2

	assert.Equal(t, expectedRouteState, test.routeState)

	// third latency worse is veto

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = core.RouteState{}
	expectedRouteState.Next = false
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.LatencyWorse = true
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

	test.nextLatency = int32(100)

	test.predictedLatency = int32(0)

	test.directPacketLoss = float32(0)

	test.nextPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.Multipath = true
	test.routeState.ReducePacketLoss = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	// first latency worse is fine

	result, nextRouteSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
	expectedRouteState.LatencyWorseCounter = 1

	assert.Equal(t, expectedRouteState, test.routeState)

	// now latency is not worse, we should recover

	test.nextLatency = int32(1)

	result, nextRouteSwitched = test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, nextRouteSwitched)

	expectedRouteState = core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.Multipath = true
	expectedRouteState.ReducePacketLoss = true
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

	test.directLatency = int32(10)

	test.directPacketLoss = float32(0)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeShader.ForceNext = true

	test.routeState.Next = false

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.True(t, result)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ForcedNext = true
	expectedRouteState.Multipath = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(50+constants.CostBias), test.routeCost)
	assert.Equal(t, int32(2), test.routeNumRelays)
	assert.Equal(t, int32(1), test.routeDiversity)
}

func TestTakeNetworkNext_ForceNext_NoRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.routeShader.ForceNext = true

	test.routeState.Next = false

	test.sliceNumber = 1

	result := test.TakeNetworkNext()

	assert.False(t, result)

	expectedRouteState := core.RouteState{}
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

	test.directLatency = int32(30)

	test.nextLatency = int32(60)

	test.predictedLatency = int32(0)

	test.nextPacketLoss = float32(5)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{10}

	test.destRelays = []int32{1}

	test.routeState.Next = true
	test.routeState.ForcedNext = true
	test.routeShader.ForceNext = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.sliceNumber = 1

	result, routeSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.False(t, routeSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ForcedNext = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(50+constants.CostBias), test.routeCost)
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

	test.routeShader.ForceNext = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.routeState.Next = true
	test.routeState.ForcedNext = true

	test.sliceNumber = 1

	result, routeSwitched := test.StayOnNetworkNext()

	assert.False(t, result)
	assert.False(t, routeSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.ForcedNext = true
	expectedRouteState.Veto = true
	expectedRouteState.NoRoute = true
	expectedRouteState.RouteLost = true

	assert.Equal(t, expectedRouteState, test.routeState)
}

func TestStayOnNetworkNext_ForceNext_RouteSwitched(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")

	env.SetCost("losangeles", "chicago", 250)
	env.SetCost("losangeles", "a", 1)
	env.SetCost("a", "chicago", 1)

	test := NewTestData(env)

	test.directLatency = int32(30)

	test.nextLatency = int32(1)

	test.sourceRelays = []int32{0}
	test.sourceRelayCosts = []int32{1}

	test.destRelays = []int32{1}

	test.routeShader.ForceNext = true

	test.currentRouteNumRelays = int32(2)
	test.currentRouteRelays = [constants.MaxRouteRelays]int32{0, 1}

	test.routeState.Next = true
	test.routeState.ForcedNext = true

	test.sliceNumber = 1

	result, routeSwitched := test.StayOnNetworkNext()

	assert.True(t, result)
	assert.True(t, routeSwitched)

	expectedRouteState := core.RouteState{}
	expectedRouteState.Next = true
	expectedRouteState.ForcedNext = true

	assert.Equal(t, expectedRouteState, test.routeState)
	assert.Equal(t, int32(3+constants.CostBias), test.routeCost)
	assert.Equal(t, int32(3), test.routeNumRelays)
}

// -------------------------------------------------------------

// todo: update tests
/*
func TestFilterSourceRelays_NoFilter(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("a", "10.0.0.1")
	env.AddRelay("b", "10.0.0.2")
	env.AddRelay("c", "10.0.0.3")
	env.AddRelay("d", "10.0.0.4")
	env.AddRelay("e", "10.0.0.5")

	relayIds := env.GetRelayIds()

	relayIdToIndex := env.GetRelayIdToIndex()

	directLatency := int32(25)
	directJitter := int32(0)
	directPacketLoss := float32(0)

	sourceRelayIds := relayIds
	sourceRelayLatency := []int32{1, 1, 1, 1, 1}
	sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	sourceRelayPacketLoss := []float32{0, 0, 0, 0, 0}

	outputSourceRelayLatency := [constants.MaxNearRelays]int32{}

	core.FilterSourceRelays(relayIdToIndex,
		directLatency,
		directJitter,
		directPacketLoss,
		sourceRelayIds,
		sourceRelayLatency,
		sourceRelayJitter,
		sourceRelayPacketLoss,
		false,
		outputSourceRelayLatency[:])

	for i := range sourceRelayIds {
		assert.Equal(t, outputSourceRelayLatency[i], int32(1))
	}
}

func TestFilterSourceRelays_ZeroLatency(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("a", "10.0.0.1")
	env.AddRelay("b", "10.0.0.2")
	env.AddRelay("c", "10.0.0.3")
	env.AddRelay("d", "10.0.0.4")
	env.AddRelay("e", "10.0.0.5")

	relayIds := env.GetRelayIds()

	relayIdToIndex := env.GetRelayIdToIndex()

	directLatency := int32(25)
	directJitter := int32(0)
	directPacketLoss := float32(0)

	sourceRelayIds := relayIds
	sourceRelayLatency := []int32{1, 1, 0, 1, 1}
	sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	sourceRelayPacketLoss := []float32{0, 0, 0, 0, 0}

	outputSourceRelayLatency := [constants.MaxNearRelays]int32{}

	core.FilterSourceRelays(relayIdToIndex,
		directLatency,
		directJitter,
		directPacketLoss,
		sourceRelayIds,
		sourceRelayLatency,
		sourceRelayJitter,
		sourceRelayPacketLoss,
		false,
		outputSourceRelayLatency[:])

	for i := range sourceRelayIds {
		if i != 2 {
			assert.Equal(t, outputSourceRelayLatency[i], int32(1))
		} else {
			assert.Equal(t, outputSourceRelayLatency[i], int32(255))
		}
	}
}

func TestFilterSourceRelays_ClampLatencyAbove255(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("a", "10.0.0.1")
	env.AddRelay("b", "10.0.0.2")
	env.AddRelay("c", "10.0.0.3")
	env.AddRelay("d", "10.0.0.4")
	env.AddRelay("e", "10.0.0.5")

	relayIds := env.GetRelayIds()

	relayIdToIndex := env.GetRelayIdToIndex()

	directLatency := int32(25)
	directJitter := int32(0)
	directPacketLoss := float32(0)

	sourceRelayIds := relayIds
	sourceRelayLatency := []int32{1, 1, 300, 1, 1}
	sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	sourceRelayPacketLoss := []float32{0, 0, 0, 0, 0}

	outputSourceRelayLatency := [constants.MaxNearRelays]int32{}

	core.FilterSourceRelays(relayIdToIndex,
		directLatency,
		directJitter,
		directPacketLoss,
		sourceRelayIds,
		sourceRelayLatency,
		sourceRelayJitter,
		sourceRelayPacketLoss,
		false,
		outputSourceRelayLatency[:])

	for i := range sourceRelayIds {
		if i != 2 {
			assert.Equal(t, outputSourceRelayLatency[i], int32(1))
		} else {
			assert.Equal(t, outputSourceRelayLatency[i], int32(255))
		}
	}
}

func TestFilterSourceRelays_RelayDoesNotExist(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("a", "10.0.0.1")
	env.AddRelay("b", "10.0.0.2")
	env.AddRelay("c", "10.0.0.3")
	env.AddRelay("d", "10.0.0.4")
	env.AddRelay("e", "10.0.0.5")

	relayIds := env.GetRelayIds()

	relayIdToIndex := env.GetRelayIdToIndex()

	directLatency := int32(10)
	directJitter := int32(0)
	directPacketLoss := float32(0)

	sourceRelayIds := relayIds
	sourceRelayLatency := []int32{1, 1, 1, 1, 1}
	sourceRelayJitter := []int32{0, 0, 0, 0, 0}
	sourceRelayPacketLoss := []float32{0, 0, 0, 0, 0}

	sourceRelayIds[2] = 423189384

	outputSourceRelayLatency := [constants.MaxNearRelays]int32{}

	core.FilterSourceRelays(relayIdToIndex,
		directLatency,
		directJitter,
		directPacketLoss,
		sourceRelayIds,
		sourceRelayLatency,
		sourceRelayJitter,
		sourceRelayPacketLoss,
		false,
		outputSourceRelayLatency[:])

	for i := range sourceRelayIds {
		if i != 2 {
			assert.Equal(t, outputSourceRelayLatency[i], int32(1))
		} else {
			assert.Equal(t, outputSourceRelayLatency[i], int32(255))
		}
	}
}
*/

// -------------------------------------------------------------

func randomBytes(buffer []byte) {
	for i := 0; i < len(buffer); i++ {
		buffer[i] = byte(rand.Intn(256))
	}
}

func TestABI(t *testing.T) {

	var output [1024]byte

	magic := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	fromAddress := [4]byte{1, 2, 3, 4}
	toAddress := [4]byte{4, 3, 2, 1}
	packetLength := 1000

	core.GeneratePittle(output[:], fromAddress[:], toAddress[:], packetLength)

	assert.Equal(t, output[0], uint8(0x3f))
	assert.Equal(t, output[1], uint8(0xb1))

	core.GenerateChonkle(output[:], magic[:], fromAddress[:], toAddress[:], packetLength)

	assert.Equal(t, output[0], uint8(0x2a))
	assert.Equal(t, output[1], uint8(0xd0))
	assert.Equal(t, output[2], uint8(0x1e))
	assert.Equal(t, output[3], uint8(0x4c))
	assert.Equal(t, output[4], uint8(0x4e))
	assert.Equal(t, output[5], uint8(0xdc))
	assert.Equal(t, output[6], uint8(0x9f))
	assert.Equal(t, output[7], uint8(0x07))
}

func TestPittleAndChonkle(t *testing.T) {
	rand.Seed(42)
	var output [constants.MaxPacketBytes]byte
	output[0] = 0x32
	iterations := 10000
	for i := 0; i < iterations; i++ {
		var magic [8]byte
		var fromAddress [4]byte
		var toAddress [4]byte
		randomBytes(magic[:])
		randomBytes(fromAddress[:])
		randomBytes(toAddress[:])
		packetLength := 18 + (i % (len(output) - 18))
		core.GeneratePittle(output[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(output[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		assert.Equal(t, true, core.BasicPacketFilter(output[:], packetLength))
		assert.Equal(t, true, core.AdvancedPacketFilter(output[:], magic[:], fromAddress[:], toAddress[:], packetLength))
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
		assert.Equal(t, false, core.BasicPacketFilter(output[:], packetLength))
	}
	assert.Equal(t, 0, pass)
}

func TestAdvancedBasicPacketFilter(t *testing.T) {
	rand.Seed(42)
	var output [constants.MaxPacketBytes]byte
	iterations := 10000
	for i := 0; i < iterations; i++ {
		var magic [8]byte
		var fromAddress [4]byte
		var toAddress [4]byte
		randomBytes(magic[:])
		randomBytes(fromAddress[:])
		randomBytes(toAddress[:])
		randomBytes(output[:])
		packetLength := i % len(output)
		assert.Equal(t, false, core.BasicPacketFilter(output[:], packetLength))
		assert.Equal(t, false, core.AdvancedPacketFilter(output[:], magic[:], fromAddress[:], toAddress[:], packetLength))
	}
}

func TestPingTokenSignatures(t *testing.T) {
	const NumTokens = 32
	key := crypto.Auth_Key()
	clientPublicAddress := common.RandomAddress()
	relayPublicAddresses := make([]net.UDPAddr, NumTokens)
	for i := range relayPublicAddresses {
		address := common.RandomAddress()
		relayPublicAddresses[i] = address
	}
	expireTimestamp := rand.Uint64()
	pingTokens := make([]byte, NumTokens*constants.PingTokenBytes)
	core.GeneratePingTokens(expireTimestamp, &clientPublicAddress, relayPublicAddresses, key, pingTokens)
	for i := 0; i < 32; i++ {
		data := make([]byte, 256)
		binary.LittleEndian.PutUint64(data[0:], expireTimestamp)
		clientAddressWithoutPort := clientPublicAddress
		clientAddressWithoutPort.Port = 0
		core.WriteAddress(data[8:], &clientAddressWithoutPort)
		core.WriteAddress(data[8+constants.NEXT_ADDRESS_BYTES:], &relayPublicAddresses[i])
		length := 8 + constants.NEXT_ADDRESS_BYTES + constants.NEXT_ADDRESS_BYTES
		assert.True(t, crypto.Auth_Verify(data[:length], key, pingTokens[i*constants.PingTokenBytes:]))
	}
}

func TestSessionScore(t *testing.T) {

	// biggest next improvement should be 0 (lowest score)

	assert.True(t, core.GetSessionScore(true, 254, 0) == uint32(0))

	// no next improvement should be 254 (no improvement)

	assert.True(t, core.GetSessionScore(true, 0, 0) == uint32(254))

	// next is worse than direct is still no improvement

	assert.True(t, core.GetSessionScore(true, 100, 200) == uint32(254))

	// biggest direct RTT values come first, after next values with no improvement

	assert.True(t, core.GetSessionScore(false, 1000, 0) == uint32(255))

	// lowest direct RTT values are last

	assert.True(t, core.GetSessionScore(false, 0, 0) == uint32(999))

	// test random direct sessions

	for i := 0; i < 10000; i++ {
		score := core.GetSessionScore(false, int32(rand.Intn(5000)-2000), int32(rand.Intn(5000)-2000))
		assert.True(t, score <= 999)
	}

	// test random next sessions

	for i := 0; i < 10000; i++ {
		score := core.GetSessionScore(true, int32(rand.Intn(5000)-2000), int32(rand.Intn(5000)-2000))
		assert.True(t, score <= 999)
	}
}

func TestPagination(t *testing.T) {

	t.Parallel()

	// if there is nothing in the list, then we should always get page 0 [0,0]

	{
		begin, end, outputPage, numPages := core.DoPagination(100, 0)
		assert.True(t, begin == 0)
		assert.True(t, end == 0)
		assert.True(t, outputPage == 0)
		assert.True(t, numPages == 0)
	}

	// if the list is less than 100 long, then we should always get page 0 [0,length]

	{
		begin, end, outputPage, numPages := core.DoPagination(100, 15)
		assert.True(t, begin == 0)
		assert.True(t, end == 15)
		assert.True(t, outputPage == 0)
		assert.True(t, numPages == 1)
	}

	// if the list is not evenly dividable by 100, we get an extra page at the end

	{
		begin, end, outputPage, numPages := core.DoPagination(0, 1001)
		assert.True(t, begin == 0)
		assert.True(t, end == 100)
		assert.True(t, outputPage == 0)
		assert.True(t, numPages == 11)
	}

	// regular positive get page cases (relative to beginning of list)

	for i := 0; i < 100; i++ {
		begin, end, outputPage, numPages := core.DoPagination(i, 100000)
		assert.True(t, begin == i*100)
		assert.True(t, end == (i+1)*100)
		assert.True(t, outputPage == i)
		assert.True(t, numPages == 1000)
	}

	// regular negative page cases (relative to end of list)

	{
		for i := 1; i < 100; i++ {
			begin, end, outputPage, numPages := core.DoPagination(-i, 100000)
			assert.True(t, end == 100000-(i*100))
			assert.True(t, begin == end-100)
			assert.True(t, outputPage == -i)
			assert.True(t, numPages == 1000)
		}
	}

	// positive pages that go past end, should clamp to page -1

	{
		begin, end, outputPage, numPages := core.DoPagination(100, 1000)
		assert.True(t, begin == 900)
		assert.True(t, end == 1000)
		assert.True(t, outputPage == -1)
		assert.True(t, numPages == 10)
	}

	// negative pages that go past beginning, should clamp to page 0

	{
		begin, end, outputPage, numPages := core.DoPagination(-100, 1000)
		assert.True(t, begin == 0)
		assert.True(t, end == 100)
		assert.True(t, outputPage == 0)
		assert.True(t, numPages == 10)
	}

}
