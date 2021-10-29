package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"cloud.google.com/go/bigquery"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
	"google.golang.org/api/iterator"
)

func dumpSession(env Environment, sessionID uint64) {

	// make a call for all the relays (there is no relay "singular" endpoint)
	var relaysReply localjsonrpc.RelaysReply
	relaysArg := localjsonrpc.RelaysArgs{} // empty args returns all relays
	if err := makeRPCCall(env, &relaysReply, "OpsService.Relays", relaysArg); err != nil {
		handleJSONRPCError(env, err)
	}

	relayNames := make(map[int64]string)
	for _, relay := range relaysReply.Relays {
		relayNames[int64(relay.ID)] = relay.Name
	}

	// make a call for the datacenters
	var dcsReply localjsonrpc.DatacentersReply
	dcsArgs := localjsonrpc.DatacentersArgs{}
	if err := makeRPCCall(env, &dcsReply, "OpsService.Datacenters", dcsArgs); err != nil {
		handleJSONRPCError(env, err)
	}
	dcNames := make(map[int64]string)
	for _, dc := range dcsReply.Datacenters {
		dcNames[int64(dc.ID)] = dc.Name
	}

	rows, err := GetAllSessionBillingInfo(int64(sessionID), env)

	var newRows []BigQueryBillingEntry

	// process returned resultset - wire up relay and datacenter names
	for index, row := range rows {

		newRows = append(newRows, row)

		newRows[index].DatacenterString = bigquery.NullString{StringVal: dcNames[row.DatacenterID.Int64], Valid: true}

		relayList := ""

		// the BQ sql query enforces the relay ordering
		if row.NextRelay0.Valid {
			relayList = relayNames[row.NextRelay0.Int64]
		}

		if row.NextRelay1.Valid {
			relayList += ", " + relayNames[row.NextRelay1.Int64]
		}

		if row.NextRelay2.Valid {
			relayList += ", " + relayNames[row.NextRelay2.Int64]
		}

		if row.NextRelay3.Valid {
			relayList += ", " + relayNames[row.NextRelay3.Int64]
		}

		if row.NextRelay4.Valid {
			relayList += ", " + relayNames[row.NextRelay4.Int64]
		}

		newRows[index].NextRelaysStrings = relayList
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
		"LackOfDiversity",
		"RouteDiversity",
		"Pro",
		"PacketLoss",
		"Tags",
		"ABTest",
		"Next",
		"Initial",
		"Committed",
		"Flagged",
		"Multipath",
		"MultipathRestricted",
		"RttReduction",
		"PacketLossReduction",
		"FallbackToDirect",
		"Mispredicted",
		"Vetoed",
		"MultipathVetoed",
		"LatencyWorse",
		"NoRoute",
		"NextLatencyTooHigh",
		"RouteChanged",
		"CommitVeto",
		"RelayWentAway",
		"RouteLost",
		"Debug String",
		"ClientToServerPacketsSent",
		"ServerToClientPacketsSent",
		"UnknownDatacenter",
		"DatacenterNotEnabled",
		"BuyerNotLive",
		"StaleRouteMatrix",
	})

	for _, billingEntry := range newRows {
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
		nextRelays := billingEntry.NextRelaysStrings
		// TotalPrice
		totalPrice := ""
		if billingEntry.TotalPrice.Valid {
			totalPrice = fmt.Sprintf("%d", billingEntry.TotalPrice.Int64)
		}
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

		initial := "false"
		if billingEntry.Initial {
			initial = "true"
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
			connType = fmt.Sprintf("%d", billingEntry.ConnectionType.Int64)
		}
		// PlatformType
		platformType := ""
		if billingEntry.PlatformType.Valid {
			platformType = fmt.Sprintf("%d", billingEntry.PlatformType.Int64)
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
		// MultipathRestricted
		multipathRestricted := ""
		if billingEntry.MultipathRestricted.Bool {
			multipathRestricted = "true"
		}
		// ClientToServerPacketsSent
		clientToServerPacketsSent := ""
		if billingEntry.ClientToServerPacketsSent.Valid {
			clientToServerPacketsSent = fmt.Sprintf("%d", billingEntry.ClientToServerPacketsSent.Int64)
		}
		// ServerToClientPacketsSent
		serverToClientPacketsSent := ""
		if billingEntry.ServerToClientPacketsSent.Valid {
			serverToClientPacketsSent = fmt.Sprintf("%d", billingEntry.ServerToClientPacketsSent.Int64)
		}

		// LackOfDiversity,
		lackOfDiversity := ""
		if billingEntry.LackOfDiversity.Valid {
			lackOfDiversity = fmt.Sprintf("%t", billingEntry.LackOfDiversity.Bool)
		}
		// PacketLoss
		packetLoss := ""
		if billingEntry.PacketLoss.Valid {
			packetLoss = fmt.Sprintf("%5.5f", billingEntry.PacketLoss.Float64)
		}
		// Pro
		pro := ""
		if billingEntry.Pro.Valid {
			pro = fmt.Sprintf("%t", billingEntry.Pro.Bool)
		}
		// RouteDiversity
		routeDiversity := ""
		if billingEntry.RouteDiversity.Valid {
			routeDiversity = fmt.Sprintf("%d", billingEntry.RouteDiversity.Int64)
		}

		// UnknownDatacenter
		unknownDatacenter := ""
		if billingEntry.UnknownDatacenter.Valid {
			unknownDatacenter = fmt.Sprintf("%t", billingEntry.UnknownDatacenter.Bool)
		}

		// DatacenterNotEnabled
		datacenterNotEnabled := ""
		if billingEntry.DatacenterNotEnabled.Valid {
			datacenterNotEnabled = fmt.Sprintf("%t", billingEntry.DatacenterNotEnabled.Bool)
		}

		// BuyerNotLive
		buyerNotLive := ""
		if billingEntry.BuyerNotLive.Valid {
			buyerNotLive = fmt.Sprintf("%t", billingEntry.BuyerNotLive.Bool)
		}

		// StaleRouteMatrix
		staleRouteMatrix := ""
		if billingEntry.StaleRouteMatrix.Valid {
			staleRouteMatrix = fmt.Sprintf("%t", billingEntry.StaleRouteMatrix.Bool)
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
			lackOfDiversity,
			routeDiversity,
			pro,
			packetLoss,
			tags,
			abTest,
			next,
			initial,
			committed,
			flagged,
			multipath,
			multipathRestricted,
			rttReduction,
			plReduction,
			fallbackToDirect,
			mispredicted,
			vetoed,
			multipathVetoed,
			latencyWorse,
			noRoute,
			nextLatencyTooHigh,
			routeChanged,
			commitVeto,
			relayWentAway,
			routeLost,
			debug,
			clientToServerPacketsSent,
			serverToClientPacketsSent,
			unknownDatacenter,
			datacenterNotEnabled,
			buyerNotLive,
			staleRouteMatrix,
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

func GetAllSessionBillingInfo(sessionID int64, env Environment) ([]BigQueryBillingEntry, error) {

	ctx := context.Background()

	var rows []BigQueryBillingEntry

	var dbName string
	var sql bytes.Buffer

	sql.Write([]byte(`select 
	timeStamp,
	buyerID,
	sessionID,
	sliceNumber,
	next,
	directRTT,
	directJitter,
	directPacketLoss,
	nextRTT,
	nextJitter,
	nextPacketLoss,
	if(ARRAY_LENGTH(nextRelays)>0, nextRelays[OFFSET(0)], NULL) as NextRelay0,
	if(ARRAY_LENGTH(nextRelays)>1, nextRelays[OFFSET(1)], NULL) as NextRelay1,
	if(ARRAY_LENGTH(nextRelays)>2, nextRelays[OFFSET(2)], NULL) as NextRelay2,
	if(ARRAY_LENGTH(nextRelays)>3, nextRelays[OFFSET(3)], NULL) as NextRelay3,
	if(ARRAY_LENGTH(nextRelays)>4, nextRelays[OFFSET(4)], NULL) as NextRelay4,
	totalPrice,
	clientToServerPacketsLost,
	serverToClientPacketsLost,
	initial,
	committed,
	flagged,
	multipath,
	multipathRestricted,
	nextBytesUp,
	nextBytesDown,
	datacenterID,
	rttReduction,
	packetLossReduction,
	nextRelaysPrice,
	userHash,
	latitude,
	longitude,
	isp,
	abTest,
	connectionType,
	platformType,
	sdkVersion,
	packetLoss,
	envelopeBytesUp,
	envelopeBytesDown,
	predictedNextRTT,
	multipathVetoed,
	debug,
	fallbackToDirect,
	clientFlags,
	userFlags,
	nearRelayRTT,
	packetsOutOfOrderClientToServer,
	packetsOutOfOrderServerToClient,
	jitterClientToServer,
	jitterServerToClient,
	numNearRelays,
	nearRelayIDs,
	nearRelayRTTs,
	nearRelayJitters,
	nearRelayPacketLosses,
	tags,
	relayWentAway,
	routeLost,
	mispredicted,
	vetoed,
	latencyWorse,
	noRoute,
	nextLatencyTooHigh,
	routeChanged,
	commitVeto,
	clientToServerPacketsSent,
	serverToClientPacketsSent,
	lackOfDiversity,
	packetLoss,
	pro,
	routeDiversity,
	unknownDatacenter,
	datacenterNotEnabled,
	buyerNotLive,
	staleRouteMatrix
    from `))

	if env.Name != "prod" && env.Name != "dev" && env.Name != "staging" {
		// env == local (unit test)
		// env == ""    (e.g. go test -run TestGetAllSessionBillingInfo)
		// var err error
		// rows, err = returnLocalTestData(reply)
		// if err != nil {
		// 	err = fmt.Errorf("GetAllSessionBillingInfo() error returning local json: %v", err)
		// 	level.Error(s.Logger).Log("err", err, "GetAllSessionBillingInfo", fmt.Sprintf("%016x", sessionID))
		// 	return err
		// }
		fmt.Println("Local/testing functionality TBD.")
	} else {
		if env.Name == "prod" {
			sql.Write([]byte("network-next-v3-prod.prod.billing"))
			dbName = "network-next-v3-prod"

		} else if env.Name == "dev" {
			sql.Write([]byte("network-next-v3-dev.dev.billing"))
			dbName = "network-next-v3-dev"
		}

		sql.Write([]byte(" where sessionId = "))
		sql.Write([]byte(fmt.Sprintf("%d", sessionID)))
		// a timestamp must be provided although it is not relevant to this query
		sql.Write([]byte(" and DATE(timestamp) >= '1968-05-01'"))
		sql.Write([]byte(" order by sliceNumber asc"))

		bqClient, err := bigquery.NewClient(ctx, dbName)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBillingInfo() failed to create BigQuery client: %v", err), 1)
			return nil, err
		}
		defer bqClient.Close()

		q := bqClient.Query(string(sql.String()))

		job, err := q.Run(ctx)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBillingInfo() failed to query BigQuery: %v", err), 1)
			return nil, err
		}

		status, err := job.Wait(ctx)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBillingInfo() error waiting for job to complete: %v", err), 1)
			return nil, err
		}
		if err := status.Err(); err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBillingInfo() job returned an error: %v", err), 1)
			return nil, err
		}

		it, err := job.Read(ctx)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBillingInfo() job.Read() error: %v", err), 1)
			return nil, err
		}

		// process result set and load rows
		for {
			var rec BigQueryBillingEntry
			err := it.Next(&rec)

			if err == iterator.Done {
				break
			}
			if err != nil {
				handleRunTimeError(fmt.Sprintf("GetAllSessionBillingInfo() BigQuery iterator error: %v", err), 1)
				return nil, err
			}
			rows = append(rows, rec)
		}
	}

	return rows, nil

}

func dumpSession2(env Environment, sessionID uint64) {

	// make a call for all the relays (there is no relay "singular" endpoint)
	var relaysReply localjsonrpc.RelaysReply
	relaysArg := localjsonrpc.RelaysArgs{} // empty args returns all relays
	if err := makeRPCCall(env, &relaysReply, "OpsService.Relays", relaysArg); err != nil {
		handleJSONRPCError(env, err)
	}

	relayNames := make(map[int64]string)
	for _, relay := range relaysReply.Relays {
		relayNames[int64(relay.ID)] = relay.Name
	}

	// make a call for the datacenters
	var dcsReply localjsonrpc.DatacentersReply
	dcsArgs := localjsonrpc.DatacentersArgs{}
	if err := makeRPCCall(env, &dcsReply, "OpsService.Datacenters", dcsArgs); err != nil {
		handleJSONRPCError(env, err)
	}
	dcNames := make(map[int64]string)
	for _, dc := range dcsReply.Datacenters {
		dcNames[int64(dc.ID)] = dc.Name
	}

	rows, err := GetAllSessionBilling2Info(int64(sessionID), env)

	var newRows []BigQueryBilling2Entry

	// process returned resultset - wire up relay and datacenter names
	for index, row := range rows {

		newRows = append(newRows, row)

		newRows[index].DatacenterString = bigquery.NullString{StringVal: dcNames[row.DatacenterID.Int64], Valid: true}

		relayList := ""

		// the BQ sql query enforces the relay ordering
		if row.NextRelay0.Valid {
			relayList = relayNames[row.NextRelay0.Int64]
		}

		if row.NextRelay1.Valid {
			relayList += ", " + relayNames[row.NextRelay1.Int64]
		}

		if row.NextRelay2.Valid {
			relayList += ", " + relayNames[row.NextRelay2.Int64]
		}

		if row.NextRelay3.Valid {
			relayList += ", " + relayNames[row.NextRelay3.Int64]
		}

		if row.NextRelay4.Valid {
			relayList += ", " + relayNames[row.NextRelay4.Int64]
		}

		newRows[index].NextRelaysStrings = relayList
	}

	bqBilling2DataEntryCSV := [][]string{{}}

	bqBilling2DataEntryCSV = append(bqBilling2DataEntryCSV, []string{
		"Timestamp",
		"SessionID",
		"SliceNumber",
		"DirectMinRTT",
		"DirectMaxRTT",
		"DirectPrimeRTT",
		"DirectJitter",
		"DirectPacketLoss",
		"RealPacketLoss",
		"RealJitter",
		"Next",
		"Flagged",
		"Summary",
		"Debug",
		"Datacenter",
		"BuyerID",
		"UserHash",
		"EnvelopeBytesUp",
		"EnvelopeBytesDown",
		"Latitude",
		"Longitude",
		"ClientAddress",
		"ISP",
		"ConnectionType",
		"PlatformType",
		"SdkVersion",
		"Tags",
		"ABTest",
		"Pro",
		"ClientToServerPacketsSent",
		"ServerToClientPacketsSent",
		"ClientToServerPacketsLost",
		"ServerToClientPacketsLost",
		"ClientToServerPacketsOutOfOrder",
		"ServerToClientPacketsOutOfOrder",
		"NearRelayIDs",
		"NearRelayRTTs",
		"NearRelayJitters",
		"NearRelayPacketLosses",
		"EverOnNext",
		"SessionDuration",
		"TotalPriceSum",
		"EnvelopeBytesUpSum",
		"EnvelopeBytesDownSum",
		"DurationOnNext",
		"StartTimestamp",
		"NextRTT",
		"NextJitter",
		"NextPacketLoss",
		"PredictedNextRTT",
		"NearRelayRTT",
		"NextRelays",
		"NextRelaysPrice",
		"TotalPrice",
		"RouteDiversity",
		"Uncommitted",
		"Multipath",
		"RTTReduction",
		"PacketLossReduction",
		"RouteChanged",
		"NextBytesUp",
		"NextBytesDown",
		"FallbackToDirect",
		"MultipathVetoed",
		"Mispredicted",
		"Vetoed",
		"LatencyWorse",
		"NoRoute",
		"NextLatencyTooHigh",
		"CommitVeto",
		"UnknownDatacenter",
		"DatacenterNotEnabled",
		"BuyerNotLive",
		"StaleRouteMatrix",
	})

	for _, billingEntry := range newRows {
		// Timestamp
		timestamp := billingEntry.Timestamp.String()
		// SessionID
		sessionID := fmt.Sprintf("%016x", uint64(billingEntry.SessionID))
		// SliceNumber
		sliceNumber := fmt.Sprintf("%d", billingEntry.SliceNumber)
		// DirectMinRTT
		directMinRTT := fmt.Sprintf("%d", billingEntry.DirectMinRTT)
		// DirectMaxRTT
		directMaxRTT := ""
		if billingEntry.DirectMaxRTT.Valid {
			directMaxRTT = fmt.Sprintf("%d", billingEntry.DirectMaxRTT.Int64)
		}
		// DirectPrimeRTT
		directPrimeRTT := ""
		if billingEntry.DirectPrimeRTT.Valid {
			directPrimeRTT = fmt.Sprintf("%d", billingEntry.DirectPrimeRTT.Int64)
		}
		// DirectJitter
		directJitter := fmt.Sprintf("%d", billingEntry.DirectJitter)
		// DirectPacketLoss
		directPacketLoss := fmt.Sprintf("%d", billingEntry.DirectPacketLoss)
		// RealPacketLoss
		realPacketLoss := fmt.Sprintf("%5.5f", billingEntry.RealPacketLoss)
		// RealJitter
		realJitter := fmt.Sprintf("%d", billingEntry.RealJitter)
		// Next
		next := ""
		if billingEntry.Next.Valid && billingEntry.Next.Bool {
			next = "true"
		}
		// Flagged
		flagged := ""
		if billingEntry.Flagged.Valid && billingEntry.Flagged.Bool {
			flagged = "true"
		}
		// Summary
		summary := ""
		if billingEntry.Summary.Valid && billingEntry.Summary.Bool {
			summary = "true"
		}
		// Debug
		debug := ""
		if billingEntry.Debug.Valid {
			debug = billingEntry.Debug.StringVal
		}
		// Datacenter
		datacenter := ""
		if billingEntry.DatacenterString.Valid {
			datacenter = billingEntry.DatacenterString.StringVal
		}
		// BuyerID
		buyerID := ""
		if billingEntry.BuyerID.Valid {
			buyerID = fmt.Sprintf("%016x", uint64(billingEntry.BuyerID.Int64))
		}
		// UserHash
		userHash := ""
		if billingEntry.UserHash.Valid {
			userHash = fmt.Sprintf("%016x", billingEntry.UserHash.Int64)
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
		// ClientAddress
		clientAddress := ""
		if billingEntry.ClientAddress.Valid {
			clientAddress = billingEntry.ClientAddress.StringVal
		}
		// ISP
		isp := ""
		if billingEntry.ISP.Valid {
			isp = billingEntry.ISP.StringVal
		}
		// ConnectionType
		connectionType := ""
		if billingEntry.ConnectionType.Valid {
			connectionType = fmt.Sprintf("%d", billingEntry.ConnectionType.Int64)
		}
		// PlatformType
		platformType := ""
		if billingEntry.PlatformType.Valid {
			platformType = fmt.Sprintf("%d", billingEntry.PlatformType.Int64)
		}
		// SdkVersion
		sdkVersion := ""
		if billingEntry.SDKVersion.Valid {
			sdkVersion = billingEntry.SDKVersion.StringVal
		}
		// Tags
		tags := ""
		if len(billingEntry.Tags) > 0 {
			for _, tag := range billingEntry.Tags {
				tags += fmt.Sprintf("%016x", uint64(tag.Int64)) + ", "
			}
			tags = strings.TrimSuffix(tags, ", ")
		}
		// ABTest
		abTest := ""
		if billingEntry.ABTest.Valid && billingEntry.ABTest.Bool {
			abTest = "true"
		}
		// Pro
		pro := ""
		if billingEntry.Pro.Valid && billingEntry.Pro.Bool {
			pro = fmt.Sprintf("%t", billingEntry.Pro.Bool)
		}
		// ClientToServerPacketsSent
		clientToServerPacketsSent := ""
		if billingEntry.ClientToServerPacketsSent.Valid {
			clientToServerPacketsSent = fmt.Sprintf("%d", billingEntry.ClientToServerPacketsSent.Int64)
		}
		// ServerToClientPacketsSent
		serverToClientPacketsSent := ""
		if billingEntry.ServerToClientPacketsSent.Valid {
			serverToClientPacketsSent = fmt.Sprintf("%d", billingEntry.ServerToClientPacketsSent.Int64)
		}
		// ClientToServerPacketsLost
		clientToServerPacketsLost := ""
		if billingEntry.ClientToServerPacketsLost.Valid {
			clientToServerPacketsLost = fmt.Sprintf("%d", billingEntry.ClientToServerPacketsLost.Int64)
		}
		// ServerToClientPacketsLost
		serverToClientPacketsLost := ""
		if billingEntry.ServerToClientPacketsLost.Valid {
			serverToClientPacketsLost = fmt.Sprintf("%d", billingEntry.ServerToClientPacketsLost.Int64)
		}
		// ClientToServerPacketsOutOfOrder
		clientToServerPacketsOutOfOrder := ""
		if billingEntry.ClientToServerPacketsOutOfOrder.Valid {
			clientToServerPacketsOutOfOrder = fmt.Sprintf("%d", billingEntry.ClientToServerPacketsOutOfOrder.Int64)
		}
		// ServerToClientPacketsOutOfOrder
		serverToClientPacketsOutOfOrder := ""
		if billingEntry.ServerToClientPacketsOutOfOrder.Valid {
			serverToClientPacketsOutOfOrder = fmt.Sprintf("%d", billingEntry.ServerToClientPacketsOutOfOrder.Int64)
		}
		// NearRelayIDs
		nearRelayIDs := ""
		if len(billingEntry.NearRelayIDs) > 0 {
			for _, relayID := range billingEntry.NearRelayIDs {
				nearRelayIDs += fmt.Sprintf("%016x", uint64(relayID.Int64)) + ", "
			}
			nearRelayIDs = strings.TrimSuffix(nearRelayIDs, ", ")
		}
		// NearRelayRTTs
		nearRelayRTTs := ""
		if len(billingEntry.NearRelayRTTs) > 0 {
			for _, relayID := range billingEntry.NearRelayRTTs {
				nearRelayRTTs += fmt.Sprintf("%d", relayID.Int64) + ", "
			}
			nearRelayRTTs = strings.TrimSuffix(nearRelayRTTs, ", ")
		}
		// NearRelayJitters
		nearRelayJitters := ""
		if len(billingEntry.NearRelayJitters) > 0 {
			for _, relayID := range billingEntry.NearRelayJitters {
				nearRelayJitters += fmt.Sprintf("%d", relayID.Int64) + ", "
			}
			nearRelayJitters = strings.TrimSuffix(nearRelayJitters, ", ")
		}
		// NearRelayPacketLosses
		nearRelayPacketLosses := ""
		if len(billingEntry.NearRelayPacketLosses) > 0 {
			for _, relayID := range billingEntry.NearRelayPacketLosses {
				nearRelayPacketLosses += fmt.Sprintf("%d", relayID.Int64) + ", "
			}
			nearRelayPacketLosses = strings.TrimSuffix(nearRelayPacketLosses, ", ")
		}
		// EverOnNext
		everOnNext := ""
		if billingEntry.EverOnNext.Valid {
			everOnNext = fmt.Sprintf("%t", billingEntry.EverOnNext.Bool)
		}
		// SessionDuration
		sessionDuration := ""
		if billingEntry.SessionDuration.Valid {
			sessionDuration = fmt.Sprintf("%d", billingEntry.SessionDuration.Int64)
		}
		// TotalPriceSum
		totalPriceSum := ""
		if billingEntry.TotalPriceSum.Valid {
			totalPriceSum = fmt.Sprintf("%d", billingEntry.TotalPriceSum.Int64)
		}
		// EnvelopeBytesUpSum
		envelopeBytesUpSum := ""
		if billingEntry.EnvelopeBytesUpSum.Valid {
			envelopeBytesUpSum = fmt.Sprintf("%d", billingEntry.EnvelopeBytesUpSum.Int64)
		}
		// EnvelopeBytesDownSum
		envelopeBytesDownSum := ""
		if billingEntry.EnvelopeBytesDownSum.Valid {
			envelopeBytesDownSum = fmt.Sprintf("%d", billingEntry.EnvelopeBytesDownSum.Int64)
		}
		// DurationOnNext
		durationOnNext := ""
		if billingEntry.DurationOnNext.Valid {
			durationOnNext = fmt.Sprintf("%d", billingEntry.DurationOnNext.Int64)
		}
		// StartTimestamp
		startTimestamp := ""
		if billingEntry.StartTimestamp.Valid {
			startTimestamp = billingEntry.StartTimestamp.String()
		}
		// NextRTT
		nextRTT := ""
		if billingEntry.NextRTT.Valid {
			nextRTT = fmt.Sprintf("%d", billingEntry.NextRTT.Int64)
		}
		// NextJitter
		nextJitter := ""
		if billingEntry.NextJitter.Valid {
			nextJitter = fmt.Sprintf("%d", billingEntry.NextJitter.Int64)
		}
		// NextPacketLoss
		nextPacketLoss := ""
		if billingEntry.NextPacketLoss.Valid {
			nextPacketLoss = fmt.Sprintf("%d", billingEntry.NextPacketLoss.Int64)
		}
		// PredictedNextRTT
		predictedNextRTT := ""
		if billingEntry.PredictedNextRTT.Valid {
			predictedNextRTT = fmt.Sprintf("%d", billingEntry.PredictedNextRTT.Int64)
		}
		// NearRelayRTT
		nearRelayRTT := ""
		if billingEntry.NearRelayRTT.Valid {
			nearRelayRTT = fmt.Sprintf("%d", billingEntry.NearRelayRTT.Int64)
		}
		// NextRelays
		nextRelays := billingEntry.NextRelaysStrings
		// NextRelaysPrice
		nextRelaysPrice := ""
		if len(billingEntry.NextRelaysPrice) > 0 {
			for _, nextRelayPrice := range billingEntry.NextRelaysPrice {
				nextRelaysPrice += fmt.Sprintf("%d", nextRelayPrice.Int64) + ", "
			}
			nextRelaysPrice = strings.TrimSuffix(nextRelaysPrice, ", ")
		}
		// TotalPrice
		totalPrice := ""
		if billingEntry.TotalPrice.Valid {
			totalPrice = fmt.Sprintf("%d", billingEntry.TotalPrice.Int64)
		}
		// RouteDiversity
		routeDiversity := ""
		if billingEntry.RouteDiversity.Valid {
			routeDiversity = fmt.Sprintf("%d", billingEntry.RouteDiversity.Int64)
		}
		// Uncommitted
		uncommitted := ""
		if billingEntry.Uncommitted.Valid && billingEntry.Uncommitted.Bool {
			uncommitted = "true"
		}
		// Multipath
		multipath := ""
		if billingEntry.Multipath.Valid && billingEntry.Multipath.Bool {
			multipath = "true"
		}
		// RTTReduction
		rttReduction := ""
		if billingEntry.RTTReduction.Valid && billingEntry.RTTReduction.Bool {
			rttReduction = "true"
		}
		// PacketLossReduction
		packetLossReduction := ""
		if billingEntry.PacketLossReduction.Valid && billingEntry.PacketLossReduction.Bool {
			packetLossReduction = "true"
		}
		// RouteChanged
		routeChanged := ""
		if billingEntry.RouteChanged.Valid && billingEntry.RouteChanged.Bool {
			routeChanged = "true"
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
		// FallbackToDirect
		fallbackToDirect := ""
		if billingEntry.FallbackToDirect.Valid && billingEntry.FallbackToDirect.Bool {
			fallbackToDirect = "true"
		}
		// MultipathVetoed
		multipathVetoed := ""
		if billingEntry.MultipathVetoed.Valid && billingEntry.MultipathVetoed.Bool {
			multipathVetoed = "true"
		}
		// Mispredicted
		mispredicted := ""
		if billingEntry.Mispredicted.Valid && billingEntry.Mispredicted.Bool {
			mispredicted = "true"
		}
		// Vetoed
		vetoed := ""
		if billingEntry.Vetoed.Valid && billingEntry.Vetoed.Bool {
			vetoed = "true"
		}
		// LatencyWorse
		latencyWorse := ""
		if billingEntry.LatencyWorse.Valid && billingEntry.LatencyWorse.Bool {
			latencyWorse = "true"
		}
		// NoRoute
		noRoute := ""
		if billingEntry.NoRoute.Valid && billingEntry.NoRoute.Bool {
			noRoute = "true"
		}
		// NextLatencyTooHigh
		nextLatencyTooHigh := ""
		if billingEntry.NextLatencyTooHigh.Valid && billingEntry.NextLatencyTooHigh.Bool {
			nextLatencyTooHigh = "true"
		}
		// CommitVeto
		commitVeto := ""
		if billingEntry.CommitVeto.Valid && billingEntry.CommitVeto.Bool {
			commitVeto = "true"
		}
		// UnknownDatacenter
		unknownDatacenter := ""
		if billingEntry.UnknownDatacenter.Valid && billingEntry.UnknownDatacenter.Bool {
			unknownDatacenter = fmt.Sprintf("%t", billingEntry.UnknownDatacenter.Bool)
		}
		// DatacenterNotEnabled
		datacenterNotEnabled := ""
		if billingEntry.DatacenterNotEnabled.Valid && billingEntry.DatacenterNotEnabled.Bool {
			datacenterNotEnabled = fmt.Sprintf("%t", billingEntry.DatacenterNotEnabled.Bool)
		}
		// BuyerNotLive
		buyerNotLive := ""
		if billingEntry.BuyerNotLive.Valid && billingEntry.BuyerNotLive.Bool {
			buyerNotLive = fmt.Sprintf("%t", billingEntry.BuyerNotLive.Bool)
		}
		// StaleRouteMatrix
		staleRouteMatrix := ""
		if billingEntry.StaleRouteMatrix.Valid && billingEntry.StaleRouteMatrix.Bool {
			staleRouteMatrix = fmt.Sprintf("%t", billingEntry.StaleRouteMatrix.Bool)
		}

		bqBilling2DataEntryCSV = append(bqBilling2DataEntryCSV, []string{
			timestamp,
			sessionID,
			sliceNumber,
			directMinRTT,
			directMaxRTT,
			directPrimeRTT,
			directJitter,
			directPacketLoss,
			realPacketLoss,
			realJitter,
			next,
			flagged,
			summary,
			debug,
			datacenter,
			buyerID,
			userHash,
			envelopeBytesUp,
			envelopeBytesDown,
			latitude,
			longitude,
			clientAddress,
			isp,
			connectionType,
			platformType,
			sdkVersion,
			tags,
			abTest,
			pro,
			clientToServerPacketsSent,
			serverToClientPacketsSent,
			clientToServerPacketsLost,
			serverToClientPacketsLost,
			clientToServerPacketsOutOfOrder,
			serverToClientPacketsOutOfOrder,
			nearRelayIDs,
			nearRelayRTTs,
			nearRelayJitters,
			nearRelayPacketLosses,
			everOnNext,
			sessionDuration,
			totalPriceSum,
			envelopeBytesUpSum,
			envelopeBytesDownSum,
			durationOnNext,
			startTimestamp,
			nextRTT,
			nextJitter,
			nextPacketLoss,
			predictedNextRTT,
			nearRelayRTT,
			nextRelays,
			nextRelaysPrice,
			totalPrice,
			routeDiversity,
			uncommitted,
			multipath,
			rttReduction,
			packetLossReduction,
			routeChanged,
			nextBytesUp,
			nextBytesDown,
			fallbackToDirect,
			multipathVetoed,
			mispredicted,
			vetoed,
			latencyWorse,
			noRoute,
			nextLatencyTooHigh,
			commitVeto,
			unknownDatacenter,
			datacenterNotEnabled,
			buyerNotLive,
			staleRouteMatrix,
		})
	}

	fileName := "./session-" + fmt.Sprintf("%016x", sessionID) + ".csv"
	f, err := os.Create(fileName)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("Error creating local CSV file %s: %v\n", fileName, err), 1)
	}

	writer := csv.NewWriter(f)
	err = writer.WriteAll(bqBilling2DataEntryCSV)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("Error writing local CSV file %s: %v\n", fileName, err), 1)
	}
	fmt.Println("CSV file written: ", fileName)
	return

}

func GetAllSessionBilling2Info(sessionID int64, env Environment) ([]BigQueryBilling2Entry, error) {

	ctx := context.Background()

	var rows []BigQueryBilling2Entry

	var dbName string
	var sql bytes.Buffer

	sql.Write([]byte(`select 
	timestamp,
	sessionID,
	sliceNumber,
	directRTT,
	directJitter,
	directPacketLoss,
	realPacketLoss,
	realJitter,
	next,
	flagged,
	summary,
	debug,
	datacenterID,
	buyerID,
	userHash,
	envelopeBytesUp,
	envelopeBytesDown,
	latitude,
	longitude,
	isp,
	connectionType,
	platformType,
	sdkVersion,
	tags,
	abTest,
	pro,
	clientToServerPacketsSent,
	serverToClientPacketsSent,
	clientToServerPacketsLost,
	serverToClientPacketsLost,
	clientToServerPacketsOutOfOrder,
	serverToClientPacketsOutOfOrder,
	nearRelayIDs,
	nearRelayRTTs,
	nearRelayJitters,
	nearRelayPacketLosses,
	nextRTT,
	nextJitter,
	nextPacketLoss,
	predictedNextRTT,
	nearRelayRTT,
	if(ARRAY_LENGTH(nextRelays)>0, nextRelays[OFFSET(0)], NULL) as NextRelay0,
	if(ARRAY_LENGTH(nextRelays)>1, nextRelays[OFFSET(1)], NULL) as NextRelay1,
	if(ARRAY_LENGTH(nextRelays)>2, nextRelays[OFFSET(2)], NULL) as NextRelay2,
	if(ARRAY_LENGTH(nextRelays)>3, nextRelays[OFFSET(3)], NULL) as NextRelay3,
	if(ARRAY_LENGTH(nextRelays)>4, nextRelays[OFFSET(4)], NULL) as NextRelay4,
	nextRelayPrice,
	totalPrice,
	routeDiversity,
	uncommitted,
	multipath,
	rttReduction,
	packetLossReduction,
	routeChanged,
	fallbackToDirect,
	multipathVetoed,
	mispredicted,
	vetoed,
	latencyWorse,
	noRoute,
	nextLatencyTooHigh,
	commitVeto,
	unknownDatacenter,
	datacenterNotEnabled,
	buyerNotLive,
	staleRouteMatrix,
	nextBytesUp,
	nextBytesDown,
	everOnNext,
	sessionDuration,
	totalPriceSum,
	envelopeBytesUpSum,
	envelopeBytesDownSum,
	durationOnNext,
	clientAddress,
	startTimestamp,
	directMaxRTT,
	directPrimeRTT,
    from `))

	if env.Name != "prod" && env.Name != "dev" && env.Name != "staging" {
		fmt.Println("Local/testing functionality TBD.")
	} else {
		if env.Name == "prod" {
			sql.Write([]byte("network-next-v3-prod.prod.billing2"))
			dbName = "network-next-v3-prod"

		} else if env.Name == "staging" {
			sql.Write([]byte("network-next-v3-staging.staging.billing2"))
			dbName = "network-next-v3-staging"
		} else if env.Name == "dev" {
			sql.Write([]byte("network-next-v3-dev.dev.billing2"))
			dbName = "network-next-v3-dev"
		}

		sql.Write([]byte(" where sessionID = "))
		sql.Write([]byte(fmt.Sprintf("%d", sessionID)))
		// a timestamp must be provided although it is not relevant to this query
		sql.Write([]byte(" and DATE(timestamp) >= '1968-05-01'"))
		sql.Write([]byte(" order by sliceNumber asc"))

		bqClient, err := bigquery.NewClient(ctx, dbName)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2Info() failed to create BigQuery client: %v", err), 1)
			return nil, err
		}
		defer bqClient.Close()

		q := bqClient.Query(string(sql.String()))

		job, err := q.Run(ctx)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2Info() failed to query BigQuery: %v", err), 1)
			return nil, err
		}

		status, err := job.Wait(ctx)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2Info() error waiting for job to complete: %v", err), 1)
			return nil, err
		}
		if err := status.Err(); err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2Info() job returned an error: %v", err), 1)
			return nil, err
		}

		it, err := job.Read(ctx)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2Info() job.Read() error: %v", err), 1)
			return nil, err
		}

		// process result set and load rows
		for {
			var rec BigQueryBilling2Entry
			err := it.Next(&rec)

			if err == iterator.Done {
				break
			}
			if err != nil {
				handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2Info() BigQuery iterator error: %v", err), 1)
				return nil, err
			}
			rows = append(rows, rec)
		}
	}

	return rows, nil

}

func dumpSession2Summary(env Environment, sessionID uint64) {

	// make a call for all the relays (there is no relay "singular" endpoint)
	var relaysReply localjsonrpc.RelaysReply
	relaysArg := localjsonrpc.RelaysArgs{} // empty args returns all relays
	if err := makeRPCCall(env, &relaysReply, "OpsService.Relays", relaysArg); err != nil {
		handleJSONRPCError(env, err)
	}

	relayNames := make(map[int64]string)
	for _, relay := range relaysReply.Relays {
		relayNames[int64(relay.ID)] = relay.Name
	}

	// make a call for the datacenters
	var dcsReply localjsonrpc.DatacentersReply
	dcsArgs := localjsonrpc.DatacentersArgs{}
	if err := makeRPCCall(env, &dcsReply, "OpsService.Datacenters", dcsArgs); err != nil {
		handleJSONRPCError(env, err)
	}
	dcNames := make(map[int64]string)
	for _, dc := range dcsReply.Datacenters {
		dcNames[int64(dc.ID)] = dc.Name
	}

	rows, err := GetAllSessionBilling2SummaryInfo(int64(sessionID), env)

	var newRows []BigQueryBilling2EntrySummary

	// process returned resultset - wire up relay and datacenter names
	for index, row := range rows {

		newRows = append(newRows, row)

		newRows[index].DatacenterString = bigquery.NullString{StringVal: dcNames[row.DatacenterID.Int64], Valid: true}
	}

	bqBilling2SummaryDataEntryCSV := [][]string{{}}

	bqBilling2SummaryDataEntryCSV = append(bqBilling2SummaryDataEntryCSV, []string{
		"SessionID",
		"BuyerID",
		"UserHash",
		"Datacenter",
		"StartTimestamp",
		"Latitude",
		"Longitude",
		"ISP",
		"ConnectionType",
		"PlatformType",
		"Tags",
		"ABTest",
		"Pro",
		"SdkVersion",
		"EnvelopeBytesUp",
		"EnvelopeBytesDown",
		"ClientToServerPacketsSent",
		"ServerToClientPacketsSent",
		"ClientToServerPacketsLost",
		"ServerToClientPacketsLost",
		"ClientToServerPacketsOutOfOrder",
		"ServerToClientPacketsOutOfOrder",
		"NearRelayIDs",
		"NearRelayRTTs",
		"NearRelayJitters",
		"NearRelayPacketLosses",
		"EverOnNext",
		"SessionDuration",
		"TotalPriceSum",
		"EnvelopeBytesUpSum",
		"EnvelopeBytesDownSum",
		"DurationOnNext",
		"ClientAddress",
	})

	for _, billingEntry := range newRows {
		// SessionID
		sessionID := fmt.Sprintf("%016x", uint64(billingEntry.SessionID))
		// BuyerID
		buyerID := ""
		if billingEntry.BuyerID.Valid {
			buyerID = fmt.Sprintf("%016x", uint64(billingEntry.BuyerID.Int64))
		}
		// UserHash
		userHash := ""
		if billingEntry.UserHash.Valid {
			userHash = fmt.Sprintf("%016x", billingEntry.UserHash.Int64)
		}
		// Datacenter
		datacenter := ""
		if billingEntry.DatacenterString.Valid {
			datacenter = billingEntry.DatacenterString.StringVal
		}
		// StartTimestamp
		startTimestamp := ""
		if billingEntry.StartTimestamp.Valid {
			startTimestamp = billingEntry.StartTimestamp.String()
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
		// ConnectionType
		connectionType := ""
		if billingEntry.ConnectionType.Valid {
			connectionType = fmt.Sprintf("%d", billingEntry.ConnectionType.Int64)
		}
		// PlatformType
		platformType := ""
		if billingEntry.PlatformType.Valid {
			platformType = fmt.Sprintf("%d", billingEntry.PlatformType.Int64)
		}
		// Tags
		tags := ""
		if len(billingEntry.Tags) > 0 {
			for _, tag := range billingEntry.Tags {
				tags += fmt.Sprintf("%016x", uint64(tag.Int64)) + ", "
			}
			tags = strings.TrimSuffix(tags, ", ")
		}
		// ABTest
		abTest := ""
		if billingEntry.ABTest.Valid && billingEntry.ABTest.Bool {
			abTest = "true"
		}
		// Pro
		pro := ""
		if billingEntry.Pro.Valid && billingEntry.Pro.Bool {
			pro = fmt.Sprintf("%t", billingEntry.Pro.Bool)
		}
		// SdkVersion
		sdkVersion := ""
		if billingEntry.SDKVersion.Valid {
			sdkVersion = billingEntry.SDKVersion.StringVal
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
		// ClientToServerPacketsSent
		clientToServerPacketsSent := ""
		if billingEntry.ClientToServerPacketsSent.Valid {
			clientToServerPacketsSent = fmt.Sprintf("%d", billingEntry.ClientToServerPacketsSent.Int64)
		}
		// ServerToClientPacketsSent
		serverToClientPacketsSent := ""
		if billingEntry.ServerToClientPacketsSent.Valid {
			serverToClientPacketsSent = fmt.Sprintf("%d", billingEntry.ServerToClientPacketsSent.Int64)
		}
		// ClientToServerPacketsLost
		clientToServerPacketsLost := ""
		if billingEntry.ClientToServerPacketsLost.Valid {
			clientToServerPacketsLost = fmt.Sprintf("%d", billingEntry.ClientToServerPacketsLost.Int64)
		}
		// ServerToClientPacketsLost
		serverToClientPacketsLost := ""
		if billingEntry.ServerToClientPacketsLost.Valid {
			serverToClientPacketsLost = fmt.Sprintf("%d", billingEntry.ServerToClientPacketsLost.Int64)
		}
		// ClientToServerPacketsOutOfOrder
		clientToServerPacketsOutOfOrder := ""
		if billingEntry.ClientToServerPacketsOutOfOrder.Valid {
			clientToServerPacketsOutOfOrder = fmt.Sprintf("%d", billingEntry.ClientToServerPacketsOutOfOrder.Int64)
		}
		// ServerToClientPacketsOutOfOrder
		serverToClientPacketsOutOfOrder := ""
		if billingEntry.ServerToClientPacketsOutOfOrder.Valid {
			serverToClientPacketsOutOfOrder = fmt.Sprintf("%d", billingEntry.ServerToClientPacketsOutOfOrder.Int64)
		}
		// NearRelayIDs
		nearRelayIDs := ""
		if len(billingEntry.NearRelayIDs) > 0 {
			for _, relayID := range billingEntry.NearRelayIDs {
				nearRelayIDs += fmt.Sprintf("%016x", uint64(relayID.Int64)) + ", "
			}
			nearRelayIDs = strings.TrimSuffix(nearRelayIDs, ", ")
		}
		// NearRelayRTTs
		nearRelayRTTs := ""
		if len(billingEntry.NearRelayRTTs) > 0 {
			for _, relayID := range billingEntry.NearRelayRTTs {
				nearRelayRTTs += fmt.Sprintf("%d", relayID.Int64) + ", "
			}
			nearRelayRTTs = strings.TrimSuffix(nearRelayRTTs, ", ")
		}
		// NearRelayJitters
		nearRelayJitters := ""
		if len(billingEntry.NearRelayJitters) > 0 {
			for _, relayID := range billingEntry.NearRelayJitters {
				nearRelayJitters += fmt.Sprintf("%d", relayID.Int64) + ", "
			}
			nearRelayJitters = strings.TrimSuffix(nearRelayJitters, ", ")
		}
		// NearRelayPacketLosses
		nearRelayPacketLosses := ""
		if len(billingEntry.NearRelayPacketLosses) > 0 {
			for _, relayID := range billingEntry.NearRelayPacketLosses {
				nearRelayPacketLosses += fmt.Sprintf("%d", relayID.Int64) + ", "
			}
			nearRelayPacketLosses = strings.TrimSuffix(nearRelayPacketLosses, ", ")
		}
		// EverOnNext
		everOnNext := ""
		if billingEntry.EverOnNext.Valid {
			everOnNext = fmt.Sprintf("%t", billingEntry.EverOnNext.Bool)
		}
		// SessionDuration
		sessionDuration := ""
		if billingEntry.SessionDuration.Valid {
			sessionDuration = fmt.Sprintf("%d", billingEntry.SessionDuration.Int64)
		}
		// TotalPriceSum
		totalPriceSum := ""
		if billingEntry.TotalPriceSum.Valid {
			totalPriceSum = fmt.Sprintf("%d", billingEntry.TotalPriceSum.Int64)
		}
		// EnvelopeBytesUpSum
		envelopeBytesUpSum := ""
		if billingEntry.EnvelopeBytesUpSum.Valid {
			envelopeBytesUpSum = fmt.Sprintf("%d", billingEntry.EnvelopeBytesUpSum.Int64)
		}
		// EnvelopeBytesDownSum
		envelopeBytesDownSum := ""
		if billingEntry.EnvelopeBytesDownSum.Valid {
			envelopeBytesDownSum = fmt.Sprintf("%d", billingEntry.EnvelopeBytesDownSum.Int64)
		}
		// DurationOnNext
		durationOnNext := ""
		if billingEntry.DurationOnNext.Valid {
			durationOnNext = fmt.Sprintf("%d", billingEntry.DurationOnNext.Int64)
		}
		// ClientAddress
		clientAddress := ""
		if billingEntry.ClientAddress.Valid {
			clientAddress = billingEntry.ClientAddress.StringVal
		}

		bqBilling2SummaryDataEntryCSV = append(bqBilling2SummaryDataEntryCSV, []string{
			sessionID,
			buyerID,
			userHash,
			datacenter,
			startTimestamp,
			latitude,
			longitude,
			isp,
			connectionType,
			platformType,
			tags,
			abTest,
			pro,
			sdkVersion,
			envelopeBytesUp,
			envelopeBytesDown,
			clientToServerPacketsSent,
			serverToClientPacketsSent,
			clientToServerPacketsLost,
			serverToClientPacketsLost,
			clientToServerPacketsOutOfOrder,
			serverToClientPacketsOutOfOrder,
			nearRelayIDs,
			nearRelayRTTs,
			nearRelayJitters,
			nearRelayPacketLosses,
			everOnNext,
			sessionDuration,
			totalPriceSum,
			envelopeBytesUpSum,
			envelopeBytesDownSum,
			durationOnNext,
			clientAddress,
		})
	}

	fileName := "./session-summary-" + fmt.Sprintf("%016x", sessionID) + ".csv"
	f, err := os.Create(fileName)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("Error creating local CSV file %s: %v\n", fileName, err), 1)
	}

	writer := csv.NewWriter(f)
	err = writer.WriteAll(bqBilling2SummaryDataEntryCSV)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("Error writing local CSV file %s: %v\n", fileName, err), 1)
	}
	fmt.Println("CSV file written: ", fileName)
	return

}

func GetAllSessionBilling2SummaryInfo(sessionID int64, env Environment) ([]BigQueryBilling2EntrySummary, error) {

	ctx := context.Background()

	var rows []BigQueryBilling2EntrySummary

	var dbName string
	var sql bytes.Buffer

	sql.Write([]byte(`select 
	sessionID,
	buyerID,
	userHash,
	datacenterID,
	startTimestamp,
	latitude,
	longitude,
	isp,
	connectionType,
	platformType,
	tags,
	abTest,
	pro,
	sdkVersion,
	envelopeBytesUp,
	envelopeBytesDown,
	clientToServerPacketsSent,
	serverToClientPacketsSent,
	clientToServerPacketsLost,
	serverToClientPacketsLost,
	clientToServerPacketsOutOfOrder,
	serverToClientPacketsOutOfOrder,
	nearRelayIDs,
	nearRelayRTTs,
	nearRelayJitters,
	nearRelayPacketLosses,
	everOnNext,
	sessionDuration,
	totalPriceSum,
	envelopeBytesUpSum,
	envelopeBytesDownSum,
	durationOnNext,
	clientAddress,
    from `))

	if env.Name != "prod" && env.Name != "dev" && env.Name != "staging" {
		fmt.Println("Local/testing functionality TBD.")
	} else {
		if env.Name == "prod" {
			sql.Write([]byte("network-next-v3-prod.prod.billing2_session_summary"))
			dbName = "network-next-v3-prod"

		} else if env.Name == "staging" {
			sql.Write([]byte("network-next-v3-staging.staging.billing2_session_summary"))
			dbName = "network-next-v3-staging"
		} else if env.Name == "dev" {
			sql.Write([]byte("network-next-v3-dev.dev.billing2_session_summary"))
			dbName = "network-next-v3-dev"
		}

		sql.Write([]byte(" where sessionID = "))
		sql.Write([]byte(fmt.Sprintf("%d", sessionID)))
		// a timestamp must be provided although it is not relevant to this query
		sql.Write([]byte(" and DATE(startTimestamp) >= '1968-05-01'"))

		bqClient, err := bigquery.NewClient(ctx, dbName)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2SummaryInfo() failed to create BigQuery client: %v", err), 1)
			return nil, err
		}
		defer bqClient.Close()

		q := bqClient.Query(string(sql.String()))

		job, err := q.Run(ctx)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2SummaryInfo() failed to query BigQuery: %v", err), 1)
			return nil, err
		}

		status, err := job.Wait(ctx)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2SummaryInfo() error waiting for job to complete: %v", err), 1)
			return nil, err
		}
		if err := status.Err(); err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2Info() job returned an error: %v", err), 1)
			return nil, err
		}

		it, err := job.Read(ctx)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2SummaryInfo() job.Read() error: %v", err), 1)
			return nil, err
		}

		// process result set and load rows
		for {
			var rec BigQueryBilling2EntrySummary
			err := it.Next(&rec)

			if err == iterator.Done {
				break
			}
			if err != nil {
				handleRunTimeError(fmt.Sprintf("GetAllSessionBilling2SummaryInfo() BigQuery iterator error: %v", err), 1)
				return nil, err
			}
			rows = append(rows, rec)
		}
	}

	return rows, nil

}
