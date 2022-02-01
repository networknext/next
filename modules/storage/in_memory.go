package storage

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport/looker"
)

type InMemory struct {
	localCustomers      []routing.Customer
	localBuyers         []routing.Buyer
	localSellers        []routing.Seller
	localRelays         []routing.Relay
	localDatacenters    []routing.Datacenter
	localDatacenterMaps []routing.DatacenterMap

	LocalMode bool
}

func (m *InMemory) DatabaseBinFileReference(ctx context.Context) (routing.DatabaseBinWrapperReference, error) {
	return routing.DatabaseBinWrapperReference{}, fmt.Errorf("Need to implement DatabaseBinFileReference for in memory storer")
}

func (m *InMemory) Buyer(ctx context.Context, id uint64) (routing.Buyer, error) {
	for _, buyer := range m.localBuyers {
		if buyer.ID == id {
			return buyer, nil
		}
	}

	return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: id}
}

func (m *InMemory) Buyers(ctx context.Context) []routing.Buyer {
	buyers := make([]routing.Buyer, len(m.localBuyers))
	for i := range buyers {
		buyers[i] = m.localBuyers[i]
	}

	return buyers
}

func (m *InMemory) BuyerWithCompanyCode(ctx context.Context, code string) (routing.Buyer, error) {
	for _, buyer := range m.localBuyers {
		if buyer.CompanyCode == code {
			return buyer, nil
		}
	}

	return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: code}
}

func (m *InMemory) AddBuyer(ctx context.Context, buyer routing.Buyer) error {
	for _, b := range m.localBuyers {
		if b.ID == buyer.ID {
			return &AlreadyExistsError{resourceType: "buyer", resourceRef: buyer.ID}
		}
	}

	m.localBuyers = append(m.localBuyers, buyer)
	return nil
}

func (m *InMemory) RemoveBuyer(ctx context.Context, id uint64) error {
	buyerIndex := -1
	for i, buyer := range m.localBuyers {
		if buyer.ID == id {
			buyerIndex = i
		}
	}

	if buyerIndex < 0 {
		return &DoesNotExistError{resourceType: "buyer", resourceRef: id}
	}

	if buyerIndex+1 == len(m.localBuyers) {
		m.localBuyers = m.localBuyers[:buyerIndex]
		return nil
	}

	frontSlice := m.localBuyers[:buyerIndex]
	backSlice := m.localBuyers[buyerIndex+1:]
	m.localBuyers = append(frontSlice, backSlice...)
	return nil
}

func (m *InMemory) SetBuyer(ctx context.Context, buyer routing.Buyer) error {
	for i := range m.localBuyers {
		if m.localBuyers[i].ID == buyer.ID {
			m.localBuyers[i] = buyer
			return nil
		}
	}

	return &DoesNotExistError{resourceType: "buyer", resourceRef: buyer.ID}
}

func (m *InMemory) Seller(ctx context.Context, id string) (routing.Seller, error) {
	for _, seller := range m.localSellers {
		if seller.ID == id {
			return seller, nil
		}
	}

	return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: id}
}

func (m *InMemory) Sellers(ctx context.Context) []routing.Seller {
	sellers := make([]routing.Seller, len(m.localSellers))
	for i := range sellers {
		sellers[i] = m.localSellers[i]
	}

	return sellers
}

func (m *InMemory) SellerWithCompanyCode(ctx context.Context, code string) (routing.Seller, error) {
	for _, seller := range m.localSellers {
		if seller.CompanyCode == code {
			return seller, nil
		}
	}

	return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: code}
}

func (m *InMemory) AddSeller(ctx context.Context, seller routing.Seller) error {
	for _, b := range m.localSellers {
		if b.ID == seller.ID {
			return &AlreadyExistsError{resourceType: "seller", resourceRef: seller.ID}
		}
	}

	m.localSellers = append(m.localSellers, seller)
	return nil
}

func (m *InMemory) RemoveSeller(ctx context.Context, id string) error {
	sellerIndex := -1
	for i, seller := range m.localSellers {
		if seller.ID == id {
			sellerIndex = i
		}
	}

	if sellerIndex < 0 {
		return &DoesNotExistError{resourceType: "seller", resourceRef: id}
	}

	if sellerIndex+1 == len(m.localSellers) {
		m.localSellers = m.localSellers[:sellerIndex]
		return nil
	}

	frontSlice := m.localSellers[:sellerIndex]
	backSlice := m.localSellers[sellerIndex+1:]
	m.localSellers = append(frontSlice, backSlice...)
	return nil
}

func (m *InMemory) SetSeller(ctx context.Context, seller routing.Seller) error {
	for i := range m.localSellers {
		if m.localSellers[i].ID == seller.ID {
			m.localSellers[i] = seller
			return nil
		}
	}

	return &DoesNotExistError{resourceType: "seller", resourceRef: seller.ID}
}

func (m *InMemory) Customer(ctx context.Context, code string) (routing.Customer, error) {
	for _, customer := range m.localCustomers {
		if customer.Code == code {
			return customer, nil
		}
	}

	return routing.Customer{}, &DoesNotExistError{resourceType: "customer", resourceRef: code}
}

func (m *InMemory) CustomerByID(ctx context.Context, id int64) (routing.Customer, error) {
	for _, customer := range m.localCustomers {
		if customer.DatabaseID == id {
			return customer, nil
		}
	}
	return routing.Customer{}, &DoesNotExistError{resourceType: "customer", resourceRef: id}
}

func (m *InMemory) Customers(ctx context.Context) []routing.Customer {
	customers := make([]routing.Customer, len(m.localCustomers))
	for i := range customers {
		customers[i] = m.localCustomers[i]
	}

	return customers
}

func (m *InMemory) CustomerWithName(ctx context.Context, name string) (routing.Customer, error) {
	for _, customer := range m.localCustomers {
		if customer.Name == name {
			return customer, nil
		}
	}

	return routing.Customer{}, &DoesNotExistError{resourceType: "customer", resourceRef: name}
}

func (m *InMemory) AddCustomer(ctx context.Context, customer routing.Customer) error {
	for _, c := range m.localCustomers {
		if c.Code == customer.Code {
			return &AlreadyExistsError{resourceType: "customer", resourceRef: customer.Code}
		}
	}

	m.localCustomers = append(m.localCustomers, customer)
	return nil
}

func (m *InMemory) RemoveCustomer(ctx context.Context, code string) error {
	customerIndex := -1
	for i, customer := range m.localCustomers {
		if customer.Code == code {
			customerIndex = i
		}
	}

	if customerIndex < 0 {
		return &DoesNotExistError{resourceType: "customer", resourceRef: code}
	}

	if customerIndex+1 == len(m.localCustomers) {
		m.localCustomers = m.localCustomers[:customerIndex]
		return nil
	}

	frontSlice := m.localCustomers[:customerIndex]
	backSlice := m.localCustomers[customerIndex+1:]
	m.localCustomers = append(frontSlice, backSlice...)
	return nil
}

func (m *InMemory) SetCustomer(ctx context.Context, customer routing.Customer) error {
	for i := range m.localCustomers {
		if m.localCustomers[i].Code == customer.Code {
			m.localCustomers[i] = customer
			return nil
		}
	}

	return &DoesNotExistError{resourceType: "customer", resourceRef: customer.Code}
}

// SetCustomerLink is a no-op since InMemory has no concept on customers
func (m *InMemory) SetCustomerLink(ctx context.Context, customerName string, buyerID uint64, sellerID string) error {
	return nil
}

// BuyerIDFromCustomerName is a no-op since InMemory has no concept on customers
func (m *InMemory) BuyerIDFromCustomerName(ctx context.Context, customerName string) (uint64, error) {
	return 0, nil
}

// SellerIDFromCustomerName is a no-op since InMemory has no concept on customers
func (m *InMemory) SellerIDFromCustomerName(ctx context.Context, customerName string) (string, error) {
	return "", nil
}

func (m *InMemory) Relay(ctx context.Context, id uint64) (routing.Relay, error) {
	for _, relay := range m.localRelays {
		if relay.ID == id {
			return relay, nil
		}
	}

	// If the relay isn't found then just return the first one, since we need one for local dev
	if m.LocalMode && len(m.localRelays) > 0 {
		return m.localRelays[0], nil
	}

	return routing.Relay{}, &DoesNotExistError{resourceType: "relay", resourceRef: id}
}

func (m *InMemory) Relays(ctx context.Context) []routing.Relay {
	relays := make([]routing.Relay, len(m.localRelays))
	for i := range relays {
		relays[i] = m.localRelays[i]
	}

	return relays
}

func (m *InMemory) AddRelay(ctx context.Context, relay routing.Relay) error {
	for _, r := range m.localRelays {
		if r.ID == relay.ID {
			return &AlreadyExistsError{resourceType: "relay", resourceRef: relay.ID}
		}
	}

	// Emulate postgres behavior by requiring the seller and datacenter to exist before adding the relay
	foundSeller := false
	for _, s := range m.localSellers {
		if s.ID == relay.Seller.ID || s.ID == relay.BillingSupplier {
			foundSeller = true
		}
	}

	if !foundSeller {
		return &DoesNotExistError{resourceType: "seller", resourceRef: relay.Seller.ID}
	}

	foundDatacenter := false
	for _, d := range m.localDatacenters {
		if d.ID == relay.Datacenter.ID {
			foundDatacenter = true
		}
	}

	if !foundDatacenter {
		return &DoesNotExistError{resourceType: "datacenter", resourceRef: relay.Datacenter.ID}
	}

	if relay.InternalAddressClientRoutable && relay.InternalAddr.String() == ":0" {
		return &DoesNotExistError{resourceType: "internalAddr", resourceRef: relay.InternalAddr.String()}
	}

	m.localRelays = append(m.localRelays, relay)
	return nil
}

func (m *InMemory) RemoveRelay(ctx context.Context, id uint64) error {
	relayIndex := -1
	for i, relay := range m.localRelays {
		if relay.ID == id {
			relayIndex = i
		}
	}

	if relayIndex < 0 {
		return &DoesNotExistError{resourceType: "relay", resourceRef: id}
	}

	if relayIndex+1 == len(m.localRelays) {
		m.localRelays = m.localRelays[:relayIndex]
		return nil
	}

	frontSlice := m.localRelays[:relayIndex]
	backSlice := m.localRelays[relayIndex+1:]
	m.localRelays = append(frontSlice, backSlice...)
	return nil
}

func (m *InMemory) SetRelay(ctx context.Context, relay routing.Relay) error {
	for i := 0; i < len(m.localRelays); i++ {
		if m.localRelays[i].ID == relay.ID {
			m.localRelays[i] = relay
			return nil
		}
	}

	// If the relay isn't found then just set the first one, since we need to set one for local dev
	if m.LocalMode && len(m.localRelays) > 0 {
		m.localRelays[0] = relay
		return nil
	}

	return &DoesNotExistError{resourceType: "relay", resourceRef: relay.ID}
}

func (m *InMemory) SetRelayMetadata(ctx context.Context, relay routing.Relay) error {
	for i := 0; i < len(m.localRelays); i++ {
		if m.localRelays[i].ID == relay.ID {
			m.localRelays[i] = relay
			return nil
		}
	}

	// If the relay isn't found then just set the first one, since we need to set one for local dev
	if m.LocalMode && len(m.localRelays) > 0 {
		m.localRelays[0] = relay
		return nil
	}

	return &DoesNotExistError{resourceType: "relay", resourceRef: relay.ID}
}

func (m *InMemory) AddDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {

	for _, dc := range m.localDatacenterMaps {
		if dc.BuyerID == dcMap.BuyerID && dc.DatacenterID == dcMap.DatacenterID {
			return &AlreadyExistsError{resourceType: "datacenterMap", resourceRef: dcMap.DatacenterID}
		}
	}

	m.localDatacenterMaps = append(m.localDatacenterMaps, dcMap)

	return nil

}

func (m *InMemory) GetDatacenterMapsForBuyer(ctx context.Context, id uint64) map[uint64]routing.DatacenterMap {
	var dcs = make(map[uint64]routing.DatacenterMap)
	for _, dc := range m.localDatacenterMaps {
		if dc.BuyerID == id {
			id := crypto.HashID(fmt.Sprintf("%x", dc.BuyerID) + fmt.Sprintf("%x", dc.DatacenterID))
			dcs[id] = dc
		}
	}

	return dcs
}

func (m *InMemory) ListDatacenterMaps(ctx context.Context, dcID uint64) map[uint64]routing.DatacenterMap {
	var dcs = make(map[uint64]routing.DatacenterMap)
	for _, dc := range m.localDatacenterMaps {
		if dc.DatacenterID == dcID || dcID == 0 {
			id := crypto.HashID(fmt.Sprintf("%x", dc.BuyerID) + fmt.Sprintf("%x", dc.DatacenterID))
			dcs[id] = dc
		}
	}

	return dcs
}

func (m *InMemory) RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {

	idx := -1
	for i, dcm := range m.localDatacenterMaps {
		if dcMap.BuyerID == dcm.BuyerID && dcMap.DatacenterID == dcm.DatacenterID {
			idx = i
		}
	}

	if idx < 0 {
		return &DoesNotExistError{resourceType: "datacenterMap", resourceRef: dcMap.DatacenterID}
	}

	if idx+1 == len(m.localDatacenterMaps) {
		m.localDatacenterMaps = m.localDatacenterMaps[:idx]
		return nil
	}

	m.localDatacenterMaps = append(m.localDatacenterMaps[:idx], m.localDatacenterMaps[idx+1:]...)
	return nil

}

func (m *InMemory) Datacenter(ctx context.Context, id uint64) (routing.Datacenter, error) {
	for _, datacenter := range m.localDatacenters {
		if datacenter.ID == id {
			return datacenter, nil
		}
	}

	return routing.Datacenter{}, &DoesNotExistError{resourceType: "datacenter", resourceRef: id}
}

func (m *InMemory) Datacenters(ctx context.Context) []routing.Datacenter {
	datacenters := make([]routing.Datacenter, len(m.localDatacenters))
	for i := range datacenters {
		datacenters[i] = m.localDatacenters[i]
	}

	return datacenters
}

func (m *InMemory) AddDatacenter(ctx context.Context, datacenter routing.Datacenter) error {
	for _, d := range m.localDatacenters {
		if d.ID == datacenter.ID {
			return &AlreadyExistsError{resourceType: "datacenter", resourceRef: datacenter.ID}
		}
	}

	m.localDatacenters = append(m.localDatacenters, datacenter)
	return nil
}

func (m *InMemory) RemoveDatacenter(ctx context.Context, id uint64) error {
	datacenterIndex := -1
	for i, datacenter := range m.localDatacenters {
		if datacenter.ID == id {
			datacenterIndex = i
		}
	}

	if datacenterIndex < 0 {
		return &DoesNotExistError{resourceType: "datacenter", resourceRef: id}
	}

	if datacenterIndex+1 == len(m.localDatacenters) {
		m.localDatacenters = m.localDatacenters[:datacenterIndex]
		return nil
	}

	frontSlice := m.localDatacenters[:datacenterIndex]
	backSlice := m.localDatacenters[datacenterIndex+1:]
	m.localDatacenters = append(frontSlice, backSlice...)
	return nil
}

func (m *InMemory) SetDatacenter(ctx context.Context, datacenter routing.Datacenter) error {
	for i := 0; i < len(m.localDatacenters); i++ {
		if m.localDatacenters[i].ID == datacenter.ID {
			m.localDatacenters[i] = datacenter
			return nil
		}
	}

	return &DoesNotExistError{resourceType: "datacenter", resourceRef: datacenter.ID}
}

func (m *InMemory) CheckSequenceNumber(ctx context.Context) (bool, int64, error) {

	err := fmt.Errorf("SetSequenceNumber not implemented in InMemory")
	return false, -1, err
}

func (m *InMemory) IncrementSequenceNumber(ctx context.Context) error {
	return fmt.Errorf("IncrementSequenceNumber not implemented in InMemory")
}

func (m *InMemory) SyncLoop(ctx context.Context, c <-chan time.Time) {
	// no-op - fulfilling Storer interface
}

func (m *InMemory) GetFeatureFlags(ctx context.Context) map[string]bool {
	return map[string]bool{}
}

func (m *InMemory) GetFeatureFlagByName(ctx context.Context, flagName string) (map[string]bool, error) {
	return map[string]bool{}, fmt.Errorf(("GetFeatureFlagByName not impemented in InMemory"))
}

func (m *InMemory) SetFeatureFlagByName(ctx context.Context, flagName string, flagVal bool) error {
	return fmt.Errorf(("SetFeatureFlagByName not impemented in InMemory"))
}

func (m *InMemory) RemoveFeatureFlagByName(ctx context.Context, flagName string) error {
	return fmt.Errorf(("RemoveFeatureFlagByName not impemented in InMemory"))
}

func (m *InMemory) SetSequenceNumber(ctx context.Context, value int64) error {
	return nil
}

func (m *InMemory) InternalConfig(ctx context.Context, buyerID uint64) (core.InternalConfig, error) {
	buyer, err := m.Buyer(ctx, buyerID)
	if err != nil {
		return core.InternalConfig{}, err
	}

	emptyInternalConfig := core.InternalConfig{}
	if buyer.InternalConfig == emptyInternalConfig {
		return core.InternalConfig{}, &DoesNotExistError{resourceType: "InternalConfig", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	return buyer.InternalConfig, nil
}

func (m *InMemory) RouteShader(ctx context.Context, buyerID uint64) (core.RouteShader, error) {
	buyer, err := m.Buyer(ctx, buyerID)
	if err != nil {
		return core.RouteShader{}, err
	}

	rs := buyer.RouteShader

	isEmptyRouteShader := (!rs.DisableNetworkNext &&
		rs.SelectionPercent == 0 &&
		!rs.ABTest &&
		!rs.ReduceLatency &&
		!rs.ReduceJitter &&
		!rs.ReducePacketLoss &&
		!rs.Multipath &&
		!rs.ProMode &&
		rs.AcceptableLatency == 0 &&
		rs.LatencyThreshold == 0 &&
		rs.AcceptablePacketLoss == 0 &&
		rs.BandwidthEnvelopeUpKbps == 0 &&
		rs.BandwidthEnvelopeDownKbps == 0 &&
		len(rs.BannedUsers) == 0 &&
		rs.PacketLossSustained == 0)

	if isEmptyRouteShader {
		return core.RouteShader{}, &DoesNotExistError{resourceType: "RouteShader", resourceRef: fmt.Sprintf("%016x", buyerID)}
	}

	return buyer.RouteShader, nil
}

func (m *InMemory) AddInternalConfig(ctx context.Context, internalConfig core.InternalConfig, buyerID uint64) error {
	for idx, buyer := range m.localBuyers {
		if buyer.ID == buyerID {
			buyer.InternalConfig = internalConfig
			m.localBuyers[idx] = buyer

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "buyer", resourceRef: buyerID}
}

func (m *InMemory) UpdateInternalConfig(ctx context.Context, buyerID uint64, field string, value interface{}) error {
	var buyerExists bool
	var buyer routing.Buyer
	var idx int

	for i, localBuyer := range m.localBuyers {
		if localBuyer.ID == buyerID {

			buyer = localBuyer
			idx = i
			buyerExists = true
			break
		}
	}

	if !buyerExists {
		return &DoesNotExistError{resourceType: "buyer", resourceRef: buyerID}
	}

	switch field {
	case "RouteSelectThreshold":
		routeSelectThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteSelectThreshold: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.InternalConfig.RouteSelectThreshold = routeSelectThreshold
	case "RouteSwitchThreshold":
		routeSwitchThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteSwitchThreshold: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.InternalConfig.RouteSwitchThreshold = routeSwitchThreshold
	case "MaxLatencyTradeOff":
		maxLatencyTradeOff, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MaxLatencyTradeOff: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.InternalConfig.MaxLatencyTradeOff = maxLatencyTradeOff
	case "RTTVeto_Default":
		rttVetoDefault, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_Default: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.InternalConfig.RTTVeto_Default = rttVetoDefault
	case "RTTVeto_PacketLoss":
		rttVetoPacketLoss, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_PacketLoss: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.InternalConfig.RTTVeto_PacketLoss = rttVetoPacketLoss
	case "RTTVeto_Multipath":
		rttVetoMultipath, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RTTVeto_Multipath: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.InternalConfig.RTTVeto_Multipath = rttVetoMultipath
	case "MultipathOverloadThreshold":
		multipathOverloadThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MultipathOverloadThreshold: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.InternalConfig.MultipathOverloadThreshold = multipathOverloadThreshold
	case "TryBeforeYouBuy":
		tryBeforeYouBuy, ok := value.(bool)
		if !ok {
			return fmt.Errorf("TryBeforeYouBuy: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.InternalConfig.TryBeforeYouBuy = tryBeforeYouBuy
	case "ForceNext":
		forceNext, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ForceNext: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.InternalConfig.ForceNext = forceNext
	case "LargeCustomer":
		largeCustomer, ok := value.(bool)
		if !ok {
			return fmt.Errorf("LargeCustomer: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.InternalConfig.LargeCustomer = largeCustomer
	case "Uncommitted":
		uncommitted, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Uncommitted: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.InternalConfig.Uncommitted = uncommitted
	case "HighFrequencyPings":
		highFrequencyPings, ok := value.(bool)
		if !ok {
			return fmt.Errorf("HighFrequencyPings: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.InternalConfig.HighFrequencyPings = highFrequencyPings
	case "MaxRTT":
		maxRTT, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MaxRTT: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.InternalConfig.MaxRTT = maxRTT
	case "RouteDiversity":
		routeDiversity, ok := value.(int32)
		if !ok {
			return fmt.Errorf("RouteDiversity: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.InternalConfig.RouteDiversity = routeDiversity
	case "MultipathThreshold":
		multipathThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("MultipathThreshold: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.InternalConfig.MultipathThreshold = multipathThreshold
	case "EnableVanityMetrics":
		enableVanityMetrics, ok := value.(bool)
		if !ok {
			return fmt.Errorf("EnableVanityMetrics: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.InternalConfig.EnableVanityMetrics = enableVanityMetrics
	case "ReducePacketLossMinSliceNumber":
		reducePacketLossMinSliceNumber, ok := value.(int32)
		if !ok {
			return fmt.Errorf("ReducePacketLossMinSliceNumber: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.InternalConfig.ReducePacketLossMinSliceNumber = reducePacketLossMinSliceNumber
	default:
		return fmt.Errorf("Field '%v' does not exist on the InternalConfig type", field)
	}

	m.localBuyers[idx] = buyer

	return nil
}

func (m *InMemory) RemoveInternalConfig(ctx context.Context, buyerID uint64) error {
	for idx, buyer := range m.localBuyers {
		if buyer.ID == buyerID {
			buyer.InternalConfig = core.InternalConfig{}
			m.localBuyers[idx] = buyer

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "buyer", resourceRef: buyerID}
}

func (m *InMemory) AddRouteShader(ctx context.Context, routeShader core.RouteShader, buyerID uint64) error {
	for idx, buyer := range m.localBuyers {
		if buyer.ID == buyerID {
			buyer.RouteShader = routeShader
			m.localBuyers[idx] = buyer

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "buyer", resourceRef: buyerID}
}

func (m *InMemory) UpdateRouteShader(ctx context.Context, buyerID uint64, field string, value interface{}) error {
	var buyerExists bool
	var buyer routing.Buyer
	var idx int

	for i, localBuyer := range m.localBuyers {
		if localBuyer.ID == buyerID {

			buyer = localBuyer
			idx = i
			buyerExists = true
			break
		}
	}

	if !buyerExists {
		return &DoesNotExistError{resourceType: "buyer", resourceRef: buyerID}
	}

	switch field {
	case "ABTest":
		abTest, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ABTest: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.RouteShader.ABTest = abTest
	case "AcceptableLatency":
		acceptableLatency, ok := value.(int32)
		if !ok {
			return fmt.Errorf("AcceptableLatency: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.RouteShader.AcceptableLatency = acceptableLatency
	case "AcceptablePacketLoss":
		acceptablePacketLoss, ok := value.(float32)
		if !ok {
			return fmt.Errorf("AcceptablePacketLoss: %v is not a valid float32 type (%T)", value, value)
		}

		buyer.RouteShader.AcceptablePacketLoss = acceptablePacketLoss
	case "AnalysisOnly":
		analysisOnly, ok := value.(bool)
		if !ok {
			return fmt.Errorf("AnalysisOnly: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.RouteShader.AnalysisOnly = analysisOnly
	case "BandwidthEnvelopeDownKbps":
		bandwidthEnvelopeDownKbps, ok := value.(int32)
		if !ok {
			return fmt.Errorf("BandwidthEnvelopeDownKbps: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.RouteShader.BandwidthEnvelopeDownKbps = bandwidthEnvelopeDownKbps
	case "BandwidthEnvelopeUpKbps":
		bandwidthEnvelopeUpKbps, ok := value.(int32)
		if !ok {
			return fmt.Errorf("BandwidthEnvelopeUpKbps: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.RouteShader.BandwidthEnvelopeUpKbps = bandwidthEnvelopeUpKbps
	case "DisableNetworkNext":
		disableNetworkNext, ok := value.(bool)
		if !ok {
			return fmt.Errorf("DisableNetworkNext: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.RouteShader.DisableNetworkNext = disableNetworkNext
	case "LatencyThreshold":
		latencyThreshold, ok := value.(int32)
		if !ok {
			return fmt.Errorf("LatencyThreshold: %v is not a valid int32 type (%T)", value, value)
		}

		buyer.RouteShader.LatencyThreshold = latencyThreshold
	case "Multipath":
		multipath, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Multipath: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.RouteShader.Multipath = multipath
	case "ProMode":
		proMode, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ProMode: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.RouteShader.ProMode = proMode
	case "ReduceLatency":
		reduceLatency, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReduceLatency: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.RouteShader.ReduceLatency = reduceLatency
	case "ReducePacketLoss":
		reducePacketLoss, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReducePacketLoss: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.RouteShader.ReducePacketLoss = reducePacketLoss
	case "ReduceJitter":
		reduceJitter, ok := value.(bool)
		if !ok {
			return fmt.Errorf("ReduceJitter: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.RouteShader.ReduceJitter = reduceJitter
	case "SelectionPercent":
		selectionPercent, ok := value.(int)
		if !ok {
			return fmt.Errorf("SelectionPercent: %v is not a valid int type (%T)", value, value)
		}

		buyer.RouteShader.SelectionPercent = selectionPercent
	case "PacketLossSustained":
		packetLossSustained, ok := value.(float32)
		if !ok {
			return fmt.Errorf("PacketLossSustained: %v is not a valid float32 type (%T)", value, value)
		}

		buyer.RouteShader.PacketLossSustained = packetLossSustained
	default:
		return fmt.Errorf("Field '%v' does not exist on the RouteShader type", field)

	}

	m.localBuyers[idx] = buyer

	return nil
}

func (m *InMemory) RemoveRouteShader(ctx context.Context, buyerID uint64) error {
	for idx, buyer := range m.localBuyers {
		if buyer.ID == buyerID {
			buyer.RouteShader = core.RouteShader{}
			m.localBuyers[idx] = buyer

			return nil
		}
	}

	return &DoesNotExistError{resourceType: "buyer", resourceRef: buyerID}
}

func (m *InMemory) UpdateRelay(ctx context.Context, relayID uint64, field string, value interface{}) error {
	var relayExists bool
	var relay routing.Relay
	var idx int

	for i, localRelay := range m.localRelays {
		if localRelay.ID == relayID {

			relay = localRelay
			idx = i
			relayExists = true
			break
		}
	}

	if !relayExists {
		return &DoesNotExistError{resourceType: "relay", resourceRef: relayID}
	}

	switch field {
	case "Name":
		name, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		relay.Name = name
	case "Addr":
		addrString, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		// "removing" a relay zeroes-out the public IP address
		if addrString == "" {
			relay.Addr = net.UDPAddr{}
		} else {
			udpAddr, err := net.ResolveUDPAddr("udp", addrString)
			if err != nil {
				return fmt.Errorf("unable to parse address %s as UDP address: %w", addrString, err)
			}
			relay.Addr = *udpAddr
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
			relay.InternalAddr = net.UDPAddr{}

		} else {
			udpAddr, err := net.ResolveUDPAddr("udp", addrString)
			if err != nil {
				return fmt.Errorf("unable to parse address %s as UDP address: %w", addrString, err)
			}
			relay.InternalAddr = *udpAddr
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

		relay.PublicKey = newPublicKey

	case "NICSpeedMbps":
		portSpeed, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		relay.NICSpeedMbps = int32(portSpeed)

	case "IncludedBandwidthGB":
		includedBW, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		relay.IncludedBandwidthGB = int32(includedBW)

	case "MaxBandwidthMbps":
		maxBandwidthMbps, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		relay.MaxBandwidthMbps = int32(maxBandwidthMbps)

	case "State":
		state, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		if state < 0 || state > 5 {
			return fmt.Errorf("%d is not a valid RelayState value", int64(state))
		}
		relay.State = routing.RelayState(state)

	case "ManagementAddr":
		// routing.Relay.ManagementIP is currently a string type although
		// the database field is inet
		managementIP, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		relay.ManagementAddr = managementIP

	case "SSHUser":
		user, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}
		relay.SSHUser = user

	case "SSHPort":
		port, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		relay.SSHPort = int64(port)

	case "MaxSessions":
		maxSessions, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		if maxSessions < 0 {
			return fmt.Errorf("%d is not a valid MaxSessions value", int32(maxSessions))
		}

		relay.MaxSessions = uint32(maxSessions)

	case "EgressPriceOverride":
		egressPriceOverrideUSD, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		relay.EgressPriceOverride = routing.DollarsToNibblins(egressPriceOverrideUSD)

	case "MRC":
		mrcUSD, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		relay.MRC = routing.DollarsToNibblins(mrcUSD)

	case "Overage":
		overageUSD, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		relay.Overage = routing.DollarsToNibblins(overageUSD)

	case "BWRule":
		bwRule, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		if bwRule < 0 || bwRule > 4 {
			return fmt.Errorf("%d is not a valid BandWidthRule value", int64(bwRule))
		}
		relay.BWRule = routing.BandWidthRule(bwRule)

	case "ContractTerm":
		term, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}
		if term < 0 {
			return fmt.Errorf("%d is not a valid ContractTerm value", int32(term))
		}
		relay.ContractTerm = int32(term)

	case "StartDate":
		startDate, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		if startDate == "" {
			relay.StartDate = time.Time{}
		} else {
			newStartDate, err := time.Parse("January 2, 2006", startDate)
			if err != nil {
				return fmt.Errorf("could not parse `%s` - must be of the form 'January 2, 2006'", startDate)
			}
			relay.StartDate = newStartDate
		}

	case "EndDate":
		endDate, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		if endDate == "" {
			relay.EndDate = time.Time{}
		} else {
			newEndDate, err := time.Parse("January 2, 2006", endDate)
			if err != nil {
				return fmt.Errorf("could not parse `%s` - must be of the form 'January 2, 2006'", endDate)
			}

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
		relay.Type = routing.MachineType(machineType)

	case "Notes":
		notes, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		relay.Notes = notes

	case "BillingSupplier":
		billingSupplier, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		if billingSupplier == "" {
			relay.BillingSupplier = ""
		} else {

			sellerDatabaseID := 0
			for _, seller := range m.Sellers(ctx) {
				if seller.ID == billingSupplier {
					sellerDatabaseID = int(seller.DatabaseID)
				}
			}

			if sellerDatabaseID == 0 {
				return fmt.Errorf("%s is not a valid seller ID", billingSupplier)
			}

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

		relay.Version = version

	case "DestFirst":
		destFirst, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%v is not a valid boolean value", value)
		}

		relay.DestFirst = destFirst

	case "InternalAddressClientRoutable":
		internalAddressClientRoutable, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%v is not a valid boolean value", value)
		}

		if internalAddressClientRoutable && relay.InternalAddr.String() == ":0" {
			// Enforce that the relay has an valid internal address
			return fmt.Errorf("relay must have valid internal address before InternalAddressClientRoutable is true")
		}

		relay.InternalAddressClientRoutable = internalAddressClientRoutable

	default:
		return fmt.Errorf("field '%v' does not exist on the routing.Relay type", field)

	}

	m.localRelays[idx] = relay
	return nil
}

func (m *InMemory) AddBannedUser(ctx context.Context, buyerID uint64, userID uint64) error {
	routeShader, err := m.RouteShader(ctx, buyerID)
	if err != nil {
		return err
	}

	routeShader.BannedUsers[userID] = true

	err = m.AddRouteShader(ctx, routeShader, buyerID)

	return err
}

func (m *InMemory) RemoveBannedUser(ctx context.Context, buyerID uint64, userID uint64) error {
	routeShader, err := m.RouteShader(ctx, buyerID)
	if err != nil {
		return err
	}

	if _, exists := routeShader.BannedUsers[userID]; exists {
		delete(routeShader.BannedUsers, userID)
		return m.AddRouteShader(ctx, routeShader, buyerID)
	}

	return nil
}

func (m *InMemory) BannedUsers(ctx context.Context, buyerID uint64) (map[uint64]bool, error) {
	routeShader, err := m.RouteShader(ctx, buyerID)
	if err != nil {
		return map[uint64]bool{}, err
	}

	return routeShader.BannedUsers, nil
}

func (m *InMemory) UpdateBuyer(ctx context.Context, buyerID uint64, field string, value interface{}) error {
	var buyerExists bool
	var buyer routing.Buyer
	var idx int

	for i, localBuyer := range m.localBuyers {
		if localBuyer.ID == buyerID {

			buyer = localBuyer
			idx = i
			buyerExists = true
			break
		}
	}

	if !buyerExists {
		return &DoesNotExistError{resourceType: "buyer", resourceRef: buyerID}
	}

	switch field {
	case "Live":
		live, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Live: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.Live = live
	case "Debug":
		debug, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Debug: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.Debug = debug
	case "Analytics":
		analytics, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Analytics: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.Analytics = analytics
	case "Billing":
		billing, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Billing: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.Billing = billing
	case "Trial":
		trial, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Trial: %v is not a valid boolean type (%T)", value, value)
		}

		buyer.Trial = trial
	case "ExoticLocationFee":
		exoticLocationFee, ok := value.(float64)
		if !ok {
			return fmt.Errorf("ExoticLocationFee: %v is not a valid float64 type (%T)", value, value)
		}

		buyer.ExoticLocationFee = exoticLocationFee
	case "StandardLocationFee":
		standardLocationFee, ok := value.(float64)
		if !ok {
			return fmt.Errorf("StandardLocationFee: %v is not a valid float64 type (%T)", value, value)
		}

		buyer.StandardLocationFee = standardLocationFee
	case "LookerSeats":
		lookerSeats, ok := value.(int64)
		if !ok {
			return fmt.Errorf("LookerSeats: %v is not a valid int64 type (%T)", value, value)
		}

		buyer.LookerSeats = lookerSeats
	case "ShortName":
		shortName, ok := value.(string)
		if !ok {
			return fmt.Errorf("ShortName: %v is not a valid string type (%T)", value, value)
		}

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
			return fmt.Errorf("PublicKey: failed to decode string public key: %v", err)
		}

		if len(newPublicKey) != crypto.KeySize+8 {
			return fmt.Errorf("PublicKey: public key is not the correct length: %d", len(newPublicKey))
		}

		newBuyerID := binary.LittleEndian.Uint64(newPublicKey[:8])

		buyer.ID = newBuyerID
		buyer.PublicKey = newPublicKey

		// TODO: datacenter maps for this buyer must be updated with the new buyer ID
	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Buyer type", field)
	}

	m.localBuyers[idx] = buyer

	return nil
}

func (m *InMemory) UpdateSeller(ctx context.Context, sellerID string, field string, value interface{}) error {
	var foundSeller bool
	var seller routing.Seller
	var idx int

	for i, localSeller := range m.localSellers {
		if localSeller.ID == sellerID {
			seller = localSeller
			idx = i
			foundSeller = true
			break
		}
	}

	if !foundSeller {
		return &DoesNotExistError{resourceType: "seller", resourceRef: sellerID}
	}

	switch field {
	case "ShortName":
		shortName, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		seller.ShortName = shortName
	case "EgressPriceNibblinsPerGB":
		egressPrice, ok := value.(float64)
		if !ok {
			return fmt.Errorf("%v is not a valid float64 type", value)
		}

		egress := routing.DollarsToNibblins(egressPrice)
		seller.EgressPriceNibblinsPerGB = egress
	case "Secret":
		secret, ok := value.(bool)
		if !ok {
			return fmt.Errorf("%v is not a valid boolean type", value)
		}

		seller.Secret = secret
	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Seller type", field)

	}

	m.localSellers[idx] = seller
	return nil
}

func (m *InMemory) UpdateCustomer(ctx context.Context, customerID string, field string, value interface{}) error {
	var foundCustomer bool
	var customer routing.Customer
	var idx int

	for i, localCustomer := range m.localCustomers {
		if localCustomer.Code == customerID {
			customer = localCustomer
			idx = i
			foundCustomer = true
			break
		}
	}

	if !foundCustomer {
		return &DoesNotExistError{resourceType: "customer", resourceRef: customerID}
	}

	switch field {
	case "AutomaticSigninDomains":
		domains, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		customer.AutomaticSignInDomains = domains
	case "Name":
		name, ok := value.(string)
		if !ok {
			return fmt.Errorf("%v is not a valid string value", value)
		}

		customer.Name = name
	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Customer type", field)
	}

	m.localCustomers[idx] = customer
	return nil
}

func (m *InMemory) UpdateDatacenter(ctx context.Context, datacenterID uint64, field string, value interface{}) error {
	var foundDatacenter bool
	var datacenter routing.Datacenter
	var idx int

	for i, localDatacenter := range m.localDatacenters {
		if localDatacenter.ID == datacenterID {
			datacenter = localDatacenter
			idx = i
			foundDatacenter = true
			break
		}
	}

	if !foundDatacenter {
		return &DoesNotExistError{resourceType: "datacenter", resourceRef: datacenterID}
	}

	switch field {
	case "Latitude":
		latitude, ok := value.(float32)
		if !ok {
			return fmt.Errorf("%v is not a valid float32 value", value)
		}

		datacenter.Location.Latitude = latitude
	case "Longitude":
		longitude, ok := value.(float32)
		if !ok {
			return fmt.Errorf("%v is not a valid float32 value", value)
		}

		datacenter.Location.Longitude = longitude
	default:
		return fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Datacenter type", field)
	}

	m.localDatacenters[idx] = datacenter
	return nil
}

func (m *InMemory) GetDatabaseBinFileMetaData(ctx context.Context) (routing.DatabaseBinFileMetaData, error) {
	return routing.DatabaseBinFileMetaData{}, fmt.Errorf("GetDatabaseBinFileMetaData not implemented in InMemory storer")
}

func (m *InMemory) UpdateDatabaseBinFileMetaData(ctx context.Context, fileMeta routing.DatabaseBinFileMetaData) error {
	return fmt.Errorf("UpdateDatabaseBinFileMetaData not implemented in InMemory storer")
}

// GetAnalyticsDashboardCategories returns all Looker dashboard categories
func (m *InMemory) GetAnalyticsDashboardCategories(ctx context.Context) ([]looker.AnalyticsDashboardCategory, error) {
	categories := make([]looker.AnalyticsDashboardCategory, 0)
	return categories, fmt.Errorf("GetAnalyticsDashboardCategories not implemented in InMemory storer")
}

// GetPremiumAnalyticsDashboardCategories returns all Looker dashboard categories
func (m *InMemory) GetPremiumAnalyticsDashboardCategories(ctx context.Context) ([]looker.AnalyticsDashboardCategory, error) {
	categories := make([]looker.AnalyticsDashboardCategory, 0)
	return categories, fmt.Errorf("GetPremiumAnalyticsDashboardCategories not implemented in InMemory storer")
}

// GetFreeAnalyticsDashboardCategories returns all free Looker dashboard categories
func (m *InMemory) GetFreeAnalyticsDashboardCategories(ctx context.Context) ([]looker.AnalyticsDashboardCategory, error) {
	categories := make([]looker.AnalyticsDashboardCategory, 0)
	return categories, fmt.Errorf("GetFreeAnalyticsDashboardCategories not implemented in InMemory storer")
}

// GetAnalyticsDashboardCategories returns all Looker dashboard categories
func (m *InMemory) GetAnalyticsDashboardCategoryByID(ctx context.Context, id int64) (looker.AnalyticsDashboardCategory, error) {
	category := looker.AnalyticsDashboardCategory{}
	return category, fmt.Errorf("GetAnalyticsDashboardCategoryByID not implemented in InMemory storer")
}

// GetAnalyticsDashboardCategories returns all Looker dashboard categories
func (m *InMemory) GetAnalyticsDashboardCategoryByLabel(ctx context.Context, label string) (looker.AnalyticsDashboardCategory, error) {
	category := looker.AnalyticsDashboardCategory{}
	return category, fmt.Errorf("GetAnalyticsDashboardCategoryByLabel not implemented in InMemory storer")
}

// AddAnalyticsDashboardCategory adds a new dashboard category
func (m *InMemory) AddAnalyticsDashboardCategory(ctx context.Context, label string, isAdmin bool, isPremium bool, isSeller bool) error {
	return fmt.Errorf("AddAnalyticsDashboardCategory not implemented in InMemory storer")
}

// RemoveAnalyticsDashboardCategory remove a dashboard category by ID
func (m *InMemory) RemoveAnalyticsDashboardCategoryByID(ctx context.Context, id int64) error {
	return fmt.Errorf("RemoveAnalyticsDashboardCategory not implemented in InMemory storer")
}

// RemoveAnalyticsDashboardCategory remove a dashboard category by label
func (m *InMemory) RemoveAnalyticsDashboardCategoryByLabel(ctx context.Context, label string) error {
	return fmt.Errorf("RemoveAnalyticsDashboardCategory not implemented in InMemory storer")

}

// UpdateAnalyticsDashboardCategoryByID update dashboard category by ID
func (m *InMemory) UpdateAnalyticsDashboardCategoryByID(ctx context.Context, id int64, field string, value interface{}) error {
	return fmt.Errorf("UpdateAnalyticsDashboardCategoryByID not implemented in InMemory storer")
}

// GetAnalyticsDashboardsByCategoryID get all looker dashboards by category id
func (m *InMemory) GetAnalyticsDashboardsByCategoryID(ctx context.Context, id int64) ([]looker.AnalyticsDashboard, error) {
	dashboards := make([]looker.AnalyticsDashboard, 0)
	return dashboards, fmt.Errorf("GetAnalyticsDashboardsByCategoryID not implemented in InMemory storer")
}

// GetAnalyticsDashboardsByCategoryLabel get all looker dashboards by category label
func (m *InMemory) GetAnalyticsDashboardsByCategoryLabel(ctx context.Context, label string) ([]looker.AnalyticsDashboard, error) {
	dashboards := make([]looker.AnalyticsDashboard, 0)
	return dashboards, fmt.Errorf("GetAnalyticsDashboardsByCategoryLabel not implemented in InMemory storer")
}

// GetPremiumAnalyticsDashboards get all premium looker dashboards
func (m *InMemory) GetPremiumAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error) {
	dashboards := make([]looker.AnalyticsDashboard, 0)
	return dashboards, fmt.Errorf("GetPremiumAnalyticsDashboards not implemented in InMemory storer")
}

// GetFreeAnalyticsDashboards get all free looker dashboards
func (m *InMemory) GetFreeAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error) {
	dashboards := make([]looker.AnalyticsDashboard, 0)
	return dashboards, fmt.Errorf("GetFreeAnalyticsDashboards not implemented in InMemory storer")
}

// GetDiscoveryAnalyticsDashboards get all discovery looker dashboards
func (m *InMemory) GetDiscoveryAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error) {
	dashboards := make([]looker.AnalyticsDashboard, 0)
	return dashboards, fmt.Errorf("GetDiscoveryAnalyticsDashboards not implemented in InMemory storer")
}

// GetDiscoveryAnalyticsDashboards get all discovery looker dashboards
func (m *InMemory) GetAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error) {
	dashboards := make([]looker.AnalyticsDashboard, 0)
	return dashboards, fmt.Errorf("GetAnalyticsDashboards not implemented in InMemory storer")
}

// GetAdminAnalyticsDashboards get all admin looker dashboards
func (m *InMemory) GetAdminAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error) {
	dashboards := make([]looker.AnalyticsDashboard, 0)
	return dashboards, fmt.Errorf("GetAdminAnalyticsDashboards not implemented in InMemory storer")
}

// GetAnalyticsDashboardsByLookerID get looker dashboard by looker id
func (m *InMemory) GetAnalyticsDashboardsByLookerID(ctx context.Context, id string) ([]looker.AnalyticsDashboard, error) {
	dashboards := []looker.AnalyticsDashboard{}
	return dashboards, fmt.Errorf("GetAnalyticsDashboardsByLookerID not implemented in InMemory storer")
}

// GetAnalyticsDashboardByID get looker dashboard by id
func (m *InMemory) GetAnalyticsDashboardByID(ctx context.Context, id int64) (looker.AnalyticsDashboard, error) {
	dashboard := looker.AnalyticsDashboard{}
	return dashboard, fmt.Errorf("GetAnalyticsDashboardByID not implemented in InMemory storer")
}

// GetAnalyticsDashboardByName get looker dashboard by name
func (m *InMemory) GetAnalyticsDashboardByName(ctx context.Context, name string) (looker.AnalyticsDashboard, error) {
	dashboard := looker.AnalyticsDashboard{}
	return dashboard, fmt.Errorf("GetAnalyticsDashboardByName not implemented in InMemory storer")
}

// AddAnalyticsDashboard adds a new dashboard
func (m *InMemory) AddAnalyticsDashboard(ctx context.Context, name string, lookerID int64, isDiscover bool, customerID int64, categoryID int64) error {
	return fmt.Errorf("AddAnalyticsDashboard not implemented in InMemory storer")
}

// RemoveAnalyticsDashboardByID remove looker dashboard by id
func (m *InMemory) RemoveAnalyticsDashboardByID(ctx context.Context, id int64) error {
	return fmt.Errorf("RemoveAnalyticsDashboardByID not implemented in InMemory storer")
}

// RemoveAnalyticsDashboardByName remove looker dashboard by name
func (m *InMemory) RemoveAnalyticsDashboardByName(ctx context.Context, name string) error {
	return fmt.Errorf("RemoveAnalyticsDashboardByName not implemented in InMemory storer")
}

// UpdateAnalyticsDashboardByID update looker dashboard looker id by dashboard id
func (m *InMemory) UpdateAnalyticsDashboardByID(ctx context.Context, id int64, field string, value interface{}) error {
	return fmt.Errorf("UpdateAnalyticsDashboardByID not implemented in InMemory storer")
}
