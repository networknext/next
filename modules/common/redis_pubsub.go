package common

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
)

type RedisPubsubConfig struct {
	RedisHostname string
	RedisPassword string
	ChannelName   string
	BatchSize     int
	BatchDuration time.Duration
}

type RedisPubsubProducer struct {
	config          RedisPubsubConfig
	mutex           sync.Mutex
	redisDB         *redis.Client
	messageBatch    [][]byte
	batchStartTime  time.Time
	numMessagesSent int
	numBatchesSent  int
}

func CreateRedisPubsubProducer(ctx context.Context, config RedisPubsubConfig) (*RedisPubsubProducer, error) {
	redisDB := redis.NewClient(&redis.Options{
		Addr:     config.RedisHostname,
		Password: config.RedisPassword,
	})
	_, err := redisDB.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return &RedisPubsubProducer{config: config, redisDB: redisDB}, nil
}

func (producer *RedisPubsubProducer) ShutdownRedisPubsubProducer(ctx context.Context) error {
	if err := producer.redisDB.Close(); err != nil {
		return err
	}

	return nil
}

func (producer *RedisPubsubProducer) SendMessage(ctx context.Context, message []byte) {

	producer.mutex.Lock()

	defer producer.mutex.Unlock()

	// add the message to the batch

	producer.messageBatch = append(producer.messageBatch, message)

	if len(producer.messageBatch) == 1 {
		producer.batchStartTime = time.Now()
	}

	// should we send the batch now?

	if len(producer.messageBatch) >= producer.config.BatchSize || time.Since(producer.batchStartTime) >= producer.config.BatchDuration {

		// yes. send the current batch of messages to redis as a single coalesced message

		messageToSend := batchMessages(producer.numBatchesSent, producer.messageBatch)
		_, err := producer.redisDB.Publish(ctx, producer.config.ChannelName, messageToSend).Result()
		if err != nil {
			core.Error("failed to send batched pubsub messages to redis: %v", err)
			return
		}

		// success!

		batchId := producer.numBatchesSent
		batchNumMessages := len(producer.messageBatch)

		producer.numBatchesSent++
		producer.numMessagesSent += batchNumMessages

		producer.messageBatch = [][]byte{}

		core.Debug("sent batch %d containing %d messages (%d bytes)\n", batchId, batchNumMessages, len(messageToSend))
	}
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

func batchMessages(batchId int, messages [][]byte) []byte {
	numMessages := len(messages)
	messageBytes := 4 + numMessages*4
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

// -----------------------------------------------

type RedisPubsubConsumer struct {
	config                   RedisPubsubConfig
	redisDB                  *redis.Client
	MessageChannel           <-chan *redis.Message
	pubsubHandler            *redis.PubSub
	NumberOfMessagesReceived int64
	NumberOfBatchesReceived  int64
}

func CreateRedisPubsubConsumer(ctx context.Context, config RedisPubsubConfig) (*RedisPubsubConsumer, error) {
	redisDB := redis.NewClient(&redis.Options{
		Addr:     config.RedisHostname,
		Password: config.RedisPassword,
	})
	_, err := redisDB.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	pubsubHandler := redisDB.Subscribe(ctx, config.ChannelName)

	messageChannel := pubsubHandler.Channel()

	return &RedisPubsubConsumer{config: config, redisDB: redisDB, MessageChannel: messageChannel, pubsubHandler: pubsubHandler}, nil
}

func (consumer *RedisPubsubConsumer) ShutdownRedisPubsubConsumer(ctx context.Context) error {
	if err := consumer.pubsubHandler.Unsubscribe(ctx, consumer.config.ChannelName); err != nil {
		return err
	}

	if err := consumer.redisDB.Close(); err != nil {
		return err
	}

	return nil
}
