package storage

import (
	"context"
	"fmt"

	"github.com/networknext/backend/routing"
)

type InMemory struct {
	localBuyers      []routing.Buyer
	localRelays      []routing.Relay
	localDatacenters []routing.Datacenter
}

func (m *InMemory) Buyer(id uint64) (routing.Buyer, error) {
	for _, buyer := range m.localBuyers {
		if buyer.ID == id {
			return buyer, nil
		}
	}

	return routing.Buyer{}, fmt.Errorf("buyer with id %d not found in memory storage", id)
}

func (m *InMemory) Buyers() []routing.Buyer {
	buyers := make([]routing.Buyer, len(m.localBuyers))
	for i := range buyers {
		buyers[i] = m.localBuyers[i]
	}

	return buyers
}

func (m *InMemory) AddBuyer(ctx context.Context, buyer routing.Buyer) error {
	for _, b := range m.localBuyers {
		if b.ID == buyer.ID {
			return fmt.Errorf("buyer with id %d already exists in memory storage", buyer.ID)
		}
	}

	m.localBuyers = append(m.localBuyers, buyer)
	return nil
}

func (m *InMemory) RemoveBuyer(ctx context.Context, id uint64) error {
	buyerIndex := -1
	for i, buyer := range m.localBuyers {
		if buyer.ID == id {
			buyerIndex = i
		}
	}

	if buyerIndex < 0 {
		return fmt.Errorf("buyer with id %d not found in memory storage", id)
	}

	if buyerIndex+1 == len(m.localBuyers) {
		m.localBuyers = m.localBuyers[:buyerIndex]
		return nil
	}

	frontSlice := m.localBuyers[:buyerIndex]
	backSlice := m.localBuyers[buyerIndex+1:]
	m.localBuyers = append(frontSlice, backSlice...)
	return nil
}

func (m *InMemory) SetBuyer(ctx context.Context, buyer routing.Buyer) error {
	for i := range m.localBuyers {
		if m.localBuyers[i].ID == buyer.ID {
			m.localBuyers[i] = buyer
			return nil
		}
	}

	return fmt.Errorf("buyer with id %d not found in memory storage", buyer.ID)
}

func (m *InMemory) Relay(id uint64) (routing.Relay, error) {
	for _, relay := range m.localRelays {
		if relay.ID == id {
			return relay, nil
		}
	}

	return routing.Relay{}, fmt.Errorf("relay with id %d not found in memory storage", id)
}

func (m *InMemory) Relays() []routing.Relay {
	relays := make([]routing.Relay, len(m.localRelays))
	for i := range relays {
		relays[i] = m.localRelays[i]
	}

	return relays
}

func (m *InMemory) AddRelay(ctx context.Context, relay routing.Relay) error {
	for _, r := range m.localRelays {
		if r.ID == relay.ID {
			return fmt.Errorf("relay with id %d already exists in memory storage", relay.ID)
		}
	}

	m.localRelays = append(m.localRelays, relay)
	return nil
}

func (m *InMemory) RemoveRelay(ctx context.Context, id uint64) error {
	relayIndex := -1
	for i, relay := range m.localRelays {
		if relay.ID == id {
			relayIndex = i
		}
	}

	if relayIndex < 0 {
		return fmt.Errorf("relay with id %d not found in memory storage", id)
	}

	if relayIndex+1 == len(m.localRelays) {
		m.localRelays = m.localRelays[:relayIndex]
		return nil
	}

	frontSlice := m.localRelays[:relayIndex]
	backSlice := m.localRelays[relayIndex+1:]
	m.localRelays = append(frontSlice, backSlice...)
	return nil
}

func (m *InMemory) SetRelay(ctx context.Context, relay routing.Relay) error {
	for i := 0; i < len(m.localRelays); i++ {
		if m.localRelays[i].ID == relay.ID {
			m.localRelays[i] = relay
			return nil
		}
	}

	return fmt.Errorf("relay with id %d not found in memory storage", relay.ID)
}

func (m *InMemory) Datacenter(id uint64) (routing.Datacenter, error) {
	for _, datacenter := range m.localDatacenters {
		if datacenter.ID == id {
			return datacenter, nil
		}
	}

	return routing.Datacenter{}, fmt.Errorf("datacenter with id %d not found in memory storage", id)
}

func (m *InMemory) Datacenters() []routing.Datacenter {
	datacenters := make([]routing.Datacenter, len(m.localDatacenters))
	for i := range datacenters {
		datacenters[i] = m.localDatacenters[i]
	}

	return datacenters
}

func (m *InMemory) AddDatacenter(ctx context.Context, datacenter routing.Datacenter) error {
	for _, d := range m.localDatacenters {
		if d.ID == datacenter.ID {
			return fmt.Errorf("datacenter with id %d already exists in memory storage", datacenter.ID)
		}
	}

	m.localDatacenters = append(m.localDatacenters, datacenter)
	return nil
}

func (m *InMemory) RemoveDatacenter(ctx context.Context, id uint64) error {
	datacenterIndex := -1
	for i, datacenter := range m.localDatacenters {
		if datacenter.ID == id {
			datacenterIndex = i
		}
	}

	if datacenterIndex < 0 {
		return fmt.Errorf("datacenter with id %d not found in memory storage", id)
	}

	if datacenterIndex+1 == len(m.localDatacenters) {
		m.localDatacenters = m.localDatacenters[:datacenterIndex]
		return nil
	}

	frontSlice := m.localDatacenters[:datacenterIndex]
	backSlice := m.localDatacenters[datacenterIndex+1:]
	m.localDatacenters = append(frontSlice, backSlice...)
	return nil
}

func (m *InMemory) SetDatacenter(ctx context.Context, datacenter routing.Datacenter) error {
	for i := 0; i < len(m.localDatacenters); i++ {
		if m.localDatacenters[i].ID == datacenter.ID {
			m.localDatacenters[i] = datacenter
			return nil
		}
	}

	return fmt.Errorf("datacenter with id %d not found in memory storage", datacenter.ID)
}
