package main

import (
	"github.com/modood/table"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func datacenters(rpcClient jsonrpc.RPCClient, filter string) {
	args := localjsonrpc.DatacentersArgs{
		Name: filter,
	}

	var reply localjsonrpc.DatacentersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Datacenters", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	table.Output(reply.Datacenters)
}
