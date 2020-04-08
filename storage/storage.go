package storage

import (
	"context"

	"github.com/networknext/backend/routing"
)

type Storer interface {
	// Buyer gets a copy of a buyer with the specified buyer ID,
	// and returns an empty buyer and an error if a buyer with that ID doesn't exist in storage.
	Buyer(id uint64) (routing.Buyer, error)

	// Buyers returns a copy of all stored buyers.
	Buyers() []routing.Buyer

	// AddBuyer adds the provided buyer to storage and returns an error if the buyer could not be added.
	AddBuyer(ctx context.Context, buyer routing.Buyer) error

	// RemoveBuyer removes a buyer with the provided buyer ID from storage and returns an error if the buyer could not be removed.
	RemoveBuyer(ctx context.Context, id uint64) error

	// SetBuyer updates the buyer in storage with the provided copy and returns an error if the buyer could not be updated.
	SetBuyer(ctx context.Context, buyer routing.Buyer) error

	// Relay gets a copy of a relay with the specified relay ID
	// and returns an empty relay and an error if a relay with that ID doesn't exist in storage.
	Relay(id uint64) (routing.Relay, error)

	// Relays returns a copy of all stored relays.
	Relays() []routing.Relay

	// AddRelay adds the provided relay to storage and returns an error if the relay could not be added.
	AddRelay(ctx context.Context, relay routing.Relay) error

	// RemoveRelay removes a relay with the provided relay ID from storage and returns an error if the relay could not be removed.
	RemoveRelay(ctx context.Context, id uint64) error

	// SetRelay updates the relay in storage with the provided copy and returns an error if the relay could not be updated.
	SetRelay(ctx context.Context, relay routing.Relay) error

	// Datacenter gets a copy of a datacenter with the specified datacenter ID
	// and returns an empty datacenter and an error if a datacenter with that ID doesn't exist in storage.
	Datacenter(id uint64) (routing.Datacenter, error)

	// Datacenters returns a copy of all stored datacenters.
	Datacenters() []routing.Datacenter

	// AddDatacenter adds the provided datacenter to storage and returns an error if the datacenter could not be added.
	AddDatacenter(ctx context.Context, datacenter routing.Datacenter) error

	// RemoveDatacenter removes a datacenter with the provided datacenter ID from storage and returns an error if the datacenter could not be removed.
	RemoveDatacenter(ctx context.Context, id uint64) error

	// SetDatacenter updates the datacenter in storage with the provided copy and returns an error if the datacenter could not be updated.
	SetDatacenter(ctx context.Context, datacenter routing.Datacenter) error
}
