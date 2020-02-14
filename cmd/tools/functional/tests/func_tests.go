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
	"strconv"
	"strings"
	"time"
)

const (
	relayBin   = "./dist/relay"
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

func relay() (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command(relayBin)
	if cmd == nil {
		return nil, nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RELAYDEV=1")
	cmd.Env = append(cmd.Env, "RELAYMASTER=127.0.0.1")
	cmd.Env = append(cmd.Env, "RELAYPORT=0")
	cmd.Env = append(cmd.Env, "RELAYNAME=local")
	cmd.Env = append(cmd.Env, "RELAYUPDATEKEY=eyqNheTBdx+97qd3Nkf/QvjaSDQVQQzHvkhX6w9cvMO276rgKZ7VIPHwaoNE7f9SiQW6yThhEC5onwpBEFFdaw==")
	cmd.Env = append(cmd.Env, "RELAYPUBLICKEY=9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=")
	cmd.Env = append(cmd.Env, "RELAYPRIVATEKEY=lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=")
	cmd.Env = append(cmd.Env, "RELAYADDRESS=127.0.0.1")

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Start()

	return cmd, &output
}

type ClientConfig struct {
	duration                   int
	customer_public_key        string
	disable_network_next       bool
	user_flags                 bool
	packet_loss                bool
	fake_direct_packet_loss    float32
	fake_direct_rtt            float32
	fake_next_packet_loss      float32
	fake_next_rtt              float32
	connect_time               float64
	connect_address            string
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
	"NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS",
	"NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS",
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

func client_check(client_counters []uint64, client_stdout *bytes.Buffer, server_stdout *bytes.Buffer, backend_stdout *bytes.Buffer, condition bool) {
	if !condition {
		if backend_stdout != nil {
			fmt.Printf("\n--------------------------------------------------------------------------\n")
			fmt.Printf("\n%s", backend_stdout)
		}
		fmt.Printf("\n--------------------------------------------------------------------------\n")
		fmt.Printf("\n%s", server_stdout)
		fmt.Printf("\n--------------------------------------------------------------------------\n")
		fmt.Printf("\n%s", client_stdout)
		fmt.Printf("\n--------------------------------------------------------------------------\n")
		fmt.Printf("\n")
		print_client_counters(client_counters)
		fmt.Printf("\n--------------------------------------------------------------------------\n")
		fmt.Printf("\n")
		panic("client check failed!")
	}
}

func test_direct_default() {

	fmt.Printf("test_direct_default\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 10.0

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	server_cmd, server_stdout := server(&ServerConfig{})

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	server_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 500)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 500)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

func test_direct_upgrade() {

	fmt.Printf("test_direct_upgrade\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 10.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	server_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 500)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 500)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

func test_direct_no_upgrade() {

	fmt.Printf("test_direct_no_upgrade\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 10.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.no_upgrade = true
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	server_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 500)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 500)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

func test_direct_with_backend() {

	fmt.Printf("test_direct_with_backend\n")

	clientConfig := &ClientConfig{}
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

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

func test_fallback_to_direct_without_backend() {

	fmt.Printf("test_fallback_to_direct_without_backend\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 40.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	server_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 2350)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 2350)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

func test_fallback_to_direct_is_not_sticky() {

	fmt.Printf("test_fallback_to_direct_is_not_sticky\n")

	clientConfig := &ClientConfig{}
	clientConfig.duration = 60.0
	clientConfig.customer_public_key = "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="
	clientConfig.connect_time = 45.0
	clientConfig.connect_address = "127.0.0.1:32202"

	client_cmd, client_stdout, client_stderr := client(clientConfig)

	serverConfig := &ServerConfig{}
	serverConfig.customer_private_key = "leN7D7+9vr3TEZexVmvbYzdH1hbpwBvioc6y1c9Dhwr4ZaTkEWyX2Li5Ph/UFrw8QS8hAD9SQZkuVP6x14tEcqxWppmrvbdn"

	server_cmd, server_stdout := server(serverConfig)

	client_cmd.Wait()

	server_cmd.Process.Signal(os.Interrupt)
	server_cmd.Wait()

	client_counters := read_client_counters(client_stderr.String())

	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 2)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 1)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, nil, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

func test_packets_over_next_with_relay_and_backend() {

	fmt.Printf("test_packets_over_next_with_relay_and_backend\n")

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

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT])
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT])
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

func test_fallback_to_direct_when_backend_goes_down() {

	fmt.Printf("test_fallback_to_direct_when_backend_goes_down\n")

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

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 2500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

func test_network_next_disabled_server() {

	fmt.Printf("test_network_next_disabled_server\n")

	clientConfig := &ClientConfig{}
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

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

func test_network_next_disabled_client() {

	fmt.Printf("test_network_next_disabled_client\n")

	clientConfig := &ClientConfig{}
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

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

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

	relay_1_cmd, _ := relay()
	relay_2_cmd, _ := relay()
	relay_3_cmd, _ := relay()

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

		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
		client_check(client_counters, client_stdout[i], server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

	}
}

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

func test_route_switching() {

	fmt.Printf("test_route_switching\n")

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

func test_on_off() {

	fmt.Printf("test_on_off\n")

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

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 3300)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT]+client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 3300)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

func test_multipath() {

	fmt.Printf("test_multipath\n")

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

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 2000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 2000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)

}

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

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)
}

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

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 3500)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] == 0)
}

func test_packet_loss() {

	fmt.Printf("test_packet_loss\n")

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

	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_OPEN_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLOSE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_UPGRADE_SESSION] == 1)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_FALLBACK_TO_DIRECT] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] > 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_SENT_NEXT] + client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] > 3000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_DIRECT] + client_counters[NEXT_CLIENT_COUNTER_PACKET_RECEIVED_NEXT] + client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] > 3000)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_MULTIPATH] == 0)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_CLIENT_TO_SERVER_PACKET_LOSS] > 250)
	client_check(client_counters, client_stdout, server_stdout, backend_stdout, client_counters[NEXT_CLIENT_COUNTER_SERVER_TO_CLIENT_PACKET_LOSS] > 250)
}

type test_function func()

func main() {

	tests := []test_function{
		/*
		test_direct_default,
		test_direct_upgrade,
		test_direct_no_upgrade,
		test_direct_with_backend,
		test_fallback_to_direct_without_backend,
		test_fallback_to_direct_is_not_sticky,
		test_packets_over_next_with_relay_and_backend,
		test_fallback_to_direct_when_backend_goes_down,
		test_network_next_disabled_server,
		test_network_next_disabled_client,
		test_server_under_load,
		test_reconnect_direct,
		test_reconnect_next,
		test_connect_to_another_server_direct,
		test_connect_to_another_server_next,
		test_route_switching,
		test_on_off,
		test_multipath,
		test_uncommitted,
		test_uncommitted_to_committed,
		test_user_flags,
		*/
		test_packet_loss,
	}

	for {
		for i := range tests {
			tests[i]()
		}
	}
}
