package main

import (
	"fmt"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
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

func addDatacenter(rpcClient jsonrpc.RPCClient, datacenter routing.Datacenter) {
	args := localjsonrpc.AddDatacenterArgs{
		Datacenter: datacenter,
	}

	var reply localjsonrpc.AddDatacenterReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddDatacenter", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	fmt.Printf("Datacenter \"%s\" added to storage.\n", datacenter.Name)
}
