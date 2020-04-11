package main

import (
	"github.com/modood/table"
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
