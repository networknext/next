package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/modood/table"
	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
)

func sellers(env Environment) {
	args := localjsonrpc.SellersArgs{}

	var reply localjsonrpc.SellersReply
	if err := makeRPCCall(env, &reply, "OpsService.Sellers", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	sellers := []struct {
		Name           string
		ID             string
		EgressPriceUSD string
		Secret         string
	}{}

	for _, seller := range reply.Sellers {
		sellers = append(sellers, struct {
			Name           string
			ID             string
			EgressPriceUSD string
			Secret         string
		}{
			Name:           seller.Name,
			ID:             seller.ID,
			EgressPriceUSD: fmt.Sprintf("$%02.2f", seller.EgressPriceNibblins.ToDollars()),
			Secret:         fmt.Sprintf("%t", seller.Secret),
		})
	}

	sort.Slice(sellers, func(i int, j int) bool {
		return sellers[i].ID < sellers[j].ID
	})

	table.Output(sellers)
}

func addSeller(env Environment, seller routing.Seller) {
	args := localjsonrpc.AddSellerArgs{
		Seller: seller,
	}

	var reply localjsonrpc.AddSellerReply
	if err := makeRPCCall(env, &reply, "OpsService.AddSeller", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Seller \"%s\" added to storage.\n", seller.Name)
}

func removeSeller(env Environment, id string) {
	args := localjsonrpc.RemoveSellerArgs{
		ID: id,
	}

	var reply localjsonrpc.RemoveSellerReply
	if err := makeRPCCall(env, &reply, "OpsService.RemoveSeller", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Seller with ID \"%s\" removed from storage.\n", id)
}

func getSellerInfo(env Environment, id string) {

	arg := localjsonrpc.SellerArg{
		ID: id,
	}

	var reply localjsonrpc.SellerReply
	if err := makeRPCCall(env, &reply, "OpsService.Seller", arg); err != nil {
		handleJSONRPCError(env, err)
	}

	sellerInfo := "Seller " + reply.Seller.Name + " info:\n"
	sellerInfo += "  ID           : " + reply.Seller.ID + "\n"
	sellerInfo += "  Name         : " + reply.Seller.Name + "\n"
	sellerInfo += "  ShortName    : " + reply.Seller.ShortName + "\n"
	sellerInfo += "  Egress Price : " + fmt.Sprintf("%4.2f", reply.Seller.EgressPriceNibblinsPerGB.ToDollars()) + "\n"
	sellerInfo += "  Secret       : " + fmt.Sprintf("%t", reply.Seller.Secret) + "\n"

	fmt.Println(sellerInfo)
	os.Exit(0)

}

func updateSeller(
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
	if err := makeRPCCall(env, &emptyReply, "OpsService.UpdateSeller", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("Seller %s updated successfully.\n", sellerID)
	return nil
}
