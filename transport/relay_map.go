package transport

import (
	"sync"
	"sync/atomic"

	"github.com/networknext/backend/routing"
)

const (
	NumRelayMapShards = 10
)

type RelayMapShard struct {
	mutex  sync.RWMutex
	relays map[uint64]*routing.Relay
}

type RelayMap struct {
	numRelays uint64
	shard     [NumRelayMapShards]*RelayMapShard
}

func NewRelayMap() *RelayMap {
	relayMap := &RelayMap{
		numNextRelaysPerBuyer:   make(map[uint64]uint64),
		numDirectRelaysPerBuyer: make(map[uint64]uint64),
	}
	for i := 0; i < NumRelayMapShards; i++ {
		relayMap.shard[i] = &RelayMapShard{}
		relayMap.shard[i].relays = make(map[uint64]*routing.Relay)
	}
	return relayMap
}

func (relayMap *RelayMap) GetRelayCount() uint64 {
	return atomic.LoadUint64(&relayMap.numRelays)
}

func (relayMap *RelayMap) Lock(relayId uint64) {
	index := relayId % NumRelayMapShards
	relayMap.shard[index].mutex.Lock()
}

func (relayMap *RelayMap) Unlock(relayId uint64) {
	index := relayId % NumRelayMapShards
	relayMap.shard[index].mutex.Unlock()
}

func (relayMap *RelayMap) RLock(relayId uint64) {
	index := relayId % NumRelayMapShards
	relayMap.shard[index].mutex.RLock()
}

func (relayMap *RelayMap) RUnlock(relayId uint64) {
	index := relayId % NumRelayMapShards
	relayMap.shard[index].mutex.RUnlock()
}

func (relayMap *RelayMap) GetRelayIndices() []uint64 {
	var indices []uint64

	for _, s := range relayMap.shard {
		for key, _ := range s.relays {
			indices = append(indices, key)
		}
	}

	return indices
}

func (relayMap *RelayMap) SetRelayData(relay *routing.Relay) error {
	index := relayId % NumRelayMapShards

	relayMap.Lock(index)
	relayMap.shard[index].relays[relayId] = relay
	relayMap.UnLock(index)

	return relayData
}

func (relayMap *RelayMap) GetRelayData(relayId uint64) *routing.Relay {
	index := relayId % NumRelayMapShards

	relayMap.RLock(index)
	relayData, _ := relayMap.shard[index].relays[relayId]
	relayMap.RUnLock(index)

	return relayData
}
