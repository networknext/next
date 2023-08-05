/*
   Network Next Accelerate.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"time"
	"net"
	"context"
	"math/rand"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	// "github.com/networknext/next/modules/crypto"
)

func Base64String(value string) []byte {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(err)
	}
	return data
}

var TestRelayPublicKey = "9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14="
var TestRelayPrivateKey = "lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8="
var TestRelayBackendPublicKey = "SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y="
var TestRelayBackendPrivateKey = "ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M="

const (
	relayBin   = "./relay-debug"
	backendBin = "./func_backend"
)

type RelayConfig struct {
	num_threads int
	log_level int
}

func relay(name string, port int, configArray ...RelayConfig) (*exec.Cmd) {

	var config RelayConfig
	if len(configArray) == 1 {
		config = configArray[0]
	}

	cmd := exec.Command(relayBin)
	if cmd == nil {
		panic("could not create relay!\n")
		return nil
	}

	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_NAME=%s", name))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_PUBLIC_ADDRESS=127.0.0.1:%d", port))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_PUBLIC_KEY=%s", TestRelayPublicKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_PRIVATE_KEY=%s", TestRelayPrivateKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_BACKEND_PUBLIC_KEY=%s", TestRelayBackendPublicKey))
	cmd.Env = append(cmd.Env, "RELAY_BACKEND_HOSTNAME=http://127.0.0.1:30000")

	if config.num_threads != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_NUM_THREADS=%d", config.num_threads))
	}

	if config.log_level > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_LOG_LEVEL=%d", config.log_level))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Start()

	return cmd
}

func backend(mode string) (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command(backendBin)
	if cmd == nil {
		panic("could not create backend!\n")
		return nil, nil
	}

	cmd.Env = os.Environ()
	if mode != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("BACKEND_MODE=%s", mode))
	}

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Start()

	return cmd, &output
}

// =======================================================================================================================

func soak_test_relay(run_forever bool) {

	fmt.Printf("\nsoak_test_relay\n\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	if backend_cmd == nil {
		panic("failed to run backend")
	}

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 8
	config.log_level = 1
	
	relay_cmd := relay("relay", 2000, config)

	if relay_cmd == nil {
		panic("failed to run relay")
	}

	const NumSockets = 1024

	conn := make([]*net.UDPConn, NumSockets)

	clientAddress := make([]net.UDPAddr, NumSockets)

	sessionId := make([]uint64, NumSockets)
	sessionVersion := make([]uint8, NumSockets)
	sessionSequence := make([]uint64, NumSockets)
	sessionKey := make([][32]byte, NumSockets)

	publicKey := Base64String(TestRelayPublicKey)
	privateKey := Base64String(TestRelayBackendPrivateKey)

	for i := 0; i < NumSockets; i++ {
		lc := net.ListenConfig{}
		lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
		if err != nil {
			panic("could not bind socket")
		}
		conn[i] = lp.(*net.UDPConn)
		clientPort := conn[i].LocalAddr().(*net.UDPAddr).Port
		clientAddress[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", clientPort))
	}

	relayAddress := core.ParseAddress("127.0.0.1:2000")

	// send packets

	startTime := time.Now()

 	for {

 		if !run_forever {
	 		duration := time.Now().Sub(startTime)
	 		if duration > 5 * time.Minute {
	 			break
	 		}
 		}

 		// send a bunch of random packets that don't pass the basic packet filter

		for i := 0; i < NumSockets; i++ {
			packet := make([]byte, common.RandomInt(1,10*1024))
			common.RandomBytes(packet[:])
			conn[i].WriteToUDP(packet, &relayAddress)
		}

 		// send a bunch of random packets that pass the packet filters

		for i := 0; i < NumSockets; i++ {
			packet := make([]byte, common.RandomInt(18,6000))
			common.RandomBytes(packet[:])
			var magic [constants.MagicBytes]byte
			var fromAddressBuffer [32]byte
			var toAddressBuffer [32]byte
			fromAddress, fromPort := core.GetAddressData(&clientAddress[i], fromAddressBuffer[:])
			toAddress, toPort := core.GetAddressData(&relayAddress, toAddressBuffer[:])
			packetLength := len(packet)
			core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			conn[i].WriteToUDP(packet, &relayAddress)
		}

		// send valid route request packets

		for i := 0; i < NumSockets; i++ {
			if rand.Intn(1000) == 0 {
				packet := make([]byte, 18 + 111*2)
				common.RandomBytes(packet[:])
				packet[0] = 9 // ROUTE_REQUEST_PACKET
				token := core.RouteToken{}
				sessionId[i] = rand.Uint64()
				sessionVersion[i] = 0
				sessionSequence[i] = 1
				token.SessionId = sessionId[i]
				token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
				token.NextAddress = clientAddress[i]
				token.PrevAddress = clientAddress[i]
				core.RandomBytes(sessionKey[i][:])
				copy(token.PrivateKey[:], sessionKey[i][:])
				core.WriteEncryptedRouteToken(&token, packet[16:], privateKey, publicKey)
				var magic [constants.MagicBytes]byte
				var fromAddressBuffer [32]byte
				var toAddressBuffer [32]byte
				fromAddress, fromPort := core.GetAddressData(&clientAddress[i], fromAddressBuffer[:])
				toAddress, toPort := core.GetAddressData(&relayAddress, toAddressBuffer[:])
				packetLength := len(packet)
				core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				conn[i].WriteToUDP(packet, &relayAddress)
			}
		}

		// send valid route response packets

		for i := 0; i < NumSockets; i++ {

			if rand.Intn(100) == 0 {

				packet := make([]byte, 18 + 33)
				
				packet[0] = 10 // ROUTE_RESPONSE_PACKET
				binary.LittleEndian.PutUint64(packet[16:], sessionSequence[i])
				binary.LittleEndian.PutUint64(packet[16+8:], sessionId[i])

				nonce := [12]byte{}
				binary.LittleEndian.PutUint32(nonce[0:], 10) // ROUTE_RESPONSE_PACKET 
				binary.LittleEndian.PutUint64(nonce[4:], sessionSequence[i])

				additional := packet[16+8:16+8+8+1]

				buffer := packet[16+8+8+1:18+33-2]

				encryptedLength := uint64(0)

				additionalLength := uint64(9)

				result := C.crypto_aead_chacha20poly1305_ietf_encrypt(
					(*C.uchar)(&buffer[0]),
					(*C.ulonglong)(&encryptedLength),
					(*C.uchar)(&buffer[0]),
					(C.ulonglong)(0),
					(*C.uchar)(&additional[0]),
					(C.ulonglong)(additionalLength),
					(*C.uchar)(nil),
					(*C.uchar)(&nonce[0]),
					(*C.uchar)(&sessionKey[i][0]),
				)

				if result != 0 {
					panic("crypto_aead_chacha20poly1305_ietf_encrypt failed")
				}

				var magic [constants.MagicBytes]byte
				var fromAddressBuffer [32]byte
				var toAddressBuffer [32]byte
				fromAddress, fromPort := core.GetAddressData(&clientAddress[i], fromAddressBuffer[:])
				toAddress, toPort := core.GetAddressData(&relayAddress, toAddressBuffer[:])
				
				packetLength := len(packet)
				
				core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				conn[i].WriteToUDP(packet, &relayAddress)

				sessionSequence[i]++
			}
		}

		// send valid continue request packets

		for i := 0; i < NumSockets; i++ {
			if rand.Intn(100) == 0 {
				packet := make([]byte, 18 + 57*2)
				common.RandomBytes(packet[:])
				packet[0] = 15 // CONTINUE_REQUEST_PACKET
				token := core.ContinueToken{}
				token.SessionId = sessionId[i]
				token.SessionVersion = sessionVersion[i]
				sessionVersion[i]++
				token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
				core.WriteEncryptedContinueToken(&token, packet[16:], privateKey, publicKey)
				var magic [constants.MagicBytes]byte
				var fromAddressBuffer [32]byte
				var toAddressBuffer [32]byte
				fromAddress, fromPort := core.GetAddressData(&clientAddress[i], fromAddressBuffer[:])
				toAddress, toPort := core.GetAddressData(&relayAddress, toAddressBuffer[:])
				packetLength := len(packet)
				core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				conn[i].WriteToUDP(packet, &relayAddress)
			}
		}

		// send valid continue response packets

		for i := 0; i < NumSockets; i++ {

			if rand.Intn(100) == 0 {

				packet := make([]byte, 18 + 33)
				
				packet[0] = 16 // CONTINUE_RESPONSE_PACKET
				binary.LittleEndian.PutUint64(packet[16:], sessionSequence[i])
				binary.LittleEndian.PutUint64(packet[16+8:], sessionId[i])

				nonce := [12]byte{}
				binary.LittleEndian.PutUint32(nonce[0:], 16) // CONTINUE_RESPONSE_PACKET 
				binary.LittleEndian.PutUint64(nonce[4:], sessionSequence[i])

				additional := packet[16+8:16+8+8+1]

				buffer := packet[16+8+8+1:18+33-2]

				encryptedLength := uint64(0)

				additionalLength := uint64(9)

				result := C.crypto_aead_chacha20poly1305_ietf_encrypt(
					(*C.uchar)(&buffer[0]),
					(*C.ulonglong)(&encryptedLength),
					(*C.uchar)(&buffer[0]),
					(C.ulonglong)(0),
					(*C.uchar)(&additional[0]),
					(C.ulonglong)(additionalLength),
					(*C.uchar)(nil),
					(*C.uchar)(&nonce[0]),
					(*C.uchar)(&sessionKey[i][0]),
				)

				if result != 0 {
					panic("crypto_aead_chacha20poly1305_ietf_encrypt failed")
				}

				var magic [constants.MagicBytes]byte
				var fromAddressBuffer [32]byte
				var toAddressBuffer [32]byte
				fromAddress, fromPort := core.GetAddressData(&clientAddress[i], fromAddressBuffer[:])
				toAddress, toPort := core.GetAddressData(&relayAddress, toAddressBuffer[:])
				
				packetLength := len(packet)
				
				core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				conn[i].WriteToUDP(packet, &relayAddress)

				sessionSequence[i]++
			}
		}

		// send valid client to server packets

		for i := 0; i < NumSockets; i++ {

			if rand.Intn(10) == 0 {

				packet := make([]byte, 18 + 33 + rand.Intn(1024))
				
				packet[0] = 11 // CLIENT_TO_SERVER_PACKET
				binary.LittleEndian.PutUint64(packet[16:], sessionSequence[i])
				binary.LittleEndian.PutUint64(packet[16+8:], sessionId[i])

				nonce := [12]byte{}
				binary.LittleEndian.PutUint32(nonce[0:], 11) // CLIENT_TO_SERVER_PACKET
				binary.LittleEndian.PutUint64(nonce[4:], sessionSequence[i])

				additional := packet[16+8:16+8+8+1]

				buffer := packet[16+8+8+1:18+33-2]

				encryptedLength := uint64(0)

				additionalLength := uint64(9)

				result := C.crypto_aead_chacha20poly1305_ietf_encrypt(
					(*C.uchar)(&buffer[0]),
					(*C.ulonglong)(&encryptedLength),
					(*C.uchar)(&buffer[0]),
					(C.ulonglong)(0),
					(*C.uchar)(&additional[0]),
					(C.ulonglong)(additionalLength),
					(*C.uchar)(nil),
					(*C.uchar)(&nonce[0]),
					(*C.uchar)(&sessionKey[i][0]),
				)

				if result != 0 {
					panic("crypto_aead_chacha20poly1305_ietf_encrypt failed")
				}

				var magic [constants.MagicBytes]byte
				var fromAddressBuffer [32]byte
				var toAddressBuffer [32]byte
				fromAddress, fromPort := core.GetAddressData(&clientAddress[i], fromAddressBuffer[:])
				toAddress, toPort := core.GetAddressData(&relayAddress, toAddressBuffer[:])
				
				packetLength := len(packet)
				
				core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				conn[i].WriteToUDP(packet, &relayAddress)

				sessionSequence[i]++
			}
		}

		// send valid server to client packets

		for i := 0; i < NumSockets; i++ {

			if rand.Intn(10) == 0 {

				packet := make([]byte, 18 + 33 + rand.Intn(1024))
				
				packet[0] = 12 // SERVER_TO_CLIENT_PACKET
				binary.LittleEndian.PutUint64(packet[16:], sessionSequence[i])
				binary.LittleEndian.PutUint64(packet[16+8:], sessionId[i])

				nonce := [12]byte{}
				binary.LittleEndian.PutUint32(nonce[0:], 12) // SERVER_TO_CLIENT_PACKET
				binary.LittleEndian.PutUint64(nonce[4:], sessionSequence[i])

				additional := packet[16+8:16+8+8+1]

				buffer := packet[16+8+8+1:18+33-2]

				encryptedLength := uint64(0)

				additionalLength := uint64(9)

				result := C.crypto_aead_chacha20poly1305_ietf_encrypt(
					(*C.uchar)(&buffer[0]),
					(*C.ulonglong)(&encryptedLength),
					(*C.uchar)(&buffer[0]),
					(C.ulonglong)(0),
					(*C.uchar)(&additional[0]),
					(C.ulonglong)(additionalLength),
					(*C.uchar)(nil),
					(*C.uchar)(&nonce[0]),
					(*C.uchar)(&sessionKey[i][0]),
				)

				if result != 0 {
					panic("crypto_aead_chacha20poly1305_ietf_encrypt failed")
				}

				var magic [constants.MagicBytes]byte
				var fromAddressBuffer [32]byte
				var toAddressBuffer [32]byte
				fromAddress, fromPort := core.GetAddressData(&clientAddress[i], fromAddressBuffer[:])
				toAddress, toPort := core.GetAddressData(&relayAddress, toAddressBuffer[:])
				
				packetLength := len(packet)
				
				core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				conn[i].WriteToUDP(packet, &relayAddress)

				sessionSequence[i]++
			}
		}

		// send session ping packets

		for i := 0; i < NumSockets; i++ {

			if rand.Intn(100) == 0 {

				packet := make([]byte, 18 + 33 + 8)
				
				packet[0] = 13 // SESSION_PING_PACKET
				binary.LittleEndian.PutUint64(packet[16:], sessionSequence[i])
				binary.LittleEndian.PutUint64(packet[16+8:], sessionId[i])

				nonce := [12]byte{}
				binary.LittleEndian.PutUint32(nonce[0:], 13) // SESSION_PING_PACKET
				binary.LittleEndian.PutUint64(nonce[4:], sessionSequence[i])

				additional := packet[16+8:16+8+8+1]

				buffer := packet[16+8+8+1:18+33-2]

				encryptedLength := uint64(0)

				additionalLength := uint64(9)

				result := C.crypto_aead_chacha20poly1305_ietf_encrypt(
					(*C.uchar)(&buffer[0]),
					(*C.ulonglong)(&encryptedLength),
					(*C.uchar)(&buffer[0]),
					(C.ulonglong)(0),
					(*C.uchar)(&additional[0]),
					(C.ulonglong)(additionalLength),
					(*C.uchar)(nil),
					(*C.uchar)(&nonce[0]),
					(*C.uchar)(&sessionKey[i][0]),
				)

				if result != 0 {
					panic("crypto_aead_chacha20poly1305_ietf_encrypt failed")
				}

				var magic [constants.MagicBytes]byte
				var fromAddressBuffer [32]byte
				var toAddressBuffer [32]byte
				fromAddress, fromPort := core.GetAddressData(&clientAddress[i], fromAddressBuffer[:])
				toAddress, toPort := core.GetAddressData(&relayAddress, toAddressBuffer[:])
				
				packetLength := len(packet)
				
				core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				conn[i].WriteToUDP(packet, &relayAddress)

				sessionSequence[i]++
			}
		}

		// send session pong packets

		for i := 0; i < NumSockets; i++ {

			if rand.Intn(100) == 0 {

				packet := make([]byte, 18 + 33 + 8)
				
				packet[0] = 14 // SESSION_PONG_PACKET
				binary.LittleEndian.PutUint64(packet[16:], sessionSequence[i])
				binary.LittleEndian.PutUint64(packet[16+8:], sessionId[i])

				nonce := [12]byte{}
				binary.LittleEndian.PutUint32(nonce[0:], 14) // SESSION_PONG_PACKET
				binary.LittleEndian.PutUint64(nonce[4:], sessionSequence[i])

				additional := packet[16+8:16+8+8+1]

				buffer := packet[16+8+8+1:18+33-2]

				encryptedLength := uint64(0)

				additionalLength := uint64(9)

				result := C.crypto_aead_chacha20poly1305_ietf_encrypt(
					(*C.uchar)(&buffer[0]),
					(*C.ulonglong)(&encryptedLength),
					(*C.uchar)(&buffer[0]),
					(C.ulonglong)(0),
					(*C.uchar)(&additional[0]),
					(C.ulonglong)(additionalLength),
					(*C.uchar)(nil),
					(*C.uchar)(&nonce[0]),
					(*C.uchar)(&sessionKey[i][0]),
				)

				if result != 0 {
					panic("crypto_aead_chacha20poly1305_ietf_encrypt failed")
				}

				var magic [constants.MagicBytes]byte
				var fromAddressBuffer [32]byte
				var toAddressBuffer [32]byte
				fromAddress, fromPort := core.GetAddressData(&clientAddress[i], fromAddressBuffer[:])
				toAddress, toPort := core.GetAddressData(&relayAddress, toAddressBuffer[:])
				
				packetLength := len(packet)
				
				core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
				
				conn[i].WriteToUDP(packet, &relayAddress)

				sessionSequence[i]++
			}
		}

		time.Sleep(10 * time.Millisecond)
	}

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	for i := 0; i < NumSockets; i++{
		conn[i].Close()
	}

	fmt.Printf("Success!\n")
}

func main() {
	run_forever := true
	if len(os.Args) > 1 {
		run_forever = false
	}
	soak_test_relay(run_forever)
}
