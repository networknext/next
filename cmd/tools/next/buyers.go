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

func routingRulesSettings(rpcClient jsonrpc.RPCClient, buyerID uint64) {
	args := localjsonrpc.RoutingRulesSettingsArgs{
		BuyerID: buyerID,
	}

	var reply localjsonrpc.RoutingRulesSettingsReply
	if err := rpcClient.CallFor(&reply, "OpsService.RoutingRulesSettings", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	table.Output(reply.RoutingRuleSettings)
}

func setRoutingRulesSettings(rpcClient jsonrpc.RPCClient, buyerID uint64, rrs routing.RoutingRulesSettings) {
	args := localjsonrpc.SetRoutingRulesSettingsArgs{
		BuyerID:              buyerID,
		RoutingRulesSettings: rrs,
	}

	var reply localjsonrpc.SetRouteShaderReply
	if err := rpcClient.CallFor(&reply, "OpsService.SetRoutingRulesSettings", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	fmt.Printf("Route shader for buyer with ID \"%d\" updated.\n", buyerID)
}
