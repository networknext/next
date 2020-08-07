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
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/metrics"
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

	portalCruncherMetrics, err := metrics.NewPortalCruncherMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create portal cruncher metrics", "err", err)
	}

	// Setup the stats print routine
	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		portalCruncherMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
		portalCruncherMetrics.MemoryAllocated.Set(memoryUsed())

		go func() {
			for {

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", int(portalCruncherMetrics.Goroutines.Value()))
				fmt.Printf("%.2f mb allocated\n", portalCruncherMetrics.MemoryAllocated.Value())
				fmt.Printf("%d messages received\n", int(portalCruncherMetrics.ReceivedMessageCount.Value()))
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

		receiveBufferSizeString, ok := os.LookupEnv("CRUNCHER_RECEIVE_BUFFER_SIZE")
		if !ok {
			level.Error(logger).Log("err", "env var CRUNCHER_RECEIVE_BUFFER_SIZE must be set")
			os.Exit(1)
		}

		receiveBufferSize, err := strconv.ParseInt(receiveBufferSizeString, 10, 64)
		if err != nil {
			level.Error(logger).Log("envvar", "CRUNCHER_RECEIVE_BUFFER_SIZE", "msg", "could not parse", "err", err)
			os.Exit(1)
		}

		portalCruncherSubscriber, err := pubsub.NewPortalCruncherSubscriber(cruncherPort, int(receiveBufferSize))
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

	// Sub to expiry events for cleanup
	{
		redisClientPortal.ConfigSet("notify-keyspace-events", "Ex")
		go func() {
			ps := redisClientPortal.Subscribe("__keyevent@0__:expired")
			for {
				// Receive expiry event message
				msg, err := ps.ReceiveMessage()
				if err != nil {
					level.Error(logger).Log("msg", "Error receiving expired message from redis pubsub", "err", err)
					os.Exit(1)
				}

				// If it is a total direct session count that is expiring...
				if strings.HasPrefix(msg.Payload, "session-count-total-direct-") {
					// Remove the total direct session count from the hash
					if err := redisClientPortal.HDel("session-count-total-direct", msg.Payload).Err(); err != nil {
						level.Error(logger).Log("msg", "failed to remove hashmap entry for total direct session count", "err", err)
						os.Exit(1)
					}
				}

				// If it is a total next session count that is expiring...
				if strings.HasPrefix(msg.Payload, "session-count-total-next-") {
					// Remove the total next session count from the hash
					if err := redisClientPortal.HDel("session-count-total-next", msg.Payload).Err(); err != nil {
						level.Error(logger).Log("msg", "failed to remove hashmap entry for total next session count", "err", err)
						os.Exit(1)
					}
				}

				// If it is a buyer direct session count that is expiring...
				if strings.HasPrefix(msg.Payload, "session-count-direct-buyer-") {
					// Get the buyer ID
					buyerID, err := strconv.ParseUint(strings.TrimPrefix(msg.Payload, "session-count-direct-buyer-")[:16], 16, 64)
					if err != nil {
						level.Error(logger).Log("msg", "failed to parse buyer ID from expired direct buyer session count", "payload", msg.Payload, "err", err)
						os.Exit(1)
					}

					// Remove the buyer direct session count from the hash
					if err := redisClientPortal.HDel(fmt.Sprintf("session-count-direct-buyer-%016x", buyerID), msg.Payload).Err(); err != nil {
						level.Error(logger).Log("msg", "failed to remove hashmap entry for direct buyer session count", "err", err)
						os.Exit(1)
					}
				}

				// If it is a buyer next session count that is expiring...
				if strings.HasPrefix(msg.Payload, "session-count-next-buyer-") {
					// Get the buyer ID
					buyerID, err := strconv.ParseUint(strings.TrimPrefix(msg.Payload, "session-count-next-buyer-")[:16], 16, 64)
					if err != nil {
						level.Error(logger).Log("msg", "failed to parse buyer ID from expired next buyer session count", "payload", msg.Payload, "err", err)
						os.Exit(1)
					}

					// Remove the buyer next session count from the hash
					if err := redisClientPortal.HDel(fmt.Sprintf("session-count-next-buyer-%016x", buyerID), msg.Payload).Err(); err != nil {
						level.Error(logger).Log("msg", "failed to remove hashmap entry for next buyer session count", "err", err)
						os.Exit(1)
					}
				}
			}
		}()
	}

	receiveGoroutineCount := int64(1)
	receiveGoroutineCountString, ok := os.LookupEnv("CRUNCHER_RECEIVE_GOROUTINE_COUNT")
	if ok {
		receiveGoroutineCount, err = strconv.ParseInt(receiveGoroutineCountString, 10, 64)
		if err != nil {
			level.Error(logger).Log("envvar", "CRUNCHER_RECEIVE_GOROUTINE_COUNT", "msg", "could not parse", "err", err)
			os.Exit(1)
		}
	}

	redisGoroutineCount := int64(1)
	redisGoroutineCountString, ok := os.LookupEnv("CRUNCHER_REDIS_GOROUTINE_COUNT")
	if ok {
		redisGoroutineCount, err = strconv.ParseInt(redisGoroutineCountString, 10, 64)
		if err != nil {
			level.Error(logger).Log("envvar", "CRUNCHER_REDIS_GOROUTINE_COUNT", "msg", "could not parse", "err", err)
			os.Exit(1)
		}
	}

	messageChanSize := int64(10000000)
	messageChanSizeString, ok := os.LookupEnv("CRUNCHER_MESSAGE_CHANNEL_SIZE")
	if ok {
		messageChanSize, err = strconv.ParseInt(messageChanSizeString, 10, 64)
		if err != nil {
			level.Error(logger).Log("envvar", "CRUNCHER_MESSAGE_CHANNEL_SIZE", "msg", "could not parse", "err", err)
			os.Exit(1)
		}
	}

	messageChan := make(chan struct {
		topic   pubsub.Topic
		message []byte
	}, messageChanSize)

	// Start receive loops
	for i := int64(0); i < receiveGoroutineCount; i++ {
		go func() {
			for {
				topic, message, err := portalSubscriber.ReceiveMessage()
				if err != nil {
					level.Error(logger).Log("msg", "error receiving message", "err", err)
					continue
				}

				portalCruncherMetrics.ReceivedMessageCount.Add(1)

				messageChan <- struct {
					topic   pubsub.Topic
					message []byte
				}{
					topic:   topic,
					message: message,
				}
			}
		}()
	}

	// Start redis insertion loop
	{
		for i := int64(0); i < redisGoroutineCount; i++ {
			go func() {
				for incoming := range messageChan {
					switch incoming.topic {
					case pubsub.TopicPortalCruncherSessionData:
						var sessionPortalData transport.SessionPortalData
						if err := sessionPortalData.UnmarshalBinary(incoming.message); err != nil {
							level.Error(logger).Log("msg", "error unmarshaling session data message", "err", err)
							continue
						}

						tx := redisClientPortal.TxPipeline()

						// set total session counts with expiration on the entire key set for safety
						switch sessionPortalData.Meta.OnNetworkNext {
						case true:
							// Remove the session from the direct set if it exists
							tx.ZRem("total-direct", sessionPortalData.Meta.ID)
							tx.ZRem(fmt.Sprintf("total-direct-buyer-%016x", sessionPortalData.Meta.BuyerID), sessionPortalData.Meta.ID)

							tx.ZAdd("total-next", &redis.Z{Score: sessionPortalData.Meta.DeltaRTT, Member: fmt.Sprintf("%016x", sessionPortalData.Meta.ID)})
							tx.Expire("total-next", redisPortalHostExp)
							tx.ZAdd(fmt.Sprintf("total-next-buyer-%016x", sessionPortalData.Meta.BuyerID), &redis.Z{Score: sessionPortalData.Meta.DeltaRTT, Member: fmt.Sprintf("%016x", sessionPortalData.Meta.ID)})
							tx.Expire(fmt.Sprintf("total-next-buyer-%016x", sessionPortalData.Meta.BuyerID), redisPortalHostExp)
						case false:
							// Remove the session from the next set if it exists
							tx.ZRem("total-next", sessionPortalData.Meta.ID)
							tx.ZRem(fmt.Sprintf("total-next-buyer-%016x", sessionPortalData.Meta.BuyerID), sessionPortalData.Meta.ID)

							tx.ZAdd("total-direct", &redis.Z{Score: -sessionPortalData.Meta.DirectRTT, Member: fmt.Sprintf("%016x", sessionPortalData.Meta.ID)})
							tx.Expire("total-direct", redisPortalHostExp)
							tx.ZAdd(fmt.Sprintf("total-direct-buyer-%016x", sessionPortalData.Meta.BuyerID), &redis.Z{Score: -sessionPortalData.Meta.DirectRTT, Member: fmt.Sprintf("%016x", sessionPortalData.Meta.ID)})
							tx.Expire(fmt.Sprintf("total-direct-buyer-%016x", sessionPortalData.Meta.BuyerID), redisPortalHostExp)
						}

						// set session and slice information with expiration on the entire key set for safety
						tx.Set(fmt.Sprintf("session-%016x-meta", sessionPortalData.Meta.ID), sessionPortalData.Meta, redisPortalHostExp)
						tx.SAdd(fmt.Sprintf("session-%016x-slices", sessionPortalData.Meta.ID), sessionPortalData.Slice)
						tx.Expire(fmt.Sprintf("session-%016x-slices", sessionPortalData.Meta.ID), redisPortalHostExp)

						// set the user session reverse lookup sets with expiration on the entire key set for safety
						tx.SAdd(fmt.Sprintf("user-%016x-sessions", sessionPortalData.Meta.UserHash), fmt.Sprintf("%016x", sessionPortalData.Meta.ID))
						tx.Expire(fmt.Sprintf("user-%016x-sessions", sessionPortalData.Meta.UserHash), redisPortalHostExp)

						// set the map point key and buyer sessions with expiration on the entire key set for safety
						tx.Set(fmt.Sprintf("session-%016x-point", sessionPortalData.Meta.ID), sessionPortalData.Point, redisPortalHostExp)
						tx.SAdd(fmt.Sprintf("map-points-%016x-buyer", sessionPortalData.Meta.BuyerID), fmt.Sprintf("%016x", sessionPortalData.Meta.ID))
						tx.Expire(fmt.Sprintf("map-points-%016x-buyer", sessionPortalData.Meta.BuyerID), redisPortalHostExp)

						if _, err := tx.Exec(); err != nil {
							level.Error(logger).Log("msg", "error sending session data to redis", "err", err)
							continue
						}

					case pubsub.TopicPortalCruncherSessionCounts:
						var countData transport.SessionCountData
						if err := countData.UnmarshalBinary(incoming.message); err != nil {
							level.Error(logger).Log("msg", "error unmarshaling session count message", "err", err)
							continue
						}

						tx := redisClientPortal.TxPipeline()

						// Regular set for expiry
						tx.Set(fmt.Sprintf("session-count-total-direct-instance-%016x", countData.InstanceID), countData.TotalNumDirectSessions, redisPortalHostExp)

						// HSet for quick summing in the portal
						tx.HSet("session-count-total-direct", fmt.Sprintf("session-count-total-direct-instance-%016x", countData.InstanceID), countData.TotalNumDirectSessions)

						// Regular set for expiry
						tx.Set(fmt.Sprintf("session-count-total-next-instance-%016x", countData.InstanceID), countData.TotalNumNextSessions, redisPortalHostExp)

						// HSet for quick summing in the portal
						tx.HSet("session-count-total-next", fmt.Sprintf("session-count-total-next-instance-%016x", countData.InstanceID), countData.TotalNumNextSessions)

						for buyerID, count := range countData.NumDirectSessionsPerBuyer {
							// Regular set for expiry
							tx.Set(fmt.Sprintf("session-count-direct-buyer-%016x-instance-%016x", buyerID, countData.InstanceID), count, redisPortalHostExp)

							// HSet for quick summing in the portal
							tx.HSet(fmt.Sprintf("session-count-direct-buyer-%016x", buyerID), fmt.Sprintf("session-count-direct-buyer-%016x-instance-%016x", buyerID, countData.InstanceID), count)
							tx.Expire(fmt.Sprintf("session-count-direct-buyer-%016x", buyerID), redisPortalHostExp)
						}

						for buyerID, count := range countData.NumNextSessionsPerBuyer {
							// Regular set for expiry
							tx.Set(fmt.Sprintf("session-count-next-buyer-%016x-instance-%016x", buyerID, countData.InstanceID), count, redisPortalHostExp)

							// HSet for quick summing in the portal
							tx.HSet(fmt.Sprintf("session-count-next-buyer-%016x", buyerID), fmt.Sprintf("session-count-next-buyer-%016x-instance-%016x", buyerID, countData.InstanceID), count)
							tx.Expire(fmt.Sprintf("session-count-next-buyer-%016x", buyerID), redisPortalHostExp)
						}

						if _, err := tx.Exec(); err != nil {
							level.Error(logger).Log("msg", "error sending session count data to redis", "err", err)
							continue
						}
					}
				}
			}()
		}
	}

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
