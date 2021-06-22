package main

import (
	"fmt"
	"io/ioutil"

	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
)

func getCostMatrix(env Environment, fileName string) {
	args := localjsonrpc.NextCostMatrixHandlerArgs{}

	var reply localjsonrpc.NextCostMatrixHandlerReply
	if err := makeRPCCall(env, &reply, "RelayFleetService.NextCostMatrixHandler", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	err := ioutil.WriteFile(fileName, reply.CostMatrix, 0777)
	if err != nil {
		err := fmt.Errorf("getCostMatrix() error writing %s to filesystem: %v", fileName, err)
		handleRunTimeError(fmt.Sprintf("could not write %s to the filesystem: %v\n", fileName, err), 0)
	}

}
