/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	_ "net/http/pprof"

    "github.com/networknext/backend/modules/tranpsort"
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

	// todo why 2 loggers
	// relayLogger, err := backend.GetLogger(ctx, gcpProjectID, "relays")
	// if err != nil {
	// 	level.Error(logger).Log("err", err)
	// 	return 1
	// }

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

	relayMetrics, err, msg := metrics.NewRelayGatewayMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", msg, "err", err)
		return 1
	}

	// storer, err := backend.GetStorer(ctx, logger, gcpProjectID, env)
	// if err != nil {
	// 	level.Error(logger).Log("err", err)
	// 	return 1
	// }

	cfg, err := newConfig()
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	updateChan := make(chan []byte, 1000)

    // Prioritize using HTTP to send updates to relay backend
	if cfg.NRBHTTP {
        client := http.Client{Timeout: cfg.HTTPTimeout}
        for {
            msg := <-updateChan
            for _, address := range cfg.RelayBackendAddresses {
                go func(address string, body []byte) {
                    buffer := bytes.NewBuffer(body)
                    resp, err := client.Post(fmt.Sprintf("http://%s/relay_update", address), "application/octet-stream", buffer)
                    if err != nil || resp.StatusCode != http.StatusOK {
                        // Don't exit here because we want to continue sending relay updates
                        _ = level.Error(logger).Log("msg", "unable to send update to relay backend", "err", err)
                    }
                    resp.Body.Close()
                }(address, msg)
            }
        }
    } else {
        // Use ZeroMQ to publish updates to relay backend
        var publishers []pubsub.Publisher
        refreshPubs := make(chan bool, 1)
        publishers, err := pubsub.NewMultiPublisher(cfg.PublishToHosts, cfg.PublisherSendBuffer)
        if err != nil {
            level.Error(logger).Log("err", err)
            os.Exit(1)
        }

        go func() {
            syncTimer := helpers.NewSyncTimer(cfg.PublisherRefreshTimer)
            for {
                syncTimer.Run()
                refreshPubs <- true
            }
        }()

        go func() {
            for {
                select {
                case <-refreshPubs:
                    newPublishers, err := pubsub.NewMultiPublisher(cfg.PublishToHosts, cfg.PublisherSendBuffer)
                    if err != nil {
                        _ = level.Error(logger).Log("err", err)
                        continue
                    }

                    for _, pub := range publishers {
                        err = pub.Close()
                        if err != nil {
                            _ = level.Error(logger).Log("err", err)
                        }
                    }

                    publishers = newPublishers

                    continue

                case msg := <-updateChan:
                    for _, pub := range publishers {
                        _, err = pub.Publish(context.Background(), pubsub.RelayUpdateTopic, msg)
                        if err != nil {
                            _ = level.Error(logger).Log("msg", "unable to send update to optimizer", "err", err)
                        }
                    }
                }
            }
        }()
    }


    // ZeroMQ
	if !cfg.NRBHTTP {

	} else {

	}

	getParams := func() *transport.GatewayHandlerConfig {
		return &transport.GatewayHandlerConfig{
			// Storer:           storer,
			InitMetrics:      relayMetrics.RelayInitMetrics,
			UpdateMetrics:    relayMetrics.RelayUpdateMetrics,
			// RouterPrivateKey: cfg.RouterPrivateKey,
			NRBNoInit:        cfg.NRBNoInit,
			LoadTest:         cfg.Loadtest,
		}
	}

	fmt.Printf("starting http server\n")
	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/relay_init", transport.GatewayRelayInitHandlerFunc(logger, getParams())).Methods("POST")
	router.HandleFunc("/relay_update", transport.GatewayRelayUpdateHandlerFunc(logger, relayLogger, getParams(), updateChan)).Methods("POST")
	router.Handle("/debug/vars", expvar.Handler())

	enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	if enablePProf {
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

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

func newConfig() (*transport.GatewayConfig, error) {
	cfg := new(transport.GatewayConfig)

	// routerPrivateKey, err := envvar.GetBase64("RELAY_ROUTER_PRIVATE_KEY", nil)
	// if err != nil || routerPrivateKey == nil {
	// 	return nil, fmt.Errorf("RELAY_ROUTER_PRIVATE_KEY not set")
	// }
	// cfg.RouterPrivateKey = routerPrivateKey

    // Verify if we are using No Init on Relays
	nrbNoInit, err := envvar.GetBool("FEATURE_NEW_RELAY_BACKEND_NO_INIT", false)
	if err != nil {
		return nil, err
	}
	cfg.NRBNoInit = nrbNoInit

    // Decide if we are using HTTP to batch-write to relay backends
	nrbHTTP, err := envvar.GetBool("FEATURE_NEW_RELAY_BACKEND_HTTP", true)
	if err != nil {
		return nil, err
	}
	cfg.NRBHTTP = nrbHTTP

	if nrbHTTP {
        // Using HTTP, get the relay backend addresses to send relay updates to
		if exists := envvar.Exists("FEATURE_NEW_RELAY_BACKEND_ADDRESSES"); !exists {
			return nil, fmt.Errorf("FEATURE_NEW_RELAY_BACKEND_ADDRESSES not set")
		}
		relayBackendAddresses := envvar.GetList("FEATURE_NEW_RELAY_BACKEND_ADDRESSES", []string{})
		cfg.RelayBackendAddresses = relayBackendAddresses

        // Get the HTTP timeout duration
		httpTimeout, err := envvar.GetDuration("HTTP_TIMEOUT", time.Second)
		if err != nil {
			return nil, err
		}
		cfg.HTTPTimeout = httpTimeout

	} else {
        // Using ZeroMQ Pub/Sub, get the relay backend addresses that will receive messages
		if exists := envvar.Exists("PUBLISH_TO_HOSTS"); !exists {
			return nil, fmt.Errorf("PUBLISH_TO_HOSTS not set")
		}
		publishToHosts := envvar.GetList("PUBLISH_TO_HOSTS", []string{"tcp://127.0.0.1:5555"})
		cfg.PublishToHosts = publishToHosts

        // Get publisher send buffer size
		publisherSendBuffer, err := envvar.GetInt("PUBLISHER_SEND_BUFFER", 100000)
		if err != nil {
			return nil, err
		}
		cfg.PublisherSendBuffer = publisherSendBuffer

        // Get publisher refresh time duration
		publisherRefresh, err := envvar.GetDuration("PUBLISHER_REFRESH_TIMER", 60*time.Second)
		if err != nil {
			return nil, err
		}
		cfg.PublisherRefreshTimer = publisherRefresh
	}

    // Decide if we are using fake relays for a load test
	featureLoadTest, err := envvar.GetBool("FEATURE_LOAD_TEST", false)
	if err != nil {
		return nil, err
	}
	cfg.Loadtest = featureLoadTest

	return cfg, nil
}
