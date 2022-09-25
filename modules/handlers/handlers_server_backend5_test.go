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
)

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
}

func CreateTestHarness() *TestHarness {

	harness := TestHarness{}

	ctx := context.Background()

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(ctx, "udp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	// todo: generate sign keypair

	harness.handler = SDK5_Handler{}
	harness.handler.MaxPacketSize = 4096
	harness.handler.ServerBackendAddress = *core.ParseAddress("127.0.0.1:45000")		// todo: get port from socket above
	// todo: harness.handler.Database
	// todo: harness.handler.RouteMatrix
	// todo: harness.handler.PrivateKey
	harness.handler.GetMagicValues = getMagicValues

	// todo: create client conn

	harness.conn = lp.(*net.UDPConn)
	harness.from = core.ParseAddress("127.0.0.1:10000")								// todo: get port from client socket

	return &harness
}

func TestPacketTooSmall(t *testing.T) {

	t.Parallel()

	harness := CreateTestHarness()

	packetData := make([]byte, 10)

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)

	assert.True(t, harness.handler.Events[SDK5_HandlerEvent_PacketTooSmall])
}

func TestUnsupportedPacketType(t *testing.T) {

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

func TestBasicPacketFilterFailed(t *testing.T) {

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

func TestAdvancedPacketFilterFailed(t *testing.T) {

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

func TestNoRouteMatrix(t *testing.T) {

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

func TestNoDatabase(t *testing.T) {

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

// ...

/*

	SDK5_PacketHandler(&harness.handler, harness.conn, harness.from, packetData)
*/