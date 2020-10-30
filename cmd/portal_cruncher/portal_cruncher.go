/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"fmt"
	"runtime"

	"net/http"
	"os"
	"os/signal"

	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/backend"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"

	portalcruncher "github.com/networknext/backend/portal_cruncher"
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
	serviceName := "portal_cruncher"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	ctx := context.Background()

	gcpProjectID := backend.GetGCPProjectID()

	backend.GetLogger(ctx, gcpProjectID, serviceName)

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

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
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

	portalCruncherMetrics, err := metrics.NewPortalCruncherMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create portal cruncher metrics", "err", err)
		return 1
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
				portalCruncherMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				portalCruncherMetrics.MemoryAllocated.Set(memoryUsed())

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", int(portalCruncherMetrics.Goroutines.Value()))
				fmt.Printf("%.2f mb allocated\n", portalCruncherMetrics.MemoryAllocated.Value())
				fmt.Printf("%d messages received\n", int(portalCruncherMetrics.ReceivedMessageCount.Value()))
				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second * 10)
			}
		}()
	}

	redisFlushCount, err := envvar.GetInt("PORTAL_CRUNCHER_REDIS_FLUSH_COUNT", 1000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Start portal cruncher subscriber
	var portalSubscriber pubsub.Subscriber
	{
		cruncherPort := envvar.Get("CRUNCHER_PORT", "5555")
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		receiveBufferSize, err := envvar.GetInt("CRUNCHER_RECEIVE_BUFFER_SIZE", 1000000)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		portalCruncherSubscriber, err := pubsub.NewPortalCruncherSubscriber(cruncherPort, int(receiveBufferSize))
		if err != nil {
			level.Error(logger).Log("msg", "could not create portal cruncher subscriber", "err", err)
			return 1
		}

		if err := portalCruncherSubscriber.Subscribe(pubsub.TopicPortalCruncherSessionData); err != nil {
			level.Error(logger).Log("msg", "could not subscribe to portal cruncher session data topic", "err", err)
			return 1
		}

		if err := portalCruncherSubscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts); err != nil {
			level.Error(logger).Log("msg", "could not subscribe to portal cruncher session counts topic", "err", err)
			return 1
		}

		portalSubscriber = portalCruncherSubscriber
	}

	receiveGoroutineCount, err := envvar.GetInt("CRUNCHER_RECEIVE_GOROUTINE_COUNT", 5)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	redisGoroutineCount, err := envvar.GetInt("CRUNCHER_REDIS_GOROUTINE_COUNT", 5)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	messageChanSize, err := envvar.GetInt("CRUNCHER_MESSAGE_CHANNEL_SIZE", 10000000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	redisHostTopSessions := envvar.Get("REDIS_HOST_TOP_SESSIONS", "127.0.0.1:6379")
	redisHostSessionMap := envvar.Get("REDIS_HOST_SESSION_MAP", "127.0.0.1:6379")
	redisHostSessionMeta := envvar.Get("REDIS_HOST_SESSION_META", "127.0.0.1:6379")
	redisHostSessionSlices := envvar.Get("REDIS_HOST_SESSION_SLICES", "127.0.0.1:6379")

	portalCruncher, err := portalcruncher.NewPortalCruncher(portalSubscriber, redisHostTopSessions, redisHostSessionMap, redisHostSessionMeta, redisHostSessionSlices, messageChanSize, portalCruncherMetrics)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	if err := portalCruncher.PingRedis(); err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	errChan := make(chan error, 1)
	go func() {
		if err := portalCruncher.Start(ctx, receiveGoroutineCount, redisGoroutineCount, time.Second, redisFlushCount); err != nil {
			level.Error(logger).Log("err", err)
			errChan <- err
			return
		}
	}()

	// Start HTTP server
	{
		go func() {
			router := mux.NewRouter()
			router.HandleFunc("/health", transport.HealthHandlerFunc())
			router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))

			port, ok := os.LookupEnv("HTTP_PORT")
			if !ok {
				level.Error(logger).Log("err", "env var HTTP_PORT must be set")
				errChan <- err
				return
			}

			err := http.ListenAndServe(":"+port, router)
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
