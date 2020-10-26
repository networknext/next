package storage_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/networknext/backend/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewRawRedisClientFailure(t *testing.T) {
	_, err := storage.NewRawRedisClient("")
	assert.Error(t, err)
}

func TestNewRawRedisClientSuccess(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	_, err = storage.NewRawRedisClient(redisServer.Addr())
	assert.NoError(t, err)
}

func TestNewRawRedisClientPingFailure(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	redisClient, err := storage.NewRawRedisClient(redisServer.Addr())
	assert.NoError(t, err)

	// Wait a little bit here to prevent a race condition within miniredis
	time.Sleep(time.Millisecond)

	redisServer.Close()

	err = redisClient.Ping()
	assert.Error(t, err)
}

func TestNewRawRedisClientPingSuccess(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	redisClient, err := storage.NewRawRedisClient(redisServer.Addr())
	assert.NoError(t, err)

	err = redisClient.Ping()
	assert.NoError(t, err)
}

func TestNewRawRedisClientCommandFailure(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	redisClient, err := storage.NewRawRedisClient(redisServer.Addr())
	assert.NoError(t, err)

	redisClient.Close()

	key := "my-key"
	value := "my-value"

	err = redisClient.Command("SET", "%s %s", key, value)
	assert.Error(t, err)
}

func TestNewRawRedisClientCommandSuccess(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	redisClient, err := storage.NewRawRedisClient(redisServer.Addr())
	assert.NoError(t, err)

	key := "my-key"
	expectedValue := "my-value"

	err = redisClient.Command("SET", "%s %s", key, expectedValue)
	assert.NoError(t, err)

	// Wait a little bit here to prevent a race condition within miniredis
	time.Sleep(time.Millisecond)

	redisServer.Close()

	actualValue, err := redisServer.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, actualValue)
}

func TestNewRawRedisClientClose(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	redisClient, err := storage.NewRawRedisClient(redisServer.Addr())
	assert.NoError(t, err)

	err = redisClient.Close()
	assert.NoError(t, err)

	err = redisClient.Close()
	assert.Error(t, err)
}
