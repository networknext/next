package main

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

	numMapUpdateThreads := envvar.GetInt("NUM_MAP_UPDATE_THREADS", 1)

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")

	core.Log("num map update threads: %d", numMapUpdateThreads)
	core.Log("redis hostname: %s", redisHostname)
	core.Log("redis password: %s", redisPassword)

	service := common.CreateService("map_cruncher")

	// todo: process map messages
	_ = numMapUpdateThreads
	/*
		for i := 0; i < numSessionUpdateThreads; i++ {
			ProcessMessages[*messages.PortalSessionUpdateMessage](service, "session update", i, ProcessSessionUpdate)
		}
	*/

	// todo: serve up map data from leader

	service.LeaderElection()

	service.StartWebServer()

	service.WaitForShutdown()
}

// -------------------------------------------------------------------------------

func ProcessMessages[T messages.Message](service *common.Service, name string, threadNumber int, process func([]byte, int)) {

	streamName := strings.ReplaceAll(name, " ", "_")
	consumerGroup := streamName

	// todo: it must be redis pubsub actually. all map crunchers must receive the same stream of map update messages

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

func ProcessMapUpdate(messageData []byte, threadNumber int) {

	// todo
	/*
		message := messages.PortalRelayUpdateMessage{}
		err := message.Read(messageData)
		if err != nil {
			core.Error("could not read relay update message: %v", err)
			return
		}

		core.Debug("received relay update message on thread %d", threadNumber)

		// ...

		_ = message
	*/
}

// -------------------------------------------------------------------------------
