package beacon_test

/*
import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/beacon"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func checkGooglePubsubEmulator(t *testing.T) {
	pubsubEmulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST")
	if pubsubEmulatorHost == "" {
		t.Skip("Pub/Sub emulator not set up, skipping beacon pub/sub tests")
	}
}

func TestNewGooglePubSubBeaconer(t *testing.T) {
	checkGooglePubsubEmulator(t)

	t.Run("no publish settings", func(t *testing.T) {
		_, err := beacon.NewGooglePubSubBeaconer(context.Background(), &metrics.EmptyBeaconMetrics, log.NewNopLogger(), "", "", 0, 0, 0, nil)
		assert.EqualError(t, err, "nil google pubsub publish settings")
	})

	t.Run("success", func(t *testing.T) {
		_, err := beacon.NewGooglePubSubBeaconer(context.Background(), &metrics.EmptyBeaconMetrics, log.NewNopLogger(), "default", "beacon", 1, 0, 0, &pubsub.DefaultPublishSettings)
		assert.NoError(t, err)
	})
}

func TestGooglePubSubSubmit(t *testing.T) {
	checkGooglePubsubEmulator(t)
	ctx := context.Background()

	t.Run("uninitialized beacon clients", func(t *testing.T) {
		beaconer := &beacon.GooglePubSubBeaconer{}
		err := beaconer.Submit(ctx, &transport.NextBeaconPacket{})
		assert.EqualError(t, err, "beacon: clients not initialized")
	})

	t.Run("success", func(t *testing.T) {
		beaconer, err := beacon.NewGooglePubSubBeaconer(context.Background(), &metrics.EmptyBeaconMetrics, log.NewNopLogger(), "default", "beacon", 1, 0, 0, &pubsub.DefaultPublishSettings)
		assert.NoError(t, err)

		err = beaconer.Submit(ctx, &transport.NextBeaconPacket{})
		assert.NoError(t, err)
	})
}

func TestLocalSubmit(t *testing.T) {
	t.Run("no logger", func(t *testing.T) {
		beaconer := beacon.LocalBeaconer{
			Metrics: &metrics.EmptyBeaconMetrics,
		}
		err := beaconer.Submit(context.Background(), &transport.NextBeaconPacket{})
		assert.EqualError(t, err, "no logger for local beaconer, can't display entry")
	})

	t.Run("success", func(t *testing.T) {
		beaconer := beacon.LocalBeaconer{
			Logger:  log.NewNopLogger(),
			Metrics: &metrics.EmptyBeaconMetrics,
		}

		err := beaconer.Submit(context.Background(), &transport.NextBeaconPacket{})
		assert.NoError(t, err)
	})
}

func TestNoOpSubmit(t *testing.T) {
	beaconer := beacon.NoOpBeaconer{}

	err := beaconer.Submit(context.Background(), nil)
	assert.NoError(t, err)
}
*/
