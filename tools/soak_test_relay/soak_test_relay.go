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
	"fmt"
	"os"
	"os/exec"
	"time"
	"net"
	"context"

	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/common"
	"github.com/networknext/accelerate/modules/core"
	// "github.com/networknext/accelerate/modules/crypto"
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
	relayBin   = "./dist/relay-debug"
	backendBin = "./dist/func_backend"
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

func soak_test_relay() {

	fmt.Printf("\nsoak_test_relay\n\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 4
	config.log_level = 1
	
	relay_cmd := relay("relay", 2000, config)

	const NumSockets = 1024

	conn := make([]*net.UDPConn, NumSockets)

	clientAddress := make([]net.UDPAddr, NumSockets)

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

	serverAddress := core.ParseAddress("127.0.0.1:2000")

 	for {

 		// send a bunch of random packets that don't pass the basic packet filter

		for i := 0; i < NumSockets; i++ {
			packet := make([]byte, common.RandomInt(1,10*1024))
			common.RandomBytes(packet[:])
			conn[i].WriteToUDP(packet, &serverAddress)
		}

 		// send a bunch of random packets that pass the packet filters

		for i := 0; i < NumSockets; i++ {
			packet := make([]byte, common.RandomInt(18,6000))
			common.RandomBytes(packet[:])
			var magic [constants.MagicBytes]byte
			var fromAddressBuffer [32]byte
			var toAddressBuffer [32]byte
			fromAddress, fromPort := core.GetAddressData(&clientAddress[i], fromAddressBuffer[:])
			toAddress, toPort := core.GetAddressData(&serverAddress, toAddressBuffer[:])
			packetLength := len(packet)
			core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			conn[i].WriteToUDP(packet, &serverAddress)
		}

		// wait

		time.Sleep(10 * time.Millisecond)
	}

	for i := 0; i < NumSockets; i++{
		conn[i].Close()
	}

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()
}

func main() {
	soak_test_relay()
}
