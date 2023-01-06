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

	// initialize relay backend services

	fmt.Printf("starting relay backend services:\n\n")

	magic_backend_stdout := run("magic-backend", "logs/magic_backend")
	relay_gateway_stdout := run("relay-gateway", "logs/relay_gateway")
	relay_backend_1_stdout := run("relay-backend", "logs/relay_backend_1")
	relay_backend_2_stdout := run("relay-backend", "logs/relay_backend_2", "HTTP_PORT=30002")

	fmt.Printf("\nverifying magic backend ...")

	magic_backend_initialized := false

	for i := 0; i < 200; i++ {
		if strings.Contains(magic_backend_stdout.String(), "starting http server on port 41007") {
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

	fmt.Printf("\nstarting server backend:\n\n")

	server_backend5_stdout := run("server-backend5", "logs/server_backend5")

	fmt.Printf("\nverifying server backend ...")

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

	analytics_1_stdout := run("analytics", "logs/analytics_1")
	analytics_2_stdout := run("analytics", "logs/analytics_2", "HTTP_PORT=40002")

	fmt.Printf("\nverifying analytics 1 ...")

	analytics_1_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(analytics_1_stdout.String(), "starting http server on port 40001") {
			analytics_1_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !analytics_1_initialized {
		fmt.Printf("\n\nerror: analytics 1 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", analytics_1_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	fmt.Printf("verifying analytics 2 ...")

	analytics_2_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(analytics_2_stdout.String(), "starting http server on port 40002") {
			analytics_2_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !analytics_2_initialized {
		fmt.Printf("\n\nerror: analytics 2 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", analytics_2_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// ==================================================================================

	fmt.Printf("\nstarting server:\n\n")

	server5_stdout := run("server5", "logs/server5")

	// initialize server5

	fmt.Printf("verifying server ...")

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

	fmt.Printf("\nstarting client:\n\n")

	client5_stdout := run("client5", "logs/client5")

	fmt.Printf("\n")

	// initialize client5

	fmt.Printf("verifying client ...")

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

	fmt.Printf("verifying leader election in relay backend ...")

	relay_backend_1_is_leader := strings.Contains(relay_backend_1_stdout.String(), "we became the leader")
	relay_backend_2_is_leader := strings.Contains(relay_backend_2_stdout.String(), "we became the leader")

	if relay_backend_1_is_leader && relay_backend_2_is_leader {
		fmt.Printf("\n\nerror: leader flap in relay backend\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_backend_1_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_backend_2_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	if relay_backend_1_is_leader && relay_backend_2_is_leader {
		fmt.Printf("\n\nerror: no relay backend leader?\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_backend_1_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_backend_2_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	fmt.Printf("verifying leader election in analytics ...")

	analytics_1_is_leader := strings.Contains(analytics_1_stdout.String(), "we became the leader")
	analytics_2_is_leader := strings.Contains(analytics_2_stdout.String(), "we became the leader")

	if analytics_1_is_leader && analytics_2_is_leader {
		fmt.Printf("\n\nerror: leader flap in analytics\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", analytics_1_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", analytics_2_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	if !analytics_1_is_leader && !analytics_2_is_leader {
		fmt.Printf("\n\nerror: no analytics leader?!\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", analytics_1_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", analytics_2_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}
	fmt.Printf(" OK\n")

	fmt.Printf("verifying analytics leader ...")

	analytics_leader_stdout := analytics_1_stdout
	if analytics_2_is_leader {
		analytics_leader_stdout = analytics_2_stdout
	}

	if strings.Contains(analytics_leader_stdout.String(), "we are no longer the leader") ||
		!strings.Contains(analytics_leader_stdout.String(), "cost matrix num relays: 10") ||
		!strings.Contains(analytics_leader_stdout.String(), "route matrix num relays: 10") {
		fmt.Printf("\n\nerror: analytics leader did not verify\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", analytics_leader_stdout)
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
