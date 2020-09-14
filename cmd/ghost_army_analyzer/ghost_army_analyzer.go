package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/networknext/backend/encoding"
	ghostarmy "github.com/networknext/backend/ghost_army"
)

const (
	SlicesInDay = 60 * 60 * 24 / 10
)

func main() {
	infile := os.Args[1]

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
		if !entries[i].ReadFrom(bin, &index) {
			fmt.Printf("can't read entry at index %d\n", i)
		}
	}

	bin = nil

	// publish to zero mq, sleep for 10 seconds, repeat
	var counts [SlicesInDay]struct {
		sliceNum int64
		count    uint64
	}

	var i int64 = 0
	for sliceNum := int64(0); sliceNum < SlicesInDay && i < int64(len(entries)); sliceNum++ {
		counts[sliceNum].sliceNum = sliceNum
		counts[sliceNum].count = 0

		// seek to the next position slices should be from
		for entries[i].Timestamp < sliceNum*10 {
			i++
		}

		// only read for the next 10 seconds
		for entries[i].Timestamp < sliceNum*10+10 {
			counts[sliceNum].count++
			i++

			if i >= int64(len(entries)) {
				fmt.Printf("reached end at index %d\n", i)
				break
			}
		}
	}

	sort.SliceStable(counts[:], func(i, j int) bool {
		return counts[i].count < counts[j].count
	})

	for i := range counts {
		c := counts[i]

		timeBegin := time.Unix(c.sliceNum*10, 0)
		timeEnd := time.Unix(c.sliceNum*10+10, 0)

		bhr, bmin, bsec := timeBegin.Clock()
		ehr, emin, esec := timeEnd.Clock()

		fmt.Printf("%d:%d:%d - %d:%d:%d = %d\n", bhr, bmin, bsec, ehr, emin, esec, c.count)
	}
}
