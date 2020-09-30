package routing_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func TestRelayMapTimeoutLoop(t *testing.T) {
	expiredRelays := new(int)
	rmap := routing.NewRelayMap(func(relay *routing.RelayData) error {
		(*expiredRelays)++
		return nil
	})

	var relay routing.RelayData
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12345")
	assert.NoError(t, err)
	relay.Addr = *addr
	relay.LastUpdateTime = time.Unix(time.Now().Unix()-2, 0)
	rmap.UpdateRelayData(relay.Addr.String(), &relay)

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

func TestRelayMapMarshalBinary(t *testing.T) {
	rmap := routing.NewRelayMap(func(relay *routing.RelayData) error { return nil })
	for i := 0; i < 10; i++ {
		relay := new(routing.RelayData)
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", 10000+i))
		assert.NoError(t, err)
		relay.Addr = *addr
		rmap.UpdateRelayData(addr.String(), relay)
	}

	t.Run("invalid version", func(t *testing.T) {
		for _, relay := range rmap.GetAllRelayData() {
			relay.Version = "invalid version"
		}

		bin, err := rmap.MarshalBinary()
		assert.Nil(t, bin)
		assert.Error(t, err)
	})

	t.Run("valid versions", func(t *testing.T) {
		for _, relay := range rmap.GetAllRelayData() {
			relay.Version = "1.0.0"
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
	})
}
