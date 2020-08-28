package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
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

	// publish to zero mq, sleep for 10 seconds, repeat

	publishChan := make(chan ghostarmy.Entry)

	ctx := context.Background()

	go func() {
		for {
			select {
			case entry := <-publishChan:
				fmt.Printf("%s\n", time.Unix(entry.Timestamp, 0).String())
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		index := 0
		var startTime int64 = 0

		for {
			endIndex := (startTime + 10) / 10

			if endIndex > SlicesInDay {
				startTime = 10
			}

			for entries[index].Timestamp < startTime {
				fmt.Printf("skipping entry at index %d\n", index)
				index++
				index = index % len(entries)
			}

			for entries[index].Timestamp < startTime+10 {
				fmt.Printf("publishing entry at index %d\n", index)
				publishChan <- entries[index]
				index++
				index = index % len(entries)
			}

			startTime += 10

			time.Sleep(time.Second * 10)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
