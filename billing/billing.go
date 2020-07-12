package billing

import (
	"context"
)

// Biller is a billing interface that handles sending billing entries to remote services
type Biller interface {
	Bill(ctx context.Context, billingEntry *BillingEntry) error
	NumSubmitted() uint64
	NumQueued() uint64
	NumFlushed() uint64
}
