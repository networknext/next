package main

import (
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/networknext/backend/modules/routing"
)

func getRelaysBin(env Environment, filename string) {
	var err error

	uri := fmt.Sprintf("%s/relays.bin", env.PortalHostname())

	// GET doesn't seem to like env.PortalHostname() for local
	if env.Name == "local" {
		uri = "http://127.0.0.1:20000/relays.bin"
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", env.AuthToken))

	r, err := client.Do(req)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not get relays.bin from the portal: %v\n", err), 1)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		handleRunTimeError(fmt.Sprintf("the portal returned an error response code: %d\n", r.StatusCode), 1)
	}

	file, err := os.Create(filename)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open file for writing: %v\n", err), 1)
	}
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		handleRunTimeError(fmt.Sprintf("error writing data to relays.bin: %v\n", err), 1)
	}

	f, err := os.Stat("./relays.bin")
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not find relays.bin? %v\n", err), 1)
	}

	fileSize := f.Size()
	fmt.Printf("Successfully retrieved ./relays.bin (%d bytes)\n", fileSize)

}

func checkRelaysBin() {

	f2, err := os.Open("relays.bin")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f2.Close()

	var incomingRelays routing.RelayBinWrapper

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingRelays)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// print a list
	sort.SliceStable(incomingRelays.Relays, func(i, j int) bool {
		return incomingRelays.Relays[i].Name < incomingRelays.Relays[j].Name
	})

	for _, relay := range incomingRelays.Relays {
		id := strings.ToUpper(fmt.Sprintf("%016x", relay.ID))
		fmt.Printf("%-25s 0x%016s %17s\n", relay.Name, id, relay.Addr.String())
	}

	// generate a csv file
	relaysCSV := [][]string{{}}

	relaysCSV = append(relaysCSV, []string{
		"Name", "MRC", "Overage", "BW Rule",
		"Term", "Start Date", "End Date", "Type", "Bandwidth", "NIC Speed", "State", "IP Address", "Notes"})

	for _, relay := range incomingRelays.Relays {
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
