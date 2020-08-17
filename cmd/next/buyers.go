package main

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func buyers(rpcClient jsonrpc.RPCClient, env Environment, signed bool) {
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

	if signed {
		for _, buyer := range reply.Buyers {
			buyers = append(buyers, struct {
				Name    string
				BuyerID string
			}{
				Name:    buyer.Name,
				BuyerID: fmt.Sprintf("%d", int64(buyer.ID)),
			})
		}
	} else {
		for _, buyer := range reply.Buyers {
			buyers = append(buyers, struct {
				Name    string
				BuyerID string
			}{
				Name:    buyer.Name,
				BuyerID: fmt.Sprintf("%016x", buyer.ID),
			})
		}
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
		if fmt.Sprintf("%016x", buyers.Buyers[i].ID) == buyerID {

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
			filtered = append(filtered, []string{buyer.Name, fmt.Sprintf("%016x", buyer.ID)})
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

func datacenterMapsForBuyer(rpcClient jsonrpc.RPCClient, env Environment, buyer string) {

	if buyer == "" {
		var reply localjsonrpc.ListDatacenterMapsReply
		var arg = localjsonrpc.ListDatacenterMapsArgs{
			DatacenterID: 0,
		}

		if err := rpcClient.CallFor(&reply, "OpsService.ListDatacenterMaps", arg); err != nil {
			fmt.Printf("rpc error: %v\n", err)
			handleJSONRPCError(env, err)
			return
		}

		table.Output(reply.DatacenterMaps)
		return
	}

	var buyerID uint64
	var err error

	buyerArgs := localjsonrpc.BuyersArgs{}
	var buyersReply localjsonrpc.BuyersReply
	if err = rpcClient.CallFor(&buyersReply, "OpsService.Buyers", buyerArgs); err != nil {
		handleJSONRPCError(env, err)
		return
	}
	r := regexp.MustCompile("(?i)" + buyer) // case-insensitive regex
	for _, buyer := range buyersReply.Buyers {
		if r.MatchString(buyer.Name) || r.MatchString(fmt.Sprintf("%016x", buyer.ID)) {
			buyerID = buyer.ID
		}
	}

	if buyerID == 0 {
		fmt.Printf("No match for provided buyer ID: %v\n\n", buyer)
		fmt.Println("Here is a current list of buyers in the system:")
		buyers(rpcClient, env, false)
		return
	}

	args := localjsonrpc.DatacenterMapsArgs{
		ID: buyerID,
	}

	var reply localjsonrpc.DatacenterMapsReply
	if err := rpcClient.CallFor(&reply, "BuyersService.DatacenterMapsForBuyer", args); err != nil {
		fmt.Printf("rpc error: %v\n", err)
		handleJSONRPCError(env, err)
		return
	}

	sort.Slice(reply.DatacenterMaps, func(i int, j int) bool {
		return reply.DatacenterMaps[i].DatacenterName < reply.DatacenterMaps[j].DatacenterName
	})

	table.Output(reply.DatacenterMaps)

}

func addDatacenterMap(rpcClient jsonrpc.RPCClient, env Environment, dcm dcMapStrings) error {

	var err error
	var buyerID uint64
	var dcID uint64

	buyerArgs := localjsonrpc.BuyersArgs{}
	var buyers localjsonrpc.BuyersReply
	if err = rpcClient.CallFor(&buyers, "OpsService.Buyers", buyerArgs); err != nil {
		handleRunTimeError(fmt.Sprintln("Unable to retrive buyer list."), 1)
	}
	r := regexp.MustCompile("(?i)" + dcm.BuyerID) // case-insensitive regex
	for _, buyer := range buyers.Buyers {
		if r.MatchString(buyer.Name) || r.MatchString(fmt.Sprintf("%016x", buyer.ID)) {
			buyerID = buyer.ID
		}
	}
	if buyerID == 0 {
		handleRunTimeError(fmt.Sprintf("Buyer %s does not seem to exist.\n", dcm.BuyerID), 0)
	}

	dcArgs := localjsonrpc.DatacentersArgs{}
	var dcReply localjsonrpc.DatacentersReply
	if err = rpcClient.CallFor(&dcReply, "OpsService.Datacenters", dcArgs); err != nil {
		handleRunTimeError(fmt.Sprintln("Unable to retrive datacenter list."), 1)
	}
	r = regexp.MustCompile("(?i)" + dcm.Datacenter) // case-insensitive regex
	for _, dc := range dcReply.Datacenters {
		if r.MatchString(dc.Name) || r.MatchString(fmt.Sprintf("%016x", dc.ID)) {
			dcID = dc.ID
		}
	}
	if dcID == 0 {
		handleRunTimeError(fmt.Sprintf("Datacenter %s does not seem to exist.\n", dcm.Datacenter), 0)
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
		return nil
	}

	return nil

}

func removeDatacenterMap(rpcClient jsonrpc.RPCClient, env Environment, dcm dcMapStrings) error {

	var err error
	var buyerID uint64
	var dcID uint64

	buyerArgs := localjsonrpc.BuyersArgs{}
	var buyers localjsonrpc.BuyersReply
	if err = rpcClient.CallFor(&buyers, "OpsService.Buyers", buyerArgs); err != nil {
		handleRunTimeError(fmt.Sprintln("Unable to retrive buyer list."), 1)
	}
	r := regexp.MustCompile("(?i)" + dcm.BuyerID) // case-insensitive regex
	for _, buyer := range buyers.Buyers {
		if r.MatchString(buyer.Name) || r.MatchString(fmt.Sprintf("%016x", buyer.ID)) {
			buyerID = buyer.ID
		}
	}
	if buyerID == 0 {
		handleRunTimeError(fmt.Sprintf("Buyer %s does not seem to exist.\n", dcm.BuyerID), 0)
	}

	dcArgs := localjsonrpc.DatacentersArgs{}
	var dcReply localjsonrpc.DatacentersReply
	if err = rpcClient.CallFor(&dcReply, "OpsService.Datacenters", dcArgs); err != nil {
		handleRunTimeError(fmt.Sprintln("Unable to retrive datacenter list."), 1)
	}
	r = regexp.MustCompile("(?i)" + dcm.Datacenter) // case-insensitive regex
	for _, dc := range dcReply.Datacenters {
		if r.MatchString(dc.Name) || r.MatchString(fmt.Sprintf("%016x", dc.ID)) {
			dcID = dc.ID
		}
	}
	if dcID == 0 {
		handleRunTimeError(fmt.Sprintf("Datacenter %s does not seem to exist.\n", dcm.Datacenter), 0)
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
		return nil
	}

	return nil

}
