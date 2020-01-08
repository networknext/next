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
	"github.com/gomodule/redigo/redis"

	"github.com/networknext/backend/transport"
)

func main() {
	var err error

	var port int64
	if port, err = strconv.ParseInt(os.Getenv("SERVER_BACKEND_PORT"), 10, 64); err != nil {
		port = 30000
		log.Printf("unable to parse port %s, defauling to 30000\n", os.Getenv("SERVER_BACKEND_PORT"))
	}

	// Attempt to connect to REDIS_URL
	// If it fails to connect then start a local in memory instance and connect to that instead
	var redisConn redis.Conn
	if redisConn, err = redis.DialURL(os.Getenv("REDIS_URL")); err != nil {
		redisServer, err := miniredis.Run()
		if err != nil {
			log.Fatal(err)
		}

		redisConn, err = redis.Dial("tcp", redisServer.Addr())
		if err != nil {
			log.Fatal(err)
		}
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

			ServerUpdateHandlerFunc:  transport.ServerUpdateHandlerFunc(redisConn),
			SessionUpdateHandlerFunc: transport.SessionUpdateHandlerFunc(redisConn),
		}

		if err := mux.Start(); err != nil {
			log.Println(err)
		}
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
