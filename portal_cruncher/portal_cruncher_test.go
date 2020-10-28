package portalcruncher_test

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/alicebob/miniredis/v2/server"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	portalcruncher "github.com/networknext/backend/portal_cruncher"
)

func getTestSessionData(largeCustomer bool, sessionID uint64, userHash uint64, onNetworkNext bool, everOnNetworkNext bool, timestamp time.Time) transport.SessionPortalData {
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

type MockSubscriber struct {
	topic        pubsub.Topic
	sessionData  []byte
	bad          bool
	receiveCount int
}

func (mock *MockSubscriber) Subscribe(topic pubsub.Topic) error {
	return nil
}

func (mock *MockSubscriber) Unsubscribe(topic pubsub.Topic) error {
	return nil
}

func (mock *MockSubscriber) ReceiveMessage(ctx context.Context) (pubsub.Topic, <-chan []byte, error) {
	if mock.bad {
		return 0, nil, errors.New("bad data")
	}

	out := make(chan []byte)
	if mock.receiveCount == 0 {
		go func() {
			out <- mock.sessionData
		}()
	}

	mock.receiveCount++
	return mock.topic, out, nil
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

func TestNewPortalCruncher(t *testing.T) {
	redisTopSessions, err := miniredis.Run()
	assert.NoError(t, err)

	redisSessionMap, err := miniredis.Run()
	assert.NoError(t, err)

	redisSessionMeta, err := miniredis.Run()
	assert.NoError(t, err)

	redisSessionSlices, err := miniredis.Run()
	assert.NoError(t, err)

	t.Run("top sessions failure", func(t *testing.T) {
		portalCruncher, err := portalcruncher.NewPortalCruncher(nil, "", redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.Nil(t, portalCruncher)
		assert.Error(t, err)
	})

	t.Run("session map failure", func(t *testing.T) {
		portalCruncher, err := portalcruncher.NewPortalCruncher(nil, redisTopSessions.Addr(), "", redisSessionMeta.Addr(), redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.Nil(t, portalCruncher)
		assert.Error(t, err)
	})

	t.Run("session meta failure", func(t *testing.T) {
		portalCruncher, err := portalcruncher.NewPortalCruncher(nil, redisTopSessions.Addr(), redisSessionMap.Addr(), "", redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.Nil(t, portalCruncher)
		assert.Error(t, err)
	})

	t.Run("session slices failure", func(t *testing.T) {
		portalCruncher, err := portalcruncher.NewPortalCruncher(nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), "", 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.Nil(t, portalCruncher)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		portalCruncher, err := portalcruncher.NewPortalCruncher(nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.NotNil(t, portalCruncher)
		assert.NoError(t, err)
	})
}

func TestReceiveMessage(t *testing.T) {
	ctx := context.Background()

	redisTopSessions, err := miniredis.Run()
	assert.NoError(t, err)

	redisSessionMap, err := miniredis.Run()
	assert.NoError(t, err)

	redisSessionMeta, err := miniredis.Run()
	assert.NoError(t, err)

	redisSessionSlices, err := miniredis.Run()
	assert.NoError(t, err)

	t.Run("receive error", func(t *testing.T) {
		subscriber := &MockSubscriber{bad: true}

		portalCruncher, err := portalcruncher.NewPortalCruncher(subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.EqualError(t, err, "error receiving message: bad data")
	})

	// todo: count tests

	t.Run("portal data unmarshal failure", func(t *testing.T) {
		subscriber := &MockSubscriber{topic: pubsub.TopicPortalCruncherSessionData, sessionData: []byte("bad data")}

		portalCruncher, err := portalcruncher.NewPortalCruncher(subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.Contains(t, err.Error(), "could not unmarshal message: ")
	})

	t.Run("portal data channel full", func(t *testing.T) {
		sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), true, true, time.Now())
		sessionDataBytes, err := sessionData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &MockSubscriber{topic: pubsub.TopicPortalCruncherSessionData, sessionData: sessionDataBytes}

		portalCruncher, err := portalcruncher.NewPortalCruncher(subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.Equal(t, err, &portalcruncher.ErrChannelFull{})
	})

	t.Run("portal data success", func(t *testing.T) {
		sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), true, true, time.Now())
		sessionDataBytes, err := sessionData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &MockSubscriber{topic: pubsub.TopicPortalCruncherSessionData, sessionData: sessionDataBytes}

		portalCruncher, err := portalcruncher.NewPortalCruncher(subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 1, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.NoError(t, err)
	})

	t.Run("unknown message", func(t *testing.T) {
		sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), true, true, time.Now())
		sessionDataBytes, err := sessionData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &MockSubscriber{topic: 255, sessionData: sessionDataBytes}

		portalCruncher, err := portalcruncher.NewPortalCruncher(subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 1, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.Equal(t, err, &portalcruncher.ErrUnknownMessage{})
	})
}

func TestPingRedis(t *testing.T) {
	t.Run("top sessions failure", func(t *testing.T) {
		redisTopSessions, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionMap, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionMeta, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionSlices, err := miniredis.Run()
		assert.NoError(t, err)

		portalCruncher, err := portalcruncher.NewPortalCruncher(nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.NotNil(t, portalCruncher)
		assert.NoError(t, err)

		time.Sleep(time.Millisecond) // have to sleep here otherwise miniredis can deadlock from closing too quickly after starting
		redisTopSessions.Close()

		err = portalCruncher.PingRedis()
		assert.Error(t, err)
	})

	t.Run("session map failure", func(t *testing.T) {
		redisTopSessions, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionMap, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionMeta, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionSlices, err := miniredis.Run()
		assert.NoError(t, err)

		portalCruncher, err := portalcruncher.NewPortalCruncher(nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.NotNil(t, portalCruncher)
		assert.NoError(t, err)

		time.Sleep(time.Millisecond)
		redisSessionMap.Close()

		err = portalCruncher.PingRedis()
		assert.Error(t, err)
	})

	t.Run("session meta failure", func(t *testing.T) {
		redisTopSessions, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionMap, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionMeta, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionSlices, err := miniredis.Run()
		assert.NoError(t, err)

		portalCruncher, err := portalcruncher.NewPortalCruncher(nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.NotNil(t, portalCruncher)
		assert.NoError(t, err)

		time.Sleep(time.Millisecond)
		redisSessionMeta.Close()

		err = portalCruncher.PingRedis()
		assert.Error(t, err)
	})

	t.Run("session slices failure", func(t *testing.T) {
		redisTopSessions, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionMap, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionMeta, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionSlices, err := miniredis.Run()
		assert.NoError(t, err)

		portalCruncher, err := portalcruncher.NewPortalCruncher(nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.NotNil(t, portalCruncher)
		assert.NoError(t, err)

		time.Sleep(time.Millisecond)
		redisSessionSlices.Close()

		err = portalCruncher.PingRedis()
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		redisTopSessions, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionMap, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionMeta, err := miniredis.Run()
		assert.NoError(t, err)

		redisSessionSlices, err := miniredis.Run()
		assert.NoError(t, err)

		portalCruncher, err := portalcruncher.NewPortalCruncher(nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), 0, 0, &metrics.EmptyPortalCruncherMetrics)
		assert.NotNil(t, portalCruncher)
		assert.NoError(t, err)

		err = portalCruncher.PingRedis()
		assert.NoError(t, err)
	})
}

func TestDirectSession(t *testing.T) {
	ctx, ctxCancelFunc := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*100))
	defer ctxCancelFunc()

	sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), false, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	mockRedises := make([]*MockRedis, 4)
	for i := range mockRedises {
		mockRedises[i], err = NewMockRedis()
		assert.NoError(t, err)
	}

	subscriber := &MockSubscriber{topic: pubsub.TopicPortalCruncherSessionData, sessionData: sessionDataBytes}
	portalCruncher, err := portalcruncher.NewPortalCruncher(subscriber, mockRedises[0].db.Addr(), mockRedises[1].db.Addr(), mockRedises[2].db.Addr(), mockRedises[3].db.Addr(), 4, 0, &metrics.EmptyPortalCruncherMetrics)
	err = portalCruncher.PingRedis()
	assert.NoError(t, err)

	err = portalCruncher.Start(ctx, 1, 1)
	assert.EqualError(t, err, "context deadline exceeded")

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
	ctx, ctxCancelFunc := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*100))
	defer ctxCancelFunc()

	sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), true, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	mockRedises := make([]*MockRedis, 4)
	for i := range mockRedises {
		mockRedises[i], err = NewMockRedis()
		assert.NoError(t, err)
	}

	subscriber := &MockSubscriber{topic: pubsub.TopicPortalCruncherSessionData, sessionData: sessionDataBytes}
	portalCruncher, err := portalcruncher.NewPortalCruncher(subscriber, mockRedises[0].db.Addr(), mockRedises[1].db.Addr(), mockRedises[2].db.Addr(), mockRedises[3].db.Addr(), 4, 0, &metrics.EmptyPortalCruncherMetrics)
	err = portalCruncher.PingRedis()
	assert.NoError(t, err)

	err = portalCruncher.Start(ctx, 1, 1)
	assert.EqualError(t, err, "context deadline exceeded")

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

func TestNextSessionLargeCustomer(t *testing.T) {
	ctx, ctxCancelFunc := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*100))
	defer ctxCancelFunc()

	sessionData := getTestSessionData(true, rand.Uint64(), rand.Uint64(), true, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	mockRedises := make([]*MockRedis, 4)
	for i := range mockRedises {
		mockRedises[i], err = NewMockRedis()
		assert.NoError(t, err)
	}

	subscriber := &MockSubscriber{topic: pubsub.TopicPortalCruncherSessionData, sessionData: sessionDataBytes}
	portalCruncher, err := portalcruncher.NewPortalCruncher(subscriber, mockRedises[0].db.Addr(), mockRedises[1].db.Addr(), mockRedises[2].db.Addr(), mockRedises[3].db.Addr(), 4, 0, &metrics.EmptyPortalCruncherMetrics)
	err = portalCruncher.PingRedis()
	assert.NoError(t, err)

	err = portalCruncher.Start(ctx, 1, 1)
	assert.EqualError(t, err, "context deadline exceeded")

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
	ctx, ctxCancelFunc := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*100))
	defer ctxCancelFunc()

	sessionID := rand.Uint64()
	userHash := rand.Uint64()

	flushTime := time.Now().Add(-time.Second * 10)

	minutes := time.Now().Unix() / 60

	var err error
	mockRedises := make([]*MockRedis, 4)
	for i := range mockRedises {
		mockRedises[i], err = NewMockRedis()
		assert.NoError(t, err)
	}

	directSessionData := getTestSessionData(true, sessionID, userHash, false, false, flushTime)

	_, err = mockRedises[0].db.ZAdd(fmt.Sprintf("s-%d", minutes), directSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", directSessionData.Meta.ID))
	assert.NoError(t, err)

	_, err = mockRedises[0].db.ZAdd(fmt.Sprintf("sc-%016x-%d", directSessionData.Meta.BuyerID, minutes), directSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", directSessionData.Meta.ID))
	assert.NoError(t, err)

	mockRedises[1].db.HSet(fmt.Sprintf("d-%016x-%d", directSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", directSessionData.Meta.ID))

	err = mockRedises[2].db.Set(fmt.Sprintf("sm-%016x", directSessionData.Meta.ID), directSessionData.Meta.RedisString())
	assert.NoError(t, err)

	_, err = mockRedises[3].db.RPush(fmt.Sprintf("ss-%016x", directSessionData.Meta.ID), directSessionData.Slice.RedisString())
	assert.NoError(t, err)

	nextSessionData := getTestSessionData(true, sessionID, userHash, true, false, flushTime)
	sessionDataBytes, err := nextSessionData.MarshalBinary()
	assert.NoError(t, err)

	subscriber := &MockSubscriber{topic: pubsub.TopicPortalCruncherSessionData, sessionData: sessionDataBytes}
	portalCruncher, err := portalcruncher.NewPortalCruncher(subscriber, mockRedises[0].db.Addr(), mockRedises[1].db.Addr(), mockRedises[2].db.Addr(), mockRedises[3].db.Addr(), 4, 0, &metrics.EmptyPortalCruncherMetrics)
	err = portalCruncher.PingRedis()
	assert.NoError(t, err)

	err = portalCruncher.Start(ctx, 1, 1)
	assert.EqualError(t, err, "context deadline exceeded")

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
	ctx, ctxCancelFunc := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*100))
	defer ctxCancelFunc()

	sessionID := rand.Uint64()
	userHash := rand.Uint64()

	flushTime := time.Now().Add(-time.Second * 10)

	minutes := time.Now().Unix() / 60

	var err error
	mockRedises := make([]*MockRedis, 4)
	for i := range mockRedises {
		mockRedises[i], err = NewMockRedis()
		assert.NoError(t, err)
	}

	nextSessionData := getTestSessionData(true, sessionID, userHash, true, false, flushTime)

	_, err = mockRedises[0].db.ZAdd(fmt.Sprintf("s-%d", minutes), nextSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", nextSessionData.Meta.ID))
	assert.NoError(t, err)

	_, err = mockRedises[0].db.ZAdd(fmt.Sprintf("sc-%016x-%d", nextSessionData.Meta.BuyerID, minutes), nextSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", nextSessionData.Meta.ID))
	assert.NoError(t, err)

	mockRedises[1].db.HSet(fmt.Sprintf("n-%016x-%d", nextSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", nextSessionData.Meta.ID))

	err = mockRedises[2].db.Set(fmt.Sprintf("sm-%016x", nextSessionData.Meta.ID), nextSessionData.Meta.RedisString())
	assert.NoError(t, err)

	_, err = mockRedises[3].db.RPush(fmt.Sprintf("ss-%016x", nextSessionData.Meta.ID), nextSessionData.Slice.RedisString())
	assert.NoError(t, err)

	directSessionData := getTestSessionData(true, sessionID, userHash, false, true, flushTime)
	sessionDataBytes, err := directSessionData.MarshalBinary()
	assert.NoError(t, err)

	subscriber := &MockSubscriber{topic: pubsub.TopicPortalCruncherSessionData, sessionData: sessionDataBytes}
	portalCruncher, err := portalcruncher.NewPortalCruncher(subscriber, mockRedises[0].db.Addr(), mockRedises[1].db.Addr(), mockRedises[2].db.Addr(), mockRedises[3].db.Addr(), 4, 0, &metrics.EmptyPortalCruncherMetrics)
	err = portalCruncher.PingRedis()
	assert.NoError(t, err)

	err = portalCruncher.Start(ctx, 1, 1)
	assert.EqualError(t, err, "context deadline exceeded")

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
