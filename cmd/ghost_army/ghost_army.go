package main

import (
	"fmt"
	"io/ioutil"

	"github.com/networknext/backend/encoding"
)

const (
	SlicesInday = 60 * 60 * 24 / 10
)

func main() {
	// read binary file
	bin, err := ioutil.ReadFile("ghost_army.dat")
	if err != nil {
		fmt.Printf("could not read 'ghost_army.dat': %v\n", err)
	}

	// unmarshal
	index := 0

	var count uint64
	if !encoding.ReadUint64(bin, &index, &count) {
		fmt.Println("could not read count")
	}

	// publish to zero mq, sleep for 10 seconds, repeat
}
