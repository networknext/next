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
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/alicebob/miniredis"
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

	if key := os.Getenv("RELAY_KEY_PUBLIC"); len(key) != 0 {
		relayPublicKey, _ = base64.StdEncoding.DecodeString(key)
	} else {
		log.Println("Env var 'RELAY_KEY_PUBLIC' is not set, exiting!")
		os.Exit(1)
	}

	if key := os.Getenv("ROUTER_KEY_PRIVATE"); len(key) != 0 {
		routerPrivateKey, _ = base64.StdEncoding.DecodeString(key)
	} else {
		log.Println("Env var 'ROUTER_KEY_PRIVATE' is not set, exiting!")
		os.Exit(1)
	}

	// Create an in-memory relay & datacenter store
	inMemoryProvider := storage.NewInMemory()
	var relayProvider transport.RelayProvider = &inMemoryProvider.RelayStore
	var datacenterProvider transport.DatacenterProvider = &inMemoryProvider.DatacenterStore

	var ipLocator routing.IPLocator

	initStubbedData(&inMemoryProvider, &ipLocator)

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

	// Attempt to connect to REDIS_HOST
	// If it fails to connect then start a local in memory instance and connect to that instead
	redisClient := redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_HOST")})
	if err := redisClient.Ping().Err(); err != nil {
		redisServer, err := miniredis.Run()
		if err != nil {
			log.Fatal(err)
		}

		redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
		if err := redisClient.Ping().Err(); err != nil {
			log.Fatal(err)
		}

		log.Printf("unable to connect to REDIS_HOST '%s', connected to in-memory redis %s", os.Getenv("REDIS_HOST"), redisServer.Addr())
	}

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "RELAY_LOCATIONS",
	}

	if uri, set := os.LookupEnv("MAXMIND_DB_URI"); set {
		mmreader, err := geoip2.Open(uri)
		if err != nil {
			log.Fatalf("failed to open Maxmind GeoIP2 database: %v", err)
		}
		ipLocator = &routing.MaxmindDB{
			Reader: mmreader,
		}
		defer mmreader.Close()
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

func initStubbedData(inMemory *storage.InMemory, ipLocator *routing.IPLocator) {
	if _, set := os.LookupEnv("RELAY_DEBUG"); set {
		filename := os.Getenv("RELAY_DEBUG_FILENAME")

		var data []byte
		if len(filename) > 0 {
			fmt.Println("Using debug file for fake data")
			data, _ = ioutil.ReadFile(filename)
		} else {
			data = []byte("{}")
		}

		type fakeRelayData struct {
			Latitude       float64
			Longitude      float64
			DatacenterName string
		}

		var fakeData map[string]fakeRelayData
		json.Unmarshal(data, &fakeData)

		type latLong struct {
			Latitude  float64
			Longitude float64
		}

		ipToLatLong := make(map[string]latLong)

		for k, v := range fakeData {
			inMemory.RelayStore.RelaysToDatacenterName[uint32(crypto.HashID(k))] = v.DatacenterName

			if udp, err := net.ResolveUDPAddr("udp", k); err == nil {
				ipToLatLong[udp.IP.String()] = latLong{
					Latitude:  v.Latitude,
					Longitude: v.Longitude,
				}
			}
		}

		*ipLocator = routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
			ll, ok := ipToLatLong[ip.String()]
			if ok {
				return routing.Location{
					Latitude:  ll.Latitude,
					Longitude: ll.Longitude,
				}, nil
			}

			return routing.Location{
				Latitude:  0.0,
				Longitude: 0.0,
			}, fmt.Errorf("could not locate lat/long for %s", ip.String())
		})
	}
}
