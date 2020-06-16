package main

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func relays(rpcClient jsonrpc.RPCClient, env Environment, filter string, relaysStateShowFlags [6]bool, relaysStateHideFlags [6]bool, relaysDownFlag bool, relaysAllFlag bool) {
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
		Address     string
		State       string
		Sessions    string
		Tx          string
		Rx          string
		Version     string
		LastUpdated string
	}{}

	for _, relay := range reply.Relays {
		relayState, err := routing.ParseRelayState(relay.State)
		if err != nil {
			log.Fatalf("could not parse invalid relay state %s", relay.State)
		}

		includeRelay := true

		if relaysStateHideFlags[relayState] {
			// Relay should be hidden, so don't include in final output
			includeRelay = false
		}

		for i, flag := range relaysStateShowFlags {
			if flag {
				if relayState != routing.RelayState(i) {
					// An "only show" flag is set and this relay doesn't match that state, so don't include it in the final output
					includeRelay = false
				} else {
					// One of the flags should include the relay, so set to true and break out, since combining the flags is an OR operation
					includeRelay = true
					break
				}
			}
		}

		tx := fmt.Sprintf("%.02fGB", float64(relay.BytesSent)/float64(1000000000))
		if relay.BytesSent < 1000000000 {
			tx = fmt.Sprintf("%.02fMB", float64(relay.BytesSent)/float64(1000000))
		}
		rx := fmt.Sprintf("%.02fGB", float64(relay.BytesReceived)/float64(1000000000))
		if relay.BytesReceived < 1000000000 {
			rx = fmt.Sprintf("%.02fMB", float64(relay.BytesReceived)/float64(1000000))
		}
		lastUpdateDuration := time.Since(relay.LastUpdateTime).Truncate(time.Second)
		lastUpdated := "n/a"
		if relay.State == "enabled" {
			lastUpdated = lastUpdateDuration.String()
		}

		if relaysDownFlag && lastUpdateDuration < 30*time.Second {
			// Relay is still up and shouldn't be included in the final output
			includeRelay = false
		}

		if relaysAllFlag {
			// We should show all relays so include it regardless of previous logic
			includeRelay = true
		}

		if !includeRelay {
			continue
		}

		address := relay.Addr

		relays = append(relays, struct {
			Name        string
			Address     string
			State       string
			Sessions    string
			Tx          string
			Rx          string
			Version     string
			LastUpdated string
		}{
			Name:        relay.Name,
			Address:     address,
			State:       relay.State,
			Sessions:    fmt.Sprintf("%d", relay.SessionCount),
			Tx:          tx,
			Rx:          rx,
			Version:     relay.Version,
			LastUpdated: lastUpdated,
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

	fmt.Printf("Relay \"%s\" decommissioned.\n", name)
}
