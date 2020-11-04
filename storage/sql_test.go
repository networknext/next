package storage_test

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/stretchr/testify/assert"
)

// TODO: first clean and reset postgresql

// TODO: sort environment in/from Makefile
func SetupEnv() {
	os.Setenv("ENV", "local")
	os.Setenv("PGSQL", "false")
	os.Setenv("DB_SYNC_INTERVAL", "10s")
}

func TestInsertSQL(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	fmt.Println("Starting Add SQL tests.")

	// NewSQLStorage syncs the local sync number from the remote and
	// runs all the sync*() methods
	db, err := storage.NewSQLStorage(ctx, logger)
	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	var outerCustomer routing.Customer
	var outerBuyer routing.Buyer
	var outerSeller routing.Seller
	var outerDatacenter routing.Datacenter

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

	t.Run("AddCustomer", func(t *testing.T) {
		customer := routing.Customer{
			Active:                 true,
			Debug:                  true,
			Code:                   "Compcode",
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		_, err := db.Customer("Compcode")
		assert.NoError(t, err)
		outerCustomer, err = db.Customer("Compcode")
		assert.NoError(t, err)
		assert.Equal(t, customer.Active, outerCustomer.Active)
		assert.Equal(t, customer.Code, outerCustomer.Code)
		assert.Equal(t, customer.Name, outerCustomer.Name)
		assert.Equal(t, customer.AutomaticSignInDomains, outerCustomer.AutomaticSignInDomains)
	})

	t.Run("AddSeller", func(t *testing.T) {
		seller := routing.Seller{
			ID:                        "Compcode",
			IngressPriceNibblinsPerGB: 10,
			EgressPriceNibblinsPerGB:  20,
			CustomerID:                outerCustomer.CustomerID,
		}

		err = db.AddSeller(ctx, seller)
		assert.NoError(t, err)

		outerSeller, err = db.Seller("Compcode")
		assert.NoError(t, err)
		assert.Equal(t, seller.ID, outerSeller.ID)
		assert.Equal(t, seller.IngressPriceNibblinsPerGB, outerSeller.IngressPriceNibblinsPerGB)
		assert.Equal(t, seller.EgressPriceNibblinsPerGB, outerSeller.EgressPriceNibblinsPerGB)
		assert.Equal(t, seller.CustomerID, outerSeller.CustomerID)
	})

	t.Run("AddDatacenter", func(t *testing.T) {

		datacenter := routing.Datacenter{
			ID:      crypto.HashID("some.locale.name"),
			Name:    "some.locale.name",
			Enabled: true,
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
			StreetAddress: "Somewhere, USA",
			SupplierName:  "supplier.local.name",
			SellerID:      outerSeller.SellerID,
		}

		err = db.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		outerDatacenter, err = db.Datacenter(datacenter.ID)
		assert.NoError(t, err)
		assert.Equal(t, outerDatacenter.ID, datacenter.ID)
		assert.Equal(t, outerDatacenter.Name, datacenter.Name)
		assert.Equal(t, outerDatacenter.StreetAddress, datacenter.StreetAddress)
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
			ID:         internalID,
			Live:       true,
			Debug:      true,
			PublicKey:  publicKey,
			CustomerID: outerCustomer.CustomerID,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		outerBuyer, err = db.Buyer(internalID)
		assert.NoError(t, err)

		assert.Equal(t, buyer.Live, outerBuyer.Live)
		assert.Equal(t, buyer.Debug, outerBuyer.Debug)
		assert.Equal(t, publicKey, outerBuyer.PublicKey)
		assert.Equal(t, buyer.CustomerID, outerBuyer.CustomerID)
	})

	t.Run("AddRelay", func(t *testing.T) {

		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		rid := crypto.HashID(addr.String())

		publicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		assert.NoError(t, err)

		updateKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(updateKey)
		assert.NoError(t, err)

		// fmt.Printf("AddRelay test - outerDatacenter: %s\n", outerDatacenter.String())

		// fields not stored in the database are not tested here
		relay := routing.Relay{
			ID:             rid,
			Name:           "local.1",
			Addr:           *addr,
			ManagementAddr: "1.2.3.4",
			SSHPort:        22,
			SSHUser:        "fred",
			MaxSessions:    1000,
			PublicKey:      publicKey,
			UpdateKey:      updateKey,
			// Datacenter:     outerDatacenter,
			MRC:          19700000000000,
			Overage:      26000000000000,
			BWRule:       routing.BWRuleBurst,
			ContractTerm: 12,
			StartDate:    time.Now(),
			EndDate:      time.Now(),
			Type:         routing.BareMetal,
			State:        routing.RelayStateMaintenance,
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
		assert.Equal(t, relay.UpdateKey, checkRelay.UpdateKey)
		assert.Equal(t, relay.Datacenter.DatacenterID, checkRelay.Datacenter.DatacenterID)
		assert.Equal(t, relay.MRC, checkRelay.MRC)
		assert.Equal(t, relay.Overage, checkRelay.Overage)
		assert.Equal(t, relay.BWRule, checkRelay.BWRule)
		assert.Equal(t, relay.ContractTerm, checkRelay.ContractTerm)
		assert.Equal(t, relay.StartDate.Format("01/02/06"), checkRelay.StartDate.Format("01/02/06"))
		assert.Equal(t, relay.EndDate.Format("01/02/06"), checkRelay.EndDate.Format("01/02/06"))
		assert.Equal(t, relay.Type, checkRelay.Type)
		assert.Equal(t, relay.State, checkRelay.State)
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
		assert.Equal(t, dcMap.Alias, checkDCMaps[outerBuyer.ID].Alias)
		assert.Equal(t, dcMap.BuyerID, checkDCMaps[outerBuyer.ID].BuyerID)
		assert.Equal(t, dcMap.DatacenterID, checkDCMaps[outerBuyer.ID].DatacenterID)
	})
}

func TestDeleteSQL(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	db, err := storage.NewSQLStorage(ctx, logger)
	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	var outerCustomer routing.Customer
	var outerBuyer routing.Buyer
	var outerSeller routing.Seller
	var outerDatacenter routing.Datacenter
	var outerDatacenterMap routing.DatacenterMap

	t.Run("ExerciseFKs", func(t *testing.T) {

		customer := routing.Customer{
			Active:                 true,
			Code:                   "Compcode",
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		outerCustomer, err = db.Customer("Compcode")
		assert.NoError(t, err)

		publicKey := make([]byte, crypto.KeySize)
		_, err := rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := binary.LittleEndian.Uint64(publicKey[:8])

		buyer := routing.Buyer{
			ID:         internalID,
			Live:       true,
			Debug:      false,
			PublicKey:  publicKey,
			CustomerID: outerCustomer.CustomerID,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		outerBuyer, err = db.Buyer(internalID)
		assert.NoError(t, err)

		seller := routing.Seller{
			ID:                        "Compcode",
			IngressPriceNibblinsPerGB: 10,
			EgressPriceNibblinsPerGB:  20,
			CustomerID:                outerCustomer.CustomerID,
		}

		err = db.AddSeller(ctx, seller)
		assert.NoError(t, err)

		outerSeller, err = db.Seller("Compcode")
		assert.NoError(t, err)

		datacenter := routing.Datacenter{
			ID:      crypto.HashID("some.locale.name"),
			Name:    "some.locale.name",
			Enabled: true,
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
			StreetAddress: "Somewhere, USA",
			SellerID:      outerSeller.SellerID,
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
		outerDatacenterMap = dcMaps[outerBuyer.ID]

		addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		rid := crypto.HashID(addr.String())

		relayPublicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(relayPublicKey)
		assert.NoError(t, err)

		updateKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(updateKey)
		assert.NoError(t, err)

		relay := routing.Relay{
			ID:             rid,
			Name:           "local.1",
			Addr:           *addr,
			ManagementAddr: "1.2.3.4",
			SSHPort:        22,
			SSHUser:        "fred",
			MaxSessions:    1000,
			PublicKey:      relayPublicKey,
			UpdateKey:      updateKey,
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

	db, err := storage.NewSQLStorage(ctx, logger)
	time.Sleep(1000 * time.Millisecond) // allow time for sync functions to complete
	assert.NoError(t, err)

	var customerWithID routing.Customer
	// var outerBuyer routing.Buyer
	// var outerSeller routing.Seller
	// var outerDatacenter routing.Datacenter
	// var outerDatacenterMap routing.DatacenterMap

	t.Run("SetCustomer", func(t *testing.T) {
		customer := routing.Customer{
			Active:                 true,
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
		customerWithID.Active = false
		customerWithID.Debug = false

		err = db.SetCustomer(ctx, customerWithID)
		assert.NoError(t, err)

		checkCustomer, err := db.Customer("Compcode")
		assert.NoError(t, err)

		assert.Equal(t, customerWithID.Active, checkCustomer.Active)
		assert.Equal(t, customerWithID.Debug, checkCustomer.Debug)
		assert.Equal(t, customerWithID.AutomaticSignInDomains, checkCustomer.AutomaticSignInDomains)
		assert.Equal(t, customerWithID.Name, checkCustomer.Name)

	})

	t.Run("SetBuyer", func(t *testing.T) {

		publicKey := make([]byte, crypto.KeySize)
		_, err = rand.Read(publicKey)
		assert.NoError(t, err)

		internalID := binary.LittleEndian.Uint64(publicKey[:8])

		buyer := routing.Buyer{
			ID:         internalID,
			Live:       true,
			Debug:      true,
			PublicKey:  publicKey,
			CustomerID: customerWithID.CustomerID,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		buyerWithID, err := db.Buyer(internalID)
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
}
