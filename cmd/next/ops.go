package main

import (
	"fmt"

	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

// without type generics there's no way to collapse this to one func
// func opsMRC(rpcClient jsonrpc.RPCClient,
// 	env Environment,
// 	relayRegex string,
// 	mrcUSD float64,
// ) {

// 	var relay routing.Relay
// 	var ok bool
// 	if relay, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
// 		// error msg printed by called function
// 		return
// 	}

// 	relay.MRC = routing.CentsToNibblins(mrcUSD * 100)

// 	args := localjsonrpc.RelayMetadataArgs{
// 		Relay: relay,
// 	}

// 	var reply localjsonrpc.RelaysReply
// 	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
// 		handleJSONRPCError(env, err)
// 		return
// 	}

// 	fmt.Println("MRC successfully updated.")
// 	return

// }

// func opsOverage(rpcClient jsonrpc.RPCClient,
// 	env Environment,
// 	relayRegex string,
// 	overageUSD float64,
// ) {
// 	var relay routing.Relay
// 	var ok bool
// 	if relay, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
// 		// error msg printed by called function
// 		return
// 	}

// 	relay.Overage = routing.CentsToNibblins(overageUSD * 100)

// 	args := localjsonrpc.RelayMetadataArgs{
// 		Relay: relay,
// 	}

// 	var reply localjsonrpc.RelaysReply
// 	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
// 		handleJSONRPCError(env, err)
// 		return
// 	}

// 	fmt.Println("Overage successfully updated.")
// 	return
// }

// func opsBWRule(rpcClient jsonrpc.RPCClient,
// 	env Environment,
// 	relayRegex string,
// 	rule string,
// ) {
// 	var relay routing.Relay
// 	var ok bool
// 	if relay, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
// 		// error msg printed by called function
// 		return
// 	}

// 	var bwRule routing.BandWidthRule
// 	switch rule {
// 	case "pool":
// 		bwRule = routing.BWRulePool
// 	case "burst":
// 		bwRule = routing.BWRuleBurst
// 	case "flat":
// 		bwRule = routing.BWRuleFlat
// 	case "none":
// 		bwRule = routing.BWRuleNone
// 	default:
// 		handleRunTimeError(fmt.Sprintln("Bandwidth rule must be one of pool, burst, flat or none."), 0)
// 	}
// 	relay.BWRule = bwRule

// 	args := localjsonrpc.RelayMetadataArgs{
// 		Relay: relay,
// 	}

// 	var reply localjsonrpc.RelaysReply
// 	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); !ok {
// 		handleJSONRPCError(env, err)
// 		return
// 	}

// 	fmt.Println("Bandwidth rule successfully updated.")
// 	return
// }

// func opsTerm(rpcClient jsonrpc.RPCClient,
// 	env Environment,
// 	relayRegex string,
// 	months int32,
// ) {
// 	var relay routing.Relay
// 	var ok bool
// 	if relay, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
// 		// error msg printed by called function
// 		return
// 	}

// 	relay.ContractTerm = months

// 	args := localjsonrpc.RelayMetadataArgs{
// 		Relay: relay,
// 	}

// 	var reply localjsonrpc.RelaysReply
// 	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); !ok {
// 		handleJSONRPCError(env, err)
// 		return
// 	}

// 	fmt.Println("Contract term successfully updated.")
// 	return

// }

// func opsStartDate(rpcClient jsonrpc.RPCClient,
// 	env Environment,
// 	relayRegex string,
// 	startDate string,
// ) {
// 	var relay routing.Relay
// 	var ok bool
// 	if relay, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
// 		// error msg printed by called function
// 		return
// 	}

// 	var err error
// 	relay.StartDate, err = time.Parse("January 2, 2006", startDate)
// 	if err != nil {
// 		handleRunTimeError(fmt.Sprintf("Could not parse `%s` - must be of the form 'January 2, 2006'\n", startDate), 0)
// 	}

// 	args := localjsonrpc.RelayMetadataArgs{
// 		Relay: relay,
// 	}

// 	var reply localjsonrpc.RelaysReply
// 	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
// 		handleJSONRPCError(env, err)
// 		return
// 	}

// 	fmt.Println("Start date successfully updated.")
// 	return
// }

// func opsEndDate(rpcClient jsonrpc.RPCClient,
// 	env Environment,
// 	relayRegex string,
// 	endDate string,
// ) {
// 	var relay routing.Relay
// 	var ok bool
// 	if relay, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
// 		// error msg printed by called function
// 		return
// 	}

// 	var err error
// 	relay.EndDate, err = time.Parse("January 2, 2006", endDate)
// 	if err != nil {
// 		handleRunTimeError(fmt.Sprintf("Could not parse `%s` - must be of the form 'January 2, 2006'\n", endDate), 0)
// 	}

// 	args := localjsonrpc.RelayMetadataArgs{
// 		Relay: relay,
// 	}

// 	var reply localjsonrpc.RelaysReply
// 	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
// 		handleJSONRPCError(env, err)
// 		return
// 	}

// 	fmt.Println("End date successfully updated.")
// 	return
// }

// func opsType(rpcClient jsonrpc.RPCClient,
// 	env Environment,
// 	relayRegex string,
// 	machineType string,
// ) {
// 	var relay routing.Relay
// 	var ok bool
// 	if relay, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
// 		// error msg printed by called function
// 		return
// 	}

// 	var serverType routing.MachineType
// 	switch machineType {
// 	case "bare":
// 		serverType = routing.BareMetal
// 	case "vm":
// 		serverType = routing.VirtualMachine
// 	case "none":
// 		serverType = routing.NoneSpecified
// 	default:
// 		handleRunTimeError(fmt.Sprintln("machine type must be one of bare, vm or none."), 0)
// 	}
// 	relay.Type = serverType

// 	args := localjsonrpc.RelayMetadataArgs{
// 		Relay: relay,
// 	}

// 	var reply localjsonrpc.RelaysReply
// 	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
// 		handleJSONRPCError(env, err)
// 		return
// 	}

// 	fmt.Println("Machine type successfully updated.")
// 	return
// }

// func opsBandwidth(rpcClient jsonrpc.RPCClient,
// 	env Environment,
// 	relayRegex string,
// 	bw int32,
// ) {
// 	var relay routing.Relay
// 	var ok bool
// 	if relay, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
// 		// error msg printed by called function
// 		return
// 	}

// 	relay.IncludedBandwidthGB = bw

// 	args := localjsonrpc.RelayMetadataArgs{
// 		Relay: relay,
// 	}

// 	var reply localjsonrpc.RelaysReply
// 	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); !ok {
// 		handleJSONRPCError(env, err)
// 		return
// 	}

// 	fmt.Println("Bandwidth allotment successfully updated.")
// 	return

// }

// func opsNic(rpcClient jsonrpc.RPCClient,
// 	env Environment,
// 	relayRegex string,
// 	nic int32,
// ) {
// 	var relay routing.Relay
// 	var ok bool
// 	if relay, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
// 		// error msg printed by called function
// 		return
// 	}

// 	relay.NICSpeedMbps = nic

// 	args := localjsonrpc.RelayMetadataArgs{
// 		Relay: relay,
// 	}

// 	var reply localjsonrpc.RelaysReply
// 	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); !ok {
// 		handleJSONRPCError(env, err)
// 		return
// 	}

// 	fmt.Println("NIC speed successfully updated.")
// 	return

// }

func checkForRelay(rpcClient jsonrpc.RPCClient, env Environment, regex string) (routing.Relay, bool) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	err := rpcClient.CallFor(&reply, "OpsService.Relays", args)
	if err != nil {
		handleJSONRPCError(env, err)
		return routing.Relay{}, false
	}

	if len(reply.Relays) == 0 {
		handleRunTimeError(fmt.Sprintf("Zero relays found matching '%s'", regex), 0)
	} else if len(reply.Relays) > 1 {
		handleRunTimeError(fmt.Sprintf("'%s' returned more than one relay - please be more specific.", regex), 0)
	}

	replyRelay := routing.Relay{
		ID:                  reply.Relays[0].ID,
		Name:                reply.Relays[0].Name,
		IncludedBandwidthGB: reply.Relays[0].IncludedBandwidthGB,
		NICSpeedMbps:        reply.Relays[0].NICSpeedMbps,
		FirestoreID:         reply.Relays[0].FirestoreID,
		MRC:                 reply.Relays[0].MRC,
		Overage:             reply.Relays[0].Overage,
		BWRule:              reply.Relays[0].BWRule,
		ContractTerm:        reply.Relays[0].ContractTerm,
		StartDate:           reply.Relays[0].StartDate,
		EndDate:             reply.Relays[0].EndDate,
		Type:                reply.Relays[0].Type,
	}

	return replyRelay, true

}
