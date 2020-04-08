package transport

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/go-redis/redis/v7"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/stats"
	"github.com/networknext/backend/storage"
)

const (
	InitRequestMagic = 0x9083708f

	MaxRelays             = 1024
	MaxRelayAddressLength = 256
)

type RelayHandlerConfig struct {
	RedisClient           redis.Cmdable
	GeoClient             *routing.GeoClient
	IpLocator             routing.IPLocator
	Storer                storage.Storer
	StatsDb               *routing.StatsDatabase
	TrafficStatsPublisher stats.Publisher
	Metrics               *metrics.RelayHandlerMetrics
	RouterPrivateKey      []byte
}

type RelayInitHandlerConfig struct {
	RedisClient      redis.Cmdable
	GeoClient        *routing.GeoClient
	IpLocator        routing.IPLocator
	Storer           storage.Storer
	Metrics          *metrics.RelayInitMetrics
	RouterPrivateKey []byte
}

type RelayUpdateHandlerConfig struct {
	RedisClient           redis.Cmdable
	StatsDb               *routing.StatsDatabase
	Metrics               *metrics.RelayUpdateMetrics
	TrafficStatsPublisher stats.Publisher
	Storer                storage.Storer
}

// RelayHandlerFunc returns the function for the relays endpoint
func RelayHandlerFunc(logger log.Logger, params *RelayHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "relay")

	return func(writer http.ResponseWriter, request *http.Request) {
		// Set up metrics
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.Metrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
			params.Metrics.Invocations.Add(1)
		}()

		// Read in the request
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(handlerLogger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		handlerLogger = log.With(handlerLogger, "req_addr", request.RemoteAddr)

		// Unmarshal the request packet
		var relayRequest RelayRequest
		if err := relayRequest.UnmarshalJSON(body); err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			params.Metrics.ErrorMetrics.UnmarshalFailure.Add(1)
			return
		}

		// Check that the request doesn't exceed the maximum number of relays that a relay can ping
		if len(relayRequest.PingStats) > MaxRelays {
			level.Error(handlerLogger).Log("msg", "max relays exceeded", "relay count", len(relayRequest.PingStats))
			http.Error(writer, "max relays exceeded", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.ExceedMaxRelays.Add(1)
			return
		}

		locallogger := log.With(logger, "relay_addr", relayRequest.Address.String())

		// Gets the relay ID as a hash of its address
		id := crypto.HashID(relayRequest.Address.String())

		// Check if the relay is registered in firestore
		relayEntry, err := params.Storer.Relay(id)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to get relay from storage", "err", err)
			http.Error(writer, "failed to get relay from storage", http.StatusNotFound)
			params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
			return
		}

		// Set the relay to the firestore entry for now
		relay := relayEntry

		// Get the relay's HTTP authorization header
		authorizationHeader := request.Header.Get("Authorization")
		if authorizationHeader == "" {
			level.Error(locallogger).Log("msg", "no authorization header")
			http.Error(writer, "no authorization header", http.StatusUnauthorized)
			params.Metrics.ErrorMetrics.NoAuthHeader.Add(1)
			return
		}

		// Get the token from the authorization header
		tokenIndex := len("Bearer ")
		if tokenIndex >= len(authorizationHeader) {
			level.Error(locallogger).Log("msg", "bad authorization header length")
			http.Error(writer, "bad authorization header length", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.BadAuthHeaderLength.Add(1)
			return
		}
		token := authorizationHeader[tokenIndex:]

		// Split the token into the base64 encoded nonce and address
		splitResult := strings.Split(token, ":")
		if splitResult == nil || len(splitResult) != 2 {
			level.Error(locallogger).Log("msg", "bad authorization token")
			http.Error(writer, "bad authorization token", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.BadAuthHeaderToken.Add(1)
			return
		}

		nonceString := splitResult[0]
		encryptedAddressString := splitResult[1]

		// Decode the base64
		nonce, err := base64.StdEncoding.DecodeString(nonceString)
		if err != nil {
			level.Error(locallogger).Log("msg", "bad nonce")
			http.Error(writer, "bad nonce", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.BadNonce.Add(1)
			return
		}

		encryptedAddress, err := base64.StdEncoding.DecodeString(encryptedAddressString)
		if err != nil {
			level.Error(locallogger).Log("msg", "bad encrypted address")
			http.Error(writer, "bad encrypted address", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.BadEncryptedAddress.Add(1)
			return
		}

		// Decrypt the address
		if _, ok := crypto.Open(encryptedAddress, nonce, relay.PublicKey, params.RouterPrivateKey); !ok {
			level.Error(locallogger).Log("msg", "crypto open failed")
			http.Error(writer, "crypto open failed", http.StatusUnauthorized)
			params.Metrics.ErrorMetrics.DecryptFailure.Add(1)
			return
		}

		// Check if the relay exists in redis
		exists := params.RedisClient.HExists(routing.HashKeyAllRelays, relay.Key())

		if exists.Err() != nil && exists.Err() != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to check if relay is registered", "err", exists.Err())
			http.Error(writer, "failed to check if relay is registered", http.StatusInternalServerError)
			params.Metrics.ErrorMetrics.RedisFailure.Add(1)
			return
		}

		// If the relay doesn't exist, add it
		if !exists.Val() {
			// Set the relay's lat long
			if loc, err := params.IpLocator.LocateIP(relay.Addr.IP); err == nil {
				relay.Latitude = loc.Latitude
				relay.Longitude = loc.Longitude
			} else {
				level.Warn(locallogger).Log("msg", "using default geolocation from storage for relay")
			}

			// Regular set for expiry
			if res := params.RedisClient.Set(relay.Key(), relay.ID, routing.RelayTimeout); res.Err() != nil && res.Err() != redis.Nil {
				level.Error(locallogger).Log("msg", "failed to set relay expiry in redis", "err", res.Err())
				http.Error(writer, "failed to initialize relay", http.StatusInternalServerError)
				return
			}

			// HSet for full relay data
			if res := params.RedisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay); res.Err() != nil && res.Err() != redis.Nil {
				level.Error(locallogger).Log("msg", "failed to store relay in redis", "err", res.Err())
				http.Error(writer, "failed to initialize relay", http.StatusInternalServerError)
				return
			}

			if err := params.GeoClient.Add(relay); err != nil {
				level.Error(locallogger).Log("msg", "failed to add relay to geoclient", "err", err)
				http.Error(writer, "failed to initialize relay", http.StatusInternalServerError)
				return
			}

			level.Debug(locallogger).Log("msg", "relay initialized")
		} else { // If the relay exists in redis, then get it and use that instead of the firestore version
			// Get the relay from redis
			hgetResult := params.RedisClient.HGet(routing.HashKeyAllRelays, relay.Key())
			if hgetResult.Err() != nil && hgetResult.Err() != redis.Nil {
				level.Error(handlerLogger).Log("msg", "failed to get relay", "err", exists.Err())
				http.Error(writer, "failed to get relay", http.StatusNotFound)
				return
			}

			data, err := hgetResult.Bytes()
			if err != nil && err != redis.Nil {
				level.Error(handlerLogger).Log("msg", "failed to get relay data", "err", err)
				http.Error(writer, "failed to get relay data", http.StatusInternalServerError)
				return
			}

			// Unmarshal the relay entry
			if err = relay.UnmarshalBinary(data); err != nil {
				level.Error(handlerLogger).Log("msg", "failed to unmarshal relay data", "err", err)
				http.Error(writer, "failed to unmarshal relay data", http.StatusInternalServerError)
				params.Metrics.ErrorMetrics.RelayUnmarshalFailure.Add(1)
				return
			}
		}

		// Update the relay's last update time
		relay.LastUpdateTime = time.Now()

		// Update the relay's ping stats in statsdb
		statsUpdate := &routing.RelayStatsUpdate{}
		statsUpdate.ID = relay.ID
		//statsUpdate.PingStats = append(statsUpdate.PingStats, relayRequest.PingStats...)

		// For compatibility, convert the ping stats to the old struct for now
		relayStatsPing := make([]routing.RelayStatsPing, len(relayRequest.PingStats))
		for i := 0; i < len(relayStatsPing); i++ {
			relayStatsPing[i] = routing.RelayStatsPing{
				RelayID:    relayRequest.PingStats[i].ID,
				RTT:        relayRequest.PingStats[i].RTT,
				Jitter:     relayRequest.PingStats[i].Jitter,
				PacketLoss: relayRequest.PingStats[i].PacketLoss,
			}
		}
		statsUpdate.PingStats = relayStatsPing

		params.StatsDb.ProcessStats(statsUpdate)

		// Store the relay back in redis

		// Regular set for expiry
		if res := params.RedisClient.Set(relay.Key(), relay.ID, routing.RelayTimeout); res.Err() != nil {
			level.Error(handlerLogger).Log("msg", "failed to set relay expiry in redis", "err", res.Err())
			http.Error(writer, "failed to update relay", http.StatusInternalServerError)
			return
		}

		// HSet for full relay data
		if res := params.RedisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay); res.Err() != nil {
			level.Error(handlerLogger).Log("msg", "failed to store relay in redis", "err", res.Err())
			http.Error(writer, "failed to update relay", http.StatusInternalServerError)
			return
		}

		// Get all of the relays to make a list of relays for the requesting relay to ping
		hgetallResult := params.RedisClient.HGetAll(routing.HashKeyAllRelays)
		if hgetallResult.Err() != nil && hgetallResult.Err() != redis.Nil {
			level.Error(handlerLogger).Log("msg", "failed to get other relays", "err", hgetallResult.Err())
			http.Error(writer, "failed to get other relays", http.StatusNotFound)
			return
		}

		level.Debug(handlerLogger).Log("msg", "relay updated")

		// Create the list of relays to ping
		relaysToPing := make([]RelayPingStats, 0)
		for k, v := range hgetallResult.Val() {
			if k != relay.Key() {
				var unmarshaledValue routing.Relay
				if err := unmarshaledValue.UnmarshalBinary([]byte(v)); err != nil {
					level.Error(handlerLogger).Log("msg", "failed to get other relay", "err", err)
					continue
				}

				relaysToPing = append(relaysToPing, RelayPingStats{
					ID:      unmarshaledValue.ID,
					Address: unmarshaledValue.Addr.String(),
				})
			}
		}

		// Send back the response
		var responseData []byte
		response := RelayRequest{}
		response.Address = relay.Addr
		response.PingStats = relaysToPing

		responseData, err = response.MarshalJSON()
		if err != nil {
			level.Error(handlerLogger).Log("msg", "failed to marshal response JSON", "err", err)
			http.Error(writer, "failed to marshal response JSON", http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(responseData)
	}
}

// RelayInitHandlerFunc returns the function for the relay init endpoint
func RelayInitHandlerFunc(logger log.Logger, params *RelayInitHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "init")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.Metrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
			params.Metrics.Invocations.Add(1)
		}()

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(handlerLogger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		handlerLogger = log.With(handlerLogger, "req_addr", request.RemoteAddr)

		var relayInitRequest RelayInitRequest
		switch request.Header.Get("Content-Type") {
		case "application/json":
			err = relayInitRequest.UnmarshalJSON(body)
		case "application/octet-stream":
			err = relayInitRequest.UnmarshalBinary(body)
		default:
			err = errors.New("unsupported content type")
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			params.Metrics.ErrorMetrics.UnmarshalFailure.Add(1)
			return
		}

		locallogger := log.With(logger, "relay_addr", relayInitRequest.Address.String())

		if relayInitRequest.Magic != InitRequestMagic {
			level.Error(locallogger).Log("msg", "magic number mismatch", "magic_number", relayInitRequest.Magic)
			http.Error(writer, "magic number mismatch", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.InvalidMagic.Add(1)
			return
		}

		if relayInitRequest.Version != VersionNumberInitRequest {
			level.Error(locallogger).Log("msg", "version mismatch", "version", relayInitRequest.Version)
			http.Error(writer, "version mismatch", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.InvalidVersion.Add(1)
			return
		}

		id := crypto.HashID(relayInitRequest.Address.String())

		relayEntry, err := params.Storer.Relay(id)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to get relay from storage", "err", err)
			http.Error(writer, "failed to get relay from storage", http.StatusNotFound)
			params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
			return
		}

		relay := routing.Relay{
			ID:             id,
			Addr:           relayInitRequest.Address,
			PublicKey:      relayEntry.PublicKey,
			Datacenter:     relayEntry.Datacenter,
			Seller:         relayEntry.Seller,
			LastUpdateTime: time.Now(),
			Latitude:       relayEntry.Latitude,
			Longitude:      relayEntry.Longitude,
		}

		if _, ok := crypto.Open(relayInitRequest.EncryptedToken, relayInitRequest.Nonce, relay.PublicKey, params.RouterPrivateKey); !ok {
			level.Error(locallogger).Log("msg", "crypto open failed")
			http.Error(writer, "crypto open failed", http.StatusUnauthorized)
			params.Metrics.ErrorMetrics.DecryptionFailure.Add(1)
			return
		}

		exists := params.RedisClient.HExists(routing.HashKeyAllRelays, relay.Key())

		if exists.Err() != nil && exists.Err() != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to check if relay is registered", "err", exists.Err())
			http.Error(writer, "failed to check if relay is registered", http.StatusInternalServerError)
			params.Metrics.ErrorMetrics.RedisFailure.Add(1)
			return
		}

		if exists.Val() {
			level.Warn(locallogger).Log("msg", "relay already initialized")
			http.Error(writer, "relay already initialized", http.StatusConflict)
			params.Metrics.ErrorMetrics.RelayAlreadyExists.Add(1)
			return
		}

		if loc, err := params.IpLocator.LocateIP(relay.Addr.IP); err == nil {
			relay.Latitude = loc.Latitude
			relay.Longitude = loc.Longitude
		} else {
			level.Warn(locallogger).Log("msg", "using default geolocation from storage for relay")
			params.Metrics.ErrorMetrics.IPLookupFailure.Add(1)
		}

		// Regular set for expiry
		if res := params.RedisClient.Set(relay.Key(), relay.ID, routing.RelayTimeout); res.Err() != nil && res.Err() != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to initialize relay", "err", res.Err())
			http.Error(writer, "failed to initialize relay", http.StatusInternalServerError)
			return
		}

		relay.State = routing.RelayStateOnline

		// HSet for full relay data
		if res := params.RedisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay); res.Err() != nil && res.Err() != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to initialize relay", "err", res.Err())
			http.Error(writer, "failed to initialize relay", http.StatusInternalServerError)
			return
		}

		if err := params.GeoClient.Add(relay); err != nil {
			level.Error(locallogger).Log("msg", "failed to initialize relay", "err", err)
			http.Error(writer, "failed to initialize relay", http.StatusInternalServerError)
			return
		}

		level.Debug(locallogger).Log("msg", "relay initialized")

		var responseData []byte
		response := RelayInitResponse{
			Version:   VersionNumberInitResponse,
			Timestamp: uint64(relay.LastUpdateTime.Unix()),
			PublicKey: relay.PublicKey,
		}

		switch request.Header.Get("Content-Type") {
		case "application/json":
			response.Timestamp = response.Timestamp * 1000

			responseData, err = response.MarshalJSON()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
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

// RelayUpdateHandlerFunc returns the function for the relay update endpoint
func RelayUpdateHandlerFunc(logger log.Logger, params *RelayUpdateHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "update")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.Metrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
			params.Metrics.Invocations.Add(1)
		}()

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		handlerLogger = log.With(handlerLogger, "req_addr", request.RemoteAddr)

		var relayUpdateRequest RelayUpdateRequest
		switch request.Header.Get("Content-Type") {
		case "application/json":
			err = relayUpdateRequest.UnmarshalJSON(body)
		case "application/octet-stream":
			err = relayUpdateRequest.UnmarshalBinary(body)
		default:
			err = errors.New("unsupported content type")
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			params.Metrics.ErrorMetrics.UnmarshalFailure.Add(1)
			return
		}

		if relayUpdateRequest.Version != VersionNumberUpdateRequest {
			level.Error(handlerLogger).Log("msg", "version mismatch", "version", relayUpdateRequest.Version)
			http.Error(writer, "version mismatch", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.InvalidVersion.Add(1)
			return
		}

		if len(relayUpdateRequest.PingStats) > MaxRelays {
			level.Error(handlerLogger).Log("msg", "max relays exceeded", "relay count", len(relayUpdateRequest.PingStats))
			http.Error(writer, "max relays exceeded", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.ExceedMaxRelays.Add(1)
			return
		}

		relay := routing.Relay{
			ID: crypto.HashID(relayUpdateRequest.Address.String()),
		}

		exists := params.RedisClient.HExists(routing.HashKeyAllRelays, relay.Key())

		if exists.Err() != nil && exists.Err() != redis.Nil {
			level.Error(handlerLogger).Log("msg", "failed to check if relay is registered", "err", exists.Err())
			http.Error(writer, "failed to check if relay is registered", http.StatusInternalServerError)
			params.Metrics.ErrorMetrics.RedisFailure.Add(1)
			return
		}

		if !exists.Val() {
			level.Warn(handlerLogger).Log("msg", "relay not initialized")
			http.Error(writer, "relay not initialized", http.StatusNotFound)
			params.Metrics.ErrorMetrics.RelayNotFound.Add(1)
			return
		}

		hgetResult := params.RedisClient.HGet(routing.HashKeyAllRelays, relay.Key())
		if hgetResult.Err() != nil && hgetResult.Err() != redis.Nil {
			level.Error(handlerLogger).Log("msg", "failed to get relays", "err", exists.Err())
			http.Error(writer, "failed to get relays", http.StatusNotFound)
			return
		}

		data, err := hgetResult.Bytes()
		if err != nil && err != redis.Nil {
			level.Error(handlerLogger).Log("msg", "failed to get relay data", "err", err)
			http.Error(writer, "failed to get relay data", http.StatusInternalServerError)
			return
		}

		if err = relay.UnmarshalBinary(data); err != nil {
			level.Error(handlerLogger).Log("msg", "failed to unmarshal relay data", "err", err)
			http.Error(writer, "failed to unmarshal relay data", http.StatusInternalServerError)
			params.Metrics.ErrorMetrics.RelayUnmarshalFailure.Add(1)
			return
		}

		if !bytes.Equal(relayUpdateRequest.Token, relay.PublicKey) {
			level.Error(handlerLogger).Log("msg", "relay public key doesn't match")
			http.Error(writer, "relay public key doesn't match", http.StatusBadRequest)
			params.Metrics.ErrorMetrics.InvalidToken.Add(1)
			return
		}

		statsUpdate := &routing.RelayStatsUpdate{}
		statsUpdate.ID = relay.ID
		statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdateRequest.PingStats...)

		params.StatsDb.ProcessStats(statsUpdate)

		relay.LastUpdateTime = time.Now()

		relaysToPing := make([]routing.RelayPingData, 0)

		// Regular set for expiry
		if res := params.RedisClient.Set(relay.Key(), 0, routing.RelayTimeout); res.Err() != nil {
			level.Error(handlerLogger).Log("msg", "failed to store relay update expiry", "err", res.Err())
			http.Error(writer, "failed to store relay update expiry", http.StatusInternalServerError)
			return
		}

		if relayUpdateRequest.ShuttingDown {
			relay.State = routing.RelayStateShuttingDown
		}

		// HSet for full relay data
		if res := params.RedisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay); res.Err() != nil {
			level.Error(handlerLogger).Log("msg", "failed to store relay update", "err", res.Err())
			http.Error(writer, "failed to store relay update", http.StatusInternalServerError)
			return
		}

		hgetallResult := params.RedisClient.HGetAll(routing.HashKeyAllRelays)
		if hgetallResult.Err() != nil && hgetallResult.Err() != redis.Nil {
			level.Error(handlerLogger).Log("msg", "failed to get other relays", "err", hgetallResult.Err())
			http.Error(writer, "failed to get other relays", http.StatusNotFound)
			return
		}

		for k, v := range hgetallResult.Val() {
			if k != relay.Key() {
				var unmarshaledValue routing.Relay
				if err := unmarshaledValue.UnmarshalBinary([]byte(v)); err != nil {
					level.Error(handlerLogger).Log("msg", "failed to get other relay", "err", err)
					continue
				}

				if unmarshaledValue.State == routing.RelayStateOnline {
					relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(unmarshaledValue.ID), Address: unmarshaledValue.Addr.String()})
				}
			}
		}

		level.Debug(handlerLogger).Log("msg", "relay updated")

		var responseData []byte
		response := RelayUpdateResponse{}
		for _, pingData := range relaysToPing {
			var token routing.LegacyPingToken
			token.Timeout = uint64(time.Now().Unix() * 100000) // some arbitrarily large number just to make things compatable for the moment
			token.RelayID = crypto.HashID(relayUpdateRequest.Address.String())
			bin, _ := token.MarshalBinary()

			var legacy routing.LegacyPingData
			legacy.ID = pingData.ID
			legacy.Address = pingData.Address
			legacy.PingToken = base64.StdEncoding.EncodeToString(bin)

			response.RelaysToPing = append(response.RelaysToPing, legacy)
		}

		switch request.Header.Get("Content-Type") {
		case "application/json":
			responseData, err = response.MarshalJSON()
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
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

func HealthzHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}
