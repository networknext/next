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
	NumSubmitted() uint64
	NumQueued() uint64
	NumFlushed() uint64
}

type NoOpBigQueryWriter struct {
	submitted uint64
}

func (bq *NoOpBigQueryWriter) Write(ctx context.Context, entry *StatsEntry) error {
	atomic.AddUint64(&bq.submitted, 1)
	return nil
}

func (writer *NoOpBigQueryWriter) NumSubmitted() uint64 {
	return atomic.LoadUint64(&writer.submitted)
}

func (writer *NoOpBigQueryWriter) NumQueued() uint64 {
	return 0
}

func (writer *NoOpBigQueryWriter) NumFlushed() uint64 {
	return atomic.LoadUint64(&writer.submitted)
}

type GoogleBigQueryWriter struct {
	Metrics       *metrics.AnalyticsMetrics
	Logger        log.Logger
	TableInserter *bigquery.Inserter

	entries chan *StatsEntry

	submitted uint64
	flushed   uint64
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
	atomic.AddUint64(&bq.submitted, 1)
	bq.entries <- entry
	return nil
}

func (bq *GoogleBigQueryWriter) WriteLoop(ctx context.Context) {
	for entry := range bq.entries {
		if err := bq.TableInserter.Put(ctx, entry); err != nil {
			level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
			bq.Metrics.ErrorMetrics.AnalyticsWriteFailure.Add(1)
		}
		atomic.AddUint64(&bq.flushed, 1)
		bq.Metrics.AnalyticsEntriesWritten.Add(1)
	}
}

func (bq *GoogleBigQueryWriter) NumSubmitted() uint64 {
	return atomic.LoadUint64(&bq.submitted)
}

func (bq *GoogleBigQueryWriter) NumQueued() uint64 {
	return uint64(len(bq.entries))
}

func (bq *GoogleBigQueryWriter) NumFlushed() uint64 {
	return atomic.LoadUint64(&bq.flushed)
}
