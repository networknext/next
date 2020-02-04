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
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/oschwald/geoip2-golang"
	"google.golang.org/grpc"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := context.Background()

	var relayPublicKey []byte
	var routerPrivateKey []byte
	{
		if key := os.Getenv("RELAY_PUBLIC_KEY"); len(key) != 0 {
			relayPublicKey, _ = base64.StdEncoding.DecodeString(key)
		} else {
			log.Fatal("env var 'RELAY_PUBLIC_KEY' is not set")
		}

		if key := os.Getenv("RELAY_ROUTER_PRIVATE_KEY"); len(key) != 0 {
			routerPrivateKey, _ = base64.StdEncoding.DecodeString(key)
		} else {
			log.Fatal("env var 'RELAY_ROUTER_PRIVATE_KEY' is not set")
		}
	}

	// Attempt to connect to REDIS_HOST, falling back to local instance if not explicitly specified
	var redisHost string
	if redisHost = os.Getenv("REDIS_HOST"); len(redisHost) == 0 {
		redisHost = "localhost:6379"
		log.Printf("env var 'REDIS_HOST' is not set, falling back to default value of '%s'\n", redisHost)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: redisHost})
	if err := redisClient.Ping().Err(); err != nil {
		log.Fatalf("unable to connect to REDIS_HOST '%s'", redisHost)
	}

	// Set the default IPLocator to resolve all lookups to 0/0 aka Null Island
	var ipLocator routing.IPLocator = routing.NullIsland

	// Create an in-memory relay & datacenter store
	// that doesn't require talking to configstore
	inMemory := storage.InMemory{
		LocalDatacenter: true,
	}

	if filename, ok := os.LookupEnv("RELAYS_STUBBED_DATA_FILENAME"); ok {
		type relays struct {
			routing.Location
			DatacenterName string
		}
		relaydata := make(map[string]relays)

		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}

		if err := json.NewDecoder(f).Decode(&relaydata); err != nil {
			log.Fatal(err)
		}

		inMemory.RelayDatacenterNames = make(map[uint32]string)
		for ip, relay := range relaydata {
			inMemory.RelayDatacenterNames[uint32(crypto.HashID(ip))] = relay.DatacenterName
		}

		ipLocator = routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
			if relay, ok := relaydata[ip.String()]; ok {
				log.Printf("found stubbed lat long for relay address: '%s'\n", ip.String())
				return routing.Location{
					Latitude:  relay.Latitude,
					Longitude: relay.Longitude,
				}, nil
			}

			return routing.Location{}, fmt.Errorf("relay address '%s' could not be found", ip.String())
		})

		log.Printf("loaded %d relays from %s\n", len(relaydata), filename)
	}

	// Set the IPLocator to use Maxmind if set
	if uri, ok := os.LookupEnv("RELAY_MAXMIND_DB_URI"); ok {
		mmreader, err := geoip2.Open(uri)
		if err != nil {
			log.Fatalf("failed to open Maxmind GeoIP2 database: %v", err)
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
	if os.Getenv("CONFIGSTORE_HOST") != "" {
		grpcconn, err := grpc.Dial(os.Getenv("CONFIGSTORE_HOST"), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("could not dial configstore: %v", err)
		}
		configstore, err := storage.ConnectToConfigstore(ctx, grpcconn)
		if err != nil {
			log.Fatalf("could not dial configstore: %v", err)
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
				log.Printf("failed to get the cost matrix: %v\n", err)
			}

			if err := costmatrix.Optimize(&routematrix, 1); err != nil {
				log.Printf("failed to optimize cost matrix into route matrix: %v", err)
			}

			log.Printf("optimized %d entries into route matrix from cost matrix\n", len(routematrix.Entries))

			time.Sleep(10 * time.Second)
		}
	}()

	port := os.Getenv("RELAY_PORT")

	if len(port) == 0 {
		port = "40000"
		fmt.Printf("RELAY_PORT env var is unset, setting port as %s\n", port)
	}

	router := transport.NewRouter(redisClient, &geoClient, ipLocator, relayProvider, datacenterProvider, statsdb, &costmatrix, &routematrix, relayPublicKey, routerPrivateKey)

	go transport.HTTPStart(port, router)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
