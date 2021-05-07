package main

import (
	"bytes"
	"context"
	"encoding/base64"
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
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/fake_relays"
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
	serviceName := "fake_relays"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	// Setup the service
	ctx, cancel := context.WithCancel(context.Background())

	gcpProjectID := backend.GetGCPProjectID()

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	fakeRelayMetrics, err := metrics.NewFakeRelayMetrics(ctx, metricsHandler, serviceName, "fake_relays", "Fake Relays", "")
	if err != nil {
		level.Error(logger).Log("msg", "failed to create fake relays metrics", "err", err)
		return 1
	}

	// Get the public key for the fake relays
	relayPublicKeyStr := envvar.Get("RELAY_PUBLIC_KEY", "8hUCRvzKh2aknL9RErM/Vj22+FGJW0tWMRz5KlHKryE=")
	relayPublicKey, err := base64.StdEncoding.DecodeString(relayPublicKeyStr)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("could not decode to base64: %s", relayPublicKeyStr), "err", err)
		return 1
	}

	// Get the number of fake relays to produce
	numRelays, err := envvar.GetInt("NUM_FAKE_RELAYS", 10)
	if err != nil {
		level.Error(logger).Log("msg", "error reading NUM_FAKE_RELAYS as int", "err", err)
	}

	// Get the Relay Gateway's Load Balancer's IP
	gatewayAddr := envvar.Get("GATEWAY_LOAD_BALANCER_IP", "127.0.0.1:30000")
	// Verify the IP is valid if not testing locally
	if gcpProjectID != "" {
		ip := net.ParseIP(gatewayAddr)
		if ip == nil {
			level.Error(logger).Log("msg", fmt.Sprintf("could not parse relay gatway's load balancer's IP: %s", gatewayAddr), "err", err)
			return 1
		}
	}

	// Get the relay update version
	relayUpdateVersion, err := envvar.GetInt("RELAY_UPDATE_VERSION", 3)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Create all the fake relays
	relays, err := fake_relays.NewFakeRelays(numRelays, relayPublicKey, gatewayAddr, relayUpdateVersion, logger, fakeRelayMetrics)
	if err != nil {
		level.Error(logger).Log("msg", "could not create fake relays", "err", err)
		return 1
	}

	// Let the relays send their updates
	var wg sync.WaitGroup
	for _, relay := range relays {
		wg.Add(1)
		go relay.StartLoop(ctx, &wg)
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
			for {
				fakeRelayMetrics.FakeRelayServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				fakeRelayMetrics.FakeRelayServiceMetrics.MemoryAllocated.Set(memoryUsed())

				statusDataString := fmt.Sprintf("%s\n", serviceName)
				statusDataString += fmt.Sprintf("git hash %s\n", sha)
				statusDataString += fmt.Sprintf("started %s\n", startTime.Format("Mon, 02 Jan 2006 15:04:05 EST"))
				statusDataString += fmt.Sprintf("uptime %s\n", time.Since(startTime))

				statusDataString += fmt.Sprintf("%d goroutines\n", int(fakeRelayMetrics.FakeRelayServiceMetrics.Goroutines.Value()))
				statusDataString += fmt.Sprintf("%.2f mb allocated\n", fakeRelayMetrics.FakeRelayServiceMetrics.MemoryAllocated.Value())
				statusDataString += fmt.Sprintf("%d update invocations\n", int(fakeRelayMetrics.UpdateInvocations.Value()))
				statusDataString += fmt.Sprintf("%d successful updates\n", int(fakeRelayMetrics.SuccessfulUpdateInvocations.Value()))
				statusDataString += fmt.Sprintf("%d marshal binary errors\n", int(fakeRelayMetrics.ErrorMetrics.MarshalBinaryError.Value()))
				statusDataString += fmt.Sprintf("%d unmarshal binary errors\n", int(fakeRelayMetrics.ErrorMetrics.UnmarshalBinaryError.Value()))
				statusDataString += fmt.Sprintf("%d update post errors\n", int(fakeRelayMetrics.ErrorMetrics.UpdatePostError.Value()))
				statusDataString += fmt.Sprintf("%d not OK response errors\n", int(fakeRelayMetrics.ErrorMetrics.NotOKResponseError.Value()))
				statusDataString += fmt.Sprintf("%d resolve UDP address errors\n", int(fakeRelayMetrics.ErrorMetrics.ResolveUDPAddressError.Value()))

				statusMutex.Lock()
				statusData = []byte(statusDataString)
				statusMutex.Unlock()

				time.Sleep(time.Second * 10)
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

	errChan := make(chan error, 1)

	port := envvar.Get("PORT", "30007")
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
				level.Error(logger).Log("err", err)
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
		level.Debug(logger).Log("msg", "Received shutdown signal")
		fmt.Println("Received shutdown signal.")

		cancel()
		// Wait for essential goroutines to finish up
		wg.Wait()

		level.Debug(logger).Log("msg", "Successfully shutdown")
		fmt.Println("Successfully shutdown.")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		return 1
	}
}
