package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
)

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
	cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH=.")

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
	} else if command == "api" {
		api()
	} else if command == "portal-cruncher" {
		portal_cruncher()
	} else if command == "map-cruncher" {
		map_cruncher()
	} else if command == "relay" {
		relay()
	} else if command == "server-backend" {
		server_backend()
	} else if command == "happy-path" {
		happy_path()
	} else if command == "happy-path-no-wait" {
		happy_path_no_wait()
	} else if command == "client" {
		client()
	} else if command == "server" {
		server()
	} else if command == "pubsub-emulator" {
		pubsub_emulator()
	} else if command == "bigquery-emulator" {
		bigquery_emulator()
	} else if command == "setup-emulators" {
		setup_emulators()
	} else if command == "func-test-sdk5" {
		func_test_sdk5(args[2:])
	} else if command == "func-test-relay" {
		func_test_relay(args[2:])
	} else if command == "func-test-backend" {
		func_test_backend(args[2:])
	} else if command == "func-test-api" {
		func_test_api(args[2:])
	} else if command == "func-test-portal" {
		func_test_portal(args[2:])
	} else if command == "func-test-database" {
		func_test_database(args[2:])
	} else if command == "raspberry-backend" {
		raspberry_backend()
	} else if command == "raspberry-server" {
		raspberry_server()
	} else if command == "raspberry-client" {
		raspberry_client()
	} else if command == "relay-keygen" {
		relay_keygen()
	} else if command == "sql-create" {
		sql_create()
	} else if command == "sql-destroy" {
		sql_destroy()
	} else if command == "sql-dev" {
		sql_dev()
	} else if command == "sql-local" {
		sql_local()
	} else if command == "sql-docker" {
		sql_docker()
	} else if command == "extract-database" {
		extract_database()
	} else if command == "func-server" {
		func_server()
	} else if command == "func-client" {
		func_client()
	} else if command == "func-backend" {
		func_backend()
	} else if command == "load-test-redis-portal" {
		load_test_redis_portal()
	} else if command == "load-test-redis-data" {
		load_test_redis_data()
	} else if command == "load-test-redis-pubsub" {
		load_test_redis_pubsub()
	} else if command == "load-test-redis-streams" {
		load_test_redis_streams()
	} else if command == "load-test-map" {
		load_test_map()
	} else if command == "load-test-optimize" {
		load_test_optimize()
	} else if command == "load-test-route-matrix" {
		load_test_route_matrix()
	} else if command == "load-test-relay-manager" {
		load_test_relay_manager()
	} else if command == "load-test-crypto-box" {
		load_test_crypto_box()
	} else if command == "load-test-crypto-sign" {
		load_test_crypto_sign()
	} else if command == "load-test-crypto-auth" {
		load_test_crypto_sign()
	} else if command == "load-test-server-update" {
		load_test_server_update()
	} else if command == "load-test-session-update" {
		load_test_session_update()
	} else if command == "amazon-config" {
		amazon_config()
	} else if command == "google-config" {
		google_config()
	} else if command == "akamai-config" {
		akamai_config()
	} else if command == "vultr-config" {
		vultr_config()
	} else if command == "soak-test-relay" {
		soak_test_relay()
	}

	cleanup()
}

func help() {
	fmt.Printf("todo: help\n")
}

func test() {
	bash("go test ./modules/...")
}

func test_sdk5() {
	bash("cd ./dist && ./test")
}

func magic_backend() {
	bash("HTTP_PORT=41007 ./dist/magic_backend")
}

func relay_gateway() {
	bash("HTTP_PORT=30000 ./dist/relay_gateway")
}

func relay_backend() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "30001"
	}
	bash(fmt.Sprintf("HTTP_PORT=%s ./dist/relay_backend", httpPort))
}

func analytics() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "40001"
	}
	bash(fmt.Sprintf("HTTP_PORT=%s ./dist/analytics", httpPort))
}

func api() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "50000"
	}
	bash(fmt.Sprintf("HTTP_PORT=%s ./dist/api", httpPort))
}

func sync() {
	bash("HTTP_PORT=40010 ./dist/sync")
}

func pingdom() {
	bash("HTTP_PORT=40020 ./dist/pingdom")
}

func relay() {
	relayPort := os.Getenv("RELAY_PORT")
	if relayPort == "" {
		relayPort = "2000"
	}
	bash(fmt.Sprintf("cd dist && RELAY_PUBLIC_ADDRESS=127.0.0.1:%s ./relay-debug", relayPort))
}

func server_backend() {
	bash("HTTP_PORT=40000 ./dist/server_backend")
}

func website_cruncher() {
	bash("HTTP_PORT=40010 ./dist/website_cruncher")
}

func portal_cruncher() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "40012"
	}
	bash(fmt.Sprintf("HTTP_PORT=%s ./dist/portal_cruncher", httpPort))
}

func map_cruncher() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "40100"
	}
	bash(fmt.Sprintf("HTTP_PORT=%s ./dist/map_cruncher", httpPort))
}

func happy_path() {
	fmt.Printf("\ndon't worry. be happy.\n\n")
	bash("go run ./tools/happy_path/happy_path.go")
}

func happy_path_no_wait() {
	fmt.Printf("\ndon't worry. be happy.\n\n")
	bash("go run ./tools/happy_path/happy_path.go 1")
}

func server() {
	bash("cd dist && ./server")
}

func client() {
	bash("cd dist && ./client")
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
	bash("go run ./tools/setup_emulators/setup_emulators.go")
}

func func_test_sdk5(tests []string) {
	command := "cd dist && ./func_test_sdk5"
	if len(tests) > 0 {
		for _, test := range tests {
			bash(fmt.Sprintf("%s %s", command, test))
		}
	} else {
		bash(command)
	}
}

func func_test_relay(tests []string) {
	command := "cd dist && ./func_test_relay"
	if len(tests) > 0 {
		for _, test := range tests {
			bash(fmt.Sprintf("%s %s", command, test))
		}
	} else {
		bash(command)
	}
}

func func_test_backend(tests []string) {
	command := "cd dist && ./func_test_backend"
	if len(tests) > 0 {
		for _, test := range tests {
			bash(fmt.Sprintf("%s %s", command, test))
		}
	} else {
		bash(command)
	}
}

func func_test_api(tests []string) {
	command := "cd dist && ./func_test_api"
	if len(tests) > 0 {
		for _, test := range tests {
			bash(fmt.Sprintf("%s %s", command, test))
		}
	} else {
		bash(command)
	}
}

func func_test_portal(tests []string) {
	command := "cd dist && ./func_test_portal"
	if len(tests) > 0 {
		for _, test := range tests {
			bash(fmt.Sprintf("%s %s", command, test))
		}
	} else {
		bash(command)
	}
}

func func_test_database(tests []string) {
	command := "cd dist && ./func_test_database"
	if len(tests) > 0 {
		for _, test := range tests {
			bash(fmt.Sprintf("%s %s", command, test))
		}
	} else {
		bash(command)
	}
}

func raspberry_backend() {
	bash("HTTP_PORT=40100 ./dist/raspberry_backend")
}

func raspberry_client() {
	bash("cd dist && ./raspberry_client")
}

func raspberry_server() {
	bash("cd dist && ./raspberry_server")
}

func relay_keygen() {
	bash("go run tools/relay_keygen/relay_keygen.go")
}

func sql_create() {
	bash("psql -U developer postgres -f ./schemas/sql/create.sql -v ON_ERROR_STOP=1")
}

func sql_destroy() {
	bash("psql -U developer postgres -f ./schemas/sql/destroy.sql -v ON_ERROR_STOP=1")
}

func sql_dev() {
	bash("psql -U developer postgres -f ./schemas/sql/dev.sql -v ON_ERROR_STOP=1")
}

func sql_local() {
	bash("psql -U developer postgres -f ./schemas/sql/local.sql -v ON_ERROR_STOP=1")
}

func sql_docker() {
	bash("psql -U developer postgres -f ./schemas/sql/docker.sql -v ON_ERROR_STOP=1")
}

func extract_database() {
	bash("go run tools/extract_database/extract_database.go")
}

func func_server() {
	bash("cd dist && ./func_server")
}

func func_client() {
	bash("cd dist && ./func_client")
}

func func_backend() {
	bash("cd dist && ./func_backend")
}

func load_test_redis_data() {
	bash("go run tools/load_test_redis_data/load_test_redis_data.go")
}

func load_test_redis_pubsub() {
	bash("go run tools/load_test_redis_pubsub/load_test_redis_pubsub.go")
}

func load_test_redis_streams() {
	bash("go run tools/load_test_redis_streams/load_test_redis_streams.go")
}

func load_test_redis_portal() {
	bash("go run tools/load_test_redis_portal/load_test_redis_portal.go")
}

func load_test_map() {
	bash("go run tools/load_test_map/load_test_map.go")
}

func load_test_optimize() {
	bash("go run tools/load_test_optimize/load_test_optimize.go")
}

func load_test_route_matrix() {
	bash("go run tools/load_test_route_matrix/load_test_route_matrix.go")
}

func load_test_relay_manager() {
	bash("go run tools/load_test_relay_manager/load_test_relay_manager.go")
}

func load_test_crypto_box() {
	bash("go run tools/load_test_crypto_box/load_test_crypto_box.go")
}

func load_test_crypto_sign() {
	bash("go run tools/load_test_crypto_sign/load_test_crypto_sign.go")
}

func load_test_crypto_auth() {
	bash("go run tools/load_test_crypto_auth/load_test_crypto_auth.go")
}

func load_test_server_update() {
	bash("go run tools/load_test_server_update/load_test_server_update.go")
}

func load_test_session_update() {
	bash("go run tools/load_test_session_update/load_test_session_update.go")
}

func amazon_config() {
	bash("go run tools/amazon_config/amazon_config.go")
}

func google_config() {
	bash("go run tools/google_config/google_config.go")
}

func akamai_config() {
	bash("go run tools/akamai_config/akamai_config.go")
}

func vultr_config() {
	bash("go run tools/vultr_config/vultr_config.go")
}

func soak_test_relay() {
	bash("go run tools/soak_test_relay/soak_test_relay.go")
}
