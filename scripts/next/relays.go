package main

import (
	"fmt"
	"strings"

	"github.com/modood/table"

	"github.com/networknext/backend/modules-old/routing"

	localjsonrpc "github.com/networknext/backend/modules-old/transport/jsonrpc"
)

func relays(
	env Environment,
	regex string,
	relaysStateShowFlags [6]bool,
	relaysStateHideFlags [6]bool,
	relaysDownFlag bool,
	relaysListFlag bool,
	csvOutputFlag bool,
	relayVersionFilter string,
	relaysCount int64,
	relayIDSigned bool,
) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	// debugPrintRelayViewFlags("function entry", relaysStateHideFlags, relaysStateShowFlags)
	// os.Exit(0)

	var reply localjsonrpc.RelaysReply
	if err := makeRPCCall(env, &reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	relays := []struct {
		Name     string
		ID       string
		Address  string
		Internal string
		State    string
	}{}

	for _, relay := range reply.Relays {
		relayState, err := routing.ParseRelayState(relay.State)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("could not parse invalid relay state %s\n", relay.State), 0)
		}

		// TODO: fix once routing.Relay.State is updated
		if relay.State == "decommissioned" {
			relay.State = "removed"
		}

		includeRelay := true

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

		if relaysStateHideFlags[relayState] {
			// Relay should be hidden, so don't include in final output
			includeRelay = false
		}

		if !includeRelay {
			continue
		}

		address := relay.Addr
		internal := relay.InternalAddr

		if relayVersionFilter == "all" || relay.Version == relayVersionFilter {
			var relayID string
			if relayIDSigned {
				relayID = fmt.Sprintf("%d", int64(relay.ID))
			} else {
				relayID = fmt.Sprintf("%016x", relay.ID)
			}
			relays = append(relays, struct {
				Name     string
				ID       string
				Address  string
				Internal string
				State    string
			}{
				Name:     relay.Name,
				ID:       relayID,
				Address:  address,
				Internal: internal,
				State:    relay.State,
			})
		}

	}

	if relaysListFlag {
		relayNames := []string{}
		for _, relay := range relays {
			relayNames = append(relayNames, relay.Name)

		}
		fmt.Println(strings.Join(relayNames, " "))
		return
	}

	if relaysCount > 0 && int(relaysCount) < len(relays) {
		relays = relays[:relaysCount]
	}

	table.Output(relays)

}
