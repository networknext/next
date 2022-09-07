package common

import (
	"context"
	"time"

    "github.com/google/uuid"
	"github.com/go-redis/redis/v8"
	"github.com/networknext/backend/modules/core"
)

type RedisSelectorConfig struct {
	RedisHostname      string
	RedisPassword      string
}

type RedisSelector struct {
	config          RedisSelectorConfig
	redisDB         *redis.Client
	startTime       time.Time
	instanceId      string
	costMatrixData  []byte
	routeMatrixData []byte
}

func CreateRedisSelector(ctx context.Context, config RedisSelectorConfig) (*RedisSelector, error) {
	
	redisDB := redis.NewClient(&redis.Options{
		Addr:     config.RedisHostname,
		Password: config.RedisPassword,
	})
	_, err := redisDB.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	
	selector := &RedisSelector{}

	selector.config = config
	selector.redisDB = redisDB
	selector.startTime = time.Now()
	selector.instanceId = uuid.New().String()

	core.Debug("redis selector instance id: %s", selector.instanceId)

	return selector, nil
}

func (selector *RedisSelector) Store(costMatrixData []byte, routeMatrixData []byte) {
	// todo: actually do the redis thing
	selector.costMatrixData = costMatrixData
	selector.routeMatrixData = routeMatrixData
}

func (selector *RedisSelector) Load() ([]byte, []byte) {
	// todo: actually do the redis thing
	return selector.costMatrixData, selector.routeMatrixData
}
