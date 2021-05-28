package transport_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"net"
	"testing"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
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

func TestServerInitHandlerFunc_Init_BuyerNotFound(t *testing.T) {
	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	buyerID := binary.LittleEndian.Uint64(publicKey[:8])

	databaseWrapper := routing.CreateEmptyDatabaseBinWrapper()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 10},
		BuyerID:        buyerID,
		DatacenterID:   crypto.HashID("datacenter.name"),
		DatacenterName: "datacenter.name",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

	getDatabase := func() *routing.DatabaseBinWrapper {
		return databaseWrapper
	}

	handler := transport.ServerInitHandlerFunc(getDatabase, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseUnknownBuyer), responsePacket.Response)

	assert.Equal(t, float64(1), metrics.ServerInitMetrics.BuyerNotFound.Value())
}

func TestServerInitHandlerFunc_Init_BuyerNotLive(t *testing.T) {
	publicKey, privateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	publicKey = publicKey[8:]
	privateKey = privateKey[8:]

	buyerID := binary.LittleEndian.Uint64(publicKey[:8])

	databaseWrapper := routing.CreateEmptyDatabaseBinWrapper()

	databaseWrapper.BuyerMap[buyerID] = routing.Buyer{
		ID:        buyerID,
		PublicKey: publicKey,
		Live:      false,
	}

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 10},
		BuyerID:        buyerID,
		DatacenterID:   crypto.HashID("datacenter.name"),
		DatacenterName: "datacenter.name",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

	getDatabase := func() *routing.DatabaseBinWrapper {
		return databaseWrapper
	}

	handler := transport.ServerInitHandlerFunc(getDatabase, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseBuyerNotActive), responsePacket.Response)

	assert.Equal(t, float64(1), metrics.ServerInitMetrics.BuyerNotActive.Value())
}

func TestServerInitHandlerFunc_Init_SDKToOld(t *testing.T) {
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

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{3, 0, 0},
		BuyerID:        buyerID,
		DatacenterID:   crypto.HashID("datacenter.name"),
		DatacenterName: "datacenter.name",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

	getDatabase := func() *routing.DatabaseBinWrapper {
		return databaseWrapper
	}

	handler := transport.ServerInitHandlerFunc(getDatabase, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseOldSDKVersion), responsePacket.Response)

	assert.Equal(t, float64(1), metrics.ServerInitMetrics.SDKTooOld.Value())
}

func TestServerInitHandlerFunc_Init_Success_DatacenterNotFound(t *testing.T) {
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

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestPacket := transport.ServerInitRequestPacket{
		Version:        transport.SDKVersion{4, 0, 10},
		BuyerID:        buyerID,
		DatacenterID:   crypto.HashID("datacenter.name"),
		DatacenterName: "datacenter.name",
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	// We need to add the packet header (packet type + 8 hash bytes) in order to get the correct signature
	requestDataHeader := append([]byte{transport.PacketTypeServerInitRequest}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)
	requestData = crypto.SignPacket(privateKey, requestData)

	// Once we have the signature, we need to take off the header before passing to the handler
	requestData = requestData[1+crypto.PacketHashSize:]

	getDatabase := func() *routing.DatabaseBinWrapper {
		return databaseWrapper
	}

	handler := transport.ServerInitHandlerFunc(getDatabase, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, requestPacket.RequestID, responsePacket.RequestID)
	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)

	assert.Equal(t, float64(1), metrics.ServerInitMetrics.DatacenterNotFound.Value())
}
