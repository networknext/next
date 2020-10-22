package storage_test

import (
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/networknext/backend/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisPool(t *testing.T) {
	addr := "127.0.0.1:6739"
	maxIdleConnections := 5
	maxActiveConnections := 64

	pool := storage.NewRedisPool(addr, maxIdleConnections, maxActiveConnections)
	assert.Equal(t, maxIdleConnections, pool.MaxIdle)
	assert.Equal(t, maxActiveConnections, pool.MaxActive)
}

func TestValidateRedisPoolFailure(t *testing.T) {
	addr := ""
	maxIdleConnections := 5
	maxActiveConnections := 64

	pool := storage.NewRedisPool(addr, maxIdleConnections, maxActiveConnections)
	err := storage.ValidateRedisPool(pool)
	assert.Error(t, err)
}

func TestValidateRedisPoolSuccess(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	maxIdleConnections := 5
	maxActiveConnections := 64

	pool := storage.NewRedisPool(redisServer.Addr(), maxIdleConnections, maxActiveConnections)
	err = storage.ValidateRedisPool(pool)
	assert.NoError(t, err)
}
