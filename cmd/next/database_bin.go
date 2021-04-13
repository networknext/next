package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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

	f, err := os.Stat("./relays.bin")
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not find database.bin? %v\n", err), 1)
	}

	fileSize := f.Size()
	fmt.Printf("Successfully retrieved ./database.bin (%d bytes)\n", fileSize)

}
