package billing_test

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/metrics"
	"github.com/stretchr/testify/assert"
)

func checkGooglePubsubEmulator(t *testing.T) {
	pubsubEmulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST")
	if pubsubEmulatorHost == "" {
		t.Skip("Pub/Sub emulator not set up, skipping billing pub/sub tests")
	}
}

func TestNewGooglePubSubBiller(t *testing.T) {
	checkGooglePubsubEmulator(t)

	t.Run("no publish settings", func(t *testing.T) {
		_, err := billing.NewGooglePubSubBiller(context.Background(), &metrics.EmptyBillingMetrics, log.NewNopLogger(), "", "", 0, 0, 0, nil)
		assert.EqualError(t, err, "nil google pubsub publish settings")
	})

	t.Run("success", func(t *testing.T) {
		_, err := billing.NewGooglePubSubBiller(context.Background(), &metrics.EmptyBillingMetrics, log.NewNopLogger(), "default", "billing", 1, 0, 0, &pubsub.DefaultPublishSettings)
		assert.NoError(t, err)
	})
}

func TestGooglePubSubBill(t *testing.T) {
	checkGooglePubsubEmulator(t)
	ctx := context.Background()

	t.Run("uninitialized billing clients", func(t *testing.T) {
		biller := &billing.GooglePubSubBiller{}
		err := biller.Bill(ctx, &billing.BillingEntry{})
		assert.EqualError(t, err, "billing: clients not initialized")
	})

	t.Run("success", func(t *testing.T) {
		biller, err := billing.NewGooglePubSubBiller(context.Background(), &metrics.EmptyBillingMetrics, log.NewNopLogger(), "default", "billing", 1, 0, 0, &pubsub.DefaultPublishSettings)
		assert.NoError(t, err)

		biller.Bill(ctx, &billing.BillingEntry{})
	})
}

func TestLocalBill(t *testing.T) {
	t.Run("no logger", func(t *testing.T) {
		biller := billing.LocalBiller{
			Metrics: &metrics.EmptyBillingMetrics,
		}
		err := biller.Bill(context.Background(), &billing.BillingEntry{})
		assert.EqualError(t, err, "no logger for local biller, can't display entry")
	})

	t.Run("success", func(t *testing.T) {
		biller := billing.LocalBiller{
			Logger:  log.NewNopLogger(),
			Metrics: &metrics.EmptyBillingMetrics,
		}

		err := biller.Bill(context.Background(), &billing.BillingEntry{})
		assert.NoError(t, err)
	})
}

func TestNoOpBill(t *testing.T) {
	biller := billing.NoOpBiller{}
	biller.Bill(context.Background(), nil)
}
