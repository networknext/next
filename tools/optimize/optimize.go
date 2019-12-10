/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"core"
	"fmt"
	"io/ioutil"
	"log"
)

const ThresholdRTT = 1.0

func WriteResult(filename string, result *core.RouteMatrix) {
	fmt.Printf("Writing result to '%s'\n", filename)
	buffer := make([]byte, 20*1024*1024)
	buffer = core.WriteRouteMatrix(buffer, result)
	err := ioutil.WriteFile(filename, buffer, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	fmt.Printf("\nWelcome to Network Next!\n\n")

	raw, err := ioutil.ReadFile("cost.bin")
	if err != nil {
		log.Fatalf(err.Error())
	}

	costMatrix, err := core.ReadCostMatrix(raw)
	if err != nil {
		log.Fatalf(err.Error())
	}

	routeMatrix := core.Optimize(costMatrix, ThresholdRTT)

	WriteResult("optimize.bin", routeMatrix)

	fmt.Printf("\nFinished.\n\n")
}
