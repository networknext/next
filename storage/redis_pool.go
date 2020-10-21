package storage

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

func NewRedisPool(hostname string, maxIdleConnections int, maxActiveConnections int) *redis.Pool {
	pool := redis.Pool{
		MaxIdle:     maxIdleConnections,
		MaxActive:   maxActiveConnections,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", hostname)
		},
	}

	return &pool
}

func ValidateRedisPool(pool *redis.Pool) error {
	redisConn := pool.Get()
	defer redisConn.Close()

	redisConn.Send("PING")
	redisConn.Flush()
	pong, err := redisConn.Receive()
	if err != nil || pong != "PONG" {
		return fmt.Errorf("could not ping: %v", err)
	}

	return nil
}
