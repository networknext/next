package transport_test

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/networknext/backend/modules/messages"

	"github.com/networknext/backend/modules-old/billing"
	md "github.com/networknext/backend/modules-old/match_data"
	"github.com/networknext/backend/modules-old/metrics"
	"github.com/networknext/backend/modules-old/routing"
	"github.com/networknext/backend/modules-old/transport"
	"github.com/networknext/backend/modules-old/transport/pubsub"

	"github.com/stretchr/testify/assert"
)

type badBiller struct {
	calledChan2 chan bool
}

func (biller *badBiller) Bill2(ctx context.Context, billingEntry *billing.BillingEntry2) error {
	biller.calledChan2 <- true
	return errors.New("bad bill")
}

func (biller *badBiller) FlushBuffer(ctx context.Context) {}

func (biller *badBiller) Close() {}

type mockBiller struct {
	calledChan2    chan bool
	billedEntries2 []billing.BillingEntry2
}

func (biller *mockBiller) Bill2(ctx context.Context, billingEntry *billing.BillingEntry2) error {
	biller.billedEntries2 = append(biller.billedEntries2, *billingEntry)

	biller.calledChan2 <- true
	return nil
}

func (biller *mockBiller) FlushBuffer(ctx context.Context) {}

func (biller *mockBiller) Close() {}

type badMatcher struct {
	calledChan chan bool
}

func (matcher *badMatcher) Match(ctx context.Context, matchDataEntry *md.MatchDataEntry) error {
	matcher.calledChan <- true
	return errors.New("bad match")
}

func (matcher *badMatcher) FlushBuffer(ctx context.Context) {}

func (matcher *badMatcher) Close() {}

type mockMatcher struct {
	calledChan     chan bool
	matchedEntries []md.MatchDataEntry
}

func (matcher *mockMatcher) Match(ctx context.Context, matchDataEntry *md.MatchDataEntry) error {
	matcher.matchedEntries = append(matcher.matchedEntries, *matchDataEntry)

	matcher.calledChan <- true
	return nil
}

func (matcher *mockMatcher) FlushBuffer(ctx context.Context) {}

func (matcher *mockMatcher) Close() {}

type badPublisher struct {
	calledChan chan bool
}

func (pub *badPublisher) Publish(ctx context.Context, topic pubsub.Topic, message []byte) (int, error) {
	pub.calledChan <- true
	return 0, errors.New("bad publish")
}

func (pub *badPublisher) Close() error {
	return nil
}

type retryPublisher struct {
	retryCount int
	retries    int
}

func (pub *retryPublisher) Publish(ctx context.Context, topic pubsub.Topic, message []byte) (int, error) {
	if pub.retries >= pub.retryCount {
		return len(message), nil
	}

	pub.retries++
	return 0, &pubsub.ErrRetry{}
}

func (pub *retryPublisher) Close() error {
	return nil
}

type mockPublisher struct {
	calledChan        chan bool
	publishedMessages [][][]byte
}

func (pub *mockPublisher) Publish(ctx context.Context, topic pubsub.Topic, message []byte) (int, error) {
	pub.publishedMessages = append(pub.publishedMessages, [][]byte{{byte(topic)}, message})

	pub.calledChan <- true
	return len(message), nil
}

func (pub *mockPublisher) Close() error {
	return nil
}

func testBillingEntry2() *billing.BillingEntry2 {
	return &billing.BillingEntry2{
		Version:                         uint32(0),
		Timestamp:                       rand.Uint32(),
		SessionID:                       rand.Uint64(),
		SliceNumber:                     rand.Uint32(),
		DirectMinRTT:                    rand.Int31(),
		DirectMaxRTT:                    rand.Int31(),
		DirectPrimeRTT:                  rand.Int31(),
		DirectJitter:                    rand.Int31(),
		DirectPacketLoss:                rand.Int31(),
		RealPacketLoss:                  rand.Int31(),
		RealPacketLoss_Frac:             rand.Uint32(),
		RealJitter:                      rand.Uint32(),
		Next:                            true,
		Flagged:                         false,
		Summary:                         true,
		UseDebug:                        false,
		Debug:                           "",
		RouteDiversity:                  5,
		UserFlags:                       rand.Uint64(),
		DatacenterID:                    rand.Uint64(),
		BuyerID:                         rand.Uint64(),
		UserHash:                        rand.Uint64(),
		EnvelopeBytesDown:               rand.Uint64(),
		EnvelopeBytesUp:                 rand.Uint64(),
		Latitude:                        rand.Float32(),
		Longitude:                       rand.Float32(),
		ClientAddress:                   "127.0.0.1",
		ServerAddress:                   "127.0.0.2",
		ISP:                             "ISP",
		ConnectionType:                  1,
		PlatformType:                    3,
		SDKVersion:                      "4.0.10",
		NumTags:                         int32(5),
		Tags:                            [billing.BillingEntryMaxTags]uint64{rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64()},
		ABTest:                          false,
		Pro:                             false,
		ClientToServerPacketsSent:       rand.Uint64(),
		ServerToClientPacketsSent:       rand.Uint64(),
		ClientToServerPacketsLost:       rand.Uint64(),
		ServerToClientPacketsLost:       rand.Uint64(),
		ClientToServerPacketsOutOfOrder: rand.Uint64(),
		ServerToClientPacketsOutOfOrder: rand.Uint64(),
		NumNearRelays:                   int32(5),
		NearRelayIDs:                    [billing.BillingEntryMaxNearRelays]uint64{rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64()},
		NearRelayRTTs:                   [billing.BillingEntryMaxNearRelays]int32{rand.Int31(), rand.Int31(), rand.Int31(), rand.Int31(), rand.Int31()},
		NearRelayJitters:                [billing.BillingEntryMaxNearRelays]int32{rand.Int31(), rand.Int31(), rand.Int31(), rand.Int31(), rand.Int31()},
		NearRelayPacketLosses:           [billing.BillingEntryMaxNearRelays]int32{rand.Int31(), rand.Int31(), rand.Int31(), rand.Int31(), rand.Int31()},
		TotalPriceSum:                   rand.Uint64(),
		EnvelopeBytesUpSum:              rand.Uint64(),
		EnvelopeBytesDownSum:            rand.Uint64(),
		SessionDuration:                 5 * billing.BillingSliceSeconds,
		EverOnNext:                      true,
		DurationOnNext:                  4 * billing.BillingSliceSeconds,
		StartTimestamp:                  rand.Uint32(),
		NextRTT:                         rand.Int31(),
		NextJitter:                      rand.Int31(),
		NextPacketLoss:                  rand.Int31(),
		PredictedNextRTT:                rand.Int31(),
		NearRelayRTT:                    rand.Int31(),
		NumNextRelays:                   int32(5),
		NextRelays:                      [billing.BillingEntryMaxRelays]uint64{rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64()},
		NextRelayPrice:                  [billing.BillingEntryMaxRelays]uint64{rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64()},
		TotalPrice:                      rand.Uint64(),
		Uncommitted:                     false,
		Multipath:                       false,
		RTTReduction:                    true,
		PacketLossReduction:             true,
		RouteChanged:                    false,
		NextBytesUp:                     rand.Uint64(),
		NextBytesDown:                   rand.Uint64(),
		FallbackToDirect:                false,
		MultipathVetoed:                 false,
		Mispredicted:                    false,
		Vetoed:                          false,
		LatencyWorse:                    false,
		NoRoute:                         false,
		NextLatencyTooHigh:              false,
		CommitVeto:                      false,
		UnknownDatacenter:               false,
		DatacenterNotEnabled:            false,
		BuyerNotLive:                    false,
		StaleRouteMatrix:                false,
		TryBeforeYouBuy:                 false,
	}
}

func testCountData() *transport.SessionCountData {
	return &transport.SessionCountData{
		ServerID:    rand.Uint64(),
		BuyerID:     rand.Uint64(),
		NumSessions: rand.Uint32(),
	}
}

func testPortalData() *transport.SessionPortalData {
	return &transport.SessionPortalData{
		Version: transport.SessionPortalDataVersion,
		Meta: transport.SessionMeta{
			Version:         transport.SessionMetaVersion,
			ID:              rand.Uint64(),
			UserHash:        rand.Uint64(),
			DatacenterName:  "local",
			DatacenterAlias: "alias",
			OnNetworkNext:   true,
			NextRTT:         rand.Float64(),
			DirectRTT:       rand.Float64(),
			DeltaRTT:        rand.Float64(),
			Location:        routing.LocationNullIsland,
			ClientAddr:      "127.0.0.1:34629",
			ServerAddr:      "127.0.0.1:50000",
			Hops: []transport.RelayHop{
				{
					Version: transport.RelayHopVersion,
					ID:      rand.Uint64(),
					Name:    "local.test_relay.0",
				},
				{
					Version: transport.RelayHopVersion,
					ID:      rand.Uint64(),
					Name:    "local.test_relay.1",
				},
			},
			SDK:        "4.0.0",
			Connection: 3,
			NearbyRelays: []transport.NearRelayPortalData{
				{
					Version: transport.NearRelayPortalDataVersion,
					ID:      rand.Uint64(),
					Name:    "local.test_relay.2",
				},
				{
					Version: transport.NearRelayPortalDataVersion,
					ID:      rand.Uint64(),
					Name:    "local.test_relay.3",
				},
			},
			Platform: 1,
			BuyerID:  rand.Uint64(),
		},
		Point: transport.SessionMapPoint{
			Version:   transport.SessionMapPointVersion,
			Latitude:  rand.Float64(),
			Longitude: rand.Float64(),
		},
		Slice: transport.SessionSlice{
			Version:   transport.SessionSliceVersion,
			Timestamp: time.Now(),
			Envelope: routing.Envelope{
				Up:   100,
				Down: 150,
			},
			OnNetworkNext: true,
		},
		LargeCustomer: false,
		EverOnNext:    true,
	}
}

func testMatchData() *md.MatchDataEntry {
	var matchValues [messages.MatchDataMaxMatchValues]float64
	for i := 0; i < messages.MatchDataMaxMatchValues; i++ {
		matchValues[i] = rand.ExpFloat64()
	}

	matchData := &md.MatchDataEntry{
		Version:        md.MatchDataEntryVersion,
		Timestamp:      uint32(time.Now().Unix()),
		BuyerID:        rand.Uint64(),
		ServerAddress:  "127.0.0.1",
		DatacenterID:   rand.Uint64(),
		UserHash:       rand.Uint64(),
		SessionID:      rand.Uint64(),
		MatchID:        rand.Uint64(),
		NumMatchValues: messages.MatchDataMaxMatchValues,
		MatchValues:    matchValues,
	}

	return matchData
}

func TestPostSessionHandlerSendBillingEntry2Full(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics)
	postSessionHandler.SendBillingEntry2(testBillingEntry2())

	assert.Equal(t, postSessionHandler.Billing2BufferSize(), uint64(0))
	assert.Equal(t, 1.0, metrics.Billing2BufferFull.Value())
}

func TestPostSessionHandlerSendBillingEntry2Success(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics)
	postSessionHandler.SendBillingEntry2(testBillingEntry2())

	assert.Equal(t, postSessionHandler.Billing2BufferSize(), uint64(1))
	assert.Equal(t, 1.0, metrics.BillingEntries2Sent.Value())
}

func TestPostSessionHandlerSendPortalCountsFull(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, false, &md.NoOpMatcher{}, metrics)
	postSessionHandler.SendPortalCounts(testCountData())

	assert.Equal(t, postSessionHandler.PortalCountBufferSize(), uint64(0))
	assert.Equal(t, 1.0, metrics.PortalBufferFull.Value())
}

func TestPostSessionHandlerSendPortalCountsSuccess(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, nil, 10, &billing.NoOpBiller{}, false, &md.NoOpMatcher{}, metrics)
	postSessionHandler.SendPortalCounts(testCountData())

	assert.Equal(t, postSessionHandler.PortalCountBufferSize(), uint64(1))
	assert.Equal(t, 1.0, metrics.PortalEntriesSent.Value())
}
func TestPostSessionHandlerSendPortalDataFull(t *testing.T) {

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, false, &md.NoOpMatcher{}, metrics)
	postSessionHandler.SendPortalData(testPortalData())

	assert.Equal(t, postSessionHandler.PortalDataBufferSize(), uint64(0))
	assert.Equal(t, 1.0, metrics.PortalBufferFull.Value())
}

func TestPostSessionHandlerSendPortalDataSuccess(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, nil, 10, &billing.NoOpBiller{}, false, &md.NoOpMatcher{}, metrics)
	postSessionHandler.SendPortalData(testPortalData())

	assert.Equal(t, postSessionHandler.PortalDataBufferSize(), uint64(1))
	assert.Equal(t, 1.0, metrics.PortalEntriesSent.Value())
}

func TestPostSessionHandlerTransmitPortalDataFailure(t *testing.T) {
	publisher := &badPublisher{
		calledChan: make(chan bool, 1),
	}

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, []pubsub.Publisher{publisher}, 10, &billing.NoOpBiller{}, false, &md.NoOpMatcher{}, &metrics.EmptyPostSessionMetrics)
	bytes, err := postSessionHandler.TransmitPortalData(context.Background(), 0, []byte("data"))

	assert.Zero(t, bytes)
	assert.EqualError(t, err, "bad publish")
}

func TestPostSessionHandlerTransmitPortalDataMaxRetries(t *testing.T) {
	publisher := &retryPublisher{
		retryCount: 11,
	}

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, []pubsub.Publisher{publisher}, 10, &billing.NoOpBiller{}, false, &md.NoOpMatcher{}, &metrics.EmptyPostSessionMetrics)
	bytes, err := postSessionHandler.TransmitPortalData(context.Background(), 0, []byte("data"))

	assert.Zero(t, bytes)
	assert.EqualError(t, err, "exceeded retry count on portal data")
}

func TestPostSessionHandlerTransmitPortalDataRetriesSuccess(t *testing.T) {
	publisher := &retryPublisher{
		retryCount: 5,
	}

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, []pubsub.Publisher{publisher}, 10, &billing.NoOpBiller{}, false, &md.NoOpMatcher{}, &metrics.EmptyPostSessionMetrics)
	bytes, err := postSessionHandler.TransmitPortalData(context.Background(), 0, []byte("data"))

	assert.Equal(t, 4, bytes)
	assert.NoError(t, err)
}

func TestPostSessionHandlerTransmitPortalDataSuccess(t *testing.T) {
	publisher := &retryPublisher{}

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, []pubsub.Publisher{publisher}, 0, &billing.NoOpBiller{}, false, &md.NoOpMatcher{}, &metrics.EmptyPostSessionMetrics)
	bytes, err := postSessionHandler.TransmitPortalData(context.Background(), 0, []byte("data"))

	assert.Equal(t, 4, bytes)
	assert.NoError(t, err)
}

func TestPostSessionHandlerTransmitPortalDataMultiplePublishersSuccess(t *testing.T) {
	// Have the first publisher retry too many times, but the second one succeeds
	publisher1 := &retryPublisher{
		retryCount: 11,
	}
	publisher2 := &retryPublisher{
		retryCount: 5,
	}

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, []pubsub.Publisher{publisher1, publisher2}, 10, &billing.NoOpBiller{}, false, &md.NoOpMatcher{}, &metrics.EmptyPostSessionMetrics)
	bytes, err := postSessionHandler.TransmitPortalData(context.Background(), 0, []byte("data"))

	assert.Equal(t, 4, bytes)
	assert.NoError(t, err)
}

func TestPostSessionHandlerSendMatchDataFull(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, false, &md.NoOpMatcher{}, metrics)
	postSessionHandler.SendMatchData(testMatchData())

	assert.Equal(t, postSessionHandler.MatchDataBufferSize(), uint64(0))
	assert.Equal(t, 1.0, metrics.MatchDataEntriesBufferFull.Value())
}

func TestPostSessionHandlerSendMatchDataSuccess(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, nil, 10, &billing.NoOpBiller{}, false, &md.NoOpMatcher{}, metrics)
	postSessionHandler.SendMatchData(testMatchData())

	assert.Equal(t, postSessionHandler.MatchDataBufferSize(), uint64(1))
	assert.Equal(t, 1.0, metrics.MatchDataEntriesSent.Value())
}

func TestPostSessionHandlerStartProcessingBilling2Failure(t *testing.T) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	biller2 := &badBiller{
		calledChan2: make(chan bool, 1),
	}
	publisher := &mockPublisher{}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller2, true, &md.NoOpMatcher{}, metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx, &wg)
		wg.Done()
	}()

	postSessionHandler.SendBillingEntry2(testBillingEntry2())
	<-biller2.calledChan2

	ctxCancelFunc()
	wg.Wait()

	assert.Equal(t, 1.0, metrics.Billing2Failure.Value())
	assert.Equal(t, 0.0, metrics.BillingEntries2Finished.Value())
}

func TestPostSessionHandlerStartProcessingBilling2Success(t *testing.T) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	biller2 := &mockBiller{
		calledChan2: make(chan bool, 1),
	}
	publisher := &mockPublisher{}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller2, true, &md.NoOpMatcher{}, metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx, &wg)
		wg.Done()
	}()

	postSessionHandler.SendBillingEntry2(testBillingEntry2())
	<-biller2.calledChan2

	ctxCancelFunc()
	wg.Wait()

	assert.Equal(t, 1.0, metrics.BillingEntries2Finished.Value())
	assert.Equal(t, 0.0, metrics.Billing2Failure.Value())
}

func TestPostSessionHandlerStartProcessingPortalCountFailure(t *testing.T) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	biller2 := &mockBiller{}

	publisher := &badPublisher{
		calledChan: make(chan bool, 1),
	}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller2, true, &md.NoOpMatcher{}, metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx, &wg)
		wg.Done()
	}()

	postSessionHandler.SendPortalCounts(testCountData())
	<-publisher.calledChan

	ctxCancelFunc()
	wg.Wait()

	assert.Equal(t, 1.0, metrics.PortalFailure.Value())
	assert.Equal(t, 0.0, metrics.PortalEntriesFinished.Value())
}

func TestPostSessionHandlerStartProcessingPortalCountSuccess(t *testing.T) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	biller2 := &mockBiller{}

	publisher := &mockPublisher{
		calledChan: make(chan bool, 1),
	}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller2, true, &md.NoOpMatcher{}, metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx, &wg)
		wg.Done()
	}()

	countData := testCountData()
	countDataBytes, err := transport.WriteSessionCountData(countData)
	assert.NoError(t, err)

	postSessionHandler.SendPortalCounts(countData)
	<-publisher.calledChan

	ctxCancelFunc()
	wg.Wait()

	assert.Equal(t, 1.0, metrics.PortalEntriesFinished.Value())
	assert.Equal(t, 0.0, metrics.PortalFailure.Value())

	assert.Len(t, publisher.publishedMessages, 1)
	assert.Len(t, publisher.publishedMessages[0], 2)
	assert.Equal(t, []byte{byte(pubsub.TopicPortalCruncherSessionCounts)}, publisher.publishedMessages[0][0])
	assert.Equal(t, countDataBytes, publisher.publishedMessages[0][1])
}

func TestPostSessionHandlerStartProcessingPortalDataFailure(t *testing.T) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	biller2 := &mockBiller{}

	publisher := &badPublisher{
		calledChan: make(chan bool, 1),
	}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller2, true, &md.NoOpMatcher{}, metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx, &wg)
		wg.Done()
	}()

	postSessionHandler.SendPortalData(testPortalData())
	<-publisher.calledChan

	ctxCancelFunc()
	wg.Wait()

	assert.Equal(t, 1.0, metrics.PortalFailure.Value())
	assert.Equal(t, 0.0, metrics.PortalEntriesFinished.Value())
}

func TestPostSessionHandlerStartProcessingPortalDataSuccess(t *testing.T) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	biller2 := &mockBiller{}

	publisher := &mockPublisher{
		calledChan: make(chan bool, 1),
	}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller2, true, &md.NoOpMatcher{}, metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx, &wg)
		wg.Done()
	}()

	portalData := testPortalData()
	portalDataBytes, err := transport.WriteSessionPortalData(portalData)
	assert.NoError(t, err)

	postSessionHandler.SendPortalData(portalData)
	<-publisher.calledChan

	ctxCancelFunc()
	wg.Wait()

	assert.Equal(t, 1.0, metrics.PortalEntriesFinished.Value())
	assert.Equal(t, 0.0, metrics.PortalFailure.Value())

	assert.Len(t, publisher.publishedMessages, 1)
	assert.Len(t, publisher.publishedMessages[0], 2)
	assert.Equal(t, []byte{byte(pubsub.TopicPortalCruncherSessionData)}, publisher.publishedMessages[0][0])
	assert.Equal(t, portalDataBytes, publisher.publishedMessages[0][1])
}

func TestPostSessionHandlerStartProcessingMatchDataFailure(t *testing.T) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	matcher := &badMatcher{
		calledChan: make(chan bool, 1),
	}
	publisher := &mockPublisher{}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, &billing.NoOpBiller{}, false, matcher, metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx, &wg)
		wg.Done()
	}()

	postSessionHandler.SendMatchData(testMatchData())
	<-matcher.calledChan

	ctxCancelFunc()
	wg.Wait()

	assert.Equal(t, 1.0, metrics.MatchDataEntriesFailure.Value())
	assert.Equal(t, 0.0, metrics.MatchDataEntriesFinished.Value())
}

func TestPostSessionHandlerStartProcessingMatchDataSuccess(t *testing.T) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	matcher := &mockMatcher{
		calledChan: make(chan bool, 1),
	}
	publisher := &mockPublisher{}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, &billing.NoOpBiller{}, false, matcher, metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx, &wg)
		wg.Done()
	}()

	postSessionHandler.SendMatchData(testMatchData())
	<-matcher.calledChan

	ctxCancelFunc()
	wg.Wait()

	assert.Equal(t, 1.0, metrics.MatchDataEntriesFinished.Value())
	assert.Equal(t, 0.0, metrics.MatchDataEntriesFailure.Value())
}
