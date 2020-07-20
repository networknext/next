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

	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/transport"

	"cloud.google.com/go/bigquery"
	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {

	fmt.Printf("billing: Git Hash: %s - Commit: %s\n", sha, commitMessage)

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

	// Create a no-op biller
	var biller billing.Biller = &billing.NoOpBiller{}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	gcpProjectID, gcpOK := os.LookupEnv("GOOGLE_PROJECT_ID")
	if gcpOK {

		// StackDriver Metrics
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

		// StackDriver Profiler
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
					Service:        "billing",
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

	// Create billing metrics
	billingMetrics, err := metrics.NewBillingMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create billing metrics", "err", err)
	}

	if gcpOK {
		// Google BigQuery
		{
			if billingDataset, ok := os.LookupEnv("GOOGLE_BIGQUERY_DATASET_BILLING"); ok {
				batchSize := billing.DefaultBigQueryBatchSize
				if size, ok := os.LookupEnv("GOOGLE_BIGQUERY_BATCH_SIZE"); ok {
					s, err := strconv.ParseInt(size, 10, 64)
					if err != nil {
						level.Error(logger).Log("err", err)
						os.Exit(1)
					}
					batchSize = int(s)
				}

				bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}
				b := billing.GoogleBigQueryClient{
					Metrics:       billingMetrics,
					Logger:        logger,
					TableInserter: bqClient.Dataset(billingDataset).Table(os.Getenv("GOOGLE_BIGQUERY_TABLE_BILLING")).Inserter(),
					BatchSize:     batchSize,
				}

				// Set the Biller to BigQuery
				biller = &b

				// Start the background WriteLoop to batch write to BigQuery
				go func() {
					b.WriteLoop(ctx)
				}()
			}
		}
	}

	_, emulatorOK := os.LookupEnv("PUBSUB_EMULATOR_HOST")
	if emulatorOK { // Prefer to use the emulator instead of actual Google pubsub
		gcpProjectID = "local"

		// Use the local biller
		biller = &billing.LocalBiller{
			Logger: logger,
		}

		level.Info(logger).Log("msg", "Detected pubsub emulator")
	}

	if gcpOK || emulatorOK {
		// Google pubsub forwarder
		{
			topicName := "billing"
			subscriptionName := "billing"

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			pubsubForwarder, err := billing.NewPubSubForwarder(pubsubCtx, biller, logger, billingMetrics, gcpProjectID, topicName, subscriptionName)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			go pubsubForwarder.Forward(ctx)
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

				billingMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				billingMetrics.MemoryAllocated.Set(memoryUsed())

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", int(billingMetrics.Goroutines.Value()))
				fmt.Printf("%.2f mb allocated\n", billingMetrics.MemoryAllocated.Value())
				fmt.Printf("%d billing entries received\n", int(billingMetrics.EntriesReceived.Value()))
				fmt.Printf("%d billing entries submitted\n", int(billingMetrics.EntriesSubmitted.Value()))
				fmt.Printf("%d billing entries queued\n", int(billingMetrics.EntriesQueued.Value()))
				fmt.Printf("%d billing entries flushed\n", int(billingMetrics.EntriesFlushed.Value()))
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
			router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))

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
