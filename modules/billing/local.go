package billing

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
)

type LocalBiller struct {
	Logger  log.Logger
	Metrics *metrics.BillingMetrics
}

func (local *LocalBiller) Bill(ctx context.Context, entry *BillingEntry) error {
	local.Metrics.EntriesSubmitted.Add(1)

	if local.Logger == nil {
		return errors.New("no logger for local biller, can't display entry")
	}

	level.Info(local.Logger).Log("msg", "submitted billing entry")

	output := fmt.Sprintf("%#v", entry)
	level.Debug(local.Logger).Log("entry", output)

	local.Metrics.EntriesFlushed.Add(1)

	return nil
}

func (local *LocalBiller) Bill2(ctx context.Context, entry *BillingEntry2) error {
	if entry.Summary {
		local.Metrics.SummaryEntries2Submitted.Add(1)
	} else {
		local.Metrics.Entries2Submitted.Add(1)
	}

	if local.Logger == nil {
		return errors.New("no logger for local biller, can't display entry")
	}

	level.Info(local.Logger).Log("msg", "submitted billing entry")

	output := fmt.Sprintf("%#v", entry)
	level.Debug(local.Logger).Log("entry", output)

	if entry.Summary {
		local.Metrics.SummaryEntries2Flushed.Add(1)
	} else {
		local.Metrics.Entries2Flushed.Add(1)
	}

	return nil
}

func (local *LocalBiller) FlushBuffer(ctx context.Context) {}

func (local *LocalBiller) Close() {}
