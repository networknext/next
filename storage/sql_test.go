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

func TestSQL(t *testing.T) {

	SetupEnv()

	ctx := context.Background()
	logger := log.NewNopLogger()

	fmt.Println("Starting SQL tests.")

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

	// TODO: test "not null" constraints and failure modes
	t.Run("AddCustomer", func(t *testing.T) {
		customer := routing.Customer{
			Active:                 true,
			Code:                   "Compcode",
			Name:                   "Company, Ltd.",
			AutomaticSignInDomains: "fredscuttle.com",
		}

		err = db.AddCustomer(ctx, customer)
		assert.NoError(t, err)

		// TODO: retrieve customer, make sure it matches

	})

	outerCustomer, err = db.Customer("Compcode")
	assert.NoError(t, err)

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

		// TODO: retrieve seller, make sure it matches
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
			SellerID:      outerSeller.SellerID,
		}

		err = db.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		// TODO: retrieve dc, make sure it matches
		outerDatacenter, err = db.Datacenter(datacenter.ID)
		assert.NoError(t, err)
	})

	t.Run("AddBuyer", func(t *testing.T) {

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

		relay := routing.Relay{
			ID:           rid,
			Name:         "local.1",
			Addr:         *addr,
			PublicKey:    publicKey,
			UpdateKey:    updateKey,
			Datacenter:   outerDatacenter,
			MRC:          19700000000000,
			Overage:      26000000000000,
			BWRule:       routing.BWRuleBurst,
			ContractTerm: 12,
			StartDate:    time.Now(),
			EndDate:      time.Now(),
			Type:         routing.BareMetal,
			State:        routing.RelayStateMaintenance,
		}

		// fmt.Printf("AddRelay test relay: %s\n", relay.String())

		err = db.AddRelay(ctx, relay)
		assert.NoError(t, err)

		// check only the fields set above
		checkRelay, err := db.Relay(rid)
		assert.NoError(t, err)
		assert.Equal(t, relay.Name, checkRelay.Name)
		assert.Equal(t, relay.Addr, checkRelay.Addr)
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
		t.Skip()
		dcMap := routing.DatacenterMap{
			Alias:        "local.map",
			BuyerID:      outerBuyer.ID,
			DatacenterID: outerDatacenter.ID,
		}

		err := db.AddDatacenterMap(ctx, dcMap)
		assert.NoError(t, err)

		checkDCMaps := db.GetDatacenterMapsForBuyer(outerBuyer.ID)
		assert.NotEqual(t, 1, len(checkDCMaps))
		assert.Equal(t, dcMap, checkDCMaps[0])

	})

}
