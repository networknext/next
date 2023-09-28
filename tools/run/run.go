package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"

	"github.com/joho/godotenv"
)

var cmd *exec.Cmd

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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-c
		if cmd.Process != nil {
			fmt.Printf("\n\n")
			if err := cmd.Process.Signal(sig); err != nil {
				fmt.Printf("error trying to signal child process: %v\n", err)
			}
			cmd.Wait()
		}
		os.Exit(1)
	}()

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
	} else if command == "test-sdk" {
		test_sdk()
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
	} else if command == "session-cruncher" {
		session_cruncher()
	} else if command == "server-cruncher" {
		server_cruncher()
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
	} else if command == "func-test-sdk" {
		func_test_sdk(args[2:])
	} else if command == "func-test-relay" {
		func_test_relay(args[2:])
	} else if command == "func-test-backend" {
		func_test_backend(args[2:])
	} else if command == "func-test-api" {
		func_test_api(args[2:])
	} else if command == "func-test-terraform" {
		func_test_terraform(args[2:])
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
	} else if command == "sql-local" {
		sql_local()
	} else if command == "sql-docker" {
		sql_docker()
	} else if command == "sql-staging" {
		sql_staging()
	} else if command == "extract-database" {
		extract_database()
	} else if command == "func-server" {
		func_server()
	} else if command == "func-client" {
		func_client()
	} else if command == "func-backend" {
		func_backend()
	} else if command == "load-test-portal" {
		load_test_portal()
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
	} else if command == "soak-test-relay" {
		soak_test_relay()
	} else if command == "config-amazon" {
		config_amazon()
	} else if command == "config-google" {
		config_google()
	} else if command == "config-akamai" {
		config_akamai()
	} else if command == "config-vultr" {
		config_vultr()
	} else if command == "portal" {
		portal()
	} else if command == "ip2location" {
		ip2location()
	} else if command == "generate-staging-sql" {
		generate_staging_sql()
	} else if command == "load-test-relays" {
		load_test_relays()
	} else if command == "load-test-servers" {
		load_test_servers()
	} else if command == "load-test-sessions" {
		load_test_sessions()
	} else if command == "redis-cluster" {
		redis_cluster()
	} else {
		fmt.Printf("\nunknown command\n\n")
	}
}

func help() {
	fmt.Printf("\nsyntax:\n\n    run <action> [args]\n\n")
}

func test() {
	fmt.Printf("\n")
	bash("go test ./modules/...")
	fmt.Printf("\n")
}

func test_sdk() {
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

func portal_cruncher() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "40012"
	}
	bash(fmt.Sprintf("HTTP_PORT=%s ./dist/portal_cruncher", httpPort))
}

func session_cruncher() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "40200"
	}
	bash(fmt.Sprintf("HTTP_PORT=%s ./dist/session_cruncher", httpPort))
}

func server_cruncher() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "40300"
	}
	bash(fmt.Sprintf("HTTP_PORT=%s ./dist/server_cruncher", httpPort))
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

func func_test_sdk(tests []string) {
	command := "cd dist && ./func_test_sdk"
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

func func_test_terraform(tests []string) {
	command := "cd dist && ./func_test_terraform"
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
	bash("psql -U developer postgres -f ./sql/create.sql -v ON_ERROR_STOP=1")
}

func sql_destroy() {
	bash("psql -U developer postgres -f ./sql/destroy.sql -v ON_ERROR_STOP=1")
}

func sql_local() {
	bash("psql -U developer postgres -f ./sql/local.sql -v ON_ERROR_STOP=1")
}

func sql_docker() {
	bash("psql -U developer postgres -f ./sql/docker.sql -v ON_ERROR_STOP=1")
}

func sql_staging() {
	bash("psql -U developer postgres -f ./sql/staging.sql -v ON_ERROR_STOP=1")
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

func load_test_portal() {
	bash("go run tools/load_test_portal/load_test_portal.go")
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

func config_amazon() {
	bash("go run sellers/amazon.go")
}

func config_google() {
	bash("go run sellers/google.go")
}

func config_akamai() {
	bash("go run sellers/akamai.go")
}

func config_vultr() {
	bash("go run sellers/vultr.go")
}

func soak_test_relay() {
	bash("cd dist && ./soak_test_relay stop")
}

type Environment struct {
	Name                     string `json:"name"`
	AdminURL                 string `json:"admin_url"`
	PortalURL                string `json:"portal_url"`
	DatabaseURL              string `json:"database_url"`
	SSHKeyFile               string `json:"ssh_key_filepath"`
	APIPrivateKey            string `json:"api_private_key"`
	APIKey                   string `json:"api_key"`
	VPNAddress               string `json:"vpn_address"`
	RelayBackendHostname     string `json:"relay_backend_hostname"`
	RelayBackendPublicKey    string `json:"relay_backend_public_key"`
	RelayArtifactsBucketName string `json:"relay_artifacts_bucket_name"`
}

func (e *Environment) Read() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	envFilePath := path.Join(homeDir, ".next")

	f, err := os.Open(envFilePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(e); err != nil {
		panic(err)
	}
}

func portal() {
	var env Environment
	env.Read()
	bash(fmt.Sprintf("cd portal && yarn serve-%s", env.Name))
}

func ip2location() {
	bash("cd dist && ./ip2location")
}

func generate_staging_sql() {
	bash("go run tools/generate_staging_sql/generate_staging_sql.go")
}

func load_test_relays() {
	bash("cd dist && ./load_test_relays")
}

func load_test_servers() {
	bash("cd dist && ./load_test_servers")
}

func load_test_sessions() {
	bash("cd dist && ./load_test_sessions")
}

func redis_cluster() {
	bash("go run tools/redis_cluster/redis_cluster.go")
}

func test_go_redis() {
	bash("go run tools/test_go_redis/test_go_redis.go")
}
