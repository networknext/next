package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"

	"github.com/modood/table"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
)

type datacenterReply struct {
	Name      string
	ID        string
	Latitude  float32
	Longitude float32
}

func datacenters(
	env Environment,
	filter string,
	signed bool,
	csvOutput bool,
) {
	args := localjsonrpc.DatacentersArgs{
		Name: filter,
	}

	datacentersCSV := [][]string{}

	datacentersCSV = append(datacentersCSV, []string{
		"Name", "HexID", "Latitude", "Longitude",
	})

	var reply localjsonrpc.DatacentersReply
	if err := makeRPCCall(env, &reply, "OpsService.Datacenters", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	if csvOutput {
		if signed {
			for _, dc := range reply.Datacenters {
				datacentersCSV = append(datacentersCSV, []string{
					dc.Name,
					fmt.Sprintf("%d", int64(dc.ID)),
					fmt.Sprintf("%.2f", dc.Latitude),
					fmt.Sprintf("%.2f", dc.Longitude),
				})
			}
		} else {
			for _, dc := range reply.Datacenters {
				datacentersCSV = append(datacentersCSV, []string{
					dc.Name,
					fmt.Sprintf("%d", dc.ID),
					fmt.Sprintf("%.2f", dc.Latitude),
					fmt.Sprintf("%.2f", dc.Longitude),
				})
			}
		}

		fileName := "./datacenters.csv"
		f, err := os.Create(fileName)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("Error creating local CSV file %s: %v\n", fileName, err), 1)
		}

		writer := csv.NewWriter(f)
		err = writer.WriteAll(datacentersCSV)
		if err != nil {
			handleRunTimeError(fmt.Sprintf("Error writing local CSV file %s: %v\n", fileName, err), 1)
		}
		fmt.Println("CSV file written: datacenters.csv")

	} else {

		var dcs []datacenterReply
		if signed {
			for _, dc := range reply.Datacenters {
				dcs = append(dcs, datacenterReply{
					Name: dc.Name,
					// ID:           fmt.Sprintf("%d", dc.SignedID), // ToDo: could come from storage (exists in firestore)
					ID:        fmt.Sprintf("%d", int64(dc.ID)),
					Latitude:  dc.Latitude,
					Longitude: dc.Longitude,
				})
			}
		} else {
			for _, dc := range reply.Datacenters {
				dcs = append(dcs, datacenterReply{
					Name:      dc.Name,
					ID:        fmt.Sprintf("%016x", dc.ID),
					Latitude:  dc.Latitude,
					Longitude: dc.Longitude,
				})
			}
		}
		table.Output(dcs)
	}

}

func addDatacenter(env Environment, dc datacenter) {

	var sellerReply localjsonrpc.SellerReply
	var sellerArg localjsonrpc.SellerArg

	sellerArg.ID = dc.SellerID
	if err := makeRPCCall(env, &sellerReply, "OpsService.Seller", sellerArg); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	did := crypto.HashID(dc.Name)
	datacenter := routing.Datacenter{
		ID:   did,
		Name: dc.Name,
		Location: routing.Location{
			Latitude:  dc.Latitude,
			Longitude: dc.Longitude,
		},
		SellerID: sellerReply.Seller.DatabaseID,
	}

	args := localjsonrpc.AddDatacenterArgs{
		Datacenter: datacenter,
	}

	var reply localjsonrpc.AddDatacenterReply
	if err := makeRPCCall(env, &reply, "OpsService.AddDatacenter", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Datacenter \"%s\" added to storage.\n", datacenter.Name)
}

func removeDatacenter(env Environment, name string) {
	args := localjsonrpc.RemoveDatacenterArgs{
		Name: name,
	}

	var reply localjsonrpc.RemoveDatacenterReply
	if err := makeRPCCall(env, &reply, "OpsService.RemoveDatacenter", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Datacenter \"%x\" removed from storage.\n", name)
}

func listDatacenterMaps(env Environment, datacenter string) {

	var dcIDs []uint64
	var err error

	// get list of datacenters matching the given id/name/substring
	datacentersArgs := localjsonrpc.DatacentersArgs{}
	var datacenters localjsonrpc.DatacentersReply
	if err = makeRPCCall(env, &datacenters, "OpsService.Datacenters", datacentersArgs); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	r := regexp.MustCompile("(?i)" + datacenter) // case-insensitive regex
	for _, dc := range datacenters.Datacenters {
		if r.MatchString(dc.Name) || r.MatchString(fmt.Sprintf("%016x", dc.ID)) {
			dcIDs = append(dcIDs, dc.ID)
		}
	}

	if len(dcIDs) == 0 {
		fmt.Printf("No match for provided datacenter ID: %v\n", datacenter)
		return
	}

	// assemble the full list of maps
	var dcMapsFull []localjsonrpc.DatacenterMapsFull
	for _, id := range dcIDs {
		var reply localjsonrpc.ListDatacenterMapsReply
		var arg = localjsonrpc.ListDatacenterMapsArgs{
			DatacenterID: id,
		}

		if err := makeRPCCall(env, &reply, "OpsService.ListDatacenterMaps", arg); err != nil {
			fmt.Printf("rpc error: %v\n", err)
			handleJSONRPCError(env, err)
			return
		}

		for _, dcMap := range reply.DatacenterMaps {
			dcMapsFull = append(dcMapsFull, dcMap)
		}
	}

	if len(dcMapsFull) == 0 {
		fmt.Printf("No buyers found for the provided datacenter ID: %v\n", datacenter)
		return
	}

	table.Output(dcMapsFull)

}
