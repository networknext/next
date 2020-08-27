package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/networknext/backend/encoding"
	ghostarmy "github.com/networknext/backend/ghost_army"
)

const (
	SlicesInDay = 60 * 60 * 24 / 10
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
	}

	for i := 0; i < 50; i++ {
		fmt.Printf("\n%d = %v\n", i, entries[i])
	}

	sliceMap := make(map[int]*ghostarmy.Entry)

	for i := range entries {
		entry := &entries[i]
		t := time.Unix(entry.Timestamp, 0)
		year, month, day := t.Date()
		t2 := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
		secsIntoDay := int(t.Sub(t2).Seconds())
		sliceIndex := (secsIntoDay / 10) % SlicesInDay
		sliceMap[sliceIndex] = entry
	}

	// publish to zero mq, sleep for 10 seconds, repeat
}
