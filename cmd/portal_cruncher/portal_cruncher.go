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
	"github.com/networknext/backend/modules/config"
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

	btMetrics, err := metrics.NewBigTableMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create bigtable metrics", "err", err)
		return 1
	}

	// Setup feature config for bigtable
	var featureConfig config.Config
	envVarConfig := config.NewEnvVarConfig([]config.Feature{
		{
			Name:        "FEATURE_BIGTABLE",
			Value:       false,
			Description: "Bigtable integration for historic session data",
		},
	})
	featureConfig = envVarConfig

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
				fmt.Printf("%d bigtable success meta writes\n", int(btMetrics.WriteMetaSuccessCount.Value()))
				fmt.Printf("%d bigtable success slice writes\n", int(btMetrics.WriteSliceSuccessCount.Value()))
				fmt.Printf("%d bigtable failed meta writes\n", int(btMetrics.WriteMetaFailureCount.Value()))
				fmt.Printf("%d bigtable failed slice writes\n", int(btMetrics.WriteSliceFailureCount.Value()))
				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second * 10)
			}
		}()
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

	redisPingFrequency, err := envvar.GetDuration("CRUNCHER_REDIS_PING_FREQUENCY", time.Second*30)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	redisFlushFrequency, err := envvar.GetDuration("CRUNCHER_REDIS_FLUSH_FREQUENCY", time.Second)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	redisFlushCount, err := envvar.GetInt("PORTAL_CRUNCHER_REDIS_FLUSH_COUNT", 1000)
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

	// Determine if should insert into Bigtable
	useBigtable := featureConfig.FeatureEnabled(0)

	// Get Bigtable instance ID
	btInstanceID := envvar.Get("BIGTABLE_INSTANCE_ID", "")
	// Get the table name
	btTableName := envvar.Get("BIGTABLE_TABLE_NAME", "")
	// Get the column family name
	btCfName := envvar.Get("BIGTABLE_CF_NAME", "")
	// Get the max number of days the data should be kept in Bigtable
	btMaxAgeDays, err := envvar.GetInt("BIGTABLE_MAX_AGE_DAYS", 90)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	btGoroutineCount, err := envvar.GetInt("BIGTABLE_CRUNCHER_GOROUTINE_COUNT", 1)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	portalCruncher, err := portalcruncher.NewPortalCruncher(ctx,
															portalSubscriber,
															redisHostTopSessions,
															redisHostSessionMap,
															redisHostSessionMeta,
															redisHostSessionSlices,
															useBigtable,
															gcpProjectID,
															btInstanceID,
															btTableName,
															btCfName,
															btMaxAgeDays,
															messageChanSize,
															logger,
															portalCruncherMetrics,
															btMetrics)
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
		if err := portalCruncher.Start(ctx, redisGoroutineCount, btGoroutineCount, redisPingFrequency, redisFlushFrequency, redisFlushCount); err != nil {
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
