package storage

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/crypto"
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

	SyncSequenceNumber int64

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

type sqlCustomer struct {
	Code                   string
	Name                   string
	AutomaticSignInDomains string
	Active                 bool
	CustomerName           string
	CustomerCode           string
}

func (db *SQL) AddCustomer(ctx context.Context, c routing.Customer) error {
	var sql bytes.Buffer

	db.customerMutex.RLock()
	_, ok := db.customers[c.Code]
	db.customerMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "customer", resourceRef: c.Code}
	}

	customer := sqlCustomer{
		CustomerCode:           c.Code,
		CustomerName:           c.Name,
		AutomaticSignInDomains: c.AutomaticSignInDomains,
		Active:                 false,
	}

	// Add the buyer in remote storage
	sql.Write([]byte("insert into customers ("))
	sql.Write([]byte("active, automatic_signin_domain, customer_name, customer_code"))
	sql.Write([]byte(") values ($1, $2, $3, $4)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error perparing AddCustomer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(customer.Active,
		customer.AutomaticSignInDomains,
		customer.CustomerName,
		customer.CustomerCode,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding datacenter", "err", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		level.Error(db.Logger).Log("during", "RowsAffected returned an error", "err", err)
		return err
	}
	if rows != 1 {
		level.Error(db.Logger).Log("during", "RowsAffected <> 1", "err", err)
		return err
	}

	// Add the buyer in cached storage
	db.customerMutex.Lock()
	db.customers[c.Code] = c
	db.customerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

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

type sqlSeller struct {
	IngressPriceNibblinsPerGB int64 `firestore:"pricePublicIngressNibblins"`
	EgressPriceNibblinsPerGB  int64 `firestore:"pricePublicEgressNibblins"`
}

// The seller_id is reqired by the schema. The client interface must already have a
// seller defined.
func (db *SQL) AddSeller(ctx context.Context, s routing.Seller) error {
	var sql bytes.Buffer
	// Check if the seller exists
	db.sellerMutex.RLock()
	_, found := db.sellers[s.ID]
	db.sellerMutex.RUnlock()

	if found {
		return &AlreadyExistsError{resourceType: "seller", resourceRef: s.ID}
	}

	var company routing.Customer
	for _, customer := range db.customers {
		if customer.Code == s.CompanyCode {
			company = customer
		}
	}

	// A relevant customer entry must exist to add a seller
	if company.Code == "" {
		return &DoesNotExistError{resourceType: "customer", resourceRef: s.CompanyCode}
	}

	newSellerData := sqlSeller{
		IngressPriceNibblinsPerGB: int64(s.IngressPriceNibblinsPerGB),
		EgressPriceNibblinsPerGB:  int64(s.EgressPriceNibblinsPerGB),
	}

	// Add the seller in remote storage
	sql.Write([]byte("insert into sellers ("))
	sql.Write([]byte("public_egress_price, public_ingress_price, customer_id"))
	sql.Write([]byte(") values ($1, $2, $3)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error perparing AddSeller SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(newSellerData.EgressPriceNibblinsPerGB,
		newSellerData.IngressPriceNibblinsPerGB,
		s.Name,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding datacenter", "err", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		level.Error(db.Logger).Log("during", "RowsAffected returned an error", "err", err)
		return err
	}
	if rows != 1 {
		level.Error(db.Logger).Log("during", "RowsAffected <> 1", "err", err)
		return err
	}

	// Add the seller in cached storage
	db.sellerMutex.Lock()
	db.sellers[s.ID] = s
	db.sellerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

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

// CheckSequenceNumber is called in the sync*() operations to see if a sync is required.
// Returns:
// 	true: if the remote sequence number > local sequence number
//	false: if the remote sequence number = local sequence number
//  false + error: if the remote sequence number < local sequence number
//
// "true" forces the caller to sync from the database and updates
// the local sequence number. "false" does not force a sync and does
// not modify the local number.
func (db *SQL) CheckSequenceNumber(ctx context.Context) (bool, error) {
	var sequenceNumber int64

	err := db.Client.QueryRowContext(ctx, "select sync_sequence_number from metadata").Scan(&sequenceNumber)

	if err == sql.ErrNoRows {
		level.Error(db.Logger).Log("during", "No sequence number returned", "err", err)
		return false, err
	} else if err != nil {
		level.Error(db.Logger).Log("during", "query error", "err", err)
		return false, err
	}

	db.sequenceNumberMutex.RLock()
	localSeqNum := db.SyncSequenceNumber
	db.sequenceNumberMutex.RUnlock()

	if localSeqNum < sequenceNumber {
		db.sequenceNumberMutex.Lock()
		db.SyncSequenceNumber = sequenceNumber
		db.sequenceNumberMutex.Unlock()
		return true, nil
	} else if localSeqNum == sequenceNumber {
		return false, nil
	} else {
		err = fmt.Errorf("local sequence number larger than remote: %d > %d", localSeqNum, sequenceNumber)
		level.Error(db.Logger).Log("during", "query error", "err", err)
		return false, err
	}
}

// IncrementSequenceNumber is called by all CRUD operations defined in the Storage interface. It only
// increments the remote seq number. When the sync() functions call CheckSequenceNumber(), if the
// local and remote numbers are not the same, the data will be sync'd from the database and the local
// sequence numbers updated.
func (db *SQL) IncrementSequenceNumber(ctx context.Context) error {

	var sequenceNumber int64

	err := db.Client.QueryRowContext(ctx, "select sync_sequence_number from metadata").Scan(&sequenceNumber)

	if err == sql.ErrNoRows {
		level.Error(db.Logger).Log("during", "No sequence number returned", "err", err)
		return err
	} else if err != nil {
		level.Error(db.Logger).Log("during", "query error", "err", err)
		return err
	}

	sequenceNumber++

	stmt, err := db.Client.PrepareContext(ctx, "update metadata set sync_sequence_number = $1")
	if err != nil {
		level.Error(db.Logger).Log("during", "error perparing SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(sequenceNumber)
	if err != nil {
		level.Error(db.Logger).Log("during", "error setting sequence number", "err", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		level.Error(db.Logger).Log("during", "RowsAffected returned an error", "err", err)
	}
	if rows != 1 {
		level.Error(db.Logger).Log("during", "RowsAffected <> 1", "err", err)
	}

	return nil
}

// SyncLoop is a helper method that calls Sync
func (db *SQL) SyncLoop(ctx context.Context, c <-chan time.Time) {
	if err := db.Sync(ctx); err != nil {
		level.Error(db.Logger).Log("during", "SyncLoop", "err", err)
	}

	for {
		select {
		case <-c:
			if err := db.Sync(ctx); err != nil {
				level.Error(db.Logger).Log("during", "SyncLoop", "err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// Sync is a utility function that calls the individual sync* methods
func (db *SQL) Sync(ctx context.Context) error {

	seqNumberNotInSync, err := db.CheckSequenceNumber(ctx)
	if err != nil {
		return err
	}
	if !seqNumberNotInSync {
		return nil
	}

	var outerErr error
	var wg sync.WaitGroup
	wg.Add(6)

	go func() {

		if err := db.syncRelays(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync relays: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := db.syncCustomers(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync customers: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := db.syncBuyers(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync buyers: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := db.syncSellers(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync sellers: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := db.syncDatacenters(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync datacenters: %v", err)
		}
		wg.Done()
	}()

	go func() {
		if err := db.syncDatacenterMaps(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync datacenterMaps: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()

	return outerErr
}

type sqlDatacenter struct {
	ID            int64
	Name          string
	Enabled       bool
	Latitude      float64
	Longitude     float64
	SupplierName  string
	StreetAddress string
	SellerID      int64
}

func (db *SQL) syncDatacenters(ctx context.Context) error {

	var sql bytes.Buffer
	var dc sqlDatacenter
	datacenters := make(map[uint64]routing.Datacenter)

	sql.Write([]byte("select id, display_name, enabled, latitude, longitude,"))
	sql.Write([]byte("supplier_name, street_address, seller_id from datacenters"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		fmt.Printf("QueryContext returned an error: %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&dc.ID, // TODO: Add PK to routing.Datacenter
			&dc.Name,
			&dc.Enabled,
			&dc.Latitude,
			&dc.Longitude,
			&dc.SupplierName,
			&dc.StreetAddress,
			&dc.SellerID,
		)

		did := crypto.HashID(dc.Name)
		datacenters[did] = routing.Datacenter{
			ID:      did,
			Name:    dc.Name,
			Enabled: dc.Enabled,
			Location: routing.Location{
				Latitude:  dc.Latitude,
				Longitude: dc.Longitude,
			},
			SupplierName: dc.SupplierName,
			SellerID:     dc.SellerID,
		}
	}

	db.datacenterMutex.Lock()
	db.datacenters = datacenters
	db.datacenterMutex.Unlock()

	level.Info(db.Logger).Log("during", "syncDatacenters", "num", len(db.datacenters))

	return nil
}

// AddDatacenter adds the provided datacenter to storage. It enforces business rule
// that datacenter names are of the form: "seller"."location"."optional number/enum"
//
// The seller_id is reqired by the schema. The client interface must already have a
// seller defined.
func (db *SQL) AddDatacenter(ctx context.Context, datacenter routing.Datacenter) error {

	var sql bytes.Buffer

	// TODO make sure it doesn't already exist

	dc := sqlDatacenter{
		Name:         strings.Split(datacenter.Name, ".")[0],
		Enabled:      datacenter.Enabled,
		Latitude:     datacenter.Location.Latitude,
		Longitude:    datacenter.Location.Longitude,
		SupplierName: datacenter.SupplierName,
		SellerID:     datacenter.SellerID,
	}

	// Add the datacenter in remote storage
	sql.Write([]byte("insert into datacenters ("))
	sql.Write([]byte("display_name, enabled, latitude, longitude, supplier_name, street_address, "))
	sql.Write([]byte("seller_id ) values ($1, $2, $3, $4, $5, $6, $7)"))

	// sql.Write([]byte(" )")))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error perparing AddDatacenter SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(dc.Name,
		dc.Enabled,
		dc.Latitude,
		dc.Longitude,
		dc.SupplierName,
		dc.StreetAddress,
		dc.SellerID,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding datacenter", "err", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		level.Error(db.Logger).Log("during", "RowsAffected returned an error", "err", err)
		return err
	}
	if rows != 1 {
		level.Error(db.Logger).Log("during", "RowsAffected <> 1", "err", err)
		return err
	}

	// Add the datacenter in cached storage
	db.datacenterMutex.Lock()
	db.datacenters[datacenter.ID] = datacenter
	db.datacenterMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

func (db *SQL) syncRelays(ctx context.Context) error {
	return nil
}
func (db *SQL) syncBuyers(ctx context.Context) error {
	return nil
}
func (db *SQL) syncSellers(ctx context.Context) error {
	return nil
}
func (db *SQL) syncDatacenterMaps(ctx context.Context) error {
	return nil
}
func (db *SQL) syncCustomers(ctx context.Context) error {
	return nil
}
