package common

import (
	"context"
	"time"
	"fmt"
	"bytes"
	"encoding/gob"
	
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
	redisClient     *redis.Client
	startTime       time.Time
	instanceId      string
	costMatrixData  []byte
	routeMatrixData []byte
	storeCounter    uint64
}

type InstanceEntry struct {
	instanceId string
	uptime uint64
	timestamp uint64
	costMatrixKey string
	routeMatrixKey string
}

func CreateRedisSelector(ctx context.Context, config RedisSelectorConfig) (*RedisSelector, error) {
	
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisHostname,
		Password: config.RedisPassword,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	
	selector := &RedisSelector{}

	selector.config = config
	selector.redisClient = redisClient
	selector.startTime = time.Now()
	selector.instanceId = uuid.New().String()

	core.Debug("redis selector instance id: %s", selector.instanceId)

	return selector, nil
}

func (selector *RedisSelector) Store(ctx context.Context, costMatrixData []byte, routeMatrixData []byte) {

	selector.storeCounter++

	instanceEntry := InstanceEntry{}
	instanceEntry.instanceId = selector.instanceId
	instanceEntry.uptime = uint64(time.Since(selector.startTime))
	instanceEntry.timestamp = uint64(time.Now().Unix())
	instanceEntry.costMatrixKey = fmt.Sprintf("cost_matrix/%s-%d", selector.instanceId, selector.storeCounter)
	instanceEntry.routeMatrixKey = fmt.Sprintf("route_matrix/%s-%d", selector.instanceId, selector.storeCounter)

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(instanceEntry)
	instanceData := buffer.Bytes()

	timeoutContext, _ := context.WithTimeout(ctx, time.Duration(time.Second))

	pipe := selector.redisClient.TxPipeline()
	pipe.Set(timeoutContext, fmt.Sprintf("instance/%s", selector.instanceId), instanceData[:], 10*time.Second)
	pipe.Set(timeoutContext, instanceEntry.costMatrixKey, costMatrixData[:], 10*time.Second)
	pipe.Set(timeoutContext, instanceEntry.routeMatrixKey, routeMatrixData[:], 10*time.Second)
	_, err := pipe.Exec(timeoutContext)

	if err != nil {
		core.Error("failed to store instance data: %v", err)
		return
	}

	// todo: remove this once load is implemented
	selector.costMatrixData = costMatrixData
	selector.routeMatrixData = routeMatrixData
}

func (selector *RedisSelector) Load(ctx context.Context) ([]byte, []byte) {
	// todo: actually do the redis thing
	return selector.costMatrixData, selector.routeMatrixData
}
