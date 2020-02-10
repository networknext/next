/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/go-redis/redis/v7"
	"github.com/oschwald/geoip2-golang"
	"google.golang.org/grpc"

	"github.com/networknext/backend/crypto"
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

	// Set the default IPLocator to resolve all lookups to 0/0 aka Null Island
	var ipLocator routing.IPLocator = routing.NullIsland

	// Create an in-memory relay & datacenter store
	// that doesn't require talking to configstore
	inMemory := storage.InMemory{
		LocalCustomerPublicKey: customerPublicKey,
		LocalRelayPublicKey:    relayPublicKey,
		LocalDatacenter:        true,
	}

	if filename, ok := os.LookupEnv("RELAYS_STUBBED_DATA_FILENAME"); ok {
		type relays struct {
			routing.Location
			DatacenterName string
			PublicKey      []byte
		}
		relaydata := make(map[string]relays)

		f, err := os.Open(filename)
		if err != nil {
			level.Error(logger).Log(err)
		}

		if err := json.NewDecoder(f).Decode(&relaydata); err != nil {
			level.Error(logger).Log(err)
		}

		inMemory.RelayDatacenterNames = make(map[uint32]string)
		inMemory.RelayPublicKeys = make(map[uint32][]byte)
		for ip, relay := range relaydata {
			inMemory.RelayDatacenterNames[uint32(crypto.HashID(ip))] = relay.DatacenterName
			inMemory.RelayPublicKeys[uint32(crypto.HashID(ip))] = relay.PublicKey
		}

		ipLocator = routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
			if relay, ok := relaydata[ip.String()]; ok {
				return routing.Location{
					Latitude:  relay.Latitude,
					Longitude: relay.Longitude,
				}, nil
			}

			return routing.Location{}, fmt.Errorf("relay address '%s' could not be found", ip.String())
		})

		level.Info(logger).Log("envvar", "RELAYS_STUBBED_DATA_FILENAME", "value", filename)
	}

	// Set the IPLocator to use Maxmind if set
	if uri, ok := os.LookupEnv("RELAY_MAXMIND_DB_URI"); ok {
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

	var relayProvider transport.RelayProvider = &inMemory
	var datacenterProvider transport.DatacenterProvider = &inMemory
	if host, ok := os.LookupEnv("CONFIGSTORE_HOST"); ok {
		grpcconn, err := grpc.Dial(host, grpc.WithInsecure())
		if err != nil {
			level.Error(logger).Log("envvar", "CONFIGSTORE_HOST", "value", host, "err", err)
		}
		configstore, err := storage.ConnectToConfigstore(ctx, grpcconn)
		if err != nil {
			level.Error(logger).Log("envvar", "CONFIGSTORE_HOST", "value", host, "err", err)
		}

		// If CONFIGSTORE_HOST exists and a successful connection was made
		// then replace the in-memory with the gRPC one
		relayProvider = configstore.Relays
		datacenterProvider = configstore.Datacenters
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

	router := transport.NewRouter(logger, redisClient, &geoClient, ipLocator, relayProvider, datacenterProvider, statsdb, &costmatrix, &routematrix, relayPublicKey, routerPrivateKey)

	go func() {
		level.Info(logger).Log("addr", ":30000")

		err := http.ListenAndServe(":30000", router)
		if err != nil {
			level.Error(logger).Log("err", err)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
