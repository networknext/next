package analytics

import (
	"context"
	"sync"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"
)

const (
	DefaultBigQueryChannelSize = 10000
	PingStatsToPublishAtOnce   = 10000
)

type GoogleBigQueryPingStatsWriter struct {
	Metrics       *metrics.AnalyticsMetrics
	TableInserter *bigquery.Inserter

	entries chan []*PingStatsEntry
}

func NewGoogleBigQueryPingStatsWriter(client *bigquery.Client, metrics *metrics.AnalyticsMetrics, dataset, table string) GoogleBigQueryPingStatsWriter {
	return GoogleBigQueryPingStatsWriter{
		Metrics:       metrics,
		TableInserter: client.Dataset(dataset).Table(table).Inserter(),
		entries:       make(chan []*PingStatsEntry, DefaultBigQueryChannelSize),
	}
}

func (bq *GoogleBigQueryPingStatsWriter) Write(ctx context.Context, entries []*PingStatsEntry) error {
	select {
	case bq.entries <- entries:
		bq.Metrics.EntriesSubmitted.Add(float64(len(entries)))
		return nil
	default:
		return &ErrPingStatsChannelFull{}
	}
}

// WriteLoop() continues to write ping stats to BigQuery until the entries channel is closed by the PubSubForwarder
// We do not use the parent context in order to continue writing to BigQuery even if the parent context is canceled
func (bq *GoogleBigQueryPingStatsWriter) WriteLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	for entries := range bq.entries {
		fullBatches := len(entries) / PingStatsToPublishAtOnce

		for i := 0; i < fullBatches; i++ {
			entriesToWrite := entries[i*PingStatsToPublishAtOnce : (i+1)*PingStatsToPublishAtOnce]
			if err := bq.TableInserter.Put(context.Background(), entriesToWrite); err != nil {
				core.Error("failed to write ping stats to BigQuery: %v", err)
				bq.Metrics.ErrorMetrics.WriteFailure.Add(float64(len(entriesToWrite)))
			} else {
				bq.Metrics.EntriesFlushed.Add(float64(len(entriesToWrite)))
			}
		}

		remainingEntriesToWrite := entries[fullBatches*PingStatsToPublishAtOnce:]
		if len(remainingEntriesToWrite) > 0 {
			if err := bq.TableInserter.Put(context.Background(), remainingEntriesToWrite); err != nil {
				core.Error("failed to write ping stats to BigQuery: %v", err)
				bq.Metrics.ErrorMetrics.WriteFailure.Add(float64(len(remainingEntriesToWrite)))
			} else {
				bq.Metrics.EntriesFlushed.Add(float64(len(remainingEntriesToWrite)))
			}
		}
	}
}

func (bq *GoogleBigQueryPingStatsWriter) Close() {
	if bq.entries != nil {
		close(bq.entries)
	}
}

type GoogleBigQueryRelayStatsWriter struct {
	Metrics       *metrics.AnalyticsMetrics
	TableInserter *bigquery.Inserter

	entries chan []*RelayStatsEntry
}

func NewGoogleBigQueryRelayStatsWriter(client *bigquery.Client, metrics *metrics.AnalyticsMetrics, dataset, table string) GoogleBigQueryRelayStatsWriter {
	return GoogleBigQueryRelayStatsWriter{
		Metrics:       metrics,
		TableInserter: client.Dataset(dataset).Table(table).Inserter(),
		entries:       make(chan []*RelayStatsEntry, DefaultBigQueryChannelSize),
	}
}

func (bq *GoogleBigQueryRelayStatsWriter) Write(ctx context.Context, entries []*RelayStatsEntry) error {
	select {
	case bq.entries <- entries:
		bq.Metrics.EntriesSubmitted.Add(float64(len(entries)))
		return nil
	default:
		return &ErrRelayStatsChannelFull{}
	}
}

// WriteLoop() continues to write relay stats to BigQuery until the entries channel is closed by the PubSubForwarder
func (bq *GoogleBigQueryRelayStatsWriter) WriteLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	for entries := range bq.entries {
		if err := bq.TableInserter.Put(context.Background(), entries); err != nil {
			core.Error("failed to write relay stats to BigQuery: %v", err)
			bq.Metrics.ErrorMetrics.WriteFailure.Add(float64(len(entries)))
		}

		bq.Metrics.EntriesFlushed.Add(float64(len(entries)))
	}
}

func (bq *GoogleBigQueryRelayStatsWriter) Close() {
	if bq.entries != nil {
		close(bq.entries)
	}
}
