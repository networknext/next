package handlers

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/packets"

	"github.com/networknext/backend/modules-old/crypto"
	"github.com/networknext/backend/modules-old/routing"
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
	handler        SDK5_Handler
	conn           *net.UDPConn
	from           *net.UDPAddr
	signPublicKey  [packets.NEXT_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
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

	fmt.Printf("server backend address is %s\n", harness.handler.ServerBackendAddress.String())

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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
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

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
}

// ---------------------------------------------------------------------------------------

// tests for the server init handler

func Test_ServerInitHandler_BuyerNotLive_SDK5(t *testing.T) {

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

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
}

func Test_ServerInitHandler_BuyerSDKTooOld_SDK5(t *testing.T) {

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

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
}

func Test_ServerInitHandler_UnknownDatacenter_SDK5(t *testing.T) {

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

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
}

func Test_ServerInitHandler_ServerInitResponse_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a UDP socket to listen on so we can get the response packet

	ctx := context.Background()

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(ctx, "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind client socket")
	}

	clientConn := lp.(*net.UDPConn)

	clientPort := clientConn.LocalAddr().(*net.UDPAddr).Port

	clientAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", clientPort))

	fmt.Printf("client address is %s\n", clientAddress.String())

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

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// setup "local" datacenter in the database

	localDatacenterId := crypto.HashID("local")

	localDatacenter := routing.Datacenter{
		ID:   localDatacenterId,
		Name: "local",
		Location: routing.Location{
			Latitude:  10,
			Longitude: 20,
		},
	}

	harness.handler.Database.DatacenterMap = make(map[uint64]routing.Datacenter)
	harness.handler.Database.DatacenterMap[localDatacenterId] = localDatacenter

	// construct a valid, signed server init request packet

	requestId := uint64(0x12345)

	packet := packets.SDK5_ServerInitRequestPacket{
		Version:        packets.SDKVersion{5, 0, 0},
		BuyerId:        buyerId,
		RequestId:      requestId,
		DatacenterId:   crypto.HashID("local"),
		DatacenterName: "local",
	}

	packetData, err := SDK5_WritePacket(&packet, packets.SDK5_SERVER_INIT_REQUEST_PACKET, 1500, clientAddress, &harness.handler.ServerBackendAddress, buyerPrivateKey[:])
	if err != nil {
		core.Error("failed to write server init request packet: %v", err)
		return
	}

	// setup a goroutine to listen for response packets from the packet handler

	var receivedResponse uint64

	go func() {

		for {

			var buffer [4096]byte

			packetBytes, from, err := clientConn.ReadFromUDP(buffer[:])
			if err != nil {
				core.Debug("failed to read udp packet: %v", err)
				continue
			}

			core.Debug("received response packet from %s", from.String())

			packetData := buffer[:packetBytes]

			// ignore any packets that aren't from the server backend we're testing

			if from.String() != harness.handler.ServerBackendAddress.String() {
				core.Debug("not from server backend")
				continue
			}

			// ignore any packets that are not server init response packets

			if packetData[0] != packets.SDK5_SERVER_INIT_RESPONSE_PACKET {
				core.Debug("wrong packet type")
				continue
			}

			// ignore any packets that are too small

			if len(packetData) < 16+3+4+packets.NEXT_CRYPTO_SIGN_BYTES+2 {
				core.Debug("too small")
				continue
			}

			// make sure basic packet filter passes

			if !core.BasicPacketFilter(packetData[:], len(packetData)) {
				core.Debug("basic packet filter failed")
				continue
			}

			// make sure advanced packet filter passes

			var emptyMagic [8]byte

			var fromAddressBuffer [32]byte
			var toAddressBuffer [32]byte

			fromAddressData, fromAddressPort := core.GetAddressData(&harness.handler.ServerBackendAddress, fromAddressBuffer[:])
			toAddressData, toAddressPort := core.GetAddressData(clientAddress, toAddressBuffer[:])

			if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packetData)) {
				core.Debug("advanced packet filter failed")
				continue
			}

			// make sure packet signature check passes

			if !SDK5_CheckPacketSignature(packetData, harness.signPublicKey[:]) {
				core.Debug("packet signature check failed")
				return
			}

			// read packet

			packetData = packetData[16 : len(packetData)-(2+packets.NEXT_CRYPTO_SIGN_BYTES)]

			responsePacket := packets.SDK5_ServerInitResponsePacket{}
			if err := packets.ReadPacket(packetData, &responsePacket); err != nil {
				core.Debug("could not read server init response packet")
				continue
			}

			// check all response packet fields match expected values

			assert.Equal(t, packet.RequestId, responsePacket.RequestId)
			assert.Equal(t, packets.SDK5_ServerInitResponseOK, int(responsePacket.Response))

			// success!

			atomic.AddUint64(&receivedResponse, 1)
			break
		}
	}()

	// loop sending the request packet until we get a response or time out

	harness.from = clientAddress

	for i := 0; i < 100; i++ {

		response := atomic.LoadUint64(&receivedResponse)
		if response != 0 {
			break
		}

		SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_PacketTooSmall])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_UnsupportedPacketType])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_BasicPacketFilterFailed])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_AdvancedPacketFilterFailed])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_NoRouteMatrix])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_NoDatabase])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SignatureCheckFailed])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_BuyerNotLive])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SDKTooOld])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_UnknownDatacenter])
		
		assert.True(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
		assert.True(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])

		if i > 10 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// verify that we received a response

	assert.True(t, receivedResponse != 0)
}

// ---------------------------------------------------------------------------------------

// tests for the server update handler

func Test_ServerUpdateHandler_BuyerNotLive_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK5_SERVER_UPDATE_REQUEST_PACKET
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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
}

func Test_ServerUpdateHandler_BuyerSDKTooOld_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK5_SERVER_UPDATE_REQUEST_PACKET
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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
}

func Test_ServerUpdateHandler_UnknownDatacenter_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK5_SERVER_UPDATE_REQUEST_PACKET
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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
}

func Test_ServerUpdateHandler_ServerUpdateResponse_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a UDP socket to listen on so we can get the response packet

	ctx := context.Background()

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(ctx, "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind client socket")
	}

	clientConn := lp.(*net.UDPConn)

	clientPort := clientConn.LocalAddr().(*net.UDPAddr).Port

	clientAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", clientPort))

	fmt.Printf("client address is %s\n", clientAddress.String())

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

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// setup "local" datacenter in the database

	localDatacenterId := crypto.HashID("local")

	localDatacenter := routing.Datacenter{
		ID:   localDatacenterId,
		Name: "local",
		Location: routing.Location{
			Latitude:  10,
			Longitude: 20,
		},
	}

	harness.handler.Database.DatacenterMap = make(map[uint64]routing.Datacenter)
	harness.handler.Database.DatacenterMap[localDatacenterId] = localDatacenter

	// construct a valid, signed server update request packet

	requestId := uint64(0x12345)

	packet := packets.SDK5_ServerUpdateRequestPacket{
		Version:      packets.SDKVersion{5, 0, 0},
		BuyerId:      buyerId,
		RequestId:    requestId,
		DatacenterId: crypto.HashID("local"),
	}

	packetData, err := SDK5_WritePacket(&packet, packets.SDK5_SERVER_UPDATE_REQUEST_PACKET, 1500, clientAddress, &harness.handler.ServerBackendAddress, buyerPrivateKey[:])
	if err != nil {
		core.Error("failed to write server update request packet: %v", err)
		return
	}

	// setup a goroutine to listen for response packets from the packet handler

	var receivedResponse uint64

	go func() {

		for {

			var buffer [4096]byte

			packetBytes, from, err := clientConn.ReadFromUDP(buffer[:])
			if err != nil {
				core.Debug("failed to read udp packet: %v", err)
				continue
			}

			core.Debug("received response packet from %s", from.String())

			packetData := buffer[:packetBytes]

			// ignore any packets that aren't from the server backend we're testing

			if from.String() != harness.handler.ServerBackendAddress.String() {
				core.Debug("not from server backend")
				continue
			}

			// ignore any packets that are not server update response packets

			if packetData[0] != packets.SDK5_SERVER_UPDATE_RESPONSE_PACKET {
				core.Debug("wrong packet type")
				continue
			}

			// ignore any packets that are too small

			if len(packetData) < 16+3+4+packets.NEXT_CRYPTO_SIGN_BYTES+2 {
				core.Debug("too small")
				continue
			}

			// make sure basic packet filter passes

			if !core.BasicPacketFilter(packetData[:], len(packetData)) {
				core.Debug("basic packet filter failed")
				continue
			}

			// make sure advanced packet filter passes

			var emptyMagic [8]byte

			var fromAddressBuffer [32]byte
			var toAddressBuffer [32]byte

			fromAddressData, fromAddressPort := core.GetAddressData(&harness.handler.ServerBackendAddress, fromAddressBuffer[:])
			toAddressData, toAddressPort := core.GetAddressData(clientAddress, toAddressBuffer[:])

			if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packetData)) {
				core.Debug("advanced packet filter failed")
				continue
			}

			// make sure packet signature check passes

			if !SDK5_CheckPacketSignature(packetData, harness.signPublicKey[:]) {
				core.Debug("packet signature check failed")
				return
			}

			// read packet

			packetData = packetData[16 : len(packetData)-(2+packets.NEXT_CRYPTO_SIGN_BYTES)]

			responsePacket := packets.SDK5_ServerUpdateResponsePacket{}
			if err := packets.ReadPacket(packetData, &responsePacket); err != nil {
				core.Debug("could not read server update response packet")
				continue
			}

			// check all response packet fields match expected values

			assert.Equal(t, packet.RequestId, responsePacket.RequestId)

			// success!

			atomic.AddUint64(&receivedResponse, 1)
			break
		}
	}()

	// loop sending the request packet until we get a response or time out

	harness.from = clientAddress

	for i := 0; i < 100; i++ {

		response := atomic.LoadUint64(&receivedResponse)
		if response != 0 {
			break
		}

		SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_PacketTooSmall])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_UnsupportedPacketType])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_BasicPacketFilterFailed])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_AdvancedPacketFilterFailed])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_NoRouteMatrix])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_NoDatabase])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SignatureCheckFailed])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_BuyerNotLive])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SDKTooOld])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_UnknownDatacenter])

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
		assert.True(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
		assert.True(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])

		if i > 10 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// verify that we received a response

	assert.True(t, receivedResponse != 0)
}

// ---------------------------------------------------------------------------------------

// tests for the match data handler

func Test_MatchUpdateHandler_BuyerNotLive_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK5_MATCH_DATA_REQUEST_PACKET
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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
}

func Test_MatchDataHandler_BuyerSDKTooOld_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK5_MATCH_DATA_REQUEST_PACKET
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

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])
}

func Test_MatchDataHandler_MatchDataResponse_SDK5(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a UDP socket to listen on so we can get the response packet

	ctx := context.Background()

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(ctx, "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind client socket")
	}

	clientConn := lp.(*net.UDPConn)

	clientPort := clientConn.LocalAddr().(*net.UDPAddr).Port

	clientAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", clientPort))

	fmt.Printf("client address is %s\n", clientAddress.String())

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

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// setup "local" datacenter in the database

	localDatacenterId := crypto.HashID("local")

	localDatacenter := routing.Datacenter{
		ID:   localDatacenterId,
		Name: "local",
		Location: routing.Location{
			Latitude:  10,
			Longitude: 20,
		},
	}

	harness.handler.Database.DatacenterMap = make(map[uint64]routing.Datacenter)
	harness.handler.Database.DatacenterMap[localDatacenterId] = localDatacenter

	// construct a valid, signed match data request packet

	packet := packets.SDK5_MatchDataRequestPacket{
		Version:        packets.SDKVersion{5, 0, 0},
		BuyerId:        buyerId,
		ServerAddress:  *core.ParseAddress("127.0.0.1:10000"),
		DatacenterId:   crypto.HashID("local"),
		UserHash:       uint64(123456789213),
		SessionId:      uint64(5213412421413),
		RetryNumber:    2,
		MatchId:        uint64(112312737131),
		NumMatchValues: 64,
	}

	for i := 0; i < 64; i++ {
		packet.MatchValues[i] = float64(i)
	}

	packetData, err := SDK5_WritePacket(&packet, packets.SDK5_MATCH_DATA_REQUEST_PACKET, 1500, clientAddress, &harness.handler.ServerBackendAddress, buyerPrivateKey[:])
	if err != nil {
		core.Error("failed to write match data request packet: %v", err)
		return
	}

	// setup a goroutine to listen for response packets from the packet handler

	var receivedResponse uint64

	go func() {

		for {

			var buffer [4096]byte

			packetBytes, from, err := clientConn.ReadFromUDP(buffer[:])
			if err != nil {
				core.Debug("failed to read udp packet: %v", err)
				continue
			}

			core.Debug("received response packet from %s", from.String())

			packetData := buffer[:packetBytes]

			// ignore any packets that aren't from the server backend we're testing

			if from.String() != harness.handler.ServerBackendAddress.String() {
				core.Debug("not from server backend")
				continue
			}

			// ignore any packets that are not server update response packets

			if packetData[0] != packets.SDK5_MATCH_DATA_RESPONSE_PACKET {
				core.Debug("wrong packet type")
				continue
			}

			// ignore any packets that are too small

			if len(packetData) < 16+3+4+packets.NEXT_CRYPTO_SIGN_BYTES+2 {
				core.Debug("too small")
				continue
			}

			// make sure basic packet filter passes

			if !core.BasicPacketFilter(packetData[:], len(packetData)) {
				core.Debug("basic packet filter failed")
				continue
			}

			// make sure advanced packet filter passes

			var emptyMagic [8]byte

			var fromAddressBuffer [32]byte
			var toAddressBuffer [32]byte

			fromAddressData, fromAddressPort := core.GetAddressData(&harness.handler.ServerBackendAddress, fromAddressBuffer[:])
			toAddressData, toAddressPort := core.GetAddressData(clientAddress, toAddressBuffer[:])

			if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packetData)) {
				core.Debug("advanced packet filter failed")
				continue
			}

			// make sure packet signature check passes

			if !SDK5_CheckPacketSignature(packetData, harness.signPublicKey[:]) {
				core.Debug("packet signature check failed")
				return
			}

			// read packet

			packetData = packetData[16 : len(packetData)-(2+packets.NEXT_CRYPTO_SIGN_BYTES)]

			responsePacket := packets.SDK5_MatchDataResponsePacket{}
			if err := packets.ReadPacket(packetData, &responsePacket); err != nil {
				core.Debug("could not read match data response packet")
				continue
			}

			// check all response packet fields match expected values

			assert.Equal(t, packet.SessionId, responsePacket.SessionId)

			// success!

			atomic.AddUint64(&receivedResponse, 1)
			break
		}
	}()

	// loop sending the request packet until we get a response or time out

	harness.from = clientAddress

	for i := 0; i < 100; i++ {

		response := atomic.LoadUint64(&receivedResponse)
		if response != 0 {
			break
		}

		SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_PacketTooSmall])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_UnsupportedPacketType])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_BasicPacketFilterFailed])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_AdvancedPacketFilterFailed])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_NoRouteMatrix])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_NoDatabase])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SignatureCheckFailed])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_BuyerNotLive])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SDKTooOld])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_UnknownDatacenter])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket])
		assert.True(t, harness.handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket])

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket])
		assert.True(t, harness.handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket])

		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket])

		if i > 10 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// verify that we received a response

	assert.True(t, receivedResponse != 0)
}

// ---------------------------------------------------------------------------------------
