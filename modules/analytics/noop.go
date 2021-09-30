package analytics

import (
	"context"
	"sync/atomic"
)

// NoOpPingStatsWriter does not perform any analytics actions. Useful for when analytics is not configured or for testing.
type NoOpPingStatsWriter struct {
	submitted uint64
}

// Write() does nothing
func (noop *NoOpPingStatsWriter) Write(ctx context.Context, entries []*PingStatsEntry) error {
	atomic.AddUint64(&noop.submitted, uint64(len(entries)))
	return nil
}

// NoOpRelayStatsWriter does not perform any analytics actions. Useful for when analytics is not configured or for testing.
type NoOpRelayStatsWriter struct {
	submitted uint64
}

// Write() does nothing
func (noop *NoOpRelayStatsWriter) Write(ctx context.Context, entries []*RelayStatsEntry) error {
	atomic.AddUint64(&noop.submitted, uint64(len(entries)))
	return nil
}
