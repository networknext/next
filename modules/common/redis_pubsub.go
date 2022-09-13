package common

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
)

type RedisPubsubConfig struct {
	RedisHostname      string
	RedisPassword      string
	PubsubChannelName  string
	BatchSize          int
	BatchDuration      time.Duration
	MessageChannelSize int
}

type RedisPubsubProducer struct {
	MessageChannel  chan []byte
	config          RedisPubsubConfig
	mutex           sync.RWMutex
	redisClient     *redis.Client
	messageBatch    [][]byte
	batchStartTime  time.Time
	numMessagesSent int
	numBatchesSent  int
}

func CreateRedisPubsubProducer(ctx context.Context, config RedisPubsubConfig) (*RedisPubsubProducer, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisHostname,
		Password: config.RedisPassword,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		core.Error("failed to create pubsub client: %v", err)
		return nil, err
	}

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 10 * 1024
	}

	producer := &RedisPubsubProducer{}
	producer.config = config
	producer.redisClient = redisClient
	producer.MessageChannel = make(chan []byte, config.MessageChannelSize)

	go producer.updateMessageChannel(ctx)

	return producer, nil
}

func (producer *RedisPubsubProducer) updateMessageChannel(ctx context.Context) {
	ticker := time.NewTicker(producer.config.BatchDuration)

	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:
			if len(producer.messageBatch) > 0 {
				producer.sendBatch(ctx)
			}

		case message := <-producer.MessageChannel:
			producer.messageBatch = append(producer.messageBatch, message)
			if len(producer.messageBatch) >= producer.config.BatchSize {
				producer.sendBatch(ctx)
			}
		}
	}
}

func (producer *RedisPubsubProducer) sendBatch(ctx context.Context) {
	messageToSend := batchMessages(producer.numBatchesSent, producer.messageBatch)

	timeoutContext, _ := context.WithTimeout(ctx, time.Duration(time.Second))

	_, err := producer.redisClient.Publish(timeoutContext, producer.config.PubsubChannelName, messageToSend).Result()
	if err != nil {
		core.Error("failed to send batched pubsub messages to redis: %v", err)
		return
	}

	batchId := producer.numBatchesSent
	batchNumMessages := len(producer.messageBatch)

	producer.mutex.Lock()
	producer.numBatchesSent++
	producer.numMessagesSent += batchNumMessages
	producer.mutex.Unlock()

	producer.messageBatch = [][]byte{}

	core.Debug("sent batch %d containing %d messages (%d bytes)", batchId, batchNumMessages, len(messageToSend))
}

func batchMessages(batchId int, messages [][]byte) []byte {
	numMessages := len(messages)
	messageBytes := 4 + 4 + numMessages*4
	for i := range messages {
		messageBytes += len(messages[i])
	}
	messageData := make([]byte, messageBytes)
	index := 0
	encoding.WriteUint32(messageData, &index, uint32(batchId))
	encoding.WriteUint32(messageData, &index, uint32(len(messages)))
	for i := range messages {
		encoding.WriteUint32(messageData, &index, uint32(len(messages[i])))
		encoding.WriteBytes(messageData, &index, messages[i], len(messages[i]))
	}
	return messageData
}

func (producer *RedisPubsubProducer) NumMessagesSent() int {
	producer.mutex.Lock()
	numMessagesSent := producer.numMessagesSent
	producer.mutex.Unlock()
	return numMessagesSent
}

func (producer *RedisPubsubProducer) NumBatchesSent() int {
	producer.mutex.Lock()
	numBatchesSent := producer.numBatchesSent
	producer.mutex.Unlock()
	return numBatchesSent
}

// -----------------------------------------------

type RedisPubsubConsumer struct {
	MessageChannel      chan []byte
	config              RedisPubsubConfig
	redisClient         *redis.Client
	pubsubSubscription  *redis.PubSub
	pubsubChannel       <-chan *redis.Message
	mutex               sync.RWMutex
	numMessagesReceived int
	numBatchesReceived  int
}

func CreateRedisPubsubConsumer(ctx context.Context, config RedisPubsubConfig) (*RedisPubsubConsumer, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisHostname,
		Password: config.RedisPassword,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 10 * 1024
	}

	consumer := &RedisPubsubConsumer{}

	consumer.config = config
	consumer.redisClient = redisClient
	consumer.pubsubSubscription = consumer.redisClient.Subscribe(ctx, config.PubsubChannelName)
	consumer.pubsubChannel = consumer.pubsubSubscription.Channel()
	consumer.MessageChannel = make(chan []byte, config.MessageChannelSize)

	go consumer.processRedisMessages(ctx)

	return consumer, nil
}

func (consumer *RedisPubsubConsumer) processRedisMessages(ctx context.Context) {
	for {
		select {

		case <-ctx.Done():
			return

		case messageBatch := <-consumer.pubsubChannel:

			batchMessages := parseMessages([]byte(messageBatch.Payload))

			core.Debug("received %d messages (%v bytes) from redis pubsub", len(batchMessages), len([]byte(messageBatch.Payload)))

			for _, message := range batchMessages {
				consumer.MessageChannel <- message
			}

			consumer.mutex.Lock()
			consumer.numBatchesReceived += 1
			consumer.numMessagesReceived += len(batchMessages)
			consumer.mutex.Unlock()
		}
	}
}

func parseMessages(messages []byte) [][]byte {
	index := 0
	var batchNum uint32
	var numMessages uint32
	if !encoding.ReadUint32(messages, &index, &batchNum) {
		core.Error("redis pubsub consumer: could not read batch number")
		return [][]byte{}
	}
	if !encoding.ReadUint32(messages, &index, &numMessages) {
		core.Error("redis pubsub consumer: could not read number of messages")
		return [][]byte{}
	}
	messagesData := make([][]byte, numMessages)
	for i := 0; i < int(numMessages); i++ {
		var messageLength uint32
		if !encoding.ReadUint32(messages, &index, &messageLength) {
			core.Error("redis pubsub consumer: could not read length of the message")
			return [][]byte{}
		}
		if !encoding.ReadBytes(messages, &index, &messagesData[i], uint32(messageLength)) {
			core.Error("redis pubsub consumer: could not read message data")
			return [][]byte{}
		}
	}
	return messagesData
}

func (consumer *RedisPubsubConsumer) NumMessageReceived() int {
	consumer.mutex.RLock()
	value := consumer.numMessagesReceived
	consumer.mutex.RUnlock()
	return value
}

func (consumer *RedisPubsubConsumer) NumBatchesReceived() int {
	consumer.mutex.RLock()
	value := consumer.numBatchesReceived
	consumer.mutex.RUnlock()
	return value
}
