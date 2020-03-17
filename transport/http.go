package transport

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"expvar"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/golang/protobuf/ptypes"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
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

// NewRouter creates a router with the specified endpoints
func NewRouter(logger log.Logger, redisClient *redis.Client, geoClient *routing.GeoClient, ipLocator routing.IPLocator, storer storage.Storer,
	statsdb *routing.StatsDatabase, initDuration metrics.Gauge, updateDuration metrics.Gauge, initCounter metrics.Counter,
	updateCounter metrics.Counter, costmatrix *routing.CostMatrix, routematrix *routing.RouteMatrix, routerPrivateKey []byte, trafficStatsPublisher stats.Publisher) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/healthz", HealthzHandlerFunc())
	router.HandleFunc("/relay_init", RelayInitHandlerFunc(logger, redisClient, geoClient, ipLocator, storer, initDuration, initCounter, routerPrivateKey)).Methods("POST")
	router.HandleFunc("/relay_update", RelayUpdateHandlerFunc(logger, redisClient, statsdb, updateDuration, updateCounter, trafficStatsPublisher, storer)).Methods("POST")
	router.Handle("/cost_matrix", costmatrix).Methods("GET")
	router.Handle("/route_matrix", routematrix).Methods("GET")
	router.HandleFunc("/near", NearHandlerFunc(nil)).Methods("GET")
	router.HandleFunc("/relay_init_json", RelayInitJSONHandlerFunc(logger, redisClient, geoClient, ipLocator, storer, initDuration, initCounter, routerPrivateKey)).Methods("POST")
	router.HandleFunc("/relay_update_json", RelayUpdateJSONHandlerFunc(logger, redisClient, statsdb, updateDuration, updateCounter, trafficStatsPublisher, storer)).Methods("POST")
	router.Handle("/debug/vars", expvar.Handler())
	return router
}

func relayInitPacketHandler(relayInitPacket *RelayInitPacket, writer http.ResponseWriter, request *http.Request, logger log.Logger, redisClient *redis.Client, geoClient *routing.GeoClient, ipLocator routing.IPLocator, storer storage.Storer, routerPrivateKey []byte) *routing.Relay {
	locallogger := log.With(logger, "req_addr", request.RemoteAddr, "relay_addr", relayInitPacket.Address.String())

	if relayInitPacket.Magic != InitRequestMagic {
		level.Error(locallogger).Log("msg", "magic number mismatch", "magic_number", relayInitPacket.Magic)
		writer.WriteHeader(http.StatusBadRequest)
		return nil
	}

	if relayInitPacket.Version != VersionNumberInitRequest {
		level.Error(locallogger).Log("msg", "version mismatch", "version", relayInitPacket.Version)
		writer.WriteHeader(http.StatusBadRequest)
		return nil
	}

	id := crypto.HashID(relayInitPacket.Address.String())

	relayEntry, ok := storer.Relay(id)
	if !ok {
		level.Error(locallogger).Log("msg", "relay not in configstore")
		writer.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	relay := routing.Relay{
		ID:             id,
		Addr:           relayInitPacket.Address,
		PublicKey:      relayEntry.PublicKey,
		Datacenter:     relayEntry.Datacenter,
		Seller:         relayEntry.Seller,
		LastUpdateTime: uint64(time.Now().Unix()),
	}

	if _, ok := crypto.Open(relayInitPacket.EncryptedToken, relayInitPacket.Nonce, relay.PublicKey, routerPrivateKey); !ok {
		level.Error(locallogger).Log("msg", "crypto open failed")
		writer.WriteHeader(http.StatusUnauthorized)
		return nil
	}

	exists := redisClient.HExists(routing.HashKeyAllRelays, relay.Key())

	if exists.Err() != nil && exists.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to check if relay is registered", "err", exists.Err())
		writer.WriteHeader(http.StatusNotFound)
		return nil
	}

	if exists.Val() {
		level.Warn(locallogger).Log("msg", "relay already initialized")
		writer.WriteHeader(http.StatusConflict)
		return nil
	}

	loc, err := ipLocator.LocateIP(relay.Addr.IP)
	if err != nil {
		level.Error(locallogger).Log("msg", "failed to locate relay")
		writer.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	relay.Latitude = loc.Latitude
	relay.Longitude = loc.Longitude

	// Regular set for expiry
	if res := redisClient.Set(relay.Key(), relay.ID, routing.RelayTimeout); res.Err() != nil && res.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to initialize relay", "err", res.Err())
		writer.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	// HSet for full relay data
	if res := redisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay); res.Err() != nil && res.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to initialize relay", "err", res.Err())
		writer.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	if err := geoClient.Add(relay); err != nil {
		level.Error(locallogger).Log("msg", "failed to initialize relay", "err", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	level.Debug(locallogger).Log("msg", "relay initialized")

	return &relay
}

func HealthzHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

// RelayInitHandlerFunc returns the function for the relay init endpoint
func RelayInitHandlerFunc(logger log.Logger, redisClient *redis.Client, geoClient *routing.GeoClient, ipLocator routing.IPLocator, storer storage.Storer, duration metrics.Gauge, counter metrics.Counter, routerPrivateKey []byte) func(writer http.ResponseWriter, request *http.Request) {
	logger = log.With(logger, "handler", "init")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			duration.Set(float64(durationSince.Milliseconds()))
			counter.Add(1)
		}()

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

		log.With(logger, "req_addr", request.RemoteAddr, "relay_addr", relayInitPacket.Address.String(), "packet_addr", relayInitPacket.Address.String())

		relay := relayInitPacketHandler(&relayInitPacket, writer, request, logger, redisClient, geoClient, ipLocator, storer, routerPrivateKey)
		if relay == nil {
			return
		}

		index := 0
		responseData := make([]byte, 64)
		encoding.WriteUint32(responseData, &index, VersionNumberInitResponse)
		encoding.WriteUint64(responseData, &index, relay.LastUpdateTime)
		encoding.WriteBytes(responseData, &index, relay.PublicKey, crypto.KeySize)

		writer.Header().Set("Content-Type", "application/octet-stream")

		writer.Write(responseData[:index])
	}
}

// RelayInitJSONHandlerFunc handles relay init data in json form
// currently it just converts the json struct into a packet struct and processes that
func RelayInitJSONHandlerFunc(logger log.Logger, redisClient *redis.Client, geoClient *routing.GeoClient, ipLocator routing.IPLocator, storer storage.Storer, duration metrics.Gauge, counter metrics.Counter, routerPrivateKey []byte) func(writer http.ResponseWriter, request *http.Request) {
	logger = log.With(logger, "handler", "init_json")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			duration.Set(float64(durationSince.Milliseconds()))
			counter.Add(1)
		}()

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			return
		}

		var jsonData RelayInitRequestJSON
		if err := json.Unmarshal(body, &jsonData); err != nil {
			level.Error(logger).Log("msg", "could not parse init json", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		var packet RelayInitPacket

		if err := jsonData.ToInitPacket(&packet); err != nil {
			level.Error(logger).Log("msg", "could not convert json data to binary packet", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		relay := relayInitPacketHandler(&packet, writer, request, logger, redisClient, geoClient, ipLocator, storer, routerPrivateKey)

		if relay == nil {
			// the packet handler func will have set the writer's status and logged something, so just log it was json and return
			level.Error(logger).Log("msg", "could not process relay data from json handler")
			return
		}

		var response RelayInitResponseJSON
		response.Timestamp = relay.LastUpdateTime * 1000 // convert to millis, this is what the curr prod relay expects, will have to change when using new relay, new relay just uses seconds

		var dat []byte
		if dat, err = json.Marshal(response); err != nil {
			level.Error(logger).Log("msg", "could not marshal init json response", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		writer.Header().Set("Content-Type", "application/json")

		writer.Write(dat)
	}
}

func relayUpdatePacketHandler(relayUpdatePacket *RelayUpdatePacket, writer http.ResponseWriter, request *http.Request, logger log.Logger, redisClient *redis.Client, statsdb *routing.StatsDatabase) []routing.RelayPingData {
	locallogger := log.With(logger, "req_addr", request.RemoteAddr, "relay_addr", relayUpdatePacket.Address.String())

	if relayUpdatePacket.Version != VersionNumberUpdateRequest || relayUpdatePacket.NumRelays > MaxRelays {
		level.Error(locallogger).Log("msg", "version mismatch", "version", relayUpdatePacket.Version)
		writer.WriteHeader(http.StatusBadRequest)
		return nil
	}

	relay := routing.Relay{
		ID: crypto.HashID(relayUpdatePacket.Address.String()),
	}

	exists := redisClient.HExists(routing.HashKeyAllRelays, relay.Key())

	if exists.Err() != nil && exists.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to check if relay is registered", "err", exists.Err())
		writer.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	if !exists.Val() {
		level.Warn(locallogger).Log("msg", "relay not initialized")
		writer.WriteHeader(http.StatusNotFound)
		return nil
	}

	hgetResult := redisClient.HGet(routing.HashKeyAllRelays, relay.Key())
	if hgetResult.Err() != nil && hgetResult.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to get relays", "err", exists.Err())
		writer.WriteHeader(http.StatusNotFound)
		return nil
	}

	data, err := hgetResult.Bytes()
	if err != nil && err != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to get relay data", "err", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	if err = relay.UnmarshalBinary(data); err != nil {
		level.Error(locallogger).Log("msg", "failed to unmarshal relay data", "err", err)
		writer.WriteHeader(http.StatusBadRequest)
		return nil
	}

	if !bytes.Equal(relayUpdatePacket.Token, relay.PublicKey) {
		level.Error(locallogger).Log("msg", "relay public key doesn't match")
		writer.WriteHeader(http.StatusBadRequest)
		return nil
	}

	statsUpdate := &routing.RelayStatsUpdate{}
	statsUpdate.ID = relay.ID
	statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdatePacket.PingStats...)

	statsdb.ProcessStats(statsUpdate)

	relay.LastUpdateTime = uint64(time.Now().Unix())

	relaysToPing := make([]routing.RelayPingData, 0)

	// Regular set for expiry
	if res := redisClient.Set(relay.Key(), 0, routing.RelayTimeout); res.Err() != nil {
		level.Error(locallogger).Log("msg", "failed to store relay update expiry", "err", res.Err())
		writer.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	// HSet for full relay data
	if res := redisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay); res.Err() != nil {
		level.Error(locallogger).Log("msg", "failed to store relay update", "err", res.Err())
		writer.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	hgetallResult := redisClient.HGetAll(routing.HashKeyAllRelays)
	if hgetallResult.Err() != nil && hgetallResult.Err() != redis.Nil {
		level.Error(locallogger).Log("msg", "failed to get other relays", "err", hgetallResult.Err())
		writer.WriteHeader(http.StatusNotFound)
		return nil
	}

	for k, v := range hgetallResult.Val() {
		if k != relay.Key() {
			var unmarshaledValue routing.Relay
			if err := unmarshaledValue.UnmarshalBinary([]byte(v)); err != nil {
				level.Error(locallogger).Log("msg", "failed to get other relay", "err", err)
				continue
			}
			relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(unmarshaledValue.ID), Address: unmarshaledValue.Addr.String()})
		}
	}

	level.Debug(locallogger).Log("msg", "relay updated")

	return relaysToPing
}

// RelayUpdateHandlerFunc returns the function for the relay update endpoint
func RelayUpdateHandlerFunc(logger log.Logger, redisClient *redis.Client, statsdb *routing.StatsDatabase, duration metrics.Gauge, counter metrics.Counter, trafficStatsPublisher stats.Publisher, storer storage.Storer) func(writer http.ResponseWriter, request *http.Request) {
	logger = log.With(logger, "handler", "update")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			duration.Set(float64(durationSince.Milliseconds()))
			counter.Add(1)
		}()

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		index := 0

		relayUpdatePacket := RelayUpdatePacket{}
		if err = relayUpdatePacket.UnmarshalBinary(body); err != nil {
			level.Error(logger).Log("msg", "could not unmarshal packet", "err", err)
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		relaysToPing := relayUpdatePacketHandler(&relayUpdatePacket, writer, request, logger, redisClient, statsdb)

		responseData := make([]byte, 10*1024)

		index = 0

		encoding.WriteUint32(responseData, &index, VersionNumberUpdateResponse)
		encoding.WriteUint32(responseData, &index, uint32(len(relaysToPing)))

		for i := range relaysToPing {
			encoding.WriteUint64(responseData, &index, relaysToPing[i].ID)
			encoding.WriteString(responseData, &index, relaysToPing[i].Address, MaxRelayAddressLength)
		}

		responseLength := index

		writer.Header().Set("Content-Type", "application/octet-stream")

		writer.Write(responseData[:responseLength])

		relayID := crypto.HashID(relayUpdatePacket.Address.String())
		if relay, ok := storer.Relay(relayID); ok {
			stats := &stats.RelayTrafficStats{
				RelayId:            stats.NewEntityID("Relay", relay.ID),
				BytesMeasurementRx: relayUpdatePacket.BytesReceived,
			}

			if err := trafficStatsPublisher.Publish(context.Background(), relay.ID, stats); err != nil {
				level.Error(logger).Log("msg", fmt.Sprintf("Publish error: %v", err))
			}
		} else {
			level.Error(logger).Log("msg", fmt.Sprintf("TrafficStats, cannot lookup relay in firestore, %d", relayID))
			return
		}
	}
}

// RelayUpdateJSONHandlerFunc handles processing json from the relays
// currently it just converts the json into a packet and passes it to a common function
func RelayUpdateJSONHandlerFunc(logger log.Logger, redisClient *redis.Client, statsdb *routing.StatsDatabase, duration metrics.Gauge, counter metrics.Counter, trafficStatsPublisher stats.Publisher, storer storage.Storer) func(writer http.ResponseWriter, request *http.Request) {
	logger = log.With(logger, "handler", "update_json")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			duration.Set(float64(durationSince.Milliseconds()))
			counter.Add(1)
		}()

		level.Error(logger).Log("msg", "received json packet")

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		var jsonPacket RelayUpdateRequestJSON
		if err := json.Unmarshal(body, &jsonPacket); err != nil {
			level.Error(logger).Log("msg", "could not parse update json", "err", err)
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		var packet RelayUpdatePacket
		if err := jsonPacket.ToUpdatePacket(&packet); err != nil {
			level.Error(logger).Log("msg", "could not convert json", "err", err)
			writer.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		relaysToPing := relayUpdatePacketHandler(&packet, writer, request, logger, redisClient, statsdb)
		if relaysToPing == nil {
			level.Error(logger).Log("msg", "could not process converted packet")
			return
		}

		var response RelayUpdateResponseJSON
		for _, pingData := range relaysToPing {
			var token routing.LegacyPingToken
			token.Timeout = uint64(time.Now().Unix() * 100000) // some arbitrarily large number just to make things compatable for the moment
			token.RelayID = crypto.HashID(jsonPacket.StringAddr)
			bin, _ := token.MarshalBinary()
			var legacy routing.LegacyPingData
			legacy.ID = pingData.ID
			legacy.Address = pingData.Address
			legacy.PingToken = base64.StdEncoding.EncodeToString(bin)
			response.RelaysToPing = append(response.RelaysToPing, legacy)
		}

		var dat []byte
		if dat, err = json.Marshal(response); err != nil {
			level.Error(logger).Log("msg", "could not marshal update json response", "err", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		writer.Header().Set("Content-Type", "application/json")

		writer.Write(dat)

		if ts, err := ptypes.TimestampProto(time.Unix(int64(jsonPacket.Timestamp), 0)); err == nil {
			if relay, ok := storer.Relay(jsonPacket.Metadata.ID); ok {
				stats := &stats.RelayTrafficStats{
					RelayId:            stats.NewEntityID("Relay", relay.ID),
					Usage:              jsonPacket.Usage,
					Timestamp:          ts,
					BytesPaidTx:        jsonPacket.TrafficStats.BytesPaidTx,
					BytesPaidRx:        jsonPacket.TrafficStats.BytesPaidRx,
					BytesManagementTx:  jsonPacket.TrafficStats.BytesManagementTx,
					BytesManagementRx:  jsonPacket.TrafficStats.BytesManagementRx,
					BytesMeasurementTx: jsonPacket.TrafficStats.BytesMeasurementTx,
					BytesMeasurementRx: jsonPacket.TrafficStats.BytesMeasurementRx,
					BytesInvalidRx:     jsonPacket.TrafficStats.BytesInvalidRx,
					SessionCount:       jsonPacket.TrafficStats.SessionCount,
				}

				if err := trafficStatsPublisher.Publish(context.Background(), relay.ID, stats); err != nil {
					level.Error(logger).Log("msg", fmt.Sprintf("Publish error: %v", err))
				}
			} else {
				level.Error(logger).Log("msg", fmt.Sprintf("TrafficStats, cannot lookup relay in firestore, %d", jsonPacket.Metadata.ID))
				return
			}
		} else {
			level.Error(logger).Log("msg", fmt.Sprintf("Can't publish to pubsub. Timestamp error: %v", err))
			return
		}
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
