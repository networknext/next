package main

import (
	"os"
	"strings"
	"time"

	"github.com/networknext/accelerate/modules/common"
	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/core"
	"github.com/networknext/accelerate/modules/envvar"
	"github.com/networknext/accelerate/modules/messages"
	"github.com/networknext/accelerate/modules/portal"

	"github.com/gomodule/redigo/redis"
)

var redisHostname string
var redisPassword string

var pool *redis.Pool

var sessionInserter []*portal.SessionInserter
var serverInserter []*portal.ServerInserter
var relayInserter []*portal.RelayInserter
var nearRelayInserter []*portal.NearRelayInserter

func main() {

	numSessionUpdateThreads := envvar.GetInt("NUM_SESSION_UPDATE_THREADS", 1)
	numServerUpdateThreads := envvar.GetInt("NUM_SERVER_UPDATE_THREADS", 1)
	numRelayUpdateThreads := envvar.GetInt("NUM_RELAY_UPDATE_THREADS", 1)
	numNearRelayUpdateThreads := envvar.GetInt("NUM_NEAR_RELAY_UPDATE_THREADS", 1)

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")
	redisPoolActive := envvar.GetInt("REDIS_POOL_ACTIVE", 1000)
	redisPoolIdle := envvar.GetInt("REDIS_POOL_IDLE", 10000)

	sessionInsertBatchSize := envvar.GetInt("SESSION_INSERT_BATCH_SIZE", 1000)
	serverInsertBatchSize := envvar.GetInt("SERVER_INSERT_BATCH_SIZE", 1000)
	relayInsertBatchSize := envvar.GetInt("RELAY_INSERT_BATCH_SIZE", 1000)
	nearRelayInsertBatchSize := envvar.GetInt("NEAR_RELAY_INSERT_BATCH_SIZE", 1000)

	service := common.CreateService("portal_cruncher")

	core.Debug("num session update threads: %d", numSessionUpdateThreads)
	core.Debug("num server update threads: %d", numServerUpdateThreads)
	core.Debug("num relay update threads: %d", numRelayUpdateThreads)
	core.Debug("num near relay update threads: %d", numNearRelayUpdateThreads)

	core.Debug("redis hostname: %s", redisHostname)
	core.Debug("redis pool active: %d", redisPoolActive)
	core.Debug("redis pool idle: %d", redisPoolIdle)

	core.Debug("session insert batch size: %d", sessionInsertBatchSize)
	core.Debug("server insert batch size: %d", serverInsertBatchSize)
	core.Debug("relay insert batch size: %d", relayInsertBatchSize)
	core.Debug("near relay insert batch size: %d", nearRelayInsertBatchSize)

	pool = common.CreateRedisPool(redisHostname, redisPoolActive, redisPoolIdle)

	sessionInserter = make([]*portal.SessionInserter, numSessionUpdateThreads)
	serverInserter = make([]*portal.ServerInserter, numServerUpdateThreads)
	relayInserter = make([]*portal.RelayInserter, numRelayUpdateThreads)
	nearRelayInserter = make([]*portal.NearRelayInserter, numNearRelayUpdateThreads)

	for i := 0; i < numSessionUpdateThreads; i++ {
		sessionInserter[i] = portal.CreateSessionInserter(pool, sessionInsertBatchSize)
		ProcessMessages[*messages.PortalSessionUpdateMessage](service, "session update", i, ProcessSessionUpdate)
	}

	for i := 0; i < numServerUpdateThreads; i++ {
		serverInserter[i] = portal.CreateServerInserter(pool, serverInsertBatchSize)
		ProcessMessages[*messages.PortalServerUpdateMessage](service, "server update", i, ProcessServerUpdate)
	}

	for i := 0; i < numRelayUpdateThreads; i++ {
		relayInserter[i] = portal.CreateRelayInserter(pool, relayInsertBatchSize)
		ProcessMessages[*messages.PortalRelayUpdateMessage](service, "relay update", i, ProcessRelayUpdate)
	}

	for i := 0; i < numNearRelayUpdateThreads; i++ {
		nearRelayInserter[i] = portal.CreateNearRelayInserter(pool, nearRelayInsertBatchSize)
		ProcessMessages[*messages.PortalNearRelayUpdateMessage](service, "near relay update", i, ProcessNearRelayUpdate)
	}

	service.StartWebServer()

	service.WaitForShutdown()
}

// -------------------------------------------------------------------------------

func ProcessMessages[T messages.Message](service *common.Service, name string, threadNumber int, process func([]byte, int)) {

	streamName := strings.ReplaceAll(name, " ", "_")
	consumerGroup := streamName

	config := common.RedisStreamsConfig{
		RedisHostname: redisHostname,
		RedisPassword: redisPassword,
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
				process(messageData, threadNumber)
			}
		}
	}()
}

// -------------------------------------------------------------------------------

func ProcessSessionUpdate(messageData []byte, threadNumber int) {

	message := messages.PortalSessionUpdateMessage{}
	err := message.Read(messageData)
	if err != nil {
		core.Error("could not read session update message: %v", err)
		return
	}

	core.Debug("received session update message on thread %d", threadNumber)

	sessionId := message.SessionId

	next := (message.SessionFlags & constants.SessionFlags_Next) != 0

	score := uint32(0)
	if next {
		score = uint32(message.NextRTT)
	} else {
		score = 10000 - uint32(message.DirectRTT)
	}

	// todo: look up ISP name from message.ClientAddress
	isp := "Comcast Internet Company, LLC"

	sessionData := portal.SessionData{
		SessionId:      message.SessionId,
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
	}

	sessionInserter[threadNumber].Insert(sessionId, score, next, &sessionData, &sliceData)
}

// -------------------------------------------------------------------------------

func ProcessServerUpdate(messageData []byte, threadNumber int) {

	message := messages.PortalServerUpdateMessage{}
	err := message.Read(messageData)
	if err != nil {
		core.Error("could not read server update message: %v", err)
		return
	}

	core.Debug("received server update message on thread %d", threadNumber)

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

	serverInserter[threadNumber].Insert(&serverData)
}

// -------------------------------------------------------------------------------

func ProcessRelayUpdate(messageData []byte, threadNumber int) {

	message := messages.PortalRelayUpdateMessage{}
	err := message.Read(messageData)
	if err != nil {
		core.Error("could not read relay update message: %v", err)
		return
	}

	core.Debug("received relay update message on thread %d", threadNumber)

	relayData := portal.RelayData{
		RelayId:      message.RelayId,
		RelayAddress: message.RelayAddress.String(),
		NumSessions:  message.SessionCount,
		MaxSessions:  message.MaxSessions,
		StartTime:    message.StartTime,
		RelayFlags:   message.RelayFlags,
		Version:      message.RelayVersion,
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

	relayInserter[threadNumber].Insert(&relayData, &relaySample)
}

// -------------------------------------------------------------------------------

func ProcessNearRelayUpdate(messageData []byte, threadNumber int) {

	message := messages.PortalNearRelayUpdateMessage{}
	err := message.Read(messageData)
	if err != nil {
		core.Error("could not read near relay update message: %v", err)
		return
	}

	core.Debug("received near relay update message on thread %d", threadNumber)

	sessionId := message.SessionId

	nearRelayData := portal.NearRelayData{
		Timestamp:           uint64(time.Now().Unix()),
		NumNearRelays:       message.NumNearRelays,
		NearRelayId:         message.NearRelayId,
		NearRelayRTT:        message.NearRelayRTT,
		NearRelayJitter:     message.NearRelayJitter,
		NearRelayPacketLoss: message.NearRelayPacketLoss,
	}

	nearRelayInserter[threadNumber].Insert(sessionId, &nearRelayData)
}

// -------------------------------------------------------------------------------
