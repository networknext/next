/*
   Network Next. Copyright 2017 - 2026 Network Next, Inc.
   Licensed under the Network Next Source Available License 1.0
*/

package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
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

	// initialize postgres

	bash("psql -U developer -h localhost postgres -f ../schemas/sql/destroy.sql")
	bash("psql -U developer -h localhost postgres -f ../schemas/sql/create.sql")
	bash("psql -U developer -h localhost postgres -f ../schemas/sql/local.sql")

	// nuke redis

	redisClient := common.CreateRedisClient("127.0.0.1:6379")

	redisClient.FlushAll(context.Background())

	// initialize session cruncher

	fmt.Printf("starting session cruncher:\n\n")

	session_cruncher_stdout := run("session-cruncher", "logs/session_cruncher")

	fmt.Printf("\nverifying session cruncher ...")

	session_cruncher_initialized := false

	for i := 0; i < 1000; i++ {
		if strings.Contains(session_cruncher_stdout.String(), "starting http server on port 40200") {
			session_cruncher_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !session_cruncher_initialized {
		fmt.Printf("\n\nerror: session cruncher failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", session_cruncher_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// initialize server cruncher

	fmt.Printf("\nstarting server cruncher:\n\n")

	server_cruncher_stdout := run("server-cruncher", "logs/server_cruncher")

	fmt.Printf("\nverifying server cruncher ...")

	server_cruncher_initialized := false

	for i := 0; i < 1000; i++ {
		if strings.Contains(server_cruncher_stdout.String(), "starting http server on port 40300") {
			server_cruncher_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !server_cruncher_initialized {
		fmt.Printf("\n\nerror: server cruncher failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server_cruncher_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// initialize api

	fmt.Printf("\nstarting api:\n\n")

	api_stdout := run("api", "logs/api")

	fmt.Printf("\nverifying api ...")

	api_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(api_stdout.String(), "starting http server on port 50000") {
			api_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !api_initialized {
		fmt.Printf("\n\nerror: api failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", api_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// initialize relay backend services

	fmt.Printf("\nstarting relay backend services:\n\n")

	magic_backend_stdout := run("magic-backend", "logs/magic_backend")
	relay_gateway_stdout := run("relay-gateway", "logs/relay_gateway")
	relay_backend_stdout := run("relay-backend", "logs/relay_backend")

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
		if strings.Contains(relay_gateway_stdout.String(), "loaded database: database.bin") &&
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

	relay_backend_initialized := false

	fmt.Printf("verifying relay backend ...")

	for i := 0; i < 300; i++ {
		if strings.Contains(relay_backend_stdout.String(), "starting http server on port 30001") &&
			strings.Contains(relay_backend_stdout.String(), "loaded database: database.bin") &&
			strings.Contains(relay_backend_stdout.String(), "initial delay completed") {
			relay_backend_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !relay_backend_initialized {
		fmt.Printf("\n\nerror: failed to initialize relay backend\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_backend_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// initialize relays

	fmt.Printf("\nstarting relays:\n\n")

	relay_1_stdout := run("relay", "logs/relay_1", "RELAY_PORT=2000 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1")
	relay_2_stdout := run("relay", "logs/relay_2", "RELAY_PORT=2001 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1")
	relay_3_stdout := run("relay", "logs/relay_3", "RELAY_PORT=2002 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1")
	relay_4_stdout := run("relay", "logs/relay_4", "RELAY_PORT=2003 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1")
	relay_5_stdout := run("relay", "logs/relay_5", "RELAY_PORT=2004 RELAY_LOG_LEVEL=3 RELAY_PRINT_COUNTERS=1")

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
		fmt.Printf("%s", relay_backend_stdout)
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
		if strings.Contains(relay_gateway_stdout.String(), "received update for local.") {
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

	fmt.Printf("verifying relay backend sees relays ...")

	relay_backend_sees_relays := false

	for i := 0; i < 200; i++ {
		if strings.Contains(relay_backend_stdout.String(), "received update for local.0") &&
			strings.Contains(relay_backend_stdout.String(), "received update for local.1") &&
			strings.Contains(relay_backend_stdout.String(), "received update for local.2") &&
			strings.Contains(relay_backend_stdout.String(), "received update for local.3") &&
			strings.Contains(relay_backend_stdout.String(), "received update for local.4") {
			relay_backend_sees_relays = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !relay_backend_sees_relays {
		fmt.Printf("\n\nerror: relay backend does not see relays\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_backend_stdout.String())
		fmt.Printf("-----------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// initialize server backends

	fmt.Printf("\nstarting server backend:\n\n")

	server_backend_stdout := run("server-backend", "logs/server_backend")

	fmt.Printf("\nverifying server backend ...")

	server_backend_initialized := false

	for i := 0; i < 1000; i++ {
		if strings.Contains(server_backend_stdout.String(), "starting http server on port 40000") &&
			strings.Contains(server_backend_stdout.String(), "starting udp server on port 40000") &&
			strings.Contains(server_backend_stdout.String(), "updated route matrix: 10 relays") &&
			strings.Contains(server_backend_stdout.String(), "updated magic values: ") {
			server_backend_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !server_backend_initialized {
		fmt.Printf("\n\nerror: server backend failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server_backend_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// ==================================================================================

	fmt.Printf("\nwaiting for leader election\n\n")

	fmt.Printf("   relay backend ...")

	relay_backend_leader_elected := false

	for i := 0; i < 250; i++ {
		relay_backend_is_leader := strings.Contains(relay_backend_stdout.String(), "we became the leader")
		if relay_backend_is_leader {
			relay_backend_leader_elected = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !relay_backend_leader_elected {
		fmt.Printf("\n\nerror: no relay backend leader?\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", relay_backend_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	// ==================================================================================

	fmt.Printf("\nstarting client and server:\n\n")

	client_stdout := run("client", "logs/client")
	server_stdout := run("server", "logs/server")

	fmt.Printf("\nverifying server ...")

	server_initialized := false

	for i := 0; i < 600; i++ {
		if strings.Contains(server_stdout.String(), "welcome to network next :)") &&
			strings.Contains(server_stdout.String(), "server is ready to receive client connections") {
			server_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !server_initialized {
		fmt.Printf("\n\nerror: server failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	fmt.Printf(" OK\n")

	fmt.Printf("verifying client ...")

	client_initialized := false

	for i := 0; i < 600; i++ {
		if strings.Contains(client_stdout.String(), "client next route") &&
			strings.Contains(client_stdout.String(), "client continues route") {
			client_initialized = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !client_initialized {
		fmt.Printf("\n\nerror: client failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", client_stdout)
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server_stdout)
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
	killList := [...]string{"api", "relay", "client", "server", "magic_backend", "relay_gateway", "relay_backend", "server_backend", "session_cruncher", "server_cruncher"}
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
