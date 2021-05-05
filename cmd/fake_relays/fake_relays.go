package main

import (
	"context"
	"encoding/base64"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	errChan := make(chan error, 1)

	port := envvar.Get("PORT", "30007")
	fmt.Printf("starting http server on :%s\n", port)

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
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
