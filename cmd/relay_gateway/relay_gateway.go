package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	
	"github.com/networknext/backend/modules-old/crypto"
	"github.com/networknext/backend/modules-old/routing"
	"github.com/networknext/backend/modules-old/transport"
)

var redisHostname string
var redisPassword string
var redisPubsubChannelName string
var relayUpdateBatchSize int
var relayUpdateBatchDuration time.Duration
var relayUpdateChannelSize int

var producer *common.RedisPubsubProducer

func main() {

	service := common.CreateService("relay_gateway_new")

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")
	redisPubsubChannelName = envvar.GetString("REDIS_PUBSUB_CHANNEL_NAME", "relay_updates")
	relayUpdateBatchSize = envvar.GetInt("RELAY_UPDATE_BATCH_SIZE", 100)
	relayUpdateBatchDuration = envvar.GetDuration("RELAY_UPDATE_BATCH_DURATION", 100*time.Millisecond)
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

func GetRelaysToPing(id uint64, relay *routing.Relay, relayArray []routing.Relay) []routing.RelayPingData {

	sellerName := relay.Seller.Name

	relaysToPing := make([]routing.RelayPingData, 0, len(relayArray)-1)

	for i := range relayArray {

		if relayArray[i].ID == id {
			continue
		}

		var address string
		if sellerName == relayArray[i].Seller.Name && relayArray[i].InternalAddr.String() != ":0" {
			address = relayArray[i].InternalAddr.String()
		} else {
			address = relayArray[i].Addr.String()
		}

		relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(relayArray[i].ID), Address: address})
	}

	return relaysToPing
}

func WriteRelayUpdateResponse(getMagicValues func() ([]byte, []byte, []byte), relay *routing.Relay, relayUpdateRequest *transport.RelayUpdateRequest, relaysToPing []routing.RelayPingData) ([]byte, error) {

	response := transport.RelayUpdateResponse{}

	response.Version = transport.VersionNumberUpdateResponse
	response.Timestamp = time.Now().Unix()

	for i := range relaysToPing {
		response.RelaysToPing = append(response.RelaysToPing, routing.RelayPingData{
			ID:      relaysToPing[i].ID,
			Address: relaysToPing[i].Address,
		})
	}

	response.TargetVersion = relay.Version

	upcomingMagic, currentMagic, previousMagic := getMagicValues()

	response.UpcomingMagic = upcomingMagic
	response.CurrentMagic = currentMagic
	response.PreviousMagic = previousMagic

	responseData, err := response.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return responseData, nil
}

func RelayUpdateHandler(getRelayData func() *common.RelayData, getMagicValues func() ([]byte, []byte, []byte)) func(writer http.ResponseWriter, request *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		startTime := time.Now()

		defer func() {
			duration := time.Since(startTime)
			if duration.Milliseconds() > 1000 {
				core.Warn("long relay update: %s", duration.String())
			}
		}()

		core.Debug("%s - relay update", request.RemoteAddr)

		if request.Header.Get("Content-Type") != "application/octet-stream" {
			core.Error("%s - relay update unsupported content type", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			core.Error("%s - relay update could not read request body: %v", request.RemoteAddr, err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		defer request.Body.Close()

		var relayUpdateRequest transport.RelayUpdateRequest
		err = relayUpdateRequest.UnmarshalBinary(body)
		if err != nil {
			core.Error("%s - relay update could not read relay update request", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		if relayUpdateRequest.Version > transport.VersionNumberUpdateRequest {
			core.Error("%s - relay update version mismatch: %d > %d", request.RemoteAddr, relayUpdateRequest.Version, transport.VersionNumberUpdateRequest)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		if len(relayUpdateRequest.PingStats) > core.MaxNearRelays {
			core.Error("%s - error: relay update too many relays in ping stats: %d > %d", request.RemoteAddr, relayUpdateRequest.PingStats, core.MaxNearRelays)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		relayData := getRelayData()

		id := crypto.HashID(relayUpdateRequest.Address.String())

		relay, ok := relayData.RelayHash[id]
		if !ok {
			core.Error("%s - could not find relay: %s [%x]", request.RemoteAddr, relayUpdateRequest.Address.String(), id)
			writer.WriteHeader(http.StatusNotFound) // 404
			return
		}

		producer.MessageChannel <- body

		relaysToPing := GetRelaysToPing(id, &relay, relayData.RelayArray)

		response, err := WriteRelayUpdateResponse(getMagicValues, &relay, &relayUpdateRequest, relaysToPing)
		if err != nil {
			core.Error("%s - relay update could not write relay update response: %v", request.RemoteAddr, err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))

		writer.Write(response)
	}
}

func GetRelayData(service *common.Service) func() *common.RelayData {
	return func() *common.RelayData {
		return service.RelayData()
	}
}

func GetMagicValues(service *common.Service) func() ([]byte, []byte, []byte) {
	return func() ([]byte, []byte, []byte) {
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
