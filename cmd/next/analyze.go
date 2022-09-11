package main

import (
	"fmt"
	"io/ioutil"

	"github.com/networknext/backend/modules/common"
)

func analyzeRouteMatrix(inputFile string) {

	routeMatrixFilename := "optimize.bin"

	routeMatrixData, err := ioutil.ReadFile(routeMatrixFilename)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not read the route matrix file: %v\n", err), 1)
	}

	var routeMatrix common.RouteMatrix

	err = routeMatrix.Read(routeMatrixData)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("error reading route matrix: %v\n", err), 1)
	}

	analysis := routeMatrix.Analyze()

	fmt.Printf("RTT Improvement\n\n")

	fmt.Printf("    None: %.1f%%\n", analysis.RTTBucket_NoImprovement)
	fmt.Printf("    0-5ms: %.1f%%\n", analysis.RTTBucket_0_5ms)
	fmt.Printf("    5-10ms: %.1f%%\n", analysis.RTTBucket_5_10ms)
	fmt.Printf("    10-15ms: %.1f%%\n", analysis.RTTBucket_10_15ms)
	fmt.Printf("    15-20ms: %.1f%%\n", analysis.RTTBucket_15_20ms)
	fmt.Printf("    20-25ms: %.1f%%\n", analysis.RTTBucket_20_25ms)
	fmt.Printf("    25-30ms: %.1f%%\n", analysis.RTTBucket_25_30ms)
	fmt.Printf("    30-35ms: %.1f%%\n", analysis.RTTBucket_30_35ms)
	fmt.Printf("    35-40ms: %.1f%%\n", analysis.RTTBucket_35_40ms)
	fmt.Printf("    40-45ms: %.1f%%\n", analysis.RTTBucket_40_45ms)
	fmt.Printf("    45-50ms: %.1f%%\n", analysis.RTTBucket_45_50ms)
	fmt.Printf("    50ms+: %.1f%%\n", analysis.RTTBucket_50ms_Plus)

	fmt.Printf("\nRoute Summary:\n\n")

	fmt.Printf("    %d relays\n", len(routeMatrix.RelayIds))
	fmt.Printf("    %d total routes\n", analysis.TotalRoutes)
	fmt.Printf("    %d relay pairs\n", analysis.NumRelayPairs)
	fmt.Printf("    %.1f routes per-relay pair on average\n", analysis.AverageNumRoutes)
	fmt.Printf("    %.1f relays per-route on average\n", analysis.AverageRouteLength)
	fmt.Printf("    %.1f%% of relay pairs have only one route\n", float64(analysis.NumRelayPairsWithOneRoute)/float64(analysis.NumRelayPairs)*100)
	fmt.Printf("    %.1f%% of relay pairs have no route\n", float64(analysis.NumRelayPairsWithNoRoutes)/float64(analysis.NumRelayPairs)*100)
}
