package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/routing"
)

func optimizeCostMatrix(costFilename, routeFilename string, costThreshold int32) {

	var costMatrix routing.CostMatrix

	costFile, err := os.Open(costFilename)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could open the cost matrix file for reading: %v\n", err), 1)
	}
	defer costFile.Close()

	if _, err := costMatrix.ReadFrom(costFile); err != nil {
		handleRunTimeError(fmt.Sprintf("error reading cost matrix: %v\n", err), 1)
	}

	numRelays := len(costMatrix.RelayIDs)

	numDestRelays := 0
	for i := range costMatrix.DestRelays {
		if costMatrix.DestRelays[i] {
			numDestRelays++
		}
	}

	numCPUs := runtime.NumCPU()
	numSegments := numRelays
	if numCPUs < numRelays {
		numSegments = numRelays / 5
		if numSegments == 0 {
			numSegments = 1
		}
	}

	routeMatrix := &routing.RouteMatrix{
		Version:            routing.RouteMatrixSerializeVersion,
		RelayIDs:           costMatrix.RelayIDs,
		RelayAddresses:     costMatrix.RelayAddresses,
		RelayNames:         costMatrix.RelayNames,
		RelayLatitudes:     costMatrix.RelayLatitudes,
		RelayLongitudes:    costMatrix.RelayLongitudes,
		RelayDatacenterIDs: costMatrix.RelayDatacenterIDs,
		DestRelays:         costMatrix.DestRelays,
		RouteEntries:       core.Optimize2(numRelays, numSegments, costMatrix.Costs, costThreshold, costMatrix.RelayDatacenterIDs, costMatrix.DestRelays),
	}

	routeFile, err := os.Create(routeFilename)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open the route matrix file for writing: %v\n", err), 1)
	}
	defer routeFile.Close()

	if _, err := routeMatrix.WriteTo(routeFile, 100*1000*1000); err != nil {
		handleRunTimeError(fmt.Sprintf("error writing route matrix: %v\n", err), 1)
	}
}
