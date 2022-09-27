/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"google.golang.org/api/option"
)

var processes []*os.Process

func run_make(action string, log string) *bytes.Buffer {

	fmt.Printf("make %s\n", action)

	cmd := exec.Command("make", action)
	if cmd == nil {
		core.Error("could not run make!\n")
		os.Exit(1)
	}

	var stdout bytes.Buffer

	stdout_pipe, err := cmd.StdoutPipe()
	if err != nil {
		core.Error("could not create stdout pipe for make")
		os.Exit(1)
	}

	cmd.Start()

	processes = append(processes, cmd.Process)

	go func(output *bytes.Buffer) {
		file, err := os.Create(log)
		if err != nil {
			core.Error("could not create log file: %s", log)
			os.Exit(1)
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

	return &stdout
}

func run_relay(port int, log string) *bytes.Buffer {

	fmt.Printf("RELAY_PORT=%d make %s\n", port, "dev-relay")

	cmd := exec.Command("./dist/reference_relay")
	if cmd == nil {
		panic("could not run relay!\n")
		return nil
	}

	cmd.Env = make([]string, 0)
	cmd.Env = append(cmd.Env, fmt.Sprintf("RELAY_ADDRESS=127.0.0.1:%d", port))
	cmd.Env = append(cmd.Env, "RELAY_PRIVATE_KEY=lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=")
	cmd.Env = append(cmd.Env, "RELAY_PUBLIC_KEY=9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=")
	cmd.Env = append(cmd.Env, "RELAY_ROUTER_PUBLIC_KEY=SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=")
	cmd.Env = append(cmd.Env, "RELAY_GATEWAY=http://127.0.0.1:30000")
	cmd.Env = append(cmd.Env, "RELAY_DEBUG=1")

	var stdout bytes.Buffer

	stdout_pipe, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	cmd.Start()

	processes = append(processes, cmd.Process)

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

	return &stdout
}

func happy_path() int {

	if envvar.GetString("ENV", "") != "local" {
		core.Error("happy path only works in local env. please run 'next select local' first")
		return 1
	}

	fmt.Printf("\nhappy path\n\n")

	os.Mkdir("logs", os.ModePerm)

	// build and run services, as a developer would via "make dev-*" as much as possible

	magic_backend_stdout := run_make("dev-magic-backend", "logs/magic_backend")
	relay_gateway_stdout := run_make("dev-relay-gateway", "logs/relay_gateway")
	relay_backend_1_stdout := run_make("dev-relay-backend", "logs/relay_backend_1")
	relay_backend_2_stdout := run_make("dev-relay-backend-2", "logs/relay_backend_2")

	time.Sleep(time.Second * 5)

	relay_1_stdout := run_make("dev-relay", "logs/relay_1")
	relay_2_stdout := run_relay(2001, "logs/relay_2")
	relay_3_stdout := run_relay(2002, "logs/relay_3")
	relay_4_stdout := run_relay(2003, "logs/relay_4")
	relay_5_stdout := run_relay(2004, "logs/relay_5")

	server_backend4_stdout := run_make("dev-server-backend4", "logs/server_backend4")
	server_backend5_stdout := run_make("dev-server-backend5", "logs/server_backend5")

	analytics_stdout := run_make("dev-analytics", "logs/analytics")

	// make sure all processes we create get cleaned up

	defer func() {
		for i := range processes {
			processes[i].Signal(syscall.SIGTERM)
		}
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
			strings.Contains(relay_backend_1_stdout.String(), "relay backend is ready") &&
			strings.Contains(relay_backend_1_stdout.String(), "received relay update for 'local.0'") &&
			strings.Contains(relay_backend_1_stdout.String(), "received relay update for 'local.1'") &&
			strings.Contains(relay_backend_1_stdout.String(), "received relay update for 'local.2'") &&
			strings.Contains(relay_backend_1_stdout.String(), "received relay update for 'local.3'") &&
			strings.Contains(relay_backend_1_stdout.String(), "received relay update for 'local.4'") {
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
			strings.Contains(relay_backend_2_stdout.String(), "relay backend is ready") &&
			strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.0'") &&
			strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.1'") &&
			strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.2'") &&
			strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.3'") &&
			strings.Contains(relay_backend_2_stdout.String(), "received relay update for 'local.4'") {
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
		if strings.Contains(server_backend5_stdout.String(), "started http server on port 45000") &&
			strings.Contains(server_backend5_stdout.String(), "started udp server on port 45000") &&
			strings.Contains(server_backend5_stdout.String(), "updated route matrix: 10 relays") {
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

	server4_stdout := run_make("dev-server4", "logs/server4")
	server5_stdout := run_make("dev-server5", "logs/server5")

	fmt.Printf("\n")

	// initialize server4

	fmt.Printf("initializing server 4\n")

	server4_initialized := false

	for i := 0; i < 100; i++ {
		if strings.Contains(server5_stdout.String(), "welcome to network next :)") &&
			strings.Contains(server5_stdout.String(), "server is ready to receive client connections") {
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

	client4_stdout := run_make("dev-client4", "logs/client4")
	client5_stdout := run_make("dev-client5", "logs/client5")

	fmt.Printf("\n")

	// initialize client4

	fmt.Printf("initializing client 4\n")

	client4_initialized := false

	for i := 0; i < 30; i++ {
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

	time.Sleep(time.Hour)

	return 0
}

func main() {
	setupPubsubAndBigquery()
	if happy_path() != 0 {
		os.Exit(1)
	}
}

var googleProjectID string
var bigqueryDataset string

// todo: lots of copy pasta below. surely we could simplify this by having a struct for message types?

var costMatrixPubsubTopic string
var costMatrixPubsubSubscription string
var costMatrixStatsTable string

var routeMatrixPubsubTopic string
var routeMatrixPubsubSubscription string
var routeMatrixStatsTable string

var billingPubsubTopic string
var billingPubsubSubscription string
var billingTable string

var summaryPubsubTopic string
var summaryPubsubSubscription string
var summaryTable string

var pingStatsPubsubTopic string
var pingStatsPubsubSubscription string
var pingStatsTable string

var relayStatsPubsubTopic string
var relayStatsPubsubSubscription string
var relayStatsTable string

func setupPubsubAndBigquery() {

	// todo: what if I want to run the services for the happy path manually?
	// ideally this function would be something that would automatically be done
	// when we run the pubsub emulator or biguery emulators from the makefile
	// vs. happy path specific

	// set up pubsub emulator

	run_make("dev-pubsub-emulator", "logs/pubsub_emulator")

	// set up bigquery emulator
	
	run_make("dev-bigquery-emulator", "logs/bigquery_emulator")

	time.Sleep(time.Second * 5)

	ctx := context.Background()

	googleProjectID = envvar.GetString("GOOGLE_PROJECT_ID", "local")

	pubsubSetupClient, err := pubsub.NewClient(ctx, googleProjectID)
	if err != nil {
		core.Error("failed to create pubsub setup client: %v", err)
		os.Exit(1)
	}

	// todo: simplify the code below

	costMatrixPubsubTopic = envvar.GetString("COST_MATRIX_STATS_PUBSUB_TOPIC", "local")
	costMatrixPubsubSubscription = envvar.GetString("COST_MATRIX_STATS_PUBSUB_SUBSCRIPTION", "local")

	pubsubSetupClient.CreateTopic(ctx, costMatrixPubsubTopic)
	pubsubSetupClient.CreateSubscription(ctx, costMatrixPubsubSubscription, pubsub.SubscriptionConfig{
		Topic: pubsubSetupClient.Topic(costMatrixPubsubTopic),
	})

	routeMatrixPubsubTopic = envvar.GetString("ROUTE_MATRIX_STATS_PUBSUB_TOPIC", "local")
	routeMatrixPubsubSubscription = envvar.GetString("ROUTE_MATRIX_STATS_PUBSUB_SUBSCRIPTION", "local")

	pubsubSetupClient.CreateTopic(ctx, routeMatrixPubsubTopic)
	pubsubSetupClient.CreateSubscription(ctx, routeMatrixPubsubSubscription, pubsub.SubscriptionConfig{
		Topic: pubsubSetupClient.Topic(routeMatrixPubsubTopic),
	})

	pingStatsPubsubTopic = envvar.GetString("PING_STATS_PUBSUB_TOPIC", "local")
	pingStatsPubsubSubscription = envvar.GetString("PING_STATS_PUBSUB_SUBSCRIPTION", "local")

	pubsubSetupClient.CreateTopic(ctx, pingStatsPubsubTopic)
	pubsubSetupClient.CreateSubscription(ctx, pingStatsPubsubSubscription, pubsub.SubscriptionConfig{
		Topic: pubsubSetupClient.Topic(pingStatsPubsubTopic),
	})

	relayStatsPubsubTopic = envvar.GetString("RELAY_STATS_PUBSUB_TOPIC", "local")
	relayStatsPubsubSubscription = envvar.GetString("RELAY_STATS_PUBSUB_SUBSCRIPTION", "local")

	pubsubSetupClient.CreateTopic(ctx, relayStatsPubsubTopic)
	pubsubSetupClient.CreateSubscription(ctx, relayStatsPubsubSubscription, pubsub.SubscriptionConfig{
		Topic: pubsubSetupClient.Topic(relayStatsPubsubTopic),
	})

	billingPubsubTopic = envvar.GetString("BILLING_PUBSUB_TOPIC", "local")
	billingPubsubSubscription = envvar.GetString("BILLING_PUBSUB_SUBSCRIPTION", "local")

	pubsubSetupClient.CreateTopic(ctx, billingPubsubTopic)
	pubsubSetupClient.CreateSubscription(ctx, billingPubsubSubscription, pubsub.SubscriptionConfig{
		Topic: pubsubSetupClient.Topic(billingPubsubTopic),
	})

	summaryPubsubTopic = envvar.GetString("SUMMARY_PUBSUB_TOPIC", "local")
	summaryPubsubSubscription = envvar.GetString("SUMMARY_PUBSUB_SUBSCRIPTION", "local")

	pubsubSetupClient.CreateTopic(ctx, summaryPubsubTopic)
	pubsubSetupClient.CreateSubscription(ctx, summaryPubsubSubscription, pubsub.SubscriptionConfig{
		Topic: pubsubSetupClient.Topic(summaryPubsubTopic),
	})

	pubsubSetupClient.Close()

	// ----------------

	bigqueryDataset = envvar.GetString("BIGQUERY_DATASET", "local")

	costMatrixStatsTable = envvar.GetString("COST_MATRIX_STATS_BIGQUERY_TABLE", "cost_matrix_stats")
	routeMatrixStatsTable = envvar.GetString("ROUTE_MATRIX_STATS_BIGQUERY_TABLE", "route_matrix_stats")
	relayStatsTable = envvar.GetString("RELAY_STATS_BIGQUERY_TABLE", "relay_stats")
	pingStatsTable = envvar.GetString("PING_STATS_BIGQUERY_TABLE", "ping_stats")
	billingTable = envvar.GetString("BILLING_BIGQUERY_TABLE", "billing")
	summaryTable = envvar.GetString("SUMMARY_BIGQUERY_TABLE", "summary")

	clientOptions := []option.ClientOption{
		option.WithEndpoint("http://127.0.0.1:9050"),
		option.WithoutAuthentication(),
	}

	bigquerySetupClient, err := bigquery.NewClient(ctx, googleProjectID, clientOptions...)
	if err != nil {
		core.Error("failed to create bigquery setup client: %v", err)
		os.Exit(1)
	}

	// Create local tables under the local dataset
	costMatrixStatsTableRef := bigquerySetupClient.Dataset(bigqueryDataset).Table(costMatrixStatsTable)
	routeMatrixStatsTableRef := bigquerySetupClient.Dataset(bigqueryDataset).Table(routeMatrixStatsTable)
	pingStatsTableRef := bigquerySetupClient.Dataset(bigqueryDataset).Table(pingStatsTable)
	relayStatsTableRef := bigquerySetupClient.Dataset(bigqueryDataset).Table(relayStatsTable)
	billingStatsTableRef := bigquerySetupClient.Dataset(bigqueryDataset).Table(billingTable)
	summaryTableRef := bigquerySetupClient.Dataset(bigqueryDataset).Table(summaryTable)

	costMatrixStatsTableRef.Create(ctx, &bigquery.TableMetadata{
		Schema: bigquery.Schema{
			{
				Name: "timestamp",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "bytes",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "numRelays",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "numDestRelays",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "numDatacenters",
				Type: bigquery.IntegerFieldType,
			},
		},
	})

	routeMatrixStatsTableRef.Create(ctx, &bigquery.TableMetadata{
		Schema: bigquery.Schema{
			{
				Name: "timestamp",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "bytes",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "numRelays",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "numDestRelays",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "numFullRelays",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "numDatacenters",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "totalRoutes",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "averageNumRoutes",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "averageRouteLength",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "noRoutePercent",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "oneRoutePercent",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "noDirectRoutePercent",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_NoImprovement",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_0_5ms",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_5_10ms",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_10_15ms",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_15_20ms",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_20_25ms",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_25_30ms",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_30_35ms",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_35_40ms",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_40_45ms",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_45_50ms",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "rttBucket_50ms_Plus",
				Type: bigquery.IntegerFieldType,
			},
		},
	})

	pingStatsTableRef.Create(ctx, &bigquery.TableMetadata{
		Schema: bigquery.Schema{
			{
				Name:     "timestamp",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "relay_a",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "relay_b",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "rtt",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "jitter",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "packet_loss",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "routable",
				Type:     bigquery.BooleanFieldType,
				Required: true,
			},
		},
	})

	relayStatsTableRef.Create(ctx, &bigquery.TableMetadata{
		Schema: bigquery.Schema{
			{
				Name:     "timestamp",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "relay_id",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "cpu_percent",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "memory_percent",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "actual_bandwidth_send_percent",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "actual_bandwidth_receive_percent",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "envelope_bandwidth_send_percent",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "envelope_bandwidth_receive_percent",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "actual_bandwidth_send_mbps",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "actual_bandwidth_receive_mbps",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "envelope_bandwidth_send_mbps",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "envelope_bandwidth_receive_mbps",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "num_sessions",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "max_sessions",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "num_routable",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "num_unroutable",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name: "num_unroutable",
				Type: bigquery.BooleanFieldType,
			},
		},
	})

	billingStatsTableRef.Create(ctx, &bigquery.TableMetadata{
		Schema: bigquery.Schema{
			{
				Name:     "sessionID",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name: "datacenterID",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "buyerID",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "userHash",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "envelopeBytesUp",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "envelopeBytesDown",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "latitude",
				Type: bigquery.FloatFieldType,
			},
			{
				Name: "longitude",
				Type: bigquery.FloatFieldType,
			},
			{
				Name: "clientAddress",
				Type: bigquery.StringFieldType,
			},
			{
				Name: "serverAddress",
				Type: bigquery.StringFieldType,
			},
			{
				Name: "isp",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "connectionType",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "platformType",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "sdkVersion",
				Type: bigquery.StringFieldType,
			},
			{
				Name:     "tags",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name: "abTest",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "pro",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "clientToServerPacketsSent",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "serverToClientPacketsSent",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "clientToServerPacketsLost",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "serverToClientPacketsLost",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "clientToServerPacketsOutOfOrder",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "serverToClientPacketsOutOfOrder",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "serverToClientPacketsSent",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name:     "nearRelayIDs",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name:     "nearRelayRTTs",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name:     "nearRelayJitters",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name:     "nearRelayPacketLosses",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name: "everOnNext",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "sessionDuration",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "totalPriceSum",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "envelopeBytesUpSum",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "envelopeBytesDownSum",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "durationOnNext",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "startTimestamp",
				Type: bigquery.IntegerFieldType,
			},
		},
	})

	summaryTableRef.Create(ctx, &bigquery.TableMetadata{
		Schema: bigquery.Schema{
			{
				Name:     "timestamp",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "sessionID",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "sliceNumber",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "directRTT",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "directMaxRTT",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "directPrimeRTT",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "directJitter",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "directPacketLoss",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name:     "realPacketLoss",
				Type:     bigquery.FloatFieldType,
				Required: true,
			},
			{
				Name:     "realJitter",
				Type:     bigquery.IntegerFieldType,
				Required: true,
			},
			{
				Name: "next",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "flagged",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "summary",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "debug",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "routeDiversity",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "userFlags",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "tryBeforeYouBuy",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "datacenterID",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "buyerID",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "userHash",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "envelopeBytesUp",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "envelopeBytesDown",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "latitude",
				Type: bigquery.FloatFieldType,
			},
			{
				Name: "longitude",
				Type: bigquery.FloatFieldType,
			},
			{
				Name: "clientAddress",
				Type: bigquery.StringFieldType,
			},
			{
				Name: "serverAddress",
				Type: bigquery.StringFieldType,
			},
			{
				Name: "isp",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "connectionType",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "platformType",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "sdkVersion",
				Type: bigquery.StringFieldType,
			},
			{
				Name:     "tags",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name: "abTest",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "pro",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "clientToServerPacketsSent",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "serverToClientPacketsSent",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "clientToServerPacketsLost",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "serverToClientPacketsLost",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "clientToServerPacketsOutOfOrder",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "serverToClientPacketsOutOfOrder",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "serverToClientPacketsSent",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name:     "nearRelayIDs",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name:     "nearRelayRTTs",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name:     "nearRelayJitters",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name:     "nearRelayPacketLosses",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name: "everOnNext",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "sessionDuration",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "totalPriceSum",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "envelopeBytesUpSum",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "envelopeBytesDownSum",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "durationOnNext",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "startTimestamp",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "nextRTT",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "nextJitter",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "nextPacketLoss",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "predictedNextRTT",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "nearRelayRTT",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name:     "nextRelays",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name:     "nextRelayPrice",
				Type:     bigquery.IntegerFieldType,
				Repeated: true,
			},
			{
				Name: "totalPrice",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "uncommitted",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "multipath",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "rttReduction",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "packetLossReduction",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "routeChanged",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "nextBytesUp",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "nextBytesDown",
				Type: bigquery.IntegerFieldType,
			},
			{
				Name: "fallbackToDirect",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "multipathVetoed",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "mispredicted",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "vetoed",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "latencyWorse",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "noRoute",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "nextLatencyTooHigh",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "commitVeto",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "nextLatencyTooHigh",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "unknownDatacenter",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "datacenterNotEnabled",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "buyerNotLive",
				Type: bigquery.BooleanFieldType,
			},
			{
				Name: "staleRouteMatrix",
				Type: bigquery.BooleanFieldType,
			},
		},
	})

	bigquerySetupClient.Close()
}
