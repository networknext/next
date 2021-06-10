//go:generate moq -out storage_test_mocks.go . Storer
package storage

import (
	"context"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/routing"
)

type Storer interface {
	Customer(code string) (routing.Customer, error)

	// TODO: chopping block (unused)
	CustomerWithName(name string) (routing.Customer, error)

	Customers() []routing.Customer

	AddCustomer(ctx context.Context, customer routing.Customer) error

	RemoveCustomer(ctx context.Context, code string) error

	// TODO: chopping block (this is dangerous)
	SetCustomer(ctx context.Context, customer routing.Customer) error

	// UpdateCustomer modifies the givien field for the specified buyer
	UpdateCustomer(ctx context.Context, customerID string, field string, value interface{}) error

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
	// TODO: chopping block (this is dangerous)
	SetBuyer(ctx context.Context, buyer routing.Buyer) error

	// UpdateBuyer modifies the givien field for the specified buyer
	UpdateBuyer(ctx context.Context, buyerID uint64, field string, value interface{}) error

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
	// TODO: chopping block - this is dangerous
	SetSeller(ctx context.Context, seller routing.Seller) error

	// UpdateSeller modifies the givien field for the specified buyer
	UpdateSeller(ctx context.Context, sellerID string, field string, value interface{}) error

	// BuyerIDFromCustomerName returns the buyer ID associated with the given customer name and an error if the customer wasn't found.
	// If the customer has no buyer linked, then it will return a buyer ID of 0 and no error.
	BuyerIDFromCustomerName(ctx context.Context, customerName string) (uint64, error)

	// SellerIDFromCustomerName returns the seller ID associated with the given customer name and an error if the customer wasn't found.
	// If the customer has no seller linked, then it will return an empty seller ID and no error.
	SellerIDFromCustomerName(ctx context.Context, customerName string) (string, error)

	SellerWithCompanyCode(code string) (routing.Seller, error)

	// SetCustomerLink update the customer's buyer and seller references.
	// TODO: chopping block (handled/required by database)
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

	// UpdateRelay updates a single field in a relay record
	UpdateRelay(ctx context.Context, relayID uint64, field string, value interface{}) error

	// SetRelay updates the relay in storage with the provided copy and returns an error if the relay could not be updated.
	// TODO: chopping block (obsoleted by UpdateRelay, and broken anyway)
	SetRelay(ctx context.Context, relay routing.Relay) error

	// Datacenter gets a copy of a datacenter with the specified datacenter ID
	// and returns an empty datacenter and an error if a datacenter with that ID doesn't exist in storage.
	Datacenter(datacenterID uint64) (routing.Datacenter, error)

	// Datacenters returns a copy of all stored datacenters.
	Datacenters() []routing.Datacenter

	// AddDatacenter adds the provided datacenter to storage and returns an error if the datacenter could not be added.
	AddDatacenter(ctx context.Context, datacenter routing.Datacenter) error

	// UpdateDatacenter modifies the givien field for the specified datacenter
	UpdateDatacenter(ctx context.Context, datacenterID uint64, field string, value interface{}) error

	// RemoveDatacenter removes a datacenter with the provided datacenter ID from storage and returns an error if the datacenter could not be removed.
	RemoveDatacenter(ctx context.Context, id uint64) error

	// SetDatacenter updates the datacenter in storage with the provided copy and returns an error if the datacenter could not be updated.
	// TODO: replace with UpdateDatacenter
	SetDatacenter(ctx context.Context, datacenter routing.Datacenter) error

	// GetDatacenterMapsForBuyer returns the list of datacenter aliases in use for a given (internally generated) buyerID. Returns
	// an empty []routing.DatacenterMap if there are no aliases for that buyerID.
	GetDatacenterMapsForBuyer(buyerID uint64) map[uint64]routing.DatacenterMap

	// AddDatacenterMap adds a new datacenter alias for the given buyer and datacenter IDs
	AddDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error

	// UpdateDatacenterMap modifies the given map in storage. The full map is required as the
	// primary key is the buyer ID and the datacenter ID, combined.
	UpdateDatacenterMap(ctx context.Context, buyerID uint64, datacenterID uint64, field string, value interface{}) error

	// ListDatacenterMaps returns a list of alias/buyer mappings for the specified datacenter ID. An
	// empty dcID returns a list of all maps.
	ListDatacenterMaps(dcID uint64) map[uint64]routing.DatacenterMap

	// RemoveDatacenterMap removes an entry from the DatacenterMaps table
	RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error

	// SetRelayMetadata provides write access to ops metadat (mrc, overage, etc)
	// TODO: chopping block (obsoleted by UpdateRelay)
	SetRelayMetadata(ctx context.Context, relay routing.Relay) error

	// CheckSequenceNumber is called in the sync*() operations to see if a sync is required.
	CheckSequenceNumber(ctx context.Context) (bool, int64, error)

	// IncrementSequenceNumber is used by all methods that make changes to the db
	IncrementSequenceNumber(ctx context.Context) error

	// SetSequenceNumber is used to setup the db for unit testing
	SetSequenceNumber(ctx context.Context, value int64) error

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

	// InternalConfig returns the internal config for the given buyer ID
	InternalConfig(buyerID uint64) (core.InternalConfig, error)

	// AddInternalConfig adds the provided InternalConfig to the database
	AddInternalConfig(ctx context.Context, internalConfig core.InternalConfig, buyerID uint64) error

	// UpdateInternalConfig updates the specified field in an InternalConfig record
	UpdateInternalConfig(ctx context.Context, buyerID uint64, field string, value interface{}) error

	// RemoveInternalConfig removes a record from the InternalConfigs table
	RemoveInternalConfig(ctx context.Context, buyerID uint64) error

	// RouteShader returns a slice of route shaders for the given buyer ID
	RouteShader(buyerID uint64) (core.RouteShader, error)

	// AddRouteShader adds the provided RouteShader to the database
	AddRouteShader(ctx context.Context, routeShader core.RouteShader, buyerID uint64) error

	// UpdateRouteShader updates the specified field in an RouteShader record
	UpdateRouteShader(ctx context.Context, buyerID uint64, field string, value interface{}) error

	// RemoveRouteShader removes a record from the RouteShaders table
	RemoveRouteShader(ctx context.Context, buyerID uint64) error

	// AddBannedUser adds a user to the banned_user table
	AddBannedUser(ctx context.Context, buyerID uint64, userID uint64) error

	// RemoveBannedUser removes a user from the banned_user table
	RemoveBannedUser(ctx context.Context, buyerID uint64, userID uint64) error

	// BannedUsers returns the set of banned users for the specified buyer ID. This method
	// is designed to be used by syncRouteShaders() though it can be used by client code.
	BannedUsers(buyerID uint64) (map[uint64]bool, error)

	// GetDatabaseBinFileMetaData returns data from the database_bin_meta table
	GetDatabaseBinFileMetaData() (routing.DatabaseBinFileMetaData, error)

	// UpdateDatabaseBinFileMetaData updates the specified field in an database_bin_meta table
	UpdateDatabaseBinFileMetaData(ctx context.Context, field string, value interface{}) error
}
