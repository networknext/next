package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
)

func optimizeCostMatrix(costMatrixFilename, routeMatrixFilename string, costThreshold int32) {

	costMatrixData, err := os.ReadFile(costMatrixFilename)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not read the cost matrix file: %v\n", err), 1)
	}

	var costMatrix common.CostMatrix

	err = costMatrix.Read(costMatrixData)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("error reading cost matrix: %v\n", err), 1)
	}

	numRelays := len(costMatrix.RelayIds)

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

	routeMatrix := &common.RouteMatrix{
		Version:            common.RouteMatrixVersion_Write,
		RelayIds:           costMatrix.RelayIds,
		RelayAddresses:     costMatrix.RelayAddresses,
		RelayNames:         costMatrix.RelayNames,
		RelayLatitudes:     costMatrix.RelayLatitudes,
		RelayLongitudes:    costMatrix.RelayLongitudes,
		RelayDatacenterIds: costMatrix.RelayDatacenterIds,
		DestRelays:         costMatrix.DestRelays,
		RouteEntries:       core.Optimize2(numRelays, numSegments, costMatrix.Costs, costThreshold, costMatrix.RelayDatacenterIds, costMatrix.DestRelays),
	}

	routeMatrixData, err := routeMatrix.Write(100 * 1024 * 1024)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not write route matrix: %v", err), 1)
	}

	err = os.WriteFile(routeMatrixFilename, routeMatrixData, 0644)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open the route matrix file for writing: %v\n", err), 1)
	}

	// todo: temporary -- print out route matrix as csv

	fmt.Printf(",")
	for i := range costMatrix.RelayNames {
		fmt.Printf("%s,", costMatrix.RelayNames[i])
	}
	fmt.Printf("\n")
	for i := range costMatrix.RelayNames {
		fmt.Printf("%s,", costMatrix.RelayNames[i])
		for j := range costMatrix.RelayNames {
			if i == j {
				fmt.Printf("-1,")
			} else {
				index := core.TriMatrixIndex(i, j)
				cost := costMatrix.Costs[index]
				fmt.Printf("%d,", cost)
			}
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")

	// todo: print out dest relays
	fmt.Printf("dest relays: ")
	for i := range costMatrix.RelayNames {
		if costMatrix.DestRelays[i] {
			fmt.Printf("%s,", costMatrix.RelayNames[i])
		}
	}
	fmt.Printf("\n\n")
}
