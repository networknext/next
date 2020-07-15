package main

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func buyers(rpcClient jsonrpc.RPCClient, env Environment) {
	args := localjsonrpc.BuyersArgs{}

	var reply localjsonrpc.BuyersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Buyers", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	sort.Slice(reply.Buyers, func(i int, j int) bool {
		return reply.Buyers[i].ID > reply.Buyers[j].ID
	})

	buyers := []struct {
		Name    string
		BuyerID string
	}{}

	for _, buyer := range reply.Buyers {
		buyers = append(buyers, struct {
			Name    string
			BuyerID string
		}{
			Name:    buyer.Name,
			BuyerID: buyer.ID,
		})
	}

	table.Output(buyers)
}

func addBuyer(rpcClient jsonrpc.RPCClient, env Environment, buyer routing.Buyer) {
	args := localjsonrpc.AddBuyerArgs{
		Buyer: buyer,
	}

	var reply localjsonrpc.AddBuyerReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddBuyer", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Buyer \"%s\" added to storage.\n", buyer.Name)
}

func removeBuyer(rpcClient jsonrpc.RPCClient, env Environment, id string) {
	args := localjsonrpc.RemoveBuyerArgs{
		ID: id,
	}

	var reply localjsonrpc.RemoveBuyerReply
	if err := rpcClient.CallFor(&reply, "OpsService.RemoveBuyer", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Buyer with ID \"%s\" removed from storage.\n", id)
}

func routingRulesSettingsByID(rpcClient jsonrpc.RPCClient, env Environment, buyerID string) {

	buyerArgs := localjsonrpc.BuyersArgs{}
	var buyers localjsonrpc.BuyersReply
	if err := rpcClient.CallFor(&buyers, "OpsService.Buyers", buyerArgs); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	for i := range buyers.Buyers {
		if buyers.Buyers[i].ID == buyerID {

			fmt.Printf(" Routing rules for %s:\n\n", buyers.Buyers[i].Name)

			transpose := []struct {
				RoutingRuleSetting string
				Value              string
			}{}

			args := localjsonrpc.RoutingRulesSettingsArgs{
				BuyerID: buyerID,
			}

			var reply localjsonrpc.RoutingRulesSettingsReply
			if err := rpcClient.CallFor(&reply, "OpsService.RoutingRulesSettings", args); err != nil {
				handleJSONRPCError(env, err)
				return
			}

			v := reflect.ValueOf(reply.RoutingRuleSettings[0])
			typeOfV := v.Type()

			for i := 0; i < v.NumField(); i++ {
				transpose = append(transpose, struct {
					RoutingRuleSetting string
					Value              string
				}{
					RoutingRuleSetting: typeOfV.Field(i).Name,
					Value:              fmt.Sprintf("%v", v.Field(i).Interface()),
				})
			}

			table.Output(transpose)
			return
		}
	}

	fmt.Printf("Buyer id %s not found", buyerID)

}

func routingRulesSettings(rpcClient jsonrpc.RPCClient, env Environment, buyerName string) {

	buyerArgs := localjsonrpc.BuyersArgs{}
	var buyers localjsonrpc.BuyersReply
	if err := rpcClient.CallFor(&buyers, "OpsService.Buyers", buyerArgs); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	var filtered [][]string

	r := regexp.MustCompile("(?i)" + buyerName) // case-insensitive regex
	for _, buyer := range buyers.Buyers {
		if r.MatchString(buyer.Name) {
			filtered = append(filtered, []string{buyer.Name, buyer.ID})
		}
	}

	if len(filtered) == 0 {
		fmt.Printf("No matches found for '%s'", buyerName)
		return
	}

	if len(filtered) > 1 {
		fmt.Printf("Found several  matches for '%s'", buyerName)
		for _, match := range filtered {
			fmt.Printf("\t%s", match[0])
		}
		return
	}

	fmt.Printf(" Routing rules for %s:\n\n", filtered[0][0])

	transpose := []struct {
		RoutingRuleSetting string
		Value              string
	}{}

	args := localjsonrpc.RoutingRulesSettingsArgs{
		BuyerID: filtered[0][1],
	}

	var reply localjsonrpc.RoutingRulesSettingsReply
	if err := rpcClient.CallFor(&reply, "OpsService.RoutingRulesSettings", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	v := reflect.ValueOf(reply.RoutingRuleSettings[0])
	typeOfV := v.Type()

	for i := 0; i < v.NumField(); i++ {
		transpose = append(transpose, struct {
			RoutingRuleSetting string
			Value              string
		}{
			RoutingRuleSetting: typeOfV.Field(i).Name,
			Value:              fmt.Sprintf("%v", v.Field(i).Interface()),
		})
	}

	table.Output(transpose)
	return
}

func setRoutingRulesSettings(rpcClient jsonrpc.RPCClient, env Environment, buyerID string, rrs routing.RoutingRulesSettings) {
	args := localjsonrpc.SetRoutingRulesSettingsArgs{
		BuyerID:              buyerID,
		RoutingRulesSettings: rrs,
	}

	var reply localjsonrpc.SetRoutingRulesSettingsReply
	if err := rpcClient.CallFor(&reply, "OpsService.SetRoutingRulesSettings", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Route shader for buyer with ID \"%s\" updated.\n", buyerID)
}

func datacenterMapsForBuyer(rpcClient jsonrpc.RPCClient, env Environment, arg uint64) {

	type dcMapStrings struct {
		BuyerID    string `json:"buyer_id"`
		Datacenter string `json:"datacenter"`
		Alias      string `json:"alias"`
	}

	var buyerID uint64
	if buyerID = returnBuyerID(rpcClient, env, arg); buyerID == 0 {
		fmt.Printf("No matches found for '%x'", arg)
		return
	}
	args := localjsonrpc.DatacenterMapsArgs{
		ID: buyerID,
	}

	var reply localjsonrpc.DatacenterMapsReply
	if err := rpcClient.CallFor(&reply, "BuyersService.DatacenterMapsForBuyer", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	var list []dcMapStrings
	for _, dc := range reply.DatacenterMaps {
		list = append(list, dcMapStrings{
			Alias:      dc.Alias,
			BuyerID:    fmt.Sprintf("%016x", dc.BuyerID),
			Datacenter: fmt.Sprintf("%016x", dc.Datacenter),
		})
	}

	table.Output(list)

}

func addDatacenterMap(rpcClient jsonrpc.RPCClient, env Environment, dcm routing.DatacenterMap) error {

	var buyerID uint64
	if buyerID = returnBuyerID(rpcClient, env, dcm.BuyerID); buyerID == 0 {
		return fmt.Errorf("No buyer matches found for '%x'", dcm.BuyerID)
	}
	var dcID uint64
	if dcID = returnDatacenterID(rpcClient, env, dcm.Datacenter); dcID == 0 {
		return fmt.Errorf("No matches found for '%x'", dcm.Datacenter)
	}
	arg := localjsonrpc.AddDatacenterMapArgs{
		DatacenterMap: routing.DatacenterMap{
			BuyerID:    buyerID,
			Datacenter: dcID,
			Alias:      dcm.Alias,
		},
	}

	var reply localjsonrpc.AddDatacenterMapReply
	if err := rpcClient.CallFor(&reply, "BuyersService.AddDatacenterMap", arg); err != nil {
		handleJSONRPCError(env, err)
		return fmt.Errorf("AddDatacenterMap error: %v\n", err)
	}

	return nil

}

func removeDatacenterMap(rpcClient jsonrpc.RPCClient, env Environment, dcm routing.DatacenterMap) error {
	var buyerID uint64
	if buyerID = returnBuyerID(rpcClient, env, dcm.BuyerID); buyerID == 0 {
		return fmt.Errorf("No buyer matches found for '%x'", dcm.BuyerID)
	}
	var dcID uint64
	if dcID = returnDatacenterID(rpcClient, env, dcm.Datacenter); dcID == 0 {
		return fmt.Errorf("No matches found for '%x'", dcm.Datacenter)
	}

	arg := localjsonrpc.RemoveDatacenterMapArgs{
		DatacenterMap: routing.DatacenterMap{
			BuyerID:    buyerID,
			Datacenter: dcID,
			Alias:      dcm.Alias,
		},
	}

	var reply localjsonrpc.RemoveDatacenterMapReply
	if err := rpcClient.CallFor(&reply, "BuyersService.RemoveDatacenterMap", arg); err != nil {
		handleJSONRPCError(env, err)
		return err
	}

	return nil

}

func returnBuyerID(rpcClient jsonrpc.RPCClient, env Environment, arg uint64) uint64 {

	// check to see if user entered name or substring (not id)
	buyerArgs := localjsonrpc.BuyersArgs{}
	var buyers localjsonrpc.BuyersReply
	if err := rpcClient.CallFor(&buyers, "OpsService.Buyers", buyerArgs); err != nil {
		handleJSONRPCError(env, err)
		return 0
	}

	var buyerID uint64

	r := regexp.MustCompile("(?i)" + fmt.Sprintf("%x", arg)) // case-insensitive regex
	for _, buyer := range buyers.Buyers {
		hexBuyerID, err := strconv.ParseUint(buyer.ID, 16, 64)
		if err != nil {
			handleJSONRPCError(env, err)
			return 0
		}
		if r.MatchString(buyer.Name) || arg == hexBuyerID {
			buyerID = arg
		}
	}

	return buyerID
}

func returnDatacenterID(rpcClient jsonrpc.RPCClient, env Environment, arg uint64) uint64 {

	args := localjsonrpc.DatacentersArgs{}

	var reply localjsonrpc.DatacentersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Datacenters", args); err != nil {
		handleJSONRPCError(env, err)
		return 0
	}

	var datacenterID uint64
	r := regexp.MustCompile("(?i)" + fmt.Sprintf("%x", arg)) // case-insensitive regex
	for _, dc := range reply.Datacenters {
		hexDatacenterID, err := strconv.ParseUint(dc.ID, 16, 64)
		if err != nil {
			handleJSONRPCError(env, err)
			return 0
		}
		if r.MatchString(dc.Name) || arg == hexDatacenterID {
			datacenterID = arg
		}
	}

	return datacenterID

}
