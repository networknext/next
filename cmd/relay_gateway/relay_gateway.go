package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/packets"
)

var redisHostname string
var redisPassword string
var redisPubsubChannelName string
var relayUpdateBatchSize int
var relayUpdateBatchDuration time.Duration
var relayUpdateChannelSize int

var producer *common.RedisPubsubProducer

func main() {

	service := common.CreateService("relay_gateway")

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")
	redisPubsubChannelName = envvar.GetString("REDIS_PUBSUB_CHANNEL_NAME", "relay_updates")
	relayUpdateBatchSize = envvar.GetInt("RELAY_UPDATE_BATCH_SIZE", 100)
	relayUpdateBatchDuration = envvar.GetDuration("RELAY_UPDATE_BATCH_DURATION", 1000*time.Millisecond)
	relayUpdateChannelSize = envvar.GetInt("RELAY_UPDATE_CHANNEL_SIZE", 10*1024)

	core.Log("redis hostname: %s", redisHostname)
	core.Log("redis password: %s", redisPassword)
	core.Log("redis pubsub channel name: %s", redisPubsubChannelName)
	core.Log("relay update batch size: %d", relayUpdateBatchSize)
	core.Log("relay update batch duration: %v", relayUpdateBatchDuration)
	core.Log("relay update channel size: %d", relayUpdateChannelSize)

	producer = CreatePubsubProducer(service)

	service.UpdateMagic()

	service.LoadDatabase()

	service.StartWebServer()

	service.Router.HandleFunc("/relay_update", RelayUpdateHandler(GetRelayData(service), GetMagicValues(service))).Methods("POST")

	service.WaitForShutdown()
}

func RelayUpdateHandler(getRelayData func() *common.RelayData, getMagicValues func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte)) func(writer http.ResponseWriter, request *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		startTime := time.Now()

		defer func() {
			duration := time.Since(startTime)
			if duration.Milliseconds() > 1000 {
				core.Warn("long relay update: %s", duration.String())
			}
		}()

		if request.Header.Get("Content-Type") != "application/octet-stream" {
			core.Error("[%s] unsupported content type", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			core.Error("[%s] could not read request body: %v", request.RemoteAddr, err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		defer request.Body.Close()

		var relayUpdateRequest packets.RelayUpdateRequestPacket

		err = relayUpdateRequest.Peek(body)
		if err != nil {
			core.Debug("[%s] could not peek relay update request", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		if relayUpdateRequest.Version != packets.VersionNumberRelayUpdateRequest {
			core.Debug("[%s] version mismatch", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		currentTimestamp := uint64(startTime.Unix())

		if relayUpdateRequest.Timestamp < currentTimestamp - 10 {
			core.Debug("[%s] relay update request is too old", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		if relayUpdateRequest.Timestamp > currentTimestamp + 10 {
			core.Debug("[%s] relay update request is in the future", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// todo: track per-address in a hash w. expiry 60 seconds, if the timestamp has already been received. if it has, then drop the packet here so people DDoSing us with replayed packets can't fill the redis queue

		relayData := getRelayData()

		relayId := common.RelayId(relayUpdateRequest.Address.String())

		relay, ok := relayData.RelayHash[relayId]
		if !ok {
			core.Debug("[%s] unknown relay %x", request.RemoteAddr, relayId)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// todo: verify packet signature

		// ...

		// relay update accepted

		relayName := relay.Name

		core.Debug("[%s] received update for %s [%x]", request.RemoteAddr, relayName, relayId)

		var responsePacket packets.RelayUpdateResponsePacket

		responsePacket.Version = packets.VersionNumberRelayUpdateResponse
		responsePacket.Timestamp = uint64(time.Now().Unix())
		responsePacket.TargetVersion = relay.Version

		index := 0

		for i := range relayData.RelayIds {

			if relayData.RelayIds[i] == relayId {
				continue
			}

			address := relayData.RelayArray[i].PublicAddress

			internal := uint8(0)
			if relay.Seller.Id == relayData.RelaySellerIds[i] && relayData.RelayArray[i].HasInternalAddress && relay.HasInternalAddress {
				address = relayData.RelayArray[i].InternalAddress
				internal = 1
			}

			responsePacket.RelayId[index] = relayData.RelayIds[i]
			responsePacket.RelayAddress[index] = address
			responsePacket.RelayInternal[index] = internal

			index++
		}

		responsePacket.NumRelays = uint32(index)

		responsePacket.UpcomingMagic, responsePacket.CurrentMagic, responsePacket.PreviousMagic = getMagicValues()

		// send the response packet back to the relay

		responseData := make([]byte, 1024*1024) // todo: would be better to tightly bound this response

		responseData = responsePacket.Write(responseData)

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))

		writer.Write(responseData)

		// forward the relay update to the relay backend

		producer.MessageChannel <- body
	}
}

func GetRelayData(service *common.Service) func() *common.RelayData {
	return func() *common.RelayData {
		return service.RelayData()
	}
}

func GetMagicValues(service *common.Service) func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte) {
	return func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte) {
		return service.GetMagicValues()
	}
}

func CreatePubsubProducer(service *common.Service) *common.RedisPubsubProducer {

	config := common.RedisPubsubConfig{}

	config.RedisHostname = redisHostname
	config.RedisPassword = redisPassword
	config.PubsubChannelName = redisPubsubChannelName
	config.BatchSize = relayUpdateBatchSize
	config.BatchDuration = relayUpdateBatchDuration
	config.MessageChannelSize = relayUpdateChannelSize

	var err error
	producer, err = common.CreateRedisPubsubProducer(service.Context, config)
	if err != nil {
		core.Error("could not create redis pubsub producer")
		os.Exit(1)
	}

	return producer
}
