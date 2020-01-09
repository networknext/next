package transport

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
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
)

var gRelayPublicKey = []byte{
	0xf5, 0x22, 0xad, 0xc1, 0xee, 0x04, 0x6a, 0xbe,
	0x7d, 0x89, 0x0c, 0x81, 0x3a, 0x08, 0x31, 0xba,
	0xdc, 0xdd, 0xb5, 0x52, 0xcb, 0x73, 0x56, 0x10,
	0xda, 0xa9, 0xc0, 0xae, 0x08, 0xa2, 0xcf, 0x5e,
}

// NewRouter creates a router with the specified endpoints
func NewRouter(relaydb *core.RelayDatabase, statsdb *core.StatsDatabase, backend *StubbedBackend) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/relay_init", RelayInitHandlerFunc(relaydb)).Methods("POST")
	router.HandleFunc("/relay_update", RelayUpdateHandlerFunc(relaydb, statsdb)).Methods("POST")
	router.HandleFunc("/cost_matrix", CostMatrixHandlerFunc(backend)).Methods("GET")
	router.HandleFunc("/route_matrix", RouteMatrixHandlerFunc(backend)).Methods("GET")
	router.HandleFunc("/near", NearHandlerFunc(backend)).Methods("GET")
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
func RelayInitHandlerFunc(relaydb *core.RelayDatabase) func(writer http.ResponseWriter, request *http.Request) {
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
			relayInitPacket.Version != VersionNumberInitRequest ||
			!crypto.Check(relayInitPacket.EncryptedToken, relayInitPacket.Nonce, gRelayPublicKey[:], core.RouterPrivateKey[:]) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		key := core.GetRelayId(relayInitPacket.Address)

		_, relayAlreadyExists := relaydb.Relays[key]
		if relayAlreadyExists {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		entry := core.RelayData{}
		entry.Name = relayInitPacket.Address
		entry.Id = core.GetRelayId(relayInitPacket.Address)
		entry.Address = relayInitPacket.Address //core.ParseAddress(relayInitPacket.address)
		entry.LastUpdateTime = uint64(time.Now().Unix())
		entry.PublicKey = core.RandomBytes(LengthOfRelayToken)

		relaydb.Relays[entry.Id] = entry

		writer.Header().Set("Content-Type", "application/octet-stream")

		index = 0
		responseData := make([]byte, 64)
		encoding.WriteUint32(responseData, &index, VersionNumberInitResponse)
		encoding.WriteUint64(responseData, &index, uint64(time.Now().Unix()))
		encoding.WriteBytes(responseData, &index, entry.PublicKey, LengthOfRelayToken)

		writer.Write(responseData[:index])
	}
}

// RelayUpdateHandlerFunc returns the function fora the relay update endpoint
func RelayUpdateHandlerFunc(relaydb *core.RelayDatabase, statsdb *core.StatsDatabase) func(writer http.ResponseWriter, request *http.Request) {
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

		key := core.GetRelayId(relayUpdatePacket.Address)
		entry, ok := relaydb.Relays[key]
		found := false
		if ok && crypto.CompareTokens(relayUpdatePacket.Token, entry.PublicKey) {
			found = true
		}

		if !found {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		statsUpdate := &core.RelayStatsUpdate{}
		statsUpdate.Id = entry.Id

		for _, ps := range relayUpdatePacket.PingStats {
			statsUpdate.PingStats = append(statsUpdate.PingStats, ps)
		}

		statsdb.ProcessStats(statsUpdate)

		entry = core.RelayData{
			Name:           relayUpdatePacket.Address,
			Id:             core.GetRelayId(relayUpdatePacket.Address),
			Address:        relayUpdatePacket.Address,
			LastUpdateTime: uint64(time.Now().Unix()),
			PublicKey:      relayUpdatePacket.Token,
		}

		type RelayPingData struct {
			id      uint64
			address string
		}

		relaysToPing := make([]RelayPingData, 0)

		relaydb.Relays[key] = entry
		hashedAddress := core.GetRelayId(relayUpdatePacket.Address)
		for k, v := range relaydb.Relays {
			if k != hashedAddress {
				relaysToPing = append(relaysToPing, RelayPingData{id: uint64(v.Id), address: v.Address})
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

		responseData = responseData[:responseLength]

		writer.Header().Set("Content-Type", "application/octet-stream")

		writer.Write(responseData)
	}
}

// CostMatrixHandlerFunc returns the function for the cost matrix endpoint
func CostMatrixHandlerFunc(backend *StubbedBackend) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		backend.mutex.RLock()
		costMatrixData := backend.costMatrixData
		backend.mutex.RUnlock()
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(costMatrixData)
	}
}

// RouteMatrixHandlerFunc returns the function for the matrix endpoint
func RouteMatrixHandlerFunc(backend *StubbedBackend) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		backend.mutex.RLock()
		routeMatrixData := backend.routeMatrixData
		backend.mutex.RUnlock()
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(routeMatrixData)
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
