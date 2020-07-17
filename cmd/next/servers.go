package main

import (
	"fmt"
	"regexp"

	"github.com/modood/table"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

type buyerServers struct {
	BuyerID    string `json:"buyer_id"`
	NumServers int    `json:"num_servers"`
}

type server struct {
	ServerAddrs string `json:"server_addrs"`
}

func servers(rpcClient jsonrpc.RPCClient, env Environment, buyerName string, serverCount int64) {
	buyerArgs := localjsonrpc.BuyersArgs{}

	var buyersReply localjsonrpc.BuyersReply
	if err := rpcClient.CallFor(&buyersReply, "OpsService.Buyers", buyerArgs); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	buyers := buyersReply.Buyers

	var serverArgs localjsonrpc.ServerArgs
	var serversReply localjsonrpc.ServerReply
	var servers []buyerServers

	if buyerName == "" {
		for _, buyer := range buyers {
			serverArgs.BuyerID = buyer.ID
			if err := rpcClient.CallFor(&serversReply, "OpsService.Servers", serverArgs); err != nil {
				handleJSONRPCError(env, err)
				return
			}
			servers = append(servers, buyerServers{BuyerID: buyer.ID, NumServers: len(serversReply.ServerAddresses)})
		}
		table.Output(servers)
		return
	}

	if len(buyers) > 0 {
		r := regexp.MustCompile("(?i)" + buyerName) // case-insensitive regex
		for _, buyer := range buyers {
			if r.MatchString(buyer.Name) {
				serverArgs.BuyerID = buyer.ID
				break
			}
		}
	}

	if err := rpcClient.CallFor(&serversReply, "OpsService.Servers", serverArgs); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	if len(serversReply.ServerAddresses) == 0 {
		fmt.Printf("No servers found for buyer ID: %v\n", serverArgs.BuyerID)
		return
	}

	var serverAddrs []server

	for _, s := range serversReply.ServerAddresses {
		serverAddrs = append(serverAddrs, server{ServerAddrs: s})
	}

	if serverCount > 0 && serverCount < int64(len(serversReply.ServerAddresses)) {
		table.Output(serverAddrs[0:serverCount])
	} else {
		table.Output(serverAddrs)
	}
}
