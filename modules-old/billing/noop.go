package billing

import (
    "context"
    "sync/atomic"
)

// NoOpBiller does not perform any billing actions. Useful for when billing is not configured or for testing.
type NoOpBiller struct {
    submitted uint64
}

// Bill2 does nothing
func (noop *NoOpBiller) Bill2(ctx context.Context, entry *BillingEntry2) error {
    atomic.AddUint64(&noop.submitted, 1)
    return nil
}

// FlushBuffer does nothing
func (noop *NoOpBiller) FlushBuffer(ctx context.Context) {}

// Close does nothing
func (noop *NoOpBiller) Close() {}

func (noop *NoOpBiller) NumSubmitted() uint64 {
    return atomic.LoadUint64(&noop.submitted)
}

func (noop *NoOpBiller) NumQueued() uint64 {
    return 0
}

func (noop *NoOpBiller) NumFlushed() uint64 {
    return atomic.LoadUint64(&noop.submitted)
}
