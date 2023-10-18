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
	Retention          int
	DisplayWindow      int
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
	keys               map[string]bool
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

	if config.Retention == 0 {
		config.Retention = 3600 * 1000 // 1 hour in milliseconds
	}

	publisher.config = config
	publisher.keys = make(map[string]bool)
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

	publisher.mutex.Lock()
	keys := publisher.keys
	publisher.mutex.Unlock()

	newKeys := make([]string, 0)

	var pipeline redis.Pipeliner
	if publisher.redisClusterClient != nil {
		pipeline = publisher.redisClusterClient.Pipeline()
	} else {
		pipeline = publisher.redisClient.Pipeline()
	}

	for i := range publisher.messageBatch {
		for j := range publisher.messageBatch[i].Keys {
			_, exists := keys[publisher.messageBatch[i].Keys[j]]
			if !exists {
				newKeys = append(newKeys, publisher.messageBatch[i].Keys[j])
			}
		}
	}

	for i := range newKeys {
		options := redis.TSOptions{}
		options.Retention = publisher.config.Retention
		options.DuplicatePolicy = "MAX"
		pipeline.TSCreateWithArgs(ctx, newKeys[i], &options)
	}

	for i := range publisher.messageBatch {
		for j := range publisher.messageBatch[i].Keys {
			pipeline.TSAdd(ctx, publisher.messageBatch[i].Keys[j], publisher.messageBatch[i].Timestamp, publisher.messageBatch[i].Values[j])
		}
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to add time series: %v", err)
	}

	batchNumMessages := len(publisher.messageBatch)

	publisher.mutex.Lock()
	publisher.numBatchesSent++
	publisher.numMessagesSent += batchNumMessages
	for i := range newKeys {
		publisher.keys[newKeys[i]] = true
	}
	publisher.mutex.Unlock()

	publisher.messageBatch = publisher.messageBatch[:0]
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
	config             RedisTimeSeriesConfig
	redisClient        *redis.Client
	redisClusterClient *redis.ClusterClient
	mutex              sync.Mutex
	keys               []string
	keyToIndex         map[string]int
	timestamps         [][]uint64
	values             [][]float64
}

func CreateRedisTimeSeriesWatcher(ctx context.Context, config RedisTimeSeriesConfig) (*RedisTimeSeriesWatcher, error) {

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

	if config.DisplayWindow == 0 {
		config.DisplayWindow = 3600 * 1000 // 1 hour in milliseconds
	}

	watcher := &RedisTimeSeriesWatcher{}

	watcher.config = config
	watcher.redisClient = client
	watcher.redisClusterClient = clusterClient
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

			watcher.mutex.Lock()
			keys := make([]string, len(watcher.keys))
			copy(keys, watcher.keys)
			watcher.mutex.Unlock()

			if len(keys) == 0 {
				break
			}

			var pipeline redis.Pipeliner
			if watcher.redisClusterClient != nil {
				pipeline = watcher.redisClusterClient.Pipeline()
			} else {
				pipeline = watcher.redisClient.Pipeline()
			}

			currentTime := int(time.Now().UnixNano() / 1000000)

			for i := range keys {
				pipeline.TSRange(ctx, keys[i], currentTime-watcher.config.DisplayWindow, currentTime)
			}

			cmds, err := pipeline.Exec(ctx)
			if err != nil {
				break
			}

			keyToIndex := make(map[string]int, len(keys))
			timestamps := make([][]uint64, len(keys))
			values := make([][]float64, len(keys))

			for i := range keys {
				keyToIndex[keys[i]] = i
			}

			for i := range keys {
				data := cmds[i].(*redis.TSTimestampValueSliceCmd).Val()
				timestamps[i] = make([]uint64, len(data))
				values[i] = make([]float64, len(data))
				for j := range data {
					timestamps[i][j] = uint64(data[j].Timestamp)
					values[i][j] = data[j].Value
				}
			}

			watcher.mutex.Lock()
			watcher.keyToIndex = keyToIndex
			watcher.timestamps = timestamps
			watcher.values = values
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

func (watcher *RedisTimeSeriesWatcher) GetIntValues(timestamps *[]uint64, values *[]int, key string) {
	index, exists := watcher.keyToIndex[key]
	if exists {
		*timestamps = make([]uint64, len(watcher.timestamps[index]))
		*values = make([]int, len(watcher.values[index]))
		copy(*timestamps, watcher.timestamps[index])
		for i := range watcher.values[index] {
			(*values)[i] = int(watcher.values[index][i])
		}
	} else {
		*timestamps = nil
		*values = nil
	}
}

func (watcher *RedisTimeSeriesWatcher) GetFloat32Values(timestamps *[]uint64, values *[]float32, key string) {
	index, exists := watcher.keyToIndex[key]
	if exists {
		*timestamps = make([]uint64, len(watcher.timestamps[index]))
		*values = make([]float32, len(watcher.values[index]))
		copy(*timestamps, watcher.timestamps[index])
		for i := range watcher.values[index] {
			(*values)[i] = float32(watcher.values[index][i])
		}
	} else {
		*timestamps = nil
		*values = nil
	}
}

func (watcher *RedisTimeSeriesWatcher) Unlock() {
	watcher.mutex.Unlock()
}

// -------------------------------------------------------------------------------
