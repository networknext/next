package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/fake_server"
	"github.com/networknext/backend/modules/transport"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "fake_server"
	fmt.Printf("fake_server: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	gcpProjectID := backend.GetGCPProjectID()

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
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

	if !envvar.Exists("NEXT_CUSTOMER_PUBLIC_KEY") {
		level.Error(logger).Log("err", errors.New("NEXT_CUSTOMER_PUBLIC_KEY not set"))
		return 1
	}

	customerPublicKey, err := envvar.GetBase64("NEXT_CUSTOMER_PUBLIC_KEY", nil)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}
	customerID := binary.LittleEndian.Uint64(customerPublicKey[:8])

	if !envvar.Exists("NEXT_CUSTOMER_PRIVATE_KEY") {
		level.Error(logger).Log("err", errors.New("NEXT_CUSTOMER_PRIVATE_KEY not set"))
		return 1
	}

	customerPrivateKey, err := envvar.GetBase64("NEXT_CUSTOMER_PRIVATE_KEY", nil)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))

		go func() {
			httpPort := envvar.Get("PORT", "50001")

			err := http.ListenAndServe(":"+httpPort, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				return
			}
		}()
	}

	lc := net.ListenConfig{}

	readBuffer, err := envvar.GetInt("READ_BUFFER", 100000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	writeBuffer, err := envvar.GetInt("WRITE_BUFFER", 100000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	serverBackendAddress, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	serverBackendAddress, err = envvar.GetAddress("SERVER_BACKEND_ADDRESS", serverBackendAddress)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	beaconAddress, err := net.ResolveUDPAddr("udp", "127.0.0.1:35000")
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	beaconAddress, err = envvar.GetAddress("NEXT_BEACON_ADDRESS", beaconAddress)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	numClients, err := envvar.GetInt("NUM_CLIENTS", 400)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	maxClientsPerServer, err := envvar.GetInt("MAX_CLIENTS_PER_SERVER", 200)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	numServers := int(math.Ceil(float64(numClients) / float64(maxClientsPerServer)))

	level.Info(logger).Log("server_count", numServers, "client_count", numClients, "msg", "starting fake server")

	// Seed the random number generator so we get different session data each run
	rand.Seed(time.Now().Unix())

	var numClientsMutex sync.Mutex

	errChan := make(chan error, 1)
	for i := 0; i < numServers; i++ {
		go func() {
			numClientsMutex.Lock()
			clients := numClients

			if numClients > maxClientsPerServer {
				numClients -= maxClientsPerServer
				clients = maxClientsPerServer
			} else {
				numClients = 0
			}
			numClientsMutex.Unlock()

			lp, err := lc.ListenPacket(ctx, "udp", "0.0.0.0:0")
			if err != nil {
				level.Error(logger).Log("err", err)
				errChan <- err
				return
			}

			conn := lp.(*net.UDPConn)
			defer conn.Close()

			if err := conn.SetReadBuffer(readBuffer); err != nil {
				level.Error(logger).Log("msg", "could not set connection read buffer size", "err", err)
				errChan <- err
				return
			}

			if err := conn.SetWriteBuffer(writeBuffer); err != nil {
				level.Error(logger).Log("msg", "could not set connection write buffer size", "err", err)
				errChan <- err
				return
			}

			// Assign a datacenter for this server
			var dcName string
			if gcpProjectID != "" {
				// Staging datacenters are between 1 and 80, inclusive
				dcNum := rand.Intn(80-1) + 1
				dcName = fmt.Sprintf("staging.%d", dcNum)
			} else {
				dcName = "local"
			}

			server, err := fake_server.NewFakeServer(conn, serverBackendAddress, beaconAddress, clients, transport.SDKVersionLatest, logger, customerID, customerPrivateKey, dcName)
			if err != nil {
				level.Error(logger).Log("err", err)
				errChan <- err
				return
			}

			if err := server.StartLoop(ctx, time.Second*10, readBuffer, writeBuffer); err != nil {
				level.Error(logger).Log("err", err)
				fmt.Printf("Error: %v\n", err)
				errChan <- err
				return
			}
		}()
	}

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-sigint:
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		return 1
	}
}
