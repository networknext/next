package common

import (
	"context"
	"sync"
	"time"

	"github.com/networknext/backend/modules/core"
)

type GooglePubsubConfig struct {
	Topic              string
	// ...
	BatchSize          int
	BatchDuration      time.Duration
	MessageChannelSize int
}

type GooglePubsubProducer struct {
	MessageChannel chan []byte
	config         GooglePubsubConfig
	// ...
	messageBatch    [][]byte
	batchStartTime  time.Time
	mutex           sync.RWMutex
	numMessagesSent int
	numBatchesSent  int
}

func CreateGooglePubsubProducer(ctx context.Context, config GooglePubsubConfig) (*GooglePubsubProducer, error) {

	producer := &GooglePubsubProducer{}

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 10 * 1024
	}

	if config.BatchDuration == 0 {
		config.BatchDuration = time.Second
	}

	producer.config = config
	// ...
	producer.MessageChannel = make(chan []byte, config.MessageChannelSize)

	go producer.updateMessageChannel(ctx)

	return producer, nil
}

func (producer *GooglePubsubProducer) updateMessageChannel(ctx context.Context) {

	ticker := time.NewTicker(producer.config.BatchDuration)

	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:
			if len(producer.messageBatch) > 0 {
				producer.sendBatch(ctx)
			}
			break

		case message := <-producer.MessageChannel:
			producer.messageBatch = append(producer.messageBatch, message)
			if len(producer.messageBatch) >= producer.config.BatchSize {
				producer.sendBatch(ctx)
			}
			break
		}
	}
}

func (producer *GooglePubsubProducer) sendBatch(ctx context.Context) {

	messageToSend := batchMessages(producer.numBatchesSent, producer.messageBatch)

	// todo: send batched messages

	batchId := producer.numBatchesSent
	batchNumMessages := len(producer.messageBatch)

	producer.mutex.Lock()
	producer.numBatchesSent++
	producer.numMessagesSent += batchNumMessages
	producer.mutex.Unlock()

	producer.messageBatch = [][]byte{}

	core.Debug("sent batch %d containing %d messages (%d bytes)", batchId, batchNumMessages, len(messageToSend))
}

// ----------------------------

type GooglePubsubConsumer struct {
	MessageChannel chan []byte
	config         GooglePubsubConfig
	// ...
	mutex               sync.RWMutex
	numMessagesReceived int
	numBatchesReceived  int
}

func CreateGooglePubsubConsumer(ctx context.Context, config GooglePubsubConfig) (*GooglePubsubConsumer, error) {

	consumer := &GooglePubsubConsumer{}

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 10 * 1024
	}

	consumer.config = config
	consumer.MessageChannel = make(chan []byte, config.MessageChannelSize)

	// ...

	go consumer.receiveMessages(ctx)

	return consumer, nil
}

func (consumer *GooglePubsubConsumer) receiveMessages(ctx context.Context) {

	/*
	for {

		// todo: quit if context is done

		// todo: read message

		if err != nil {
			core.Error("error reading from google pubsub: %s", err)
			continue
		}

		for _, stream := range streamMessages[0].Messages {

			batchData := []byte(stream.Values["data"].(string))

			batchMessages := parseMessages(batchData)

			core.Debug("received %d messages (%d bytes) from redis streams", len(batchMessages), len(batchData))

			for _, message := range batchMessages {
				consumer.MessageChannel <- message
			}

			consumer.redisClient.XAck(ctx, consumer.config.StreamName, consumer.config.StreamName, stream.ID)

			consumer.mutex.Lock()
			consumer.numBatchesReceived += 1
			consumer.numMessagesReceived += len(batchMessages)
			consumer.mutex.Unlock()
		}
	}
	*/
}

func (consumer *GooglePubsubConsumer) NumMessageReceived() int {
	consumer.mutex.RLock()
	value := consumer.numMessagesReceived
	consumer.mutex.RUnlock()
	return value
}

func (consumer *GooglePubsubConsumer) NumBatchesReceived() int {
	consumer.mutex.RLock()
	value := consumer.numBatchesReceived
	consumer.mutex.RUnlock()
	return value
}
