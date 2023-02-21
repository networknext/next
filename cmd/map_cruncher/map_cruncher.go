package main

import (
	"fmt"
)

func main() {
	fmt.Printf("map_cruncher")
}

/*
import (
	"os"
	"strings"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/messages"
)

var redisHostname string
var redisPassword string

func main() {

	numSessionUpdateThreads := envvar.GetInt("NUM_SESSION_UPDATE_THREADS", 1)
	numServerUpdateThreads := envvar.GetInt("NUM_SERVER_UPDATE_THREADS", 1)
	numRelayUpdateThreads := envvar.GetInt("NUM_RELAY_UPDATE_THREADS", 1)

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")

	core.Log("num session update threads: %d", numSessionUpdateThreads)
	core.Log("num server update threads: %d", numServerUpdateThreads)
	core.Log("num relay update threads: %d", numRelayUpdateThreads)
	core.Log("redis hostname: %s", redisHostname)
	core.Log("redis password: %s", redisPassword)

	service := common.CreateService("portal_cruncher")

	for i := 0; i < numSessionUpdateThreads; i++ {
		ProcessMessages[*messages.PortalSessionUpdateMessage](service, "session update", i, ProcessSessionUpdate)
	}

	for i := 0; i < numServerUpdateThreads; i++ {
		ProcessMessages[*messages.PortalServerUpdateMessage](service, "server update", i, ProcessServerUpdate)
	}

	for i := 0; i < numRelayUpdateThreads; i++ {
		ProcessMessages[*messages.PortalRelayUpdateMessage](service, "relay update", i, ProcessRelayUpdate)
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

	// ...

	_ = message
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

	// ...

	_ = message
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

	// ...

	_ = message
}

// -------------------------------------------------------------------------------
*/