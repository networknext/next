package routing

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/networknext/backend/encoding"
)

const (
	VersionNumberRelayMap = 0

	// | id (8) | sessions (8) | tx (8) | rx (8) | version major (1), minor (1), patch (1) | last update time (8) | cpu usage (4) | mem usage (4) |
	RelayDataBytes = 8 + 8 + 8 + 8 + 1 + 1 + 1 + 8 + 4 + 4
)

type RelayData struct {
	ID               uint64
	Name             string
	Addr             net.UDPAddr
	PublicKey        []byte
	Seller           Seller
	Datacenter       Datacenter
	LastUpdateTime   time.Time
	TrafficStats     RelayTrafficStats
	PeakTrafficStats PeakRelayTrafficStats
	MaxSessions      uint32
	CPUUsage         float32
	MemUsage         float32
	Version          string
}

// RelayCleanupCallback is a callback function that will be called
// right before a relay is timed out from the RelayMap
type RelayCleanupCallback func(relayData *RelayData) error

type RelayMap struct {
	relays          map[string]*RelayData
	cleanupCallback RelayCleanupCallback

	mutex sync.RWMutex
}

func NewRelayMap(callback RelayCleanupCallback) *RelayMap {
	relayMap := &RelayMap{
		cleanupCallback: callback,
	}
	relayMap.relays = make(map[string]*RelayData)
	return relayMap
}

func (rmap *RelayMap) Lock() {
	rmap.mutex.Lock()
}

func (rmap *RelayMap) Unlock() {
	rmap.mutex.Unlock()
}

func (rmap *RelayMap) RLock() {
	rmap.mutex.RLock()
}

func (rmap *RelayMap) RUnlock() {
	rmap.mutex.RUnlock()
}

func (relayMap *RelayMap) GetRelayCount() uint64 {
	count := uint64(len(relayMap.relays))
	return count
}

func (relayMap *RelayMap) UpdateRelayData(relayAddress string, relayData *RelayData) {
	relayMap.relays[relayAddress] = relayData
}

func (relayMap *RelayMap) GetRelayData(relayAddress string) *RelayData {
	return relayMap.relays[relayAddress]
}

func (relayMap *RelayMap) GetAllRelayData() []*RelayData {
	relays := make([]*RelayData, len(relayMap.relays))

	relayMap.mutex.RLock()
	index := 0
	for _, relayData := range relayMap.relays {
		relays[index] = relayData
		index++
	}
	relayMap.mutex.RUnlock()

	return relays
}

func (relayMap *RelayMap) RemoveRelayData(relayAddress string) {
	relayMap.mutex.Lock()
	delete(relayMap.relays, relayAddress)
	relayMap.mutex.Unlock()
}

func (relayMap *RelayMap) TimeoutLoop(ctx context.Context, timeoutSeconds int64, c <-chan time.Time) {
	deleteList := make([]string, 0)
	for {
		select {
		case <-c:
			deleteList = deleteList[:0]
			timeoutTimestamp := time.Now().Unix() - timeoutSeconds

			relayMap.mutex.RLock()
			for k, v := range relayMap.relays {
				fmt.Println("last update time:", v.LastUpdateTime.Unix(), "timeout:", timeoutTimestamp)
				if v.LastUpdateTime.Unix() < timeoutTimestamp {
					deleteList = append(deleteList, k)
				}
			}
			relayMap.mutex.RUnlock()

			if len(deleteList) > 0 {
				relayMap.mutex.Lock()
				for i := range deleteList {
					relayMap.cleanupCallback(relayMap.relays[deleteList[i]])
					delete(relayMap.relays, deleteList[i])
				}
				relayMap.mutex.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}

// | version | count | relay data ... |
func (r *RelayMap) MarshalBinary() ([]byte, error) {
	r.mutex.RLock()
	numRelays := uint64(len(r.relays))

	// preallocate the entire buffer size
	data := make([]byte, 1+8+numRelays*RelayDataBytes)

	index := 0
	encoding.WriteUint8(data, &index, VersionNumberRelayMap)
	encoding.WriteUint64(data, &index, numRelays)

	for _, relay := range r.relays {

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
		encoding.WriteUint64(data, &index, relay.TrafficStats.SessionCount)
		encoding.WriteUint64(data, &index, relay.TrafficStats.BytesSent)
		encoding.WriteUint64(data, &index, relay.TrafficStats.BytesReceived)
		encoding.WriteUint8(data, &index, major)
		encoding.WriteUint8(data, &index, minor)
		encoding.WriteUint8(data, &index, patch)
		encoding.WriteUint64(data, &index, uint64(relay.LastUpdateTime.Unix()))
		encoding.WriteFloat32(data, &index, relay.CPUUsage)
		encoding.WriteFloat32(data, &index, relay.MemUsage)
	}
	r.mutex.RUnlock()

	return data, nil
}
