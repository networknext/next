package storage

import (
	"context"
	"fmt"
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

	// Emulate firestore behavior by requiring the seller and datacenter to exist before adding the relay
	foundSeller := false
	for _, s := range m.localSellers {
		if s.ID == relay.Seller.ID {
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
	return core.InternalConfig{}, fmt.Errorf(("InternalConfig not impemented in InMemory storer"))
}

func (m *InMemory) RouteShader(ctx context.Context, buyerID uint64) (core.RouteShader, error) {
	return core.RouteShader{}, fmt.Errorf(("RouteShaders not impemented in InMemory storer"))
}

func (m *InMemory) AddInternalConfig(ctx context.Context, internalConfig core.InternalConfig, buyerID uint64) error {
	return fmt.Errorf("AddInternalConfig not yet impemented in InMemory storer")
}

func (m *InMemory) UpdateInternalConfig(ctx context.Context, buyerID uint64, field string, value interface{}) error {
	return fmt.Errorf("UpdateInternalConfig not yet impemented in InMemory storer")
}

func (m *InMemory) RemoveInternalConfig(ctx context.Context, buyerID uint64) error {
	return fmt.Errorf("RemoveInternalConfig not yet impemented in InMemory storer")
}

func (m *InMemory) AddRouteShader(ctx context.Context, routeShader core.RouteShader, buyerID uint64) error {
	return fmt.Errorf("AddRouteShader not yet impemented in InMemory storer")
}

func (m *InMemory) UpdateRouteShader(ctx context.Context, buyerID uint64, field string, value interface{}) error {
	return fmt.Errorf("UpdateRouteShader not yet impemented in InMemory storer")
}

func (m *InMemory) RemoveRouteShader(ctx context.Context, buyerID uint64) error {
	return fmt.Errorf("RemoveRouteShader not yet impemented in InMemory storer")
}

func (m *InMemory) UpdateRelay(ctx context.Context, relayID uint64, field string, value interface{}) error {
	return fmt.Errorf(("UpdateRelay not impemented in Firestore storer"))
}

func (m *InMemory) AddBannedUser(ctx context.Context, buyerID uint64, userID uint64) error {
	return fmt.Errorf(("AddBannedUser not yet impemented in InMemory storer"))
}

func (m *InMemory) RemoveBannedUser(ctx context.Context, buyerID uint64, userID uint64) error {
	return fmt.Errorf(("RemoveBannedUser not yet impemented in InMemory storer"))
}

func (m *InMemory) BannedUsers(ctx context.Context, buyerID uint64) (map[uint64]bool, error) {
	return map[uint64]bool{}, fmt.Errorf(("BannedUsers not yet impemented in InMemory storer"))
}

func (m *InMemory) UpdateBuyer(ctx context.Context, buyerID uint64, field string, value interface{}) error {
	return fmt.Errorf("UpdateBuyer not impemented in InMemory storer")
}

func (m *InMemory) UpdateSeller(ctx context.Context, sellerID string, field string, value interface{}) error {
	return fmt.Errorf("UpdateSeller not impemented in InMemory storer")
}

func (m *InMemory) UpdateCustomer(ctx context.Context, customerID string, field string, value interface{}) error {
	return fmt.Errorf("UpdateCustomer not impemented in InMemory storer")
}

func (m *InMemory) UpdateDatacenter(ctx context.Context, datacenterID uint64, field string, value interface{}) error {
	return fmt.Errorf("UpdateDatacenter not impemented in InMemory storer")
}

func (m *InMemory) UpdateDatacenterMap(ctx context.Context, buyerID uint64, datacenterID uint64, field string, value interface{}) error {
	return fmt.Errorf("UpdateDatacenterMap not implemented in InMemory storer")
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
func (m *InMemory) AddAnalyticsDashboardCategory(ctx context.Context, label string, isAdmin bool, isPremium bool) error {
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
