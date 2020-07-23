package main

import (
	"fmt"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func sellers(rpcClient jsonrpc.RPCClient, env Environment) {
	args := localjsonrpc.SellersArgs{}

	var reply localjsonrpc.SellersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Sellers", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	sellers := []struct {
		Name            string
		ID              string
		IngressPriceUSD float64
		EgressPriceUSD  float64
	}{}

	for _, seller := range reply.Sellers {
		sellers = append(sellers, struct {
			Name            string
			ID              string
			IngressPriceUSD float64
			EgressPriceUSD  float64
		}{
			Name:            seller.Name,
			ID:              seller.ID,
			IngressPriceUSD: seller.IngressPriceNibblins.ToDollars(),
			EgressPriceUSD:  seller.EgressPriceNibblins.ToDollars(),
		})
	}

	table.Output(sellers)
}

func addSeller(rpcClient jsonrpc.RPCClient, env Environment, seller routing.Seller) {
	args := localjsonrpc.AddSellerArgs{
		Seller: seller,
	}

	var reply localjsonrpc.AddSellerReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddSeller", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Seller \"%s\" added to storage.\n", seller.Name)
}

func removeSeller(rpcClient jsonrpc.RPCClient, env Environment, id string) {
	args := localjsonrpc.RemoveSellerArgs{
		ID: id,
	}

	var reply localjsonrpc.RemoveSellerReply
	if err := rpcClient.CallFor(&reply, "OpsService.RemoveSeller", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Seller with ID \"%s\" removed from storage.\n", id)
}
