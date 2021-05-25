package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/modood/table"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
)

func getFleetRelays(
	env Environment,
	relayCount int64,
	alphaSort bool,
	regexName string,

) {

	var reply localjsonrpc.RelayFleetReply = localjsonrpc.RelayFleetReply{}
	var args = localjsonrpc.RelayFleetArgs{}

	if err := makeRPCCall(env, &reply, "RelayFleetService.RelayFleet", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	// name,address,id,status,sessions,version
	relays := []struct {
		Name     string
		Address  string
		Id       string
		Status   string
		Sessions int
		Version  string
	}{}

	filtered := []struct {
		Name     string
		Address  string
		Id       string
		Status   string
		Sessions int
		Version  string
	}{}

	for _, relay := range reply.RelayFleet {

		maxSessions, err := strconv.Atoi(relay.Sessions)
		if err != nil {
			maxSessions = -1
		}

		relays = append(relays, struct {
			Name     string
			Address  string
			Id       string
			Status   string
			Sessions int
			Version  string
		}{
			relay.Name,
			strings.Split(relay.Address, ":")[0],
			strings.ToUpper(relay.Id),
			relay.Status,
			maxSessions,
			relay.Version,
		})
	}

	for _, relay := range relays {
		if match, err := regexp.Match(regexName, []byte(relay.Name)); match && err == nil {
			filtered = append(filtered, relay)
			continue
		}
	}

	if alphaSort {
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].Name < filtered[j].Name
		})
	} else {
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].Sessions > filtered[j].Sessions
		})
	}

	outputRelays := []struct {
		Name     string
		Address  string
		Id       string
		Status   string
		Sessions string
		Version  string
	}{}

	for _, relay := range filtered {

		sessions := fmt.Sprintf("%d", relay.Sessions)

		if relay.Sessions == -1 {
			sessions = ""
		}
		outputRelays = append(outputRelays, struct {
			Name     string
			Address  string
			Id       string
			Status   string
			Sessions string
			Version  string
		}{
			relay.Name,
			relay.Address,
			relay.Id,
			relay.Status,
			sessions,
			relay.Version,
		})
	}

	//  limit the number of relays displayed
	if relayCount != 0 {
		table.Output(outputRelays[0:relayCount])
	} else {
		table.Output(outputRelays)
	}

}
