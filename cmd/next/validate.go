package main

import (
	"fmt"
	"os"

	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func validate(rpcClient jsonrpc.RPCClient, env Environment, inputFile string) {
	file, err := os.Open(inputFile)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open the route matrix file for reading: %v\n", err), 1)
	}
	defer file.Close()

	var routeMatrix routing.RouteMatrix
	if _, err := routeMatrix.ReadFrom(file); err != nil {
		handleRunTimeError(fmt.Sprintf("error reading route matrix: %v\n", err), 1)
	}

	var reply localjsonrpc.DatacentersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Datacenters", localjsonrpc.DatacentersArgs{}); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	// Loop over all DCs and make sure there is a relay within them
	for _, dc := range allDCs {
		dcRelaysIDs := routeMatrix.GetDatacenterRelayIDs(dc.ID)
		if len(dcRelaysIDs) == 0 {
			fmt.Println("No dc relays found")
		}
	}
}
