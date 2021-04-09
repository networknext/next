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
	"io"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	gcStorage "cloud.google.com/go/storage"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/rjeczalik/notify"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string

	author    string
	timestamp string
	env       string

	relayArray_internal []routing.Relay
	relayHash_internal  map[uint64]routing.Relay

	relayArrayMutex sync.RWMutex
	relayHashMutex  sync.RWMutex
)

func init() {
	var binWrapper routing.RelayBinWrapper
	relayHash_internal = make(map[uint64]routing.Relay)

	filePath := envvar.Get("RELAYS_BIN_PATH", "./relays.bin")
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("could not load relay binary: %s\n", filePath)
		return
	}
	defer file.Close()

	if err = decodeBinWrapper(file, &binWrapper); err != nil {
		fmt.Printf("decodeBinWrapper() error: %v\n", err)
		os.Exit(1)
	}

	relayArray_internal = binWrapper.Relays

	gcpProjectID := backend.GetGCPProjectID()
	sortAndHashRelayArray(relayArray_internal, relayHash_internal, gcpProjectID)
	displayLoadedRelays(relayArray_internal)

	// TODO: update the author, timestamp, and env for the RelaysBinVersionFunc handler using the other fields in binWrapper
}

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

	/*
		routerPrivateKey, err := envvar.GetBase64("RELAY_ROUTER_PRIVATE_KEY", nil)
		if err != nil {
			level.Error(logger).Log("err", "RELAY_ROUTER_PRIVATE_KEY not set")
			return 1
		}
	*/

	// create metrics

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

	// get the max jitter and max packet loss env vars

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

	// Setup file watchman on relays.bin
	{
		// Get absolute path of relays.bin
		relaysFilePath := envvar.Get("RELAYS_BIN_PATH", "./relays.bin")
		absPath, err := filepath.Abs(relaysFilePath)
		if err != nil {
			level.Error(logger).Log("msg", fmt.Sprintf("error getting absolute path %s", relaysFilePath), "err", err)
			return 1
		}

		// Check if file exists
		if _, err := os.Stat(absPath); err != nil {
			level.Error(logger).Log("msg", fmt.Sprintf("%s does not exist", absPath), "err", err)
			return 1
		}

		// Get the directory of the relays.bin
		// Used to watch over file creation and modification
		directoryPath := filepath.Dir(absPath)

		// Create channel to store notifications of file changing
		// Use a size of 2 to ensure a change is detected even with io.EOF error
		fileChan := make(chan notify.EventInfo, 2)

		// Check for Create and InModify events because cloud scheduler will be deleting and replacing each file with updates
		// Create covers the case where a file is explicitly deleted and then re-added
		// InModify covers the case where a file is replaced with the same filename (i.e. using mv or cp)
		if err := notify.Watch(directoryPath, fileChan, notify.Create, notify.InModify); err != nil {
			level.Error(logger).Log("msg", fmt.Sprintf("could not create watchman on %s", directoryPath), "err", err)
			return 1
		}
		defer notify.Stop(fileChan)

		// Setup goroutine to watch for replaced file and update relayArray_internal and relayHash_internal
		go func() {
			level.Debug(logger).Log("msg", fmt.Sprintf("started watchman on %s", directoryPath))
			for {
				select {
				case ei := <-fileChan:
					if strings.Contains(ei.Path(), absPath) {
						// File has changed
						level.Debug(logger).Log("msg", fmt.Sprintf("detected file change type %s at %s", ei.Event().String(), ei.Path()))
						file, err := os.Open(absPath)
						if err != nil {
							level.Error(logger).Log("msg", fmt.Sprintf("could not load relay binary at %s", absPath), "err", err)
							continue
						}

						// Setup relay array and hash to read into
						var binWrapperNew routing.RelayBinWrapper
						relayHashNew := make(map[uint64]routing.Relay)

						if err = decodeBinWrapper(file, &binWrapperNew); err == io.EOF {
							// Sometimes we receive an EOF error since the file is still being replaced
							// so early out here and proceed on the next notification
							file.Close()
							level.Debug(logger).Log("msg", "decodeBinWrapper() EOF error, will wait for next notification")
							continue
						} else if err != nil {
							file.Close()
							level.Error(logger).Log("msg", "decodeBinWrapper() error", "err", err)
							continue
						}

						// Close the file since it is no longer needed
						file.Close()

						// Get the new relay array
						relayArrayNew := binWrapperNew.Relays
						// Proceed to fill up the new relay hash
						sortAndHashRelayArray(relayArrayNew, relayHashNew, gcpProjectID)

						// Pointer swap the relay array
						relayArrayMutex.Lock()
						relayArray_internal = relayArrayNew
						relayArrayMutex.Unlock()

						// Pointer swap the relay hash
						relayHashMutex.Lock()
						relayHash_internal = relayHashNew
						relayHashMutex.Unlock()

						// TODO: update the author, timestamp, and env for the RelaysBinVersionFunc handler using the other fields in binWrapperNew
						level.Debug(logger).Log("msg", "successfully updated the relay array and hash")

						// Print the new list of relays
						displayLoadedRelays(relayArray_internal)
					}
				}
			}
		}()
	}

	// Create the relay map
	cleanupCallback := func(relayData routing.RelayData) error {
		statsdb.DeleteEntry(relayData.ID)
		return nil
	}

	relayMap := routing.NewRelayMap(cleanupCallback)
	go func() {
		timeout := int64(routing.RelayTimeout.Seconds())
		frequency := time.Second * 10
		ticker := time.NewTicker(frequency)
		relayMap.TimeoutLoop(ctx, GetRelayData, timeout, ticker.C)
	}()

	// relay ping stats

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

	var costMatrixData []byte
	var routeMatrixData []byte
	var relaysData []byte

	costMatrix := &routing.CostMatrix{}
	routeMatrix := &routing.RouteMatrix{}

	var costMatrixMutex sync.RWMutex
	var routeMatrixMutex sync.RWMutex
	var relaysMutex sync.RWMutex

	_ = costMatrix

	go func() {

		syncTimer := helpers.NewSyncTimer(syncInterval)

		for {
			syncTimer.Run()

			// build set active relays that are *also* in the current relays.bin

			_, relayHash := GetRelayData()

			type ActiveRelayData struct {
				ID           uint64
				Name         string
				Addr         net.UDPAddr
				SessionCount int
				Version      string
				Latitude     float32
				Longitude    float32
				SellerID     string
				DatacenterID uint64
			}

			activeRelays := make([]ActiveRelayData, 0)
			{
				activeRelayIds, activeRelaySessionCounts, activeRelayVersions := relayMap.GetActiveRelayData()

				for i := range activeRelayIds {

					id := activeRelayIds[i]
					relay, ok := relayHash[id]
					if !ok {
						continue
					}

					relayData := ActiveRelayData{}
					relayData.ID = relay.ID
					relayData.Addr = relay.Addr
					relayData.Name = relay.Name
					relayData.Latitude = float32(relay.Datacenter.Location.Latitude)
					relayData.Longitude = float32(relay.Datacenter.Location.Longitude)
					relayData.SellerID = relay.Seller.ID
					relayData.DatacenterID = relay.Datacenter.ID
					relayData.SessionCount = activeRelaySessionCounts[i]
					relayData.Version = activeRelayVersions[i]

					activeRelays = append(activeRelays, relayData)
				}
			}

			sort.SliceStable(activeRelays, func(i, j int) bool { return activeRelays[i].Name < activeRelays[j].Name })

			// gather relay data required for building cost matrix

			numActiveRelays := len(activeRelays)

			relayIDs := make([]uint64, numActiveRelays)
			relayAddresses := make([]net.UDPAddr, numActiveRelays)
			relayNames := make([]string, numActiveRelays)
			relayLatitudes := make([]float32, numActiveRelays)
			relayLongitudes := make([]float32, numActiveRelays)
			relayDatacenterIDs := make([]uint64, numActiveRelays)

			for i := range activeRelays {
				relayIDs[i] = activeRelays[i].ID
				relayNames[i] = activeRelays[i].Name
				relayAddresses[i] = activeRelays[i].Addr
				relayLatitudes[i] = float32(activeRelays[i].Latitude)
				relayLongitudes[i] = float32(activeRelays[i].Longitude)
				relayDatacenterIDs[i] = activeRelays[i].DatacenterID
			}

			// build relays data to serve up on "relays" endpoint (CSV)

			// active relays

			relaysDataString := "name,address,id,status,sessions,version"

			for i := range activeRelays {
				name := activeRelays[i].Name
				address := activeRelays[i].Addr.String()
				id := activeRelays[i].ID
				status := "active"
				sessions := activeRelays[i].SessionCount
				version := activeRelays[i].Version
				relaysDataString = fmt.Sprintf("%s\n%s,%s,%x,%s,%d,%s", relaysDataString, name, address, id, status, sessions, version)
			}

			// inactive relays

			inactiveRelays := make([]routing.Relay, 0)

			relayMap.RLock()
			for _, v := range relayHash {
				_, exists := relayMap.GetRelayData(v.Addr.String())
				if !exists {
					inactiveRelays = append(inactiveRelays, v)
				}
			}
			relayMap.RUnlock()

			sort.SliceStable(inactiveRelays, func(i, j int) bool { return inactiveRelays[i].Name < inactiveRelays[j].Name })

			for i := range inactiveRelays {
				name := inactiveRelays[i].Name
				address := inactiveRelays[i].Addr.String()
				id := inactiveRelays[i].ID
				relaysDataString = fmt.Sprintf("%s\n%s,%s,%x,inactive,,", relaysDataString, name, address, id)
			}

			// shutting down relays

			shuttingDownRelays := make([]routing.Relay, 0)

			relayMap.RLock()
			for _, v := range relayHash {
				relayData, exists := relayMap.GetRelayData(v.Addr.String())
				if exists && relayData.ShuttingDown {
					shuttingDownRelays = append(shuttingDownRelays, v)
				}
			}
			relayMap.RUnlock()

			sort.SliceStable(shuttingDownRelays, func(i, j int) bool { return shuttingDownRelays[i].Name < shuttingDownRelays[j].Name })

			for i := range shuttingDownRelays {
				name := shuttingDownRelays[i].Name
				address := shuttingDownRelays[i].Addr.String()
				id := shuttingDownRelays[i].ID
				relaysDataString = fmt.Sprintf("%s\n%s,%s,%x,shutting down,,", relaysDataString, name, address, id)
			}

			relaysMutex.Lock()
			relaysData = []byte(relaysDataString)
			relaysMutex.Unlock()

			// build cost matrix

			costMatrixMetrics.Invocations.Add(1)
			costMatrixDurationStart := time.Now()

			costMatrixNew := routing.CostMatrix{
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

			costMatrixDataNew := costMatrixNew.GetResponseData()

			costMatrixMetrics.Bytes.Set(float64(len(costMatrixDataNew)))

			costMatrixMutex.Lock()
			costMatrix = &costMatrixNew
			costMatrixData = costMatrixDataNew
			costMatrixMutex.Unlock()

			// optimize

			numCPUs := runtime.NumCPU()
			numSegments := numActiveRelays
			if numCPUs < numActiveRelays {
				numSegments = numActiveRelays / 5
				if numSegments == 0 {
					numSegments = 1
				}
			}

			optimizeMetrics.Invocations.Add(1)
			optimizeDurationStart := time.Now()

			costThreshold := int32(1)

			routeEntries := core.Optimize(numActiveRelays, numSegments, costMatrixNew.Costs, costThreshold, relayDatacenterIDs)

			optimizeDurationSince := time.Since(optimizeDurationStart)
			optimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

			if optimizeDurationSince.Seconds() > 1.0 {
				optimizeMetrics.LongUpdateCount.Add(1)
			}

			routeMatrixNew := routing.RouteMatrix{
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

			routeMatrixDataNew := routeMatrixNew.GetResponseData()

			relayBackendMetrics.RouteMatrix.Bytes.Set(float64(len(routeMatrixDataNew)))
			relayBackendMetrics.RouteMatrix.RelayCount.Set(float64(len(routeMatrixNew.RelayIDs)))
			relayBackendMetrics.RouteMatrix.DatacenterCount.Set(float64(len(routeMatrixNew.RelayDatacenterIDs)))

			routeMatrixMutex.Lock()
			routeMatrix = &routeMatrixNew
			routeMatrixData = routeMatrixDataNew
			routeMatrixMutex.Unlock()

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

			// optionally write route matrix to cloud storage

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
		}
	}()

	commonUpdateParams := transport.RelayUpdateHandlerConfig{
		RelayMap:     relayMap,
		StatsDB:      statsdb,
		Metrics:      relayUpdateMetrics,
		GetRelayData: GetRelayData,
	}

	serveRelaysFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		relaysMutex.RLock()
		data := relaysData
		relaysMutex.RUnlock()
		buffer := bytes.NewBuffer(data)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	serveRouteMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		routeMatrixMutex.RLock()
		data := routeMatrixData
		routeMatrixMutex.RUnlock()
		buffer := bytes.NewBuffer(data)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	serveCostMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		costMatrixMutex.RLock()
		data := costMatrixData
		costMatrixMutex.RUnlock()
		buffer := bytes.NewBuffer(data)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	getRouteMatrixFunc := func() *routing.RouteMatrix {
		routeMatrixMutex.RLock()
		rm := routeMatrix
		routeMatrixMutex.RUnlock()
		return rm
	}

	port := envvar.Get("PORT", "30000")

	fmt.Printf("starting http server on port %s\n\n", port)

	router := mux.NewRouter()

	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/bin_version", transport.RelaysBinVersionFunc(author, timestamp, env))
	router.HandleFunc("/relay_update", transport.RelayUpdateHandlerFunc(&commonUpdateParams)).Methods("POST")
	router.HandleFunc("/cost_matrix", serveCostMatrixFunc).Methods("GET")
	router.HandleFunc("/route_matrix", serveRouteMatrixFunc).Methods("GET")
	router.HandleFunc("/relay_dashboard", transport.RelayDashboardHandlerFunc(relayMap, getRouteMatrixFunc, statsdb, "local", "local", maxJitter))
	router.HandleFunc("/relays", serveRelaysFunc).Methods("GET")

	router.Handle("/debug/vars", expvar.Handler())

	enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	if enablePProf {
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

	go func() {
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

func GetRelayData() ([]routing.Relay, map[uint64]routing.Relay) {
	relayArrayMutex.RLock()
	relayArrayData := relayArray_internal
	relayArrayMutex.RUnlock()

	relayHashMutex.RLock()
	relayHashData := relayHash_internal
	relayHashMutex.RUnlock()

	return relayArrayData, relayHashData
}

func decodeBinWrapper(file *os.File, binWrapper *routing.RelayBinWrapper) error {
	decoder := gob.NewDecoder(file)
	err := decoder.Decode(binWrapper)
	return err
}

func sortAndHashRelayArray(relayArray []routing.Relay, relayHash map[uint64]routing.Relay, gcpProjectID string) {
	sort.SliceStable(relayArray, func(i, j int) bool {
		return relayArray[i].Name < relayArray[j].Name
	})

	if gcpProjectID == "" {
		// TODO: hack override for local testing for single relay
		relayArray[0].Addr = *ParseAddress("127.0.0.1:35000")
		relayArray[0].ID = 0xde0fb1e9a25b1948
	}

	for i := range relayArray {
		relayHash[relayArray[i].ID] = relayArray[i]
	}
}

func displayLoadedRelays(relayArray []routing.Relay) {
	fmt.Printf("\n=======================================\n")
	fmt.Printf("\nLoaded %d relays:\n\n", len(relayArray))
	for i := range relayArray {
		fmt.Printf("\t%s - %s [%x]\n", relayArray[i].Name, relayArray[i].Addr.String(), relayArray[i].ID)
	}
	fmt.Printf("\n=======================================\n")
}
