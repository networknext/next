/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"expvar"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/backend"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/relay_gateway"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-kit/kit/log/level"
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
	serviceName := "relay_gateway"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	ctx := context.Background()

	gcpProjectID := backend.GetGCPProjectID()

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	//todo why 2 loggers
	relaysLogger, err := backend.GetLogger(ctx, gcpProjectID, "relays")
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

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	relayMetrics, err, msg := relay_gateway.NewMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", msg, "err", err)
		return 1
	}

	storer, err := backend.GetStorer(ctx, logger, gcpProjectID, env)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	cfg, err := relay_gateway.NewConfig()
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	relayStore, err := storage.NewRedisRelayStore(cfg.RelayStoreAddress, cfg.RelayStoreReadTimeout, cfg.RelayStoreWriteTimeout, cfg.RelayStoreRelayTimeout)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}


	publishers, err := pubsub.NewMultiPublisher([]string{cfg.PublishToHosts}, cfg.PublisherSendBuffer)

	gateway := &relay_gateway.Gateway{
		Cfg: cfg,
		Logger: logger,
		RelayLogger: relaysLogger,
		Metrics: relayMetrics,
		Publishers: publishers,
		Store: &storer,
		RelayStore: relayStore,
		RelayCache: storage.NewRelayCache(),
		ShutdownSvc: false,
	}
	
	go func() {
	err = gateway.RelayCacheRunner()
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
	}()

	fmt.Printf("starting http server\n")
	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))
	router.HandleFunc("/relay_init", gateway.RelayInitHandlerFunc()).Methods("POST")
	router.HandleFunc("/relay_update", gateway.RelayUpdateHandlerFunc()).Methods("POST")
	router.Handle("/debug/vars", expvar.Handler())

	fmt.Println("starting Http")
	go func() {
		port := envvar.Get("PORT", "30000")

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
