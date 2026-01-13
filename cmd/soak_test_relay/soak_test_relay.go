/*
   Network Next. Copyright 2017 - 2025 Network Next, Inc.
   Licensed under the Network Next Source Available License 1.0
*/

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
)

const ROUTE_REQUEST_PACKET = 1
const ROUTE_RESPONSE_PACKET = 2
const CLIENT_TO_SERVER_PACKET = 3
const SERVER_TO_CLIENT_PACKET = 4
const SESSION_PING_PACKET = 5
const SESSION_PONG_PACKET = 6
const CONTINUE_REQUEST_PACKET = 7
const CONTINUE_RESPONSE_PACKET = 8
const CLIENT_PING_PACKET = 9
const CLIENT_PONG_PACKET = 10
const RELAY_PING_PACKET = 11
const RELAY_PONG_PACKET = 12
const SERVER_PING_PACKET = 13
const SERVER_PONG_PACKET = 14

func Base64String(value string) []byte {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(err)
	}
	return data
}

const TestRelayPublicKey = "1nTj7bQmo8gfIDqG+o//GFsak/g1TRo4hl6XXw1JkyI="
const TestRelayPrivateKey = "cwvK44Pr5aHI3vE3siODS7CUgdPI/l1VwjVZ2FvEyAo="
const TestRelayBackendPublicKey = "IsjRpWEz9H7qslhWWupW4A9LIpVh+PzWoLleuXL1NUE="
const TestRelayBackendPrivateKey = "qXeUdLPZxaMnZ/zFHLHkmgkQOmunWq1AmRv55nqTYMg="

const (
	relayBin   = "./relay-debug"
	backendBin = "./func_backend"
)

func relay(name string, port int) *exec.Cmd {

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
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_BACKEND_PRIVATE_KEY=%s", TestRelayBackendPrivateKey))
	cmd.Env = append(cmd.Env, "RELAY_BACKEND_URL=http://127.0.0.1:30000")

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

func GenerateHeaderTag(packetType uint8, packetSequence uint64, sessionId uint64, sessionVersion uint8, sessionPrivateKey []byte) []byte {
	data := make([]byte, 32+1+8+8+1)
	index := 0
	copy(data[index:], sessionPrivateKey)
	index += 32
	data[index] = packetType
	index += 1
	binary.LittleEndian.PutUint64(data[index:], packetSequence)
	index += 8
	binary.LittleEndian.PutUint64(data[index:], sessionId)
	index += 8
	data[index] = sessionVersion
	result := sha256.Sum256(data)
	return result[0:8]
}

// =======================================================================================================================

func soak_test_relay(run_forever bool) {

	fmt.Printf("\nsoak_test_relay\n\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	if backend_cmd == nil {
		panic("failed to run backend")
	}

	time.Sleep(time.Second)

	relay_cmd := relay("local", 2000)

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

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	_ = testSecretKey

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

	_ = relayAddress

	// send packets

	startTime := time.Now()

	for {

		if !run_forever {
			duration := time.Now().Sub(startTime)
			if duration > time.Minute {
				break
			}
		}

		/*
			// send a bunch of random packets that don't pass the basic packet filter

			for i := 0; i < NumSockets; i++ {
				packet := make([]byte, common.RandomInt(1, 10*1024))
				common.RandomBytes(packet[:])
				conn[i].WriteToUDP(packet, &relayAddress)
			}

			// send a bunch of random packets that do pass the packet filters

			for i := 0; i < NumSockets; i++ {
				packet := make([]byte, common.RandomInt(18, 6000))
				common.RandomBytes(packet[:])
				var magic [constants.MagicBytes]byte
				fromAddress := core.GetAddressData(&clientAddress[i])
				toAddress := core.GetAddressData(&relayAddress)
				packetLength := len(packet)
				core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
				core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
				conn[i].WriteToUDP(packet, &relayAddress)
			}
		*/

		// send valid route request packets

		for i := 0; i < NumSockets; i++ {
			if rand.Intn(1000) == 0 {
				packet := make([]byte, 18+111*2)
				common.RandomBytes(packet[:])
				packet[0] = ROUTE_REQUEST_PACKET
				token := core.RouteToken{}
				sessionId[i] = rand.Uint64()
				sessionVersion[i] = 0
				sessionSequence[i] = 1
				token.SessionId = sessionId[i]
				token.SessionVersion = sessionVersion[i]
				token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
				token.NextAddress = clientAddress[i]
				token.PrevAddress = clientAddress[i]
				core.RandomBytes(sessionKey[i][:])
				copy(token.SessionPrivateKey[:], sessionKey[i][:])
				core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
				var magic [constants.MagicBytes]byte
				fromAddress := core.GetAddressData(&clientAddress[i])
				toAddress := core.GetAddressData(&relayAddress)
				packetLength := len(packet)
				core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
				core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
				conn[i].WriteToUDP(packet, &relayAddress)
			}
		}

		// send valid route response packets

		for i := 0; i < NumSockets; i++ {

			if rand.Intn(100) == 0 {

				packet := make([]byte, 18+25)

				packet[0] = ROUTE_RESPONSE_PACKET
				binary.LittleEndian.PutUint64(packet[18:], sessionSequence[i])
				binary.LittleEndian.PutUint64(packet[18+8:], sessionId[i])
				packet[18+8+8] = sessionVersion[i]

				tag := GenerateHeaderTag(ROUTE_RESPONSE_PACKET, sessionSequence[i], sessionId[i], sessionVersion[i], sessionKey[i][:])
				copy(packet[18+8+8+1:], tag)

				var magic [constants.MagicBytes]byte
				fromAddress := core.GetAddressData(&clientAddress[i])
				toAddress := core.GetAddressData(&relayAddress)

				packetLength := len(packet)

				core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

				core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

				conn[i].WriteToUDP(packet, &relayAddress)

				sessionSequence[i]++
			}
		}

		// send valid continue request packets

		for i := 0; i < NumSockets; i++ {
			if rand.Intn(100) == 0 {
				packet := make([]byte, 18+57*2)
				common.RandomBytes(packet[:])
				packet[0] = CONTINUE_REQUEST_PACKET
				token := core.ContinueToken{}
				token.SessionId = sessionId[i]
				token.SessionVersion = sessionVersion[i]
				sessionVersion[i]++
				token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
				core.WriteEncryptedContinueToken(&token, packet[18:], testSecretKey)
				var magic [constants.MagicBytes]byte
				fromAddress := core.GetAddressData(&clientAddress[i])
				toAddress := core.GetAddressData(&relayAddress)
				packetLength := len(packet)
				core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
				core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
				conn[i].WriteToUDP(packet, &relayAddress)
			}
		}

		// send valid continue response packets

		for i := 0; i < NumSockets; i++ {

			if rand.Intn(100) == 0 {

				packet := make([]byte, 18+25)

				packet[0] = CONTINUE_RESPONSE_PACKET
				binary.LittleEndian.PutUint64(packet[18:], sessionSequence[i])
				binary.LittleEndian.PutUint64(packet[18+8:], sessionId[i])
				packet[18+8+8] = sessionVersion[i]

				tag := GenerateHeaderTag(CONTINUE_RESPONSE_PACKET, sessionSequence[i], sessionId[i], sessionVersion[i], sessionKey[i][:])
				copy(packet[18+8+8+1:], tag)

				var magic [constants.MagicBytes]byte
				fromAddress := core.GetAddressData(&clientAddress[i])
				toAddress := core.GetAddressData(&relayAddress)

				packetLength := len(packet)

				core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

				core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

				conn[i].WriteToUDP(packet, &relayAddress)

				sessionSequence[i]++
			}
		}

		// send valid client to server packets

		for i := 0; i < NumSockets; i++ {

			if rand.Intn(10) == 0 {

				packet := make([]byte, 18+25+rand.Intn(1024))

				packet[0] = CLIENT_TO_SERVER_PACKET
				binary.LittleEndian.PutUint64(packet[18:], sessionSequence[i])
				binary.LittleEndian.PutUint64(packet[18+8:], sessionId[i])
				packet[18+8+8] = sessionVersion[i]

				tag := GenerateHeaderTag(CLIENT_TO_SERVER_PACKET, sessionSequence[i], sessionId[i], sessionVersion[i], sessionKey[i][:])
				copy(packet[18+8+8+1:], tag)

				var magic [constants.MagicBytes]byte
				fromAddress := core.GetAddressData(&clientAddress[i])
				toAddress := core.GetAddressData(&relayAddress)

				packetLength := len(packet)

				core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

				core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

				conn[i].WriteToUDP(packet, &relayAddress)

				sessionSequence[i]++
			}
		}

		// send valid server to client packets

		for i := 0; i < NumSockets; i++ {

			if rand.Intn(10) == 0 {

				packet := make([]byte, 18+25+rand.Intn(1024))

				packet[0] = SERVER_TO_CLIENT_PACKET
				binary.LittleEndian.PutUint64(packet[18:], sessionSequence[i])
				binary.LittleEndian.PutUint64(packet[18+8:], sessionId[i])
				packet[18+8+8] = sessionVersion[i]

				tag := GenerateHeaderTag(SERVER_TO_CLIENT_PACKET, sessionSequence[i], sessionId[i], sessionVersion[i], sessionKey[i][:])
				copy(packet[18+8+8+1:], tag)

				var magic [constants.MagicBytes]byte
				fromAddress := core.GetAddressData(&clientAddress[i])
				toAddress := core.GetAddressData(&relayAddress)

				packetLength := len(packet)

				core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

				core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

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

	for i := 0; i < NumSockets; i++ {
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
