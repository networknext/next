package core_test

import (
	"testing"

	"github.com/networknext/backend/core"
	"github.com/stretchr/testify/assert"
)

func TestRelayDatabase(t *testing.T) {
	t.Run("updating relay", func(t *testing.T) {
		t.Run("shutdown = true also deletes database entry", func(t *testing.T) {
			relaydb := core.NewRelayDatabase()
			addr := "127.0.0.1"
			id := core.GetRelayID(addr)
			relaydb.Relays[id] = core.RelayData{}

			update := core.RelayUpdate{}
			update.Shutdown = true
			update.Id = id

			_, ok := relaydb.Relays[id]
			assert.True(t, ok)
			assert.False(t, relaydb.UpdateRelay(&update))
			_, ok = relaydb.Relays[id]
			assert.False(t, ok)
		})
	})
}
