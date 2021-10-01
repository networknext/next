package analytics

import (
	"context"
	"fmt"
	"os"
	"sync"

	"cloud.google.com/go/pubsub"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"
)

type PingStatsPubSubForwarder struct {
	Writer  PingStatsWriter
	Metrics *metrics.AnalyticsMetrics

	pubsubSubscription *pubsub.Subscription
}

func NewPingStatsPubSubForwarder(ctx context.Context, writer PingStatsWriter, metrics *metrics.AnalyticsMetrics, gcpProjectID string, topicName string, subscriptionName string) (*PingStatsPubSubForwarder, error) {
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

	return &PingStatsPubSubForwarder{
		Writer:             writer,
		Metrics:            metrics,
		pubsubSubscription: pubsubClient.Subscription(subscriptionName),
	}, nil
}

// Forward() reads the ping stats entries from Google Pub/Sub and writes them to BigQuery
func (psf *PingStatsPubSubForwarder) Forward(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	err := psf.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {

		if entries, ok := ReadPingStatsEntries(m.Data); ok {

			psf.Metrics.EntriesReceived.Add(float64(len(entries)))

			for i := range entries {
				entries[i].Timestamp = uint64(m.PublishTime.Unix())
			}

			if err := psf.Writer.Write(ctx, entries); err != nil {
				core.Error("failed to submit ping stats batch entry: %v", err)

				// Nack if we fail to submit the batch of ping stats entries at once
				m.Nack()
				return
			}

			// Successfully submitted all ping stats to the channel
			m.Ack()
		} else {
			core.Error("failed to read ping stats batch entry")
			psf.Metrics.ErrorMetrics.ReadFailure.Add(1)

			m.Nack()
			return
		}
	})

	if err != nil && err != context.Canceled {
		// If the Receive function returns for any reason besides shutdown, we want to immediately exit and restart the service
		core.Error("stopped receive loop: %v", err)
		os.Exit(1)
	}

	psf.Writer.Close()
	core.Debug("receive canceled, closed ping stats entries channel")
}

type RelayStatsPubSubForwarder struct {
	Writer  RelayStatsWriter
	Metrics *metrics.AnalyticsMetrics

	pubsubSubscription *pubsub.Subscription
}

func NewRelayStatsPubSubForwarder(ctx context.Context, writer RelayStatsWriter, metrics *metrics.AnalyticsMetrics, gcpProjectID string, topicName string, subscriptionName string) (*RelayStatsPubSubForwarder, error) {
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

	return &RelayStatsPubSubForwarder{
		Writer:             writer,
		Metrics:            metrics,
		pubsubSubscription: pubsubClient.Subscription(subscriptionName),
	}, nil
}

// Forward() reads the relay stats entries from Google Pub/Sub and writes them to BigQuery
func (psf *RelayStatsPubSubForwarder) Forward(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	err := psf.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {

		if entries, ok := ReadRelayStatsEntries(m.Data); ok {

			psf.Metrics.EntriesReceived.Add(float64(len(entries)))

			// Assign a timestamp for the entries
			for i := range entries {
				entries[i].Timestamp = uint64(m.PublishTime.Unix())
			}

			if err := psf.Writer.Write(ctx, entries); err != nil {
				core.Error("failed to submit relay stats batch entry: %v", err)

				// Nack if we fail to submit the batch of relay stats entries at once
				m.Nack()
				return
			}

			// Successfully submitted all relay stats to the channel
			m.Ack()
		} else {
			core.Error("failed to read relay stats batch entry")
			psf.Metrics.ErrorMetrics.ReadFailure.Add(1)

			m.Nack()
			return
		}
	})

	if err != nil && err != context.Canceled {
		// If the Receive function returns for any reason besides shutdown, we want to immediately exit and restart the service
		core.Error("stopped receive loop: %v", err)
		os.Exit(1)
	}

	psf.Writer.Close()
	core.Debug("receive canceled, closed relay stats entries channel")
}
