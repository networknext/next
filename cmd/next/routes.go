package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/modood/table"
	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func routes(rpcClient jsonrpc.RPCClient, env Environment, srcrelays []string, destrelays []string, rtt float64, routehash uint64) {
	args := localjsonrpc.RouteSelectionArgs{
		SourceRelays:      srcrelays,
		DestinationRelays: destrelays,
		RTT:               rtt,
		RouteHash:         routehash,
	}

	var reply localjsonrpc.RouteSelectionReply
	if err := rpcClient.CallFor(&reply, "OpsService.RouteSelection", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	for _, route := range reply.Routes {
		fmt.Printf("RTT(%v) ", route.Stats.RTT)
		for _, relayName := range route.RelayNames {
			fmt.Print(relayName, " ")
		}
		fmt.Println()
	}
}

func viewCostMatrix(inputFile string) {
	file, err := os.Open(inputFile)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open the cost matrix file for reading: %v\n", err), 1)
	}
	defer file.Close()

	var costMatrix routing.CostMatrix
	if _, err := costMatrix.ReadFrom(file); err != nil {
		handleRunTimeError(fmt.Sprintf("error reading cost matrix: %v\n", err), 1)
	}

	var entries []struct {
		FirstRelayName  string
		SecondRelayName string
		RTT             int32
	}
	for i := 0; i < len(costMatrix.RelayIDs); i++ {
		firstRelayName := costMatrix.RelayNames[i]

		for j := 0; j < i; j++ {
			secondRelayName := costMatrix.RelayNames[j]

			ijIndex := routing.TriMatrixIndex(i, j)
			rtt := costMatrix.Costs[ijIndex]

			entries = append(entries, struct {
				FirstRelayName  string
				SecondRelayName string
				RTT             int32
			}{
				FirstRelayName:  firstRelayName,
				SecondRelayName: secondRelayName,
				RTT:             rtt,
			})
		}
	}

	sort.SliceStable(entries, func(i int, j int) bool {
		if len(entries[i].FirstRelayName) == 0 {
			return true
		}

		if len(entries[j].FirstRelayName) == 0 {
			return true
		}

		return entries[i].FirstRelayName[0] < entries[j].FirstRelayName[0]
	})

	table.Output(entries)

	fmt.Printf("\nListed %d entries\n", len(entries))
}

func viewRouteMatrix(inputFile string, srcRelayNameFilter string, destRelayNameFilter string) {
	file, err := os.Open(inputFile)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open the route matrix file for reading: %v\n", err), 1)
	}
	defer file.Close()

	var routeMatrix routing.RouteMatrix
	if _, err := routeMatrix.ReadFrom(file); err != nil {
		handleRunTimeError(fmt.Sprintf("error reading route matrix: %v\n", err), 1)
	}

	var displayCount int
	for _, entry := range routeMatrix.RouteEntries {
		directRTT := entry.DirectCost
		numRoutes := entry.NumRoutes

		displayEntries := []struct {
			IndirectRTT int32
			Relays      string
		}{}

		if len(entry.RouteNumRelays) == 0 {
			handleRunTimeError(fmt.Sprintln("RouteNumRelays empty or nil"), 0)
		}

		if len(entry.RouteRelays) == 0 {
			handleRunTimeError(fmt.Sprintln("RouteRelays empty or nil"), 0)
		}

		if len(entry.RouteRelays[0]) == 0 {
			handleRunTimeError(fmt.Sprintln("RouteRelays[0] empty or nil"), 0)
		}

		if len(routeMatrix.RelayNames) == 0 {
			handleRunTimeError(fmt.Sprintln("RelayNames empty or nil"), 0)
		}

		srcRelayName := routeMatrix.RelayNames[entry.RouteRelays[0][0]]
		destRelayName := routeMatrix.RelayNames[entry.RouteRelays[0][entry.RouteNumRelays[0]-1]]

		if srcRelayNameFilter != "" && destRelayNameFilter != "" && (srcRelayName != srcRelayNameFilter || destRelayName != destRelayNameFilter) {
			continue
		}

		for i := int32(0); i < numRoutes; i++ {
			rtt := entry.RouteCost[i]
			numRelays := entry.RouteNumRelays[i]

			var relays string
			for j := int32(0); j < numRelays; j++ {
				relayID := entry.RouteRelays[i][j]
				relayName := routeMatrix.RelayNames[relayID]
				relays += relayName + " "
			}

			displayEntries = append(displayEntries, struct {
				IndirectRTT int32
				Relays      string
			}{
				IndirectRTT: rtt,
				Relays:      relays,
			})
		}

		if len(displayEntries) > 0 {
			// Pop off the last entry since it's just the direct route
			displayEntries = displayEntries[:len(displayEntries)-1]
		}

		fmt.Println("Direct RTT:", directRTT)
		fmt.Printf("%s <--> %s\n", srcRelayName, destRelayName)
		table.Output(displayEntries)
		fmt.Println()

		displayCount++
	}

	fmt.Printf("\n%d Entries, showing %d\n", len(routeMatrix.RouteEntries), displayCount)
}
