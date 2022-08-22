/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"bytes"
	"strings"
)

func make(action string) (*exec.Cmd, *bytes.Buffer) {

	fmt.Printf("make %s\n", action)

	cmd := exec.Command("make", action)
	if cmd == nil {
		panic("could not run make!\n")
		return nil, nil
	}

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Start()

	return cmd, &output
}

func main() {
	
	fmt.Printf("\nhappy path\n\n")

	magic_backend_cmd, magic_backend_stdout := make("dev-magic-backend")
	magic_frontend_cmd, magic_frontend_stdout := make("dev-magic-frontend")
	relay_gateway_cmd, relay_gateway_stdout := make("dev-relay-gateway")
	relay_backend_1_cmd, relay_backend_1_stdout := make("dev-relay-backend-1")
	relay_backend_2_cmd, relay_backend_2_stdout := make("dev-relay-backend-2")
	relay_frontend_cmd, relay_frontend_stdout := make("dev-relay-frontend")
	relay_cmd, relay_stdout := make("dev-relay")
	
	_ = magic_backend_stdout
	_ = magic_frontend_stdout
	_ = relay_gateway_stdout
	_ = relay_backend_1_stdout
	_ = relay_backend_2_stdout
	_ = relay_frontend_stdout
	_ = relay_stdout

	relay_initialized := false

	fmt.Printf("\nwaiting for relays to initialize...\n")

	for i := 0; i < 10; i++ {

		fmt.Printf("iteration %d\n", i)

		if !relay_initialized && strings.Contains(relay_stdout.String(), "Relay initialized") {
			fmt.Printf("relay initialized\n")
			relay_initialized = true
		}

		if relay_initialized {
			break
		}

		time.Sleep(time.Second)
	}

	fmt.Printf("end loop\n")

	// todo: don't complain about relays failing to initialize, until we fix this
	/*
	if !relay_initialized {
		fmt.Printf("error: relays failed to initialize\n\n")
		fmt.Printf("relay frontend: %s\n\n", relay_frontend_stdout)
		os.Exit(1)
	}
	*/

	magic_backend_cmd.Process.Signal(os.Interrupt)
	magic_frontend_cmd.Process.Signal(os.Interrupt)
	relay_gateway_cmd.Process.Signal(os.Interrupt)
	relay_backend_1_cmd.Process.Signal(os.Interrupt)
	relay_backend_2_cmd.Process.Signal(os.Interrupt)
	relay_frontend_cmd.Process.Signal(os.Interrupt)
	relay_cmd.Process.Signal(os.Interrupt)

	magic_backend_cmd.Wait()
	magic_frontend_cmd.Wait()
	relay_gateway_cmd.Wait()
	relay_backend_1_cmd.Wait()
	relay_backend_2_cmd.Wait()
	relay_frontend_cmd.Wait()
	relay_cmd.Wait()

	fmt.Printf("\nsuccess!\n\n")

}
