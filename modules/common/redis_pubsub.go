package common

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
)

type RedisPubsubConfig struct {
	RedisHostname     string
	RedisPassword     string
	PubsubChannelName string
	BatchSize         int
	BatchDuration     time.Duration
}

type RedisPubsubProducer struct {
	MessageChannel  chan []byte
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
	
	producer := &RedisPubsubProducer{}

	producer.config = config
	producer.redisDB = redisDB
	producer.MessageChannel = make(chan []byte)

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
				producer.sendBatchToRedis()
			}
			break

		case message := <-producer.MessageChannel:
			producer.messageBatch = append(producer.messageBatch, message)
			if len(producer.messageBatch) >= producer.config.BatchSize {
				producer.sendBatchToRedis()
			}
			break
		}
	}
}

func (producer *RedisPubsubProducer) sendBatchToRedis() {

	messageToSend := batchMessages(producer.numBatchesSent, producer.messageBatch)

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(time.Second))

	_, err := producer.redisDB.Publish(ctx, producer.config.PubsubChannelName, messageToSend).Result()
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
	config              RedisPubsubConfig
	redisDB             *redis.Client
	subscriber          *redis.PubSub
	MessageChannel      chan []byte
	numMessagesReceived int
	numBatchesReceived  int
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

	consumer := &RedisPubsubConsumer{}

	consumer.config = config
	consumer.redisDB = redisDB
	consumer.subscriber = redisDB.Subscribe(ctx, config.PubsubChannelName)
	consumer.MessageChannel = make(chan []byte, 10*1024) // todo: make the chan length configurable via config

	return consumer, nil
}

func (consumer *RedisPubsubConsumer) ProcessRedisMessages(ctx context.Context) {

	for {

		messageBatch, err := consumer.subscriber.ReceiveMessage(ctx)
		if err != nil {
			core.Error("failed to receive message batch from redis pubsub: %v", err)
			continue
		}

		fmt.Printf("received message batch from redis")

		_ = messageBatch

		// todo: unbatch message and stick the messages in the chan []byte MessageChannel for the caller to receive

		// todo: update new messages received

		// todo: update num batches received
	}
}

func (consumer *RedisPubsubConsumer) NumMessageReceived() int {
	return consumer.numMessagesReceived
}

func (consumer *RedisPubsubConsumer) NumBatchesReceived() int {
	return consumer.numBatchesReceived
}
