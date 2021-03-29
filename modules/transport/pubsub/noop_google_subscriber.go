package pubsub

import (
	"context"
	"sync/atomic"
)

// NoOpPubSubPublisher does not perform any receiving actions.
type NoOpPubSubSubscriber struct {
	received uint64
}

// ReceiveAndSubmit() increments a counter.
func (subscriber *NoOpPubSubSubscriber) ReceiveAndSubmit(ctx context.Context) error {
	atomic.AddUint64(&subscriber.received, 1)
	return nil
}

// Close() does nothing.
func (subscriber *NoOpPubSubSubscriber) Close() {}

// WriteLoop() does nothing.
func (subscriber *NoOpPubSubSubscriber) WriteLoop(ctx context.Context) error {
	return nil
}

func (subscriber *NoOpPubSubSubscriber) NumReceived() uint64 {
	return atomic.LoadUint64(&subscriber.received)
}

func (subscriber *NoOpPubSubSubscriber) NumQueuedToWrite() uint64 {
	return 0
}

func (subscriber *NoOpPubSubSubscriber) NumSubmitted() uint64 {
	return atomic.LoadUint64(&subscriber.received)
}
