package billing

import (
	"context"
)

const BillingSliceSeconds = 10

// Biller is a billing interface that handles sending billing entries to remote services
type Biller interface {
	Bill(ctx context.Context, billingEntry *BillingEntry) error
}
