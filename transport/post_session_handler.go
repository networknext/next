package transport

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/transport/pubsub"
)

type PostSessionHandler struct {
	numGoroutines              int
	postSessionBillingChannel  chan *billing.BillingEntry
	sessionPortalCountsChannel chan *SessionCountData
	sessionPortalDataChannel   chan *SessionPortalData
	portalPublisher            pubsub.Publisher
	portalPublishMaxRetries    int
	biller                     billing.Biller
	logger                     log.Logger
	metrics                    *metrics.PostSessionMetrics
}

func NewPostSessionHandler(numGoroutines int, chanBufferSize int, portalPublisher pubsub.Publisher, portalPublishMaxRetries int,
	biller billing.Biller, logger log.Logger, metrics *metrics.PostSessionMetrics) *PostSessionHandler {
	return &PostSessionHandler{
		numGoroutines:              numGoroutines,
		postSessionBillingChannel:  make(chan *billing.BillingEntry, chanBufferSize),
		sessionPortalCountsChannel: make(chan *SessionCountData, chanBufferSize),
		sessionPortalDataChannel:   make(chan *SessionPortalData, chanBufferSize),
		portalPublisher:            portalPublisher,
		portalPublishMaxRetries:    portalPublishMaxRetries,
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

					portalDataBytes, err := transmitPortalData(post.portalPublisher, pubsub.TopicPortalCruncherSessionCounts, countBytes, post.portalPublishMaxRetries)
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

					portalDataBytes, err := transmitPortalData(post.portalPublisher, pubsub.TopicPortalCruncherSessionData, sessionBytes, post.portalPublishMaxRetries)
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

func (post *PostSessionHandler) BillingBufferSize() uint64 {
	return uint64(len(post.postSessionBillingChannel))
}

func (post *PostSessionHandler) PortalCountBufferSize() uint64 {
	return uint64(len(post.sessionPortalCountsChannel))
}

func (post *PostSessionHandler) PortalDataBufferSize() uint64 {
	return uint64(len(post.sessionPortalDataChannel))
}

func transmitPortalData(publisher pubsub.Publisher, topic pubsub.Topic, data []byte, maxRetries int) (int, error) {
	var byteCount int
	var retryCount int

	for retryCount < maxRetries+1 { // only retry so many times, then error out after that
		singleByteCount, err := publisher.Publish(topic, data)
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
		byteCount += singleByteCount
		break
	}

	if retryCount >= maxRetries {
		return byteCount, errors.New("exceeded retry count on portal data")
	}

	return byteCount, nil

}
