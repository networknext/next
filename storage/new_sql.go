package storage

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
)

// NewSQLite3 returns an SQLite3 backed database pointer. This package is
// only used for unit testing.
func NewSQLite3(ctx context.Context, logger log.Logger) (*SQL, error) {

	// remove the old db file if it exists (SQLite3 save one by default when
	// exiting)
	if _, err := os.Stat("network_next.db"); err == nil || os.IsExist(err) {
		err = os.Remove("network_next.db")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error removing old db file: %w", err)
			return nil, err
		}
	}

	sqlite3, err := sql.Open("sqlite3", "network_next.db")
	if err != nil {
		err = fmt.Errorf("NewSQLite3() error creating db connection: %w", err)
		return nil, err
	}

	db := &SQL{
		Client:             sqlite3,
		Logger:             logger,
		datacenters:        make(map[uint64]routing.Datacenter),
		datacenterMaps:     make(map[uint64]routing.DatacenterMap),
		relays:             make(map[uint64]routing.Relay),
		customers:          make(map[string]routing.Customer),
		buyers:             make(map[uint64]routing.Buyer),
		sellers:            make(map[string]routing.Seller),
		SyncSequenceNumber: -1,
	}

	// populate the db with some data from dev
	file, err := ioutil.ReadFile("sqlite3-empty.sql")
	if err != nil {
		err = fmt.Errorf("NewSQLite3() error opening seed file: %w", err)
		return nil, err
	}

	requests := strings.Split(string(file), ";")

	for _, request := range requests {
		_, err := db.Client.Exec(request)
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error executing seed file sql line: %v\n", err)
			return nil, err
		}
	}
	syncIntervalStr := os.Getenv("DB_SYNC_INTERVAL")
	syncInterval, err := time.ParseDuration(syncIntervalStr)
	if err != nil {
		level.Error(logger).Log("envvar", "DB_SYNC_INTERVAL", "value", syncIntervalStr, "err", err)
		os.Exit(1)
	}
	// Start a goroutine to sync from the database
	go func() {
		ticker := time.NewTicker(syncInterval)
		db.SyncLoop(ctx, ticker.C)
	}()

	return db, nil
}

// NewPostgreSQL returns an PostgreSQL backed database pointer
func NewPostgreSQL(ctx context.Context, logger log.Logger) (*SQL, error) {

	// move sensitive stuff to env w/ GCP vars
	const (
		host     = "localhost"
		port     = 5432
		user     = "engineering"
		password = "0xdeadbeef"
		dbname   = "nn"
	)

	pgsqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	pgsql, err := sql.Open("postgres", pgsqlInfo)
	if err != nil {
		err = fmt.Errorf("NewPostgreSQL() error creating db connection: %w", err)
		return nil, err
	}

	// db.Ping actually establishes the connection and validates the parameters
	err = pgsql.Ping()
	if err != nil {
		err = fmt.Errorf("NewPostgreSQL() error pinging db: %w", err)
		return nil, err
	}

	db := &SQL{
		Client:             pgsql,
		Logger:             logger,
		datacenters:        make(map[uint64]routing.Datacenter),
		datacenterMaps:     make(map[uint64]routing.DatacenterMap),
		relays:             make(map[uint64]routing.Relay),
		customers:          make(map[string]routing.Customer),
		buyers:             make(map[uint64]routing.Buyer),
		sellers:            make(map[string]routing.Seller),
		SyncSequenceNumber: -1,
	}

	syncIntervalStr := os.Getenv("DB_SYNC_INTERVAL")
	syncInterval, err := time.ParseDuration(syncIntervalStr)
	if err != nil {
		level.Error(logger).Log("envvar", "DB_SYNC_INTERVAL", "value", syncIntervalStr, "err", err)
		os.Exit(1)
	}
	// Start a goroutine to sync from the database
	go func() {

		ticker := time.NewTicker(syncInterval)
		db.SyncLoop(ctx, ticker.C)
	}()

	return db, nil

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

// Sync is a utility function that calls the individual sync* methods in the proper order.
func (db *SQL) Sync(ctx context.Context) error {

	seqNumberNotInSync, value, err := db.CheckSequenceNumber(ctx)
	if err != nil {
		return err
	}

	if !seqNumberNotInSync {
		return nil
	}

	db.sequenceNumberMutex.Lock()
	db.SyncSequenceNumber = value
	db.sequenceNumberMutex.Unlock()

	var outerErr error
	var wg sync.WaitGroup
	wg.Add(6)

	// Due to foreign key relationships in the tables, they must
	// be synced in this order:
	// 	1 Customers
	//	2 Buyers
	//	3 Sellers
	// 	4 Datacenters
	// 	5 DatacenterMaps
	//	6 Relays

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

	go func() {
		if err := db.syncRelays(ctx); err != nil {
			outerErr = fmt.Errorf("failed to sync relays: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()

	return outerErr
}

func (db *SQL) syncDatacenters(ctx context.Context) error {

	var sql bytes.Buffer
	var dc sqlDatacenter

	datacenters := make(map[uint64]routing.Datacenter)
	datacenterIDs := make(map[int64]uint64)

	sql.Write([]byte("select id, enabled, latitude, longitude,"))
	sql.Write([]byte("supplier_name, street_address, seller_id from datacenters"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return err
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

		datacenterIDs[dc.SellerID] = did

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

	db.datacenterIDsMutex.Lock()
	db.datacenterIDs = datacenterIDs
	db.datacenterIDsMutex.Unlock()

	db.datacenterMutex.Lock()
	db.datacenters = datacenters
	db.datacenterMutex.Unlock()

	level.Info(db.Logger).Log("during", "syncDatacenters", "num", len(db.datacenters))

	return nil
}

func (db *SQL) syncRelays(ctx context.Context) error {

	var sql bytes.Buffer
	var relay sqlRelay

	relays := make(map[uint64]routing.Relay)
	relayIDs := make(map[int64]uint64)

	sql.Write([]byte("select relays.id, relays.contract_term, relays.end_date, relays.included_bandwidth_gb, relays.management_ip, "))
	sql.Write([]byte("relays.max_sessions, relays.mrc, relays.overage, relays.port_speed, relays.public_ip, relays.public_ip_port, relays.public_key, "))
	sql.Write([]byte("relays.ssh_port, relays.ssh_user, relays.start_date, relays.update_key, bw_billing_rules.name, relays.datacenter, "))
	sql.Write([]byte("machine_types.name, relay_states.name from relays "))
	sql.Write([]byte("inner join relay_states on relays.relay_state = relay_states.id "))
	sql.Write([]byte("inner join machine_types on relays.machine_type = machine_types.id "))
	sql.Write([]byte("inner join bw_billing_rules on relays.bw_billing_rule = bw_billing_rules.id "))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&relay.RelayID,
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
			&relay.UpdateKey,
			&relay.BWRule,
			&relay.DatacenterID,
			&relay.MachineType,
			&relay.State,
			&relay.RelayID,
		)

		fullPublicAddress := relay.PublicIP + ":" + fmt.Sprintf("%d", relay.PublicIPPort)
		rid := crypto.HashID(fullPublicAddress)

		publicAddr, err := net.ResolveUDPAddr("udp", fullPublicAddress)
		if err != nil {
			level.Error(db.Logger).Log("during", "net.ResolveUDPAddr returned an error parsing public address", "err", err)
		}

		managementAddr, err := net.ResolveUDPAddr("udp", relay.ManagementIP)
		if err != nil {
			level.Error(db.Logger).Log("during", "net.ResolveUDPAddr returned an error parsing management address", "err", err)
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

		datacenter := db.datacenters[db.datacenterIDs[relay.RelayID]]

		relayIDs[relay.RelayID] = rid

		relays[rid] = routing.Relay{
			ID:                  rid,
			Name:                relay.Name,
			Addr:                *publicAddr,
			PublicKey:           relay.PublicKey,
			Datacenter:          datacenter,
			NICSpeedMbps:        int32(relay.NICSpeedMbps),
			IncludedBandwidthGB: int32(relay.IncludedBandwithGB),
			State:               relayState,
			ManagementAddr:      managementAddr.String(),
			SSHUser:             relay.SSHUser,
			SSHPort:             relay.SSHPort,
			MaxSessions:         uint32(relay.MaxSessions),
			UpdateKey:           relay.UpdateKey,
			MRC:                 routing.Nibblin(relay.MRC),
			Overage:             routing.Nibblin(relay.Overage),
			BWRule:              bwRule,
			ContractTerm:        int32(relay.ContractTerm),
			StartDate:           relay.StartDate,
			EndDate:             relay.EndDate,
			Type:                machineType,
			RelayID:             relay.RelayID,
		}
	}

	db.relayMutex.Lock()
	db.relays = relays
	db.relayMutex.Unlock()

	db.relayIDsMutex.Lock()
	db.relays = relays
	db.relayIDsMutex.Unlock()

	level.Info(db.Logger).Log("during", "syncRelays", "num", len(db.relays))

	return nil
}

type sqlBuyer struct {
	ID             uint64
	IsLiveCustomer bool
	Name           string
	PublicKey      []byte
	CompanyCode    string
	BuyerID        int64 // sql PK
	CustomerID     int64 // sql PK
}

func (db *SQL) syncBuyers(ctx context.Context) error {

	var sql bytes.Buffer
	var buyer sqlBuyer

	buyers := make(map[uint64]routing.Buyer)
	buyerIDs := make(map[int64]uint64)

	sql.Write([]byte("select id, is_live_customer, public_key, customer_id "))
	sql.Write([]byte("from buyers"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&buyer.BuyerID,
			&buyer.IsLiveCustomer,
			&buyer.PublicKey,
			&buyer.CustomerID,
		)
		buyer.ID = binary.LittleEndian.Uint64(buyer.PublicKey[:8])

		buyerIDs[buyer.BuyerID] = buyer.ID

		rs, err := db.GetRouteShaderForBuyerID(ctx, buyer.BuyerID)
		if err != nil {
			level.Warn(db.Logger).Log("msg", fmt.Sprintf("failed to completely read route shader for buyer %v, some fields will have default values", buyer.ID), "err", err)
		}

		ic, err := db.GetInternalConfigForBuyerID(ctx, buyer.BuyerID)
		if err != nil {
			level.Warn(db.Logger).Log("msg", fmt.Sprintf("failed to completely read internal config for buyer %v, some fields will have default values", buyer.ID), "err", err)
		}

		buyers[buyer.ID] = routing.Buyer{
			// CompanyCode:    db.customerIDs[buyer.CustomerID],
			ID:             buyer.ID,
			Live:           buyer.IsLiveCustomer,
			PublicKey:      buyer.PublicKey,
			RouteShader:    rs,
			InternalConfig: ic,
			CustomerID:     buyer.CustomerID,
		}

	}

	db.buyerIDsMutex.Lock()
	db.buyerIDs = buyerIDs
	db.buyerIDsMutex.Unlock()

	db.buyerMutex.Lock()
	db.buyers = buyers
	db.buyerMutex.Unlock()

	level.Info(db.Logger).Log("during", "syncBuyers", "num", len(db.customers))

	return nil
}
func (db *SQL) syncSellers(ctx context.Context) error {

	var sql bytes.Buffer
	var seller sqlSeller

	sellers := make(map[string]routing.Seller)
	sellerIDs := make(map[int64]string)

	sql.Write([]byte("select id, public_egress_price, public_ingress_price, "))
	sql.Write([]byte("customer_id from sellers"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&seller.SellerID,
			&seller.EgressPriceNibblinsPerGB,
			&seller.IngressPriceNibblinsPerGB,
			&seller.CustomerID,
		)

		// seller name is defined by the parent customer
		sellerIDs[seller.SellerID] = db.customerIDs[seller.CustomerID]

		sellers[db.customerIDs[seller.CustomerID]] = routing.Seller{
			ID:                        db.customerIDs[seller.CustomerID],
			IngressPriceNibblinsPerGB: routing.Nibblin(seller.IngressPriceNibblinsPerGB),
			EgressPriceNibblinsPerGB:  routing.Nibblin(seller.EgressPriceNibblinsPerGB),
			SellerID:                  seller.SellerID,
			CustomerID:                seller.CustomerID,
		}

	}

	db.sellerIDsMutex.Lock()
	db.sellerIDs = sellerIDs
	db.sellerIDsMutex.Unlock()

	db.sellerMutex.Lock()
	db.sellers = sellers
	db.sellerMutex.Unlock()

	level.Info(db.Logger).Log("during", "syncCustomers", "num", len(db.customers))

	return nil
}
func (db *SQL) syncDatacenterMaps(ctx context.Context) error {
	return nil
}
func (db *SQL) syncCustomers(ctx context.Context) error {
	var sql bytes.Buffer
	var customer sqlCustomer

	customers := make(map[string]routing.Customer)
	customerIDs := make(map[int64]string)

	sql.Write([]byte("select id, active, automatic_signin_domain, "))
	sql.Write([]byte("customer_name, customer_code from customers"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&customer.ID,
			&customer.Active,
			&customer.AutomaticSignInDomains,
			&customer.CustomerName,
			&customer.CustomerCode,
		)

		customerIDs[customer.ID] = customer.CustomerCode

		customers[customer.CustomerCode] = routing.Customer{
			Code:                   customer.CustomerCode,
			Name:                   customer.Name,
			AutomaticSignInDomains: customer.AutomaticSignInDomains,
			Active:                 customer.Active,
			CustomerID:             customer.ID,
		}
	}

	db.customerIDsMutex.Lock()
	db.customerIDs = customerIDs
	db.customerIDsMutex.Unlock()

	db.customerMutex.Lock()
	db.customers = customers
	db.customerMutex.Unlock()

	level.Info(db.Logger).Log("during", "syncCustomers", "num", len(db.customers))

	return nil
}
