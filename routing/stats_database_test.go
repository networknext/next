package routing_test

import (
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	t.Run("HistoryMax()", func(t *testing.T) {
		t.Run("returns the max value in the array", func(t *testing.T) {
			history := []float32{1, 2, 3, 4, 5, 4, 3, 2, 1}
			assert.Equal(t, float32(5), routing.HistoryMax(history))
		})
	})

	t.Run("HistoryNotSet()", func(t *testing.T) {
		t.Run("returns a []float32 the size of HistorySize containing nothing but InvalidHistoryValue", func(t *testing.T) {
			history := routing.HistoryNotSet()
			assert.Len(t, history, routing.HistorySize)
			for i := 0; i < routing.HistorySize; i++ {
				assert.Equal(t, routing.HistoryInvalidValue, int(history[i]))
			}
		})
	})

	t.Run("HistoryMean()", func(t *testing.T) {
		t.Run("returns a float32 that is the average of the input", func(t *testing.T) {
			history := []float32{1, 2, 3, 4, 5, 4, 3, 2, 1}
			assert.Equal(t, float32(25.0/9.0), routing.HistoryMean(history))
		})
	})

	t.Run("Old tests", func(t *testing.T) {
		t.Run("TestHistoryMax()", func(t *testing.T) {

			t.Parallel()

			history := routing.HistoryNotSet()

			assert.Equal(t, float32(0.0), routing.HistoryMax(history[:]))

			history[0] = 5.0
			history[1] = 3.0
			history[2] = 100.0

			assert.Equal(t, float32(100.0), routing.HistoryMax(history[:]))
		})

		t.Run("TestHistoryMean", func(t *testing.T) {

			t.Parallel()

			history := routing.HistoryNotSet()

			assert.Equal(t, float32(0.0), routing.HistoryMean(history[:]))

			history[0] = 5.0
			history[1] = 3.0
			history[2] = 100.0

			assert.Equal(t, float32(36.0), routing.HistoryMean(history[:]))
		})
	})
}

func TestStatsDatabase(t *testing.T) {
	sourceID := crypto.HashID("127.0.0.1")
	relay1ID := crypto.HashID("127.9.9.9")
	relay2ID := crypto.HashID("999.999.9.9")

	makeBasicStats := func() *routing.StatsEntryRelay {
		entry := routing.NewStatsEntryRelay()
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
		update := routing.RelayStatsUpdate{
			ID: sourceID,
			PingStats: []routing.RelayStatsPing{
				{
					RelayID:    relay1ID,
					RTT:        0.5,
					Jitter:     0.7,
					PacketLoss: 0.9,
				},
				{
					RelayID:    relay2ID,
					RTT:        0,
					Jitter:     0,
					PacketLoss: 0,
				},
			},
		}

		t.Run("if the entry for the source relay does not exist, it is created properly", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.ProcessStats(&update)
			entry, ok := statsdb.Entries[update.ID]
			assert.True(t, ok)
			assert.NotNil(t, entry.Relays)
		})

		t.Run("if the entry for the destination relay does not exist in the source relay's collection, then it is created properly", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.ProcessStats(&update)
			entry := statsdb.Entries[update.ID]
			for _, r := range entry.Relays {
				assert.NotNil(t, r.RTTHistory)
				assert.NotNil(t, r.JitterHistory)
				assert.NotNil(t, r.PacketLossHistory)
			}
		})

		t.Run("source and destination both don't exist but entries are set to invalid value", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.ProcessStats(&update)
			entry := statsdb.Entries[update.ID]
			assert.Equal(t, len(entry.Relays), len(update.PingStats))
			for _, stats := range update.PingStats {
				destRelay := entry.Relays[stats.RelayID]
				assert.Equal(t, float32(routing.InvalidRouteValue), destRelay.RTT)
				assert.Equal(t, float32(routing.InvalidRouteValue), destRelay.Jitter)
				assert.Equal(t, float32(routing.InvalidRouteValue), destRelay.PacketLoss)
				assert.Equal(t, float32(routing.InvalidRouteValue), destRelay.RTTHistory[0])
				assert.Equal(t, float32(routing.InvalidRouteValue), destRelay.JitterHistory[0])
				assert.Equal(t, float32(routing.InvalidRouteValue), destRelay.PacketLossHistory[0])
			}
		})

		t.Run("source and destination do exist and their entries are updated", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entry := routing.NewStatsEntry()
			statsEntry1 := makeBasicStats()
			entry.Relays[relay1ID] = statsEntry1
			statsdb.Entries[update.ID] = *entry

			statsdb.ProcessStats(&update)

			assert.Equal(t, 2, statsEntry1.Index)
			assert.Equal(t, float32(1.0), statsEntry1.RTT)
			assert.Equal(t, float32(1.0), statsEntry1.Jitter)
			assert.Equal(t, float32(1.0), statsEntry1.PacketLoss)
		})
	})

	t.Run("MakeCopy()", func(t *testing.T) {
		t.Run("makes a copy", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entry := routing.NewStatsEntry()
			statsEntry1 := makeBasicStats()
			entry.Relays[relay1ID] = statsEntry1
			statsdb.Entries[sourceID] = *entry

			cpy := statsdb.MakeCopy()

			assert.Equal(t, statsdb, cpy)
		})
	})

	t.Run("GetEntry()", func(t *testing.T) {
		t.Run("entry does not exist", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()

			stats := statsdb.GetEntry(sourceID, relay1ID)

			assert.Nil(t, stats)
		})

		t.Run("entry exists but stats for the internal entry does not", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entry := routing.NewStatsEntry()
			statsdb.Entries[sourceID] = *entry

			stats := statsdb.GetEntry(sourceID, relay1ID)

			assert.Nil(t, stats)
		})

		t.Run("both the entry and the internal entry exist", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entry := routing.NewStatsEntry()
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

		makeBasicConnection := func(statsdb *routing.StatsDatabase, id1, id2 uint64) {
			statsEntryRelay := routing.NewStatsEntryRelay()
			entry := routing.NewStatsEntry()
			entry.Relays[id2] = statsEntryRelay
			statsdb.Entries[id1] = *entry
		}

		t.Run("relay 1 -> relay 2 does not exist", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.Entries[id1] = *routing.NewStatsEntry()
			makeBasicConnection(statsdb, id2, id1)

			rtt, jitter, packetLoss := statsdb.GetSample(id1, id2)

			assert.Equal(t, float32(routing.InvalidRouteValue), rtt)
			assert.Equal(t, float32(routing.InvalidRouteValue), jitter)
			assert.Equal(t, float32(routing.InvalidRouteValue), packetLoss)
		})

		t.Run("relay 2 -> relay 1 does not exist", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.Entries[id2] = *routing.NewStatsEntry()
			makeBasicConnection(statsdb, id1, id2)

			rtt, jitter, packetLoss := statsdb.GetSample(id1, id2)

			assert.Equal(t, float32(routing.InvalidRouteValue), rtt)
			assert.Equal(t, float32(routing.InvalidRouteValue), jitter)
			assert.Equal(t, float32(routing.InvalidRouteValue), packetLoss)
		})

		t.Run("both relays are valid", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entryForID1 := routing.NewStatsEntry()
			entryForID2 := routing.NewStatsEntry()
			statsForID1 := routing.NewStatsEntryRelay()
			statsForID2 := routing.NewStatsEntryRelay()

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
			statsdb := routing.NewStatsDatabase()
			relayMap := routing.NewRelayMap(func(relayData *routing.RelayData) error {
				statsdb.DeleteEntry(relayData.ID)
				return nil
			})
			// Setup

			fillRelayDatabase(relayMap)
			fillStatsDatabase(statsdb)

			allRelayData := relayMap.GetAllRelayData()

			i := 0
			for _, r := range allRelayData {
				if i == 0 {
					r.Datacenter.ID = 0
					relayMap.AddRelayDataEntry(r.Addr.String(), r)
				} else {
					r.Datacenter.ID = uint64(i)
					relayMap.AddRelayDataEntry(r.Addr.String(), r)
				}
				i++
			}

			modifyEntry := func(addr1, addr2 string, rtt, jitter, packetloss float32) {
				entry := statsdb.GetEntry(crypto.HashID(addr1), crypto.HashID(addr2))
				entry.RTT = rtt
				entry.Jitter = jitter
				entry.PacketLoss = packetloss
			}

			maxJitter := float32(10.0)
			maxPacketLoss := float32(0.1)

			// valid
			modifyEntry("127.0.0.1:40000", "127.0.0.2:40000", 123.00, 1.3, 0.0)

			// invalid - rtt = invalid route value
			modifyEntry("127.0.0.1:40000", "127.0.0.3:40000", routing.InvalidRouteValue, 0.3, 0.0)

			// invalid - jitter > MaxJitter
			modifyEntry("127.0.0.1:40000", "127.0.0.4:40000", 1.0, maxJitter+1, 0.0)

			// invalid - packet loss > MaxPacketLoss
			modifyEntry("127.0.0.1:40000", "127.0.0.5:40000", 1.0, 0.3, maxPacketLoss+1)

			relayIDs := make([]uint64, 0)
			for _, relayData := range allRelayData {
				relayIDs = append(relayIDs, relayData.ID)
			}

			costMatrix := statsdb.GenerateCostMatrix(relayIDs, maxJitter, maxPacketLoss)
			assert.NotEmpty(t, costMatrix)

			// Testing
			// assert that all non-invalid rtt's are within the cost matrix
			getAddressIndex := func(addr1, addr2 string) int {
				addr1ID := crypto.HashID(addr1)
				addr2ID := crypto.HashID(addr2)
				indxOfI := -1
				indxOfJ := -1

				for i, id := range relayIDs {
					if uint64(id) == addr1ID {
						indxOfI = i
					} else if uint64(id) == addr2ID {
						indxOfJ = i
					}

					if indxOfI != -1 && indxOfJ != -1 {
						break
					}
				}

				return routing.TriMatrixIndex(indxOfI, indxOfJ)
			}

			assert.Equal(t, int32(123), costMatrix[getAddressIndex("127.0.0.1:40000", "127.0.0.2:40000")])
			assert.Equal(t, int32(-1), costMatrix[getAddressIndex("127.0.0.1:40000", "127.0.0.3:40000")])
			assert.Equal(t, int32(-1), costMatrix[getAddressIndex("127.0.0.1:40000", "654.0.0.4:40000")])
			assert.Equal(t, int32(-1), costMatrix[getAddressIndex("127.0.0.1:40000", "000.0.0.5:40000")])
		})
	})
}
