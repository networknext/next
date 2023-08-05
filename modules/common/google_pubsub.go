package common

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/networknext/next/modules/core"
	"google.golang.org/api/option"
)

type GooglePubsubConfig struct {
	ProjectId          string
	Topic              string
	Subscription       string
	ClientOptions      []option.ClientOption
	BatchSize          int
	BatchDuration      time.Duration
	MessageChannelSize int
}

type GooglePubsubProducer struct {
	MessageChannel  chan []byte
	resultChannel   chan *pubsub.PublishResult
	config          GooglePubsubConfig
	pubsubClient    *pubsub.Client
	pubsubTopic     *pubsub.Topic
	batchStartTime  time.Time
	mutex           sync.RWMutex
	numMessagesSent int
	numBatchesSent  int
}

func CreateGooglePubsubProducer(ctx context.Context, config GooglePubsubConfig) (*GooglePubsubProducer, error) {

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 1024
	}

	if config.BatchDuration == 0 {
		config.BatchDuration = time.Second
	}

	if config.BatchSize == 0 {
		config.BatchSize = 1000
	}

	pubsubClient, err := pubsub.NewClient(ctx, config.ProjectId, config.ClientOptions...)
	if err != nil {
		core.Error("failed to create google pubsub client: %v", err)
		return nil, err
	}

	pubsubTopic := pubsubClient.Topic(config.Topic)
	if pubsubTopic == nil {
		core.Error("failed to create google pubsub topic")
		return nil, err
	}

	pubsubTopic.PublishSettings.CountThreshold = config.BatchSize
	pubsubTopic.PublishSettings.DelayThreshold = config.BatchDuration

	producer := &GooglePubsubProducer{}

	producer.config = config
	producer.pubsubClient = pubsubClient
	producer.pubsubTopic = pubsubTopic
	producer.MessageChannel = make(chan []byte, config.MessageChannelSize)
	producer.resultChannel = make(chan *pubsub.PublishResult, config.MessageChannelSize)

	go producer.monitorResults(ctx)

	go producer.updateMessageChannel(ctx)

	return producer, nil
}

func (producer *GooglePubsubProducer) monitorResults(ctx context.Context) {

	for {
		select {

		case <-ctx.Done():
			return

		case result := <-producer.resultChannel:
			_, err := result.Get(ctx)
			if err != nil {
				core.Error("failed to send message batch: %v", err)
				break
			}

			producer.mutex.Lock()
			producer.numBatchesSent++
			producer.mutex.Unlock()
		}
	}
}

func (producer *GooglePubsubProducer) updateMessageChannel(ctx context.Context) {

	for {
		select {

		case <-ctx.Done():
			return

		case message := <-producer.MessageChannel:
			producer.sendMessage(ctx, message)
			break
		}
	}
}

func (producer *GooglePubsubProducer) sendMessage(ctx context.Context, message []byte) {

	result := producer.pubsubTopic.Publish(ctx, &pubsub.Message{Data: message})

	producer.resultChannel <- result

	producer.mutex.Lock()
	producer.numMessagesSent++
	producer.mutex.Unlock()
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
		consumer.MessageChannel <- m
		consumer.mutex.Lock()
		consumer.numMessagesReceived++
		consumer.mutex.Unlock()
	})
}

func (consumer *GooglePubsubConsumer) NumMessageReceived() int {
	consumer.mutex.RLock()
	value := consumer.numMessagesReceived
	consumer.mutex.RUnlock()
	return value
}

func (consumer *GooglePubsubConsumer) Close(ctx context.Context) {
	consumer.pubsubClient.Close()
}
