package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/messages"
	"github.com/networknext/next/modules/portal"

	"github.com/redis/go-redis/v9"
)

var service *common.Service

var redisPortalHostname string
var redisPortalCluster []string
var redisServerBackendHostname string
var redisServerBackendCluster []string
var redisRelayBackendHostname string
var sessionCruncherURL string
var serverCruncherURL string

var enableRedisTimeSeries bool
var redisTimeSeriesCluster []string
var redisTimeSeriesHostname string

func main() {

	enableRedisTimeSeries = envvar.GetBool("ENABLE_REDIS_TIME_SERIES", false)
	redisTimeSeriesCluster = envvar.GetStringArray("REDIS_TIME_SERIES_CLUSTER", []string{})
	redisTimeSeriesHostname = envvar.GetString("REDIS_TIME_SERIES_HOSTNAME", "127.0.0.1:6379")

	if enableRedisTimeSeries {
		core.Debug("redis time series cluster: %s", redisTimeSeriesCluster)
		core.Debug("redis time series hostname: %s", redisTimeSeriesHostname)
	}

	redisPortalCluster = envvar.GetStringArray("REDIS_PORTAL_CLUSTER", []string{})
	redisPortalHostname = envvar.GetString("REDIS_PORTAL_HOSTNAME", "127.0.0.1:6379")
	redisServerBackendCluster = envvar.GetStringArray("REDIS_SERVER_BACKEND_CLUSTER", []string{})
	redisServerBackendHostname = envvar.GetString("REDIS_SERVER_BACKEND_HOSTNAME", "127.0.0.1:6379")
	redisRelayBackendHostname = envvar.GetString("REDIS_RELAY_BACKEND_HOSTNAME", "127.0.0.1:6379")
	sessionCruncherURL = envvar.GetString("SESSION_CRUNCHER_URL", "http://127.0.0.1:40200")
	serverCruncherURL = envvar.GetString("SERVER_CRUNCHER_URL", "http://127.0.0.1:40300")

	sessionInsertBatchSize := envvar.GetInt("SESSION_INSERT_BATCH_SIZE", 10000)
	serverInsertBatchSize := envvar.GetInt("SERVER_INSERT_BATCH_SIZE", 10000)
	relayInsertBatchSize := envvar.GetInt("RELAY_INSERT_BATCH_SIZE", 10000)
	nearRelayInsertBatchSize := envvar.GetInt("NEAR_RELAY_INSERT_BATCH_SIZE", 10000)

	reps := envvar.GetInt("REPS", 1)

	service = common.CreateService("portal_cruncher")

	core.Debug("redis portal cluster: %v", redisPortalCluster)
	core.Debug("redis portal hostname: %s", redisPortalHostname)
	core.Debug("redis server backend cluster: %s", redisServerBackendCluster)
	core.Debug("redis server backend hostname: %s", redisServerBackendHostname)
	core.Debug("redis relay backend hostname: %s", redisRelayBackendHostname)
	core.Debug("session cruncher url: %s", sessionCruncherURL)
	core.Debug("server cruncher url: %s", serverCruncherURL)

	core.Debug("session insert batch size: %d", sessionInsertBatchSize)
	core.Debug("server insert batch size: %d", serverInsertBatchSize)
	core.Debug("relay insert batch size: %d", relayInsertBatchSize)
	core.Debug("near relay insert batch size: %d", nearRelayInsertBatchSize)

	if !service.Local {
		service.LoadIP2Location()
	}

	for j := 0; j < reps; j++ {
		ProcessSessionUpdateMessages(service, sessionInsertBatchSize)
		ProcessServerUpdateMessages(service, serverInsertBatchSize)
		ProcessNearRelayUpdateMessages(service, nearRelayInsertBatchSize)
		ProcessRelayUpdateMessages(service, redisRelayBackendHostname, relayInsertBatchSize)
	}

	service.StartWebServer()

	service.WaitForShutdown()
}

// -------------------------------------------------------------------------------

func ProcessSessionUpdateMessages(service *common.Service, batchSize int) {

	name := "session update"

	var redisClient redis.Cmdable
	if len(redisPortalCluster) > 0 {
		redisClient = common.CreateRedisClusterClient(redisPortalCluster)
	} else {
		redisClient = common.CreateRedisClient(redisPortalHostname)
	}

	sessionInserter := portal.CreateSessionInserter(service.Context, redisClient, sessionCruncherURL, batchSize)

	streamName := strings.ReplaceAll(name, " ", "_")
	consumerGroup := streamName

	config := common.RedisStreamsConfig{
		RedisHostname: redisServerBackendHostname,
		RedisCluster:  redisServerBackendCluster,
		StreamName:    streamName,
		ConsumerGroup: consumerGroup,
	}

	consumer, err := common.CreateRedisStreamsConsumer(service.Context, config)
	if err != nil {
		core.Error("could not create redis streams consumer for %s: %v", name, err)
		os.Exit(1)
	}

	var countersPublisher *common.RedisCountersPublisher

	if enableRedisTimeSeries {

		countersConfig := common.RedisCountersConfig{
			RedisHostname: redisTimeSeriesHostname,
			RedisCluster:  redisTimeSeriesCluster,
		}
		countersPublisher, err = common.CreateRedisCountersPublisher(service.Context, countersConfig)
		if err != nil {
			core.Error("could not create redis counters publisher: %v", err)
			os.Exit(1)
		}
	}

	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case messageData := <-consumer.MessageChannel:
				ProcessSessionUpdate(messageData, sessionInserter, countersPublisher)
			}
		}
	}()
}

func ProcessSessionUpdate(messageData []byte, sessionInserter *portal.SessionInserter, countersPublisher *common.RedisCountersPublisher) {

	message := messages.PortalSessionUpdateMessage{}
	err := message.Read(messageData)
	if err != nil {
		core.Error("could not read session update message: %v", err)
		return
	}

	core.Debug("received session update message")

	sessionId := message.SessionId

	userHash := message.UserHash

	var isp string
	if !service.Local {
		isp = service.GetISP(message.ClientAddress.IP)
	} else {
		isp = "Local"
	}

	sessionData := portal.SessionData{
		SessionId:      message.SessionId,
		UserHash:       message.UserHash,
		StartTime:      message.StartTime,
		ISP:            isp,
		ConnectionType: message.ConnectionType,
		PlatformType:   message.PlatformType,
		Latitude:       message.Latitude,
		Longitude:      message.Longitude,
		DirectRTT:      message.BestDirectRTT,
		NextRTT:        message.BestNextRTT,
		BuyerId:        message.BuyerId,
		DatacenterId:   message.DatacenterId,
		ServerAddress:  message.ServerAddress.String(),
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
		InternalEvents:   message.InternalEvents,
		SessionEvents:    message.SessionEvents,
		DirectKbpsUp:     message.DirectKbpsUp,
		DirectKbpsDown:   message.DirectKbpsUp,
		NextKbpsUp:       message.NextKbpsUp,
		NextKbpsDown:     message.NextKbpsDown,
		Next:             message.Next,
	}

	sessionInserter.Insert(service.Context, sessionId, userHash, message.Next, message.BestScore, &sessionData, &sliceData)

	if enableRedisTimeSeries {
		countersPublisher.MessageChannel <- "session_update"
		if message.Retry {
			countersPublisher.MessageChannel <- "retry"
		}
		if message.FallbackToDirect {
			countersPublisher.MessageChannel <- "fallback_to_direct"
		}
	}
}

// -------------------------------------------------------------------------------

func ProcessServerUpdateMessages(service *common.Service, batchSize int) {

	name := "server update"

	var redisClient redis.Cmdable
	if len(redisPortalCluster) > 0 {
		redisClient = common.CreateRedisClusterClient(redisPortalCluster)
	} else {
		redisClient = common.CreateRedisClient(redisPortalHostname)
	}

	serverInserter := portal.CreateServerInserter(service.Context, redisClient, serverCruncherURL, batchSize)

	streamName := strings.ReplaceAll(name, " ", "_")
	consumerGroup := streamName

	config := common.RedisStreamsConfig{
		RedisHostname: redisServerBackendHostname,
		RedisCluster:  redisServerBackendCluster,
		StreamName:    streamName,
		ConsumerGroup: consumerGroup,
	}

	consumer, err := common.CreateRedisStreamsConsumer(service.Context, config)
	if err != nil {
		core.Error("could not create redis streams consumer for %s: %v", name, err)
		os.Exit(1)
	}

	var countersPublisher *common.RedisCountersPublisher

	if enableRedisTimeSeries {

		countersConfig := common.RedisCountersConfig{
			RedisHostname: redisTimeSeriesHostname,
			RedisCluster:  redisTimeSeriesCluster,
		}
		countersPublisher, err = common.CreateRedisCountersPublisher(service.Context, countersConfig)
		if err != nil {
			core.Error("could not create redis counters publisher: %v", err)
			os.Exit(1)
		}
	}

	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case messageData := <-consumer.MessageChannel:
				ProcessServerUpdate(messageData, serverInserter, countersPublisher)
			}
		}
	}()
}

func ProcessServerUpdate(messageData []byte, serverInserter *portal.ServerInserter, countersPublisher *common.RedisCountersPublisher) {

	message := messages.PortalServerUpdateMessage{}
	err := message.Read(messageData)
	if err != nil {
		core.Error("could not read server update message: %v", err)
		return
	}

	core.Debug("received server update message")

	serverData := portal.ServerData{
		ServerAddress:    message.ServerAddress.String(),
		SDKVersion_Major: message.SDKVersion_Major,
		SDKVersion_Minor: message.SDKVersion_Minor,
		SDKVersion_Patch: message.SDKVersion_Patch,
		BuyerId:          message.BuyerId,
		DatacenterId:     message.DatacenterId,
		NumSessions:      message.NumSessions,
		Uptime:           message.Uptime,
	}

	serverInserter.Insert(service.Context, &serverData)

	if enableRedisTimeSeries {
		countersPublisher.MessageChannel <- "server_update"
	}
}

// -------------------------------------------------------------------------------

func ProcessNearRelayUpdateMessages(service *common.Service, batchSize int) {

	name := "near relay update"

	var redisClient redis.Cmdable
	if len(redisPortalCluster) > 0 {
		redisClient = common.CreateRedisClusterClient(redisPortalCluster)
	} else {
		redisClient = common.CreateRedisClient(redisPortalHostname)
	}

	nearRelayInserter := portal.CreateNearRelayInserter(redisClient, batchSize)

	streamName := strings.ReplaceAll(name, " ", "_")
	consumerGroup := streamName

	config := common.RedisStreamsConfig{
		RedisHostname: redisServerBackendHostname,
		RedisCluster:  redisServerBackendCluster,
		StreamName:    streamName,
		ConsumerGroup: consumerGroup,
	}

	consumer, err := common.CreateRedisStreamsConsumer(service.Context, config)
	if err != nil {
		core.Error("could not create redis streams consumer for %s: %v", name, err)
		os.Exit(1)
	}

	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case messageData := <-consumer.MessageChannel:
				ProcessNearRelayUpdate(messageData, nearRelayInserter)
			}
		}
	}()
}

func ProcessNearRelayUpdate(messageData []byte, nearRelayInserter *portal.NearRelayInserter) {

	message := messages.PortalNearRelayUpdateMessage{}
	err := message.Read(messageData)
	if err != nil {
		core.Error("could not read near relay update message: %v", err)
		return
	}

	core.Debug("received near relay update message")

	sessionId := message.SessionId

	nearRelayData := portal.NearRelayData{
		Timestamp:           message.Timestamp,
		NumNearRelays:       message.NumNearRelays,
		NearRelayId:         message.NearRelayId,
		NearRelayRTT:        message.NearRelayRTT,
		NearRelayJitter:     message.NearRelayJitter,
		NearRelayPacketLoss: message.NearRelayPacketLoss,
	}

	nearRelayInserter.Insert(service.Context, sessionId, &nearRelayData)
}

// -------------------------------------------------------------------------------

func ProcessRelayUpdateMessages(service *common.Service, redisStreams string, batchSize int) {

	name := "relay update"

	var redisClient redis.Cmdable
	if len(redisPortalCluster) > 0 {
		redisClient = common.CreateRedisClusterClient(redisPortalCluster)
	} else {
		redisClient = common.CreateRedisClient(redisPortalHostname)
	}

	relayInserter := portal.CreateRelayInserter(redisClient, batchSize)

	streamName := strings.ReplaceAll(name, " ", "_")
	consumerGroup := streamName

	config := common.RedisStreamsConfig{
		RedisHostname: redisStreams,
		StreamName:    streamName,
		ConsumerGroup: consumerGroup,
	}

	consumer, err := common.CreateRedisStreamsConsumer(service.Context, config)
	if err != nil {
		core.Error("could not create redis streams consumer for %s: %v", name, err)
		os.Exit(1)
	}

	var timeSeriesPublisher *common.RedisTimeSeriesPublisher

	var countersPublisher *common.RedisCountersPublisher

	if enableRedisTimeSeries {

		timeSeriesConfig := common.RedisTimeSeriesConfig{
			RedisHostname: redisTimeSeriesHostname,
			RedisCluster:  redisTimeSeriesCluster,
		}
		var err error
		timeSeriesPublisher, err = common.CreateRedisTimeSeriesPublisher(service.Context, timeSeriesConfig)
		if err != nil {
			core.Error("could not create redis time series publisher: %v", err)
			os.Exit(1)
		}

		countersConfig := common.RedisCountersConfig{
			RedisHostname: redisTimeSeriesHostname,
			RedisCluster:  redisTimeSeriesCluster,
		}
		countersPublisher, err = common.CreateRedisCountersPublisher(service.Context, countersConfig)
		if err != nil {
			core.Error("could not create redis counters publisher: %v", err)
			os.Exit(1)
		}
	}

	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case messageData := <-consumer.MessageChannel:
				ProcessRelayUpdate(messageData, relayInserter, timeSeriesPublisher, countersPublisher)
			}
		}
	}()
}

func ProcessRelayUpdate(messageData []byte, relayInserter *portal.RelayInserter, timeSeriesPublisher *common.RedisTimeSeriesPublisher, countersPublisher *common.RedisCountersPublisher) {

	message := messages.PortalRelayUpdateMessage{}
	err := message.Read(messageData)
	if err != nil {
		core.Error("could not read relay update message: %v", err)
		return
	}

	core.Debug("received relay update message")

	relayData := portal.RelayData{
		RelayId:      message.RelayId,
		RelayName:    message.RelayName,
		RelayAddress: message.RelayAddress.String(),
		NumSessions:  message.SessionCount,
		MaxSessions:  message.MaxSessions,
		StartTime:    message.StartTime,
		RelayFlags:   message.RelayFlags,
		RelayVersion: message.RelayVersion,
	}

	relayInserter.Insert(service.Context, &relayData)

	if enableRedisTimeSeries {

		// send time series to redis

		timeSeriesMessage := common.RedisTimeSeriesMessage{}

		timeSeriesMessage.Timestamp = uint64(time.Now().UnixNano() / 1000000)

		timeSeriesMessage.Keys = []string{
			fmt.Sprintf("relay_%016x_session_count", message.RelayId),
			fmt.Sprintf("relay_%016x_envelope_bandwidth_up_kbps", message.RelayId),
			fmt.Sprintf("relay_%016x_envelope_bandwidth_down_kbps", message.RelayId),
			fmt.Sprintf("relay_%016x_packets_sent_per_second", message.RelayId),
			fmt.Sprintf("relay_%016x_packets_received_per_second", message.RelayId),
			fmt.Sprintf("relay_%016x_bandwidth_sent_kbps", message.RelayId),
			fmt.Sprintf("relay_%016x_bandwidth_received_kbps", message.RelayId),
			fmt.Sprintf("relay_%016x_near_pings_per_second", message.RelayId),
			fmt.Sprintf("relay_%016x_relay_pings_per_second", message.RelayId),
			fmt.Sprintf("relay_%016x_num_routable", message.RelayId),
			fmt.Sprintf("relay_%016x_num_unroutable", message.RelayId),
		}

		timeSeriesMessage.Values = []float64{
			float64(message.SessionCount),
			float64(message.EnvelopeBandwidthUpKbps),
			float64(message.EnvelopeBandwidthDownKbps),
			float64(message.PacketsSentPerSecond),
			float64(message.PacketsReceivedPerSecond),
			float64(message.BandwidthSentKbps),
			float64(message.BandwidthReceivedKbps),
			float64(message.NearPingsPerSecond),
			float64(message.RelayPingsPerSecond),
			float64(message.NumRoutable),
			float64(message.NumUnroutable),
		}

		timeSeriesPublisher.MessageChannel <- &timeSeriesMessage

		// send counters to redis

		countersPublisher.MessageChannel <- "relay_update"
	}
}

// -------------------------------------------------------------------------------
