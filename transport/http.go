package transport

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
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

var gRelayPublicKey = []byte{
	0xf5, 0x22, 0xad, 0xc1, 0xee, 0x04, 0x6a, 0xbe,
	0x7d, 0x89, 0x0c, 0x81, 0x3a, 0x08, 0x31, 0xba,
	0xdc, 0xdd, 0xb5, 0x52, 0xcb, 0x73, 0x56, 0x10,
	0xda, 0xa9, 0xc0, 0xae, 0x08, 0xa2, 0xcf, 0x5e,
}

type RelayProvider interface {
	GetAndCheckByRelayCoreId(key uint32) (*storage.Relay, bool)
}

type DatacenterProvider interface {
	GetAndCheck(key *storage.Key) (*storage.Datacenter, bool)
}

// NewRouter creates a router with the specified endpoints
func NewRouter(redisClient *redis.Client, geoClient *routing.GeoClient, ipLocator routing.IPLocator, relayProvider RelayProvider, datacenterProvider DatacenterProvider, statsdb *routing.StatsDatabase, costmatrix *routing.CostMatrix, routematrix *routing.RouteMatrix, relayPublicKey []byte, routerPrivateKey []byte) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/relay_init", RelayInitHandlerFunc(redisClient, geoClient, ipLocator, relayProvider, datacenterProvider, relayPublicKey, routerPrivateKey)).Methods("POST")
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
func RelayInitHandlerFunc(redisClient *redis.Client, geoClient *routing.GeoClient, ipLocator routing.IPLocator, relayProvider RelayProvider, datacenterProvider DatacenterProvider, relayPublicKey []byte, routerPrivateKey []byte) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Println("Received Relay Init Packet")
		body, err := ioutil.ReadAll(request.Body)

		if err != nil {
			log.Printf("Could not read init packet: %v", err)
			return
		}

		index := 0

		relayInitPacket := RelayInitPacket{}

		if err = relayInitPacket.UnmarshalBinary(body); err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			log.Printf("Could not read init packet: %v", err)
			return
		}

		if relayInitPacket.Magic != InitRequestMagic ||
			relayInitPacket.Version != VersionNumberInitRequest {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		if _, ok := crypto.Open(relayInitPacket.EncryptedToken, relayInitPacket.Nonce, relayPublicKey, routerPrivateKey); !ok {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		relay := routing.Relay{
			ID:             crypto.HashID(relayInitPacket.Address.String()),
			Addr:           relayInitPacket.Address,
			LastUpdateTime: uint64(time.Now().Unix()),
		}

		if dcID, err := RelayIDToDatacenterIDFunc(&relay, relayProvider, datacenterProvider); err != nil {
			log.Printf("Unable to get datacenter ID for relay '%s': %v", relay.Addr.String(), err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			relay.Datacenter = dcID
		}

		exists := redisClient.HExists(routing.HashKeyAllRelays, relay.Key())

		if exists.Err() != nil && exists.Err() != redis.Nil {
			log.Printf("failed to get relay %s from redis: %v", relayInitPacket.Address.String(), exists.Err())
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
		relay.PublicKey = make([]byte, crypto.KeySize)
		rand.Read(relay.PublicKey)

		loc, err := ipLocator.LocateIP(relay.Addr.IP)
		if err != nil {
			log.Printf("failed to lookup relay ip '%s': %v", relay.Addr.String(), err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		relay.Latitude = loc.Latitude
		relay.Longitude = loc.Longitude

		if res := redisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay); res.Err() != nil && res.Err() != redis.Nil {
			log.Printf("failed to set relay %s into redis hash: %v", relayInitPacket.Address.String(), res.Err())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := geoClient.Add(relay); err != nil {
			log.Printf("failed to add relay %s into geo client: %v", relayInitPacket.Address.String(), err)
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

// RelayUpdateHandlerFunc returns the function for the relay update endpoint
func RelayUpdateHandlerFunc(redisClient *redis.Client, statsdb *routing.StatsDatabase) func(writer http.ResponseWriter, request *http.Request) {
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
			log.Printf("Could not read update packet: %v", err)
			return
		}

		if relayUpdatePacket.Version != VersionNumberUpdateRequest || relayUpdatePacket.NumRelays > MaxRelays {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		relay := routing.Relay{
			ID: crypto.HashID(relayUpdatePacket.Address.String()),
		}

		exists := redisClient.HExists(routing.HashKeyAllRelays, relay.Key())

		if exists.Err() != nil && exists.Err() != redis.Nil {
			log.Printf("failed to check if relay %s exists: %v", relayUpdatePacket.Address.String(), exists.Err())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !exists.Val() {
			log.Printf("failed to find relay with address '%s' in redis", relayUpdatePacket.Address.String())
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		hgetResult := redisClient.HGet(routing.HashKeyAllRelays, relay.Key())
		if hgetResult.Err() != nil && hgetResult.Err() != redis.Nil {
			log.Printf("failed to get relay %s from redis: %v", relayUpdatePacket.Address.String(), hgetResult.Err())
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		data, err := hgetResult.Bytes()

		if err != nil && err != redis.Nil {
			log.Printf("failed to get bytes from redis: %v", hgetResult.Err())
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = relay.UnmarshalBinary(data); err != nil {
			log.Printf("failed to marshal data into struct: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		if !bytes.Equal(relayUpdatePacket.Token, relay.PublicKey) {
			log.Printf("update packet for address '%s' not equal to existing entry", relayUpdatePacket.Address.String())
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
			log.Printf("failed to get all relays from redis: %v", hgetallResult.Err())
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		for k, v := range hgetallResult.Val() {
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

// Dev debug logic
type fakeRelayData struct {
	Latitude     float64
	Longitude    float64
	DatacenterID uint64
}

var StubbedRelayData map[int]fakeRelayData

var RelayIDToDatacenterIDFunc func(relay *routing.Relay, rp RelayProvider, dp DatacenterProvider) (uint64, error) = ReleaseRelayIDToDatacenterIDFunc

func DebugRelayIDToDatacenterIDFunc(relay *routing.Relay, rp RelayProvider, dp DatacenterProvider) (uint64, error) {
	fakeData, ok := StubbedRelayData[relay.Addr.Port]
	if ok {
		fmt.Printf("Found fake datacenter for port %d\n", relay.Addr.Port)
		return fakeData.DatacenterID, nil
	} else {
		return 0, nil
	}
}

func ReleaseRelayIDToDatacenterIDFunc(relay *routing.Relay, rp RelayProvider, dp DatacenterProvider) (uint64, error) {
	// TODO config store will have to use uint64's at some later point
	dbRelay, ok := rp.GetAndCheckByRelayCoreId(uint32(relay.ID))

	if !ok {
		return 0, errors.New("relay not found in configstore")
	}

	dbDatacenter, ok := dp.GetAndCheck(dbRelay.Datacenter)

	if !ok {
		return 0, errors.New("datacenter not found in configstore")
	}

	return crypto.HashID(dbDatacenter.Name), nil
}

var IpLookupFunc func(relay *routing.Relay, ipLocator routing.IPLocator) (routing.Location, error) = ReleaseIPLookupFunc

func DebugIPLookupFunc(relay *routing.Relay, ipLocator routing.IPLocator) (routing.Location, error) {
	fakeData, ok := StubbedRelayData[relay.Addr.Port]
	if ok {
		fmt.Printf("Found fake location for port %d\n", relay.Addr.Port)
		return routing.Location{
			Latitude:  fakeData.Latitude,
			Longitude: fakeData.Longitude,
		}, nil
	} else {
		return routing.Location{
			Latitude:  0,
			Longitude: 0,
		}, nil
	}
}

func ReleaseIPLookupFunc(relay *routing.Relay, ipLocator routing.IPLocator) (routing.Location, error) {
	return ipLocator.LocateIP(relay.Addr.IP)
}
