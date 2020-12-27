package main

import (
	"fmt"

	"github.com/modood/table"
	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func customers(rpcClient jsonrpc.RPCClient, env Environment) {
	args := localjsonrpc.BuyersArgs{}

	var reply localjsonrpc.CustomersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Customers", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	table.Output(reply.Customers)
}

func addCustomer(rpcClient jsonrpc.RPCClient, env Environment, c routing.Customer) {

	arg := localjsonrpc.AddCustomerArgs{
		Customer: c,
	}

	var reply localjsonrpc.AddCustomerReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddCustomer", arg); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Customer added: %s\n", c.Name)

}

// func customerLink(rpcClient jsonrpc.RPCClient, env Environment, customerName string, buyerID uint64, sellerID string) {
// 	args := localjsonrpc.SetCustomerLinkArgs{
// 		CustomerName: customerName,
// 		BuyerID:      buyerID,
// 		SellerID:     sellerID,
// 	}

// 	var reply localjsonrpc.SetCustomerLinkReply
// 	if err := rpcClient.CallFor(&reply, "OpsService.SetCustomerLink", args); err != nil {
// 		handleJSONRPCError(env, err)
// 		return
// 	}

// 	if buyerID != 0 {
// 		fmt.Printf("Customer %s linked to buyer ID %d successfully\n", customerName, buyerID)
// 	}

// 	if sellerID != "" {
// 		fmt.Printf("Customer %s linked to seller ID %s successfully\n", customerName, sellerID)
// 	}
// }
