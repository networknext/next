package main

import (
	"os"
	"log"
	"fmt"
	"sort"

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
	id uint64
	name string
	index int
}

func debug(relayName string, inputFile string) {

	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalln(fmt.Errorf("could not open the route matrix file for reading: %w", err))
	}
	defer file.Close()

	var routeMatrix routing.RouteMatrix
	if _, err := routeMatrix.ReadFrom(file); err != nil {
		log.Fatalln(fmt.Errorf("error reading route matrix: %w", err))
	}

	relayIndex := GetRelayIndex(&routeMatrix, relayName)

	if relayIndex == -1 {
		log.Fatalf("error: can't find relay called '%s'\n", relayName)
	}

	fmt.Printf("Routes to '%s':\n\n", relayName)

	numRelays := len(routeMatrix.RelayIDs)

	relays := make([]RelayEntry, numRelays)
	for i := 0; i < numRelays; i++ {
		relays[i].id = routeMatrix.RelayIDs[i]
		relays[i].name = routeMatrix.RelayNames[i]
		relays[i].index = i
	}

	sort.Slice(relays, func(i int, j int) bool {
		return relays[i].name < relays[j].name
	})

	a := relayIndex

	for b := 0; b < numRelays; b++ {

		if b > 0 {
			fmt.Printf("\n")
		}

		dest := relays[b]

		if a == dest.index {
			continue
		}

		fmt.Printf("    %s:\n\n", dest.name)

		index := routing.TriMatrixIndex(a, dest.index)

		numRoutes := int(routeMatrix.Entries[index].NumRoutes)

		for i := 0; i < numRoutes; i++ {
			routeRTT := routeMatrix.Entries[index].RouteRTT[i]
			routeNumRelays := int(routeMatrix.Entries[index].RouteNumRelays[i])
			fmt.Printf("    %*dms: ", 5, routeRTT)
			for j := 0; j < routeNumRelays; j++ {
				fmt.Printf( "%s", routeMatrix.RelayNames[routeMatrix.Entries[index].RouteRelays[i][j]])
				if j != routeNumRelays -1 {
					fmt.Printf(" - ")
				} else {
					fmt.Printf("\n")
				}
			}
		}
	}
}
