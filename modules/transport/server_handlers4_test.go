package transport_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	md "github.com/networknext/backend/modules/match_data"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/test"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func TestGetRouteAddressesAndPublicKeys(t *testing.T) {
	t.Parallel()

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

func TestGetRouteAddressesAndPublicKeys_InternalAddressClientRoutable_MissingInternalAddr(t *testing.T) {
	t.Parallel()

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
	relayMap[crypto.HashID(relayAddr1.String())] = routing.Relay{ID: crypto.HashID(relayAddr1.String()), Addr: *relayAddr1, PublicKey: relayPublicKey1, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}
	relayMap[crypto.HashID(relayAddr2.String())] = routing.Relay{ID: crypto.HashID(relayAddr2.String()), Addr: *relayAddr2, PublicKey: relayPublicKey2, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}
	relayMap[crypto.HashID(relayAddr3.String())] = routing.Relay{ID: crypto.HashID(relayAddr3.String()), Addr: *relayAddr3, PublicKey: relayPublicKey3, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}

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

func TestGetRouteAddressesAndPublicKeys_InternalAddressClientRoutable_Success(t *testing.T) {
	t.Parallel()

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

	relayInternalAddr1, err := net.ResolveUDPAddr("udp", "128.0.0.1:10000")
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
	relayMap[crypto.HashID(relayAddr1.String())] = routing.Relay{ID: crypto.HashID(relayAddr1.String()), Addr: *relayAddr1, InternalAddr: *relayInternalAddr1, PublicKey: relayPublicKey1, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}
	relayMap[crypto.HashID(relayAddr2.String())] = routing.Relay{ID: crypto.HashID(relayAddr2.String()), Addr: *relayAddr2, PublicKey: relayPublicKey2, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}
	relayMap[crypto.HashID(relayAddr3.String())] = routing.Relay{ID: crypto.HashID(relayAddr3.String()), Addr: *relayAddr3, PublicKey: relayPublicKey3, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}

	database := routing.DatabaseBinWrapper{RelayMap: relayMap, SellerMap: sellerMap, DatacenterMap: datacenterMap}

	allRelayIDs := []uint64{crypto.HashID(relayAddr1.String()), crypto.HashID(relayAddr2.String()), crypto.HashID(relayAddr3.String())}
	routeRelays := []int32{0, 1, 2}

	routeAddresses, routePublicKeys := transport.GetRouteAddressesAndPublicKeys(clientAddr, clientPublicKey, serverAddr, serverPublicKey, 5, routeRelays, allRelayIDs, &database)

	expectedRouteAddresses := []*net.UDPAddr{clientAddr, relayInternalAddr1, relayAddr2, relayAddr3, serverAddr}
	expectedRoutePublicKeys := [][]byte{clientPublicKey, relayPublicKey1, relayPublicKey2, relayPublicKey3, serverPublicKey}

	for i := range routeAddresses {
		assert.Equal(t, expectedRouteAddresses[i].String(), routeAddresses[i].String())
	}

	for i := range routePublicKeys {
		assert.Equal(t, expectedRoutePublicKeys[i], routePublicKeys[i])
	}
}

func TestGetRouteAddressesAndPublicKeys_UseRelayPrivateAddress(t *testing.T) {
	t.Parallel()

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

	relayInternalAddr1, err := net.ResolveUDPAddr("udp", "128.0.0.1:10000")
	assert.NoError(t, err)
	relayInternalAddr2, err := net.ResolveUDPAddr("udp", "128.0.0.1:10001")
	assert.NoError(t, err)
	relayInternalAddr3, err := net.ResolveUDPAddr("udp", "128.0.0.1:10002")
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
	relayMap[crypto.HashID(relayAddr1.String())] = routing.Relay{ID: crypto.HashID(relayAddr1.String()), Addr: *relayAddr1, InternalAddr: *relayInternalAddr1, PublicKey: relayPublicKey1, Seller: seller, Datacenter: datacenter}
	relayMap[crypto.HashID(relayAddr2.String())] = routing.Relay{ID: crypto.HashID(relayAddr2.String()), Addr: *relayAddr2, InternalAddr: *relayInternalAddr2, PublicKey: relayPublicKey2, Seller: seller, Datacenter: datacenter}
	relayMap[crypto.HashID(relayAddr3.String())] = routing.Relay{ID: crypto.HashID(relayAddr3.String()), Addr: *relayAddr3, InternalAddr: *relayInternalAddr3, PublicKey: relayPublicKey3, Seller: seller, Datacenter: datacenter}

	database := routing.DatabaseBinWrapper{RelayMap: relayMap, SellerMap: sellerMap, DatacenterMap: datacenterMap}

	allRelayIDs := []uint64{crypto.HashID(relayAddr1.String()), crypto.HashID(relayAddr2.String()), crypto.HashID(relayAddr3.String())}
	routeRelays := []int32{0, 1, 2}

	routeAddresses, routePublicKeys := transport.GetRouteAddressesAndPublicKeys(clientAddr, clientPublicKey, serverAddr, serverPublicKey, 5, routeRelays, allRelayIDs, &database)

	expectedRouteAddresses := []*net.UDPAddr{clientAddr, relayAddr1, relayInternalAddr2, relayInternalAddr3, serverAddr}
	expectedRoutePublicKeys := [][]byte{clientPublicKey, relayPublicKey1, relayPublicKey2, relayPublicKey3, serverPublicKey}

	for i := range routeAddresses {
		assert.Equal(t, expectedRouteAddresses[i].String(), routeAddresses[i].String())
	}

	for i := range routePublicKeys {
		assert.Equal(t, expectedRoutePublicKeys[i], routePublicKeys[i])
	}
}

func TestGetRouteAddressesAndPublicKeys_UseRelayPrivateAddress_InternalAddressClientRoutable_MissingInternalAddr(t *testing.T) {
	t.Parallel()

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

	relayInternalAddr2, err := net.ResolveUDPAddr("udp", "128.0.0.1:10001")
	assert.NoError(t, err)
	relayInternalAddr3, err := net.ResolveUDPAddr("udp", "128.0.0.1:10002")
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
	relayMap[crypto.HashID(relayAddr1.String())] = routing.Relay{ID: crypto.HashID(relayAddr1.String()), Addr: *relayAddr1, PublicKey: relayPublicKey1, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}
	relayMap[crypto.HashID(relayAddr2.String())] = routing.Relay{ID: crypto.HashID(relayAddr2.String()), Addr: *relayAddr2, InternalAddr: *relayInternalAddr2, PublicKey: relayPublicKey2, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}
	relayMap[crypto.HashID(relayAddr3.String())] = routing.Relay{ID: crypto.HashID(relayAddr3.String()), Addr: *relayAddr3, InternalAddr: *relayInternalAddr3, PublicKey: relayPublicKey3, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}

	database := routing.DatabaseBinWrapper{RelayMap: relayMap, SellerMap: sellerMap, DatacenterMap: datacenterMap}

	allRelayIDs := []uint64{crypto.HashID(relayAddr1.String()), crypto.HashID(relayAddr2.String()), crypto.HashID(relayAddr3.String())}
	routeRelays := []int32{0, 1, 2}

	routeAddresses, routePublicKeys := transport.GetRouteAddressesAndPublicKeys(clientAddr, clientPublicKey, serverAddr, serverPublicKey, 5, routeRelays, allRelayIDs, &database)

	expectedRouteAddresses := []*net.UDPAddr{clientAddr, relayAddr1, relayInternalAddr2, relayInternalAddr3, serverAddr}
	expectedRoutePublicKeys := [][]byte{clientPublicKey, relayPublicKey1, relayPublicKey2, relayPublicKey3, serverPublicKey}

	for i := range routeAddresses {
		assert.Equal(t, expectedRouteAddresses[i].String(), routeAddresses[i].String())
	}

	for i := range routePublicKeys {
		assert.Equal(t, expectedRoutePublicKeys[i], routePublicKeys[i])
	}
}

func TestGetRouteAddressesAndPublicKeys_UseRelayPrivateAddress_InternalAddressClientRoutable_Success(t *testing.T) {
	t.Parallel()

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

	relayInternalAddr1, err := net.ResolveUDPAddr("udp", "128.0.0.1:10000")
	assert.NoError(t, err)
	relayInternalAddr2, err := net.ResolveUDPAddr("udp", "128.0.0.1:10001")
	assert.NoError(t, err)
	relayInternalAddr3, err := net.ResolveUDPAddr("udp", "128.0.0.1:10002")
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
	relayMap[crypto.HashID(relayAddr1.String())] = routing.Relay{ID: crypto.HashID(relayAddr1.String()), Addr: *relayAddr1, InternalAddr: *relayInternalAddr1, PublicKey: relayPublicKey1, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}
	relayMap[crypto.HashID(relayAddr2.String())] = routing.Relay{ID: crypto.HashID(relayAddr2.String()), Addr: *relayAddr2, InternalAddr: *relayInternalAddr2, PublicKey: relayPublicKey2, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}
	relayMap[crypto.HashID(relayAddr3.String())] = routing.Relay{ID: crypto.HashID(relayAddr3.String()), Addr: *relayAddr3, InternalAddr: *relayInternalAddr3, PublicKey: relayPublicKey3, Seller: seller, Datacenter: datacenter, InternalAddressClientRoutable: true}

	database := routing.DatabaseBinWrapper{RelayMap: relayMap, SellerMap: sellerMap, DatacenterMap: datacenterMap}

	allRelayIDs := []uint64{crypto.HashID(relayAddr1.String()), crypto.HashID(relayAddr2.String()), crypto.HashID(relayAddr3.String())}
	routeRelays := []int32{0, 1, 2}

	routeAddresses, routePublicKeys := transport.GetRouteAddressesAndPublicKeys(clientAddr, clientPublicKey, serverAddr, serverPublicKey, 5, routeRelays, allRelayIDs, &database)

	expectedRouteAddresses := []*net.UDPAddr{clientAddr, relayInternalAddr1, relayInternalAddr2, relayInternalAddr3, serverAddr}
	expectedRoutePublicKeys := [][]byte{clientPublicKey, relayPublicKey1, relayPublicKey2, relayPublicKey3, serverPublicKey}

	for i := range routeAddresses {
		assert.Equal(t, expectedRouteAddresses[i].String(), routeAddresses[i].String())
	}

	for i := range routePublicKeys {
		assert.Equal(t, expectedRoutePublicKeys[i], routePublicKeys[i])
	}
}

// Server init handler tests

func TestServerInitHandlerFunc_BuyerNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false, false)

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

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, unknownBuyerID, datacenterID, datacenterName, unknownPrivateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, serverTracker, metrics.ServerInitMetrics)
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
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", false, false)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, privateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, serverTracker, metrics.ServerInitMetrics)
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
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, privateKey[2:])

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, serverTracker, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseSignatureCheckFailed), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.SignatureCheckFailed.Value())
}

func TestServerInitHandlerFunc_SDKTooOld(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersion{3, 0, 0}, buyerID, datacenterID, datacenterName, privateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, serverTracker, metrics.ServerInitMetrics)
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
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, privateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, serverTracker, metrics.ServerInitMetrics)
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
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, buyerID, datacenter.ID, datacenter.Name, privateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, serverTracker, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)
	assert.Equal(t, float64(0), metrics.ServerInitMetrics.DatacenterNotFound.Value())
}

func TestServerInitHandlerFunc_ServerTracker_DatacenterNotFound_WithoutName(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := ""
	datacenterID := crypto.HashID(datacenterName)

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, privateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, serverTracker, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.DatacenterNotFound.Value())

	assert.NotEmpty(t, serverTracker.Tracker)

	buyerHexID := fmt.Sprintf("%016x", buyerID)
	for _, serverInfo := range serverTracker.Tracker[buyerHexID] {
		assert.Equal(t, serverInfo.DatacenterID, fmt.Sprintf("%016x", datacenterID))
		assert.Equal(t, serverInfo.DatacenterName, "unknown_init")
	}
}

func TestServerInitHandlerFunc_ServerTracker_DatacenterNotFound_WithName(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	serverTracker := storage.NewServerTracker()

	requestData := env.GenerateServerInitRequestPacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, privateKey)

	handler := transport.ServerInitHandlerFunc(env.GetDatabaseWrapper, serverTracker, metrics.ServerInitMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.ServerInitResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.InitResponseOK), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.ServerInitMetrics.DatacenterNotFound.Value())

	assert.NotEmpty(t, serverTracker.Tracker)

	buyerHexID := fmt.Sprintf("%016x", buyerID)
	for _, serverInfo := range serverTracker.Tracker[buyerHexID] {
		assert.Equal(t, serverInfo.DatacenterID, fmt.Sprintf("%016x", datacenterID))
		assert.Equal(t, serverInfo.DatacenterName, datacenterName)
	}
}

// Server update handler tests

func TestServerUpdateHandlerFunc_BuyerNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false, false)

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

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	serverTracker := storage.NewServerTracker()

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, serverTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.BuyerNotFound.Value())
}

func TestServerUpdateHandlerFunc_BuyerNotLive(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", false, false)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, 10, "10.0.0.1", privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	serverTracker := storage.NewServerTracker()

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, serverTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.BuyerNotLive.Value())
}

func TestServerUpdateHandlerFunc_SigCheckFail(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, 10, "10.0.0.1", privateKey[2:])

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	serverTracker := storage.NewServerTracker()

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, serverTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.SignatureCheckFailed.Value())
}

func TestServerUpdateHandlerFunc_SDKToOld(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersion{3, 0, 0}, buyerID, datacenterID, datacenterName, 10, "10.0.0.1", privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	serverTracker := storage.NewServerTracker()

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, serverTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.SDKTooOld.Value())
}

func TestServerUpdateHandlerFunc_DatacenterNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, 10, "10.0.0.1", privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	serverTracker := storage.NewServerTracker()

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, serverTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.DatacenterNotFound.Value())
}

func TestServerUpdateHandlerFunc_Success(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersionLatest, buyerID, datacenter.ID, datacenter.Name, 10, "10.0.0.1", privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	serverTracker := storage.NewServerTracker()

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, serverTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(0), metrics.ServerUpdateMetrics.DatacenterNotFound.Value())
}

func TestServerUpdateHandlerFunc_ServerTracker_DatacenterNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	datacenterName := "datacenter.name"
	datacenterID := crypto.HashID(datacenterName)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersionLatest, buyerID, datacenterID, datacenterName, 10, "10.0.0.1", privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	serverTracker := storage.NewServerTracker()

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, serverTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(1), metrics.ServerUpdateMetrics.DatacenterNotFound.Value())

	assert.NotEmpty(t, serverTracker.Tracker)

	buyerHexID := fmt.Sprintf("%016x", buyerID)
	for _, serverInfo := range serverTracker.Tracker[buyerHexID] {
		assert.Equal(t, serverInfo.DatacenterID, fmt.Sprintf("%016x", datacenterID))
		assert.Equal(t, serverInfo.DatacenterName, "unknown_update")
	}
}

func TestServerUpdateHandlerFunc_ServerTracker_DatacenterFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateServerUpdatePacket(transport.SDKVersionLatest, buyerID, datacenter.ID, datacenter.Name, 10, "10.0.0.1", privateKey)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	serverTracker := storage.NewServerTracker()

	handler := transport.ServerUpdateHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, serverTracker, metrics.ServerUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	assert.Equal(t, float64(0), metrics.ServerUpdateMetrics.DatacenterNotFound.Value())

	assert.NotEmpty(t, serverTracker.Tracker)

	buyerHexID := fmt.Sprintf("%016x", buyerID)
	for _, serverInfo := range serverTracker.Tracker[buyerHexID] {
		assert.Equal(t, serverInfo.DatacenterID, fmt.Sprintf("%016x", datacenter.ID))
		assert.Equal(t, serverInfo.DatacenterName, datacenter.Name)
	}
}

// Session update handler

func getErrorLocator() *routing.MaxmindDB {
	return &routing.MaxmindDB{}
}

func getSuccessLocator(t *testing.T) *routing.MaxmindDB {
	// Set IsStaging to true to use sessionID instead of IP Address
	// when we are not testing IP2Location
	mmdb := &routing.MaxmindDB{
		CityFile:  "../../testdata/GeoIP2-City-Test.mmdb",
		IspFile:   "../../testdata/GeoIP2-ISP-Test.mmdb",
		IsStaging: true,
	}

	err := mmdb.OpenCity(context.Background(), mmdb.CityFile)
	assert.NoError(t, err)
	err = mmdb.OpenISP(context.Background(), mmdb.IspFile)
	assert.NoError(t, err)

	return mmdb
}

func TestSessionUpdateHandlerFunc_Pre_BuyerNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false, false)

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
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = unknownBuyerID

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.BuyerNotFound)
	assert.Equal(t, float64(1), state.Metrics.BuyerNotFound.Value())
}

func TestSessionUpdateHandlerFunc_Pre_BuyerNotLive(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, _ := env.AddBuyer("local", false, false)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.BuyerNotLive)
	assert.Equal(t, float64(1), state.Metrics.BuyerNotLive.Value())
}

func TestSessionUpdateHandlerFunc_Pre_SigCheckFail(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey[2:])
	state.PacketData = requestData

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.SignatureCheckFailed)
	assert.Equal(t, float64(1), state.Metrics.SignatureCheckFailed.Value())
}

func TestSessionUpdateHandlerFunc_Pre_ClientTimedOut(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.Packet.ClientPingTimedOut = true

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.Packet.ClientPingTimedOut)
	assert.Equal(t, float64(1), state.Metrics.ClientPingTimedOut.Value())
}

func TestSessionUpdateHandlerFunc_Pre_LocationVeto(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.Packet.ClientPingTimedOut = false

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	state.IpLocator = getErrorLocator()

	transport.SessionPre(&state)

	assert.True(t, state.Output.RouteState.LocationVeto)
	assert.Equal(t, float64(1), state.Metrics.ClientLocateFailure.Value())
}

func TestSessionUpdateHandlerFunc_Pre_DatacenterNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = crypto.HashID("unknown.datacenter.name")
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.UnknownDatacenter)
	assert.Equal(t, float64(1), state.Metrics.DatacenterNotFound.Value())
}

func TestSessionUpdateHandlerFunc_Pre_DatacenterNotFound_AnalysisOnly(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, true)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = crypto.HashID("unknown.datacenter.name")
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.RouteMatrix = env.GetRouteMatrix()

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	assert.False(t, transport.SessionPre(&state))

	assert.True(t, state.UnknownDatacenter)
	assert.Equal(t, float64(1), state.Metrics.DatacenterNotFound.Value())
	assert.Equal(t, routing.UnknownDatacenter, state.Datacenter)
}

func TestSessionUpdateHandlerFunc_Pre_DatacenterNotEnabled(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	assert.True(t, transport.SessionPre(&state))

	assert.True(t, state.DatacenterNotEnabled)
	assert.Equal(t, float64(1), state.Metrics.DatacenterNotEnabled.Value())
}

func TestSessionUpdateHandlerFunc_Pre_DatacenterNotEnabled_AnalysisOnly(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, true)
	datacenter := env.AddDatacenter("datacenter.name")

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.RouteMatrix = env.GetRouteMatrix()

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	assert.False(t, transport.SessionPre(&state))

	assert.False(t, state.DatacenterNotEnabled)
	assert.Equal(t, float64(0), state.Metrics.DatacenterNotEnabled.Value())
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerFunc_Pre_NoRelaysInDatacenter(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, true)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.RouteMatrix = env.GetRouteMatrix()

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	assert.True(t, transport.SessionPre(&state))

	assert.Equal(t, float64(1), state.Metrics.NoRelaysInDatacenter.Value())
}

func TestSessionUpdateHandlerFunc_Pre_NoRelaysInDatacenter_AnalysisOnly(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, true)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, true)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.RouteMatrix = env.GetRouteMatrix()

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	assert.False(t, transport.SessionPre(&state))

	assert.Equal(t, float64(0), state.Metrics.NoRelaysInDatacenter.Value())
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerFunc_Pre_StaleRouteMatrix(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, true)
	env.AddRelay("losangeles.1", "10.0.0.2", datacenter.ID)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.StaleDuration = time.Second * 20
	state.RouteMatrix = env.GetRouteMatrix()

	// Make the route matrix creation time older by 30 seconds
	state.RouteMatrix.CreatedAt = uint64(time.Now().Add(-(time.Second * 30)).Unix())

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData

	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	assert.True(t, transport.SessionPre(&state))
	assert.True(t, state.StaleRouteMatrix)
	assert.Equal(t, float64(1), state.Metrics.StaleRouteMatrix.Value())
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerFunc_Pre_Success(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, true)
	env.AddRelay("losangeles.1", "10.0.0.2", datacenter.ID)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.StaleDuration = time.Second * 20
	state.RouteMatrix = env.GetRouteMatrix()

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	assert.False(t, transport.SessionPre(&state))
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerFunc_Pre_Success_AnalysisOnly(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, true)
	datacenter := env.AddDatacenter("datacenter.name")
	env.AddDCMap(buyerID, datacenter.ID, datacenter.Name, true)
	env.AddRelay("losangeles.1", "10.0.0.2", datacenter.ID)

	state := transport.SessionHandlerState{}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)

	state.Metrics = metrics.SessionUpdateMetrics
	state.Database = env.GetDatabaseWrapper()
	state.Packet.DatacenterID = datacenter.ID
	state.PostSessionHandler = transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)
	state.Packet.BuyerID = buyerID
	state.StaleDuration = time.Second * 20
	state.RouteMatrix = env.GetRouteMatrix()

	requestData := env.GenerateEmptySessionUpdatePacket(privateKey)
	state.PacketData = requestData
	state.IpLocator = getSuccessLocator(t)
	defer state.IpLocator.CloseCity()
	defer state.IpLocator.CloseISP()

	assert.False(t, transport.SessionPre(&state))
	assert.False(t, state.UnknownDatacenter)
	assert.False(t, state.DatacenterNotEnabled)
	assert.Equal(t, float64(0), state.Metrics.DatacenterNotEnabled.Value())
	assert.Equal(t, float64(0), state.Metrics.NoRelaysInDatacenter.Value())
	assert.Equal(t, datacenter, state.Datacenter)
}

func TestSessionUpdateHandlerFunc_NewSession_Success(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Output: transport.SessionData{
			SliceNumber: 0,
		},
	}
	transport.SessionUpdateNewSession(&state)

	assert.Equal(t, uint32(1), state.Output.SliceNumber)
	assert.Equal(t, state.Output, state.Input)
}

func TestSessionUpdateHandlerFunc_ExistingSession_BadSessionID(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_ClientTimedOut(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Packet: transport.SessionUpdatePacket{
			FallbackToDirect: true,
			Flags:            (1 << 7),
		},
		Output: transport.SessionData{
			FellBackToDirect: false,
		},
	}

	assert.True(t, transport.SessionHandleFallbackToDirect(&state))
	assert.Equal(t, float64(1), state.Metrics.FallbackToDirectClientTimedOut.Value())
}

func TestSessionUpdateHandlerFunc_SessionHandleFallbackToDirect_UpgradeResponseTimedOut(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

func TestSessionUpdateHandlerFunc_SessionUpdateNearRelayStats_AnalysisOnly(t *testing.T) {
	t.Parallel()

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

	state.Buyer = routing.Buyer{
		RouteShader: core.RouteShader{
			AnalysisOnly: true,
		},
	}

	assert.False(t, transport.SessionUpdateNearRelayStats(&state))
	assert.Equal(t, float64(0), state.Metrics.NoRelaysInDatacenter.Value())
	assert.Zero(t, len(state.Packet.NearRelayIDs))
}

func TestSessionUpdateHandlerFunc_SessionUpdateNearRelayStats_NoRelaysInDatacenter(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics:                       metrics.SessionUpdateMetrics,
		DatacenterAccelerationEnabled: true,
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

func TestSessionUpdateHandlerFunc_SessionUpdateNearRelayStats_HoldNearRelays(t *testing.T) {
	t.Parallel()

	t.Run("Large Customer Transition false -> true before slice 4", func(t *testing.T) {
		updatePacket := transport.SessionUpdatePacket{
			SliceNumber: uint32(2),
		}

		metricsHandler := metrics.LocalHandler{}
		metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
		assert.NoError(t, err)

		buyer := routing.Buyer{
			InternalConfig: core.InternalConfig{
				LargeCustomer: false,
			},
		}

		dc := routing.Datacenter{
			ID:        crypto.HashID("datacenter.name"),
			Name:      "datacenter.name",
			AliasName: "datacenter.name",
		}

		routeMatrix := &routing.RouteMatrix{
			RelayIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
			RelayDatacenterIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
		}

		state := transport.SessionHandlerState{
			Packet:                        updatePacket,
			Metrics:                       metrics.SessionUpdateMetrics,
			RouteMatrix:                   routeMatrix,
			Datacenter:                    dc,
			Buyer:                         buyer,
			DatacenterAccelerationEnabled: true,
		}

		assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStats(&state))
		assert.False(t, state.Output.HoldNearRelays)

		state.Packet.SliceNumber++
		state.Buyer.InternalConfig.LargeCustomer = true

		assert.True(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStats(&state))
		assert.False(t, state.Output.HoldNearRelays)
	})

	t.Run("Large Customer Transition true -> false before slice 4", func(t *testing.T) {
		updatePacket := transport.SessionUpdatePacket{
			SliceNumber: uint32(2),
		}

		metricsHandler := metrics.LocalHandler{}
		metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
		assert.NoError(t, err)

		buyer := routing.Buyer{
			InternalConfig: core.InternalConfig{
				LargeCustomer: true,
			},
		}

		dc := routing.Datacenter{
			ID:        crypto.HashID("datacenter.name"),
			Name:      "datacenter.name",
			AliasName: "datacenter.name",
		}

		routeMatrix := &routing.RouteMatrix{
			RelayIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
			RelayDatacenterIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
		}

		state := transport.SessionHandlerState{
			Packet:                        updatePacket,
			Metrics:                       metrics.SessionUpdateMetrics,
			RouteMatrix:                   routeMatrix,
			Datacenter:                    dc,
			Buyer:                         buyer,
			DatacenterAccelerationEnabled: true,
		}

		assert.True(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStats(&state))
		assert.False(t, state.Output.HoldNearRelays)

		state.Packet.SliceNumber++
		state.Buyer.InternalConfig.LargeCustomer = false

		assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStats(&state))
		assert.False(t, state.Output.HoldNearRelays)
	})

	t.Run("Large Customer Transition false -> true on or after slice 4", func(t *testing.T) {
		rand.Seed(time.Now().Unix())

		updatePacket := transport.SessionUpdatePacket{
			SliceNumber: uint32(4),
			NearRelayIDs: []uint64{
				rand.Uint64(),
				rand.Uint64(),
				rand.Uint64(),
			},
			NearRelayRTT: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayJitter: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayPacketLoss: []int32{
				rand.Int31n(100),
				rand.Int31n(100),
				rand.Int31n(100),
			},
		}

		metricsHandler := metrics.LocalHandler{}
		metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
		assert.NoError(t, err)

		buyer := routing.Buyer{
			InternalConfig: core.InternalConfig{
				LargeCustomer: false,
			},
		}

		dc := routing.Datacenter{
			ID:        crypto.HashID("datacenter.name"),
			Name:      "datacenter.name",
			AliasName: "datacenter.name",
		}

		routeMatrix := &routing.RouteMatrix{
			RelayIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
			RelayDatacenterIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
		}

		state := transport.SessionHandlerState{
			Packet:                        updatePacket,
			Metrics:                       metrics.SessionUpdateMetrics,
			RouteMatrix:                   routeMatrix,
			Datacenter:                    dc,
			Buyer:                         buyer,
			DatacenterAccelerationEnabled: true,
		}

		assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStats(&state))
		assert.False(t, state.Output.HoldNearRelays)

		state.Packet.SliceNumber++
		state.Buyer.InternalConfig.LargeCustomer = true
		state.Input = state.Output

		assert.True(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStats(&state))
		assert.True(t, state.Output.HoldNearRelays)

		updatePacket = transport.SessionUpdatePacket{
			SliceNumber: state.Packet.SliceNumber + 1,
			NearRelayIDs: []uint64{
				rand.Uint64(),
				rand.Uint64(),
				rand.Uint64(),
			},
			NearRelayRTT: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayJitter: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayPacketLoss: []int32{
				rand.Int31n(100),
				rand.Int31n(100),
				rand.Int31n(100),
			},
		}

		state.Packet = updatePacket
		state.Input = state.Output

		assert.True(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStats(&state))
		assert.True(t, state.Output.HoldNearRelays)
		assert.Equal(t, state.Packet.NearRelayIDs, updatePacket.NearRelayIDs)
		for i := 0; i < len(state.Packet.NearRelayIDs); i++ {
			assert.Equal(t, state.Packet.NearRelayRTT[i], state.Input.HoldNearRelayRTT[i])
			assert.Equal(t, state.Packet.NearRelayRTT[i], updatePacket.NearRelayRTT[i])
		}
	})

	t.Run("Large Customer Transition true -> false on or after slice 4", func(t *testing.T) {
		rand.Seed(time.Now().Unix())

		updatePacket := transport.SessionUpdatePacket{
			SliceNumber: uint32(4),
			NearRelayIDs: []uint64{
				rand.Uint64(),
				rand.Uint64(),
				rand.Uint64(),
			},
			NearRelayRTT: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayJitter: []int32{
				rand.Int31n(255),
				rand.Int31n(255),
				rand.Int31n(255),
			},
			NearRelayPacketLoss: []int32{
				rand.Int31n(100),
				rand.Int31n(100),
				rand.Int31n(100),
			},
		}

		metricsHandler := metrics.LocalHandler{}
		metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
		assert.NoError(t, err)

		buyer := routing.Buyer{
			InternalConfig: core.InternalConfig{
				LargeCustomer: true,
			},
		}

		dc := routing.Datacenter{
			ID:        crypto.HashID("datacenter.name"),
			Name:      "datacenter.name",
			AliasName: "datacenter.name",
		}

		routeMatrix := &routing.RouteMatrix{
			RelayIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
			RelayDatacenterIDs: []uint64{
				crypto.HashID("datacenter.name"),
				uint64(1234),
				uint64(12345),
				uint64(123456),
			},
		}

		state := transport.SessionHandlerState{
			Packet:                        updatePacket,
			Metrics:                       metrics.SessionUpdateMetrics,
			RouteMatrix:                   routeMatrix,
			Datacenter:                    dc,
			Buyer:                         buyer,
			DatacenterAccelerationEnabled: true,
		}

		assert.True(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStats(&state))
		assert.True(t, state.Output.HoldNearRelays)

		state.Packet.SliceNumber++
		state.Buyer.InternalConfig.LargeCustomer = false
		state.Input = state.Output

		assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStats(&state))
		assert.True(t, state.Output.HoldNearRelays)
		assert.Equal(t, state.Packet.NearRelayIDs, updatePacket.NearRelayIDs)
		for i := 0; i < len(state.Packet.NearRelayIDs); i++ {
			assert.Equal(t, state.Packet.NearRelayRTT[i], state.Input.HoldNearRelayRTT[i])
			assert.Equal(t, state.Packet.NearRelayRTT[i], updatePacket.NearRelayRTT[i])
		}

		state.Packet.SliceNumber++
		state.Input = state.Output

		assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
		assert.True(t, transport.SessionUpdateNearRelayStats(&state))
		assert.True(t, state.Output.HoldNearRelays)
		assert.Equal(t, state.Packet.NearRelayIDs, updatePacket.NearRelayIDs)
		for i := 0; i < len(state.Packet.NearRelayIDs); i++ {
			assert.Equal(t, state.Packet.NearRelayRTT[i], state.Input.HoldNearRelayRTT[i])
			assert.Equal(t, state.Packet.NearRelayRTT[i], updatePacket.NearRelayRTT[i])
		}
	})
}

func TestSessionUpdateHandlerFunc_SessionUpdateNearRelayStats_RelayNoLongerExists(t *testing.T) {
	t.Parallel()

	updatePacket := transport.SessionUpdatePacket{
		SliceNumber:  uint32(2),
		NearRelayIDs: []uint64{1234, 12345, 123456},
	}

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	buyer := routing.Buyer{
		InternalConfig: core.InternalConfig{
			LargeCustomer: false,
		},
	}

	dc := routing.Datacenter{
		ID:        crypto.HashID("datacenter.name"),
		Name:      "datacenter.name",
		AliasName: "datacenter.name",
	}

	relayIDsToIndices := make(map[uint64]int32)
	relayIDsToIndices[1234] = 0
	relayIDsToIndices[12345] = 1

	routeMatrix := &routing.RouteMatrix{
		RelayIDsToIndices: relayIDsToIndices,
		RelayIDs: []uint64{
			crypto.HashID("datacenter.name"),
			uint64(1234),
			uint64(12345),
			uint64(123456),
		},
		RelayDatacenterIDs: []uint64{
			crypto.HashID("datacenter.name"),
			uint64(1234),
			uint64(12345),
			uint64(123456),
		},
	}

	state := transport.SessionHandlerState{
		Packet:                        updatePacket,
		Metrics:                       metrics.SessionUpdateMetrics,
		RouteMatrix:                   routeMatrix,
		Datacenter:                    dc,
		Buyer:                         buyer,
		DatacenterAccelerationEnabled: true,
	}

	assert.False(t, state.Buyer.InternalConfig.LargeCustomer)
	assert.True(t, transport.SessionUpdateNearRelayStats(&state))
	assert.False(t, state.Output.HoldNearRelays)
	assert.Equal(t, int32(0), state.NearRelayIndices[0])
	assert.Equal(t, int32(1), state.NearRelayIndices[1])
	assert.Equal(t, int32(-1), state.NearRelayIndices[2])
}

func TestSessionUpdateHandlerFunc_SessionMakeRouteDecision_NextWithoutRouteRelays(t *testing.T) {
	t.Parallel()

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

func TestSessionUpdateHandlerFunc_SessionMakeRouteDecision_SDKAbortedSession(t *testing.T) {
	t.Parallel()

	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	state := transport.SessionHandlerState{
		Metrics: metrics.SessionUpdateMetrics,
		Input: transport.SessionData{
			RouteState: core.RouteState{
				Next: true,
			},
			RouteNumRelays: 5,
		},
		Packet: transport.SessionUpdatePacket{
			Next: false,
		},
	}

	transport.SessionMakeRouteDecision(&state)

	assert.False(t, state.Output.RouteState.Next)
	assert.True(t, state.Output.RouteState.Veto)
	assert.Equal(t, float64(1), state.Metrics.SDKAborted.Value())
}

type testLocator struct{}

func (locator *testLocator) LocateIP(ip net.IP) (routing.Location, error) {
	return routing.Location{
		Latitude:  10,
		Longitude: 10,
	}, nil
}

func TestSessionUpdateHandlerFunc_BuyerNotFound_NoResponse(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false, false)
	datacenter := env.AddDatacenter("datacenter.name")

	unknownPublicKey, unknownPrivateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	unknownBuyerID := binary.LittleEndian.Uint64(unknownPublicKey[:8])

	unknownPublicKey = unknownPublicKey[8:]
	unknownPrivateKey = unknownPrivateKey[8:]

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	getIPLocatorFunc := func() *routing.MaxmindDB {
		return getSuccessLocator(t)
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], unknownPrivateKey)

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", env.GetDatabaseWrapper)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	sessionUpdateConfig := test.SessionUpdatePacketConfig{
		Version:      transport.SDKVersionLatest,
		BuyerID:      unknownBuyerID,
		DatacenterID: datacenter.ID,
		SessionID:    123456789,
		SliceNumber:  0,
		Next:         false,
		PublicKey:    unknownPublicKey,
		PrivateKey:   unknownPrivateKey,
	}

	requestData := env.GenerateSessionUpdatePacket(sessionUpdateConfig)

	handler := transport.SessionUpdateHandlerFunc(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, env.GetDatabaseWrapper, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	// Buyer not found - no response
	assert.Equal(t, 0, len(responseBuffer.Bytes()))
}

func TestSessionUpdateHandlerFunc_SigCheckFailed_NoResponse(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, publicKey, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("datacenter.name")

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	getIPLocatorFunc := func() *routing.MaxmindDB {
		return getSuccessLocator(t)
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], privateKey)

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", env.GetDatabaseWrapper)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	sessionUpdateConfig := test.SessionUpdatePacketConfig{
		Version:      transport.SDKVersionLatest,
		BuyerID:      buyerID,
		DatacenterID: datacenter.ID,
		SessionID:    123456789,
		SliceNumber:  0,
		Next:         false,
		PublicKey:    publicKey,
		PrivateKey:   privateKey[2:],
	}

	requestData := env.GenerateSessionUpdatePacket(sessionUpdateConfig)

	handler := transport.SessionUpdateHandlerFunc(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, env.GetDatabaseWrapper, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	// SigCheck Failed - no response
	assert.Equal(t, 0, len(responseBuffer.Bytes()))
}

func TestSessionUpdateHandlerFunc_DirectResponse(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, publicKey, privateKey := env.AddBuyer("local", true, false)
	datacenter := env.AddDatacenter("losangeles")
	env.AddRelay("los.angeles.1", "10.0.0.2", datacenter.ID)
	env.AddRelay("los.angeles.2", "10.0.0.3", datacenter.ID)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	getIPLocatorFunc := func() *routing.MaxmindDB {
		return getSuccessLocator(t)
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], privateKey)

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", env.GetDatabaseWrapper)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	sessionUpdateConfig := test.SessionUpdatePacketConfig{
		Version:      transport.SDKVersionLatest,
		BuyerID:      buyerID,
		DatacenterID: datacenter.ID,
		SessionID:    123456789,
		SliceNumber:  0,
		Next:         false,
		PublicKey:    publicKey,
		PrivateKey:   privateKey,
	}

	requestData := env.GenerateSessionUpdatePacket(sessionUpdateConfig)

	handler := transport.SessionUpdateHandlerFunc(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, env.GetDatabaseWrapper, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = transport.SDKVersionLatest // Make sure we unmarshal the response the same way we marshaled the request
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, int32(routing.RouteTypeDirect), responsePacket.RouteType)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.False(t, sessionData.RouteState.Next)
}

func TestSessionUpdateHandlerFunc_SessionMakeRouteDecision_NextResponse(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, publicKey, privateKey := env.AddBuyer("local", true, false)
	datacenterLA := env.AddDatacenter("losangeles")
	env.AddRelay("losangeles.1", "10.0.0.2", datacenterLA.ID)
	datacenterChicago := env.AddDatacenter("chicago")
	env.AddRelay("chicago.1", "10.0.0.4", datacenterChicago.ID)
	env.SetCost("losangeles.1", "chicago.1", 10)

	env.DatabaseWrapper.DatacenterMaps[buyerID][datacenterLA.ID] = routing.DatacenterMap{
		BuyerID:            buyerID,
		DatacenterID:       datacenterLA.ID,
		EnableAcceleration: true,
	}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	getIPLocatorFunc := func() *routing.MaxmindDB {
		return getSuccessLocator(t)
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], privateKey)

	relayIDs := env.GetRelayIds()

	for i, id := range relayIDs {
		env.DatabaseWrapper.RelayMap[id] = routing.Relay{
			Addr:      env.GetRelayAddresses()[i],
			PublicKey: publicKey,
		}
	}

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", env.GetDatabaseWrapper)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	sessionDataConfig := test.SessionDataConfig{
		Version:     transport.SessionDataVersion,
		Initial:     false,
		SessionID:   123456789,
		SliceNumber: 3,
	}

	sessionDataPacket, sessionDataSize := env.GenerateSessionDataPacket(sessionDataConfig)

	sessionUpdateConfig := test.SessionUpdatePacketConfig{
		Version:          transport.SDKVersionLatest,
		BuyerID:          buyerID,
		DatacenterID:     datacenterLA.ID,
		SessionID:        sessionDataConfig.SessionID,
		SliceNumber:      sessionDataConfig.SliceNumber,
		PublicKey:        publicKey,
		PrivateKey:       privateKey,
		SessionData:      sessionDataPacket,
		SessionDataBytes: int32(sessionDataSize),
		NearRelayRTT:     10,
		NearRelayJitter:  10,
		NearRelayPL:      1,
		NextRTT:          10,
		NextJitter:       10,
		NextPacketLoss:   1,
		DirectRTT:        1000,
		DirectJitter:     1000,
		DirectPacketLoss: 100,
		ClientAddress:    "10.0.0.9",
		ServerAddress:    "10.0.0.10",
		UserHash:         100,
	}

	requestData := env.GenerateSessionUpdatePacket(sessionUpdateConfig)

	handler := transport.SessionUpdateHandlerFunc(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, env.GetDatabaseWrapper, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = transport.SDKVersionLatest // Make sure we unmarshal the response the same way we marshaled the request
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
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, publicKey, privateKey := env.AddBuyer("local", true, false)
	datacenterLA := env.AddDatacenter("losangeles")
	env.AddRelay("losangeles.1", "10.0.0.2", datacenterLA.ID)
	datacenterChicago := env.AddDatacenter("chicago")
	env.AddRelay("chicago.1", "10.0.0.4", datacenterChicago.ID)
	env.SetCost("losangeles.1", "chicago.1", 10)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	getIPLocatorFunc := func() *routing.MaxmindDB {
		return getSuccessLocator(t)
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		return env.GetRouteMatrix()
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], privateKey)

	relayIDs := env.GetRelayIds()

	for i, id := range relayIDs {
		env.DatabaseWrapper.RelayMap[id] = routing.Relay{
			Addr:      env.GetRelayAddresses()[i],
			PublicKey: publicKey,
		}
	}

	env.DatabaseWrapper.DatacenterMaps[buyerID][datacenterLA.ID] = routing.DatacenterMap{
		BuyerID:            buyerID,
		DatacenterID:       datacenterLA.ID,
		EnableAcceleration: true,
	}

	localMultiPathVetoHandler, err := storage.NewLocalMultipathVetoHandler("", env.GetDatabaseWrapper)
	assert.NoError(t, err)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	sessionDataConfig := test.SessionDataConfig{
		Version:        transport.SessionDataVersion,
		Initial:        true,
		SessionID:      123456789,
		SliceNumber:    5,
		RouteNumRelays: 2,
		RouteRelayIDs:  [5]uint64{relayIDs[0], relayIDs[1], 0, 0, 0},
		RouteState: core.RouteState{
			UserID:          100,
			Next:            true,
			NumNearRelays:   2,
			NearRelayRTT:    [32]int32{10, 10},
			NearRelayJitter: [32]int32{10, 10},
		},
	}

	sessionDataPacket, sessionDataSize := env.GenerateSessionDataPacket(sessionDataConfig)

	sessionUpdateConfig := test.SessionUpdatePacketConfig{
		Version:          transport.SDKVersionLatest,
		BuyerID:          buyerID,
		DatacenterID:     datacenterLA.ID,
		SessionID:        sessionDataConfig.SessionID,
		SliceNumber:      sessionDataConfig.SliceNumber,
		PublicKey:        publicKey,
		PrivateKey:       privateKey,
		SessionData:      sessionDataPacket,
		SessionDataBytes: int32(sessionDataSize),
		Next:             true,
		Committed:        true,
		NearRelayRTT:     10,
		NearRelayJitter:  10,
		NearRelayPL:      1,
		NextRTT:          11,
		NextJitter:       11,
		NextPacketLoss:   1,
		DirectRTT:        1000,
		DirectJitter:     1000,
		DirectPacketLoss: 100,
		ClientAddress:    "10.0.0.9",
		ServerAddress:    "10.0.0.10",
		UserHash:         100,
	}

	requestData := env.GenerateSessionUpdatePacket(sessionUpdateConfig)

	handler := transport.SessionUpdateHandlerFunc(getIPLocatorFunc, getRouteMatrixFunc, localMultiPathVetoHandler, env.GetDatabaseWrapper, routerPrivateKey, postSessionHandler, metrics.SessionUpdateMetrics, time.Minute)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = transport.SDKVersionLatest // Make sure we unmarshal the response the same way we marshaled the request
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, int32(routing.RouteTypeContinue), responsePacket.RouteType)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.True(t, sessionData.RouteState.Next)
	assert.False(t, sessionData.Initial)
}

func TestCalculateTotalPriceNibblins_SellerPrice(t *testing.T) {
	t.Parallel()

	routeNumRelays := 2

	seller := routing.Seller{
		ID:                       "test-seller",
		Name:                     "test-seller",
		CompanyCode:              "test-seller",
		ShortName:                "test-seller",
		Secret:                   false,
		EgressPriceNibblinsPerGB: routing.Nibblin(10000000000000),
		DatabaseID:               0,
		CustomerID:               0,
	}

	var relaySellers [core.MaxRelaysPerRoute]routing.Seller
	var relayEgressPriceOverride [core.MaxRelaysPerRoute]routing.Nibblin

	assert.Less(t, routeNumRelays, core.MaxRelaysPerRoute)
	for i := 0; i < routeNumRelays; i++ {
		relaySellers[i] = seller
		relayEgressPriceOverride[i] = routing.Nibblin(0)
	}

	envelopeBytesDown := uint64(1)
	envelopeBytesUp := uint64(1)

	totalPrice := transport.CalculateTotalPriceNibblins(routeNumRelays, relaySellers, relayEgressPriceOverride, envelopeBytesUp, envelopeBytesDown)

	assert.Equal(t, routing.Nibblin(40002), totalPrice)
}

func TestCalculateTotalPriceNibblins_EgressPriceOverride(t *testing.T) {
	t.Parallel()

	routeNumRelays := 2

	seller := routing.Seller{
		ID:                       "test-seller",
		Name:                     "test-seller",
		CompanyCode:              "test-seller",
		ShortName:                "test-seller",
		Secret:                   false,
		EgressPriceNibblinsPerGB: routing.Nibblin(10000000000000),
		DatabaseID:               0,
		CustomerID:               0,
	}

	var relaySellers [core.MaxRelaysPerRoute]routing.Seller
	var relayEgressPriceOverride [core.MaxRelaysPerRoute]routing.Nibblin

	assert.Less(t, routeNumRelays, core.MaxRelaysPerRoute)
	for i := 0; i < routeNumRelays; i++ {
		relaySellers[i] = seller
		relayEgressPriceOverride[i] = routing.Nibblin(20000000000000)
	}

	envelopeBytesDown := uint64(1)
	envelopeBytesUp := uint64(1)

	totalPrice := transport.CalculateTotalPriceNibblins(routeNumRelays, relaySellers, relayEgressPriceOverride, envelopeBytesUp, envelopeBytesDown)

	assert.Equal(t, routing.Nibblin(80002), totalPrice)
}

func TestCalculateRouteRelaysPrice_SellerPrice(t *testing.T) {
	t.Parallel()

	routeNumRelays := 2

	seller := routing.Seller{
		ID:                       "test-seller",
		Name:                     "test-seller",
		CompanyCode:              "test-seller",
		ShortName:                "test-seller",
		Secret:                   false,
		EgressPriceNibblinsPerGB: routing.Nibblin(10000000000000),
		DatabaseID:               0,
		CustomerID:               0,
	}

	var relaySellers [core.MaxRelaysPerRoute]routing.Seller
	var relayEgressPriceOverride [core.MaxRelaysPerRoute]routing.Nibblin

	assert.Less(t, routeNumRelays, core.MaxRelaysPerRoute)
	for i := 0; i < routeNumRelays; i++ {
		relaySellers[i] = seller
		relayEgressPriceOverride[i] = routing.Nibblin(0)
	}

	envelopeBytesDown := uint64(1)
	envelopeBytesUp := uint64(1)

	routeRelaysPrice := transport.CalculateRouteRelaysPrice(routeNumRelays, relaySellers, relayEgressPriceOverride, envelopeBytesUp, envelopeBytesDown)

	expectedRouteRelaysPrice := [core.MaxRelaysPerRoute]routing.Nibblin{routing.Nibblin(20000), routing.Nibblin(20000), routing.Nibblin(0), routing.Nibblin(0), routing.Nibblin(0)}

	assert.Equal(t, expectedRouteRelaysPrice, routeRelaysPrice)
}

func TestCalculateRouteRelaysPrice_EgressPriceOverride(t *testing.T) {
	t.Parallel()

	routeNumRelays := 2

	seller := routing.Seller{
		ID:                       "test-seller",
		Name:                     "test-seller",
		CompanyCode:              "test-seller",
		ShortName:                "test-seller",
		Secret:                   false,
		EgressPriceNibblinsPerGB: routing.Nibblin(10000000000000),
		DatabaseID:               0,
		CustomerID:               0,
	}

	var relaySellers [core.MaxRelaysPerRoute]routing.Seller
	var relayEgressPriceOverride [core.MaxRelaysPerRoute]routing.Nibblin

	assert.Less(t, routeNumRelays, core.MaxRelaysPerRoute)
	for i := 0; i < routeNumRelays; i++ {
		relaySellers[i] = seller
		relayEgressPriceOverride[i] = routing.Nibblin(20000000000000)
	}

	envelopeBytesDown := uint64(1)
	envelopeBytesUp := uint64(1)

	routeRelaysPrice := transport.CalculateRouteRelaysPrice(routeNumRelays, relaySellers, relayEgressPriceOverride, envelopeBytesUp, envelopeBytesDown)

	expectedRouteRelaysPrice := [core.MaxRelaysPerRoute]routing.Nibblin{routing.Nibblin(40000), routing.Nibblin(40000), routing.Nibblin(0), routing.Nibblin(0), routing.Nibblin(0)}

	assert.Equal(t, expectedRouteRelaysPrice, routeRelaysPrice)
}

// Match data handler tests

func TestMatchDataHandlerFunc_BuyerNotFound(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	env.AddBuyer("local", false, false)

	unknownPublicKey, unknownPrivateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	unknownPrivateKey = unknownPrivateKey[8:]

	unknownBuyerID := binary.LittleEndian.Uint64(unknownPublicKey[:8])

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)

	matchDataConfig := test.MatchDataPacketConfig{
		Version:       transport.SDKVersionLatest,
		BuyerID:       unknownBuyerID,
		ServerAddress: *serverAddr,
		DatacenterID:  rand.Uint64(),
		UserHash:      rand.Uint64(),
		SessionID:     crypto.GenerateSessionID(),
		MatchID:       rand.Uint64(),
		PrivateKey:    unknownPrivateKey,
	}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateMatchDataRequestPacket(matchDataConfig)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.MatchDataHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, metrics.MatchDataHandlerMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.MatchDataResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.MatchDataResponseUnknownBuyer), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.MatchDataHandlerMetrics.BuyerNotFound.Value())
}

func TestMatchDataHandlerFunc_BuyerNotLive(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", false, false)

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)

	matchDataConfig := test.MatchDataPacketConfig{
		Version:       transport.SDKVersionLatest,
		BuyerID:       buyerID,
		ServerAddress: *serverAddr,
		DatacenterID:  rand.Uint64(),
		UserHash:      rand.Uint64(),
		SessionID:     crypto.GenerateSessionID(),
		MatchID:       rand.Uint64(),
		PrivateKey:    privateKey,
	}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateMatchDataRequestPacket(matchDataConfig)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.MatchDataHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, metrics.MatchDataHandlerMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.MatchDataResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.MatchDataResponseBuyerNotActive), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.MatchDataHandlerMetrics.BuyerNotActive.Value())
}

func TestMatchDataHandlerFunc_SigCheckFail(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)

	matchDataConfig := test.MatchDataPacketConfig{
		Version:       transport.SDKVersionLatest,
		BuyerID:       buyerID,
		ServerAddress: *serverAddr,
		DatacenterID:  rand.Uint64(),
		UserHash:      rand.Uint64(),
		SessionID:     crypto.GenerateSessionID(),
		MatchID:       rand.Uint64(),
		PrivateKey:    privateKey[2:],
	}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateMatchDataRequestPacket(matchDataConfig)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.MatchDataHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, metrics.MatchDataHandlerMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.MatchDataResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.MatchDataResponseSignatureCheckFailed), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.MatchDataHandlerMetrics.SignatureCheckFailed.Value())
}

func TestMatchDataHandlerFunc_Success(t *testing.T) {
	t.Parallel()

	env := test.NewTestEnvironment(t)
	buyerID, _, privateKey := env.AddBuyer("local", true, false)

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)

	var matchValues [transport.MaxMatchValues]float64
	for i := 0; i < transport.MaxMatchValues; i++ {
		matchValues[i] = rand.ExpFloat64()
	}

	matchDataConfig := test.MatchDataPacketConfig{
		Version:        transport.SDKVersionLatest,
		BuyerID:        buyerID,
		ServerAddress:  *serverAddr,
		DatacenterID:   rand.Uint64(),
		UserHash:       rand.Uint64(),
		SessionID:      crypto.GenerateSessionID(),
		MatchID:        rand.Uint64(),
		NumMatchValues: transport.MaxMatchValues,
		MatchValues:    matchValues,
		PrivateKey:     privateKey,
	}

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &env.MetricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)

	requestData := env.GenerateMatchDataRequestPacket(matchDataConfig)

	postSessionHandler := transport.NewPostSessionHandler(4, 0, nil, 10, &billing.NoOpBiller{}, true, &md.NoOpMatcher{}, metrics.PostSessionMetrics)

	handler := transport.MatchDataHandlerFunc(env.GetDatabaseWrapper, postSessionHandler, metrics.MatchDataHandlerMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.MatchDataResponsePacket
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	assert.Equal(t, uint32(transport.MatchDataResponseOK), responsePacket.Response)
	assert.Equal(t, float64(1), metrics.MatchDataHandlerMetrics.HandlerMetrics.Invocations.Value())
	assert.Zero(t, metrics.MatchDataHandlerMetrics.BuyerNotFound.Value())
	assert.Zero(t, metrics.MatchDataHandlerMetrics.BuyerNotActive.Value())
	assert.Zero(t, metrics.MatchDataHandlerMetrics.SignatureCheckFailed.Value())
}
