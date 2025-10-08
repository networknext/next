package main

import (
	_ "embed"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/handlers"
	"github.com/networknext/next/modules/messages"
	"github.com/networknext/next/modules/portal"

	"github.com/hamba/avro"
	"github.com/redis/go-redis/v9"
)

var service *common.Service

var pingKey []byte

var channelSize int
var maxPacketSize int
var serverBackendAddress net.UDPAddr
var serverBackendPublicKey []byte
var serverBackendPrivateKey []byte
var relayBackendPublicKey []byte
var relayBackendPrivateKey []byte

var fallbackToDirectChannel chan uint64

var portalSessionUpdateMessageChannel chan *messages.PortalSessionUpdateMessage
var portalServerUpdateMessageChannel chan *messages.PortalServerUpdateMessage
var portalClientRelayUpdateMessageChannel chan *messages.PortalClientRelayUpdateMessage
var portalServerRelayUpdateMessageChannel chan *messages.PortalServerRelayUpdateMessage

var analyticsServerInitMessageChannel chan *messages.AnalyticsServerInitMessage
var analyticsServerUpdateMessageChannel chan *messages.AnalyticsServerUpdateMessage
var analyticsSessionUpdateMessageChannel chan *messages.AnalyticsSessionUpdateMessage
var analyticsSessionSummaryMessageChannel chan *messages.AnalyticsSessionSummaryMessage
var analyticsClientRelayPingMessageChannel chan *messages.AnalyticsClientRelayPingMessage
var analyticsServerRelayPingMessageChannel chan *messages.AnalyticsServerRelayPingMessage

var enableGooglePubsub bool

var shuttingDownMutex sync.Mutex
var shuttingDown bool

var portalNextSessionsOnly bool

var sessionInserter *portal.SessionInserter
var sessionCruncherURL string
var serverCruncherURL string
var sessionInsertBatchSize int
var serverInsertBatchSize int
var clientRelayInsertBatchSize int
var serverRelayInsertBatchSize int

var enableRedisTimeSeries bool
var redisTimeSeriesHostname string
var redisTimeSeriesCluster []string

var redisPortalHostname string
var redisPortalCluster []string

var countersPublisher *common.RedisCountersPublisher

var startTime int64

var initialDelay int

//go:embed client_relay_ping.json
var clientRelayPingSchemaData string

//go:embed server_relay_ping.json
var serverRelayPingSchemaData string

//go:embed server_update.json
var serverUpdateSchemaData string

//go:embed server_init.json
var serverInitSchemaData string

//go:embed session_update.json
var sessionUpdateSchemaData string

//go:embed session_summary.json
var sessionSummarySchemaData string

var clientRelayPingSchema avro.Schema
var serverRelayPingSchema avro.Schema
var serverUpdateSchema avro.Schema
var serverInitSchema avro.Schema
var sessionUpdateSchema avro.Schema
var sessionSummarySchema avro.Schema

var enableIP2Location bool

func main() {

	startTime = time.Now().Unix()

	service = common.CreateService("server_backend")

	service.ConnectionDrain = true

	pingKey = envvar.GetBase64("PING_KEY", []byte{})

	channelSize = envvar.GetInt("CHANNEL_SIZE", 10*1024*1024)
	maxPacketSize = envvar.GetInt("UDP_MAX_PACKET_SIZE", 1384)
	serverBackendAddress = envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:40000"))
	serverBackendPublicKey = envvar.GetBase64("SERVER_BACKEND_PUBLIC_KEY", []byte{})
	serverBackendPrivateKey = envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", []byte{})
	relayBackendPrivateKey = envvar.GetBase64("RELAY_BACKEND_PRIVATE_KEY", []byte{})
	relayBackendPublicKey = envvar.GetBase64("RELAY_BACKEND_PUBLIC_KEY", []byte{})
	enableGooglePubsub = envvar.GetBool("ENABLE_GOOGLE_PUBSUB", false)
	portalNextSessionsOnly = envvar.GetBool("PORTAL_NEXT_SESSIONS_ONLY", false)
	sessionCruncherURL = envvar.GetString("SESSION_CRUNCHER_URL", "http://127.0.0.1:40200")
	serverCruncherURL = envvar.GetString("SERVER_CRUNCHER_URL", "http://127.0.0.1:40300")
	sessionInsertBatchSize = envvar.GetInt("SESSION_INSERT_BATCH_SIZE", 10000)
	serverInsertBatchSize = envvar.GetInt("SERVER_INSERT_BATCH_SIZE", 10000)
	clientRelayInsertBatchSize = envvar.GetInt("CLIENT_RELAY_INSERT_BATCH_SIZE", 10000)
	clientRelayInsertBatchSize = envvar.GetInt("SERVER_RELAY_INSERT_BATCH_SIZE", 10000)
	enableRedisTimeSeries = envvar.GetBool("ENABLE_REDIS_TIME_SERIES", false)
	redisTimeSeriesCluster = envvar.GetStringArray("REDIS_TIME_SERIES_CLUSTER", []string{})
	redisTimeSeriesHostname = envvar.GetString("REDIS_TIME_SERIES_HOSTNAME", "127.0.0.1:6379")
	redisPortalCluster = envvar.GetStringArray("REDIS_PORTAL_CLUSTER", []string{})
	redisPortalHostname = envvar.GetString("REDIS_PORTAL_HOSTNAME", "127.0.0.1:6379")
	initialDelay = envvar.GetInt("INITIAL_DELAY", 90)
	enableIP2Location = envvar.GetBool("ENABLE_IP2LOCATION", false)

	if enableRedisTimeSeries {
		core.Debug("redis time series cluster: %s", redisTimeSeriesCluster)
		core.Debug("redis time series hostname: %s", redisTimeSeriesHostname)
	}

	core.Debug("redis portal cluster: %s", redisPortalCluster)
	core.Debug("redis portal hostname: %s", redisPortalHostname)

	core.Debug("channel size: %d", channelSize)
	core.Debug("max packet size: %d bytes", maxPacketSize)
	core.Debug("server backend address: %s", serverBackendAddress.String())
	core.Debug("enable google pubsub: %v", enableGooglePubsub)
	core.Debug("portal next sessions only: %v", portalNextSessionsOnly)
	core.Debug("session cruncher url: %s", sessionCruncherURL)
	core.Debug("server cruncher url: %s", serverCruncherURL)
	core.Debug("session insert batch size: %d", sessionInsertBatchSize)
	core.Debug("server insert batch size: %d", serverInsertBatchSize)
	core.Debug("client relay insert batch size: %d", clientRelayInsertBatchSize)
	core.Debug("server relay insert batch size: %d", serverRelayInsertBatchSize)
	core.Debug("enable ip2location: %v", enableIP2Location)

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

	if len(relayBackendPublicKey) == 0 {
		panic("RELAY_BACKEND_PUBLIC_KEY must be specified")
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

	// initialize avro schemas for sending analytics data to pubsub -> bigquery ingestion

	if enableGooglePubsub {

		var err error
		sessionUpdateSchema, err = avro.Parse(sessionUpdateSchemaData)
		if err != nil {
			panic(fmt.Sprintf("invalid session update schema: %v", err))
		}

		sessionSummarySchema, err = avro.Parse(sessionSummarySchemaData)
		if err != nil {
			panic(fmt.Sprintf("invalid session summary schema: %v", err))
		}

		clientRelayPingSchema, err = avro.Parse(clientRelayPingSchemaData)
		if err != nil {
			panic(fmt.Sprintf("invalid client relay ping schema: %v", err))
		}

		serverRelayPingSchema, err = avro.Parse(serverRelayPingSchemaData)
		if err != nil {
			panic(fmt.Sprintf("invalid server relay ping schema: %v", err))
		}

		serverUpdateSchema, err = avro.Parse(serverUpdateSchemaData)
		if err != nil {
			panic(fmt.Sprintf("invalid server update schema: %v", err))
		}

		serverInitSchema, err = avro.Parse(serverInitSchemaData)
		if err != nil {
			panic(fmt.Sprintf("invalid server init schema: %v", err))
		}
	}

	// initialize fallback to direct channel

	fallbackToDirectChannel = make(chan uint64, channelSize)

	processFallbackToDirect(service, fallbackToDirectChannel)

	// initialize portal message channels

	portalSessionUpdateMessageChannel = make(chan *messages.PortalSessionUpdateMessage, channelSize)
	portalServerUpdateMessageChannel = make(chan *messages.PortalServerUpdateMessage, channelSize)
	portalClientRelayUpdateMessageChannel = make(chan *messages.PortalClientRelayUpdateMessage, channelSize)
	portalServerRelayUpdateMessageChannel = make(chan *messages.PortalServerRelayUpdateMessage, channelSize)

	if enableRedisTimeSeries {

		countersConfig := common.RedisCountersConfig{
			RedisHostname: redisTimeSeriesHostname,
			RedisCluster:  redisTimeSeriesCluster,
		}
		var err error
		countersPublisher, err = common.CreateRedisCountersPublisher(service.Context, countersConfig)
		if err != nil {
			core.Error("could not create redis counters publisher: %v", err)
			os.Exit(1)
		}
	}

	processPortalSessionUpdateMessages(service, portalSessionUpdateMessageChannel)
	processPortalServerUpdateMessages(service, portalServerUpdateMessageChannel)
	processPortalClientRelayUpdateMessages(service, portalClientRelayUpdateMessageChannel)
	processPortalServerRelayUpdateMessages(service, portalServerRelayUpdateMessageChannel)

	// initialize analytics message channels

	analyticsServerInitMessageChannel = make(chan *messages.AnalyticsServerInitMessage, channelSize)
	analyticsServerUpdateMessageChannel = make(chan *messages.AnalyticsServerUpdateMessage, channelSize)
	analyticsSessionUpdateMessageChannel = make(chan *messages.AnalyticsSessionUpdateMessage, channelSize)
	analyticsSessionSummaryMessageChannel = make(chan *messages.AnalyticsSessionSummaryMessage, channelSize)
	analyticsClientRelayPingMessageChannel = make(chan *messages.AnalyticsClientRelayPingMessage, channelSize)
	analyticsServerRelayPingMessageChannel = make(chan *messages.AnalyticsServerRelayPingMessage, channelSize)

	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsServerInitMessage]("server init", analyticsServerInitMessageChannel, serverInitSchema)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsServerUpdateMessage]("server update", analyticsServerUpdateMessageChannel, serverUpdateSchema)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsClientRelayPingMessage]("client relay ping", analyticsClientRelayPingMessageChannel, clientRelayPingSchema)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsServerRelayPingMessage]("server relay ping", analyticsServerRelayPingMessageChannel, serverRelayPingSchema)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsSessionUpdateMessage]("session update", analyticsSessionUpdateMessageChannel, sessionUpdateSchema)
	processAnalyticsMessages_GooglePubsub[*messages.AnalyticsSessionSummaryMessage]("session summary", analyticsSessionSummaryMessageChannel, sessionSummarySchema)

	// start the service

	updateShuttingDown()

	service.UpdateRouteMatrix(relayBackendPublicKey, relayBackendPrivateKey)

	service.SetHealthFunctions(sendTrafficToMe, machineIsHealthy, ready)

	if enableIP2Location {
		service.LoadIP2Location()
	}

	service.UpdateMagic()

	service.StartUDPServer(packetHandler)

	service.StartWebServer()

	service.WaitForShutdown()
}

func updateShuttingDown() {

	// grab google cloud instance name from metadata

	result, instanceName := common.Bash("curl -s http://metadata/computeMetadata/v1/instance/hostname -H \"Metadata-Flavor: Google\" --max-time 5 -s 2>/dev/null")
	if !result {
		return // not in google cloud
	}

	instanceName = strings.TrimSuffix(instanceName, "\n")

	tokens := strings.Split(instanceName, ".")

	instanceName = tokens[0]

	core.Log("google cloud instance name is '%s'", instanceName)

	// grab google cloud zone from metadata

	var zone string
	result, zone = common.Bash("curl -s http://metadata/computeMetadata/v1/instance/zone -H \"Metadata-Flavor: Google\" --max-time 5 -s 2>/dev/null")
	if !result {
		return // not in google cloud
	}

	zone = strings.TrimSuffix(zone, "\n")

	tokens = strings.Split(zone, "/")

	zone = tokens[len(tokens)-1]

	core.Log("google cloud zone is '%s'", zone)

	// turn zone into region

	tokens = strings.Split(zone, "-")

	region := strings.Join(tokens[:len(tokens)-1], "-")

	core.Log("google cloud region is '%s'", region)

	go func() {

		ticker := time.NewTicker(100 * time.Millisecond)

		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				_, output := common.Bash(fmt.Sprintf("gcloud compute instance-groups managed list-instances server-backend --region %s", region))

				lines := strings.Split(output, "\n")

				for i := range lines {
					if strings.Contains(lines[i], instanceName) && (strings.Contains(lines[i], "STOPPING") || strings.Contains(lines[i], "DELETING")) {
						shuttingDownMutex.Lock()
						if !shuttingDown {
							core.Log("*** SHUTTING DOWN ***")
							shuttingDown = true
						}
						shuttingDownMutex.Unlock()
						break
					}
				}
			}
		}
	}()
}

func isShuttingDown() bool {
	shuttingDownMutex.Lock()
	value := shuttingDown
	shuttingDownMutex.Unlock()
	return value
}

func sendTrafficToMe() bool {
	routeMatrix, database := service.RouteMatrixAndDatabase()
	return time.Now().Unix() > startTime+int64(initialDelay) && routeMatrix != nil && database != nil && !isShuttingDown() && !service.Stopping
}

func machineIsHealthy() bool {
	return true
}

func ready() bool {
	routeMatrix, database := service.RouteMatrixAndDatabase()
	return routeMatrix != nil && database != nil
}

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	handler := handlers.SDK_Handler{}

	handler.PortalNextSessionsOnly = portalNextSessionsOnly

	handler.PingKey = pingKey
	handler.ServerBackendAddress = serverBackendAddress
	handler.ServerBackendPublicKey = serverBackendPublicKey
	handler.ServerBackendPrivateKey = serverBackendPrivateKey
	handler.RelayBackendPublicKey = relayBackendPublicKey
	handler.RelayBackendPrivateKey = relayBackendPrivateKey
	handler.RouteMatrix, handler.Database = service.RouteMatrixAndDatabase()
	handler.MaxPacketSize = maxPacketSize
	handler.GetMagicValues = func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte) {
		return service.GetMagicValues()
	}

	handler.FallbackToDirectChannel = fallbackToDirectChannel

	handler.PortalSessionUpdateMessageChannel = portalSessionUpdateMessageChannel
	handler.PortalServerUpdateMessageChannel = portalServerUpdateMessageChannel
	handler.PortalClientRelayUpdateMessageChannel = portalClientRelayUpdateMessageChannel
	handler.PortalServerRelayUpdateMessageChannel = portalServerRelayUpdateMessageChannel

	handler.AnalyticsServerInitMessageChannel = analyticsServerInitMessageChannel
	handler.AnalyticsServerUpdateMessageChannel = analyticsServerUpdateMessageChannel
	handler.AnalyticsSessionUpdateMessageChannel = analyticsSessionUpdateMessageChannel
	handler.AnalyticsSessionSummaryMessageChannel = analyticsSessionSummaryMessageChannel
	handler.AnalyticsClientRelayPingMessageChannel = analyticsClientRelayPingMessageChannel
	handler.AnalyticsServerRelayPingMessageChannel = analyticsServerRelayPingMessageChannel

	handler.LocateIP = locateIP_Real
	if service.Env == "dev" {
		handler.LocateIP = locateIP_Dev
	} else if service.Env == "local" && service.Env == "docker" {
		handler.LocateIP = locateIP_Local
	}

	handlers.SDK_PacketHandler(&handler, conn, from, packetData)
}

func locateIP_Local(ip net.IP) (float32, float32) {
	return 41, -93 // iowa
}

func locateIP_Dev(ip net.IP) (float32, float32) {
	ipv4 := ip.To4()
	if ipv4[0] == 34 || ipv4[0] == 35 {
		// This is a raspberry client running in google cloud: mock lat/long of major US cities for testing
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
	}
	// this is a real client. do ip2location with maxmind if enabled, otherwise fallback to fixed location for debugging
	if enableIP2Location {
		return service.GetLocation(ip)
	} else {
		return 40.7128, -74.0060 // new york
	}
}

func locateIP_Real(ip net.IP) (float32, float32) {
	return service.GetLocation(ip)
}

// ------------------------------------------------------------------------------------

func processFallbackToDirect(service *common.Service, channel chan uint64) {
	go func() {
		for {
			_ = <-channel
			if enableRedisTimeSeries {
				countersPublisher.MessageChannel <- "fallback_to_direct"
			}
		}
	}()
}

// ------------------------------------------------------------------------------------

func processPortalSessionUpdateMessages(service *common.Service, inputChannel chan *messages.PortalSessionUpdateMessage) {

	var redisClient redis.Cmdable
	if len(redisPortalCluster) > 0 {
		redisClient = common.CreateRedisClusterClient(redisPortalCluster)
	} else {
		redisClient = common.CreateRedisClient(redisPortalHostname)
	}

	sessionInserter = portal.CreateSessionInserter(service.Context, redisClient, sessionCruncherURL, sessionInsertBatchSize)

	go func() {
		for {
			message := <-inputChannel

			core.Debug("processing portal session update message")

			sessionId := message.SessionId

			var isp string
			if enableIP2Location {
				isp = service.GetISP(message.ClientAddress.IP)
			} else {
				isp = "Local"
			}

			sessionData := portal.SessionData{
				SessionId:      message.SessionId,
				StartTime:      message.StartTime,
				ISP:            isp,
				ConnectionType: message.ConnectionType,
				PlatformType:   message.PlatformType,
				Latitude:       message.Latitude,
				Longitude:      message.Longitude,
				DirectRTT:      message.BestDirectRTT,
				NextRTT:        message.BestNextRTT,
				BuyerId:        message.BuyerId,
				ServerId:       message.ServerId,
				DatacenterId:   message.DatacenterId,
			}

			if message.Next {
				sessionData.NumRouteRelays = int(message.NextNumRouteRelays)
				for i := 0; i < int(message.NextNumRouteRelays); i++ {
					sessionData.RouteRelays[i] = message.NextRouteRelayId[i]
				}
			}

			sliceData := portal.SliceData{
				Timestamp:        message.Timestamp,
				SliceNumber:      message.SliceNumber,
				DirectRTT:        uint32(message.DirectRTT),
				NextRTT:          uint32(message.NextRTT),
				PredictedRTT:     uint32(message.NextPredictedRTT),
				DirectJitter:     uint32(message.DirectJitter),
				NextJitter:       uint32(message.NextJitter),
				RealJitter:       uint32(message.RealJitter),
				DirectPacketLoss: float32(message.DirectPacketLoss),
				NextPacketLoss:   float32(message.NextPacketLoss),
				RealPacketLoss:   float32(message.RealPacketLoss),
				RealOutOfOrder:   float32(message.RealOutOfOrder),
				DeltaTimeMin:     float32(message.DeltaTimeMin),
				DeltaTimeMax:     float32(message.DeltaTimeMax),
				DeltaTimeAvg:     float32(message.DeltaTimeAvg),
				InternalEvents:   message.InternalEvents,
				SessionEvents:    message.SessionEvents,
				DirectKbpsUp:     message.DirectKbpsUp,
				DirectKbpsDown:   message.DirectKbpsDown,
				NextKbpsUp:       message.NextKbpsUp,
				NextKbpsDown:     message.NextKbpsDown,
				Next:             message.Next,
				GameRTT:          message.GameRTT,
				GameJitter:       message.GameJitter,
				GamePacketLoss:   message.GamePacketLoss,
			}

			if message.SendToPortal {
				sessionInserter.Insert(service.Context, sessionId, message.BestNextRTT > 0, message.BestScore, &sessionData, &sliceData)
			}

			if enableRedisTimeSeries {

				if !message.Retry {
					countersPublisher.MessageChannel <- "session_update"
					if message.Next {
						countersPublisher.MessageChannel <- "next_session_update"
					}
					countersPublisher.MessageChannel <- fmt.Sprintf("session_update_%016x", message.BuyerId)
					if message.Next {
						countersPublisher.MessageChannel <- fmt.Sprintf("next_session_update_%016x", message.BuyerId)
					}
				} else {
					countersPublisher.MessageChannel <- "retry"
				}
			}
		}
	}()
}

func processPortalServerUpdateMessages(service *common.Service, inputChannel chan *messages.PortalServerUpdateMessage) {

	var redisClient redis.Cmdable
	if len(redisPortalCluster) > 0 {
		redisClient = common.CreateRedisClusterClient(redisPortalCluster)
	} else {
		redisClient = common.CreateRedisClient(redisPortalHostname)
	}

	serverInserter := portal.CreateServerInserter(service.Context, redisClient, serverCruncherURL, serverInsertBatchSize)

	go func() {
		for {
			message := <-inputChannel

			core.Debug("processing portal server update message")

			serverData := portal.ServerData{
				SDKVersion_Major: message.SDKVersion_Major,
				SDKVersion_Minor: message.SDKVersion_Minor,
				SDKVersion_Patch: message.SDKVersion_Patch,
				BuyerId:          message.BuyerId,
				ServerId:         message.ServerId,
				DatacenterId:     message.DatacenterId,
				NumSessions:      message.NumSessions,
				Uptime:           message.Uptime,
			}

			serverInserter.Insert(service.Context, &serverData)

			if enableRedisTimeSeries {

				countersPublisher.MessageChannel <- "server_update"

				countersPublisher.MessageChannel <- fmt.Sprintf("server_update_%016x", message.BuyerId)
			}
		}
	}()
}

func processPortalClientRelayUpdateMessages(service *common.Service, inputChannel chan *messages.PortalClientRelayUpdateMessage) {

	var redisClient redis.Cmdable
	if len(redisPortalCluster) > 0 {
		redisClient = common.CreateRedisClusterClient(redisPortalCluster)
	} else {
		redisClient = common.CreateRedisClient(redisPortalHostname)
	}

	clientRelayInserter := portal.CreateClientRelayInserter(redisClient, clientRelayInsertBatchSize)

	go func() {
		for {
			message := <-inputChannel

			core.Debug("processing client relay update message")

			sessionId := message.SessionId

			clientRelayData := portal.ClientRelayData{
				Timestamp:             message.Timestamp,
				NumClientRelays:       message.NumClientRelays,
				ClientRelayId:         message.ClientRelayId,
				ClientRelayRTT:        message.ClientRelayRTT,
				ClientRelayJitter:     message.ClientRelayJitter,
				ClientRelayPacketLoss: message.ClientRelayPacketLoss,
			}

			clientRelayInserter.Insert(service.Context, sessionId, &clientRelayData)

			if enableRedisTimeSeries {

				countersPublisher.MessageChannel <- "client_relay_update"
			}
		}
	}()
}

func processPortalServerRelayUpdateMessages(service *common.Service, inputChannel chan *messages.PortalServerRelayUpdateMessage) {

	var redisClient redis.Cmdable
	if len(redisPortalCluster) > 0 {
		redisClient = common.CreateRedisClusterClient(redisPortalCluster)
	} else {
		redisClient = common.CreateRedisClient(redisPortalHostname)
	}

	serverRelayInserter := portal.CreateServerRelayInserter(redisClient, serverRelayInsertBatchSize)

	go func() {
		for {
			message := <-inputChannel

			core.Debug("processing server relay update message")

			sessionId := message.SessionId

			serverRelayData := portal.ServerRelayData{
				Timestamp:             message.Timestamp,
				NumServerRelays:       message.NumServerRelays,
				ServerRelayId:         message.ServerRelayId,
				ServerRelayRTT:        message.ServerRelayRTT,
				ServerRelayJitter:     message.ServerRelayJitter,
				ServerRelayPacketLoss: message.ServerRelayPacketLoss,
			}

			serverRelayInserter.Insert(service.Context, sessionId, &serverRelayData)

			if enableRedisTimeSeries {

				countersPublisher.MessageChannel <- "server_relay_update"
			}
		}
	}()
}

// ------------------------------------------------------------------------------------

func processAnalyticsMessages_GooglePubsub[T any](name string, inputChannel chan T, schema avro.Schema) {

	var googlePubsubProducer *common.GooglePubsubProducer

	if enableGooglePubsub {

		defaultPubsubTopic := strings.ReplaceAll(name, " ", "_")

		envVarName := strings.ToUpper(defaultPubsubTopic) + "_PUBSUB_TOPIC"

		pubsubTopic := envvar.GetString(envVarName, defaultPubsubTopic)

		core.Debug("analytics %s google pubsub topic: %s", name, pubsubTopic)

		config := common.GooglePubsubConfig{
			ProjectId:          service.GoogleProjectId,
			Topic:              pubsubTopic,
			MessageChannelSize: 1024 * 1024,
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
			if enableGooglePubsub {
				data, err := avro.Marshal(schema, &message)
				if err != nil {
					core.Warn("failed to encode %s message: %v", name, err)
					continue
				}
				core.Debug("sent analytics %s message to google pubsub", name)
				googlePubsubProducer.MessageChannel <- data
			}
		}
	}()
}

// ------------------------------------------------------------------------------------
