package main

import (
	"fmt"

	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
)

func getDetailedRelayInfo(
	env Environment,
	relayRegex string,
) {
	var relayID uint64
	var ok bool
	if relayID, ok = checkForRelay(env, relayRegex); !ok {
		// error msg printed by called function
		return
	}

	args := localjsonrpc.GetRelayArgs{
		RelayID: relayID,
	}

	var reply localjsonrpc.GetRelayReply
	if err := makeRPCCall(env, &reply, "OpsService.GetRelay", args); !ok {
		handleJSONRPCError(env, err)
		return
	}

	var bwRule string
	switch reply.Relay.BWRule {
	case routing.BWRuleNone:
		bwRule = "none specified"
	case routing.BWRuleFlat:
		bwRule = "flat"
	case routing.BWRulePool:
		bwRule = "pool"
	case routing.BWRuleBurst:
		bwRule = "burst"
	}

	var machineType string
	switch reply.Relay.Type {
	case routing.NoneSpecified:
		machineType = "none specified"
	case routing.BareMetal:
		machineType = "bare metal"
	case routing.VirtualMachine:
		machineType = "virtual machine"
	}

	var startDate, endDate string

	if !reply.Relay.StartDate.IsZero() {
		startDate = reply.Relay.StartDate.Format("January 2, 2006")
	}

	if !reply.Relay.EndDate.IsZero() {
		endDate = reply.Relay.EndDate.Format("January 2, 2006")
	}

	// TODO: remove once routing.Relay.State is updated
	if reply.Relay.State == "decommissioned" {
		reply.Relay.State = "removed"
	}

	relay := "\nrelay info:\n"
	relay += "  ID                 : " + fmt.Sprintf("%016x", reply.Relay.ID) + "\n"
	relay += "  Name               : " + reply.Relay.Name + "\n"
	relay += "  Addr               : " + reply.Relay.Addr + "\n"
	relay += "  InternalAddr       : " + reply.Relay.InternalAddr + "\n"
	relay += "  PublicKey          : " + string(reply.Relay.PublicKey) + "\n"
	relay += "  Datacenter         : " + fmt.Sprintf("%016x", reply.Relay.DatacenterID) + "\n"
	relay += "  BillingSupplier    : " + fmt.Sprintf("%s", reply.Relay.BillingSupplier) + "\n"
	relay += "  Seller             : " + reply.Relay.SellerName + "\n"
	relay += "  NICSpeedMbps       : " + fmt.Sprintf("%d", reply.Relay.NICSpeedMbps) + "\n"
	relay += "  IncludedBandwidthGB: " + fmt.Sprintf("%d", reply.Relay.IncludedBandwidthGB) + "\n"
	relay += "  State              : " + fmt.Sprintf("%v", reply.Relay.State) + "\n"
	relay += "  ManagementAddr     : " + reply.Relay.ManagementAddr + "\n"
	relay += "  SSHUser            : " + reply.Relay.SSHUser + "\n"
	relay += "  SSHPort            : " + fmt.Sprintf("%d", reply.Relay.SSHPort) + "\n"
	relay += "  MaxSessions        : " + fmt.Sprintf("%d", reply.Relay.MaxSessionCount) + "\n"
	relay += "  MRC                : " + fmt.Sprintf("%4.2f", reply.Relay.MRC.ToDollars()) + "\n"
	relay += "  Overage            : " + fmt.Sprintf("%4.2f", reply.Relay.Overage.ToDollars()) + "\n"
	relay += "  BWRule             : " + fmt.Sprintf("%s", bwRule) + "\n"
	relay += "  ContractTerm       : " + fmt.Sprintf("%d", reply.Relay.ContractTerm) + "\n"
	relay += "  StartDate          : " + startDate + "\n"
	relay += "  EndDate            : " + endDate + "\n"
	relay += "  Type               : " + fmt.Sprintf("%s", machineType) + "\n"
	relay += "  Version            : " + reply.Relay.Version + "\n"
	relay += "  Notes:\n" + reply.Relay.Notes + "\n"

	fmt.Printf("%s\n", relay)

	return

}

func checkForRelay(env Environment, regex string) (uint64, bool) {
	args := localjsonrpc.RelaysArgs{
		Regex: regex,
	}

	var reply localjsonrpc.RelaysReply
	err := makeRPCCall(env, &reply, "OpsService.Relays", args)
	if err != nil {
		handleJSONRPCError(env, err)
		return 0, false
	}

	if len(reply.Relays) == 0 {
		handleRunTimeError(fmt.Sprintf("Zero relays found matching '%s'", regex), 0)
	} else if len(reply.Relays) > 1 {
		handleRunTimeError(fmt.Sprintf("'%s' returned more than one relay - please be more specific.", regex), 0)
	}

	return reply.Relays[0].ID, true

}
