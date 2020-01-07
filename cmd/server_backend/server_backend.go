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

	"github.com/networknext/backend/transport"
)

func main() {
	var err error

	var port int64
	if port, err = strconv.ParseInt(os.Getenv("SERVER_BACKEND_PORT"), 10, 64); err != nil {
		port = 30000
		log.Printf("unable to parse port %s, defauling to 30000\n", os.Getenv("SERVER_BACKEND_PORT"))
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

			ServerUpdateHandlerFunc:  transport.ServerUpdateHandlerFunc,
			SessionUpdateHandlerFunc: transport.SessionUpdateHandlerFunc,
		}

		if err := mux.Start(); err != nil {
			log.Println(err)
		}
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
