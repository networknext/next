package analytics

import (
	"context"
	"sync/atomic"
)

// NoOpPingStatsPublisher does not perform any analytics actions. Useful for when analytics is not configured or for testing.
type NoOpPingStatsPublisher struct {
	submitted uint64
}

// Publish() does nothing
func (noop *NoOpPingStatsPublisher) Publish(ctx context.Context, entries []PingStatsEntry) error {
	atomic.AddUint64(&noop.submitted, uint64(len(entries)))
	return nil
}

// NoOpRelayStatsPublisher does not perform any analytics actions. Useful for when analytics is not configured or for testing.
type NoOpRelayStatsPublisher struct {
	submitted uint64
}

// Publish() does nothing
func (noop *NoOpRelayStatsPublisher) Publish(ctx context.Context, entries []RelayStatsEntry) error {
	atomic.AddUint64(&noop.submitted, uint64(len(entries)))
	return nil
}

// NoOpPingStatsWriter does not perform any analytics actions. Useful for when analytics is not configured or for testing.
type NoOpPingStatsWriter struct {
	submitted uint64
}

// Write() does nothing
func (noop *NoOpPingStatsWriter) Write(ctx context.Context, entries []*PingStatsEntry) error {
	atomic.AddUint64(&noop.submitted, uint64(len(entries)))
	return nil
}

// Close() does nothing
func (noop *NoOpPingStatsWriter) Close() {}

// NoOpRelayStatsWriter does not perform any analytics actions. Useful for when analytics is not configured or for testing.
type NoOpRelayStatsWriter struct {
	submitted uint64
}

// Write() does nothing
func (noop *NoOpRelayStatsWriter) Write(ctx context.Context, entries []*RelayStatsEntry) error {
	atomic.AddUint64(&noop.submitted, uint64(len(entries)))
	return nil
}

// Close() does nothing
func (noop *NoOpRelayStatsWriter) Close() {}
