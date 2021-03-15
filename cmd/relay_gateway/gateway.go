/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/pubsub"

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

	fmt.Printf("gcp id: %v \n", gcpProjectID)

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	fmt.Printf("logger 1\n")

	// todo why 2 loggers
	relayLogger, err := backend.GetLogger(ctx, gcpProjectID, "relays")
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	fmt.Printf("logger 2\n")

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
	fmt.Printf("stack driver profiler\n")

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	fmt.Printf("metrics handler\n")
	relayMetrics, err, msg := metrics.NewRelayGatewayMetrics(ctx, metricsHandler)
	if err != nil {
		fmt.Printf("error: %v \n", err.Error())
		level.Error(logger).Log("msg", msg, "err", err)
		return 1
	}

	fmt.Printf("metrics gateway\n")
	storer, err := backend.GetStorer(ctx, logger, gcpProjectID, env)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	fmt.Printf("storer\n")
	cfg, err := newConfig()
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}
	fmt.Printf("config\n")
	var publishers []pubsub.Publisher
	if !cfg.NRBHTTP {
		publishers, err = pubsub.NewMultiPublisher(cfg.PublishToHosts, cfg.PublisherSendBuffer)
		if err != nil {
			level.Error(logger).Log("err", err)
		}
	}

	getParams := func() *transport.GatewayHandlerConfig {
		return &transport.GatewayHandlerConfig{
			Storer:                storer,
			InitMetrics:           relayMetrics.RelayInitMetrics,
			UpdateMetrics:         relayMetrics.RelayUpdateMetrics,
			RouterPrivateKey:      cfg.RouterPrivateKey,
			Publishers:            publishers,
			NRBNoInit:             cfg.NRBNoInit,
			NRBHTTP:               cfg.NRBHTTP,
			RelayBackendAddresses: cfg.RelayBackendAddresses,
			LoadTest:              cfg.Loadtest,
		}
	}

	fmt.Printf("starting http server\n")
	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/relay_init", transport.GatewayRelayInitHandlerFunc(logger, getParams())).Methods("POST")
	router.HandleFunc("/relay_update", transport.GatewayRelayUpdateHandlerFunc(logger, relayLogger, getParams())).Methods("POST")
	router.Handle("/debug/vars", expvar.Handler())

	enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	if enablePProf {
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

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

type Config struct {
	PublisherSendBuffer   int
	PublishToHosts        []string
	RouterPrivateKey      []byte
	NRBNoInit             bool
	NRBHTTP               bool
	RelayBackendAddresses []string
	Loadtest              bool
}

func newConfig() (*Config, error) {
	cfg := new(Config)

	routerPrivateKey, err := envvar.GetBase64("RELAY_ROUTER_PRIVATE_KEY", nil)
	if err != nil {
		return nil, fmt.Errorf("RELAY_ROUTER_PRIVATE_KEY not set")
	}
	cfg.RouterPrivateKey = routerPrivateKey

	nrbNoInit, err := envvar.GetBool("FEATURE_NEW_RELAY_BACKEND_NO_INIT", false)
	if err != nil {
		return nil, err
	}
	cfg.NRBNoInit = nrbNoInit

	nrbHTTP, err := envvar.GetBool("FEATURE_NEW_RELAY_BACKEND_HTTP", false)
	if err != nil {
		return nil, err
	}
	cfg.NRBHTTP = nrbHTTP

	if nrbHTTP {
		if exists := envvar.Exists("FEATURE_NEW_RELAY_BACKEND_ADDRESSES"); !exists {
			return nil, fmt.Errorf("FEATURE_NEW_RELAY_BACKEND_ADDRESSES not set")
		}
		relayBackendAddresses := envvar.GetList("FEATURE_NEW_RELAY_BACKEND_ADDRESSES", []string{})
		cfg.RelayBackendAddresses = relayBackendAddresses
	} else {
		if exists := envvar.Exists("PUBLISH_TO_HOSTS"); !exists {
			return nil, fmt.Errorf("PUBLISH_TO_HOSTS not set")
		}
		publishToHosts := envvar.GetList("PUBLISH_TO_HOSTS", []string{"tcp://127.0.0.1:5555"})
		cfg.PublishToHosts = publishToHosts

		publisherSendBuffer, err := envvar.GetInt("PUBLISHER_SEND_BUFFER", 100000)
		if err != nil {
			return nil, err
		}
		cfg.PublisherSendBuffer = publisherSendBuffer
	}

	featureLoadTest, err := envvar.GetBool("FEATURE_LOAD_TEST", false)
	if err != nil {
		return nil, err
	}
	cfg.Loadtest = featureLoadTest

	return cfg, nil
}
