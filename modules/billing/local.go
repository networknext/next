package billing

import (
	"context"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"
)

type LocalBiller struct {
	Metrics *metrics.BillingMetrics
}

func (local *LocalBiller) Bill2(ctx context.Context, entry *BillingEntry2) error {
	if entry.Summary {
		local.Metrics.SummaryEntries2Submitted.Add(1)
		core.Debug("submitted billing entry 2 summary")
	} else {
		local.Metrics.Entries2Submitted.Add(1)
		core.Debug("submitted billing entry 2")
	}

	// core.Debug("%+v", entry)

	if entry.Summary {
		local.Metrics.SummaryEntries2Flushed.Add(1)
	} else {
		local.Metrics.Entries2Flushed.Add(1)
	}

	return nil
}

func (local *LocalBiller) FlushBuffer(ctx context.Context) {}

func (local *LocalBiller) Close() {}
