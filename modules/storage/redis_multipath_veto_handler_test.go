package storage_test

// todo: these need to be converted to functional tests

/*
import (
	"fmt"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisMultipathVetoHandlerCouldNotPing(t *testing.T) {
	_, err := storage.NewRedisMultipathVetoHandler("", "", 0, 0, nil)
	assert.Contains(t, err.Error(), "could not ping multipath veto redis instance")
}

func TestNewRedisMultipathVetoHandlerSuccess(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	getDatabase := func() *routing.DatabaseBinWrapper {
		return &routing.DatabaseBinWrapper{}
	}

	_, err = storage.NewRedisMultipathVetoHandler(redisServer.Addr(), "", 1, 1, getDatabase)
	assert.NoError(t, err)
}

func TestMultipathVetoUserRedisError(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	getDatabase := func() *routing.DatabaseBinWrapper {
		buyerMap := make(map[uint64]routing.Buyer)
		buyerMap[123] = routing.Buyer{ID: 123, CompanyCode: "local"}

		return &routing.DatabaseBinWrapper{BuyerMap: buyerMap}
	}

	multipathVetoHandler, err := storage.NewRedisMultipathVetoHandler(redisServer.Addr(), "", 1, 1, getDatabase)
	assert.NoError(t, err)

	redisServer.Close()

	err = multipathVetoHandler.MultipathVetoUser("local", 1234567890)
	assert.EqualError(t, err, fmt.Sprintf("failed setting multipath veto on user %016x for buyer %s: %v", 1234567890, "local", "EOF"))
}

func TestMultipathVetoUserSuccess(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	getDatabase := func() *routing.DatabaseBinWrapper {
		buyerMap := make(map[uint64]routing.Buyer)
		buyerMap[123] = routing.Buyer{ID: 123, CompanyCode: "local"}

		return &routing.DatabaseBinWrapper{BuyerMap: buyerMap}
	}

	multipathVetoHandler, err := storage.NewRedisMultipathVetoHandler(redisServer.Addr(), "", 1, 1, getDatabase)
	assert.NoError(t, err)
	err = multipathVetoHandler.MultipathVetoUser("local", 1234567890)
	assert.NoError(t, err)

	result, err := redisServer.Get(fmt.Sprintf("local-%016x", 1234567890))
	assert.NoError(t, err)

	assert.Equal(t, "1", result)
}

func TestGetMapCopyNewCustomerCode(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	getDatabase := func() *routing.DatabaseBinWrapper {
		return &routing.DatabaseBinWrapper{}
	}

	multipathVetoHandler, err := storage.NewRedisMultipathVetoHandler(redisServer.Addr(), "", 1, 1, getDatabase)
	assert.NoError(t, err)

	multipathVetoMap := multipathVetoHandler.GetMapCopy("unknown")
	assert.Equal(t, map[uint64]bool{}, multipathVetoMap)
}

func TestGetMapCopyExistingCustomerCode(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	getDatabase := func() *routing.DatabaseBinWrapper {
		buyerMap := make(map[uint64]routing.Buyer)
		buyerMap[123] = routing.Buyer{ID: 123, CompanyCode: "local"}

		return &routing.DatabaseBinWrapper{BuyerMap: buyerMap}
	}

	multipathVetoHandler, err := storage.NewRedisMultipathVetoHandler(redisServer.Addr(), "", 1, 1, getDatabase)
	assert.NoError(t, err)
	err = multipathVetoHandler.MultipathVetoUser("local", 1234567890)
	assert.NoError(t, err)

	multipathVetoMap := multipathVetoHandler.GetMapCopy("local")
	assert.Equal(t, map[uint64]bool{1234567890: true}, multipathVetoMap)
}

func TestMultipathVetoHandlerSyncRedisError(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	getDatabase := func() *routing.DatabaseBinWrapper {
		buyerMap := make(map[uint64]routing.Buyer)
		buyerMap[123] = routing.Buyer{ID: 123, CompanyCode: "local"}

		return &routing.DatabaseBinWrapper{BuyerMap: buyerMap}
	}

	multipathVetoHandler, err := storage.NewRedisMultipathVetoHandler(redisServer.Addr(), "", 1, 1, getDatabase)
	assert.NoError(t, err)

	redisServer.Close()

	err = multipathVetoHandler.Sync()
	assert.EqualError(t, err, "failed to get customer keys from multipath veto redis: EOF")
}

func TestMultipathVetoHandlerSyncParseError(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	getDatabase := func() *routing.DatabaseBinWrapper {
		buyerMap := make(map[uint64]routing.Buyer)
		buyerMap[123] = routing.Buyer{ID: 123, CompanyCode: "local"}

		return &routing.DatabaseBinWrapper{BuyerMap: buyerMap}
	}

	multipathVetoHandler, err := storage.NewRedisMultipathVetoHandler(redisServer.Addr(), "", 1, 1, getDatabase)
	assert.NoError(t, err)

	redisServer.Set("local-zyxwvut", "1")

	err = multipathVetoHandler.Sync()
	assert.EqualError(t, err, "failed to parse user hash zyxwvut: strconv.ParseUint: parsing \"zyxwvut\": invalid syntax")
}

func TestMultipathVetoHandlerSyncSuccess(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	getDatabase := func() *routing.DatabaseBinWrapper {
		buyerMap := make(map[uint64]routing.Buyer)
		buyerMap[123] = routing.Buyer{ID: 123, CompanyCode: "local"}

		return &routing.DatabaseBinWrapper{BuyerMap: buyerMap}
	}

	multipathVetoHandler, err := storage.NewRedisMultipathVetoHandler(redisServer.Addr(), "", 1, 1, getDatabase)
	assert.NoError(t, err)

	for i := 0; i < 100; i++ {
		redisServer.Set(fmt.Sprintf("local-%016x", i), "1")
	}

	err = multipathVetoHandler.Sync()
	assert.NoError(t, err)

	multipathVetoMap := multipathVetoHandler.GetMapCopy("local")
	for i := uint64(0); i < 100; i++ {
		assert.Equal(t, true, multipathVetoMap[i])
	}
}
*/