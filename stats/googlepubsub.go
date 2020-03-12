package stats

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gogo/protobuf/proto"
	"github.com/networknext/backend/billing"
)

// GooglePubSubTrafficStatsPublisher ...
type GooglePubSubTrafficStatsPublisher struct {
	clients []*billing.GooglePubSubClient
}

// NewTrafficStatsPublisher ...
func NewTrafficStatsPublisher(ctx context.Context, resultLogger log.Logger, projectID string, statsTopicID string, descriptor *billing.Descriptor) (*GooglePubSubTrafficStatsPublisher, error) {
	var clientCount int
	if descriptor != nil {
		clientCount = descriptor.ClientCount
	}

	publisher := &GooglePubSubTrafficStatsPublisher{
		clients: make([]*billing.GooglePubSubClient, clientCount),
	}

	for i := 0; i < clientCount; i++ {
		var err error
		publisher.clients[i] = &billing.GooglePubSubClient{}
		publisher.clients[i].PubsubClient, err = pubsub.NewClient(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("could not create pubsub client %v: %v", i, err)
		}

		publisher.clients[i].Topic = publisher.clients[i].PubsubClient.Topic(statsTopicID)

		if descriptor.CountThreshold > pubsub.MaxPublishRequestCount {
			descriptor.CountThreshold = pubsub.MaxPublishRequestCount
		}

		if descriptor.ByteThreshold > pubsub.MaxPublishRequestBytes {
			descriptor.ByteThreshold = pubsub.MaxPublishRequestBytes
		}

		publisher.clients[i].Topic.PublishSettings = pubsub.PublishSettings{
			DelayThreshold:    descriptor.DelayThreshold,
			CountThreshold:    descriptor.CountThreshold,
			ByteThreshold:     descriptor.ByteThreshold,
			NumGoroutines:     descriptor.NumGoroutines,
			Timeout:           descriptor.Timeout,
			BufferedByteLimit: pubsub.DefaultPublishSettings.BufferedByteLimit,
		}
		publisher.clients[i].ResultChan = make(chan *pubsub.PublishResult, descriptor.ResultChannelBuffer)
		go printPubSubResults(ctx, resultLogger, publisher.clients[i].ResultChan)
	}

	return publisher, nil
}

func (gps *GooglePubSubTrafficStatsPublisher) Publish(ctx context.Context, relayID uint64, entry *RelayTrafficStats) error {
	data, err := proto.Marshal(entry)
	if err != nil {
		return err
	}

	if gps.clients == nil {
		return fmt.Errorf("billing: clients not initialized")
	}

	index := relayID % uint64(len(gps.clients))
	topic := gps.clients[index].Topic
	resultChan := gps.clients[index].ResultChan

	result := topic.Publish(ctx, &pubsub.Message{Data: data})
	resultChan <- result
	return nil
}

func printPubSubResults(ctx context.Context, logger log.Logger, results chan *pubsub.PublishResult) {
	for {
		select {
		case result := <-results:
			_, err := result.Get(ctx)
			if err != nil {
				level.Error(logger).Log("traffic_stats", "failed to publish to pub/sub", "err", err)
			} else {
				level.Debug(logger).Log("traffic_stats", "successfully published traffic stats data")
			}
		case <-ctx.Done():
			return
		}
	}
}
