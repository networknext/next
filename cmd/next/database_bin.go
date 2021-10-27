package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
)

func getLocalDatabaseBin() {
	ctx := context.Background()
	logger := log.NewLogfmtLogger(os.Stdout)
	gcpProjectID := ""
	db, err := backend.GetStorer(ctx, logger, gcpProjectID, "local")
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
	var dbWrapper routing.DatabaseBinWrapper
	var enabledRelays []routing.Relay
	relayMap := make(map[uint64]routing.Relay)
	buyerMap := make(map[uint64]routing.Buyer)
	sellerMap := make(map[string]routing.Seller)
	datacenterMap := make(map[uint64]routing.Datacenter)
	datacenterMaps := make(map[uint64]map[uint64]routing.DatacenterMap)

	buyers := db.Buyers(ctx)
	for _, buyer := range buyers {
		buyerMap[buyer.ID] = buyer
		dcMapsForBuyer := db.GetDatacenterMapsForBuyer(ctx, buyer.ID)
		datacenterMaps[buyer.ID] = dcMapsForBuyer
	}

	for _, seller := range db.Sellers(ctx) {
		sellerMap[seller.ShortName] = seller
	}

	for _, datacenter := range db.Datacenters(ctx) {
		datacenterMap[datacenter.ID] = datacenter
	}

	for _, localRelay := range db.Relays(ctx) {
		if localRelay.State == routing.RelayStateEnabled {
			enabledRelays = append(enabledRelays, localRelay)
			relayMap[localRelay.ID] = localRelay
		}
	}

	dbWrapper.Relays = enabledRelays
	dbWrapper.RelayMap = relayMap
	dbWrapper.BuyerMap = buyerMap
	dbWrapper.SellerMap = sellerMap
	dbWrapper.DatacenterMap = datacenterMap
	dbWrapper.DatacenterMaps = datacenterMaps

	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(dbWrapper)

	err = ioutil.WriteFile("./database.bin", buffer.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Failed to write database file")
	}
}

func getDatabaseBin(
	env Environment,
) {

	args := localjsonrpc.NextBinFileHandlerArgs{}
	var reply localjsonrpc.NextBinFileHandlerReply

	if err := makeRPCCall(env, &reply, "RelayFleetService.NextBinFileHandler", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(reply.DBWrapper)

	err := ioutil.WriteFile("database.bin", buffer.Bytes(), 0777)
	if err != nil {
		err := fmt.Errorf("BinFileHandler() error writing database.bin to filesystem: %v", err)
		handleRunTimeError(fmt.Sprintf("could not write database.bin to the filesystem: %v\n", err), 0)
	}

}

func createStagingDatabaseBin(numRelays int) {
	dbWrapper := routing.CreateEmptyDatabaseBinWrapper()

	// Create Buyers
	ghostArmyPK, err := base64.StdEncoding.DecodeString("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==")
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not decode ghost army public key: %v\n", err), 1)
	}
	nextPK, err := base64.StdEncoding.DecodeString("uLk+H9QWvDxBLyEAP1JBmS5U/rHXi0RyrFammau9t2c=")
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not decode next public key: %v\n", err), 1)
	}
	stagingSellerPK, err := base64.StdEncoding.DecodeString("5w4h6mAzN5Vembvv8LC/9WePTEGuPcXgPiEj4yK1zyk=")
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not decode staging seller public key: %v\n", err), 1)
	}

	defaultRouteShader := core.RouteShader{
		DisableNetworkNext:        false,
		SelectionPercent:          100,
		PacketLossSustained:       100,
		ABTest:                    false,
		ProMode:                   false,
		ReduceLatency:             true,
		ReduceJitter:              true,
		ReducePacketLoss:          true,
		Multipath:                 false,
		AcceptableLatency:         0,
		LatencyThreshold:          10,
		AcceptablePacketLoss:      1,
		BandwidthEnvelopeUpKbps:   1024,
		BandwidthEnvelopeDownKbps: 1024,
		BannedUsers:               make(map[uint64]bool),
	}
	defaultInternalConfig := core.InternalConfig{
		ReducePacketLossMinSliceNumber: 0,
		RouteSelectThreshold:           2,
		RouteSwitchThreshold:           5,
		MaxLatencyTradeOff:             20,
		RTTVeto_Default:                -10,
		RTTVeto_Multipath:              -20,
		RTTVeto_PacketLoss:             -30,
		MultipathOverloadThreshold:     500,
		TryBeforeYouBuy:                false,
		ForceNext:                      false,
		LargeCustomer:                  false,
		Uncommitted:                    false,
		MaxRTT:                         300,
		HighFrequencyPings:             true,
		RouteDiversity:                 0,
		MultipathThreshold:             25,
		EnableVanityMetrics:            false,
	}
	nextInternalConfig := core.InternalConfig{
		RouteSelectThreshold:       0,
		RouteSwitchThreshold:       5,
		MaxLatencyTradeOff:         10,
		RTTVeto_Default:            -5,
		RTTVeto_Multipath:          -20,
		RTTVeto_PacketLoss:         -20,
		MultipathOverloadThreshold: 500,
		TryBeforeYouBuy:            false,
		ForceNext:                  true,
		LargeCustomer:              true,
		Uncommitted:                false,
		MaxRTT:                     300,
		HighFrequencyPings:         false,
		RouteDiversity:             0,
		MultipathThreshold:         0,
		EnableVanityMetrics:        true,
	}

	buyerGhostArmy := routing.Buyer{
		CompanyCode:    "ghost-army",
		ShortName:      "ghost-army",
		ID:             uint64(0),
		HexID:          "0000000000000000",
		Live:           true,
		Debug:          false,
		PublicKey:      ghostArmyPK,
		RouteShader:    defaultRouteShader,
		InternalConfig: defaultInternalConfig,
		DatabaseID:     3,
		CustomerID:     1,
	}

	nextRouteShader := defaultRouteShader
	nextRouteShader.ReduceJitter = false
	nextRouteShader.AcceptableLatency = 25
	nextRouteShader.LatencyThreshold = 5

	buyerNext := routing.Buyer{
		CompanyCode:    "next",
		ShortName:      "next",
		ID:             uint64(13672574147039585173),
		HexID:          "bdbebdbf0f7be395",
		Live:           true,
		Debug:          false,
		PublicKey:      nextPK,
		RouteShader:    nextRouteShader,
		InternalConfig: nextInternalConfig,
		DatabaseID:     1,
		CustomerID:     3,
	}

	stagingSellerRouteShader := defaultRouteShader
	stagingSellerRouteShader.SelectionPercent = 1
	stagingSellerInternalConfig := defaultInternalConfig
	stagingSellerInternalConfig.LargeCustomer = true
	stagingSellerInternalConfig.EnableVanityMetrics = true

	buyerStagingSeller := routing.Buyer{
		CompanyCode:    "stagingseller",
		ShortName:      "stagingseller",
		ID:             uint64(13053258624167246632),
		HexID:          "b5267d8f3ecafb28",
		Live:           true,
		Debug:          true,
		PublicKey:      stagingSellerPK,
		RouteShader:    stagingSellerRouteShader,
		InternalConfig: stagingSellerInternalConfig,
		DatabaseID:     2,
		CustomerID:     2,
	}

	// Fill in buyers
	dbWrapper.BuyerMap[buyerGhostArmy.ID] = buyerGhostArmy
	dbWrapper.BuyerMap[buyerNext.ID] = buyerNext
	dbWrapper.BuyerMap[buyerStagingSeller.ID] = buyerStagingSeller

	// Fill in sellers
	seller := routing.Seller{
		ID:                       "stagingseller",
		Name:                     "staging seller",
		CompanyCode:              "stagingseller",
		ShortName:                "stagingseller",
		Secret:                   false,
		EgressPriceNibblinsPerGB: routing.Nibblin(2000000000),
		DatabaseID:               1,
		CustomerID:               2,
	}
	dbWrapper.SellerMap[seller.ID] = seller

	// Save the list of dest datacenters
	destDatacenters := make(map[uint64]routing.Datacenter)

	// Create and fill in 80 dest datacenters
	for i := 0; i < 80; i++ {
		name := fmt.Sprintf("staging.%d", i+1)
		dc := routing.Datacenter{
			ID:         crypto.HashID(name),
			Name:       name,
			Location:   generateRandomLocation(),
			SellerID:   1,
			DatabaseID: int64(i + 1),
		}
		dbWrapper.DatacenterMap[dc.ID] = dc
		destDatacenters[dc.ID] = dc
	}

	// Create and fill in fake relays and any additional datacenters
	relayPK, err := base64.StdEncoding.DecodeString("8hUCRvzKh2aknL9RErM/Vj22+FGJW0tWMRz5KlHKryE=")
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not decode relay public key: %v\n", err), 1)
	}

	var maxDatacenterDbID int64
	var maxRelayDbID int64
	for i := 0; i < numRelays; i++ {
		// Create the IP and UDP address for the fake relay
		ipAddress := fmt.Sprintf("127.0.0.1:%d", 10000+i)
		udpAddr, err := net.ResolveUDPAddr("udp", ipAddress)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("could not resolve %s to udp addr: %v\n", ipAddress, err), 1)
		}

		// Get the datacenter for this relay
		var datacenter routing.Datacenter

		dcName := fmt.Sprintf("staging.%d", i+1)
		dcID := crypto.HashID(dcName)
		dc, exists := dbWrapper.DatacenterMap[dcID]
		if exists {
			// Use one of the pre-made 80 datacenters
			datacenter = dc
			if dc.DatabaseID > maxDatacenterDbID {
				maxDatacenterDbID = dc.DatabaseID
			}
		} else {
			// Create a new datacenter for this fake relay
			datacenter = routing.Datacenter{
				ID:         dcID,
				Name:       dcName,
				Location:   generateRandomLocation(),
				SellerID:   1,
				DatabaseID: maxDatacenterDbID + 1,
			}

			// Increment datacenter database ID counter
			maxDatacenterDbID++

			// Add the new datacenter to the map
			dbWrapper.DatacenterMap[dcID] = datacenter
		}

		// Create the fake relay
		relayID := crypto.HashID(ipAddress)
		relay := routing.Relay{
			ID:                  relayID,
			Name:                fmt.Sprintf("staging.relay.%d", i+1),
			Addr:                *udpAddr,
			PublicKey:           relayPK,
			Seller:              seller,
			Datacenter:          datacenter,
			NICSpeedMbps:        1000,
			IncludedBandwidthGB: 1,
			State:               routing.RelayStateEnabled,
			MaxSessions:         8000,
			DatabaseID:          maxRelayDbID,
			Notes:               "I am a staging relay." + fmt.Sprintf("%d", i+1) + " - let me load test!",
			Version:             "2.0.8",
		}

		// Increment the relay database ID counter
		maxRelayDbID++

		// Add the fake relay to the map and array
		dbWrapper.RelayMap[relayID] = relay
		dbWrapper.Relays = append(dbWrapper.Relays, relay)
	}

	// Create datacenter maps for next and stagingseller
	dcMapsNext := make(map[uint64]routing.DatacenterMap)
	dcMapsStagingSeller := make(map[uint64]routing.DatacenterMap)
	for dcID, _ := range dbWrapper.DatacenterMap {
		dcMapNext := routing.DatacenterMap{
			BuyerID:      buyerNext.ID,
			DatacenterID: dcID,
		}
		dcMapStagingSeller := routing.DatacenterMap{
			BuyerID:      buyerStagingSeller.ID,
			DatacenterID: dcID,
		}
		// Add the datacenter map per buyer to the mapping
		// Only include 80 dest datacenters
		if _, exists := destDatacenters[dcID]; exists {
			dcMapsNext[dcID] = dcMapNext
			dcMapsStagingSeller[dcID] = dcMapStagingSeller
		}
	}

	// Fill in the datacenter maps for the buyers
	dbWrapper.DatacenterMaps[buyerGhostArmy.ID] = make(map[uint64]routing.DatacenterMap)
	dbWrapper.DatacenterMaps[buyerNext.ID] = dcMapsNext
	dbWrapper.DatacenterMaps[buyerStagingSeller.ID] = dcMapsStagingSeller

	// Fill in metadata
	now := time.Now().UTC()
	dbWrapper.CreationTime = fmt.Sprintf("%s %d, %d %02d:%02d UTC\n", now.Month(), now.Day(), now.Year(), now.Hour(), now.Minute())
	dbWrapper.Creator = "next"

	// Encode the database bin wrapper using gob and write to disk
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(dbWrapper)

	err = ioutil.WriteFile("./database.bin", buffer.Bytes(), 0644)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("failed to write staging database file: %v\n", err), 1)
	}

	fmt.Printf("Generated staging ./database.bin with %d fake relays\n", numRelays)
}

func generateRandomLocation() routing.Location {
	// Create a random sessionID because it has 8 bytes
	sessionID := crypto.GenerateSessionID()

	// Generate a random lat/long from the session ID
	sessionIDBytes := [8]byte{}
	binary.LittleEndian.PutUint64(sessionIDBytes[0:8], sessionID)

	// Randomize the location by using 4 bits of the sessionID for the lat, and the other 4 for the long
	latBits := binary.LittleEndian.Uint32(sessionIDBytes[0:4])
	longBits := binary.LittleEndian.Uint32(sessionIDBytes[4:8])

	lat := (float32(latBits)) / 0xFFFFFFFF
	long := (float32(longBits)) / 0xFFFFFFFF

	return routing.Location{
		Latitude:  (-90.0 + lat*180.0) * 0.5,
		Longitude: -180.0 + long*360.0,
	}
}

func checkMetaData() {
	f2, err := os.Open("database.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Creator     : %s\n", incomingDB.Creator)
	fmt.Printf("CreationTime: %s\n\n", incomingDB.CreationTime)
}

func checkRelaysInBinFile() {

	f2, err := os.Open("database.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// print a list
	sort.SliceStable(incomingDB.Relays, func(i, j int) bool {
		return incomingDB.Relays[i].Name < incomingDB.Relays[j].Name
	})

	fmt.Println("Relays:")
	fmt.Printf("\t%-25s %-18s %-22s %10s\n", "Name", "ID", "Address", "Version")
	fmt.Printf("\t%s\n", strings.Repeat("-", 80))
	for _, relay := range incomingDB.Relays {
		id := strings.ToUpper(fmt.Sprintf("%016x", relay.ID))
		fmt.Printf("\t%-25s %016s %22s %+10s\n", relay.Name, id, relay.Addr.String(), relay.Version)
	}
	fmt.Println()

}

func checkDatacentersInBinFile() {
	f2, err := os.Open("database.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var datacenters []routing.Datacenter
	for _, datacenter := range incomingDB.DatacenterMap {
		datacenters = append(datacenters, datacenter)
	}

	// print a list
	sort.SliceStable(datacenters, func(i, j int) bool {
		return datacenters[i].Name < datacenters[j].Name
	})

	fmt.Println("Datacenters:")
	fmt.Printf("\t%-25s %-16s\n", "Name", "ID")
	fmt.Printf("\t%s\n", strings.Repeat("-", 43))
	for _, datacenter := range datacenters {
		id := strings.ToUpper(fmt.Sprintf("%016x", datacenter.ID))
		fmt.Printf("\t%-25s %016s\n", datacenter.Name, id)
	}
	fmt.Println()
}

func checkSellersInBinFile() {
	f2, err := os.Open("database.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var sellers []routing.Seller
	for _, seller := range incomingDB.SellerMap {
		sellers = append(sellers, seller)
	}

	// print a list
	sort.SliceStable(sellers, func(i, j int) bool {
		return sellers[i].ID < sellers[j].ID
	})

	fmt.Println("Sellers:")
	fmt.Printf("\t%-25s\n", "ID")
	fmt.Printf("\t%s\n", strings.Repeat("-", 25))
	for _, seller := range sellers {
		fmt.Printf("\t%-25s\n", seller.ID)
	}
	fmt.Println()

}

func checkBuyersInBinFile() {
	f2, err := os.Open("database.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var buyers []routing.Buyer
	for _, buyer := range incomingDB.BuyerMap {
		buyers = append(buyers, buyer)
	}

	// print a list
	sort.SliceStable(buyers, func(i, j int) bool {
		return buyers[i].ShortName < buyers[j].ShortName
	})

	fmt.Println("Buyers:")
	fmt.Printf("\t%-25s %-16s  %-5s\n", "ShortName", "ID", "Live")
	fmt.Printf("\t%s\n", strings.Repeat("-", 50))
	for _, buyer := range buyers {
		id := strings.ToUpper(fmt.Sprintf("%016x", buyer.ID))
		fmt.Printf("\t%-25s %016s %5t\n", buyer.ShortName, id, buyer.Live)
	}
	fmt.Println()
}

func checkDCMapsInBinFile() {
	f2, err := os.Open("database.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for buyerID, buyer := range incomingDB.BuyerMap {
		if dcMaps, ok := incomingDB.DatacenterMaps[buyerID]; ok {
			fmt.Printf("\t%s:\n", buyer.ShortName)
			fmt.Printf("\t\t%-25s %-16s\n", "Datacenter", "Alias")
			fmt.Printf("\t\t%s\n", strings.Repeat("-", 50))
			for _, dcMap := range dcMaps {

				dcName := "name not found"
				if datacenter, ok := incomingDB.DatacenterMap[dcMap.DatacenterID]; ok {
					dcName = datacenter.Name
				}
				fmt.Printf("\t\t%-25s\n", dcName)
			}
			fmt.Println()
		}
	}

}

func commitDatabaseBin(env Environment) {

	// dev    : development_artifacts
	// prod   : prod_artifacts
	// staging: staging_artifacts

	bucketName := "gs://"

	switch env.Name {
	case "dev":
		bucketName += "dev_database_bin"
	case "prod":
		bucketName += "prod_database_bin"
	case "staging":
		bucketName += "staging_database_bin"
	case "local":
		fmt.Println("No need to commit database.bin for the happy path.")
		os.Exit(0)
	}

	if _, err := os.Stat("./database.bin"); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Local file database.bin does not exist.")
		os.Exit(0)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("This command will copy database.bin to %s\n", bucketName)
	fmt.Println("Are you sure you want to do this? (N/y)")
	fmt.Print("-> ")

	answer, _ := reader.ReadString('\n')
	answer = strings.Replace(answer, "\n", "", -1)

	if strings.Compare("y", answer) == 0 {
		// make a local copy in case things go pear-shaped
		// gsutil cp gs://development_artifacts/database.bin ./database.bin.remote
		remoteFileName := bucketName + "/database.bin"
		localCopy := fmt.Sprintf("database.bin.%d", time.Now().Unix())
		gsutilCpCommand := exec.Command("gsutil", "cp", remoteFileName, localCopy)

		err := gsutilCpCommand.Run()
		if err != nil {
			fmt.Println("Remote database.bin file does not exist (!!), so no local backup made.")
		}

		// gsutil cp database.bin gs://${bucketName}
		gsutilCpCommand = exec.Command("gsutil", "cp", "database.bin", bucketName)

		err = gsutilCpCommand.Run()
		if err != nil {
			handleRunTimeError(fmt.Sprintf("Error copying database.bin to %s: %v\n", bucketName, err), 1)
		}

		var timeStampArgs = localjsonrpc.NextBinFileCommitTimeStampArgs{}
		var timeStampReply = localjsonrpc.NextBinFileCommitTimeStampReply{}

		if err := makeRPCCall(env, &timeStampReply, "RelayFleetService.NextBinFileCommitTimeStamp", timeStampArgs); err != nil {
			handleJSONRPCError(env, err)
			return
		}

		fmt.Printf("\ndatabase.bin copied to %s.\n", bucketName)
	} else {
		fmt.Printf("\nOk - not pushing database.bin to %s\n", bucketName)
	}

}
