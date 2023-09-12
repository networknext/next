package main

import (
	"time"
	"fmt"
	"net"
	"encoding/binary"
	"context"
	"math/rand"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
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

	numSessions = envvar.GetInt("NUM_SESSIONS", 10000)

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

	datacenterId := common.DatacenterId(fmt.Sprintf("test.%03", index%numRelays))

	bindAddress := fmt.Sprintf("0.0.0.0:%d", 10000+index)

	serverAddress := common.RandomAddress()

	lc := net.ListenConfig{}
	lp, err := lc.ListenPacket(context.Background(), "udp", bindAddress)
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	go func() {

		for {

			buffer := make([]byte, 4096)
			_, _, err := conn.ReadFromUDP(buffer[:])
			if err != nil {
				fmt.Printf("udp receive error: %v\n", err)
				break
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

				fmt.Printf("update session %03d\n", index)

				// todo
				_ = address

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
					/*
					ClientRoutePublicKey            [crypto.Box_PublicKeySize]byte
					ServerRoutePublicKey            [crypto.Box_PublicKeySize]byte
					*/

					/*
					SessionDataBytes                int32
					SessionData                     [SDK_MaxSessionDataSize]byte
					SessionDataSignature            [SDK_SignatureBytes]byte
					*/

					/*
					Next                            bool
					FallbackToDirect                bool
					ClientPingTimedOut              bool

					DirectRTT                       float32
					DirectJitter                    float32
					DirectPacketLoss                float32
					DirectMaxPacketLossSeen         float32

					NextRTT                         float32
					NextJitter                      float32
					NextPacketLoss                  float32
					NumNearRelays                   int32

					HasNearRelayPings               bool
					NearRelayIds                    [SDK_MaxNearRelays]uint64
					NearRelayRTT                    [SDK_MaxNearRelays]int32
					NearRelayJitter                 [SDK_MaxNearRelays]int32
					NearRelayPacketLoss             [SDK_MaxNearRelays]float32
					*/
				}

				packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_SERVER_UPDATE_REQUEST_PACKET, 4096, &address, &serverBackendAddress, buyerPrivateKey)
				if err != nil {
					core.Error("failed to write response packet: %v", err)
					return
				}

				if _, err := conn.WriteToUDP(packetData, &serverBackendAddress); err != nil {
					core.Error("failed to send packet: %v", err)
					return
				}

				// todo: retry 5 times, once second apart until session update response is received

				sliceNumber += 1
			}
		}
	}()

}
