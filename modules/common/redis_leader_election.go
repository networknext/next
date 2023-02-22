package common

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sort"
	"sync"
	"time"

	// todo: switch to redigo!
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/networknext/backend/modules/core"
)

const RedisLeaderElectionVersion = 2 // IMPORTANT: bump this anytime you change the redis data structures!

type RedisLeaderElectionConfig struct {
	RedisHostname string
	RedisPassword string
	ServiceName   string
	Timeout       time.Duration
}

type RedisLeaderElection struct {
	config      RedisLeaderElectionConfig
	redisClient *redis.Client
	startTime   time.Time
	instanceId  string
	leaderInstanceId string

	leaderMutex sync.RWMutex
	isLeader    bool
}

type InstanceEntry struct {
	InstanceId string
	StartTime  uint64
}

func CreateRedisLeaderElection(ctx context.Context, config RedisLeaderElectionConfig) (*RedisLeaderElection, error) {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisHostname,
		Password: config.RedisPassword,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

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

func (leaderElection *RedisLeaderElection) Update(ctx context.Context) {

	// write

	instanceEntry := InstanceEntry{}
	instanceEntry.InstanceId = leaderElection.instanceId
	instanceEntry.StartTime = uint64(leaderElection.startTime.UnixNano())

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(instanceEntry)
	if err != nil {
		core.Error("failed to write instance entry\n")
	}

	instanceData := buffer.Bytes()

	timeoutContext, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	pipe := leaderElection.redisClient.TxPipeline()
	pipe.Set(timeoutContext, fmt.Sprintf("%s-instance-%d/%s", leaderElection.config.ServiceName, RedisLeaderElectionVersion, leaderElection.instanceId), instanceData[:], leaderElection.config.Timeout)
	cmds, err := pipe.Exec(timeoutContext)

	// todo: don't use "SCAN" it's slow

	// get all "instance/*" keys

	instanceKeys := []string{}
	itor := leaderElection.redisClient.Scan(timeoutContext, 0, fmt.Sprintf("%s-instance-%d/*", leaderElection.config.ServiceName, RedisLeaderElectionVersion), 0).Iterator()
	for itor.Next(timeoutContext) {
		instanceKeys = append(instanceKeys, itor.Val())
	}
	if err := itor.Err(); err != nil {
		core.Error("failed to get instance keys: %v", err)
		return
	}

	// query all instance data

	pipe = leaderElection.redisClient.Pipeline()
	for i := range instanceKeys {
		pipe.Get(timeoutContext, instanceKeys[i])
	}
	cmds, err = pipe.Exec(timeoutContext)

	if err != nil {
		core.Error("failed to get instance entries: %v", err)
		return
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
		instanceEntries = append(instanceEntries, instanceEntry)
	}

	// if there are no instance entries, we cannot be leader and there is no leader (yet)

	if len(instanceEntries) == 0 {
		return
	}

	// wait at least timeout to ensure we don't flap leader when a bunch of services start close together

	if time.Since(leaderElection.startTime) < leaderElection.config.Timeout {
		return
	}

	// leader is the one with longest uptime, instance id used for tie-break

	sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].StartTime < instanceEntries[j].StartTime })

	sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].InstanceId > instanceEntries[j].InstanceId })

	leaderInstance := instanceEntries[0]

	leaderElection.leaderInstanceId = leaderInstance.InstanceId

	leaderElection.leaderMutex.Lock()
	previousValue := leaderElection.isLeader
	currentValue := leaderInstance.InstanceId == leaderElection.instanceId
	leaderElection.isLeader = currentValue
	leaderElection.leaderMutex.Unlock()

	if !previousValue && currentValue {
		core.Log("we became the leader")
	} else if previousValue && !currentValue {
		core.Log("we are no longer the leader")
	}
}

func (leaderElection *RedisLeaderElection) Store(ctx context.Context, name string, data []byte) {
	timeoutContext, cancel := context.WithTimeout(ctx, time.Duration(10*time.Second))
	defer cancel()
	pipe := leaderElection.redisClient.TxPipeline()
	pipe.Set(timeoutContext, fmt.Sprintf("%s-data-%d/%s/%s", leaderElection.config.ServiceName, RedisLeaderElectionVersion, leaderElection.instanceId, name), data[:], leaderElection.config.Timeout)
	_, err := pipe.Exec(timeoutContext)
	if err != nil {
		core.Error("failed to store data '%s': %v", name, err)
		return
	}
}

func (leaderElection *RedisLeaderElection) Load(ctx context.Context, name string) []byte {
	timeoutContext, cancel := context.WithTimeout(ctx, time.Duration(10*time.Second))
	defer cancel()
	pipe := leaderElection.redisClient.Pipeline()
	pipe.Get(timeoutContext, fmt.Sprintf("%s-data-%d/%s/%s", leaderElection.config.ServiceName, RedisLeaderElectionVersion, leaderElection.leaderInstanceId, name))
	cmds, err := pipe.Exec(timeoutContext)
	if err != nil {
		return nil
	}
	return []byte(cmds[0].(*redis.StringCmd).Val())
}

func (leaderElection *RedisLeaderElection) IsLeader() bool {
	leaderElection.leaderMutex.RLock()
	value := leaderElection.isLeader
	leaderElection.leaderMutex.RUnlock()
	return value
}
