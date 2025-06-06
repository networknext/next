package common

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/networknext/next/modules/core"

	"github.com/redis/go-redis/v9"
)

type RedisCountersConfig struct {
	RedisHostname      string
	RedisCluster       []string
	BatchDuration      time.Duration
	MessageChannelSize int
	Retention          int
	SumWindow          int // the time period to sum into a single sample in milliseconds. default is 1 minute.
	DisplayWindow      int // how far back from current time to query summed samples in milliseconds. default is 1 hour.
}

// -------------------------------------------------------------------------------

type RedisCountersPublisher struct {
	config             RedisCountersConfig
	redisClient        *redis.Client
	redisClusterClient *redis.ClusterClient
	mutex              sync.Mutex
	keys               map[string]bool
	MessageChannel     chan string
}

func CreateRedisCountersPublisher(ctx context.Context, config RedisCountersConfig) (*RedisCountersPublisher, error) {

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

	publisher := &RedisCountersPublisher{}

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 1024 * 1024
	}

	if config.BatchDuration == 0 {
		config.BatchDuration = time.Second
	}

	if config.Retention == 0 {
		config.Retention = 3600 * 1000 // 1 hour in milliseconds
	}

	if config.SumWindow == 0 {
		config.SumWindow = 60 * 1000 // 60 seconds in milliseconds
	}

	publisher.config = config
	publisher.keys = make(map[string]bool)
	publisher.MessageChannel = make(chan string, config.MessageChannelSize)
	publisher.redisClient = client
	publisher.redisClusterClient = clusterClient

	go publisher.updateMessageChannel(ctx)

	return publisher, nil
}

func (publisher *RedisCountersPublisher) updateMessageChannel(ctx context.Context) {

	ticker := time.NewTicker(publisher.config.BatchDuration)

	newKeys := make(map[string]bool)
	counters := make(map[string]uint64, 64)

	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:
			publisher.sendBatch(ctx, counters, newKeys)
			for k := range counters {
				counters[k] = 0
			}
			if len(newKeys) > 0 {
				newKeys = make(map[string]bool)
			}

		case key := <-publisher.MessageChannel:
			counter, exists := counters[key]
			if !exists {
				newKeys[key] = true
			}
			counters[key] = counter + 1
		}
	}
}

func (publisher *RedisCountersPublisher) sendBatch(ctx context.Context, counters map[string]uint64, newKeys map[string]bool) {

	timestamp := time.Now().UnixNano() / 1000000

	var pipeline redis.Pipeliner
	if publisher.redisClusterClient != nil {
		pipeline = publisher.redisClusterClient.Pipeline()
	} else {
		pipeline = publisher.redisClient.Pipeline()
	}

	for k := range newKeys {
		options := redis.TSOptions{}
		options.Retention = publisher.config.Retention
		options.DuplicatePolicy = "SUM"
		pipeline.TSCreateWithArgs(ctx, fmt.Sprintf("%s-internal", k), &options)
		pipeline.TSCreateWithArgs(ctx, k, &options)
		pipeline.TSCreateRule(ctx, fmt.Sprintf("%s-internal", k), k, redis.Sum, publisher.config.SumWindow)
	}

	for k, v := range counters {
		if v > 0 {
			pipeline.TSAdd(ctx, fmt.Sprintf("%s-internal", k), timestamp, float64(v))
		}
	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		if !strings.Contains(err.Error(), "key already exists") {
			core.Warn("failed to add counters: %v", err)
		}
	}
}

// -------------------------------------------------------------------------------

type RedisCountersWatcher struct {
	config             RedisCountersConfig
	redisClient        *redis.Client
	redisClusterClient *redis.ClusterClient
	mutex              sync.Mutex
	keys               []string
	keyToIndex         map[string]int
	timestamps         [][]uint64
	values             [][]float64
}

func CreateRedisCountersWatcher(ctx context.Context, config RedisCountersConfig) (*RedisCountersWatcher, error) {

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

	if config.SumWindow == 0 {
		config.SumWindow = 60 * 1000 // 60 seconds in milliseconds
	}

	if config.DisplayWindow == 0 {
		config.DisplayWindow = 3600 * 1000 // 1 hour in milliseconds
	}

	watcher := &RedisCountersWatcher{}

	watcher.config = config
	watcher.redisClient = client
	watcher.redisClusterClient = clusterClient
	watcher.keys = []string{}
	watcher.keyToIndex = make(map[string]int)

	go watcher.watcherThread(ctx)

	return watcher, nil
}

func (watcher *RedisCountersWatcher) watcherThread(ctx context.Context) {
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
				core.Error("failed to check existing counters keys: %v", err)
				break
			}

			exists := make([]bool, len(keys))

			for i := range exists {
				if cmds[i].(*redis.IntCmd).Val() > 0 {
					exists[i] = true
				}
			}

			// get counter data for existing keys

			if watcher.redisClusterClient != nil {
				pipeline = watcher.redisClusterClient.Pipeline()
			} else {
				pipeline = watcher.redisClient.Pipeline()
			}

			currentTime := int(time.Now().UnixNano()) / 1000000

			for i := range keys {
				if exists[i] {
					pipeline.TSRange(ctx, keys[i], currentTime-watcher.config.DisplayWindow, currentTime)
				}
			}

			cmds, err = pipeline.Exec(ctx)
			if err != nil {
				core.Error("failed to get counters data: %v", err)
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

				startTimestamp := uint64(currentTime) - uint64(watcher.config.DisplayWindow)
				startTimestamp -= startTimestamp % uint64(watcher.config.SumWindow)
				startTimestamp += uint64(watcher.config.SumWindow)

				endTimestamp := startTimestamp + uint64(watcher.config.DisplayWindow)
				endTimestamp -= endTimestamp % uint64(watcher.config.SumWindow)
				endTimestamp -= uint64(watcher.config.SumWindow) * 2

				if exists[i] {

					data := cmds[index].(*redis.TSTimestampValueSliceCmd).Val()

					valueMap := make(map[uint64]float64)

					for j := range data {
						valueMap[uint64(data[j].Timestamp)] = data[j].Value
					}

					timestamps[i] = make([]uint64, 0)
					values[i] = make([]float64, 0)

					for timestamp := startTimestamp; timestamp < endTimestamp; timestamp += uint64(watcher.config.SumWindow) {
						timestamps[i] = append(timestamps[i], timestamp)
						value, exists := valueMap[timestamp]
						if exists {
							values[i] = append(values[i], value)
						} else {
							values[i] = append(values[i], 0.0)
						}
					}

					index++

				} else {

					timestamps[i] = []uint64{startTimestamp, endTimestamp}
					values[i] = []float64{0.0, 0.0}

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

func (watcher *RedisCountersWatcher) SetKeys(keys []string) {
	watcher.mutex.Lock()
	watcher.keys = keys
	watcher.mutex.Unlock()
}

func (watcher *RedisCountersWatcher) Lock() {
	watcher.mutex.Lock()
}

func (watcher *RedisCountersWatcher) GetIntValues(timestamps *[]uint64, values *[]int, key string) {
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

func (watcher *RedisCountersWatcher) GetFloat32Values(timestamps *[]uint64, values *[]float32, key string) {
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

func (watcher *RedisCountersWatcher) GetFloatValue(key string) float64 {
	index, exists := watcher.keyToIndex[key]
	if exists {
		// IMPORTANT: Sometimes redis counters hasn't processed the most recent sample yet and it is zero
		// In this wase, walk back up to 5 samples to find a non-zero sample and return that. Without this
		// the session counts in the portal (driven by counters) flicker to zero occasionally.
		lastIndex := len(watcher.values[index]) - 1
		for i := 0; i < 5; i++ {
			if lastIndex-i < 0 {
				break
			}
			if watcher.values[index][lastIndex-i] != 0 {
				return watcher.values[index][lastIndex-i]
			}
		}
		return watcher.values[index][lastIndex]
	} else {
		return 0.0
	}
}

func (watcher *RedisCountersWatcher) Unlock() {
	watcher.mutex.Unlock()
}

// -------------------------------------------------------------------------------
