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
		fmt.Println("GCP project ID not defined")
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

	// Create beacon inserter metrics
	relayPusherServiceMetrics, err := metrics.NewRelayPusherServiceMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create beacon service metrics", "err", err)
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
		level.Error(logger).Log("msg", "gcp bucket not specified", "err")
		return 1
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
		level.Error(logger).Log("msg", "ISP temp file not defined", "err", err)
		return 1
	}

	cityFileName := envvar.Get("MAXMIND_CITY_DB_FILE_NAME", "")
	if cityFileName == "" {
		level.Error(logger).Log("msg", "city temp file not defined", "err", err)
		return 1
	}

	ispURI := envvar.Get("MAXMIND_ISP_DB_URI", "")
	if ispURI == "" {
		level.Error(logger).Log("msg", "maxmind DB ISP location not defined", "err", err)
		return 1
	}

	cityURI := envvar.Get("MAXMIND_CITY_DB_URI", "")
	if cityURI == "" {
		level.Error(logger).Log("msg", "maxmind DB city location not defined", "err", err)
		return 1
	}

	// Call gsutil to copy the tmp file over to the instance
	runnable := exec.Command("gcloud", "compute", "--project", gcpProjectID, "instance-groups", "managed", "list-instances", "server-backend-mig", "--zone", "us-central1-a", "--format", "value(instance)")

	buffer, err := runnable.CombinedOutput()
	if err != nil {
		level.Error(logger).Log("msg", "failed to fetch server backend instance names", "err", err)
		return 1
	}

	serverBackendInstanceNames := strings.Split(string(buffer), "\n")

	// Using the method above causes an empty string to be added at the end of the slice - remove it
	if len(serverBackendInstanceNames) > 0 {
		serverBackendInstanceNames = serverBackendInstanceNames[:len(serverBackendInstanceNames)-1]
	}

	// Setup maxmind download go routine
	maxmindSyncInterval, err := envvar.GetDuration("MAXMIND_SYNC_DB_INTERVAL", time.Hour*24)
	if err != nil {
		level.Error(logger).Log("msg", "maxmind DB sync interval not defined", "err", err)
		return 1
	}

	ticker := time.NewTicker(maxmindSyncInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
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
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.SCPWriteFailure.Add(1)
					continue
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

				if err := gcpStorage.CopyFromBytesToRemote(buf.Bytes(), serverBackendInstanceNames, cityFileName); err != nil {
					level.Error(logger).Log("msg", "failed to copy maxmind city file to server backends", "err", err)
					relayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics.SCPWriteFailure.Add(1)
					continue
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
