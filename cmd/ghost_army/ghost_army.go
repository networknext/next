package main

import (
	"fmt"
	"io/ioutil"

	"github.com/networknext/backend/transport"
)

const (
	SlicesInday = 60 * 60 * 24 / 10
)

func main() {
	// read binary file
	data, err := ioutil.ReadFile("ghost_army.dat")
	if err != nil {
		fmt.Printf("could not read 'ghost_army.dat': %v\n", err)
	}

	// sort & store in some data structure

	// publish to zero mq, sleep for 10 seconds, repeat
	var session transport.SessionPortalData
}
