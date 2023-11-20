package common

import (
	"context"
	"strings"
	"sync"
	"time"
	"fmt"

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
	AverageWindow      int // the time period to average samples into a single sample in milliseconds. default is 1 minute.
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

	if config.AverageWindow == 0 {
		config.AverageWindow = 60 * 1000 // 60 seconds in milliseconds
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
		pipeline.TSCreateWithArgs(ctx, fmt.Sprintf("%s-internal", newKeys[i]), &options)
		pipeline.TSCreateWithArgs(ctx, newKeys[i], &options)
		pipeline.TSCreateRule(ctx, fmt.Sprintf("%s-internal", newKeys[i]), newKeys[i], redis.Avg, publisher.config.AverageWindow)
	}

	for i := range publisher.messageBatch {
		for j := range publisher.messageBatch[i].Keys {
			pipeline.TSAdd(ctx, publisher.messageBatch[i].Keys[j], publisher.messageBatch[i].Timestamp, publisher.messageBatch[i].Values[j])
		}
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		if !strings.Contains(err.Error(), "key already exists") {
			core.Warn("failed to add time series: %v", err)
		}
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

	if config.AverageWindow == 0 {
		config.AverageWindow = 60 * 1000 // 60 seconds in milliseconds
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

			// first, work out which time series keys exist

			var pipeline redis.Pipeliner
			if watcher.redisClusterClient != nil {
				pipeline = watcher.redisClusterClient.Pipeline()
			} else {
				pipeline = watcher.redisClient.Pipeline()
			}

			for i := range keys {
				pipeline.Exists(ctx, keys[i])
			}

			cmds, err := pipeline.Exec(ctx)
			if err != nil {
				core.Error("failed to check existing time series keys: %v", err)
				break
			}

			exists := make([]bool, len(keys))

			for i := range exists {
				if cmds[i].(*redis.IntCmd).Val() > 0 {
					exists[i] = true
				}
			}

			// get time series for existing keys only

			if watcher.redisClusterClient != nil {
				pipeline = watcher.redisClusterClient.Pipeline()
			} else {
				pipeline = watcher.redisClient.Pipeline()
			}

			currentTime := int(time.Now().UnixNano() / 1000000)

			for i := range keys {
				if exists[i] {
					pipeline.TSRange(ctx, keys[i], currentTime-watcher.config.DisplayWindow, currentTime)
				}
			}

			cmds, err = pipeline.Exec(ctx)
			if err != nil {
				core.Error("failed to get time series data: %v", err)
				break
			}

			keyToIndex := make(map[string]int, len(keys))
			timestamps := make([][]uint64, len(keys))
			values := make([][]float64, len(keys))

			for i := range keys {
				keyToIndex[keys[i]] = i
			}

			index := 0
			for i := range keys {

				sampleRate := uint64(watcher.config.AverageWindow)

				startTimestamp := uint64(currentTime) - uint64(watcher.config.DisplayWindow)
				startTimestamp -= startTimestamp % uint64(sampleRate)
				startTimestamp += uint64(sampleRate)

				endTimestamp := startTimestamp + uint64(watcher.config.DisplayWindow)
				endTimestamp -= endTimestamp % uint64(sampleRate)
				endTimestamp -= uint64(sampleRate) * 2

				if exists[i] {

					// time series exists

					data := cmds[index].(*redis.TSTimestampValueSliceCmd).Val()

					dataLength := len(data)
					firstTimestamp := endTimestamp
					lastTimestamp := endTimestamp
					if dataLength > 0 {
						firstTimestamp = uint64(data[0].Timestamp)
						lastTimestamp = uint64(data[dataLength-1].Timestamp)
					}

					timestamps[i] = make([]uint64, 0)
					values[i] = make([]float64, 0)

					// pad in front with zero samples

					for timestamp := startTimestamp; timestamp < firstTimestamp; timestamp += sampleRate {
						timestamps[i] = append(timestamps[i], timestamp)
						values[i] = append(values[i], 0.0)
					}

					// insert real samples in middle

					for j := range data {
						timestamps[i] = append(timestamps[i], uint64(data[j].Timestamp))
						values[i] = append(values[i], data[j].Value)
					}

					// pad after with zero samples

					for timestamp := lastTimestamp + sampleRate; timestamp <= endTimestamp; timestamp += sampleRate {
						timestamps[i] = append(timestamps[i], timestamp)
						values[i] = append(values[i], 0.0)
					}

					index++

				} else {

					// does not exist
					numSamples := 2
					timestamps[i] = make([]uint64, numSamples)
					values[i] = make([]float64, numSamples)
					timestamps[i][0] = uint64(startTimestamp)
					timestamps[i][1] = uint64(endTimestamp)

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
