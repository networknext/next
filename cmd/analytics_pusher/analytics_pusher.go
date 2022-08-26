package main

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/networknext/backend/modules/analytics"
	pusher "github.com/networknext/backend/modules/analytics_pusher"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"

	"github.com/gorilla/mux"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "analytics_pusher"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, commitHash, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	// Setup the service
	ctx, cancel := context.WithCancel(context.Background())
	gcpProjectID := backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		core.Error("error getting logger: %v", err)
		return 1
	}

	env := backend.GetEnv()

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("error getting metrics handler: %v", err)
		return 1
	}

	// Create analytics pusher metrics
	analyticsPusherMetrics, err := metrics.NewAnalyticsPusherMetrics(ctx, metricsHandler, serviceName)
	if err != nil {
		core.Error("failed to create analytics pusher metrics: %v", err)
		return 1
	}

	if gcpOK {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
			return 1
		}
	}

	// Determine how frequently we should publish ping and relay stats
	pingStatsPublishInterval := envvar.GetDuration("PING_STATS_PUBLISH_INTERVAL", 1*time.Minute)
	relayStatsPublishInterval := envvar.GetDuration("RELAY_STATS_PUBLISH_INTERVAL", 10*time.Second)

	// Get HTTP timeout for route matrix
	httpTimeout := envvar.GetDuration("HTTP_TIMEOUT", 4*time.Second)

	// Get route matrix URI
	routeMatrixURI := envvar.Get("ROUTE_MATRIX_URI", "")
	if routeMatrixURI == "" {
		core.Error("ROUTE_MATRIX_URI not set")
		return 1
	}

	// Get route matrix stale duration
	routeMatrixStaleDuration := envvar.GetDuration("ROUTE_MATRIX_STALE_DURATION", 20*time.Second)

	// Setup ping stats and relay stats publishers
	var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
	var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
	{
		emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
		if gcpOK || emulatorOK {

			pubsubCtx := ctx
			if emulatorOK {
				gcpProjectID = "local"

				var cancelFunc context.CancelFunc
				pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(60*time.Minute))
				defer cancelFunc()

				core.Debug("Detected pubsub emulator")
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

				pingPubsub, err := analytics.NewGooglePubSubPingStatsPublisher(pubsubCtx, &analyticsPusherMetrics.PingStatsMetrics, gcpProjectID, "ping_stats", settings)
				if err != nil {
					core.Error("could not create ping stats analytics pubsub publisher: %v", err)
					return 1
				}

				pingStatsPublisher = pingPubsub

				relayPubsub, err := analytics.NewGooglePubSubRelayStatsPublisher(pubsubCtx, &analyticsPusherMetrics.RelayStatsMetrics, gcpProjectID, "relay_stats", settings)
				if err != nil {
					core.Error("could not create relay stats analytics pubsub publisher: %v", err)
					return 1
				}

				relayStatsPublisher = relayPubsub
			}
		}
	}

	// Create an error chan for exiting from goroutines
	errChan := make(chan error, 1)

	// Create a waitgroup for goroutines
	var wg sync.WaitGroup

	// Start the relay stats and ping stats goroutines
	analyticsPusher, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, relayStatsPublishInterval, pingStatsPublishInterval, httpTimeout, routeMatrixURI, routeMatrixStaleDuration, analyticsPusherMetrics)
	if err != nil {
		core.Error("could not create analytics pusher: %v", err)
		return 1
	}

	wg.Add(1)
	go analyticsPusher.Start(ctx, &wg, errChan)

	// Setup the status handler info
	statusData := &metrics.AnalyticsPusherStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				analyticsPusherMetrics.AnalyticsPusherServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				analyticsPusherMetrics.AnalyticsPusherServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.AnalyticsPusherStatus{}

				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = commitHash
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				newStatusData.Goroutines = int(analyticsPusherMetrics.AnalyticsPusherServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = analyticsPusherMetrics.AnalyticsPusherServiceMetrics.MemoryAllocated.Value()
				newStatusData.RouteMatrixInvocations = int(analyticsPusherMetrics.RouteMatrixInvocations.Value())
				newStatusData.RouteMatrixSuccesses = int(analyticsPusherMetrics.RouteMatrixSuccess.Value())
				newStatusData.RouteMatrixDuration = int(analyticsPusherMetrics.RouteMatrixUpdateDuration.Value())
				newStatusData.RouteMatrixLongDurations = int(analyticsPusherMetrics.RouteMatrixUpdateLongDuration.Value())
				newStatusData.PingStatsEntriesReceived = int(analyticsPusherMetrics.PingStatsMetrics.EntriesReceived.Value())
				newStatusData.PingStatsEntriesSubmitted = int(analyticsPusherMetrics.PingStatsMetrics.EntriesSubmitted.Value())
				newStatusData.PingStatsEntriesFlushed = int(analyticsPusherMetrics.PingStatsMetrics.EntriesFlushed.Value())
				newStatusData.RelayStatsEntriesReceived = int(analyticsPusherMetrics.RelayStatsMetrics.EntriesReceived.Value())
				newStatusData.RelayStatsEntriesSubmitted = int(analyticsPusherMetrics.RelayStatsMetrics.EntriesSubmitted.Value())
				newStatusData.RelayStatsEntriesFlushed = int(analyticsPusherMetrics.RelayStatsMetrics.EntriesFlushed.Value())
				newStatusData.PingStatusPublishFailures = int(analyticsPusherMetrics.PingStatsMetrics.ErrorMetrics.PublishFailure.Value())
				newStatusData.RelayStatsPublishFailures = int(analyticsPusherMetrics.RelayStatsMetrics.ErrorMetrics.PublishFailure.Value())
				newStatusData.RouteMatrixReaderNilErrors = int(analyticsPusherMetrics.ErrorMetrics.RouteMatrixReaderNil.Value())
				newStatusData.RouteMatrixReadErrors = int(analyticsPusherMetrics.ErrorMetrics.RouteMatrixReaderNil.Value())
				newStatusData.RouteMatrixBufferEmptyErrors = int(analyticsPusherMetrics.ErrorMetrics.RouteMatrixBufferEmpty.Value())
				newStatusData.RouteMatrixSerializeErrors = int(analyticsPusherMetrics.ErrorMetrics.RouteMatrixSerializeFailure.Value())
				newStatusData.RouteMatrixStaleErrors = int(analyticsPusherMetrics.ErrorMetrics.StaleRouteMatrix.Value())

				statusMutex.Lock()
				statusData = newStatusData
				statusMutex.Unlock()

				time.Sleep(time.Second * 10)
			}
		}()
	}

	serveStatusFunc := func(w http.ResponseWriter, r *http.Request) {
		statusMutex.RLock()
		data := statusData
		statusMutex.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			core.Error("could not write status data to json: %v\n%+v", err, data)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// Start HTTP Server
	{
		port := envvar.Get("PORT", "41002")
		if port == "" {
			core.Error("PORT not set")
			return 1
		}

		fmt.Printf("starting http server on port %s\n", port)

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.Handle("/debug/vars", expvar.Handler())

		enablePProf := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if enablePProf {
			router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
		}

		go func() {
			err := http.ListenAndServe(":"+port, router)
			if err != nil {
				core.Error("error starting http server: %v", err)
				errChan <- err
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		core.Debug("received shutdown signal")
		cancel()

		// Wait for essential goroutines to finish up
		wg.Wait()

		core.Debug("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		cancel()
		return 1
	}
}
