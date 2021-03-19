package pubsub

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	googlepubsub "cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
)

type PubSubPublisher struct {
	clients []*GooglePubSubClient
}

type GooglePubSubClient struct {
	PubsubClient         *pubsub.Client
	Topic                *pubsub.Topic
	ResultChan           chan *pubsub.PublishResult
	Logger               log.Logger
	Metrics              *metrics.GooglePublisherMetrics
	BufferCountThreshold int
	MinBufferBytes       int
	CancelContextFunc    context.CancelFunc

	buffer             []byte
	bufferMessageCount int
	bufferMutex        sync.Mutex
}

// Creates a new GooglePubSubPublisher with clientCount clients that will publish to the given topicID in Google Pub/Sub
func NewPubSubPublisher(ctx context.Context, publisherMetrics *metrics.GooglePublisherMetrics, logger log.Logger, gcpProjectID string, topicID string, clientCount int, clientBufferCountThreshold int, clientMinBufferBytes int, settings *pubsub.PublishSettings) (*PubSubPublisher, error) {
	if settings == nil {
		return nil, errors.New("NewPubSubPublisher(): nil google pubsub publish settings")
	}

	clients := make([]*GooglePubSubClient, clientCount)

	for i := 0; i < clientCount; i++ {
		var client *GooglePubSubClient
		var err error
		client = &GooglePubSubClient{}
		client.PubsubClient, err = googlepubsub.NewClient(ctx, gcpProjectID)
		client.Metrics = billingMetrics
		client.Logger = resultLogger
		if err != nil {
			return nil, fmt.Errorf("NewPubSubPublisher(): could not create pubsub client %v: %v", i, err)
		}

		// Create the pubsub topic if running locally with the pubsub emulator
		if gcpProjectID == "local" {
			if _, err := client.PubsubClient.CreateTopic(ctx, topicID); err != nil {
				// Not the best, but the underlying error type is internal so we can't check for it
				if err.Error() != "rpc error: code = AlreadyExists desc = Topic already exists" {
					return nil, fmt.Errorf("NewPubSubPublisher(): %v", err)
				}
			}
		}

		// Define thresholds for publishing
		client.buffer = make([]byte, 0)
		client.BufferCountThreshold = clientBufferCountThreshold
		client.MinBufferBytes = clientMinBufferBytes

		ctx := context.Background()

		// Ensure topic is valid
		client.Topic = client.PubsubClient.Topic(topicID)
		ok, err := client.Topic.Exists(ctx)
		if err != nil {
			return nil, fmt.Errorf("NewPubSubPublisher(): could not verify if topic %s exists: %v", topicID, err)
		}
		if !ok {
			return nil, fmt.Errorf("NewPubSubPublisher(): topic %s does not exist", topicID)
		}

		client.Topic.PublishSettings = *settings
		client.ResultChan = make(chan *googlepubsub.PublishResult)

		// Define a cancel func to stop clients
		cancelCtx, cancelFunc := context.WithCancel(ctx)
		client.CancelContextFunc = cancelFunc

		// Start a goroutine for monitoring publish results
		go client.pubsubResults(cancelCtx)

		clients[i] = client
	}

	publisher := &PubSubPublisher{
		clients: clients,
	}

	return publisher, nil
}

// Batch-writes the given entry when the buffer count and minimum buffer byte thresholds have been met
func (publisher *PubSubPublisher) Publish(ctx context.Context, entry *Entry) error {
	if publisher.clients == nil {
		return fmt.Errorf("GooglePubSubPublisher Publish(): clients not initialized")
	}

	// Assign a client for this entry
	index := uint64(time.Now().Unix()) % uint64(len(publisher.clients))
	client := publisher.clients[index]

	// Get the bytes for the entry
	entryBytes, err := entry.WriteEntry()
	if err != nil {
		return fmt.Errorf("PubSubPublisher Publish(): %v", err)
	}

	// Add an offset for the entry for chaining entries together
	data := make([]byte, 4+len(entryBytes))
	var offset int
	encoding.WriteUint32(data, &offset, uint32(len(entryBytes)))
	encoding.WriteBytes(data, &offset, entryBytes, len(entryBytes))

	client.bufferMutex.Lock()
	defer client.bufferMutex.Unlock()

	// Add this data to the buffer if the message or minimum buffer byte thresholds haven't been met yet
	if client.bufferMessageCount < client.BufferCountThreshold || len(client.buffer) < client.MinBufferBytes {
		client.buffer = append(client.buffer, data...)
		client.bufferMessageCount++
		client.Metrics.EntriesQueuedToPublish.Add(1)
	}

	// Publish the results if we are over the message and minimum buffer byte thresholds
	var result *googlepubsub.PublishResult
	if client.bufferMessageCount >= client.BufferCountThreshold && len(client.buffer) >= client.MinBufferBytes {
		result = client.Topic.Publish(ctx, &pubsub.Message{Data: client.buffer})
		if result != nil {
			client.ResultChan <- result
		}

		client.Metrics.EntriesPublished.Add(float64(client.bufferMessageCount))
		client.buffer = make([]byte, 0)
		client.bufferMessageCount = 0
	}

	return nil
}

// Stops all clients for the PubSubPublisher
func (publisher *PubSubPublisher) Stop() {
	for _, client := range publisher.clients {
		client.CancelContextFunc()
	}
}

// Handles publish results and closing clients when canceled
func (client *GooglePubSubClient) pubsubResults(ctx context.Context) {
	for {
		select {
		case result := <-client.ResultChan:
			_, err := result.Get(ctx)
			if err != nil {
				// Critical error, print to stdout
				level.Error(client.Logger).Log(client.Topic.String(), "failed to publish to Google Pub/Sub", "err", err)
				fmt.Printf("%s failed to publish to Google Pub/Sub: %v\n", client.Topic.String(), err)
				client.Metrics.ErrorMetrics.PublishFailure.Add(1)
			} else {
				level.Debug(client.Logger).Log(client.Topic.String(), "successfully published data to Google Pub/Sub")
			}
		case <-ctx.Done():
			// Close the clients and finish up
			err := client.Close()
			if err != nil {
				level.Error(client.Logger).Log("err", err)
			}
			return
		}
	}
}
