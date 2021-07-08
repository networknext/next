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
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport/notifications"
)

// SQL is an implementation of the Storer interface. It can
// be backed up by PostgreSQL or SQLite3. There is a
// dichotomy set in concrete, here. Using the routing.Relay
// type as an example:
//
// 	Relay.ID     : Internal-use, calculated from public IP:Port
//  Relay.RelayID: Database primary key for the relays table
//
// The PK is required to enforce business rules in the DB.
type SQL struct {
	Client *sql.DB
	Logger log.Logger
}

type sqlBuyer struct {
	SdkID          int64
	ID             uint64
	IsLiveCustomer bool
	Debug          bool
	Name           string
	PublicKey      []byte
	ShortName      string
	CompanyCode    string // should not be needed
	DatabaseID     int64  // sql PK
	CustomerID     int64  // sql PK
}

type sqlDatacenterMap struct {
	Alias        string
	BuyerID      int64
	DatacenterID int64
}

type sqlInternalConfig struct {
	RouteSelectThreshold           int64
	RouteSwitchThreshold           int64
	MaxLatencyTradeOff             int64
	RTTVetoDefault                 int64
	RTTVetoPacketLoss              int64
	RTTVetoMultipath               int64
	MultipathOverloadThreshold     int64
	TryBeforeYouBuy                bool
	ForceNext                      bool
	LargeCustomer                  bool
	Uncommitted                    bool
	MaxRTT                         int64
	HighFrequencyPings             bool
	RouteDiversity                 int64
	MultipathThreshold             int64
	EnableVanityMetrics            bool
	ReducePacketLossMinSliceNumber int64
	BuyerID                        int64
}

type sqlRouteShader struct {
	ABTest                    bool
	AcceptableLatency         int64
	AcceptablePacketLoss      float64
	BandwidthEnvelopeDownKbps int64
	BandwidthEnvelopeUpKbps   int64
	DisableNetworkNext        bool
	LatencyThreshold          int64
	Multipath                 bool
	ProMode                   bool
	ReduceLatency             bool
	ReducePacketLoss          bool
	ReduceJitter              bool
	SelectionPercent          int64
	PacketLossSustained       float64
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

func (db *SQL) CustomerIDToCode(id int64) (string, error) {
	var querySQL bytes.Buffer
	var customer sqlCustomer

	querySQL.Write([]byte("select customer_code from customers where id = $1"))

	row := db.Client.QueryRow(querySQL.String(), id)
	err := row.Scan(&customer.CustomerCode)
	switch err {
	case sql.ErrNoRows:
		level.Error(db.Logger).Log("during", "CustomerIDToCode() no rows were returned!")
		return "", &DoesNotExistError{resourceType: "customer id", resourceRef: id}
	case nil:
		return customer.CustomerCode, nil
	default:
		level.Error(db.Logger).Log("during", "CustomerIDToCode() QueryRow returned an error: %v", err)
		return "", err
	}
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

	sqlBuyerID := int64(ephemeralBuyerID)

	var querySQL bytes.Buffer
	var buyer sqlBuyer

	querySQL.Write([]byte("select id, short_name, is_live_customer, debug, public_key, customer_id "))
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
			ic = core.NewInternalConfig()
		}

		rs, err := db.RouteShader(ephemeralBuyerID)
		if err != nil {
			rs = core.NewRouteShader()
		}

		b := routing.Buyer{
			ID:             ephemeralBuyerID,
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
func (db *SQL) BuyerWithCompanyCode(companyCode string) (routing.Buyer, error) {
	var querySQL bytes.Buffer
	var buyer sqlBuyer

	querySQL.Write([]byte("select id, sdk_generated_id, is_live_customer, debug, public_key, customer_id "))
	querySQL.Write([]byte("from buyers where short_name = $1"))

	row := db.Client.QueryRow(querySQL.String(), companyCode)
	err := row.Scan(
		&buyer.DatabaseID,
		&buyer.SdkID,
		&buyer.IsLiveCustomer,
		&buyer.Debug,
		&buyer.PublicKey,
		&buyer.CustomerID,
	)
	switch err {
	case sql.ErrNoRows:
		return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer short_name", resourceRef: companyCode}
	case nil:
		buyer.ID = uint64(buyer.SdkID)
		ic, err := db.InternalConfig(buyer.ID)
		if err != nil {
			ic = core.NewInternalConfig()
		}

		rs, err := db.RouteShader(buyer.ID)
		if err != nil {
			rs = core.NewRouteShader()
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
		level.Error(db.Logger).Log("during", "BuyerWithCompanyCode() QueryRow returned an error: %v", err)
		return routing.Buyer{}, err
	}
}

// Buyers returns a copy of all stored buyers.
func (db *SQL) Buyers() []routing.Buyer {
	var sql bytes.Buffer
	var buyer sqlBuyer

	buyers := []routing.Buyer{}
	buyerIDs := make(map[uint64]int64)

	sql.Write([]byte("select sdk_generated_id, id, short_name, is_live_customer, debug, public_key, customer_id "))
	sql.Write([]byte("from buyers"))

	rows, err := db.Client.QueryContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "Buyers(): QueryContext returned an error", "err", err)
		return []routing.Buyer{}
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&buyer.SdkID,
			&buyer.DatabaseID,
			&buyer.ShortName,
			&buyer.IsLiveCustomer,
			&buyer.Debug,
			&buyer.PublicKey,
			&buyer.CustomerID,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "Buyers(): error parsing returned row", "err", err)
			return []routing.Buyer{}
		}

		buyer.ID = uint64(buyer.SdkID)

		buyerIDs[buyer.ID] = buyer.DatabaseID

		ic, err := db.InternalConfig(buyer.ID)
		if err != nil {
			ic = core.NewInternalConfig()
		}

		rs, err := db.RouteShader(buyer.ID)
		if err != nil {
			rs = core.NewRouteShader()
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

		buyers = append(buyers, b)

	}

	sort.Slice(buyers, func(i int, j int) bool { return buyers[i].ID < buyers[j].ID })
	return buyers
}

// AddBuyer adds the provided buyer to storage and returns an error if the buyer could not be added.
func (db *SQL) AddBuyer(ctx context.Context, b routing.Buyer) error {
	var sql bytes.Buffer

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

	return nil
}

// RemoveBuyer removes a buyer with the provided buyer ID from storage
// and returns an error if:
//  1. The buyer ID does not exist
//  2. Removing the buyer would break the foreigh key relationship (datacenter_maps)
//  3. Any other error returned from the database
func (db *SQL) RemoveBuyer(ctx context.Context, ephemeralBuyerID uint64) error {
	var sql bytes.Buffer

	sql.Write([]byte("delete from buyers where sdk_generated_id = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveBuyer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(int64(ephemeralBuyerID))

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

	return nil
}

// Seller gets a copy of a seller with the specified seller ID,
// and returns an empty seller and an error if a seller with that ID doesn't exist in storage.
func (db *SQL) Seller(id string) (routing.Seller, error) {

	var querySQL bytes.Buffer
	var seller sqlSeller

	querySQL.Write([]byte("select id, short_name, public_egress_price, secret, "))
	querySQL.Write([]byte("customer_id from sellers where short_name = $1"))

	row := db.Client.QueryRow(querySQL.String(), id)
	err := row.Scan(&seller.DatabaseID,
		&seller.ShortName,
		&seller.EgressPriceNibblinsPerGB,
		&seller.Secret,
		&seller.CustomerID)
	switch err {
	case sql.ErrNoRows:
		level.Error(db.Logger).Log("during", "Seller() no rows were returned!")
		return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: id}
	case nil:
		c, err := db.Customer(id)
		if err != nil {
			return routing.Seller{}, &DoesNotExistError{resourceType: "customer", resourceRef: id}
		}
		s := routing.Seller{
			ID:                       id,
			ShortName:                seller.ShortName,
			Secret:                   seller.Secret,
			CompanyCode:              c.Code,
			Name:                     c.Name,
			EgressPriceNibblinsPerGB: routing.Nibblin(seller.EgressPriceNibblinsPerGB),
			DatabaseID:               seller.DatabaseID,
			CustomerID:               seller.CustomerID,
		}
		return s, nil
	default:
		level.Error(db.Logger).Log("during", "Seller() QueryRow returned an error: %v", err)
		return routing.Seller{}, err
	}
}

// SellerByDbId returns the sellers table entry for the given ID
// TODO: add to storer interface?
func (db *SQL) SellerByDbId(id int64) (routing.Seller, error) {

	var querySQL bytes.Buffer
	var seller sqlSeller

	querySQL.Write([]byte("select short_name, public_egress_price, secret, "))
	querySQL.Write([]byte("customer_id from sellers where id = $1"))

	row := db.Client.QueryRow(querySQL.String(), id)
	err := row.Scan(
		&seller.ShortName,
		&seller.EgressPriceNibblinsPerGB,
		&seller.Secret,
		&seller.CustomerID)
	switch err {
	case sql.ErrNoRows:
		level.Error(db.Logger).Log("during", "Seller() no rows were returned!")
		return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: id}
	case nil:
		c, err := db.Customer(seller.ShortName)
		if err != nil {
			return routing.Seller{}, &DoesNotExistError{resourceType: "customer", resourceRef: id}
		}
		s := routing.Seller{
			ID:                       seller.ShortName,
			ShortName:                seller.ShortName,
			Secret:                   seller.Secret,
			CompanyCode:              c.Code,
			Name:                     c.Name,
			EgressPriceNibblinsPerGB: routing.Nibblin(seller.EgressPriceNibblinsPerGB),
			DatabaseID:               id,
			CustomerID:               seller.CustomerID,
		}
		return s, nil
	default:
		level.Error(db.Logger).Log("during", "Seller() QueryRow returned an error: %v", err)
		return routing.Seller{}, err
	}
}

// Sellers returns a copy of all stored sellers.
func (db *SQL) Sellers() []routing.Seller {

	var sql bytes.Buffer
	var seller sqlSeller

	sellers := []routing.Seller{}

	sql.Write([]byte("select id, short_name, public_egress_price, secret, "))
	sql.Write([]byte("customer_id from sellers"))

	rows, err := db.Client.QueryContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "Sellers(): QueryContext returned an error", "err", err)
		return []routing.Seller{}
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&seller.DatabaseID,
			&seller.ShortName,
			&seller.EgressPriceNibblinsPerGB,
			&seller.Secret,
			&seller.CustomerID,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "Sellers(): error parsing returned row", "err", err)
			return []routing.Seller{}
		}

		c, err := db.Customer(seller.ShortName)
		if err != nil {
			level.Error(db.Logger).Log("during", "Sellers(): customer does not exist", "err", err)
			return []routing.Seller{}
		}
		s := routing.Seller{
			ID:                       c.Code,
			ShortName:                seller.ShortName,
			Secret:                   seller.Secret,
			CompanyCode:              c.Code,
			Name:                     c.Name,
			EgressPriceNibblinsPerGB: routing.Nibblin(seller.EgressPriceNibblinsPerGB),
			DatabaseID:               seller.DatabaseID,
			CustomerID:               seller.CustomerID,
		}

		sellers = append(sellers, s)
	}

	sort.Slice(sellers, func(i int, j int) bool { return sellers[i].ShortName < sellers[j].ShortName })
	return sellers
}

type sqlSeller struct {
	ID                       string
	CompanyCode              string
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

	// This check only pertains to the next tool. Stateful clients would already
	// have the customer id.
	c, err := db.Customer(s.CompanyCode)
	if err != nil {
		return &DoesNotExistError{resourceType: "customer", resourceRef: s.CompanyCode}
	}

	newSellerData := sqlSeller{
		ID:                       s.ID,
		ShortName:                s.ShortName,
		CompanyCode:              s.ShortName,
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

	return nil
}

// RemoveSeller removes a seller with the provided seller ID from storage and
// returns an error if:
//  1. The seller ID does not exist
//  2. Removing the seller would break the foreigh key relationship (datacenters)
//  3. Any other error returned from the database
func (db *SQL) RemoveSeller(ctx context.Context, id string) error {
	var sql bytes.Buffer

	sql.Write([]byte("delete from sellers where short_name = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveBuyer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(id)

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
	var querySQL bytes.Buffer
	var seller sqlSeller

	querySQL.Write([]byte("select id, short_name, public_egress_price, secret, "))
	querySQL.Write([]byte("customer_id from sellers where short_name = $1"))

	row := db.Client.QueryRow(querySQL.String(), code)
	err := row.Scan(&seller.DatabaseID,
		&seller.ShortName,
		&seller.EgressPriceNibblinsPerGB,
		&seller.Secret,
		&seller.CustomerID)
	switch err {
	case sql.ErrNoRows:
		return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: code}
	case nil:
		c, err := db.Customer(code)
		if err != nil {
			return routing.Seller{}, &DoesNotExistError{resourceType: "customer", resourceRef: code}
		}
		s := routing.Seller{
			ID:                       code,
			ShortName:                seller.ShortName,
			Secret:                   seller.Secret,
			CompanyCode:              c.Code,
			Name:                     c.Name,
			EgressPriceNibblinsPerGB: routing.Nibblin(seller.EgressPriceNibblinsPerGB),
			DatabaseID:               seller.DatabaseID,
			CustomerID:               seller.CustomerID,
		}
		return s, nil
	default:
		level.Error(db.Logger).Log("during", "SellerWithCompanyCode() QueryRow returned an error", "err", err)
		return routing.Seller{}, err
	}
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
	hexID := fmt.Sprintf("%016x", id)

	var sqlQuery bytes.Buffer
	var relay sqlRelay

	sqlQuery.Write([]byte("select relays.id, relays.display_name, relays.contract_term, relays.end_date, "))
	sqlQuery.Write([]byte("relays.included_bandwidth_gb, relays.management_ip, "))
	sqlQuery.Write([]byte("relays.max_sessions, relays.mrc, relays.overage, relays.port_speed, "))
	sqlQuery.Write([]byte("relays.public_ip, relays.public_ip_port, relays.public_key, "))
	sqlQuery.Write([]byte("relays.ssh_port, relays.ssh_user, relays.start_date, relays.internal_ip, "))
	sqlQuery.Write([]byte("relays.internal_ip_port, relays.bw_billing_rule, relays.datacenter, "))
	sqlQuery.Write([]byte("relays.machine_type, relays.relay_state, "))
	sqlQuery.Write([]byte("relays.internal_ip, relays.internal_ip_port, relays.notes , "))
	sqlQuery.Write([]byte("relays.billing_supplier, relays.relay_version from relays where hex_id = $1"))

	rows := db.Client.QueryRow(sqlQuery.String(), hexID)
	err := rows.Scan(&relay.DatabaseID,
		&relay.Name,
		&relay.ContractTerm,
		&relay.EndDate,
		&relay.IncludedBandwithGB,
		&relay.ManagementIP,
		&relay.MaxSessions,
		&relay.MRC,
		&relay.Overage,
		&relay.NICSpeedMbps,
		&relay.PublicIP,
		&relay.PublicIPPort,
		&relay.PublicKey,
		&relay.SSHPort,
		&relay.SSHUser,
		&relay.StartDate,
		&relay.InternalIP,
		&relay.InternalIPPort,
		&relay.BWRule,
		&relay.DatacenterID,
		&relay.MachineType,
		&relay.State,
		&relay.InternalIP,
		&relay.InternalIPPort,
		&relay.Notes,
		&relay.BillingSupplier,
		&relay.Version,
	)

	switch err {
	case sql.ErrNoRows:
		level.Error(db.Logger).Log("during", "Relay() no rows were returned!")
		return routing.Relay{}, &DoesNotExistError{resourceType: "relay", resourceRef: hexID}
	case nil:
		relayState, err := routing.GetRelayStateSQL(relay.State)
		if err != nil {
			level.Error(db.Logger).Log("during", "invalid relay state", "err", err)
		}

		bwRule, err := routing.GetBandwidthRuleSQL(relay.BWRule)
		if err != nil {
			level.Error(db.Logger).Log("during", "routing.ParseBandwidthRule returned an error", "err", err)
		}

		machineType, err := routing.GetMachineTypeSQL(relay.MachineType)
		if err != nil {
			level.Error(db.Logger).Log("during", "routing.ParseMachineType returned an error", "err", err)
		}

		datacenter, err := db.DatacenterByDbId(relay.DatacenterID)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncRelays error dereferencing datacenter", "err", err)
		}

		seller, err := db.SellerByDbId(datacenter.SellerID)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncRelays error dereferencing seller", "err", err)
		}

		internalID, err := strconv.ParseUint(hexID, 16, 64)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncRelays error parsing hex_id", "err", err)
		}

		r := routing.Relay{
			ID:                  internalID,
			Name:                relay.Name,
			PublicKey:           relay.PublicKey,
			Datacenter:          datacenter,
			NICSpeedMbps:        int32(relay.NICSpeedMbps),
			IncludedBandwidthGB: int32(relay.IncludedBandwithGB),
			State:               relayState,
			ManagementAddr:      relay.ManagementIP,
			SSHUser:             relay.SSHUser,
			SSHPort:             relay.SSHPort,
			MaxSessions:         uint32(relay.MaxSessions),
			MRC:                 routing.Nibblin(relay.MRC),
			Overage:             routing.Nibblin(relay.Overage),
			BWRule:              bwRule,
			ContractTerm:        int32(relay.ContractTerm),
			Type:                machineType,
			Seller:              seller,
			DatabaseID:          relay.DatabaseID,
			Version:             relay.Version,
		}

		// nullable values follow
		if relay.InternalIP.Valid {
			fullInternalAddress := relay.InternalIP.String + ":" + fmt.Sprintf("%d", relay.InternalIPPort.Int64)
			internalAddr, err := net.ResolveUDPAddr("udp", fullInternalAddress)
			if err != nil {
				level.Error(db.Logger).Log("during", "net.ResolveUDPAddr returned an error parsing internal address", "err", err)
			}
			r.InternalAddr = *internalAddr
		}

		if relay.PublicIP.Valid {
			fullPublicAddress := relay.PublicIP.String + ":" + fmt.Sprintf("%d", relay.PublicIPPort.Int64)
			publicAddr, err := net.ResolveUDPAddr("udp", fullPublicAddress)
			if err != nil {
				level.Error(db.Logger).Log("during", "net.ResolveUDPAddr returned an error parsing public address", "err", err)
			}
			r.Addr = *publicAddr
		}

		if relay.BillingSupplier.Valid {
			found := false
			for _, seller := range db.Sellers() {
				if seller.DatabaseID == relay.BillingSupplier.Int64 {
					found = true
					r.BillingSupplier = seller.ID
					break
				}
			}

			if !found {
				errString := fmt.Sprintf("Relay() Unable to find Seller matching BillingSupplier ID %d", relay.BillingSupplier.Int64)
				level.Error(db.Logger).Log("during", errString, "err", err)
			}

		}

		if relay.StartDate.Valid {
			r.StartDate = relay.StartDate.Time
		}

		if relay.EndDate.Valid {
			r.EndDate = relay.EndDate.Time
		}

		if relay.Notes.Valid {
			r.Notes = relay.Notes.String
		}

		return r, nil

	default:
		level.Error(db.Logger).Log("during", "Relay() QueryRow returned an error: %v", err)
		return routing.Relay{}, err
	}

}

// Relays returns a copy of all stored relays.
func (db *SQL) Relays() []routing.Relay {

	var sqlQuery bytes.Buffer
	var relay sqlRelay

	relays := []routing.Relay{}

	sqlQuery.Write([]byte("select relays.id, relays.hex_id, relays.display_name, relays.contract_term, relays.end_date, "))
	sqlQuery.Write([]byte("relays.included_bandwidth_gb, relays.management_ip, "))
	sqlQuery.Write([]byte("relays.max_sessions, relays.mrc, relays.overage, relays.port_speed, "))
	sqlQuery.Write([]byte("relays.public_ip, relays.public_ip_port, relays.public_key, "))
	sqlQuery.Write([]byte("relays.ssh_port, relays.ssh_user, relays.start_date, relays.internal_ip, "))
	sqlQuery.Write([]byte("relays.internal_ip_port, relays.bw_billing_rule, relays.datacenter, "))
	sqlQuery.Write([]byte("relays.machine_type, relays.relay_state, "))
	sqlQuery.Write([]byte("relays.internal_ip, relays.internal_ip_port, relays.notes , "))
	sqlQuery.Write([]byte("relays.billing_supplier, relays.relay_version from relays "))

	rows, err := db.Client.QueryContext(context.Background(), sqlQuery.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "syncRelays(): QueryContext returned an error", "err", err)
		return []routing.Relay{}
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&relay.DatabaseID,
			&relay.HexID,
			&relay.Name,
			&relay.ContractTerm,
			&relay.EndDate,
			&relay.IncludedBandwithGB,
			&relay.ManagementIP,
			&relay.MaxSessions,
			&relay.MRC,
			&relay.Overage,
			&relay.NICSpeedMbps,
			&relay.PublicIP,
			&relay.PublicIPPort,
			&relay.PublicKey,
			&relay.SSHPort,
			&relay.SSHUser,
			&relay.StartDate,
			&relay.InternalIP,
			&relay.InternalIPPort,
			&relay.BWRule,
			&relay.DatacenterID,
			&relay.MachineType,
			&relay.State,
			&relay.InternalIP,
			&relay.InternalIPPort,
			&relay.Notes,
			&relay.BillingSupplier,
			&relay.Version,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncRelays(): error parsing returned row", "err", err)
			return []routing.Relay{}
		}

		relayState, err := routing.GetRelayStateSQL(relay.State)
		if err != nil {
			level.Error(db.Logger).Log("during", "invalid relay state", "err", err)
		}

		bwRule, err := routing.GetBandwidthRuleSQL(relay.BWRule)
		if err != nil {
			level.Error(db.Logger).Log("during", "routing.ParseBandwidthRule returned an error", "err", err)
		}

		machineType, err := routing.GetMachineTypeSQL(relay.MachineType)
		if err != nil {
			level.Error(db.Logger).Log("during", "routing.ParseMachineType returned an error", "err", err)
		}

		datacenter, err := db.DatacenterByDbId(relay.DatacenterID)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncRelays error dereferencing datacenter", "err", err)
		}

		seller, err := db.SellerByDbId(datacenter.SellerID)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncRelays error dereferencing seller", "err", err)
		}

		internalID, err := strconv.ParseUint(relay.HexID, 16, 64)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncRelays error parsing hex_id", "err", err)
		}

		r := routing.Relay{
			ID:                  internalID,
			Name:                relay.Name,
			PublicKey:           relay.PublicKey,
			Datacenter:          datacenter,
			NICSpeedMbps:        int32(relay.NICSpeedMbps),
			IncludedBandwidthGB: int32(relay.IncludedBandwithGB),
			State:               relayState,
			ManagementAddr:      relay.ManagementIP,
			SSHUser:             relay.SSHUser,
			SSHPort:             relay.SSHPort,
			MaxSessions:         uint32(relay.MaxSessions),
			MRC:                 routing.Nibblin(relay.MRC),
			Overage:             routing.Nibblin(relay.Overage),
			BWRule:              bwRule,
			ContractTerm:        int32(relay.ContractTerm),
			Type:                machineType,
			Seller:              seller,
			DatabaseID:          relay.DatabaseID,
			Version:             relay.Version,
		}

		// nullable values follow
		if relay.InternalIP.Valid {
			fullInternalAddress := relay.InternalIP.String + ":" + fmt.Sprintf("%d", relay.InternalIPPort.Int64)
			internalAddr, err := net.ResolveUDPAddr("udp", fullInternalAddress)
			if err != nil {
				level.Error(db.Logger).Log("during", "net.ResolveUDPAddr returned an error parsing internal address", "err", err)
			}
			r.InternalAddr = *internalAddr
		}

		if relay.PublicIP.Valid {
			fullPublicAddress := relay.PublicIP.String + ":" + fmt.Sprintf("%d", relay.PublicIPPort.Int64)
			publicAddr, err := net.ResolveUDPAddr("udp", fullPublicAddress)
			if err != nil {
				level.Error(db.Logger).Log("during", "net.ResolveUDPAddr returned an error parsing public address", "err", err)
			}
			r.Addr = *publicAddr
		}

		if relay.BillingSupplier.Valid {
			found := false
			for _, seller := range db.Sellers() {
				if seller.DatabaseID == relay.BillingSupplier.Int64 {
					found = true
					r.BillingSupplier = seller.ID
					break
				}
			}

			if !found {
				errString := fmt.Sprintf("syncRelays() Unable to find Seller matching BillingSupplier ID %d", relay.BillingSupplier.Int64)
				level.Error(db.Logger).Log("during", errString, "err", err)
			}

		}

		if relay.StartDate.Valid {
			r.StartDate = relay.StartDate.Time
		}

		if relay.EndDate.Valid {
			r.EndDate = relay.EndDate.Time
		}

		if relay.Notes.Valid {
			r.Notes = relay.Notes.String
		}

		relays = append(relays, r)
	}

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

	case "InternalAddr":
		addrString, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		if addrString == "" {
			updateSQL.Write([]byte("update relays set (internal_ip, internal_ip_port) = (null, null) "))
			updateSQL.Write([]byte("where id=$1"))
			args = append(args, relay.DatabaseID)

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

	case "NICSpeedMbps":
		portSpeed, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		updateSQL.Write([]byte("update relays set port_speed=$1 where id=$2"))
		args = append(args, portSpeed, relay.DatabaseID)

	case "IncludedBandwidthGB":
		includedBW, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		updateSQL.Write([]byte("update relays set included_bandwidth_gb=$1 where id=$2"))
		args = append(args, includedBW, relay.DatabaseID)

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

	case "ManagementAddr":
		// routing.Relay.ManagementIP is currently a string type although
		// the database field is inet
		managementIP, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update relays set management_ip=$1 where id=$2"))
		args = append(args, managementIP, relay.DatabaseID)

	case "SSHUser":
		user, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string type", value)
		}
		updateSQL.Write([]byte("update relays set ssh_user=$1 where id=$2"))
		args = append(args, user, relay.DatabaseID)

	case "SSHPort":
		port, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		updateSQL.Write([]byte("update relays set ssh_port=$1 where id=$2"))
		args = append(args, port, relay.DatabaseID)

	case "MaxSessions":
		maxSessions, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		updateSQL.Write([]byte("update relays set max_sessions=$1 where id=$2"))
		args = append(args, int64(maxSessions), relay.DatabaseID)

	case "MRC":
		mrcUSD, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		mrc := routing.DollarsToNibblins(mrcUSD)
		updateSQL.Write([]byte("update relays set mrc=$1 where id=$2"))
		args = append(args, int64(mrc), relay.DatabaseID)

	case "Overage":
		overageUSD, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		overage := routing.DollarsToNibblins(overageUSD)
		updateSQL.Write([]byte("update relays set overage=$1 where id=$2"))
		args = append(args, int64(overage), relay.DatabaseID)

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
			args = append(args, newStartDate, relay.DatabaseID)
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
			args = append(args, newEndDate, relay.DatabaseID)
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

	case "Notes":
		notes, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		if notes == "" {
			updateSQL.Write([]byte("update relays set notes=null where id=$1"))
			args = append(args, relay.DatabaseID)
		} else {
			updateSQL.Write([]byte("update relays set notes=$1 where id=$2"))
			args = append(args, notes, relay.DatabaseID)
		}

	case "BillingSupplier":
		billingSupplier, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		if billingSupplier == "" {
			updateSQL.Write([]byte("update relays set billing_supplier=null where id=$1"))
			args = append(args, relay.DatabaseID)
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

	return nil
}

// RemoveRelay removes a relay with the provided relay ID from storage and
// returns any database errors to the caller
func (db *SQL) RemoveRelay(ctx context.Context, id uint64) error {
	var sql bytes.Buffer

	hexID := fmt.Sprintf("%016x", id)
	sql.Write([]byte("delete from relays where hex_id = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveRelay SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(hexID)

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
// TODO: chopping block, used in OpsService
func (db *SQL) SetRelay(ctx context.Context, r routing.Relay) error {

	var sqlQuery bytes.Buffer
	var err error

	hexID := fmt.Sprintf("%016x", r.ID)

	var publicIP sql.NullString
	var publicIPPort sql.NullInt64

	if r.Addr.String() != ":0" && r.Addr.String() != "" {
		publicIP.String = strings.Split(r.Addr.String(), ":")[0]
		publicIP.Valid = true
		publicIPPort.Int64, err = strconv.ParseInt(strings.Split(r.Addr.String(), ":")[1], 10, 64)
		publicIPPort.Valid = true
		if err != nil {
			return fmt.Errorf("unable to convert InternalIP Port %s to int: %v", strings.Split(r.Addr.String(), ":")[1], err)
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
		HexID:              hexID,
	}

	sqlQuery.Write([]byte("update relays set ("))
	sqlQuery.Write([]byte("hex_id, contract_term, display_name, end_date, included_bandwidth_gb, "))
	sqlQuery.Write([]byte("management_ip, max_sessions, mrc, overage, port_speed, public_ip, "))
	sqlQuery.Write([]byte("public_ip_port, public_key, ssh_port, ssh_user, start_date, "))
	sqlQuery.Write([]byte("bw_billing_rule, datacenter, machine_type, relay_state, internal_ip, internal_ip_port "))
	sqlQuery.Write([]byte(") = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, "))
	sqlQuery.Write([]byte("$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22) where id = $23"))

	stmt, err := db.Client.PrepareContext(ctx, sqlQuery.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing SetRelay SQL", "err", err)
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

	return nil
}

// Datacenter gets a copy of a datacenter with the specified datacenter ID
// and returns an empty datacenter and an error if a datacenter with that ID doesn't exist in storage.
func (db *SQL) Datacenter(datacenterID uint64) (routing.Datacenter, error) {

	hexID := fmt.Sprintf("%016x", datacenterID)
	var querySQL bytes.Buffer
	var dc sqlDatacenter

	querySQL.Write([]byte("select id, display_name, latitude, longitude,"))
	querySQL.Write([]byte("seller_id from datacenters where hex_id = $1"))

	row := db.Client.QueryRow(querySQL.String(), hexID)
	err := row.Scan(&dc.ID,
		&dc.Name,
		&dc.Latitude,
		&dc.Longitude,
		&dc.SellerID)
	switch err {
	case sql.ErrNoRows:
		level.Error(db.Logger).Log("during", "Datacenter() no rows were returned!")
		return routing.Datacenter{}, &DoesNotExistError{resourceType: "datacenter", resourceRef: hexID}
	case nil:
		d := routing.Datacenter{
			ID:   datacenterID,
			Name: dc.Name,
			Location: routing.Location{
				Latitude:  dc.Latitude,
				Longitude: dc.Longitude,
			},
			SellerID:   dc.SellerID,
			DatabaseID: dc.ID}
		return d, nil
	default:
		level.Error(db.Logger).Log("during", "Datacenter() QueryRow returned an error: %v", err)
		return routing.Datacenter{}, err
	}
}

// DatacenterByDbId retrives the entry in the datacenters table for the provided ID
// TODO: add to storer interface?
func (db *SQL) DatacenterByDbId(databaseID int64) (routing.Datacenter, error) {

	var querySQL bytes.Buffer
	var dc sqlDatacenter

	querySQL.Write([]byte("select hex_id, display_name, latitude, longitude,"))
	querySQL.Write([]byte("seller_id from datacenters where id = $1"))

	row := db.Client.QueryRow(querySQL.String(), databaseID)
	err := row.Scan(
		&dc.HexID,
		&dc.Name,
		&dc.Latitude,
		&dc.Longitude,
		&dc.SellerID)
	switch err {
	case sql.ErrNoRows:
		level.Error(db.Logger).Log("during", "DatacenterByDbId() no rows were returned!")
		return routing.Datacenter{}, &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%d", databaseID)}
	case nil:
		datacenterID, err := strconv.ParseUint(dc.HexID, 16, 64)
		if err != nil {
			level.Error(db.Logger).Log("during", "DatacenterByDbId()error parsing hex ID")
			return routing.Datacenter{}, &HexStringConversionError{hexString: dc.HexID}
		}
		d := routing.Datacenter{
			ID:   datacenterID,
			Name: dc.Name,
			Location: routing.Location{
				Latitude:  dc.Latitude,
				Longitude: dc.Longitude,
			},
			SellerID:   dc.SellerID,
			DatabaseID: databaseID,
		}
		return d, nil
	default:
		level.Error(db.Logger).Log("during", "DatacenterByDbId() QueryRow returned an error: %v", err)
		return routing.Datacenter{}, err
	}
}

// Datacenters returns a copy of all stored datacenters.
func (db *SQL) Datacenters() []routing.Datacenter {

	var sql bytes.Buffer
	var dc sqlDatacenter

	datacenters := []routing.Datacenter{}

	sql.Write([]byte("select id, display_name, latitude, longitude,"))
	sql.Write([]byte("seller_id from datacenters"))

	rows, err := db.Client.QueryContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "syncDatacenters(): QueryContext returned an error", "err", err)
		return []routing.Datacenter{}
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&dc.ID,
			&dc.Name,
			&dc.Latitude,
			&dc.Longitude,
			&dc.SellerID,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncDatacenters(): error parsing returned row", "err", err)
			return []routing.Datacenter{}
		}

		did := crypto.HashID(dc.Name)

		d := routing.Datacenter{
			ID:   did,
			Name: dc.Name,
			Location: routing.Location{
				Latitude:  dc.Latitude,
				Longitude: dc.Longitude,
			},
			SellerID:   dc.SellerID,
			DatabaseID: dc.ID,
		}

		datacenters = append(datacenters, d)
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

	hexID := fmt.Sprintf("%016x", id)
	sql.Write([]byte("delete from datacenters where hex_id = $1"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveDatacenter SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(hexID)

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

	return nil
}

// GetDatacenterMapsForBuyer returns a map of datacenter aliases in use for a given
// (internally generated) buyerID. The map is indexed by the datacenter ID. Returns
// an empty map if there are no aliases for that buyerID.
func (db *SQL) GetDatacenterMapsForBuyer(ephemeralBuyerID uint64) map[uint64]routing.DatacenterMap {

	var querySQL bytes.Buffer
	var dcMaps = make(map[uint64]routing.DatacenterMap)
	// var sqlMap sqlDatacenterMap

	dbBuyerID := int64(ephemeralBuyerID)

	querySQL.Write([]byte("select datacenters.hex_id from datacenter_maps "))
	querySQL.Write([]byte("inner join datacenters on datacenter_maps.datacenter_id "))
	querySQL.Write([]byte("= datacenters.id where datacenter_maps.buyer_id = "))
	querySQL.Write([]byte("(select id from buyers where sdk_generated_id = $1)"))

	rows, err := db.Client.QueryContext(context.Background(), querySQL.String(), dbBuyerID)
	if err != nil {
		level.Error(db.Logger).Log("during", "GetDatacenterMapsForBuyer(): QueryContext returned an error", "err", err)
		return map[uint64]routing.DatacenterMap{}
	}
	defer rows.Close()

	for rows.Next() {
		var hexID string
		err = rows.Scan(&hexID)
		if err != nil {
			level.Error(db.Logger).Log("during", "GetDatacenterMapsForBuyer(): error parsing returned row", "err", err)
			return map[uint64]routing.DatacenterMap{}
		}

		dcID, err := strconv.ParseUint(hexID, 16, 64)
		if err != nil {
			level.Error(db.Logger).Log("during", "GetDatacenterMapsForBuyer() error parsing datacenter hex ID")
			return map[uint64]routing.DatacenterMap{}
		}

		dcMap := routing.DatacenterMap{
			BuyerID:      ephemeralBuyerID,
			DatacenterID: dcID,
		}

		dcMaps[dcID] = dcMap
	}

	return dcMaps
}

// AddDatacenterMap adds a new datacenter map for the given buyer and datacenter IDs
func (db *SQL) AddDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {

	var sql bytes.Buffer

	buyer, err := db.Buyer(dcMap.BuyerID)
	if err != nil {
		return &DoesNotExistError{resourceType: "Buyer.ID", resourceRef: fmt.Sprintf("%016x", dcMap.BuyerID)}
	}

	datacenter, err := db.Datacenter(dcMap.DatacenterID)
	if err != nil {
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

	return nil
}

// ListDatacenterMaps returns a list of alias/buyer mappings for the specified datacenter ID. An
// empty dcID returns a list of all maps.
func (db *SQL) ListDatacenterMaps(dcID uint64) map[uint64]routing.DatacenterMap {

	var querySQL bytes.Buffer
	var dcMaps = make(map[uint64]routing.DatacenterMap)
	var sqlMap sqlDatacenterMap

	hexID := fmt.Sprintf("%016x", dcID)

	querySQL.Write([]byte("select sdk_generated_id from buyers where id in ( "))
	querySQL.Write([]byte("select buyer_id from datacenter_maps where datacenter_id = ( "))
	querySQL.Write([]byte("select id from datacenters where hex_id = $1 )) "))

	rows, err := db.Client.QueryContext(context.Background(), querySQL.String(), hexID)
	if err != nil {
		level.Error(db.Logger).Log("during", "ListDatacenterMaps(): QueryContext returned an error", "err", err)
		return map[uint64]routing.DatacenterMap{}
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&sqlMap.BuyerID)
		if err != nil {
			level.Error(db.Logger).Log("during", "ListDatacenterMaps(): error parsing returned row", "err", err)
			return map[uint64]routing.DatacenterMap{}
		}

		dcMap := routing.DatacenterMap{
			BuyerID:      uint64(sqlMap.BuyerID),
			DatacenterID: dcID,
		}

		id := crypto.HashID(fmt.Sprintf("%016x", dcMap.BuyerID) + fmt.Sprintf("%016x", dcMap.DatacenterID))
		dcMaps[id] = dcMap
	}

	return dcMaps
}

// RemoveDatacenterMap removes an entry from the DatacenterMaps table
func (db *SQL) RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {
	var sql bytes.Buffer

	sql.Write([]byte("delete from datacenter_maps where buyer_id = "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $1) "))
	sql.Write([]byte("and datacenter_id = (select id from datacenters where hex_id = $2)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveDatacenterMap SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(int64(dcMap.BuyerID), fmt.Sprintf("%016x", dcMap.DatacenterID))

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

	return nil
}

// RouteShaders returns a slice of route shaders for the given buyer ID
func (db *SQL) RouteShader(ephemeralBuyerID uint64) (core.RouteShader, error) {

	var querySQL bytes.Buffer
	var sqlRS sqlRouteShader

	querySQL.Write([]byte("select ab_test, acceptable_latency, acceptable_packet_loss, bw_envelope_down_kbps, "))
	querySQL.Write([]byte("bw_envelope_up_kbps, disable_network_next, latency_threshold, multipath, pro_mode, "))
	querySQL.Write([]byte("reduce_latency, reduce_packet_loss, reduce_jitter, selection_percent, "))
	querySQL.Write([]byte("packet_loss_sustained from route_shaders where buyer_id = ( "))
	querySQL.Write([]byte("select id from buyers where sdk_generated_id = $1)"))

	row := db.Client.QueryRow(querySQL.String(), int64(ephemeralBuyerID))
	err := row.Scan(
		&sqlRS.ABTest,
		&sqlRS.AcceptableLatency,
		&sqlRS.AcceptablePacketLoss,
		&sqlRS.BandwidthEnvelopeDownKbps,
		&sqlRS.BandwidthEnvelopeUpKbps,
		&sqlRS.DisableNetworkNext,
		&sqlRS.LatencyThreshold,
		&sqlRS.Multipath,
		&sqlRS.ProMode,
		&sqlRS.ReduceLatency,
		&sqlRS.ReducePacketLoss,
		&sqlRS.ReduceJitter,
		&sqlRS.SelectionPercent,
		&sqlRS.PacketLossSustained,
	)
	switch err {
	case sql.ErrNoRows:
		// By default buyers do not have a custom route shader so will not
		// have an entry in the route_shaders table. We probably don't need
		// to log an error here. However, the return is checked and the
		// default core.RouteShader is applied to the buyer if an error
		// is returned.
		// level.Error(db.Logger).Log("during", "RouteShader() no rows were returned!")
		return core.RouteShader{}, &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%016x", ephemeralBuyerID)}
	case nil:
		routeShader := core.RouteShader{
			DisableNetworkNext:        sqlRS.DisableNetworkNext,
			SelectionPercent:          int(sqlRS.SelectionPercent),
			ABTest:                    sqlRS.ABTest,
			ProMode:                   sqlRS.ProMode,
			ReduceLatency:             sqlRS.ReduceLatency,
			ReducePacketLoss:          sqlRS.ReducePacketLoss,
			ReduceJitter:              sqlRS.ReduceJitter,
			Multipath:                 sqlRS.Multipath,
			AcceptableLatency:         int32(sqlRS.AcceptableLatency),
			LatencyThreshold:          int32(sqlRS.LatencyThreshold),
			AcceptablePacketLoss:      float32(sqlRS.AcceptablePacketLoss),
			BandwidthEnvelopeUpKbps:   int32(sqlRS.BandwidthEnvelopeUpKbps),
			BandwidthEnvelopeDownKbps: int32(sqlRS.BandwidthEnvelopeDownKbps),
			PacketLossSustained:       float32(sqlRS.PacketLossSustained),
		}

		bannedUserList, err := db.BannedUsers(ephemeralBuyerID)
		if err != nil {
			level.Error(db.Logger).Log("during", "RouteShader() -> BannedUsers() returned an error")
			return core.RouteShader{}, fmt.Errorf("RouteShader() -> BannedUser() returned an error: %v for Buyer %s", err, fmt.Sprintf("%016x", ephemeralBuyerID))
		}
		routeShader.BannedUsers = bannedUserList
		return routeShader, nil
	default:
		level.Error(db.Logger).Log("during", "RouteShader() QueryRow returned an error: %v", err)
		return core.RouteShader{}, err
	}
}

// InternalConfig returns the InternalConfig entry for the specified buyer
func (db *SQL) InternalConfig(ephemeralBuyerID uint64) (core.InternalConfig, error) {

	var querySQL bytes.Buffer
	var sqlIC sqlInternalConfig

	querySQL.Write([]byte("select max_latency_tradeoff, max_rtt, multipath_overload_threshold, "))
	querySQL.Write([]byte("route_switch_threshold, route_select_threshold, rtt_veto_default, "))
	querySQL.Write([]byte("rtt_veto_multipath, rtt_veto_packetloss, try_before_you_buy, force_next, "))
	querySQL.Write([]byte("large_customer, is_uncommitted, high_frequency_pings, route_diversity, "))
	querySQL.Write([]byte("multipath_threshold, enable_vanity_metrics, reduce_pl_min_slice_number "))
	querySQL.Write([]byte("from rs_internal_configs where buyer_id = ( "))
	querySQL.Write([]byte("select id from buyers where sdk_generated_id = $1)"))

	row := db.Client.QueryRow(querySQL.String(), int64(ephemeralBuyerID))
	err := row.Scan(
		&sqlIC.MaxLatencyTradeOff,
		&sqlIC.MaxRTT,
		&sqlIC.MultipathOverloadThreshold,
		&sqlIC.RouteSwitchThreshold,
		&sqlIC.RouteSelectThreshold,
		&sqlIC.RTTVetoDefault,
		&sqlIC.RTTVetoMultipath,
		&sqlIC.RTTVetoPacketLoss,
		&sqlIC.TryBeforeYouBuy,
		&sqlIC.ForceNext,
		&sqlIC.LargeCustomer,
		&sqlIC.Uncommitted,
		&sqlIC.HighFrequencyPings,
		&sqlIC.RouteDiversity,
		&sqlIC.MultipathThreshold,
		&sqlIC.EnableVanityMetrics,
		&sqlIC.ReducePacketLossMinSliceNumber,
	)
	switch err {
	case sql.ErrNoRows:
		// By default buyers do not have a custom internal config so will not
		// have an entry in the rs_internal_configs table. We probably don't need
		// to log an error here. However, the return is checked and the
		// default core.InternalConfig is applied to the buyer if an error
		// is returned.
		// level.Error(db.Logger).Log("during", "InternalConfig() no rows were returned!")
		return core.InternalConfig{}, &DoesNotExistError{resourceType: "InternalConfig", resourceRef: fmt.Sprintf("%016x", ephemeralBuyerID)}
	case nil:
		internalConfig := core.InternalConfig{
			RouteSelectThreshold:           int32(sqlIC.RouteSelectThreshold),
			RouteSwitchThreshold:           int32(sqlIC.RouteSwitchThreshold),
			MaxLatencyTradeOff:             int32(sqlIC.MaxLatencyTradeOff),
			RTTVeto_Default:                int32(sqlIC.RTTVetoDefault),
			RTTVeto_PacketLoss:             int32(sqlIC.RTTVetoPacketLoss),
			RTTVeto_Multipath:              int32(sqlIC.RTTVetoMultipath),
			MultipathOverloadThreshold:     int32(sqlIC.MultipathOverloadThreshold),
			TryBeforeYouBuy:                sqlIC.TryBeforeYouBuy,
			ForceNext:                      sqlIC.ForceNext,
			LargeCustomer:                  sqlIC.LargeCustomer,
			Uncommitted:                    sqlIC.Uncommitted,
			MaxRTT:                         int32(sqlIC.MaxRTT),
			HighFrequencyPings:             sqlIC.HighFrequencyPings,
			RouteDiversity:                 int32(sqlIC.RouteDiversity),
			MultipathThreshold:             int32(sqlIC.MultipathThreshold),
			EnableVanityMetrics:            sqlIC.EnableVanityMetrics,
			ReducePacketLossMinSliceNumber: int32(sqlIC.ReducePacketLossMinSliceNumber),
		}
		return internalConfig, nil
	default:
		level.Error(db.Logger).Log("during", "InternalConfig() QueryRow returned an error: %v", err)
		return core.InternalConfig{}, err
	}

}

// AddInternalConfig adds an InternalConfig for the specified buyer
func (db *SQL) AddInternalConfig(ctx context.Context, ic core.InternalConfig, ephemeralBuyerID uint64) error {

	var sql bytes.Buffer

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
	sql.Write([]byte("values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $18)"))
	sql.Write([]byte(")"))

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
		int64(ephemeralBuyerID),
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

	return nil
}

func (db *SQL) RemoveInternalConfig(ctx context.Context, ephemeralBuyerID uint64) error {
	var sql bytes.Buffer

	buyerID := int64(ephemeralBuyerID)
	sql.Write([]byte("delete from rs_internal_configs where buyer_id = "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $1)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveInternalConfig SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(buyerID)

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

	return nil
}

func (db *SQL) UpdateInternalConfig(ctx context.Context, ephemeralBuyerID uint64, field string, value interface{}) error {

	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt
	var err error

	switch field {
	case "RouteSelectThreshold":
		routeSelectThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteSelectThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set route_select_threshold=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, routeSelectThreshold, ephemeralBuyerID)
	case "RouteSwitchThreshold":
		routeSwitchThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteSwitchThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set route_switch_threshold=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, routeSwitchThreshold, ephemeralBuyerID)
	case "MaxLatencyTradeOff":
		maxLatencyTradeOff, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MaxLatencyTradeOff: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set max_latency_tradeoff=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, maxLatencyTradeOff, ephemeralBuyerID)
	case "RTTVeto_Default":
		rttVetoDefault, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_Default: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_default=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, rttVetoDefault, ephemeralBuyerID)
	case "RTTVeto_PacketLoss":
		rttVetoPacketLoss, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_PacketLoss: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_packetloss=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, rttVetoPacketLoss, ephemeralBuyerID)
	case "RTTVeto_Multipath":
		rttVetoMultipath, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_Multipath: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_multipath=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, rttVetoMultipath, ephemeralBuyerID)
	case "MultipathOverloadThreshold":
		multipathOverloadThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MultipathOverloadThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set multipath_overload_threshold=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, multipathOverloadThreshold, ephemeralBuyerID)
	case "TryBeforeYouBuy":
		tryBeforeYouBuy, ok := value.(bool)
		if !ok {
			return fmt.Errorf("TryBeforeYouBuy: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set try_before_you_buy=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, tryBeforeYouBuy, ephemeralBuyerID)
	case "ForceNext":
		forceNext, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ForceNext: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set force_next=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, forceNext, ephemeralBuyerID)
	case "LargeCustomer":
		largeCustomer, ok := value.(bool)
		if !ok {
			return fmt.Errorf("LargeCustomer: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set large_customer=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, largeCustomer, ephemeralBuyerID)
	case "Uncommitted":
		uncommitted, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Uncommitted: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set is_uncommitted=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, uncommitted, ephemeralBuyerID)
	case "HighFrequencyPings":
		highFrequencyPings, ok := value.(bool)
		if !ok {
			return fmt.Errorf("HighFrequencyPings: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set high_frequency_pings=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, highFrequencyPings, ephemeralBuyerID)
	case "MaxRTT":
		maxRTT, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MaxRTT: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set max_rtt=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, maxRTT, ephemeralBuyerID)
	case "RouteDiversity":
		routeDiversity, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteDiversity: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set route_diversity=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, routeDiversity, ephemeralBuyerID)
	case "MultipathThreshold":
		multipathThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MultipathThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set multipath_threshold=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, multipathThreshold, ephemeralBuyerID)
	case "EnableVanityMetrics":
		enableVanityMetrics, ok := value.(bool)
		if !ok {
			return fmt.Errorf("EnableVanityMetrics: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set enable_vanity_metrics=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, enableVanityMetrics, ephemeralBuyerID)
	case "ReducePacketLossMinSliceNumber":
		reducePacketLossMinSliceNumber, ok := value.(int32)
		if !ok {
			return fmt.Errorf("ReducePacketLossMinSliceNumber: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set reduce_pl_min_slice_number=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, reducePacketLossMinSliceNumber, ephemeralBuyerID)

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

	return nil
}

func (db *SQL) AddRouteShader(ctx context.Context, rs core.RouteShader, ephemeralBuyerID uint64) error {

	var sql bytes.Buffer

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
	sql.Write([]byte(") values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $15))"))

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
		int64(ephemeralBuyerID),
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

	return nil

}

func (db *SQL) UpdateRouteShader(ctx context.Context, ephemeralBuyerID uint64, field string, value interface{}) error {

	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt
	var err error

	switch field {
	case "ABTest":
		abTest, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ABTest: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set ab_test=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, abTest, ephemeralBuyerID)
	case "AcceptableLatency":
		acceptableLatency, ok := value.(int32)
		if !ok {
			return fmt.Errorf("AcceptableLatency: %v is not a valid int32 type ( %T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set acceptable_latency=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, acceptableLatency, ephemeralBuyerID)
	case "AcceptablePacketLoss":
		acceptablePacketLoss, ok := value.(float32)
		if !ok {
			return fmt.Errorf("AcceptablePacketLoss: %v is not a valid float32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set acceptable_packet_loss=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, acceptablePacketLoss, ephemeralBuyerID)
	case "BandwidthEnvelopeDownKbps":
		bandwidthEnvelopeDownKbps, ok := value.(int32)
		if !ok {
			return fmt.Errorf("BandwidthEnvelopeDownKbps: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set bw_envelope_down_kbps=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, bandwidthEnvelopeDownKbps, ephemeralBuyerID)
	case "BandwidthEnvelopeUpKbps":
		bandwidthEnvelopeUpKbps, ok := value.(int32)
		if !ok {
			return fmt.Errorf("BandwidthEnvelopeUpKbps: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set bw_envelope_up_kbps=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, bandwidthEnvelopeUpKbps, ephemeralBuyerID)
	case "DisableNetworkNext":
		disableNetworkNext, ok := value.(bool)
		if !ok {
			return fmt.Errorf("DisableNetworkNext: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set disable_network_next=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, disableNetworkNext, ephemeralBuyerID)
	case "LatencyThreshold":
		latencyThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("LatencyThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set latency_threshold=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, latencyThreshold, ephemeralBuyerID)
	case "Multipath":
		multipath, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Multipath: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set multipath=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, multipath, ephemeralBuyerID)
	case "ProMode":
		proMode, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ProMode: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set pro_mode=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, proMode, ephemeralBuyerID)
	case "ReduceLatency":
		reduceLatency, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReduceLatency: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set reduce_latency=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, reduceLatency, ephemeralBuyerID)
	case "ReducePacketLoss":
		reducePacketLoss, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReducePacketLoss: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set reduce_packet_loss=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, reducePacketLoss, ephemeralBuyerID)
	case "ReduceJitter":
		reduceJitter, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReduceJitter: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set reduce_jitter=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, reduceJitter, ephemeralBuyerID)
	case "SelectionPercent":
		selectionPercent, ok := value.(int)
		if !ok {
			return fmt.Errorf("SelectionPercent: %v is not a valid int type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set selection_percent=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, selectionPercent, ephemeralBuyerID)
	case "PacketLossSustained":
		packetLossSustained, ok := value.(float32)
		if !ok {
			return fmt.Errorf("PacketLossSustained: %v is not a valid float type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set packet_loss_sustained=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, packetLossSustained, ephemeralBuyerID)
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

	return nil

}

func (db *SQL) RemoveRouteShader(ctx context.Context, ephemeralBuyerID uint64) error {
	var sql bytes.Buffer

	sql.Write([]byte("delete from route_shaders where buyer_id = "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $1)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveRouteShader SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(int64(ephemeralBuyerID))

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

	return nil
}

// AddBannedUser adds a user to the banned_user table
func (db *SQL) AddBannedUser(ctx context.Context, ephemeralBuyerID uint64, userID uint64) error {

	var sql bytes.Buffer

	sql.Write([]byte("insert into banned_users (user_id, buyer_id) values ($1, "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $2))"))
	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddBannedUser SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(int64(userID), int64(ephemeralBuyerID))
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

	return nil

}

// RemoveBannedUser removes a user from the banned_user table
func (db *SQL) RemoveBannedUser(ctx context.Context, ephemeralBuyerID uint64, userID uint64) error {

	var sql bytes.Buffer

	sql.Write([]byte("delete from banned_users where user_id = $1 and buyer_id = "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddInternalConfig SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(int64(userID), int64(ephemeralBuyerID))
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

	return nil

}

// BannedUsers returns the set of banned users for the specified buyer ID.
func (db *SQL) BannedUsers(ephemeralBuyerID uint64) (map[uint64]bool, error) {

	var sql bytes.Buffer
	bannedUserList := make(map[uint64]bool)

	sql.Write([]byte("select user_id from banned_users where buyer_id = "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $1)"))

	rows, err := db.Client.QueryContext(context.Background(), sql.String(), int64(ephemeralBuyerID))
	if err != nil {
		level.Error(db.Logger).Log("during", "BannedUsers(): QueryContext returned an error", "err", err)
		return bannedUserList, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int64
		err := rows.Scan(&userID)
		if err != nil {
			level.Error(db.Logger).Log("during", "BannedUsers() error parsing user and buyer IDs", "err", err)
			return bannedUserList, err
		}

		bannedUserList[uint64(userID)] = true
	}

	return bannedUserList, nil

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

	switch field {
	case "Live":
		live, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Live: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set is_live_customer=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, live, ephemeralBuyerID)
	case "Debug":
		debug, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Debug: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set debug=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, debug, ephemeralBuyerID)
	case "ShortName":
		shortName, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update buyers set short_name=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, shortName, ephemeralBuyerID)
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
		updateSQL.Write([]byte("update buyers set public_key=$1, sdk_generated_id=$2 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $3)"))
		args = append(args, newPublicKey[8:], int64(newBuyerID), ephemeralBuyerID)

		// TODO: datacenter maps for this buyer must be updated with the new buyer ID

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

	return nil
}

func (db *SQL) UpdateSeller(ctx context.Context, sellerID string, field string, value interface{}) error {

	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt

	seller, err := db.Seller(sellerID)
	if err != nil {
		return &DoesNotExistError{resourceType: "seller", resourceRef: sellerID}
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

type sqlNotification struct {
	ID         int64
	Timestamp  time.Time
	Author     string
	Title      string
	Message    string
	Type       int64
	CustomerID int64
	Priority   int64
	Public     bool
	Paid       bool
	Data       string
}

// Notifications returns all notifications in the database
func (db *SQL) Notifications() []notifications.Notification {
	var sqlQuery bytes.Buffer
	var notification sqlNotification

	allNotifications := []notifications.Notification{}

	sqlQuery.Write([]byte("select id, creation_date, author, "))
	sqlQuery.Write([]byte("card_title, card_body, type_id, "))
	sqlQuery.Write([]byte("customer_id, priority_id, public, paid, "))
	sqlQuery.Write([]byte("json_string from notifications"))

	rows, err := db.Client.QueryContext(context.Background(), sqlQuery.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "Notifications(): QueryContext returned an error", "err", err)
		return allNotifications
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&notification.ID,
			&notification.Timestamp,
			&notification.Author,
			&notification.Title,
			&notification.Message,
			&notification.Type,
			&notification.CustomerID,
			&notification.Priority,
			&notification.Public,
			&notification.Paid,
			&notification.Data,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "Notifications(): error parsing returned row", "err", err)
			return allNotifications
		}

		notificationType, _ := db.NotificationTypeByID(notification.Type)
		notificationPriority, _ := db.NotificationPriorityByID(notification.Priority)
		customerCode, _ := db.CustomerIDToCode(notification.CustomerID)

		n := notifications.Notification{
			ID:           notification.ID,
			Timestamp:    notification.Timestamp,
			Author:       notification.Author,
			Title:        notification.Title,
			Message:      notification.Message,
			Type:         notificationType,
			CustomerCode: customerCode,
			Priority:     notificationPriority,
			Public:       notification.Public,
			Paid:         notification.Paid,
			Data:         notification.Data,
		}

		allNotifications = append(allNotifications, n)
	}

	sort.Slice(allNotifications, func(i int, j int) bool { return allNotifications[i].Timestamp.Before(allNotifications[j].Timestamp) })
	return allNotifications
}

// NotificationsByCustomer Get all notifications in the database
func (db *SQL) NotificationsByCustomer(customerCode string) []notifications.Notification {
	var sqlQuery bytes.Buffer
	var notification sqlNotification

	allNotifications := []notifications.Notification{}

	customer, err := db.Customer(customerCode)
	if err != nil {
		level.Error(db.Logger).Log("during", "NotificationsByCustomer(): ", "err", err)
		return allNotifications
	}

	sqlQuery.Write([]byte("select id, creation_date, author, "))
	sqlQuery.Write([]byte("card_title, card_body, type_id, "))
	sqlQuery.Write([]byte("customer_id, priority_id, public, paid, "))
	sqlQuery.Write([]byte("json_string from notifications where customer_id = $1"))

	rows, err := db.Client.QueryContext(context.Background(), sqlQuery.String(), customer.DatabaseID)
	if err != nil {
		level.Error(db.Logger).Log("during", "NotificationsByCustomer(): QueryContext returned an error", "err", err)
		return allNotifications
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&notification.ID,
			&notification.Timestamp,
			&notification.Author,
			&notification.Title,
			&notification.Message,
			&notification.Type,
			&notification.CustomerID,
			&notification.Priority,
			&notification.Public,
			&notification.Paid,
			&notification.Data,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "NotificationsByCustomer(): error parsing returned row", "err", err)
			return allNotifications
		}

		notificationType, _ := db.NotificationTypeByID(notification.Type)
		notificationPriority, _ := db.NotificationPriorityByID(notification.Priority)

		n := notifications.Notification{
			ID:           notification.ID,
			Timestamp:    notification.Timestamp,
			Author:       notification.Author,
			Title:        notification.Title,
			Message:      notification.Message,
			Type:         notificationType,
			CustomerCode: customerCode,
			Priority:     notificationPriority,
			Public:       notification.Public,
			Paid:         notification.Paid,
			Data:         notification.Data,
		}

		allNotifications = append(allNotifications, n)
	}

	return allNotifications
}

// NotificationByID Remove a specific notification by ID
func (db *SQL) NotificationByID(id int64) (notifications.Notification, error) {
	var sqlQuery bytes.Buffer
	var notification sqlNotification

	sqlQuery.Write([]byte("select id, creation_date, author, "))
	sqlQuery.Write([]byte("card_title, card_body, type_id, "))
	sqlQuery.Write([]byte("customer_id, priority_id, public, paid, "))
	sqlQuery.Write([]byte("json_string from notifications where id = $1"))

	row := db.Client.QueryRow(sqlQuery.String(), id)
	err := row.Scan(&notification.ID,
		&notification.Timestamp,
		&notification.Author,
		&notification.Title,
		&notification.Message,
		&notification.Type,
		&notification.CustomerID,
		&notification.Priority,
		&notification.Public,
		&notification.Paid,
		&notification.Data,
	)
	switch err {
	case sql.ErrNoRows:
		return notifications.Notification{}, &DoesNotExistError{resourceType: "notification id", resourceRef: id}
	case nil:
		notificiationType, _ := db.NotificationTypeByID(notification.Type)
		notificiationPriority, _ := db.NotificationPriorityByID(notification.Priority)

		// TODO: this functionality needs to be a sub select
		customerCode, err := db.CustomerIDToCode(notification.CustomerID)
		if err != nil {
			return notifications.Notification{}, err
		}

		n := notifications.Notification{
			ID:           notification.ID,
			Timestamp:    notification.Timestamp,
			Author:       notification.Author,
			Title:        notification.Title,
			Message:      notification.Message,
			Type:         notificiationType,
			CustomerCode: customerCode,
			Priority:     notificiationPriority,
			Public:       notification.Public,
			Paid:         notification.Paid,
			Data:         notification.Data,
		}
		return n, nil
	default:
		level.Error(db.Logger).Log("during", "NotificationByID() QueryRow returned an error: %v", err)
		return notifications.Notification{}, err
	}
}

func (db *SQL) AddNotification(notification notifications.Notification) error {
	var sqlQuery bytes.Buffer

	// Add the buyer in remote storage

	sqlQuery.Write([]byte("insert into notifications ("))
	sqlQuery.Write([]byte("creation_date, author, card_title, card_body, type_id, "))
	sqlQuery.Write([]byte("customer_id, priority_id, public, paid, json_string"))
	sqlQuery.Write([]byte(") values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"))

	stmt, err := db.Client.PrepareContext(context.Background(), sqlQuery.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddNotification SQL", "err", err)
		return err
	}

	customer, err := db.Customer(notification.CustomerCode)
	if err != nil {
		level.Error(db.Logger).Log("during", "NotificationsByCustomer(): ", "err", err)
		return err
	}

	result, err := stmt.Exec(
		notification.Timestamp,
		notification.Author,
		notification.Title,
		notification.Message,
		notification.Type.ID,
		customer.DatabaseID,
		notification.Priority.ID,
		notification.Public,
		notification.Paid,
		notification.Data,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding notification", "err", err)
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

// UpdateNotification Update a specific notification
func (db *SQL) UpdateNotification(id int64, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt
	var err error

	switch field {
	case "Author":
		author, ok := value.(string)
		if !ok {
			return fmt.Errorf("Author: %v is not a valid string type (%T)", value, value)
		}
		updateSQL.Write([]byte("update notifications set author=$1 where id="))
		updateSQL.Write([]byte("(select id from notifications where id = $2)"))
		args = append(args, author, id)
	case "Title":
		title, ok := value.(string)
		if !ok {
			return fmt.Errorf("Title: %v is not a valid string type (%T)", value, value)
		}
		updateSQL.Write([]byte("update notifications set card_title=$1 where id="))
		updateSQL.Write([]byte("(select id from notifications where id = $2)"))
		args = append(args, title, id)
	case "Message":
		message, ok := value.(string)
		if !ok {
			return fmt.Errorf("Message: %v is not a valid string type (%T)", value, value)
		}
		updateSQL.Write([]byte("update notifications set card_body=$1 where id="))
		updateSQL.Write([]byte("(select id from notifications where id = $2)"))
		args = append(args, message, id)
	case "Type":
		newType, ok := value.(int64)
		if !ok {
			return fmt.Errorf("Type: %v is not a valid int64 type (%T)", value, value)
		}

		updateSQL.Write([]byte("update notifications set type_id=$1 where id="))
		updateSQL.Write([]byte("(select id from notifications where id = $2)"))
		args = append(args, newType, id)
	case "CustomerCode":
		customerCode, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		customer, err := db.Customer(customerCode)
		if err != nil {
			level.Error(db.Logger).Log("during", "customer does not exist", "err", err)
			return err
		}

		updateSQL.Write([]byte("update notifications set customer_id=$1 where id="))
		updateSQL.Write([]byte("(select id from notifications where id = $2)"))
		args = append(args, customer.DatabaseID, id)
	case "Priority":
		newPriority, ok := value.(int64)
		if !ok {
			return fmt.Errorf("Type: %v is not a valid int64 type (%T)", value, value)
		}

		updateSQL.Write([]byte("update notifications set priority_id=$1 where id="))
		updateSQL.Write([]byte("(select id from notifications where id = $2)"))
		args = append(args, newPriority, id)
	case "Public":
		public, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Public: %v is not a valid bool type (%T)", value, value)
		}

		updateSQL.Write([]byte("update notifications set public=$1 where id="))
		updateSQL.Write([]byte("(select id from notifications where id = $2)"))
		args = append(args, public, id)
	case "Paid":
		paid, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Paid: %v is not a valid bool type (%T)", value, value)
		}

		updateSQL.Write([]byte("update notifications set paid=$1 where id="))
		updateSQL.Write([]byte("(select id from notifications where id = $2)"))
		args = append(args, paid, id)
	case "Data":
		data, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update notifications set json_string=$1 where id="))
		updateSQL.Write([]byte("(select id from notifications where id = $2)"))
		args = append(args, data, id)
	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the notifications.Notification type", field)

	}

	stmt, err = db.Client.PrepareContext(context.Background(), updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateNotification SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(args...)
	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying notification record", "err", err)
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

	return nil
}

// RemoveNotification Remove a specific notification
func (db *SQL) RemoveNotification(id int64) error {
	var sql bytes.Buffer

	sql.Write([]byte("delete from notifications where id = $1"))

	stmt, err := db.Client.PrepareContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveNotification SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(id)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing notification", "err", err)
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

type sqlNotificationType struct {
	ID   int64
	Name string
}

// NotificationTypes returns a list of notification types
func (db *SQL) NotificationTypes() []notifications.NotificationType {
	var sql bytes.Buffer
	var notificationType sqlNotificationType

	allNotificationTypes := []notifications.NotificationType{}

	sql.Write([]byte("select id, priority_type from notification_types"))

	rows, err := db.Client.QueryContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "NotificationTypes(): QueryContext returned an error", "err", err)
		return allNotificationTypes
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&notificationType.ID,
			&notificationType.Name,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "Notifications(): error parsing returned row", "err", err)
			return allNotificationTypes
		}

		t := notifications.NotificationType{
			ID:   notificationType.ID,
			Name: notificationType.Name,
		}

		allNotificationTypes = append(allNotificationTypes, t)
	}

	sort.Slice(allNotificationTypes, func(i int, j int) bool { return allNotificationTypes[i].ID < allNotificationTypes[j].ID })
	return allNotificationTypes
}

// NotificationTypeByID Get a specific notification type by ID
func (db *SQL) NotificationTypeByID(id int64) (notifications.NotificationType, error) {
	var sqlQuery bytes.Buffer
	var notificationType sqlNotificationType

	sqlQuery.Write([]byte("select id, priority_type from notification_types where id = $1"))

	row := db.Client.QueryRow(sqlQuery.String(), id)
	err := row.Scan(&notificationType.ID,
		&notificationType.Name,
	)
	switch err {
	case sql.ErrNoRows:
		return notifications.NotificationType{}, &DoesNotExistError{resourceType: "notificationType id", resourceRef: id}
	case nil:
		t := notifications.NotificationType{
			ID:   notificationType.ID,
			Name: notificationType.Name,
		}
		return t, nil
	default:
		level.Error(db.Logger).Log("during", "NotificationByID() QueryRow returned an error: %v", err)
		return notifications.NotificationType{}, err
	}
}

// NotificationTypeByName Remove a specific notification priority by name
func (db *SQL) NotificationTypeByName(name string) (notifications.NotificationType, error) {
	var sqlQuery bytes.Buffer
	var notificationType sqlNotificationType

	sqlQuery.Write([]byte("select id, priority_type from notification_types where priority_type = $1"))

	row := db.Client.QueryRow(sqlQuery.String(), name)
	err := row.Scan(&notificationType.ID,
		&notificationType.Name,
	)
	switch err {
	case sql.ErrNoRows:
		return notifications.NotificationType{}, &DoesNotExistError{resourceType: "notificationType name", resourceRef: name}
	case nil:
		t := notifications.NotificationType{
			ID:   notificationType.ID,
			Name: notificationType.Name,
		}
		return t, nil
	default:
		level.Error(db.Logger).Log("during", "NotificationByID() QueryRow returned an error: %v", err)
		return notifications.NotificationType{}, err
	}
}

// AddNotificationType Add a notification type to the database
func (db *SQL) AddNotificationType(notificationType notifications.NotificationType) error {
	var sql bytes.Buffer

	// Add the buyer in remote storage
	sql.Write([]byte("insert into notification_types (priority_type) values ($1)"))

	stmt, err := db.Client.PrepareContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddNotificationType SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
		notificationType.Name,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding notification type", "err", err)
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

// UpdateNotificationType Update a specific notification type
func (db *SQL) UpdateNotificationType(id int64, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt
	var err error

	switch field {
	case "Name":
		name, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update notification_types set priority_type=$1 where id="))
		updateSQL.Write([]byte("(select id from notification_types where id = $2)"))
		args = append(args, name, id)
	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the notifications.NotificationType type", field)

	}

	stmt, err = db.Client.PrepareContext(context.Background(), updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateNotificationType SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(args...)
	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying notification record", "err", err)
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

	return nil
}

// RemoveNotificationTypeByID Remove a specific notification type
func (db *SQL) RemoveNotificationTypeByID(id int64) error {
	var sql bytes.Buffer

	sql.Write([]byte("delete from notification_types where id = $1"))

	stmt, err := db.Client.PrepareContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveNotificationTypeByID SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(id)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing notification type by id", "err", err)
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

// RemoveNotificationTypeByName Remove a specific notification type
func (db *SQL) RemoveNotificationTypeByName(name string) error {
	var sql bytes.Buffer

	sql.Write([]byte("delete from notification_types where priority_type = $1"))

	stmt, err := db.Client.PrepareContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveNotificationTypeByName SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(name)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing notification type by name", "err", err)
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

type sqlNotificationPriority struct {
	ID    int64
	Name  string
	Color int64
}

// NotificationPriorities returns a list of priorities
func (db *SQL) NotificationPriorities() []notifications.NotificationPriority {
	var sql bytes.Buffer
	var notificationPriority sqlNotificationPriority

	allNotificationPriorities := []notifications.NotificationPriority{}

	sql.Write([]byte("select id, priority_name, color from notification_priorities"))

	rows, err := db.Client.QueryContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "NotificationPriorities(): QueryContext returned an error", "err", err)
		return allNotificationPriorities
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&notificationPriority.ID,
			&notificationPriority.Name,
			&notificationPriority.Color,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "NotificationPriorities(): error parsing returned row", "err", err)
			return allNotificationPriorities
		}

		p := notifications.NotificationPriority{
			ID:    notificationPriority.ID,
			Name:  notificationPriority.Name,
			Color: notificationPriority.Color,
		}

		allNotificationPriorities = append(allNotificationPriorities, p)
	}

	sort.Slice(allNotificationPriorities, func(i int, j int) bool { return allNotificationPriorities[i].ID < allNotificationPriorities[j].ID })
	return allNotificationPriorities
}

// NotificationPriorityByID Get a specific notification priority by ID
func (db *SQL) NotificationPriorityByID(id int64) (notifications.NotificationPriority, error) {
	var sqlQuery bytes.Buffer
	var notificationPriority sqlNotificationPriority

	sqlQuery.Write([]byte("select id, priority_name, color from notification_priorities where id = $1"))

	row := db.Client.QueryRow(sqlQuery.String(), id)
	err := row.Scan(&notificationPriority.ID,
		&notificationPriority.Name,
		&notificationPriority.Color,
	)
	switch err {
	case sql.ErrNoRows:
		return notifications.NotificationPriority{}, &DoesNotExistError{resourceType: "notificationPriority id", resourceRef: id}
	case nil:
		p := notifications.NotificationPriority{
			ID:    notificationPriority.ID,
			Name:  notificationPriority.Name,
			Color: notificationPriority.Color,
		}
		return p, nil
	default:
		level.Error(db.Logger).Log("during", "NotificationPriorityByID() QueryRow returned an error: %v", err)
		return notifications.NotificationPriority{}, err
	}
}

// NotificationPriorityByName Remove a specific notification priority by name
func (db *SQL) NotificationPriorityByName(name string) (notifications.NotificationPriority, error) {
	var sqlQuery bytes.Buffer
	var notificationPriority sqlNotificationPriority

	sqlQuery.Write([]byte("select id, priority_name, color from notification_priorities where priority_name = $1"))

	row := db.Client.QueryRow(sqlQuery.String(), name)
	err := row.Scan(&notificationPriority.ID,
		&notificationPriority.Name,
		&notificationPriority.Color,
	)
	switch err {
	case sql.ErrNoRows:
		return notifications.NotificationPriority{}, &DoesNotExistError{resourceType: "notificationPriority name", resourceRef: name}
	case nil:
		p := notifications.NotificationPriority{
			ID:    notificationPriority.ID,
			Name:  notificationPriority.Name,
			Color: notificationPriority.Color,
		}
		return p, nil
	default:
		level.Error(db.Logger).Log("during", "NotificationPriorityByName() QueryRow returned an error: %v", err)
		return notifications.NotificationPriority{}, err
	}
}

// AddNotificationPriority Add a notification priority to the database
func (db *SQL) AddNotificationPriority(priority notifications.NotificationPriority) error {
	var sql bytes.Buffer

	// Add the buyer in remote storage
	sql.Write([]byte("insert into notification_priorities (priority_name, color) values ($1, $2)"))

	stmt, err := db.Client.PrepareContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing AddNotificationPriority SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(
		priority.Name,
		priority.Color,
	)

	if err != nil {
		level.Error(db.Logger).Log("during", "error adding notification priority", "err", err)
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

// UpdateNotificationPriority Update a specific notification priority
func (db *SQL) UpdateNotificationPriority(id int64, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}
	var stmt *sql.Stmt
	var err error

	switch field {
	case "Name":
		name, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update notification_priorities set priority_name=$1 where id="))
		updateSQL.Write([]byte("(select id from notification_priorities where id = $2)"))
		args = append(args, name, id)
	case "Color":
		color, ok := value.(int64)
		if !ok {
			return fmt.Errorf("%v is not a valid int64 value", value)
		}
		updateSQL.Write([]byte("update notification_priorities set color=$1 where id="))
		updateSQL.Write([]byte("(select id from notification_priorities where id = $2)"))
		args = append(args, color, id)
	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the notifications.NotificationPriority type", field)

	}

	stmt, err = db.Client.PrepareContext(context.Background(), updateSQL.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing UpdateNotificationPriority SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(args...)
	if err != nil {
		level.Error(db.Logger).Log("during", "error modifying notification priority record", "err", err)
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

	return nil
}

// RemoveNotificationPriorityByID Remove a specific notification priority by ID
func (db *SQL) RemoveNotificationPriorityByID(id int64) error {
	var sql bytes.Buffer

	sql.Write([]byte("delete from notification_priorities where id = $1"))

	stmt, err := db.Client.PrepareContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveNotificationPriorityByID SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(id)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing notification priority by id", "err", err)
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

// RemoveNotificationPriorityByName Remove a specific notification priority by name
func (db *SQL) RemoveNotificationPriorityByName(name string) error {
	var sql bytes.Buffer

	sql.Write([]byte("delete from notification_priorities where priority_name = $1"))

	stmt, err := db.Client.PrepareContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error preparing RemoveNotificationPriorityByName SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(name)

	if err != nil {
		level.Error(db.Logger).Log("during", "error removing notification priority by name", "err", err)
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
