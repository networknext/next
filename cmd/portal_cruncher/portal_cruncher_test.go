package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/alicebob/miniredis/v2/server"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
	"github.com/stretchr/testify/assert"
)

func getTestSessionData(t *testing.T, largeCustomer bool, sessionID uint64, userHash uint64, onNetworkNext bool, everOnNetworkNext bool, timestamp time.Time) transport.SessionPortalData {
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
		LargeCustomer: largeCustomer,
		EverOnNext:    everOnNetworkNext,
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

type MockRedis struct {
	db *miniredis.Miniredis
}

func NewMockRedis() (*MockRedis, error) {
	db, err := miniredis.Run()
	if err != nil {
		return nil, err
	}

	return &MockRedis{
		db: db,
	}, nil
}

func (m *MockRedis) Ping() error {
	var replyBuffer bytes.Buffer
	w := bufio.NewWriter(&replyBuffer)

	peer := server.NewPeer(w)
	m.db.Server().Dispatch(peer, []string{"PING"})
	peer.Flush()

	reader := bufio.NewReader(&replyBuffer)
	reader.ReadString('\n')

	return nil
}

func (m *MockRedis) Command(command string, format string, args ...interface{}) error {
	cmdArgsString := fmt.Sprintf(format, args...)
	var cmdArgs []string

	if cmdArgsString != "" {
		var err error

		// Split the args string so that we can allow for args with spaces
		reader := csv.NewReader(strings.NewReader(cmdArgsString))
		reader.Comma = ' '
		cmdArgs, err = reader.Read()
		if err != nil {
			return fmt.Errorf("failed to split command args: %v", err)
		}
	}

	cmdArgs = append([]string{command}, cmdArgs...)

	var replyBuffer bytes.Buffer
	w := bufio.NewWriter(&replyBuffer)

	peer := server.NewPeer(w)
	m.db.Server().Dispatch(peer, cmdArgs)
	peer.Flush()

	return nil
}

func (m *MockRedis) Close() error {
	return nil
}

func setupMockRedises(t *testing.T) []*MockRedis {
	mockRedises := make([]*MockRedis, 4)

	var err error
	for i := range mockRedises {
		mockRedises[i], err = NewMockRedis()
		assert.NoError(t, err)
	}

	err = pingRedis(mockRedises[0], mockRedises[1], mockRedises[2], mockRedises[3])
	assert.NoError(t, err)

	return mockRedises
}

// Use this function to get a free port so we can run tests while also running the happy path
func getFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "0", err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "0", err
	}
	defer l.Close()
	return fmt.Sprintf("%d", l.Addr().(*net.TCPAddr).Port), nil
}

func TestReceive(t *testing.T) {
	port, err := getFreePort()
	assert.NoError(t, err)

	portalSubscriber, err := pubsub.NewPortalCruncherSubscriber(port, 1000)
	assert.NoError(t, err)

	err = portalSubscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)
	assert.NoError(t, err)

	portalPublisher, err := pubsub.NewPortalCruncherPublisher("tcp://127.0.0.1:"+port, 1000)
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
	expected := getTestSessionData(t, false, rand.Uint64(), rand.Uint64(), true, false, time.Now())
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
	mockRedises := setupMockRedises(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	flushCount := 1000

	sessionData := getTestSessionData(t, false, rand.Uint64(), rand.Uint64(), false, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(mockRedises[0], mockRedises[1], mockRedises[2], mockRedises[3], messageChan, portalDataBuffer, flushTime, pingTime, flushCount)
	assert.NoError(t, err)

	minutes := time.Now().Unix() / 60

	{
		topSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		customerTopSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("sc-%016x-%d", sessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err := strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		pointVal := mockRedises[1].db.HGet(fmt.Sprintf("d-%016x-%d", sessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", sessionData.Meta.ID))
		assert.Equal(t, sessionData.Point.RedisString(), pointVal)
	}

	{
		metaVal, err := mockRedises[2].db.Get(fmt.Sprintf("sm-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.RedisString(), metaVal)
	}

	{
		sliceVals, err := mockRedises[3].db.List(fmt.Sprintf("ss-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)
		assert.Len(t, sliceVals, 1)

		sliceVal := sliceVals[0]

		assert.Equal(t, sessionData.Slice.RedisString(), sliceVal)
	}
}

func TestNextSession(t *testing.T) {
	mockRedises := setupMockRedises(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	flushCount := 1000

	sessionData := getTestSessionData(t, false, rand.Uint64(), rand.Uint64(), true, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(mockRedises[0], mockRedises[1], mockRedises[2], mockRedises[3], messageChan, portalDataBuffer, flushTime, pingTime, flushCount)
	assert.NoError(t, err)

	minutes := time.Now().Unix() / 60

	{
		topSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		customerTopSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("sc-%016x-%d", sessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err := strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		pointVal := mockRedises[1].db.HGet(fmt.Sprintf("n-%016x-%d", sessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", sessionData.Meta.ID))
		assert.Equal(t, sessionData.Point.RedisString(), pointVal)
	}

	{
		metaVal, err := mockRedises[2].db.Get(fmt.Sprintf("sm-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.RedisString(), metaVal)
	}

	{
		sliceVals, err := mockRedises[3].db.List(fmt.Sprintf("ss-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)
		assert.Len(t, sliceVals, 1)

		sliceVal := sliceVals[0]

		assert.Equal(t, sessionData.Slice.RedisString(), sliceVal)
	}
}

func TestDirectSessionLargeCustomer(t *testing.T) {
	mockRedises := setupMockRedises(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	flushCount := 1000

	sessionData := getTestSessionData(t, true, rand.Uint64(), rand.Uint64(), false, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(mockRedises[0], mockRedises[1], mockRedises[2], mockRedises[3], messageChan, portalDataBuffer, flushTime, pingTime, flushCount)
	assert.NoError(t, err)

	minutes := time.Now().Unix() / 60

	{
		topSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.Error(t, err)
		assert.Len(t, topSessionIDs, 0)
	}

	{
		customerTopSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("sc-%016x-%d", sessionData.Meta.BuyerID, minutes))
		assert.Error(t, err)
		assert.Len(t, customerTopSessionIDs, 0)
	}

	{
		pointVal := mockRedises[1].db.HGet(fmt.Sprintf("d-%016x-%d", sessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", sessionData.Meta.ID))
		assert.Empty(t, pointVal)
	}

	{
		metaVal, err := mockRedises[2].db.Get(fmt.Sprintf("sm-%016x", sessionData.Meta.ID))
		assert.Error(t, err)
		assert.Empty(t, metaVal)
	}

	{
		sliceVals, err := mockRedises[3].db.List(fmt.Sprintf("ss-%016x", sessionData.Meta.ID))
		assert.Error(t, err)
		assert.Len(t, sliceVals, 0)
	}
}

func TestNextSessionLargeCustomer(t *testing.T) {
	mockRedises := setupMockRedises(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	flushCount := 1000

	sessionData := getTestSessionData(t, true, rand.Uint64(), rand.Uint64(), true, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(mockRedises[0], mockRedises[1], mockRedises[2], mockRedises[3], messageChan, portalDataBuffer, flushTime, pingTime, flushCount)
	assert.NoError(t, err)

	minutes := time.Now().Unix() / 60

	{
		topSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		customerTopSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("sc-%016x-%d", sessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err := strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		pointVal := mockRedises[1].db.HGet(fmt.Sprintf("n-%016x-%d", sessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", sessionData.Meta.ID))
		assert.Equal(t, sessionData.Point.RedisString(), pointVal)
	}

	{
		metaVal, err := mockRedises[2].db.Get(fmt.Sprintf("sm-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.RedisString(), metaVal)
	}

	{
		sliceVals, err := mockRedises[3].db.List(fmt.Sprintf("ss-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)
		assert.Len(t, sliceVals, 1)

		sliceVal := sliceVals[0]

		assert.Equal(t, sessionData.Slice.RedisString(), sliceVal)
	}
}

func TestDirectToNextLargeCustomer(t *testing.T) {
	mockRedises := setupMockRedises(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	flushCount := 1000

	sessionID := rand.Uint64()
	userHash := rand.Uint64()
	directSessionData := getTestSessionData(t, true, sessionID, userHash, false, false, flushTime)

	minutes := time.Now().Unix() / 60

	_, err := mockRedises[0].db.ZAdd(fmt.Sprintf("s-%d", minutes), directSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", directSessionData.Meta.ID))
	assert.NoError(t, err)

	_, err = mockRedises[0].db.ZAdd(fmt.Sprintf("sc-%016x-%d", directSessionData.Meta.BuyerID, minutes), directSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", directSessionData.Meta.ID))
	assert.NoError(t, err)

	mockRedises[1].db.HSet(fmt.Sprintf("d-%016x-%d", directSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", directSessionData.Meta.ID))

	err = mockRedises[2].db.Set(fmt.Sprintf("sm-%016x", directSessionData.Meta.ID), directSessionData.Meta.RedisString())
	assert.NoError(t, err)

	_, err = mockRedises[3].db.RPush(fmt.Sprintf("ss-%016x", directSessionData.Meta.ID), directSessionData.Slice.RedisString())
	assert.NoError(t, err)

	nextSessionData := getTestSessionData(t, true, sessionID, userHash, true, false, flushTime)

	sessionDataBytes, err := nextSessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(mockRedises[0], mockRedises[1], mockRedises[2], mockRedises[3], messageChan, portalDataBuffer, flushTime, pingTime, flushCount)
	assert.NoError(t, err)

	{
		topSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, nextSessionData.Meta.ID, sessionID)
	}

	{
		customerTopSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("sc-%016x-%d", nextSessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err := strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, nextSessionData.Meta.ID, sessionID)
	}

	{
		pointVal := mockRedises[1].db.HGet(fmt.Sprintf("n-%016x-%d", nextSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", nextSessionData.Meta.ID))
		assert.Equal(t, nextSessionData.Point.RedisString(), pointVal)
	}

	{
		metaVal, err := mockRedises[2].db.Get(fmt.Sprintf("sm-%016x", nextSessionData.Meta.ID))
		assert.NoError(t, err)

		assert.Equal(t, nextSessionData.Meta.RedisString(), metaVal)
	}

	{
		sliceVals, err := mockRedises[3].db.List(fmt.Sprintf("ss-%016x", nextSessionData.Meta.ID))
		assert.NoError(t, err)
		assert.Len(t, sliceVals, 2)

		directSliceVal := sliceVals[0]
		nextSliceVal := sliceVals[1]

		assert.Equal(t, directSessionData.Slice.RedisString(), directSliceVal)
		assert.Equal(t, nextSessionData.Slice.RedisString(), nextSliceVal)
	}
}

func TestNextToDirectLargeCustomer(t *testing.T) {
	mockRedises := setupMockRedises(t)

	portalDataBuffer := make([]transport.SessionPortalData, 0)

	flushTime := time.Now().Add(-time.Second * 10)
	pingTime := time.Now().Add(-time.Second * 10)

	flushCount := 1000

	sessionID := rand.Uint64()
	userHash := rand.Uint64()
	nextSessionData := getTestSessionData(t, true, sessionID, userHash, true, false, flushTime)

	minutes := time.Now().Unix() / 60

	_, err := mockRedises[0].db.ZAdd(fmt.Sprintf("s-%d", minutes), nextSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", nextSessionData.Meta.ID))
	assert.NoError(t, err)

	_, err = mockRedises[0].db.ZAdd(fmt.Sprintf("sc-%016x-%d", nextSessionData.Meta.BuyerID, minutes), nextSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", nextSessionData.Meta.ID))
	assert.NoError(t, err)

	mockRedises[1].db.HSet(fmt.Sprintf("n-%016x-%d", nextSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", nextSessionData.Meta.ID))

	err = mockRedises[2].db.Set(fmt.Sprintf("sm-%016x", nextSessionData.Meta.ID), nextSessionData.Meta.RedisString())
	assert.NoError(t, err)

	_, err = mockRedises[3].db.RPush(fmt.Sprintf("ss-%016x", nextSessionData.Meta.ID), nextSessionData.Slice.RedisString())
	assert.NoError(t, err)

	directSessionData := getTestSessionData(t, true, sessionID, userHash, false, true, flushTime)

	sessionDataBytes, err := directSessionData.MarshalBinary()
	assert.NoError(t, err)

	messageChan := make(chan []byte, 1)
	messageChan <- sessionDataBytes

	err = RedisHandler(mockRedises[0], mockRedises[1], mockRedises[2], mockRedises[3], messageChan, portalDataBuffer, flushTime, pingTime, flushCount)
	assert.NoError(t, err)

	{
		topSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, directSessionData.Meta.ID, sessionID)
	}

	{
		customerTopSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("sc-%016x-%d", directSessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err := strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, directSessionData.Meta.ID, sessionID)
	}

	{
		pointVal := mockRedises[1].db.HGet(fmt.Sprintf("d-%016x-%d", directSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", directSessionData.Meta.ID))
		assert.Equal(t, directSessionData.Point.RedisString(), pointVal)
	}

	{
		metaVal, err := mockRedises[2].db.Get(fmt.Sprintf("sm-%016x", directSessionData.Meta.ID))
		assert.NoError(t, err)

		assert.Equal(t, directSessionData.Meta.RedisString(), metaVal)
	}

	{
		sliceVals, err := mockRedises[3].db.List(fmt.Sprintf("ss-%016x", directSessionData.Meta.ID))
		assert.NoError(t, err)
		assert.Len(t, sliceVals, 2)

		nextSliceVal := sliceVals[0]
		directSliceVal := sliceVals[1]

		assert.Equal(t, nextSessionData.Slice.RedisString(), nextSliceVal)
		assert.Equal(t, directSessionData.Slice.RedisString(), directSliceVal)
	}
}
