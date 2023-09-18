package handlers

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/database"
	"github.com/networknext/next/modules/encoding"
	"github.com/networknext/next/modules/messages"
	"github.com/networknext/next/modules/packets"
)

// ---------------------------------------------------------------------------------------

// tests that apply to the SDK handler for all packet types

func getMagicValues() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte) {
	upcoming := [constants.MagicBytes]byte{}
	current := [constants.MagicBytes]byte{}
	previous := [constants.MagicBytes]byte{}
	for i := 0; i < constants.MagicBytes; i++ {
		upcoming[i] = 1
		current[i] = 2
		previous[i] = 3
	}
	return upcoming, current, previous
}

type TestHarness struct {
	handler                             SDK_Handler
	conn                                *net.UDPConn
	from                                net.UDPAddr
	signPublicKey                       [packets.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	signPrivateKey                      [packets.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	analyticsServerInitMessageChannel   chan *messages.AnalyticsServerInitMessage
	analyticsServerUpdateMessageChannel chan *messages.AnalyticsServerUpdateMessage
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

	harness.handler = SDK_Handler{}
	harness.handler.MaxPacketSize = 4096
	harness.handler.ServerBackendAddress = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", backendPort))
	harness.handler.GetMagicValues = getMagicValues

	fmt.Printf("server backend address is %s\n", harness.handler.ServerBackendAddress.String())

	harness.from = core.ParseAddress("127.0.0.1:10000")

	packets.SDK_SignKeypair(harness.signPublicKey[:], harness.signPrivateKey[:])

	harness.handler.ServerBackendPrivateKey = harness.signPrivateKey[:]

	harness.analyticsServerInitMessageChannel = make(chan *messages.AnalyticsServerInitMessage, 1024)
	harness.analyticsServerUpdateMessageChannel = make(chan *messages.AnalyticsServerUpdateMessage, 1024)

	harness.handler.AnalyticsServerInitMessageChannel = harness.analyticsServerInitMessageChannel
	harness.handler.AnalyticsServerUpdateMessageChannel = harness.analyticsServerUpdateMessageChannel

	return &harness
}

func TestPacketTooSmall_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 10)

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_PacketTooSmall])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])
}

func TestUnsupportedPacketType_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	for i := 0; i < 256; i++ {

		packetType := uint8(i)

		if packetType == packets.SDK_SERVER_INIT_REQUEST_PACKET || packetType == packets.SDK_SERVER_UPDATE_REQUEST_PACKET || packetType == packets.SDK_SESSION_UPDATE_REQUEST_PACKET {
			continue
		}

		packetData[0] = packetType

		SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

		assert.True(t, harness.handler.Events[SDK_HandlerEvent_UnsupportedPacketType])

		assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

		assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])
	}
}

func TestBasicPacketFilterFailed_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	packetData[0] = packets.SDK_SERVER_INIT_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_BasicPacketFilterFailed])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])
}

func TestAdvancedPacketFilterFailed_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	packetData[0] = packets.SDK_SERVER_INIT_REQUEST_PACKET
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

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_AdvancedPacketFilterFailed])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])
}

func TestNoRouteMatrix_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	packetData[0] = packets.SDK_SERVER_INIT_REQUEST_PACKET
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

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_NoRouteMatrix])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])
}

func TestNoDatabase_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	packetData[0] = packets.SDK_SERVER_INIT_REQUEST_PACKET
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

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_NoDatabase])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])
}

func TestUnknownBuyer_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 100)

	packetData[0] = packets.SDK_SERVER_INIT_REQUEST_PACKET
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
	harness.handler.Database = database.CreateDatabase()

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_UnknownBuyer])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])
}

func TestSignatureCheckFailed_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 100)
	packetData[0] = packets.SDK_SERVER_INIT_REQUEST_PACKET
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
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [packets.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 16 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// run the packet through the handler, it should fail the signature check

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SignatureCheckFailed])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])
}

// ---------------------------------------------------------------------------------------

// tests for the server init handler

func Test_ServerInitHandler_BuyerNotLive_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_SERVER_INIT_REQUEST_PACKET
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
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [packets.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 16 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	packets.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, it should pass the signature check then fail on buyer not live

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_BuyerNotLive])

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])

	// verify that we get a server init message sent over the channel

	select {
	case _ = <-harness.analyticsServerInitMessageChannel:
	default:
		panic("no server init message found on channel")
	}
}

func Test_ServerInitHandler_BuyerSDKTooOld_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_SERVER_INIT_REQUEST_PACKET
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
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [packets.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 16 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// modify the packet so it has an old SDK version of 1.2.3

	packetData[16] = 1
	packetData[17] = 2
	packetData[18] = 3

	// actually sign the packet, so it passes the signature check

	packets.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see that the SDK is too old

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SDKTooOld])

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])

	// verify that we get a server init message sent over the channel

	select {
	case _ = <-harness.analyticsServerInitMessageChannel:
	default:
		panic("no server init message found on channel")
	}
}

func Test_ServerInitHandler_UnknownDatacenter_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_SERVER_INIT_REQUEST_PACKET
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
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [packets.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 16 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	packets.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see the datacenter is unknown

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_UnknownDatacenter])

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])

	// verify that we get a server init message sent over the channel

	select {
	case _ = <-harness.analyticsServerInitMessageChannel:
	default:
		panic("no server init message found on channel")
	}
}

func Test_ServerInitHandler_ServerInitResponse_SDK(t *testing.T) {

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
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [packets.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// setup "local" datacenter in the database

	localDatacenterId := common.DatacenterId("local")

	localDatacenter := &database.Datacenter{
		Id:        localDatacenterId,
		Name:      "local",
		Latitude:  10,
		Longitude: 20,
	}

	harness.handler.Database.DatacenterMap[localDatacenterId] = localDatacenter

	// construct a valid, signed server init request packet

	requestId := uint64(0x12345)

	packet := packets.SDK_ServerInitRequestPacket{
		Version:        packets.SDKVersion{5, 0, 0},
		BuyerId:        buyerId,
		RequestId:      requestId,
		DatacenterId:   common.DatacenterId("local"),
		DatacenterName: "local",
	}

	packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_SERVER_INIT_REQUEST_PACKET, 1500, &clientAddress, &harness.handler.ServerBackendAddress, buyerPrivateKey[:])
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

			if packetData[0] != packets.SDK_SERVER_INIT_RESPONSE_PACKET {
				core.Debug("wrong packet type")
				continue
			}

			// ignore any packets that are too small

			if len(packetData) < 16+3+4+packets.SDK_CRYPTO_SIGN_BYTES+2 {
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
			toAddressData, toAddressPort := core.GetAddressData(&clientAddress, toAddressBuffer[:])

			if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packetData)) {
				core.Debug("advanced packet filter failed")
				continue
			}

			// make sure packet signature check passes

			if !packets.SDK_CheckPacketSignature(packetData, harness.signPublicKey[:]) {
				core.Debug("packet signature check failed")
				return
			}

			// read packet

			packetData = packetData[16 : len(packetData)-(2+packets.SDK_CRYPTO_SIGN_BYTES)]

			responsePacket := packets.SDK_ServerInitResponsePacket{}
			if err := packets.ReadPacket(packetData, &responsePacket); err != nil {
				core.Debug("could not read server init response packet")
				continue
			}

			// check all response packet fields match expected values

			assert.Equal(t, packet.RequestId, responsePacket.RequestId)
			assert.Equal(t, packets.SDK_ServerInitResponseOK, int(responsePacket.Response))

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

		SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

		assert.False(t, harness.handler.Events[SDK_HandlerEvent_PacketTooSmall])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_UnsupportedPacketType])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_BasicPacketFilterFailed])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_AdvancedPacketFilterFailed])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_NoRouteMatrix])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_NoDatabase])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SignatureCheckFailed])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_BuyerNotLive])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SDKTooOld])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_UnknownDatacenter])

		assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
		assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])

		assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

		assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])

		if i > 10 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// verify that we received a response

	assert.True(t, receivedResponse != 0)

	// verify that we get at least one server init message sent over the channel

	select {
	case message := <-harness.analyticsServerInitMessageChannel:
		assert.Equal(t, message.SDKVersion_Major, byte(5))
		assert.Equal(t, message.SDKVersion_Minor, byte(0))
		assert.Equal(t, message.SDKVersion_Patch, byte(0))
		assert.Equal(t, message.BuyerId, packet.BuyerId)
		assert.Equal(t, message.DatacenterId, packet.DatacenterId)
		assert.Equal(t, message.DatacenterName, packet.DatacenterName)
	default:
		panic("no server init message found on channel")
	}
}

// ---------------------------------------------------------------------------------------

// tests for the server update handler

func Test_ServerUpdateHandler_BuyerNotLive_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_SERVER_UPDATE_REQUEST_PACKET
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
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [packets.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 16 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	packets.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, it should pass the signature check then fail on buyer not live

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_BuyerNotLive])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])

	// verify that we get a server update message sent over the channel

	select {
	case _ = <-harness.analyticsServerUpdateMessageChannel:
	default:
		panic("no server update message found on channel")
	}
}

func Test_ServerUpdateHandler_BuyerSDKTooOld_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_SERVER_UPDATE_REQUEST_PACKET
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
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [packets.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 16 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// modify the packet so it has an old SDK version of 1.2.3

	packetData[16] = 1
	packetData[17] = 2
	packetData[18] = 3

	// actually sign the packet, so it passes the signature check

	packets.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see that the SDK is too old

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SDKTooOld])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])

	// verify that we get a server update message sent over the channel

	select {
	case _ = <-harness.analyticsServerUpdateMessageChannel:
	default:
		panic("no server update message found on channel")
	}
}

func Test_ServerUpdateHandler_UnknownDatacenter_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_SERVER_UPDATE_REQUEST_PACKET
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
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [packets.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 16 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	packets.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see the datacenter is unknown

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
	assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

	assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])

	// verify that we get a server update message sent over the channel

	select {
	case _ = <-harness.analyticsServerUpdateMessageChannel:
	default:
		panic("no server update message found on channel")
	}
}

func Test_ServerUpdateHandler_ServerUpdateResponse_SDK(t *testing.T) {

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
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [packets.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [packets.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	packets.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// setup "local" datacenter in the database

	localDatacenterId := common.DatacenterId("local")

	localDatacenter := &database.Datacenter{
		Id:        localDatacenterId,
		Name:      "local",
		Latitude:  10,
		Longitude: 20,
	}

	harness.handler.Database.DatacenterMap[localDatacenterId] = localDatacenter

	// construct a valid, signed server update request packet

	requestId := uint64(0x12345)

	packet := packets.SDK_ServerUpdateRequestPacket{
		Version:      packets.SDKVersion{5, 0, 0},
		BuyerId:      buyerId,
		RequestId:    requestId,
		DatacenterId: common.DatacenterId("local"),
	}

	packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_SERVER_UPDATE_REQUEST_PACKET, 1500, &clientAddress, &harness.handler.ServerBackendAddress, buyerPrivateKey[:])
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

			if packetData[0] != packets.SDK_SERVER_UPDATE_RESPONSE_PACKET {
				core.Debug("wrong packet type")
				continue
			}

			// ignore any packets that are too small

			if len(packetData) < 16+3+4+packets.SDK_CRYPTO_SIGN_BYTES+2 {
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
			toAddressData, toAddressPort := core.GetAddressData(&clientAddress, toAddressBuffer[:])

			if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packetData)) {
				core.Debug("advanced packet filter failed")
				continue
			}

			// make sure packet signature check passes

			if !packets.SDK_CheckPacketSignature(packetData, harness.signPublicKey[:]) {
				core.Debug("packet signature check failed")
				return
			}

			// read packet

			packetData = packetData[16 : len(packetData)-(2+packets.SDK_CRYPTO_SIGN_BYTES)]

			responsePacket := packets.SDK_ServerUpdateResponsePacket{}
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

		SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

		assert.False(t, harness.handler.Events[SDK_HandlerEvent_PacketTooSmall])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_UnsupportedPacketType])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_BasicPacketFilterFailed])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_AdvancedPacketFilterFailed])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_NoRouteMatrix])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_NoDatabase])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SignatureCheckFailed])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_BuyerNotLive])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SDKTooOld])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_UnknownDatacenter])

		assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
		assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket])

		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
		assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket])

		assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket])
		assert.False(t, harness.handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket])

		assert.False(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])
		assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])

		if i > 10 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// verify that we received a response

	assert.True(t, receivedResponse != 0)

	// verify that we get at least one server update message sent over the channel

	select {
	case message := <-harness.analyticsServerUpdateMessageChannel:
		assert.Equal(t, message.SDKVersion_Major, byte(5))
		assert.Equal(t, message.SDKVersion_Minor, byte(0))
		assert.Equal(t, message.SDKVersion_Patch, byte(0))
		assert.Equal(t, message.BuyerId, packet.BuyerId)
		assert.Equal(t, message.DatacenterId, packet.DatacenterId)
	default:
		panic("no server update message found on channel")
	}
}

// ---------------------------------------------------------------------------------------
