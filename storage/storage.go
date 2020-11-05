package storage

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/routing"
)

type Storer interface {
	Customer(code string) (routing.Customer, error)

	CustomerWithName(name string) (routing.Customer, error)

	Customers() []routing.Customer

	AddCustomer(ctx context.Context, customer routing.Customer) error

	RemoveCustomer(ctx context.Context, code string) error

	SetCustomer(ctx context.Context, customer routing.Customer) error

	// Buyer gets a copy of a buyer with the specified buyer ID,
	// and returns an empty buyer and an error if a buyer with that ID doesn't exist in storage.
	Buyer(id uint64) (routing.Buyer, error)

	// BuyerWithCompanyCode gets the Buyer with the matching company code
	BuyerWithCompanyCode(code string) (routing.Buyer, error)

	// Buyers returns a copy of all stored buyers.
	Buyers() []routing.Buyer

	// AddBuyer adds the provided buyer to storage and returns an error if the buyer could not be added.
	AddBuyer(ctx context.Context, buyer routing.Buyer) error

	// RemoveBuyer removes a buyer with the provided buyer ID from storage and returns an error if the buyer could not be removed.
	RemoveBuyer(ctx context.Context, id uint64) error

	// SetBuyer updates the buyer in storage with the provided copy and returns an error if the buyer could not be updated.
	SetBuyer(ctx context.Context, buyer routing.Buyer) error

	// Seller gets a copy of a seller with the specified seller ID,
	// and returns an empty seller and an error if a seller with that ID doesn't exist in storage.
	Seller(id string) (routing.Seller, error)

	// Sellers returns a copy of all stored sellers.
	Sellers() []routing.Seller

	// AddSeller adds the provided seller to storage and returns an error if the seller could not be added.
	AddSeller(ctx context.Context, seller routing.Seller) error

	// RemoveSeller removes a seller with the provided seller ID from storage and returns an error if the seller could not be removed.
	RemoveSeller(ctx context.Context, id string) error

	// SetSeller updates the seller in storage with the provided copy and returns an error if the seller could not be updated.
	SetSeller(ctx context.Context, seller routing.Seller) error

	// BuyerIDFromCustomerName returns the buyer ID associated with the given customer name and an error if the customer wasn't found.
	// If the customer has no buyer linked, then it will return a buyer ID of 0 and no error.
	BuyerIDFromCustomerName(ctx context.Context, customerName string) (uint64, error)

	// SellerIDFromCustomerName returns the seller ID associated with the given customer name and an error if the customer wasn't found.
	// If the customer has no seller linked, then it will return an empty seller ID and no error.
	SellerIDFromCustomerName(ctx context.Context, customerName string) (string, error)

	SellerWithCompanyCode(code string) (routing.Seller, error)

	// SetCustomerLink update the customer's buyer and seller references.
	SetCustomerLink(ctx context.Context, customerName string, buyerID uint64, sellerID string) error

	// Relay gets a copy of a relay with the specified relay ID
	// and returns an empty relay and an error if a relay with that ID doesn't exist in storage.
	Relay(id uint64) (routing.Relay, error)

	// Relays returns a copy of all stored relays.
	Relays() []routing.Relay

	// AddRelay adds the provided relay to storage and returns an error if the relay could not be added.
	AddRelay(ctx context.Context, relay routing.Relay) error

	// RemoveRelay removes a relay with the provided relay ID from storage and returns an error if the relay could not be removed.
	RemoveRelay(ctx context.Context, id uint64) error

	// SetRelay updates the relay in storage with the provided copy and returns an error if the relay could not be updated.
	SetRelay(ctx context.Context, relay routing.Relay) error

	// Datacenter gets a copy of a datacenter with the specified datacenter ID
	// and returns an empty datacenter and an error if a datacenter with that ID doesn't exist in storage.
	Datacenter(datacenterID uint64) (routing.Datacenter, error)

	// Datacenters returns a copy of all stored datacenters.
	Datacenters() []routing.Datacenter

	// AddDatacenter adds the provided datacenter to storage and returns an error if the datacenter could not be added.
	AddDatacenter(ctx context.Context, datacenter routing.Datacenter) error

	// RemoveDatacenter removes a datacenter with the provided datacenter ID from storage and returns an error if the datacenter could not be removed.
	RemoveDatacenter(ctx context.Context, id uint64) error

	// SetDatacenter updates the datacenter in storage with the provided copy and returns an error if the datacenter could not be updated.
	SetDatacenter(ctx context.Context, datacenter routing.Datacenter) error

	// DatacenterMaps returns the list of datacenter aliases in use for a given (internally generated) buyerID. Returns
	// an empty []routing.DatacenterMap if there are no aliases for that buyerID.
	GetDatacenterMapsForBuyer(buyerID uint64) map[uint64]routing.DatacenterMap

	// AddDatacenterMap adds a new datacenter alias for the given buyer and datacenter IDs
	AddDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error

	// ListDatacenterMaps returns a list of alias/buyer mappings for the specified datacenter ID. An
	// empty dcID returns a list of all maps.
	ListDatacenterMaps(dcID uint64) map[uint64]routing.DatacenterMap

	// RemoveDatacenterMap removes an entry from the DatacenterMaps table
	RemoveDatacenterMap(ctx context.Context, dcMap routing.DatacenterMap) error

	// SetRelayMetadata provides write access to ops metadat (mrc, overage, etc)
	SetRelayMetadata(ctx context.Context, relay routing.Relay) error
}

type UnmarshalError struct {
	err error
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("unmarshal error: %v", e.err)
}

type DoesNotExistError struct {
	resourceType string
	resourceRef  interface{}
}

func (e *DoesNotExistError) Error() string {
	return fmt.Sprintf("%s with reference %v not found", e.resourceType, e.resourceRef)
}

type AlreadyExistsError struct {
	resourceType string
	resourceRef  interface{}
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("%s with reference %v already exists", e.resourceType, e.resourceRef)
}

type HexStringConversionError struct {
	hexString string
}

func (e *HexStringConversionError) Error() string {
	return fmt.Sprintf("error converting hex string %s to uint64", e.hexString)
}

type SequenceNumbersOutOfSync struct {
	localSequenceNumber  int64
	remoteSequenceNumber int64
}

func (e *SequenceNumbersOutOfSync) Error() string {
	return fmt.Sprintf("sequence number out of sync: remote %d != local %d", e.remoteSequenceNumber, e.localSequenceNumber)
}

func SeedStorage(logger log.Logger, ctx context.Context, db Storer, relayPublicKey []byte, customerID uint64, customerPublicKey []byte) {
	routeShader := core.NewRouteShader()
	internalConfig := core.NewInternalConfig()
	internalConfig.ForceNext = true

	shouldFill := false
	switch db := db.(type) {
	case *Firestore:
		level.Info(logger).Log("msg", "adding sequence number to firestore emulator")
		_, _, err := db.CheckSequenceNumber(ctx)
		if err != nil {
			level.Error(logger).Log("msg", "unable to check sequence number, attempting to reset value", "err", err)
			if err := db.SetSequenceNumber(ctx, 0); err != nil {
				level.Error(logger).Log("msg", "unable to set sequence number", "err", err)
			}
			if err := db.IncrementSequenceNumber(ctx); err != nil {
				level.Error(logger).Log("msg", "unable to increment sequence number", "err", err)
			}
			shouldFill = true
		}
	default:
		shouldFill = true
	}
	if shouldFill {
		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Network Next",
			Code:                   "next",
			Active:                 true,
			AutomaticSignInDomains: "networknext.com",
		}); err != nil {
			level.Error(logger).Log("msg", "could not add customer to storage", "err", err)
			os.Exit(1)
		}
		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Ghost Army",
			Code:                   "ghost-army",
			Active:                 true,
			AutomaticSignInDomains: "",
		}); err != nil {
			level.Error(logger).Log("msg", "could not add customer to storage", "err", err)
			os.Exit(1)
		}
		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Local",
			Code:                   "local",
			Active:                 true,
			AutomaticSignInDomains: "",
		}); err != nil {
			level.Error(logger).Log("msg", "could not add customer to storage", "err", err)
			os.Exit(1)
		}
		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Valve",
			Code:                   "valve",
			Active:                 true,
			AutomaticSignInDomains: "",
		}); err != nil {
			level.Error(logger).Log("msg", "could not add customer to storage", "err", err)
			os.Exit(1)
		}
		if err := db.AddBuyer(ctx, routing.Buyer{
			ID:                   customerID,
			CompanyCode:          "local",
			Live:                 true,
			PublicKey:            customerPublicKey,
			RouteShader:          routeShader,
			InternalConfig:       internalConfig,
			RoutingRulesSettings: routing.LocalRoutingRulesSettings,
		}); err != nil {
			level.Error(logger).Log("msg", "could not add buyer to storage", "err", err)
			os.Exit(1)
		}
		if err := db.AddBuyer(ctx, routing.Buyer{
			ID:                   0,
			CompanyCode:          "ghost-army",
			Live:                 true,
			PublicKey:            customerPublicKey,
			RouteShader:          routeShader,
			InternalConfig:       internalConfig,
			RoutingRulesSettings: routing.LocalRoutingRulesSettings,
		}); err != nil {
			level.Error(logger).Log("msg", "could not add buyer to storage", "err", err)
			os.Exit(1)
		}
		seller := routing.Seller{
			ID:                        "sellerID",
			CompanyCode:               "local",
			Name:                      "local",
			IngressPriceNibblinsPerGB: 0.1 * 1e9,
			EgressPriceNibblinsPerGB:  0.2 * 1e9,
		}
		valveSeller := routing.Seller{
			ID:                        "valve",
			CompanyCode:               "valve",
			Name:                      "Valve",
			IngressPriceNibblinsPerGB: 0.1 * 1e9,
			EgressPriceNibblinsPerGB:  0.5 * 1e9,
		}
		did := crypto.HashID("local")
		datacenter := routing.Datacenter{
			ID:           did,
			SignedID:     int64(did),
			Name:         "local",
			SupplierName: "usw2-az4",
		}
		if err := db.AddSeller(ctx, seller); err != nil {
			level.Error(logger).Log("msg", "could not add seller to storage", "err", err)
			os.Exit(1)
		}
		if err := db.AddSeller(ctx, valveSeller); err != nil {
			level.Error(logger).Log("msg", "could not add seller to storage", "err", err)
			os.Exit(1)
		}
		if err := db.AddDatacenter(ctx, datacenter); err != nil {
			level.Error(logger).Log("msg", "could not add datacenter to storage", "err", err)
			os.Exit(1)
		}
		if val, ok := os.LookupEnv("LOCAL_RELAYS"); ok {
			numRelays := uint64(10)
			numRelays, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				level.Warn(logger).Log("msg", fmt.Sprintf("LOCAL_RELAYS not valid number, defaulting to 10: %v", err))
			}
			level.Info(logger).Log("msg", fmt.Sprintf("adding %d relays to local firestore", numRelays))
			for i := uint64(0); i < numRelays; i++ {
				addrExternal := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000 + int(i)}
				addrInternal := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000 + int(i)}
				id := crypto.HashID(addrExternal.String())
				if err := db.AddRelay(ctx, routing.Relay{
					Name:           fmt.Sprintf("local.test_relay.%d", i),
					ID:             id,
					SignedID:       int64(id),
					Addr:           addrExternal,
					InternalAddr:   addrInternal,
					PublicKey:      relayPublicKey,
					Seller:         seller,
					Datacenter:     datacenter,
					ManagementAddr: addrExternal.String(),
					SSHUser:        "root",
					SSHPort:        22,
					MaxSessions:    3000,
					MRC:            19700000000000,
					Overage:        26000000000000,
					BWRule:         routing.BWRuleBurst,
					ContractTerm:   12,
					StartDate:      time.Now(),
					EndDate:        time.Now(),
					Type:           routing.BareMetal,
					State:          routing.RelayStateOffline,
					NICSpeedMbps:   1000,
				}); err != nil {
					level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
					os.Exit(1)
				}
				if i%25 == 0 {
					time.Sleep(time.Millisecond * 500)
				}
			}
		} else {
			addr1External := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000}
			addr1Internal := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000}
			rid1 := crypto.HashID(addr1External.String())
			if err := db.AddRelay(ctx, routing.Relay{
				Name:           "local.test_relay.a",
				ID:             rid1,
				SignedID:       int64(rid1),
				Addr:           addr1External,
				InternalAddr:   addr1Internal,
				PublicKey:      relayPublicKey,
				Seller:         seller,
				Datacenter:     datacenter,
				ManagementAddr: "127.0.0.1",
				SSHUser:        "root",
				SSHPort:        22,
				MaxSessions:    3000,
				MRC:            19700000000000,
				Overage:        26000000000000,
				BWRule:         routing.BWRuleBurst,
				ContractTerm:   12,
				StartDate:      time.Now(),
				EndDate:        time.Now(),
				Type:           routing.BareMetal,
			}); err != nil {
				level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
				os.Exit(1)
			}
			addr2External := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10001}
			addr2Internal := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10001}
			rid2 := crypto.HashID(addr2External.String())
			if err := db.AddRelay(ctx, routing.Relay{
				Name:           "local.test_relay.b",
				ID:             rid2,
				SignedID:       int64(rid2),
				Addr:           addr2External,
				InternalAddr:   addr2Internal,
				PublicKey:      relayPublicKey,
				Seller:         seller,
				Datacenter:     datacenter,
				ManagementAddr: "127.0.0.1",
				SSHUser:        "root",
				SSHPort:        22,
				MaxSessions:    3000,
			}); err != nil {
				level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
				os.Exit(1)
			}
			addr3External := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10002}
			addr3Internal := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10002}
			rid3 := crypto.HashID(addr3External.String())
			if err := db.AddRelay(ctx, routing.Relay{
				Name:           "abc.xyz",
				ID:             rid3,
				SignedID:       int64(rid3),
				Addr:           addr3External,
				InternalAddr:   addr3Internal,
				PublicKey:      relayPublicKey,
				Seller:         seller,
				Datacenter:     datacenter,
				ManagementAddr: "127.0.0.1",
				SSHUser:        "root",
				SSHPort:        22,
				MaxSessions:    3000,
			}); err != nil {
				level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
				os.Exit(1)
			}
		}
	}
}
