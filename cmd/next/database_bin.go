package main

import (
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/networknext/backend/modules/routing"
)

func getDatabaseBin(env Environment) {
	var err error

	uri := fmt.Sprintf("%s/database.bin", env.PortalHostname())

	// GET doesn't seem to like env.PortalHostname() for local
	if env.Name == "local" {
		uri = "http://127.0.0.1:20000/database.bin"
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", env.AuthToken))

	r, err := client.Do(req)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not get database.bin from the portal: %v\n", err), 1)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		handleRunTimeError(fmt.Sprintf("the portal returned an error response code: %d\n", r.StatusCode), 1)
	}

	file, err := os.Create("database.bin")
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open database.bin file for writing: %v\n", err), 1)
	}
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		handleRunTimeError(fmt.Sprintf("error writing data to database.bin: %v\n", err), 1)
	}

	f, err := os.Stat("./database.bin")
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not find database.bin? %v\n", err), 1)
	}

	fileSize := f.Size()
	fmt.Printf("Successfully retrieved ./database.bin (%d bytes)\n", fileSize)

}

func checkRelaysInBinFile() {

	f2, err := os.Open("database.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// print a list
	sort.SliceStable(incomingDB.Relays, func(i, j int) bool {
		return incomingDB.Relays[i].Name < incomingDB.Relays[j].Name
	})

	fmt.Println("Relays:")
	fmt.Printf("\t%-25s %-18s %-17s\n", "Name", "ID", "Address")
	fmt.Printf("\t%s\n", strings.Repeat("-", 61))
	for _, relay := range incomingDB.Relays {
		id := strings.ToUpper(fmt.Sprintf("%016x", relay.ID))
		fmt.Printf("\t%-25s %016s %17s\n", relay.Name, id, relay.Addr.String())
	}
	fmt.Println()

}

func checkDatacentersInBinFile() {
	f2, err := os.Open("database.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var datacenters []routing.Datacenter
	for _, datacenter := range incomingDB.DatacenterMap {
		datacenters = append(datacenters, datacenter)
	}

	// print a list
	sort.SliceStable(datacenters, func(i, j int) bool {
		return datacenters[i].Name < datacenters[j].Name
	})

	fmt.Println("Datacenters:")
	fmt.Printf("\t%-25s %-16s\n", "Name", "ID")
	fmt.Printf("\t%s\n", strings.Repeat("-", 43))
	for _, datacenter := range datacenters {
		id := strings.ToUpper(fmt.Sprintf("%016x", datacenter.ID))
		fmt.Printf("\t%-25s %016s\n", datacenter.Name, id)
	}
	fmt.Println()
}

func checkSellersInBinFile() {
	f2, err := os.Open("database.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var sellers []routing.Seller
	for _, seller := range incomingDB.SellerMap {
		sellers = append(sellers, seller)
	}

	// print a list
	sort.SliceStable(sellers, func(i, j int) bool {
		return sellers[i].ID < sellers[j].ID
	})

	fmt.Println("Sellers:")
	fmt.Printf("\t%-25s\n", "ID")
	fmt.Printf("\t%s\n", strings.Repeat("-", 25))
	for _, seller := range sellers {
		fmt.Printf("\t%-25s\n", seller.ID)
	}
	fmt.Println()

}

func checkBuyersInBinFile() {
	f2, err := os.Open("database.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var buyers []routing.Buyer
	for _, buyer := range incomingDB.BuyerMap {
		buyers = append(buyers, buyer)
	}

	// print a list
	sort.SliceStable(buyers, func(i, j int) bool {
		return buyers[i].ShortName < buyers[j].ShortName
	})

	fmt.Println("Buyers:")
	fmt.Printf("\t%-25s %-16s  %-5s\n", "ShortName", "ID", "Live")
	fmt.Printf("\t%s\n", strings.Repeat("-", 50))
	for _, buyer := range buyers {
		id := strings.ToUpper(fmt.Sprintf("%016x", buyer.ID))
		fmt.Printf("\t%-25s %016s %5t\n", buyer.ShortName, id, buyer.Live)
	}
	fmt.Println()
}

func checkDCMapsInBinFile() {
	f2, err := os.Open("database.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingDB routing.DatabaseBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingDB)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for buyerID, buyer := range incomingDB.BuyerMap {
		if dcMaps, ok := incomingDB.DatacenterMaps[buyerID]; ok {
			fmt.Printf("\t%s:\n", buyer.ShortName)
			fmt.Printf("\t\t%-25s %-16s\n", "Datacenter", "Alias")
			fmt.Printf("\t\t%s\n", strings.Repeat("-", 50))
			for _, dcMap := range dcMaps {

				dcName := "name not found"
				if datacenter, ok := incomingDB.DatacenterMap[dcMap.DatacenterID]; ok {
					dcName = datacenter.Name
				}
				fmt.Printf("\t\t%-25s %-16s\n", dcName, dcMap.Alias)
			}
			fmt.Println()
		}
	}

}

func commitDatabaseBin(env Environment) {

	// dev    : development_artifacts
	// prod   : prod_artifacts
	// staging: staging_artifacts

	bucketName := "gs://"

	switch env.Name {
	case "dev":
		bucketName += "dev_database_bin"
	case "prod":
		bucketName += "prod_database_bin"
	case "staging":
		bucketName += "staging_database_bin"
	case "local":
		fmt.Println("No need to commit database.bin for the happy path.")
		os.Exit(0)
	}

	if _, err := os.Stat("./database.bin"); errors.Is(err, os.ErrNotExist) {
		fmt.Println("Local file database.bin does not exist.")
		os.Exit(0)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("This command will copy database.bin to %s\n", bucketName)
	fmt.Println("Are you sure you want to do this? (N/y)")
	fmt.Print("-> ")

	answer, _ := reader.ReadString('\n')
	answer = strings.Replace(answer, "\n", "", -1)

	if strings.Compare("y", answer) == 0 {
		// make a local copy in case things go pear-shaped
		// gsutil cp gs://development_artifacts/database.bin ./database.bin.remote
		remoteFileName := bucketName + "/database.bin"
		localCopy := fmt.Sprintf("database.bin.%d", time.Now().Unix())
		gsutilCpCommand := exec.Command("gsutil", "cp", remoteFileName, localCopy)

		err := gsutilCpCommand.Run()
		if err != nil {
			fmt.Println("Remote database.bin file does not exist (!!), so no local backup made.")
		}

		// gsutil cp database.bin gs://${bucketName}
		gsutilCpCommand = exec.Command("gsutil", "cp", "database.bin", bucketName)

		err = gsutilCpCommand.Run()
		if err != nil {
			handleRunTimeError(fmt.Sprintf("Error copying database.bin to %s: %v\n", bucketName, err), 1)
		}

		fmt.Printf("\ndatabase.bin copied to %s.\n", bucketName)
	} else {
		fmt.Printf("\nOk - not pushing database.bin to %s\n", bucketName)
	}

}
