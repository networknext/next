package analytics_test

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/analytics"
	"github.com/networknext/backend/metrics"
	"github.com/stretchr/testify/assert"
)

func checkGooglePubsubEmulator(t *testing.T) {
	pubsubEmulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST")
	if pubsubEmulatorHost == "" {
		t.Skip("Pub/Sub emulator not set up, skipping analytics pub/sub tests")
	}
}

func TestNewGooglePubSubPublisher(t *testing.T) {
	checkGooglePubsubEmulator(t)
	_, err := analytics.NewGooglePubSubPingStatsPublisher(context.Background(), &metrics.EmptyAnalyticsMetrics, log.NewNopLogger(), "default", "analytics", pubsub.DefaultPublishSettings)
	assert.NoError(t, err)
}

func TestGooglePubSubPublisher(t *testing.T) {
	checkGooglePubsubEmulator(t)
	ctx := context.Background()

	t.Run("uninitialized writing client", func(t *testing.T) {
		publisher := &analytics.GooglePubSubPingStatsPublisher{}
		err := publisher.Publish(ctx, []analytics.PingStatsEntry{})
		assert.EqualError(t, err, "analytics: client not initialized")
	})

	t.Run("success", func(t *testing.T) {
		publisher, err := analytics.NewGooglePubSubPingStatsPublisher(ctx, &metrics.EmptyAnalyticsMetrics, log.NewNopLogger(), "default", "analytics", pubsub.DefaultPublishSettings)
		assert.NoError(t, err)
		err = publisher.Publish(ctx, []analytics.PingStatsEntry{})
		assert.NoError(t, err)
	})
}

func TestLocalBigQueryWriter(t *testing.T) {
	t.Run("no logger", func(t *testing.T) {
		writer := analytics.LocalPingStatsWriter{}
		err := writer.Write(context.Background(), &analytics.PingStatsEntry{})
		assert.EqualError(t, err, "no logger for local big query writer, can't display entry")
	})

	t.Run("success", func(t *testing.T) {
		writer := analytics.LocalPingStatsWriter{
			Logger: log.NewNopLogger(),
		}

		err := writer.Write(context.Background(), &analytics.PingStatsEntry{})
		assert.NoError(t, err)
	})
}

func TestNoOp(t *testing.T) {
	t.Run("pubsub", func(t *testing.T) {
		publisher := analytics.NoOpPingStatsPublisher{}
		publisher.Publish(context.Background(), []analytics.PingStatsEntry{})
	})

	t.Run("bigquery", func(t *testing.T) {
		writer := analytics.NoOpPingStatsWriter{}
		writer.Write(context.Background(), &analytics.PingStatsEntry{})
	})
}
