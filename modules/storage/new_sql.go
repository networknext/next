package storage

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
)

// NewSQLite3 returns an SQLite3 backed database pointer. This package is
// only used for unit testing.
func NewSQLite3(ctx context.Context, logger log.Logger) (*SQL, error) {

	var sqlite3 *sql.DB

	// pwd, _ := os.Getwd()
	// fmt.Printf("NewSQLite3() pwd: %s\n", pwd)

	// remove the old db file if it exists (SQLite3 saves one by default when exiting)
	// fmt.Println("--> Attempting to remove db file")
	if _, err := os.Stat("testdata/sqlite3-empty.sql"); err == nil || os.IsExist(err) { // happy path
		// fmt.Println("--> Removing testdata/network_next.db")
		err = os.Remove("testdata/network_next.db")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error removing old db file: %w", err)
			// return nil, err
		}
		// fmt.Println("--> Removed testdata/network_next.db")
		sqlite3, err = sql.Open("sqlite3", "file:testdata/network_next.db?_foreign_keys=on&_locking_mode=NORMAL")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error creating db connection: %w", err)
			return nil, err
		}
		// fmt.Println("--> opened testdata/network_next.db")
	} else if _, err := os.Stat("../../testdata/sqlite3-empty.sql"); err == nil || os.IsExist(err) { // unit test
		// fmt.Println("--> Removing ../../testdata/network_next.db")
		err = os.Remove("../../testdata/network_next.db")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error removing old db file: %w", err)
			// return nil, err
		}
		// fmt.Println("--> Removed ../../testdata/network_next.db")
		sqlite3, err = sql.Open("sqlite3", "file:../../testdata/network_next.db?_foreign_keys=on&_locking_mode=NORMAL")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error creating db connection: %w", err)
			return nil, err
		}
		// fmt.Println("--> opened ../../testdata/network_next.db")
	} else {
		fmt.Println("--> did not find db file?")
		os.Exit(0)
	}

	// db.Ping actually establishes the connection and validates the parameters
	err := sqlite3.Ping()
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
		routeShaders:       make(map[uint64]core.RouteShader),
		internalConfigs:    make(map[uint64]core.InternalConfig),
		bannedUsers:        make(map[uint64]map[uint64]bool),
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

func NewSQLite3Staging(ctx context.Context, logger log.Logger) (*SQL, error) {
	var sqlite3 *sql.DB

	if _, err := os.Stat("/app/sqlite3-empty.sql"); err == nil || os.IsExist(err) {

		err = os.Remove("/app/network_next.db")
		if err != nil {
			err = fmt.Errorf("NewSQLite3() error removing old db file: %v", err)
		}

		// Boiler plate SQL file exists, load it in
		sqlite3, err = sql.Open("sqlite3", "file:/app/network_next.db?_foreign_keys=on&_locking_mode=NORMAL")
		if err != nil {
			return nil, fmt.Errorf("NewSQLite3Staging() error creating db connection: %v", err)
		}
	} else {
		return nil, fmt.Errorf("NewSQLite3Staging() could not find /app/sqlite3-empty.sql")
	}

	// db.Ping actually establishes the connection and validates the parameters
	err := sqlite3.Ping()
	if err != nil {
		err = fmt.Errorf("NewSQLite3Staging() error pinging db: %v", err)
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
		routeShaders:       make(map[uint64]core.RouteShader),
		internalConfigs:    make(map[uint64]core.InternalConfig),
		bannedUsers:        make(map[uint64]map[uint64]bool),
		SyncSequenceNumber: -1,
	}

	// populate the db with basic tables
	file, err := ioutil.ReadFile("/app/sqlite3-empty.sql") // happy path
	if err != nil {
		return nil, fmt.Errorf("NewSQLite3Staging() error reading from ")
	}

	requests := strings.Split(string(file), ";")

	for _, request := range requests {
		_, err := db.Client.Exec(request)
		if err != nil {
			err = fmt.Errorf("NewSQLite3Staging() error executing seed file sql line: %v", err)
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
func NewPostgreSQL(
	ctx context.Context,
	logger log.Logger,
	pgHostIP string,
	pgUserName string,
	pgPassword string,
) (*SQL, error) {

	fmt.Println("Creating PostgreSQL Storer.")

	// -- port and db name are the same regardless of the environment
	// -- sslmode is a driver req, connection is internal/IP and encrypted
	pgsqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		pgHostIP, 5432, pgUserName, pgPassword, "network_next")

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
		buyerIDs:           make(map[uint64]int64),
		sellerIDs:          make(map[int64]string),
		routeShaders:       make(map[uint64]core.RouteShader),
		internalConfigs:    make(map[uint64]core.InternalConfig),
		bannedUsers:        make(map[uint64]map[uint64]bool),
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

	// seqNumberNotInSync, value, err := db.CheckSequenceNumber(ctx)
	// if err != nil {
	// 	return err
	// }

	// if !seqNumberNotInSync {
	// 	return nil
	// }

	// db.sequenceNumberMutex.Lock()
	// db.SyncSequenceNumber = value
	// db.sequenceNumberMutex.Unlock()

	// Due to foreign key relationships in the tables, they must
	// be synced in this order:
	// 	1 Customers
	//	2 Buyers
	//  3 InternalConfigs
	//  4 BannedUsers
	//  5 RouteShaders
	//	6 Sellers
	// 	7 Datacenters
	// 	8 DatacenterMaps
	//	9 Relays

	if err := db.syncCustomers(ctx); err != nil {
		return fmt.Errorf("failed to sync customers: %v", err)
	}

	if err := db.syncBuyers(ctx); err != nil {
		return fmt.Errorf("failed to sync buyers: %v", err)
	}

	if err := db.syncInternalConfigs(ctx); err != nil {
		return fmt.Errorf("failed to sync internal configs: %v", err)
	}

	if err := db.syncBannedUsers(ctx); err != nil {
		return fmt.Errorf("failed to sync banned users: %v", err)
	}

	if err := db.syncRouteShaders(ctx); err != nil {
		return fmt.Errorf("failed to sync route shaders: %v", err)
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

	sql.Write([]byte("select id, display_name, latitude, longitude,"))
	sql.Write([]byte("seller_id from datacenters"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "syncDatacenters(): QueryContext returned an error", "err", err)
		return err
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
			return err
		}

		did := crypto.HashID(dc.Name)

		datacenterIDs[dc.ID] = did

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

		datacenters[did] = d

	}

	db.datacenterIDsMutex.Lock()
	db.datacenterIDs = datacenterIDs
	db.datacenterIDsMutex.Unlock()

	db.datacenterMutex.Lock()
	db.datacenters = datacenters
	db.datacenterMutex.Unlock()

	return nil
}

func (db *SQL) syncRelays(ctx context.Context) error {

	var sqlQuery bytes.Buffer
	var relay sqlRelay

	relays := make(map[uint64]routing.Relay)
	relayIDs := make(map[int64]uint64)

	sqlQuery.Write([]byte("select relays.id, relays.hex_id, relays.display_name, relays.contract_term, relays.end_date, "))
	sqlQuery.Write([]byte("relays.included_bandwidth_gb, relays.management_ip, "))
	sqlQuery.Write([]byte("relays.max_sessions, relays.mrc, relays.overage, relays.port_speed, "))
	sqlQuery.Write([]byte("relays.public_ip, relays.public_ip_port, relays.public_key, "))
	sqlQuery.Write([]byte("relays.ssh_port, relays.ssh_user, relays.start_date, relays.internal_ip, "))
	sqlQuery.Write([]byte("relays.internal_ip_port, relays.bw_billing_rule, relays.datacenter, "))
	sqlQuery.Write([]byte("relays.machine_type, relays.relay_state, "))
	sqlQuery.Write([]byte("relays.internal_ip, relays.internal_ip_port, relays.notes , "))
	sqlQuery.Write([]byte("relays.billing_supplier, relays.relay_version from relays "))
	// sql.Write([]byte("inner join relay_states on relays.relay_state = relay_states.id "))
	// sql.Write([]byte("inner join machine_types on relays.machine_type = machine_types.id "))
	// sql.Write([]byte("inner join bw_billing_rules on relays.bw_billing_rule = bw_billing_rules.id "))

	rows, err := db.Client.QueryContext(ctx, sqlQuery.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "syncRelays(): QueryContext returned an error", "err", err)
		return err
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
			return err
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

		datacenter, err := db.Datacenter(db.datacenterIDs[relay.DatacenterID])
		if err != nil {
			level.Error(db.Logger).Log("during", "syncRelays error dereferencing datacenter", "err", err)
		}

		seller, err := db.Seller(db.sellerIDs[datacenter.SellerID])
		if err != nil {
			level.Error(db.Logger).Log("during", "syncRelays error dereferencing seller", "err", err)
		}

		internalID, err := strconv.ParseUint(relay.HexID, 16, 64)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncRelays error parsing hex_id", "err", err)
		}

		relayIDs[relay.DatabaseID] = internalID

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

		relays[internalID] = r

	}

	db.relayMutex.Lock()
	db.relays = relays
	db.relayMutex.Unlock()

	db.relayIDsMutex.Lock()
	db.relayIDs = relayIDs
	db.relayIDsMutex.Unlock()

	return nil
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

func (db *SQL) syncBuyers(ctx context.Context) error {

	var sql bytes.Buffer
	var buyer sqlBuyer

	buyers := make(map[uint64]routing.Buyer)
	buyerIDs := make(map[uint64]int64)

	sql.Write([]byte("select sdk_generated_id, id, short_name, is_live_customer, debug, public_key, customer_id "))
	sql.Write([]byte("from buyers"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "syncBuyers(): QueryContext returned an error", "err", err)
		return err
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
			level.Error(db.Logger).Log("during", "syncBuyers(): error parsing returned row", "err", err)
			return err
		}

		buyer.ID = uint64(buyer.SdkID)

		buyerIDs[buyer.ID] = buyer.DatabaseID

		// apply default values - custom values will be attached in
		// syncInternalConfigs() and syncRouteShaders() if they exist
		rs := core.NewRouteShader()
		ic := core.NewInternalConfig()

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

		buyers[uint64(buyer.DatabaseID)] = b

	}

	db.buyerIDsMutex.Lock()
	db.buyerIDs = buyerIDs
	db.buyerIDsMutex.Unlock()

	db.buyerMutex.Lock()
	db.buyers = buyers
	db.buyerMutex.Unlock()

	return nil
}
func (db *SQL) syncSellers(ctx context.Context) error {

	var sql bytes.Buffer
	var seller sqlSeller

	sellers := make(map[string]routing.Seller)
	sellerIDs := make(map[int64]string)

	sql.Write([]byte("select id, short_name, public_egress_price, secret, "))
	sql.Write([]byte("customer_id from sellers"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "syncSellers(): QueryContext returned an error", "err", err)
		return err
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
			level.Error(db.Logger).Log("during", "syncSellers(): error parsing returned row", "err", err)
			return err
		}

		// seller name is defined by the parent customer
		sellerIDs[seller.DatabaseID] = db.customerIDs[seller.CustomerID]

		sellers[db.customerIDs[seller.CustomerID]] = routing.Seller{
			ID:                       db.customerIDs[seller.CustomerID],
			ShortName:                seller.ShortName,
			Secret:                   seller.Secret,
			CompanyCode:              db.customers[db.customerIDs[seller.CustomerID]].Code,
			Name:                     db.customers[db.customerIDs[seller.CustomerID]].Name,
			EgressPriceNibblinsPerGB: routing.Nibblin(seller.EgressPriceNibblinsPerGB),
			DatabaseID:               seller.DatabaseID,
			CustomerID:               seller.CustomerID,
		}

	}

	db.sellerIDsMutex.Lock()
	db.sellerIDs = sellerIDs
	db.sellerIDsMutex.Unlock()

	db.sellerMutex.Lock()
	db.sellers = sellers
	db.sellerMutex.Unlock()

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

	sql.Write([]byte("select buyer_id, datacenter_id from datacenter_maps"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "syncDatacenterMaps(): QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&sqlMap.BuyerID, &sqlMap.DatacenterID)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncDatacenterMaps(): error parsing returned row", "err", err)
			return err
		}

		db.buyerMutex.RLock()
		buyer := db.buyers[uint64(sqlMap.BuyerID)]
		db.buyerMutex.RUnlock()

		ephemeralBuyerID := buyer.ID

		dcMap := routing.DatacenterMap{
			BuyerID:      ephemeralBuyerID,
			DatacenterID: db.datacenterIDs[sqlMap.DatacenterID],
		}

		id := crypto.HashID(fmt.Sprintf("%016x", dcMap.BuyerID) + fmt.Sprintf("%016x", dcMap.DatacenterID))
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

	sql.Write([]byte("select id, automatic_signin_domain, "))
	sql.Write([]byte("customer_name, customer_code from customers"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "syncCustomers(): QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&customer.ID,
			&customer.AutomaticSignInDomains,
			&customer.Name,
			&customer.CustomerCode,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncCustomers(): error parsing returned row", "err", err)
			return err
		}

		customerIDs[customer.ID] = customer.CustomerCode

		c := routing.Customer{
			Code:                   customer.CustomerCode,
			Name:                   customer.Name,
			AutomaticSignInDomains: customer.AutomaticSignInDomains,
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

	return nil
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

func (db *SQL) syncInternalConfigs(ctx context.Context) error {

	var sql bytes.Buffer
	var sqlIC sqlInternalConfig

	internalConfigs := make(map[uint64]core.InternalConfig)

	sql.Write([]byte("select max_latency_tradeoff, max_rtt, multipath_overload_threshold, "))
	sql.Write([]byte("route_switch_threshold, route_select_threshold, rtt_veto_default, "))
	sql.Write([]byte("rtt_veto_multipath, rtt_veto_packetloss, try_before_you_buy, force_next, "))
	sql.Write([]byte("large_customer, is_uncommitted, high_frequency_pings, route_diversity, "))
	sql.Write([]byte("multipath_threshold, enable_vanity_metrics, reduce_pl_min_slice_number, buyer_id from rs_internal_configs"))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "syncInternalConfigs(): QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var buyerID int64
		err := rows.Scan(
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
			&buyerID,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncInternalConfigs(): error parsing returned row", "err", err)
			return err
		}

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

		db.buyerMutex.RLock()
		buyer := db.buyers[uint64(buyerID)]
		db.buyerMutex.RUnlock()

		buyer.InternalConfig = internalConfig

		db.buyerMutex.Lock()
		db.buyers[uint64(buyerID)] = buyer
		db.buyerMutex.Unlock()

		internalConfigs[uint64(buyerID)] = internalConfig
	}

	db.internalConfigMutex.Lock()
	db.internalConfigs = internalConfigs
	db.internalConfigMutex.Unlock()
	return nil
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

func (db *SQL) syncRouteShaders(ctx context.Context) error {

	var sql bytes.Buffer
	var sqlRS sqlRouteShader

	routeShaders := make(map[uint64]core.RouteShader)

	sql.Write([]byte("select ab_test, acceptable_latency, acceptable_packet_loss, bw_envelope_down_kbps, "))
	sql.Write([]byte("bw_envelope_up_kbps, disable_network_next, latency_threshold, multipath, pro_mode, "))
	sql.Write([]byte("reduce_latency, reduce_packet_loss, reduce_jitter, selection_percent, packet_loss_sustained, buyer_id from route_shaders "))

	rows, err := db.Client.QueryContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "syncRouteShaders(): QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var buyerID int64
		err := rows.Scan(
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
			&buyerID,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncRouteShaders(): error parsing returned row", "err", err)
			return err
		}

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

		db.buyerMutex.RLock()
		buyer := db.buyers[uint64(buyerID)]
		db.buyerMutex.RUnlock()

		buyer.RouteShader = routeShader

		db.buyerMutex.Lock()
		db.buyers[uint64(buyerID)] = buyer
		db.buyerMutex.Unlock()

		if bannedUsers, ok := db.bannedUsers[uint64(buyerID)]; ok {
			routeShader.BannedUsers = bannedUsers
		}
		routeShaders[uint64(buyerID)] = routeShader
	}

	for buyerID, rs := range routeShaders {
		bannedUserList, ok := db.bannedUsers[buyerID]
		if ok {
			rs.BannedUsers = bannedUserList
		}
	}

	db.routeShaderMutex.Lock()
	db.routeShaders = routeShaders
	db.routeShaderMutex.Unlock()
	return nil
}

func (db *SQL) syncBannedUsers(ctx context.Context) error {

	var sql bytes.Buffer
	bannedUserList := make(map[uint64]map[uint64]bool)

	sql.Write([]byte("select user_id, buyer_id from banned_users order by buyer_id asc"))

	rows, err := db.Client.QueryContext(context.Background(), sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "syncBannedUsers(): QueryContext returned an error", "err", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var userID, dbBuyerID int64
		err := rows.Scan(&userID, &dbBuyerID)
		if err != nil {
			level.Error(db.Logger).Log("during", "syncBannedUsers() error parsing user and buyer IDs", "err", err)
			return err
		}

		buyerID := db.buyerIDs[uint64(dbBuyerID)]

		bannedUser := make(map[uint64]bool)
		bannedUser[uint64(userID)] = true
		if _, ok := bannedUserList[uint64(buyerID)]; !ok {
			bannedUserList[uint64(buyerID)] = make(map[uint64]bool)
			bannedUserList[uint64(buyerID)] = bannedUser
		} else {
			bannedUserList[uint64(buyerID)][uint64(userID)] = true
		}
	}

	db.bannedUserMutex.Lock()
	db.bannedUsers = bannedUserList
	db.bannedUserMutex.Unlock()

	return nil

}
