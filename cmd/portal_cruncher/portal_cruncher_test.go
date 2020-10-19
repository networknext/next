package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
	"github.com/stretchr/testify/assert"
)

func getTestSessionData(t *testing.T, sessionID uint64, userHash uint64, onNetworkNext bool, everOnNetworkNext bool, timestamp time.Time) transport.SessionPortalData {
	relayID1 := crypto.HashID("127.0.0.1:10000")
	relayID2 := crypto.HashID("127.0.0.1:10001")
	relayID3 := crypto.HashID("127.0.0.1:10002")
	relayID4 := crypto.HashID("127.0.0.1:10003")

	return transport.SessionPortalData{
		Meta: transport.SessionMeta{
			ID:              sessionID,
			UserHash:        userHash,
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
			Timestamp: timestamp.Truncate(time.Second),
			Envelope: routing.Envelope{
				Up:   100,
				Down: 150,
			},
			OnNetworkNext: onNetworkNext,
		},
		EverOnNext: everOnNetworkNext,
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
	topSessions   *miniredis.Miniredis
	sessionMap    *miniredis.Miniredis
	sessionMeta   *miniredis.Miniredis
	sessionSlices *miniredis.Miniredis
}

type redisClients struct {
	topSessions   *storage.RawRedisClient
	sessionMap    *storage.RawRedisClient
	sessionMeta   *storage.RawRedisClient
	sessionSlices *storage.RawRedisClient
}

func setupMockRedis(t *testing.T) (*redisServers, *redisClients) {
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

	// Poll for 100ms to set up ZeroMQ's internal state correctly and not miss the first message
	err = portalSubscriber.Poll(time.Millisecond * 100)
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
	expected := getTestSessionData(t, rand.Uint64(), rand.Uint64(), true, false, time.Now())
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
	ctx := context.Background()

	servers, clients := setupMockRedis(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	storer := &storage.InMemory{}
	err := storer.AddBuyer(ctx, routing.Buyer{
		ID:          12345,
		CompanyCode: "local",
	})
	assert.NoError(t, err)

	flushCount := 1000

	sessionData := getTestSessionData(t, rand.Uint64(), rand.Uint64(), false, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(clients.topSessions, clients.sessionMap, clients.sessionMeta, clients.sessionSlices, messageChan, portalDataBuffer, flushTime, pingTime, storer, flushCount)
	assert.NoError(t, err)

	// Add a small delay so that the data has time to go over the tcp sockets
	time.Sleep(time.Millisecond * 50)

	minutes := flushTime.Unix() / 60

	{
		topSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		customerTopSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("sc-%016x-%d", sessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err := strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		pointVal := servers.sessionMap.HGet(fmt.Sprintf("d-%016x-%d", sessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", sessionData.Meta.ID))
		pointVal = pointVal[1 : len(pointVal)-1] // Remove the extra quotes
		assert.Equal(t, sessionData.Point.RedisString(), pointVal)
	}

	{
		metaVal, err := servers.sessionMeta.Get(fmt.Sprintf("sm-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)

		metaVal = metaVal[1 : len(metaVal)-1] // Remove the extra quotes
		assert.Equal(t, sessionData.Meta.RedisString(), metaVal)
	}

	{
		sliceVals, err := servers.sessionSlices.List(fmt.Sprintf("ss-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)
		assert.Len(t, sliceVals, 1)

		sliceVal := sliceVals[0]

		sliceVal = sliceVal[1 : len(sliceVal)-1] // Remove the extra quotes
		assert.Equal(t, sessionData.Slice.RedisString(), sliceVal)
	}
}

func TestNextSession(t *testing.T) {
	ctx := context.Background()

	servers, clients := setupMockRedis(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	storer := &storage.InMemory{}
	err := storer.AddBuyer(ctx, routing.Buyer{
		ID:          12345,
		CompanyCode: "local",
	})
	assert.NoError(t, err)

	flushCount := 1000

	sessionData := getTestSessionData(t, rand.Uint64(), rand.Uint64(), true, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(clients.topSessions, clients.sessionMap, clients.sessionMeta, clients.sessionSlices, messageChan, portalDataBuffer, flushTime, pingTime, storer, flushCount)
	assert.NoError(t, err)

	// Add a small delay so that the data has time to go over the tcp sockets
	time.Sleep(time.Millisecond * 50)

	minutes := flushTime.Unix() / 60

	{
		topSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		customerTopSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("sc-%016x-%d", sessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err := strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		pointVal := servers.sessionMap.HGet(fmt.Sprintf("n-%016x-%d", sessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", sessionData.Meta.ID))
		pointVal = pointVal[1 : len(pointVal)-1] // Remove the extra quotes
		assert.Equal(t, sessionData.Point.RedisString(), pointVal)
	}

	{
		metaVal, err := servers.sessionMeta.Get(fmt.Sprintf("sm-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)

		metaVal = metaVal[1 : len(metaVal)-1] // Remove the extra quotes
		assert.Equal(t, sessionData.Meta.RedisString(), metaVal)
	}

	{
		sliceVals, err := servers.sessionSlices.List(fmt.Sprintf("ss-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)
		assert.Len(t, sliceVals, 1)

		sliceVal := sliceVals[0]

		sliceVal = sliceVal[1 : len(sliceVal)-1] // Remove the extra quotes
		assert.Equal(t, sessionData.Slice.RedisString(), sliceVal)
	}
}

func TestDirectSessionLargeCustomer(t *testing.T) {
	ctx := context.Background()

	servers, clients := setupMockRedis(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	storer := &storage.InMemory{}
	err := storer.AddBuyer(ctx, routing.Buyer{
		ID:          12345,
		CompanyCode: "local",
		InternalConfig: core.InternalConfig{
			LargeCustomer: true,
		},
	})
	assert.NoError(t, err)

	flushCount := 1000

	sessionData := getTestSessionData(t, rand.Uint64(), rand.Uint64(), false, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(clients.topSessions, clients.sessionMap, clients.sessionMeta, clients.sessionSlices, messageChan, portalDataBuffer, flushTime, pingTime, storer, flushCount)
	assert.NoError(t, err)

	// Add a small delay so that the data has time to go over the tcp sockets
	time.Sleep(time.Millisecond * 50)

	minutes := flushTime.Unix() / 60

	{
		topSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.Error(t, err)
		assert.Len(t, topSessionIDs, 0)
	}

	{
		customerTopSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("sc-%016x-%d", sessionData.Meta.BuyerID, minutes))
		assert.Error(t, err)
		assert.Len(t, customerTopSessionIDs, 0)
	}

	{
		pointVal := servers.sessionMap.HGet(fmt.Sprintf("d-%016x-%d", sessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", sessionData.Meta.ID))
		assert.Empty(t, pointVal)
	}

	{
		metaVal, err := servers.sessionMeta.Get(fmt.Sprintf("sm-%016x", sessionData.Meta.ID))
		assert.Error(t, err)
		assert.Empty(t, metaVal)
	}

	{
		sliceVals, err := servers.sessionSlices.List(fmt.Sprintf("ss-%016x", sessionData.Meta.ID))
		assert.Error(t, err)
		assert.Len(t, sliceVals, 0)
	}
}

func TestNextSessionLargeCustomer(t *testing.T) {
	ctx := context.Background()

	servers, clients := setupMockRedis(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	storer := &storage.InMemory{}
	err := storer.AddBuyer(ctx, routing.Buyer{
		ID:          12345,
		CompanyCode: "local",
		InternalConfig: core.InternalConfig{
			LargeCustomer: true,
		},
	})
	assert.NoError(t, err)

	flushCount := 1000

	sessionData := getTestSessionData(t, rand.Uint64(), rand.Uint64(), true, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(clients.topSessions, clients.sessionMap, clients.sessionMeta, clients.sessionSlices, messageChan, portalDataBuffer, flushTime, pingTime, storer, flushCount)
	assert.NoError(t, err)

	// Add a small delay so that the data has time to go over the tcp sockets
	time.Sleep(time.Millisecond * 50)

	minutes := flushTime.Unix() / 60

	{
		topSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		customerTopSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("sc-%016x-%d", sessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err := strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		pointVal := servers.sessionMap.HGet(fmt.Sprintf("n-%016x-%d", sessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", sessionData.Meta.ID))
		pointVal = pointVal[1 : len(pointVal)-1] // Remove the extra quotes
		assert.Equal(t, sessionData.Point.RedisString(), pointVal)
	}

	{
		metaVal, err := servers.sessionMeta.Get(fmt.Sprintf("sm-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)

		metaVal = metaVal[1 : len(metaVal)-1] // Remove the extra quotes
		assert.Equal(t, sessionData.Meta.RedisString(), metaVal)
	}

	{
		sliceVals, err := servers.sessionSlices.List(fmt.Sprintf("ss-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)
		assert.Len(t, sliceVals, 1)

		sliceVal := sliceVals[0]

		sliceVal = sliceVal[1 : len(sliceVal)-1] // Remove the extra quotes
		assert.Equal(t, sessionData.Slice.RedisString(), sliceVal)
	}
}

func TestDirectToNextLargeCustomer(t *testing.T) {
	ctx := context.Background()

	servers, clients := setupMockRedis(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	storer := &storage.InMemory{}
	err := storer.AddBuyer(ctx, routing.Buyer{
		ID:          12345,
		CompanyCode: "local",
		InternalConfig: core.InternalConfig{
			LargeCustomer: true,
		},
	})
	assert.NoError(t, err)

	flushCount := 1000

	sessionID := rand.Uint64()
	userHash := rand.Uint64()
	directSessionData := getTestSessionData(t, sessionID, userHash, false, false, flushTime)

	minutes := flushTime.Unix() / 60

	_, err = servers.topSessions.ZAdd(fmt.Sprintf("s-%d", minutes), directSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", directSessionData.Meta.ID))
	assert.NoError(t, err)

	_, err = servers.topSessions.ZAdd(fmt.Sprintf("sc-%016x-%d", directSessionData.Meta.BuyerID, minutes), directSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", directSessionData.Meta.ID))
	assert.NoError(t, err)

	servers.sessionMap.HSet(fmt.Sprintf("d-%016x-%d", directSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", directSessionData.Meta.ID))

	err = servers.sessionMeta.Set(fmt.Sprintf("sm-%016x", directSessionData.Meta.ID), "\""+directSessionData.Meta.RedisString()+"\"")
	assert.NoError(t, err)

	_, err = servers.sessionSlices.RPush(fmt.Sprintf("ss-%016x", directSessionData.Meta.ID), "\""+directSessionData.Slice.RedisString()+"\"")
	assert.NoError(t, err)

	nextSessionData := getTestSessionData(t, sessionID, userHash, true, false, flushTime)

	sessionDataBytes, err := nextSessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(clients.topSessions, clients.sessionMap, clients.sessionMeta, clients.sessionSlices, messageChan, portalDataBuffer, flushTime, pingTime, storer, flushCount)
	assert.NoError(t, err)

	// Add a small delay so that the data has time to go over the tcp sockets
	time.Sleep(time.Millisecond * 50)

	{
		topSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, nextSessionData.Meta.ID, sessionID)
	}

	{
		customerTopSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("sc-%016x-%d", nextSessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err := strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, nextSessionData.Meta.ID, sessionID)
	}

	{
		pointVal := servers.sessionMap.HGet(fmt.Sprintf("n-%016x-%d", nextSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", nextSessionData.Meta.ID))
		pointVal = pointVal[1 : len(pointVal)-1] // Remove the extra quotes
		assert.Equal(t, nextSessionData.Point.RedisString(), pointVal)
	}

	{
		metaVal, err := servers.sessionMeta.Get(fmt.Sprintf("sm-%016x", nextSessionData.Meta.ID))
		assert.NoError(t, err)

		metaVal = metaVal[1 : len(metaVal)-1] // Remove the extra quotes
		assert.Equal(t, nextSessionData.Meta.RedisString(), metaVal)
	}

	{
		sliceVals, err := servers.sessionSlices.List(fmt.Sprintf("ss-%016x", nextSessionData.Meta.ID))
		assert.NoError(t, err)
		assert.Len(t, sliceVals, 2)

		directSliceVal := sliceVals[0]
		nextSliceVal := sliceVals[1]

		directSliceVal = directSliceVal[1 : len(directSliceVal)-1] // Remove the extra quotes
		assert.Equal(t, directSessionData.Slice.RedisString(), directSliceVal)

		nextSliceVal = nextSliceVal[1 : len(nextSliceVal)-1] // Remove the extra quotes
		assert.Equal(t, nextSessionData.Slice.RedisString(), nextSliceVal)
	}
}

func TestNextToDirectLargeCustomer(t *testing.T) {
	ctx := context.Background()

	servers, clients := setupMockRedis(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	storer := &storage.InMemory{}
	err := storer.AddBuyer(ctx, routing.Buyer{
		ID:          12345,
		CompanyCode: "local",
		InternalConfig: core.InternalConfig{
			LargeCustomer: true,
		},
	})
	assert.NoError(t, err)

	flushCount := 1000

	sessionID := rand.Uint64()
	userHash := rand.Uint64()
	nextSessionData := getTestSessionData(t, sessionID, userHash, true, false, flushTime)

	minutes := flushTime.Unix() / 60

	_, err = servers.topSessions.ZAdd(fmt.Sprintf("s-%d", minutes), nextSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", nextSessionData.Meta.ID))
	assert.NoError(t, err)

	_, err = servers.topSessions.ZAdd(fmt.Sprintf("sc-%016x-%d", nextSessionData.Meta.BuyerID, minutes), nextSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", nextSessionData.Meta.ID))
	assert.NoError(t, err)

	servers.sessionMap.HSet(fmt.Sprintf("n-%016x-%d", nextSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", nextSessionData.Meta.ID))

	err = servers.sessionMeta.Set(fmt.Sprintf("sm-%016x", nextSessionData.Meta.ID), "\""+nextSessionData.Meta.RedisString()+"\"")
	assert.NoError(t, err)

	_, err = servers.sessionSlices.RPush(fmt.Sprintf("ss-%016x", nextSessionData.Meta.ID), "\""+nextSessionData.Slice.RedisString()+"\"")
	assert.NoError(t, err)

	directSessionData := getTestSessionData(t, sessionID, userHash, false, true, flushTime)

	sessionDataBytes, err := directSessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(clients.topSessions, clients.sessionMap, clients.sessionMeta, clients.sessionSlices, messageChan, portalDataBuffer, flushTime, pingTime, storer, flushCount)
	assert.NoError(t, err)

	// Add a small delay so that the data has time to go over the tcp sockets
	time.Sleep(time.Millisecond * 50)

	{
		topSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, directSessionData.Meta.ID, sessionID)
	}

	{
		customerTopSessionIDs, err := servers.topSessions.ZMembers(fmt.Sprintf("sc-%016x-%d", directSessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err := strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, directSessionData.Meta.ID, sessionID)
	}

	{
		pointVal := servers.sessionMap.HGet(fmt.Sprintf("d-%016x-%d", directSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", directSessionData.Meta.ID))
		pointVal = pointVal[1 : len(pointVal)-1] // Remove the extra quotes
		assert.Equal(t, directSessionData.Point.RedisString(), pointVal)
	}

	{
		metaVal, err := servers.sessionMeta.Get(fmt.Sprintf("sm-%016x", directSessionData.Meta.ID))
		assert.NoError(t, err)

		metaVal = metaVal[1 : len(metaVal)-1] // Remove the extra quotes
		assert.Equal(t, directSessionData.Meta.RedisString(), metaVal)
	}

	{
		sliceVals, err := servers.sessionSlices.List(fmt.Sprintf("ss-%016x", directSessionData.Meta.ID))
		assert.NoError(t, err)
		assert.Len(t, sliceVals, 2)

		nextSliceVal := sliceVals[0]
		directSliceVal := sliceVals[1]

		nextSliceVal = nextSliceVal[1 : len(nextSliceVal)-1] // Remove the extra quotes
		assert.Equal(t, nextSessionData.Slice.RedisString(), nextSliceVal)

		directSliceVal = directSliceVal[1 : len(directSliceVal)-1] // Remove the extra quotes
		assert.Equal(t, directSessionData.Slice.RedisString(), directSliceVal)
	}
}
