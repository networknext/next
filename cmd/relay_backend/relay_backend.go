/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
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
	"sync"
	"syscall"
	"time"

	gcStorage "cloud.google.com/go/storage"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string

	binCreator          string
	binCreationTime     string
	overlayCreator      string
	overlayCreationTime string
	env                 string

	database_internal *routing.DatabaseBinWrapper = routing.CreateEmptyDatabaseBinWrapper()
	overlay_internal  *routing.OverlayBinWrapper  = routing.CreateEmptyOverlayBinWrapper()

	relayArray_internal []routing.Relay
	relayHash_internal  map[uint64]routing.Relay

	databaseMutex   sync.RWMutex
	overlayMutex    sync.RWMutex
	relayArrayMutex sync.RWMutex
	relayHashMutex  sync.RWMutex

	startTime time.Time
)

func init() {
	relayHash_internal = make(map[uint64]routing.Relay)

	databaseFilePath := envvar.Get("BIN_PATH", "./database.bin")
	databaseFile, err := os.Open(databaseFilePath)
	if err != nil {
		fmt.Printf("could not load database binary: %s\n", databaseFilePath)
		return
	}
	defer databaseFile.Close()

	if err = backend.DecodeBinWrapper(databaseFile, database_internal); err != nil {
		core.Error("failed to read database: %v", err)
		os.Exit(1)
	}

	relayArray_internal = database_internal.Relays

	backend.SortAndHashRelayArray(relayArray_internal, relayHash_internal)
	// backend.DisplayLoadedRelays(relayArray_internal)

	binCreator = database_internal.Creator
	binCreationTime = database_internal.CreationTime

	overlayFilePath := envvar.Get("OVERLAY_PATH", "./overlay.bin")
	overlayFile, err := os.Open(overlayFilePath)
	if err != nil {
		fmt.Printf("could not load overlay binary: %s\n", overlayFilePath)
		return
	}
	defer overlayFile.Close()

	if err = backend.DecodeOverlayWrapper(overlayFile, overlay_internal); err != nil {
		core.Error("failed to read overlay: %v", err)
	}
}

func main() {
	os.Exit(mainReturnWithCode())
}

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func mainReturnWithCode() int {
	serviceName := "relay_backend"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime = time.Now().In(est)

	isDebug, err := envvar.GetBool("NEXT_DEBUG", false)
	if err != nil {
		core.Error("Failed to get debug status")
		isDebug = false
	}

	if isDebug {
		core.Debug("Instance is running as a debug instance")
	}

	ctx, ctxCancelFunc := context.WithCancel(context.Background())

	gcpProjectID := backend.GetGCPProjectID()

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		core.Error("failed to get logger: %v", err)
		return 1
	}

	env, err := backend.GetEnv()
	if err != nil {
		core.Error("failed to get env: %v", err)
		return 1
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("failed to get metrics handler: %v", err)
		return 1
	}

	if gcpProjectID != "" {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
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
		core.Error("failed to create relay update metrics: %v", err)
		return 1
	}

	costMatrixMetrics, err := metrics.NewCostMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to create cost matrix metrics: %v", err)
		return 1
	}

	optimizeMetrics, err := metrics.NewOptimizeMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to create optimize metrics metrics: %v", err)
		return 1
	}

	relayBackendMetrics, err := metrics.NewRelayBackendMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to create relay backend metrics: %v", err)
		return 1
	}

	statsdb := routing.NewStatsDatabase()

	// get the max jitter and max packet loss env vars

	if !envvar.Exists("RELAY_ROUTER_MAX_JITTER") {
		core.Error("RELAY_ROUTER_MAX_JITTER not set")
		return 1
	}

	maxJitter, err := envvar.GetFloat("RELAY_ROUTER_MAX_JITTER", 0)
	if err != nil {
		core.Error("failed to parse RELAY_ROUTER_MAX_JITTER %v", err)
		return 1
	}

	if !envvar.Exists("RELAY_ROUTER_MAX_PACKET_LOSS") {
		core.Error("RELAY_ROUTER_MAX_PACKET_LOSS not set")
		return 1
	}

	maxPacketLoss, err := envvar.GetFloat("RELAY_ROUTER_MAX_PACKET_LOSS", 0)
	if err != nil {
		core.Error("failed to parse RELAY_ROUTER_MAX_PACKET_LOSS: %v", err)
		return 1
	}

	if !envvar.Exists("RELAY_ROUTER_MAX_BANDWIDTH_PERCENTAGE") {
		core.Error("RELAY_ROUTER_MAX_BANDWIDTH_PERCENTAGE not set")
		return 1
	}

	maxBandwidthPercentage, err := envvar.GetFloat("RELAY_ROUTER_MAX_BANDWIDTH_PERCENTAGE", 0)
	if err != nil {
		core.Error("failed to parse RELAY_ROUTER_MAX_BANDWIDTH_PERCENTAGE: %v", err)
		return 1
	}

	featureRelayFullBandwidth, err := envvar.GetBool("FEATURE_RELAY_FULL_BANDWIDTH", false)
	if err != nil {
		core.Error("failed to parse FEATURE_RELAY_FULL_BANDWIDTH: %v", err)
		return 1
	}

	instanceID, err := backend.GetInstanceID(env)
	if err != nil {
		core.Error("failed to get relay backend instance ID: %v", err)
		return 1
	}
	core.Debug("VM Instance ID: %s", instanceID)

	// Create an error channel for goroutines
	errChan := make(chan error, 1)

	// Create a waitgroup to manage clean shutdown
	var wg sync.WaitGroup

	// Setup file watchman on database.bin
	{
		// Get absolute path of database.bin
		databaseFilePath := envvar.Get("BIN_PATH", "./database.bin")
		databaseAbsPath, err := filepath.Abs(databaseFilePath)
		if err != nil {
			core.Error("error getting database absolute path %s: %v", databaseFilePath, err)
			return 1
		}

		// Check if file exists
		if _, err := os.Stat(databaseAbsPath); err != nil {
			core.Error("%s does not exist: %v", databaseAbsPath, err)
			return 1
		}

		// Get the directory of the database.bin
		// Used to watch over file creation and modification
		databaseDirectoryPath := filepath.Dir(databaseAbsPath)

		// Get absolute path of database.bin
		overlayFilePath := envvar.Get("OVERLAY_PATH", "./overlay.bin")
		overlayAbsPath, err := filepath.Abs(overlayFilePath)
		if err != nil {
			core.Error("error getting overlay absolute path %s: %v", overlayFilePath, err)
			return 1
		}

		// Check if file exists
		if _, err := os.Stat(overlayAbsPath); err != nil {
			core.Error("%s does not exist: %v", overlayAbsPath, err)
		}

		binSyncInterval, err := envvar.GetDuration("BIN_SYNC_INTERVAL", time.Minute*1)
		if err != nil {
			core.Error("failed to parse BIN_SYNC_INTERVAL: %v", err)
			return 1
		}

		// Setup goroutine to watch for latest database file and update relayArray_internal and relayHash_internal
		wg.Add(1)
		go func() {
			defer wg.Done()

			ticker := time.NewTicker(binSyncInterval)

			core.Debug("started watchman on %s", databaseDirectoryPath)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					// File has changed
					databaseFile, err := os.Open(databaseAbsPath)
					if err != nil {
						core.Error("could not load database binary at %s: %v", databaseAbsPath, err)
						continue
					}

					// Setup relay array and hash to read into
					databaseNew := routing.CreateEmptyDatabaseBinWrapper()

					relayHashNew := make(map[uint64]routing.Relay)

					if err = backend.DecodeBinWrapper(databaseFile, databaseNew); err == io.EOF {
						// Sometimes we receive an EOF error since the file is still being replaced
						// so early out here and proceed on the next notification
						databaseFile.Close()
						core.Debug("DecodeBinWrapper() EOF error, will wait for next notification")
						continue
					} else if err != nil {
						databaseFile.Close()
						core.Error("DecodeBinWrapper() error: %v", err)
						continue
					}

					// Close the file since it is no longer needed
					databaseFile.Close()

					if databaseNew.IsEmpty() {
						// Don't want to use an empty bin wrapper
						// so early out here and use existing array and hash
						core.Error("new database file is empty, keeping previous values")
						continue
					}

					overlayNew := routing.CreateEmptyOverlayBinWrapper()

					// File has changed
					overlayFile, err := os.Open(overlayAbsPath)
					if err != nil {
						core.Error("could not load overlay binary at %s: %v", overlayAbsPath, err)
					} else {
						if err = backend.DecodeOverlayWrapper(overlayFile, overlayNew); err == io.EOF {
							// Sometimes we receive an EOF error since the file is still being replaced
							// so early out here and proceed on the next notification
							core.Debug("DecodeOverlayWrapper() EOF error, will wait for next notification")
						} else if err != nil {
							core.Error("DecodeOverlayWrapper() error: %v", err)
						}

						// Close the file since it is no longer needed
						overlayFile.Close()
					}

					// Only update the internal overlay cache if it is new and not empty
					if !overlayNew.IsEmpty() && overlayNew.CreationTime != overlay_internal.CreationTime {
						overlayMutex.Lock()
						overlay_internal = overlayNew
						overlayMutex.Unlock()
					}

					for _, buyer := range overlay_internal.BuyerMap {
						binBuyer, ok := databaseNew.BuyerMap[buyer.ID]
						// If the buyer does not exist in database.bin or does and is still under trial, use the overlay
						if !ok || (ok && binBuyer.Trial) {
							databaseNew.BuyerMap[buyer.ID] = buyer
						}

						// TODO: Support other buyer driven settings changes here
					}

					// Get the new relay array
					relayArrayNew := databaseNew.Relays
					// Proceed to fill up the new relay hash
					backend.SortAndHashRelayArray(relayArrayNew, relayHashNew)

					// Pointer swap the database bin wrapper
					databaseMutex.Lock()
					database_internal = databaseNew
					binCreator = database_internal.Creator
					binCreationTime = database_internal.CreationTime
					databaseMutex.Unlock()

					// Pointer swap the relay array
					relayArrayMutex.Lock()
					relayArray_internal = relayArrayNew
					relayArrayMutex.Unlock()

					// Pointer swap the relay hash
					relayHashMutex.Lock()
					relayHash_internal = relayHashNew
					relayHashMutex.Unlock()

					// Print the new list of relays
					// backend.DisplayLoadedRelays(relayArray_internal)
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
	wg.Add(1)
	go func() {
		defer wg.Done()
		timeout := int64(routing.RelayTimeout.Seconds())
		frequency := time.Second * 10
		ticker := time.NewTicker(frequency)
		relayMap.TimeoutLoop(ctx, GetRelayData, timeout, ticker.C)
	}()

	var gcBucket *gcStorage.BucketHandle
	gcStoreActive, err := envvar.GetBool("FEATURE_MATRIX_CLOUDSTORE", false)
	if err != nil {
		core.Error("failed to parse FEATURE_MATRIX_CLOUDSTORE: %v", err)
	}
	if gcStoreActive {
		gcBucket, err = GCStoreConnect(ctx, gcpProjectID)
		if err != nil {
			core.Error("failed to connect to GCStore: %v", err)
		}
	}

	syncInterval, err := envvar.GetDuration("COST_MATRIX_INTERVAL", time.Second)
	if err != nil {
		core.Error("failed to parse COST_MATRIX_INTERVAL: %v", err)
		return 1
	}

	matrixBufferSize, err := envvar.GetInt("MATRIX_BUFFER_SIZE", 100000)
	if err != nil {
		core.Error("failed to parse MATRIX_BUFFER_SIZE: %v", err)
		return 1
	}

	port := envvar.Get("PORT", "30001")
	if port == "" {
		core.Error("PORT not set")
		return 1
	}

	// Setup redis so that the Relay Frontend knows this backend is live
	var matrixStore *storage.RedisMatrixStore
	var backendLiveData storage.RelayBackendLiveData
	{
		// Determine which relay backend address this instance is
		backendAddresses := envvar.GetList("RELAY_BACKEND_ADDRESSES", []string{})
		if len(backendAddresses) == 0 {
			core.Error("RELAY_BACKEND_ADDRESSES not set")
			return 1
		}
		foundAddress, backendAddress, err := getBackendAddress(backendAddresses, env)
		if err != nil {
			core.Error("error searching through list of backend addresses: %v", err)
			return 1
		}
		if !foundAddress {
			core.Error("relay backend address not found in list %+v", backendAddresses)
			return 1
		}

		matrixStoreAddress := envvar.Get("MATRIX_STORE_ADDRESS", "")
		if matrixStoreAddress == "" {
			core.Error("MATRIX_STORE_ADDRESS not set")
			return 1
		}

		matrixStorePassword := envvar.Get("MATRIX_STORE_PASSWORD", "")

		maxIdleConnections, err := envvar.GetInt("MATRIX_STORE_MAX_IDLE_CONNS", 5)
		if err != nil {
			core.Error("failed to parse MATRIX_STORE_MAX_IDLE_CONNS: %v", err)
			return 1
		}

		maxActiveConnections, err := envvar.GetInt("MATRIX_STORE_MAX_ACTIVE_CONNS", 5)
		if err != nil {
			core.Error("failed to parse MATRIX_STORE_MAX_ACTIVE_CONNS: %v", err)
			return 1
		}

		readTimeout, err := envvar.GetDuration("MATRIX_STORE_READ_TIMEOUT", 250*time.Millisecond)
		if err != nil {
			core.Error("failed to parse MATRIX_STORE_READ_TIMEOUT: %v", err)
			return 1
		}

		writeTimeout, err := envvar.GetDuration("MATRIX_STORE_WRITE_TIMEOUT", 250*time.Millisecond)
		if err != nil {
			core.Error("failed to parse MATRIX_STORE_WRITE_TIMEOUT: %v", err)
			return 1
		}

		expireTimeout, err := envvar.GetDuration("MATRIX_STORE_EXPIRE_TIMEOUT", 5*time.Second)
		if err != nil {
			core.Error("failed to parse MATRIX_STORE_EXPIRE_TIMEOUT: %v", err)
			return 1
		}

		matrixStore, err = storage.NewRedisMatrixStore(matrixStoreAddress, matrixStorePassword, maxIdleConnections, maxActiveConnections, readTimeout, writeTimeout, expireTimeout)
		if err != nil {
			core.Error("failed to create redis matrix store: %v", err)
			return 1
		}

		backendLiveData.ID = instanceID
		backendLiveData.Address = fmt.Sprintf("%s:%s", backendAddress, port)
		backendLiveData.InitAt = time.Now().UTC()
	}

	var costMatrixData []byte
	var routeMatrixData []byte
	var relaysData []byte
	var destRelaysData []byte

	costMatrix := &routing.CostMatrix{}
	routeMatrix := &routing.RouteMatrix{}

	var costMatrixMutex sync.RWMutex
	var routeMatrixMutex sync.RWMutex
	var relaysMutex sync.RWMutex
	var destRelaysMutex sync.RWMutex

	_ = costMatrix

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(syncInterval)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Encode the current database.bin to attach to route matrix
				databaseMutex.RLock()
				databaseCopy := database_internal
				databaseMutex.RUnlock()

				var databaseBuffer bytes.Buffer
				encoder := gob.NewEncoder(&databaseBuffer)
				encoder.Encode(databaseCopy)

				// build set active relays that are *also* in the current database.bin

				_, relayHash := GetRelayData()

				type ActiveRelayData struct {
					ID                            uint64
					Name                          string
					Addr                          net.UDPAddr
					InternalAddr                  net.UDPAddr
					SessionCount                  int
					Version                       string
					Latitude                      float32
					Longitude                     float32
					SellerID                      string
					DatacenterID                  uint64
					InternalAddressClientRoutable bool
					DestFirst                     bool
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
						relayData.InternalAddr = relay.InternalAddr
						relayData.Name = relay.Name
						relayData.Latitude = float32(relay.Datacenter.Location.Latitude)
						relayData.Longitude = float32(relay.Datacenter.Location.Longitude)
						relayData.SellerID = relay.Seller.ID
						relayData.DatacenterID = relay.Datacenter.ID
						relayData.SessionCount = activeRelaySessionCounts[i]
						relayData.Version = activeRelayVersions[i]
						relayData.InternalAddressClientRoutable = relay.InternalAddressClientRoutable
						relayData.DestFirst = relay.DestFirst

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
				var relayInternalAddressClientRoutable []uint64
				var relayInternalAddressClientRoutableAddresses []net.UDPAddr
				var relayDestFirst []uint64

				for i := range activeRelays {
					relayIDs[i] = activeRelays[i].ID
					relayNames[i] = activeRelays[i].Name
					relayAddresses[i] = activeRelays[i].Addr
					relayLatitudes[i] = float32(activeRelays[i].Latitude)
					relayLongitudes[i] = float32(activeRelays[i].Longitude)
					relayDatacenterIDs[i] = activeRelays[i].DatacenterID

					if activeRelays[i].InternalAddressClientRoutable {
						if activeRelays[i].InternalAddr.String() == ":0" {
							// Do not add this relay as client routable if it is missing an internal address
							core.Error("relay %s (%016x) internal address is client routable but is missing internal address (%s)", activeRelays[i].Name, activeRelays[i].ID, activeRelays[i].InternalAddr.String())
						} else {
							relayInternalAddressClientRoutable = append(relayInternalAddressClientRoutable, activeRelays[i].ID)
							relayInternalAddressClientRoutableAddresses = append(relayInternalAddressClientRoutableAddresses, activeRelays[i].InternalAddr)
						}
					}

					if activeRelays[i].DestFirst {
						relayDestFirst = append(relayDestFirst, activeRelays[i].ID)
					}
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

				var destRelays []bool = make([]bool, len(activeRelays))

				buyers := databaseCopy.BuyerMap
				buyerDCMaps := databaseCopy.DatacenterMaps

				relayIDsToIndices := make(map[uint64]int32)
				numRelays := uint32(len(relayIDs))

				for i := uint32(0); i < numRelays; i++ {
					relayIDsToIndices[relayIDs[i]] = int32(i)
				}

				destRelayNames := []string{}
				// Loop over buyers
				for _, buyer := range buyers {
					// If live check for dest relays
					if buyer.Live {
						for _, dc := range buyerDCMaps[buyer.ID] {
							relaysInDC := routeMatrix.GetDatacenterRelayIDs(dc.DatacenterID)

							if len(relaysInDC) == 0 {
								continue
							}

							for _, relayID := range relaysInDC {
								relayIndex, ok := relayIDsToIndices[relayID]
								if ok {
									destRelays[relayIndex] = true
									destRelayNames = append(destRelayNames, relayNames[relayIndex])
								}
							}
						}
					}
				}

				sort.Strings(destRelayNames)

				destRelaysDataString := ""
				for _, name := range destRelayNames {
					destRelaysDataString += fmt.Sprintf("%s", name)
					destRelaysDataString += fmt.Sprintf("\n")
				}

				destRelaysMutex.Lock()
				destRelaysData = []byte(destRelaysDataString)
				destRelaysMutex.Unlock()

				var costs []int32
				if env == "local" {
					costs = statsdb.GetCostsLocal(relayIDs, float32(maxJitter), float32(maxPacketLoss))
				} else {
					costs = statsdb.GetCosts(relayIDs, float32(maxJitter), float32(maxPacketLoss))
				}

				costMatrixNew := routing.CostMatrix{
					RelayIDs:           relayIDs,
					RelayAddresses:     relayAddresses,
					RelayNames:         relayNames,
					RelayLatitudes:     relayLatitudes,
					RelayLongitudes:    relayLongitudes,
					RelayDatacenterIDs: relayDatacenterIDs,
					Costs:              costs,
					Version:            routing.CostMatrixSerializeVersion,
					DestRelays:         destRelays,
				}

				costMatrixDurationSince := time.Since(costMatrixDurationStart)
				costMatrixMetrics.DurationGauge.Set(float64(costMatrixDurationSince.Milliseconds()))
				if costMatrixDurationSince.Seconds() > 1.0 {
					costMatrixMetrics.LongUpdateCount.Add(1)
				}

				if err := costMatrixNew.WriteResponseData(matrixBufferSize); err != nil {
					core.Error("could not write response data for cost matrix: %v", err)
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

				routeEntries := core.Optimize2(numActiveRelays, numSegments, costMatrixNew.Costs, costThreshold, relayDatacenterIDs, costMatrixNew.DestRelays)

				optimizeDurationSince := time.Since(optimizeDurationStart)
				optimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

				if optimizeDurationSince.Seconds() > 1.0 {
					optimizeMetrics.LongUpdateCount.Add(1)
				}

				pingStats := statsdb.ExtractPingStats(float32(maxJitter), float32(maxPacketLoss), instanceID, isDebug)

				allRelayData := relayMap.GetAllRelayData()
				entries := make([]analytics.RelayStatsEntry, len(allRelayData))

				var fullRelayIDs []uint64

				count := 0
				for i := range allRelayData {
					relay := &allRelayData[i]

					numSessions := relay.SessionCount

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

					if relay.NICSpeedMbps > 0 {
						bwSentPercent = relay.BandwidthSentMbps / float32(relay.NICSpeedMbps) * 100.0
						bwRecvPercent = relay.BandwidthRecvMbps / float32(relay.NICSpeedMbps) * 100.0

						if relay.EnvelopeUpMbps > 0 {
							envSentPercent = relay.BandwidthSentMbps / relay.EnvelopeUpMbps * 100.0
							envRecvPercent = relay.BandwidthRecvMbps / relay.EnvelopeDownMbps * 100.0
						}
					}

					// Track the relays that are near max capacity based on max sessions and bandwidth
					var full bool

					// Relays with MaxSessions set to 0 are never considered full based on session count
					maxSessions := int(relay.MaxSessions)
					if maxSessions != 0 && numSessions >= maxSessions {
						fullRelayIDs = append(fullRelayIDs, relay.ID)
						full = true
						core.Debug("Relay ID %016x is full (%d/%d sessions)", relay.ID, numSessions, maxSessions)
					}

					// Relays with MaxBandwidthMbps set to 0 use maxBandwidthPercentage by default to determine if full
					if featureRelayFullBandwidth && !full {
						if relay.MaxBandwidthMbps != 0 {
							if relay.BandwidthSentMbps > float32(relay.MaxBandwidthMbps) || relay.BandwidthRecvMbps > float32(relay.MaxBandwidthMbps) {
								fullRelayIDs = append(fullRelayIDs, relay.ID)
								full = true
								core.Debug("Relay ID %016x is full (BW Sent Mbps: %.2f | BW Recv Mbps: %.2f | Max BW Mbps: %d)", relay.ID, relay.BandwidthSentMbps, relay.BandwidthRecvMbps, relay.MaxBandwidthMbps)
							}
						} else if float64(bwSentPercent) > maxBandwidthPercentage || float64(bwRecvPercent) > maxBandwidthPercentage {
							fullRelayIDs = append(fullRelayIDs, relay.ID)
							full = true
							core.Debug(`Relay ID %016x is full (BW Sent Percent: %.2f | BW Recv Percent: %.2f | Max BW Percent: %.2f)`, relay.ID, bwSentPercent, bwRecvPercent, maxBandwidthPercentage)
						}
					}

					entries[count] = analytics.RelayStatsEntry{
						ID:                       relay.ID,
						MaxSessions:              relay.MaxSessions,
						NumSessions:              uint32(numSessions),
						NumRoutable:              numRouteable,
						NumUnroutable:            uint32(len(allRelayData)) - 1 - numRouteable,
						Timestamp:                uint64(time.Now().Unix()),
						Full:                     full,
						CPUUsage:                 float32(relay.CPU),
						BandwidthSentPercent:     bwSentPercent,
						BandwidthReceivedPercent: bwRecvPercent,
						EnvelopeSentPercent:      envSentPercent,
						EnvelopeReceivedPercent:  envRecvPercent,
						BandwidthSentMbps:        relay.BandwidthSentMbps,
						BandwidthReceivedMbps:    relay.BandwidthRecvMbps,
						EnvelopeSentMbps:         relay.EnvelopeUpMbps,
						EnvelopeReceivedMbps:     relay.EnvelopeDownMbps,
					}

					count++
				}

				relayStats := entries[:count]

				routeMatrixNew := routing.RouteMatrix{
					RelayIDs:                              relayIDs,
					RelayAddresses:                        relayAddresses,
					RelayNames:                            relayNames,
					RelayLatitudes:                        relayLatitudes,
					RelayLongitudes:                       relayLongitudes,
					RelayDatacenterIDs:                    relayDatacenterIDs,
					RouteEntries:                          routeEntries,
					BinFileBytes:                          int32(len(databaseBuffer.Bytes())),
					BinFileData:                           databaseBuffer.Bytes(),
					CreatedAt:                             uint64(time.Now().Unix()),
					Version:                               routing.RouteMatrixSerializeVersion,
					DestRelays:                            destRelays,
					PingStats:                             pingStats,
					RelayStats:                            relayStats,
					FullRelayIDs:                          fullRelayIDs,
					InternalAddressClientRoutableRelayIDs: relayInternalAddressClientRoutable,
					InternalAddressClientRoutableRelayAddresses: relayInternalAddressClientRoutableAddresses,
					DestFirstRelayIDs: relayDestFirst,
				}

				if err := routeMatrixNew.WriteResponseData(matrixBufferSize); err != nil {
					core.Error("could not write response data for route matrix: %v", err)
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

				// update status data for "/status" handler

				numRoutes := int32(0)
				for i := range routeMatrixNew.RouteEntries {
					numRoutes += routeMatrixNew.RouteEntries[i].NumRoutes
				}
				relayBackendMetrics.RouteMatrix.RouteCount.Set(float64(numRoutes))

				// Update redis with last update time
				// Debug instance should not store this data in redis
				if !isDebug {
					backendLiveData.UpdatedAt = time.Now().UTC()
					err = matrixStore.SetRelayBackendLiveData(backendLiveData)
					if err != nil {
						core.Error("failed to set relay backend live data for address %s: %v", backendLiveData.Address, err)
					}
				}

				// optionally write route matrix to cloud storage
				if gcStoreActive {
					if gcBucket == nil {
						gcBucket, err = GCStoreConnect(ctx, gcpProjectID)
						if err != nil {
							core.Error("failed to connect to GC Store: %v", err)
							continue
						}
					}

					timestamp := time.Now().UTC()
					err = GCStoreMatrix(gcBucket, "cost", timestamp, costMatrixNew.GetResponseData())
					if err != nil {
						core.Error("failed to write cost matrix to GC Storage: %v", err)
						continue
					}
					err = GCStoreMatrix(gcBucket, "route", timestamp, routeMatrixNew.GetResponseData())
					if err != nil {
						core.Error("failed to write route matrix to GC Storage: %v", err)
						continue
					}
				}
			}
		}
	}()

	// Setup the status handler info

	statusData := &metrics.RelayBackendStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				relayBackendMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				relayBackendMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.RelayBackendStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				// Service Metrics
				newStatusData.Goroutines = int(relayBackendMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = relayBackendMetrics.MemoryAllocated.Value()

				// Relay Information
				newStatusData.DatacenterCount = int(relayBackendMetrics.RouteMatrix.DatacenterCount.Value())
				newStatusData.RelayCount = int(relayBackendMetrics.RouteMatrix.RelayCount.Value())
				newStatusData.RouteCount = int(relayBackendMetrics.RouteMatrix.RouteCount.Value())

				// Relay Update Information
				newStatusData.RelayUpdateInvocations = int(relayUpdateMetrics.Invocations.Value())
				newStatusData.RelayUpdateContentTypeFailure = int(relayUpdateMetrics.ErrorMetrics.ContentTypeFailure.Value())
				newStatusData.RelayUpdateUnbatchFailure = int(relayUpdateMetrics.ErrorMetrics.UnbatchFailure.Value())
				newStatusData.RelayUpdateUnmarshalFailure = int(relayUpdateMetrics.ErrorMetrics.UnmarshalFailure.Value())
				newStatusData.RelayUpdateRelayNotFound = int(relayUpdateMetrics.ErrorMetrics.RelayNotFound.Value())

				// Durations
				newStatusData.LongCostMatrixUpdates = int(costMatrixMetrics.LongUpdateCount.Value())
				newStatusData.LongRouteMatrixUpdates = int(optimizeMetrics.LongUpdateCount.Value())
				newStatusData.CostMatrixUpdateMs = costMatrixMetrics.DurationGauge.Value()
				newStatusData.RouteMatrixUpdateMs = optimizeMetrics.DurationGauge.Value()
				newStatusData.RelayUpdateMs = relayUpdateMetrics.DurationGauge.Value()

				// Size
				newStatusData.CostMatrixBytes = int(costMatrixMetrics.Bytes.Value())
				newStatusData.RouteMatrixBytes = int(relayBackendMetrics.RouteMatrix.Bytes.Value())

				statusMutex.Lock()
				statusData = newStatusData
				statusMutex.Unlock()

				time.Sleep(time.Second * 10)
			}
		}()
	}

	commonUpdateParams := transport.RelayUpdateHandlerConfig{
		RelayMap:     relayMap,
		StatsDB:      statsdb,
		Metrics:      relayUpdateMetrics,
		GetRelayData: GetRelayData,
	}

	destRelayFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		destRelaysMutex.RLock()
		data := destRelaysData
		destRelaysMutex.RUnlock()
		buffer := bytes.NewBuffer(data)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
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

	serveStatusFunc := func(w http.ResponseWriter, r *http.Request) {
		statusMutex.RLock()
		data := statusData
		statusMutex.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			core.Error("could not write status data to json: %v\n%+v", err, data)
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

	fmt.Printf("starting http server on port %s\n\n", port)

	router := mux.NewRouter()

	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/database_version", transport.DatabaseBinVersionFunc(&binCreator, &binCreationTime, &env))
	router.HandleFunc("/relay_update", transport.RelayUpdateHandlerFunc(&commonUpdateParams)).Methods("POST")
	router.HandleFunc("/route_matrix", serveRouteMatrixFunc).Methods("GET")
	router.HandleFunc("/relay_dashboard_data", transport.RelayDashboardDataHandlerFunc(relayMap, getRouteMatrixFunc, statsdb, maxJitter))
	router.HandleFunc("/relay_dashboard_analysis", transport.RelayDashboardAnalysisHandlerFunc(getRouteMatrixFunc))
	router.HandleFunc("/status", serveStatusFunc).Methods("GET")
	router.HandleFunc("/dest_relays", destRelayFunc).Methods("GET")
	router.Handle("/debug/vars", expvar.Handler())

	router.HandleFunc("/relays", serveRelaysFunc)
	router.HandleFunc("/cost_matrix", serveCostMatrixFunc)

	enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
	if err != nil {
		core.Error("failed to parse FEATURE_ENABLE_PPROF: %v", err)
	}
	if enablePProf {
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

	go func() {
		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			core.Error("error starting http server: %v", err)
			errChan <- err
		}
	}()

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan: // Exit with an error code of 0 if we receive SIGINT or SIGTERM
		fmt.Println("Received shutdown signal.")

		ctxCancelFunc()
		// Wait for essential goroutines to finish up
		wg.Wait()

		fmt.Println("Successfully shutdown.")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		// Still let essential goroutines finish even though we got an error
		ctxCancelFunc()
		wg.Wait()
		return 1
	}
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

func GetRelayData() ([]routing.Relay, map[uint64]routing.Relay) {
	relayArrayMutex.RLock()
	relayArrayData := relayArray_internal
	relayArrayMutex.RUnlock()

	relayHashMutex.RLock()
	relayHashData := relayHash_internal
	relayHashMutex.RUnlock()

	return relayArrayData, relayHashData
}

// Determines if this instance is in the backend address list and
// gets the backend address
func getBackendAddress(backendAddresses []string, env string) (bool, string, error) {
	var host string
	var err error

	if env == "local" {
		// Running local env, default IP to 127.0.0.1
		host = "127.0.0.1"
	} else {
		// Get the host
		host, err = os.Hostname()
		if err != nil {
			return false, "", err
		}
	}

	// Get a list of IPv4 and IPv6 addresses for the host
	addresses, err := net.LookupIP(host)
	if err != nil {
		return false, "", err
	}

	// Get the hosts from the backend addresses if local
	var backendAddressHosts []string
	if env == "local" {
		for _, address := range backendAddresses {
			backendHost, _, err := net.SplitHostPort(address)
			if err != nil {
				return false, "", err
			}
			backendAddressHosts = append(backendAddressHosts, backendHost)
		}
	} else {
		backendAddressHosts = backendAddresses
	}

	for _, address := range addresses {
		// Get the IPv4 of the address
		if ipv4 := address.To4(); ipv4 != nil {
			// Search through the list to see if there's a match
			for _, validAddress := range backendAddressHosts {
				if ipv4.String() == validAddress {
					return true, ipv4.String(), nil
				}
			}
		}
	}

	return false, "", nil
}
