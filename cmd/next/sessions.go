package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

func flushsessions(rpcClient jsonrpc.RPCClient, env Environment) {
	relaysargs := localjsonrpc.FlushSessionsArgs{}

	var relaysreply localjsonrpc.FlushSessionsReply
	if err := rpcClient.CallFor(&relaysreply, "BuyersService.FlushSessions", relaysargs); err != nil {
		handleJSONRPCError(env, err)
		return
	}
}

func sessions(rpcClient jsonrpc.RPCClient, env Environment, sessionID string, sessionCount int64) {
	if sessionID != "" {
		relaysargs := localjsonrpc.RelaysArgs{}

		var relaysreply localjsonrpc.RelaysReply
		if err := rpcClient.CallFor(&relaysreply, "OpsService.Relays", relaysargs); err != nil {
			handleJSONRPCError(env, err)
			return
		}

		args := localjsonrpc.SessionDetailsArgs{
			SessionID: sessionID,
		}

		var reply localjsonrpc.SessionDetailsReply
		if err := rpcClient.CallFor(&reply, "BuyersService.SessionDetails", args); err != nil {
			handleJSONRPCErrorCustom(env, err, "Session not found")
			return
		}

		stats := []struct {
			Name       string
			RTT        string
			Jitter     string
			PacketLoss string
		}{}

		if len(reply.Slices) == 0 {
			log.Fatalln(fmt.Errorf("session has no slices yet"))
		}

		lastSlice := reply.Slices[len(reply.Slices)-1]

		if reply.Meta.OnNetworkNext {
			fmt.Printf("Session is on Network Next\n\n")
			fmt.Printf("RTT improvement is %.1fms\n\n", lastSlice.Direct.RTT-lastSlice.Next.RTT)
		} else {
			fmt.Printf("Session is going direct\n\n")
		}

		stats = append(stats, struct {
			Name       string
			RTT        string
			Jitter     string
			PacketLoss string
		}{
			Name:       "Direct",
			RTT:        fmt.Sprintf("%.01f", lastSlice.Direct.RTT),
			Jitter:     fmt.Sprintf("%.01f", lastSlice.Direct.Jitter),
			PacketLoss: fmt.Sprintf("%.01f", lastSlice.Direct.PacketLoss),
		})

		if reply.Meta.OnNetworkNext {
			stats = append(stats, struct {
				Name       string
				RTT        string
				Jitter     string
				PacketLoss string
			}{
				Name:       "Next",
				RTT:        fmt.Sprintf("%.01f", lastSlice.Next.RTT),
				Jitter:     fmt.Sprintf("%.01f", lastSlice.Next.Jitter),
				PacketLoss: fmt.Sprintf("%.01f", lastSlice.Next.PacketLoss),
			})
		}

		table.Output(stats)

		// todo: why are near relays not sent down for direct sessions? they should be...

		if len(reply.Meta.NearbyRelays) != 0 {

			fmt.Printf("\nNear Relays:\n")

			near := []struct {
				Name       string
				RTT        string
				Jitter     string
				PacketLoss string
			}{}

			for _, relay := range reply.Meta.NearbyRelays {
				for _, r := range relaysreply.Relays {
					if relay.ID == r.ID {
						relay.Name = r.Name
					}
				}
				near = append(near, struct {
					Name       string
					RTT        string
					Jitter     string
					PacketLoss string
				}{
					Name:       relay.Name,
					RTT:        fmt.Sprintf("%.1f", relay.ClientStats.RTT),
					Jitter:     fmt.Sprintf("%.1f", relay.ClientStats.Jitter),
					PacketLoss: fmt.Sprintf("%.1f", relay.ClientStats.PacketLoss),
				})
			}

			table.Output(near)
		}

		fmt.Printf("\nCurrent Route:\n\n")

		cost := int(math.Ceil(float64(reply.Meta.DirectRTT)))
		if reply.Meta.OnNetworkNext {
			cost = int(reply.Meta.NextRTT)
		}

		fmt.Printf("    %*dms: ", 5, cost)

		if reply.Meta.OnNetworkNext {
			for index, hop := range reply.Meta.Hops {
				for _, relay := range relaysreply.Relays {
					if hop.ID == relay.ID {
						hop.Name = relay.Name
					}
				}
				if index != 0 {
					fmt.Printf(" - %s", hop.Name)
				} else {
					fmt.Printf("%s", hop.Name)
				}
			}
			fmt.Printf("\n")
		} else {
			fmt.Printf("direct\n")
		}

		// =======================================================

		if len(reply.Meta.NearbyRelays) == 0 {
			return
		}

		// todo: want the datacenter id directly, without going through hops. lets us check available routes even for direct

		if len(reply.Meta.Hops) == 0 {
			return
		}

		type AvailableRoute struct {
			cost   int
			relays string
		}

		availableRoutes := make([]AvailableRoute, 0)

		// todo: get datacenter for relay. iterate across all relays in datacenter

		destRelayId := reply.Meta.Hops[len(reply.Meta.Hops)-1].ID

		file, err := os.Open("optimize.bin")
		if err != nil {
			return
		}
		defer file.Close()

		var routeMatrix routing.RouteMatrix
		if _, err := routeMatrix.ReadFrom(file); err != nil {
			log.Fatalln(fmt.Errorf("error reading route matrix: %w", err))
		}

		numRelays := len(routeMatrix.RelayIDs)

		relays := make([]RelayEntry, numRelays)
		for i := 0; i < numRelays; i++ {
			relays[i].id = routeMatrix.RelayIDs[i]
			relays[i].name = routeMatrix.RelayNames[i]
		}

		destRelayIndex, ok := routeMatrix.RelayIndices[destRelayId]
		if !ok {
			log.Fatalln(fmt.Errorf("dest relay %x not in matrix", destRelayId))
		}

		for _, relay := range reply.Meta.NearbyRelays {

			sourceRelayId := relay.ID

			if sourceRelayId == destRelayId {
				continue
			}

			sourceRelayIndex, ok := routeMatrix.RelayIndices[sourceRelayId]
			if !ok {
				log.Fatalln(fmt.Errorf("source relay %x not in matrix", sourceRelayId))
			}

			nearRelayRTT := relay.ClientStats.RTT

			index := routing.TriMatrixIndex(sourceRelayIndex, destRelayIndex)

			numRoutes := int(routeMatrix.Entries[index].NumRoutes)

			for i := 0; i < numRoutes; i++ {
				routeRTT := routeMatrix.Entries[index].RouteRTT[i]
				routeNumRelays := int(routeMatrix.Entries[index].RouteNumRelays[i])
				routeCost := int(nearRelayRTT + float64(routeRTT))
				if routeCost >= int(lastSlice.Direct.RTT) {
					continue
				}
				var availableRoute AvailableRoute
				availableRoute.cost = routeCost
				reverse := sourceRelayIndex < destRelayIndex
				if reverse {
					for j := routeNumRelays - 1; j >= 0; j-- {
						availableRoute.relays += routeMatrix.RelayNames[routeMatrix.Entries[index].RouteRelays[i][j]]
						if j != 0 {
							availableRoute.relays += " - "
						}
					}
				} else {
					for j := 0; j < routeNumRelays; j++ {
						availableRoute.relays += routeMatrix.RelayNames[routeMatrix.Entries[index].RouteRelays[i][j]]
						if j != routeNumRelays-1 {
							availableRoute.relays += (" - ")
						}
					}
				}
				availableRoutes = append(availableRoutes, availableRoute)
			}
		}

		fmt.Printf("\nAvailable Routes:\n\n")

		sort.SliceStable(availableRoutes[:], func(i, j int) bool { return availableRoutes[i].cost < availableRoutes[j].cost })

		for i := range availableRoutes {
			fmt.Printf("    %*dms: %s\n", 5, availableRoutes[i].cost, availableRoutes[i].relays)
		}

		// =======================================================

		return
	}

	args := localjsonrpc.TopSessionsArgs{}

	var reply localjsonrpc.TopSessionsReply
	if err := rpcClient.CallFor(&reply, "BuyersService.TopSessions", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	sessions := []struct {
		ID          string
		UserHash    string
		ISP         string
		Datacenter  string
		DirectRTT   string
		NextRTT     string
		Improvement string
	}{}

	for _, session := range reply.Sessions {
		directRTT := fmt.Sprintf("%.02f", session.DirectRTT)
		if session.DirectRTT == 0 {
			directRTT = "-"
		}
		nextRTT := fmt.Sprintf("%.02f", session.NextRTT)
		if session.NextRTT == 0 {
			nextRTT = "-"
		}
		improvement := fmt.Sprintf("%.02f", session.DeltaRTT)
		if nextRTT == "-" || directRTT == "-" {
			improvement = "-"
		}
		sessions = append(sessions, struct {
			ID          string
			UserHash    string
			ISP         string
			Datacenter  string
			DirectRTT   string
			NextRTT     string
			Improvement string
		}{
			ID:          session.ID,
			UserHash:    session.UserHash,
			ISP:         fmt.Sprintf("%.32s", session.Location.ISP),
			Datacenter:  session.DatacenterName,
			DirectRTT:   directRTT,
			NextRTT:     nextRTT,
			Improvement: improvement,
		})
	}

	if sessionCount > 0 {
		table.Output(sessions[0:sessionCount])
	} else {
		table.Output(sessions)
	}

}
