package main

import (
	"fmt"
	"os"

	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func validate(rpcClient jsonrpc.RPCClient, env Environment, relaysStateShowFlags [6]bool, relaysStateHideFlags [6]bool, inputFile string) {
	file, err := os.Open(inputFile)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open the route matrix file for reading: %v\n", err), 1)
	}
	defer file.Close()

	var routeMatrix routing.RouteMatrix
	if _, err := routeMatrix.ReadFrom(file); err != nil {
		handleRunTimeError(fmt.Sprintf("error reading route matrix: %v\n", err), 1)
	}

	var relayReply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&relayReply, "OpsService.Relays", localjsonrpc.RelaysArgs{}); err != nil {
		handleJSONRPCError(env, err)
	}

	var dcReply localjsonrpc.DatacentersReply
	if err := rpcClient.CallFor(&dcReply, "OpsService.Datacenters", localjsonrpc.DatacentersArgs{}); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	allRelays := relayReply.Relays
	allDCs := dcReply.Datacenters

	// Loop over all DCs and make sure there is a relay within them
	for _, dc := range allDCs {
		for _, relay := range allRelays {
			relayState, err := routing.ParseRelayState(relay.State)
			if err != nil {
				handleRunTimeError(fmt.Sprintf("could not parse invalid relay state %s\n", relay.State), 0)
			}

			includeRelay := true

			for i, flag := range relaysStateShowFlags {
				if flag {
					if relayState != routing.RelayState(i) {
						// An "only show" flag is set and this relay doesn't match that state, so don't include it in the final output
						includeRelay = false
					} else {
						// One of the flags should include the relay, so set to true and break out, since combining the flags is an OR operation
						includeRelay = true
						break
					}
				}
			}
			// IMPORTANT: This is an assumption based off naming convention - a relay with the same name as a datacenter is a dest relay so it must have the same ID as the datacenter
			if includeRelay && relay.Name == dc.Name && relay.DatacenterHexID != dc.HexID {
				fmt.Println(fmt.Sprintf("Datacenter - %s: %s, Relay - %s: %s", dc.Name, dc.HexID, relay.Name, relay.DatacenterHexID))
				fmt.Println("")
			}
		}
	}
}
