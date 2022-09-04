package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"
)

func main() {

	service := common.CreateService("relay_gateway_new")

	service.UpdateMagic()

	service.LoadDatabase()

	service.StartWebServer()

	relayUpdateChannel := CreateRelayUpdateChannel()

	service.Router.HandleFunc("/relay_update", RelayUpdateHandler(GetRelayData(service), GetMagicValues(service), relayUpdateChannel)).Methods("POST")

	ProcessRelayUpdates(service.Context, relayUpdateChannel)

	service.WaitForShutdown()
}

func ProcessRelayUpdates(ctx context.Context, relayUpdateChannel chan []byte) {
	go func() {
		select {
		case message := <-relayUpdateChannel:
			fmt.Printf("processed relay update")
			// todo: send message to redis pubsub producer
			_ = message
		case <-ctx.Done():
			return
		}
	}()
}

func GetRelaysToPing(id uint64, relay *routing.Relay, relayArray []routing.Relay) ([]routing.RelayPingData) {

	// todo: be absolutely fucking sure that the set of relays being pinged is the set of ENABLED relays in the database !!!
	// otherwise we are pinging a bunch of relays that have been deleted for a year, and don't exist anymore
	
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

func RelayUpdateHandler(getRelayData func() (map[uint64]routing.Relay, []routing.Relay), getMagicValues func() ([]byte, []byte, []byte), relayUpdateChannel chan []byte) func(writer http.ResponseWriter, request *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		startTime := time.Now()

		defer func() {
			duration := time.Since(startTime)
			if duration.Milliseconds() > 100 {
				core.Error("long relay update: %s", duration.String())
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

		if len(relayUpdateRequest.PingStats) > transport.MaxRelays {
			core.Error("%s - error: relay update too many relays in ping stats: %d > %d", request.RemoteAddr, relayUpdateRequest.PingStats, transport.MaxRelays)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		relayHash, relayArray := getRelayData()

		id := crypto.HashID(relayUpdateRequest.Address.String())

		relay, ok := relayHash[id]
		if !ok {
			core.Error("%s - could not find relay: %s [%x]", request.RemoteAddr, relayUpdateRequest.Address.String(), id)
			writer.WriteHeader(http.StatusNotFound) // 404
			return
		}

		relayUpdateChannel <- body

		relaysToPing := GetRelaysToPing(id, &relay, relayArray)

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

func GetRelayData(service *common.Service) func() (map[uint64]routing.Relay, []routing.Relay) {
	return func() (map[uint64]routing.Relay, []routing.Relay) {
		_, _, relayHash, relayArray := service.DatabaseAll()
		return relayHash, relayArray
	}
}

func GetMagicValues(service *common.Service) func() ([]byte, []byte, []byte) {
	return func() ([]byte, []byte, []byte) {
		return service.GetMagicValues()
	}
}

func CreateRelayUpdateChannel() chan []byte {
	relayUpdateChannelSize := envvar.GetInt("RELAY_UPDATE_CHANNEL_SIZE", 10*1024)
	core.Debug("relay update channel size: %d", relayUpdateChannelSize)
	return make(chan []byte, relayUpdateChannelSize)
}
