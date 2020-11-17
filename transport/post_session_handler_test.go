package transport_test

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
	"github.com/stretchr/testify/assert"
)

type badBiller struct {
	calledChan chan bool
}

func (biller *badBiller) Bill(ctx context.Context, billingEntry *billing.BillingEntry) error {
	biller.calledChan <- true
	return errors.New("bad bill")
}

type mockBiller struct {
	calledChan    chan bool
	billedEntries []billing.BillingEntry
}

func (biller *mockBiller) Bill(ctx context.Context, billingEntry *billing.BillingEntry) error {
	biller.billedEntries = append(biller.billedEntries, *billingEntry)

	biller.calledChan <- true
	return nil
}

type badPublisher struct {
	calledChan chan bool
}

func (pub *badPublisher) Publish(ctx context.Context, topic pubsub.Topic, message []byte) (int, error) {
	pub.calledChan <- true
	return 0, errors.New("bad publish")
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

type mockPublisher struct {
	calledChan        chan bool
	publishedMessages [][][]byte
}

func (pub *mockPublisher) Publish(ctx context.Context, topic pubsub.Topic, message []byte) (int, error) {
	pub.publishedMessages = append(pub.publishedMessages, [][]byte{{byte(topic)}, message})

	pub.calledChan <- true
	return len(message), nil
}

func testBillingEntry() *billing.BillingEntry {
	return &billing.BillingEntry{
		BuyerID:                   rand.Uint64(),
		UserHash:                  rand.Uint64(),
		SessionID:                 rand.Uint64(),
		SliceNumber:               rand.Uint32(),
		DirectRTT:                 rand.Float32(),
		DirectJitter:              rand.Float32(),
		DirectPacketLoss:          rand.Float32(),
		Next:                      true,
		NextRTT:                   rand.Float32(),
		NextJitter:                rand.Float32(),
		NextPacketLoss:            rand.Float32(),
		NumNextRelays:             5,
		NextRelays:                [5]uint64{rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64()},
		TotalPrice:                1000000,
		ClientToServerPacketsLost: 0,
		ServerToClientPacketsLost: 0,
		Committed:                 true,
		Flagged:                   false,
		Multipath:                 false,
		Initial:                   false,
		NextBytesUp:               512,
		NextBytesDown:             256,
		EnvelopeBytesUp:           1024,
		EnvelopeBytesDown:         1024,
		DatacenterID:              rand.Uint64(),
		RTTReduction:              true,
		PacketLossReduction:       true,
		NextRelaysPrice:           [5]uint64{rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64()},
		Latitude:                  rand.Float32(),
		Longitude:                 rand.Float32(),
		ISP:                       "ISP",
		ABTest:                    false,
		RouteDecision:             0,
		ConnectionType:            1,
		PlatformType:              3,
		SDKVersion:                "4.0.0",
		PacketLoss:                rand.Float32(),
		PredictedNextRTT:          rand.Float32(),
		MultipathVetoed:           false,
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
		Meta: transport.SessionMeta{
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
					ID:   rand.Uint64(),
					Name: "local.test_relay.0",
				},
				{
					ID:   rand.Uint64(),
					Name: "local.test_relay.1",
				},
			},
			SDK:        "4.0.0",
			Connection: 3,
			NearbyRelays: []transport.NearRelayPortalData{
				{
					ID:   rand.Uint64(),
					Name: "local.test_relay.2",
				},
				{
					ID:   rand.Uint64(),
					Name: "local.test_relay.3",
				},
			},
			Platform: 1,
			BuyerID:  rand.Uint64(),
		},
		Point: transport.SessionMapPoint{
			Latitude:  rand.Float64(),
			Longitude: rand.Float64(),
		},
		Slice: transport.SessionSlice{
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

func TestPostSessionHandlerSendBillingEntryFull(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, log.NewNopLogger(), metrics)
	postSessionHandler.SendBillingEntry(testBillingEntry())

	assert.Equal(t, postSessionHandler.BillingBufferSize(), uint64(0))
	assert.Equal(t, 1.0, metrics.BillingBufferFull.Value())
}

func TestPostSessionHandlerSendBillingEntrySuccess(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, nil, 10, &billing.NoOpBiller{}, log.NewNopLogger(), metrics)
	postSessionHandler.SendBillingEntry(testBillingEntry())

	assert.Equal(t, postSessionHandler.BillingBufferSize(), uint64(1))
	assert.Equal(t, 1.0, metrics.BillingEntriesSent.Value())
}

func TestPostSessionHandlerSendPortalCountsFull(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, log.NewNopLogger(), metrics)
	postSessionHandler.SendPortalCounts(testCountData())

	assert.Equal(t, postSessionHandler.PortalCountBufferSize(), uint64(0))
	assert.Equal(t, 1.0, metrics.PortalBufferFull.Value())
}

func TestPostSessionHandlerSendPortalCountsSuccess(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, nil, 10, &billing.NoOpBiller{}, log.NewNopLogger(), metrics)
	postSessionHandler.SendPortalCounts(testCountData())

	assert.Equal(t, postSessionHandler.PortalCountBufferSize(), uint64(1))
	assert.Equal(t, 1.0, metrics.PortalEntriesSent.Value())
}
func TestPostSessionHandlerSendPortalDataFull(t *testing.T) {

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, log.NewNopLogger(), metrics)
	postSessionHandler.SendPortalData(testPortalData())

	assert.Equal(t, postSessionHandler.PortalDataBufferSize(), uint64(0))
	assert.Equal(t, 1.0, metrics.PortalBufferFull.Value())
}

func TestPostSessionHandlerSendPortalDataSuccess(t *testing.T) {
	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(context.Background(), metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, nil, 10, &billing.NoOpBiller{}, log.NewNopLogger(), metrics)
	postSessionHandler.SendPortalData(testPortalData())

	assert.Equal(t, postSessionHandler.PortalDataBufferSize(), uint64(1))
	assert.Equal(t, 1.0, metrics.PortalEntriesSent.Value())
}

func TestPostSessionHandlerTransmitPortalDataFailure(t *testing.T) {
	publisher := &badPublisher{
		calledChan: make(chan bool, 1),
	}

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, []pubsub.Publisher{publisher}, 10, &billing.NoOpBiller{}, log.NewNopLogger(), &metrics.EmptyPostSessionMetrics)
	bytes, err := postSessionHandler.TransmitPortalData(context.Background(), 0, []byte("data"))

	assert.Zero(t, bytes)
	assert.EqualError(t, err, "bad publish")
}

func TestPostSessionHandlerTransmitPortalDataMaxRetries(t *testing.T) {
	publisher := &retryPublisher{
		retryCount: 11,
	}

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, []pubsub.Publisher{publisher}, 10, &billing.NoOpBiller{}, log.NewNopLogger(), &metrics.EmptyPostSessionMetrics)
	bytes, err := postSessionHandler.TransmitPortalData(context.Background(), 0, []byte("data"))

	assert.Zero(t, bytes)
	assert.EqualError(t, err, "exceeded retry count on portal data")
}

func TestPostSessionHandlerTransmitPortalDataRetriesSuccess(t *testing.T) {
	publisher := &retryPublisher{
		retryCount: 5,
	}

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, []pubsub.Publisher{publisher}, 10, &billing.NoOpBiller{}, log.NewNopLogger(), &metrics.EmptyPostSessionMetrics)
	bytes, err := postSessionHandler.TransmitPortalData(context.Background(), 0, []byte("data"))

	assert.Equal(t, 4, bytes)
	assert.NoError(t, err)
}

func TestPostSessionHandlerTransmitPortalDataSuccess(t *testing.T) {
	publisher := &retryPublisher{}

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, []pubsub.Publisher{publisher}, 0, &billing.NoOpBiller{}, log.NewNopLogger(), &metrics.EmptyPostSessionMetrics)
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

	postSessionHandler := transport.NewPostSessionHandler(4, 1000, []pubsub.Publisher{publisher1, publisher2}, 10, &billing.NoOpBiller{}, log.NewNopLogger(), &metrics.EmptyPostSessionMetrics)
	bytes, err := postSessionHandler.TransmitPortalData(context.Background(), 0, []byte("data"))

	assert.Equal(t, 4, bytes)
	assert.NoError(t, err)
}

func TestPostSessionHandlerStartProcessingBillingFailure(t *testing.T) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	biller := &badBiller{
		calledChan: make(chan bool, 1),
	}
	publisher := &mockPublisher{}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller, log.NewNopLogger(), metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx)
		wg.Done()
	}()

	postSessionHandler.SendBillingEntry(testBillingEntry())
	<-biller.calledChan

	ctxCancelFunc()
	wg.Wait()

	assert.Equal(t, 1.0, metrics.BillingFailure.Value())
	assert.Equal(t, 0.0, metrics.BillingEntriesFinished.Value())
}

func TestPostSessionHandlerStartProcessingBillingSuccess(t *testing.T) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	biller := &mockBiller{
		calledChan: make(chan bool, 1),
	}
	publisher := &mockPublisher{}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller, log.NewNopLogger(), metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx)
		wg.Done()
	}()

	postSessionHandler.SendBillingEntry(testBillingEntry())
	<-biller.calledChan

	ctxCancelFunc()
	wg.Wait()

	assert.Equal(t, 1.0, metrics.BillingEntriesFinished.Value())
	assert.Equal(t, 0.0, metrics.BillingFailure.Value())
}

func TestPostSessionHandlerStartProcessingPortalCountFailure(t *testing.T) {
	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	biller := &mockBiller{}
	publisher := &badPublisher{
		calledChan: make(chan bool, 1),
	}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller, log.NewNopLogger(), metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx)
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

	biller := &mockBiller{}
	publisher := &mockPublisher{
		calledChan: make(chan bool, 1),
	}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller, log.NewNopLogger(), metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx)
		wg.Done()
	}()

	countData := testCountData()
	countDataBytes, err := countData.MarshalBinary()
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

	biller := &mockBiller{}
	publisher := &badPublisher{
		calledChan: make(chan bool, 1),
	}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller, log.NewNopLogger(), metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx)
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

	biller := &mockBiller{}
	publisher := &mockPublisher{
		calledChan: make(chan bool, 1),
	}

	metricsHandler := &metrics.LocalHandler{}
	metrics, err := metrics.NewPostSessionMetrics(ctx, metricsHandler, "server_backend")
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(1, 1000, []pubsub.Publisher{publisher}, 0, biller, log.NewNopLogger(), metrics)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		postSessionHandler.StartProcessing(ctx)
		wg.Done()
	}()

	portalData := testPortalData()
	portalDataBytes, err := portalData.MarshalBinary()
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
