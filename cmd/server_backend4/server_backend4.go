/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"sync"
	"syscall"
	"time"

	"os"
	"os/signal"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/envvar"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
	"golang.org/x/sys/unix"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
	googlepubsub "cloud.google.com/go/pubsub"
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

// A mock locator used in staging to set each session to a random, unique lat/long
type stagingLocator struct {
	SessionID uint64
}

func (locator *stagingLocator) LocateIP(ip net.IP) (routing.Location, error) {
	// Generate a random lat/long from the session ID
	sessionIDBytes := [8]byte{}
	binary.LittleEndian.PutUint64(sessionIDBytes[0:8], locator.SessionID)

	// Randomize the location by using 4 bits of the sessionID for the lat, and the other 4 for the long
	latBits := binary.LittleEndian.Uint32(sessionIDBytes[0:4])
	longBits := binary.LittleEndian.Uint32(sessionIDBytes[4:8])

	lat := (float64(latBits)) / 0xFFFFFFFF
	long := (float64(longBits)) / 0xFFFFFFFF

	return routing.Location{
		Latitude:  (-90.0 + lat*180.0) * 0.5,
		Longitude: -180.0 + long*360.0,
	}, nil
}

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

	maxNearRelays, err := envvar.GetInt("MAX_NEAR_RELAYS", 32)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
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
					return 1
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
					return 1
				}
			}
		}
	}

	// Create metrics
	backendMetrics, err := metrics.NewServerBackend4Metrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create server_backend4 metrics", "err", err)
	}

	// Create maxmindb sync metrics
	maxmindSyncMetrics, err := metrics.NewMaxmindSyncMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create session metrics", "err", err)
	}

	// Create a goroutine to update metrics
	go func() {
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		for {
			backendMetrics.ServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
			backendMetrics.ServiceMetrics.MemoryAllocated.Set(memoryUsed())

			time.Sleep(time.Second * 10)
		}
	}()

	// var db storage.Storer
	storer, err := storage.NewStorage(ctx, logger)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
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
		customerID := binary.LittleEndian.Uint64(customerPublicKey[:8])

		if !envvar.Exists("RELAY_PUBLIC_KEY") {
			level.Error(logger).Log("err", "RELAY_PUBLIC_KEY not set")
			return 1
		}

		relayPublicKey, err := envvar.GetBase64("RELAY_PUBLIC_KEY", nil)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		// Create dummy buyer and datacenter for local testing
		if err = storage.SeedStorage(logger, ctx, storer, relayPublicKey, customerID, customerPublicKey); err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}
	}

	// Create datacenter tracker
	datacenterTracker := transport.NewDatacenterTracker()

	go func() {
		for {
			unknownDatacenters := datacenterTracker.GetUnknownDatacenters()
			emptyDatacenters := datacenterTracker.GetEmptyDatacenters()

			for _, datacenter := range unknownDatacenters {
				level.Warn(logger).Log("msg", "unknown datacenter", "datacenter", datacenter)
			}

			for _, datacenter := range emptyDatacenters {
				level.Warn(logger).Log("msg", "empty datacenter", "datacenter", datacenter)
			}

			time.Sleep(10 * time.Second)
		}
	}()

	if !envvar.Exists("SERVER_BACKEND_PRIVATE_KEY") {
		level.Error(logger).Log("err", "SERVER_BACKEND_PRIVATE_KEY not set")
		return 1
	}

	privateKey, err := envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", nil)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	routerPrivateKeySlice, err := envvar.GetBase64("RELAY_ROUTER_PRIVATE_KEY", nil)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	routerPrivateKey := [crypto.KeySize]byte{}
	copy(routerPrivateKey[:], routerPrivateKeySlice)

	getIPLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return routing.NullIsland
	}

	// Open the Maxmind DB and create a routing.MaxmindDB from it
	maxmindCityURI := envvar.Get("MAXMIND_CITY_DB_URI", "")
	maxmindISPURI := envvar.Get("MAXMIND_ISP_DB_URI", "")
	if maxmindCityURI != "" && maxmindISPURI != "" {
		mmdb := &routing.MaxmindDB{
			HTTPClient: http.DefaultClient,
			CityURI:    maxmindCityURI,
			IspURI:     maxmindISPURI,
		}
		var mmdbMutex sync.RWMutex

		getIPLocatorFunc = func(sessionID uint64) routing.IPLocator {
			mmdbMutex.RLock()
			defer mmdbMutex.RUnlock()

			mmdbRet := mmdb
			return mmdbRet
		}

		if err := mmdb.Sync(ctx, maxmindSyncMetrics); err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		// todo: disable the sync for now until we can find out why it's causing session drops

		// if envvar.Exists("MAXMIND_SYNC_DB_INTERVAL") {
		// 	syncInterval, err := envvar.GetDuration("MAXMIND_SYNC_DB_INTERVAL", time.Hour*24)
		// 	if err != nil {
		// 		level.Error(logger).Log("err", err)
		// 		return 1
		// 	}

		// 	// Start a goroutine to sync from Maxmind.com
		// 	go func() {
		// 		ticker := time.NewTicker(syncInterval)
		// 		for {
		// 			newMMDB := &routing.MaxmindDB{}

		// 			select {
		// 			case <-ticker.C:
		// 				if err := newMMDB.Sync(ctx, maxmindSyncMetrics); err != nil {
		// 					level.Error(logger).Log("err", err)
		// 					continue
		// 				}

		// 				// Pointer swap the mmdb so we can sync from Maxmind.com lock free
		// 				mmdbMutex.Lock()
		// 				mmdb = newMMDB
		// 				mmdbMutex.Unlock()
		// 			case <-ctx.Done():
		// 				return
		// 			}

		// 			time.Sleep(syncInterval)
		// 		}
		// 	}()
		// }
	}

	// Use a custom IP locator for staging so that clients
	// have different, random lat/longs
	if env == "staging" {
		getIPLocatorFunc = func(sessionID uint64) routing.IPLocator {
			return &stagingLocator{
				SessionID: sessionID,
			}
		}
	}

	routeMatrix4 := &routing.RouteMatrix4{}
	var routeMatrix4Mutex sync.RWMutex

	getRouteMatrix4Func := func() *routing.RouteMatrix4 {
		routeMatrix4Mutex.RLock()
		rm4 := routeMatrix4
		routeMatrix4Mutex.RUnlock()
		return rm4
	}

	// Sync route matrix
	{
		if envvar.Exists("ROUTE_MATRIX_URI") {
			uri := envvar.Get("ROUTE_MATRIX_URI", "")
			syncInterval, err := envvar.GetDuration("ROUTE_MATRIX_SYNC_INTERVAL", time.Second)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			go func() {
				httpClient := &http.Client{
					Timeout: time.Second * 2,
				}
				for {
					var routeEntriesReader io.ReadCloser

					// Default to reading route matrix from file
					if f, err := os.Open(uri); err == nil {
						routeEntriesReader = f
					}

					// Prefer to get it remotely if possible
					if r, err := httpClient.Get(uri); err == nil {
						routeEntriesReader = r.Body
					}

					start := time.Now()

					if routeEntriesReader == nil {
						time.Sleep(syncInterval)
						continue
					}

					buffer, err := ioutil.ReadAll(routeEntriesReader)

					if routeEntriesReader != nil {
						routeEntriesReader.Close()
					}

					if err != nil {
						level.Error(logger).Log("envvar", "ROUTE_MATRIX_URI", "value", uri, "msg", "could not read route matrix", "err", err)
						time.Sleep(syncInterval)
						continue // Don't swap route matrix if we fail to read
					}

					var newRouteMatrix4 routing.RouteMatrix4
					rs := encoding.CreateReadStream(buffer)
					if err := newRouteMatrix4.Serialize(rs); err != nil {
						level.Error(logger).Log("msg", "could not serialize route matrix", "err", err)
						time.Sleep(syncInterval)
						continue // Don't swap route matrix if we fail to serialize
					}

					routeEntriesTime := time.Since(start)

					duration := float64(routeEntriesTime.Milliseconds())
					backendMetrics.RouteMatrixUpdateDuration.Set(duration)

					if duration > 100 {
						backendMetrics.RouteMatrixUpdateLongDuration.Add(1)
					}

					numRoutes := int32(0)
					for i := range newRouteMatrix4.RouteEntries {
						numRoutes += newRouteMatrix4.RouteEntries[i].NumRoutes
					}
					backendMetrics.RouteMatrixNumRoutes.Set(float64(numRoutes))
					backendMetrics.RouteMatrixBytes.Set(float64(len(buffer)))

					routeMatrix4Mutex.Lock()
					routeMatrix4 = &newRouteMatrix4
					routeMatrix4Mutex.Unlock()

					time.Sleep(syncInterval)
				}
			}()
		}
	}

	// Create a local biller
	var biller billing.Biller = &billing.LocalBiller{
		Logger:  logger,
		Metrics: backendMetrics.BillingMetrics,
	}

	pubsubEmulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
	if gcpOK || pubsubEmulatorOK {

		pubsubCtx := ctx
		if pubsubEmulatorOK {
			gcpProjectID = "local"

			var cancelFunc context.CancelFunc
			pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			level.Info(logger).Log("msg", "Detected pubsub emulator")
		}

		// Google Pubsub
		{
			clientCount, err := envvar.GetInt("BILLING_CLIENT_COUNT", 1)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			countThreshold, err := envvar.GetInt("BILLING_BATCHED_MESSAGE_COUNT", 100)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			byteThreshold, err := envvar.GetInt("BILLING_BATCHED_MESSAGE_MIN_BYTES", 1024)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			// We do our own batching so don't stack the library's batching on top of ours
			// Specifically, don't stack the message count thresholds
			settings := googlepubsub.DefaultPublishSettings
			settings.CountThreshold = 1
			settings.ByteThreshold = byteThreshold
			settings.NumGoroutines = runtime.GOMAXPROCS(0)

			pubsub, err := billing.NewGooglePubSubBiller(pubsubCtx, backendMetrics.BillingMetrics, logger, gcpProjectID, "billing", clientCount, countThreshold, byteThreshold, &settings)
			if err != nil {
				level.Error(logger).Log("msg", "could not create pubsub biller", "err", err)
				return 1
			}

			biller = pubsub
		}
	}

	// Start portal cruncher publisher
	var portalPublisher pubsub.Publisher
	{
		portalCruncherHost := envvar.Get("PORTAL_CRUNCHER_HOST", "tcp://127.0.0.1:5555")

		postSessionPortalSendBufferSize, err := envvar.GetInt("POST_SESSION_PORTAL_SEND_BUFFER_SIZE", 1000000)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		portalCruncherPublisher, err := pubsub.NewPortalCruncherPublisher(portalCruncherHost, postSessionPortalSendBufferSize)
		if err != nil {
			level.Error(logger).Log("msg", "could not create portal cruncher publisher", "err", err)
			return 1
		}

		portalPublisher = portalCruncherPublisher
	}

	numPostSessionGoroutines, err := envvar.GetInt("POST_SESSION_THREAD_COUNT", 1000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	postSessionBufferSize, err := envvar.GetInt("POST_SESSION_BUFFER_SIZE", 1000000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	postSessionPortalMaxRetries, err := envvar.GetInt("POST_SESSION_PORTAL_MAX_RETRIES", 10)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Create a post session handler to handle the post process of session updates.
	// This way, we can quickly return from the session update handler and not spawn a
	// ton of goroutines if things get backed up.
	postSessionHandler := transport.NewPostSessionHandler(numPostSessionGoroutines, postSessionBufferSize, portalPublisher, postSessionPortalMaxRetries, biller, logger, backendMetrics.PostSessionMetrics)
	postSessionHandler.StartProcessing(ctx)

	// Create the multipath veto handler to handle syncing multipath vetoes to and from redis
	redisMultipathVetoHost := envvar.Get("REDIS_HOST_MULTIPATH_VETO", "127.0.0.1:6379")
	multipathVetoSyncFrequency, err := envvar.GetDuration("MULTIPATH_VETO_SYNC_FREQUENCY", time.Second*10)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisMultipathVetoHost, storer)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	if err := multipathVetoHandler.Sync(); err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Start a routine to sync multipath vetoed users from redis to this instance
	{
		ticker := time.NewTicker(multipathVetoSyncFrequency)
		go func(ctx context.Context) {
			for {
				select {
				case <-ticker.C:
					if err := multipathVetoHandler.Sync(); err != nil {
						level.Error(logger).Log("err", err)
					}
					break
				case <-ctx.Done():
					break
				}
			}
		}(ctx)
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

	serverInitHandler := transport.ServerInitHandlerFunc4(log.With(logger, "handler", "server_init"), storer, datacenterTracker, backendMetrics.ServerInitMetrics)
	serverUpdateHandler := transport.ServerUpdateHandlerFunc4(log.With(logger, "handler", "server_update"), storer, datacenterTracker, backendMetrics.ServerUpdateMetrics)
	sessionUpdateHandler := transport.SessionUpdateHandlerFunc4(log.With(logger, "handler", "session_update"), getIPLocatorFunc, getRouteMatrix4Func, multipathVetoHandler, storer, maxNearRelays, routerPrivateKey, postSessionHandler, backendMetrics.SessionUpdateMetrics)

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

				packetType := data[0]
				data = data[crypto.PacketHashSize+1 : size]

				var buffer bytes.Buffer
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
					crypto.HashPacket(crypto.PacketHashKey, response)

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
