package stats

import "context"

// NoOpTrafficStatsPublisher does not perform any traffic stats publishing actions. Useful for when traffic stats are not configured or for testing.
type NoOpTrafficStatsPublisher struct{}

// Publish does nothing
func (noop *NoOpTrafficStatsPublisher) Publish(ctx context.Context, relayID uint64, entry *RelayTrafficStats) error {
	return nil
}
