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
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/config"
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

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func mainReturnWithCode() int {
	serviceName := "billing"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	logger := log.NewNopLogger()
	wg := &sync.WaitGroup{}

	// Setup the service
	gcpProjectID := backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""

	env, err := backend.GetEnv()
	if err != nil {
		core.Error("error getting env: %v", err)
		return 1
	}

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("error getting metrics handler: %v", err)
		return 1
	}

	// Create billing metrics
	billingServiceMetrics, err := metrics.NewBillingServiceMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to create billing service metrics: %v", err)
		return 1
	}

	if gcpOK {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
			return 1
		}
	}

	// Get billing feature config
	var featureConfig config.Config
	envVarConfig := config.NewEnvVarConfig([]config.Feature{
		{
			Name:        "FEATURE_BILLING2",
			Enum:        config.FEATURE_BILLING2,
			Value:       true,
			Description: "Receives BillingEntry2 types from Google Pub/Sub and writes them to BigQuery",
		},
	})
	featureConfig = envVarConfig
	featureBilling2 := featureConfig.FeatureEnabled(config.FEATURE_BILLING2)

	// Create no-op biller
	var biller2 billing.Biller = &billing.NoOpBiller{}

	if gcpOK {
		// Google BigQuery
		if featureBilling2 {
			billingDataset := envvar.Get("GOOGLE_BIGQUERY_DATASET_BILLING", "")
			if billingDataset == "" {
				core.Error("GOOGLE_BIGQUERY_DATASET_BILLING not set")
				return 1
			}

			batchSize, err := envvar.GetInt("GOOGLE_BIGQUERY_BATCH_SIZE", billing.DefaultBigQueryBatchSize)
			if err != nil {
				core.Error("could not parse GOOGLE_BIGQUERY_BATCH_SIZE: %v", err)
				return 1
			}

			summaryBatchSize, err := envvar.GetInt("GOOGLE_BIGQUERY_SUMMARY_BATCH_SIZE", int(billing.DefaultBigQueryBatchSize/10))
			if err != nil {
				core.Error("could not parse GOOGLE_BIGQUERY_SUMMARY_BATCH_SIZE: %v", err)
				return 1
			}

			batchSizePercent, err := envvar.GetFloat("FEATURE_BILLING2_BATCH_SIZE_PERCENT", 0.80)
			if err != nil {
				core.Error("could not parse FEATURE_BILLING2_BATCH_SIZE_PERCENT: %v", err)
				return 1
			}

			billing2TableName := envvar.Get("FEATURE_BILLING2_GOOGLE_BIGQUERY_TABLE_BILLING", "billing2")

			billing2SummaryTableName := envvar.Get("FEATURE_BILLING2_GOOGLE_BIGQUERY_TABLE_BILLING_SUMMARY", "billing2_session_summary")

			// Pass context without cancel to ensure writing continues even past reception of shutdown signal
			bqClient, err := bigquery.NewClient(context.Background(), gcpProjectID)
			if err != nil {
				core.Error("could not create BigQuery client: %v", err)
				return 1
			}

			b := billing.GoogleBigQueryClient{
				Metrics:              &billingServiceMetrics.BillingMetrics,
				TableInserter:        bqClient.Dataset(billingDataset).Table(billing2TableName).Inserter(),
				SummaryTableInserter: bqClient.Dataset(billingDataset).Table(billing2SummaryTableName).Inserter(),
				BatchSize:            batchSize,
				SummaryBatchSize:     summaryBatchSize,
				BatchSizePercent:     batchSizePercent,
				FeatureBilling2:      featureBilling2,
			}

			// Set the Biller to BigQuery
			biller2 = &b

			// Start the background WriteLoop and SummaryWriteLoop to batch write to BigQuery
			wg.Add(1)
			go b.WriteLoop2(ctx, wg)

			wg.Add(1)
			go b.SummaryWriteLoop2(ctx, wg)
		}
	}

	errChan := make(chan error, 1)

	emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
	if emulatorOK { // Prefer to use the emulator instead of actual Google pubsub
		gcpProjectID = "local"

		// Use the local biller
		biller2 = &billing.LocalBiller{
			Metrics: &billingServiceMetrics.BillingMetrics,
		}

		core.Debug("detected pubsub emulator")
	}

	if gcpOK || emulatorOK {
		// Google pubsub forwarder
		{
			numRecvGoroutines, err := envvar.GetInt("NUM_RECEIVE_GOROUTINES", 10)
			if err != nil {
				core.Error("could not parse NUM_RECEIVE_GOROUTINES: %v", err)
				return 1
			}

			if featureBilling2 {

				core.Debug("Billing2 enabled")

				entryVeto, err := envvar.GetBool("BILLING_ENTRY_VETO", false)
				if err != nil {
					core.Error("could not parse BILLING_ENTRY_VETO: %v", err)
					return 1
				}

				maxRetries, err := envvar.GetInt("FEATURE_BILLING2_MAX_RETRIES", 25)
				if err != nil {
					core.Error("could not parse FEATURE_BILLING2_MAX_RETRIES: %v", err)
					return 1
				}

				retryTime, err := envvar.GetDuration("FEATURE_BILLING2_RETRY_TIME", time.Second*1)
				if err != nil {
					core.Error("could not parse FEATURE_BILLING2_RETRY_TIME: %v", err)
					return 1
				}

				topicName := envvar.Get("FEATURE_BILLING2_TOPIC_NAME", "billing2")
				subscriptionName := envvar.Get("FEATURE_BILLING2_SUBSCRIPTION_NAME", "billing2")

				pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
				defer cancelFunc()

				pubsubForwarder, err := billing.NewPubSubForwarder(pubsubCtx, biller2, entryVeto, maxRetries, retryTime, &billingServiceMetrics.BillingMetrics, gcpProjectID, topicName, subscriptionName, numRecvGoroutines)
				if err != nil {
					core.Error("could not create pubsub forwarder: %v", err)
					return 1
				}

				// Start the pubsub forwarder
				wg.Add(1)
				go pubsubForwarder.Forward2(ctx, wg, errChan)
			}
		}
	}

	// Setup the status handler info
	statusData := &metrics.BillingStatus{}
	var statusMutex sync.RWMutex

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

				newStatusData := &metrics.BillingStatus{}

				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				newStatusData.Goroutines = int(billingServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = billingServiceMetrics.MemoryAllocated.Value()
				newStatusData.Billing2EntriesReceived = int(billingServiceMetrics.BillingMetrics.Entries2Received.Value())
				newStatusData.Billing2EntriesSubmitted = int(billingServiceMetrics.BillingMetrics.Entries2Submitted.Value())
				newStatusData.Billing2EntriesQueued = int(billingServiceMetrics.BillingMetrics.Entries2Queued.Value())
				newStatusData.Billing2EntriesFlushed = int(billingServiceMetrics.BillingMetrics.Entries2Flushed.Value())
				newStatusData.Billing2SummaryEntriesSubmitted = int(billingServiceMetrics.BillingMetrics.SummaryEntries2Submitted.Value())
				newStatusData.Billing2SummaryEntriesQueued = int(billingServiceMetrics.BillingMetrics.SummaryEntries2Queued.Value())
				newStatusData.Billing2SummaryEntriesFlushed = int(billingServiceMetrics.BillingMetrics.SummaryEntries2Flushed.Value())
				newStatusData.Billing2EntriesWithNaN = int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2EntriesWithNaN.Value())
				newStatusData.Billing2InvalidEntries = int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2InvalidEntries.Value())
				newStatusData.Billing2ReadFailures = int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2ReadFailure.Value())
				newStatusData.Billing2WriteFailures = int(billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2WriteFailure.Value())

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
		port := envvar.Get("PORT", "41000")
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

		enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if err != nil {
			core.Error("could not parse FEATURE_ENABLE_PPROF: %v", err)
		}
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
