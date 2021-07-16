package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	localjsonrpc "github.com/networknext/backend/modules/transport/jsonrpc"
)

func getLiveServers(
	env Environment,
	serverBackendIPs []string,
) {

	var reply localjsonrpc.LiveServerReply = localjsonrpc.LiveServerReply{}
	var args = localjsonrpc.LiveServerArgs{ServerBackendIPs: serverBackendIPs}

	if err := makeRPCCall(env, &reply, "LiveServerService.LiveServers", args); err != nil {
		handleJSONRPCError(env, err)
		return
	}

	// Write each json to local file
	for i, tracker := range reply.ServerTrackers {
		jsonData, err := json.MarshalIndent(tracker.Tracker, "", "\t")
		if err != nil {
			handleJSONRPCError(env, err)
			continue
		}

		currentDate := time.Now().Local().Format("2006-01-02")

		jsonFile, err := os.Create(fmt.Sprintf("./servers_%d_%s.json", i, currentDate))
		if err != nil {
			handleJSONRPCError(env, err)
			continue
		}

		jsonFile.Write(jsonData)

		fmt.Printf("Wrote JSON output to %s\n", jsonFile.Name())

		jsonFile.Close()
	}
}
