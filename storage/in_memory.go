package storage

import (
	"github.com/networknext/backend/routing"
)

type InMemory struct {
	LocalBuyer  *routing.Buyer
	LocalRelays []*routing.Relay
}

func (m *InMemory) Buyer(id uint64) (*routing.Buyer, bool) {
	if m.LocalBuyer != nil {
		return m.LocalBuyer, true
	}

	return nil, false
}

func (m *InMemory) Relay(id uint64) (*routing.Relay, bool) {
	for _, relay := range m.LocalRelays {
		if relay.ID == id {
			return relay, true
		}
	}

	return nil, false
}
