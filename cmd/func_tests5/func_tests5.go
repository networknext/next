/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	relayBin   = "./dist/reference_relay"
	backendBin = "./dist/func_backend5"
	clientBin  = "./dist/func_client5"
	serverBin  = "./dist/func_server5"
)

func backend(mode string) (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command(backendBin)
	if cmd == nil {
		panic("could not create backend!\n")
		return nil, nil
	}

	if mode != "" {
		cmd.Env = os.Environ()
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

func relay(configArray ...RelayConfig) (*exec.Cmd, *bytes.Buffer) {

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
	cmd.Env = append(cmd.Env, "RELAY_DEV=1")
	cmd.Env = append(cmd.Env, "RELAY_MASTER=127.0.0.1")
	cmd.Env = append(cmd.Env, "RELAY_PORT=0")
	cmd.Env = append(cmd.Env, "RELAY_NAME=local")
	cmd.Env = append(cmd.Env, "RELAY_UPDATE_KEY=eyqNheTBdx+97qd3Nkf/QvjaSDQVQQzHvkhX6w9cvMO276rgKZ7VIPHwaoNE7f9SiQW6yThhEC5onwpBEFFdaw==")
	cmd.Env = append(cmd.Env, "RELAY_PUBLIC_KEY=9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=")
	cmd.Env = append(cmd.Env, "RELAY_PRIVATE_KEY=lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=")
	cmd.Env = append(cmd.Env, "RELAY_ADDRESS=127.0.0.1")
	cmd.Env = append(cmd.Env, "RELAY_BACKEND_HOSTNAME=http://127.0.0.1:30000")
	cmd.Env = append(cmd.Env, "RELAY_ROUTER_PUBLIC_KEY=SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=")
	cmd.Env = append(cmd.Env, "RELAY_BIND_ADDRESS=127.0.0.1")
	cmd.Env = append(cmd.Env, "RELAY_PUBLIC_ADDRESS=127.0.0.1")
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
	customer_public_key       string
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
}

func client(config *ClientConfig) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {

	cmd := exec.Command(clientBin)
	if cmd == nil {
		panic("could not create client!\n")
		return nil, nil, nil
	}

	cmd.Env = os.Environ()

	if config.duration != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("CLIENT_DURATION=%d", config.duration))
	}

	if config.customer_public_key != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_CUSTOMER_PUBLIC_KEY=%s", config.customer_public_key))
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
	customer_private_key           string
	disable_network_next           bool
	server_address                 string
	server_port                    int
	restart_time                   float64
	tags_multi                     bool
	datacenter                     string
	disable_autodetect             bool
	force_resolve_hostname_timeout bool
	force_autodetect_timeout       bool
	server_events                  bool
	match_data                     bool
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
	cmd.Env = append(cmd.Env, "NEXT_HOSTNAME=127.0.0.1")
	cmd.Env = append(cmd.Env, "NEXT_PORT=40000")
	cmd.Env = append(cmd.Env, "NEXT_CUSTOMER_PRIVATE_KEY=no")
	cmd.Env = append(cmd.Env, "NEXT_CUSTOMER_PUBLIC_KEY=no")

	if config.duration != 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("SERVER_DURATION=%d", config.duration))
	}

	if config.no_upgrade {
		cmd.Env = append(cmd.Env, "SERVER_NO_UPGRADE=1")
	}

	if config.upgrade_count > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("SERVER_UPGRADE_COUNT=%d", config.upgrade_count))
	}

	if config.customer_private_key != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("NEXT_CUSTOMER_PRIVATE_KEY=%s", config.customer_private_key))
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

	if config.tags_multi {
		cmd.Env = append(cmd.Env, "SERVER_TAGS_MULTI=1")
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

	if config.server_events {
		cmd.Env = append(cmd.Env, "SERVER_EVENTS=1")
	}

	if config.match_data {
		cmd.Env = append(cmd.Env, "SERVER_MATCH_DATA=1")
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
	Test that when a client connects to a server with no backend running, and with no customer public or private
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
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
	Run a backend and several relays. Verify that the session is upgraded and starts sending and receiving packets
	over network next. This is the first test that will likely fail if something is wrong with the backend or the
	relays.
*/

func test_network_next_route() {

	fmt.Printf("test_network_next_route\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay()
	relay_2_cmd, relay_2_stdout := relay()
	relay_3_cmd, relay_3_stdout := relay()

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

	backendSawClientPingTimeOut := strings.Contains(backend_stdout.String(), "client ping timed out")
	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientPingTimeOut == false)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] >= 40*60, relay_1_stdout, relay_2_stdout, relay_3_stdout)

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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay()
	relay_2_cmd, relay_2_stdout := relay()
	relay_3_cmd, relay_3_stdout := relay()

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_fallback_to_direct_client_side() {

	fmt.Printf("test_fallback_to_direct_client_side\n")

	clientConfig := &ClientConfig{}
	clientConfig.fallback_to_direct_time = 30.0
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay()
	relay_2_cmd, relay_2_stdout := relay()
	relay_3_cmd, relay_3_stdout := relay()

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.stop_sending_packets_time = 55.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.restart_time = 15.0
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"
	serverConfig.disable_network_next = true

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="
	clientConfig.disable_network_next = true

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="
	clientConfig.connect_time = 30.0
	clientConfig.connect_address = "127.0.0.1:32202"

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="
	clientConfig.connect_time = 30.0
	clientConfig.connect_address = "127.0.0.1:32202"

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.upgrade_count = 1
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 1750)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 1750)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] > 1250)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] > 1250)
}

/*
	Connect to a server over network next, and then reconnect to that server over network next.
	This verifies that our sequence numbers are working properly for network next packets across reconnect.
	We've had a lot of problems in the past with this not working properly, so this test locks in correct behavior.
*/

func test_reconnect_next() {

	fmt.Printf("test_reconnect_next\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="
	clientConfig.connect_time = 30.0
	clientConfig.connect_address = "127.0.0.1:32202"

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="
	clientConfig.connect_time = 30.0
	clientConfig.connect_address = "127.0.0.1:32203"

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig1 := &ServerConfig{}
	serverConfig1.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"
	server_1_cmd, _ := server(serverConfig1)

	serverConfig2 := &ServerConfig{}
	serverConfig2.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
	Make sure a client can connect over network next to one server, and then connect to another server over network next.
*/

func test_connect_to_another_server_next() {

	fmt.Printf("test_connect_to_another_server_next\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="
	clientConfig.connect_time = 30.0
	clientConfig.connect_address = "127.0.0.1:32203"

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig1 := &ServerConfig{}
	serverConfig1.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"
	server_1_cmd, _ := server(serverConfig1)

	serverConfig2 := &ServerConfig{}
	serverConfig2.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"
	serverConfig2.server_address = "127.0.0.1"
	serverConfig2.server_port = 32203
	server_2_cmd, server_stdout := server(serverConfig2)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

	backend_cmd, backend_stdout := backend("DEFAULT")

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
	Multipath feature sends packets across network next and direct at the same time.
	Verify that it actually works as advertised, by making sure we see send and received packets across both network next and direct.
*/

func test_multipath() {

	fmt.Printf("test_multipath\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
	Verify that we can connect and go multipath, and weather 100% packet loss on the network next route.
	This means that the direct route successfully acts as a backup, greatly reducing risk for players (eg. ESL pros)
	that are getting Network Next acceleration. At worst case, NN route is broken, but direct takes over!
*/

func test_multipath_next_packet_loss() {

	fmt.Printf("test_multipath_next_packet_loss\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relayConfig := RelayConfig{
		fake_packet_loss_percent:    100.0,
		fake_packet_loss_start_time: 20.0,
	}

	relay_1_cmd, _ := relay(relayConfig)
	relay_2_cmd, _ := relay(relayConfig)
	relay_3_cmd, _ := relay(relayConfig)

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
	Make sure that fallback to direct works if the backend goes down while in multipath.
*/

func test_multipath_fallback_to_direct() {

	fmt.Printf("test_multipath_fallback_to_direct\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relayConfig := RelayConfig{
		fake_packet_loss_percent:    100.0,
		fake_packet_loss_start_time: 20.0,
	}

	relay_1_cmd, _ := relay(relayConfig)
	relay_2_cmd, _ := relay(relayConfig)
	relay_3_cmd, _ := relay(relayConfig)

	backend_cmd, backend_stdout := backend("MULTIPATH")

	go func(cmd *exec.Cmd) {
		time.Sleep(time.Second * 20)
		cmd.Process.Signal(os.Interrupt)
	}(backend_cmd)

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] > 3000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 400)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 400)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
	Put the backend into a mode where it sets "committed" flag to false in routes returned to the SDK.
	Verify that the SDK gets network next routes, but doesn't actually send packets across them if committed is false.
*/

func test_uncommitted() {

	fmt.Printf("test_uncommitted\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

	backend_cmd, backend_stdout := backend("UNCOMMITTED")

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
	Test that the SDK is able to transition from uncommitted to comitted state.
	This is what we use to implement "try before you buy" feature in the backend, eg. get a route, trial it first for a slice or more
	before actually sending packets over network next. This test makes sure that packets are actually sent over network next after
	we transition from committed = false to committed = true for a session.
*/

func test_uncommitted_to_committed() {

	fmt.Printf("test_uncommitted_to_committed\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

	backend_cmd, backend_stdout := backend("UNCOMMITTED_TO_COMMITTED")

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

/*
	Simulate packet loss between the client, server and backend.
	Make sure we can still get a direct route.
*/

func test_packet_loss_direct() {

	fmt.Printf("test_packet_loss_direct\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="
	clientConfig.packet_loss = true

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] > 250)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] > 250)

}

/*
	Simulate packet loss between the client, server and backend.
	Make sure we can still get a network next route.
*/

func test_packet_loss_next() {

	fmt.Printf("test_packet_loss_next\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="
	clientConfig.packet_loss = true

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"
	serverConfig.packet_loss = true

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

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

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	const MaxClients = 32

	client_cmd := make([]*exec.Cmd, MaxClients)
	client_stdout := make([]*bytes.Buffer, MaxClients)
	client_stderr := make([]*bytes.Buffer, MaxClients)
	for i := 0; i < MaxClients; i++ {
		client_cmd[i], client_stdout[i], client_stderr[i] = client(clientConfig)
	}

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay()
	relay_2_cmd, relay_2_stdout := relay()
	relay_3_cmd, relay_3_stdout := relay()

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 60)

	for i := 0; i < MaxClients; i++ {
		client_cmd[i].Process.Signal(os.Interrupt)
	}

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

	for i := 0; i < MaxClients; i++ {

		client_cmd[i].Wait()

		client_counters := read_client_counters(client_stderr[i].String())

		backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
		backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)

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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay()
	relay_2_cmd, relay_2_stdout := relay()
	relay_3_cmd, relay_3_stdout := relay()

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] >= 30*60, relay_1_stdout, relay_2_stdout, relay_3_stdout)

}

/*
	Send bandwidth over the limit between client and server.
	Make sure the backends sees both client and server going over the bandwidth limit.
	This way we can be sure that whenever real customers go over the limit, the SDK uploads that data
	to the backend so we can analyze it.
*/

func test_bandwidth_over_limit() {

	fmt.Printf("test_bandwidth_over_limit\n")

	clientConfig := &ClientConfig{}
	clientConfig.high_bandwidth = true
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay()
	relay_2_cmd, relay_2_stdout := relay()
	relay_3_cmd, relay_3_stdout := relay()

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

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == true)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == true)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)

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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relayConfig := RelayConfig{
		fake_packet_loss_percent:    1.0,
		fake_packet_loss_start_time: 10.0,
	}

	relay_1_cmd, relay_1_stdout := relay(relayConfig)
	relay_2_cmd, relay_2_stdout := relay(relayConfig)
	relay_3_cmd, relay_3_stdout := relay(relayConfig)

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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay()
	relay_2_cmd, relay_2_stdout := relay()
	relay_3_cmd, relay_3_stdout := relay()

	backend_cmd, backend_stdout := backend("BANDWIDTH")

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

	backendSawBandwidthUp := strings.Contains(backend_stdout.String(), "kbps up")
	backendSawBandwidthDown := strings.Contains(backend_stdout.String(), "kbps down")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawBandwidthUp)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawBandwidthDown)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] >= 40*60, relay_1_stdout, relay_2_stdout, relay_3_stdout)

}

/*
	Make sure that the backend sees non-zero jitter reported from the SDK.
*/

func test_jitter() {

	fmt.Printf("test_jitter\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay()
	relay_2_cmd, relay_2_stdout := relay()
	relay_3_cmd, relay_3_stdout := relay()

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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawJitterUp := strings.Contains(backend_stdout.String(), "jitter up")
	backendSawJitterDown := strings.Contains(backend_stdout.String(), "jitter down")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawJitterUp)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawJitterDown)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] >= 40*60, relay_1_stdout, relay_2_stdout, relay_3_stdout)

}

/*
	Make sure the backend sees the tag applied on the server.
*/

func test_tags() {

	fmt.Printf("test_tags\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("TAGS")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawTag := strings.Contains(backend_stdout.String(), "tag f9e6e6ef197c2b25")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawTag)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 40*60)

}

/*
	Make sure the backend sees multiple tags applied on the server.
*/

func test_tags_multi() {

	fmt.Printf("test_tags_multi\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.tags_multi = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("TAGS")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_PASSTHROUGH] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawTag1 := strings.Contains(backend_stdout.String(), "tag 77fd571956a1f7f8")
	backendSawTag2 := strings.Contains(backend_stdout.String(), "tag 528662164ef579d6")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawTag1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawTag2)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_PASSTHROUGH]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 40*60)

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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay()
	relay_2_cmd, relay_2_stdout := relay()
	relay_3_cmd, relay_3_stdout := relay()

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, relay_1_stdout := relay()
	relay_2_cmd, relay_2_stdout := relay()
	relay_3_cmd, relay_3_stdout := relay()

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

	backendSawClientReportSession := strings.Contains(backend_stdout.String(), "client reported session")
	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientReportSession == true)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] >= 40*60, relay_1_stdout, relay_2_stdout, relay_3_stdout)

}

/*
	Test that the backend sees the client ping timeout on the server when the client is stopped.
*/

func test_client_ping_timed_out() {

	fmt.Printf("test_client_ping_timed_out\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 30.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

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
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 15)

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
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	time.Sleep(time.Second * 15)

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
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 15)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverInitSuccessful := strings.Contains(server_stdout.String(), "info: welcome to network next :)")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'cloud'")
	serverAutodetecting := strings.Contains(server_stdout.String(), "info: server attempting to autodetect datacenter")
	serverGoogleAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in google cloud")
	serverAmazonAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in AWS")
	serverMultiplayAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in multiplay")
	serverAutodetectFailed := strings.Contains(server_stdout.String(), "info: server autodetect datacenter failed. sticking with 'cloud' [9ebb5c9513bac4fe]")

	server_check(server_stdout, backend_stdout, serverInitSuccessful)
	server_check(server_stdout, backend_stdout, serverReady)
	server_check(server_stdout, backend_stdout, serverDatacenter)
	server_check(server_stdout, backend_stdout, serverAutodetecting)
	server_check(server_stdout, backend_stdout, serverGoogleAutodetect)
	server_check(server_stdout, backend_stdout, serverAmazonAutodetect)
	server_check(server_stdout, backend_stdout, serverMultiplayAutodetect)
	server_check(server_stdout, backend_stdout, serverAutodetectFailed)

}

func test_server_ready_autodetect_multiplay_success() {

	fmt.Printf("test_server_ready_autodetect_multiplay_success\n")

	f, err := os.Create("whois.txt")
	if err != nil {
		panic(err)
	}
	f.WriteString("Internap\n")
	f.Close()

	serverConfig := &ServerConfig{}
	serverConfig.datacenter = "multiplay.newyork"
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 15)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverInitSuccessful := strings.Contains(server_stdout.String(), "info: welcome to network next :)")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverInputDatacenter := strings.Contains(server_stdout.String(), "info: server input datacenter is 'multiplay.newyork'")
	serverReadyDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'inap.newyork'")
	serverAutodetecting := strings.Contains(server_stdout.String(), "info: server attempting to autodetect datacenter")
	serverGoogleAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in google cloud")
	serverAmazonAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in AWS")
	serverFoundInap := strings.Contains(server_stdout.String(), "debug: found supplier inap")
	serverAutodetectSucceeded := strings.Contains(server_stdout.String(), "info: server autodetected datacenter 'inap.newyork' [323c61af696bff50]")

	os.Remove("whois.txt")

	server_check(server_stdout, backend_stdout, serverInitSuccessful)
	server_check(server_stdout, backend_stdout, serverReady)
	server_check(server_stdout, backend_stdout, serverInputDatacenter)
	server_check(server_stdout, backend_stdout, serverReadyDatacenter)
	server_check(server_stdout, backend_stdout, serverAutodetecting)
	server_check(server_stdout, backend_stdout, serverGoogleAutodetect)
	server_check(server_stdout, backend_stdout, serverAmazonAutodetect)
	server_check(server_stdout, backend_stdout, serverFoundInap)
	server_check(server_stdout, backend_stdout, serverAutodetectSucceeded)

}

func test_server_ready_autodetect_multiplay_fail() {

	fmt.Printf("test_server_ready_autodetect_multiplay_fail\n")

	os.Remove("whois.txt")

	serverConfig := &ServerConfig{}
	serverConfig.datacenter = "multiplay.newyork"
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 15)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverInitSuccessful := strings.Contains(server_stdout.String(), "info: welcome to network next :)")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'multiplay.newyork'")
	serverAutodetecting := strings.Contains(server_stdout.String(), "info: server attempting to autodetect datacenter")
	serverGoogleAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in google cloud")
	serverAmazonAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in AWS")
	serverMultiplayAutodetect := strings.Contains(server_stdout.String(), "info: could not autodetect multiplay datacenter :(")
	serverAutodetectFailed := strings.Contains(server_stdout.String(), "info: server autodetect datacenter failed. sticking with 'multiplay.newyork' [c38c805d821cd0d3]")

	serverCachedWhois := true
	if _, err := os.Stat("whois.txt"); errors.Is(err, os.ErrNotExist) {
		serverCachedWhois = false
	}

	os.Remove("whois.txt")

	server_check(server_stdout, backend_stdout, serverInitSuccessful)
	server_check(server_stdout, backend_stdout, serverReady)
	server_check(server_stdout, backend_stdout, serverDatacenter)
	server_check(server_stdout, backend_stdout, serverAutodetecting)
	server_check(server_stdout, backend_stdout, serverGoogleAutodetect)
	server_check(server_stdout, backend_stdout, serverAmazonAutodetect)
	server_check(server_stdout, backend_stdout, serverMultiplayAutodetect)
	server_check(server_stdout, backend_stdout, serverAutodetectFailed)
	server_check(server_stdout, backend_stdout, serverCachedWhois)

}

func test_server_ready_disable_autodetect_cloud() {

	fmt.Printf("test_server_ready_disable_autodetect_cloud\n")

	serverConfig := &ServerConfig{}
	serverConfig.datacenter = "cloud"
	serverConfig.disable_autodetect = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 15)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverInitSuccessful := strings.Contains(server_stdout.String(), "info: welcome to network next :)")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'cloud'")
	serverAutodetecting := strings.Contains(server_stdout.String(), "info: server attempting to autodetect datacenter")
	serverGoogleAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in google cloud")
	serverAmazonAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in AWS")
	serverMultiplayAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in multiplay")
	serverAutodetectFailed := strings.Contains(server_stdout.String(), "info: server autodetect datacenter failed. sticking with 'cloud' [9ebb5c9513bac4fe]")

	server_check(server_stdout, backend_stdout, serverInitSuccessful)
	server_check(server_stdout, backend_stdout, serverReady)
	server_check(server_stdout, backend_stdout, serverDatacenter)
	server_check(server_stdout, backend_stdout, !serverAutodetecting)
	server_check(server_stdout, backend_stdout, !serverGoogleAutodetect)
	server_check(server_stdout, backend_stdout, !serverAmazonAutodetect)
	server_check(server_stdout, backend_stdout, !serverMultiplayAutodetect)
	server_check(server_stdout, backend_stdout, !serverAutodetectFailed)

}

func test_server_ready_disable_autodetect_multiplay() {

	fmt.Printf("test_server_ready_disable_autodetect_multiplay\n")

	serverConfig := &ServerConfig{}
	serverConfig.datacenter = "multiplay.newyork"
	serverConfig.disable_autodetect = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 15)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverInitSuccessful := strings.Contains(server_stdout.String(), "info: welcome to network next :)")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'multiplay.newyork'")
	serverAutodetecting := strings.Contains(server_stdout.String(), "info: server attempting to autodetect datacenter")
	serverGoogleAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in google cloud")
	serverAmazonAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in AWS")
	serverMultiplayAutodetect := strings.Contains(server_stdout.String(), "info: server autodetect datacenter: not in multiplay")
	serverAutodetectFailed := strings.Contains(server_stdout.String(), "info: server autodetect datacenter failed. sticking with 'cloud' [9ebb5c9513bac4fe]")

	server_check(server_stdout, backend_stdout, serverInitSuccessful)
	server_check(server_stdout, backend_stdout, serverReady)
	server_check(server_stdout, backend_stdout, serverDatacenter)
	server_check(server_stdout, backend_stdout, !serverAutodetecting)
	server_check(server_stdout, backend_stdout, !serverGoogleAutodetect)
	server_check(server_stdout, backend_stdout, !serverAmazonAutodetect)
	server_check(server_stdout, backend_stdout, !serverMultiplayAutodetect)
	server_check(server_stdout, backend_stdout, !serverAutodetectFailed)

}

func test_server_ready_resolve_hostname_timeout() {

	fmt.Printf("test_server_ready_resolve_hostname_timeout\n")

	serverConfig := &ServerConfig{}
	serverConfig.datacenter = "local"
	serverConfig.force_resolve_hostname_timeout = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 15)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverFallbackToDirect := strings.Contains(server_stdout.String(), "info: server init timed out. falling back to direct mode only :(")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'local'")
	serverTimedOutHostnameResolve := strings.Contains(server_stdout.String(), "info: resolve hostname timed out")

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
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("DEFAULT")

	time.Sleep(time.Second * 15)

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	serverFallbackToDirect := strings.Contains(server_stdout.String(), "info: server init timed out. falling back to direct mode only :(")
	serverReady := strings.Contains(server_stdout.String(), "info: server is ready to receive client connections")
	serverDatacenter := strings.Contains(server_stdout.String(), "info: server datacenter is 'local'")
	serverTimedOutAutodetect := strings.Contains(server_stdout.String(), "info: autodetect timed out. sticking with 'local' [249f1fb6f3a680e8]")

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
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.force_resolve_hostname_timeout = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

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
	serverTimedOutHostnameResolve := strings.Contains(server_stdout.String(), "info: resolve hostname timed out")

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_server_events() {

	fmt.Printf("test_server_events\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.server_events = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

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

	backendSawServerEvents := strings.Contains(backend_stdout.String(), "server events 40100400")

	serverFlushedServerEvents := strings.Contains(server_stdout.String(), "server flushed events 40100400 to backend")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerEvents)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverFlushedServerEvents)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_match_id() {

	fmt.Printf("test_match_id\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.match_data = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("MATCH_ID")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawMatchID := strings.Contains(backend_stdout.String(), "match id d5f5127019cac4e5")

	serverAddsMatchData := strings.Contains(server_stdout.String(), "server adds match data")
	serverSawMatchDataRequest := strings.Contains(server_stdout.String(), "server sent match data packet")
	serverSawMatchDataResponse := strings.Contains(server_stdout.String(), "server successfully recorded match data")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverAddsMatchData)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawMatchDataRequest)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawMatchDataResponse)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawMatchID)

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

func test_match_values() {

	fmt.Printf("test_match_values\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.match_data = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("MATCH_VALUES")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawMatchValue1 := strings.Contains(backend_stdout.String(), "match value 10.10")
	backendSawMatchValue2 := strings.Contains(backend_stdout.String(), "match value 20.20")
	backendSawMatchValue3 := strings.Contains(backend_stdout.String(), "match value 30.30")

	serverAddsMatchData := strings.Contains(server_stdout.String(), "server adds match data")
	serverSawMatchDataRequest := strings.Contains(server_stdout.String(), "server sent match data packet")
	serverSawMatchDataResponse := strings.Contains(server_stdout.String(), "server successfully recorded match data")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawMatchValue1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawMatchValue2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawMatchValue3)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverAddsMatchData)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawMatchDataRequest)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawMatchDataResponse)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_match_data_retry() {

	fmt.Printf("test_match_data_retry\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.match_data = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("FORCE_RETRY")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	backend_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()
	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawMatchID := strings.Contains(backend_stdout.String(), "match id d5f5127019cac4e5")
	backendSawMatchValue1 := strings.Contains(backend_stdout.String(), "match value 10.10")
	backendSawMatchValue2 := strings.Contains(backend_stdout.String(), "match value 20.20")
	backendSawMatchValue3 := strings.Contains(backend_stdout.String(), "match value 30.30")

	serverAddsMatchData := strings.Contains(server_stdout.String(), "server adds match data")
	serverSawMatchDataRequest := strings.Contains(server_stdout.String(), "server sent match data packet")
	serverSawMatchDataResponse := strings.Contains(server_stdout.String(), "server successfully recorded match data")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawMatchID)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawMatchValue1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawMatchValue2)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawMatchValue3)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverAddsMatchData)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawMatchDataRequest)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawMatchDataResponse)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_flush() {

	fmt.Printf("test_flush\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.flush = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_flush_retry() {

	fmt.Printf("test_flush_retry\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.flush = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 40*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_flush_server_events_and_match_data() {

	fmt.Printf("test_flush_server_events_and_match_data\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 50.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.server_events = true
	serverConfig.match_data = true
	serverConfig.flush = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("MATCH_ID")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()

	backend_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawMatchID := strings.Contains(backend_stdout.String(), "match id d5f5127019cac4e5")
	backendSawSessionUpdate := strings.Contains(backend_stdout.String(), "client ping timed out")
	backendSawServerEvents := strings.Contains(backend_stdout.String(), "server events 40100400")

	serverFlushedServerEvents := strings.Contains(server_stdout.String(), "server flushed events 40100400 to backend")
	serverAddsMatchData := strings.Contains(server_stdout.String(), "server adds match data")
	serverSawFlushRequest := strings.Contains(server_stdout.String(), "server flush started")
	serverSawSessionUpdateFlush := strings.Contains(server_stdout.String(), "server flushed session update")
	serverSawMatchDataFlush := strings.Contains(server_stdout.String(), "server flushed match data")
	serverSawFlushComplete := strings.Contains(server_stdout.String(), "server flush finished")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawMatchID)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawSessionUpdate)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerEvents)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverFlushedServerEvents)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverAddsMatchData)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushRequest)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawSessionUpdateFlush)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawMatchDataFlush)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushComplete)
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT] == 0)

}

func test_flush_server_events_and_match_data_retry() {

	fmt.Printf("test_flush_server_events_and_match_data_retry\n")

	clientConfig := &ClientConfig{}
	clientConfig.stop_sending_packets_time = 30.0
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.server_events = true
	serverConfig.match_data = true
	serverConfig.flush = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	backend_cmd, backend_stdout := backend("MATCH_ID")

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)

	server_cmd.Wait()

	backend_cmd.Process.Signal(os.Interrupt)

	backend_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	backendSawMatchID := strings.Contains(backend_stdout.String(), "match id d5f5127019cac4e5")
	backendSawSessionUpdate := strings.Contains(backend_stdout.String(), "client ping timed out")
	backendSawServerEvents := strings.Contains(backend_stdout.String(), "server events 40100400")

	serverFlushedServerEvents := strings.Contains(server_stdout.String(), "server flushed events 40100400 to backend")
	serverAddsMatchData := strings.Contains(server_stdout.String(), "server adds match data")
	serverSawFlushRequest := strings.Contains(server_stdout.String(), "server flush started")
	serverSawSessionUpdateFlush := strings.Contains(server_stdout.String(), "server flushed session update")
	serverSawMatchDataFlush := strings.Contains(server_stdout.String(), "server flushed match data")
	serverSawFlushComplete := strings.Contains(server_stdout.String(), "server flush finished")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawMatchID)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawSessionUpdate)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerEvents)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverFlushedServerEvents)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverAddsMatchData)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushRequest)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawSessionUpdateFlush)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawMatchDataFlush)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, serverSawFlushComplete)

}

type test_function func()

func main() {
	allTests := []test_function{
		test_passthrough,
		test_direct_upgraded,
		test_network_next_route,
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
		test_multipath,
		test_multipath_next_packet_loss,
		test_multipath_fallback_to_direct,
		test_uncommitted,
		test_uncommitted_to_committed,
		test_packet_loss_direct,
		test_packet_loss_next,
		test_server_under_load,
		test_session_update_retry,
		test_bandwidth_over_limit,
		test_packet_loss,
		test_bandwidth,
		test_jitter,
		test_tags,
		test_tags_multi,
		test_direct_stats,
		test_next_stats,
		test_report_session,
		test_client_ping_timed_out,
		test_server_ready_success,
 		test_server_ready_fallback_to_direct,
 		test_server_ready_autodetect_cloud,
 		test_server_ready_autodetect_multiplay_success,
 		test_server_ready_autodetect_multiplay_fail,
 		test_server_ready_disable_autodetect_cloud,
		test_server_ready_disable_autodetect_multiplay,
		test_server_ready_resolve_hostname_timeout,
		test_server_ready_autodetect_timeout,
		test_client_connect_before_ready,
		test_server_events,
		test_match_id,
		test_match_values,
		test_match_data_retry,
		test_flush,
		test_flush_retry,
		test_flush_server_events_and_match_data,
		test_flush_server_events_and_match_data_retry,
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
		time.Sleep(time.Duration(len(tests)*80) * time.Second)
		panic("tests took too long!")
	}()

	for i := range tests {
		tests[i]()
	}
}
