package routing_test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func newRelay() *routing.RelayData {
	relay := new(routing.RelayData)
	relay.ID = rand.Uint64()

	bufflen := 26 * 8
	buff := make([]byte, bufflen)
	for i := 0; i < bufflen; i++ {
		buff[i] = byte(rand.Int())
	}
	index := 0
	relay.TrafficStats.ReadFrom(buff, &index, 2)
	relay.Version = fmt.Sprintf("%d.%d.%d", byte(rand.Uint32()), byte(rand.Uint32()), byte(rand.Uint32()))
	relay.LastUpdateTime = time.Unix(rand.Int63(), 0)
	relay.CPUUsage = rand.Float32()
	relay.MemUsage = rand.Float32()

	return relay
}

func TestRelayMapTimeoutLoop(t *testing.T) {
	expiredRelays := new(int)
	rmap := routing.NewRelayMap(func(relay *routing.RelayData) error {
		(*expiredRelays)++
		return nil
	})

	var relay routing.RelayData
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	relay.Addr = *addr
	relay.LastUpdateTime = time.Unix(time.Now().Unix()-2, 0)
	rmap.AddRelayDataEntry(relay.Addr.String(), &relay)

	ctx := context.Background()

	go func() {
		var timeout int64 = 1
		frequency := time.Millisecond * 100
		ticker := time.NewTicker(frequency)
		rmap.TimeoutLoop(ctx, timeout, ticker.C)
	}()

	time.Sleep(time.Millisecond * 200)

	assert.Equal(t, 1, *expiredRelays)
	assert.Zero(t, rmap.GetRelayCount())

	ctx.Done()
}

func TestRelayMapGetAllRelayData(t *testing.T) {
	relays := make([]*routing.RelayData, 10)
	rmap := routing.NewRelayMap(func(relay *routing.RelayData) error { return nil })
	for i := 0; i < len(relays); i++ {
		relay := newRelay()
		relays[i] = relay
		addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", 10000+i))
		relay.Addr = *addr
		rmap.AddRelayDataEntry(relay.Addr.String(), relay)
	}

	for _, relay := range rmap.GetAllRelayData() {
		relay.Version = "some other version"
		expected := rmap.GetRelayData(relay.Addr.String())
		assert.NotNil(t, expected)
		assert.NotEqual(t, relay.Version, expected.Version)
	}
}

func TestRelayMapGetAllRelayIDs(t *testing.T) {
	excludeSeller := new(routing.Seller)
	excludeSeller.ID = "valve"
	normalSeller := new(routing.Seller)
	normalSeller.ID = "normal"

	rmap := routing.NewRelayMap(func(relay *routing.RelayData) error { return nil })
	for i := 0; i < 6; i++ {
		relay := newRelay()
		relay.ID = uint64(i)
		if i == 0 || i == 3{
			relay.Seller = *excludeSeller
		}else{
			relay.Seller = *normalSeller
		}
		addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", 10000+i))
		relay.Addr = *addr
		rmap.AddRelayDataEntry(relay.Addr.String(), relay)
	}

	relayIDs := rmap.GetAllRelayIDs([]string{})
	assert.Equal(t, 6, len(relayIDs))

	relayIDsWithExclude := rmap.GetAllRelayIDs([]string{excludeSeller.ID})
	assert.Equal(t,4, len(relayIDsWithExclude))
	for _, relayID := range relayIDsWithExclude{
		assert.NotEqual(t,0, relayID)
		assert.NotEqual(t,3, relayID)
	}
}

func TestRelayMapRemoveRelay(t *testing.T) {
	removedRelays := new(int)
	rmap := routing.NewRelayMap(func(relay *routing.RelayData) error {
		(*removedRelays)++
		return nil
	})

	relays := make([]*routing.RelayData, 10)
	for i := 0; i < len(relays); i++ {
		relay := newRelay()
		relays[i] = relay
		rmap.AddRelayDataEntry(fmt.Sprintf("127.0.0.1:%d", 10000+i), relay)
	}

	rmap.RemoveRelayData("127.0.0.1:10000")
	assert.Equal(t, 1, *removedRelays)
	assert.Equal(t, uint64(9), rmap.GetRelayCount())
}

func TestRelayMapMarshalBinary(t *testing.T) {
	t.Run("invalid version", func(t *testing.T) {
		rmap := routing.NewRelayMap(func(relay *routing.RelayData) error { return nil })
		for i := 0; i < 10; i++ {
			relay := newRelay()
			relay.Version = "invalid version"
			rmap.AddRelayDataEntry(fmt.Sprintf("127.0.0.1:%d", 10000+i), relay)
		}

		bin, err := rmap.MarshalBinary()
		assert.Nil(t, bin)
		assert.Error(t, err)
	})

	t.Run("valid versions", func(t *testing.T) {
		relays := make([]*routing.RelayData, 10)
		rmap := routing.NewRelayMap(func(relay *routing.RelayData) error { return nil })
		for i := 0; i < len(relays); i++ {
			relay := newRelay()
			relays[i] = relay
			rmap.AddRelayDataEntry(fmt.Sprintf("127.0.0.1:%d", 10000+i), relay)
		}

		bin, err := rmap.MarshalBinary()
		assert.NotNil(t, bin)
		assert.NoError(t, err)

		index := 0

		var version uint8
		assert.True(t, encoding.ReadUint8(bin, &index, &version))

		assert.Equal(t, uint8(routing.VersionNumberRelayMap), version)

		var numRelays uint64
		assert.True(t, encoding.ReadUint64(bin, &index, &numRelays))

		assert.Equal(t, uint64(10), numRelays)

		checkedRelays := make(map[uint64]bool)
		for i := uint64(0); i < numRelays; i++ {
			var relay routing.RelayData
			assert.True(t, encoding.ReadUint64(bin, &index, &relay.ID))

			assert.NoError(t, relay.TrafficStats.ReadFrom(bin, &index, 2))

			var major, minor, patch uint8
			assert.True(t, encoding.ReadUint8(bin, &index, &major))
			assert.True(t, encoding.ReadUint8(bin, &index, &minor))
			assert.True(t, encoding.ReadUint8(bin, &index, &patch))
			relay.Version = fmt.Sprintf("%d.%d.%d", major, minor, patch)

			var updateTime uint64
			assert.True(t, encoding.ReadUint64(bin, &index, &updateTime))
			relay.LastUpdateTime = time.Unix(int64(updateTime), 0)

			assert.True(t, encoding.ReadFloat32(bin, &index, &relay.CPUUsage))

			assert.True(t, encoding.ReadFloat32(bin, &index, &relay.MemUsage))

			// relays are written via iterating the map, so they are in a semi-random order
			var expected *routing.RelayData = nil
			for j := uint64(0); j < numRelays; j++ {
				if relay.ID == relays[j].ID {
					expected = relays[j]
					checkedRelays[relay.ID] = true
					break
				}
			}
			assert.NotNil(t, expected)

			assert.Equal(t, expected.ID, relay.ID)
			assert.Equal(t, expected.TrafficStats, relay.TrafficStats)
			assert.Equal(t, expected.Version, relay.Version)
			assert.Equal(t, expected.LastUpdateTime, relay.LastUpdateTime)
			assert.Equal(t, expected.CPUUsage, relay.CPUUsage)
			assert.Equal(t, expected.MemUsage, relay.MemUsage)
		}
		assert.Equal(t, numRelays, uint64(len(checkedRelays)))
	})
}
