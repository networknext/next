package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
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

func (m *InMemory) Buyer(id uint64) (routing.Buyer, error) {
	for _, buyer := range m.localBuyers {
		if buyer.ID == id {
			return buyer, nil
		}
	}

	return routing.Buyer{}, &DoesNotExistError{resourceType: "buyer", resourceRef: id}
}

func (m *InMemory) Buyers() []routing.Buyer {
	buyers := make([]routing.Buyer, len(m.localBuyers))
	for i := range buyers {
		buyers[i] = m.localBuyers[i]
	}

	return buyers
}

func (m *InMemory) BuyerWithCompanyCode(code string) (routing.Buyer, error) {
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

func (m *InMemory) Seller(id string) (routing.Seller, error) {
	for _, seller := range m.localSellers {
		if seller.ID == id {
			return seller, nil
		}
	}

	return routing.Seller{}, &DoesNotExistError{resourceType: "seller", resourceRef: id}
}

func (m *InMemory) Sellers() []routing.Seller {
	sellers := make([]routing.Seller, len(m.localSellers))
	for i := range sellers {
		sellers[i] = m.localSellers[i]
	}

	return sellers
}

func (m *InMemory) SellerWithCompanyCode(code string) (routing.Seller, error) {
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

func (m *InMemory) Customer(code string) (routing.Customer, error) {
	for _, customer := range m.localCustomers {
		if customer.Code == code {
			return customer, nil
		}
	}

	return routing.Customer{}, &DoesNotExistError{resourceType: "customer", resourceRef: code}
}

func (m *InMemory) Customers() []routing.Customer {
	customers := make([]routing.Customer, len(m.localCustomers))
	for i := range customers {
		customers[i] = m.localCustomers[i]
	}

	return customers
}

func (m *InMemory) CustomerWithName(name string) (routing.Customer, error) {
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

func (m *InMemory) Relay(id uint64) (routing.Relay, error) {
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

func (m *InMemory) Relays() []routing.Relay {
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
		if dc.BuyerID == dcMap.BuyerID && dc.Alias == dcMap.Alias && dc.DatacenterID == dcMap.DatacenterID {
			return &AlreadyExistsError{resourceType: "datacenterMap", resourceRef: dcMap.Alias}
		}
	}

	m.localDatacenterMaps = append(m.localDatacenterMaps, dcMap)

	return nil

}

func (m *InMemory) GetDatacenterMapsForBuyer(id uint64) map[uint64]routing.DatacenterMap {
	var dcs = make(map[uint64]routing.DatacenterMap)
	for _, dc := range m.localDatacenterMaps {
		if dc.BuyerID == id {
			id := crypto.HashID(dc.Alias + fmt.Sprintf("%x", dc.BuyerID) + fmt.Sprintf("%x", dc.DatacenterID))
			dcs[id] = dc
		}
	}

	return dcs
}

func (m *InMemory) ListDatacenterMaps(dcID uint64) map[uint64]routing.DatacenterMap {
	var dcs = make(map[uint64]routing.DatacenterMap)
	for _, dc := range m.localDatacenterMaps {
		if dc.DatacenterID == dcID || dcID == 0 {
			id := crypto.HashID(dc.Alias + fmt.Sprintf("%x", dc.BuyerID) + fmt.Sprintf("%x", dc.DatacenterID))
			dcs[id] = dc
		}
	}

	return dcs
}

func (m *InMemory) RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error {

	idx := -1
	for i, dcm := range m.localDatacenterMaps {
		if dcMap.Alias == dcm.Alias && dcMap.BuyerID == dcm.BuyerID && dcMap.DatacenterID == dcm.DatacenterID {
			idx = i
		}
	}

	if idx < 0 {
		return &DoesNotExistError{resourceType: "datacenterMap", resourceRef: dcMap.Alias}
	}

	if idx+1 == len(m.localDatacenterMaps) {
		m.localDatacenterMaps = m.localDatacenterMaps[:idx]
		return nil
	}

	m.localDatacenterMaps = append(m.localDatacenterMaps[:idx], m.localDatacenterMaps[idx+1:]...)
	return nil

}

func (m *InMemory) Datacenter(id uint64) (routing.Datacenter, error) {
	for _, datacenter := range m.localDatacenters {
		if datacenter.ID == id {
			return datacenter, nil
		}
	}

	return routing.Datacenter{}, &DoesNotExistError{resourceType: "datacenter", resourceRef: id}
}

func (m *InMemory) Datacenters() []routing.Datacenter {
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

func (m *InMemory) GetFeatureFlags() map[string]bool {
	return map[string]bool{}
}

func (m *InMemory) GetFeatureFlagByName(flagName string) (map[string]bool, error) {
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

func (m *InMemory) UpdateRelay(ctx context.Context, relayID uint64, field string, value interface{}) error {
	return fmt.Errorf(("UpdateRelay not impemented in Firestore storer"))
}
