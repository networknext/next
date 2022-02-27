package match_data

import (
	"context"
	"errors"
	"strings"
	"sync"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"
)

const (
	DefaultBigQueryBatchSize   = 100
	DefaultBigQueryChannelSize = 20000
)

type GoogleBigQueryClient struct {
	Metrics              *metrics.MatchDataMetrics
	TableInserter        *bigquery.Inserter
	BatchSize            int
	BatchSizePercent     float64

	buffer      []*MatchDataEntry
	bufferMutex sync.RWMutex

	entries chan *MatchDataEntry
}

// Match pushes a MatchDataEntry to the entries channel
func (bq *GoogleBigQueryClient) Match(ctx context.Context, entry *MatchDataEntry) error {
	if bq.entries == nil {
		bq.entries = make(chan *MatchDataEntry, DefaultBigQueryChannelSize)
	}

	bq.bufferMutex.RLock()
	bufferLength := len(bq.buffer)
	bq.bufferMutex.RUnlock()

	if bufferLength >= bq.BatchSize {
		return &ErrEntriesBufferFull{}
	}

	bq.Metrics.EntriesSubmitted.Add(1)

	hasNanOrInf, nanOrInfFields := entry.CheckNaNOrInf()
	if hasNanOrInf {
		bq.Metrics.ErrorMetrics.MatchDataEntriesWithNaN.Add(1)
		core.Debug("match data entry had NaN or Inf values for %s\n%+v", strings.Join(nanOrInfFields, " "), entry)
	}

	if !entry.Validate() {
		bq.Metrics.ErrorMetrics.MatchDataInvalidEntries.Add(1)
		core.Error("match data entry not valid\n%+v", entry)
		return errors.New("invalid match data entry")
	}

	select {
	case bq.entries <- entry:
		return nil
	default:
		return errors.New("entries channel full")
	}
}

// WriteLoop ranges over the incoming channel of MatchDataEntry types and fills an internal buffer.
// Once the buffer fills to the BatchSize it will write all entries to BigQuery. This should
// only be called from 1 goroutine to avoid using a mutex around the internal buffer slice
func (bq *GoogleBigQueryClient) WriteLoop(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	if bq.entries == nil {
		bq.entries = make(chan *MatchDataEntry, DefaultBigQueryChannelSize)
	}
	for {
		select {
		case entry := <-bq.entries:
			bq.Metrics.EntriesQueued.Set(float64(len(bq.entries)))
			bq.bufferMutex.Lock()
			bq.buffer = append(bq.buffer, entry)
			bufferLength := len(bq.buffer)

			if float64(bufferLength) >= float64(bq.BatchSize)*bq.BatchSizePercent {
				if err := bq.TableInserter.Put(context.Background(), bq.buffer); err != nil {
					bq.bufferMutex.Unlock()

					core.Error("failed to write buffer to BigQuery: %v", err)
					bq.Metrics.ErrorMetrics.MatchDataWriteFailure.Add(float64(bufferLength))
					continue
				}
				bq.buffer = bq.buffer[:0]

				core.Debug("flushed %d match data entries to BigQuery", bufferLength)
				bq.Metrics.EntriesFlushed.Add(float64(bufferLength))
			}

			bq.bufferMutex.Unlock()
		case <-ctx.Done():
			var bufferLength int

			// Received shutdown signal, write remaining entries to BigQuery
			bq.bufferMutex.Lock()
			for entry := range bq.entries {
				// Add the remaining entries to the buffer
				bq.buffer = append(bq.buffer, entry)
				bufferLength = len(bq.buffer)
				bq.Metrics.EntriesQueued.Set(float64(bufferLength))
			}

			// Emptied out the entries channel, flush to BigQuery
			if err := bq.TableInserter.Put(context.Background(), bq.buffer); err != nil {
				bq.bufferMutex.Unlock()

				core.Error("failed to write buffer to BigQuery: %v", err)
				bq.Metrics.ErrorMetrics.MatchDataWriteFailure.Add(float64(bufferLength))
				return err
			}
			bq.buffer = bq.buffer[:0]
			bq.bufferMutex.Unlock()

			core.Debug("final flush of %d match data entries to BigQuery", bufferLength)
			bq.Metrics.EntriesFlushed.Add(float64(bufferLength))

			return nil
		}
	}
}

// FlushBuffer satisfies the Biller interface, actual flushing is done in write loop
func (bq *GoogleBigQueryClient) FlushBuffer(ctx context.Context) {}

// Closes the entries channel. Should only be done by the entry sender.
func (bq *GoogleBigQueryClient) Close() {
	close(bq.entries)
}
