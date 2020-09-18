package transport_test

import (
	"bytes"
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
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

func TestSessionUpdateHandler4ReadPacketFailure(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	handler := transport.SessionUpdateHandlerFunc4(logger, nil, nil, nil, [crypto.KeySize]byte{}, nil, metrics)
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
	storer := &storage.InMemory{}
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var badIPLocator badIPLocator
	ipLocatorFunc := func() routing.IPLocator {
		return &badIPLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       make([]uint64, 0),
		NearRelayAddresses: make([]net.UDPAddr, 0),
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.PostSessionHandler{}
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, storer, [crypto.KeySize]byte{}, &postSessionHandler, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.ErrorMetrics.ClientLocateFailure.Value(), 1.0)
}

func TestSessionUpdateHandler4NoNearRelays(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)

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

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       make([]uint64, 0),
		NearRelayAddresses: make([]net.UDPAddr, 0),
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.PostSessionHandler{}
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, storer, [crypto.KeySize]byte{}, &postSessionHandler, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.ErrorMetrics.NearRelaysLocateFailure.Value(), 1.0)
}

func TestSessionUpdateHandler4ReadSessionDataFailure(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	sessionDataMetrics, err := metrics.NewSessionDataMetrics(context.Background(), &metricsHandler)
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	metrics.SessionDataMetrics = *sessionDataMetrics
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		SliceNumber:          1,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func() routing.IPLocator {
		return &goodIPLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       []uint64{},
		NearRelayAddresses: []net.UDPAddr{},
	}

	expectedSessionData := transport.SessionData4{}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.PostSessionHandler{}
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, storer, [crypto.KeySize]byte{}, &postSessionHandler, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionDataMetrics.ReadSessionDataFailure.Value(), 1.0)
}

// todo: datacenter lookup failure test

func TestSessionUpdateHandler4SessionDataBadSessionID(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	sessionDataMetrics, err := metrics.NewSessionDataMetrics(context.Background(), &metricsHandler)
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	metrics.SessionDataMetrics = *sessionDataMetrics
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)

	sessionDataStruct := transport.SessionData4{
		Version:     transport.SessionDataVersion4,
		SessionID:   1,
		SliceNumber: 1,
		Location:    routing.LocationNullIsland,
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		SliceNumber:          1,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func() routing.IPLocator {
		return &goodIPLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       []uint64{},
		NearRelayAddresses: []net.UDPAddr{},
	}

	expectedSessionData := transport.SessionData4{
		Version:     transport.SessionDataVersion4,
		SessionID:   1,
		SliceNumber: 1,
		Location:    routing.LocationNullIsland,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.PostSessionHandler{}
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, storer, [crypto.KeySize]byte{}, &postSessionHandler, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionDataMetrics.BadSessionID.Value(), 1.0)
}

func TestSessionUpdateHandler4SessionDataBadSliceNumber(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	sessionDataMetrics, err := metrics.NewSessionDataMetrics(context.Background(), &metricsHandler)
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	metrics.SessionDataMetrics = *sessionDataMetrics
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)

	sessionDataStruct := transport.SessionData4{
		Version:     transport.SessionDataVersion4,
		SessionID:   1111,
		SliceNumber: 1,
		Location:    routing.LocationNullIsland,
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		SliceNumber:          2,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func() routing.IPLocator {
		return &goodIPLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       []uint64{},
		NearRelayAddresses: []net.UDPAddr{},
	}

	expectedSessionData := transport.SessionData4{
		Version:     transport.SessionDataVersion4,
		SessionID:   1111,
		SliceNumber: 1,
		Location:    routing.LocationNullIsland,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.PostSessionHandler{}
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, storer, [crypto.KeySize]byte{}, &postSessionHandler, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionDataMetrics.BadSliceNumber.Value(), 1.0)
}

func TestSessionUpdateHandler4FirstSlice(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)

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

	relayAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	routeMatrix := routing.RouteMatrix4{
		RelayIDs:           []uint64{1},
		RelayAddresses:     []net.UDPAddr{*relayAddr},
		RelayNames:         []string{"test.relay"},
		RelayLatitudes:     []float32{90},
		RelayLongitudes:    []float32{180},
		RelayDatacenterIDs: []uint64{10},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NumNearRelays:      1,
		NearRelayIDs:       []uint64{1},
		NearRelayAddresses: []net.UDPAddr{*relayAddr},
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.PostSessionHandler{}
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, storer, [crypto.KeySize]byte{}, &postSessionHandler, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)
}

func TestSessionUpdateHandler4DirectRoute(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewSessionMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)

	sessionDataStruct := transport.SessionData4{
		Version:         transport.SessionDataVersion4,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
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

	relayAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	routeMatrix := routing.RouteMatrix4{
		RelayIDs:           []uint64{1},
		RelayAddresses:     []net.UDPAddr{*relayAddr},
		RelayNames:         []string{"test.relay"},
		RelayLatitudes:     []float32{90},
		RelayLongitudes:    []float32{180},
		RelayDatacenterIDs: []uint64{10},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NumNearRelays:      1,
		NearRelayIDs:       []uint64{1},
		NearRelayAddresses: []net.UDPAddr{*relayAddr},
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.PostSessionHandler{}
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, storer, [crypto.KeySize]byte{}, &postSessionHandler, metrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes())
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)
}
