package routing

import (
	"math"
	"sync"

	"github.com/networknext/backend/modules/analytics"
)

const (
	HistoryInvalidValue = -1
	HistorySize         = 300 // 5 minutes @ 1 relay update per-second
)

func TriMatrixLength(size int) int {
	return (size * (size - 1)) / 2
}

func TriMatrixIndex(i, j int) int {
	if i <= j {
		i, j = j, i
	}
	return i*(i+1)/2 - i + j
}

func HistoryMax(history []float32) float32 {
	var max float32
	for i := 0; i < len(history); i++ {
		if history[i] > max {
			max = history[i]
		}
	}
	return max
}

func HistoryNotSet() [HistorySize]float32 {
	var res [HistorySize]float32
	for i := 0; i < HistorySize; i++ {
		res[i] = HistoryInvalidValue
	}
	return res
}

const InvalidRouteValue = 10000.0

type RelayStatsPing struct {
	RelayID    uint64  `json:"RelayId"`
	RTT        float32 `json:"RTT"`
	Jitter     float32 `json:"Jitter"`
	PacketLoss float32 `json:"PacketLoss"`
}

type RelayStatsUpdate struct {
	ID        uint64
	PingStats []RelayStatsPing
}

type StatsEntryRelay struct {
	RTT               float32
	Jitter            float32
	PacketLoss        float32
	Index             int
	RTTHistory        [HistorySize]float32
	JitterHistory     [HistorySize]float32
	PacketLossHistory [HistorySize]float32
}

type StatsEntry struct {
	Relays map[uint64]*StatsEntryRelay
}

type StatsDatabase struct {
	Entries map[uint64]*StatsEntry
	mu      sync.Mutex
}

func NewStatsDatabase() *StatsDatabase {
	database := &StatsDatabase{}
	database.Entries = make(map[uint64]*StatsEntry)
	return database
}

func NewStatsEntry() *StatsEntry {
	entry := new(StatsEntry)
	entry.Relays = make(map[uint64]*StatsEntryRelay)
	return entry
}

func NewStatsEntryRelay() *StatsEntryRelay {
	entry := new(StatsEntryRelay)
	entry.RTTHistory = HistoryNotSet()
	entry.JitterHistory = HistoryNotSet()
	entry.PacketLossHistory = HistoryNotSet()
	return entry
}

func (database *StatsDatabase) ExtractPingStats(maxJitter float32, maxPacketLoss float32, instanceID string, isDebug bool) []analytics.PingStatsEntry {
	database.mu.Lock()
	length := TriMatrixLength(len(database.Entries))
	entries := make([]analytics.PingStatsEntry, length)

	ids := make([]uint64, len(database.Entries))

	idx := 0
	for k := range database.Entries {
		ids[idx] = k
		idx++
	}
	database.mu.Unlock()

	if length == 0 {
		return entries
	}

	for i := 1; i < len(ids); i++ {
		for j := 0; j < i; j++ {
			idA := ids[i]
			idB := ids[j]

			rtt, jitter, pl := database.GetSample(idA, idB)
			routable := rtt != InvalidRouteValue && jitter != InvalidRouteValue && pl != InvalidRouteValue

			if jitter > maxJitter {
				routable = false
			}

			if pl > maxPacketLoss {
				routable = false
			}

			entries[TriMatrixIndex(i, j)] = analytics.PingStatsEntry{
				RelayA:     idA,
				RelayB:     idB,
				RTT:        rtt,
				Jitter:     jitter,
				PacketLoss: pl,
				Routable:   routable,
				InstanceID: instanceID,
				Debug:      isDebug,
			}
		}
	}

	return entries
}

// Process ping stats stats coming up from a relay.
// Stats are filtered and we take the worst values across the last 5 minutes
// for latency, packet loss and jitter...

func (database *StatsDatabase) ProcessStats(statsUpdate *RelayStatsUpdate) {

	sourceRelayID := statsUpdate.ID

	if statsUpdate.PingStats == nil {
		return
	}

	database.mu.Lock()
	entry, entryExists := database.Entries[sourceRelayID]
	database.mu.Unlock()

	if !entryExists {
		entry = NewStatsEntry()
		database.mu.Lock()
		database.Entries[sourceRelayID] = entry
		database.mu.Unlock()
	}

	for _, stats := range statsUpdate.PingStats {

		destRelayID := stats.RelayID

		database.mu.Lock()
		relay, relayExists := entry.Relays[destRelayID]
		database.mu.Unlock()

		if !relayExists {

			// new entries are set to invalid route value. it takes 5 minutes past this point before the route becomes valid

			relay = NewStatsEntryRelay()
			relay.RTTHistory[relay.Index] = InvalidRouteValue
			relay.JitterHistory[relay.Index] = InvalidRouteValue
			relay.PacketLossHistory[relay.Index] = InvalidRouteValue

		} else {

			// stash the RTT, jitter and PL into the history buffer

			relay.RTTHistory[relay.Index] = stats.RTT
			relay.JitterHistory[relay.Index] = stats.Jitter
			relay.PacketLossHistory[relay.Index] = stats.PacketLoss

		}

		relay.Index = (relay.Index + 1) % HistorySize

		// By taking the maximum value seen across the last 5 minutes
		// we plan routes very conservatively. It's better for us to never
		// accelerate somebody that we otherwise could, than to accelerate
		// somebody and make their packet loss, latency or jitter worse.

		relay.RTT = HistoryMax(relay.RTTHistory[:])
		relay.Jitter = HistoryMax(relay.JitterHistory[:])
		relay.PacketLoss = HistoryMax(relay.PacketLossHistory[:])

		database.mu.Lock()
		entry.Relays[destRelayID] = relay
		database.mu.Unlock()
	}
}

func (database *StatsDatabase) DeleteEntry(relayID uint64) {
	database.mu.Lock()
	delete(database.Entries, relayID)
	database.mu.Unlock()
}

func (database *StatsDatabase) MakeCopy() *StatsDatabase {
	database.mu.Lock()
	defer database.mu.Unlock()
	databaseCopy := NewStatsDatabase()
	for k, v := range database.Entries {
		newEntry := NewStatsEntry()
		for k2, v2 := range v.Relays {
			vCopy := *v2
			newEntry.Relays[k2] = &vCopy
		}
		databaseCopy.Entries[k] = newEntry
	}
	return databaseCopy
}

func (database *StatsDatabase) GetEntry(relay1, relay2 uint64) *StatsEntryRelay {
	var relay *StatsEntryRelay
	database.mu.Lock()
	entry, entryExists := database.Entries[relay1]
	if entryExists {
		relay = entry.Relays[relay2]
	}
	database.mu.Unlock()
	return relay
}

func (database *StatsDatabase) GetSample(relay1, relay2 uint64) (float32, float32, float32) {
	a := database.GetEntry(relay1, relay2)
	b := database.GetEntry(relay2, relay1)
	if a != nil && b != nil {
		return float32(math.Max(float64(a.RTT), float64(b.RTT))),
			float32(math.Max(float64(a.Jitter), float64(b.Jitter))),
			float32(math.Max(float64(a.PacketLoss), float64(b.PacketLoss)))
	}
	return InvalidRouteValue, InvalidRouteValue, InvalidRouteValue
}

// This function builds the cost matrix from the statistics values in the stats db
// We exclude any routes between relays with jitter or packet loss above the max thresholds.
// Inputs into this function are already filtered stats values across the last 5 minutes,
// so the cost matrix generated is conservative.

func (database *StatsDatabase) GetCosts(relayIDs []uint64, maxJitter float32, maxPacketLoss float32) []int32 {

	numRelays := len(relayIDs)

	costs := make([]int32, TriMatrixLength(numRelays))

	for i := 0; i < numRelays; i++ {

		for j := 0; j < i; j++ {

			ijIndex := TriMatrixIndex(i, j)

			idI := uint64(relayIDs[i])
			idJ := uint64(relayIDs[j])
			rtt, jitter, packetLoss := database.GetSample(idI, idJ)

			if rtt != InvalidRouteValue && jitter <= maxJitter && packetLoss <= maxPacketLoss {
				costs[ijIndex] = int32(math.Ceil(float64(rtt)))
			} else {
				costs[ijIndex] = -1
			}

		}
	}

	return costs
}

// Hack function for getting a highly permissive local cost matrix
// This version just assumes all routes between relays are valid

func (database *StatsDatabase) GetCostsLocal(relayIDs []uint64, maxJitter float32, maxPacketLoss float32) []int32 {

	numRelays := len(relayIDs)

	costs := make([]int32, TriMatrixLength(numRelays))

	for i := 0; i < numRelays; i++ {
		for j := 0; j < i; j++ {
			ijIndex := TriMatrixIndex(i, j)
			costs[ijIndex] = 0
		}
	}

	return costs
}
