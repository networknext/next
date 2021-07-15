/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
)

func main() {
	fmt.Printf("not today\n")
}

// todo: not today
/*
package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"time"

	gcpPub "cloud.google.com/go/pubsub"
	gcStorage "cloud.google.com/go/storage"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/backend" // todo: bad name for module
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/optimizer"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
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
	serviceName := "optimizer"
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

	matrixStore, err := storage.NewRedisMatrixStore(cfg.RelayStoreAddress, cfg.RelayStoreReadTimeout, cfg.RelayStoreWriteTimeout, 0)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	svc := optimizer.NewBaseOptimizer(cfg)
	svc.Metrics = metrics
	svc.MatrixStore = matrixStore
	svc.RelayMap = relayMap
	svc.RelayStore = relayStore
	svc.Store = storer
	svc.StatsDB = statsDB
	svc.Logger = logger

	if cfg.CloudStoreActive {
		client, err := gcStorage.NewClient(ctx)
		if err != nil {
			level.Error(logger).Log("err", err)
		}
		bkt := client.Bucket(fmt.Sprintf("%s-matrices", gcpProjectID))
		err = bkt.Create(ctx, gcpProjectID, nil)
		svc.CloudBucket = bkt
	}

	go func() {
		err = svc.RelayCacheRunner()
		if err != nil {
			level.Error(logger).Log("error", err)
			os.Exit(1)
		}
	}()

	//sleep for 10 seconds to allow the first pass of the cacheRunner to finish populating the relay map
	time.Sleep(10 * time.Second)

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

			if cfg.CloudStoreActive {
				timeStamp := time.Now().UTC()
				err := svc.CloudStoreMatrix("cost", timeStamp, costMatrix.GetResponseData())
				if err != nil {
					level.Error(logger).Log("err", err)
				}
				err = svc.CloudStoreMatrix("route", timeStamp, routeMatrix.GetResponseData())
				if err != nil {
					level.Error(logger).Log("err", err)
				}
			}

			if !cfg.CloudStoreOnly {
				svc.UpdateMatrix(*routeMatrix, storage.MatrixTypeNormal)
			}
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

			if !cfg.CloudStoreOnly {
				svc.UpdateMatrix(*routeMatrix, storage.MatrixTypeValve)
			}
		}
	}()

	//Publishers
	{
		emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
		if gcpProjectID != "" || emulatorOK {

			pubsubCtx := ctx
			if emulatorOK {
				gcpProjectID = "local"

				var cancelFunc context.CancelFunc
				pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(60*time.Minute))
				defer cancelFunc()

				level.Info(svc.Logger).Log("msg", "Detected pubsub emulator")
			}

			// Google Pubsub settings
			settings := gcpPub.PublishSettings{
				DelayThreshold: time.Second,
				CountThreshold: 1,
				ByteThreshold:  1 << 14,
				NumGoroutines:  runtime.GOMAXPROCS(0),
				Timeout:        time.Minute,
			}

			pingStatsPublisher, err := analytics.NewGooglePubSubPingStatsPublisher(pubsubCtx, &svc.Metrics.RelayBackendMetrics.PingStatsMetrics, svc.Logger, gcpProjectID, "ping_stats", settings)
			if err != nil {
				level.Error(logger).Log("msg", "could not create analytics pubsub publisher", "err", err)
				return 1
			}

			psPublishInterval, err := envvar.GetDuration("PING_STATS_PUBLISH_INTERVAL", time.Minute)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
			}

			go func() {
				err := svc.PingPublishRunner(pingStatsPublisher, ctx, psPublishInterval)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
				}
			}()

			relayStatsPublisher, err := analytics.NewGooglePubSubRelayStatsPublisher(pubsubCtx, &svc.Metrics.RelayBackendMetrics.RelayStatsMetrics, svc.Logger, gcpProjectID, "relay_stats", settings)
			if err != nil {
				level.Error(logger).Log("msg", "could not create analytics pubsub publisher", "err", err)
				return 1
			}

			rsPublishInterval, err := envvar.GetDuration("RELAY_STATS_PUBLISH_INTERVAL", time.Second*10)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
			}

			go func() {
				err := svc.RelayPublishRunner(relayStatsPublisher, ctx, rsPublishInterval)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
				}
			}()
		}
	}

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.Handle("/debug/vars", expvar.Handler())

	enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	if enablePProf {
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

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
*/
