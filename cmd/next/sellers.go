package main

import (
	"fmt"
	"os"

	"github.com/modood/table"
	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
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
		IngressPriceUSD string
		EgressPriceUSD  string
	}{}

	for _, seller := range reply.Sellers {
		sellers = append(sellers, struct {
			Name            string
			ID              string
			IngressPriceUSD string
			EgressPriceUSD  string
		}{
			Name:            seller.Name,
			ID:              seller.ID,
			IngressPriceUSD: fmt.Sprintf("$%02.2f", seller.IngressPriceNibblins.ToDollars()),
			EgressPriceUSD:  fmt.Sprintf("$%02.2f", seller.EgressPriceNibblins.ToDollars()),
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

func getSellerInfo(rpcClient jsonrpc.RPCClient, env Environment, id string) {

	arg := localjsonrpc.SellerArg{
		ID: id,
	}

	var reply localjsonrpc.SellerReply
	if err := rpcClient.CallFor(&reply, "OpsService.Seller", arg); err != nil {
		handleJSONRPCError(env, err)
	}

	sellerInfo := "Seller " + reply.Seller.Name + " info:\n"
	sellerInfo += "  ID           : " + reply.Seller.ID + "\n"
	sellerInfo += "  Name         : " + reply.Seller.Name + "\n"
	sellerInfo += "  ShortName    : " + reply.Seller.ShortName + "\n"
	sellerInfo += "  Egress Price : " + fmt.Sprintf("%4.2f", reply.Seller.EgressPriceNibblinsPerGB.ToDollars()) + "\n"
	sellerInfo += "  Ingress Price: " + fmt.Sprintf("%4.2f", reply.Seller.IngressPriceNibblinsPerGB.ToDollars()) + "\n"

	fmt.Println(sellerInfo)
	os.Exit(0)

}

func updateSeller(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	sellerID string,
	field string,
	value string,
) error {

	emptyReply := localjsonrpc.UpdateSellerReply{}

	args := localjsonrpc.UpdateSellerArgs{
		SellerID: sellerID,
		Field:    field,
		Value:    value,
	}
	if err := rpcClient.CallFor(&emptyReply, "OpsService.UpdateSeller", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("Seller %s updated successfully.\n", sellerID)
	return nil
}
