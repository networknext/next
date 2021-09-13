package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"net"
	"os"
	"sort"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/storage"
)

// Chose the utility function to use
func main() {
	// DatacenterReverseLookup()

	// // Set these variables depending on the environment
	// // Remember to also export GOOGLE_APPLICATION_CREDENTIALS env var
	// gcpProjectID := "local"
	// btInstanceID := "localhost:8086"
	// btTableName := "portal-session-history"
	// prefix := "prefix_of_rows_to_delete_goes_here"
	// DeleteBigtableRows(gcpProjectID, btInstanceID, btTableName, prefix)

	err := ExamineUnknowns("./servers_2_2021-09-13.json")
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
}
	
func ExamineUnknowns(fileName string) error {
	// buyerID -> Server IP (w/ Port) -> ServerInfo
	buyerMap := make(map[string]map[string]storage.ServerInfo)

	// Read in the JSON
	data, _ := ioutil.ReadFile(fileName)
	
	err := json.Unmarshal(data, &buyerMap)
	if err != nil {
		return err
	}


	type DCMapTime struct{
		Address string
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
						Address: addr.String(),
						Timestamp: info.Timestamp,
					})
				buyerDCMap[buyerID][info.DatacenterName] = dcList
			} else {
				// First time seeing this datacenter
				var dcList []DCMapTime
				buyerDCMap[buyerID][info.DatacenterName] = append(dcList, DCMapTime{
					Address: addr.String(),
					Timestamp: info.Timestamp,
				})
			}
		}
	}

	for buyerID, dcNames := range buyerDCMap {
		fmt.Printf("Datacenters for %s\n", buyerID)
		for dcName, serverInfo := range dcNames {
			fmt.Printf("\t%s\n", dcName)

			// Sort the server info
			sort.SliceStable(serverInfo, func(i, j int) bool {
				return serverInfo[i].Timestamp > serverInfo[j].Timestamp
			})

			for _, info := range serverInfo {
				fmt.Printf("\t\t%s\n", info.Address)
			}			
		}
		fmt.Println()
	}

	jsonData, err := json.MarshalIndent(buyerDCMap, "", "\t")
	if err != nil {
		return err
	}

	jsonFile, err := os.Create(fmt.Sprintf("./parsed_%s", fileName[2:]))
	if err != nil {
		return err
	}

	jsonFile.Write(jsonData)

	fmt.Printf("Wrote JSON output to %s\n", jsonFile.Name())

	jsonFile.Close()

	// Write as CSV also

	csvFile, err := os.Create(fmt.Sprintf("./parsed_%s.csv", fileName[2:len(fileName)-5]))
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
				addrRow = append(addrRow, fmt.Sprintf("%d",info.Timestamp))

				writer.Write(addrRow)
			}			
		}
	}

	writer.Flush()

	fmt.Printf("Wrote CSV output to %s\n", csvFile.Name())

	return nil
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
