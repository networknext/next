package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/networknext/accelerate/modules/common"
	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/core"
	"github.com/networknext/accelerate/modules/crypto"
	"github.com/networknext/accelerate/modules/encoding"
	"github.com/networknext/accelerate/modules/envvar"
	"github.com/networknext/accelerate/modules/packets"
)

var redisHostname string
var redisPassword string
var redisPubsubChannelName string
var relayUpdateBatchSize int
var relayUpdateBatchDuration time.Duration
var relayUpdateChannelSize int
var pingKey []byte
var relayBackendPublicKey []byte
var relayBackendPrivateKey []byte

var producer *common.RedisPubsubProducer

func main() {

	service := common.CreateService("relay_gateway")

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")
	redisPubsubChannelName = envvar.GetString("REDIS_PUBSUB_CHANNEL_NAME", "relay_update")
	relayUpdateBatchSize = envvar.GetInt("RELAY_UPDATE_BATCH_SIZE", 100)
	relayUpdateBatchDuration = envvar.GetDuration("RELAY_UPDATE_BATCH_DURATION", 1000*time.Millisecond)
	relayUpdateChannelSize = envvar.GetInt("RELAY_UPDATE_CHANNEL_SIZE", 10*1024)
	pingKey = envvar.GetBase64("PING_KEY", []byte{})
	relayBackendPublicKey = envvar.GetBase64("RELAY_BACKEND_PUBLIC_KEY", []byte{})
	relayBackendPrivateKey = envvar.GetBase64("RELAY_BACKEND_PRIVATE_KEY", []byte{})

	core.Debug("redis hostname: %s", redisHostname)
	core.Debug("redis pubsub channel name: %s", redisPubsubChannelName)
	core.Debug("relay update batch size: %d", relayUpdateBatchSize)
	core.Debug("relay update batch duration: %v", relayUpdateBatchDuration)
	core.Debug("relay update channel size: %d", relayUpdateChannelSize)

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

	core.Debug("ping key: %x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x,%x\n",
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

	producer = CreatePubsubProducer(service)

	service.UpdateMagic()

	service.LoadDatabase()

	service.StartWebServer()

	service.Router.HandleFunc("/relay_update", RelayUpdateHandler(GetRelayData(service), GetMagicValues(service))).Methods("POST")

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
			core.Debug("[%s] unsupported content type", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			core.Error("[%s] could not read request body: %v", request.RemoteAddr, err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		defer request.Body.Close()

		// ignore the relay update if it's too small to be valid

		packetBytes := len(body)

		if packetBytes < 1+1+4+2+crypto.Box_MacSize+crypto.Box_NonceSize {
			core.Debug("[%s] relay update packet is too small to be valid", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// read the version and decide if we can handle it

		index := 0
		packetData := body
		var packetVersion uint8
		encoding.ReadUint8(packetData, &index, &packetVersion)

		if packetVersion < packets.RelayUpdateRequestPacket_VersionMin || packetVersion > packets.RelayUpdateRequestPacket_VersionMax {
			core.Debug("[%s] invalid relay update packet version: %d", request.RemoteAddr, packetVersion)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// read the relay address

		var relayAddress net.UDPAddr
		if !encoding.ReadAddress(packetData, &index, &relayAddress) {
			core.Debug("[%s] could not read relay address", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// check if the relay exists via relay id derived from relay address

		relayData := getRelayData()

		relayId := common.RelayId(relayAddress.String())

		relay, ok := relayData.RelayHash[relayId]
		if !ok {
			core.Debug("[%s] unknown relay %x", request.RemoteAddr, relayId)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// decrypt the relay update

		nonce := packetData[packetBytes-crypto.Box_NonceSize:]

		encryptedData := packetData[index : packetBytes-crypto.Box_NonceSize]
		encryptedBytes := len(encryptedData)

		relayPublicKey := relay.PublicKey[:]

		if len(relayPublicKey) == 0 {
			core.Debug("[%s] relay public key of length 0", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		err = crypto.Box_Decrypt(relayPublicKey, relayBackendPrivateKey, nonce, encryptedData, encryptedBytes)
		if err != nil {
			core.Debug("[%s] failed to decrypt relay update", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// read the timestamp in the packet

		var packetTimestamp uint64

		encoding.ReadUint64(packetData, &index, &packetTimestamp)

		currentTimestamp := uint64(startTime.Unix())

		if packetTimestamp < currentTimestamp-10 {
			core.Debug("[%s] relay update request is too old", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		if packetTimestamp > currentTimestamp+10 {
			core.Debug("[%s] relay update request is in the future", request.RemoteAddr)
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// relay update accepted

		relayName := relay.Name

		core.Debug("[%s] received update for %s [%x]", request.RemoteAddr, relayName, relayId)

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

		token := core.RouteToken{}
		core.WriteEncryptedRouteToken(&token, responsePacket.TestToken[:], relayBackendPrivateKey, relay.PublicKey)

		copy(responsePacket.PingKey[:], pingKey)

		// send the response packet back to the relay

		responseData := make([]byte, responsePacket.GetMaxSize())

		responseData = responsePacket.Write(responseData)

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))

		writer.Write(responseData)

		// forward the relay update to the relay backend, sans crypto stuff (it's now decrypted...)

		messageData := body[:packetBytes-(crypto.Box_MacSize+crypto.Box_NonceSize)]

		producer.MessageChannel <- messageData
	}
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

func CreatePubsubProducer(service *common.Service) *common.RedisPubsubProducer {

	config := common.RedisPubsubConfig{}

	config.RedisHostname = redisHostname
	config.RedisPassword = redisPassword
	config.PubsubChannelName = redisPubsubChannelName
	config.BatchSize = relayUpdateBatchSize
	config.BatchDuration = relayUpdateBatchDuration
	config.MessageChannelSize = relayUpdateChannelSize

	var err error
	producer, err = common.CreateRedisPubsubProducer(service.Context, config)
	if err != nil {
		core.Error("could not create redis pubsub producer")
		os.Exit(1)
	}

	return producer
}
