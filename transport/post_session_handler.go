package transport

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/transport/pubsub"
)

type PostSessionHandler struct {
	numGoroutines      int
	postSessionChannel chan *PostSessionData
	portalPublisher    pubsub.Publisher
	biller             billing.Biller
	logger             log.Logger
	metrics            *metrics.SessionErrorMetrics
}

func NewPostSessionHandler(numGoroutines int, chanBufferSize int, portalPublisher pubsub.Publisher, biller billing.Biller, logger log.Logger, metrics *metrics.SessionErrorMetrics) *PostSessionHandler {
	return &PostSessionHandler{
		numGoroutines:      numGoroutines,
		postSessionChannel: make(chan *PostSessionData, chanBufferSize),
		portalPublisher:    portalPublisher,
		biller:             biller,
		logger:             logger,
		metrics:            metrics,
	}
}

func (post *PostSessionHandler) StartProcessing(ctx context.Context) {
	for i := 0; i < post.numGoroutines; i++ {
		go func() {
			for {
				select {
				case postSessionData := <-post.postSessionChannel:
					go func() {
						fmt.Println("calling process portal")
						if portalDataBytes, err := postSessionData.ProcessPortalData(post.portalPublisher); err != nil {
							level.Error(post.logger).Log("msg", "could not update portal data", "err", err)
							post.metrics.UpdatePortalFailure.Add(1)
						} else {
							level.Debug(post.logger).Log("msg", fmt.Sprintf("published %d bytes to portal cruncher", portalDataBytes))
						}
					}()
					go func() {

						fmt.Println("calling process billing")
						if err := postSessionData.ProcessBillingEntry(post.biller); err != nil {
							level.Error(post.logger).Log("msg", "could not submit billing entry", "err", err)
							post.metrics.BillingFailure.Add(1)
						}
					}()
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

type PostSessionData struct {
	PortalData      *SessionPortalData
	PortalCountData *SessionCountData
	BillingEntry    *billing.BillingEntry
}

func (post *PostSessionData) ProcessBillingEntry(biller billing.Biller) error {
	return biller.Bill(context.Background(), post.BillingEntry)
}

func (post *PostSessionData) ProcessPortalData(publisher pubsub.Publisher) (int, error) {
	sessionBytes, err := post.PortalData.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("could not marshal portal data: %v", err)
	}

	countBytes, err := post.PortalCountData.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("could not marshal portal count data: %v", err)
	}

	var byteCount int
	singleByteCount, err := publisher.Publish(pubsub.TopicPortalCruncherSessionData, sessionBytes)
	byteCount += singleByteCount
	singleByteCount, err = publisher.Publish(pubsub.TopicPortalCruncherSessionCounts, countBytes)
	byteCount += singleByteCount

	if err != nil {
		return 0, fmt.Errorf("could not update portal data: %v", err)

	}

	return byteCount, nil

}
