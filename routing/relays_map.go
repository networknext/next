package routing

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/networknext/backend/crypto"
)

const NumRelayMapShards = 10

type RelayData struct {
	ID             uint64
	Name           string
	Addr           net.UDPAddr
	PublicKey      []byte
	Seller         Seller
	Datacenter     Datacenter
	LastUpdateTime time.Time
	TrafficStats   RelayTrafficStats
	MaxSessions    uint32
	CPUUsage       float32
	MemUsage       float32
	Version        string
}

type RelayMapShard struct {
	mutex  sync.RWMutex
	relays map[string]*RelayData
}

type RelayCleanupCallback func(relayID uint64) error

type RelayMap struct {
	numRelays       uint64
	timeoutShard    int
	shard           [NumRelayMapShards]*RelayMapShard
	cleanupCallback RelayCleanupCallback
}

func NewRelayMap(callback RelayCleanupCallback) *RelayMap {
	relayMap := &RelayMap{
		cleanupCallback: callback,
	}
	for i := 0; i < NumRelayMapShards; i++ {
		relayMap.shard[i] = &RelayMapShard{}
		relayMap.shard[i].relays = make(map[string]*RelayData)
	}
	return relayMap
}

func (relayMap *RelayMap) GetRelayCount() uint64 {
	return atomic.LoadUint64(&relayMap.numRelays)
}

func (relayMap *RelayMap) RLock(relayAddress string) {
	relayHash := crypto.HashID(relayAddress)
	index := relayHash % NumRelayMapShards
	relayMap.shard[index].mutex.RLock()
}

func (relayMap *RelayMap) RUnlock(relayAddress string) {
	relayHash := crypto.HashID(relayAddress)
	index := relayHash % NumRelayMapShards
	relayMap.shard[index].mutex.RUnlock()
}

func (relayMap *RelayMap) Lock(relayAddress string) {
	relayHash := crypto.HashID(relayAddress)
	index := relayHash % NumRelayMapShards
	relayMap.shard[index].mutex.Lock()
}

func (relayMap *RelayMap) Unlock(relayAddress string) {
	relayHash := crypto.HashID(relayAddress)
	index := relayHash % NumRelayMapShards
	relayMap.shard[index].mutex.Unlock()
}

func (relayMap *RelayMap) UpdateRelayData(relayAddress string, relayData *RelayData) {
	relayHash := crypto.HashID(relayAddress)
	index := relayHash % NumRelayMapShards
	_, exists := relayMap.shard[index].relays[relayAddress]
	relayMap.shard[index].relays[relayAddress] = relayData
	if !exists {
		atomic.AddUint64(&relayMap.numRelays, 1)
	}
}

func (relayMap *RelayMap) GetRelayData(relayAddress string) *RelayData {
	relayHash := crypto.HashID(relayAddress)
	index := relayHash % NumRelayMapShards
	relayData, _ := relayMap.shard[index].relays[relayAddress]
	return relayData
}

func (relayMap *RelayMap) GetAllRelayData() []*RelayData {
	relays := make([]*RelayData, relayMap.numRelays)
	var index int

	for _, shard := range relayMap.shard {
		shard.mutex.RLock()
		for _, relayData := range shard.relays {
			relays[index] = relayData
			index++
		}
		shard.mutex.RUnlock()
	}

	return relays
}

func (relayMap *RelayMap) RemoveRelayData(relayAddress string) {
	relayHash := crypto.HashID(relayAddress)
	index := relayHash % NumRelayMapShards
	relayMap.shard[index].mutex.Lock()
	delete(relayMap.shard[index].relays, relayAddress)
	relayMap.shard[index].mutex.Unlock()
}

func (relayMap *RelayMap) TimeoutLoop(ctx context.Context, timeoutSeconds int64, c <-chan time.Time) {
	maxShards := 1
	maxIterations := 10
	deleteList := make([]string, maxIterations)
	for {
		select {
		case <-c:
			timeoutTimestamp := time.Now().Unix() - timeoutSeconds
			for i := 0; i < maxShards; i++ {
				index := (relayMap.timeoutShard + i) % NumRelayMapShards
				deleteList = deleteList[:0]
				relayMap.shard[index].mutex.RLock()
				numIterations := 0
				for k, v := range relayMap.shard[index].relays {
					if numIterations >= maxIterations || numIterations >= len(relayMap.shard[index].relays) {
						break
					}
					if v.LastUpdateTime.Unix() < timeoutTimestamp {
						deleteList = append(deleteList, k)
					}
					numIterations++
				}
				relayMap.shard[index].mutex.RUnlock()
				if len(deleteList) > 0 {
					relayMap.shard[index].mutex.Lock()
					for i := range deleteList {
						relayMap.cleanupCallback(relayMap.shard[index].relays[deleteList[i]].ID)
						// fmt.Printf("timeout relay %x\n", deleteList[i])
						delete(relayMap.shard[index].relays, deleteList[i])
						atomic.AddUint64(&relayMap.numRelays, ^uint64(0))
					}
					relayMap.shard[index].mutex.Unlock()
				}
			}
			relayMap.timeoutShard = (relayMap.timeoutShard + maxShards) % NumRelayMapShards
		case <-ctx.Done():
			return
		}
	}
}
