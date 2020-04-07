package storage

import (
	"github.com/networknext/backend/routing"
)

type Storer interface {
	Buyer(uint64) (*routing.Buyer, bool)
	Buyers() []routing.Buyer
	Relay(uint64) (*routing.Relay, bool)
	Relays() []routing.Relay
	Datacenters() []routing.Datacenter
}
