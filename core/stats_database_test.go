package core_test

import (
	"testing"

	"github.com/networknext/backend/core"
	"github.com/stretchr/testify/assert"
)

func TestStatsDatabase(t *testing.T) {
	t.Run("ProcessStats()", func(t *testing.T) {
		update := core.RelayStatsUpdate{
			ID: core.GetRelayID("127.0.0.1"),
			PingStats: []core.RelayStatsPing{
				core.RelayStatsPing{
					RelayID:    core.GetRelayID("111.111.1.1"),
					RTT:        1.123,
					Jitter:     0.2,
					PacketLoss: 0.5,
				},
				core.RelayStatsPing{
					RelayID:    core.GetRelayID("999.999.9.9"),
					RTT:        0,
					Jitter:     0,
					PacketLoss: 0,
				},
			},
		}

		t.Run("if the stats entry does not exist, it is created properly", func(t *testing.T) {
			statsdb := core.NewStatsDatabase()
			statsdb.ProcessStats(&update)
			entry, ok := statsdb.Entries[update.ID]
			assert.True(t, ok)
			assert.NotNil(t, entry.Relays)
		})
	})
}
