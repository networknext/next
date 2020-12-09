package transport_test

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"errors"
	"math/rand"
	"net"
	"testing"
	"time"

	"golang.org/x/crypto/nacl/box"

	"github.com/alicebob/miniredis"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
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

func assertResponseEqual(t *testing.T, expectedResponse transport.SessionResponsePacket, actualResponse transport.SessionResponsePacket) {
	// We can't check if the entire response is equal since the response's tokens will be different each time
	// since the encryption generates random bytes for the nonce
	assert.Equal(t, expectedResponse.SessionID, actualResponse.SessionID)
	assert.Equal(t, expectedResponse.SliceNumber, actualResponse.SliceNumber)
	assert.Equal(t, expectedResponse.RouteType, actualResponse.RouteType)
	assert.Equal(t, expectedResponse.NumNearRelays, actualResponse.NumNearRelays)
	assert.Equal(t, expectedResponse.NearRelayIDs, actualResponse.NearRelayIDs)
	assert.Equal(t, expectedResponse.NearRelayAddresses, actualResponse.NearRelayAddresses)
	assert.Equal(t, expectedResponse.NumTokens, actualResponse.NumTokens)
	assert.Equal(t, expectedResponse.HasDebug, actualResponse.HasDebug)

	if expectedResponse.HasDebug {
		assert.NotEmpty(t, actualResponse.Debug)
	} else {
		assert.Empty(t, actualResponse.Debug)
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

func TestSessionUpdateHandlerReadPacketFailure(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	handler := transport.SessionUpdateHandlerFunc(logger, nil, nil, nil, nil, 32, [crypto.KeySize]byte{}, nil, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: nil,
	})

	assert.Nil(t, responseBuffer.Bytes())
	assert.Equal(t, metrics.SessionUpdateMetrics.ReadPacketFailure.Value(), 1.0)
}

func TestSessionUpdateHandlerClientPingTimedOut(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 2}, // can only be true in 4.0.2 or higher
		SessionID:            1111,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
		ClientPingTimedOut:   true,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var badIPLocator badIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &badIPLocator
	}

	var routeMatrix routing.RouteMatrix
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       make([]uint64, 0),
		NearRelayAddresses: make([]net.UDPAddr, 0),
	}

	expectedSessionData := transport.SessionData{}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.ClientPingTimedOut.Value(), 1.0)
}

func TestSessionUpdateHandlerBuyerNotFound(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}

	requestPacket := transport.SessionUpdatePacket{
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

	var routeMatrix routing.RouteMatrix
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       make([]uint64, 0),
		NearRelayAddresses: make([]net.UDPAddr, 0),
	}

	expectedSessionData := transport.SessionData{}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.BuyerNotFound.Value(), 1.0)
}

func TestSessionUpdateHandlerDatacenterNotFound(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{Live: true})

	requestPacket := transport.SessionUpdatePacket{
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

	var routeMatrix routing.RouteMatrix
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       []uint64{},
		NearRelayAddresses: []net.UDPAddr{},
	}

	expectedSessionData := transport.SessionData{}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.DatacenterNotFound.Value(), 1.0)
}

func TestSessionUpdateHandlerMisconfiguredDatacenterAlias(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{Live: true})

	datacenterAlias := "alias"
	storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{Alias: datacenterAlias})

	requestPacket := transport.SessionUpdatePacket{
		SessionID:            1111,
		DatacenterID:         crypto.HashID(datacenterAlias),
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	var routeMatrix routing.RouteMatrix
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       []uint64{},
		NearRelayAddresses: []net.UDPAddr{},
	}

	expectedSessionData := transport.SessionData{}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.MisconfiguredDatacenterAlias.Value(), 1.0)
}

func TestSessionUpdateHandlerDatacenterNotAllowed(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{Live: true})

	datacenterName := "datacenter.name"
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: crypto.HashID(datacenterName)})

	requestPacket := transport.SessionUpdatePacket{
		SessionID:            1111,
		DatacenterID:         crypto.HashID(datacenterName),
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	var routeMatrix routing.RouteMatrix
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       []uint64{},
		NearRelayAddresses: []net.UDPAddr{},
	}

	expectedSessionData := transport.SessionData{}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.DatacenterNotAllowed.Value(), 1.0)
}

func TestSessionUpdateHandlerClientLocateFailure(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{})
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)
	storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{})

	requestPacket := transport.SessionUpdatePacket{
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

	var routeMatrix routing.RouteMatrix
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       make([]uint64, 0),
		NearRelayAddresses: make([]net.UDPAddr, 0),
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
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
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.ClientLocateFailure.Value(), 1.0)
}

func TestSessionUpdateHandlerReadSessionDataFailure(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{})
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)
	storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{})

	requestPacket := transport.SessionUpdatePacket{
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

	var routeMatrix routing.RouteMatrix
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       []uint64{},
		NearRelayAddresses: []net.UDPAddr{},
	}

	expectedSessionData := transport.SessionData{}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.ReadSessionDataFailure.Value(), 1.0)
}

func TestSessionUpdateHandlerSessionDataBadSessionID(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{})
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)
	storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{})

	sessionDataStruct := transport.SessionData{
		Version:     transport.SessionDataVersion,
		SessionID:   1,
		SliceNumber: 1,
		Location:    routing.LocationNullIsland,
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	requestPacket := transport.SessionUpdatePacket{
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

	var routeMatrix routing.RouteMatrix
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       []uint64{},
		NearRelayAddresses: []net.UDPAddr{},
	}

	expectedSessionData := transport.SessionData{
		Version:     transport.SessionDataVersion,
		SessionID:   1,
		SliceNumber: 1,
		Location:    routing.LocationNullIsland,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.BadSessionID.Value(), 1.0)
}

func TestSessionUpdateHandlerSessionDataBadSliceNumber(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{})
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)
	storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{})

	sessionDataStruct := transport.SessionData{
		Version:     transport.SessionDataVersion,
		SessionID:   1111,
		SliceNumber: 1,
		Location:    routing.LocationNullIsland,
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	requestPacket := transport.SessionUpdatePacket{
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

	var routeMatrix routing.RouteMatrix
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       []uint64{},
		NearRelayAddresses: []net.UDPAddr{},
	}

	expectedSessionData := transport.SessionData{
		Version:     transport.SessionDataVersion,
		SessionID:   1111,
		SliceNumber: 1,
		Location:    routing.LocationNullIsland,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.BadSliceNumber.Value(), 1.0)
}

func TestSessionUpdateHandlerBuyerNotLive(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{})
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)
	storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{})

	requestPacket := transport.SessionUpdatePacket{
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

	var routeMatrix routing.RouteMatrix
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       make([]uint64, 0),
		NearRelayAddresses: make([]net.UDPAddr, 0),
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
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
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.BuyerNotLive.Value(), 1.0)
}

func TestSessionUpdateHandlerFallbackToDirect(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{Live: true})
	storer.AddDatacenter(context.Background(), routing.UnknownDatacenter)
	storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{})

	requestPacket := transport.SessionUpdatePacket{
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

	var routeMatrix routing.RouteMatrix
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       make([]uint64, 0),
		NearRelayAddresses: make([]net.UDPAddr, 0),
	}

	expectedSessionData := transport.SessionData{
		Version:          transport.SessionDataVersion,
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
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.FallbackToDirectUnknownReason.Value(), 1.0)
}

func TestSessionUpdateHandlerFirstSlice(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.DirectSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.DirectSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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
	storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 10})

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
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

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0},
		RelayIDs:           []uint64{1},
		RelayAddresses:     []net.UDPAddr{*relayAddr},
		RelayNames:         []string{"test.relay"},
		RelayLatitudes:     []float32{90},
		RelayLongitudes:    []float32{180},
		RelayDatacenterIDs: []uint64{10},
	}
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		Version:            requestPacket.Version,
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NumNearRelays:      1,
		NearRelayIDs:       []uint64{1},
		NearRelayAddresses: []net.UDPAddr{*relayAddr},
		NearRelaysChanged:  true,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		RouteState: core.RouteState{
			NumNearRelays: 1,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version // Do this as a sort of hack to read in the debug values just like SDK 4.0.4 does
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandlerNoDestRelays(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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
	storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 10})

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	requestPacket := transport.SessionUpdatePacket{
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
		NumNearRelays:        1,
		NearRelayIDs:         []uint64{1},
		NearRelayRTT:         []int32{0},
		NearRelayJitter:      []int32{0},
		NearRelayPacketLoss:  []int32{0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{}
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		SessionID:          requestPacket.SessionID,
		SliceNumber:        requestPacket.SliceNumber,
		RouteType:          routing.RouteTypeDirect,
		NearRelayIDs:       []uint64{},
		NearRelayAddresses: []net.UDPAddr{},
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
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
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assert.Equal(t, metrics.SessionUpdateMetrics.NoRelaysInDatacenter.Value(), 1.0)
}

func TestSessionUpdateHandlerDirectRoute(t *testing.T) {
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.DirectSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.DirectSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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
	storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 10})

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		ServerRoutePublicKey: make([]byte, crypto.KeySize),
		NumNearRelays:        1,
		NearRelayIDs:         []uint64{1},
		NearRelayRTT:         []int32{0},
		NearRelayJitter:      []int32{0},
		NearRelayPacketLoss:  []int32{0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var unmarshaled transport.SessionUpdatePacket
	err = transport.UnmarshalPacket(&unmarshaled, requestData)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	relayAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0},
		RelayIDs:           []uint64{1},
		RelayAddresses:     []net.UDPAddr{*relayAddr},
		RelayNames:         []string{"test.relay"},
		RelayLatitudes:     []float32{90},
		RelayLongitudes:    []float32{180},
		RelayDatacenterIDs: []uint64{10},
	}
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeDirect,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		RouteState: core.RouteState{
			NumNearRelays:   1,
			NearRelayRTT:    [core.MaxNearRelays]int32{255},
			NearRelayJitter: [core.MaxNearRelays]int32{0},
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, [crypto.KeySize]byte{}, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)
	assert.Equal(t, expectedResponse, responsePacket)

	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandlerNextRoute(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
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
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteState: core.RouteState{
			NearRelayRTT: [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
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
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
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
	routeMatrixFunc := func() *routing.RouteMatrix {
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
	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeNew,
		NumTokens:   4,
		Tokens:      tokenData,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         true,
		RouteNumRelays:  2,
		RouteCost:       45 + core.CostBias,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
			NumNearRelays: 2,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
		EverOnNext: true,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandlerNextRouteExternalIPs(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 12})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 12})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "other_seller"})
	assert.NoError(t, err)

	relayAddr1External, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr1Internal, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	relayAddr2External, err := net.ResolveUDPAddr("udp", "127.0.0.1:10002")
	assert.NoError(t, err)
	relayAddr2Internal, err := net.ResolveUDPAddr("udp", "127.0.0.1:10003")
	assert.NoError(t, err)

	relayAddr3External, err := net.ResolveUDPAddr("udp", "127.0.0.1:10004")
	assert.NoError(t, err)
	relayAddr3Internal, err := net.ResolveUDPAddr("udp", "127.0.0.1:10005")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:           1,
		Addr:         *relayAddr1External,
		InternalAddr: *relayAddr1Internal,
		PublicKey:    publicKey,
		Seller:       routing.Seller{ID: "seller"},
		Datacenter:   routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:           2,
		Addr:         *relayAddr2External,
		InternalAddr: *relayAddr2Internal,
		PublicKey:    publicKey,
		Seller:       routing.Seller{ID: "other_seller"},
		Datacenter:   routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:           3,
		Addr:         *relayAddr3External,
		InternalAddr: *relayAddr3Internal,
		PublicKey:    publicKey,
		Seller:       routing.Seller{ID: "seller"},
		Datacenter:   routing.Datacenter{ID: 12},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteState: core.RouteState{
			NearRelayRTT: [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	assert.NoError(t, err)
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         12,
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
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1, 3: 2},
		RelayIDs:           []uint64{1, 2, 3},
		RelayAddresses:     []net.UDPAddr{*relayAddr1External, *relayAddr2External, *relayAddr3External},
		RelayNames:         []string{"test.relay.1", "test.relay.2", "test.relay.3"},
		RelayLatitudes:     []float32{90, 89, 88},
		RelayLongitudes:    []float32{180, 179, 178},
		RelayDatacenterIDs: []uint64{10, 11, 12},
		RouteEntries: []core.RouteEntry{
			// route entries identical so there's no randomness to account for
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{3},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						0, 1, 2,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1, 2)},
			},
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{3},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						0, 1, 2,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1, 2)},
			},
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{3},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						0, 1, 2,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1, 2)},
			},
		},
	}
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expireTimestamp := uint64(time.Now().Unix()) + billing.BillingSliceSeconds*2
	sessionVersion := sessionDataStruct.SessionVersion + 1

	tokenData := make([]byte, core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES*5)
	routeAddresses := make([]*net.UDPAddr, 0)
	routeAddresses = append(routeAddresses, clientAddr, relayAddr1External, relayAddr2External, relayAddr3External, serverAddr)
	routePublicKeys := make([][]byte, 0)
	routePublicKeys = append(routePublicKeys, publicKey, publicKey, publicKey, publicKey, publicKey)
	core.WriteRouteTokens(tokenData, expireTimestamp, requestPacket.SessionID, uint8(sessionVersion), 1024, 1024, 4, routeAddresses, routePublicKeys, privateKey)
	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeNew,
		NumTokens:   5,
		Tokens:      tokenData,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         true,
		RouteNumRelays:  3,
		RouteCost:       45 + core.CostBias,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{3, 2, 1},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
			NumNearRelays: 2,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
		EverOnNext: true,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestFeatureInternalIP(t *testing.T) {

	relayAddr1External, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr1Internal, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	relayAddr2External, err := net.ResolveUDPAddr("udp", "127.0.0.1:10002")
	assert.NoError(t, err)
	relayAddr2Internal, err := net.ResolveUDPAddr("udp", "127.0.0.1:10003")
	assert.NoError(t, err)

	relayAddr3External, err := net.ResolveUDPAddr("udp", "127.0.0.1:10004")
	assert.NoError(t, err)
	relayAddr3Internal, err := net.ResolveUDPAddr("udp", "127.0.0.1:10005")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	publicKeyArr, _, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)
	copy(publicKey, publicKeyArr[:])

	relayIDs := []uint64{0, 1, 2}
	seller := routing.Seller{ID: "seller_id", Name: "seller_name"}
	seller2 := routing.Seller{ID: "seller_id2", Name: "seller_name2"}

	relays := make([]routing.Relay, 3)
	relays[0] = routing.Relay{
		ID:           0,
		Addr:         *relayAddr1External,
		InternalAddr: *relayAddr1Internal,
		PublicKey:    publicKey,
		Seller:       seller,
		Datacenter:   routing.Datacenter{ID: 10},
	}

	relays[1] = routing.Relay{
		ID:           1,
		Addr:         *relayAddr2External,
		InternalAddr: *relayAddr2Internal,
		PublicKey:    publicKey,
		Seller:       seller,
		Datacenter:   routing.Datacenter{ID: 11},
	}

	relays[2] = routing.Relay{
		ID:           2,
		Addr:         *relayAddr3External,
		InternalAddr: *relayAddr3Internal,
		PublicKey:    publicKey,
		Seller:       seller2,
		Datacenter:   routing.Datacenter{ID: 12},
	}
	var storer storage.Storer
	storer = &storage.StorerMock{RelayFunc: func(id uint64) (routing.Relay, error) {
		return relays[id], nil
	}}

	routeRelays := []int32{0, 1, 2}

	//feature off
	routeAddressesOff := make([]*net.UDPAddr, 4)
	for i := int32(0); i < 3; i++ {
		routeAddressesOff = transport.AddAddress(false, i, relays[i], relayIDs, storer, routeRelays, routeAddressesOff)
	}

	assert.Equal(t, relays[0].Addr.String(), routeAddressesOff[1].String())
	assert.Equal(t, relays[1].Addr.String(), routeAddressesOff[2].String())
	assert.Equal(t, relays[2].Addr.String(), routeAddressesOff[3].String())

	//feature off
	routeAddressesOn := make([]*net.UDPAddr, 4)
	for i := int32(0); i < 3; i++ {
		routeAddressesOn = transport.AddAddress(true, i, relays[i], relayIDs, storer, routeRelays, routeAddressesOn)
	}

	assert.Equal(t, relays[0].Addr.String(), routeAddressesOn[1].String())
	assert.Equal(t, relays[1].InternalAddr.String(), routeAddressesOn[2].String())
	assert.Equal(t, relays[2].Addr.String(), routeAddressesOn[3].String())

}

//todo: test does not currently work with inline feature flag set to false.
//func TestSessionUpdateHandlerNextRouteInternalIPs(t *testing.T) {
//	// Seed the RNG so we don't get different results from running `make test`
//	// and running the test directly in VSCode
//	rand.Seed(0)
//	logger := log.NewNopLogger()
//	metricsHandler := metrics.LocalHandler{}
//
//	expectedMetrics := metrics.EmptyServerBackendMetrics
//	var err error
//	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
//	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
//	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
//	assert.NoError(t, err)
//	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)
//
//	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
//	assert.NoError(t, err)
//	responseBuffer := bytes.NewBuffer(nil)
//	storer := &storage.InMemory{}
//	err = storer.AddBuyer(context.Background(), routing.Buyer{
//		ID:             100,
//		Live:           true,
//		RouteShader:    core.NewRouteShader(),
//		InternalConfig: core.NewInternalConfig(),
//	})
//	assert.NoError(t, err)
//
//	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
//	assert.NoError(t, err)
//
//	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
//	assert.NoError(t, err)
//
//	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 12})
//	assert.NoError(t, err)
//
//	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 12})
//	assert.NoError(t, err)
//
//	seller := routing.Seller{ID: "seller_id", Name: "seller_name"}
//	err = storer.AddSeller(context.Background(), seller)
//	assert.NoError(t, err)
//
//	relayAddr1External, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
//	assert.NoError(t, err)
//	relayAddr1Internal, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
//	assert.NoError(t, err)
//
//	relayAddr2External, err := net.ResolveUDPAddr("udp", "127.0.0.1:10002")
//	assert.NoError(t, err)
//	relayAddr2Internal, err := net.ResolveUDPAddr("udp", "127.0.0.1:10003")
//	assert.NoError(t, err)
//
//	relayAddr3External, err := net.ResolveUDPAddr("udp", "127.0.0.1:10004")
//	assert.NoError(t, err)
//	relayAddr3Internal, err := net.ResolveUDPAddr("udp", "127.0.0.1:10005")
//	assert.NoError(t, err)
//
//	publicKey := make([]byte, crypto.KeySize)
//	publicKeyArr, privateKey, err := box.GenerateKey(crand.Reader)
//	assert.NoError(t, err)
//	copy(publicKey, publicKeyArr[:])
//
//	err = storer.AddRelay(context.Background(), routing.Relay{
//		ID:           1,
//		Addr:         *relayAddr1External,
//		InternalAddr: *relayAddr1Internal,
//		PublicKey:    publicKey,
//		Seller:       seller,
//		Datacenter:   routing.Datacenter{ID: 10},
//	})
//	assert.NoError(t, err)
//
//	err = storer.AddRelay(context.Background(), routing.Relay{
//		ID:           2,
//		Addr:         *relayAddr2External,
//		InternalAddr: *relayAddr2Internal,
//		PublicKey:    publicKey,
//		Seller:       seller,
//		Datacenter:   routing.Datacenter{ID: 11},
//	})
//	assert.NoError(t, err)
//
//	err = storer.AddRelay(context.Background(), routing.Relay{
//		ID:           3,
//		Addr:         *relayAddr3External,
//		InternalAddr: *relayAddr3Internal,
//		PublicKey:    publicKey,
//		Seller:       seller,
//		Datacenter:   routing.Datacenter{ID: 12},
//	})
//	assert.NoError(t, err)
//
//	sessionDataStruct := transport.SessionData{
//		Version:         transport.SessionDataVersion,
//		SessionID:       1111,
//		SliceNumber:     1,
//		Location:        routing.LocationNullIsland,
//		ExpireTimestamp: uint64(time.Now().Unix()),
//		RouteState: core.RouteState{
//			NearRelayRTT: [core.MaxNearRelays]int32{10, 15},
//		},
//	}
//
//	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
//	assert.NoError(t, err)
//
//	sessionDataArray := [transport.MaxSessionDataSize]byte{}
//	copy(sessionDataArray[:], sessionDataSlice)
//
//	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
//	assert.NoError(t, err)
//	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
//	assert.NoError(t, err)
//
//	requestPacket := transport.SessionUpdatePacket{
//		Version:              transport.SDKVersion{4, 0, 4},
//		SessionID:            1111,
//		CustomerID:           100,
//		DatacenterID:         12,
//		SliceNumber:          1,
//		SessionDataBytes:     int32(len(sessionDataSlice)),
//		SessionData:          sessionDataArray,
//		ClientAddress:        *clientAddr,
//		ServerAddress:        *serverAddr,
//		ClientRoutePublicKey: publicKey,
//		ServerRoutePublicKey: publicKey,
//		DirectRTT:            60,
//		NumNearRelays:        2,
//		NearRelayIDs:         []uint64{1, 2},
//		NearRelayRTT:         []int32{10, 15},
//		NearRelayJitter:      []int32{0, 0},
//		NearRelayPacketLoss:  []int32{0, 0},
//	}
//	requestData, err := transport.MarshalPacket(&requestPacket)
//	assert.NoError(t, err)
//
//	var goodIPLocator goodIPLocator
//	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
//		return &goodIPLocator
//	}
//
//	routeMatrix := routing.RouteMatrix{
//		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1, 3: 2},
//		RelayIDs:           []uint64{1, 2, 3},
//		RelayAddresses:     []net.UDPAddr{*relayAddr1External, *relayAddr2External, *relayAddr3External},
//		RelayNames:         []string{"test.relay.1", "test.relay.2", "test.relay.3"},
//		RelayLatitudes:     []float32{90, 89, 88},
//		RelayLongitudes:    []float32{180, 179, 178},
//		RelayDatacenterIDs: []uint64{10, 11, 12},
//		RouteEntries: []core.RouteEntry{
//			// route entries identical so there's no randomness to account for
//			{
//				DirectCost:     65,
//				NumRoutes:      int32(core.TriMatrixLength(2)),
//				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
//				RouteNumRelays: [core.MaxRoutesPerEntry]int32{3},
//				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
//					{
//						0, 1, 2,
//					},
//				},
//				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1, 2)},
//			},
//			{
//				DirectCost:     65,
//				NumRoutes:      int32(core.TriMatrixLength(2)),
//				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
//				RouteNumRelays: [core.MaxRoutesPerEntry]int32{3},
//				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
//					{
//						0, 1, 2,
//					},
//				},
//				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1, 2)},
//			},
//			{
//				DirectCost:     65,
//				NumRoutes:      int32(core.TriMatrixLength(2)),
//				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
//				RouteNumRelays: [core.MaxRoutesPerEntry]int32{3},
//				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
//					{
//						0, 1, 2,
//					},
//				},
//				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1, 2)},
//			},
//		},
//	}
//	routeMatrixFunc := func() *routing.RouteMatrix {
//		return &routeMatrix
//	}
//
//	redisServer, err := miniredis.Run()
//	assert.NoError(t, err)
//
//	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
//	assert.NoError(t, err)
//
//	expireTimestamp := uint64(time.Now().Unix()) + billing.BillingSliceSeconds*2
//	sessionVersion := sessionDataStruct.SessionVersion + 1
//
//	tokenData := make([]byte, core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES*5)
//	routeAddresses := make([]*net.UDPAddr, 0)
//	routeAddresses = append(routeAddresses, clientAddr, relayAddr3External, relayAddr2Internal, relayAddr1Internal, serverAddr)
//	routePublicKeys := make([][]byte, 0)
//	routePublicKeys = append(routePublicKeys, publicKey, publicKey, publicKey, publicKey, publicKey)
//	core.WriteRouteTokens(tokenData, expireTimestamp, requestPacket.SessionID, uint8(sessionVersion), 1024, 1024, 4, routeAddresses, routePublicKeys, *privateKey)
//	expectedResponse := transport.SessionResponsePacket{
//		Version:     requestPacket.Version,
//		SessionID:   requestPacket.SessionID,
//		SliceNumber: requestPacket.SliceNumber,
//		RouteType:   routing.RouteTypeNew,
//		NumTokens:   5,
//		Tokens:      tokenData,
//	}
//
//	expectedSessionData := transport.SessionData{
//		Version:         transport.SessionDataVersion,
//		SessionID:       requestPacket.SessionID,
//		SessionVersion:  sessionVersion,
//		SliceNumber:     requestPacket.SliceNumber + 1,
//		Location:        routing.LocationNullIsland,
//		ExpireTimestamp: expireTimestamp,
//		Initial:         true,
//		RouteNumRelays:  3,
//		RouteCost:       45 + core.CostBias,
//		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{3, 2, 1},
//		RouteState: core.RouteState{
//			UserID:        requestPacket.UserHash,
//			Next:          true,
//			ReduceLatency: true,
//			Committed:     true,
//			NumNearRelays: 2,
//			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
//		},
//		EverOnNext: true,
//	}
//
//	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
//	assert.NoError(t, err)
//
//	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
//	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)
//
//	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
//	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, *privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
//	handler(responseBuffer, &transport.UDPPacket{
//		Data: requestData,
//	})
//
//	var responsePacket transport.SessionResponsePacket
//	responsePacket.Version = requestPacket.Version
//	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
//	assert.NoError(t, err)
//
//	var sessionData transport.SessionData
//	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
//	assert.NoError(t, err)
//
//	assert.Equal(t, expectedSessionData, sessionData)
//
//	assertResponseEqual(t, expectedResponse, responsePacket)
//
//	assert.Equal(t, 5*core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES, len(responsePacket.Tokens))
//
//	var clientToken core.RouteToken
//	assert.NoError(t, core.ReadEncryptedRouteToken(&clientToken, responsePacket.Tokens[0*core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], publicKey, privateKey[:]))
//
//	var relay1Token core.RouteToken
//	assert.NoError(t, core.ReadEncryptedRouteToken(&relay1Token, responsePacket.Tokens[1*core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], publicKey, privateKey[:]))
//
//	var relay2Token core.RouteToken
//	assert.NoError(t, core.ReadEncryptedRouteToken(&relay2Token, responsePacket.Tokens[2*core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], publicKey, privateKey[:]))
//
//	var relay3Token core.RouteToken
//	assert.NoError(t, core.ReadEncryptedRouteToken(&relay3Token, responsePacket.Tokens[3*core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], publicKey, privateKey[:]))
//
//	assert.Equal(t, routeAddresses[1], clientToken.NextAddress)
//	assert.Equal(t, routeAddresses[2], relay1Token.NextAddress)
//	assert.Equal(t, routeAddresses[3], relay2Token.NextAddress)
//	assert.Equal(t, routeAddresses[4], relay3Token.NextAddress)
//
//	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
//}

func TestSessionUpdateHandlerContinueRoute(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
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
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
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
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
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
	routeMatrixFunc := func() *routing.RouteMatrix {
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
	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeContinue,
		NumTokens:   4,
		Tokens:      tokenData,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         false,
		RouteNumRelays:  2,
		RouteCost:       50 + core.CostBias,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
			NumNearRelays: 2,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
		EverOnNext: true,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandlerRouteNoLongerExists(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{ID: "next_slices"})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.RouteDoesNotExist, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{ID: "route_does_not_exist"})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)
	expectedMetrics.SessionUpdateMetrics.RouteDoesNotExist.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
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
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{5, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
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
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
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
	routeMatrixFunc := func() *routing.RouteMatrix {
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
	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeNew,
		NumTokens:   4,
		Tokens:      tokenData,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         true,
		RouteNumRelays:  2,
		RouteCost:       45 + core.CostBias,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
			NumNearRelays: 2,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
			RelayWentAway: true,
		},
		EverOnNext: true,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandlerRouteSwitched(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{ID: "next_slices"})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.RouteSwitched, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{ID: "route_switched"})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)
	expectedMetrics.SessionUpdateMetrics.RouteSwitched.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
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
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{1, 2},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
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
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
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
	routeMatrixFunc := func() *routing.RouteMatrix {
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
	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeNew,
		NumTokens:   4,
		Tokens:      tokenData,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         true,
		RouteNumRelays:  2,
		RouteCost:       45 + core.CostBias,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			NumNearRelays: 2,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
			RouteLost:     true,
		},
		EverOnNext: true,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandlerVetoNoRoute(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
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
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
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
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
		RouteEntries:       []core.RouteEntry{{}},
	}
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeDirect,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
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
			NumNearRelays: 2,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
			RouteLost:     true,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assert.Equal(t, 1.0, metrics.SessionUpdateMetrics.NoRoute.Value())
}

func TestSessionUpdateHandlerVetoMultipathOverloaded(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
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
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        1234567890,
			Next:          true,
			ReduceLatency: true,
			Multipath:     true,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
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
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
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
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeDirect,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
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
			NumNearRelays:     2,
			NearRelayRTT:      [core.MaxNearRelays]int32{10, 15},
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assert.Equal(t, 1.0, metrics.SessionUpdateMetrics.MultipathOverload.Value())
}

func TestSessionUpdateHandlerVetoLatencyWorse(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
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
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
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
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
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
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeDirect,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		Initial:         false,
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Veto:          true,
			Committed:     true,
			ReduceLatency: true,
			LatencyWorse:  true,
			NumNearRelays: 2,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assert.Equal(t, 1.0, metrics.SessionUpdateMetrics.LatencyWorse.Value())
}

func TestSessionUpdateHandlerCommitPending(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
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
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			CommitCounter: 1,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
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
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
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
	routeMatrixFunc := func() *routing.RouteMatrix {
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
	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeContinue,
		NumTokens:   4,
		Tokens:      tokenData,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         false,
		RouteNumRelays:  2,
		RouteCost:       50 + core.CostBias,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			CommitCounter: 2,
			NumNearRelays: 2,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
		EverOnNext: true,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandlerCommitVeto(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.DirectSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.DirectSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
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
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteNumRelays:  2,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			CommitCounter: 3,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
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
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
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
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeDirect,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
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
			CommitCounter: 4,
			CommitVeto:    true,
			NumNearRelays: 2,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateDebugResponse(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		Debug:          true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
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
		Name:       "test.relay.1",
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Name:       "test.relay.2",
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteState: core.RouteState{
			NearRelayRTT: [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		DirectRTT:            80,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
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
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeNew,
		NumTokens:   4,
		HasDebug:    true,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion + 1,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds*2,
		Initial:         true,
		RouteNumRelays:  2,
		RouteCost:       45 + core.CostBias,
		RouteRelayIDs:   [5]uint64{1, 2, 0, 0, 0},
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
			NumNearRelays: 2,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
		EverOnNext: true,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateDesyncedNearRelays(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.DirectSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.DirectSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
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
		Name:       "test.relay.1",
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Name:       "test.relay.2",
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteState: core.RouteState{
			NearRelayRTT: [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		DirectRTT:            80,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
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
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeDirect,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		RouteState: core.RouteState{
			UserID: requestPacket.UserHash,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateOneRelayInRouteMatrix(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.DirectSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.DirectSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
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

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 10})
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
		Name:       "test.relay.1",
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Name:       "test.relay.2",
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
		EverOnNext: true,
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
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
		DirectRTT:            80,
		Next:                 true,
		NextRTT:              60,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0},
		RelayIDs:           []uint64{1},
		RelayAddresses:     []net.UDPAddr{*relayAddr1},
		RelayNames:         []string{"test.relay.1"},
		RelayLatitudes:     []float32{90},
		RelayLongitudes:    []float32{180},
		RelayDatacenterIDs: []uint64{10},
	}
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeDirect,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		RouteState: core.RouteState{
			UserID:        requestPacket.UserHash,
			ReduceLatency: true,
			Committed:     true,
			NumNearRelays: 2,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 255},
		},
		EverOnNext: true,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}
