package analytics

import (
	"context"
)

// PingStatsWriter is an interface that handles publishing ping stats entries to Google Pub/Sub
type PingStatsPublisher interface {
	Publish(ctx context.Context, entries []PingStatsEntry) error
	Close()
}

// RelayStatsWriter is an interface that handles publishing relay stats entries to Google Pub/Sub
type RelayStatsPublisher interface {
	Publish(ctx context.Context, entries []RelayStatsEntry) error
	Close()
}

type ErrPingStatsChannelFull struct{}

func (e *ErrPingStatsChannelFull) Error() string {
	return fmt.Sprintf("ping stats channel full")
}

// PingStatsWriter is an interface that handles writing ping stats entries to BigQuery
type PingStatsWriter interface {
	Write(ctx context.Context, entries []*PingStatsEntry) error
	Close()
}

type ErrRelayStatsChannelFull struct{}

func (e *ErrRelayStatsChannelFull) Error() string {
	return fmt.Sprintf("relay stats channel full")
}

// RelayStatsWriter is an interface that handles writing relay stats entries to BigQuery
type RelayStatsWriter interface {
	Write(ctx context.Context, entries []*RelayStatsEntry) error
	Close()
}
