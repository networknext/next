package billing

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

// GooglePubSubBiller is an implementation of a billing handler that sends billing data to Google Pub/Sub through multiple clients
type GooglePubSubBiller struct {
	Logger log.Logger

	clients   []*GooglePubSubClient
	submitted uint64
	flushed   uint64
}

// GooglePubSubClient represents a single client that publishes billing data
type GooglePubSubClient struct {
	PubsubClient *pubsub.Client
	Topic        *pubsub.Topic
	ResultChan   chan *pubsub.PublishResult
	Metrics      *metrics.BillingMetrics
}

// NewBiller creates a new GooglePubSubBiller, sets up the pubsub clients, and starts goroutines to listen for publish results.
// To clean up the results goroutine, use ctx.Done().
func NewGooglePubSubBiller(ctx context.Context, billingMetrics *metrics.BillingMetrics, resultLogger log.Logger, projectID string, billingTopicID string, clientCount int, clientChanBufferSize int, settings *pubsub.PublishSettings) (Biller, error) {
	if settings == nil {
		return nil, errors.New("nil google pubsub publish settings")
	}

	biller := &GooglePubSubBiller{
		Logger:  resultLogger,
		clients: make([]*GooglePubSubClient, clientCount),
	}

	for i := 0; i < clientCount; i++ {

		var err error
		biller.clients[i] = &GooglePubSubClient{}
		biller.clients[i].PubsubClient, err = pubsub.NewClient(ctx, projectID)
		biller.clients[i].Metrics = billingMetrics
		if err != nil {
			return nil, fmt.Errorf("could not create pubsub client %v: %v", i, err)
		}

		// Create the billing topic if running locally with the pubsub emulator
		if projectID == "local" {
			if _, err := biller.clients[i].PubsubClient.CreateTopic(ctx, billingTopicID); err != nil {
				// Not the best, but the underlying error type is internal so we can't check for it
				if err.Error() != "rpc error: code = AlreadyExists desc = Topic already exists" {
					return nil, err
				}
			}
		}

		biller.clients[i].Topic = biller.clients[i].PubsubClient.Topic(billingTopicID)
		biller.clients[i].Topic.PublishSettings = *settings
		biller.clients[i].ResultChan = make(chan *pubsub.PublishResult, clientChanBufferSize)

		go biller.clients[i].pubsubResults(ctx, biller)
	}

	return biller, nil
}

func (biller *GooglePubSubBiller) Bill(ctx context.Context, entry *BillingEntry) error {

	atomic.AddUint64(&biller.submitted, 1)

	data := WriteBillingEntry(entry)

	if biller.clients == nil {
		return fmt.Errorf("billing: clients not initialized")
	}

	index := entry.SessionID % uint64(len(biller.clients))
	topic := biller.clients[index].Topic
	resultChan := biller.clients[index].ResultChan

	result := topic.Publish(ctx, &pubsub.Message{Data: data})
	resultChan <- result

	return nil
}

func (biller *GooglePubSubBiller) NumSubmitted() uint64 {
	return atomic.LoadUint64(&biller.submitted)
}

func (biller *GooglePubSubBiller) NumQueued() uint64 {
	return atomic.LoadUint64(&biller.submitted) - atomic.LoadUint64(&biller.flushed)
}

func (biller *GooglePubSubBiller) NumFlushed() uint64 {
	return atomic.LoadUint64(&biller.flushed)
}

func (client *GooglePubSubClient) pubsubResults(ctx context.Context, biller *GooglePubSubBiller) {
	for {
		select {
		case result := <-client.ResultChan:
			_, err := result.Get(ctx)
			if err != nil {
				level.Error(biller.Logger).Log("billing", "failed to publish to pub/sub", "err", err)
				client.Metrics.ErrorMetrics.BillingPublishFailure.Add(1)
			} else {
				level.Debug(biller.Logger).Log("billing", "successfully published billing data")
				atomic.AddUint64(&biller.flushed, 1)
			}
		case <-ctx.Done():
			return
		}
	}
}
