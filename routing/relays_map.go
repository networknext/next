package routing

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
)

const (
	NumRelayMapShards     = 10
	VersionNumberRelayMap = 0

	// | id (8) | sessions (8) | tx (8) | rx (8) | version strlen (uint32 for size + 5 for the 1.0.x) | last update time (8) | cpu usage (4) | mem usage (4) |
	RelayDataBytes = 8 + 8 + 8 + 8 + 4 + 5 + 8 + 4 + 4
)

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

// RelayCleanupCallback is a callback function that will be called
// right before a relay is timed out from the RelayMap
type RelayCleanupCallback func(relayData *RelayData) error

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
	relays := make([]*RelayData, 0)
	for _, shard := range relayMap.shard {
		shard.mutex.RLock()
		for _, relayData := range shard.relays {
			relays = append(relays, relayData)
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
						relayMap.cleanupCallback(relayMap.shard[index].relays[deleteList[i]])
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

// | version | count | relay stats ... |
func (r *RelayMap) MarshalBinary() ([]byte, error) {
	data := make([]byte, 1+8+r.numRelays*RelayDataBytes)

	index := 0
	encoding.WriteUint8(data, &index, VersionNumberRelayMap)
	encoding.WriteUint64(data, &index, r.numRelays)

	for i := range r.shard {
		shard := r.shard[i]
		shard.mutex.RLock()
		defer shard.mutex.RUnlock()
		for _, relay := range shard.relays {
			encoding.WriteUint64(data, &index, relay.ID)
			encoding.WriteUint64(data, &index, relay.TrafficStats.SessionCount)
			encoding.WriteUint64(data, &index, relay.TrafficStats.BytesSent)
			encoding.WriteUint64(data, &index, relay.TrafficStats.BytesReceived)
			encoding.WriteString(data, &index, relay.Version, uint32(len(relay.Version)))
			encoding.WriteUint64(data, &index, uint64(relay.LastUpdateTime.Unix()))
			encoding.WriteFloat32(data, &index, relay.CPUUsage)
			encoding.WriteFloat32(data, &index, relay.MemUsage)
		}
	}

	return data, nil
}
