package main

import (
	"fmt"
	"os"
	"text/tabwriter"

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

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)

	fmt.Fprint(tw, "Next\tDirect\tRoute\n")
	for _, route := range reply.Routes {
		fmt.Fprintf(tw, "%.1f\t%.1f\t", route.Stats.RTT, route.DirectStats.RTT)
		for idx, relay := range route.Relays {
			fmt.Fprint(tw, relay.Name)
			if idx+1 < len(route.Relays) {
				fmt.Fprint(tw, " - ")
			}
		}
		fmt.Fprint(tw, "\n")
	}
	tw.Flush()
}
