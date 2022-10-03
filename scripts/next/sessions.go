package main

import (
    "fmt"
    "math"
    "os"
    "regexp"
    "sort"

    "github.com/modood/table"

    "github.com/networknext/backend/modules-old/routing"

    localjsonrpc "github.com/networknext/backend/modules-old/transport/jsonrpc"
)

func flushsessions(env Environment) {
    relaysargs := localjsonrpc.FlushSessionsArgs{}

    var relaysreply localjsonrpc.FlushSessionsReply
    if err := makeRPCCall(env, &relaysreply, "BuyersService.FlushSessions", relaysargs); err != nil {
        handleJSONRPCError(env, err)
        return
    }
}

func sessions(env Environment, sessionID string, sessionCount int64) {
    if sessionID != "" {
        args := localjsonrpc.SessionDetailsArgs{
            SessionID: sessionID,
        }

        var reply localjsonrpc.SessionDetailsReply
        if err := makeRPCCall(env, &reply, "BuyersService.SessionDetails", args); err != nil {
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
            handleRunTimeError(fmt.Sprintln("session has no slices yet"), 0)
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
            RTT:        fmt.Sprintf("%.02f", lastSlice.Direct.RTT),
            Jitter:     fmt.Sprintf("%.02f", lastSlice.Direct.Jitter),
            PacketLoss: fmt.Sprintf("%.02f", lastSlice.Direct.PacketLoss),
        })

        if reply.Meta.OnNetworkNext {
            stats = append(stats, struct {
                Name       string
                RTT        string
                Jitter     string
                PacketLoss string
            }{
                Name:       "Next",
                RTT:        fmt.Sprintf("%.02f", lastSlice.Next.RTT),
                Jitter:     fmt.Sprintf("%.02f", lastSlice.Next.Jitter),
                PacketLoss: fmt.Sprintf("%.02f", lastSlice.Next.PacketLoss),
            })
        }

        table.Output(stats)

        if len(reply.Meta.NearbyRelays) != 0 {

            fmt.Printf("\nNear Relays:\n")

            near := []struct {
                Name       string
                RTT        string
                Jitter     string
                PacketLoss string
            }{}

            for _, relay := range reply.Meta.NearbyRelays {
                near = append(near, struct {
                    Name       string
                    RTT        string
                    Jitter     string
                    PacketLoss string
                }{
                    Name:       relay.Name,
                    RTT:        fmt.Sprintf("%.2f", relay.ClientStats.RTT),
                    Jitter:     fmt.Sprintf("%.2f", relay.ClientStats.Jitter),
                    PacketLoss: fmt.Sprintf("%.2f", relay.ClientStats.PacketLoss),
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
            handleRunTimeError(fmt.Sprintf("error reading route matrix: %v\n", err), 1)
        }

        numRelays := len(routeMatrix.RelayIDs)

        relays := make([]RelayEntry, numRelays)
        for i := 0; i < numRelays; i++ {
            relays[i].id = routeMatrix.RelayIDs[i]
            relays[i].name = routeMatrix.RelayNames[i]
        }

        destRelayIndex, ok := routeMatrix.RelayIDsToIndices[destRelayId]
        if !ok {
            handleRunTimeError(fmt.Sprintf("dest relay %x not in matrix\n", destRelayId), 1)
        }

        for _, relay := range reply.Meta.NearbyRelays {

            sourceRelayId := relay.ID

            if sourceRelayId == destRelayId {
                continue
            }

            sourceRelayIndex, ok := routeMatrix.RelayIDsToIndices[sourceRelayId]
            if !ok {
                handleRunTimeError(fmt.Sprintf("source relay %x not in matrix\n", sourceRelayId), 1)
            }

            nearRelayRTT := relay.ClientStats.RTT

            index := routing.TriMatrixIndex(int(sourceRelayIndex), int(destRelayIndex))

            numRoutes := int(routeMatrix.RouteEntries[index].NumRoutes)

            for i := 0; i < numRoutes; i++ {
                routeRTT := routeMatrix.RouteEntries[index].RouteCost[i]
                routeNumRelays := int(routeMatrix.RouteEntries[index].RouteNumRelays[i])
                routeCost := int(nearRelayRTT + float64(routeRTT))
                if routeCost >= int(lastSlice.Direct.RTT) {
                    continue
                }
                var availableRoute AvailableRoute
                availableRoute.cost = routeCost
                reverse := sourceRelayIndex < destRelayIndex
                if reverse {
                    for j := routeNumRelays - 1; j >= 0; j-- {
                        availableRoute.relays += routeMatrix.RelayNames[routeMatrix.RouteEntries[index].RouteRelays[i][j]]
                        if j != 0 {
                            availableRoute.relays += " - "
                        }
                    }
                } else {
                    for j := 0; j < routeNumRelays; j++ {
                        availableRoute.relays += routeMatrix.RelayNames[routeMatrix.RouteEntries[index].RouteRelays[i][j]]
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
    sessionsByBuyer(env, "", sessionCount)
}

func sessionsByBuyer(env Environment, buyerName string, sessionCount int64) {

    buyerArgs := localjsonrpc.BuyersArgs{}

    var buyersReply localjsonrpc.BuyersReply
    if err := makeRPCCall(env, &buyersReply, "OpsService.Buyers", buyerArgs); err != nil {
        handleJSONRPCError(env, err)
        return
    }

    buyers := buyersReply.Buyers
    topSessionArgs := localjsonrpc.TopSessionsArgs{}

    var buyerID uint64
    if len(buyers) > 0 && buyerName != "" {
        r := regexp.MustCompile("(?i)" + buyerName) // case-insensitive regex
        for _, buyer := range buyers {
            if r.MatchString(buyer.CompanyName) {
                topSessionArgs.CompanyCode = buyer.CompanyCode
                buyerID = buyer.ID
                break
            }
        }
    }

    var topSessionsReply localjsonrpc.TopSessionsReply
    if err := makeRPCCall(env, &topSessionsReply, "BuyersService.TopSessions", topSessionArgs); err != nil {
        handleJSONRPCError(env, err)
        return
    }

    if len(topSessionsReply.Sessions) == 0 {
        handleRunTimeError(fmt.Sprintf("No sessions found for buyer ID: %v\n", buyerID), 0)
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

    for _, session := range topSessionsReply.Sessions {
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
            ID:          fmt.Sprintf("%016x", session.ID),
            UserHash:    fmt.Sprintf("%016x", session.UserHash),
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
