package routing_test

import (
	"math"
	"math/rand"
	"net"
	"testing"

	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestRelayDatabase(t *testing.T) {
	t.Run("UpdateRelay()", func(t *testing.T) {
		t.Run("shutdown = true also deletes database entry", func(t *testing.T) {
			relaydb := routing.NewRelayDatabase()
			addr := "127.0.0.1"
			id := routing.GetRelayID(addr)
			relaydb.Relays[id] = routing.Relay{}

			update := routing.RelayUpdate{}
			update.Shutdown = true
			update.ID = id

			_, ok := relaydb.Relays[id]
			assert.True(t, ok)
			assert.False(t, relaydb.UpdateRelay(&update))
			_, ok = relaydb.Relays[id]
			assert.False(t, ok)
		})

		t.Run("relay did not already exist", func(t *testing.T) {
			relaydb := routing.NewRelayDatabase()
			update := routing.RelayUpdate{}
			assert.True(t, relaydb.UpdateRelay(&update))
		})

		t.Run("relay did already exist", func(t *testing.T) {
			relaydb := routing.NewRelayDatabase()
			addr := "127.0.0.1"
			id := routing.GetRelayID(addr)
			relaydb.Relays[id] = routing.Relay{}

			update := routing.RelayUpdate{}
			update.ID = id

			assert.False(t, relaydb.UpdateRelay(&update))
		})

		t.Run("updates correctly", func(t *testing.T) {
			relaydb := routing.NewRelayDatabase()
			addr := "127.0.0.1:40000"
			id := routing.GetRelayID(addr)
			udp, _ := net.ResolveUDPAddr("udp", addr)
			update := routing.RelayUpdate{
				ID:             id,
				Name:           "Some Name",
				Address:        *udp,
				Datacenter:     uint64(rand.Int()%math.MaxInt32 + 1),
				DatacenterName: "Some Datacenter",
				PublicKey:      RandomPublicKey(),
				Shutdown:       false,
			}

			assert.True(t, relaydb.UpdateRelay(&update))
			value, ok := relaydb.Relays[id]
			assert.True(t, ok)

			// is there a go equivalent for c++ operator== overloading? or Java's .equal() method? Googling did me no help
			assert.Equal(t, update.ID, value.ID)
			assert.Equal(t, update.Name, value.Name)
			assert.Equal(t, update.Address, value.Addr)
			assert.Equal(t, update.Datacenter, value.Datacenter)
			assert.Equal(t, update.DatacenterName, value.DatacenterName)
			assert.Equal(t, update.PublicKey, value.PublicKey)
		})
	})

	/* MAY OCCASIONALLY FAIL DUE TO TIMING if so rerun and pray */
	t.Run("CheckForTimeouts()", func(t *testing.T) {
		t.Skip("indeterminate tests")

		t.Run("dead relays are present", func(t *testing.T) {
			relaydb := routing.NewRelayDatabase()
			FillRelayDatabase(relaydb)
			expectedDeadRelays := []uint64{routing.GetRelayID("654.3.2.1"), routing.GetRelayID("999.9.9.9")}

			deadRelays := relaydb.CheckForTimeouts(50)
			assert.Equal(t, expectedDeadRelays, deadRelays)
			for _, id := range expectedDeadRelays {
				_, ok := relaydb.Relays[id]
				assert.False(t, ok, "ID: %x", id)
			}
		})

		t.Run("all relays are alive", func(t *testing.T) {
			relaydb := routing.NewRelayDatabase()
			FillRelayDatabase(relaydb)
			deadRelays := relaydb.CheckForTimeouts(2000)
			assert.Empty(t, deadRelays)
			assert.Len(t, relaydb.Relays, 6)
		})
	})

	t.Run("MakeCopy()", func(t *testing.T) {
		t.Run("returns an exact copy", func(t *testing.T) {
			relaydb := routing.NewRelayDatabase()
			FillRelayDatabase(relaydb)
			cpy := relaydb.MakeCopy()
			assert.Equal(t, relaydb, cpy)
		})
	})
}
