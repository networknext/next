package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func routes(rpcClient jsonrpc.RPCClient, env Environment, srcrelays []string, destrelays []string, rtt float64, routehash uint64) {
	args := localjsonrpc.RouteSelectionArgs{
		SourceRelays:      srcrelays,
		DestinationRelays: destrelays,
		RTT:               rtt,
		RouteHash:         routehash,
	}

	var reply localjsonrpc.RouteSelectionReply
	if err := rpcClient.CallFor(&reply, "OpsService.RouteSelection", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	for _, route := range reply.Routes {
		fmt.Printf("Next RTT(%v) ", route.Stats.RTT)
		fmt.Printf("Direct RTT(%v) ", route.DirectStats.RTT)
		for _, relay := range route.Relays {
			fmt.Print(relay.Name, " ")
		}
		fmt.Println()
	}
}

// todo: move to cost
func saveCostMatrix(env Environment, filename string) {
	var uri string
	var err error

	if uri, err = env.RelayBackendURL(); err != nil {
		log.Fatalf("Cannot get get relay backend hostname: %v\n", err)
	}

	uri += "/cost_matrix"

	r, err := http.Get(uri)
	if err != nil {
		log.Fatalln(fmt.Errorf("could not get the route matrix from the backend: %w", err))
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		log.Fatalf("relay backend returns non 200 response code: %d\n", r.StatusCode)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalln(fmt.Errorf("could not open file for writing: %w", err))
	}
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		log.Fatalln(fmt.Errorf("error writing cost matrix to file: %w", err))
	}
}

// todo: move to optimize
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

// todo: move to analyze
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
