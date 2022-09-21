package main

import (
	"net"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/packets"
)

var serverBackendAddress net.UDPAddr

func main() {

	service := common.CreateService("new_server_backend5")

	serverBackendAddress = *envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:45000"))

	core.Log("server backend address: %s", serverBackendAddress.String())

	service.StartUDPServer(packetHandler)

	service.StartWebServer()

	service.WaitForShutdown()
}

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packet []byte) {

	// make sure the basic packet filter passes

	if !core.BasicPacketFilter(packet[:], len(packet)) {
		core.Debug("basic packet filter failed for %d byte packet from %s", len(packet), from.String())
		return
	}

	// make sure the advanced packet filter passes

	to := &serverBackendAddress

	var emptyMagic [8]byte

	var fromAddressBuffer [32]byte
	var toAddressBuffer [32]byte

	fromAddressData, fromAddressPort := core.GetAddressData(from, fromAddressBuffer[:])
	toAddressData, toAddressPort := core.GetAddressData(to, toAddressBuffer[:])

	if !core.AdvancedPacketFilter(packet, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packet)) {
		core.Debug("advanced packet filter failed for %d byte packet from %s to %s", len(packet), from.String(), to.String())
		return
	}

	// process the packet according to type

	packetType := packet[0]

	switch packetType {

	case packets.SDK5_SERVER_INIT_REQUEST_PACKET:
		core.Debug("received server init request packet from %s", from.String())
		break

	case packets.SDK5_SERVER_UPDATE_REQUEST_PACKET:
		core.Debug("received server update request packet from %s", from.String())
		break

	case packets.SDK5_SESSION_UPDATE_REQUEST_PACKET:
		core.Debug("received session update request packet from %s", from.String())
		break

	case packets.SDK5_MATCH_DATA_REQUEST_PACKET:
		core.Debug("received match data request packet from %s", from.String())
		break

	default:
		core.Debug("received unknown packet type %d from %s", packetType, from.String())
	}
}
