package common

import (
	"sort"
	"sync"
	"time"

	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/storage"
)

type relayEnabled struct {
	name string
	id   uint64
}

type RelayEnabledCache struct {
	storer       storage.Storer
	mux          sync.RWMutex
	activeRelays []relayEnabled
	shutdown     bool
}

func NewRelayEnabledCache(storer storage.Storer) *RelayEnabledCache {
	rec := new(RelayEnabledCache)
	rec.storer = storer

	return rec
}

func (rec *RelayEnabledCache) StartRunner(interval time.Duration) {
	go func() {
		syncTimer := helpers.NewSyncTimer(interval)
		for !rec.shutdown {
			syncTimer.Run()
			rec.runner()
		}
	}()
}

func (rec *RelayEnabledCache) runner() {
	newActiveRelays := make([]relayEnabled, 0)

	allRelayData := rec.storer.Relays()
	for _, relay := range allRelayData {
		if relay.State == 0 {
			newActiveRelays = append(newActiveRelays, relayEnabled{relay.Name, relay.ID})
		}
	}

	rec.mux.Lock()
	rec.activeRelays = newActiveRelays
	rec.mux.Unlock()
}

func (rec *RelayEnabledCache) GetEnabledRelays() ([]string, []uint64) {
	if len(rec.activeRelays) == 0 {
		return []string{}, []uint64{}
	}

	relays := make(map[string]uint64)
	relayNames := make([]string, 0)
	rec.mux.RLock()
	activeRelays := rec.activeRelays
	rec.mux.RUnlock()

	//find
	for _, dbRelay := range activeRelays {
		relayNames = append(relayNames, dbRelay.name)
		relays[dbRelay.name] = dbRelay.id
	}

	//sort names and populate ids
	if len(relayNames) > 1 {
		sort.Strings(relayNames)
	}
	relayIDs := make([]uint64, len(relayNames))
	for i, name := range relayNames {
		relayIDs[i] = relays[name]
	}

	return relayNames, relayIDs
}

func (rec *RelayEnabledCache) GetDownRelays(runningRelays []string) ([]string, []uint64) {
	downRelays := make(map[string]uint64)
	downRelayNames := make([]string, 0)
	rec.mux.RLock()
	activeRelays := rec.activeRelays
	rec.mux.RUnlock()

	//find
	for _, dbRelay := range activeRelays {
		found := false
		for _, rName := range runningRelays {
			if dbRelay.name == rName {
				found = true
				break
			}
		}
		if !found {
			downRelayNames = append(downRelayNames, dbRelay.name)
			downRelays[dbRelay.name] = dbRelay.id
		}
	}

	if len(downRelays) == 0 {
		return []string{}, []uint64{}
	}

	//sort names and populate ids
	if len(downRelays) > 1 {
		sort.Strings(downRelayNames)
	}
	downRelayIDs := make([]uint64, len(downRelayNames))
	for i, name := range downRelayNames {
		downRelayIDs[i] = downRelays[name]
	}

	return downRelayNames, downRelayIDs
}
