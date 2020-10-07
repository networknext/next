package routing

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
)

const (
	NumRelayMapShards     = 10
	VersionNumberRelayMap = 1

	RelayDataBytes = 8 + // id
		8 + // sessions
		8 + // tx
		8 + // rx
		8 + // outbound ping tx
		8 + // route request rx
		8 + // route request tx
		8 + // route response rx
		8 + // route response tx
		8 + // client to server rx
		8 + // client to server tx
		8 + // server to client rx
		8 + // server to client tx
		8 + // inbound ping rx
		8 + // inbound ping tx
		8 + // pong rx
		8 + // session ping rx
		8 + // session ping tx
		8 + // session pong rx
		8 + // session pong tx
		8 + // continue request rx
		8 + // continue request tx
		8 + // continue response rx
		8 + // continue response tx
		8 + // near ping rx
		8 + // near ping tx
		8 + // unknown Rx
		1 + // version major
		1 + // version minor
		1 + // version patch
		8 + // last update time
		4 + // cpu usage
		4 // mem usage
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
	numRelaysRightNow := r.GetRelayCount()

	// preallocate the entire buffer size
	data := make([]byte, 1+8+numRelaysRightNow*RelayDataBytes)

	index := 0
	encoding.WriteUint8(data, &index, VersionNumberRelayMap)
	index += 8 // skip the relay count for now

	// since this loops using a range, if one or more relays expire
	// after the number of relays in the map is queried it'll be less
	// than the expected amount which will cause the portal to read
	// garbage data. Manually counting and compareing accounts for that edge case
	var count uint64 = 0
	for i := range r.shard {
		shard := r.shard[i]
		shard.mutex.RLock()
		defer shard.mutex.RUnlock()
		for _, relay := range shard.relays {
			s := strings.Split(relay.Version, ".")
			if len(s) != 3 {
				return nil, fmt.Errorf("invalid relay version for relay %s: %s", relay.Addr.String(), relay.Version)
			}

			var major uint8
			if v, err := strconv.ParseUint(s[0], 10, 32); err == nil {
				major = uint8(v)
			} else {
				return nil, fmt.Errorf("invalid relay major version for relay %s: %s", relay.Addr.String(), s[0])
			}

			var minor uint8
			if v, err := strconv.ParseUint(s[1], 10, 32); err == nil {
				minor = uint8(v)
			} else {
				return nil, fmt.Errorf("invalid relay minor version for relay %s: %s", relay.Addr.String(), s[1])
			}

			var patch uint8
			if v, err := strconv.ParseUint(s[2], 10, 32); err == nil {
				patch = uint8(v)
			} else {
				return nil, fmt.Errorf("invalid relay patch version for relay %s: %s", relay.Addr.String(), s[2])
			}

			encoding.WriteUint64(data, &index, relay.ID)
			relay.TrafficStats.WriteTo(data, &index)
			encoding.WriteUint8(data, &index, major)
			encoding.WriteUint8(data, &index, minor)
			encoding.WriteUint8(data, &index, patch)
			encoding.WriteUint64(data, &index, uint64(relay.LastUpdateTime.Unix()))
			encoding.WriteFloat32(data, &index, relay.CPUUsage)
			encoding.WriteFloat32(data, &index, relay.MemUsage)

			count++

			// if a relay inits into a shard after the current one
			// the number will be greater than the amount of space
			// preallocated so break early and the next update can
			// get the missing relay(s)
			if count > numRelaysRightNow {
				break
			}
		}

		// same reason as above
		if count > numRelaysRightNow {
			break
		}
	}

	// write the count now for accuracy
	index = 1
	encoding.WriteUint64(data, &index, count)

	fmt.Printf("%d relays sent to portal\n", count)

	// truncate the data in case the expire edge case ocurred
	return data[:1+8+count*RelayDataBytes], nil
}
