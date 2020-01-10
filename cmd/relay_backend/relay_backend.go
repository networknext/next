/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/transport"
)

func main() {
	statsdb := core.NewStatsDatabase()
	backend := transport.NewStubbedBackend()
	port := os.Getenv("NN_PORT")

	if len(port) == 0 {
		port = "30000"
		fmt.Printf("NN_RELAY_BACKEND_PORT env var is unset, settings port as 30000")
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

	router := transport.NewRouter(redisClient, statsdb, backend)

	go optimizeRoutine()

	go timeoutRoutine()

	go transport.HTTPStart(port, router)

	// so my pc doesn't kill itself with an infinite loop
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}

// TODO
func optimizeRoutine() {
	fmt.Println("TODO optimizeRoutine()")
}

// TODO
func timeoutRoutine() {
	fmt.Println("TODO timeoutRoutine()")
}
