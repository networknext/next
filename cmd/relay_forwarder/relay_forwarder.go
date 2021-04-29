package main

import (
	"bytes"
	"context"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"

	"github.com/go-kit/kit/log/level"
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
		fmt.Println(err.Error())
		return 1
	}

	env, err := backend.GetEnv()
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	if gcpProjectID != "" {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
			return 1
		}
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	forwarderMetrics, err := metrics.NewRelayForwarderMetrics(ctx, metricsHandler, serviceName, "relay_forwarder", "Relay Forwarder", "riot relay packet")
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Get the Relay Gateway's Load Balancer's IP
	var gatewayAddr string
	gatewayAddr = envvar.Get("GATEWAY_LOAD_BALANCER_IP", "127.0.0.1:30000")
	// Verify the IP is valid if not testing locally
	if gcpProjectID != "" {
		ip := net.ParseIP(gatewayAddr)
		if ip == nil {
			level.Error(logger).Log("msg", fmt.Sprintf("could not parse relay gatway's load balancer's IP: %s", gatewayAddr), "err", err)
			return 1
		}
	}

	// Create an error channel for goroutines
	errChan := make(chan error, 1)

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
			syncTimer := helpers.NewSyncTimer(10 * time.Second)
			for {
				syncTimer.Run()

				forwarderMetrics.ForwarderServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				forwarderMetrics.ForwarderServiceMetrics.MemoryAllocated.Set(memoryUsed())

				statusDataString := fmt.Sprintf("%s\n", serviceName)
				statusDataString += fmt.Sprintf("git hash %s\n", sha)
				statusDataString += fmt.Sprintf("started %s\n", startTime.Format("Mon, 02 Jan 2006 15:04:05 EST"))
				statusDataString += fmt.Sprintf("uptime %s\n", time.Since(startTime))
				statusDataString += fmt.Sprintf("%d goroutines\n", int(forwarderMetrics.ForwarderServiceMetrics.Goroutines.Value()))
				statusDataString += fmt.Sprintf("%.2f mb allocated\n", forwarderMetrics.ForwarderServiceMetrics.MemoryAllocated.Value())
				statusDataString += fmt.Sprintf("%d invocations\n", int(forwarderMetrics.HandlerMetrics.Invocations.Value()))
				statusDataString += fmt.Sprintf("%d long durations\n", int(forwarderMetrics.HandlerMetrics.LongDuration.Value()))
				statusDataString += fmt.Sprintf("%d parse URL errors\n", int(forwarderMetrics.ErrorMetrics.ParseURLError.Value()))
				statusDataString += fmt.Sprintf("%d forward post errors\n", int(forwarderMetrics.ErrorMetrics.ForwardPostError.Value()))

				statusMutex.Lock()
				statusData = []byte(statusDataString)
				statusMutex.Unlock()
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

	// Create params for the main relay forwarder handlers
	forwarderParams := &transport.RelayForwarderParams{
		GatewayAddr: gatewayAddr,
		Metrics:     forwarderMetrics,
		Logger:      logger,
	}

	port := envvar.Get("PORT", "30006")
	fmt.Printf("starting http server on :%s\n", port)

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/status", serveStatusFunc).Methods("GET")
	router.HandleFunc("/relay_init", transport.ForwardPostHandlerFunc(forwarderParams)).Methods("POST")
	router.HandleFunc("/relay_update", transport.ForwardPostHandlerFunc(forwarderParams)).Methods("POST")
	router.Handle("/debug/vars", expvar.Handler())

	go func() {
		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			level.Error(logger).Log("err", err)
			errChan <- err
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-sigint:
		cancelFunc()
	case <-errChan:
		cancelFunc()
		return 1
	}
	return 0
}
