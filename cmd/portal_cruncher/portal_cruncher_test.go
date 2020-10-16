package main

import (
	"context"
	"testing"
	"time"

	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
	"github.com/stretchr/testify/assert"
)

func RunTestPubSub(t *testing.T, portalPublisher pubsub.Publisher, portalSubscriber pubsub.Subscriber, topic pubsub.Topic, message []byte, totalMessages int, messageChan chan []byte, metrics *metrics.PortalCruncherMetrics) error {
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

		err = RunTestPubSub(t, portalPublisher, portalSubscriber, pubsub.TopicPortalCruncherSessionData, sessionData, messageChanSize+1, messageChan, metrics)
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

		err = RunTestPubSub(t, portalPublisher, portalSubscriber, pubsub.TopicPortalCruncherSessionData, sessionData, messageChanSize, messageChan, metrics)
		assert.NoError(t, err)

		assert.Equal(t, float64(messageChanSize), metrics.ReceivedMessageCount.Value())

		assert.Len(t, messageChan, messageChanSize)
		for i := 0; i < messageChanSize; i++ {
			actualSessionData := <-messageChan
			assert.Equal(t, sessionData, actualSessionData)
		}
	})
}

func TestRedisConnectionFailure(t *testing.T) {

}

func TestPortalDataUnmarshalFailure(t *testing.T) {

}

func TestNoEarlyFlush(t *testing.T) {

}

func TestRedisPingFailure(t *testing.T) {

}

func TestBuyerNotFoundDuringLargeCustomerCacheInsertion(t *testing.T) {

}

func TestBuyerNotFoundDuringLargeCustomerCacheFilteringTopSessions(t *testing.T) {

}

func TestBuyerNotFoundDuringLargeCustomerCacheFilteringRedis(t *testing.T) {

}

func TestDirectSession(t *testing.T) {

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
