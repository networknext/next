package transport

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

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
)

// NewRouter creates a router with the specified endpoints
func NewRouter(relaydb *core.RelayDatabase, statsdb *core.StatsDatabase, costmatrix *routing.CostMatrix, routematrix *routing.RouteMatrix) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/relay_init", RelayInitHandlerFunc(relaydb)).Methods("POST")
	router.HandleFunc("/relay_update", RelayUpdateHandlerFunc(relaydb, statsdb)).Methods("POST")
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
			relayInitPacket.Version != VersionNumberInitRequest {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		if _, ok := crypto.Open(relayInitPacket.EncryptedToken, relayInitPacket.Nonce, crypto.RelayPublicKey[:], crypto.RouterPrivateKey[:]); !ok {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		key := core.GetRelayID(relayInitPacket.Address)

		_, relayAlreadyExists := relaydb.Relays[key]
		if relayAlreadyExists {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		entry := core.RelayData{}
		entry.Name = relayInitPacket.Address
		entry.ID = core.GetRelayID(relayInitPacket.Address)
		entry.Address = relayInitPacket.Address //core.ParseAddress(relayInitPacket.address)
		entry.LastUpdateTime = uint64(time.Now().Unix())

		entry.PublicKey = make([]byte, crypto.KeySize)
		if _, err := rand.Read(entry.PublicKey); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		relaydb.Relays[entry.ID] = entry

		writer.Header().Set("Content-Type", "application/octet-stream")

		index = 0
		responseData := make([]byte, 64)
		encoding.WriteUint32(responseData, &index, VersionNumberInitResponse)
		encoding.WriteUint64(responseData, &index, uint64(time.Now().Unix()))
		encoding.WriteBytes(responseData, &index, entry.PublicKey, crypto.KeySize)

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

		if relayUpdatePacket.Version != VersionNumberUpdateRequest || relayUpdatePacket.NumRelays == 0 || relayUpdatePacket.NumRelays > MaxRelays {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		key := core.GetRelayID(relayUpdatePacket.Address)
		entry, ok := relaydb.Relays[key]
		found := false
		if ok && bytes.Equal(relayUpdatePacket.Token, entry.PublicKey) {
			found = true
		}

		if !found {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		statsUpdate := &core.RelayStatsUpdate{}
		statsUpdate.ID = entry.ID
		statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdatePacket.PingStats...)

		statsdb.ProcessStats(statsUpdate)

		entry = core.RelayData{
			Name:           relayUpdatePacket.Address,
			ID:             core.GetRelayID(relayUpdatePacket.Address),
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
		hashedAddress := core.GetRelayID(relayUpdatePacket.Address)
		for k, v := range relaydb.Relays {
			if k != hashedAddress {
				relaysToPing = append(relaysToPing, RelayPingData{id: uint64(v.ID), address: v.Address})
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
