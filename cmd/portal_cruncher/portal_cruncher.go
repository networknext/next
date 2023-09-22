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
)

var service *common.Service

var redisPortalHostname string
var redisRelayBackendHostname string
var redisServerBackendHostname string
var sessionCruncherURL string

func main() {

	// todo: redis cluster config
	redisCluster := []string{"127.0.0.1:7000", "127.0.0.1:7001", "127.0.0.1:7002", "127.0.0.1:7003", "127.0.0.1:7004", "127.0.0.1:7005"}

	redisServerBackendHostname = envvar.GetString("REDIS_SERVER_BACKEND_HOSTNAME", "127.0.0.1:6379")
	redisRelayBackendHostname = envvar.GetString("REDIS_RELAY_BACKEND_HOSTNAME", "127.0.0.1:6379")
	sessionCruncherURL = envvar.GetString("SESSION_CRUNCHER_URL", "http://127.0.0.1:40200/session_batch")

	sessionInsertBatchSize := envvar.GetInt("SESSION_INSERT_BATCH_SIZE", 1000)
	serverInsertBatchSize := envvar.GetInt("SERVER_INSERT_BATCH_SIZE", 1000)
	relayInsertBatchSize := envvar.GetInt("RELAY_INSERT_BATCH_SIZE", 1000)
	nearRelayInsertBatchSize := envvar.GetInt("NEAR_RELAY_INSERT_BATCH_SIZE", 1000)

	reps := envvar.GetInt("REPS", 1)

	service = common.CreateService("portal_cruncher")

	core.Debug("redis cluster: %v", redisCluster)
	core.Debug("redis relay backend hostname: %s", redisRelayBackendHostname)
	core.Debug("redis server backend hostname: %s", redisServerBackendHostname)
	core.Debug("session cruncher url: %s", sessionCruncherURL)

	core.Debug("session insert batch size: %d", sessionInsertBatchSize)
	core.Debug("server insert batch size: %d", serverInsertBatchSize)
	core.Debug("relay insert batch size: %d", relayInsertBatchSize)
	core.Debug("near relay insert batch size: %d", nearRelayInsertBatchSize)

	if !service.Local {
		service.LoadIP2Location()
	}

	for j := 0; j < reps; j++ {
		ProcessSessionUpdateMessages(service, redisServerBackendHostname, redisCluster, sessionInsertBatchSize)
		ProcessServerUpdateMessages(service, redisServerBackendHostname, redisCluster, serverInsertBatchSize)
		ProcessNearRelayUpdateMessages(service, redisServerBackendHostname, redisCluster, nearRelayInsertBatchSize)
		ProcessRelayUpdateMessages(service, redisRelayBackendHostname, redisCluster, relayInsertBatchSize)
	}

	service.StartWebServer()

	service.WaitForShutdown()
}

// -------------------------------------------------------------------------------

func ProcessSessionUpdateMessages(service *common.Service, redisStreams string, redisCluster []string, batchSize int) {

	name := "session update"

	redisClient := common.CreateRedisClusterClient(redisCluster)

	sessionInserter := portal.CreateSessionInserter(service.Context, redisClient, sessionCruncherURL, batchSize)

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

	score := int32(0)
	if next {
	 	score = 500 - int32(message.NextRTT)
	} else {
	 	score = 500 + int32(message.DirectRTT)
	}

	if score < 0 {
		score = 0
	} else if score > 999 {
		score = 999
	}
	
	// todo: hack
	currentScore := uint32(score)
	previousScore := uint32(score)

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

	sessionInserter.Insert(service.Context, sessionId, userHash, next, currentScore, previousScore, &sessionData, &sliceData)
}

// -------------------------------------------------------------------------------

func ProcessServerUpdateMessages(service *common.Service, redisStreams string, redisCluster []string, batchSize int) {

	name := "server update"

	redisClient := common.CreateRedisClusterClient(redisCluster)

	serverInserter := portal.CreateServerInserter(redisClient, batchSize)

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
		BuyerId:          message.BuyerId,
		DatacenterId:     message.DatacenterId,
		NumSessions:      message.NumSessions,
		StartTime:        message.StartTime,
	}

	serverInserter.Insert(service.Context, &serverData)
}

// -------------------------------------------------------------------------------

func ProcessNearRelayUpdateMessages(service *common.Service, redisStreams string, redisCluster []string, batchSize int) {

	name := "near relay update"

	redisClient := common.CreateRedisClusterClient(redisCluster)

	nearRelayInserter := portal.CreateNearRelayInserter(redisClient, batchSize)

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

	nearRelayInserter.Insert(service.Context, sessionId, &nearRelayData)
}

// -------------------------------------------------------------------------------

func ProcessRelayUpdateMessages(service *common.Service, redisStreams string, redisCluster []string, batchSize int) {

	name := "relay update"

	redisClient := common.CreateRedisClusterClient(redisCluster)

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

	relayInserter.Insert(service.Context, &relayData)
}

// -------------------------------------------------------------------------------

/*
	// todo: this should be time series
	// relaySample := portal.RelaySample{
	// 	Timestamp:                 message.Timestamp,
	// 	NumSessions:               message.SessionCount,
	// 	EnvelopeBandwidthUpKbps:   message.EnvelopeBandwidthUpKbps,
	// 	EnvelopeBandwidthDownKbps: message.EnvelopeBandwidthDownKbps,
	// 	PacketsSentPerSecond:      message.PacketsSentPerSecond,
	// 	PacketsReceivedPerSecond:  message.PacketsReceivedPerSecond,
	// 	BandwidthSentKbps:         message.BandwidthSentKbps,
	// 	BandwidthReceivedKbps:     message.BandwidthReceivedKbps,
	// 	NearPingsPerSecond:        message.NearPingsPerSecond,
	// 	RelayPingsPerSecond:       message.RelayPingsPerSecond,
	// 	RelayFlags:                message.RelayFlags,
	// 	NumRoutable:               message.NumRoutable,
	// 	NumUnroutable:             message.NumUnroutable,
	// 	CurrentTime:               message.CurrentTime,
	// }
*/
