/*
   Network Next Accelerate.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
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

	relay_cmd, relay_stdout := relay("relay", 2000)

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

	relay_cmd, relay_stdout := relay("relay", 2000)

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

// fmt.Printf("=======================================\n%s=============================================\n", relay_stdout)

type test_function func()

func main() {

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
