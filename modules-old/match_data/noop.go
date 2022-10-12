package match_data

import (
	"context"
	"sync/atomic"
)

// NoOpMatcher does not perform any matching actions. Useful for when match data is not configured or for testing.
type NoOpMatcher struct {
	submitted uint64
}

// Match does nothing
func (noop *NoOpMatcher) Match(ctx context.Context, entry *MatchDataEntry) error {
	atomic.AddUint64(&noop.submitted, 1)
	return nil
}

// FlushBuffer does nothing
func (noop *NoOpMatcher) FlushBuffer(ctx context.Context) {}

// Close does nothing
func (noop *NoOpMatcher) Close() {}

func (noop *NoOpMatcher) NumSubmitted() uint64 {
	return atomic.LoadUint64(&noop.submitted)
}

func (noop *NoOpMatcher) NumQueued() uint64 {
	return 0
}

func (noop *NoOpMatcher) NumFlushed() uint64 {
	return atomic.LoadUint64(&noop.submitted)
}
