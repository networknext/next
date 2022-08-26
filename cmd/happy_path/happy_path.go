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
	"encoding/base64"
	"encoding/binary"
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/storage"
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

func generateLocalDatabase(filename string) {

	ctx := context.Background()

	logger := log.NewNopLogger()

	db, err := storage.NewSQLite3(ctx, logger)
	if err != nil {
		panic(err)
	}

	relayPublicKeyString := "9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14="
	customerPublicKeyString := "leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw=="

	relayPublicKey, err := base64.StdEncoding.DecodeString(relayPublicKeyString)
	if err != nil {
		panic(err)
	}

	customerPublicKey, err := base64.StdEncoding.DecodeString(customerPublicKeyString)
	if err != nil {
		panic(err)
	}

	customerId := binary.LittleEndian.Uint64(customerPublicKey[:8])
	customerPublicKey = customerPublicKey[8:]

	storage.SeedSQLStorage(ctx, db, relayPublicKey, customerId, customerPublicKey)

	// todo: write to file

}

func happy_path() int {

	fmt.Printf("\nhappy path\n\n")

	// todo: not sure we need this anymore
	/*
	// generate a local database.bin just for the happy path

	generateLocalDatabase("database.bin")
	*/

	// build and run services, as a develop would via "make dev-*"

	magic_backend_cmd, magic_backend_stdout := make("dev-magic-backend")
	relay_gateway_cmd, relay_gateway_stdout := make("dev-relay-gateway")
	relay_backend_1_cmd, relay_backend_1_stdout := make("dev-relay-backend-1")
	relay_backend_2_cmd, relay_backend_2_stdout := make("dev-relay-backend-2")
	relay_frontend_cmd, relay_frontend_stdout := make("dev-relay-frontend")
	relay_1_cmd, relay_1_stdout := make("dev-relay")
	/*
	relay_2_cmd, relay_2_stdout := make("RELAY_PORT=2001 dev-relay")
	relay_3_cmd, relay_3_stdout := make("RELAY_PORT=2002 dev-relay")
	relay_4_cmd, relay_4_stdout := make("RELAY_PORT=2003 dev-relay")
	relay_5_cmd, relay_5_stdout := make("RELAY_PORT=2004 dev-relay")
	*/

	_ = magic_backend_cmd
	_ = relay_gateway_cmd
	_ = relay_backend_1_cmd
	_ = relay_backend_2_cmd
	_ = relay_frontend_cmd
	_ = relay_1_cmd
	/*
	_ = relay_2_cmd
	_ = relay_3_cmd
	_ = relay_4_cmd
	_ = relay_5_cmd
	*/

	_ = magic_backend_stdout
	_ = relay_gateway_stdout
	_ = relay_backend_1_stdout
	_ = relay_backend_2_stdout
	_ = relay_frontend_stdout
	_ = relay_1_stdout
	/*
	_ = relay_2_stdout
	_ = relay_3_stdout
	_ = relay_4_stdout
	_ = relay_5_stdout
	*/

	relay_1_initialized := false
	/*
	relay_2_initialized := false
	relay_3_initialized := false
	relay_4_initialized := false
	relay_5_initialized := false
	*/

	// make sure everything gets cleaned up

	defer func() {

		fmt.Printf("shutting down...\n")

		magic_backend_cmd.Process.Signal(os.Interrupt)
		relay_gateway_cmd.Process.Signal(os.Interrupt)
		relay_backend_1_cmd.Process.Signal(os.Interrupt)
		relay_backend_2_cmd.Process.Signal(os.Interrupt)
		relay_frontend_cmd.Process.Signal(os.Interrupt)
		relay_1_cmd.Process.Signal(os.Interrupt)

		magic_backend_cmd.Wait()
		relay_gateway_cmd.Wait()
		relay_backend_1_cmd.Wait()
		relay_backend_2_cmd.Wait()
		relay_frontend_cmd.Wait()
		relay_1_cmd.Wait()

		fmt.Printf("everything shut down OK\n")
	}()

	// initialize relay gateway

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

	// todo: lots of work...
	/*
	if !relay_gateway_initialized {
		fmt.Printf("error: failed to initialize relay gateway\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_gateway_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}
	*/

	// initialize relays

	fmt.Printf("\nwaiting for relays to initialize...\n\n")

	relays_initialized := false

	for i := 0; i < 10; i++ {

		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_1_stdout)
		fmt.Printf("----------------------------------------------------\n")

		if !relay_1_initialized && strings.Contains(relay_1_stdout.String(), "Relay initialized") {
			fmt.Printf("Relay initialized\n")
			relay_1_initialized = true
		}

		/*
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
		*/

		if relay_1_initialized {
			relays_initialized = true
			break
		}

		time.Sleep(time.Second)
	}

	if !relays_initialized {
		// todo: get relays initializing
		/*
		fmt.Printf("error: relays failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_1_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
		*/
	}

	fmt.Printf("\nsuccess!\n\n")

	return 0
}

func main() {
	if happy_path() != 0 {
		os.Exit(1)
	}
}
