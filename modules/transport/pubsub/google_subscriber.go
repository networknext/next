package pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	googlepubsub "cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
)

const (
	DefaultBigQueryBatchSize   = 1000
	DefaultBigQueryChannelSize = 10000
)

// PubSubSubscriber pulls data from Google Pub/Sub and writes them to BigQuery.
type PubSubSubscriber struct {
	Logger        log.Logger
	Metrics       *metrics.GoogleSubscriberMetrics
	TableInserter *bigquery.Inserter
	BatchSize     int
	ChannelSize   int
	WriteDuration time.Duration
	WriteTimeout  time.Duration

	PubSubSubscription *googlepubsub.Subscription
	buffer             []*Entry
	bufferMutex        sync.RWMutex

	entries   chan *Entry
	EntryVeto bool

	AckMap      map[*Entry]*googlepubsub.Message
	AckMapMutex sync.RWMutex
}

func NewPubSubSubscriber(
	ctx context.Context,
	logger log.Logger,
	metrics *metrics.GoogleSubscriberMetrics,
	gcpProjectID string,
	topicID string,
	subscriptionID string,
	recvNumGoroutines int,
	subscriberDatasetName string,
	subscriberTableName string,
	subscriberBatchSize int,
	subscriberChannelSize int,
	subscriberWriteDuration time.Duration,
	subscriberWriteTimeout time.Duration,
	entryVeto bool,
) (*PubSubSubscriber, error) {
	pubsubClient, err := googlepubsub.NewClient(ctx, gcpProjectID)
	if err != nil {
		return nil, fmt.Errorf("NewPubSubSubscriber(): could not create pubsub client: %v", err)
	}

	// Verify the subscriptionID exists
	subscriber := pubsubClient.Subscription(subscriptionID)
	ok, err := subscriber.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewPubSubSubscriber(): could not verify if subscription %s exists", subscriptionID)
	}
	if !ok {
		return nil, fmt.Errorf("NewPubSubSubscriber(): subscription %s does not exists", subscriptionID)
	}

	// Change the number of receive goroutines to the default if given less than the default
	if recvNumGoroutines < googlepubsub.DefaultReceiveSettings.NumGoroutines {
		recvNumGoroutines = googlepubsub.DefaultReceiveSettings.NumGoroutines
	}

	// Update the default receive settings to allow for unlimited number of messages and bytes on unprocessed messages (unacked but not yet expired)
	// and set the number of receive goroutines
	recvSettings := googlepubsub.ReceiveSettings{
		MaxOutstandingMessages: -1,
		MaxOutstandingBytes:    -1,
		NumGoroutines:          recvNumGoroutines,
	}
	subscriber.ReceiveSettings = recvSettings

	// Create BigQuery client
	subscriberClient, err := bigquery.NewClient(ctx, gcpProjectID)
	if err != nil {
		return nil, fmt.Errorf("NewPubSubSubscriber(): could not create BigQuery client: %v", err)
	}

	// Ensure the BigQuery dataset and table exists
	subscriberDataset := subscriberClient.Dataset(subscriberDatasetName)
	subscriberDatasetMetadata, err := subscriberDataset.Metadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewPubSubSubscriber(): could not access BigQuery dataset %s metadata: %v", subscriberDatasetName, err)
	}
	if subscriberDatasetMetadata.Name == "" {
		return nil, fmt.Errorf("NewPubSubSubscriber(): BigQuery dataset %s is invalid", subscriberDatasetName)
	}
	subscriberTable := subscriberDataset.Table(subscriberTableName)
	subscriberTableMetadata, err := subscriberTable.Metadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewPubSubSubscriber(): could not access BigQuery table %s metadata: %v", subscriberTableName, err)
	}
	if subscriberTableMetadata == nil || subscriberTableMetadata.Name == "" {
		return nil, fmt.Errorf("NewPubSubSubscriber(): BigQuery table %s does not exist", subscriberTableName)
	}

	// Force batch size to default size if invalid batch size provided
	if subscriberBatchSize < 1 {
		subscriberBatchSize = DefaultBigQueryBatchSize
	}

	// Force channel size to default size if provided size is less than batch size
	if subscriberChannelSize < subscriberBatchSize {
		subscriberChannelSize = DefaultBigQueryChannelSize
	}

	// Force write timeout to be <= 25 seconds because when context is canceled for shutdown,
	// we only have 30 seconds to finish cleaning up before a SIGKILL is sent
	if subscriberWriteTimeout > time.Second*25 {
		subscriberWriteTimeout = time.Second * 25
	}

	return &PubSubSubscriber{
		Logger:        logger,
		Metrics:       metrics,
		TableInserter: subscriberTable.Inserter(),
		BatchSize:     subscriberBatchSize,
		ChannelSize:   subscriberChannelSize,
		WriteDuration: subscriberWriteDuration,
		WriteTimeout:  subscriberWriteTimeout,

		PubSubSubscription: subscriber,
		EntryVeto:          entryVeto,
		AckMap:             make(map[*Entry]*googlepubsub.Message),
	}, nil
}

// Spawns goroutines to receive messages from Google Pub/Sub and submits them to the internal channel.
// Messages are acked when they are successfully written to BigQuery, and are otherwise nacked.
func (subscriber *PubSubSubscriber) ReceiveAndSubmit(ctx context.Context) error {
	err := subscriber.PubSubSubscription.Receive(ctx, func(ctx context.Context, m *googlepubsub.Message) {
		entries, err := subscriber.unbatchMessages(m)
		if err != nil {
			level.Error(subscriber.Logger).Log("err", err)
			subscriber.Metrics.ErrorMetrics.BatchedReadFailure.Add(1)
		}

		subscriber.Metrics.EntriesReceived.Add(float64(len(entries)))

		unbatchedEntries := make([]Entry, len(entries))
		for i := range unbatchedEntries {
			if (unbatchedEntries[i]).ReadEntry(entries[i]) {

				// If we are on the last entry for this message, add the message to the ack map to be acked after writing to BigQuery
				if i == len(entries)-1 {
					subscriber.AckMapMutex.Lock()
					subscriber.AckMap[&unbatchedEntries[i]] = m
					subscriber.AckMapMutex.Unlock()
				}

				// Submit the entry for writing to BigQuery
				if err := subscriber.submit(context.Background(), &unbatchedEntries[i]); err != nil {
					subscriber.Metrics.ErrorMetrics.QueueFailure.Add(1)
					level.Error(subscriber.Logger).Log("msg", "PubSubSubscriber Receive(): could not submit entry", "err", err)

					// Delete the entry from the map if fail to submit
					if i == len(entries)-1 {
						subscriber.AckMapMutex.Lock()
						delete(subscriber.AckMap, &unbatchedEntries[i])
						subscriber.AckMapMutex.Unlock()
					}

					// Nack the message so it can be redelivered
					m.Nack()
					return
				}
			} else {
				if subscriber.EntryVeto {
					m.Ack()
					return
				}

				// Nack if we failed to read the billing entry
				m.Nack()
				subscriber.Metrics.ErrorMetrics.ReadFailure.Add(1)
			}
		}
	})

	if err != nil && err != context.Canceled {
		// If the Receive function returns for any reason besides shutdown, we want to immediately exit and restart the service
		level.Error(subscriber.Logger).Log("msg", "stopped receive loop", "err", err)
		return err
	}

	// Close entries channel to ensure messages are drained for the final write to BigQuery
	subscriber.Close()
	level.Debug(subscriber.Logger).Log("msg", "receive canceled, closed entries channel")

	return nil
}

// Unbatches a message from Google Pub/Sub since we do our own batching.
func (subscriber *PubSubSubscriber) unbatchMessages(m *googlepubsub.Message) ([][]byte, error) {
	messages := make([][]byte, 0)

	var offset int
	for {
		if offset >= len(m.Data) {
			break
		}

		var messageLength uint32
		var message []byte
		if !encoding.ReadUint32(m.Data, &offset, &messageLength) {
			return nil, fmt.Errorf("PubSubSubscriber UnbatchMessages(): failed to read batched message length at offset %d (length %d)", offset, len(m.Data))
		}

		if !encoding.ReadBytes(m.Data, &offset, &message, messageLength) {
			return nil, fmt.Errorf("PubSubSubscriber UnbatchMessages(): failed to read batched message bytes at offset %d (length %d)", offset, len(m.Data))
		}

		messages = append(messages, message)
	}

	return messages, nil
}

// Submits the entries to the internal channel if space is available
func (subscriber *PubSubSubscriber) submit(ctx context.Context, entry *Entry) error {
	if subscriber.entries == nil {
		subscriber.entries = make(chan *Entry, subscriber.ChannelSize)
	}

	subscriber.bufferMutex.RLock()
	bufferLength := len(subscriber.buffer)
	subscriber.bufferMutex.RUnlock()

	if bufferLength >= subscriber.BatchSize {
		return errors.New("PubSubSubscriber submit(): entries buffer full")
	}

	select {
	case subscriber.entries <- entry:
		return nil
	default:
		return errors.New("PubSubSubscriber submit(): entries channel full")
	}
}

// Closes the entries channel. Should only be done by the entry sender.
func (subscriber *PubSubSubscriber) Close() {
	close(subscriber.entries)
}

// WriteLoop ranges over the incoming channel of Entry types and fills an internal buffer.
// Once the buffer fills to the BatchSize it will write all entries to BigQuery. This should
// only be called from 1 goroutine to avoid using a mutex around the internal buffer slice.
func (subscriber *PubSubSubscriber) WriteLoop(ctx context.Context) error {
	if subscriber.entries == nil {
		subscriber.entries = make(chan *Entry, subscriber.ChannelSize)
	}

	// Track which messages to ack once they have been written successfully
	messagesToAck := make([]*googlepubsub.Message, subscriber.BatchSize)
	lastWriteTime := time.Now()

	for {
		select {
		case entry := <-subscriber.entries:
			subscriber.Metrics.EntriesQueuedToWrite.Set(float64(len(subscriber.entries)))
			subscriber.bufferMutex.Lock()
			subscriber.buffer = append(subscriber.buffer, entry)
			bufferLength := len(subscriber.buffer)

			// See if any of these entries have a message that needs to be acked
			subscriber.AckMapMutex.RLock()
			message, exists := subscriber.AckMap[entry]
			subscriber.AckMapMutex.RUnlock()
			if exists {
				messagesToAck = append(messagesToAck, message)
				// Delete the entry and message from the ack map
				subscriber.AckMapMutex.Lock()
				delete(subscriber.AckMap, entry)
				subscriber.AckMapMutex.Unlock()
			}

			// Create context with timeout to not indefinitely retry writing to BigQuery
			ctxTimeout, cancel := context.WithTimeout(context.Background(), subscriber.WriteTimeout)

			// Write to BigQuery when buffer reached batch size, or it has been longer than WriteDuration since the last write
			if bufferLength >= subscriber.BatchSize || time.Since(lastWriteTime) > subscriber.WriteDuration {
				if err := subscriber.TableInserter.Put(ctxTimeout, subscriber.buffer); err != nil {
					// Cancel the context with timeout
					cancel()

					subscriber.bufferMutex.Unlock()
					// Nack all messages
					for _, nackMessage := range messagesToAck {
						nackMessage.Nack()
					}

					// Critical error, log to stdout
					level.Error(subscriber.Logger).Log("msg", "failed to write to BigQuery", "err", err)
					fmt.Printf("PubSubSubscriber WriteLoop(): failed to write %d entries to BigQuery: %v\n", bufferLength, err)
					subscriber.Metrics.ErrorMetrics.WriteFailure.Add(float64(bufferLength))
				} else {
					// Cancel the context with timeout
					cancel()

					// Ack all messages
					for _, ackMessage := range messagesToAck {
						ackMessage.Ack()
					}

					// Clear the buffers
					messagesToAck = messagesToAck[:0]
					subscriber.buffer = subscriber.buffer[:0]

					level.Info(subscriber.Logger).Log("msg", "wrote entries to BigQuery", "size", subscriber.BatchSize, "total", bufferLength)
					subscriber.Metrics.EntriesSubmitted.Add(float64(bufferLength))
				}

				// Reset last attempted write
				lastWriteTime = time.Now()
			}

			subscriber.bufferMutex.Unlock()
		case <-ctx.Done():
			subscriber.Metrics.EntriesQueuedToWrite.Set(float64(len(subscriber.entries)))
			var bufferLength int

			// Received shutdown signal, write remaining entries to BigQuery
			subscriber.bufferMutex.Lock()
			for entry := range subscriber.entries {
				subscriber.buffer = append(subscriber.buffer, entry)
				bufferLength = len(subscriber.buffer)

				// See if any of these entries have a message that needs to be acked
				subscriber.AckMapMutex.RLock()
				message, exists := subscriber.AckMap[entry]
				subscriber.AckMapMutex.RUnlock()
				if exists {
					messagesToAck = append(messagesToAck, message)
					// Delete the entry and message from the ack map
					subscriber.AckMapMutex.Lock()
					delete(subscriber.AckMap, entry)
					subscriber.AckMapMutex.Unlock()
				}
			}

			// Nack all remaining messages in AckMap
			for _, message := range subscriber.AckMap {
				message.Nack()
			}

			// Create context with timeout to not indefinitely retry writing to BigQuery
			ctxTimeout, cancel := context.WithTimeout(context.Background(), subscriber.WriteTimeout)

			// Emptied out the entries channel, flush to BigQuery
			if err := subscriber.TableInserter.Put(ctxTimeout, subscriber.buffer); err != nil {
				// Cancel the context with timeout
				cancel()

				// Nack all messages
				for _, nackMessage := range messagesToAck {
					nackMessage.Nack()
				}

				// Critical error, log to stdout
				level.Error(subscriber.Logger).Log("msg", "failed final write to BigQuery", "err", err)
				fmt.Printf("PubSubSubscriber WriteLoop(): failed to write %d entries to BigQuery: %v\n", bufferLength, err)

				subscriber.Metrics.ErrorMetrics.WriteFailure.Add(float64(bufferLength))
			} else {
				// Cancel the context with timeout
				cancel()

				// Ack all messages
				for _, ackMessage := range messagesToAck {
					ackMessage.Ack()
				}

				// Clear the buffers
				messagesToAck = messagesToAck[:0]
				subscriber.buffer = subscriber.buffer[:0]

				level.Info(subscriber.Logger).Log("msg", "wrote final entries to BigQuery", "size", bufferLength, "total", bufferLength)
				fmt.Printf("Final flush of %d entries to BigQuery.\n", bufferLength)

				subscriber.Metrics.EntriesSubmitted.Add(float64(bufferLength))
			}

			subscriber.bufferMutex.Unlock()
			return nil
		}
	}

	return nil
}
