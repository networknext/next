package routing_test

import (
	"context"
	"net"
	"testing"
	"time"

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
	relay.LastUpdateTime = time.Now().Add(-time.Second * 2)
	rmap.UpdateRelayData(relay.Addr.String(), &relay)

	ctx := context.Background()

	go func() {
		timeout := int64(time.Second)
		frequency := time.Millisecond * 100
		ticker := time.NewTicker(frequency)
		rmap.TimeoutLoop(ctx, timeout, ticker.C)
	}()

	time.Sleep(time.Second * 2)

	assert.Equal(t, 1, *expiredRelays)

	ctx.Done()
}
