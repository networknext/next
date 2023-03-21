package main

import (
	"net"
	"os"
	"strings"

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
var portalServerUpdateMessageChannel chan *messages.PortalServerUpdateMessage
var portalNearRelayUpdateMessageChannel chan *messages.PortalNearRelayUpdateMessage
var portalMapUpdateMessageChannel chan *messages.PortalMapUpdateMessage

var analyticsServerInitMessageChannel chan *messages.AnalyticsServerInitMessage
var analyticsServerUpdateMessageChannel chan *messages.AnalyticsServerUpdateMessage
var analyticsSessionUpdateMessageChannel chan *messages.AnalyticsSessionUpdateMessage
var analyticsSessionSummaryMessageChannel chan *messages.AnalyticsSessionSummaryMessage
var analyticsMatchDataMessageChannel chan *messages.AnalyticsMatchDataMessage
var analyticsNearRelayUpdateMessageChannel chan *messages.AnalyticsNearRelayUpdateMessage

var enableGooglePubsub bool
var enableRedisStreams bool

var redisHostname string
var redisPassword string

func main() {

	service = common.CreateService("server_backend")

	channelSize = envvar.GetInt("CHANNEL_SIZE", 10*1024)
	maxPacketSize = envvar.GetInt("UDP_MAX_PACKET_SIZE", 4096)
	serverBackendAddress = envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:45000"))
	serverBackendPublicKey = envvar.GetBase64("SERVER_BACKEND_PUBLIC_KEY", []byte{})
	serverBackendPrivateKey = envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", []byte{})
	relayBackendPrivateKey = envvar.GetBase64("RELAY_BACKEND_PRIVATE_KEY", []byte{})
	enableGooglePubsub = envvar.GetBool("ENABLE_GOOGLE_PUBSUB", false)
	enableRedisStreams = envvar.GetBool("ENABLE_REDIS_STREAMS", true)
	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")

	core.Log("channel size: %d", channelSize)
	core.Log("max packet size: %d bytes", maxPacketSize)
	core.Log("server backend address: %s", serverBackendAddress.String())
	core.Log("enable google pubsub: %v", enableGooglePubsub)
	core.Log("enable redis streams: %v", enableRedisStreams)
	core.Log("redis hostname: %s", redisHostname)

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
	portalServerUpdateMessageChannel = make(chan *messages.PortalServerUpdateMessage, channelSize)
	portalNearRelayUpdateMessageChannel = make(chan *messages.PortalNearRelayUpdateMessage, channelSize)
	portalMapUpdateMessageChannel = make(chan *messages.PortalMapUpdateMessage, channelSize)

	processPortalMessages_RedisStreams[*messages.PortalSessionUpdateMessage](service, "session update", portalSessionUpdateMessageChannel)
	processPortalMessages_RedisStreams[*messages.PortalServerUpdateMessage](service, "server update", portalServerUpdateMessageChannel)
	processPortalMessages_RedisStreams[*messages.PortalNearRelayUpdateMessage](service, "near relay update", portalNearRelayUpdateMessageChannel)
	processPortalMessages_RedisPubsub[*messages.PortalMapUpdateMessage](service, "map update", portalMapUpdateMessageChannel)

	// initialize analytics message channels

	analyticsServerInitMessageChannel = make(chan *messages.AnalyticsServerInitMessage, channelSize)
	analyticsServerUpdateMessageChannel = make(chan *messages.AnalyticsServerUpdateMessage, channelSize)
	analyticsSessionUpdateMessageChannel = make(chan *messages.AnalyticsSessionUpdateMessage, channelSize)
	analyticsSessionSummaryMessageChannel = make(chan *messages.AnalyticsSessionSummaryMessage, channelSize)
	analyticsMatchDataMessageChannel = make(chan *messages.AnalyticsMatchDataMessage, channelSize)
	analyticsNearRelayUpdateMessageChannel = make(chan *messages.AnalyticsNearRelayUpdateMessage, channelSize)

	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsServerInitMessage]("server init", analyticsServerInitMessageChannel)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsServerUpdateMessage]("server update", analyticsServerUpdateMessageChannel)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsNearRelayUpdateMessage]("near relay update", analyticsNearRelayUpdateMessageChannel)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsSessionUpdateMessage]("session update", analyticsSessionUpdateMessageChannel)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsSessionSummaryMessage]("session summary", analyticsSessionSummaryMessageChannel)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsMatchDataMessage]("match data", analyticsMatchDataMessageChannel)

	// start the service

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
	handler.PortalServerUpdateMessageChannel = portalServerUpdateMessageChannel
	handler.PortalNearRelayUpdateMessageChannel = portalNearRelayUpdateMessageChannel
	handler.PortalMapUpdateMessageChannel = portalMapUpdateMessageChannel

	handler.AnalyticsServerInitMessageChannel = analyticsServerInitMessageChannel
	handler.AnalyticsServerUpdateMessageChannel = analyticsServerUpdateMessageChannel
	handler.AnalyticsSessionUpdateMessageChannel = analyticsSessionUpdateMessageChannel
	handler.AnalyticsSessionSummaryMessageChannel = analyticsSessionSummaryMessageChannel
	handler.AnalyticsNearRelayUpdateMessageChannel = analyticsNearRelayUpdateMessageChannel

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

func processPortalMessages_RedisStreams[T messages.Message](service *common.Service, name string, inputChannel chan T) {

	streamName := strings.ReplaceAll(name, " ", "_")

	redisStreamsProducer, err := common.CreateRedisStreamsProducer(service.Context, common.RedisStreamsConfig{
		RedisHostname: redisHostname,
		RedisPassword: redisPassword,
		StreamName:    streamName,
	})

	if err != nil {
		core.Error("could not create redis streams producer for %s", name)
		os.Exit(1)
	}

	go func() {
		for {
			message := <-inputChannel
			core.Debug("processing portal %s message", name)
			messageData := message.Write(make([]byte, message.GetMaxSize()))
			if enableRedisStreams {
				core.Debug("sent portal %s message to redis streams", name)
				redisStreamsProducer.MessageChannel <- messageData
			}
		}
	}()
}

func processPortalMessages_RedisPubsub[T messages.Message](service *common.Service, name string, inputChannel chan T) {

	channelName := strings.ReplaceAll(name, " ", "_")

	redisPubsubProducer, err := common.CreateRedisPubsubProducer(service.Context, common.RedisPubsubConfig{
		RedisHostname:     redisHostname,
		RedisPassword:     redisPassword,
		PubsubChannelName: channelName,
	})

	if err != nil {
		core.Error("could not create redis pubsub producer for %s", name)
		os.Exit(1)
	}

	go func() {
		for {
			message := <-inputChannel
			core.Debug("processing portal %s message", name)
			messageData := message.Write(make([]byte, message.GetMaxSize()))
			if enableRedisStreams {
				core.Debug("sent portal %s message to redis pubsub", name)
				redisPubsubProducer.MessageChannel <- messageData
			}
		}
	}()
}

func processAnalyticsMessages_GooglePubsub[T messages.Message](name string, inputChannel chan T) {

	var googlePubsubProducer *common.GooglePubsubProducer

	if enableGooglePubsub {

		defaultPubsubTopic := strings.ReplaceAll(name, " ", "_")

		envVarName := strings.ToUpper(defaultPubsubTopic) + "_PUBSUB_TOPIC"

		pubsubTopic := envvar.GetString(envVarName, defaultPubsubTopic)

		core.Log("analytics %s google pubsub topic: %s", name, pubsubTopic)

		config := common.GooglePubsubConfig{
			ProjectId:          service.GoogleProjectId,
			Topic:              pubsubTopic,
			MessageChannelSize: 10 * 1024, // todo: env var
		}

		var err error
		googlePubsubProducer, err = common.CreateGooglePubsubProducer(service.Context, config)
		if err != nil {
			core.Error("could not create google pubsub producer for analytics %s: %v", name, err)
			os.Exit(1)
		}
	}

	go func() {
		for {
			message := <-inputChannel
			core.Debug("processing analytics %s message", name)
			messageData := message.Write(make([]byte, message.GetMaxSize()))
			if enableGooglePubsub {
				core.Debug("sent analytics %s message to google pubsub", name)
				googlePubsubProducer.MessageChannel <- messageData
			}
		}
	}()
}
