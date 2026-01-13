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
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"syscall"
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
	omit_relay_backend_url            bool
	bind_to_port_zero                 bool
	print_counters                    bool
	disable_destroy                   bool
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

	if !config.omit_relay_backend_url {
		cmd.Env = append(cmd.Env, "RELAY_BACKEND_URL=http://127.0.0.1:30000")
	}

	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_FAKE_PACKET_LOSS_PERCENT=%f", config.fake_packet_loss_percent))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_FAKE_PACKET_LOSS_START_TIME=%f", config.fake_packet_loss_start_time))

	if config.bind_to_port_zero {
		cmd.Env = append(cmd.Env, "RELAY_PUBLIC_ADDRESS=127.0.0.1:0")
	}

	if config.print_counters {
		cmd.Env = append(cmd.Env, "RELAY_PRINT_COUNTERS=1")
	}

	if config.disable_destroy {
		cmd.Env = append(cmd.Env, "RELAY_DISABLE_DESTROY=1")
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

// =======================================================================================================================

func test_initialize_success() {

	fmt.Printf("test_initialize_success\n")

	backend_cmd, _ := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	time.Sleep(15 * time.Second)

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

// =======================================================================================================================

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

	if !strings.Contains(relay_stdout.String(), "error: relay update response is 400. the relay backend is down or the relay is misconfigured. check RELAY_BACKEND_PUBLIC_KEY") {
		panic("relay cannot talk to the relay backend unless it has the correct relay backend public key")
	}
}

func test_relay_backend_url_not_set() {

	fmt.Printf("test_relay_backend_url_not_set\n")

	config := RelayConfig{}
	config.omit_relay_backend_url = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "error: RELAY_BACKEND_URL not set") {
		panic("relay should not start without a relay backend url")
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

// =======================================================================================================================

func test_relay_pings() {

	fmt.Printf("test_relay_pings\n")

	backend_cmd, _ := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
		checkNoCounter("RELAY_COUNTER_RELAY_PONG_PACKET_WRONG_SIZE", relay_stdout[i])
		checkNoCounter("RELAY_COUNTER_RELAY_PONG_PACKET_UNKNOWN_RELAY", relay_stdout[i])
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

	buffer, err := io.ReadAll(response.Body)
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

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.print_counters = true

	relay_1_cmd, relay_1_stdout := relay("relay", 2000, config)
	relay_2_cmd, relay_2_stdout := relay("relay", 2001, config)
	relay_3_cmd, relay_3_stdout := relay("relay", 2002, config)

	time.Sleep(5 * time.Second)

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
		fmt.Printf("--------------------------------------------------\n")
		fmt.Printf("%s", backend_stdout.String())
		fmt.Printf("--------------------------------------------------\n")
		panic(fmt.Sprintf("cost matrix entries are invalid: %+v", costMatrix.Costs))
	}
}

// =======================================================================================================================

func test_basic_packet_filter() {

	fmt.Printf("test_basic_packet_filter\n")

	backend_cmd, _ := backend("DEFAULT")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.print_counters = true

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	conn, err := net.Dial("udp", "127.0.0.1:2000")
	if err != nil {
		panic("could not create udp socket")
	}

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, common.RandomInt(1, constants.MaxPacketBytes))
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
			packet := make([]byte, common.RandomInt(18, constants.MaxPacketBytes))
			common.RandomBytes(packet[:])
			var magic [8]byte
			var fromAddress [4]byte
			var toAddress [4]byte
			common.RandomBytes(magic[:])
			common.RandomBytes(fromAddress[:])
			common.RandomBytes(toAddress[:])
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

// =======================================================================================================================

func test_clean_shutdown() {

	fmt.Printf("test_clean_shutdown\n")

	backend_cmd, backend_stdout := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
		panic("relay crashed while shutting down")
	}

	if !strings.Contains(backend_stdout.String(), "relay is shutting down") {
		fmt.Printf("=============================================\n%s==============================================\n", backend_stdout.String())
		panic("missing relay is shutting down on backend")
	}
}

// =======================================================================================================================

func test_client_ping_packet_wrong_size() {

	fmt.Printf("test_client_ping_packet_wrong_size\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, common.RandomInt(18, constants.MaxPacketBytes))
			common.RandomBytes(packet[:])
			packet[0] = CLIENT_PING_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CLIENT_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_PING_PACKET_WRONG_SIZE", relay_stdout.String())
}

func test_client_ping_packet_expired() {

	fmt.Printf("test_client_ping_packet_expired\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+8+8+8+32)
			packet[0] = CLIENT_PING_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CLIENT_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_PING_PACKET_EXPIRED", relay_stdout.String())
}

func test_client_ping_packet_did_not_verify() {

	fmt.Printf("test_client_ping_packet_did_not_verify\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+8+8+8+32)
			packet[0] = CLIENT_PING_PACKET
			binary.LittleEndian.PutUint64(packet[18+8+8:], uint64(expireTimestamp))
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CLIENT_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_PING_PACKET_DID_NOT_VERIFY", relay_stdout.String())
}

func test_client_ping_packet_responded_with_pong() {

	fmt.Printf("test_client_ping_packet_responded_with_pong\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			receiveBuffer := make([]byte, constants.MaxPacketBytes)
			receivePacketBytes, from, err := conn.ReadFromUDP(receiveBuffer[:])
			if err != nil {
				break
			}
			if receivePacketBytes == 18+8+8 && receiveBuffer[0] == CLIENT_PONG_PACKET && from.String() == serverAddress.String() {
				receivedPong = true
				break
			}
		}
	}()

	for i := 0; i < 10; i++ {

		expireTimestamp := uint64(time.Now().Unix()) + 10

		pingToken := make([]byte, 32)

		clientAddressWithoutPort := clientAddress
		clientAddressWithoutPort.Port = 0

		core.GeneratePingToken(expireTimestamp, &clientAddressWithoutPort, &serverAddress, pingKey, pingToken)

		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+8+8+8+32)
			packet[0] = CLIENT_PING_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequence)
			binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
			binary.LittleEndian.PutUint64(packet[18+8+8:], expireTimestamp)
			copy(packet[18+8+8+8:18+8+8+8+32], pingToken)
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CLIENT_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_PING_PACKET_RESPONDED_WITH_PONG", relay_stdout.String())

	if !receivedPong {
		panic("did not receive any pong packets")
	}
}

// =======================================================================================================================

func test_server_ping_packet_wrong_size() {

	fmt.Printf("test_server_ping_packet_wrong_size\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, common.RandomInt(18, constants.MaxPacketBytes))
			common.RandomBytes(packet[:])
			packet[0] = SERVER_PING_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SERVER_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_PING_PACKET_WRONG_SIZE", relay_stdout.String())
}

func test_server_ping_packet_expired() {

	fmt.Printf("test_server_ping_packet_expired\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+8+8+32)
			packet[0] = SERVER_PING_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SERVER_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_PING_PACKET_EXPIRED", relay_stdout.String())
}

func test_server_ping_packet_did_not_verify() {

	fmt.Printf("test_server_ping_packet_did_not_verify\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+8+8+32)
			packet[0] = SERVER_PING_PACKET
			binary.LittleEndian.PutUint64(packet[18+8:], uint64(expireTimestamp))
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SERVER_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_PING_PACKET_DID_NOT_VERIFY", relay_stdout.String())
}

func test_server_ping_packet_responded_with_pong() {

	fmt.Printf("test_server_ping_packet_responded_with_pong\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	pingKey := make([]byte, 32)

	receivedPong := false

	go func() {
		for {
			receiveBuffer := make([]byte, constants.MaxPacketBytes)
			receivePacketBytes, from, err := conn.ReadFromUDP(receiveBuffer[:])
			if err != nil {
				break
			}
			if receivePacketBytes == 18+8 && receiveBuffer[0] == SERVER_PONG_PACKET && from.String() == serverAddress.String() {
				receivedPong = true
				break
			}
		}
	}()

	for i := 0; i < 10; i++ {

		expireTimestamp := uint64(time.Now().Unix()) + 10

		pingToken := make([]byte, 32)

		core.GeneratePingToken(expireTimestamp, &clientAddress, &serverAddress, pingKey, pingToken)

		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+8+8+32)
			packet[0] = SERVER_PING_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequence)
			binary.LittleEndian.PutUint64(packet[18+8:], expireTimestamp)
			copy(packet[18+8+8:18+8+8+32], pingToken)
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SERVER_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_PING_PACKET_RESPONDED_WITH_PONG", relay_stdout.String())

	if !receivedPong {
		panic("did not receive any pong packets")
	}
}

// =======================================================================================================================

func test_relay_pong_packet_wrong_size() {

	fmt.Printf("test_relay_pong_packet_wrong_size\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, common.RandomInt(18, constants.MaxPacketBytes))
			common.RandomBytes(packet[:])
			packet[0] = RELAY_PONG_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_RELAY_PONG_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_RELAY_PONG_PACKET_WRONG_SIZE", relay_stdout.String())
}

func test_relay_ping_packet_wrong_size() {

	fmt.Printf("test_relay_ping_packet_wrong_size\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, common.RandomInt(18, constants.MaxPacketBytes))
			common.RandomBytes(packet[:])
			packet[0] = RELAY_PING_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_RELAY_PING_PACKET_WRONG_SIZE", relay_stdout.String())
}

func test_relay_ping_packet_expired() {

	fmt.Printf("test_relay_ping_packet_expired\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+8+8+1+32)
			packet[0] = RELAY_PING_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_RELAY_PING_PACKET_EXPIRED", relay_stdout.String())
}

func test_relay_ping_packet_did_not_verify() {

	fmt.Printf("test_relay_ping_packet_did_not_verify\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+8+8+1+32)
			packet[0] = RELAY_PING_PACKET
			binary.LittleEndian.PutUint64(packet[18+8:], uint64(expireTimestamp))
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_RELAY_PING_PACKET_DID_NOT_VERIFY", relay_stdout.String())
}

// =======================================================================================================================

func test_route_request_packet_wrong_size() {

	fmt.Printf("test_route_request_packet_wrong_size\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, common.RandomInt(18, constants.MaxPacketBytes))
			common.RandomBytes(packet[:])
			packet[0] = ROUTE_REQUEST_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_WRONG_SIZE", relay_stdout.String())
}

func test_route_request_packet_could_not_decrypt_route_token() {

	fmt.Printf("test_route_request_packet_could_not_decrypt_route_token\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+111*2)
			common.RandomBytes(packet[:])
			packet[0] = ROUTE_REQUEST_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_COULD_NOT_DECRYPT_ROUTE_TOKEN", relay_stdout.String())
}

func test_route_request_packet_token_expired() {

	fmt.Printf("test_route_request_packet_token_expired\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+111*2)
			common.RandomBytes(packet[:])
			packet[0] = ROUTE_REQUEST_PACKET
			token := core.RouteToken{}
			token.NextAddress = net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 0}
			token.PrevAddress = net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 0}
			core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_TOKEN_EXPIRED", relay_stdout.String())
}

func test_route_request_packet_forward_to_next_hop() {

	fmt.Printf("test_route_request_packet_forward_to_next_hop\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	receivedRouteRequestPacket := false

	go func() {
		for {
			receiveBuffer := make([]byte, constants.MaxPacketBytes)
			receivePacketBytes, from, err := conn.ReadFromUDP(receiveBuffer[:])
			if err != nil {
				break
			}
			if receivePacketBytes == 18+111 && receiveBuffer[0] == ROUTE_REQUEST_PACKET && from.String() == serverAddress.String() {
				receivedRouteRequestPacket = true
				break
			}
		}
	}()

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+111*2)
			common.RandomBytes(packet[:])
			packet[0] = ROUTE_REQUEST_PACKET
			token := core.RouteToken{}
			token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
			token.NextAddress = clientAddress
			token.PrevAddress = clientAddress
			core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())

	if !receivedRouteRequestPacket {
		panic("did not receive forwarded route request packet")
	}
}

// =======================================================================================================================

func test_route_response_packet_wrong_size() {

	fmt.Printf("test_route_response_packet_wrong_size\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, common.RandomInt(18, constants.MaxPacketBytes))
			common.RandomBytes(packet[:])
			packet[0] = ROUTE_RESPONSE_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_WRONG_SIZE", relay_stdout.String())
}

func test_route_response_packet_could_not_find_session() {

	fmt.Printf("test_route_response_packet_could_not_find_session\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+25)
			packet[0] = ROUTE_RESPONSE_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION", relay_stdout.String())
}

func test_route_response_packet_already_received() {

	fmt.Printf("test_route_response_packet_already_received\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	// now send a bunch of route response packets with sequence number 0, they will trigger already received
	// (sequence number starts at zero...)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+25)
			packet[0] = ROUTE_RESPONSE_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED", relay_stdout.String())
}

func test_route_response_packet_header_did_not_verify() {

	fmt.Printf("test_route_response_packet_header_did_not_verify\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	time.Sleep(time.Second)

	// send a route response packet with sequence number > 0, so it passes already received test, but does not verify

	{
		packet := make([]byte, 18+25)
		packet[0] = ROUTE_RESPONSE_PACKET
		binary.LittleEndian.PutUint64(packet[18:], 1)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	time.Sleep(time.Second)

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY", relay_stdout.String())
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

func test_route_response_packet_forward_to_previous_hop() {

	fmt.Printf("test_route_response_packet_forward_to_previous_hop\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(0x1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.SessionId = sessionId
	token.SessionVersion = sessionVersion
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	copy(token.SessionPrivateKey[:], sessionKey)
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	time.Sleep(time.Second)

	// send a valid route response packet so it gets forwarded to previous hop (client address)

	{
		packet := make([]byte, 18+25)

		sequenceNumber := uint64(1)

		packet[0] = ROUTE_RESPONSE_PACKET
		binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
		binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
		packet[18+8+8] = sessionVersion

		tag := GenerateHeaderTag(ROUTE_RESPONSE_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
		copy(packet[18+8+8+1:], tag)

		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)

		packetLength := len(packet)

		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

		conn.WriteToUDP(packet, &serverAddress)
	}

	time.Sleep(time.Second)

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP", relay_stdout.String())
}

// =======================================================================================================================

func test_continue_request_packet_wrong_size() {

	fmt.Printf("test_continue_request_packet_wrong_size\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, common.RandomInt(18, constants.MaxPacketBytes))
			common.RandomBytes(packet[:])
			packet[0] = CONTINUE_REQUEST_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_WRONG_SIZE", relay_stdout.String())
}

func test_continue_request_packet_could_not_decrypt_continue_token() {

	fmt.Printf("test_continue_request_packet_could_not_decrypt_continue_token\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+57*2)
			common.RandomBytes(packet[:])
			packet[0] = CONTINUE_REQUEST_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_DECRYPT_CONTINUE_TOKEN", relay_stdout.String())
}

func test_continue_request_packet_token_expired() {

	fmt.Printf("test_continue_request_packet_token_expired\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+57*2)
			common.RandomBytes(packet[:])
			packet[0] = CONTINUE_REQUEST_PACKET
			token := core.ContinueToken{}
			core.WriteEncryptedContinueToken(&token, packet[18:], testSecretKey)
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED", relay_stdout.String())
}

func test_continue_request_packet_could_not_find_session() {

	fmt.Printf("test_continue_request_packet_could_not_find_session\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+57*2)
			common.RandomBytes(packet[:])
			packet[0] = CONTINUE_REQUEST_PACKET
			token := core.ContinueToken{}
			token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
			core.WriteEncryptedContinueToken(&token, packet[18:], testSecretKey)
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_FIND_SESSION", relay_stdout.String())
}

func test_continue_request_packet_forward_to_next_hop() {

	fmt.Printf("test_continue_request_packet_forward_to_next_hop\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// first send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now send continue request packets and listen to see that they get forwarded

	receivedContinueRequestPacket := false

	go func() {
		for {
			receiveBuffer := make([]byte, constants.MaxPacketBytes)
			receivePacketBytes, from, err := conn.ReadFromUDP(receiveBuffer[:])
			if err != nil {
				break
			}
			if receivePacketBytes == 18+57 && receiveBuffer[0] == CONTINUE_REQUEST_PACKET && from.String() == serverAddress.String() {
				receivedContinueRequestPacket = true
				break
			}
		}
	}()

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+57*2)
			common.RandomBytes(packet[:])
			packet[0] = CONTINUE_REQUEST_PACKET
			token := core.ContinueToken{}
			token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
			core.WriteEncryptedContinueToken(&token, packet[18:], testSecretKey)
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_CONTINUED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())

	if !receivedContinueRequestPacket {
		panic("did not receive forwarded continue request packet")
	}
}

// =======================================================================================================================

func test_continue_response_packet_wrong_size() {

	fmt.Printf("test_continue_response_packet_wrong_size\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, common.RandomInt(18, constants.MaxPacketBytes))
			common.RandomBytes(packet[:])
			packet[0] = CONTINUE_RESPONSE_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_WRONG_SIZE", relay_stdout.String())
}

func test_continue_response_packet_could_not_find_session() {

	fmt.Printf("test_continue_response_packet_could_not_find_session\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+25)
			packet[0] = CONTINUE_RESPONSE_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION", relay_stdout.String())
}

func test_continue_response_packet_already_received() {

	fmt.Printf("test_continue_response_packet_already_received\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	// now send a bunch of continue response packets with sequence number 0, they will trigger already received
	// (sequence number starts at zero...)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+25)
			packet[0] = CONTINUE_RESPONSE_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED", relay_stdout.String())
}

func test_continue_response_packet_header_did_not_verify() {

	fmt.Printf("test_continue_response_packet_header_did_not_verify\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	time.Sleep(time.Second)

	// send a continue response packet with sequence number > 0, so it passes already received test, but does not verify

	{
		packet := make([]byte, 18+25)
		packet[0] = CONTINUE_RESPONSE_PACKET
		binary.LittleEndian.PutUint64(packet[18:], 1)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	time.Sleep(time.Second)

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY", relay_stdout.String())
}

func test_continue_response_packet_forward_to_previous_hop() {

	fmt.Printf("test_continue_response_packet_forward_to_previous_hop\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.SessionId = sessionId
	token.SessionVersion = sessionVersion
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	copy(token.SessionPrivateKey[:], sessionKey)
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	time.Sleep(time.Second)

	// send a valid continue response packet so it gets forwarded to previous hop (client address)

	{
		packet := make([]byte, 18+25)

		sequenceNumber := uint64(1)

		packet[0] = CONTINUE_RESPONSE_PACKET
		binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
		binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
		packet[18+8+8] = sessionVersion

		tag := GenerateHeaderTag(CONTINUE_RESPONSE_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
		copy(packet[18+8+8+1:], tag)

		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)

		packetLength := len(packet)

		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

		conn.WriteToUDP(packet, &serverAddress)
	}

	time.Sleep(time.Second)

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP", relay_stdout.String())
}

// =======================================================================================================================

func test_client_to_server_packet_too_small() {

	fmt.Printf("test_client_to_server_packet_too_small\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	for i := 0; i < 30; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, common.RandomInt(18, 18+25-1))
			common.RandomBytes(packet[:])
			packet[0] = CLIENT_TO_SERVER_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_SMALL", relay_stdout.String())
}

func test_client_to_server_packet_too_big() {

	fmt.Printf("test_client_to_server_packet_too_big\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	for i := 0; i < 30; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, common.RandomInt(constants.MaxPacketBytes, 4095))
			common.RandomBytes(packet[:])
			packet[0] = CLIENT_TO_SERVER_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_BIG", relay_stdout.String())
}

func test_client_to_server_packet_could_not_find_session() {

	fmt.Printf("test_client_to_server_packet_could_not_find_session\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+25+256)
			packet[0] = CLIENT_TO_SERVER_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_FIND_SESSION", relay_stdout.String())
}

func test_client_to_server_packet_already_received() {

	fmt.Printf("test_client_to_server_packet_already_received\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	// now send a bunch of client to server packets with sequence number 0, they will trigger already received
	// (sequence number starts at zero...)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+25+256)
			packet[0] = CLIENT_TO_SERVER_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED", relay_stdout.String())
}

func test_client_to_server_packet_header_did_not_verify() {

	fmt.Printf("test_client_to_server_packet_header_did_not_verify\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	time.Sleep(time.Second)

	// send a client to server packet with sequence number > 0, so it passes already received test, but does not verify

	{
		packet := make([]byte, 18+25+256)
		packet[0] = CLIENT_TO_SERVER_PACKET
		binary.LittleEndian.PutUint64(packet[18:], 1)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	time.Sleep(time.Second)

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_HEADER_DID_NOT_VERIFY", relay_stdout.String())
}

func test_client_to_server_packet_forward_to_next_hop() {

	fmt.Printf("test_client_to_server_packet_forward_to_next_hop\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// first send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		copy(token.SessionPrivateKey[:], sessionKey)
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now send client to server packets and listen to see that they get forwarded

	receivedClientToServerPacket := false

	go func() {
		for {
			receiveBuffer := make([]byte, constants.MaxPacketBytes)
			receivePacketBytes, from, err := conn.ReadFromUDP(receiveBuffer[:])
			if err != nil {
				break
			}
			if receivePacketBytes == 18+25+256 && receiveBuffer[0] == CLIENT_TO_SERVER_PACKET && from.String() == serverAddress.String() {
				receivedClientToServerPacket = true
				break
			}
		}
	}()

	sequenceNumber := uint64(1)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {

			packet := make([]byte, 18+25+256)

			packet[0] = CLIENT_TO_SERVER_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
			binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
			packet[18+8+8] = sessionVersion

			tag := GenerateHeaderTag(CLIENT_TO_SERVER_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
			copy(packet[18+8+8+1:], tag)

			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)

			packetLength := len(packet)

			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

			conn.WriteToUDP(packet, &serverAddress)

			sequenceNumber++
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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())

	if !receivedClientToServerPacket {
		panic("did not receive forwarded client to server packet")
	}
}

// =======================================================================================================================

func test_server_to_client_packet_too_small() {

	fmt.Printf("test_server_to_client_packet_too_small\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	for i := 0; i < 30; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, common.RandomInt(18, 18+25-1))
			common.RandomBytes(packet[:])
			packet[0] = SERVER_TO_CLIENT_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_SMALL", relay_stdout.String())
}

func test_server_to_client_packet_too_big() {

	fmt.Printf("test_server_to_client_packet_too_big\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	for i := 0; i < 30; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, common.RandomInt(constants.MaxPacketBytes, 4095))
			common.RandomBytes(packet[:])
			packet[0] = SERVER_TO_CLIENT_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_BIG", relay_stdout.String())
}

func test_server_to_client_packet_could_not_find_session() {

	fmt.Printf("test_server_to_client_packet_could_not_find_session\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+25+256)
			packet[0] = SERVER_TO_CLIENT_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_FIND_SESSION", relay_stdout.String())
}

func test_server_to_client_packet_already_received() {

	fmt.Printf("test_server_to_client_packet_already_received\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	// now send a bunch of server to client packets with sequence number 0, they will trigger already received
	// (sequence number starts at zero...)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+25+256)
			packet[0] = SERVER_TO_CLIENT_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED", relay_stdout.String())
}

func test_server_to_client_packet_header_did_not_verify() {

	fmt.Printf("test_server_to_client_packet_header_did_not_verify\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	time.Sleep(time.Second)

	// send a server to client packet with sequence number > 0, so it passes already received test, but does not verify

	{
		packet := make([]byte, 18+25+256)
		packet[0] = SERVER_TO_CLIENT_PACKET
		binary.LittleEndian.PutUint64(packet[18:], 1)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	time.Sleep(time.Second)

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_HEADER_DID_NOT_VERIFY", relay_stdout.String())
}

func test_server_to_client_packet_forward_to_previous_hop() {

	fmt.Printf("test_server_to_client_packet_forward_to_previous_hop\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// first send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		copy(token.SessionPrivateKey[:], sessionKey)
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now send server to client packets and listen to see that they get forwarded

	receivedServerToClientPacket := false

	go func() {
		for {
			receiveBuffer := make([]byte, constants.MaxPacketBytes)
			receivePacketBytes, from, err := conn.ReadFromUDP(receiveBuffer[:])
			if err != nil {
				break
			}
			if receivePacketBytes == 18+25+256 && receiveBuffer[0] == SERVER_TO_CLIENT_PACKET && from.String() == serverAddress.String() {
				receivedServerToClientPacket = true
				break
			}
		}
	}()

	sequenceNumber := uint64(1)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {

			packet := make([]byte, 18+25+256)

			packet[0] = SERVER_TO_CLIENT_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
			binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
			packet[18+8+8] = sessionVersion

			tag := GenerateHeaderTag(SERVER_TO_CLIENT_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
			copy(packet[18+8+8+1:], tag)

			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)

			packetLength := len(packet)

			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

			conn.WriteToUDP(packet, &serverAddress)

			sequenceNumber++
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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP", relay_stdout.String())

	if !receivedServerToClientPacket {
		panic("did not receive forwarded server to client packet")
	}
}

// =======================================================================================================================

func test_session_ping_packet_wrong_size() {

	fmt.Printf("test_session_ping_packet_wrong_size\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, common.RandomInt(18, constants.MaxPacketBytes))
			common.RandomBytes(packet[:])
			packet[0] = SESSION_PING_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PING_PACKET_WRONG_SIZE", relay_stdout.String())
}

func test_session_ping_packet_could_not_find_session() {

	fmt.Printf("test_session_ping_packet_could_not_find_session\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+25+8)
			packet[0] = SESSION_PING_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_FIND_SESSION", relay_stdout.String())
}

func test_session_ping_packet_already_received() {

	fmt.Printf("test_session_ping_packet_already_received\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	// now send a bunch of session ping packets with sequence number 0, they will trigger already received
	// (sequence number starts at zero...)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+25+8)
			packet[0] = SESSION_PING_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PING_PACKET_ALREADY_RECEIVED", relay_stdout.String())
}

func test_session_ping_packet_header_did_not_verify() {

	fmt.Printf("test_session_ping_packet_header_did_not_verify\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	time.Sleep(time.Second)

	// send a session ping packet with sequence number > 0, so it passes already received test, but does not verify

	{
		packet := make([]byte, 18+25+8)
		packet[0] = SESSION_PING_PACKET
		binary.LittleEndian.PutUint64(packet[18:], 1)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	time.Sleep(time.Second)

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PING_PACKET_HEADER_DID_NOT_VERIFY", relay_stdout.String())
}

func test_session_ping_packet_forward_to_next_hop() {

	fmt.Printf("test_session_ping_packet_forward_to_next_hop\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// first send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		copy(token.SessionPrivateKey[:], sessionKey)
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now send session ping packets and listen to see that they get forwarded

	receivedSessionPingPacket := false

	go func() {
		for {
			receiveBuffer := make([]byte, constants.MaxPacketBytes)
			receivePacketBytes, from, err := conn.ReadFromUDP(receiveBuffer[:])
			if err != nil {
				break
			}
			if receivePacketBytes == 18+25+8 && receiveBuffer[0] == SESSION_PING_PACKET && from.String() == serverAddress.String() {
				receivedSessionPingPacket = true
				break
			}
		}
	}()

	sequenceNumber := uint64(1)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {

			packet := make([]byte, 18+25+8)

			packet[0] = SESSION_PING_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
			binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
			packet[18+8+8] = sessionVersion

			tag := GenerateHeaderTag(SESSION_PING_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
			copy(packet[18+8+8+1:], tag)

			binary.LittleEndian.PutUint64(packet[18+25:], sequenceNumber)

			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)

			packetLength := len(packet)

			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())

	if !receivedSessionPingPacket {
		panic("did not receive forwarded session ping packet")
	}
}

// =======================================================================================================================

func test_session_pong_packet_wrong_size() {

	fmt.Printf("test_session_pong_packet_wrong_size\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, common.RandomInt(18, constants.MaxPacketBytes))
			common.RandomBytes(packet[:])
			packet[0] = SESSION_PONG_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PONG_PACKET_WRONG_SIZE", relay_stdout.String())
}

func test_session_pong_packet_could_not_find_session() {

	fmt.Printf("test_session_pong_packet_could_not_find_session\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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
			packet := make([]byte, 18+25+8)
			packet[0] = SESSION_PONG_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_FIND_SESSION", relay_stdout.String())
}

func test_session_pong_packet_already_received() {

	fmt.Printf("test_session_pong_packet_already_received\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	// send a route request packet to create a session on the relay

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	// now send a bunch of session ping packets with sequence number 0, they will trigger already received
	// (sequence number starts at zero...)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {
			packet := make([]byte, 18+25+8)
			packet[0] = SESSION_PONG_PACKET
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PONG_PACKET_ALREADY_RECEIVED", relay_stdout.String())
}

func test_session_pong_packet_header_did_not_verify() {

	fmt.Printf("test_session_pong_packet_header_did_not_verify\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// send a route request packet to create a session on the relay

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	time.Sleep(time.Second)

	// send a session ping packet with sequence number > 0, so it passes already received test, but does not verify

	{
		packet := make([]byte, 18+25+8)
		packet[0] = SESSION_PONG_PACKET
		binary.LittleEndian.PutUint64(packet[18:], 1)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	time.Sleep(time.Second)

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PONG_PACKET_HEADER_DID_NOT_VERIFY", relay_stdout.String())
}

func test_session_pong_packet_forward_to_previous_hop() {

	fmt.Printf("test_session_pong_packet_forward_to_previous_hop\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// first send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		copy(token.SessionPrivateKey[:], sessionKey)
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now send session pong packets and listen to see that they get forwarded

	receivedSessionPongPacket := false

	go func() {
		for {
			receiveBuffer := make([]byte, constants.MaxPacketBytes)
			receivePacketBytes, from, err := conn.ReadFromUDP(receiveBuffer[:])
			if err != nil {
				break
			}
			if receivePacketBytes == 18+25+8 && receiveBuffer[0] == SESSION_PONG_PACKET && from.String() == serverAddress.String() {
				receivedSessionPongPacket = true
				break
			}
		}
	}()

	sequenceNumber := uint64(1)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {

			packet := make([]byte, 18+25+8)

			packet[0] = SESSION_PONG_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
			binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
			packet[18+8+8] = sessionVersion

			tag := GenerateHeaderTag(SESSION_PONG_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
			copy(packet[18+8+8+1:], tag)

			binary.LittleEndian.PutUint64(packet[18+25:], sequenceNumber)

			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)

			packetLength := len(packet)

			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

			conn.WriteToUDP(packet, &serverAddress)

			sequenceNumber++
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

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP", relay_stdout.String())

	if !receivedSessionPongPacket {
		panic("did not receive forwarded session pong packet")
	}
}

// =======================================================================================================================

func test_session_expired_route_response_packet() {

	fmt.Printf("test_session_expired_route_response_packet\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.print_counters = true
	config.disable_destroy = true

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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.ExpireTimestamp = uint64(time.Now().Unix())
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now wait until the session should expire

	time.Sleep(10 * time.Second)

	// now throw a bunch of packets at the relay, and verify that the session is expired

	sequenceNumber := uint64(1)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {

			packet := make([]byte, 18+25)

			packet[0] = ROUTE_RESPONSE_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
			binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
			packet[18+8+8] = sessionVersion

			tag := GenerateHeaderTag(ROUTE_RESPONSE_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
			copy(packet[18+8+8+1:], tag)

			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)

			packetLength := len(packet)

			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

			conn.WriteToUDP(packet, &serverAddress)

			sequenceNumber++
		}

		time.Sleep(time.Second)
	}

	// verify everything is OK

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_SESSION_EXPIRED", relay_stdout.String())
	checkNoCounter("RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP", relay_stdout.String())
}

func test_session_expired_continue_request_packet() {

	fmt.Printf("test_session_expired_continue_request_packet\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.print_counters = true
	config.disable_destroy = true

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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.ExpireTimestamp = uint64(time.Now().Unix())
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now wait until the session should expire

	time.Sleep(10 * time.Second)

	// now throw a bunch of packets at the relay, and verify that the session is expired

	for i := 0; i < 10; i++ {
		for j := 0; j < 1; j++ {
			packet := make([]byte, 18+57*2)
			common.RandomBytes(packet[:])
			packet[0] = CONTINUE_REQUEST_PACKET
			token := core.ContinueToken{}
			token.SessionId = sessionId
			token.SessionVersion = sessionVersion
			token.ExpireTimestamp = uint64(time.Now().Unix()) + 1000
			core.WriteEncryptedContinueToken(&token, packet[18:], testSecretKey)
			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)
			packetLength := len(packet)
			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
			conn.WriteToUDP(packet, &serverAddress)
		}
		time.Sleep(time.Second)
	}

	// verify everything is OK

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_SESSION_EXPIRED", relay_stdout.String())
	checkNoCounter("RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
}

func test_session_expired_continue_response_packet() {

	fmt.Printf("test_session_expired_continue_response_packet\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.print_counters = true
	config.disable_destroy = true

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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.ExpireTimestamp = uint64(time.Now().Unix())
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now wait until the session should expire

	time.Sleep(10 * time.Second)

	// now throw a bunch of packets at the relay, and verify that the session is expired

	sequenceNumber := uint64(1)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {

			packet := make([]byte, 18+25)

			packet[0] = CONTINUE_RESPONSE_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
			binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
			packet[18+8+8] = sessionVersion

			tag := GenerateHeaderTag(CONTINUE_RESPONSE_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
			copy(packet[18+8+8+1:], tag)

			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)

			packetLength := len(packet)

			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

			conn.WriteToUDP(packet, &serverAddress)

			sequenceNumber++
		}

		time.Sleep(time.Second)
	}

	// verify everything is OK

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_SESSION_EXPIRED", relay_stdout.String())
	checkNoCounter("RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP", relay_stdout.String())
}

func test_session_expired_client_to_server_packet() {

	fmt.Printf("test_session_expired_client_to_server_packet\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.print_counters = true
	config.disable_destroy = true

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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.ExpireTimestamp = uint64(time.Now().Unix())
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now wait until the session should expire

	time.Sleep(10 * time.Second)

	// now throw a bunch of packets at the relay, and verify that the session is expired

	sequenceNumber := uint64(1)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {

			packet := make([]byte, 18+25+100)

			packet[0] = CLIENT_TO_SERVER_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
			binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
			packet[18+8+8] = sessionVersion

			tag := GenerateHeaderTag(CLIENT_TO_SERVER_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
			copy(packet[18+8+8+1:], tag)

			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)

			packetLength := len(packet)

			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

			conn.WriteToUDP(packet, &serverAddress)

			sequenceNumber++
		}

		time.Sleep(time.Second)
	}

	// verify everything is OK

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_SESSION_EXPIRED", relay_stdout.String())
	checkNoCounter("RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
}

func test_session_expired_server_to_client_packet() {

	fmt.Printf("test_session_expired_server_to_client_packet\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.print_counters = true
	config.disable_destroy = true

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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.ExpireTimestamp = uint64(time.Now().Unix())
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now wait until the session should expire

	time.Sleep(10 * time.Second)

	// now throw a bunch of packets at the relay, and verify that the session is expired

	sequenceNumber := uint64(1)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {

			packet := make([]byte, 18+25+100)

			packet[0] = SERVER_TO_CLIENT_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
			binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
			packet[18+8+8] = sessionVersion

			tag := GenerateHeaderTag(SERVER_TO_CLIENT_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
			copy(packet[18+8+8+1:], tag)

			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)

			packetLength := len(packet)

			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

			conn.WriteToUDP(packet, &serverAddress)

			sequenceNumber++
		}

		time.Sleep(time.Second)
	}

	// verify everything is OK

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_SESSION_EXPIRED", relay_stdout.String())
	checkNoCounter("RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP", relay_stdout.String())
}

func test_session_expired_session_ping_packet() {

	fmt.Printf("test_session_expired_session_ping_packet\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.print_counters = true
	config.disable_destroy = true

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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.ExpireTimestamp = uint64(time.Now().Unix())
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now wait until the session should expire

	time.Sleep(10 * time.Second)

	// now throw a bunch of packets at the relay, and verify that the session is expired

	sequenceNumber := uint64(1)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {

			packet := make([]byte, 18+25+8)

			packet[0] = SESSION_PING_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
			binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
			packet[18+8+8] = sessionVersion

			tag := GenerateHeaderTag(SESSION_PING_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
			copy(packet[18+8+8+1:], tag)

			binary.LittleEndian.PutUint64(packet[18+25:], sequenceNumber)

			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)

			packetLength := len(packet)

			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

			conn.WriteToUDP(packet, &serverAddress)

			sequenceNumber++
		}

		time.Sleep(time.Second)
	}

	// verify everything is OK

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PING_PACKET_SESSION_EXPIRED", relay_stdout.String())
	checkNoCounter("RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP", relay_stdout.String())
}

func test_session_expired_session_pong_packet() {

	fmt.Printf("test_session_expired_session_pong_packet\n")

	backend_cmd, _ := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}
	config.print_counters = true
	config.disable_destroy = true

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

	sessionId := uint64(0x12345)
	sessionVersion := uint8(1)

	sessionKey := make([]byte, crypto.Box_PrivateKeySize)
	common.RandomBytes(sessionKey)

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	// send a route request packet to create the session
	{
		packet := make([]byte, 18+111*2)
		common.RandomBytes(packet[:])
		packet[0] = ROUTE_REQUEST_PACKET
		token := core.RouteToken{}
		token.SessionId = sessionId
		token.SessionVersion = sessionVersion
		token.ExpireTimestamp = uint64(time.Now().Unix())
		token.NextAddress = clientAddress
		token.PrevAddress = clientAddress
		core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
		var magic [constants.MagicBytes]byte
		fromAddress := core.GetAddressData(&clientAddress)
		toAddress := core.GetAddressData(&serverAddress)
		packetLength := len(packet)
		core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
		core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
		conn.WriteToUDP(packet, &serverAddress)
	}

	// now wait until the session should expire

	time.Sleep(10 * time.Second)

	// now throw a bunch of packets at the relay, and verify that the session is expired

	sequenceNumber := uint64(1)

	for i := 0; i < 10; i++ {
		for j := 0; j < 1000; j++ {

			packet := make([]byte, 18+25+8)

			packet[0] = SESSION_PONG_PACKET
			binary.LittleEndian.PutUint64(packet[18:], sequenceNumber)
			binary.LittleEndian.PutUint64(packet[18+8:], sessionId)
			packet[18+8+8] = sessionVersion

			tag := GenerateHeaderTag(SESSION_PONG_PACKET, sequenceNumber, sessionId, sessionVersion, sessionKey)
			copy(packet[18+8+8+1:], tag)

			var magic [constants.MagicBytes]byte
			fromAddress := core.GetAddressData(&clientAddress)
			toAddress := core.GetAddressData(&serverAddress)

			packetLength := len(packet)

			core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)

			core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)

			conn.WriteToUDP(packet, &serverAddress)

			sequenceNumber++
		}

		time.Sleep(time.Second)
	}

	// verify everything is OK

	conn.Close()

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	checkCounter("RELAY_COUNTER_SESSION_CREATED", relay_stdout.String())
	checkCounter("RELAY_COUNTER_SESSION_PONG_PACKET_SESSION_EXPIRED", relay_stdout.String())
	checkNoCounter("RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP", relay_stdout.String())
}

// =======================================================================================================================

func test_relay_backend_stats() {

	fmt.Printf("test_relay_backend_stats\n")

	backend_cmd, backend_stdout := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	time.Sleep(5 * time.Second)

	time.Sleep(5 * time.Second)

	// send a route request packet to create a session on the relay

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind socket")
	}

	conn := lp.(*net.UDPConn)

	clientPort := conn.LocalAddr().(*net.UDPAddr).Port

	clientAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", clientPort))

	serverAddress := core.ParseAddress("127.0.0.1:2000")

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	token.EnvelopeKbpsUp = 512
	token.EnvelopeKbpsDown = 1024
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	// wait and make sure we see stats in the backend

	time.Sleep(5 * time.Second)

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		fmt.Printf("=====================================================\n%s======================================================\n", relay_stdout.String())
		panic("could not initialize relay")
	}

	if !strings.Contains(backend_stdout.String(), "session count = 1") {
		fmt.Printf("=====================================================\n%s======================================================\n", backend_stdout.String())
		panic("missing session count")
	}

	if !strings.Contains(backend_stdout.String(), "envelope bandwidth up kbps = 512") {
		fmt.Printf("=====================================================\n%s======================================================\n", backend_stdout.String())
		panic("missing envelope bandwidth up kbps")
	}

	if !strings.Contains(backend_stdout.String(), "envelope bandwidth down kbps = 1024") {
		fmt.Printf("=====================================================\n%s======================================================\n", backend_stdout.String())
		panic("missing envelope bandwidth down kbps")
	}
}

func test_relay_backend_counters() {

	fmt.Printf("test_relay_backend_counters\n")

	backend_cmd, backend_stdout := backend("ZERO_MAGIC")

	time.Sleep(time.Second)

	config := RelayConfig{}

	relay_cmd, relay_stdout := relay("relay", 2000, config)

	time.Sleep(5 * time.Second)

	// send a route request packet to create a session on the relay

	lc := net.ListenConfig{}

	lp, err := lc.ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	if err != nil {
		panic("could not bind socket")
	}

	conn := lp.(*net.UDPConn)

	clientPort := conn.LocalAddr().(*net.UDPAddr).Port

	clientAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", clientPort))

	serverAddress := core.ParseAddress("127.0.0.1:2000")

	testRelayPublicKey := Base64String(TestRelayPublicKey)
	testRelayPrivateKey := Base64String(TestRelayPrivateKey)
	testRelayBackendPublicKey := Base64String(TestRelayBackendPublicKey)

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(testRelayPublicKey, testRelayPrivateKey, testRelayBackendPublicKey)

	packet := make([]byte, 18+111*2)
	packet[0] = ROUTE_REQUEST_PACKET
	token := core.RouteToken{}
	token.ExpireTimestamp = uint64(time.Now().Unix()) + 15
	token.NextAddress = clientAddress
	token.PrevAddress = clientAddress
	core.WriteEncryptedRouteToken(&token, packet[18:], testSecretKey)
	var magic [constants.MagicBytes]byte
	fromAddress := core.GetAddressData(&clientAddress)
	toAddress := core.GetAddressData(&serverAddress)
	packetLength := len(packet)
	core.GeneratePittle(packet[1:3], fromAddress[:], toAddress[:], packetLength)
	core.GenerateChonkle(packet[3:18], magic[:], fromAddress[:], toAddress[:], packetLength)
	conn.WriteToUDP(packet, &serverAddress)

	// verify that we see the session created counter on the relay backned

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
	}

	backend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_cmd.Wait()

	if !strings.Contains(relay_stdout.String(), "Relay initialized") {
		panic("could not initialize relay")
	}

	if !strings.Contains(backend_stdout.String(), "counter 6: 1") {
		fmt.Printf("=====================================================\n%s======================================================\n", backend_stdout.String())
		panic("missing session created counter on backend")
	}
}

// =======================================================================================================================

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
		test_relay_backend_public_key_not_set,
		test_relay_backend_public_key_invalid,
		test_relay_backend_public_key_mismatch,
		test_relay_backend_url_not_set,

		test_relay_cant_bind_to_port_zero,

		test_relay_pings,

		test_cost_matrix,

		test_basic_packet_filter,
		test_advanced_packet_filter,

		test_clean_shutdown,

		test_client_ping_packet_wrong_size,
		test_client_ping_packet_expired,
		test_client_ping_packet_did_not_verify,
		test_client_ping_packet_responded_with_pong,

		test_server_ping_packet_wrong_size,
		test_server_ping_packet_expired,
		test_server_ping_packet_did_not_verify,
		test_server_ping_packet_responded_with_pong,

		test_relay_pong_packet_wrong_size,

		test_relay_ping_packet_wrong_size,
		test_relay_ping_packet_expired,
		test_relay_ping_packet_did_not_verify,

		test_route_request_packet_wrong_size,
		test_route_request_packet_could_not_decrypt_route_token,
		test_route_request_packet_token_expired,
		test_route_request_packet_forward_to_next_hop,

		test_route_response_packet_wrong_size,
		test_route_response_packet_could_not_find_session,
		test_route_response_packet_already_received,
		test_route_response_packet_header_did_not_verify,
		test_route_response_packet_forward_to_previous_hop,

		test_continue_request_packet_wrong_size,
		test_continue_request_packet_could_not_decrypt_continue_token,
		test_continue_request_packet_token_expired,
		test_continue_request_packet_could_not_find_session,
		test_continue_request_packet_forward_to_next_hop,

		test_continue_response_packet_wrong_size,
		test_continue_response_packet_could_not_find_session,
		test_continue_response_packet_already_received,
		test_continue_response_packet_header_did_not_verify,
		test_continue_response_packet_forward_to_previous_hop,

		test_client_to_server_packet_too_small,
		test_client_to_server_packet_too_big,
		test_client_to_server_packet_could_not_find_session,
		test_client_to_server_packet_already_received,
		test_client_to_server_packet_header_did_not_verify,
		test_client_to_server_packet_forward_to_next_hop,

		test_server_to_client_packet_too_small,
		test_server_to_client_packet_too_big,
		test_server_to_client_packet_could_not_find_session,
		test_server_to_client_packet_already_received,
		test_server_to_client_packet_header_did_not_verify,
		test_server_to_client_packet_forward_to_previous_hop,

		test_session_ping_packet_wrong_size,
		test_session_ping_packet_could_not_find_session,
		test_session_ping_packet_already_received,
		test_session_ping_packet_header_did_not_verify,
		test_session_ping_packet_forward_to_next_hop,

		test_session_pong_packet_wrong_size,
		test_session_pong_packet_could_not_find_session,
		test_session_pong_packet_already_received,
		test_session_pong_packet_header_did_not_verify,
		test_session_pong_packet_forward_to_previous_hop,

		test_session_expired_route_response_packet,
		test_session_expired_continue_request_packet,
		test_session_expired_continue_response_packet,
		test_session_expired_client_to_server_packet,
		test_session_expired_server_to_client_packet,
		test_session_expired_session_ping_packet,
		test_session_expired_session_pong_packet,

		test_relay_backend_stats,

		test_relay_backend_counters,
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
	// awk '/^#define RELAY_COUNTER_/ {print "    counterNames["$3"] = \""$2"\""}' ./relay/reference/reference_relay.cpp
	counterNames[0] = "RELAY_COUNTER_PACKETS_SENT"
	counterNames[1] = "RELAY_COUNTER_PACKETS_RECEIVED"
	counterNames[2] = "RELAY_COUNTER_BYTES_SENT"
	counterNames[3] = "RELAY_COUNTER_BYTES_RECEIVED"
	counterNames[4] = "RELAY_COUNTER_BASIC_PACKET_FILTER_DROPPED_PACKET"
	counterNames[5] = "RELAY_COUNTER_ADVANCED_PACKET_FILTER_DROPPED_PACKET"
	counterNames[6] = "RELAY_COUNTER_SESSION_CREATED"
	counterNames[7] = "RELAY_COUNTER_SESSION_CONTINUED"
	counterNames[8] = "RELAY_COUNTER_SESSION_DESTROYED"
	counterNames[10] = "RELAY_COUNTER_RELAY_PING_PACKET_SENT"
	counterNames[11] = "RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED"
	counterNames[12] = "RELAY_COUNTER_RELAY_PING_PACKET_DID_NOT_VERIFY"
	counterNames[13] = "RELAY_COUNTER_RELAY_PING_PACKET_EXPIRED"
	counterNames[14] = "RELAY_COUNTER_RELAY_PING_PACKET_WRONG_SIZE"
	counterNames[15] = "RELAY_COUNTER_RELAY_PING_PACKET_UNKNOWN_RELAY"
	counterNames[15] = "RELAY_COUNTER_RELAY_PONG_PACKET_SENT"
	counterNames[16] = "RELAY_COUNTER_RELAY_PONG_PACKET_RECEIVED"
	counterNames[17] = "RELAY_COUNTER_RELAY_PONG_PACKET_WRONG_SIZE"
	counterNames[18] = "RELAY_COUNTER_RELAY_PONG_PACKET_UNKNOWN_RELAY"
	counterNames[20] = "RELAY_COUNTER_CLIENT_PING_PACKET_RECEIVED"
	counterNames[21] = "RELAY_COUNTER_CLIENT_PING_PACKET_WRONG_SIZE"
	counterNames[22] = "RELAY_COUNTER_CLIENT_PING_PACKET_RESPONDED_WITH_PONG"
	counterNames[23] = "RELAY_COUNTER_CLIENT_PING_PACKET_DID_NOT_VERIFY"
	counterNames[24] = "RELAY_COUNTER_CLIENT_PING_PACKET_EXPIRED"
	counterNames[30] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED"
	counterNames[31] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_WRONG_SIZE"
	counterNames[32] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_COULD_NOT_DECRYPT_ROUTE_TOKEN"
	counterNames[33] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_TOKEN_EXPIRED"
	counterNames[34] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP"
	counterNames[40] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED"
	counterNames[41] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_WRONG_SIZE"
	counterNames[42] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[43] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_SESSION_EXPIRED"
	counterNames[44] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED"
	counterNames[45] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[46] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP"
	counterNames[50] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED"
	counterNames[51] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_WRONG_SIZE"
	counterNames[52] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_DECRYPT_CONTINUE_TOKEN"
	counterNames[53] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED"
	counterNames[54] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[55] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_SESSION_EXPIRED"
	counterNames[56] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP"
	counterNames[60] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED"
	counterNames[61] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_WRONG_SIZE"
	counterNames[62] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED"
	counterNames[63] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[64] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_SESSION_EXPIRED"
	counterNames[65] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[66] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP"
	counterNames[70] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED"
	counterNames[71] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_SMALL"
	counterNames[72] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_BIG"
	counterNames[73] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[74] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_SESSION_EXPIRED"
	counterNames[75] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED"
	counterNames[76] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[77] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP"
	counterNames[80] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED"
	counterNames[81] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_SMALL"
	counterNames[82] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_BIG"
	counterNames[83] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[84] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_SESSION_EXPIRED"
	counterNames[85] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED"
	counterNames[86] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[87] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP"
	counterNames[90] = "RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED"
	counterNames[91] = "RELAY_COUNTER_SESSION_PING_PACKET_WRONG_SIZE"
	counterNames[92] = "RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[93] = "RELAY_COUNTER_SESSION_PING_PACKET_SESSION_EXPIRED"
	counterNames[94] = "RELAY_COUNTER_SESSION_PING_PACKET_ALREADY_RECEIVED"
	counterNames[95] = "RELAY_COUNTER_SESSION_PING_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[96] = "RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP"
	counterNames[100] = "RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED"
	counterNames[101] = "RELAY_COUNTER_SESSION_PONG_PACKET_WRONG_SIZE"
	counterNames[102] = "RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[103] = "RELAY_COUNTER_SESSION_PONG_PACKET_SESSION_EXPIRED"
	counterNames[104] = "RELAY_COUNTER_SESSION_PONG_PACKET_ALREADY_RECEIVED"
	counterNames[105] = "RELAY_COUNTER_SESSION_PONG_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[106] = "RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP"
	counterNames[110] = "RELAY_COUNTER_SERVER_PING_PACKET_RECEIVED"
	counterNames[111] = "RELAY_COUNTER_SERVER_PING_PACKET_WRONG_SIZE"
	counterNames[112] = "RELAY_COUNTER_SERVER_PING_PACKET_RESPONDED_WITH_PONG"
	counterNames[113] = "RELAY_COUNTER_SERVER_PING_PACKET_DID_NOT_VERIFY"
	counterNames[114] = "RELAY_COUNTER_SERVER_PING_PACKET_EXPIRED"
	counterNames[120] = "RELAY_COUNTER_PACKET_TOO_LARGE"
	counterNames[121] = "RELAY_COUNTER_PACKET_TOO_SMALL"
	counterNames[122] = "RELAY_COUNTER_DROP_FRAGMENT"
	counterNames[123] = "RELAY_COUNTER_DROP_LARGE_IP_HEADER"
	counterNames[124] = "RELAY_COUNTER_REDIRECT_NOT_IN_WHITELIST"
	counterNames[125] = "RELAY_COUNTER_DROPPED_PACKETS"
	counterNames[126] = "RELAY_COUNTER_DROPPED_BYTES"
	counterNames[127] = "RELAY_COUNTER_NOT_IN_WHITELIST"
	counterNames[128] = "RELAY_COUNTER_WHITELIST_ENTRY_EXPIRED"
	counterNames[130] = "RELAY_COUNTER_SESSIONS"
	counterNames[131] = "RELAY_COUNTER_ENVELOPE_KBPS_UP"
	counterNames[132] = "RELAY_COUNTER_ENVELOPE_KBPS_DOWN"

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
		fmt.Printf("=======================================\n%s=============================================\n", stdout)
		panic(fmt.Sprintf("missing counter: %s", name))
	}
}

func checkNoCounter(name string, stdout string) {
	index := getCounterIndex(name)
	if strings.Contains(stdout, fmt.Sprintf("counter %d: ", index)) {
		fmt.Printf("=======================================\n%s=============================================\n", stdout)
		panic(fmt.Sprintf("unexpected counter: %s", name))
	}
}
