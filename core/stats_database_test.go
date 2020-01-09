package core_test

import (
	"testing"

	"github.com/networknext/backend/core"
	"github.com/stretchr/testify/assert"
)

func TestStatsDatabase(t *testing.T) {
	t.Run("ProcessStats()", func(t *testing.T) {
		relay1ID := core.GetRelayID("127.0.0.1")
		relay2ID := core.GetRelayID("999.999.9.9")
		update := core.RelayStatsUpdate{
			ID: core.GetRelayID("127.0.0.1"),
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
			statsEntry1 := core.NewStatsEntryRelay()
			statsEntry1.Index = 1
			statsEntry1.Rtt = 1
			statsEntry1.Jitter = 1
			statsEntry1.PacketLoss = 1
			statsEntry1.RttHistory[0] = 1
			statsEntry1.JitterHistory[0] = 1
			statsEntry1.PacketLossHistory[0] = 1
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
}
