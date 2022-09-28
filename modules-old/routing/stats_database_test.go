package routing_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/networknext/backend/modules-old/crypto"
	"github.com/networknext/backend/modules-old/routing"

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
			statsdb.Entries[update.ID] = entry

			statsdb.ProcessStats(&update)

			assert.Equal(t, 2, statsEntry1.Index)
			assert.Equal(t, float32(0.75), statsEntry1.RTT)
			assert.Equal(t, float32(0.85), statsEntry1.Jitter)
			assert.Equal(t, float32(0.95), statsEntry1.PacketLoss)
		})
	})

	t.Run("MakeCopy()", func(t *testing.T) {
		t.Run("makes a copy", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entry := routing.NewStatsEntry()
			statsEntry1 := makeBasicStats()
			entry.Relays[relay1ID] = statsEntry1
			statsdb.Entries[sourceID] = entry

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
			statsdb.Entries[sourceID] = entry

			stats := statsdb.GetEntry(sourceID, relay1ID)

			assert.Nil(t, stats)
		})

		t.Run("both the entry and the internal entry exist", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			entry := routing.NewStatsEntry()
			// this entry makes no sense, test puposes only
			statsEntry1 := makeBasicStats()
			entry.Relays[relay1ID] = statsEntry1
			statsdb.Entries[sourceID] = entry

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
			statsdb.Entries[id1] = entry
		}

		t.Run("relay 1 -> relay 2 does not exist", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.Entries[id1] = routing.NewStatsEntry()
			makeBasicConnection(statsdb, id2, id1)

			rtt, jitter, packetLoss := statsdb.GetSample(id1, id2)

			assert.Equal(t, float32(routing.InvalidRouteValue), rtt)
			assert.Equal(t, float32(routing.InvalidRouteValue), jitter)
			assert.Equal(t, float32(routing.InvalidRouteValue), packetLoss)
		})

		t.Run("relay 2 -> relay 1 does not exist", func(t *testing.T) {
			statsdb := routing.NewStatsDatabase()
			statsdb.Entries[id2] = routing.NewStatsEntry()
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

			statsdb.Entries[id1] = entryForID1
			statsdb.Entries[id2] = entryForID2

			rtt, jitter, packetLoss := statsdb.GetSample(id1, id2)

			assert.Equal(t, float32(1000.0), rtt)
			assert.Equal(t, float32(13.0), jitter)
			assert.Equal(t, float32(989.0), packetLoss)
		})
	})
}

func TestDeleteEntry(t *testing.T) {
	db := routing.NewStatsDatabase()
	sourceID := crypto.HashID("127.0.0.1")
	relay1ID := crypto.HashID("127.9.9.9")
	relay2ID := crypto.HashID("999.999.9.9")

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

	t.Run("delete existing entry", func(t *testing.T) {
		db.ProcessStats(&update)

		entry, ok := db.Entries[update.ID]
		assert.True(t, ok)
		assert.NotNil(t, entry.Relays)

		db.DeleteEntry(update.ID)

		entry, ok = db.Entries[update.ID]
		assert.False(t, ok)
		assert.Nil(t, entry)
	})

	t.Run("delete non existing entry", func(t *testing.T) {
		db.DeleteEntry(update.ID)
	})
}

func TestGetCosts(t *testing.T) {
	db := routing.NewStatsDatabase()
	sourceID := crypto.HashID("127.0.0.1")
	relay1ID := crypto.HashID("127.9.9.9")
	relay2ID := crypto.HashID("999.999.9.9")

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

	relayIDs := []uint64{
		sourceID,
		relay1ID,
		relay2ID,
	}

	db.ProcessStats(&update)

	costs := db.GetCosts(relayIDs, 5, 5)

	expectedCosts := []int32{
		-1,
		-1,
		-1,
		1,
	}

	for i := 0; i < len(costs); i++ {
		assert.Equal(t, expectedCosts[i], costs[i])
	}
}

func TestGetCostsLocal(t *testing.T) {
	db := routing.NewStatsDatabase()
	sourceID := crypto.HashID("127.0.0.1")
	relay1ID := crypto.HashID("127.9.9.9")
	relay2ID := crypto.HashID("999.999.9.9")

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

	relayIDs := []uint64{
		sourceID,
		relay1ID,
		relay2ID,
	}

	db.ProcessStats(&update)

	costs := db.GetCostsLocal(relayIDs, 0, 0)

	for i := 0; i < len(costs); i++ {
		assert.Equal(t, int32(0), costs[i])
	}
}

func TestTriMatrixLength(t *testing.T) {
	t.Run("size of 0", func(t *testing.T) {
		length := routing.TriMatrixLength(0)
		assert.Equal(t, 0, length)
	})

	t.Run("size of 1", func(t *testing.T) {
		length := routing.TriMatrixLength(1)
		assert.Equal(t, 0, length)
	})

	t.Run("size of 2", func(t *testing.T) {
		length := routing.TriMatrixLength(2)
		assert.Equal(t, 1, length)
	})

	t.Run("size of 5", func(t *testing.T) {
		length := routing.TriMatrixLength(5)
		assert.Equal(t, 10, length)
	})

	t.Run("size of 10", func(t *testing.T) {
		length := routing.TriMatrixLength(10)
		assert.Equal(t, 45, length)
	})

	t.Run("size of 27", func(t *testing.T) {
		length := routing.TriMatrixLength(27)
		assert.Equal(t, 351, length)
	})

	t.Run("size of 138", func(t *testing.T) {
		length := routing.TriMatrixLength(138)
		assert.Equal(t, 9453, length)
	})

	t.Run("size of 2148", func(t *testing.T) {
		length := routing.TriMatrixLength(2148)
		assert.Equal(t, 2305878, length)
	})
}

func TestTriMatrixIndex(t *testing.T) {
	t.Run("index at 0,0", func(t *testing.T) {
		index := routing.TriMatrixIndex(0, 0)
		assert.Equal(t, 0, index)
	})

	t.Run("index i > j", func(t *testing.T) {
		index := routing.TriMatrixIndex(10, 20)
		assert.Equal(t, 200, index)
	})

	t.Run("index i < j", func(t *testing.T) {
		index := routing.TriMatrixIndex(20, 10)
		assert.Equal(t, 200, index)
	})

	t.Run("index i == j", func(t *testing.T) {
		index := routing.TriMatrixIndex(20, 20)
		assert.Equal(t, 210, index)
	})
}

func TestExtractPingStats(t *testing.T) {
	numRelays := 10

	statsdb := routing.NewStatsDatabase()

	for i := 0; i < numRelays; i++ {
		var update routing.RelayStatsUpdate
		update.ID = uint64(i)
		update.PingStats = make([]routing.RelayStatsPing, numRelays-1)

		for j, idx := 0, 0; j < numRelays; j++ {
			if i == j {
				continue
			}

			update.PingStats[idx].RelayID = uint64(j)
			update.PingStats[idx].RTT = rand.Float32()
			update.PingStats[idx].Jitter = rand.Float32()
			update.PingStats[idx].PacketLoss = rand.Float32()

			idx++
		}

		statsdb.ProcessStats(&update)
	}

	maxJitter := float32(5.0)
	maxPacketLoss := float32(0.1)
	instanceID := "12345"
	isDebug := false

	pairs := statsdb.ExtractPingStats(maxJitter, maxPacketLoss, instanceID, isDebug)
	assert.Len(t, pairs, numRelays*(numRelays-1)/2)

	for i := 0; i < numRelays; i++ {
		for j := 0; j < numRelays; j++ {
			var expectedTimesFound int
			if i == j {
				// this pair should not be in the list
				expectedTimesFound = 0
			} else {
				// this pair should be in the list only once
				expectedTimesFound = 1
			}

			timesFound := 0

			for k := range pairs {
				pair := &pairs[k]
				if (pair.RelayA == uint64(i) && pair.RelayB == uint64(j)) || (pair.RelayA == uint64(j) && pair.RelayB == uint64(i)) {
					timesFound++
				}
			}

			assert.Equal(t, expectedTimesFound, timesFound, fmt.Sprintf("i = %d, j = %d, pairs = %v", i, j, pairs))
		}
	}
}
