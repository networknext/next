package local

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

func SeedFirestore(logger log.Logger, ctx context.Context, db storage.Storer, relayPublicKey []byte, customerID uint64, customerPublicKey []byte) {
	if err := db.AddBuyer(ctx, routing.Buyer{
		ID:                   customerID,
		Name:                 "local",
		PublicKey:            customerPublicKey,
		RoutingRulesSettings: routing.LocalRoutingRulesSettings,
	}); err != nil {
		level.Error(logger).Log("msg", "could not add buyer to storage", "err", err)
		os.Exit(1)
	}

	if err := db.AddBuyer(ctx, routing.Buyer{
		ID:                   0,
		Name:                 "Ghost Army",
		PublicKey:            customerPublicKey,
		RoutingRulesSettings: routing.LocalRoutingRulesSettings,
	}); err != nil {
		level.Error(logger).Log("msg", "could not add buyer to storage", "err", err)
		os.Exit(1)
	}

	seller := routing.Seller{
		ID:                        "sellerID",
		Name:                      "local",
		IngressPriceNibblinsPerGB: 0.1 * 1e9,
		EgressPriceNibblinsPerGB:  0.2 * 1e9,
	}

	valveSeller := routing.Seller{
		ID:                        "valve",
		Name:                      "valve",
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
			fmt.Printf("LOCAL_PORTAL_RELAYS not valid number, defaulting to 10: %v\n", err)
		}

		fmt.Printf("adding %d relays to local firestore\n", numRelays)

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
				MRC:            19700000000000,
				Overage:        26000000000000,
				BWRule:         routing.BWRuleBurst,
				ContractTerm:   12,
				StartDate:      time.Now(),
				EndDate:        time.Now(),
				Type:           routing.BareMetal,
				State:          routing.RelayStateOffline,
			}); err != nil {
				level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
				os.Exit(1)
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
		}); err != nil {
			level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
			os.Exit(1)
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
		}); err != nil {
			level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
			os.Exit(1)
		}

		switch db := db.(type) {
		case *storage.Firestore:
			db.SetSequenceNumber(ctx, 0)
		}
	}
}
