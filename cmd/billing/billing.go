/*
	Network Next. You control the network.
	Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"expvar"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/config"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/logging"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"

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

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Configure local logging
	logger := log.NewLogfmtLogger(os.Stdout)

	// Create a no-op metrics handler
	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

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

				logger = logging.NewStackdriverLogger(loggingClient, "billing")
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

	// Get billing feature config
	var featureConfig config.Config
	envVarConfig := config.NewEnvVarConfig([]config.Feature{
		{
			Name:        "FEATURE_BILLING",
			Enum:        config.FEATURE_BILLING,
			Value:       false,
			Description: "Receives BillingEntry types from Google Pub/Sub and writes them to BigQuery",
		},
		{
			Name:        "FEATURE_BILLING2",
			Enum:        config.FEATURE_BILLING2,
			Value:       true,
			Description: "Receives BillingEntry2 types from Google Pub/Sub and writes them to BigQuery",
		},
	})
	featureConfig = envVarConfig
	featureBilling := featureConfig.FeatureEnabled(config.FEATURE_BILLING)
	featureBilling2 := featureConfig.FeatureEnabled(config.FEATURE_BILLING2)

	// Create no-op billers
	var biller billing.Biller = &billing.NoOpBiller{}
	var biller2 billing.Biller = &billing.NoOpBiller{}

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
	billingServiceMetrics, err := metrics.NewBillingServiceMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create billing service metrics", "err", err)
	}

	if gcpOK {
		// Google BigQuery
		if featureBilling {
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

				// Pass context without cancel to ensure writing continues even past reception of shutdown signal
				bqClient, err := bigquery.NewClient(context.Background(), gcpProjectID)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}
				b := billing.GoogleBigQueryClient{
					Metrics:        &billingServiceMetrics.BillingMetrics,
					Logger:         logger,
					TableInserter:  bqClient.Dataset(billingDataset).Table(os.Getenv("GOOGLE_BIGQUERY_TABLE_BILLING")).Inserter(),
					BatchSize:      batchSize,
					FeatureBilling: featureBilling,
				}

				// Set the Biller to BigQuery
				biller = &b

				// Start the background WriteLoop to batch write to BigQuery
				wg.Add(1)
				go func() {
					b.WriteLoop(ctx, wg)
				}()
			}
		}

		if featureBilling2 {
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

				summaryBatchSize := billing.DefaultBigQueryBatchSize / 10
				if size, ok := os.LookupEnv("GOOGLE_BIGQUERY_SUMMARY_BATCH_SIZE"); ok {
					s, err := strconv.ParseInt(size, 10, 64)
					if err != nil {
						level.Error(logger).Log("err", err)
						os.Exit(1)
					}
					summaryBatchSize = int(s)
				}

				batchSizePercent, err := envvar.GetFloat("FEATURE_BILLING2_BATCH_SIZE_PERCENT", 0.80)
				if err != nil {
					level.Error(logger).Log("envvar", "FEATURE_BILLING2_BATCH_SIZE_PERCENT", "msg", "failed to parse envvar", "err", err)
					os.Exit(1)
				}

				billing2TableName := envvar.Get("FEATURE_BILLING2_GOOGLE_BIGQUERY_TABLE_BILLING", "billing2")

				billing2SummaryTableName := envvar.Get("FEATURE_BILLING2_GOOGLE_BIGQUERY_TABLE_BILLING_SUMMARY", "billing2_session_summary")

				// Pass context without cancel to ensure writing continues even past reception of shutdown signal
				bqClient, err := bigquery.NewClient(context.Background(), gcpProjectID)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}
				b := billing.GoogleBigQueryClient{
					Metrics:              &billingServiceMetrics.BillingMetrics,
					Logger:               logger,
					TableInserter:        bqClient.Dataset(billingDataset).Table(billing2TableName).Inserter(),
					SummaryTableInserter: bqClient.Dataset(billingDataset).Table(billing2SummaryTableName).Inserter(),
					BatchSize:            batchSize,
					SummaryBatchSize:     summaryBatchSize,
					BatchSizePercent:     batchSizePercent,
					FeatureBilling2:      featureBilling2,
				}

				// Set the Biller to BigQuery
				biller2 = &b

				// Start the background WriteLoop to batch write to BigQuery
				wg.Add(1)
				go func() {
					b.WriteLoop2(ctx, wg)
				}()

				wg.Add(1)
				go func() {
					b.SummaryWriteLoop2(ctx, wg)
				}()

			}
		}
	}

	_, emulatorOK := os.LookupEnv("PUBSUB_EMULATOR_HOST")
	if emulatorOK { // Prefer to use the emulator instead of actual Google pubsub
		gcpProjectID = "local"

		// Use the local biller
		biller = &billing.LocalBiller{
			Logger:  logger,
			Metrics: &billingServiceMetrics.BillingMetrics,
		}

		biller2 = &billing.LocalBiller{
			Logger:  logger,
			Metrics: &billingServiceMetrics.BillingMetrics,
		}

		level.Info(logger).Log("msg", "Detected pubsub emulator")
	}

	if gcpOK || emulatorOK {
		// Google pubsub forwarder
		{
			numRecvGoroutines, err := envvar.GetInt("NUM_RECEIVE_GOROUTINES", 10)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			if featureBilling {

				level.Debug(logger).Log("msg", "Billing enabled")

				topicName := envvar.Get("BILLING_TOPIC_NAME", "billing")
				subscriptionName := envvar.Get("BILLING_SUBSCRIPTION_NAME", "billing")

				pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
				defer cancelFunc()

				pubsubForwarder, err := billing.NewPubSubForwarder(pubsubCtx, biller, 25, time.Second, logger, &billingServiceMetrics.BillingMetrics, gcpProjectID, topicName, subscriptionName, numRecvGoroutines)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}

				wg.Add(1)
				go pubsubForwarder.Forward(ctx, wg)
			}

			if featureBilling2 {

				level.Debug(logger).Log("msg", "Billing2 enabled")

				maxRetries, err := envvar.GetInt("FEATURE_BILLING2_MAX_RETRIES", 25)
				if err != nil {
					level.Error(logger).Log("envvar", "FEATURE_BILLING2_MAX_RETRIES", "msg", "failed to parse envvar", "err", err)
					os.Exit(1)
				}

				retryTime, err := envvar.GetDuration("FEATURE_BILLING2_RETRY_TIME", time.Second*1)
				if err != nil {
					level.Error(logger).Log("envvar", "FEATURE_BILLING2_RETRY_TIME", "msg", "failed to parse envvar", "err", err)
					os.Exit(1)
				}

				topicName := envvar.Get("FEATURE_BILLING2_TOPIC_NAME", "billing2")
				subscriptionName := envvar.Get("FEATURE_BILLING2_SUBSCRIPTION_NAME", "billing2")

				pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
				defer cancelFunc()

				pubsubForwarder, err := billing.NewPubSubForwarder(pubsubCtx, biller2, maxRetries, retryTime, logger, &billingServiceMetrics.BillingMetrics, gcpProjectID, topicName, subscriptionName, numRecvGoroutines)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}

				wg.Add(1)
				go pubsubForwarder.Forward2(ctx, wg)
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

				billingServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				billingServiceMetrics.MemoryAllocated.Set(memoryUsed())

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", int(billingServiceMetrics.Goroutines.Value()))
				fmt.Printf("%.2f mb allocated\n", billingServiceMetrics.MemoryAllocated.Value())
				fmt.Printf("%d billing entries received\n", int(billingServiceMetrics.BillingMetrics.EntriesReceived.Value()))
				fmt.Printf("%d billing entries submitted\n", int(billingServiceMetrics.BillingMetrics.EntriesSubmitted.Value()))
				fmt.Printf("%d billing entries queued\n", int(billingServiceMetrics.BillingMetrics.EntriesQueued.Value()))
				fmt.Printf("%d billing entries flushed\n", int(billingServiceMetrics.BillingMetrics.EntriesFlushed.Value()))
				fmt.Printf("%d billing entries with NaN\n", int(billingServiceMetrics.BillingMetrics.ErrorMetrics.BillingEntriesWithNaN.Value()))
				fmt.Printf("%d invalid billing entries\n", int(billingServiceMetrics.BillingMetrics.ErrorMetrics.BillingInvalidEntries.Value()))
				fmt.Printf("%d billing entry write failures\n", int(billingServiceMetrics.BillingMetrics.ErrorMetrics.BillingWriteFailure.Value()))
				fmt.Printf("%d billing entry 2s received\n", int(billingServiceMetrics.BillingMetrics.Entries2Received.Value()))
				fmt.Printf("%d billing entry 2s submitted\n", int(billingServiceMetrics.BillingMetrics.Entries2Submitted.Value()))
				fmt.Printf("%d billing entry 2s queued\n", int(billingServiceMetrics.BillingMetrics.Entries2Queued.Value()))
				fmt.Printf("%d billing entry 2s flushed\n", int(billingServiceMetrics.BillingMetrics.Entries2Flushed.Value()))
				fmt.Printf("%d billing summary entry 2s submitted\n", int(billingServiceMetrics.BillingMetrics.SummaryEntries2Submitted.Value()))
				fmt.Printf("%d billing sumary entry 2s queued\n", int(billingServiceMetrics.BillingMetrics.SummaryEntries2Queued.Value()))
				fmt.Printf("%d billing summary entry 2s flushed\n", int(billingServiceMetrics.BillingMetrics.SummaryEntries2Flushed.Value()))
				fmt.Printf("%d billing entry 2s with NaN\n", int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2EntriesWithNaN.Value()))
				fmt.Printf("%d invalid billing entry 2s\n", int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2InvalidEntries.Value()))
				fmt.Printf("%d billing entry 2 read failures\n", int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2ReadFailure.Value()))
				fmt.Printf("%d billing entry 2 write failures\n", int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2WriteFailure.Value()))
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
			router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
			router.Handle("/debug/vars", expvar.Handler())

			enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
			if err != nil {
				level.Error(logger).Log("err", err)
			}
			if enablePProf {
				router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
			}

			port, ok := os.LookupEnv("PORT")
			if !ok {
				level.Error(logger).Log("err", "env var PORT must be set")
				os.Exit(1)
			}

			level.Info(logger).Log("addr", ":"+port)

			err = http.ListenAndServe(":"+port, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan
	level.Debug(logger).Log("msg", "Received shutdown signal")
	fmt.Println("Received shutdown signal.")

	cancel()
	// Wait for essential goroutines to finish up
	wg.Wait()

	level.Debug(logger).Log("msg", "Successfully shutdown.")
	fmt.Println("Successfully shutdown.")
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
