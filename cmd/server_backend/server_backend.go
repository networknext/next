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
	"runtime"
	"sync"

	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
	googlepubsub "cloud.google.com/go/pubsub"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {

	fmt.Printf("Welcome to the nerd zone 2.0\n")

	// fmt.Printf("server_backend: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	// Configure local logging
	logger := log.NewLogfmtLogger(os.Stdout)

	// StackDriver Logging
	{
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
					level.Error(logger).Log("msg", "failed to create GCP logging client", "err", err)
					os.Exit(1)
				}

				logger = logging.NewStackdriverLogger(loggingClient, "server-backend")
			}
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
	}

	// Get env
	env, ok := os.LookupEnv("ENV")
	if !ok {
		level.Error(logger).Log("err", "ENV not set")
		os.Exit(1)
	}

	var customerPublicKey []byte
	var serverPrivateKey []byte
	var routerPrivateKey []byte
	var err error
	{
		if key := os.Getenv("SERVER_BACKEND_PRIVATE_KEY"); len(key) != 0 {
			serverPrivateKey, err = base64.StdEncoding.DecodeString(key)
			if err != nil {
				level.Error(logger).Log("envvar", "SERVER_BACKEND_PRIVATE_KEY", "msg", "could not parse", "err", err)
				os.Exit(1)
			}
		} else {
			level.Error(logger).Log("err", "SERVER_BACKEND_PRIVATE_KEY not set")
			os.Exit(1)
		}

		if key := os.Getenv("RELAY_ROUTER_PRIVATE_KEY"); len(key) != 0 {
			routerPrivateKey, err = base64.StdEncoding.DecodeString(key)
			if err != nil {
				level.Error(logger).Log("envvar", "RELAY_ROUTER_PRIVATE_KEY", "msg", "could not parse", "err", err)
				os.Exit(1)
			}
		} else {
			level.Error(logger).Log("err", "RELAY_ROUTER_PRIVATE_KEY not set")
			os.Exit(1)
		}

		if key := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(key) != 0 {
			customerPublicKey, err = base64.StdEncoding.DecodeString(key)
			if err != nil {
				level.Error(logger).Log("envvar", "NEXT_CUSTOMER_PUBLIC_KEY", "msg", "could not parse", "err", err)
				os.Exit(1)
			}
			customerPublicKey = customerPublicKey[8:]
		}
	}

	redisHost := os.Getenv("REDIS_HOST_RELAYS")
	redisClientRelays := storage.NewRedisClient(redisHost)
	if err := redisClientRelays.Ping().Err(); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_RELAYS", "value", redisHost, "msg", "could not ping", "err", err)
		os.Exit(1)
	}

	// Create an in-memory db
	var db storage.Storer = &storage.InMemory{
		LocalMode: true,
	}

	// Create a no-op biller
	var biller billing.Biller = &billing.NoOpBiller{}

	// Create a no-op metrics handler
	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	gcpProjectID, gcpOK := os.LookupEnv("GOOGLE_PROJECT_ID")
	_, firestoreEmulatorOK := os.LookupEnv("FIRESTORE_EMULATOR_HOST")
	if firestoreEmulatorOK {
		gcpProjectID = "local"

		level.Info(logger).Log("msg", "Detected firestore emulator")
	}

	if gcpOK || firestoreEmulatorOK {
		// Firestore
		{
			// Create a Firestore Storer
			fs, err := storage.NewFirestore(ctx, gcpProjectID, logger)
			if err != nil {
				level.Error(logger).Log("msg", "could not create firestore", "err", err)
				os.Exit(1)
			}

			fssyncinterval := os.Getenv("GOOGLE_FIRESTORE_SYNC_INTERVAL")
			syncInterval, err := time.ParseDuration(fssyncinterval)
			if err != nil {
				level.Error(logger).Log("envvar", "GOOGLE_FIRESTORE_SYNC_INTERVAL", "value", fssyncinterval, "msg", "could not parse", "err", err)
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

	// Create dummy buyer and datacenter for local testing
	if env == "local" {
		if err := db.AddBuyer(ctx, routing.Buyer{
			ID:                   13672574147039585173,
			Name:                 "local",
			Live:                 true,
			PublicKey:            customerPublicKey,
			RoutingRulesSettings: routing.LocalRoutingRulesSettings,
		}); err != nil {
			level.Error(logger).Log("msg", "could not add buyer to storage", "err", err)
			os.Exit(1)
		}
		if err := db.AddDatacenter(ctx, routing.Datacenter{
			ID:      crypto.HashID("local"),
			Name:    "local",
			Enabled: true,
		}); err != nil {
			level.Error(logger).Log("msg", "could not add datacenter to storage", "err", err)
			os.Exit(1)
		}
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpOK {
		// StackDriver Metrics
		{
			var enableSDMetrics bool
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

		// StackDriver Profiler
		{
			var enableSDProfiler bool
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
					Service:        "server_backend",
					ServiceVersion: env,
					ProjectID:      gcpProjectID,
					MutexProfiling: true,
				}); err != nil {
					level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
					os.Exit(1)
				}
			}
		}
	}

	// Create server init metrics
	serverInitMetrics, err := metrics.NewServerInitMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create server init metrics", "err", err)
	}

	// Create server update metrics
	serverUpdateMetrics, err := metrics.NewServerUpdateMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create server update metrics", "err", err)
	}

	// Create session update metrics
	sessionUpdateMetrics, err := metrics.NewSessionMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create session metrics", "err", err)
	}

	// Create maxmindb sync metrics
	maxmindSyncMetrics, err := metrics.NewMaxmindSyncMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create session metrics", "err", err)
	}

	// Create server backend metrics
	serverBackendMetrics, err := metrics.NewServerBackendMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create server backend metrics", "err", err)
	}

	_, pubsubEmulatorOK := os.LookupEnv("PUBSUB_EMULATOR_HOST")
	if gcpOK || pubsubEmulatorOK {

		pubsubCtx := ctx
		if pubsubEmulatorOK {
			gcpProjectID = "local"

			var cancelFunc context.CancelFunc
			pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			level.Info(logger).Log("msg", "Detected pubsub emulator")
		}

		// Google Pubsub
		{
			settings := googlepubsub.PublishSettings{
				DelayThreshold: time.Hour,
				CountThreshold: 100,
				ByteThreshold:  60 * 1024,
				NumGoroutines:  runtime.GOMAXPROCS(0),
				Timeout:        time.Minute,
			}

			pubsub, err := billing.NewGooglePubSubBiller(pubsubCtx, &serverBackendMetrics.BillingMetrics, logger, gcpProjectID, "billing", 1, 100, 1024, &settings)
			if err != nil {
				level.Error(logger).Log("msg", "could not create pubsub biller", "err", err)
				os.Exit(1)
			}

			biller = pubsub
		}
	}

	getIPLocatorFunc := func() routing.IPLocator {
		return routing.NullIsland
	}

	// Open the Maxmind DB and create a routing.MaxmindDB from it
	mmcitydburi := os.Getenv("MAXMIND_CITY_DB_URI")
	mmispdburi := os.Getenv("MAXMIND_ISP_DB_URI")
	if mmcitydburi != "" && mmispdburi != "" {
		mmdb := &routing.MaxmindDB{
			HTTPClient: http.DefaultClient,
			CityURI:    mmcitydburi,
			IspURI:     mmispdburi,
		}
		var mmdbMutex sync.RWMutex

		getIPLocatorFunc = func() routing.IPLocator {
			mmdbMutex.RLock()
			defer mmdbMutex.RUnlock()

			mmdbRet := mmdb
			return mmdbRet
		}

		if err := mmdb.Sync(ctx, maxmindSyncMetrics); err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		if mmsyncinterval, ok := os.LookupEnv("MAXMIND_SYNC_DB_INTERVAL"); ok {
			syncInterval, err := time.ParseDuration(mmsyncinterval)
			if err != nil {
				level.Error(logger).Log("envvar", "MAXMIND_SYNC_DB_INTERVAL", "value", mmsyncinterval, "msg", "could not parse", "err", err)
				os.Exit(1)
			}

			// Start a goroutine to sync from Maxmind.com
			go func() {
				ticker := time.NewTicker(syncInterval)
				for {
					newMMDB := &routing.MaxmindDB{}

					select {
					case <-ticker.C:
						if err := newMMDB.Sync(ctx, maxmindSyncMetrics); err != nil {
							level.Error(logger).Log("err", err)
							continue
						}

						// Pointer swap the mmdb so we can sync from Maxmind.com lock free
						mmdbMutex.Lock()
						mmdb = newMMDB
						mmdbMutex.Unlock()
					case <-ctx.Done():
						return
					}

					time.Sleep(syncInterval)
				}
			}()
		}
	}

	routeMatrix := &routing.RouteMatrix{}
	var routeMatrixMutex sync.RWMutex

	getRouteMatrixFunc := func() transport.RouteProvider {
		routeMatrixMutex.RLock()
		rm := routeMatrix
		routeMatrixMutex.RUnlock()
		return rm
	}

	// Sync route matrix
	{
		if uri, ok := os.LookupEnv("ROUTE_MATRIX_URI"); ok {
			rmsyncinterval := os.Getenv("ROUTE_MATRIX_SYNC_INTERVAL")
			syncInterval, err := time.ParseDuration(rmsyncinterval)
			if err != nil {
				level.Error(logger).Log("envvar", "ROUTE_MATRIX_SYNC_INTERVAL", "value", rmsyncinterval, "msg", "could not parse", "err", err)
				os.Exit(1)
			}

			go func() {
				for {
					newRouteMatrix := routing.RouteMatrix{}
					var matrixReader io.Reader

					// Default to reading route matrix from file
					if f, err := os.Open(uri); err == nil {
						matrixReader = f
					}

					// Prefer to get it remotely if possible
					if r, err := http.Get(uri); err == nil {
						matrixReader = r.Body
					}

					start := time.Now()

					// Don't swap route matrix if we fail to read
					routeMatrixBytes, err := newRouteMatrix.ReadFrom(matrixReader)
					if err != nil {
						if env != "local" {
							level.Warn(logger).Log("envvar", "ROUTE_MATRIX_URI", "value", uri, "msg", "could not read route matrix", "err", err)
						}
						time.Sleep(syncInterval)
						continue
					}

					routeMatrixTime := time.Since(start)

					serverBackendMetrics.RouteMatrixUpdateDuration.Set(float64(routeMatrixTime.Milliseconds()))

					if routeMatrixTime.Seconds() > 1.0 {
						serverBackendMetrics.LongRouteMatrixUpdateCount.Add(1)
					}

					// Swap the route matrix pointer to the new one
					// This double buffered route matrix approach makes the route matrix lockless
					routeMatrixMutex.Lock()
					routeMatrix = &newRouteMatrix
					routeMatrixMutex.Unlock()

					serverBackendMetrics.RouteMatrixBytes.Set(float64(routeMatrixBytes))

					time.Sleep(syncInterval)
				}
			}()
		}
	}

	udpPortString, ok := os.LookupEnv("UDP_PORT")
	if !ok {
		level.Error(logger).Log("err", "env var UDP_PORT must be set")
		os.Exit(1)
	}

	udpPort, err := strconv.ParseInt(udpPortString, 10, 64)
	if err != nil {
		level.Error(logger).Log("envvar", "UDP_PORT", "msg", "could not parse", "err", err)
		os.Exit(1)
	}

	vetoMap := transport.NewVetoMap()
	serverMap := transport.NewServerMap()
	sessionMap := transport.NewSessionMap()
	{
		// todo: ryan, please add the number of iterations to perform each check to each map timeout func below. currently hardcoded.

		// Start a goroutine to timeout vetoes
		go func() {
			timeout := int64(60 * 5)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			vetoMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()

		// Start a goroutine to timeout servers
		go func() {
			timeout := int64(30)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			serverMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()

		// Start a goroutine to timeout sessions
		go func() {
			timeout := int64(30)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			sessionMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()
	}

	// Initialize the datacenter tracker
	datacenterTracker := transport.NewDatacenterTracker()
	go func() {
		timeout := time.Minute
		frequency := time.Millisecond * 10
		ticker := time.NewTicker(frequency)
		datacenterTracker.TimeoutLoop(ctx, timeout, ticker.C)
	}()

	// Setup the stats print routine
	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				serverBackendMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				serverBackendMetrics.MemoryAllocated.Set(memoryUsed())

				numVetoes := vetoMap.NumVetoes()
				serverBackendMetrics.VetoCount.Set(float64(numVetoes))

				numServers := serverMap.NumServers()
				serverBackendMetrics.ServerCount.Set(float64(numServers))

				numSessions := sessionMap.NumSessions()
				serverBackendMetrics.SessionCount.Set(float64(numSessions))

				numEntriesQueued := serverBackendMetrics.BillingMetrics.EntriesSubmitted.Value() - serverBackendMetrics.BillingMetrics.EntriesFlushed.Value()
				serverBackendMetrics.BillingMetrics.EntriesQueued.Set(numEntriesQueued)

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%.2f mb allocated\n", serverBackendMetrics.MemoryAllocated.Value())
				fmt.Printf("%d goroutines\n", int(serverBackendMetrics.Goroutines.Value()))
				fmt.Printf("%d vetoes\n", numVetoes)
				fmt.Printf("%d servers\n", numServers)
				fmt.Printf("%d sessions\n", numSessions)
				fmt.Printf("%d billing entries submitted\n", int(serverBackendMetrics.BillingMetrics.EntriesSubmitted.Value()))
				fmt.Printf("%d billing entries queued\n", int(serverBackendMetrics.BillingMetrics.EntriesQueued.Value()))
				fmt.Printf("%d billing entries flushed\n", int(serverBackendMetrics.BillingMetrics.EntriesFlushed.Value()))
				fmt.Printf("%d server init packets processed\n", int(serverInitMetrics.Invocations.Value()))
				fmt.Printf("%d server update packets processed\n", int(serverUpdateMetrics.Invocations.Value()))
				fmt.Printf("%d session update packets processed\n", int(sessionUpdateMetrics.Invocations.Value()))
				fmt.Printf("%d long route matrix updates\n", int(serverBackendMetrics.LongRouteMatrixUpdateCount.Value()))
				fmt.Printf("route matrix update: %.2f milliseconds\n", serverBackendMetrics.RouteMatrixUpdateDuration.Value())
				fmt.Printf("route matrix bytes: %d\n", int(serverBackendMetrics.RouteMatrixBytes.Value()))

				if env != "local" {
					unknownDatacentersLength := datacenterTracker.UnknownDatacenterLength()
					serverBackendMetrics.UnknownDatacenterCount.Set(float64(unknownDatacentersLength))
					if unknownDatacentersLength > 0 {
						fmt.Printf("unknown datacenters: %v\n", datacenterTracker.GetUnknownDatacenters())
					}

					emptyDatacentersLength := datacenterTracker.EmptyDatacenterLength()
					serverBackendMetrics.EmptyDatacenterCount.Set(float64(emptyDatacentersLength))
					if emptyDatacentersLength > 0 {
						fmt.Printf("empty datacenters: %v\n", datacenterTracker.GetEmptyDatacenters())
					}
				}

				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second)
			}
		}()
	}

	// Start portal cruncher publisher
	var portalPublisher pubsub.Publisher
	{
		portalCruncherHost, ok := os.LookupEnv("PORTAL_CRUNCHER_HOST")
		if !ok {
			level.Error(logger).Log("err", "env var PORTAL_CRUNCHER_HOST must be set")
			os.Exit(1)
		}

		portalCruncherPublisher, err := pubsub.NewPortalCruncherPublisher(portalCruncherHost)
		if err != nil {
			level.Error(logger).Log("msg", "could not create portal cruncher publisher", "err", err)
			os.Exit(1)
		}

		portalPublisher = portalCruncherPublisher
	}

	// Start UDP server
	{
		serverInitConfig := &transport.ServerInitParams{
			ServerPrivateKey:  serverPrivateKey,
			Storer:            db,
			Metrics:           serverInitMetrics,
			Logger:            logger,
			DatacenterTracker: datacenterTracker,
		}

		serverUpdateConfig := &transport.ServerUpdateParams{
			Storer:            db,
			Metrics:           serverUpdateMetrics,
			Logger:            logger,
			ServerMap:         serverMap,
			DatacenterTracker: datacenterTracker,
		}

		sessionUpdateConfig := &transport.SessionUpdateParams{
			ServerPrivateKey:  serverPrivateKey,
			RouterPrivateKey:  routerPrivateKey,
			GetRouteProvider:  getRouteMatrixFunc,
			GetIPLocator:      getIPLocatorFunc,
			Storer:            db,
			Biller:            biller,
			Metrics:           sessionUpdateMetrics,
			Logger:            logger,
			VetoMap:           vetoMap,
			ServerMap:         serverMap,
			SessionMap:        sessionMap,
			DatacenterTracker: datacenterTracker,
			PortalPublisher:   portalPublisher,
		}

		mux := transport.UDPServerMux2{
			Logger:                   logger,
			Port:                     udpPort,
			MaxPacketSize:            transport.DefaultMaxPacketSize,
			ServerInitHandlerFunc:    transport.ServerInitHandlerFunc(serverInitConfig),
			ServerUpdateHandlerFunc:  transport.ServerUpdateHandlerFunc(serverUpdateConfig),
			SessionUpdateHandlerFunc: transport.SessionUpdateHandlerFunc(sessionUpdateConfig),
		}

		go func() {
			if err := mux.Start(ctx); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}()
	}

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))
		router.Handle("/debug/vars", expvar.Handler())

		go func() {
			httpPort, ok := os.LookupEnv("HTTP_PORT")
			if !ok {
				level.Error(logger).Log("err", "env var HTTP_PORT must be set")
				os.Exit(1)
			}

			level.Info(logger).Log("addr", ":"+httpPort)

			err := http.ListenAndServe(":"+httpPort, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}
		}()
	}

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
