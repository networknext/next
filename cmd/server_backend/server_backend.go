/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/base64"
	_ "expvar"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"time"

	"github.com/networknext/backend/logging"

	gcplogging "cloud.google.com/go/logging"
	monitoring "cloud.google.com/go/monitoring/apiv3"

	gkmetrics "github.com/go-kit/kit/metrics"

	"github.com/go-kit/kit/metrics/expvar"

	"cloud.google.com/go/firestore"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis/v7"
	"github.com/oschwald/geoip2-golang"

	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
)

func main() {
	ctx := context.Background()

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)
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
	if projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		loggingClient, err := gcplogging.NewClient(ctx, projectID)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		logger = logging.NewStackdriverLogger(loggingClient, "server-backend")
	}

	// var serverPublicKey []byte
	var customerPublicKey []byte
	var serverPrivateKey []byte
	var routerPrivateKey []byte
	{
		if key := os.Getenv("SERVER_BACKEND_PUBLIC_KEY"); len(key) != 0 {
			// serverPublicKey, _ = base64.StdEncoding.DecodeString(key)
		} else {
			level.Error(logger).Log("err", "SERVER_BACKEND_PUBLIC_KEY not set")
			os.Exit(1)
		}

		if key := os.Getenv("SERVER_BACKEND_PRIVATE_KEY"); len(key) != 0 {
			serverPrivateKey, _ = base64.StdEncoding.DecodeString(key)
		} else {
			level.Error(logger).Log("err", "SERVER_BACKEND_PRIVATE_KEY not set")
			os.Exit(1)
		}

		if key := os.Getenv("RELAY_ROUTER_PRIVATE_KEY"); len(key) != 0 {
			routerPrivateKey, _ = base64.StdEncoding.DecodeString(key)
		} else {
			level.Error(logger).Log("err", "RELAY_ROUTER_PRIVATE_KEY not set")
			os.Exit(1)
		}

		if key := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(key) != 0 {
			customerPublicKey, _ = base64.StdEncoding.DecodeString(key)
			customerPublicKey = customerPublicKey[8:]
		}
	}

	// Attempt to connect to REDIS_HOST, falling back to local instance if not explicitly specified
	redisHost, ok := os.LookupEnv("REDIS_HOST")
	if !ok {
		redisHost = "localhost:6379"
		level.Warn(logger).Log("envvar", "REDIS_HOST", "value", redisHost)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: redisHost})
	if err := redisClient.Ping().Err(); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST", "value", redisHost, "err", err)
		os.Exit(1)
	}

	// Open the Maxmind DB and create a routing.MaxmindDB from it
	var ipLocator routing.IPLocator = routing.NullIsland
	if filename, ok := os.LookupEnv("MAXMIND_DB_URI"); ok {
		if mmreader, err := geoip2.Open(filename); err != nil {
			if err != nil {
				level.Error(logger).Log("envvar", "RELAY_MAXMIND_DB_URI", "value", filename, "err", err)
			}
			ipLocator = &routing.MaxmindDB{
				Reader: mmreader,
			}
			defer mmreader.Close()
		}
	}

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "RELAY_LOCATIONS",
	}

	// Create an in-memory db
	var db storage.Storer = &storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey:            customerPublicKey,
			RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
		},
	}

	// Create a no-op biller
	var biller billing.Biller = &billing.NoOpBiller{}

	// Create a no-op metrics handler
	var metricsHandler metrics.Handler = &metrics.NoOpHandler{}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpProjectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		firestoreClient, err := firestore.NewClient(ctx, gcpProjectID)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// Create a Firestore Storer
		fs := storage.Firestore{
			Client: firestoreClient,
			Logger: logger,
		}

		// Start a goroutine to sync from Firestore
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			fs.SyncLoop(ctx, ticker.C)
		}()

		// Set the Firestore Storer to give to handlers
		db = &fs

		if billingTopicID, ok := os.LookupEnv("GOOGLE_PUBSUB_TOPIC_BILLING"); ok {
			b, err := billing.NewBiller(ctx, logger, gcpProjectID, billingTopicID, &billing.Descriptor{
				ClientCount:         4,
				DelayThreshold:      time.Millisecond,
				CountThreshold:      100,
				ByteThreshold:       1e6,
				NumGoroutines:       (25 * runtime.GOMAXPROCS(0)) / 4,
				Timeout:             time.Minute,
				ResultChannelBuffer: 10000 * 60 * 10, // 10,000 messages per second for 10 minutes
			})
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			// Set the Biller to the Pub/Sub version
			biller = b
		}

		stackdriverClient, err := monitoring.NewMetricClient(ctx)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		sd := metrics.StackDriverHandler{
			Client:    stackdriverClient,
			ProjectID: gcpProjectID,
		}

		go func() {
			metricsHandler.MetricSubmitRoutine(ctx, logger, time.Minute, 200)
		}()

		metricsHandler = &sd
	}

	var routeMatrix routing.RouteMatrix
	{
		if uri, ok := os.LookupEnv("ROUTE_MATRIX_URI"); ok {
			go func() {
				for {
					var matrixReader io.Reader

					// Default to reading route matrix from file
					if f, err := os.Open(uri); err == nil {
						matrixReader = f
					}

					// Prefer to get it remotely if possible
					if r, err := http.Get(uri); err == nil {
						matrixReader = r.Body
					}

					// Attempt to read, and intentionally force to empty route matrix if any errors are encountered to avoid stale routes
					_, err := routeMatrix.ReadFrom(matrixReader)
					if err != nil {
						routeMatrix = routing.RouteMatrix{}
						level.Warn(logger).Log("matrix", "route", "op", "read", "envvar", "ROUTE_MATRIX_URI", "value", uri, "err", err, "msg", "forcing empty route matrix to avoid stale routes")
					}

					level.Info(logger).Log("matrix", "route", "entries", len(routeMatrix.Entries))

					time.Sleep(10 * time.Second)
				}
			}()
		}
	}

	var serverUpdateDuration gkmetrics.Histogram
	var serverUpdateCounter gkmetrics.Counter
	var sessionUpdateDuration gkmetrics.Histogram
	var sessionUpdateCounter gkmetrics.Counter
	{
		serverUpdateDuration = expvar.NewHistogram("server.update.duration", 50)
		serverUpdateCounter = expvar.NewCounter("server.update.counter")
		sessionUpdateDuration = expvar.NewHistogram("session.update.duration", 50)
		sessionUpdateCounter = expvar.NewCounter("session.update.counter")
	}

	{
		port, ok := os.LookupEnv("PORT")
		if !ok {
			level.Error(logger).Log("err", "env var PORT must be set")
			os.Exit(1)
		}
		iport, err := strconv.ParseInt(port, 10, 64)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		addr := net.UDPAddr{
			Port: int(iport),
		}

		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			level.Error(logger).Log("addr", conn.LocalAddr().String(), "err", err)
			os.Exit(1)
		}

		mux := transport.UDPServerMux{
			Conn:          conn,
			MaxPacketSize: transport.DefaultMaxPacketSize,

			ServerUpdateHandlerFunc:  transport.ServerUpdateHandlerFunc(logger, redisClient, db, serverUpdateDuration, serverUpdateCounter, metricsHandler),
			SessionUpdateHandlerFunc: transport.SessionUpdateHandlerFunc(logger, redisClient, db, sessionUpdateDuration, sessionUpdateCounter, &routeMatrix, ipLocator, &geoClient, metricsHandler, biller, serverPrivateKey, routerPrivateKey),
		}

		go func() {
			level.Info(logger).Log("protocol", "udp", "addr", conn.LocalAddr().String())
			if err := mux.Start(ctx, runtime.NumCPU()); err != nil {
				level.Error(logger).Log("protocol", "udp", "addr", conn.LocalAddr().String(), "err", err)
				os.Exit(1)
			}
		}()

		go func() {
			level.Info(logger).Log("protocol", "http", "addr", conn.LocalAddr().String())
			if err := http.ListenAndServe(conn.LocalAddr().String(), nil); err != nil {
				level.Error(logger).Log("protocol", "http", "addr", conn.LocalAddr().String(), "err", err)
				os.Exit(1)
			}
		}()
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
