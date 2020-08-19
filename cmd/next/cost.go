package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func saveCostMatrix(env Environment, filename string) {
	var uri string
	var err error

	if uri, err = env.RelayBackendURL(); err != nil {
		handleRunTimeError(fmt.Sprintf("Cannot get get relay backend hostname: %v\n", err), 1)
	}

	uri += "/cost_matrix"

	r, err := http.Get(uri)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not get the route matrix from the backend: %v\n", err), 1)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		handleRunTimeError(fmt.Sprintf("relay backend return an error response code: %d\n", r.StatusCode), 1)
	}

	file, err := os.Create(filename)
	if err != nil {
		handleRunTimeError(fmt.Sprintf("could not open file for writing: %v\n", err), 1)
	}
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		handleRunTimeError(fmt.Sprintf("error writing cost matrix to file: %v\n", err), 1)
	}
}
