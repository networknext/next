package billing_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/billing"
	"github.com/stretchr/testify/assert"
)

func checkPubsubEmulator(t *testing.T) {
	pubsubEmulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST")
	if pubsubEmulatorHost == "" {
		t.Skip("Pub/Sub emulator not set up, skipping billing pub/sub tests")
	}
}

func TestNewPubSubBiller(t *testing.T) {
	checkPubsubEmulator(t)

	// Test base case
	_, err := billing.NewBiller(context.Background(), log.NewNopLogger(), "", "", nil)
	assert.NoError(t, err)

	// Test success case
	projectID := "default"

	descriptor := &billing.Descriptor{
		ClientCount:         1,
		DelayThreshold:      time.Millisecond,
		CountThreshold:      1001, // Set this higher than the allowed pubsub max
		ByteThreshold:       1e8,  // Set this higher than the allowed pubsub max
		NumGoroutines:       1,
		Timeout:             time.Minute,
		ResultChannelBuffer: 10000 * 60 * 10,
	}
	_, err = billing.NewBiller(context.Background(), log.NewNopLogger(), projectID, "billing", descriptor)
	assert.NoError(t, err)
	assert.Equal(t, billing.Descriptor{
		ClientCount:         1,
		DelayThreshold:      time.Millisecond,
		CountThreshold:      1000,
		ByteThreshold:       1e7,
		NumGoroutines:       1,
		Timeout:             time.Minute,
		ResultChannelBuffer: 10000 * 60 * 10,
	}, *descriptor)
}

func TestPubSubBill(t *testing.T) {
	checkPubsubEmulator(t)
	ctx := context.Background()

	// Call Bill() with bad data
	var biller billing.Biller
	biller = &billing.GooglePubSubBiller{}
	err := biller.Bill(ctx, 0, nil)
	assert.Error(t, err)

	// Call Bill() with an uninitialized biller
	err = biller.Bill(ctx, 0, &billing.Entry{})
	assert.EqualError(t, err, "billing: clients not initialized")

	// Success case
	descriptor := &billing.Descriptor{
		ClientCount:         1,
		DelayThreshold:      time.Millisecond,
		CountThreshold:      1001, // Set this higher than the allowed pubsub max
		ByteThreshold:       1e8,  // Set this higher than the allowed pubsub max
		NumGoroutines:       1,
		Timeout:             time.Minute,
		ResultChannelBuffer: 10000 * 60 * 10,
	}
	biller, err = billing.NewBiller(context.Background(), log.NewNopLogger(), "default", "billing", descriptor)
	assert.NoError(t, err)
	assert.Equal(t, billing.Descriptor{
		ClientCount:         1,
		DelayThreshold:      time.Millisecond,
		CountThreshold:      1000,
		ByteThreshold:       1e7,
		NumGoroutines:       1,
		Timeout:             time.Minute,
		ResultChannelBuffer: 10000 * 60 * 10,
	}, *descriptor)

	biller.Bill(ctx, 0, &billing.Entry{})
}

func TestNoOpBill(t *testing.T) {
	biller := billing.NoOpBiller{}
	biller.Bill(context.Background(), 0, nil)
}
