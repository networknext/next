package billing

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
	DefaultBigQueryBatchSize   = 1000
	DefaultBigQueryChannelSize = 20000
)

type GoogleBigQueryClient struct {
	Metrics              *metrics.BillingMetrics
	TableInserter        *bigquery.Inserter
	SummaryTableInserter *bigquery.Inserter
	BatchSize            int
	SummaryBatchSize     int
	BatchSizePercent     float64
	FeatureBilling2      bool

	buffer2      []*BillingEntry2
	bufferMutex2 sync.RWMutex

	summaryBuffer2      []*BillingEntry2Summary
	summaryBufferMutex2 sync.RWMutex

	entries2 chan *BillingEntry2

	summaryEntries2 chan *BillingEntry2Summary
}

// Bill2 pushes a BillingEntry2 to the entries2 channel
// Entries with summary slices are pushed to the summaryEntries2 channel
func (bq *GoogleBigQueryClient) Bill2(ctx context.Context, entry *BillingEntry2) error {
	if bq.entries2 == nil {
		bq.entries2 = make(chan *BillingEntry2, DefaultBigQueryChannelSize)
	}
	if bq.summaryEntries2 == nil {
		bq.summaryEntries2 = make(chan *BillingEntry2Summary, DefaultBigQueryChannelSize)
	}

	if entry.Summary {
		bq.summaryBufferMutex2.RLock()
		bufferLength := len(bq.summaryEntries2)
		bq.summaryBufferMutex2.RUnlock()

		if bufferLength >= bq.SummaryBatchSize {
			return &ErrSummaryEntries2BufferFull{}
		}

		bq.Metrics.SummaryEntries2Submitted.Add(1)
	} else {
		bq.bufferMutex2.RLock()
		bufferLength := len(bq.buffer2)
		bq.bufferMutex2.RUnlock()

		if bufferLength >= bq.BatchSize {
			return &ErrEntries2BufferFull{}
		}

		bq.Metrics.Entries2Submitted.Add(1)
	}

	hasNanOrInf, nanOrInfFields := entry.CheckNaNOrInf()
	if hasNanOrInf {
		bq.Metrics.ErrorMetrics.Billing2EntriesWithNaN.Add(1)
		core.Debug("billing entry 2 had NaN or Inf values for %s\n%+v", strings.Join(nanOrInfFields, " "), entry)
	}

	if !entry.Validate() {
		bq.Metrics.ErrorMetrics.Billing2InvalidEntries.Add(1)
		core.Error("billing entry 2 not valid\n%+v", entry)
		return errors.New("invalid billing entry 2")
	}

	if entry.Summary {
		select {
		case bq.summaryEntries2 <- entry.GetSummaryStruct():
			return nil
		default:
			return errors.New("summaryEntries2 channel full")
		}
	} else {
		select {
		case bq.entries2 <- entry:
			return nil
		default:
			return errors.New("entries2 channel full")
		}
	}
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

			if float64(bufferLength) >= float64(bq.BatchSize)*bq.BatchSizePercent {
				if err := bq.TableInserter.Put(context.Background(), bq.buffer2); err != nil {
					bq.bufferMutex2.Unlock()

					core.Error("failed to write buffer2 to BigQuery: %v", err)
					bq.Metrics.ErrorMetrics.Billing2WriteFailure.Add(float64(bufferLength))
					continue
				}
				bq.buffer2 = bq.buffer2[:0]

				core.Debug("flushed %d billing entries 2 to BigQuery", bufferLength)
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

				core.Error("failed to write buffer2 to BigQuery: %v", err)
				bq.Metrics.ErrorMetrics.Billing2WriteFailure.Add(float64(bufferLength))
				return err
			}
			bq.buffer2 = bq.buffer2[:0]
			bq.bufferMutex2.Unlock()

			core.Debug("final flush of %d billing entries 2 to BigQuery", bufferLength)
			bq.Metrics.Entries2Flushed.Add(float64(bufferLength))

			return nil
		}
	}
}

// SummaryWriteLoop2 ranges over the incoming summary channel of BillingEntry2 types and fills an internal buffer.
// Once the buffer fills to the SummaryBatchSize it will write all entries to BigQuery. This should
// only be called from 1 goroutine to avoid using a mutex around the internal buffer slice
func (bq *GoogleBigQueryClient) SummaryWriteLoop2(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	if bq.summaryEntries2 == nil {
		bq.summaryEntries2 = make(chan *BillingEntry2Summary, DefaultBigQueryChannelSize)
	}
	for {
		select {
		case entry := <-bq.summaryEntries2:
			bq.Metrics.SummaryEntries2Queued.Set(float64(len(bq.summaryEntries2)))
			bq.summaryBufferMutex2.Lock()
			bq.summaryBuffer2 = append(bq.summaryBuffer2, entry)
			bufferLength := len(bq.summaryBuffer2)

			if float64(bufferLength) >= float64(bq.SummaryBatchSize)*bq.BatchSizePercent {
				if err := bq.SummaryTableInserter.Put(context.Background(), bq.summaryBuffer2); err != nil {
					bq.summaryBufferMutex2.Unlock()

					core.Error("failed to write summaryBuffer2 to BigQuery: %v", err)
					bq.Metrics.ErrorMetrics.Billing2WriteFailure.Add(float64(bufferLength))
					continue
				}

				bq.summaryBuffer2 = bq.summaryBuffer2[:0]

				core.Debug("flushed %d summary billing entries 2 to BigQuery", bufferLength)
				bq.Metrics.BillingEntry2Size.Set(float64(bufferLength * MaxBillingEntry2Bytes))
				bq.Metrics.SummaryEntries2Flushed.Add(float64(bufferLength))
			}

			bq.summaryBufferMutex2.Unlock()
		case <-ctx.Done():
			var bufferLength int

			// Received shutdown signal, write remaining entries to BigQuery
			bq.summaryBufferMutex2.Lock()
			for entry := range bq.summaryEntries2 {
				// Add the remaining entries to the buffer
				bq.summaryBuffer2 = append(bq.summaryBuffer2, entry)
				bufferLength = len(bq.summaryBuffer2)
				bq.Metrics.SummaryEntries2Queued.Set(float64(bufferLength))
			}

			// Emptied out the entries channel, flush to BigQuery
			if err := bq.SummaryTableInserter.Put(context.Background(), bq.summaryBuffer2); err != nil {
				bq.summaryBufferMutex2.Unlock()

				core.Error("failed to write summaryBuffer2 to BigQuery: %v", err)
				bq.Metrics.ErrorMetrics.Billing2WriteFailure.Add(float64(bufferLength))
				return err
			}
			bq.summaryBuffer2 = bq.summaryBuffer2[:0]
			bq.summaryBufferMutex2.Unlock()

			core.Debug("final flush of %d summary billing entries 2 to BigQuery", bufferLength)
			bq.Metrics.SummaryEntries2Flushed.Add(float64(bufferLength))

			return nil
		}
	}
}

// FlushBuffer satisfies the Biller interface, actual flushing is done in write loop
func (bq *GoogleBigQueryClient) FlushBuffer(ctx context.Context) {}

// Closes the entries channel. Should only be done by the entry sender.
func (bq *GoogleBigQueryClient) Close() {
	if bq.FeatureBilling2 {
		close(bq.entries2)
		close(bq.summaryEntries2)
	}
}
