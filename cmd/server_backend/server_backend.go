package main

import (
	"net"
	"os"
	"strings"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/handlers"
	"github.com/networknext/next/modules/messages"
)

var service *common.Service

var pingKey []byte

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
var analyticsNearRelayPingMessageChannel chan *messages.AnalyticsNearRelayPingMessage

var enableGooglePubsub bool
var enableRedisStreams bool

var redisHostname string
var redisPassword string

func main() {

	service = common.CreateService("server_backend")

	pingKey = envvar.GetBase64("PING_KEY", []byte{})

	channelSize = envvar.GetInt("CHANNEL_SIZE", 10*1024)
	maxPacketSize = envvar.GetInt("UDP_MAX_PACKET_SIZE", 4096)
	serverBackendAddress = envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:40000"))
	serverBackendPublicKey = envvar.GetBase64("SERVER_BACKEND_PUBLIC_KEY", []byte{})
	serverBackendPrivateKey = envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", []byte{})
	relayBackendPrivateKey = envvar.GetBase64("RELAY_BACKEND_PRIVATE_KEY", []byte{})
	enableGooglePubsub = envvar.GetBool("ENABLE_GOOGLE_PUBSUB", false)
	enableRedisStreams = envvar.GetBool("ENABLE_REDIS_STREAMS", true)
	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")

	core.Debug("channel size: %d", channelSize)
	core.Debug("max packet size: %d bytes", maxPacketSize)
	core.Debug("server backend address: %s", serverBackendAddress.String())
	core.Debug("enable google pubsub: %v", enableGooglePubsub)
	core.Debug("enable redis streams: %v", enableRedisStreams)
	core.Debug("redis hostname: %s", redisHostname)

	if len(pingKey) == 0 {
		core.Error("You must supply PING_KEY")
		os.Exit(1)
	}

	if len(serverBackendPublicKey) == 0 {
		panic("SERVER_BACKEND_PUBLIC_KEY must be specified")
	}

	if len(serverBackendPrivateKey) == 0 {
		panic("SERVER_BACKEND_PRIVATE_KEY must be specified")
	}

	if len(relayBackendPrivateKey) == 0 {
		panic("RELAY_BACKEND_PRIVATE_KEY must be specified")
	}

	core.Debug("ping key: %x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x",
		pingKey[0],
		pingKey[1],
		pingKey[2],
		pingKey[3],
		pingKey[4],
		pingKey[5],
		pingKey[6],
		pingKey[7],
		pingKey[8],
		pingKey[9],
		pingKey[10],
		pingKey[11],
		pingKey[12],
		pingKey[13],
		pingKey[14],
		pingKey[15],
		pingKey[16],
		pingKey[17],
		pingKey[18],
		pingKey[19],
		pingKey[20],
		pingKey[21],
		pingKey[22],
		pingKey[23],
		pingKey[24],
		pingKey[25],
		pingKey[26],
		pingKey[27],
		pingKey[28],
		pingKey[29],
		pingKey[30],
		pingKey[31],
	)

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
	analyticsNearRelayPingMessageChannel = make(chan *messages.AnalyticsNearRelayPingMessage, channelSize)

	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsServerInitMessage]("server init", analyticsServerInitMessageChannel)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsServerUpdateMessage]("server update", analyticsServerUpdateMessageChannel)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsNearRelayPingMessage]("near relay ping", analyticsNearRelayPingMessageChannel)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsSessionUpdateMessage]("session update", analyticsSessionUpdateMessageChannel)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsSessionSummaryMessage]("session summary", analyticsSessionSummaryMessageChannel)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsMatchDataMessage]("match data", analyticsMatchDataMessageChannel)

	// start the service

	service.UpdateRouteMatrix()

	service.SetHealthFunctions(sendTrafficToMe, machineIsHealthy, ready)

	if !service.Local {
		service.LoadIP2Location()
	}

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

func ready() bool {
	routeMatrix, database := service.RouteMatrixAndDatabase()
	return routeMatrix != nil && len(routeMatrix.RelayIds) > 0 && database != nil
}

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	// todo
	core.Debug("received packet type %d from %s (%d bytes)", packetData[0], from.String(), len(packetData) )

	handler := handlers.SDK_Handler{}

	handler.PingKey = pingKey
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
	handler.AnalyticsNearRelayPingMessageChannel = analyticsNearRelayPingMessageChannel

	handler.LocateIP = locateIP_Local
	if service.Env == "dev" {
		handler.LocateIP = locateIP_Dev
	} else if service.Env != "local" {
		handler.LocateIP = locateIP_Real
	}

	handlers.SDK_PacketHandler(&handler, conn, from, packetData)
}

func locateIP_Local(ip net.IP) (float32, float32) {
	return 41, -93 // iowa
}

func locateIP_Dev(ip net.IP) (float32, float32) {
	index := common.RandomInt(0, 22)
	switch index {
	case 0:
		return 33.748798, -84.387703 // atlanta
	case 1:
		return 32.776699, -96.796997 // dallas
	case 2:
		return 40.712799, -74.005997 // new york
	case 3:
		return 34.052200, -118.243698 // los angeles
	case 4:
		return 25.761700, -80.191803 // miami
	case 5:
		return 41.878101, -87.629799 // chicago
	case 6:
		return 47.606201, -122.332100 // seattle
	case 7:
		return 37.338699, -121.885300 // sanjose
	case 8:
		return 39.043800, -77.487396 // virginia
	case 9:
		return 42.360100, -71.058899 // boston
	case 10:
		return 29.760401, -95.369797 // houston
	case 11:
		return 39.099701, -94.578598 // kansas
	case 12:
		return 44.977798, -93.264999 // minneapolis
	case 13:
		return 39.952599, -75.165199 // philadelphia
	case 14:
		return 40.417301, -82.907097 // ohio
	case 15:
		return 45.839901, -119.700600 // oregon
	case 16:
		return 39.739201, -104.990303 // denver
	case 17:
		return 36.171600, -115.139099 // las vegas
	case 18:
		return 45.515202, -122.678398 // portland
	case 19:
		return 33.448399, -112.073997 // phoenix
	case 20:
		return 41.877998, -93.097702 // iowa
	case 21:
		return 33.836102, -81.163696 // south carolina
	case 22:
		return 40.760799, -111.890999 // salt lake city
	}
	return 0, 0
}

func locateIP_Real(ip net.IP) (float32, float32) {
	return service.GetLocation(ip)
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

		core.Debug("analytics %s google pubsub topic: %s", name, pubsubTopic)

		config := common.GooglePubsubConfig{
			ProjectId:          service.GoogleProjectId,
			Topic:              pubsubTopic,
			MessageChannelSize: 10 * 1024,
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
