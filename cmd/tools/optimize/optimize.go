/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/networknext/backend/core"
)

func main() {
	rtt := flag.Int64("threshold-rtt", 1.0, "set the threshold RTT")
	flag.Parse()

	costraw, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln(fmt.Errorf("error reading from stdin: %w", err))
	}

	costmatrix, err := core.ReadCostMatrix(costraw)
	if err != nil {
		log.Fatalln(fmt.Errorf("error reading cost matrix: %w", err))
	}

	routeMatrix := core.Optimize(costmatrix, int32(*rtt))

	buffer := make([]byte, 20*1024*1024)
	buffer = core.WriteRouteMatrix(buffer, routeMatrix)

	buf := bytes.NewBuffer(buffer)
	if _, err := io.Copy(os.Stdout, buf); err != nil {
		log.Fatalln(fmt.Errorf("error writing optimize matrix to stdout: %w", err))
	}
}
