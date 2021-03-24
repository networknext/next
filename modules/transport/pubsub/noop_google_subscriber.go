package pubsub

import (
	"context"
	"sync/atomic"
)

type NoOpPubSubSubscriber struct {
	received uint64
}

func (subscriber *NoOpPubSubSubscriber) Receive(ctx context.Context) error {
	atomic.AddUint64(&subscriber.received, 1)
	return nil
}

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
