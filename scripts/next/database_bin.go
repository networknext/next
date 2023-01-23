package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	// todo: we don't wan to be using old modules
	"github.com/networknext/backend/modules-old/backend"
	"github.com/networknext/backend/modules-old/routing"
	localjsonrpc "github.com/networknext/backend/modules-old/transport/jsonrpc"
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
