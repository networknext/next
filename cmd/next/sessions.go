package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/modood/table"
	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
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
	sessionsByBuyer(rpcClient, env, "", sessionCount)
}

func sessionsByBuyer(rpcClient jsonrpc.RPCClient, env Environment, buyerName string, sessionCount int64) {

	buyerArgs := localjsonrpc.BuyersArgs{}

	var buyersReply localjsonrpc.BuyersReply
	if err := rpcClient.CallFor(&buyersReply, "OpsService.Buyers", buyerArgs); err != nil {
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
	if err := rpcClient.CallFor(&topSessionsReply, "BuyersService.TopSessions", topSessionArgs); err != nil {
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

func dumpSession(rpcClient jsonrpc.RPCClient, env Environment, sessionID uint64) {

	arg := localjsonrpc.GetAllSessionBillingInfoArg{
		SessionID: sessionID,
	}

	var reply localjsonrpc.GetAllSessionBillingInfoReply

	if err := rpcClient.CallFor(&reply, "BuyersService.GetAllSessionBillingInfo", arg); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	bqBillingDataEntryCSV := [][]string{{}}

	bqBillingDataEntryCSV = append(bqBillingDataEntryCSV, []string{
		"SliceNumber",
		"Timestamp",
		"SessionID",
		"Datacenter",
		"UserHash",
		"Latitude",
		"Longitude",
		"ISP",
		"ConnectionType",
		"PlatformType",
		"SdkVersion",
		"DirectRTT",
		"NextRTT",
		"PredictedNextRTT",
		"DirectJitter",
		"NextJitter",
		"DirectPacketLoss",
		"NextPacketLoss",
		"RouteDiversity",
		"NextRelays",
		"NextRelaysPrice",
		"TotalPrice",
		"NextBytesUp",
		"NextBytesDown",
		"EnvelopeBytesUp",
		"EnvelopeBytesDown",
		"ClientToServerPacketsLost",
		"ServerToClientPacketsLost",
		"PacketsOutOfOrderClientToServer",
		"PacketsOutOfOrderServerToClient",
		"JitterClientToServer",
		"JitterServerToClient",
		"ClientFlags",
		"UserFlags",
		"NearRelayRTT",
		"NumNearRelays",
		"NearRelayIDs",
		"NearRelayRTTs",
		"NearRelayJitters",
		"NearRelayPacketLosses",
		"Tags",
		"ABTest",
		"Next",
		"Committed",
		"Flagged",
		"Multipath",
		"RttReduction",
		"PacketLossReduction",
		"FallbackToDirect",
		"Mispredicted",
		"Vetoed",
		"MultipathVetoed",
		"LatencyWorse",
		"Pro",
		"LackOfDiversity",
		"NoRoute",
		"NextLatencyTooHigh",
		"RouteChanged",
		"CommitVeto",
		"RelayWentAway",
		"RouteLost",
		"Debug String",
	})

	for _, billingEntry := range reply.SessionBillingInfo {
		// Timestamp
		timeStamp := billingEntry.Timestamp.String()
		// BuyerString
		// buyerName := billingEntry.BuyerString
		// SessionID
		sessionID := fmt.Sprintf("%016x", uint64(billingEntry.SessionID))
		// SliceNumber
		sliceNumber := fmt.Sprintf("%d", billingEntry.SliceNumber)
		// Next
		next := ""
		if billingEntry.Next.Valid {
			if billingEntry.Next.Bool {
				next = "true"
			}
		}
		// DirectRTT
		directRTT := fmt.Sprintf("%5.5f", billingEntry.DirectRTT)
		// DirectJitter
		directJitter := fmt.Sprintf("%5.5f", billingEntry.DirectJitter)
		// DirectPacketLoss
		directPacketLoss := fmt.Sprintf("%5.5f", billingEntry.DirectPacketLoss)
		// NextRTT
		nextRTT := ""
		if billingEntry.NextRTT.Valid {
			nextRTT = fmt.Sprintf("%5.5f", billingEntry.NextRTT.Float64)
		}
		// NextJitter
		nextJitter := ""
		if billingEntry.NextJitter.Valid {
			nextJitter = fmt.Sprintf("%5.5f", billingEntry.NextJitter.Float64)
		}
		// NextPacketLoss
		nextPacketLoss := ""
		if billingEntry.NextPacketLoss.Valid {
			nextPacketLoss = fmt.Sprintf("%5.5f", billingEntry.NextPacketLoss.Float64)
		}
		// NextRelays
		nextRelays := ""
		if len(billingEntry.NextRelays) > 0 {
			for _, relay := range billingEntry.NextRelaysStrings {
				nextRelays += relay + ", "
			}
			nextRelays = strings.TrimSuffix(nextRelays, ", ")
		}
		// TotalPrice
		totalPrice := fmt.Sprintf("%d", billingEntry.TotalPrice)
		// ClientToServerPacketsLost
		clientToServerPacketsLost := ""
		if billingEntry.ClientToServerPacketsLost.Valid {
			clientToServerPacketsLost = fmt.Sprintf("%d", billingEntry.ClientToServerPacketsLost.Int64)
		}
		// ServerToClientPacketsLost
		serverToClientPacketsLost := ""
		if billingEntry.ClientToServerPacketsLost.Valid {
			serverToClientPacketsLost = fmt.Sprintf("%d", billingEntry.ClientToServerPacketsLost.Int64)
		}
		// Committed
		committed := ""
		if billingEntry.Committed.Valid {
			if billingEntry.Next.Bool {
				committed = "true"
			}
		}
		// Flagged
		flagged := ""
		if billingEntry.Flagged.Bool {
			flagged = "true"
		}
		// Multipath
		multipath := ""
		if billingEntry.Multipath.Bool {
			multipath = "true"
		}
		// NextBytesUp
		nextBytesUp := ""
		if billingEntry.NextBytesUp.Valid {
			nextBytesUp = fmt.Sprintf("%d", billingEntry.NextBytesUp.Int64)
		}
		// NextBytesDown
		nextBytesDown := ""
		if billingEntry.NextBytesDown.Valid {
			nextBytesDown = fmt.Sprintf("%d", billingEntry.NextBytesDown.Int64)
		}
		// DatacenterString
		datacenterName := ""
		if billingEntry.DatacenterString.Valid {
			datacenterName = billingEntry.DatacenterString.StringVal
		}
		// RttReduction
		rttReduction := ""
		if billingEntry.RttReduction.Bool {
			rttReduction = "true"
		}
		// PacketLossReduction
		plReduction := ""
		if billingEntry.PacketLossReduction.Bool {
			plReduction = "true"
		}
		// NextRelaysPrice
		nextRelaysPrice := ""
		if len(billingEntry.NextRelaysPrice) > 0 {
			nextRelaysPrice += fmt.Sprintf("%d", billingEntry.NextRelaysPrice[0])
		}
		// UserHash
		userHash := ""
		if billingEntry.UserHash.Valid {
			userHash = fmt.Sprintf("%016x", billingEntry.UserHash.Int64)
		}
		// Latitude
		latitude := ""
		if billingEntry.Latitude.Valid {
			latitude = fmt.Sprintf("%3.2f", billingEntry.Latitude.Float64)
		}
		// Longitude
		longitude := ""
		if billingEntry.Longitude.Valid {
			longitude = fmt.Sprintf("%3.2f", billingEntry.Longitude.Float64)
		}
		// ISP
		isp := ""
		if billingEntry.ISP.Valid {
			isp = billingEntry.ISP.StringVal
		}
		// ABTest
		abTest := ""
		if billingEntry.ABTest.Bool {
			abTest = "true"
		}
		// ConnectionType
		connType := ""
		if billingEntry.ConnectionType.Valid {
			switch billingEntry.ConnectionType.Int64 {
			case 0:
				connType = "Unknown"
			case 1:
				connType = "Wired"
			case 2:
				connType = "Wi-Fi"
			case 3:
				connType = "Cellular"
			default:
				connType = "none specified?"
			}
		}
		// PlatformType
		platformType := ""
		if billingEntry.PlatformType.Valid {
			switch billingEntry.PlatformType.Int64 {
			case 0:
				platformType = "Unknown"
			case 1:
				platformType = "Windows"
			case 2:
				platformType = "Mac"
			case 3:
				platformType = "Linux"
			case 4:
				platformType = "Nintendo Switch"
			case 5:
				platformType = "PS4"
			case 6:
				platformType = "IOS"
			case 7:
				platformType = "XBox One"
			case 8:
				platformType = "XBox Series X"
			case 9:
				platformType = "PS5"
			default:
				platformType = "none specified?"
			}
		}
		// SdkVersion
		sdkVersion := ""
		if billingEntry.SdkVersion.Valid {
			sdkVersion = billingEntry.SdkVersion.StringVal
		}
		// EnvelopeBytesUp
		envelopeBytesUp := ""
		if billingEntry.EnvelopeBytesUp.Valid {
			envelopeBytesUp = fmt.Sprintf("%d", billingEntry.EnvelopeBytesUp.Int64)
		}
		// EnvelopeBytesDown
		envelopeBytesDown := ""
		if billingEntry.EnvelopeBytesDown.Valid {
			envelopeBytesDown = fmt.Sprintf("%d", billingEntry.EnvelopeBytesDown.Int64)
		}
		// PredictedNextRTT
		predictedNextRTT := ""
		if billingEntry.PredictedNextRTT.Valid {
			predictedNextRTT = fmt.Sprintf("%5.2f", billingEntry.PredictedNextRTT.Float64)
		}
		// MultipathVetoed
		multipathVetoed := ""
		if billingEntry.MultipathVetoed.Bool {
			multipathVetoed = "true"
		}
		// Debug
		debug := ""
		if billingEntry.Debug.Valid {
			debug = billingEntry.Debug.StringVal
		}
		// FallbackToDirect
		fallbackToDirect := ""
		if billingEntry.FallbackToDirect.Bool {
			fallbackToDirect = "true"
		}
		// ClientFlags
		clientFlags := ""
		if billingEntry.ClientFlags.Valid {
			clientFlags = "0b" + strconv.FormatInt(billingEntry.ClientFlags.Int64, 2)
		}
		// UserFlags
		userFlags := ""
		if billingEntry.UserFlags.Valid {
			userFlags = "0b" + strconv.FormatInt(billingEntry.UserFlags.Int64, 2)
		}
		// NearRelayRTT
		nearRelayRTT := ""
		if billingEntry.NearRelayRTT.Valid {
			nearRelayRTT = fmt.Sprintf("%5.2f", billingEntry.NearRelayRTT.Float64)
		}
		// PacketsOutOfOrderClientToServer
		packetsOutOfOrderClientToServer := ""
		if billingEntry.PacketsOutOfOrderClientToServer.Valid {
			packetsOutOfOrderClientToServer = fmt.Sprintf("%d", billingEntry.PacketsOutOfOrderClientToServer.Int64)
		}
		// PacketsOutOfOrderServerToClient
		packetsOutOfOrderServerToClient := ""
		if billingEntry.PacketsOutOfOrderServerToClient.Valid {
			packetsOutOfOrderServerToClient = fmt.Sprintf("%d", billingEntry.PacketsOutOfOrderServerToClient.Int64)
		}
		// JitterClientToServer
		jitterClientToServer := ""
		if billingEntry.JitterClientToServer.Valid {
			jitterClientToServer = fmt.Sprintf("%5.2f", billingEntry.JitterClientToServer.Float64)
		}
		// JitterServerToClient
		jitterServerToClient := ""
		if billingEntry.JitterServerToClient.Valid {
			jitterServerToClient = fmt.Sprintf("%5.2f", billingEntry.JitterServerToClient.Float64)
		}
		// NumNearRelays
		numNearRelays := ""
		if billingEntry.NumNearRelays.Valid {
			numNearRelays = fmt.Sprintf("%d", billingEntry.NumNearRelays.Int64)
		}
		// NearRelayIDs
		nearRelayIDs := ""
		if len(billingEntry.NearRelayIDs) > 0 {
			for _, relayID := range billingEntry.NearRelayIDs {
				nearRelayIDs += fmt.Sprintf("%016x", uint64(relayID)) + ", "
			}
			nearRelayIDs = strings.TrimSuffix(nearRelayIDs, ", ")
		}
		// NearRelayRTTs
		nearRelayRTTs := ""
		if len(billingEntry.NearRelayRTTs) > 0 {
			for _, relayID := range billingEntry.NearRelayRTTs {
				nearRelayRTTs += fmt.Sprintf("%5.5f", relayID) + ", "
			}
			nearRelayRTTs = strings.TrimSuffix(nearRelayRTTs, ", ")
		}
		// NearRelayJitters
		nearRelayJitters := ""
		if len(billingEntry.NearRelayJitters) > 0 {
			for _, relayID := range billingEntry.NearRelayJitters {
				nearRelayJitters += fmt.Sprintf("%5.5f", relayID) + ", "
			}
			nearRelayJitters = strings.TrimSuffix(nearRelayJitters, ", ")
		}
		// NearRelayPacketLosses
		nearRelayPacketLosses := ""
		if len(billingEntry.NearRelayPacketLosses) > 0 {
			for _, relayID := range billingEntry.NearRelayPacketLosses {
				nearRelayPacketLosses += fmt.Sprintf("%5.5f", relayID) + ", "
			}
			nearRelayPacketLosses = strings.TrimSuffix(nearRelayPacketLosses, ", ")
		}
		// RelayWentAway
		relayWentAway := ""
		if billingEntry.RelayWentAway.Bool {
			relayWentAway = "true"
		}
		// RouteLost
		routeLost := ""
		if billingEntry.RouteLost.Bool {
			routeLost = "true"
		}
		// Tags
		tags := ""
		if len(billingEntry.Tags) > 0 {
			for _, tag := range billingEntry.Tags {
				tags += fmt.Sprintf("%016x", uint64(tag)) + ", "
			}
			tags = strings.TrimSuffix(tags, ", ")
		}
		// Mispredicted
		mispredicted := ""
		if billingEntry.Mispredicted.Bool {
			mispredicted = "true"
		}
		// Vetoed
		vetoed := ""
		if billingEntry.Vetoed.Bool {
			vetoed = "true"
		}
		// LatencyWorse
		latencyWorse := ""
		if billingEntry.LatencyWorse.Bool {
			latencyWorse = "true"
		}
		// NoRoute
		noRoute := ""
		if billingEntry.NoRoute.Bool {
			noRoute = "true"
		}
		// NextLatencyTooHigh
		nextLatencyTooHigh := ""
		if billingEntry.NextLatencyTooHigh.Bool {
			nextLatencyTooHigh = "true"
		}
		// RouteChanged
		routeChanged := ""
		if billingEntry.RouteChanged.Bool {
			routeChanged = "true"
		}
		// CommitVeto
		commitVeto := ""
		if billingEntry.CommitVeto.Bool {
			commitVeto = "true"
		}
		// Pro
		pro := ""
		if billingEntry.Pro.Bool {
			pro = "true"
		}
		// LackOfDiversity
		lackOfDiversity := ""
		if billingEntry.LackOfDiversity.Bool {
			lackOfDiversity = "true"
		}
		// RouteDiversity
		routeDiversity := ""
		if billingEntry.RouteDiversity.Valid {
			routeDiversity = fmt.Sprintf("%d", billingEntry.RouteDiversity.Int64)
		}

		bqBillingDataEntryCSV = append(bqBillingDataEntryCSV, []string{
			sliceNumber,
			timeStamp,
			sessionID,
			datacenterName,
			userHash,
			latitude,
			longitude,
			isp,
			connType,
			platformType,
			sdkVersion,
			directRTT,
			nextRTT,
			predictedNextRTT,
			directJitter,
			nextJitter,
			directPacketLoss,
			nextPacketLoss,
			routeDiversity,
			nextRelays,
			nextRelaysPrice,
			totalPrice,
			nextBytesUp,
			nextBytesDown,
			envelopeBytesUp,
			envelopeBytesDown,
			clientToServerPacketsLost,
			serverToClientPacketsLost,
			packetsOutOfOrderClientToServer,
			packetsOutOfOrderServerToClient,
			jitterClientToServer,
			jitterServerToClient,
			clientFlags,
			userFlags,
			nearRelayRTT,
			numNearRelays,
			nearRelayIDs,
			nearRelayRTTs,
			nearRelayJitters,
			nearRelayPacketLosses,
			tags,
			abTest,
			next,
			committed,
			flagged,
			multipath,
			rttReduction,
			plReduction,
			fallbackToDirect,
			mispredicted,
			vetoed,
			multipathVetoed,
			latencyWorse,
			pro,
			lackOfDiversity,
			noRoute,
			nextLatencyTooHigh,
			routeChanged,
			commitVeto,
			relayWentAway,
			routeLost,
			debug,
		})
	}

	fileName := "./session-" + fmt.Sprintf("%016x", sessionID) + ".csv"
	f, err := os.Create(fileName)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("Error creating local CSV file %s: %v\n", fileName, err), 1)
	}

	writer := csv.NewWriter(f)
	err = writer.WriteAll(bqBillingDataEntryCSV)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("Error writing local CSV file %s: %v\n", fileName, err), 1)
	}
	fmt.Println("CSV file written: ", fileName)
	return

}
