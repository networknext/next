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
	"hash/fnv"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/common/helpers"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/backend" // todo: not a good name for a module
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
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	ctx := context.Background()

	gcpProjectID := backend.GetGCPProjectID()

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	relayslogger, err := backend.GetLogger(ctx, gcpProjectID, "relays")
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

	// relay stats
	var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
	{
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

					fsrelay, err := storer.Relay(relay.ID)
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

	var relayNamesHashPublisher analytics.RouteMatrixStatsPublisher = &analytics.NoOpRouteMatrixStatsPublisher{}
	{
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

	relayEnabledCache := common.NewRelayEnabledCache(storer)
	relayEnabledCache.StartRunner(1 * time.Minute)

	var gcBucket *gcStorage.BucketHandle
	gcStoreActive, err := envvar.GetBool("FEATURE_MATRIX_CLOUDSTORE", false)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	if gcStoreActive {
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

	matrixBufferSize, err := envvar.GetInt("MATRIX_BUFFER_SIZE", 100000)
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
			// For now, exclude all valve relays
			baseRelayIDs := relayMap.GetAllRelayIDs([]string{"valve"}) // Filter out any relays whose seller has a Firestore key of "valve"

			namesMap := make(map[string]routing.Relay)
			numRelays := len(baseRelayIDs)
			relayAddresses := make([]net.UDPAddr, numRelays)
			relayNames := make([]string, numRelays)
			relayLatitudes := make([]float32, numRelays)
			relayLongitudes := make([]float32, numRelays)
			relayDatacenterIDs := make([]uint64, numRelays)
			relayIDs := make([]uint64, numRelays)

			for i, relayID := range baseRelayIDs {
				relay, err := storer.Relay(relayID)
				if err != nil {
					continue
				}

				relayNames[i] = relay.Name
				namesMap[relay.Name] = relay
			}
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
				level.Warn(logger).Log("matrix", "cost", "op", "optimize", "warn", "no route entries generated from cost matrix")
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

			gcStoreActive, err := envvar.GetBool("FEATURE_MATRIX_CLOUDSTORE", false)
			if err != nil {
				level.Error(logger).Log("err", err)
				continue
			}
			if gcStoreActive {
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

			hashing, err := envvar.GetBool("FEATURE_ROUTE_MATRIX_STATS", true)
			if err != nil {
				level.Error(logger).Log("err", err)
			}
			if hashing {
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

	// Separate route matrix specifically for Valve
	valveMatrixData := new(helpers.MatrixData)

	// Generate the route matrix specifically for valve
	go func() {
		syncTimer := helpers.NewSyncTimer(syncInterval)
		for {
			syncTimer.Run()
			// All relays included
			baseRelayIDs := relayMap.GetAllRelayIDs([]string{})

			namesMap := make(map[string]routing.Relay)
			numRelays := len(baseRelayIDs)
			relayAddresses := make([]net.UDPAddr, numRelays)
			relayNames := make([]string, numRelays)
			relayLatitudes := make([]float32, numRelays)
			relayLongitudes := make([]float32, numRelays)
			relayDatacenterIDs := make([]uint64, numRelays)
			relayIDs := make([]uint64, numRelays)

			for i, relayID := range baseRelayIDs {
				relay, err := storer.Relay(relayID)
				if err != nil {
					continue
				}

				relayNames[i] = relay.Name
				namesMap[relay.Name] = relay
			}
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

			valveCostMatrixMetrics.Invocations.Add(1)
			costMatrixDurationStart := time.Now()

			valveCostMatrix := statsdb.GetCosts(relayIDs, float32(maxJitter), float32(maxPacketLoss))

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
				continue
			}

			optimizeDurationSince := time.Since(optimizeDurationStart)
			valveOptimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

			if optimizeDurationSince.Seconds() > 1.0 {
				valveOptimizeMetrics.LongUpdateCount.Add(1)
			}

			valveRouteMatrixNew := &routing.RouteMatrix{
				RelayIDs:           relayIDs,
				RelayAddresses:     relayAddresses,
				RelayNames:         relayNames,
				RelayLatitudes:     relayLatitudes,
				RelayLongitudes:    relayLongitudes,
				RelayDatacenterIDs: relayDatacenterIDs,
				RouteEntries:       routeEntries,
			}

			if err := valveRouteMatrixNew.WriteResponseData(matrixBufferSize); err != nil {
				level.Error(logger).Log("matrix", "route", "op", "write_response", "msg", "could not write response data", "err", err)
				continue
			}

			valveRouteMatrixNew.WriteAnalysisData()

			valveMatrixData.SetMatrix(valveRouteMatrixNew.GetResponseData())

			valveRouteMatrixMetrics.Bytes.Set(float64(valveMatrixData.GetMatrixDataSize()))
			valveRouteMatrixMetrics.RelayCount.Set(float64(len(valveRouteMatrixNew.RelayIDs)))
			valveRouteMatrixMetrics.DatacenterCount.Set(float64(len(valveRouteMatrixNew.RelayDatacenterIDs)))

			// todo: calculate this in optimize and store in route matrix so we don't have to calc this here
			numRoutes := int32(0)
			for i := range valveRouteMatrixNew.RouteEntries {
				numRoutes += valveRouteMatrixNew.RouteEntries[i].NumRoutes
			}
			valveRouteMatrixMetrics.RouteCount.Set(float64(numRoutes))
		}
	}()

	commonInitParams := transport.RelayInitHandlerConfig{
		RelayMap:         relayMap,
		Storer:           storer,
		Metrics:          relayInitMetrics,
		RouterPrivateKey: routerPrivateKey,
	}

	commonUpdateParams := transport.RelayUpdateHandlerConfig{
		RelayMap: relayMap,
		StatsDB:  statsdb,
		Metrics:  relayUpdateMetrics,
		Storer:   storer,
	}

	serveRouteMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		buffer := bytes.NewBuffer(routeMatrixData.GetMatrix())
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	serveValveRouteMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		buffer := bytes.NewBuffer(valveMatrixData.GetMatrix())
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

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, false, []string{}))
	router.HandleFunc("/relay_init", transport.RelayInitHandlerFunc(logger, &commonInitParams)).Methods("POST")
	router.HandleFunc("/relay_update", transport.RelayUpdateHandlerFunc(logger, relayslogger, &commonUpdateParams)).Methods("POST")
	router.HandleFunc("/cost_matrix", serveCostMatrixFunc).Methods("GET")
	router.HandleFunc("/route_matrix", serveRouteMatrixFunc).Methods("GET")
	router.HandleFunc("/route_matrix_valve", serveValveRouteMatrixFunc).Methods("GET")
	router.Handle("/debug/vars", expvar.Handler())
	router.HandleFunc("/relay_dashboard", transport.RelayDashboardHandlerFunc(relayMap, getRouteMatrixFunc, statsdb, "local", "local", maxJitter))
	router.HandleFunc("/relay_stats", transport.RelayStatsFunc(logger, relayMap))

	enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	if enablePProf {
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
