package beacon

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
)

const (
	DefaultBigQueryBatchSize   = 1000
	DefaultBigQueryChannelSize = 20000
)

type GoogleBigQueryClient struct {
	Metrics       *metrics.BeaconInserterMetrics
	Logger        log.Logger
	TableInserter *bigquery.Inserter
	BatchSize     int

	buffer      []*transport.NextBeaconPacket
	bufferMutex sync.RWMutex

	entries chan *transport.NextBeaconPacket
}

// Submit pushes an Entry to the channel
func (bq *GoogleBigQueryClient) Submit(ctx context.Context, entry *transport.NextBeaconPacket) error {
	if bq.entries == nil {
		bq.entries = make(chan *transport.NextBeaconPacket, DefaultBigQueryChannelSize)
	}

	bq.bufferMutex.RLock()
	bufferLength := len(bq.buffer)
	bq.bufferMutex.RUnlock()

	if bufferLength >= bq.BatchSize {
		return errors.New("entries buffer full")
	}

	select {
	case bq.entries <- entry:
		bq.Metrics.EntriesSubmitted.Add(1)
		return nil
	default:
		return errors.New("entries channel full")
	}
}

// WriteLoop ranges over the incoming channel of Entry types and fills an internal buffer.
// Once the buffer fills to the BatchSize it will write all entries to BigQuery. This should
// only be called from 1 goroutine to avoid using a mutex around the internal buffer slice
func (bq *GoogleBigQueryClient) WriteLoop(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	if bq.entries == nil {
		bq.entries = make(chan *transport.NextBeaconPacket, DefaultBigQueryChannelSize)
	}
	for {
		select {
		case entry := <-bq.entries:
			bq.Metrics.EntriesQueued.Set(float64(len(bq.entries)))
			bq.bufferMutex.Lock()
			bq.buffer = append(bq.buffer, entry)
			bufferLength := len(bq.buffer)

			if bufferLength >= bq.BatchSize {
				if err := bq.TableInserter.Put(context.Background(), bq.buffer); err != nil {
					bq.bufferMutex.Unlock()

					level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
					fmt.Printf("Failed to write to BigQuery: %v\n", err)

					bq.Metrics.ErrorMetrics.BeaconInserterWriteFailure.Add(float64(bufferLength))
					continue
				}

				bq.buffer = bq.buffer[:0]
				level.Info(bq.Logger).Log("msg", "flushed entries to BigQuery", "size", bq.BatchSize, "total", bufferLength)
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

				level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
				fmt.Printf("Failed to write to BigQuery: %v\n", err)

				bq.Metrics.ErrorMetrics.BeaconInserterWriteFailure.Add(float64(bufferLength))
				return err
			}
			bq.buffer = bq.buffer[:0]
			bq.bufferMutex.Unlock()

			level.Info(bq.Logger).Log("msg", "flushed entries to BigQuery", "size", bufferLength, "total", bufferLength)
			fmt.Printf("Final flush of %d entries to BigQuery.\n", bufferLength)

			bq.Metrics.EntriesFlushed.Add(float64(bufferLength))

			return nil
		}
	}
	return nil
}

// Closes the entries channel. Should only be done by the entry sender.
func (bq *GoogleBigQueryClient) Close() {
	close(bq.entries)
}
