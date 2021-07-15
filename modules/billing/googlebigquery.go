package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
)

const (
	DefaultBigQueryBatchSize   = 1000
	DefaultBigQueryChannelSize = 20000
)

type GoogleBigQueryClient struct {
	Metrics         *metrics.BillingMetrics
	Logger          log.Logger
	TableInserter   *bigquery.Inserter
	BatchSize       int
	FeatureBilling  bool
	FeatureBilling2 bool

	buffer      []*BillingEntry
	bufferMutex sync.RWMutex

	buffer2      []*BillingEntry2
	bufferMutex2 sync.RWMutex

	entries chan *BillingEntry

	entries2 chan *BillingEntry2
}

// Bill pushes a BillingEntry to the entries channel
func (bq *GoogleBigQueryClient) Bill(ctx context.Context, entry *BillingEntry) error {
	bq.Metrics.EntriesSubmitted.Add(1)
	if bq.entries == nil {
		bq.entries = make(chan *BillingEntry, DefaultBigQueryChannelSize)
	}

	bq.bufferMutex.RLock()
	bufferLength := len(bq.buffer)
	bq.bufferMutex.RUnlock()

	if bufferLength >= bq.BatchSize {
		return errors.New("entries buffer full")
	}

	hasNanOrInf, nanOrInfFields := entry.CheckNaNOrInf()
	if hasNanOrInf {
		bq.Metrics.ErrorMetrics.BillingEntriesWithNaN.Add(1)
		fieldStr := strings.Join(nanOrInfFields, " ")
		fmt.Printf("Warn: billing entry had NaN or Inf values for %v.\n%+v\n", nanOrInfFields, entry)
		level.Warn(bq.Logger).Log("msg", "Billing entry had NaN or Inf values", "fields", fieldStr)
	}

	if !entry.Validate() {
		bq.Metrics.ErrorMetrics.BillingInvalidEntries.Add(1)
		fmt.Printf("Error: billing entry not valid.\n%+v\n", entry)
		return errors.New("invalid billing entry")
	}

	select {
	case bq.entries <- entry:
		return nil
	default:
		return errors.New("entries channel full")
	}
}

// Bill2 pushes a BillingEntry2 to the entries2 channel
func (bq *GoogleBigQueryClient) Bill2(ctx context.Context, entry *BillingEntry2) error {
	bq.Metrics.Entries2Submitted.Add(1)
	if bq.entries2 == nil {
		bq.entries2 = make(chan *BillingEntry2, DefaultBigQueryChannelSize)
	}

	bq.bufferMutex2.RLock()
	bufferLength := len(bq.buffer2)
	bq.bufferMutex2.RUnlock()

	if bufferLength >= bq.BatchSize {
		return errors.New("entries2 buffer full")
	}

	hasNanOrInf, nanOrInfFields := entry.CheckNaNOrInf()
	if hasNanOrInf {
		bq.Metrics.ErrorMetrics.Billing2EntriesWithNaN.Add(1)
		fieldStr := strings.Join(nanOrInfFields, " ")
		fmt.Printf("Warn: billing entry 2 had NaN or Inf values for %v.\n%+v\n", nanOrInfFields, entry)
		level.Warn(bq.Logger).Log("msg", "Billing entry 2 had NaN or Inf values", "fields", fieldStr)
	}

	if !entry.Validate() {
		bq.Metrics.ErrorMetrics.Billing2InvalidEntries.Add(1)
		fmt.Printf("Error: billing entry 2 not valid.\n%+v\n", entry)
		return errors.New("invalid billing entry 2")
	}

	select {
	case bq.entries2 <- entry:
		return nil
	default:
		return errors.New("entries2 channel full")
	}
}

// WriteLoop ranges over the incoming channel of BillingEntry types and fills an internal buffer.
// Once the buffer fills to the BatchSize it will write all entries to BigQuery. This should
// only be called from 1 goroutine to avoid using a mutex around the internal buffer slice
func (bq *GoogleBigQueryClient) WriteLoop(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	if bq.entries == nil {
		bq.entries = make(chan *BillingEntry, DefaultBigQueryChannelSize)
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

					level.Error(bq.Logger).Log("msg", "failed to write buffer to BigQuery", "err", err)
					fmt.Printf("Failed to write buffer to BigQuery: %v\n", err)

					bq.Metrics.ErrorMetrics.BillingWriteFailure.Add(float64(bufferLength))
					continue
				}

				bq.buffer = bq.buffer[:0]
				level.Info(bq.Logger).Log("msg", "flushed billing entries to BigQuery", "size", bq.BatchSize, "total", bufferLength)
				bq.Metrics.BillingEntrySize.Set(float64(bufferLength * MaxBillingEntryBytes))
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

				level.Error(bq.Logger).Log("msg", "failed to write buffer to BigQuery", "err", err)
				fmt.Printf("Failed to write buffer to BigQuery: %v\n", err)

				bq.Metrics.ErrorMetrics.BillingWriteFailure.Add(float64(bufferLength))
				return err
			}
			bq.buffer = bq.buffer[:0]
			bq.bufferMutex.Unlock()

			level.Info(bq.Logger).Log("msg", "flushed billing entries to BigQuery", "size", bufferLength, "total", bufferLength)
			fmt.Printf("Final flush of %d billing entries to BigQuery.\n", bufferLength)

			bq.Metrics.EntriesFlushed.Add(float64(bufferLength))

			return nil
		}
	}
	return nil
}

// WriteLoop2 ranges over the incoming channel of BillingEntry2 types and fills an internal buffer.
// Once the buffer fills to the BatchSize it will write all entries to BigQuery. This should
// only be called from 1 goroutine to avoid using a mutex around the internal buffer slice
func (bq *GoogleBigQueryClient) WriteLoop2(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	if bq.entries2 == nil {
		bq.entries2 = make(chan *BillingEntry2, DefaultBigQueryChannelSize)
	}
	for {
		select {
		case entry := <-bq.entries2:
			bq.Metrics.Entries2Queued.Set(float64(len(bq.entries2)))
			bq.bufferMutex2.Lock()
			bq.buffer2 = append(bq.buffer2, entry)
			bufferLength := len(bq.buffer2)

			if bufferLength >= bq.BatchSize {
				if err := bq.TableInserter.Put(context.Background(), bq.buffer2); err != nil {
					bq.bufferMutex2.Unlock()

					level.Error(bq.Logger).Log("msg", "failed to write buffer2 to BigQuery", "err", err)
					fmt.Printf("Failed to write buffer2 to BigQuery: %v\n", err)

					bq.Metrics.ErrorMetrics.Billing2WriteFailure.Add(float64(bufferLength))
					continue
				}

				bq.buffer2 = bq.buffer2[:0]
				level.Info(bq.Logger).Log("msg", "flushed billing entries 2 to BigQuery", "size", bq.BatchSize, "total", bufferLength)
				bq.Metrics.BillingEntry2Size.Set(float64(bufferLength * MaxBillingEntry2Bytes))
				bq.Metrics.Entries2Flushed.Add(float64(bufferLength))
			}

			bq.bufferMutex2.Unlock()
		case <-ctx.Done():
			var bufferLength int

			// Received shutdown signal, write remaining entries to BigQuery
			bq.bufferMutex2.Lock()
			for entry := range bq.entries2 {
				// Add the remaining entries to the buffer
				bq.buffer2 = append(bq.buffer2, entry)
				bufferLength = len(bq.buffer2)
				bq.Metrics.Entries2Queued.Set(float64(bufferLength))
			}

			// Emptied out the entries channel, flush to BigQuery
			if err := bq.TableInserter.Put(context.Background(), bq.buffer2); err != nil {
				bq.bufferMutex2.Unlock()

				level.Error(bq.Logger).Log("msg", "failed to write buffer2 to BigQuery", "err", err)
				fmt.Printf("Failed to write buffer2 to BigQuery: %v\n", err)

				bq.Metrics.ErrorMetrics.Billing2WriteFailure.Add(float64(bufferLength))
				return err
			}
			bq.buffer2 = bq.buffer2[:0]
			bq.bufferMutex2.Unlock()

			level.Info(bq.Logger).Log("msg", "flushed billing entries 2 to BigQuery", "size", bufferLength, "total", bufferLength)
			fmt.Printf("Final flush of %d billing entries 2 to BigQuery.\n", bufferLength)

			bq.Metrics.Entries2Flushed.Add(float64(bufferLength))

			return nil
		}
	}
	return nil
}

// Closes the entries channel. Should only be done by the entry sender.
func (bq *GoogleBigQueryClient) Close() {
	if bq.FeatureBilling {
		close(bq.entries)
	}
	if bq.FeatureBilling2 {
		close(bq.entries2)
	}
}
