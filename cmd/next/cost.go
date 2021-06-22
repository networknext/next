package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
)

func saveCostMatrix(env Environment, filename string) {
	var uri string
	var err error

	if uri, err = env.RelayBackendURL(); err != nil {
		handleRunTimeError(fmt.Sprintf("Cannot get get relay backend hostname: %v\n", err), 1)
	}

	uri += "/cost_matrix"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", env.AuthToken))

	r, err := client.Do(req)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not get the route matrix from the backend: %v\n", err), 1)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		handleRunTimeError(fmt.Sprintf("relay backend return an error response code: %d\n", r.StatusCode), 1)
	}

	file, err := os.Create(filename)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open file for writing: %v\n", err), 1)
	}
	defer file.Close()

	fmt.Printf("%v\n", r.Body)

	if _, err := io.Copy(file, r.Body); err != nil {
		handleRunTimeError(fmt.Sprintf("error writing cost matrix to file: %v\n", err), 1)
	}
}

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
