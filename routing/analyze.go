package routing

import (
	"fmt"
	"io"
)

func Analyze(writer io.Writer, routeMatrix *RouteMatrix) {
	type routeData struct {
		Improvement int32
	}

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
				abFlatIndex := TriMatrixIndex(i, j)
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

	fmt.Fprintf(writer, "\n%s Improvement:\n\n", "RTT")

	fmt.Fprintf(writer, "    None: %d (%.2f%%)\n", int(numValidRelayPairsWithoutImprovement), numValidRelayPairsWithoutImprovement/numValidRelayPairs*100.0)

	for i := range buckets {
		if i != len(buckets)-1 {
			fmt.Fprintf(writer, "    %d-%d%s: %d (%.2f%%)\n", i*5, (i+1)*5, "ms", buckets[i], float64(buckets[i])/numValidRelayPairs*100.0)
		} else {
			fmt.Fprintf(writer, "    %d%s+: %d (%.2f%%)\n", i*5, "ms", buckets[i], float64(buckets[i])/numValidRelayPairs*100.0)
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
				ijFlatIndex := TriMatrixIndex(i, j)
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

	fmt.Fprintf(writer, "\n%s Summary:\n\n", "RTT")
	fmt.Fprintf(writer, "    %.1f routes per relay pair on average (%d max)\n", averageNumRoutes, maxRoutesPerRelayPair)
	fmt.Fprintf(writer, "    %.1f relays per route on average (%d max)\n", averageRouteLength, maxRouteLength)
	fmt.Fprintf(writer, "    %.1f%% of relay pairs have no route\n", float64(relayPairsWithNoRoutes)/float64(numRelayPairs)*100)
	fmt.Fprintf(writer, "    %.1f%% of relay pairs have only one route\n", float64(relayPairsWithOneRoute)/float64(numRelayPairs)*100)
	fmt.Fprintf(writer, "\n")
}
