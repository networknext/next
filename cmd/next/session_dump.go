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
	"github.com/ybbus/jsonrpc"
	"google.golang.org/api/iterator"
)

func dumpSession(rpcClient jsonrpc.RPCClient, env Environment, sessionID uint64) {

	// make a call for all the relays (there is no relay "singular" endpoint)
	var relaysReply localjsonrpc.RelaysReply
	relaysArg := localjsonrpc.RelaysArgs{} // empty args returns all relays
	if err := rpcClient.CallFor(&relaysReply, "OpsService.Relays", relaysArg); err != nil {
		handleJSONRPCError(env, err)
	}

	relayNames := make(map[int64]string)
	for _, relay := range relaysReply.Relays {
		relayNames[int64(relay.ID)] = relay.Name
	}

	// make a call for the datacenters
	var dcsReply localjsonrpc.DatacentersReply
	dcsArgs := localjsonrpc.DatacentersArgs{}
	if err := rpcClient.CallFor(&dcsReply, "OpsService.Datacenters", dcsArgs); err != nil {
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
		"NoRoute",
		"NextLatencyTooHigh",
		"RouteChanged",
		"CommitVeto",
		"RelayWentAway",
		"RouteLost",
		"Debug String",
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
	committed,
	flagged,
	multipath,
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
	relayWentAway,
	routeLost,
	mispredicted,
	vetoed,
	latencyWorse,
	noRoute,
	nextLatencyTooHigh,
	routeChanged,
	commitVeto,
	tags
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
				return nil, err
			}
			rows = append(rows, rec)
		}
	}

	return rows, nil

}

// the test should be ported from jsonrpc/transport/buyers_test.go
// func returnLocalTestData(reply *GetAllSessionBillingInfoReply) ([]transport.BigQueryBillingEntry, error) {
// 	var localRow transport.BigQueryBillingEntry
// 	var rows []transport.BigQueryBillingEntry

// 	bqRow, err := ioutil.ReadFile("../../../testdata/bq_billing_row.json")
// 	if err != nil {
// 		err = fmt.Errorf("returnLocalTestData() error opening local testdata file: %v", err)
// 		return []transport.BigQueryBillingEntry{}, err
// 	}
// 	err = json.Unmarshal(bqRow, &localRow)
// 	if err != nil {
// 		err = fmt.Errorf("returnLocalTestData() error unmarshalling json from local file: %v", err)
// 		return []transport.BigQueryBillingEntry{}, err
// 	}

// 	rows = append(rows, localRow)

// 	return rows, nil
// }

// func slicesAreEqual(a, b []int64) bool {
// 	if len(a) != len(b) {
// 		return false
// 	}
// 	for i, v := range a {
// 		if v != b[i] {
// 			return false
// 		}
// 	}
// 	return true
// }
