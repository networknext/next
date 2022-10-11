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

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"

	"github.com/networknext/backend/modules-old/backend"
	"github.com/networknext/backend/modules-old/metrics"
	"github.com/networknext/backend/modules-old/pingdom"
	"github.com/networknext/backend/modules-old/transport"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "pingdom"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, commitHash, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, cancel := context.WithCancel(context.Background())

	env := backend.GetEnv()

	gcpProjectID := backend.GetGCPProjectID()
	if gcpProjectID == "" {
		core.Error("pingdom must be run in the cloud because requires BigQuery read/write access")
		return 1
	}

	logger := log.NewNopLogger()

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("failed to get metrics handler: %v", err)
		return 1
	}

	// Create pingdom metrics
	pingdomMetrics, err := metrics.NewPingdomMetrics(ctx, metricsHandler, serviceName, "pingdom", "Pingdom")
	if err != nil {
		core.Error("failed to create pingdom metrics: %v", err)
		return 1
	}

	if gcpProjectID != "" {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
			return 1
		}
	}

	pingdomApiToken := envvar.GetString("PINGDOM_API_TOKEN", "")
	if pingdomApiToken == "" {
		core.Error("PINGDOM_API_TOKEN not set")
		return 1
	}

	bqClient, err := bigquery.NewClient(context.Background(), gcpProjectID)
	if err != nil {
		core.Error("failed to create BigQuery client: %v", err)
		return 1
	}

	bqDatasetName := envvar.GetString("GOOGLE_BIGQUERY_DATASET_PINGDOM", "")
	if bqDatasetName == "" {
		core.Error("GOOGLE_BIGQUERY_DATASET_PINGDOM not set")
		return 1
	}

	bqTableName := envvar.GetString("GOOGLE_BIGQUERY_TABLE_PINGDOM", "")
	if bqTableName == "" {
		core.Error("GOOGLE_BIGQUERY_TABLE_PINGDOM not set")
		return 1
	}

	chanSize := envvar.GetInt("PINGDOM_CHANNEL_SIZE", 100)

	pingdomClient, err := pingdom.NewPingdomClient(pingdomApiToken, pingdomMetrics, bqClient, gcpProjectID, bqDatasetName, bqTableName, chanSize)
	if err != nil {
		core.Error("failed to create pingdom client: %v", err)
		return 1
	}

	portalHostname := envvar.GetString("PORTAL_HOSTNAME", "")
	if portalHostname == "" {
		core.Error("PORTAL_HOSTNAME not set")
		return 1
	}

	serverBackendHostname := envvar.GetString("SERVER_BACKEND_HOSTNAME", "")
	if serverBackendHostname == "" {
		core.Error("SERVER_BACKEND_HOSTNAME not set")
		return 1
	}

	portalID, err := pingdomClient.GetIDForHostname(portalHostname)
	if err != nil {
		core.Error("failed to get portal pingdom ID: %v", err)
		return 1
	}

	serverBackendID, err := pingdomClient.GetIDForHostname(serverBackendHostname)
	if err != nil {
		core.Error("failed to get server backend pingdom ID: %v", err)
		return 1
	}

	pingFrequency := envvar.GetDuration("PINGDOM_API_PING_FREQUENCY", time.Second*10)

	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// Start the goroutine for calculating uptime from the Pingdom API
	wg.Add(1)
	go pingdomClient.GetUptimeForIDs(ctx, portalID, serverBackendID, pingFrequency, &wg, errChan)

	// Start the goroutine for inserting uptime data to BigQuery
	wg.Add(1)
	go pingdomClient.WriteLoop(ctx, &wg)

	// Setup the status handler info
	statusData := &metrics.PingdomStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				pingdomMetrics.PingdomServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				pingdomMetrics.PingdomServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.PingdomStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = commitHash
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				// Service Metrics
				newStatusData.Goroutines = int(pingdomMetrics.PingdomServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = pingdomMetrics.PingdomServiceMetrics.MemoryAllocated.Value()

				// Success Metrics
				newStatusData.CreatePingdomUptime = int(pingdomMetrics.CreatePingdomUptime.Value())
				newStatusData.BigQueryWriteSuccess = int(pingdomMetrics.BigQueryWriteSuccess.Value())

				// Error Metrics
				newStatusData.PingdomAPICallFailure = int(pingdomMetrics.ErrorMetrics.PingdomAPICallFailure.Value())
				newStatusData.BigQueryReadFailure = int(pingdomMetrics.ErrorMetrics.BigQueryReadFailure.Value())
				newStatusData.BigQueryWriteFailure = int(pingdomMetrics.ErrorMetrics.BigQueryWriteFailure.Value())
				newStatusData.BadSummaryPerformanceRequest = int(pingdomMetrics.ErrorMetrics.BadSummaryPerformanceRequest.Value())

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
		port := envvar.GetString("PORT", "41006")
		if port == "" {
			core.Error("PORT not set")
			return 1
		}

		fmt.Printf("starting http server on port %s\n", port)

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))
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
		cancel()

		// Wait for essential goroutines to finish up
		wg.Wait()

		core.Debug("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		cancel()
		return 1
	}
}
