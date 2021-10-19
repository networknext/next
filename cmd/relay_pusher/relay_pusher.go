/*
   Network Next. You control the network.
   Copyright © 2017 - 2021 Network Next, Inc. All rights reserved.
*/

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/storage"

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
	serviceName := "relay_pusher"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	gcpProjectID := backend.GetGCPProjectID()
	if gcpProjectID == "" {
		core.Error("cannot run relay pusher without gcpProjectID")
		return 1
	}

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

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("failed to get metrics handler: %v", err)
		return 1
	}

	// Create relay pusher metrics
	relayPusherServiceMetrics, err := metrics.NewRelayPusherServiceMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to create relay pusher service metrics: %v", err)
		return 1
	}

	// Stackdriver Profiler
	if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
		core.Error("failed to initialize stackdriver profiler: %v", err)
		return 1
	}

	// Setup GCP storage
	bucketName := envvar.Get("ARTIFACT_BUCKET", "")
	if bucketName == "" {
		core.Error("ARTIFACT_BUCKET not set")
		return 1
	}

	databaseBinFileName := envvar.Get("DATABASE_FILE_NAME", "")
	if databaseBinFileName == "" {
		core.Error("DATABASE_FILE_NAME not set")
		return 1
	}

	databaseBinFileOutputLocation := envvar.Get("DB_OUTPUT_LOCATION", "")
	if databaseBinFileOutputLocation == "" {
		core.Error("DB_OUTPUT_LOCATION not set")
		return 1
	}

	dbSyncInterval, err := envvar.GetDuration("DB_SYNC_INTERVAL", time.Minute*1)
	if err != nil {
		core.Error("failed to parse DB_SYNC_INTERVAL: %v", err)
		return 1
	}

	remoteDBLocations := make([]string, 0)

	relayBackendNames := envvar.GetList("RELAY_BACKEND_INSTANCE_NAMES", []string{})
	if len(relayBackendNames) == 0 {
		core.Error("RELAY_BACKEND_INSTANCE_NAMES not set")
		return 1
	}
	for _, relayBackendName := range relayBackendNames {
		remoteDBLocations = append(remoteDBLocations, relayBackendName)
	}

	debugRelayBackendName := envvar.Get("DEBUG_RELAY_BACKEND_INSTANCE_NAME", "")
	if debugRelayBackendName == "" {
		core.Error("DEBUG_RELAY_BACKEND_INSTANCE_NAME not set. Will not send data to debug instance.")
	} else {
		remoteDBLocations = append(remoteDBLocations, debugRelayBackendName)
	}

	relayGatewayMIGName := envvar.Get("RELAY_GATEWAY_MIG_NAME", "")
	if relayGatewayMIGName == "" {
		core.Error("RELAY_GATEWAY_MIG_NAME not set")
		return 1
	}

	apiMIGName := envvar.Get("API_MIG_NAME", "")
	if apiMIGName == "" {
		core.Error("API_MIG_NAME not set")
		return 1
	}

	serverBackendMIGName := envvar.Get("SERVER_BACKEND_MIG_NAME", "")
	if serverBackendMIGName == "" {
		core.Error("SERVER_BACKEND_MIG_NAME not set")
		return 1
	}

	serverBackendInstanceNames := make([]string, 0)

	// Check if the debug server backend exists and push files to it as well
	debugServerBackendName := envvar.Get("DEBUG_SERVER_BACKEND_NAME", "")
	if debugServerBackendName == "" {
		core.Error("DEBUG_SERVER_BACKEND_NAME not set. Will not send data to debug instance.")
	} else {
		serverBackendInstanceNames = append(serverBackendInstanceNames, debugServerBackendName)
	}

	gcpStorage, err := storage.NewGCPStorageClient(ctx, bucketName)
	if err != nil {
		core.Error("failed to create gcp storage client: %v", err)
		return 1
	}

	// Setup http client for maxmind DB file
	maxmindHttpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	ispFileName := envvar.Get("MAXMIND_ISP_DB_FILE_NAME", "")
	if ispFileName == "" {
		core.Error("MAXMIND_ISP_DB_FILE_NAME not set")
		return 1
	}

	cityFileName := envvar.Get("MAXMIND_CITY_DB_FILE_NAME", "")
	if cityFileName == "" {
		core.Error("MAXMIND_CITY_DB_FILE_NAME not set")
		return 1
	}

	ispStorageName := envvar.Get("MAXMIND_ISP_STORAGE_FILE_NAME", "")
	if ispStorageName == "" {
		core.Error("MAXMIND_ISP_STORAGE_FILE_NAME not set")
		return 1
	}

	cityStorageName := envvar.Get("MAXMIND_CITY_STORAGE_FILE_NAME", "")
	if cityStorageName == "" {
		core.Error("MAXMIND_CITY_STORAGE_FILE_NAME not set")
		return 1
	}

	ispURI := envvar.Get("MAXMIND_ISP_DB_URI", "")
	if ispURI == "" {
		core.Error("MAXMIND_ISP_DB_URI not set")
		return 1
	}

	cityURI := envvar.Get("MAXMIND_CITY_DB_URI", "")
	if cityURI == "" {
		core.Error("MAXMIND_CITY_DB_URI not set")
		return 1
	}

	// Setup maxmind download go routine
	maxmindSyncInterval, err := envvar.GetDuration("MAXMIND_SYNC_DB_INTERVAL", time.Hour*24)
	if err != nil {
		core.Error("failed to parse MAXMIND_SYNC_DB_INTERVAL: %v", err)
		return 1
	}

	// Maxmind sync routine
	maxmindTicker := time.NewTicker(maxmindSyncInterval)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-maxmindTicker.C:
				start := time.Now()

				ispRes, err := maxmindHttpClient.Get(ispURI)
				if err != nil {
					core.Error("failed to get ISP file from maxmind: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureISP.Add(1)
					continue
				}

				defer ispRes.Body.Close()

				if ispRes.StatusCode != http.StatusOK {
					core.Error("HTTP GET was not successful for ISP file. Status Code: %d, Status Text: %s", ispRes.StatusCode, http.StatusText(ispRes.StatusCode))
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureISP.Add(1)
					continue
				}

				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulHTTPCallsISP.Add(1)

				gz, err := gzip.NewReader(ispRes.Body)
				if err != nil {
					core.Error("failed to open ISP file with GZIP: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
					continue
				}

				buf := bytes.NewBuffer(nil)
				tr := tar.NewReader(gz)
				for {
					hdr, err := tr.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						core.Error("failed to read from GZIP file: %v", err)
						relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
						continue
					}

					if strings.HasSuffix(hdr.Name, "mmdb") {
						_, err := io.Copy(buf, tr)
						if err != nil {
							core.Error("failed to copy ISP data to buffer: %v", err)
							relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
							continue
						}
					}
				}
				gz.Close()

				if err := gcpStorage.CopyFromBytesToRemote(buf.Bytes(), serverBackendInstanceNames, ispFileName); err != nil {
					core.Error("failed to copy maxmind ISP file to server backends: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindSCPWriteFailure.Add(1)
					// Don't continue here, we need to try out the city file as well
				} else {
					relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulISPSCP.Add(1)
				}

				if err := gcpStorage.CopyFromBytesToStorage(ctx, buf.Bytes(), ispStorageName); err != nil {
					core.Error("failed to copy maxmind ISP file to GCP Cloud Storage: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindStorageUploadFailureISP.Add(1)
					// Don't continue here, we need to try out the city file as well
				} else {
					relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulISPStorageUploads.Add(1)
				}

				updateTime := time.Since(start)
				duration := float64(updateTime.Milliseconds())

				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBISPUpdateDuration.Set(duration)

				cityRes, err := maxmindHttpClient.Get(cityURI)
				if err != nil {
					core.Error("failed to get City file from maxmind: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureCity.Add(1)
					continue
				}

				defer cityRes.Body.Close()

				if cityRes.StatusCode != http.StatusOK {
					core.Error("HTTP GET was not successful for City file. Status Code: %d, Status Text: %s", cityRes.StatusCode, http.StatusText(cityRes.StatusCode))
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureCity.Add(1)
					continue
				}

				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulHTTPCallsCity.Add(1)

				gz, err = gzip.NewReader(cityRes.Body)
				if err != nil {
					core.Error("failed to open City file with GZIP: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
					continue
				}

				buf = bytes.NewBuffer(nil)
				tr = tar.NewReader(gz)
				for {
					hdr, err := tr.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						core.Error("failed to read from GZIP file: %v", err)
						relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
						continue
					}

					if strings.HasSuffix(hdr.Name, "mmdb") {
						_, err := io.Copy(buf, tr)
						if err != nil {
							core.Error("failed to copy City data to buffer: %v", err)
							relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
							continue
						}
					}
				}
				gz.Close()

				// Store the known list of instance names
				maxmindInstanceNames := serverBackendInstanceNames

				// The names of instances in a MIG can change, so get them each time
				serverBackendMIGInstanceNames, err := getMIGInstanceNames(gcpProjectID, serverBackendMIGName)
				if err != nil {
					core.Error("failed to fetch server backend mig instance names: %v", err)
				} else {
					// Add the server backend mig instance names to the list
					maxmindInstanceNames = append(maxmindInstanceNames, serverBackendMIGInstanceNames...)
				}

				if err := gcpStorage.CopyFromBytesToStorage(ctx, buf.Bytes(), cityStorageName); err != nil {
					core.Error("failed to copy maxmind City file to GCP Cloud Storage: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindStorageUploadFailureCity.Add(1)
				} else {
					relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulCityStorageUploads.Add(1)
				}

				if err := gcpStorage.CopyFromBytesToRemote(buf.Bytes(), maxmindInstanceNames, cityFileName); err != nil {
					core.Error("failed to copy maxmind City file to server backends: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindSCPWriteFailure.Add(1)
					// Don't continue here, need to record update duration
				} else {
					relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulCitySCP.Add(1)
				}

				updateTime = time.Since(start)
				duration = float64(updateTime.Milliseconds())

				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBCityUpdateDuration.Set(duration)
				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBTotalUpdateDuration.Set(duration)

			case <-ctx.Done():
				return
			}
		}
	}()

	// database binary sync routine
	dbTicker := time.NewTicker(dbSyncInterval)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-dbTicker.C:
				start := time.Now()

				// Store the known list of instance names
				databaseInstanceNames := remoteDBLocations

				// The names of instances in a MIG can change, so get them each time
				relayGatewayMIGInstanceNames, err := getMIGInstanceNames(gcpProjectID, relayGatewayMIGName)
				if err != nil {
					core.Error("failed to fetch relay gateway mig instance names: %v", err)
				} else {
					// Add the gateway mig instance names to the list
					databaseInstanceNames = append(databaseInstanceNames, relayGatewayMIGInstanceNames...)
				}

				apiMIGInstanceNames, err := getMIGInstanceNames(gcpProjectID, apiMIGName)
				if err != nil {
					core.Error("failed to fetch api mig instance names: %v", err)
				} else {
					// Add the gateway mig instance names to the list
					databaseInstanceNames = append(databaseInstanceNames, apiMIGInstanceNames...)
				}

				if err := gcpStorage.CopyFromBucketToRemote(ctx, databaseBinFileName, databaseInstanceNames, databaseBinFileOutputLocation); err != nil {
					core.Error("failed to copy database bin file to database locations: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.DatabaseSCPWriteFailure.Add(1)
					// Don't continue here, need to record update duration
				}

				updateTime := time.Since(start)
				duration := float64(updateTime.Milliseconds())

				relayPusherServiceMetrics.RelayPusherMetrics.DBBinaryTotalUpdateDuration.Set(duration)

			case <-ctx.Done():
				return
			}
		}
	}()

	// Create error channel to error out from any goroutines
	errChan := make(chan error, 1)

	// Setup the status handler info
	statusData := &metrics.RelayPusherStatus{}
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				relayPusherServiceMetrics.ServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				relayPusherServiceMetrics.ServiceMetrics.MemoryAllocated.Set(memoryUsed())

				newStatusData := &metrics.RelayPusherStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				// Service Metrics
				newStatusData.Goroutines = int(relayPusherServiceMetrics.ServiceMetrics.Goroutines.Value())
				newStatusData.MemoryAllocated = relayPusherServiceMetrics.ServiceMetrics.MemoryAllocated.Value()

				// Success Metrics
				newStatusData.MaxmindSuccessfulHTTPCallsISP = int(relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulHTTPCallsISP.Value())
				newStatusData.MaxmindSuccessfulHTTPCallsCity = int(relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulHTTPCallsCity.Value())
				newStatusData.MaxmindSuccessfulISPSCP = int(relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulISPSCP.Value())
				newStatusData.MaxmindSuccessfulCitySCP = int(relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulCitySCP.Value())
				newStatusData.MaxmindSuccessfulISPStorageUploads = int(relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulISPStorageUploads.Value())
				newStatusData.MaxmindSuccessfulCityStorageUploads = int(relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulCityStorageUploads.Value())

				// Error Metrics
				newStatusData.MaxmindSuccessfulCityStorageUploads = int(relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureISP.Value())
				newStatusData.MaxmindHTTPFailureCity = int(relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureCity.Value())
				newStatusData.MaxmindGZIPReadFailure = int(relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Value())
				newStatusData.MaxmindTempFileWriteFailure = int(relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindTempFileWriteFailure.Value())
				newStatusData.MaxmindSCPWriteFailure = int(relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindSCPWriteFailure.Value())
				newStatusData.MaxmindStorageUploadFailureISP = int(relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindStorageUploadFailureISP.Value())
				newStatusData.MaxmindStorageUploadFailureCity = int(relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindStorageUploadFailureCity.Value())
				newStatusData.DatabaseSCPWriteFailure = int(relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.DatabaseSCPWriteFailure.Value())

				// Durations
				newStatusData.DBBinaryTotalUpdateDurationMs = relayPusherServiceMetrics.RelayPusherMetrics.DBBinaryTotalUpdateDuration.Value()
				newStatusData.MaxmindDBTotalUpdateDurationMs = relayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBTotalUpdateDuration.Value()
				newStatusData.MaxmindDBCityUpdateDurationMs = relayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBCityUpdateDuration.Value()
				newStatusData.MaxmindDBISPUpdateDurationMs = relayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBISPUpdateDuration.Value()

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

	// Start HTTP server
	{
		port := envvar.Get("PORT", "")
		if port == "" {
			core.Error("PORT not set")
			return 1
		}
		fmt.Printf("starting http server on :%s\n", port)

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.Handle("/debug/vars", expvar.Handler())

		enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if err != nil {
			core.Error("could not parse envvar FEATURE_ENABLE_PPROF: %v", err)
		}
		if enablePProf {
			router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
		}

		go func() {
			err := http.ListenAndServe(":"+port, router)
			if err != nil {
				core.Error("failed to start http server: %v", err)
				errChan <- err
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		fmt.Println("Received shutdown signal.")

		cancel()
		// Wait for essential goroutines to finish up
		wg.Wait()

		fmt.Println("Successfully shutdown.")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		cancel()

		return 1
	}
}

func getMIGInstanceNames(gcpProjectID string, migName string) ([]string, error) {
	// Get the latest instance names in the relay gateway mig
	runnable := exec.Command("gcloud", "compute", "--project", gcpProjectID, "instance-groups", "managed", "list-instances", migName, "--zone", "us-central1-a", "--format", "value(instance)")

	buffer, err := runnable.CombinedOutput()
	if err != nil {
		return []string{}, err
	}

	migInstanceNames := strings.Split(string(buffer), "\n")

	// Using the method above causes an empty string to be added at the end of the slice - remove it
	if len(migInstanceNames) > 0 {
		migInstanceNames = migInstanceNames[:len(migInstanceNames)-1]
	}

	return migInstanceNames, nil
}
