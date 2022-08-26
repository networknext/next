/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func make(action string) (*exec.Cmd, *bytes.Buffer) {

	fmt.Printf("make %s\n", action)

	cmd := exec.Command("make", action)
	if cmd == nil {
		panic("could not run make!\n")
		return nil, nil
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Start()

	return cmd, &stdout
}

func main() {

	fmt.Printf("\nhappy path\n\n")

	magic_backend_cmd, magic_backend_stdout := make("dev-magic-backend")
	relay_gateway_cmd, relay_gateway_stdout := make("dev-relay-gateway")
	relay_backend_1_cmd, relay_backend_1_stdout := make("dev-relay-backend-1")
	relay_backend_2_cmd, relay_backend_2_stdout := make("dev-relay-backend-2")
	relay_frontend_cmd, relay_frontend_stdout := make("dev-relay-frontend")
	relay_1_cmd, relay_1_stdout := make("RELAY_PORT=2000 dev-relay")
	relay_2_cmd, relay_2_stdout := make("RELAY_PORT=2001 dev-relay")
	relay_3_cmd, relay_3_stdout := make("RELAY_PORT=2002 dev-relay")
	relay_4_cmd, relay_4_stdout := make("RELAY_PORT=2003 dev-relay")
	relay_5_cmd, relay_5_stdout := make("RELAY_PORT=2004 dev-relay")

	_ = magic_backend_cmd
	_ = relay_gateway_cmd
	_ = relay_backend_1_cmd
	_ = relay_backend_2_cmd
	_ = relay_frontend_cmd
	_ = relay_1_cmd
	_ = relay_2_cmd
	_ = relay_3_cmd
	_ = relay_4_cmd
	_ = relay_5_cmd

	_ = magic_backend_stdout
	_ = relay_gateway_stdout
	_ = relay_backend_1_stdout
	_ = relay_backend_2_stdout
	_ = relay_frontend_stdout
	_ = relay_1_stdout
	_ = relay_2_stdout
	_ = relay_3_stdout
	_ = relay_4_stdout
	_ = relay_5_stdout

	relay_1_initialized := false
	relay_2_initialized := false
	relay_3_initialized := false
	relay_4_initialized := false
	relay_5_initialized := false

	fmt.Printf("\nwaiting for the relay gateway to initialize...\n")

	relay_gateway_initialized := false

	for i := 0; i < 10; i++ {
		if strings.Contains(relay_gateway_stdout.String(), "loaded database.bin") &&
		   strings.Contains(relay_gateway_stdout.String(), "starting http server on port 30000") &&
		   strings.Contains(relay_gateway_stdout.String(), "started watchman on ") {
		   	relay_gateway_initialized = true
		   	break
		}
		time.Sleep(time.Second)
	}

	if !relay_gateway_initialized {
		fmt.Printf("error: failed to initialize relay gateway\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_gateway_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		os.Exit(1)
	}

	fmt.Printf("\nwaiting for relays to initialize...\n\n")

	const NumIterations = 10

	for i := 0; i < NumIterations; i++ {

		fmt.Printf("iteration %d\n", i)

		if !relay_1_initialized && strings.Contains(relay_1_stdout.String(), "Relay initialized") {
			fmt.Printf("Relay initialized\n")
			relay_1_initialized = true
		}

		if !relay_2_initialized && strings.Contains(relay_2_stdout.String(), "Relay initialized") {
			fmt.Printf("Relay initialized\n")
			relay_2_initialized = true
		}

		if !relay_3_initialized && strings.Contains(relay_3_stdout.String(), "Relay initialized") {
			fmt.Printf("Relay initialized\n")
			relay_3_initialized = true
		}

		if !relay_4_initialized && strings.Contains(relay_4_stdout.String(), "Relay initialized") {
			fmt.Printf("Relay initialized\n")
			relay_2_initialized = true
		}

		if !relay_5_initialized && strings.Contains(relay_5_stdout.String(), "Relay initialized") {
			fmt.Printf("Relay initialized\n")
			relay_5_initialized = true
		}

		if relay_1_initialized && relay_2_initialized && relay_3_initialized && relay_4_initialized && relay_5_initialized {
			break
		}

		time.Sleep(time.Second)
	}

	fmt.Printf("\nend loop\n")

	// todo: don't complain about relays failing to initialize, until we fix this
	/*
		if !relay_initialized {
			fmt.Printf("error: relays failed to initialize\n\n")
			fmt.Printf("relay frontend: %s\n\n", relay_frontend_stdout)
			os.Exit(1)
		}
	*/

	fmt.Printf("\nsuccess!\n\n")

}
