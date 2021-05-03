package storage

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/routing"
)

func SeedSQLStorage(
	ctx context.Context,
	db Storer,
	relayPublicKey []byte,
	customerID uint64,
	customerPublicKey []byte,
) error {
	// When using SQLite it is ok to "seed" each version of the storer
	// and let them sync up later on. When using a local PostgreSQL server
	// we can only seed storage once, externally (via SQL file).
	// TODO: setup "only seed once" checking for PostgreSQL

	pgsql, err := envvar.GetBool("FEATURE_POSTGRESQL", false)
	if err != nil {
		return fmt.Errorf("could not parse FEATURE_POSTGRESQL boolean: %v", err)
	}

	// only seed if we're using sqlite3
	if !pgsql {

		// Add customers
		// fmt.Println("Adding customers")
		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Network Next",
			Code:                   "next",
			AutomaticSignInDomains: "networknext.com",
		}); err != nil {
			return fmt.Errorf("AddCustomer() err: %w", err)
		}

		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Happy Path",
			Code:                   "happypath",
			AutomaticSignInDomains: "happypath.com",
		}); err != nil {
			return fmt.Errorf("AddCustomer() err: %w", err)
		}

		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Ghost Army",
			Code:                   "ghost-army",
			AutomaticSignInDomains: "",
		}); err != nil {
			return fmt.Errorf("AddCustomer() err: %w", err)
		}

		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Local",
			Code:                   "local",
			AutomaticSignInDomains: "",
		}); err != nil {
			return fmt.Errorf("AddCustomer() err: %w", err)
		}

		// retrieve entities so we can get the database-assigned keys
		localCust, err := db.Customer("local")
		if err != nil {
			return fmt.Errorf("Error getting local customer: %v", err)
		}

		hpCust, err := db.Customer("happypath")
		if err != nil {
			return fmt.Errorf("Error getting happypath customer: %v", err)
		}

		ghostCust, err := db.Customer("ghost-army")
		if err != nil {
			return fmt.Errorf("Error getting ghost customer: %v", err)
		}

		// Add buyers
		// fmt.Println("Adding buyers")
		if err := db.AddBuyer(ctx, routing.Buyer{
			ID:          customerID,
			ShortName:   "local",
			CompanyCode: localCust.Code,
			Live:        true,
			PublicKey:   customerPublicKey,
			CustomerID:  localCust.DatabaseID,
			Debug:       true,
		}); err != nil {
			return fmt.Errorf("AddBuyer() err: %w", err)
		}

		publicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		if err != nil {
			return fmt.Errorf("Error generating buyer public key: %v", err)
		}
		internalBuyerIDGhost := binary.LittleEndian.Uint64(publicKey[:8])

		if err := db.AddBuyer(ctx, routing.Buyer{
			ID:          internalBuyerIDGhost,
			ShortName:   "ghost-army",
			CompanyCode: ghostCust.Code,
			Live:        true,
			PublicKey:   publicKey,
			CustomerID:  ghostCust.DatabaseID,
		}); err != nil {
			return fmt.Errorf("AddBuyer() err: %w", err)
		}

		localBuyer, err := db.Buyer(customerID)
		if err != nil {
			return fmt.Errorf("Error getting local buyer: %v", err)
		}

		ghostBuyer, err := db.Buyer(internalBuyerIDGhost)
		if err != nil {
			return fmt.Errorf("Error getting local buyer: %v", err)
		}

		// fmt.Println("Adding sellers")
		localSeller := routing.Seller{
			ID:                        localCust.Code,
			ShortName:                 "local",
			CompanyCode:               "local",
			Secret:                    false,
			Name:                      localCust.Name,
			IngressPriceNibblinsPerGB: 0.1 * 1e11,
			EgressPriceNibblinsPerGB:  0.2 * 1e11,
			CustomerID:                localCust.DatabaseID,
		}

		ghostSeller := routing.Seller{
			ID:                        ghostCust.Code,
			ShortName:                 "ghost",
			CompanyCode:               "ghost-army",
			Secret:                    false,
			Name:                      ghostCust.Name,
			IngressPriceNibblinsPerGB: 0.3 * 1e11,
			EgressPriceNibblinsPerGB:  0.4 * 1e11,
			CustomerID:                ghostCust.DatabaseID,
		}

		hpSeller := routing.Seller{
			ID:                        hpCust.Code,
			ShortName:                 hpCust.Code,
			CompanyCode:               hpCust.Code,
			Secret:                    false,
			Name:                      hpCust.Name,
			IngressPriceNibblinsPerGB: 0.3 * 1e11,
			EgressPriceNibblinsPerGB:  0.4 * 1e11,
			CustomerID:                hpCust.DatabaseID,
		}

		if err := db.AddSeller(ctx, localSeller); err != nil {
			return fmt.Errorf("AddSeller() err adding localSeller: %w", err)
		}

		if err := db.AddSeller(ctx, ghostSeller); err != nil {
			return fmt.Errorf("AddSeller() err adding ghostSeller: %w", err)
		}

		if err := db.AddSeller(ctx, hpSeller); err != nil {
			return fmt.Errorf("AddSeller() err adding hpSeller: %w", err)
		}

		localSeller, err = db.Seller("local")
		if err != nil {
			return fmt.Errorf("Error getting local seller: %v", err)
		}

		ghostSeller, err = db.Seller("ghost-army")
		if err != nil {
			return fmt.Errorf("Error getting ghost seller: %v", err)
		}

		hpSeller, err = db.Seller("happypath")
		if err != nil {
			return fmt.Errorf("Error getting happypath seller: %v", err)
		}

		var localDCID uint64
		for i := uint64(0); i < 10; i++ {
			dcName := "local." + fmt.Sprintf("%d", i)
			localDCID = crypto.HashID(dcName)
			localDatacenter2 := routing.Datacenter{
				ID:   localDCID,
				Name: dcName,
				Location: routing.Location{
					Latitude:  0,
					Longitude: 0,
				},
				SellerID: hpSeller.DatabaseID,
			}
			if err := db.AddDatacenter(ctx, localDatacenter2); err != nil {
				return fmt.Errorf("AddDatacenter() error adding local datacenter: %w", err)
			}
		}

		// fmt.Println("Adding datacenters")
		// req for happy path
		localDCID = crypto.HashID("local")
		localDatacenter := routing.Datacenter{
			ID:   localDCID,
			Name: "local",
			Location: routing.Location{
				Latitude:  0,
				Longitude: 0,
			},
			SellerID: localSeller.DatabaseID,
		}
		if err := db.AddDatacenter(ctx, localDatacenter); err != nil {
			return fmt.Errorf("AddDatacenter() error adding local datacenter: %w", err)
		}

		ghostDCID := crypto.HashID("ghost-army.local.name")
		ghostDatacenter := routing.Datacenter{
			ID:   ghostDCID,
			Name: "ghost-army.local.name",
			Location: routing.Location{
				Latitude:  0,
				Longitude: 0,
			},
			SellerID: ghostSeller.DatabaseID,
		}
		if err := db.AddDatacenter(ctx, ghostDatacenter); err != nil {
			return fmt.Errorf("AddDatacenter() error adding ghost datacenter: %w", err)
		}

		localDatacenter, err = db.Datacenter(localDCID)
		if err != nil {
			return fmt.Errorf("Error getting local datacenter: %v", err)
		}

		ghostDatacenter, err = db.Datacenter(ghostDCID)
		if err != nil {
			return fmt.Errorf("Error getting local datacenter: %v", err)
		}

		// add datacenter maps
		// fmt.Println("Adding datacenter_maps")
		localDcMap := routing.DatacenterMap{
			Alias:        "local",
			BuyerID:      localBuyer.ID,
			DatacenterID: localDatacenter.ID,
		}

		err = db.AddDatacenterMap(ctx, localDcMap)
		if err != nil {
			return fmt.Errorf("Error creating local datacenter map: %v", err)
		}

		ghostDcMap := routing.DatacenterMap{
			Alias:        "ghost-army.map",
			BuyerID:      ghostBuyer.ID,
			DatacenterID: ghostDatacenter.ID,
		}

		err = db.AddDatacenterMap(ctx, ghostDcMap)
		if err != nil {
			return fmt.Errorf("Error creating local datacenter map: %v", err)
		}

		// fmt.Println("Adding relays")
		// Add the number of relays provided by LOCAL_RELAYS for each datacenter
		numRelays := uint64(10)
		if val, ok := os.LookupEnv("LOCAL_RELAYS"); ok {
			numRelays, err = strconv.ParseUint(val, 10, 64)
			if err != nil {
				return fmt.Errorf("LOCAL_RELAYS ParseUint() err: %w", err)
			}
		}

		for i := uint64(0); i < numRelays; i++ {
			// fmt.Printf("\tSeedSQLStorage adding relay %d\n", i)
			updateKey := make([]byte, crypto.KeySize)
			_, err = rand.Read(updateKey)
			if err != nil {
				return fmt.Errorf("Error generating relay update key: %v", err)
			}

			addr := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000 + int(i)}
			rid := crypto.HashID(addr.String())

			internalAddr := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000 + int(i)}

			if err := db.AddRelay(ctx, routing.Relay{
				ID:                  rid,
				Name:                "local." + fmt.Sprintf("%d", i),
				Addr:                addr,
				InternalAddr:        internalAddr,
				ManagementAddr:      "1.2.3.4" + fmt.Sprintf("%d", i),
				SSHPort:             22,
				SSHUser:             "root",
				MaxSessions:         uint32(1000 + i),
				PublicKey:           relayPublicKey,
				Datacenter:          localDatacenter,
				MRC:                 19700000000000,
				Overage:             26000000000000,
				BWRule:              routing.BWRuleBurst,
				ContractTerm:        12,
				StartDate:           time.Now(),
				EndDate:             time.Now(),
				Type:                routing.BareMetal,
				State:               routing.RelayStateEnabled,
				IncludedBandwidthGB: 10000,
				NICSpeedMbps:        1000,
				Notes:               "I am relay local." + fmt.Sprintf("%d", i) + " - hear me roar!",
			}); err != nil {
				return fmt.Errorf("AddRelay() error adding local relay: %w", err)
			}

			// ghost
			publicKey = make([]byte, crypto.KeySize)
			_, err = rand.Read(publicKey)
			if err != nil {
				return fmt.Errorf("Error generating ghost relay public key: %v", err)
			}

			updateKey = make([]byte, crypto.KeySize)
			_, err = rand.Read(updateKey)
			if err != nil {
				return fmt.Errorf("Error generating ghost  relay update key: %v", err)
			}

			addr = net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 10000 + int(i)}
			rid = crypto.HashID(addr.String())

			internalAddr = net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 10000 + int(i)}

			// set ghost-army relays to random states
			var ghostRelayState routing.RelayState
			rand.Seed(time.Now().UnixNano())
			state := rand.Int63n(6)
			ghostRelayState, _ = routing.GetRelayStateSQL(state)

			if err := db.AddRelay(ctx, routing.Relay{
				ID:                  rid,
				Name:                "ghost-army.local.1" + fmt.Sprintf("%d", i),
				Addr:                addr,
				InternalAddr:        internalAddr,
				ManagementAddr:      "4.3.2.1" + fmt.Sprintf("%d", i),
				SSHPort:             22,
				SSHUser:             "root",
				MaxSessions:         uint32(1000 + i),
				PublicKey:           publicKey,
				Datacenter:          ghostDatacenter,
				MRC:                 19700000000000,
				Overage:             26000000000000,
				BWRule:              routing.BWRuleBurst,
				ContractTerm:        12,
				StartDate:           time.Now(),
				EndDate:             time.Now(),
				Type:                routing.BareMetal,
				State:               ghostRelayState,
				IncludedBandwidthGB: 10000,
				NICSpeedMbps:        1000,
				Notes:               "I am relay ghost-army.local.1" + fmt.Sprintf("%d", i) + " - hear me roar!",
			}); err != nil {
				return fmt.Errorf("AddRelay() error adding ghost relay: %w", err)
			}

			if i%25 == 0 {
				time.Sleep(time.Millisecond * 500)
			}
		}

		// add InternalConfigs, RouteShaders and BannedUsers

		// fmt.Printf("localBuyer ID: %016x\n", localBuyer.ID)
		// fmt.Printf("ghostBuyer ID: %016x\n", ghostBuyer.ID)

		internalConfig := core.InternalConfig{
			RouteSelectThreshold:       5,
			RouteSwitchThreshold:       10,
			MaxLatencyTradeOff:         10,
			RTTVeto_Default:            -10,
			RTTVeto_PacketLoss:         -20,
			RTTVeto_Multipath:          -20,
			MultipathOverloadThreshold: 500,
			TryBeforeYouBuy:            false,
			ForceNext:                  true,
			LargeCustomer:              false,
			Uncommitted:                false,
			MaxRTT:                     300,
			HighFrequencyPings:         true,
			RouteDiversity:             0,
			MultipathThreshold:         35,
			EnableVanityMetrics:        true,
		}

		err = db.AddInternalConfig(ctx, internalConfig, localBuyer.ID)
		if err != nil {
			return fmt.Errorf("Error adding InternalConfig for local buyer: %v", err)
		}

		err = db.AddInternalConfig(ctx, internalConfig, ghostBuyer.ID)
		if err != nil {
			return fmt.Errorf("Error adding InternalConfig for local buyer: %v", err)
		}

		localRouteShader := core.RouteShader{
			ABTest:                    false,
			AcceptableLatency:         int32(25),
			AcceptablePacketLoss:      float32(0),
			BandwidthEnvelopeDownKbps: int32(1200),
			BandwidthEnvelopeUpKbps:   int32(500),
			DisableNetworkNext:        false,
			LatencyThreshold:          int32(0),
			Multipath:                 true,
			ProMode:                   false,
			ReduceLatency:             true,
			ReducePacketLoss:          true,
			SelectionPercent:          int(100),
		}

		gaRouteShader := core.RouteShader{
			ABTest:                    false,
			AcceptableLatency:         int32(25),
			AcceptablePacketLoss:      float32(1),
			BandwidthEnvelopeDownKbps: int32(1200),
			BandwidthEnvelopeUpKbps:   int32(500),
			DisableNetworkNext:        false,
			LatencyThreshold:          int32(5),
			Multipath:                 false,
			ProMode:                   false,
			ReduceLatency:             true,
			ReducePacketLoss:          true,
			SelectionPercent:          int(100),
		}

		err = db.AddRouteShader(ctx, localRouteShader, localBuyer.ID)
		if err != nil {
			return fmt.Errorf("Error adding RouteShader for local buyer: %v", err)
		}

		err = db.AddRouteShader(ctx, gaRouteShader, ghostBuyer.ID)
		if err != nil {
			return fmt.Errorf("Error adding RouteShader for ghost army buyer: %v", err)
		}

		userID1 := rand.Uint64()
		userID2 := rand.Uint64()
		userID3 := rand.Uint64()

		err = db.AddBannedUser(ctx, localBuyer.ID, userID1)
		if err != nil {
			return fmt.Errorf("Error adding BannedUser for local buyer: %v", err)
		}

		err = db.AddBannedUser(ctx, localBuyer.ID, userID2)
		if err != nil {
			return fmt.Errorf("Error adding BannedUser for local buyer: %v", err)
		}

		err = db.AddBannedUser(ctx, localBuyer.ID, userID3)
		if err != nil {
			return fmt.Errorf("Error adding BannedUser for local buyer: %v", err)
		}

		err = db.AddBannedUser(ctx, ghostBuyer.ID, userID1)
		if err != nil {
			return fmt.Errorf("Error adding BannedUser for local buyer: %v", err)
		}

		err = db.AddBannedUser(ctx, ghostBuyer.ID, userID2)
		if err != nil {
			return fmt.Errorf("Error adding BannedUser for local buyer: %v", err)
		}

		err = db.AddBannedUser(ctx, ghostBuyer.ID, userID3)
		if err != nil {
			return fmt.Errorf("Error adding BannedUser for local buyer: %v", err)
		}

	}

	return nil
}
