package main

import (
	"fmt"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

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

func listDatacenterMaps(rpcClient jsonrpc.RPCClient, env Environment, datacenter string) {
	dcID := returnDatacenterID(rpcClient, env, datacenter)

	if dcID == "" {
		fmt.Printf("Datacenter '%s' not found.\n", datacenter)
		return
	}

	var reply localjsonrpc.ListDatacenterMapsReply
	if err := rpcClient.CallFor(&reply, "OpsService.ListDatacenterMaps", dcID); err != nil {
		handleJSONRPCError(env, err)
		return
	}

}
