package main

import (
	"fmt"
	"os"

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

func getCustomerInfo(rpcClient jsonrpc.RPCClient, env Environment, id string) {

	arg := localjsonrpc.CustomerArg{
		CustomerID: id,
	}

	var reply localjsonrpc.CustomerReply
	if err := rpcClient.CallFor(&reply, "OpsService.Customer", arg); err != nil {
		handleJSONRPCError(env, err)
	}

	customerInfo := "Customer " + reply.Customer.Name + " info:\n"
	customerInfo += "  Code         : " + reply.Customer.Code + "\n"
	customerInfo += "  Name         : " + reply.Customer.Name + "\n\n"
	customerInfo += "  Automatic Sign-In Domains:\n"
	if reply.Customer.AutomaticSignInDomains == "" {
		customerInfo += "\tnone"
	} else {
		customerInfo += "\t" + reply.Customer.AutomaticSignInDomains + "\n"
	}

	fmt.Println(customerInfo)
	os.Exit(0)

}
