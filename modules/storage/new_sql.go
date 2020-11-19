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
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
)

// NewSQLite3 returns an SQLite3 backed database pointer. This package is
// only used for unit testing.
func NewSQLite3(ctx context.Context, logger log.Logger) (*SQL, error) {

	fmt.Println("Creating SQLite3 Storer.")
	pwd, _ := os.Getwd()
	fmt.Printf("NewSQLite3() pwd: %s\n", pwd)

	// remove the old db file if it exists (SQLite3 save one by default when
	// exiting)
	fmt.Println("--> Attempting to remove db file")
	if _, err := os.Stat("../../testdata/network_next.db"); err == nil || os.IsExist(err) {
		err = os.Remove("../../testdata/network_next.db")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error removing old db file: %w", err)
			return nil, err
		}
	}

	sqlite3, err := sql.Open("sqlite3", "file:../../testdata/network_next.db?_foreign_keys=on&_locking_mode=NORMAL")
	if err != nil {
		err = fmt.Errorf("NewSQLite3() error creating db connection: %w", err)
		return nil, err
	}

	// db.Ping actually establishes the connection and validates the parameters
	err = sqlite3.Ping()
	if err != nil {
		err = fmt.Errorf("NewSQLite3() error pinging db: %w", err)
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
	file, err := ioutil.ReadFile("testdata/sqlite3-empty.sql") // happy path
	if err != nil {
		file, err = ioutil.ReadFile("../../testdata/sqlite3-empty.sql") // unit test
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error opening seed file: %w", err)
			return nil, err
		}
	}

	requests := strings.Split(string(file), ";")

	for _, request := range requests {
		_, err := db.Client.Exec(request)
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error executing seed file sql line: %v", err)
			return nil, err
		}
	}

	syncIntervalStr := os.Getenv("GOOGLE_CLOUD_SQL_SYNC_INTERVAL")
	syncInterval, err := time.ParseDuration(syncIntervalStr)
	if err != nil {
		level.Error(logger).Log("envvar", "GOOGLE_CLOUD_SQL_SYNC_INTERVAL", "value", syncIntervalStr, "err", err)
		return nil, err
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

	fmt.Println("Creating PostgreSQL Storer.")

	// TODO: move sensitive stuff to env w/ GCP vars
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
		datacenterIDs:      make(map[int64]uint64),
		relayIDs:           make(map[int64]uint64),
		customerIDs:        make(map[int64]string),
		buyerIDs:           make(map[int64]uint64),
		sellerIDs:          make(map[int64]string),
		SyncSequenceNumber: -1,
	}

	syncIntervalStr := os.Getenv("GOOGLE_CLOUD_SQL_SYNC_INTERVAL")
	syncInterval, err := time.ParseDuration(syncIntervalStr)
	if err != nil {
		level.Error(logger).Log("envvar", "GOOGLE_CLOUD_SQL_SYNC_INTERVAL", "value", syncIntervalStr, "err", err)
		return nil, err
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

	// Due to foreign key relationships in the tables, they must
	// be synced in this order:
	// 	1 Customers
	//	2 Buyers
	//	3 Sellers
	// 	4 Datacenters
	// 	5 DatacenterMaps
	//	6 Relays

	if err := db.syncCustomers(ctx); err != nil {
		return fmt.Errorf("failed to sync customers: %v", err)
	}

	if err := db.syncBuyers(ctx); err != nil {
		return fmt.Errorf("failed to sync buyers: %v", err)
	}

	if err := db.syncSellers(ctx); err != nil {
		return fmt.Errorf("failed to sync sellers: %v", err)
	}

	if err := db.syncDatacenters(ctx); err != nil {
		return fmt.Errorf("failed to sync datacenters: %v", err)
	}

	if err := db.syncDatacenterMaps(ctx); err != nil {
		return fmt.Errorf("failed to sync datacenterMaps: %v", err)
	}

	if err := db.syncRelays(ctx); err != nil {
		return fmt.Errorf("failed to sync relays: %v", err)
	}

	return nil
}

func (db *SQL) syncDatacenters(ctx context.Context) error {

	var sql bytes.Buffer
	var dc sqlDatacenter

	datacenters := make(map[uint64]routing.Datacenter)
	datacenterIDs := make(map[int64]uint64)

	sql.Write([]byte("select id, display_name, enabled, latitude, longitude,"))
	sql.Write([]byte("supplier_name, street_address, seller_id from datacenters"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&dc.ID,
			&dc.Name,
			&dc.Enabled,
			&dc.Latitude,
			&dc.Longitude,
			&dc.SupplierName,
			&dc.StreetAddress,
			&dc.SellerID,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "error parsing returned row", "err", err)
			return err
		}

		did := crypto.HashID(dc.Name)

		datacenterIDs[dc.ID] = did

		d := routing.Datacenter{
			ID:      did,
			Name:    dc.Name,
			Enabled: dc.Enabled,
			Location: routing.Location{
				Latitude:  dc.Latitude,
				Longitude: dc.Longitude,
			},
			SupplierName:  dc.SupplierName,
			StreetAddress: dc.StreetAddress,
			SellerID:      dc.SellerID,
			DatabaseID:    dc.ID,
		}

		datacenters[did] = d

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

	sql.Write([]byte("select relays.id, relays.display_name, relays.contract_term, relays.end_date, "))
	sql.Write([]byte("relays.included_bandwidth_gb, relays.management_ip, "))
	sql.Write([]byte("relays.max_sessions, relays.mrc, relays.overage, relays.port_speed, "))
	sql.Write([]byte("relays.public_ip, relays.public_ip_port, relays.public_key, "))
	sql.Write([]byte("relays.ssh_port, relays.ssh_user, relays.start_date, relays.update_key, "))
	sql.Write([]byte("relays.bw_billing_rule, relays.datacenter, "))
	sql.Write([]byte("relays.machine_type, relays.relay_state from relays "))
	// sql.Write([]byte("inner join relay_states on relays.relay_state = relay_states.id "))
	// sql.Write([]byte("inner join machine_types on relays.machine_type = machine_types.id "))
	// sql.Write([]byte("inner join bw_billing_rules on relays.bw_billing_rule = bw_billing_rules.id "))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&relay.DatabaseID,
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
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "error parsing returned row", "err", err)
			return err
		}

		fullPublicAddress := relay.PublicIP + ":" + fmt.Sprintf("%d", relay.PublicIPPort)
		rid := crypto.HashID(fullPublicAddress)

		publicAddr, err := net.ResolveUDPAddr("udp", fullPublicAddress)
		if err != nil {
			level.Error(db.Logger).Log("during", "net.ResolveUDPAddr returned an error parsing public address", "err", err)
		}

		// TODO: this should be treated as a legit address
		// managementAddr, err := net.ResolveUDPAddr("udp", relay.ManagementIP)
		// if err != nil {
		// 	fmt.Printf("error parsing mgmt ip: %v\n", err)
		// 	level.Error(db.Logger).Log("during", "net.ResolveUDPAddr returned an error parsing management address", "err", err)
		// }

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

		datacenter := db.datacenters[db.datacenterIDs[relay.DatacenterID]]

		relayIDs[relay.DatabaseID] = rid

		r := routing.Relay{
			ID:                  rid,
			Name:                relay.Name,
			Addr:                *publicAddr,
			PublicKey:           relay.PublicKey,
			Datacenter:          datacenter,
			NICSpeedMbps:        int32(relay.NICSpeedMbps),
			IncludedBandwidthGB: int32(relay.IncludedBandwithGB),
			State:               relayState,
			ManagementAddr:      relay.ManagementIP,
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
			DatabaseID:          relay.DatabaseID,
		}
		relays[rid] = r

	}

	db.relayMutex.Lock()
	db.relays = relays
	db.relayMutex.Unlock()

	db.relayIDsMutex.Lock()
	db.relayIDs = relayIDs
	db.relayIDsMutex.Unlock()

	level.Info(db.Logger).Log("during", "syncRelays", "num", len(db.relays))

	return nil
}

type sqlBuyer struct {
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

func (db *SQL) syncBuyers(ctx context.Context) error {

	var sql bytes.Buffer
	var buyer sqlBuyer

	buyers := make(map[uint64]routing.Buyer)
	buyerIDs := make(map[int64]uint64)

	sql.Write([]byte("select id, short_name, is_live_customer, debug, public_key, customer_id "))
	sql.Write([]byte("from buyers"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&buyer.DatabaseID,
			&buyer.ShortName,
			&buyer.IsLiveCustomer,
			&buyer.Debug,
			&buyer.PublicKey,
			&buyer.CustomerID,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "error parsing returned row", "err", err)
			return err
		}

		buyer.ID = binary.LittleEndian.Uint64(buyer.PublicKey[:8])

		buyerIDs[buyer.DatabaseID] = buyer.ID

		rs, err := db.GetRouteShaderForBuyerID(ctx, buyer.DatabaseID)
		if err != nil {
			// level.Warn(db.Logger).Log("msg", fmt.Sprintf("failed to completely read route shader for buyer %v, some fields will have default values", buyer.ID), "err", err)
		}

		ic, err := db.GetInternalConfigForBuyerID(ctx, buyer.DatabaseID)
		if err != nil {
			level.Warn(db.Logger).Log("msg", fmt.Sprintf("failed to completely read internal config for buyer %v, some fields will have default values", buyer.ID), "err", err)
		}

		b := routing.Buyer{
			ID:             buyer.ID,
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

		buyers[buyer.ID] = b

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

	sql.Write([]byte("select id, short_name, public_egress_price, public_ingress_price, "))
	sql.Write([]byte("customer_id from sellers"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&seller.DatabaseID,
			&seller.ShortName,
			&seller.EgressPriceNibblinsPerGB,
			&seller.IngressPriceNibblinsPerGB,
			&seller.CustomerID,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "error parsing returned row", "err", err)
			return err
		}

		// seller name is defined by the parent customer
		sellerIDs[seller.DatabaseID] = db.customerIDs[seller.CustomerID]

		sellers[db.customerIDs[seller.CustomerID]] = routing.Seller{
			ID:                        db.customerIDs[seller.CustomerID],
			ShortName:                 seller.ShortName,
			CompanyCode:               db.customers[db.customerIDs[seller.CustomerID]].Code,
			Name:                      db.customers[db.customerIDs[seller.CustomerID]].Name,
			IngressPriceNibblinsPerGB: routing.Nibblin(seller.IngressPriceNibblinsPerGB),
			EgressPriceNibblinsPerGB:  routing.Nibblin(seller.EgressPriceNibblinsPerGB),
			DatabaseID:                seller.DatabaseID,
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

type sqlDatacenterMap struct {
	Alias        string
	BuyerID      int64
	DatacenterID int64
}

func (db *SQL) syncDatacenterMaps(ctx context.Context) error {

	var sql bytes.Buffer
	var sqlMap sqlDatacenterMap

	dcMaps := make(map[uint64]routing.DatacenterMap)

	sql.Write([]byte("select alias, buyer_id, datacenter_id from datacenter_maps"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&sqlMap.Alias, &sqlMap.BuyerID, &sqlMap.DatacenterID)
		if err != nil {
			level.Error(db.Logger).Log("during", "error parsing returned row", "err", err)
			return err
		}

		dcMap := routing.DatacenterMap{
			Alias:        sqlMap.Alias,
			BuyerID:      db.buyerIDs[sqlMap.BuyerID],
			DatacenterID: db.datacenterIDs[sqlMap.DatacenterID],
		}

		id := crypto.HashID(dcMap.Alias + fmt.Sprintf("%x", dcMap.BuyerID) + fmt.Sprintf("%x", dcMap.DatacenterID))
		dcMaps[id] = dcMap
	}

	db.datacenterMapMutex.Lock()
	db.datacenterMaps = dcMaps
	db.datacenterMapMutex.Unlock()
	return nil
}
func (db *SQL) syncCustomers(ctx context.Context) error {
	var sql bytes.Buffer
	var customer sqlCustomer

	customers := make(map[string]routing.Customer)
	customerIDs := make(map[int64]string)

	// sql.Write([]byte("select customers.id, customers.active, customers.debug, "))
	// sql.Write([]byte("customers.automatic_signin_domain, customers.customer_name, "))
	// sql.Write([]byte("customers.customer_code, buyers.id as buyer_id, "))
	// sql.Write([]byte("sellers.id as seller_id from customers "))
	// sql.Write([]byte("left join buyers on customers.id = buyers.customer_id "))
	// sql.Write([]byte("left join sellers on customers.id = sellers.customer_id"))

	sql.Write([]byte("select id, active, debug, automatic_signin_domain, "))
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
			&customer.Debug,
			&customer.AutomaticSignInDomains,
			&customer.Name,
			&customer.CustomerCode,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "error parsing returned row", "err", err)
			fmt.Printf("error parsing returned row: %v\n", err)
			return err
		}

		customerIDs[customer.ID] = customer.CustomerCode

		c := routing.Customer{
			Code:                   customer.CustomerCode,
			Name:                   customer.Name,
			AutomaticSignInDomains: customer.AutomaticSignInDomains,
			Active:                 customer.Active,
			DatabaseID:             customer.ID,
		}

		customers[customer.CustomerCode] = c
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
