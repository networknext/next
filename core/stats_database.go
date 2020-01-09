package core

import (
	"math"
	"sort"
)

const (
	InvalidRouteValue = 10000.0
)

// RelayStatsPing is the ping stats for a relay
type RelayStatsPing struct {
	RelayId    RelayId
	RTT        float32
	Jitter     float32
	PacketLoss float32
}

// RelayStatsUpdate is a struct for updating relay stats
type RelayStatsUpdate struct {
	Id        RelayId
	PingStats []RelayStatsPing
}

// StatsEntryRelay is a entry for relay stats in the stats db
type StatsEntryRelay struct {
	Rtt               float32
	Jitter            float32
	PacketLoss        float32
	Index             int
	RttHistory        [HistorySize]float32
	JitterHistory     [HistorySize]float32
	PacketLossHistory [HistorySize]float32
}

// StatsEntry is a entry in the stats db
type StatsEntry struct {
	Relays map[RelayId]*StatsEntryRelay
}

// StatsDatabase is a relay statistics database
// Each entry contains data about the entry relay to other relays
type StatsDatabase struct {
	Entries map[RelayId]StatsEntry
}

// NewStatsDatabase creates a new stats database
func NewStatsDatabase() *StatsDatabase {
	database := &StatsDatabase{}
	database.Entries = make(map[RelayId]StatsEntry)
	return database
}

// NewStatsEntry creates a new stats entry
func NewStatsEntry() *StatsEntry {
	entry := new(StatsEntry)
	entry.Relays = make(map[RelayId]*StatsEntryRelay)
	return entry
}

// NewStatsEntryRelay creates a new stats entry relay
func NewStatsEntryRelay() *StatsEntryRelay {
	entry := new(StatsEntryRelay)
	entry.RttHistory = HistoryNotSet()
	entry.JitterHistory = HistoryNotSet()
	entry.PacketLossHistory = HistoryNotSet()
	return entry
}

// ProcessStats processes the stats update, creating the needed entries if they do not already exist
func (database *StatsDatabase) ProcessStats(statsUpdate *RelayStatsUpdate) {
	sourceRelayId := statsUpdate.Id

	entry, entryExists := database.Entries[sourceRelayId]
	if !entryExists {
		entry = *NewStatsEntry()
		database.Entries[sourceRelayId] = entry
	}

	for _, stats := range statsUpdate.PingStats {

		destRelayId := stats.RelayId

		relay, relayExists := entry.Relays[destRelayId]

		if !relayExists {
			relay = NewStatsEntryRelay()
		}

		relay.RttHistory[relay.Index] = stats.RTT
		relay.JitterHistory[relay.Index] = stats.Jitter
		relay.PacketLossHistory[relay.Index] = stats.PacketLoss
		relay.Index = (relay.Index + 1) % HistorySize
		relay.Rtt = HistoryMean(relay.RttHistory[:])
		relay.Jitter = HistoryMean(relay.JitterHistory[:])
		relay.PacketLoss = HistoryMean(relay.PacketLossHistory[:])

		entry.Relays[destRelayId] = relay // is this needed? relay is a pointer
	}
}

// MakeCopy makes a exact copy of the stats db
func (database *StatsDatabase) MakeCopy() *StatsDatabase {
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
func (database *StatsDatabase) GetEntry(relay1 RelayId, relay2 RelayId) *StatsEntryRelay {
	if entry, entryExists := database.Entries[relay1]; entryExists {
		if relay, relayExists := entry.Relays[relay2]; relayExists {
			return relay
		}
	}

	return nil
}

// GetSample returns the max values of each stats field of the bidirectional entries in the database
func (database *StatsDatabase) GetSample(relay1 RelayId, relay2 RelayId) (float32, float32, float32) {
	a := database.GetEntry(relay1, relay2)
	b := database.GetEntry(relay2, relay1)
	if a != nil && b != nil {
		// math.Max requires float64 but we're returning float32's hence... whatever this is
		return float32(math.Max(float64(a.Rtt), float64(b.Rtt))),
			float32(math.Max(float64(a.Jitter), float64(b.Jitter))),
			float32(math.Max(float64(a.PacketLoss), float64(b.PacketLoss)))
	}
	return InvalidRouteValue, InvalidRouteValue, InvalidRouteValue
}

// GetCostMatrix TODO
func (database *StatsDatabase) GetCostMatrix(relaydb *RelayDatabase) *CostMatrix {

	numRelays := len(relaydb.Relays)

	entryCount := TriMatrixLength(numRelays)

	costMatrix := &CostMatrix{}
	costMatrix.RelayIds = make([]RelayId, numRelays)
	costMatrix.RelayNames = make([]string, numRelays)
	costMatrix.RelayAddresses = make([][]byte, numRelays)
	costMatrix.RelayPublicKeys = make([][]byte, numRelays)
	costMatrix.DatacenterRelays = make(map[DatacenterId][]RelayId)
	costMatrix.RTT = make([]int32, entryCount)

	datacenterNameMap := make(map[DatacenterId]string)

	var stableRelays []RelayData
	for _, relayData := range relaydb.Relays {
		stableRelays = append(stableRelays, relayData)
	}

	sort.SliceStable(stableRelays, func(i, j int) bool {
		return stableRelays[i].Id < stableRelays[j].Id
	})

	for i, relayData := range stableRelays {
		costMatrix.RelayIds[i] = relayData.Id
		costMatrix.RelayNames[i] = relayData.Name
		costMatrix.RelayPublicKeys[i] = relayData.PublicKey
		if relayData.Datacenter != DatacenterId(0) {
			datacenter := costMatrix.DatacenterRelays[relayData.Datacenter]
			datacenter = append(datacenter, RelayId(relayData.Id))
			costMatrix.DatacenterRelays[relayData.Datacenter] = datacenter
			datacenterNameMap[relayData.Datacenter] = relayData.DatacenterName
		}
	}

	for id, name := range datacenterNameMap {
		costMatrix.DatacenterIds = append(costMatrix.DatacenterIds, id)
		costMatrix.DatacenterNames = append(costMatrix.DatacenterNames, name)
	}

	for i := 0; i < numRelays; i++ {
		for j := 0; j < i; j++ {
			idI := costMatrix.RelayIds[i]
			idJ := costMatrix.RelayIds[j]
			rtt, jitter, packetLoss := database.GetSample(idI, idJ)
			ijIndex := TriMatrixIndex(i, j)
			if rtt != InvalidRouteValue && jitter <= MaxJitter && packetLoss <= MaxPacketLoss {
				costMatrix.RTT[ijIndex] = int32(math.Floor(float64(rtt + jitter)))
			} else {
				costMatrix.RTT[ijIndex] = -1
			}
		}
	}

	return costMatrix
}
