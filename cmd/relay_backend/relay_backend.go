/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
)

func main() {
	relaydb := core.NewRelayDatabase()
	statsdb := core.NewStatsDatabase()

	var costmatrix routing.CostMatrix
	var routematrix routing.RouteMatrix
	go func() {
		for {
			if err := costmatrix.Optimize(&routematrix, 1); err != nil {
				log.Printf("failed to optimize cost matrix into route matrix: %v", err)
			}

			log.Printf("optimized %d entries into route matrix from cost matrix\n", len(routematrix.Entries))

			time.Sleep(10 * time.Second)
		}
	}()

	port := os.Getenv("NN_RELAY_BACKEND_PORT")

	if len(port) == 0 {
		port = "30000"
	}

	router := transport.NewRouter(relaydb, statsdb, &costmatrix, &routematrix)

	go transport.HTTPStart(port, router)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
