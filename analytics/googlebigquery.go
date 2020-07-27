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

type BigQueryWriter interface {
	Write(ctx context.Context, entry *StatsEntry) error
}

type NoOpBigQueryWriter struct {
	written uint64
}

func (bq *NoOpBigQueryWriter) Write(ctx context.Context, entry *StatsEntry) error {
	atomic.AddUint64(&bq.written, 1)
	return nil
}

type GoogleBigQueryWriter struct {
	Metrics       *metrics.AnalyticsMetrics
	Logger        log.Logger
	TableInserter *bigquery.Inserter

	entries chan *StatsEntry
}

func NewGoogleBigQueryWriter(client *bigquery.Client, logger log.Logger, metrics *metrics.AnalyticsMetrics, dataset, table string) GoogleBigQueryWriter {
	return GoogleBigQueryWriter{
		Metrics:       metrics,
		Logger:        logger,
		TableInserter: client.Dataset(dataset).Table(table).Inserter(),
		entries:       make(chan *StatsEntry, DefaultBigQueryChannelSize),
	}
}

func (bq *GoogleBigQueryWriter) Write(ctx context.Context, entry *StatsEntry) error {
	bq.Metrics.EntriesSubmitted.Add(1)
	bq.entries <- entry
	return nil
}

func (bq *GoogleBigQueryWriter) WriteLoop(ctx context.Context) {
	for entry := range bq.entries {
		if err := bq.TableInserter.Put(ctx, entry); err != nil {
			level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
			bq.Metrics.ErrorMetrics.WriteFailure.Add(1)
		}
		bq.Metrics.EntriesFlushed.Add(1)
	}
}
