package billing_test

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"

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
		_, err := billing.NewGooglePubSubBiller(context.Background(), &metrics.EmptyBillingMetrics, "", "", 0, 0, 0, nil)
		assert.EqualError(t, err, "nil google pubsub publish settings")
	})

	t.Run("success", func(t *testing.T) {
		_, err := billing.NewGooglePubSubBiller(context.Background(), &metrics.EmptyBillingMetrics, "default", "billing", 1, 0, 0, &pubsub.DefaultPublishSettings)
		assert.NoError(t, err)
	})
}

func TestGooglePubSubBill2(t *testing.T) {
	checkGooglePubsubEmulator(t)
	ctx := context.Background()

	t.Run("uninitialized billing clients", func(t *testing.T) {
		biller := &billing.GooglePubSubBiller{}
		err := biller.Bill2(ctx, &billing.BillingEntry2{})
		assert.EqualError(t, err, "billing: clients not initialized")
	})

	t.Run("success", func(t *testing.T) {
		biller, err := billing.NewGooglePubSubBiller(context.Background(), &metrics.EmptyBillingMetrics, "default", "billing", 1, 0, 0, &pubsub.DefaultPublishSettings)
		assert.NoError(t, err)

		err = biller.Bill2(ctx, &billing.BillingEntry2{})
		assert.NoError(t, err)
	})
}

func TestLocalBill2(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		biller := billing.LocalBiller{
			Metrics: &metrics.EmptyBillingMetrics,
		}

		err := biller.Bill2(context.Background(), &billing.BillingEntry2{})
		assert.NoError(t, err)
	})
}

func TestNoOpBill2(t *testing.T) {
	biller := billing.NoOpBiller{}
	err := biller.Bill2(context.Background(), nil)
	assert.NoError(t, err)
}
