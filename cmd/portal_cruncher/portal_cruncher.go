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

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/envvar"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"

	"cloud.google.com/go/bigtable"
	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
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

	fmt.Printf("portal-cruncher: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	// Configure local logging
	logger := log.NewLogfmtLogger(os.Stdout)

	// Create a no-op metrics handler
	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	// Get GCP project ID
	gcpOK := envvar.Exists("GOOGLE_PROJECT_ID")
	gcpProjectID := envvar.Get("GOOGLE_PROJECT_ID", "")

	// StackDriver Logging
	{
		var enableSDLogging bool

		enableSDLogging, err := envvar.GetBool("ENABLE_STACKDRIVER_LOGGING", false)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		if enableSDLogging && gcpOK {
			loggingClient, err := gcplogging.NewClient(ctx, gcpProjectID)
			if err != nil {
				level.Error(logger).Log("msg", "failed to create GCP logging client", "err", err)
				return 1
			}

			logger = logging.NewStackdriverLogger(loggingClient, "portal-cruncher")
		}
	}

	{
		backendLogLevel := envvar.Get("BACKEND_LOG_LEVEL", "none")
		switch backendLogLevel {
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
	if !envvar.Exists("ENV") {
		level.Error(logger).Log("err", "ENV not set")
		return 1
	}
	env := envvar.Get("ENV", "")

	redisFlushCount, err := envvar.GetInt("PORTAL_CRUNCHER_REDIS_FLUSH_COUNT", 1000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	if gcpOK {
		// StackDriver Metrics
		{
			enableSDMetrics, err := envvar.GetBool("ENABLE_STACKDRIVER_METRICS", false)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
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
					return 1
				}

				metricsHandler = &sd

				sdWriteInterval, err := envvar.GetDuration("GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL", time.Minute)
				if err != nil {
					level.Error(logger).Log("err", err)
					return 1
				}
				go func() {
					metricsHandler.WriteLoop(ctx, logger, sdWriteInterval, 200)
				}()
			}
		}

		// StackDriver Profiler
		{
			enableSDProfiler, err := envvar.GetBool("ENABLE_STACKDRIVER_PROFILER", false)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
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
					return 1
				}
			}
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

	// Setup Bigtable

	btEmulatorOK := envvar.Exists("BIGTABLE_EMULATOR_HOST")
	if btEmulatorOK {
		// Emulator is used for local testing
		// Requires that emulator has been started in another terminal to work as intended
		gcpProjectID = "local"
		level.Info(logger).Log("msg", "Detected bigtable emulator")
	}

	// Get Bigtable client, table, and column family name
	var btClient *storage.BigTable
	var btTbl *bigtable.Table
	var btCfNames []string

	if gcpProjectID != "" || btEmulatorOK {
		// Get Bigtable instance ID
		btInstanceID := envvar.Get("GOOGLE_BIGTABLE_INSTANCE_ID", "")

		// Get the table name
		btTableName := envvar.Get("GOOGLE_BIGTABLE_TABLE_NAME", "")

		// Get the column family names and put them in a slice
		btCfName := envvar.Get("GOOGLE_BIGTABLE_CF_NAME", "")
		btCfNames = []string{btCfName}

		// Create a bigtable admin for setup
		btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID, logger)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		// Check if the table exists in the instance
		tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		// Verify if the table needed exists
		if !tableExists {
			// Create a table with the given name and column families
			if err = btAdmin.CreateTable(ctx, btTableName, btCfNames); err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			// Get the max number of days the data should be kept in Bigtable
			maxDays, err := envvar.GetInt("GOOGLE_BIGTABLE_MAX_AGE_DAYS", 90)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			// Set a garbage collection policy of 90 days
			maxAge := time.Hour * time.Duration(24*maxDays)
			if err = btAdmin.SetMaxAgePolicy(ctx, btTableName, btCfNames, maxAge); err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}
		}

		// Close the admin client
		if err = btAdmin.Close(); err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		// Create a standard client for writing to the table
		btClient, err = storage.NewBigTable(ctx, gcpProjectID, btInstanceID, logger)
		if err != nil {
			level.Error(logger).Log("err", err)
			return 1
		}

		btTbl = btClient.GetTable(btTableName)

	} else {
		level.Error(logger).Log("msg", "Could not find $BIGTABLE_EMULATOR_HOST")
		return 1
	}

	// Start portal cruncher subscriber
	redisCruncherPort := envvar.Get("CRUNCHER_PORT_REDIS", "5555")
	redisPortalSubscriber, err := getPortalSubscriber(logger, redisCruncherPort)
	if err != nil {
		return 1
	}

	btCruncherPort := envvar.Get("CRUNCHER_PORT_BIGTABLE", "5556")
	btPortalSubscriber, err := getPortalSubscriber(logger, btCruncherPort)
	if err != nil {
		return 1
	}
	

	receiveGoroutineCount, err := envvar.GetInt("CRUNCHER_RECEIVE_GOROUTINE_COUNT", 1)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	redisGoroutineCount, err := envvar.GetInt("CRUNCHER_REDIS_GOROUTINE_COUNT", 1)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	btGoroutineCount, err := envvar.GetInt("CRUNCHER_BIGTABLE_GOROUTINE_COUNT", 1)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	messageChanSize, err := envvar.GetInt("CRUNCHER_MESSAGE_CHANNEL_SIZE", 10000000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	redisMessageChan := make(chan []byte, messageChanSize)
	btMessageChan := make(chan []byte, messageChanSize)

	// Start receive loops
	for i := 0; i < receiveGoroutineCount; i += 2 {
		go func() {
			for {
				if err := ReceivePortalMessage(redisPortalSubscriber, portalCruncherMetrics, redisMessageChan); err != nil {
					switch err.(type) {
					case *ErrReceiveMessage:
						level.Error(logger).Log("err", err)
						os.Exit(1) // todo: don't os.Exit() here, but somehow quit
					case *ErrChannelFull:
						level.Error(logger).Log("err", err)
					}
				}
			}
		}()
		go func() {
			for {
				if err := ReceivePortalMessage(btPortalSubscriber, portalCruncherMetrics, btMessageChan); err != nil {
					switch err.(type) {
					case *ErrReceiveMessage:
						level.Error(logger).Log("err", err)
						os.Exit(1) // todo: don't os.Exit() here, but somehow quit
					case *ErrChannelFull:
						level.Error(logger).Log("err", err)
					}
				}
			}
		}()
	}

	redisHostTopSessions := envvar.Get("REDIS_HOST_TOP_SESSIONS", "127.0.0.1:6379")
	redisHostSessionMap := envvar.Get("REDIS_HOST_SESSION_MAP", "127.0.0.1:6379")
	redisHostSessionMeta := envvar.Get("REDIS_HOST_SESSION_META", "127.0.0.1:6379")
	redisHostSessionSlices := envvar.Get("REDIS_HOST_SESSION_SLICES", "127.0.0.1:6379")

	// Start redis insertion loop
	{
		for i := 0; i < redisGoroutineCount; i++ {
			go func() {
				clientTopSessions, clientSessionMap, clientSessionMeta, clientSessionSlices, err := createRedis(redisHostTopSessions, redisHostSessionMap, redisHostSessionMeta, redisHostSessionSlices)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1) // todo: don't exit here but find some way to return
				}

				if err := pingRedis(clientTopSessions, clientSessionMap, clientSessionMeta, clientSessionSlices); err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1) // todo: don't exit here but find some way to return
				}

				redisPortalDataBuffer := make([]transport.SessionPortalData, 0)

				flushTime := time.Now()
				pingTime := time.Now()

				for {
					// Pull the message out of the channel
					sessionData, err := PullMessage(redisMessageChan)
					if err != nil {
						level.Error(logger).Log("err", err)
						os.Exit(1) // todo: don't exit here but find some way to return
					}

					// Handle Redis functionality
					if err := RedisHandler(clientTopSessions, clientSessionMap, clientSessionMeta, clientSessionSlices, sessionData, redisPortalDataBuffer, flushTime, pingTime, redisFlushCount); err != nil {
						level.Error(logger).Log("err", err)
						os.Exit(1) // todo: don't exit here but find some way to return
					}
				}
			}()
		}
	}

	// Start Bigtable insertion loop
	{
		for i := 0; i < btGoroutineCount; i++ {
			go func() {
				
				btPortalDataBuffer := make([]transport.SessionPortalData, 0)

				for {
					// Pull the message out of the channel
					sessionData, err := PullMessage(btMessageChan)
					if err != nil {
						level.Error(logger).Log("err", err)
						os.Exit(1) // todo: don't exit here but find some way to return
					}

					// Handle Bigtable functionality
					if err := BTHandler(ctx, btClient, btTbl, btCfNames, sessionData, btPortalDataBuffer); err != nil {
						level.Error(logger).Log("err", err)
						os.Exit(1) // todo: don't exit here but find some way to return
					}
				}
			}()
		}
	}


	// Start HTTP server
	{
		go func() {
			router := mux.NewRouter()
			router.HandleFunc("/health", transport.HealthHandlerFunc())
			router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))

			port, ok := os.LookupEnv("HTTP_PORT")
			if !ok {
				level.Error(logger).Log("err", "env var HTTP_PORT must be set")
				os.Exit(1) // todo: don't os.Exit() here, but somehow quit
			}

			err := http.ListenAndServe(":"+port, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1) // todo: don't os.Exit() here, but somehow quit
			}
		}()
	}

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	// Close the Bigtable client
	btClient.Close()

	return 0
}

type ErrReceiveMessage struct {
	err error
}

func (e *ErrReceiveMessage) Error() string {
	return fmt.Sprintf("error receiving message: %v", e.err)
}

type ErrChannelFull struct{}

func (e *ErrChannelFull) Error() string {
	return "message channel full, dropping message"
}

type ErrUnmarshalMessage struct {
	err error
}

func (e *ErrUnmarshalMessage) Error() string {
	return fmt.Sprintf("could not unmarshal message: %v", e.err)
}

func ReceivePortalMessage(portalSubscriber pubsub.Subscriber, metrics *metrics.PortalCruncherMetrics, messageChan chan []byte) error {
	_, message, err := portalSubscriber.ReceiveMessage()
	if err != nil {
		return &ErrReceiveMessage{err: err}
	}

	metrics.ReceivedMessageCount.Add(1)

	select {
	case messageChan <- message:
	default:
		return &ErrChannelFull{}
	}

	return nil
}

func BTHandler(
	ctx context.Context,
	btClient *storage.BigTable,
	btTbl *bigtable.Table,
	btCfNames []string,
	sessionData transport.SessionPortalData,
	portalDataBuffer []transport.SessionPortalData) error {

	portalDataBuffer = append(portalDataBuffer, sessionData)

	InsertToBT(ctx, btClient, btTbl, btCfNames, portalDataBuffer)
	return nil
}

func InsertToBT(
	ctx context.Context,
	btClient *storage.BigTable,
	btTbl *bigtable.Table,
	btCfNames []string,
	portalDataBuffer []transport.SessionPortalData) error {

	for j := range portalDataBuffer {
		meta := &portalDataBuffer[j].Meta
		slice := &portalDataBuffer[j].Slice

		// Use customer (buyer) ID and user hash as our row key prefixes
		// Then have session ID as the identifier in the row key to indicate what to group by
		sessionRowKey := fmt.Sprintf("%d#%d", meta.BuyerID, meta.ID)
		userRowKey := fmt.Sprintf("%d#%d", meta.UserHash, meta.ID)

		rowKeys := []string{sessionRowKey, userRowKey}

		// Have 2 columns under 1 column family
		// 1) Meta
		// 2) Slice

		// Create byte slices of the session data
		metaBinary, err := meta.MarshalBinary()
		if err != nil {
			return err
		}
		sliceBinary, err := slice.MarshalBinary()
		if err != nil {
			return err
		}

		// Create a map of column name to session data
		sessionDataMap := make(map[string][]byte)
		sessionDataMap["meta"] = metaBinary
		sessionDataMap["slice"] = sliceBinary

		// Create a map of column name to column family
		// Always map meta and slice to the first column family
		cfMap := make(map[string]string)
		cfMap["meta"] = btCfNames[0]
		cfMap["slice"] = btCfNames[0]

		if err := btClient.InsertRowInTable(ctx, btTbl, rowKeys, sessionDataMap, cfMap); err != nil {
			return err
		}
	}

	portalDataBuffer = portalDataBuffer[:0]

	return nil
}

func RedisHandler(
	clientTopSessions storage.RedisClient,
	clientSessionMap storage.RedisClient,
	clientSessionMeta storage.RedisClient,
	clientSessionSlices storage.RedisClient,
	sessionData transport.SessionPortalData,
	portalDataBuffer []transport.SessionPortalData,
	flushTime time.Time,
	pingTime time.Time,
	redisFlushCount int) error {

	portalDataBuffer = append(portalDataBuffer, sessionData)

	if time.Since(flushTime) < time.Second && len(portalDataBuffer) < redisFlushCount {
		return nil
	}

	// Periodically ping the redis instances and restart if we don't get a pong
	if time.Since(pingTime) >= time.Second*10 {
		if err := pingRedis(clientTopSessions, clientSessionMap, clientSessionMeta, clientSessionSlices); err != nil {
			return err
		}

		pingTime = time.Now()
	}

	flushTime = time.Now()
	minutes := flushTime.Unix() / 60

	InsertToRedis(clientTopSessions, clientSessionMap, clientSessionMeta, clientSessionSlices, portalDataBuffer, minutes)
	return nil
}

func PullMessage(messageChan chan []byte) (transport.SessionPortalData, error) {
	message := <-messageChan

	var sessionPortalData transport.SessionPortalData
	if err := sessionPortalData.UnmarshalBinary(message); err != nil {
		return transport.SessionPortalData{}, &ErrUnmarshalMessage{err: err}
	}

	return sessionPortalData, nil
}

func InsertToRedis(
	clientTopSessions storage.RedisClient,
	clientSessionMap storage.RedisClient,
	clientSessionMeta storage.RedisClient,
	clientSessionSlices storage.RedisClient,
	portalDataBuffer []transport.SessionPortalData,
	minutes int64) {

	// Remove the old global top sessions minute bucket from 2 minutes ago if it didn't expire
	clientTopSessions.Command("DEL", "s-%d", minutes-2)

	// Update the current global top sessions minute bucket
	var format string
	args := make([]interface{}, 0)

	format += "s-%d"
	args = append(args, minutes)

	for j := range portalDataBuffer {
		meta := portalDataBuffer[j].Meta
		largeCustomer := portalDataBuffer[j].LargeCustomer
		everOnNext := portalDataBuffer[j].EverOnNext

		// For large customers, only insert the session if they have ever taken network next
		if largeCustomer && !meta.OnNetworkNext && !everOnNext {
			continue // Early out if we shouldn't add this session
		}

		sessionID := fmt.Sprintf("%016x", meta.ID)
		score := meta.DeltaRTT
		if score < 0 {
			score = 0
		}
		if !meta.OnNetworkNext {
			score = -meta.DirectRTT
		}

		format += " %.2f %s"
		args = append(args, score, sessionID)
	}

	clientTopSessions.Command("ZADD", format, args...)
	clientTopSessions.Command("EXPIRE", "s-%d %d", minutes, 30)

	for j := range portalDataBuffer {
		meta := &portalDataBuffer[j].Meta
		largeCustomer := portalDataBuffer[j].LargeCustomer
		everOnNext := portalDataBuffer[j].EverOnNext

		// For large customers, only insert the session if they have ever taken network next
		if largeCustomer && !meta.OnNetworkNext && !everOnNext {
			continue // Early out if we shouldn't add this session
		}

		slice := &portalDataBuffer[j].Slice
		point := &portalDataBuffer[j].Point
		sessionID := fmt.Sprintf("%016x", meta.ID)
		customerID := fmt.Sprintf("%016x", meta.BuyerID)
		next := meta.OnNetworkNext
		score := meta.DeltaRTT
		if score < 0 {
			score = 0
		}
		if !next {
			score = -100000 + meta.DirectRTT
		}

		// Remove the old per-buyer top sessions minute bucket from 2 minutes ago if it didnt expire
		// and update the current per-buyer top sessions list
		clientTopSessions.Command("DEL", "sc-%s-%d", customerID, minutes-2)
		clientTopSessions.Command("ZADD", "sc-%s-%d %.2f %s", customerID, minutes, score, sessionID)
		clientTopSessions.Command("EXPIRE", "sc-%s-%d %d", customerID, minutes, 30)

		// Remove the old map points minute buckets from 2 minutes ago if it didn't expire
		clientSessionMap.Command("HDEL", "d-%s-%d %s", customerID, minutes-2, sessionID)
		clientSessionMap.Command("HDEL", "n-%s-%d %s", customerID, minutes-2, sessionID)

		// Update the map points for this minute bucket
		// Make sure to remove the session ID from the opposite bucket in case the session
		// has switched from direct -> next or next -> direct
		pointString := point.RedisString()
		if next {
			clientSessionMap.Command("HSET", "n-%s-%d %s %s", customerID, minutes, sessionID, pointString)
			clientSessionMap.Command("HDEL", "d-%s-%d %s", customerID, minutes-1, sessionID)
			clientSessionMap.Command("HDEL", "d-%s-%d %s", customerID, minutes, sessionID)
		} else {
			clientSessionMap.Command("HSET", "d-%s-%d %s %s", customerID, minutes, sessionID, pointString)
			clientSessionMap.Command("HDEL", "n-%s-%d %s", customerID, minutes-1, sessionID)
			clientSessionMap.Command("HDEL", "n-%s-%d %s", customerID, minutes, sessionID)
		}

		// Expire map points
		clientSessionMap.Command("EXPIRE", "n-%s-%d %d", customerID, minutes, 30)
		clientSessionMap.Command("EXPIRE", "d-%s-%d %d", customerID, minutes, 30)

		// Update session meta
		clientSessionMeta.Command("SET", "sm-%s \"%s\" EX %d", sessionID, meta.RedisString(), 120)

		// Update session slices
		clientSessionSlices.Command("RPUSH", "ss-%s %s", sessionID, slice.RedisString())
		clientSessionSlices.Command("EXPIRE", "ss-%s %d", sessionID, 120)
	}

	portalDataBuffer = portalDataBuffer[:0]
}

func createRedis(redisHostTopSessions string, redisHostSessionMap string, redisHostSessionMeta string, redisHostSessionSlices string) (clientTopSessions *storage.RawRedisClient, clientSessionMap *storage.RawRedisClient, clientSessionMeta *storage.RawRedisClient, clientSessionSlices *storage.RawRedisClient, err error) {
	clientTopSessions, err = storage.NewRawRedisClient(redisHostTopSessions)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create redis client for %s: %v", redisHostTopSessions, err)
	}

	clientSessionMap, err = storage.NewRawRedisClient(redisHostSessionMap)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create redis client for %s: %v", redisHostSessionMap, err)
	}

	clientSessionMeta, err = storage.NewRawRedisClient(redisHostSessionMeta)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create redis client for %s: %v", redisHostSessionMeta, err)
	}

	clientSessionSlices, err = storage.NewRawRedisClient(redisHostSessionSlices)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create redis client for %s: %v", redisHostSessionSlices, err)
	}

	return clientTopSessions, clientSessionMap, clientSessionMeta, clientSessionSlices, nil
}

func pingRedis(clientTopSessions storage.RedisClient, clientSessionMap storage.RedisClient, clientSessionMeta storage.RedisClient, clientSessionSlices storage.RedisClient) error {
	if err := clientTopSessions.Ping(); err != nil {
		return err
	}

	if err := clientSessionMap.Ping(); err != nil {
		return err
	}

	if err := clientSessionMeta.Ping(); err != nil {
		return err
	}

	if err := clientSessionSlices.Ping(); err != nil {
		return err
	}

	return nil
}

func getPortalSubscriber(logger log.Logger, cruncherPort string) (pubsub.Subscriber, error) {

	receiveBufferSize, err := envvar.GetInt("CRUNCHER_RECEIVE_BUFFER_SIZE", 1000000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return nil, err
	}

	portalCruncherSubscriber, err := pubsub.NewPortalCruncherSubscriber(cruncherPort, int(receiveBufferSize))
	if err != nil {
		level.Error(logger).Log("msg", "could not create portal cruncher subscriber", "err", err)
		return nil, err
	}

	if err := portalCruncherSubscriber.Subscribe(pubsub.TopicPortalCruncherSessionData); err != nil {
		level.Error(logger).Log("msg", "could not subscribe to portal cruncher session data topic", "err", err)
		return nil, err
	}

	if err := portalCruncherSubscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts); err != nil {
		level.Error(logger).Log("msg", "could not subscribe to portal cruncher session counts topic", "err", err)
		return nil, err
	}

	return portalCruncherSubscriber, nil
}