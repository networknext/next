package main

import (
    "os"
    "os/exec"
    "os/signal"
    "fmt"
    "syscall"

    "github.com/joho/godotenv"
)

var cmd *exec.Cmd

func cleanup() {
    cmd.Process.Kill()
    fmt.Print("\n")
}

func bash(command string) {

    cmd = exec.Command("bash", "-c", command)
    if cmd == nil {
        fmt.Printf("error: could not run bash!\n")
        os.Exit(1)
    }

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
	    fmt.Printf("error: failed to run command: %v\n", err)
	    os.Exit(1)
	}

	cmd.Wait()
}

func main() {

	c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        cleanup()
        os.Exit(1)
    }()

	args := os.Args
	
	if len(args) < 2 || (len(args) == 2 && args[1]=="help") {
		help()
		return
	}

	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("error: could not load .env file")
		os.Exit(1)
	}

	command := args[1]

	if command == "test" {
		test()
	} else if command == "test-sdk4" || command == "test4" {
		test_sdk4()
	} else if command == "test-sdk5" || command == "test5" {
		test_sdk5()
	} else if command == "magic-backend" {
		magic_backend()
	} else if command == "relay-gateway" {
		relay_gateway()
	} else if command == "relay-backend" {
		relay_backend()
	} else if command == "analytics" {
		analytics()
	} else if command == "relay" {
		relay()
	} else if command == "server-backend4" {
		server_backend4()
	} else if command == "server-backend5" {
		server_backend5()
	} else if command == "happy-path" {
		happy_path()
	} else if command == "server4" {
		server4()
	} else if command == "server5" {
		server5()
	} else if command == "client4" {
		client4()
	} else if command == "client5" {
		client5()
	}

	cleanup()
}

func help() {
	fmt.Printf("todo: help\n")
}

func test() {
	bash("./scripts/test-backend.sh")
}

func test_sdk4() {
	bash("make ./dist/test4 && cd ./dist && ./test4")
}

func test_sdk5() {
	bash("make ./dist/test5 && cd ./dist && ./test5")
}

func magic_backend() {
	bash("make ./dist/magic_backend && HTTP_PORT=41007 ./dist/magic_backend")
}

func relay_gateway() {
	bash("make ./dist/relay_gateway && HTTP_PORT=30000 ./dist/relay_gateway")
}

func relay_backend() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "30001"
	}
	bash(fmt.Sprintf("make ./dist/relay_backend && HTTP_PORT=%s ./dist/relay_backend", httpPort))
}

func analytics() {
	bash("make ./dist/analytics && HTTP_PORT=40001 ./dist/analytics")
}

func relay() {
	relayPort := os.Getenv("RELAY_PORT")
	if relayPort == "" {
		relayPort = "2000"
	}
	bash(fmt.Sprintf("make ./dist/reference_relay && cd dist && RELAY_ADDRESS=127.0.0.1:%s ./reference_relay", relayPort))
}

func server_backend4() {
	bash("make ./dist/server_backend4 && HTTP_PORT=40000 UDP_PORT=40000 ./dist/server_backend4")
}

func server_backend5() {
	bash("make ./dist/server_backend5 && HTTP_PORT=45000 UDP_PORT=45000 ./dist/server_backend5")
}

func happy_path() {
	fmt.Printf("\ndon't worry. be happy.\n\n")
	bash("./build.sh && go run ./cmd/happy_path/happy_path.go")
}

func server4() {
	bash("make ./dist/server4 && cd dist && ./server4")
}

func server5() {
	bash("make ./dist/server5 && cd dist && ./server5")
}

func client4() {
	bash("make ./dist/client4 && cd dist && ./client4")
}

func client5() {
	bash("make ./dist/client5 && cd dist && ./client5")
}
