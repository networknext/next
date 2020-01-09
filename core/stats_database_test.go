package core_test

import (
	"testing"

	"github.com/networknext/backend/core"
	"github.com/stretchr/testify/assert"
)

func TestStatsDatabase(t *testing.T) {
	sourceID := core.GetRelayID("127.0.0.1")
	relay1ID := core.GetRelayID("127.0.0.1")
	relay2ID := core.GetRelayID("999.999.9.9")

	makeBasicStats := func() *core.StatsEntryRelay {
		entry := core.NewStatsEntryRelay()
		entry.Index = 1
		entry.Rtt = 1
		entry.Jitter = 1
		entry.PacketLoss = 1
		entry.RttHistory[0] = 1
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
				assert.NotNil(t, r.RttHistory)
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
				assert.Equal(t, stats.RTT, destRelay.Rtt)
				assert.Equal(t, stats.Jitter, destRelay.Jitter)
				assert.Equal(t, stats.PacketLoss, destRelay.PacketLoss)
				assert.Equal(t, stats.RTT, destRelay.RttHistory[0])
				assert.Equal(t, stats.Jitter, destRelay.JitterHistory[0])
				assert.Equal(t, stats.PacketLoss, destRelay.PacketLossHistory[0])
			}
		})

		t.Run("source and destination do exist and their entries are updated", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			entry := core.NewStatsEntry()
			// this entry makes no sense, test puposes only
			statsEntry1 := makeBasicStats()
			entry.Relays[relay1ID] = statsEntry1
			statsdb.Entries[update.ID] = *entry

			statsdb.ProcessStats(&update)

			assert.Equal(t, 2, statsEntry1.Index)
			assert.Equal(t, float32(0.75), statsEntry1.Rtt)
			assert.Equal(t, float32(0.85), statsEntry1.Jitter)
			assert.Equal(t, float32(0.95), statsEntry1.PacketLoss)
			// can't assert length of history, well you can but it'll always be the length of HistorySize
		})
	})

	t.Run("MakeCopy()", func(t *testing.T) {
		t.Run("makes a copy", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			entry := core.NewStatsEntry()
			// this entry makes no sense, test puposes only
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

	t.Run("GetCostMatrix()", func(t *testing.T) {
		// TODO taking a moment away from this to work on PR fixes
	})
}
