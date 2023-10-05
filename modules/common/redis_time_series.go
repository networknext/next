package common

import (
	"context"
	"sync"
	"time"

	"github.com/networknext/next/modules/core"
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
	config RedisTimeSeriesConfig
	// redisClient     redis.StreamCmdable
	mutex           sync.Mutex
	messageBatch    []*RedisTimeSeriesMessage
	numMessagesSent int
	numBatchesSent  int
	MessageChannel  chan *RedisTimeSeriesMessage
}

func CreateRedisTimeSeriesPublisher(ctx context.Context, config RedisTimeSeriesConfig) (*RedisTimeSeriesPublisher, error) {

	/*
		var redisClient redis.StreamCmdable
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
	*/

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

	/*
		publisher.redisClient = redisClient
	*/

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

	// todo: send time series data batch to redis
	/*
		messageToSend := batchMessages(producer.numBatchesSent, producer.messageBatch)

		timeoutContext, _ := context.WithTimeout(ctx, time.Duration(time.Second))

		_, err := producer.redisClient.Publish(timeoutContext, producer.config.PubsubChannelName, messageToSend).Result()
		if err != nil {
			core.Error("failed to send batched pubsub messages to redis: %v", err)
			return
		}
	*/

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
	config     RedisTimeSeriesConfig
	mutex      sync.Mutex
	keys       []string
	timestamps []uint64
	values     [][]float64
}

func CreateRedisTimeSeriesWatcher(ctx context.Context, config RedisTimeSeriesConfig, keys []string) (*RedisTimeSeriesWatcher, error) {

	/*
		var redisClient redis.StreamCmdable
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
	*/

	watcher := &RedisTimeSeriesWatcher{}

	watcher.config = config
	watcher.keys = keys

	// todo: create watcher thread

	return watcher, nil
}

func (watcher *RedisTimeSeriesWatcher) GetTimeSeries() (keys []string, timestamps []uint64, values [][]float64) {
	watcher.mutex.Lock()
	keys = watcher.keys
	timestamps = watcher.timestamps
	values = watcher.values
	watcher.mutex.Unlock()
	return keys, timestamps, values
}

// -------------------------------------------------------------------------------
