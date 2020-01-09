/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v7"

	"github.com/networknext/backend/transport"
)

func main() {
	var err error

	var port int64
	if port, err = strconv.ParseInt(os.Getenv("SERVER_BACKEND_PORT"), 10, 64); err != nil {
		port = 30000
		log.Printf("unable to parse SERVER_BACKEND_PORT '%s', defaulting to 30000\n", os.Getenv("SERVER_BACKEND_PORT"))
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

	{
		addr := net.UDPAddr{
			Port: int(port),
			IP:   net.ParseIP("0.0.0.0"),
		}

		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			log.Printf("error: could not listen on %s\n", addr.String())
		}

		mux := transport.UDPServerMux{
			Conn:          conn,
			MaxPacketSize: transport.DefaultMaxPacketSize,

			ServerUpdateHandlerFunc:  transport.ServerUpdateHandlerFunc(redisClient),
			SessionUpdateHandlerFunc: transport.SessionUpdateHandlerFunc(redisClient),
		}

		if err := mux.Start(); err != nil {
			log.Println(err)
		}
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
