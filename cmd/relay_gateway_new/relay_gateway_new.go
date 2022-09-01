package main

import (
	"net/http"
	"time"
	"io/ioutil"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
)

var relayUpdateChannel chan []byte

func relayUpdateHandler(writer http.ResponseWriter, request *http.Request) {

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

	// ...

	_ = body

	/*
	// Unmarshal the request

	// Check if the version number is valid
	if relayUpdateRequest.Version > VersionNumberUpdateRequest {
		params.Metrics.ErrorMetrics.InvalidVersion.Add(1)
		core.Error("%s - error: relay update version mismatch: %d > %d", request.RemoteAddr, relayUpdateRequest.Version, VersionNumberUpdateRequest)
		writer.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// Check if we have too many relays in the ping stats
	if len(relayUpdateRequest.PingStats) > MaxRelays {
		params.Metrics.ErrorMetrics.ExceedMaxRelays.Add(1)
		core.Error("%s - error: relay update too many relays in ping stats: %d > %d", request.RemoteAddr, relayUpdateRequest.PingStats, MaxRelays)
		writer.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// Check if relay exists
	relayArray, relayHash := params.GetRelayData()

	id := crypto.HashID(relayUpdateRequest.Address.String())

	relay, ok := relayHash[id]
	if !ok {
		params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
		core.Error("%s - error: could not find relay: %s [%x]", request.RemoteAddr, relayUpdateRequest.Address.String(), id)
		writer.WriteHeader(http.StatusNotFound) // 404
		return
	}

	// todo: bring back crypto check

	// Request is valid, insert the body into the channel
	params.RequestChan <- body

	// Get relays to ping
	relaysToPing := make([]routing.RelayPingData, 0)
	sellerName := relay.Seller.Name

	var internalAddrs []string

	for i := range relayArray {
		if relayArray[i].ID == id {
			continue
		}

		var address string
		if sellerName == relayArray[i].Seller.Name && relayArray[i].InternalAddr.String() != ":0" {
			// If the relay is under the same seller, prefer the internal address
			address = relayArray[i].InternalAddr.String()
			// Store the internal address to send to SDK5 supported relays
			internalAddrs = append(internalAddrs, address)
		} else {
			// Use the relay's external address
			address = relayArray[i].Addr.String()
		}

		relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(relayArray[i].ID), Address: address})
	}

	// Build and write the response
	var responseData []byte
	response := RelayUpdateResponse{}

	relayVersion, err := routing.ParseRelayVersion(relayUpdateRequest.RelayVersion)
	if err != nil {
		core.Error("failed to parse relay version: %v", err)
		relayVersion = routing.RelayVersion{}
	}

	response.Version = 0
	if relayVersion.AtLeast(routing.RelayVersion{2, 1, 0}) {
		// Relay version 2.1.0 started using update response version 1
		response.Version = VersionNumberUpdateResponse
	}

	for i := range relaysToPing {
		response.RelaysToPing = append(response.RelaysToPing, routing.RelayPingData{
			ID:      relaysToPing[i].ID,
			Address: relaysToPing[i].Address,
		})
	}
	response.Timestamp = time.Now().Unix()
	response.TargetVersion = relay.Version

	if response.Version >= 1 {
		response.UpcomingMagic, response.CurrentMagic, response.PreviousMagic = params.GetMagicData()

		response.InternalAddrs = internalAddrs
	}

	responseData, err = response.MarshalBinary()
	if err != nil {
		params.Metrics.ErrorMetrics.MarshalBinaryResponseFailure.Add(1)
		core.Error("%s - error: failed to write relay update response: %v", request.RemoteAddr, err)
		writer.WriteHeader(http.StatusInternalServerError) // 500
		return
	}

	writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
	writer.Write(responseData)
	*/
}

func main() {

	service := common.CreateService("relay_gateway_new")

	// register the relay update handler. this is where updates are posted to us from relays.

	service.Router.HandleFunc("/relay_update", relayUpdateHandler).Methods("POST")

	// create a relay update channel. the "relay_update" handler pushes updates onto this channel for processing

	relayUpdateChannelSize := envvar.GetInt("RELAY_GATEWAY_CHANNEL_SIZE", 10*1024)

	core.Debug("relay update channel size: %d", relayUpdateChannelSize)

	relayUpdateChannel = make(chan []byte, relayUpdateChannelSize)

	_ = relayUpdateChannel

	// run the service

	service.UpdateMagic()

	service.LoadDatabase()

	service.StartWebServer()

	service.WaitForShutdown()
}
