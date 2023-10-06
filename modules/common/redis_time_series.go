package common

import (
	"context"
	"sync"
	"time"

	"github.com/networknext/next/modules/core"

	"github.com/redis/go-redis/v9"
)

type RedisTimeSeriesConfig struct {
	RedisHostname      string
	RedisCluster       []string
	BatchSize          int
	BatchDuration      time.Duration
	MessageChannelSize int
}

// -------------------------------------------------------------------------------

type RedisTimeSeriesMessage struct {
	Timestamp uint64
	Keys      []string
	Values    []float64
}

type RedisTimeSeriesPublisher struct {
	config             RedisTimeSeriesConfig
	redisClient        *redis.Client
	redisClusterClient *redis.ClusterClient
	mutex              sync.Mutex
	messageBatch       []*RedisTimeSeriesMessage
	numMessagesSent    int
	numBatchesSent     int
	MessageChannel     chan *RedisTimeSeriesMessage
}

func CreateRedisTimeSeriesPublisher(ctx context.Context, config RedisTimeSeriesConfig) (*RedisTimeSeriesPublisher, error) {

	var client *redis.Client
	var clusterClient *redis.ClusterClient
	if len(config.RedisCluster) > 0 {
		clusterClient = CreateRedisClusterClient(config.RedisCluster)
		_, err := clusterClient.Ping(ctx).Result()
		if err != nil {
			return nil, err
		}
	} else {
		client = CreateRedisClient(config.RedisHostname)
		_, err := client.Ping(ctx).Result()
		if err != nil {
			return nil, err
		}
	}

	publisher := &RedisTimeSeriesPublisher{}

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 1024 * 1024
	}

	if config.BatchDuration == 0 {
		config.BatchDuration = time.Second
	}

	if config.BatchSize == 0 {
		config.BatchSize = 10000
	}

	publisher.config = config
	publisher.MessageChannel = make(chan *RedisTimeSeriesMessage, config.MessageChannelSize)
	publisher.redisClient = client
	publisher.redisClusterClient = clusterClient

	go publisher.updateMessageChannel(ctx)

	return publisher, nil
}

func (publisher *RedisTimeSeriesPublisher) updateMessageChannel(ctx context.Context) {
	ticker := time.NewTicker(publisher.config.BatchDuration)
	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:
			if len(publisher.messageBatch) > 0 {
				publisher.sendBatch(ctx)
			}

		case message := <-publisher.MessageChannel:
			publisher.messageBatch = append(publisher.messageBatch, message)
			if len(publisher.messageBatch) >= publisher.config.BatchSize {
				publisher.sendBatch(ctx)
			}
		}
	}
}

func (publisher *RedisTimeSeriesPublisher) sendBatch(ctx context.Context) {

	var pipeline redis.Pipeliner
	if publisher.redisClusterClient != nil {
		pipeline = publisher.redisClusterClient.Pipeline()
	} else {
		pipeline = publisher.redisClient.Pipeline()
	}

	for i := range publisher.messageBatch {
		for j := range publisher.messageBatch[i].Keys {
			pipeline.TSAdd(ctx, publisher.messageBatch[i].Keys[j], publisher.messageBatch[i].Timestamp, publisher.messageBatch[i].Values[j])
		}
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to add time series batch: %v", err)
	}

	batchNumMessages := len(publisher.messageBatch)

	publisher.mutex.Lock()
	publisher.numBatchesSent++
	publisher.numMessagesSent += batchNumMessages
	publisher.mutex.Unlock()

	publisher.messageBatch = publisher.messageBatch[:0]

	core.Log("batch")
}

func (publisher *RedisTimeSeriesPublisher) NumMessagesSent() int {
	publisher.mutex.Lock()
	numMessagesSent := publisher.numMessagesSent
	publisher.mutex.Unlock()
	return numMessagesSent
}

func (publisher *RedisTimeSeriesPublisher) NumBatchesSent() int {
	publisher.mutex.Lock()
	numBatchesSent := publisher.numBatchesSent
	publisher.mutex.Unlock()
	return numBatchesSent
}

// -------------------------------------------------------------------------------

type RedisTimeSeriesWatcher struct {
	redisClient redis.TimeseriesCmdable
	config      RedisTimeSeriesConfig
	mutex       sync.Mutex
	keys        []string
	keyToIndex  map[string]int
	timestamps  []uint64
	values      [][]float64
}

func CreateRedisTimeSeriesWatcher(ctx context.Context, config RedisTimeSeriesConfig) (*RedisTimeSeriesWatcher, error) {

	var redisClient redis.TimeseriesCmdable
	if len(config.RedisCluster) > 0 {
		client := CreateRedisClusterClient(config.RedisCluster)
		_, err := client.Ping(ctx).Result()
		if err != nil {
			return nil, err
		}
		redisClient = client
	} else {
		client := CreateRedisClient(config.RedisHostname)
		_, err := client.Ping(ctx).Result()
		if err != nil {
			return nil, err
		}
		redisClient = client
	}

	watcher := &RedisTimeSeriesWatcher{}

	watcher.config = config
	watcher.redisClient = redisClient
	watcher.keys = []string{}
	watcher.keyToIndex = make(map[string]int)

	go watcher.watcherThread(ctx)

	return watcher, nil
}

func (watcher *RedisTimeSeriesWatcher) watcherThread(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:
			// todo: pump time series from redis
			watcher.mutex.Lock()
			// todo: stash data from redis
			watcher.mutex.Unlock()
		}
	}
}

func (watcher *RedisTimeSeriesWatcher) SetKeys(keys []string) {
	watcher.mutex.Lock()
	watcher.keys = keys
	watcher.mutex.Unlock()
}

func (watcher *RedisTimeSeriesWatcher) Lock() {
	watcher.mutex.Lock()
}

func (watcher *RedisTimeSeriesWatcher) GetTimestamps(timestamps *[]uint64) {
	*timestamps = make([]uint64, len(watcher.timestamps))
	copy(*timestamps, watcher.timestamps)
}

func (watcher *RedisTimeSeriesWatcher) GetIntValues(values *[]int, key string) {
	index, exists := watcher.keyToIndex[key]
	if exists {
		*values = make([]int, len(watcher.values[index]))
		for i := range watcher.values[index] {
			(*values)[i] = int(watcher.values[index][i])
		}
	} else {
		*values = nil
	}
}

func (watcher *RedisTimeSeriesWatcher) GetFloat32Values(values *[]float32, key string) {
	index, exists := watcher.keyToIndex[key]
	if exists {
		*values = make([]float32, len(watcher.values[index]))
		for i := range watcher.values[index] {
			(*values)[i] = float32(watcher.values[index][i])
		}
	} else {
		*values = nil
	}
}

func (watcher *RedisTimeSeriesWatcher) Unlock() {
	watcher.mutex.Unlock()
}

// -------------------------------------------------------------------------------
