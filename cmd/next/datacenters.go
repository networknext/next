package main

import (
	"fmt"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func datacenterMaps(rpcClient jsonrpc.RPCClient, env Environment, buyerID string) {
	args := localjsonrpc.DatacenterMapsArgs{
		ID: buyerID,
	}

	var reply localjsonrpc.DatacenterMapsReply
	if err := rpcClient.CallFor(&reply, "BuyersService.DatacenterMaps", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	table.Output(reply.DatacenterMaps)

}

// func addDatacenterMap(rpcClient jsonrpc.RPCClient, env Environment, ))

func datacenters(rpcClient jsonrpc.RPCClient, env Environment, filter string) {
	args := localjsonrpc.DatacentersArgs{
		Name: filter,
	}

	var reply localjsonrpc.DatacentersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Datacenters", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	table.Output(reply.Datacenters)
}

func addDatacenter(rpcClient jsonrpc.RPCClient, env Environment, datacenter routing.Datacenter) {
	args := localjsonrpc.AddDatacenterArgs{
		Datacenter: datacenter,
	}

	var reply localjsonrpc.AddDatacenterReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddDatacenter", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Datacenter \"%s\" added to storage.\n", datacenter.Name)
}

func removeDatacenter(rpcClient jsonrpc.RPCClient, env Environment, name string) {
	args := localjsonrpc.RemoveDatacenterArgs{
		Name: name,
	}

	var reply localjsonrpc.RemoveDatacenterReply
	if err := rpcClient.CallFor(&reply, "OpsService.RemoveDatacenter", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Datacenter \"%s\" removed from storage.\n", name)
}
