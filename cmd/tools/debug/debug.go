/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/networknext/backend/core"
)

func GetRelayIndex(routeMatrix *core.RouteMatrix, relayName string) int {
	for i := range routeMatrix.RelayNames {
		if routeMatrix.RelayNames[i] == relayName {
			return i
		}
	}
	return -1
}

func main() {
	relay := flag.String("relay", "", "name of the relay")
	flag.Parse()

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal("error reading from stdin")
	}

	routeMatrix, err := core.ReadRouteMatrix(data)
	if err != nil {
		log.Fatalln("error reading route matrix")
	}

	relayName := *relay

	relayIndex := GetRelayIndex(routeMatrix, relayName)

	if relayIndex == -1 {
		log.Fatalf("error: can't find relay called '%s'\n", relayName)
	}

	fmt.Printf("Debug %s:\n", relayName)

	numRelays := len(routeMatrix.RelayIDs)

	a := relayIndex
	for b := 0; b < numRelays; b++ {
		if a == b {
			continue
		}
		index := core.TriMatrixIndex(a, b)
		if routeMatrix.Entries[index].NumRoutes != 0 {
			fmt.Printf("    %*dms (%d) %s\n", 5, routeMatrix.Entries[index].RouteRTT[0], routeMatrix.Entries[index].NumRoutes, routeMatrix.RelayNames[b])
		} else {
			fmt.Printf("       ---- (0) %s\n", routeMatrix.RelayNames[b])
		}
	}
}
