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
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/packets"
)

var service *common.Service
var numRelays int
var numSessions int
var clientAddress string
var serverBackendAddress net.UDPAddr
var buyerId uint64
var buyerPrivateKey []byte

func main() {

	service = common.CreateService("load_test_sessions")

	numRelays = envvar.GetInt("NUM_RELAYS", 1000)

	numSessions = envvar.GetInt("NUM_SESSIONS", 1000)

	clientAddress = envvar.GetString("CLIENT_ADDRESS", "127.0.0.1")

	serverBackendAddress = envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:40000"))

	core.Log("num relays = %d", numRelays)
	core.Log("num sessions = %d", numSessions)
	core.Log("client address = %s", clientAddress)
	core.Log("server backend address = %s", serverBackendAddress.String())

	customerPrivateKey := envvar.GetBase64("NEXT_CUSTOMER_PRIVATE_KEY", nil)

	if customerPrivateKey == nil {
		panic("you must supply the customer private key")
	}

	clientAddress = DetectGoogleClientAddress(clientAddress)

	buyerId = binary.LittleEndian.Uint64(customerPrivateKey[0:8])

	buyerPrivateKey = customerPrivateKey[8:]

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

	initialDelay := time.Duration(r.Intn(300)) * time.Second

	time.Sleep(initialDelay)

	address := core.ParseAddress(fmt.Sprintf("%s:%d", clientAddress, 10000+index))

	userHash := r.Uint64()

	sessionId := r.Uint64()

	sliceNumber := 0

	retryNumber := 0

	datacenterId := common.DatacenterId(fmt.Sprintf("test.%03d", (index%numRelays)/10*10))

	bindAddress := fmt.Sprintf("0.0.0.0:%d", 10000+index)

	serverAddress := core.ParseAddress("127.0.0.1:50000")

	clientPublicKey, _ := crypto.Box_KeyPair()
	serverPublicKey, _ := crypto.Box_KeyPair()

	sessionDuration := 0

	sessionTimeout := 300

	var mutex sync.Mutex

	var receivedResponse bool

	var next, fallbackToDirect, clientPingTimedOut bool

	var sessionDataBytes int32
	var sessionData [packets.SDK_MaxSessionDataSize]byte
	var sessionDataSignature [packets.SDK_SignatureBytes]byte

	var numNearRelays int32
	var nearRelayIds [packets.SDK_MaxNearRelays]uint64

	lc := net.ListenConfig{}
	lp, err := lc.ListenPacket(context.Background(), "udp", bindAddress)
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	go func() {

		for {

			buffer := make([]byte, 4096)
			packetBytes, from, err := conn.ReadFromUDP(buffer[:])
			if err != nil {
				core.Error("udp receive error: %v", err)
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

				packetData = packetData[16 : len(packetData)-2]

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

					if packet.HasNearRelays {
						numNearRelays = packet.NumNearRelays
						copy(nearRelayIds[:], packet.NearRelayIds[:])
					}

					receivedResponse = true

				}

				mutex.Unlock()

			}

		}

		conn.Close()

	}()

	ticker := time.NewTicker(10 * time.Second)

	go func() {
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
						core.Debug("update session %03d", index)
					} else {
						core.Debug("update session %03d (retry %d)", index, retryNumber)
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
						packet.HasNearRelayPings = true
						packet.NumNearRelays = numNearRelays
						copy(packet.NearRelayIds[:], nearRelayIds[:])
						for i := range packet.NearRelayRTT {
							packet.NearRelayRTT[i] = int32((sessionId ^ nearRelayIds[i]) % 10)
						}
					}

					if sliceNumber >= 1 {
						packet.DirectRTT = float32(sessionId % 500)
					}

					if next {
						packet.NextRTT = 1
					}

					mutex.Unlock()

					packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_SESSION_UPDATE_REQUEST_PACKET, 4096, &address, &serverBackendAddress, buyerPrivateKey)
					if err != nil {
						core.Error("failed to write response packet: %v", err)
						return
					}

					if _, err := conn.WriteToUDP(packetData, &serverBackendAddress); err != nil {
						core.Error("failed to send packet: %v", err)
						return
					}

					time.Sleep(time.Second)

					mutex.Lock()
					done := receivedResponse
					mutex.Unlock()
					if done {
						break
					}

					mutex.Lock()
					retryNumber += 1
					mutex.Unlock()
				}

				mutex.Lock()

				if !receivedResponse {
					core.Error("fallback to direct")
					fallbackToDirect = true
					os.Exit(1)
				}

				sliceNumber += 1
				retryNumber = 0
				receivedResponse = false

				mutex.Unlock()

				sessionDuration += 10

				if sessionDuration > sessionTimeout {
					if !clientPingTimedOut {
						core.Debug("client ping timed out")
						clientPingTimedOut = true
					}
				}

				if sessionDuration > sessionTimeout+60 {
					mutex.Lock()
					core.Debug("new session %03d\n", index)
					sessionId = r.Uint64()
					sliceNumber = 0
					retryNumber = 0
					sessionDuration = 0
					next = false
					fallbackToDirect = false
					clientPingTimedOut = false
					sessionDataBytes = 0
					numNearRelays = 0
					mutex.Unlock()
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
	result, output := Bash("curl -s http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip -H \"Metadata-Flavor: Google\" --max-time 10 -vs 2>/dev/null")
	if !result {
		return input
	}
	output = strings.TrimSuffix(output, "\n")
	core.Log("google cloud client address is '%s'", output)
	return output
}
