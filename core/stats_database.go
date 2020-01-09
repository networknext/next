package core

import (
	"math"
	"sort"
)

// RelayStatsPing is the ping stats for a relay
type RelayStatsPing struct {
	RelayID    RelayId
	RTT        float32
	Jitter     float32
	PacketLoss float32
}

// RelayStatsUpdate is a struct for updating relay stats
type RelayStatsUpdate struct {
	ID        RelayId
	PingStats []RelayStatsPing
}

// StatsEntryRelay is a entry for relay stats in the stats db
type StatsEntryRelay struct {
	rtt               float32
	jitter            float32
	packetLoss        float32
	index             int
	rttHistory        [HistorySize]float32
	jitterHistory     [HistorySize]float32
	packetLossHistory [HistorySize]float32
}

// StatsEntry is a entry in the stats db
type StatsEntry struct {
	Relays map[RelayId]*StatsEntryRelay
}

// StatsDatabase is a relay statistics database (shocking right?)
type StatsDatabase struct {
	Entries map[RelayId]StatsEntry
}

// NewStatsDatabase creates a new stats database (never would have guessed that)
func NewStatsDatabase() *StatsDatabase {
	database := &StatsDatabase{}
	database.Entries = make(map[RelayId]StatsEntry)
	return database
}

// ProcessStats TODO
func (database *StatsDatabase) ProcessStats(statsUpdate *RelayStatsUpdate) {
	sourceRelay := statsUpdate.ID

	entry, entryExists := database.Entries[sourceRelay]
	if !entryExists {
		entry = StatsEntry{
			Relays: make(map[RelayId]*StatsEntryRelay),
		}
		database.Entries[sourceRelay] = entry
	}

	for _, stats := range statsUpdate.PingStats {

		destRelay := stats.RelayID

		relay, relayExists := entry.Relays[destRelay]

		if !relayExists {
			relay = &StatsEntryRelay{
				rttHistory:        HistoryNotSet(),
				jitterHistory:     HistoryNotSet(),
				packetLossHistory: HistoryNotSet(),
			}
		}

		relay.rttHistory[relay.index] = stats.RTT
		relay.jitterHistory[relay.index] = stats.Jitter
		relay.packetLossHistory[relay.index] = stats.PacketLoss
		relay.index = (relay.index + 1) % HistorySize
		relay.rtt = HistoryMean(relay.rttHistory[:])
		relay.jitter = HistoryMean(relay.jitterHistory[:])
		relay.packetLoss = HistoryMean(relay.packetLossHistory[:])

		entry.Relays[destRelay] = relay
	}
}

// MakeCopy TODO
func (database *StatsDatabase) MakeCopy() *StatsDatabase {
	databaseCopy := NewStatsDatabase()
	for k, v := range database.Entries {
		newEntry := StatsEntry{
			Relays: make(map[RelayId]*StatsEntryRelay),
		}
		for k2, v2 := range v.Relays {
			vCopy := *v2
			newEntry.Relays[k2] = &vCopy
		}
		databaseCopy.Entries[k] = newEntry
	}
	return databaseCopy
}

// GetEntry TODO
func (database *StatsDatabase) GetEntry(relay1 RelayId, relay2 RelayId) *StatsEntryRelay {
	entry, entryExists := database.Entries[relay1]
	if entryExists {
		relay, relayExists := entry.Relays[relay2]
		if relayExists {
			return relay
		}
	}
	return nil
}

// GetSample TODO
func (database *StatsDatabase) GetSample(relays *RelayDatabase, relay1 RelayId, relay2 RelayId) (float32, float32, float32) {
	a := database.GetEntry(relay1, relay2)
	b := database.GetEntry(relay2, relay1)
	if a != nil && b != nil {
		return float32(math.Max(float64(a.rtt), float64(b.rtt))),
			float32(math.Max(float64(a.jitter), float64(b.jitter))),
			float32(math.Max(float64(a.packetLoss), float64(b.packetLoss)))
	}
	return InvalidRouteValue, InvalidRouteValue, InvalidRouteValue
}

// GetCostMatrix TODO
func (database *StatsDatabase) GetCostMatrix(relays *RelayDatabase) *CostMatrix {

	numRelays := len(relays.Relays)

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
	for _, relayData := range relays.Relays {
		stableRelays = append(stableRelays, relayData)
	}

	sort.SliceStable(stableRelays, func(i, j int) bool {
		return stableRelays[i].ID < stableRelays[j].ID
	})

	for i, relayData := range stableRelays {
		costMatrix.RelayIds[i] = relayData.ID
		costMatrix.RelayNames[i] = relayData.Name
		costMatrix.RelayPublicKeys[i] = relayData.PublicKey
		if relayData.Datacenter != DatacenterId(0) {
			datacenter := costMatrix.DatacenterRelays[relayData.Datacenter]
			datacenter = append(datacenter, RelayId(relayData.ID))
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
			rtt, jitter, packetLoss := database.GetSample(relays, idI, idJ)
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
