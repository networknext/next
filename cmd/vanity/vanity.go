/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/backend"
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

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "vanity_metrics"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	ctx := context.Background()

	gcpProjectID := backend.GetGCPProjectID()

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	env, err := backend.GetEnv()
	if err != nil {
		level.Error(logger).Log("err", "ENV not set")
		return 1
	}

	gcpProjectID = backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""
	if gcpOK {
		level.Info(logger).Log("envvar", "GOOGLE_PROJECT_ID", "msg", "Found GOOGLE_PROJECT_ID")
	} else {
		level.Info(logger).Log("envvar", "GOOGLE_PROJECT_ID", "msg", "GOOGLE_PROJECT_ID not set. Vanity Metrics will be written to local metrics.")
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then

	// StackDriver Logging
	logger, err = backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// StackDriver Profiler
	if err = backend.InitStackDriverProfiler(gcpProjectID, "vanity_metrics", env); err != nil {
		level.Error(logger).Log("msg", "failed to initialize StackDriver profiler", "err", err)
		return 1
	}

	// Get the time series metrics handler for vanity metrics
	tsMetricsHandler, err := backend.GetTSMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Get standard metrics handler for observational usage
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Get metrics for evaluating the performance of vanity metrics
	vanityServiceMetrics, err := metrics.NewVanityServiceMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create vanity metric metrics", "err", err)
		return 1
	}

	// Get vanity metric subscribers
	var vanitySubscriber pubsub.Subscriber
	{
		vanityPort := envvar.Get("FEATURE_VANITY_METRIC_PORT", "6666")

		receiveBufferSize, err := envvar.GetInt("FEATURE_VANITY_METRIC_RECEIVE_BUFFER_SIZE", 1000000)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		vanityMetricSubscriber, err := pubsub.NewVanityMetricSubscriber(vanityPort, int(receiveBufferSize))
		if err != nil {
			level.Error(logger).Log("msg", "could not create vanity metric subscriber", "err", err)
			return 1
		}

		if err := vanityMetricSubscriber.Subscribe(pubsub.TopicVanityMetricData); err != nil {
			level.Error(logger).Log("msg", "could not subscribe to vanity metric data topic", "err", err)
			return 1
		}

		vanitySubscriber = vanityMetricSubscriber
	}

	// Get the message size for internal storage
	messageChanSize, err := envvar.GetInt("FEATURE_VANITY_METRIC_MESSAGE_CHANNEL_SIZE", 10000000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Get the redis host for the user session map
	redisUserSessions := envvar.Get("FEATURE_VANITY_METRIC_REDIS_HOST_USER_SESSIONS_MAP", "127.0.0.1:6379")

	// Get the redis password for the user session map
	redisPassword := envvar.Get("FEATURE_VANITY_METRIC_REDIS_PASSWORD_USER_SESSIONS_MAP", "")

	// Get the max idle time for a sessionID in redis
	vanityMaxUserIdleTime, err := envvar.GetDuration("FEATURE_VANITY_METRIC_MAX_USER_IDLE_TIME", time.Minute*5)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Get the name of the set for redis
	vanitySetName := envvar.Get("FEATURE_VANITY_METRIC_REDIS_SET_NAME", "RecentSessionIDs")

	// Get the max idle connections for redis
	redisMaxIdleConnections, err := envvar.GetInt("FEATURE_VANITY_METRIC_REDIS_MAX_IDLE_CONNECTIONS", 5)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Get the max active connections for redis
	redisMaxActiveConnections, err := envvar.GetInt("FEATURE_VANITY_METRIC_REDIS_MAX_ACTIVE_CONNECTIONS", 5)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Get the number of update goroutines for the vanity metric handler
	numVanityUpdateGoroutines, err := envvar.GetInt("FEATURE_VANITY_METRIC_GOROUTINE_COUNT", 5)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Setup the stats print routine
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

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", int(vanityServiceMetrics.Goroutines.Value()))
				fmt.Printf("%.2f mb allocated\n", vanityServiceMetrics.MemoryAllocated.Value())
				fmt.Printf("%d messages received\n", int(vanityServiceMetrics.ReceivedVanityCount.Value()))
				fmt.Printf("%d successful updates\n", int(vanityServiceMetrics.UpdateVanitySuccessCount.Value()))
				fmt.Printf("%d failed updates\n", int(vanityServiceMetrics.UpdateVanityFailureCount.Value()))
				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second * 10)
			}
		}()
	}

	// Get the vanity metric handler for writing to StackDriver
	vanityMetricHandler, err := vanity.NewVanityMetricHandler(tsMetricsHandler, vanityServiceMetrics, messageChanSize, vanitySubscriber, redisUserSessions, redisPassword, redisMaxIdleConnections, redisMaxActiveConnections, vanityMaxUserIdleTime, vanitySetName, env, logger)
	if err != nil {
		level.Error(logger).Log("msg", "error creating vanity metric handler", "err", err)
		return 1
	}

	// Start the goroutines for receiving vanity metrics from the backend and updating metrics
	errChan := make(chan error, 1)
	go func() {
		if err := vanityMetricHandler.Start(ctx, numVanityUpdateGoroutines); err != nil {
			level.Error(logger).Log("err", err)
			errChan <- err
			return
		}
	}()

	// Start HTTP server
	{
		go func() {
			router := mux.NewRouter()
			router.HandleFunc("/health", transport.HealthHandlerFunc())
			router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))

			enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
			if err != nil {
				level.Error(logger).Log("err", err)
			}
			if enablePProf {
				router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
			}

			port, ok := os.LookupEnv("HTTP_PORT")
			if !ok {
				level.Error(logger).Log("err", "env var HTTP_PORT must be set")
				errChan <- err
				return
			}

			err = http.ListenAndServe(":"+port, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				errChan <- err
				return
			}
		}()
	}

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-sigint:
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		return 1
	}
}
