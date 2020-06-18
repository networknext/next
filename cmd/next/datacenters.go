package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/modood/table"
	"github.com/networknext/backend/routing"
	localjsonrpc "github.com/networknext/backend/transport/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

type datacenter struct {
	Name      string
	Enabled   bool
	Latitude  float64
	Longitude float64
}

func datacenters(rpcClient jsonrpc.RPCClient, env Environment, filter string) {
	args := localjsonrpc.DatacentersArgs{
		Name: filter,
	}

	var reply localjsonrpc.DatacentersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Datacenters", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	datacenters := make([]struct {
		Name      string
		ID        string
		Latitude  float64
		Longitude float64
		Enabled   bool
	}, len(reply.Datacenters))

	for i := 0; i < len(datacenters); i++ {
		datacenters[i] = struct {
			Name      string
			ID        string
			Latitude  float64
			Longitude float64
			Enabled   bool
		}{
			Name:      reply.Datacenters[i].Name,
			ID:        fmt.Sprintf("%x", reply.Datacenters[i].ID),
			Latitude:  reply.Datacenters[i].Latitude,
			Longitude: reply.Datacenters[i].Longitude,
			Enabled:   reply.Datacenters[i].Enabled,
		}
	}

	table.Output(datacenters)
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

	fmt.Printf("Datacenter \"%s\" removed from storage.\n", name)
}

func viewDatacenter(rpcClient jsonrpc.RPCClient, env Environment, name string) {
	args := localjsonrpc.DatacentersArgs{
		Name: name,
	}

	var reply localjsonrpc.DatacentersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Datacenters", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	if len(reply.Datacenters) == 0 {
		log.Fatalf("Could not find datacenter with name %s", name)
	}

	if len(reply.Datacenters) > 1 {
		log.Fatalf("Found more than one datacenter matching %s", name)
	}

	datacenter := datacenter{
		Name:      reply.Datacenters[0].Name,
		Enabled:   reply.Datacenters[0].Enabled,
		Latitude:  reply.Datacenters[0].Latitude,
		Longitude: reply.Datacenters[0].Longitude,
	}

	jsonData, err := json.MarshalIndent(datacenter, "", "\t")
	if err != nil {
		log.Fatalf("Could not marshal json data for datacenter: %v", err)
	}

	fmt.Println(string(jsonData))
}

func editDatacenter(rpcClient jsonrpc.RPCClient, env Environment, name string, datacenterData map[string]interface{}) {
	args := localjsonrpc.DatacentersArgs{
		Name: name,
	}

	var reply localjsonrpc.DatacentersReply
	if err := rpcClient.CallFor(&reply, "OpsService.Datacenters", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	if len(reply.Datacenters) == 0 {
		log.Fatalf("Could not find datacenter with name %s", name)
	}

	if len(reply.Datacenters) > 1 {
		log.Fatalf("Found more than one datacenter matching %s", name)
	}

	editArgs := localjsonrpc.DatacenterEditArgs{
		DatacenterID:   reply.Datacenters[0].ID,
		DatacenterData: datacenterData,
	}

	var editReply localjsonrpc.DatacenterEditReply
	if err := rpcClient.CallFor(&editReply, "OpsService.DatacenterEdit", editArgs); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	fmt.Printf("Datacenter %s edited\n", reply.Datacenters[0].Name)
}
