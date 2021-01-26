package beacon

import (
	"context"
	"errors"
	"sync"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
)

const (
	DefaultBigQueryBatchSize   = 1000
	DefaultBigQueryChannelSize = 10000
)

type GoogleBigQueryClient struct {
	Metrics       *metrics.BeaconMetrics
	Logger        log.Logger
	TableInserter *bigquery.Inserter
	BatchSize     int

	buffer      []*NextBeaconPacket
	bufferMutex sync.RWMutex

	entries chan *NextBeaconPacket
}

// Submit pushes an Entry to the channel
func (bq *GoogleBigQueryClient) Submit(ctx context.Context, entry *NextBeaconPacket) error {
	if bq.entries == nil {
		bq.entries = make(chan *NextBeaconPacket, DefaultBigQueryChannelSize)
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
func (bq *GoogleBigQueryClient) WriteLoop(ctx context.Context) error {
	if bq.entries == nil {
		bq.entries = make(chan *NextBeaconPacket, DefaultBigQueryChannelSize)
	}
	for entry := range bq.entries {
		bq.Metrics.EntriesQueued.Set(float64(len(bq.entries)))

		bq.bufferMutex.Lock()
		bq.buffer = append(bq.buffer, entry)
		bufferLength := len(bq.buffer)
		if bufferLength >= bq.BatchSize {
			if err := bq.TableInserter.Put(ctx, bq.buffer); err != nil {
				bq.bufferMutex.Unlock()

				level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
				bq.Metrics.ErrorMetrics.BeaconWriteFailure.Add(float64(bufferLength))
				continue
			}

			bq.buffer = bq.buffer[:0]

			level.Info(bq.Logger).Log("msg", "flushed entries to BigQuery", "size", bq.BatchSize, "total", bufferLength)
			bq.Metrics.EntriesFlushed.Add(float64(bufferLength))
		}

		bq.bufferMutex.Unlock()
	}
	return nil
}

// Save implements the bigquery.ValueSaver interface for an Entry
// so it can be used in Put()
func (entry *NextBeaconPacket) Save() (map[string]bigquery.Value, string, error) {
	e := make(map[string]bigquery.Value)

	e["version"] = int(entry.Version)
	e["customerID"] = int(entry.CustomerID)
	e["datacenterID"] = int(entry.DatacenterID)
	e["userHash"] = int(entry.UserHash)
	e["addressHash"] = int(entry.AddressHash)
	e["sessionID"] = int(entry.SessionID)
	e["platformID"] = int(entry.PlatformID)
	e["connectionType"] = int(entry.ConnectionType)

	e["enabled"] = entry.Enabled
	e["upgraded"] = entry.Upgraded
	e["next"] = entry.Next
	e["fallbackToDirect"] = entry.FallbackToDirect

	return e, "", nil
}
