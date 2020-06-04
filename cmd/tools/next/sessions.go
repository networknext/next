package main

import (
	"fmt"

	"github.com/modood/table"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func sessions(rpcClient jsonrpc.RPCClient, env Environment, sessionID string) {
	if sessionID != "" {
		relaysargs := localjsonrpc.RelaysArgs{}

		var relaysreply localjsonrpc.RelaysReply
		if err := rpcClient.CallFor(&relaysreply, "OpsService.Relays", relaysargs); err != nil {
			handleJSONRPCError(env, err)
			return
		}

		args := localjsonrpc.SessionDetailsArgs{
			SessionID: sessionID,
		}

		var reply localjsonrpc.SessionDetailsReply
		if err := rpcClient.CallFor(&reply, "BuyersService.SessionDetails", args); err != nil {
			handleJSONRPCError(env, err)
			return
		}

		fmt.Println("Session ID:", sessionID)
		fmt.Println("User Hash:", reply.Meta.UserHash)
		fmt.Printf("Current Route: %s, %s ->", reply.Meta.Location.City, reply.Meta.Location.Region)
		for _, hop := range reply.Meta.Hops {
			for _, relay := range relaysreply.Relays {
				if hop.ID == relay.ID {
					hop.Name = relay.Name
				}
			}

			fmt.Printf(" %s ->", hop.Name)
		}
		fmt.Println(" ", reply.Meta.Datacenter)

		return
	}

	args := localjsonrpc.TopSessionsArgs{}

	var reply localjsonrpc.TopSessionsReply
	if err := rpcClient.CallFor(&reply, "BuyersService.TopSessions", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	sessions := []struct {
		ID          string
		UserHash    string
		Location    string
		Datacenter  string
		DirectRTT   string
		NextRTT     string
		Improvement string
	}{}

	for _, session := range reply.Sessions {
		sessions = append(sessions, struct {
			ID          string
			UserHash    string
			Location    string
			Datacenter  string
			DirectRTT   string
			NextRTT     string
			Improvement string
		}{
			ID:          session.ID,
			UserHash:    session.UserHash,
			Location:    fmt.Sprintf("%s, %s", session.Location.City, session.Location.Region),
			Datacenter:  session.Datacenter,
			DirectRTT:   fmt.Sprintf("%.02f", session.DirectRTT),
			NextRTT:     fmt.Sprintf("%.02f", session.NextRTT),
			Improvement: fmt.Sprintf("%.02f", session.DeltaRTT),
		})
	}

	table.Output(sessions)
}
