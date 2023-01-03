/*
   Network Next. You control the network.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"runtime"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
)

var commands []*exec.Cmd

func run(action string, log string, env ...string) *bytes.Buffer {

	fmt.Printf("   run %s %s\n", action, strings.Join(env, " "))

	cmd := exec.Command("./run", action)
	if cmd == nil {
		core.Error("could not run %s!\n", action)
		os.Exit(1)
	}

	cmd.Env = os.Environ()
	for i := range env {
		cmd.Env = append(cmd.Env, env[i])
	}

	var stdout bytes.Buffer
	stdout_pipe, err := cmd.StdoutPipe()
	if err != nil {
		core.Error("could not create stdout pipe for run")
		os.Exit(1)
	}

	cmd.Start()

	commands = append(commands, cmd)

	go func(output *bytes.Buffer) {
		file, err := os.Create(log)
		if err != nil {
			core.Error("could not create log file: %s", log)
			os.Exit(1)
		}
		writer := bufio.NewWriter(file)
		buf := bufio.NewReader(stdout_pipe)
		for {
			line, _, err := buf.ReadLine()
			if err != nil {
				break
			}
			writer.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), string(line)))
			writer.Flush()
			output.Write(line)
			output.Write([]byte("\n"))
		}
	}(&stdout)

	return &stdout
}

func happy_path(wait bool) int {

	if envvar.GetString("ENV", "") != "local" {
		core.Error("happy path only works in local env. please run 'next select local' first")
		return 1
	}

	os.Mkdir("logs", os.ModePerm)

	fmt.Printf("setting up emulators:\n\n")

	// run pubsub emulator

	pubsub_emulator_stdout := run("pubsub-emulator", "logs/pubsub_emulator")

	if runtime.GOOS != "linux" {

		pubsub_emulator_initialized := false

		for i := 0; i < 50; i++ {
			if strings.Contains(pubsub_emulator_stdout.String(), "[pubsub] INFO: Server started, listening on 9000") {
				pubsub_emulator_initialized = true
				break
			}
			time.Sleep(100 * time.Millisecond)
		}	

		if !pubsub_emulator_initialized {
			fmt.Printf("\nerror: failed to initialize pubsub emulator\n")
			fmt.Printf("-----------------------------------------\n")
			fmt.Printf("%s", pubsub_emulator_stdout.String())
			fmt.Printf("-----------------------------------------\n")
			return 1
		}

	} else {

		// hack: we can't reliably seem to get output from pubsub emulator via "run" on linux
		// so this is the best we can do. probably related to python buffering stdout from inside
		// gcloud.py

		time.Sleep(1*time.Second)
	}

	// setup emulators

	setup_emulators_stdout := run("setup-emulators", "logs/setup_emulators")

	setup_emulators_initialized := false

	for i := 0; i < 50; i++ {
		if strings.Contains(setup_emulators_stdout.String(), "finished setting up pubsub") {
			setup_emulators_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}	

	if !setup_emulators_initialized {
		fmt.Printf("\nerror: failed to setup emulators\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", setup_emulators_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	// initialize relay backend services

	fmt.Printf("\nstarting relay backend services:\n\n")

	magic_backend_stdout := run("magic-backend", "logs/magic_backend")
	relay_gateway_stdout := run("relay-gateway", "logs/relay_gateway")
	relay_backend_1_stdout := run("relay-backend", "logs/relay_backend_1")
	relay_backend_2_stdout := run("relay-backend", "logs/relay_backend_2", "HTTP_PORT=30002")

	fmt.Printf("\nverifying magic backend ...")

	magic_backend_initialized := false

	for i := 0; i < 200; i++ {
		if strings.Contains(magic_backend_stdout.String(), "starting http server on port 41007") &&
			strings.Contains(magic_backend_stdout.String(), "served magic values") {
			magic_backend_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !magic_backend_initialized {
		fmt.Printf("\n\nerror: failed to initialize magic backend\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", magic_backend_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	fmt.Printf("verifying relay gateway ...")

	relay_gateway_initialized := false

	for i := 0; i < 10; i++ {
		if strings.Contains(relay_gateway_stdout.String(), "loaded database: 'database.bin'") &&
			strings.Contains(relay_gateway_stdout.String(), "starting http server on port 30000") &&
			strings.Contains(relay_gateway_stdout.String(), "updated magic values: ") {
			relay_gateway_initialized = true
			break
		}
		time.Sleep(time.Second)
	}

	if !relay_gateway_initialized {
		fmt.Printf("\n\nerror: failed to initialize relay gateway\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_gateway_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	relay_backend_1_initialized := false

	fmt.Printf("verifying relay backend 1 ...")

	for i := 0; i < 300; i++ {
		if strings.Contains(relay_backend_1_stdout.String(), "starting http server on port 30001") &&
			strings.Contains(relay_backend_1_stdout.String(), "loaded database: 'database.bin'") &&
			strings.Contains(relay_backend_1_stdout.String(), "relay backend is ready") {
			relay_backend_1_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !relay_backend_1_initialized {
		fmt.Printf("\n\nerror: failed to initialize relay backend 1\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_backend_1_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	relay_backend_2_initialized := false

	fmt.Printf("verifying relay backend 2 ...")

	for i := 0; i < 300; i++ {
		if strings.Contains(relay_backend_2_stdout.String(), "starting http server on port 30002") &&
			strings.Contains(relay_backend_2_stdout.String(), "loaded database: 'database.bin'") &&
			strings.Contains(relay_backend_2_stdout.String(), "relay backend is ready") {
			relay_backend_2_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !relay_backend_2_initialized {
		fmt.Printf("\nerror: failed to initialize relay backend 2\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_backend_2_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// initialize relays

	fmt.Printf("\nstarting relays:\n\n")

	relay_1_stdout := run("relay", "logs/relay_1")
	relay_2_stdout := run("relay", "logs/relay_2", "RELAY_PORT=2001")
	relay_3_stdout := run("relay", "logs/relay_3", "RELAY_PORT=2002")
	relay_4_stdout := run("relay", "logs/relay_4", "RELAY_PORT=2003")
	relay_5_stdout := run("relay", "logs/relay_5", "RELAY_PORT=2004")

	fmt.Printf("\nverifying relays ...")

	relays_initialized := false

	relay_1_initialized := false
	relay_2_initialized := false
	relay_3_initialized := false
	relay_4_initialized := false
	relay_5_initialized := false

	for i := 0; i < 10; i++ {

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
			relay_4_initialized = true
		}

		if !relay_5_initialized && strings.Contains(relay_5_stdout.String(), "Relay initialized") {
			relay_5_initialized = true
		}

		if relay_1_initialized && relay_2_initialized && relay_3_initialized && relay_4_initialized && relay_5_initialized {
			relays_initialized = true
			break
		}

		time.Sleep(time.Second)
	}

	if !relays_initialized {
		fmt.Printf("\n\nerror: relays failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_gateway_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_backend_1_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_backend_2_stdout)
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

	fmt.Printf(" OK\n")

	fmt.Printf("verifying relay gateway sees relays ...")

	relay_gateway_sees_relays := false

	for i := 0; i < 200; i++ {
		if strings.Contains(relay_gateway_stdout.String(), " - relay update") &&
			strings.Contains(relay_gateway_stdout.String(), "sent batch 0 containing 5 messages") {
			relay_gateway_sees_relays = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !relay_gateway_sees_relays {
		fmt.Printf("\n\nerror: relay gateway does not see relays\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_gateway_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	fmt.Printf("verifying relay backend 1 sees relays ...")

	relay_backend_1_sees_relays := false

	for i := 0; i < 200; i++ {
		if strings.Contains(relay_backend_1_stdout.String(), "received relay update for 'local.0'") &&
			strings.Contains(relay_backend_1_stdout.String(), "received relay update for 'local.1'") &&
			strings.Contains(relay_backend_1_stdout.String(), "received relay update for 'local.2'") &&
			strings.Contains(relay_backend_1_stdout.String(), "received relay update for 'local.3'") &&
			strings.Contains(relay_backend_1_stdout.String(), "received relay update for 'local.4'") {
			relay_backend_1_sees_relays = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !relay_backend_1_sees_relays {
		fmt.Printf("\n\nerror: relay backend 1 does not see relays\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_backend_1_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	fmt.Printf("verifying relay backend 2 sees relays ...")

	relay_backend_2_sees_relays := false

	for i := 0; i < 200; i++ {
		if strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.0'") &&
			strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.1'") &&
			strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.2'") &&
			strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.3'") &&
			strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.4'") {
			relay_backend_2_sees_relays = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !relay_backend_2_sees_relays {
		fmt.Printf("\n\nerror: relay backend 2 does not see relays\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_backend_2_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// initialize server backends

	fmt.Printf("\nstarting server backends:\n\n")

	server_backend4_stdout := run("server-backend4", "logs/server_backend4")
	server_backend5_stdout := run("server-backend5", "logs/server_backend5")

	fmt.Printf("\nverifying server backend 4 ...")

	server_backend4_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(server_backend4_stdout.String(), "started http server on port 40000") &&
			strings.Contains(server_backend4_stdout.String(), "started udp server on port 40000") &&
			strings.Contains(server_backend4_stdout.String(), "updated route matrix: 10 relays") {
			server_backend4_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !server_backend4_initialized {
		fmt.Printf("\n\nerror: server backend 4 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server_backend4_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	fmt.Printf("verifying server backend 5 ...")

	server_backend5_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(server_backend5_stdout.String(), "starting http server on port 45000") &&
			strings.Contains(server_backend5_stdout.String(), "starting udp server on port 45000") &&
			strings.Contains(server_backend5_stdout.String(), "updated route matrix: 10 relays") &&
			strings.Contains(server_backend5_stdout.String(), "updated magic values: ") {
			server_backend5_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !server_backend5_initialized {
		fmt.Printf("\n\nerror: server backend 5 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server_backend5_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// initialize analytics

	fmt.Printf("\nstarting analytics:\n\n")

	analytics_stdout := run("analytics", "logs/analytics")

	fmt.Printf("\nverifying analytics ...")

	analytics_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(analytics_stdout.String(), "we became the leader") &&
		    strings.Contains(analytics_stdout.String(), "cost matrix num relays: 10") &&
			strings.Contains(analytics_stdout.String(), "route matrix num relays: 10") {
			// todo: additional checks, like we see each type of pubsub message we expect being processed at least once
			analytics_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !analytics_initialized {
		fmt.Printf("\n\nerror: analytics failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", analytics_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// ==================================================================================

	fmt.Printf("\nstarting servers:\n\n")

	server4_stdout := run("server4", "logs/server4")
	server5_stdout := run("server5", "logs/server5")

	// initialize server4

	fmt.Printf("\nverifying server 4 ...")

	server4_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(server4_stdout.String(), "welcome to network next :)") &&
			strings.Contains(server4_stdout.String(), "server is ready to receive client connections") {
			server4_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !server4_initialized {
		fmt.Printf("\n\nerror: server 4 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server4_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// initialize server5

	fmt.Printf("verifying server 5 ...")

	server5_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(server5_stdout.String(), "welcome to network next :)") &&
			strings.Contains(server5_stdout.String(), "server is ready to receive client connections") {
			server5_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !server5_initialized {
		fmt.Printf("\n\nerror: server 5 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server5_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// ==================================================================================

	fmt.Printf("\nstarting clients:\n\n")

	client4_stdout := run("client4", "logs/client4")
	client5_stdout := run("client5", "logs/client5")

	fmt.Printf("\n")

	// initialize client4

	fmt.Printf("verifying client 4 ...")

	client4_initialized := false

	for i := 0; i < 45; i++ {
		if strings.Contains(client4_stdout.String(), "client next route (committed)") &&
			strings.Contains(client4_stdout.String(), "client continues route (committed)") {
			client4_initialized = true
			break
		}
		time.Sleep(time.Second)
	}

	if !client4_initialized {
		fmt.Printf("\n\nerror: client 4 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", client4_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// initialize client5

	fmt.Printf("verifying client 5 ...")

	client5_initialized := false

	for i := 0; i < 30; i++ {
		if strings.Contains(client5_stdout.String(), "client next route (committed)") &&
			strings.Contains(client5_stdout.String(), "client continues route (committed)") {
			client5_initialized = true
			break
		}
		time.Sleep(time.Second)
	}

	if !client5_initialized {
		fmt.Printf("\n\nerror: client 5 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", client5_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// ==================================================================================

	fmt.Printf("\n*** SUCCESS! ***\n\n")

	if wait {
		waitDuration := envvar.GetDuration("WAIT_DURATION", 24*time.Hour)
		time.Sleep(waitDuration)
	}

	return 0
}

func bash(command string) {
	cmd := exec.Command("bash", "-c", command)
	if cmd == nil {
		fmt.Printf("error: could not run bash!\n")
		os.Exit(1)
	}
	cmd.Run()
	cmd.Wait()
}

func cleanup() {
	killList := [...]string{"reference_relay", "client4", "server4", "client5", "server5", "magic_backend", "relay_gateway", "relay_backend", "server_backend4", "server_backend5", "analytics"}
	for i := range killList {
		bash(fmt.Sprintf("pkill -f %s", killList[i]))
	}
}

func main() {

	wait := true
	if len(os.Args) > 1 {
		wait = false
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()

	result := happy_path(wait)

	cleanup()

	if result != 0 {
		os.Exit(1)
	}
}
