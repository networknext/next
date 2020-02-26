/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"google.golang.org/api/option"

	"github.com/go-redis/redis/v7"
	"github.com/oschwald/geoip2-golang"

	"github.com/networknext/backend/crypto"
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

	var customerPublicKey []byte
	{
		if key := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(key) != 0 {
			customerPublicKey, _ = base64.StdEncoding.DecodeString(key)
			customerPublicKey = customerPublicKey[8:]
		}
	}

	var relayPublicKey []byte
	{
		if key := os.Getenv("RELAY_PUBLIC_KEY"); len(key) != 0 {
			relayPublicKey, _ = base64.StdEncoding.DecodeString(key)
		}
	}

	var routerPrivateKey []byte
	{
		if key := os.Getenv("RELAY_ROUTER_PRIVATE_KEY"); len(key) != 0 {
			routerPrivateKey, _ = base64.StdEncoding.DecodeString(key)
		} else {
			level.Error(logger).Log("err", "RELAY_ROUTER_PRIVATE_KEY not set")
			os.Exit(1)
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

	var ipLocator routing.IPLocator = routing.NullIsland
	if uri, ok := os.LookupEnv("MAXMIND_DB_URI"); ok {
		mmreader, err := geoip2.Open(uri)
		if err != nil {
			level.Error(logger).Log("envvar", "RELAY_MAXMIND_DB_URI", "value", uri, "err", err)
		}
		ipLocator = &routing.MaxmindDB{
			Reader: mmreader,
		}
		defer mmreader.Close()
	}

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "RELAY_LOCATIONS",
	}

	// Create an in-memory relay & datacenter store
	// that doesn't require talking to configstore
	var db storage.Storer = &storage.InMemory{
		LocalRelay: &routing.Relay{
			PublicKey: relayPublicKey,
			Datacenter: routing.Datacenter{
				ID:   crypto.HashID("local"),
				Name: "local",
			},
		},
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
		client, err := firestore.NewClient(ctx, firestore.DetectProjectID, option.WithCredentialsJSON(gcpcredsjson))
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
	}

	// If GCP_CREDENTIALS_METRICS are set then override the no-op metric handler and connect to StackDriver
	// This has its own credentials because the StackDriver metrics are in a separate workspace
	if stackdrivercreds, ok := os.LookupEnv("GCP_CREDENTIALS_METRICS"); ok {
		if stackDriverProjectID, ok := os.LookupEnv("GCP_METRICS_PROJECT"); ok {
			var stackdrivercredsjson []byte

			_, err := os.Stat(stackdrivercreds)
			switch err := err.(type) {
			case *os.PathError:
				stackdrivercredsjson = []byte(stackdrivercreds)
				level.Info(logger).Log("envvar", "GCP_CREDENTIALS_METRICS", "value", "<JSON>")
			case nil:
				stackdrivercredsjson, err = ioutil.ReadFile(stackdrivercreds)
				if err != nil {
					level.Error(logger).Log("envvar", "GCP_CREDENTIALS_METRICS", "value", stackdrivercreds, "err", err)
					os.Exit(1)
				}
				level.Info(logger).Log("envvar", "GCP_CREDENTIALS_METRICS", "value", stackdrivercreds)
			}

			// Create the metrics handler
			metricsHandler = &metrics.StackDriverHandler{
				ProjectID:       stackDriverProjectID,
				ClusterLocation: os.Getenv("GCP_METRICS_CLUSTER_LOCATION"),
				ClusterName:     os.Getenv("GCP_METRICS_CLUSTER_NAME"),
				PodName:         os.Getenv("GCP_METRICS_POD_NAME"),
				ContainerName:   os.Getenv("GCP_METRICS_CONTAINER_NAME"),
				NamespaceName:   os.Getenv("GCP_METRICS_NAMESPACE_NAME"),
			}

			if err := metricsHandler.Open(ctx, stackdrivercredsjson); err == nil {
				go metricsHandler.MetricSubmitRoutine(ctx, logger, time.Minute, 200)
			} else {
				level.Error(logger).Log("msg", "Failed to create StackDriver metrics client", "err", err)
			}
		}
	}

	statsdb := routing.NewStatsDatabase()
	var costmatrix routing.CostMatrix
	var routematrix routing.RouteMatrix

	go func() {
		for {
			if err := statsdb.GetCostMatrix(&costmatrix, redisClient); err != nil {
				level.Warn(logger).Log("matrix", "cost", "op", "generate", "err", err)
			}

			if err := costmatrix.Optimize(&routematrix, 1); err != nil {
				level.Warn(logger).Log("matrix", "cost", "op", "optimize", "err", err)
			}

			level.Info(logger).Log("matrix", "route", "entries", len(routematrix.Entries))

			time.Sleep(10 * time.Second)
		}
	}()

	router := transport.NewRouter(logger, redisClient, &geoClient, ipLocator, db, statsdb, metricsHandler, &costmatrix, &routematrix, routerPrivateKey)

	go func() {
		port, ok := os.LookupEnv("PORT")
		if !ok {
			level.Error(logger).Log("err", "env var PORT must be set")
			os.Exit(1)
		}

		level.Info(logger).Log("addr", ":"+port)

		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
