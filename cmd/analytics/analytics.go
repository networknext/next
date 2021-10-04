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

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
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

func mainReturnWithCode() int {
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

	// Create an error chan for exiting from goroutines
	errChan := make(chan error, 1)

	var pingStatsWriter analytics.PingStatsWriter = &analytics.NoOpPingStatsWriter{}
	var relayStatsWriter analytics.RelayStatsWriter = &analytics.NoOpRelayStatsWriter{}

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

			pingStatsToPublishAtOnce, err := envvar.GetInt("PING_STATS_TO_PUBLISH_AT_ONCE", 10000)
			if err != nil {
				core.Error("could not parse PING_STATS_TO_PUBLISH_AT_ONCE: %v", err)
				return 1
			}

			bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
			if err != nil {
				core.Error("could not create ping stats bigquery client: %v", err)
				return 1
			}

			b := analytics.NewGoogleBigQueryPingStatsWriter(bqClient, &analyticsMetrics.PingStatsMetrics, pingStatsDataset, pingStatsTableName, pingStatsToPublishAtOnce)
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

			b := analytics.NewGoogleBigQueryRelayStatsWriter(bqClient, &analyticsMetrics.RelayStatsMetrics, relayStatsDataset, relayStatsTableName)
			relayStatsWriter = &b

			go b.WriteLoop(wg)
		}
	}

	pubsubEmulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")

	if gcpOK || pubsubEmulatorOK {

		if pubsubEmulatorOK {
			// Prefer to use the emulator instead of actual Google pubsub
			gcpProjectID = "local"

			// Use local ping stats and relay stats writer
			pingStatsWriter = &analytics.LocalPingStatsWriter{Metrics: &analyticsMetrics.PingStatsMetrics}
			relayStatsWriter = &analytics.LocalRelayStatsWriter{Metrics: &analyticsMetrics.RelayStatsMetrics}

			core.Debug("Detected pubsub emulator")
		}

		// Google pubsub forwarder
		{
			topicName := envvar.Get("PING_STATS_TOPIC_NAME", "ping_stats")
			subscriptionName := envvar.Get("PING_STATS_SUBSCRIPTION_NAME", "ping_stats")

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			pubsubForwarder, err := analytics.NewPingStatsPubSubForwarder(pubsubCtx, pingStatsWriter, &analyticsMetrics.PingStatsMetrics, gcpProjectID, topicName, subscriptionName)
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

			pubsubForwarder, err := analytics.NewRelayStatsPubSubForwarder(pubsubCtx, relayStatsWriter, &analyticsMetrics.RelayStatsMetrics, gcpProjectID, topicName, subscriptionName)
			if err != nil {
				core.Error("could not create relay stats pub sub forwarder: %v", err)
				return 1
			}

			wg.Add(1)
			go pubsubForwarder.Forward(ctx, wg)
		}
	}

	// Setup the status handler info
	type AnalyticsStatus struct {
		// Service Information
		ServiceName string `json:"service_name"`
		GitHash     string `json:"git_hash"`
		Started     string `json:"started"`
		Uptime      string `json:"uptime"`

		// Metrics
		Goroutines                 int     `json:"goroutines"`
		MemoryAllocated            float64 `json:"mb_allocated"`
		PingStatsEntriesReceived   int     `json:"ping_stats_entries_received"`
		PingStatsEntriesSubmitted  int     `json:"ping_stats_entries_submitted"`
		PingStatsEntriesQueued     int     `json:"ping_stats_entries_queued"`
		PingStatsEntriesFlushed    int     `json:"ping_stats_entries_flushed"`
		RelayStatsEntriesReceived  int     `json:"relay_stats_entries_received"`
		RelayStatsEntriesSubmitted int     `json:"relay_stats_entries_submitted"`
		RelayStatsEntriesQueued    int     `json:"relay_stats_entries_queued"`
		RelayStatsEntriesFlushed   int     `json:"relay_stats_entries_flushed"`
	}

	statusData := &AnalyticsStatus{}
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

				newStatusData := &AnalyticsStatus{}

				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				newStatusData.Goroutines = int(analyticsMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = analyticsMetrics.MemoryAllocated.Value()
				newStatusData.PingStatsEntriesReceived = int(analyticsMetrics.PingStatsMetrics.EntriesReceived.Value())
				newStatusData.PingStatsEntriesSubmitted = int(analyticsMetrics.PingStatsMetrics.EntriesSubmitted.Value())
				newStatusData.PingStatsEntriesQueued = int(analyticsMetrics.PingStatsMetrics.EntriesQueued.Value())
				newStatusData.PingStatsEntriesFlushed = int(analyticsMetrics.PingStatsMetrics.EntriesFlushed.Value())
				newStatusData.RelayStatsEntriesReceived = int(analyticsMetrics.RelayStatsMetrics.EntriesReceived.Value())
				newStatusData.RelayStatsEntriesSubmitted = int(analyticsMetrics.RelayStatsMetrics.EntriesSubmitted.Value())
				newStatusData.RelayStatsEntriesQueued = int(analyticsMetrics.RelayStatsMetrics.EntriesQueued.Value())
				newStatusData.RelayStatsEntriesFlushed = int(analyticsMetrics.RelayStatsMetrics.EntriesFlushed.Value())

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
