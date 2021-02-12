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

const (
	// MaxClientsPerServer is the maximum number of clients (sessions) that a server can hold
	MaxClientsPerServer = 200
)

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "server_backend"
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
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, false, []string{}))

		go func() {
			httpPort := envvar.Get("HTTP_PORT", "50001")

			err := http.ListenAndServe(":"+httpPort, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				return
			}
		}()
	}

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

	numClients, err := envvar.GetInt("NUM_CLIENTS", 400)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	numServers := int(math.Ceil(float64(numClients) / float64(MaxClientsPerServer)))

	level.Info(logger).Log("server_count", numServers, "client_count", numClients, "msg", "starting fake server")

	// Seed the random number generator so we get different session data each run
	rand.Seed(time.Now().Unix())

	var numClientsMutex sync.Mutex

	errChan := make(chan error, 1)
	for i := 0; i < numServers; i++ {
		go func() {
			bindAddress, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
			if err != nil {
				level.Error(logger).Log("err", err)
				errChan <- err
				return
			}

			numClientsMutex.Lock()
			clients := numClients

			if numClients > MaxClientsPerServer {
				numClients -= MaxClientsPerServer
				clients = MaxClientsPerServer
			} else {
				numClients = 0
			}
			numClientsMutex.Unlock()

			server, err := fake_server.NewFakeServer(clients, transport.SDKVersion{4, 0, 6}, bindAddress, serverBackendAddress, logger, customerID, customerPrivateKey)
			if err != nil {
				level.Error(logger).Log("err", err)
				errChan <- err
				return
			}

			if err := server.StartLoop(readBuffer, writeBuffer); err != nil {
				level.Error(logger).Log("err", err)
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
