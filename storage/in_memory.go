package storage

import (
	"context"
	"errors"

	"github.com/networknext/backend/routing"
)

type InMemory struct {
	LocalBuyer       *routing.Buyer
	LocalRelays      []routing.Relay
	LocalDatacenters []routing.Datacenter
}

func (m *InMemory) Buyer(id uint64) (*routing.Buyer, bool) {
	if m.LocalBuyer != nil {
		return m.LocalBuyer, true
	}

	return nil, false
}

func (m *InMemory) Buyers() []routing.Buyer {
	return []routing.Buyer{*m.LocalBuyer}
}

func (m *InMemory) Relay(id uint64) (*routing.Relay, bool) {
	// Fail if literally nothing is set
	if len(m.LocalRelays) == 0 {
		return nil, false
	}

	// Do attempt to get appropriate relay if it exists
	for _, relay := range m.LocalRelays {
		if relay.ID == id {
			return &relay, true
		}
	}

	// Failing this, just return first one since we need something for local dev
	return &m.LocalRelays[0], true
}

func (m *InMemory) Relays() []routing.Relay {
	return m.LocalRelays
}

func (m *InMemory) SetRelayState(ctx context.Context, relay *routing.Relay) error {
	for i := 0; i < len(m.LocalRelays); i++ {
		if m.LocalRelays[i].ID == relay.ID {
			m.LocalRelays[i].State = relay.State
			return nil
		}
	}

	return errors.New("could not find relay")
}

func (m *InMemory) Datacenters() []routing.Datacenter {
	return m.LocalDatacenters
}
