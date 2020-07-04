package billing

import (
	"context"
	"fmt"
	"sync/atomic"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"

	"github.com/go-kit/kit/log"
	// "github.com/go-kit/kit/log/level"
)

// GooglePubSubBiller is an implementation of a billing handler that sends billing data to Google Pub/Sub through multiple clients
type GooglePubSubBiller struct {
	clients []*GooglePubSubClient
	submitted uint64
	flushed   uint64
}

// GooglePubSubClient represents a single client that publishes billing data
type GooglePubSubClient struct {
	PubsubClient *pubsub.Client
	Topic        *pubsub.Topic
	ResultChan   chan *pubsub.PublishResult
}

// NewBiller creates a new GooglePubSubBiller, sets up the pubsub clients, and starts goroutines to listen for publish results.
// To clean up the results goroutine, use ctx.Done().
func NewBiller(ctx context.Context, resultLogger log.Logger, projectID string, billingTopicID string, descriptor *Descriptor) (Biller, error) {
	
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
		biller.clients[i].PubsubClient, err = pubsub.NewClient(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("could not create pubsub client %v: %v", i, err)
		}

		biller.clients[i].Topic = biller.clients[i].PubsubClient.Topic(billingTopicID)

		if descriptor.CountThreshold > pubsub.MaxPublishRequestCount {
			descriptor.CountThreshold = pubsub.MaxPublishRequestCount
		}

		if descriptor.ByteThreshold > pubsub.MaxPublishRequestBytes {
			descriptor.ByteThreshold = pubsub.MaxPublishRequestBytes
		}

		biller.clients[i].Topic.PublishSettings = pubsub.PublishSettings{
			DelayThreshold:    descriptor.DelayThreshold,
			CountThreshold:    descriptor.CountThreshold,
			ByteThreshold:     descriptor.ByteThreshold,
			NumGoroutines:     descriptor.NumGoroutines,
			Timeout:           descriptor.Timeout,
			BufferedByteLimit: pubsub.DefaultPublishSettings.BufferedByteLimit,
		}

		biller.clients[i].ResultChan = make(chan *pubsub.PublishResult, descriptor.ResultChannelBuffer)

		go pubsubResults(biller, ctx, resultLogger, biller.clients[i].ResultChan)
	}

	return biller, nil
}

// Bill sends the billing entry to Google Pub/Sub
func (biller *GooglePubSubBiller) Bill(ctx context.Context, sessionID uint64, entry *Entry) error {

	atomic.AddUint64(&biller.submitted, 1)

	data, err := proto.Marshal(entry)
	if err != nil {
		return err
	}

	if biller.clients == nil {
		return fmt.Errorf("billing: clients not initialized")
	}

	index := sessionID % uint64(len(biller.clients))
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

func pubsubResults(biller *GooglePubSubBiller, ctx context.Context, logger log.Logger, results chan *pubsub.PublishResult) {
	for {
		select {
		case result := <-results:
			_, err := result.Get(ctx)
			if err != nil {
				// level.Error(logger).Log("billing", "failed to publish to pub/sub", "err", err)
				// todo: ryan, please increase pubsub error count metric
			} else {
				// level.Debug(logger).Log("billing", "successfully published billing data")
				atomic.AddUint64(&biller.flushed, 1)
			}
		case <-ctx.Done():
			return
		}
	}
}
