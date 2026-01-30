package handlers

import (
	"context"
	"fmt"
	"net"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
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
	signPublicKey                       [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	signPrivateKey                      [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
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
	harness.handler.MaxPacketSize = 1384
	harness.handler.ServerBackendAddress = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", backendPort))
	harness.handler.GetMagicValues = getMagicValues

	fmt.Printf("server backend address is %s\n", harness.handler.ServerBackendAddress.String())

	harness.from = core.ParseAddress("127.0.0.1:10000")

	crypto.SDK_SignKeypair(harness.signPublicKey[:], harness.signPrivateKey[:])

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
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_AdvancedPacketFilterFailed])
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
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_NoRouteMatrix])
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
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	harness.handler.RouteMatrix = &common.RouteMatrix{}

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_NoDatabase])
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
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_UnknownBuyer])
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
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// run the packet through the handler, it should fail the signature check

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SignatureCheckFailed])
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
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, it should pass the signature check then fail on buyer not live

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_BuyerNotLive])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])

	// verify that we get a server init message sent over the channel

	select {
	case _ = <-harness.analyticsServerInitMessageChannel:
	default:
		panic("no server init message found on channel")
	}
}

func Test_ServerInitHandler_SDKTooOld_SDK(t *testing.T) {

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
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// modify the packet so it has an old SDK version of 0.1.2

	packetData[18] = 0
	packetData[19] = 1
	packetData[20] = 2

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see that the SDK is too old

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SDKTooOld])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])

	// verify that we get a server init message sent over the channel

	select {
	case _ = <-harness.analyticsServerInitMessageChannel:
	default:
		panic("no server init message found on channel")
	}
}

func Test_ServerInitHandler_DatacenterNotEnabled_SDK(t *testing.T) {

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
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see the datacenter is not enabled for the buyer

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_DatacenterNotEnabled])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])

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

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

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

	// make sure the "local" datacenter is enabled for the buyer

	harness.handler.Database.BuyerDatacenterSettings[buyerId] = make(map[uint64]*database.BuyerDatacenterSettings)

	harness.handler.Database.BuyerDatacenterSettings[buyerId][localDatacenterId] = &database.BuyerDatacenterSettings{BuyerId: buyerId, DatacenterId: localDatacenterId, EnableAcceleration: true}

	// construct a valid, signed server init request packet

	requestId := uint64(0x12345)

	packet := packets.SDK_ServerInitRequestPacket{
		Version:        packets.SDKVersion{1, 0, 0},
		BuyerId:        buyerId,
		RequestId:      requestId,
		DatacenterId:   localDatacenterId,
		DatacenterName: "local",
	}

	packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_SERVER_INIT_REQUEST_PACKET, constants.MaxPacketBytes, &clientAddress, &harness.handler.ServerBackendAddress, buyerPrivateKey[:])
	if err != nil {
		core.Error("failed to write server init request packet: %v", err)
		return
	}

	// setup a goroutine to listen for response packets from the packet handler

	var receivedResponse uint64

	go func() {

		for {

			var buffer [constants.MaxPacketBytes]byte

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

			if len(packetData) < 18+3+4+crypto.SDK_CRYPTO_SIGN_BYTES+2 {
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

			fromAddressData := core.GetAddressData(&harness.handler.ServerBackendAddress)
			toAddressData := core.GetAddressData(&clientAddress)

			if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, toAddressData, len(packetData)) {
				core.Debug("advanced packet filter failed")
				continue
			}

			// make sure packet signature check passes

			if !crypto.SDK_CheckPacketSignature(packetData, harness.signPublicKey[:]) {
				core.Debug("packet signature check failed")
				return
			}

			// read packet

			packetData = packetData[18 : len(packetData)-(crypto.SDK_CRYPTO_SIGN_BYTES)]

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

	iterations := 100

	for i := 0; i < iterations; i++ {

		response := atomic.LoadUint64(&receivedResponse)
		if response != 0 {
			break
		}

		SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

		assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket])
		assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket])
		assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage])

		if i > 10 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// verify that we received a response

	assert.True(t, receivedResponse != 0)

	// verify that we get at least one server init message sent over the channel

	select {
	case message := <-harness.analyticsServerInitMessageChannel:
		assert.Equal(t, message.SDKVersion_Major, int32(1))
		assert.Equal(t, message.SDKVersion_Minor, int32(0))
		assert.Equal(t, message.SDKVersion_Patch, int32(0))
		assert.Equal(t, message.BuyerId, int64(packet.BuyerId))
		assert.Equal(t, message.DatacenterId, int64(packet.DatacenterId))
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
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, it should pass the signature check then fail on buyer not live

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_BuyerNotLive])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage])

	// verify that we get a server update message sent over the channel

	select {
	case _ = <-harness.analyticsServerUpdateMessageChannel:
	default:
		panic("no server update message found on channel")
	}
}

func Test_ServerUpdateHandler_SDKTooOld_SDK(t *testing.T) {

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
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// modify the packet so it has an old SDK version of 0.1.2

	packetData[18] = 0
	packetData[19] = 1
	packetData[20] = 2

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see that the SDK is too old

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SDKTooOld])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
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
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see the datacenter is unknown

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_UnknownDatacenter])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
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

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

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
		Version:      packets.SDKVersion{1, 0, 0},
		BuyerId:      buyerId,
		RequestId:    requestId,
		DatacenterId: common.DatacenterId("local"),
	}

	packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_SERVER_UPDATE_REQUEST_PACKET, constants.MaxPacketBytes, &clientAddress, &harness.handler.ServerBackendAddress, buyerPrivateKey[:])
	if err != nil {
		core.Error("failed to write server update request packet: %v", err)
		return
	}

	// setup a goroutine to listen for response packets from the packet handler

	var receivedResponse uint64

	go func() {

		for {

			var buffer [constants.MaxPacketBytes]byte

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

			if len(packetData) < 18+3+4+crypto.SDK_CRYPTO_SIGN_BYTES+2 {
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

			fromAddressData := core.GetAddressData(&harness.handler.ServerBackendAddress)
			toAddressData := core.GetAddressData(&clientAddress)

			if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, toAddressData, len(packetData)) {
				core.Debug("advanced packet filter failed")
				continue
			}

			// make sure packet signature check passes

			if !crypto.SDK_CheckPacketSignature(packetData, harness.signPublicKey[:]) {
				core.Debug("packet signature check failed")
				return
			}

			// read packet

			packetData = packetData[18 : len(packetData)-(crypto.SDK_CRYPTO_SIGN_BYTES)]

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

		assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket])
		assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket])
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
		assert.Equal(t, message.SDKVersion_Major, int32(1))
		assert.Equal(t, message.SDKVersion_Minor, int32(0))
		assert.Equal(t, message.SDKVersion_Patch, int32(0))
		assert.Equal(t, message.BuyerId, int64(packet.BuyerId))
		assert.Equal(t, message.DatacenterId, int64(packet.DatacenterId))
	default:
		panic("no server update message found on channel")
	}
}

// ---------------------------------------------------------------------------------------

// tests for the client relay request handler

func Test_ClientRelayRequestHandler_BuyerNotLive_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_CLIENT_RELAY_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// generate pittle and chonkle so the packet gets through the basic and advanced packet filters

	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, it should pass the signature check then fail on buyer not live

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_BuyerNotLive])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessClientRelayRequestPacket])
}

func Test_ClientRelayRequestHandler_SDKTooOld_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_CLIENT_RELAY_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// generate pittle and chonkle so the packet gets through the basic and advanced packet filters

	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// modify the packet so it has an old SDK version of 0.1.2

	packetData[18] = 0
	packetData[19] = 1
	packetData[20] = 2

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see that the SDK is too old

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SDKTooOld])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessClientRelayRequestPacket])
}

func Test_ClientRelayRequestHandler_UnknownDatacenter_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_CLIENT_RELAY_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// generate pittle and chonkle so the packet gets through the basic and advanced packet filters

	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see the datacenter is unknown

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_UnknownDatacenter])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessClientRelayRequestPacket])
}

func locateIP_Local(ip net.IP) (float32, float32) {
	return 41, -93 // iowa
}

func generateRouteMatrix(relayIds []uint64, costMatrix []uint8, relayDatacenters []uint64) *common.RouteMatrix {

	numRelays := len(relayIds)

	numSegments := numRelays

	relayPrice := make([]uint8, numRelays)

	routeEntries := core.Optimize(numRelays, numSegments, costMatrix, relayPrice, relayDatacenters[:])

	destRelays := make([]bool, numRelays)

	relayIdToIndex := make(map[uint64]int32)
	for i := range relayIds {
		relayIdToIndex[relayIds[i]] = int32(i)
	}

	relayAddresses := make([]net.UDPAddr, numRelays)
	for i := range relayAddresses {
		relayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 40000+i))
	}

	relayNames := make([]string, numRelays)
	for i := range relayNames {
		relayNames[i] = fmt.Sprintf("relay-%d", i)
	}

	relayLatitudes := make([]float32, numRelays)
	relayLongitudes := make([]float32, numRelays)
	for i := range relayLatitudes {
		relayLatitudes[i] = 41
		relayLongitudes[i] = -93
	}

	routeMatrix := &common.RouteMatrix{}

	routeMatrix.CreatedAt = uint64(time.Now().Unix())
	routeMatrix.Version = common.RouteMatrixVersion_Write
	routeMatrix.RelayDatacenterIds = relayDatacenters
	routeMatrix.DestRelays = destRelays
	routeMatrix.RelayIds = relayIds
	routeMatrix.RelayAddresses = relayAddresses
	routeMatrix.RelayLatitudes = relayLatitudes
	routeMatrix.RelayLongitudes = relayLongitudes
	routeMatrix.RelayIdToIndex = relayIdToIndex
	routeMatrix.RelayNames = relayNames
	routeMatrix.RouteEntries = routeEntries

	return routeMatrix
}

func Test_ClientRelayRequestResponse_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	harness.handler.LocateIP = locateIP_Local

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

	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// initialize database with some relays

	seller_a := &database.Seller{Id: 1, Name: "a"}
	seller_b := &database.Seller{Id: 2, Name: "b"}
	seller_c := &database.Seller{Id: 3, Name: "c"}

	datacenter_a := &database.Datacenter{Id: 1, Name: "a", Latitude: 41, Longitude: -93}
	datacenter_b := &database.Datacenter{Id: 2, Name: "b", Latitude: 41, Longitude: -93}
	datacenter_c := &database.Datacenter{Id: 3, Name: "c", Latitude: 41, Longitude: -93}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	harness.handler.Database.Relays = make([]database.Relay, 3)

	harness.handler.Database.Relays[0] = database.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	harness.handler.Database.Relays[1] = database.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	harness.handler.Database.Relays[2] = database.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	harness.handler.Database.SellerMap[1] = seller_a
	harness.handler.Database.SellerMap[2] = seller_b
	harness.handler.Database.SellerMap[3] = seller_c

	harness.handler.Database.DatacenterMap[1] = datacenter_a
	harness.handler.Database.DatacenterMap[2] = datacenter_b
	harness.handler.Database.DatacenterMap[3] = datacenter_c

	harness.handler.Database.RelayMap[1] = &harness.handler.Database.Relays[0]
	harness.handler.Database.RelayMap[2] = &harness.handler.Database.Relays[1]
	harness.handler.Database.RelayMap[3] = &harness.handler.Database.Relays[2]

	// setup "local" datacenter in the database

	localDatacenterId := common.DatacenterId("local")

	localDatacenter := &database.Datacenter{
		Id:        localDatacenterId,
		Name:      "local",
		Latitude:  10,
		Longitude: 20,
	}

	harness.handler.Database.DatacenterMap[localDatacenterId] = localDatacenter

	// generate a route matrix with the relays

	relayIds := []uint64{1, 2, 3}

	costMatrix := make([]uint8, 3)

	for i := range costMatrix {
		costMatrix[i] = 255
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	relayDatacenters := [...]uint64{1, 2, 3}

	harness.handler.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:])

	// construct a valid, signed server update request packet

	requestId := uint64(0x12345)

	packet := packets.SDK_ClientRelayRequestPacket{
		Version:       packets.SDKVersion{1, 0, 0},
		BuyerId:       buyerId,
		RequestId:     requestId,
		DatacenterId:  common.DatacenterId("local"),
		ClientAddress: clientAddress,
	}

	packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_CLIENT_RELAY_REQUEST_PACKET, constants.MaxPacketBytes, &clientAddress, &harness.handler.ServerBackendAddress, buyerPrivateKey[:])
	if err != nil {
		core.Error("failed to write client update request packet: %v", err)
		return
	}

	// setup a goroutine to listen for response packets from the packet handler

	var receivedResponse uint64

	go func() {

		for {

			var buffer [constants.MaxPacketBytes]byte

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

			// ignore any packets that are not client relay response packets

			if packetData[0] != packets.SDK_CLIENT_RELAY_RESPONSE_PACKET {
				core.Debug("wrong packet type")
				continue
			}

			// ignore any packets that are too small

			if len(packetData) < 18+3+4+crypto.SDK_CRYPTO_SIGN_BYTES+2 {
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

			fromAddressData := core.GetAddressData(&harness.handler.ServerBackendAddress)
			toAddressData := core.GetAddressData(&clientAddress)

			if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, toAddressData, len(packetData)) {
				core.Debug("advanced packet filter failed")
				continue
			}

			// make sure packet signature check passes

			if !crypto.SDK_CheckPacketSignature(packetData, harness.signPublicKey[:]) {
				core.Debug("packet signature check failed")
				return
			}

			// read packet

			packetData = packetData[18 : len(packetData)-(crypto.SDK_CRYPTO_SIGN_BYTES)]

			responsePacket := packets.SDK_ClientRelayResponsePacket{}
			if err := packets.ReadPacket(packetData, &responsePacket); err != nil {
				core.Debug("could not read client relay response packet")
				continue
			}

			// check all response packet fields match expected values

			sort.Slice(responsePacket.ClientRelayIds[:responsePacket.NumClientRelays], func(i, j int) bool { return responsePacket.ClientRelayIds[i] < responsePacket.ClientRelayIds[j] })
			sort.Slice(responsePacket.ClientRelayAddresses[:responsePacket.NumClientRelays], func(i, j int) bool {
				return responsePacket.ClientRelayAddresses[i].String() < responsePacket.ClientRelayAddresses[j].String()
			})

			fmt.Printf("%v\n", responsePacket.ClientRelayIds[:])

			assert.Equal(t, responsePacket.RequestId, packet.RequestId)
			assert.Equal(t, responsePacket.NumClientRelays, int32(3))
			assert.Equal(t, responsePacket.Latitude, float32(41))
			assert.Equal(t, responsePacket.Longitude, float32(-93))
			assert.Equal(t, responsePacket.ClientRelayIds[0], uint64(1))
			assert.Equal(t, responsePacket.ClientRelayIds[1], uint64(2))
			assert.Equal(t, responsePacket.ClientRelayIds[2], uint64(3))
			assert.Equal(t, responsePacket.ClientRelayAddresses[0].String(), "127.0.0.1:40000")
			assert.Equal(t, responsePacket.ClientRelayAddresses[1].String(), "127.0.0.1:40001")
			assert.Equal(t, responsePacket.ClientRelayAddresses[2].String(), "127.0.0.1:40002")

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

		assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessClientRelayRequestPacket])
		assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentClientRelayResponsePacket])

		if i > 10 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// verify that we received a response

	assert.True(t, receivedResponse != 0)
}

// ---------------------------------------------------------------------------------------

// tests for the server relay request handler

func Test_ServerRelayRequestHandler_BuyerNotLive_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_SERVER_RELAY_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// generate pittle and chonkle so the packet gets through the basic and advanced packet filters

	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, it should pass the signature check then fail on buyer not live

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_BuyerNotLive])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerRelayRequestPacket])
}

func Test_ServerRelayRequestHandler_SDKTooOld_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_SERVER_RELAY_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// generate pittle and chonkle so the packet gets through the basic and advanced packet filters

	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// modify the packet so it has an old SDK version of 0.1.2

	packetData[18] = 0
	packetData[19] = 1
	packetData[20] = 2

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see that the SDK is too old

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_SDKTooOld])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerRelayRequestPacket])
}

func Test_ServerRelayRequestHandler_UnknownDatacenter_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	// setup a dummy packet that will get through the packet type check

	packetData := make([]byte, 256)
	packetData[0] = packets.SDK_SERVER_RELAY_REQUEST_PACKET
	for i := 1; i < len(packetData); i++ {
		packetData[i] = byte(i)
	}

	// generate pittle and chonkle so the packet gets through the basic and advanced packet filters

	magic := [8]byte{}
	fromAddress := [4]byte{127, 0, 0, 1}
	toAddress := [4]byte{127, 0, 0, 1}
	packetLength := len(packetData)

	core.GeneratePittle(packetData[1:3], fromAddress[:], toAddress[:], packetLength)

	core.GenerateChonkle(packetData[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

	// setup a buyer in the database with keypair

	harness.handler.RouteMatrix = &common.RouteMatrix{}
	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]
	_ = buyerPrivateKey

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// modify the packet so it has the buyer id of the new buyer, so it passes the unknown buyer check

	index := 18 + 3
	encoding.WriteUint64(packetData[:], &index, buyerId)

	// actually sign the packet, so it passes the signature check

	crypto.SDK_SignPacket(packetData[:], buyerPrivateKey[:])

	// run the packet through the handler, we should see the datacenter is unknown

	SDK_PacketHandler(&harness.handler, harness.conn, &harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK_HandlerEvent_UnknownDatacenter])
	assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerRelayRequestPacket])
}

func Test_ServerRelayRequestResponse_SDK(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	harness.handler.LocateIP = locateIP_Local

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

	harness.handler.Database = database.CreateDatabase()

	buyerId := uint64(0x1111111122222222)

	var buyerPublicKey [crypto.SDK_CRYPTO_SIGN_PUBLIC_KEY_BYTES]byte
	var buyerPrivateKey [crypto.SDK_CRYPTO_SIGN_PRIVATE_KEY_BYTES]byte
	crypto.SDK_SignKeypair(buyerPublicKey[:], buyerPrivateKey[:])

	buyer := &database.Buyer{}
	buyer.Live = true
	buyer.PublicKey = buyerPublicKey[:]

	harness.handler.Database.BuyerMap[buyerId] = buyer

	// initialize database with some relays

	seller_a := &database.Seller{Id: 1, Name: "a"}
	seller_b := &database.Seller{Id: 2, Name: "b"}
	seller_c := &database.Seller{Id: 3, Name: "c"}

	datacenter_a := &database.Datacenter{Id: 1, Name: "a", Latitude: 41, Longitude: -93}
	datacenter_b := &database.Datacenter{Id: 2, Name: "b", Latitude: 41, Longitude: -93}
	datacenter_c := &database.Datacenter{Id: 3, Name: "c", Latitude: 41, Longitude: -93}

	relay_address_a := core.ParseAddress("127.0.0.1:40000")
	relay_address_b := core.ParseAddress("127.0.0.1:40001")
	relay_address_c := core.ParseAddress("127.0.0.1:40002")

	relay_public_key_a, _ := crypto.Box_KeyPair()
	relay_public_key_b, _ := crypto.Box_KeyPair()
	relay_public_key_c, _ := crypto.Box_KeyPair()

	harness.handler.Database.Relays = make([]database.Relay, 3)

	harness.handler.Database.Relays[0] = database.Relay{Id: 1, Name: "a", PublicAddress: relay_address_a, Seller: seller_a, PublicKey: relay_public_key_a}
	harness.handler.Database.Relays[1] = database.Relay{Id: 2, Name: "b", PublicAddress: relay_address_b, Seller: seller_b, PublicKey: relay_public_key_b}
	harness.handler.Database.Relays[2] = database.Relay{Id: 3, Name: "c", PublicAddress: relay_address_c, Seller: seller_c, PublicKey: relay_public_key_c}

	harness.handler.Database.SellerMap[1] = seller_a
	harness.handler.Database.SellerMap[2] = seller_b
	harness.handler.Database.SellerMap[3] = seller_c

	harness.handler.Database.DatacenterMap[1] = datacenter_a
	harness.handler.Database.DatacenterMap[2] = datacenter_b
	harness.handler.Database.DatacenterMap[3] = datacenter_c

	harness.handler.Database.RelayMap[1] = &harness.handler.Database.Relays[0]
	harness.handler.Database.RelayMap[2] = &harness.handler.Database.Relays[1]
	harness.handler.Database.RelayMap[3] = &harness.handler.Database.Relays[2]

	harness.handler.Database.DatacenterRelays = make(map[uint64][]uint64)
	harness.handler.Database.DatacenterRelays[1] = []uint64{1}

	// setup "local" datacenter in the database

	localDatacenterId := common.DatacenterId("local")

	localDatacenter := &database.Datacenter{
		Id:        localDatacenterId,
		Name:      "local",
		Latitude:  10,
		Longitude: 20,
	}

	harness.handler.Database.DatacenterMap[localDatacenterId] = localDatacenter

	// generate a route matrix with the relays

	relayIds := []uint64{1, 2, 3}

	costMatrix := make([]uint8, 3)

	for i := range costMatrix {
		costMatrix[i] = 255
	}

	costMatrix[core.TriMatrixIndex(0, 1)] = 10
	costMatrix[core.TriMatrixIndex(1, 2)] = 10
	costMatrix[core.TriMatrixIndex(0, 2)] = 100

	relayDatacenters := [...]uint64{1, 2, 3}

	harness.handler.RouteMatrix = generateRouteMatrix(relayIds[:], costMatrix, relayDatacenters[:])

	// construct a valid, signed server update request packet

	requestId := uint64(0x12345)

	packet := packets.SDK_ServerRelayRequestPacket{
		Version:      packets.SDKVersion{1, 0, 0},
		BuyerId:      buyerId,
		RequestId:    requestId,
		DatacenterId: 1,
	}

	packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_SERVER_RELAY_REQUEST_PACKET, constants.MaxPacketBytes, &clientAddress, &harness.handler.ServerBackendAddress, buyerPrivateKey[:])
	if err != nil {
		core.Error("failed to write client update request packet: %v", err)
		return
	}

	// setup a goroutine to listen for response packets from the packet handler

	var receivedResponse uint64

	go func() {

		for {

			var buffer [constants.MaxPacketBytes]byte

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

			// ignore any packets that are not client relay response packets

			if packetData[0] != packets.SDK_SERVER_RELAY_RESPONSE_PACKET {
				core.Debug("wrong packet type")
				continue
			}

			// ignore any packets that are too small

			if len(packetData) < 18+3+4+crypto.SDK_CRYPTO_SIGN_BYTES+2 {
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

			fromAddressData := core.GetAddressData(&harness.handler.ServerBackendAddress)
			toAddressData := core.GetAddressData(&clientAddress)

			if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, toAddressData, len(packetData)) {
				core.Debug("advanced packet filter failed")
				continue
			}

			// make sure packet signature check passes

			if !crypto.SDK_CheckPacketSignature(packetData, harness.signPublicKey[:]) {
				core.Debug("packet signature check failed")
				return
			}

			// read packet

			packetData = packetData[18 : len(packetData)-(crypto.SDK_CRYPTO_SIGN_BYTES)]

			responsePacket := packets.SDK_ServerRelayResponsePacket{}
			if err := packets.ReadPacket(packetData, &responsePacket); err != nil {
				core.Debug("could not read client relay response packet")
				continue
			}

			// check all response packet fields match expected values

			assert.Equal(t, responsePacket.RequestId, packet.RequestId)
			assert.Equal(t, responsePacket.NumServerRelays, int32(1))
			assert.Equal(t, responsePacket.ServerRelayIds[0], uint64(1))
			assert.Equal(t, responsePacket.ServerRelayAddresses[0].String(), "127.0.0.1:40000")

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

		assert.True(t, harness.handler.Events[SDK_HandlerEvent_ProcessServerRelayRequestPacket])
		assert.True(t, harness.handler.Events[SDK_HandlerEvent_SentServerRelayResponsePacket])

		if i > 10 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// verify that we received a response

	assert.True(t, receivedResponse != 0)
}

// ---------------------------------------------------------------------------------------
