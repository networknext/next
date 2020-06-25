package main

import (
	"fmt"
	"log"
	"os"

	"github.com/networknext/backend/routing"
)

func analyzeRouteMatrix(inputFile string) {
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalln(fmt.Errorf("could not open the route matrix file for reading: %w", err))
	}
	defer file.Close()

	var routeMatrix routing.RouteMatrix
	if _, err := routeMatrix.ReadFrom(file); err != nil {
		log.Fatalln(fmt.Errorf("error reading route matrix: %w", err))
	}

	routeMatrix.WriteAnalysisTo(os.Stdout)
}
