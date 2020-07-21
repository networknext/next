package analytics

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/metrics"
)

type PubSubWriter interface {
	Write(ctx context.Context, entry *StatsEntry) error
}

type NoOpPubSubWriter struct {
	written uint64
}

func (writer *NoOpPubSubWriter) Write(ctx context.Context, entry StatsEntry) error {
	atomic.AddUint64(&writer.written, 1)
	return nil
}

type GooglePubSubWriter struct {
	Logger log.Logger

	clients   []*GooglePubSubClient
	submitted uint64
	flushed   uint64
}

type GooglePubSubClient struct {
	PubsubClient *pubsub.Client
	Topic        *pubsub.Topic
	ResultChan   chan *pubsub.PublishResult
	Metrics      *metrics.AnalyticsMetrics
}

func NewGooglePubSubWriter(ctx context.Context, statsMetrics *metrics.AnalyticsMetrics, resultLogger log.Logger, projectID string, topicID string, clientCount int, clientChanBufferSize int, settings *pubsub.PublishSettings) (*GooglePubSubWriter, error) {
	if settings == nil {
		return nil, errors.New("nil google pubsub publish settings")
	}

	writer := &GooglePubSubWriter{
		Logger:  resultLogger,
		clients: make([]*GooglePubSubClient, clientCount),
	}

	for i := 0; i < clientCount; i++ {
		var err error
		client := &GooglePubSubClient{}
		client.PubsubClient, err = pubsub.NewClient(ctx, projectID)
		client.Metrics = statsMetrics
		if err != nil {
			return nil, fmt.Errorf("could not create pubsub client %d: %v", i, err)
		}

		if projectID == "local" {
			if _, err := client.PubsubClient.CreateTopic(ctx, topicID); err != nil {
				if err.Error() != "rpc error: code = AlreadyExists desc = Topic already exists" {
					return nil, err
				}
			}
		}

		client.Topic = client.PubsubClient.Topic(topicID)
		client.Topic.PublishSettings = *settings
		client.ResultChan = make(chan *pubsub.PublishResult, clientChanBufferSize)

		go client.pubsubResults(ctx, writer)

		writer.clients[i] = client
	}

	return writer, nil
}

func (writer *GooglePubSubWriter) Write(ctx context.Context, entry *StatsEntry) error {
	atomic.AddUint64(&writer.submitted, 1)

	data := WriteStatsEntry(entry)

	if writer.clients == nil {
		return fmt.Errorf("statsdb: clients not initialized")
	}

	index := entry.RelayA % uint64(len(writer.clients))
	topic := writer.clients[index].Topic
	resultChan := writer.clients[index].ResultChan

	result := topic.Publish(ctx, &pubsub.Message{Data: data})
	resultChan <- result

	return nil
}

func (client *GooglePubSubClient) pubsubResults(ctx context.Context, writer *GooglePubSubWriter) {
	for {
		select {
		case result := <-client.ResultChan:
			_, err := result.Get(ctx)
			if err != nil {
				level.Error(writer.Logger).Log("statsdb", "failed to publish to pubsub", "err", err)
				client.Metrics.ErrorMetrics.AnalyticsPublishFailure.Add(1)
			} else {
				level.Debug(writer.Logger).Log("statsdb", "successfully published billing data")
				atomic.AddUint64(&writer.flushed, 1)
			}
		case <-ctx.Done():
			return
		}
	}
}
