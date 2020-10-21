package storage_test

import (
	"context"
	"fmt"
	"os"
	"testing"

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

	// NewSQLStorage syncs the local sync number from the remote
	db, err := storage.NewSQLStorage(ctx, logger)
	assert.NoError(t, err)

	t.Run("Sync", func(t *testing.T) {
		sync, _, err := db.CheckSequenceNumber(ctx)
		assert.NoError(t, err)
		assert.Equal(t, true, sync)
	})

	t.Run("Do Not Sync", func(t *testing.T) {
		// CheckSequenceNumber() (above) syncs local to remote
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

	// Due to the foreign key constraints, we can't add a
	// datacenter if there isn't a relevant extant seller, and
	// we can't add a seller if there isn't a relevant extant
	// customer

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
			IngressPriceNibblinsPerGB: 10,
			EgressPriceNibblinsPerGB:  20,
		}

		err = db.AddSeller(ctx, seller)
		assert.NoError(t, err)
	})

	t.Run("AddDatacenter", func(t *testing.T) {

		sellers := db.Sellers()
		if len(sellers) < 1 {
			assert.Error(t, fmt.Errorf("no sellers returned"))
		}

		// id := sellers[0]

	})

}
