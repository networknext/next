package test

import (
	"hash/fnv"
	"net"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/routing"
)

func RelayHash64(name string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(name))
	return hash.Sum64()
}

type TestRelayData struct {
	name       string
	address    *net.UDPAddr
	publicKey  []byte
	privateKey []byte
	index      int
	latitude   float32
	longitude  float32
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
	relay.address = core.ParseAddress(relayAddress)
	relay.latitude = 0
	relay.longitude = 0

	var err error
	relay.publicKey, relay.privateKey, err = core.GenerateRelayKeyPair()
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

func (env *TestEnvironment) GetRelayAddresses() []net.UDPAddr {
	relayAddresses := make([]net.UDPAddr, len(env.relayArray))
	for i := range env.relayArray {
		relayAddresses[i] = *env.relayArray[i].address
	}
	return relayAddresses
}

func (env *TestEnvironment) GetRelayLatitudes() []float32 {
	relayLatitudues := make([]float32, len(env.relayArray))
	for i := range env.relayArray {
		relayLatitudues[i] = env.relayArray[i].latitude
	}
	return relayLatitudues
}

func (env *TestEnvironment) GetRelayLongitudes() []float32 {
	relayLongitudes := make([]float32, len(env.relayArray))
	for i := range env.relayArray {
		relayLongitudes[i] = env.relayArray[i].longitude
	}
	return relayLongitudes
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
	entryCount := core.TriMatrixLength(numRelays)
	costMatrix := make([]int32, entryCount)
	for i := 0; i < numRelays; i++ {
		for j := 0; j < i; j++ {
			index := core.TriMatrixIndex(i, j)
			costMatrix[index] = env.cost[i][j]
		}
	}
	return costMatrix, numRelays
}

func (env *TestEnvironment) GetRouteMatrix() *routing.RouteMatrix {
	costMatrix, numRelays := env.GetCostMatrix()

	relayDatacenters := env.GetRelayDatacenters()

	numSegments := numRelays
	costThreshold := int32(5)
	routeEntries := core.Optimize(numRelays, numSegments, costMatrix, costThreshold, relayDatacenters)

	return &routing.RouteMatrix{
		CreatedAt:          uint64(time.Now().Unix()),
		RouteEntries:       routeEntries,
		RelayNames:         env.GetRelayNames(),
		RelayIDs:           env.GetRelayIds(),
		RelayIDsToIndices:  env.GetRelayIdToIndex(),
		RelayAddresses:     env.GetRelayAddresses(),
		RelayLatitudes:     env.GetRelayLatitudes(),
		RelayLongitudes:    env.GetRelayLongitudes(),
		RelayDatacenterIDs: env.GetRelayDatacenters(),
	}
}
