/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"runtime"

	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"

	"time"

	"github.com/gorilla/mux"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/transport"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
	"cloud.google.com/go/pubsub"
)

var (
	buildtime string
	sha       string
	tag       string
)

func main() {

	fmt.Printf("welcome to the nerd zone 1.0\n")

	ctx := context.Background()

	// Configure local logging
	logger := log.NewLogfmtLogger(os.Stdout)

	// Create a no-op metrics handler
	var metricsHandler metrics.Handler = &metrics.NoOpHandler{}

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

	var billingEntriesProcessed uint64

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpProjectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {

		// Google pubsub

		fmt.Printf("google project: %s\n", gcpProjectID)

		pubsubClient, err := pubsub.NewClient(ctx, gcpProjectID)
		if err != nil {
			fmt.Printf("could not create pubsub client\n")
			os.Exit(1)
		}

		pubsubTopic, err := pubsubClient.CreateTopic(ctx, "projects/network-next-v3-prod/topics/billing")
		if err != nil {
			fmt.Printf("could not create pubsub topic\n")
			os.Exit(1)
		}

		pubsubSubscription, err := pubsubClient.CreateSubscription(ctx, "billing", pubsub.SubscriptionConfig{
			Topic:            pubsubTopic,
			AckDeadline:      10 * time.Second,
			ExpirationPolicy: time.Duration(0),
		})
		if err != nil {
			fmt.Printf("could not create pubsub subscription\n")
			os.Exit(1)
		}
	
		err = pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
			// todo: process billing entry
			atomic.AddUint64(&billingEntriesProcessed, 1)
			m.Ack()
		})
		if err != context.Canceled {
			fmt.Printf("could not setup to receive pubsub messages\n")
			os.Exit(1)
		}

		// todo: goroutine or something to do work
		_ = pubsubSubscription

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

	// Setup the stats print routine
	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", runtime.NumGoroutine())
				fmt.Printf("%.2f mb allocated\n", memoryUsed())
				fmt.Printf("%d billing entries processed\n", billingEntriesProcessed)
				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second * 10)
			}
		}()
	}

	// Start HTTP server
	{
		go func() {
			router := mux.NewRouter()
			router.HandleFunc("/health", HealthHandlerFunc())
			router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag))
		
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
	}

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}

func HealthHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		statusCode := http.StatusOK

		w.WriteHeader(statusCode)
		w.Write([]byte(http.StatusText(statusCode)))
	}
}
