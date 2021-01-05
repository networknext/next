package vanity_test

import (
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

	vanityMetrics := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), 5, 5, 5)
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

	t.Run("receive error", func(t *testing.T) {
		subscriber := &BadMockSubscriber{}

		vanityMetrics := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, subscriber, redisServer.Addr(), 5, 5, 5)

		err = vanityMetrics.ReceiveMessage(ctx)
		assert.EqualError(t, err, "error receiving message: bad data")
	})

	t.Run("vanity data unmarshal failure", func(t *testing.T) {
		subscriber := &SimpleMockSubscriber{vanityData: []byte("bad data")}
		subscriber.Subscribe(pubsub.TopicVanityMetricData)

		vanityMetrics := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, subscriber, redisServer.Addr(), 5, 5, 5)

		err = vanityMetrics.ReceiveMessage(ctx)
		assert.Contains(t, err.Error(), "could not unmarshal message: ")
	})

	t.Run("vanity data channel full", func(t *testing.T) {
		vanityData := getTestVanityData(rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64())
		vanityDataBytes, err := vanityData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &SimpleMockSubscriber{vanityData: vanityDataBytes}
		subscriber.Subscribe(pubsub.TopicVanityMetricData)

		vanityMetrics := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 0, subscriber, redisServer.Addr(), 5, 5, 5)

		err = vanityMetrics.ReceiveMessage(ctx)
		assert.Equal(t, err, &vanity.ErrChannelFull{})
	})

	t.Run("vanity data success", func(t *testing.T) {
		vanityData := getTestVanityData(rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64())
		vanityDataBytes, err := vanityData.MarshalBinary()
		assert.NoError(t, err)

		subscriber := &SimpleMockSubscriber{vanityData: vanityDataBytes}
		subscriber.Subscribe(pubsub.TopicVanityMetricData)

		vanityMetrics := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, subscriber, redisServer.Addr(), 5, 5, 5)

		err = vanityMetrics.ReceiveMessage(ctx)
		assert.NoError(t, err)
	})

	t.Run("unknown message", func(t *testing.T) {
		subscriber := &SimpleMockSubscriber{}
		subscriber.Subscribe(0)

		vanityMetrics := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, subscriber, redisServer.Addr(), 5, 5, 5)

		err = vanityMetrics.ReceiveMessage(ctx)
		assert.Equal(t, &vanity.ErrUnknownMessage{}, err)
	})
}

func TestUpdateMetrics(t *testing.T) {
	ctx := context.Background()
	redisServer, _ := miniredis.Run()

	// Get the time series metrics handler for vanity metrics (local since not actively writing)
	tsMetricsHandler := &metrics.LocalHandler{}

	// Get metrics for evaluating the performance of vanity metrics
	vanityServiceMetrics, err := metrics.NewVanityServiceMetrics(ctx, tsMetricsHandler)
	assert.NoError(t, err)

	t.Run("put user session in redis", func(t *testing.T) {
		userHash := rand.Uint64()
		sessionID := rand.Uint64()
		timestamp := rand.Uint64()

		userSession := vanity.UserSession{SessionID: fmt.Sprintf("%016x", sessionID), Timestamp: fmt.Sprintf("%016x", timestamp)}

		vanityMetrics := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), 5, 5, 15)
		err := vanityMetrics.PutUserSession(fmt.Sprintf("%016x", userHash), &userSession)
		assert.NoError(t, err)
	})

	t.Run("get user session from redis", func(t *testing.T) {
		userHash := rand.Uint64()
		sessionID := rand.Uint64()
		timestamp := rand.Uint64()

		conn := storage.NewRedisPool(redisServer.Addr(), 5, 5).Get()

		userSession := vanity.UserSession{SessionID: fmt.Sprintf("%016x", sessionID), Timestamp: fmt.Sprintf("%016x", timestamp)}
		key := fmt.Sprintf("uh-%s", fmt.Sprintf("%016x", userHash))
		_, err := conn.Do("HMSET", redis.Args{}.Add(key).AddFlat(&userSession)...)
		assert.NoError(t, err)

		conn.Close()

		vanityMetrics := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), 5, 5, 15)
		recvSession, exists, err := vanityMetrics.GetUserSession(fmt.Sprintf("%016x", userHash))
		assert.NoError(t, err)
		assert.Equal(t, true, exists)
		assert.Equal(t, &userSession, recvSession)
	})

	t.Run("check user hash expiration in redis", func(t *testing.T) {
		userHash := rand.Uint64()
		sessionID := rand.Uint64()
		timestamp := rand.Uint64()

		userSession := vanity.UserSession{SessionID: fmt.Sprintf("%016x", sessionID), Timestamp: fmt.Sprintf("%016x", timestamp)}

		vanityMetrics := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), 5, 5, 1)
		key := fmt.Sprintf("uh-%s", fmt.Sprintf("%016x", userHash))
		err := vanityMetrics.PutUserSession(key, &userSession)
		assert.NoError(t, err)

		// Sleep for 3 seconds to let the userHash expire
		time.Sleep(3)

		conn := storage.NewRedisPool(redisServer.Addr(), 5, 5).Get()
		defer conn.Close()

		ttl, err := conn.Do("TTL", redis.Args{}.Add(key))
		assert.NoError(t, err)
		assert.Equal(t, int64(-2), ttl) // -2 indicates key does not exist
	})

	t.Run("vanity data update metric success", func(t *testing.T) {
		vanityData := getTestVanityData(rand.Uint64(), rand.Uint64(), rand.Uint64(), rand.Uint64())

		vanityMetrics := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), 5, 5, 5)

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

	err = os.Setenv("FEATURE_VANITY_METRIC_WRITE_INTERVAL", "3s")
	assert.NoError(t, err)
	defer os.Unsetenv("FEATURE_VANITY_METRIC_WRITE_INTERVAL")

	// Get the time series metrics handler for vanity metrics
	tsMetricsHandler, err := backend.GetTSMetricsHandler(ctx, logger, gcpProjectID)
	assert.NoError(t, err)

	// Get metrics for evaluating the performance of vanity metrics
	vanityServiceMetrics, err := metrics.NewVanityServiceMetrics(ctx, tsMetricsHandler)
	assert.NoError(t, err)

	vanityMetrics := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, 1, nil, redisServer.Addr(), 5, 5, 5)

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

		maxAgg := &monitoringpb.Aggregation{
			AlignmentPeriod:    durationpb.New(duration),
			PerSeriesAligner:   monitoringpb.Aggregation_Aligner(11), // Get max values per alignment period
			CrossSeriesReducer: monitoringpb.Aggregation_Reducer(3),  // Get single max value across alignment periods
		}

		pointsList, err := vanityMetrics.GetPointDetails(ctx, sd, tsName, tsFilter, tsInterval, maxAgg)
		assert.NoError(t, err)
		assert.NotNil(t, pointsList)
	})

	t.Run("get vanity metrics json success", func(t *testing.T) {
		buyerID := rand.Uint64()
		userHash := rand.Uint64()
		sessionID := rand.Uint64()
		timestamp := rand.Uint64()

		vanityData := getTestVanityData(buyerID, userHash, sessionID, timestamp)

		err := vanityMetrics.UpdateMetrics(ctx, []*vanity.VanityMetrics{&vanityData})
		assert.NoError(t, err)

		// Wait for StackDriver to write the result
		time.Sleep(5)

		jsonMarshal, err := vanityMetrics.GetVanityMetricJSON(ctx, sd, gcpProjectID, fmt.Sprintf("%016x", buyerID), startTime, endTime)
		assert.NoError(t, err)
		assert.NotNil(t, jsonMarshal)
	})

	// Stop the submit routine
	cancelWriteLoop()

	// Close the metric client
	err = sd.Close()
	assert.NoError(t, err)
}
