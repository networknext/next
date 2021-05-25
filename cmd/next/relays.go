package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/modood/table"
	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
)

func opsRelays(
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

	// debugPrintRelayViewFlags("function entry", relaysStateHideFlags, relaysStateShowFlags)
	// os.Exit(0)

	var reply localjsonrpc.RelaysReply
	if err := makeRPCCall(env, &reply, "OpsService.Relays", args); err != nil {
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
		State               string
		IPAddress           string
		Notes               string
	}{}

	relaysCSV := [][]string{{}}

	relaysCSV = append(relaysCSV, []string{
		"Name", "MRC", "Overage", "BW Rule",
		"Term", "Start Date", "End Date", "Type", "Bandwidth", "NIC Speed", "State", "IP Address", "Notes"})

	for _, relay := range reply.Relays {
		relayState, err := routing.ParseRelayState(relay.State)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("could not parse invalid relay state %s\n", relay.State), 1)
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

		mrc := ""
		if relay.MRC > 0 {
			mrc = fmt.Sprintf("%.2f", relay.MRC.ToCents()/100)
		}
		overage := ""
		if relay.Overage > 0 {
			overage = fmt.Sprintf("%.5f", relay.Overage.ToCents()/100)
		}

		var bwRule string
		switch relay.BWRule {
		case routing.BWRuleNone:
			bwRule = ""
		case routing.BWRuleFlat:
			bwRule = "flat"
		case routing.BWRuleBurst:
			bwRule = "burst"
		case routing.BWRulePool:
			bwRule = "pool"
		default:
			bwRule = ""
		}

		var machineType string
		switch relay.Type {
		case routing.NoneSpecified:
			machineType = ""
		case routing.BareMetal:
			machineType = "bare metal"
		case routing.VirtualMachine:
			machineType = "virtual machine"
		default:
			machineType = ""
		}

		contractTerm := ""
		if relay.ContractTerm != 0 {
			contractTerm = fmt.Sprintf("%d", relay.ContractTerm)
		}

		startDate := ""
		if !relay.StartDate.IsZero() {
			startDate = relay.StartDate.Format("January 2, 2006")
		}

		endDate := ""
		if !relay.EndDate.IsZero() {
			endDate = relay.EndDate.Format("January 2, 2006")
		}

		bandwidth := strconv.FormatInt(int64(relay.IncludedBandwidthGB), 10)
		if bandwidth == "0" {
			bandwidth = ""
		}

		nicSpeed := strconv.FormatInt(int64(relay.NICSpeedMbps), 10)
		if nicSpeed == "0" {
			nicSpeed = ""
		}

		// return csv file
		if relayVersionFilter == "all" || relay.Version == relayVersionFilter {
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
					relay.State,
					strings.Split(relay.Addr, ":")[0],
					relay.Notes,
				})
			} else {
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
					State               string
					IPAddress           string
					Notes               string
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
					relay.State,
					strings.Split(relay.Addr, ":")[0],
					relay.Notes,
				})
			}
		}

	}

	if csvOutputFlag {
		if relaysCount > 0 && int(relaysCount) < len(relaysCSV) {
			relaysCSV = relaysCSV[:relaysCount+2] // +2 for heading lines
		}

		// return csv file of structs
		// fileName := "./relays-" + strconv.FormatInt(time.Now().Unix(), 10) + ".csv"
		fileName := "./relays.csv"
		f, err := os.Create(fileName)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("Error creating local CSV file %s: %v\n", fileName, err), 1)
		}

		writer := csv.NewWriter(f)
		err = writer.WriteAll(relaysCSV)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("Error writing local CSV file %s: %v\n", fileName, err), 1)
		}
		fmt.Println("CSV file written: relays.csv")
		return
	}

	if relaysCount > 0 && int(relaysCount) < len(relays) {
		relays = relays[:relaysCount]
	}

	table.Output(relays)
}

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
		Name        string
		ID          string
		Address     string
		State       string
		Sessions    string
		Tx          string
		Rx          string
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
			handleRunTimeError(fmt.Sprintf("could not parse invalid relay state %s\n", relay.State), 0)
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
					"n/a",
					"n/a",
					"n/a",
					relay.Version,
					"n/a",
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
				LastUpdated string
			}{
				Name:    relay.Name,
				ID:      relayID,
				Address: address,
				State:   relay.State,
			})
		}

	}

	if csvOutputFlag {
		if relaysCount > 0 && int(relaysCount) < len(relaysCSV) {
			relaysCSV = relaysCSV[:relaysCount+2] // +2 for heading lines
		}

		// return csv file of structs
		// fileName := "./relays-" + strconv.FormatInt(time.Now().Unix(), 10) + ".csv"
		fileName := "./relays.csv"
		f, err := os.Create(fileName)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("Error creating local CSV file %s: %v\n", fileName, err), 1)
		}

		writer := csv.NewWriter(f)
		err = writer.WriteAll(relaysCSV)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("Error writing local CSV file %s: %v\n", fileName, err), 1)
		}
		fmt.Println("CSV file written: relays.csv")
		return
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

func addRelayJS(env Environment, r relay) {

	bwRule, err := routing.ParseBandwidthRule(r.BWRule)
	if err != nil {
		handleJSONRPCError(env, err)
		return
	}

	machineType, err := routing.ParseMachineType(r.Type)
	if err != nil {
		handleJSONRPCError(env, err)
		return
	}

	args := localjsonrpc.JSAddRelayArgs{
		Name:                r.Name,
		Addr:                r.Addr,
		InternalAddr:        r.InternalAddr,
		PublicKey:           r.PublicKey,
		SellerID:            r.Seller,       // not used
		DatacenterID:        r.DatacenterID, // hex
		NICSpeedMbps:        int64(r.NicSpeedMbps),
		IncludedBandwidthGB: int64(r.IncludedBandwidthGB),
		ManagementAddr:      r.ManagementAddr,
		SSHUser:             r.SSHUser,
		SSHPort:             r.SSHPort,
		MaxSessions:         int64(r.MaxSessions),
		MRC:                 int64(routing.DollarsToNibblins(r.MRC)),
		Overage:             int64(routing.DollarsToNibblins(r.Overage)),
		BWRule:              int64(bwRule),
		ContractTerm:        int64(r.ContractTerm),
		StartDate:           r.StartDate,
		EndDate:             r.EndDate,
		Type:                int64(machineType),
		BillingSupplier:     r.BillingSupplier,
		Version:             r.Version,
	}

	var reply localjsonrpc.JSAddRelayReply
	if err := makeRPCCall(env, &reply, "OpsService.JSAddRelay", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Relay \"%s\" added to storage.\n", r.Name)

}

func removeRelay(env Environment, name string) {
	relays := getRelayInfo(env, name)

	if len(relays) == 0 {
		handleRunTimeError(fmt.Sprintf("no relays matched the name '%s'\n", name), 0)
	}

	info := relays[0]

	if info.state == routing.RelayStateDecommissioned.String() {
		fmt.Printf("Relay \"%s\" already removed\n", info.name)
		os.Exit(0)
	}

	if info.state != routing.RelayStateDisabled.String() {
		fmt.Printf("Relay %s must be disabled prior to removal.\n\n", info.name)
		os.Exit(0)
	}

	args := localjsonrpc.RemoveRelayArgs{
		RelayID: info.id,
	}

	var reply localjsonrpc.RemoveRelayReply
	if err := makeRPCCall(env, &reply, "OpsService.RemoveRelay", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Relay \"%s\" removed.\n", info.name)
}

func countRelays(env Environment, regex string) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	if err := makeRPCCall(env, &reply, "OpsService.Relays", args); err != nil {
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

	var totalRelays int

	for key, relayCount := range relayCountList {
		if key != "decommissioned" {
			totalRelays += relayCount
			relayList = append(relayList, struct {
				State string
				Count string
			}{
				State: key,
				Count: strconv.Itoa(relayCount),
			})
		}
	}

	relayList = append(relayList, struct {
		State string
		Count string
	}{
		State: "total",
		Count: strconv.Itoa(totalRelays),
	})

	table.Output(relayList)

}

func modifyRelayField(
	env Environment,
	relayRegex string,
	field string,
	value string,
) error {

	args := localjsonrpc.RelaysArgs{
		Regex: relayRegex,
	}

	var reply localjsonrpc.RelaysReply
	if err := makeRPCCall(env, &reply, "OpsService.Relays", args); err != nil {
		handleJSONRPCError(env, err)
		return nil
	}

	if len(reply.Relays) == 0 {
		handleRunTimeError(fmt.Sprintf("No relay matches found for '%s'", relayRegex), 0)
	}

	if len(reply.Relays) > 1 {
		fmt.Printf("Found several  matches for '%s'", relayRegex)
		for _, relay := range reply.Relays {
			fmt.Printf("\t%s\n", relay.Name)
		}
		handleRunTimeError(fmt.Sprintln("Please be more specific."), 0)
	}

	emptyReply := localjsonrpc.ModifyRelayFieldReply{}

	modifyArgs := localjsonrpc.ModifyRelayFieldArgs{
		RelayID: reply.Relays[0].ID,
		Field:   field,
		Value:   value,
	}
	if err := makeRPCCall(env, &emptyReply, "OpsService.ModifyRelayField", modifyArgs); err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}

	fmt.Printf("Field %s for relay %s updated successfully.\n", field, reply.Relays[0].Name)
	return nil
}

func debugPrintRelayViewFlags(where string, hide [6]bool, show [6]bool) {
	fmt.Printf("\nHidden States ( %s ):\n", where)
	for i, hiddenRelayState := range hide {
		state, err := routing.GetRelayStateSQL(int64(i))
		if err != nil {
			handleRunTimeError(fmt.Sprintf("error parsing relay state: %v'", err), 0)
		}
		fmt.Printf("\t%s: %t\n", state, hiddenRelayState)
	}

	fmt.Printf("\nShown States ( %s ):\n", where)
	for i, shownRelayState := range show {
		state, err := routing.GetRelayStateSQL(int64(i))
		if err != nil {
			handleRunTimeError(fmt.Sprintf("error parsing relay state: %v'", err), 0)
		}
		fmt.Printf("\t%s: %t\n", state, shownRelayState)
	}
	fmt.Println()
}
