package transport

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/pubsub"
)

type GatewayConfig struct {
	PublisherSendBuffer   int
	PublishToHosts        []string
	// RouterPrivateKey      []byte
	NRBNoInit             bool
	NRBHTTP               bool
	RelayBackendAddresses []string
	Loadtest              bool
	PublisherRefreshTimer time.Duration
	HTTPTimeout           time.Duration
}

type GatewayHandlerConfig struct {
	// Storer                storage.Storer
	InitMetrics           *metrics.RelayInitMetrics
	UpdateMetrics         *metrics.RelayUpdateMetrics
	// RouterPrivateKey      []byte
	Publishers            []pubsub.Publisher
	RelayBackendAddresses []string
	NRBNoInit             bool
	NRBHTTP               bool
	LoadTest              bool
}

// RelayInitHandlerFunc returns the function for the relay init endpoint
// func GatewayRelayInitHandlerFunc(logger log.Logger, params *GatewayHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {
// 	handlerLogger := log.With(logger, "handler", "init")

// 	return func(writer http.ResponseWriter, request *http.Request) {
// 		durationStart := time.Now()
// 		defer func() {
// 			durationSince := time.Since(durationStart)
// 			params.InitMetrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
// 			params.InitMetrics.Invocations.Add(1)
// 		}()

// 		localLogger := log.With(handlerLogger, "req_addr", request.RemoteAddr)

// 		body, err := ioutil.ReadAll(request.Body)
// 		if err != nil {
// 			level.Error(localLogger).Log("msg", "could not read packet", "err", err)
// 			writer.WriteHeader(http.StatusBadRequest)
// 			return
// 		}
// 		defer request.Body.Close()

// 		var relayInitRequest RelayInitRequest
// 		switch request.Header.Get("Content-Type") {
// 		case "application/octet-stream":
// 			err = relayInitRequest.UnmarshalBinary(body)
// 		default:
// 			err = errors.New("unsupported content type")
// 		}
// 		if err != nil {
// 			http.Error(writer, err.Error(), http.StatusBadRequest)
// 			params.InitMetrics.ErrorMetrics.UnmarshalFailure.Add(1)
// 			return
// 		}

// 		localLogger = log.With(localLogger, "relay_addr", relayInitRequest.Address.String())

// 		if relayInitRequest.Magic != InitRequestMagic {
// 			level.Error(localLogger).Log("msg", "magic number mismatch", "magic_number", relayInitRequest.Magic)
// 			http.Error(writer, "magic number mismatch", http.StatusBadRequest)
// 			params.InitMetrics.ErrorMetrics.InvalidMagic.Add(1)
// 			return
// 		}

// 		id := crypto.HashID(relayInitRequest.Address.String())
// 		var relay routing.Relay
// 		if !params.LoadTest {
// 			relay, err = params.Storer.Relay(id)
// 			if err != nil {
// 				level.Error(localLogger).Log("msg", "failed to get relay from storage", "err", err)
// 				http.Error(writer, "failed to get relay from storage", http.StatusNotFound)
// 				params.InitMetrics.ErrorMetrics.RelayNotFound.Add(1)
// 				return
// 			}
// 		} else {
// 			relay = loadTestRelay(relayInitRequest.Address.String(), routing.RelayStateDisabled)
// 		}

// 		if _, ok := crypto.Open(relayInitRequest.EncryptedToken, relayInitRequest.Nonce, relay.PublicKey, params.RouterPrivateKey); !ok {
// 			level.Error(localLogger).Log("msg", "crypto open failed")
// 			http.Error(writer, "crypto open failed", http.StatusUnauthorized)
// 			params.InitMetrics.ErrorMetrics.DecryptionFailure.Add(1)
// 			return
// 		}

// 		if relay.State == routing.RelayStateEnabled {
// 			level.Error(localLogger).Log("msg", "relay already exist", "relay address", relay.Addr.String())
// 			params.InitMetrics.ErrorMetrics.RelayAlreadyExists.Add(1)
// 			if !params.NRBNoInit {
// 				http.Error(writer, "relay already active", http.StatusConflict)
// 				return
// 			}
// 		}

// 		if !params.LoadTest {
// 			err, errCode := initRelayOnGateway(&relay, relayInitRequest.RelayVersion, localLogger, params)
// 			if err != nil {
// 				http.Error(writer, err.Error(), errCode)
// 			}
// 		}

// 		var responseData []byte
// 		response := RelayInitResponse{
// 			Version:   VersionNumberInitResponse,
// 			Timestamp: uint64(time.Now().Unix()),
// 			PublicKey: relay.PublicKey,
// 		}

// 		switch request.Header.Get("Content-Type") {
// 		case "application/octet-stream":
// 			responseData, err = response.MarshalBinary()
// 			if err != nil {
// 				writer.WriteHeader(http.StatusInternalServerError)
// 				return
// 			}
// 		}

// 		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))
// 		writer.Write(responseData)
// 	}
// }

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
	// todo change to update relay when sql is working
	if err := params.Storer.SetRelay(ctx, *relay); err != nil {
		level.Error(logger).Log("msg", "failed to set relay state in storage", "err", err)
		return fmt.Errorf("failed to set relay state in storer"), http.StatusInternalServerError
	}

	level.Debug(logger).Log("msg", "relay initialized")
	return nil, 0
}

// GatewayRelayUpdateHandlerFunc returns the function for the relay update endpoint

func GatewayRelayUpdateHandlerFunc(logger log.Logger, relayslogger log.Logger, params *GatewayHandlerConfig, requestChan chan []byte) func(writer http.ResponseWriter, request *http.Request) {
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
		level.Debug(localLogger).Log("msg", "relay update received")

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

		var relay routing.Relay
		if !params.LoadTest {
			// If the relay does not exist in Firestore it's a ghost, ignore it
			relay, err = params.Storer.Relay(id)
			if err != nil {
				level.Error(localLogger).Log("msg", "relay does not exist in Firestore (ghost)", "err", err)
				http.Error(writer, "relay does not exist in Firestore (ghost)", http.StatusNotFound)
				params.UpdateMetrics.ErrorMetrics.RelayNotFound.Add(1)
				return
			}
		} else {
			relay = loadTestRelay(relayUpdateRequest.Address.String(), routing.RelayStateEnabled)
		}

		if !bytes.Equal(relayUpdateRequest.Token, relay.PublicKey) {
			level.Error(localLogger).Log("msg", "relay public key doesn't match")
			http.Error(writer, "relay public key doesn't match", http.StatusBadRequest)
			params.UpdateMetrics.ErrorMetrics.InvalidToken.Add(1)
			return
		}

		if relay.State != routing.RelayStateEnabled {
			if params.NRBNoInit {
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
		if relayUpdateRequest.ShuttingDown && !params.LoadTest {

			if relay.State == routing.RelayStateEnabled {
				relay.State = routing.RelayStateMaintenance
			}

			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			// todo update instead of set when sql ready
			if err := params.Storer.SetRelay(ctx, relay); err != nil {
				level.Error(localLogger).Log("msg", "failed to set relay state in storage while shutting down", "err", err)
				http.Error(writer, "failed to set relay state in storage while shutting down", http.StatusInternalServerError)
				// todo error metric??
				return
			}

		}

		requestChan <- body

		relaysToPing := make([]routing.RelayPingData, 0)
		if !params.LoadTest {
			allRelayData := params.Storer.Relays()
			enableInternalIPs, err := envvar.GetBool("FEATURE_ENABLE_INTERNAL_IPS", false)
			if err != nil {
				level.Error(logger).Log("msg", "unable to parse value of 'ENABLE_INTERNAL_IPS'", "err", err)
			}

			for _, v := range allRelayData {
				if v.ID != relay.ID {
					if v.State == routing.RelayStateEnabled {
						address := v.Addr.String()
						if enableInternalIPs && relay.Seller.Name == v.Seller.Name && relay.InternalAddr.String() != ":0" && v.InternalAddr.String() != ":0" {
							address = v.InternalAddr.String()
						}
						relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(v.ID), Address: address})
					}
				}
			}
		}
		level.Debug(relayslogger).Log(
			"id", relay.ID,
			"name", relay.Name,
			"addr", relay.Addr.String(),
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
