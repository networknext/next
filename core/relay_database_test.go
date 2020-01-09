package core_test

import (
	"hash/fnv"
	"testing"
	"time"

	"github.com/networknext/backend/core"
	"github.com/stretchr/testify/assert"
)

func TestRelayDatabase(t *testing.T) {
	fillDB := func(relaydb *core.RelayDatabase) {
		fillData := func(relaydb *core.RelayDatabase, addr string, updateTime int64) {
			id := core.GetRelayID(addr)
			data := core.RelayData{
				ID:             id,
				Name:           "n/a",
				Address:        addr,
				Datacenter:     core.DatacenterId(123),
				DatacenterName: "n/a",
				PublicKey:      []byte{0x01, 0x02, 0x03, 0x04},
				LastUpdateTime: uint64(updateTime),
			}
			relaydb.Relays[id] = data
		}

		fillData(relaydb, "127.0.0.1", time.Now().Unix()-1)
		fillData(relaydb, "123.4.5.6", time.Now().Unix()-10)
		fillData(relaydb, "654.3.2.1", time.Now().Unix()-100)
		fillData(relaydb, "000.0.0.0", time.Now().Unix()-25)
		fillData(relaydb, "999.9.9.9", time.Now().Unix()-1000)
	}

	t.Run("UpdateRelay()", func(t *testing.T) {
		t.Run("shutdown = true also deletes database entry", func(t *testing.T) {
			relaydb := core.NewRelayDatabase()
			addr := "127.0.0.1"
			id := core.GetRelayID(addr)
			relaydb.Relays[id] = core.RelayData{}

			update := core.RelayUpdate{}
			update.Shutdown = true
			update.ID = id

			_, ok := relaydb.Relays[id]
			assert.True(t, ok)
			assert.False(t, relaydb.UpdateRelay(&update))
			_, ok = relaydb.Relays[id]
			assert.False(t, ok)
		})

		t.Run("relay did not already exist", func(t *testing.T) {
			relaydb := core.NewRelayDatabase()
			update := core.RelayUpdate{}
			assert.True(t, relaydb.UpdateRelay(&update))
		})

		t.Run("relay did already exist", func(t *testing.T) {
			relaydb := core.NewRelayDatabase()
			addr := "127.0.0.1"
			id := core.GetRelayID(addr)
			relaydb.Relays[id] = core.RelayData{}

			update := core.RelayUpdate{}
			update.ID = id

			assert.False(t, relaydb.UpdateRelay(&update))
		})

		t.Run("updates correctly", func(t *testing.T) {
			relaydb := core.NewRelayDatabase()
			addr := "127.0.0.1"
			id := core.GetRelayID(addr)
			update := core.RelayUpdate{
				ID:             id,
				Name:           "I don't know what this is supposed to be",
				Address:        addr,
				Datacenter:     core.DatacenterId(123),
				DatacenterName: "I also don't know what a good stub name should be",
				PublicKey:      []byte{0x01, 0x02, 0x03, 0x04},
				Shutdown:       false,
			}

			assert.True(t, relaydb.UpdateRelay(&update))
			value, ok := relaydb.Relays[id]
			assert.True(t, ok)

			// is there a go equivalent for c++ operator== overloading? or Java's .equal() method? Googling did me no help
			assert.Equal(t, update.ID, value.ID)
			assert.Equal(t, update.Name, value.Name)
			assert.Equal(t, update.Address, value.Address)
			assert.Equal(t, update.Datacenter, value.Datacenter)
			assert.Equal(t, update.DatacenterName, value.DatacenterName)
			assert.Equal(t, update.PublicKey, value.PublicKey)
		})
	})

	/* MAY OCCASIONALLY FAIL DUE TO TIMING if so rerun and pray */
	t.Run("CheckForTimeouts()", func(t *testing.T) {

		t.Run("dead relays are present", func(t *testing.T) {
			relaydb := core.NewRelayDatabase()
			fillDB(relaydb)
			expectedDeadRelays := []core.RelayId{core.GetRelayID("654.3.2.1"), core.GetRelayID("999.9.9.9")}

			deadRelays := relaydb.CheckForTimeouts(50)
			assert.Equal(t, expectedDeadRelays, deadRelays)
			for _, id := range expectedDeadRelays {
				_, ok := relaydb.Relays[id]
				assert.False(t, ok, "ID: %x", id)
			}
		})

		t.Run("all relays are alive", func(t *testing.T) {
			relaydb := core.NewRelayDatabase()
			fillDB(relaydb)
			deadRelays := relaydb.CheckForTimeouts(2000)
			assert.Empty(t, deadRelays)
			assert.Len(t, relaydb.Relays, 5)
		})
	})

	t.Run("MakeCopy()", func(t *testing.T) {
		t.Run("returns an exact copy", func(t *testing.T) {
			relaydb := core.NewRelayDatabase()
			fillDB(relaydb)
			cpy := relaydb.MakeCopy()
			assert.Equal(t, relaydb, cpy)
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
