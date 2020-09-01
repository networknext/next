package main

import (
	"fmt"
	"regexp"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

type datacenterReply struct {
	Name         string
	ID           string
	Latitude     float64
	Longitude    float64
	Enabled      bool
	SupplierName string
}

func datacenters(rpcClient jsonrpc.RPCClient, env Environment, filter string, signed bool) {
	args := localjsonrpc.DatacentersArgs{
		Name: filter,
	}

	var reply localjsonrpc.DatacentersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Datacenters", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	var dcs []datacenterReply

	if signed {
		for _, dc := range reply.Datacenters {
			dcs = append(dcs, datacenterReply{
				Name: dc.Name,
				// ID:           fmt.Sprintf("%d", dc.SignedID), // ToDo: could come from storage (exists in firestore)
				ID:           fmt.Sprintf("%d", int64(dc.ID)),
				Latitude:     dc.Latitude,
				Longitude:    dc.Longitude,
				Enabled:      dc.Enabled,
				SupplierName: dc.SupplierName,
			})
		}
	} else {
		for _, dc := range reply.Datacenters {
			dcs = append(dcs, datacenterReply{
				Name:         dc.Name,
				ID:           fmt.Sprintf("%016x", dc.ID),
				Latitude:     dc.Latitude,
				Longitude:    dc.Longitude,
				Enabled:      dc.Enabled,
				SupplierName: dc.SupplierName,
			})
		}
	}

	table.Output(dcs)
}

func addDatacenter(rpcClient jsonrpc.RPCClient, env Environment, datacenter routing.Datacenter) {
	args := localjsonrpc.AddDatacenterArgs{
		Datacenter: datacenter,
	}

	var reply localjsonrpc.AddDatacenterReply
	if err := rpcClient.CallFor(&reply, "OpsService.AddDatacenter", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Datacenter \"%s\" added to storage.\n", datacenter.Name)
}

func removeDatacenter(rpcClient jsonrpc.RPCClient, env Environment, name string) {
	args := localjsonrpc.RemoveDatacenterArgs{
		Name: name,
	}

	var reply localjsonrpc.RemoveDatacenterReply
	if err := rpcClient.CallFor(&reply, "OpsService.RemoveDatacenter", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Datacenter \"%x\" removed from storage.\n", name)
}

func listDatacenterMaps(rpcClient jsonrpc.RPCClient, env Environment, datacenter string) {

	var dcIDs []uint64
	var err error

	// get list of datacenters matching the given id/name/substring
	datacentersArgs := localjsonrpc.DatacentersArgs{}
	var datacenters localjsonrpc.DatacentersReply
	if err = rpcClient.CallFor(&datacenters, "OpsService.Datacenters", datacentersArgs); err != nil {
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

		if err := rpcClient.CallFor(&reply, "OpsService.ListDatacenterMaps", arg); err != nil {
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
