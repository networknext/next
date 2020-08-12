package routing

import (
	"math"
	"sort"
	"sync"
)

// HistorySize is the limit to how big the history of the relay entries should be
const (
	HistoryInvalidValue = -1
	HistorySize         = 300 // 5 minutes @ 1 relay update per-second
)

// TriMatrixLength returns the length of a triangular shaped matrix
func TriMatrixLength(size int) int {
	return (size * (size - 1)) / 2
}

// TriMatrixIndex returns the index of the ij coord for a triangular shaped matrix
func TriMatrixIndex(i, j int) int {
	if i <= j {
		i, j = j, i
	}
	return i*(i+1)/2 - i + j
}

// HistoryMax returns the max value in the history array
func HistoryMax(history []float32) float32 {
	var max float32
	for i := 0; i < len(history); i++ {
		if history[i] > max {
			max = history[i]
		}
	}
	return max
}

// HistoryNotSet returns a history array initialized with invalid history values
func HistoryNotSet() [HistorySize]float32 {
	var res [HistorySize]float32
	for i := 0; i < HistorySize; i++ {
		res[i] = HistoryInvalidValue
	}
	return res
}

// HistoryMean returns the average value of all the history entries
func HistoryMean(history []float32) float32 {
	var sum float32
	var size int
	for i := 0; i < len(history); i++ {
		if history[i] != HistoryInvalidValue {
			sum += history[i]
			size++
		}
	}
	if size == 0 {
		return 0
	}
	return sum / float32(size)
}

// InvalidRouteValue ...
const InvalidRouteValue = 10000.0

// RelayStatsPing is the ping stats for a relay
type RelayStatsPing struct {
	RelayID    uint64  `json:"RelayId"`
	RTT        float32 `json:"RTT"`
	Jitter     float32 `json:"Jitter"`
	PacketLoss float32 `json:"PacketLoss"`
}

// RelayStatsUpdate is a struct for updating relay stats
type RelayStatsUpdate struct {
	ID        uint64
	PingStats []RelayStatsPing
}

// StatsEntryRelay is an entry for relay stats in the stats db
type StatsEntryRelay struct {
	RTT               float32
	Jitter            float32
	PacketLoss        float32
	Index             int
	RTTHistory        [HistorySize]float32
	JitterHistory     [HistorySize]float32
	PacketLossHistory [HistorySize]float32
}

// StatsEntry is an entry in the stats db
type StatsEntry struct {
	Relays map[uint64]*StatsEntryRelay
}

// StatsDatabase is a relay statistics database.
// Each entry contains data about the entry relay to other relays
type StatsDatabase struct {
	Entries map[uint64]StatsEntry

	mu sync.Mutex
}

// NewStatsDatabase creates a new stats database
func NewStatsDatabase() *StatsDatabase {
	database := &StatsDatabase{}
	database.Entries = make(map[uint64]StatsEntry)
	return database
}

// NewStatsEntry creates a new stats entry
func NewStatsEntry() *StatsEntry {
	entry := new(StatsEntry)
	entry.Relays = make(map[uint64]*StatsEntryRelay)
	return entry
}

// NewStatsEntryRelay creates a new stats entry relay
func NewStatsEntryRelay() *StatsEntryRelay {
	entry := new(StatsEntryRelay)
	entry.RTTHistory = HistoryNotSet()
	entry.JitterHistory = HistoryNotSet()
	entry.PacketLossHistory = HistoryNotSet()
	return entry
}

// ProcessStats processes the stats update, creating the needed entries if they do not already exist
func (database *StatsDatabase) ProcessStats(statsUpdate *RelayStatsUpdate) {
	sourceRelayID := statsUpdate.ID

	if statsUpdate.PingStats == nil {
		return
	}

	database.mu.Lock()
	entry, entryExists := database.Entries[sourceRelayID]
	database.mu.Unlock()

	if !entryExists {
		entry = *NewStatsEntry()
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
			relay = NewStatsEntryRelay()

			relay.RTTHistory[relay.Index] = InvalidRouteValue
			relay.JitterHistory[relay.Index] = InvalidRouteValue
			relay.PacketLossHistory[relay.Index] = InvalidRouteValue

		} else {
			// Make sure that relays with 100% packet loss do not have 0 RTT
			// and will definitely be excluded during route optimzation
			if stats.PacketLoss > 99 {
				relay.RTTHistory[relay.Index] = InvalidRouteValue
				relay.JitterHistory[relay.Index] = InvalidRouteValue
				relay.PacketLossHistory[relay.Index] = InvalidRouteValue
			} else {
				relay.RTTHistory[relay.Index] = stats.RTT
				relay.JitterHistory[relay.Index] = stats.Jitter
				relay.PacketLossHistory[relay.Index] = stats.PacketLoss
			}
		}

		relay.Index = (relay.Index + 1) % HistorySize
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

// MakeCopy makes a exact copy of the stats db
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
		databaseCopy.Entries[k] = *newEntry
	}
	return databaseCopy
}

// GetEntry retrieves the stats for the supplied relay id's, if either or both do not exist the function returns nil
func (database *StatsDatabase) GetEntry(relay1, relay2 uint64) *StatsEntryRelay {
	database.mu.Lock()
	entry, entryExists := database.Entries[relay1]
	relay, relayExists := entry.Relays[relay2]
	database.mu.Unlock()

	if entryExists && relayExists {
		return relay
	}

	return nil
}

// GetSample returns the max values of each stats field of the bidirectional entries in the database
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

// GetCostMatrix returns the cost matrix composed of all current information
func (database *StatsDatabase) GetCostMatrix(
	costMatrix *CostMatrix,
	allRelayData []*RelayData,
	maxJitter float32,
	maxPacketLoss float32) error {

	numRelays := len(allRelayData)

	var stableRelays []*RelayData
	for _, relay := range allRelayData {
		stableRelays = append(stableRelays, relay)
	}

	sort.SliceStable(stableRelays, func(i, j int) bool {
		return stableRelays[i].ID < stableRelays[j].ID
	})

	costMatrix.RelayIndices = make(map[uint64]int)
	costMatrix.RelayIDs = make([]uint64, numRelays)
	costMatrix.RelayNames = make([]string, numRelays)
	costMatrix.RelayAddresses = make([][]byte, numRelays)
	costMatrix.RelayLatitude = make([]float64, numRelays)
	costMatrix.RelayLongitude = make([]float64, numRelays)
	costMatrix.RelayPublicKeys = make([][]byte, numRelays)
	// DatacenterIDs is handled below
	// DatacenterNames is handled below
	costMatrix.DatacenterRelays = make(map[uint64][]uint64)
	costMatrix.RTT = make([]int32, TriMatrixLength(numRelays))
	costMatrix.RelaySellers = make([]Seller, numRelays)
	costMatrix.RelaySessionCounts = make([]uint32, numRelays)
	costMatrix.RelayMaxSessionCounts = make([]uint32, numRelays)

	datacenterNameMap := make(map[uint64]string)

	for i, relayData := range stableRelays {
		costMatrix.RelayIndices[relayData.ID] = i
		costMatrix.RelayIDs[i] = relayData.ID
		costMatrix.RelayNames[i] = relayData.Name
		costMatrix.RelaySellers[i] = relayData.Seller
		costMatrix.RelaySessionCounts[i] = uint32(relayData.TrafficStats.SessionCount)
		costMatrix.RelayMaxSessionCounts[i] = relayData.MaxSessions

		costMatrix.RelayAddresses[i] = make([]byte, MaxRelayAddressLength)
		copy(costMatrix.RelayAddresses[i], []byte(relayData.Addr.String()))

		costMatrix.RelayPublicKeys[i] = relayData.PublicKey
		if relayData.Datacenter.ID != 0 {
			datacenter := costMatrix.DatacenterRelays[relayData.Datacenter.ID]
			datacenter = append(datacenter, relayData.ID)
			costMatrix.DatacenterRelays[relayData.Datacenter.ID] = datacenter
			datacenterNameMap[relayData.Datacenter.ID] = relayData.Datacenter.Name
		}
	}

	costMatrix.DatacenterIDs = make([]uint64, len(datacenterNameMap))
	costMatrix.DatacenterNames = make([]string, len(datacenterNameMap))
	idx := 0
	for id, name := range datacenterNameMap {
		costMatrix.DatacenterIDs[idx] = id
		costMatrix.DatacenterNames[idx] = name
		idx++
	}

	for i := 0; i < numRelays; i++ {
		for j := 0; j < i; j++ {
			idI := uint64(costMatrix.RelayIDs[i])
			idJ := uint64(costMatrix.RelayIDs[j])
			rtt, jitter, packetLoss := database.GetSample(idI, idJ)
			ijIndex := TriMatrixIndex(i, j)
			if rtt != InvalidRouteValue && jitter <= maxJitter && packetLoss <= maxPacketLoss {
				costMatrix.RTT[ijIndex] = int32(math.Floor(float64(rtt) + float64(jitter)))
			} else {
				costMatrix.RTT[ijIndex] = -1
			}
		}
	}

	return nil
}
