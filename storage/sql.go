package storage

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
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

// Customer retrieves a Customer record using the company code
func (db *SQL) Customer(companyCode string) (routing.Customer, error) {

	db.customerMutex.RLock()
	defer db.customerMutex.RUnlock()

	c, found := db.customers[companyCode]
	if !found {
		return routing.Customer{}, &DoesNotExistError{resourceType: "customer", resourceRef: fmt.Sprintf("%s", companyCode)}
	}

	return c, nil
}

// CustomerWithName retrieves a record using the customer's name
func (db *SQL) CustomerWithName(name string) (routing.Customer, error) {
	db.customerMutex.RLock()
	defer db.customerMutex.RUnlock()

	for _, customer := range db.customers {
		if customer.Name == name {
			return customer, nil
		}
	}

	return routing.Customer{}, &DoesNotExistError{resourceType: "customer", resourceRef: name}
}

// Customers retrieves the full list
func (db *SQL) Customers() []routing.Customer {
	var customers []routing.Customer
	for _, customer := range db.customers {
		customers = append(customers, customer)
	}

	sort.Slice(customers, func(i int, j int) bool { return customers[i].Name < customers[j].Name })
	return customers
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
	db.buyerMutex.RLock()
	defer db.buyerMutex.RUnlock()

	b, found := db.buyers[id]
	if !found {
		return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%x", id)}
	}

	return b, nil
}

// BuyerWithCompanyCode gets the Buyer with the matching company code
func (db *SQL) BuyerWithCompanyCode(code string) (routing.Buyer, error) {
	db.buyerMutex.RLock()
	defer db.buyerMutex.RUnlock()

	for _, buyer := range db.buyers {
		if buyer.CompanyCode == code {
			return buyer, nil
		}
	}
	return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: code}
}

// Buyers returns a copy of all stored buyers.
func (db *SQL) Buyers() []routing.Buyer {
	db.buyerMutex.RLock()
	defer db.buyerMutex.RUnlock()

	var buyers []routing.Buyer
	for _, buyer := range db.buyers {
		buyers = append(buyers, buyer)
	}

	sort.Slice(buyers, func(i int, j int) bool { return buyers[i].ID < buyers[j].ID })
	return buyers
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
	db.sellerMutex.RLock()
	defer db.sellerMutex.RUnlock()

	s, found := db.sellers[id]
	if !found {
		return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: id}
	}

	return s, nil
}

// Sellers returns a copy of all stored sellers.
func (db *SQL) Sellers() []routing.Seller {
	db.sellerMutex.RLock()
	defer db.sellerMutex.RUnlock()

	var sellers []routing.Seller
	for _, seller := range db.sellers {
		sellers = append(sellers, seller)
	}

	sort.Slice(sellers, func(i int, j int) bool { return sellers[i].ID < sellers[j].ID })
	return sellers
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
	db.sellerMutex.RLock()
	defer db.sellerMutex.RUnlock()

	for _, seller := range db.sellers {
		if seller.CompanyCode == code {
			return seller, nil
		}
	}
	return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: code}
}

// SetCustomerLink update the customer's buyer and seller references.
func (db *SQL) SetCustomerLink(ctx context.Context, customerName string, buyerID uint64, sellerID string) error {
	return nil
}

// Relay gets a copy of a relay with the specified relay ID
// and returns an empty relay and an error if a relay with that ID doesn't exist in storage.
func (db *SQL) Relay(id uint64) (routing.Relay, error) {
	db.relayMutex.RLock()
	defer db.relayMutex.RUnlock()

	relay, found := db.relays[id]
	if !found {
		return routing.Relay{}, &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%x", id)}
	}

	return relay, nil
}

// Relays returns a copy of all stored relays.
func (db *SQL) Relays() []routing.Relay {
	db.relayMutex.RLock()
	defer db.relayMutex.RUnlock()

	var relays []routing.Relay
	for _, relay := range db.relays {
		relays = append(relays, relay)
	}

	sort.Slice(relays, func(i int, j int) bool { return relays[i].ID < relays[j].ID })
	return relays
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
	db.datacenterMutex.RLock()
	defer db.datacenterMutex.RUnlock()

	d, found := db.datacenters[datacenterID]
	if !found {
		return routing.Datacenter{}, &DoesNotExistError{resourceType: "datacenter", resourceRef: datacenterID}
	}

	return d, nil
}

// Datacenters returns a copy of all stored datacenters.
func (db *SQL) Datacenters() []routing.Datacenter {
	db.datacenterMutex.RLock()
	defer db.datacenterMutex.RUnlock()

	var datacenters []routing.Datacenter
	for _, datacenter := range db.datacenters {
		datacenters = append(datacenters, datacenter)
	}

	sort.Slice(datacenters, func(i int, j int) bool { return datacenters[i].ID < datacenters[j].ID })
	return datacenters
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
