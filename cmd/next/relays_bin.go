package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"

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
	var relayNames []string

	decoder := gob.NewDecoder(f2)
	err = decoder.Decode(&incomingRelays)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, relay := range incomingRelays.Relays {
		relayNames = append(relayNames, relay.Name)
	}

	sort.Strings(relayNames)

	for _, name := range relayNames {
		fmt.Println(name)
	}

}
