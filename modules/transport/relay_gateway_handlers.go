package transport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/pubsub"
)

type GatewayHandlerConfig struct {
	RelayStore            storage.RelayStore
	RelayCache            storage.RelayCache
	Storer                storage.Storer
	InitMetrics           *metrics.RelayInitMetrics
	UpdateMetrics         *metrics.RelayUpdateMetrics
	RouterPrivateKey      []byte
	Publishers            []pubsub.Publisher
	RelayBackendAddresses []string
	RB15Enabled           bool
	RB15NoInit            bool
	RB2Enabled            bool
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

		id := crypto.HashID(relayInitRequest.Address.String())
		relay, err := params.Storer.Relay(id)
		if err != nil {
			level.Error(localLogger).Log("msg", "failed to get relay from storage", "err", err)
			http.Error(writer, "failed to get relay from storage", http.StatusNotFound)
			params.InitMetrics.ErrorMetrics.RelayNotFound.Add(1)
			return
		}

		if _, ok := crypto.Open(relayInitRequest.EncryptedToken, relayInitRequest.Nonce, relay.PublicKey, params.RouterPrivateKey); !ok {
			level.Error(localLogger).Log("msg", "crypto open failed")
			http.Error(writer, "crypto open failed", http.StatusUnauthorized)
			params.InitMetrics.ErrorMetrics.DecryptionFailure.Add(1)
			return
		}

		relayData, err := params.RelayStore.Get(id)
		if err == nil && relayData.ID == id {
			level.Error(localLogger).Log("msg", "relay already exist", "relay address", relay.Addr.String())
			http.Error(writer, "relay already exists", http.StatusConflict)
			params.InitMetrics.ErrorMetrics.RelayAlreadyExists.Add(1)
			return
		}
		if err != nil && err.Error() != "unable to find relay data" {
			level.Error(localLogger).Log("msg", "relay already exist", "relay address", relay.Addr.String())
			http.Error(writer, "issue storing relay", http.StatusInternalServerError)
			return
		}

		err, errCode := initRelayOnGateway(&relay, relayInitRequest.RelayVersion, localLogger, params)
		if err != nil {
			http.Error(writer, err.Error(), errCode)
		}

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

		//send relay init to relay backend feature_relay_backend_1.5
		//this is after all the checks to ensure that the relay backend will pass all the checks. double the work but stops from waiting.
		if params.RB15Enabled && !params.RB15NoInit {
			for _, address := range params.RelayBackendAddresses {
				go func(address string) {
					resp, err := http.Post(fmt.Sprintf("http://%s/relay_init", address), "application/octet-stream", request.Body)
					if err != nil || resp.StatusCode != http.StatusOK {
						_ = level.Error(localLogger).Log("msg", "unable to send update to relay backend", "err", err)
					}
				}(address)
			}
		}

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
		writer.Write(responseData)
	}
}

func initRelayOnGateway(relay *routing.Relay, relayVersion string, logger log.Logger, params *GatewayHandlerConfig) (error, int) {
	// Don't allow quarantined relays back in
	if relay.State == routing.RelayStateQuarantine {
		level.Error(logger).Log("msg", "quaratined relay attempted to reconnect", "relay", relay.Name)
		params.InitMetrics.ErrorMetrics.RelayQuarantined.Add(1)
		return fmt.Errorf("cannot permit quarantined relay"), http.StatusUnauthorized
	}

	// Set the relay's state to enabled
	relay.State = routing.RelayStateEnabled

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := params.Storer.SetRelay(ctx, *relay); err != nil {
		level.Error(logger).Log("msg", "failed to set relay state in storage", "err", err)
		return fmt.Errorf("failed to set relay state in storer"), http.StatusInternalServerError
	}

	relayData := storage.NewRelayStoreData(relay.ID, relayVersion, relay.Addr)
	err := params.RelayStore.Set(*relayData)
	if err != nil {
		level.Error(logger).Log("redis error %s \n", err.Error())
		return fmt.Errorf("failed to set relay state in relay store"), http.StatusInternalServerError
	}

	level.Debug(logger).Log("msg", "relay initialized")
	return nil, 0

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

		if len(relayUpdateRequest.PingStats) > MaxRelays {
			level.Error(localLogger).Log("msg", "max relays exceeded", "relay count", len(relayUpdateRequest.PingStats))
			http.Error(writer, "max relays exceeded", http.StatusBadRequest)
			params.UpdateMetrics.ErrorMetrics.ExceedMaxRelays.Add(1)
			return
		}

		id := crypto.HashID(relayUpdateRequest.Address.String())
		// If the relay does not exist in Firestore it's a ghost, ignore it
		relay, err := params.Storer.Relay(id)
		if err != nil {
			level.Error(localLogger).Log("msg", "relay does not exist in Firestore (ghost)", "err", err)
			http.Error(writer, "relay does not exist in Firestore (ghost)", http.StatusNotFound)
			params.UpdateMetrics.ErrorMetrics.RelayNotFound.Add(1)
			return
		}

		relayData, err := params.RelayStore.Get(id)
		if relayData == nil || err != nil {
			if params.RB15NoInit {
				err, errCode := initRelayOnGateway(&relay, relayUpdateRequest.RelayVersion, localLogger, params)
				if err != nil {
					http.Error(writer, err.Error(), errCode)
				}
			} else {
				level.Warn(localLogger).Log("msg", "relay not initialized")
				http.Error(writer, "relay not initialized", http.StatusNotFound)
				params.UpdateMetrics.ErrorMetrics.RelayNotFound.Add(1)
				return
			}
		}

		if !bytes.Equal(relayUpdateRequest.Token, relay.PublicKey) {
			level.Error(localLogger).Log("msg", "relay public key doesn't match")
			http.Error(writer, "relay public key doesn't match", http.StatusBadRequest)
			params.UpdateMetrics.ErrorMetrics.InvalidToken.Add(1)
			return
		}

		// Check if the relay state isn't set to enabled, and as a failsafe quarantine the relay
		if relay.State != routing.RelayStateEnabled {
			if params.RB15NoInit {
				err, errCode := initRelayOnGateway(&relay, relayUpdateRequest.RelayVersion, localLogger, params)
				if err != nil {
					http.Error(writer, err.Error(), errCode)
					return
				}
			} else {
				level.Error(localLogger).Log("msg", "non-enabled relay attempting to update", "relay_name", relay.Name, "relay_address", relay.Addr.String(), "relay_state", relay.State)
				http.Error(writer, "cannot allow non-enabled relay to update", http.StatusUnauthorized)
				params.UpdateMetrics.ErrorMetrics.RelayNotEnabled.Add(1)
				return
			}
		}

		// If the relay is shutting down, set the state to maintenance if it was previously operating correctly
		if relayUpdateRequest.ShuttingDown {
			relay, err := params.Storer.Relay(relayData.ID)
			if err != nil {
				level.Error(localLogger).Log("msg", "failed to get relay from storage while shutting down", "err", err)
				http.Error(writer, "failed to get relay from storage while shutting down", http.StatusInternalServerError)
				//todo error metric??
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
				//todo error metric??
				return
			}

			// no need to http error as the relay store item has a expire
			err = params.RelayStore.Delete(id)
			if err != nil {
				level.Error(localLogger).Log("msg", "failed to remove relay from redis store", "err", err)
				//todo error metric??
			}
			return
		}

		//send to relay backend feature_relay_backend_1.5
		//this is after all the checks to ensure that the relay backend pass all the checks. double the work but stops from waiting.
		if params.RB15Enabled {
			for _, address := range params.RelayBackendAddresses {
				go func(address string) {
					resp, err := http.Post(fmt.Sprintf("http://%s/relay_update", address), "application/octet-stream", request.Body)
					if err != nil || resp.StatusCode != http.StatusOK {
						_ = level.Error(localLogger).Log("msg", "unable to send update to relay backend", "err", err)
					}
				}(address)
			}
		}

		//send to optimizer
		if params.RB2Enabled {
			for _, pub := range params.Publishers {
				go func() {
					_, err = pub.Publish(context.Background(), pubsub.RelayUpdateTopic, body)
					if err != nil {
						_ = level.Error(localLogger).Log("msg", "unable to send update to optimizer", "err", err)
					}
				}()
			}
		}

		relaysToPing := make([]routing.RelayPingData, 0)
		allRelayData, err := params.RelayCache.GetAll()

		enableInternalIPs, err := envvar.GetBool("FEATURE_ENABLE_INTERNAL_IPS", false)
		if err != nil {
			level.Error(logger).Log("msg", "unable to parse value of 'ENABLE_INTERNAL_IPS'", "err", err)
		}

		for _, v := range allRelayData {
			if v.ID != relay.ID {
				otherRelay, err := params.Storer.Relay(v.ID)
				if err != nil {
					level.Error(logger).Log("msg", "failed to get other relay from storage", "err", err)
					continue
				}

				if otherRelay.State == routing.RelayStateEnabled {
					var address string
					if enableInternalIPs && relay.Seller.Name == otherRelay.Seller.Name && relay.InternalAddr.String() != ":0" && otherRelay.InternalAddr.String() != ":0" {
						address = otherRelay.InternalAddr.String()
					} else {
						address = otherRelay.Addr.String()
					}
					relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(v.ID), Address: address})
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
