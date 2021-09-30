package analytics

import (
	"context"

	"github.com/networknext/backend/modules/metrics"
)

type LocalPingStatsWriter struct {
	Metrics *metrics.AnalyticsMetrics
}

func (writer *LocalPingStatsWriter) Write(ctx context.Context, entries []*PingStatsEntry) error {
	writer.Metrics.EntriesSubmitted.Add(1)

	core.Debug("wrote analytics ping stats entries")

	for i := range entries {
		entry := entries[i]
		core.Debug("entry %d: %+v", i, *entry)
	}

	writer.Metrics.EntriesFlushed.Add(1)

	return nil
}

type LocalRelayStatsWriter struct {
	Metrics *metrics.AnalyticsMetrics
}

func (writer *LocalRelayStatsWriter) Write(ctx context.Context, entries []*RelayStatsEntry) error {
	writer.Metrics.EntriesSubmitted.Add(1)

	core.Debug("wrote analytics relay stats entries")

	for i := range entries {
		entry := entries[i]
		core.Debug("entry %d: %+v", i, *entry)
	}

	return nil
}
