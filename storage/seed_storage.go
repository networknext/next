package storage

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/routing"
)

func SeedStorage(
	ctx context.Context,
	db Storer,
	relayPublicKey []byte,
	customerID uint64,
	customerPublicKey []byte,
) error {
	routeShader := core.NewRouteShader()
	internalConfig := core.NewInternalConfig()
	internalConfig.ForceNext = true

	shouldFill := false
	switch db := db.(type) {
	case *Firestore:
		_, _, err := db.CheckSequenceNumber(ctx)
		if err != nil {
			return fmt.Errorf("CheckSequenceNumber() err: %w", err)
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
			return fmt.Errorf("AddCustomer() err: %w", err)
		}
		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Ghost Army",
			Code:                   "ghost-army",
			Active:                 true,
			AutomaticSignInDomains: "",
		}); err != nil {
			return fmt.Errorf("AddCustomer() err: %w", err)
		}
		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Local",
			Code:                   "local",
			Active:                 true,
			AutomaticSignInDomains: "",
		}); err != nil {
			return fmt.Errorf("AddCustomer() err: %w", err)
		}
		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Valve",
			Code:                   "valve",
			Active:                 true,
			AutomaticSignInDomains: "",
		}); err != nil {
			return fmt.Errorf("AddCustomer() err: %w", err)
		}
		if err := db.AddBuyer(ctx, routing.Buyer{
			ID:             customerID,
			CompanyCode:    "local",
			Live:           true,
			PublicKey:      customerPublicKey,
			RouteShader:    routeShader,
			InternalConfig: internalConfig,
		}); err != nil {
			return fmt.Errorf("AddBuyer() err: %w", err)
		}
		if err := db.AddBuyer(ctx, routing.Buyer{
			ID:             0,
			CompanyCode:    "ghost-army",
			Live:           true,
			PublicKey:      customerPublicKey,
			RouteShader:    routeShader,
			InternalConfig: internalConfig,
		}); err != nil {
			return fmt.Errorf("AddBuyer() err: %w", err)
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
			return fmt.Errorf("AddSeller() err: %w", err)
		}
		if err := db.AddSeller(ctx, valveSeller); err != nil {
			return fmt.Errorf("AddSeller() err: %w", err)
		}
		if err := db.AddDatacenter(ctx, datacenter); err != nil {
			return fmt.Errorf("AddDatacenter() err: %w", err)
		}
		if val, ok := os.LookupEnv("LOCAL_RELAYS"); ok {
			numRelays := uint64(10)
			numRelays, err := strconv.ParseUint(val, 10, 64)
			if err != nil {
				return fmt.Errorf("LOCAL_RELAYS ParseUint() err: %w", err)
			}
			for i := uint64(0); i < numRelays; i++ {
				addr := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000 + int(i)}
				id := crypto.HashID(addr.String())
				if err := db.AddRelay(ctx, routing.Relay{
					Name:           fmt.Sprintf("local.test_relay.%d", i),
					ID:             id,
					SignedID:       int64(id),
					Addr:           addr,
					PublicKey:      relayPublicKey,
					Seller:         seller,
					Datacenter:     datacenter,
					ManagementAddr: addr.String(),
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
					return fmt.Errorf("AddRelay() err: %w", err)
				}
				if i%25 == 0 {
					time.Sleep(time.Millisecond * 500)
				}
			}
		} else {
			addr1 := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000}
			rid1 := crypto.HashID(addr1.String())
			if err := db.AddRelay(ctx, routing.Relay{
				Name:           "local.test_relay.a",
				ID:             rid1,
				SignedID:       int64(rid1),
				Addr:           addr1,
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
				return fmt.Errorf("AddRelay() err: %w", err)
			}
			addr2 := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10001}
			rid2 := crypto.HashID(addr2.String())
			if err := db.AddRelay(ctx, routing.Relay{
				Name:           "local.test_relay.b",
				ID:             rid2,
				SignedID:       int64(rid2),
				Addr:           addr2,
				PublicKey:      relayPublicKey,
				Seller:         seller,
				Datacenter:     datacenter,
				ManagementAddr: "127.0.0.1",
				SSHUser:        "root",
				SSHPort:        22,
				MaxSessions:    3000,
			}); err != nil {
				return fmt.Errorf("AddRelay() err: %w", err)
			}
			addr3 := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10002}
			rid3 := crypto.HashID(addr3.String())
			if err := db.AddRelay(ctx, routing.Relay{
				Name:           "abc.xyz",
				ID:             rid3,
				SignedID:       int64(rid3),
				Addr:           addr3,
				PublicKey:      relayPublicKey,
				Seller:         seller,
				Datacenter:     datacenter,
				ManagementAddr: "127.0.0.1",
				SSHUser:        "root",
				SSHPort:        22,
				MaxSessions:    3000,
			}); err != nil {
				return fmt.Errorf("AddRelay() err: %w", err)
			}
		}
	}
	return nil
}
