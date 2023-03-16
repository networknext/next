package main

import (
	"net"
	"fmt"

	"github.com/networknext/backend/modules/common"
)

func main() {

	service := common.CreateService("udp_test")

	service.StartUDPServer(packetHandler)

	service.StartWebServer()

	service.WaitForShutdown()
}

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {
	fmt.Printf("received %d byte udp packet from %s\n", len(packetData), from.String())
}
