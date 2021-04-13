package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
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

	for _, relay := range incomingDB.Relays {
		id := strings.ToUpper(fmt.Sprintf("%016x", relay.ID))
		fmt.Printf("%-25s 0x%016s %17s\n", relay.Name, id, relay.Addr.String())
	}

	// generate a csv file
	relaysCSV := [][]string{{}}

	relaysCSV = append(relaysCSV, []string{
		"Name", "MRC", "Overage", "BW Rule",
		"Term", "Start Date", "End Date", "Type", "Bandwidth", "NIC Speed", "State", "IP Address", "Notes"})

	for _, relay := range incomingDB.Relays {
		mrc := ""
		if relay.MRC > 0 {
			mrc = fmt.Sprintf("%.2f", relay.MRC.ToCents()/100)
		}
		overage := ""
		if relay.Overage > 0 {
			overage = fmt.Sprintf("%.5f", relay.Overage.ToCents()/100)
		}

		var bwRule string
		switch relay.BWRule {
		case routing.BWRuleNone:
			bwRule = ""
		case routing.BWRuleFlat:
			bwRule = "flat"
		case routing.BWRuleBurst:
			bwRule = "burst"
		case routing.BWRulePool:
			bwRule = "pool"
		default:
			bwRule = ""
		}

		var machineType string
		switch relay.Type {
		case routing.NoneSpecified:
			machineType = ""
		case routing.BareMetal:
			machineType = "bare metal"
		case routing.VirtualMachine:
			machineType = "virtual machine"
		default:
			machineType = ""
		}

		contractTerm := ""
		if relay.ContractTerm != 0 {
			contractTerm = fmt.Sprintf("%d", relay.ContractTerm)
		}

		startDate := ""
		if !relay.StartDate.IsZero() {
			startDate = relay.StartDate.Format("January 2, 2006")
		}

		endDate := ""
		if !relay.EndDate.IsZero() {
			endDate = relay.EndDate.Format("January 2, 2006")
		}

		bandwidth := strconv.FormatInt(int64(relay.IncludedBandwidthGB), 10)
		if bandwidth == "0" {
			bandwidth = ""
		}

		nicSpeed := strconv.FormatInt(int64(relay.NICSpeedMbps), 10)
		if nicSpeed == "0" {
			nicSpeed = ""
		}

		relaysCSV = append(relaysCSV, []string{
			relay.Name,
			mrc,
			overage,
			bwRule,
			contractTerm,
			startDate,
			endDate,
			machineType,
			bandwidth,
			nicSpeed,
			relay.State.String(),
			relay.Addr.String(),
			relay.Notes,
		})
	}

	fileName := "./relays_bin.csv"
	f, err := os.Create(fileName)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("Error creating local CSV file %s: %v\n", fileName, err), 1)
	}

	writer := csv.NewWriter(f)
	err = writer.WriteAll(relaysCSV)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("Error writing local CSV file %s: %v\n", fileName, err), 1)
	}
	fmt.Printf("\nCSV file written: relays_bin.csv\n")

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
