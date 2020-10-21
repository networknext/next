package transport_test

import (
	"bytes"
	"context"
	"errors"
	"math/rand"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/modules/core"
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

func assertAllMetricsEqual(t *testing.T, expectedMetrics metrics.SessionUpdate4Metrics, actualMetrics metrics.SessionUpdate4Metrics) {
	expectedMetricsValue := reflect.ValueOf(expectedMetrics)
	actualMetricsValue := reflect.ValueOf(actualMetrics)
	for i := 0; i < actualMetricsValue.NumField(); i++ {
		if expectedMetricsValue.Field(i).CanInterface() && actualMetricsValue.Field(i).CanInterface() {
			expectedField := expectedMetricsValue.Field(i).Interface()
			actualField := actualMetricsValue.Field(i).Interface()

			expectedValuer, ok := expectedField.(metrics.Valuer)
			if !ok {
				continue
			}

			actualValuer, ok := actualField.(metrics.Valuer)
			if !ok {
				continue
			}

			assert.Equal(t, expectedValuer.Value(), actualValuer.Value(), expectedMetricsValue.Type().Field(i).Name)
		}
	}
}

// todo: these should be their own type/file and not tested alongside the session update handler
func TestGetRouteAddressesAndPublicKeysFailure(t *testing.T) {
	ctx := context.Background()

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:34567")
	assert.NoError(t, err)
	clientPublicKey := make([]byte, crypto.KeySize)
	core.RandomBytes(clientPublicKey)

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)
	serverPublicKey := make([]byte, crypto.KeySize)
	core.RandomBytes(serverPublicKey)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)
	relayAddr3, err := net.ResolveUDPAddr("udp", "127.0.0.1:10002")
	assert.NoError(t, err)

	relayPublicKey1 := make([]byte, crypto.KeySize)
	core.RandomBytes(relayPublicKey1)
	relayPublicKey2 := make([]byte, crypto.KeySize)
	core.RandomBytes(relayPublicKey2)
	relayPublicKey3 := make([]byte, crypto.KeySize)
	core.RandomBytes(relayPublicKey3)

	seller := routing.Seller{ID: "seller"}
	datacenter := routing.Datacenter{ID: crypto.HashID("local"), Name: "local"}

	storer := &storage.InMemory{}

	err = storer.AddSeller(ctx, seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(ctx, datacenter)
	assert.NoError(t, err)

	err = storer.AddRelay(ctx, routing.Relay{ID: crypto.HashID(relayAddr1.String()), Addr: *relayAddr1, PublicKey: relayPublicKey1, Seller: seller, Datacenter: datacenter})
	assert.NoError(t, err)
	err = storer.AddRelay(ctx, routing.Relay{ID: crypto.HashID(relayAddr2.String()), Addr: *relayAddr2, PublicKey: relayPublicKey2, Seller: seller, Datacenter: datacenter})
	assert.NoError(t, err)

	allRelayIDs := []uint64{crypto.HashID(relayAddr1.String()), crypto.HashID(relayAddr2.String()), crypto.HashID(relayAddr3.String())}
	routeRelays := []int32{1, 0, 2}

	routeAddresses, routePublicKeys := transport.GetRouteAddressesAndPublicKeys(clientAddr, clientPublicKey, serverAddr, serverPublicKey, 5, routeRelays, allRelayIDs, storer)
	assert.Nil(t, routeAddresses)
	assert.Nil(t, routePublicKeys)
}

func TestGetRouteAddressesAndPublicKeysSuccess(t *testing.T) {
	ctx := context.Background()

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:34567")
	assert.NoError(t, err)
	clientPublicKey := make([]byte, crypto.KeySize)
	core.RandomBytes(clientPublicKey)

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)
	serverPublicKey := make([]byte, crypto.KeySize)
	core.RandomBytes(serverPublicKey)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)
	relayAddr3, err := net.ResolveUDPAddr("udp", "127.0.0.1:10002")
	assert.NoError(t, err)

	relayPublicKey1 := make([]byte, crypto.KeySize)
	core.RandomBytes(relayPublicKey1)
	relayPublicKey2 := make([]byte, crypto.KeySize)
	core.RandomBytes(relayPublicKey2)
	relayPublicKey3 := make([]byte, crypto.KeySize)
	core.RandomBytes(relayPublicKey3)

	seller := routing.Seller{ID: "seller"}
	datacenter := routing.Datacenter{ID: crypto.HashID("local"), Name: "local"}

	storer := &storage.InMemory{}

	err = storer.AddSeller(ctx, seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(ctx, datacenter)
	assert.NoError(t, err)

	err = storer.AddRelay(ctx, routing.Relay{ID: crypto.HashID(relayAddr1.String()), Addr: *relayAddr1, PublicKey: relayPublicKey1, Seller: seller, Datacenter: datacenter})
	assert.NoError(t, err)
	err = storer.AddRelay(ctx, routing.Relay{ID: crypto.HashID(relayAddr2.String()), Addr: *relayAddr2, PublicKey: relayPublicKey2, Seller: seller, Datacenter: datacenter})
	assert.NoError(t, err)
	err = storer.AddRelay(ctx, routing.Relay{ID: crypto.HashID(relayAddr3.String()), Addr: *relayAddr3, PublicKey: relayPublicKey3, Seller: seller, Datacenter: datacenter})
	assert.NoError(t, err)

	expectedRouteAddresses := []*net.UDPAddr{clientAddr, relayAddr2, relayAddr1, relayAddr3, serverAddr}
	expectedRoutePublicKeys := [][]byte{clientPublicKey, relayPublicKey2, relayPublicKey1, relayPublicKey3, serverPublicKey}

	allRelayIDs := []uint64{crypto.HashID(relayAddr1.String()), crypto.HashID(relayAddr2.String()), crypto.HashID(relayAddr3.String())}
	routeRelays := []int32{1, 0, 2}

	routeAddresses, routePublicKeys := transport.GetRouteAddressesAndPublicKeys(clientAddr, clientPublicKey, serverAddr, serverPublicKey, 5, routeRelays, allRelayIDs, storer)
	assert.Equal(t, expectedRouteAddresses, routeAddresses)
	assert.Equal(t, expectedRoutePublicKeys, routePublicKeys)
}

func TestSessionUpdateHandler4ReadPacketFailure(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	handler := transport.SessionUpdateHandlerFunc4(logger, nil, nil, nil, nil, 32, [crypto.KeySize]byte{}, nil, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: nil,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.SessionUpdateMetrics.ReadPacketFailure.Value(), 1.0)
}

func TestSessionUpdateHandler4BuyerNotFound(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	handler := transport.SessionUpdateHandlerFunc4(logger, nil, nil, nil, storer, 32, [crypto.KeySize]byte{}, nil, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.SessionUpdateMetrics.BuyerNotFound.Value(), 1.0)
}

func TestSessionUpdateHandler4ClientLocateFailure(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{})
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var badIPLocator badIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &badIPLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

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
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.ClientLocateFailure.Value(), 1.0)
}

func TestSessionUpdateHandler4ReadSessionDataFailure(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{})
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
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

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

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.ReadSessionDataFailure.Value(), 1.0)
}

func TestSessionUpdateHandler4SessionDataBadSessionID(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{})
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
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

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

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.BadSessionID.Value(), 1.0)
}

func TestSessionUpdateHandler4SessionDataBadSliceNumber(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{})
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
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

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

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.BadSliceNumber.Value(), 1.0)
}

func TestSessionUpdateHandler4BuyerNotLive(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{})
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var ipLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &ipLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

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
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.BuyerNotLive.Value(), 1.0)
}

func TestSessionUpdateHandler4FallbackToDirect(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{Live: true})
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
		FallbackToDirect:     true,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var ipLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &ipLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       make([]uint64, 0),
		NearRelayAddresses: make([]net.UDPAddr, 0),
	}

	expectedSessionData := transport.SessionData4{
		SessionID:        requestPacket.SessionID,
		SliceNumber:      requestPacket.SliceNumber + 1,
		Location:         routing.LocationNullIsland,
		ExpireTimestamp:  uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		FellBackToDirect: true,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.FallbackToDirect.Value(), 1.0)
}

func TestSessionUpdateHandler4DatacenterNotFound(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{Live: true})

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       []uint64{},
		NearRelayAddresses: []net.UDPAddr{},
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

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.DatacenterNotFound.Value(), 1.0)
}

func TestSessionUpdateHandler4NoNearRelays(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{Live: true})
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	var routeMatrix routing.RouteMatrix4
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

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
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.NearRelaysLocateFailure.Value(), 1.0)
}

func TestSessionUpdateHandler4FirstSlice(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackend4Metrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdate4Metrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.DirectSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.DirectSlices.Add(1)

	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
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

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

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

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandler4NoDestRelays(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})

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
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
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
		RelayDatacenterIDs: []uint64{0},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

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

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.NoRelaysInDatacenter.Value(), 1.0)
}

func TestSessionUpdateHandler4DirectRoute(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackend4Metrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdate4Metrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.DirectSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.DirectSlices.Add(1)

	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})

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
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
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

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

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

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandler4NextRoute(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackend4Metrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdate4Metrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)

	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

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

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		DirectRTT:            60,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []float32{10, 15},
		NearRelayJitter:      []float32{0, 0},
		NearRelayPacketLoss:  []float32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix4{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						0, 1,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1)},
			},
		},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expireTimestamp := uint64(time.Now().Unix()) + billing.BillingSliceSeconds*2
	sessionVersion := sessionDataStruct.SessionVersion + 1

	tokenData := make([]byte, core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES*4)
	routeAddresses := make([]*net.UDPAddr, 0)
	routeAddresses = append(routeAddresses, clientAddr, relayAddr1, relayAddr2, serverAddr)
	routePublicKeys := make([][]byte, 0)
	routePublicKeys = append(routePublicKeys, publicKey, publicKey, publicKey, publicKey)
	core.WriteRouteTokens(tokenData, expireTimestamp, requestPacket.SessionID, uint8(sessionVersion), 1024, 1024, 4, routeAddresses, routePublicKeys, privateKey)
	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeNew,
		NumNearRelays:      2,
		NearRelayIDs:       []uint64{1, 2},
		NearRelayAddresses: []net.UDPAddr{*relayAddr1, *relayAddr2},
		NumTokens:          4,
		Tokens:             tokenData,
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         true,
		RouteNumRelays:  2,
		RouteCost:       45,
		RouteRelayIDs:   [routing.MaxRelays]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	// We can't check if the entire response is equal since the response's tokens will be different each time
	// since the encryption generates random bytes for the nonce
	assert.Equal(t, expectedResponse.SessionID, responsePacket.SessionID)
	assert.Equal(t, expectedResponse.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, expectedResponse.RouteType, responsePacket.RouteType)
	assert.Equal(t, expectedResponse.NumNearRelays, responsePacket.NumNearRelays)
	assert.Equal(t, expectedResponse.NearRelayIDs, responsePacket.NearRelayIDs)
	assert.Equal(t, expectedResponse.NearRelayAddresses, responsePacket.NearRelayAddresses)
	assert.Equal(t, expectedResponse.NumTokens, responsePacket.NumTokens)

	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandler4ContinueRoute(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackend4Metrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdate4Metrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)

	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData4{
		Version:         transport.SessionDataVersion4,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [routing.MaxRelays]uint64{2, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		DirectRTT:            60,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []float32{10, 15},
		NearRelayJitter:      []float32{0, 0},
		NearRelayPacketLoss:  []float32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix4{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						1, 0,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(1, 0)},
			},
		},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expireTimestamp := uint64(time.Now().Unix()) + billing.BillingSliceSeconds

	tokenData := make([]byte, core.NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES*4)
	routePublicKeys := make([][]byte, 0)
	routePublicKeys = append(routePublicKeys, publicKey, publicKey, publicKey, publicKey)
	core.WriteContinueTokens(tokenData, expireTimestamp, requestPacket.SessionID, uint8(sessionDataStruct.SessionVersion), 4, routePublicKeys, privateKey)
	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeContinue,
		NumNearRelays:      2,
		NearRelayIDs:       []uint64{1, 2},
		NearRelayAddresses: []net.UDPAddr{*relayAddr1, *relayAddr2},
		NumTokens:          4,
		Tokens:             tokenData,
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         false,
		RouteNumRelays:  2,
		RouteCost:       50,
		RouteRelayIDs:   [routing.MaxRelays]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	// We can't check if the entire response is equal since the response's tokens will be different each time
	// since the encryption generates random bytes for the nonce
	assert.Equal(t, expectedResponse.SessionID, responsePacket.SessionID)
	assert.Equal(t, expectedResponse.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, expectedResponse.RouteType, responsePacket.RouteType)
	assert.Equal(t, expectedResponse.NumNearRelays, responsePacket.NumNearRelays)
	assert.Equal(t, expectedResponse.NearRelayIDs, responsePacket.NearRelayIDs)
	assert.Equal(t, expectedResponse.NearRelayAddresses, responsePacket.NearRelayAddresses)
	assert.Equal(t, expectedResponse.NumTokens, responsePacket.NumTokens)

	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandler4RouteNoLongerExists(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackend4Metrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdate4Metrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)

	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData4{
		Version:         transport.SessionDataVersion4,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [routing.MaxRelays]uint64{5, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		Committed:            true,
		DirectRTT:            60,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []float32{10, 15},
		NearRelayJitter:      []float32{0, 0},
		NearRelayPacketLoss:  []float32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix4{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						0, 1,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1)},
			},
		},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expireTimestamp := uint64(time.Now().Unix()) + billing.BillingSliceSeconds*2
	sessionVersion := sessionDataStruct.SessionVersion + 1

	tokenData := make([]byte, core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES*4)
	routeAddresses := make([]*net.UDPAddr, 0)
	routeAddresses = append(routeAddresses, clientAddr, relayAddr1, relayAddr2, serverAddr)
	routePublicKeys := make([][]byte, 0)
	routePublicKeys = append(routePublicKeys, publicKey, publicKey, publicKey, publicKey)
	core.WriteRouteTokens(tokenData, expireTimestamp, requestPacket.SessionID, uint8(sessionVersion), 1024, 1024, 4, routeAddresses, routePublicKeys, privateKey)
	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeNew,
		NumNearRelays:      2,
		NearRelayIDs:       []uint64{1, 2},
		NearRelayAddresses: []net.UDPAddr{*relayAddr1, *relayAddr2},
		NumTokens:          4,
		Tokens:             tokenData,
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         true,
		RouteNumRelays:  2,
		RouteCost:       45,
		RouteRelayIDs:   [routing.MaxRelays]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	// We can't check if the entire response is equal since the response's tokens will be different each time
	// since the encryption generates random bytes for the nonce
	assert.Equal(t, expectedResponse.SessionID, responsePacket.SessionID)
	assert.Equal(t, expectedResponse.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, expectedResponse.RouteType, responsePacket.RouteType)
	assert.Equal(t, expectedResponse.NumNearRelays, responsePacket.NumNearRelays)
	assert.Equal(t, expectedResponse.NearRelayIDs, responsePacket.NearRelayIDs)
	assert.Equal(t, expectedResponse.NearRelayAddresses, responsePacket.NearRelayAddresses)
	assert.Equal(t, expectedResponse.NumTokens, responsePacket.NumTokens)

	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandler4RouteSwitched(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackend4Metrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdate4Metrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)

	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData4{
		Version:         transport.SessionDataVersion4,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [routing.MaxRelays]uint64{1, 2},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		Committed:            true,
		DirectRTT:            60,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []float32{10, 15},
		NearRelayJitter:      []float32{0, 0},
		NearRelayPacketLoss:  []float32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix4{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						0, 1,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1)},
			},
		},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expireTimestamp := uint64(time.Now().Unix()) + billing.BillingSliceSeconds*2
	sessionVersion := sessionDataStruct.SessionVersion + 1

	tokenData := make([]byte, core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES*4)
	routeAddresses := make([]*net.UDPAddr, 0)
	routeAddresses = append(routeAddresses, clientAddr, relayAddr1, relayAddr2, serverAddr)
	routePublicKeys := make([][]byte, 0)
	routePublicKeys = append(routePublicKeys, publicKey, publicKey, publicKey, publicKey)
	core.WriteRouteTokens(tokenData, expireTimestamp, requestPacket.SessionID, uint8(sessionVersion), 1024, 1024, 4, routeAddresses, routePublicKeys, privateKey)
	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeNew,
		NumNearRelays:      2,
		NearRelayIDs:       []uint64{1, 2},
		NearRelayAddresses: []net.UDPAddr{*relayAddr1, *relayAddr2},
		NumTokens:          4,
		Tokens:             tokenData,
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         true,
		RouteNumRelays:  2,
		RouteCost:       45,
		RouteRelayIDs:   [routing.MaxRelays]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	// We can't check if the entire response is equal since the response's tokens will be different each time
	// since the encryption generates random bytes for the nonce
	assert.Equal(t, expectedResponse.SessionID, responsePacket.SessionID)
	assert.Equal(t, expectedResponse.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, expectedResponse.RouteType, responsePacket.RouteType)
	assert.Equal(t, expectedResponse.NumNearRelays, responsePacket.NumNearRelays)
	assert.Equal(t, expectedResponse.NearRelayIDs, responsePacket.NearRelayIDs)
	assert.Equal(t, expectedResponse.NearRelayAddresses, responsePacket.NearRelayAddresses)
	assert.Equal(t, expectedResponse.NumTokens, responsePacket.NumTokens)

	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandler4VetoNoRoute(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData4{
		Version:         transport.SessionDataVersion4,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [routing.MaxRelays]uint64{2, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		DirectRTT:            60,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []float32{10, 15},
		NearRelayJitter:      []float32{0, 0},
		NearRelayPacketLoss:  []float32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix4{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		RouteEntries:       []core.RouteEntry{{}},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NumNearRelays:      2,
		NearRelayIDs:       []uint64{1, 2},
		NearRelayAddresses: []net.UDPAddr{*relayAddr1, *relayAddr2},
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		Initial:         false,
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Veto:          true,
			NoRoute:       true,
			ReduceLatency: true,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	// We can't check if the entire response is equal since the response's tokens will be different each time
	// since the encryption generates random bytes for the nonce
	assert.Equal(t, expectedResponse.SessionID, responsePacket.SessionID)
	assert.Equal(t, expectedResponse.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, expectedResponse.RouteType, responsePacket.RouteType)
	assert.Equal(t, expectedResponse.NumNearRelays, responsePacket.NumNearRelays)
	assert.Equal(t, expectedResponse.NearRelayIDs, responsePacket.NearRelayIDs)
	assert.Equal(t, expectedResponse.NearRelayAddresses, responsePacket.NearRelayAddresses)
	assert.Equal(t, expectedResponse.NumTokens, responsePacket.NumTokens)

	assert.Equal(t, 1.0, metrics.SessionUpdateMetrics.NoRoute.Value())
}

func TestSessionUpdateHandler4VetoMultipathOverloaded(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	buyer := routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	}
	err = storer.AddBuyer(context.Background(), buyer)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData4{
		Version:         transport.SessionDataVersion4,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [routing.MaxRelays]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        1234567890,
			Next:          true,
			ReduceLatency: true,
			Multipath:     true,
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		UserHash:             sessionDataStruct.RouteState.UserID,
		DirectRTT:            500,
		Next:                 true,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []float32{10, 15},
		NearRelayJitter:      []float32{0, 0},
		NearRelayPacketLoss:  []float32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix4{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						0, 1,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1)},
			},
		},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NumNearRelays:      2,
		NearRelayIDs:       []uint64{1, 2},
		NearRelayAddresses: []net.UDPAddr{*relayAddr1, *relayAddr2},
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		Initial:         false,
		RouteState: core.RouteState{
			UserID:            requestPacket.UserHash,
			Veto:              true,
			Multipath:         true,
			MultipathOverload: true,
			ReduceLatency:     true,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	// We can't check if the entire response is equal since the response's tokens will be different each time
	// since the encryption generates random bytes for the nonce
	assert.Equal(t, expectedResponse.SessionID, responsePacket.SessionID)
	assert.Equal(t, expectedResponse.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, expectedResponse.RouteType, responsePacket.RouteType)
	assert.Equal(t, expectedResponse.NumNearRelays, responsePacket.NumNearRelays)
	assert.Equal(t, expectedResponse.NearRelayIDs, responsePacket.NearRelayIDs)
	assert.Equal(t, expectedResponse.NearRelayAddresses, responsePacket.NearRelayAddresses)
	assert.Equal(t, expectedResponse.NumTokens, responsePacket.NumTokens)

	assert.Equal(t, 1.0, metrics.SessionUpdateMetrics.MultipathOverload.Value())
}

func TestSessionUpdateHandler4VetoLatencyWorse(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData4{
		Version:         transport.SessionDataVersion4,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [routing.MaxRelays]uint64{2, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		DirectRTT:            60,
		Next:                 true,
		NextRTT:              80,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []float32{10, 15},
		NearRelayJitter:      []float32{0, 0},
		NearRelayPacketLoss:  []float32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix4{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						0, 1,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1)},
			},
		},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NumNearRelays:      2,
		NearRelayIDs:       []uint64{1, 2},
		NearRelayAddresses: []net.UDPAddr{*relayAddr1, *relayAddr2},
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		Initial:         false,
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Veto:          true,
			ReduceLatency: true,
			LatencyWorse:  true,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	// We can't check if the entire response is equal since the response's tokens will be different each time
	// since the encryption generates random bytes for the nonce
	assert.Equal(t, expectedResponse.SessionID, responsePacket.SessionID)
	assert.Equal(t, expectedResponse.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, expectedResponse.RouteType, responsePacket.RouteType)
	assert.Equal(t, expectedResponse.NumNearRelays, responsePacket.NumNearRelays)
	assert.Equal(t, expectedResponse.NearRelayIDs, responsePacket.NearRelayIDs)
	assert.Equal(t, expectedResponse.NearRelayAddresses, responsePacket.NearRelayAddresses)
	assert.Equal(t, expectedResponse.NumTokens, responsePacket.NumTokens)

	assert.Equal(t, 1.0, metrics.SessionUpdateMetrics.LatencyWorse.Value())
}

func TestSessionUpdateHandler4CommitPending(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackend4Metrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdate4Metrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)

	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	internalConfig := core.NewInternalConfig()
	internalConfig.TryBeforeYouBuy = true
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: internalConfig,
	})
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData4{
		Version:         transport.SessionDataVersion4,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [routing.MaxRelays]uint64{2, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			CommitPending: true,
			CommitCounter: 1,
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		DirectRTT:            60,
		Next:                 true,
		NextRTT:              62,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []float32{10, 15},
		NearRelayJitter:      []float32{0, 0},
		NearRelayPacketLoss:  []float32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix4{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						1, 0,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(1, 0)},
			},
		},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expireTimestamp := uint64(time.Now().Unix()) + billing.BillingSliceSeconds

	tokenData := make([]byte, core.NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES*4)
	routePublicKeys := make([][]byte, 0)
	routePublicKeys = append(routePublicKeys, publicKey, publicKey, publicKey, publicKey)
	core.WriteContinueTokens(tokenData, expireTimestamp, requestPacket.SessionID, uint8(sessionDataStruct.SessionVersion), 4, routePublicKeys, privateKey)
	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeContinue,
		NumNearRelays:      2,
		NearRelayIDs:       []uint64{1, 2},
		NearRelayAddresses: []net.UDPAddr{*relayAddr1, *relayAddr2},
		NumTokens:          4,
		Tokens:             tokenData,
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         false,
		RouteNumRelays:  2,
		RouteCost:       50,
		RouteRelayIDs:   [routing.MaxRelays]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			CommitPending: true,
			CommitCounter: 2,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	// We can't check if the entire response is equal since the response's tokens will be different each time
	// since the encryption generates random bytes for the nonce
	assert.Equal(t, expectedResponse.SessionID, responsePacket.SessionID)
	assert.Equal(t, expectedResponse.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, expectedResponse.RouteType, responsePacket.RouteType)
	assert.Equal(t, expectedResponse.NumNearRelays, responsePacket.NumNearRelays)
	assert.Equal(t, expectedResponse.NearRelayIDs, responsePacket.NearRelayIDs)
	assert.Equal(t, expectedResponse.NearRelayAddresses, responsePacket.NearRelayAddresses)
	assert.Equal(t, expectedResponse.NumTokens, responsePacket.NumTokens)

	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandler4CommitVeto(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackend4Metrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdate4Metrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.DirectSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.DirectSlices.Add(1)

	metrics, err := metrics.NewServerBackend4Metrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	internalConfig := core.NewInternalConfig()
	internalConfig.TryBeforeYouBuy = true
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: internalConfig,
	})
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData4{
		Version:         transport.SessionDataVersion4,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [routing.MaxRelays]uint64{2, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			CommitPending: true,
			CommitCounter: 3,
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket4{
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		DirectRTT:            60,
		Next:                 true,
		NextRTT:              62,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []float32{10, 15},
		NearRelayJitter:      []float32{0, 0},
		NearRelayPacketLoss:  []float32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix4{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						1, 0,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(1, 0)},
			},
		},
	}
	routeMatrixFunc := func() *routing.RouteMatrix4 {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket4{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NumNearRelays:      2,
		NearRelayIDs:       []uint64{1, 2},
		NearRelayAddresses: []net.UDPAddr{*relayAddr1, *relayAddr2},
	}

	expectedSessionData := transport.SessionData4{
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		Initial:         false,
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Veto:          true,
			ReduceLatency: true,
			CommitVeto:    true,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc4(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket4
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData4
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	// We can't check if the entire response is equal since the response's tokens will be different each time
	// since the encryption generates random bytes for the nonce
	assert.Equal(t, expectedResponse.SessionID, responsePacket.SessionID)
	assert.Equal(t, expectedResponse.SliceNumber, responsePacket.SliceNumber)
	assert.Equal(t, expectedResponse.RouteType, responsePacket.RouteType)
	assert.Equal(t, expectedResponse.NumNearRelays, responsePacket.NumNearRelays)
	assert.Equal(t, expectedResponse.NearRelayIDs, responsePacket.NearRelayIDs)
	assert.Equal(t, expectedResponse.NearRelayAddresses, responsePacket.NearRelayAddresses)
	assert.Equal(t, expectedResponse.NumTokens, responsePacket.NumTokens)

	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}
