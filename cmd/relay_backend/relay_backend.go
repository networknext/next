/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

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

	statsdb := routing.NewStatsDatabase()
	var costmatrix routing.CostMatrix
	var routematrix routing.RouteMatrix

	go func() {
		for {
			if err := statsdb.GetCostMatrix(&costmatrix, redisClient); err != nil {
				log.Printf("failed to get the cost matrix: %v", err)
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

	router := transport.NewRouter(redisClient, statsdb, &costmatrix, &routematrix, relayPublicKey, routerPrivateKey)

	go transport.HTTPStart(port, router)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
