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
	SlicesInDay = 60 * 60 * 24 / 10
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

		// slice begin is the current number of seconds into the day
		t := time.Now()
		year, month, day := t.Date()
		t2 := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
		sliceBegin := int64(t.Sub(t2).Seconds())

		index := 0
		dateOffset := getLastMidnight()
		for {
			begin := time.Now()
			endIndex := (sliceBegin + 10) / 10

			if endIndex > SlicesInDay*3 {
				index = 0
				sliceBegin = 0
				dateOffset = getLastMidnight()
			}

			for slices[index].Slice.Timestamp.Unix() < sliceBegin {
				index++
				index = index % len(slices)
			}

			for slices[index].Slice.Timestamp.Unix() < sliceBegin+10 {
				slice := slices[index]

				// slice timestamp will be in the range of 0 - SecondsInDay * 3,
				// so adjust the timestamp by the time the loop was started
				slice.Slice.Timestamp = dateOffset.Add(time.Second*-10 + time.Second*time.Duration(slice.Slice.Timestamp.Unix()))

				publishChan <- slice
				index++
				index = index % len(slices)
			}

			sliceBegin += 10

			end := time.Now()

			diff := end.Sub(begin)

			time.Sleep((time.Second * 10) - diff)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
