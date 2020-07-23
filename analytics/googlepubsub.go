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
	NumSubmitted() uint64
	NumQueued() uint64
	NumFlushed() uint64
}

type NoOpPingStatsPublisher struct {
	submitted uint64
}

func (publisher *NoOpPingStatsPublisher) Publish(ctx context.Context, entries []PingStatsEntry) error {
	atomic.AddUint64(&publisher.submitted, uint64(len(entries)))
	return nil
}

func (publisher *NoOpPingStatsPublisher) NumSubmitted() uint64 {
	return atomic.LoadUint64(&publisher.submitted)
}

func (publisher *NoOpPingStatsPublisher) NumQueued() uint64 {
	return 0
}

func (publisher *NoOpPingStatsPublisher) NumFlushed() uint64 {
	return atomic.LoadUint64(&publisher.submitted)
}

type GooglePubSubClient struct {
	PubsubClient *pubsub.Client
	Topic        *pubsub.Topic
	ResultChan   chan *pubsub.PublishResult
	Metrics      *metrics.AnalyticsMetrics
}

func newGooglePubSubClient(ctx context.Context, statsMetrics *metrics.AnalyticsMetrics, projectID string, topicID string, settings pubsub.PublishSettings) (*GooglePubSubClient, error) {
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

	return client, nil
}

type GooglePubSubPingStatsPublisher struct {
	Logger log.Logger

	client    *GooglePubSubClient
	submitted uint64
	flushed   uint64
}

func NewGooglePubSubPingStatsPublisher(ctx context.Context, statsMetrics *metrics.AnalyticsMetrics, resultLogger log.Logger, projectID string, topicID string, settings pubsub.PublishSettings) (*GooglePubSubPingStatsPublisher, error) {
	publisher := &GooglePubSubPingStatsPublisher{
		Logger: resultLogger,
	}

	client, err := newGooglePubSubClient(ctx, statsMetrics, projectID, topicID, settings)
	if err != nil {
		return nil, err
	}
	publisher.client = client

	go client.pubsubResults(ctx, publisher)

	return publisher, nil
}

func (publisher *GooglePubSubPingStatsPublisher) Publish(ctx context.Context, entries []PingStatsEntry) error {
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

func (publisher *GooglePubSubPingStatsPublisher) NumSubmitted() uint64 {
	return atomic.LoadUint64(&publisher.submitted)
}

func (publisher *GooglePubSubPingStatsPublisher) NumQueued() uint64 {
	return atomic.LoadUint64(&publisher.submitted) - atomic.LoadUint64(&publisher.flushed)
}

func (publisher *GooglePubSubPingStatsPublisher) NumFlushed() uint64 {
	return atomic.LoadUint64(&publisher.flushed)
}

func (client *GooglePubSubClient) pubsubResults(ctx context.Context, publisher *GooglePubSubPingStatsPublisher) {
	for {
		select {
		case result := <-client.ResultChan:
			_, err := result.Get(ctx)
			if err != nil {
				level.Error(publisher.Logger).Log("analytics", "failed to publish to pubsub", "err", err)
				client.Metrics.PingStatsErrorMetrics.AnalyticsPublishFailure.Add(1)
			} else {
				level.Debug(publisher.Logger).Log("analytics", "successfully published analytics data")
				atomic.AddUint64(&publisher.flushed, 1)
			}
		case <-ctx.Done():
			level.Debug(publisher.Logger).Log("msg", "SHOULD NOT GET HERE")
			return
		}
	}
}

type RelayStatsPublisher interface {
	Publish(ctx context.Context, entries []RelayStatsEntry) error
	NumSubmitted() uint64
	NumQueued() uint64
	NumFlushed() uint64
}

type NoOpRelayStatsPublisher struct {
	submitted uint64
}

func (publisher *NoOpRelayStatsPublisher) Publish(ctx context.Context, entries []RelayStatsEntry) error {
	atomic.AddUint64(&publisher.submitted, uint64(len(entries)))
	return nil
}

func (publisher *NoOpRelayStatsPublisher) NumSubmitted() uint64 {
	return atomic.LoadUint64(&publisher.submitted)
}

func (publisher *NoOpRelayStatsPublisher) NumQueued() uint64 {
	return 0
}

func (publisher *NoOpRelayStatsPublisher) NumFlushed() uint64 {
	return atomic.LoadUint64(&publisher.submitted)
}
