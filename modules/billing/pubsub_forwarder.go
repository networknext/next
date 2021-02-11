package billing

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
)

type PubSubForwarder struct {
	Biller  Biller
	Logger  log.Logger
	Metrics metrics.ReceiverMetrics

	pubsubSubscription *pubsub.Subscription
}

func NewPubSubForwarder(ctx context.Context, biller Biller, logger log.Logger, metrics metrics.ReceiverMetrics, gcpProjectID string, topicName string, subscriptionName string) (*PubSubForwarder, error) {
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

	return &PubSubForwarder{
		Biller:             biller,
		Logger:             logger,
		Metrics:            metrics,
		pubsubSubscription: pubsubClient.Subscription(subscriptionName),
	}, nil
}

// Forward reads the billing entry from pubsub and writes it to BigQuery
func (psf *PubSubForwarder) Forward(ctx context.Context) {
	err := psf.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		entries, err := psf.unbatchMessages(m)
		if err != nil {
			level.Error(psf.Logger).Log("err", err)
			psf.Metrics.UnmarshalFailure.Add(1)
		}

		psf.Metrics.EntriesReceived.Add(float64(len(entries)))

		billingEntries := make([]BillingEntry, len(entries))
		for i := range billingEntries {
			if ReadBillingEntry(&billingEntries[i], entries[i]) {
				// Starting with version 13 of the billing entry the timestamp is now stored per entry
				// This means we don't want to use pubsub's publish time anymore, unless it's an older
				// version where the timestamp wouldn't be deserialized.
				if billingEntries[i].Timestamp == 0 {
					billingEntries[i].Timestamp = uint64(m.PublishTime.Unix())
				}

				if err := psf.Biller.Bill(context.Background(), &billingEntries[i]); err != nil {
					level.Error(psf.Logger).Log("msg", "could not submit billing entry", "err", err)
					return
				}

				m.Ack()
			} else {
				entryVetoStr := os.Getenv("BILLING_ENTRY_VETO")
				entryVeto, err := strconv.ParseBool(entryVetoStr)

				if err != nil {
					level.Error(psf.Logger).Log("msg", "failed to parse veto env var", "err", err)
					psf.Metrics.UnmarshalFailure.Add(1)
					return
				}

				if entryVeto {
					m.Ack()
					return
				}

				psf.Metrics.UnmarshalFailure.Add(1)
			}
		}
	})

	// If the Receive function returns for any reason, we want to immediately exit and restart the service
	level.Error(psf.Logger).Log("msg", "stopped receive loop", "err", err)
	os.Exit(1)
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
