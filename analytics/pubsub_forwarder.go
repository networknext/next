package analytics

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
	Writer  BigQueryWriter
	Logger  log.Logger
	Metrics *metrics.AnalyticsMetrics

	pubsubSubscription *pubsub.Subscription
}

func NewPubSubForwarder(ctx context.Context, writer BigQueryWriter, logger log.Logger, metrics *metrics.AnalyticsMetrics, gcpProjectID string, topicName string, subscriptionName string) (*PubSubForwarder, error) {
	pubsubClient, err := pubsub.NewClient(ctx, gcpProjectID)
	if err != nil {
		return nil, fmt.Errorf("could not create pubsub client: %v", err)
	}

	if gcpProjectID == "local" {
		if _, err := pubsubClient.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic: pubsubClient.Topic(topicName),
		}); err != nil && err.Error() != "rpc error: code = AlreadyExists desc = Subscription already exists" {
			return nil, fmt.Errorf("could not create local pubsub subscription '%s' for topic '%s' on project '%s': %v", subscriptionName, topicName, gcpProjectID, err)
		}
	}

	return &PubSubForwarder{
		Writer:             writer,
		Logger:             logger,
		Metrics:            metrics,
		pubsubSubscription: pubsubClient.Subscription(subscriptionName),
	}, nil
}

// Forward reads the analytics entry from pubsub and writes it to BigQuery
func (psf *PubSubForwarder) Forward(ctx context.Context) {
	err := psf.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		psf.Metrics.EntriesReceived.Add(1)
		if entries, ok := ReadStatsEntries(m.Data); ok {
			m.Ack()

			for i := range entries {
				entry := &entries[i]
				entry.Timestamp = uint64(m.PublishTime.Unix())
				psf.Writer.Write(context.Background(), entry)
			}
		} else {
			psf.Metrics.ErrorMetrics.ReadFailure.Add(1)
		}
	})

	if err != context.Canceled {
		level.Error(psf.Logger).Log("msg", "could not setup to receive pubsub messages", "err", err)
		os.Exit(1)
	}
}
