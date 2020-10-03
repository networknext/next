/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
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
	backendBin = "./dist/func_backend"
	clientBin  = "./dist/func_client"
	serverBin  = "./dist/func_server"
)

func backend(mode string) (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command(backendBin)
	if cmd == nil {
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
	duration                	int
	customer_public_key     	string
	disable_network_next    	bool
	user_flags              	bool
	packet_loss             	bool
	fake_direct_packet_loss 	float32
	fake_direct_rtt         	float32
	fake_next_packet_loss   	float32
	fake_next_rtt           	float32
	connect_time            	float64
	connect_address         	string
	stop_sending_packets_time 	float64
	fallback_to_direct_time     float64
}

func client(config *ClientConfig) (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {

	cmd := exec.Command(clientBin)
	if cmd == nil {
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

	if config.user_flags {
		cmd.Env = append(cmd.Env, "CLIENT_USER_FLAGS=1")
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

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Start()

	return cmd, &stdout, &stderr
}

type ServerConfig struct {
	duration             int
	no_upgrade           bool
	packet_loss          bool
	customer_private_key string
	disable_network_next bool
	server_address       string
	server_port          int
}

func server(config *ServerConfig) (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command(serverBin)
	if cmd == nil {
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
const NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT = 4
const NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT = 5
const NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT = 6
const NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT = 7
const NEXT_CLIENT_COUNTER_MULTIPATH = 8
const NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS = 9
const NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS = 10
const NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_CLIENT_TO_SERVER = 11
const NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_SERVER_TO_CLIENT = 12
const NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_RAW = 13
const NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_UPGRADED = 14
const NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT_RAW = 15
const NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT_UPGRADED = 16

var ClientCounterNames = []string{
	"NEXT_CLIENT_COUNTER_OPEN_SESSION",
	"NEXT_CLIENT_COUNTER_CLOSE_SESSION",
	"NEXT_CLIENT_COUNTER_UPGRADE_SESSION",
	"NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT",
	"NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT",
	"NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT",
	"NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT",
	"NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT",
	"NEXT_CLIENT_COUNTER_MULTIPATH",
	"NEXT_CLIENT_COUNTER_PACKETS_LOST_CLIENT_TO_SERVER",
	"NEXT_CLIENT_COUNTER_PACKETS_LOST_SERVER_TO_CLIENT",
	"NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_CLIENT_TO_SERVER",
	"NEXT_CLIENT_COUNTER_PACKETS_OUT_OF_ORDER_SERVER_TO_CLIENT",
    "NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_RAW",
	"NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_UPGRADED",
	"NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT_RAW",
	"NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT_UPGRADED",
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

/*
	Test that when a client connects to a server with no backend running, and with no customer public or private
	keys set on either client and server, that packets are sent and received direct. This is network next in direct mode.
*/

func test_direct_raw() {

	fmt.Printf("test_direct_raw\n")

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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, nil, totalPacketsSent >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, nil, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] == client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_RAW])
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_UPGRADED] == 0)

	// todo: make sure raw direct packets are sent and received

	// todo: make sure we don't have upgraded direct packets

}

/*
	Run a backend but no relays. Make sure that we send and receive all packets direct.
	This tests the path where we prefix upgraded session direct packets with [255][sequence]
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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_RAW] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_UPGRADED] == client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT])
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_UPGRADED] >= 40*60)

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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)
   
}

func test_fallback_to_direct_client_side() {

	fmt.Printf("fallback_to_direct_client_side\n")

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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)

}

/*
	Have network next enabled on a client, but disable it on a server.
	Verify that the client is still able to connect to the server, but all packets are sent direct.
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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_RAW] >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT_RAW] >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_UPGRADED] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT_UPGRADED] == 0)

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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_RAW] >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT_RAW] >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT_UPGRADED] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT_UPGRADED] == 0)

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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

	backendSawClientBandwidthOverLimit := strings.Contains(backend_stdout.String(), "client bandwidth over limit")
	backendSawServerBandwidthOverLimit := strings.Contains(backend_stdout.String(), "server bandwidth over limit")

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawClientBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, backendSawServerBandwidthOverLimit == false)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

/*
	Test that a client is able to connect to a server direct, and then reconnect to the same server without problems.
	This verifies that our code in the SDK to distinguish the old session from the new one is working properly for
	upgraded direct packets (255 prefix).
*/

func test_reconnect_direct() {

	fmt.Printf("test_reconnect_direct\n")

	clientConfig := &ClientConfig{}
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

/*
	User flags are a new feature in SDK 3.4.0.
	Verify that user flags set on the client get plumbed all the way up to the backend.
*/

func test_user_flags() {

	fmt.Printf("test_user_flags\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.user_flags = true
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

	backend_cmd, backend_stdout := backend("USER_FLAGS")

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

/*
	Simulate packet loss between the client, server and backend.
	Make sure we can still get a direct route.
*/

func test_packet_loss_direct() {

	fmt.Printf("test_packet_loss_direct\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.user_flags = true
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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] > 250)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] > 250)

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]+client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]+client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] > 250)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] > 250)

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
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)

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

	totalPacketsSent := client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT]
	totalPacketsReceived := client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT]

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
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsSent >= 50*60)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, totalPacketsReceived == totalPacketsSent)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0, relay_1_stdout, relay_2_stdout, relay_3_stdout)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] >= 30*60, relay_1_stdout, relay_2_stdout, relay_3_stdout)

}

type test_function func()

func main() {
	allTests := []test_function{
		test_direct_raw,
		test_direct_upgraded,
		test_network_next_route,
		test_fallback_to_direct_backend,
		test_fallback_to_direct_client_side,
		test_disable_on_server,
		test_disable_on_client,
		test_route_switching,
		test_on_off,
		test_on_on_off,
		test_reconnect_direct,
		test_reconnect_next,
		test_connect_to_another_server_direct,
		test_connect_to_another_server_next,
		test_multipath,
		test_multipath_next_packet_loss,
		test_multipath_fallback_to_direct,
		test_uncommitted,
		test_uncommitted_to_committed,
		test_user_flags,
		test_packet_loss_direct,
		test_packet_loss_next,
		test_server_under_load,
		test_session_update_retry,
	}

	// If there are command line arguments, use reflection to see what tests to run
	var tests []test_function
	prefix := "main."
	if len(os.Args) > 1 {
		for _, funcName := range os.Args[1:] {
			for _, test := range allTests {
				name := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
				name = name[len(prefix):]
				if funcName == name {
					tests = append(tests, test)
				}
			}
		}
	} else {
		tests = allTests // No command line args, run all tests
	}

	for {
		for i := range tests {
			tests[i]()
		}
	}

}

// todo: check that no bandwidth is over limit in regular tests (eg. network next route tests)

// todo: add test for bandwidth over limit

// todo: add test to verify packet loss tracker results gets up to backend

// todo: add test to verify jitter tracker results get up to backend

// todo: add test to verify out of order tracker results get up to backend

// todo: verify backend sees when client exceeds bandwidth envelope up/down (separate counters)
