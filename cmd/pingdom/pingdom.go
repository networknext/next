package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/pingdom"

	"cloud.google.com/go/bigquery"
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {

	ctx, cancel := context.WithCancel(context.Background())

	// DEBUG: set default value to ""
	pingdomApiToken := envvar.Get("PINGDOM_API_TOKEN", "nEVIN5R8WjzcRDEO6FA6wMwjBgla63NqNbqZ2dX-D8TTtPUF_3sGAqg_d0db2OPxkdCdMh8")
	if pingdomApiToken == "" {
		core.Error("PINGDOM_API_TOKEN not set")
		return 1
	}

	gcpProjectID := backend.GetGCPProjectID()
	// if gcpProjectID == "" {
	// 	core.Error("pingdom must be run in the cloud because requires BigQuery read/write access")
	// 	return 1
	// }

	// DEBUG
	gcpProjectID = "network-next-v3-dev"

	bqClient, err := bigquery.NewClient(context.Background(), gcpProjectID)
	if err != nil {
		core.Error("failed to create BigQuery client: %v", err)
		return 1
	}

	// DEBUG: set default value to ""
	bqDatasetName := envvar.Get("GOOGLE_BIGQUERY_DATASET_PINGDOM", "dev")
	if bqDatasetName == "" {
		core.Error("GOOGLE_BIGQUERY_DATASET_PINGDOM not set")
		return 1
	}

	// DEBUG: set default value to ""
	bqTableName := envvar.Get("GOOGLE_BIGQUERY_TABLE_PINGDOM", "pingdom")
	if bqTableName == "" {
		core.Error("GOOGLE_BIGQUERY_TABLE_PINGDOM not set")
		return 1
	}

	chanSize, err := envvar.GetInt("PINGDOM_CHANNEL_SIZE", 100)
	if err != nil {
		core.Error("failed to parse PINGDOM_CHANNEL_SIZE: %v", err)
		return 1
	}

	pingdomClient, err := pingdom.NewPingdomClient(pingdomApiToken, bqClient, gcpProjectID, bqDatasetName, bqTableName, chanSize)
	if err != nil {
		core.Error("failed to create pingdom client: %v", err)
		return 1
	}

	// DEBUG: set default value to ""
	portalHostname := envvar.Get("PORTAL_HOSTNAME", "portal.networknext.com")
	if portalHostname == "" {
		core.Error("PORTAL_HOSTNAME not set")
		return 1
	}

	// DEBUG: set default value to ""
	serverBackendHostname := envvar.Get("SERVER_BACKEND_HOSTNAME", "server_backend.prod.networknext.com")
	if serverBackendHostname == "" {
		core.Error("SERVER_BACKEND_HOSTNAME not set")
		return 1
	}

	portalID, err := pingdomClient.GetIDForHostname(portalHostname)
	if err != nil {
		core.Error("failed to get portal pingdom ID: %v", err)
		return 1
	}

	serverBackendID, err := pingdomClient.GetIDForHostname(serverBackendHostname)
	if err != nil {
		core.Error("failed to get server backend pingdom ID: %v", err)
		return 1
	}

	pingFrequency, err := envvar.GetDuration("PINGDOM_API_PING_FREQUENCY", time.Second*10)
	if err != nil {
		core.Error("failed to parse PINGDOM_API_PING_FREQUENCY: %v", err)
		return 1
	}

	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	// Start the goroutine for calculating uptime from the Pingdom API
	wg.Add(1)
	go pingdomClient.GetUptimeForIDs(ctx, portalID, serverBackendID, pingFrequency, &wg, errChan)

	// Start the goroutine for inserting uptime data to BigQuery
	wg.Add(1)
	go pingdomClient.WriteLoop(ctx, &wg)

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		core.Debug("received shutdown signal")
		cancel()

		// Wait for essential goroutines to finish up
		wg.Wait()

		core.Debug("successfully shutdown")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		cancel()
		return 1
	}
}
