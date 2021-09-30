package analytics

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"
)

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
	client.ResultChan = make(chan *pubsub.PublishResult)

	return client, nil
}

func (client *googlePubSubClient) pubsubResults(ctx context.Context) {
	for {
		select {
		case result := <-client.ResultChan:
			_, err := result.Get(ctx)
			if err != nil {
				core.Error("failed to publish to Google Pub/Sub: %v", err)
				client.Metrics.ErrorMetrics.PublishFailure.Add(1)
			} else {
				core.Debug("successfully published to Google Pub/Sub")
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

// NewGooglePubSubPingStatsPublisher() returns a GooglePubSubPingStatsPublisher that publishes ping stats to Google Pub/Sub
// TODO: remove resultLogger once Analytics Pusher no longer uses the gokit logger
func NewGooglePubSubPingStatsPublisher(ctx context.Context, statsMetrics *metrics.AnalyticsMetrics, resultLogger log.Logger, projectID string, topicID string, settings pubsub.PublishSettings) (*GooglePubSubPingStatsPublisher, error) {
	publisher := &GooglePubSubPingStatsPublisher{}

	client, err := newGooglePubSubClient(ctx, statsMetrics, projectID, topicID, settings)
	if err != nil {
		return nil, err
	}
	publisher.client = client

	go client.pubsubResults(ctx)

	return publisher, nil
}

func (publisher *GooglePubSubPingStatsPublisher) Publish(ctx context.Context, entries []PingStatsEntry) error {
	data := WritePingStatsEntries(entries)

	if publisher.client == nil {
		return fmt.Errorf("analytics: ping stats pub/sub client not initialized")
	}

	topic := publisher.client.Topic
	resultChan := publisher.client.ResultChan

	result := topic.Publish(ctx, &pubsub.Message{Data: data})

	if result != nil {
		resultChan <- result
		publisher.client.Metrics.EntriesSubmitted.Add(float64(len(entries)))
	}

	return nil
}

// TODO: call Close() when context is canceled in the Analytics Pusher
func (publisher *GooglePubSubPingStatsPublisher) Close() {
	publisher.client.Topic.Stop()
}

type GooglePubSubRelayStatsPublisher struct {
	client *googlePubSubClient
}

// NewGooglePubSubPingStatsPublisher() returns a GooglePubSubRelayStatsPublisher that publishes relay stats to Google Pub/Sub
// TODO: remove resultLogger once Analytics Pusher no longer uses the gokit logger
func NewGooglePubSubRelayStatsPublisher(ctx context.Context, statsMetrics *metrics.AnalyticsMetrics, resultLogger log.Logger, projectID string, topicID string, settings pubsub.PublishSettings) (*GooglePubSubRelayStatsPublisher, error) {
	publisher := &GooglePubSubRelayStatsPublisher{}

	client, err := newGooglePubSubClient(ctx, statsMetrics, projectID, topicID, settings)
	if err != nil {
		return nil, err
	}
	publisher.client = client

	go client.pubsubResults(ctx)

	return publisher, nil
}

func (publisher *GooglePubSubRelayStatsPublisher) Publish(ctx context.Context, entries []RelayStatsEntry) error {
	data := WriteRelayStatsEntries(entries)

	if publisher.client == nil {
		return fmt.Errorf("analytics: relay stats pub/sub client not initialized")
	}

	topic := publisher.client.Topic
	resultChan := publisher.client.ResultChan

	result := topic.Publish(ctx, &pubsub.Message{Data: data})

	if result != nil {
		resultChan <- result
		publisher.client.Metrics.EntriesSubmitted.Add(float64(len(entries)))
	}

	return nil
}

// TODO: call Close() when context is canceled in the Analytics Pusher
func (publisher *GooglePubSubRelayStatsPublisher) Close() {
	publisher.client.Topic.Stop()
}
