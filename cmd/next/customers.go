package main

import (
	"fmt"
	"os"
	"sort"

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

	sort.Slice(reply.Customers, func(i int, j int) bool {
		return reply.Customers[i].Code < reply.Customers[j].Code
	})

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

func updateCustomer(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	customerCode string,
	field string,
	value string,
) error {

	emptyReply := localjsonrpc.UpdateCustomerReply{}

	args := localjsonrpc.UpdateCustomerArgs{
		CustomerID: customerCode,
		Field:      field,
		Value:      value,
	}
	if err := rpcClient.CallFor(&emptyReply, "OpsService.UpdateCustomer", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("Customer %s updated successfully.\n", customerCode)
	return nil
}
