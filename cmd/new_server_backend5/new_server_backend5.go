package main

import (
	"net"
	"net/http"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/handlers"
)

var service *common.Service

var maxPacketSize int
var serverBackendAddress net.UDPAddr
var serverBackendPrivateKey []byte

func main() {

	service = common.CreateService("new_server_backend5")

	maxPacketSize = envvar.GetInt("UDP_MAX_PACKET_SIZE", 4096)
	serverBackendAddress = *envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:45000"))
	serverBackendPrivateKey = envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", []byte{})

	core.Log("max packet size: %d bytes", maxPacketSize)
	core.Log("server backend address: %s", serverBackendAddress.String())

	service.UpdateRouteMatrix()

	service.OverrideHealthHandler(healthHandler)

	service.StartUDPServer(packetHandler)

	service.UpdateMagic()

	service.StartWebServer()

	service.WaitForShutdown()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {

	routeMatrix, database := service.RouteMatrixAndDatabase()

	not_ready := routeMatrix == nil || database == nil

	if not_ready {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	handler := handlers.SDK5_Handler{}
	handler.RouteMatrix, handler.Database = service.RouteMatrixAndDatabase()
	handler.MaxPacketSize = maxPacketSize
	handler.ServerBackendAddress = serverBackendAddress
	handler.PrivateKey = serverBackendPrivateKey
	handler.GetMagicValues = func() ([]byte, []byte, []byte) {
		return service.GetMagicValues()
	}

	handlers.SDK5_PacketHandler(&handler, conn, from, packetData)
}
