package pubsub

import (
	"testing"
	"time"
	
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"

	googlepubsub "cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

type TestEntry struct {
	Number uint64
}

func (entry *TestEntry) WriteEntry() ([]byte, error) {
	data := make([]byte, 8)
	index := 0
	
	encoding.WriteUint64(data, &index, entry.Number)

	return data[:index], nil
}

func (entry *TestEntry) ReadEntry(data []byte) bool {
	index := 0
	
}


func checkGooglePubsubEmulator(t *testing.T) {
	pubsubEmulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST")
	if pubsubEmulatorHost == "" {
		t.Skip("Pub/Sub emulator not set up, skipping pub/sub publisher tests")
	}
}

func TestNewPubSubPublisher(t *testing.T) {
	checkGooglePubsubEmulator(t)

	ctx, cancel := context.WithCancel()
	logger := log.NewNopLogger()

	t.Run("no publish settings", func(t *testing.T) {
		_, err := NewPubSubPublisher(ctx, cancel, metrics.EmptyGooglePublisherMetrics, logger, "", "", 0, 0, 0, nil)
		assert.EqualError(t, err, "NewPubSubPublisher(): nil google pubsub publish settings")
	})

	t.Run("success", func(t *testing.T) {
		publisher, err := NewPubSubPublisher(ctx, cancel, metrics.EmptyGooglePublisherMetrics, logger, "local", "test_topic", 1, 1, 1, &googlepubsub.DefaultPublishSettings)
		assert.NoError(t, err)
		assert.NotNil(t, publisher)
	})
}

func TestPubSubPublisherPublish(t *testing.T) {
	checkGooglePubsubEmulator(t)

	ctx, cancel := context.WithCancel()
	logger := log.NewNopLogger()
	gcpProjectID := "local"
	topicID := "test_topic"
	clientCount := 1
	clientBufferCountThreshold := 1
	clientMinBufferBytes := 1
	settings := &googlepubsub.DefaultPublishSettings

	t.Run("uninitialized clients", func(t *testing.T) {
		publisher := &PubSubPublisher{}
		err := publisher.Publish(ctx, &TestEntry{})
		assert.EqualError(t, err, "GooglePubSubPublisher Publish(): clients not initialized")
	})




}