package billing

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/metrics"
)

type PubSubForwarder struct {
	Biller  Biller
	Logger  log.Logger
	Metrics *metrics.BillingMetrics

	pubsubSubscription *pubsub.Subscription
}

func NewPubSubForwarder(ctx context.Context, biller Biller, logger log.Logger, metrics *metrics.BillingMetrics, gcpProjectID string, topicName string, subscriptionName string) (*PubSubForwarder, error) {
	pubsubClient, err := pubsub.NewClient(ctx, gcpProjectID)
	if err != nil {
		return nil, fmt.Errorf("could not create pubsub client: %v", err)
	}

	if gcpProjectID == "local" {
		if _, err := pubsubClient.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic: pubsubClient.Topic(topicName),
		}); err != nil && err.Error() != "rpc error: code = AlreadyExists desc = Subscription already exists" {
			// Not the best error check, but the underlying error type is internal so we can't check for it
			return nil, fmt.Errorf("could not create local pubsub subscription: %v", err)
		}
	}

	return &PubSubForwarder{
		Biller:             biller,
		Logger:             logger,
		Metrics:            metrics,
		pubsubSubscription: pubsubClient.Subscription(subscriptionName),
	}, nil
}

// Forward reads the billing entry from pubsub and writes it to BigQuery
func (psf *PubSubForwarder) Forward(ctx context.Context) {
	err := psf.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		// Check if the message is batched or unbatched (for compatibility with old data)
		var messageLength uint32
		var offset int
		if !encoding.ReadUint32(m.Data, &offset, &messageLength) {
			level.Error(psf.Logger).Log("msg", "failed to detect if message is batched or unbatched", "offset", offset, "length", len(m.Data))
			psf.Metrics.ErrorMetrics.BillingBatchedReadFailure.Add(1)
		}

		var entries [][]byte
		if messageLength <= uint32(len(m.Data)) {
			// This is a new, batched message
			var err error
			entries, err = psf.unbatchMessages(m)
			if err != nil {
				level.Error(psf.Logger).Log("err", err)
				psf.Metrics.ErrorMetrics.BillingBatchedReadFailure.Add(1)
			}
		} else {
			// This is an old, unbatched message. ignore it
			return
		}

		psf.Metrics.EntriesReceived.Add(float64(len(entries)))

		billingEntries := make([]BillingEntry, len(entries))
		for i := range billingEntries {
			if ReadBillingEntry(&billingEntries[i], entries[i]) {
				m.Ack()
				billingEntries[i].Timestamp = uint64(m.PublishTime.Unix())
				if err := psf.Biller.Bill(context.Background(), &billingEntries[i]); err != nil {
					level.Error(psf.Logger).Log("msg", "could not submit billing entry", "err", err)
				}
			} else {
				psf.Metrics.ErrorMetrics.BillingReadFailure.Add(1)
			}
		}
	})
	if err != context.Canceled {
		level.Error(psf.Logger).Log("msg", "could not setup to receive pubsub messages", "err", err)
		os.Exit(1)
	}
}

func (psf *PubSubForwarder) unbatchMessages(m *pubsub.Message) ([][]byte, error) {
	messages := make([][]byte, 0)

	var offset int
	for {
		if offset >= len(m.Data) {
			break
		}

		var messageLength uint32
		var message []byte
		if !encoding.ReadUint32(m.Data, &offset, &messageLength) {
			return nil, fmt.Errorf("failed to read batched message length at offset %d (length %d)", offset, len(m.Data))
		}

		if !encoding.ReadBytes(m.Data, &offset, &message, messageLength) {
			return nil, fmt.Errorf("failed to read batched message bytes at offset %d (length %d)", offset, len(m.Data))
		}

		messages = append(messages, message)
	}

	return messages, nil
}
