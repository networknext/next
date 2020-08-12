package storage_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/networknext/backend/storage"
	"github.com/stretchr/testify/assert"
)

func TestRawRedisClientPing(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	redisClient, err := storage.NewRawRedisClient(redisServer.Addr())
	assert.NoError(t, err)

	err = redisClient.Ping()
	assert.NoError(t, err)
}
