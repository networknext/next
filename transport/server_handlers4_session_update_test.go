package transport_test

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

type badIPLocator struct{}

func (locator *badIPLocator) LocateIP(ip net.IP) (routing.Location, error) {
	return routing.LocationNullIsland, errors.New("bad location")
}

type goodIPLocator struct{}

func (locator *goodIPLocator) LocateIP(ip net.IP) (routing.Location, error) {
	return routing.LocationNullIsland, nil
}

type badRouteProvider struct{}

func (rp *badRouteProvider) GetDatacenterRelayIDs(datacenter routing.Datacenter) []uint64 {
	return nil
}
func (rp *badRouteProvider) GetAcceptableRoutes(nearIDs []routing.NearRelayData, destIDs []uint64, prevRouteHash uint64, rttEpsilon int32) ([]routing.Route, error) {
	return nil, errors.New("no acceptable routes")
}
func (rp *badRouteProvider) GetNearRelays(latitude float64, longitude float64, maxNearRelays int) ([]routing.NearRelayData, error) {
	return nil, errors.New("no near relays")
}

type goodRouteProvider struct{}

func (rp *goodRouteProvider) GetDatacenterRelayIDs(datacenter routing.Datacenter) []uint64 {
	return nil
}
func (rp *goodRouteProvider) GetAcceptableRoutes(nearIDs []routing.NearRelayData, destIDs []uint64, prevRouteHash uint64, rttEpsilon int32) ([]routing.Route, error) {
	return []routing.Route{}, nil
}
func (rp *goodRouteProvider) GetNearRelays(latitude float64, longitude float64, maxNearRelays int) ([]routing.NearRelayData, error) {
	return []routing.NearRelayData{}, nil
}

func TestSessionUpdateHandler4ReadPacketFailure(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	handler := transport.SessionUpdateHandlerFunc4(logger, nil, nil, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: nil,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.ErrorMetrics.ReadPacketFailure.Value(), 1.0)
}

func TestSessionUpdateHandler4ClientLocateFailure(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		SessionDataBytes:     1,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var badIPLocator badIPLocator
	ipLocatorFunc := func() routing.IPLocator {
		return &badIPLocator
	}

	var badRouteProvider badRouteProvider
	routeProviderFunc := func() transport.RouteProvider {
		return &badRouteProvider
	}

	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeProviderFunc, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, requestPacket.SessionID, responsePacket.SessionID)
	assert.Equal(t, int32(routing.RouteTypeDirect), responsePacket.RouteType)

	assert.Equal(t, int32(14), responsePacket.SessionDataBytes)
	assert.Equal(t, requestPacket.SessionID, sessionData.SessionID)
	assert.Equal(t, uint32(requestPacket.SliceNumber+1), sessionData.SliceNumber)

	assert.Equal(t, metrics.ErrorMetrics.ClientLocateFailure.Value(), 1.0)
}

func TestSessionUpdateHandler4NoNearRelays(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		SessionDataBytes:     1,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func() routing.IPLocator {
		return &goodIPLocator
	}

	var badRouteProvider badRouteProvider
	routeProviderFunc := func() transport.RouteProvider {
		return &badRouteProvider
	}

	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeProviderFunc, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, requestPacket.SessionID, responsePacket.SessionID)
	assert.Equal(t, int32(routing.RouteTypeDirect), responsePacket.RouteType)

	assert.Equal(t, int32(14), responsePacket.SessionDataBytes)
	assert.Equal(t, requestPacket.SessionID, sessionData.SessionID)
	assert.Equal(t, uint32(requestPacket.SliceNumber+1), sessionData.SliceNumber)

	assert.Equal(t, metrics.ErrorMetrics.NearRelaysLocateFailure.Value(), 1.0)
}

func TestSessionUpdateHandler4SessionDataBadSessionID(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	sessionDataMetrics, err := metrics.NewSessionDataMetrics(context.Background(), &metricsHandler)
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	metrics.SessionDataMetrics = *sessionDataMetrics
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		SliceNumber:          1,
		SessionDataBytes:     1,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func() routing.IPLocator {
		return &goodIPLocator
	}

	var goodRouteProvider goodRouteProvider
	routeProviderFunc := func() transport.RouteProvider {
		return &goodRouteProvider
	}

	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeProviderFunc, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.SessionDataMetrics.BadSessionID.Value(), 1.0)
}

func TestSessionUpdateHandler4SessionDataBadSequenceNumber(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	sessionDataMetrics, err := metrics.NewSessionDataMetrics(context.Background(), &metricsHandler)
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	metrics.SessionDataMetrics = *sessionDataMetrics
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	sessionDataStruct := transport.SessionData4{
		Version:     transport.SessionDataVersion4,
		SessionID:   1111,
		SliceNumber: 1,
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		SliceNumber:          1,
		SessionDataBytes:     8,
		SessionData:          sessionDataArray,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func() routing.IPLocator {
		return &goodIPLocator
	}

	var goodRouteProvider goodRouteProvider
	routeProviderFunc := func() transport.RouteProvider {
		return &goodRouteProvider
	}

	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeProviderFunc, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.SessionDataMetrics.BadSequenceNumber.Value(), 1.0)
}

func TestSessionUpdateHandler4InitialSlice(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func() routing.IPLocator {
		return &goodIPLocator
	}

	var goodRouteProvider goodRouteProvider
	routeProviderFunc := func() transport.RouteProvider {
		return &goodRouteProvider
	}

	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeProviderFunc, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, requestPacket.SessionID, responsePacket.SessionID)
	assert.Equal(t, int32(routing.RouteTypeDirect), responsePacket.RouteType)

	assert.Equal(t, int32(14), responsePacket.SessionDataBytes)
	assert.Equal(t, requestPacket.SessionID, sessionData.SessionID)
	assert.Equal(t, uint32(requestPacket.SliceNumber+1), sessionData.SliceNumber)
}

func TestSessionUpdateHandler4DirectRoute(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	sessionDataStruct := transport.SessionData4{
		Version:     transport.SessionDataVersion4,
		SessionID:   1111,
		SliceNumber: 1,
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		SliceNumber:          1,
		SessionDataBytes:     13,
		SessionData:          sessionDataArray,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func() routing.IPLocator {
		return &goodIPLocator
	}

	var goodRouteProvider goodRouteProvider
	routeProviderFunc := func() transport.RouteProvider {
		return &goodRouteProvider
	}

	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeProviderFunc, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, requestPacket.SessionID, responsePacket.SessionID)
	assert.Equal(t, int32(routing.RouteTypeDirect), responsePacket.RouteType)

	assert.Equal(t, int32(14), responsePacket.SessionDataBytes)
	assert.Equal(t, requestPacket.SessionID, sessionData.SessionID)
	assert.Equal(t, uint32(requestPacket.SliceNumber+1), sessionData.SliceNumber)
}
