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

	"github.com/networknext/backend/modules/common/helpers"

	"os"
	"os/signal"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/config"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/pubsub"
	"golang.org/x/sys/unix"

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

	lat := (float32(latBits)) / 0xFFFFFFFF
	long := (float32(longBits)) / 0xFFFFFFFF

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
	serviceName := "server_backend"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

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

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
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

	storer, err := backend.GetStorer(ctx, logger, gcpProjectID, env)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Create server backend metrics
	backendMetrics, err := metrics.NewServerBackendMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create server_backend metrics", "err", err)
		return 1
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

	if !envvar.Exists("SERVER_BACKEND_PRIVATE_KEY") {
		level.Error(logger).Log("err", "SERVER_BACKEND_PRIVATE_KEY not set")
		return 1
	}

	privateKey, err := envvar.GetBase64("SERVER_BACKEND_PRIVATE_KEY", nil)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	if !envvar.Exists("RELAY_ROUTER_PRIVATE_KEY") {
		level.Error(logger).Log("err", "RELAY_ROUTER_PRIVATE_KEY not set")
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

		if err := mmdb.Sync(ctx, backendMetrics.MaxmindSyncMetrics); err != nil {
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

	var matrixStore storage.MatrixStore
	matrixStoreAddr := envvar.Get("MATRIX_STORE_ADDRESS", "")
	if matrixStoreAddr != "" {
		mSReadTimeout, err := envvar.GetDuration("MATRIX_STORE_READ_TIMEOUT", 250*time.Millisecond)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}
		mSWriteTimeout, err := envvar.GetDuration("MATRIX_STORE_WRITE_TIMEOUT", 250*time.Millisecond)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		matrixStore, err = storage.NewRedisMatrixStore(matrixStoreAddr, mSReadTimeout, mSWriteTimeout, 0*time.Second)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}
	}

	routeMatrix := &routing.RouteMatrix{}
	var routeMatrixMutex sync.RWMutex

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		routeMatrixMutex.RLock()
		rm4 := routeMatrix
		routeMatrixMutex.RUnlock()
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

				valveBackend, err := envvar.GetBool("VALVE_SERVER_BACKEND", false)
				if err != nil {
					level.Error(logger).Log("err", err)
				}

				syncTimer := helpers.NewSyncTimer(syncInterval)
				for {
					syncTimer.Run()

					var buffer []byte
					start := time.Now()

					backendMetrics.RouteMatrixUpdateMetrics.Invocations.Add(1)

					newRelayBackend, err := envvar.GetBool("FEATURE_NEW_RELAY_BACKEND", false)
					if err != nil {
						level.Error(logger).Log("err", err)
					}
					if newRelayBackend && matrixStore != nil {
						if valveBackend {
							buffer, err = matrixStore.GetLiveMatrix(storage.MatrixTypeValve)
							if err != nil {
								level.Error(logger).Log("err", err)
							}
						} else {
							buffer, err = matrixStore.GetLiveMatrix(storage.MatrixTypeNormal)
							if err != nil {
								level.Error(logger).Log("err", err)
							}
						}

					} else {
						var routeEntriesReader io.ReadCloser

						// Default to reading route matrix from file
						if f, err := os.Open(uri); err == nil {
							routeEntriesReader = f
						}

						// Prefer to get it remotely if possible
						if r, err := httpClient.Get(uri); err == nil {
							routeEntriesReader = r.Body
						}

						if routeEntriesReader == nil {
							continue
						}

						buffer, err = ioutil.ReadAll(routeEntriesReader)

						if routeEntriesReader != nil {
							routeEntriesReader.Close()
						}

						if err != nil {
							level.Error(logger).Log("envvar", "ROUTE_MATRIX_URI", "value", uri, "msg", "could not read route matrix", "err", err)
							continue // Don't swap route matrix if we fail to read
						}
					}
					var newRouteMatrix routing.RouteMatrix
					if len(buffer) > 0 {
						rs := encoding.CreateReadStream(buffer)
						if err := newRouteMatrix.Serialize(rs); err != nil {
							level.Error(logger).Log("msg", "could not serialize route matrix", "err", err)
							continue // Don't swap route matrix if we fail to serialize
						}
					}

					routeEntriesTime := time.Since(start)

					duration := float64(routeEntriesTime.Milliseconds())
					backendMetrics.RouteMatrixUpdateMetrics.Duration.Set(duration)

					if duration > 100 {
						backendMetrics.RouteMatrixUpdateMetrics.LongDuration.Add(1)
					}

					numRoutes := int32(0)
					for i := range newRouteMatrix.RouteEntries {
						numRoutes += newRouteMatrix.RouteEntries[i].NumRoutes
					}

					backendMetrics.RouteMatrixMetrics.RelayCount.Set(float64(len(newRouteMatrix.RelayIDs)))
					backendMetrics.RouteMatrixMetrics.DatacenterCount.Set(float64(len(newRouteMatrix.RelayDatacenterIDs)))
					backendMetrics.RouteMatrixMetrics.RouteCount.Set(float64(numRoutes))
					backendMetrics.RouteMatrixMetrics.Bytes.Set(float64(len(buffer)))

					routeMatrixMutex.Lock()
					routeMatrix = &newRouteMatrix
					routeMatrixMutex.Unlock()
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
	if gcpProjectID != "" || pubsubEmulatorOK {

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
	portalPublishers := make([]pubsub.Publisher, 0)
	{
		portalCruncherHosts := envvar.GetList("PORTAL_CRUNCHER_HOSTS", []string{"tcp://127.0.0.1:5555"})

		postSessionPortalSendBufferSize, err := envvar.GetInt("POST_SESSION_PORTAL_SEND_BUFFER_SIZE", 1000000)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		for _, host := range portalCruncherHosts {
			portalCruncherPublisher, err := pubsub.NewPortalCruncherPublisher(host, postSessionPortalSendBufferSize)
			if err != nil {
				level.Error(logger).Log("msg", "could not create portal cruncher publisher", "err", err)
				return 1
			}

			portalPublishers = append(portalPublishers, portalCruncherPublisher)
		}
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

	// Setup feature config for vanity metrics
	var featureConfig config.Config
	envVarConfig := config.NewEnvVarConfig([]config.Feature{
		{
			Name:        "FEATURE_VANITY_METRIC",
			Enum:        config.FEATURE_VANITY_METRIC,
			Value:       false,
			Description: "Vanity metrics for fast aggregate statistic lookup",
		},
	})
	featureConfig = envVarConfig
	// Determine if should use vanity metrics
	useVanityMetrics := featureConfig.FeatureEnabled(config.FEATURE_VANITY_METRIC)

	// Start vanity metrics publisher
	vanityPublishers := make([]pubsub.Publisher, 0)
	{
		vanityMetricHosts := envvar.GetList("FEATURE_VANITY_METRIC_HOSTS", []string{"tcp://127.0.0.1:6666"})

		postVanityMetricSendBufferSize, err := envvar.GetInt("FEATURE_VANITY_METRIC_POST_SEND_BUFFER_SIZE", 1000000)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		for _, host := range vanityMetricHosts {
			vanityPublisher, err := pubsub.NewVanityMetricPublisher(host, postVanityMetricSendBufferSize)
			if err != nil {
				level.Error(logger).Log("msg", "could not create vanity metric publisher", "err", err)
				return 1
			}

			vanityPublishers = append(vanityPublishers, vanityPublisher)
		}
	}

	postVanityMetricMaxRetries, err := envvar.GetInt("FEATURE_VANITY_METRIC_POST_MAX_RETRIES", 10)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Create a post session handler to handle the post process of session updates.
	// This way, we can quickly return from the session update handler and not spawn a
	// ton of goroutines if things get backed up.
	postSessionHandler := transport.NewPostSessionHandler(numPostSessionGoroutines, postSessionBufferSize, portalPublishers, postSessionPortalMaxRetries, vanityPublishers, postVanityMetricMaxRetries, useVanityMetrics, biller, logger, backendMetrics.PostSessionMetrics)
	go postSessionHandler.StartProcessing(ctx)

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
				case <-ctx.Done():
					return
				}
			}
		}(ctx)
	}

	maxNearRelays, err := envvar.GetInt("MAX_NEAR_RELAYS", 32)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	if maxNearRelays > 32 {
		level.Error(logger).Log("err", "cannot support more than 32 near relays")
		return 1
	}

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, false, []string{}))
		router.Handle("/debug/vars", expvar.Handler())

		go func() {
			httpPort := envvar.Get("HTTP_PORT", "40001")

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

	serverInitHandler := transport.ServerInitHandlerFunc(log.With(logger, "handler", "server_init"), storer, backendMetrics.ServerInitMetrics)
	serverUpdateHandler := transport.ServerUpdateHandlerFunc(log.With(logger, "handler", "server_update"), storer, postSessionHandler, backendMetrics.ServerUpdateMetrics)
	sessionUpdateHandler := transport.SessionUpdateHandlerFunc(log.With(logger, "handler", "session_update"), getIPLocatorFunc, getRouteMatrixFunc, multipathVetoHandler, storer, maxNearRelays, routerPrivateKey, postSessionHandler, backendMetrics.SessionUpdateMetrics)

	for i := 0; i < numThreads; i++ {
		go func(thread int) {
			lp, err := lc.ListenPacket(ctx, "udp", "0.0.0.0:"+udpPort)
			if err != nil {
				panic(fmt.Sprintf("could not bind socket: %v", err))
			}

			conn := lp.(*net.UDPConn)
			defer conn.Close()

			if err := conn.SetReadBuffer(readBuffer); err != nil {
				panic(fmt.Sprintf("could not set connection read buffer size: %v", err))
			}

			if err := conn.SetWriteBuffer(writeBuffer); err != nil {
				panic(fmt.Sprintf("could not set connection write buffer size: %v", err))
			}

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
				case transport.PacketTypeServerInitRequest:
					serverInitHandler(&buffer, &packet)
				case transport.PacketTypeServerUpdate:
					serverUpdateHandler(&buffer, &packet)
				case transport.PacketTypeSessionUpdate:
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

	return 0
}
