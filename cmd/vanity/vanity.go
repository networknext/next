/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/pubsub"
	"github.com/networknext/backend/modules/vanity"
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

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func mainReturnWithCode() int {
	serviceName := "vanity_metrics"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, ctxCancelFunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	logger := log.NewNopLogger()

	env, err := backend.GetEnv()
	if err != nil {
		core.Error("error getting env: %v", err)
		return 1
	}

	gcpProjectID := backend.GetGCPProjectID()
	if gcpProjectID != "" {
		core.Debug("Found GOOGLE_PROJECT_ID: %s", gcpProjectID)
	} else {
		core.Debug("GOOGLE_PROJECT_ID not set. Vanity Metrics will be written to local metrics")
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then

	// StackDriver Profiler
	if err = backend.InitStackDriverProfiler(gcpProjectID, "vanity_metrics", env); err != nil {
		core.Error("failed to initialize StackDriver profiler: %v", err)
		return 1
	}

	// Get the time series metrics handler for vanity metrics
	tsMetricsHandler, err := backend.GetTSMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("could not create time series metrics handler: %v", err)
		return 1
	}

	// Get standard metrics handler for observational usage
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("could not create metrics handler: %v", err)
		return 1
	}

	// Get metrics for evaluating the performance of vanity metrics
	vanityServiceMetrics, err := metrics.NewVanityServiceMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to create vanity service metrics: %v", err)
		return 1
	}

	// Get vanity metric subscribers
	var vanitySubscriber pubsub.Subscriber
	{
		vanityPort := envvar.Get("FEATURE_VANITY_METRIC_PORT", "6666")

		receiveBufferSize, err := envvar.GetInt("FEATURE_VANITY_METRIC_RECEIVE_BUFFER_SIZE", 1000000)
		if err != nil {
			core.Error("could not parse FEATURE_VANITY_METRIC_RECEIVE_BUFFER_SIZE: %v", err)
			return 1
		}

		vanityMetricSubscriber, err := pubsub.NewVanityMetricSubscriber(vanityPort, int(receiveBufferSize))
		if err != nil {
			core.Error("could not create vanity metric subscriber: %v", err)
			return 1
		}

		if err := vanityMetricSubscriber.Subscribe(pubsub.TopicVanityMetricData); err != nil {
			core.Error("could not subscribe to vanity metric data topic: %v", err)
			return 1
		}

		vanitySubscriber = vanityMetricSubscriber
	}

	// Get the message size for internal storage
	messageChanSize, err := envvar.GetInt("FEATURE_VANITY_METRIC_MESSAGE_CHANNEL_SIZE", 10000000)
	if err != nil {
		core.Error("could not parse FEATURE_VANITY_METRIC_MESSAGE_CHANNEL_SIZEL: %v", err)
		return 1
	}

	// Get the redis host for the user session map
	redisUserSessions := envvar.Get("FEATURE_VANITY_METRIC_REDIS_HOST_USER_SESSIONS_MAP", "127.0.0.1:6379")

	// Get the redis password for the user session map
	redisPassword := envvar.Get("FEATURE_VANITY_METRIC_REDIS_PASSWORD_USER_SESSIONS_MAP", "")

	// Get the max idle time for a sessionID in redis
	vanityMaxUserIdleTime, err := envvar.GetDuration("FEATURE_VANITY_METRIC_MAX_USER_IDLE_TIME", time.Minute*5)
	if err != nil {
		core.Error("could not parse FEATURE_VANITY_METRIC_MAX_USER_IDLE_TIME: %v", err)
		return 1
	}

	// Get the name of the set for redis
	vanitySetName := envvar.Get("FEATURE_VANITY_METRIC_REDIS_SET_NAME", "RecentSessionIDs")

	// Get the max idle connections for redis
	redisMaxIdleConnections, err := envvar.GetInt("FEATURE_VANITY_METRIC_REDIS_MAX_IDLE_CONNECTIONS", 5)
	if err != nil {
		core.Error("could not parse FEATURE_VANITY_METRIC_REDIS_MAX_IDLE_CONNECTIONS: %v", err)
		return 1
	}

	// Get the max active connections for redis
	redisMaxActiveConnections, err := envvar.GetInt("FEATURE_VANITY_METRIC_REDIS_MAX_ACTIVE_CONNECTIONS", 5)
	if err != nil {
		core.Error("could not parse FEATURE_VANITY_METRIC_REDIS_MAX_ACTIVE_CONNECTIONS: %v", err)
		return 1
	}

	// Get the number of update goroutines for the vanity metric handler
	numVanityUpdateGoroutines, err := envvar.GetInt("FEATURE_VANITY_METRIC_GOROUTINE_COUNT", 5)
	if err != nil {
		core.Error("could not parse FEATURE_VANITY_METRIC_GOROUTINE_COUNT: %v", err)
		return 1
	}

	// Get the vanity metric handler for writing to StackDriver
	vanityMetricHandler, err := vanity.NewVanityMetricHandler(
		tsMetricsHandler,
		vanityServiceMetrics,
		messageChanSize,
		vanitySubscriber,
		redisUserSessions,
		redisPassword,
		redisMaxIdleConnections,
		redisMaxActiveConnections,
		vanityMaxUserIdleTime,
		vanitySetName,
		env,
	)
	if err != nil {
		core.Error("failed to create vanity metric handler: %v", err)
		return 1
	}

	// Start the goroutines for receiving vanity metrics from the backend and updating metrics
	errChan := make(chan error, 1)
	go vanityMetricHandler.Start(ctx, numVanityUpdateGoroutines, wg, errChan)

	// Setup the status handler info
	type VanityStatus struct {
		// Service Information
		ServiceName string `json:"service_name"`
		GitHash     string `json:"git_hash"`
		Started     string `json:"started"`
		Uptime      string `json:"uptime"`

		// Metrics
		Goroutines        int     `json:"goroutines"`
		MemoryAllocated   float64 `json:"mb_allocated"`
		MessagesReceived  int     `json:"messages_received"`
		SuccessfulUpdates int     `json:"successful_updates"`
		FailedUpdates     int     `json:"failed_updates"`
	}

	statusData := &VanityStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				vanityServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				vanityServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &VanityStatus{}

				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()
				newStatusData.Goroutines = int(vanityServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = vanityServiceMetrics.MemoryAllocated.Value()
				newStatusData.MessagesReceived = int(vanityServiceMetrics.ReceivedVanityCount.Value())
				newStatusData.SuccessfulUpdates = int(vanityServiceMetrics.UpdateVanitySuccessCount.Value())
				newStatusData.FailedUpdates = int(vanityServiceMetrics.UpdateVanityFailureCount.Value())

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

	// Start HTTP server
	{
		port := envvar.Get("HTTP_PORT", "41005")
		if port == "" {
			core.Error("HTTP_PORT not set")
			return 1
		}

		fmt.Printf("starting http server on port %s\n", port)

		go func() {
			router := mux.NewRouter()
			router.HandleFunc("/health", transport.HealthHandlerFunc())
			router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
			router.HandleFunc("/status", serveStatusFunc).Methods("GET")

			enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
			if err != nil {
				core.Error("could not parse envvar FEATURE_ENABLE_PPROF: %v", err)
			}
			if enablePProf {
				router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
			}

			err = http.ListenAndServe(":"+port, router)
			if err != nil {
				core.Error("error starting http server: %v", err)
				errChan <- err
				return
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		core.Debug("received shutdown signal")
		ctxCancelFunc()

		// Wait for essential goroutines to finish up
		wg.Wait()

		core.Debug("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		ctxCancelFunc()
		return 1
	}
}
