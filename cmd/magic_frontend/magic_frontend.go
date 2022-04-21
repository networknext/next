package main

import (
	"bytes"
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
	serviceName := "magic_frontend"
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
	mfMetrics, err := metrics.NewMagicFrontendMetrics(ctx, metricsHandler, serviceName, serviceName, "Magic Frontend")
	if err != nil {
		core.Error("failed to create magic backend metrics: %v", err)
		return 1
	}

	magicPollFrequency, err := envvar.GetDuration("MAGIC_POLL_FREQUENCY", time.Second)
	if err != nil {
		core.Error("failed to parse MAGIC_POLL_FREQUENCY: %v", err)
		return 1
	}

	magicPollRetryLimit, err := envvar.GetInt("MAGIC_POLL_RETRY_LIMIT", 5)
	if err != nil {
		core.Error("failed to parse MAGIC_POLL_RETRY_LIMIT: %v", err)
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

	// Create the magic service for the magic frontend
	magicService, err := magic.NewMagicService(time.Second, "", time.Now().UTC(), redisHostname, redisPassword, redisMaxIdleConns, redisMaxActiveConns, mfMetrics)
	if err != nil {
		core.Error("failed to create magic service: %v", err)
		return 1
	}

	// Track the most recent magic values
	// The order of the magic values is always upcoming, current, previous
	var magicMutex sync.RWMutex
	var combinedMagic [24]byte

	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// Start the poll goroutine
	pollTicker := time.NewTicker(magicPollFrequency)

	wg.Add(1)
	go func() {
		defer wg.Done()

		var err error
		var numRetries int

		var cachedUpcomingMagic [8]byte
		var cachedCurrentMagic [8]byte
		var cachedPreviousMagic [8]byte

		for {
			select {
			case <-ctx.Done():
				return
			case <-pollTicker.C:
				if numRetries > magicPollRetryLimit {
					core.Error("reached poll retry limit (%d)")
					errChan <- err
					return
				}

				newUpcomingMagic, err := magicService.GetMagicValue(magic.MagicUpcomingKey)
				if err != nil {
					numRetries++
					core.Error("failed to get upcoming magic (retry count %d): %v", numRetries, err)
					continue
				}

				newCurrentMagic, err := magicService.GetMagicValue(magic.MagicCurrentKey)
				if err != nil {
					numRetries++
					core.Error("failed to get current magic (retry count %d): %v", numRetries, err)
					continue
				}

				newPreviousMagic, err := magicService.GetMagicValue(magic.MagicPreviousKey)
				if err != nil {
					numRetries++
					core.Error("failed to get previous magic (retry count %d): %v", numRetries, err)
					continue
				}

				// Reset retry counter
				numRetries = 0

				if newUpcomingMagic.MagicBytes == cachedUpcomingMagic && newCurrentMagic.MagicBytes == cachedCurrentMagic && newPreviousMagic.MagicBytes == cachedPreviousMagic {
					continue
				}

				// Combine the magic values under mutex
				magicMutex.Lock()
				for i, val := range newUpcomingMagic.MagicBytes {
					combinedMagic[i] = val
				}
				for i, val := range newCurrentMagic.MagicBytes {
					combinedMagic[i+8] = val
				}
				for i, val := range newPreviousMagic.MagicBytes {
					combinedMagic[i+16] = val
				}
				magicMutex.Unlock()

				// Update the cached values
				cachedUpcomingMagic = newUpcomingMagic.MagicBytes
				cachedCurrentMagic = newCurrentMagic.MagicBytes
				cachedPreviousMagic = newPreviousMagic.MagicBytes

				core.Debug("received new magic values")
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
				mfMetrics.MagicServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				mfMetrics.MagicServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.MagicStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				// Service Metrics
				newStatusData.Goroutines = int(mfMetrics.MagicServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = mfMetrics.MagicServiceMetrics.MemoryAllocated.Value()

				// Success Metrics
				newStatusData.GetMagicValueSuccess = int(mfMetrics.GetMagicValueSuccess.Value())

				// Error Metrics
				newStatusData.GetMagicValueFailure = int(mfMetrics.ErrorMetrics.GetMagicValueFailure.Value())
				newStatusData.ReadFromRedisFailure = int(mfMetrics.ErrorMetrics.ReadFromRedisFailure.Value())
				newStatusData.MarshalFailure = int(mfMetrics.ErrorMetrics.MarshalFailure.Value())
				newStatusData.UnmarshalFailure = int(mfMetrics.ErrorMetrics.UnmarshalFailure.Value())

				statusMutex.Lock()
				statusData = newStatusData
				statusMutex.Unlock()

				time.Sleep(time.Second * 10)
			}
		}()
	}

	// Setup HTTP handlers
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

	serveCombinedMagicFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		magicMutex.RLock()
		data := combinedMagic
		magicMutex.RUnlock()
		buffer := bytes.NewBuffer(data[:])
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// Start HTTP server
	{
		port := envvar.Get("PORT", "41008")
		if port == "" {
			core.Error("PORT not set")
			return 1
		}

		fmt.Printf("starting http server on port %s\n", port)

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.HandleFunc("/magic", serveCombinedMagicFunc).Methods("GET")
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
