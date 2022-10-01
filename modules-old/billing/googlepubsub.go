package billing

import (
    "context"
    "errors"
    "fmt"
    "sync"

    "cloud.google.com/go/pubsub"

    "github.com/networknext/backend/modules/core"
    "github.com/networknext/backend/modules/encoding"

    "github.com/networknext/backend/modules-old/metrics"
)

// GooglePubSubBiller is an implementation of a billing handler that sends billing data to Google Pub/Sub through multiple clients
type GooglePubSubBiller struct {
    clients []*GooglePubSubClient
}

// GooglePubSubClient represents a single client that publishes billing data
type GooglePubSubClient struct {
    PubsubClient         *pubsub.Client
    Topic                *pubsub.Topic
    ResultChan           chan *pubsub.PublishResult
    Metrics              *metrics.BillingMetrics
    BufferCountThreshold int
    MinBufferBytes       int
    CancelContextFunc    context.CancelFunc

    buffer             []byte
    bufferMessageCount int
    bufferMutex        sync.Mutex
}

// NewBiller creates a new GooglePubSubBiller, sets up the pubsub clients, and starts goroutines to listen for publish results.
// To clean up the results goroutine, use ctx.Done().
// NOTE: use SEPARATE GooglePubSubBillers for writing BillingEntry and BillingEntry2 to utilize different pubsub topics
func NewGooglePubSubBiller(ctx context.Context, billingMetrics *metrics.BillingMetrics, projectID string, billingTopicID string, clientCount int, clientBufferCountThreshold int, clientMinBufferBytes int, settings *pubsub.PublishSettings) (Biller, error) {
    if settings == nil {
        return nil, errors.New("nil google pubsub publish settings")
    }

    clients := make([]*GooglePubSubClient, clientCount)

    for i := 0; i < clientCount; i++ {
        var client *GooglePubSubClient
        var err error
        client = &GooglePubSubClient{}
        client.PubsubClient, err = pubsub.NewClient(ctx, projectID)
        client.Metrics = billingMetrics
        if err != nil {
            return nil, fmt.Errorf("could not create pubsub client %v: %v", i, err)
        }

        // Create the billing topic if running locally with the pubsub emulator
        if projectID == "local" {
            if _, err := client.PubsubClient.CreateTopic(ctx, billingTopicID); err != nil {
                // Not the best, but the underlying error type is internal so we can't check for it
                if err.Error() != "rpc error: code = AlreadyExists desc = Topic already exists" {
                    return nil, err
                }
            }
        }

        client.buffer = make([]byte, 0)
        client.BufferCountThreshold = clientBufferCountThreshold
        client.MinBufferBytes = clientMinBufferBytes

        client.Topic = client.PubsubClient.Topic(billingTopicID)
        client.Topic.PublishSettings = *settings
        client.ResultChan = make(chan *pubsub.PublishResult)

        cancelCtx, cancelFunc := context.WithCancel(context.Background())
        client.CancelContextFunc = cancelFunc

        go client.pubsubResults(cancelCtx)

        clients[i] = client
    }

    biller := &GooglePubSubBiller{
        clients: clients,
    }

    return biller, nil
}

func (biller *GooglePubSubBiller) Bill2(ctx context.Context, entry *BillingEntry2) error {
    if biller.clients == nil {
        return fmt.Errorf("billing: clients not initialized")
    }

    index := entry.SessionID % uint64(len(biller.clients))
    client := biller.clients[index]

    entryBytes, err := WriteBillingEntry2(entry)
    if err != nil {
        return err
    }

    client.Metrics.PubsubBillingEntry2Size.Set(float64(len(entryBytes)))

    data := make([]byte, 4+len(entryBytes))
    var offset int
    encoding.WriteUint32(data, &offset, uint32(len(entryBytes)))
    encoding.WriteBytes(data, &offset, entryBytes, len(entryBytes))

    client.bufferMutex.Lock()

    if client.bufferMessageCount < client.BufferCountThreshold || len(client.buffer) < client.MinBufferBytes {
        client.buffer = append(client.buffer, data...)
        client.bufferMessageCount++
        client.Metrics.Entries2Submitted.Add(1)
    }

    var result *pubsub.PublishResult
    if client.bufferMessageCount >= client.BufferCountThreshold && len(client.buffer) >= client.MinBufferBytes {
        result = client.Topic.Publish(ctx, &pubsub.Message{Data: client.buffer})

        client.Metrics.Entries2Flushed.Add(float64(client.bufferMessageCount))

        client.buffer = make([]byte, 0)
        client.bufferMessageCount = 0
    }

    client.bufferMutex.Unlock()

    if result != nil {
        client.ResultChan <- result
    }

    return nil
}

// Force the biller to publish remaining buffer to Google Pub/Sub, even if we haven't met the batch-publishing criteria
func (biller *GooglePubSubBiller) FlushBuffer(ctx context.Context) {
    for _, client := range biller.clients {
        client.bufferMutex.Lock()

        if client.bufferMessageCount > 0 {

            result := client.Topic.Publish(ctx, &pubsub.Message{Data: client.buffer})
            if result != nil {
                client.ResultChan <- result
            }

            client.Metrics.Entries2Flushed.Add(float64(client.bufferMessageCount))

            client.buffer = make([]byte, 0)
            client.bufferMessageCount = 0
        }

        client.bufferMutex.Unlock()
    }
}

func (biller *GooglePubSubBiller) Close() {
    for _, client := range biller.clients {
        client.Topic.Stop()
        client.CancelContextFunc()
    }
}

func (client *GooglePubSubClient) pubsubResults(ctx context.Context) {
    for {
        select {
        case result := <-client.ResultChan:
            _, err := result.Get(ctx)
            if err != nil {
                core.Error("failed to publish to pubsub: %v", err)
                client.Metrics.ErrorMetrics.Billing2PublishFailure.Add(1)
            }
        case <-ctx.Done():
            return
        }
    }
}
