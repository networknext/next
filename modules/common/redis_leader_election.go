package common

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

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

	leaderMutex sync.RWMutex
	isLeader    bool

	autoRefresh bool
}

type InstanceEntry struct {
	InstanceId string
	StartTime  uint64
	Keys       []string
}

type DataStoreConfig struct {
	Name string
	Data []byte
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

	core.Debug("redis leader election start time: %s", leaderElection.startTime)
	core.Debug("redis leader election instance id: %s", leaderElection.instanceId)

	return leaderElection, nil
}

func (leaderElection *RedisLeaderElection) Start(ctx context.Context) {

	leaderElection.autoRefresh = true

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
	leaderElection.Store(ctx)
	leaderElection.Load(ctx)
}

func (leaderElection *RedisLeaderElection) Store(ctx context.Context, dataStores ...DataStoreConfig) {

	instanceEntry := InstanceEntry{}
	instanceEntry.InstanceId = leaderElection.instanceId
	instanceEntry.StartTime = uint64(leaderElection.startTime.UnixNano())

	numStores := len(dataStores)
	instanceEntry.Keys = make([]string, numStores)

	for i := 0; i < numStores; i++ {
		instanceEntry.Keys[i] = fmt.Sprintf("%s-%d/%s", dataStores[i].Name, RedisLeaderElectionVersion, leaderElection.instanceId)
	}

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

	for i := 0; i < numStores; i++ {
		pipe.Set(timeoutContext, instanceEntry.Keys[i], dataStores[i].Data[:], leaderElection.config.Timeout)
	}

	_, err = pipe.Exec(timeoutContext)
	if err != nil {
		core.Error("failed to store instance data: %v", err)
		return
	}
}

func (leaderElection *RedisLeaderElection) Load(ctx context.Context) []DataStoreConfig {

	timeoutContext, cancel := context.WithTimeout(ctx, time.Duration(5*time.Second))
	defer cancel()

	// get all "instance/*" keys

	instanceKeys := []string{}
	dataStores := []DataStoreConfig{}
	itor := leaderElection.redisClient.Scan(timeoutContext, 0, fmt.Sprintf("%s-instance-%d/*", leaderElection.config.ServiceName, RedisLeaderElectionVersion), 0).Iterator()
	for itor.Next(timeoutContext) {
		instanceKeys = append(instanceKeys, itor.Val())
	}
	if err := itor.Err(); err != nil {
		core.Error("failed to get instance keys: %v", err)
		return dataStores
	}

	// query all instance data

	pipe := leaderElection.redisClient.Pipeline()
	for i := range instanceKeys {
		pipe.Get(timeoutContext, instanceKeys[i])
	}
	cmds, err := pipe.Exec(timeoutContext)

	if err != nil {
		core.Error("failed to get instance entries: %v", err)
		return dataStores
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

	// no instance entries? we have no instance data...

	if len(instanceEntries) == 0 {
		core.Error("no instance entries found")
		return dataStores
	}

	// IMPORTANT: if there is only one entry, wait at least 10 seconds to ensure
	// we don't flap leader when a bunch of services start close together

	if len(instanceEntries) == 1 && time.Since(leaderElection.startTime) < leaderElection.config.Timeout {
		core.Debug("only one instance entry. waiting for other entries to join...")
		return dataStores
	}

	sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].StartTime < instanceEntries[j].StartTime })

	sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].InstanceId > instanceEntries[j].InstanceId })

	masterInstance := instanceEntries[0]

	// are we the leader?

	leaderElection.leaderMutex.Lock()
	previousValue := leaderElection.isLeader
	currentValue := masterInstance.InstanceId == leaderElection.instanceId
	leaderElection.isLeader = currentValue
	leaderElection.leaderMutex.Unlock()

	if !previousValue && currentValue {
		core.Log("we became the leader")
	} else if previousValue && !currentValue {
		core.Log("we are no longer the leader")
	}

	core.Debug("master instance: %+v", masterInstance)

	// get data from master instance

	pipe = leaderElection.redisClient.Pipeline()

	dataStores = make([]DataStoreConfig, len(masterInstance.Keys))

	for i := 0; i < len(dataStores); i++ {
		key := masterInstance.Keys[i]
		dataStores[i].Name = strings.Split(key, "-")[0]
		pipe.Get(timeoutContext, key)
	}

	cmds, err = pipe.Exec(timeoutContext)
	if err != nil {
		core.Error("failed to get data from redis: %v", err)
		return dataStores
	}

	for i := 0; i < len(dataStores); i++ {
		dataStores[i].Data = []byte(cmds[i].(*redis.StringCmd).Val())
	}

	return dataStores
}

func (leaderElection *RedisLeaderElection) IsLeader() bool {
	leaderElection.leaderMutex.RLock()
	value := leaderElection.isLeader
	leaderElection.leaderMutex.RUnlock()
	return value
}
