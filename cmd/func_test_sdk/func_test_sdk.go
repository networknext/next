/*
   Network Next. Copyright 2017 - 2025 Network Next, Inc.
   Licensed under the Network Next Source Available License 1.0
*/

package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

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
const TestServerBackendPublicKey = "1wXeogqOEL/UuMnHy3lwpdkdklcg4IktO/39mJiYfgc="
const TestServerBackendPrivateKey = "peZ17P29VgtnOiEv5wwNPDDo9lWweFV7dBVac0KoaXHXBd6iCo4Qv9S4ycfLeXCl2R2SVyDgiS07/f2YmJh+Bw=="
const TestBuyerPublicKey = "AzcqXbdP3Txq3rHIjRBS4BfG7OoKV9PAZfB0rY7a+ArdizBzFAd2vQ=="
const TestBuyerPrivateKey = "AzcqXbdP3TwX+9o9VfR7RcX2cq34UPdEsR2ztUnwxlTb/R49EiV5a2resciNEFLgF8bs6gpX08Bl8HStjtr4Ct2LMHMUB3a9"

const (
	relayBin   = "./relay-debug"
	backendBin = "./func_backend"
	clientBin  = "./func_client"
	serverBin  = "./func_server"
)

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

type RelayConfig struct {
	fake_packet_loss_percent    float32
	fake_packet_loss_start_time float32
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

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_NAME=%s", name))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_PUBLIC_ADDRESS=127.0.0.1:%d", port))
	cmd.Env = append(cmd.Env, "RELAY_BACKEND_URL=http://127.0.0.1:30000")
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_PUBLIC_KEY=%s", TestRelayPublicKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_PRIVATE_KEY=%s", TestRelayPrivateKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_BACKEND_PUBLIC_KEY=%s", TestRelayBackendPublicKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_FAKE_PACKET_LOSS_PERCENT=%f", config.fake_packet_loss_percent))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_FAKE_PACKET_LOSS_START_TIME=%f", config.fake_packet_loss_start_time))

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Start()

	return cmd, &output
}

type ClientConfig struct {
	duration                  int
	buyer_public_key          string
	disable_network_next      bool
	packet_loss               bool
	fake_direct_packet_loss   float32
	fake_direct_rtt           float32
	fake_next_packet_loss     float32
	fake_next_rtt             float32
	connect_time              float64
	connect_address           string
	stop_sending_packets_time float64
	fallback_to_direct_time   float64
	high_bandwidth            bool
	report_session            bool
	big_packets               bool
}

func client(config *ClientConfig) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {

	cmd := exec.Command(clientBin)
	if cmd == nil {
		panic("could not create client!\n")
		return nil, nil, nil
	}

	cmd.Env = os.Environ()

	cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_SERVER_BACKEND_PUBLIC_KEY=%s", TestServerBackendPublicKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_RELAY_BACKEND_PUBLIC_KEY=%s", TestRelayBackendPublicKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_BUYER_PUBLIC_KEY=%s", TestBuyerPublicKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_BUYER_PRIVATE_KEY=%s", TestBuyerPrivateKey))

	if config.duration != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("CLIENT_DURATION=%d", config.duration))
	}

	if config.buyer_public_key != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_BUYER_PUBLIC_KEY=%s", config.buyer_public_key))
	}

	if config.disable_network_next {
		cmd.Env = append(cmd.Env, "CLIENT_DISABLE_NETWORK_NEXT=1")
	}

	if config.fake_direct_packet_loss > 0.0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("CLIENT_FAKE_DIRECT_PACKET_LOSS=%f", config.fake_direct_packet_loss))
	}

	if config.fake_direct_rtt > 0.0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("CLIENT_FAKE_DIRECT_RTT=%f", config.fake_direct_rtt))
	}

	if config.fake_next_packet_loss > 0.0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("CLIENT_FAKE_NEXT_PACKET_LOSS=%f", config.fake_next_packet_loss))
	}

	if config.fake_next_rtt > 0.0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("CLIENT_FAKE_NEXT_RTT=%f", config.fake_next_rtt))
	}

	if config.connect_time > 0.0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("CLIENT_CONNECT_TIME=%f", config.connect_time))
		cmd.Env = append(cmd.Env, fmt.Sprintf("CLIENT_CONNECT_ADDRESS=%s", config.connect_address))
	}

	if config.packet_loss {
		cmd.Env = append(cmd.Env, "CLIENT_PACKET_LOSS=1")
	}

	if config.stop_sending_packets_time > 0.0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("CLIENT_STOP_SENDING_PACKETS_TIME=%f", config.stop_sending_packets_time))
	}

	if config.fallback_to_direct_time > 0.0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("CLIENT_FALLBACK_TO_DIRECT_TIME=%f", config.fallback_to_direct_time))
	}

	if config.high_bandwidth {
		cmd.Env = append(cmd.Env, "CLIENT_HIGH_BANDWIDTH=1")
	}

	if config.report_session {
		cmd.Env = append(cmd.Env, "CLIENT_REPORT_SESSION=1")
	}

	if config.big_packets {
		cmd.Env = append(cmd.Env, "CLIENT_BIG_PACKETS=1")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Start()

	return cmd, &stdout, &stderr
}

type ServerConfig struct {
	duration                       int
	no_upgrade                     bool
	upgrade_count                  int
	packet_loss                    bool
	buyer_private_key              string
	disable_network_next           bool
	server_address                 string
	server_port                    int
	restart_time                   float64
	tags_multi                     bool
	datacenter                     string
	disable_autodetect             bool
	force_resolve_hostname_timeout bool
	force_autodetect_timeout       bool
	session_events                 bool
	flush                          bool
}

func server(config *ServerConfig) (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command(serverBin)
	if cmd == nil {
		panic("could not create server!\n")
		return nil, nil
	}

	cmd.Env = os.Environ()

	cmd.Env = append(cmd.Env, "NEXT_DATACENTER=local")
	cmd.Env = append(cmd.Env, "NEXT_SERVER_BACKEND_HOSTNAME=127.0.0.1")
	cmd.Env = append(cmd.Env, "NEXT_PORT=45000")
	cmd.Env = append(cmd.Env, "NEXT_BUYER_PRIVATE_KEY=no")
	cmd.Env = append(cmd.Env, "NEXT_BUYER_PUBLIC_KEY=no")
	cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_SERVER_BACKEND_PUBLIC_KEY=%s", TestServerBackendPublicKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_RELAY_BACKEND_PUBLIC_KEY=%s", TestRelayBackendPublicKey))
	cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_BUYER_PUBLIC_KEY=%s", TestBuyerPublicKey))

	if config.duration != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("SERVER_DURATION=%d", config.duration))
	}

	if config.no_upgrade {
		cmd.Env = append(cmd.Env, "SERVER_NO_UPGRADE=1")
	}

	if config.upgrade_count > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("SERVER_UPGRADE_COUNT=%d", config.upgrade_count))
	}

	if config.buyer_private_key != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_BUYER_PRIVATE_KEY=%s", config.buyer_private_key))
	}

	if config.disable_network_next {
		cmd.Env = append(cmd.Env, "SERVER_DISABLE_NETWORK_NEXT=1")
	}

	if config.server_address != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_SERVER_ADDRESS=%s:%d", config.server_address, config.server_port))
		cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_BIND_ADDRESS=0.0.0.0:%d", config.server_port))
	}

	if config.packet_loss {
		cmd.Env = append(cmd.Env, "SERVER_PACKET_LOSS=1")
	}

	if config.restart_time > 0.0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("SERVER_RESTART_TIME=%.2f", config.restart_time))
	}

	if config.datacenter != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_DATACENTER=%s", config.datacenter))
	}

	if config.disable_autodetect {
		cmd.Env = append(cmd.Env, "NEXT_DISABLE_AUTODETECT=1")
	}

	if config.force_resolve_hostname_timeout {
		cmd.Env = append(cmd.Env, "NEXT_FORCE_RESOLVE_HOSTNAME_TIMEOUT=1")
	}

	if config.force_autodetect_timeout {
		cmd.Env = append(cmd.Env, "NEXT_FORCE_AUTODETECT_TIMEOUT=1")
	}

	if config.session_events {
		cmd.Env = append(cmd.Env, "SESSION_EVENTS=1")
	}

	if config.flush {
		cmd.Env = append(cmd.Env, "SERVER_FLUSH=1")
	}

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Start()

	return cmd, &output
}

const NEXT_CLIENT_COUNTER_OPEN_SESSION = 0
const NEXT_CLIENT_COUNTER_CLOSE_SESSION = 1
const NEXT_CLIENT_COUNTER_UPGRADE_SESSION = 2
const NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT = 3
const NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH = 4
const NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH = 5
const NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT = 6
const NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT = 7
const NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT = 8
const NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT = 9
const NEXT_CLIENT_COUNTER_MULTIPATH = 10
const NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER = 11
const NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT = 12
const NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_CLIENT_TO_SERVER = 13
const NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_SERVER_TO_CLIENT = 14

var ClientCounterNames = []string{
	"NEXT_CLIENT_COUNTER_OPEN_SESSION",
	"NEXT_CLIENT_COUNTER_CLOSE_SESSION",
	"NEXT_CLIENT_COUNTER_UPGRADE_SESSION",
	"NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT",
	"NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH",
	"NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH",
	"NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT",
	"NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT",
	"NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT",
	"NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT",
	"NEXT_CLIENT_COUNTER_MULTIPATH",
	"NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER",
	"NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT",
	"NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_CLIENT_TO_SERVER",
	"NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_SERVER_TO_CLIENT",
}

const NEXT_CLIENT_COUNTER_MAX = 64

func read_client_counters(output string) []uint64 {
	result_uint64 := make([]uint64, NEXT_CLIENT_COUNTER_MAX)
	result_strings := strings.Split(output, ",")
	for i := range result_strings {
		if i == NEXT_CLIENT_COUNTER_MAX-1 {
			break
		}
		result_uint64[i], _ = strconv.ParseUint(result_strings[i], 10, 64)
	}
	return result_uint64
}

func print_client_counters(client_counters []uint64) {
	for i := range client_counters {
		if i == len(ClientCounterNames) {
			break
		}
		fmt.Printf("%s = %d\n", ClientCounterNames[i], client_counters[i])
	}
}

func client_check(client_counters []uint64, client_stdout *bytes.Buffer, server_stdout *bytes.Buffer, backend_stdout *bytes.Buffer, condition bool, relay_stdouts ...*bytes.Buffer) {
	if !condition {
		if backend_stdout != nil {
			fmt.Printf("\n--------------------------------------------------------------------------\n")
			fmt.Printf("\n%s", backend_stdout)
		}
		fmt.Printf("\n--------------------------------------------------------------------------\n")
		fmt.Printf("\n%s", server_stdout)
		fmt.Printf("\n--------------------------------------------------------------------------\n")
		fmt.Printf("\n%s", client_stdout)
		for i, buff := range relay_stdouts {
			fmt.Printf("----------------------------Relay %d--------------------------------------\n\n%s", i, buff)
		}
		fmt.Printf("\n--------------------------------------------------------------------------\n")
		fmt.Printf("\n")
		print_client_counters(client_counters)
		fmt.Printf("\n--------------------------------------------------------------------------\n")
		fmt.Printf("\n")
		panic("client check failed!")
	}
}

func server_check(server_stdout *bytes.Buffer, backend_stdout *bytes.Buffer, condition bool) {
	if !condition {
		if backend_stdout != nil {
			fmt.Printf("\n--------------------------------------------------------------------------\n")
			fmt.Printf("\n%s", backend_stdout)
		}
		fmt.Printf("\n--------------------------------------------------------------------------\n")
		fmt.Printf("\n%s", server_stdout)
		fmt.Printf("\n--------------------------------------------------------------------------\n")
		panic("server check failed!")
	}
}

/*
   Test that when a client connects to a server with no backend running, and with no buyer public or private
   keys set on either client and server, that packets are sent and received direct. This is network next in direct mode.
*/

func test_passthrough() {

	fmt.Printf("test_passthrough\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.no_upgrade = true
	server_cmd, server_stdout := server(serverConfig)

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	server_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] > 40*60)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] == client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH])
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_SERVER_TO_CLIENT] == 0)

}

/*
   Run a backend but no relays. Make sure that we send and receive all packets direct.
   This tests the path where we prefix upgraded session direct packets with [1][sequence]
*/

func test_direct_upgraded() {

	fmt.Printf("test_direct_upgraded\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()

	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
	Test accelerated
   Verify that it actually works as advertised, by making sure we see send and received packets across both network next and direct.
*/

func test_accelerated() {

	fmt.Printf("test_accelerated\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("MULTIPATH")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 2000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] >= 2000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] >= 2000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] >= 2000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
   Verify that we can accelerate and weather 100% packet loss on the network next route.
   This means that the direct route successfully acts as a backup, greatly reducing risk for players
   that are getting Network Next acceleration. At worst case, NN route is broken, but direct takes over!
*/

func test_next_packet_loss() {

	fmt.Printf("test_next_packet_loss\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relayConfig := RelayConfig{
		fake_packet_loss_percent:    100.0,
		fake_packet_loss_start_time: 20.0,
	}

	relay_1_cmd, _ := relay("relay.1", 2000, relayConfig)
	relay_2_cmd, _ := relay("relay.2", 2001, relayConfig)
	relay_3_cmd, _ := relay("relay.3", 2002, relayConfig)

	backend_cmd, backend_stdout := backend("MULTIPATH")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] > 2000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 400)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 400)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
   Test that once we have upgraded a session and start sending packets over network next, that if the backend goes down
   we fallback to direct. Verify that we don't lose any packets when we do this. This is a critical test. It ensures that
   when our backend goes down in production we don't drop packets or disconnect players.
*/

func test_fallback_to_direct_backend() {

	fmt.Printf("test_fallback_to_direct_backend\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 70.0
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay("relay.1", 2000)
	relay_2_cmd, relay_2_stdout := relay("relay.2", 2001)
	relay_3_cmd, relay_3_stdout := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("DEFAULT")

	go func(cmd *exec.Cmd) {
		time.Sleep(time.Second * 30)
		cmd.Process.Signal(os.Interrupt)
	}(backend_cmd)

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_fallback_to_direct_client_side() {

	fmt.Printf("test_fallback_to_direct_client_side\n")

	clientConfig := &ClientConfig{}
	clientConfig.fallback_to_direct_time = 30.0
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay("relay.1", 2000)
	relay_2_cmd, relay_2_stdout := relay("relay.2", 2001)
	relay_3_cmd, relay_3_stdout := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	backendSawFallbackToDirect := strings.Contains(backend_stdout.String(), "error: fallback to direct")

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawFallbackToDirect)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)

}

/*
   Verify that the client falls back to direct and exchanges passthrough packets with the server
   after the server is restarted. Without this client side fallback to direct, it is possible
   for the client to get stuck in a state where it keep sending direct (upgraded) packets which
   won't get through to the server post restart.
*/

func test_fallback_to_direct_server_restart() {

	fmt.Printf("test_fallback_to_direct_server_restart\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 85.0
	clientConfig.duration = 90.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.restart_time = 15.0
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] > 2000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] > 2000)
}

/*
   Have network next enabled on a client, but disable it on a server.
   Verify that the client is still able to connect to the server, but all packets are sent as passthrough.
*/

func test_disable_on_server() {

	fmt.Printf("test_disable_on_server\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey
	serverConfig.disable_network_next = true

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] == 0)

}

/*
   Have network next enabled on the server, but disable network next via config bool in the SDK on the client.
   Verify that the client is still able to connect to the server, but all packets are sent direct.
*/

func test_disable_on_client() {

	fmt.Printf("test_disable_on_client\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey
	clientConfig.disable_network_next = true

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("DEFAULT")

	fmt.Printf("waiting for client\n")
	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	fmt.Printf("waiting for server\n")
	server_cmd.Wait()

	fmt.Printf("waiting for backend\n")
	backend_cmd.Wait()

	fmt.Printf("waiting for relays\n")
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] == 0)

}

/*
   Put the backend into a mode where it randomly switches routes every slice.
   Verify that the SDK is able to properly handle route switches without dropping packets.
*/

func test_route_switching() {

	fmt.Printf("test_route_switching\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("ROUTE_SWITCHING")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
   Put the backend into a mode where it serves up even slices on network next, and odd slices going direct.
   Verify that the SDK is able to handle transitioning from direct -> next, and next -> direct without dropping packets.
*/

func test_on_off() {

	fmt.Printf("test_on_off\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("ON_OFF")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
   Put the backend into a mode where every three slices, it serves two slices on network next, and the third slice going direct.
   Verify that the SDK is able to handle transitioning from direct -> next -> continue -> direct without dropping packets or falling back to direct.
*/

func test_on_on_off() {

	fmt.Printf("test_on_on_off\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("ON_ON_OFF")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
   Test that a client is able to connect to a server direct, and then reconnect without problems.
   This verifies that our code in the SDK to distinguish the old session from the new one is working properly for
   upgraded direct packets (1 prefix).
*/

func test_reconnect_direct() {

	fmt.Printf("test_reconnect_direct\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 55.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey
	clientConfig.connect_time = 30.0
	clientConfig.connect_address = "127.0.0.1:32202"

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 2900)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 2900)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
}

/*
   Test that a client is able to connect to a server direct, and then reconnect (without upgrading) without problems.
   This verifies that the previous session doesn't interfere with passthrough packets sent in the reconnect session.
*/

func test_reconnect_direct_no_upgrade() {

	fmt.Printf("test_reconnect_direct_no_upgrade\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 55.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey
	clientConfig.connect_time = 30.0
	clientConfig.connect_address = "127.0.0.1:32202"

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.upgrade_count = 1
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 1000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 1000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] > 1000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] > 1000)
}

/*
   Connect to a server over network next, and then reconnect to that server over network next.
   This verifies that our sequence numbers are working properly for network next packets across reconnect.
   We've had a lot of problems in the past with this not working properly, so this test locks in correct behavior.
*/

func test_reconnect_next() {

	fmt.Printf("test_reconnect_next\n")

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("DEFAULT")

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	// IMPORTANT: give the server time to ping server relays and get ready
	time.Sleep(time.Second * 10)

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey
	clientConfig.connect_time = 30.0
	clientConfig.connect_address = "127.0.0.1:32202"

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT])
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT])
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 3000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 2900)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
   Make sure a client can connect direct to one server, and then connect direct to another without problems.
*/

func test_connect_to_another_server_direct() {

	fmt.Printf("test_connect_to_another_server_direct\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey
	clientConfig.connect_time = 30.0
	clientConfig.connect_address = "127.0.0.1:32203"

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig1 := &ServerConfig{}
	serverConfig1.buyer_private_key = TestBuyerPrivateKey
	server_1_cmd, _ := server(serverConfig1)

	serverConfig2 := &ServerConfig{}
	serverConfig2.buyer_private_key = TestBuyerPrivateKey
	serverConfig2.server_address = "127.0.0.1"
	serverConfig2.server_port = 32203
	server_2_cmd, server_stdout := server(serverConfig2)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_1_cmd.Process.Signal(os.Interrupt)
	server_2_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_1_cmd.Wait()
	server_2_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
   Make sure a client can connect over network next to one server, and then connect to another server over network next.
*/

func test_connect_to_another_server_next() {

	fmt.Printf("test_connect_to_another_server_next\n")

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("DEFAULT")

	// IMPORTANT: give the relays time initialize with the backend
	time.Sleep(time.Second * 10)

	serverConfig1 := &ServerConfig{}
	serverConfig1.buyer_private_key = TestBuyerPrivateKey
	server_1_cmd, _ := server(serverConfig1)

	serverConfig2 := &ServerConfig{}
	serverConfig2.buyer_private_key = TestBuyerPrivateKey
	serverConfig2.server_address = "127.0.0.1"
	serverConfig2.server_port = 32203
	server_2_cmd, server_stdout := server(serverConfig2)

	// IMPORTANT: give the servers time to ping server relays and get ready
	time.Sleep(time.Second * 10)

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey
	clientConfig.connect_time = 30.0
	clientConfig.connect_address = "127.0.0.1:32203"

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	client_cmd.Wait()

	server_1_cmd.Process.Signal(os.Interrupt)
	server_2_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_1_cmd.Wait()
	server_2_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT])
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT])
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 2900)
}

/*
   Simulate packet loss between the client, server and backend.
   Make sure we can still get a direct route.
*/

func test_packet_loss_direct() {

	fmt.Printf("test_packet_loss_direct\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey
	clientConfig.packet_loss = true

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey
	serverConfig.packet_loss = true

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] > 250)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] > 250)

}

/*
   Simulate packet loss between the client, server and backend.
   Make sure we can still get a network next route.
*/

func test_packet_loss_next() {

	fmt.Printf("test_packet_loss_next\n")

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("DEFAULT")

	// IMPORTANT: give the relays time to initialize with the backend
	time.Sleep(time.Second * 10)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey
	serverConfig.packet_loss = true

	server_cmd, server_stdout := server(serverConfig)

	// IMPORTANT: give the server time to ping server relays and get ready
	time.Sleep(time.Second * 10)

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey
	clientConfig.packet_loss = true

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]+client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]+client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] > 200)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] > 200)

}

/*
   Run a bunch of clients and make sure that we are able to connect to the server and exchange packets over network next.
   This is sort of a miniature load test, it verifies that the server SDK is able to handle multiple client connections
   without dropping packets or getting confused (eg. crossed wires).
*/

func test_server_under_load() {

	fmt.Printf("test_server_under_load\n")

	backend_cmd, backend_stdout := backend("DEFAULT")

	relay_1_cmd, relay_1_stdout := relay("relay.1", 2000)
	relay_2_cmd, relay_2_stdout := relay("relay.2", 2001)
	relay_3_cmd, relay_3_stdout := relay("relay.3", 2002)

	relaysInited := false
	for i := 0; i < 60; i++ {
		if strings.Contains(relay_1_stdout.String(), "Relay initialized") && strings.Contains(relay_2_stdout.String(), "Relay initialized") && strings.Contains(relay_3_stdout.String(), "Relay initialized") {
			relaysInited = true
			break
		}
		time.Sleep(time.Second)
	}

	if !relaysInited {
		panic("relays did not init")
	}

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	const MaxClients = 100

	clientConfig := &ClientConfig{}
	clientConfig.duration = 50
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd := make([]*exec.Cmd, MaxClients)
	client_stdout := make([]*bytes.Buffer, MaxClients)
	client_stderr := make([]*bytes.Buffer, MaxClients)
	for i := 0; i < MaxClients; i++ {
		client_cmd[i], client_stdout[i], client_stderr[i] = client(clientConfig)
	}

	time.Sleep(time.Second * 60)

	for i := 0; i < MaxClients; i++ {
		server_cmd.Process.Signal(os.Interrupt)
	}

	server_cmd.Wait()

	for i := 0; i < MaxClients; i++ {
		client_cmd[i].Process.Signal(os.Interrupt)
		client_cmd[i].Wait()
	}

	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	for i := 0; i < MaxClients; i++ {

		client_counters := read_client_counters(client_stderr[i].String())

		backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
		backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	}
}

/*
   Put the backend in a mode where it ignores session update packets until the final retry.
   Verify that the network next route works perfectly even though retries are necessary.
*/

func test_session_update_retry() {

	fmt.Printf("test_session_update_retry\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay("relay.1", 2000)
	relay_2_cmd, relay_2_stdout := relay("relay.2", 2001)
	relay_3_cmd, relay_3_stdout := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("FORCE_RETRY")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] >= 30*60, relay_1_stdout, relay_2_stdout, relay_3_stdout)

}

/*
   Create fake packet loss between the client and server.
   Make sure the packet loss is reported up to the backend.
*/

func test_packet_loss() {

	fmt.Printf("test_packet_loss\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relayConfig := RelayConfig{
		fake_packet_loss_percent:    1.0,
		fake_packet_loss_start_time: 10.0,
	}

	relay_1_cmd, relay_1_stdout := relay("relay.1", 2000, relayConfig)
	relay_2_cmd, relay_2_stdout := relay("relay.2", 2001, relayConfig)
	relay_3_cmd, relay_3_stdout := relay("relay.3", 2002, relayConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawClientToServerPacketLoss := strings.Contains(backend_stdout.String(), "client to server packets lost")
	backendSawServerToClientPacketLoss := strings.Contains(backend_stdout.String(), "server to client packets lost")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientToServerPacketLoss == true)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerToClientPacketLoss == true)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived < totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] != 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] != 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)

}

/*
   Make sure that the backend sees non-zero bandwidth up/down reported from the SDK.
*/

func test_bandwidth() {

	fmt.Printf("test_bandwidth\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("BANDWIDTH")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()

	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawBandwidthUp := strings.Contains(backend_stdout.String(), "bandwidth kbps up")
	backendSawBandwidthDown := strings.Contains(backend_stdout.String(), "bandwidth kbps down")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawBandwidthUp)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawBandwidthDown)

}

/*
   Make sure that the backend sees non-zero jitter reported from the SDK.
*/

func test_jitter() {

	fmt.Printf("test_jitter\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("JITTER")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawJitterUp := strings.Contains(backend_stdout.String(), "jitter up")
	backendSawJitterDown := strings.Contains(backend_stdout.String(), "jitter down")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawJitterUp)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawJitterDown)
}

/*
   Make sure all the direct stats (RTT, jitter, PL) are uploaded to the backend.
*/

func test_direct_stats() {

	fmt.Printf("test_direct_stats\n")

	clientConfig := &ClientConfig{}
	clientConfig.fake_direct_packet_loss = 10.0
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DIRECT_STATS")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]

	backendSawDirectStats := strings.Contains(backend_stdout.String(), "direct rtt =")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawDirectStats)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 30*60)

}

/*
   Make sure all the next stats (RTT, jitter, PL) are uploaded to the backend.
*/

func test_next_stats() {

	fmt.Printf("test_next_stats\n")

	clientConfig := &ClientConfig{}
	clientConfig.fake_next_packet_loss = 10.0
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay("relay.1", 2000)
	relay_2_cmd, relay_2_stdout := relay("relay.2", 2001)
	relay_3_cmd, relay_3_stdout := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("NEXT_STATS")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]

	backendSawNextStats := strings.Contains(backend_stdout.String(), "next rtt =")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawNextStats)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] >= 30*60, relay_1_stdout, relay_2_stdout, relay_3_stdout)

}

/*
   Test that the backend sees the client report a session
*/

func test_report_session() {

	fmt.Printf("test_report_session\n")

	clientConfig := &ClientConfig{}
	clientConfig.report_session = true
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawClientReportSession := strings.Contains(backend_stdout.String(), "client reported session")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientReportSession == true)
}

/*
   Test that the backend sees the client ping timeout on the server when the client is stopped.
*/

func test_client_ping_timed_out() {

	fmt.Printf("test_client_ping_timed_out\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 30.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay("relay.1", 2000)
	relay_2_cmd, _ := relay("relay.2", 2001)
	relay_3_cmd, _ := relay("relay.3", 2002)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 60)

	client_cmd.Process.Signal(os.Interrupt)
	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)
	relay_1_cmd.Process.Signal(os.Interrupt)
	relay_2_cmd.Process.Signal(os.Interrupt)
	relay_3_cmd.Process.Signal(os.Interrupt)

	client_cmd.Wait()
	server_cmd.Wait()
	backend_cmd.Wait()
	relay_1_cmd.Wait()
	relay_2_cmd.Wait()
	relay_3_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawClientPingTimedOut := strings.Contains(backend_stdout.String(), "client ping timed out")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientPingTimedOut == true)

}

func test_server_ready_success() {

	fmt.Printf("test_server_ready_success\n")

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 25)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverInitSuccessful := strings.Contains(server_stdout.String(), "info: welcome to network next :)")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'local'")

	server_check(server_stdout, backend_stdout, serverInitSuccessful)
	server_check(server_stdout, backend_stdout, serverReady)
	server_check(server_stdout, backend_stdout, serverDatacenter)

}

func test_server_ready_fallback_to_direct() {

	fmt.Printf("test_server_ready_fallback_to_direct\n")

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	time.Sleep(time.Second * 25)

	server_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()

	serverFallbackToDirect := strings.Contains(server_stdout.String(), "info: server init timed out. falling back to direct mode only :(")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'local'")

	server_check(server_stdout, nil, serverFallbackToDirect)
	server_check(server_stdout, nil, serverReady)
	server_check(server_stdout, nil, serverDatacenter)

}

func test_server_ready_autodetect_cloud() {

	fmt.Printf("test_server_ready_autodetect_cloud\n")

	serverConfig := &ServerConfig{}
	serverConfig.datacenter = "cloud"
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 25)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverAutodetecting := strings.Contains(server_stdout.String(), "info: server attempting to autodetect datacenter")
	serverGoogleAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in google cloud")
	serverAmazonAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in amazon cloud")
	serverAutodetectFailed := strings.Contains(server_stdout.String(), "info: server autodetect datacenter failed. sticking with 'cloud' [9ebb5c9513bac4fe]")
	serverAutodetectTimedOut := strings.Contains(server_stdout.String(), "info: server autodetect datacenter timed out. sticking with 'cloud' [9ebb5c9513bac4fe]")

	server_check(server_stdout, backend_stdout, serverAutodetecting)
	server_check(server_stdout, backend_stdout, serverGoogleAutodetect)
	server_check(server_stdout, backend_stdout, serverAmazonAutodetect)
	server_check(server_stdout, backend_stdout, serverAutodetectFailed || serverAutodetectTimedOut)

}

func test_server_ready_disable_autodetect_cloud() {

	fmt.Printf("test_server_ready_disable_autodetect_cloud\n")

	serverConfig := &ServerConfig{}
	serverConfig.datacenter = "cloud"
	serverConfig.disable_autodetect = true
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 25)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverInitSuccessful := strings.Contains(server_stdout.String(), "info: welcome to network next :)")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'cloud'")
	serverAutodetecting := strings.Contains(server_stdout.String(), "info: server attempting to autodetect datacenter")
	serverGoogleAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in google cloud")
	serverAmazonAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in amazon cloud")
	serverAutodetectFailed := strings.Contains(server_stdout.String(), "info: server autodetect datacenter failed. sticking with 'cloud' [9ebb5c9513bac4fe]")

	server_check(server_stdout, backend_stdout, serverInitSuccessful)
	server_check(server_stdout, backend_stdout, serverReady)
	server_check(server_stdout, backend_stdout, serverDatacenter)
	server_check(server_stdout, backend_stdout, !serverAutodetecting)
	server_check(server_stdout, backend_stdout, !serverGoogleAutodetect)
	server_check(server_stdout, backend_stdout, !serverAmazonAutodetect)
	server_check(server_stdout, backend_stdout, !serverAutodetectFailed)

}

func test_server_ready_resolve_hostname_timeout() {

	fmt.Printf("test_server_ready_resolve_hostname_timeout\n")

	serverConfig := &ServerConfig{}
	serverConfig.datacenter = "local"
	serverConfig.force_resolve_hostname_timeout = true
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 25)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverFallbackToDirect := strings.Contains(server_stdout.String(), "info: server init timed out. falling back to direct mode only :(")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'local'")
	serverTimedOutHostnameResolve := strings.Contains(server_stdout.String(), "resolve hostname timed out")

	server_check(server_stdout, backend_stdout, serverFallbackToDirect)
	server_check(server_stdout, backend_stdout, serverReady)
	server_check(server_stdout, backend_stdout, serverDatacenter)
	server_check(server_stdout, backend_stdout, serverTimedOutHostnameResolve)

}

func test_server_ready_autodetect_timeout() {

	fmt.Printf("test_server_ready_autodetect_timeout\n")

	serverConfig := &ServerConfig{}
	serverConfig.datacenter = "local"
	serverConfig.force_autodetect_timeout = true
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 25)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverFallbackToDirect := strings.Contains(server_stdout.String(), "info: server init timed out. falling back to direct mode only :(")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'local'")
	serverTimedOutAutodetect := strings.Contains(server_stdout.String(), "autodetect timed out. sticking with 'local' [249f1fb6f3a680e8]")

	server_check(server_stdout, backend_stdout, serverFallbackToDirect)
	server_check(server_stdout, backend_stdout, serverReady)
	server_check(server_stdout, backend_stdout, serverDatacenter)
	server_check(server_stdout, backend_stdout, serverTimedOutAutodetect)

}

func test_client_connect_before_ready() {

	fmt.Printf("test_client_connect_before_ready\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.force_resolve_hostname_timeout = true
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverFallbackToDirect := strings.Contains(server_stdout.String(), "info: server init timed out. falling back to direct mode only :(")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'local'")
	serverTimedOutHostnameResolve := strings.Contains(server_stdout.String(), "resolve hostname timed out")

	server_check(server_stdout, backend_stdout, serverFallbackToDirect)
	server_check(server_stdout, backend_stdout, serverReady)
	server_check(server_stdout, backend_stdout, serverDatacenter)
	server_check(server_stdout, backend_stdout, serverTimedOutHostnameResolve)

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_session_events() {

	fmt.Printf("test_session_events\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.session_events = true
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawSessionEvents := strings.Contains(backend_stdout.String(), "session events 40100400")

	serverFlushedSessionEvents := strings.Contains(server_stdout.String(), "server flushed session events 40100400 to backend")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawSessionEvents)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverFlushedSessionEvents)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_flush() {

	fmt.Printf("test_flush\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 40.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.flush = true
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()

	backend_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawSessionUpdate := strings.Contains(backend_stdout.String(), "client ping timed out")

	serverSawFlushRequest := strings.Contains(server_stdout.String(), "server flush started")
	serverSawSessionUpdateFlush := strings.Contains(server_stdout.String(), "server flushed session update")
	serverSawFlushComplete := strings.Contains(server_stdout.String(), "server flush finished")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawSessionUpdate)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushRequest)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawSessionUpdateFlush)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushComplete)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 30*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] >= 30*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 30*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_flush_retry() {

	fmt.Printf("test_flush_retry\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 40.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.flush = true
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("FORCE_RETRY")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()

	backend_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawSessionUpdate := strings.Contains(backend_stdout.String(), "client ping timed out")

	serverSawFlushRequest := strings.Contains(server_stdout.String(), "server flush started")
	serverSawSessionUpdateFlush := strings.Contains(server_stdout.String(), "server flushed session update")
	serverSawFlushComplete := strings.Contains(server_stdout.String(), "server flush finished")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawSessionUpdate)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushRequest)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawSessionUpdateFlush)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushComplete)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 30*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] >= 30*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 30*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_flush_session_events() {

	fmt.Printf("test_flush_session_events\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 40.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.session_events = true
	serverConfig.flush = true
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()

	backend_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawSessionUpdate := strings.Contains(backend_stdout.String(), "client ping timed out")
	backendSawSessionEvents := strings.Contains(backend_stdout.String(), "session events 40100400")

	serverFlushedSessionEvents := strings.Contains(server_stdout.String(), "server flushed session events 40100400 to backend")
	serverSawFlushRequest := strings.Contains(server_stdout.String(), "server flush started")
	serverSawSessionUpdateFlush := strings.Contains(server_stdout.String(), "server flushed session update")
	serverSawFlushComplete := strings.Contains(server_stdout.String(), "server flush finished")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawSessionUpdate)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawSessionEvents)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverFlushedSessionEvents)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushRequest)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawSessionUpdateFlush)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushComplete)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 30*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] >= 30*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 30*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_flush_session_events_retry() {

	fmt.Printf("test_flush_session_events_retry\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 30.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.session_events = true
	serverConfig.flush = true
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()

	backend_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawSessionUpdate := strings.Contains(backend_stdout.String(), "client ping timed out")
	backendSawSessionEvents := strings.Contains(backend_stdout.String(), "session events 40100400")

	serverFlushedSessionEvents := strings.Contains(server_stdout.String(), "server flushed session events 40100400 to backend")
	serverSawFlushRequest := strings.Contains(server_stdout.String(), "server flush started")
	serverSawSessionUpdateFlush := strings.Contains(server_stdout.String(), "server flushed session update")
	serverSawFlushComplete := strings.Contains(server_stdout.String(), "server flush finished")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawSessionUpdate)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawSessionEvents)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverFlushedSessionEvents)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushRequest)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawSessionUpdateFlush)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushComplete)

}

func test_big_packets() {

	fmt.Printf("test_big_packets\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.buyer_public_key = TestBuyerPublicKey
	clientConfig.big_packets = true

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.buyer_private_key = TestBuyerPrivateKey

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

type test_function func()

func main() {
	allTests := []test_function{
		test_passthrough,
		test_direct_upgraded,
		test_accelerated,
		test_next_packet_loss,
		test_fallback_to_direct_backend,
		test_fallback_to_direct_client_side,
		test_fallback_to_direct_server_restart,
		test_disable_on_server,
		test_disable_on_client,
		test_route_switching,
		test_on_off,
		test_on_on_off,
		test_reconnect_direct,
		test_reconnect_direct_no_upgrade,
		test_reconnect_next,
		test_connect_to_another_server_direct,
		test_connect_to_another_server_next,
		test_packet_loss_direct,
		test_packet_loss_next,
		test_server_under_load,
		test_session_update_retry,
		test_packet_loss,
		test_bandwidth,
		test_jitter,
		test_direct_stats,
		test_next_stats,
		test_report_session,
		test_client_ping_timed_out,
		test_server_ready_success,
		test_server_ready_fallback_to_direct,
		test_server_ready_autodetect_cloud,
		test_server_ready_disable_autodetect_cloud,
		test_server_ready_resolve_hostname_timeout,
		test_server_ready_autodetect_timeout,
		test_client_connect_before_ready,
		test_session_events,
		test_flush,
		test_flush_retry,
		test_flush_session_events,
		test_flush_session_events_retry,
		test_big_packets,
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
