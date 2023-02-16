package main

import (
	"os"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/messages"
	"github.com/networknext/backend/modules/envvar"
)

var redisHostname string
var redisPassword string

func main() {

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")

	core.Log("redis hostname: %s", redisHostname)
	core.Log("redis password: %s", redisPassword)

	service := common.CreateService("portal_cruncher")

	service.StartWebServer()

	service.LeaderElection(true)

	ProcessSessionUpdateMessages(service)

	service.WaitForShutdown()
}

func ProcessSessionUpdateMessages(service *common.Service) {

	name := "session update"
	streamName := "session_update"
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

				core.Debug("received %s message", name)

				message := messages.PortalSessionUpdateMessage{}
				err := message.Read(messageData)
				if err != nil {
					core.Error("could not read %s message: %v", name)
					break
				}

				// todo: process the message
				_ = message
			}
		}
	}()
}
