package core_test

import (
	"hash/fnv"
	"testing"

	"github.com/networknext/backend/core"
	"github.com/stretchr/testify/assert"
)

func TestRelayDatabase(t *testing.T) {
	t.Run("UpdateRelay()", func(t *testing.T) {
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

		t.Run("relay did not already exist", func(t *testing.T) {
			t.Skip()
		})

		t.Run("relay did already exist", func(t *testing.T) {
			t.Skip()
		})
	})

	t.Run("CheckForTimeouts()", func(t *testing.T) {
		t.Run("dead relays are present", func(t *testing.T) {
			t.Skip()
		})

		t.Run("all relays are alive", func(t *testing.T) {
			t.Skip()
		})
	})

	t.Run("MakeCopy()", func(t *testing.T) {
		t.Run("returns an exact copy", func(t *testing.T) {
			t.Skip()
		})
	})
}

func TestGetRelayID(t *testing.T) {
	t.Run("returns the hash of the supplied value", func(t *testing.T) {
		duplicateFunction := func(value string) core.RelayId {
			hash := fnv.New64a()
			hash.Write([]byte(value))
			return core.RelayId(hash.Sum64())
		}

		value := "127.0.0.1"
		assert.Equal(t, duplicateFunction(value), core.GetRelayID(value))
	})
}
