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

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	// make sure the basic packet filter passes

	if !core.BasicPacketFilter(packetData[:], len(packetData)) {
		core.Debug("basic packet filter failed for %d byte packet from %s", len(packetData), from.String())
		return
	}

	// make sure the advanced packet filter passes

	to := &serverBackendAddress

	var emptyMagic [8]byte

	var fromAddressBuffer [32]byte
	var toAddressBuffer [32]byte

	fromAddressData, fromAddressPort := core.GetAddressData(from, fromAddressBuffer[:])
	toAddressData, toAddressPort := core.GetAddressData(to, toAddressBuffer[:])

	if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packetData)) {
		core.Debug("advanced packet filter failed for %d byte packet from %s to %s", len(packetData), from.String(), to.String())
		return
	}

	// check packet signature

	// ...

	// process the packet according to type

	packetType := packetData[0]

	packetData = packetData[16:len(packetData)-(2+packets.NEXT_CRYPTO_SIGN_BYTES)]

	switch packetType {

	case packets.SDK5_SERVER_INIT_REQUEST_PACKET:
		packet := packets.SDK5_ServerInitRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server init request packet from %s", from.String())
			return;
		}
		ProcessServerInitRequestPacket(conn, from, &packet)
		break

	case packets.SDK5_SERVER_UPDATE_REQUEST_PACKET:
		packet := packets.SDK5_ServerUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server update request packet from %s", from.String())
			return;
		}
		ProcessServerUpdateRequestPacket(conn, from, &packet)
		break

	case packets.SDK5_SESSION_UPDATE_REQUEST_PACKET:
		packet := packets.SDK5_SessionUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read session update request packet from %s", from.String())
			return;
		}
		ProcessSessionUpdateRequestPacket(conn, from, &packet)
		break

	case packets.SDK5_MATCH_DATA_REQUEST_PACKET:
		packet := packets.SDK5_MatchDataRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read match data request packet from %s", from.String())
			return;
		}
		ProcessMatchDataRequestPacket(conn, from, &packet)
		break

	default:
		core.Debug("received unknown packet type %d from %s", packetType, from.String())
	}
}

func ProcessServerInitRequestPacket(conn *net.UDPConn, from *net.UDPAddr, packet *packets.SDK5_ServerInitRequestPacket) {
	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received server init request packet from %s", from.String())
	core.Debug("version: %d.%d.%d", packet.Version.Major, packet.Version.Minor, packet.Version.Patch)
	core.Debug("buyer id: %016x", packet.BuyerId)
	core.Debug("request id: %016x", packet.RequestId)
	core.Debug("datacenter: \"%s\" [%016x]", packet.DatacenterName, packet.DatacenterId)
	core.Debug("---------------------------------------------------------------------------")
}

func ProcessServerUpdateRequestPacket(conn *net.UDPConn, from *net.UDPAddr, packet *packets.SDK5_ServerUpdateRequestPacket) {
	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received server update request packet from %s", from.String())
	// ...
	core.Debug("---------------------------------------------------------------------------")
}

func ProcessSessionUpdateRequestPacket(conn *net.UDPConn, from *net.UDPAddr, packet *packets.SDK5_SessionUpdateRequestPacket) {
	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received session update request packet from %s", from.String())
	// ...
	core.Debug("---------------------------------------------------------------------------")
}

func ProcessMatchDataRequestPacket(conn *net.UDPConn, from *net.UDPAddr, packet *packets.SDK5_MatchDataRequestPacket) {
	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received match data request packet from %s", from.String())
	// ...
	core.Debug("---------------------------------------------------------------------------")
}
