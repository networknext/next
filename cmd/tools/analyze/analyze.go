/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/networknext/backend/routing"
)

func main() {
	var routeMatrix routing.RouteMatrix
	_, err := routeMatrix.ReadFrom(os.Stdin)
	if err != nil {
		log.Fatalln(fmt.Errorf("error reading route matrix from stdin: %w", err))
	}

	routeMatrix.WriteAnalysisTo(os.Stdout)
}
