package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
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
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/logging"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
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

func mainReturnWithCode() {
	serviceName := "analytics"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, ctxCancelFunc := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	logger := log.NewNopLogger()

	gcpProjectID := backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("could not get metrics handler: %v", err)
		return 1
	}

	env, err := backend.GetEnv()
	if err != nil {
		core.Error("could not get env: %v", err)
		return 1
	}

	if gcpOK {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("could not initialize stackdriver profiler: %v", err)
			return 1
		}
	}

	analyticsMetrics, err := metrics.NewAnalyticsServiceMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to create analytics metrics: %v", err)
		return 1
	}

	if gcpOK {
		// Google BigQuery
		{
			pingStatsDataset := envvar.Get("GOOGLE_BIGQUERY_DATASET_PING_STATS", "")
			if pingStatsDataset == "" {
				core.Error("envvar GOOGLE_BIGQUERY_DATASET_PING_STATS not set")
				return 1
			}

			pingStatsTableName := envvar.Get("GOOGLE_BIGQUERY_TABLE_PING_STATS", "")
			if pingStatsTableName == "" {
				core.Error("envvar GOOGLE_BIGQUERY_TABLE_PING_STATS not set")
				return 1
			}

			bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
			if err != nil {
				core.Error("could not create ping stats bigquery client: %v", err)
				return 1
			}

			b := analytics.NewGoogleBigQueryPingStatsWriter(bqClient, logger, &analyticsMetrics.PingStatsMetrics, pingStatsDataset, pingStatsTableName)
			pingStatsWriter = &b

			go b.WriteLoop(wg)
		}

		{
			relayStatsDataset := envvar.Get("GOOGLE_BIGQUERY_DATASET_RELAY_STATS", "")
			if relayStatsDataset == "" {
				core.Error("envvar GOOGLE_BIGQUERY_DATASET_RELAY_STATS not set")
				return 1
			}

			relayStatsTableName := envvar.Get("GOOGLE_BIGQUERY_TABLE_RELAY_STATS", "")
			if relayStatsTableName == "" {
				core.Error("envvar GOOGLE_BIGQUERY_TABLE_RELAY_STATS not set")
				return 1
			}

			bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
			if err != nil {
				core.Error("could not create relay stats bigquery client: %v", err)
				return 1
			}

			b := analytics.NewGoogleBigQueryRelayStatsWriter(bqClient, logger, &analyticsMetrics.RelayStatsMetrics, relayStatsDataset, relayStatsTableName)
			relayStatsWriter = &b

			go b.WriteLoop(wg)
		}
	}

	var pingStatsWriter analytics.PingStatsWriter = &analytics.NoOpPingStatsWriter{}
	var relayStatsWriter analytics.RelayStatsWriter = &analytics.NoOpRelayStatsWriter{}

	pubsubEmulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")

	if gcpOK || pubsubEmulatorOK {

		if pubsubEmulatorOK {
			// Prefer to use the emulator instead of actual Google pubsub
			gcpProjectID = "local"

			// Use local ping stats and relay stats writer
			pingStatsWriter = analytics.LocalPingStatsWriter{Logger: logger}
			relayStatsWriter = analytics.LocalRelayStatsWriter{Logger: logger}
		}

		// Google pubsub forwarder
		{
			topicName := envvar.Get("PING_STATS_TOPIC_NAME", "ping_stats")
			subscriptionName := envvar.Get("PING_STATS_SUBSCRIPTION_NAME", "ping_stats")

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			pubsubForwarder, err := analytics.NewPingStatsPubSubForwarder(pubsubCtx, pingStatsWriter, logger, &analyticsMetrics.PingStatsMetrics, gcpProjectID, topicName, subscriptionName)
			if err != nil {
				core.Error("could not create ping stats pub sub forwarder: %v", err)
				return 1
			}

			wg.Add(1)
			go pubsubForwarder.Forward(ctx, wg)
		}

		{
			topicName := envvar.Get("RELAY_STATS_TOPIC_NAME", "ping_stats")
			subscriptionName := envvar.Get("RELAY_STATS_SUBSCRIPTION_NAME", "ping_stats")

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			pubsubForwarder, err := analytics.NewRelayStatsPubSubForwarder(pubsubCtx, relayStatsWriter, logger, &analyticsMetrics.RelayStatsMetrics, gcpProjectID, topicName, subscriptionName)
			if err != nil {
				core.Error("could not create relay stats pub sub forwarder: %v", err)
				return 1
			}

			wg.Add(1)
			go pubsubForwarder.Forward(ctx, wg)
		}
	}

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
				analyticsMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				analyticsMetrics.MemoryAllocated.Set(memoryUsed())

				statusDataString := fmt.Sprintf("%s\n", serviceName)
				statusDataString += fmt.Sprintf("git hash %s\n", sha)
				statusDataString += fmt.Sprintf("started %s\n", startTime.Format("Mon, 02 Jan 2006 15:04:05 EST"))
				statusDataString += fmt.Sprintf("uptime %s\n", time.Since(startTime))
				statusDataString += fmt.Sprintf("%d goroutines\n", int(analyticsMetrics.Goroutines.Value()))
				statusDataString += fmt.Sprintf("%.2f mb allocated\n", analyticsMetrics.MemoryAllocated.Value())
				statusDataString += fmt.Sprintf("%d ping stats entries received\n", int(analyticsMetrics.PingStatsMetrics.EntriesReceived.Value()))
				statusDataString += fmt.Sprintf("%d ping stats entries submitted\n", int(analyticsMetrics.PingStatsMetrics.EntriesSubmitted.Value()))
				statusDataString += fmt.Sprintf("%d ping stats entries queued\n", int(analyticsMetrics.PingStatsMetrics.EntriesQueued.Value()))
				statusDataString += fmt.Sprintf("%d ping stats entries flushed\n", int(analyticsMetrics.PingStatsMetrics.EntriesFlushed.Value()))
				statusDataString += fmt.Sprintf("%d relay stats entries received\n", int(analyticsMetrics.RelayStatsMetrics.EntriesReceived.Value()))
				statusDataString += fmt.Sprintf("%d relay stats entries submitted\n", int(analyticsMetrics.RelayStatsMetrics.EntriesSubmitted.Value()))
				statusDataString += fmt.Sprintf("%d relay stats entries queued\n", int(analyticsMetrics.RelayStatsMetrics.EntriesQueued.Value()))
				statusDataString += fmt.Sprintf("%d relay stats entries flushed\n", int(analyticsMetrics.RelayStatsMetrics.EntriesFlushed.Value()))

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

	// Start HTTP server
	{
		port := envvar.Get("PORT", "41001")
		if port == "" {
			core.Error("envvar PORT not set: %v", err)
			return 1
		}

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.Handle("/debug/vars", expvar.Handler())

		enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if err != nil {
			core.Error("could not parse envvar FEATURE_ENABLE_PPROF: %v", err)
		}
		if enablePProf {
			router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
		}

		fmt.Printf("starting http server on port %s\n", port)

		go func() {
			err = http.ListenAndServe(":"+port, router)
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
