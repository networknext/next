package storage

import (
	"context"
	"time"

	"github.com/networknext/backend/routing"
)

type Storer interface {
	Customer(code string) (routing.Customer, error)

	CustomerWithName(name string) (routing.Customer, error)

	Customers() []routing.Customer

	AddCustomer(ctx context.Context, customer routing.Customer) error

	RemoveCustomer(ctx context.Context, code string) error

	SetCustomer(ctx context.Context, customer routing.Customer) error

	// Buyer gets a copy of a buyer with the specified buyer ID,
	// and returns an empty buyer and an error if a buyer with that ID doesn't exist in storage.
	Buyer(id uint64) (routing.Buyer, error)

	// BuyerWithCompanyCode gets the Buyer with the matching company code
	BuyerWithCompanyCode(code string) (routing.Buyer, error)

	// Buyers returns a copy of all stored buyers.
	Buyers() []routing.Buyer

	// AddBuyer adds the provided buyer to storage and returns an error if the buyer could not be added.
	AddBuyer(ctx context.Context, buyer routing.Buyer) error

	// RemoveBuyer removes a buyer with the provided buyer ID from storage and returns an error if the buyer could not be removed.
	RemoveBuyer(ctx context.Context, id uint64) error

	// SetBuyer updates the buyer in storage with the provided copy and returns an error if the buyer could not be updated.
	SetBuyer(ctx context.Context, buyer routing.Buyer) error

	// Seller gets a copy of a seller with the specified seller ID,
	// and returns an empty seller and an error if a seller with that ID doesn't exist in storage.
	Seller(id string) (routing.Seller, error)

	// Sellers returns a copy of all stored sellers.
	Sellers() []routing.Seller

	// AddSeller adds the provided seller to storage and returns an error if the seller could not be added.
	AddSeller(ctx context.Context, seller routing.Seller) error

	// RemoveSeller removes a seller with the provided seller ID from storage and returns an error if the seller could not be removed.
	RemoveSeller(ctx context.Context, id string) error

	// SetSeller updates the seller in storage with the provided copy and returns an error if the seller could not be updated.
	SetSeller(ctx context.Context, seller routing.Seller) error

	// BuyerIDFromCustomerName returns the buyer ID associated with the given customer name and an error if the customer wasn't found.
	// If the customer has no buyer linked, then it will return a buyer ID of 0 and no error.
	BuyerIDFromCustomerName(ctx context.Context, customerName string) (uint64, error)

	// SellerIDFromCustomerName returns the seller ID associated with the given customer name and an error if the customer wasn't found.
	// If the customer has no seller linked, then it will return an empty seller ID and no error.
	SellerIDFromCustomerName(ctx context.Context, customerName string) (string, error)

	SellerWithCompanyCode(code string) (routing.Seller, error)

	// SetCustomerLink update the customer's buyer and seller references.
	SetCustomerLink(ctx context.Context, customerName string, buyerID uint64, sellerID string) error

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
	Datacenter(datacenterID uint64) (routing.Datacenter, error)

	// Datacenters returns a copy of all stored datacenters.
	Datacenters() []routing.Datacenter

	// AddDatacenter adds the provided datacenter to storage and returns an error if the datacenter could not be added.
	AddDatacenter(ctx context.Context, datacenter routing.Datacenter) error

	// RemoveDatacenter removes a datacenter with the provided datacenter ID from storage and returns an error if the datacenter could not be removed.
	RemoveDatacenter(ctx context.Context, id uint64) error

	// SetDatacenter updates the datacenter in storage with the provided copy and returns an error if the datacenter could not be updated.
	SetDatacenter(ctx context.Context, datacenter routing.Datacenter) error

	// GetDatacenterMapsForBuyer returns the list of datacenter aliases in use for a given (internally generated) buyerID. Returns
	// an empty []routing.DatacenterMap if there are no aliases for that buyerID.
	GetDatacenterMapsForBuyer(buyerID uint64) map[uint64]routing.DatacenterMap

	// AddDatacenterMap adds a new datacenter alias for the given buyer and datacenter IDs
	AddDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error

	// ListDatacenterMaps returns a list of alias/buyer mappings for the specified datacenter ID. An
	// empty dcID returns a list of all maps.
	ListDatacenterMaps(dcID uint64) map[uint64]routing.DatacenterMap

	// RemoveDatacenterMap removes an entry from the DatacenterMaps table
	RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error

	// SetRelayMetadata provides write access to ops metadat (mrc, overage, etc)
	SetRelayMetadata(ctx context.Context, relay routing.Relay) error

	// CheckSequenceNumber is called in the sync*() operations to see if a sync is required.
	CheckSequenceNumber(ctx context.Context) (bool, int64, error)

	// IncrementSequenceNumber is used by all methods that make changes to the db
	IncrementSequenceNumber(ctx context.Context) error

	// SyncLoop sets up the ticker for database syncs
	SyncLoop(ctx context.Context, c <-chan time.Time)

	// New for ConfigService

	// GetFeatureFlags returns all feature flags currently in the database
	GetFeatureFlags() map[string]bool

	// GetFeatureFlagByName returns a specific flag or an error if it does not exist
	GetFeatureFlagByName(flagName string) (map[string]bool, error)

	// SetFeatureFlagByName adds a new feature or updates the value of an existing feature
	SetFeatureFlagByName(ctx context.Context, flagName string, flagVal bool) error

	// RemoveFeatureFlagByName removes an existing flag from storage
	RemoveFeatureFlagByName(ctx context.Context, flagName string) error
}
