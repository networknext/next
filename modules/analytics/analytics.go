package analytics

import (
	"context"
)

// PingStatsWriter is an interface that handles publishing ping stats entries to Google Pub/Sub
type PingStatsPublisher interface {
	Publish(ctx context.Context, entries []PingStatsEntry) error
}

// RelayStatsWriter is an interface that handles publishing relay stats entries to Google Pub/Sub
type RelayStatsPublisher interface {
	Publish(ctx context.Context, entries []RelayStatsEntry) error
}

// PingStatsWriter is an interface that handles writing ping stats entries to BigQuery
type PingStatsWriter interface {
	Write(ctx context.Context, entries []*PingStatsEntry) error
}

// RelayStatsWriter is an interface that handles writing relay stats entries to BigQuery
type RelayStatsWriter interface {
	Write(ctx context.Context, entries []*RelayStatsEntry) error
}
