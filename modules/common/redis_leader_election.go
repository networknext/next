package common

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/networknext/next/modules/core"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const RedisLeaderElectionVersion = 1 // IMPORTANT: bump this anytime you change the redis data structures!

type RedisLeaderElectionConfig struct {
	RedisHostname string
	RedisPassword string
	ServiceName   string
	Timeout       time.Duration
}

type RedisLeaderElection struct {
	config           RedisLeaderElectionConfig
	redisClient      redis.Cmdable
	startTime        time.Time
	instanceId       string
	leaderInstanceId string

	leaderMutex sync.RWMutex
	isLeader    bool
	isReady     bool
}

type InstanceEntry struct {
	InstanceId string
	StartTime  uint64
	UpdateTime uint64
}

func CreateRedisLeaderElection(redisClient redis.Cmdable, config RedisLeaderElectionConfig) (*RedisLeaderElection, error) {

	leaderElection := &RedisLeaderElection{}

	if config.Timeout == 0 {
		config.Timeout = time.Second * 10
	}

	leaderElection.config = config
	leaderElection.redisClient = redisClient
	leaderElection.startTime = time.Now()
	leaderElection.instanceId = uuid.New().String()

	// core.Debug("redis leader election start time: %s", leaderElection.startTime)
	// core.Debug("redis leader election instance id: %s", leaderElection.instanceId)

	return leaderElection, nil
}

func (leaderElection *RedisLeaderElection) Start(ctx context.Context) {

	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				leaderElection.Update(ctx)
			}
		}
	}()
}

func getInstanceEntries(ctx context.Context, redisClient redis.Cmdable, service string, minutes int64) []InstanceEntry {

	// get all instance keys and values

	key_a := fmt.Sprintf("%s-instance-%d-%d", service, RedisLeaderElectionVersion, minutes)
	key_b := fmt.Sprintf("%s-instance-%d-%d", service, RedisLeaderElectionVersion, minutes-1)

	pipeline := redisClient.Pipeline()

	pipeline.HGetAll(ctx, key_a)
	pipeline.HGetAll(ctx, key_b)

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("get instance entries error: %v", err)
	}

	instances_a := cmds[0].(*redis.MapStringStringCmd).Val()

	instances_b := cmds[1].(*redis.MapStringStringCmd).Val()

	// merge instance entries

	instanceMap := make(map[string]string)

	for k, v := range instances_b {
		instanceMap[k] = v
	}

	for k, v := range instances_a {
		instanceMap[k] = v
	}

	// convert instance data to instance entries

	instanceEntries := []InstanceEntry{}

	for _, v := range instanceMap {
		instanceEntry := InstanceEntry{}
		buffer := bytes.NewBuffer([]byte(v))
		decoder := gob.NewDecoder(buffer)
		err := decoder.Decode(&instanceEntry)
		if err != nil {
			core.Debug("could not decode instance entry: %v", err)
			continue
		}
		instanceEntries = append(instanceEntries, instanceEntry)
	}

	// leader is the one with longest uptime, instance id used for tie-break

	sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].StartTime < instanceEntries[j].StartTime })

	sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].InstanceId > instanceEntries[j].InstanceId })

	return instanceEntries
}

func (leaderElection *RedisLeaderElection) Update(ctx context.Context) {

	// write our instance entry

	seconds := time.Now().Unix()
	minutes := seconds / 60

	instanceEntry := InstanceEntry{}
	instanceEntry.InstanceId = leaderElection.instanceId
	instanceEntry.StartTime = uint64(leaderElection.startTime.UnixNano())
	instanceEntry.UpdateTime = uint64(seconds)

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(instanceEntry)
	if err != nil {
		core.Error("failed to write instance entry\n")
	}

	instanceData := buffer.Bytes()

	key := fmt.Sprintf("%s-instance-%d-%d", leaderElection.config.ServiceName, RedisLeaderElectionVersion, minutes)

	field := instanceEntry.InstanceId
	value := instanceData

	leaderElection.redisClient.HSet(ctx, key, field, value)

	// wait at least timeout to ensure we don't flap leader when a bunch of services start close together

	if time.Since(leaderElection.startTime) < leaderElection.config.Timeout {
		// core.Debug("wait timeout\n")
		return
	}

	// get all instance entries for this service

	instanceEntries := getInstanceEntries(ctx, leaderElection.redisClient, leaderElection.config.ServiceName, minutes)

	// if there are no instance entries, we cannot be leader and there is no leader (yet)

	if len(instanceEntries) == 0 {
		core.Debug("no instance entries?\n")
		return
	}

	// leader is the first instance entry

	leaderInstance := instanceEntries[0]

	leaderElection.leaderInstanceId = leaderInstance.InstanceId

	leaderElection.leaderMutex.Lock()
	previousValue := leaderElection.isLeader
	currentValue := leaderInstance.InstanceId == leaderElection.instanceId
	leaderElection.isLeader = currentValue
	leaderElection.isReady = true
	leaderElection.leaderMutex.Unlock()

	if !previousValue && currentValue {
		core.Log("we became the leader")
	} else if previousValue && !currentValue {
		core.Log("we are no longer the leader")
	}
}

func (leaderElection *RedisLeaderElection) Store(ctx context.Context, name string, data []byte) {
	key := fmt.Sprintf("%s-instance-data-%d-%s-%s", leaderElection.config.ServiceName, RedisLeaderElectionVersion, leaderElection.instanceId, name)
	err := leaderElection.redisClient.Set(ctx, key, data, 0).Err()
	if err != nil {
		core.Error("failed to store data: %v", err)
	}
}

func (leaderElection *RedisLeaderElection) Load(ctx context.Context, name string) []byte {
	key := fmt.Sprintf("%s-instance-data-%d-%s-%s", leaderElection.config.ServiceName, RedisLeaderElectionVersion, leaderElection.leaderInstanceId, name)
	value, err := leaderElection.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil
	} else {
		return []byte(value)
	}
}

func (leaderElection *RedisLeaderElection) IsLeader() bool {
	leaderElection.leaderMutex.RLock()
	value := leaderElection.isLeader
	leaderElection.leaderMutex.RUnlock()
	return value
}

func (leaderElection *RedisLeaderElection) IsReady() bool {
	leaderElection.leaderMutex.RLock()
	value := leaderElection.isReady
	leaderElection.leaderMutex.RUnlock()
	return value
}

func LoadMasterServiceData(ctx context.Context, redisClient redis.Cmdable, service string, name string) []byte {
	seconds := time.Now().Unix()
	minutes := seconds / 60
	instanceEntries := getInstanceEntries(ctx, redisClient, service, minutes)
	if len(instanceEntries) == 0 {
		return nil
	}
	masterInstance := instanceEntries[0]
	key := fmt.Sprintf("%s-instance-data-%d-%s-%s", service, RedisLeaderElectionVersion, masterInstance.InstanceId, name)
	value, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil
	}
	return []byte(value)
}
