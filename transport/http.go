package transport

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	// Relay entry
)

const gInitRequestMagic = uint32(0x9083708f)
const gInitRequestVersion = 0
const gInitResponseVersion = 0
const gUpdateRequestVersion = 0
const gUpdateResponseVersion = 0
const gMaxRelayIdLength = 256
const gMaxRelayAddressLength = 256
const gRelayTokenBytes = 32
const gMaxRelays = 1024

var gRelayPublicKey = []byte{
	0xf5, 0x22, 0xad, 0xc1, 0xee, 0x04, 0x6a, 0xbe,
	0x7d, 0x89, 0x0c, 0x81, 0x3a, 0x08, 0x31, 0xba,
	0xdc, 0xdd, 0xb5, 0x52, 0xcb, 0x73, 0x56, 0x10,
	0xda, 0xa9, 0xc0, 0xae, 0x08, 0xa2, 0xcf, 0x5e,
}

func RelayInitHandlerFunc(backend interface{}) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			// Error handling besides return?
			return
		}

		index := 0

		var magic uint32
		if !crypto.ReadUint32(body, &index, &magic) || magic != gInitRequestMagic {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		var version uint32
		if !crypto.ReadUint32(body, &index, &version) || version != gInitRequestVersion {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		var nonce []byte
		if !crypto.ReadBytes(body, &index, &nonce, C.crypto_box_NONCEBYTES) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		var relay_address string
		if !crypto.ReadString(body, &index, &relay_address, gMaxRelayAddressLength) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		var encrypted_token []byte
		if !crypto.ReadBytes(body, &index, &encrypted_token, gRelayTokenBytes+C.crypto_box_MACBYTES) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		if !crypto.CryptoCheck(encrypted_token, nonce, gRelayPublicKey[:], core.RouterPrivateKey[:]) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		key := relay_address

		backend.mutex.RLock()
		_, relayAlreadyExists := backend.relayDatabase[key]
		backend.mutex.RUnlock()

		if relayAlreadyExists {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		relayEntry := RelayEntry{}
		relayEntry.name = relay_address
		relayEntry.id = crypto.GetRelayId(relay_address)
		relayEntry.address = core.ParseAddress(relay_address)
		relayEntry.lastUpdate = time.Now().Unix()
		relayEntry.token = core.RandomBytes(gRelayTokenBytes)

		backend.mutex.Lock()
		backend.relayDatabase[key] = relayEntry
		backend.dirty = true
		backend.mutex.Unlock()

		writer.Header().Set("Content-Type", "application/octet-stream")

		index = 0
		responseData := make([]byte, 64)
		crypto.WriteUint32(responseData, &index, gInitResponseVersion)
		crypto.WriteUint64(responseData, &index, uint64(time.Now().Unix()))
		crypto.WriteBytes(responseData, &index, relayEntry.token, gRelayTokenBytes)

		responseData = responseData[:index]
		writer.Write(responseData)
	}
}

func RelayUpdateHandlerFunc(backend interface{}) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			return
		}

		index := 0

		var version uint32
		if !crypto.ReadUint32(body, &index, &version) || version != gUpdateRequestVersion {
			return
		}

		var relay_address string
		if !crypto.ReadString(body, &index, &relay_address, gMaxRelayAddressLength) {
			return
		}

		var token []byte
		if !crypto.ReadBytes(body, &index, &token, gRelayTokenBytes) {
			return
		}

		key := relay_address

		backend.mutex.RLock()
		relayEntry, ok := backend.relayDatabase[key]
		found := false
		if ok && CompareTokens(token, relayEntry.token) {
			found = true
		}
		backend.mutex.RUnlock()

		if !found {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		var num_relays uint32
		if !crypto.ReadUint32(body, &index, &num_relays) {
			return
		}

		if num_relays > gMaxRelays {
			return
		}

		statsUpdate := &core.RelayStatsUpdate{}
		statsUpdate.Id = core.RelayId(relayEntry.id)

		for i := 0; i < int(num_relays); i++ {
			var id uint64
			var rtt, jitter, packet_loss float32
			if !crypto.ReadUint64(body, &index, &id) {
				return
			}
			if !crypto.ReadFloat32(body, &index, &rtt) {
				return
			}
			if !crypto.ReadFloat32(body, &index, &jitter) {
				return
			}
			if !crypto.ReadFloat32(body, &index, &packet_loss) {
				return
			}

			ping := core.RelayStatsPing{}
			ping.RelayId = core.RelayId(id)
			ping.RTT = rtt
			ping.Jitter = jitter
			ping.PacketLoss = packet_loss
			statsUpdate.PingStats = append(statsUpdate.PingStats, ping)
		}

		backend.mutex.Lock()
		backend.statsDatabase.ProcessStats(statsUpdate)
		backend.mutex.Unlock()

		relayEntry = RelayEntry{}
		relayEntry.name = relay_address
		relayEntry.id = crypto.GetRelayId(relay_address)
		relayEntry.address = core.ParseAddress(relay_address)
		relayEntry.lastUpdate = time.Now().Unix()
		relayEntry.token = token

		type RelayPingData struct {
			id      uint64
			address string
		}

		relaysToPing := make([]RelayPingData, 0)

		backend.mutex.Lock()
		backend.relayDatabase[key] = relayEntry
		for k, v := range backend.relayDatabase {
			if k != relay_address {
				if k != relay_address {
					relaysToPing = append(relaysToPing, RelayPingData{id: v.id, address: k})
				}
			}
		}
		backend.mutex.Unlock()

		responseData := make([]byte, 10*1024)

		index = 0

		crypto.WriteUint32(responseData, &index, gUpdateResponseVersion)
		crypto.WriteUint32(responseData, &index, uint32(len(relaysToPing)))

		for i := range relaysToPing {
			crypto.WriteUint64(responseData, &index, relaysToPing[i].id)
			crypto.WriteString(responseData, &index, relaysToPing[i].address, gMaxRelayAddressLength)
		}

		responseLength := index

		responseData = responseData[:responseLength]

		writer.Header().Set("Content-Type", "application/octet-stream")

		writer.Write(responseData)
	}
}

func CostMatrixHandlerFunc(backend interface{}) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		backend.mutex.RLock()
		costMatrixData := backend.costMatrixData
		backend.mutex.RUnlock()
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(costMatrixData)
	}
}

func RouteMatrixHandlerFunc(backend interface{}) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		backend.mutex.RLock()
		routeMatrixData := backend.routeMatrixData
		backend.mutex.RUnlock()
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(routeMatrixData)
	}
}

func NearHandlerFunc(backend interface{}) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		backend.mutex.RLock()
		nearData := backend.nearData
		backend.mutex.RUnlock()
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(nearData)
	}
}
