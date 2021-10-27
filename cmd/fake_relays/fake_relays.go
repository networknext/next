package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/fake_relays"
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
	serviceName := "fake_relays"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	// Setup the service
	ctx, cancel := context.WithCancel(context.Background())

	gcpProjectID := backend.GetGCPProjectID()

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		core.Error("failed to get logger: %v", err)
		return 1
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("failed to get metrics handler: %v", err)
		return 1
	}

	fakeRelayMetrics, err := metrics.NewFakeRelayMetrics(ctx, metricsHandler, serviceName, "fake_relays", "Fake Relays", "")
	if err != nil {
		core.Error("failed to create fake relays metrics: %v", err)
		return 1
	}

	// Get the public key for the fake relays
	relayPublicKeyStr := envvar.Get("RELAY_PUBLIC_KEY", "8hUCRvzKh2aknL9RErM/Vj22+FGJW0tWMRz5KlHKryE=")
	relayPublicKey, err := base64.StdEncoding.DecodeString(relayPublicKeyStr)
	if err != nil {
		core.Error("failed to decode RELAY_PUBLIC_KEY %s to base64: %v", relayPublicKeyStr, err)
		return 1
	}

	// Get the number of fake relays to produce
	numRelays, err := envvar.GetInt("NUM_FAKE_RELAYS", 10)
	if err != nil {
		core.Error("failed to parse NUM_FAKE_RELAYS: %v", err)
		return 1
	}

	// Get the Relay Gateway's Load Balancer's IP
	gatewayAddr := envvar.Get("GATEWAY_LOAD_BALANCER_IP", "127.0.0.1:30000")
	// Verify the IP is valid if not testing locally
	if gcpProjectID != "" {
		ip := net.ParseIP(gatewayAddr)
		if ip == nil {
			core.Error("failed to parse the relay gateway's load balancer's IP: %s", gatewayAddr)
			return 1
		}
	}

	// Get the relay update version
	relayUpdateVersion, err := envvar.GetInt("RELAY_UPDATE_VERSION", 4)
	if err != nil {
		core.Error("failed to parse RELAY_UPDATE_VERSION: %v", err)
		return 1
	}

	// Create all the fake relays
	relays, err := fake_relays.NewFakeRelays(numRelays, relayPublicKey, gatewayAddr, relayUpdateVersion, fakeRelayMetrics)
	if err != nil {
		core.Error("failed to create fake relays: %v", err)
		return 1
	}

	// Let the relays send their updates
	var wg sync.WaitGroup
	for _, relay := range relays {
		wg.Add(1)
		go relay.StartLoop(ctx, &wg)
	}

	// Setup the status handler info
	statusData := &metrics.FakeRelayStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				fakeRelayMetrics.FakeRelayServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				fakeRelayMetrics.FakeRelayServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.FakeRelayStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				// Service Metrics
				newStatusData.Goroutines = int(fakeRelayMetrics.FakeRelayServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = fakeRelayMetrics.FakeRelayServiceMetrics.MemoryAllocated.Value()

				// Invocations
				newStatusData.UpdateInvocations = int(fakeRelayMetrics.UpdateInvocations.Value())
				newStatusData.SuccessfulUpdateInvocations = int(fakeRelayMetrics.SuccessfulUpdateInvocations.Value())

				// Error Metrics
				newStatusData.MarshalBinaryError = int(fakeRelayMetrics.ErrorMetrics.MarshalBinaryError.Value())
				newStatusData.UnmarshalBinaryError = int(fakeRelayMetrics.ErrorMetrics.UnmarshalBinaryError.Value())
				newStatusData.NotOKResponseError = int(fakeRelayMetrics.ErrorMetrics.NotOKResponseError.Value())
				newStatusData.ResolveUDPAddressError = int(fakeRelayMetrics.ErrorMetrics.ResolveUDPAddressError.Value())

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

	errChan := make(chan error, 1)

	port := envvar.Get("PORT", "30007")
	if port == "" {
		core.Error("PORT not set")
		return 1
	}
	fmt.Printf("starting http server on :%s\n", port)

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.Handle("/debug/vars", expvar.Handler())

		go func() {

			err := http.ListenAndServe(":"+port, router)
			if err != nil {
				core.Error("failed to start http server: %v", err)
				errChan <- err
				return
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		fmt.Println("Received shutdown signal.")

		cancel()
		// Wait for essential goroutines to finish up
		wg.Wait()

		fmt.Println("Successfully shutdown.")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		return 1
	}
}
