package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func getRelaysBin(env Environment, filename string) {
	var err error

	uri := fmt.Sprintf("%s/relays.bin", env.PortalHostname())

	r, err := http.Get(uri)
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
}
