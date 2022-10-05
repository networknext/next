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

const RedisSelectorVersion = 0 // IMPORTANT: bump this anytime you change the redis data structures!

type RedisSelectorConfig struct {
    RedisHostname string
    RedisPassword string
    Timeout       time.Duration
}

type RedisSelector struct {
    config       RedisSelectorConfig
    redisClient  *redis.Client
    startTime    time.Time
    instanceId   string
    storeCounter uint64

    leaderMutex sync.RWMutex
    isLeader    bool
}

type InstanceEntry struct {
    InstanceId string
    Uptime     uint64
    Timestamp  uint64
    Keys       []string
}

type DataStoreConfig struct {
    Name string
    Data []byte
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

func (selector *RedisSelector) Store(ctx context.Context, dataStores []DataStoreConfig) {

    selector.storeCounter++

    instanceEntry := InstanceEntry{}
    instanceEntry.InstanceId = selector.instanceId
    instanceEntry.Uptime = uint64(time.Since(selector.startTime))
    instanceEntry.Timestamp = uint64(time.Now().Unix())

    numStores := len(dataStores)
    instanceEntry.Keys = make([]string, numStores)

    for i := 0; i < numStores; i++ {
        instanceEntry.Keys[i] = fmt.Sprintf("%s-%d/%s-%d", dataStores[i].Name, RedisSelectorVersion, selector.instanceId, selector.storeCounter)
    }

    var buffer bytes.Buffer
    encoder := gob.NewEncoder(&buffer)
    err := encoder.Encode(instanceEntry)
    if err != nil {
        core.Error("failed to write instance entry\n")
    }
    instanceData := buffer.Bytes()

    timeoutContext, _ := context.WithTimeout(ctx, time.Duration(time.Second))

    pipe := selector.redisClient.TxPipeline()
    pipe.Set(timeoutContext, fmt.Sprintf("instance-%d/%s", RedisSelectorVersion, selector.instanceId), instanceData[:], selector.config.Timeout)

    for i := 0; i < numStores; i++ {
        pipe.Set(timeoutContext, instanceEntry.Keys[i], dataStores[i].Data[:], selector.config.Timeout)
    }

    _, err = pipe.Exec(timeoutContext)
    if err != nil {
        core.Error("failed to store instance data: %v", err)
        return
    }
}

func (selector *RedisSelector) Load(ctx context.Context) []DataStoreConfig {

    timeoutContext, _ := context.WithTimeout(ctx, time.Duration(time.Second))

    // get all "instance/*" keys via scan to be safe

    instanceKeys := []string{}
    dataStores := []DataStoreConfig{}
    itor := selector.redisClient.Scan(timeoutContext, 0, fmt.Sprintf("instance-%d/*", RedisSelectorVersion), 0).Iterator()
    for itor.Next(timeoutContext) {
        key := itor.Val()
        instanceKeys = append(instanceKeys, key)
    }
    if err := itor.Err(); err != nil {
        core.Error("failed to get instance keys: %v", err)
        return dataStores

    }

    // query all instance data

    pipe := selector.redisClient.Pipeline()
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

        // IMPORTANT: ignore any instance entries more than 5 seconds old
        if instanceEntry.Timestamp >= uint64(time.Now().Unix()-5) {
            instanceEntries = append(instanceEntries, instanceEntry)
        }
    }

    // no instance entries? we have no route matrix data...

    if len(instanceEntries) == 0 {
        core.Error("no instance entries found")
        return dataStores
    }

    // select master instance (most uptime, instance id as tie breaker)

    sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].InstanceId > instanceEntries[j].InstanceId })

    sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].Uptime > instanceEntries[j].Uptime })

    masterInstance := instanceEntries[0]

    // are we the leader?

    selector.leaderMutex.Lock()
    previousValue := selector.isLeader
    currentValue := masterInstance.InstanceId == selector.instanceId
    selector.isLeader = currentValue
    selector.leaderMutex.Unlock()

    if !previousValue && currentValue {
        core.Log("we became the leader")
    } else if previousValue && !currentValue {
        core.Log("we are no longer the leader")
    }

    // get data for master instance

    pipe = selector.redisClient.Pipeline()

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

func (selector *RedisSelector) IsLeader() bool {
    selector.leaderMutex.RLock()
    value := selector.isLeader
    selector.leaderMutex.RUnlock()
    return value
}
