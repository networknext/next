package routing

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/networknext/backend/modules/core"
)

type RelayData struct {
	ID                uint64
	Name              string
	Addr              net.UDPAddr
	PublicKey         []byte
	MaxSessions       uint32
	SessionCount      int
	ShuttingDown      bool
	LastUpdateTime    time.Time
	Version           string
	CPU               uint8
	NICSpeedMbps      int32
	EnvelopeUpMbps    float32
	EnvelopeDownMbps  float32
	BandwidthSentMbps float32
	BandwidthRecvMbps float32
}

type RelayCleanupCallback func(relayData RelayData) error

type RelayMap struct {
	relays          map[string]RelayData
	cleanupCallback RelayCleanupCallback
	mutex           sync.RWMutex
}

func NewRelayMap(callback RelayCleanupCallback) *RelayMap {
	relayMap := &RelayMap{
		cleanupCallback: callback,
	}
	relayMap.relays = make(map[string]RelayData)
	return relayMap
}

func (relayMap *RelayMap) Lock() {
	relayMap.mutex.Lock()
}

func (relayMap *RelayMap) Unlock() {
	relayMap.mutex.Unlock()
}

func (relayMap *RelayMap) RLock() {
	relayMap.mutex.RLock()
}

func (relayMap *RelayMap) RUnlock() {
	relayMap.mutex.RUnlock()
}

func (relayMap *RelayMap) GetRelayCount() uint64 {
	return uint64(len(relayMap.relays))
}

func (relayMap *RelayMap) UpdateRelayData(relayData RelayData) {
	relayMap.relays[relayData.Addr.String()] = relayData
}

func (relayMap *RelayMap) GetRelayData(relayAddress string) (RelayData, bool) {
	relayData, ok := relayMap.relays[relayAddress]
	return relayData, ok
}

func (relayMap *RelayMap) GetActiveRelayData() ([]uint64, []int, []string) {
	relayIds := make([]uint64, len(relayMap.relays))
	relaySessionCounts := make([]int, len(relayMap.relays))
	relayVersions := make([]string, len(relayMap.relays))
	relayMap.RLock()
	index := 0
	for _, v := range relayMap.relays {
		if v.ShuttingDown {
			continue
		}
		relayIds[index] = v.ID
		relaySessionCounts[index] = v.SessionCount
		relayVersions[index] = v.Version
		index++
	}
	relayMap.RUnlock()
	relayIds = relayIds[:index]
	relaySessionCounts = relaySessionCounts[:index]
	relayVersions = relayVersions[:index]
	return relayIds, relaySessionCounts, relayVersions
}

// todo: this is really a pretty naff function and we should deprecate it
func (relayMap *RelayMap) GetAllRelayData() []RelayData {
	relays := make([]RelayData, len(relayMap.relays))
	relayMap.RLock()
	index := 0
	for _, relayData := range relayMap.relays {
		relays[index] = relayData
		index++
	}
	relayMap.RUnlock()
	return relays
}

func (relayMap *RelayMap) GetAllRelayAddresses() []string {
	relayMap.RLock()
	defer relayMap.RUnlock()
	relayAddresses := make([]string, 0)
	for _, relayData := range relayMap.relays {
		relayAddresses = append(relayAddresses, relayData.Addr.String())
	}
	return relayAddresses
}

func (relayMap *RelayMap) RemoveRelayData(relayAddress string) {
	if relay, ok := relayMap.relays[relayAddress]; ok {
		relayMap.cleanupCallback(relay)
		delete(relayMap.relays, relayAddress)
	}
}

type RelayStatsEntry struct {
	name         string
	sessionCount int
}

func (relayMap *RelayMap) TimeoutLoop(ctx context.Context, getRelayData func() ([]Relay, map[uint64]Relay), timeoutSeconds int64, c <-chan time.Time) {
	deleteList := make([]string, 0)
	for {
		select {
		case <-c:

			deleteList = deleteList[:0]
			currentTime := time.Now().Unix()
			timeoutTimestamp := currentTime - timeoutSeconds

			relayMap.RLock()
			for k, v := range relayMap.relays {
				timeSinceLastUpdate := currentTime - v.LastUpdateTime.Unix()
				if timeSinceLastUpdate > 10 {
					core.Error("%s: %s hasn't received an update for %d seconds (%d)", v.Addr.String(), v.Name, timeSinceLastUpdate, v.LastUpdateTime.Unix())
				}
				if v.LastUpdateTime.Unix() < timeoutTimestamp {
					core.Error("%s: %s timed out", v.Addr.String(), v.Name)
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
