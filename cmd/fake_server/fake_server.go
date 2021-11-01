package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/fake_server"
	"github.com/networknext/backend/modules/transport"

	"github.com/gorilla/mux"
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
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	ctx, cancel := context.WithCancel(context.Background())

	gcpProjectID := backend.GetGCPProjectID()

	env, err := backend.GetEnv()
	if err != nil {
		core.Error("failed to get env: %v", err)
		return 1
	}

	if gcpProjectID != "" {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
			return 1
		}
	}

	if !envvar.Exists("NEXT_CUSTOMER_PUBLIC_KEY") {
		core.Error("NEXT_CUSTOMER_PUBLIC_KEY not set")
		return 1
	}

	customerPublicKey, err := envvar.GetBase64("NEXT_CUSTOMER_PUBLIC_KEY", nil)
	if err != nil {
		core.Error("failed to parse NEXT_CUSTOMER_PUBLIC_KEY: %v", err)
		return 1
	}
	customerID := binary.LittleEndian.Uint64(customerPublicKey[:8])

	if !envvar.Exists("NEXT_CUSTOMER_PRIVATE_KEY") {
		core.Error("NEXT_CUSTOMER_PRIVATE_KEY not set")
		return 1
	}

	customerPrivateKey, err := envvar.GetBase64("NEXT_CUSTOMER_PRIVATE_KEY", nil)
	if err != nil {
		core.Error("failed to parse NEXT_CUSTOMER_PRIVATE_KEY: %v", err)
		return 1
	}

	httpPort := envvar.Get("PORT", "50001")
	if httpPort == "" {
		core.Error("PORT not set")
		return 1
	}

	// Setup an err channel to exit from goroutines
	errChan := make(chan error, 1)

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))

		go func() {
			err := http.ListenAndServe(":"+httpPort, router)
			if err != nil {
				core.Error("failed to start http server: %v", err)
				errChan <- err
			}
		}()
	}

	lc := net.ListenConfig{}

	readBuffer, err := envvar.GetInt("READ_BUFFER", 100000)
	if err != nil {
		core.Error("failed to parse READ_BUFFER: %v", err)
		return 1
	}

	writeBuffer, err := envvar.GetInt("WRITE_BUFFER", 100000)
	if err != nil {
		core.Error("failed to parse WRITE_BUFFER: %v", err)
		return 1
	}

	var serverBackendAddress *net.UDPAddr
	if !envvar.Exists("SERVER_BACKEND_ADDRESS") {
		serverBackendAddress, err = net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		if err != nil {
			core.Error("failed to resolve default server backend udp address 127.0.0.1:40000: %v", err)
			return 1
		}
	} else {
		serverBackendAddress, err = envvar.GetAddress("SERVER_BACKEND_ADDRESS", serverBackendAddress)
		if err != nil {
			core.Error("failed to get SERVER_BACKEND_ADDRESS: %v", err)
			return 1
		}
	}

	sendBeaconPackets, err := envvar.GetBool("SEND_NEXT_BEACON_PACKETS", false)
	if err != nil {
		core.Error("failed to parse SEND_NEXT_BEACON_PACKETS: %v", err)
		return 1
	}

	var beaconAddress *net.UDPAddr
	if !envvar.Exists("NEXT_BEACON_ADDRESS") {
		beaconAddress, err = net.ResolveUDPAddr("udp", "127.0.0.1:35000")
		if err != nil {
			core.Error("failed to resolve default beacon udp address 127.0.0.1:35000: %v", err)
			return 1
		}
	} else {
		beaconAddress, err = envvar.GetAddress("NEXT_BEACON_ADDRESS", beaconAddress)
		if err != nil {
			core.Error("failed to get NEXT_BEACON_ADDRESS: %v", err)
			return 1
		}
	}

	numClients, err := envvar.GetInt("NUM_CLIENTS", 400)
	if err != nil {
		core.Error("failed to parse NUM_CLIENTS: %v", err)
		return 1
	}

	maxClientsPerServer, err := envvar.GetInt("MAX_CLIENTS_PER_SERVER", 200)
	if err != nil {
		core.Error("failed to parse MAX_CLIENTS_PER_SERVER: %v", err)
		return 1
	}

	numServers := int(math.Ceil(float64(numClients) / float64(maxClientsPerServer)))

	core.Debug("starting %s with %d servers and %d clients per server", serviceName, numServers, numClients)

	// Seed the random number generator so we get different session data each run
	rand.Seed(time.Now().Unix())

	var numClientsMutex sync.Mutex

	var wg sync.WaitGroup

	for i := 0; i < numServers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

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
				core.Error("failed to listen for udp packets: %v", err)
				errChan <- err
				return
			}

			conn := lp.(*net.UDPConn)
			defer conn.Close()

			if err := conn.SetReadBuffer(readBuffer); err != nil {
				core.Error("could not set connection read buffer size: %v", err)
				errChan <- err
				return
			}

			if err := conn.SetWriteBuffer(writeBuffer); err != nil {
				core.Error("could not set connection write buffer size: %v", err)
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

			server, err := fake_server.NewFakeServer(conn, serverBackendAddress, beaconAddress, clients, transport.SDKVersionLatest, customerID, customerPrivateKey, dcName, sendBeaconPackets)
			if err != nil {
				core.Error("failed to start fake server: %v", err)
				errChan <- err
				return
			}

			if err := server.StartLoop(ctx, time.Second*10, readBuffer, writeBuffer); err != nil {
				core.Error("error during fake server operation: %v", err)
				errChan <- err
				return
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan: // Exit with an error code of 0 if we receive SIGINT or SIGTERM
		fmt.Println("Received shutdown signal.")

		cancel()
		wg.Wait()

		fmt.Println("Successfully shutdown.")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		cancel()
		wg.Wait()
		return 1
	}
}
