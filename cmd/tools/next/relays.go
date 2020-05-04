package main

import (
	"fmt"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func relays(rpcClient jsonrpc.RPCClient, filter string) {
	args := localjsonrpc.RelaysArgs{
		Name: filter,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	table.Output(reply.Relays)
}
func addRelay(rpcClient jsonrpc.RPCClient, relay routing.Relay) {
	args := localjsonrpc.AddRelayArgs{
		Relay: relay,
	}

	var reply localjsonrpc.AddRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddRelay", args); err != nil {
		switch e := err.(type) {
		case *storage.UnknownSellerError:
			// Prompt to add seller
		case *storage.UnknownDatacenterError:
			// Prompt to add datacenter
		default:
			handleJSONRPCError(e)
		}

		return
	}

	fmt.Printf("Relay \"%s\" added to storage.\n", relay.Name)
}

func removeRelay(rpcClient jsonrpc.RPCClient, name string) {
	info := getRelayInfo(rpcClient, name)

	args := localjsonrpc.RemoveRelayArgs{
		RelayID: info.id,
	}

	var reply localjsonrpc.RemoveRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.RemoveRelay", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	fmt.Printf("Relay \"%s\" removed from storage.\n", name)
}
