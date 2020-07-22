package analytics

import (
	"context"
	"fmt"
	"sync/atomic"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/metrics"
)

type PubSubPublisher interface {
	Publish(ctx context.Context, entries []StatsEntry) error
	NumSubmitted() uint64
	NumQueued() uint64
	NumFlushed() uint64
}

type NoOpPubSubPublisher struct {
	submitted uint64
}

func (publisher *NoOpPubSubPublisher) Publish(ctx context.Context, entries []StatsEntry) error {
	atomic.AddUint64(&publisher.submitted, uint64(len(entries)))
	return nil
}

func (publisher *NoOpPubSubPublisher) NumSubmitted() uint64 {
	return atomic.LoadUint64(&publisher.submitted)
}

func (publisher *NoOpPubSubPublisher) NumQueued() uint64 {
	return 0
}

func (publisher *NoOpPubSubPublisher) NumFlushed() uint64 {
	return atomic.LoadUint64(&publisher.submitted)
}

type GooglePubSubPublisher struct {
	Logger log.Logger

	client    *GooglePubSubClient
	submitted uint64
	flushed   uint64
}

type GooglePubSubClient struct {
	PubsubClient *pubsub.Client
	Topic        *pubsub.Topic
	ResultChan   chan *pubsub.PublishResult
	Metrics      *metrics.AnalyticsMetrics
}

func NewGooglePubSubPublisher(ctx context.Context, statsMetrics *metrics.AnalyticsMetrics, resultLogger log.Logger, projectID string, topicID string, settings pubsub.PublishSettings) (*GooglePubSubPublisher, error) {
	publisher := &GooglePubSubPublisher{
		Logger: resultLogger,
	}

	var err error
	client := &GooglePubSubClient{}

	client.PubsubClient, err = pubsub.NewClient(ctx, projectID)

	if err != nil {
		return nil, fmt.Errorf("could not create pubsub client: %v", err)
	}

	if projectID == "local" {
		if _, err := client.PubsubClient.CreateTopic(ctx, topicID); err != nil {
			if err.Error() != "rpc error: code = AlreadyExists desc = Topic already exists" {
				return nil, err
			}
		}
	}

	client.Metrics = statsMetrics
	client.Topic = client.PubsubClient.Topic(topicID)
	client.Topic.PublishSettings = settings
	client.ResultChan = make(chan *pubsub.PublishResult, 1)

	go client.pubsubResults(ctx, publisher)

	publisher.client = client

	return publisher, nil
}

func (publisher *GooglePubSubPublisher) Publish(ctx context.Context, entries []StatsEntry) error {
	atomic.AddUint64(&publisher.submitted, 1)

	data := WriteStatsEntries(entries)

	if publisher.client == nil {
		return fmt.Errorf("analytics: client not initialized")
	}

	topic := publisher.client.Topic
	resultChan := publisher.client.ResultChan

	result := topic.Publish(ctx, &pubsub.Message{Data: data})
	resultChan <- result

	return nil
}

func (publisher *GooglePubSubPublisher) NumSubmitted() uint64 {
	return atomic.LoadUint64(&publisher.submitted)
}

func (publisher *GooglePubSubPublisher) NumQueued() uint64 {
	return atomic.LoadUint64(&publisher.submitted) - atomic.LoadUint64(&publisher.flushed)
}

func (publisher *GooglePubSubPublisher) NumFlushed() uint64 {
	return atomic.LoadUint64(&publisher.flushed)
}

func (client *GooglePubSubClient) pubsubResults(ctx context.Context, publisher *GooglePubSubPublisher) {
	for {
		select {
		case result := <-client.ResultChan:
			_, err := result.Get(ctx)
			if err != nil {
				level.Error(publisher.Logger).Log("analytics", "failed to publish to pubsub", "err", err)
				client.Metrics.ErrorMetrics.AnalyticsPublishFailure.Add(1)
			} else {
				level.Debug(publisher.Logger).Log("analytics", "successfully published billing data")
				atomic.AddUint64(&publisher.flushed, 1)
			}
		case <-ctx.Done():
			return
		}
	}
}
