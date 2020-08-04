package main

import (
	"fmt"
	"os"

	"github.com/networknext/backend/routing"
)

func GetRelayIndex(routeMatrix *routing.RouteMatrix, relayName string) int {
	for i := range routeMatrix.RelayNames {
		if routeMatrix.RelayNames[i] == relayName {
			return i
		}
	}
	return -1
}

type RelayEntry struct {
	id   uint64
	name string
}

func debug(relayName string, inputFile string) {

	file, err := os.Open(inputFile)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open the route matrix file for reading: %v\n", err), 1)
	}
	defer file.Close()

	var routeMatrix routing.RouteMatrix
	if _, err := routeMatrix.ReadFrom(file); err != nil {
		handleRunTimeError(fmt.Sprintf("error reading route matrix: %v\n", err), 1)
	}

	relayIndex := GetRelayIndex(&routeMatrix, relayName)

	if relayIndex == -1 {
		handleRunTimeError(fmt.Sprintf("error: can't find relay called '%s'\n", relayName), 0)
	}

	fmt.Printf("Routes to '%s':\n\n", relayName)

	numRelays := len(routeMatrix.RelayIDs)

	relays := make([]RelayEntry, numRelays)
	for i := 0; i < numRelays; i++ {
		relays[i].id = routeMatrix.RelayIDs[i]
		relays[i].name = routeMatrix.RelayNames[i]
	}

	a := relayIndex

	for b := 0; b < numRelays; b++ {

		dest := relays[b]

		if a == b {
			continue
		}

		if b > 0 {
			fmt.Printf("\n")
		}

		index := routing.TriMatrixIndex(a, b)

		directRTT := routeMatrix.Entries[index].DirectRTT

		fmt.Printf("    %s (%d)\n", dest.name, directRTT)

		numRoutes := int(routeMatrix.Entries[index].NumRoutes)

		for i := 0; i < numRoutes; i++ {
			if i == 0 {
				fmt.Printf("\n")
			}
			routeRTT := routeMatrix.Entries[index].RouteRTT[i]
			routeNumRelays := int(routeMatrix.Entries[index].RouteNumRelays[i])
			fmt.Printf("    %*dms: ", 5, routeRTT)
			reverse := a >= b
			if reverse {
				for j := routeNumRelays - 1; j >= 0; j-- {
					fmt.Printf("%s", routeMatrix.RelayNames[routeMatrix.Entries[index].RouteRelays[i][j]])
					if j != 0 {
						fmt.Printf(" - ")
					} else {
						fmt.Printf("\n")
					}
				}
			} else {
				for j := 0; j < routeNumRelays; j++ {
					fmt.Printf("%s", routeMatrix.RelayNames[routeMatrix.Entries[index].RouteRelays[i][j]])
					if j != routeNumRelays-1 {
						fmt.Printf(" - ")
					} else {
						fmt.Printf("\n")
					}
				}
			}
		}
	}
}
