/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"syscall"
	"time"

	"os"
	"os/signal"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/envvar"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"golang.org/x/sys/unix"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
)

// MaxRelayCount is the maximum number of relays you can run locally with the firestore emulator
// An equal number of valve relays will also be added
const MaxRelayCount = 10

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
	fmt.Printf("server_backend4: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	// Configure local logging
	logger := log.NewLogfmtLogger(os.Stdout)

	gcpOK := envvar.Exists("GOOGLE_PROJECT_ID")
	gcpProjectID := envvar.Get("GOOGLE_PROJECT_ID", "")

	// StackDriver Logging
	{
		enableSDLogging, err := envvar.GetBool("ENABLE_STACKDRIVER_LOGGING", false)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		if enableSDLogging && gcpOK {
			loggingClient, err := gcplogging.NewClient(ctx, gcpProjectID)
			if err != nil {
				level.Error(logger).Log("msg", "failed to create GCP logging client", "err", err)
				return 1
			}

			logger = logging.NewStackdriverLogger(loggingClient, "server-backend4")
		}
	}
	{
		backendLogLevel := envvar.Get("BACKEND_LOG_LEVEL", "none")
		switch backendLogLevel {
		case "none":
			logger = level.NewFilter(logger, level.AllowNone())
		case level.ErrorValue().String():
			logger = level.NewFilter(logger, level.AllowError())
		case level.WarnValue().String():
			logger = level.NewFilter(logger, level.AllowWarn())
		case level.InfoValue().String():
			logger = level.NewFilter(logger, level.AllowInfo())
		case level.DebugValue().String():
			logger = level.NewFilter(logger, level.AllowDebug())
		default:
			logger = level.NewFilter(logger, level.AllowWarn())
		}

		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	// Get env
	if !envvar.Exists("ENV") {
		level.Error(logger).Log("err", "ENV not set")
		return 1
	}
	env := envvar.Get("ENV", "")

	// Create an in-memory storer
	var storer storage.Storer = &storage.InMemory{
		LocalMode: true,
	}

	// Create dummy entries in storer for local testing
	if env == "local" {
		if !envvar.Exists("NEXT_CUSTOMER_PUBLIC_KEY") {
			level.Error(logger).Log("err", "NEXT_CUSTOMER_PUBLIC_KEY not set")
			return 1
		}

		customerPublicKey, err := envvar.GetBase64("NEXT_CUSTOMER_PUBLIC_KEY", nil)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		if !envvar.Exists("RELAY_PUBLIC_KEY") {
			level.Error(logger).Log("err", "RELAY_PUBLIC_KEY not set")
			return 1
		}

		relayPublicKey, err := envvar.GetBase64("RELAY_PUBLIC_KEY", nil)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		if err := storer.AddBuyer(ctx, routing.Buyer{
			ID:                   13672574147039585173,
			Name:                 "local",
			Live:                 true,
			PublicKey:            customerPublicKey,
			RoutingRulesSettings: routing.LocalRoutingRulesSettings,
		}); err != nil {
			level.Error(logger).Log("msg", "could not add buyer to storage", "err", err)
			os.Exit(1)
		}
		seller := routing.Seller{
			ID:                        "sellerID",
			Name:                      "local",
			IngressPriceNibblinsPerGB: 0.1 * 1e9,
			EgressPriceNibblinsPerGB:  0.5 * 1e9,
		}

		valveSeller := routing.Seller{
			ID:                        "valve",
			Name:                      "valve",
			IngressPriceNibblinsPerGB: 0.1 * 1e9,
			EgressPriceNibblinsPerGB:  0.5 * 1e9,
		}

		datacenter := routing.Datacenter{
			ID:       crypto.HashID("local"),
			Name:     "local",
			Enabled:  true,
			Location: routing.LocationNullIsland,
		}

		if err := storer.AddSeller(ctx, seller); err != nil {
			level.Error(logger).Log("msg", "could not add seller to storage", "err", err)
			os.Exit(1)
		}

		if err := storer.AddSeller(ctx, valveSeller); err != nil {
			level.Error(logger).Log("msg", "could not add valve seller to storage", "err", err)
			os.Exit(1)
		}

		if err := storer.AddDatacenter(ctx, datacenter); err != nil {
			level.Error(logger).Log("msg", "could not add datacenter to storage", "err", err)
			os.Exit(1)
		}

		for i := int64(0); i < MaxRelayCount; i++ {
			addressString := "127.0.0.1:1000" + strconv.FormatInt(i, 10)
			addr, err := net.ResolveUDPAddr("udp", addressString)
			if err != nil {
				level.Error(logger).Log("msg", "could parse udp address", "address", addressString, "err", err)
				os.Exit(1)
			}

			if err := storer.AddRelay(ctx, routing.Relay{
				ID:          crypto.HashID(addr.String()),
				Name:        addr.String(),
				Addr:        *addr,
				PublicKey:   relayPublicKey,
				Seller:      seller,
				Datacenter:  datacenter,
				MaxSessions: 3000,
			}); err != nil {
				level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
				os.Exit(1)
			}
		}

		for i := int64(0); i < MaxRelayCount; i++ {
			addressString := "127.0.0.1:1001" + strconv.FormatInt(i, 10)
			addr, err := net.ResolveUDPAddr("udp", addressString)
			if err != nil {
				level.Error(logger).Log("msg", "could parse udp address", "address", addressString, "err", err)
				os.Exit(1)
			}

			if err := storer.AddRelay(ctx, routing.Relay{
				ID:          crypto.HashID(addr.String()),
				Name:        addr.String(),
				Addr:        *addr,
				PublicKey:   relayPublicKey,
				Seller:      valveSeller,
				Datacenter:  datacenter,
				MaxSessions: 3000,
			}); err != nil {
				level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
				os.Exit(1)
			}
		}
	}

	// Check for the firestore emulator
	firestoreEmulatorOK := envvar.Exists("FIRESTORE_EMULATOR_HOST")
	if firestoreEmulatorOK {
		gcpProjectID = "local"
		level.Info(logger).Log("msg", "Detected firestore emulator")
	}

	if gcpOK || firestoreEmulatorOK {
		// Firestore
		{
			// Create a Firestore Storer
			fs, err := storage.NewFirestore(ctx, gcpProjectID, logger)
			if err != nil {
				level.Error(logger).Log("msg", "could not create firestore", "err", err)
				return 1
			}

			fsSyncInterval, err := envvar.GetDuration("GOOGLE_FIRESTORE_SYNC_INTERVAL", time.Second*10)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			// Start a goroutine to sync from Firestore
			go func() {
				ticker := time.NewTicker(fsSyncInterval)
				fs.SyncLoop(ctx, ticker.C)
			}()

			// Set the Firestore Storer to give to handlers
			storer = fs
		}
	}

	// Create a local metrics handler
	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	if gcpOK {
		// StackDriver Metrics
		{
			enableSDMetrics, err := envvar.GetBool("ENABLE_STACKDRIVER_METRICS", false)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			if enableSDMetrics {
				// Set up StackDriver metrics
				sd := metrics.StackDriverHandler{
					ProjectID:          gcpProjectID,
					OverwriteFrequency: time.Second,
					OverwriteTimeout:   10 * time.Second,
				}

				if err := sd.Open(ctx); err != nil {
					level.Error(logger).Log("msg", "Failed to create StackDriver metrics client", "err", err)
					os.Exit(1)
				}

				metricsHandler = &sd

				sdWriteInterval, err := envvar.GetDuration("GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL", time.Minute)
				if err != nil {
					level.Error(logger).Log("err", err)
					return 1
				}

				go func() {
					metricsHandler.WriteLoop(ctx, logger, sdWriteInterval, 200)
				}()
			}
		}

		// StackDriver Profiler
		{
			enableSDProfiler, err := envvar.GetBool("ENABLE_STACKDRIVER_PROFILER", false)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			if enableSDProfiler {
				// Set up StackDriver profiler
				if err := profiler.Start(profiler.Config{
					Service:        "server_backend",
					ServiceVersion: env,
					ProjectID:      gcpProjectID,
					MutexProfiling: true,
				}); err != nil {
					level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
					os.Exit(1)
				}
			}
		}
	}

	// Create metrics
	serverInitMetrics, err := metrics.NewServerInitMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create server init metrics", "err", err)
		return 1
	}

	serverUpdateMetrics, err := metrics.NewServerUpdateMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create server update metrics", "err", err)
		return 1
	}

	sessionUpdateMetrics, err := metrics.NewSessionMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create session update metrics", "err", err)
		return 1
	}

	// Create datacenter tracker
	datacenterTracker := transport.NewDatacenterTracker()

	if !envvar.Exists("SERVER_BACKEND_PRIVATE_KEY") {
		level.Error(logger).Log("err", "SERVER_BACKEND_PRIVATE_KEY not set")
		return 1
	}

	privateKey, err := envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", nil)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))
		router.Handle("/debug/vars", expvar.Handler())

		go func() {
			if !envvar.Exists("HTTP_PORT") {
				level.Error(logger).Log("err", "env var HTTP_PORT must be set")
				return
			}

			httpPort := envvar.Get("HTTP_PORT", "40000")

			err := http.ListenAndServe(":"+httpPort, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				return
			}
		}()
	}

	numThreads, err := envvar.GetInt("NUM_THREADS", 1)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
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

	udpPort := envvar.Get("UDP_PORT", "40000")

	var wg sync.WaitGroup

	wg.Add(numThreads)

	lc := net.ListenConfig{
		Control: func(network string, address string, c syscall.RawConn) error {
			err := c.Control(func(fileDescriptor uintptr) {
				err := unix.SetsockoptInt(int(fileDescriptor), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if err != nil {
					panic(fmt.Sprintf("failed to set reuse address socket option: %v", err))
				}

				err = unix.SetsockoptInt(int(fileDescriptor), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if err != nil {
					panic(fmt.Sprintf("failed to set reuse port socket option: %v", err))
				}
			})

			return err
		},
	}

	connections := make([]*net.UDPConn, numThreads)

	serverInitHandler := transport.ServerInitHandlerFunc4(logger, storer, datacenterTracker, serverInitMetrics)
	serverUpdateHandler := transport.ServerUpdateHandlerFunc4(logger, storer, datacenterTracker, serverUpdateMetrics)
	sessionUpdateHandler := transport.SessionUpdateHandlerFunc4(logger, storer, sessionUpdateMetrics)

	for i := 0; i < numThreads; i++ {
		go func(thread int) {
			lp, err := lc.ListenPacket(ctx, "udp", "0.0.0.0:"+udpPort)
			if err != nil {
				panic(fmt.Sprintf("could not bind socket: %v", err))
			}

			conn := lp.(*net.UDPConn)

			if err := conn.SetReadBuffer(readBuffer); err != nil {
				panic(fmt.Sprintf("could not set connection read buffer size: %v", err))
			}

			if err := conn.SetWriteBuffer(writeBuffer); err != nil {
				panic(fmt.Sprintf("could not set connection write buffer size: %v", err))
			}

			connections[thread] = conn

			dataArray := [transport.DefaultMaxPacketSize]byte{}
			for {
				data := dataArray[:]
				size, fromAddr, err := conn.ReadFromUDP(data)
				if err != nil {
					level.Error(logger).Log("msg", "failed to read UDP packet", "err", err)
					break
				}

				if size <= 0 {
					continue
				}

				data = data[:size]

				// Check the packet hash is legit and remove the hash from the beginning of the packet
				// to continue processing the packet as normal
				if !crypto.IsNetworkNextPacket(crypto.PacketHashKey, data) {
					level.Error(logger).Log("err", "received non network next packet")
					continue
				}

				data = data[crypto.PacketHashSize:size]

				var buffer bytes.Buffer
				packetType := data[0]
				packet := transport.UDPPacket{SourceAddr: *fromAddr, Data: data}

				switch packetType {
				case transport.PacketTypeServerInitRequest4:
					serverInitHandler(&buffer, &packet)
				case transport.PacketTypeServerUpdate4:
					serverUpdateHandler(&buffer, &packet)
				case transport.PacketTypeSessionUpdate4:
					sessionUpdateHandler(&buffer, &packet)
				default:
					level.Error(logger).Log("err", "unknown packet type", "packet_type", packet.Data[0])
				}

				if buffer.Len() > 0 {
					response := buffer.Bytes()

					// Sign and hash the response
					response = crypto.SignPacket(privateKey, response)
					response = crypto.HashPacket(crypto.PacketHashKey, response)

					if _, err := conn.WriteToUDP(response, fromAddr); err != nil {
						level.Error(logger).Log("msg", "failed to write UDP response", "err", err)
					}
				}
			}

			wg.Done()
		}(i)
	}

	level.Info(logger).Log("msg", "waiting for incoming connections")

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	for _, connection := range connections {
		connection.Close()
	}

	return 0
}
