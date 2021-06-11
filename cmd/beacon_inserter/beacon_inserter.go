/*
   Network Next. You control the network.
   Copyright © 2017 - 2021 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/beacon"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"

	"cloud.google.com/go/bigquery"
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
	fmt.Printf("beacon_inserter: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	serviceName := "beacon_inserter"

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	gcpProjectID := backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""

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
	beaconInserterServiceMetrics, err := metrics.NewBeaconInserterServiceMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create beacon service metrics", "err", err)
		return 1
	}

	// Create a no-op beaconer
	var beaconer beacon.Beaconer = &beacon.NoOpBeaconer{}

	if gcpOK {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
			return 1
		}

		// Google Bigquery
		{
			beaconDataset := envvar.Get("GOOGLE_BIGQUERY_DATASET_BEACON", "")
			if beaconDataset != "" {
				batchSize := beacon.DefaultBigQueryBatchSize

				batchSize, err := envvar.GetInt("GOOGLE_BIGQUERY_BATCH_SIZE", beacon.DefaultBigQueryBatchSize)
				if err != nil {
					level.Error(logger).Log("err", err)
					return 1
				}

				// Create biquery client
				// Pass context without cancel to ensure writing continues even past reception of shutdown signal
				bqClient, err := bigquery.NewClient(context.Background(), gcpProjectID)
				if err != nil {
					level.Error(logger).Log("err", err)
					return 1
				}

				beaconTable := envvar.Get("GOOGLE_BIGQUERY_TABLE_BEACON", "")

				b := beacon.GoogleBigQueryClient{
					Metrics:       beaconInserterServiceMetrics.BeaconInserterMetrics,
					Logger:        logger,
					TableInserter: bqClient.Dataset(beaconDataset).Table(beaconTable).Inserter(),
					BatchSize:     batchSize,
				}

				// Set the Beaconer to BigQuery
				beaconer = &b

				// Start the background WriteLoop to batch write to BigQuery
				wg.Add(1)
				go func() {
					b.WriteLoop(ctx, wg)
				}()
			}
		}
	}

	pubsubEmulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
	if gcpOK || pubsubEmulatorOK {
		// Google pubsub forwarder
		{
			if pubsubEmulatorOK {
				gcpProjectID = "local"
				level.Info(logger).Log("msg", "Detected pubsub emulator")
			}

			topicName := "beacon"
			subscriptionName := "beacon"

			numRecvGoroutines, err := envvar.GetInt("NUM_RECEIVE_GOROUTINES", 10)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			pubsubForwarder, err := beacon.NewPubSubForwarder(pubsubCtx, beaconer, logger, beaconInserterServiceMetrics.BeaconInserterMetrics, gcpProjectID, topicName, subscriptionName, numRecvGoroutines)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			wg.Add(1)
			go pubsubForwarder.Forward(ctx, wg)
		}
	}

	// Create error channel to error out from any goroutines
	errChan := make(chan error, 1)

	// Setup the stats print routine
	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {

				beaconInserterServiceMetrics.ServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				beaconInserterServiceMetrics.ServiceMetrics.MemoryAllocated.Set(memoryUsed())

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", int(beaconInserterServiceMetrics.ServiceMetrics.Goroutines.Value()))
				fmt.Printf("%.2f mb allocated\n", beaconInserterServiceMetrics.ServiceMetrics.MemoryAllocated.Value())
				fmt.Printf("%d beacon entries transfered\n", int(beaconInserterServiceMetrics.BeaconInserterMetrics.EntriesTransfered.Value()))
				fmt.Printf("%d beacon entries submitted\n", int(beaconInserterServiceMetrics.BeaconInserterMetrics.EntriesSubmitted.Value()))
				fmt.Printf("%d beacon entries queued\n", int(beaconInserterServiceMetrics.BeaconInserterMetrics.EntriesQueued.Value()))
				fmt.Printf("%d beacon entries flushed\n", int(beaconInserterServiceMetrics.BeaconInserterMetrics.EntriesFlushed.Value()))
				fmt.Printf("%d beacon entry read failures\n", int(beaconInserterServiceMetrics.BeaconInserterMetrics.ErrorMetrics.BeaconInserterReadFailure.Value()))
				fmt.Printf("%d beacon entry batched read failures\n", int(beaconInserterServiceMetrics.BeaconInserterMetrics.ErrorMetrics.BeaconInserterBatchedReadFailure.Value()))
				fmt.Printf("%d beacon entry write failures\n", int(beaconInserterServiceMetrics.BeaconInserterMetrics.ErrorMetrics.BeaconInserterWriteFailure.Value()))
				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second * 10)
			}
		}()
	}

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
		router.Handle("/debug/vars", expvar.Handler())

		go func() {
			httpPort := envvar.Get("PORT", "40002")

			err := http.ListenAndServe(":"+httpPort, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				errChan <- err
				return
			}
		}()
	}

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
