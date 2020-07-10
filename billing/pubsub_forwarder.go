package billing

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/metrics"
)

type PubSubForwarder struct {
	Biller  Biller
	Logger  log.Logger
	Metrics *metrics.GooglePubSubForwarderMetrics

	pubsubSubscription *pubsub.Subscription
}

func NewPubSubForwarder(ctx context.Context, biller Biller, logger log.Logger, metrics *metrics.GooglePubSubForwarderMetrics, gcpProjectID string, topicName string, subscriptionName string) (*PubSubForwarder, error) {
	pubsubClient, err := pubsub.NewClient(ctx, gcpProjectID)
	if err != nil {
		return nil, fmt.Errorf("could not create pubsub client: %v", err)
	}

	if _, err := pubsubClient.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
		Topic: pubsubClient.Topic(topicName),
	}); err != nil && err.Error() != "rpc error: code = AlreadyExists desc = Subscription already exists" {
		// Not the best error check, but the underlying error type is internal so we can't check for it
		return nil, fmt.Errorf("could not create local pubsub subscription: %v", err)
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
		psf.Metrics.BillingEntriesReceived.Add(1)
		billingEntry := BillingEntry{}
		if ReadBillingEntry(&billingEntry, m.Data) {
			m.Ack()
			billingEntry.Timestamp = uint64(m.PublishTime.Unix())
			if err := psf.Biller.Bill(context.Background(), &billingEntry); err != nil {
				level.Error(psf.Logger).Log("msg", "could not submit billing entry", "err", err)
				psf.Metrics.ErrorMetrics.BillingWriteFailure.Add(1)
			}
		} else {
			psf.Metrics.ErrorMetrics.BillingReadFailure.Add(1)
		}
	})
	if err != context.Canceled {
		level.Error(psf.Logger).Log("msg", "could not setup to receive pubsub messages", "err", err)
		os.Exit(1)
	}
}
