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
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/storage"
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
	fmt.Printf("relay_pusher: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	serviceName := "relay_pusher"

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	gcpProjectID := backend.GetGCPProjectID()
	if gcpProjectID == "" {
		return 1
	}

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

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Create relay pusher metrics
	relayPusherServiceMetrics, err := metrics.NewRelayPusherServiceMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create relay pusher service metrics", "err", err)
		return 1
	}

	// Stackdriver Profiler
	if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
		level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
		return 1
	}

	// Setup GCP storage
	bucketName := envvar.Get("ARTIFACT_BUCKET", "")
	if bucketName == "" {
		level.Error(logger).Log("err", "gcp bucket not specified")
		return 1
	}

	databaseBinFileName := envvar.Get("DATABASE_FILE_NAME", "")
	if databaseBinFileName == "" {
		level.Error(logger).Log("err", "DB binary file name not specified")
		return 1
	}

	databaseBinFileOutputLocation := envvar.Get("DB_OUTPUT_LOCATION", "")
	if databaseBinFileOutputLocation == "" {
		level.Error(logger).Log("err", "DB output file location not specified")
		return 1
	}

	dbSyncInterval, err := envvar.GetDuration("DB_SYNC_INTERVAL", time.Minute*1)
	if err != nil {
		level.Error(logger).Log("err", "failed to get DB sync interval")
		return 1
	}

	remoteDBLocations := make([]string, 0)

	relayBackendNames := envvar.GetList("RELAY_BACKEND_INSTANCE_NAMES", []string{})
	if len(relayBackendNames) == 0 {
		level.Error(logger).Log("err", "relay backend names not specified")
		return 1
	}
	for _, relayBackendName := range relayBackendNames {
		remoteDBLocations = append(remoteDBLocations, relayBackendName)
	}

	debugRelayBackendName := envvar.Get("DEBUG_RELAY_BACKEND_INSTANCE_NAME", "")
	if debugRelayBackendName == "" {
		level.Error(logger).Log("err", "debug relay backend name not specified")
	} else {
		remoteDBLocations = append(remoteDBLocations, debugRelayBackendName)
	}

	relayGatewayMIGName := envvar.Get("RELAY_GATEWAY_MIG_NAME", "")
	if relayGatewayMIGName == "" {
		level.Error(logger).Log("err", "relay gateway mig name not specified")
		return 1
	}

	apiMIGName := envvar.Get("API_MIG_NAME", "")
	if apiMIGName == "" {
		level.Error(logger).Log("err", "api mig name not specified")
		return 1
	}

	serverBackendMIGName := envvar.Get("SERVER_BACKEND_MIG_NAME", "")
	if serverBackendMIGName == "" {
		level.Error(logger).Log("err", "server backend mig name not specified")
		return 1
	}

	// We sometimes have two server backend MIGs running depending on the env
	serverBackendMIGName2 := envvar.Get("SERVER_BACKEND_MIG_NAME_2", "")
	if serverBackendMIGName2 == "" {
		level.Error(logger).Log("err", "server backend 2 mig name not specified")
	}

	serverBackendInstanceNames := make([]string, 0)

	// Check if the debug server backend exists and push files to it as well
	debugServerBackendName := envvar.Get("DEBUG_SERVER_BACKEND_NAME", "")
	if debugServerBackendName == "" {
		level.Error(logger).Log("err", "debug server backend name not specified")
	} else {
		serverBackendInstanceNames = append(serverBackendInstanceNames, debugServerBackendName)
	}

	gcpStorage, err := storage.NewGCPStorageClient(ctx, bucketName, logger)
	if err != nil {
		level.Error(logger).Log("msg", "failed to initialze gcp storage", "err", err)
		return 1
	}

	// Setup http client for maxmind DB file
	maxmindHttpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	ispFileName := envvar.Get("MAXMIND_ISP_DB_FILE_NAME", "")
	if ispFileName == "" {
		level.Error(logger).Log("err", "ISP temp file not defined", "err")
		return 1
	}

	cityFileName := envvar.Get("MAXMIND_CITY_DB_FILE_NAME", "")
	if cityFileName == "" {
		level.Error(logger).Log("err", "city temp file not defined", "err")
		return 1
	}

	ispStorageName := envvar.Get("MAXMIND_ISP_STORAGE_FILE_NAME", "")
	if ispStorageName == "" {
		level.Error(logger).Log("err", "MAXMIND_ISP_STORAGE_FILE_NAME not defined", "err")
		return 1
	}

	cityStorageName := envvar.Get("MAXMIND_CITY_STORAGE_FILE_NAME", "")
	if cityStorageName == "" {
		level.Error(logger).Log("err", "MAXMIND_CITY_STORAGE_FILE_NAME not defined", "err")
		return 1
	}

	ispURI := envvar.Get("MAXMIND_ISP_DB_URI", "")
	if ispURI == "" {
		level.Error(logger).Log("err", "maxmind DB ISP location not defined")
		return 1
	}

	cityURI := envvar.Get("MAXMIND_CITY_DB_URI", "")
	if cityURI == "" {
		level.Error(logger).Log("err", "maxmind DB city location not defined")
		return 1
	}

	// Setup maxmind download go routine
	maxmindSyncInterval, err := envvar.GetDuration("MAXMIND_SYNC_DB_INTERVAL", time.Hour*24)
	if err != nil {
		level.Error(logger).Log("msg", "maxmind DB sync interval not defined", "err", err)
		return 1
	}

	// Maxmind sync routine
	maxmindTicker := time.NewTicker(maxmindSyncInterval)
	go func() {
		for {
			select {
			case <-maxmindTicker.C:
				start := time.Now()

				ispRes, err := maxmindHttpClient.Get(ispURI)
				if err != nil {
					level.Error(logger).Log("msg", "failed to get ISP file from maxmind", "err", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureISP.Add(1)
					continue
				}

				defer ispRes.Body.Close()

				if ispRes.StatusCode != http.StatusOK {
					level.Error(logger).Log("msg", "http get was not successful for ISP file", ispRes.StatusCode, http.StatusText(ispRes.StatusCode))
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureISP.Add(1)
					continue
				}

				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulHTTPCallsISP.Add(1)

				gz, err := gzip.NewReader(ispRes.Body)
				if err != nil {
					level.Error(logger).Log("msg", "failed to open isp file with gzip", "err", err)
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
						level.Error(logger).Log("msg", "failed reading from gzip file", "err", err)
						relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
						continue
					}

					if strings.HasSuffix(hdr.Name, "mmdb") {
						_, err := io.Copy(buf, tr)
						if err != nil {
							level.Error(logger).Log("msg", "failed to copy ISP data to buffer", "err", err)
							relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
							continue
						}
					}
				}
				gz.Close()

				if err := gcpStorage.CopyFromBytesToRemote(buf.Bytes(), serverBackendInstanceNames, ispFileName); err != nil {
					level.Error(logger).Log("msg", "failed to copy maxmind ISP file to server backends", "err", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindSCPWriteFailure.Add(1)
					// Don't continue here, we need to try out the city file as well
				}

				if err := gcpStorage.CopyFromBytesToStorage(ctx, buf.Bytes(), ispStorageName); err != nil {
					level.Error(logger).Log("msg", "failed to copy maxmind ISP file to gcp cloud storage", "err", err)
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
					level.Error(logger).Log("msg", "failed to get city file from maxmind", "err", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureCity.Add(1)
					continue
				}

				defer cityRes.Body.Close()

				if cityRes.StatusCode != http.StatusOK {
					level.Error(logger).Log("msg", "http get was not successful for ISP file", cityRes.StatusCode, http.StatusText(cityRes.StatusCode))
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindHTTPFailureCity.Add(1)
					continue
				}

				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulHTTPCallsCity.Add(1)

				gz, err = gzip.NewReader(cityRes.Body)
				if err != nil {
					level.Error(logger).Log("msg", "failed to open isp file with gzip", "err", err)
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
						level.Error(logger).Log("msg", "failed reading from gzip file", "err", err)
						relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindGZIPReadFailure.Add(1)
						continue
					}

					if strings.HasSuffix(hdr.Name, "mmdb") {
						_, err := io.Copy(buf, tr)
						if err != nil {
							level.Error(logger).Log("msg", "failed to copy ISP data to buffer", "err", err)
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
					level.Error(logger).Log("msg", "failed to fetch server backend mig instance names", "err", err)
				} else {
					// Add the server backend mig instance names to the list
					maxmindInstanceNames = append(maxmindInstanceNames, serverBackendMIGInstanceNames...)
				}

				// Add the instances for the second server backend MIG if it is in use
				if serverBackendMIGName2 != "" {
					serverBackendMIG2InstanceNames, err := getMIGInstanceNames(gcpProjectID, serverBackendMIGName2)
					if err != nil {
						level.Error(logger).Log("msg", "failed to fetch server backend 2 mig instance names", "err", err)
					} else {
						// Add the server backend 2 mig instance names to the list
						maxmindInstanceNames = append(maxmindInstanceNames, serverBackendMIG2InstanceNames...)
					}
				}

				if err := gcpStorage.CopyFromBytesToStorage(ctx, buf.Bytes(), cityStorageName); err != nil {
					level.Error(logger).Log("msg", "failed to copy maxmind City file to gcp cloud storage", "err", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindStorageUploadFailureCity.Add(1)
				}

				if err := gcpStorage.CopyFromBytesToRemote(buf.Bytes(), maxmindInstanceNames, cityFileName); err != nil {
					level.Error(logger).Log("msg", "failed to copy maxmind city file to server backends", "err", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.MaxmindSCPWriteFailure.Add(1)
					continue
				}

				relayPusherServiceMetrics.RelayPusherMetrics.MaxmindSuccessfulCityStorageUploads.Add(1)

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
	go func() {
		for {
			select {
			case <-dbTicker.C:
				start := time.Now()

				// Store the known list of instance names
				databaseInstanceNames := remoteDBLocations

				// The names of instances in a MIG can change, so get them each time
				relayGatewayMIGInstanceNames, err := getMIGInstanceNames(gcpProjectID, relayGatewayMIGName)
				if err != nil {
					level.Error(logger).Log("msg", "failed to fetch relay gateway mig instance names", "err", err)
				} else {
					// Add the gateway mig instance names to the list
					databaseInstanceNames = append(databaseInstanceNames, relayGatewayMIGInstanceNames...)
				}

				apiMIGInstanceNames, err := getMIGInstanceNames(gcpProjectID, apiMIGName)
				if err != nil {
					level.Error(logger).Log("msg", "failed to fetch api mig instance names", "err", err)
				} else {
					// Add the gateway mig instance names to the list
					databaseInstanceNames = append(databaseInstanceNames, apiMIGInstanceNames...)
				}

				if err := gcpStorage.CopyFromBucketToRemote(ctx, databaseBinFileName, databaseInstanceNames, databaseBinFileOutputLocation); err != nil {
					level.Error(logger).Log("msg", "failed to copy database bin file to database locations", "err", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.DatabaseSCPWriteFailure.Add(1)
					continue
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

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		level.Debug(logger).Log("msg", "Received shutdown signal")
		fmt.Println("Received shutdown signal.")

		cancel()
		// Wait for essential goroutines to finish up
		wg.Wait()

		level.Debug(logger).Log("msg", "Successfully shutdown")
		fmt.Println("Successfully shutdown.")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
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
