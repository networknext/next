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
	"bufio"
)

func run_make(action string, log string) (*exec.Cmd, *bytes.Buffer) {

	fmt.Printf("make %s\n", action)

	cmd := exec.Command("make", action)
	if cmd == nil {
		panic("could not run make!\n")
		return nil, nil
	}

	var stdout bytes.Buffer

    stdout_pipe, err := cmd.StdoutPipe()
    if err != nil {
        panic(err)
    }

    cmd.Start()

    go func(output *bytes.Buffer) {
		file, err := os.Create(log)
		if err != nil {
			panic(err)
		}
		writer := bufio.NewWriter(file)
	    buf := bufio.NewReader(stdout_pipe) 
	    for {
	        line, _, _ := buf.ReadLine()
	        writer.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), string(line)))
	        writer.Flush()
	        output.Write(line)
	        output.Write([]byte("\n"))
	    }
    }(&stdout)

	return cmd, &stdout
}

func run_relay(port int, log string) (*exec.Cmd, *bytes.Buffer) {

	fmt.Printf("PORT=%d make %s\n", port, "dev-relay")

	cmd := exec.Command("./dist/reference_relay")
	if cmd == nil {
		panic("could not run relay!\n")
		return nil, nil
	}

	cmd.Env = make([]string, 0)
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_ADDRESS=127.0.0.1:%d", port))
	cmd.Env = append(cmd.Env, "RELAY_PRIVATE_KEY=lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=")
	cmd.Env = append(cmd.Env, "RELAY_PUBLIC_KEY=9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=")
	cmd.Env = append(cmd.Env, "RELAY_ROUTER_PUBLIC_KEY=SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=")
	cmd.Env = append(cmd.Env, "RELAY_GATEWAY=http://127.0.0.1:30000")

	var stdout bytes.Buffer

    stdout_pipe, err := cmd.StdoutPipe()
    if err != nil {
        panic(err)
    }

    cmd.Start()

    go func(output *bytes.Buffer) {
		file, err := os.Create(log)
		if err != nil {
			panic(err)
		}
		writer := bufio.NewWriter(file)
	    buf := bufio.NewReader(stdout_pipe) 
	    for {
	        line, _, _ := buf.ReadLine()
	        writer.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), string(line)))
	        writer.Flush()
	        output.Write(line)
	        output.Write([]byte("\n"))	    
	    }
    }(&stdout)

	return cmd, &stdout
}

func happy_path() int {

	fmt.Printf("\nhappy path\n\n")

	os.Mkdir("logs", os.ModePerm)

	// build and run services, as a developer would via "make dev-*" as much as possible

	magic_backend_cmd, magic_backend_stdout := run_make("dev-magic-backend", "logs/magic_backend")
	relay_gateway_cmd, relay_gateway_stdout := run_make("dev-relay-gateway", "logs/relay_gateway")
	relay_backend_1_cmd, relay_backend_1_stdout := run_make("dev-relay-backend-1", "logs/relay_backend_1")
	relay_backend_2_cmd, relay_backend_2_stdout := run_make("dev-relay-backend-2", "logs/relay_backend_2")
	relay_frontend_cmd, relay_frontend_stdout := run_make("dev-relay-frontend", "logs/relay_frontend")

	relay_1_cmd, relay_1_stdout := run_make("dev-relay", "logs/relay_1")
	relay_2_cmd, relay_2_stdout := run_relay(2001, "logs/relay_2")
	relay_3_cmd, relay_3_stdout := run_relay(2002, "logs/relay_3")
	relay_4_cmd, relay_4_stdout := run_relay(2003, "logs/relay_4")
	relay_5_cmd, relay_5_stdout := run_relay(2004, "logs/relay_5")

	server_backend4_cmd, server_backend4_stdout := run_make("dev-server-backend4", "logs/server_backend4")
	server4_cmd, server4_stdout := run_make("dev-server4", "logs/server4")
	client4_cmd, client4_stdout := run_make("dev-client4", "logs/client4")

	// make sure everything gets cleaned up

	defer func() {

		magic_backend_cmd.Process.Signal(syscall.SIGTERM)
		relay_gateway_cmd.Process.Signal(syscall.SIGTERM)
		relay_backend_1_cmd.Process.Signal(syscall.SIGTERM)
		relay_backend_2_cmd.Process.Signal(syscall.SIGTERM)
		relay_frontend_cmd.Process.Signal(syscall.SIGTERM)
		relay_1_cmd.Process.Signal(syscall.SIGTERM)
		relay_2_cmd.Process.Signal(syscall.SIGTERM)
		relay_3_cmd.Process.Signal(syscall.SIGTERM)
		relay_4_cmd.Process.Signal(syscall.SIGTERM)
		relay_5_cmd.Process.Signal(syscall.SIGTERM)
		server_backend4_cmd.Process.Signal(syscall.SIGTERM)
		server4_cmd.Process.Signal(syscall.SIGTERM)
		client4_cmd.Process.Signal(syscall.SIGTERM)

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
		server_backend4_cmd.Wait()
		server4_cmd.Wait()
		client4_cmd.Wait()
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
		if strings.Contains(relay_gateway_stdout.String(), "loaded database.bin") &&
		   strings.Contains(relay_gateway_stdout.String(), "starting http server on port 30000") &&
		   strings.Contains(relay_gateway_stdout.String(), "started watchman on ") {
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
		fmt.Printf("\nerror: failed to initialize relay backend 1\n")
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
		fmt.Printf("\nerror: failed to initialize relay backend 2\n")
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
		fmt.Printf("\nerror: failed to initialize relay frontend\n")
		fmt.Printf("-----------------------------------------\n")
		fmt.Printf("%s", relay_frontend_stdout.String())
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

	// initialize server backend 4

	fmt.Printf("initializing server backend 4\n")

	server_backend4_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(server_backend4_stdout.String(), "started http server on port 40000") &&
		   strings.Contains(server_backend4_stdout.String(), "started udp server on port 40000") &&
		   strings.Contains(server_backend4_stdout.String(), "updated route matrix: 5 relays") {
		   	server_backend4_initialized = true
		   	break
		}
		time.Sleep(100*time.Millisecond)
	}

	if !server_backend4_initialized {
		fmt.Printf("\nerror: server backend 4 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server_backend4_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	// initialize server4

	fmt.Printf("initializing server 4\n")

	server4_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(server4_stdout.String(), "welcome to network next :)") {
		   	server4_initialized = true
		   	break
		}
		time.Sleep(100*time.Millisecond)
	}

	if !server4_initialized {
		fmt.Printf("\nerror: server 4 failed to initialize\n\n")
		fmt.Printf("----------------------------------------------------\n")
		fmt.Printf("%s", server4_stdout)
		fmt.Printf("----------------------------------------------------\n")
		return 1
	}

	// initialize client4

	fmt.Printf("initializing client 4\n")

	client4_initialized := false

	for i := 0; i < 20; i++ {
		if strings.Contains(client4_stdout.String(), "client next route (committed)") {
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

	// -------------------------------------------

	fmt.Printf("\nsuccess!\n\n")

	time.Sleep(time.Hour)

	return 0
}

func main() {
	if happy_path() != 0 {
		os.Exit(1)
	}
}
