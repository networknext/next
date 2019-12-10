/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"github.com/networknext/backend/core"
	"io/ioutil"
	"os"
)

func LoadRouteMatrix(filename string) *core.RouteMatrix {
	fmt.Printf("Loading '%s'\n", filename)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("error: could not read %s\n", filename)
		os.Exit(1)
	}
	routeMatrix, err := core.ReadRouteMatrix(data)
	if err != nil {
		fmt.Printf("error: could not read route matrix\n")
		os.Exit(1)
	}
	return routeMatrix
}

func FindRelayByName(routeMatrix *core.RouteMatrix, relayName string) int {
	for i := range routeMatrix.RelayNames {
		if routeMatrix.RelayNames[i] == relayName {
			return i
		}
	}
	return -1
}

func FindRelayById(routeMatrix *core.RouteMatrix, relayId core.RelayCoreID) int {
	for i := range routeMatrix.RelayIds {
		if routeMatrix.RelayIds[i] == relayId {
			return i
		}
	}
	return -1
}

func GetDatacenterIndex(routeMatrix *core.RouteMatrix, datacenterName string) int {
	for i := range routeMatrix.DatacenterNames {
		if routeMatrix.DatacenterNames[i] == datacenterName {
			return i
		}
	}
	return -1
}

/*
func RelayNamesString(relayIds []uint64, relayIdToIndex map[uint64]int, relayNames []string) string {
	if len(relayIds) == 0 {
		panic("must have at least one entry in relay ids")
	}

	relays := make([]string, len(relayIds))
	for i, v := range relayIds {
		index := relayIdToIndex[v]
		relays[i] = relayNames[index]
	}

	return strings.Join(relays, " - ")
}
*/

func main() {

	args := os.Args[1:]

	if len(args) != 2 {
		fmt.Printf("\nUsage: 'next routes [relay] [datacenter]'\n\n")
		return
	}

	fmt.Printf("\nWelcome to Network Next!\n\n")

	routeMatrix := LoadRouteMatrix("optimize.bin")

	relayName := args[0]
	datacenterName := args[1]

	relayIndex := FindRelayByName(routeMatrix, relayName)

	if relayIndex == -1 {
		fmt.Printf("\nerror: can't find relay called '%s'\n\n", relayName)
		os.Exit(1)
	}

	datacenterIndex := GetDatacenterIndex(routeMatrix, datacenterName)

	if datacenterIndex == -1 {
		fmt.Printf("\nerror: can't find datacenter called '%s'\n\n", datacenterName)
		os.Exit(1)
	}

	datacenterId := routeMatrix.DatacenterIds[datacenterIndex]

	relaysInDatacenter := routeMatrix.RelayDatacenters[datacenterId]

	fmt.Printf("%d relays in datacenter\n", len(relaysInDatacenter))

	for i := range relaysInDatacenter {

		destRelayId := relaysInDatacenter[i]

		destRelayIndex := FindRelayById(routeMatrix, core.RelayCoreID(destRelayId))

		if destRelayIndex == -1 {
			panic("WTF!")
		}

		destRelayName := routeMatrix.RelayNames[destRelayIndex]

		fmt.Printf("%s -> %s\n", relayName, destRelayName)
	}

	/*
		a_id := core.RelayId(a)
		b_id := core.RelayId(b)

		abFlatIndex := core.TriMatrixIndex(a_index, b_index)

		entry_rtt := route_matrix_rtt.Entries[abFlatIndex]
		entry_jitter := route_matrix_jitter.Entries[abFlatIndex]
		entry_packet_loss := route_matrix_packet_loss.Entries[abFlatIndex]

		if len(entry_rtt.Routes) == 0 && len(entry_jitter.Routes) == 0 && len(entry_packet_loss.Routes) == 0 {
			fmt.Printf("No routes found!\n\n")
			os.Exit(1)
		}

		entry := entry_rtt
		if len(entry.Routes) == 0 {
			entry = entry_jitter
			if len(entry.Routes) == 0 {
				entry = entry_packet_loss
			}
		}

		fmt.Printf("Direct RTT = %.2f\n", entry.DirectRTT)
		fmt.Printf("Direct Jitter = %.2f\n", entry.DirectJitter)
		fmt.Printf("Direct Packet Loss = %.2f\n\n", entry.DirectPacketLoss)

		numRelays := len(route_matrix_rtt.RelayIds)
		relayIdToIndex := make(map[uint64]int)
		for i := 0; i < numRelays; i++ {
			id := route_matrix_rtt.RelayIds[i]
			relayIdToIndex[id] = i
		}

		relayNames := route_matrix_rtt.RelayNames

		routes_rtt := core.GetRoutesBetweenRelays(a_id, b_id, route_matrix_rtt, relayIdToIndex)
		routes_jitter := core.GetRoutesBetweenRelays(a_id, b_id, route_matrix_jitter, relayIdToIndex)
		routes_packet_loss := core.GetRoutesBetweenRelays(a_id, b_id, route_matrix_packet_loss, relayIdToIndex)

		if len(routes_rtt) > 1 {
			fmt.Printf("RTT Reducing Routes:\n\n")
			for i := range routes_rtt {
				route := routes_rtt[i]
				fmt.Printf("    %s (%.2fms)\n", RelayNamesString(route.RelayIds, relayIdToIndex, relayNames), entry.DirectRTT-route.RTT)
			}
			fmt.Printf("\n")
		}

		if len(routes_jitter) > 1 {
			fmt.Printf("Jitter Reducing Routes:\n\n")
			for i := range routes_jitter {
				route := routes_jitter[i]
				fmt.Printf("    %s (%.2fms)\n", RelayNamesString(route.RelayIds, relayIdToIndex, relayNames), entry.DirectJitter-route.Jitter)
			}
			fmt.Printf("\n")
		}

		if len(routes_packet_loss) > 1 {
			fmt.Printf("Packet Loss Reducing Routes:\n\n")
			for i := range routes_packet_loss {
				route := routes_packet_loss[i]
				fmt.Printf("    %s (%.2f%%)\n", RelayNamesString(route.RelayIds, relayIdToIndex, relayNames), entry.DirectPacketLoss-route.PacketLoss)
			}
			fmt.Printf("\n")
		}
	*/
}
