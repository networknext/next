/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/base64"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/stats"
	"github.com/networknext/backend/transport"

	gcplogging "cloud.google.com/go/logging"

	"cloud.google.com/go/firestore"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/oschwald/geoip2-golang"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

func main() {
	ctx := context.Background()

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)
	{
		switch os.Getenv("BACKEND_LOG_LEVEL") {
		case "none":
			logger = level.NewFilter(logger, level.AllowNone())
		case level.ErrorValue().String():
			logger = level.NewFilter(logger, level.AllowError())
		case level.WarnValue().String():
			logger = level.NewFilter(logger, level.AllowWarn())
		case level.InfoValue().String():
			logger = level.NewFilter(logger, level.AllowInfo())
		case level.DebugValue().String():
			logger = level.NewFilter(logger, level.AllowDebug())
		default:
			logger = level.NewFilter(logger, level.AllowWarn())
		}

		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}
	if projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		loggingClient, err := gcplogging.NewClient(ctx, projectID)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		logger = logging.NewStackdriverLogger(loggingClient, "relay-backend")
	}

	var customerPublicKey []byte
	{
		if key := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(key) != 0 {
			customerPublicKey, _ = base64.StdEncoding.DecodeString(key)
			customerPublicKey = customerPublicKey[8:]
		}
	}

	var relayPublicKey []byte
	{
		if key := os.Getenv("RELAY_PUBLIC_KEY"); len(key) != 0 {
			relayPublicKey, _ = base64.StdEncoding.DecodeString(key)
		}
	}

	var routerPrivateKey []byte
	{
		if key := os.Getenv("RELAY_ROUTER_PRIVATE_KEY"); len(key) != 0 {
			routerPrivateKey, _ = base64.StdEncoding.DecodeString(key)
		} else {
			level.Error(logger).Log("err", "RELAY_ROUTER_PRIVATE_KEY not set")
			os.Exit(1)
		}
	}

	redisHost := os.Getenv("REDIS_HOST_RELAYS")
	redisClientRelays := storage.NewRedisClient(redisHost)
	if err := redisClientRelays.Ping().Err(); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_RELAYS", "value", redisHost, "err", err)
		os.Exit(1)
	}

	var ipLocator routing.IPLocator = routing.NullIsland
	if uri, ok := os.LookupEnv("MAXMIND_DB_URI"); ok {
		mmreader, err := geoip2.Open(uri)
		if err != nil {
			level.Error(logger).Log("envvar", "MAXMIND_DB_URI", "value", uri, "err", err)
		}
		ipLocator = &routing.MaxmindDB{
			Reader: mmreader,
		}
		defer mmreader.Close()
	}

	geoClient := routing.GeoClient{
		RedisClient: redisClientRelays,
		Namespace:   "RELAY_LOCATIONS",
	}

	// Create an in-memory relay & datacenter store
	// that doesn't require talking to configstore
	var db storage.Storer = &storage.InMemory{
		LocalRelays: []routing.Relay{
			routing.Relay{
				PublicKey: relayPublicKey,
				Datacenter: routing.Datacenter{
					ID:   crypto.HashID("local"),
					Name: "local",
				},
				Seller: routing.Seller{
					Name:              "local",
					IngressPriceCents: 10,
					EgressPriceCents:  20,
				},
			},
		},
	}

	// Create a no-op relay traffic stats publisher
	var trafficStatsPublisher stats.Publisher = &stats.NoOpTrafficStatsPublisher{}

	// Create a local metrics handler
	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpProjectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		firestoreClient, err := firestore.NewClient(ctx, gcpProjectID)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// Create a Firestore Storer
		fs := storage.Firestore{
			Client: firestoreClient,
			Logger: logger,
		}

		// Start a goroutine to sync from Firestore
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			fs.SyncLoop(ctx, ticker.C)
		}()

		// Set the Firestore Storer to give to handlers
		db = &fs

		if trafficStatsTopicID, ok := os.LookupEnv("GOOGLE_PUBSUB_TOPIC_TRAFFIC_STATS"); ok {
			t, err := stats.NewTrafficStatsPublisher(ctx, logger, gcpProjectID, trafficStatsTopicID, &billing.Descriptor{
				ClientCount:         4,
				DelayThreshold:      time.Millisecond,
				CountThreshold:      1024 / 4, // max relays / number of clients
				ByteThreshold:       1e6,
				NumGoroutines:       (25 * runtime.GOMAXPROCS(0)) / 4,
				Timeout:             time.Minute,
				ResultChannelBuffer: 1024 * 60 * 10, // 1,024 messages per second for 10 minutes
			})

			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			// Set the Publisher to the Pub/Sub version
			trafficStatsPublisher = t
		}

		// Set up StackDriver metrics
		sd := metrics.StackDriverHandler{
			ProjectID:          gcpProjectID,
			OverwriteFrequency: time.Second,
			OverwriteTimeout:   10 * time.Second,
		}

		if err := sd.Open(ctx); err != nil {
			level.Error(logger).Log("msg", "Failed to create StackDriver metrics client", "err", err)
			os.Exit(1)
		}

		metricsHandler = &sd

		go func() {
			metricsHandler.WriteLoop(ctx, logger, time.Minute, 200)
		}()
	}

	// Create relay init metrics
	relayInitMetrics, err := metrics.NewRelayInitMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create relay init metrics", "err", err)
	}

	// Create relay update metrics
	relayUpdateMetrics, err := metrics.NewRelayUpdateMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create relay update metrics", "err", err)
	}

	// Create relay stat metrics
	relayStatMetrics, err := metrics.NewRelayStatMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create relay stat metrics", "err", err)
	}

	statsdb := routing.NewStatsDatabase()
	var costmatrix routing.CostMatrix
	var routematrix routing.RouteMatrix

	// Periodically generate cost matrix from stats db
	go func() {
		for {
			if err := statsdb.GetCostMatrix(&costmatrix, redisClientRelays); err != nil {
				level.Warn(logger).Log("matrix", "cost", "op", "generate", "err", err)
			}

			relayStatMetrics.NumRelays.Set(float64(len(statsdb.Entries)))

			if err := costmatrix.Optimize(&routematrix, 1); err != nil {
				level.Warn(logger).Log("matrix", "cost", "op", "optimize", "err", err)
			}

			relayStatMetrics.NumRoutes.Set(float64(len(routematrix.Entries)))

			level.Info(logger).Log("matrix", "route", "entries", len(routematrix.Entries))

			time.Sleep(10 * time.Second)
		}
	}()

	// Sub to expiry events for cleanup
	redisClientRelays.ConfigSet("notify-keyspace-events", "Ex")
	go func() {
		ps := redisClientRelays.Subscribe("__keyevent@0__:expired")
		for {
			// Recieve expiry event message
			msg, err := ps.ReceiveMessage()
			if err != nil {
				level.Error(logger).Log("msg", "Error recieving expired message from pubsub", "err", err)
				os.Exit(1)
			}

			// If it is a relay that is expiring...
			if strings.HasPrefix(msg.Payload, routing.HashKeyPrefixRelay) {

				// Retrieve the ID of the relay that has expired
				rawID, err := strconv.ParseUint(strings.TrimPrefix(msg.Payload, routing.HashKeyPrefixRelay), 10, 64)
				if err != nil {
					level.Error(logger).Log("msg", "Failed to parse expired Relay ID from payload", "payload", msg.Payload, "err", err)
					os.Exit(1)
				}

				// Log the ID
				level.Warn(logger).Log("msg", fmt.Sprintf("relay with id %v has disconnected.", rawID))

				// Remove geo location data associated with this relay
				if err := geoClient.Remove(rawID); err != nil {
					level.Error(logger).Log("msg", fmt.Sprintf("Failed to remove geoClient entry for relay with ID %v", rawID), "err", err)
					os.Exit(1)
				}

				// Remove relay entry from Hashmap
				if err := redisClientRelays.HDel(routing.HashKeyAllRelays, msg.Payload).Err(); err != nil {
					level.Error(logger).Log("msg", fmt.Sprintf("Failed to remove hashmap entry for relay with ID %v", rawID), "err", err)
					os.Exit(1)
				}

				// Remove relay entry from statsDB (which in turn means it won't appear in cost matrix)
				statsdb.DeleteEntry(rawID)
			}
		}
	}()

	commonInitParams := transport.RelayInitHandlerConfig{
		RedisClient:      redisClientRelays,
		GeoClient:        &geoClient,
		IpLocator:        ipLocator,
		Storer:           db,
		Metrics:          relayInitMetrics,
		RouterPrivateKey: routerPrivateKey,
	}

	commonUpdateParams := transport.RelayUpdateHandlerConfig{
		RedisClient:           redisClientRelays,
		StatsDb:               statsdb,
		Metrics:               relayUpdateMetrics,
		TrafficStatsPublisher: trafficStatsPublisher,
		Storer:                db,
	}

	router := mux.NewRouter()
	router.HandleFunc("/healthz", transport.HealthzHandlerFunc())
	router.HandleFunc("/relay_init", transport.RelayInitHandlerFunc(logger, &commonInitParams)).Methods("POST")
	router.HandleFunc("/relay_update", transport.RelayUpdateHandlerFunc(logger, &commonUpdateParams)).Methods("POST")
	router.Handle("/cost_matrix", &costmatrix).Methods("GET")
	router.Handle("/route_matrix", &routematrix).Methods("GET")
	router.Handle("/debug/vars", expvar.Handler())

	go func() {
		port, ok := os.LookupEnv("PORT")
		if !ok {
			level.Error(logger).Log("err", "env var PORT must be set")
			os.Exit(1)
		}

		level.Info(logger).Log("addr", ":"+port)

		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
