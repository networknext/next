package billing

import (
	"context"
	"sync/atomic"
)

// NoOpBiller does not perform any billing actions. Useful for when billing is not configured or for testing.
type NoOpBiller struct {
	submitted uint64
}

// Bill does nothing
func (noop *NoOpBiller) Bill(ctx context.Context, sessionID uint64, entry *Entry) error {
	atomic.AddUint64(&noop.submitted, 1)
	return nil
}

func (noop *NoOpBiller) NumSubmitted() uint64 {
	return noop.submitted
}

func (noop *NoOpBiller) NumQueued() uint64 {
	return 0
}

func (noop *NoOpBiller) NumFlushed() uint64 {
	return noop.submitted
}
