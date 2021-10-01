package analytics_test

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/metrics"

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

	t.Parallel()

	t.Run("ping stats publisher", func(t *testing.T) {
		_, err := analytics.NewGooglePubSubPingStatsPublisher(context.Background(), &metrics.EmptyAnalyticsMetrics, "default", "analytics", pubsub.DefaultPublishSettings)
		assert.NoError(t, err)
	})

	t.Run("relay stats publisher", func(t *testing.T) {
		_, err := analytics.NewGooglePubSubRelayStatsPublisher(context.Background(), &metrics.EmptyAnalyticsMetrics, "default", "analytics", pubsub.DefaultPublishSettings)
		assert.NoError(t, err)
	})

}

func TestGooglePubSubPublisher(t *testing.T) {
	checkGooglePubsubEmulator(t)

	t.Parallel()

	ctx := context.Background()

	t.Run("ping stats uninitialized writing client", func(t *testing.T) {
		publisher := &analytics.GooglePubSubPingStatsPublisher{}
		err := publisher.Publish(ctx, []analytics.PingStatsEntry{})
		assert.EqualError(t, err, "analytics: ping stats pub/sub client not initialized")
	})

	t.Run("ping stats publisher success", func(t *testing.T) {
		analyticsMetrics, err := metrics.NewAnalyticsServiceMetrics(ctx, &metrics.LocalHandler{})
		assert.NoError(t, err)

		pingStatsMetrics := analyticsMetrics.PingStatsMetrics

		publisher, err := analytics.NewGooglePubSubPingStatsPublisher(ctx, &pingStatsMetrics, "default", "analytics", pubsub.DefaultPublishSettings)
		assert.NoError(t, err)

		entry := analytics.PingStatsEntry{}
		err = publisher.Publish(ctx, []analytics.PingStatsEntry{entry})
		assert.NoError(t, err)

		assert.Equal(t, float64(1), pingStatsMetrics.EntriesSubmitted.Value())
	})

	t.Run("relay stats uninitialized writing client", func(t *testing.T) {
		publisher := &analytics.GooglePubSubRelayStatsPublisher{}
		err := publisher.Publish(ctx, []analytics.RelayStatsEntry{})
		assert.EqualError(t, err, "analytics: relay stats pub/sub client not initialized")
	})

	t.Run("relay stats publisher success", func(t *testing.T) {
		analyticsMetrics, err := metrics.NewAnalyticsServiceMetrics(ctx, &metrics.LocalHandler{})
		assert.NoError(t, err)

		relayStatsMetrics := analyticsMetrics.RelayStatsMetrics

		publisher, err := analytics.NewGooglePubSubRelayStatsPublisher(ctx, &relayStatsMetrics, "default", "analytics", pubsub.DefaultPublishSettings)
		assert.NoError(t, err)

		entry := analytics.RelayStatsEntry{}
		err = publisher.Publish(ctx, []analytics.RelayStatsEntry{entry})
		assert.NoError(t, err)

		assert.Equal(t, float64(1), relayStatsMetrics.EntriesSubmitted.Value())
	})
}

func TestLocalBigQueryWriter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("ping stats success", func(t *testing.T) {

		analyticsMetrics, err := metrics.NewAnalyticsServiceMetrics(ctx, &metrics.LocalHandler{})
		assert.NoError(t, err)

		writer := analytics.LocalPingStatsWriter{
			Metrics: &analyticsMetrics.PingStatsMetrics,
		}

		entry := &analytics.PingStatsEntry{}

		err = writer.Write(context.Background(), []*analytics.PingStatsEntry{entry})
		assert.NoError(t, err)

		assert.Equal(t, float64(1), writer.Metrics.EntriesSubmitted.Value())
		assert.Equal(t, float64(1), writer.Metrics.EntriesFlushed.Value())
	})

	t.Run("relay stats success", func(t *testing.T) {

		analyticsMetrics, err := metrics.NewAnalyticsServiceMetrics(ctx, &metrics.LocalHandler{})
		assert.NoError(t, err)

		writer := analytics.LocalRelayStatsWriter{
			Metrics: &analyticsMetrics.RelayStatsMetrics,
		}

		entry := &analytics.RelayStatsEntry{}

		err = writer.Write(context.Background(), []*analytics.RelayStatsEntry{entry})
		assert.NoError(t, err)

		assert.Equal(t, float64(1), writer.Metrics.EntriesSubmitted.Value())
		assert.Equal(t, float64(1), writer.Metrics.EntriesFlushed.Value())
	})
}

func TestNoOp(t *testing.T) {
	t.Parallel()

	t.Run("ping stats publisher", func(t *testing.T) {
		publisher := analytics.NoOpPingStatsPublisher{}
		err := publisher.Publish(context.Background(), []analytics.PingStatsEntry{})
		assert.NoError(t, err)
	})

	t.Run("ping stats writer", func(t *testing.T) {
		writer := analytics.NoOpPingStatsWriter{}
		err := writer.Write(context.Background(), []*analytics.PingStatsEntry{})
		assert.NoError(t, err)
	})

	t.Run("relay stats publisher", func(t *testing.T) {
		publisher := analytics.NoOpRelayStatsPublisher{}
		err := publisher.Publish(context.Background(), []analytics.RelayStatsEntry{})
		assert.NoError(t, err)
	})

	t.Run("relay stats writer", func(t *testing.T) {
		writer := analytics.NoOpRelayStatsWriter{}
		err := writer.Write(context.Background(), []*analytics.RelayStatsEntry{})
		assert.NoError(t, err)
	})
}
