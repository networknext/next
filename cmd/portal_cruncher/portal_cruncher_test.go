package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
	"github.com/stretchr/testify/assert"
)

func getTestSessionData(t *testing.T, onNetworkNext bool) transport.SessionPortalData {
	relayID1 := crypto.HashID("127.0.0.1:10000")
	relayID2 := crypto.HashID("127.0.0.1:10001")
	relayID3 := crypto.HashID("127.0.0.1:10002")
	relayID4 := crypto.HashID("127.0.0.1:10003")

	return transport.SessionPortalData{
		Meta: transport.SessionMeta{
			ID:              rand.Uint64(),
			UserHash:        rand.Uint64(),
			DatacenterName:  "local",
			DatacenterAlias: "alias",
			OnNetworkNext:   onNetworkNext,
			NextRTT:         20,
			DirectRTT:       50,
			DeltaRTT:        30,
			Location:        routing.LocationNullIsland,
			ClientAddr:      "127.0.0.1:34629",
			ServerAddr:      "127.0.0.1:50000",
			Hops: []transport.RelayHop{
				{
					ID:   relayID1,
					Name: "local.test_relay.0",
				},
				{
					ID:   relayID2,
					Name: "local.test_relay.1",
				},
			},
			SDK:        "4.0.0",
			Connection: 3,
			NearbyRelays: []transport.NearRelayPortalData{
				{
					ID:   relayID3,
					Name: "local.test_relay.2",
				},
				{
					ID:   relayID4,
					Name: "local.test_relay.3",
				},
			},
			Platform: 1,
			BuyerID:  12345,
		},
		Point: transport.SessionMapPoint{
			Latitude:  45,
			Longitude: 90,
		},
		Slice: transport.SessionSlice{
			Timestamp: time.Now().Truncate(time.Second),
			Envelope: routing.Envelope{
				Up:   100,
				Down: 150,
			},
			OnNetworkNext: onNetworkNext,
		},
	}
}

func runTestPubSub(t *testing.T, portalPublisher pubsub.Publisher, portalSubscriber pubsub.Subscriber, topic pubsub.Topic, message []byte, totalMessages int, messageChan chan []byte, metrics *metrics.PortalCruncherMetrics) error {
	messagesSent := 0
	for messagesSent < totalMessages {
		_, err := portalPublisher.Publish(topic, message)
		if err != nil {
			return err
		}

		messagesSent++
	}

	messagesReceived := 0
	for messagesReceived < totalMessages {
		err := ReceivePortalMessage(portalSubscriber, metrics, messageChan)
		if err != nil {
			return err
		}

		messagesReceived++
	}

	return nil
}

type redisServers struct {
	topSessions   *storage.MockRedisServer
	sessionMap    *storage.MockRedisServer
	sessionMeta   *storage.MockRedisServer
	sessionSlices *storage.MockRedisServer
}

type redisClients struct {
	topSessions   *storage.RawRedisClient
	sessionMap    *storage.RawRedisClient
	sessionMeta   *storage.RawRedisClient
	sessionSlices *storage.RawRedisClient
}

func setupMockRedis(ctx context.Context, t *testing.T) (*redisServers, *redisClients) {
	redisTopSessions, err := storage.NewMockRedisServer(ctx, "127.0.0.1:6380")
	assert.NoError(t, err)

	redisSessionMap, err := storage.NewMockRedisServer(ctx, "127.0.0.1:6381")
	assert.NoError(t, err)

	redisSessionMeta, err := storage.NewMockRedisServer(ctx, "127.0.0.1:6382")
	assert.NoError(t, err)

	redisSessionSlices, err := storage.NewMockRedisServer(ctx, "127.0.0.1:6383")
	assert.NoError(t, err)

	errCallback := func(err error) {
		fmt.Println("err", err)
	}

	go func() {
		redisTopSessions.Start(errCallback)
		fmt.Println("1 finished")
	}()

	go func() {
		redisSessionMap.Start(errCallback)
		fmt.Println("2 finished")
	}()

	go func() {
		redisSessionMeta.Start(errCallback)
		fmt.Println("3 finished")
	}()

	go func() {
		redisSessionSlices.Start(errCallback)
		fmt.Println("4 finished")
	}()

	clientTopSessions, clientSessionMap, clientSessionMeta, clientSessionSlices, err := createRedis(redisTopSessions.Addr().String(), redisSessionMap.Addr().String(), redisSessionMeta.Addr().String(), redisSessionSlices.Addr().String())
	assert.NoError(t, err)

	err = pingRedis(clientTopSessions, clientSessionMap, clientSessionMeta, clientSessionSlices)
	assert.NoError(t, err)

	return &redisServers{
			topSessions:   redisTopSessions,
			sessionMap:    redisSessionMap,
			sessionMeta:   redisSessionMeta,
			sessionSlices: redisSessionSlices,
		}, &redisClients{
			topSessions:   clientTopSessions,
			sessionMap:    clientSessionMap,
			sessionMeta:   clientSessionMeta,
			sessionSlices: clientSessionSlices,
		}
}

func TestReceive(t *testing.T) {
	portalSubscriber, err := pubsub.NewPortalCruncherSubscriber("5555", 1000)
	assert.NoError(t, err)

	err = portalSubscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)
	assert.NoError(t, err)

	portalPublisher, err := pubsub.NewPortalCruncherPublisher("tcp://127.0.0.1:5555", 1000)
	assert.NoError(t, err)

	// Poll for 10ms to set up ZeroMQ's internal state correctly and not miss the first message
	err = portalSubscriber.Poll(time.Millisecond * 10)
	assert.NoError(t, err)

	t.Run("channel full", func(t *testing.T) {
		var sessionPortalData transport.SessionPortalData
		sessionData, err := sessionPortalData.MarshalBinary()
		assert.NoError(t, err)

		messageChanSize := 100
		messageChan := make(chan []byte, messageChanSize)

		metricsHandler := &metrics.LocalHandler{}
		metrics, err := metrics.NewPortalCruncherMetrics(context.Background(), metricsHandler)
		assert.NoError(t, err)

		err = runTestPubSub(t, portalPublisher, portalSubscriber, pubsub.TopicPortalCruncherSessionData, sessionData, messageChanSize+1, messageChan, metrics)
		expectedErr := &ErrChannelFull{}
		assert.EqualError(t, err, expectedErr.Error())

		assert.Equal(t, float64(messageChanSize+1), metrics.ReceivedMessageCount.Value())
		assert.Len(t, messageChan, messageChanSize)
		for i := 0; i < messageChanSize; i++ {
			actualSessionData := <-messageChan
			assert.Equal(t, sessionData, actualSessionData)
		}
	})

	t.Run("success", func(t *testing.T) {
		var sessionPortalData transport.SessionPortalData
		sessionData, err := sessionPortalData.MarshalBinary()
		assert.NoError(t, err)

		messageChanSize := 100
		messageChan := make(chan []byte, messageChanSize)

		metricsHandler := &metrics.LocalHandler{}
		metrics, err := metrics.NewPortalCruncherMetrics(context.Background(), metricsHandler)
		assert.NoError(t, err)

		err = runTestPubSub(t, portalPublisher, portalSubscriber, pubsub.TopicPortalCruncherSessionData, sessionData, messageChanSize, messageChan, metrics)
		assert.NoError(t, err)

		assert.Equal(t, float64(messageChanSize), metrics.ReceivedMessageCount.Value())

		assert.Len(t, messageChan, messageChanSize)
		for i := 0; i < messageChanSize; i++ {
			actualSessionData := <-messageChan
			assert.Equal(t, sessionData, actualSessionData)
		}
	})
}

func TestPullMessageUnmarshalFailure(t *testing.T) {
	messageChan := make(chan []byte, 1)
	messageChan <- []byte("test")

	sessionData, err := PullMessage(messageChan)
	assert.Error(t, err)
	assert.Empty(t, sessionData)
}

func TestPullMessageSuccess(t *testing.T) {
	expected := getTestSessionData(t, true)
	expectedBytes, err := expected.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- expectedBytes

	actual, err := PullMessage(messageChan)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestRedisClientCreateFailure(t *testing.T) {
	redisTopSessions, err := miniredis.Run()
	assert.NoError(t, err)

	redisSessionMap, err := miniredis.Run()
	assert.NoError(t, err)

	redisSessionSlices, err := miniredis.Run()
	assert.NoError(t, err)

	clientTopSessions, clientSessionMap, clientSessionMeta, clientSessionSlices, err := createRedis(redisTopSessions.Addr(), redisSessionMap.Addr(), "", redisSessionSlices.Addr())
	assert.Error(t, err)
	assert.Nil(t, clientTopSessions)
	assert.Nil(t, clientSessionMap)
	assert.Nil(t, clientSessionMeta)
	assert.Nil(t, clientSessionSlices)
}

func TestRedisClientCreateSuccess(t *testing.T) {
	redisTopSessions, err := miniredis.Run()
	assert.NoError(t, err)

	redisSessionMap, err := miniredis.Run()
	assert.NoError(t, err)

	redisSessionMeta, err := miniredis.Run()
	assert.NoError(t, err)

	redisSessionSlices, err := miniredis.Run()
	assert.NoError(t, err)

	clientTopSessions, clientSessionMap, clientSessionMeta, clientSessionSlices, err := createRedis(redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr())
	assert.NoError(t, err)
	assert.NotNil(t, clientTopSessions)
	assert.NotNil(t, clientSessionMap)
	assert.NotNil(t, clientSessionMeta)
	assert.NotNil(t, clientSessionSlices)
}

func TestDirectSession(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	servers, clients := setupMockRedis(ctx, t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	storer := &storage.InMemory{}
	err := storer.AddBuyer(ctx, routing.Buyer{
		ID:          12345,
		CompanyCode: "local",
	})
	assert.NoError(t, err)

	var largeCustomerCacheMap sync.Map

	flushCount := 1000

	sessionData := getTestSessionData(t, false)
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(clients.topSessions, clients.sessionMap, clients.sessionMeta, clients.sessionSlices, messageChan, portalDataBuffer, flushTime, pingTime, storer, &largeCustomerCacheMap, flushCount)
	assert.NoError(t, err)

	// Add a small delay so that the data has time to go over the tcp sockets
	time.Sleep(time.Millisecond * 10)

	cancelFunc()

	val, err := servers.sessionMeta.Storage.Get(fmt.Sprintf("sm-%016x", sessionData.Meta.ID))
	assert.NoError(t, err)

	val = val[1 : len(val)-1] // Remove the extra quotes

	assert.Equal(t, sessionData.Meta.RedisString(), val)

	fmt.Println("top sessions", servers.topSessions.Dump())
	fmt.Println("session map", servers.sessionMap.Dump())
	fmt.Println("session meta", servers.sessionMeta.Dump())
	fmt.Println("session slices", servers.sessionSlices.Dump())
}

func TestNextSession(t *testing.T) {

}

func TestDirectSessionLargeCustomer(t *testing.T) {

}

func TestNextSessionLargeCustomer(t *testing.T) {

}

func TestDirectToNextLargeCustomer(t *testing.T) {

}

func TestNextToDirectLargeCustomer(t *testing.T) {

}
