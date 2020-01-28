/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/base64"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v7"
	"github.com/oschwald/geoip2-golang"
	"google.golang.org/grpc"

	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
)

func main() {
	ctx := context.Background()

	var err error

	var serverPrivateKey []byte
	var routerPrivateKey []byte

	if key := os.Getenv("SERVER_KEY_PRIVATE"); len(key) != 0 {
		serverPrivateKey, _ = base64.StdEncoding.DecodeString(key)
	} else {
		log.Fatal("env var 'SERVER_KEY_PRIVATE' is not set")
	}

	if key := os.Getenv("ROUTER_KEY_PRIVATE"); len(key) != 0 {
		routerPrivateKey, _ = base64.StdEncoding.DecodeString(key)
	} else {
		log.Fatal("env var 'ROUTER_KEY_PRIVATE' is not set")
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

		log.Printf("unable to connect to REDIS_HOST '%s', connected to in-memory redis %s", os.Getenv("REDIS_URL"), redisServer.Addr())
	}

	// Open the Maxmind DB and create a routing.MaxmindDB from it
	mmreader, err := geoip2.Open(os.Getenv("MAXMIND_DB_URI"))
	if err != nil {
		log.Fatalf("failed to open Maxmind GeoIP2 database: %v", err)
	}
	mmdb := routing.MaxmindDB{
		Reader: mmreader,
	}
	defer mmreader.Close()

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "RELAY_LOCATIONS",
	}

	// Create an in-memory buyer provider
	var buyerProvider transport.BuyerProvider = &storage.InMemory{}
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
		buyerProvider = configstore.Buyers
	}

	// For demo reaons just read a local cost.bin file and optimze it once.
	// This will change so we can periodically get an up to date RouteMatrix
	// to get Routes for Sessions.
	var routeMatrix routing.RouteMatrix
	{
		if os.Getenv("ROUTE_MATRIX_URI") != "" {
			go func() {
				for {
					res, err := http.Get(os.Getenv("ROUTE_MATRIX_URI"))
					if err != nil {
						log.Fatalf("failed to get route matrix: %v\n", err)
					}

					n, err := routeMatrix.ReadFom(res.Body)
					if err != nil {
						log.Printf("failed to read route matrix: %v\n", err)
					}

					log.Printf("read %d bytes into route matrix for %d entries\n", n, len(routeMatrix.Entries))

					time.Sleep(10 * time.Second)
				}
			}()
		}
	}

	{
		addr := net.UDPAddr{
			Port: 30000,
			IP:   net.ParseIP("0.0.0.0"),
		}

		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			log.Printf("error: could not listen on %s\n", addr.String())
		}

		mux := transport.UDPServerMux{
			Conn:          conn,
			MaxPacketSize: transport.DefaultMaxPacketSize,

			ServerUpdateHandlerFunc:  transport.ServerUpdateHandlerFunc(redisClient, buyerProvider),
			SessionUpdateHandlerFunc: transport.SessionUpdateHandlerFunc(redisClient, buyerProvider, nil, &mmdb, &geoClient, serverPrivateKey, routerPrivateKey),
		}

		go func() {
			log.Printf("started on %s\n", addr.String())
			if err := mux.Start(ctx, runtime.NumCPU()); err != nil {
				log.Println(err)
			}
		}()
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
