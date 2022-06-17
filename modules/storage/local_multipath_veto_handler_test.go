package storage_test

// todo: this needs to be converted to a functional test

/*
import (
	"testing"

	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewLocalMultipathVetoHandlerSuccess(t *testing.T) {
	getDatabase := func() *routing.DatabaseBinWrapper {
		return &routing.DatabaseBinWrapper{}
	}

	_, err := storage.NewLocalMultipathVetoHandler("", getDatabase)
	assert.NoError(t, err)
}

func TestLocalMultipathVetoUserSuccess(t *testing.T) {
	getDatabase := func() *routing.DatabaseBinWrapper {
		buyerMap := make(map[uint64]routing.Buyer)
		buyerMap[123] = routing.Buyer{ID: 123, CompanyCode: "local"}

		return &routing.DatabaseBinWrapper{BuyerMap: buyerMap}
	}

	multipathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", getDatabase)
	assert.NoError(t, err)
	err = multipathVetoHandler.MultipathVetoUser("local", 1234567890)
	assert.NoError(t, err)
}

func TestLocalGetMapCopyNewCustomerCode(t *testing.T) {
	getDatabase := func() *routing.DatabaseBinWrapper {
		return &routing.DatabaseBinWrapper{}
	}

	multipathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", getDatabase)
	assert.NoError(t, err)

	multipathVetoMap := multipathVetoHandler.GetMapCopy("unknown")
	assert.Equal(t, map[uint64]bool{}, multipathVetoMap)
}

func TestLocalGetMapCopyExistingCustomerCode(t *testing.T) {
	getDatabase := func() *routing.DatabaseBinWrapper {
		buyerMap := make(map[uint64]routing.Buyer)
		buyerMap[123] = routing.Buyer{ID: 123, CompanyCode: "local"}

		return &routing.DatabaseBinWrapper{BuyerMap: buyerMap}
	}

	multipathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", getDatabase)
	assert.NoError(t, err)
	err = multipathVetoHandler.MultipathVetoUser("local", 1234567890)
	assert.NoError(t, err)

	multipathVetoMap := multipathVetoHandler.GetMapCopy("local")
	assert.Equal(t, map[uint64]bool{1234567890: true}, multipathVetoMap)
}
*/