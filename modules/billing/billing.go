package billing

import (
	"context"
	"fmt"
)

const BillingSliceSeconds = 10

type ErrEntries2BufferFull struct{}

func (e *ErrEntries2BufferFull) Error() string {
	return fmt.Sprintf("entries2 buffer full")
}

type ErrSummaryEntries2BufferFull struct{}

func (e *ErrSummaryEntries2BufferFull) Error() string {
	return fmt.Sprintf("entries2 buffer full")
}

// Biller is a billing interface that handles sending billing entries to remote services
type Biller interface {
	Bill(ctx context.Context, billingEntry *BillingEntry) error
	Bill2(ctx context.Context, billingEntry *BillingEntry2) error
	FlushBuffer(ctx context.Context)
	Close()
}
