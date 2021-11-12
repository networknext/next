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

		if err := db.AddAnalyticsDashboardCategory(ctx, "General", false, false); err != nil {
			return fmt.Errorf("AddAnalyticsDashboardCategory() err: %w", err)
		}

		if err := db.AddAnalyticsDashboardCategory(ctx, "Regional", false, true); err != nil {
			return fmt.Errorf("AddAnalyticsDashboardCategory() err: %w", err)
		}

		if err := db.AddAnalyticsDashboardCategory(ctx, "Platform", false, true); err != nil {
			return fmt.Errorf("AddAnalyticsDashboardCategory() err: %w", err)
		}

		if err := db.AddAnalyticsDashboardCategory(ctx, "Discovery", false, true); err != nil {
			return fmt.Errorf("AddAnalyticsDashboardCategory() err: %w", err)
		}

		if err := db.AddAnalyticsDashboardCategory(ctx, "System", true, false); err != nil {
			return fmt.Errorf("AddAnalyticsDashboardCategory() err: %w", err)
		}

		generalCategory, err := db.GetAnalyticsDashboardCategoryByLabel(ctx, "General")
		if err != nil {
			return fmt.Errorf("GetAnalyticsDashboardCategoryByLabel() err: %w", err)
		}

		regionalCategory, err := db.GetAnalyticsDashboardCategoryByLabel(ctx, "Regional")
		if err != nil {
			return fmt.Errorf("GetAnalyticsDashboardCategoryByLabel() err: %w", err)
		}

		platformCategory, err := db.GetAnalyticsDashboardCategoryByLabel(ctx, "Platform")
		if err != nil {
			return fmt.Errorf("GetAnalyticsDashboardCategoryByLabel() err: %w", err)
		}

		discoveryCategory, err := db.GetAnalyticsDashboardCategoryByLabel(ctx, "Discovery")
		if err != nil {
			return fmt.Errorf("GetAnalyticsDashboardCategoryByLabel() err: %w", err)
		}

		systemCategory, err := db.GetAnalyticsDashboardCategoryByLabel(ctx, "System")
		if err != nil {
			return fmt.Errorf("GetAnalyticsDashboardCategoryByLabel() err: %w", err)
		}

		// Add customers
		// fmt.Println("Adding customers")
		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Network Next",
			Code:                   "next",
			AutomaticSignInDomains: "networknext.com",
		}); err != nil {
			return fmt.Errorf("AddCustomer() err: %w", err)
		}

		nextCustomer, err := db.Customer(ctx, "next")
		if err == nil {
			if err := db.AddAnalyticsDashboard(ctx, "Relay Stats", 11, false, nextCustomer.DatabaseID, systemCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
		}

		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Happy Path",
			Code:                   "happypath",
			AutomaticSignInDomains: "happypath.com",
		}); err != nil {
			return fmt.Errorf("AddCustomer() err: %w", err)
		}

		happyPathCustomer, err := db.Customer(ctx, "happypath")
		if err == nil {
			if err := db.AddAnalyticsDashboard(ctx, "General Analytics", 14, false, happyPathCustomer.DatabaseID, generalCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
		}

		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Ghost Army",
			Code:                   "ghost-army",
			AutomaticSignInDomains: "",
		}); err != nil {
			return fmt.Errorf("AddCustomer() err: %w", err)
		}

		ghostCustomer, err := db.Customer(ctx, "ghost-army")
		if err == nil {
			if err := db.AddAnalyticsDashboard(ctx, "General Analytics", 14, false, ghostCustomer.DatabaseID, generalCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
			if err := db.AddAnalyticsDashboard(ctx, "Regional Analytics", 18, false, ghostCustomer.DatabaseID, regionalCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
			if err := db.AddAnalyticsDashboard(ctx, "Platform Analytics", 12, false, ghostCustomer.DatabaseID, platformCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
			if err := db.AddAnalyticsDashboard(ctx, "Some Discovery!", 11, true, ghostCustomer.DatabaseID, discoveryCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
			if err := db.AddAnalyticsDashboard(ctx, "Another Discovery!", 11, true, ghostCustomer.DatabaseID, discoveryCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
		}

		if err := db.AddCustomer(ctx, routing.Customer{
			Name:                   "Local",
			Code:                   "local",
			AutomaticSignInDomains: "",
		}); err != nil {
			return fmt.Errorf("AddCustomer() err: %w", err)
		}

		localCustomer, err := db.Customer(ctx, "local")
		if err == nil {
			if err := db.AddAnalyticsDashboard(ctx, "General Analytics", 14, false, localCustomer.DatabaseID, generalCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
			if err := db.AddAnalyticsDashboard(ctx, "Regional Analytics", 18, false, localCustomer.DatabaseID, regionalCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
			if err := db.AddAnalyticsDashboard(ctx, "Platform Analytics", 12, false, localCustomer.DatabaseID, platformCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
			if err := db.AddAnalyticsDashboard(ctx, "Some Discovery!", 11, true, localCustomer.DatabaseID, discoveryCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
			if err := db.AddAnalyticsDashboard(ctx, "Another Discovery!", 11, true, localCustomer.DatabaseID, discoveryCategory.ID); err != nil {
				return fmt.Errorf("AddAnalyticsDashboard() err: %w", err)
			}
		}

		// retrieve entities so we can get the database-assigned keys
		localCust, err := db.Customer(ctx, "local")
		if err != nil {
			return fmt.Errorf("Error getting local customer: %v", err)
		}

		hpCust, err := db.Customer(ctx, "happypath")
		if err != nil {
			return fmt.Errorf("Error getting happypath customer: %v", err)
		}

		ghostCust, err := db.Customer(ctx, "ghost-army")
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
			Analytics:   true,
			Billing:     true,
			Debug:       true,
			Trial:       true,
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
			Analytics:   false,
			Billing:     false,
			Debug:       false,
			Trial:       true,
			PublicKey:   publicKey,
			CustomerID:  ghostCust.DatabaseID,
		}); err != nil {
			return fmt.Errorf("AddBuyer() err: %w", err)
		}

		localBuyer, err := db.Buyer(ctx, customerID)
		if err != nil {
			return fmt.Errorf("Error getting local buyer: %v", err)
		}

		ghostBuyer, err := db.Buyer(ctx, internalBuyerIDGhost)
		if err != nil {
			return fmt.Errorf("Error getting ghost army buyer: %v", err)
		}

		// fmt.Println("Adding sellers")
		localSeller := routing.Seller{
			ID:                       localCust.Code,
			ShortName:                "local",
			CompanyCode:              "local",
			Secret:                   false,
			Name:                     localCust.Name,
			EgressPriceNibblinsPerGB: 0.2 * 1e11,
			CustomerID:               localCust.DatabaseID,
		}

		ghostSeller := routing.Seller{
			ID:                       ghostCust.Code,
			ShortName:                "ghost-army",
			CompanyCode:              "ghost-army",
			Secret:                   false,
			Name:                     ghostCust.Name,
			EgressPriceNibblinsPerGB: 0.4 * 1e11,
			CustomerID:               ghostCust.DatabaseID,
		}

		hpSeller := routing.Seller{
			ID:                       hpCust.Code,
			ShortName:                hpCust.Code,
			CompanyCode:              hpCust.Code,
			Secret:                   false,
			Name:                     hpCust.Name,
			EgressPriceNibblinsPerGB: 0.4 * 1e11,
			CustomerID:               hpCust.DatabaseID,
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

		localSeller, err = db.Seller(ctx, "local")
		if err != nil {
			return fmt.Errorf("Error getting local seller: %v", err)
		}

		ghostSeller, err = db.Seller(ctx, "ghost-army")
		if err != nil {
			return fmt.Errorf("Error getting ghost seller: %v", err)
		}

		hpSeller, err = db.Seller(ctx, "happypath")
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

		localDatacenter, err = db.Datacenter(ctx, localDCID)
		if err != nil {
			return fmt.Errorf("Error getting local datacenter: %v", err)
		}

		ghostDatacenter, err = db.Datacenter(ctx, ghostDCID)
		if err != nil {
			return fmt.Errorf("Error getting local datacenter: %v", err)
		}

		// add datacenter maps
		// fmt.Println("Adding datacenter_maps")
		localDcMap := routing.DatacenterMap{
			BuyerID:      localBuyer.ID,
			DatacenterID: localDatacenter.ID,
		}

		err = db.AddDatacenterMap(ctx, localDcMap)
		if err != nil {
			return fmt.Errorf("Error creating local datacenter map: %v", err)
		}

		ghostDcMap := routing.DatacenterMap{
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
				EgressPriceOverride: 0,
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
				Version:             "2.0.6",
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
				EgressPriceOverride: 0,
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
				Version:             "2.0.6",
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
			RouteSelectThreshold:           5,
			RouteSwitchThreshold:           10,
			MaxLatencyTradeOff:             10,
			RTTVeto_Default:                -10,
			RTTVeto_PacketLoss:             -20,
			RTTVeto_Multipath:              -20,
			MultipathOverloadThreshold:     500,
			TryBeforeYouBuy:                false,
			ForceNext:                      true,
			LargeCustomer:                  false,
			Uncommitted:                    false,
			MaxRTT:                         300,
			HighFrequencyPings:             true,
			RouteDiversity:                 0,
			MultipathThreshold:             35,
			EnableVanityMetrics:            true,
			ReducePacketLossMinSliceNumber: 10,
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
			PacketLossSustained:       float32(100),
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
			PacketLossSustained:       float32(100),
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

		// set creation time to 1.5 hours ago to avoid cooldown ticker in UI
		now := time.Now()
		duration, _ := time.ParseDuration("-1.5h")
		then := now.Add(duration)

		metaData := routing.DatabaseBinFileMetaData{
			DatabaseBinFileAuthor:       "arthur@networknext.com",
			DatabaseBinFileCreationTime: then,
		}

		err = db.UpdateDatabaseBinFileMetaData(context.Background(), metaData)
		if err != nil {
			return fmt.Errorf("AdminBinFileHandler() error writing bin file metadata to db: %v", err)
		}

	}

	return nil
}

// Seeds the SQLite storer for the staging environment
func SeedSQLStorageStaging(
	ctx context.Context,
	db Storer,
	database *routing.DatabaseBinWrapper,
) error {
	// When using SQLite it is ok to "seed" each version of the storer
	// and let them sync up later on. When using a local PostgreSQL server
	// we can only seed storage once, externally (via SQL file).
	// TODO: setup "only seed once" checking for PostgreSQL
	var err error

	pgsql, err := envvar.GetBool("FEATURE_POSTGRESQL", false)
	if err != nil {
		return fmt.Errorf("could not parse FEATURE_POSTGRESQL boolean: %v", err)
	}

	// only seed if we're using sqlite3
	if pgsql {
		return nil
	}

	// Add customers manually in order by customerID

	if err = db.AddCustomer(ctx, routing.Customer{
		Name:                   "Ghost Army",
		Code:                   "ghost-army",
		AutomaticSignInDomains: "ghost_army.com.net.gov",
		DatabaseID:             1,
	}); err != nil {
		return fmt.Errorf("AddCustomer() ghost army err: %v", err)
	}

	if err = db.AddCustomer(ctx, routing.Customer{
		Name:                   "staging seller",
		Code:                   "stagingseller",
		AutomaticSignInDomains: "",
		DatabaseID:             2,
	}); err != nil {
		return fmt.Errorf("AddCustomer() staging seller err: %v", err)
	}

	if err = db.AddCustomer(ctx, routing.Customer{
		Name:                   "Network Next",
		Code:                   "next",
		AutomaticSignInDomains: "networknext.com",
		DatabaseID:             3,
	}); err != nil {
		return fmt.Errorf("AddCustomer() next err: %v", err)
	}

	// Add buyers in order
	nextBuyerID := uint64(13672574147039585173)
	stagingSellerBuyerID := uint64(13053258624167246632)
	ghostArmyBuyerID := uint64(0)

	if err = db.AddBuyer(ctx, database.BuyerMap[nextBuyerID]); err != nil {
		return fmt.Errorf("AddBuyer() next err: %v", err)
	}
	if err = db.AddBuyer(ctx, database.BuyerMap[stagingSellerBuyerID]); err != nil {
		return fmt.Errorf("AddBuyer() staging seller err: %v", err)
	}
	if err = db.AddBuyer(ctx, database.BuyerMap[ghostArmyBuyerID]); err != nil {
		return fmt.Errorf("AddBuyer() ghost army err: %v", err)
	}

	// Add sellers
	for _, seller := range database.SellerMap {
		if err = db.AddSeller(ctx, seller); err != nil {
			return fmt.Errorf("AddSeller() err: %v", err)
		}
	}

	// Add buyer internal configs and route shaders
	for buyerID, buyer := range database.BuyerMap {
		if err = db.AddInternalConfig(ctx, buyer.InternalConfig, buyerID); err != nil {
			return fmt.Errorf("AddInternalConfig() err: %v", err)
		}
		if err = db.AddRouteShader(ctx, buyer.RouteShader, buyerID); err != nil {
			return fmt.Errorf("AddRouteShader() err: %v", err)
		}
	}

	// Add datacenters
	for _, datacenter := range database.DatacenterMap {
		if err = db.AddDatacenter(ctx, datacenter); err != nil {
			return fmt.Errorf("AddDatacenter() err: %v", err)
		}
	}

	// Add datacenter maps
	for buyerID := range database.BuyerMap {
		if dcMaps, ok := database.DatacenterMaps[buyerID]; ok {
			for _, dcMap := range dcMaps {
				if err = db.AddDatacenterMap(ctx, dcMap); err != nil {
					return fmt.Errorf("AddDatacenterMap() err: %v", err)
				}
			}
		}
	}

	// Add relays
	for _, relay := range database.Relays {
		if err = db.AddRelay(ctx, relay); err != nil {
			return fmt.Errorf("AddRelay() err: %v", err)
		}
	}

	return nil
}
