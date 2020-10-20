/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/networknext/backend/analytics"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/transport"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
	"cloud.google.com/go/pubsub"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"

	"github.com/networknext/backend/modules/core"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {

	fmt.Printf("relay_backend: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)
	relayslogger := log.NewLogfmtLogger(os.Stdout)

	var enableSDLogging bool
	enableSDLoggingString, ok := os.LookupEnv("ENABLE_STACKDRIVER_LOGGING")
	if ok {
		var err error
		enableSDLogging, err = strconv.ParseBool(enableSDLoggingString)
		if err != nil {
			level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_LOGGING", "msg", "could not parse", "err", err)
			os.Exit(1)
		}
	}

	if enableSDLogging {
		if projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
			loggingClient, err := gcplogging.NewClient(ctx, projectID)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			logger = logging.NewStackdriverLogger(loggingClient, "relay-backend")
			relayslogger = logging.NewStackdriverLogger(loggingClient, "relays")
		}
	}

	{
		switch os.Getenv("BACKEND_LOG_LEVEL") {
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

		switch os.Getenv("RELAYS_LOG_LEVEL") {
		case "none":
			relayslogger = level.NewFilter(relayslogger, level.AllowNone())
		case level.ErrorValue().String():
			relayslogger = level.NewFilter(relayslogger, level.AllowError())
		case level.WarnValue().String():
			relayslogger = level.NewFilter(relayslogger, level.AllowWarn())
		case level.InfoValue().String():
			relayslogger = level.NewFilter(relayslogger, level.AllowInfo())
		case level.DebugValue().String():
			relayslogger = level.NewFilter(relayslogger, level.AllowDebug())
		default:
			relayslogger = level.NewFilter(relayslogger, level.AllowWarn())
		}
		relayslogger = log.With(relayslogger, "ts", log.DefaultTimestampUTC)
	}

	// Get env
	env, ok := os.LookupEnv("ENV")
	if !ok {
		level.Error(logger).Log("err", "ENV not set")
		os.Exit(1)
	}

	fmt.Printf("env is %s\n", env)

	var customerPublicKey []byte
	{
		if key := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(key) != 0 {
			customerPublicKey, _ = base64.StdEncoding.DecodeString(key)
			customerPublicKey = customerPublicKey[8:]
		}
	}

	var customerID uint64
	if key := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(key) != 0 {
		customerPublicKey, _ = base64.StdEncoding.DecodeString(key)
		customerID = binary.LittleEndian.Uint64(customerPublicKey[:8])
		customerPublicKey = customerPublicKey[8:]
	}

	var relayPublicKey []byte
	{
		if key := os.Getenv("RELAY_PUBLIC_KEY"); len(key) != 0 {
			relayPublicKey, _ = base64.StdEncoding.DecodeString(key)
		}
	}

	var routerPrivateKey []byte
	{
		if key := os.Getenv("RELAY_ROUTER_PRIVATE_KEY"); len(key) != 0 {
			routerPrivateKey, _ = base64.StdEncoding.DecodeString(key)
		} else {
			level.Error(logger).Log("err", "RELAY_ROUTER_PRIVATE_KEY not set")
			os.Exit(1)
		}
	}

	var db storage.Storer = &storage.InMemory{
		LocalMode: true,
	}

	gcpProjectID, gcpOK := os.LookupEnv("GOOGLE_PROJECT_ID")
	_, emulatorOK := os.LookupEnv("FIRESTORE_EMULATOR_HOST")
	if emulatorOK {
		fmt.Printf("using firestore emulator\n")
		gcpProjectID = "local"
		level.Info(logger).Log("msg", "Detected firestore emulator")
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpOK || emulatorOK {

		// Firestore
		{
			fmt.Printf("initializing firestore\n")

			// Create a Firestore Storer
			fs, err := storage.NewFirestore(ctx, gcpProjectID, logger)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			fssyncinterval := os.Getenv("GOOGLE_FIRESTORE_SYNC_INTERVAL")
			syncInterval, err := time.ParseDuration(fssyncinterval)
			if err != nil {
				level.Error(logger).Log("envvar", "GOOGLE_FIRESTORE_SYNC_INTERVAL", "value", fssyncinterval, "err", err)
				os.Exit(1)
			}
			// Start a goroutine to sync from Firestore
			go func() {

				ticker := time.NewTicker(syncInterval)
				fs.SyncLoop(ctx, ticker.C)
			}()

			// Set the Firestore Storer to give to handlers
			db = fs
		}
	}

	if env == "local" {
		storage.SeedStorage(logger, ctx, db, relayPublicKey, customerID, customerPublicKey)
	}

	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	if gcpOK {
		// Stackdriver Metrics
		{
			fmt.Printf("initializing stackdriver metrics\n")

			var enableSDMetrics bool
			var err error
			enableSDMetricsString, ok := os.LookupEnv("ENABLE_STACKDRIVER_METRICS")
			if ok {
				enableSDMetrics, err = strconv.ParseBool(enableSDMetricsString)
				if err != nil {
					level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_METRICS", "msg", "could not parse", "err", err)
					os.Exit(1)
				}
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

				sdwriteinterval := os.Getenv("GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL")
				writeInterval, err := time.ParseDuration(sdwriteinterval)
				if err != nil {
					level.Error(logger).Log("envvar", "GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL", "value", sdwriteinterval, "err", err)
					os.Exit(1)
				}
				go func() {
					metricsHandler.WriteLoop(ctx, logger, writeInterval, 200)
				}()
			}
		}

		// Stackdriver Profiler
		{
			fmt.Printf("initializing stackdriver profiler\n")

			var enableSDProfiler bool
			var err error
			enableSDProfilerString, ok := os.LookupEnv("ENABLE_STACKDRIVER_PROFILER")
			if ok {
				enableSDProfiler, err = strconv.ParseBool(enableSDProfilerString)
				if err != nil {
					level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_PROFILER", "msg", "could not parse", "err", err)
					os.Exit(1)
				}
			}

			if enableSDProfiler {
				// Set up StackDriver profiler
				if err := profiler.Start(profiler.Config{
					Service:        "relay_backend",
					ServiceVersion: env,
					ProjectID:      gcpProjectID,
					MutexProfiling: true,
				}); err != nil {
					level.Error(logger).Log("msg", "Failed to initialize StackDriver profiler", "err", err)
					os.Exit(1)
				}
			}
		}
	}

	// Create relay init metrics
	relayInitMetrics, err := metrics.NewRelayInitMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create relay init metrics", "err", err)
	}

	// Create relay update metrics
	relayUpdateMetrics, err := metrics.NewRelayUpdateMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create relay update metrics", "err", err)
	}

	costMatrixMetrics, err := metrics.NewCostMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create cost matrix metrics", "err", err)
	}

	optimizeMetrics, err := metrics.NewOptimizeMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create optimize metrics", "err", err)
	}

	valveCostMatrixMetrics, err := metrics.NewValveCostMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create valve cost matrix metrics", "err", err)
	}

	valveOptimizeMetrics, err := metrics.NewValveOptimizeMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create valve optimize metrics", "err", err)
	}

	valveRouteMatrixMetrics, err := metrics.NewValveRouteMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create valve route matrix metrics", "err", err)
	}

	relayBackendMetrics, err := metrics.NewRelayBackendMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create relay backend metrics", "err", err)
	}

	statsdb := routing.NewStatsDatabase()
	costMatrix := []int32{}
	routeMatrix := &routing.RouteMatrix4{}

	var costMatrixMutex sync.RWMutex
	getCostMatrixFunc := func() []int32 {
		costMatrixMutex.RLock()
		cm := costMatrix
		costMatrixMutex.RUnlock()
		return cm
	}

	var routeMatrixMutex sync.RWMutex
	getRouteMatrixFunc := func() *routing.RouteMatrix4 {
		routeMatrixMutex.RLock()
		rm := routeMatrix
		routeMatrixMutex.RUnlock()
		return rm
	}

	// Get the max jitter and max packet loss env vars
	var maxJitter float64
	var maxPacketLoss float64
	{
		maxJitterString, ok := os.LookupEnv("RELAY_ROUTER_MAX_JITTER")
		if !ok {
			level.Error(logger).Log("msg", "env var not set", "envvar", "RELAY_ROUTER_MAX_JITTER")
			os.Exit(1)
		}

		maxJitter, err = strconv.ParseFloat(maxJitterString, 32)
		if err != nil {
			level.Error(logger).Log("err", "could not parse max jitter", "value", maxJitterString)
			os.Exit(1)
		}

		maxPacketLossString, ok := os.LookupEnv("RELAY_ROUTER_MAX_PACKET_LOSS")
		if !ok {
			level.Error(logger).Log("msg", "env var not set", "envvar", "RELAY_ROUTER_MAX_PACKET_LOSS")
			os.Exit(1)
		}

		maxPacketLoss, err = strconv.ParseFloat(maxPacketLossString, 32)
		if err != nil {
			level.Error(logger).Log("err", "could not parse max packet loss", "value", maxPacketLossString)
			os.Exit(1)
		}
	}

	// Create the relay map
	cleanupCallback := func(relayData *routing.RelayData) error {
		// Remove relay entry from statsDB (which in turn means it won't appear in cost matrix)
		statsdb.DeleteEntry(relayData.ID)
		level.Warn(logger).Log("msg", "relay timed out", "relay ID", relayData.ID, "relay addr", relayData.Addr.String(), "relay name", relayData.Name)
		return nil
	}

	relayMap := routing.NewRelayMap(cleanupCallback)
	go func() {
		timeout := int64(routing.RelayTimeout.Seconds())
		frequency := time.Second
		ticker := time.NewTicker(frequency)
		relayMap.TimeoutLoop(ctx, timeout, ticker.C)
	}()

	// ping stats
	var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
	{
		// Create a no-op publisher
		_, emulatorOK = os.LookupEnv("PUBSUB_EMULATOR_HOST")
		if gcpOK || emulatorOK {

			pubsubCtx := ctx
			if emulatorOK {
				gcpProjectID = "local"

				var cancelFunc context.CancelFunc
				pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(60*time.Minute))
				defer cancelFunc()

				level.Info(logger).Log("msg", "Detected pubsub emulator")
			}

			// Google Pubsub
			{
				settings := pubsub.PublishSettings{
					DelayThreshold: time.Second,
					CountThreshold: 1,
					ByteThreshold:  1 << 14,
					NumGoroutines:  runtime.GOMAXPROCS(0),
					Timeout:        time.Minute,
				}

				pubsub, err := analytics.NewGooglePubSubPingStatsPublisher(pubsubCtx, &relayBackendMetrics.PingStatsMetrics, logger, gcpProjectID, "ping_stats", settings)
				if err != nil {
					level.Error(logger).Log("msg", "could not create analytics pubsub publisher", "err", err)
					os.Exit(1)
				}

				pingStatsPublisher = pubsub
			}
		}

		go func() {
			sleepTime := time.Minute
			if publishInterval, ok := os.LookupEnv("PING_STATS_PUBLISH_INTERVAL"); ok {
				if duration, err := time.ParseDuration(publishInterval); err == nil {
					sleepTime = duration
				} else {
					level.Error(logger).Log("msg", "could not parse publish interval", "err", err)
				}
			}

			for {
				time.Sleep(sleepTime)
				cpy := statsdb.MakeCopy()
				entries := analytics.ExtractPingStats(cpy)
				if err := pingStatsPublisher.Publish(ctx, entries); err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}
			}
		}()
	}

	// relay stats
	var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
	{
		_, emulatorOK = os.LookupEnv("PUBSUB_EMULATOR_HOST")
		if gcpOK || emulatorOK {

			pubsubCtx := ctx
			if emulatorOK {
				gcpProjectID = "local"

				var cancelFunc context.CancelFunc
				pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(60*time.Minute))
				defer cancelFunc()

				level.Info(logger).Log("msg", "Detected pubsub emulator")
			}

			// Google Pubsub
			{
				settings := pubsub.PublishSettings{
					DelayThreshold: time.Second,
					CountThreshold: 1,
					ByteThreshold:  1 << 14,
					NumGoroutines:  runtime.GOMAXPROCS(0),
					Timeout:        time.Minute,
				}

				pubsub, err := analytics.NewGooglePubSubRelayStatsPublisher(pubsubCtx, &relayBackendMetrics.RelayStatsMetrics, logger, gcpProjectID, "relay_stats", settings)
				if err != nil {
					level.Error(logger).Log("msg", "could not create analytics pubsub publisher", "err", err)
					os.Exit(1)
				}

				relayStatsPublisher = pubsub
			}
		}

		go func() {
			sleepTime := time.Second * 10
			if publishInterval, ok := os.LookupEnv("RELAY_STATS_PUBLISH_INTERVAL"); ok {
				if duration, err := time.ParseDuration(publishInterval); err == nil {
					sleepTime = duration
				} else {
					level.Error(logger).Log("msg", "could not parse publish interval", "err", err)
				}
			}

			for {
				time.Sleep(sleepTime)
				allRelayData := relayMap.GetAllRelayData()
				entries := make([]analytics.RelayStatsEntry, len(allRelayData))

				count := 0
				for _, relay := range allRelayData {
					// convert peak to mbps

					var traffic routing.TrafficStats

					relay.TrafficMu.Lock()
					for i := range relay.TrafficStatsBuff {
						stats := &relay.TrafficStatsBuff[i]
						traffic = traffic.Add(stats)
					}
					relay.TrafficStatsBuff = relay.TrafficStatsBuff[:0]
					numSessions := relay.PeakTrafficStats.SessionCount
					envUp := relay.PeakTrafficStats.EnvelopeUpKbps
					envDown := relay.PeakTrafficStats.EnvelopeDownKbps
					relay.PeakTrafficStats.SessionCount = 0
					relay.PeakTrafficStats.EnvelopeUpKbps = 0
					relay.PeakTrafficStats.EnvelopeDownKbps = 0
					relay.TrafficMu.Unlock()

					elapsed := time.Since(relay.LastStatsPublishTime)
					relay.LastStatsPublishTime = time.Now()

					fsrelay, err := db.Relay(relay.ID)
					if err != nil {
						continue
					}

					// use the sum of all the stats since the last publish here and convert to mbps
					bwSentMbps := float32(float64(traffic.AllTx()) * 8.0 / 1000000.0 / elapsed.Seconds())
					bwRecvMbps := float32(float64(traffic.AllRx()) * 8.0 / 1000000.0 / elapsed.Seconds())

					// use the peak envelope values here and convert, it's already per second so no need for time adjustment
					envSentMbps := float32(float64(envUp) / 1000.0)
					envRecvMbps := float32(float64(envDown) / 1000.0)

					var numRouteable uint32 = 0
					for _, otherRelay := range allRelayData {
						if relay.ID == otherRelay.ID {
							continue
						}

						rtt, jitter, pl := statsdb.GetSample(relay.ID, otherRelay.ID)
						if rtt != routing.InvalidRouteValue && jitter != routing.InvalidRouteValue && pl != routing.InvalidRouteValue {
							numRouteable++
						}
					}

					var bwSentPercent float32
					var bwRecvPercent float32
					var envSentPercent float32
					var envRecvPercent float32
					if fsrelay.NICSpeedMbps != 0 {
						bwSentPercent = bwSentMbps / float32(fsrelay.NICSpeedMbps) * 100.0
						bwRecvPercent = bwRecvMbps / float32(fsrelay.NICSpeedMbps) * 100.0
						envSentPercent = envSentMbps / float32(fsrelay.NICSpeedMbps) * 100.0
						envRecvPercent = envRecvMbps / float32(fsrelay.NICSpeedMbps) * 100.0
					}

					entries[count] = analytics.RelayStatsEntry{
						ID:                       relay.ID,
						CPUUsage:                 relay.CPUUsage,
						MemUsage:                 relay.MemUsage,
						BandwidthSentPercent:     bwSentPercent,
						BandwidthReceivedPercent: bwRecvPercent,
						EnvelopeSentPercent:      envSentPercent,
						EnvelopeReceivedPercent:  envRecvPercent,
						BandwidthSentMbps:        bwSentMbps,
						BandwidthReceivedMbps:    bwRecvMbps,
						EnvelopeSentMbps:         envSentMbps,
						EnvelopeReceivedMbps:     envRecvMbps,
						NumSessions:              uint32(numSessions),
						MaxSessions:              relay.MaxSessions,
						NumRoutable:              numRouteable,
						NumUnroutable:            uint32(len(allRelayData)) - 1 - numRouteable,
					}

					count++
				}

				if err := relayStatsPublisher.Publish(ctx, entries[:count]); err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}
			}
		}()

	}

	cmsyncinterval := os.Getenv("COST_MATRIX_INTERVAL")
	syncInterval, err := time.ParseDuration(cmsyncinterval)
	if err != nil {
		level.Error(logger).Log("envvar", "COST_MATRIX_INTERVAL", "value", cmsyncinterval, "err", err)
		os.Exit(1)
	}

	// Separate route matrix specifically for Valve
	valveRouteMatrix := &routing.RouteMatrix4{}
	var valveRouteMatrixMutex sync.RWMutex

	getValveRouteMatrixFunc := func() *routing.RouteMatrix4 {
		valveRouteMatrixMutex.RLock()
		rm := valveRouteMatrix
		valveRouteMatrixMutex.RUnlock()
		return rm
	}

	routeMatrixBufferSize := 100000
	if routeMatrixBufferSizeString, ok := os.LookupEnv("ROUTE_MATRIX_BUFFER_SIZE"); ok {
		routeMatrixBufferSize, err = strconv.Atoi(routeMatrixBufferSizeString)
		if err != nil {
			level.Error(logger).Log("envvar", "ROUTE_MATRIX_BUFFER_SIZE", "value", routeMatrixBufferSize, "err", err)
			os.Exit(1)
		}
	}

	// Generate the route matrix
	go func() {
		for {
			// For now, exclude all valve relays
			relayIDs := make([]uint64, 0)
			allRelayData := relayMap.GetAllRelayData()
			for _, relayData := range allRelayData {
				if relayData.Seller.ID != "valve" { // Filter out any relays whose seller has a Firestore key of "valve"
					relayIDs = append(relayIDs, relayData.ID)
				}
			}

			numRelays := len(relayIDs)
			relayAddresses := make([]net.UDPAddr, numRelays)
			relayNames := make([]string, numRelays)
			relayLatitudes := make([]float32, numRelays)
			relayLongitudes := make([]float32, numRelays)
			relayDatacenterIDs := make([]uint64, numRelays)

			for i, relayID := range relayIDs {
				relay, err := db.Relay(relayID)
				if err != nil {
					continue
				}

				relayAddresses[i] = relay.Addr
				relayNames[i] = relay.Name
				relayLatitudes[i] = float32(relay.Datacenter.Location.Latitude)
				relayLongitudes[i] = float32(relay.Datacenter.Location.Longitude)
				relayDatacenterIDs[i] = relay.Datacenter.ID
			}

			costMatrixMetrics.Invocations.Add(1)
			costMatrixDurationStart := time.Now()

			costMatrixNew := statsdb.GenerateCostMatrix(relayIDs, float32(maxJitter), float32(maxPacketLoss))

			costMatrixDurationSince := time.Since(costMatrixDurationStart)
			costMatrixMetrics.DurationGauge.Set(float64(costMatrixDurationSince.Milliseconds()))
			if costMatrixDurationSince.Seconds() > 1.0 {
				costMatrixMetrics.LongUpdateCount.Add(1)
			}

			costMatrixMetrics.Bytes.Set(float64(len(costMatrix) * 4))

			costMatrixMutex.Lock()
			costMatrix = costMatrixNew
			costMatrixMutex.Unlock()

			numCPUs := runtime.NumCPU()
			numSegments := numRelays
			if numCPUs < numRelays {
				numSegments = numRelays / 5
				if numSegments == 0 {
					numSegments = 1
				}
			}

			optimizeMetrics.Invocations.Add(1)
			optimizeDurationStart := time.Now()

			routeEntries := core.Optimize(numRelays, numSegments, costMatrixNew, 5, relayDatacenterIDs)
			if len(routeEntries) == 0 {
				level.Warn(logger).Log("matrix", "cost", "op", "optimize", "warn", "no route entries generated from cost matrix")
				time.Sleep(syncInterval)
				continue
			}

			optimizeDurationSince := time.Since(optimizeDurationStart)
			optimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

			if optimizeDurationSince.Seconds() > 1.0 {
				optimizeMetrics.LongUpdateCount.Add(1)
			}

			routeMatrixNew := &routing.RouteMatrix4{
				RelayIDs:           relayIDs,
				RelayAddresses:     relayAddresses,
				RelayNames:         relayNames,
				RelayLatitudes:     relayLatitudes,
				RelayLongitudes:    relayLongitudes,
				RelayDatacenterIDs: relayDatacenterIDs,
				RouteEntries:       routeEntries,
			}

			if err := routeMatrixNew.WriteResponseData(routeMatrixBufferSize); err != nil {
				level.Error(logger).Log("matrix", "route", "op", "write_response", "msg", "could not write response data", "err", err)
				time.Sleep(syncInterval)
				continue
			}

			routeMatrixNew.WriteAnalysisData()

			routeMatrixMutex.Lock()
			routeMatrix = routeMatrixNew
			routeMatrixMutex.Unlock()

			relayBackendMetrics.RouteMatrix.Bytes.Set(float64(len(routeMatrixNew.GetResponseData())))
			relayBackendMetrics.RouteMatrix.RelayCount.Set(float64(len(routeMatrixNew.RelayIDs)))
			relayBackendMetrics.RouteMatrix.DatacenterCount.Set(float64(len(routeMatrixNew.RelayDatacenterIDs)))

			// todo: calculate this in optimize and store in route matrix so we don't have to calc this here
			numRoutes := int32(0)
			for i := range routeMatrixNew.RouteEntries {
				numRoutes += routeMatrixNew.RouteEntries[i].NumRoutes
			}
			relayBackendMetrics.RouteMatrix.RouteCount.Set(float64(numRoutes))

			memoryUsed := func() float64 {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				return float64(m.Alloc) / (1000.0 * 1000.0)
			}

			relayBackendMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
			relayBackendMetrics.MemoryAllocated.Set(memoryUsed())

			fmt.Printf("-----------------------------\n")
			fmt.Printf("%.2f mb allocated\n", relayBackendMetrics.MemoryAllocated.Value())
			fmt.Printf("%d goroutines\n", int(relayBackendMetrics.Goroutines.Value()))
			fmt.Printf("%d datacenters\n", int(relayBackendMetrics.RouteMatrix.DatacenterCount.Value()))
			fmt.Printf("%d relays\n", int(relayBackendMetrics.RouteMatrix.RelayCount.Value()))
			fmt.Printf("%d relays in map\n", relayMap.GetRelayCount())
			fmt.Printf("%d routes\n", int(relayBackendMetrics.RouteMatrix.RouteCount.Value()))
			fmt.Printf("%d long cost matrix updates\n", int(costMatrixMetrics.LongUpdateCount.Value()))
			fmt.Printf("%d long route matrix updates\n", int(optimizeMetrics.LongUpdateCount.Value()))
			fmt.Printf("cost matrix update: %.2f milliseconds\n", costMatrixMetrics.DurationGauge.Value())
			fmt.Printf("route matrix update: %.2f milliseconds\n", optimizeMetrics.DurationGauge.Value())
			fmt.Printf("cost matrix bytes: %d\n", int(costMatrixMetrics.Bytes.Value()))
			fmt.Printf("route matrix bytes: %d\n", int(relayBackendMetrics.RouteMatrix.Bytes.Value()))
			fmt.Printf("%d ping stats entries submitted\n", int(relayBackendMetrics.PingStatsMetrics.EntriesSubmitted.Value()))
			fmt.Printf("%d ping stats entries queued\n", int(relayBackendMetrics.PingStatsMetrics.EntriesQueued.Value()))
			fmt.Printf("%d ping stats entries flushed\n", int(relayBackendMetrics.PingStatsMetrics.EntriesFlushed.Value()))
			fmt.Printf("%d relay stats entries submitted\n", int(relayBackendMetrics.RelayStatsMetrics.EntriesSubmitted.Value()))
			fmt.Printf("%d relay stats entries queued\n", int(relayBackendMetrics.RelayStatsMetrics.EntriesQueued.Value()))
			fmt.Printf("%d relay stats entries flushed\n", int(relayBackendMetrics.RelayStatsMetrics.EntriesFlushed.Value()))
			fmt.Printf("-----------------------------\n")

			time.Sleep(syncInterval)
		}
	}()

	// Generate the route matrix specifically for valve
	go func() {
		for {
			// All relays included
			relayIDs := make([]uint64, 0)
			allRelayData := relayMap.GetAllRelayData()
			for _, relayData := range allRelayData {
				relayIDs = append(relayIDs, relayData.ID)
			}

			numRelays := len(relayIDs)
			relayAddresses := make([]net.UDPAddr, numRelays)
			relayNames := make([]string, numRelays)
			relayLatitudes := make([]float32, numRelays)
			relayLongitudes := make([]float32, numRelays)
			relayDatacenterIDs := make([]uint64, numRelays)

			for i, relayID := range relayIDs {
				relay, err := db.Relay(relayID)
				if err != nil {
					continue
				}

				relayAddresses[i] = relay.Addr
				relayNames[i] = relay.Name
				relayLatitudes[i] = float32(relay.Datacenter.Location.Latitude)
				relayLongitudes[i] = float32(relay.Datacenter.Location.Longitude)
				relayDatacenterIDs[i] = relay.Datacenter.ID
			}

			valveCostMatrixMetrics.Invocations.Add(1)
			costMatrixDurationStart := time.Now()

			valveCostMatrix := statsdb.GenerateCostMatrix(relayIDs, float32(maxJitter), float32(maxPacketLoss))

			costMatrixDurationSince := time.Since(costMatrixDurationStart)
			valveCostMatrixMetrics.DurationGauge.Set(float64(costMatrixDurationSince.Milliseconds()))
			if costMatrixDurationSince.Seconds() > 1.0 {
				valveCostMatrixMetrics.LongUpdateCount.Add(1)
			}

			valveCostMatrixMetrics.Bytes.Set(float64(len(valveCostMatrix) * 4))

			numCPUs := runtime.NumCPU()
			numSegments := numRelays
			if numCPUs < numRelays {
				numSegments = numRelays / 5
				if numSegments == 0 {
					numSegments = 1
				}
			}

			valveOptimizeMetrics.Invocations.Add(1)
			optimizeDurationStart := time.Now()

			routeEntries := core.Optimize(numRelays, numSegments, valveCostMatrix, 5, relayDatacenterIDs)
			if len(routeEntries) == 0 {
				level.Warn(logger).Log("matrix", "cost", "op", "optimize", "warn", "no route entries generated from cost matrix")
				time.Sleep(syncInterval)
				continue
			}

			optimizeDurationSince := time.Since(optimizeDurationStart)
			valveOptimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

			if optimizeDurationSince.Seconds() > 1.0 {
				valveOptimizeMetrics.LongUpdateCount.Add(1)
			}

			valveRouteMatrixNew := &routing.RouteMatrix4{
				RelayIDs:           relayIDs,
				RelayAddresses:     relayAddresses,
				RelayNames:         relayNames,
				RelayLatitudes:     relayLatitudes,
				RelayLongitudes:    relayLongitudes,
				RelayDatacenterIDs: relayDatacenterIDs,
				RouteEntries:       routeEntries,
			}

			if err := valveRouteMatrixNew.WriteResponseData(routeMatrixBufferSize); err != nil {
				level.Error(logger).Log("matrix", "route", "op", "write_response", "msg", "could not write response data", "err", err)
				time.Sleep(syncInterval)
				continue
			}

			valveRouteMatrixNew.WriteAnalysisData()

			valveRouteMatrixMutex.Lock()
			valveRouteMatrix = valveRouteMatrixNew
			valveRouteMatrixMutex.Unlock()

			valveRouteMatrixMetrics.Bytes.Set(float64(len(valveRouteMatrixNew.GetResponseData())))
			valveRouteMatrixMetrics.RelayCount.Set(float64(len(valveRouteMatrixNew.RelayIDs)))
			valveRouteMatrixMetrics.DatacenterCount.Set(float64(len(valveRouteMatrixNew.RelayDatacenterIDs)))

			// todo: calculate this in optimize and store in route matrix so we don't have to calc this here
			numRoutes := int32(0)
			for i := range valveRouteMatrixNew.RouteEntries {
				numRoutes += valveRouteMatrixNew.RouteEntries[i].NumRoutes
			}
			valveRouteMatrixMetrics.RouteCount.Set(float64(numRoutes))

			time.Sleep(syncInterval)
		}
	}()

	commonInitParams := transport.RelayInitHandlerConfig{
		RelayMap:         relayMap,
		Storer:           db,
		Metrics:          relayInitMetrics,
		RouterPrivateKey: routerPrivateKey,
	}

	commonUpdateParams := transport.RelayUpdateHandlerConfig{
		RelayMap: relayMap,
		StatsDB:  statsdb,
		Metrics:  relayUpdateMetrics,
		Storer:   db,
	}

	serveRouteMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		routeMatrix := getRouteMatrixFunc()

		data := routeMatrix.GetResponseData()

		buffer := bytes.NewBuffer(data)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	serveValveRouteMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		m := getValveRouteMatrixFunc()

		data := m.GetResponseData()

		buffer := bytes.NewBuffer(data)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	serveCostMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		m := getCostMatrixFunc()

		data := make([]byte, len(m)*4)
		var index int

		encoding.WriteUint32(data, &index, uint32(len(m)))
		for _, v := range m {
			encoding.WriteUint32(data, &index, uint32(v))
		}

		buffer := bytes.NewBuffer(data)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))
	router.HandleFunc("/relay_init", transport.RelayInitHandlerFunc(logger, &commonInitParams)).Methods("POST")
	router.HandleFunc("/relay_update", transport.RelayUpdateHandlerFunc(logger, relayslogger, &commonUpdateParams)).Methods("POST")
	router.HandleFunc("/cost_matrix", serveCostMatrixFunc).Methods("GET")
	router.HandleFunc("/route_matrix", serveRouteMatrixFunc).Methods("GET")
	router.HandleFunc("/route_matrix_valve", serveValveRouteMatrixFunc).Methods("GET")
	router.Handle("/debug/vars", expvar.Handler())
	router.HandleFunc("/relay_dashboard", transport.RelayDashboardHandlerFunc(relayMap, getRouteMatrixFunc, statsdb, "local", "local", maxJitter))
	router.HandleFunc("/relay_stats", transport.RelayStatsFunc(logger, relayMap))

	go func() {
		port, ok := os.LookupEnv("PORT")
		if !ok {
			level.Error(logger).Log("err", "env var PORT must be set")
			os.Exit(1)
		}

		level.Info(logger).Log("addr", ":"+port)

		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
