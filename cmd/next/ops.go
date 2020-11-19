package main

import (
	"fmt"

	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

// without type generics there's no way to collapse this to one func
func opsMRC(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	mrcUSD float64,
) {

	var ok bool
	var relayID uint64
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "MRC",
		Value:   mrcUSD,
	}

	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); err != nil {
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
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "Overage",
		Value:   overageUSD,
	}

	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); err != nil {
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
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	var bwRule int32
	switch rule {
	case "pool":
		bwRule = 3
	case "burst":
		bwRule = 2
	case "flat":
		bwRule = 1
	case "none":
		bwRule = 0
	default:
		handleRunTimeError(fmt.Sprintln("Bandwidth rule must be one of pool, burst, flat or none."), 0)
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "BWRule",
		Value:   bwRule,
	}
	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); !ok {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Bandwidth rule successfully updated.")
	return
}

func opsTerm(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	months int32,
) {
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "ContractTerm",
		Value:   months,
	}
	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); !ok {
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
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "StartDate",
		Value:   startDate,
	}
	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); err != nil {
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
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "EndDate",
		Value:   endDate,
	}

	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); err != nil {
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
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	var serverType int32
	switch machineType {
	case "bare":
		serverType = 1
	case "vm":
		serverType = 2
	case "none":
		serverType = 0
	default:
		handleRunTimeError(fmt.Sprintln("machine type must be one of bare, vm or none."), 0)
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "Type",
		Value:   serverType,
	}

	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Machine type successfully updated.")
	return
}

func opsBandwidth(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	bw int32,
) {
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "IncludedBandwidthGB",
		Value:   bw,
	}

	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); !ok {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Bandwidth allotment successfully updated.")
	return

}

func opsNic(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	nic int32,
) {
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "NICSpeedMbps",
		Value:   nic,
	}

	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); !ok {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("NIC speed successfully updated.")
	return

}

func opsExternalAddr(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	addr string,
) {
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "Addr",
		Value:   addr,
	}

	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); !ok {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Relay external address successfully updated.")
	return

}

func opsManagementAddr(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	addr string,
) {
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "ManagementAddr",
		Value:   addr,
	}

	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); !ok {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Relay management address successfully updated.")
	return

}

func opsSSHUser(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	username string,
) {
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "SSHUser",
		Value:   username,
	}

	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); !ok {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Relay SSH user name successfully updated.")
	return

}

func opsSSHPort(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	port int64,
) {
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "SSHPort",
		Value:   port,
	}

	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); !ok {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Relay SSH port number successfully updated.")
	return

}

func opsMaxSessions(rpcClient jsonrpc.RPCClient,
	env Environment,
	relayRegex string,
	sessions int64,
) {
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(rpcClient, env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.UpdateRelayArgs{
		RelayID: relayID,
		Field:   "SSHPort",
		Value:   sessions,
	}

	var reply localjsonrpc.UpdateRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.UpdateRelay", args); !ok {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Println("Relay max sessions value successfully updated.")
	return

}

func checkForRelay(rpcClient jsonrpc.RPCClient, env Environment, regex string) (uint64, bool) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	err := rpcClient.CallFor(&reply, "OpsService.Relays", args)
	if err != nil {
		handleJSONRPCError(env, err)
		return 0, false
	}

	if len(reply.Relays) == 0 {
		handleRunTimeError(fmt.Sprintf("Zero relays found matching '%s'", regex), 0)
	} else if len(reply.Relays) > 1 {
		handleRunTimeError(fmt.Sprintf("'%s' returned more than one relay - please be more specific.", regex), 0)
	}

	return reply.Relays[0].ID, true

}
