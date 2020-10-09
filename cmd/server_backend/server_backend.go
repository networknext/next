/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/base64"
	"encoding/binary"
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
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
	googlepubsub "cloud.google.com/go/pubsub"

	metadataapi "cloud.google.com/go/compute/metadata"
)

// MaxRelayCount is the maximum number of relays you can run locally with the firestore emulator
// An equal number of valve relays will also be added
const MaxRelayCount = 10

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {

	fmt.Printf("server_backend: Git Hash: %s - Commit: %s\n", sha, commitMessage)

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

	fmt.Printf("env is %s\n", env)

	var customerPublicKey []byte
	var customerID uint64
	var serverPrivateKey []byte
	var routerPrivateKey []byte
	var relayPublicKey []byte
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

		if env == "local" {
			if key := os.Getenv("RELAY_PUBLIC_KEY"); len(key) != 0 {
				relayPublicKey, err = base64.StdEncoding.DecodeString(key)
				if err != nil {
					level.Error(logger).Log("envvar", "RELAY_PUBLIC_KEY", "msg", "could not parse", "err", err)
					os.Exit(1)
				}
			} else {
				level.Error(logger).Log("err", "RELAY_PUBLIC_KEY not set")
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

			if key := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(key) != 0 {
				customerPublicKey, _ = base64.StdEncoding.DecodeString(key)
				customerID = binary.LittleEndian.Uint64(customerPublicKey[:8])
				customerPublicKey = customerPublicKey[8:]
			}
		}
	}

	gcpProjectID, gcpOK := os.LookupEnv("GOOGLE_PROJECT_ID")

	// var db storage.Storer
	db, err := storage.NewFirestore(ctx, gcpProjectID, logger)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
	// Create a no-op biller
	var biller billing.Biller = &billing.NoOpBiller{}

	// Create a no-op metrics handler
	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	// Create dummy buyer and datacenter for local testing
	if env == "local" {
		if err = storage.SeedStorage(logger, ctx, db, relayPublicKey, customerID, customerPublicKey); err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	var instanceID uint64
	if gcpOK {
		// Get the instance number of this server_backend instance
		{
			instanceIDString, err := metadataapi.InstanceID()
			if err != nil {
				level.Error(logger).Log("msg", "could not read instance id from GCP", "err", err)
				os.Exit(1)
			}

			instanceIDInt, err := strconv.Atoi(instanceIDString)
			if err != nil {
				level.Error(logger).Log("msg", "could not parse instance id", "id", instanceIDString, "err", err)
				os.Exit(1)
			}

			instanceID = uint64(instanceIDInt)
		}

		// StackDriver Metrics
		{
			fmt.Printf("setting up stackdriver metrics\n")

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
			fmt.Printf("setting up stackdriver profiler\n")

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
			fmt.Printf("setting up pubsub emulator\n")

			gcpProjectID = "local"

			var cancelFunc context.CancelFunc
			pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			level.Info(logger).Log("msg", "Detected pubsub emulator")
		}

		// Google Pubsub
		{
			fmt.Printf("setting up pubsub\n")

			clientCount, err := strconv.Atoi(os.Getenv("BILLING_CLIENT_COUNT"))
			if err != nil {
				level.Error(logger).Log("envvar", "BILLING_CLIENT_COUNT", "msg", "could not parse", "err", err)
				os.Exit(1)
			}

			countThreshold, err := strconv.Atoi(os.Getenv("BILLING_BATCHED_MESSAGE_COUNT"))
			if err != nil {
				level.Error(logger).Log("envvar", "BILLING_BATCHED_MESSAGE_COUNT", "msg", "could not parse", "err", err)
				os.Exit(1)
			}

			byteThreshold, err := strconv.Atoi(os.Getenv("BILLING_BATCHED_MESSAGE_MIN_BYTES"))
			if err != nil {
				level.Error(logger).Log("envvar", "BILLING_BATCHED_MESSAGE_MIN_BYTES", "msg", "could not parse", "err", err)
				os.Exit(1)
			}

			// We do our own batching so don't stack the library's batching on top of ours
			// Specifically, don't stack the message count thresholds
			settings := googlepubsub.DefaultPublishSettings
			settings.CountThreshold = 1
			settings.ByteThreshold = byteThreshold
			settings.NumGoroutines = runtime.GOMAXPROCS(0)

			pubsub, err := billing.NewGooglePubSubBiller(pubsubCtx, &serverBackendMetrics.BillingMetrics, logger, gcpProjectID, "billing", clientCount, countThreshold, byteThreshold, &settings)
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

		fmt.Printf("setting up maxmind ip2location\n")

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

		// todo: disable the sync for now until we can find out why it's causing session drops

		// if mmsyncinterval, ok := os.LookupEnv("MAXMIND_SYNC_DB_INTERVAL"); ok {
		// 	syncInterval, err := time.ParseDuration(mmsyncinterval)
		// 	if err != nil {
		// 		level.Error(logger).Log("envvar", "MAXMIND_SYNC_DB_INTERVAL", "value", mmsyncinterval, "msg", "could not parse", "err", err)
		// 		os.Exit(1)
		// 	}

		// 	// Start a goroutine to sync from Maxmind.com
		// 	go func() {
		// 		ticker := time.NewTicker(syncInterval)
		// 		for {
		// 			newMMDB := &routing.MaxmindDB{}

		// 			select {
		// 			case <-ticker.C:
		// 				if err := newMMDB.Sync(ctx, maxmindSyncMetrics); err != nil {
		// 					level.Error(logger).Log("err", err)
		// 					continue
		// 				}

		// 				// Pointer swap the mmdb so we can sync from Maxmind.com lock free
		// 				mmdbMutex.Lock()
		// 				mmdb = newMMDB
		// 				mmdbMutex.Unlock()
		// 			case <-ctx.Done():
		// 				return
		// 			}

		// 			time.Sleep(syncInterval)
		// 		}
		// 	}()
		// }
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
				httpClient := &http.Client{
					Timeout: time.Second * 2,
				}
				for {
					newRouteMatrix := routing.RouteMatrix{}
					var matrixReader io.Reader

					// Default to reading route matrix from file
					if f, err := os.Open(uri); err == nil {
						matrixReader = f
					}

					// Prefer to get it remotely if possible
					if r, err := httpClient.Get(uri); err == nil {
						matrixReader = r.Body
						// todo: need to close the response body!
					}

					start := time.Now()

					bytes, err := newRouteMatrix.ReadFrom(matrixReader)
					if err != nil {
						if env != "local" {
							level.Warn(logger).Log("envvar", "ROUTE_MATRIX_URI", "value", uri, "msg", "could not read route matrix", "err", err)
						}
						time.Sleep(syncInterval)
						continue // Don't swap route matrix if we fail to read
					}

					routeMatrixTime := time.Since(start)

					serverBackendMetrics.RouteMatrixUpdateDuration.Set(float64(routeMatrixTime.Milliseconds()))

					if routeMatrixTime.Seconds() > 1.0 {
						serverBackendMetrics.LongRouteMatrixUpdateCount.Add(1)
					}

					serverBackendMetrics.RouteMatrix.RelayCount.Set(float64(len(newRouteMatrix.RelayIDs)))
					serverBackendMetrics.RouteMatrix.DatacenterCount.Set(float64(len(newRouteMatrix.DatacenterIDs)))

					// todo: calculate this in optimize and store in route matrix so we don't have to calc this here
					numRoutes := int32(0)
					for i := range newRouteMatrix.Entries {
						numRoutes += newRouteMatrix.Entries[i].NumRoutes
					}
					serverBackendMetrics.RouteMatrix.RouteCount.Set(float64(numRoutes))
					serverBackendMetrics.RouteMatrix.Bytes.Set(float64(bytes))

					// Swap the route matrix pointer to the new one
					// This double buffered route matrix approach makes the route matrix lockless
					routeMatrixMutex.Lock()
					routeMatrix = &newRouteMatrix
					routeMatrixMutex.Unlock()

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
	multipathVetoMap := transport.NewVetoMap()
	serverMap := transport.NewServerMap()
	sessionMap := transport.NewSessionMap()
	{
		// Start a goroutine to timeout vetoes
		go func() {
			timeout := int64(60 * 5)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			vetoMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()

		// Start a goroutine to timeout multipath vetoes
		go func() {
			timeout := int64(60 * 60 * 24 * 7)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			multipathVetoMap.TimeoutLoop(ctx, timeout, ticker.C)
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

	// Start portal cruncher publisher
	var portalPublisher pubsub.Publisher
	{
		fmt.Printf("setting up portal cruncher\n")

		portalCruncherHost, ok := os.LookupEnv("PORTAL_CRUNCHER_HOST")
		if !ok {
			level.Error(logger).Log("err", "env var PORTAL_CRUNCHER_HOST must be set")
			os.Exit(1)
		}

		postSessionPortalSendBufferSizeString, ok := os.LookupEnv("POST_SESSION_PORTAL_SEND_BUFFER_SIZE")
		if !ok {
			level.Error(logger).Log("err", "env var POST_SESSION_PORTAL_SEND_BUFFER_SIZE must be set")
			os.Exit(1)
		}

		postSessionPortalSendBufferSize, err := strconv.ParseInt(postSessionPortalSendBufferSizeString, 10, 64)
		if err != nil {
			level.Error(logger).Log("envvar", "POST_SESSION_PORTAL_SEND_BUFFER_SIZE", "msg", "could not parse", "err", err)
			os.Exit(1)
		}

		portalCruncherPublisher, err := pubsub.NewPortalCruncherPublisher(portalCruncherHost, int(postSessionPortalSendBufferSize))
		if err != nil {
			level.Error(logger).Log("msg", "could not create portal cruncher publisher", "err", err)
			os.Exit(1)
		}

		portalPublisher = portalCruncherPublisher
	}

	numPostSessionGoroutinesString, ok := os.LookupEnv("POST_SESSION_THREAD_COUNT")
	if !ok {
		level.Error(logger).Log("err", "env var POST_SESSION_THREAD_COUNT must be set")
		os.Exit(1)
	}

	numPostSessionGoroutines, err := strconv.ParseInt(numPostSessionGoroutinesString, 10, 64)
	if err != nil {
		level.Error(logger).Log("envvar", "POST_SESSION_THREAD_COUNT", "msg", "could not parse", "err", err)
		os.Exit(1)
	}

	postSessionBufferSizeString, ok := os.LookupEnv("POST_SESSION_BUFFER_SIZE")
	if !ok {
		level.Error(logger).Log("err", "env var POST_SESSION_BUFFER_SIZE must be set")
		os.Exit(1)
	}

	postSessionBufferSize, err := strconv.ParseInt(postSessionBufferSizeString, 10, 64)
	if err != nil {
		level.Error(logger).Log("envvar", "POST_SESSION_BUFFER_SIZE", "msg", "could not parse", "err", err)
		os.Exit(1)
	}

	postSessionPortalMaxRetriesString, ok := os.LookupEnv("POST_SESSION_PORTAL_MAX_RETRIES")
	if !ok {
		level.Error(logger).Log("err", "env var POST_SESSION_PORTAL_MAX_RETRIES must be set")
		os.Exit(1)
	}

	postSessionPortalMaxRetries, err := strconv.ParseInt(postSessionPortalMaxRetriesString, 10, 64)
	if err != nil {
		level.Error(logger).Log("envvar", "POST_SESSION_PORTAL_MAX_RETRIES", "msg", "could not parse", "err", err)
		os.Exit(1)
	}

	// Create a post session handler to handle the post process of session updates.
	// This way, we can quickly return from the session update handler and not spawn a
	// ton of goroutines if things get backed up.
	postSessionHandler := transport.NewPostSessionHandler(int(numPostSessionGoroutines), int(postSessionBufferSize), portalPublisher, int(postSessionPortalMaxRetries), biller, logger, &serverBackendMetrics.PostSessionMetrics)
	postSessionHandler.StartProcessing(ctx)

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

				numVetoes := vetoMap.GetVetoCount()
				serverBackendMetrics.VetoCount.Set(float64(numVetoes))

				numMultipathVetoes := multipathVetoMap.GetVetoCount()
				serverBackendMetrics.MultipathVetoCount.Set(float64(numMultipathVetoes))

				numServers := serverMap.GetServerCount()
				serverBackendMetrics.ServerCount.Set(float64(numServers))

				numSessions := sessionMap.GetSessionCount()
				serverBackendMetrics.SessionCount.Set(float64(numSessions))

				numDirectSessions := sessionMap.GetDirectSessionCount()
				serverBackendMetrics.SessionDirectCount.Set(float64(numDirectSessions))

				numNextSessions := sessionMap.GetNextSessionCount()
				serverBackendMetrics.SessionNextCount.Set(float64(numNextSessions))

				numEntriesQueued := serverBackendMetrics.BillingMetrics.EntriesSubmitted.Value() - serverBackendMetrics.BillingMetrics.EntriesFlushed.Value()
				serverBackendMetrics.BillingMetrics.EntriesQueued.Set(numEntriesQueued)

				serverBackendMetrics.PostSessionMetrics.BillingBufferLength.Set(float64(postSessionHandler.BillingBufferSize()))
				serverBackendMetrics.PostSessionMetrics.PortalBufferLength.Set(float64(postSessionHandler.PortalBufferSize()))

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%.2f mb allocated\n", serverBackendMetrics.MemoryAllocated.Value())
				fmt.Printf("%d goroutines\n", int(serverBackendMetrics.Goroutines.Value()))
				fmt.Printf("%d vetoes\n", numVetoes)
				fmt.Printf("%d multipath vetoes\n", numMultipathVetoes)
				fmt.Printf("%d servers\n", numServers)
				fmt.Printf("%d sessions\n", numSessions)
				fmt.Printf("%d direct sessions\n", numDirectSessions)
				fmt.Printf("%d next sessions\n", numNextSessions)
				fmt.Printf("%d billing entries submitted\n", int(serverBackendMetrics.BillingMetrics.EntriesSubmitted.Value()))
				fmt.Printf("%d billing entries queued\n", int(serverBackendMetrics.BillingMetrics.EntriesQueued.Value()))
				fmt.Printf("%d billing entries flushed\n", int(serverBackendMetrics.BillingMetrics.EntriesFlushed.Value()))
				fmt.Printf("%d server init packets processed\n", int(serverInitMetrics.Invocations.Value()))
				fmt.Printf("%d server update packets processed\n", int(serverUpdateMetrics.Invocations.Value()))
				fmt.Printf("%d session update packets processed\n", int(sessionUpdateMetrics.Invocations.Value()))
				fmt.Printf("%d post session billing entries sent\n", int(serverBackendMetrics.PostSessionMetrics.BillingEntriesSent.Value()))
				fmt.Printf("%d post session billing entries queued\n", int(serverBackendMetrics.PostSessionMetrics.BillingBufferLength.Value()))
				fmt.Printf("%d post session billing entries finished\n", int(serverBackendMetrics.PostSessionMetrics.BillingEntriesFinished.Value()))
				fmt.Printf("%d post session portal entries sent\n", int(serverBackendMetrics.PostSessionMetrics.PortalEntriesSent.Value()))
				fmt.Printf("%d post session portal entries queued\n", int(serverBackendMetrics.PostSessionMetrics.PortalBufferLength.Value()))
				fmt.Printf("%d post session portal entries finished\n", int(serverBackendMetrics.PostSessionMetrics.PortalEntriesFinished.Value()))
				fmt.Printf("%d datacenters\n", int(serverBackendMetrics.RouteMatrix.DatacenterCount.Value()))
				fmt.Printf("%d relays\n", int(serverBackendMetrics.RouteMatrix.RelayCount.Value()))
				fmt.Printf("%d routes\n", int(serverBackendMetrics.RouteMatrix.RouteCount.Value()))
				fmt.Printf("%d long route matrix updates\n", int(serverBackendMetrics.LongRouteMatrixUpdateCount.Value()))
				fmt.Printf("route matrix update: %.2f milliseconds\n", serverBackendMetrics.RouteMatrixUpdateDuration.Value())
				fmt.Printf("route matrix bytes: %d\n", int(serverBackendMetrics.RouteMatrix.Bytes.Value()))

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

	// Start UDP server
	{
		fmt.Printf("starting udp server\n")

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
			MultipathVetoMap:  multipathVetoMap,
			ServerMap:         serverMap,
			SessionMap:        sessionMap,
			DatacenterTracker: datacenterTracker,
			PortalPublisher:   portalPublisher,
			InstanceID:        instanceID,
		}

		mux := transport.UDPServerMux2{
			Logger:                   logger,
			PostSessionHandler:       postSessionHandler,
			Port:                     udpPort,
			MaxPacketSize:            transport.DefaultMaxPacketSize,
			ServerInitHandlerFunc:    transport.ServerInitHandlerFunc(serverInitConfig),
			ServerUpdateHandlerFunc:  transport.ServerUpdateHandlerFunc(serverUpdateConfig),
			SessionUpdateHandlerFunc: transport.SessionUpdateHandlerFunc(sessionUpdateConfig),
		}

		var selectionPercent uint64 = 100
		if valueStr, ok := os.LookupEnv("PACKET_SELECTION_PERCENT"); ok {
			if valueUint, err := strconv.ParseUint(valueStr, 10, 64); err == nil {
				selectionPercent = valueUint
			} else {
				level.Error(logger).Log("msg", "cannot parse value of 'PACKET_SELECTION_PERCENT' env var", "err", err)
			}
		}

		go func() {
			if err := mux.Start(ctx, selectionPercent); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}()
	}

	// Start HTTP server
	{
		fmt.Printf("starting http server\n")

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
