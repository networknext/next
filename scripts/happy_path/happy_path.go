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

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
)

var commands []*exec.Cmd

func run(action string, log string, env ...string) *bytes.Buffer {

	fmt.Printf("run %s %s\n", action, strings.Join(env, " "))

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

	// build and run services as a developer would

	setup_emulators_stdout := run("setup-emulators", "logs/setup_emulators")
	_ = setup_emulators_stdout // todo

	// todo: initialize emulators

	magic_backend_stdout := run("magic-backend", "logs/magic_backend")
	relay_gateway_stdout := run("relay-gateway", "logs/relay_gateway")
	relay_backend_1_stdout := run("relay-backend", "logs/relay_backend_1")
	relay_backend_2_stdout := run("relay-backend", "logs/relay_backend_2", "HTTP_PORT=30002")

	time.Sleep(time.Second * 10)

	relay_1_stdout := run("relay", "logs/relay_1")
	relay_2_stdout := run("relay", "logs/relay_2", "RELAY_PORT=2001")
	relay_3_stdout := run("relay", "logs/relay_3", "RELAY_PORT=2002")
	relay_4_stdout := run("relay", "logs/relay_4", "RELAY_PORT=2003")
	relay_5_stdout := run("relay", "logs/relay_5", "RELAY_PORT=2004")

	server_backend4_stdout := run("server-backend4", "logs/server_backend4")
	server_backend5_stdout := run("server-backend5", "logs/server_backend5")

	analytics_stdout := run("analytics", "logs/analytics")

	// initialize the magic backend

	fmt.Printf("\ninitializing magic backend\n")

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
		fmt.Printf("\nerror: failed to initialize magic backend\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", magic_backend_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	// initialize relay gateway

	fmt.Printf("initializing relay gateway\n")

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
		fmt.Printf("\nerror: failed to initialize relay gateway\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_gateway_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	// initialize relay backend 1

	relay_backend_1_initialized := false

	fmt.Printf("initializing relay backend 1\n")

	for i := 0; i < 300; i++ {
		if strings.Contains(relay_backend_1_stdout.String(), "starting http server on port 30001") &&
			strings.Contains(relay_backend_1_stdout.String(), "loaded database: 'database.bin'") &&
			strings.Contains(relay_backend_1_stdout.String(), "route optimization: 10 relays in") &&
			strings.Contains(relay_backend_1_stdout.String(), "relay backend is ready") {
			relay_backend_1_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !relay_backend_1_initialized {
		fmt.Printf("\nerror: failed to initialize relay backend 1\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_backend_1_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	// initialize relay backend 2

	relay_backend_2_initialized := false

	fmt.Printf("initializing relay backend 2\n")

	for i := 0; i < 300; i++ {
		if strings.Contains(relay_backend_2_stdout.String(), "starting http server on port 30002") &&
			strings.Contains(relay_backend_2_stdout.String(), "loaded database: 'database.bin'") &&
			strings.Contains(relay_backend_2_stdout.String(), "route optimization: 10 relays in") &&
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

	// initialize relays

	fmt.Printf("initializing relays\n")

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
		fmt.Printf("\nerror: relays failed to initialize\n\n")
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

	// todo: verify that both relay backends see relay updates from each relay

	/*
	   strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.0'") &&
	   strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.1'") &&
	   strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.2'") &&
	   strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.3'") &&
	   strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.4'") {
	*/

	// initialize server backend 4

	fmt.Printf("initializing server backend 4\n")

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
		fmt.Printf("\nerror: server backend 4 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server_backend4_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	// initialize server backend 5

	fmt.Printf("initializing server backend 5\n")

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
		fmt.Printf("\nerror: server backend 5 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server_backend5_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	// initialize analytics

	fmt.Printf("initializing analytics\n")

	analytics_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(analytics_stdout.String(), "cost matrix num relays: 10") &&
			strings.Contains(analytics_stdout.String(), "route matrix num relays: 10") {
			// todo: additional checks, like we see each type of pubsub message we expect being processed at least once
			analytics_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !analytics_initialized {
		fmt.Printf("\nerror: analytics failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", analytics_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	// ==================================================================================

	fmt.Printf("\n")

	server4_stdout := run("server4", "logs/server4")
	server5_stdout := run("server5", "logs/server5")

	fmt.Printf("\n")

	// initialize server4

	fmt.Printf("initializing server 4\n")

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
		fmt.Printf("\nerror: server 4 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server4_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	// initialize server5

	fmt.Printf("initializing server 5\n")

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
		fmt.Printf("\nerror: server 5 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server5_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	// ==================================================================================

	fmt.Printf("\n")

	client4_stdout := run("client4", "logs/client4")
	client5_stdout := run("client5", "logs/client5")

	fmt.Printf("\n")

	// initialize client4

	fmt.Printf("initializing client 4\n")

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
		fmt.Printf("\nerror: client 4 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", client4_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	// initialize client5

	fmt.Printf("initializing client 5\n")

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
		fmt.Printf("\nerror: client 5 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", client5_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	// ==================================================================================

	fmt.Printf("\nsuccess!\n\n")

	if wait {
		waitDuration := envvar.GetDuration("WAIT_DURATION", 24*time.Hour)
		time.Sleep(waitDuration)
	}

	return 0
}

func cleanup() {
	for i := range commands {
		commands[i].Process.Kill()
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
