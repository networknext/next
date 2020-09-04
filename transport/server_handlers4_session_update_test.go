package transport_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestSessionUpdateHandler4ReadPacketFailure(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	handler := transport.SessionUpdateHandlerFunc4(logger, storer, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: nil,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.ErrorMetrics.ReadPacketFailure.Value(), 1.0)
}

func TestSessionUpdateHandler4SessionDataBadSessionID(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	metricsHandler := metrics.LocalHandler{}
	sessionDataMetrics, err := metrics.NewSessionDataMetrics(context.Background(), &metricsHandler)
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	metrics.SessionDataMetrics = *sessionDataMetrics
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.SessionUpdatePacket4{
		Sequence:             1,
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
		SessionDataBytes:     1,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.SessionUpdateHandlerFunc4(logger, storer, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.SessionDataMetrics.BadSessionID.Value(), 1.0)
}

func TestSessionUpdateHandler4SessionDataBadSequenceNumber(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	metricsHandler := metrics.LocalHandler{}
	sessionDataMetrics, err := metrics.NewSessionDataMetrics(context.Background(), &metricsHandler)
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	metrics.SessionDataMetrics = *sessionDataMetrics
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	sessionDataStruct := transport.SessionData4{
		Version:   transport.SessionDataVersion4,
		SessionID: 1111,
		Sequence:  1,
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	requestPacket := transport.SessionUpdatePacket4{
		Sequence:             1,
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
		SessionDataBytes:     8,
		SessionData:          sessionDataArray,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.SessionUpdateHandlerFunc4(logger, storer, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.SessionDataMetrics.BadSequenceNumber.Value(), 1.0)
}

func TestSessionUpdateHandler4DirectRoute(t *testing.T) {
	logger := log.NewNopLogger()
	storer := &storage.InMemory{}
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	sessionDataStruct := transport.SessionData4{
		Version:   transport.SessionDataVersion4,
		SessionID: 1111,
		Sequence:  1,
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	requestPacket := transport.SessionUpdatePacket4{
		Sequence:             1,
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
		SessionDataBytes:     12,
		SessionData:          sessionDataArray,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.SessionUpdateHandlerFunc4(logger, storer, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.Sequence, responsePacket.Sequence)
	assert.Equal(t, requestPacket.SessionID, responsePacket.SessionID)
	assert.Equal(t, requestPacket.ServerRoutePublicKey, responsePacket.ServerRoutePublicKey)
	assert.Equal(t, int32(routing.RouteTypeDirect), responsePacket.RouteType)

	assert.Equal(t, int32(13), responsePacket.SessionDataBytes)
	assert.Equal(t, requestPacket.SessionID, sessionData.SessionID)
	assert.Equal(t, uint32(requestPacket.Sequence+1), sessionData.Sequence)
}
