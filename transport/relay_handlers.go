package transport

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
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

type RelayInitHandlerConfig struct {
	RedisClient      redis.Cmdable
	GeoClient        *routing.GeoClient
	IpLocator        routing.IPLocator
	Storer           storage.Storer
	Duration         metrics.Gauge
	Counter          metrics.Counter
	RouterPrivateKey []byte
}

type RelayUpdateHandlerConfig struct {
	RedisClient           redis.Cmdable
	StatsDb               *routing.StatsDatabase
	Duration              metrics.Gauge
	Counter               metrics.Counter
	TrafficStatsPublisher stats.Publisher
	Storer                storage.Storer
}

// RelayInitHandlerFunc returns the function for the relay init endpoint
func RelayInitHandlerFunc(logger log.Logger, params *RelayInitHandlerConfig) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "init")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			params.Duration.Set(float64(durationSince.Milliseconds()))
			params.Counter.Add(1)
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
			return
		}

		locallogger := log.With(logger, "relay_addr", relayInitRequest.Address.String())

		if relayInitRequest.Magic != InitRequestMagic {
			level.Error(locallogger).Log("msg", "magic number mismatch", "magic_number", relayInitRequest.Magic)
			http.Error(writer, "magic number mismatch", http.StatusBadRequest)
			return
		}

		if relayInitRequest.Version != VersionNumberInitRequest {
			level.Error(locallogger).Log("msg", "version mismatch", "version", relayInitRequest.Version)
			http.Error(writer, "version mismatch", http.StatusBadRequest)
			return
		}

		id := crypto.HashID(relayInitRequest.Address.String())

		relayEntry, ok := params.Storer.Relay(id)
		if !ok {
			level.Error(locallogger).Log("msg", "relay not in firestore")
			http.Error(writer, "relay not in firestore", http.StatusInternalServerError)
			return
		}

		relay := routing.Relay{
			ID:             id,
			Addr:           relayInitRequest.Address,
			PublicKey:      relayEntry.PublicKey,
			Datacenter:     relayEntry.Datacenter,
			Seller:         relayEntry.Seller,
			LastUpdateTime: uint64(time.Now().Unix()),
			Latitude:       relayEntry.Latitude,
			Longitude:      relayEntry.Longitude,
		}

		if _, ok := crypto.Open(relayInitRequest.EncryptedToken, relayInitRequest.Nonce, relay.PublicKey, params.RouterPrivateKey); !ok {
			level.Error(locallogger).Log("msg", "crypto open failed")
			http.Error(writer, "crypto open failed", http.StatusUnauthorized)
			return
		}

		exists := params.RedisClient.HExists(routing.HashKeyAllRelays, relay.Key())

		if exists.Err() != nil && exists.Err() != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to check if relay is registered", "err", exists.Err())
			http.Error(writer, "failed to check if relay is registered", http.StatusNotFound)
			return
		}

		if exists.Val() {
			level.Warn(locallogger).Log("msg", "relay already initialized")
			http.Error(writer, "relay already initialized", http.StatusConflict)
			return
		}

		if loc, err := params.IpLocator.LocateIP(relay.Addr.IP); err == nil {
			relay.Latitude = loc.Latitude
			relay.Longitude = loc.Longitude
		} else {
			level.Warn(locallogger).Log("msg", "using default geolocation from storage for relay")
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
			Timestamp: relay.LastUpdateTime,
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
			params.Duration.Set(float64(durationSince.Milliseconds()))
			params.Counter.Add(1)
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
			return
		}

		if relayUpdateRequest.Version != VersionNumberUpdateRequest {
			level.Error(handlerLogger).Log("msg", "version mismatch", "version", relayUpdateRequest.Version)
			http.Error(writer, "version mismatch", http.StatusBadRequest)
			return
		}

		if len(relayUpdateRequest.PingStats) > MaxRelays {
			level.Error(handlerLogger).Log("msg", "max relays exceeded", "relay count", len(relayUpdateRequest.PingStats))
			http.Error(writer, "max relays exceeded", http.StatusBadRequest)
			return
		}

		relay := routing.Relay{
			ID: crypto.HashID(relayUpdateRequest.Address.String()),
		}

		exists := params.RedisClient.HExists(routing.HashKeyAllRelays, relay.Key())

		if exists.Err() != nil && exists.Err() != redis.Nil {
			level.Error(handlerLogger).Log("msg", "failed to check if relay is registered", "err", exists.Err())
			http.Error(writer, "failed to check if relay is registered", http.StatusInternalServerError)
			return
		}

		if !exists.Val() {
			level.Warn(handlerLogger).Log("msg", "relay not initialized")
			http.Error(writer, "relay not initialized", http.StatusNotFound)
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
			http.Error(writer, "failed to unmarshal relay data", http.StatusBadRequest)
			return
		}

		if !bytes.Equal(relayUpdateRequest.Token, relay.PublicKey) {
			level.Error(handlerLogger).Log("msg", "relay public key doesn't match")
			http.Error(writer, "relay public key doesn't match", http.StatusBadRequest)
			return
		}

		statsUpdate := &routing.RelayStatsUpdate{}
		statsUpdate.ID = relay.ID
		statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdateRequest.PingStats...)

		params.StatsDb.ProcessStats(statsUpdate)

		relay.LastUpdateTime = uint64(time.Now().Unix())

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
