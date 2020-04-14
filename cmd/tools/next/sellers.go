package main

import (
	"fmt"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func sellers(rpcClient jsonrpc.RPCClient) {
	args := localjsonrpc.SellersArgs{}

	var reply localjsonrpc.SellersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Sellers", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	table.Output(reply.Sellers)
}

func addSeller(rpcClient jsonrpc.RPCClient, seller routing.Seller) {
	args := localjsonrpc.AddSellerArgs{
		Seller: seller,
	}

	var reply localjsonrpc.AddSellerReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddSeller", args); err != nil {
		handleJSONRPCError(err)
		return
	}

	fmt.Printf("Seller \"%s\" added to storage.\n", seller.Name)
}
