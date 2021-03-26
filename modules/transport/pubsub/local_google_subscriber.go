package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	googlepubsub "cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
)

const (
	DefaultBatchSize   = 10
	DefaultChannelSize = 1000
)

// LocalPubSubSubscriber is a local version of PubSubSubscriber that writes to a local json file instead of BigQuery.
type LocalPubSubSubscriber struct {
	Logger             log.Logger
	Metrics            *metrics.GoogleSubscriberMetrics
	PubSubSubscription *PubSubSubscriber

	jsonFileName string
}

func NewLocalPubSubSubscriber(
	ctx context.Context,
	logger log.Logger,
	metrics *metrics.GoogleSubscriberMetrics,
	gcpProjectID string,
	topicID string,
	subscriptionID string,
	recvNumGoroutines int,
	subscriberBatchSize int,
	subscriberChannelSize int,
	subscriberWriteDuration time.Duration,
	entryVeto bool,
	jsonFileName string,
) (*LocalPubSubSubscriber, error) {
	pubsubClient, err := googlepubsub.NewClient(ctx, gcpProjectID)
	if err != nil {
		return nil, fmt.Errorf("NewLocalPubSubSubscriber(): could not create pubsub client: %v", err)
	}

	var subscriber *googlepubsub.Subscription
	// Extra check to make sure we are running locally
	if gcpProjectID == "local" {
		// Verify the subscriptionID exists
		subscriber = pubsubClient.Subscription(subscriptionID)
		ok, err := subscriber.Exists(ctx)
		if err != nil {
			return nil, fmt.Errorf("NewPubSubSubscriber(): could not verify if subscription %s exists", subscriptionID)
		}
		if !ok {
			// Create the subscription for the local environment
			if _, err := pubsubClient.CreateSubscription(ctx, subscriptionID, googlepubsub.SubscriptionConfig{
				Topic: pubsubClient.Topic(topicID),
			}); err != nil {
				// Not the best error check, but the underlying error type is internal so we can't check for it
				return nil, fmt.Errorf("NewPubSubSubscriber(): could not create local pubsub subscription %s: %v", subscriptionID, err)
			}
		}
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

	// Force batch size to default size if invalid batch size provided
	if subscriberBatchSize < 1 {
		subscriberBatchSize = DefaultBatchSize
	}

	// Force channel size to default size if provided size is less than batch size
	if subscriberChannelSize < subscriberBatchSize {
		subscriberChannelSize = DefaultChannelSize
	}

	// Check if json file name has .json extension
	if !strings.HasSuffix(jsonFileName, ".json") {
		// Append .json extension
		jsonFileName = jsonFileName + ".json"
	}

	// Create a PubSubSubscriber to use its ReceiveAndSubmit() method
	googlePubSubSubscriber := &PubSubSubscriber{
		Logger:             logger,
		Metrics:            metrics,
		TableInserter:      nil,
		BatchSize:          subscriberBatchSize,
		ChannelSize:        subscriberChannelSize,
		WriteDuration:      subscriberWriteDuration,
		PubSubSubscription: subscriber,
		EntryVeto:          entryVeto,
		AckMap:             make(map[*Entry]*googlepubsub.Message),
	}

	return &LocalPubSubSubscriber{
		Logger:             logger,
		Metrics:            metrics,
		PubSubSubscription: googlePubSubSubscriber,

		jsonFileName: jsonFileName,
	}, nil
}

// Uses PubSubSubscriber's ReceiveAndSubmit()
func (subscriber *LocalPubSubSubscriber) ReceiveAndSubmit(ctx context.Context) error {
	err := subscriber.PubSubSubscription.ReceiveAndSubmit(ctx)

	level.Error(subscriber.Logger).Log("err", err)
	return err
}

// WriteLoop ranges over the incoming channel of Entry types and fills an internal buffer.
// Once the buffer fills to the BatchSize it will write all entries to a local JSON File.
// This should only be called from 1 goroutine to avoid using a mutex around the internal buffer slice.
func (subscriber *LocalPubSubSubscriber) WriteLoop(ctx context.Context) error {
	if subscriber.PubSubSubscription.entries == nil {
		subscriber.PubSubSubscription.entries = make(chan *Entry, subscriber.PubSubSubscription.ChannelSize)
	}

	// Track which messages to ack once they have been written successfully
	messagesToAck := make([]*googlepubsub.Message, subscriber.PubSubSubscription.BatchSize)
	lastWriteTime := time.Now()

	for entry := range subscriber.PubSubSubscription.entries {
		subscriber.Metrics.EntriesQueuedToWrite.Set(float64(len(subscriber.PubSubSubscription.entries)))
		subscriber.PubSubSubscription.bufferMutex.Lock()
		subscriber.PubSubSubscription.buffer = append(subscriber.PubSubSubscription.buffer, entry)
		bufferLength := len(subscriber.PubSubSubscription.buffer)

		// See if any of these entries have a message that needs to be acked
		subscriber.PubSubSubscription.AckMapMutex.RLock()
		message, exists := subscriber.PubSubSubscription.AckMap[entry]
		subscriber.PubSubSubscription.AckMapMutex.RUnlock()
		if exists {
			messagesToAck = append(messagesToAck, message)
			// Delete the entry and message from the ack map
			subscriber.PubSubSubscription.AckMapMutex.Lock()
			delete(subscriber.PubSubSubscription.AckMap, entry)
			subscriber.PubSubSubscription.AckMapMutex.Unlock()
		}

		if bufferLength >= subscriber.PubSubSubscription.BatchSize || time.Since(lastWriteTime) > subscriber.PubSubSubscription.WriteDuration {
			if err := subscriber.writeToJSON(); err != nil {
				subscriber.PubSubSubscription.bufferMutex.Unlock()
				// Nack all messages
				for _, nackMessage := range messagesToAck {
					nackMessage.Nack()
				}

				// Critical error, log to stdout
				level.Error(subscriber.Logger).Log("msg", "failed to write to JSON file", "err", err)
				fmt.Printf("LocalPubSubSubscriber WriteLoop(): failed to write to jsonFileName %s: %v\n", subscriber.jsonFileName, err)
				subscriber.Metrics.ErrorMetrics.WriteFailure.Add(float64(bufferLength))
			} else {
				// Ack all messages
				for _, ackMessage := range messagesToAck {
					ackMessage.Ack()
				}

				// Clear the buffers
				messagesToAck = messagesToAck[:0]
				subscriber.PubSubSubscription.buffer = subscriber.PubSubSubscription.buffer[:0]

				level.Info(subscriber.Logger).Log("msg", "wrote entries to JSON file", "size", subscriber.PubSubSubscription.BatchSize, "total", bufferLength)
				subscriber.Metrics.EntriesSubmitted.Add(float64(bufferLength))
			}
		}

		subscriber.PubSubSubscription.bufferMutex.Unlock()
		lastWriteTime = time.Now()
	}

	return nil
}

// Writes the buffer to the local JSON file
func (subscriber *LocalPubSubSubscriber) writeToJSON() error {
	var dataList []map[string]bigquery.Value

	// Append to file if it exists
	if _, err := os.Stat(subscriber.jsonFileName); err == nil {
		// Read from the existing file
		prevData, err := ioutil.ReadFile(subscriber.jsonFileName)
		if err != nil {
			return fmt.Errorf("LocalPubSubSubscriber writeToJSON(): could not read from jsonFileName %s: %v", subscriber.jsonFileName, err)
		}

		// Unmarshal the data
		if err := json.Unmarshal(prevData, &dataList); err != nil {
			return fmt.Errorf("LocalPubSubSubscriber writeToJSON(): could not unmarshal previous data from %s: %v", subscriber.jsonFileName, err)
		}

		// Add the new data
		for _, entry := range subscriber.PubSubSubscription.buffer {
			// Save the data to ensure BigQuery can accept the data
			entryMap, _, err := (*entry).Save()
			if err != nil {
				return fmt.Errorf("Could not Save() entry %v: %v", entry, err)
			}

			// Append the new data to previous data
			dataList = append(dataList, entryMap)
		}

		// Marshal all the data
		updatedData, err := json.MarshalIndent(dataList, "", "\t")
		if err != nil {
			return fmt.Errorf("LocalPubSubSubscriber writeToJSON(): could not marshal updated data: %v", err)
		}

		// Overwrite the file with the updated data
		if err = ioutil.WriteFile(subscriber.jsonFileName, updatedData, 0644); err != nil {
			return fmt.Errorf("LocalPubSubSubscriber writeToJSON(): could not overwrite jsonFileName %s with updated data: %v", subscriber.jsonFileName, err)
		}
	} else if os.IsNotExist(err) {
		// File does not exist, create it
		f, err := os.OpenFile(subscriber.jsonFileName, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("LocalPubSubSubscriber writeToJSON(): could not create jsonFileName %s: %v", subscriber.jsonFileName, err)
		}

		// Close the file with a defer func
		defer func() {
			_ = f.Close()
		}()

		// Add the data to the slice
		for _, entry := range subscriber.PubSubSubscription.buffer {
			// Save the data to ensure BigQuery can accept the data
			entryMap, _, err := (*entry).Save()
			if err != nil {
				return fmt.Errorf("Could not Save() entry %v: %v", entry, err)
			}

			// Append the data
			dataList = append(dataList, entryMap)
		}

		// Marshal the data
		newData, err := json.MarshalIndent(dataList, "", "\t")
		if err != nil {
			return fmt.Errorf("LocalPubSubSubscriber writeToJSON(): could not marshal new data: %v", err)
		}

		// Write the data
		if _, err := f.Write(newData); err != nil {
			fmt.Errorf("LocalPubSubSubscriber writeToJSON(): could not write to jsonFileName %s with new data: %v", subscriber.jsonFileName, err)
		}
	} else {
		// Could not confirm if file exists
		return fmt.Errorf("LocalPubSubSubscriber writeToJSON(): could not confirm if jsonFileName %s exists: %v", subscriber.jsonFileName, err)
	}

	return nil
}
