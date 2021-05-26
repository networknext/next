package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"

	"github.com/modood/table"
	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
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
				Name:    buyer.CompanyName,
				BuyerID: fmt.Sprintf("%d", int64(buyer.ID)),
			})
		}
	} else {
		for _, buyer := range reply.Buyers {
			buyers = append(buyers, struct {
				Name    string
				BuyerID string
			}{
				Name:    buyer.CompanyName,
				BuyerID: fmt.Sprintf("%016x", buyer.ID),
			})
		}
	}

	table.Output(buyers)
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

			fmt.Printf(" Routing rules for %s:\n\n", buyers.Buyers[i].CompanyName)

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
		if r.MatchString(buyer.CompanyName) {
			filtered = append(filtered, []string{buyer.CompanyName, fmt.Sprintf("%016x", buyer.ID)})
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

func datacenterMapsForBuyer(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyer string,
	csvOutput bool,
	signedIDs bool,
) {

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

		if signedIDs {
			var newMaps []localjsonrpc.DatacenterMapsFull
			for _, dcMap := range reply.DatacenterMaps {
				buyerID, err := strconv.ParseUint(dcMap.BuyerID, 16, 64)
				if err != nil {
					handleRunTimeError(fmt.Sprintf("Error converting BuyerID hex to signed int: %s\n", dcMap.BuyerID), 1)
				}
				dcID, err := strconv.ParseUint(dcMap.DatacenterID, 16, 64)
				if err != nil {
					handleRunTimeError(fmt.Sprintf("Error converting DatacenterID hex to signed int: %s\n", dcMap.DatacenterID), 1)
				}
				dcMap.BuyerID = fmt.Sprintf("%d", int64(buyerID))
				dcMap.DatacenterID = fmt.Sprintf("%d", int64(dcID))

				newMaps = append(newMaps, dcMap)
			}

			reply.DatacenterMaps = newMaps
		}

		if csvOutput {
			var csvInfo [][]string
			csvInfo = append(csvInfo, []string{
				"Alias", "DatacenterName", "DatacenterID", "BuyerName", "BuyerID", "SupplierName"})
			for _, dcMap := range reply.DatacenterMaps {

				csvInfo = append(csvInfo, []string{
					dcMap.Alias,
					dcMap.DatacenterName,
					dcMap.DatacenterID,
					dcMap.BuyerName,
					dcMap.BuyerID,
					dcMap.SupplierName,
				})
			}

			fileName := "./dcmaps.csv"
			f, err := os.Create(fileName)
			if err != nil {
				fmt.Printf("Error creating local CSV file %s: %v\n", fileName, err)
				return
			}

			writer := csv.NewWriter(f)
			err = writer.WriteAll(csvInfo)
			if err != nil {
				fmt.Printf("Error writing local CSV file %s: %v\n", fileName, err)
			}
			fmt.Println("CSV file written: dcmaps.csv")
		} else {
			table.Output(reply.DatacenterMaps)
		}

	} else {
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
			if r.MatchString(buyer.CompanyName) || r.MatchString(fmt.Sprintf("%016x", buyer.ID)) {
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

		if signedIDs {
			var newMaps []localjsonrpc.DatacenterMapsFull
			for _, dcMap := range reply.DatacenterMaps {
				buyerID, err := strconv.ParseUint(dcMap.BuyerID, 16, 64)
				if err != nil {
					handleRunTimeError(fmt.Sprintf("Error converting BuyerID hex to signed int: %s\n", dcMap.BuyerID), 1)
				}
				dcID, err := strconv.ParseUint(dcMap.DatacenterID, 16, 64)
				if err != nil {
					handleRunTimeError(fmt.Sprintf("Error converting DatacenterID hex to signed int: %s\n", dcMap.DatacenterID), 1)
				}
				dcMap.BuyerID = fmt.Sprintf("%d", int64(buyerID))
				dcMap.DatacenterID = fmt.Sprintf("%d", int64(dcID))

				newMaps = append(newMaps, dcMap)
			}

			reply.DatacenterMaps = newMaps
		}

		if csvOutput {
			var csvInfo [][]string
			csvInfo = append(csvInfo, []string{
				"Alias", "DatacenterName", "DatacenterID", "BuyerName", "BuyerID", "SupplierName"})
			for _, dcMap := range reply.DatacenterMaps {
				csvInfo = append(csvInfo, []string{
					dcMap.Alias,
					dcMap.DatacenterName,
					dcMap.DatacenterID,
					dcMap.BuyerName,
					dcMap.BuyerID,
					dcMap.SupplierName,
				})
			}

			fileName := "./dcmaps.csv"
			f, err := os.Create(fileName)
			if err != nil {
				fmt.Printf("Error creating local CSV file %s: %v\n", fileName, err)
				return
			}

			writer := csv.NewWriter(f)
			err = writer.WriteAll(csvInfo)
			if err != nil {
				fmt.Printf("Error writing local CSV file %s: %v\n", fileName, err)
			}
			fmt.Println("CSV file written: dcmaps.csv")
		} else {
			table.Output(reply.DatacenterMaps)
		}
	}

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
		if r.MatchString(buyer.CompanyName) || r.MatchString(fmt.Sprintf("%016x", buyer.ID)) {
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
			BuyerID:      buyerID,
			DatacenterID: dcID,
			Alias:        dcm.Alias,
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
		if r.MatchString(buyer.CompanyName) || r.MatchString(fmt.Sprintf("%016x", buyer.ID)) {
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
			BuyerID:      buyerID,
			DatacenterID: dcID,
			Alias:        dcm.Alias,
		},
	}

	var reply localjsonrpc.RemoveDatacenterMapReply
	if err := rpcClient.CallFor(&reply, "BuyersService.RemoveDatacenterMap", arg); err != nil {
		handleJSONRPCError(env, err)
		return nil
	}

	return nil

}

func buyerIDFromName(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerRegex string,
) (string, uint64) {

	buyerArgs := localjsonrpc.BuyersArgs{}
	var buyers localjsonrpc.BuyersReply
	if err := rpcClient.CallFor(&buyers, "OpsService.Buyers", buyerArgs); err != nil {
		handleJSONRPCError(env, err)
	}

	var filtered []string
	var buyerID uint64

	r := regexp.MustCompile("(?i)" + buyerRegex) // case-insensitive regex
	for _, buyer := range buyers.Buyers {
		if r.MatchString(buyer.CompanyName) {
			filtered = append(filtered, buyer.CompanyName)
			buyerID = buyer.ID
		}
	}

	if len(filtered) == 0 {
		handleRunTimeError(fmt.Sprintf("No buyer matches found for '%s'", buyerRegex), 0)
	}

	if len(filtered) > 1 {
		fmt.Printf("Found several  matches for '%s'", buyerRegex)
		for _, match := range filtered {
			fmt.Printf("\t%s\n", match)
		}
		handleRunTimeError(fmt.Sprintln("Please be more specific."), 0)
	}

	return filtered[0], buyerID
}

func getInternalConfig(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerRegex string,
) error {
	var reply localjsonrpc.InternalConfigReply

	buyerName, buyerID := buyerIDFromName(rpcClient, env, buyerRegex)

	buyerIDHex := fmt.Sprintf("%016x", buyerID)

	arg := localjsonrpc.InternalConfigArg{
		BuyerID: buyerIDHex,
	}

	if err := rpcClient.CallFor(&reply, "BuyersService.InternalConfig", arg); err != nil {
		handleJSONRPCError(env, err)
	}

	fmt.Printf("InternalConfig for buyer %s:\n", buyerName)
	fmt.Printf("  RouteSelectThreshold          : %d\n", reply.InternalConfig.RouteSelectThreshold)
	fmt.Printf("  RouteSwitchThreshold          : %d\n", reply.InternalConfig.RouteSwitchThreshold)
	fmt.Printf("  MaxLatencyTradeOff            : %d\n", reply.InternalConfig.MaxLatencyTradeOff)
	fmt.Printf("  RTTVeto_Default               : %d\n", reply.InternalConfig.RTTVeto_Default)
	fmt.Printf("  RTTVeto_PacketLoss            : %d\n", reply.InternalConfig.RTTVeto_PacketLoss)
	fmt.Printf("  RTTVeto_Multipath             : %d\n", reply.InternalConfig.RTTVeto_Multipath)
	fmt.Printf("  MultipathOverloadThreshold    : %d\n", reply.InternalConfig.MultipathOverloadThreshold)
	fmt.Printf("  TryBeforeYouBuy               : %t\n", reply.InternalConfig.TryBeforeYouBuy)
	fmt.Printf("  ForceNext                     : %t\n", reply.InternalConfig.ForceNext)
	fmt.Printf("  LargeCustomer                 : %t\n", reply.InternalConfig.LargeCustomer)
	fmt.Printf("  Uncommitted                   : %t\n", reply.InternalConfig.Uncommitted)
	fmt.Printf("  MaxRTT                        : %d\n", reply.InternalConfig.MaxRTT)
	fmt.Printf("  HighFrequencyPings            : %t\n", reply.InternalConfig.HighFrequencyPings)
	fmt.Printf("  RouteDiversity                : %d\n", reply.InternalConfig.RouteDiversity)
	fmt.Printf("  MultipathThreshold            : %d\n", reply.InternalConfig.MultipathThreshold)
	fmt.Printf("  EnableVanityMetrics           : %t\n", reply.InternalConfig.EnableVanityMetrics)
	fmt.Printf("  ReducePacketLossMinSliceNumber: %d\n", reply.InternalConfig.ReducePacketLossMinSliceNumber)

	return nil
}

func getRouteShader(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerRegex string,
) error {
	var reply localjsonrpc.RouteShaderReply

	buyerName, buyerID := buyerIDFromName(rpcClient, env, buyerRegex)

	buyerIDHex := fmt.Sprintf("%016x", buyerID)

	arg := localjsonrpc.RouteShaderArg{
		BuyerID: buyerIDHex,
	}
	if err := rpcClient.CallFor(&reply, "BuyersService.RouteShader", arg); err != nil {
		fmt.Println("No RouteShader stored for this buyer (they use the defaults).")
		return nil
	}

	fmt.Printf("RouteShader for buyer %s:\n", buyerName)
	fmt.Printf("  DisableNetworkNext       : %t\n", reply.RouteShader.DisableNetworkNext)
	fmt.Printf("  SelectionPercent         : %d\n", reply.RouteShader.SelectionPercent)
	fmt.Printf("  ABTest                   : %t\n", reply.RouteShader.ABTest)
	fmt.Printf("  ProMode                  : %t\n", reply.RouteShader.ProMode)
	fmt.Printf("  ReduceLatency            : %t\n", reply.RouteShader.ReduceLatency)
	fmt.Printf("  ReduceJitter             : %t\n", reply.RouteShader.ReduceJitter)
	fmt.Printf("  ReducePacketLoss         : %t\n", reply.RouteShader.ReducePacketLoss)
	fmt.Printf("  Multipath                : %t\n", reply.RouteShader.Multipath)
	fmt.Printf("  AcceptableLatency        : %d\n", reply.RouteShader.AcceptableLatency)
	fmt.Printf("  LatencyThreshold         : %d\n", reply.RouteShader.LatencyThreshold)
	fmt.Printf("  AcceptablePacketLoss     : %5.5f\n", reply.RouteShader.AcceptablePacketLoss)
	fmt.Printf("  BandwidthEnvelopeUpKbps  : %d\n", reply.RouteShader.BandwidthEnvelopeUpKbps)
	fmt.Printf("  BandwidthEnvelopeDownKbps: %d\n", reply.RouteShader.BandwidthEnvelopeDownKbps)
	fmt.Printf("  PacketLossSustained      : %.2f\n", reply.RouteShader.PacketLossSustained)

	return nil
}

func addInternalConfig(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerID uint64,
	ic localjsonrpc.JSInternalConfig,
) error {

	emptyReply := localjsonrpc.JSAddInternalConfigReply{}

	args := localjsonrpc.JSAddInternalConfigArgs{
		BuyerID:        fmt.Sprintf("%016x", buyerID),
		InternalConfig: ic,
	}
	// Storer method checks BuyerID validity
	if err := rpcClient.CallFor(&emptyReply, "BuyersService.JSAddInternalConfig", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Println("InternalConfig added successfully.")
	return nil
}

func removeInternalConfig(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerRegex string,
) error {

	buyerName, buyerID := buyerIDFromName(rpcClient, env, buyerRegex)

	emptyReply := localjsonrpc.RemoveInternalConfigReply{}

	args := localjsonrpc.RemoveInternalConfigArg{
		BuyerID: fmt.Sprintf("%016x", buyerID),
	}
	// Storer method checks BuyerID validity
	if err := rpcClient.CallFor(&emptyReply, "BuyersService.RemoveInternalConfig", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("InternalConfig for %s removed successfully.\n", buyerName)
	return nil
}

func updateInternalConfig(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerRegex string,
	field string,
	value string,
) error {

	buyerName, buyerID := buyerIDFromName(rpcClient, env, buyerRegex)

	emptyReply := localjsonrpc.UpdateInternalConfigReply{}

	args := localjsonrpc.UpdateInternalConfigArgs{
		BuyerID: buyerID,
		Field:   field,
		Value:   value,
	}
	if err := rpcClient.CallFor(&emptyReply, "BuyersService.UpdateInternalConfig", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("InternalConfig for %s updated successfully.\n", buyerName)
	return nil
}

func addRouteShader(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerID uint64,
	rs localjsonrpc.JSRouteShader,
) error {

	emptyReply := localjsonrpc.JSAddRouteShaderReply{}

	args := localjsonrpc.JSAddRouteShaderArgs{
		BuyerID:     fmt.Sprintf("%016x", buyerID),
		RouteShader: rs,
	}
	if err := rpcClient.CallFor(&emptyReply, "BuyersService.JSAddRouteShader", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Println("RouteShader added successfully.")
	return nil
}

func removeRouteShader(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerRegex string,
) error {

	buyerName, buyerID := buyerIDFromName(rpcClient, env, buyerRegex)

	emptyReply := localjsonrpc.RemoveRouteShaderReply{}

	args := localjsonrpc.RemoveRouteShaderArg{
		BuyerID: fmt.Sprintf("%016x", buyerID),
	}
	if err := rpcClient.CallFor(&emptyReply, "BuyersService.RemoveRouteShader", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("RouteShader for %s removed successfully.\n", buyerName)
	return nil
}

func updateRouteShader(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerRegex string,
	field string,
	value string,
) error {

	buyerName, buyerID := buyerIDFromName(rpcClient, env, buyerRegex)

	emptyReply := localjsonrpc.UpdateRouteShaderReply{}

	args := localjsonrpc.UpdateRouteShaderArgs{
		BuyerID: buyerID,
		Field:   field,
		Value:   value,
	}
	if err := rpcClient.CallFor(&emptyReply, "BuyersService.UpdateRouteShader", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("RouteShader for %s updated successfully.\n", buyerName)
	return nil
}

func getBannedUsers(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerRegex string,
) error {

	buyerName, buyerID := buyerIDFromName(rpcClient, env, buyerRegex)

	reply := localjsonrpc.GetBannedUserReply{}

	args := localjsonrpc.GetBannedUserArg{
		BuyerID: buyerID,
	}
	if err := rpcClient.CallFor(&reply, "BuyersService.GetBannedUsers", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("Banned users for buyer %s:\n", buyerName)
	for _, userID := range reply.BannedUsers {
		fmt.Printf("  %s\n", userID)
	}
	return nil
}

func addBannedUser(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerRegex string,
	userID uint64,
) error {

	buyerName, buyerID := buyerIDFromName(rpcClient, env, buyerRegex)

	emptyReply := localjsonrpc.BannedUserReply{}

	args := localjsonrpc.BannedUserArgs{
		BuyerID: buyerID,
		UserID:  userID,
	}
	if err := rpcClient.CallFor(&emptyReply, "BuyersService.AddBannedUser", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("Banned user %016x added for buyer %s successfully.\n", userID, buyerName)
	return nil
}

func removeBannedUser(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerRegex string,
	userID uint64,
) error {

	buyerName, buyerID := buyerIDFromName(rpcClient, env, buyerRegex)

	emptyReply := localjsonrpc.BannedUserReply{}

	args := localjsonrpc.BannedUserArgs{
		BuyerID: buyerID,
		UserID:  userID,
	}
	if err := rpcClient.CallFor(&emptyReply, "BuyersService.RemoveBannedUser", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("Banned user %016x successfully removed  for buyer %s.\n", userID, buyerName)
	return nil
}

func getBuyerInfo(rpcClient jsonrpc.RPCClient, env Environment, buyerRegex string) {

	_, buyerID := buyerIDFromName(rpcClient, env, buyerRegex)

	arg := localjsonrpc.BuyerArg{
		BuyerID: buyerID,
	}

	var reply localjsonrpc.BuyerReply
	if err := rpcClient.CallFor(&reply, "BuyersService.Buyer", arg); err != nil {
		handleJSONRPCError(env, err)
	}

	buyerInfo := "Buyer " + reply.Buyer.ShortName + " info:\n"
	buyerInfo += "  CompanyCode: " + reply.Buyer.CompanyCode + "\n"
	buyerInfo += "  ShortName  : " + reply.Buyer.ShortName + "\n"
	buyerInfo += "  Live       : " + fmt.Sprintf("%t", reply.Buyer.Live) + "\n"
	buyerInfo += "  Debug      : " + fmt.Sprintf("%t", reply.Buyer.Debug) + "\n"
	buyerInfo += "  ID         : " + fmt.Sprintf("%016x", uint64(reply.Buyer.ID)) + "\n"
	buyerInfo += "  Public Key : " + reply.Buyer.EncodedPublicKey() + "\n"

	fmt.Println(buyerInfo)
	os.Exit(0)

}

func updateBuyer(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	buyerRegex string,
	field string,
	value string,
) error {

	buyerName, buyerID := buyerIDFromName(rpcClient, env, buyerRegex)

	emptyReply := localjsonrpc.UpdateBuyerReply{}

	args := localjsonrpc.UpdateBuyerArgs{
		BuyerID: buyerID,
		Field:   field,
		Value:   value,
	}
	if err := rpcClient.CallFor(&emptyReply, "BuyersService.UpdateBuyer", args); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("Buyer %s updated successfully.\n", buyerName)
	return nil
}
