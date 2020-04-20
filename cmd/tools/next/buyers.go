package main

import (
	"fmt"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func buyers(rpcClient jsonrpc.RPCClient) {
	args := localjsonrpc.BuyersArgs{}

	var reply localjsonrpc.BuyersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Buyers", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	table.Output(reply.Buyers)
}

func addBuyer(rpcClient jsonrpc.RPCClient, buyer routing.Buyer) {
	args := localjsonrpc.AddBuyerArgs{
		Buyer: buyer,
	}

	var reply localjsonrpc.AddBuyerReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddBuyer", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	fmt.Printf("Buyer \"%s\" added to storage.\n", buyer.Name)
}

func removeBuyer(rpcClient jsonrpc.RPCClient, id uint64) {
	args := localjsonrpc.RemoveBuyerArgs{
		ID: id,
	}

	var reply localjsonrpc.RemoveBuyerReply
	if err := rpcClient.CallFor(&reply, "OpsService.RemoveBuyer", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	fmt.Printf("Buyer with ID \"%d\" removed from storage.\n", id)
}

func routeShader(rpcClient jsonrpc.RPCClient, name string) {
	args := localjsonrpc.RouteShaderArgs{
		Name: name,
	}

	var reply localjsonrpc.RouteShaderReply
	if err := rpcClient.CallFor(&reply, "OpsService.RouteShader", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	table.Output(reply.RoutingRuleSettings)
}

func setRouteShader(rpcClient jsonrpc.RPCClient, name string, rrs routing.RoutingRulesSettings) {
	args := localjsonrpc.SetRouteShaderArgs{
		Name:                name,
		RoutingRuleSettings: rrs,
	}

	var reply localjsonrpc.SetRouteShaderReply
	if err := rpcClient.CallFor(&reply, "OpsService.SetRouteShader", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	fmt.Printf("Route shader for buyer \"%s\" updated.\n", name)
}
