package transport

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
)

const (
	InitRequestMagic = uint32(0x9083708f)

	LengthOfRelayToken = 32

	MaxRelays             = 1024
	MaxRelayAddressLength = 256

	VersionNumberInitRequest    = 0
	VersionNumberInitResponse   = 0
	VersionNumberUpdateRequest  = 0
	VersionNumberUpdateResponse = 0

	RedisHashName = "ALL_RELAYS"
)

var gRelayPublicKey = []byte{
	0xf5, 0x22, 0xad, 0xc1, 0xee, 0x04, 0x6a, 0xbe,
	0x7d, 0x89, 0x0c, 0x81, 0x3a, 0x08, 0x31, 0xba,
	0xdc, 0xdd, 0xb5, 0x52, 0xcb, 0x73, 0x56, 0x10,
	0xda, 0xa9, 0xc0, 0xae, 0x08, 0xa2, 0xcf, 0x5e,
}

// NewRouter creates a router with the specified endpoints
func NewRouter(redisClient *redis.Client, statsdb *core.StatsDatabase, costmatrix *routing.CostMatrix, routematrix *routing.RouteMatrix) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/relay_init", RelayInitHandlerFunc(redisClient)).Methods("POST")
	router.HandleFunc("/relay_update", RelayUpdateHandlerFunc(redisClient, statsdb)).Methods("POST")
	router.Handle("/cost_matrix", costmatrix).Methods("GET")
	router.Handle("/route_matrix", routematrix).Methods("GET")
	router.HandleFunc("/near", NearHandlerFunc(nil)).Methods("GET")
	return router
}

// HTTPStart starts a http server on the supplied port with the supplied router
func HTTPStart(port string, router *mux.Router) {
	log.Printf("Starting server with port %s\n", port) // log
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), router)
	if err != nil {
		fmt.Println(err)
	}
}

// RelayInitHandlerFunc returns the function for the relay init endpoint
func RelayInitHandlerFunc(redisClient *redis.Client) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Println("Received Relay Init Packet")
		body, err := ioutil.ReadAll(request.Body)

		if err != nil {
			return
		}

		index := 0

		relayInitPacket := RelayInitPacket{}

		if err = relayInitPacket.UnmarshalBinary(body); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		if relayInitPacket.Magic != InitRequestMagic ||
			relayInitPacket.Version != VersionNumberInitRequest {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		if _, ok := crypto.Open(relayInitPacket.EncryptedToken, relayInitPacket.Nonce, crypto.RelayPublicKey[:], crypto.RouterPrivateKey[:]); !ok {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		udpAddr, _ := net.ResolveUDPAddr("udp", relayInitPacket.Address)
		relay := routing.Relay{
			ID:             core.GetRelayID(relayInitPacket.Address),
			Addr:           *udpAddr,
			LastUpdateTime: uint64(time.Now().Unix()),
		}

		exists := redisClient.HExists(RedisHashName, relay.Key())

		if exists.Err() != nil && exists.Err() != redis.Nil {
			log.Printf("failed to get relay %s from redis: %v", relayInitPacket.Address, exists.Err())
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		if exists.Val() {
			log.Println("relay entry exists, returning")
			writer.WriteHeader(http.StatusNotFound)
		}

		relay.PublicKey = make([]byte, crypto.KeySize)
		if _, err := rand.Read(relay.PublicKey); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		relay.LastUpdateTime = uint64(time.Now().Unix())
		relay.PublicKey = core.RandomBytes(LengthOfRelayToken)

		res := redisClient.HSet(RedisHashName, relay.Key(), relay)

		if res.Err() != nil && res.Err() != redis.Nil {
			log.Printf("failed to set relay %s into redis hash: %v", relayInitPacket.Address, res.Err())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/octet-stream")

		index = 0
		responseData := make([]byte, 64)
		encoding.WriteUint32(responseData, &index, VersionNumberInitResponse)
		encoding.WriteUint64(responseData, &index, uint64(time.Now().Unix()))
		encoding.WriteBytes(responseData, &index, relay.PublicKey, crypto.KeySize)

		writer.Write(responseData[:index])
	}
}

// RelayUpdateHandlerFunc returns the function fora the relay update endpoint
func RelayUpdateHandlerFunc(redisClient *redis.Client, statsdb *core.StatsDatabase) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Println("Received Relay Update Packet")
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return
		}

		index := 0

		relayUpdatePacket := RelayUpdatePacket{}
		if err = relayUpdatePacket.UnmarshalBinary(body); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		if relayUpdatePacket.Version != VersionNumberUpdateRequest || relayUpdatePacket.NumRelays > MaxRelays {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		relay := routing.Relay{
			ID: core.GetRelayID(relayUpdatePacket.Address),
		}

		found := false
		exists := redisClient.HExists(RedisHashName, relay.Key())

		if exists.Err() != nil && exists.Err() != redis.Nil {
			log.Printf("failed to check if relay %s exists: %v", relayUpdatePacket.Address, exists.Err())
		}

		if exists.Val() {
			result := redisClient.HGet(RedisHashName, relay.Key())
			if result.Err() != nil && result.Err() != redis.Nil {
				log.Printf("failed to get relay %s from redis: %v", relayUpdatePacket.Address, result.Err())
				writer.WriteHeader(http.StatusNotFound)
				return
			}

			data, err := result.Bytes()

			if err != nil && err != redis.Nil {
				log.Printf("failed to get bytes from redis: %v", result.Err())
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			if err = relay.UnmarshalBinary(data); err != nil {
				log.Printf("failed to marshal data into struct: %v", err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}

			if bytes.Equal(relayUpdatePacket.Token, relay.PublicKey) {
				found = true
			}
		}

		if bytes.Equal(relayUpdatePacket.Token, relay.PublicKey) {
			found = true
		}

		if !found {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		statsUpdate := &core.RelayStatsUpdate{}
		statsUpdate.ID = relay.ID
		statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdatePacket.PingStats...)

		statsdb.ProcessStats(statsUpdate)

		relay.LastUpdateTime = uint64(time.Now().Unix())

		type RelayPingData struct {
			id      uint64
			address string
		}

		relaysToPing := make([]RelayPingData, 0)

		redisClient.HSet(RedisHashName, relay.Key(), relay)
		result := redisClient.HGetAll(RedisHashName)

		if result.Err() != nil && result.Err() != redis.Nil {
			log.Printf("failed to get all relays from redis: %v", result.Err())
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		for k, v := range result.Val() {
			if k != relay.Key() {
				var unmarshaledValue routing.Relay
				unmarshaledValue.UnmarshalBinary([]byte(v))
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
