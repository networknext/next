package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
)

// Chose the utility function to use
func main() {
	DatacenterReverseLookup()

	// Set these variables depending on the environment
	// Remember to also export GOOGLE_APPLICATION_CREDENTIALS env var
	gcpProjectID := "local"
	btInstanceID := "localhost:8086"
	btTableName := "portal-session-history"
	prefix := "prefix_of_rows_to_delete_goes_here"
	DeleteBigtableRows(gcpProjectID, btInstanceID, btTableName, prefix)

	// Provide the external IPs of server backend instances
	serverBackendIPs := []string{"http://34.121.78.228", "http://35.238.240.228"}
	// Provide the prod database.bin path
	prodDatabaseBinPath := "./database.bin"
	err := GetLiveServers(serverBackendIPs, prodDatabaseBinPath)
	if err != nil {
		fmt.Printf("err: %v", err)
	}
}

// Fill in the "matches" string slice with datacenter hashes you want to brute force search.
// The program will pair up all known suppliers with a list of common cities to try and find a match.
func DatacenterReverseLookup() {
	cityNames := []string{
		"",
		"amsterdam",
		"atlanta",
		"beijing",
		"bangalore",
		"chennai",
		"chicago",
		"copenhagen",
		"dallas",
		"daressalaam",
		"delhi",
		"denpasar",
		"dubai",
		"eindhoven",
		"florida",
		"frankfurt",
		"fujairah",
		"heerlen",
		"hongkong",
		"hyderabad",
		"jakarta",
		"johannesburg",
		"kolkata",
		"kuala",
		"kyiv",
		"london",
		"losangeles",
		"luanda",
		"luxembourg",
		"madrid",
		"manila",
		"mexico",
		"miami",
		"montreal",
		"moscow",
		"mumbai",
		"newdelhi",
		"newjersey",
		"newyork",
		"osaka",
		"paris",
		"perth",
		"phoenix",
		"rotterdam",
		"saintlouis",
		"saintpetersburg",
		"saltlakecity",
		"sanjose",
		"santaclara",
		"sao",
		"saopaolo",
		"saopaulo",
		"seattle",
		"semarang",
		"seoul",
		"singapore",
		"shanghai",
		"stlouis",
		"stpetersburg",
		"stockholm",
		"strasbourg",
		"surabaya",
		"sydney",
		"tokyo",
		"toronto",
		"vancouver",
		"warsaw",
		"washingtondc",
	}

	supplierNames := []string{
		"",
		"multiplay",
		"100tb",
		"amazon",
		"atman",
		"azure",
		"buzinessware",
		"colocrossing",
		"datapacket",
		"dedicatedsolutions",
		"dinahosting",
		"gcore",
		"glesys",
		"google",
		"handynetworks",
		"hostdime",
		"hosthink",
		"hostirian",
		"hostroyale",
		"i3d",
		"ibm",
		"inap",
		"intergrid",
		"kamatera",
		"leaseweb",
		"limelight",
		"maxihost",
		"netdedi",
		"nforce",
		"opencolo",
		"ovh",
		"packet",
		"pallada",
		"phoenixnap",
		"psychz",
		"radore",
		"riot",
		"selectel",
		"serversaustralia",
		"serversdotcom",
		"serverzoo",
		"singlehop",
		"springshosting",
		"stackpath",
		"totalserversolutions",
		"vultr",
		"zenlayer",
	}

	matches := []string{
		"6d7c3a02b218131d",
		"83462735ba7ef971",
		"c38c805d821cd0d3",
		"f45cb29b3b74e986",
	}

	oneMatchFound := false
	for _, supplierName := range supplierNames {
		for _, cityName := range cityNames {
			datacenterName := supplierName + "." + cityName

			for _, match := range matches {
				if match == fmt.Sprintf("%016x", crypto.HashID(datacenterName)) {
					fmt.Printf("match found! datacenter name for %s is %s\n", match, datacenterName)
					oneMatchFound = true
				}
			}
		}
	}

	if !oneMatchFound {
		fmt.Println("no matches found")
	}
}

// Delete rows from a Bigtable instance based on a prefix
func DeleteBigtableRows(gcpProjectID, btInstanceID, btTableName, prefix string) {
	ctx := context.Background()
	logger := log.NewNopLogger()

	if os.Getenv("BIGTABLE_EMULATOR_HOST") != "" {
		fmt.Println("Detected Bigtable emulator")
	}

	// Get a bigtable admin client
	btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID, logger)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	defer func() {
		// Close the admin client
		err = btAdmin.Close()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}()

	// Verify table exists
	exists, err := btAdmin.VerifyTableExists(ctx, btTableName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	if !exists {
		fmt.Printf("Table %s does not exist in instance %s. Aborting.\n", btTableName, btInstanceID)
		return
	}

	// Delete rows with prefix from table
	err = btAdmin.DropRowsByPrefix(ctx, btTableName, prefix)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Successfully deleted rows with prefix %s from table %s\n", prefix, btTableName)
}

// Gets the datacenter names, IPs, and timestamps for all servers connected to server backend instances per buyer
func GetLiveServers(serverBackendIPs []string, databaseBinPath string) error {
	type DatacenterInfo struct {
		Timestamp uint64
		ServerIPs []string
	}

	// Output mapping will be {Buyer Name: {Datacenter Name: [Timestamp, IP]}}
	output := make(map[string]map[string]DatacenterInfo)

	client := &http.Client{Timeout: 5 * time.Second}

	var trackers []storage.ServerTracker

	// Load in JSON from server backend's /server endpoint
	for _, serverBackendIP := range serverBackendIPs {
		endpoint := fmt.Sprintf("%s/servers", serverBackendIP)

		r, err := client.Get(endpoint)
		if err != nil {
			return err
		}

		tracker := storage.NewServerTracker()

		json.NewDecoder(r.Body).Decode(&tracker.Tracker)
		r.Body.Close()

		trackers = append(trackers, *tracker)
	}

	// Load in database.bin
	f2, err := os.Open(databaseBinPath)
	if err != nil {
		return err
	}

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		return err
	}

	f2.Close()

	// Create map of buyer hex ID to buyer
	buyerMap := make(map[string]routing.Buyer)
	for _, buyer := range incomingDB.BuyerMap {
		hexID := fmt.Sprintf("%016x", buyer.ID)
		buyerMap[hexID] = buyer
	}

	// Create map of datacenter hex ID datacneter
	datacenterMap := make(map[string]routing.Datacenter)
	for _, dc := range incomingDB.DatacenterMap {
		hexID := fmt.Sprintf("%016x", dc.ID)
		datacenterMap[hexID] = dc
	}

	// Loop through trackers and add to output mapping
	for _, tracker := range trackers {
		for buyerHexID, ipMapping := range tracker.Tracker {
			buyer, ok := buyerMap[buyerHexID]
			if !ok {
				return fmt.Errorf("buyer %s does not exist in buyer map", buyerHexID)
			}

			buyerName := buyer.CompanyCode

			dcMap := make(map[string]DatacenterInfo)

			for serverIP, serverInfo := range ipMapping {
				host := strings.Split(serverIP, ":")[0]
				datacenterHexID := serverInfo.DatacenterID
				datacenter, ok := datacenterMap[datacenterHexID]

				var datacenterName string
				if !ok {
					datacenterName = fmt.Sprintf("unknown - %s", host)
					fmt.Printf("datacenter %s does not exist in datacenter map. Buyer Name: %s, Server IP: %s\n", datacenterHexID, buyerName, serverIP)
				} else {
					datacenterName = datacenter.Name
				}

				var info DatacenterInfo
				var exists bool
				info, exists = dcMap[datacenterName]
				if !exists {
					// First time we see this datacenter for this buyer
					dcMap[datacenterName] = DatacenterInfo{
						Timestamp: serverInfo.Timestamp,
						ServerIPs: []string{host},
					}
				} else {
					// Use the latest timestamp
					if serverInfo.Timestamp > info.Timestamp {
						info.Timestamp = serverInfo.Timestamp
					}

					// Don't add duplicate server IPs
					unionMap := make(map[string]bool)
					for _, existingServerIP := range info.ServerIPs {
						unionMap[existingServerIP] = true
					}

					if _, alreadyExists := unionMap[host]; !alreadyExists {
						info.ServerIPs = append(info.ServerIPs, host)
					}

					dcMap[datacenterName] = info
				}
			}

			// Add the non-duplicate values of the datacenter map for the buyer
			if existingDcMapInfo, exists := output[buyerName]; !exists {
				output[buyerName] = dcMap
			} else {
				// Need to de-dupe values via iteration
				for datacenterName, existingDcInfo := range existingDcMapInfo {
					if dcInfo, ok := dcMap[datacenterName]; ok {
						if dcInfo.Timestamp > existingDcInfo.Timestamp {
							existingDcInfo.Timestamp = dcInfo.Timestamp
						}

						unionMap := make(map[string]bool)
						for _, existingServerIP := range existingDcInfo.ServerIPs {
							unionMap[existingServerIP] = true
						}

						for _, serverIP := range dcInfo.ServerIPs {
							if _, alreadyExists := unionMap[serverIP]; !alreadyExists {
								existingDcInfo.ServerIPs = append(existingDcInfo.ServerIPs, serverIP)
							}
						}
					}

					output[buyerName][datacenterName] = existingDcInfo
				}
			}
		}
	}

	// Save the file, one per buyer
	for buyer, liveServers := range output {
		jsonData, err := json.MarshalIndent(liveServers, "", "\t")
		if err != nil {
			return err
		}

		currentDate := time.Now().Local().Format("2006-01-02")

		jsonFile, err := os.Create(fmt.Sprintf("./datacenter_list_%s_%s.json", buyer, currentDate))
		if err != nil {
			return err
		}

		jsonFile.Write(jsonData)

		fmt.Printf("Wrote JSON output to %s\n", jsonFile.Name())

		jsonFile.Close()
	}

	return nil
}
