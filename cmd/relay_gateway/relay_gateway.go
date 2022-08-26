/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	gateway "github.com/networknext/backend/modules/relay_gateway"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"

	"github.com/gorilla/mux"
)

var (
	buildTime     string
	commitMessage string
	commitHash    string

	binCreator      string
	binCreationTime string

	relayArray_internal []routing.Relay
	relayHash_internal  map[uint64]routing.Relay

	relayArrayMutex sync.RWMutex
	relayHashMutex  sync.RWMutex

	magicUpcoming_internal []byte
	magicCurrent_internal  []byte
	magicPrevious_internal []byte

	magicMutex sync.RWMutex
)

func initialize() {

	database := routing.CreateEmptyDatabaseBinWrapper()

	relayHash_internal = make(map[uint64]routing.Relay)

	filePath := envvar.Get("BIN_PATH", "./database.bin")
	file, err := os.Open(filePath)
	if err != nil {
		// fmt.Printf("could not load database binary: %s\n", filePath)
		return
	}
	defer file.Close()

	if err = backend.DecodeBinWrapper(file, database); err != nil {
		core.Error("failed to read database: %v", err)
		os.Exit(1)
	}

	fmt.Printf("loaded database.bin\n")

	relayArray_internal = database.Relays

	backend.SortAndHashRelayArray(relayArray_internal, relayHash_internal)

	// backend.DisplayLoadedRelays(relayArray_internal)

	// Store the creator and creation time from the database
	binCreator = database.Creator
	binCreationTime = database.CreationTime

	magicUpcoming_internal = make([]byte, 8)
	magicCurrent_internal = make([]byte, 8)
	magicPrevious_internal = make([]byte, 8)
}

func main() {
	os.Exit(mainReturnWithCode())
}

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func mainReturnWithCode() int {
	
	serviceName := "relay_gateway"

	fmt.Printf("%s\n", serviceName)
	fmt.Printf("commit hash: %s\n", commitHash)
	fmt.Printf("commit message: %s\n", commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	initialize()

	// Setup the service
	ctx, cancel := context.WithCancel(context.Background())

	gcpProjectID := backend.GetGCPProjectID()

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		core.Error("failed to get logger: %v", err)
		return 1
	}

	env := backend.GetEnv()

	if gcpProjectID != "" {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			core.Error("failed to initialze StackDriver profiler: %v", err)
			return 1
		}
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("failed to get metrics handler: %v", err)
		return 1
	}

	gatewayMetrics, err := metrics.NewRelayGatewayMetrics(ctx, metricsHandler, serviceName, "relay_gateway", "Relay Gateway", "relay update request")
	if err != nil {
		core.Error("failed to create relay gateway metrics: %v", err)
		return 1
	}

	// Get a config for how the Gateway should operate
	cfg, err := newConfig()
	if err != nil {
		core.Error("failed to create relay gateway config: %v", err)
		return 1
	}

	// todo: this "file watcher" should be a helper class in a module

	// Setup file watchman on database.bin
	{
		// Get absolute path of database.bin
		databaseFilePath := envvar.Get("BIN_PATH", "./database.bin")
		absPath, err := filepath.Abs(databaseFilePath)
		if err != nil {
			core.Error("error getting absolute path %s: %v", databaseFilePath, err)
			return 1
		}

		// Check if file exists
		if _, err := os.Stat(absPath); err != nil {
			core.Error("%s does not exist: %v", absPath, err)
			return 1
		}

		// Get the directory of the database.bin
		// Used to watch over file creation and modification
		directoryPath := filepath.Dir(absPath)

		ticker := time.NewTicker(cfg.BinSyncInterval)

		// Setup goroutine to watch for replaced file and update relayArray_internal and relayHash_internal
		go func() {
			fmt.Printf("started watchman on %s\n", directoryPath)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					// File has changed
					file, err := os.Open(absPath)
					if err != nil {
						core.Error("could not load database binary at %s: %v", absPath, err)
						continue
					}

					// Setup relay array and hash to read into
					databaseNew := routing.CreateEmptyDatabaseBinWrapper()

					relayHashNew := make(map[uint64]routing.Relay)

					if err = backend.DecodeBinWrapper(file, databaseNew); err == io.EOF {
						// Sometimes we receive an EOF error since the file is still being replaced
						// so early out here and proceed on the next notification
						file.Close()
						core.Debug("DecodeBinWrapper() EOF error, will wait for next notification")
						continue
					} else if err != nil {
						file.Close()
						core.Error("DecodeBinWrapper() error: %v", err)
						continue
					}

					// Close the file since it is no longer needed
					file.Close()

					if databaseNew.IsEmpty() {
						// Don't want to use an empty bin wrapper
						// so early out here and use existing array and hash
						core.Error("new database file is empty, keeping previous values")
						continue
					}

					// Store the creator and creation time from the database
					binCreator = databaseNew.Creator
					binCreationTime = databaseNew.CreationTime

					// Get the new relay array
					relayArrayNew := databaseNew.Relays
					// Proceed to fill up the new relay hash
					backend.SortAndHashRelayArray(relayArrayNew, relayHashNew)

					// Pointer swap the relay array
					relayArrayMutex.Lock()
					relayArray_internal = relayArrayNew
					relayArrayMutex.Unlock()

					// Pointer swap the relay hash
					relayHashMutex.Lock()
					relayHash_internal = relayHashNew
					relayHashMutex.Unlock()
				}
			}
		}()
	}

	// Setup magic goroutine
	{
		go func() {
			var cachedCombinedMagic []byte

			httpClient := &http.Client{
				Timeout: cfg.HTTPTimeout,
			}

			magicTicker := time.NewTicker(cfg.MagicPollFrequency)
			magicURI := fmt.Sprintf("http://%s/magic", cfg.MagicBackendIP)
			for {
				select {
				case <-ctx.Done():
					return
				case <-magicTicker.C:
					var magicReader io.ReadCloser

					if r, err := httpClient.Get(magicURI); err == nil {
						magicReader = r.Body
					}

					if magicReader == nil {
						core.Error("failed to get magic values: %v", err)
						gatewayMetrics.ErrorMetrics.MagicReaderNil.Add(1)
						continue
					}

					buffer, err := ioutil.ReadAll(magicReader)
					magicReader.Close()
					if err != nil {
						core.Error("failed to read magic data: %v", err)
						gatewayMetrics.ErrorMetrics.MagicReadFailure.Add(1)
						continue
					}

					if len(buffer) == 0 {
						core.Error("magic data buffer is empty")
						gatewayMetrics.ErrorMetrics.MagicBufferEmpty.Add(1)
						continue
					}

					if len(buffer) != 24 {
						core.Error("expected combined magic to be 24 bytes, got %d", len(buffer))
						gatewayMetrics.ErrorMetrics.MagicUnexpectedLengthError.Add(1)
						continue
					}

					if bytes.Equal(cachedCombinedMagic, buffer) {
						// Magic values are the same
						continue
					}

					magicMutex.Lock()
					magicUpcoming_internal = buffer[0:8]
					magicCurrent_internal = buffer[8:16]
					magicPrevious_internal = buffer[16:24]
					magicMutex.Unlock()

					cachedCombinedMagic = buffer

					upcomingMagic := buffer[0:8]
					currentMagic := buffer[8:16]
					previousMagic := buffer[16:24]

					core.Debug("updated magic values: %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x | %02x,%02x,%02x,%02x,%02x,%02x,%02x,%02x",
						upcomingMagic[0],
						upcomingMagic[1],
						upcomingMagic[2],
						upcomingMagic[3],
						upcomingMagic[4],
						upcomingMagic[5],
						upcomingMagic[6],
						upcomingMagic[7],
						currentMagic[0],
						currentMagic[1],
						currentMagic[2],
						currentMagic[3],
						currentMagic[4],
						currentMagic[5],
						currentMagic[6],
						currentMagic[7],
						previousMagic[0],
						previousMagic[1],
						previousMagic[2],
						previousMagic[3],
						previousMagic[4],
						previousMagic[5],
						previousMagic[6],
						previousMagic[7])

					gatewayMetrics.RefreshedMagicValues.Add(1)
				}
			}
		}()
	}

	// Create an error channel for goroutines
	errChan := make(chan error, 1)

	// Create a channel to hold incoming relay update requests
	updateChan := make(chan []byte, cfg.ChannelBufferSize)

	// Create a waitgroup to manage clean shutdown
	var wg sync.WaitGroup

	// Prioritize using HTTP to batch-send updates to relay backends
	if cfg.UseHTTP {
		// Create a Gateway HTTP Client
		gatewayHTTPClient, err := gateway.NewGatewayHTTPClient(cfg, updateChan, gatewayMetrics)
		if err != nil {
			core.Error("could not create gateway http client: %v", err)
			return 1
		}

		// Start up goroutines to POST to relay backends
		go gatewayHTTPClient.Start(ctx, &wg)

	} else {
		// TODO: implement ZeroMQ functionality
		core.Error("ZeroMQ is not yet supported")
		return 1

		// // Use ZeroMQ to publish updates to relay backend
		// var publishers []pubsub.Publisher
		// refreshPubs := make(chan bool, 1)
		// publishers, err := pubsub.NewMultiPublisher(cfg.PublishToHosts, cfg.PublisherSendBuffer)
		// if err != nil {
		//     level.Error(logger).Log("err", err)
		//     os.Exit(1)
		// }

		// go func() {
		//     syncTimer := helpers.NewSyncTimer(cfg.PublisherRefreshTimer)
		//     for {
		//         syncTimer.Run()
		//         refreshPubs <- true
		//     }
		// }()

		// go func() {
		//     for {
		//         select {
		//         case <-refreshPubs:
		//             newPublishers, err := pubsub.NewMultiPublisher(cfg.PublishToHosts, cfg.PublisherSendBuffer)
		//             if err != nil {
		//                 _ = level.Error(logger).Log("err", err)
		//                 continue
		//             }

		//             for _, pub := range publishers {
		//                 err = pub.Close()
		//                 if err != nil {
		//                     _ = level.Error(logger).Log("err", err)
		//                 }
		//             }

		//             publishers = newPublishers

		//             continue

		//         case msg := <-updateChan:
		//             for _, pub := range publishers {
		//                 _, err = pub.Publish(context.Background(), pubsub.RelayUpdateTopic, msg)
		//                 if err != nil {
		//                     _ = level.Error(logger).Log("msg", "unable to send update to optimizer", "err", err)
		//                 }
		//             }
		//         }
		//     }
		// }()
	}

	// Setup the status handler info

	statusData := &metrics.RelayGatewayStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				gatewayMetrics.GatewayServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				gatewayMetrics.GatewayServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.RelayGatewayStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = commitHash
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				// Service Metrics
				newStatusData.Goroutines = int(gatewayMetrics.GatewayServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = gatewayMetrics.GatewayServiceMetrics.MemoryAllocated.Value()

				// Requests
				newStatusData.UpdateRequestsReceived = int(gatewayMetrics.UpdatesReceived.Value())
				newStatusData.UpdateRequestsQueued = int(gatewayMetrics.UpdatesQueued.Value())
				newStatusData.UpdateRequestsFlushed = int(gatewayMetrics.UpdatesFlushed.Value())
				newStatusData.RefreshedMagicValues = int(gatewayMetrics.RefreshedMagicValues.Value())

				// Errors
				newStatusData.UpdateRequestReadPacketFailure = int(gatewayMetrics.ErrorMetrics.ReadPacketFailure.Value())
				newStatusData.UpdateRequestContentTypeFailure = int(gatewayMetrics.ErrorMetrics.ContentTypeFailure.Value())
				newStatusData.UpdateRequestUnmarshalFailure = int(gatewayMetrics.ErrorMetrics.UnmarshalFailure.Value())
				newStatusData.UpdateRequestExceedMaxRelaysError = int(gatewayMetrics.ErrorMetrics.ExceedMaxRelays.Value())
				newStatusData.UpdateRequestRelayNotFoundError = int(gatewayMetrics.ErrorMetrics.RelayNotFound.Value())
				newStatusData.UpdateResponseMarshalBinaryFailure = int(gatewayMetrics.ErrorMetrics.MarshalBinaryResponseFailure.Value())
				newStatusData.BatchUpdateRequestMarshalBinaryFailure = int(gatewayMetrics.ErrorMetrics.MarshalBinaryFailure.Value())
				newStatusData.BatchUpdateRequestBackendSendFailure = int(gatewayMetrics.ErrorMetrics.BackendSendFailure.Value())
				newStatusData.MagicReaderNil = int(gatewayMetrics.ErrorMetrics.MagicReaderNil.Value())
				newStatusData.MagicReadFailure = int(gatewayMetrics.ErrorMetrics.MagicReadFailure.Value())
				newStatusData.MagicBufferEmpty = int(gatewayMetrics.ErrorMetrics.MagicBufferEmpty.Value())
				newStatusData.MagicUnexpectedLengthError = int(gatewayMetrics.ErrorMetrics.MagicUnexpectedLengthError.Value())

				statusMutex.Lock()
				statusData = newStatusData
				statusMutex.Unlock()

				time.Sleep(time.Second * 10)
			}
		}()
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

	updateParams := transport.GatewayRelayUpdateHandlerConfig{
		RequestChan:  updateChan,
		Metrics:      gatewayMetrics,
		GetRelayData: GetRelayData,
		GetMagicData: GetMagicData,
	}

	port := envvar.Get("PORT", "30000")
	
	fmt.Printf("starting http server on port %s\n", port)

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildTime, commitMessage, commitHash, []string{}))
	router.HandleFunc("/status", serveStatusFunc).Methods("GET")
	router.HandleFunc("/database_version", transport.DatabaseBinVersionFunc(&binCreator, &binCreationTime, &env))
	router.HandleFunc("/relay_init", transport.GatewayRelayInitHandlerFunc()).Methods("POST")
	router.HandleFunc("/relay_update", transport.GatewayRelayUpdateHandlerFunc(updateParams)).Methods("POST")
	router.Handle("/debug/vars", expvar.Handler())

	enablePProf := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
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
		fmt.Println("received shutdown signal")

		cancel()
		// Wait for essential goroutines to finish up
		wg.Wait()

		fmt.Println("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		// Still let essential goroutines finish even though we got an error
		cancel()
		wg.Wait()
		return 1
	}
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

func GetMagicData() ([]byte, []byte, []byte) {
	magicMutex.RLock()
	magicUpcoming := magicUpcoming_internal
	magicCurrent := magicCurrent_internal
	magicPrevious := magicPrevious_internal
	magicMutex.RUnlock()

	return magicUpcoming, magicCurrent, magicPrevious
}

// Get the config for how this relay gateway should operate
func newConfig() (*gateway.GatewayConfig, error) {
	cfg := new(gateway.GatewayConfig)
	// Get the channel size
	channelBufferSize := envvar.GetInt("GATEWAY_CHANNEL_BUFFER_SIZE", 100000)
	cfg.ChannelBufferSize = channelBufferSize

	binSyncInterval := envvar.GetDuration("BIN_SYNC_INTERVAL", time.Minute)
	cfg.BinSyncInterval = binSyncInterval

	magicPollFrequency := envvar.GetDuration("MAGIC_POLL_FREQUENCY", time.Second)
	cfg.MagicPollFrequency = magicPollFrequency

	cfg.MagicBackendIP = envvar.Get("MAGIC_BACKEND_IP", "127.0.0.1:41007")
	if cfg.MagicBackendIP == "" {
		return nil, fmt.Errorf("MAGIC_BACKEND_IP not set")
	}

	// Decide if we are using HTTP to batch-write to relay backends
	useHTTP := envvar.GetBool("GATEWAY_USE_HTTP", true)
	cfg.UseHTTP = useHTTP

	// Load env vars depending on relay update delivery method
	if useHTTP {
		// Using HTTP, get the relay backend addresses to send relay updates to
		if exists := envvar.Exists("RELAY_BACKEND_ADDRESSES"); !exists {
			return nil, fmt.Errorf("RELAY_BACKEND_ADDRESSES not set")
		}
		relayBackendAddresses := envvar.GetList("RELAY_BACKEND_ADDRESSES", []string{})
		cfg.RelayBackendAddresses = relayBackendAddresses

		// Get the HTTP timeout duration
		httpTimeout := envvar.GetDuration("HTTP_TIMEOUT", time.Second)
		cfg.HTTPTimeout = httpTimeout

		// Get the batch size threshold for sending updates to relay backends
		batchSize := envvar.GetInt("GATEWAY_BACKEND_BATCH_SIZE", 10)
		cfg.BatchSize = batchSize
	} else {

		panic("not supported yet!")

	}

	return cfg, nil
}
