package main

import (
	"fmt"
	"strings"

	"github.com/modood/table"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func flushsessions(rpcClient jsonrpc.RPCClient, env Environment) {
	relaysargs := localjsonrpc.FlushSessionsArgs{}

	var relaysreply localjsonrpc.FlushSessionsReply
	if err := rpcClient.CallFor(&relaysreply, "BuyersService.FlushSessions", relaysargs); err != nil {
		handleJSONRPCError(env, err)
		return
	}
}

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

		stats := []struct {
			Name       string
			RTT        string
			Jitter     string
			PacketLoss string
		}{}

		if len(reply.Slices) > 0 {
			
			lastSlice := reply.Slices[len(reply.Slices)-1]

			if reply.Meta.OnNetworkNext {
				fmt.Printf( "Session is on Network Next\n\n" )
				fmt.Printf( "RTT improvement is %.1fms\n\n", lastSlice.Direct.RTT - lastSlice.Next.RTT)
			} else {
				fmt.Printf( "Session is going direct\n\n" )
			}

			stats = append(stats, struct {
				Name       string
				RTT        string
				Jitter     string
				PacketLoss string
			}{
				Name:       "Direct",
				RTT:        fmt.Sprintf("%.01f", lastSlice.Direct.RTT),
				Jitter:     fmt.Sprintf("%.01f", lastSlice.Direct.Jitter),
				PacketLoss: fmt.Sprintf("%.01f", lastSlice.Direct.PacketLoss),
			})

			if reply.Meta.OnNetworkNext {
				stats = append(stats, struct {
					Name       string
					RTT        string
					Jitter     string
					PacketLoss string
				}{
					Name:       "Next",
					RTT:        fmt.Sprintf("%.01f", lastSlice.Next.RTT),
					Jitter:     fmt.Sprintf("%.01f", lastSlice.Next.Jitter),
					PacketLoss: fmt.Sprintf("%.01f", lastSlice.Next.PacketLoss),
				})
			}
	
			table.Output(stats)
		}

		fmt.Println("\nNear Relays:\n")

		near := []struct {
			Name       string
			RTT        string
			Jitter     string
			PacketLoss string
		}{}

		for _, relay := range reply.Meta.NearbyRelays {
			for _, r := range relaysreply.Relays {
				if relay.ID == r.ID {
				    relay.Name = r.Name
				}
			}
			near = append(near, struct {
				Name       string
				RTT        string
				Jitter     string
				PacketLoss string
			}{
				Name:       relay.Name,
				RTT:        fmt.Sprintf("%.1f", relay.ClientStats.RTT),
				Jitter:     fmt.Sprintf("%.1f", relay.ClientStats.Jitter),
				PacketLoss: fmt.Sprintf("%.1f", relay.ClientStats.PacketLoss),
			})
		}

		table.Output(near)

		fmt.Printf( "\nRoute:\n\n" )

		clientAddr := strings.Replace(reply.Meta.ClientAddr, ":0", "", -1)
		fmt.Printf("    %s\n", clientAddr)
		for _, hop := range reply.Meta.Hops {
			for _, relay := range relaysreply.Relays {
				if hop.ID == relay.ID {
					hop.Name = relay.Name
				}
			}
			fmt.Printf("    %s\n", hop.Name)
		}
		fmt.Printf("    %s\n", reply.Meta.ServerAddr)

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
		ISP         string
		Datacenter  string
		DirectRTT   string
		NextRTT     string
		Improvement string
	}{}

	for _, session := range reply.Sessions {
		improvement := fmt.Sprintf("%.02f", session.DeltaRTT)
		if session.NextRTT <= 0 || session.DeltaRTT <= 0 {
			improvement = "-"
		}
		sessions = append(sessions, struct {
			ID          string
			UserHash    string
			ISP         string
			Datacenter  string
			DirectRTT   string
			NextRTT     string
			Improvement string
		}{
			ID:          session.ID,
			UserHash:    session.UserHash,
			ISP:         session.Location.ISP,
			Datacenter:  session.Datacenter,
			DirectRTT:   fmt.Sprintf("%.02f", session.DirectRTT),
			NextRTT:     fmt.Sprintf("%.02f", session.NextRTT),
			Improvement: improvement,
		})
	}

	table.Output(sessions)
}
