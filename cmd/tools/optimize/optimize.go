/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/networknext/backend/routing"
)

func main() {
	rtt := flag.Int64("threshold-rtt", 1.0, "set the threshold RTT")
	flag.Parse()

	var costMatrix routing.CostMatrix
	_, err := costMatrix.ReadFrom(os.Stdin)
	if err != nil {
		log.Fatalln(fmt.Errorf("error reading cost matrix from stdin: %w", err))
	}

	fmt.Println("Analyzing route matrix")

	var routeMatrix routing.RouteMatrix
	err = costMatrix.Optimize(&routeMatrix, int32(*rtt))

	if _, err := routeMatrix.WriteTo(os.Stdout); err != nil {
		log.Fatalln(fmt.Errorf("error writing route matrix to stdout: %w", err))
	}

}
