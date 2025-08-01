package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/packets"
)

var service *common.Service
var numRelays int
var numSessions int
var relayAddress string
var clientAddress string
var serverBackendAddress net.UDPAddr
var buyerId uint64
var buyerPrivateKey []byte
var basePort int

func main() {

	service = common.CreateService("load_test_sessions")

	numRelays = envvar.GetInt("NUM_RELAYS", 1000)

	numSessions = envvar.GetInt("NUM_SESSIONS", 1000)

	clientAddress = envvar.GetString("CLIENT_ADDRESS", "127.0.0.1")

	relayAddress = envvar.GetString("RELAY_ADDRESS", "127.0.0.1")

	serverBackendAddress = envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:40000"))

	basePort = envvar.GetInt("BASE_PORT", 10000)

	core.Log("num relays = %d", numRelays)
	core.Log("num sessions = %d", numSessions)
	core.Log("base port = %d", basePort)
	core.Log("relay address = %s", relayAddress)
	core.Log("client address = %s", clientAddress)
	core.Log("server backend address = %s", serverBackendAddress.String())

	buyerPrivateKey = envvar.GetBase64("NEXT_BUYER_PRIVATE_KEY", nil)

	if buyerPrivateKey == nil {
		panic("you must supply the buyer private key")
	}

	clientAddress = DetectGoogleClientAddress(clientAddress)

	buyerId = binary.LittleEndian.Uint64(buyerPrivateKey[0:8])

	buyerPrivateKey = buyerPrivateKey[8:]

	core.Log("simulating %d sessions", numSessions)

	go SimulateSessions()

	service.WaitForShutdown()
}

func SimulateSessions() {
	for i := 0; i < numSessions; i++ {
		go RunSession(i)
	}
}

func RunSession(index int) {

	var r = rand.New(rand.NewSource(time.Now().UnixNano()))

	time.Sleep(time.Duration(r.Intn(1000)) * time.Millisecond) // jitter delay

	time.Sleep(time.Duration(r.Intn(360)) * time.Second) // initial delay

	address := core.ParseAddress(fmt.Sprintf("%s:%d", clientAddress, basePort+index))

	userHash := r.Uint64()

	go func() {

		for {

			// session loop

			sessionId := r.Uint64()

			core.Debug("[%016x] new (%d)", sessionId, index)

			sliceNumber := 0

			retryNumber := 0

			datacenterId := common.DatacenterId(fmt.Sprintf("test.%03d", (index%numRelays)/10*10))

			bindAddress := fmt.Sprintf("0.0.0.0:%d", basePort+index)

			serverAddress := core.ParseAddress("127.0.0.1:50000")

			clientPublicKey, _ := crypto.Box_KeyPair()
			serverPublicKey, _ := crypto.Box_KeyPair()

			sessionDuration := 0

			var mutex sync.Mutex

			var receivedResponse bool

			var next, fallbackToDirect, clientPingTimedOut bool

			var sessionDataBytes int32
			var sessionData [packets.SDK_MaxSessionDataSize]byte
			var sessionDataSignature [packets.SDK_SignatureBytes]byte

			numClientRelays := int32(packets.SDK_MaxClientRelays)
			var clientRelayIds [packets.SDK_MaxClientRelays]uint64
			for i := range clientRelayIds {
				clientRelayIds[i] = common.RelayId(fmt.Sprintf("%s:%d", relayAddress, 10000+common.RandomInt(0, numRelays)))
			}

			numServerRelays := int32(packets.SDK_MaxServerRelays)
			var serverRelayIds [packets.SDK_MaxServerRelays]uint64
			for i := range serverRelayIds {
				serverRelayIds[i] = common.RelayId(fmt.Sprintf("%s:%d", relayAddress, 10000+common.RandomInt(0, numRelays)))
			}

			lc := net.ListenConfig{}
			lp, err := lc.ListenPacket(context.Background(), "udp", bindAddress)
			if err != nil {
				panic(fmt.Sprintf("could not bind socket: %v", err))
			}

			conn := lp.(*net.UDPConn)

			go func() {

				for {

					buffer := make([]byte, constants.MaxPacketBytes)
					packetBytes, from, err := conn.ReadFromUDP(buffer[:])
					if err != nil {
						break
					}

					if packetBytes < 1 {
						continue
					}

					if from.String() != serverBackendAddress.String() {
						continue
					}

					packetData := buffer[:packetBytes]

					packetType := packetData[0]

					if packetType == packets.SDK_SESSION_UPDATE_RESPONSE_PACKET {

						packetData = packetData[18:]

						packet := packets.SDK_SessionUpdateResponsePacket{}
						if err := packets.ReadPacket(packetData, &packet); err != nil {
							core.Error("failed to read packet: %v", err)
							continue
						}

						mutex.Lock()

						if packet.SessionId == sessionId && packet.SliceNumber == uint32(sliceNumber) {

							sessionDataBytes = packet.SessionDataBytes
							copy(sessionData[:], packet.SessionData[:])
							copy(sessionDataSignature[:], packet.SessionDataSignature[:])

							next = packet.RouteType != packets.SDK_RouteTypeDirect

							receivedResponse = true

						}

						mutex.Unlock()

					}

				}

				conn.Close()

			}()

			ticker := time.NewTicker(10 * time.Second)

			for {
				select {

				case <-service.Context.Done():
					return

				case <-ticker.C:

					mutex.Lock()
					receivedResponse = false
					mutex.Unlock()

					for i := 0; i < 9; i++ {

						if retryNumber == 0 {
							core.Debug("[%016x] update (%d)", sessionId, index)
						} else {
							core.Log("[%016x] retry #%d (%d)", sessionId, retryNumber, index)
						}

						mutex.Lock()

						packet := packets.SDK_SessionUpdateRequestPacket{
							Version:            packets.SDKVersion{255, 255, 255},
							BuyerId:            buyerId,
							DatacenterId:       datacenterId,
							SessionId:          sessionId,
							SliceNumber:        uint32(sliceNumber),
							RetryNumber:        int32(retryNumber),
							ClientAddress:      address,
							ServerAddress:      serverAddress,
							UserHash:           userHash,
							Next:               next,
							FallbackToDirect:   fallbackToDirect,
							ClientPingTimedOut: clientPingTimedOut,
						}

						copy(packet.ClientRoutePublicKey[:], clientPublicKey[:])
						copy(packet.ServerRoutePublicKey[:], serverPublicKey[:])

						if sliceNumber != 0 {
							packet.SessionDataBytes = sessionDataBytes
							copy(packet.SessionData[:], sessionData[:])
							copy(packet.SessionDataSignature[:], sessionDataSignature[:])
						}

						if sliceNumber >= 1 {
							packet.ClientRelayPingsHaveChanged = sliceNumber == 1 || (sliceNumber%30) == 0
							packet.HasClientRelayPings = (sessionId % 10) == 0
							packet.NumClientRelays = numClientRelays
							copy(packet.ClientRelayIds[:], clientRelayIds[:])
							for i := range packet.ClientRelayRTT {
								packet.ClientRelayRTT[i] = 1 + int32((sessionId^clientRelayIds[i])%30)
							}
						}

						if sliceNumber >= 1 {
							packet.ServerRelayPingsHaveChanged = sliceNumber == 1 || (sliceNumber%30) == 0
							packet.HasServerRelayPings = (sessionId % 10) == 0
							packet.NumServerRelays = numServerRelays
							copy(packet.ServerRelayIds[:], clientRelayIds[:])
							for i := range packet.ServerRelayRTT {
								packet.ServerRelayRTT[i] = 1
							}
						}

						packet.DirectRTT = float32(sessionId%250) + 10

						if next {
							packet.NextRTT = 1
						}

						mutex.Unlock()

						packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_SESSION_UPDATE_REQUEST_PACKET, constants.MaxPacketBytes, &address, &serverBackendAddress, buyerPrivateKey)
						if err != nil {
							core.Error("failed to write session update request packet: %v", err)
							return
						}

						if _, err := conn.WriteToUDP(packetData, &serverBackendAddress); err != nil {
							core.Error("failed to send packet: %v", err)
							return
						}

						time.Sleep(1 * time.Second)

						mutex.Lock()
						done := receivedResponse
						mutex.Unlock()
						if done || fallbackToDirect {
							break
						}

						mutex.Lock()
						retryNumber += 1
						mutex.Unlock()
					}

					mutex.Lock()

					if !receivedResponse && !fallbackToDirect {
						core.Error("[%016x] fallback to direct (%d)", sessionId, index)
						fallbackToDirect = true
					}

					// send client relay request packets once every 5 minutes

					if sliceNumber == 1 || sliceNumber > 0 && (sliceNumber%30) == 0 {

						packet := packets.SDK_ClientRelayRequestPacket{
							Version:      packets.SDKVersion{255, 255, 255},
							BuyerId:      buyerId,
							RequestId:    rand.Uint64(),
							DatacenterId: datacenterId,
						}

						packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_CLIENT_RELAY_REQUEST_PACKET, constants.MaxPacketBytes, &address, &serverBackendAddress, buyerPrivateKey)
						if err != nil {
							core.Error("failed to write client relay request packet: %v", err)
							return
						}

						if _, err := conn.WriteToUDP(packetData, &serverBackendAddress); err != nil {
							core.Error("failed to send packet: %v", err)
							return
						}
					}

					// send server relay request packets once every 5 minutes

					if sliceNumber == 1 || sliceNumber > 0 && (sliceNumber%30) == 0 {

						packet := packets.SDK_ServerRelayRequestPacket{
							Version:      packets.SDKVersion{255, 255, 255},
							BuyerId:      buyerId,
							RequestId:    rand.Uint64(),
							DatacenterId: datacenterId,
						}

						packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_SERVER_RELAY_REQUEST_PACKET, constants.MaxPacketBytes, &address, &serverBackendAddress, buyerPrivateKey)
						if err != nil {
							core.Error("failed to write server relay request packet: %v", err)
							return
						}

						if _, err := conn.WriteToUDP(packetData, &serverBackendAddress); err != nil {
							core.Error("failed to send packet: %v", err)
							return
						}
					}

					// next slice

					sliceNumber += 1
					retryNumber = 0
					receivedResponse = false

					mutex.Unlock()

					sessionDuration += 10
				}
			}
		}

	}()

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

	cmd.Stderr = os.Stderr

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

func DetectGoogleClientAddress(input string) string {
	result, output := Bash("curl -s http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip -H \"Metadata-Flavor: Google\" --max-time 10 -s 2>/dev/null")
	if !result {
		return input
	}
	output = strings.TrimSuffix(output, "\n")
	core.Log("google cloud client address is '%s'", output)
	return output
}
