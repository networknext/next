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
	"runtime"
	"time"

	"os"
	"os/signal"

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

	ctx := context.Background()

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
	beaconInserterMetrics, err := metrics.NewBeaconInserterMetrics(ctx, metricsHandler)
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

				// Create biquery client and start write loop
				bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
				if err != nil {
					level.Error(logger).Log("err", err)
					return 1
				}

				beaconTable := envvar.Get("GOOGLE_BIGQUERY_TABLE_BEACON", "")

				b := beacon.GoogleBigQueryClient{
					Metrics:       beaconInserterMetrics.PublishMetrics,
					Logger:        logger,
					TableInserter: bqClient.Dataset(beaconDataset).Table(beaconTable).Inserter(),
					BatchSize:     batchSize,
				}

				// Set the Beaconer to BigQuery
				beaconer = &b

				// Start the background WriteLoop to batch write to BigQuery
				go func() {
					b.WriteLoop(ctx)
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
			}

			topicName := "beacon"
			subscriptionName := "beacon"

			pubsubCtx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			pubsubForwarder, err := beacon.NewPubSubForwarder(pubsubCtx, beaconer, logger, beaconInserterMetrics.ReceiveMetrics, gcpProjectID, topicName, subscriptionName)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			go pubsubForwarder.Forward(ctx)
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

				beaconInserterMetrics.ServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				beaconInserterMetrics.ServiceMetrics.MemoryAllocated.Set(memoryUsed())

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", int(beaconInserterMetrics.ServiceMetrics.Goroutines.Value()))
				fmt.Printf("%.2f mb allocated\n", beaconInserterMetrics.ServiceMetrics.MemoryAllocated.Value())
				fmt.Printf("%d beacon entries received\n", int(beaconInserterMetrics.ReceiveMetrics.EntriesReceived.Value()))
				fmt.Printf("%d beacon entries submitted\n", int(beaconInserterMetrics.PublishMetrics.EntriesSubmitted.Value()))
				fmt.Printf("%d beacon entries queued\n", int(beaconInserterMetrics.PublishMetrics.EntriesQueued.Value()))
				fmt.Printf("%d beacon entries flushed\n", int(beaconInserterMetrics.PublishMetrics.EntriesFlushed.Value()))
				fmt.Printf("%d beacon entry read failures\n", int(beaconInserterMetrics.ReceiveMetrics.UnmarshalFailure.Value()))
				fmt.Printf("%d beacon entry write failures\n", int(beaconInserterMetrics.PublishMetrics.PublishFailure.Value()))
				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second * 10)
			}
		}()
	}

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, false, []string{}))
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

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-sigint:
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		return 1
	}
}
