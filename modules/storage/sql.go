package storage

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
)

// SQL is an implementation of the Storer interface. It can
// be backed up by PostgreSQL or SQLite3. There is a
// dichotomy set in concrete, here. Using the routing.Relay
// type as an example:
//
// 	Relay.ID     : Internal-use, calculated from public IP:Port
//  Relay.RelayID: Database primary key for the relays table
//
// The PK is required to enforce business rules in the DB. The *IDs
// maps below provide an easy look-up mechanism for the PK. This also
// enforces the "unique" nature of the internally generated IDs (map
// keys must be unique).
type SQL struct {
	Client *sql.DB
	Logger log.Logger

	datacenters    map[uint64]routing.Datacenter
	relays         map[uint64]routing.Relay
	customers      map[string]routing.Customer
	buyers         map[uint64]routing.Buyer
	sellers        map[string]routing.Seller
	datacenterMaps map[uint64]routing.DatacenterMap

	internalConfigs map[uint64]core.InternalConfig // index: buyer ID
	routeShaders    map[uint64][]core.RouteShader  // index: buyer ID

	datacenterMutex     sync.RWMutex
	relayMutex          sync.RWMutex
	customerMutex       sync.RWMutex
	buyerMutex          sync.RWMutex
	sellerMutex         sync.RWMutex
	datacenterMapMutex  sync.RWMutex
	sequenceNumberMutex sync.RWMutex
	internalConfigMutex sync.RWMutex
	routeShaderMutex    sync.RWMutex

	datacenterIDs map[int64]uint64
	relayIDs      map[int64]uint64
	customerIDs   map[int64]string
	buyerIDs      map[int64]uint64
	sellerIDs     map[int64]string

	datacenterIDsMutex sync.RWMutex
	relayIDsMutex      sync.RWMutex
	customerIDsMutex   sync.RWMutex
	buyerIDsMutex      sync.RWMutex
	sellerIDsMutex     sync.RWMutex

	SyncSequenceNumber int64
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
	ID                     int64
	Name                   string
	AutomaticSignInDomains string
	Active                 bool
	Debug                  bool
	CustomerCode           string
	DatabaseID             int64
	BuyerID                sql.NullInt64 // loaded during syncCustomers()
	SellerID               sql.NullInt64 // loaded during syncCustomers()
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
		Name:                   c.Name,
		AutomaticSignInDomains: c.AutomaticSignInDomains,
		Active:                 c.Active,
	}

	// Add the buyer in remote storage
	sql.Write([]byte("insert into customers ("))
	sql.Write([]byte("active, automatic_signin_domain, customer_name, customer_code"))
	sql.Write([]byte(") values ($1, $2, $3, $4)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddCustomer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(customer.Active,
		customer.AutomaticSignInDomains,
		customer.Name,
		customer.CustomerCode,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding customer", "err", err)
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

	db.syncCustomers(ctx)

	db.IncrementSequenceNumber(ctx)

	return nil
}

// RemoveCustomer removes a customer from the database. An error is returned if
//  1. The customer ID does not exist
//  2. Removing the customer will break a foreign key relationship (buyer, seller)
//  3. Any other error returned from the database
//
// #2 is not checked here - it is enforced by the database
func (db *SQL) RemoveCustomer(ctx context.Context, customerCode string) error {

	var sql bytes.Buffer

	db.customerMutex.RLock()
	customer, ok := db.customers[customerCode]
	db.customerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "customer", resourceRef: fmt.Sprintf("%s", customerCode)}
	}

	sql.Write([]byte("delete from customers where id = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveCustomer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(customer.DatabaseID)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing customer", "err", err)
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

	db.customerMutex.Lock()
	delete(db.customers, customerCode)
	db.customerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

// SetCustomer modifies a subset of fields in a customer record
// in the database. Modifield fields:
//		Name
//		AutomaticSigninDomains
//		Active
//		Debug
func (db *SQL) SetCustomer(ctx context.Context, c routing.Customer) error {

	var sql bytes.Buffer

	db.customerMutex.RLock()
	_, ok := db.customers[c.Code]
	db.customerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "customer", resourceRef: fmt.Sprintf("%s", c.Code)}
	}

	sql.Write([]byte("update customers set (active, debug, automatic_signin_domain, customer_name) ="))
	sql.Write([]byte("($1, $2, $3, $4) where id = $5"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing SetCustomer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(c.Active, c.Debug, c.AutomaticSignInDomains, c.Name, c.DatabaseID)
	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying customer record", "err", err)
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

	db.customerMutex.Lock()
	db.customers[c.Code] = c
	db.customerMutex.Unlock()

	return nil
}

// Buyer gets a copy of a buyer with the specified buyer ID,
// and returns an empty buyer and an error if a buyer with that ID doesn't exist in storage.
func (db *SQL) Buyer(id uint64) (routing.Buyer, error) {
	db.buyerMutex.RLock()
	b, found := db.buyers[id]
	db.buyerMutex.RUnlock()

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

	var buyers []routing.Buyer
	for _, buyer := range db.buyers {
		buyers = append(buyers, buyer)
	}

	db.buyerMutex.RUnlock()

	sort.Slice(buyers, func(i int, j int) bool { return buyers[i].ID < buyers[j].ID })
	return buyers
}

// AddBuyer adds the provided buyer to storage and returns an error if the buyer could not be added.
func (db *SQL) AddBuyer(ctx context.Context, b routing.Buyer) error {
	var sql bytes.Buffer

	db.buyerMutex.RLock()
	_, ok := db.buyers[b.ID]
	db.buyerMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "buyer", resourceRef: b.ID}
	}

	// This check only pertains to the next tool. Stateful clients would already
	// have the customer id.
	c, err := db.Customer(b.CompanyCode)
	if err != nil {
		return &DoesNotExistError{resourceType: "customer", resourceRef: b.CompanyCode}
	}

	internalID := binary.LittleEndian.Uint64(b.PublicKey[:8])

	buyer := sqlBuyer{
		ID:             internalID,
		CompanyCode:    b.CompanyCode,
		ShortName:      b.CompanyCode,
		IsLiveCustomer: b.Live,
		Debug:          b.Debug,
		PublicKey:      b.PublicKey,
		CustomerID:     c.DatabaseID,
	}

	// Add the buyer in remote storage
	sql.Write([]byte("insert into buyers ("))
	sql.Write([]byte("short_name, is_live_customer, debug, public_key, customer_id"))
	sql.Write([]byte(") values ($1, $2, $3, $4, $5)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddBuyer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
		buyer.ShortName,
		buyer.IsLiveCustomer,
		buyer.Debug,
		buyer.PublicKey,
		buyer.CustomerID,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding buyer", "err", err)
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

	db.syncBuyers(ctx)

	db.customerMutex.Lock()
	db.customers[c.Code] = c
	db.customerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

// RemoveBuyer removes a buyer with the provided buyer ID from storage
// and returns an error if:
//  1. The buyer ID does not exist
//  2. Removing the buyer would break the foreigh key relationship (datacenter_maps)
//  3. Any other error returned from the database
func (db *SQL) RemoveBuyer(ctx context.Context, id uint64) error {
	var sql bytes.Buffer

	db.buyerMutex.RLock()
	buyer, ok := db.buyers[id]
	db.buyerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%016x", id)}
	}

	sql.Write([]byte("delete from buyers where id = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveBuyer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(buyer.DatabaseID)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing buyer", "err", err)
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

	db.buyerMutex.Lock()
	delete(db.buyers, buyer.ID)
	db.buyerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

// SetBuyer updates a subset of the fields in the buyers table and
// updates the local copy.
//		Live
//		Debug
//		PublicKey
func (db *SQL) SetBuyer(ctx context.Context, b routing.Buyer) error {

	var sql bytes.Buffer

	db.buyerMutex.RLock()
	_, ok := db.buyers[b.ID]
	db.buyerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%016x", b.ID)}
	}

	sql.Write([]byte("update buyers set (is_live_customer, debug, public_key) = ($1, $2, $3) where id = $4 "))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing SetBuyer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(b.Live, b.Debug, b.PublicKey, b.DatabaseID)
	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying buyer record", "err", err)
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

	db.buyerMutex.Lock()
	db.buyers[b.ID] = b
	db.buyerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

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
	ID                        string
	ShortName                 string
	IngressPriceNibblinsPerGB int64
	EgressPriceNibblinsPerGB  int64
	CustomerID                int64
	DatabaseID                int64
}

// The seller_id is required by the schema. The client interface must already have a
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

	// This check only pertains to the next tool. Stateful clients would already
	// have the customer id.
	c, err := db.Customer(s.CompanyCode)
	if err != nil {
		return &DoesNotExistError{resourceType: "customer", resourceRef: s.CompanyCode}
	}

	newSellerData := sqlSeller{
		ID:                        s.ID,
		ShortName:                 s.ShortName,
		IngressPriceNibblinsPerGB: int64(s.IngressPriceNibblinsPerGB),
		EgressPriceNibblinsPerGB:  int64(s.EgressPriceNibblinsPerGB),
		CustomerID:                c.DatabaseID,
	}

	// Add the seller in remote storage
	sql.Write([]byte("insert into sellers ("))
	sql.Write([]byte("short_name, public_egress_price, public_ingress_price, customer_id"))
	sql.Write([]byte(") values ($1, $2, $3, $4)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddSeller SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(newSellerData.ShortName, newSellerData.EgressPriceNibblinsPerGB,
		newSellerData.IngressPriceNibblinsPerGB,
		newSellerData.CustomerID,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding seller", "err", err)
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

	// Must re-sync to get the relevant SQL IDs
	db.syncSellers(ctx)

	db.syncCustomers(ctx) // pick up new seller ID

	db.IncrementSequenceNumber(ctx)

	return nil
}

// RemoveSeller removes a seller with the provided seller ID from storage and
// returns an error if:
//  1. The seller ID does not exist
//  2. Removing the seller would break the foreigh key relationship (datacenters)
//  3. Any other error returned from the database
func (db *SQL) RemoveSeller(ctx context.Context, id string) error {
	var sql bytes.Buffer

	db.sellerMutex.RLock()
	seller, ok := db.sellers[id]
	db.sellerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "seller", resourceRef: fmt.Sprintf("%s", id)}
	}

	sql.Write([]byte("delete from sellers where id = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveBuyer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(seller.DatabaseID)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing seller", "err", err)
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

	db.sellerMutex.Lock()
	delete(db.sellers, seller.ID)
	db.sellerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

// SetSeller updates a subset of the sellers table entry:
//		Name		(not yet implemented, derived from parent customer)
//		CompanyCode (not yet implemented, awaiting business rule decision)
//		IngressPriceNibblinsPerGB
//  	EgressPriceNibblinsPerGB
func (db *SQL) SetSeller(ctx context.Context, seller routing.Seller) error {

	var sql bytes.Buffer

	db.sellerMutex.RLock()
	_, ok := db.sellers[seller.ID]
	db.sellerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "seller", resourceRef: fmt.Sprintf("%s", seller.ID)}
	}

	sql.Write([]byte("update sellers set (public_egress_price, public_ingress_price) = ($1, $2) where id = $3 "))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing SetBuyer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(seller.EgressPriceNibblinsPerGB, seller.IngressPriceNibblinsPerGB, seller.DatabaseID)
	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying seller record", "err", err)
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

	db.sellerMutex.Lock()
	db.sellers[seller.ID] = seller
	db.sellerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

// BuyerIDFromCustomerName is called by the SetCustomerLink endpoint, which is deprecated.
func (db *SQL) BuyerIDFromCustomerName(ctx context.Context, customerName string) (uint64, error) {
	return 0, fmt.Errorf("BuyerIDFromCustomerName() not implemented in SQL Storer")

}

// SellerIDFromCustomerName is called by the SetCustomerLink endpoint, which is deprecated.
func (db *SQL) SellerIDFromCustomerName(ctx context.Context, customerName string) (string, error) {
	return "", fmt.Errorf("BuyerIDFromCustomerName() not implemented in SQL Storer")
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
// The Customer/Buyer/Seller relationship is controlled by primary and
// foreign keys and can not be modified by any client. Also, the relevant
// fields (BuyerRef and SellerRef) Are being dropped from the Customer type.
func (db *SQL) SetCustomerLink(ctx context.Context, customerName string, buyerID uint64, sellerID string) error {
	return fmt.Errorf("SetCustomerLink() not implemented in SQL Storer")
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

// UpdateRelay updates one field (2 for addr) in a relay record - field names
// are those provided by routing.Relay.
// value:
//	addr           : ipaddress:port (string)
//  bw_billing_rule: float64 (json number)
//  machine_type   : float64 (json number)
//  relay_state    : float64 (json number)
//  MRC            : USD float64 (json number)
//  Overage        : USD float64 (json number)
//  StartDate      : string ('January 2, 2006')
//  EndDate        : string ('January 2, 2006')
//  all others are bool, float64 or string, based on field type
func (db *SQL) UpdateRelay(ctx context.Context, relayID uint64, field string, value interface{}) error {

	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt

	relay, err := db.Relay(relayID)
	if err != nil {
		return &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%016x", relayID)}
	}

	switch field {
	case "Name":
		name, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update relays set display_name=$1 where id=$2"))
		args = append(args, name, relay.DatabaseID)
		relay.Name = name

	case "Addr":
		addrString, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		uriTuple := strings.Split(addrString, ":")
		if uriTuple[0] == "" || uriTuple[1] == "" {
			return fmt.Errorf("Unable to parse URI fo Add field: %v", value)
		}
		updateSQL.Write([]byte("update relays set (public_ip, public_ip_port) = ($1, $2) "))
		updateSQL.Write([]byte("where id=$3"))
		args = append(args, uriTuple[0], uriTuple[1], relay.DatabaseID)

		addr, err := net.ResolveUDPAddr("udp", addrString)
		if err != nil {
			return fmt.Errorf("Error converting relay address %s: %v", addrString, err)
		}
		relay.Addr = *addr

	case "InternalAddr":
		addrString, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		uriTuple := strings.Split(addrString, ":")
		if uriTuple[0] == "" || uriTuple[1] == "" {
			return fmt.Errorf("Unable to parse URI fo InternalAddr field: %v", value)
		}
		updateSQL.Write([]byte("update relays set (internal_ip, internal_ip_port) = ($1, $2) "))
		updateSQL.Write([]byte("where id=$3"))
		args = append(args, uriTuple[0], uriTuple[1], relay.DatabaseID)

		addr, err := net.ResolveUDPAddr("udp", addrString)
		if err != nil {
			return fmt.Errorf("Error converting relay address %s: %v", addrString, err)
		}
		relay.Addr = *addr

	case "PublicKey":
		publicKey, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("%v is not a valid []byte type", value)
		}
		updateSQL.Write([]byte("update relays set public_key=$1 where id=$2"))
		args = append(args, publicKey, relay.DatabaseID)
		relay.PublicKey = publicKey

	case "NICSpeedMbps":
		portSpeed, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		updateSQL.Write([]byte("update relays set port_speed=$1 where id=$2"))
		args = append(args, portSpeed, relay.DatabaseID)
		relay.NICSpeedMbps = int32(portSpeed)

	case "IncludedBandwidthGB":
		includedBW, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		updateSQL.Write([]byte("update relays set included_bandwidth_gb=$1 where id=$2"))
		args = append(args, includedBW, relay.DatabaseID)
		relay.IncludedBandwidthGB = int32(includedBW)

	case "State":
		state, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		if state < 0 || state > 5 {
			return fmt.Errorf("%d is not a valid BandWidthRule value", int64(state))
		}
		updateSQL.Write([]byte("update relays set relay_state=$1 where id=$2"))
		args = append(args, int64(state), relay.DatabaseID)
		// already checked int validity above
		relay.State, _ = routing.GetRelayStateSQL(int64(state))

	case "ManagementAddr":
		// routing.Relay.ManagementIP is currently a string type although
		// the database field is inet
		managementIP, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update relays set management_ip=$1 where id=$2"))
		args = append(args, managementIP, relay.DatabaseID)
		relay.ManagementAddr = managementIP

	case "SSHUser":
		user, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string type", value)
		}
		updateSQL.Write([]byte("update relays set ssh_user=$1 where id=$2"))
		args = append(args, user, relay.DatabaseID)
		relay.SSHUser = user

	case "SSHPort":
		port, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		updateSQL.Write([]byte("update relays set ssh_port=$1 where id=$2"))
		args = append(args, port, relay.DatabaseID)
		relay.SSHPort = int64(port)

	case "MaxSessions":
		maxSessions, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		updateSQL.Write([]byte("update relays set max_sessions=$1 where id=$2"))
		args = append(args, int64(maxSessions), relay.DatabaseID)
		relay.MaxSessions = uint32(maxSessions)

	case "MRC":
		mrcUSD, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		mrc := routing.DollarsToNibblins(mrcUSD)
		updateSQL.Write([]byte("update relays set mrc=$1 where id=$2"))
		args = append(args, int64(mrc), relay.DatabaseID)
		relay.MRC = mrc

	case "Overage":
		overageUSD, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		overage := routing.DollarsToNibblins(overageUSD)
		updateSQL.Write([]byte("update relays set overage=$1 where id=$2"))
		args = append(args, int64(overage), relay.DatabaseID)
		relay.Overage = overage

	case "BWRule":
		bwRule, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		if bwRule < 0 || bwRule > 4 {
			return fmt.Errorf("%d is not a valid BandWidthRule value", int64(bwRule))
		}
		updateSQL.Write([]byte("update relays set bw_billing_rule=$1 where id=$2"))
		args = append(args, int64(bwRule), relay.DatabaseID)
		// already checked int validity above
		relay.BWRule, _ = routing.GetBandwidthRuleSQL(int64(bwRule))

	case "ContractTerm":
		term, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		if term < 0 {
			return fmt.Errorf("%d is not a valid ContractTerm value", int32(term))
		}
		updateSQL.Write([]byte("update relays set contract_term=$1 where id=$2"))
		args = append(args, int64(term), relay.DatabaseID)
		relay.ContractTerm = int32(term)

	case "StartDate":
		startDate, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		newStartDate, err := time.Parse("January 2, 2006", startDate)
		if err != nil {
			return fmt.Errorf("Could not parse `%s` - must be of the form 'January 2, 2006'", startDate)
		}

		updateSQL.Write([]byte("update relays set start_date=$1 where id=$2"))
		args = append(args, startDate, relay.DatabaseID)
		relay.StartDate = newStartDate

	case "EndDate":
		endDate, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		newEndDate, err := time.Parse("January 2, 2006", endDate)
		if err != nil {
			return fmt.Errorf("Could not parse `%s` - must be of the form 'January 2, 2006'", endDate)
		}

		updateSQL.Write([]byte("update relays set end_date=$1 where id=$2"))
		args = append(args, endDate, relay.DatabaseID)
		relay.EndDate = newEndDate

	case "Type":
		machineType, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		if machineType < 0 || machineType > 2 {
			return fmt.Errorf("%d is not a valid MachineType value", int64(machineType))
		}
		updateSQL.Write([]byte("update relays set machine_type=$1 where id=$2"))
		args = append(args, int64(machineType), relay.DatabaseID)
		// already checked int validity above
		relay.Type, _ = routing.GetMachineTypeSQL(int64(machineType))

	default:
		return fmt.Errorf("Field %v not available for update (not part of schema)", value)

	}

	// fmt.Printf("--> updateSQL: %s\n", updateSQL.String())

	stmt, err = db.Client.PrepareContext(ctx, updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateRelay SQL", "err", err)
		return err
	}

	// fmt.Println("--> UpdateRelay() stmt.Exec()")
	result, err := stmt.Exec(args...)
	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying relay record", "err", err)
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

	db.relayMutex.Lock()
	db.relays[relayID] = relay
	db.relayMutex.Unlock()

	return nil
}

type sqlRelay struct {
	ID                 uint64
	Name               string
	PublicIP           string // []byte?
	PublicIPPort       int64
	InternalIP         string // []byte?
	InternalIPPort     int64
	PublicKey          []byte
	NICSpeedMbps       int64
	IncludedBandwithGB int64
	DatacenterID       int64
	ManagementIP       string // []byte?
	SSHUser            string
	SSHPort            int64
	State              int64
	MaxSessions        int64
	MRC                int64
	Overage            int64
	BWRule             int64
	ContractTerm       int64
	StartDate          time.Time
	EndDate            time.Time
	MachineType        int64
	DatabaseID         int64
}

// AddRelay adds the provided relay to storage and returns an error if the relay could not be added.
func (db *SQL) AddRelay(ctx context.Context, r routing.Relay) error {

	var sql bytes.Buffer

	db.relayMutex.RLock()
	_, ok := db.relays[r.ID]
	db.relayMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "relay", resourceRef: r.ID}
	}

	publicIPPort, err := strconv.ParseInt(strings.Split(r.Addr.String(), ":")[1], 10, 64)
	if err != nil {
		return fmt.Errorf("Unable to convert PublicIP Port %s to int: %v", strings.Split(r.Addr.String(), ":")[1], err)
	}
	internalIP := ""
	internalIPPort := int64(0)
	if r.InternalAddr.String() != "" {
		internalIP = strings.Split(r.InternalAddr.String(), ":")[0]
		internalIPPort, err = strconv.ParseInt(strings.Split(r.InternalAddr.String(), ":")[1], 10, 64)
		if err != nil {
			return fmt.Errorf("Unable to convert InternalIP Port %s to int: %v", strings.Split(r.InternalAddr.String(), ":")[1], err)
		}
	}

	relay := sqlRelay{
		Name:               r.Name,
		PublicIP:           strings.Split(r.Addr.String(), ":")[0],
		PublicIPPort:       publicIPPort,
		InternalIP:         internalIP,
		InternalIPPort:     internalIPPort,
		PublicKey:          r.PublicKey,
		NICSpeedMbps:       int64(r.NICSpeedMbps),
		IncludedBandwithGB: int64(r.IncludedBandwidthGB),
		DatacenterID:       r.Datacenter.DatabaseID,
		ManagementIP:       r.ManagementAddr,
		SSHUser:            r.SSHUser,
		SSHPort:            r.SSHPort,
		State:              int64(r.State),
		MaxSessions:        int64(r.MaxSessions),
		MRC:                int64(r.MRC),
		Overage:            int64(r.Overage),
		BWRule:             int64(r.BWRule),
		ContractTerm:       int64(r.ContractTerm),
		StartDate:          r.StartDate,
		EndDate:            r.EndDate,
		MachineType:        int64(r.Type),
	}

	sql.Write([]byte("insert into relays ("))
	sql.Write([]byte("contract_term, display_name, end_date, included_bandwidth_gb, "))
	sql.Write([]byte("management_ip, max_sessions, mrc, overage, port_speed, public_ip, "))
	sql.Write([]byte("public_ip_port, public_key, ssh_port, ssh_user, start_date, "))
	sql.Write([]byte("bw_billing_rule, datacenter, machine_type, relay_state, internal_ip, internal_ip_port "))
	sql.Write([]byte(") values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, "))
	sql.Write([]byte("$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddRelay SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
		relay.ContractTerm,
		relay.Name,
		relay.EndDate,
		relay.IncludedBandwithGB,
		relay.ManagementIP,
		relay.MaxSessions,
		relay.MRC,
		relay.Overage,
		relay.NICSpeedMbps,
		relay.PublicIP,
		relay.PublicIPPort,
		relay.PublicKey,
		relay.SSHPort,
		relay.SSHUser,
		relay.StartDate,
		relay.BWRule,
		relay.DatacenterID,
		relay.MachineType,
		relay.State,
		relay.InternalIP,
		relay.InternalIPPort,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding relay", "err", err)
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

	db.syncRelays(ctx)

	db.IncrementSequenceNumber(ctx)

	return nil
}

// RemoveRelay removes a relay with the provided relay ID from storage and
// returns any database errors to the caller
func (db *SQL) RemoveRelay(ctx context.Context, id uint64) error {
	var sql bytes.Buffer

	db.relayMutex.RLock()
	relay, ok := db.relays[id]
	db.relayMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%016x", id)}
	}

	sql.Write([]byte("delete from relays where id = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveRelay SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(relay.DatabaseID)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing relay", "err", err)
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

	db.relayMutex.Lock()
	delete(db.relays, relay.ID)
	db.relayMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

// SetRelay updates the relay in storage with the provided copy and returns an
// error if the relay could not be updated.
// TODO: chopping block
func (db *SQL) SetRelay(ctx context.Context, r routing.Relay) error {

	var sql bytes.Buffer

	db.relayMutex.RLock()
	_, ok := db.relays[r.ID]
	db.relayMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%016x", r.ID)}
	}

	publicIPPort, err := strconv.ParseInt(strings.Split(r.Addr.String(), ":")[1], 10, 64)
	if err != nil {
		return fmt.Errorf("Unable to convert PublicIP Port %s to int: %v", strings.Split(r.Addr.String(), ":")[1], err)
	}

	internalIP := ""
	internalIPPort := int64(0)
	if r.InternalAddr.String() != "" {
		internalIP = strings.Split(r.InternalAddr.String(), ":")[0]
		internalIPPort, err = strconv.ParseInt(strings.Split(r.InternalAddr.String(), ":")[1], 10, 64)
		if err != nil {
			return fmt.Errorf("Unable to convert InternalIP Port %s to int: %v", strings.Split(r.InternalAddr.String(), ":")[1], err)
		}
	}

	relay := sqlRelay{
		Name:               r.Name,
		PublicIP:           strings.Split(r.Addr.String(), ":")[0],
		PublicIPPort:       publicIPPort,
		InternalIP:         internalIP,
		InternalIPPort:     internalIPPort,
		PublicKey:          r.PublicKey,
		NICSpeedMbps:       int64(r.NICSpeedMbps),
		IncludedBandwithGB: int64(r.IncludedBandwidthGB),
		DatacenterID:       r.Datacenter.DatabaseID,
		ManagementIP:       r.ManagementAddr,
		SSHUser:            r.SSHUser,
		SSHPort:            r.SSHPort,
		State:              int64(r.State),
		MaxSessions:        int64(r.MaxSessions),
		MRC:                int64(r.MRC),
		Overage:            int64(r.Overage),
		BWRule:             int64(r.BWRule),
		ContractTerm:       int64(r.ContractTerm),
		StartDate:          r.StartDate,
		EndDate:            r.EndDate,
		MachineType:        int64(r.Type),
	}

	sql.Write([]byte("update relays set ("))
	sql.Write([]byte("contract_term, display_name, end_date, included_bandwidth_gb, "))
	sql.Write([]byte("management_ip, max_sessions, mrc, overage, port_speed, public_ip, "))
	sql.Write([]byte("public_ip_port, public_key, ssh_port, ssh_user, start_date, "))
	sql.Write([]byte("bw_billing_rule, datacenter, machine_type, relay_state, internal_ip, internal_ip_port "))
	sql.Write([]byte(") = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, "))
	sql.Write([]byte("$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) where id = $22"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing SetRelay SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
		relay.ContractTerm,
		relay.Name,
		relay.EndDate,
		relay.IncludedBandwithGB,
		relay.ManagementIP,
		relay.MaxSessions,
		relay.MRC,
		relay.Overage,
		relay.NICSpeedMbps,
		relay.PublicIP,
		relay.PublicIPPort,
		relay.PublicKey,
		relay.SSHPort,
		relay.SSHUser,
		relay.StartDate,
		relay.BWRule,
		relay.DatacenterID,
		relay.MachineType,
		relay.State,
		relay.InternalIP,
		relay.InternalIPPort,
		r.DatabaseID,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying relay", "err", err)
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

	db.relayMutex.Lock()
	db.relays[r.ID] = r
	db.relayMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

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

// RemoveDatacenter removes a datacenter with the provided datacenter ID from storage
// and returns an error if:
//  1. The datacenter ID does not exist
//  2. Removing the datacenter would break foreigh key relationships (datacenter_maps, relays)
//  3. Any other error returned from the database
func (db *SQL) RemoveDatacenter(ctx context.Context, id uint64) error {
	var sql bytes.Buffer

	db.datacenterMutex.RLock()
	datacenter, ok := db.datacenters[id]
	db.datacenterMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%016x", id)}
	}

	sql.Write([]byte("delete from datacenters where id = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveDatacenter SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(datacenter.DatabaseID)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing datacenter", "err", err)
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

	db.datacenterMutex.Lock()
	delete(db.datacenters, datacenter.ID)
	db.datacenterMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

// SetDatacenter updates a subset of fields in the datacenters table.
//		Name
//		Enabled
//		Latitude
//		Longitude
//		SupplierName
func (db *SQL) SetDatacenter(ctx context.Context, d routing.Datacenter) error {

	var sql bytes.Buffer

	db.datacenterMutex.RLock()
	_, ok := db.datacenters[d.ID]
	db.datacenterMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%016x", d.ID)}
	}

	dc := sqlDatacenter{
		Name:          d.Name,
		Enabled:       d.Enabled,
		Latitude:      d.Location.Latitude,
		Longitude:     d.Location.Longitude,
		SupplierName:  d.SupplierName,
		SellerID:      d.SellerID,
		StreetAddress: d.StreetAddress,
	}

	sql.Write([]byte("update datacenters set ("))
	sql.Write([]byte("display_name, enabled, latitude, longitude, supplier_name, street_address, "))
	sql.Write([]byte("seller_id ) = ($1, $2, $3, $4, $5, $6, $7) where id = $8"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing SetDatacenter SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(dc.Name,
		dc.Enabled,
		dc.Latitude,
		dc.Longitude,
		dc.SupplierName,
		dc.StreetAddress,
		dc.SellerID,
		d.DatabaseID,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying datacenter", "err", err)
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

	db.datacenterMutex.Lock()
	db.datacenters[d.ID] = d
	db.datacenterMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

// GetDatacenterMapsForBuyer returns the list of datacenter aliases in use for a given (internally generated) buyerID. Returns
// an empty []routing.DatacenterMap if there are no aliases for that buyerID.
func (db *SQL) GetDatacenterMapsForBuyer(buyerID uint64) map[uint64]routing.DatacenterMap {
	db.datacenterMapMutex.RLock()
	defer db.datacenterMapMutex.RUnlock()

	// buyer can have multiple dc aliases
	var dcs = make(map[uint64]routing.DatacenterMap)
	for _, dc := range db.datacenterMaps {
		if dc.BuyerID == buyerID {
			dcs[dc.DatacenterID] = dc
		}
	}

	return dcs
}

// AddDatacenterMap adds a new datacenter alias for the given buyer and datacenter IDs
func (db *SQL) AddDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {

	var sql bytes.Buffer

	bID := dcMap.BuyerID

	dcID := dcMap.DatacenterID

	buyer, ok := db.buyers[bID]
	if !ok {
		return &DoesNotExistError{resourceType: "BuyerID", resourceRef: dcMap.BuyerID}
	}

	datacenter, ok := db.datacenters[dcID]
	if !ok {
		return &DoesNotExistError{resourceType: "DatacenterID", resourceRef: dcMap.DatacenterID}
	}

	sql.Write([]byte("insert into datacenter_maps (alias, buyer_id, datacenter_id) "))
	sql.Write([]byte("values ($1, $2, $3)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddDatacenterMap SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(dcMap.Alias,
		buyer.DatabaseID,
		datacenter.DatabaseID,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding DatacenterMap", "err", err)
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

	db.syncDatacenterMaps(ctx)

	db.IncrementSequenceNumber(ctx)

	return nil
}

// ListDatacenterMaps returns a list of alias/buyer mappings for the specified datacenter ID. An
// empty dcID returns a list of all maps.
func (db *SQL) ListDatacenterMaps(dcID uint64) map[uint64]routing.DatacenterMap {
	db.datacenterMapMutex.RLock()
	defer db.datacenterMapMutex.RUnlock()

	var dcs = make(map[uint64]routing.DatacenterMap)
	for _, dc := range db.datacenterMaps {
		if dc.DatacenterID == dcID || dcID == 0 {
			id := crypto.HashID(dc.Alias + fmt.Sprintf("%x", dc.BuyerID) + fmt.Sprintf("%x", dc.DatacenterID))
			dcs[id] = dc
		}
	}

	return dcs
}

// RemoveDatacenterMap removes an entry from the DatacenterMaps table
func (db *SQL) RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {
	var sql bytes.Buffer

	id := crypto.HashID(dcMap.Alias + fmt.Sprintf("%x", dcMap.BuyerID) + fmt.Sprintf("%x", dcMap.DatacenterID))

	db.datacenterMapMutex.RLock()
	dcMap, ok := db.datacenterMaps[id]
	db.datacenterMapMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "datacenter map", resourceRef: fmt.Sprintf("%016x", id)}
	}

	buyer := db.buyers[dcMap.BuyerID]
	datacenter := db.datacenters[dcMap.DatacenterID]

	sql.Write([]byte("delete from datacenter_maps where buyer_id = $1 and datacenter_id = $2"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveDatacenterMap SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(buyer.DatabaseID, datacenter.DatabaseID)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing datacenter map", "err", err)
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

	db.datacenterMapMutex.Lock()
	delete(db.datacenterMaps, id)
	db.datacenterMapMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

// SetRelayMetadata provides write access to ops metadat (mrc, overage, etc)
func (db *SQL) SetRelayMetadata(ctx context.Context, relay routing.Relay) error {
	// return fmt.Errorf("SetRelayMetadata() not implemented in SQL Storer")
	fmt.Printf("SetRelayMetadata(): %s\n", relay.String())
	err := db.SetRelay(ctx, relay)
	return err
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
func (db *SQL) CheckSequenceNumber(ctx context.Context) (bool, int64, error) {
	var sequenceNumber int64

	err := db.Client.QueryRowContext(ctx, "select sync_sequence_number from metadata").Scan(&sequenceNumber)

	if err == sql.ErrNoRows {
		level.Error(db.Logger).Log("during", "No sequence number returned", "err", err)
		return false, -1, err
	} else if err != nil {
		level.Error(db.Logger).Log("during", "query error", "err", err)
		return false, -1, err
	}

	db.sequenceNumberMutex.RLock()
	localSeqNum := db.SyncSequenceNumber
	db.sequenceNumberMutex.RUnlock()

	if localSeqNum < sequenceNumber {
		db.sequenceNumberMutex.Lock()
		db.SyncSequenceNumber = sequenceNumber
		db.sequenceNumberMutex.Unlock()
		return true, sequenceNumber, nil
	} else if localSeqNum == sequenceNumber {
		return false, -1, nil
	} else {
		err = fmt.Errorf("local sequence number larger than remote: %d > %d", localSeqNum, sequenceNumber)
		level.Error(db.Logger).Log("during", "query error", "err", err)
		return false, -1, err
	}
}

// SetSequenceNumber is required for testing with the Firestore emulator
func (db *SQL) SetSequenceNumber(ctx context.Context, sequenceNumber int64) error {
	stmt, err := db.Client.PrepareContext(ctx, "update metadata set sync_sequence_number = $1")
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing SQL", "err", err)
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

	db.sequenceNumberMutex.Lock()
	db.SyncSequenceNumber = sequenceNumber
	db.sequenceNumberMutex.Unlock()

	return nil
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
		level.Error(db.Logger).Log("during", "error preparing SQL", "err", err)
		return err
	}

	_, err = stmt.Exec(sequenceNumber)
	if err != nil {
		level.Error(db.Logger).Log("during", "error setting sequence number", "err", err)
	}

	// SQLite3 does not like this check but it is necessary...
	// TODO: fix/research
	// result, err := stmt.Exec(sequenceNumber)
	// if err != nil {
	// 	level.Error(db.Logger).Log("during", "error setting sequence number", "err", err)
	// }
	// rows, err := result.RowsAffected()
	// if err != nil {
	// 	level.Error(db.Logger).Log("during", "RowsAffected returned an error", "err", err)
	// }
	// if rows != 1 {
	// 	level.Error(db.Logger).Log("during", "RowsAffected <> 1", "err", err)
	// }

	return nil
}

type sqlDatacenter struct {
	ID            int64
	Name          string
	Enabled       bool
	Latitude      float32
	Longitude     float32
	SupplierName  string
	StreetAddress string
	SellerID      int64
}

// AddDatacenter adds the provided datacenter to storage. It enforces business rule
// that datacenter names are of the form: "seller"."location"."optional number/enum"
//
// The seller_id is reqired by the schema. The client interface must already have a
// seller defined.
func (db *SQL) AddDatacenter(ctx context.Context, datacenter routing.Datacenter) error {

	var sql bytes.Buffer

	db.datacenterMutex.RLock()
	_, ok := db.datacenters[datacenter.ID]
	db.datacenterMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "datacenter", resourceRef: datacenter.ID}
	}

	dc := sqlDatacenter{
		Name:          datacenter.Name,
		Enabled:       datacenter.Enabled,
		Latitude:      datacenter.Location.Latitude,
		Longitude:     datacenter.Location.Longitude,
		SupplierName:  datacenter.SupplierName,
		SellerID:      datacenter.SellerID,
		StreetAddress: datacenter.StreetAddress,
	}

	sql.Write([]byte("insert into datacenters ("))
	sql.Write([]byte("display_name, enabled, latitude, longitude, supplier_name, street_address, "))
	sql.Write([]byte("seller_id ) values ($1, $2, $3, $4, $5, $6, $7)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddDatacenter SQL", "err", err)
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

	db.syncDatacenters(ctx)

	db.IncrementSequenceNumber(ctx)

	return nil
}

// RouteShaders returns a slice of route shaders for the given buyer ID
func (db *SQL) RouteShaders(buyerID uint64) ([]core.RouteShader, error) {
	db.routeShaderMutex.RLock()
	defer db.routeShaderMutex.RUnlock()

	routeShaders, found := db.routeShaders[buyerID]
	if !found {
		return []core.RouteShader{}, &DoesNotExistError{resourceType: "route shaders", resourceRef: fmt.Sprintf("%x", buyerID)}
	}

	return routeShaders, nil
}

// InternalConfig returns the InternalConfig entry for the specified buyer
func (db *SQL) InternalConfig(buyerID uint64) (core.InternalConfig, error) {
	db.internalConfigMutex.RLock()
	defer db.internalConfigMutex.RUnlock()

	internalConfig, found := db.internalConfigs[buyerID]
	if !found {
		return core.InternalConfig{}, &DoesNotExistError{resourceType: "internal config", resourceRef: fmt.Sprintf("%x", buyerID)}
	}

	return internalConfig, nil

}

// AddInternalConfig adds an InternalConfig for the specified buyer
func (db *SQL) AddInternalConfig(ctx context.Context, ic core.InternalConfig, buyerID uint64) error {

	var sql bytes.Buffer

	db.internalConfigMutex.RLock()
	_, ok := db.internalConfigs[buyerID]
	db.internalConfigMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "InternalConfig", resourceRef: buyerID}
	}

	db.buyerMutex.RLock()
	buyer, err := db.Buyer(buyerID)
	db.buyerMutex.RUnlock()

	if err != nil {
		return &DoesNotExistError{resourceType: "Buyer", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	internalConfig := sqlInternalConfig{
		RouteSelectThreshold:       int64(ic.RouteSelectThreshold),
		RouteSwitchThreshold:       int64(ic.RouteSwitchThreshold),
		MaxLatencyTradeOff:         int64(ic.MaxLatencyTradeOff),
		RTTVetoDefault:             int64(ic.RTTVeto_Default),
		RTTVetoPacketLoss:          int64(ic.RTTVeto_PacketLoss),
		RTTVetoMultipath:           int64(ic.RTTVeto_Multipath),
		MultipathOverloadThreshold: int64(ic.MultipathOverloadThreshold),
		TryBeforeYouBuy:            ic.TryBeforeYouBuy,
		ForceNext:                  ic.ForceNext,
		LargeCustomer:              ic.LargeCustomer,
		Uncommitted:                ic.Uncommitted,
		MaxRTT:                     int64(ic.MaxRTT),
	}

	sql.Write([]byte("insert into rs_internal_configs "))
	sql.Write([]byte("(max_latency_tradeoff, max_rtt, multipath_overload_threshold, "))
	sql.Write([]byte("route_switch_threshold, route_select_threshold, rtt_veto_default, "))
	sql.Write([]byte("rtt_veto_multipath, rtt_veto_packetloss, try_before_you_buy, force_next, "))
	sql.Write([]byte("large_customer, is_uncommitted, buyer_id) "))
	sql.Write([]byte("values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddInternalConfig SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
		internalConfig.MaxLatencyTradeOff,
		internalConfig.MaxRTT,
		internalConfig.MultipathOverloadThreshold,
		internalConfig.RouteSwitchThreshold,
		internalConfig.RouteSelectThreshold,
		internalConfig.RTTVetoDefault,
		internalConfig.RTTVetoMultipath,
		internalConfig.RTTVetoPacketLoss,
		internalConfig.TryBeforeYouBuy,
		internalConfig.ForceNext,
		internalConfig.LargeCustomer,
		internalConfig.Uncommitted,
		buyer.DatabaseID,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding internal config", "err", err)
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

	db.syncInternalConfigs(ctx)

	db.IncrementSequenceNumber(ctx)

	return nil
}

func (db *SQL) RemoveInternalConfig(ctx context.Context, buyerID uint64) error {
	var sql bytes.Buffer

	db.internalConfigMutex.RLock()
	_, ok := db.internalConfigs[buyerID]
	db.internalConfigMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "InternalConfig", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	buyer, err := db.Buyer(buyerID)
	if err != nil {
		return &DoesNotExistError{resourceType: "Buyer", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	sql.Write([]byte("delete from internal_configs where where buyer_id = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveRelay SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(buyer.DatabaseID)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing internal config", "err", err)
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

	db.internalConfigMutex.Lock()
	delete(db.internalConfigs, buyerID)
	db.internalConfigMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

func (db *SQL) UpdateInternalConfig(ctx context.Context, buyerID uint64, field string, value interface{}) error {

	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt

	ic, err := db.InternalConfig(buyerID)
	if err != nil {
		return &DoesNotExistError{resourceType: "internal config", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	db.buyerMutex.RLock()
	buyer, err := db.Buyer(buyerID)
	db.buyerMutex.RUnlock()

	if err != nil {
		return &DoesNotExistError{resourceType: "Buyer", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	switch field {
	case "RouteSelectThreshold":
		routeSelectThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteSelectThreshold: %v is not a valid int32 type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set route_select_threshold=$1 where buyer_id=$2"))
		args = append(args, routeSelectThreshold, buyer.DatabaseID)
		ic.RouteSelectThreshold = routeSelectThreshold
	case "RouteSwitchThreshold":
		routeSwitchThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteSwitchThreshold: %v is not a valid int32 type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set route_switch_threshold=$1 where buyer_id=$2"))
		args = append(args, routeSwitchThreshold, buyer.DatabaseID)
		ic.RouteSwitchThreshold = routeSwitchThreshold
	case "MaxLatencyTradeOff":
		maxLatencyTradeOff, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MaxLatencyTradeOff: %v is not a valid int32 type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set max_latency_tradeoff=$1 where buyer_id=$2"))
		args = append(args, maxLatencyTradeOff, buyer.DatabaseID)
		ic.MaxLatencyTradeOff = maxLatencyTradeOff
	case "RTTVeto_Default":
		rttVetoDefault, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_Default: %v is not a valid int32 type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_default=$1 where buyer_id=$2"))
		args = append(args, rttVetoDefault, buyer.DatabaseID)
		ic.RTTVeto_Default = rttVetoDefault
	case "RTTVeto_PacketLoss":
		rttVetoPacketLoss, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_PacketLoss: %v is not a valid int32 type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_packetloss=$1 where buyer_id=$2"))
		args = append(args, rttVetoPacketLoss, buyer.DatabaseID)
		ic.RTTVeto_PacketLoss = rttVetoPacketLoss
	case "RTTVeto_Multipath":
		rttVetoMultipath, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_Multipath: %v is not a valid int32 type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_multipath=$1 where buyer_id=$2"))
		args = append(args, rttVetoMultipath, buyer.DatabaseID)
		ic.RTTVeto_Multipath = rttVetoMultipath
	case "MultipathOverloadThreshold":
		multipathOverloadThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MultipathOverloadThreshold: %v is not a valid int32 type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set multipath_overload_threshold=$1 where buyer_id=$2"))
		args = append(args, multipathOverloadThreshold, buyer.DatabaseID)
		ic.MultipathOverloadThreshold = multipathOverloadThreshold
	case "TryBeforeYouBuy":
		tryBeforeYouBuy, ok := value.(bool)
		if !ok {
			return fmt.Errorf("TryBeforeYouBuy: %v is not a valid boolean type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set try_before_you_buy=$1 where buyer_id=$2"))
		args = append(args, tryBeforeYouBuy, buyer.DatabaseID)
		ic.TryBeforeYouBuy = tryBeforeYouBuy
	case "ForceNext":
		forceNext, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ForceNext: %v is not a valid boolean type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set force_next=$1 where buyer_id=$2"))
		args = append(args, forceNext, buyer.DatabaseID)
		ic.ForceNext = forceNext
	case "LargeCustomer":
		largeCustomer, ok := value.(bool)
		if !ok {
			return fmt.Errorf("LargeCustomer: %v is not a valid boolean type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set large_customer=$1 where buyer_id=$2"))
		args = append(args, largeCustomer, buyer.DatabaseID)
		ic.LargeCustomer = largeCustomer
	case "Uncommitted":
		uncommitted, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Uncommitted: %v is not a valid boolean type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set is_uncommitted=$1 where buyer_id=$2"))
		args = append(args, uncommitted, buyer.DatabaseID)
		ic.Uncommitted = uncommitted
	case "MaxRTT":
		maxRTT, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MaxRTT: %v is not a valid int32 type", value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set max_rtt=$1 where buyer_id=$2"))
		args = append(args, maxRTT, buyer.DatabaseID)
		ic.MaxRTT = maxRTT

	}

	stmt, err = db.Client.PrepareContext(ctx, updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateRelay SQL", "err", err)
		return err
	}

	// fmt.Println("--> UpdateInternalConfig() stmt.Exec()")
	result, err := stmt.Exec(args...)
	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying internal_config record", "err", err)
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

	db.internalConfigMutex.Lock()
	db.internalConfigs[buyerID] = ic
	db.internalConfigMutex.Unlock()

	return nil
}

type featureFlag struct {
	Name        string
	Description string
	Enabled     bool
}

func (db *SQL) GetFeatureFlags() map[string]bool {
	return map[string]bool{}
}

func (db *SQL) GetFeatureFlagByName(flagName string) (map[string]bool, error) {
	return map[string]bool{}, fmt.Errorf(("GetFeatureFlagByName not yet impemented in SQL storer"))
}

func (db *SQL) SetFeatureFlagByName(ctx context.Context, flagName string, flagVal bool) error {
	return fmt.Errorf("SetFeatureFlagByName not yet impemented in SQL storer")
}

func (db *SQL) RemoveFeatureFlagByName(ctx context.Context, flagName string) error {
	return fmt.Errorf("RemoveFeatureFlagByName not yet impemented in SQL storer")
}

func (db *SQL) AddRouteShader(ctx context.Context, routeShader core.RouteShader, buyerID uint64) error {
	return fmt.Errorf("AddRouteShader not yet impemented in SQL storer")
}

func (db *SQL) UpdateRouteShader(ctx context.Context, buyerID uint64, index uint64, field string, value interface{}) error {
	return fmt.Errorf("UpdateRouteShader not yet impemented in SQL storer")
}

func (db *SQL) RemoveRouteSHader(ctx context.Context, buyerID uint64, index uint64) error {
	return fmt.Errorf("RemoveRouteSHader not yet impemented in SQL storer")
}
