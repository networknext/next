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

type PingStatsPublisher interface {
	Publish(ctx context.Context, entries []PingStatsEntry) error
}

type NoOpPingStatsPublisher struct {
	submitted uint64
}

func (publisher *NoOpPingStatsPublisher) Publish(ctx context.Context, entries []PingStatsEntry) error {
	atomic.AddUint64(&publisher.submitted, uint64(len(entries)))
	return nil
}

type googlePubSubClient struct {
	PubsubClient *pubsub.Client
	Topic        *pubsub.Topic
	ResultChan   chan *pubsub.PublishResult
	Metrics      *metrics.AnalyticsMetrics
}

func newGooglePubSubClient(ctx context.Context, statsMetrics *metrics.AnalyticsMetrics, projectID string, topicID string, settings pubsub.PublishSettings) (*googlePubSubClient, error) {
	var err error

	client := &googlePubSubClient{}

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

	return client, nil
}

func (client *googlePubSubClient) pubsubResults(ctx context.Context, logger log.Logger) {
	for {
		select {
		case result := <-client.ResultChan:
			_, err := result.Get(ctx)
			if err != nil {
				level.Error(logger).Log("analytics", "failed to publish to pubsub", "err", err)
				client.Metrics.ErrorMetrics.PublishFailure.Add(1)
			} else {
				level.Debug(logger).Log("analytics", "successfully published analytics data")
				client.Metrics.EntriesFlushed.Add(1)
			}
		case <-ctx.Done():
			return
		}
	}
}

type GooglePubSubPingStatsPublisher struct {
	client *googlePubSubClient
}

func NewGooglePubSubPingStatsPublisher(ctx context.Context, statsMetrics *metrics.AnalyticsMetrics, resultLogger log.Logger, projectID string, topicID string, settings pubsub.PublishSettings) (*GooglePubSubPingStatsPublisher, error) {
	publisher := &GooglePubSubPingStatsPublisher{}

	client, err := newGooglePubSubClient(ctx, statsMetrics, projectID, topicID, settings)
	if err != nil {
		return nil, err
	}
	publisher.client = client

	go client.pubsubResults(ctx, resultLogger)

	return publisher, nil
}

func (publisher *GooglePubSubPingStatsPublisher) Publish(ctx context.Context, entries []PingStatsEntry) error {
	data := WritePingStatsEntries(entries)

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

type RelayStatsPublisher interface {
	Publish(ctx context.Context, entries []RelayStatsEntry) error
}

type NoOpRelayStatsPublisher struct {
	submitted uint64
}

func (publisher *NoOpRelayStatsPublisher) Publish(ctx context.Context, entries []RelayStatsEntry) error {
	atomic.AddUint64(&publisher.submitted, uint64(len(entries)))
	return nil
}

type GooglePubSubRelayStatsPublisher struct {
	client *googlePubSubClient
}

func NewGooglePubSubRelayStatsPublisher(ctx context.Context, statsMetrics *metrics.AnalyticsMetrics, resultLogger log.Logger, projectID string, topicID string, settings pubsub.PublishSettings) (*GooglePubSubRelayStatsPublisher, error) {
	publisher := &GooglePubSubRelayStatsPublisher{}

	client, err := newGooglePubSubClient(ctx, statsMetrics, projectID, topicID, settings)
	if err != nil {
		return nil, err
	}
	publisher.client = client

	go client.pubsubResults(ctx, resultLogger)

	return publisher, nil
}

func (publisher *GooglePubSubRelayStatsPublisher) Publish(ctx context.Context, entries []RelayStatsEntry) error {
	data := WriteRelayStatsEntries(entries)

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
