//go:generate moq -out storage_test_mocks.go . Storer
package storage

import (
	"context"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport/looker"
)

type Storer interface {
	DatabaseBinFileReference(ctx context.Context) (routing.DatabaseBinWrapperReference, error)

	Customer(ctx context.Context, code string) (routing.Customer, error)

	CustomerByID(ctx context.Context, id int64) (routing.Customer, error)

	Customers(ctx context.Context) []routing.Customer

	AddCustomer(ctx context.Context, customer routing.Customer) error

	RemoveCustomer(ctx context.Context, code string) error

	// TODO: chopping block (this is dangerous)
	SetCustomer(ctx context.Context, customer routing.Customer) error

	// UpdateCustomer modifies the givien field for the specified buyer
	UpdateCustomer(ctx context.Context, customerID string, field string, value interface{}) error

	// Buyer gets a copy of a buyer with the specified buyer ID,
	// and returns an empty buyer and an error if a buyer with that ID doesn't exist in storage.
	Buyer(ctx context.Context, id uint64) (routing.Buyer, error)

	// BuyerWithCompanyCode gets the Buyer with the matching company code
	BuyerWithCompanyCode(ctx context.Context, code string) (routing.Buyer, error)

	// Buyers returns a copy of all stored buyers.
	Buyers(ctx context.Context) []routing.Buyer

	// AddBuyer adds the provided buyer to storage and returns an error if the buyer could not be added.
	AddBuyer(ctx context.Context, buyer routing.Buyer) error

	// RemoveBuyer removes a buyer with the provided buyer ID from storage and returns an error if the buyer could not be removed.
	RemoveBuyer(ctx context.Context, id uint64) error

	// UpdateBuyer modifies the givien field for the specified buyer
	UpdateBuyer(ctx context.Context, buyerID uint64, field string, value interface{}) error

	// Seller gets a copy of a seller with the specified seller ID,
	// and returns an empty seller and an error if a seller with that ID doesn't exist in storage.
	Seller(ctx context.Context, id string) (routing.Seller, error)

	// Sellers returns a copy of all stored sellers.
	Sellers(ctx context.Context) []routing.Seller

	// AddSeller adds the provided seller to storage and returns an error if the seller could not be added.
	AddSeller(ctx context.Context, seller routing.Seller) error

	// RemoveSeller removes a seller with the provided seller ID from storage and returns an error if the seller could not be removed.
	RemoveSeller(ctx context.Context, id string) error

	// UpdateSeller modifies the givien field for the specified buyer
	UpdateSeller(ctx context.Context, sellerID string, field string, value interface{}) error

	// BuyerIDFromCustomerName returns the buyer ID associated with the given customer name and an error if the customer wasn't found.
	// If the customer has no buyer linked, then it will return a buyer ID of 0 and no error.
	BuyerIDFromCustomerName(ctx context.Context, customerName string) (uint64, error)

	// SellerIDFromCustomerName returns the seller ID associated with the given customer name and an error if the customer wasn't found.
	// If the customer has no seller linked, then it will return an empty seller ID and no error.
	SellerIDFromCustomerName(ctx context.Context, customerName string) (string, error)

	SellerWithCompanyCode(ctx context.Context, code string) (routing.Seller, error)

	// SetCustomerLink update the customer's buyer and seller references.
	// TODO: chopping block (handled/required by database)
	SetCustomerLink(ctx context.Context, customerName string, buyerID uint64, sellerID string) error

	// Relay gets a copy of a relay with the specified relay ID
	// and returns an empty relay and an error if a relay with that ID doesn't exist in storage.
	Relay(ctx context.Context, id uint64) (routing.Relay, error)

	// Relays returns a copy of all stored relays.
	Relays(ctx context.Context) []routing.Relay

	// AddRelay adds the provided relay to storage and returns an error if the relay could not be added.
	AddRelay(ctx context.Context, relay routing.Relay) error

	// RemoveRelay removes a relay with the provided relay ID from storage and returns an error if the relay could not be removed.
	RemoveRelay(ctx context.Context, id uint64) error

	// UpdateRelay updates a single field in a relay record
	UpdateRelay(ctx context.Context, relayID uint64, field string, value interface{}) error

	// SetRelay updates the relay in storage with the provided copy and returns an error if the relay could not be updated.
	// TODO: chopping block (obsoleted by UpdateRelay, and broken anyway)
	SetRelay(ctx context.Context, relay routing.Relay) error

	// Datacenter gets a copy of a datacenter with the specified datacenter ID
	// and returns an empty datacenter and an error if a datacenter with that ID doesn't exist in storage.
	Datacenter(ctx context.Context, datacenterID uint64) (routing.Datacenter, error)

	// Datacenters returns a copy of all stored datacenters.
	Datacenters(ctx context.Context) []routing.Datacenter

	// AddDatacenter adds the provided datacenter to storage and returns an error if the datacenter could not be added.
	AddDatacenter(ctx context.Context, datacenter routing.Datacenter) error

	// UpdateDatacenter modifies the givien field for the specified datacenter
	UpdateDatacenter(ctx context.Context, datacenterID uint64, field string, value interface{}) error

	// RemoveDatacenter removes a datacenter with the provided datacenter ID from storage and returns an error if the datacenter could not be removed.
	RemoveDatacenter(ctx context.Context, id uint64) error

	// GetDatacenterMapsForBuyer returns the list of datacenter aliases in use for a given (internally generated) buyerID. Returns
	// an empty []routing.DatacenterMap if there are no aliases for that buyerID.
	GetDatacenterMapsForBuyer(ctx context.Context, buyerID uint64) map[uint64]routing.DatacenterMap

	// AddDatacenterMap adds a new datacenter alias for the given buyer and datacenter IDs
	AddDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error

	// ListDatacenterMaps returns a list of alias/buyer mappings for the specified datacenter ID. An
	// empty dcID returns a list of all maps.
	ListDatacenterMaps(ctx context.Context, dcID uint64) map[uint64]routing.DatacenterMap

	// RemoveDatacenterMap removes an entry from the DatacenterMaps table
	RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error

	// New for ConfigService

	// GetFeatureFlags returns all feature flags currently in the database
	GetFeatureFlags(ctx context.Context) map[string]bool

	// GetFeatureFlagByName returns a specific flag or an error if it does not exist
	GetFeatureFlagByName(ctx context.Context, flagName string) (map[string]bool, error)

	// SetFeatureFlagByName adds a new feature or updates the value of an existing feature
	SetFeatureFlagByName(ctx context.Context, flagName string, flagVal bool) error

	// RemoveFeatureFlagByName removes an existing flag from storage
	RemoveFeatureFlagByName(ctx context.Context, flagName string) error

	// InternalConfig returns the internal config for the given buyer ID
	InternalConfig(ctx context.Context, buyerID uint64) (core.InternalConfig, error)

	// AddInternalConfig adds the provided InternalConfig to the database
	AddInternalConfig(ctx context.Context, internalConfig core.InternalConfig, buyerID uint64) error

	// UpdateInternalConfig updates the specified field in an InternalConfig record
	UpdateInternalConfig(ctx context.Context, buyerID uint64, field string, value interface{}) error

	// RemoveInternalConfig removes a record from the InternalConfigs table
	RemoveInternalConfig(ctx context.Context, buyerID uint64) error

	// RouteShader returns a slice of route shaders for the given buyer ID
	RouteShader(ctx context.Context, buyerID uint64) (core.RouteShader, error)

	// AddRouteShader adds the provided RouteShader to the database
	AddRouteShader(ctx context.Context, routeShader core.RouteShader, buyerID uint64) error

	// UpdateRouteShader updates the specified field in an RouteShader record
	UpdateRouteShader(ctx context.Context, buyerID uint64, field string, value interface{}) error

	// RemoveRouteShader removes a record from the RouteShaders table
	RemoveRouteShader(ctx context.Context, buyerID uint64) error

	// AddBannedUser adds a user to the banned_user table
	AddBannedUser(ctx context.Context, buyerID uint64, userID uint64) error

	// RemoveBannedUser removes a user from the banned_user table
	RemoveBannedUser(ctx context.Context, buyerID uint64, userID uint64) error

	// BannedUsers returns the set of banned users for the specified buyer ID. This method
	// is designed to be used by syncRouteShaders() though it can be used by client code.
	BannedUsers(ctx context.Context, buyerID uint64) (map[uint64]bool, error)

	// GetDatabaseBinFileMetaData returns data from the database_bin_meta table
	GetDatabaseBinFileMetaData(ctx context.Context) (routing.DatabaseBinFileMetaData, error)

	// UpdateDatabaseBinFileMetaData updates the specified field in an database_bin_meta table
	UpdateDatabaseBinFileMetaData(context.Context, routing.DatabaseBinFileMetaData) error

	// GetAnalyticsDashboardCategories returns all Looker dashboard categories
	GetAnalyticsDashboardCategories(ctx context.Context) ([]looker.AnalyticsDashboardCategory, error)

	// GetPremiumAnalyticsDashboardCategories returns all Looker dashboard categories
	GetPremiumAnalyticsDashboardCategories(ctx context.Context) ([]looker.AnalyticsDashboardCategory, error)

	// GetFreeAnalyticsDashboardCategories returns all Looker dashboard categories
	GetFreeAnalyticsDashboardCategories(ctx context.Context) ([]looker.AnalyticsDashboardCategory, error)

	// GetAnalyticsDashboardCategories returns all Looker dashboard categories
	GetAnalyticsDashboardCategoryByID(ctx context.Context, id int64) (looker.AnalyticsDashboardCategory, error)

	// GetAnalyticsDashboardCategories returns all Looker dashboard categories
	GetAnalyticsDashboardCategoryByLabel(ctx context.Context, label string) (looker.AnalyticsDashboardCategory, error)

	// AddAnalyticsDashboardCategory adds a new dashboard category
	AddAnalyticsDashboardCategory(ctx context.Context, label string, isAdmin bool, isPremium bool, isSeller bool) error

	// RemoveAnalyticsDashboardCategory remove a dashboard category by ID
	RemoveAnalyticsDashboardCategoryByID(ctx context.Context, id int64) error

	// RemoveAnalyticsDashboardCategory remove a dashboard category by label
	RemoveAnalyticsDashboardCategoryByLabel(ctx context.Context, label string) error

	// UpdateAnalyticsDashboardByID update a dashboard category label by id
	UpdateAnalyticsDashboardCategoryByID(ctx context.Context, id int64, field string, value interface{}) error

	// GetAnalyticsDashboardsByCategoryID get all looker dashboards by category id
	GetAnalyticsDashboardsByCategoryID(ctx context.Context, id int64) ([]looker.AnalyticsDashboard, error)

	// GetAnalyticsDashboardsByCategoryLabel get all looker dashboards by category label
	GetAnalyticsDashboardsByCategoryLabel(ctx context.Context, label string) ([]looker.AnalyticsDashboard, error)

	// GetPremiumAnalyticsDashboards get all premium looker dashboards
	GetPremiumAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error)

	// GetFreeAnalyticsDashboards get all free looker dashboards
	GetFreeAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error)

	// GetDiscoveryAnalyticsDashboards get all discovery looker dashboards
	GetDiscoveryAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error)

	// GetAdminAnalyticsDashboards get all admin looker dashboards
	GetAdminAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error)

	// GetAnalyticsDashboardByLookerID get looker dashboard by looker id
	GetAnalyticsDashboardsByLookerID(ctx context.Context, id string) ([]looker.AnalyticsDashboard, error)

	// GetDiscoveryAnalyticsDashboards get all discovery looker dashboards
	GetAnalyticsDashboards(ctx context.Context) ([]looker.AnalyticsDashboard, error)

	// GetAnalyticsDashboardByID get looker dashboard by id
	GetAnalyticsDashboardByID(ctx context.Context, id int64) (looker.AnalyticsDashboard, error)

	// GetAnalyticsDashboardByName get looker dashboard by name
	GetAnalyticsDashboardByName(ctx context.Context, name string) (looker.AnalyticsDashboard, error)

	// AddAnalyticsDashboard adds a new dashboard
	AddAnalyticsDashboard(ctx context.Context, name string, lookerID int64, isDiscover bool, customerID int64, categoryID int64) error

	// RemoveAnalyticsDashboardByID remove looker dashboard by id
	RemoveAnalyticsDashboardByID(ctx context.Context, id int64) error

	// RemoveAnalyticsDashboardByName remove looker dashboard by name
	RemoveAnalyticsDashboardByName(ctx context.Context, name string) error

	// UpdateAnalyticsDashboardByID update looker dashboard looker id by dashboard id
	UpdateAnalyticsDashboardByID(ctx context.Context, id int64, field string, value interface{}) error
}
