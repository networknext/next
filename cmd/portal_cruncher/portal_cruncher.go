/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"

	"net/http"
	"os"
	"os/signal"
	"strconv"

	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {

	fmt.Printf("portal-cruncher: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	// Configure local logging
	logger := log.NewLogfmtLogger(os.Stdout)

	// Create a no-op metrics handler
	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	// StackDriver Logging
	{
		var enableSDLogging bool
		enableSDLoggingString, ok := os.LookupEnv("ENABLE_STACKDRIVER_LOGGING")
		if ok {
			var err error
			enableSDLogging, err = strconv.ParseBool(enableSDLoggingString)
			if err != nil {
				level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_LOGGING", "msg", "could not parse", "err", err)
				os.Exit(1)
			}
		}

		if enableSDLogging {
			if projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
				loggingClient, err := gcplogging.NewClient(ctx, projectID)
				if err != nil {
					level.Error(logger).Log("msg", "failed to create GCP logging client", "err", err)
					os.Exit(1)
				}

				logger = logging.NewStackdriverLogger(loggingClient, "portal-cruncher")
			}
		}
	}

	{
		switch os.Getenv("BACKEND_LOG_LEVEL") {
		case "none":
			logger = level.NewFilter(logger, level.AllowNone())
		case level.ErrorValue().String():
			logger = level.NewFilter(logger, level.AllowError())
		case level.WarnValue().String():
			logger = level.NewFilter(logger, level.AllowWarn())
		case level.InfoValue().String():
			logger = level.NewFilter(logger, level.AllowInfo())
		case level.DebugValue().String():
			logger = level.NewFilter(logger, level.AllowDebug())
		default:
			logger = level.NewFilter(logger, level.AllowWarn())
		}

		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	// Get env
	env, ok := os.LookupEnv("ENV")
	if !ok {
		level.Error(logger).Log("err", "ENV not set")
		os.Exit(1)
	}

	redisPortalHosts := os.Getenv("REDIS_HOST_PORTAL")
	splitPortalHosts := strings.Split(redisPortalHosts, ",")
	redisClientPortal := storage.NewRedisClient(splitPortalHosts...)
	if err := redisClientPortal.Ping().Err(); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_PORTAL", "value", redisPortalHosts, "msg", "could not ping", "err", err)
		os.Exit(1)
	}

	redisPortalHostExp, err := time.ParseDuration(os.Getenv("REDIS_HOST_PORTAL_EXPIRATION"))
	if err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_PORTAL_EXPIRATION", "msg", "could not parse", "err", err)
		os.Exit(1)
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	gcpProjectID, gcpOK := os.LookupEnv("GOOGLE_PROJECT_ID")
	if gcpOK {

		// StackDriver Metrics
		{
			var enableSDMetrics bool
			var err error
			enableSDMetricsString, ok := os.LookupEnv("ENABLE_STACKDRIVER_METRICS")
			if ok {
				enableSDMetrics, err = strconv.ParseBool(enableSDMetricsString)
				if err != nil {
					level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_METRICS", "msg", "could not parse", "err", err)
					os.Exit(1)
				}
			}

			if enableSDMetrics {
				// Set up StackDriver metrics
				sd := metrics.StackDriverHandler{
					ProjectID:          gcpProjectID,
					OverwriteFrequency: time.Second,
					OverwriteTimeout:   10 * time.Second,
				}

				if err := sd.Open(ctx); err != nil {
					level.Error(logger).Log("msg", "Failed to create StackDriver metrics client", "err", err)
					os.Exit(1)
				}

				metricsHandler = &sd

				sdwriteinterval := os.Getenv("GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL")
				writeInterval, err := time.ParseDuration(sdwriteinterval)
				if err != nil {
					level.Error(logger).Log("envvar", "GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL", "value", sdwriteinterval, "err", err)
					os.Exit(1)
				}
				go func() {
					metricsHandler.WriteLoop(ctx, logger, writeInterval, 200)
				}()
			}
		}

		// StackDriver Profiler
		{
			var enableSDProfiler bool
			var err error
			enableSDProfilerString, ok := os.LookupEnv("ENABLE_STACKDRIVER_PROFILER")
			if ok {
				enableSDProfiler, err = strconv.ParseBool(enableSDProfilerString)
				if err != nil {
					level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_PROFILER", "msg", "could not parse", "err", err)
					os.Exit(1)
				}
			}

			if enableSDProfiler {
				// Set up StackDriver profiler
				if err := profiler.Start(profiler.Config{
					Service:        "portal_cruncher",
					ServiceVersion: env,
					ProjectID:      gcpProjectID,
					MutexProfiling: true,
				}); err != nil {
					level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
					os.Exit(1)
				}
			}
		}
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

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", runtime.NumGoroutine())
				fmt.Printf("%.2f mb allocated\n", memoryUsed())
				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second * 10)
			}
		}()
	}

	// Start portal cruncher subscriber
	var portalSubscriber pubsub.Subscriber
	{
		cruncherPort, ok := os.LookupEnv("CRUNCHER_PORT")
		if !ok {
			level.Error(logger).Log("err", "env var CRUNCHER_PORT must be set")
			os.Exit(1)
		}

		portalCruncherSubscriber, err := pubsub.NewPortalCruncherSubscriber(cruncherPort)
		if err != nil {
			level.Error(logger).Log("msg", "could not create portal cruncher subscriber", "err", err)
			os.Exit(1)
		}

		if err := portalCruncherSubscriber.Subscribe(pubsub.TopicPortalCruncherSessionData); err != nil {
			level.Error(logger).Log("msg", "could not subscribe to portal cruncher session data topic", "err", err)
			os.Exit(1)
		}

		if err := portalCruncherSubscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts); err != nil {
			level.Error(logger).Log("msg", "could not subscribe to portal cruncher session counts topic", "err", err)
			os.Exit(1)
		}

		portalSubscriber = portalCruncherSubscriber
	}

	// Start receive loop
	go func() {
		for {
			topic, message, err := portalSubscriber.ReceiveMessage()
			if err != nil {
				level.Error(logger).Log("msg", "error receiving message", "err", err)
				continue
			}

			switch topic {
			case pubsub.TopicPortalCruncherSessionData:
				var sessionData routing.SessionData
				if err := sessionData.UnmarshalBinary(message); err != nil {
					level.Error(logger).Log("msg", "error unmarshaling session data message", "err", err)
					continue
				}

				fmt.Println("received session data")
				fmt.Printf("%#v\n", sessionData)

				tx := redisClientPortal.TxPipeline()

				// set session and slice information with expiration on the entire key set for safety
				tx.Set(fmt.Sprintf("session-%s-meta", sessionData.Meta.ID), sessionData.Meta, redisPortalHostExp)
				tx.SAdd(fmt.Sprintf("session-%s-slices", sessionData.Meta.ID), sessionData.Slice)
				tx.Expire(fmt.Sprintf("session-%s-slices", sessionData.Meta.ID), redisPortalHostExp)

				// set the user session reverse lookup sets with expiration on the entire key set for safety
				tx.SAdd(fmt.Sprintf("user-%s-sessions", sessionData.Meta.UserHash), sessionData.Meta.ID)
				tx.Expire(fmt.Sprintf("user-%s-sessions", sessionData.Meta.UserHash), redisPortalHostExp)

				// set the map point key and buyer sessions with expiration on the entire key set for safety
				tx.Set(fmt.Sprintf("session-%s-point", sessionData.Meta.ID), sessionData.Point, redisPortalHostExp)
				tx.SAdd(fmt.Sprintf("map-points-%s-buyer", sessionData.Meta.BuyerID), sessionData.Meta.ID)
				tx.Expire(fmt.Sprintf("map-points-%s-buyer", sessionData.Meta.BuyerID), redisPortalHostExp)

				if _, err := tx.Exec(); err != nil {
					level.Error(logger).Log("msg", "error sending session data to redis", "err", err)
					continue
				}

			case pubsub.TopicPortalCruncherSessionCounts:
				var countData routing.SessionCountData
				if err := countData.UnmarshalBinary(message); err != nil {
					level.Error(logger).Log("msg", "error unmarshaling session count message", "err", err)
					continue
				}

				fmt.Println("received count data")
				fmt.Printf("%#v\n", countData)

				tx := redisClientPortal.TxPipeline()

				tx.Set("total-direct-count", countData.TotalNumDirectSessions, redisPortalHostExp)
				tx.Set("total-next-count", countData.TotalNumNextSessions, redisPortalHostExp)

				for buyerID, count := range countData.NumDirectSessionsPerBuyer {
					tx.Set(fmt.Sprintf("direct-count-buyer-%016x", buyerID), count, redisPortalHostExp)
				}

				for buyerID, count := range countData.NumNextSessionsPerBuyer {
					tx.Set(fmt.Sprintf("next-count-buyer-%016x", buyerID), count, redisPortalHostExp)
				}

				if _, err := tx.Exec(); err != nil {
					level.Error(logger).Log("msg", "error sending session count data to redis", "err", err)
					continue
				}
			}
		}
	}()

	// Start HTTP server
	{
		go func() {
			router := mux.NewRouter()
			router.HandleFunc("/health", HealthHandlerFunc())
			router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))

			port, ok := os.LookupEnv("HTTP_PORT")
			if !ok {
				level.Error(logger).Log("err", "env var HTTP_PORT must be set")
				os.Exit(1)
			}

			level.Info(logger).Log("addr", ":"+port)

			err := http.ListenAndServe(":"+port, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}
		}()
	}

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}

func HealthHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		statusCode := http.StatusOK

		w.WriteHeader(statusCode)
		w.Write([]byte(http.StatusText(statusCode)))
	}
}
