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
	"os"

	"github.com/alicebob/miniredis/v2"
	"github.com/alicebob/miniredis/v2/server"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	portalcruncher "github.com/networknext/backend/portal_cruncher"
)

func getTestCountData(serverID uint64, buyerID uint64) transport.SessionCountData {
	return transport.SessionCountData{
		ServerID:    serverID,
		BuyerID:     buyerID,
		NumSessions: rand.Uint32(),
	}
}

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
			NextRTT:         50,
			DirectRTT:       20,
			DeltaRTT:        -30,
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
	topics     []pubsub.Topic
	topicIndex int

	countData   []byte
	sessionData []byte

	bad bool

	receiveCount int
}

func (mock *MockSubscriber) Subscribe(topic pubsub.Topic) error {
	mock.topics = append(mock.topics, topic)
	return nil
}

func (mock *MockSubscriber) Unsubscribe(topic pubsub.Topic) error {
	for i, t := range mock.topics {
		if t == topic {
			mock.topics = append(mock.topics[:i], mock.topics[i+1:]...)
			return nil
		}
	}

	return nil
}

func (mock *MockSubscriber) ReceiveMessage(ctx context.Context) (pubsub.Topic, <-chan []byte, error) {
	if mock.bad {
		return 0, nil, errors.New("bad data")
	}

	if mock.receiveCount >= 2 {
		return 0, nil, nil
	}

	mock.receiveCount++

	topic := mock.topics[mock.topicIndex]
	mock.topicIndex = (mock.topicIndex + 1) % len(mock.topics)

	out := make(chan []byte)

	switch topic {
	case pubsub.TopicPortalCruncherSessionCounts:
		go func() {
			out <- mock.countData
		}()

		return topic, out, nil
	case pubsub.TopicPortalCruncherSessionData:
		go func() {
			out <- mock.sessionData
		}()

		return topic, out, nil

	default:
		go func() {
			out <- []byte("bad data")
		}()
		return topic, out, nil
	}
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

func checkBigtableEmulation() bool {
	bigtableEmulatorHost := os.Getenv("BIGTABLE_EMULATOR_HOST")
	return bigtableEmulatorHost != ""
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

	var useBigtable bool
	{
		bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
		assert.NoError(t, err)

		bigtableEmulation := checkBigtableEmulation()

		useBigtable = bigtableEnabled && bigtableEmulation
	}

	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
	btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
	btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
	
	btMaxAgeDays := 1

	ctx := context.Background()
	logger := log.NewNopLogger()

	t.Run("top sessions failure", func(t *testing.T) {
		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, "", redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.Nil(t, portalCruncher)
		assert.Error(t, err)
	})

	t.Run("session map failure", func(t *testing.T) {
		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisTopSessions.Addr(), "", redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.Nil(t, portalCruncher)
		assert.Error(t, err)
	})

	t.Run("session meta failure", func(t *testing.T) {
		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisTopSessions.Addr(), redisSessionMap.Addr(), "", redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.Nil(t, portalCruncher)
		assert.Error(t, err)
	})

	t.Run("session slices failure", func(t *testing.T) {
		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), "", useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.Nil(t, portalCruncher)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
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

	var useBigtable bool
	{
		bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
		assert.NoError(t, err)

		bigtableEmulation := checkBigtableEmulation()

		useBigtable = bigtableEnabled && bigtableEmulation
	}

	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
	btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
	btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
	
	btMaxAgeDays := 1

	logger := log.NewNopLogger()

	t.Run("receive error", func(t *testing.T) {
		subscriber := &MockSubscriber{bad: true}

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.EqualError(t, err, "error receiving message: bad data")
	})

	t.Run("count data unmarshal failure", func(t *testing.T) {
		subscriber := &MockSubscriber{countData: []byte("bad data")}
		subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.Contains(t, err.Error(), "could not unmarshal message: ")
	})

	t.Run("count data channel full", func(t *testing.T) {
		countData := getTestCountData(rand.Uint64(), rand.Uint64())
		countDataBytes, err := countData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &MockSubscriber{countData: countDataBytes}
		subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.Equal(t, err, &portalcruncher.ErrChannelFull{})
	})

	t.Run("count data success", func(t *testing.T) {
		countData := getTestCountData(rand.Uint64(), rand.Uint64())
		countDataBytes, err := countData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &MockSubscriber{countData: countDataBytes}
		subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 1, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.NoError(t, err)
	})

	t.Run("portal data unmarshal failure", func(t *testing.T) {
		subscriber := &MockSubscriber{sessionData: []byte("bad data")}
		subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.Contains(t, err.Error(), "could not unmarshal message: ")
	})

	t.Run("portal data channel full", func(t *testing.T) {
		sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), true, true, time.Now())
		sessionDataBytes, err := sessionData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &MockSubscriber{sessionData: sessionDataBytes}
		subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.Equal(t, err, &portalcruncher.ErrChannelFull{})
	})

	t.Run("portal data success", func(t *testing.T) {
		sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), true, true, time.Now())
		sessionDataBytes, err := sessionData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &MockSubscriber{sessionData: sessionDataBytes}
		subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 1, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.NoError(t, err)
	})

	t.Run("unknown message", func(t *testing.T) {
		sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), true, true, time.Now())
		sessionDataBytes, err := sessionData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &MockSubscriber{sessionData: sessionDataBytes}
		subscriber.Subscribe(255)

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 1, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NoError(t, err)

		err = portalCruncher.ReceiveMessage(ctx)
		assert.Equal(t, &portalcruncher.ErrUnknownMessage{}, err)
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

		var useBigtable bool
		{
			bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
			assert.NoError(t, err)

			bigtableEmulation := checkBigtableEmulation()

			useBigtable = bigtableEnabled && bigtableEmulation
		}

		gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
		btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
		btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
		btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
		
		btMaxAgeDays := 1

		ctx := context.Background()
		logger := log.NewNopLogger()

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NotNil(t, portalCruncher)
		assert.NoError(t, err)

		time.Sleep(time.Millisecond * 100) // have to sleep here otherwise miniredis can deadlock from closing too quickly after starting
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

		var useBigtable bool
		{
			bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
			assert.NoError(t, err)

			bigtableEmulation := checkBigtableEmulation()

			useBigtable = bigtableEnabled && bigtableEmulation
		}

		gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
		btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
		btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
		btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
		
		btMaxAgeDays := 1

		ctx := context.Background()
		logger := log.NewNopLogger()

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NotNil(t, portalCruncher)
		assert.NoError(t, err)

		time.Sleep(time.Millisecond * 100)
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

		var useBigtable bool
		{
			bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
			assert.NoError(t, err)

			bigtableEmulation := checkBigtableEmulation()

			useBigtable = bigtableEnabled && bigtableEmulation
		}

		gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
		btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
		btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
		btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
		
		btMaxAgeDays := 1

		ctx := context.Background()
		logger := log.NewNopLogger()

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NotNil(t, portalCruncher)
		assert.NoError(t, err)

		time.Sleep(time.Millisecond * 100)
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

		var useBigtable bool
		{
			bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
			assert.NoError(t, err)

			bigtableEmulation := checkBigtableEmulation()

			useBigtable = bigtableEnabled && bigtableEmulation
		}

		gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
		btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
		btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
		btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
		
		btMaxAgeDays := 1

		ctx := context.Background()
		logger := log.NewNopLogger()

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NotNil(t, portalCruncher)
		assert.NoError(t, err)

		time.Sleep(time.Millisecond * 100)
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

		var useBigtable bool
		{
			bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
			assert.NoError(t, err)

			bigtableEmulation := checkBigtableEmulation()

			useBigtable = bigtableEnabled && bigtableEmulation
		}

		gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
		btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
		btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
		btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
		
		btMaxAgeDays := 1

		ctx := context.Background()
		logger := log.NewNopLogger()

		portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisTopSessions.Addr(), redisSessionMap.Addr(), redisSessionMeta.Addr(), redisSessionSlices.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 0, 0, logger, &metrics.EmptyPortalCruncherMetrics)
		assert.NotNil(t, portalCruncher)
		assert.NoError(t, err)

		err = portalCruncher.PingRedis()
		assert.NoError(t, err)
	})
}

func TestDirectSession(t *testing.T) {
	ctx, ctxCancelFunc := context.WithDeadline(context.Background(), time.Now().Add(time.Millisecond*100))
	defer ctxCancelFunc()

	countData := getTestCountData(rand.Uint64(), rand.Uint64())
	countDataBytes, err := countData.MarshalBinary()
	assert.NoError(t, err)

	sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), false, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	mockRedises := make([]*MockRedis, 4)
	for i := range mockRedises {
		mockRedises[i], err = NewMockRedis()
		assert.NoError(t, err)
	}

	subscriber := &MockSubscriber{countData: countDataBytes, sessionData: sessionDataBytes}
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)

	var useBigtable bool
	{
		bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
		assert.NoError(t, err)

		bigtableEmulation := checkBigtableEmulation()

		useBigtable = bigtableEnabled && bigtableEmulation
	}

	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
	btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
	btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
	
	btMaxAgeDays := 1

	logger := log.NewNopLogger()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, mockRedises[0].db.Addr(), mockRedises[1].db.Addr(), mockRedises[2].db.Addr(), mockRedises[3].db.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 4, 0, logger, &metrics.EmptyPortalCruncherMetrics)
	err = portalCruncher.PingRedis()
	assert.NoError(t, err)

	err = portalCruncher.Start(ctx, 1, 1, 1)
	assert.EqualError(t, err, "context deadline exceeded")

	minutes := time.Now().Unix() / 60

	{
		assert.Len(t, mockRedises[0].db.Keys(), 2)

		topSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)

		customerTopSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("sc-%016x-%d", sessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err = strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		assert.Len(t, mockRedises[1].db.Keys(), 2)

		pointVal := mockRedises[1].db.HGet(fmt.Sprintf("d-%016x-%d", sessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", sessionData.Meta.ID))
		assert.Equal(t, sessionData.Point.RedisString(), pointVal)

		fields, err := mockRedises[1].db.HKeys(fmt.Sprintf("c-%016x-%d", countData.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, fmt.Sprintf("%016x", countData.ServerID), fields[0])

		countVal := mockRedises[1].db.HGet(fmt.Sprintf("c-%016x-%d", countData.BuyerID, minutes), fmt.Sprintf("%016x", countData.ServerID))
		assert.Equal(t, fmt.Sprintf("%d", countData.NumSessions), countVal)
	}

	{
		assert.Len(t, mockRedises[2].db.Keys(), 1)

		metaVal, err := mockRedises[2].db.Get(fmt.Sprintf("sm-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.RedisString(), metaVal)
	}

	{
		assert.Len(t, mockRedises[3].db.Keys(), 1)

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

	countData := getTestCountData(rand.Uint64(), rand.Uint64())
	countDataBytes, err := countData.MarshalBinary()
	assert.NoError(t, err)

	sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), true, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	mockRedises := make([]*MockRedis, 4)
	for i := range mockRedises {
		mockRedises[i], err = NewMockRedis()
		assert.NoError(t, err)
	}

	subscriber := &MockSubscriber{countData: countDataBytes, sessionData: sessionDataBytes}
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)

	var useBigtable bool
	{
		bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
		assert.NoError(t, err)

		bigtableEmulation := checkBigtableEmulation()

		useBigtable = bigtableEnabled && bigtableEmulation
	}

	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
	btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
	btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
	
	btMaxAgeDays := 1

	logger := log.NewNopLogger()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, mockRedises[0].db.Addr(), mockRedises[1].db.Addr(), mockRedises[2].db.Addr(), mockRedises[3].db.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 4, 0, logger, &metrics.EmptyPortalCruncherMetrics)
	err = portalCruncher.PingRedis()
	assert.NoError(t, err)

	err = portalCruncher.Start(ctx, 1, 1, 1)
	assert.EqualError(t, err, "context deadline exceeded")

	minutes := time.Now().Unix() / 60

	{
		assert.Len(t, mockRedises[0].db.Keys(), 2)

		topSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)

		customerTopSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("sc-%016x-%d", sessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err = strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.ID, sessionID)
	}

	{
		assert.Len(t, mockRedises[1].db.Keys(), 2)

		pointVal := mockRedises[1].db.HGet(fmt.Sprintf("n-%016x-%d", sessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", sessionData.Meta.ID))
		assert.Equal(t, sessionData.Point.RedisString(), pointVal)

		fields, err := mockRedises[1].db.HKeys(fmt.Sprintf("c-%016x-%d", countData.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, fmt.Sprintf("%016x", countData.ServerID), fields[0])

		countVal := mockRedises[1].db.HGet(fmt.Sprintf("c-%016x-%d", countData.BuyerID, minutes), fmt.Sprintf("%016x", countData.ServerID))
		assert.Equal(t, fmt.Sprintf("%d", countData.NumSessions), countVal)
	}

	{
		assert.Len(t, mockRedises[2].db.Keys(), 1)

		metaVal, err := mockRedises[2].db.Get(fmt.Sprintf("sm-%016x", sessionData.Meta.ID))
		assert.NoError(t, err)

		assert.Equal(t, sessionData.Meta.RedisString(), metaVal)
	}

	{
		assert.Len(t, mockRedises[3].db.Keys(), 1)

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

	countData := getTestCountData(rand.Uint64(), rand.Uint64())
	countDataBytes, err := countData.MarshalBinary()
	assert.NoError(t, err)

	sessionData := getTestSessionData(true, rand.Uint64(), rand.Uint64(), false, false, time.Now())
	sessionDataBytes, err := sessionData.MarshalBinary()
	assert.NoError(t, err)

	mockRedises := make([]*MockRedis, 4)
	for i := range mockRedises {
		mockRedises[i], err = NewMockRedis()
		assert.NoError(t, err)
	}

	subscriber := &MockSubscriber{countData: countDataBytes, sessionData: sessionDataBytes}
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)

	var useBigtable bool
	{
		bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
		assert.NoError(t, err)

		bigtableEmulation := checkBigtableEmulation()

		useBigtable = bigtableEnabled && bigtableEmulation
	}

	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
	btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
	btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
	
	btMaxAgeDays := 1

	logger := log.NewNopLogger()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, mockRedises[0].db.Addr(), mockRedises[1].db.Addr(), mockRedises[2].db.Addr(), mockRedises[3].db.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 4, 0, logger, &metrics.EmptyPortalCruncherMetrics)
	err = portalCruncher.PingRedis()
	assert.NoError(t, err)

	err = portalCruncher.Start(ctx, 1, 1, 1)
	assert.EqualError(t, err, "context deadline exceeded")

	minutes := time.Now().Unix() / 60

	{
		assert.Empty(t, mockRedises[0].db.Keys())
	}

	{
		assert.Len(t, mockRedises[1].db.Keys(), 1)

		fields, err := mockRedises[1].db.HKeys(fmt.Sprintf("c-%016x-%d", countData.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, fmt.Sprintf("%016x", countData.ServerID), fields[0])

		countVal := mockRedises[1].db.HGet(fmt.Sprintf("c-%016x-%d", countData.BuyerID, minutes), fmt.Sprintf("%016x", countData.ServerID))
		assert.Equal(t, fmt.Sprintf("%d", countData.NumSessions), countVal)
	}

	{
		assert.Empty(t, mockRedises[2].db.Keys())
	}

	{
		assert.Empty(t, mockRedises[3].db.Keys())
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

	serverID := rand.Uint64()
	buyerID := rand.Uint64()
	oldCountData := getTestCountData(serverID, buyerID)

	directSessionData := getTestSessionData(true, sessionID, userHash, false, false, flushTime)

	_, err = mockRedises[0].db.ZAdd(fmt.Sprintf("s-%d", minutes), directSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", directSessionData.Meta.ID))
	assert.NoError(t, err)

	_, err = mockRedises[0].db.ZAdd(fmt.Sprintf("sc-%016x-%d", directSessionData.Meta.BuyerID, minutes), directSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", directSessionData.Meta.ID))
	assert.NoError(t, err)

	mockRedises[1].db.HSet(fmt.Sprintf("d-%016x-%d", directSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", directSessionData.Meta.ID))

	mockRedises[1].db.HSet(fmt.Sprintf("c-%016x-%d", oldCountData.BuyerID, minutes), fmt.Sprintf("%016x", oldCountData.ServerID))

	err = mockRedises[2].db.Set(fmt.Sprintf("sm-%016x", directSessionData.Meta.ID), directSessionData.Meta.RedisString())
	assert.NoError(t, err)

	_, err = mockRedises[3].db.RPush(fmt.Sprintf("ss-%016x", directSessionData.Meta.ID), directSessionData.Slice.RedisString())
	assert.NoError(t, err)

	newCountData := getTestCountData(serverID, buyerID)
	countDataBytes, err := newCountData.MarshalBinary()
	assert.NoError(t, err)

	nextSessionData := getTestSessionData(true, sessionID, userHash, true, false, flushTime)
	sessionDataBytes, err := nextSessionData.MarshalBinary()
	assert.NoError(t, err)

	subscriber := &MockSubscriber{countData: countDataBytes, sessionData: sessionDataBytes}
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)

	var useBigtable bool
	{
		bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
		assert.NoError(t, err)

		bigtableEmulation := checkBigtableEmulation()

		useBigtable = bigtableEnabled && bigtableEmulation
	}

	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
	btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
	btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
	
	btMaxAgeDays := 1

	logger := log.NewNopLogger()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, mockRedises[0].db.Addr(), mockRedises[1].db.Addr(), mockRedises[2].db.Addr(), mockRedises[3].db.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 4, 0, logger, &metrics.EmptyPortalCruncherMetrics)
	err = portalCruncher.PingRedis()
	assert.NoError(t, err)

	err = portalCruncher.Start(ctx, 1, 1, 1)
	assert.EqualError(t, err, "context deadline exceeded")

	{
		assert.Len(t, mockRedises[0].db.Keys(), 2)

		topSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, nextSessionData.Meta.ID, sessionID)

		customerTopSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("sc-%016x-%d", nextSessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err = strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, nextSessionData.Meta.ID, sessionID)
	}

	{
		assert.Len(t, mockRedises[1].db.Keys(), 2)

		pointVal := mockRedises[1].db.HGet(fmt.Sprintf("n-%016x-%d", nextSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", nextSessionData.Meta.ID))
		assert.Equal(t, nextSessionData.Point.RedisString(), pointVal)

		fields, err := mockRedises[1].db.HKeys(fmt.Sprintf("c-%016x-%d", newCountData.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, fmt.Sprintf("%016x", newCountData.ServerID), fields[0])

		countVal := mockRedises[1].db.HGet(fmt.Sprintf("c-%016x-%d", newCountData.BuyerID, minutes), fmt.Sprintf("%016x", newCountData.ServerID))
		assert.Equal(t, fmt.Sprintf("%d", newCountData.NumSessions), countVal)
	}

	{
		assert.Len(t, mockRedises[2].db.Keys(), 1)

		metaVal, err := mockRedises[2].db.Get(fmt.Sprintf("sm-%016x", nextSessionData.Meta.ID))
		assert.NoError(t, err)

		assert.Equal(t, nextSessionData.Meta.RedisString(), metaVal)
	}

	{
		assert.Len(t, mockRedises[3].db.Keys(), 1)

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

	serverID := rand.Uint64()
	buyerID := rand.Uint64()
	oldCountData := getTestCountData(serverID, buyerID)

	nextSessionData := getTestSessionData(true, sessionID, userHash, true, false, flushTime)

	_, err = mockRedises[0].db.ZAdd(fmt.Sprintf("s-%d", minutes), nextSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", nextSessionData.Meta.ID))
	assert.NoError(t, err)

	_, err = mockRedises[0].db.ZAdd(fmt.Sprintf("sc-%016x-%d", nextSessionData.Meta.BuyerID, minutes), nextSessionData.Meta.DeltaRTT, fmt.Sprintf("%016x", nextSessionData.Meta.ID))
	assert.NoError(t, err)

	mockRedises[1].db.HSet(fmt.Sprintf("n-%016x-%d", nextSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", nextSessionData.Meta.ID))

	mockRedises[1].db.HSet(fmt.Sprintf("c-%016x-%d", oldCountData.BuyerID, minutes), fmt.Sprintf("%016x", oldCountData.ServerID))

	err = mockRedises[2].db.Set(fmt.Sprintf("sm-%016x", nextSessionData.Meta.ID), nextSessionData.Meta.RedisString())
	assert.NoError(t, err)

	_, err = mockRedises[3].db.RPush(fmt.Sprintf("ss-%016x", nextSessionData.Meta.ID), nextSessionData.Slice.RedisString())
	assert.NoError(t, err)

	newCountData := getTestCountData(serverID, buyerID)
	countDataBytes, err := newCountData.MarshalBinary()
	assert.NoError(t, err)

	directSessionData := getTestSessionData(true, sessionID, userHash, false, true, flushTime)
	sessionDataBytes, err := directSessionData.MarshalBinary()
	assert.NoError(t, err)

	subscriber := &MockSubscriber{countData: countDataBytes, sessionData: sessionDataBytes}
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)

	var useBigtable bool
	{
		bigtableEnabled, err := strconv.ParseBool(os.Getenv("ENABLE_BIGTABLE"))
		assert.NoError(t, err)

		bigtableEmulation := checkBigtableEmulation()

		useBigtable = bigtableEnabled && bigtableEmulation
	}

	gcpProjectID := os.Getenv("GOOGLE_PROJECT_ID")
	btInstanceID:= os.Getenv("GOOGLE_BIGTABLE_INSTANCE_ID")
	btTableName := os.Getenv("GOOGLE_BIGTABLE_TABLE_NAME")
	btCfName := os.Getenv("GOOGLE_BIGTABLE_CF_NAME")
	
	btMaxAgeDays := 1

	logger := log.NewNopLogger()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, mockRedises[0].db.Addr(), mockRedises[1].db.Addr(), mockRedises[2].db.Addr(), mockRedises[3].db.Addr(), useBigtable, gcpProjectID, btInstanceID, btTableName, btCfName, btMaxAgeDays, 4, 0, logger, &metrics.EmptyPortalCruncherMetrics)
	err = portalCruncher.PingRedis()
	assert.NoError(t, err)

	err = portalCruncher.Start(ctx, 1, 1, 1)
	assert.EqualError(t, err, "context deadline exceeded")

	{
		assert.Len(t, mockRedises[0].db.Keys(), 2)

		topSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("s-%d", minutes))
		assert.NoError(t, err)
		assert.Len(t, topSessionIDs, 1)

		sessionID, err := strconv.ParseUint(topSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, directSessionData.Meta.ID, sessionID)

		customerTopSessionIDs, err := mockRedises[0].db.ZMembers(fmt.Sprintf("sc-%016x-%d", directSessionData.Meta.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, customerTopSessionIDs, 1)

		sessionID, err = strconv.ParseUint(customerTopSessionIDs[0], 16, 64)
		assert.NoError(t, err)

		assert.Equal(t, directSessionData.Meta.ID, sessionID)
	}

	{
		assert.Len(t, mockRedises[1].db.Keys(), 2)

		pointVal := mockRedises[1].db.HGet(fmt.Sprintf("d-%016x-%d", directSessionData.Meta.BuyerID, minutes), fmt.Sprintf("%016x", directSessionData.Meta.ID))
		assert.Equal(t, directSessionData.Point.RedisString(), pointVal)

		fields, err := mockRedises[1].db.HKeys(fmt.Sprintf("c-%016x-%d", newCountData.BuyerID, minutes))
		assert.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, fmt.Sprintf("%016x", newCountData.ServerID), fields[0])

		countVal := mockRedises[1].db.HGet(fmt.Sprintf("c-%016x-%d", newCountData.BuyerID, minutes), fmt.Sprintf("%016x", newCountData.ServerID))
		assert.Equal(t, fmt.Sprintf("%d", newCountData.NumSessions), countVal)
	}

	{
		assert.Len(t, mockRedises[2].db.Keys(), 1)

		metaVal, err := mockRedises[2].db.Get(fmt.Sprintf("sm-%016x", directSessionData.Meta.ID))
		assert.NoError(t, err)

		assert.Equal(t, directSessionData.Meta.RedisString(), metaVal)
	}

	{
		assert.Len(t, mockRedises[3].db.Keys(), 1)

		sliceVals, err := mockRedises[3].db.List(fmt.Sprintf("ss-%016x", directSessionData.Meta.ID))
		assert.NoError(t, err)
		assert.Len(t, sliceVals, 2)

		nextSliceVal := sliceVals[0]
		directSliceVal := sliceVals[1]

		assert.Equal(t, nextSessionData.Slice.RedisString(), nextSliceVal)
		assert.Equal(t, directSessionData.Slice.RedisString(), directSliceVal)
	}
}
