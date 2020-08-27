package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/networknext/backend/encoding"
	ghostarmy "github.com/networknext/backend/ghost_army"
)

const (
	SlicesInday = 60 * 60 * 24 / 10
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("you must supply an argument (input file name)")
		os.Exit(1)
	}

	infile := os.Args[1]

	// read binary file
	bin, err := ioutil.ReadFile(infile)
	if err != nil {
		fmt.Printf("could not read '%s': %v\n", infile, err)
	}

	// unmarshal
	index := 0

	var count uint64
	if !encoding.ReadUint64(bin, &index, &count) {
		fmt.Println("could not read count")
	}

	fmt.Printf("reading in %d entries\n", count)

	entries := make([]ghostarmy.Entry, count)
	for i := uint64(0); i < count; i++ {
		entry := &entries[i]
		if !entry.ReadFrom(bin, &index) {
			fmt.Printf("can't read entry at index %d\n", i)
		}

		break
	}

	fmt.Printf("first entry = %v", entries[0])

	// publish to zero mq, sleep for 10 seconds, repeat
}
