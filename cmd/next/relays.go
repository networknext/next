package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func opsRelays(
	rpcClient jsonrpc.RPCClient,
	env Environment,
	regex string,
	relaysStateShowFlags [6]bool,
	relaysStateHideFlags [6]bool,
	relaysDownFlag bool,
	csvOutputFlag bool,
	relayVersionFilter string,
	relaysCount int64,
	relayIDSigned bool,
	relayBWSort bool,
) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	if relayBWSort {
		sort.Slice(reply.Relays, func(i int, j int) bool {
			return reply.Relays[i].IncludedBandwidthGB > reply.Relays[j].IncludedBandwidthGB
		})
	}

	relays := []struct {
		Name                string
		MRC                 string
		Overage             string
		BWRule              string
		ContractTerm        string
		StartDate           string
		EndDate             string
		Type                string
		IncludedBandwidthGB string
		NICSpeedMbps        string
	}{}

	relaysCSV := [][]string{{}}

	relaysCSV = append(relaysCSV, []string{
		"Name", "Address", "MRC", "Overage", "BW Rule",
		"Term", "Start Date", "End Date", "Type", "Bandwidth", "NIC Speed"})

	for _, relay := range reply.Relays {
		relayState, err := routing.ParseRelayState(relay.State)
		if err != nil {
			fmt.Printf("could not parse invalid relay state %s", relay.State)
			os.Exit(0)
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

		if relaysDownFlag {
			// Relay is still up and shouldn't be included in the final output
			includeRelay = false
		}

		if !includeRelay {
			continue
		}

		mrc := "n/a"
		if relay.MRC > 0 {
			mrc = fmt.Sprintf("%.2f", relay.MRC.ToCents()/100)
		}
		overage := "n/a"
		if relay.Overage > 0 {
			overage = fmt.Sprintf("%.2f", relay.Overage.ToCents()/100)
		}

		var bwRule string
		switch relay.BWRule {
		case routing.BWRuleNone:
			bwRule = "n/a"
		case routing.BWRuleFlat:
			bwRule = "flat"
		case routing.BWRuleBurst:
			bwRule = "burst"
		case routing.BWRulePool:
			bwRule = "pool"
		default:
			bwRule = "n/a"
		}

		var machineType string
		switch relay.Type {
		case routing.NoneSpecified:
			machineType = "n/a"
		case routing.BareMetal:
			machineType = "bare metal"
		case routing.VirtualMachine:
			machineType = "virtual machine"
		default:
			machineType = "n/a"
		}

		contractTerm := "n/a"
		if relay.ContractTerm != 0 {
			contractTerm = fmt.Sprintf("%d", relay.ContractTerm)
		}

		startDate := "n/a"
		if !relay.StartDate.IsZero() {
			startDate = relay.StartDate.Format("January 2, 2006")
		}

		endDate := "n/a"
		if !relay.EndDate.IsZero() {
			endDate = relay.EndDate.Format("January 2, 2006")
		}

		bandwidth := strconv.FormatInt(int64(relay.IncludedBandwidthGB), 10)
		if bandwidth == "0" {
			bandwidth = "n/a"
		}

		nicSpeed := strconv.FormatInt(int64(relay.NICSpeedMbps), 10)
		if nicSpeed == "0" {
			nicSpeed = "n/a"
		}

		// return csv file
		if csvOutputFlag {
			relaysCSV = append(relaysCSV, []string{
				relay.Name,
				mrc,
				overage,
				bwRule,
				contractTerm,
				startDate,
				endDate,
				machineType,
				bandwidth,
				nicSpeed,
			})
		} else if relayVersionFilter == "all" || relay.Version == relayVersionFilter {
			relays = append(relays, struct {
				Name                string
				MRC                 string
				Overage             string
				BWRule              string
				ContractTerm        string
				StartDate           string
				EndDate             string
				Type                string
				IncludedBandwidthGB string
				NICSpeedMbps        string
			}{
				relay.Name,
				mrc,
				overage,
				bwRule,
				contractTerm,
				startDate,
				endDate,
				machineType,
				bandwidth,
				nicSpeed,
			})
		}

	}

	if csvOutputFlag {
		if relaysCount > 0 {
			relaysCSV = relaysCSV[:relaysCount]
		}

		// return csv file of structs
		// fileName := "./relays-" + strconv.FormatInt(time.Now().Unix(), 10) + ".csv"
		fileName := "./relays.csv"
		f, err := os.Create(fileName)
		if err != nil {
			fmt.Printf("Error creating local CSV file %s: %v\n", fileName, err)
			return
		}

		writer := csv.NewWriter(f)
		err = writer.WriteAll(relaysCSV)
		if err != nil {
			fmt.Printf("Error writing local CSV file %s: %v\n", fileName, err)
		}
		fmt.Println("CSV file written: relays.csv")
		return
	}

	if relaysCount > 0 {
		relays = relays[:relaysCount]
	}

	table.Output(relays)
}

func relays(
	rpcClient jsonrpc.RPCClient,
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
		ID          string
		Address     string
		State       string
		Sessions    string
		Tx          string
		Rx          string
		Version     string
		CPUUsage    string `table:"CPU Usage"`
		MemUsage    string `table:"Memory Usage"`
		LastUpdated string
	}{}

	relaysCSV := [][]string{{}}

	if relaysListFlag {
		relaysCSV = append(relaysCSV, []string{"Name"})
	} else {
		relaysCSV = append(relaysCSV, []string{
			"Name", "ID", "Address", "State", "Sessions", "Tx", "Rx", "Version", "LastUpdated"})
	}

	for _, relay := range reply.Relays {
		relayState, err := routing.ParseRelayState(relay.State)
		if err != nil {
			log.Fatalf("could not parse invalid relay state %s", relay.State)
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

		tx := fmt.Sprintf("%.02fGB", float64(relay.TrafficStats.BytesSent)/float64(1000000000))
		if relay.TrafficStats.BytesSent < 1000000000 {
			tx = fmt.Sprintf("%.02fMB", float64(relay.TrafficStats.BytesSent)/float64(1000000))
		}
		rx := fmt.Sprintf("%.02fGB", float64(relay.TrafficStats.BytesReceived)/float64(1000000000))
		if relay.TrafficStats.BytesReceived < 1000000000 {
			rx = fmt.Sprintf("%.02fMB", float64(relay.TrafficStats.BytesReceived)/float64(1000000))
		}
		cpuUsage := fmt.Sprintf("%.02f%%", relay.CPUUsage)
		memUsage := fmt.Sprintf("%.02f%%", relay.MemUsage)
		lastUpdateDuration := time.Since(relay.LastUpdateTime).Truncate(time.Second)
		lastUpdated := "n/a"
		if relay.State == "enabled" {
			lastUpdated = lastUpdateDuration.String()
		}

		if relaysDownFlag && lastUpdateDuration < 30*time.Second {
			// Relay is still up and shouldn't be included in the final output
			includeRelay = false
		}

		if !includeRelay {
			continue
		}

		address := relay.Addr

		// return csv file
		if csvOutputFlag {
			if relaysListFlag && (relayVersionFilter == "all" || relay.Version == relayVersionFilter) {
				relaysCSV = append(relaysCSV, []string{
					relay.Name,
				})
			} else if relayVersionFilter == "all" || relay.Version == relayVersionFilter {
				var relayID string
				if relayIDSigned {
					relayID = fmt.Sprintf("%d", int64(relay.ID))
				} else {
					relayID = fmt.Sprintf("%016x", relay.ID)
				}
				relaysCSV = append(relaysCSV, []string{
					relay.Name,
					relayID,
					address,
					relay.State,
					fmt.Sprintf("%d", relay.SessionCount),
					tx,
					rx,
					relay.Version,
					lastUpdated,
				})
			}

		} else if relayVersionFilter == "all" || relay.Version == relayVersionFilter {
			var relayID string
			if relayIDSigned {
				relayID = fmt.Sprintf("%d", int64(relay.ID))
			} else {
				relayID = fmt.Sprintf("%016x", relay.ID)
			}
			relays = append(relays, struct {
				Name        string
				ID          string
				Address     string
				State       string
				Sessions    string
				Tx          string
				Rx          string
				Version     string
				CPUUsage    string `table:"CPU Usage"`
				MemUsage    string `table:"Memory Usage"`
				LastUpdated string
			}{
				Name:        relay.Name,
				ID:          relayID,
				Address:     address,
				State:       relay.State,
				Sessions:    fmt.Sprintf("%d", relay.SessionCount),
				Tx:          tx,
				Rx:          rx,
				Version:     relay.Version,
				CPUUsage:    cpuUsage,
				MemUsage:    memUsage,
				LastUpdated: lastUpdated,
			})
		}

	}

	if csvOutputFlag {
		if relaysCount > 0 {
			relaysCSV = relaysCSV[:relaysCount]
		}

		// return csv file of structs
		// fileName := "./relays-" + strconv.FormatInt(time.Now().Unix(), 10) + ".csv"
		fileName := "./relays.csv"
		f, err := os.Create(fileName)
		if err != nil {
			fmt.Printf("Error creating local CSV file %s: %v\n", fileName, err)
			return
		}

		writer := csv.NewWriter(f)
		err = writer.WriteAll(relaysCSV)
		if err != nil {
			fmt.Printf("Error writing local CSV file %s: %v\n", fileName, err)
		}
		fmt.Println("CSV file written: relays.csv")
		return
	}

	if relaysCount > 0 {
		relays = relays[:relaysCount]
	}

	if relaysListFlag {
		relayNames := []string{}
		for _, relay := range relays {
			relayNames = append(relayNames, relay.Name)

		}
		fmt.Println(strings.Join(relayNames, " "))
		return
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
	relays := getRelayInfo(rpcClient, env, name)

	if len(relays) == 0 {
		log.Fatalf("no relays matched the name '%s'\n", name)
	}

	info := relays[0]

	if info.state == routing.RelayStateDecommissioned.String() {
		log.Fatalf("Relay \"%s\" already removed\n", info.name)
	}

	args := localjsonrpc.RemoveRelayArgs{
		RelayID: info.id,
	}

	var reply localjsonrpc.RemoveRelayReply
	if err := rpcClient.CallFor(&reply, "OpsService.RemoveRelay", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Relay \"%s\" removed.\n", info.name)
}

func countRelays(rpcClient jsonrpc.RPCClient, env Environment, regex string) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	relayList := []struct {
		State string
		Count string
	}{}

	relayCountList := make(map[string]int)

	for _, relay := range reply.Relays {
		if _, ok := relayCountList[relay.State]; ok {
			relayCountList[relay.State]++
			continue
		}
		relayCountList[relay.State] = 1
	}

	for key, relayCount := range relayCountList {
		relayList = append(relayList, struct {
			State string
			Count string
		}{
			State: key,
			Count: strconv.Itoa(relayCount),
		})
	}

	relayList = append(relayList, struct {
		State string
		Count string
	}{
		State: "total",
		Count: strconv.Itoa(len(reply.Relays)),
	})

	table.Output(relayList)

}

func relayTraffic(rpcClient jsonrpc.RPCClient, env Environment, regex string) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := rpcClient.CallFor(&reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	statsList := []struct {
		Name       string `table:"Name"`
		InternalTx uint64 `table:"Internal Rx"`
		InternalRx uint64 `table:"Internal Tx"`
		GameTx     uint64 `table:"Game Rx"`
		GameRx     uint64 `table:"Game Tx"`
		UnknownRx  uint64 `table:"Unknown Rx"`
	}{}

	for i := range reply.Relays {
		relay := &reply.Relays[i]

		statsList = append(statsList, struct {
			Name       string `table:"Name"`
			InternalTx uint64 `table:"Internal Rx"`
			InternalRx uint64 `table:"Internal Tx"`
			GameTx     uint64 `table:"Game Rx"`
			GameRx     uint64 `table:"Game Tx"`
			UnknownRx  uint64 `table:"Unknown Rx"`
		}{
			Name:       relay.Name,
			InternalRx: relay.TrafficStats.InternalStatsRx(),
			InternalTx: relay.TrafficStats.InternalStatsTx(),
			GameRx:     relay.TrafficStats.GameStatsRx(),
			GameTx:     relay.TrafficStats.GameStatsTx(),
			UnknownRx:  relay.TrafficStats.UnknownRx,
		})
	}

	table.Output(statsList)
}
