package analytics

import (
	"context"
	"sync/atomic"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
)

const (
	DefaultBigQueryChannelSize = 10000
	PingStatsToPublishAtOnce   = 10000
)

type PingStatsWriter interface {
	Write(ctx context.Context, entries []*PingStatsEntry) error
}

type NoOpPingStatsWriter struct {
	written uint64
}

func (bq *NoOpPingStatsWriter) Write(ctx context.Context, entries []*PingStatsEntry) error {
	atomic.AddUint64(&bq.written, uint64(len(entries)))
	return nil
}

type GoogleBigQueryPingStatsWriter struct {
	Metrics       *metrics.AnalyticsMetrics
	Logger        log.Logger
	TableInserter *bigquery.Inserter

	entries chan []*PingStatsEntry
}

func NewGoogleBigQueryPingStatsWriter(client *bigquery.Client, logger log.Logger, metrics *metrics.AnalyticsMetrics, dataset, table string) GoogleBigQueryPingStatsWriter {
	return GoogleBigQueryPingStatsWriter{
		Metrics:       metrics,
		Logger:        logger,
		TableInserter: client.Dataset(dataset).Table(table).Inserter(),
		entries:       make(chan []*PingStatsEntry, DefaultBigQueryChannelSize),
	}
}

func (bq *GoogleBigQueryPingStatsWriter) Write(ctx context.Context, entries []*PingStatsEntry) error {
	bq.Metrics.EntriesSubmitted.Add(1)
	bq.entries <- entries
	return nil
}

func (bq *GoogleBigQueryPingStatsWriter) WriteLoop(ctx context.Context) {
	for entries := range bq.entries {
		fullBatches := len(entries) / PingStatsToPublishAtOnce
		for i := 0; i < fullBatches; i++ {
			if err := bq.TableInserter.Put(ctx, entries[i*PingStatsToPublishAtOnce:(i+1)*PingStatsToPublishAtOnce]); err != nil {
				level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
				bq.Metrics.ErrorMetrics.WriteFailure.Add(1)
			}
		}

		if len(entries[fullBatches*PingStatsToPublishAtOnce:]) > 0 {
			if err := bq.TableInserter.Put(ctx, entries[fullBatches*PingStatsToPublishAtOnce:]); err != nil {
				level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
				bq.Metrics.ErrorMetrics.WriteFailure.Add(1)
			}
		}

		bq.Metrics.EntriesFlushed.Add(1)
	}
}

type RelayStatsWriter interface {
	Write(ctx context.Context, entries []*RelayStatsEntry) error
}

type NoOpRelayStatsWriter struct {
	submitted uint64
}

func (bq *NoOpRelayStatsWriter) Write(ctx context.Context, entries []*RelayStatsEntry) error {
	atomic.AddUint64(&bq.submitted, uint64(len(entries)))
	return nil
}

type GoogleBigQueryRelayStatsWriter struct {
	Metrics       *metrics.AnalyticsMetrics
	Logger        log.Logger
	TableInserter *bigquery.Inserter

	entries chan []*RelayStatsEntry
}

func NewGoogleBigQueryRelayStatsWriter(client *bigquery.Client, logger log.Logger, metrics *metrics.AnalyticsMetrics, dataset, table string) GoogleBigQueryRelayStatsWriter {
	return GoogleBigQueryRelayStatsWriter{
		Metrics:       metrics,
		Logger:        logger,
		TableInserter: client.Dataset(dataset).Table(table).Inserter(),
		entries:       make(chan []*RelayStatsEntry, DefaultBigQueryChannelSize),
	}
}

func (bq *GoogleBigQueryRelayStatsWriter) Write(ctx context.Context, entries []*RelayStatsEntry) error {
	bq.Metrics.EntriesSubmitted.Add(1)
	bq.entries <- entries
	return nil
}

func (bq *GoogleBigQueryRelayStatsWriter) WriteLoop(ctx context.Context) {
	for entries := range bq.entries {
		if err := bq.TableInserter.Put(ctx, entries); err != nil {
			level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
			bq.Metrics.ErrorMetrics.WriteFailure.Add(1)
		}
		bq.Metrics.EntriesFlushed.Add(float64(len(entries)))
	}
}

type RelayNamesHashWriter interface {
	Write(ctx context.Context, entry *RelayNamesHashEntry) error
}

type NoOpRelayNamesHashWriter struct {
	written uint64
}

func (bq *NoOpRelayNamesHashWriter) Write(ctx context.Context, entry *RelayNamesHashEntry) error {
	atomic.AddUint64(&bq.written, 1)
	return nil
}

type GoogleBigQueryRelayNamesHashWriter struct {
	Metrics       *metrics.AnalyticsMetrics
	Logger        log.Logger
	TableInserter *bigquery.Inserter

	entries chan *RelayNamesHashEntry
}

func NewGoogleBigQueryRelayNamesHashWriter(client *bigquery.Client, logger log.Logger, metrics *metrics.AnalyticsMetrics, dataset, table string) (GoogleBigQueryRelayNamesHashWriter, error) {

	bqTable := client.Dataset(dataset).Table(table)
	meta, err := bqTable.Metadata(context.Background())
	if meta == nil || err != nil {
		schema, err := bigquery.InferSchema(RelayNamesHashEntry{})
		if err != nil {
			level.Error(logger).Log("err", err)
			return GoogleBigQueryRelayNamesHashWriter{}, err
		}

		tblMeta := &bigquery.TableMetadata{
			Name:        table,
			Description: "Relay Names and Hash Table",
			Schema:      schema,
		}
		err = bqTable.Create(context.Background(), tblMeta)
		if err != nil {
			level.Error(logger).Log("err", err)
			return GoogleBigQueryRelayNamesHashWriter{}, err
		}
	}

	return GoogleBigQueryRelayNamesHashWriter{
		Metrics:       metrics,
		Logger:        logger,
		TableInserter: bqTable.Inserter(),
		entries:       make(chan *RelayNamesHashEntry, DefaultBigQueryChannelSize),
	}, nil
}

func (bq *GoogleBigQueryRelayNamesHashWriter) Write(ctx context.Context, entries *RelayNamesHashEntry) error {
	bq.Metrics.EntriesSubmitted.Add(1)
	bq.entries <- entries
	return nil
}

func (bq *GoogleBigQueryRelayNamesHashWriter) WriteLoop(ctx context.Context) {
	for entries := range bq.entries {
		if err := bq.TableInserter.Put(ctx, entries); err != nil {
			level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
			bq.Metrics.ErrorMetrics.WriteFailure.Add(1)
		}
		bq.Metrics.EntriesFlushed.Add(1)
	}
}
