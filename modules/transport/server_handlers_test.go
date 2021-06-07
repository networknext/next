package transport_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/test"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func TestGetRouteAddressesAndPublicKeys(t *testing.T) {
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

	sellerMap := make(map[string]routing.Seller)
	sellerMap[seller.ID] = seller

	datacenterMap := make(map[uint64]routing.Datacenter)
	datacenterMap[datacenter.ID] = datacenter

	relayMap := make(map[uint64]routing.Relay)
	relayMap[crypto.HashID(relayAddr1.String())] = routing.Relay{ID: crypto.HashID(relayAddr1.String()), Addr: *relayAddr1, PublicKey: relayPublicKey1, Seller: seller, Datacenter: datacenter}
	relayMap[crypto.HashID(relayAddr2.String())] = routing.Relay{ID: crypto.HashID(relayAddr2.String()), Addr: *relayAddr2, PublicKey: relayPublicKey2, Seller: seller, Datacenter: datacenter}
	relayMap[crypto.HashID(relayAddr3.String())] = routing.Relay{ID: crypto.HashID(relayAddr3.String()), Addr: *relayAddr3, PublicKey: relayPublicKey3, Seller: seller, Datacenter: datacenter}

	database := routing.DatabaseBinWrapper{RelayMap: relayMap, SellerMap: sellerMap, DatacenterMap: datacenterMap}

	allRelayIDs := []uint64{crypto.HashID(relayAddr1.String()), crypto.HashID(relayAddr2.String()), crypto.HashID(relayAddr3.String())}
	routeRelays := []int32{0, 1, 2}

	routeAddresses, routePublicKeys := transport.GetRouteAddressesAndPublicKeys(clientAddr, clientPublicKey, serverAddr, serverPublicKey, 5, routeRelays, allRelayIDs, &database)

	expectedRouteAddresses := []*net.UDPAddr{clientAddr, relayAddr1, relayAddr2, relayAddr3, serverAddr}
	expectedRoutePublicKeys := [][]byte{clientPublicKey, relayPublicKey1, relayPublicKey2, relayPublicKey3, serverPublicKey}

	for i := range routeAddresses {
		assert.Equal(t, expectedRouteAddresses[i].String(), routeAddresses[i].String())
	}

	for i := range routePublicKeys {
		assert.Equal(t, expectedRoutePublicKeys[i], routePublicKeys[i])
	}
}

// todo: there should be a test here that verifies correct behavior with private relay addresses

// Server init handler tests

func TestServerInitHandlerFunc_BuyerNotFound(t *testing.T) {
	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false)

	unknownPublicKey, unknownPrivateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	unknownBuyerID := binary.LittleEndian.Uint64(unknownPublicKey[:8])

	unknownPublicKey = unknownPublicKey[8:]
	unknownPrivateKey = unknownPrivateKey[8:]

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, unknownBuyerID, datacenterID, datacenterName, unknownPrivateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseUnknownBuyer), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.BuyerNotFound.Value())
}

func TestServerInitHandlerFunc_BuyerNotLive(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", false)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, privateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseBuyerNotActive), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.BuyerNotActive.Value())
}

func TestServerInitHandlerFunc_SigCheckFail(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, privateKey[2:])

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseSignatureCheckFailed), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.SignatureCheckFailed.Value())
}

func TestServerInitHandlerFunc_SDKToOld(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersion{3, 0, 0}, buyerID, datacenterID, datacenterName, privateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseOldSDKVersion), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.SDKTooOld.Value())
}

func TestServerInitHandlerFunc_Success_DatacenterNotFound(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, privateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.DatacenterNotFound.Value())
}

func TestServerInitHandlerFunc_Success(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)
	datacenter := env.AddDatacenter("datacenter.name")

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, buyerID, datacenter.ID, datacenter.Name, privateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)
	assert.Equal(t, float64(0), metrics.ServerInitMetrics.DatacenterNotFound.Value())
}

// Server update handler tests

func TestServerUpdateHandlerFunc_BuyerNotFound(t *testing.T) {
	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false)

	unknownPublicKey, unknownPrivateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	unknownBuyerID := binary.LittleEndian.Uint64(unknownPublicKey[:8])

	unknownPublicKey = unknownPublicKey[8:]
	unknownPrivateKey = unknownPrivateKey[8:]

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersionLatest, unknownBuyerID, datacenterID, datacenterName, 10, "10.0.0.1", unknownPrivateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.BuyerNotFound.Value())
}

func TestServerUpdateHandlerFunc_BuyerNotLive(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", false)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, 10, "10.0.0.1", privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.BuyerNotLive.Value())
}

func TestServerUpdateHandlerFunc_SigCheckFail(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, 10, "10.0.0.1", privateKey[2:])

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.SignatureCheckFailed.Value())
}

func TestServerUpdateHandlerFunc_SDKToOld(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersion{3, 0, 0}, buyerID, datacenterID, datacenterName, 10, "10.0.0.1", privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.SDKTooOld.Value())
}

func TestServerUpdateHandlerFunc_DatacenterNotFound(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, 10, "10.0.0.1", privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.DatacenterNotFound.Value())
}

func TestServerUpdateHandlerFunc_Success(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)
	datacenter := env.AddDatacenter("datacenter.name")

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersionLatest, buyerID, datacenter.ID, datacenter.Name, 10, "10.0.0.1", privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(0), metrics.ServerUpdateMetrics.DatacenterNotFound.Value())
}

// Session update handler
func TestSessionUpdateHandlerFunc_Pre_BuyerNotFound(t *testing.T) {
	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false)

	unknownPublicKey, unknownPrivateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	unknownBuyerID := binary.LittleEndian.Uint64(unknownPublicKey[:8])
	unknownPublicKey = unknownPublicKey[8:]
	unknownPrivateKey = unknownPrivateKey[8:]

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)
	state.Packet.BuyerID = unknownBuyerID

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.BuyerNotFound)
	assert.Equal(t, float64(1), state.Metrics.BuyerNotFound.Value())
}

func TestSessionUpdateHandlerFunc_Pre_BuyerNotLive(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, _ := env.AddBuyer("local", false)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.BuyerNotLive)
	assert.Equal(t, float64(1), state.Metrics.BuyerNotLive.Value())
}

func TestSessionUpdateHandlerFunc_Pre_SigCheckFail(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey[2:])
	state.PacketData = requestData

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.SignatureCheckFailed)
	assert.Equal(t, float64(1), state.Metrics.SignatureCheckFailed.Value())
}

func TestSessionUpdateHandlerFunc_Pre_ClientTimedOut(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.Packet.ClientPingTimedOut = true

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.Packet.ClientPingTimedOut)
	assert.Equal(t, float64(1), state.Metrics.ClientPingTimedOut.Value())
}

func TestSessionUpdateHandlerFunc_Pre_DatacenterNotFound(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = crypto.HashID("unknown.datacenter.name")
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.UnknownDatacenter)
	assert.Equal(t, float64(1), state.Metrics.DatacenterNotFound.Value())
}

func TestSessionUpdateHandlerFunc_Pre_DatacenterNotEnabled(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)
	datacenter := env.AddDatacenter("datacenter.name")

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.DatacenterNotEnabled)
	assert.Equal(t, float64(1), state.Metrics.DatacenterNotEnabled.Value())
}

func TestSessionUpdateHandlerFunc_Pre_NoRelaysInDatacenter(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.RouteMatrix = env.GetRouteMatrix()

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPre(&state))

	assert.Equal(t, float64(1), state.Metrics.NoRelaysInDatacenter.Value())
}

func TestSessionUpdateHandlerFunc_Pre_StaleRouteMatrix(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name)
	env.AddRelay("losangeles.1", "10.0.0.2", datacenter.ID)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.StaleDuration = time.Second * 20
	state.RouteMatrix = env.GetRouteMatrix()

	// Make the route matrix creation time older by 30 seconds
	state.RouteMatrix.CreatedAt = uint64(time.Now().Add(-(time.Second * 30)).Unix())

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPre(&state))
	assert.True(t, state.StaleRouteMatrix)
	assert.Equal(t, float64(1), state.Metrics.StaleRouteMatrix.Value())
}

func TestSessionUpdateHandlerFunc_Pre_Success(t *testing.T) {
	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name)
	env.AddRelay("losangeles.1", "10.0.0.2", datacenter.ID)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.StaleDuration = time.Second * 20
	state.RouteMatrix = env.GetRouteMatrix()

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	assert.False(t, transport.SessionPre(&state))
}

type errLocator struct{}

func (locator *errLocator) LocateIP(ip net.IP) (routing.Location, error) {
	return routing.Location{}, fmt.Errorf("failed")
}

type successLoccator struct{}

func (locator *successLoccator) LocateIP(ip net.IP) (routing.Location, error) {
	return routing.Location{
		Continent:   "North America",
		Country:     "United States",
		CountryCode: "USA",
		Region:      "New York",
		City:        "Albany",
		Latitude:    float32(42.652580),
		Longitude:   float32(-73.756233),
		ISP:         "Spectrum",
		ASN:         10,
	}, nil
}

func TestSessionUpdateHandlerFunc_NewSession_LocationVeto(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		IpLocator: &errLocator{},
		Metrics:   metrics.SessionUpdateMetrics,
	}
	transport.SessionUpdateNewSession(&state)

	assert.True(t, state.Output.RouteState.LocationVeto)
	assert.Equal(t, float64(1), state.Metrics.ClientLocateFailure.Value())

	state.IpLocator = routing.NullIsland
	state.Metrics.ClientLocateFailure.ValueReset()

	transport.SessionUpdateNewSession(&state)

	assert.True(t, state.Output.RouteState.LocationVeto)
	assert.Equal(t, float64(1), state.Metrics.ClientLocateFailure.Value())
}

func TestSessionUpdateHandlerFunc_NewSession_Success(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		IpLocator: &successLoccator{},
		Metrics:   metrics.SessionUpdateMetrics,
		Output: transport.SessionData{
			SliceNumber: 0,
		},
	}
	transport.SessionUpdateNewSession(&state)

	assert.Equal(t, uint32(1), state.Output.SliceNumber)
	assert.Equal(t, state.Output, state.Input)

	assert.False(t, state.Output.RouteState.LocationVeto)
	assert.Equal(t, float64(0), state.Metrics.ClientLocateFailure.Value())
}

func TestSessionUpdateHandlerFunc_ExistingSession_BadSessionID(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	sessionData := transport.SessionData{
		SessionID: uint64(123456789),
	}

	var sessionDataBytesFixed [511]byte
	sessionDataBytes, err := transport.MarshalSessionData(&sessionData)
	assert.NoError(t, err)

	copy(sessionDataBytesFixed[:], sessionDataBytes)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			SessionID:   uint64(9876543120),
			SessionData: sessionDataBytesFixed,
		},
	}

	transport.SessionUpdateExistingSession(&state)

	assert.Equal(t, float64(1), state.Metrics.BadSessionID.Value())
}

func TestSessionUpdateHandlerFunc_ExistingSession_BadSliceNumber(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	sessionData := transport.SessionData{
		SessionID:   uint64(123456789),
		SliceNumber: 23,
	}

	var sessionDataBytesFixed [511]byte
	sessionDataBytes, err := transport.MarshalSessionData(&sessionData)
	assert.NoError(t, err)

	copy(sessionDataBytesFixed[:], sessionDataBytes)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			SessionID:   uint64(123456789),
			SliceNumber: 199,
			SessionData: sessionDataBytesFixed,
		},
	}

	transport.SessionUpdateExistingSession(&state)

	assert.Equal(t, float64(1), state.Metrics.BadSliceNumber.Value())
}

func TestSessionUpdateHandlerFunc_ExistingSession_Success(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	sessionData := transport.SessionData{
		SessionID:   uint64(123456789),
		SliceNumber: 1,
		Initial:     true,
	}

	var sessionDataBytesFixed [511]byte
	sessionDataBytes, err := transport.MarshalSessionData(&sessionData)
	assert.NoError(t, err)

	copy(sessionDataBytesFixed[:], sessionDataBytes)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			SessionID:   uint64(123456789),
			SliceNumber: 1,
			SessionData: sessionDataBytesFixed,
		},
	}

	transport.SessionUpdateExistingSession(&state)

	assert.False(t, state.Output.Initial)
	assert.Equal(t, uint32(2), state.Output.SliceNumber)
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_NoFallback(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: false,
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.False(t, transport.SessionHandleFallbackToDirect(&state))
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_BadRouteToken(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 0),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectBadRouteToken.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_NoNextRouteToContinue(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 1),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectNoNextRouteToContinue.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_PreviousUpdateStillPending(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 2),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectPreviousUpdateStillPending.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_BadContinueToken(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 3),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectBadContinueToken.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_RouteExpired(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 4),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectRouteExpired.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_RouteRequestTimedOut(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 5),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectRouteRequestTimedOut.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_ContinueRequestTimedOut(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 6),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectContinueRequestTimedOut.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_UpgradeResponseTimedOut(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 8),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectUpgradeResponseTimedOut.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_RouteUpdateTimedOut(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 9),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectRouteUpdateTimedOut.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_DirectPongTimedOut(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 10),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectDirectPongTimedOut.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_NextPongTimedOut(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 11),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectNextPongTimedOut.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_Unknown(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 12),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectUnknownReason.Value())
}

func TestSessionUpdateHandlerFunc_SessionUpdateNearRelayStats_NoRelaysInDatacenter(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
	}

	state.Datacenter = routing.Datacenter{
		ID:        crypto.HashID("datacenter.name"),
		Name:      "datacenter.name",
		AliasName: "datacenter.name",
	}

	state.RouteMatrix = &routing.RouteMatrix{
		RelayDatacenterIDs: []uint64{
			12345,
			123423,
			12351321,
		},
	}

	assert.False(t, transport.SessionUpdateNearRelayStats(&state))
	assert.Equal(t, float64(1), state.Metrics.NoRelaysInDatacenter.Value())
}

// todo add more SessionUpdateNearRelayStats tests here

func TestSessionUpdateHandlerFunc_SessionMakeRouteDecision_NextWithoutRouteRelays(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Input: transport.SessionData{
			RouteState: core.RouteState{
				Next: true,
			},
			RouteNumRelays: 0,
		},
	}

	transport.SessionMakeRouteDecision(&state)

	assert.False(t, state.Output.RouteState.Next)
	assert.True(t, state.Output.RouteState.Veto)
	assert.Equal(t, float64(1), state.Metrics.NextWithoutRouteRelays.Value())
}

// todo add more SessionMakeRouteDecision tests here

type testLocator struct{}

func (locator *testLocator) LocateIP(ip net.IP) (routing.Location, error) {
	return routing.Location{
		Latitude:  10,
		Longitude: 10,
	}, nil
}

func TestSessionUpdateHandlerFunc_BuyerNotFound_NoResponse(t *testing.T) {
	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	buyerID := binary.LittleEndian.Uint64(publicKey[:8])

	databaseWrapper := routing.CreateEmptyDatabaseBinWrapper()

	routeMatrix := &routing.RouteMatrix{}

	getDatabase := func() *routing.DatabaseBinWrapper {
		return databaseWrapper
	}

	getIPLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &testLocator{}
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return routeMatrix
	}

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", getDatabase)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)

	routerPrivateKeySlice, err := base64.StdEncoding.DecodeString("ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=")
	assert.NoError(t, err)

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], routerPrivateKeySlice)

	responseBuffer := &bytes.Buffer{}
	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersionMin,
		BuyerID:              buyerID,
		DatacenterID:         crypto.HashID("datacenter.name"),
		SessionID:            uint64(123456789),
		SliceNumber:          uint32(0),
		Next:                 false,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// Add the packet type byte and hash bytes to the request data so we can sign it properly
	requestDataHeader := append([]byte{transport.PacketTypeSessionUpdate}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)

	// Sign the packet
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once the packet is signed, we need to remove the header before passing to the session update handler
	requestData = requestData[1+crypto.PacketHashSize:]

	handler := transport.SessionUpdateHandlerFunc(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, getDatabase, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	// Buyer not found - no response
	assert.Equal(t, 0, len(responseBuffer.Bytes()))
}

/*

func TestSessionUpdateHandlerFunc_SigCheckFailed_NoResponse(t *testing.T) {
	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	buyerID := binary.LittleEndian.Uint64(publicKey[:8])

	databaseWrapper := routing.CreateEmptyDatabaseBinWrapper()

	routeMatrix := &routing.RouteMatrix{}

	getDatabase := func() *routing.DatabaseBinWrapper {
		return databaseWrapper
	}

	getIPLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &testLocator{}
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return routeMatrix
	}

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", getDatabase)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)

	routerPrivateKeySlice, err := base64.StdEncoding.DecodeString("ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=")
	assert.NoError(t, err)

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], routerPrivateKeySlice)

	responseBuffer := &bytes.Buffer{}
	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersionMin,
		BuyerID:              buyerID,
		DatacenterID:         crypto.HashID("datacenter.name"),
		SessionID:            uint64(123456789),
		SliceNumber:          uint32(0),
		Next:                 false,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// Add the packet type byte and hash bytes to the request data so we can sign it properly
	requestDataHeader := append([]byte{transport.PacketTypeSessionUpdate}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)

	// Sign the packet
	requestData = crypto.SignPacket(privateKey[1:], requestData)

	// Once the packet is signed, we need to remove the header before passing to the session update handler
	requestData = requestData[1+crypto.PacketHashSize:]

	handler := transport.SessionUpdateHandlerFunc(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, getDatabase, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	// Sig check failed - no response
	assert.Equal(t, 0, len(responseBuffer.Bytes()))
}

func TestSessionUpdateHandlerFunc_DirectResponse(t *testing.T) {
	env := test.NewTestEnvironment(t)

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	buyerID := binary.LittleEndian.Uint64(publicKey[:8])

	databaseWrapper := routing.CreateEmptyDatabaseBinWrapper()

	databaseWrapper.BuyerMap[buyerID] = routing.Buyer{
		ID:        buyerID,
		PublicKey: publicKey,
		Live:      true,
	}

	datacenterName := "datacenter.name.0"
	datacenterID := uint64(0)
	databaseWrapper.DatacenterMap[datacenterID] = routing.Datacenter{
		ID:        0,
		Name:      datacenterName,
		AliasName: datacenterName,
	}

	databaseWrapper.DatacenterMaps[buyerID] = make(map[uint64]routing.DatacenterMap, 0)
	databaseWrapper.DatacenterMaps[buyerID][datacenterID] = routing.DatacenterMap{
		BuyerID:      buyerID,
		DatacenterID: datacenterID,
		Alias:        datacenterName,
	}

	getDatabase := func() *routing.DatabaseBinWrapper {
		return databaseWrapper
	}

	getIPLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &testLocator{}
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", getDatabase)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)

	routerPrivateKeySlice, err := base64.StdEncoding.DecodeString("ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=")
	assert.NoError(t, err)

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], routerPrivateKeySlice)

	responseBuffer := &bytes.Buffer{}
	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersionMin,
		BuyerID:              buyerID,
		DatacenterID:         datacenterID,
		SessionID:            uint64(123456789),
		SliceNumber:          uint32(0),
		Next:                 false,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// Add the packet type byte and hash bytes to the request data so we can sign it properly
	requestDataHeader := append([]byte{transport.PacketTypeSessionUpdate}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)

	// Sign the packet
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once the packet is signed, we need to remove the header before passing to the session update handler
	requestData = requestData[1+crypto.PacketHashSize:]

	handler := transport.SessionUpdateHandlerFunc(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, getDatabase, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version // Make sure we unmarshal the response the same way we marshaled the request
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, int32(routing.RouteTypeDirect), responsePacket.RouteType)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.False(t, sessionData.RouteState.Next)
}

func TestSessionUpdateHandlerFunc_SessionMakeRouteDecision_NextResponse(t *testing.T) {
	env := test.NewTestEnvironment(t)

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	buyerID := binary.LittleEndian.Uint64(publicKey[:8])

	databaseWrapper := routing.CreateEmptyDatabaseBinWrapper()

	databaseWrapper.BuyerMap[buyerID] = routing.Buyer{
		ID:             buyerID,
		PublicKey:      publicKey,
		Live:           true,
		Debug:          true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	}

	datacenterName := "datacenter.name.0"
	datacenterID := uint64(0)
	databaseWrapper.DatacenterMap[datacenterID] = routing.Datacenter{
		ID:        0,
		Name:      datacenterName,
		AliasName: datacenterName,
	}

	databaseWrapper.DatacenterMaps[buyerID] = make(map[uint64]routing.DatacenterMap, 0)
	databaseWrapper.DatacenterMaps[buyerID][datacenterID] = routing.DatacenterMap{
		BuyerID:      buyerID,
		DatacenterID: datacenterID,
		Alias:        datacenterName,
	}

	getDatabase := func() *routing.DatabaseBinWrapper {
		return databaseWrapper
	}

	getIPLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &testLocator{}
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	for i, id := range env.GetRelayIds() {
		databaseWrapper.RelayMap[id] = routing.Relay{
			Addr:      env.GetRelayAddresses()[i],
			PublicKey: publicKey,
		}
	}

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", getDatabase)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)

	routerPrivateKeySlice, err := base64.StdEncoding.DecodeString("ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=")
	assert.NoError(t, err)

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], routerPrivateKeySlice)

	requestSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		ExpireTimestamp: uint64(time.Now().Unix()) + 100,
		SessionID:       uint64(123456789),
		SliceNumber:     uint32(3),
		Initial:         false,
	}

	var requestSessionDataBytesFixed [511]byte
	requestSessionDataBytes, err := transport.MarshalSessionData(&requestSessionData)
	assert.NoError(t, err)

	copy(requestSessionDataBytesFixed[:], requestSessionDataBytes)

	relayIDs := env.GetRelayIds()
	responseBuffer := &bytes.Buffer{}
	requestPacket := transport.SessionUpdatePacket{
		Version:                  transport.SDKVersionLatest,
		BuyerID:                  buyerID,
		DatacenterID:             datacenterID,
		SessionID:                uint64(123456789),
		SliceNumber:              uint32(3),
		RetryNumber:              0,
		SessionDataBytes:         int32(len(requestSessionDataBytes)),
		SessionData:              requestSessionDataBytesFixed,
		ClientAddress:            *core.ParseAddress("10.0.0.9"),
		ServerAddress:            *core.ParseAddress("10.0.0.10"),
		ClientRoutePublicKey:     publicKey,
		ServerRoutePublicKey:     publicKey,
		UserHash:                 uint64(100),
		PlatformType:             0,
		ConnectionType:           0,
		Next:                     false,
		Committed:                false,
		Reported:                 false,
		FallbackToDirect:         false,
		ClientBandwidthOverLimit: false,
		ServerBandwidthOverLimit: false,
		ClientPingTimedOut:       false,
		NumTags:                  0,
		Tags:                     [8]uint64{},
		Flags:                    0,
		UserFlags:                0,
		DirectRTT:                1000,
		DirectPacketLoss:         100,
		DirectJitter:             1000,
		NextRTT:                  11,
		NextJitter:               11,
		NextPacketLoss:           1,
		NumNearRelays:            2,
		NearRelayIDs:             relayIDs,
		NearRelayRTT:             []int32{10, 10},
		NearRelayJitter:          []int32{10, 10},
		NearRelayPacketLoss:      []int32{1, 1},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// Add the packet type byte and hash bytes to the request data so we can sign it properly
	requestDataHeader := append([]byte{transport.PacketTypeSessionUpdate}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)

	// Sign the packet
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once the packet is signed, we need to remove the header before passing to the session update handler
	requestData = requestData[1+crypto.PacketHashSize:]

	handler := transport.SessionUpdateHandlerFunc(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, getDatabase, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version // Make sure we unmarshal the response the same way we marshaled the request
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, int32(routing.RouteTypeNew), responsePacket.RouteType)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.True(t, sessionData.RouteState.Next)
	assert.True(t, sessionData.Initial)
}

func TestSessionUpdateHandlerFunc_SessionMakeRouteDecision_ContinueResponse(t *testing.T) {
	env := test.NewTestEnvironment(t)

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")

	env.SetCost("losangeles", "chicago", 10)

	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	buyerID := binary.LittleEndian.Uint64(publicKey[:8])

	databaseWrapper := routing.CreateEmptyDatabaseBinWrapper()

	databaseWrapper.BuyerMap[buyerID] = routing.Buyer{
		ID:             buyerID,
		PublicKey:      publicKey,
		Live:           true,
		Debug:          true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	}

	datacenterName := "datacenter.name.0"
	datacenterID := uint64(0)
	databaseWrapper.DatacenterMap[datacenterID] = routing.Datacenter{
		ID:        0,
		Name:      datacenterName,
		AliasName: datacenterName,
	}

	databaseWrapper.DatacenterMaps[buyerID] = make(map[uint64]routing.DatacenterMap, 0)
	databaseWrapper.DatacenterMaps[buyerID][datacenterID] = routing.DatacenterMap{
		BuyerID:      buyerID,
		DatacenterID: datacenterID,
		Alias:        datacenterName,
	}

	getDatabase := func() *routing.DatabaseBinWrapper {
		return databaseWrapper
	}

	getIPLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &testLocator{}
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	for i, id := range env.GetRelayIds() {
		databaseWrapper.RelayMap[id] = routing.Relay{
			Addr:      env.GetRelayAddresses()[i],
			PublicKey: publicKey,
		}
	}

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", getDatabase)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, nil, 0, false, &billing.NoOpBiller{}, &billing.NoOpBiller{}, true, false, log.NewNopLogger(), metrics.PostSessionMetrics)

	routerPrivateKeySlice, err := base64.StdEncoding.DecodeString("ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=")
	assert.NoError(t, err)

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], routerPrivateKeySlice)

	relayIDs := env.GetRelayIds()

	requestSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		ExpireTimestamp: uint64(time.Now().Unix()) + 100,
		SessionID:       uint64(123456789),
		SliceNumber:     uint32(5),
		Initial:         true,
		RouteNumRelays:  2,
		RouteRelayIDs:   [5]uint64{13303252451600817855, 11215140690934626978, 0, 0, 0},
		RouteState: core.RouteState{
			UserID:          100,
			Next:            true,
			NumNearRelays:   2,
			NearRelayRTT:    [32]int32{10, 10},
			NearRelayJitter: [32]int32{10, 10},
		},
	}

	var requestSessionDataBytesFixed [511]byte
	requestSessionDataBytes, err := transport.MarshalSessionData(&requestSessionData)
	assert.NoError(t, err)

	copy(requestSessionDataBytesFixed[:], requestSessionDataBytes)

	responseBuffer := &bytes.Buffer{}
	requestPacket := transport.SessionUpdatePacket{
		Version:                  transport.SDKVersionLatest,
		BuyerID:                  buyerID,
		DatacenterID:             datacenterID,
		SessionID:                uint64(123456789),
		SliceNumber:              uint32(5),
		RetryNumber:              0,
		SessionDataBytes:         int32(len(requestSessionDataBytes)),
		SessionData:              requestSessionDataBytesFixed,
		ClientAddress:            *core.ParseAddress("10.0.0.9"),
		ServerAddress:            *core.ParseAddress("10.0.0.10"),
		ClientRoutePublicKey:     publicKey,
		ServerRoutePublicKey:     publicKey,
		UserHash:                 uint64(100),
		PlatformType:             0,
		ConnectionType:           0,
		Next:                     true,
		Committed:                true,
		Reported:                 false,
		FallbackToDirect:         false,
		ClientBandwidthOverLimit: false,
		ServerBandwidthOverLimit: false,
		ClientPingTimedOut:       false,
		NumTags:                  0,
		Tags:                     [8]uint64{},
		Flags:                    0,
		UserFlags:                0,
		DirectRTT:                1000,
		DirectPacketLoss:         100,
		DirectJitter:             1000,
		NextRTT:                  11,
		NextJitter:               11,
		NextPacketLoss:           1,
		NumNearRelays:            2,
		NearRelayIDs:             relayIDs,
		NearRelayRTT:             []int32{10, 10},
		NearRelayJitter:          []int32{10, 10},
		NearRelayPacketLoss:      []int32{1, 1},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// Add the packet type byte and hash bytes to the request data so we can sign it properly
	requestDataHeader := append([]byte{transport.PacketTypeSessionUpdate}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)

	// Sign the packet
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once the packet is signed, we need to remove the header before passing to the session update handler
	requestData = requestData[1+crypto.PacketHashSize:]

	handler := transport.SessionUpdateHandlerFunc(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, getDatabase, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version // Make sure we unmarshal the response the same way we marshaled the request
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, int32(routing.RouteTypeContinue), responsePacket.RouteType)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.True(t, sessionData.RouteState.Next)
	assert.False(t, sessionData.Initial)
}
*/
