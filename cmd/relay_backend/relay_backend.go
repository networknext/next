/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
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
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/networknext/backend/analytics"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/transport"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
	"cloud.google.com/go/pubsub"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {

	fmt.Printf("relay_backend: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)
	relayslogger := log.NewLogfmtLogger(os.Stdout)

	var enableSDLogging bool
	enableSDLoggingString, ok := os.LookupEnv("ENABLE_STACKDRIVER_LOGGING")
	if ok {
		var err error
		enableSDLogging, err = strconv.ParseBool(enableSDLoggingString)
		if err != nil {
			level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_LOGGING", "msg", "could not parse", "err", err)
			os.Exit(1)
		}
	}

	gcpOK := false
	var gcpProjectID string

	if enableSDLogging {
		if gcpProjectID, gcpOK = os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
			loggingClient, err := gcplogging.NewClient(ctx, gcpProjectID)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			logger = logging.NewStackdriverLogger(loggingClient, "relay-backend")
			relayslogger = logging.NewStackdriverLogger(loggingClient, "relays")
		}
	}

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

	// Get env
	env, ok := os.LookupEnv("ENV")
	if !ok {
		level.Error(logger).Log("err", "ENV not set")
		os.Exit(1)
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

	geoClient := routing.GeoClient{
		RedisClient: redisClientRelays,
		Namespace:   "RELAY_LOCATIONS",
	}

	// Create an in-memory relay & datacenter store
	// that doesn't require talking to configstore
	var db storage.Storer = &storage.InMemory{
		LocalMode: true,
	}

	if env == "local" {
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

	// Create a local metrics handler
	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpProjectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {

		// Firestore
		{
			// Create a Firestore Storer
			fs, err := storage.NewFirestore(ctx, gcpProjectID, logger)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			fssyncinterval := os.Getenv("GOOGLE_FIRESTORE_SYNC_INTERVAL")
			syncInterval, err := time.ParseDuration(fssyncinterval)
			if err != nil {
				level.Error(logger).Log("envvar", "GOOGLE_FIRESTORE_SYNC_INTERVAL", "value", fssyncinterval, "err", err)
				os.Exit(1)
			}
			// Start a goroutine to sync from Firestore
			go func() {
				ticker := time.NewTicker(syncInterval)
				fs.SyncLoop(ctx, ticker.C)
			}()

			// Set the Firestore Storer to give to handlers
			db = fs
		}

		// Stackdriver Metrics
		{
			var enableSDMetrics bool
			var err error
			enableSDMetricsString, ok := os.LookupEnv("ENABLE_STACKDRIVER_METRICS")
			if ok {
				enableSDMetrics, err = strconv.ParseBool(enableSDMetricsString)
				if err != nil {
					level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_METRICS", "msg", "could not parse", "err", err)
					os.Exit(1)
				}
			}

			if enableSDMetrics {
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

				sdwriteinterval := os.Getenv("GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL")
				writeInterval, err := time.ParseDuration(sdwriteinterval)
				if err != nil {
					level.Error(logger).Log("envvar", "GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL", "value", sdwriteinterval, "err", err)
					os.Exit(1)
				}
				go func() {
					metricsHandler.WriteLoop(ctx, logger, writeInterval, 200)
				}()
			}
		}

		// Stackdriver Profiler
		{
			var enableSDProfiler bool
			var err error
			enableSDProfilerString, ok := os.LookupEnv("ENABLE_STACKDRIVER_PROFILER")
			if ok {
				enableSDProfiler, err = strconv.ParseBool(enableSDProfilerString)
				if err != nil {
					level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_PROFILER", "msg", "could not parse", "err", err)
					os.Exit(1)
				}
			}

			if enableSDProfiler {
				// Set up StackDriver profiler
				if err := profiler.Start(profiler.Config{
					Service:        "relay_backend",
					ServiceVersion: env,
					ProjectID:      gcpProjectID,
					MutexProfiling: true,
				}); err != nil {
					level.Error(logger).Log("msg", "Failed to initialize StackDriver profiler", "err", err)
					os.Exit(1)
				}
			}
		}
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

	newCostMatrixGenMetrics, err := metrics.NewCostMatrixGenMetrics(context.Background(), metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create CostMatrixGenMetrics", "err", err)
	}

	newOptimizeMetrics, err := metrics.NewOptimizeMetrics(context.Background(), metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create NewOptimizeGenMetrics", "err", err)
	}

	statsdb := routing.NewStatsDatabase()
	var costMatrix routing.CostMatrix

	routeMatrix := &routing.RouteMatrix{}
	var routeMatrixMutex sync.RWMutex

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		routeMatrixMutex.RLock()
		rm := routeMatrix
		routeMatrixMutex.RUnlock()
		return rm
	}

	// Clean up any relays that may have expired while the relay_backend was down (due to a deploy, maintenance, etc.)
	hgetallResult := redisClientRelays.HGetAll(routing.HashKeyAllRelays)
	for key, raw := range hgetallResult.Val() {
		// Check if the key has expired and if it should be removed from the hash set
		getCmd := redisClientRelays.Get(key)
		if getCmd.Val() == "" {

			level.Debug(logger).Log("msg", "Found lingering relay", "key", key)

			var relay routing.RelayCacheEntry
			if err := relay.UnmarshalBinary([]byte(raw)); err != nil {
				level.Error(logger).Log("msg", "detected lingering relay but failed to unmarshal relay from redis hash set", "err", err)
				os.Exit(1)
			}

			if err := transport.RemoveRelayCacheEntry(ctx, relay.ID, key, redisClientRelays, &geoClient, statsdb); err != nil {
				level.Error(logger).Log("msg", "detected lingering relay but failed to remove relay from redis hash set", "err", err)
				os.Exit(1)
			}

			level.Debug(logger).Log("msg", "Lingering relay removed", "relay_id", relay.ID)
		}
	}

	// Get the max jitter and max packet loss env vars
	var maxJitter float64
	var maxPacketLoss float64
	{
		maxJitterString, ok := os.LookupEnv("RELAY_ROUTER_MAX_JITTER")
		if !ok {
			level.Error(logger).Log("msg", "env var not set", "envvar", "RELAY_ROUTER_MAX_JITTER")
			os.Exit(1)
		}

		maxJitter, err = strconv.ParseFloat(maxJitterString, 32)
		if err != nil {
			level.Error(logger).Log("err", "could not parse max jitter", "value", maxJitterString)
			os.Exit(1)
		}

		maxPacketLossString, ok := os.LookupEnv("RELAY_ROUTER_MAX_PACKET_LOSS")
		if !ok {
			level.Error(logger).Log("msg", "env var not set", "envvar", "RELAY_ROUTER_MAX_PACKET_LOSS")
			os.Exit(1)
		}

		maxPacketLoss, err = strconv.ParseFloat(maxPacketLossString, 32)
		if err != nil {
			level.Error(logger).Log("err", "could not parse max packet loss", "value", maxPacketLossString)
			os.Exit(1)
		}
	}

	analyticsMetrics, err := metrics.NewAnalyticsMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create statsdb metrics", "err", err)
	}

	// Create a no-op biller
	var publisher analytics.PubSubPublisher = &analytics.NoOpPubSubPublisher{}
	_, emulatorOK := os.LookupEnv("PUBSUB_EMULATOR_HOST")
	if gcpOK || emulatorOK {
		pubsubCtx := ctx

		if emulatorOK {
			gcpProjectID = "local"

			var cancelFunc context.CancelFunc
			pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			level.Info(logger).Log("msg", "Detected pubsub emulator")
		}

		// Google Pubsub
		{
			settings := pubsub.PublishSettings{
				DelayThreshold: time.Hour,
				CountThreshold: 1000,
				ByteThreshold:  60 * 1024,
				NumGoroutines:  runtime.GOMAXPROCS(0),
				Timeout:        time.Minute,
			}

			pubsub, err := analytics.NewGooglePubSubPublisher(pubsubCtx, analyticsMetrics, logger, gcpProjectID, "analytics", 1, 1000, &settings)
			if err != nil {
				level.Error(logger).Log("msg", "could not create analytics pubsub publisher", "err", err)
				os.Exit(1)
			}

			publisher = pubsub
		}
	}

	go func() {
		time.Sleep(time.Minute)
		for {
			cpy := statsdb.MakeCopy()
			length := len(cpy.Entries) * 2
			entries := make([]analytics.StatsEntry, length)

			i := 0
			for k1, s := range cpy.Entries {
				for k2, r := range s.Relays {
					entries[i] = analytics.StatsEntry{
						RelayA:     k1,
						RelayB:     k2,
						RTT:        r.RTT,
						Jitter:     r.Jitter,
						PacketLoss: r.PacketLoss,
					}

					i++
				}
			}

			publisher.Publish(ctx, entries)
		}
	}()

	// Periodically generate cost matrix from stats db
	cmsyncinterval := os.Getenv("COST_MATRIX_INTERVAL")
	syncInterval, err := time.ParseDuration(cmsyncinterval)
	if err != nil {
		level.Error(logger).Log("envvar", "COST_MATRIX_INTERVAL", "value", cmsyncinterval, "err", err)
		os.Exit(1)
	}
	go func() {

		var longCostMatrixUpdates uint64
		var longRouteMatrixUpdates uint64
		var costMatrixBytes int
		var routeMatrixBytes int

		for {

			costMatrixDurationStart := time.Now()
			err := statsdb.GetCostMatrix(&costMatrix, redisClientRelays, float32(maxJitter), float32(maxPacketLoss))
			costMatrixDurationSince := time.Since(costMatrixDurationStart)

			if costMatrixDurationSince.Seconds() > 1.0 {
				// todo: ryan, same treatment for cost matrix duration. thanks
				longCostMatrixUpdates++
			}

			if err != nil {
				level.Warn(logger).Log("matrix", "cost", "op", "generate", "err", err)
				costMatrix = routing.CostMatrix{}
			}

			costMatrixBytes = len(costMatrix.GetResponseData())

			newCostMatrixGenMetrics.DurationGauge.Set(float64(costMatrixDurationSince.Milliseconds()))

			newCostMatrixGenMetrics.Invocations.Add(1)

			// IMPORTANT: Fill the cost matrix with near relay lat/longs
			// these are then passed in to the route matrix via "Optimize"
			// and the server_backend uses them to find near relays.
			for i := range costMatrix.RelayIDs {
				relay, err := db.Relay(costMatrix.RelayIDs[i])
				if err == nil {
					costMatrix.RelayLatitude[i] = relay.Datacenter.Location.Latitude
					costMatrix.RelayLongitude[i] = relay.Datacenter.Location.Longitude
				}
			}

			relayStatMetrics.NumRelays.Set(float64(len(statsdb.Entries)))

			// todo: ryan, would be nice to upload the size of the cost matrix in bytes

			newRouteMatrix := &routing.RouteMatrix{}

			optimizeDurationStart := time.Now()
			if err := costMatrix.Optimize(newRouteMatrix, 1); err != nil {
				level.Warn(logger).Log("matrix", "cost", "op", "optimize", "err", err)
			}
			optimizeDurationSince := time.Since(optimizeDurationStart)
			newOptimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))
			newOptimizeMetrics.Invocations.Add(1)

			if optimizeDurationSince.Seconds() > 1.0 {
				longRouteMatrixUpdates++
			}

			relayStatMetrics.NumRoutes.Set(float64(len(newRouteMatrix.Entries)))

			level.Info(logger).Log("matrix", "route", "entries", len(newRouteMatrix.Entries))

			// Write the cost matrix to a buffer and serve that instead
			// of writing a new buffer every time we want to serve the cost matrix
			err = costMatrix.WriteResponseData()
			if err != nil {
				level.Error(logger).Log("matrix", "cost", "msg", "failed to write cost matrix response data", "err", err)
			}

			// Write the route matrix to a buffer and serve that instead
			// of writing a new buffer every time we want to serve the route matrix
			err = newRouteMatrix.WriteResponseData()
			if err != nil {
				level.Error(logger).Log("matrix", "route", "msg", "failed to write route matrix response data", "err", err)
				continue // Don't store the new route matrix if we fail to write response data
			}

			// todo: ryan, would be nice to upload the size of the route matrix in bytes as a metric
			routeMatrixBytes = len(routeMatrix.GetResponseData())

			// Write the route matrix analysis to a buffer and serve that instead
			// of writing a new analysis every time we want to view the analysis in the relay dashboard
			newRouteMatrix.WriteAnalysisData()

			// todo: calculate this in optimize and store in route matrix so we don't have to calc this here
			numRoutes := 0
			for i := range newRouteMatrix.Entries {
				numRoutes += int(newRouteMatrix.Entries[i].NumRoutes)
			}

			memoryUsed := func() float64 {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				return float64(m.Alloc) / (1000.0 * 1000.0)
			}

			// todo: ryan please put everything below into metrics for this service (where not already)
			fmt.Printf("-----------------------------\n")
			fmt.Printf("%d goroutines\n", runtime.NumGoroutine())
			fmt.Printf("%.2f mb allocated\n", memoryUsed())
			fmt.Printf("%d datacenters\n", len(newRouteMatrix.DatacenterIDs))
			fmt.Printf("%d relays\n", len(newRouteMatrix.RelayIDs))
			fmt.Printf("%d routes\n", numRoutes)
			fmt.Printf("%d long cost matrix updates\n", longCostMatrixUpdates)
			fmt.Printf("%d long route matrix updates\n", longRouteMatrixUpdates)
			fmt.Printf("cost matrix bytes: %d\n", costMatrixBytes)
			fmt.Printf("route matrix bytes: %d\n", routeMatrixBytes)
			fmt.Printf("cost matrix update: %.2f seconds\n", costMatrixDurationSince.Seconds())
			fmt.Printf("route matrix update: %.2f seconds\n", optimizeDurationSince.Seconds())
			fmt.Printf("%d analytics entries submitted\n", publisher.NumSubmitted())
			fmt.Printf("%d analytics entries queued\n", publisher.NumQueued())
			fmt.Printf("%d analytics entries flushed\n", publisher.NumFlushed())
			fmt.Printf("-----------------------------\n")

			// Swap the route matrix pointer to the new one
			// This double buffered route matrix approach makes the route matrix lockless
			routeMatrixMutex.Lock()
			routeMatrix = newRouteMatrix
			routeMatrixMutex.Unlock()

			time.Sleep(syncInterval)
		}
	}()

	// Sub to expiry events for cleanup
	redisClientRelays.ConfigSet("notify-keyspace-events", "Ex")
	go func() {
		ps := redisClientRelays.Subscribe("__keyevent@0__:expired")
		for {
			// Receive expiry event message
			msg, err := ps.ReceiveMessage()
			if err != nil {
				level.Error(logger).Log("msg", "Error receiving expired message from pubsub", "err", err)
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
				if err := transport.RemoveRelayCacheEntry(ctx, rawID, msg.Payload, redisClientRelays, &geoClient, statsdb); err != nil {
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
		RedisClient: redisClientRelays,
		GeoClient:   &geoClient,
		StatsDb:     statsdb,
		Metrics:     relayUpdateMetrics,
		Storer:      db,
	}

	commonHandlerParams := transport.RelayHandlerConfig{
		RedisClient:      redisClientRelays,
		GeoClient:        &geoClient,
		Storer:           db,
		StatsDb:          statsdb,
		Metrics:          relayHandlerMetrics,
		RouterPrivateKey: routerPrivateKey,
	}

	// todo: ryan, relay backend health check should only become healthy once it is ready to serve up a quality route matrix in prod.
	// in the current production environment, this probably means that it has generated route matrices for 6 minutes. the reason for this
	// being that the timeout for bad routes are 5 minutes, so when the relay backend first starts up, the route matrix is in a bad state
	// (intentionally) for the first 5 minutes, as it assumes all routes are initially bad. only routes that are good past 5 minutes
	// are good enough to serve up to our customers.

	// todo: ryan, important. make sure on the first iteration of the stats db, that we always put in, for every relay pair, 100% packet loss.
	// this will ensure the "takes 5 minutes to stabilize" policy, since the stats db is configured to take the *worst* packet loss it sees
	// for the past 5 minutes...

	serveRouteMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		m := getRouteMatrixFunc()

		data := m.GetResponseData()

		buffer := bytes.NewBuffer(data)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))
	router.HandleFunc("/relay_init", transport.RelayInitHandlerFunc(logger, &commonInitParams)).Methods("POST")
	router.HandleFunc("/relay_update", transport.RelayUpdateHandlerFunc(logger, relayslogger, &commonUpdateParams)).Methods("POST")
	router.HandleFunc("/relays", transport.RelayHandlerFunc(logger, relayslogger, &commonHandlerParams)).Methods("POST")
	router.Handle("/cost_matrix", &costMatrix).Methods("GET")
	router.HandleFunc("/route_matrix", serveRouteMatrixFunc).Methods("GET")
	router.Handle("/debug/vars", expvar.Handler())
	router.HandleFunc("/relay_dashboard", transport.RelayDashboardHandlerFunc(redisClientRelays, getRouteMatrixFunc, statsdb, "local", "local", maxJitter))
	router.HandleFunc("/routes", transport.RoutesHandlerFunc(redisClientRelays, getRouteMatrixFunc, statsdb, "local", "local"))

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
