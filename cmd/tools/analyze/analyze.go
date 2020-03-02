/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/routing"
)

type routeData struct {
	Improvement int32
}

func Analyze(name string, unit string, routeMatrix *routing.RouteMatrix) {

	src := routeMatrix.RelayIDs
	dest := routeMatrix.RelayIDs

	entries := make([]*routeData, 0, len(src)*len(dest))

	numRelayPairs := 0.0
	numValidRelayPairs := 0.0
	numValidRelayPairsWithoutImprovement := 0.0

	buckets := make([]int, 11)

	for i := range src {
		for j := range dest {
			if j < i {
				numRelayPairs++
				abFlatIndex := core.TriMatrixIndex(i, j)
				if len(routeMatrix.Entries[abFlatIndex].RouteRTT) > 0 {
					numValidRelayPairs++
					improvement := routeMatrix.Entries[abFlatIndex].DirectRTT - routeMatrix.Entries[abFlatIndex].RouteRTT[0]
					if improvement > 0.0 {
						entry := &routeData{}
						entry.Improvement = improvement
						entries = append(entries, entry)
						if improvement <= 5 {
							buckets[0]++
						} else if improvement <= 10 {
							buckets[1]++
						} else if improvement <= 15 {
							buckets[2]++
						} else if improvement <= 20 {
							buckets[3]++
						} else if improvement <= 25 {
							buckets[4]++
						} else if improvement <= 30 {
							buckets[5]++
						} else if improvement <= 35 {
							buckets[6]++
						} else if improvement <= 40 {
							buckets[7]++
						} else if improvement <= 45 {
							buckets[8]++
						} else if improvement <= 50 {
							buckets[9]++
						} else {
							buckets[10]++
						}
					} else {
						numValidRelayPairsWithoutImprovement++
					}
				}
			}
		}
	}

	fmt.Printf("\n%s Improvement:\n\n", name)

	fmt.Printf("    None: %d (%.2f%%)\n", int(numValidRelayPairsWithoutImprovement), numValidRelayPairsWithoutImprovement/numValidRelayPairs*100.0)

	for i := range buckets {
		if i != len(buckets)-1 {
			fmt.Printf("    %d-%d%s: %d (%.2f%%)\n", i*5, (i+1)*5, unit, buckets[i], float64(buckets[i])/numValidRelayPairs*100.0)
		} else {
			fmt.Printf("    %d%s+: %d (%.2f%%)\n", i*5, unit, buckets[i], float64(buckets[i])/numValidRelayPairs*100.0)
		}
	}

	totalRoutes := uint64(0)
	maxRouteLength := int32(0)
	maxRoutesPerRelayPair := int32(0)
	relayPairsWithNoRoutes := 0
	relayPairsWithOneRoute := 0
	averageRouteLength := 0.0

	for i := range src {
		for j := range dest {
			if j < i {
				ijFlatIndex := core.TriMatrixIndex(i, j)
				n := routeMatrix.Entries[ijFlatIndex].NumRoutes
				if n > maxRoutesPerRelayPair {
					maxRoutesPerRelayPair = n
				}
				totalRoutes += uint64(n)
				if n == 0 {
					relayPairsWithNoRoutes++
				}
				if n == 1 {
					relayPairsWithOneRoute++
				}
				for k := 0; k < int(routeMatrix.Entries[ijFlatIndex].NumRoutes); k++ {
					numRelays := routeMatrix.Entries[ijFlatIndex].RouteNumRelays[k]
					averageRouteLength += float64(numRelays)
					if numRelays > maxRouteLength {
						maxRouteLength = numRelays
					}
				}
			}
		}
	}

	averageNumRoutes := float64(totalRoutes) / float64(numRelayPairs)
	averageRouteLength = averageRouteLength / float64(totalRoutes)

	fmt.Printf("\n%s Summary:\n\n", name)
	fmt.Printf("    %.1f routes per relay pair on average (%d max)\n", averageNumRoutes, maxRoutesPerRelayPair)
	fmt.Printf("    %.1f relays per route on average (%d max)\n", averageRouteLength, maxRouteLength)
	fmt.Printf("    %.1f%% of relay pairs have no route\n", float64(relayPairsWithNoRoutes)/float64(numRelayPairs)*100)
	fmt.Printf("    %.1f%% of relay pairs have only one route\n", float64(relayPairsWithOneRoute)/float64(numRelayPairs)*100)
	fmt.Printf("\n")
}

func main() {
	var routeMatrix routing.RouteMatrix
	_, err := routeMatrix.ReadFrom(os.Stdin)
	if err != nil {
		log.Fatalln(fmt.Errorf("error reading route matrix from stdin: %w", err))
	}

	Analyze("RTT", "ms", &routeMatrix)
}
