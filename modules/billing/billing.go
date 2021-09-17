package billing

import (
	"context"
)

const BillingSliceSeconds = 10

// Biller is a billing interface that handles sending billing entries to remote services
type Biller interface {
	Bill(ctx context.Context, billingEntry *BillingEntry) error
	Bill2(ctx context.Context, billingEntry *BillingEntry2) error
	FlushBuffer(ctx context.Context)
	Close()
}
