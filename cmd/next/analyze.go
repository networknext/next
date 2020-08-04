package main

import (
	"fmt"
	"os"

	"github.com/networknext/backend/routing"
)

func analyzeRouteMatrix(inputFile string) {
	file, err := os.Open(inputFile)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open the route matrix file for reading: %v\n", err), 1)
	}
	defer file.Close()

	var routeMatrix routing.RouteMatrix
	if _, err := routeMatrix.ReadFrom(file); err != nil {
		handleRunTimeError(fmt.Sprintf("error reading route matrix: %v\n", err), 1)
	}

	routeMatrix.WriteAnalysisTo(os.Stdout)
}
