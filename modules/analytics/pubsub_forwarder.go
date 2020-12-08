package analytics

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
)

type PingStatsPubSubForwarder struct {
	Writer  PingStatsWriter
	Logger  log.Logger
	Metrics *metrics.AnalyticsMetrics

	pubsubSubscription *pubsub.Subscription
}

func NewPingStatsPubSubForwarder(ctx context.Context, writer PingStatsWriter, logger log.Logger, metrics *metrics.AnalyticsMetrics, gcpProjectID string, topicName string, subscriptionName string) (*PingStatsPubSubForwarder, error) {
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
		Logger:             logger,
		Metrics:            metrics,
		pubsubSubscription: pubsubClient.Subscription(subscriptionName),
	}, nil
}

// Forward reads the analytics entry from pubsub and writes it to BigQuery
func (psf *PingStatsPubSubForwarder) Forward(ctx context.Context) {
	err := psf.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		psf.Metrics.EntriesReceived.Add(1)
		if entries, ok := ReadPingStatsEntries(m.Data); ok {
			m.Ack()

			for i := range entries {
				entries[i].Timestamp = uint64(m.PublishTime.Unix())
			}
			psf.Writer.Write(context.Background(), entries)
		} else {
			psf.Metrics.ErrorMetrics.ReadFailure.Add(1)
		}
	})

	if err != context.Canceled {
		level.Error(psf.Logger).Log("msg", "could not setup to receive pubsub messages", "err", err)
		os.Exit(1)
	}
}

type RelayStatsPubSubForwarder struct {
	Writer  RelayStatsWriter
	Logger  log.Logger
	Metrics *metrics.AnalyticsMetrics

	pubsubSubscription *pubsub.Subscription
}

func NewRelayStatsPubSubForwarder(ctx context.Context, writer RelayStatsWriter, logger log.Logger, metrics *metrics.AnalyticsMetrics, gcpProjectID string, topicName string, subscriptionName string) (*RelayStatsPubSubForwarder, error) {
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
		Logger:             logger,
		Metrics:            metrics,
		pubsubSubscription: pubsubClient.Subscription(subscriptionName),
	}, nil
}

// Forward reads the analytics entry from pubsub and writes it to BigQuery
func (psf *RelayStatsPubSubForwarder) Forward(ctx context.Context) {
	err := psf.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		psf.Metrics.EntriesReceived.Add(1)
		if entries, ok := ReadRelayStatsEntries(m.Data); ok {
			m.Ack()

			for i := range entries {
				entries[i].Timestamp = uint64(m.PublishTime.Unix())
			}
			psf.Writer.Write(context.Background(), entries)
		} else {
			psf.Metrics.ErrorMetrics.ReadFailure.Add(1)
		}
	})

	if err != context.Canceled {
		level.Error(psf.Logger).Log("msg", "could not setup to receive pubsub messages", "err", err)
		os.Exit(1)
	}
}

type RouteMatrixStatsPubSubForwarder struct {
	Writer  RouteMatrixStatsWriter
	Logger  log.Logger
	Metrics *metrics.AnalyticsMetrics

	pubsubSubscription *pubsub.Subscription
}

func NewRouteMatrixStatsPubSubForwarder(ctx context.Context, writer RouteMatrixStatsWriter, logger log.Logger, metrics *metrics.AnalyticsMetrics,
	gcpProjectID string, topicName string, subscriptionName string) (*RouteMatrixStatsPubSubForwarder, error) {
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

	return &RouteMatrixStatsPubSubForwarder{
		Writer:             writer,
		Logger:             logger,
		Metrics:            metrics,
		pubsubSubscription: pubsubClient.Subscription(subscriptionName),
	}, nil
}

// Forward reads the analytics entry from pubsub and writes it to BigQuery
func (psf *RouteMatrixStatsPubSubForwarder) Forward(ctx context.Context) {
	err := psf.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		psf.Metrics.EntriesReceived.Add(1)
		if entry, ok := ReadRouteMatrixStatsEntry(m.Data); ok {
			m.Ack()

			psf.Writer.Write(context.Background(), entry)
		} else {
			psf.Metrics.ErrorMetrics.ReadFailure.Add(1)
		}
	})

	if err != context.Canceled {
		level.Error(psf.Logger).Log("msg", "could not setup to receive pubsub messages", "err", err)
		os.Exit(1)
	}
}
