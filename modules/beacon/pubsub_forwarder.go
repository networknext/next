package beacon

import (
	"context"
	"fmt"
	"os"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
)

type PubSubForwarder struct {
	Beaconer Beaconer
	Logger   log.Logger
	Metrics  *metrics.BeaconInserterMetrics

	pubsubSubscription *pubsub.Subscription
}

func NewPubSubForwarder(ctx context.Context, beaconer Beaconer, logger log.Logger, metrics *metrics.BeaconInserterMetrics, gcpProjectID string, topicName string, subscriptionName string, numRecvGoroutines int) (*PubSubForwarder, error) {
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
		Beaconer:           beaconer,
		Logger:             logger,
		Metrics:            metrics,
		pubsubSubscription: subscriber,
	}, nil
}

// Forward reads the beacon entry from pubsub and writes it to BigQuery
func (psf *PubSubForwarder) Forward(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	err := psf.pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
		entries, err := psf.unbatchMessages(m)
		if err != nil {
			level.Error(psf.Logger).Log("err", err)
			psf.Metrics.ErrorMetrics.BeaconInserterBatchedReadFailure.Add(1)
		}

		psf.Metrics.EntriesTransfered.Add(float64(len(entries)))

		beaconEntries := make([]transport.NextBeaconPacket, len(entries))
		for i := range beaconEntries {

			if err = transport.ReadBeaconEntry(&beaconEntries[i], entries[i]); err == nil {
				if err := psf.Beaconer.Submit(ctx, &beaconEntries[i]); err != nil {
					level.Error(psf.Logger).Log("msg", "could not submit beacon entry", "err", err)
					// Nack if we failed to submit the beacon entry
					m.Nack()
					return
				}

				m.Ack()
			} else {
				entryVeto, err := envvar.GetBool("BEACON_ENTRY_VETO", false)
				if err != nil {
					level.Error(psf.Logger).Log("msg", "failed to parse veto env var", "err", err)
					psf.Metrics.ErrorMetrics.BeaconInserterReadFailure.Add(1)
					// Nack if we failed to read the beacon entry
					m.Nack()
					return
				}

				if entryVeto {
					m.Ack()
					return
				}

				psf.Metrics.ErrorMetrics.BeaconInserterReadFailure.Add(1)
				// Nack if we failed to read the beacon entry
				m.Nack()
			}
		}
	})

	if err != nil && err != context.Canceled {
		// If the Receive function returns for any reason besides shutdown, we want to immediately exit and restart the service
		level.Error(psf.Logger).Log("msg", "stopped receive loop", "err", err)
		os.Exit(1)
	}

	// Close entries channel to ensure messages are drained for the final write to BigQuery
	psf.Beaconer.Close()
	level.Debug(psf.Logger).Log("msg", "receive canceled, closed entries channel")
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
