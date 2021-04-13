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
	"runtime"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	gateway "github.com/networknext/backend/modules/relay_gateway"
	"github.com/networknext/backend/modules/transport"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
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

	// Setup the service
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

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	gatewayMetrics, err := metrics.NewRelayGatewayMetrics(ctx, metricsHandler, serviceName, "relay_gateway", "Relay Gateway", "relay update request")
	if err != nil {
		level.Error(logger).Log("msg", "could not create gateway metrics", "err", err)
		return 1
	}

	// Get a config for how the Gateway should operate
	cfg, err := newConfig()
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Create an error channel for goroutines
	errChan := make(chan error, 1)

	// Create a channel to hold incoming relay update requests
	updateChan := make(chan []byte, cfg.ChannelBufferSize)

	// Prioritize using HTTP to batch-send updates to relay backends
	if cfg.UseHTTP {
		// Create a Gateway HTTP Client
		gatewayHTTPClient, err := gateway.NewGatewayHTTPClient(cfg, updateChan, gatewayMetrics, logger)
		if err != nil {
			level.Error(logger).Log("msg", "could not create gateway http client", "err", err)
			return 1
		}

		go func() {
			// Start up goroutins to POST to relay backends
			if err := gatewayHTTPClient.Start(ctx); err != nil {
				level.Error(logger).Log("err", err)
				errChan <- err
			}
		}()

	} else {
		// TODO: implement ZeroMQ functionality
		level.Error(logger).Log("err", "ZeroMQ is not yet supported")
		return 1

		// // Use ZeroMQ to publish updates to relay backend
		// var publishers []pubsub.Publisher
		// refreshPubs := make(chan bool, 1)
		// publishers, err := pubsub.NewMultiPublisher(cfg.PublishToHosts, cfg.PublisherSendBuffer)
		// if err != nil {
		//     level.Error(logger).Log("err", err)
		//     os.Exit(1)
		// }

		// go func() {
		//     syncTimer := helpers.NewSyncTimer(cfg.PublisherRefreshTimer)
		//     for {
		//         syncTimer.Run()
		//         refreshPubs <- true
		//     }
		// }()

		// go func() {
		//     for {
		//         select {
		//         case <-refreshPubs:
		//             newPublishers, err := pubsub.NewMultiPublisher(cfg.PublishToHosts, cfg.PublisherSendBuffer)
		//             if err != nil {
		//                 _ = level.Error(logger).Log("err", err)
		//                 continue
		//             }

		//             for _, pub := range publishers {
		//                 err = pub.Close()
		//                 if err != nil {
		//                     _ = level.Error(logger).Log("err", err)
		//                 }
		//             }

		//             publishers = newPublishers

		//             continue

		//         case msg := <-updateChan:
		//             for _, pub := range publishers {
		//                 _, err = pub.Publish(context.Background(), pubsub.RelayUpdateTopic, msg)
		//                 if err != nil {
		//                     _ = level.Error(logger).Log("msg", "unable to send update to optimizer", "err", err)
		//                 }
		//             }
		//         }
		//     }
		// }()
	}

	// Setup the stats print routine
	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				gatewayMetrics.GatewayServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				gatewayMetrics.GatewayServiceMetrics.MemoryAllocated.Set(memoryUsed())

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", int(gatewayMetrics.GatewayServiceMetrics.Goroutines.Value()))
				fmt.Printf("%.2f mb allocated\n", gatewayMetrics.GatewayServiceMetrics.MemoryAllocated.Value())
				fmt.Printf("%d update requests received\n", int(gatewayMetrics.UpdatesReceived.Value()))
				fmt.Printf("%d update requests queued\n", int(gatewayMetrics.UpdatesQueued.Value()))
				fmt.Printf("%d update requests flushed\n", int(gatewayMetrics.UpdatesFlushed.Value()))
				fmt.Printf("%d update request read packet failures\n", int(gatewayMetrics.ErrorMetrics.ReadPacketFailure.Value()))
				fmt.Printf("%d update request content type failures\n", int(gatewayMetrics.ErrorMetrics.ContentTypeFailure.Value()))
				fmt.Printf("%d batch update request marshal binary failures\n", int(gatewayMetrics.ErrorMetrics.MarshalBinaryFailure.Value()))
				fmt.Printf("%d batch update request backend send failures\n", int(gatewayMetrics.ErrorMetrics.BackendSendFailure.Value()))
				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second * 10)
			}
		}()
	}

	fmt.Printf("starting http server\n")
	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/relay_update", transport.GatewayRelayUpdateHandlerFunc(logger, gatewayMetrics, updateChan)).Methods("POST")
	router.Handle("/debug/vars", expvar.Handler())

	enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	if enablePProf {
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

	port := envvar.Get("PORT", "30000")
	fmt.Printf("starting http server on :%s\n", port)

	go func() {
		level.Info(logger).Log("addr", ":"+port)

		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1) // TODO: don't os.Exit() here, but find a way to exit using errChan
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-sigint:
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		// TODO: implement clean shutdown to flush update requests in buffer
		return 1
	}
}

// Get the config for how this relay gateway should operate
func newConfig() (*gateway.GatewayConfig, error) {
	cfg := new(gateway.GatewayConfig)
	// Get the channel size
	channelBufferSize, err := envvar.GetInt("FEATURE_NEW_RELAY_BACKEND_CHANNEL_BUFFER_SIZE", 100000)
	if err != nil {
		return nil, err
	}
	cfg.ChannelBufferSize = channelBufferSize

	// Decide if we are using HTTP to batch-write to relay backends
	useHTTP, err := envvar.GetBool("FEATURE_NEW_RELAY_BACKEND_HTTP", true)
	if err != nil {
		return nil, err
	}
	cfg.UseHTTP = useHTTP

	// Load env vars depending on relay update delivery method
	if useHTTP {
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

		// Get the batch size threshold for sending updates to relay backends
		batchSize, err := envvar.GetInt("FEATURE_NEW_RELAY_BACKEND_BATCH_SIZE", 20)
		if err != nil {
			return nil, err
		}
		cfg.BatchSize = batchSize

		numGoroutines, err := envvar.GetInt("FEATURE_NEW_RELAY_BACKEND_NUM_GOROUTINES", 1)
		if err != nil {
			return nil, err
		}
		cfg.NumGoroutines = numGoroutines
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

	return cfg, nil
}
