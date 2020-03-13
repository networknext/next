package stats

import "context"

type Publisher interface {
	Publish(ctx context.Context, relayID uint64, entry *RelayTrafficStats) error
}
