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
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/routing"
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

	datacenterMutex     sync.RWMutex
	relayMutex          sync.RWMutex
	customerMutex       sync.RWMutex
	buyerMutex          sync.RWMutex
	sellerMutex         sync.RWMutex
	datacenterMapMutex  sync.RWMutex
	sequenceNumberMutex sync.RWMutex

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

	// Add the buyer in cached storage
	db.syncCustomers(ctx)

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
func (db *SQL) AddBuyer(ctx context.Context, b routing.Buyer) error {
	var sql bytes.Buffer

	db.buyerMutex.RLock()
	_, ok := db.buyers[b.ID]
	db.buyerMutex.RUnlock()

	if ok {
		return &AlreadyExistsError{resourceType: "buyer", resourceRef: b.ID}
	}

	// skip Name and CompanyCode - they are part of the parent routing.Customer
	buyer := sqlBuyer{
		ID:             b.ID,
		IsLiveCustomer: b.Live,
		PublicKey:      b.PublicKey,
		CustomerID:     b.CustomerID,
	}

	// Add the buyer in remote storage
	sql.Write([]byte("insert into buyers ("))
	sql.Write([]byte("is_live_customer, public_key, customer_id"))
	sql.Write([]byte(") values ($1, $2, $3)"))

	stmt, err := db.Client.PrepareContext(ctx, sql.String())
	if err != nil {
		level.Error(db.Logger).Log("during", "error perparing AddBuyer SQL", "err", err)
		return err
	}

	result, err := stmt.Exec(buyer.IsLiveCustomer,
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

	// Add the buyer in cached storage
	db.syncBuyers(ctx)

	db.IncrementSequenceNumber(ctx)

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
	ID                        string
	IngressPriceNibblinsPerGB int64
	EgressPriceNibblinsPerGB  int64
	CustomerID                int64
	SellerID                  int64
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

	newSellerData := sqlSeller{
		ID:                        s.ID,
		IngressPriceNibblinsPerGB: int64(s.IngressPriceNibblinsPerGB),
		EgressPriceNibblinsPerGB:  int64(s.EgressPriceNibblinsPerGB),
		CustomerID:                s.CustomerID,
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

type sqlRelay struct {
	Name               string
	PublicIP           string // []byte?
	PublicIPPort       int64
	PublicKey          []byte
	UpdateKey          []byte
	NICSpeedMbps       int64
	IncludedBandwithGB int64
	DatacenterID       int64
	ManagementIP       string // []byte?
	SSHUser            string
	SSHPort            int64
	State              string
	MaxSessions        int64
	MRC                int64
	Overage            int64
	BWRule             string
	ContractTerm       int64
	StartDate          time.Time
	EndDate            time.Time
	MachineType        string
	RelayID            int64
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

type sqlRouteShader struct {
	ABTest                    bool
	AcceptableLatency         int64
	AcceptablePacketLoss      float64
	BandwidthEnvelopeDownKbps int64
	BandwidthEnvelopeUpKbps   int64
	DisableNetworkNext        bool
	DisplayName               string
	LatencyThreshold          int64
	Multipath                 bool
	ProMode                   bool
	ReduceLatency             bool
	ReducePacketLoss          bool
	SelectionPercent          int64
}

// GetRouteShaderForBuyerID TODO: will need to either return a slice of routeshaders or accept another
// arg - buyerID returns (can return) multiple records.
func (db *SQL) GetRouteShaderForBuyerID(ctx context.Context, buyerID int64) (core.RouteShader, error) {
	var sql bytes.Buffer
	var coreRS core.RouteShader

	sql.Write([]byte("select ab_test, acceptable_latency, acceptable_packet_loss, bw_envelope_down_kbps, "))
	sql.Write([]byte("bw_envelope_up_kbps, disable_network_next, latency_threshold, multipath, pro_mode, "))
	sql.Write([]byte("reduce_latency, reduce_packet_loss, selection_percent from route_shaders "))
	sql.Write([]byte("where id=$1"))

	rows, err := db.Client.QueryContext(ctx, sql.String(), buyerID)
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return core.RouteShader{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var rs sqlRouteShader

		err := rows.Scan(
			&rs.ABTest,
			&rs.AcceptableLatency,
			&rs.AcceptablePacketLoss,
			&rs.BandwidthEnvelopeDownKbps,
			&rs.BandwidthEnvelopeUpKbps,
			&rs.DisableNetworkNext,
			&rs.LatencyThreshold,
			&rs.Multipath,
			&rs.ProMode,
			&rs.ReduceLatency,
			&rs.ReducePacketLoss,
			&rs.SelectionPercent,
		)
		if err != nil {
			level.Error(db.Logger).Log("during", "rows.Scan returned an error", "err", err)
			return core.RouteShader{}, err
		}

		coreRS = core.RouteShader{
			ABTest:                    rs.ABTest,
			AcceptableLatency:         int32(rs.AcceptableLatency),
			AcceptablePacketLoss:      float32(rs.AcceptablePacketLoss),
			BandwidthEnvelopeDownKbps: int32(rs.BandwidthEnvelopeDownKbps),
			BandwidthEnvelopeUpKbps:   int32(rs.BandwidthEnvelopeUpKbps),
			DisableNetworkNext:        rs.DisableNetworkNext,
			LatencyThreshold:          int32(rs.LatencyThreshold),
			Multipath:                 rs.Multipath,
			ProMode:                   rs.ProMode,
			ReduceLatency:             rs.ReduceLatency,
			ReducePacketLoss:          rs.ReducePacketLoss,
			SelectionPercent:          int(rs.SelectionPercent),
		}
		coreRS.BannedUsers = make(map[uint64]bool) // not implemented yet

		return coreRS, nil

	}

	return coreRS, &DoesNotExistError{resourceType: "RouteShader", resourceRef: fmt.Sprintf("%x", buyerID)}
}

type sqlInternalConfig struct {
	MaxLatencyTradeOff         int64
	MultipathOverloadThreshold int64
	RouteSwitchThreshold       int64
	RttVetoDefault             int64
	RttVetoMultipath           int64
	RttVetoPacketLoss          int64
	TryBeforeYouBuy            bool
	ForceNext                  bool
	LargeCustomer              bool
}

// GetInternalConfigForBuyerID TODO: will need to either return a slice of routeshaders or accept another
// arg - buyerID returns (can return) multiple records.
func (db *SQL) GetInternalConfigForBuyerID(ctx context.Context, buyerID int64) (core.InternalConfig, error) {
	var sql bytes.Buffer
	var coreIC core.InternalConfig

	sql.Write([]byte("select max_latency_tradeoff, multipath_overload_threshold, "))
	sql.Write([]byte("route_switch_threshold, rtt_veto_default, rtt_veto_multipath, "))
	sql.Write([]byte("rtt_veto_packetloss, try_before_you_buy, force_next, large_customer "))
	sql.Write([]byte(" from rs_internal_configs where buyer_id=$1"))

	rows, err := db.Client.QueryContext(ctx, sql.String(), buyerID)
	if err != nil {
		level.Error(db.Logger).Log("during", "QueryContext returned an error", "err", err)
		return core.InternalConfig{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var ic sqlInternalConfig

		err := rows.Scan(
			&ic.MaxLatencyTradeOff,
			&ic.MultipathOverloadThreshold,
			&ic.RouteSwitchThreshold,
			&ic.RttVetoDefault,
			&ic.RttVetoMultipath,
			&ic.RttVetoPacketLoss,
			&ic.TryBeforeYouBuy,
			&ic.ForceNext,
			&ic.LargeCustomer,
		)

		if err != nil {
			level.Error(db.Logger).Log("during", "rows.Scan returned an error", "err", err)
			return core.InternalConfig{}, err
		}

		coreIC = core.InternalConfig{
			RouteSwitchThreshold:       int32(ic.RouteSwitchThreshold),
			MaxLatencyTradeOff:         int32(ic.MaxLatencyTradeOff),
			RTTVeto_Default:            int32(ic.RttVetoDefault),
			RTTVeto_PacketLoss:         int32(ic.RttVetoPacketLoss),
			RTTVeto_Multipath:          int32(ic.RttVetoMultipath),
			MultipathOverloadThreshold: int32(ic.MultipathOverloadThreshold),
			TryBeforeYouBuy:            ic.TryBeforeYouBuy,
			ForceNext:                  ic.ForceNext,
			LargeCustomer:              ic.LargeCustomer,
		}

		return coreIC, nil

	}

	return coreIC, &DoesNotExistError{resourceType: "InternalConfig", resourceRef: fmt.Sprintf("%x", buyerID)}
}

// utility functions for unit testing - methods not defined by the interface
// are not available in testing
func GetCustomerID(ctx context.Context, db *SQL, companyCode string) (int64, error) {

	customer, ok := db.customers[companyCode]
	if !ok {
		return -1, &DoesNotExistError{resourceType: "Customer", resourceRef: fmt.Sprintf("%s", companyCode)}
	}

	sqlID := customer.CustomerID
	return sqlID, nil
}
