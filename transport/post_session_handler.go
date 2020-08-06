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
	numGoroutines           int
	postSessionChannel      chan *PostSessionData
	portalPublisher         pubsub.Publisher
	portalPublishMaxRetries int
	biller                  billing.Biller
	logger                  log.Logger
	metrics                 *metrics.SessionMetrics
}

func NewPostSessionHandler(numGoroutines int, chanBufferSize int, portalPublisher pubsub.Publisher, portalPublishMaxRetries int,
	biller billing.Biller, logger log.Logger, metrics *metrics.SessionMetrics) *PostSessionHandler {
	return &PostSessionHandler{
		numGoroutines:           numGoroutines,
		postSessionChannel:      make(chan *PostSessionData, chanBufferSize),
		portalPublisher:         portalPublisher,
		portalPublishMaxRetries: portalPublishMaxRetries,
		biller:                  biller,
		logger:                  logger,
		metrics:                 metrics,
	}
}

func (post *PostSessionHandler) StartProcessing(ctx context.Context) {
	for i := 0; i < post.numGoroutines; i++ {
		go func() {
			for {
				select {
				case postSessionData := <-post.postSessionChannel:
					if err := postSessionData.ProcessBillingEntry(post.biller); err != nil {
						level.Error(post.logger).Log("msg", "could not submit billing entry", "err", err)
						post.metrics.ErrorMetrics.BillingFailure.Add(1)
					}

					if portalDataBytes, err := postSessionData.ProcessPortalData(post.portalPublisher, post.portalPublishMaxRetries); err != nil {
						level.Error(post.logger).Log("msg", "could not update portal data", "err", err)
						post.metrics.ErrorMetrics.UpdatePortalFailure.Add(1)
					} else {
						level.Debug(post.logger).Log("msg", fmt.Sprintf("published %d bytes to portal cruncher", portalDataBytes))
					}

					post.metrics.PostSessionEntriesFinished.Add(1)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func (post *PostSessionHandler) Send(postSessionData *PostSessionData) {
	post.postSessionChannel <- postSessionData
}

func (post *PostSessionHandler) QueueSize() uint64 {
	return uint64(len(post.postSessionChannel))
}

type PostSessionData struct {
	PortalData      *SessionPortalData
	PortalCountData *SessionCountData
	BillingEntry    *billing.BillingEntry
}

func (post *PostSessionData) ProcessBillingEntry(biller billing.Biller) error {
	return biller.Bill(context.Background(), post.BillingEntry)
}

func (post *PostSessionData) ProcessPortalData(publisher pubsub.Publisher, maxRetries int) (int, error) {
	sessionBytes, err := post.PortalData.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("could not marshal portal data: %v", err)
	}

	countBytes, err := post.PortalCountData.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("could not marshal portal count data: %v", err)
	}

	var byteCount int

	var retryCount int

	if fmt.Sprintf("%016x", post.PortalData.Meta.BuyerID) != "b8e4f84ca63b2021" {
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
	}

	retryCount = 0
	for retryCount < maxRetries { // only retry so many times, then error out after that
		singleByteCount, err := publisher.Publish(pubsub.TopicPortalCruncherSessionCounts, countBytes)
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
		return byteCount, errors.New("exceeded retry count on session counts")
	}

	return byteCount, nil

}
