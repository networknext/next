package main

import (
	"github.com/modood/table"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func invoiceBuyer(rpcClient jsonrpc.RPCClient, env Environment) {
	args := localjsonrpc.InvoiceArgs{}

	var reply localjsonrpc.InvoiceReply
	if err := rpcClient.CallFor(&reply, "InvoiceService.InvoiceBuyer", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	// need to generate csv
	table.Output(reply.Invoices)
}
