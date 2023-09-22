package common

// todo
/*
import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/networknext/next/modules/core"
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
	pool             *redis.Pool
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

func CreateRedisLeaderElection(pool *redis.Pool, config RedisLeaderElectionConfig) (*RedisLeaderElection, error) {

	leaderElection := &RedisLeaderElection{}

	if config.Timeout == 0 {
		config.Timeout = time.Second * 10
	}

	leaderElection.config = config
	leaderElection.pool = pool
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

func getInstanceEntries(pool *redis.Pool, service string, minutes int64) []InstanceEntry {

	// get all instance keys and values

	redisClient := pool.Get()

	key_a := fmt.Sprintf("%s-instance-%d-%d", service, RedisLeaderElectionVersion, minutes)
	key_b := fmt.Sprintf("%s-instance-%d-%d", service, RedisLeaderElectionVersion, minutes-1)

	redisClient.Send("HGETALL", key_a)
	redisClient.Send("HGETALL", key_b)

	redisClient.Flush()

	instances_a, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get instances a failed: %v", err)
		return nil
	}

	instances_b, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get instances b failed: %v", err)
		return nil
	}

	redisClient.Close()

	// merge instance entries

	instanceMap := make(map[string]string)

	for i := 0; i < len(instances_b); i += 2 {
		instanceMap[instances_b[i]] = instances_b[i+1]
	}

	for i := 0; i < len(instances_a); i += 2 {
		instanceMap[instances_a[i]] = instances_a[i+1]
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

	redisClient := leaderElection.pool.Get()

	defer redisClient.Close()

	key := fmt.Sprintf("%s-instance-%d-%d", leaderElection.config.ServiceName, RedisLeaderElectionVersion, minutes)

	field := instanceEntry.InstanceId
	value := instanceData

	redisClient.Send("HSET", key, field, value)

	redisClient.Flush()

	redisClient.Receive()

	redisClient.Close()

	// wait at least timeout to ensure we don't flap leader when a bunch of services start close together

	if time.Since(leaderElection.startTime) < leaderElection.config.Timeout {
		// core.Debug("wait timeout\n")
		return
	}

	// get all instance entries for this service

	instanceEntries := getInstanceEntries(leaderElection.pool, leaderElection.config.ServiceName, minutes)

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

	redisClient := leaderElection.pool.Get()

	defer redisClient.Close()

	key := fmt.Sprintf("%s-instance-data-%d-%s-%s", leaderElection.config.ServiceName, RedisLeaderElectionVersion, leaderElection.instanceId, name)

	redisClient.Send("SET", key, data)

	redisClient.Flush()

	redisClient.Receive()
}

func (leaderElection *RedisLeaderElection) Load(ctx context.Context, name string) []byte {

	redisClient := leaderElection.pool.Get()

	defer redisClient.Close()

	key := fmt.Sprintf("%s-instance-data-%d-%s-%s", leaderElection.config.ServiceName, RedisLeaderElectionVersion, leaderElection.leaderInstanceId, name)

	redisClient.Send("GET", key)

	redisClient.Flush()

	value, err := redis.String(redisClient.Receive())
	if err != nil {
		return nil
	}

	return []byte(value)
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

func LoadMasterServiceData(pool *redis.Pool, service string, name string) []byte {
	seconds := time.Now().Unix()
	minutes := seconds / 60
	instanceEntries := getInstanceEntries(pool, service, minutes)
	if len(instanceEntries) == 0 {
		return nil
	}
	masterInstance := instanceEntries[0]
	redisClient := pool.Get()
	defer redisClient.Close()
	key := fmt.Sprintf("%s-instance-data-%d-%s-%s", service, RedisLeaderElectionVersion, masterInstance.InstanceId, name)
	redisClient.Send("GET", key)
	redisClient.Flush()
	value, err := redis.String(redisClient.Receive())
	if err != nil {
		return nil
	}
	return []byte(value)
}
*/