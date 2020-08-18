package main

import (
	"fmt"
	"os"

	"github.com/networknext/backend/routing"
)

func optimizeCostMatrix(costFilename, routeFilename string, rtt int32) {
	var costMatrix routing.CostMatrix

	costFile, err := os.Open(costFilename)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could open the cost matrix file for reading: %v\n", err), 1)
	}
	defer costFile.Close()

	if _, err := costMatrix.ReadFrom(costFile); err != nil {
		handleRunTimeError(fmt.Sprintf("error reading cost matrix: %v\n", err), 1)
	}

	var routeMatrix routing.RouteMatrix
	if err := costMatrix.Optimize(&routeMatrix, rtt); err != nil {
		handleRunTimeError(fmt.Sprintf("error optimizing cost matrix: %v\n", err), 1)
	}

	routeFile, err := os.Create(routeFilename)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open the route matrix file for writing: %v\n", err), 1)
	}
	defer routeFile.Close()

	if _, err := routeMatrix.WriteTo(routeFile); err != nil {
		handleRunTimeError(fmt.Sprintf("error writing route matrix: %v\n", err), 1)
	}
}
