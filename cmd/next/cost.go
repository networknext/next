package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func saveCostMatrix(env Environment, filename string) {
	var uri string
	var err error

	if uri, err = env.RelayBackendURL(); err != nil {
		log.Fatalf("Cannot get get relay backend hostname: %v\n", err)
	}

	uri += "/cost_matrix"

	r, err := http.Get(uri)
	if err != nil {
		log.Fatalln(fmt.Errorf("could not get the route matrix from the backend: %w", err))
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		log.Fatalf("relay backend returns non 200 response code: %d\n", r.StatusCode)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalln(fmt.Errorf("could not open file for writing: %w", err))
	}
	defer file.Close()

	if _, err := io.Copy(file, r.Body); err != nil {
		log.Fatalln(fmt.Errorf("error writing cost matrix to file: %w", err))
	}
}
