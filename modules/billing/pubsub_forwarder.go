package billing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
)

type PubSubForwarder struct {
	Biller     Biller
	EntryVeto  bool
	MaxRetries int
	RetryTime  time.Duration
	Metrics    *metrics.BillingMetrics

	pubsubSubscription *pubsub.Subscription
}

// PubSubForwarder receives batches of billing entries from Google Pub/Sub, unbatches, and inserts them into an internal channel.
// NOTE: use SEPARATE PubSubForwarders for receiving BillingEntry and BillingEntry2 from different Pub/Sub subscriptions.
func NewPubSubForwarder(ctx context.Context,
	biller Biller,
	entryVeto bool,
	maxRetries int,
	retryTime time.Duration,
	metrics *metrics.BillingMetrics,
	gcpProjectID string,
	topicName string,
	subscriptionName string,
	numRecvGoroutines int,
) (*PubSubForwarder, error) {
	pubsubClient, err := pubsub.NewClient(ctx, gcpProjectID)
	if err != nil {
		return nil, fmt.Errorf("could not create pubsub client: %v", err)
	}

	if gcpProjectID == "local" {
		if _, err := pubsubClient.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
			Topic: pubsubClient.Topic(topicName),
		}); err != nil && err.Error() != "rpc error: code = AlreadyExists desc = Subscription already exists" {
			// Not the best error check, but the underlying error type is internal so we can't check for it
			return nil, fmt.Errorf("could not create local pubsub subscription: %v", err)
		}
	}

	// Set the number goroutines for pulling from Google Pub/Sub
	subscriber := pubsubClient.Subscription(subscriptionName)
	subscriber.ReceiveSettings.NumGoroutines = numRecvGoroutines

	return &PubSubForwarder{
		Biller:             biller,
		EntryVeto:          entryVeto,
		MaxRetries:         maxRetries,
		RetryTime:          retryTime,
		Metrics:            metrics,
		pubsubSubscription: subscriber,
	}, nil
}

// Forward reads the billing entry 2 from pubsub and writes it to BigQuery
func (psf *PubSubForwarder) Forward2(ctx context.Context, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	err := psf.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		entries, err := psf.unbatchMessages(m)
		if err != nil {
			core.Error("failed to unbatch messages: %v", err)
			psf.Metrics.ErrorMetrics.Billing2BatchedReadFailure.Add(1)
		}

		psf.Metrics.Entries2Received.Add(float64(len(entries)))

		billingEntries := make([]BillingEntry2, len(entries))

		for i := range billingEntries {
			if err = ReadBillingEntry2(&billingEntries[i], entries[i]); err == nil {

				// Only retry so many times to submit the entry to the channel
				var retryCount int

				for retryCount < psf.MaxRetries+1 {
					if err := psf.Biller.Bill2(ctx, &billingEntries[i]); err != nil {

						switch err.(type) {
						case *ErrEntries2BufferFull:
							retryCount++
							time.Sleep(psf.RetryTime)
							continue
						case *ErrSummaryEntries2BufferFull:
							retryCount++
							time.Sleep(psf.RetryTime)
							continue
						default:
							// Nack if we failed to submit the billing entry
							core.Error("could not submit billing entry 2: %v", err)
							m.Nack()
							return
						}

					}

					// Submitted the entry, break out
					retryCount -= 1
					break
				}

				if retryCount > psf.MaxRetries {
					// Failed to submit after max retries, nack the message
					core.Error("exceeded max retries (%d), could not submit billing entry 2 (entry %d / %d entries)", psf.MaxRetries, i, len(entries))
					psf.Metrics.ErrorMetrics.Billing2RetryLimitReached.Add(1)
					m.Nack()
					return
				}
			} else {
				core.Error("could not read billing entry 2: %v", err)
				core.Error("bytes for unread entry (%d / %d): %+v", i, len(entries), entries[i])

				if psf.EntryVeto {
					core.Debug("entry veto enabled, acking entry %d (out of %d)", i, len(entries))
					m.Ack()
					return
				}

				psf.Metrics.ErrorMetrics.Billing2ReadFailure.Add(1)
				// Nack if we failed to read the billing entry
				m.Nack()
				return
			}
		}

		// Successfully submit all entries in the message
		m.Ack()
	})

	if err != nil && err != context.Canceled {
		// If the Receive function returns for any reason besides shutdown, we want to immediately exit and restart the service
		core.Error("stopped receive loop: %v", err)
		errChan <- err
	}

	// Close entries channel to ensure messages are drained for the final write to BigQuery
	psf.Biller.Close()
	core.Debug("receive loop canceled, closed entries channel")
}

func (psf *PubSubForwarder) unbatchMessages(m *pubsub.Message) ([][]byte, error) {
	messages := make([][]byte, 0)

	var offset int
	for {
		if offset >= len(m.Data) {
			break
		}

		var messageLength uint32
		var message []byte
		if !encoding.ReadUint32(m.Data, &offset, &messageLength) {
			return nil, fmt.Errorf("failed to read batched message length at offset %d (length %d)", offset, len(m.Data))
		}

		if !encoding.ReadBytes(m.Data, &offset, &message, messageLength) {
			return nil, fmt.Errorf("failed to read batched message bytes at offset %d (length %d)", offset, len(m.Data))
		}

		messages = append(messages, message)
	}

	return messages, nil
}
