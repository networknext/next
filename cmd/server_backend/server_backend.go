/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"os"

	"github.com/networknext/backend/transport"
)

func main() {
	port := os.Getenv("NN_BACKEND_PORT")

	if len(port) == 0 {
		port = "30000"
	}

	router := transport.MakeRouter()

	go optimizeRoutine()

	go timeoutRoutine()

	go transport.HttpStart(port, router)

	for {
	}
}

// TODO
func optimizeRoutine() {
	fmt.Println("TODO optimizeRoutine()")
}

// TODO
func timeoutRoutine() {
	fmt.Println("TODO timeoutRoutine()")
}
