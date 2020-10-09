package storage

import (
	"context"
	"database/sql"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/routing"
)

// SQL can be backed up by PostgreSQL or SQLite3
type SQL struct {
	Client *sql.DB
	Logger log.Logger

	datacenters    map[uint64]routing.Datacenter
	relays         map[uint64]routing.Relay
	customers      map[string]routing.Customer
	buyers         map[uint64]routing.Buyer
	sellers        map[string]routing.Seller
	datacenterMaps map[uint64]routing.DatacenterMap

	syncSequenceNumber int64

	datacenterMutex     sync.RWMutex
	relayMutex          sync.RWMutex
	customerMutex       sync.RWMutex
	buyerMutex          sync.RWMutex
	sellerMutex         sync.RWMutex
	datacenterMapMutex  sync.RWMutex
	sequenceNumberMutex sync.RWMutex
}

func (db *SQL) Customer(code string) (routing.Customer, error) {

	return routing.Customer{}, nil
}

func (db *SQL) CustomerWithName(name string) (routing.Customer, error) {

	return routing.Customer{}, nil
}

func (db *SQL) Customers() []routing.Customer {
	return []routing.Customer{}
}

func (db *SQL) AddCustomer(ctx context.Context, customer routing.Customer) error {
	return nil
}

func (db *SQL) RemoveCustomer(ctx context.Context, code string) error {
	return nil
}

func (db *SQL) SetCustomer(ctx context.Context, customer routing.Customer) error {
	return nil
}

// Buyer gets a copy of a buyer with the specified buyer ID,
// and returns an empty buyer and an error if a buyer with that ID doesn't exist in storage.
func (db *SQL) Buyer(id uint64) (routing.Buyer, error) {
	return routing.Buyer{}, nil
}

// BuyerWithCompanyCode gets the Buyer with the matching company code
func (db *SQL) BuyerWithCompanyCode(code string) (routing.Buyer, error) {
	return routing.Buyer{}, nil
}

// Buyers returns a copy of all stored buyers.
func (db *SQL) Buyers() []routing.Buyer {
	return []routing.Buyer{}
}

// AddBuyer adds the provided buyer to storage and returns an error if the buyer could not be added.
func (db *SQL) AddBuyer(ctx context.Context, buyer routing.Buyer) error {
	return nil
}

// RemoveBuyer removes a buyer with the provided buyer ID from storage and returns an error if the buyer could not be removed.
func (db *SQL) RemoveBuyer(ctx context.Context, id uint64) error {
	return nil
}

// SetBuyer updates the buyer in storage with the provided copy and returns an error if the buyer could not be updated.
func (db *SQL) SetBuyer(ctx context.Context, buyer routing.Buyer) error {
	return nil
}

// Seller gets a copy of a seller with the specified seller ID,
// and returns an empty seller and an error if a seller with that ID doesn't exist in storage.
func (db *SQL) Seller(id string) (routing.Seller, error) {
	return routing.Seller{}, nil
}

// Sellers returns a copy of all stored sellers.
func (db *SQL) Sellers() []routing.Seller {
	return []routing.Seller{}
}

// AddSeller adds the provided seller to storage and returns an error if the seller could not be added.
func (db *SQL) AddSeller(ctx context.Context, seller routing.Seller) error {
	return nil
}

// RemoveSeller removes a seller with the provided seller ID from storage and returns an error if the seller could not be removed.
func (db *SQL) RemoveSeller(ctx context.Context, id string) error {
	return nil
}

// SetSeller updates the seller in storage with the provided copy and returns an error if the seller could not be updated.
func (db *SQL) SetSeller(ctx context.Context, seller routing.Seller) error {
	return nil
}

// BuyerIDFromCustomerName returns the buyer ID associated with the given customer name and an error if the customer wasn't found.
// If the customer has no buyer linked, then it will return a buyer ID of 0 and no error.
func (db *SQL) BuyerIDFromCustomerName(ctx context.Context, customerName string) (uint64, error) {
	return 0, nil
}

// SellerIDFromCustomerName returns the seller ID associated with the given customer name and an error if the customer wasn't found.
// If the customer has no seller linked, then it will return an empty seller ID and no error.
func (db *SQL) SellerIDFromCustomerName(ctx context.Context, customerName string) (string, error) {
	return "", nil
}

func (db *SQL) SellerWithCompanyCode(code string) (routing.Seller, error) {
	return routing.Seller{}, nil
}

// SetCustomerLink update the customer's buyer and seller references.
func (db *SQL) SetCustomerLink(ctx context.Context, customerName string, buyerID uint64, sellerID string) error {
	return nil
}

// Relay gets a copy of a relay with the specified relay ID
// and returns an empty relay and an error if a relay with that ID doesn't exist in storage.
func (db *SQL) Relay(id uint64) (routing.Relay, error) {
	return routing.Relay{}, nil
}

// Relays returns a copy of all stored relays.
func (db *SQL) Relays() []routing.Relay {
	return []routing.Relay{}
}

// AddRelay adds the provided relay to storage and returns an error if the relay could not be added.
func (db *SQL) AddRelay(ctx context.Context, relay routing.Relay) error {
	return nil
}

// RemoveRelay removes a relay with the provided relay ID from storage and returns an error if the relay could not be removed.
func (db *SQL) RemoveRelay(ctx context.Context, id uint64) error {
	return nil
}

// SetRelay updates the relay in storage with the provided copy and returns an error if the relay could not be updated.
func (db *SQL) SetRelay(ctx context.Context, relay routing.Relay) error {
	return nil
}

// Datacenter gets a copy of a datacenter with the specified datacenter ID
// and returns an empty datacenter and an error if a datacenter with that ID doesn't exist in storage.
func (db *SQL) Datacenter(datacenterID uint64) (routing.Datacenter, error) {
	return routing.Datacenter{}, nil
}

// Datacenters returns a copy of all stored datacenters.
func (db *SQL) Datacenters() []routing.Datacenter {
	return []routing.Datacenter{}
}

// AddDatacenter adds the provided datacenter to storage and returns an error if the datacenter could not be added.
func (db *SQL) AddDatacenter(ctx context.Context, datacenter routing.Datacenter) error {
	return nil
}

// RemoveDatacenter removes a datacenter with the provided datacenter ID from storage and returns an error if the datacenter could not be removed.
func (db *SQL) RemoveDatacenter(ctx context.Context, id uint64) error {
	return nil
}

// SetDatacenter updates the datacenter in storage with the provided copy and returns an error if the datacenter could not be updated.
func (db *SQL) SetDatacenter(ctx context.Context, datacenter routing.Datacenter) error {
	return nil
}

// GetDatacenterMapsForBuyer returns the list of datacenter aliases in use for a given (internally generated) buyerID. Returns
// an empty []routing.DatacenterMap if there are no aliases for that buyerID.
func (db *SQL) GetDatacenterMapsForBuyer(buyerID uint64) map[uint64]routing.DatacenterMap {
	return map[uint64]routing.DatacenterMap{}
}

// AddDatacenterMap adds a new datacenter alias for the given buyer and datacenter IDs
func (db *SQL) AddDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {
	return nil
}

// ListDatacenterMaps returns a list of alias/buyer mappings for the specified datacenter ID. An
// empty dcID returns a list of all maps.
func (db *SQL) ListDatacenterMaps(dcID uint64) map[uint64]routing.DatacenterMap {
	return map[uint64]routing.DatacenterMap{}
}

// RemoveDatacenterMap removes an entry from the DatacenterMaps table
func (db *SQL) RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {
	return nil
}

// SetRelayMetadata provides write access to ops metadat (mrc, overage, etc)
func (db *SQL) SetRelayMetadata(ctx context.Context, relay routing.Relay) error {
	return nil
}
