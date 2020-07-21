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
	Write(ctx context.Context, entry *StatsEntry)
}

type NoOpBigQueryWriter struct {
	written uint64
}

type GoogleBigQueryClient struct {
	Metrics       *metrics.AnalyticsMetrics
	Logger        log.Logger
	TableInserter *bigquery.Inserter

	entries chan *StatsEntry

	written uint64
	flushed uint64
}

func NewGoogleBigQueryClient(client *bigquery.Client, logger log.Logger, metrics *metrics.AnalyticsMetrics, dataset, table string) GoogleBigQueryClient {
	return GoogleBigQueryClient{
		Metrics:       metrics,
		Logger:        logger,
		TableInserter: client.Dataset(dataset).Table(table).Inserter(),
		entries:       make(chan *StatsEntry, DefaultBigQueryChannelSize),
	}
}

func (bq *GoogleBigQueryClient) Write(ctx context.Context, entry *StatsEntry) {
	atomic.AddUint64(&bq.written, 1)
	bq.entries <- entry
}

func (bq *GoogleBigQueryClient) WriteLoop(ctx context.Context) {
	for entry := range bq.entries {
		if err := bq.TableInserter.Put(ctx, entry); err != nil {
			level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
			bq.Metrics.ErrorMetrics.AnalyticsWriteFailure.Add(1)
		}
		atomic.AddUint64(&bq.flushed, 1)
		bq.Metrics.AnalyticsEntriesWritten.Add(1)
	}
}

func (bq *NoOpBigQueryWriter) Write(ctx context.Context, entry *StatsEntry) {
	atomic.AddUint64(&bq.written, 1)
}
