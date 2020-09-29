package main

import (
	"fmt"

	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

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
