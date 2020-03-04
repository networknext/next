package storage

import (
	"github.com/networknext/backend/routing"
)

type InMemory struct {
	LocalBuyer  *routing.Buyer
	LocalRelays []routing.Relay
}

func (m *InMemory) Buyer(id uint64) (*routing.Buyer, bool) {
	if m.LocalBuyer != nil {
		return m.LocalBuyer, true
	}

	return nil, false
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
