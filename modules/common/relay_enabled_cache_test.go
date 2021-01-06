package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
)

func TestRelayEnabledCache_runner(t *testing.T) {
	relays := make([]routing.Relay, 3)
	relays[0] = routing.Relay{
		ID:    0,
		Name:  "relay0",
		State: 0,
	}

	relays[1] = routing.Relay{
		ID:    1,
		Name:  "relay1",
		State: 1,
	}

	relays[2] = routing.Relay{
		ID:    2,
		Name:  "relay2",
		State: 0,
	}

	var storer storage.Storer
	storer = &storage.StorerMock{RelaysFunc: func() []routing.Relay {
		return relays
	}}

	rec := NewRelayEnabledCache(storer)
	rec.runner()

	assert.Equal(t, 2, len(rec.activeRelays))
	assert.Equal(t, "relay0", rec.activeRelays[0].name)
	assert.Equal(t, uint64(0), rec.activeRelays[0].id)
	assert.Equal(t, "relay2", rec.activeRelays[1].name)
	assert.Equal(t, uint64(2), rec.activeRelays[1].id)
}

func TestRelayEnabledCache_GetEnabledRelays(t *testing.T) {
	relays := make([]routing.Relay, 3)
	relays[0] = routing.Relay{
		ID:    0,
		Name:  "relay0",
		State: 0,
	}

	relays[1] = routing.Relay{
		ID:    1,
		Name:  "relay1",
		State: 1,
	}

	relays[2] = routing.Relay{
		ID:    2,
		Name:  "relay2",
		State: 0,
	}

	var storer storage.Storer
	storerm := &storage.StorerMock{RelaysFunc: func() []routing.Relay {
		return relays
	}}
	storer = storerm

	rec := NewRelayEnabledCache(storer)
	rec.runner()

	relayNames, relayIDs := rec.GetEnabledRelays()
	assert.Equal(t, 2, len(relayNames))
	assert.Equal(t, "relay0", relayNames[0])
	assert.Equal(t, uint64(0), relayIDs[0])
	assert.Equal(t, "relay2", relayNames[1])
	assert.Equal(t, uint64(2), relayIDs[1])
}

func TestRelayEnabledCache_GetDownRelays(t *testing.T) {
	relays := make([]routing.Relay, 5)
	relays[0] = routing.Relay{
		ID:    0,
		Name:  "relay0",
		State: 0,
	}

	relays[1] = routing.Relay{
		ID:    1,
		Name:  "relay1",
		State: 1,
	}

	relays[2] = routing.Relay{
		ID:    2,
		Name:  "relay2",
		State: 0,
	}

	relays[3] = routing.Relay{
		ID:    3,
		Name:  "relay3",
		State: 0,
	}

	relays[4] = routing.Relay{
		ID:    4,
		Name:  "relay4",
		State: 0,
	}

	var storer storage.Storer
	storer = &storage.StorerMock{RelaysFunc: func() []routing.Relay {
		return relays
	}}
	rec := NewRelayEnabledCache(storer)
	rec.runner()
	runningRelayNames := []uint64{0, 3}

	downRelayNames, downRelayIDs := rec.GetDownRelays(runningRelayNames)
	assert.Equal(t, 2, len(downRelayNames))
	assert.Equal(t, "relay2", downRelayNames[0])
	assert.Equal(t, uint64(2), downRelayIDs[0])
	assert.Equal(t, "relay4", downRelayNames[1])
	assert.Equal(t, uint64(4), downRelayIDs[1])
}
