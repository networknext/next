package main

import (
	"os"
	"strings"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/messages"
	"github.com/networknext/next/modules/portal"

	"github.com/gomodule/redigo/redis"
)

var service *common.Service

var redisPortalHostname string
var redisRelayBackendHostname string
var redisServerBackendHostname string

var pool *redis.Pool

func main() {

	redisPortalHostname = envvar.GetString("REDIS_PORTAL_HOSTNAME", "127.0.0.1:6379")
	redisServerBackendHostname = envvar.GetString("REDIS_SERVER_BACKEND_HOSTNAME", "127.0.0.1:6379")
	redisRelayBackendHostname = envvar.GetString("REDIS_RELAY_BACKEND_HOSTNAME", "127.0.0.1:6379")
	redisPoolActive := envvar.GetInt("REDIS_POOL_ACTIVE", 1000)
	redisPoolIdle := envvar.GetInt("REDIS_POOL_IDLE", 10000)

	sessionInsertBatchSize := envvar.GetInt("SESSION_INSERT_BATCH_SIZE", 1000)
	serverInsertBatchSize := envvar.GetInt("SERVER_INSERT_BATCH_SIZE", 1000)
	relayInsertBatchSize := envvar.GetInt("RELAY_INSERT_BATCH_SIZE", 1000)
	nearRelayInsertBatchSize := envvar.GetInt("NEAR_RELAY_INSERT_BATCH_SIZE", 1000)

	reps := envvar.GetInt("REPS", 1)

	service = common.CreateService("portal_cruncher")

	core.Debug("redis portal hostname: %s", redisPortalHostname)
	core.Debug("redis relay backend hostname: %s", redisRelayBackendHostname)
	core.Debug("redis server backend hostname: %s", redisServerBackendHostname)
	core.Debug("redis pool active: %d", redisPoolActive)
	core.Debug("redis pool idle: %d", redisPoolIdle)

	core.Debug("session insert batch size: %d", sessionInsertBatchSize)
	core.Debug("server insert batch size: %d", serverInsertBatchSize)
	core.Debug("relay insert batch size: %d", relayInsertBatchSize)
	core.Debug("near relay insert batch size: %d", nearRelayInsertBatchSize)

	if !service.Local {
		service.LoadIP2Location()
	}

	pool = common.CreateRedisPool(redisPortalHostname, redisPoolActive, redisPoolIdle)

	for j := 0; j < reps; j++ {

		sessionInserter := portal.CreateSessionInserter(pool, sessionInsertBatchSize)
		serverInserter := portal.CreateServerInserter(pool, serverInsertBatchSize)
		nearRelayInserter := portal.CreateNearRelayInserter(pool, nearRelayInsertBatchSize)
		relayInserter := portal.CreateRelayInserter(pool, relayInsertBatchSize)

		ProcessSessionUpdateMessages(service, redisServerBackendHostname, "session update", sessionInserter)
		ProcessServerUpdateMessages(service, redisServerBackendHostname, "server update", serverInserter)
		ProcessNearRelayUpdateMessages(service, redisServerBackendHostname, "near relay update", nearRelayInserter)
		ProcessRelayUpdateMessages(service, redisRelayBackendHostname, "relay update", relayInserter)

	}

	service.StartWebServer()

	service.WaitForShutdown()
}

// -------------------------------------------------------------------------------

func ProcessSessionUpdateMessages(service *common.Service, redisHostname string, name string, sessionInserter *portal.SessionInserter) {

	streamName := strings.ReplaceAll(name, " ", "_")
	consumerGroup := streamName

	config := common.RedisStreamsConfig{
		RedisHostname: redisHostname,
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
				ProcessSessionUpdate(messageData, sessionInserter)
			}
		}
	}()
}

func ProcessSessionUpdate(messageData []byte, sessionInserter *portal.SessionInserter) {

	message := messages.PortalSessionUpdateMessage{}
	err := message.Read(messageData)
	if err != nil {
		core.Error("could not read session update message: %v", err)
		return
	}

	core.Debug("received session update message")

	sessionId := message.SessionId

	userHash := message.UserHash

	next := message.Next

	score := uint32(0)
	if next {
		score = uint32(message.NextRTT)
	} else {
		score = 10000 - uint32(message.DirectRTT)
	}

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
		DirectRTT:      uint32(message.DirectRTT),
		NextRTT:        uint32(message.NextRTT),
		MatchId:        message.MatchId,
		BuyerId:        message.BuyerId,
		DatacenterId:   message.DatacenterId,
		ServerAddress:  message.ServerAddress.String(),
	}

	sliceData := portal.SliceData{
		Timestamp:        uint64(time.Now().Unix()),
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

	sessionInserter.Insert(sessionId, userHash, score, next, &sessionData, &sliceData)
}

// -------------------------------------------------------------------------------

func ProcessServerUpdateMessages(service *common.Service, redisHostname string, name string, serverInserter *portal.ServerInserter) {

	streamName := strings.ReplaceAll(name, " ", "_")
	consumerGroup := streamName

	config := common.RedisStreamsConfig{
		RedisHostname: redisHostname,
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
				ProcessServerUpdate(messageData, serverInserter)
			}
		}
	}()
}

func ProcessServerUpdate(messageData []byte, serverInserter *portal.ServerInserter) {

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
		MatchId:          message.MatchId,
		BuyerId:          message.BuyerId,
		DatacenterId:     message.DatacenterId,
		NumSessions:      message.NumSessions,
		StartTime:        message.StartTime,
	}

	serverInserter.Insert(&serverData)
}

// -------------------------------------------------------------------------------

func ProcessNearRelayUpdateMessages(service *common.Service, redisHostname string, name string, nearRelayInserter *portal.NearRelayInserter) {

	streamName := strings.ReplaceAll(name, " ", "_")
	consumerGroup := streamName

	config := common.RedisStreamsConfig{
		RedisHostname: redisHostname,
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
		Timestamp:           uint64(time.Now().Unix()),
		NumNearRelays:       message.NumNearRelays,
		NearRelayId:         message.NearRelayId,
		NearRelayRTT:        message.NearRelayRTT,
		NearRelayJitter:     message.NearRelayJitter,
		NearRelayPacketLoss: message.NearRelayPacketLoss,
	}

	nearRelayInserter.Insert(sessionId, &nearRelayData)
}

// -------------------------------------------------------------------------------

func ProcessRelayUpdateMessages(service *common.Service, redisHostname string, name string, relayInserter *portal.RelayInserter) {

	streamName := strings.ReplaceAll(name, " ", "_")
	consumerGroup := streamName

	config := common.RedisStreamsConfig{
		RedisHostname: redisHostname,
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
				ProcessRelayUpdate(messageData, relayInserter)
			}
		}
	}()
}

func ProcessRelayUpdate(messageData []byte, relayInserter *portal.RelayInserter) {

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

	relaySample := portal.RelaySample{
		Timestamp:                 message.Timestamp,
		NumSessions:               message.SessionCount,
		EnvelopeBandwidthUpKbps:   message.EnvelopeBandwidthUpKbps,
		EnvelopeBandwidthDownKbps: message.EnvelopeBandwidthDownKbps,
		PacketsSentPerSecond:      message.PacketsSentPerSecond,
		PacketsReceivedPerSecond:  message.PacketsReceivedPerSecond,
		BandwidthSentKbps:         message.BandwidthSentKbps,
		BandwidthReceivedKbps:     message.BandwidthReceivedKbps,
		NearPingsPerSecond:        message.NearPingsPerSecond,
		RelayPingsPerSecond:       message.RelayPingsPerSecond,
		RelayFlags:                message.RelayFlags,
		NumRoutable:               message.NumRoutable,
		NumUnroutable:             message.NumUnroutable,
		CurrentTime:               message.CurrentTime,
	}

	relayInserter.Insert(&relayData, &relaySample)
}

// -------------------------------------------------------------------------------
