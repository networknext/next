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

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/magic"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
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
	serviceName := "magic_backend"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, cancel := context.WithCancel(context.Background())

	env, err := backend.GetEnv()
	if err != nil {
		core.Error("error getting env: %v", err)
		return 1
	}

	gcpProjectID := backend.GetGCPProjectID()
	if gcpProjectID != "" {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
			return 1
		}
	}

	logger := log.NewNopLogger()

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("failed to get metrics handler: %v", err)
		return 1
	}

	// Create magic metrics
	mbMetrics, err := metrics.NewMagicBackendMetrics(ctx, metricsHandler, serviceName, serviceName, "Magic Backend")
	if err != nil {
		core.Error("failed to create magic backend metrics: %v", err)
		return 1
	}

	magicMetadataInsertionFrequency, err := envvar.GetDuration("MAGIC_METADATA_INSERTION_FREQUENCY", time.Second)
	if err != nil {
		core.Error("failed to parse MAGIC_METADATA_INSERTION_FREQUENCY: %v", err)
		return 1
	}

	magicInstanceMetadataTimeout, err := envvar.GetDuration("MAGIC_INSTANCE_METADATA_TIMEOUT", time.Second*5)
	if err != nil {
		core.Error("failed to parse MAGIC_METADATA_INSERTION_FREQUENCY: %v", err)
		return 1
	}

	magicUpdateFrequency, err := envvar.GetDuration("MAGIC_UPDATE_FREQUENCY", time.Minute)
	if err != nil {
		core.Error("failed to parse MAGIC_UPDATE_FREQUENCY: %v", err)
		return 1
	}

	redisHostname := envvar.Get("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword := envvar.Get("REDIS_PASSWORD", "")
	redisMaxIdleConns, err := envvar.GetInt("REDIS_MAX_IDLE_CONNS", 10)
	if err != nil {
		core.Error("failed to parse REDIS_MAX_IDLE_CONNS: %v", err)
		return 1
	}
	redisMaxActiveConns, err := envvar.GetInt("REDIS_MAX_ACTIVE_CONNS", 64)
	if err != nil {
		core.Error("failed to parse REDIS_MAX_ACTIVE_CONNS: %v", err)
		return 1
	}

	instanceID, err := backend.GetInstanceID(env)
	if err != nil {
		core.Error("failed to get relay backend instance ID: %v", err)
		return 1
	}
	core.Debug("VM Instance ID: %s", instanceID)

	// Create the magic service for the magic backend
	magicService, err := magic.NewMagicService(magicInstanceMetadataTimeout, instanceID, time.Now().UTC(), redisHostname, redisPassword, redisMaxIdleConns, redisMaxActiveConns, mbMetrics)
	if err != nil {
		core.Error("failed to create magic service: %v", err)
		return 1
	}

	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// Start the metdata insertion goroutine
	metadataTicker := time.NewTicker(magicMetadataInsertionFrequency)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-metadataTicker.C:
				latestMetadata := magicService.CreateInstanceMetadata()
				if err := magicService.InsertInstanceMetadata(latestMetadata); err != nil {
					core.Error("failed to insert instance metadata: %v", err)
					errChan <- err
					return
				}
			}
		}
	}()

	// Start the magic update goroutine
	var isOldestInstance bool
	updateMagicTicker := time.NewTicker(magicUpdateFrequency)

	wg.Add(1)
	go func() {
		defer wg.Done()

		var err error

		// For local testing, initialize redis with magic values
		if env == "local" {
			core.Debug("initializing magic values for local testing")

			if err = magicService.UpdateMagicValues(); err != nil {
				core.Error("failed to update magic values: %v", err)
				errChan <- err
				return
			}
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-updateMagicTicker.C:
				isOldestInstance, err = magicService.IsOldestInstance()
				if err != nil {
					core.Error("failed to verify if oldest instance: %v", err)
					errChan <- err
					return
				}

				if !isOldestInstance {
					continue
				}

				if err = magicService.UpdateMagicValues(); err != nil {
					core.Error("failed to update magic values: %v", err)
					errChan <- err
					return
				}

				core.Debug("updated magic values successfully")
			}
		}
	}()

	// Setup the status handler info
	statusData := &metrics.MagicStatus{}
	var statusMutex sync.RWMutex
	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				mbMetrics.MagicServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				mbMetrics.MagicServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.MagicStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()
				newStatusData.OldestInstance = isOldestInstance

				// Service Metrics
				newStatusData.Goroutines = int(mbMetrics.MagicServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = mbMetrics.MagicServiceMetrics.MemoryAllocated.Value()

				// Success Metrics
				newStatusData.InsertInstanceMetadataSuccess = int(mbMetrics.InsertInstanceMetadataSuccess.Value())
				newStatusData.UpdateMagicValuesSuccess = int(mbMetrics.UpdateMagicValuesSuccess.Value())
				newStatusData.GetMagicValueSuccess = int(mbMetrics.GetMagicValueSuccess.Value())
				newStatusData.SetMagicValueSuccess = int(mbMetrics.SetMagicValueSuccess.Value())

				// Error Metrics
				newStatusData.InsertInstanceMetadataFailure = int(mbMetrics.ErrorMetrics.InsertInstanceMetadataFailure.Value())
				newStatusData.UpdateMagicValuesFailure = int(mbMetrics.ErrorMetrics.UpdateMagicValuesFailure.Value())
				newStatusData.GetMagicValueFailure = int(mbMetrics.ErrorMetrics.GetMagicValueFailure.Value())
				newStatusData.SetMagicValueFailure = int(mbMetrics.ErrorMetrics.SetMagicValueFailure.Value())
				newStatusData.ReadFromRedisFailure = int(mbMetrics.ErrorMetrics.ReadFromRedisFailure.Value())
				newStatusData.MarshalFailure = int(mbMetrics.ErrorMetrics.MarshalFailure.Value())
				newStatusData.UnmarshalFailure = int(mbMetrics.ErrorMetrics.UnmarshalFailure.Value())

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
		port := envvar.Get("PORT", "41007")
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
		cancel()

		// Wait for essential goroutines to finish up
		wg.Wait()

		core.Debug("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		core.Debug("received error from goroutine")
		cancel()

		// Wait for essential goroutines to finish up
		wg.Wait()

		return 1
	}
}
