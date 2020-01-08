package transport

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"fmt"
	"github.com/gorilla/mux"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/rw"
	// Relay entry
)

const gInitRequestMagic = uint32(0x9083708f)
const gInitRequestVersion = 0
const gInitResponseVersion = 0
const gUpdateRequestVersion = 0
const gUpdateResponseVersion = 0
const gRelayTokenBytes = 32
const gMaxRelays = 1024

var gRelayPublicKey = []byte{
	0xf5, 0x22, 0xad, 0xc1, 0xee, 0x04, 0x6a, 0xbe,
	0x7d, 0x89, 0x0c, 0x81, 0x3a, 0x08, 0x31, 0xba,
	0xdc, 0xdd, 0xb5, 0x52, 0xcb, 0x73, 0x56, 0x10,
	0xda, 0xa9, 0xc0, 0xae, 0x08, 0xa2, 0xcf, 0x5e,
}

func getRelayID(name string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(name))
	return hash.Sum64()
}

// MakeRouter creates a router with the specified endpoints
func MakeRouter(backend *Backend) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/relay_init", RelayInitHandlerFunc(backend)).Methods("POST")
	router.HandleFunc("/relay_update", RelayUpdateHandlerFunc(backend)).Methods("POST")
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
func RelayInitHandlerFunc(backend *Backend) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
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

		if !crypto.Check(relayInitPacket.encryptedToken, relayInitPacket.nonce, gRelayPublicKey[:], core.RouterPrivateKey[:]) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		key := relayInitPacket.address

		backend.Mutex.RLock()
		_, relayAlreadyExists := backend.RelayDatabase[key]
		backend.Mutex.RUnlock()
		if relayAlreadyExists {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		relayEntry := RelayEntry{}
		relayEntry.name = relayInitPacket.address
		relayEntry.id = getRelayID(relayInitPacket.address)
		relayEntry.address = core.ParseAddress(relayInitPacket.address)
		relayEntry.lastUpdate = time.Now().Unix()
		relayEntry.token = core.RandomBytes(gRelayTokenBytes)

		backend.Mutex.Lock()
		backend.RelayDatabase[key] = relayEntry
		backend.Dirty = true
		backend.Mutex.Unlock()

		writer.Header().Set("Content-Type", "application/octet-stream")

		index = 0
		responseData := make([]byte, 64)
		rw.WriteUint32(responseData, &index, gInitResponseVersion)
		rw.WriteUint64(responseData, &index, uint64(time.Now().Unix()))
		rw.WriteBytes(responseData, &index, relayEntry.token, gRelayTokenBytes)

		writer.Write(responseData[:index])
	}
}

// RelayUpdateHandlerFunc returns the function fora the relay update endpoint
func RelayUpdateHandlerFunc(backend *Backend) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return
		}

		index := 0

		var version uint32
		if !rw.ReadUint32(body, &index, &version) || version != gUpdateRequestVersion {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		var relayAddress string
		if !rw.ReadString(body, &index, &relayAddress, gMaxRelayAddressLength) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		var token []byte
		if !rw.ReadBytes(body, &index, &token, gRelayTokenBytes) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		key := relayAddress
		// --VerifyRelay()?--
		backend.Mutex.RLock()
		relayEntry, ok := backend.RelayDatabase[key]
		found := false
		if ok && crypto.CompareTokens(token, relayEntry.token) {
			found = true
		}
		backend.Mutex.RUnlock()
		// ------------------

		if !found {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		var numRelays uint32
		if !rw.ReadUint32(body, &index, &numRelays) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		if numRelays > gMaxRelays {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		statsUpdate := &core.RelayStatsUpdate{}
		statsUpdate.Id = core.RelayId(relayEntry.id)

		for i := 0; i < int(numRelays); i++ {
			var id uint64
			var rtt, jitter, packetLoss float32
			if !rw.ReadUint64(body, &index, &id) {
				return
			}
			if !rw.ReadFloat32(body, &index, &rtt) {
				return
			}
			if !rw.ReadFloat32(body, &index, &jitter) {
				return
			}
			if !rw.ReadFloat32(body, &index, &packetLoss) {
				return
			}

			ping := core.RelayStatsPing{}
			ping.RelayId = core.RelayId(id)
			ping.RTT = rtt
			ping.Jitter = jitter
			ping.PacketLoss = packetLoss
			statsUpdate.PingStats = append(statsUpdate.PingStats, ping)
		}

		backend.Mutex.Lock()
		backend.StatsDatabase.ProcessStats(statsUpdate)
		backend.Mutex.Unlock()

		relayEntry = RelayEntry{}
		relayEntry.name = relayAddress
		relayEntry.id = getRelayID(relayAddress)
		relayEntry.address = core.ParseAddress(relayAddress)
		relayEntry.lastUpdate = time.Now().Unix()
		relayEntry.token = token

		type RelayPingData struct {
			id      uint64
			address string
		}

		relaysToPing := make([]RelayPingData, 0)

		backend.Mutex.Lock()
		backend.RelayDatabase[key] = relayEntry
		for k, v := range backend.RelayDatabase {
			if k != relayAddress {
				if k != relayAddress {
					relaysToPing = append(relaysToPing, RelayPingData{id: v.id, address: k})
				}
			}
		}
		backend.Mutex.Unlock()

		responseData := make([]byte, 10*1024)

		index = 0

		rw.WriteUint32(responseData, &index, gUpdateResponseVersion)
		rw.WriteUint32(responseData, &index, uint32(len(relaysToPing)))

		for i := range relaysToPing {
			rw.WriteUint64(responseData, &index, relaysToPing[i].id)
			rw.WriteString(responseData, &index, relaysToPing[i].address, gMaxRelayAddressLength)
		}

		responseLength := index

		responseData = responseData[:responseLength]

		writer.Header().Set("Content-Type", "application/octet-stream")

		writer.Write(responseData)
	}
}

// CostMatrixHandlerFunc returns the function for the cost matrix endpoint
func CostMatrixHandlerFunc(backend *Backend) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		backend.Mutex.RLock()
		costMatrixData := backend.CostMatrixData
		backend.Mutex.RUnlock()
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(costMatrixData)
	}
}

// RouteMatrixHandlerFunc returns the function for the matrix endpoint
func RouteMatrixHandlerFunc(backend *Backend) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		backend.Mutex.RLock()
		routeMatrixData := backend.RouteMatrixData
		backend.Mutex.RUnlock()
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(routeMatrixData)
	}
}

// NearHandlerFunc returns the function for the near endpoint
func NearHandlerFunc(backend *Backend) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		backend.Mutex.RLock()
		nearData := backend.NearData
		backend.Mutex.RUnlock()
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(nearData)
	}
}
