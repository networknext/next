/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	url := flag.String("url", "http://localhost:30000/cost_matrix", "http://localhost:30000/cost_matrix")
	flag.Parse()

	resp, err := http.Get(*url)
	if err != nil {
		log.Fatalf("error: could not get cost matrix: %v\n", err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		log.Fatalln(err)
	}
}
