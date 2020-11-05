package transport

import (
	"bytes"
	"context"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport/pubsub"
	"io/ioutil"
	"net/http"
	"time"
)

type GatewayHandlerConfig struct {
	RelayStore 		 storage.RelayStore
	RelayCache	     storage.RelayCache
	Storer           storage.Storer
	InitMetrics      *metrics.RelayInitMetrics
	UpdateMetrics	 *metrics.RelayUpdateMetrics
	RouterPrivateKey []byte
	Publisher        pubsub.Publisher
}

// RelayInitHandlerFunc returns the function for the relay init endpoint
func GatewayRelayInitHandlerFunc(logger log.Logger, params *GatewayHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "init")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.InitMetrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
			params.InitMetrics.Invocations.Add(1)
		}()

		localLogger := log.With(handlerLogger, "req_addr", request.RemoteAddr)

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(localLogger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		defer request.Body.Close()

		var relayInitRequest RelayInitRequest
		switch request.Header.Get("Content-Type") {
		case "application/octet-stream":
			err = relayInitRequest.UnmarshalBinary(body)
		default:
			err = errors.New("unsupported content type")
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			params.InitMetrics.ErrorMetrics.UnmarshalFailure.Add(1)
			return
		}

		localLogger = log.With(localLogger, "relay_addr", relayInitRequest.Address.String())

		if relayInitRequest.Magic != InitRequestMagic {
			level.Error(localLogger).Log("msg", "magic number mismatch", "magic_number", relayInitRequest.Magic)
			http.Error(writer, "magic number mismatch", http.StatusBadRequest)
			params.InitMetrics.ErrorMetrics.InvalidMagic.Add(1)
			return
		}

		if relayInitRequest.Version > VersionNumberInitRequest {
			level.Error(localLogger).Log("msg", "version mismatch", "version", relayInitRequest.Version)
			http.Error(writer, "version mismatch", http.StatusBadRequest)
			params.InitMetrics.ErrorMetrics.InvalidVersion.Add(1)
			return
		}

		id := crypto.HashID(relayInitRequest.Address.String())

		relay, err := params.Storer.Relay(id)
		if err != nil {
			level.Error(localLogger).Log("msg", "failed to get relay from storage", "err", err)
			http.Error(writer, "failed to get relay from storage", http.StatusNotFound)
			params.InitMetrics.ErrorMetrics.RelayNotFound.Add(1)
			return
		}

		// Don't allow quarantined relays back in
		if relay.State == routing.RelayStateQuarantine {
			level.Error(localLogger).Log("msg", "quaratined relay attempted to reconnect", "relay", relay.Name)
			params.InitMetrics.ErrorMetrics.RelayQuarantined.Add(1)
			http.Error(writer, "cannot permit quarantined relay", http.StatusUnauthorized)
			return
		}

		if _, ok := crypto.Open(relayInitRequest.EncryptedToken, relayInitRequest.Nonce, relay.PublicKey, params.RouterPrivateKey); !ok {
			level.Error(localLogger).Log("msg", "crypto open failed")
			http.Error(writer, "crypto open failed", http.StatusUnauthorized)
			params.InitMetrics.ErrorMetrics.DecryptionFailure.Add(1)
			return
		}

		// Set the relay's state to enabled
		relay.State = routing.RelayStateEnabled

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := params.Storer.SetRelay(ctx, relay); err != nil {
			level.Error(localLogger).Log("msg", "failed to set relay state in storage", "err", err)
			http.Error(writer, "failed to set relay state in storage", http.StatusInternalServerError)
			return
		}

		relayData := storage.NewRelayStoreData(id, relayInitRequest.RelayVersion, relayInitRequest.Address)
		params.RelayStore.Set(*relayData)


		level.Debug(localLogger).Log("msg", "relay initialized")

		var responseData []byte
		response := RelayInitResponse{
			Version:   VersionNumberInitResponse,
			Timestamp: uint64(time.Now().Unix()),
			PublicKey: relay.PublicKey,
		}

		switch request.Header.Get("Content-Type") {
		case "application/octet-stream":
			responseData, err = response.MarshalBinary()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
		writer.Write(responseData)
	}
}

// GatewayRelayUpdateHandlerFunc returns the function for the relay update endpoint
func GatewayRelayUpdateHandlerFunc(logger log.Logger, relayslogger log.Logger, params *GatewayHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "update")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.UpdateMetrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
			params.UpdateMetrics.Invocations.Add(1)
		}()

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(handlerLogger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer request.Body.Close()

		localLogger := log.With(handlerLogger, "req_addr", request.RemoteAddr)

		var relayUpdateRequest RelayUpdateRequest
		switch request.Header.Get("Content-Type") {
		case "application/octet-stream":
			err = relayUpdateRequest.UnmarshalBinary(body)
		default:
			err = errors.New("unsupported content type")
		}
		if err != nil {
			level.Error(localLogger).Log("msg", "error unmarshaling relay update request", "err", err)
			http.Error(writer, err.Error(), http.StatusBadRequest)
			params.UpdateMetrics.ErrorMetrics.UnmarshalFailure.Add(1)
			return
		}

		if relayUpdateRequest.Version > VersionNumberUpdateRequest {
			level.Error(localLogger).Log("msg", "version mismatch", "version", relayUpdateRequest.Version)
			http.Error(writer, "version mismatch", http.StatusBadRequest)
			params.UpdateMetrics.ErrorMetrics.InvalidVersion.Add(1)
			return
		}

		if len(relayUpdateRequest.PingStats) > MaxRelays {
			level.Error(localLogger).Log("msg", "max relays exceeded", "relay count", len(relayUpdateRequest.PingStats))
			http.Error(writer, "max relays exceeded", http.StatusBadRequest)
			params.UpdateMetrics.ErrorMetrics.ExceedMaxRelays.Add(1)
			return
		}

		id := crypto.HashID(relayUpdateRequest.Address.String())
		relayData, err:= params.RelayStore.Get(id)
		if relayData == nil  || err != nil{
			level.Warn(localLogger).Log("msg", "relay not initialized")
			http.Error(writer, "relay not initialized", http.StatusNotFound)
			params.UpdateMetrics.ErrorMetrics.RelayNotFound.Add(1)
			return
		}

		// If the relay does not exist in Firestore it's a ghost, ignore it
		relay, err := params.Storer.Relay(relayData.ID)
		if err != nil {
			level.Error(localLogger).Log("msg", "relay does not exist in Firestore (ghost)", "err", err)
			http.Error(writer, "relay does not exist in Firestore (ghost)", http.StatusNotFound)
			params.UpdateMetrics.ErrorMetrics.RelayNotFound.Add(1)
			return
		}

		// If the relay is shutting down, set the state to maintenance if it was previously operating correctly
		if relayUpdateRequest.ShuttingDown {
			relay, err := params.Storer.Relay(relayData.ID)
			if err != nil {
				level.Error(localLogger).Log("msg", "failed to get relay from storage while shutting down", "err", err)
				http.Error(writer, "failed to get relay from storage while shutting down", http.StatusInternalServerError)
				return
			}

			if relay.State == routing.RelayStateEnabled {
				relay.State = routing.RelayStateMaintenance
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := params.Storer.SetRelay(ctx, relay); err != nil {
				level.Error(localLogger).Log("msg", "failed to set relay state in storage while shutting down", "err", err)
				http.Error(writer, "failed to set relay state in storage while shutting down", http.StatusInternalServerError)
				return
			}

			params.RelayStore.Delete(id)
			return
		}

		if !bytes.Equal(relayUpdateRequest.Token, relay.PublicKey) {
			level.Error(localLogger).Log("msg", "relay public key doesn't match")
			http.Error(writer, "relay public key doesn't match", http.StatusBadRequest)
			params.UpdateMetrics.ErrorMetrics.InvalidToken.Add(1)
			return
		}

		// Check if the relay state isn't set to enabled, and as a failsafe quarantine the relay
		if relay.State != routing.RelayStateEnabled {
			level.Error(localLogger).Log("msg", "non-enabled relay attempting to update", "relay_name", relay.Name, "relay_address", relay.Addr.String(), "relay_state", relay.State)
			http.Error(writer, "cannot allow non-enabled relay to update", http.StatusUnauthorized)
			params.UpdateMetrics.ErrorMetrics.RelayNotEnabled.Add(1)
			return
		}

		var topic pubsub.Topic = 1
		params.Publisher.Publish(context.Background(),topic, body)

		relaysToPing := make([]routing.RelayPingData, 0)
		allRelayData, err := params.RelayCache.GetAll()
		for _, v := range allRelayData {
			if v.Address.String() != relayData.Address.String() {
				relay, err := params.Storer.Relay(v.ID)
				if err != nil {
					level.Error(localLogger).Log("msg", "failed to get other relay from storage", "err", err)
					continue
				}

				if relay.State == routing.RelayStateEnabled {
					relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(v.ID), Address: v.Addr.String()})
				}
			}
		}

		// Update the relay data
		err = params.RelayStore.ExpireReset(id)
		if err != nil {
			level.Error(localLogger).Log("msg", "failed to update relay", "err", err)
			http.Error(writer, "failed to update relay", http.StatusInternalServerError)
		}


		level.Debug(relayslogger).Log(
			"id", relayData.ID,
			"name", relay.Name,
			"addr", relayData.Address.String(),
			"datacenter", relay.Datacenter.Name,
			"session_count", relayUpdateRequest.TrafficStats.SessionCount,
			"bytes_received", relayUpdateRequest.TrafficStats.AllRx(),
			"bytes_send", relayUpdateRequest.TrafficStats.AllTx(),
		)

		level.Debug(localLogger).Log("msg", "relay updated")

		var responseData []byte
		response := RelayUpdateResponse{}
		for _, pingData := range relaysToPing {
			response.RelaysToPing = append(response.RelaysToPing, routing.RelayPingData{
				ID:      pingData.ID,
				Address: pingData.Address,
			})
		}
		response.Timestamp = time.Now().Unix()

		switch request.Header.Get("Content-Type") {
		case "application/octet-stream":
			responseData, err = response.MarshalBinary()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
		writer.Write(responseData)
	}
}