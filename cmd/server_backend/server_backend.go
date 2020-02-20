/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/base64"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"google.golang.org/api/option"

	"github.com/go-redis/redis/v7"
	"github.com/oschwald/geoip2-golang"

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
		LocalBuyer: &routing.Buyer{PublicKey: customerPublicKey},
	}

	// Create a no-op metrics handler in case metrics aren't set up
	var metricsHandler metrics.Handler
	metricsHandler = &metrics.NoOpHandler{}

	// If GCP_CREDENTIALS are set then override the local in memory
	// and connect to Firestore
	if gcpcreds, ok := os.LookupEnv("GCP_CREDENTIALS"); ok {
		var gcpcredsjson []byte

		_, err := os.Stat(gcpcreds)
		switch err := err.(type) {
		case *os.PathError:
			gcpcredsjson = []byte(gcpcreds)
			level.Info(logger).Log("envvar", "GCP_CREDENTIALS", "value", "<JSON>")
		case nil:
			gcpcredsjson, err = ioutil.ReadFile(gcpcreds)
			if err != nil {
				level.Error(logger).Log("envvar", "GCP_CREDENTIALS", "value", gcpcreds, "err", err)
				os.Exit(1)
			}
			level.Info(logger).Log("envvar", "GCP_CREDENTIALS", "value", gcpcreds)
		default:
			//log.Fatalf("unable to load GCP_CREDENTIALS: %v\n", err)
		}

		// Create a Firestore client
		client, err := firestore.NewClient(context.Background(), firestore.DetectProjectID, option.WithCredentialsJSON(gcpcredsjson))
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// Create a Firestore Storer
		fs := storage.Firestore{
			Client: client,
			Logger: logger,
		}

		// Start a goroutine to sync from Firestore
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			fs.SyncLoop(ctx, ticker.C)
		}()

		// Set the Firestore Storer to give to handlers
		db = &fs

		// Get all metric env vars to set up metrics
		metricEnvVars := []string{
			"GOOGLE_CLOUD_METRICS_CLUSTER_LOCATION",
			"GOOGLE_CLOUD_METRICS_CLUSTER_LOCATION",
			"GOOGLE_CLOUD_METRICS_POD_NAME",
			"GOOGLE_CLOUD_METRICS_CONTAINER_NAME",
			"GOOGLE_CLOUD_METRICS_NAMESPACE_NAME",
			"GOOGLE_CLOUD_METRICS_PROJECT",
		}
		metricEnvVarValues := make([]string, len(metricEnvVars))
		var ok bool
		for i := 0; i < len(metricEnvVarValues); i++ {
			metricEnvVarValues[i], ok = os.LookupEnv(metricEnvVars[i])
			if !ok {
				level.Warn(logger).Log("msg", "metric env var not set, metrics will not be tracked", "envvar", metricEnvVars[i])
				break
			}
		}

		if ok {
			// Create the metrics handler
			metricsHandler = &metrics.StackDriverHandler{
				ClusterLocation: metricEnvVarValues[0],
				ClusterName:     metricEnvVarValues[1],
				PodName:         metricEnvVarValues[2],
				ContainerName:   metricEnvVarValues[3],
				NamespaceName:   metricEnvVarValues[4],
				ProjectID:       metricEnvVarValues[5],
			}

			// Use a separate context for the metrics so that the metric submit routine can be stopped if need be
			metricsContext, _ := context.WithCancel(ctx)

			if err := metricsHandler.Open(metricsContext, gcpcredsjson); err == nil {
				go metricsHandler.MetricSubmitRoutine(metricsContext, logger, time.Minute, 200)
			} else {
				level.Error(logger).Log("msg", "Failed to create StackDriver metrics client", "err", err)
			}
		}
	}

	var routeMatrix routing.RouteMatrix
	{
		if uri, ok := os.LookupEnv("ROUTE_MATRIX_URI"); ok {
			go func() {
				for {
					var matrixReader io.Reader

					if f, err := os.Open(uri); err == nil {
						matrixReader = f
					}

					if r, err := http.Get(uri); err == nil {
						matrixReader = r.Body
					}

					if matrixReader != nil {
						_, err := routeMatrix.ReadFrom(matrixReader)
						if err != nil {
							level.Error(logger).Log("matrix", "route", "op", "read", "envvar", "ROUTE_MATRIX_URI", "value", uri, "err", err)
						}

						level.Info(logger).Log("matrix", "route", "entries", len(routeMatrix.Entries))
					}

					time.Sleep(10 * time.Second)
				}
			}()
		}
	}

	{
		addr := net.UDPAddr{
			Port: 40000,
			IP:   net.ParseIP("0.0.0.0"),
		}

		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			level.Error(logger).Log("addr", conn.LocalAddr().String(), "err", err)
			os.Exit(1)
		}

		mux := transport.UDPServerMux{
			Conn:          conn,
			MaxPacketSize: transport.DefaultMaxPacketSize,

			ServerUpdateHandlerFunc:  transport.ServerUpdateHandlerFunc(logger, redisClient, db, metricsHandler),
			SessionUpdateHandlerFunc: transport.SessionUpdateHandlerFunc(logger, redisClient, db, &routeMatrix, ipLocator, &geoClient, metricsHandler, serverPrivateKey, routerPrivateKey),
		}

		go func() {
			level.Info(logger).Log("addr", conn.LocalAddr().String())
			if err := mux.Start(ctx, runtime.NumCPU()); err != nil {
				level.Error(logger).Log("addr", conn.LocalAddr().String(), "err", err)
				os.Exit(1)
			}
		}()
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
