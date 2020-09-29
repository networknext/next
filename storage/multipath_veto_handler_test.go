package storage_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewMultipathVetoHandlerCouldNotPing(t *testing.T) {
	_, err := storage.NewMultipathVetoHandler("0.1.2.3:6379", nil)
	assert.EqualError(t, err, "could not ping multipath veto redis instance: dial tcp 0.1.2.3:6379: connect: invalid argument")
}

func TestNewMultipathVetoHandlerSuccess(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	_, err = storage.NewMultipathVetoHandler(redisServer.Addr(), nil)
	assert.NoError(t, err)
}

func TestMultipathVetoUserRedisError(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{
		ID:          123,
		CompanyCode: "local",
	})
	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	redisServer.Close()

	err = multipathVetoHandler.MultipathVetoUser("local", 1234567890)
	assert.EqualError(t, err, fmt.Sprintf("failed setting multipath veto on user %016x for buyer %s: %v", 1234567890, "local", "EOF"))
}

func TestMultipathVetoUserSuccess(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{
		ID:          123,
		CompanyCode: "local",
	})
	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
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
	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), nil)
	assert.NoError(t, err)

	multipathVetoMap := multipathVetoHandler.GetMapCopy("unknown")
	assert.Equal(t, map[uint64]bool{}, multipathVetoMap)
}

func TestGetMapCopyExistingCustomerCode(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{
		ID:          123,
		CompanyCode: "local",
	})
	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)
	err = multipathVetoHandler.MultipathVetoUser("local", 1234567890)
	assert.NoError(t, err)

	multipathVetoMap := multipathVetoHandler.GetMapCopy("local")
	assert.Equal(t, map[uint64]bool{1234567890: true}, multipathVetoMap)
}

func TestMultipathVetoHandlerSyncRedisError(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{
		ID:          123,
		CompanyCode: "local",
	})
	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	redisServer.Close()

	err = multipathVetoHandler.Sync()
	assert.EqualError(t, err, "failed to get customer keys from multipath veto redis: EOF")
}

func TestMultipathVetoHandlerSyncParseError(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{
		ID:          123,
		CompanyCode: "local",
	})
	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	redisServer.Set("local-zyxwvut", "1")

	err = multipathVetoHandler.Sync()
	assert.EqualError(t, err, "failed to parse user hash zyxwvut: strconv.ParseUint: parsing \"zyxwvut\": invalid syntax")
}

func TestMultipathVetoHandlerSyncSuccess(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{
		ID:          123,
		CompanyCode: "local",
	})
	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
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
