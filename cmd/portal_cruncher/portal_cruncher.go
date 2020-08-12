/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"runtime"

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

func createRedisClient(hostname string) (net.Conn, error) {
	client, err := net.Dial("tcp", hostname)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func createAndValidateRedisClient(hostname string) (net.Conn, error) {
	client, err := createRedisClient(hostname)
	if err != nil {
		return nil, fmt.Errorf("could not dial: %v", err)
	}

	// Test if the redis client can be pinged
	fmt.Fprint(client, "PING\r\n")
	fmt.Fprint(client, "FLUSHDB\r\n")

	redisReplyReader := bufio.NewReader(client)
	reply, err := redisReplyReader.ReadString('\n')
	if err != nil || reply != "+PONG\r\n" {
		client.Close()
		return nil, fmt.Errorf("could not ping: %v", err)
	}

	return client, nil
}

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

	redisPortalHostExp, err := time.ParseDuration(os.Getenv("REDIS_HOST_PORTAL_EXPIRATION"))
	if err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_PORTAL_EXPIRATION", "msg", "could not parse", "err", err)
		os.Exit(1)
	}

	redisFlushCount := 1000
	redisFlushCountString, ok := os.LookupEnv("PORTAL_CRUNCHER_REDIS_FLUSH_COUNT")
	if ok {
		if redisFlushCount, err = strconv.Atoi(redisFlushCountString); err != nil {
			level.Error(logger).Log("envvar", "PORTAL_CRUNCHER_REDIS_FLUSH_COUNT", "msg", "could not parse", "err", err)
			os.Exit(1)
		}
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

				if topic != pubsub.TopicPortalCruncherSessionData {
					continue
				}

				if int64(len(messageChan)) < messageChanSize { // Drop messages if redis insertion is backed up
					messageChan <- struct {
						topic   pubsub.Topic
						message []byte
					}{
						topic:   topic,
						message: message,
					}
				}
			}
		}()
	}

	// Start redis insertion loop
	{
		for i := int64(0); i < redisGoroutineCount; i++ {
			go func() {
				// Each goroutine should use its own TCP socket
				clientTopSessions, err := createAndValidateRedisClient(os.Getenv("REDIS_HOST_TOP_SESSIONS"))
				if err != nil {
					level.Error(logger).Log("envvar", "REDIS_HOST_TOP_SESSIONS", "err", err)
					os.Exit(1)
				}

				clientSessionMap, err := createAndValidateRedisClient(os.Getenv("REDIS_HOST_SESSION_MAP"))
				if err != nil {
					level.Error(logger).Log("envvar", "REDIS_HOST_SESSION_MAP", "err", err)
					os.Exit(1)
				}

				clientSessionMeta, err := createAndValidateRedisClient(os.Getenv("REDIS_HOST_SESSION_META"))
				if err != nil {
					level.Error(logger).Log("envvar", "REDIS_HOST_SESSION_META", "err", err)
					os.Exit(1)
				}

				clientSessionSlices, err := createAndValidateRedisClient(os.Getenv("REDIS_HOST_SESSION_SLICES"))
				if err != nil {
					level.Error(logger).Log("envvar", "REDIS_HOST_SESSION_SLICES", "err", err)
					os.Exit(1)
				}

				portalDataBuffer := make([]transport.SessionPortalData, redisFlushCount)

				now := time.Now()

				for incoming := range messageChan {
					var sessionPortalData transport.SessionPortalData
					if err := sessionPortalData.UnmarshalBinary(incoming.message); err != nil {
						level.Error(logger).Log("msg", "error unmarshaling session data message", "err", err)
						continue
					}
					portalDataBuffer = append(portalDataBuffer, sessionPortalData)

					if time.Since(now) < time.Second && len(portalDataBuffer) < redisFlushCount {
						continue
					}

					now = time.Now()
					secs := now.Unix()
					minutes := secs / 60

					// Remove the old global top sessions minute bucket from 2 minutes ago if it didn't expire
					fmt.Fprintf(clientTopSessions, "DEL s-%d\r\n", minutes-2)

					// Update the current global top sessions minute bucket
					fmt.Fprintf(clientTopSessions, "ZADD s-%d", minutes)
					for j := range portalDataBuffer {
						sessionID := fmt.Sprintf("%016x", portalDataBuffer[j].Meta.ID)
						score := portalDataBuffer[j].Meta.DeltaRTT
						fmt.Fprintf(clientTopSessions, " %.2f %s", score, sessionID)
					}
					fmt.Fprintf(clientTopSessions, "\r\n")
					fmt.Fprintf(clientTopSessions, "EXPIRE s-%d %d\r\n", minutes, int(redisPortalHostExp.Seconds()))

					for j := range portalDataBuffer {
						meta := &portalDataBuffer[j].Meta
						slice := &portalDataBuffer[j].Slice
						sessionID := fmt.Sprintf("%016x", meta.ID)
						customerID := fmt.Sprintf("%016x", meta.BuyerID)
						score := meta.DeltaRTT
						next := meta.OnNetworkNext
						location := fmt.Sprintf("%.2f|%.2f", meta.Location.Latitude, meta.Location.Longitude)

						// Remove the old per-buyer top sessions minute bucket from 2 minutes ago if it didnt expire
						// and update the current per-buyer top sessions list
						fmt.Fprintf(clientTopSessions, "DEL sc-%s-%d\r\n", customerID, minutes-2)
						fmt.Fprintf(clientTopSessions, "ZADD sc-%s-%d %.2f %s\r\n", customerID, minutes, score, sessionID)
						fmt.Fprintf(clientTopSessions, "EXPIRE sc-%s-%d %d\r\n", customerID, minutes, int(redisPortalHostExp.Seconds()))

						// Remove the old map points minute buckets from 2 minutes ago if it didn't expire
						fmt.Fprintf(clientTopSessions, "HDEL d-%s-%d %s\r\n", customerID, minutes-2, sessionID)
						fmt.Fprintf(clientTopSessions, "HDEL n-%s-%d %s\r\n", customerID, minutes-2, sessionID)

						// Update the map points for this minute bucket
						// Make sure to remove the session ID from the opposite bucket in case the session
						// has switched from direct -> next or next -> direct
						if next {
							fmt.Fprintf(clientSessionMap, "HSET n-%s-%d %s %s\r\n", customerID, minutes, sessionID, location)
							fmt.Fprintf(clientSessionMap, "HDEL d-%s-%d %s\r\n", customerID, minutes, sessionID)
						} else {
							fmt.Fprintf(clientSessionMap, "HSET d-%s-%d %s %s\r\n", customerID, minutes, sessionID, location)
							fmt.Fprintf(clientSessionMap, "HDEL n-%s-%d %s\r\n", customerID, minutes, sessionID)
						}

						// Expire map points
						fmt.Fprintf(clientSessionMap, "EXPIRE n-%s-%d %d\r\n", customerID, minutes, int(redisPortalHostExp.Seconds()))
						fmt.Fprintf(clientSessionMap, "EXPIRE d-%s-%d %d\r\n", customerID, minutes, int(redisPortalHostExp.Seconds()))

						// Update session meta
						fmt.Fprintf(clientSessionMeta, "SET sm-%s %v EX %d\r\n", sessionID, meta.RedisString(), int(redisPortalHostExp.Seconds()))

						// Update session slices
						fmt.Fprintf(clientSessionSlices, "RPUSH ss-%s %s\r\n", sessionID, slice.RedisString())
						fmt.Fprintf(clientSessionSlices, "EXPIRE ss-%s %d\r\n", sessionID, int(redisPortalHostExp.Seconds()))
					}

					portalDataBuffer = portalDataBuffer[:0]
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
