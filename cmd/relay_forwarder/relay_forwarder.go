package main

import (
	"bytes"
	"context"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
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

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

type RelayForwarderParams struct {
	GatewayAddr string
	Metrics     *metrics.RelayForwarderMetrics
	Logger      log.Logger
}

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
	if gcpProjectID != "" {
		gatewayAddr = envvar.Get("GATEWAY_LOAD_BALANCER_IP", "")
		ip := net.ParseIP(gatewayAddr)
		if ip == nil {
			level.Error(logger).Log("msg", fmt.Sprintf("could not parse relay gatway's load balancer's IP: %s", gatewayAddr), "err", err)
			return 1
		}
	} else {
		gatewayAddr = envvar.Get("RELAY_GATEWAY_ADDRESS", "127.0.0.1:30000")
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
	forwarderParams := &RelayForwarderParams{
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
	router.HandleFunc("/relay_init", forwardPost(forwarderParams)).Methods("POST")
	router.HandleFunc("/relay_update", forwardPost(forwarderParams)).Methods("POST")
	router.Handle("/debug/vars", expvar.Handler())

	go func() {
		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1) // TODO: don't os.Exit() here, but find a way to exit
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-sigint:
		cancelFunc()
	}
	return 0
}

func forwardPost(params *RelayForwarderParams) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		durationStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(durationStart).Milliseconds())
			params.Metrics.HandlerMetrics.Duration.Set(float64(milliseconds))
			if milliseconds > 100 {
				params.Metrics.HandlerMetrics.LongDuration.Add(1)
			}
			params.Metrics.HandlerMetrics.Invocations.Add(1)
		}()

		// Parse the remote address to get the origin URL
		origin, err := url.Parse(fmt.Sprintf("//%s", r.RemoteAddr))
		if err != nil {
			level.Error(params.Logger).Log("msg", fmt.Sprintf("error parsing request remote addr as URL: %s", r.RemoteAddr), "err", err)
			w.WriteHeader(http.StatusInternalServerError)
			params.Metrics.ErrorMetrics.ParseURLError.Add(1)
			return
		}

		// Get the requested path (i.e. /relay_update)
		requestedPath := r.RequestURI

		// Create a reverse proxy
		reverseProxy := httputil.NewSingleHostReverseProxy(origin)

		// Modify the director to forward the request to the relay gateway
		reverseProxy.Director = func(req *http.Request) {
			req.Header.Add("X-Forwarded-Host", req.Host)
			req.Header.Add("X-Origin-Host", origin.Host)
			req.URL.Scheme = "http"
			req.URL.Host = params.GatewayAddr
			req.URL.Path = requestedPath
		}

		// Add an error handler to use our logger
		reverseProxy.ErrorHandler = func(writer http.ResponseWriter, req *http.Request, err error) {
			if err != nil {
				level.Error(params.Logger).Log("msg", "error reaching relay gateway", "err", err)
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				params.Metrics.ErrorMetrics.ForwardPostError.Add(1)
			}
		}

		// Serve the request
		reverseProxy.ServeHTTP(w, r)
	}
}
