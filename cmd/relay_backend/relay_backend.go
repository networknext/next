/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"expvar"
	"fmt"
	"hash/fnv"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/networknext/backend/modules/config"

	"github.com/networknext/backend/modules/storage"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/common/helpers"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"

	gcStorage "cloud.google.com/go/storage"
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

	serviceName := "relay_backend"

	fmt.Printf("\n%s\n\n", serviceName)

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

	routerPrivateKey, err := envvar.GetBase64("RELAY_ROUTER_PRIVATE_KEY", nil)
	if err != nil {
		level.Error(logger).Log("err", "RELAY_ROUTER_PRIVATE_KEY not set")
		return 1
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

	relayBackendMetrics, err := metrics.NewRelayBackendMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create relay backend metrics", "err", err)
	}

	statsdb := routing.NewStatsDatabase()

	// Get the max jitter and max packet loss env vars
	if !envvar.Exists("RELAY_ROUTER_MAX_JITTER") {
		level.Error(logger).Log("err", "RELAY_ROUTER_MAX_JITTER not set")
		return 1
	}

	maxJitter, err := envvar.GetFloat("RELAY_ROUTER_MAX_JITTER", 0)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	if !envvar.Exists("RELAY_ROUTER_MAX_PACKET_LOSS") {
		level.Error(logger).Log("err", "RELAY_ROUTER_MAX_PACKET_LOSS not set")
		return 1
	}

	maxPacketLoss, err := envvar.GetFloat("RELAY_ROUTER_MAX_PACKET_LOSS", 0)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	featureConfig := config.NewEnvVarConfig([]config.Feature{
		{
			Name:        "FEATURE_NEW_RELAY_BACKEND_ENABLED",
			Enum:        config.FEATURE_NEW_RELAY_BACKEND,
			Value:       false,
			Description: "Enables the New Relay Backend project if true",
		},
		{
			Name:        "FEATURE_LOAD_TEST",
			Enum:        config.FEATURE_LOAD_TEST,
			Value:       false,
			Description: "Disables pubsub and storer usage when true",
		},
		{
			Name:        "FEATURE_ENABLE_PPROF",
			Enum:        config.FEATURE_ENABLE_PPROF,
			Value:       false,
			Description: "Allows access to PPROF http handlers when true",
		},
		{
			Name:        "FEATURE_ROUTE_MATRIX_STATS",
			Enum:        config.FEATURE_ROUTE_MATRIX_STATS,
			Value:       true,
			Description: "writes Route Matrix Stats to pubsub when true",
		},
		{
			Name:        "FEATURE_MATRIX_CLOUDSTORE",
			Enum:        config.FEATURE_MATRIX_CLOUDSTORE,
			Value:       false,
			Description: "writes Route Matrix to cloudstore when true",
		},
		{
			Name:        "FEATURE_VALVE_MATRIX",
			Enum:        config.FEATURE_VALVE_MATRIX,
			Value:       false,
			Description: "creates the valve matrix when true",
		},
	})

	var matrixStore storage.MatrixStore
	var backendLiveData storage.RelayBackendLiveData

	// update redis so relay frontend knows this backend is live
	if featureConfig.FeatureEnabled(config.FEATURE_NEW_RELAY_BACKEND) {
		var backendAddr string
		if !envvar.Exists("FEATURE_NEW_RELAY_BACKEND_ADDRESSES") {
			level.Error(logger).Log("FEATURE_NEW_RELAY_BACKEND_ADDRESSES not set")
			return 1
		}
		backendAddresses := envvar.GetList("FEATURE_NEW_RELAY_BACKEND_ADDRESSES", []string{})

		if env == "local" {
			backendAddr = backendAddresses[0]
		} else {

			addrFound := false
			host, _ := os.Hostname()
			addrs, _ := net.LookupIP(host)
			for _, addr := range addrs {
				if ipv4 := addr.To4(); ipv4 != nil {
					for _, addr := range backendAddresses {
						if ipv4.String() == addr {
							backendAddr = addr
							addrFound = true
							break
						}
					}
				}
				if addrFound {
					break
				}
			}

			if !addrFound {
				level.Error(logger).Log("relay backend address not found")
				return 1
			}
		}

		if !envvar.Exists("MATRIX_STORE_ADDRESS") {
			level.Error(logger).Log("MATRIX_STORE_ADDRESS not set")
			return 1
		}
		matrixStoreAddress := envvar.Get("MATRIX_STORE_ADDRESS", "")

		matrixStoreReadTimeout, err := envvar.GetDuration("MATRIX_STORE_READ_TIMEOUT", 250*time.Millisecond)
		if err != nil {
			level.Error(logger).Log(err.Error())
			return 1
		}

		matrixStoreWriteTimeout, err := envvar.GetDuration("MATRIX_STORE_WRITE_TIMEOUT", 250*time.Millisecond)
		if err != nil {
			level.Error(logger).Log(err.Error())
			return 1
		}

		matrixStore, err = storage.NewRedisMatrixStore(matrixStoreAddress, matrixStoreReadTimeout, matrixStoreWriteTimeout, 5*time.Second)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		backendLiveData.Id = gcpProjectID
		backendLiveData.Address = backendAddr
		backendLiveData.InitAt = time.Now()
	}

	// Create the relay map
	cleanupCallback := func(relayData routing.RelayData) error {
		// Remove relay entry from statsDB (which in turn means it won't appear in cost matrix)
		statsdb.DeleteEntry(relayData.ID)
		level.Warn(logger).Log("msg", "relay timed out", "relay ID", relayData.ID, "relay addr", relayData.Addr.String(), "relay name", relayData.Name)
		return nil
	}

	relayMap := routing.NewRelayMap(cleanupCallback)
	go func() {
		timeout := int64(routing.RelayTimeout.Seconds())
		frequency := time.Second * 10
		ticker := time.NewTicker(frequency)
		relayMap.TimeoutLoop(ctx, GetRelayData, timeout, ticker.C)
	}()

	// ping stats
	var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
	{
		if !featureConfig.FeatureEnabled(config.FEATURE_LOAD_TEST) {
			emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
			if gcpProjectID != "" || emulatorOK {

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
						return 1
					}

					pingStatsPublisher = pubsub
				}
			}
		}

		go func() {
			publishInterval, err := envvar.GetDuration("PING_STATS_PUBLISH_INTERVAL", time.Minute)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
			}

			syncTimer := helpers.NewSyncTimer(publishInterval)
			for {
				syncTimer.Run()

				cpy := statsdb.MakeCopy()
				entries := analytics.ExtractPingStats(cpy, float32(maxJitter), float32(maxPacketLoss))
				if err := pingStatsPublisher.Publish(ctx, entries); err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
				}
			}
		}()
	}

	// todo: nope
	/*
		// relay stats
		var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
		{
			if !!featureConfig.FeatureEnabled(config.FEATURE_LOAD_TEST) {
				emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
				if gcpProjectID != "" || emulatorOK {

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
							return 1
						}

						relayStatsPublisher = pubsub
					}
				}
			}

			go func() {
				publishInterval, err := envvar.GetDuration("RELAY_STATS_PUBLISH_INTERVAL", time.Second*10)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
				}

				syncTimer := helpers.NewSyncTimer(publishInterval)
				for {
					syncTimer.Run()

					allRelayData := relayMap.GetAllRelayData()
					entries := make([]analytics.RelayStatsEntry, len(allRelayData))

					count := 0
					for i := range allRelayData {
						relay := &allRelayData[i]

						// convert peak to mbps

						var traffic routing.TrafficStats

						relay.TrafficMu.Lock()
						for i := range relay.TrafficStatsBuff {
							stats := &relay.TrafficStatsBuff[i]
							traffic = traffic.Add(stats)
						}
						numSessions := relay.PeakTrafficStats.SessionCount
						envUp := relay.PeakTrafficStats.EnvelopeUpKbps
						envDown := relay.PeakTrafficStats.EnvelopeDownKbps
						elapsed := time.Since(relay.LastStatsPublishTime)

						relayMap.ClearRelayData(relay.Addr.String())
						relay.TrafficMu.Unlock()

						var storeRelay routing.Relay
						if !!featureConfig.FeatureEnabled(config.FEATURE_LOAD_TEST) {
							storeRelay, err = storer.Relay(relay.ID)
							if err != nil {
								level.Error(logger).Log("err", err)
								continue
							}
						} else {
							storeRelay = routing.Relay{NICSpeedMbps: 1}
						}

						// use the sum of all the stats since the last publish here and convert to mbps
						bwSentMbps := float32(float64(traffic.AllTx()) * 8.0 / 1000000.0 / elapsed.Seconds())
						bwRecvMbps := float32(float64(traffic.AllRx()) * 8.0 / 1000000.0 / elapsed.Seconds())

						// use the peak envelope values here and convert, it's already per second so no need for time adjustment
						envSentMbps := float32(float64(envUp) / 1000.0)
						envRecvMbps := float32(float64(envDown) / 1000.0)

						var numRouteable uint32 = 0
						for i := range allRelayData {
							otherRelay := &allRelayData[i]

							if relay.ID == otherRelay.ID {
								continue
							}

							rtt, jitter, pl := statsdb.GetSample(relay.ID, otherRelay.ID)
							if rtt != routing.InvalidRouteValue && jitter != routing.InvalidRouteValue && pl != routing.InvalidRouteValue {
								if jitter <= float32(maxJitter) && pl <= float32(maxPacketLoss) {
									numRouteable++
								}
							}
						}

						var bwSentPercent float32
						var bwRecvPercent float32
						var envSentPercent float32
						var envRecvPercent float32
						if storeRelay.NICSpeedMbps != 0 {
							bwSentPercent = bwSentMbps / float32(storeRelay.NICSpeedMbps) * 100.0
							bwRecvPercent = bwRecvMbps / float32(storeRelay.NICSpeedMbps) * 100.0
							envSentPercent = envSentMbps / float32(storeRelay.NICSpeedMbps) * 100.0
							envRecvPercent = envRecvMbps / float32(storeRelay.NICSpeedMbps) * 100.0
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

					entriesToPublish := entries[:count]
					if len(entriesToPublish) > 0 {
						if err := relayStatsPublisher.Publish(ctx, entriesToPublish); err != nil {
							level.Error(logger).Log("err", err)
							os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
						}
					}
				}
			}()
		}
	*/

	var relayNamesHashPublisher analytics.RouteMatrixStatsPublisher = &analytics.NoOpRouteMatrixStatsPublisher{}
	{
		if !!featureConfig.FeatureEnabled(config.FEATURE_LOAD_TEST) {
			emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
			if gcpProjectID != "" || emulatorOK {

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

					pubsub, err := analytics.NewGooglePubSubRouteMatrixStatsPublisher(pubsubCtx, &relayBackendMetrics.RouteMatrixStatsMetrics, logger, gcpProjectID, "route_matrix_stats", settings)
					if err != nil {
						level.Error(logger).Log("msg", "could not create analytics pubsub publisher", "err", err)
						return 1
					}

					relayNamesHashPublisher = pubsub
				}
			}
		}
	}

	var relayEnabledCache common.RelayEnabledCache
	if !!featureConfig.FeatureEnabled(config.FEATURE_LOAD_TEST) {
		relayEnabledCache := common.NewRelayEnabledCache(storer)
		relayEnabledCache.StartRunner(1 * time.Minute)
	}
	var gcBucket *gcStorage.BucketHandle
	if featureConfig.FeatureEnabled(config.FEATURE_MATRIX_CLOUDSTORE) {
		gcBucket, err = GCStoreConnect(ctx, gcpProjectID)
		if err != nil {
			level.Error(logger).Log("err", err)
		}
	}

	syncInterval, err := envvar.GetDuration("COST_MATRIX_INTERVAL", time.Second)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	matrixBufferSize, err := envvar.GetInt("MATRIX_BUFFER_SIZE", 100000000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	costMatrixData := new(helpers.MatrixData)
	routeMatrixData := new(helpers.MatrixData)

	routeMatrix := routing.RouteMatrix{} //still needed for the route dashboard
	var routeMatrixMutex sync.RWMutex
	getRouteMatrixFunc := func() *routing.RouteMatrix { // makes copy and returns pointer to copy
		routeMatrixMutex.RLock()
		rm := routeMatrix
		routeMatrixMutex.RUnlock()
		return &rm
	}

	// Generate the route matrix
	go func() {
		syncTimer := helpers.NewSyncTimer(syncInterval)
		for {
			syncTimer.Run()

			baseRelayIDs := relayMap.GetAllRelayIDs([]string{})

			namesMap := make(map[string]routing.Relay)
			numRelays := len(baseRelayIDs)
			relayAddresses := make([]net.UDPAddr, numRelays)
			relayNames := make([]string, numRelays)
			relayLatitudes := make([]float32, numRelays)
			relayLongitudes := make([]float32, numRelays)
			relayDatacenterIDs := make([]uint64, numRelays)
			relayIDs := make([]uint64, numRelays)

			// baseRelayAddresses undefinded
			// for i, relayAddress := range baseRelayAddresses {
			// 	relay := relayMap.GetCopyRelayData(relayAddress)
			// 	relayNames[i] = relay.Name
			// 	namesMap[relay.Name] = relay
			// }

			//sort relay names then populate other arrays
			sort.Strings(relayNames)
			for i, relayName := range relayNames {
				relay := namesMap[relayName]
				relayIDs[i] = relay.ID
				relayAddresses[i] = relay.Addr
				relayLatitudes[i] = float32(relay.Datacenter.Location.Latitude)
				relayLongitudes[i] = float32(relay.Datacenter.Location.Longitude)
				relayDatacenterIDs[i] = relay.Datacenter.ID
			}

			costMatrixMetrics.Invocations.Add(1)
			costMatrixDurationStart := time.Now()

			costMatrixNew := &routing.CostMatrix{
				RelayIDs:           relayIDs,
				RelayAddresses:     relayAddresses,
				RelayNames:         relayNames,
				RelayLatitudes:     relayLatitudes,
				RelayLongitudes:    relayLongitudes,
				RelayDatacenterIDs: relayDatacenterIDs,
				Costs:              statsdb.GetCosts(relayIDs, float32(maxJitter), float32(maxPacketLoss)),
			}

			costMatrixDurationSince := time.Since(costMatrixDurationStart)
			costMatrixMetrics.DurationGauge.Set(float64(costMatrixDurationSince.Milliseconds()))
			if costMatrixDurationSince.Seconds() > 1.0 {
				costMatrixMetrics.LongUpdateCount.Add(1)
			}

			if err := costMatrixNew.WriteResponseData(matrixBufferSize); err != nil {
				level.Error(logger).Log("matrix", "cost", "op", "write_response", "msg", "could not write response data", "err", err)
				continue
			}

			costMatrixData.SetMatrix(costMatrixNew.GetResponseData())
			costMatrixMetrics.Bytes.Set(float64(costMatrixData.GetMatrixDataSize()))

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

			costThreshold := int32(1)

			routeEntries := core.Optimize(numRelays, numSegments, costMatrixNew.Costs, costThreshold, relayDatacenterIDs)
			if len(routeEntries) == 0 {
				continue
			}

			optimizeDurationSince := time.Since(optimizeDurationStart)
			optimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

			if optimizeDurationSince.Seconds() > 1.0 {
				optimizeMetrics.LongUpdateCount.Add(1)
			}

			routeMatrixNew := &routing.RouteMatrix{
				RelayIDs:           relayIDs,
				RelayAddresses:     relayAddresses,
				RelayNames:         relayNames,
				RelayLatitudes:     relayLatitudes,
				RelayLongitudes:    relayLongitudes,
				RelayDatacenterIDs: relayDatacenterIDs,
				RouteEntries:       routeEntries,
			}

			if err := routeMatrixNew.WriteResponseData(matrixBufferSize); err != nil {
				level.Error(logger).Log("matrix", "route", "op", "write_response", "msg", "could not write response data", "err", err)
				continue
			}

			routeMatrixNew.WriteAnalysisData()

			routeMatrixMutex.Lock()
			routeMatrix = *routeMatrixNew
			routeMatrixMutex.Unlock()

			routeMatrixData.SetMatrix(routeMatrixNew.GetResponseData())

			relayBackendMetrics.RouteMatrix.Bytes.Set(float64(routeMatrixData.GetMatrixDataSize()))
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

			if featureConfig.FeatureEnabled(config.FEATURE_NEW_RELAY_BACKEND) {
				backendLiveData.UpdatedAt = time.Now()
				err = matrixStore.SetRelayBackendLiveData(backendLiveData)
				if err != nil {
					level.Error(logger).Log(err)
				}
			}

			if featureConfig.FeatureEnabled(config.FEATURE_MATRIX_CLOUDSTORE) {
				if gcBucket == nil {
					gcBucket, err = GCStoreConnect(ctx, gcpProjectID)
					if err != nil {
						level.Error(logger).Log("err", err)
						continue
					}
				}

				timestamp := time.Now().UTC()
				err = GCStoreMatrix(gcBucket, "cost", timestamp, costMatrixNew.GetResponseData())
				if err != nil {
					level.Error(logger).Log("err", err)
					continue
				}
				err = GCStoreMatrix(gcBucket, "route", timestamp, routeMatrixNew.GetResponseData())
				if err != nil {
					level.Error(logger).Log("err", err)
					continue
				}
			}

			if featureConfig.FeatureEnabled(config.FEATURE_ROUTE_MATRIX_STATS) && !featureConfig.FeatureEnabled(config.FEATURE_LOAD_TEST) {
				timestamp := time.Now().UTC().Unix()
				downRelayNames, downRelayIDs := relayEnabledCache.GetDownRelays(relayIDs)
				namesHashEntry := analytics.RouteMatrixStatsEntry{Timestamp: uint64(timestamp), Hash: uint64(0), IDs: downRelayIDs}
				if len(downRelayNames) != 0 {
					relayHash := fnv.New64a()
					for _, name := range downRelayNames {
						relayHash.Write([]byte(name))
					}

					hash := relayHash.Sum64()
					namesHashEntry.Hash = hash
				}

				err = relayNamesHashPublisher.Publish(ctx, namesHashEntry)
				if err != nil {
					level.Error(logger).Log("err", err)
				}
			}
		}
	}()

	commonInitParams := transport.RelayInitHandlerConfig{
		RelayMap:         relayMap,
		Metrics:          relayInitMetrics,
		RouterPrivateKey: routerPrivateKey,
		GetRelayData:     GetRelayData,
	}

	commonUpdateParams := transport.RelayUpdateHandlerConfig{
		RelayMap:     relayMap,
		StatsDB:      statsdb,
		Metrics:      relayUpdateMetrics,
		GetRelayData: GetRelayData,
	}

	serveRouteMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		buffer := bytes.NewBuffer(routeMatrixData.GetMatrix())
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	serveCostMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		buffer := bytes.NewBuffer(costMatrixData.GetMatrix())
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	fmt.Printf("starting http server\n\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/relay_init", transport.RelayInitHandlerFunc(&commonInitParams)).Methods("POST")
	router.HandleFunc("/relay_update", transport.RelayUpdateHandlerFunc(&commonUpdateParams)).Methods("POST")
	router.HandleFunc("/cost_matrix", serveCostMatrixFunc).Methods("GET")
	router.HandleFunc("/route_matrix", serveRouteMatrixFunc).Methods("GET")
	router.Handle("/debug/vars", expvar.Handler())
	router.HandleFunc("/relay_dashboard", transport.RelayDashboardHandlerFunc(relayMap, getRouteMatrixFunc, statsdb, "local", "local", maxJitter))

	// if featureConfig.FeatureEnabled(config.FEATURE_NEW_RELAY_BACKEND) {
	// 	// new backend handlers
	// 	router.HandleFunc("/relay_update", transport.NewRelayUpdateHandlerFunc(logger, relayslogger, &commonUpdateParams)).Methods("POST")
	// } else {
	// 	// old backend handlers
	// 	router.HandleFunc("/relay_init", transport.RelayInitHandlerFunc(logger, &commonInitParams)).Methods("POST")
	// 	router.HandleFunc("/relay_update", transport.RelayUpdateHandlerFunc(logger, relayslogger, &commonUpdateParams)).Methods("POST")
	// }

	if featureConfig.FeatureEnabled(config.FEATURE_ENABLE_PPROF) {
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

	go func() {
		port := envvar.Get("PORT", "30000")

		level.Info(logger).Log("addr", ":"+port)

		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	return 0
}

func GCStoreConnect(ctx context.Context, gcpProjectID string) (*gcStorage.BucketHandle, error) {
	client, err := gcStorage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	bkt := client.Bucket(fmt.Sprintf("%s-matrices", gcpProjectID))
	err = bkt.Create(ctx, gcpProjectID, nil)
	if err != nil {
		return nil, err
	}
	return bkt, nil
}

func GCStoreMatrix(bkt *gcStorage.BucketHandle, matrixType string, timestamp time.Time, matrix []byte) error {
	dir := fmt.Sprintf("matrix/relay-backend/0/%d/%d/%d/%d/%d/%s-%d", timestamp.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(), timestamp.Minute(), matrixType, timestamp.Second())
	obj := bkt.Object(dir)
	writer := obj.NewWriter(context.Background())
	defer writer.Close()
	_, err := writer.Write(matrix)
	return err
}

var relayArray_internal []routing.Relay
var relayHash_internal map[uint64]routing.Relay

func ParseAddress(input string) *net.UDPAddr {
	address := &net.UDPAddr{}
	ip_string, port_string, err := net.SplitHostPort(input)
	if err != nil {
		address.IP = net.ParseIP(input)
		address.Port = 0
		return address
	}
	address.IP = net.ParseIP(ip_string)
	address.Port, _ = strconv.Atoi(port_string)
	return address
}

func init() {

	relayHash_internal = make(map[uint64]routing.Relay)

	filePath := envvar.Get("RELAYS_BIN_PATH", "./relays.bin")
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("could not load relay binary: %s\n", filePath)
		return
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&relayArray_internal)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sort.SliceStable(relayArray_internal, func(i, j int) bool {
		return relayArray_internal[i].Name < relayArray_internal[j].Name
	})

	// todo: hack override for local testing
	relayArray_internal[0].Addr = *ParseAddress("127.0.0.1:35000")
	relayArray_internal[0].ID = 0xde0fb1e9a25b1948

	for i := range relayArray_internal {
		relayHash_internal[relayArray_internal[i].ID] = relayArray_internal[i]
	}

	fmt.Printf("\n=======================================\n")
	fmt.Printf("\nLoaded %d relays:\n\n", len(relayArray_internal))
	for i := range relayArray_internal {
		fmt.Printf("    %s - %s [%x]\n", relayArray_internal[i].Name, relayArray_internal[i].Addr.String(), relayArray_internal[i].ID)
	}
	fmt.Printf("\n=======================================\n")
}

func GetRelayData() ([]routing.Relay, map[uint64]routing.Relay) {
	return relayArray_internal, relayHash_internal
}
