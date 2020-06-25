package main

import (
	"fmt"
	"log"
	"os"

	"github.com/networknext/backend/routing"
)

func optimizeCostMatrix(costFilename, routeFilename string, rtt int32) {
	var costMatrix routing.CostMatrix

	costFile, err := os.Open(costFilename)
	if err != nil {
		log.Fatalln(fmt.Errorf("could open the cost matrix file for reading: %w", err))
	}
	defer costFile.Close()

	if _, err := costMatrix.ReadFrom(costFile); err != nil {
		log.Fatalln(fmt.Errorf("error reading cost matrix: %w", err))
	}

	var routeMatrix routing.RouteMatrix
	if err := costMatrix.Optimize(&routeMatrix, rtt); err != nil {
		log.Fatalln(fmt.Errorf("error optimizing cost matrix: %w", err))
	}

	routeFile, err := os.Create(routeFilename)
	if err != nil {
		log.Fatalln(fmt.Errorf("could not open the route matrix file for writing: %w", err))
	}
	defer routeFile.Close()

	if _, err := routeMatrix.WriteTo(routeFile); err != nil {
		log.Fatalln(fmt.Errorf("error writing route matrix: %w", err))
	}
}
