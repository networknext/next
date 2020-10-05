package transport

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/transport/pubsub"
	"github.com/pebbe/zmq4"
)

type PostSessionHandler struct {
	numGoroutines             int
	postSessionBillingChannel chan *billing.BillingEntry
	sessionPortalDataChannel  chan *SessionPortalData
	portalPublisher           pubsub.Publisher
	portalPublishMaxRetries   int
	biller                    billing.Biller
	logger                    log.Logger
	metrics                   *metrics.PostSessionMetrics

	maxBufferSize int
}

func NewPostSessionHandler(numGoroutines int, chanBufferSize int, portalPublisher pubsub.Publisher, portalPublishMaxRetries int,
	biller billing.Biller, logger log.Logger, metrics *metrics.PostSessionMetrics) *PostSessionHandler {
	return &PostSessionHandler{
		numGoroutines:             numGoroutines,
		postSessionBillingChannel: make(chan *billing.BillingEntry, chanBufferSize),
		sessionPortalDataChannel:  make(chan *SessionPortalData, chanBufferSize),
		portalPublisher:           portalPublisher,
		portalPublishMaxRetries:   portalPublishMaxRetries,
		biller:                    biller,
		logger:                    logger,
		metrics:                   metrics,
		maxBufferSize:             chanBufferSize,
	}
}

func (post *PostSessionHandler) StartProcessing(ctx context.Context) {
	for i := 0; i < post.numGoroutines; i++ {
		go func() {
			for {
				select {
				case billingEntry := <-post.postSessionBillingChannel:
					if err := post.biller.Bill(ctx, billingEntry); err != nil {
						level.Error(post.logger).Log("msg", "could not submit billing entry", "err", err)
						post.metrics.BillingFailure.Add(1)
					}

					post.metrics.BillingEntriesFinished.Add(1)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	for i := 0; i < post.numGoroutines; i++ {
		go func() {
			for {
				select {
				case postSessionPortalData := <-post.sessionPortalDataChannel:
					if portalDataBytes, err := postSessionPortalData.ProcessPortalData(post.portalPublisher, post.portalPublishMaxRetries); err != nil {
						level.Error(post.logger).Log("msg", "could not update portal data", "err", err)
						post.metrics.PortalFailure.Add(1)
					} else {
						level.Debug(post.logger).Log("msg", fmt.Sprintf("published %d bytes to portal cruncher", portalDataBytes))
					}

					post.metrics.PortalEntriesFinished.Add(1)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (post *PostSessionHandler) SendBillingEntry(billingEntry *billing.BillingEntry) {
	post.postSessionBillingChannel <- billingEntry
}

func (post *PostSessionHandler) SendPortalData(sessionPortalData *SessionPortalData) {
	post.sessionPortalDataChannel <- sessionPortalData
}

func (post *PostSessionHandler) BillingBufferSize() uint64 {
	return uint64(len(post.postSessionBillingChannel))
}

func (post *PostSessionHandler) PortalBufferSize() uint64 {
	return uint64(len(post.sessionPortalDataChannel))
}

func (post *PostSessionHandler) IsBillingBufferFull() bool {
	return len(post.postSessionBillingChannel) >= post.maxBufferSize
}

func (post *PostSessionHandler) IsPortalBufferFull() bool {
	return len(post.sessionPortalDataChannel) >= post.maxBufferSize
}

func (data *SessionPortalData) ProcessPortalData(publisher pubsub.Publisher, maxRetries int) (int, error) {
	sessionBytes, err := data.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("could not marshal portal data: %v", err)
	}

	var byteCount int

	var retryCount int

	for retryCount < maxRetries { // only retry so many times, then error out after that
		singleByteCount, err := publisher.Publish(pubsub.TopicPortalCruncherSessionData, sessionBytes)
		if err != nil {
			errno := zmq4.AsErrno(err)
			switch errno {
			case zmq4.AsErrno(syscall.EAGAIN):
				retryCount++
				time.Sleep(time.Millisecond * 100) // If the send queue is backed up, wait a little bit and try again
			default:
				return 0, err
			}
		} else {
			retryCount = -1
			byteCount += singleByteCount
			break
		}
	}

	if retryCount >= maxRetries {
		return byteCount, errors.New("exceeded retry count on portal data")
	}

	return byteCount, nil

}
