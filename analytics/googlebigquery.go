package analytics

import (
	"context"
	"sync/atomic"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/metrics"
)

const (
	DefaultBigQueryChannelSize = 10000
)

type PingStatsWriter interface {
	Write(ctx context.Context, entry *PingStatsEntry) error
}

type NoOpPingStatsWriter struct {
	written uint64
}

func (bq *NoOpPingStatsWriter) Write(ctx context.Context, entry *PingStatsEntry) error {
	atomic.AddUint64(&bq.written, 1)
	return nil
}

type GoogleBigQueryPingStatsWriter struct {
	Metrics       *metrics.AnalyticsMetrics
	Logger        log.Logger
	TableInserter *bigquery.Inserter

	entries chan *PingStatsEntry
}

func NewGoogleBigQueryPingStatsWriter(client *bigquery.Client, logger log.Logger, metrics *metrics.AnalyticsMetrics, dataset, table string) GoogleBigQueryPingStatsWriter {
	return GoogleBigQueryPingStatsWriter{
		Metrics:       metrics,
		Logger:        logger,
		TableInserter: client.Dataset(dataset).Table(table).Inserter(),
		entries:       make(chan *PingStatsEntry, DefaultBigQueryChannelSize),
	}
}

func (bq *GoogleBigQueryPingStatsWriter) Write(ctx context.Context, entry *PingStatsEntry) error {
	bq.Metrics.EntriesSubmitted.Add(1)
	bq.entries <- entry
	return nil
}

func (bq *GoogleBigQueryPingStatsWriter) WriteLoop(ctx context.Context) {
	for entry := range bq.entries {
		if err := bq.TableInserter.Put(ctx, entry); err != nil {
			level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
			bq.Metrics.ErrorMetrics.WriteFailure.Add(1)
		}
		bq.Metrics.EntriesFlushed.Add(1)
	}
}

type RelayStatsWriter interface {
	Write(ctx context.Context, entry *RelayStatsEntry) error
}

type NoOpRelayStatsWriter struct {
	submitted uint64
}

func (bq *NoOpRelayStatsWriter) Write(ctx context.Context, entry *RelayStatsEntry) error {
	atomic.AddUint64(&bq.submitted, 1)
	return nil
}

type GoogleBigQueryRelayStatsWriter struct {
	Metrics       *metrics.AnalyticsMetrics
	Logger        log.Logger
	TableInserter *bigquery.Inserter

	entries chan *RelayStatsEntry
}

func NewGoogleBigQueryRelayStatsWriter(client *bigquery.Client, logger log.Logger, metrics *metrics.AnalyticsMetrics, dataset, table string) GoogleBigQueryRelayStatsWriter {
	return GoogleBigQueryRelayStatsWriter{
		Metrics:       metrics,
		Logger:        logger,
		TableInserter: client.Dataset(dataset).Table(table).Inserter(),
		entries:       make(chan *RelayStatsEntry, DefaultBigQueryChannelSize),
	}
}

func (bq *GoogleBigQueryRelayStatsWriter) Write(ctx context.Context, entry *RelayStatsEntry) error {
	bq.Metrics.EntriesSubmitted.Add(1)
	bq.entries <- entry
	return nil
}

func (bq *GoogleBigQueryRelayStatsWriter) WriteLoop(ctx context.Context) {
	for entry := range bq.entries {
		if err := bq.TableInserter.Put(ctx, entry); err != nil {
			level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
			bq.Metrics.ErrorMetrics.WriteFailure.Add(1)
		}
		bq.Metrics.EntriesFlushed.Add(1)
	}
}
