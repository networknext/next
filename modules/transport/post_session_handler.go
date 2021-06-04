package transport

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport/pubsub"
	"github.com/networknext/backend/modules/vanity"
)

type PostSessionHandler struct {
	numGoroutines              int
	postSessionBillingChannel  chan *billing.BillingEntry
	sessionPortalCountsChannel chan *SessionCountData
	sessionPortalDataChannel   chan *SessionPortalData
	vanityMetricChannel        chan vanity.VanityMetrics
	portalPublishers           []pubsub.Publisher
	portalPublisherIndex       int
	portalPublishMaxRetries    int
	vanityPublishers           []pubsub.Publisher
	vanityPublisherIndex       int
	vanityPublishMaxRetries    int
	useVanityMetrics           bool
	biller                     billing.Biller
	logger                     log.Logger
	metrics                    *metrics.PostSessionMetrics
}

func NewPostSessionHandler(numGoroutines int, chanBufferSize int, portalPublishers []pubsub.Publisher, portalPublishMaxRetries int,
	vanityPublishers []pubsub.Publisher, vanityPublishMaxRetries int, useVanityMetrics bool, biller billing.Biller, logger log.Logger, metrics *metrics.PostSessionMetrics) *PostSessionHandler {

	return &PostSessionHandler{
		numGoroutines:              numGoroutines,
		postSessionBillingChannel:  make(chan *billing.BillingEntry, chanBufferSize),
		sessionPortalCountsChannel: make(chan *SessionCountData, chanBufferSize),
		sessionPortalDataChannel:   make(chan *SessionPortalData, chanBufferSize),
		vanityMetricChannel:        make(chan vanity.VanityMetrics, chanBufferSize),
		portalPublishers:           portalPublishers,
		portalPublishMaxRetries:    portalPublishMaxRetries,
		vanityPublishers:           vanityPublishers,
		vanityPublishMaxRetries:    vanityPublishMaxRetries,
		useVanityMetrics:           useVanityMetrics,
		biller:                     biller,
		logger:                     logger,
		metrics:                    metrics,
	}
}

func (post *PostSessionHandler) StartProcessing(ctx context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < post.numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case billingEntry := <-post.postSessionBillingChannel:
					if err := post.biller.Bill(ctx, billingEntry); err != nil {
						level.Error(post.logger).Log("msg", "could not submit billing entry", "err", err)
						post.metrics.BillingFailure.Add(1)
						continue
					}

					post.metrics.BillingEntriesFinished.Add(1)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	for i := 0; i < post.numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case postSessionCountData := <-post.sessionPortalCountsChannel:
					countBytes, err := postSessionCountData.MarshalBinary()
					if err != nil {
						level.Error(post.logger).Log("msg", "could not marshal count data", "err", err)
						post.metrics.PortalFailure.Add(1)
						continue
					}

					portalDataBytes, err := post.TransmitPortalData(ctx, pubsub.TopicPortalCruncherSessionCounts, countBytes)
					if err != nil {
						level.Error(post.logger).Log("msg", "could not update portal counts", "err", err)
						post.metrics.PortalFailure.Add(1)
						continue
					}

					level.Debug(post.logger).Log("type", "session counts", "msg", fmt.Sprintf("published %d bytes to portal cruncher", portalDataBytes))
					post.metrics.PortalEntriesFinished.Add(1)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	for i := 0; i < post.numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case postSessionPortalData := <-post.sessionPortalDataChannel:
					sessionBytes, err := postSessionPortalData.MarshalBinary()
					if err != nil {
						level.Error(post.logger).Log("msg", "could not marshal portal data", "err", err)
						post.metrics.PortalFailure.Add(1)
						continue
					}

					portalDataBytes, err := post.TransmitPortalData(ctx, pubsub.TopicPortalCruncherSessionData, sessionBytes)
					if err != nil {
						level.Error(post.logger).Log("msg", "could not update portal data", "err", err)
						post.metrics.PortalFailure.Add(1)
						continue
					}

					level.Debug(post.logger).Log("type", "session data", "msg", fmt.Sprintf("published %d bytes to portal cruncher", portalDataBytes))
					post.metrics.PortalEntriesFinished.Add(1)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	if post.useVanityMetrics {

		for i := 0; i < post.numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for {
					select {
					case extractedMetrics := <-post.vanityMetricChannel:
						// Check if received empty struct (signifies not on Next)
						emptyVanity := vanity.VanityMetrics{}
						if extractedMetrics == emptyVanity {
							// If not on Next, no need to send the metric
							level.Debug(post.logger).Log("type", "vanity metrics", "msg", "billingEntry not on next, not sending vanity metric")
							continue
						}

						// Marshal the metrics
						metricBinary, err := extractedMetrics.MarshalBinary()
						if err != nil {
							level.Error(post.logger).Log("msg", "could not marshal vanity metric", "err", err)
							post.metrics.VanityMarshalFailure.Add(1)
							continue
						}

						// Push the data over ZeroMQ
						metricBytes, err := post.TransmitVanityMetrics(ctx, pubsub.TopicVanityMetricData, metricBinary)
						if err != nil {
							level.Error(post.logger).Log("msg", "could not update vanity metrics", "err", err)
							post.metrics.VanityTransmitFailure.Add(1)
							continue
						}

						level.Debug(post.logger).Log("type", "vanity metrics", "msg", fmt.Sprintf("published %d bytes to vanity metrics", metricBytes))
						post.metrics.VanityMetricsFinished.Add(1)
					case <-ctx.Done():
						return
					}
				}
			}()
		}
	}

	wg.Wait()
}

func (post *PostSessionHandler) SendBillingEntry(billingEntry *billing.BillingEntry) {
	select {
	case post.postSessionBillingChannel <- billingEntry:
		post.metrics.BillingEntriesSent.Add(1)
	default:
		post.metrics.BillingBufferFull.Add(1)
	}

}

func (post *PostSessionHandler) SendPortalCounts(sessionPortalCounts *SessionCountData) {
	select {
	case post.sessionPortalCountsChannel <- sessionPortalCounts:
		post.metrics.PortalEntriesSent.Add(1)
	default:
		post.metrics.PortalBufferFull.Add(1)
	}

}

func (post *PostSessionHandler) SendPortalData(sessionPortalData *SessionPortalData) {
	select {
	case post.sessionPortalDataChannel <- sessionPortalData:
		post.metrics.PortalEntriesSent.Add(1)
	default:
		post.metrics.PortalBufferFull.Add(1)
	}

}

func (post *PostSessionHandler) SendVanityMetric(billingEntry *billing.BillingEntry) {
	select {
	case post.vanityMetricChannel <- post.ExtractVanityMetrics(billingEntry):
		post.metrics.VanityMetricsSent.Add(1)
		level.Info(post.logger).Log("msg", "sent vanity metric")
	default:
		post.metrics.VanityBufferFull.Add(1)
	}

}

func (post *PostSessionHandler) BillingBufferSize() uint64 {
	return uint64(len(post.postSessionBillingChannel))
}

func (post *PostSessionHandler) PortalCountBufferSize() uint64 {
	return uint64(len(post.sessionPortalCountsChannel))
}

func (post *PostSessionHandler) PortalDataBufferSize() uint64 {
	return uint64(len(post.sessionPortalDataChannel))
}

func (post *PostSessionHandler) VanityBufferSize() uint64 {
	return uint64(len(post.vanityMetricChannel))
}

func (post *PostSessionHandler) TransmitPortalData(ctx context.Context, topic pubsub.Topic, data []byte) (int, error) {
	var byteCount int
	var err error

	for i := range post.portalPublishers {
		var retryCount int

		// Calculate the index of the portal publisher to use for this iteration
		index := (post.portalPublisherIndex + i) % len(post.portalPublishers)

		for retryCount < post.portalPublishMaxRetries+1 { // only retry so many times
			byteCount, err = post.portalPublishers[index].Publish(ctx, topic, data)
			if err != nil {
				switch err.(type) {
				case *pubsub.ErrRetry:
					retryCount++
					continue
				default:
					return 0, err
				}
			}

			retryCount = -1
			break
		}

		// We published the message, break out
		if retryCount < post.portalPublishMaxRetries {
			break
		}

		// If we've hit the retry limit, try again using another portal publisher.
		// If this is the last iteration and we still can't publish the message, error out.
		if i == len(post.portalPublishers)-1 {
			return byteCount, errors.New("exceeded retry count on portal data")
		}
	}

	// If we've successfully published the message, increment the portal publisher index
	// so that we evenly distribute the load across each publisher.
	post.portalPublisherIndex = (post.portalPublisherIndex + 1) % len(post.portalPublishers)
	return byteCount, nil

}

func (post *PostSessionHandler) TransmitVanityMetrics(ctx context.Context, topic pubsub.Topic, data []byte) (int, error) {
	var byteCount int
	var err error

	for i := range post.vanityPublishers {
		var retryCount int

		// Calculate the index of the vanity publisher to use for this iteration
		index := (post.vanityPublisherIndex + i) % len(post.vanityPublishers)

		for retryCount < post.vanityPublishMaxRetries+1 { // only retry so many times
			byteCount, err = post.vanityPublishers[index].Publish(ctx, topic, data)
			if err != nil {
				switch err.(type) {
				case *pubsub.ErrRetry:
					retryCount++
					continue
				default:
					return 0, err
				}
			}

			retryCount = -1
			break
		}

		// We published the message, break out
		if retryCount < post.vanityPublishMaxRetries {
			break
		}

		// If we've hit the retry limit, try again using another vanity publisher.
		// If this is the last iteration and we still can't publish the message, error out.
		if i == len(post.vanityPublishers)-1 {
			return byteCount, errors.New("exceeded retry count on vanity metric data")
		}
	}

	// If we've successfully published the message, increment the vanity publisher index
	// so that we evenly distribute the load across each publisher.
	post.vanityPublisherIndex = (post.vanityPublisherIndex + 1) % len(post.vanityPublishers)
	return byteCount, nil
}

func (post *PostSessionHandler) ExtractVanityMetrics(billingEntry *billing.BillingEntry) vanity.VanityMetrics {
	if billingEntry.Next {
		latencyReduced := 0
		if billingEntry.RTTReduction {
			latencyReduced = 1
		}

		packetLossReduced := 0
		if billingEntry.PacketLossReduction {
			packetLossReduced = 1
		}

		jitterReduced := 0
		if billingEntry.DirectJitter-billingEntry.NextJitter > 0 {
			jitterReduced = 1
		}

		vm := vanity.VanityMetrics{
			BuyerID:                 billingEntry.BuyerID,
			UserHash:                billingEntry.UserHash,
			SessionID:               billingEntry.SessionID,
			Timestamp:               billingEntry.Timestamp,
			SlicesAccelerated:       uint64(1),
			SlicesLatencyReduced:    uint64(latencyReduced),
			SlicesPacketLossReduced: uint64(packetLossReduced),
			SlicesJitterReduced:     uint64(jitterReduced),
		}

		return vm
	}

	return vanity.VanityMetrics{}
}
