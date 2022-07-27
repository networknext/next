/*
   Network Next. You control the network.
   Copyright © 2017 - 2022 Network Next, Inc. All rights reserved.
*/

package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"

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

	overlayBinFileName := envvar.Get("OVERLAY_FILE_NAME", "")
	if overlayBinFileName == "" {
		core.Error("OVERLAY_FILE_NAME not set")
		return 1
	}

	overlayBinFileOutputLocation := envvar.Get("OVERLAY_OUTPUT_LOCATION", "")
	if overlayBinFileOutputLocation == "" {
		core.Error("OVERLAY_OUTPUT_LOCATION not set")
		return 1
	}

	binFileGCPTimeout, err := envvar.GetDuration("BIN_FILE_GCP_TIMEOUT", time.Second*5)
	if err != nil {
		core.Error("failed to parse BIN_FILE_GCP_TIMEOUT: %v", err)
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
	defer gcpStorage.Client.Close()

	// Setup http client for maxmind DB file
	maxmindHttpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	ispOutputLocation := envvar.Get("MAXMIND_ISP_OUTPUT_LOCATION", "")
	if ispOutputLocation == "" {
		core.Error("MAXMIND_ISP_OUTPUT_LOCATION not set")
		return 1
	}

	cityOutputLocation := envvar.Get("MAXMIND_CITY_OUTPUT_LOCATION", "")
	if cityOutputLocation == "" {
		core.Error("MAXMIND_CITY_OUTPUT_LOCATION not set")
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

	// Setup maxmind download and sync go routine
	maxmindISPDownloadInterval, err := envvar.GetDuration("MAXMIND_ISP_DOWNLOAD_DB_INTERVAL", time.Hour*24)
	if err != nil {
		core.Error("failed to parse MAXMIND_ISP_DOWNLOAD_DB_INTERVAL: %v", err)
		return 1
	}

	maxmindISPSyncInterval, err := envvar.GetDuration("MAXMIND_ISP_SYNC_DB_INTERVAL", time.Hour*25)
	if err != nil {
		core.Error("failed to parse MAXMIND_ISP_SYNC_DB_INTERVAL: %v", err)
		return 1
	}

	maxmindCityDownloadInterval, err := envvar.GetDuration("MAXMIND_CITY_DOWNLOAD_DB_INTERVAL", time.Hour*24)
	if err != nil {
		core.Error("failed to parse MAXMIND_CITY_DOWNLOAD_DB_INTERVAL: %v", err)
		return 1
	}

	maxmindCitySyncInterval, err := envvar.GetDuration("MAXMIND_CITY_SYNC_DB_INTERVAL", time.Hour*25)
	if err != nil {
		core.Error("failed to parse MAXMIND_CITY_SYNC_DB_INTERVAL: %v", err)
		return 1
	}

	maxmindServerBackendSync, err := envvar.GetBool("MAXMIND_SERVER_BACKEND_SYNC", false)
	if err != nil {
		core.Error("failed to parse MAXMIND_SERVER_BACKEND_SYNC: %v", err)
		return 1
	}

	// Maxmind ISP download goroutine
	var maxmindISPMutex sync.RWMutex
	maxmindISPDownloadTicker := time.NewTicker(maxmindISPDownloadInterval)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-maxmindISPDownloadTicker.C:
				var err error

				start := time.Now()

				ispRes, err := maxmindHttpClient.Get(ispURI)
				if err != nil {
					core.Error("failed to get ISP file from maxmind: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureISP.Add(1)
					continue
				}

				if ispRes.StatusCode != http.StatusOK {
					core.Error("HTTP GET was not successful for ISP file. Status Code: %d, Status Text: %s", ispRes.StatusCode, http.StatusText(ispRes.StatusCode))
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureISP.Add(1)
					ispRes.Body.Close()
					continue
				}

				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulHTTPCallsISP.Add(1)

				// Decompress file in memory
				gz, err := gzip.NewReader(ispRes.Body)
				if err != nil {
					core.Error("failed to open ISP file with GZIP: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
					ispRes.Body.Close()
					continue
				}

				bufISP := bytes.NewBuffer(nil)
				tr := tar.NewReader(gz)
				for {
					var hdr *tar.Header

					hdr, err = tr.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						core.Error("failed to read from GZIP file: %v", err)
						relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
						break
					}

					if strings.HasSuffix(hdr.Name, "mmdb") {
						_, err = io.Copy(bufISP, tr)
						if err != nil {
							core.Error("failed to copy ISP data to buffer: %v", err)
							relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
							break
						}
					}
				}

				gz.Close()
				ispRes.Body.Close()

				if err != nil && err != io.EOF {
					continue
				}

				// Write file to disk
				maxmindISPMutex.Lock()
				ispFilePath, err := os.Create(ispOutputLocation)
				if err != nil {
					core.Error("failed to create ISP file at %s: %v", ispOutputLocation, err)
					maxmindISPMutex.Unlock()
					continue
				}

				bytesWritten, err := io.Copy(ispFilePath, bufISP)
				if err != nil {
					core.Error("failed to write ISP file to disk at %s: %v", ispOutputLocation, err)
					maxmindISPMutex.Unlock()
					continue
				}

				maxmindISPMutex.Unlock()

				updateTime := time.Since(start)
				duration := float64(updateTime.Milliseconds())

				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBISPUpdateDuration.Set(duration)

				core.Debug("Wrote ISP file to disk at %s (%d bytes)", ispOutputLocation, bytesWritten)
			}
		}
	}()

	// Maxmind City download goroutine
	var maxmindCityMutex sync.RWMutex
	maxmindCityDownloadTicker := time.NewTicker(maxmindCityDownloadInterval)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-maxmindCityDownloadTicker.C:
				var err error

				start := time.Now()

				cityRes, err := maxmindHttpClient.Get(cityURI)
				if err != nil {
					core.Error("failed to get City file from maxmind: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureCity.Add(1)
					continue
				}

				if cityRes.StatusCode != http.StatusOK {
					core.Error("HTTP GET was not successful for City file. Status Code: %d, Status Text: %s", cityRes.StatusCode, http.StatusText(cityRes.StatusCode))
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureCity.Add(1)
					cityRes.Body.Close()
					continue
				}

				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulHTTPCallsCity.Add(1)

				// Decompress file in memory
				gz, err := gzip.NewReader(cityRes.Body)
				if err != nil {
					core.Error("failed to open City file with GZIP: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
					cityRes.Body.Close()
					continue
				}

				bufCity := bytes.NewBuffer(nil)
				tr := tar.NewReader(gz)
				for {
					var hdr *tar.Header

					hdr, err = tr.Next()
					if err == io.EOF {
						break
					}
					if err != nil {
						core.Error("failed to read from GZIP file: %v", err)
						relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
						break
					}

					if strings.HasSuffix(hdr.Name, "mmdb") {
						_, err = io.Copy(bufCity, tr)
						if err != nil {
							core.Error("failed to copy City data to buffer: %v", err)
							relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
							break
						}
					}
				}

				gz.Close()
				cityRes.Body.Close()

				if err != nil && err != io.EOF {
					continue
				}

				// Write file to disk
				maxmindCityMutex.Lock()
				cityFilePath, err := os.Create(cityOutputLocation)
				if err != nil {
					core.Error("failed to create City file at %s: %v", cityOutputLocation, err)
					maxmindCityMutex.Unlock()
					continue
				}

				bytesWritten, err := io.Copy(cityFilePath, bufCity)
				if err != nil {
					core.Error("failed to write City file to disk at %s: %v", cityOutputLocation, err)
					maxmindCityMutex.Unlock()
					continue
				}

				maxmindCityMutex.Unlock()

				updateTime := time.Since(start)
				duration := float64(updateTime.Milliseconds())

				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBCityUpdateDuration.Set(duration)

				core.Debug("Wrote City file to disk at %s (%d bytes)", cityOutputLocation, bytesWritten)
			}
		}
	}()

	// Maxmind ISP cloud storage goroutine
	maxmindISPUploadTicker := time.NewTicker(maxmindISPSyncInterval)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-maxmindISPUploadTicker.C:
				// Copy the ISP file to GCP Storage
				maxmindISPMutex.RLock()

				if err := validateISPFile(ctx, env, ispStorageName); err != nil {
					maxmindISPMutex.RUnlock()
					core.Error("failed to validate ISP file: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindValidationFailureISP.Add(1)
					continue
				}

				if err := gcpStorage.CopyFromLocalToBucket(ctx, ispOutputLocation, bucketName, ispStorageName); err != nil {
					maxmindISPMutex.RUnlock()
					core.Error("failed to copy maxmind ISP file to GCP Cloud Storage: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindStorageUploadFailureISP.Add(1)
					continue
				} else {
					relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulISPStorageUploads.Add(1)
				}
				maxmindISPMutex.RUnlock()
			}
		}
	}()

	if maxmindServerBackendSync {
		// Maxmind ISP sync goroutine
		maxmindISPSyncTicker := time.NewTicker(maxmindISPSyncInterval)

		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case <-maxmindISPSyncTicker.C:
					start := time.Now()

					// Store the known list of instance names
					allBackendInstanceNames := serverBackendInstanceNames

					// The names of instances in a MIG can change, so get them each time
					serverBackendMIGInstanceNames, err := getMIGInstanceNames(gcpProjectID, serverBackendMIGName)
					if err != nil {
						core.Error("failed to fetch server backend mig instance names: %v", err)
					} else {
						// Add the server backend mig instance names to the list
						allBackendInstanceNames = append(allBackendInstanceNames, serverBackendMIGInstanceNames...)
					}

					// Copy the ISP file to each Server Backend
					maxmindISPMutex.RLock()

					if err := validateISPFile(ctx, env, ispStorageName); err != nil {
						maxmindISPMutex.RUnlock()
						core.Error("failed to validate ISP file: %v", err)
						relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindValidationFailureISP.Add(1)
						continue
					}

					for _, instanceName := range allBackendInstanceNames {
						if err := gcpStorage.CopyFromLocalToRemote(ctx, ispOutputLocation, instanceName); err != nil {
							core.Error("failed to copy maxmind ISP file to instance %s: %v", instanceName, err)
							relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindSCPWriteFailure.Add(1)
						} else {
							core.Debug("successfully copied maxmind ISP file to instance %s", instanceName)
							relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulISPSCP.Add(1)
						}
					}
					maxmindISPMutex.RUnlock()

					updateTime := time.Since(start)
					duration := float64(updateTime.Milliseconds())

					relayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBISPUpdateDuration.Set(duration)
				}
			}
		}()
	}

	// Maxmind City cloud storage goroutine
	maxmindCityUploadTicker := time.NewTicker(maxmindCitySyncInterval)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-maxmindCityUploadTicker.C:
				// Copy the City file to GCP Storage
				maxmindCityMutex.RLock()

				if err := validateCityFile(ctx, env, cityStorageName); err != nil {
					maxmindCityMutex.RUnlock()
					core.Error("failed to validate ISP file: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindValidationFailureISP.Add(1)
					continue
				}

				if err := gcpStorage.CopyFromLocalToBucket(ctx, cityOutputLocation, bucketName, cityStorageName); err != nil {
					maxmindCityMutex.RUnlock()
					core.Error("failed to copy maxmind City file to GCP Cloud Storage: %v", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindStorageUploadFailureCity.Add(1)
					continue
				} else {
					relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulCityStorageUploads.Add(1)
				}
				maxmindCityMutex.RUnlock()
			}
		}
	}()

	if maxmindServerBackendSync {
		// Maxmind City sync goroutine
		maxmindCitySyncTicker := time.NewTicker(maxmindCitySyncInterval)

		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case <-maxmindCitySyncTicker.C:
					start := time.Now()

					// Store the known list of instance names
					allBackendInstanceNames := serverBackendInstanceNames

					// The names of instances in a MIG can change, so get them each time
					serverBackendMIGInstanceNames, err := getMIGInstanceNames(gcpProjectID, serverBackendMIGName)
					if err != nil {
						core.Error("failed to fetch server backend mig instance names: %v", err)
					} else {
						// Add the server backend mig instance names to the list
						allBackendInstanceNames = append(allBackendInstanceNames, serverBackendMIGInstanceNames...)
					}

					// Copy the City file to each Server Backend
					maxmindCityMutex.RLock()

					if err := validateCityFile(ctx, env, cityStorageName); err != nil {
						maxmindCityMutex.RUnlock()
						core.Error("failed to validate ISP file: %v", err)
						relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindValidationFailureISP.Add(1)
						continue
					}

					for _, instanceName := range allBackendInstanceNames {
						if err := gcpStorage.CopyFromLocalToRemote(ctx, cityOutputLocation, instanceName); err != nil {
							core.Error("failed to copy maxmind City file to instance %s: %v", instanceName, err)
							relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindSCPWriteFailure.Add(1)
						} else {
							core.Debug("successfully copied maxmind City file to instance %s", instanceName)
							relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulCitySCP.Add(1)
						}
					}
					maxmindCityMutex.RUnlock()

					updateTime := time.Since(start)
					duration := float64(updateTime.Milliseconds())

					relayPusherServiceMetrics.RelayPusherMetrics.MaxmindDBCityUpdateDuration.Set(duration)
				}
			}
		}()
	}

	// Create error channel to error out from any goroutines
	errChan := make(chan error, 1)

	// Database binary sync goroutine
	dbTicker := time.NewTicker(dbSyncInterval)
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-dbTicker.C:
				// Use anonymous function to allow for defers to complete
				func() {
					start := time.Now()

					defer func() {
						updateTime := time.Since(start)
						duration := float64(updateTime.Milliseconds())
						relayPusherServiceMetrics.RelayPusherMetrics.BinaryTotalUpdateDuration.Set(duration)
					}()

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

					// Use specific context to let service restart if it takes too long for overlay.bin to be pulled from cloud storage
					gcpOverlayCtx, gcpOverlayCancel := context.WithTimeout(ctx, binFileGCPTimeout)
					defer gcpOverlayCancel()

					// Do the overlay.bin first. Relay backends pull in these changes only when database.bin changes
					if err := gcpStorage.CopyFromBucketToRemote(gcpOverlayCtx, overlayBinFileName, databaseInstanceNames, overlayBinFileOutputLocation); err != nil {
						core.Error("failed to copy overlay bin file to overlay locations: %v", err)

						switch err {
						case context.DeadlineExceeded:
							relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.BinFilePullTimeoutError.Add(1)
							errChan <- err
							return
						default:
							relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.OverlaySCPWriteFailure.Add(1)
							// Don't error out here, need to proceed to next step
						}
					}

					// Use specific context to let service restart if it takes too long for database.bin to be pulled from cloud storage
					gcpDBCtx, gcpDBCancel := context.WithTimeout(ctx, binFileGCPTimeout)
					defer gcpDBCancel()

					if err := gcpStorage.CopyFromBucketToRemote(gcpDBCtx, databaseBinFileName, databaseInstanceNames, databaseBinFileOutputLocation); err != nil {
						core.Error("failed to copy database bin file to database locations: %v", err)

						switch err {
						case context.DeadlineExceeded:
							relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.BinFilePullTimeoutError.Add(1)
							errChan <- err
							return
						default:
							relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.DatabaseSCPWriteFailure.Add(1)
							// Don't need to return here since end of operations and defers will execute
						}
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	}()

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
				newStatusData.BinaryTotalUpdateDurationMs = relayPusherServiceMetrics.RelayPusherMetrics.BinaryTotalUpdateDuration.Value()
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

func validateISPFile(ctx context.Context, env string, ispStorageName string) error {
	mmdb := &routing.MaxmindDB{
		IspFile:   ispStorageName,
		IsStaging: env == "staging",
	}

	// Validate the ISP file
	if err := mmdb.OpenISP(ctx); err != nil {
		return err
	}

	if err := mmdb.ValidateISP(); err != nil {
		return err
	}

	return nil
}

func validateCityFile(ctx context.Context, env string, cityStorageName string) error {
	mmdb := &routing.MaxmindDB{
		CityFile:  cityStorageName,
		IsStaging: env == "staging",
	}

	// Validate the ISP file
	if err := mmdb.OpenCity(ctx); err != nil {
		return err
	}

	if err := mmdb.ValidateCity(); err != nil {
		return err
	}

	return nil
}
