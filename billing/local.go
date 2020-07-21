package billing

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/metrics"
)

type LocalBiller struct {
	Logger  log.Logger
	Metrics *metrics.BillingMetrics
}

func (local *LocalBiller) Bill(ctx context.Context, entry *BillingEntry) error {
	fmt.Println(local)
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
