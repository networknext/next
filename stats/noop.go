package stats

import "context"

// NoOpPublisher ...
type NoOpTrafficStatsPublisher struct{}

func (noop *NoOpTrafficStatsPublisher) Publish(ctx context.Context, relayID uint64, entry *RelayTrafficStats) error {
	return nil
}
