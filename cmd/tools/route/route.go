/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/networknext/backend/routing"
)

func FindRelayByName(routeMatrix *routing.RouteMatrix, relayName string) int {
	for i := range routeMatrix.RelayNames {
		if routeMatrix.RelayNames[i] == relayName {
			return i
		}
	}
	return -1
}

func FindRelayByID(routeMatrix *routing.RouteMatrix, relayID uint64) int {
	for i := range routeMatrix.RelayIDs {
		if routeMatrix.RelayIDs[i] == relayID {
			return i
		}
	}
	return -1
}

func GetDatacenterIndex(routeMatrix *routing.RouteMatrix, datacenterName string) int {
	for i := range routeMatrix.DatacenterNames {
		if routeMatrix.DatacenterNames[i] == datacenterName {
			return i
		}
	}
	return -1
}

/*
func RelayNamesString(relayIDs []uint64, relayIDToIndex map[uint64]int, relayNames []string) string {
	if len(relayIDs) == 0 {
		panic("must have at least one entry in relay ids")
	}

	relays := make([]string, len(relayIDs))
	for i, v := range relayIDs {
		index := relayIDToIndex[v]
		relays[i] = relayNames[index]
	}

	return strings.Join(relays, " - ")
}
*/

func main() {
	relay := flag.String("relay", "", "name of the relay")
	datacenter := flag.String("datacenter", "", "name of the relay")
	flag.Parse()

	var routeMatrix routing.RouteMatrix
	_, err := routeMatrix.ReadFrom(os.Stdin)
	if err != nil {
		log.Fatalln(fmt.Errorf("error reading route matrix from stdin: %w", err))
	}

	relayName := *relay
	datacenterName := *datacenter

	relayIndex := FindRelayByName(&routeMatrix, relayName)

	if relayIndex == -1 {
		log.Fatalf("error: can't find relay called '%s'\n", relayName)
	}

	datacenterIndex := GetDatacenterIndex(&routeMatrix, datacenterName)

	if datacenterIndex == -1 {
		log.Fatalf("\nerror: can't find datacenter called '%s'\n\n", datacenterName)
	}

	datacenterID := routeMatrix.DatacenterIDs[datacenterIndex]

	datacenterRelays := routeMatrix.DatacenterRelays[datacenterID]

	fmt.Printf("%d relays in datacenter\n", len(datacenterRelays))

	for i := range datacenterRelays {

		destRelayID := datacenterRelays[i]

		destRelayIndex := FindRelayByID(&routeMatrix, destRelayID)

		if destRelayIndex == -1 {
			log.Fatalln("WTF!")
		}

		destRelayName := routeMatrix.RelayNames[destRelayIndex]

		fmt.Printf("%s -> %s\n", relayName, destRelayName)
	}

	/*
		a_id := core.RelayID(a)
		b_id := core.RelayID(b)

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

		numRelays := len(route_matrix_rtt.RelayIDs)
		relayIDToIndex := make(map[uint64]int)
		for i := 0; i < numRelays; i++ {
			id := route_matrix_rtt.RelayIDs[i]
			relayIDToIndex[id] = i
		}

		relayNames := route_matrix_rtt.RelayNames

		routes_rtt := core.GetRoutesBetweenRelays(a_id, b_id, route_matrix_rtt, relayIDToIndex)
		routes_jitter := core.GetRoutesBetweenRelays(a_id, b_id, route_matrix_jitter, relayIDToIndex)
		routes_packet_loss := core.GetRoutesBetweenRelays(a_id, b_id, route_matrix_packet_loss, relayIDToIndex)

		if len(routes_rtt) > 1 {
			fmt.Printf("RTT Reducing Routes:\n\n")
			for i := range routes_rtt {
				route := routes_rtt[i]
				fmt.Printf("    %s (%.2fms)\n", RelayNamesString(route.RelayIDs, relayIDToIndex, relayNames), entry.DirectRTT-route.RTT)
			}
			fmt.Printf("\n")
		}

		if len(routes_jitter) > 1 {
			fmt.Printf("Jitter Reducing Routes:\n\n")
			for i := range routes_jitter {
				route := routes_jitter[i]
				fmt.Printf("    %s (%.2fms)\n", RelayNamesString(route.RelayIDs, relayIDToIndex, relayNames), entry.DirectJitter-route.Jitter)
			}
			fmt.Printf("\n")
		}

		if len(routes_packet_loss) > 1 {
			fmt.Printf("Packet Loss Reducing Routes:\n\n")
			for i := range routes_packet_loss {
				route := routes_packet_loss[i]
				fmt.Printf("    %s (%.2f%%)\n", RelayNamesString(route.RelayIDs, relayIDToIndex, relayNames), entry.DirectPacketLoss-route.PacketLoss)
			}
			fmt.Printf("\n")
		}
	*/
}
