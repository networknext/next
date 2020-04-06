package storage

import (
	"context"

	"github.com/networknext/backend/routing"
)

type Storer interface {
	Buyer(uint64) (*routing.Buyer, bool)
	Relay(uint64) (*routing.Relay, bool)
	Relays() []routing.Relay
	SetRelayState(ctx context.Context, relay *routing.Relay) error
	Datacenters() []routing.Datacenter
}
