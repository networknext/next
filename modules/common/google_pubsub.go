package common

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/networknext/backend/modules/core"
)

type GooglePubsubConfig struct {
	ProjectID    string
	Topic        string
	Subscription string
	// ...
	pubsubClient       *pubsub.Client
	pubsubTopic        *pubsub.Topic
	pubsubSubscription *pubsub.Subscription
	settings           *pubsub.PublishSettings
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

	pubsubClient, err := pubsub.NewClient(ctx, config.ProjectID)
	if err != nil {
		core.Error("failed to create pubsub client: %v", err)
		return nil, err
	}

	config.pubsubClient = pubsubClient
	config.pubsubTopic = pubsubClient.Topic(config.Topic)

	if config.pubsubTopic == nil {
		core.Error("failed to create google pubsub consumer: pubsub topic/subscription was not configured correctly")
		return nil, err
	}

	producer.config = config

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

	timeoutContext, _ := context.WithTimeout(ctx, time.Duration(time.Second))

	_, err := producer.config.pubsubTopic.Publish(timeoutContext, &pubsub.Message{Data: messageToSend}).Get(ctx)
	if err != nil {
		core.Error("failed to send batched pubsub messages to google: %v", err) // todo: Not sure if this is ok or not. Billing is important and needs to always succeed
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

func (producer *GooglePubsubProducer) NumMessagesSent() int {
	producer.mutex.RLock()
	value := producer.numMessagesSent
	producer.mutex.RUnlock()
	return value
}

func (producer *GooglePubsubProducer) NumBatchesSent() int {
	producer.mutex.RLock()
	value := producer.numBatchesSent
	producer.mutex.RUnlock()
	return value
}

func (producer *GooglePubsubProducer) Close(ctx context.Context) {
	producer.config.pubsubClient.Close()
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

	pubsubClient, err := pubsub.NewClient(ctx, config.ProjectID)
	if err != nil {
		core.Error("failed to create google pubsub consumer: %v", err)
		return nil, err
	}

	config.pubsubClient = pubsubClient
	config.pubsubTopic = pubsubClient.Topic(config.Topic)
	config.pubsubSubscription = pubsubClient.Subscription(config.Subscription)

	if config.pubsubTopic == nil || config.pubsubSubscription == nil {
		core.Error("failed to create google pubsub consumer: pubsub topic/subscription was not configured correctly")
		return nil, err
	}

	consumer.config = config
	consumer.MessageChannel = make(chan []byte, config.MessageChannelSize)

	go consumer.receiveMessages(ctx)

	return consumer, nil
}

func (consumer *GooglePubsubConsumer) receiveMessages(ctx context.Context) {
	consumer.config.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		batchMessages := parseMessages(m.Data)

		core.Debug("received %d messages (%v bytes) from redis pubsub", len(batchMessages), len(m.Data))

		for _, message := range batchMessages {
			consumer.MessageChannel <- message
		}

		consumer.mutex.Lock()
		consumer.numBatchesReceived += 1
		consumer.numMessagesReceived += len(batchMessages)
		consumer.mutex.Unlock()

		m.Ack()
	})
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

func (consumer *GooglePubsubConsumer) Close(ctx context.Context) {
	consumer.config.pubsubClient.Close()
}
