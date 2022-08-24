/*
	Network Next. You control the network.
	Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

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

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	md "github.com/networknext/backend/modules/match_data"
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

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func mainReturnWithCode() int {
	serviceName := "match_data"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	logger := log.NewNopLogger()
	wg := &sync.WaitGroup{}

	// Setup the service
	gcpProjectID := backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""

	env := backend.GetEnv()

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("error getting metrics handler: %v", err)
		return 1
	}

	// Create match data service metrics
	matchDataServiceMetrics, err := metrics.NewMatchDataServiceMetrics(ctx, metricsHandler, serviceName, "match_data", "Match Data", "match data entry")
	if err != nil {
		core.Error("failed to create match data service metrics: %v", err)
		return 1
	}

	if gcpOK {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
			return 1
		}
	}

	// Create no-op matcher
	var matcher md.Matcher = &md.NoOpMatcher{}

	if gcpOK {
		// Google BigQuery
		matchDataDataset := envvar.Get("GOOGLE_BIGQUERY_DATASET_MATCH_DATA", "")
		if matchDataDataset == "" {
			core.Error("GOOGLE_BIGQUERY_DATASET_MATCH_DATA not set")
			return 1
		}

		batchSize  := envvar.GetInt("GOOGLE_BIGQUERY_BATCH_SIZE", md.DefaultBigQueryBatchSize)

		batchSizePercent := envvar.GetFloat("GOOGLE_BIGQUERY_BATCH_SIZE_PERCENT", 0.80)
		if err != nil {
			core.Error("could not parse GOOGLE_BIGQUERY_BATCH_SIZE_PERCENT: %v", err)
			return 1
		}

		matchDataTableName := envvar.Get("GOOGLE_BIGQUERY_TABLE_MATCH_DATA", "match_data")

		// Pass context without cancel to ensure writing continues even past reception of shutdown signal
		bqClient, err := bigquery.NewClient(context.Background(), gcpProjectID)
		if err != nil {
			core.Error("could not create BigQuery client: %v", err)
			return 1
		}

		b := md.GoogleBigQueryClient{
			Metrics:          matchDataServiceMetrics.MatchDataMetrics,
			TableInserter:    bqClient.Dataset(matchDataDataset).Table(matchDataTableName).Inserter(),
			BatchSize:        batchSize,
			BatchSizePercent: batchSizePercent,
		}

		// Set the Matcher to BigQuery
		matcher = &b

		// Start the background WriteLoop to batch write to BigQuery
		wg.Add(1)
		go b.WriteLoop(ctx, wg)
	}

	errChan := make(chan error, 1)

	emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
	if emulatorOK { // Prefer to use the emulator instead of actual Google pubsub
		gcpProjectID = "local"

		// Use the local matcher
		matcher = &md.LocalMatcher{
			Metrics: matchDataServiceMetrics.MatchDataMetrics,
		}

		core.Debug("detected pubsub emulator")
	}

	if gcpOK || emulatorOK {
		// Google pubsub forwarder
		{
			numRecvGoroutines := envvar.GetInt("NUM_RECEIVE_GOROUTINES", 10)
			entryVeto := envvar.GetBool("MATCH_DATA_ENTRY_VETO", false)
			maxRetries := envvar.GetInt("MATCH_DATA_MAX_RETRIES", 25)
			retryTime := envvar.GetDuration("MATCH_DATA_RETRY_TIME", time.Second*1)
			topicName := envvar.Get("MATCH_DATA_TOPIC_NAME", "match_data")
			subscriptionName := envvar.Get("MATCH_DATA_SUBSCRIPTION_NAME", "match_data")

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			pubsubForwarder, err := md.NewPubSubForwarder(pubsubCtx, matcher, entryVeto, maxRetries, retryTime, matchDataServiceMetrics.MatchDataMetrics, gcpProjectID, topicName, subscriptionName, numRecvGoroutines)
			if err != nil {
				core.Error("could not create pubsub forwarder: %v", err)
				return 1
			}

			// Start the pubsub forwarder
			wg.Add(1)
			go pubsubForwarder.Forward(ctx, wg, errChan)
		}
	}

	// Setup the status handler info
	statusData := &metrics.MatchDataStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				matchDataServiceMetrics.ServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				matchDataServiceMetrics.ServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.MatchDataStatus{}

				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				newStatusData.Goroutines = int(matchDataServiceMetrics.ServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = matchDataServiceMetrics.ServiceMetrics.MemoryAllocated.Value()
				newStatusData.MatchDataEntriesReceived = int(matchDataServiceMetrics.MatchDataMetrics.EntriesReceived.Value())
				newStatusData.MatchDataEntriesSubmitted = int(matchDataServiceMetrics.MatchDataMetrics.EntriesSubmitted.Value())
				newStatusData.MatchDataEntriesQueued = int(matchDataServiceMetrics.MatchDataMetrics.EntriesQueued.Value())
				newStatusData.MatchDataEntriesFlushed = int(matchDataServiceMetrics.MatchDataMetrics.EntriesFlushed.Value())
				newStatusData.MatchDataEntriesWithNaN = int(matchDataServiceMetrics.MatchDataMetrics.ErrorMetrics.MatchDataEntriesWithNaN.Value())
				newStatusData.MatchDataInvalidEntries = int(matchDataServiceMetrics.MatchDataMetrics.ErrorMetrics.MatchDataInvalidEntries.Value())
				newStatusData.MatchDataReadFailures = int(matchDataServiceMetrics.MatchDataMetrics.ErrorMetrics.MatchDataReadFailure.Value())
				newStatusData.MatchDataWriteFailures = int(matchDataServiceMetrics.MatchDataMetrics.ErrorMetrics.MatchDataWriteFailure.Value())

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
		port := envvar.Get("PORT", "41003")
		if port == "" {
			core.Error("PORT not set")
			return 1
		}

		fmt.Printf("starting http server on port %s\n", port)

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.Handle("/debug/vars", expvar.Handler())

		enablePProf := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
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
