package portalcruncher_test

// todo: this needs to become a functional test

/*
import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/alicebob/miniredis/v2/server"

	"github.com/networknext/backend/modules/crypto"
	ghostarmy "github.com/networknext/backend/modules/ghost_army"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/pubsub"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	portalcruncher "github.com/networknext/backend/modules/portal_cruncher"
)

func getTestCountData(serverID uint64, buyerID uint64) transport.SessionCountData {
	return transport.SessionCountData{
		Version:     transport.SessionCountDataVersion,
		ServerID:    serverID,
		BuyerID:     buyerID,
		NumSessions: rand.Uint32(),
	}
}

func getTestSessionData(largeCustomer bool, sessionID uint64, userHash uint64, buyerID uint64, onNetworkNext bool, everOnNetworkNext bool, timestamp time.Time) transport.SessionPortalData {
	relayID1 := crypto.HashID("127.0.0.1:10000")
	relayID2 := crypto.HashID("127.0.0.1:10001")
	relayID3 := crypto.HashID("127.0.0.1:10002")
	relayID4 := crypto.HashID("127.0.0.1:10003")

	return transport.SessionPortalData{
		Version: transport.SessionPortalDataVersion,
		Meta: transport.SessionMeta{
			Version:         transport.SessionMetaVersion,
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
					Version: transport.RelayHopVersion,
					ID:      relayID1,
					Name:    "local.test_relay.0",
				},
				{
					Version: transport.RelayHopVersion,
					ID:      relayID2,
					Name:    "local.test_relay.1",
				},
			},
			SDK:        "4.0.0",
			Connection: 3,
			NearbyRelays: []transport.NearRelayPortalData{
				{
					Version: transport.NearRelayPortalDataVersion,
					ID:      relayID3,
					Name:    "local.test_relay.2",
				},
				{
					Version: transport.NearRelayPortalDataVersion,
					ID:      relayID4,
					Name:    "local.test_relay.3",
				},
			},
			Platform: 1,
			BuyerID:  buyerID,
		},
		Point: transport.SessionMapPoint{
			Version:   transport.SessionMapPointVersion,
			Latitude:  45,
			Longitude: 90,
			SessionID: sessionID,
		},
		Slice: transport.SessionSlice{
			Version:   transport.SessionSliceVersion,
			Timestamp: timestamp.Truncate(time.Second),
			Next: routing.Stats{
				RTT:        50,
				Jitter:     0.1,
				PacketLoss: 0,
			},
			Direct: routing.Stats{
				RTT:        20,
				Jitter:     0.2,
				PacketLoss: 0.1,
			},
			Predicted: routing.Stats{
				RTT: 55,
			},
			ClientToServerStats: routing.Stats{
				Jitter:     0.1,
				PacketLoss: 0.01,
			},
			ServerToClientStats: routing.Stats{
				Jitter:     0.1,
				PacketLoss: 0.01,
			},
			RouteDiversity: uint32(31),
			Envelope: routing.Envelope{
				Up:   100,
				Down: 150,
			},
			OnNetworkNext:     onNetworkNext,
			IsMultiPath:       false,
			IsTryBeforeYouBuy: false,
		},
		LargeCustomer: largeCustomer,
		EverOnNext:    everOnNetworkNext,
	}
}

type BadMockSubscriber struct{}

func (mock *BadMockSubscriber) Subscribe(topic pubsub.Topic) error {
	return nil
}

func (mock *BadMockSubscriber) Unsubscribe(topic pubsub.Topic) error {
	return nil
}

func (mock *BadMockSubscriber) ReceiveMessage() <-chan pubsub.MessageInfo {
	resultChan := make(chan pubsub.MessageInfo)
	resultFunc := func(topic pubsub.Topic, message []byte, err error) {
		resultChan <- pubsub.MessageInfo{
			Topic:   topic,
			Message: message,
			Err:     err,
		}
	}

	go resultFunc(0, nil, errors.New("bad data"))
	return resultChan
}

type SimpleMockSubscriber struct {
	topic       pubsub.Topic
	countData   []byte
	sessionData []byte
}

func (mock *SimpleMockSubscriber) Subscribe(topic pubsub.Topic) error {
	mock.topic = topic
	return nil
}

func (mock *SimpleMockSubscriber) Unsubscribe(topic pubsub.Topic) error {
	mock.topic = 0
	return nil
}

func (mock *SimpleMockSubscriber) ReceiveMessage() <-chan pubsub.MessageInfo {
	resultChan := make(chan pubsub.MessageInfo)
	resultFunc := func(topic pubsub.Topic, message []byte, err error) {
		resultChan <- pubsub.MessageInfo{
			Topic:   topic,
			Message: message,
			Err:     err,
		}
	}

	switch mock.topic {
	case pubsub.TopicPortalCruncherSessionCounts:
		go resultFunc(mock.topic, mock.countData, nil)
		return resultChan

	case pubsub.TopicPortalCruncherSessionData:
		go resultFunc(mock.topic, mock.sessionData, nil)
		return resultChan

	default:
		go resultFunc(mock.topic, []byte("bad topic"), nil)
		return resultChan
	}
}

type MockSubscriber struct {
	topics []pubsub.Topic

	redises *MockRedis
	expire  bool

	countData   [][]byte
	sessionData [][]byte

	receiveCount int
	maxMessages  int
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

func (mock *MockSubscriber) ReceiveMessage() <-chan pubsub.MessageInfo {
	messageIndex := mock.receiveCount % 2

	resultChan := make(chan pubsub.MessageInfo)
	resultFunc := func(topic pubsub.Topic, message []byte, err error) {
		resultChan <- pubsub.MessageInfo{
			Topic:   topic,
			Message: message,
			Err:     err,
		}
	}

	if mock.receiveCount >= mock.maxMessages {
		return resultChan
	}

	if mock.receiveCount > 0 && messageIndex == 0 {
		time.Sleep(time.Millisecond * 10)
		mock.redises.db.FastForward(time.Minute * 2)
	}

	topic := mock.topics[messageIndex]
	defer func() { mock.receiveCount++ }()

	switch topic {
	case pubsub.TopicPortalCruncherSessionCounts:
		go resultFunc(topic, mock.countData[mock.receiveCount/2], nil)
		return resultChan

	case pubsub.TopicPortalCruncherSessionData:
		go resultFunc(topic, mock.sessionData[mock.receiveCount/2], nil)
		return resultChan

	default:
		go resultFunc(topic, []byte("bad data"), nil)
		return resultChan
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

func checkBigtableEmulation(t *testing.T) {
	bigtableEmulatorHost := os.Getenv("BIGTABLE_EMULATOR_HOST")
	if bigtableEmulatorHost == "" {
		t.Skip("Bigtable emulator not set up, skipping portal cruncher bigtable tests")
	}
}

func CreateBigtableTable(ctx context.Context, btAdmin *storage.BigTableAdmin, btTableName string, btCfNames []string, btMaxAgeDays int) error {
	// Create a table with the given name and column families
	if err := btAdmin.CreateTable(ctx, btTableName, btCfNames); err != nil {
		return err
	}

	// Set a garbage collection policy of maxAgeDays
	maxAge := time.Hour * time.Duration(24*btMaxAgeDays)
	if err := btAdmin.SetMaxAgePolicy(ctx, btTableName, btCfNames, maxAge); err != nil {
		return fmt.Errorf("SetupBigtable() Failed to set max age policy: %v", err)
	}

	return nil
}

func TestNewPortalCruncher_RedisPoolMissingHostname(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, "", "", 0, 0, false, "", "", "", "", false, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.Nil(t, portalCruncher)
	assert.Error(t, err)
}

func TestNewPortalCruncher_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)
}

func TestNewPortalCruncher_Bigtable_NoGoogleProjectID(t *testing.T) {
	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, true, "", "", "", "", false, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.Nil(t, portalCruncher)
	assert.EqualError(t, err, "SetupBigtable() No GCP Project ID found. Could not find $BIGTABLE_EMULATOR_HOST for local testing.")
}

func TestNewPortalCruncher_Bigtable_TableDoesNotExist(t *testing.T) {
	checkBigtableEmulation(t)

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, true, "local", "", "table-does-not-exist", "", true, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.Nil(t, portalCruncher)
	assert.EqualError(t, err, "SetupBigtable() Could not find table table-does-not-exist in bigtable instance. Create the table before starting the portal cruncher")
}

func TestNewPortalCruncher_Bigtable_Success(t *testing.T) {
	checkBigtableEmulation(t)

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	// Create the Bigtable table
	gcpProjectID := "local"
	btInstanceID := ""
	btTableName := "test-table"
	btCfName := "test-table-cf"
	btMaxAgeDays := 1

	btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID)
	assert.NoError(t, err)
	err = CreateBigtableTable(ctx, btAdmin, btTableName, []string{btCfName}, btMaxAgeDays)
	assert.NoError(t, err)
	tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	assert.NoError(t, err)
	assert.True(t, tableExists, true)

	defer func() {
		err := btAdmin.DeleteTable(ctx, btTableName)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	}()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, true, gcpProjectID, btInstanceID, btTableName, btCfName, true, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)
}

func TestReceiveMessage_ReceiveError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	subscriber := &BadMockSubscriber{}

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	err = <-portalCruncher.ReceiveMessage(ctx)
	assert.EqualError(t, err, "error receiving message: bad data")
}

func TestReceiveMessage_CountData_SerializeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	subscriber := &SimpleMockSubscriber{countData: []byte("bad data")}
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	err = <-portalCruncher.ReceiveMessage(ctx)
	assert.Contains(t, err.Error(), "could not unmarshal message: ")
}

func TestReceiveMessage_CountData_ChannelFull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	countData := getTestCountData(rand.Uint64(), rand.Uint64())
	countDataBytes, err := transport.WriteSessionCountData(&countData)
	assert.NoError(t, err)

	subscriber := &SimpleMockSubscriber{countData: countDataBytes}
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	err = <-portalCruncher.ReceiveMessage(ctx)
	assert.Equal(t, err, &portalcruncher.ErrChannelFull{})
}

func TestReceiveMessage_CountData_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	countData := getTestCountData(rand.Uint64(), rand.Uint64())
	countDataBytes, err := transport.WriteSessionCountData(&countData)
	assert.NoError(t, err)

	subscriber := &SimpleMockSubscriber{countData: countDataBytes}
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 1, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	err = <-portalCruncher.ReceiveMessage(ctx)
	assert.NoError(t, err)
}

func TestReceiveMessage_SessionData_SerializeFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	subscriber := &SimpleMockSubscriber{countData: []byte("bad data")}
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	err = <-portalCruncher.ReceiveMessage(ctx)
	assert.Contains(t, err.Error(), "could not unmarshal message: ")
}

func TestReceiveMessage_SessionData_ChannelFull(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), rand.Uint64(), true, true, time.Now())
	sessionDataBytes, err := transport.WriteSessionPortalData(&sessionData)
	assert.NoError(t, err)

	subscriber := &SimpleMockSubscriber{sessionData: sessionDataBytes}
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	err = <-portalCruncher.ReceiveMessage(ctx)
	assert.Equal(t, err, &portalcruncher.ErrChannelFull{})
}

func TestReceiveMessage_SessionData_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), rand.Uint64(), true, true, time.Now())
	sessionDataBytes, err := transport.WriteSessionPortalData(&sessionData)
	assert.NoError(t, err)

	subscriber := &SimpleMockSubscriber{sessionData: sessionDataBytes}
	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 1, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	err = <-portalCruncher.ReceiveMessage(ctx)
	assert.NoError(t, err)
}

func TestReceiveMessage_UnknownMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	subscriber := &SimpleMockSubscriber{}
	subscriber.Subscribe(0)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, subscriber, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 1, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	err = <-portalCruncher.ReceiveMessage(ctx)
	assert.Equal(t, err, &portalcruncher.ErrUnknownMessage{})
}

func TestPingRedis_PoolClosed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 200) // have to sleep here otherwise miniredis can deadlock from closing too quickly after starting
	redisPool.Close()

	err = portalCruncher.PingRedis()
	assert.Contains(t, err.Error(), "could not ping: ")
}

func TestPingRedis_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	err = portalCruncher.PingRedis()
	assert.NoError(t, err)
}

func TestCloseRedisPool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, false, "", "", "", "", false, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	time.Sleep(time.Millisecond * 200) // have to sleep here otherwise miniredis can deadlock from closing too quickly after starting
	portalCruncher.CloseRedisPool()

	err = portalCruncher.PingRedis()
	assert.Contains(t, err.Error(), "could not ping: ")
}

func TestInsertIntoBigtable_DoNotInsertGhostArmyInProd(t *testing.T) {
	checkBigtableEmulation(t)

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	// Create the Bigtable table
	gcpProjectID := "local"
	btInstanceID := ""
	btTableName := "test-table"
	btCfName := "test-table-cf"
	btMaxAgeDays := 1

	btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID)
	assert.NoError(t, err)
	err = CreateBigtableTable(ctx, btAdmin, btTableName, []string{btCfName}, btMaxAgeDays)
	assert.NoError(t, err)
	tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	assert.NoError(t, err)
	assert.True(t, tableExists, true)

	defer func() {
		err := btAdmin.DeleteTable(ctx, btTableName)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	}()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, true, gcpProjectID, btInstanceID, btTableName, btCfName, true, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), ghostarmy.GhostArmyBuyerID("prod"), true, true, time.Now())

	err = portalCruncher.InsertIntoBigtable(ctx, []*transport.SessionPortalData{&sessionData}, "prod", ghostarmy.GhostArmyBuyerID("prod"))
	assert.NoError(t, err)

	// Create temp bigtable client to verify row does not exist
	btClient, err := storage.NewBigTable(ctx, gcpProjectID, btInstanceID, btTableName)
	assert.NoError(t, err)
	defer func() {
		err := btClient.Close()
		assert.NoError(t, err)
	}()

	sessionRowKey := fmt.Sprintf("%016x", sessionData.Meta.ID)
	row, err := btClient.GetRowWithRowKey(ctx, sessionRowKey)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(row))
}

func TestInsertIntoBigtable_InsertGhostArmyInNonProdEnv_Success(t *testing.T) {
	checkBigtableEmulation(t)

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	// Create the Bigtable table
	gcpProjectID := "local"
	btInstanceID := ""
	btTableName := "test-table"
	btCfName := "test-table-cf"
	btMaxAgeDays := 1

	btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID)
	assert.NoError(t, err)
	err = CreateBigtableTable(ctx, btAdmin, btTableName, []string{btCfName}, btMaxAgeDays)
	assert.NoError(t, err)
	tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	assert.NoError(t, err)
	assert.True(t, tableExists, true)

	defer func() {
		err := btAdmin.DeleteTable(ctx, btTableName)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	}()

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, true, gcpProjectID, btInstanceID, btTableName, btCfName, true, "", 0, &metrics.EmptyPortalCruncherMetrics, &metrics.EmptyBigTableMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), ghostarmy.GhostArmyBuyerID("local"), true, true, time.Now())

	err = portalCruncher.InsertIntoBigtable(ctx, []*transport.SessionPortalData{&sessionData}, "local", ghostarmy.GhostArmyBuyerID("local"))
	assert.NoError(t, err)

	// Create temp bigtable client to verify row does not exist
	btClient, err := storage.NewBigTable(ctx, gcpProjectID, btInstanceID, btTableName)
	assert.NoError(t, err)
	defer func() {
		err := btClient.Close()
		assert.NoError(t, err)
	}()

	sessionRowKey := fmt.Sprintf("%016x", sessionData.Meta.ID)
	row, err := btClient.GetRowWithRowKey(ctx, sessionRowKey)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(row))
}

func TestInsertIntoBigtable_InsertSessionDataFailure(t *testing.T) {
	checkBigtableEmulation(t)

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	// Create the Bigtable table
	gcpProjectID := "local"
	btInstanceID := ""
	btTableName := "test-table"
	btCfName := "test-table-cf"
	btMaxAgeDays := 1

	btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID)
	assert.NoError(t, err)
	err = CreateBigtableTable(ctx, btAdmin, btTableName, []string{btCfName}, btMaxAgeDays)
	assert.NoError(t, err)
	tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	assert.NoError(t, err)
	assert.True(t, tableExists, true)

	defer func() {
		err := btAdmin.DeleteTable(ctx, btTableName)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	}()

	btMetrics, err := metrics.NewBigTableMetrics(ctx, &metrics.LocalHandler{})
	assert.NoError(t, err)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, true, gcpProjectID, btInstanceID, btTableName, "cf-dne", true, "", 0, &metrics.EmptyPortalCruncherMetrics, btMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), rand.Uint64(), true, true, time.Now())

	// SessionMeta and SessionSlice are inserted the same way
	err = portalCruncher.InsertIntoBigtable(ctx, []*transport.SessionPortalData{&sessionData}, "local", ghostarmy.GhostArmyBuyerID("local"))
	assert.Error(t, err)
	assert.Equal(t, 1, int(btMetrics.WriteMetaFailureCount.Value()))
}

func TestInsertIntoBigtable_Success(t *testing.T) {
	checkBigtableEmulation(t)

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	// Create the Bigtable table
	gcpProjectID := "local"
	btInstanceID := ""
	btTableName := "test-table"
	btCfName := "test-table-cf"
	btMaxAgeDays := 1

	btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID)
	assert.NoError(t, err)
	err = CreateBigtableTable(ctx, btAdmin, btTableName, []string{btCfName}, btMaxAgeDays)
	assert.NoError(t, err)
	tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	assert.NoError(t, err)
	assert.True(t, tableExists, true)

	defer func() {
		err := btAdmin.DeleteTable(ctx, btTableName)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	}()

	btMetrics, err := metrics.NewBigTableMetrics(ctx, &metrics.LocalHandler{})
	assert.NoError(t, err)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, true, gcpProjectID, btInstanceID, btTableName, btCfName, true, "", 0, &metrics.EmptyPortalCruncherMetrics, btMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), rand.Uint64(), true, true, time.Now())

	err = portalCruncher.InsertIntoBigtable(ctx, []*transport.SessionPortalData{&sessionData}, "local", ghostarmy.GhostArmyBuyerID("local"))
	assert.NoError(t, err)
	assert.Equal(t, 1, int(btMetrics.WriteMetaSuccessCount.Value()))
	assert.Equal(t, 1, int(btMetrics.WriteSliceSuccessCount.Value()))
}

func TestCloseBigTable(t *testing.T) {
	checkBigtableEmulation(t)

	ctx := context.Background()

	redisPool, err := miniredis.Run()
	assert.NoError(t, err)
	defer redisPool.Close()

	// Create the Bigtable table
	gcpProjectID := "local"
	btInstanceID := ""
	btTableName := "test-table"
	btCfName := "test-table-cf"
	btMaxAgeDays := 1

	btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID)
	assert.NoError(t, err)
	err = CreateBigtableTable(ctx, btAdmin, btTableName, []string{btCfName}, btMaxAgeDays)
	assert.NoError(t, err)
	tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	assert.NoError(t, err)
	assert.True(t, tableExists, true)

	defer func() {
		err := btAdmin.DeleteTable(ctx, btTableName)
		assert.NoError(t, err)

		err = btAdmin.Close()
		assert.NoError(t, err)
	}()

	btMetrics, err := metrics.NewBigTableMetrics(ctx, &metrics.LocalHandler{})
	assert.NoError(t, err)

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx, nil, redisPool.Addr(), "", 1, 1, true, gcpProjectID, btInstanceID, btTableName, btCfName, true, "", 0, &metrics.EmptyPortalCruncherMetrics, btMetrics)
	assert.NotNil(t, portalCruncher)
	assert.NoError(t, err)

	portalCruncher.CloseBigTable()

	sessionData := getTestSessionData(false, rand.Uint64(), rand.Uint64(), rand.Uint64(), true, true, time.Now())

	err = portalCruncher.InsertIntoBigtable(ctx, []*transport.SessionPortalData{&sessionData}, "local", ghostarmy.GhostArmyBuyerID("local"))
	assert.Error(t, err)
}
*/
