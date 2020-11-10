package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"time"

	"cloud.google.com/go/bigquery"
	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/logging"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/transport"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {
	fmt.Printf("analytics: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	logger := log.NewLogfmtLogger(os.Stdout)

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

				logger = logging.NewStackdriverLogger(loggingClient, "analytics")
			}
		}
	}

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

	env, ok := os.LookupEnv("ENV")
	if !ok {
		level.Error(logger).Log("err", "ENV not set")
		os.Exit(1)
	}

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
				go metricsHandler.WriteLoop(ctx, logger, writeInterval, 200)
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
					Service:        "analytics",
					ServiceVersion: env,
					ProjectID:      gcpProjectID,
					MutexProfiling: true,
				}); err != nil {
					level.Error(logger).Log("msg", "failed to initialize StackDriver profiler", "err", err)
					os.Exit(1)
				}
			}
		}
	}

	analyticsMetrics, err := metrics.NewAnalyticsServiceMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create analytics metrics", "err", err)
	}

	var pingStatsWriter analytics.PingStatsWriter = &analytics.NoOpPingStatsWriter{}
	{
		// BigQuery
		if gcpOK {
			if analyticsDataset, ok := os.LookupEnv("GOOGLE_BIGQUERY_DATASET_PING_STATS"); ok {
				bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}
				b := analytics.NewGoogleBigQueryPingStatsWriter(bqClient, logger, &analyticsMetrics.PingStatsMetrics, analyticsDataset, os.Getenv("GOOGLE_BIGQUERY_TABLE_PING_STATS"))
				pingStatsWriter = &b

				go b.WriteLoop(ctx)
			}
		}

		_, emulatorOK := os.LookupEnv("PUBSUB_EMULATOR_HOST")
		if emulatorOK {
			gcpProjectID = "local"

			pingStatsWriter = &analytics.LocalPingStatsWriter{
				Logger: logger,
			}

			level.Info(logger).Log("msg", "detected pubsub emulator")
		}

		// google pubsub forwarder
		if gcpOK || emulatorOK {
			topicName := "ping_stats"
			subscriptionName := "ping_stats"

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(60*time.Minute))
			defer cancelFunc()

			pubsubForwarder, err := analytics.NewPingStatsPubSubForwarder(pubsubCtx, pingStatsWriter, logger, &analyticsMetrics.PingStatsMetrics, gcpProjectID, topicName, subscriptionName)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			go pubsubForwarder.Forward(ctx)
		}
	}

	var relayStatsWriter analytics.RelayStatsWriter = &analytics.NoOpRelayStatsWriter{}
	{
		// BigQuery
		if gcpOK {
			if analyticsDataset, ok := os.LookupEnv("GOOGLE_BIGQUERY_DATASET_RELAY_STATS"); ok {
				bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}
				b := analytics.NewGoogleBigQueryRelayStatsWriter(bqClient, logger, &analyticsMetrics.RelayStatsMetrics, analyticsDataset, os.Getenv("GOOGLE_BIGQUERY_TABLE_RELAY_STATS"))
				relayStatsWriter = &b

				go b.WriteLoop(ctx)
			}
		}

		_, emulatorOK := os.LookupEnv("PUBSUB_EMULATOR_HOST")
		if emulatorOK {
			gcpProjectID = "local"

			relayStatsWriter = &analytics.LocalRelayStatsWriter{
				Logger: logger,
			}

			level.Info(logger).Log("msg", "detected pubsub emulator")
		}

		// google pubsub forwarder
		if gcpOK || emulatorOK {
			topicName := "relay_stats"
			subscriptionName := "relay_stats"

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(60*time.Minute))
			defer cancelFunc()

			pubsubForwarder, err := analytics.NewRelayStatsPubSubForwarder(pubsubCtx, relayStatsWriter, logger, &analyticsMetrics.RelayStatsMetrics, gcpProjectID, topicName, subscriptionName)
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
				analyticsMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				analyticsMetrics.MemoryAllocated.Set(memoryUsed())

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", int(analyticsMetrics.Goroutines.Value()))
				fmt.Printf("%.2f mb allocated\n", analyticsMetrics.MemoryAllocated.Value())
				fmt.Printf("%d ping stats entries received\n", int(analyticsMetrics.PingStatsMetrics.EntriesReceived.Value()))
				fmt.Printf("%d ping stats entries submitted\n", int(analyticsMetrics.PingStatsMetrics.EntriesSubmitted.Value()))
				fmt.Printf("%d ping stats entries queued\n", int(analyticsMetrics.PingStatsMetrics.EntriesQueued.Value()))
				fmt.Printf("%d ping stats entries flushed\n", int(analyticsMetrics.PingStatsMetrics.EntriesFlushed.Value()))
				fmt.Printf("%d relay stats entries received\n", int(analyticsMetrics.RelayStatsMetrics.EntriesReceived.Value()))
				fmt.Printf("%d relay stats entries submitted\n", int(analyticsMetrics.RelayStatsMetrics.EntriesSubmitted.Value()))
				fmt.Printf("%d relay stats entries queued\n", int(analyticsMetrics.RelayStatsMetrics.EntriesQueued.Value()))
				fmt.Printf("%d relay stats entries flushed\n", int(analyticsMetrics.RelayStatsMetrics.EntriesFlushed.Value()))
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
