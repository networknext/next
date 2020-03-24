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

func TestNewPubSubBiller(t *testing.T) {
	// Test base case
	_, err := billing.NewBiller(context.Background(), log.NewNopLogger(), "", "", nil)
	assert.NoError(t, err)

	// Test success case
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if projectID == "" {
		t.Skip() // Ignore this test if billing isn't configured
	}

	assert.NotEmpty(t, projectID)

	topicID := os.Getenv("GOOGLE_PUBSUB_TOPIC_BILLING")
	assert.NotEmpty(t, topicID)

	descriptor := &billing.Descriptor{
		ClientCount:         1,
		DelayThreshold:      time.Millisecond,
		CountThreshold:      1001, // Set this higher than the allowed pubsub max
		ByteThreshold:       1e8,  // Set this higher than the allowed pubsub max
		NumGoroutines:       1,
		Timeout:             time.Minute,
		ResultChannelBuffer: 10000 * 60 * 10,
	}
	_, err = billing.NewBiller(context.Background(), log.NewNopLogger(), projectID, topicID, descriptor)
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
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	if projectID == "" {
		t.Skip() // Ignore this test if billing isn't configured
	}
	assert.NotEmpty(t, projectID)

	topicID := os.Getenv("GOOGLE_PUBSUB_TOPIC_BILLING")
	assert.NotEmpty(t, topicID)

	descriptor := &billing.Descriptor{
		ClientCount:         1,
		DelayThreshold:      time.Millisecond,
		CountThreshold:      1001, // Set this higher than the allowed pubsub max
		ByteThreshold:       1e8,  // Set this higher than the allowed pubsub max
		NumGoroutines:       1,
		Timeout:             time.Minute,
		ResultChannelBuffer: 10000 * 60 * 10,
	}
	biller, err = billing.NewBiller(context.Background(), log.NewNopLogger(), projectID, topicID, descriptor)
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
