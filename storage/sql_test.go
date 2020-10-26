package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
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
	})

	t.Run("AddSeller", func(t *testing.T) {
		seller := routing.Seller{
			CompanyCode:               "Compcode",
			ID:                        "sellerID",
			IngressPriceNibblinsPerGB: 10,
			EgressPriceNibblinsPerGB:  20,
		}

		err = db.AddSeller(ctx, seller)
		assert.NoError(t, err)

		newSeller, err := db.Seller("sellerID")
		fmt.Printf("newSeller: %v\n", newSeller)
		assert.NoError(t, err)
	})

	t.Run("AddDatacenter", func(t *testing.T) {

		sellers := db.Sellers()
		if len(sellers) < 1 {
			assert.Error(t, fmt.Errorf("no sellers returned"))
		}

		// id := sellers[0]

	})

	// t.Run("syncDatacenters", func(t *testing.T) {
	// 	err =
	// })

}
