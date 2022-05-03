package routing_test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"

	"github.com/stretchr/testify/assert"
)

func testRelayData(t *testing.T) *routing.RelayData {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", rand.Int31n(50000)))
	assert.NoError(t, err)

	relay := new(routing.RelayData)
	relay.ID = rand.Uint64()
	relay.Name = backend.GenerateRandomStringSequence(routing.MaxRelayNameLength)

	relay.Addr = *addr
	relay.PublicKey = make([]byte, crypto.KeySize)
	relay.MaxSessions = 0
	relay.SessionCount = rand.Int()
	relay.ShuttingDown = false
	relay.LastUpdateTime = time.Unix(rand.Int63(), 0)
	relay.Version = "2.1.0"
	relay.CPU = uint8(rand.Int())
	relay.NICSpeedMbps = int32(1000)
	relay.MaxBandwidthMbps = int32(900)
	relay.EnvelopeUpMbps = rand.Float32()
	relay.EnvelopeDownMbps = rand.Float32()
	relay.BandwidthSentMbps = rand.Float32()
	relay.BandwidthRecvMbps = rand.Float32()

	return relay
}

func TestRelayMap_NewRelayMap(t *testing.T) {
	t.Parallel()

	rmap := routing.NewRelayMap(func(relay routing.RelayData) error {
		return nil
	})

	assert.NotNil(t, rmap)
}

func TestRelayMap_TimeoutLoop(t *testing.T) {
	t.Parallel()

	callbackChan := make(chan bool, 1)

	expiredRelays := new(int)
	rmap := routing.NewRelayMap(func(relay routing.RelayData) error {
		(*expiredRelays)++
		callbackChan <- true
		return nil
	})

	relay := testRelayData(t)
	relay.LastUpdateTime = time.Unix(time.Now().Unix()-2, 0)
	rmap.UpdateRelayData(*relay)

	getRelayData := func() ([]routing.Relay, map[uint64]routing.Relay) {
		relayHash := make(map[uint64]routing.Relay)
		return []routing.Relay{}, relayHash
	}

	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		var timeout int64 = 1
		frequency := time.Millisecond * 100
		ticker := time.NewTicker(frequency)
		rmap.TimeoutLoop(ctx, getRelayData, timeout, ticker.C)
	}()

	<-callbackChan
	ctxCancelFunc()
	wg.Wait()

	assert.Equal(t, 1, *expiredRelays)
	assert.Zero(t, rmap.GetRelayCount())
}

func TestRelayMap_UpdateRelayData(t *testing.T) {
	t.Parallel()

	rmap := routing.NewRelayMap(func(relay routing.RelayData) error {
		return nil
	})

	relay := testRelayData(t)
	rmap.UpdateRelayData(*relay)
	assert.Equal(t, uint64(1), rmap.GetRelayCount())
}

func TestRelayMap_GetRelayData(t *testing.T) {
	t.Parallel()

	rmap := routing.NewRelayMap(func(relay routing.RelayData) error {
		return nil
	})

	relay := testRelayData(t)

	data, exists := rmap.GetRelayData(relay.Addr.String())
	assert.False(t, exists)
	assert.Equal(t, routing.RelayData{}, data)

	rmap.UpdateRelayData(*relay)
	assert.Equal(t, uint64(1), rmap.GetRelayCount())

	data, exists = rmap.GetRelayData(relay.Addr.String())
	assert.True(t, exists)
	assert.Equal(t, data, *relay)
}

func TestRelayMap_GetActiveRelayData(t *testing.T) {
	t.Parallel()

	rmap := routing.NewRelayMap(func(relay routing.RelayData) error {
		return nil
	})

	relay1 := testRelayData(t)
	rmap.UpdateRelayData(*relay1)

	relay2 := testRelayData(t)
	relay2.ShuttingDown = true
	rmap.UpdateRelayData(*relay2)

	assert.Equal(t, uint64(2), rmap.GetRelayCount())

	relayIDs, relaySessionCounts, relayVersions := rmap.GetActiveRelayData()
	assert.Equal(t, len(relayIDs), 1)
	assert.Equal(t, relayIDs[0], relay1.ID)

	assert.Equal(t, len(relaySessionCounts), 1)
	assert.Equal(t, relaySessionCounts[0], relay1.SessionCount)

	assert.Equal(t, len(relayVersions), 1)
	assert.Equal(t, relayVersions[0], relay1.Version)
}

func TestRelayMap_GetAllRelayData(t *testing.T) {
	t.Parallel()

	rmap := routing.NewRelayMap(func(relay routing.RelayData) error {
		return nil
	})

	relay1 := testRelayData(t)
	rmap.UpdateRelayData(*relay1)

	relay2 := testRelayData(t)
	rmap.UpdateRelayData(*relay2)

	assert.Equal(t, uint64(2), rmap.GetRelayCount())

	allData := rmap.GetAllRelayData()
	assert.Equal(t, len(allData), 2)

	var foundRelay1 bool
	var foundRelay2 bool
	for _, data := range allData {
		if data.Addr.String() != relay1.Addr.String() {
			assert.Equal(t, data, *relay2)
			foundRelay2 = true
		} else {
			assert.Equal(t, data, *relay1)
			foundRelay1 = true
		}
	}
	assert.True(t, foundRelay1)
	assert.True(t, foundRelay2)
}

func TestRelayMap_GetAllRelayAddresses(t *testing.T) {
	t.Parallel()

	rmap := routing.NewRelayMap(func(relay routing.RelayData) error {
		return nil
	})

	relay1 := testRelayData(t)
	rmap.UpdateRelayData(*relay1)

	relay2 := testRelayData(t)
	rmap.UpdateRelayData(*relay2)

	assert.Equal(t, uint64(2), rmap.GetRelayCount())

	allAddrs := rmap.GetAllRelayAddresses()
	assert.Equal(t, len(allAddrs), 2)

	var foundRelay1 bool
	var foundRelay2 bool
	for _, addr := range allAddrs {
		if addr != relay1.Addr.String() {
			assert.Equal(t, addr, relay2.Addr.String())
			foundRelay2 = true
		} else {
			assert.Equal(t, addr, relay1.Addr.String())
			foundRelay1 = true
		}
	}
	assert.True(t, foundRelay1)
	assert.True(t, foundRelay2)
}

func TestRelayMap_RemoveRelayData(t *testing.T) {
	t.Parallel()

	rmap := routing.NewRelayMap(func(relay routing.RelayData) error {
		return nil
	})

	relay1 := testRelayData(t)

	assert.Zero(t, rmap.GetRelayCount())

	rmap.RemoveRelayData(relay1.Addr.String())

	assert.Zero(t, rmap.GetRelayCount())

	rmap.UpdateRelayData(*relay1)
	assert.Equal(t, uint64(1), rmap.GetRelayCount())

	rmap.RemoveRelayData(relay1.Addr.String())
	assert.Zero(t, rmap.GetRelayCount())
}
