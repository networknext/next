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
	"github.com/networknext/backend/routing"
)

func SeedSQLStorage(
	ctx context.Context,
	db Storer,
	// relayPublicKey []byte,
	// customerID uint64,
	// customerPublicKey []byte,
) error {
	fmt.Println("SeedSQLStorage()")
	fmt.Println("getting route shader")
	routeShader := core.NewRouteShader()
	fmt.Println("getting internal config")
	internalConfig := core.NewInternalConfig()
	internalConfig.ForceNext = true

	// Add customers
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

	// retrieve entities so we can get the database-assigned keys
	localCust, err := db.Customer("local")
	if err != nil {
		return fmt.Errorf("Error getting local customer: %v", err)
	}

	ghostCust, err := db.Customer("ghost-army")
	if err != nil {
		return fmt.Errorf("Error getting ghost customer: %v", err)
	}

	// Add buyers
	publicKey := make([]byte, crypto.KeySize)
	_, err = rand.Read(publicKey)
	if err != nil {
		return fmt.Errorf("Error generating buyer public key: %v", err)
	}
	internalBuyerIDLocal := binary.LittleEndian.Uint64(publicKey[:8])

	if err := db.AddBuyer(ctx, routing.Buyer{
		ID:             internalBuyerIDLocal,
		ShortName:      "local",
		CompanyCode:    localCust.Code,
		Live:           true,
		PublicKey:      publicKey,
		RouteShader:    routeShader,
		InternalConfig: internalConfig,
		CustomerID:     localCust.DatabaseID,
	}); err != nil {
		return fmt.Errorf("AddBuyer() err: %w", err)
	}

	publicKey = make([]byte, crypto.KeySize)
	_, err = rand.Read(publicKey)
	if err != nil {
		return fmt.Errorf("Error generating buyer public key: %v", err)
	}
	internalBuyerIDGhost := binary.LittleEndian.Uint64(publicKey[:8])

	if err := db.AddBuyer(ctx, routing.Buyer{
		ID:             internalBuyerIDGhost,
		ShortName:      "ghost-army",
		CompanyCode:    ghostCust.Code,
		Live:           true,
		PublicKey:      publicKey,
		RouteShader:    routeShader,
		InternalConfig: internalConfig,
		CustomerID:     ghostCust.DatabaseID,
	}); err != nil {
		return fmt.Errorf("AddBuyer() err: %w", err)
	}

	localBuyer, err := db.Buyer(internalBuyerIDLocal)
	if err != nil {
		return fmt.Errorf("Error getting local buyer: %v", err)
	}

	ghostBuyer, err := db.Buyer(internalBuyerIDGhost)
	if err != nil {
		return fmt.Errorf("Error getting local buyer: %v", err)
	}

	localSeller := routing.Seller{
		ID:                        localCust.Code,
		ShortName:                 "local",
		Name:                      localCust.Name,
		IngressPriceNibblinsPerGB: 0.1 * 1e11,
		EgressPriceNibblinsPerGB:  0.2 * 1e11,
		CustomerID:                localCust.DatabaseID,
	}

	ghostSeller := routing.Seller{
		ID:                        ghostCust.Code,
		ShortName:                 "ghost",
		Name:                      ghostCust.Name,
		IngressPriceNibblinsPerGB: 0.3 * 1e11,
		EgressPriceNibblinsPerGB:  0.4 * 1e11,
		CustomerID:                ghostCust.DatabaseID,
	}

	if err := db.AddSeller(ctx, localSeller); err != nil {
		return fmt.Errorf("AddSeller() err adding localSeller: %w", err)
	}

	if err := db.AddSeller(ctx, ghostSeller); err != nil {
		return fmt.Errorf("AddSeller() err adding ghostSeller: %w", err)
	}

	localSeller, err = db.Seller("local")
	if err != nil {
		return fmt.Errorf("Error getting local seller: %v", err)
	}

	ghostSeller, err = db.Seller("ghost-army")
	if err != nil {
		return fmt.Errorf("Error getting ghost seller: %v", err)
	}

	localDCID := crypto.HashID("local.locale.name")
	localDatacenter := routing.Datacenter{
		ID:      localDCID,
		Name:    "local.locale.name",
		Enabled: true,
		Location: routing.Location{
			Latitude:  70.5,
			Longitude: 120.5,
		},
		StreetAddress: "Somewhere, USA",
		SupplierName:  "supplier.local.name",
		SellerID:      localSeller.DatabaseID,
	}
	if err := db.AddDatacenter(ctx, localDatacenter); err != nil {
		return fmt.Errorf("AddDatacenter() error adding local datacenter: %w", err)
	}

	ghostDCID := crypto.HashID("ghost-army.locale.name")
	ghostDatacenter := routing.Datacenter{
		ID:      ghostDCID,
		Name:    "ghost-army.locale.name",
		Enabled: true,
		Location: routing.Location{
			Latitude:  70.5,
			Longitude: 120.5,
		},
		StreetAddress: "Somewhere, Else, USA",
		SupplierName:  "supplier.ghost.name",
		SellerID:      ghostSeller.DatabaseID,
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
	localDcMap := routing.DatacenterMap{
		Alias:        "local.map",
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

	// Add the number of relays provided by LOCAL_RELAYS for each datacenter
	numRelays := uint64(10)
	if val, ok := os.LookupEnv("LOCAL_RELAYS"); ok {
		numRelays, err = strconv.ParseUint(val, 10, 64)
		if err != nil {
			return fmt.Errorf("LOCAL_RELAYS ParseUint() err: %w", err)
		}
	}

	for i := uint64(0); i < numRelays; i++ {

		// local
		publicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		if err != nil {
			return fmt.Errorf("Error generating relay public key: %v", err)
		}

		updateKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(updateKey)
		if err != nil {
			return fmt.Errorf("Error generating relay update key: %v", err)
		}

		addr := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000 + int(i)}
		rid := crypto.HashID(addr.String())

		if err := db.AddRelay(ctx, routing.Relay{
			ID:             rid,
			Name:           "local.locale." + fmt.Sprintf("%d", i),
			Addr:           addr,
			ManagementAddr: "1.2.3.4" + fmt.Sprintf("%d", i),
			SSHPort:        22,
			SSHUser:        "root",
			MaxSessions:    uint32(1000 + i),
			PublicKey:      publicKey,
			UpdateKey:      updateKey,
			Datacenter:     localDatacenter,
			MRC:            19700000000000,
			Overage:        26000000000000,
			BWRule:         routing.BWRuleBurst,
			ContractTerm:   12,
			StartDate:      time.Now(),
			EndDate:        time.Now(),
			Type:           routing.BareMetal,
			State:          routing.RelayStateMaintenance,
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

		if err := db.AddRelay(ctx, routing.Relay{
			ID:             rid,
			Name:           "ghost-army.locale.1" + fmt.Sprintf("%d", i),
			Addr:           addr,
			ManagementAddr: "4.3.2.1" + fmt.Sprintf("%d", i),
			SSHPort:        22,
			SSHUser:        "root",
			MaxSessions:    uint32(1000 + i),
			PublicKey:      publicKey,
			UpdateKey:      updateKey,
			Datacenter:     ghostDatacenter,
			MRC:            19700000000000,
			Overage:        26000000000000,
			BWRule:         routing.BWRuleBurst,
			ContractTerm:   12,
			StartDate:      time.Now(),
			EndDate:        time.Now(),
			Type:           routing.BareMetal,
			State:          routing.RelayStateMaintenance,
		}); err != nil {
			return fmt.Errorf("AddRelay() error adding ghost relay: %w", err)
		}

		if i%25 == 0 {
			time.Sleep(time.Millisecond * 500)
		}
	}
	return nil
}
