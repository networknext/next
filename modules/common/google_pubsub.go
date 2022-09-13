package common

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/networknext/backend/modules/core"
)

type GooglePubsubConfig struct {
	ProjectId          string
	Topic              string
	Subscription       string
	BatchSize          int
	BatchDuration      time.Duration
	MessageChannelSize int
}

type GooglePubsubProducer struct {
	MessageChannel  chan []byte
	config          GooglePubsubConfig
	pubsubClient    *pubsub.Client
	pubsubTopic     *pubsub.Topic
	messageBatch    [][]byte
	batchStartTime  time.Time
	mutex           sync.RWMutex
	numMessagesSent int
	numBatchesSent  int
}

func CreateGooglePubsubProducer(ctx context.Context, config GooglePubsubConfig) (*GooglePubsubProducer, error) {

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 10 * 1024
	}

	if config.BatchDuration == 0 {
		config.BatchDuration = time.Second
	}

	pubsubClient, err := pubsub.NewClient(ctx, config.ProjectId)
	if err != nil {
		core.Error("failed to create google pubsub client: %v", err)
		return nil, err
	}

	pubsubTopic := pubsubClient.Topic(config.Topic)
	if pubsubTopic == nil {
		core.Error("failed to create google pubsub topic")
		return nil, err
	}

	producer := &GooglePubsubProducer{}

	producer.config = config
	producer.pubsubClient = pubsubClient
	producer.pubsubTopic = pubsubTopic
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

	_, err := producer.pubsubTopic.Publish(timeoutContext, &pubsub.Message{Data: messageToSend}).Get(ctx)
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
	producer.pubsubClient.Close()
}

// ----------------------------

type GooglePubsubConsumer struct {
	MessageChannel      chan *pubsub.Message
	config              GooglePubsubConfig
	pubsubClient        *pubsub.Client
	pubsubSubscription  *pubsub.Subscription
	mutex               sync.RWMutex
	numMessagesReceived int
	numBatchesReceived  int
	numBatchesACKd      int
	numBatchesNACKd     int
}

func CreateGooglePubsubConsumer(ctx context.Context, config GooglePubsubConfig) (*GooglePubsubConsumer, error) {

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 10 * 1024
	}

	pubsubClient, err := pubsub.NewClient(ctx, config.ProjectId)
	if err != nil {
		core.Error("failed to create google pubsub consumer: %v", err)
		return nil, err
	}

	pubsubSubscription := pubsubClient.Subscription(config.Subscription)
	if pubsubSubscription == nil {
		core.Error("failed to create google pubsub subscription")
		return nil, err
	}

	consumer := &GooglePubsubConsumer{}

	consumer.config = config
	consumer.pubsubClient = pubsubClient
	consumer.pubsubSubscription = pubsubSubscription

	consumer.MessageChannel = make(chan *pubsub.Message, config.MessageChannelSize)

	go consumer.receiveMessages(ctx)

	return consumer, nil
}

func (consumer *GooglePubsubConsumer) receiveMessages(ctx context.Context) {

	consumer.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {

		batchMessages := ParseMessages(m.Data)

		core.Debug("received %d messages (%v bytes) from google pubsub", len(batchMessages), len(m.Data))

		consumer.MessageChannel <- m

		consumer.mutex.Lock()
		consumer.numBatchesReceived += 1
		consumer.numMessagesReceived += len(batchMessages)
		consumer.mutex.Unlock()
	})
}

func (consumer *GooglePubsubConsumer) ACKMessage() {
	consumer.mutex.Lock()
	consumer.numBatchesACKd++
	consumer.mutex.Unlock()
}

func (consumer *GooglePubsubConsumer) NACKMessage() {
	consumer.mutex.Lock()
	consumer.numBatchesNACKd++
	consumer.mutex.Unlock()
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

func (consumer *GooglePubsubConsumer) NumBatchesACKd() int {
	consumer.mutex.RLock()
	value := consumer.numBatchesACKd
	consumer.mutex.RUnlock()
	return value
}

func (consumer *GooglePubsubConsumer) NumBatchesNACKd() int {
	consumer.mutex.RLock()
	value := consumer.numBatchesNACKd
	consumer.mutex.RUnlock()
	return value
}

func (consumer *GooglePubsubConsumer) Close(ctx context.Context) {
	consumer.pubsubClient.Close()
}
