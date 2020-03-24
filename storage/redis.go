package storage

import (
	"fmt"

	"github.com/go-redis/redis/v7"
)

func NewRedisClient(addrs ...string) redis.UniversalClient {
	if len(addrs) == 1 {
		return redis.NewClient(&redis.Options{Addr: addrs[0]})
	}

	opts := redis.RingOptions{
		Addrs: make(map[string]string),
	}

	for _, addr := range addrs {
		opts.Addrs[fmt.Sprintf("shard-%s", addr)] = addr
	}

	return redis.NewRing(&opts)
}
