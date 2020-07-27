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
}

type NoOpPubSubPublisher struct {
	submitted uint64
}

func (publisher *NoOpPubSubPublisher) Publish(ctx context.Context, entries []StatsEntry) error {
	atomic.AddUint64(&publisher.submitted, uint64(len(entries)))
	return nil
}

type GooglePubSubPublisher struct {
	Logger log.Logger
	client *GooglePubSubClient
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
	publisher.client = client

	go client.pubsubResults(ctx, publisher)

	return publisher, nil
}

func (publisher *GooglePubSubPublisher) Publish(ctx context.Context, entries []StatsEntry) error {
	data := WriteStatsEntries(entries)

	if publisher.client == nil {
		return fmt.Errorf("analytics: client not initialized")
	}

	topic := publisher.client.Topic
	resultChan := publisher.client.ResultChan

	result := topic.Publish(ctx, &pubsub.Message{Data: data})
	resultChan <- result

	publisher.client.Metrics.EntriesSubmitted.Add(1)

	return nil
}

func (client *GooglePubSubClient) pubsubResults(ctx context.Context, publisher *GooglePubSubPublisher) {
	for {
		select {
		case result := <-client.ResultChan:
			_, err := result.Get(ctx)
			if err != nil {
				level.Error(publisher.Logger).Log("analytics", "failed to publish to pubsub", "err", err)
				client.Metrics.ErrorMetrics.PublishFailure.Add(1)
			} else {
				level.Debug(publisher.Logger).Log("analytics", "successfully published analytics data")
				publisher.client.Metrics.EntriesFlushed.Add(1)
			}
		case <-ctx.Done():
			level.Debug(publisher.Logger).Log("msg", "SHOULD NOT GET HERE")
			return
		}
	}
}
