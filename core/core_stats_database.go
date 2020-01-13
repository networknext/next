package core

import (
	"math"
	"sort"
)

// HistorySize is the limit to how big the history of the relay entries should be
const HistorySize = 6

// RelayStatsPing is the ping stats for a relay
type RelayStatsPing struct {
	RelayID    uint64
	RTT        float32
	Jitter     float32
	PacketLoss float32
}

// RelayStatsUpdate
type RelayStatsUpdate struct {
	ID        uint64
	PingStats []RelayStatsPing
}

// StatsEntryRelay is an entry for relay stats in the stats db
type StatsEntryRelay struct {
	rtt               float32
	jitter            float32
	packetLoss        float32
	index             int
	rttHistory        [HistorySize]float32
	jitterHistory     [HistorySize]float32
	packetLossHistory [HistorySize]float32
}

// StatsEntry is an entry in the stats db
type StatsEntry struct {
	relays map[uint64]*StatsEntryRelay
}

// StatsDatabase is a relay statistics database (shocking right?)
type StatsDatabase struct {
	entries map[uint64]StatsEntry
}

// NewStatsDatabase creates a new stats database (never would have guessed that)
func NewStatsDatabase() *StatsDatabase {
	database := &StatsDatabase{}
	database.entries = make(map[uint64]StatsEntry)
	return database
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
		res[i] = InvalidHistoryValue
	}
	return res
}

func HistoryMean(history []float32) float32 {
	var sum float32
	var size int
	for i := 0; i < len(history); i++ {
		if history[i] != InvalidHistoryValue {
			sum += history[i]
			size++
		}
	}
	if size == 0 {
		return 0
	}
	return sum / float32(size)
}

func (database *StatsDatabase) ProcessStats(statsUpdate *RelayStatsUpdate) {

	sourceRelay := statsUpdate.ID

	entry, entryExists := database.entries[sourceRelay]
	if !entryExists {
		entry = StatsEntry{
			relays: make(map[uint64]*StatsEntryRelay),
		}
		database.entries[sourceRelay] = entry
	}

	for _, stats := range statsUpdate.PingStats {

		destRelay := stats.RelayID

		relay, relayExists := entry.relays[destRelay]

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

		entry.relays[destRelay] = relay
	}
}

func (database *StatsDatabase) MakeCopy() *StatsDatabase {
	database_copy := NewStatsDatabase()
	for k, v := range database.entries {
		newEntry := StatsEntry{
			relays: make(map[uint64]*StatsEntryRelay),
		}
		for k2, v2 := range v.relays {
			v_copy := *v2
			newEntry.relays[k2] = &v_copy
		}
		database_copy.entries[k] = newEntry
	}
	return database_copy
}

func (database *StatsDatabase) GetEntry(relay1 uint64, relay2 uint64) *StatsEntryRelay {
	entry, entryExists := database.entries[relay1]
	if entryExists {
		relay, relayExists := entry.relays[relay2]
		if relayExists {
			return relay
		}
	}
	return nil
}

func max(x float32, y float32) float32 {
	if x > y {
		return x
	} else {
		return y
	}
}

func (database *StatsDatabase) GetSample(relays *RelayDatabase, relay1 uint64, relay2 uint64) (float32, float32, float32) {
	a := database.GetEntry(relay1, relay2)
	b := database.GetEntry(relay2, relay1)
	if a != nil && b != nil {
		return max(a.rtt, b.rtt), max(a.jitter, b.jitter), max(a.packetLoss, b.packetLoss)
	} else {
		return InvalidRouteValue, InvalidRouteValue, InvalidRouteValue
	}
}

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
		costMatrix.RelayIds[i] = RelayId(relayData.ID)
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
			id_i := uint64(costMatrix.RelayIds[i])
			id_j := uint64(costMatrix.RelayIds[j])
			rtt, jitter, packetLoss := database.GetSample(relays, id_i, id_j)
			ij_index := TriMatrixIndex(i, j)
			if rtt != InvalidRouteValue && jitter <= MaxJitter && packetLoss <= MaxPacketLoss {
				costMatrix.RTT[ij_index] = int32(math.Floor(float64(rtt + jitter)))
			} else {
				costMatrix.RTT[ij_index] = -1
			}
		}
	}

	return costMatrix
}
