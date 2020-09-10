package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/networknext/backend/encoding"
	ghostarmy "github.com/networknext/backend/ghost_army"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
)

const (
	SecondsInDay = 86400
)

func main() {
	var infile string
	if v, ok := os.LookupEnv("GHOST_ARMY_BIN"); ok {
		infile = v
	} else {
		fmt.Println("you must set GHOST_ARMY_BIN to a file")
		os.Exit(1)
	}

	var datacenterCSV string
	if v, ok := os.LookupEnv("DATACENTERS_CSV"); ok {
		datacenterCSV = v
	} else {
		fmt.Println("you must set DATACENTERS_CSV to a file")
		os.Exit(1)
	}

	buyerID := ghostarmy.GhostArmyBuyerID(os.Getenv("ENV"))

	// parse datacenter csv
	inputfile, err := os.Open(datacenterCSV)
	if err != nil {
		fmt.Printf("could not open '%s': %v\n", datacenterCSV, err)
		os.Exit(1)
	}
	defer inputfile.Close()

	lines, err := csv.NewReader(inputfile).ReadAll()
	if err != nil {
		fmt.Printf("could not read csv data: %v\n", err)
		os.Exit(1)
	}

	var dcmap ghostarmy.DatacenterMap
	dcmap = make(map[uint64]ghostarmy.StrippedDatacenter)

	for lineNum, line := range lines {
		if lineNum == 0 {
			continue
		}

		var datacenter ghostarmy.StrippedDatacenter
		datacenter.Name = line[0]
		id, err := strconv.ParseUint(line[1], 10, 64)
		if err != nil {
			fmt.Printf("could not parse id for dc %s", datacenter.Name)
			continue
		}
		datacenter.Lat, err = strconv.ParseFloat(line[2], 64)
		if err != nil {
			fmt.Printf("could not parse lat for dc %s", datacenter.Name)
			continue
		}
		datacenter.Long, err = strconv.ParseFloat(line[3], 64)
		if err != nil {
			fmt.Printf("could not parse long for dc %s", datacenter.Name)
			continue
		}

		dcmap[id] = datacenter
	}

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

	slices := make([]transport.SessionPortalData, count)
	for i := uint64(0); i < count; i++ {
		var entry ghostarmy.Entry
		if !entry.ReadFrom(bin, &index) {
			fmt.Printf("can't read entry at index %d\n", i)
		}

		entry.Into(&slices[i], dcmap, buyerID)
	}

	bin = nil

	// publish to zero mq, sleep for 10 seconds, repeat

	publishChan := make(chan transport.SessionPortalData)

	ctx := context.Background()

	var portalPublisher pubsub.Publisher
	{
		fmt.Printf("setting up portal cruncher\n")

		portalCruncherHost, ok := os.LookupEnv("PORTAL_CRUNCHER_HOST")
		if !ok {
			fmt.Println("env var PORTAL_CRUNCHER_HOST must be set")
			os.Exit(1)
		}

		postSessionPortalSendBufferSizeString, ok := os.LookupEnv("POST_SESSION_PORTAL_SEND_BUFFER_SIZE")
		if !ok {
			fmt.Println("env var POST_SESSION_PORTAL_SEND_BUFFER_SIZE must be set")
			os.Exit(1)
		}

		postSessionPortalSendBufferSize, err := strconv.ParseInt(postSessionPortalSendBufferSizeString, 10, 64)
		if err != nil {
			fmt.Printf("could not parse envvar POST_SESSION_PORTAL_SEND_BUFFER_SIZE: %v\n", err)
			os.Exit(1)
		}

		portalCruncherPublisher, err := pubsub.NewPortalCruncherPublisher(portalCruncherHost, int(postSessionPortalSendBufferSize))
		if err != nil {
			fmt.Printf("could not create portal cruncher publisher: %v\n", err)
			os.Exit(1)
		}

		portalPublisher = portalCruncherPublisher
	}

	go func() {
		for {
			select {
			case slice := <-publishChan:
				sessionBytes, err := slice.MarshalBinary()
				if err != nil {
					fmt.Printf("could not marshal binary for slice session id %d", slice.Meta.ID)
					continue
				}
				portalPublisher.Publish(pubsub.TopicPortalCruncherSessionData, sessionBytes)
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		getLastMidnight := func() time.Time {
			t := time.Now()
			year, month, day := t.Date()
			return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
		}

		currentSecs := func() int64 {
			t := time.Now()
			year, month, day := t.Date()
			t2 := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
			return int64(t.Sub(t2).Seconds())
		}

		// slice begin is the current number of seconds into the day
		sliceBegin := currentSecs()

		// date offset is used to adjust the slice timestamp to the current day
		dateOffset := getLastMidnight()

		i := 0
		for {
			begin := time.Now()

			// reset if at end of day
			if sliceBegin == SecondsInDay {
				i = 0
				sliceBegin = currentSecs() // account for any inaccuracy by calling sleep()
				dateOffset = getLastMidnight()
			}

			var interval int64 = 10

			// only useful at 11:50:5x pm - midnight
			// forces the last slice to be sent within the above interval
			if sliceBegin+10 > SecondsInDay {
				interval = SecondsInDay - sliceBegin
			}

			// seek to the next position slices should be from
			// mainly useful when starting the program
			for slices[i].Slice.Timestamp.Unix() < sliceBegin && i < len(slices) {
				i++
			}

			before := i

			// only read for the next 10 seconds
			for slices[i].Slice.Timestamp.Unix() < sliceBegin+interval && i < len(slices) {
				slice := slices[i]

				// slice timestamp will be in the range of 0 - SecondsInDay * 3,
				// so adjust the timestamp by the time the loop was started
				slice.Slice.Timestamp = dateOffset.Add(time.Second * time.Duration(slice.Slice.Timestamp.Unix()))

				publishChan <- slice
				i++
			}

			// check if somehow too many slices were sent, using the analyzer the
			// interval with the most slices was 10,265, so give it a little buffer
			// since the program may start anywhere within an interval
			slicesSent := i - before
			if slicesSent > 12000 {
				fmt.Printf("sent too many slices at %s = %d\n", time.Now().String(), slicesSent)
			}

			// increment by the interval, usually 10 seconds
			// at 11:50:5x this will make sliceBegin == 86400
			sliceBegin += interval

			time.Sleep((time.Second * time.Duration(interval)) - time.Since(begin))
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
