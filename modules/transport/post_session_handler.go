package transport

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport/pubsub"
	"github.com/networknext/backend/modules/vanity"
)

type UserSession struct {
	sessionID	uint64
	timestamp	int64
}

type UserSessionMap struct {
	m 		map[uint64]*UserSession
	lock 	sync.RWMutex
}

type PostSessionHandler struct {
	numGoroutines              int
	postSessionBillingChannel  chan *billing.BillingEntry
	sessionPortalCountsChannel chan *SessionCountData
	sessionPortalDataChannel   chan *SessionPortalData
	vanityMetricChannel		   chan *billing.BillingEntry
	portalPublishers           []pubsub.Publisher
	portalPublisherIndex       int
	portalPublishMaxRetries    int
	vanityPublishers		   []pubsub.Publisher
	vanityPublisherIndex	   int
	vanityPublishMaxRetries	   int
	vanityAggregator	   	   map[uint64]vanity.VanityMetrics
	vanityAggregatorMutex	   sync.RWMutex
	vanityUserSessionMap	   *UserSessionMap
	vanityPushDuration		   time.Duration
	useVanityMetrics 		   bool
	biller                     billing.Biller
	logger                     log.Logger
	metrics                    *metrics.PostSessionMetrics
}

func NewPostSessionHandler(numGoroutines int, chanBufferSize int, portalPublishers []pubsub.Publisher, portalPublishMaxRetries int,
	vanityPublishers []pubsub.Publisher, vanityPublishMaxRetries int, vanityPushDuration time.Duration, vanityMaxUserIdleTime time.Duration, 
	vanityExpirationFrequencyCheck time.Duration, useVanityMetrics bool, biller billing.Biller, logger log.Logger, metrics *metrics.PostSessionMetrics) *PostSessionHandler {
	
	return &PostSessionHandler{
		numGoroutines:              numGoroutines,
		postSessionBillingChannel:  make(chan *billing.BillingEntry, chanBufferSize),
		sessionPortalCountsChannel: make(chan *SessionCountData, chanBufferSize),
		sessionPortalDataChannel:   make(chan *SessionPortalData, chanBufferSize),
		vanityMetricChannel:		make(chan *billing.BillingEntry, chanBufferSize),
		portalPublishers:           portalPublishers,
		portalPublishMaxRetries:    portalPublishMaxRetries,
		vanityPublishers:			vanityPublishers,
		vanityPublishMaxRetries:	vanityPublishMaxRetries,
		vanityAggregator:	   	    make(map[uint64]vanity.VanityMetrics),
		vanityAggregatorMutex:		sync.RWMutex{},
		vanityUserSessionMap:		NewUserSessionMap(vanityMaxUserIdleTime, vanityExpirationFrequencyCheck),
		vanityPushDuration:		   	vanityPushDuration,
		useVanityMetrics:			useVanityMetrics,
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
				pushTime := time.Now()

				for {
					select {
					case billingEntry := <-post.vanityMetricChannel:
						// Check if buyer ID is in map
						post.vanityAggregatorMutex.RLock()
						_, ok := post.vanityAggregator[billingEntry.BuyerID]
						post.vanityAggregatorMutex.RUnlock()
						if !ok {
						    // Create a vanity metric for this buyer ID
						    post.vanityAggregatorMutex.Lock()
						    post.vanityAggregator[billingEntry.BuyerID] = vanity.VanityMetrics{BuyerID: billingEntry.BuyerID}
						    post.vanityAggregatorMutex.Unlock()
						}

						// Aggregate elements of the billing entry to respective vanity metric per buyer
						err := post.AggregateVanityMetrics(billingEntry)
						if err != nil {
							level.Error(post.logger).Log("err", err)
							post.metrics.VanityFailure.Add(1)
							continue
						}

						// Early out if not enough time has passed since the last push to ZeroMQ
						if time.Since(pushTime) < post.vanityPushDuration {
							continue
						}
						pushTime = time.Now()

						// Marshal the aggregate vanity metric per buyer
						var aggregateBinary [][]byte
						post.vanityAggregatorMutex.RLock()
						for _, metric := range post.vanityAggregator {
							metricBinary, err := metric.MarshalBinary()
							if err != nil {
								level.Error(post.logger).Log("msg", "could not marshal vanity metric", "err", err)
								post.metrics.VanityFailure.Add(1)
							} else {
								aggregateBinary = append(aggregateBinary, metricBinary)
							}
						}
						post.vanityAggregatorMutex.RUnlock()

						// Push the data over ZeroMQ
						aggregateBytes, err := post.TransmitVanityMetrics(ctx, pubsub.TopicVanityMetricData, aggregateBinary)
						if err != nil {
							level.Error(post.logger).Log("msg", "could not update vanity metrics", "err", err)
							post.metrics.VanityFailure.Add(1)
							continue
						}

						// Reset the map for the buyer IDs
						post.vanityAggregatorMutex.Lock()
						for buyerID, _ := range post.vanityAggregator {
							post.vanityAggregator[buyerID] = vanity.VanityMetrics{BuyerID: buyerID}
						}
						post.vanityAggregatorMutex.Unlock()

						level.Debug(post.logger).Log("type", "vanity metrics", "msg", fmt.Sprintf("published %d bytes to vanity metrics", aggregateBytes))
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
	case post.vanityMetricChannel <- billingEntry:
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

func (post *PostSessionHandler) TransmitVanityMetrics(ctx context.Context, topic pubsub.Topic, data [][]byte) (int, error) {
	var byteCount int
	
	for i := range post.vanityPublishers {
		var retryCount int

		// Calculate the index of the vanity publisher to use for this iteration
		index := (post.vanityPublisherIndex + i) % len(post.vanityPublishers)
		for _, subData := range data {
			for retryCount < post.vanityPublishMaxRetries+1 { // only retry so many times
			
				subByteCount, err := post.vanityPublishers[index].Publish(ctx, topic, subData)
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
				byteCount += subByteCount
				// Finished publishing this subdata, break out
				break
			}
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

// Aggregates vanity metrics per buyer
// Each billing entry is after a slice
func (post *PostSessionHandler) AggregateVanityMetrics(billingEntry *billing.BillingEntry) error {
	post.vanityAggregatorMutex.RLock()
	vanityForBuyer, ok := post.vanityAggregator[billingEntry.BuyerID]
	if !ok {
		post.vanityAggregatorMutex.RUnlock()
		errStr := fmt.Sprintf("Buyer ID %d does not exist in vanityAggregator map", billingEntry.BuyerID)
		return errors.New(errStr)
	}
	post.vanityAggregatorMutex.RUnlock()

	slicesAccelerated := 0
	if billingEntry.Next {
		slicesAccelerated = 1
	}

	latencyReduced := 0
	if billingEntry.Next && billingEntry.RTTReduction {
		latencyReduced = 1
	}

	packetLossReduced := 0
	if billingEntry.Next && billingEntry.PacketLossReduction {
		packetLossReduced = 1
	}
	
	// Jitter is continuous so safe to determine jitter reduction using values
	jitterReduced := 0
	if billingEntry.Next && billingEntry.NextJitter - billingEntry.DirectJitter < 0 {
		jitterReduced = 1
	}

	sessionsAccelerated := 0
	if billingEntry.Next {
		// There is only at most one active session per userHash,
		// so if the latest session is different than the last one, then increment the counter
		if sessionID, ok := post.vanityUserSessionMap.Get(billingEntry.UserHash); ok {
			// Verify if the session is different
			if sessionID != billingEntry.SessionID {
				// Update the map with the latest session
				post.vanityUserSessionMap.Put(billingEntry.UserHash, billingEntry.SessionID)
				sessionsAccelerated = 1
			}
		} else {
			// First session for this userHash
			post.vanityUserSessionMap.Put(billingEntry.UserHash, billingEntry.SessionID)
			sessionsAccelerated = 1
		}
	}

	atomic.AddUint64(&vanityForBuyer.SlicesAccelerated, uint64(slicesAccelerated))
	atomic.AddUint64(&vanityForBuyer.SlicesLatencyReduced, uint64(latencyReduced))
	atomic.AddUint64(&vanityForBuyer.SlicesPacketLossReduced, uint64(packetLossReduced))
	atomic.AddUint64(&vanityForBuyer.SlicesJitterReduced, uint64(jitterReduced))
	atomic.AddUint64(&vanityForBuyer.SessionsAccelerated, uint64(sessionsAccelerated))

	return nil
}


// Creates a new UserSessionMap, which maps a userHash to a userID
// Keys expire after they have not been accessed for maxUserIdleTime
// Key expiration is checked at every expirationFrequencyCheckInterval
func NewUserSessionMap(maxUserIdleTime time.Duration, expirationFrequencyCheck time.Duration) *UserSessionMap {
	m := &UserSessionMap{
		m: make(map[uint64]*UserSession), 
		lock: sync.RWMutex{},
	}
	go func() {
		for _ = range time.Tick(expirationFrequencyCheck) {
			m.lock.Lock()
			for key, session := range m.m {
				if time.Since(time.Unix(session.timestamp, 0)) > maxUserIdleTime {
					delete(m.m, key)
				}
			}
			m.lock.Unlock()
		}
	}()

	return m
}

// Gets the length of the UserSessionMap
func (m *UserSessionMap) Len() int {
	return len(m.m)
}

// Inserts a (userHash, sessionID) pair in the UserSessionMap
// and refreshes the last accessed timestamp
func (m *UserSessionMap) Put(userHash uint64, sessionID uint64) {
	m.lock.RLock()
	session, ok := m.m[userHash]
	m.lock.RUnlock()
	if !ok {
		session = &UserSession{sessionID: sessionID}
		m.lock.Lock()
		m.m[userHash] = session
		m.lock.Unlock()
	}
	
	m.lock.Lock()
	session.timestamp = time.Now().Unix()
	m.lock.Unlock()
}

// Gets the sessionID for a userHash from the UserSessionmap
// and refreshes the last accessed timestamp
func (m *UserSessionMap) Get(userHash uint64) (uint64, bool) {
	m.lock.Lock()
	if session, ok := m.m[userHash]; ok {
		sessionID := session.sessionID
		session.timestamp = time.Now().Unix()
		m.lock.Unlock()

		return sessionID, true
	}
	m.lock.Unlock()
	return 0, false
}