/*
   Network Next Accelerate.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"time"
	"net"
	"net/http"
	"io/ioutil"
	"context"
	"syscall"

	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/common"
	"github.com/networknext/accelerate/modules/core"
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
	relayBin   = "./relay"
	backendBin = "./func_backend"
)

type RelayConfig struct {
	fake_packet_loss_percent          float32
	fake_packet_loss_start_time       float32
	omit_relay_name                   bool
	omit_relay_public_address         bool
	invalid_relay_public_address      bool
	invalid_relay_internal_address    bool
	omit_relay_public_key             bool
	invalid_relay_public_key          bool
	omit_relay_private_key            bool
	invalid_relay_private_key         bool
	invalid_relay_keypair             bool
	omit_relay_backend_public_key     bool
	invalid_relay_backend_public_key  bool
	mismatch_relay_backend_public_key bool
	omit_relay_backend_hostname       bool
	bind_to_port_zero                 bool
	num_threads                       int
	print_counters                    bool
}

func relay(name string, port int, configArray ...RelayConfig) (*exec.Cmd, *bytes.Buffer) {

	var config RelayConfig
	if len(configArray) == 1 {
		config = configArray[0]
	}

	cmd := exec.Command(relayBin)
	if cmd == nil {
		panic("could not create relay!\n")
		return nil, nil
	}

	if !config.omit_relay_name {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_NAME=%s", name))
	}

	if !config.omit_relay_public_address {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_PUBLIC_ADDRESS=127.0.0.1:%d", port))
	}

	if config.invalid_relay_public_address {
		cmd.Env = append(cmd.Env, "RELAY_PUBLIC_ADDRESS=blahblahblah")
	}

	if config.invalid_relay_internal_address {
		cmd.Env = append(cmd.Env, "RELAY_INTERNAL_ADDRESS=blahblahblah")
	}

	if !config.omit_relay_public_key {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_PUBLIC_KEY=%s", TestRelayPublicKey))
	}

	if config.invalid_relay_public_key {
		cmd.Env = append(cmd.Env, "RELAY_PUBLIC_KEY=blahblahblah")
	}

	if !config.omit_relay_private_key {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_PRIVATE_KEY=%s", TestRelayPrivateKey))
	}

	if config.invalid_relay_private_key {
		cmd.Env = append(cmd.Env, "RELAY_PRIVATE_KEY=blahblahblah")
	}

	if config.invalid_relay_keypair {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_PUBLIC_KEY=%s", TestRelayPrivateKey))
		cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_PRIVATE_KEY=%s", TestRelayPublicKey))
	}

	if !config.omit_relay_backend_public_key {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_BACKEND_PUBLIC_KEY=%s", TestRelayBackendPublicKey))
	}

	if config.invalid_relay_backend_public_key {
		cmd.Env = append(cmd.Env, "RELAY_BACKEND_PUBLIC_KEY=blahblahblah")
	}

	if config.mismatch_relay_backend_public_key {
		cmd.Env = append(cmd.Env, "RELAY_BACKEND_PUBLIC_KEY=9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=")
	}

	if !config.omit_relay_backend_hostname {
		cmd.Env = append(cmd.Env, "RELAY_BACKEND_HOSTNAME=http://127.0.0.1:30000")
	}

	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_FAKE_PACKET_LOSS_PERCENT=%f", config.fake_packet_loss_percent))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_FAKE_PACKET_LOSS_START_TIME=%f", config.fake_packet_loss_start_time))

	if config.bind_to_port_zero {
		cmd.Env = append(cmd.Env, "RELAY_PUBLIC_ADDRESS=127.0.0.1:0")
	}

	if config.num_threads != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_NUM_THREADS=%d", config.num_threads))
	}

	if config.print_counters {
		cmd.Env = append(cmd.Env, "RELAY_PRINT_COUNTERS=1")
	}

	// fmt.Printf("%s\n", cmd.Env)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Start()

	return cmd, &output
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

func test_initialize_success() {

	fmt.Printf("test_initialize_success\n")

	backend_cmd, _ := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 16

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	time.Sleep(10 * time.Second)

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}
}

func test_initialize_fail() {

	fmt.Printf("test_initialize_fail\n")

	config := RelayConfig{}
	config.num_threads = 16

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: could not post relay update") {
		panic("relay should not be able to post relay update, relay backend does not exist")
	}

	if !strings.Contains(relay_stdout.String(), "error: could not update relay 30 times in a row. shutting down") {
		panic("relay should shut down when it can't initialize")
	}

	if !strings.Contains(relay_stdout.String(), "Done.\n") {
		panic("relay should shut down clean")
	}
}

func test_relay_name_not_set() {

	fmt.Printf("test_relay_name_not_set\n")

	config := RelayConfig{}
	config.omit_relay_name = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: RELAY_NAME not set") {
		panic("relay should not start without a relay name")
	}
}

func test_relay_public_address_not_set() {

	fmt.Printf("test_relay_public_address_not_set\n")

	config := RelayConfig{}
	config.omit_relay_public_address = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: RELAY_PUBLIC_ADDRESS not set") {
		panic("relay should not start without a public address")
	}
}

func test_relay_public_address_invalid() {

	fmt.Printf("test_relay_public_address_invalid\n")

	config := RelayConfig{}
	config.invalid_relay_public_address = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: invalid relay public address 'blahblahblah'") {
		panic("relay should not start with an invalid public address")
	}
}

func test_relay_internal_address_invalid() {

	fmt.Printf("test_relay_internal_address_invalid\n")

	config := RelayConfig{}
	config.invalid_relay_internal_address = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: invalid relay internal address 'blahblahblah'") {
		panic("relay should not start with an invalid internal address")
	}
}

func test_relay_public_key_not_set() {

	fmt.Printf("test_relay_public_key_not_set\n")

	config := RelayConfig{}
	config.omit_relay_public_key = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: RELAY_PUBLIC_KEY not set") {
		panic("relay should not start without a relay public key")
	}
}

func test_relay_public_key_invalid() {

	fmt.Printf("test_relay_public_key_invalid\n")

	config := RelayConfig{}
	config.invalid_relay_public_key = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: invalid relay public key") {
		panic("relay should not start with an invalid relay public key")
	}
}

func test_relay_private_key_not_set() {

	fmt.Printf("test_relay_private_key_not_set\n")

	config := RelayConfig{}
	config.omit_relay_private_key = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: RELAY_PRIVATE_KEY not set") {
		panic("relay should not start without a relay private key")
	}
}

func test_relay_private_key_invalid() {

	fmt.Printf("test_relay_private_key_invalid\n")

	config := RelayConfig{}
	config.invalid_relay_private_key = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: invalid relay private key") {
		panic("relay should not start with an invalid relay private key")
	}
}

func test_relay_keypair_invalid() {

	fmt.Printf("test_relay_keypair_invalid\n")

	config := RelayConfig{}
	config.invalid_relay_keypair = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: relay keypair is invalid") {
		panic("relay should not start with an invalid relay keypair")
	}
}

func test_relay_backend_public_key_not_set() {

	fmt.Printf("test_relay_backend_public_key_not_set\n")

	config := RelayConfig{}
	config.omit_relay_backend_public_key = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: RELAY_BACKEND_PUBLIC_KEY not set") {
		panic("relay should not start without a relay backend public key")
	}
}

func test_relay_backend_public_key_invalid() {

	fmt.Printf("test_relay_backend_public_key_invalid\n")

	config := RelayConfig{}
	config.invalid_relay_backend_public_key = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: invalid relay backend public key") {
		panic("relay should not start with an invalid relay backend public key")
	}
}

func test_relay_backend_public_key_mismatch() {

	fmt.Printf("test_relay_backend_public_key_mismatch\n")

	backend_cmd, _ := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.mismatch_relay_backend_public_key = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	backend_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()

	fmt.Printf("=======================================\n%s=============================================\n", relay_stdout)

	if !strings.Contains(relay_stdout.String(), "error: relay update response is 400. the relay backend is down or the relay is misconfigured. check RELAY_BACKEND_PUBLIC_KEY") {
		panic("relay cannot talk to the relay backend unless it has the correct relay backend public key")
	}
}

func test_relay_backend_hostname_not_set() {

	fmt.Printf("test_relay_backend_hostname_not_set\n")

	config := RelayConfig{}
	config.omit_relay_backend_hostname = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: RELAY_BACKEND_HOSTNAME not set") {
		panic("relay should not start without a relay backend hostname")
	}
}

func test_relay_cant_bind_to_port_zero() {

	fmt.Printf("test_relay_cant_bind_to_port_zero\n")

	config := RelayConfig{}
	config.bind_to_port_zero = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: you must specify a valid port number!") {
		panic("relay should not be able to bind to port zero")
	}
}

func test_num_threads() {

	fmt.Printf("test_num_threads\n")

	backend_cmd, _ := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 16

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	time.Sleep(10 * time.Second)

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	for i := 0; i < config.num_threads; i++ {
		if !strings.Contains(relay_stdout.String(), fmt.Sprintf("Creating relay thread %d", i)) {
			panic("missing relay thread")
		}
	}
}

func test_relay_pings() {

	fmt.Printf("test_relay_pings\n")

	backend_cmd, _ := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 4
	config.print_counters = true

	relay_1_cmd, relay_1_stdout := relay("relay", 2000, config)
	relay_2_cmd, relay_2_stdout := relay("relay", 2001, config)
	relay_3_cmd, relay_3_stdout := relay("relay", 2002, config)

	time.Sleep(15 * time.Second)

	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	if !strings.Contains(relay_1_stdout.String(), "Relay initialized") {
		panic("could not initialize relay 1")
	}

	if !strings.Contains(relay_2_stdout.String(), "Relay initialized") {
		panic("could not initialize relay 2")
	}

	if !strings.Contains(relay_3_stdout.String(), "Relay initialized") {
		panic("could not initialize relay 3")
	}

	relay_stdout := []string{relay_1_stdout.String(), relay_2_stdout.String(), relay_3_stdout.String()}

	for i := range relay_stdout {

		checkCounter("RELAY_COUNTER_RELAY_PING_PACKET_SENT", relay_stdout[i])
		checkCounter("RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED", relay_stdout[i])
		checkCounter("RELAY_COUNTER_RELAY_PONG_PACKET_SENT", relay_stdout[i])
		checkCounter("RELAY_COUNTER_RELAY_PONG_PACKET_RECEIVED", relay_stdout[i])
		checkCounter("RELAY_COUNTER_PONGS_PROCESSED", relay_stdout[i])
	}
}

func getCostMatrix() *common.CostMatrix {

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := httpClient.Get("http://127.0.0.1:30000/cost_matrix")
	if err != nil {
		panic("failed to http get cost matrix")
	}

	buffer, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic("failed to read response body")
	}

	response.Body.Close()

	if len(buffer) == 0 {
		panic("cost matrix is empty")
	}

	costMatrix := common.CostMatrix{}

	err = costMatrix.Read(buffer)
	if err != nil {
		panic("failed to read cost matrix")
	}

	return &costMatrix
}

func test_cost_matrix() {

	fmt.Printf("test_cost_matrix\n")

	backend_cmd, _ := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 4
	config.print_counters = true

	relay_1_cmd, relay_1_stdout := relay("relay", 2000, config)
	relay_2_cmd, relay_2_stdout := relay("relay", 2001, config)
	relay_3_cmd, relay_3_stdout := relay("relay", 2002, config)

	time.Sleep(15 * time.Second)

	costMatrix := getCostMatrix()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	if !strings.Contains(relay_1_stdout.String(), "Relay initialized") {
		panic("could not initialize relay 1")
	}

	if !strings.Contains(relay_2_stdout.String(), "Relay initialized") {
		panic("could not initialize relay 2")
	}

	if !strings.Contains(relay_3_stdout.String(), "Relay initialized") {
		panic("could not initialize relay 3")
	}

	if len(costMatrix.Costs) != 3 {
		panic("cost matrix should have three entries")
	}

	if costMatrix.Costs[0] > 5 || costMatrix.Costs[1] > 5 || costMatrix.Costs[2] > 5 {
		panic(fmt.Sprintf("cost matrix entries are invalid: %+v", costMatrix.Costs))
	}
}

func test_basic_packet_filter() {

	fmt.Printf("test_basic_packet_filter\n")

	backend_cmd, _ := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 4
	config.print_counters = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	conn, err := net.Dial("udp", "127.0.0.1:2000")
	if err != nil {
		panic("could not create udp socket")
	}
 
 	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, common.RandomInt(1,1500))
			common.RandomBytes(packet[:])
			conn.Write(packet)
		}
		time.Sleep(time.Second)
	}

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_BASIC_PACKET_FILTER_DROPPED_PACKET", relay_stdout.String())
}

func test_advanced_packet_filter() {

	fmt.Printf("test_advanced_packet_filter\n")

	backend_cmd, _ := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 4
	config.print_counters = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind socket")
	}

	conn := lp.(*net.UDPConn)

	serverAddress := core.ParseAddress("127.0.0.1:2000")

 	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, common.RandomInt(18,1500))
			common.RandomBytes(packet[:])
			var magic [8]byte
			var fromAddress [4]byte
			var toAddress [4]byte
			common.RandomBytes(magic[:])
			common.RandomBytes(fromAddress[:])
			common.RandomBytes(toAddress[:])
			fromPort := uint16(j + 1000000)
			toPort := uint16(j + 5000)
			packetLength := len(packet)
			core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			conn.WriteToUDP(packet, &serverAddress)
		}
		time.Sleep(time.Second)
	}

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_ADVANCED_PACKET_FILTER_DROPPED_PACKET", relay_stdout.String())
}

func test_clean_shutdown() {

	fmt.Printf("test_clean_shutdown\n")

	backend_cmd, _ := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 4
	config.print_counters = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	time.Sleep(10 * time.Second)
	
	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	relay_cmd.Process.Signal(syscall.SIGHUP)

	relay_cmd.Wait()

	backend_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Clean shutdown...") {
		panic("did not detect clean shutdown start")
	}

	for i := 0; i <= 60; i++ {
		if !strings.Contains(relay_stdout.String(), fmt.Sprintf("Shutting down in %d seconds", i)) {
			panic(fmt.Sprintf("missing shutdown in %d seconds", i))
		}
	}

	if !strings.Contains(relay_stdout.String(), "Clean shutdown completed") {
		panic("clean shutdown did not complete")
	}

	if !strings.Contains(relay_stdout.String(), "Done.") {
		panic("relay crashed while shutting down\n")
	}
}

func test_unknown_packets() {

	fmt.Printf("test_unknown_packets\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 4
	config.print_counters = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	time.Sleep(5 * time.Second)

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind socket")
	}

	conn := lp.(*net.UDPConn)

	clientPort := conn.LocalAddr().(*net.UDPAddr).Port

	clientAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", clientPort))

	serverAddress := core.ParseAddress("127.0.0.1:2000")

 	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, common.RandomInt(18,1500))
			common.RandomBytes(packet[:])
			var magic [constants.MagicBytes]byte
			var fromAddressBuffer [32]byte
			var toAddressBuffer [32]byte
			fromAddress, fromPort := core.GetAddressData(&clientAddress, fromAddressBuffer[:])
			toAddress, toPort := core.GetAddressData(&serverAddress, toAddressBuffer[:])
			packetLength := len(packet)
			core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			conn.WriteToUDP(packet, &serverAddress)
		}
		time.Sleep(time.Second)
	}

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_UNKNOWN_PACKETS", relay_stdout.String())
}

func test_near_ping_packet_wrong_size() {

	fmt.Printf("test_near_ping_packet_wrong_size\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 4
	config.print_counters = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	time.Sleep(5 * time.Second)

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind socket")
	}

	conn := lp.(*net.UDPConn)

	clientPort := conn.LocalAddr().(*net.UDPAddr).Port

	clientAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", clientPort))

	serverAddress := core.ParseAddress("127.0.0.1:2000")

 	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, common.RandomInt(18,1500))
			packet[0] = 20 // NEAR_PING_PACKET
			common.RandomBytes(packet[:])
			var magic [constants.MagicBytes]byte
			var fromAddressBuffer [32]byte
			var toAddressBuffer [32]byte
			fromAddress, fromPort := core.GetAddressData(&clientAddress, fromAddressBuffer[:])
			toAddress, toPort := core.GetAddressData(&serverAddress, toAddressBuffer[:])
			packetLength := len(packet)
			core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			conn.WriteToUDP(packet, &serverAddress)
		}
		time.Sleep(time.Second)
	}

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_NEAR_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_NEAR_PING_PACKET_WRONG_SIZE", relay_stdout.String())
}

func test_near_ping_packet_expired() {

	fmt.Printf("test_near_ping_packet_expired\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 4
	config.print_counters = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	time.Sleep(5 * time.Second)

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind socket")
	}

	conn := lp.(*net.UDPConn)

	clientPort := conn.LocalAddr().(*net.UDPAddr).Port

	clientAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", clientPort))

	serverAddress := core.ParseAddress("127.0.0.1:2000")

 	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18 + 8 + 8 + 8 + 32)
			packet[0] = 20 // NEAR_PING_PACKET
			var magic [constants.MagicBytes]byte
			var fromAddressBuffer [32]byte
			var toAddressBuffer [32]byte
			fromAddress, fromPort := core.GetAddressData(&clientAddress, fromAddressBuffer[:])
			toAddress, toPort := core.GetAddressData(&serverAddress, toAddressBuffer[:])
			packetLength := len(packet)
			core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			conn.WriteToUDP(packet, &serverAddress)
		}
		time.Sleep(time.Second)
	}

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_NEAR_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_NEAR_PING_PACKET_EXPIRED", relay_stdout.String())
}

func test_near_ping_packet_did_not_verify() {

	fmt.Printf("test_near_ping_packet_did_not_verify\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 4
	config.print_counters = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	time.Sleep(5 * time.Second)

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind socket")
	}

	conn := lp.(*net.UDPConn)

	clientPort := conn.LocalAddr().(*net.UDPAddr).Port

	clientAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", clientPort))

	serverAddress := core.ParseAddress("127.0.0.1:2000")

 	for i := 0; i < 10; i++ {
 		expireTimestamp := time.Now().Unix() + 10
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18 + 8 + 8 + 8 + 32)
			packet[0] = 20 // NEAR_PING_PACKET
			binary.LittleEndian.PutUint64(packet[16+8+8:], uint64(expireTimestamp))
			var magic [constants.MagicBytes]byte
			var fromAddressBuffer [32]byte
			var toAddressBuffer [32]byte
			fromAddress, fromPort := core.GetAddressData(&clientAddress, fromAddressBuffer[:])
			toAddress, toPort := core.GetAddressData(&serverAddress, toAddressBuffer[:])
			packetLength := len(packet)
			core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			conn.WriteToUDP(packet, &serverAddress)
		}
		time.Sleep(time.Second)
	}

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_NEAR_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_NEAR_PING_PACKET_DID_NOT_VERIFY", relay_stdout.String())
}

func test_near_ping_packet_responded_with_pong() {

	fmt.Printf("test_near_ping_packet_responded_with_pong\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.num_threads = 4
	config.print_counters = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	time.Sleep(5 * time.Second)

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind socket")
	}

	conn := lp.(*net.UDPConn)

	clientPort := conn.LocalAddr().(*net.UDPAddr).Port

	clientAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", clientPort))

	serverAddress := core.ParseAddress("127.0.0.1:2000")

	sequence := uint64(0)
	sessionId := uint64(0x123498173981377)

	pingKey := make([]byte, 32)

	receivedPong := false

	go func() {
		for {
			receiveBuffer := make([]byte, 1500)
			receivePacketBytes, from, err := conn.ReadFromUDP(receiveBuffer[:])
			if err != nil {
				break
			}
			if receivePacketBytes == 18 + 8 + 8 && receiveBuffer[0] == 21 && from.String() == serverAddress.String() {
				receivedPong = true
				break
			}
		}
	}()

 	for i := 0; i < 10; i++ {

 		expireTimestamp := uint64(time.Now().Unix()) + 10

		pingToken := make([]byte, 32)

		core.GeneratePingTokens(expireTimestamp, &clientAddress, []net.UDPAddr{serverAddress}, pingKey, pingToken)

		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18 + 8 + 8 + 8 + 32)
			packet[0] = 20 // NEAR_PING_PACKET
			binary.LittleEndian.PutUint64(packet[16:], sequence)
			binary.LittleEndian.PutUint64(packet[16+1:], sessionId)
			binary.LittleEndian.PutUint64(packet[16+8+8:], expireTimestamp)
			copy(packet[16+8+8+8:16+8+8+8+32], pingToken)
			var magic [constants.MagicBytes]byte
			var fromAddressBuffer [32]byte
			var toAddressBuffer [32]byte
			fromAddress, fromPort := core.GetAddressData(&clientAddress, fromAddressBuffer[:])
			toAddress, toPort := core.GetAddressData(&serverAddress, toAddressBuffer[:])
			packetLength := len(packet)
			core.GenerateChonkle(packet[1:], magic[:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			core.GeneratePittle(packet[packetLength-2:], fromAddress[:], fromPort, toAddress[:], toPort, packetLength)
			conn.WriteToUDP(packet, &serverAddress)
			sequence++
		}

		time.Sleep(time.Second)
	}

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_NEAR_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_NEAR_PING_PACKET_RESPONDED_WITH_PONG", relay_stdout.String())

	if !receivedPong {
		panic("did not receive any pong packets")
	}
}

// todo: RELAY_COUNTER_RELAY_PONG_PACKET_WRONG_SIZE

// fmt.Printf("=======================================\n%s=============================================\n", relay_stdout)

type test_function func()

func main() {

	initCounterNames()

	allTests := []test_function{
		test_initialize_success,
		test_initialize_fail,
		test_relay_name_not_set,
		test_relay_public_address_not_set,
		test_relay_public_address_invalid,
		test_relay_internal_address_invalid,
		test_relay_public_key_not_set,
		test_relay_public_key_invalid,
		test_relay_private_key_not_set,
		test_relay_private_key_invalid,
		test_relay_keypair_invalid,
		test_relay_backend_public_key_not_set,
		test_relay_backend_public_key_invalid,
		test_relay_backend_public_key_mismatch,
		test_relay_backend_hostname_not_set,
		test_relay_cant_bind_to_port_zero,
		test_num_threads,
		test_relay_pings,
		test_cost_matrix,
		test_basic_packet_filter,
		test_advanced_packet_filter,
		test_clean_shutdown,
		test_unknown_packets,
		test_near_ping_packet_wrong_size,
		test_near_ping_packet_expired,
		test_near_ping_packet_did_not_verify,
		test_near_ping_packet_responded_with_pong,
	}

	var tests []test_function

	if len(os.Args) > 1 {
		funcName := os.Args[1]
		for _, test := range allTests {
			name := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
			name = name[len("main."):]
			if funcName == name {
				tests = append(tests, test)
				break
			}
		}
		if len(tests) == 0 {
			panic(fmt.Sprintf("could not find any test: '%s'", funcName))
		}
	} else {
		tests = allTests // No command line args, run all tests
	}

	go func() {
		time.Sleep(time.Duration(len(tests)*120) * time.Second)
		panic("tests took too long!")
	}()

	for i := range tests {
		tests[i]()
	}
}

var counterNames [constants.NumRelayCounters]string

var counterHash map[string]int

func initCounterNames() {
	// awk '/^#define RELAY_COUNTER_/ {print "    counterNames["$3"] = \""$2"\""}' ./relay/relay.cpp
	counterNames[0] = "RELAY_COUNTER_PACKETS_SENT"
	counterNames[1] = "RELAY_COUNTER_PACKETS_RECEIVED"
	counterNames[2] = "RELAY_COUNTER_BYTES_SENT"
	counterNames[3] = "RELAY_COUNTER_BYTES_RECEIVED"
	counterNames[4] = "RELAY_COUNTER_BASIC_PACKET_FILTER_DROPPED_PACKET"
	counterNames[5] = "RELAY_COUNTER_ADVANCED_PACKET_FILTER_DROPPED_PACKET"
	counterNames[6] = "RELAY_COUNTER_SESSION_CREATED"
	counterNames[7] = "RELAY_COUNTER_SESSION_CONTINUED"
	counterNames[8] = "RELAY_COUNTER_SESSION_DESTROYED"
	counterNames[9] = "RELAY_COUNTER_SESSION_EXPIRED"
	counterNames[10] = "RELAY_COUNTER_RELAY_PING_PACKET_SENT"
	counterNames[11] = "RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED"
	counterNames[12] = "RELAY_COUNTER_RELAY_PING_PACKET_DID_NOT_VERIFY"
	counterNames[13] = "RELAY_COUNTER_RELAY_PING_PACKET_EXPIRED"
	counterNames[14] = "RELAY_COUNTER_RELAY_PING_PACKET_WRONG_SIZE"
	counterNames[15] = "RELAY_COUNTER_RELAY_PONG_PACKET_SENT"
	counterNames[16] = "RELAY_COUNTER_RELAY_PONG_PACKET_RECEIVED"
	counterNames[17] = "RELAY_COUNTER_RELAY_PONG_PACKET_WRONG_SIZE"
	counterNames[20] = "RELAY_COUNTER_NEAR_PING_PACKET_RECEIVED"
	counterNames[21] = "RELAY_COUNTER_NEAR_PING_PACKET_WRONG_SIZE"
	counterNames[22] = "RELAY_COUNTER_NEAR_PING_PACKET_RESPONDED_WITH_PONG"
	counterNames[23] = "RELAY_COUNTER_NEAR_PING_PACKET_DID_NOT_VERIFY"
	counterNames[24] = "RELAY_COUNTER_NEAR_PING_PACKET_EXPIRED"
	counterNames[30] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED"
	counterNames[31] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_WRONG_SIZE"
	counterNames[32] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_COULD_NOT_READ_TOKEN"
	counterNames[33] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_TOKEN_EXPIRED"
	counterNames[34] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[35] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[40] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED"
	counterNames[41] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_WRONG_SIZE"
	counterNames[42] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[43] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[45] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED"
	counterNames[46] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[47] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[48] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[50] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED"
	counterNames[51] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_WRONG_SIZE"
	counterNames[52] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_READ_TOKEN"
	counterNames[53] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED"
	counterNames[55] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[56] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[60] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED"
	counterNames[61] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_WRONG_SIZE"
	counterNames[62] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[63] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED"
	counterNames[64] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[66] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[67] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[68] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[70] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED"
	counterNames[71] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_SMALL"
	counterNames[72] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_BIG"
	counterNames[73] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[74] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[76] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED"
	counterNames[77] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_VERIFY_HEADER"
	counterNames[78] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[79] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[80] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED"
	counterNames[81] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_SMALL"
	counterNames[82] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_BIG"
	counterNames[83] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[84] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[86] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED"
	counterNames[87] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_VERIFY_HEADER"
	counterNames[88] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[89] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[90] = "RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED"
	counterNames[91] = "RELAY_COUNTER_SESSION_PING_PACKET_WRONG_SIZE"
	counterNames[92] = "RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[93] = "RELAY_COUNTER_SESSION_PING_PACKET_SESSION_DOES_NOT_EXIST"
	counterNames[95] = "RELAY_COUNTER_SESSION_PING_PACKET_ALREADY_RECEIVED"
	counterNames[96] = "RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_VERIFY_HEADER"
	counterNames[97] = "RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[98] = "RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[100] = "RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED"
	counterNames[101] = "RELAY_COUNTER_SESSION_PONG_PACKET_WRONG_SIZE"
	counterNames[102] = "RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[103] = "RELAY_COUNTER_SESSION_PONG_PACKET_SESSION_DOES_NOT_EXIST"
	counterNames[105] = "RELAY_COUNTER_SESSION_PONG_PACKET_ALREADY_RECEIVED"
	counterNames[106] = "RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_VERIFY_HEADER"
	counterNames[107] = "RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[108] = "RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[110] = "RELAY_COUNTER_PACKETS_RECEIVED_BEFORE_INITIALIZE"
	counterNames[111] = "RELAY_COUNTER_UNKNOWN_PACKETS"
	counterNames[112] = "RELAY_COUNTER_PONGS_PROCESSED"

	counterHash = make(map[string]int)

	for i := 0; i < constants.NumRelayCounters; i++ {
		if counterNames[i] != "" {
			counterHash[counterNames[i]] = i
		}
	}
}

func getCounterIndex(name string) int {
	index, exists := counterHash[name]
	if !exists {
		panic(fmt.Sprintf("unknown counter: %s", name))
	}
	return index
}

func checkCounter(name string, stdout string) {
	index := getCounterIndex(name)
	if !strings.Contains(stdout, fmt.Sprintf("counter %d: ", index)) {
		panic(fmt.Sprintf("missing counter: %s", name))
	}
}
