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

var redisServerBackendCluster []string
var redisServerBackendHostname string

var mapInstance *portal.Map

func main() {

	redisServerBackendCluster = envvar.GetStringArray("REDIS_SERVER_BACKEND_CLUSTER", []string{})
	redisServerBackendHostname = envvar.GetString("REDIS_SERVER_BACKEND_HOSTNAME", "127.0.0.1:6379")

	reps := envvar.GetInt("REPS", 1)

	service := common.CreateService("map_cruncher")

	core.Debug("redis server backend cluster: %v", redisServerBackendCluster)
	core.Debug("redis server backend hostname: %s", redisServerBackendHostname)

	for i := 0; i < reps; i++ {
		ProcessMessages[*messages.PortalMapUpdateMessage](service, "map update", ProcessMapUpdate)
	}

	mapInstance = portal.CreateMap(service.Context)

	WriteMapDataToRedis(service)

	service.LeaderElection()

	service.StartWebServer()

	service.WaitForShutdown()
}

// -------------------------------------------------------------------------------

func WriteMapDataToRedis(service *common.Service) {

	go func() {

		previousSize := 0

		for {

			time.Sleep(time.Second)

			entries := make([]portal.CellEntry, 0, previousSize)

			for i := 0; i < portal.NumCells; i++ {
				for {
					var output *portal.CellOutput
					select {
					case output = <-mapInstance.Cells[i].OutputChan:
					default:
					}
					if output == nil {
						break
					}
					entries = append(entries, output.Entries...)
				}
			}

			data := portal.WriteMapData(entries)

			if service.IsLeader() {
				service.Store("map_data", data)
			}

			previousSize = len(entries)
		}
	}()
}

// -------------------------------------------------------------------------------

func ProcessMessages[T messages.Message](service *common.Service, name string, process func([]byte)) {

	channelName := strings.ReplaceAll(name, " ", "_")

	config := common.RedisPubsubConfig{}

	config.RedisCluster = redisServerBackendCluster
	config.RedisHostname = redisServerBackendHostname
	config.PubsubChannelName = channelName

	consumer, err := common.CreateRedisPubsubConsumer(service.Context, config)

	if err != nil {
		core.Error("could not create redis pubsub consumer for map updates")
		os.Exit(1)
	}

	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case messageData := <-consumer.MessageChannel:
				process(messageData)
			}
		}
	}()
}

// -------------------------------------------------------------------------------

func ProcessMapUpdate(messageData []byte) {

	message := messages.PortalMapUpdateMessage{}
	err := message.Read(messageData)
	if err != nil {
		core.Error("could not read map update message: %v", err)
		return
	}

	core.Debug("received map update message")

	update := portal.CellUpdate{}
	update.SessionId = message.SessionId
	update.Latitude = message.Latitude
	update.Longitude = message.Longitude
	update.Next = message.Next

	cellIndex := portal.GetCellIndex(update.Latitude, update.Longitude)
	if cellIndex < 0 {
		core.Error("bad map update lat/long")
		return
	}

	mapInstance.Cells[cellIndex].UpdateChan <- &update
}

// -------------------------------------------------------------------------------
