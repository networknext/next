package routing

import (
	"context"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/networknext/backend/modules/core"
)

type RelayData struct {
	ID             uint64
	Name           string
	Addr           net.UDPAddr
	PublicKey      []byte
	Seller         Seller
	Datacenter     Datacenter
	MaxSessions    uint32
	SessionCount   int
	LastUpdateTime time.Time
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

func (relayMap *RelayMap) AddRelayDataEntry(relayAddress string, data RelayData) {
	relayMap.relays[relayAddress] = data
}

func (relayMap *RelayMap) UpdateRelayDataEntry(relayAddress string, sessionCount int) {
	entry := relayMap.relays[relayAddress]
	entry.LastUpdateTime = time.Now()
	entry.SessionCount = sessionCount
	relayMap.relays[relayAddress] = entry
}

func (relayMap *RelayMap) GetRelayData(relayAddress string) (RelayData, bool) {
	relayData, ok := relayMap.relays[relayAddress]
	return relayData, ok
}

// relay indirection below generates compiler error
// func (relayMap *RelayMap) GetCopyRelayData(relayAddress string) RelayData {
// 	relayMap.RLock()
// 	defer relayMap.RUnlock()
// 	relay := relayMap.relays[relayAddress]
// 	return *relay
// }

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

func (relayMap *RelayMap) GetAllRelayIDs(excludeList []string) []uint64 {
	relayIDs := make([]uint64, 0)
	relayMap.RLock()
	defer relayMap.RUnlock()
	if len(excludeList) == 0 {
		for _, relayData := range relayMap.relays {
			relayIDs = append(relayIDs, relayData.ID)
		}
		return relayIDs
	}
	excludeMap := make(map[string]bool)
	for _, exclude := range excludeList {
		excludeMap[exclude] = true
	}
	for _, relayData := range relayMap.relays {
		if _, ok := excludeMap[relayData.Seller.ID]; !ok {
			relayIDs = append(relayIDs, relayData.ID)
		}
	}
	return relayIDs
}

func (relayMap *RelayMap) GetAllRelayAddresses(excludeList []string) []string {
	relayMap.RLock()
	defer relayMap.RUnlock()
	relayAddresses := make([]string, 0)

	if len(excludeList) == 0 {
		for _, relayData := range relayMap.relays {
			relayAddresses = append(relayAddresses, relayData.Addr.String())
		}
		return relayAddresses
	}

	excludeMap := make(map[string]bool)
	for _, exclude := range excludeList {
		excludeMap[exclude] = true
	}

	for _, relayData := range relayMap.relays {
		if _, ok := excludeMap[relayData.Seller.ID]; !ok {
			relayAddresses = append(relayAddresses, relayData.Addr.String())
		}
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

			_, relayHash := getRelayData()

			relayStats := make([]RelayStatsEntry, 0)

			inactiveRelays := make([]string, 0)

			relayMap.RLock()
			for _, v := range relayMap.relays {
				relayStats = append(relayStats, RelayStatsEntry{name: v.Name, sessionCount: v.SessionCount})
			}
			for _, v := range relayHash {
				_, exists := relayMap.relays[v.Addr.String()]
				if !exists {
					inactiveRelays = append(inactiveRelays, v.Name)
				}
			}
			relayMap.RUnlock()

			sort.SliceStable(relayStats, func(i, j int) bool {
				return relayStats[i].name < relayStats[j].name
			})

			sort.SliceStable(relayStats, func(i, j int) bool {
				return relayStats[i].sessionCount > relayStats[j].sessionCount
			})

			sort.SliceStable(inactiveRelays, func(i, j int) bool {
				return inactiveRelays[i] < inactiveRelays[j]
			})

			fmt.Printf("\n-----------------------------------------\n")
			fmt.Printf("\n%d active relays:\n\n", len(relayStats))
			for i := range relayStats {
				fmt.Printf("    %s [%d]\n", relayStats[i].name, relayStats[i].sessionCount)
			}
			fmt.Printf("\n%d inactive relays:\n\n", len(inactiveRelays))
			for i := range inactiveRelays {
				fmt.Printf("    %s\n", inactiveRelays[i])
			}
			fmt.Printf("\n-----------------------------------------\n\n")

			deleteList = deleteList[:0]
			currentTime := time.Now().Unix()
			timeoutTimestamp := currentTime - timeoutSeconds

			relayMap.RLock()
			for k, v := range relayMap.relays {
				timeSinceLastUpdate := currentTime - v.LastUpdateTime.Unix()
				if timeSinceLastUpdate > 10 {
					core.Debug("error: %s - %s hasn't received an update for %d seconds (%d)", v.Addr.String(), v.Name, timeSinceLastUpdate, v.LastUpdateTime.Unix())
				}
				if v.LastUpdateTime.Unix() < timeoutTimestamp {
					core.Debug("error: %s - %s timed out", v.Addr.String(), v.Name)
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
