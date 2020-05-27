package main

import (
	"fmt"
	"time"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func relays(rpcClient jsonrpc.RPCClient, env Environment, filter string) {
	args := localjsonrpc.RelaysArgs{
		Name: filter,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	relays := []struct {
		Name        string
		State       string
		BindAddr    string
		SSHAddr     string
		Location    string
		Speed       string
		Bandwidth   string
		LastUpdated string
		Sessions    string
		BytesTxRx   string
	}{}

	for _, relay := range reply.Relays {
		relays = append(relays, struct {
			Name        string
			State       string
			BindAddr    string
			SSHAddr     string
			Location    string
			Speed       string
			Bandwidth   string
			LastUpdated string
			Sessions    string
			BytesTxRx   string
		}{
			Name:        relay.Name,
			State:       relay.State,
			BindAddr:    relay.Addr,
			SSHAddr:     fmt.Sprintf("%s@%s:%d", relay.SSHUser, relay.ManagementAddr, relay.SSHPort),
			Location:    fmt.Sprintf("%.2f, %.2f", relay.Latitude, relay.Longitude),
			Speed:       fmt.Sprintf("%dGB", relay.NICSpeedMbps/1000),
			Bandwidth:   fmt.Sprintf("%dGB", relay.IncludedBandwidthGB),
			LastUpdated: time.Since(relay.StateUpdateTime).Truncate(time.Second).String(),
			Sessions:    fmt.Sprintf("%d/%d", relay.SessionCount, relay.MaxSessionCount),
			BytesTxRx:   fmt.Sprintf("%d/%d", relay.BytesSent, relay.BytesReceived),
		})
	}

	table.Output(relays)
}
func addRelay(rpcClient jsonrpc.RPCClient, env Environment, relay routing.Relay) {
	args := localjsonrpc.AddRelayArgs{
		Relay: relay,
	}

	var reply localjsonrpc.AddRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddRelay", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Relay \"%s\" added to storage.\n", relay.Name)
}

func removeRelay(rpcClient jsonrpc.RPCClient, env Environment, name string) {
	info := getRelayInfo(rpcClient, name)

	args := localjsonrpc.RemoveRelayArgs{
		RelayID: info.id,
	}

	var reply localjsonrpc.RemoveRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.RemoveRelay", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Relay \"%s\" removed from storage.\n", name)
}
