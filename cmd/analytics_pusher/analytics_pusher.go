package main

import (
	"bytes"
	"context"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/networknext/backend/modules/analytics"
	pusher "github.com/networknext/backend/modules/analytics_pusher"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport/middleware"

	frontend "github.com/networknext/backend/modules/relay_frontend"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "analytics_pusher"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

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

	env, err := backend.GetEnv()
	if err != nil {
		core.Error("error getting env: %v", err)
		return 1
	}

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("error getting metrics handler: %v", err)
		return 1
	}

	// Create analytics pusher metrics
	analyticsPusherMetrics, err := metrics.NewAnalyticsPusherMetrics(ctx, metricsHandler)
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
	pingStatsPublishInterval, err := envvar.GetDuration("PING_STATS_PUBLISH_INTERVAL", 1*time.Minute)
	if err != nil {
		core.Error("error getting PING_STATS_PUBLISH_INTERVAL: %v", err)
		return 1
	}

	relayStatsPublishInterval, err := envvar.GetDuration("RELAY_STATS_PUBLISH_INTERVAL", 10*time.Second)
	if err != nil {
		core.Error("error getting RELAY_STATS_PUBLISH_INTERVAL: %v", err)
		return 1
	}

	// Get HTTP timeout for route matrix
	httpTimeout, err := envvar.GetDuration("HTTP_TIMEOUT", 4*time.Second)
	if err != nil {
		core.Error("error getting HTTP_TIMEOUT: %v", err)
		return 1
	}

	// Get route matrix URI
	routeMatrixURI := envvar.Get("ROUTE_MATRIX_URI", "")
	if routeMatrixURI == "" {
		core.Error("ROUTE_MATRIX_URI not set")
		return 1
	}

	// Setup ping stats and relay stats publishers
	var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
	var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
	{
		emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
		if gcpOK || emulatorOK {

			pubsubCtx := context.Background()
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

				pingPubsub, err := analytics.NewGooglePubSubPingStatsPublisher(pubsubCtx, &analyticsPusherMetrics.PingStatsMetrics, logger, gcpProjectID, "ping_stats", settings)
				if err != nil {
					core.Error("could not create ping stats analytics pubsub publisher: %v", err)
					return 1
				}

				pingStatsPublisher = pingPubsub

				relayPubsub, err := analytics.NewGooglePubSubRelayStatsPublisher(pubsubCtx, &analyticsPusherMetrics.RelayStatsMetrics, logger, gcpProjectID, "relay_stats", settings)
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
	analyticsPusher, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, relayStatsPublishInterval, pingStatsPublishInterval, httpTimeout, routeMatrixURI, analyticsPusherMetrics)
	if err != nil {
		core.Error("could not create analytics pusher: %v", err)
		return 1
	}

	analyticsPusher.Start(ctx, errChan, &wg)

	// Setup the status handler info
	var statusData []byte
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				analyticsPusherMetrics.ServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				analyticsPusherMetrics.ServiceMetrics.MemoryAllocated.Set(memoryUsed())

				statusDataString := fmt.Sprintf("%s\n", serviceName)
				statusDataString += fmt.Sprintf("git hash %s\n", sha)
				statusDataString += fmt.Sprintf("started %s\n", startTime.Format("Mon, 02 Jan 2006 15:04:05 EST"))
				statusDataString += fmt.Sprintf("uptime %s\n", time.Since(startTime))
				statusDataString += fmt.Sprintf("%d goroutines\n", int(analyticsPusherMetrics.FrontendServiceMetrics.Goroutines.Value()))
				statusDataString += fmt.Sprintf("%.2f mb allocated\n", analyticsPusherMetrics.FrontendServiceMetrics.MemoryAllocated.Value())
				statusDataString += fmt.Sprintf("%d route matrix invocations\n", int(analyticsPusherMetrics.RouteMatrix.Invocations.Value()))
				statusDataString += fmt.Sprintf("%d route matrix success count\n", int(analyticsPusherMetrics.RouteMatrix.Success.Value()))
				statusDataString += fmt.Sprintf("%d route matrix errors\n", int(analyticsPusherMetrics.RouteMatrix.Error.Value()))
				statusDataString += fmt.Sprintf("%d ping stats entries submitted\n", int(analyticsPusherMetrics.PingStatsMetrics.EntriesSubmitted.Value()))
				statusDataString += fmt.Sprintf("%d ping stats entries queued\n", int(analyticsPusherMetrics.PingStatsMetrics.EntriesQueued.Value()))
				statusDataString += fmt.Sprintf("%d ping stats entries flushed\n", int(analyticsPusherMetrics.PingStatsMetrics.EntriesFlushed.Value()))
				statusDataString += fmt.Sprintf("%d relay stats entries submitted\n", int(analyticsPusherMetrics.RelayStatsMetrics.EntriesSubmitted.Value()))
				statusDataString += fmt.Sprintf("%d relay stats entries queued\n", int(analyticsPusherMetrics.RelayStatsMetrics.EntriesQueued.Value()))
				statusDataString += fmt.Sprintf("%d relay stats entries flushed\n", int(analyticsPusherMetrics.RelayStatsMetrics.EntriesFlushed.Value()))

				statusMutex.Lock()
				statusData = []byte(statusDataString)
				statusMutex.Unlock()

				time.Sleep(time.Second * 10)
			}
		}()
	}

	serveStatusFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		statusMutex.RLock()
		data := statusData
		statusMutex.RUnlock()
		buffer := bytes.NewBuffer(data)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// Start HTTP Server
	{
		port := envvar.Get("PORT", "30005")
		if port == "" {
			core.Error("PORT not set")
			return 1
		}

		fmt.Printf("starting http server on port %s\n", port)

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.HandleFunc("/route_matrix", frontendClient.GetRouteMatrixHandlerFunc()).Methods("GET")
		router.Handle("/debug/vars", expvar.Handler())

		enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if err != nil {
			level.Error(logger).Log("err", err)
		}
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
