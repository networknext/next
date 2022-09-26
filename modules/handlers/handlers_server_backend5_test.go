package handlers

import (
	"testing"
	"context"
	"net"
	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/packets"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/crypto"
)

// ---------------------------------------------------------------------------------------

// tests that apply to the SDK5 handler for all packet types

func getMagicValues() ([]byte, []byte, []byte) {
	upcoming := make([]byte, 8)
	current := make([]byte, 8)
	previous := make([]byte, 8)
	for i := 0; i < 8; i++ {
		upcoming[i] = 1
		current[i] = 2
		previous[i] = 3
	}
	return upcoming, current, previous
}

type TestHarness struct {
	handler SDK5_Handler
	conn *net.UDPConn
	from *net.UDPAddr
	signPublicKey [packets.NEXT_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	signPrivateKey [packets.NEXT_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
}

func CreateTestHarness() *TestHarness {

	harness := TestHarness{}

	ctx := context.Background()

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(ctx, "udp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	harness.conn = lp.(*net.UDPConn)

	backendPort := harness.conn.LocalAddr().(*net.UDPAddr).Port

	harness.handler = SDK5_Handler{}
	harness.handler.MaxPacketSize = 4096
	harness.handler.ServerBackendAddress = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", backendPort))
	harness.handler.GetMagicValues = getMagicValues

	harness.from = core.ParseAddress("127.0.0.1:10000")

	SDK5_SignKeypair(harness.signPublicKey[:], harness.signPrivateKey[:])

	harness.handler.PrivateKey = harness.signPrivateKey[:]

	return &harness
}

func TestPacketTooSmall_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 10)

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_PacketTooSmall])
}

func TestUnsupportedPacketType_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	for i := 0; i < 256; i++ {
	
		packetType := uint8(i)

		if packetType == packets.SDK5_SERVER_INIT_REQUEST_PACKET || packetType == packets.SDK5_SERVER_UPDATE_REQUEST_PACKET || packetType == packets.SDK5_SESSION_UPDATE_REQUEST_PACKET || packetType == packets.SDK5_MATCH_DATA_REQUEST_PACKET {
			continue
		}

		packetData[0] = packetType

		SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

		assert.True(t, harness.handler.Events[SDK5_HandlerEvent_UnsupportedPacketType])
	}
}

func TestBasicPacketFilterFailed_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	packetData[0] = packets.SDK5_SERVER_INIT_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_BasicPacketFilterFailed])
}

func TestAdvancedPacketFilterFailed_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	packetData[0] = packets.SDK5_SERVER_INIT_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// intentionally incorrect inputs -> will pass basic packet filter, but fail advanced
	magic := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	fromAddress := [4]byte{1, 2, 3, 4}
	toAddress := [4]byte{4, 3, 2, 1}
	fromPort := uint16(1000)
	toPort := uint16(5000)
	packetLength := len(packetData)

	core.GenerateChonkle(packetData[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	core.GeneratePittle(packetData[len(packetData)-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_AdvancedPacketFilterFailed])
}

func TestNoRouteMatrix_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	packetData[0] = packets.SDK5_SERVER_INIT_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// correct inputs -> passes advanced packet filter
	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	fromPort := uint16(harness.from.Port)
	toPort := uint16(harness.handler.ServerBackendAddress.Port)
	packetLength := len(packetData)

	core.GenerateChonkle(packetData[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	core.GeneratePittle(packetData[len(packetData)-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_NoRouteMatrix])
}

func TestNoDatabase_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	packetData[0] = packets.SDK5_SERVER_INIT_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// correct inputs -> passes advanced packet filter
	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	fromPort := uint16(harness.from.Port)
	toPort := uint16(harness.handler.ServerBackendAddress.Port)
	packetLength := len(packetData)

	core.GenerateChonkle(packetData[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	core.GeneratePittle(packetData[len(packetData)-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	harness.handler.RouteMatrix = &common.RouteMatrix{}

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_NoDatabase])
}

func TestUnknownBuyer_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	packetData[0] = packets.SDK5_SERVER_INIT_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// correct inputs -> passes advanced packet filter
	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	fromPort := uint16(harness.from.Port)
	toPort := uint16(harness.handler.ServerBackendAddress.Port)
	packetLength := len(packetData)

	core.GenerateChonkle(packetData[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	core.GeneratePittle(packetData[len(packetData)-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = &routing.DatabaseBinWrapper{}

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_UnknownBuyer])
}

func TestSignatureCheckFailed_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 100)
	packetData[0] = packets.SDK5_SERVER_INIT_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// generate pittle and chonkle so the packet gets through the basic and advanced packet filters

	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	fromPort := uint16(harness.from.Port)
	toPort := uint16(harness.handler.ServerBackendAddress.Port)
	packetLength := len(packetData)

	core.GenerateChonkle(packetData[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	core.GeneratePittle(packetData[len(packetData)-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = &routing.DatabaseBinWrapper{}

	buyerId := uint64(0x1111111122222222)
	
	var buyerPublicKey [packets.NEXT_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.NEXT_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	SDK5_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	harness.handler.Database.BuyerMap = make(map[uint64]routing.Buyer)

	buyer := routing.Buyer{}
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 16 + 3
	common.WriteUint64(packetData[:], &index, buyerId)

	// run the packet through the handler, it should fail the signature check

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_SignatureCheckFailed])
}

// ---------------------------------------------------------------------------------------

// tests for the server init handler

func TestBuyerNotLive_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK5_SERVER_INIT_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// generate pittle and chonkle so the packet gets through the basic and advanced packet filters

	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	fromPort := uint16(harness.from.Port)
	toPort := uint16(harness.handler.ServerBackendAddress.Port)
	packetLength := len(packetData)

	core.GenerateChonkle(packetData[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	core.GeneratePittle(packetData[len(packetData)-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = &routing.DatabaseBinWrapper{}

	buyerId := uint64(0x1111111122222222)
	
	var buyerPublicKey [packets.NEXT_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.NEXT_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	SDK5_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	harness.handler.Database.BuyerMap = make(map[uint64]routing.Buyer)

	buyer := routing.Buyer{}
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 16 + 3
	common.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	SDK5_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, it should pass the signature check then fail on buyer not live

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_BuyerNotLive])
}

func TestBuyerSDKTooOld_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK5_SERVER_INIT_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// generate pittle and chonkle so the packet gets through the basic and advanced packet filters

	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	fromPort := uint16(harness.from.Port)
	toPort := uint16(harness.handler.ServerBackendAddress.Port)
	packetLength := len(packetData)

	core.GenerateChonkle(packetData[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	core.GeneratePittle(packetData[len(packetData)-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = &routing.DatabaseBinWrapper{}

	buyerId := uint64(0x1111111122222222)
	
	var buyerPublicKey [packets.NEXT_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.NEXT_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	SDK5_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	harness.handler.Database.BuyerMap = make(map[uint64]routing.Buyer)

	buyer := routing.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 16 + 3
	common.WriteUint64(packetData[:], &index, buyerId)

	// modify the packet so it has an old SDK version of 1.2.3

	packetData[16] = 1
	packetData[17] = 2	
	packetData[18] = 3

	// actually sign the packet, so it passes the signature check

	SDK5_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see that the SDK is too old

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_SDKTooOld])
}

func TestUnknownDatacenter_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK5_SERVER_INIT_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// generate pittle and chonkle so the packet gets through the basic and advanced packet filters

	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	fromPort := uint16(harness.from.Port)
	toPort := uint16(harness.handler.ServerBackendAddress.Port)
	packetLength := len(packetData)

	core.GenerateChonkle(packetData[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	core.GeneratePittle(packetData[len(packetData)-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = &routing.DatabaseBinWrapper{}

	buyerId := uint64(0x1111111122222222)
	
	var buyerPublicKey [packets.NEXT_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.NEXT_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	SDK5_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	harness.handler.Database.BuyerMap = make(map[uint64]routing.Buyer)

	buyer := routing.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 16 + 3
	common.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	SDK5_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see the datacenter is unknown

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_UnknownDatacenter])
}

func TestServerInitRequestResponse_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = &routing.DatabaseBinWrapper{}

	buyerId := uint64(0x1111111122222222)
	
	var buyerPublicKey [packets.NEXT_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.NEXT_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	SDK5_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	harness.handler.Database.BuyerMap = make(map[uint64]routing.Buyer)

	buyer := routing.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// construct a valid, signed server init request packet

	requestId := uint64(0x12345)

	packet := packets.SDK5_ServerInitRequestPacket{
		Version: packets.SDKVersion{5,0,0},
		BuyerId: buyerId,
		RequestId: requestId,
		DatacenterId: crypto.HashID("local"),
		DatacenterName: "local",
	}

	packetData := make([]byte, 1500)

	writeStream := common.CreateWriteStream(packetData[:])

	// todo: serialize 16 bytes dummy

	err := packet.Serialize(writeStream)
	assert.Nil(t, err)

	// todo: packet type, chonkle, pittle, sign

	writeStream.Flush()

	// setup a UDP socket to listen on so we can get the response

	ctx := context.Background()

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(ctx, "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind client socket")
	}

	clientConn := lp.(*net.UDPConn)

	clientPort := clientConn.LocalAddr().(*net.UDPAddr).Port

	fmt.Printf("client port is %d\n", clientPort)

	// loop to process the packet, until we can get a response, up to n times

	/*
	for {

		packetBytes, from, err := conn.ReadFromUDP(buffer[:])
		if err != nil {
			core.Debug("failed to read udp packet: %v", err)
			break
		}
		*/

	// ...

	_ = clientConn
}

// ---------------------------------------------------------------------------------------

// tests for the server update handler

// ...

// ---------------------------------------------------------------------------------------
