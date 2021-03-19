package pubsub

import (
	"context"
	"fmt"
	"sync/atomic"
)

// NoOpPubSubPublisher does not perform any publishing actions. Useful for when the Google Pub/Sub Subscriber is not configured or for testing.
type NoOpPubSubPublisher struct {
	submitted uint64
}

// Publish does nothing besides checking if the given entry can be marshaled and incrementing the appropriate metrics.
func (publisher *NoOpPubSubPublisher) Publish(ctx context.Context, entry *Entry) error {
	// Ensure the entry can be marshaled
	_, err := entry.WriteEntry()
	if err != nil {
		return fmt.Errorf("NoOpPubSubPublisher Publish(): %v", err)
	}

	atomic.AddUint64(&publisher.submitted, 1)
	return nil
}

func (publisher *NoOpPubSubPublisher) NumSubmitted() uint64 {
	return atomic.LoadUint64(&publisher.submitted)
}

func (publisher *NoOpPubSubPublisher) NumQueued() uint64 {
	return 0
}

func (publisher *NoOpPubSubPublisher) NumFlushed() uint64 {
	return atomic.LoadUint64(&publisher.submitted)
}
