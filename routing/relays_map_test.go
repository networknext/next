package routing_test

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"gopkg.in/go-playground/assert.v1"
)

func TestRelayMap(t *testing.T) {
	callbackHit := new(int)
	rmap := routing.NewRelayMap(func(relay *routing.RelayData) error {
		(*callbackHit)++
		return nil
	})

	relays := make([]routing.RelayData, 1000)

	var wg sync.WaitGroup
	wg.Add(10)

	// add relays asynchronously
	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 100; j++ {
				relay := &relays[i*100+j]
				addrStr := fmt.Sprintf("192.168.%d.%d:40000", i, j)
				addr, err := net.ResolveUDPAddr("udp", addrStr)
				assert.Equal(t, err, nil)
				relay.Addr = *addr
				relay.ID = crypto.HashID(addrStr)
				relay.LastUpdateTime = time.Now()
				rmap.Lock(addrStr)
				rmap.UpdateRelayData(addrStr, relay)
				rmap.Unlock(addrStr)
			}
			wg.Done()
		}(i)
	}

	wg.Wait()

	assert.Equal(t, rmap.GetRelayCount(), uint64(1000))

	ctx := context.Background()

	rmap.RemoveRelayData(relays[0].Addr.String())

	assert.Equal(t, rmap.GetRelayCount(), uint64(999))
	assert.Equal(t, callbackHit, 1)

	ctx.Done()
}
