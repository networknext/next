package beacon

import (
	"context"
	"sync/atomic"

	"github.com/networknext/backend/modules/transport"
)

// NoOpBeaconer does not perform any beacon actions. Useful for when the beacon is not configured or for testing.
type NoOpBeaconer struct {
	submitted uint64
}

// Submit does nothing
func (noop *NoOpBeaconer) Submit(ctx context.Context, entry *transport.NextBeaconPacket) error {
	atomic.AddUint64(&noop.submitted, 1)
	return nil
}

// Close does nothing
func (noop *NoOpBeaconer) Close() {}

func (noop *NoOpBeaconer) NumSubmitted() uint64 {
	return atomic.LoadUint64(&noop.submitted)
}

func (noop *NoOpBeaconer) NumQueued() uint64 {
	return 0
}

func (noop *NoOpBeaconer) NumFlushed() uint64 {
	return atomic.LoadUint64(&noop.submitted)
}
