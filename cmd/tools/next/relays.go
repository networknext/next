package main

import (
	"log"

	"github.com/modood/table"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func relays(rpcClient jsonrpc.RPCClient, filter string) {
	args := localjsonrpc.RelaysArgs{
		Name: filter,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		log.Fatal(err)
	}

	table.Output(reply.Relays)
}
