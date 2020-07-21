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

func TestNewGooglePubSubWriter(t *testing.T) {
	checkGooglePubsubEmulator(t)

	t.Run("no publish settings", func(t *testing.T) {
		_, err := analytics.NewGooglePubSubWriter(context.Background(), &metrics.EmptyAnalyticsMetrics, log.NewNopLogger(), "", "", 0, 0, nil)
		assert.EqualError(t, err, "nil google pubsub publish settings")
	})

	t.Run("success", func(t *testing.T) {
		_, err := analytics.NewGooglePubSubWriter(context.Background(), &metrics.EmptyAnalyticsMetrics, log.NewNopLogger(), "default", "analytics", 1, 0, &pubsub.DefaultPublishSettings)
		assert.NoError(t, err)
	})
}

func TestGooglePubSubWriter(t *testing.T) {
	checkGooglePubsubEmulator(t)
	ctx := context.Background()

}
