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

const RedisLeaderElectionVersion = 0 // IMPORTANT: bump this anytime you change the redis data structures!

type RedisLeaderElectionConfig struct {
    RedisHostname string
    RedisPassword string
    ServiceName   string
}

type RedisLeaderElection struct {
    config      RedisLeaderElectionConfig
    redisClient *redis.Client
    startTime   time.Time
    instanceId  string
    isLeader    bool
}

type RedisLeaderElectionEntry struct {
    InstanceId string
    Uptime     uint64
    Timestamp  uint64
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

    leaderElection.config = config
    leaderElection.redisClient = redisClient
    leaderElection.startTime = time.Now()
    leaderElection.instanceId = uuid.New().String()

    core.Debug("redis leader election instance id: %s", leaderElection.instanceId)

    return leaderElection, nil
}

func (leaderElection *RedisLeaderElection) Update(ctx context.Context) {

    // store our instance entry

    instanceEntry := InstanceEntry{}
    instanceEntry.InstanceId = leaderElection.instanceId
    instanceEntry.Uptime = uint64(time.Since(leaderElection.startTime))
    instanceEntry.Timestamp = uint64(time.Now().Unix())

    var buffer bytes.Buffer
    encoder := gob.NewEncoder(&buffer)
    err := encoder.Encode(instanceEntry)
    if err != nil {
        core.Error("failed to write instance entry\n")
        leaderElection.isLeader = false
        return
    }
    instanceData := buffer.Bytes()

    timeoutContext, _ := context.WithTimeout(ctx, time.Duration(time.Second))

    pipe := leaderElection.redisClient.TxPipeline()
    pipe.Set(timeoutContext, fmt.Sprintf("leader-election-%s-%d/%s", leaderElection.config.ServiceName, RedisLeaderElectionVersion, leaderElection.instanceId), instanceData[:], 10*time.Second)
    _, err = pipe.Exec(timeoutContext)

    if err != nil {
        core.Error("failed to store instance data: %v", err)
        leaderElection.isLeader = false
        return
    }

    // get all "instance/*" keys via scan to be safe

    instanceKeys := []string{}
    itor := leaderElection.redisClient.Scan(timeoutContext, 0, fmt.Sprintf("leader-election-%s-%d/*", leaderElection.config.ServiceName, RedisLeaderElectionVersion), 0).Iterator()
    for itor.Next(timeoutContext) {
        instanceKeys = append(instanceKeys, itor.Val())
    }
    if err := itor.Err(); err != nil {
        core.Error("failed to get instance keys: %v", err)
        leaderElection.isLeader = false
        return
    }

    // query all instance data

    pipe = leaderElection.redisClient.Pipeline()
    for i := range instanceKeys {
        pipe.Get(timeoutContext, instanceKeys[i])
    }
    cmds, err := pipe.Exec(timeoutContext)

    if err != nil {
        core.Error("failed to get instance entries: %v", err)
        leaderElection.isLeader = false
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

        // IMPORTANT: ignore any instance entries more than 5 seconds old
        if instanceEntry.Timestamp >= uint64(time.Now().Unix()-5) {
            instanceEntries = append(instanceEntries, instanceEntry)
        }
    }

    // no instance entries? we are not the leader

    if len(instanceEntries) == 0 {
        core.Error("no instance entries found")
        leaderElection.isLeader = false
        return
    }

    // select master instance (most uptime, instance id as tie breaker)

    sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].InstanceId > instanceEntries[j].InstanceId })

    sort.SliceStable(instanceEntries, func(i, j int) bool { return instanceEntries[i].Uptime > instanceEntries[j].Uptime })

    masterInstance := instanceEntries[0]

    newLeader := (masterInstance.InstanceId == leaderElection.instanceId)

    if newLeader && !leaderElection.isLeader {
        core.Log("we are the leader")
    }

    if !newLeader && leaderElection.isLeader {
        core.Log("we are no longer the leader")
    }

    leaderElection.isLeader = newLeader
}

func (leaderElection *RedisLeaderElection) IsLeader() bool {
    return leaderElection.isLeader
}
