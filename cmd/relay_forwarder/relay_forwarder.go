package main

import (
	"bytes"
	"expvar"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/envvar"
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
	return mainReturnWithCode()
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

	fowarderMetrics, err := metrics.NewRelayForwarderMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Get the Relay Gateway's Load Balancer's IP
	var lbAddr string
	if gcpProjectID != "" {
		lbAddr = envvar.Get("GATEWAY_LOAD_BALANCER_IP", "")
		ip := net.ParseIP(lbAddr)
		if ip == nil {
			level.Error(logger).Log("msg", "could not parse relay gatway's load balancer's IP", "err", err)
			return 1
		}
	} else {
		lbAddr = envvar.Get("RELAY_GATEWAY_ADDRESS", "127.0.0.1:30000")
	}

	// Create init and update URIs
	initURI, err := url.Parse(fmt.Sprintf("http://%s/relay_init", lbAddr))
	if err != nil {
		level.Error(logger).Log("msg", "could not parse relay init URI", "err", err)
		return 1
	}

	updateURI, err := url.Parse(fmt.Sprintf("http://%s/relay_update", lbAddr))
	if err != nil {
		level.Error(logger).Log("msg", "could not parse relay update URI", "err", err)
		return 1
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

	port := envvar.Get("PORT", "30006")
	fmt.Printf("starting http server on port %s\n", port)

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/status", serveStatusFunc).Methods("GET")
	router.HandleFunc("/relay_init", forwardPost(initURI)).Methods("POST")
	router.HandleFunc("/relay_update", forwardPost(updateURI)).Methods("POST")
	router.Handle("/debug/vars", expvar.Handler())

	go func() {
		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
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

// func forwardGet(address string, octet bool) func(w http.ResponseWriter, r *http.Request) {

// 	return func(w http.ResponseWriter, r *http.Request) {
// 		resp, err := http.Get(address)
// 		if err != nil {
// 			log.Printf("error forwarding get: %s", err.Error())
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		body, err := ioutil.ReadAll(resp.Body)
// 		if err != nil {
// 			log.Printf("error reading response body: %s", err.Error())
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		defer resp.Body.Close()
// 		w.WriteHeader(resp.StatusCode)
// 		if octet {
// 			w.Header().Set("Content-Type", "application/octet-stream")
// 		}
// 		w.Write(body)
// 	}
// }

func forwardPost(requestURI *url.URL) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		// Create a reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(requestURI)
		// Serve the request
		proxy.ServeHTTP(w, r)

		// reqBody, err := ioutil.ReadAll(r.Body)
		// if err != nil {
		// 	fmt.Printf("error reading response body: %s", err.Error())
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }
		// r.Body.Close()

		// reqBuf := bytes.NewBuffer(reqBody)
		// resp, err := http.Post(address, r.Header.Get("Content-Type"), reqBuf)
		// if err != nil {
		// 	fmt.Printf("error forwarding get: %s", err.Error())
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }
		// respBody, err := ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	fmt.Printf("error reading response body: %s", err.Error())
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }
		// resp.Body.Close()

		// w.WriteHeader(resp.StatusCode)
		// w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		// w.Write(respBody)
	}
}
