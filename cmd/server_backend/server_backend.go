/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
)

var (
	release string
)

func main() {
	ctx := context.Background()

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)
	if projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		loggingClient, err := gcplogging.NewClient(ctx, projectID)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		logger = logging.NewStackdriverLogger(loggingClient, "server-backend")
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

	redisPortalHosts := os.Getenv("REDIS_HOST_PORTAL")
	splitPortalHosts := strings.Split(redisPortalHosts, ",")
	redisClientPortal := storage.NewRedisClient(splitPortalHosts...)
	if err := redisClientPortal.Ping().Err(); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_PORTAL", "value", redisPortalHosts, "err", err)
		os.Exit(1)
	}

	redisPortalHostExpiration, err := time.ParseDuration(os.Getenv("REDIS_HOST_PORTAL_EXPIRATION"))
	if err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_PORTAL_EXPIRATION", "err", err)
		os.Exit(1)
	}

	redisHost := os.Getenv("REDIS_HOST_RELAYS")
	redisClientRelays := storage.NewRedisClient(redisHost)
	if err := redisClientRelays.Ping().Err(); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_RELAYS", "value", redisHost, "err", err)
		os.Exit(1)
	}

	redisHosts := strings.Split(os.Getenv("REDIS_HOST_CACHE"), ",")
	redisClientCache := storage.NewRedisClient(redisHosts...)
	if err := redisClientCache.Ping().Err(); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_CACHE", "value", redisHosts, "err", err)
		os.Exit(1)
	}

	// Open the Maxmind DB and create a routing.MaxmindDB from it
	var ipLocator routing.IPLocator = routing.NullIsland
	mmcitydburi := os.Getenv("MAXMIND_CITY_DB_URI")
	mmispdburi := os.Getenv("MAXMIND_ISP_DB_URI")
	if mmcitydburi != "" && mmispdburi != "" {
		cityreader, err := routing.NewMaxmindReader(http.DefaultClient, mmcitydburi)
		if err != nil {
			level.Error(logger).Log("envvar", "MAXMIND_CITY_DB_URI", "value", mmcitydburi, "err", err)
			os.Exit(1)
		}

		ispreader, err := routing.NewMaxmindReader(http.DefaultClient, mmispdburi)
		if err != nil {
			level.Error(logger).Log("envvar", "MAXMIND_ISP_DB_URI", "value", mmispdburi, "err", err)

		}

		ipLocator = &routing.MaxmindDB{
			CityReader: cityreader,
			ISPReader:  ispreader,
		}
		defer func() {
			cityreader.Close()
			ispreader.Close()
		}()
	}
	// Commented out to ensure it really does not load the IPStack version
	// if key, ok := os.LookupEnv("IPSTACK_ACCESS_KEY"); ok {
	// 	ipLocator = &routing.IPStack{
	// 		Client:    http.DefaultClient,
	// 		AccessKey: key,
	// 	}
	// }

	geoClient := routing.GeoClient{
		RedisClient: redisClientRelays,
		Namespace:   "RELAY_LOCATIONS",
	}

	// Create an in-memory db
	var db storage.Storer = &storage.InMemory{
		LocalMode: true,
	}

	if err := db.AddBuyer(ctx, routing.Buyer{
		ID:                   13672574147039585173,
		Name:                 "local",
		PublicKey:            customerPublicKey,
		RoutingRulesSettings: routing.LocalRoutingRulesSettings,
	}); err != nil {
		level.Error(logger).Log("msg", "could not add buyer to storage", "err", err)
		os.Exit(1)
	}
	if err := db.AddDatacenter(ctx, routing.Datacenter{
		ID:      crypto.HashID("local"),
		Name:    "local",
		Enabled: true,
	}); err != nil {
		level.Error(logger).Log("msg", "could not add datacenter to storage", "err", err)
		os.Exit(1)
	}

	// Create a no-op biller
	var biller billing.Biller = &billing.NoOpBiller{}

	// Create a local metrics handler
	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpProjectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		// Create a Firestore Storer
		fs, err := storage.NewFirestore(ctx, gcpProjectID, logger)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// Start a goroutine to sync from Firestore
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			fs.SyncLoop(ctx, ticker.C)
		}()

		// Set the Firestore Storer to give to handlers
		db = fs

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

		go func() {
			metricsHandler.WriteLoop(ctx, logger, time.Minute, 200)
		}()

		// Set up StackDriver profiler
		if err := profiler.Start(profiler.Config{
			Service:        "server_backend",
			ServiceVersion: env,
			ProjectID:      gcpProjectID,
			MutexProfiling: true,
		}); err != nil {
			level.Error(logger).Log("msg", "Failed to initialze StackDriver profiler", "err", err)
			os.Exit(1)
		}
	}

	// Create server update metrics
	serverInitMetrics, err := metrics.NewServerInitMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create server update metrics", "err", err)
	}

	// Create server update metrics
	serverUpdateMetrics, err := metrics.NewServerUpdateMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create server update metrics", "err", err)
	}

	// Create session update metrics
	sessionMetrics, err := metrics.NewSessionMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create session metrics", "err", err)
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

					time.Sleep(1 * time.Second)
				}
			}()
		}
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
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		mux := transport.UDPServerMux{
			Conn:          conn,
			MaxPacketSize: transport.DefaultMaxPacketSize,

			ServerInitHandlerFunc:    transport.ServerInitHandlerFunc(logger, redisClientCache, db, serverInitMetrics, serverPrivateKey),
			ServerUpdateHandlerFunc:  transport.ServerUpdateHandlerFunc(logger, redisClientCache, db, serverUpdateMetrics),
			SessionUpdateHandlerFunc: transport.SessionUpdateHandlerFunc(logger, redisClientCache, redisClientPortal, redisPortalHostExpiration, db, &routeMatrix, ipLocator, &geoClient, sessionMetrics, biller, serverPrivateKey, routerPrivateKey),
		}

		go func() {
			level.Info(logger).Log("protocol", "udp", "addr", conn.LocalAddr().String())
			if err := mux.Start(ctx); err != nil {
				level.Error(logger).Log("protocol", "udp", "addr", conn.LocalAddr().String(), "err", err)
				os.Exit(1)
			}
		}()

		go func() {
			http.HandleFunc("/healthz", transport.HealthzHandlerFunc())

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
