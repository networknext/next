package main

import (
	"fmt"

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
