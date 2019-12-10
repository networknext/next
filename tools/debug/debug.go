/*
   Network Next. You control the network.
   Copyright © 2017 - 2019 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"core"
)

func LoadRouteMatrix(filename string) *core.RouteMatrix {
	fmt.Printf("Loading '%s'\n", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("error: could not read %s\n", filename)
		os.Exit(1)
	}
	routeMatrix, err := core.ReadRouteMatrix(data)
	if err != nil {
		fmt.Printf("error: could not read route matrix\n")
		os.Exit(1)
	}
	return routeMatrix
}

func GetRelayIndex(routeMatrix *core.RouteMatrix, relayName string) int {
	for i := range routeMatrix.RelayNames {
		if routeMatrix.RelayNames[i] == relayName {
			return i
		}
	}
	return -1
}

func main() {

	args := os.Args[1:]

	if len(args) != 1 {
		fmt.Printf("\nUsage: 'next debug [relayname]'\n\n")
		return
	}

	fmt.Printf("\nWelcome to Network Next!\n\n")

	routeMatrix := LoadRouteMatrix("optimize.bin")

	relayName := args[0]

	relayIndex := GetRelayIndex(routeMatrix, relayName)

	if relayIndex == -1 {
		fmt.Printf("\nerror: can't find relay called '%s'\n\n", relayName)
		os.Exit(1)
	}

	fmt.Printf("\nDebug %s:\n\n", relayName)

	numRelays := len(routeMatrix.RelayIds)

	a := relayIndex
	for b := 0; b < numRelays; b++ {
		if a == b {
			continue
		}
		index := core.TriMatrixIndex(a, b)
		if routeMatrix.Entries[index].NumRoutes != 0 {
			fmt.Printf("    %*dms (%d) %s\n", 5, routeMatrix.Entries[index].RouteRTT[0], routeMatrix.Entries[index].NumRoutes, routeMatrix.RelayNames[b])
		} else {
			fmt.Printf("       ---- (0) %s\n", routeMatrix.RelayNames[b] )
		}
	}

	fmt.Printf("\n")
}
