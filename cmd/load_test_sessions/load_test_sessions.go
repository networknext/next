package main

import (
	"time"
	"fmt"
	"net"
	"encoding/binary"
	"context"
	"math/rand"
	"sync"
	"os"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/packets"
	"github.com/networknext/next/modules/crypto"
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

	customerPrivateKey := envvar.GetBase64("NEXT_CUSTOMER_PRIVATE_KEY", nil)

	if customerPrivateKey == nil {
		panic("you must supply the customer private key")
	}

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

	time.Sleep(time.Duration(common.RandomInt(0, 10000)) * time.Millisecond)

	address := core.ParseAddress(fmt.Sprintf("%s:%d", clientAddress, 10000+index))

	userHash := rand.Uint64()

	sessionId := rand.Uint64()

	sliceNumber := 0

	retryNumber := 0

	datacenterId := common.DatacenterId(fmt.Sprintf("test.%03d", (index%numRelays)/10*10))

	bindAddress := fmt.Sprintf("0.0.0.0:%d", 10000+index)

	serverAddress := common.RandomAddress()

	clientPublicKey, _ := crypto.Box_KeyPair()
	serverPublicKey, _ := crypto.Box_KeyPair()

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
				fmt.Printf("udp receive error: %v\n", err)
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

				packetData = packetData[16:len(packetData)-2]

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

	ticker := time.NewTicker(10*time.Second)

	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				for i := 0; i < 5; i++ {

					if retryNumber == 0 {
						fmt.Printf("update session %03d\n", index)
					} else {
						fmt.Printf("update session %03d (retry %d)\n", index, retryNumber)
					}

					mutex.Lock()

					receivedResponse = false

					packet := packets.SDK_SessionUpdateRequestPacket{
						Version: packets.SDKVersion{255,255,255},
						BuyerId: buyerId,
						DatacenterId: datacenterId,
						SessionId: sessionId,
						SliceNumber: uint32(sliceNumber),
						RetryNumber: int32(retryNumber),
						ClientAddress: address,
						ServerAddress: serverAddress,
						UserHash: userHash,
						Next: next,
						FallbackToDirect: fallbackToDirect,
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
							packet.NearRelayRTT[i] = int32((sessionId^nearRelayIds[i])%10)
						}
					}

					if sliceNumber >= 1 {
						// give one-in-ten sessions a very high direct RTT, so they tend to go over network next
						if (sessionId % 10) == 0 {
							packet.DirectRTT = 250
						} else {
							packet.DirectRTT = 1
						}
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
					core.Error("did not receive response")
					fallbackToDirect = true
					os.Exit(1)
				}

				sliceNumber += 1
				retryNumber = 0
				receivedResponse = false

				mutex.Unlock()
			}
		}
	}()

}
