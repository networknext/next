package core_test

import (
	"testing"

	"github.com/networknext/backend/core"
	"github.com/stretchr/testify/assert"
)

func TestStatsDatabase(t *testing.T) {
	sourceID := core.GetRelayID("127.0.0.1")
	relay1ID := core.GetRelayID("127.9.9.9")
	relay2ID := core.GetRelayID("999.999.9.9")

	makeBasicStats := func() *core.StatsEntryRelay {
		entry := core.NewStatsEntryRelay()
		entry.Index = 1
		entry.RTT = 1
		entry.Jitter = 1
		entry.PacketLoss = 1
		entry.RTTHistory[0] = 1
		entry.JitterHistory[0] = 1
		entry.PacketLossHistory[0] = 1
		return entry
	}

	t.Run("ProcessStats()", func(t *testing.T) {
		update := core.RelayStatsUpdate{
			ID: sourceID,
			PingStats: []core.RelayStatsPing{
				core.RelayStatsPing{
					RelayID:    relay1ID,
					RTT:        0.5,
					Jitter:     0.7,
					PacketLoss: 0.9,
				},
				core.RelayStatsPing{
					RelayID:    relay2ID,
					RTT:        0,
					Jitter:     0,
					PacketLoss: 0,
				},
			},
		}

		t.Run("if the entry for the source relay does not exist, it is created properly", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			statsdb.ProcessStats(&update)
			entry, ok := statsdb.Entries[update.ID]
			assert.True(t, ok)
			assert.NotNil(t, entry.Relays)
		})

		t.Run("if the entry for the destination relay does not exist in the source relay's collection, then it is created properly", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			statsdb.ProcessStats(&update)
			entry, _ := statsdb.Entries[update.ID]
			for _, r := range entry.Relays {
				assert.NotNil(t, r.RTTHistory)
				assert.NotNil(t, r.JitterHistory)
				assert.NotNil(t, r.PacketLossHistory)
			}
		})

		t.Run("source and destination both don't exist but entries are updated properly", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			statsdb.ProcessStats(&update)
			entry, _ := statsdb.Entries[update.ID]
			assert.Equal(t, len(entry.Relays), len(update.PingStats))
			for _, stats := range update.PingStats {
				destRelay, _ := entry.Relays[stats.RelayID]
				assert.Equal(t, stats.RTT, destRelay.RTT)
				assert.Equal(t, stats.Jitter, destRelay.Jitter)
				assert.Equal(t, stats.PacketLoss, destRelay.PacketLoss)
				assert.Equal(t, stats.RTT, destRelay.RTTHistory[0])
				assert.Equal(t, stats.Jitter, destRelay.JitterHistory[0])
				assert.Equal(t, stats.PacketLoss, destRelay.PacketLossHistory[0])
			}
		})

		t.Run("source and destination do exist and their entries are updated", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			entry := core.NewStatsEntry()
			statsEntry1 := makeBasicStats()
			entry.Relays[relay1ID] = statsEntry1
			statsdb.Entries[update.ID] = *entry

			statsdb.ProcessStats(&update)

			assert.Equal(t, 2, statsEntry1.Index)
			assert.Equal(t, float32(0.75), statsEntry1.RTT)
			assert.Equal(t, float32(0.85), statsEntry1.Jitter)
			assert.Equal(t, float32(0.95), statsEntry1.PacketLoss)
		})
	})

	t.Run("MakeCopy()", func(t *testing.T) {
		t.Run("makes a copy", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			entry := core.NewStatsEntry()
			statsEntry1 := makeBasicStats()
			entry.Relays[relay1ID] = statsEntry1
			statsdb.Entries[sourceID] = *entry

			cpy := statsdb.MakeCopy()

			assert.Equal(t, statsdb, cpy)
		})
	})

	t.Run("GetEntry()", func(t *testing.T) {
		t.Run("entry does not exist", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()

			stats := statsdb.GetEntry(sourceID, relay1ID)

			assert.Nil(t, stats)
		})

		t.Run("entry exists but stats for the internal entry does not", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			entry := core.NewStatsEntry()
			statsdb.Entries[sourceID] = *entry

			stats := statsdb.GetEntry(sourceID, relay1ID)

			assert.Nil(t, stats)
		})

		t.Run("both the entry and the internal entry exist", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			entry := core.NewStatsEntry()
			// this entry makes no sense, test puposes only
			statsEntry1 := makeBasicStats()
			entry.Relays[relay1ID] = statsEntry1
			statsdb.Entries[sourceID] = *entry

			stats := statsdb.GetEntry(sourceID, relay1ID)

			assert.Equal(t, statsEntry1, stats)
		})
	})

	t.Run("GetSample()", func(t *testing.T) {
		id1 := sourceID
		id2 := relay1ID

		makeBasicConnection := func(statsdb *core.StatsDatabase, id1, id2 uint64) {
			stats := core.NewStatsEntryRelay()
			entry := core.NewStatsEntry()
			entry.Relays[id2] = stats
			statsdb.Entries[id1] = *entry
		}

		t.Run("relay 1 -> relay 2 does not exist", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			statsdb.Entries[id1] = *core.NewStatsEntry()
			makeBasicConnection(statsdb, id2, id1)

			rtt, jitter, packetLoss := statsdb.GetSample(id1, id2)

			assert.Equal(t, float32(core.InvalidRouteValue), rtt)
			assert.Equal(t, float32(core.InvalidRouteValue), jitter)
			assert.Equal(t, float32(core.InvalidRouteValue), packetLoss)
		})

		t.Run("relay 2 -> relay 1 does not exist", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			statsdb.Entries[id2] = *core.NewStatsEntry()
			makeBasicConnection(statsdb, id1, id2)

			rtt, jitter, packetLoss := statsdb.GetSample(id1, id2)

			assert.Equal(t, float32(core.InvalidRouteValue), rtt)
			assert.Equal(t, float32(core.InvalidRouteValue), jitter)
			assert.Equal(t, float32(core.InvalidRouteValue), packetLoss)
		})

		t.Run("both relays are valid", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			entryForID1 := core.NewStatsEntry()
			entryForID2 := core.NewStatsEntry()
			statsForID1 := core.NewStatsEntryRelay()
			statsForID2 := core.NewStatsEntryRelay()

			statsForID1.RTT = 1000.0
			statsForID1.Jitter = 12.345
			statsForID1.PacketLoss = 987.654

			statsForID2.RTT = 999.99
			statsForID2.Jitter = 13.0
			statsForID2.PacketLoss = 989.0

			entryForID1.Relays[id2] = statsForID2
			entryForID2.Relays[id1] = statsForID1

			statsdb.Entries[id1] = *entryForID1
			statsdb.Entries[id2] = *entryForID2

			rtt, jitter, packetLoss := statsdb.GetSample(id1, id2)

			assert.Equal(t, float32(1000.0), rtt)
			assert.Equal(t, float32(13.0), jitter)
			assert.Equal(t, float32(989.0), packetLoss)
		})
	})

	t.Run("GetCostMatrix()", func(t *testing.T) {
		t.Run("returns the cost matrix", func(t *testing.T) {
			relaydb := core.NewRelayDatabase()
			statsdb := core.NewStatsDatabase()

			// Setup

			FillRelayDatabase(relaydb)
			FillStatsDatabase(statsdb)

			// make the datacenter of the first relay 0
			// otherwise push the rest into the validDcIDs array
			// same with Dc names
			validDcIDs := make([]core.DatacenterID, 0)
			validDcNames := make([]string, 0)

			i := 0
			for _, r := range relaydb.Relays {
				if i == 0 {
					r.Datacenter = core.DatacenterID(0)
					relaydb.Relays[core.GetRelayID(r.Address)] = r
				} else {
					r.Datacenter = core.DatacenterID(i)
					relaydb.Relays[core.GetRelayID(r.Address)] = r
					validDcIDs = append(validDcIDs, r.Datacenter)
					validDcNames = append(validDcNames, r.DatacenterName)
				}
				i++
			}

			modifyEntry := func(addr1, addr2 string, rtt, jitter, packetloss float32) {
				entry := statsdb.GetEntry(core.GetRelayID(addr1), core.GetRelayID(addr2))
				entry.RTT = rtt
				entry.Jitter = jitter
				entry.PacketLoss = packetloss
			}

			// valid
			modifyEntry("127.0.0.1", "127.0.0.2", 123.00, 1.3, 0.0)

			// invalid - rtt = invalid route value
			modifyEntry("127.0.0.1", "123.4.5.6", core.InvalidRouteValue, 0.3, 0.0)

			// invalid - jitter > MaxJitter
			modifyEntry("127.0.0.1", "654.3.2.1", 1.0, core.MaxJitter+1, 0.0)

			// invalid - packet loss > MaxPacketLoss
			modifyEntry("127.0.0.1", "000.0.0.0", 1.0, 0.3, core.MaxPacketLoss+1)

			costMatrix := statsdb.GetCostMatrix(relaydb)

			// Testing

			// assert each entry in the relay db is present in the cost matrix
			for _, relay := range relaydb.Relays {
				assert.Contains(t, costMatrix.RelayIDs, core.RelayID(relay.ID))
				assert.Contains(t, costMatrix.RelayNames, relay.Name)
				assert.Contains(t, costMatrix.RelayPublicKeys, relay.PublicKey)
			}

			// assert the length of the valid ids equals the length of all the datacenter ids in the matrix
			assert.Equal(t, len(validDcIDs), len(costMatrix.DatacenterIDs))
			assert.Equal(t, len(validDcNames), len(costMatrix.DatacenterNames))
			for i, id := range validDcIDs {
				// assert all valid ids are present in the matrix
				assert.Contains(t, costMatrix.DatacenterIDs, id)

				// find the relays whose datacenter id matches this one
				validRelayIDs := make([]uint64, 0)
				for _, relay := range relaydb.Relays {
					if relay.Datacenter == id {
						validRelayIDs = append(validRelayIDs, relay.ID)
						break
					}
				}

				// assert the datacenter id -> relay ids mapping contains the actual ids
				// i + 1 because in the first for loop each datacenter id is reset as i which is > 0
				for _, relayID := range validRelayIDs {
					assert.Contains(t, costMatrix.DatacenterRelays[core.DatacenterID(i+1)], core.RelayID(relayID))
				}
			}

			// assert all names valid Dc names are within the cost matrix
			for _, name := range validDcNames {
				assert.Contains(t, costMatrix.DatacenterNames, name)
			}

			// assert that all non-invalid rtt's are within the cost matrix
			getAddressIndex := func(addr1, addr2 string) int {
				addr1ID := core.GetRelayID(addr1)
				addr2ID := core.GetRelayID(addr2)
				indxOfI := -1
				indxOfJ := -1

				for i, id := range costMatrix.RelayIDs {
					if uint64(id) == addr1ID {
						indxOfI = i
					} else if uint64(id) == addr2ID {
						indxOfJ = i
					}

					if indxOfI != -1 && indxOfJ != -1 {
						break
					}
				}

				return core.TriMatrixIndex(indxOfI, indxOfJ)
			}

			assert.Equal(t, int32(124), costMatrix.RTT[getAddressIndex("127.0.0.1", "127.0.0.2")])
			assert.Equal(t, int32(-1), costMatrix.RTT[getAddressIndex("127.0.0.1", "123.4.5.6")])
			assert.Equal(t, int32(-1), costMatrix.RTT[getAddressIndex("127.0.0.1", "654.3.2.1")])
			assert.Equal(t, int32(-1), costMatrix.RTT[getAddressIndex("127.0.0.1", "000.0.0.0")])
		})
	})
}
