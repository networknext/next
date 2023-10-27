package main

import (
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"
	"bytes"
	"fmt"
	"sync"
	"bufio"
	"os/exec"
	"strings"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	"github.com/networknext/next/modules/encoding"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/packets"
)

var redisHostname string
var redisPubsubChannelName string
var relayUpdateBatchSize int
var relayUpdateBatchDuration time.Duration
var relayUpdateChannelSize int
var pingKey []byte
var relayBackendPublicKey []byte
var relayBackendPrivateKey []byte
var relayBackendAddress string

var mutex sync.Mutex
var relayBackendAddresses []string

func main() {

	service := common.CreateService("relay_gateway")

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPubsubChannelName = envvar.GetString("REDIS_PUBSUB_CHANNEL_NAME", "relay_update")
	relayUpdateBatchSize = envvar.GetInt("RELAY_UPDATE_BATCH_SIZE", 1)
	relayUpdateBatchDuration = envvar.GetDuration("RELAY_UPDATE_BATCH_DURATION", 1000*time.Millisecond)
	relayUpdateChannelSize = envvar.GetInt("RELAY_UPDATE_CHANNEL_SIZE", 1024*1024)
	pingKey = envvar.GetBase64("PING_KEY", []byte{})
	relayBackendPublicKey = envvar.GetBase64("RELAY_BACKEND_PUBLIC_KEY", []byte{})
	relayBackendPrivateKey = envvar.GetBase64("RELAY_BACKEND_PRIVATE_KEY", []byte{})
	relayBackendAddress = envvar.GetString("RELAY_BACKEND_ADDRESS", "127.0.0.1:30001")

	core.Debug("redis hostname: %s", redisHostname)
	core.Debug("redis pubsub channel name: %s", redisPubsubChannelName)
	core.Debug("relay update batch size: %d", relayUpdateBatchSize)
	core.Debug("relay update batch duration: %v", relayUpdateBatchDuration)
	core.Debug("relay update channel size: %d", relayUpdateChannelSize)
	core.Debug("relay backend address: %s", relayBackendAddress)

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

	TrackRelayBackendInstances(service)

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
			core.Error("[%s] unsupported content type", request.RemoteAddr)
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
			core.Error("[%s] unknown relay %x", request.RemoteAddr, relayId)
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
			core.Error("[%s] failed to decrypt relay update", request.RemoteAddr)
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

		// todo: disable spam for now
		// relayName := relay.Name

//		core.Debug("[%s] received update for %s [%016x]", request.RemoteAddr, relayName, relayId)

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
		token.NextAddress = net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 10000}
		token.PrevAddress = net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 20000}
		core.WriteEncryptedRouteToken(&token, responsePacket.TestToken[:], relayBackendPrivateKey, relay.PublicKey)

		copy(responsePacket.PingKey[:], pingKey)

		// send the response packet back to the relay

		responseData := make([]byte, responsePacket.GetMaxSize())

		responseData = responsePacket.Write(responseData)

		writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))

		writer.Write(responseData)

		// adjust the packet to current time so we can detect when redis is overloaded in the relay backend

		encoding.WriteUint64(packetData, &timestampIndex, currentTimestamp)

		// forward the decrypted relay update to the relay backends

		buffer := bytes.NewBuffer(body[:packetBytes-(crypto.Box_MacSize+crypto.Box_NonceSize)])

		addresses := []string{}
		if relayBackendAddress != "" {
			addresses = []string{relayBackendAddress}
		} else {
			mutex.Lock()
			addresses = relayBackendAddresses
			mutex.Unlock()
		}

		for i := range addresses {
			go func() {
				url := fmt.Sprintf("http://%s/relay_update", addresses[i])
				forward_request, err := http.NewRequest("POST", url, buffer)
				if err == nil {
					httpClient := http.Client{
					    Timeout: time.Second,
					}
					response, _ := httpClient.Do(forward_request)
					if response != nil {
						_,_  = ioutil.ReadAll(response.Body)
						response.Body.Close()
					}
				}
			}()
		}
	}
}

func RunCommand(command string, args []string) (bool, string) {

	cmd := exec.Command(command, args...)

	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return false, ""
	}

	var wait sync.WaitGroup
	var mutex sync.Mutex

	output := ""

	stdoutScanner := bufio.NewScanner(stdoutReader)
	wait.Add(1)
	go func() {
		for stdoutScanner.Scan() {
			mutex.Lock()
			output += stdoutScanner.Text() + "\n"
			mutex.Unlock()
		}
		wait.Done()
	}()

	err = cmd.Start()
	if err != nil {
		return false, output
	}

	wait.Wait()

	err = cmd.Wait()
	if err != nil {
		return false, output
	}

	return true, output
}

func Bash(command string) (bool, string) {
	return RunCommand("bash", []string{"-c", command})
}

func TrackRelayBackendInstances(service *common.Service) {

	// grab google cloud instance name from metadata

	result, instanceName := Bash("curl -s http://metadata/computeMetadata/v1/instance/hostname -H \"Metadata-Flavor: Google\" --max-time 1 -vs 2>/dev/null")
	if !result {
		return // not in google cloud
	}

	instanceName = strings.TrimSuffix(instanceName, "\n")

	tokens := strings.Split(instanceName, ".")

	instanceName = tokens[0]

	core.Log("google cloud instance name is '%s'", instanceName)

	// grab google cloud zone from metadata

	var zone string
	result, zone = Bash("curl -s http://metadata/computeMetadata/v1/instance/zone -H \"Metadata-Flavor: Google\" --max-time 1 -vs 2>/dev/null")
	if !result {
		return // not in google cloud
	}

	zone = strings.TrimSuffix(zone, "\n")

	tokens = strings.Split(zone, "/")

	zone = tokens[len(tokens)-1]

	core.Log("google cloud zone is '%s'", zone)

	// turn zone into region

	tokens = strings.Split(zone, "-")

	region := strings.Join(tokens[:len(tokens)-1], "-")

	core.Log("google cloud region is '%s'", region)

	go func() {

		ticker := time.NewTicker(1000 * time.Millisecond)

		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				_, list_output := Bash(fmt.Sprintf("gcloud compute instance-groups managed list-instances relay-backend --region %s", region))

				list_lines := strings.Split(list_output, "\n")

				instanceIds := make([]string, 0)
				zones := make([]string, 0)
				for i := range list_lines {
					if strings.Contains(list_lines[i], "relay-backend-") {
						values := strings.Fields(list_lines[i])
						instanceId := values[0]
						zone := values[1]
						instanceIds = append(instanceIds, instanceId)
						zones = append(zones, zone)
					}
				}

				addresses := make([]string, len(instanceIds))
				for i := range instanceIds {
					_, describe_output := Bash(fmt.Sprintf("gcloud compute instances describe %s --zone %s", instanceIds[i], zones[i]))
					describe_lines := strings.Split(describe_output, "\n")
					address := ""
					for i := range describe_lines {
						if strings.Contains(describe_lines[i], "networkIP: ") {
							values := strings.Fields(describe_lines[i])
							address = values[1]
						}
					}
					addresses[i] = address
				}

				fmt.Printf("==========================================\n")
				for i := range instanceIds {
					fmt.Printf("%s -> '%s'\n", instanceIds[i], addresses[i])
				}
				fmt.Printf("==========================================\n")

				ok := make([]bool, len(addresses))
				waitGroup := sync.WaitGroup{}
				waitGroup.Add(len(addresses))
				for i := range addresses {
					go func() {
						ok[i], _ = Bash("curl http://%s/health_fanout --max-time 1", addresses[i])
						waitGroup.Done()
					}()
				}
				waitGroup.Wait()

				verified := []string{}
				for i := range addresses {
					if ok[i] {
						verified = append(verified, addresses[i])
					}
				}

				fmt.Printf("==========================================\n")
				for i := range verified {
					fmt.Printf("%s\n", verified[i])
				}
				fmt.Printf("==========================================\n")

				mutex.Lock()
				relayBackendAddresses = verified
				mutex.Unlock()
			}
		}
	}()
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
