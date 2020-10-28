package routing

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/networknext/backend/modules/encoding"
)

const (
	VersionNumberRelayMap = 2

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
	MaxSessions    uint32
	CPUUsage       float32
	MemUsage       float32
	Version        string

	// Traffic stats from last update
	TrafficStats TrafficStats

	// Highest values from the traffic stats seen since the last publis interval
	PeakTrafficStats PeakTrafficStats

	// contains all the traffic stats updates since the last publish
	TrafficStatsBuff []TrafficStats

	// for locking access to the traffic stats buffer & peak stats specifically
	TrafficMu sync.Mutex

	// only modified within the stats publish loop, so no need to lock access
	LastStatsPublishTime time.Time
}

func NewRelayData() *RelayData {
	return &RelayData{
		TrafficStatsBuff: make([]TrafficStats, 0),
	}
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

func (rmap *RelayMap) GetRelayCount() uint64 {
	return uint64(len(rmap.relays))
}

// NewRelayData inserts a new entry into the map and returns the pointer
func (rmap *RelayMap) AddRelayDataEntry(relayAddress string, data *RelayData) {
	rmap.relays[relayAddress] = data
}

// UpdateRelayDataEntry updates specific fields that may change per update
func (relayMap *RelayMap) UpdateRelayDataEntry(relayAddress string, newTraffic TrafficStats, cpuUsage float32, memUsage float32) {
	entry := relayMap.relays[relayAddress]
	entry.LastUpdateTime = time.Now()

	entry.TrafficStats = newTraffic
	entry.CPUUsage = cpuUsage
	entry.MemUsage = memUsage

	entry.TrafficMu.Lock()
	entry.PeakTrafficStats = entry.PeakTrafficStats.MaxValues(PeakTrafficStats{
		SessionCount:     newTraffic.SessionCount,
		EnvelopeUpKbps:   newTraffic.EnvelopeUpKbps,
		EnvelopeDownKbps: newTraffic.EnvelopeDownKbps,
	})
	entry.TrafficStatsBuff = append(entry.TrafficStatsBuff, newTraffic)
	entry.TrafficMu.Unlock()
}

func (relayMap *RelayMap) GetRelayData(relayAddress string) *RelayData {
	return relayMap.relays[relayAddress]
}

func (relayMap *RelayMap) GetAllRelayData() []*RelayData {
	relayMap.RLock()
	relays := make([]*RelayData, len(relayMap.relays))

	index := 0
	for _, relayData := range relayMap.relays {
		relays[index] = relayData
		index++
	}

	relayMap.RUnlock()

	return relays
}

func (relayMap *RelayMap) RemoveRelayData(relayAddress string) {
	if relay, ok := relayMap.relays[relayAddress]; ok {
		relayMap.cleanupCallback(relay)
		delete(relayMap.relays, relayAddress)
	}
}

func (relayMap *RelayMap) TimeoutLoop(ctx context.Context, timeoutSeconds int64, c <-chan time.Time) {
	deleteList := make([]string, 0)
	for {
		select {
		case <-c:
			deleteList = deleteList[:0]
			timeoutTimestamp := time.Now().Unix() - timeoutSeconds

			relayMap.RLock()
			for k, v := range relayMap.relays {
				if v.LastUpdateTime.Unix() < timeoutTimestamp {
					deleteList = append(deleteList, k)
				}
			}
			relayMap.RUnlock()

			if len(deleteList) > 0 {
				relayMap.Lock()
				for i := range deleteList {
					relayMap.cleanupCallback(relayMap.relays[deleteList[i]])
					delete(relayMap.relays, deleteList[i])
				}
				relayMap.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}

// | version | count | relay data ... |
func (r *RelayMap) MarshalBinary() ([]byte, error) {
	r.RLock()
	defer r.RUnlock()

	numRelays := r.GetRelayCount()

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
		relay.TrafficStats.WriteTo(data, &index, 2)
		encoding.WriteUint8(data, &index, major)
		encoding.WriteUint8(data, &index, minor)
		encoding.WriteUint8(data, &index, patch)
		encoding.WriteUint64(data, &index, uint64(relay.LastUpdateTime.Unix()))
		encoding.WriteFloat32(data, &index, relay.CPUUsage)
		encoding.WriteFloat32(data, &index, relay.MemUsage)
	}

	return data, nil
}
