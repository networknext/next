package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/crypto"
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

	customer, err := db.Customer("Compcode")
	assert.NoError(t, err)

	t.Run("AddSeller", func(t *testing.T) {
		seller := routing.Seller{
			ID:                        "Compcode",
			IngressPriceNibblinsPerGB: 10,
			EgressPriceNibblinsPerGB:  20,
			CustomerID:                customer.CustomerID,
		}

		err = db.AddSeller(ctx, seller)
		assert.NoError(t, err)

		_, err := db.Seller("Compcode")
		assert.NoError(t, err)

		// TODO: retrieve seller, make sure it matches
	})

	t.Run("AddDatacenter", func(t *testing.T) {

		seller, err := db.Seller("Compcode")
		assert.NoError(t, err)

		datacenter := routing.Datacenter{
			ID:      crypto.HashID("datacenter name"),
			Name:    "datacenter.name",
			Enabled: true,
			Location: routing.Location{
				Latitude:  70.5,
				Longitude: 120.5,
			},
			StreetAddress: "Somewhere, USA",
			SellerID:      seller.SellerID,
		}

		err = db.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		// TODO: retrieve dc, make sure it matches
	})

	t.Run("AddBuyer", func(t *testing.T) {
		buyer := routing.Buyer{
			ID:         1,
			Live:       true,
			Debug:      false,
			PublicKey:  make([]byte, crypto.KeySize),
			CustomerID: customer.CustomerID,
		}

		err = db.AddBuyer(ctx, buyer)
		assert.NoError(t, err)
	})

}
