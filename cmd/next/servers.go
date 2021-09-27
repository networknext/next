package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/storage"
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

func parseLiveServerOutput(
	fileName string,
) error {
	// buyerID -> Server IP (w/ Port) -> ServerInfo
	buyerMap := make(map[string]map[string]storage.ServerInfo)

	// Read in the JSON
	data, _ := ioutil.ReadFile(fileName)

	err := json.Unmarshal(data, &buyerMap)
	if err != nil {
		return err
	}

	type DCMapTime struct {
		Address   string
		Timestamp int64
	}

	// buyerID -> DCName -> DCMapTime
	buyerDCMap := make(map[string]map[string][]DCMapTime)

	for buyerID, serverIPs := range buyerMap {
		buyerDCMap[buyerID] = make(map[string][]DCMapTime)

		for serverIP, info := range serverIPs {
			addr := core.ParseAddress(serverIP)
			if dcList, ok := buyerDCMap[buyerID][info.DatacenterName]; ok {
				dcList = append(dcList, DCMapTime{
					Address:   addr.String(),
					Timestamp: info.Timestamp,
				})
				buyerDCMap[buyerID][info.DatacenterName] = dcList
			} else {
				// First time seeing this datacenter
				var dcList []DCMapTime
				buyerDCMap[buyerID][info.DatacenterName] = append(dcList, DCMapTime{
					Address:   addr.String(),
					Timestamp: info.Timestamp,
				})
			}
		}
	}

	for _, dcNames := range buyerDCMap {
		for _, serverInfo := range dcNames {
			// Sort the server info
			sort.SliceStable(serverInfo, func(i, j int) bool {
				return serverInfo[i].Timestamp > serverInfo[j].Timestamp
			})
		}
	}

	jsonData, err := json.MarshalIndent(buyerDCMap, "", "\t")
	if err != nil {
		return err
	}

	jsonFile, err := os.Create(fmt.Sprintf("./parsed_%s", fileName[0:]))
	if err != nil {
		return err
	}

	jsonFile.Write(jsonData)

	fmt.Printf("Wrote JSON output to %s\n", jsonFile.Name())

	jsonFile.Close()

	// Write as CSV also

	csvFile, err := os.Create(fmt.Sprintf("./parsed_%s.csv", fileName[0:len(fileName)-5]))
	if err != nil {
		return err
	}

	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)

	for buyerID, dcNames := range buyerDCMap {
		var buyerIDRow []string
		buyerIDRow = append(buyerIDRow, buyerID)

		writer.Write(buyerIDRow)

		for dcName, serverInfo := range dcNames {

			var dcRow []string
			dcRow = append(dcRow, "")
			dcRow = append(dcRow, dcName)

			writer.Write(dcRow)

			for _, info := range serverInfo {
				var addrRow []string
				addrRow = append(addrRow, "")
				addrRow = append(addrRow, "")

				host, port, err := net.SplitHostPort(info.Address)
				if err != nil {
					fmt.Printf("err splitting: %v\n", err)
					continue
				}
				addrRow = append(addrRow, host)
				addrRow = append(addrRow, port)
				addrRow = append(addrRow, fmt.Sprintf("%d", info.Timestamp))

				writer.Write(addrRow)
			}
		}
	}

	writer.Flush()

	fmt.Printf("Wrote CSV output to %s\n", csvFile.Name())

	return nil
}
