package main

import (
	"github.com/modood/table"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func customers(rpcClient jsonrpc.RPCClient, env Environment) {
	args := localjsonrpc.BuyersArgs{}

	var reply localjsonrpc.CustomersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Customers", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	customers := []struct {
		Name     string
		BuyerID  uint64
		SellerID string
	}{}

	for _, customer := range reply.Customers {
		customers = append(customers, struct {
			Name     string
			BuyerID  uint64
			SellerID string
		}{
			Name:     customer.Name,
			BuyerID:  customer.BuyerID,
			SellerID: customer.SellerID,
		})
	}

	table.Output(customers)
}
