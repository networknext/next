package match_data

import (
	"context"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"
)

type LocalMatcher struct {
	Metrics *metrics.MatchDataMetrics
}

func (local *LocalMatcher) Match(ctx context.Context, entry *MatchDataEntry) error {

	local.Metrics.EntriesSubmitted.Add(1)
	core.Debug("submitted match data entry")

	local.Metrics.EntriesFlushed.Add(1)

	return nil
}

func (local *LocalMatcher) FlushBuffer(ctx context.Context) {}

func (local *LocalMatcher) Close() {}
