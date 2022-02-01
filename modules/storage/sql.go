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

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport/looker"
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
}

const (
	SQL_TIMEOUT = 10 * time.Second
	MAX_RETRIES = 8
)

func QueryMultipleRowsRetry(ctx context.Context, db *SQL, queryString bytes.Buffer, queryArgs ...interface{}) (*sql.Rows, error) {
	var err error
	var sqlRows *sql.Rows

	retryCount := 0
	for retryCount < MAX_RETRIES {
		if len(queryArgs) > 0 {
			sqlRows, err = db.Client.QueryContext(ctx, queryString.String(), queryArgs...)
		} else {
			sqlRows, err = db.Client.QueryContext(ctx, queryString.String())
		}
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}
	return sqlRows, err
}

func ExecRetry(ctx context.Context, db *SQL, queryString bytes.Buffer, queryArgs ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error
	retryCount := 0

	stmt, err := db.Client.PrepareContext(ctx, queryString.String())
	if err != nil {
		core.Error("Failed to prepare ExecRetry SQL: %v", err)
		return nil, err
	}

	for retryCount < MAX_RETRIES {
		result, err = stmt.Exec(queryArgs...)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	return result, err
}

type sqlBuyer struct {
	SdkID               int64
	ID                  uint64
	IsLiveCustomer      bool
	Debug               bool
	Analytics           bool
	Billing             bool
	Trial               bool
	ExoticLocationFee   float64
	StandardLocationFee float64
	Name                string
	PublicKey           []byte
	ShortName           string
	CompanyCode         string // should not be needed
	DatabaseID          int64  // sql PK
	CustomerID          int64  // sql PK
	LookerSeats         int64
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
	AnalysisOnly              bool
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

func (db *SQL) DatabaseBinFileReference(ctx context.Context) (routing.DatabaseBinWrapperReference, error) {
	var sqlQuery bytes.Buffer

	dbReference := routing.DatabaseBinWrapperReference{}

	relays := make([]routing.RelayReference, 0)
	relayMap := make(map[uint64]routing.RelayReference)
	buyers := make([]uint64, 0)
	sellers := make([]string, 0)
	datacenters := make([]string, 0)
	datacenterMaps := make(map[uint64][]uint64)

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sqlQuery.Write([]byte("select sdk_generated_id "))
	sqlQuery.Write([]byte("from buyers"))

	rows, err := QueryMultipleRowsRetry(ctx, db, sqlQuery)
	if err != nil {
		core.Error("DatabaseBinFileReference(): QueryMultipleRowsRetry returned an error")
		return dbReference, err
	}

	var buyerID int64
	for rows.Next() {
		err = rows.Scan(
			&buyerID,
		)
		if err != nil {
			core.Error("DatabaseBinFileReference(): error parsing returned row")
			return dbReference, err
		}

		buyers = append(buyers, uint64(buyerID))
	}

	rows.Close()
	sqlQuery.Reset()

	sqlQuery.Write([]byte("select short_name "))
	sqlQuery.Write([]byte("from sellers"))

	rows, err = QueryMultipleRowsRetry(ctx, db, sqlQuery)
	if err != nil {
		core.Error("DatabaseBinFileReference(): QueryMultipleRowsRetry returned an error")
		return dbReference, err
	}

	var sellerShortName string
	for rows.Next() {
		err = rows.Scan(
			&sellerShortName,
		)
		if err != nil {
			core.Error("DatabaseBinFileReference(): error parsing returned row")
			return dbReference, err
		}

		sellers = append(sellers, sellerShortName)
	}

	rows.Close()
	sqlQuery.Reset()

	sqlQuery.Write([]byte("select display_name, hex_id, public_ip, public_ip_port "))
	sqlQuery.Write([]byte("from relays where relay_state = 0"))

	rows, err = QueryMultipleRowsRetry(ctx, db, sqlQuery)
	if err != nil {
		core.Error("DatabaseBinFileReference(): QueryMultipleRowsRetry returned an error")
		return dbReference, err
	}

	var relayDisplayName string
	var relayHexID string
	var relayPublicIP sql.NullString
	var relayPort sql.NullInt64
	for rows.Next() {
		err = rows.Scan(
			&relayDisplayName,
			&relayHexID,
			&relayPublicIP,
			&relayPort,
		)
		if err != nil {
			core.Error("DatabaseBinFileReference(): error parsing returned row")
			return dbReference, err
		}

		relayID, err := strconv.ParseUint(relayHexID, 16, 64)
		if err != nil {
			core.Error("DatabaseBinFileReference() error parsing datacenter hex ID")
			return dbReference, err
		}

		relayRef := routing.RelayReference{
			DisplayName: relayDisplayName,
		}

		if relayPublicIP.Valid {
			fullPublicAddress := relayPublicIP.String + ":" + fmt.Sprintf("%d", relayPort.Int64)
			publicAddr, err := net.ResolveUDPAddr("udp", fullPublicAddress)
			if err != nil {
				core.Error("Relay() net.ResolveUDPAddr returned an error parsing public address: %v", err)
			}
			relayRef.PublicIP = *publicAddr
		}

		relays = append(relays, relayRef)
		relayMap[relayID] = relayRef
	}

	rows.Close()
	sqlQuery.Reset()

	// TODO: merge this for loop into the buyer ID query
	for _, buyer := range buyers {
		sqlQuery.Write([]byte("select datacenters.hex_id from datacenter_maps "))
		sqlQuery.Write([]byte("inner join datacenters on datacenter_maps.datacenter_id "))
		sqlQuery.Write([]byte("= datacenters.id where datacenter_maps.buyer_id = "))
		sqlQuery.Write([]byte("(select id from buyers where sdk_generated_id = $1)"))

		rows, err := QueryMultipleRowsRetry(ctx, db, sqlQuery, int64(buyer))
		if err != nil {
			core.Error("DatabaseBinFileReference(): QueryMultipleRowsRetry returned an error: %v", err)
			return dbReference, err
		}

		for rows.Next() {
			var hexID string
			err = rows.Scan(&hexID)
			if err != nil {
				core.Error("DatabaseBinFileReference(): error parsing returned row: %v", err)
				return dbReference, err
			}

			dcID, err := strconv.ParseUint(hexID, 16, 64)
			if err != nil {
				core.Error("DatabaseBinFileReference() error parsing datacenter hex ID")
				return dbReference, err
			}

			if _, ok := datacenterMaps[dcID]; !ok {
				datacenterMaps[dcID] = make([]uint64, 0)
			}

			datacenterMaps[dcID] = append(datacenterMaps[dcID], buyer)
		}

		rows.Close()
		sqlQuery.Reset()
	}

	sqlQuery.Write([]byte("select display_name "))
	sqlQuery.Write([]byte("from datacenters "))

	rows, err = QueryMultipleRowsRetry(ctx, db, sqlQuery)
	if err != nil {
		core.Error("DatabaseBinFileReference(): QueryMultipleRowsRetry returned an error: %v", err)
		return dbReference, err
	}

	var datacenterDisplayName string
	for rows.Next() {
		err = rows.Scan(
			&datacenterDisplayName,
		)
		if err != nil {
			core.Error("DatabaseBinFileReference(): error parsing returned row")
			return dbReference, err
		}

		datacenters = append(datacenters, datacenterDisplayName)
	}

	rows.Close()
	sqlQuery.Reset()

	dbReference.Version = routing.DatabaseBinWrapperReferenceVersion
	dbReference.Buyers = buyers
	dbReference.Sellers = sellers
	dbReference.Datacenters = datacenters
	dbReference.DatacenterMaps = datacenterMaps
	dbReference.RelayMap = relayMap
	dbReference.Relays = relays

	return dbReference, err
}

// Customer retrieves a Customer record using the company code
func (db *SQL) Customer(ctx context.Context, customerCode string) (routing.Customer, error) {
	var querySQL bytes.Buffer
	var customer sqlCustomer
	var row *sql.Row
	var err error
	retryCount := 0

	querySQL.Write([]byte("select id, automatic_signin_domain,"))
	querySQL.Write([]byte("customer_name, customer_code from customers where customer_code = $1"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), customerCode)
		err = row.Scan(
			&customer.ID,
			&customer.AutomaticSignInDomains,
			&customer.Name,
			&customer.CustomerCode,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("Customer() connection with database timed out!")
		return routing.Customer{}, err
	case sql.ErrNoRows:
		core.Error("Customer() no rows were returned!")
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
		core.Error("Customer() QueryRow returned an error: %v", err)
		return routing.Customer{}, err
	}
}

func (db *SQL) CustomerByID(ctx context.Context, id int64) (routing.Customer, error) {
	var querySQL bytes.Buffer
	var customer sqlCustomer
	var row *sql.Row
	var err error
	retryCount := 0

	querySQL.Write([]byte("select id, automatic_signin_domain,"))
	querySQL.Write([]byte("customer_name, customer_code from customers where id = $1"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), id)
		err = row.Scan(
			&customer.ID,
			&customer.AutomaticSignInDomains,
			&customer.Name,
			&customer.CustomerCode,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("during", "Customer() connection with the database timed out!")
		return routing.Customer{}, err
	case sql.ErrNoRows:
		core.Error("during", "Customer() no rows were returned!")
		return routing.Customer{}, &DoesNotExistError{resourceType: "customer", resourceRef: id}
	case nil:
		c := routing.Customer{
			Code:                   customer.CustomerCode,
			Name:                   customer.Name,
			AutomaticSignInDomains: customer.AutomaticSignInDomains,
			DatabaseID:             customer.ID,
		}
		return c, nil
	default:
		core.Error("Customer() QueryRow returned an error: %v", err)
		return routing.Customer{}, err
	}
}

// Customers retrieves the full list
// TODO: not covered by sql_test.go
func (db *SQL) Customers(ctx context.Context) []routing.Customer {
	var sql bytes.Buffer
	var customer sqlCustomer

	customers := []routing.Customer{}
	customerIDs := make(map[int64]string)

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("select id, automatic_signin_domain, "))
	sql.Write([]byte("customer_name, customer_code from customers"))

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("Customers(): QueryMultipleRowsRetry returned an error")
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
			core.Error("Customers(): error parsing returned row")
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

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("insert into customers ("))
	sql.Write([]byte("automatic_signin_domain, customer_name, customer_code"))
	sql.Write([]byte(") values ($1, $2, $3)"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		customer.AutomaticSignInDomains,
		customer.Name,
		customer.CustomerCode,
	)
	if err != nil {
		core.Error("AddCustomer() error adding customer: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("AddCustomer() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("AddCustomer() RowsAffected <> 1")
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

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("delete from customers where customer_code = $1"))

	result, err := ExecRetry(ctx, db, sql, customerCode)
	if err != nil {
		core.Error("RemoveCustomer() error removing customer: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveCustomer() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveCustomer() RowsAffected <> 1")
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

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("update customers set (automatic_signin_domain, customer_name) ="))
	sql.Write([]byte("($1, $2) where customer_code = $3"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		c.AutomaticSignInDomains,
		c.Name,
		c.Code,
	)
	if err != nil {
		core.Error("SetCustomer() error modifying customer record: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("SetCustomer() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("SetCustomer() RowsAffected <> 1")
		return err
	}

	return nil
}

// Buyer gets a copy of a buyer with the specified buyer ID,
// and returns an empty buyer and an error if a buyer with that ID doesn't exist in storage.
func (db *SQL) Buyer(ctx context.Context, ephemeralBuyerID uint64) (routing.Buyer, error) {
	var querySQL bytes.Buffer
	var buyer sqlBuyer
	var row *sql.Row
	var err error
	retryCount := 0

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sqlBuyerID := int64(ephemeralBuyerID)

	querySQL.Write([]byte("select id, short_name, is_live_customer, debug, analytics, billing, trial, exotic_location_fee, standard_location_fee, public_key, customer_id, looker_seats "))
	querySQL.Write([]byte("from buyers where sdk_generated_id = $1"))

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), sqlBuyerID)
		err = row.Scan(
			&buyer.DatabaseID,
			&buyer.ShortName,
			&buyer.IsLiveCustomer,
			&buyer.Debug,
			&buyer.Analytics,
			&buyer.Billing,
			&buyer.Trial,
			&buyer.ExoticLocationFee,
			&buyer.StandardLocationFee,
			&buyer.PublicKey,
			&buyer.CustomerID,
			&buyer.LookerSeats,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("Buyer() connection with the database timed out!")
		return routing.Buyer{}, err
	case sql.ErrNoRows:
		core.Error("Buyer() no rows were returned!")
		return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%016x", ephemeralBuyerID)}
	case nil:
		ic, err := db.InternalConfig(ctx, ephemeralBuyerID)
		if err != nil {
			ic = core.NewInternalConfig()
		}

		rs, err := db.RouteShader(ctx, ephemeralBuyerID)
		if err != nil {
			rs = core.NewRouteShader()
		}

		b := routing.Buyer{
			ID:                  ephemeralBuyerID,
			HexID:               fmt.Sprintf("%016x", buyer.ID),
			ShortName:           buyer.ShortName,
			CompanyCode:         buyer.ShortName,
			Live:                buyer.IsLiveCustomer,
			Debug:               buyer.Debug,
			Analytics:           buyer.Analytics,
			Billing:             buyer.Billing,
			Trial:               buyer.Trial,
			ExoticLocationFee:   buyer.ExoticLocationFee,
			StandardLocationFee: buyer.StandardLocationFee,
			PublicKey:           buyer.PublicKey,
			RouteShader:         rs,
			InternalConfig:      ic,
			CustomerID:          buyer.CustomerID,
			DatabaseID:          buyer.DatabaseID,
			LookerSeats:         buyer.LookerSeats,
		}
		return b, nil
	default:
		core.Error("Buyer() QueryRow returned an error: %v", err)
		return routing.Buyer{}, err
	}

}

// BuyerWithCompanyCode gets the Buyer with the matching company code
func (db *SQL) BuyerWithCompanyCode(ctx context.Context, companyCode string) (routing.Buyer, error) {
	var querySQL bytes.Buffer
	var buyer sqlBuyer
	var row *sql.Row
	var err error
	retryCount := 0

	querySQL.Write([]byte("select id, sdk_generated_id, is_live_customer, debug, analytics, billing, trial, exotic_location_fee, standard_location_fee, public_key, customer_id, looker_seats "))
	querySQL.Write([]byte("from buyers where short_name = $1"))

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), companyCode)
		err = row.Scan(
			&buyer.DatabaseID,
			&buyer.SdkID,
			&buyer.IsLiveCustomer,
			&buyer.Debug,
			&buyer.Analytics,
			&buyer.Billing,
			&buyer.Trial,
			&buyer.ExoticLocationFee,
			&buyer.StandardLocationFee,
			&buyer.PublicKey,
			&buyer.CustomerID,
			&buyer.LookerSeats,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("BuyerWithCompanyCode() connection with the database timed out!")
		return routing.Buyer{}, err
	case sql.ErrNoRows:
		return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer short_name", resourceRef: companyCode}
	case nil:
		buyer.ID = uint64(buyer.SdkID)
		ic, err := db.InternalConfig(ctx, buyer.ID)
		if err != nil {
			ic = core.NewInternalConfig()
		}

		rs, err := db.RouteShader(ctx, buyer.ID)
		if err != nil {
			rs = core.NewRouteShader()
		}

		b := routing.Buyer{
			ID:                  buyer.ID,
			HexID:               fmt.Sprintf("%016x", buyer.ID),
			ShortName:           companyCode,
			CompanyCode:         companyCode,
			Live:                buyer.IsLiveCustomer,
			Debug:               buyer.Debug,
			Analytics:           buyer.Analytics,
			Billing:             buyer.Billing,
			Trial:               buyer.Trial,
			ExoticLocationFee:   buyer.ExoticLocationFee,
			StandardLocationFee: buyer.StandardLocationFee,
			PublicKey:           buyer.PublicKey,
			RouteShader:         rs,
			InternalConfig:      ic,
			CustomerID:          buyer.CustomerID,
			DatabaseID:          buyer.DatabaseID,
			LookerSeats:         buyer.LookerSeats,
		}
		return b, nil
	default:
		core.Error("BuyerWithCompanyCode() QueryRow returned an error: %v", err)
		return routing.Buyer{}, err
	}
}

// Buyers returns a copy of all stored buyers.
func (db *SQL) Buyers(ctx context.Context) []routing.Buyer {
	var sql bytes.Buffer
	var buyer sqlBuyer

	buyers := []routing.Buyer{}
	buyerIDs := make(map[uint64]int64)

	sql.Write([]byte("select sdk_generated_id, id, short_name, is_live_customer, debug, analytics, billing, trial, exotic_location_fee, standard_location_fee, public_key, customer_id, looker_seats "))
	sql.Write([]byte("from buyers"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("Buyers(): QueryMultipleRowsRetry returned an error")
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
			&buyer.Analytics,
			&buyer.Billing,
			&buyer.Trial,
			&buyer.ExoticLocationFee,
			&buyer.StandardLocationFee,
			&buyer.PublicKey,
			&buyer.CustomerID,
			&buyer.LookerSeats,
		)
		if err != nil {
			core.Error("Buyers(): error parsing returned row")
			return []routing.Buyer{}
		}

		buyer.ID = uint64(buyer.SdkID)

		buyerIDs[buyer.ID] = buyer.DatabaseID

		ic, err := db.InternalConfig(ctx, buyer.ID)
		if err != nil {
			ic = core.NewInternalConfig()
		}

		rs, err := db.RouteShader(ctx, buyer.ID)
		if err != nil {
			rs = core.NewRouteShader()
		}

		b := routing.Buyer{
			ID:                  buyer.ID,
			HexID:               fmt.Sprintf("%016x", buyer.ID),
			ShortName:           buyer.ShortName,
			CompanyCode:         buyer.ShortName,
			Live:                buyer.IsLiveCustomer,
			Debug:               buyer.Debug,
			Analytics:           buyer.Analytics,
			Billing:             buyer.Billing,
			Trial:               buyer.Trial,
			ExoticLocationFee:   buyer.ExoticLocationFee,
			StandardLocationFee: buyer.StandardLocationFee,
			PublicKey:           buyer.PublicKey,
			RouteShader:         rs,
			InternalConfig:      ic,
			CustomerID:          buyer.CustomerID,
			DatabaseID:          buyer.DatabaseID,
			LookerSeats:         buyer.LookerSeats,
		}

		buyers = append(buyers, b)

	}

	sort.Slice(buyers, func(i int, j int) bool { return buyers[i].ID < buyers[j].ID })
	return buyers
}

// AddBuyer adds the provided buyer to storage and returns an error if the buyer could not be added.
func (db *SQL) AddBuyer(ctx context.Context, b routing.Buyer) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	c, err := db.Customer(ctx, b.CompanyCode)
	if err != nil {
		return &DoesNotExistError{resourceType: "customer", resourceRef: b.CompanyCode}
	}

	buyer := sqlBuyer{
		ID:                  b.ID,
		CompanyCode:         b.CompanyCode,
		ShortName:           b.CompanyCode,
		IsLiveCustomer:      b.Live,
		Debug:               b.Debug,
		Analytics:           b.Analytics,
		Billing:             b.Billing,
		Trial:               b.Trial,
		ExoticLocationFee:   b.ExoticLocationFee,
		StandardLocationFee: b.StandardLocationFee,
		PublicKey:           b.PublicKey,
		CustomerID:          c.DatabaseID,
		LookerSeats:         b.LookerSeats,
	}

	// Add the buyer in remote storage
	sql.Write([]byte("insert into buyers ("))
	sql.Write([]byte("sdk_generated_id, short_name, is_live_customer, debug, analytics, billing, trial, exotic_location_fee, standard_location_fee, public_key, customer_id, looker_seats"))
	sql.Write([]byte(") values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		int64(buyer.ID),
		buyer.ShortName,
		buyer.IsLiveCustomer,
		buyer.Debug,
		buyer.Analytics,
		buyer.Billing,
		buyer.Trial,
		buyer.ExoticLocationFee,
		buyer.StandardLocationFee,
		buyer.PublicKey,
		buyer.CustomerID,
		buyer.LookerSeats,
	)
	if err != nil {
		core.Error("AddBuyer() error adding buyer: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("AddBuyer() RowsAffected returned an error")
		return err
	}

	if rows != 1 {
		core.Error("AddBuyer() RowsAffected <> 1")
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

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("delete from buyers where sdk_generated_id = $1"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		int64(ephemeralBuyerID),
	)
	if err != nil {
		core.Error("RemoveBuyer() error removing buyer: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveBuyer() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveBuyer() RowsAffected <> 1")
		return err
	}

	return nil
}

// Seller gets a copy of a seller with the specified seller ID,
// and returns an empty seller and an error if a seller with that ID doesn't exist in storage.
func (db *SQL) Seller(ctx context.Context, id string) (routing.Seller, error) {
	var querySQL bytes.Buffer
	var seller sqlSeller
	var row *sql.Row
	var err error
	retryCount := 0

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	querySQL.Write([]byte("select id, short_name, public_egress_price, secret, "))
	querySQL.Write([]byte("customer_id from sellers where short_name = $1"))

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), id)
		err = row.Scan(
			&seller.DatabaseID,
			&seller.ShortName,
			&seller.EgressPriceNibblinsPerGB,
			&seller.Secret,
			&seller.CustomerID,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("Seller() connection with the database timed out!")
		return routing.Seller{}, err
	case sql.ErrNoRows:
		core.Error("Seller() no rows were returned!")
		return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: id}
	case nil:
		c, err := db.Customer(ctx, id)
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
		core.Error("Seller() QueryRow returned an error: %v", err)
		return routing.Seller{}, err
	}
}

// SellerByDbId returns the sellers table entry for the given ID
// TODO: add to storer interface?
func (db *SQL) SellerByDbId(ctx context.Context, id int64) (routing.Seller, error) {
	var querySQL bytes.Buffer
	var seller sqlSeller
	var row *sql.Row
	var err error
	retryCount := 0

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	querySQL.Write([]byte("select short_name, public_egress_price, secret, "))
	querySQL.Write([]byte("customer_id from sellers where id = $1"))

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), id)
		err = row.Scan(
			&seller.ShortName,
			&seller.EgressPriceNibblinsPerGB,
			&seller.Secret,
			&seller.CustomerID,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("SellerByDbId() connection with the database timed out!")
		return routing.Seller{}, err
	case sql.ErrNoRows:
		core.Error("SellerByDbId() no rows were returned!")
		return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: id}
	case nil:
		c, err := db.Customer(ctx, seller.ShortName)
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
		core.Error("SellerByDbId() QueryRow returned an error: %v", err)
		return routing.Seller{}, err
	}
}

// Sellers returns a copy of all stored sellers.
func (db *SQL) Sellers(ctx context.Context) []routing.Seller {
	var sql bytes.Buffer
	var seller sqlSeller
	sellers := []routing.Seller{}

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("select id, short_name, public_egress_price, secret, "))
	sql.Write([]byte("customer_id from sellers"))

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("Sellers(): QueryMultipleRowsRetry returned an error")
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
			core.Error("Sellers(): error parsing returned row")
			return []routing.Seller{}
		}

		c, err := db.Customer(ctx, seller.ShortName)
		if err != nil {
			core.Error("Sellers(): customer does not exist")
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

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	// This check only pertains to the next tool. Stateful clients would already
	// have the customer id.
	c, err := db.Customer(ctx, s.CompanyCode)
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

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		newSellerData.ShortName,
		newSellerData.EgressPriceNibblinsPerGB,
		newSellerData.Secret,
		newSellerData.CustomerID,
	)

	if err != nil {
		core.Error("AddSeller() error adding seller: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("AddSeller() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("AddSeller() RowsAffected <> 1")
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

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("delete from sellers where short_name = $1"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		id,
	)
	if err != nil {
		core.Error("RemoveSeller() error removing seller: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveSeller() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveSeller() RowsAffected <> 1")
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

func (db *SQL) SellerWithCompanyCode(ctx context.Context, code string) (routing.Seller, error) {
	var querySQL bytes.Buffer
	var seller sqlSeller
	var row *sql.Row
	var err error

	retryCount := 0

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	querySQL.Write([]byte("select id, short_name, public_egress_price, secret, "))
	querySQL.Write([]byte("customer_id from sellers where short_name = $1"))

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), code)
		err = row.Scan(
			&seller.DatabaseID,
			&seller.ShortName,
			&seller.EgressPriceNibblinsPerGB,
			&seller.Secret,
			&seller.CustomerID,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("SellerWithCompanyCode() connection with the database timed out!")
		return routing.Seller{}, err
	case sql.ErrNoRows:
		return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: code}
	case nil:
		c, err := db.Customer(ctx, code)
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
		core.Error("SellerWithCompanyCode() QueryRow returned an error: %v", err)
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
func (db *SQL) Relay(ctx context.Context, id uint64) (routing.Relay, error) {
	var sqlQuery bytes.Buffer
	var relay sqlRelay
	var row *sql.Row
	var err error

	retryCount := 0
	hexID := fmt.Sprintf("%016x", id)

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sqlQuery.Write([]byte("select relays.id, relays.display_name, relays.contract_term, relays.end_date, "))
	sqlQuery.Write([]byte("relays.included_bandwidth_gb, relays.management_ip, "))
	sqlQuery.Write([]byte("relays.max_sessions, relays.egress_price_override, relays.mrc, relays.overage, relays.port_speed, relays.max_bandwidth_mbps,"))
	sqlQuery.Write([]byte("relays.public_ip, relays.public_ip_port, relays.public_key, "))
	sqlQuery.Write([]byte("relays.ssh_port, relays.ssh_user, relays.start_date, relays.internal_ip, "))
	sqlQuery.Write([]byte("relays.internal_ip_port, relays.bw_billing_rule, relays.datacenter, "))
	sqlQuery.Write([]byte("relays.machine_type, relays.relay_state, "))
	sqlQuery.Write([]byte("relays.internal_ip, relays.internal_ip_port, relays.notes, "))
	sqlQuery.Write([]byte("relays.billing_supplier, relays.relay_version, relays.dest_first, "))
	sqlQuery.Write([]byte("relays.internal_address_client_routable from relays where hex_id = $1"))

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, sqlQuery.String(), hexID)
		err = row.Scan(&relay.DatabaseID,
			&relay.Name,
			&relay.ContractTerm,
			&relay.EndDate,
			&relay.IncludedBandwithGB,
			&relay.ManagementIP,
			&relay.MaxSessions,
			&relay.EgressPriceOverride,
			&relay.MRC,
			&relay.Overage,
			&relay.NICSpeedMbps,
			&relay.MaxBandwidthMbps,
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
			&relay.DestFirst,
			&relay.InternalAddressClientRoutable,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("Relay() connection with the database timed out!")
		return routing.Relay{}, err
	case sql.ErrNoRows:
		core.Error("Relay() no rows were returned!")
		return routing.Relay{}, &DoesNotExistError{resourceType: "relay", resourceRef: hexID}
	case nil:
		relayState, err := routing.GetRelayStateSQL(relay.State)
		if err != nil {
			core.Error("Relay() invalid relay state: %v", err)
		}

		bwRule, err := routing.GetBandwidthRuleSQL(relay.BWRule)
		if err != nil {
			core.Error("Relay() routing.ParseBandwidthRule returned an error: %v", err)
		}

		machineType, err := routing.GetMachineTypeSQL(relay.MachineType)
		if err != nil {
			core.Error("Relay() routing.ParseMachineType returned an error: %v", err)
		}

		datacenter, err := db.DatacenterByDbId(ctx, relay.DatacenterID)
		if err != nil {
			core.Error("Relay() syncRelays error dereferencing datacenter: %v", err)
		}

		seller, err := db.SellerByDbId(ctx, datacenter.SellerID)
		if err != nil {
			core.Error("Relay() syncRelays error dereferencing seller: %v", err)
		}

		internalID, err := strconv.ParseUint(hexID, 16, 64)
		if err != nil {
			core.Error("Relay() syncRelays error parsing hex_id: %v", err)
		}

		r := routing.Relay{
			ID:                            internalID,
			Name:                          relay.Name,
			PublicKey:                     relay.PublicKey,
			Datacenter:                    datacenter,
			NICSpeedMbps:                  int32(relay.NICSpeedMbps),
			IncludedBandwidthGB:           int32(relay.IncludedBandwithGB),
			MaxBandwidthMbps:              int32(relay.MaxBandwidthMbps),
			State:                         relayState,
			ManagementAddr:                relay.ManagementIP,
			SSHUser:                       relay.SSHUser,
			SSHPort:                       relay.SSHPort,
			MaxSessions:                   uint32(relay.MaxSessions),
			EgressPriceOverride:           routing.Nibblin(relay.EgressPriceOverride),
			MRC:                           routing.Nibblin(relay.MRC),
			Overage:                       routing.Nibblin(relay.Overage),
			BWRule:                        bwRule,
			ContractTerm:                  int32(relay.ContractTerm),
			Type:                          machineType,
			Seller:                        seller,
			DatabaseID:                    relay.DatabaseID,
			Version:                       relay.Version,
			DestFirst:                     relay.DestFirst,
			InternalAddressClientRoutable: relay.InternalAddressClientRoutable,
		}

		// nullable values follow
		if relay.InternalIP.Valid {
			fullInternalAddress := relay.InternalIP.String + ":" + fmt.Sprintf("%d", relay.InternalIPPort.Int64)
			internalAddr, err := net.ResolveUDPAddr("udp", fullInternalAddress)
			if err != nil {
				core.Error("Relay() net.ResolveUDPAddr returned an error parsing internal address: %v", err)
			}
			r.InternalAddr = *internalAddr
		}

		if relay.PublicIP.Valid {
			fullPublicAddress := relay.PublicIP.String + ":" + fmt.Sprintf("%d", relay.PublicIPPort.Int64)
			publicAddr, err := net.ResolveUDPAddr("udp", fullPublicAddress)
			if err != nil {
				core.Error("Relay() net.ResolveUDPAddr returned an error parsing public address: %v", err)
			}
			r.Addr = *publicAddr
		}

		if relay.BillingSupplier.Valid {
			found := false
			for _, seller := range db.Sellers(ctx) {
				if seller.DatabaseID == relay.BillingSupplier.Int64 {
					found = true
					r.BillingSupplier = seller.ID
					break
				}
			}

			if !found {
				errString := fmt.Sprintf("Relay() Unable to find Seller matching BillingSupplier ID %d", relay.BillingSupplier.Int64)
				core.Error(errString, err)
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
		core.Error("Relay() QueryRow returned an error: %v", err)
		return routing.Relay{}, err
	}
}

// Relays returns a copy of all stored relays.
func (db *SQL) Relays(ctx context.Context) []routing.Relay {
	var sqlQuery bytes.Buffer
	var relay sqlRelay

	relays := []routing.Relay{}

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sqlQuery.Write([]byte("select relays.id, relays.hex_id, relays.display_name, relays.contract_term, relays.end_date, "))
	sqlQuery.Write([]byte("relays.included_bandwidth_gb, relays.management_ip, "))
	sqlQuery.Write([]byte("relays.max_sessions, relays.egress_price_override, relays.mrc, relays.overage, relays.port_speed, relays.max_bandwidth_mbps,"))
	sqlQuery.Write([]byte("relays.public_ip, relays.public_ip_port, relays.public_key, "))
	sqlQuery.Write([]byte("relays.ssh_port, relays.ssh_user, relays.start_date, relays.internal_ip, "))
	sqlQuery.Write([]byte("relays.internal_ip_port, relays.bw_billing_rule, relays.datacenter, "))
	sqlQuery.Write([]byte("relays.machine_type, relays.relay_state, "))
	sqlQuery.Write([]byte("relays.internal_ip, relays.internal_ip_port, relays.notes , "))
	sqlQuery.Write([]byte("relays.billing_supplier, relays.relay_version, relays.dest_first, "))
	sqlQuery.Write([]byte("relays.internal_address_client_routable from relays "))

	rows, err := QueryMultipleRowsRetry(ctx, db, sqlQuery)
	if err != nil {
		core.Error("Relays(): QueryMultipleRowsRetry returned an error: %v", err)
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
			&relay.EgressPriceOverride,
			&relay.MRC,
			&relay.Overage,
			&relay.NICSpeedMbps,
			&relay.MaxBandwidthMbps,
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
			&relay.DestFirst,
			&relay.InternalAddressClientRoutable,
		)
		if err != nil {
			core.Error("Relays(): error parsing returned row: %v", err)
			return []routing.Relay{}
		}

		relayState, err := routing.GetRelayStateSQL(relay.State)
		if err != nil {
			core.Error("Relays() invalid relay state: %v", err)
		}

		bwRule, err := routing.GetBandwidthRuleSQL(relay.BWRule)
		if err != nil {
			core.Error("Relays() routing.ParseBandwidthRule returned an error: %v", err)
		}

		machineType, err := routing.GetMachineTypeSQL(relay.MachineType)
		if err != nil {
			core.Error("Relays() routing.ParseMachineType returned an error: %v", err)
		}

		datacenter, err := db.DatacenterByDbId(ctx, relay.DatacenterID)
		if err != nil {
			core.Error("Relays() error dereferencing datacenter: %v", err)
		}

		seller, err := db.SellerByDbId(ctx, datacenter.SellerID)
		if err != nil {
			core.Error("Relays() error dereferencing seller: %v", err)
		}

		internalID, err := strconv.ParseUint(relay.HexID, 16, 64)
		if err != nil {
			core.Error("Relays() error parsing hex_id: %v", err)
		}

		r := routing.Relay{
			ID:                            internalID,
			Name:                          relay.Name,
			PublicKey:                     relay.PublicKey,
			Datacenter:                    datacenter,
			NICSpeedMbps:                  int32(relay.NICSpeedMbps),
			IncludedBandwidthGB:           int32(relay.IncludedBandwithGB),
			MaxBandwidthMbps:              int32(relay.MaxBandwidthMbps),
			State:                         relayState,
			ManagementAddr:                relay.ManagementIP,
			SSHUser:                       relay.SSHUser,
			SSHPort:                       relay.SSHPort,
			MaxSessions:                   uint32(relay.MaxSessions),
			EgressPriceOverride:           routing.Nibblin(relay.EgressPriceOverride),
			MRC:                           routing.Nibblin(relay.MRC),
			Overage:                       routing.Nibblin(relay.Overage),
			BWRule:                        bwRule,
			ContractTerm:                  int32(relay.ContractTerm),
			Type:                          machineType,
			Seller:                        seller,
			DatabaseID:                    relay.DatabaseID,
			Version:                       relay.Version,
			DestFirst:                     relay.DestFirst,
			InternalAddressClientRoutable: relay.InternalAddressClientRoutable,
		}

		// nullable values follow
		if relay.InternalIP.Valid {
			fullInternalAddress := relay.InternalIP.String + ":" + fmt.Sprintf("%d", relay.InternalIPPort.Int64)
			internalAddr, err := net.ResolveUDPAddr("udp", fullInternalAddress)
			if err != nil {
				core.Error("Relays() net.ResolveUDPAddr returned an error parsing internal address: %v", err)
			}
			r.InternalAddr = *internalAddr
		}

		if relay.PublicIP.Valid {
			fullPublicAddress := relay.PublicIP.String + ":" + fmt.Sprintf("%d", relay.PublicIPPort.Int64)
			publicAddr, err := net.ResolveUDPAddr("udp", fullPublicAddress)
			if err != nil {
				core.Error("Relays() net.ResolveUDPAddr returned an error parsing public address: %v", err)
			}
			r.Addr = *publicAddr
		}

		if relay.BillingSupplier.Valid {
			found := false
			for _, seller := range db.Sellers(ctx) {
				if seller.DatabaseID == relay.BillingSupplier.Int64 {
					found = true
					r.BillingSupplier = seller.ID
					break
				}
			}

			if !found {
				errString := fmt.Sprintf("Relays() Unable to find Seller matching BillingSupplier ID %d", relay.BillingSupplier.Int64)
				core.Error(errString, err)
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
//	addr               : ipaddress:port (string)
//  bw_billing_rule    : float64 (json number)
//  machine_type       : float64 (json number)
//  relay_state        : float64 (json number)
//  EgressPriceOverride: USD float64 (json number)
//  MRC                : USD float64 (json number)
//  Overage            : USD float64 (json number)
//  StartDate          : string ('January 2, 2006')
//  EndDate            : string ('January 2, 2006')
//  all others are bool, float64 or string, based on field type
func (db *SQL) UpdateRelay(ctx context.Context, relayID uint64, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	relay, err := db.Relay(ctx, relayID)
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
			if relay.InternalAddressClientRoutable {
				return fmt.Errorf("cannot remove internal address while InternalAddressClientRoutable is true")
			}

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

	case "MaxBandwidthMbps":
		maxBandwidthMbps, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		updateSQL.Write([]byte("update relays set max_bandwidth_mbps=$1 where id=$2"))
		args = append(args, maxBandwidthMbps, relay.DatabaseID)

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

	case "EgressPriceOverride":
		egressPriceOverrideUSD, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		egressPriceOverride := routing.DollarsToNibblins(egressPriceOverrideUSD)
		updateSQL.Write([]byte("update relays set egress_price_override=$1 where id=$2"))
		args = append(args, int64(egressPriceOverride), relay.DatabaseID)

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
			for _, seller := range db.Sellers(ctx) {
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

	case "DestFirst":
		destFirst, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%v is not a valid boolean value", value)
		}

		updateSQL.Write([]byte("update relays set dest_first=$1 where id=$2"))
		args = append(args, destFirst, relay.DatabaseID)

	case "InternalAddressClientRoutable":
		internalAddressClientRoutable, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%v is not a valid boolean value", value)
		}

		if internalAddressClientRoutable && relay.InternalAddr.String() == ":0" {
			// Enforce that the relay has an valid internal address
			return fmt.Errorf("relay must have valid internal address before InternalAddressClientRoutable is true")
		}

		updateSQL.Write([]byte("update relays set internal_address_client_routable=$1 where id=$2"))
		args = append(args, internalAddressClientRoutable, relay.DatabaseID)

	default:
		return fmt.Errorf("field '%v' does not exist on the routing.Relay type", field)

	}

	result, err := ExecRetry(ctx, db, updateSQL, args...)
	if err != nil {
		core.Error("UpdateRelay() error modifying relay record: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("UpdateRelay() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("UpdateRelay() RowsAffected <> 1")
		return err
	}

	return nil
}

type sqlRelay struct {
	ID                            uint64
	HexID                         string
	Name                          string
	PublicIP                      sql.NullString
	PublicIPPort                  sql.NullInt64
	InternalIP                    sql.NullString
	InternalIPPort                sql.NullInt64
	BillingSupplier               sql.NullInt64
	PublicKey                     []byte
	NICSpeedMbps                  int64
	IncludedBandwithGB            int64
	MaxBandwidthMbps              int64
	DatacenterID                  int64
	ManagementIP                  string
	SSHUser                       string
	SSHPort                       int64
	State                         int64
	MaxSessions                   int64
	EgressPriceOverride           int64
	MRC                           int64
	Overage                       int64
	BWRule                        int64
	ContractTerm                  int64
	Notes                         sql.NullString
	StartDate                     sql.NullTime
	EndDate                       sql.NullTime
	MachineType                   int64
	Version                       string
	DestFirst                     bool
	InternalAddressClientRoutable bool
	DatabaseID                    int64
}

// AddRelay adds the provided relay to storage and returns an error if the relay could not be added.
func (db *SQL) AddRelay(ctx context.Context, r routing.Relay) error {
	var sqlQuery bytes.Buffer
	var err error

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	// Routing.Addr is possibly null during syncRelays (due to removed/renamed
	// relays) but *must* have a value when adding a relay
	publicIP := strings.Split(r.Addr.String(), ":")[0]
	publicIPPort, err := strconv.ParseInt(strings.Split(r.Addr.String(), ":")[1], 10, 64)
	if err != nil {
		return fmt.Errorf("unable to convert PublicIP Port %s to int: %v", strings.Split(r.Addr.String(), ":")[1], err)
	}
	rid := crypto.HashID(r.Addr.String())

	relays := db.Relays(ctx)

	// Check that we don't have a relay with the same Hex ID already
	for _, relay := range relays {
		if relay.ID == rid {
			// If a relay with this IP exists already, throw an error
			return fmt.Errorf("relay %s (%016x) (state: %s) already exists with this IP address. please reuse this relay.", relay.Name, relay.ID, relay.State.String())
		}
	}

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
		supplier, err := db.Seller(ctx, r.BillingSupplier)
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
		return fmt.Errorf("relay version cannot be an empty string and must be a valid value (e.g. '2.0.6')")
	}

	// flag is not null but should not be true unless internal address is valid
	if r.InternalAddressClientRoutable && (!internalIP.Valid || !internalIPPort.Valid) {
		return fmt.Errorf("relay flag InternalAddressClientRoutable cannot be true without valid internal IP")
	}

	relay := sqlRelay{
		Name:                          r.Name,
		HexID:                         fmt.Sprintf("%016x", rid),
		PublicIP:                      nullablePublicIP,
		PublicIPPort:                  nullablePublicIPPort,
		InternalIP:                    internalIP,
		InternalIPPort:                internalIPPort,
		PublicKey:                     r.PublicKey,
		NICSpeedMbps:                  int64(r.NICSpeedMbps),
		IncludedBandwithGB:            int64(r.IncludedBandwidthGB),
		MaxBandwidthMbps:              int64(r.MaxBandwidthMbps),
		DatacenterID:                  r.Datacenter.DatabaseID,
		ManagementIP:                  r.ManagementAddr,
		BillingSupplier:               billingSupplier,
		SSHUser:                       r.SSHUser,
		SSHPort:                       r.SSHPort,
		State:                         int64(r.State),
		MaxSessions:                   int64(r.MaxSessions),
		EgressPriceOverride:           int64(r.EgressPriceOverride),
		MRC:                           int64(r.MRC),
		Overage:                       int64(r.Overage),
		BWRule:                        int64(r.BWRule),
		ContractTerm:                  int64(r.ContractTerm),
		StartDate:                     startDate,
		EndDate:                       endDate,
		MachineType:                   int64(r.Type),
		Notes:                         nullableNotes,
		Version:                       r.Version,
		DestFirst:                     r.DestFirst,
		InternalAddressClientRoutable: r.InternalAddressClientRoutable,
	}

	sqlQuery.Write([]byte("insert into relays ("))
	sqlQuery.Write([]byte("hex_id, contract_term, display_name, end_date, included_bandwidth_gb, "))
	sqlQuery.Write([]byte("management_ip, max_sessions, egress_price_override, mrc, overage, port_speed, max_bandwidth_mbps, public_ip, "))
	sqlQuery.Write([]byte("public_ip_port, public_key, ssh_port, ssh_user, start_date, "))
	sqlQuery.Write([]byte("bw_billing_rule, datacenter, machine_type, relay_state, "))
	sqlQuery.Write([]byte("internal_ip, internal_ip_port, notes, billing_supplier, relay_version, dest_first, internal_address_client_routable"))
	sqlQuery.Write([]byte(") values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, "))
	sqlQuery.Write([]byte("$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)"))

	result, err := ExecRetry(
		ctx,
		db,
		sqlQuery,
		relay.HexID,
		relay.ContractTerm,
		relay.Name,
		relay.EndDate,
		relay.IncludedBandwithGB,
		relay.ManagementIP,
		relay.MaxSessions,
		relay.EgressPriceOverride,
		relay.MRC,
		relay.Overage,
		relay.NICSpeedMbps,
		relay.MaxBandwidthMbps,
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
		relay.DestFirst,
		relay.InternalAddressClientRoutable,
	)
	if err != nil {
		core.Error("AddRelay() error adding relay: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("AddRelay() RowsAffected returned an error")
		return err
	}

	if rows != 1 {
		core.Error("AddRelay() RowsAffected <> 1")
		return err
	}

	return nil
}

// RemoveRelay removes a relay with the provided relay ID from storage and
// returns any database errors to the caller
func (db *SQL) RemoveRelay(ctx context.Context, id uint64) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	hexID := fmt.Sprintf("%016x", id)
	sql.Write([]byte("delete from relays where hex_id = $1"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		hexID,
	)
	if err != nil {
		core.Error("RemoveRelay() error removing relay: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveRelay() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveRelay() RowsAffected <> 1")
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

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

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
		Name:                          r.Name,
		PublicIP:                      publicIP,
		PublicIPPort:                  publicIPPort,
		InternalIP:                    internalIP,
		InternalIPPort:                internalIPPort,
		PublicKey:                     r.PublicKey,
		NICSpeedMbps:                  int64(r.NICSpeedMbps),
		IncludedBandwithGB:            int64(r.IncludedBandwidthGB),
		MaxBandwidthMbps:              int64(r.MaxBandwidthMbps),
		DatacenterID:                  r.Datacenter.DatabaseID,
		ManagementIP:                  r.ManagementAddr,
		SSHUser:                       r.SSHUser,
		SSHPort:                       r.SSHPort,
		State:                         int64(r.State),
		MaxSessions:                   int64(r.MaxSessions),
		EgressPriceOverride:           int64(r.EgressPriceOverride),
		MRC:                           int64(r.MRC),
		Overage:                       int64(r.Overage),
		BWRule:                        int64(r.BWRule),
		ContractTerm:                  int64(r.ContractTerm),
		StartDate:                     startDate,
		EndDate:                       endDate,
		MachineType:                   int64(r.Type),
		HexID:                         hexID,
		DestFirst:                     r.DestFirst,
		InternalAddressClientRoutable: r.InternalAddressClientRoutable,
	}

	sqlQuery.Write([]byte("update relays set ("))
	sqlQuery.Write([]byte("hex_id, contract_term, display_name, end_date, included_bandwidth_gb, "))
	sqlQuery.Write([]byte("management_ip, max_sessions, egress_price_override, mrc, overage, port_speed, max_bandwidth_mbps, public_ip, "))
	sqlQuery.Write([]byte("public_ip_port, public_key, ssh_port, ssh_user, start_date, "))
	sqlQuery.Write([]byte("bw_billing_rule, datacenter, machine_type, relay_state, internal_ip, internal_ip_port, "))
	sqlQuery.Write([]byte("dest_first, internal_address_client_routable"))
	sqlQuery.Write([]byte(") = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, "))
	sqlQuery.Write([]byte("$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26) where id = $27"))

	result, err := ExecRetry(
		ctx,
		db,
		sqlQuery,
		relay.HexID,
		relay.ContractTerm,
		relay.Name,
		relay.EndDate,
		relay.IncludedBandwithGB,
		relay.ManagementIP,
		relay.MaxSessions,
		relay.EgressPriceOverride,
		relay.MRC,
		relay.Overage,
		relay.NICSpeedMbps,
		relay.MaxBandwidthMbps,
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
		relay.DestFirst,
		relay.InternalAddressClientRoutable,
		r.DatabaseID,
	)
	if err != nil {
		core.Error("SetRelay() error modifying relay: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("SetRelay() RowsAffected returned an error")
		return err
	}

	if rows != 1 {
		core.Error("SetRelay() RowsAffected <> 1")
		return err
	}

	return nil
}

// Datacenter gets a copy of a datacenter with the specified datacenter ID
// and returns an empty datacenter and an error if a datacenter with that ID doesn't exist in storage.
func (db *SQL) Datacenter(ctx context.Context, datacenterID uint64) (routing.Datacenter, error) {
	var querySQL bytes.Buffer
	var dc sqlDatacenter
	var row *sql.Row
	var err error

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	retryCount := 0
	hexID := fmt.Sprintf("%016x", datacenterID)

	querySQL.Write([]byte("select id, display_name, latitude, longitude,"))
	querySQL.Write([]byte("seller_id from datacenters where hex_id = $1"))

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), hexID)

		err = row.Scan(
			&dc.ID,
			&dc.Name,
			&dc.Latitude,
			&dc.Longitude,
			&dc.SellerID,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("Datacenter() connection with database timed out!")
		return routing.Datacenter{}, err
	case sql.ErrNoRows:
		core.Error("Datacenter() no rows were returned!")
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
		core.Error("Datacenter() QueryRow returned an error: %v", err)
		return routing.Datacenter{}, err
	}
}

// DatacenterByDbId retrives the entry in the datacenters table for the provided ID
// TODO: add to storer interface?
func (db *SQL) DatacenterByDbId(ctx context.Context, databaseID int64) (routing.Datacenter, error) {
	var querySQL bytes.Buffer
	var dc sqlDatacenter
	var row *sql.Row
	var err error

	retryCount := 0

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	querySQL.Write([]byte("select hex_id, display_name, latitude, longitude,"))
	querySQL.Write([]byte("seller_id from datacenters where id = $1"))

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), databaseID)
		err = row.Scan(
			&dc.HexID,
			&dc.Name,
			&dc.Latitude,
			&dc.Longitude,
			&dc.SellerID,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("DatacenterByDbId() connection with the database timed out!")
		return routing.Datacenter{}, err
	case sql.ErrNoRows:
		core.Error("DatacenterByDbId() no rows were returned!")
		return routing.Datacenter{}, &DoesNotExistError{resourceType: "datacenter", resourceRef: fmt.Sprintf("%d", databaseID)}
	case nil:
		datacenterID, err := strconv.ParseUint(dc.HexID, 16, 64)
		if err != nil {
			core.Error("DatacenterByDbId() error parsing hex ID")
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
		core.Error("DatacenterByDbId() QueryRow returned an error: %v", err)
		return routing.Datacenter{}, err
	}
}

// Datacenters returns a copy of all stored datacenters.
func (db *SQL) Datacenters(ctx context.Context) []routing.Datacenter {
	var sql bytes.Buffer
	var dc sqlDatacenter

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	datacenters := []routing.Datacenter{}

	sql.Write([]byte("select id, display_name, latitude, longitude,"))
	sql.Write([]byte("seller_id from datacenters"))

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("Datacenters(): QueryContext returned an error: %v", err)
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
			core.Error("Datacenters(): error parsing returned row: %v", err)
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

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	hexID := fmt.Sprintf("%016x", id)
	sql.Write([]byte("delete from datacenters where hex_id = $1"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		hexID,
	)
	if err != nil {
		core.Error("RemoveDatacenter() error removing datacenter: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveDatacenter() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveDatacenter() RowsAffected <> 1")
		return err
	}

	return nil
}

// GetDatacenterMapsForBuyer returns a map of datacenter aliases in use for a given
// (internally generated) buyerID. The map is indexed by the datacenter ID. Returns
// an empty map if there are no aliases for that buyerID.
func (db *SQL) GetDatacenterMapsForBuyer(ctx context.Context, ephemeralBuyerID uint64) map[uint64]routing.DatacenterMap {
	var querySQL bytes.Buffer
	var dcMaps = make(map[uint64]routing.DatacenterMap)

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	dbBuyerID := int64(ephemeralBuyerID)

	querySQL.Write([]byte("select datacenters.hex_id from datacenter_maps "))
	querySQL.Write([]byte("inner join datacenters on datacenter_maps.datacenter_id "))
	querySQL.Write([]byte("= datacenters.id where datacenter_maps.buyer_id = "))
	querySQL.Write([]byte("(select id from buyers where sdk_generated_id = $1)"))

	rows, err := QueryMultipleRowsRetry(ctx, db, querySQL, dbBuyerID)
	if err != nil {
		core.Error("GetDatacenterMapsForBuyer(): QueryMultipleRowsRetry returned an error: %v", err)
		return map[uint64]routing.DatacenterMap{}
	}
	defer rows.Close()

	for rows.Next() {
		var hexID string
		err = rows.Scan(&hexID)
		if err != nil {
			core.Error("GetDatacenterMapsForBuyer(): error parsing returned row: %v", err)
			return map[uint64]routing.DatacenterMap{}
		}

		dcID, err := strconv.ParseUint(hexID, 16, 64)
		if err != nil {
			core.Error("GetDatacenterMapsForBuyer() error parsing datacenter hex ID")
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

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	buyer, err := db.Buyer(ctx, dcMap.BuyerID)
	if err != nil {
		return &DoesNotExistError{resourceType: "Buyer.ID", resourceRef: fmt.Sprintf("%016x", dcMap.BuyerID)}
	}

	datacenter, err := db.Datacenter(ctx, dcMap.DatacenterID)
	if err != nil {
		return &DoesNotExistError{resourceType: "DatacenterID", resourceRef: dcMap.DatacenterID}
	}

	sql.Write([]byte("insert into datacenter_maps (buyer_id, datacenter_id) "))
	sql.Write([]byte("values ($1, $2)"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		buyer.DatabaseID,
		datacenter.DatabaseID,
	)

	if err != nil {
		core.Error("AddDatacenterMap() error adding DatacenterMap: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("AddDatacenterMap() RowsAffected returned an error")
		return err
	}

	if rows != 1 {
		core.Error("AddDatacenterMap() RowsAffected <> 1")
		return err
	}

	return nil
}

// ListDatacenterMaps returns a list of alias/buyer mappings for the specified datacenter ID. An
// empty dcID returns a list of all maps.
func (db *SQL) ListDatacenterMaps(ctx context.Context, dcID uint64) map[uint64]routing.DatacenterMap {
	var querySQL bytes.Buffer
	var dcMaps = make(map[uint64]routing.DatacenterMap)
	var sqlMap sqlDatacenterMap

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	hexID := fmt.Sprintf("%016x", dcID)

	querySQL.Write([]byte("select sdk_generated_id from buyers where id in ( "))
	querySQL.Write([]byte("select buyer_id from datacenter_maps where datacenter_id = ( "))
	querySQL.Write([]byte("select id from datacenters where hex_id = $1 )) "))

	rows, err := QueryMultipleRowsRetry(ctx, db, querySQL, hexID)
	if err != nil {
		core.Error("ListDatacenterMaps(): QueryContext returned an error: %v", err)
		return map[uint64]routing.DatacenterMap{}
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&sqlMap.BuyerID)
		if err != nil {
			core.Error("ListDatacenterMaps(): error parsing returned row: %v", err)
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

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("delete from datacenter_maps where buyer_id = "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $1) "))
	sql.Write([]byte("and datacenter_id = (select id from datacenters where hex_id = $2)"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		int64(dcMap.BuyerID),
		fmt.Sprintf("%016x", dcMap.DatacenterID),
	)
	if err != nil {
		core.Error("RemoveDatacenterMap() error removing datacenter map: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveDatacenterMap() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveDatacenterMap() RowsAffected <> 1")
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

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

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

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		dc.Name,
		dc.HexID,
		dc.Latitude,
		dc.Longitude,
		dc.SellerID,
	)

	if err != nil {
		core.Error("RemoveDatacenterMap() error adding datacenter: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveDatacenterMap() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveDatacenterMap() RowsAffected <> 1")
		return err
	}

	return nil
}

// RouteShaders returns a slice of route shaders for the given buyer ID
func (db *SQL) RouteShader(ctx context.Context, ephemeralBuyerID uint64) (core.RouteShader, error) {
	var querySQL bytes.Buffer
	var sqlRS sqlRouteShader
	var row *sql.Row
	var err error

	retryCount := 0

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	querySQL.Write([]byte("select ab_test, acceptable_latency, acceptable_packet_loss, analysis_only, bw_envelope_down_kbps, "))
	querySQL.Write([]byte("bw_envelope_up_kbps, disable_network_next, latency_threshold, multipath, pro_mode, "))
	querySQL.Write([]byte("reduce_latency, reduce_packet_loss, reduce_jitter, selection_percent, "))
	querySQL.Write([]byte("packet_loss_sustained from route_shaders where buyer_id = ( "))
	querySQL.Write([]byte("select id from buyers where sdk_generated_id = $1)"))

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), int64(ephemeralBuyerID))
		err = row.Scan(
			&sqlRS.ABTest,
			&sqlRS.AcceptableLatency,
			&sqlRS.AcceptablePacketLoss,
			&sqlRS.AnalysisOnly,
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
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("RouteShader() connection with the database timed out!")
		return core.RouteShader{}, err
	case sql.ErrNoRows:
		// By default buyers do not have a custom route shader so will not
		// have an entry in the route_shaders table. We probably don't need
		// to log an error here. However, the return is checked and the
		// default core.RouteShader is applied to the buyer if an error
		// is returned.
		// core.Error("RouteShader() no rows were returned!")
		return core.RouteShader{}, &DoesNotExistError{resourceType: "buyer", resourceRef: fmt.Sprintf("%016x", ephemeralBuyerID)}
	case nil:
		routeShader := core.RouteShader{
			DisableNetworkNext:        sqlRS.DisableNetworkNext,
			AnalysisOnly:              sqlRS.AnalysisOnly,
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

		bannedUserList, err := db.BannedUsers(ctx, ephemeralBuyerID)
		if err != nil {
			core.Error("RouteShader() -> BannedUsers() returned an error")
			return core.RouteShader{}, fmt.Errorf("RouteShader() -> BannedUser() returned an error: %v for Buyer %s", err, fmt.Sprintf("%016x", ephemeralBuyerID))
		}
		routeShader.BannedUsers = bannedUserList
		return routeShader, nil
	default:
		core.Error("RouteShader() QueryRow returned an error: %v", err)
		return core.RouteShader{}, err
	}
}

// InternalConfig returns the InternalConfig entry for the specified buyer
func (db *SQL) InternalConfig(ctx context.Context, ephemeralBuyerID uint64) (core.InternalConfig, error) {
	var querySQL bytes.Buffer
	var sqlIC sqlInternalConfig
	var row *sql.Row
	var err error

	retryCount := 0

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	querySQL.Write([]byte("select max_latency_tradeoff, max_rtt, multipath_overload_threshold, "))
	querySQL.Write([]byte("route_switch_threshold, route_select_threshold, rtt_veto_default, "))
	querySQL.Write([]byte("rtt_veto_multipath, rtt_veto_packetloss, try_before_you_buy, force_next, "))
	querySQL.Write([]byte("large_customer, is_uncommitted, high_frequency_pings, route_diversity, "))
	querySQL.Write([]byte("multipath_threshold, enable_vanity_metrics, reduce_pl_min_slice_number "))
	querySQL.Write([]byte("from rs_internal_configs where buyer_id = ( "))
	querySQL.Write([]byte("select id from buyers where sdk_generated_id = $1)"))

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), int64(ephemeralBuyerID))
		err = row.Scan(
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
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("InternalConfig() connection with the database timed out!")
		return core.InternalConfig{}, err
	case sql.ErrNoRows:
		// By default buyers do not have a custom internal config so will not
		// have an entry in the rs_internal_configs table. We probably don't need
		// to log an error here. However, the return is checked and the
		// default core.InternalConfig is applied to the buyer if an error
		// is returned.
		// core.Error("InternalConfig() no rows were returned!")
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
		core.Error("InternalConfig() QueryRow returned an error: %v", err)
		return core.InternalConfig{}, err
	}

}

// AddInternalConfig adds an InternalConfig for the specified buyer
func (db *SQL) AddInternalConfig(ctx context.Context, ic core.InternalConfig, ephemeralBuyerID uint64) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

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

	result, err := ExecRetry(
		ctx,
		db,
		sql,
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
		core.Error("AddInternalConfig() error adding internal config: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("AddInternalConfig() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("AddInternalConfig() RowsAffected <> 1")
		return err
	}

	return nil
}

func (db *SQL) RemoveInternalConfig(ctx context.Context, ephemeralBuyerID uint64) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	buyerID := int64(ephemeralBuyerID)
	sql.Write([]byte("delete from rs_internal_configs where buyer_id = "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $1)"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		buyerID,
	)
	if err != nil {
		core.Error("RemoveInternalConfig() error removing internal config: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveInternalConfig() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveInternalConfig() RowsAffected <> 1")
		return err
	}

	return nil
}

func (db *SQL) UpdateInternalConfig(ctx context.Context, ephemeralBuyerID uint64, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}
	var err error

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	switch field {
	case "RouteSelectThreshold":
		routeSelectThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteSelectThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set route_select_threshold=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, routeSelectThreshold, int64(ephemeralBuyerID))
	case "RouteSwitchThreshold":
		routeSwitchThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteSwitchThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set route_switch_threshold=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, routeSwitchThreshold, int64(ephemeralBuyerID))
	case "MaxLatencyTradeOff":
		maxLatencyTradeOff, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MaxLatencyTradeOff: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set max_latency_tradeoff=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, maxLatencyTradeOff, int64(ephemeralBuyerID))
	case "RTTVeto_Default":
		rttVetoDefault, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_Default: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_default=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, rttVetoDefault, int64(ephemeralBuyerID))
	case "RTTVeto_PacketLoss":
		rttVetoPacketLoss, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_PacketLoss: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_packetloss=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, rttVetoPacketLoss, int64(ephemeralBuyerID))
	case "RTTVeto_Multipath":
		rttVetoMultipath, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_Multipath: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set rtt_veto_multipath=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, rttVetoMultipath, int64(ephemeralBuyerID))
	case "MultipathOverloadThreshold":
		multipathOverloadThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MultipathOverloadThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set multipath_overload_threshold=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, multipathOverloadThreshold, int64(ephemeralBuyerID))
	case "TryBeforeYouBuy":
		tryBeforeYouBuy, ok := value.(bool)
		if !ok {
			return fmt.Errorf("TryBeforeYouBuy: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set try_before_you_buy=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, tryBeforeYouBuy, int64(ephemeralBuyerID))
	case "ForceNext":
		forceNext, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ForceNext: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set force_next=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, forceNext, int64(ephemeralBuyerID))
	case "LargeCustomer":
		largeCustomer, ok := value.(bool)
		if !ok {
			return fmt.Errorf("LargeCustomer: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set large_customer=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, largeCustomer, int64(ephemeralBuyerID))
	case "Uncommitted":
		uncommitted, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Uncommitted: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set is_uncommitted=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, uncommitted, int64(ephemeralBuyerID))
	case "HighFrequencyPings":
		highFrequencyPings, ok := value.(bool)
		if !ok {
			return fmt.Errorf("HighFrequencyPings: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set high_frequency_pings=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, highFrequencyPings, int64(ephemeralBuyerID))
	case "MaxRTT":
		maxRTT, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MaxRTT: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set max_rtt=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, maxRTT, int64(ephemeralBuyerID))
	case "RouteDiversity":
		routeDiversity, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteDiversity: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set route_diversity=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, routeDiversity, int64(ephemeralBuyerID))
	case "MultipathThreshold":
		multipathThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MultipathThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set multipath_threshold=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, multipathThreshold, int64(ephemeralBuyerID))
	case "EnableVanityMetrics":
		enableVanityMetrics, ok := value.(bool)
		if !ok {
			return fmt.Errorf("EnableVanityMetrics: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set enable_vanity_metrics=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, enableVanityMetrics, int64(ephemeralBuyerID))
	case "ReducePacketLossMinSliceNumber":
		reducePacketLossMinSliceNumber, ok := value.(int32)
		if !ok {
			return fmt.Errorf("ReducePacketLossMinSliceNumber: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update rs_internal_configs set reduce_pl_min_slice_number=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, reducePacketLossMinSliceNumber, int64(ephemeralBuyerID))

	default:
		return fmt.Errorf("Field '%v' does not exist on the InternalConfig type", field)
	}

	result, err := ExecRetry(
		ctx,
		db,
		updateSQL,
		args...,
	)
	if err != nil {
		core.Error("UpdateInternalConfig() error modifying internal_config record: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("UpdateInternalConfig() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("UpdateInternalConfig() RowsAffected <> 1")
		return err
	}

	return nil
}

func (db *SQL) AddRouteShader(ctx context.Context, rs core.RouteShader, ephemeralBuyerID uint64) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	routeShader := sqlRouteShader{
		ABTest:                    rs.ABTest,
		AcceptableLatency:         int64(rs.AcceptableLatency),
		AcceptablePacketLoss:      float64(rs.AcceptablePacketLoss),
		AnalysisOnly:              rs.AnalysisOnly,
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
	sql.Write([]byte("ab_test, acceptable_latency, acceptable_packet_loss, analysis_only, bw_envelope_down_kbps, "))
	sql.Write([]byte("bw_envelope_up_kbps, disable_network_next, latency_threshold, multipath, "))
	sql.Write([]byte("pro_mode, reduce_latency, reduce_packet_loss, reduce_jitter, selection_percent, packet_loss_sustained, buyer_id"))
	sql.Write([]byte(") values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $16))"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		routeShader.ABTest,
		routeShader.AcceptableLatency,
		routeShader.AcceptablePacketLoss,
		routeShader.AnalysisOnly,
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
		core.Error("AddRouteShader() error adding route shader: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("AddRouteShader() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("AddRouteShader() RowsAffected <> 1")
		return err
	}

	return nil

}

func (db *SQL) UpdateRouteShader(ctx context.Context, ephemeralBuyerID uint64, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}
	var err error

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	switch field {
	case "ABTest":
		abTest, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ABTest: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set ab_test=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, abTest, int64(ephemeralBuyerID))
	case "AcceptableLatency":
		acceptableLatency, ok := value.(int32)
		if !ok {
			return fmt.Errorf("AcceptableLatency: %v is not a valid int32 type ( %T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set acceptable_latency=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, acceptableLatency, int64(ephemeralBuyerID))
	case "AcceptablePacketLoss":
		acceptablePacketLoss, ok := value.(float32)
		if !ok {
			return fmt.Errorf("AcceptablePacketLoss: %v is not a valid float32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set acceptable_packet_loss=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, acceptablePacketLoss, int64(ephemeralBuyerID))
	case "AnalysisOnly":
		analysisOnly, ok := value.(bool)
		if !ok {
			return fmt.Errorf("AnalysisOnly: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set analysis_only=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, analysisOnly, int64(ephemeralBuyerID))
	case "BandwidthEnvelopeDownKbps":
		bandwidthEnvelopeDownKbps, ok := value.(int32)
		if !ok {
			return fmt.Errorf("BandwidthEnvelopeDownKbps: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set bw_envelope_down_kbps=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, bandwidthEnvelopeDownKbps, int64(ephemeralBuyerID))
	case "BandwidthEnvelopeUpKbps":
		bandwidthEnvelopeUpKbps, ok := value.(int32)
		if !ok {
			return fmt.Errorf("BandwidthEnvelopeUpKbps: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set bw_envelope_up_kbps=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, bandwidthEnvelopeUpKbps, int64(ephemeralBuyerID))
	case "DisableNetworkNext":
		disableNetworkNext, ok := value.(bool)
		if !ok {
			return fmt.Errorf("DisableNetworkNext: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set disable_network_next=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, disableNetworkNext, int64(ephemeralBuyerID))
	case "LatencyThreshold":
		latencyThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("LatencyThreshold: %v is not a valid int32 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set latency_threshold=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, latencyThreshold, int64(ephemeralBuyerID))
	case "Multipath":
		multipath, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Multipath: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set multipath=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, multipath, int64(ephemeralBuyerID))
	case "ProMode":
		proMode, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ProMode: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set pro_mode=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, proMode, int64(ephemeralBuyerID))
	case "ReduceLatency":
		reduceLatency, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReduceLatency: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set reduce_latency=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, reduceLatency, int64(ephemeralBuyerID))
	case "ReducePacketLoss":
		reducePacketLoss, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReducePacketLoss: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set reduce_packet_loss=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, reducePacketLoss, int64(ephemeralBuyerID))
	case "ReduceJitter":
		reduceJitter, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReduceJitter: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set reduce_jitter=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, reduceJitter, int64(ephemeralBuyerID))
	case "SelectionPercent":
		selectionPercent, ok := value.(int)
		if !ok {
			return fmt.Errorf("SelectionPercent: %v is not a valid int type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set selection_percent=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, selectionPercent, int64(ephemeralBuyerID))
	case "PacketLossSustained":
		packetLossSustained, ok := value.(float32)
		if !ok {
			return fmt.Errorf("PacketLossSustained: %v is not a valid float type (%T)", value, value)
		}
		updateSQL.Write([]byte("update route_shaders set packet_loss_sustained=$1 where buyer_id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, packetLossSustained, int64(ephemeralBuyerID))
	default:
		return fmt.Errorf("Field '%v' does not exist on the RouteShader type", field)

	}

	result, err := ExecRetry(
		ctx,
		db,
		updateSQL,
		args...,
	)
	if err != nil {
		core.Error("UpdateRouteShader() error modifying route_shader record: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("UpdateRouteShader() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("UpdateRouteShader() RowsAffected <> 1")
		return err
	}

	return nil
}

func (db *SQL) RemoveRouteShader(ctx context.Context, ephemeralBuyerID uint64) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("delete from route_shaders where buyer_id = "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $1)"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		int64(ephemeralBuyerID),
	)

	if err != nil {
		core.Error("RemoveRouteShader() error removing route shader: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveRouteShader() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveRouteShader() RowsAffected <> 1")
		return err
	}

	return nil
}

// AddBannedUser adds a user to the banned_user table
func (db *SQL) AddBannedUser(ctx context.Context, ephemeralBuyerID uint64, userID uint64) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("insert into banned_users (user_id, buyer_id) values ($1, "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $2))"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		int64(userID), int64(ephemeralBuyerID),
	)
	if err != nil {
		core.Error("AddBannedUser() error adding banned user: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("AddBannedUser() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("AddBannedUser() RowsAffected <> 1")
		return err
	}

	return nil

}

// RemoveBannedUser removes a user from the banned_user table
func (db *SQL) RemoveBannedUser(ctx context.Context, ephemeralBuyerID uint64, userID uint64) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("delete from banned_users where user_id = $1 and buyer_id = "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		int64(userID),
		int64(ephemeralBuyerID),
	)
	if err != nil {
		core.Error("RemoveBannedUser() error removing banned user: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveBannedUser() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveBannedUser() RowsAffected <> 1")
		return err
	}

	return nil

}

// BannedUsers returns the set of banned users for the specified buyer ID.
func (db *SQL) BannedUsers(ctx context.Context, ephemeralBuyerID uint64) (map[uint64]bool, error) {
	var sql bytes.Buffer
	bannedUserList := make(map[uint64]bool)

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("select user_id from banned_users where buyer_id = "))
	sql.Write([]byte("(select id from buyers where sdk_generated_id = $1)"))

	rows, err := QueryMultipleRowsRetry(ctx, db, sql, int64(ephemeralBuyerID))
	if err != nil {
		core.Error("BannedUsers(): QueryContext returned an error: %v", err)
		return bannedUserList, err
	}
	defer rows.Close()

	for rows.Next() {
		var userID int64
		err := rows.Scan(&userID)
		if err != nil {
			core.Error("BannedUsers() error parsing user and buyer IDs: %v", err)
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

func (db *SQL) GetFeatureFlags(ctx context.Context) map[string]bool {
	return map[string]bool{}
}

func (db *SQL) GetFeatureFlagByName(ctx context.Context, flagName string) (map[string]bool, error) {
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
	var err error

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	switch field {
	case "Live":
		live, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Live: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set is_live_customer=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, live, int64(ephemeralBuyerID))
	case "Debug":
		debug, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Debug: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set debug=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, debug, int64(ephemeralBuyerID))
	case "Analytics":
		analytics, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Analytics: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set analytics=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, analytics, int64(ephemeralBuyerID))
	case "Billing":
		billing, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Billing: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set billing=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, billing, int64(ephemeralBuyerID))
	case "Trial":
		trial, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Trial: %v is not a valid boolean type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set trial=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, trial, int64(ephemeralBuyerID))
	case "ExoticLocationFee":
		exoticLocationFee, ok := value.(float64)
		if !ok {
			return fmt.Errorf("ExoticLocationFee: %v is not a valid float64 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set exotic_location_fee=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, exoticLocationFee, int64(ephemeralBuyerID))
	case "StandardLocationFee":
		standardLocationFee, ok := value.(float64)
		if !ok {
			return fmt.Errorf("StandardLocationFee: %v is not a valid float64 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set standard_location_fee=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, standardLocationFee, int64(ephemeralBuyerID))
	case "LookerSeats":
		lookerSeats, ok := value.(int64)
		if !ok {
			return fmt.Errorf("LookerSeats: %v is not a valid int64 type (%T)", value, value)
		}
		updateSQL.Write([]byte("update buyers set looker_seats=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, lookerSeats, int64(ephemeralBuyerID))
	case "ShortName":
		shortName, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		updateSQL.Write([]byte("update buyers set short_name=$1 where id="))
		updateSQL.Write([]byte("(select id from buyers where sdk_generated_id = $2)"))
		args = append(args, shortName, int64(ephemeralBuyerID))
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
		args = append(args, newPublicKey[8:], int64(newBuyerID), int64(ephemeralBuyerID))

		// TODO: datacenter maps for this buyer must be updated with the new buyer ID

	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Buyer type", field)

	}

	result, err := ExecRetry(ctx, db, updateSQL, args...)
	if err != nil {
		core.Error("UpdateBuyer() error modifying buyer record: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("UpdateBuyer() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("UpdateBuyer() RowsAffected <> 1")
		return err
	}

	return nil
}

func (db *SQL) UpdateCustomer(ctx context.Context, customerID string, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	customer, err := db.Customer(ctx, customerID)
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

	result, err := ExecRetry(ctx, db, updateSQL, args...)
	if err != nil {
		core.Error("UpdateCustomer() error modifying customer record: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("UpdateCustomer() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("UpdateCustomer() RowsAffected <> 1")
		return err
	}

	return nil
}

func (db *SQL) UpdateSeller(ctx context.Context, sellerID string, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	seller, err := db.Seller(ctx, sellerID)
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

	result, err := ExecRetry(ctx, db, updateSQL, args...)
	if err != nil {
		core.Error("UpdateSeller() error modifying seller record: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("UpdateSeller() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("UpdateSeller() RowsAffected <> 1")
		return err
	}

	return nil
}

func (db *SQL) UpdateDatacenter(ctx context.Context, datacenterID uint64, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	datacenter, err := db.Datacenter(ctx, datacenterID)
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

	result, err := ExecRetry(ctx, db, updateSQL, args...)
	if err != nil {
		core.Error("UpdateDatacenter() error modifying datacenter record: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("UpdateDatacenter() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("UpdateDatacenter() RowsAffected <> 1")
		return err
	}

	return nil
}

func (db *SQL) GetDatabaseBinFileMetaData(ctx context.Context) (routing.DatabaseBinFileMetaData, error) {
	var querySQL bytes.Buffer
	var dashboardData routing.DatabaseBinFileMetaData
	var row *sql.Row
	var err error

	retryCount := 0

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	querySQL.Write([]byte("select bin_file_creation_time, bin_file_author "))
	querySQL.Write([]byte("from database_bin_meta order by bin_file_creation_time desc limit 1"))

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String())
		err = row.Scan(&dashboardData.DatabaseBinFileCreationTime, &dashboardData.DatabaseBinFileAuthor)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("GetFleetDashboardData() connection with the database timed out!")
		return routing.DatabaseBinFileMetaData{}, err
	case sql.ErrNoRows:
		core.Error("GetFleetDashboardData() no rows were returned!")
		return routing.DatabaseBinFileMetaData{}, err
	case nil:
		return dashboardData, nil
	default:
		core.Error("GetFleetDashboardData() QueryRow returned an error: %v", err)
		return routing.DatabaseBinFileMetaData{}, err
	}
}

func (db *SQL) UpdateDatabaseBinFileMetaData(ctx context.Context, metaData routing.DatabaseBinFileMetaData) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	// Add the metadata record to the database_bin_meta table
	sql.Write([]byte("insert into database_bin_meta ("))
	sql.Write([]byte("bin_file_creation_time, bin_file_author, sha"))
	sql.Write([]byte(") values ($1, $2, $3)"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		metaData.DatabaseBinFileCreationTime,
		metaData.DatabaseBinFileAuthor,
		"",
	)

	if err != nil {
		core.Error("UpdateDatabaseBinFileMetaData() error adding DatabaseBinFileMetaData: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("UpdateDatabaseBinFileMetaData() RowsAffected returned an error: %v", err)
		return err
	}
	if rows != 1 {
		core.Error("UpdateDatabaseBinFileMetaData() RowsAffected <> 1: %v", err)
		return err
	}

	return nil
}

// GetAnalyticsDashboardCategories returns all Looker dashboard categories
func (db *SQL) GetAnalyticsDashboardCategories(ctx context.Context) ([]looker.AnalyticsDashboardCategory, error) {
	var sql bytes.Buffer
	category := looker.AnalyticsDashboardCategory{}
	categories := make([]looker.AnalyticsDashboardCategory, 0)

	sql.Write([]byte("select id, tab_label, premium, admin_only, seller_only "))
	sql.Write([]byte("from analytics_dashboard_categories"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("GetAnalyticsDashboardCategories(): QueryMultipleRowsRetry returned an error: %v", err)
		return []looker.AnalyticsDashboardCategory{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&category.ID,
			&category.Label,
			&category.Premium,
			&category.Admin,
			&category.Seller,
		)
		if err != nil {
			core.Error("GetAnalyticsDashboardCategories(): error parsing returned row: %v", err)
			return []looker.AnalyticsDashboardCategory{}, err
		}

		categories = append(categories, category)
	}

	sort.Slice(categories, func(i int, j int) bool { return categories[i].ID < categories[j].ID })
	return categories, nil
}

// GetPremiumAnalyticsDashboardCategories returns all Looker dashboard categories
func (db *SQL) GetPremiumAnalyticsDashboardCategories(ctx context.Context) ([]looker.AnalyticsDashboardCategory, error) {
	var sql bytes.Buffer
	category := looker.AnalyticsDashboardCategory{}
	categories := make([]looker.AnalyticsDashboardCategory, 0)

	sql.Write([]byte("select id, tab_label, premium, seller_only "))
	sql.Write([]byte("from analytics_dashboard_categories where premium = true"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("GetPremiumAnalyticsDashboardCategories(): QueryMultipleRowsRetry returned an error: %v", err)
		return []looker.AnalyticsDashboardCategory{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&category.ID,
			&category.Label,
			&category.Premium,
			&category.Seller,
		)
		if err != nil {
			core.Error("GetPremiumAnalyticsDashboardCategories(): error parsing returned row: %v", err)
			return []looker.AnalyticsDashboardCategory{}, err
		}

		categories = append(categories, category)
	}

	sort.Slice(categories, func(i int, j int) bool { return categories[i].ID < categories[j].ID })
	return categories, nil
}

// GetFreeAnalyticsDashboardCategories returns all Looker dashboard categories
func (db *SQL) GetFreeAnalyticsDashboardCategories(ctx context.Context) ([]looker.AnalyticsDashboardCategory, error) {
	var sql bytes.Buffer
	category := looker.AnalyticsDashboardCategory{}
	categories := make([]looker.AnalyticsDashboardCategory, 0)

	sql.Write([]byte("select id, tab_label, premium, seller_only "))
	sql.Write([]byte("from analytics_dashboard_categories where premium = false"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("GetFreeAnalyticsDashboardCategories(): QueryMultipleRowsRetry returned an error: %v", err)
		return []looker.AnalyticsDashboardCategory{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&category.ID,
			&category.Label,
			&category.Premium,
			&category.Seller,
		)
		if err != nil {
			core.Error("GetFreeAnalyticsDashboardCategories(): error parsing returned row: %v", err)
			return []looker.AnalyticsDashboardCategory{}, err
		}

		categories = append(categories, category)
	}

	sort.Slice(categories, func(i int, j int) bool { return categories[i].ID < categories[j].ID })
	return categories, nil
}

// GetAnalyticsDashboardCategories returns all Looker dashboard categories
func (db *SQL) GetAnalyticsDashboardCategoryByID(ctx context.Context, id int64) (looker.AnalyticsDashboardCategory, error) {
	var querySQL bytes.Buffer
	var row *sql.Row
	var err error

	category := looker.AnalyticsDashboardCategory{}
	retryCount := 0

	querySQL.Write([]byte("select id, tab_label, premium, admin_only, seller_only from analytics_dashboard_categories where id = $1"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), id)
		err = row.Scan(
			&category.ID,
			&category.Label,
			&category.Premium,
			&category.Admin,
			&category.Seller,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("GetAnalyticsDashboardCategoryByID() connection with the database timed out!")
		return looker.AnalyticsDashboardCategory{}, err
	case sql.ErrNoRows:
		core.Error("GetAnalyticsDashboardCategoryByID() no rows were returned!")
		return looker.AnalyticsDashboardCategory{}, &DoesNotExistError{resourceType: "analytics dashboard category", resourceRef: id}
	case nil:
		return category, nil
	default:
		core.Error("GetAnalyticsDashboardCategoryByID() QueryRow returned an error: %v", err)
		return looker.AnalyticsDashboardCategory{}, err
	}
}

// GetAnalyticsDashboardCategories returns all Looker dashboard categories
func (db *SQL) GetAnalyticsDashboardCategoryByLabel(ctx context.Context, label string) (looker.AnalyticsDashboardCategory, error) {
	var querySQL bytes.Buffer
	var row *sql.Row
	var err error

	category := looker.AnalyticsDashboardCategory{}
	retryCount := 0

	querySQL.Write([]byte("select id, tab_label, premium, admin_only, seller_only from analytics_dashboard_categories where tab_label = $1"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), label)
		err = row.Scan(
			&category.ID,
			&category.Label,
			&category.Premium,
			&category.Admin,
			&category.Seller,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("GetAnalyticsDashboardCategoryByLabel() connection with the database timed out!")
		return looker.AnalyticsDashboardCategory{}, err
	case sql.ErrNoRows:
		core.Error("GetAnalyticsDashboardCategoryByLabel() no rows were returned!")
		return looker.AnalyticsDashboardCategory{}, &DoesNotExistError{resourceType: "analytics dashboard category", resourceRef: label}
	case nil:
		return category, nil
	default:
		core.Error("GetAnalyticsDashboardCategoryByLabel() QueryRow returned an error: %v", err)
		return looker.AnalyticsDashboardCategory{}, err
	}
}

// AddAnalyticsDashboardCategory adds a new dashboard category
func (db *SQL) AddAnalyticsDashboardCategory(ctx context.Context, label string, isAdmin bool, isPremium bool, isSeller bool) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("insert into analytics_dashboard_categories (tab_label, premium, admin_only, seller_only) values ($1, $2, $3, $4)"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		label, isPremium, isAdmin, isSeller,
	)
	if err != nil {
		core.Error("AddAnalyticsDashboardCategory() error adding analytics dashboard category: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("AddAnalyticsDashboardCategory() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("AddAnalyticsDashboardCategory() RowsAffected <> 1")
		return err
	}

	return nil
}

// RemoveAnalyticsDashboardCategory remove a dashboard category by ID
func (db *SQL) RemoveAnalyticsDashboardCategoryByID(ctx context.Context, id int64) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("delete from analytics_dashboard_categories where id = $1"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		id,
	)
	if err != nil {
		core.Error("RemoveAnalyticsDashboardCategoryByID() error removing analytics dashboard category: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveAnalyticsDashboardCategoryByID() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveAnalyticsDashboardCategoryByID() RowsAffected <> 1")
		return err
	}

	return nil
}

// RemoveAnalyticsDashboardCategory remove a dashboard category by label
func (db *SQL) RemoveAnalyticsDashboardCategoryByLabel(ctx context.Context, label string) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("delete from analytics_dashboard_categories where tab_label = $1"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		label,
	)
	if err != nil {
		core.Error("RemoveAnalyticsDashboardCategoryByID() error removing analytics dashboard category: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveAnalyticsDashboardCategoryByID() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveAnalyticsDashboardCategoryByID() RowsAffected <> 1")
		return err
	}

	return nil
}

// UpdateAnalyticsDashboardCategoryByID update dashboard category by id
func (db *SQL) UpdateAnalyticsDashboardCategoryByID(ctx context.Context, id int64, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	switch field {
	case "Label":
		label, ok := value.(string)
		if !ok || label == "" {
			return fmt.Errorf("%v is not a valid string or empty", value)
		}
		updateSQL.Write([]byte("update analytics_dashboard_categories set tab_label=$1 where id=$2"))
		args = append(args, label, id)

	case "Premium":
		premium, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%v is not a valid bool value", value)
		}
		updateSQL.Write([]byte("update analytics_dashboard_categories set premium=$1 where id=$2"))
		args = append(args, premium, id)

	case "Admin":
		admin, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%v is not a valid bool value", value)
		}
		updateSQL.Write([]byte("update analytics_dashboard_categories set admin_only=$1 where id=$2"))
		args = append(args, admin, id)

	case "Seller":
		seller, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%v is not a valid bool value", value)
		}
		updateSQL.Write([]byte("update analytics_dashboard_categories set seller_only=$1 where id=$2"))
		args = append(args, seller, id)

	default:
		return fmt.Errorf("field '%v' does not exist on the analytics dashboard category type", field)

	}

	result, err := ExecRetry(ctx, db, updateSQL, args...)
	if err != nil {
		core.Error("UpdateAnalyticsDashboardCategoryByID() error modifying relay record: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("UpdateAnalyticsDashboardCategoryByID() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("UpdateAnalyticsDashboardCategoryByID() RowsAffected <> 1")
		return err
	}

	return nil
}

// GetAdminAnalyticsDashboards get all admin looker dashboards
func (db *SQL) GetAdminAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error) {
	var sql bytes.Buffer
	var customerID int64
	var categoryID int64
	dashboard := looker.AnalyticsDashboard{}
	dashboards := []looker.AnalyticsDashboard{}

	sql.Write([]byte("select id, dashboard_name, looker_dashboard_id, discovery, customer_id, category_id "))
	sql.Write([]byte("from analytics_dashboards"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("GetAdminAnalyticsDashboards(): QueryMultipleRowsRetry returned an error: %v", err)
		return []looker.AnalyticsDashboard{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&dashboard.ID,
			&dashboard.Name,
			&dashboard.LookerID,
			&dashboard.Discovery,
			&customerID,
			&categoryID,
		)
		if err != nil {
			core.Error("GetAdminAnalyticsDashboards(): error parsing returned row: %v", err)
			return []looker.AnalyticsDashboard{}, err
		}

		customer, err := db.CustomerByID(ctx, customerID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		category, err := db.GetAnalyticsDashboardCategoryByID(ctx, categoryID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		if category.Admin {
			dashboard.CustomerCode = customer.Code
			dashboard.Category = category

			dashboards = append(dashboards, dashboard)
		}
	}

	sort.Slice(dashboards, func(i int, j int) bool { return dashboards[i].ID < dashboards[j].ID })
	return dashboards, nil
}

// GetAnalyticsDashboardsByCategoryID get all looker dashboards by category id
func (db *SQL) GetAnalyticsDashboardsByCategoryID(ctx context.Context, id int64) ([]looker.AnalyticsDashboard, error) {
	var sql bytes.Buffer
	var customerID int64
	var categoryID int64
	dashboard := looker.AnalyticsDashboard{}
	dashboards := []looker.AnalyticsDashboard{}

	sql.Write([]byte("select id, dashboard_name, looker_dashboard_id, discovery, customer_id, category_id "))
	sql.Write([]byte("from analytics_dashboards where category_id = $1"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql, id)
	if err != nil {
		core.Error("GetAnalyticsDashboardsByCategoryID(): QueryMultipleRowsRetry returned an error: %v", err)
		return []looker.AnalyticsDashboard{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&dashboard.ID,
			&dashboard.Name,
			&dashboard.LookerID,
			&dashboard.Discovery,
			&customerID,
			&categoryID,
		)
		if err != nil {
			core.Error("GetAnalyticsDashboardsByCategoryID(): error parsing returned row: %v", err)
			return []looker.AnalyticsDashboard{}, err
		}

		customer, err := db.CustomerByID(ctx, customerID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		category, err := db.GetAnalyticsDashboardCategoryByID(ctx, categoryID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		dashboard.CustomerCode = customer.Code
		dashboard.Category = category

		dashboards = append(dashboards, dashboard)
	}

	sort.Slice(dashboards, func(i int, j int) bool { return dashboards[i].ID < dashboards[j].ID })
	return dashboards, nil
}

// GetAnalyticsDashboardsByCategoryLabel get all looker dashboards by category label
func (db *SQL) GetAnalyticsDashboardsByCategoryLabel(ctx context.Context, label string) ([]looker.AnalyticsDashboard, error) {
	var sql bytes.Buffer
	var customerID int64
	var categoryID int64
	dashboard := looker.AnalyticsDashboard{}
	dashboards := []looker.AnalyticsDashboard{}

	sql.Write([]byte("select id, dashboard_name, looker_dashboard_id, discovery, customer_id, category_id "))
	sql.Write([]byte("from analytics_dashboards"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("GetAnalyticsDashboardsByCategoryLabel(): QueryMultipleRowsRetry returned an error: %v", err)
		return []looker.AnalyticsDashboard{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&dashboard.ID,
			&dashboard.Name,
			&dashboard.LookerID,
			&dashboard.Discovery,
			&customerID,
			&categoryID,
		)
		if err != nil {
			core.Error("GetAnalyticsDashboardsByCategoryLabel(): error parsing returned row: %v", err)
			return []looker.AnalyticsDashboard{}, err
		}

		customer, err := db.CustomerByID(ctx, customerID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		category, err := db.GetAnalyticsDashboardCategoryByID(ctx, categoryID)
		if err != nil {

			return []looker.AnalyticsDashboard{}, err
		}

		dashboard.CustomerCode = customer.Code
		dashboard.Category = category

		if dashboard.Category.Label == label {
			dashboards = append(dashboards, dashboard)
		}
	}

	sort.Slice(dashboards, func(i int, j int) bool { return dashboards[i].ID < dashboards[j].ID })
	return dashboards, nil
}

// GetPremiumAnalyticsDashboards get all premium looker dashboards
func (db *SQL) GetPremiumAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error) {
	var sql bytes.Buffer
	var customerID int64
	var categoryID int64
	dashboard := looker.AnalyticsDashboard{}
	dashboards := []looker.AnalyticsDashboard{}

	sql.Write([]byte("select id, dashboard_name, looker_dashboard_id, discovery, customer_id, category_id "))
	sql.Write([]byte("from analytics_dashboards"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("GetPremiumAnalyticsDashboards(): QueryMultipleRowsRetry returned an error: %v", err)
		return []looker.AnalyticsDashboard{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&dashboard.ID,
			&dashboard.Name,
			&dashboard.LookerID,
			&dashboard.Discovery,
			&customerID,
			&categoryID,
		)
		if err != nil {
			core.Error("GetPremiumAnalyticsDashboards(): error parsing returned row: %v", err)
			return []looker.AnalyticsDashboard{}, err
		}

		customer, err := db.CustomerByID(ctx, customerID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		category, err := db.GetAnalyticsDashboardCategoryByID(ctx, categoryID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		dashboard.CustomerCode = customer.Code
		dashboard.Category = category

		if dashboard.Category.Premium {
			dashboards = append(dashboards, dashboard)
		}
	}

	sort.Slice(dashboards, func(i int, j int) bool { return dashboards[i].ID < dashboards[j].ID })
	return dashboards, nil
}

// GetFreeAnalyticsDashboards get all free looker dashboards
func (db *SQL) GetFreeAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error) {
	var sql bytes.Buffer
	var customerID int64
	var categoryID int64
	dashboard := looker.AnalyticsDashboard{}
	dashboards := []looker.AnalyticsDashboard{}

	sql.Write([]byte("select id, dashboard_name, looker_dashboard_id, discovery, customer_id, category_id "))
	sql.Write([]byte("from analytics_dashboards where discovery = false"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("GetFreeAnalyticsDashboards(): QueryMultipleRowsRetry returned an error: %v", err)
		return []looker.AnalyticsDashboard{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&dashboard.ID,
			&dashboard.Name,
			&dashboard.LookerID,
			&dashboard.Discovery,
			&customerID,
			&categoryID,
		)
		if err != nil {
			core.Error("GetFreeAnalyticsDashboards(): error parsing returned row: %v", err)
			return []looker.AnalyticsDashboard{}, err
		}

		customer, err := db.CustomerByID(ctx, customerID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		category, err := db.GetAnalyticsDashboardCategoryByID(ctx, categoryID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		dashboard.CustomerCode = customer.Code
		dashboard.Category = category

		if !dashboard.Category.Premium {
			dashboards = append(dashboards, dashboard)
		}
	}

	sort.Slice(dashboards, func(i int, j int) bool { return dashboards[i].ID < dashboards[j].ID })
	return dashboards, nil
}

// GetDiscoveryAnalyticsDashboards get all discovery looker dashboards
func (db *SQL) GetDiscoveryAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error) {
	var sql bytes.Buffer
	var customerID int64
	var categoryID int64
	dashboard := looker.AnalyticsDashboard{}
	dashboards := []looker.AnalyticsDashboard{}

	sql.Write([]byte("select id, dashboard_name, looker_dashboard_id, discovery, customer_id, category_id "))
	sql.Write([]byte("from analytics_dashboards where discovery = true"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("GetDiscoveryAnalyticsDashboards(): QueryMultipleRowsRetry returned an error: %v", err)
		return []looker.AnalyticsDashboard{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&dashboard.ID,
			&dashboard.Name,
			&dashboard.LookerID,
			&dashboard.Discovery,
			&customerID,
			&categoryID,
		)
		if err != nil {
			core.Error("GetDiscoveryAnalyticsDashboards(): error parsing returned row: %v", err)
			return []looker.AnalyticsDashboard{}, err
		}

		customer, err := db.CustomerByID(ctx, customerID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		category, err := db.GetAnalyticsDashboardCategoryByID(ctx, categoryID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		dashboard.CustomerCode = customer.Code
		dashboard.Category = category

		dashboards = append(dashboards, dashboard)
	}

	sort.Slice(dashboards, func(i int, j int) bool { return dashboards[i].ID < dashboards[j].ID })
	return dashboards, nil
}

// GetDiscoveryAnalyticsDashboards get all discovery looker dashboards
func (db *SQL) GetAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error) {
	var sql bytes.Buffer
	var customerID int64
	var categoryID int64
	dashboard := looker.AnalyticsDashboard{}
	dashboards := []looker.AnalyticsDashboard{}

	sql.Write([]byte("select id, dashboard_name, looker_dashboard_id, discovery, customer_id, category_id "))
	sql.Write([]byte("from analytics_dashboards"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql)
	if err != nil {
		core.Error("GetAnalyticsDashboards(): QueryMultipleRowsRetry returned an error: %v", err)
		return []looker.AnalyticsDashboard{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&dashboard.ID,
			&dashboard.Name,
			&dashboard.LookerID,
			&dashboard.Discovery,
			&customerID,
			&categoryID,
		)
		if err != nil {
			core.Error("GetAnalyticsDashboards(): error parsing returned row: %v", err)
			return []looker.AnalyticsDashboard{}, err
		}

		customer, err := db.CustomerByID(ctx, customerID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		category, err := db.GetAnalyticsDashboardCategoryByID(ctx, categoryID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		dashboard.CustomerCode = customer.Code
		dashboard.Category = category

		dashboards = append(dashboards, dashboard)
	}

	sort.Slice(dashboards, func(i int, j int) bool { return dashboards[i].ID < dashboards[j].ID })
	return dashboards, nil
}

// GetAnalyticsDashboardsByLookerID get looker dashboards by looker id
func (db *SQL) GetAnalyticsDashboardsByLookerID(ctx context.Context, id string) ([]looker.AnalyticsDashboard, error) {
	var sql bytes.Buffer
	var customerID int64
	var categoryID int64
	dashboard := looker.AnalyticsDashboard{}
	dashboards := []looker.AnalyticsDashboard{}

	sql.Write([]byte("select id, dashboard_name, looker_dashboard_id, discovery, customer_id, category_id "))
	sql.Write([]byte("from analytics_dashboards where looker_dashboard_id = $1"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	rows, err := QueryMultipleRowsRetry(ctx, db, sql, id)
	if err != nil {
		core.Error("GetAnalyticsDashboardsByLookerID(): QueryMultipleRowsRetry returned an error: %v", err)
		return []looker.AnalyticsDashboard{}, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&dashboard.ID,
			&dashboard.Name,
			&dashboard.LookerID,
			&dashboard.Discovery,
			&customerID,
			&categoryID,
		)
		if err != nil {
			core.Error("GetAnalyticsDashboardsByLookerID(): error parsing returned row: %v", err)
			return []looker.AnalyticsDashboard{}, err
		}

		customer, err := db.CustomerByID(ctx, customerID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		category, err := db.GetAnalyticsDashboardCategoryByID(ctx, categoryID)
		if err != nil {
			return []looker.AnalyticsDashboard{}, err
		}

		dashboard.CustomerCode = customer.Code
		dashboard.Category = category

		dashboards = append(dashboards, dashboard)
	}

	sort.Slice(dashboards, func(i int, j int) bool { return dashboards[i].ID < dashboards[j].ID })
	return dashboards, nil
}

// GetAnalyticsDashboardByID get looker dashboard by id
func (db *SQL) GetAnalyticsDashboardByID(ctx context.Context, id int64) (looker.AnalyticsDashboard, error) {
	var querySQL bytes.Buffer
	var customerID int64
	var categoryID int64
	var row *sql.Row
	var err error

	dashboard := looker.AnalyticsDashboard{}
	retryCount := 0

	querySQL.Write([]byte("select id, dashboard_name, looker_dashboard_id, discovery, customer_id, category_id from analytics_dashboards where id = $1"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), id)
		err = row.Scan(
			&dashboard.ID,
			&dashboard.Name,
			&dashboard.LookerID,
			&dashboard.Discovery,
			&customerID,
			&categoryID,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("GetAnalyticsDashboardByID() connection with the database timed out!")
		return looker.AnalyticsDashboard{}, err
	case sql.ErrNoRows:
		core.Error("GetAnalyticsDashboardByID() no rows were returned!")
		return looker.AnalyticsDashboard{}, &DoesNotExistError{resourceType: "analytics dashboard id", resourceRef: id}
	case nil:
		customer, err := db.CustomerByID(ctx, customerID)
		if err != nil {
			return looker.AnalyticsDashboard{}, err
		}

		category, err := db.GetAnalyticsDashboardCategoryByID(ctx, categoryID)
		if err != nil {
			return looker.AnalyticsDashboard{}, err
		}

		dashboard.CustomerCode = customer.Code
		dashboard.Category = category

		return dashboard, nil
	default:
		core.Error("GetAnalyticsDashboardByID() QueryRow returned an error: %v", err)
		return looker.AnalyticsDashboard{}, err
	}
}

// GetAnalyticsDashboardByName get looker dashboard by name
func (db *SQL) GetAnalyticsDashboardByName(ctx context.Context, name string) (looker.AnalyticsDashboard, error) {
	var querySQL bytes.Buffer
	var customerID int64
	var categoryID int64
	var row *sql.Row
	var err error

	dashboard := looker.AnalyticsDashboard{}
	retryCount := 0

	querySQL.Write([]byte("select id, dashboard_name, looker_dashboard_id, discovery, customer_id, category_id from analytics_dashboards where dashboard_name = $1"))

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	for retryCount < MAX_RETRIES {
		row = db.Client.QueryRowContext(ctx, querySQL.String(), name)
		err = row.Scan(
			&dashboard.ID,
			&dashboard.Name,
			&dashboard.LookerID,
			&dashboard.Discovery,
			&customerID,
			&categoryID,
		)
		switch err {
		case context.Canceled:
			retryCount = retryCount + 1
		default:
			retryCount = MAX_RETRIES
		}
	}

	switch err {
	case context.Canceled:
		core.Error("GetAnalyticsDashboardByName() connection with the database timed out!")
		return looker.AnalyticsDashboard{}, err
	case sql.ErrNoRows:
		core.Error("GetAnalyticsDashboardByName() no rows were returned!")
		return looker.AnalyticsDashboard{}, &DoesNotExistError{resourceType: "analytics dashboard name", resourceRef: name}
	case nil:
		customer, err := db.CustomerByID(ctx, customerID)
		if err != nil {
			return looker.AnalyticsDashboard{}, err
		}

		category, err := db.GetAnalyticsDashboardCategoryByID(ctx, categoryID)
		if err != nil {
			return looker.AnalyticsDashboard{}, err
		}

		dashboard.CustomerCode = customer.Code
		dashboard.Category = category

		return dashboard, nil
	default:
		core.Error("GetAnalyticsDashboardByName() QueryRow returned an error: %v", err)
		return looker.AnalyticsDashboard{}, err
	}
}

// AddAnalyticsDashboard adds a new dashboard
func (db *SQL) AddAnalyticsDashboard(ctx context.Context, name string, lookerID int64, isDiscover bool, customerID int64, categoryID int64) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("insert into analytics_dashboards (dashboard_name, looker_dashboard_id, discovery, customer_id, category_id) values ($1, $2, $3, $4, $5)"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		name, lookerID, isDiscover, customerID, categoryID,
	)
	if err != nil {
		core.Error("AddAnalyticsDashboard() error adding analytics dashboard category: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("AddAnalyticsDashboard() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("AddAnalyticsDashboard() RowsAffected <> 1")
		return err
	}

	return nil
}

// RemoveAnalyticsDashboardByID remove looker dashboard by id
func (db *SQL) RemoveAnalyticsDashboardByID(ctx context.Context, id int64) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("delete from analytics_dashboards where id = $1"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		id,
	)
	if err != nil {
		core.Error("RemoveAnalyticsDashboardByID() error removing analytics dashboard: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveAnalyticsDashboardByID() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveAnalyticsDashboardByID() RowsAffected <> 1")
		return err
	}

	return nil
}

// RemoveAnalyticsDashboardByName remove looker dashboard by name
func (db *SQL) RemoveAnalyticsDashboardByName(ctx context.Context, name string) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("delete from analytics_dashboards where dashboard_name = $1"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		name,
	)
	if err != nil {
		core.Error("RemoveAnalyticsDashboardByName() error removing analytics dashboard: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveAnalyticsDashboardByName() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveAnalyticsDashboardByName() RowsAffected <> 1")
		return err
	}

	return nil
}

// RemoveAnalyticsDashboardByLookerID remove looker dashboard by looker id
func (db *SQL) RemoveAnalyticsDashboardByLookerID(ctx context.Context, id string) error {
	var sql bytes.Buffer

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	sql.Write([]byte("delete from analytics_dashboards where looker_dashboard_id = $1"))

	result, err := ExecRetry(
		ctx,
		db,
		sql,
		id,
	)
	if err != nil {
		core.Error("RemoveAnalyticsDashboardByLookerID() error removing analytics dashboard: %v", err)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("RemoveAnalyticsDashboardByLookerID() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("RemoveAnalyticsDashboardByLookerID() RowsAffected <> 1")
		return err
	}

	return nil
}

// UpdateAnalyticsDashboardByID update dashboard by id
func (db *SQL) UpdateAnalyticsDashboardByID(ctx context.Context, id int64, field string, value interface{}) error {
	var updateSQL bytes.Buffer
	var args []interface{}

	ctx, cancel := context.WithTimeout(ctx, SQL_TIMEOUT)
	defer cancel()

	switch field {
	case "Name":
		name, ok := value.(string)
		if !ok || name == "" {
			return fmt.Errorf("%v is not a valid string or is empty", value)
		}
		updateSQL.Write([]byte("update analytics_dashboards set dashboard_name=$1 where id=$2"))
		args = append(args, name, id)

	case "LookerID":
		dashID, ok := value.(int64)
		if !ok {
			return fmt.Errorf("%v is not a valid int64 value", value)
		}
		updateSQL.Write([]byte("update analytics_dashboards set looker_dashboard_id=$1 where id=$2"))
		args = append(args, dashID, id)

	case "Discovery":
		isDiscovery, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%v is not a valid bool value", value)
		}
		updateSQL.Write([]byte("update analytics_dashboards set discovery=$1 where id=$2"))
		args = append(args, isDiscovery, id)

	case "CustomerCode":
		customerCode, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid bool value", value)
		}

		customer, err := db.Customer(ctx, customerCode)
		if err != nil {
			return &DoesNotExistError{resourceType: "customer", resourceRef: customerCode}
		}

		updateSQL.Write([]byte("update analytics_dashboards set customer_id=$1 where id=$2"))
		args = append(args, customer.DatabaseID, id)

	case "Category":
		categoryID, ok := value.(int64)
		if !ok {
			return fmt.Errorf("%v is not a valid int64 value", value)
		}

		updateSQL.Write([]byte("update analytics_dashboards set category_id=$1 where id=$2"))
		args = append(args, categoryID, id)

	default:
		return fmt.Errorf("field '%v' does not exist on the analytics dashboard type", field)
	}

	result, err := ExecRetry(ctx, db, updateSQL, args...)
	if err != nil {
		core.Error("UpdateAnalyticsDashboardByID() error modifying relay record: %v", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		core.Error("UpdateAnalyticsDashboardByID() RowsAffected returned an error")
		return err
	}
	if rows != 1 {
		core.Error("UpdateAnalyticsDashboardByID() RowsAffected <> 1")
		return err
	}

	return nil
}
