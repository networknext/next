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

type PubSubSubscriber struct {
	Logger                log.Logger
	Metrics               *metrics.GoogleSubscriberMetrics
	TableInserter         *bigquery.Inserter
	BatchSize             int
	ChannelSize           int
	BigQueryWriteDuration time.Duration

	pubsubSubscriber *googlepubsub.Subscription
	buffer           []*Entry
	bufferMutex      sync.RWMutex

	entries   chan *Entry
	entryVeto bool

	ackMap      map[*Entry]*googlepubsub.Message
	ackMapMutex sync.RWMutex
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
	entryVeto bool,
) (*PubSubSubscriber, error) {
	pubsubClient, err := googlepubsub.NewClient(ctx, gcpProjectID)
	if err != nil {
		return nil, fmt.Errorf("NewPubSubSubscriber(): could not create pubsub client: %v", err)
	}

	if gcpProjectID == "local" {
		// Create the subscription for the local environment
		if _, err := pubsubClient.CreateSubscription(ctx, subscriptionID, googlepubsub.SubscriptionConfig{
			Topic: pubsubClient.Topic(topicID),
		}); err != nil && err.Error() != "rpc error: code = AlreadyExists desc = Subscription already exists" {
			// Not the best error check, but the underlying error type is internal so we can't check for it
			return nil, fmt.Errorf("NewPubSubSubscriber(): could not create local pubsub subscription: %v", err)
		}
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

	return &PubSubSubscriber{
		Logger:                logger,
		Metrics:               metrics,
		TableInserter:         subscriberTable.Inserter(),
		BatchSize:             subscriberBatchSize,
		ChannelSize:           subscriberChannelSize,
		BigQueryWriteDuration: subscriberWriteDuration,

		pubsubSubscriber: subscriber,
		entryVeto:        entryVeto,
	}, nil
}

func (subscriber *PubSubSubscriber) Receive(ctx context.Context) error {
	err := subscriber.pubsubSubscriber.Receive(ctx, func(ctx context.Context, m *googlepubsub.Message) {
		entries, err := subscriber.UnbatchMessages(m)
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
					subscriber.ackMapMutex.Lock()
					subscriber.ackMap[&unbatchedEntries[i]] = m
					subscriber.ackMapMutex.Unlock()
				}

				// Submit the entry for writing to BigQuery
				if err := subscriber.Submit(context.Background(), &unbatchedEntries[i]); err != nil {
					subscriber.Metrics.ErrorMetrics.QueueFailure.Add(1)
					level.Error(subscriber.Logger).Log("msg", "PubSubSubscriber Receive(): could not submit entry", "err", err)

					// Delete the entry from the map if fail to submit
					if i == len(entries)-1 {
						subscriber.ackMapMutex.Lock()
						delete(subscriber.ackMap, &unbatchedEntries[i])
						subscriber.ackMapMutex.Unlock()
					}

					// Nack the message so it can be redelivered
					m.Nack()
					return
				}
			} else {
				if subscriber.entryVeto {
					m.Ack()
					return
				}

				subscriber.Metrics.ErrorMetrics.ReadFailure.Add(1)
			}
		}
	})

	// If the Receive function returns for any reason, we want to immediately return and exit
	level.Error(subscriber.Logger).Log("msg", "stopped receive loop", "err", err)
	return err
}

func (subscriber *PubSubSubscriber) UnbatchMessages(m *googlepubsub.Message) ([][]byte, error) {
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

func (subscriber *PubSubSubscriber) Submit(ctx context.Context, entry *Entry) error {
	if subscriber.entries == nil {
		subscriber.entries = make(chan *Entry, subscriber.ChannelSize)
	}

	subscriber.bufferMutex.RLock()
	bufferLength := len(subscriber.buffer)
	subscriber.bufferMutex.RUnlock()

	if bufferLength >= subscriber.BatchSize {
		return errors.New("PubSubSubscriber Submit(): entries buffer full")
	}

	select {
	case subscriber.entries <- entry:
		return nil
	default:
		return errors.New("PubSubSubscriber Submit(): entries channel full")
	}
}

func (subscriber *PubSubSubscriber) WriteLoop(ctx context.Context) error {
	if subscriber.entries == nil {
		subscriber.entries = make(chan *Entry, subscriber.ChannelSize)
	}

	messagesToAck := make([]*googlepubsub.Message, subscriber.BatchSize)
	lastWriteTime := time.Now()

	for entry := range subscriber.entries {
		subscriber.Metrics.EntriesQueuedToWrite.Set(float64(len(subscriber.entries)))
		subscriber.bufferMutex.Lock()
		subscriber.buffer = append(subscriber.buffer, entry)
		bufferLength := len(subscriber.buffer)

		// See if any of these entries have a message that needs to be acked
		subscriber.ackMapMutex.RLock()
		message, exists := subscriber.ackMap[entry]
		subscriber.ackMapMutex.RUnlock()
		if exists {
			messagesToAck = append(messagesToAck, message)
			// Delete the entry and message from the ack map
			subscriber.ackMapMutex.Lock()
			delete(subscriber.ackMap, entry)
			subscriber.ackMapMutex.Unlock()
		}

		if bufferLength >= subscriber.BatchSize || time.Since(lastWriteTime) > subscriber.BigQueryWriteDuration {
			if err := subscriber.TableInserter.Put(ctx, subscriber.buffer); err != nil {
				subscriber.bufferMutex.Unlock()
				// Nack all messages
				for _, nackMessage := range messagesToAck {
					nackMessage.Nack()
				}

				// Critical error, log to stdout
				level.Error(subscriber.Logger).Log("msg", "failed to write to BigQuery", "err", err)
				fmt.Printf("PubSubSubscriber WriteLoop(): failed to write to BigQuery: %v\n", err)
				subscriber.Metrics.ErrorMetrics.BigQueryWriteFailure.Add(float64(bufferLength))
			} else {
				// Ack all messages
				for _, ackMessage := range messagesToAck {
					ackMessage.Ack()
				}

				messagesToAck = messagesToAck[:0]
				subscriber.buffer = subscriber.buffer[:0]

				level.Info(subscriber.Logger).Log("msg", "wrote entries to BigQuery", "size", subscriber.BatchSize, "total", bufferLength)
				subscriber.Metrics.EntriesSubmitted.Add(float64(bufferLength))
			}
		}

		subscriber.bufferMutex.Unlock()
		lastWriteTime = time.Now()
	}
	return nil
}
