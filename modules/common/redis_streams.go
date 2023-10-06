package common

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/networknext/next/modules/core"
	"github.com/redis/go-redis/v9"
)

type RedisStreamsConfig struct {
	RedisHostname      string
	RedisCluster       []string
	StreamName         string
	ConsumerGroup      string
	BatchSize          int
	BatchDuration      time.Duration
	MessageChannelSize int
}

type RedisStreamsProducer struct {
	MessageChannel  chan []byte
	config          RedisStreamsConfig
	redisClient     redis.StreamCmdable
	messageBatch    [][]byte
	batchStartTime  time.Time
	mutex           sync.RWMutex
	numMessagesSent int
	numBatchesSent  int
}

func CreateRedisStreamsProducer(ctx context.Context, config RedisStreamsConfig) (*RedisStreamsProducer, error) {

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

	producer := &RedisStreamsProducer{}

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 1024 * 1024
	}

	if config.BatchDuration == 0 {
		config.BatchDuration = time.Second
	}

	if config.BatchSize == 0 {
		config.BatchSize = 10000
	}

	producer.config = config
	producer.redisClient = redisClient
	producer.MessageChannel = make(chan []byte, config.MessageChannelSize)

	go producer.updateMessageChannel(ctx)

	return producer, nil
}

func (producer *RedisStreamsProducer) updateMessageChannel(ctx context.Context) {

	ticker := time.NewTicker(producer.config.BatchDuration)

	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:
			if len(producer.messageBatch) > 0 {
				// core.Debug("send %s batch (time)", producer.config.StreamName)
				producer.sendBatch(ctx)
			}
			break

		case message := <-producer.MessageChannel:
			producer.messageBatch = append(producer.messageBatch, message)
			if len(producer.messageBatch) >= producer.config.BatchSize {
				// core.Debug("send %s batch (count)", producer.config.StreamName)
				producer.sendBatch(ctx)
			}
			break
		}
	}
}

func (producer *RedisStreamsProducer) sendBatch(ctx context.Context) {

	messageToSend := batchMessages(producer.numBatchesSent, producer.messageBatch)

	args := &redis.XAddArgs{
		Stream:     producer.config.StreamName,
		NoMkStream: false,
		Approx:     true,
		MaxLen:     1000,
		Limit:      0,
		ID:         "",
		Values:     map[string]interface{}{"type": "message", "data": messageToSend},
	}

	if _, err := producer.redisClient.XAdd(ctx, args).Result(); err != nil {
		core.Error("failed to send batch to redis: %v", err)
		return
	}

	batchNumMessages := len(producer.messageBatch)

	producer.mutex.Lock()
	producer.numBatchesSent++
	producer.numMessagesSent += batchNumMessages
	producer.mutex.Unlock()

	producer.messageBatch = producer.messageBatch[:0]

	// core.Debug("sent batch %d containing %d messages (%d bytes)", producer.numBatchesSent - 1, batchNumMessages, len(messageToSend))
}

func (producer *RedisStreamsProducer) NumMessagesSent() int {
	producer.mutex.Lock()
	numMessagesSent := producer.numMessagesSent
	producer.mutex.Unlock()
	return numMessagesSent
}

func (producer *RedisStreamsProducer) NumBatchesSent() int {
	producer.mutex.Lock()
	numBatchesSent := producer.numBatchesSent
	producer.mutex.Unlock()
	return numBatchesSent
}

// ----------------------------

type RedisStreamsConsumer struct {
	MessageChannel      chan []byte
	config              RedisStreamsConfig
	consumerId          string
	redisClient         redis.StreamCmdable
	mutex               sync.RWMutex
	numMessagesReceived int
	numBatchesReceived  int
}

func CreateRedisStreamsConsumer(ctx context.Context, config RedisStreamsConfig) (*RedisStreamsConsumer, error) {

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

	consumerId := uuid.New().String()

	// core.Debug("redis streams: stream name = %s, consumer id = %s, consumer group = %s", config.StreamName, consumerId, config.ConsumerGroup)

	_, err := redisClient.XGroupCreateMkStream(ctx, config.StreamName, config.ConsumerGroup, "$").Result()
	if err != nil {
		if strings.Contains(err.Error(), "BUSYGROUP") {
			// core.Debug("consumer group %v already exists", config.ConsumerGroup)
		} else {
			core.Error("error creating redis streams group: %v", err)
		}
	}

	consumer := &RedisStreamsConsumer{}

	consumer.config = config
	consumer.consumerId = consumerId
	consumer.redisClient = redisClient
	consumer.MessageChannel = make(chan []byte, config.MessageChannelSize)

	go consumer.receiveMessages(ctx)

	return consumer, nil
}

func (consumer *RedisStreamsConsumer) receiveMessages(ctx context.Context) {

	for {

		start := ">"

		args := &redis.XReadGroupArgs{
			Streams:  []string{consumer.config.StreamName, start},
			Group:    consumer.config.ConsumerGroup,
			Consumer: consumer.consumerId,
			Count:    10,
			Block:    time.Second * 10,
			NoAck:    false,
		}

		streamMessages, err := consumer.redisClient.XReadGroup(ctx, args).Result()

		if err == context.Canceled {
			break
		}

		if err != nil {
			continue
		}

		for _, stream := range streamMessages[0].Messages {

			batchData := []byte(stream.Values["data"].(string))

			batchMessages := parseMessages(batchData)

			// core.Debug("received %d messages (%d bytes) from redis streams", len(batchMessages), len(batchData))

			for _, message := range batchMessages {
				consumer.MessageChannel <- message
			}

			// core.Debug("batch sent to channel")

			ackResponse := consumer.redisClient.XAck(ctx, consumer.config.StreamName, consumer.config.ConsumerGroup, stream.ID)
			if ackResponse.Err() != nil {
				core.Error("failed to ack messagee: %v", err)
				continue
			}

			// core.Debug("acked message batch")

			consumer.mutex.Lock()
			consumer.numBatchesReceived += 1
			consumer.numMessagesReceived += len(batchMessages)
			consumer.mutex.Unlock()
		}
	}
}

func (consumer *RedisStreamsConsumer) NumMessageReceived() int {
	consumer.mutex.RLock()
	value := consumer.numMessagesReceived
	consumer.mutex.RUnlock()
	return value
}

func (consumer *RedisStreamsConsumer) NumBatchesReceived() int {
	consumer.mutex.RLock()
	value := consumer.numBatchesReceived
	consumer.mutex.RUnlock()
	return value
}
