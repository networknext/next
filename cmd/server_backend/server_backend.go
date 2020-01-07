/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/networknext/backend/transport"
)

func main() {
	backend := transport.NewBackend()
	port := os.Getenv("NN_BACKEND_PORT")

	if len(port) == 0 {
		port = "30000"
	}

	router := transport.MakeRouter(backend)

	go optimizeRoutine()

	go timeoutRoutine()

	go transport.HTTPStart(port, router)

	// so my pc doesn't kill itself with a infinite loop
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
