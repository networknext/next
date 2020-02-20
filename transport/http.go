package transport

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

const (
	InitRequestMagic = 0x9083708f

	MaxRelays             = 1024
	MaxRelayAddressLength = 256

	VersionNumberInitRequest    = 0
	VersionNumberInitResponse   = 0
	VersionNumberUpdateRequest  = 0
	VersionNumberUpdateResponse = 0
)

// NewRouter creates a router with the specified endpoints
func NewRouter(logger log.Logger, redisClient *redis.Client, geoClient *routing.GeoClient, ipLocator routing.IPLocator, storer storage.Storer, statsdb *routing.StatsDatabase, metricsHandler metrics.Handler, costmatrix *routing.CostMatrix, routematrix *routing.RouteMatrix, routerPrivateKey []byte) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/relay_init", RelayInitHandlerFunc(logger, redisClient, geoClient, ipLocator, storer, metricsHandler, routerPrivateKey)).Methods("POST")
	router.HandleFunc("/relay_update", RelayUpdateHandlerFunc(logger, redisClient, statsdb, metricsHandler)).Methods("POST")
	router.Handle("/cost_matrix", costmatrix).Methods("GET")
	router.Handle("/route_matrix", routematrix).Methods("GET")
	router.HandleFunc("/near", NearHandlerFunc(nil)).Methods("GET")
	return router
}

// RelayInitHandlerFunc returns the function for the relay init endpoint
func RelayInitHandlerFunc(logger log.Logger, redisClient *redis.Client, geoClient *routing.GeoClient, ipLocator routing.IPLocator, storer storage.Storer, metricsHandler metrics.Handler, routerPrivateKey []byte) func(writer http.ResponseWriter, request *http.Request) {
	logger = log.With(logger, "handler", "init")

	return func(writer http.ResponseWriter, request *http.Request) {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		relayInitPacket := RelayInitPacket{}

		if err = relayInitPacket.UnmarshalBinary(body); err != nil {
			level.Error(logger).Log("msg", "could not unmarshal packet", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		locallogger := log.With(logger, "req_addr", request.RemoteAddr, "relay_addr", relayInitPacket.Address.String())

		if relayInitPacket.Magic != InitRequestMagic {
			level.Error(locallogger).Log("msg", "magic number mismatch", "magic_number", relayInitPacket.Magic)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		if relayInitPacket.Version != VersionNumberInitRequest {
			level.Error(locallogger).Log("msg", "version mismatch", "version", relayInitPacket.Version)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		id := crypto.HashID(relayInitPacket.Address.String())

		relayEntry, ok := storer.Relay(id)
		if !ok {
			level.Error(locallogger).Log("msg", "relay not in configstore")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		relay := routing.Relay{
			ID:             id,
			Addr:           relayInitPacket.Address,
			PublicKey:      relayEntry.PublicKey,
			Datacenter:     relayEntry.Datacenter,
			LastUpdateTime: uint64(time.Now().Unix()),
		}

		if _, ok := crypto.Open(relayInitPacket.EncryptedToken, relayInitPacket.Nonce, relay.PublicKey, routerPrivateKey); !ok {
			level.Error(locallogger).Log("msg", "crypto open failed")
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		exists := redisClient.HExists(routing.HashKeyAllRelays, relay.Key())

		if exists.Err() != nil && exists.Err() != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to check if relay is registered", "err", exists.Err())
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		if exists.Val() {
			level.Warn(locallogger).Log("msg", "relay already initialized")
			writer.WriteHeader(http.StatusConflict)
			return
		}

		loc, err := ipLocator.LocateIP(relay.Addr.IP)
		if err != nil {
			level.Error(locallogger).Log("msg", "failed to locate relay")
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		relay.Latitude = loc.Latitude
		relay.Longitude = loc.Longitude

		if res := redisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay); res.Err() != nil && res.Err() != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to initialize relay", "err", res.Err())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := geoClient.Add(relay); err != nil {
			level.Error(locallogger).Log("msg", "failed to initialize relay", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		level.Debug(locallogger).Log("msg", "relay initialized")

		writer.Header().Set("Content-Type", "application/octet-stream")

		index := 0
		responseData := make([]byte, 64)
		encoding.WriteUint32(responseData, &index, VersionNumberInitResponse)
		encoding.WriteUint64(responseData, &index, relay.LastUpdateTime)
		encoding.WriteBytes(responseData, &index, relay.PublicKey, crypto.KeySize)

		writer.Write(responseData[:index])
	}
}

// RelayUpdateHandlerFunc returns the function for the relay update endpoint
func RelayUpdateHandlerFunc(logger log.Logger, redisClient *redis.Client, statsdb *routing.StatsDatabase, metricsHandler metrics.Handler) func(writer http.ResponseWriter, request *http.Request) {
	logger = log.With(logger, "handler", "update")

	return func(writer http.ResponseWriter, request *http.Request) {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			return
		}

		index := 0

		relayUpdatePacket := RelayUpdatePacket{}
		if err = relayUpdatePacket.UnmarshalBinary(body); err != nil {
			level.Error(logger).Log("msg", "could not unmarshal packet", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		locallogger := log.With(logger, "req_addr", request.RemoteAddr, "relay_addr", relayUpdatePacket.Address.String())

		if relayUpdatePacket.Version != VersionNumberUpdateRequest || relayUpdatePacket.NumRelays > MaxRelays {
			level.Error(locallogger).Log("msg", "version mismatch", "version", relayUpdatePacket.Version)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		relay := routing.Relay{
			ID: crypto.HashID(relayUpdatePacket.Address.String()),
		}

		exists := redisClient.HExists(routing.HashKeyAllRelays, relay.Key())

		if exists.Err() != nil && exists.Err() != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to check if relay is registered", "err", exists.Err())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !exists.Val() {
			level.Warn(locallogger).Log("msg", "relay not initialized")
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		hgetResult := redisClient.HGet(routing.HashKeyAllRelays, relay.Key())
		if hgetResult.Err() != nil && hgetResult.Err() != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to get relays", "err", exists.Err())
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		data, err := hgetResult.Bytes()
		if err != nil && err != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to get relay data", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = relay.UnmarshalBinary(data); err != nil {
			level.Error(locallogger).Log("msg", "failed to unmarshal relay data", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		if !bytes.Equal(relayUpdatePacket.Token, relay.PublicKey) {
			level.Error(locallogger).Log("msg", "failed to get public key")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		statsUpdate := &routing.RelayStatsUpdate{}
		statsUpdate.ID = relay.ID
		statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdatePacket.PingStats...)

		statsdb.ProcessStats(statsUpdate)

		relay.LastUpdateTime = uint64(time.Now().Unix())

		type RelayPingData struct {
			id      uint64
			address string
		}

		relaysToPing := make([]RelayPingData, 0)

		redisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay)

		hgetallResult := redisClient.HGetAll(routing.HashKeyAllRelays)
		if hgetallResult.Err() != nil && hgetallResult.Err() != redis.Nil {
			level.Error(locallogger).Log("msg", "failed to get other relays", "err", hgetallResult.Err())
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		for k, v := range hgetallResult.Val() {
			if k != relay.Key() {
				var unmarshaledValue routing.Relay
				if err := unmarshaledValue.UnmarshalBinary([]byte(v)); err != nil {
					level.Error(locallogger).Log("msg", "failed to get other relay", "err", err)
					continue
				}
				relaysToPing = append(relaysToPing, RelayPingData{id: uint64(unmarshaledValue.ID), address: unmarshaledValue.Addr.String()})
			}
		}

		responseData := make([]byte, 10*1024)

		index = 0

		encoding.WriteUint32(responseData, &index, VersionNumberUpdateResponse)
		encoding.WriteUint32(responseData, &index, uint32(len(relaysToPing)))

		for i := range relaysToPing {
			encoding.WriteUint64(responseData, &index, relaysToPing[i].id)
			encoding.WriteString(responseData, &index, relaysToPing[i].address, MaxRelayAddressLength)
		}

		level.Debug(locallogger).Log("msg", "relay updated")

		responseLength := index

		writer.Header().Set("Content-Type", "application/octet-stream")

		writer.Write(responseData[:responseLength])
	}
}

// NearHandlerFunc returns the function for the near endpoint
func NearHandlerFunc(backend *StubbedBackend) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		backend.mutex.RLock()
		nearData := backend.nearData
		backend.mutex.RUnlock()
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(nearData)
	}
}
