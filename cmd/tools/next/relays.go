package main

import (
	"fmt"
	"sort"
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

	sort.Slice(reply.Relays, func(i int, j int) bool {
		return reply.Relays[i].SessionCount > reply.Relays[j].SessionCount
	})

	relays := []struct {
		Name        string
		State       string
		Sessions    string
		Tx          string
		Rx          string
		LastUpdated string
	}{}

	for _, relay := range reply.Relays {
		tx := fmt.Sprintf("%.02fGB", float64(relay.BytesSent)/float64(1000000000))
		if relay.BytesSent < 1000000000 {
			tx = fmt.Sprintf("%.02fMB", float64(relay.BytesSent)/float64(1000000))
		}
		rx := fmt.Sprintf("%.02fGB", float64(relay.BytesReceived)/float64(1000000000))
		if relay.BytesReceived < 1000000000 {
			rx = fmt.Sprintf("%.02fMB", float64(relay.BytesReceived)/float64(1000000))
		}

		relays = append(relays, struct {
			Name        string
			State       string
			Sessions    string
			Tx          string
			Rx          string
			LastUpdated string
		}{
			Name:        relay.Name,
			State:       relay.State,
			Sessions:    fmt.Sprintf("%d", relay.SessionCount),
			Tx:          tx,
			Rx:          rx,
			LastUpdated: time.Since(relay.LastUpdateTime).Truncate(time.Second).String(),
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
