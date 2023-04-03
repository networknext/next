package main

import (
	"os"
	"strings"
	"time"

	"github.com/networknext/accelerate/modules/common"
	"github.com/networknext/accelerate/modules/core"
	"github.com/networknext/accelerate/modules/envvar"
	"github.com/networknext/accelerate/modules/messages"
	"github.com/networknext/accelerate/modules/portal"
)

var redisHostname string
var redisPassword string

var mapInstance *portal.Map

func main() {

	numMapUpdateThreads := envvar.GetInt("NUM_MAP_UPDATE_THREADS", 1)

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")

	service := common.CreateService("map_cruncher")

	core.Debug("num map update threads: %d", numMapUpdateThreads)
	core.Debug("redis hostname: %s", redisHostname)

	for i := 0; i < numMapUpdateThreads; i++ {
		ProcessMessages[*messages.PortalMapUpdateMessage](service, "map update", i, ProcessMapUpdate)
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

func ProcessMessages[T messages.Message](service *common.Service, name string, threadNumber int, process func([]byte, int)) {

	channelName := strings.ReplaceAll(name, " ", "_")

	config := common.RedisPubsubConfig{}

	config.RedisHostname = redisHostname
	config.RedisPassword = redisPassword
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
				process(messageData, threadNumber)
			}
		}
	}()
}

// -------------------------------------------------------------------------------

func ProcessMapUpdate(messageData []byte, threadNumber int) {

	message := messages.PortalMapUpdateMessage{}
	err := message.Read(messageData)
	if err != nil {
		core.Error("could not read map update message: %v", err)
		return
	}

	core.Debug("received map update message on thread %d", threadNumber)

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
