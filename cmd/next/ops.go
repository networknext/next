package main

import (
	"fmt"
	"time"

	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

// without type generics there's no way to collapse this to one func
func opsMRC(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	mrcUSD float64,
) {

	var relay routing.Relay
	var err error
	if relay, err = checkForRelay(rpcClient, env, relayRegex); err != nil {
		// error msg printed by called function
		return
	}

	relay.MRC = routing.CentsToNibblins(mrcUSD * 100)

	args := localjsonrpc.RelayMetadataArgs{
		Relay: relay,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("MRC successfully updated.")
	return

}

func opsOverage(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	overageUSD float64,
) {
	var relay routing.Relay
	var err error
	if relay, err = checkForRelay(rpcClient, env, relayRegex); err != nil {
		// error msg printed by called function
		return
	}

	relay.Overage = routing.CentsToNibblins(overageUSD * 100)

	args := localjsonrpc.RelayMetadataArgs{
		Relay: relay,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Overage successfully updated.")
	return
}

func opsBWRule(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	rule string,
) {
	var relay routing.Relay
	var err error
	if relay, err = checkForRelay(rpcClient, env, relayRegex); err != nil {
		// error msg printed by called function
		return
	}

	var bwRule routing.BandWidthRule
	switch rule {
	case "pool":
		bwRule = routing.BWRulePool
	case "burst":
		bwRule = routing.BWRuleBurst
	case "flat":
		bwRule = routing.BWRuleFlat
	case "none":
		bwRule = routing.BWRuleNone
	default:
		fmt.Println("Bandwidth rule must be one of pool, burst, flat or none.")
		return
	}
	relay.BWRule = bwRule

	args := localjsonrpc.RelayMetadataArgs{
		Relay: relay,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Bandwidth rule successfully updated.")
	return
}

func opsTerm(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	months uint32,
) {
	var relay routing.Relay
	var err error
	if relay, err = checkForRelay(rpcClient, env, relayRegex); err != nil {
		// error msg printed by called function
		return
	}

	relay.ContractTerm = months

	args := localjsonrpc.RelayMetadataArgs{
		Relay: relay,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Contract term successfully updated.")
	return

}

func opsStartDate(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	startDate string,
) {
	var relay routing.Relay
	var err error
	if relay, err = checkForRelay(rpcClient, env, relayRegex); err != nil {
		// error msg printed by called function
		return
	}

	relay.StartDate, err = time.Parse("January 2, 2006", startDate)
	if err != nil {
		fmt.Printf("Could not parse `%s` - must be of the form 'January 2, 2006'\n", startDate)
	}

	args := localjsonrpc.RelayMetadataArgs{
		Relay: relay,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Start date successfully updated.")
	return
}

func opsEndDate(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	endDate string,
) {
	var relay routing.Relay
	var err error
	if relay, err = checkForRelay(rpcClient, env, relayRegex); err != nil {
		// error msg printed by called function
		return
	}

	relay.EndDate, err = time.Parse("January 2, 2006", endDate)
	if err != nil {
		fmt.Printf("Could not parse `%s` - must be of the form 'January 2, 2006'\n", endDate)
	}

	args := localjsonrpc.RelayMetadataArgs{
		Relay: relay,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("End date successfully updated.")
	return
}

func opsType(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	machineType string,
) {
	var relay routing.Relay
	var err error
	if relay, err = checkForRelay(rpcClient, env, relayRegex); err != nil {
		// error msg printed by called function
		return
	}

	var serverType routing.MachineType
	switch machineType {
	case "bare":
		serverType = routing.BareMetal
	case "vm":
		serverType = routing.VirtualMachine
	case "none":
		serverType = routing.NoneSpecified
	default:
		fmt.Println("Bandwidth rule must be one of bare, vm or none.")
		return
	}
	relay.Type = serverType

	args := localjsonrpc.RelayMetadataArgs{
		Relay: relay,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.RelayMetadata", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Bandwidth rule successfully updated.")
	return
}

func opsModify(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	jsonFile string,
) {
	fmt.Println("Modify not yet implemented.")
}

func checkForRelay(rpcClient jsonrpc.RPCClient, env Environment, regex string) (routing.Relay, error) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	err := rpcClient.CallFor(&reply, "OpsService.Relays", args)
	if err != nil {
		handleJSONRPCError(env, err)
		return routing.Relay{}, err
	}

	if len(reply.Relays) == 0 {
		fmt.Printf("Zero relays found matching '%s'", regex)
		return routing.Relay{}, err
	} else if len(reply.Relays) > 1 {
		fmt.Printf("'%s' returned more than one relay - please be more specific.", regex)
		return routing.Relay{}, err
	}

	replyRelay := routing.Relay{
		ID:           reply.Relays[0].ID,
		Name:         reply.Relays[0].Name,
		FirestoreID:  reply.Relays[0].FirestoreID,
		MRC:          reply.Relays[0].MRC,
		Overage:      reply.Relays[0].Overage,
		BWRule:       reply.Relays[0].BWRule,
		ContractTerm: reply.Relays[0].ContractTerm,
		StartDate:    reply.Relays[0].StartDate,
		EndDate:      reply.Relays[0].EndDate,
		Type:         reply.Relays[0].Type,
	}

	return replyRelay, nil

}
