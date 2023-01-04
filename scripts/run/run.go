package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

const TestRouterPrivateKey = "ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M="
const TestBackendPrivateKey = "FXwFqzjGlIwUDwiq1N5Um5VUesdr4fP2hVV2cnJ+yARMYcqMR4c+1KC1l8PK4M9xCC0lPJEO1G8ZIq+6JZajQA=="

var cmd *exec.Cmd

func cleanup() {
	if cmd != nil {
		cmd.Process.Kill()
	}
	fmt.Print("\n")
}

func bash(command string) {

	cmd = exec.Command("bash", "-c", command)
	if cmd == nil {
		fmt.Printf("error: could not run bash!\n")
		os.Exit(1)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH=.") // IMPORTANT: linux needs this to run server4 etc.

	if err := cmd.Run(); err != nil {
		fmt.Printf("error: failed to run command: %v\n", err)
		os.Exit(1)
	}

	cmd.Wait()
}

func bash_ignore_result(command string) {

	cmd = exec.Command("bash", "-c", command)
	if cmd == nil {
		fmt.Printf("error: could not run bash!\n")
		os.Exit(1)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	cmd.Run()

	cmd.Wait()
}

func bash_no_wait(command string) {

	cmd = exec.Command("bash", "-c", command)
	if cmd == nil {
		fmt.Printf("error: could not run bash!\n")
		os.Exit(1)
	}

	cmd.Run()
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

	if len(args) < 2 || (len(args) == 2 && args[1] == "help") {
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
	} else if command == "test-relay" {
		test_relay()
	} else if command == "magic-backend" {
		magic_backend()
	} else if command == "relay-gateway" {
		relay_gateway()
	} else if command == "relay-backend" {
		relay_backend()
	} else if command == "analytics" {
		analytics()
	} else if command == "website-cruncher" {
		website_cruncher()
	} else if command == "portal-cruncher" {
		portal_cruncher()
	} else if command == "portal" {
		portal()
	} else if command == "pusher" {
		pusher()
	} else if command == "pingdom" {
		pingdom()
	} else if command == "relay" {
		relay()
	} else if command == "server-backend4" {
		server_backend4()
	} else if command == "server-backend5" {
		server_backend5()
	} else if command == "happy-path" {
		happy_path()
	} else if command == "happy-path-no-wait" {
		happy_path_no_wait()
	} else if command == "server4" {
		server4()
	} else if command == "server5" {
		server5()
	} else if command == "client4" {
		client4()
	} else if command == "client5" {
		client5()
	} else if command == "pubsub-emulator" {
		pubsub_emulator()
	} else if command == "bigquery-emulator" {
		bigquery_emulator()
	} else if command == "setup-emulators" {
		setup_emulators()
	} else if command == "func-sdk4" {
		func_sdk4()
	} else if command == "func-sdk5" {
		func_sdk5()
	} else if command == "func-backend4" {
		func_backend4()
	} else if command == "func-backend5" {
		func_backend5()
	} else if command == "func-backend" {
		func_backend(args[2:])
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
	bash("make ./dist/test4 -j && cd ./dist && ./test4")
}

func test_sdk5() {
	bash("make ./dist/test5 -j && cd ./dist && ./test5")
}

func test_relay() {
	bash("make ./dist/reference_relay -j && cd dist && ./reference_relay test")
}

func magic_backend() {
	bash("make ./dist/magic_backend -j && HTTP_PORT=41007 ./dist/magic_backend")
}

func relay_gateway() {
	bash("make ./dist/relay_gateway -j && HTTP_PORT=30000 ./dist/relay_gateway")
}

func relay_backend() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "30001"
	}
	bash(fmt.Sprintf("make ./dist/relay_backend -j && HTTP_PORT=%s ./dist/relay_backend", httpPort))
}

func analytics() {
	bash("make ./dist/analytics -j && HTTP_PORT=40001 ./dist/analytics")
}

func pusher() {
	bash("make ./dist/pusher -j && HTTP_PORT=40010 ./dist/pusher")
}

func pingdom() {
	bash("make ./dist/pingdom -j && HTTP_PORT=40011 ./dist/pingdom")
}

func relay() {
	relayPort := os.Getenv("RELAY_PORT")
	if relayPort == "" {
		relayPort = "2000"
	}
	bash(fmt.Sprintf("make -j ./dist/reference_relay -j && cd dist && RELAY_ADDRESS=127.0.0.1:%s ./reference_relay", relayPort))
}

func server_backend4() {
	bash("make ./dist/server_backend4 -j && HTTP_PORT=40000 UDP_PORT=40000 ./dist/server_backend4")
}

func server_backend5() {
	bash("make ./dist/server_backend5 -j && HTTP_PORT=45000 UDP_PORT=45000 ./dist/server_backend5")
}

func website_cruncher() {
	bash("make ./dist/website_cruncher -j && HTTP_PORT=40010 ./dist/website_cruncher")
}

func portal_cruncher() {
	bash("make ./dist/portal_cruncher -j && HTTP_PORT=40012 ./dist/portal_cruncher")
}

func portal() {
	bash("make ./dist/portal -j && PORT=20000 ./dist/portal")
}

func happy_path() {
	fmt.Printf("\ndon't worry. be happy.\n\n")
	bash("go run ./scripts/happy_path/happy_path.go")
}

func happy_path_no_wait() {
	fmt.Printf("\ndon't worry. be happy.\n\n")
	bash("go run ./scripts/happy_path/happy_path.go 1")
}

func server4() {
	bash("make ./dist/server4 -j && cd dist && ./server4")
}

func server5() {
	bash("make ./dist/server5 -j && cd dist && ./server5")
}

func client4() {
	bash("make ./dist/client4 -j && cd dist && ./client4")
}

func client5() {
	bash("make ./dist/client5 -j && cd dist && ./client5")
}

func pubsub_emulator() {
	bash_ignore_result("pkill -f pubsub-emulator")
	bash("gcloud beta emulators pubsub start --project=local --host-port=127.0.0.1:9000")
}

func bigquery_emulator() {
	bash_ignore_result("pkill -f bigquery-emulator")
	bash("bigquery-emulator --project=local --dataset=local")
}

func setup_emulators() {
	bash("go run ./scripts/setup_emulators/setup_emulators.go")
}

func func_sdk4() {
	bash("make func-test-sdk4 -j && cd dist && ./func_tests_sdk4")
}

func func_sdk5() {
	bash(fmt.Sprintf("make func-test-sdk5 -j && cd dist && TEST_ROUTER_PRIVATE_KEY=%s TEST_BACKEND_PRIVATE_KEY=%s ./func_tests_sdk5", TestRouterPrivateKey, TestBackendPrivateKey))
}

func func_backend4() {
	bash("make ./dist/func_backend4 -j && cd dist && ./func_backend4")
}

func func_backend5() {
	bash(fmt.Sprintf("make ./dist/func_backend5 -j && cd dist && TEST_ROUTER_PRIVATE_KEY=%s TEST_BACKEND_PRIVATE_KEY=%s ./func_backend5", TestRouterPrivateKey, TestBackendPrivateKey))
}

func func_backend(tests []string) {
	command := "make ./dist/func_tests_backend && cd dist && ./func_tests_backend"
	if len(tests) > 0 {
		for _, test := range tests {
			bash(fmt.Sprintf("%s %s", command, test))
		}
	} else {
		bash(command)
	}
}
