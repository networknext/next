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
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
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

// MaxRelayCount is the maximum number of relays you can run locally with the firestore emulator
const MaxRelayCount = 10

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

	if enableSDLogging {
		if projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
			loggingClient, err := gcplogging.NewClient(ctx, projectID)
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

	fmt.Printf("env is %s\n", env)

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

	var db storage.Storer = &storage.InMemory{
		LocalMode: true,
	}

	gcpProjectID, gcpOK := os.LookupEnv("GOOGLE_PROJECT_ID")
	_, emulatorOK := os.LookupEnv("FIRESTORE_EMULATOR_HOST")
	if emulatorOK {
		fmt.Printf("using firestore emulator\n")
		gcpProjectID = "local"
		level.Info(logger).Log("msg", "Detected firestore emulator")
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpOK || emulatorOK {

		// Firestore
		{
			fmt.Printf("initializing firestore\n")

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
	}

	if env == "local" {

		fmt.Printf("creating dummy local seller\n")

		seller := routing.Seller{
			ID:                        "sellerID",
			Name:                      "local",
			IngressPriceNibblinsPerGB: 0.1 * 1e9,
			EgressPriceNibblinsPerGB:  0.5 * 1e9,
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

		for i := int64(0); i < MaxRelayCount; i++ {
			addressString := "127.0.0.1:1000" + strconv.FormatInt(i, 10)
			addr, err := net.ResolveUDPAddr("udp", addressString)
			if err != nil {
				level.Error(logger).Log("msg", "could parse udp address", "address", addressString, "err", err)
				os.Exit(1)
			}

			if err := db.AddRelay(ctx, routing.Relay{
				ID:          crypto.HashID(addr.String()),
				Name:        "", // needs to be blank so the relay_dashboard shows ips and the stats
				Addr:        *addr,
				PublicKey:   relayPublicKey,
				Seller:      seller,
				Datacenter:  datacenter,
				MaxSessions: 3000,
			}); err != nil {
				level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
				os.Exit(1)
			}
		}
	}

	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	if gcpOK {
		// Stackdriver Metrics
		{
			fmt.Printf("initializing stackdriver metrics\n")

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
			fmt.Printf("initializing stackdriver profiler\n")

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

	costMatrixMetrics, err := metrics.NewCostMatrixMetrics(context.Background(), metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create cost matrix metrics", "err", err)
	}

	optimizeMetrics, err := metrics.NewOptimizeMetrics(context.Background(), metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create optimize metrics", "err", err)
	}

	routeMatrixMetrics, err := metrics.NewRouteMatrixMetrics(context.Background(), metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create route matrix metrics", "err", err)
	}

	relayBackendMetrics, err := metrics.NewRelayBackendMetrics(context.Background(), metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create relay backend metrics", "err", err)
	}

	statsdb := routing.NewStatsDatabase()
	costMatrix := &routing.CostMatrix{}
	routeMatrix := &routing.RouteMatrix{}
	var routeMatrixMutex sync.RWMutex

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		routeMatrixMutex.RLock()
		rm := routeMatrix
		routeMatrixMutex.RUnlock()
		return rm
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

	// Create the relay map
	cleanupCallback := func(relayID uint64) error {
		// Remove relay entry from statsDB (which in turn means it won't appear in cost matrix)
		statsdb.DeleteEntry(relayID)

		return nil
	}

	relayMap := routing.NewRelayMap(cleanupCallback)
	go func() {
		timeout := int64(routing.RelayTimeout.Seconds())
		frequency := time.Millisecond * 100
		ticker := time.NewTicker(frequency)
		relayMap.TimeoutLoop(ctx, timeout, ticker.C)
	}()

	// ping stats
	var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
	{
		// Create a no-op publisher
		_, emulatorOK = os.LookupEnv("PUBSUB_EMULATOR_HOST")
		if gcpOK || emulatorOK {

			pubsubCtx := ctx
			if emulatorOK {
				gcpProjectID = "local"

				var cancelFunc context.CancelFunc
				pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(5*time.Minute))
				defer cancelFunc()

				level.Info(logger).Log("msg", "Detected pubsub emulator")
			}

			// Google Pubsub
			{
				settings := pubsub.PublishSettings{
					DelayThreshold: time.Second,
					CountThreshold: 1,
					ByteThreshold:  1 << 14,
					NumGoroutines:  runtime.GOMAXPROCS(0),
					Timeout:        time.Minute,
				}

				pubsub, err := analytics.NewGooglePubSubPingStatsPublisher(pubsubCtx, &relayBackendMetrics.PingStatsMetrics, logger, gcpProjectID, "ping_stats", settings)
				if err != nil {
					level.Error(logger).Log("msg", "could not create analytics pubsub publisher", "err", err)
					os.Exit(1)
				}

				pingStatsPublisher = pubsub
			}
		}

		go func() {
			sleepTime := time.Minute
			if publishInterval, ok := os.LookupEnv("PING_STATS_PUBLISH_INTERVAL"); ok {
				if duration, err := time.ParseDuration(publishInterval); err == nil {
					sleepTime = duration
				} else {
					level.Error(logger).Log("msg", "could not parse publish interval", "err", err)
				}
			}

			for {
				time.Sleep(sleepTime)
				cpy := statsdb.MakeCopy()
				length := routing.TriMatrixLength(len(cpy.Entries))
				if length > 0 { // prevent crash with only 1 relay
					entries := make([]analytics.PingStatsEntry, length)
					ids := make([]uint64, length)

					idx := 0
					for k := range cpy.Entries {
						ids[idx] = k
						idx++
					}

					for i := 1; i < len(cpy.Entries); i++ {
						for j := 0; j < i; j++ {
							idA := ids[i]
							idB := ids[j]

							rtt, jitter, pl := cpy.GetSample(idA, idB)

							entries[routing.TriMatrixIndex(i, j)] = analytics.PingStatsEntry{
								RelayA:     idA,
								RelayB:     idB,
								RTT:        rtt,
								Jitter:     jitter,
								PacketLoss: pl,
							}
						}
					}

					if err := pingStatsPublisher.Publish(ctx, entries); err != nil {
						level.Error(logger).Log("err", err)
						os.Exit(1)
					}
				}
			}
		}()
	}

	// relay stats
	var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
	{
		_, emulatorOK = os.LookupEnv("PUBSUB_EMULATOR_HOST")
		if gcpOK || emulatorOK {

			pubsubCtx := ctx
			if emulatorOK {
				gcpProjectID = "local"

				var cancelFunc context.CancelFunc
				pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(5*time.Minute))
				defer cancelFunc()

				level.Info(logger).Log("msg", "Detected pubsub emulator")
			}

			// Google Pubsub
			{
				settings := pubsub.PublishSettings{
					DelayThreshold: time.Second,
					CountThreshold: 1,
					ByteThreshold:  1 << 14,
					NumGoroutines:  runtime.GOMAXPROCS(0),
					Timeout:        time.Minute,
				}

				pubsub, err := analytics.NewGooglePubSubRelayStatsPublisher(pubsubCtx, &relayBackendMetrics.RelayStatsMetrics, logger, gcpProjectID, "relay_stats", settings)
				if err != nil {
					level.Error(logger).Log("msg", "could not create analytics pubsub publisher", "err", err)
					os.Exit(1)
				}

				relayStatsPublisher = pubsub
			}
		}

		go func() {
			sleepTime := time.Second * 10
			if publishInterval, ok := os.LookupEnv("RELAY_STATS_PUBLISH_INTERVAL"); ok {
				if duration, err := time.ParseDuration(publishInterval); err == nil {
					sleepTime = duration
				} else {
					level.Error(logger).Log("msg", "could not parse publish interval", "err", err)
				}
			}

			for {
				time.Sleep(sleepTime)

				allRelayData := relayMap.GetAllRelayData()

				entries := make([]analytics.RelayStatsEntry, len(allRelayData))

				for i, relay := range allRelayData {
					entries[i] = analytics.RelayStatsEntry{
						ID:          relay.ID,
						NumSessions: relay.TrafficStats.SessionCount,
						CPUUsage:    relay.CPUUsage,
						MemUsage:    relay.MemUsage,
						Tx:          relay.TrafficStats.BytesSent,
						Rx:          relay.TrafficStats.BytesReceived,
					}
				}

				if err := relayStatsPublisher.Publish(ctx, entries); err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}
			}
		}()

	}

	// Periodically generate cost matrix from stats db
	cmsyncinterval := os.Getenv("COST_MATRIX_INTERVAL")
	syncInterval, err := time.ParseDuration(cmsyncinterval)
	if err != nil {
		level.Error(logger).Log("envvar", "COST_MATRIX_INTERVAL", "value", cmsyncinterval, "err", err)
		os.Exit(1)
	}
	go func() {
		for {
			costMatrixMetrics.Invocations.Add(1)

			costMatrixNew := routing.CostMatrix{}

			costMatrixDurationStart := time.Now()

			err := statsdb.GetCostMatrix(&costMatrixNew, relayMap.GetAllRelayData(), float32(maxJitter), float32(maxPacketLoss))

			costMatrixDurationSince := time.Since(costMatrixDurationStart)
			costMatrixMetrics.DurationGauge.Set(float64(costMatrixDurationSince.Milliseconds()))
			if costMatrixDurationSince.Seconds() > 1.0 {
				costMatrixMetrics.LongUpdateCount.Add(1)
			}

			// todo: we need to handle this better in future, but just hold the previous cost matrix for the moment on error
			if err == nil {
				costMatrix = &costMatrixNew
			} else {
				costMatrixMetrics.ErrorMetrics.GenFailure.Add(1)
			}

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

			newRouteMatrix := &routing.RouteMatrix{}

			optimizeMetrics.Invocations.Add(1)

			optimizeDurationStart := time.Now()
			if err := costMatrix.Optimize(newRouteMatrix, 1); err != nil {
				level.Warn(logger).Log("matrix", "cost", "op", "optimize", "err", err)
			}
			optimizeDurationSince := time.Since(optimizeDurationStart)
			optimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

			if optimizeDurationSince.Seconds() > 1.0 {
				optimizeMetrics.LongUpdateCount.Add(1)
			}

			// Write the cost matrix to a buffer and serve that instead
			// of writing a new buffer every time we want to serve the cost matrix
			err = costMatrix.WriteResponseData()
			if err != nil {
				level.Error(logger).Log("matrix", "cost", "msg", "failed to write cost matrix response data", "err", err)
				continue // Don't store the new route matrix if we fail to write cost matrix data
			}

			// Write the route matrix to a buffer and serve that instead
			// of writing a new buffer every time we want to serve the route matrix
			err = newRouteMatrix.WriteResponseData()
			if err != nil {
				level.Error(logger).Log("matrix", "route", "msg", "failed to write route matrix response data", "err", err)
				continue // Don't store the new route matrix if we fail to write response data
			}

			// Write the route matrix analysis to a buffer and serve that instead
			// of writing a new analysis every time we want to view the analysis in the relay dashboard
			newRouteMatrix.WriteAnalysisData()

			// Swap the route matrix pointer to the new one
			// This double buffered route matrix approach makes the route matrix lockless
			routeMatrixMutex.Lock()
			routeMatrix = newRouteMatrix
			routeMatrixMutex.Unlock()

			costMatrixMetrics.Bytes.Set(float64(len(costMatrix.GetResponseData())))

			routeMatrixMetrics.Bytes.Set(float64(len(newRouteMatrix.GetResponseData())))
			routeMatrixMetrics.RelayCount.Set(float64(len(newRouteMatrix.RelayIDs)))
			routeMatrixMetrics.DatacenterCount.Set(float64(len(newRouteMatrix.DatacenterIDs)))

			// todo: calculate this in optimize and store in route matrix so we don't have to calc this here
			numRoutes := int32(0)
			for i := range newRouteMatrix.Entries {
				numRoutes += newRouteMatrix.Entries[i].NumRoutes
			}
			routeMatrixMetrics.RouteCount.Set(float64(numRoutes))

			memoryUsed := func() float64 {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				return float64(m.Alloc) / (1000.0 * 1000.0)
			}

			relayBackendMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
			relayBackendMetrics.MemoryAllocated.Set(memoryUsed())

			fmt.Printf("-----------------------------\n")
			fmt.Printf("%.2f mb allocated\n", relayBackendMetrics.MemoryAllocated.Value())
			fmt.Printf("%d goroutines\n", int(relayBackendMetrics.Goroutines.Value()))
			fmt.Printf("%d datacenters\n", int(routeMatrixMetrics.DatacenterCount.Value()))
			fmt.Printf("%d relays\n", int(routeMatrixMetrics.RelayCount.Value()))
			fmt.Printf("%d routes\n", int(routeMatrixMetrics.RouteCount.Value()))
			fmt.Printf("%d long cost matrix updates\n", int(costMatrixMetrics.LongUpdateCount.Value()))
			fmt.Printf("%d long route matrix updates\n", int(optimizeMetrics.LongUpdateCount.Value()))
			fmt.Printf("cost matrix update: %.2f milliseconds\n", costMatrixMetrics.DurationGauge.Value())
			fmt.Printf("route matrix update: %.2f milliseconds\n", optimizeMetrics.DurationGauge.Value())
			fmt.Printf("cost matrix bytes: %d\n", int(costMatrixMetrics.Bytes.Value()))
			fmt.Printf("route matrix bytes: %d\n", int(routeMatrixMetrics.Bytes.Value()))
			fmt.Printf("%d ping stats entries submitted\n", int(relayBackendMetrics.PingStatsMetrics.EntriesSubmitted.Value()))
			fmt.Printf("%d ping stats entries queued\n", int(relayBackendMetrics.PingStatsMetrics.EntriesQueued.Value()))
			fmt.Printf("%d ping stats entries flushed\n", int(relayBackendMetrics.PingStatsMetrics.EntriesFlushed.Value()))
			fmt.Printf("%d relay stats entries submitted\n", int(relayBackendMetrics.RelayStatsMetrics.EntriesSubmitted.Value()))
			fmt.Printf("%d relay stats entries queued\n", int(relayBackendMetrics.RelayStatsMetrics.EntriesQueued.Value()))
			fmt.Printf("%d relay stats entries flushed\n", int(relayBackendMetrics.RelayStatsMetrics.EntriesFlushed.Value()))
			fmt.Printf("-----------------------------\n")

			time.Sleep(syncInterval)
		}
	}()

	commonInitParams := transport.RelayInitHandlerConfig{
		RelayMap:         relayMap,
		Storer:           db,
		Metrics:          relayInitMetrics,
		RouterPrivateKey: routerPrivateKey,
	}

	commonUpdateParams := transport.RelayUpdateHandlerConfig{
		RelayMap: relayMap,
		StatsDb:  statsdb,
		Metrics:  relayUpdateMetrics,
		Storer:   db,
	}

	commonHandlerParams := transport.RelayHandlerConfig{
		RelayMap:         relayMap,
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

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))
	router.HandleFunc("/relay_init", transport.RelayInitHandlerFunc(logger, &commonInitParams)).Methods("POST")
	router.HandleFunc("/relay_update", transport.RelayUpdateHandlerFunc(logger, relayslogger, &commonUpdateParams)).Methods("POST")
	router.HandleFunc("/relays", transport.RelayHandlerFunc(logger, relayslogger, &commonHandlerParams)).Methods("POST")
	router.Handle("/cost_matrix", costMatrix).Methods("GET")
	router.HandleFunc("/route_matrix", serveRouteMatrixFunc).Methods("GET")
	router.Handle("/debug/vars", expvar.Handler())
	router.HandleFunc("/relay_dashboard", transport.RelayDashboardHandlerFunc(relayMap, getRouteMatrixFunc, statsdb, "local", "local", maxJitter))
	router.HandleFunc("/routes", transport.RoutesHandlerFunc(getRouteMatrixFunc, statsdb, "local", "local"))

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
