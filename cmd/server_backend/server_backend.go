/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/base64"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/oschwald/geoip2-golang"
	"google.golang.org/grpc"

	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := context.Background()

	// var serverPublicKey []byte
	var serverPrivateKey []byte
	var routerPrivateKey []byte
	{
		if key := os.Getenv("SERVER_BACKEND_PUBLIC_KEY"); len(key) != 0 {
			// serverPublicKey, _ = base64.StdEncoding.DecodeString(key)
			log.Printf("using SERVER_BACKEND_PUBLIC_KEY '%s'\n", key)
		} else {
			log.Fatal("env var 'SERVER_BACKEND_PUBLIC_KEY' is not set")
		}

		if key := os.Getenv("SERVER_BACKEND_PRIVATE_KEY"); len(key) != 0 {
			serverPrivateKey, _ = base64.StdEncoding.DecodeString(key)
			log.Printf("using SERVER_BACKEND_PRIVATE_KEY '%s'\n", key)
		} else {
			log.Fatal("env var 'SERVER_BACKEND_PRIVATE_KEY' is not set")
		}

		if key := os.Getenv("RELAY_ROUTER_PRIVATE_KEY"); len(key) != 0 {
			routerPrivateKey, _ = base64.StdEncoding.DecodeString(key)
			log.Printf("using RELAY_ROUTER_PRIVATE_KEY '%s'\n", key)
		} else {
			log.Fatal("env var 'RELAY_ROUTER_PRIVATE_KEY' is not set")
		}
	}

	// Attempt to connect to REDIS_HOST, falling back to local instance if not explicitly specified
	redisHost, ok := os.LookupEnv("REDIS_HOST")
	if !ok {
		redisHost = "localhost:6379"
		log.Printf("env var 'REDIS_HOST' is not set, falling back to default value of '%s'\n", redisHost)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: redisHost})
	if err := redisClient.Ping().Err(); err != nil {
		log.Fatalf("unable to connect to REDIS_HOST '%s'", redisHost)
	}

	// Open the Maxmind DB and create a routing.MaxmindDB from it
	var ipLocator routing.IPLocator = routing.NullIsland
	if filename, ok := os.LookupEnv("MAXMIND_DB_URI"); ok {
		if mmreader, err := geoip2.Open(filename); err != nil {
			if err != nil {
				log.Fatalf("failed to open Maxmind GeoIP2 database: %v", err)
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
						n, err := routeMatrix.ReadFrom(matrixReader)
						if err != nil {
							log.Printf("failed to read route matrix: %v\n", err)
						}

						log.Printf("read %d bytes from %s into route matrix for %d entries\n", n, uri, len(routeMatrix.Entries))
					}

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
			SessionUpdateHandlerFunc: transport.SessionUpdateHandlerFunc(redisClient, buyerProvider, &routeMatrix, ipLocator, &geoClient, serverPrivateKey, routerPrivateKey),
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
