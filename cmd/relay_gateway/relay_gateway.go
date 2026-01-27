package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
	"context"
	"slices"
	"maps"
	"sort"

	"github.com/redis/go-redis/v9"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	"github.com/networknext/next/modules/encoding"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/packets"
)

var redisHostname string
var redisCluster []string
var pingKey []byte
var relayBackendPublicKey []byte
var relayBackendPrivateKey []byte

var mutex sync.Mutex
var relayBackendAddresses []string

var httpClient *http.Client

func main() {

	service := common.CreateService("relay_gateway")

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisCluster = envvar.GetStringArray("REDIS_CLUSTER", []string{})
	pingKey = envvar.GetBase64("PING_KEY", []byte{})
	relayBackendPublicKey = envvar.GetBase64("RELAY_BACKEND_PUBLIC_KEY", []byte{})
	relayBackendPrivateKey = envvar.GetBase64("RELAY_BACKEND_PRIVATE_KEY", []byte{})

	if len(redisCluster) > 0 {
		core.Debug("redis cluster: %v", redisCluster)
	} else {
		core.Debug("redis hostname: %s", redisHostname)
	}

	if len(pingKey) == 0 {
		core.Error("You must supply PING_KEY")
		os.Exit(1)
	}

	if len(relayBackendPublicKey) == 0 {
		core.Error("You must supply RELAY_BACKEND_PUBLIC_KEY")
		os.Exit(1)
	}

	if len(relayBackendPrivateKey) == 0 {
		core.Error("You must supply RELAY_BACKEND_PRIVATE_KEY")
		os.Exit(1)
	}

	core.Debug("ping key: %x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x",
		pingKey[0],
		pingKey[1],
		pingKey[2],
		pingKey[3],
		pingKey[4],
		pingKey[5],
		pingKey[6],
		pingKey[7],
		pingKey[8],
		pingKey[9],
		pingKey[10],
		pingKey[11],
		pingKey[12],
		pingKey[13],
		pingKey[14],
		pingKey[15],
		pingKey[16],
		pingKey[17],
		pingKey[18],
		pingKey[19],
		pingKey[20],
		pingKey[21],
		pingKey[22],
		pingKey[23],
		pingKey[24],
		pingKey[25],
		pingKey[26],
		pingKey[27],
		pingKey[28],
		pingKey[29],
		pingKey[30],
		pingKey[31],
	)

	httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}

	TrackRelayBackendInstances(service)

	service.UpdateMagic()

	service.LoadDatabase(relayBackendPublicKey, relayBackendPrivateKey)

	service.StartWebServer()

	service.Router.HandleFunc("/relay_update", RelayUpdateHandler(GetRelayData(service), GetMagicValues(service))).Methods("POST")

	service.Router.HandleFunc("/relay_backends", RelayBackendsHandler)

	service.WaitForShutdown()
}

func RelayUpdateHandler(getRelayData func() *common.RelayData, getMagicValues func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte)) func(writer http.ResponseWriter, request *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		startTime := time.Now()

		defer func() {
			duration := time.Since(startTime)
			if duration.Milliseconds() > 1000 {
				core.Warn("long relay update: %s", duration.String())
			}
		}()

		if request.Header.Get("Content-Type") != "application/octet-stream" {
			core.Error("[%s] unsupported content type", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		body, err := io.ReadAll(request.Body)
		if err != nil {
			core.Error("[%s] could not read request body: %v", request.RemoteAddr, err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		defer request.Body.Close()

		// ignore the relay update if it's too small to be valid

		packetBytes := len(body)

		if packetBytes < 1+1+4+2+crypto.Box_MacSize+crypto.Box_NonceSize {
			core.Error("[%s] relay update packet is too small to be valid", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// read the version and decide if we can handle it

		index := 0
		packetData := body
		var packetVersion uint8
		encoding.ReadUint8(packetData, &index, &packetVersion)

		if packetVersion < packets.RelayUpdateRequestPacket_VersionMin || packetVersion > packets.RelayUpdateRequestPacket_VersionMax {
			core.Error("[%s] invalid relay update packet version: %d", request.RemoteAddr, packetVersion)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// read the relay address

		var relayAddress net.UDPAddr
		if !encoding.ReadAddress(packetData, &index, &relayAddress) {
			core.Error("[%s] could not read relay address", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// check if the relay exists via relay id derived from relay address

		relayData := getRelayData()

		relayId := common.RelayId(relayAddress.String())

		relay, ok := relayData.RelayHash[relayId]
		if !ok {
			core.Error("[%s] unknown relay %s [%x]", request.RemoteAddr, relayAddress.String(), relayId)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// decrypt the relay update

		nonce := packetData[packetBytes-crypto.Box_NonceSize:]

		encryptedData := packetData[index : packetBytes-crypto.Box_NonceSize]
		encryptedBytes := len(encryptedData)

		relayPublicKey := relay.PublicKey[:]

		if len(relayPublicKey) == 0 {
			core.Error("[%s] relay public key of length 0", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		err = crypto.Box_Decrypt(relayPublicKey, relayBackendPrivateKey, nonce, encryptedData, encryptedBytes)
		if err != nil {
			core.Error("[%s] failed to decrypt relay update (%d bytes)", request.RemoteAddr, encryptedBytes)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// read the timestamp in the packet

		var packetTimestamp uint64

		timestampIndex := index

		encoding.ReadUint64(packetData, &index, &packetTimestamp)

		currentTimestamp := uint64(startTime.Unix())

		if packetTimestamp < currentTimestamp-10 {
			core.Error("[%s] relay update request is too old", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		if packetTimestamp > currentTimestamp+10 {
			core.Error("[%s] relay update request is in the future", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// relay update accepted

		relayName := relay.Name

		core.Log("[%s] received update for %s [%016x] (%d bytes)", request.RemoteAddr, relayName, relayId, encryptedBytes)

		var responsePacket packets.RelayUpdateResponsePacket

		responsePacket.Version = packets.RelayUpdateResponsePacket_VersionWrite
		responsePacket.Timestamp = uint64(time.Now().Unix())
		responsePacket.TargetVersion = relay.Version

		relayIndex := 0

		for i := range relayData.RelayIds {

			if relayData.RelayIds[i] == relayId {
				continue
			}

			address := relayData.RelayArray[i].PublicAddress

			internal := uint8(0)
			if relay.Seller.Id == relayData.RelaySellerIds[i] &&
				relayData.RelayArray[i].HasInternalAddress && relay.HasInternalAddress &&
				relayData.RelayArray[i].InternalGroup == relay.InternalGroup {
				address = relayData.RelayArray[i].InternalAddress
				internal = 1
			}

			responsePacket.RelayId[relayIndex] = relayData.RelayIds[i]
			responsePacket.RelayAddress[relayIndex] = address
			responsePacket.RelayInternal[relayIndex] = internal

			relayIndex++
		}

		responsePacket.NumRelays = uint32(relayIndex)

		responsePacket.UpcomingMagic, responsePacket.CurrentMagic, responsePacket.PreviousMagic = getMagicValues()

		responsePacket.ExpectedPublicAddress = relay.PublicAddress

		if relay.HasInternalAddress {
			responsePacket.ExpectedHasInternalAddress = 1
			responsePacket.ExpectedInternalAddress = relay.InternalAddress
		}

		copy(responsePacket.ExpectedRelayPublicKey[:], relay.PublicKey)
		copy(responsePacket.ExpectedRelayBackendPublicKey[:], relayBackendPublicKey)

		relaySecretKey, ok := relayData.RelaySecretKeys[relay.Id]
		if !ok {
			core.Error("[%s] could not find relay secret key", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		token := core.RouteToken{}
		token.NextAddress = net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 10000}
		token.PrevAddress = net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 20000}
		core.WriteEncryptedRouteToken(&token, responsePacket.TestToken[:], relaySecretKey)

		copy(responsePacket.PingKey[:], pingKey)

		// send the response packet back to the relay

		responseData := make([]byte, responsePacket.GetMaxSize())

		responseData = responsePacket.Write(responseData)

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))

		writer.Write(responseData)

		// adjust the packet to current time so we can detect when redis is overloaded in the relay backend

		encoding.WriteUint64(packetData, &timestampIndex, currentTimestamp)

		// forward the decrypted relay update to the relay backends

		mutex.Lock()
		addresses := make([]string, len(relayBackendAddresses))
		copy(addresses, relayBackendAddresses)
		mutex.Unlock()

		for i := range addresses {
			core.Debug("forwarding relay update to %s", addresses[i])
			go func(index int) {
				url := fmt.Sprintf("http://%s/relay_update", addresses[index])
				buffer := bytes.NewBuffer(body[:packetBytes-(crypto.Box_MacSize+crypto.Box_NonceSize)])
				forward_request, err := http.NewRequest("POST", url, buffer)
				if err == nil {
					response, err := httpClient.Do(forward_request)
					if err != nil && response != nil {
						io.Copy(io.Discard, response.Body)
						response.Body.Close()
					}
				}
			}(i)
		}
	}
}

func RelayBackendsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	mutex.Lock()
	for i := range relayBackendAddresses {
		fmt.Fprintf(w, "%s\n", relayBackendAddresses[i])		
	}
	mutex.Unlock()
}

func TrackRelayBackendInstances(service *common.Service) {

	var redisClient redis.Cmdable
	if len(redisCluster) > 0 {
		redisClient = common.CreateRedisClusterClient(redisCluster)
	} else {
		redisClient = common.CreateRedisClient(redisHostname)
	}

	go func() {

		ticker := time.NewTicker(10 * time.Second)

		updateRelayBackendInstances(service, redisClient)

		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:
				updateRelayBackendInstances(service, redisClient)
			}
		}
	}()
}

func updateRelayBackendInstances(service *common.Service, redisClient redis.Cmdable) {

	ctx := context.Background()

	core.Debug("updating relay backend instances")

	currentMinutes := time.Now().Unix() / 60
	previousMinutes := currentMinutes - 1

	currentKeys, err := redisClient.HKeys(ctx, fmt.Sprintf("relay-backends-%d", currentMinutes)).Result()
	if err != nil {
		core.Warn("could not get current relay backends from redis: %v", err)
		return
	}

	previousKeys, err := redisClient.HKeys(ctx, fmt.Sprintf("relay-backends-%d", previousMinutes)).Result()
	if err != nil {
		core.Warn("could not get previous relay backends from redis: %v", err)
		return
	}

	addressMap := map[string]int{}

	for i := range currentKeys {
	    addressMap[currentKeys[i]] = 1
	}

	for i := range previousKeys {
		addressMap[previousKeys[i]] = 1
	}

	addresses := slices.Collect(maps.Keys(addressMap))

	sort.Strings(addresses)

	if len(addresses) == 0 {
		core.Warn("(no relay backends)")
	} else {
		core.Debug("relay backends: %v", addresses)
	}

	mutex.Lock()
	relayBackendAddresses = addresses
	mutex.Unlock()
}

func GetRelayData(service *common.Service) func() *common.RelayData {
	return func() *common.RelayData {
		return service.RelayData()
	}
}

func GetMagicValues(service *common.Service) func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte) {
	return func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte) {
		return service.GetMagicValues()
	}
}
