package main

import (
	"net"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
)

func main() {

	service := common.CreateService("new_server_backend4")

	service.UpdateRouteMatrix()

	service.StartUDPServer(packetHandler)

	service.StartWebServer()

	service.WaitForShutdown()
}

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packet []byte) {
	core.Debug("received %d byte udp packet from %s", len(packet), from.String())
}
