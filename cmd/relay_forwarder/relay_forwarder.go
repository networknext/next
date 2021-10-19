package main

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"net"
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
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"

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
	serviceName := "relay_forwarder"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	// Setup the service
	ctx, cancelFunc := context.WithCancel(context.Background())
	gcpProjectID := backend.GetGCPProjectID()
	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		core.Error("failed to create logger: %v", err)
		return 1
	}

	env, err := backend.GetEnv()
	if err != nil {
		core.Error("failed to get env: %v", err)
		return 1
	}

	if gcpProjectID != "" {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialize stackdriver profiler: %v", err)
			return 1
		}
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("failed to get metrics handler: %v", err)
		return 1
	}

	forwarderMetrics, err := metrics.NewRelayForwarderMetrics(ctx, metricsHandler, serviceName, "relay_forwarder", "Relay Forwarder", "riot relay packet")
	if err != nil {
		core.Error("failed to create relay forwarder metrics: %v", err)
		return 1
	}

	// Get the Relay Gateway's Load Balancer's IP
	gatewayAddr := envvar.Get("GATEWAY_LOAD_BALANCER_IP", "127.0.0.1:30000")
	// Verify the IP is valid if not testing locally
	if gcpProjectID != "" {
		ip := net.ParseIP(gatewayAddr)
		if ip == nil {
			core.Error("failed to parse relay gateway's load balancer's IP (%s)", gatewayAddr)
			return 1
		}
	}

	// Create an error channel for goroutines
	errChan := make(chan error, 1)

	// Setup the status handler info
	statusData := &metrics.RelayForwarderStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				forwarderMetrics.ForwarderServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				forwarderMetrics.ForwarderServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.RelayForwarderStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				// Service Metrics
				newStatusData.Goroutines = int(forwarderMetrics.ForwarderServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = forwarderMetrics.ForwarderServiceMetrics.MemoryAllocated.Value()

				// Handler Metrics
				newStatusData.Invocations = int(forwarderMetrics.HandlerMetrics.Invocations.Value())
				newStatusData.DurationMs = forwarderMetrics.HandlerMetrics.Duration.Value()

				// Error Metrics
				newStatusData.ParseURLError = int(forwarderMetrics.ErrorMetrics.ParseURLError.Value())
				newStatusData.ForwardPostError = int(forwarderMetrics.ErrorMetrics.ForwardPostError.Value())

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

	// Create params for the main relay forwarder handlers
	forwarderParams := &transport.RelayForwarderParams{
		GatewayAddr: gatewayAddr,
		Metrics:     forwarderMetrics,
	}

	port := envvar.Get("PORT", "30006")
	if port == "" {
		core.Error("PORT not set")
		return 1
	}

	fmt.Printf("starting http server on :%s\n", port)

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/status", serveStatusFunc).Methods("GET")
	router.HandleFunc("/relay_init", transport.ForwardPostHandlerFunc(forwarderParams)).Methods("POST")
	router.HandleFunc("/relay_update", transport.ForwardPostHandlerFunc(forwarderParams)).Methods("POST")
	router.Handle("/debug/vars", expvar.Handler())

	enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
	if err != nil {
		core.Error("could not parse envvar FEATURE_ENABLE_PPROF: %v", err)
	}
	if enablePProf {
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

	go func() {
		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			core.Error("failed to start http server: %v", err)
			errChan <- err
		}
	}()

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan: // Exit with an error code of 0 if we receive SIGINT or SIGTERM
		fmt.Println("Received shutdown signal.")

		cancelFunc()

		fmt.Println("Successfully shutdown.")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		cancelFunc()
		return 1
	}
}
