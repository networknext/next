package vanity_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/alicebob/miniredis"
	"github.com/gomodule/redigo/redis"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/pubsub"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	vanity "github.com/networknext/backend/modules/vanity"
)

func getTestVanityData(buyerID uint64, userHash uint64, sessionID uint64, timestamp uint64) vanity.VanityMetrics {
	return vanity.VanityMetrics{
		BuyerID:                 buyerID,
		UserHash:                userHash,
		SessionID:               sessionID,
		Timestamp:               timestamp,
		SlicesAccelerated:       uint64(5),
		SlicesLatencyReduced:    uint64(5),
		SlicesPacketLossReduced: uint64(5),
		SlicesJitterReduced:     uint64(4),
		SessionsAccelerated:     uint64(1),
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
	topic      pubsub.Topic
	vanityData []byte
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
	case pubsub.TopicVanityMetricData:
		go resultFunc(mock.topic, mock.vanityData, nil)
		return resultChan
	default:
		go resultFunc(mock.topic, []byte("bad topic"), nil)
		return resultChan
	}
}

type MockSubscriber struct {
	topics []pubsub.Topic

	expire bool

	vanityData [][]byte

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

	topic := mock.topics[messageIndex]
	defer func() { mock.receiveCount++ }()

	switch topic {
	case pubsub.TopicVanityMetricData:
		go resultFunc(topic, mock.vanityData[mock.receiveCount/2], nil)
		return resultChan

	default:
		go resultFunc(topic, []byte("bad data"), nil)
		return resultChan
	}
}

func TestNewVanityMetrics(t *testing.T) {
	ctx := context.Background()
	redisServer, _ := miniredis.Run()

	// Get the time series metrics handler for vanity metrics (local since not actively writing)
	tsMetricsHandler := &metrics.LocalHandler{}

	// Get metrics for evaluating the performance of vanity metrics
	vanityServiceMetrics, err := metrics.NewVanityServiceMetrics(ctx, tsMetricsHandler)
	assert.NoError(t, err)

	logger := log.NewNopLogger()

	vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), "", 5, 5, time.Second*5, "testSet", "", logger)
	assert.NoError(t, err)
	assert.NotNil(t, vanityMetrics)
}

func TestReceiveMessage(t *testing.T) {
	ctx := context.Background()
	redisServer, _ := miniredis.Run()

	// Get the time series metrics handler for vanity metrics (local since not actively writing)
	tsMetricsHandler := &metrics.LocalHandler{}

	// Get metrics for evaluating the performance of vanity metrics
	vanityServiceMetrics, err := metrics.NewVanityServiceMetrics(ctx, tsMetricsHandler)
	assert.NoError(t, err)

	logger := log.NewNopLogger()

	t.Run("receive error", func(t *testing.T) {
		subscriber := &BadMockSubscriber{}

		vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, subscriber, redisServer.Addr(), "", 5, 5, time.Second*5, "testSet", "", logger)
		assert.NoError(t, err)

		err = vanityMetrics.ReceiveMessage(ctx)
		assert.EqualError(t, err, "error receiving message: bad data")
	})

	t.Run("vanity data unmarshal failure", func(t *testing.T) {
		subscriber := &SimpleMockSubscriber{vanityData: []byte("bad data")}
		subscriber.Subscribe(pubsub.TopicVanityMetricData)

		vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, subscriber, redisServer.Addr(), "", 5, 5, time.Second*5, "testSet", "", logger)
		assert.NoError(t, err)

		err = vanityMetrics.ReceiveMessage(ctx)
		assert.Contains(t, err.Error(), "could not unmarshal message: ")
	})

	t.Run("vanity data channel full", func(t *testing.T) {
		vanityData := getTestVanityData(rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64())
		vanityDataBytes, err := vanityData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &SimpleMockSubscriber{vanityData: vanityDataBytes}
		subscriber.Subscribe(pubsub.TopicVanityMetricData)

		vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 0, subscriber, redisServer.Addr(), "", 5, 5, time.Second*5, "testSet", "", logger)
		assert.NoError(t, err)

		err = vanityMetrics.ReceiveMessage(ctx)
		assert.Equal(t, err, &vanity.ErrChannelFull{})
	})

	t.Run("vanity data success", func(t *testing.T) {
		vanityData := getTestVanityData(rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64())
		vanityDataBytes, err := vanityData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &SimpleMockSubscriber{vanityData: vanityDataBytes}
		subscriber.Subscribe(pubsub.TopicVanityMetricData)

		vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, subscriber, redisServer.Addr(), "", 5, 5, time.Second*5, "testSet", "", logger)
		assert.NoError(t, err)

		err = vanityMetrics.ReceiveMessage(ctx)
		assert.NoError(t, err)
	})

	t.Run("unknown message", func(t *testing.T) {
		subscriber := &SimpleMockSubscriber{}
		subscriber.Subscribe(0)

		vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, subscriber, redisServer.Addr(), "", 5, 5, time.Second*5, "testSet", "", logger)
		assert.NoError(t, err)

		err = vanityMetrics.ReceiveMessage(ctx)
		assert.Equal(t, &vanity.ErrUnknownMessage{}, err)
	})
}

func TestUpdateMetrics(t *testing.T) {
	ctx := context.Background()

	// Get the time series metrics handler for vanity metrics (local since not actively writing)
	tsMetricsHandler := &metrics.LocalHandler{}

	// Get metrics for evaluating the performance of vanity metrics
	vanityServiceMetrics, err := metrics.NewVanityServiceMetrics(ctx, tsMetricsHandler)
	assert.NoError(t, err)

	logger := log.NewNopLogger()

	t.Run("add session id to redis", func(t *testing.T) {
		redisServer, _ := miniredis.Run()

		sessionID := rand.Uint64()
		sessionIDStr := fmt.Sprintf("%016x", sessionID)

		vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), "", 5, 5, time.Second*15, "testSet", "", logger)
		assert.NoError(t, err)

		err = vanityMetrics.AddSessionID(sessionIDStr)
		assert.NoError(t, err)
	})

	t.Run("check session id exists in redis", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		sessionID := rand.Uint64()
		sessionIDStr := fmt.Sprintf("%016x", sessionID)

		conn := storage.NewRedisPool(redisServer.Addr(), "", 5, 5).Get()

		member := fmt.Sprintf("sid-%s", sessionIDStr)
		_, err := conn.Do("ZADD", redis.Args{}.Add("testSet").Add(time.Now().UnixNano()).Add(member)...)
		assert.NoError(t, err)

		conn.Close()

		vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), "", 5, 5, time.Second*15, "testSet", "", logger)
		assert.NoError(t, err)

		exists, err := vanityMetrics.SessionIDExists(sessionIDStr)
		assert.NoError(t, err)
		assert.Equal(t, true, exists)
	})

	t.Run("check session ID expiration in redis", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		sessionID := rand.Uint64()
		sessionIDStr := fmt.Sprintf("%016x", sessionID)

		vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), "", 5, 5, time.Millisecond*100, "testSet", "", logger)
		assert.NoError(t, err)

		conn := storage.NewRedisPool(redisServer.Addr(), "", 5, 5).Get()
		defer conn.Close()

		err = vanityMetrics.AddSessionID(sessionIDStr)
		assert.NoError(t, err)

		// Ensure set has the sessionID
		members, err := conn.Do("ZCARD", redis.Args{}.Add("testSet")...)
		assert.NoError(t, err)
		assert.NotNil(t, members)
		assert.Equal(t, int64(1), members)

		// Sleep for 200 milliseconds to let the expiration time limit reach
		time.Sleep(time.Millisecond * 200)

		// Expire old sessions
		err = vanityMetrics.ExpireOldSessions(conn)
		assert.NoError(t, err)

		members, err = conn.Do("ZCARD", redis.Args{}.Add("testSet")...)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), members)
	})

	t.Run("vanity data update metric failure", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		buyerID := rand.Uint64()
		vanityData := getTestVanityData(buyerID, rand.Uint64(), rand.Uint64(), rand.Uint64())

		vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), "", 5, 5, time.Second*5, "testSet", "", logger)
		assert.NoError(t, err)

		err = vanityMetrics.UpdateMetrics(ctx, []*vanity.VanityMetrics{&vanityData})
		errStr := fmt.Sprintf("Could not find buyerID %016x in map", buyerID)
		assert.EqualError(t, err, errStr)
		assert.Equal(t, float64(0), vanityServiceMetrics.UpdateVanitySuccessCount.Value())
	})

	t.Run("vanity data update metric success", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		buyerID := rand.Uint64()
		vanityData := getTestVanityData(buyerID, rand.Uint64(), rand.Uint64(), rand.Uint64())

		vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), "", 5, 5, time.Second*5, "testSet", "", logger)
		assert.NoError(t, err)

		exists, err := vanityMetrics.AddNewBuyerID(ctx, fmt.Sprintf("%016x", buyerID))
		assert.NoError(t, err)
		assert.Equal(t, true, exists)

		err = vanityMetrics.UpdateMetrics(ctx, []*vanity.VanityMetrics{&vanityData})
		assert.NoError(t, err)
		assert.Equal(t, 1.0, vanityServiceMetrics.UpdateVanitySuccessCount.Value())
	})
}

// Requires environment to not be local and for StackDriver
func TestReadingMetrics(t *testing.T) {
	ctx, cancelWriteLoop := context.WithCancel(context.Background())

	gcpProjectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID")
	if !ok {
		t.Skip() // Skip the test if GCP project ID isn't set
	}

	redisServer, _ := miniredis.Run()

	var enableSDMetrics bool
	var err error
	enableSDMetricsString, ok := os.LookupEnv("ENABLE_STACKDRIVER_METRICS")
	if ok {
		enableSDMetrics, err = strconv.ParseBool(enableSDMetricsString)
		assert.NoError(t, err)
	}

	if !enableSDMetrics {
		t.Skip()
	}

	var sd *metrics.StackDriverHandler
	if enableSDMetrics {
		// Set up StackDriver metrics
		sdHandler := metrics.StackDriverHandler{
			ProjectID:          gcpProjectID,
			OverwriteFrequency: time.Second,
			OverwriteTimeout:   10 * time.Second,
		}

		err := sdHandler.Open(ctx)
		assert.NoError(t, err)

		sd = &sdHandler
	}

	logger := log.NewNopLogger()

	err = os.Setenv("FEATURE_VANITY_METRIC_WRITE_INTERVAL", "1s")
	assert.NoError(t, err)
	defer os.Unsetenv("FEATURE_VANITY_METRIC_WRITE_INTERVAL")

	// Get the time series metrics handler for vanity metrics
	tsMetricsHandler, err := backend.GetTSMetricsHandler(ctx, logger, gcpProjectID)
	assert.NoError(t, err)

	// Get metrics for evaluating the performance of vanity metrics
	vanityServiceMetrics, err := metrics.NewVanityServiceMetrics(ctx, tsMetricsHandler)
	assert.NoError(t, err)

	vanityMetrics, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), "", 5, 5, time.Second*5, "testSet", "", logger)
	assert.NoError(t, err)

	startTime, err := time.Parse(time.RFC3339, "2020-12-01T15:04:05+07:00")
	assert.NoError(t, err)

	endTime := time.Now()

	t.Run("get custom metric types success", func(t *testing.T) {
		customMetrics, err := vanityMetrics.GetCustomMetricTypes(ctx, sd, gcpProjectID, "server_backend")
		assert.NoError(t, err)
		assert.NotNil(t, customMetrics)
	})

	t.Run("get point details success", func(t *testing.T) {
		customMetrics, err := vanityMetrics.GetCustomMetricTypes(ctx, sd, gcpProjectID, "server_backend")
		assert.NoError(t, err)
		assert.NotNil(t, customMetrics)

		metricType, ok := customMetrics["Server Backend Billing Entries Queued"]
		assert.Equal(t, true, ok)

		tsFilter := vanityMetrics.GetTimeSeriesFilter(metricType)
		tsName := vanityMetrics.GetTimeSeriesName(gcpProjectID, metricType)

		tsInterval := &monitoringpb.TimeInterval{EndTime: timestamppb.New(endTime), StartTime: timestamppb.New(startTime)}
		duration := endTime.Sub(startTime)

		sumAgg := &monitoringpb.Aggregation{
			AlignmentPeriod:    durationpb.New(duration),
			PerSeriesAligner:   monitoringpb.Aggregation_Aligner(14), // Get summed values per alignment period
			CrossSeriesReducer: monitoringpb.Aggregation_Reducer(4),  // Sum across each alignment period
		}

		pointsList, err := vanityMetrics.GetPointDetails(ctx, sd, tsName, tsFilter, tsInterval, sumAgg)
		assert.NoError(t, err)
		assert.NotNil(t, pointsList)
	})

	t.Run("get vanity metrics json success", func(t *testing.T) {
		buyerID := uint64(0)
		userHash := rand.Uint64()
		sessionID := rand.Uint64()
		timestamp := rand.Uint64()

		vanityData := getTestVanityData(buyerID, userHash, sessionID, timestamp)

		err := vanityMetrics.UpdateMetrics(ctx, []*vanity.VanityMetrics{&vanityData})
		assert.NoError(t, err)

		// Wait for StackDriver to write the result
		time.Sleep(time.Second * 2)

		jsonMarshal, err := vanityMetrics.GetVanityMetricJSON(ctx, sd, gcpProjectID, fmt.Sprintf("%016x", buyerID), startTime, endTime)
		assert.NoError(t, err)
		assert.NotNil(t, jsonMarshal)

		emptyMapJSON, err := json.Marshal(make(map[string]float64))
		assert.NoError(t, err)
		assert.NotEqual(t, emptyMapJSON, jsonMarshal)
	})

	// Stop the submit routine
	cancelWriteLoop()

	// Close the metric client
	err = sd.Close()
	assert.NoError(t, err)
}
