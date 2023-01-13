package main

import (
	"net"
	
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/handlers"
	"github.com/networknext/backend/modules/packets"
)

var service *common.Service

var maxPacketSize int
var serverBackendAddress net.UDPAddr
var serverBackendPublicKey []byte
var serverBackendPrivateKey []byte
var routingPrivateKey []byte

func main() {

	service = common.CreateService("new_server_backend5")

	maxPacketSize = envvar.GetInt("UDP_MAX_PACKET_SIZE", 4096)
	serverBackendAddress = *envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:45000")) // IMPORTANT: This must be the LB public address in dev/prod
	serverBackendPublicKey = envvar.GetBase64("SERVER_BACKEND_PUBLIC_KEY", []byte{})
	serverBackendPrivateKey = envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", []byte{})
	routingPrivateKey = envvar.GetBase64("ROUTING_PRIVATE_KEY", []byte{})

	core.Log("max packet size: %d bytes", maxPacketSize)
	core.Log("server backend address: %s", serverBackendAddress.String())

	if len(serverBackendPublicKey) == 0 {
		panic("SERVER_BACKEND_PUBLIC_KEY must be specified")
	}

	if len(serverBackendPrivateKey) == 0 {
		panic("SERVER_BACKEND_PRIVATE_KEY must be specified")
	}

	if len(routingPrivateKey) == 0 {
		panic("ROUTING_PRIVATE_KEY must be specified")
	}

	service.UpdateRouteMatrix()

	service.SetHealthFunctions(sendTrafficToMe, machineIsHealthy)

	service.StartUDPServer(packetHandler)

	service.UpdateMagic()

	service.StartWebServer()

	service.WaitForShutdown()
}

func sendTrafficToMe() bool {
	routeMatrix, database := service.RouteMatrixAndDatabase()
	return routeMatrix != nil && database != nil
}

func machineIsHealthy() bool {
	return true
}

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	handler := handlers.SDK5_Handler{}

	handler.ServerBackendAddress = serverBackendAddress
	handler.ServerBackendPublicKey = serverBackendPublicKey
	handler.ServerBackendPrivateKey = serverBackendPrivateKey
	handler.RoutingPrivateKey = routingPrivateKey
	handler.RouteMatrix, handler.Database = service.RouteMatrixAndDatabase()
	handler.MaxPacketSize = maxPacketSize
	handler.GetMagicValues = func() ([]byte, []byte, []byte) { return service.GetMagicValues() }
	handler.LocateIP = locateIP

	handlers.SDK5_PacketHandler(&handler, conn, from, packetData)
}

func locateIP(ip net.IP) (packets.SDK5_LocationData, error) {
	// todo: this needs to be hooked up to the proper ip2location when we are not running in local env!!!
	location := packets.SDK5_LocationData{}
	location.Latitude = 43
	location.Longitude = -75
	return location, nil
}
