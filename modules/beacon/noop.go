package beacon

import (
	"context"
	"sync/atomic"

	"github.com/networknext/backend/modules/transport"
)

// NoOpBeacon does not perform any beacon actions. Useful for when the beacon is not configured or for testing.
type NoOpBeacon struct {
	submitted uint64
}

// Bill does nothing
func (noop *NoOpBeacon) Submit(ctx context.Context, entry *transport.NextBeaconPacket) error {
	atomic.AddUint64(&noop.submitted, 1)
	return nil
}

func (noop *NoOpBeacon) NumSubmitted() uint64 {
	return atomic.LoadUint64(&noop.submitted)
}

func (noop *NoOpBeacon) NumQueued() uint64 {
	return 0
}

func (noop *NoOpBeacon) NumFlushed() uint64 {
	return atomic.LoadUint64(&noop.submitted)
}
