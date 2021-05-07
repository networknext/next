package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/modood/table"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func getFleetRelays(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	relayCount int64,
	alphaSort bool,
	regexName string,

) {

	var reply localjsonrpc.RelayFleetReply
	var args localjsonrpc.RelayFleetArgs

	if err := rpcClient.CallFor(&reply, "RelayFleetService.RelayFleet", args); err != nil {
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

func queryRelayBackend(
	env Environment,
	relayCount int64,
	alphaSort bool,
	regexName string,
) {

	var uri string
	var err error

	if uri, err = env.RelayBackendURL(); err != nil {
		handleRunTimeError(fmt.Sprintf("Cannot get get relay backend hostname: %v\n", err), 1)
	}

	uri += "/relays"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", env.AuthToken))

	r, err := client.Do(req)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not get relays csv from the portal: %v\n", err), 1)
	}
	defer r.Body.Close()

	reader := csv.NewReader(r.Body)
	relayData, err := reader.ReadAll()
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not parse relays csv file from %s: %v\n", uri, err), 1)
	}

	// drop headings row
	relayData = append(relayData[:0], relayData[1:]...)

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

	for _, relay := range relayData {

		maxSessions, err := strconv.Atoi(relay[4])
		if err != nil {
			// tbd
			// fmt.Printf("Error parsing MaxSessions for %s: %v\n", relay[0], err)
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
			relay[0],
			strings.Split(relay[1], ":")[0],
			strings.ToUpper(relay[2]),
			relay[3],
			maxSessions,
			relay[5],
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
