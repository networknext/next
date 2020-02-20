package billing

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/proto"
	"google.golang.org/api/option"
)

// GooglePubSubBiller is an implementation of a billing handler that sends billing data to Google Pub/Sub through multiple clients
type GooglePubSubBiller struct {
	clients []*GooglePubSubClient
}

// GooglePubSubClient represents a single client that publishes billing data
type GooglePubSubClient struct {
	pubsubClient *pubsub.Client
	topic        *pubsub.Topic
	resultChan   chan *pubsub.PublishResult
}

// NewBiller creates a new GooglePubSubBiller, sets up the pubsub clients, and starts goroutines to listen for publish results.
// To clean up the results goroutine, use ctx.Done().
func NewBiller(ctx context.Context, resultLogger log.Logger, projectID string, billingTopicID string, credentials []byte, descriptor *Descriptor) (Biller, error) {
	var clientCount int
	if descriptor != nil {
		clientCount = descriptor.ClientCount
	}

	biller := &GooglePubSubBiller{
		clients: make([]*GooglePubSubClient, clientCount),
	}

	for i := 0; i < clientCount; i++ {
		var err error
		biller.clients[i] = &GooglePubSubClient{}
		biller.clients[i].pubsubClient, err = pubsub.NewClient(ctx, projectID, option.WithCredentialsJSON(credentials))
		if err != nil {
			return nil, fmt.Errorf("could not create pubsub client %v: %v", i, err)
		}

		biller.clients[i].topic = biller.clients[i].pubsubClient.Topic(billingTopicID)

		if descriptor.CountThreshold > pubsub.MaxPublishRequestCount {
			descriptor.CountThreshold = pubsub.MaxPublishRequestCount
		}

		if descriptor.ByteThreshold > pubsub.MaxPublishRequestBytes {
			descriptor.ByteThreshold = pubsub.MaxPublishRequestBytes
		}

		biller.clients[i].topic.PublishSettings = pubsub.PublishSettings{
			DelayThreshold:    descriptor.DelayThreshold,
			CountThreshold:    descriptor.CountThreshold,
			ByteThreshold:     descriptor.ByteThreshold,
			NumGoroutines:     descriptor.NumGoroutines,
			Timeout:           descriptor.Timeout,
			BufferedByteLimit: pubsub.DefaultPublishSettings.BufferedByteLimit,
		}
		biller.clients[i].resultChan = make(chan *pubsub.PublishResult, descriptor.ResultChannelBuffer)
		go printPubSubResults(ctx, resultLogger, biller.clients[i].resultChan)
	}

	return biller, nil
}

// Bill sends the billing entry to Google Pub/Sub
func (biller *GooglePubSubBiller) Bill(ctx context.Context, sessionID uint64, entry *Entry) error {
	data, err := proto.Marshal(entry)
	if err != nil {
		return err
	}

	if biller.clients == nil {
		return fmt.Errorf("billing: clients not initialized")
	}

	index := sessionID % uint64(len(biller.clients))
	topic := biller.clients[index].topic
	resultChan := biller.clients[index].resultChan

	result := topic.Publish(ctx, &pubsub.Message{Data: data})
	resultChan <- result
	return nil
}

func printPubSubResults(ctx context.Context, logger log.Logger, results chan *pubsub.PublishResult) {
	select {
	case result := <-results:
		_, err := result.Get(ctx)
		if err != nil {
			level.Error(logger).Log("billing", "failed to publish to pub/sub", "err", err)
		} else {
			level.Debug(logger).Log("billing", "successfully pushed billing data")
		}
	case <-ctx.Done():
		return
	}
}
