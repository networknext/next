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

var portalSessionUpdateMessageChannel chan *messages.PortalSessionUpdateMessage

var analyticsServerInitMessageChannel chan *messages.AnalyticsServerInitMessage
var analyticsServerUpdateMessageChannel chan *messages.AnalyticsServerUpdateMessage
var analyticsSessionUpdateMessageChannel chan *messages.AnalyticsSessionUpdateMessage
var analyticsSessionSummaryMessageChannel chan *messages.AnalyticsSessionSummaryMessage
var analyticsMatchDataMessageChannel chan *messages.AnalyticsMatchDataMessage
var analyticsNearRelayPingsMessageChannel chan *messages.AnalyticsNearRelayPingsMessage

var enableGooglePubsub bool

func main() {

	service = common.CreateService("new_server_backend5")

	channelSize = envvar.GetInt("CHANNEL_SIZE", 10*1024)
	maxPacketSize = envvar.GetInt("UDP_MAX_PACKET_SIZE", 4096)
	serverBackendAddress = envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:45000")) // IMPORTANT: This must be the LB public address in dev/prod
	serverBackendPublicKey = envvar.GetBase64("SERVER_BACKEND_PUBLIC_KEY", []byte{})
	serverBackendPrivateKey = envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", []byte{})
	relayBackendPrivateKey = envvar.GetBase64("RELAY_BACKEND_PRIVATE_KEY", []byte{})
	enableGooglePubsub = envvar.GetBool("ENABLE_GOOGLE_PUBSUB", false)

	core.Log("channel size: %d", channelSize)
	core.Log("max packet size: %d bytes", maxPacketSize)
	core.Log("server backend address: %s", serverBackendAddress.String())
	core.Log("enable google pubsub: %v", enableGooglePubsub)

	if len(serverBackendPublicKey) == 0 {
		panic("SERVER_BACKEND_PUBLIC_KEY must be specified")
	}

	if len(serverBackendPrivateKey) == 0 {
		panic("SERVER_BACKEND_PRIVATE_KEY must be specified")
	}

	if len(relayBackendPrivateKey) == 0 {
		panic("RELAY_BACKEND_PRIVATE_KEY must be specified")
	}

	// initialize portal message channels

	portalSessionUpdateMessageChannel = make(chan *messages.PortalSessionUpdateMessage, channelSize)

	// initialize analytics message channels

	analyticsServerInitMessageChannel = make(chan *messages.AnalyticsServerInitMessage, channelSize)
	analyticsServerUpdateMessageChannel = make(chan *messages.AnalyticsServerUpdateMessage, channelSize)
	analyticsSessionUpdateMessageChannel = make(chan *messages.AnalyticsSessionUpdateMessage, channelSize)
	analyticsSessionSummaryMessageChannel = make(chan *messages.AnalyticsSessionSummaryMessage, channelSize)
	analyticsMatchDataMessageChannel = make(chan *messages.AnalyticsMatchDataMessage, channelSize)
	analyticsNearRelayPingsMessageChannel = make(chan *messages.AnalyticsNearRelayPingsMessage, channelSize)

	processPortalSessionUpdateMessages()
	// todo: portal match update
	// todo: portal server update

	processAnalyticsServerInitMessages()
	processAnalyticsServerUpdateMessages()
	processAnalyticsNearRelayPingsMessages()
	processAnalyticsSessionUpdateMessages()
	processAnalyticsSessionSummaryMessages()
	processAnalyticsMatchDataMessages()

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

	handler.PortalSessionUpdateMessageChannel = portalSessionUpdateMessageChannel

	handler.AnalyticsServerInitMessageChannel = analyticsServerInitMessageChannel
	handler.AnalyticsServerUpdateMessageChannel = analyticsServerUpdateMessageChannel
	handler.AnalyticsSessionUpdateMessageChannel = analyticsSessionUpdateMessageChannel
	handler.AnalyticsSessionSummaryMessageChannel = analyticsSessionSummaryMessageChannel
	handler.AnalyticsNearRelayPingsMessageChannel = analyticsNearRelayPingsMessageChannel

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

func processPortalSessionUpdateMessages() {
	go func() {
		for {
			message := <-portalSessionUpdateMessageChannel
			_ = message
			core.Debug("processed portal session update message")
		}
	}()
}

func processAnalyticsServerInitMessages() {
	go func() {
		for {
			message := <-analyticsServerInitMessageChannel
			core.Debug("processing analytics server init message")
			messageData := message.Write(make([]byte, message.GetMaxSize()))
			if enableGooglePubsub {
				// googlePubsubProducer.MessageChannel <- messageData
				_ = messageData
			}
		}
	}()
}

func processAnalyticsServerUpdateMessages() {
	go func() {
		for {
			message := <-analyticsServerUpdateMessageChannel
			core.Debug("processing analytics server update message")
			messageData := message.Write(make([]byte, message.GetMaxSize()))
			if enableGooglePubsub {
				// googlePubsubProducer.MessageChannel <- messageData
				_ = messageData
			}
		}
	}()
}

func processAnalyticsNearRelayPingsMessages() {
	go func() {
		for {
			message := <-analyticsNearRelayPingsMessageChannel
			core.Debug("processing analytics near relay pings message")
			messageData := message.Write(make([]byte, message.GetMaxSize()))
			if enableGooglePubsub {
				// googlePubsubProducer.MessageChannel <- messageData
				_ = messageData
			}
		}
	}()
}

func processAnalyticsSessionUpdateMessages() {
	go func() {
		for {
			message := <-analyticsSessionUpdateMessageChannel
			core.Debug("processing analytics session update message")
			messageData := message.Write(make([]byte, message.GetMaxSize()))
			if enableGooglePubsub {
				// googlePubsubProducer.MessageChannel <- messageData
				_ = messageData
			}
		}
	}()
}

func processAnalyticsSessionSummaryMessages() {
	go func() {
		for {
			message := <-analyticsSessionSummaryMessageChannel
			core.Debug("processing analytics session summary message")
			messageData := message.Write(make([]byte, message.GetMaxSize()))
			if enableGooglePubsub {
				// googlePubsubProducer.MessageChannel <- messageData
				_ = messageData
			}
		}
	}()
}

func processAnalyticsMatchDataMessages() {
	go func() {
		for {
			message := <-analyticsMatchDataMessageChannel
			core.Debug("processing analytics match data message")
			messageData := message.Write(make([]byte, message.GetMaxSize()))
			if enableGooglePubsub {
				// googlePubsubProducer.MessageChannel <- messageData
				_ = messageData
			}
		}
	}()
}
