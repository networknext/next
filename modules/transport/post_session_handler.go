package transport

import (
	"context"
	"errors"
	"sync"

	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/core"
	md "github.com/networknext/backend/modules/match_data"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport/pubsub"
)

type PostSessionHandler struct {
	numGoroutines              int
	postSessionBilling2Channel chan *billing.BillingEntry2
	sessionPortalCountsChannel chan *SessionCountData
	sessionPortalDataChannel   chan *SessionPortalData
	matchDataChannel           chan *md.MatchDataEntry
	portalPublishers           []pubsub.Publisher
	portalPublisherIndex       int
	portalPublishMaxRetries    int
	biller2                    billing.Biller
	featureBilling2            bool
	matcher                    md.Matcher
	metrics                    *metrics.PostSessionMetrics
}

func NewPostSessionHandler(
	numGoroutines int, chanBufferSize int, portalPublishers []pubsub.Publisher, portalPublishMaxRetries int,
	biller2 billing.Biller, featureBilling2 bool, matcher md.Matcher, metrics *metrics.PostSessionMetrics,
) *PostSessionHandler {

	return &PostSessionHandler{
		numGoroutines:              numGoroutines,
		postSessionBilling2Channel: make(chan *billing.BillingEntry2, chanBufferSize),
		sessionPortalCountsChannel: make(chan *SessionCountData, chanBufferSize),
		sessionPortalDataChannel:   make(chan *SessionPortalData, chanBufferSize),
		matchDataChannel:           make(chan *md.MatchDataEntry, chanBufferSize),
		portalPublishers:           portalPublishers,
		portalPublishMaxRetries:    portalPublishMaxRetries,
		biller2:                    biller2,
		featureBilling2:            featureBilling2,
		matcher:                    matcher,
		metrics:                    metrics,
	}
}

func (post *PostSessionHandler) StartProcessing(ctx context.Context, wg *sync.WaitGroup) {

	if post.featureBilling2 {

		for i := 0; i < post.numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				for {
					select {
					case billingEntry := <-post.postSessionBilling2Channel:
						if err := post.biller2.Bill2(ctx, billingEntry); err != nil {
							core.Error("could not submit billing entry 2: %v", err)
							post.metrics.Billing2Failure.Add(1)
							continue
						}

						post.metrics.BillingEntries2Finished.Add(1)
					case <-ctx.Done():
						post.biller2.FlushBuffer(context.Background())
						post.biller2.Close()
						return
					}
				}
			}()
		}
	}

	for i := 0; i < post.numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case postSessionCountData := <-post.sessionPortalCountsChannel:
					countBytes, err := WriteSessionCountData(postSessionCountData)
					if err != nil {
						core.Error("could not serialize count data: %v", err)
						post.metrics.PortalFailure.Add(1)
						continue
					}

					_, err = post.TransmitPortalData(ctx, pubsub.TopicPortalCruncherSessionCounts, countBytes)
					if err != nil {
						core.Error("could not update portal counts: %v", err)
						post.metrics.PortalFailure.Add(1)
						continue
					}

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
					sessionBytes, err := WriteSessionPortalData(postSessionPortalData)
					if err != nil {
						core.Error("could not serialize portal data: %v", err)
						post.metrics.PortalFailure.Add(1)
						continue
					}

					_, err = post.TransmitPortalData(ctx, pubsub.TopicPortalCruncherSessionData, sessionBytes)
					if err != nil {
						core.Error("could not update portal data: %v", err)
						post.metrics.PortalFailure.Add(1)
						continue
					}

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
				case matchData := <-post.matchDataChannel:
					if err := post.matcher.Match(ctx, matchData); err != nil {
						core.Error("could not submit match data entry: %v", err)
						post.metrics.MatchDataEntriesFailure.Add(1)
						continue
					}

					post.metrics.MatchDataEntriesFinished.Add(1)
				case <-ctx.Done():
					post.matcher.FlushBuffer(context.Background())
					post.matcher.Close()
					return
				}
			}
		}()
	}
}

func (post *PostSessionHandler) SendBillingEntry2(billingEntry *billing.BillingEntry2) {
	select {
	case post.postSessionBilling2Channel <- billingEntry:
		post.metrics.BillingEntries2Sent.Add(1)
	default:
		post.metrics.Billing2BufferFull.Add(1)
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

func (post *PostSessionHandler) SendMatchData(matchData *md.MatchDataEntry) {
	select {
	case post.matchDataChannel <- matchData:
		post.metrics.MatchDataEntriesSent.Add(1)
	default:
		post.metrics.MatchDataEntriesBufferFull.Add(1)
	}

}

func (post *PostSessionHandler) Billing2BufferSize() uint64 {
	return uint64(len(post.postSessionBilling2Channel))
}

func (post *PostSessionHandler) PortalCountBufferSize() uint64 {
	return uint64(len(post.sessionPortalCountsChannel))
}

func (post *PostSessionHandler) PortalDataBufferSize() uint64 {
	return uint64(len(post.sessionPortalDataChannel))
}

func (post *PostSessionHandler) MatchDataBufferSize() uint64 {
	return uint64(len(post.matchDataChannel))
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
