/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"expvar"
	"fmt"
	"github.com/networknext/backend/optimizer"
	"github.com/networknext/backend/storage"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/networknext/backend/backend"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/transport"

	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/routing"
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
	serviceName := "Optimizer"
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

	if gcpProjectID != "" {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
			return 1
		}
	}

	storer, err := backend.GetStorer(ctx, logger, gcpProjectID, env)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	cfg, err := optimizer.GetConfig()
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	metrics, err, msg := optimizer.NewMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", msg, "err", err)
	}

	statsDB := routing.NewStatsDatabase()

	// Create the relay map
	cleanupCallback := func(relayData *routing.RelayData) error {
		// Remove relay entry from statsDB (which in turn means it won't appear in cost matrix)
		statsDB.DeleteEntry(relayData.ID)
		level.Warn(logger).Log("msg", "relay timed out", "relay ID", relayData.ID, "relay addr", relayData.Addr.String(), "relay name", relayData.Name)
		return nil
	}

	relayMap := routing.NewRelayMap(cleanupCallback)
	go func() {
		timeout := int64(routing.RelayTimeout.Seconds())
		frequency := time.Second
		ticker := time.NewTicker(frequency)
		relayMap.TimeoutLoop(ctx, timeout, ticker.C)
	}()

	relayStore, err := storage.NewRedisRelayStore(cfg.RelayStoreAddress, cfg.RelayStoreReadTimeout, cfg.RelayStoreWriteTimeout, cfg.RelayStoreRelayTimeout)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	svc := optimizer.NewBaseOptimizer(cfg)
	svc.Metrics = metrics
	svc.RelayMap = relayMap
	svc.RelayStore = relayStore
	svc.Store = storer
	svc.StatsDB = statsDB
	svc.Logger = logger

	go func() {
	err = svc.RelayCacheRunner()
	if err != nil{
		level.Error(logger).Log("error", err)
		os.Exit(1)
	}
	}()
	go func() {
		err = svc.StartSubscriber()
		if err != nil {
			level.Error(logger).Log("error", err)
			os.Exit(1)
		}
	}()


	syncInterval, err := envvar.GetDuration("COST_MATRIX_INTERVAL", time.Second)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	costMatrixData := new(helpers.MatrixData)
	routeMatrixData := new(helpers.MatrixData)
	valveRouteMatrixData := new(helpers.MatrixData)

	// Generate the route matrix
	go func() {
		syncTimer := helpers.NewSyncTimer(syncInterval)
		for {

			syncTimer.Run()

			costMatrix, routeMatrix := svc.GetRouteMatrix()
			if costMatrix == nil || routeMatrix == nil {
				continue
			}

			costMatrixData.SetMatrix(costMatrix.GetResponseData())
			routeMatrixData.SetMatrix(routeMatrix.GetResponseData())

			svc.UpdateMatrix(*routeMatrix, storage.MatrixTypeNormal)

			svc.MetricsOutput()
		}
	}()

	// Generate the route matrix specifically for valve
	go func() {
		syncTimer := helpers.NewSyncTimer(syncInterval)
		for {
			syncTimer.Run()

			_, routeMatrix := svc.GetValveRouteMatrix()
			if routeMatrix == nil {
				continue
			}

			valveRouteMatrixData.SetMatrix(routeMatrix.GetResponseData())

			svc.UpdateMatrix(*routeMatrix, storage.MatrixTypeValve)
		}
	}()

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))
	router.Handle("/debug/vars", expvar.Handler())

	//todo fix getRouteMatrixFunc??
	//router.HandleFunc("/relay_dashboard", transport.RelayDashboardHandlerFunc(relayMap, getRouteMatrixFunc, statsdb, "local", "local", maxJitter))
	router.HandleFunc("/relay_stats", transport.RelayStatsFunc(logger, relayMap))

	go func() {
		port := envvar.Get("PORT", "30005")

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

