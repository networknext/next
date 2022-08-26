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
	"syscall"
)

func make(action string) (*exec.Cmd, *bytes.Buffer) {

	// IMPORTANT: need to install "expect" package, eg. "brew install expect", "sudo apt install -y expect"
	// without this, the output from make is buffered and we can't read it reliabliy until the process finishes :(

	fmt.Printf("make %s\n", action)

	// cmd := exec.Command("unbuffer", "make", action)
	cmd := exec.Command("make", action)
	if cmd == nil {
		panic("could not run make!\n")
		return nil, nil
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stdout = &stdout
	cmd.Start()

	return cmd, &stdout
}

func happy_path() int {

	fmt.Printf("\nhappy path\n\n")

	// build and run services, as a develop would via "make dev-*"

	magic_backend_cmd, magic_backend_stdout := make("dev-magic-backend")
	relay_gateway_cmd, relay_gateway_stdout := make("dev-relay-gateway")
	relay_backend_1_cmd, relay_backend_1_stdout := make("dev-relay-backend-1")
	relay_backend_2_cmd, relay_backend_2_stdout := make("dev-relay-backend-2")
	relay_frontend_cmd, relay_frontend_stdout := make("dev-relay-frontend")
	relay_1_cmd, relay_1_stdout := make("dev-relay")
	relay_2_cmd, relay_2_stdout := make("RELAY_PORT=2001 dev-relay")
	relay_3_cmd, relay_3_stdout := make("RELAY_PORT=2002 dev-relay")
	relay_4_cmd, relay_4_stdout := make("RELAY_PORT=2003 dev-relay")
	relay_5_cmd, relay_5_stdout := make("RELAY_PORT=2004 dev-relay")

	relay_1_initialized := false
	relay_2_initialized := false
	relay_3_initialized := false
	relay_4_initialized := false
	relay_5_initialized := false

	// make sure everything gets cleaned up

	defer func() {

		magic_backend_cmd.Process.Signal(syscall.SIGTERM)
		relay_gateway_cmd.Process.Signal(syscall.SIGTERM)
		relay_backend_1_cmd.Process.Signal(syscall.SIGTERM)
		relay_backend_2_cmd.Process.Signal(syscall.SIGTERM)
		relay_frontend_cmd.Process.Signal(syscall.SIGTERM)
		relay_1_cmd.Process.Signal(syscall.SIGTERM)

		magic_backend_cmd.Wait()
		relay_gateway_cmd.Wait()
		relay_backend_1_cmd.Wait()
		relay_backend_2_cmd.Wait()
		relay_frontend_cmd.Wait()
		relay_1_cmd.Wait()
		relay_2_cmd.Wait()
		relay_3_cmd.Wait()
		relay_4_cmd.Wait()
		relay_5_cmd.Wait()
	}()

	// initialize the magic backend

	fmt.Printf("\ninitializing magic backend\n")

	magic_backend_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(magic_backend_stdout.String(), "starting http server on port 41007") &&
		   strings.Contains(magic_backend_stdout.String(), "served magic values") {
		   	magic_backend_initialized = true
		   	break
		}
		time.Sleep(100*time.Millisecond)
	}

	if !magic_backend_initialized {
		fmt.Printf("error: failed to initialize magic backend\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", magic_backend_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	// initialize relay gateway

	fmt.Printf("initializing relay gateway\n")

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
		return 1
	}

	// initialize relay backend 1

	relay_backend_1_initialized := false

	fmt.Printf("initializing relay backend 1\n")

	for i := 0; i < 100; i++ {
		if strings.Contains(relay_backend_1_stdout.String(), "starting http server on port 30001") &&
		   strings.Contains(relay_backend_1_stdout.String(), "started watchman on ") && 
		   strings.Contains(relay_backend_1_stdout.String(), "wrote route matrix to redis") {
		   	relay_backend_1_initialized = true
		   	break
		}
		time.Sleep(100*time.Millisecond)
	}

	if !relay_backend_1_initialized {
		fmt.Printf("error: failed to initialize relay backend 1\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_backend_1_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	// initialize relay backend 2

	relay_backend_2_initialized := false

	fmt.Printf("initializing relay backend 2\n")

	for i := 0; i < 100; i++ {
		if strings.Contains(relay_backend_2_stdout.String(), "starting http server on port 30002") &&
		   strings.Contains(relay_backend_2_stdout.String(), "started watchman on ") && 
		   strings.Contains(relay_backend_2_stdout.String(), "wrote route matrix to redis") {
		   	relay_backend_2_initialized = true
		   	break
		}
		time.Sleep(100*time.Millisecond)
	}

	if !relay_backend_2_initialized {
		fmt.Printf("error: failed to initialize relay backend 2\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_backend_2_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	// initialize relay frontend

	relay_frontend_initialized := false

	fmt.Printf("initializing relay frontend\n")

	for i := 0; i < 100; i++ {
		if strings.Contains(relay_frontend_stdout.String(), "starting http server on port 30005") {
		   	relay_frontend_initialized = true
		   	break
		}
		time.Sleep(100*time.Millisecond)
	}

	if !relay_frontend_initialized {
		fmt.Printf("error: failed to initialize relay frontend\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_frontend_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	// initialize relays

	fmt.Printf("initializing relays\n")

	relays_initialized := false

	for i := 0; i < 100; i++ {

		if !relay_1_initialized && strings.Contains(relay_1_stdout.String(), "Relay initialized") {
			relay_1_initialized = true
		}

		if !relay_2_initialized && strings.Contains(relay_2_stdout.String(), "Relay initialized") {
			relay_2_initialized = true
		}

		if !relay_3_initialized && strings.Contains(relay_3_stdout.String(), "Relay initialized") {
			relay_3_initialized = true
		}

		if !relay_4_initialized && strings.Contains(relay_4_stdout.String(), "Relay initialized") {
			relay_2_initialized = true
		}

		if !relay_5_initialized && strings.Contains(relay_5_stdout.String(), "Relay initialized") {
			relay_5_initialized = true
		}

		if relay_1_initialized && relay_2_initialized && relay_3_initialized && relay_4_initialized && relay_5_initialized {
			break
		}

		if relay_1_initialized {
			relays_initialized = true
			break
		}

		time.Sleep(100*time.Millisecond)
	}

	if !relays_initialized {
		fmt.Printf("error: relays failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_gateway_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_1_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_2_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_3_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_4_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_5_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf("\nsuccess!\n\n")

	return 0
}

func main() {
	if happy_path() != 0 {
		os.Exit(1)
	}
}
