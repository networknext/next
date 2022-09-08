package common

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/networknext/backend/modules/core"
)

type RedisSelectorConfig struct {
	RedisHostname string
	RedisPassword string
}

type RedisSelector struct {
	config          RedisSelectorConfig
	redisClient     *redis.Client
	startTime       time.Time
	instanceId      string
	relaysData      []byte
	costMatrixData  []byte
	routeMatrixData []byte
	storeCounter    uint64
}

type InstanceEntry struct {
	InstanceId     string
	Uptime         uint64
	Timestamp      uint64
	RelaysKey      string
	CostMatrixKey  string
	RouteMatrixKey string
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

func (selector *RedisSelector) Store(ctx context.Context, relaysData []byte, costMatrixData []byte, routeMatrixData []byte) {

	selector.storeCounter++

	instanceEntry := InstanceEntry{}
	instanceEntry.InstanceId = selector.instanceId
	instanceEntry.Uptime = uint64(time.Since(selector.startTime))
	instanceEntry.Timestamp = uint64(time.Now().Unix())
	instanceEntry.RelaysKey = fmt.Sprintf("relays/%s-%d", selector.instanceId, selector.storeCounter)
	instanceEntry.CostMatrixKey = fmt.Sprintf("cost_matrix/%s-%d", selector.instanceId, selector.storeCounter)
	instanceEntry.RouteMatrixKey = fmt.Sprintf("route_matrix/%s-%d", selector.instanceId, selector.storeCounter)

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(instanceEntry)
	if err != nil {
		core.Error("failed to write instance entry\n")
	}
	instanceData := buffer.Bytes()

	timeoutContext, _ := context.WithTimeout(ctx, time.Duration(time.Second))

	pipe := selector.redisClient.TxPipeline()
	pipe.Set(timeoutContext, fmt.Sprintf("instance/%s", selector.instanceId), instanceData[:], 10*time.Second)
	pipe.Set(timeoutContext, instanceEntry.RelaysKey, relaysData[:], 10*time.Second)
	pipe.Set(timeoutContext, instanceEntry.CostMatrixKey, costMatrixData[:], 10*time.Second)
	pipe.Set(timeoutContext, instanceEntry.RouteMatrixKey, routeMatrixData[:], 10*time.Second)
	_, err = pipe.Exec(timeoutContext)

	if err != nil {
		core.Error("failed to store instance data: %v", err)
		return
	}
}

func (selector *RedisSelector) Load(ctx context.Context) ([]byte, []byte, []byte) {

	timeoutContext, _ := context.WithTimeout(ctx, time.Duration(time.Second))

	// get all "instance/*" keys via scan to be safe

	instanceKeys := []string{}
	itor := selector.redisClient.Scan(timeoutContext, 0, "instance/*", 0).Iterator()
	for itor.Next(timeoutContext) {
		instanceKeys = append(instanceKeys, itor.Val())
	}
	if err := itor.Err(); err != nil {
		core.Error("failed to get instance keys: %v", err)
		return nil, nil, nil
	}

	// query all instance data

	pipe := selector.redisClient.Pipeline()
	for i := range instanceKeys {
		pipe.Get(timeoutContext, instanceKeys[i])
	}
	cmds, err := pipe.Exec(timeoutContext)

	if err != nil {
		core.Error("failed to get instance entries: %v", err)
		return nil, nil, nil
	}

	// convert instance data to instance entries

	instanceEntries := []InstanceEntry{}

	for _, cmd := range cmds {

		instanceData := cmd.(*redis.StringCmd).Val()

		instanceEntry := InstanceEntry{}
		buffer := bytes.NewBuffer([]byte(instanceData))
		decoder := gob.NewDecoder(buffer)
		err := decoder.Decode(&instanceEntry)
		if err != nil {
			core.Debug("could not decode instance entry: %v", err)
			continue
		}

		// IMPORTANT: ignore any instance entries more than 5 seconds old
		if instanceEntry.Timestamp >= uint64(time.Now().Unix()-5) {
			instanceEntries = append(instanceEntries, instanceEntry)
		}
	}

	// no instance entries? we have no route matrix data...

	if len(instanceEntries) == 0 {
		core.Error("no instance entries found")
		return nil, nil, nil
	}

	// select master instance (most uptime, instance id as tie breaker)

	sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].InstanceId > instanceEntries[j].InstanceId })

	sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].Uptime > instanceEntries[j].Uptime })

	masterInstance := instanceEntries[0]

	// get cost matrix and route matrix for master instance

	pipe = selector.redisClient.Pipeline()
	pipe.Get(timeoutContext, masterInstance.RelaysKey)
	pipe.Get(timeoutContext, masterInstance.CostMatrixKey)
	pipe.Get(timeoutContext, masterInstance.RouteMatrixKey)
	cmds, err = pipe.Exec(timeoutContext)

	if err != nil {
		core.Error("failed to get data from redis: %v", err)
		return nil, nil, nil
	}

	selector.relaysData = []byte(cmds[0].(*redis.StringCmd).Val())
	selector.costMatrixData = []byte(cmds[1].(*redis.StringCmd).Val())
	selector.routeMatrixData = []byte(cmds[2].(*redis.StringCmd).Val())

	return selector.relaysData, selector.costMatrixData, selector.routeMatrixData
}
