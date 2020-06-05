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

	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/stats"
	"github.com/networknext/backend/transport"

	gcplogging "cloud.google.com/go/logging"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

var (
	release string
)

func main() {
	ctx := context.Background()

	env := os.Getenv("ENV")

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)
	relayslogger := log.NewLogfmtLogger(os.Stdout)
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

		switch os.Getenv("RELAYS_LOG_LEVEL") {
		case "none":
			relayslogger = level.NewFilter(relayslogger, level.AllowNone())
		case level.ErrorValue().String():
			relayslogger = level.NewFilter(relayslogger, level.AllowError())
		case level.WarnValue().String():
			relayslogger = level.NewFilter(relayslogger, level.AllowWarn())
		case level.InfoValue().String():
			relayslogger = level.NewFilter(relayslogger, level.AllowInfo())
		case level.DebugValue().String():
			relayslogger = level.NewFilter(relayslogger, level.AllowDebug())
		default:
			relayslogger = level.NewFilter(relayslogger, level.AllowWarn())
		}
		relayslogger = log.With(relayslogger, "ts", log.DefaultTimestampUTC)
	}
	if projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		loggingClient, err := gcplogging.NewClient(ctx, projectID)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		logger = logging.NewStackdriverLogger(loggingClient, "relay-backend")
		relayslogger = logging.NewStackdriverLogger(loggingClient, "relays")
	}

	sentryOpts := sentry.ClientOptions{
		ServerName:       "Relay Backend",
		Release:          release,
		Dist:             "linux",
		AttachStacktrace: true,
		Debug:            true,
	}

	if err := sentry.Init(sentryOpts); err != nil {
		level.Error(logger).Log("msg", "failed to initialize sentry", "err", err)
		os.Exit(1)
	}

	// force sentry to post any updates upon program exit
	defer sentry.Flush(time.Second * 2)

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

	geoClient := routing.GeoClient{
		RedisClient: redisClientRelays,
		Namespace:   "RELAY_LOCATIONS",
	}

	// Create an in-memory relay & datacenter store
	// that doesn't require talking to configstore
	var db storage.Storer = &storage.InMemory{
		LocalMode: true,
	}

	{
		seller := routing.Seller{
			ID:                "sellerID",
			Name:              "local",
			IngressPriceCents: 10,
			EgressPriceCents:  20,
		}

		datacenter := routing.Datacenter{
			ID:       crypto.HashID("local"),
			Name:     "local",
			Location: routing.LocationNullIsland,
		}

		if err := db.AddSeller(ctx, seller); err != nil {
			level.Error(logger).Log("msg", "could not add seller to storage", "err", err)
			os.Exit(1)
		}

		if err := db.AddDatacenter(ctx, datacenter); err != nil {
			level.Error(logger).Log("msg", "could not add datacenter to storage", "err", err)
			os.Exit(1)
		}

		if err := db.AddRelay(ctx, routing.Relay{
			Name:        "", // needs to be blank so the relay_dashboard shows ips and the stats
			PublicKey:   relayPublicKey,
			Seller:      seller,
			Datacenter:  datacenter,
			MaxSessions: 3000,
		}); err != nil {
			level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
			os.Exit(1)
		}
	}

	// Create a no-op relay traffic stats publisher
	var trafficStatsPublisher stats.Publisher = &stats.NoOpTrafficStatsPublisher{}

	// Create a local metrics handler
	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpProjectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		// Create a Firestore Storer
		fs, err := storage.NewFirestore(ctx, gcpProjectID, logger)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// Start a goroutine to sync from Firestore
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			fs.SyncLoop(ctx, ticker.C)
		}()

		// Set the Firestore Storer to give to handlers
		db = fs

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

	// Create relay handler metrics
	relayHandlerMetrics, err := metrics.NewRelayHandlerMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create relay handler metrics", "err", err)
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

			if len(routematrix.Entries) == 0 {
				sentry.CaptureMessage("no routes within route matrix")
			}

			// todo: should be one
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

				// Remove the relay cache entry
				if err := transport.RemoveRelayCacheEntry(ctx, rawID, msg.Payload, env, http.DefaultClient, redisClientRelays, &geoClient, statsdb, db); err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}
			}
		}
	}()

	commonInitParams := transport.RelayInitHandlerConfig{
		RedisClient:      redisClientRelays,
		GeoClient:        &geoClient,
		Storer:           db,
		Metrics:          relayInitMetrics,
		RouterPrivateKey: routerPrivateKey,
	}

	commonUpdateParams := transport.RelayUpdateHandlerConfig{
		Environment:           env,
		HTTPClient:            http.DefaultClient,
		RedisClient:           redisClientRelays,
		GeoClient:             &geoClient,
		StatsDb:               statsdb,
		Metrics:               relayUpdateMetrics,
		TrafficStatsPublisher: trafficStatsPublisher,
		Storer:                db,
	}

	commonHandlerParams := transport.RelayHandlerConfig{
		Environment:           env,
		HTTPClient:            http.DefaultClient,
		RedisClient:           redisClientRelays,
		GeoClient:             &geoClient,
		Storer:                db,
		StatsDb:               statsdb,
		TrafficStatsPublisher: trafficStatsPublisher,
		Metrics:               relayHandlerMetrics,
		RouterPrivateKey:      routerPrivateKey,
	}

	router := mux.NewRouter()
	router.HandleFunc("/healthz", transport.HealthzHandlerFunc())
	router.HandleFunc("/relay_init", transport.RelayInitHandlerFunc(logger, &commonInitParams)).Methods("POST")
	router.HandleFunc("/relay_update", transport.RelayUpdateHandlerFunc(logger, relayslogger, &commonUpdateParams)).Methods("POST")
	router.HandleFunc("/relays", transport.RelayHandlerFunc(logger, relayslogger, &commonHandlerParams)).Methods("POST")
	router.Handle("/cost_matrix", &costmatrix).Methods("GET")
	router.Handle("/route_matrix", &routematrix).Methods("GET")
	router.Handle("/debug/vars", expvar.Handler())
	router.HandleFunc("/relay_dashboard", transport.RelayDashboardHandlerFunc(redisClientRelays, &routematrix, statsdb, "local", "local"))
	router.HandleFunc("/routes", transport.RoutesHandlerFunc(redisClientRelays, &routematrix, statsdb, "local", "local"))

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
