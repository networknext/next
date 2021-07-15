package beacon

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
)

type LocalBeaconer struct {
	Logger  log.Logger
	Metrics *metrics.BeaconMetrics
}

func (local *LocalBeaconer) Submit(ctx context.Context, entry *transport.NextBeaconPacket) error {
	local.Metrics.EntriesSubmitted.Add(1)

	if local.Logger == nil {
		return errors.New("no logger for local beaconer, can't display entry")
	}

	level.Info(local.Logger).Log("msg", "submitted beacon entry")

	output := fmt.Sprintf("%#v", entry)
	level.Debug(local.Logger).Log("entry", output)

	local.Metrics.EntriesFlushed.Add(1)

	return nil
}

func (local *LocalBeaconer) Close() {}
