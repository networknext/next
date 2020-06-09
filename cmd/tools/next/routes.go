package main

import (
	"fmt"
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

func saveCostMatrix(env Environment, filename string) {
	var uri string
	var err error

	if uri, err = env.RelayBackendHostname(); err != nil {
		log.Fatalf("Cannot get get relay backend hostname: %v\n", err)
	}

	uri += "/cost_matrix"

	if r, err := http.Get(uri); err == nil {
		if r.StatusCode == http.StatusOK {
			var matrix routing.CostMatrix
			if _, err := matrix.ReadFrom(r.Body); err == nil {
				if file, err := os.Create(filename); err == nil {
					if _, err := matrix.WriteTo(file); err != nil {
						log.Fatalln(fmt.Errorf("error writing cost matrix to file: %w", err))
					}
				} else {
					log.Fatalln(fmt.Errorf("could not open file for writing: %w", err))
				}
			} else {
				log.Fatalln(fmt.Errorf("error reading cost matrix: %w", err))
			}
		} else {
			log.Fatalf("relay backend returns non 200 response code: %d\n", r.StatusCode)
		}
	} else {
		log.Fatalln(fmt.Errorf("could not get the route matrix from the backend: %w", err))
	}
}

func optimizeCostMatrix(costFile, routeFile string, rtt int32) {
	var costMatrix routing.CostMatrix
	if file, err := os.Open(costFile); err == nil {
		if _, err := costMatrix.ReadFrom(file); err != nil {
			log.Fatalln(fmt.Errorf("error reading cost matrix: %w", err))
		}
	} else {
		log.Fatalln(fmt.Errorf("could open the cost matrix file for reading: %w", err))
	}

	var routeMatrix routing.RouteMatrix
	if err := costMatrix.Optimize(&routeMatrix, rtt); err != nil {
		log.Fatalln(fmt.Errorf("error optimizing cost matrix: %w", err))
	}

	if file, err := os.Create(routeFile); err == nil {
		if _, err := routeMatrix.WriteTo(file); err != nil {
			log.Fatalln(fmt.Errorf("error writing route matrix: %w", err))
		}
	} else {
		log.Fatalln(fmt.Errorf("could not open the route matrix file for writing: %w", err))
	}
}

func analyzeRouteMatrix(inputFile string) {
	if file, err := os.Open(inputFile); err == nil {
		var routeMatrix routing.RouteMatrix
		if _, err := routeMatrix.ReadFrom(file); err == nil {
			routeMatrix.WriteAnalysisTo(os.Stdout)
			routeMatrix.WriteRoutesTo(os.Stdout)
		} else {
			log.Fatalln(fmt.Errorf("error reading route matrix: %w", err))
		}
	} else {
		log.Fatalln(fmt.Errorf("could not open the route matrix file for reading: %w", err))
	}
}
