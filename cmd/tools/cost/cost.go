/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
)

func main() {

	fmt.Printf("\nWelcome to Network Next!\n\n")

	resp, err := http.Get("http://localhost:30000/cost_matrix")
	if err != nil {
		fmt.Printf("error: could not get cost matrix: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	ioutil.WriteFile("./dist/cost.bin", body, 0644)

	fmt.Printf("Wrote cost matrix to 'dist/cost.bin'\n\n")
}
