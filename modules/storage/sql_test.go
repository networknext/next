package storage_test

import (
	"context"
	"encoding/binary"
	"math/rand"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/stretchr/testify/assert"
)

// TODO: first clean and reset postgresql

// TODO: sort environment in/from Makefile
func SetupEnv() {
	os.Setenv("ENV", "local")
	os.Setenv("FEATURE_POSTGRESQL", "false")
	os.Setenv("DB_SYNC_INTERVAL", "10s")
}

func TestInsertSQL(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	// NewSQLStorage syncs the local sync number from the remote and
	// runs all the sync*() methods
	// db, err := storage.NewSQLStorage(ctx, logger)
	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete

	var outerCustomer routing.Customer
	var outerBuyer routing.Buyer
	var outerSeller routing.Seller
	var outerDatacenter routing.Datacenter

	// currentLocation, err := os.Getwd()
	// assert.NoError(t, err)
	// fmt.Printf("Current disk location: %s\n", currentLocation)

	err = db.SetSequenceNumber(ctx, -1)
	assert.NoError(t, err)

	// err = db.IncrementSequenceNumber(ctx)
	// assert.NoError(t, err)

	// NewSQLStorage() Sync() above sets up seq number
	t.Run("Do Not Sync", func(t *testing.T) {
		sync, _, err := db.CheckSequenceNumber(ctx)
		assert.NoError(t, err)
		assert.Equal(t, false, sync)
	})

	t.Run("IncrementSequenceNumber", func(t *testing.T) {
		err = db.IncrementSequenceNumber(ctx)
		sync, _, err := db.CheckSequenceNumber(ctx)
		assert.NoError(t, err)
		assert.Equal(t, true, sync)
	})

	customerShortname := "Compcode"

	t.Run("AddCustomer", func(t *testing.T) {
		customer := routing.Customer{
			Code:                   customerShortname,
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		outerCustomer, err = db.Customer(customerShortname)
		assert.NoError(t, err)
		assert.Equal(t, customer.Code, outerCustomer.Code)
		assert.Equal(t, customer.Name, outerCustomer.Name)
		assert.Equal(t, customer.AutomaticSignInDomains, outerCustomer.AutomaticSignInDomains)
	})

	t.Run("AddSeller", func(t *testing.T) {
		seller := routing.Seller{
			ID:                        customerShortname,
			ShortName:                 customerShortname,
			CompanyCode:               customerShortname,
			IngressPriceNibblinsPerGB: 10,
			EgressPriceNibblinsPerGB:  20,
			CustomerID:                outerCustomer.DatabaseID,
		}

		err = db.AddSeller(ctx, seller)
		assert.NoError(t, err)

		outerSeller, err = db.Seller("Compcode")
		assert.NoError(t, err)
		assert.Equal(t, seller.ID, outerSeller.ID)
		assert.Equal(t, seller.ShortName, outerSeller.ShortName)
		assert.Equal(t, seller.IngressPriceNibblinsPerGB, outerSeller.IngressPriceNibblinsPerGB)
		assert.Equal(t, seller.EgressPriceNibblinsPerGB, outerSeller.EgressPriceNibblinsPerGB)
		assert.Equal(t, seller.CustomerID, outerSeller.CustomerID)
	})

	t.Run("AddDatacenter", func(t *testing.T) {

		datacenter := routing.Datacenter{
			ID:   crypto.HashID("some.locale.name"),
			Name: "some.locale.name",
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
			SupplierName: "supplier.local.name",
			SellerID:     outerSeller.DatabaseID,
		}

		err = db.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		outerDatacenter, err = db.Datacenter(datacenter.ID)
		assert.NoError(t, err)
		assert.Equal(t, outerDatacenter.ID, datacenter.ID)
		assert.Equal(t, outerDatacenter.Name, datacenter.Name)
		assert.Equal(t, outerDatacenter.Location.Latitude, datacenter.Location.Latitude)
		assert.Equal(t, outerDatacenter.Location.Longitude, datacenter.Location.Longitude)
		assert.Equal(t, outerDatacenter.SupplierName, datacenter.SupplierName)
		assert.Equal(t, outerDatacenter.SellerID, datacenter.SellerID)
	})

	t.Run("AddBuyer", func(t *testing.T) {

		publicKey := make([]byte, crypto.KeySize)
		_, err := rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := binary.LittleEndian.Uint64(publicKey[:8])

		buyer := routing.Buyer{
			// ID:          internalID,
			ShortName:   outerCustomer.Code,
			CompanyCode: outerCustomer.Code,
			Live:        true,
			Debug:       true,
			PublicKey:   publicKey,
			// CustomerID:  outerCustomer.DatabaseID,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		outerBuyer, err = db.Buyer(internalID)
		assert.NoError(t, err)

		assert.Equal(t, internalID, outerBuyer.ID)
		assert.Equal(t, buyer.Live, outerBuyer.Live)
		assert.Equal(t, buyer.Debug, outerBuyer.Debug)
		assert.Equal(t, publicKey, outerBuyer.PublicKey)
		// assert.Equal(t, buyer.CustomerID, outerBuyer.CustomerID)
		assert.Equal(t, buyer.ShortName, outerBuyer.ShortName)
		assert.Equal(t, buyer.CompanyCode, outerBuyer.CompanyCode)
	})

	t.Run("AddRelay", func(t *testing.T) {

		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		internalAddr, err := net.ResolveUDPAddr("udp", "172.20.2.6:40000")
		assert.NoError(t, err)

		rid := crypto.HashID(addr.String())

		publicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		assert.NoError(t, err)

		// fields not stored in the database are not tested here
		relay := routing.Relay{
			ID:             rid,
			Name:           "local.1",
			Addr:           *addr,
			InternalAddr:   *internalAddr,
			ManagementAddr: "1.2.3.4",
			SSHPort:        22,
			SSHUser:        "fred",
			MaxSessions:    1000,
			PublicKey:      publicKey,
			// Datacenter:     outerDatacenter,
			MRC:                 19700000000000,
			Overage:             26000000000000,
			BWRule:              routing.BWRuleBurst,
			ContractTerm:        12,
			StartDate:           time.Now(),
			EndDate:             time.Now(),
			Type:                routing.BareMetal,
			State:               routing.RelayStateMaintenance,
			IncludedBandwidthGB: 10000,
			NICSpeedMbps:        1000,
		}

		// adding a relay w/o a valid datacenter should return an FK violation error
		err = db.AddRelay(ctx, relay)
		assert.Error(t, err)

		// TODO repeat the above test with bwrule, type and state

		relay.Datacenter = outerDatacenter
		err = db.AddRelay(ctx, relay)
		assert.NoError(t, err)

		// check only the fields set above
		checkRelay, err := db.Relay(rid)
		assert.NoError(t, err)

		assert.Equal(t, relay.Name, checkRelay.Name)
		assert.Equal(t, relay.Addr, checkRelay.Addr)
		assert.Equal(t, relay.ManagementAddr, checkRelay.ManagementAddr)
		assert.Equal(t, relay.SSHPort, checkRelay.SSHPort)
		assert.Equal(t, relay.SSHUser, checkRelay.SSHUser)
		assert.Equal(t, relay.MaxSessions, checkRelay.MaxSessions)
		assert.Equal(t, relay.PublicKey, checkRelay.PublicKey)
		assert.Equal(t, relay.Datacenter.DatabaseID, checkRelay.Datacenter.DatabaseID)
		assert.Equal(t, relay.MRC, checkRelay.MRC)
		assert.Equal(t, relay.Overage, checkRelay.Overage)
		assert.Equal(t, relay.BWRule, checkRelay.BWRule)
		assert.Equal(t, relay.ContractTerm, checkRelay.ContractTerm)
		assert.Equal(t, relay.StartDate.Format("01/02/06"), checkRelay.StartDate.Format("01/02/06"))
		assert.Equal(t, relay.EndDate.Format("01/02/06"), checkRelay.EndDate.Format("01/02/06"))
		assert.Equal(t, relay.Type, checkRelay.Type)
		assert.Equal(t, relay.State, checkRelay.State)
		assert.Equal(t, int32(10000), checkRelay.IncludedBandwidthGB)
		assert.Equal(t, int32(1000), checkRelay.NICSpeedMbps)

		assert.Equal(t, customerShortname, checkRelay.Seller.ID)
		assert.Equal(t, customerShortname, checkRelay.Seller.ShortName)
		assert.Equal(t, customerShortname, checkRelay.Seller.CompanyCode)
		assert.Equal(t, routing.Nibblin(10), checkRelay.Seller.IngressPriceNibblinsPerGB)
		assert.Equal(t, routing.Nibblin(20), checkRelay.Seller.EgressPriceNibblinsPerGB)
		assert.Equal(t, outerCustomer.DatabaseID, checkRelay.Seller.CustomerID)
	})

	t.Run("AddDatacenterMap", func(t *testing.T) {
		dcMap := routing.DatacenterMap{
			Alias:        "local.map",
			BuyerID:      outerBuyer.ID,
			DatacenterID: outerDatacenter.ID,
		}

		err := db.AddDatacenterMap(ctx, dcMap)
		assert.NoError(t, err)

		checkDCMaps := db.GetDatacenterMapsForBuyer(outerBuyer.ID)
		assert.Equal(t, 1, len(checkDCMaps))
		assert.Equal(t, dcMap.Alias, checkDCMaps[outerDatacenter.ID].Alias)
		assert.Equal(t, dcMap.BuyerID, checkDCMaps[outerDatacenter.ID].BuyerID)
		assert.Equal(t, dcMap.DatacenterID, checkDCMaps[outerDatacenter.ID].DatacenterID)
	})
}

func TestDeleteSQL(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	// db, err := storage.NewSQLStorage(ctx, logger)
	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	var outerCustomer routing.Customer
	var outerBuyer routing.Buyer
	var outerSeller routing.Seller
	var outerDatacenter routing.Datacenter
	var outerDatacenterMap routing.DatacenterMap

	t.Run("ExerciseFKs", func(t *testing.T) {

		customerCode := "Compcode"
		customer := routing.Customer{
			Code:                   customerCode,
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		outerCustomer, err = db.Customer(customerCode)
		assert.NoError(t, err)

		publicKey := make([]byte, crypto.KeySize)
		_, err := rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := binary.LittleEndian.Uint64(publicKey[:8])

		buyer := routing.Buyer{
			// ID:          internalID,
			ShortName:   outerCustomer.Code,
			CompanyCode: outerCustomer.Code,
			Live:        true,
			Debug:       true,
			PublicKey:   publicKey,
			// CustomerID:  outerCustomer.DatabaseID,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		outerBuyer, err = db.Buyer(internalID)
		assert.NoError(t, err)

		seller := routing.Seller{
			ID:                        "Compcode",
			IngressPriceNibblinsPerGB: 10,
			EgressPriceNibblinsPerGB:  20,
			CustomerID:                outerCustomer.DatabaseID,
			CompanyCode:               outerCustomer.Code,
		}

		err = db.AddSeller(ctx, seller)
		assert.NoError(t, err)

		outerSeller, err = db.Seller("Compcode")
		assert.NoError(t, err)

		datacenter := routing.Datacenter{
			ID:   crypto.HashID("some.locale.name"),
			Name: "some.locale.name",
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
			SellerID: outerSeller.DatabaseID,
		}

		err = db.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		outerDatacenter, err = db.Datacenter(datacenter.ID)
		assert.NoError(t, err)

		dcMap := routing.DatacenterMap{
			Alias:        "local.map",
			BuyerID:      outerBuyer.ID,
			DatacenterID: outerDatacenter.ID,
		}

		err = db.AddDatacenterMap(ctx, dcMap)
		assert.NoError(t, err)

		dcMaps := db.GetDatacenterMapsForBuyer(outerBuyer.ID)
		assert.Equal(t, 1, len(dcMaps))
		outerDatacenterMap = dcMaps[outerDatacenter.ID]

		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		internalAddr, err := net.ResolveUDPAddr("udp", "172.20.2.6:40000")
		assert.NoError(t, err)

		rid := crypto.HashID(addr.String())

		relayPublicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(relayPublicKey)
		assert.NoError(t, err)

		relay := routing.Relay{
			ID:             rid,
			Name:           "local.1",
			Addr:           *addr,
			InternalAddr:   *internalAddr,
			ManagementAddr: "1.2.3.4",
			SSHPort:        22,
			SSHUser:        "fred",
			MaxSessions:    1000,
			PublicKey:      relayPublicKey,
			Datacenter:     outerDatacenter,
			MRC:            19700000000000,
			Overage:        26000000000000,
			BWRule:         routing.BWRuleBurst,
			ContractTerm:   12,
			StartDate:      time.Now(),
			EndDate:        time.Now(),
			Type:           routing.BareMetal,
			State:          routing.RelayStateMaintenance,
		}

		err = db.AddRelay(ctx, relay)
		assert.NoError(t, err)

		// Attempting to remove the customer should return a foreign
		// key violation error (for buyer and/or seller)
		// sqlite3: FOREIGN KEY constraint failed
		err = db.RemoveCustomer(ctx, "Compcode")
		assert.Error(t, err)

		// Attempting to remove the buyer should return an FK
		// violation error (for datacenter maps and banned users (TBD))
		err = db.RemoveBuyer(ctx, outerBuyer.ID)
		assert.Error(t, err)

		// Attempting to remove the seller should return an FK
		// violation error (for the datacenter)
		err = db.RemoveSeller(ctx, outerSeller.ID)
		assert.Error(t, err)

		// Attempting to remove the datacenter should return an FK
		// violation error (for the datacenter map)
		err = db.RemoveDatacenter(ctx, outerDatacenter.ID)
		assert.Error(t, err)

		err = db.RemoveDatacenterMap(ctx, outerDatacenterMap)
		assert.NoError(t, err)

		dcMapsCheck := db.GetDatacenterMapsForBuyer(outerBuyer.ID)
		assert.Equal(t, 0, len(dcMapsCheck))

		err = db.RemoveBuyer(ctx, outerBuyer.ID)
		assert.NoError(t, err)

		_, err = db.Buyer(outerBuyer.ID)
		assert.Error(t, err)

		err = db.RemoveRelay(ctx, relay.ID)
		assert.NoError(t, err)

		_, err = db.Relay(relay.ID)
		assert.Error(t, err)

		err = db.RemoveDatacenter(ctx, outerDatacenter.ID)
		assert.NoError(t, err)

		_, err = db.Datacenter(outerDatacenter.ID)
		assert.Error(t, err)

		err = db.RemoveSeller(ctx, outerSeller.ID)
		assert.NoError(t, err)

		_, err = db.Seller(outerSeller.ID)
		assert.Error(t, err)

		err = db.RemoveCustomer(ctx, "Compcode")
		assert.NoError(t, err)

		_, err = db.Customer("Compcode")
		assert.Error(t, err)
	})
}

func TestUpdateSQL(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	// db, err := storage.NewSQLStorage(ctx, logger)
	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	var customerWithID routing.Customer
	var buyerWithID routing.Buyer
	var sellerWithID routing.Seller
	var datacenterWithID routing.Datacenter
	// var outerDatacenterMap routing.DatacenterMap

	t.Run("SetCustomer", func(t *testing.T) {
		customer := routing.Customer{
			Code:                   "Compcode",
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		// the CustomerID field is the PK and is set by AddCustomer(). In
		// production usage this field would already be set and sync'd.
		customerWithID, err = db.Customer("Compcode")

		customerWithID.Name = "No Longer The Company, Ltd."
		customerWithID.AutomaticSignInDomains = "fredscuttle.com,swampthing.com"

		err = db.SetCustomer(ctx, customerWithID)
		assert.NoError(t, err)

		checkCustomer, err := db.Customer("Compcode")
		assert.NoError(t, err)

		assert.Equal(t, customerWithID.AutomaticSignInDomains, checkCustomer.AutomaticSignInDomains)
		assert.Equal(t, customerWithID.Name, checkCustomer.Name)

	})

	t.Run("SetBuyer", func(t *testing.T) {

		publicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := binary.LittleEndian.Uint64(publicKey[:8])

		buyer := routing.Buyer{
			// ID:          internalID,
			ShortName:   customerWithID.Code,
			CompanyCode: customerWithID.Code,
			Live:        true,
			Debug:       true,
			PublicKey:   publicKey,
			// CustomerID:  customerWithID.DatabaseID,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		buyerWithID, err = db.Buyer(internalID)
		assert.NoError(t, err)

		buyerWithID.Live = false
		buyerWithID.Debug = false
		buyerWithID.PublicKey = []byte("")

		err = db.SetBuyer(ctx, buyerWithID)
		assert.NoError(t, err)

		checkBuyer, err := db.Buyer(internalID)
		assert.NoError(t, err)
		assert.Equal(t, checkBuyer.Live, buyerWithID.Live)
		assert.Equal(t, checkBuyer.Debug, buyerWithID.Debug)
		assert.Equal(t, checkBuyer.PublicKey, buyerWithID.PublicKey)

	})

	t.Run("SetSeller", func(t *testing.T) {
		seller := routing.Seller{
			ID:                        "Compcode",
			IngressPriceNibblinsPerGB: 10,
			EgressPriceNibblinsPerGB:  20,
			CustomerID:                customerWithID.DatabaseID,
			CompanyCode:               customerWithID.Code,
		}

		err = db.AddSeller(ctx, seller)
		assert.NoError(t, err)

		sellerWithID, err = db.Seller("Compcode")
		assert.NoError(t, err)

		sellerWithID.IngressPriceNibblinsPerGB = 100
		sellerWithID.EgressPriceNibblinsPerGB = 200

		err = db.SetSeller(ctx, sellerWithID)
		assert.NoError(t, err)

		checkSeller, err := db.Seller("Compcode")
		assert.NoError(t, err)
		assert.Equal(t, checkSeller.IngressPriceNibblinsPerGB, sellerWithID.IngressPriceNibblinsPerGB)
		assert.Equal(t, checkSeller.EgressPriceNibblinsPerGB, sellerWithID.EgressPriceNibblinsPerGB)
	})

	t.Run("SetDatacenter", func(t *testing.T) {

		did := crypto.HashID("some.locale.name")
		datacenter := routing.Datacenter{
			ID:   did,
			Name: "some.locale.name",
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
			SupplierName: "supplier.local.name",
			SellerID:     sellerWithID.DatabaseID,
		}

		err = db.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		datacenterWithID, err = db.Datacenter(did)
		assert.NoError(t, err)

		modifiedDatacenter := datacenterWithID
		modifiedDatacenter.Name = "some.newlocale.name"
		modifiedDatacenter.Location.Longitude = 70.5
		modifiedDatacenter.Location.Latitude = 120.5
		modifiedDatacenter.SupplierName = "supplier.nonlocal.name"

		err = db.SetDatacenter(ctx, modifiedDatacenter)
		assert.NoError(t, err)

		checkModDC, err := db.Datacenter(did)
		assert.NoError(t, err)
		assert.Equal(t, modifiedDatacenter.Name, checkModDC.Name)
		assert.Equal(t, modifiedDatacenter.Location.Longitude, checkModDC.Location.Longitude)
		assert.Equal(t, modifiedDatacenter.Location.Latitude, checkModDC.Location.Latitude)
		assert.Equal(t, modifiedDatacenter.SupplierName, checkModDC.SupplierName)
	})

	t.Run("UpdateCustomer", func(t *testing.T) {
		err := db.UpdateCustomer(ctx, customerWithID.Code, "Name", "A Brand New Name")
		assert.NoError(t, err)

		err = db.UpdateCustomer(ctx, customerWithID.Code, "AutomaticSigninDomains", "somewhere.com,somewhere.else.com")
		assert.NoError(t, err)

		checkCustomer, err := db.Customer(customerWithID.Code)
		assert.NoError(t, err)

		assert.Equal(t, "A Brand New Name", checkCustomer.Name)
		assert.Equal(t, "somewhere.com,somewhere.else.com", checkCustomer.AutomaticSignInDomains)
	})

	t.Run("UpdateBuyer", func(t *testing.T) {
		err := db.UpdateBuyer(ctx, buyerWithID.ID, "Live", false)
		assert.NoError(t, err)

		err = db.UpdateBuyer(ctx, buyerWithID.ID, "Debug", false)
		assert.NoError(t, err)

		err = db.UpdateBuyer(ctx, buyerWithID.ID, "ShortName", "newname")
		assert.NoError(t, err)

		checkBuyer, err := db.Buyer(buyerWithID.ID)
		assert.NoError(t, err)

		assert.Equal(t, false, checkBuyer.Live)
		assert.Equal(t, false, checkBuyer.Debug)
		assert.Equal(t, "newname", checkBuyer.ShortName)
	})

	t.Run("UpdateSeller", func(t *testing.T) {
		err := db.UpdateSeller(ctx, sellerWithID.ID, "EgressPriceNibblinsPerGB", 133.44)
		assert.NoError(t, err)

		err = db.UpdateSeller(ctx, sellerWithID.ID, "IngressPriceNibblinsPerGB", 144.33)
		assert.NoError(t, err)

		err = db.UpdateSeller(ctx, sellerWithID.ID, "ShortName", "newname")
		assert.NoError(t, err)

		checkSeller, err := db.Seller(sellerWithID.ID)
		assert.NoError(t, err)

		assert.Equal(t, routing.Nibblin(13344000000000), checkSeller.EgressPriceNibblinsPerGB)
		assert.Equal(t, routing.Nibblin(14433000000000), checkSeller.IngressPriceNibblinsPerGB)
		assert.Equal(t, "newname", checkSeller.ShortName)
	})

	t.Run("UpdateRelay", func(t *testing.T) {

		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		rid := crypto.HashID(addr.String())

		internalAddr, err := net.ResolveUDPAddr("udp", "172.20.2.6:40000")
		assert.NoError(t, err)

		publicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		assert.NoError(t, err)

		relay := routing.Relay{
			ID:                  rid,
			Name:                "local.1",
			Addr:                *addr,
			InternalAddr:        *internalAddr,
			ManagementAddr:      "1.2.3.4",
			SSHPort:             22,
			SSHUser:             "fred",
			MaxSessions:         1000,
			PublicKey:           publicKey,
			Datacenter:          datacenterWithID,
			MRC:                 19700000000000,
			Overage:             26000000000000,
			BWRule:              routing.BWRuleBurst,
			NICSpeedMbps:        1000,
			IncludedBandwidthGB: 10000,
			ContractTerm:        12,
			StartDate:           time.Now(),
			EndDate:             time.Now(),
			Type:                routing.BareMetal,
			State:               routing.RelayStateMaintenance,
		}

		err = db.AddRelay(ctx, relay)
		assert.NoError(t, err)

		_, err = db.Relay(rid)
		assert.NoError(t, err)

		// relay.Name
		err = db.UpdateRelay(ctx, rid, "Name", "local.2")
		assert.NoError(t, err)
		checkRelay, err := db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, "local.2", checkRelay.Name)

		// relay.Addr
		newAddr, err := net.ResolveUDPAddr("udp", "192.168.0.1:40000")
		assert.NoError(t, err)
		err = db.UpdateRelay(ctx, rid, "Addr", "192.168.0.1:40000")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, *newAddr, checkRelay.Addr)

		// relay.InternalAddr
		intAddr, err := net.ResolveUDPAddr("udp", "192.168.0.2:40000")
		assert.NoError(t, err)
		err = db.UpdateRelay(ctx, rid, "InternalAddr", "192.168.0.2:40000")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, *intAddr, checkRelay.InternalAddr)

		// relay.ManagementAddr
		err = db.UpdateRelay(ctx, rid, "ManagementAddr", "9.8.7.6")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, "9.8.7.6", checkRelay.ManagementAddr)

		// relay.SSHPort
		// Note: ints in json are unmarshalled as float64
		err = db.UpdateRelay(ctx, rid, "SSHPort", float64(13))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, int64(13), checkRelay.SSHPort)

		// checkRelay.SSHUser
		err = db.UpdateRelay(ctx, rid, "SSHUser", "Abercrombie")
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, "Abercrombie", checkRelay.SSHUser)

		// relay.MaxSessions
		err = db.UpdateRelay(ctx, rid, "MaxSessions", float64(25000))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, uint32(25000), checkRelay.MaxSessions)

		// relay.PublicKey
		err = db.UpdateRelay(ctx, rid, "PublicKey", []byte("public key"))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, []byte("public key"), checkRelay.PublicKey)

		// relay.Datacenter = only one datacenter available...

		// relay.MRC
		err = db.UpdateRelay(ctx, rid, "MRC", float64(397))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, routing.Nibblin(39700000000000), checkRelay.MRC)

		// relay.Overage
		err = db.UpdateRelay(ctx, rid, "Overage", float64(260))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, routing.Nibblin(26000000000000), checkRelay.Overage)

		// relay.BWRule
		err = db.UpdateRelay(ctx, rid, "BWRule", float64(3))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, routing.BWRulePool, checkRelay.BWRule)

		// relay.ContractTerm
		err = db.UpdateRelay(ctx, rid, "ContractTerm", float64(1))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), checkRelay.ContractTerm)

		// relay.StartDate
		// We use a string as type-switching (in UpdateRelay()) doesn't work with a time.Time type
		startDate := "July 7, 2023"
		err = db.UpdateRelay(ctx, rid, "StartDate", startDate)
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		startDateFormatted, err := time.Parse("January 2, 2006", startDate)
		assert.NoError(t, err)
		assert.Equal(t, startDateFormatted, checkRelay.StartDate)

		// relay.EndDate
		endDate := "July 7, 2025"
		err = db.UpdateRelay(ctx, rid, "EndDate", endDate)
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		endDateFormatted, err := time.Parse("January 2, 2006", endDate)
		assert.NoError(t, err)
		assert.Equal(t, endDateFormatted, checkRelay.EndDate)

		// relay.Type
		err = db.UpdateRelay(ctx, rid, "Type", float64(2))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, routing.VirtualMachine, checkRelay.Type)

		// relay.State
		err = db.UpdateRelay(ctx, rid, "State", float64(0))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, routing.RelayStateEnabled, checkRelay.State)

		// relay.NICSpeedMbps
		err = db.UpdateRelay(ctx, rid, "NICSpeedMbps", float64(20000))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, int32(20000), checkRelay.NICSpeedMbps)

		// relay.IncludedBandwidthGB
		err = db.UpdateRelay(ctx, rid, "IncludedBandwidthGB", float64(25000))
		assert.NoError(t, err)
		checkRelay, err = db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, int32(25000), checkRelay.IncludedBandwidthGB)

	})
}

func TestInternalConfig(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	// db, err := storage.NewSQLStorage(ctx, logger)
	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	var outerCustomer routing.Customer
	var outerBuyer routing.Buyer
	var outerInternalConfig core.InternalConfig

	t.Run("AddInternalConfig", func(t *testing.T) {

		customerCode := "Compcode"
		customer := routing.Customer{
			Code:                   customerCode,
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		outerCustomer, err = db.Customer(customerCode)
		assert.NoError(t, err)

		publicKey := make([]byte, crypto.KeySize)
		_, err := rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := binary.LittleEndian.Uint64(publicKey[:8])

		buyer := routing.Buyer{
			ShortName:   outerCustomer.Code,
			CompanyCode: outerCustomer.Code,
			Live:        true,
			Debug:       true,
			PublicKey:   publicKey,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		outerBuyer, err = db.Buyer(internalID)
		assert.NoError(t, err)

		internalConfig := core.InternalConfig{
			RouteSelectThreshold:       2,
			RouteSwitchThreshold:       5,
			MaxLatencyTradeOff:         10,
			RTTVeto_Default:            -10,
			RTTVeto_PacketLoss:         -20,
			RTTVeto_Multipath:          -20,
			MultipathOverloadThreshold: 500,
			TryBeforeYouBuy:            true,
			ForceNext:                  true,
			LargeCustomer:              true,
			Uncommitted:                true,
			MaxRTT:                     300,
		}

		err = db.AddInternalConfig(ctx, internalConfig, outerBuyer.ID)
		assert.NoError(t, err)

		outerInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)

		assert.Equal(t, int32(2), outerInternalConfig.RouteSelectThreshold)
		assert.Equal(t, int32(5), outerInternalConfig.RouteSwitchThreshold)
		assert.Equal(t, int32(10), outerInternalConfig.MaxLatencyTradeOff)
		assert.Equal(t, int32(-10), outerInternalConfig.RTTVeto_Default)
		assert.Equal(t, int32(-20), outerInternalConfig.RTTVeto_PacketLoss)
		assert.Equal(t, int32(-20), outerInternalConfig.RTTVeto_Multipath)
		assert.Equal(t, int32(500), outerInternalConfig.MultipathOverloadThreshold)
		assert.Equal(t, true, outerInternalConfig.TryBeforeYouBuy)
		assert.Equal(t, true, outerInternalConfig.ForceNext)
		assert.Equal(t, true, outerInternalConfig.LargeCustomer)
		assert.Equal(t, true, outerInternalConfig.Uncommitted)
		assert.Equal(t, int32(300), outerInternalConfig.MaxRTT)
	})

	t.Run("UpdateInternalConfig", func(t *testing.T) {
		t.Skip() // working on it

		// RouteSelectThreshold
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "RouteSelectThreshold", int32(1))
		assert.NoError(t, err)
		checkInternalConfig, err := db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), checkInternalConfig.RouteSelectThreshold)

		// RouteSwitchThreshold
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "RouteSwitchThreshold", int32(4))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(4), checkInternalConfig.RouteSwitchThreshold)

		// MaxLatencyTradeOff
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "MaxLatencyTradeOff", int32(11))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(11), checkInternalConfig.MaxLatencyTradeOff)

		// RTTVeto_Default
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "RTTVeto_Default", int32(-20))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(-20), checkInternalConfig.RTTVeto_Default)

		// RTTVeto_PacketLoss
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "RTTVeto_PacketLoss", int32(-30))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(-30), checkInternalConfig.RTTVeto_PacketLoss)

		// RTTVeto_Multipath
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "RTTVeto_Multipath", int32(-40))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(-40), checkInternalConfig.RTTVeto_Multipath)

		// MultipathOverloadThreshold
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "MultipathOverloadThreshold", int32(600))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(600), checkInternalConfig.MultipathOverloadThreshold)

		// TryBeforeYouBuy
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "TryBeforeYouBuy", false)
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkInternalConfig.TryBeforeYouBuy)

		// ForceNext
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "ForceNext", false)
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkInternalConfig.ForceNext)

		// LargeCustomer
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "LargeCustomer", false)
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkInternalConfig.LargeCustomer)

		// Uncommitted
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "Uncommitted", false)
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkInternalConfig.Uncommitted)

		// MaxRTT
		err = db.UpdateInternalConfig(ctx, outerBuyer.ID, "MaxRTT", int32(400))
		assert.NoError(t, err)
		checkInternalConfig, err = db.InternalConfig(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(400), checkInternalConfig.MaxRTT)

	})

	t.Run("RemoveInternalConfig", func(t *testing.T) {

		err := db.RemoveInternalConfig(context.Background(), outerBuyer.ID)
		assert.NoError(t, err)

		_, err = db.InternalConfig(outerBuyer.ID)
		assert.Error(t, err)

	})

}

func TestRouteShaders(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	// db, err := storage.NewSQLStorage(ctx, logger)
	env, err := backend.GetEnv()
	assert.NoError(t, err)
	db, err := backend.GetStorer(ctx, logger, "local", env)
	assert.NoError(t, err)

	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	var outerCustomer routing.Customer
	var outerBuyer routing.Buyer
	var outerRouteShader core.RouteShader

	t.Run("AddRouteShader", func(t *testing.T) {

		customerCode := "Compcode"
		customer := routing.Customer{
			Code:                   customerCode,
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		outerCustomer, err = db.Customer(customerCode)
		assert.NoError(t, err)

		publicKey := make([]byte, crypto.KeySize)
		_, err := rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := binary.LittleEndian.Uint64(publicKey[:8])

		buyer := routing.Buyer{
			ShortName:   outerCustomer.Code,
			CompanyCode: outerCustomer.Code,
			Live:        true,
			Debug:       true,
			PublicKey:   publicKey,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		outerBuyer, err = db.Buyer(internalID)
		assert.NoError(t, err)

		routeShader := core.RouteShader{
			ABTest:                    true,
			AcceptableLatency:         int32(25),
			AcceptablePacketLoss:      float32(1),
			BandwidthEnvelopeDownKbps: int32(1200),
			BandwidthEnvelopeUpKbps:   int32(500),
			DisableNetworkNext:        true,
			LatencyThreshold:          int32(5),
			Multipath:                 true,
			ProMode:                   true,
			ReduceLatency:             true,
			ReducePacketLoss:          true,
			ReduceJitter:              true,
			SelectionPercent:          int(100),
		}

		err = db.AddRouteShader(ctx, routeShader, outerBuyer.ID)
		assert.NoError(t, err)

		outerRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)

		assert.Equal(t, true, outerRouteShader.ABTest)
		assert.Equal(t, int32(25), outerRouteShader.AcceptableLatency)
		assert.Equal(t, float32(1), outerRouteShader.AcceptablePacketLoss)
		assert.Equal(t, int32(1200), outerRouteShader.BandwidthEnvelopeDownKbps)
		assert.Equal(t, int32(500), outerRouteShader.BandwidthEnvelopeUpKbps)
		assert.Equal(t, true, outerRouteShader.DisableNetworkNext)
		assert.Equal(t, int32(5), outerRouteShader.LatencyThreshold)
		assert.Equal(t, true, outerRouteShader.Multipath)
		assert.Equal(t, true, outerRouteShader.ProMode)
		assert.Equal(t, true, outerRouteShader.ReduceLatency)
		assert.Equal(t, true, outerRouteShader.ReducePacketLoss)
		assert.Equal(t, true, outerRouteShader.ReduceJitter)
		assert.Equal(t, int(100), outerRouteShader.SelectionPercent)
	})

	t.Run("UpdateRouteShader", func(t *testing.T) {

		time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete

		// ABTest
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "ABTest", false)
		assert.NoError(t, err)
		checkRouteShader, err := db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.ABTest)

		// AcceptableLatency
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "AcceptableLatency", int32(35))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(35), checkRouteShader.AcceptableLatency)

		// AcceptablePacketLoss
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "AcceptablePacketLoss", float32(10))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, float32(10), checkRouteShader.AcceptablePacketLoss)

		// BandwidthEnvelopeDownKbps
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "BandwidthEnvelopeDownKbps", int32(1000))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(1000), checkRouteShader.BandwidthEnvelopeDownKbps)

		// BandwidthEnvelopeUpKbps
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "BandwidthEnvelopeUpKbps", int32(400))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(400), checkRouteShader.BandwidthEnvelopeUpKbps)

		// DisableNetworkNext
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "DisableNetworkNext", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.DisableNetworkNext)

		// LatencyThreshold
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "LatencyThreshold", int32(15))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int32(15), checkRouteShader.LatencyThreshold)

		// Multipath
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "Multipath", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.Multipath)

		// ProMode
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "ProMode", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.ProMode)

		// ReduceLatency
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "ReduceLatency", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.ReduceLatency)

		// ReducePacketLoss
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "ReducePacketLoss", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.ReducePacketLoss)

		// ReduceJitter
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "ReduceJitter", false)
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, false, checkRouteShader.ReduceJitter)

		// SelectionPercent
		err = db.UpdateRouteShader(ctx, outerBuyer.ID, "SelectionPercent", int(90))
		assert.NoError(t, err)
		checkRouteShader, err = db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, int(90), checkRouteShader.SelectionPercent)

	})

	t.Run("BannedUser tests", func(t *testing.T) {
		// random user IDs scraped from the portal
		userID1, err := strconv.ParseUint("77c556007df7c02e", 16, 64)
		assert.NoError(t, err)
		userID2, err := strconv.ParseUint("a731e14c521514a4", 16, 64)
		assert.NoError(t, err)
		userID3, err := strconv.ParseUint("fb6fa90ad67bc76a", 16, 64)
		assert.NoError(t, err)

		err = db.AddBannedUser(ctx, outerBuyer.ID, userID1)
		assert.NoError(t, err)
		err = db.AddBannedUser(ctx, outerBuyer.ID, userID2)
		assert.NoError(t, err)
		err = db.AddBannedUser(ctx, outerBuyer.ID, userID3)
		assert.NoError(t, err)

		bannedUserList, err := db.BannedUsers(outerBuyer.ID)
		assert.NoError(t, err)

		assert.True(t, bannedUserList[userID1])
		assert.True(t, bannedUserList[userID2])
		assert.True(t, bannedUserList[userID3])

		checkRouteShader, err := db.RouteShader(outerBuyer.ID)
		assert.NoError(t, err)
		assert.True(t, len(checkRouteShader.BannedUsers) > 0)
		assert.True(t, checkRouteShader.BannedUsers[userID1])
		assert.True(t, checkRouteShader.BannedUsers[userID2])
		assert.True(t, checkRouteShader.BannedUsers[userID3])

		err = db.RemoveBannedUser(ctx, outerBuyer.ID, userID1)
		assert.NoError(t, err)

		bannedUserList2, err := db.BannedUsers(outerBuyer.ID)
		assert.NoError(t, err)
		assert.False(t, bannedUserList2[userID1])

	})

	t.Run("RemoveRouteShader", func(t *testing.T) {
		t.Skip()
		err := db.RemoveRouteShader(context.Background(), outerBuyer.ID)
		assert.NoError(t, err)

		_, err = db.RouteShader(outerBuyer.ID)
		assert.Error(t, err)

	})

}
