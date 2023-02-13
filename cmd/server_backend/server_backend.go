package main

import (
	"net"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/handlers"
	"github.com/networknext/backend/modules/messages"
)

var service *common.Service

var channelSize int
var maxPacketSize int
var serverBackendAddress net.UDPAddr
var serverBackendPublicKey []byte
var serverBackendPrivateKey []byte
var relayBackendPrivateKey []byte

var serverInitMessageChannel chan *messages.ServerInitMessage
var serverUpdateMessageChannel chan *messages.ServerUpdateMessage
var portalMessageChannel chan *messages.PortalMessage
var nearRelayPingsMessageChannel chan *messages.NearRelayPingsMessage
var sessionUpdateMessageChannel chan *messages.SessionUpdateMessage
var matchDataMessageChannel chan *messages.MatchDataMessage

func main() {

	service = common.CreateService("new_server_backend5")

	channelSize = envvar.GetInt("CHANNEL_SIZE", 10*1024)
	maxPacketSize = envvar.GetInt("UDP_MAX_PACKET_SIZE", 4096)
	serverBackendAddress = envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:45000")) // IMPORTANT: This must be the LB public address in dev/prod
	serverBackendPublicKey = envvar.GetBase64("SERVER_BACKEND_PUBLIC_KEY", []byte{})
	serverBackendPrivateKey = envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", []byte{})
	relayBackendPrivateKey = envvar.GetBase64("RELAY_BACKEND_PRIVATE_KEY", []byte{})

	core.Log("channel size: %d", channelSize)
	core.Log("max packet size: %d bytes", maxPacketSize)
	core.Log("server backend address: %s", serverBackendAddress.String())

	if len(serverBackendPublicKey) == 0 {
		panic("SERVER_BACKEND_PUBLIC_KEY must be specified")
	}

	if len(serverBackendPrivateKey) == 0 {
		panic("SERVER_BACKEND_PRIVATE_KEY must be specified")
	}

	if len(relayBackendPrivateKey) == 0 {
		panic("RELAY_BACKEND_PRIVATE_KEY must be specified")
	}

	// initialize message channels

	serverInitMessageChannel = make(chan *messages.ServerInitMessage, channelSize)
	serverUpdateMessageChannel = make(chan *messages.ServerUpdateMessage, channelSize)
	portalMessageChannel = make(chan *messages.PortalMessage, channelSize)
	nearRelayPingsMessageChannel = make(chan *messages.NearRelayPingsMessage, channelSize)
	sessionUpdateMessageChannel = make(chan *messages.SessionUpdateMessage, channelSize)
	matchDataMessageChannel = make(chan *messages.MatchDataMessage, channelSize)

	processServerInitMessages()
	processServerUpdateMessages()
	processPortalMessages()
	processNearRelayPingsMessages()
	processSessionUpdateMessages()
	processMatchDataMessages()

	// start service

	service.UpdateRouteMatrix()

	service.SetHealthFunctions(sendTrafficToMe, machineIsHealthy)

	// todo: not ready yet
	/*
		if !service.Local {
			service.LoadIP2Location()
		}
	*/

	service.UpdateMagic()

	service.StartUDPServer(packetHandler)

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
	handler.RelayBackendPrivateKey = relayBackendPrivateKey
	handler.RouteMatrix, handler.Database = service.RouteMatrixAndDatabase()
	handler.MaxPacketSize = maxPacketSize
	handler.GetMagicValues = func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte) {
		return service.GetMagicValues()
	}

	handler.ServerInitMessageChannel = serverInitMessageChannel
	handler.ServerUpdateMessageChannel = serverUpdateMessageChannel
	handler.PortalMessageChannel = portalMessageChannel
	handler.NearRelayPingsMessageChannel = nearRelayPingsMessageChannel
	handler.SessionUpdateMessageChannel = sessionUpdateMessageChannel
	handler.MatchDataMessageChannel = matchDataMessageChannel

	// todo: not ready yet
	handler.LocateIP = locateIP_Local
	/*
		if service.Local {
			handler.LocateIP = locateIP_Local
		} else {
			handler.LocateIP = locateIP_Real
		}
	*/

	handlers.SDK5_PacketHandler(&handler, conn, from, packetData)
}

func locateIP_Local(ip net.IP) (float32, float32) {
	return 43, -75
}

func locateIP_Real(ip net.IP) (float32, float32) {
	return service.LocateIP(ip)
}

func processServerInitMessages() {
	go func() {
		for {
			message := <-serverInitMessageChannel
			_ = message
			core.Debug("processed server init message")
		}
	}()
}

func processServerUpdateMessages() {
	go func() {
		for {
			message := <-serverUpdateMessageChannel
			_ = message
			core.Debug("processed server update message")
		}
	}()
}

func processPortalMessages() {
	go func() {
		for {
			message := <-portalMessageChannel
			_ = message
			core.Debug("processed portal message")
		}
	}()
}

func processNearRelayPingsMessages() {
	go func() {
		for {
			message := <-nearRelayPingsMessageChannel
			_ = message
			core.Debug("processed near relay pings message")
		}
	}()
}

func processSessionUpdateMessages() {
	go func() {
		for {
			message := <-sessionUpdateMessageChannel
			_ = message
			core.Debug("processed session update message")
		}
	}()
}

func processMatchDataMessages() {
	go func() {
		for {
			message := <-matchDataMessageChannel
			_ = message
			core.Debug("processed match data message")
		}
	}()
}
