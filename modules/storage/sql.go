package storage

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
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
	buyers         map[uint64]routing.Buyer // this index is in fact the database ID cast to uint64
	sellers        map[string]routing.Seller
	datacenterMaps map[uint64]routing.DatacenterMap

	internalConfigs map[uint64]core.InternalConfig // index: buyer ID
	routeShaders    map[uint64]core.RouteShader    // index: buyer ID
	bannedUsers     map[uint64]map[uint64]bool     // index: buyerID

	datacenterMutex     sync.RWMutex
	relayMutex          sync.RWMutex
	customerMutex       sync.RWMutex
	buyerMutex          sync.RWMutex
	sellerMutex         sync.RWMutex
	datacenterMapMutex  sync.RWMutex
	sequenceNumberMutex sync.RWMutex
	internalConfigMutex sync.RWMutex
	routeShaderMutex    sync.RWMutex
	bannedUserMutex     sync.RWMutex

	//  int64: PostgreSQL primary key
	// uint64: backend/storer internal ID
	datacenterIDs map[int64]uint64
	relayIDs      map[int64]uint64
	customerIDs   map[int64]string
	buyerIDs      map[uint64]int64 // buyerIDs map is inverted
	sellerIDs     map[int64]string

	datacenterIDsMutex  sync.RWMutex
	relayIDsMutex       sync.RWMutex
	customerIDsMutex    sync.RWMutex
	buyerIDsMutex       sync.RWMutex
	sellerIDsMutex      sync.RWMutex
	datacenterMapsMutex sync.RWMutex

	SyncSequenceNumber int64
}

// Customer retrieves a Customer record using the company code
func (db *SQL) Customer(customerCode string) (routing.Customer, error) {

	var querySQL bytes.Buffer
	var customer sqlCustomer

	querySQL.Write([]byte("select id, automatic_signin_domain,"))
	querySQL.Write([]byte("customer_name, customer_code from customers where customer_code = $1"))

	row := db.Client.QueryRow(querySQL.String(), customerCode)
	err := row.Scan(&customer.ID,
		&customer.AutomaticSignInDomains,
		&customer.Name,
		&customer.CustomerCode)
	switch err {
	case sql.ErrNoRows:
		level.Error(db.Logger).Log("during", "Customer() no rows were returned!")
		return routing.Customer{}, &DoesNotExistError{resourceType: "customer", resourceRef: customerCode}
	case nil:
		c := routing.Customer{
			Code:                   customer.CustomerCode,
			Name:                   customer.Name,
			AutomaticSignInDomains: customer.AutomaticSignInDomains,
			DatabaseID:             customer.ID,
		}
		return c, nil
	default:
		level.Error(db.Logger).Log("during", "Customer() QueryRow returned an error: %v", err)
		return routing.Customer{}, err
	}

}

// CustomerWithName retrieves a record using the customer's name
// func (db *SQL) CustomerWithName(name string) (routing.Customer, error) {
// 	var querySQL bytes.Buffer
// 	var customer sqlCustomer

// 	querySQL.Write([]byte("select id, automatic_signin_domain,"))
// 	querySQL.Write([]byte("customer_name, customer_code from customers where customer_name = $1"))

// 	row := db.Client.QueryRow(querySQL.String(), name)
// 	err := row.Scan(&customer.ID,
// 		&customer.AutomaticSignInDomains,
// 		&customer.Name,
// 		&customer.CustomerCode)
// 	switch err {
// 	case sql.ErrNoRows:
// 		level.Error(db.Logger).Log("during", "CustomerWithName() no rows were returned!")
// 		return routing.Customer{}, &DoesNotExistError{resourceType: "customer", resourceRef: fmt.Sprintf("%s", name)}
// 	case nil:
// 		c := routing.Customer{
// 			Code:                   customer.CustomerCode,
// 			Name:                   customer.Name,
// 			AutomaticSignInDomains: customer.AutomaticSignInDomains,
// 			DatabaseID:             customer.ID,
// 		}
// 		return c, nil
// 	default:
// 		level.Error(db.Logger).Log("during", "CustomerWithName() QueryRow returned an error: %v", err)
// 		return routing.Customer{}, err
// 	}
// }

// Customers retrieves the full list
// TODO: not covered by sql_test.go
func (db *SQL) Customers() []routing.Customer {
	var sql bytes.Buffer
	var customer sqlCustomer

	customers := []routing.Customer{}
	customerIDs := make(map[int64]string)

	sql.Write([]byte("select id, automatic_signin_domain, "))
	sql.Write([]byte("customer_name, customer_code from customers"))

	rows, err := db.Client.QueryContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "Customers(): QueryContext returned an error", "err", err)
		return []routing.Customer{}
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&customer.ID,
			&customer.AutomaticSignInDomains,
			&customer.Name,
			&customer.CustomerCode,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "Customers(): error parsing returned row", "err", err)
			return []routing.Customer{}
		}

		customerIDs[customer.ID] = customer.CustomerCode

		c := routing.Customer{
			Code:                   customer.CustomerCode,
			Name:                   customer.Name,
			AutomaticSignInDomains: customer.AutomaticSignInDomains,
			DatabaseID:             customer.ID,
		}

		customers = append(customers, c)
	}

	sort.Slice(customers, func(i int, j int) bool { return customers[i].Name < customers[j].Name })
	return customers
}

type sqlCustomer struct {
	ID                     int64
	Name                   string
	AutomaticSignInDomains string
	Debug                  bool
	CustomerCode           string
	DatabaseID             int64
	BuyerID                sql.NullInt64 // loaded during syncCustomers()
	SellerID               sql.NullInt64 // loaded during syncCustomers()
}

func (db *SQL) AddCustomer(ctx context.Context, c routing.Customer) error {
	var sql bytes.Buffer

	customer := sqlCustomer{
		CustomerCode:           c.Code,
		Name:                   c.Name,
		AutomaticSignInDomains: c.AutomaticSignInDomains,
	}

	sql.Write([]byte("insert into customers ("))
	sql.Write([]byte("automatic_signin_domain, customer_name, customer_code"))
	sql.Write([]byte(") values ($1, $2, $3)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddCustomer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
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

	sql.Write([]byte("delete from customers where customer_code = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveCustomer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(customerCode)

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

	return nil
}

// SetCustomer modifies a subset of fields in a customer record
// in the database. Modifield fields:
//		Name
//		AutomaticSigninDomains
//		Active
//		Debug
// TODO: remove - need to modify AuthService.UpdateAutoSignupDomains to
//       use UpdateCustomer() and then drop this method
func (db *SQL) SetCustomer(ctx context.Context, c routing.Customer) error {

	var sql bytes.Buffer

	sql.Write([]byte("update customers set (automatic_signin_domain, customer_name) ="))
	sql.Write([]byte("($1, $2) where customer_code = $3"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing SetCustomer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(c.AutomaticSignInDomains, c.Name, c.Code)
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

	return nil
}

// Buyer gets a copy of a buyer with the specified buyer ID,
// and returns an empty buyer and an error if a buyer with that ID doesn't exist in storage.
func (db *SQL) Buyer(ephemeralBuyerID uint64) (routing.Buyer, error) {

	// dbBuyerID := uint64(db.buyerIDs[ephemeralBuyerID])
	// db.buyerMutex.RLock()
	// b, found := db.buyers[dbBuyerID]
	// db.buyerMutex.RUnlock()

	// if !found {
	// 	return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%x", ephemeralBuyerID)}
	// }

	sqlBuyerID := int64(ephemeralBuyerID)

	var querySQL bytes.Buffer
	var buyer sqlBuyer

	querySQL.Write([]byte("select short_name, is_live_customer, debug, public_key, customer_id "))
	querySQL.Write([]byte("from buyers where sdk_generated_id = $1"))

	row := db.Client.QueryRow(querySQL.String(), sqlBuyerID)
	err := row.Scan(
		&buyer.DatabaseID,
		&buyer.ShortName,
		&buyer.IsLiveCustomer,
		&buyer.Debug,
		&buyer.PublicKey,
		&buyer.CustomerID,
	)
	switch err {
	case sql.ErrNoRows:
		level.Error(db.Logger).Log("during", "Customer() no rows were returned!")
		return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%016x", ephemeralBuyerID)}
	case nil:

		ic, err := db.InternalConfig(ephemeralBuyerID)
		if err != nil {
			level.Error(db.Logger).Log("during", "Buyer() InternalConfig query returned an error: %v", err)
			return routing.Buyer{}, err
		}

		rs, err := db.RouteShader(ephemeralBuyerID)
		if err != nil {
			level.Error(db.Logger).Log("during", "Buyer() RouteShader query returned an error: %v", err)
			return routing.Buyer{}, err
		}

		b := routing.Buyer{
			ID:             buyer.ID,
			HexID:          fmt.Sprintf("%016x", buyer.ID),
			ShortName:      buyer.ShortName,
			CompanyCode:    buyer.ShortName,
			Live:           buyer.IsLiveCustomer,
			Debug:          buyer.Debug,
			PublicKey:      buyer.PublicKey,
			RouteShader:    rs,
			InternalConfig: ic,
			CustomerID:     buyer.CustomerID,
			DatabaseID:     buyer.DatabaseID,
		}
		return b, nil
	default:
		level.Error(db.Logger).Log("during", "Buyer() QueryRow returned an error: %v", err)
		return routing.Buyer{}, err
	}

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
	_, ok := db.buyers[uint64(b.DatabaseID)]
	db.buyerMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "buyer", resourceRef: b.ID}
	}

	c, err := db.Customer(b.CompanyCode)
	if err != nil {
		return &DoesNotExistError{resourceType: "customer", resourceRef: b.CompanyCode}
	}

	buyer := sqlBuyer{
		ID:             b.ID,
		CompanyCode:    b.CompanyCode,
		ShortName:      b.CompanyCode,
		IsLiveCustomer: b.Live,
		Debug:          b.Debug,
		PublicKey:      b.PublicKey,
		CustomerID:     c.DatabaseID,
	}

	// Add the buyer in remote storage
	sql.Write([]byte("insert into buyers ("))
	sql.Write([]byte("sdk_generated_id, short_name, is_live_customer, debug, public_key, customer_id"))
	sql.Write([]byte(") values ($1, $2, $3, $4, $5, $6)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddBuyer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
		int64(buyer.ID),
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

	// get the DatabaseID loaded
	dbBuyerID := uint64(db.buyerIDs[buyer.ID])

	db.buyerMutex.RLock()
	newBuyer := db.buyers[dbBuyerID]
	db.buyerMutex.RUnlock()

	newBuyer.HexID = fmt.Sprintf("%016x", buyer.ID)
	newBuyer.RouteShader = core.NewRouteShader()
	newBuyer.InternalConfig = core.NewInternalConfig()

	// update local fields
	db.buyerMutex.Lock()
	db.buyers[dbBuyerID] = newBuyer
	db.buyerMutex.Unlock()

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
func (db *SQL) RemoveBuyer(ctx context.Context, ephemeralBuyerID uint64) error {
	var sql bytes.Buffer

	buyerID := db.buyerIDs[ephemeralBuyerID]

	db.buyerMutex.RLock()
	buyer, ok := db.buyers[uint64(buyerID)]
	db.buyerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%016x", ephemeralBuyerID)}
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
	delete(db.buyers, uint64(buyerID))
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

	ephemeralBuyerID := b.ID
	buyerID := db.buyerIDs[ephemeralBuyerID]

	db.buyerMutex.RLock()
	_, ok := db.buyers[uint64(buyerID)]
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
	db.buyers[uint64(b.DatabaseID)] = b
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
	ID                       string
	ShortName                string
	Secret                   bool
	EgressPriceNibblinsPerGB int64
	CustomerID               int64
	DatabaseID               int64
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
		ID:                       s.ID,
		ShortName:                s.ShortName,
		Secret:                   s.Secret,
		EgressPriceNibblinsPerGB: int64(s.EgressPriceNibblinsPerGB),
		CustomerID:               c.DatabaseID,
	}

	// Add the seller in remote storage
	sql.Write([]byte("insert into sellers ("))
	sql.Write([]byte("short_name, public_egress_price, secret, customer_id"))
	sql.Write([]byte(") values ($1, $2, $3, $4)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddSeller SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
		newSellerData.ShortName,
		newSellerData.EgressPriceNibblinsPerGB,
		newSellerData.Secret,
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

	sql.Write([]byte("update sellers set public_egress_price = $1 where id = $2 "))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing SetBuyer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(seller.EgressPriceNibblinsPerGB, seller.DatabaseID)
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

		// "removing" a relay zeroes-out the public IP address
		if addrString == "" {
			updateSQL.Write([]byte("update relays set (public_ip, public_ip_port) = (null, null) "))
			updateSQL.Write([]byte("where id=$1"))
			args = append(args, relay.DatabaseID)
		} else {
			uriTuple := strings.Split(addrString, ":")
			if len(uriTuple) < 2 {
				return fmt.Errorf("unable to parse URI fo Add field: %v - you may be missing the port number?", value)
			} else if uriTuple[0] == "" || uriTuple[1] == "" {
				return fmt.Errorf("unable to parse URI fo Add field: %v", value)
			}
			updateSQL.Write([]byte("update relays set (public_ip, public_ip_port) = ($1, $2) "))
			updateSQL.Write([]byte("where id=$3"))
			args = append(args, uriTuple[0], uriTuple[1], relay.DatabaseID)
		}

		// addr will be ':0' for "removed" relays
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

		if addrString == "" {
			updateSQL.Write([]byte("update relays set (internal_ip, internal_ip_port) = (null, null) "))
			updateSQL.Write([]byte("where id=$1"))
			args = append(args, relay.DatabaseID)
			relay.InternalAddr = net.UDPAddr{}

		} else {
			uriTuple := strings.Split(addrString, ":")
			if len(uriTuple) < 2 {
				return fmt.Errorf("unable to parse URI fo Add field: %v - you may be missing the port number?", value)
			} else if uriTuple[0] == "" || uriTuple[1] == "" {
				return fmt.Errorf("unable to parse URI fo Add field: %v", value)
			}
			updateSQL.Write([]byte("update relays set (internal_ip, internal_ip_port) = ($1, $2) "))
			updateSQL.Write([]byte("where id=$3"))
			args = append(args, uriTuple[0], uriTuple[1], relay.DatabaseID)

			addr, err := net.ResolveUDPAddr("udp", addrString)
			if err != nil {
				return fmt.Errorf("Error converting relay address %s: %v", addrString, err)
			}
			relay.InternalAddr = *addr
		}

	case "PublicKey":
		publicKey, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string type", value)
		}

		newPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
		if err != nil {
			return fmt.Errorf("PublicKey: failed to encode string public key: %v", err)
		}

		updateSQL.Write([]byte("update relays set public_key=$1 where id=$2"))
		args = append(args, newPublicKey, relay.DatabaseID)
		relay.PublicKey = newPublicKey

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
			return fmt.Errorf("%d is not a valid RelayState value", int64(state))
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

		if startDate == "" {
			updateSQL.Write([]byte("update relays set start_date=null where id=$1"))
			args = append(args, relay.DatabaseID)
			relay.StartDate = time.Time{}
		} else {
			newStartDate, err := time.Parse("January 2, 2006", startDate)
			if err != nil {
				return fmt.Errorf("could not parse `%s` - must be of the form 'January 2, 2006'", startDate)
			}

			updateSQL.Write([]byte("update relays set start_date=$1 where id=$2"))
			args = append(args, startDate, relay.DatabaseID)
			relay.StartDate = newStartDate
		}

	case "EndDate":
		endDate, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		if endDate == "" {
			updateSQL.Write([]byte("update relays set end_date=null where id=$1"))
			args = append(args, relay.DatabaseID)
			relay.EndDate = time.Time{}
		} else {
			newEndDate, err := time.Parse("January 2, 2006", endDate)
			if err != nil {
				return fmt.Errorf("could not parse `%s` - must be of the form 'January 2, 2006'", endDate)
			}

			updateSQL.Write([]byte("update relays set end_date=$1 where id=$2"))
			args = append(args, endDate, relay.DatabaseID)
			relay.EndDate = newEndDate
		}

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

	case "Notes":
		notes, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		if notes == "" {
			updateSQL.Write([]byte("update relays set notes=null where id=$1"))
			args = append(args, relay.DatabaseID)
			relay.Notes = ""
		} else {
			updateSQL.Write([]byte("update relays set notes=$1 where id=$2"))
			args = append(args, notes, relay.DatabaseID)
			relay.Notes = notes
		}

	case "BillingSupplier":
		billingSupplier, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		if billingSupplier == "" {
			updateSQL.Write([]byte("update relays set billing_supplier=null where id=$1"))
			args = append(args, relay.DatabaseID)
			relay.BillingSupplier = ""
		} else {

			sellerDatabaseID := 0
			for _, seller := range db.Sellers() {
				if seller.ID == billingSupplier {
					sellerDatabaseID = int(seller.DatabaseID)
				}
			}

			if sellerDatabaseID == 0 {
				return fmt.Errorf("%s is not a valid seller ID", billingSupplier)
			}

			updateSQL.Write([]byte("update relays set billing_supplier=$1 where id=$2"))
			args = append(args, sellerDatabaseID, relay.DatabaseID)
			relay.BillingSupplier = billingSupplier
		}

	case "Version":
		version, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		if version == "" {
			return fmt.Errorf("relay version must not be an empty string")
		}

		updateSQL.Write([]byte("update relays set relay_version=$1 where id=$2"))
		args = append(args, version, relay.DatabaseID)
		relay.Version = version

	default:
		return fmt.Errorf("field '%v' does not exist on the routing.Relay type", field)

	}

	stmt, err = db.Client.PrepareContext(ctx, updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateRelay SQL", "err", err)
		return err
	}

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
	HexID              string
	Name               string
	PublicIP           sql.NullString
	PublicIPPort       sql.NullInt64
	InternalIP         sql.NullString
	InternalIPPort     sql.NullInt64
	BillingSupplier    sql.NullInt64
	PublicKey          []byte
	NICSpeedMbps       int64
	IncludedBandwithGB int64
	DatacenterID       int64
	ManagementIP       string
	SSHUser            string
	SSHPort            int64
	State              int64
	MaxSessions        int64
	MRC                int64
	Overage            int64
	BWRule             int64
	ContractTerm       int64
	Notes              sql.NullString
	StartDate          sql.NullTime
	EndDate            sql.NullTime
	MachineType        int64
	Version            string
	DatabaseID         int64
}

// AddRelay adds the provided relay to storage and returns an error if the relay could not be added.
func (db *SQL) AddRelay(ctx context.Context, r routing.Relay) error {

	var sqlQuery bytes.Buffer
	var err error

	db.relayMutex.RLock()
	_, ok := db.relays[r.ID]
	db.relayMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "relay", resourceRef: r.Name}
	}

	// Routing.Addr is possibly null during syncRelays (due to removed/renamed
	// relays) but *must* have a value when adding a relay
	publicIP := strings.Split(r.Addr.String(), ":")[0]
	publicIPPort, err := strconv.ParseInt(strings.Split(r.Addr.String(), ":")[1], 10, 64)
	if err != nil {
		return fmt.Errorf("unable to convert PublicIP Port %s to int: %v", strings.Split(r.Addr.String(), ":")[1], err)
	}
	rid := crypto.HashID(r.Addr.String())

	var internalIP sql.NullString
	var internalIPPort sql.NullInt64
	if r.InternalAddr.String() != "" && r.InternalAddr.String() != ":0" {
		internalIP.String = strings.Split(r.InternalAddr.String(), ":")[0]
		internalIP.Valid = true
		internalIPPort.Int64, err = strconv.ParseInt(strings.Split(r.InternalAddr.String(), ":")[1], 10, 64)
		internalIPPort.Valid = true
		if err != nil {
			return fmt.Errorf("unable to convert InternalIP Port %s to int: %v", strings.Split(r.InternalAddr.String(), ":")[1], err)
		}
	} else {
		internalIP = sql.NullString{}
		internalIPPort = sql.NullInt64{}
	}

	var startDate sql.NullTime
	if !r.StartDate.IsZero() {
		startDate.Time = r.StartDate
		startDate.Valid = true
	} else {
		startDate = sql.NullTime{}
	}

	var endDate sql.NullTime
	if !r.EndDate.IsZero() {
		endDate.Time = r.EndDate
		endDate.Valid = true
	} else {
		endDate = sql.NullTime{}
	}

	var billingSupplier sql.NullInt64
	if r.BillingSupplier != "" {
		supplier, err := db.Seller(r.BillingSupplier)
		if err != nil {
			return fmt.Errorf("Seller %s does not exist %v", r.BillingSupplier, err)
		}
		billingSupplier.Valid = true
		billingSupplier.Int64 = supplier.DatabaseID
	} else {
		billingSupplier.Valid = false
	}

	nullablePublicIP := sql.NullString{
		Valid:  true,
		String: publicIP,
	}

	nullablePublicIPPort := sql.NullInt64{
		Valid: true,
		Int64: publicIPPort,
	}

	nullableNotes := sql.NullString{
		Valid: false,
	}

	if r.Notes != "" {
		nullableNotes.Valid = true
		nullableNotes.String = r.Notes
	}

	// field is not null but we also don't want an empty string
	if r.Version == "" {
		return fmt.Errorf("relay version can not be an empty string and must be a valid value (e.g. '2.0.6')")
	}

	relay := sqlRelay{
		Name:               r.Name,
		HexID:              fmt.Sprintf("%016x", rid),
		PublicIP:           nullablePublicIP,
		PublicIPPort:       nullablePublicIPPort,
		InternalIP:         internalIP,
		InternalIPPort:     internalIPPort,
		PublicKey:          r.PublicKey,
		NICSpeedMbps:       int64(r.NICSpeedMbps),
		IncludedBandwithGB: int64(r.IncludedBandwidthGB),
		DatacenterID:       r.Datacenter.DatabaseID,
		ManagementIP:       r.ManagementAddr,
		BillingSupplier:    billingSupplier,
		SSHUser:            r.SSHUser,
		SSHPort:            r.SSHPort,
		State:              int64(r.State),
		MaxSessions:        int64(r.MaxSessions),
		MRC:                int64(r.MRC),
		Overage:            int64(r.Overage),
		BWRule:             int64(r.BWRule),
		ContractTerm:       int64(r.ContractTerm),
		StartDate:          startDate,
		EndDate:            endDate,
		MachineType:        int64(r.Type),
		Notes:              nullableNotes,
		Version:            r.Version,
	}

	sqlQuery.Write([]byte("insert into relays ("))
	sqlQuery.Write([]byte("hex_id, contract_term, display_name, end_date, included_bandwidth_gb, "))
	sqlQuery.Write([]byte("management_ip, max_sessions, mrc, overage, port_speed, public_ip, "))
	sqlQuery.Write([]byte("public_ip_port, public_key, ssh_port, ssh_user, start_date, "))
	sqlQuery.Write([]byte("bw_billing_rule, datacenter, machine_type, relay_state, "))
	sqlQuery.Write([]byte("internal_ip, internal_ip_port, notes, billing_supplier, relay_version "))
	sqlQuery.Write([]byte(") values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, "))
	sqlQuery.Write([]byte("$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)"))

	stmt, err := db.Client.PrepareContext(ctx, sqlQuery.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddRelay SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
		relay.HexID,
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
		relay.Notes,
		relay.BillingSupplier,
		relay.Version,
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

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// SetRelay updates the relay in storage with the provided copy and returns an
// error if the relay could not be updated.
// TODO: chopping block
func (db *SQL) SetRelay(ctx context.Context, r routing.Relay) error {

	var sqlQuery bytes.Buffer
	var err error

	db.relayMutex.RLock()
	_, ok := db.relays[r.ID]
	db.relayMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "relay", resourceRef: fmt.Sprintf("%016x", r.ID)}
	}

	var publicIP sql.NullString
	var publicIPPort sql.NullInt64

	if r.Addr.String() != ":0" && r.Addr.String() != "" {
		publicIP.String = strings.Split(r.Addr.String(), ":")[0]
		publicIP.Valid = true
		publicIPPort.Int64, err = strconv.ParseInt(strings.Split(r.Addr.String(), ":")[1], 10, 64)
		publicIPPort.Valid = true
		if err != nil {
			return fmt.Errorf("Unable to convert InternalIP Port %s to int: %v", strings.Split(r.Addr.String(), ":")[1], err)
		}
	} else {
		publicIP = sql.NullString{}
		publicIPPort = sql.NullInt64{}
	}

	var internalIP sql.NullString
	var internalIPPort sql.NullInt64
	if r.InternalAddr.String() != ":0" && r.InternalAddr.String() != "" {
		internalIP.String = strings.Split(r.InternalAddr.String(), ":")[0]
		internalIP.Valid = true
		internalIPPort.Int64, err = strconv.ParseInt(strings.Split(r.InternalAddr.String(), ":")[1], 10, 64)
		internalIPPort.Valid = true
		if err != nil {
			return fmt.Errorf("Unable to convert InternalIP Port %s to int: %v", strings.Split(r.InternalAddr.String(), ":")[1], err)
		}
	} else {
		internalIP = sql.NullString{}
		internalIPPort = sql.NullInt64{}
	}

	var startDate sql.NullTime
	if !r.StartDate.IsZero() {
		startDate.Time = r.StartDate
		startDate.Valid = true
	} else {
		startDate = sql.NullTime{}
	}

	var endDate sql.NullTime
	if !r.EndDate.IsZero() {
		endDate.Time = r.EndDate
		endDate.Valid = true
	} else {
		endDate = sql.NullTime{}
	}

	relay := sqlRelay{
		Name:               r.Name,
		PublicIP:           publicIP,
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
		StartDate:          startDate,
		EndDate:            endDate,
		MachineType:        int64(r.Type),
	}

	sqlQuery.Write([]byte("update relays set ("))
	sqlQuery.Write([]byte("contract_term, display_name, end_date, included_bandwidth_gb, "))
	sqlQuery.Write([]byte("management_ip, max_sessions, mrc, overage, port_speed, public_ip, "))
	sqlQuery.Write([]byte("public_ip_port, public_key, ssh_port, ssh_user, start_date, "))
	sqlQuery.Write([]byte("bw_billing_rule, datacenter, machine_type, relay_state, internal_ip, internal_ip_port "))
	sqlQuery.Write([]byte(") = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, "))
	sqlQuery.Write([]byte("$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) where id = $22"))

	stmt, err := db.Client.PrepareContext(ctx, sqlQuery.String())
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
func (db *SQL) SetDatacenter(ctx context.Context, d routing.Datacenter) error {

	var sql bytes.Buffer

	db.datacenterMutex.RLock()
	_, ok := db.datacenters[d.ID]
	db.datacenterMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%016x", d.ID)}
	}

	dc := sqlDatacenter{
		Name:      d.Name,
		Latitude:  d.Location.Latitude,
		Longitude: d.Location.Longitude,
		SellerID:  d.SellerID,
	}

	sql.Write([]byte("update datacenters set ("))
	sql.Write([]byte("display_name, latitude, longitude, "))
	sql.Write([]byte("seller_id ) = ($1, $2, $3, $4) where id = $5"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing SetDatacenter SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(dc.Name,
		dc.Latitude,
		dc.Longitude,
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

// GetDatacenterMapsForBuyer returns a map of datacenter aliases in use for a given
// (internally generated) buyerID. The map is indexed by the datacenter ID. Returns
// an empty map if there are no aliases for that buyerID.
func (db *SQL) GetDatacenterMapsForBuyer(ephemeralBuyerID uint64) map[uint64]routing.DatacenterMap {
	db.datacenterMapMutex.RLock()
	defer db.datacenterMapMutex.RUnlock()

	// buyer can have multiple dc maps but only one alias per datacenter
	var dcs = make(map[uint64]routing.DatacenterMap)
	for _, dc := range db.datacenterMaps {
		if dc.BuyerID == ephemeralBuyerID {
			dcs[dc.DatacenterID] = dc
		}
	}

	return dcs
}

// AddDatacenterMap adds a new datacenter map for the given buyer and datacenter IDs
func (db *SQL) AddDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {

	var sql bytes.Buffer

	ephemeralBuyerID := dcMap.BuyerID
	buyerID := db.buyerIDs[ephemeralBuyerID]

	dcID := dcMap.DatacenterID

	buyer, ok := db.buyers[uint64(buyerID)]
	if !ok {
		fmt.Printf("buyer does not exist: %016x\n", dcMap.BuyerID)
		return &DoesNotExistError{resourceType: "BuyerID", resourceRef: dcMap.BuyerID}
	}

	datacenter, ok := db.datacenters[dcID]
	if !ok {
		fmt.Printf("datacenter does not exist: %016x\n", dcMap.DatacenterID)
		return &DoesNotExistError{resourceType: "DatacenterID", resourceRef: dcMap.DatacenterID}
	}

	sql.Write([]byte("insert into datacenter_maps (buyer_id, datacenter_id) "))
	sql.Write([]byte("values ($1, $2)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddDatacenterMap SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
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

func (db *SQL) UpdateDatacenterMap(ctx context.Context, ephemeralBuyerID uint64, datacenterID uint64, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt
	var err error

	dcmID := crypto.HashID(fmt.Sprintf("%016x", ephemeralBuyerID) + fmt.Sprintf("%016x", datacenterID))
	workingDatacenterMap, ok := db.datacenterMaps[dcmID]
	if !ok {
		return fmt.Errorf("Datacenter map for buyerID %016x, datacenterID %016x does not exist", ephemeralBuyerID, datacenterID)
	}

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])
	// if the dcMap exists then the buyer and datacenter IDs are legit
	originalDatacenter := db.datacenters[datacenterID]
	originalBuyer := db.buyers[buyerID]

	switch field {
	case "HexDatacenterID":
		hexDatacenterID, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		newDatacenterID, err := strconv.ParseUint(hexDatacenterID, 16, 64)
		if err != nil {
			return fmt.Errorf("Could not parse hexDatacenterID: %v", value)
		}

		newDatacenter := db.datacenters[newDatacenterID]

		updateSQL.Write([]byte("update datacenter_maps set datacenter_id=$1 where datacenter_id=$2 and buyer_id=$3"))
		args = append(args, newDatacenter.DatabaseID, originalDatacenter.DatabaseID, originalBuyer.DatabaseID)
		workingDatacenterMap.DatacenterID = newDatacenterID

		// changing the datacenter ID in the alias changes the datacenter map ID so
		// delete the old one and add the new one
		db.datacenterMapsMutex.Lock()
		delete(db.datacenterMaps, dcmID)
		dcmID = crypto.HashID(fmt.Sprintf("%x", ephemeralBuyerID) + fmt.Sprintf("%x", newDatacenterID))
		db.datacenterMaps[dcmID] = workingDatacenterMap
		db.datacenterMapsMutex.Unlock()

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.DatacenterMap type", field)

	}

	stmt, err = db.Client.PrepareContext(ctx, updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateDatacenterMap SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(args...)
	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying datacenter map record", "err", err)
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
			id := crypto.HashID(fmt.Sprintf("%x", dc.BuyerID) + fmt.Sprintf("%x", dc.DatacenterID))
			dcs[id] = dc
		}
	}

	return dcs
}

// RemoveDatacenterMap removes an entry from the DatacenterMaps table
func (db *SQL) RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {
	var sql bytes.Buffer

	id := crypto.HashID(fmt.Sprintf("%016x", dcMap.BuyerID) + fmt.Sprintf("%016x", dcMap.DatacenterID))

	db.datacenterMapMutex.RLock()
	dcMap, ok := db.datacenterMaps[id]
	db.datacenterMapMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "datacenter map", resourceRef: fmt.Sprintf("%016x", id)}
	}

	buyerID := db.buyerIDs[dcMap.BuyerID]
	buyer := db.buyers[uint64(buyerID)]
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

	return nil
}

type sqlDatacenter struct {
	ID        int64
	HexID     string
	Name      string
	Latitude  float32
	Longitude float32
	SellerID  int64
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

	did := fmt.Sprintf("%016x", crypto.HashID(datacenter.Name))
	dc := sqlDatacenter{
		Name:      datacenter.Name,
		HexID:     did,
		Latitude:  datacenter.Location.Latitude,
		Longitude: datacenter.Location.Longitude,
		SellerID:  datacenter.SellerID,
	}

	sql.Write([]byte("insert into datacenters ("))
	sql.Write([]byte("display_name, hex_id, latitude, longitude, "))
	sql.Write([]byte("seller_id ) values ($1, $2, $3, $4, $5)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddDatacenter SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
		dc.Name,
		dc.HexID,
		dc.Latitude,
		dc.Longitude,
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
func (db *SQL) RouteShader(ephemeralBuyerID uint64) (core.RouteShader, error) {
	db.routeShaderMutex.RLock()
	defer db.routeShaderMutex.RUnlock()

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])

	routeShader, found := db.routeShaders[buyerID]
	if !found {
		return core.RouteShader{}, &DoesNotExistError{resourceType: "route shader", resourceRef: fmt.Sprintf("%x", buyerID)}
	}

	return routeShader, nil
}

// InternalConfig returns the InternalConfig entry for the specified buyer
func (db *SQL) InternalConfig(ephemeralBuyerID uint64) (core.InternalConfig, error) {
	db.internalConfigMutex.RLock()
	defer db.internalConfigMutex.RUnlock()

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])

	internalConfig, found := db.internalConfigs[buyerID]
	if !found {
		return core.InternalConfig{}, &DoesNotExistError{resourceType: "internal config", resourceRef: fmt.Sprintf("%x", buyerID)}
	}

	return internalConfig, nil

}

// AddInternalConfig adds an InternalConfig for the specified buyer
func (db *SQL) AddInternalConfig(ctx context.Context, ic core.InternalConfig, ephemeralBuyerID uint64) error {

	var sql bytes.Buffer

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])

	db.internalConfigMutex.RLock()
	_, ok := db.internalConfigs[buyerID]
	db.internalConfigMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "InternalConfig", resourceRef: buyerID}
	}

	db.buyerMutex.RLock()
	buyer, ok := db.buyers[buyerID]
	db.buyerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "Buyer", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	internalConfig := sqlInternalConfig{
		RouteSelectThreshold:           int64(ic.RouteSelectThreshold),
		RouteSwitchThreshold:           int64(ic.RouteSwitchThreshold),
		MaxLatencyTradeOff:             int64(ic.MaxLatencyTradeOff),
		RTTVetoDefault:                 int64(ic.RTTVeto_Default),
		RTTVetoPacketLoss:              int64(ic.RTTVeto_PacketLoss),
		RTTVetoMultipath:               int64(ic.RTTVeto_Multipath),
		MultipathOverloadThreshold:     int64(ic.MultipathOverloadThreshold),
		TryBeforeYouBuy:                ic.TryBeforeYouBuy,
		ForceNext:                      ic.ForceNext,
		LargeCustomer:                  ic.LargeCustomer,
		Uncommitted:                    ic.Uncommitted,
		MaxRTT:                         int64(ic.MaxRTT),
		HighFrequencyPings:             ic.HighFrequencyPings,
		RouteDiversity:                 int64(ic.RouteDiversity),
		MultipathThreshold:             int64(ic.MultipathThreshold),
		EnableVanityMetrics:            ic.EnableVanityMetrics,
		ReducePacketLossMinSliceNumber: int64(ic.ReducePacketLossMinSliceNumber),
	}

	sql.Write([]byte("insert into rs_internal_configs "))
	sql.Write([]byte("(max_latency_tradeoff, max_rtt, multipath_overload_threshold, "))
	sql.Write([]byte("route_switch_threshold, route_select_threshold, rtt_veto_default, "))
	sql.Write([]byte("rtt_veto_multipath, rtt_veto_packetloss, try_before_you_buy, force_next, "))
	sql.Write([]byte("large_customer, is_uncommitted, high_frequency_pings, route_diversity, "))
	sql.Write([]byte("multipath_threshold, enable_vanity_metrics, reduce_pl_min_slice_number, buyer_id) "))
	sql.Write([]byte("values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)"))

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
		internalConfig.HighFrequencyPings,
		internalConfig.RouteDiversity,
		internalConfig.MultipathThreshold,
		internalConfig.EnableVanityMetrics,
		internalConfig.ReducePacketLossMinSliceNumber,
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

	buyer.InternalConfig = ic
	db.buyerMutex.Lock()
	db.buyers[buyerID] = buyer
	db.buyerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

func (db *SQL) RemoveInternalConfig(ctx context.Context, ephemeralBuyerID uint64) error {
	var sql bytes.Buffer

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])

	db.internalConfigMutex.RLock()
	_, ok := db.internalConfigs[buyerID]
	db.internalConfigMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "InternalConfig", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	buyer, ok := db.buyers[buyerID]
	if !ok {
		return &DoesNotExistError{resourceType: "InternalConfig", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	sql.Write([]byte("delete from rs_internal_configs where buyer_id = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveInternalConfig SQL", "err", err)
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

func (db *SQL) UpdateInternalConfig(ctx context.Context, ephemeralBuyerID uint64, field string, value interface{}) error {

	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt
	var err error

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])
	ic, ok := db.internalConfigs[buyerID]
	if !ok {
		return &DoesNotExistError{resourceType: "internal config", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	db.buyerMutex.RLock()
	buyer, ok := db.buyers[buyerID]
	db.buyerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "Buyer", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	switch field {
	case "RouteSelectThreshold":
		routeSelectThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteSelectThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set route_select_threshold=$1 where buyer_id=$2"))
		args = append(args, routeSelectThreshold, buyer.DatabaseID)
		ic.RouteSelectThreshold = routeSelectThreshold
	case "RouteSwitchThreshold":
		routeSwitchThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteSwitchThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set route_switch_threshold=$1 where buyer_id=$2"))
		args = append(args, routeSwitchThreshold, buyer.DatabaseID)
		ic.RouteSwitchThreshold = routeSwitchThreshold
	case "MaxLatencyTradeOff":
		maxLatencyTradeOff, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MaxLatencyTradeOff: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set max_latency_tradeoff=$1 where buyer_id=$2"))
		args = append(args, maxLatencyTradeOff, buyer.DatabaseID)
		ic.MaxLatencyTradeOff = maxLatencyTradeOff
	case "RTTVeto_Default":
		rttVetoDefault, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_Default: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_default=$1 where buyer_id=$2"))
		args = append(args, rttVetoDefault, buyer.DatabaseID)
		ic.RTTVeto_Default = rttVetoDefault
	case "RTTVeto_PacketLoss":
		rttVetoPacketLoss, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_PacketLoss: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_packetloss=$1 where buyer_id=$2"))
		args = append(args, rttVetoPacketLoss, buyer.DatabaseID)
		ic.RTTVeto_PacketLoss = rttVetoPacketLoss
	case "RTTVeto_Multipath":
		rttVetoMultipath, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_Multipath: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_multipath=$1 where buyer_id=$2"))
		args = append(args, rttVetoMultipath, buyer.DatabaseID)
		ic.RTTVeto_Multipath = rttVetoMultipath
	case "MultipathOverloadThreshold":
		multipathOverloadThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MultipathOverloadThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set multipath_overload_threshold=$1 where buyer_id=$2"))
		args = append(args, multipathOverloadThreshold, buyer.DatabaseID)
		ic.MultipathOverloadThreshold = multipathOverloadThreshold
	case "TryBeforeYouBuy":
		tryBeforeYouBuy, ok := value.(bool)
		if !ok {
			return fmt.Errorf("TryBeforeYouBuy: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set try_before_you_buy=$1 where buyer_id=$2"))
		args = append(args, tryBeforeYouBuy, buyer.DatabaseID)
		ic.TryBeforeYouBuy = tryBeforeYouBuy
	case "ForceNext":
		forceNext, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ForceNext: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set force_next=$1 where buyer_id=$2"))
		args = append(args, forceNext, buyer.DatabaseID)
		ic.ForceNext = forceNext
	case "LargeCustomer":
		largeCustomer, ok := value.(bool)
		if !ok {
			return fmt.Errorf("LargeCustomer: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set large_customer=$1 where buyer_id=$2"))
		args = append(args, largeCustomer, buyer.DatabaseID)
		ic.LargeCustomer = largeCustomer
	case "Uncommitted":
		uncommitted, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Uncommitted: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set is_uncommitted=$1 where buyer_id=$2"))
		args = append(args, uncommitted, buyer.DatabaseID)
		ic.Uncommitted = uncommitted
	case "HighFrequencyPings":
		highFrequencyPings, ok := value.(bool)
		if !ok {
			return fmt.Errorf("HighFrequencyPings: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set high_frequency_pings=$1 where buyer_id=$2"))
		args = append(args, highFrequencyPings, buyer.DatabaseID)
		ic.HighFrequencyPings = highFrequencyPings
	case "MaxRTT":
		maxRTT, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MaxRTT: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set max_rtt=$1 where buyer_id=$2"))
		args = append(args, maxRTT, buyer.DatabaseID)
		ic.MaxRTT = maxRTT
	case "RouteDiversity":
		routeDiversity, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteDiversity: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set route_diversity=$1 where buyer_id=$2"))
		args = append(args, routeDiversity, buyer.DatabaseID)
		ic.RouteDiversity = routeDiversity
	case "MultipathThreshold":
		multipathThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MultipathThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set multipath_threshold=$1 where buyer_id=$2"))
		args = append(args, multipathThreshold, buyer.DatabaseID)
		ic.MultipathThreshold = multipathThreshold
	case "EnableVanityMetrics":
		enableVanityMetrics, ok := value.(bool)
		if !ok {
			return fmt.Errorf("EnableVanityMetrics: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set enable_vanity_metrics=$1 where buyer_id=$2"))
		args = append(args, enableVanityMetrics, buyer.DatabaseID)
		ic.EnableVanityMetrics = enableVanityMetrics
	case "ReducePacketLossMinSliceNumber":
		reducePacketLossMinSliceNumber, ok := value.(int32)
		if !ok {
			return fmt.Errorf("ReducePacketLossMinSliceNumber: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set reduce_pl_min_slice_number=$1 where buyer_id=$2"))
		args = append(args, reducePacketLossMinSliceNumber, buyer.DatabaseID)
		ic.ReducePacketLossMinSliceNumber = reducePacketLossMinSliceNumber

	default:
		return fmt.Errorf("Field '%v' does not exist on the InternalConfig type", field)
	}

	stmt, err = db.Client.PrepareContext(ctx, updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateInternalConfig SQL", "err", err)
		return err
	}

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

	db.IncrementSequenceNumber(ctx)

	return nil
}

func (db *SQL) AddRouteShader(ctx context.Context, rs core.RouteShader, ephemeralBuyerID uint64) error {

	var sql bytes.Buffer

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])

	db.routeShaderMutex.RLock()
	_, ok := db.routeShaders[buyerID]
	db.routeShaderMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "RouteShader", resourceRef: buyerID}
	}

	db.buyerMutex.RLock()
	buyer, ok := db.buyers[buyerID]
	db.buyerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "Buyer", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	routeShader := sqlRouteShader{
		ABTest:                    rs.ABTest,
		AcceptableLatency:         int64(rs.AcceptableLatency),
		AcceptablePacketLoss:      float64(rs.AcceptablePacketLoss),
		BandwidthEnvelopeDownKbps: int64(rs.BandwidthEnvelopeDownKbps),
		BandwidthEnvelopeUpKbps:   int64(rs.BandwidthEnvelopeUpKbps),
		DisableNetworkNext:        rs.DisableNetworkNext,
		LatencyThreshold:          int64(rs.LatencyThreshold),
		Multipath:                 rs.Multipath,
		ProMode:                   rs.ProMode,
		ReduceLatency:             rs.ReduceLatency,
		ReducePacketLoss:          rs.ReducePacketLoss,
		ReduceJitter:              rs.ReduceJitter,
		SelectionPercent:          int64(rs.SelectionPercent),
		PacketLossSustained:       float64(rs.PacketLossSustained),
	}

	sql.Write([]byte("insert into route_shaders ("))
	sql.Write([]byte("ab_test, acceptable_latency, acceptable_packet_loss, bw_envelope_down_kbps, "))
	sql.Write([]byte("bw_envelope_up_kbps, disable_network_next, latency_threshold, multipath, "))
	sql.Write([]byte("pro_mode, reduce_latency, reduce_packet_loss, reduce_jitter, selection_percent, packet_loss_sustained, buyer_id"))
	sql.Write([]byte(") values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddInternalConfig SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
		routeShader.ABTest,
		routeShader.AcceptableLatency,
		routeShader.AcceptablePacketLoss,
		routeShader.BandwidthEnvelopeDownKbps,
		routeShader.BandwidthEnvelopeUpKbps,
		routeShader.DisableNetworkNext,
		routeShader.LatencyThreshold,
		routeShader.Multipath,
		routeShader.ProMode,
		routeShader.ReduceLatency,
		routeShader.ReducePacketLoss,
		routeShader.ReduceJitter,
		routeShader.SelectionPercent,
		routeShader.PacketLossSustained,
		buyer.DatabaseID,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding route shader", "err", err)
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

	db.routeShaderMutex.Lock()
	db.routeShaders[buyerID] = rs
	db.routeShaderMutex.Unlock()

	buyer.RouteShader = rs
	db.buyerMutex.Lock()
	db.buyers[buyer.ID] = buyer
	db.buyerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil

}

func (db *SQL) UpdateRouteShader(ctx context.Context, ephemeralBuyerID uint64, field string, value interface{}) error {

	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt
	var err error

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])
	rs, ok := db.routeShaders[buyerID]
	if !ok {
		return &DoesNotExistError{resourceType: "route shader", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	db.buyerMutex.RLock()
	buyer, ok := db.buyers[buyerID]
	db.buyerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "Buyer", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	switch field {
	case "ABTest":
		abTest, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ABTest: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set ab_test=$1 where buyer_id=$2"))
		args = append(args, abTest, buyer.DatabaseID)
		rs.ABTest = abTest
	case "AcceptableLatency":
		acceptableLatency, ok := value.(int32)
		if !ok {
			return fmt.Errorf("AcceptableLatency: %v is not a valid int32 type ( %T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set acceptable_latency=$1 where buyer_id=$2"))
		args = append(args, acceptableLatency, buyer.DatabaseID)
		rs.AcceptableLatency = acceptableLatency
	case "AcceptablePacketLoss":
		acceptablePacketLoss, ok := value.(float32)
		if !ok {
			return fmt.Errorf("AcceptablePacketLoss: %v is not a valid float32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set acceptable_packet_loss=$1 where buyer_id=$2"))
		args = append(args, acceptablePacketLoss, buyer.DatabaseID)
		rs.AcceptablePacketLoss = acceptablePacketLoss
	case "BandwidthEnvelopeDownKbps":
		bandwidthEnvelopeDownKbps, ok := value.(int32)
		if !ok {
			return fmt.Errorf("BandwidthEnvelopeDownKbps: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set bw_envelope_down_kbps=$1 where buyer_id=$2"))
		args = append(args, bandwidthEnvelopeDownKbps, buyer.DatabaseID)
		rs.BandwidthEnvelopeDownKbps = bandwidthEnvelopeDownKbps
	case "BandwidthEnvelopeUpKbps":
		bandwidthEnvelopeUpKbps, ok := value.(int32)
		if !ok {
			return fmt.Errorf("BandwidthEnvelopeUpKbps: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set bw_envelope_up_kbps=$1 where buyer_id=$2"))
		args = append(args, bandwidthEnvelopeUpKbps, buyer.DatabaseID)
		rs.BandwidthEnvelopeUpKbps = bandwidthEnvelopeUpKbps
	case "DisableNetworkNext":
		disableNetworkNext, ok := value.(bool)
		if !ok {
			return fmt.Errorf("DisableNetworkNext: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set disable_network_next=$1 where buyer_id=$2"))
		args = append(args, disableNetworkNext, buyer.DatabaseID)
		rs.DisableNetworkNext = disableNetworkNext
	case "LatencyThreshold":
		latencyThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("LatencyThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set latency_threshold=$1 where buyer_id=$2"))
		args = append(args, latencyThreshold, buyer.DatabaseID)
		rs.LatencyThreshold = latencyThreshold
	case "Multipath":
		multipath, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Multipath: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set multipath=$1 where buyer_id=$2"))
		args = append(args, multipath, buyer.DatabaseID)
		rs.Multipath = multipath
	case "ProMode":
		proMode, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ProMode: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set pro_mode=$1 where buyer_id=$2"))
		args = append(args, proMode, buyer.DatabaseID)
		rs.ProMode = proMode
	case "ReduceLatency":
		reduceLatency, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReduceLatency: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set reduce_latency=$1 where buyer_id=$2"))
		args = append(args, reduceLatency, buyer.DatabaseID)
		rs.ReduceLatency = reduceLatency
	case "ReducePacketLoss":
		reducePacketLoss, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReducePacketLoss: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set reduce_packet_loss=$1 where buyer_id=$2"))
		args = append(args, reducePacketLoss, buyer.DatabaseID)
		rs.ReducePacketLoss = reducePacketLoss
	case "ReduceJitter":
		reduceJitter, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReduceJitter: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set reduce_jitter=$1 where buyer_id=$2"))
		args = append(args, reduceJitter, buyer.DatabaseID)
		rs.ReduceJitter = reduceJitter
	case "SelectionPercent":
		selectionPercent, ok := value.(int)
		if !ok {
			return fmt.Errorf("SelectionPercent: %v is not a valid int type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set selection_percent=$1 where buyer_id=$2"))
		args = append(args, selectionPercent, buyer.DatabaseID)
		rs.SelectionPercent = selectionPercent
	case "PacketLossSustained":
		packetLossSustained, ok := value.(float32)
		if !ok {
			return fmt.Errorf("PacketLossSustained: %v is not a valid float type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set packet_loss_sustained=$1 where buyer_id=$2"))
		args = append(args, packetLossSustained, buyer.DatabaseID)
		rs.PacketLossSustained = packetLossSustained
	default:
		return fmt.Errorf("Field '%v' does not exist on the RouteShader type", field)

	}

	stmt, err = db.Client.PrepareContext(ctx, updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateRouteShader SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(args...)
	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying route_shader record", "err", err)
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

	db.routeShaderMutex.Lock()
	db.routeShaders[buyerID] = rs
	db.routeShaderMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil

}

func (db *SQL) RemoveRouteShader(ctx context.Context, ephemeralBuyerID uint64) error {
	var sql bytes.Buffer

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])
	db.routeShaderMutex.RLock()
	_, ok := db.routeShaders[buyerID]
	db.routeShaderMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "RouteShader", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	buyer, ok := db.buyers[buyerID]
	if !ok {
		return &DoesNotExistError{resourceType: "RouteShader", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	sql.Write([]byte("delete from route_shaders where buyer_id = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveRouteShader SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(buyer.DatabaseID)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing route shader", "err", err)
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

	db.routeShaderMutex.Lock()
	delete(db.routeShaders, buyerID)
	db.routeShaderMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

// AddBannedUser adds a user to the banned_user table
func (db *SQL) AddBannedUser(ctx context.Context, ephemeralBuyerID uint64, userID uint64) error {

	var sql bytes.Buffer

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])

	db.buyerMutex.RLock()
	buyer, ok := db.buyers[buyerID]
	db.buyerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "Buyer", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	db.bannedUserMutex.RLock()
	_, ok = db.bannedUsers[buyerID][userID]
	db.bannedUserMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "banned user", resourceRef: fmt.Sprintf("%016x", userID)}
	}

	sql.Write([]byte("insert into banned_users (user_id, buyer_id) values ($1, $2)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddBannedUser SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(int64(userID), buyer.DatabaseID)
	if err != nil {
		level.Error(db.Logger).Log("during", "error adding banned user", "err", err)
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

	db.bannedUserMutex.Lock()
	if _, ok := db.bannedUsers[buyerID]; !ok {
		bannedUsers := make(map[uint64]bool)
		db.bannedUsers[buyerID] = bannedUsers
	}
	db.bannedUsers[buyerID][userID] = true
	db.bannedUserMutex.Unlock()

	// we need to handle the case where the buyer is using the default
	// route shader (and therefore does not have an entry in db.routeShaders)
	var rs core.RouteShader
	db.routeShaderMutex.Lock()
	if rs, ok = db.routeShaders[buyerID]; !ok {
		rs = core.NewRouteShader()
	}

	if len(rs.BannedUsers) == 0 {
		rs.BannedUsers = make(map[uint64]bool)
	}
	rs.BannedUsers[userID] = true
	db.routeShaders[buyerID] = rs
	db.routeShaderMutex.Unlock()

	buyer.RouteShader = rs
	db.buyerMutex.Lock()
	db.buyers[buyerID] = buyer
	db.buyerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil

}

// RemoveBannedUser removes a user from the banned_user table
func (db *SQL) RemoveBannedUser(ctx context.Context, ephemeralBuyerID uint64, userID uint64) error {

	var sql bytes.Buffer

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])
	db.buyerMutex.RLock()
	buyer, ok := db.buyers[buyerID]
	db.buyerMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "Buyer", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	db.routeShaderMutex.RLock()
	rs, ok := db.routeShaders[buyerID]
	db.routeShaderMutex.RUnlock()

	if !ok {
		return &DoesNotExistError{resourceType: "RouteShader", resourceRef: fmt.Sprintf("%016x", buyerID)}
	} else if !rs.BannedUsers[userID] {
		return &DoesNotExistError{resourceType: "Banned User", resourceRef: fmt.Sprintf("%016x", userID)}
	}

	sql.Write([]byte("delete from banned_users where user_id = $1 and buyer_id = $2"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddInternalConfig SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(int64(userID), buyer.DatabaseID)
	if err != nil {
		level.Error(db.Logger).Log("during", "error removing banned user", "err", err)
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

	db.bannedUserMutex.Lock()
	delete(db.bannedUsers[buyerID], userID)
	db.bannedUserMutex.Unlock()

	delete(rs.BannedUsers, userID)
	db.routeShaderMutex.Lock()
	db.routeShaders[buyerID] = rs
	db.routeShaderMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil

}

// BannedUsers returns the set of banned users for the specified buyer ID.
func (db *SQL) BannedUsers(ephemeralBuyerID uint64) (map[uint64]bool, error) {

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])

	db.bannedUserMutex.RLock()
	bannedUsers, found := db.bannedUsers[buyerID]
	db.bannedUserMutex.RUnlock()

	if !found {
		return map[uint64]bool{}, &DoesNotExistError{resourceType: "banned user", resourceRef: fmt.Sprintf("%x", buyerID)}
	}

	return bannedUsers, nil
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

func (db *SQL) UpdateBuyer(ctx context.Context, ephemeralBuyerID uint64, field string, value interface{}) error {

	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt
	var err error

	buyerID := uint64(db.buyerIDs[ephemeralBuyerID])

	buyer, ok := db.buyers[buyerID]
	if !ok {
		return &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	switch field {
	case "Live":
		live, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Live: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set is_live_customer=$1 where id=$2"))
		args = append(args, live, buyer.DatabaseID)
		buyer.Live = live
	case "Debug":
		debug, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Debug: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set debug=$1 where id=$2"))
		args = append(args, debug, buyer.DatabaseID)
		buyer.Debug = debug
	case "ShortName":
		shortName, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update buyers set short_name=$1 where id=$2"))
		args = append(args, shortName, buyer.DatabaseID)
		buyer.ShortName = shortName

	case "PublicKey":
		pubKey, ok := value.(string)
		if !ok {
			return fmt.Errorf("PublicKey: %v is not a valid string type (%T)", value, value)
		}

		// Changing the public key also requires changing the ID field and fixing any
		// extant datacenter maps for this buyer
		newPublicKey, err := base64.StdEncoding.DecodeString(pubKey)
		if err != nil {
			return fmt.Errorf("PublicKey: failed to encode string public key: %v", err)
		}

		if len(newPublicKey) != crypto.KeySize+8 {
			return fmt.Errorf("PublicKey: public key is not the correct length: %d", len(newPublicKey))
		}

		newBuyerID := binary.LittleEndian.Uint64(newPublicKey[:8])
		updateSQL.Write([]byte("update buyers set public_key=$1, sdk_generated_id=$2 where id=$3"))
		args = append(args, newPublicKey[8:], int64(newBuyerID), buyer.DatabaseID)

		buyer.ID = newBuyerID
		buyer.PublicKey = newPublicKey[8:]

		db.buyerIDsMutex.Lock()
		delete(db.buyerIDs, ephemeralBuyerID)
		db.buyerIDs[newBuyerID] = buyer.DatabaseID
		db.buyerIDsMutex.Unlock()

		// Fix existing datacenter maps
		dcMaps := db.GetDatacenterMapsForBuyer(ephemeralBuyerID)
		if len(dcMaps) > 0 {
			for key, dcMap := range dcMaps {
				dcMap.BuyerID = newBuyerID
				dcmID := crypto.HashID(fmt.Sprintf("%x", newBuyerID) + fmt.Sprintf("%x", dcMap.DatacenterID))

				db.datacenterMapsMutex.Lock()
				delete(db.datacenterMaps, key)
				db.datacenterMaps[dcmID] = dcMap
				db.datacenterMapsMutex.Unlock()
			}
		}

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Buyer type", field)

	}

	stmt, err = db.Client.PrepareContext(ctx, updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateBuyer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(args...)
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
		level.Error(db.Logger).Log("during", "RowsAffected <> 1")
		return err
	}

	db.buyerMutex.Lock()
	db.buyers[buyerID] = buyer
	db.buyerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

func (db *SQL) UpdateCustomer(ctx context.Context, customerID string, field string, value interface{}) error {

	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt

	customer, err := db.Customer(customerID)
	if err != nil {
		return &DoesNotExistError{resourceType: "customer", resourceRef: fmt.Sprintf("%016x", customerID)}
	}

	switch field {
	case "AutomaticSigninDomains":
		domains, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update customers set automatic_signin_domain=$1 where id=$2"))
		args = append(args, domains, customer.DatabaseID)
		customer.AutomaticSignInDomains = domains
	case "Name":
		name, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update customers set customer_name=$1 where id=$2"))
		args = append(args, name, customer.DatabaseID)
		customer.Name = name

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Customer type", field)

	}

	stmt, err = db.Client.PrepareContext(ctx, updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateCustomer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(args...)
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
	db.customers[customerID] = customer
	db.customerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

func (db *SQL) UpdateSeller(ctx context.Context, sellerID string, field string, value interface{}) error {

	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt

	seller, err := db.Seller(sellerID)
	if err != nil {
		return &DoesNotExistError{resourceType: "seller", resourceRef: fmt.Sprintf("%016x", sellerID)}
	}

	switch field {
	case "ShortName":
		shortName, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update sellers set short_name=$1 where id=$2"))
		args = append(args, shortName, seller.DatabaseID)
		seller.ShortName = shortName
	case "EgressPriceNibblinsPerGB":
		egressPrice, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		egress := routing.DollarsToNibblins(egressPrice)
		updateSQL.Write([]byte("update sellers set public_egress_price=$1 where id=$2"))
		args = append(args, int64(egress), seller.DatabaseID)
		seller.EgressPriceNibblinsPerGB = egress

	case "Secret":
		secret, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%v is not a valid boolean type", value)
		}
		updateSQL.Write([]byte("update sellers set secret=$1 where id=$2"))
		args = append(args, secret, seller.DatabaseID)
		seller.Secret = secret

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Seller type", field)

	}

	stmt, err = db.Client.PrepareContext(ctx, updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateSeller SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(args...)
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
	db.sellers[sellerID] = seller
	db.sellerMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

func (db *SQL) UpdateDatacenter(ctx context.Context, datacenterID uint64, field string, value interface{}) error {

	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt

	datacenter, err := db.Datacenter(datacenterID)
	if err != nil {
		return &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%016x", datacenterID)}
	}

	switch field {
	case "Latitude":
		latitude, ok := value.(float32)
		if !ok {
			return fmt.Errorf("%v is not a valid float32 value", value)
		}
		updateSQL.Write([]byte("update datacenters set latitude=$1 where id=$2"))
		args = append(args, latitude, datacenter.DatabaseID)
		datacenter.Location.Latitude = latitude
	case "Longitude":
		longitude, ok := value.(float32)
		if !ok {
			return fmt.Errorf("%v is not a valid float32 value", value)
		}
		updateSQL.Write([]byte("update datacenters set longitude=$1 where id=$2"))
		args = append(args, longitude, datacenter.DatabaseID)
		datacenter.Location.Longitude = longitude
	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Datacenter type", field)

	}

	stmt, err = db.Client.PrepareContext(ctx, updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateDatacenter SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(args...)
	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying datacenter record", "err", err)
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
	db.datacenters[datacenterID] = datacenter
	db.datacenterMutex.Unlock()

	db.IncrementSequenceNumber(ctx)

	return nil
}

func (db *SQL) GetDatabaseBinFileMetaData() (routing.DatabaseBinFileMetaData, error) {
	var querySQL bytes.Buffer
	var dashboardData routing.DatabaseBinFileMetaData

	querySQL.Write([]byte("select bin_file_creation_time, bin_file_author "))
	querySQL.Write([]byte("from database_bin_meta order by bin_file_creation_time desc limit 1"))

	row := db.Client.QueryRow(querySQL.String())
	switch err := row.Scan(&dashboardData.DatabaseBinFileCreationTime, &dashboardData.DatabaseBinFileAuthor); err {
	case sql.ErrNoRows:
		level.Error(db.Logger).Log("during", "GetFleetDashboardData() no rows were returned!")
		return routing.DatabaseBinFileMetaData{}, err
	case nil:
		return dashboardData, nil
	default:
		level.Error(db.Logger).Log("during", "GetFleetDashboardData() QueryRow returned an error: %v", err)
		return routing.DatabaseBinFileMetaData{}, err
	}

}

func (db *SQL) UpdateDatabaseBinFileMetaData(ctx context.Context, metaData routing.DatabaseBinFileMetaData) error {

	var sql bytes.Buffer

	// Add the metadata record to the database_bin_meta table
	sql.Write([]byte("insert into database_bin_meta ("))
	sql.Write([]byte("bin_file_creation_time, bin_file_author "))
	sql.Write([]byte(") values ($1, $2)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateDatabaseBinFileMetaData SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(metaData.DatabaseBinFileCreationTime, metaData.DatabaseBinFileAuthor)

	if err != nil {
		level.Error(db.Logger).Log("during", "UpdateDatabaseBinFileMetaData() error adding DatabaseBinFileMetaData", "err", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		level.Error(db.Logger).Log("during", "UpdateDatabaseBinFileMetaData() RowsAffected returned an error", "err", err)
		return err
	}
	if rows != 1 {
		level.Error(db.Logger).Log("during", "UpdateDatabaseBinFileMetaData() RowsAffected <> 1", "err", err)
		return err
	}

	return nil
}
