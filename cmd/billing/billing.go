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
	"cloud.google.com/go/pubsub"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
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

	var billingEntriesReceived uint64

	// Create a no-op biller
	var biller billing.Biller = &billing.NoOpBiller{}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpProjectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {

		// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
		// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
		// on creation so we can use that for the default then
		// if gcpProjectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {

		/*
			// Create a Firestore Storer
			fs, err := storage.NewFirestore(ctx, gcpProjectID)//, logger)
			if err != nil {
				// level.Error(logger).Log("err", err)
				fmt.Printf("could not create firestore: %v\n", err)
				os.Exit(1)
			}

			fssyncinterval := os.Getenv("GOOGLE_FIRESTORE_SYNC_INTERVAL")
			syncInterval, err := time.ParseDuration(fssyncinterval)
			if err != nil {
				// level.Error(logger).Log("envvar", "GOOGLE_FIRESTORE_SYNC_INTERVAL", "value", fssyncinterval, "err", err)
				fmt.Printf("bad GOOGLE_FIRESTORE_SYNC_INTERVAL\n")
				os.Exit(1)
			}
			// Start a goroutine to sync from Firestore
			go func() {
				ticker := time.NewTicker(syncInterval)
				fs.SyncLoop(ctx, ticker.C)
			}()

			// Set the Firestore Storer to give to handlers
			db = fs
		*/

		// Google BigQuery

		if billingDataset, ok := os.LookupEnv("GOOGLE_BIGQUERY_DATASET_BILLING"); ok {
			batchSize := billing.DefaultBigQueryBatchSize
			if size, ok := os.LookupEnv("GOOGLE_BIGQUERY_BATCH_SIZE"); ok {
				s, err := strconv.ParseInt(size, 10, 64)
				if err != nil {
					// level.Error(logger).Log("err", err)
					fmt.Println(err)
					os.Exit(1)
				}
				batchSize = int(s)
			}

			bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
			if err != nil {
				// level.Error(logger).Log("err", err)
				fmt.Println(err)
				os.Exit(1)
			}
			b := billing.GoogleBigQueryClient{
				// Logger:        logger,
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

		// Google pubsub

		fmt.Printf("google project: %s\n", gcpProjectID)

		pubsubClient, err := pubsub.NewClient(ctx, gcpProjectID)
		if err != nil {
			fmt.Printf("could not create pubsub client\n")
			os.Exit(1)
		}

		subscriptionName := "billing"

		fmt.Printf("subscription name: %s\n", subscriptionName)

		pubsubSubscription := pubsubClient.Subscription(subscriptionName)

		go func() {
			err = pubsubSubscription.Receive(ctx, func(ctx context.Context, m *pubsub.Message) {
				atomic.AddUint64(&billingEntriesReceived, 1)
				billingEntry := billing.BillingEntry{}
				if billing.ReadBillingEntry(&billingEntry, m.Data) {
					m.Ack()
					billingEntry.Timestamp = uint64(m.PublishTime.Unix())
					if err := biller.Bill(context.Background(), &billingEntry); err != nil {
						fmt.Printf("could not submit billing entry: %v\n", err)
						// level.Error(params.Logger).Log("msg", "could not submit billing entry", "err", err)
						// params.Metrics.ErrorMetrics.BillingFailure.Add(1)
					}
				} else {
					// todo: metric for read failures
				}
			})
			if err != context.Canceled {
				fmt.Printf("could not setup to receive pubsub messages\n")
				os.Exit(1)
			}
		}()

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
				fmt.Printf("%d billing entries received\n", billingEntriesReceived)
				fmt.Printf("%d billing entries submitted\n", biller.NumSubmitted())
				fmt.Printf("%d billing entries queued\n", biller.NumQueued())
				fmt.Printf("%d billing entries flushed\n", biller.NumFlushed())
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
