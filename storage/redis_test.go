package storage_test

import (
	"testing"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/storage"
)

func TestNewRedisClient(t *testing.T) {
	client := storage.NewRedisClient("localhost:9999")
	if _, ok := client.(*redis.Client); !ok {
		t.Error("client expected to be *redis.Client with a single host")
	}

	client = storage.NewRedisClient("localhost:9999", "localhost:8888", "localhost:7777")
	if _, ok := client.(*redis.Ring); !ok {
		t.Error("client expected to be *redis.Ring with a multiple hosts")
	}
}
