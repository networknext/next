package transport

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"

	"github.com/networknext/backend/modules/core"
)

const (
	InitRequestMagic = 0x9083708f
	MaxRelays        = 1024
)

/*
	GatewayRelayInitHandlerFunc() initializes all relays with version < 2.0.
	It does not perform any crypto checks and responds with an OK to get the relay
	to start making relay updates, where the primary work is done.

	NOTE: Relay init is deprecated. Remove this function once all relays have been upgraded to 2.0.x.
*/
func GatewayRelayInitHandlerFunc() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/octet-stream")
		responseData := make([]byte, 64)
		index := 0
		var nilVersion uint32 = 0
		var nilPubKey []byte = make([]byte, 32)

		encoding.WriteUint32(responseData, &index, nilVersion)
		encoding.WriteUint64(responseData, &index, uint64(time.Now().Unix()))
		encoding.WriteBytes(responseData, &index, nilPubKey, len(nilPubKey))
		responseData = responseData[:index]

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))

		writer.Write(responseData)
	}
}

type GatewayRelayUpdateHandlerConfig struct {
	Logger       log.Logger
	RequestChan  chan []byte
	Metrics      *metrics.RelayGatewayMetrics
	GetRelayData func() ([]routing.Relay, map[uint64]routing.Relay)
}

/*
	GatewayRelayUpdateHandlerFunc() receives relay updates. It's purpose is to verify the update
	and insert it into the update channel as quickly as possible so that the relay backends can
	receive the latest info and produce an accurate route matrix.

	Additionally, it responds to each relay with the set of relays to ping, which is derived from the database.
*/
func GatewayRelayUpdateHandlerFunc(params GatewayRelayUpdateHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.Metrics.HandlerMetrics.Duration.Set(float64(durationSince.Milliseconds()))
			params.Metrics.HandlerMetrics.Invocations.Add(1)
			if durationSince.Milliseconds() > 100 {
				params.Metrics.HandlerMetrics.LongDuration.Add(1)
			}
		}()

		core.Debug("%s - relay update", request.RemoteAddr)

		// Read the body from the request
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			params.Metrics.ErrorMetrics.ReadPacketFailure.Add(1)
			level.Error(params.Logger).Log("msg", "could not read packet", "err", err)
			core.Debug("%s - error: relay update could not read request body: %v", request.RemoteAddr, err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		defer request.Body.Close()

		// Ensure the content type is valid
		if request.Header.Get("Content-Type") != "application/octet-stream" {
			params.Metrics.ErrorMetrics.ContentTypeFailure.Add(1)
			level.Error(params.Logger).Log("err", fmt.Sprintf("%s - error: relay update unsupported content type", request.RemoteAddr))
			core.Debug("%s - error: relay update unsupported content type", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// Unmarshal the request
		var relayUpdateRequest RelayUpdateRequest
		err = relayUpdateRequest.UnmarshalBinary(body)
		if err != nil {
			params.Metrics.ErrorMetrics.UnmarshalFailure.Add(1)
			level.Error(params.Logger).Log("err", fmt.Sprintf("%s - error: relay update could not read request packet", request.RemoteAddr))
			core.Debug("%s - error: relay update could not read request packet", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// Check if the version number is valid
		if relayUpdateRequest.Version > VersionNumberUpdateRequest {
			params.Metrics.ErrorMetrics.InvalidVersion.Add(1)
			level.Error(params.Logger).Log("err", fmt.Sprintf("%s - error: relay update version mismatch: %d > %d", request.RemoteAddr, relayUpdateRequest.Version, VersionNumberUpdateRequest))
			core.Debug("%s - error: relay update version mismatch: %d > %d", request.RemoteAddr, relayUpdateRequest.Version, VersionNumberUpdateRequest)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// Check if we have too many relays in the ping stats
		if len(relayUpdateRequest.PingStats) > MaxRelays {
			params.Metrics.ErrorMetrics.ExceedMaxRelays.Add(1)
			level.Error(params.Logger).Log("err", fmt.Sprintf("%s - error: relay update too many relays in ping stats: %d > %d", request.RemoteAddr, len(relayUpdateRequest.PingStats), MaxRelays))
			core.Debug("%s - error: relay update too many relays in ping stats: %d > %d", request.RemoteAddr, relayUpdateRequest.PingStats, MaxRelays)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// Check if relay exists
		relayArray, relayHash := params.GetRelayData()

		id := crypto.HashID(relayUpdateRequest.Address.String())

		relay, ok := relayHash[id]
		if !ok {
			params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
			level.Error(params.Logger).Log("err", fmt.Sprintf("%s - error: could not find relay: %s [%x]", request.RemoteAddr, relayUpdateRequest.Address.String(), id))
			core.Debug("%s - error: could not find relay: %s [%x]", request.RemoteAddr, relayUpdateRequest.Address.String(), id)
			writer.WriteHeader(http.StatusNotFound) // 404
			return
		}
		// TODO: bring back crypto check

		// Request is valid, insert the body into the channel
		params.RequestChan <- body
		level.Debug(params.Logger).Log("msg", fmt.Sprintf("inserted relay update %s body into channel", request.RemoteAddr))

		// Get relays to ping
		relaysToPing := make([]routing.RelayPingData, 0)
		sellerName := relayHash[id].Seller.Name

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

		// Build and write the response
		var responseData []byte
		response := RelayUpdateResponse{}

		for i := range relaysToPing {
			response.RelaysToPing = append(response.RelaysToPing, routing.RelayPingData{
				ID:      relaysToPing[i].ID,
				Address: relaysToPing[i].Address,
			})
		}
		response.Timestamp = time.Now().Unix()
		response.TargetVersion = relay.Version

		responseData, err = response.MarshalBinary()
		if err != nil {
			params.Metrics.ErrorMetrics.MarshalBinaryResponseFailure.Add(1)
			level.Error(params.Logger).Log("err", fmt.Sprintf("%s - error: failed to write relay update response: %v", request.RemoteAddr, err))
			core.Debug("%s - error: failed to write relay update response: %v", request.RemoteAddr, err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
		writer.Write(responseData)
	}
}
