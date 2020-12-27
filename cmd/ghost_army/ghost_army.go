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

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/pubsub"

	ghostarmy "github.com/networknext/backend/modules/ghost_army"
)

const (
	SecondsInDay = 86400
)

func main() {
	// this var is just used to catch the situation where ghost army publishes
	// way too many slices in a given interval. Shouldn't happen anymore but just in case

	var estimatedPeakSessionCount int
	if v, ok := os.LookupEnv("GHOST_ARMY_PEAK_SESSION_COUNT"); ok {
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			estimatedPeakSessionCount = int(num)
		} else {
			fmt.Printf("could not parse GHOST_ARMY_PEAK_SESSION_COUNT: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("GHOST_ARMY_PEAK_SESSION_COUNT not set")
		os.Exit(1)
	}

	var infile string
	if v, ok := os.LookupEnv("GHOST_ARMY_BIN"); ok {
		infile = v
	} else {
		fmt.Println("GHOST_ARMY_BIN not set")
		os.Exit(1)
	}

	var datacenterCSV string
	if v, ok := os.LookupEnv("DATACENTERS_CSV"); ok {
		datacenterCSV = v
	} else {
		fmt.Println("DATACENTERS_CSV not set")
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

	var numSegments uint64
	if !encoding.ReadUint64(bin, &index, &numSegments) {
		fmt.Println("could not read num segments")
	}

	slices := make([]transport.SessionPortalData, 0)
	entityIndex := 0
	for s := uint64(0); s < numSegments; s++ {
		var count uint64
		if !encoding.ReadUint64(bin, &index, &count) {
			fmt.Println("could not read count")
		}

		newSlice := make([]transport.SessionPortalData, count)
		slices = append(slices, newSlice...)
		fmt.Printf("reading in %d entries\n", count)

		for i := uint64(0); i < count; i++ {
			var entry ghostarmy.Entry
			if !entry.ReadFrom(bin, &index) {
				fmt.Printf("can't read entry at index %d\n", i)
			}

			entry.Into(&slices[entityIndex], dcmap, buyerID)
			entityIndex++
		}
	}
	bin = nil
	// publish to zero mq, sleep for 10 seconds, repeat

	publishChan := make(chan transport.SessionPortalData)

	ctx := context.Background()

	portalPublishers := make([]pubsub.Publisher, 0)
	{
		fmt.Printf("setting up portal cruncher\n")

		portalCruncherHosts := envvar.GetList("PORTAL_CRUNCHER_HOSTS", []string{"tcp://127.0.0.1:5555"})

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

		for _, host := range portalCruncherHosts {
			portalCruncherPublisher, err := pubsub.NewPortalCruncherPublisher(host, int(postSessionPortalSendBufferSize))
			if err != nil {
				fmt.Printf("could not create portal cruncher publisher: %v\n", err)
				os.Exit(1)
			}

			portalPublishers = append(portalPublishers, portalCruncherPublisher)
		}
	}

	go func() {
		publisherIndex := 0

		for {
			select {
			case slice := <-publishChan:
				sessionBytes, err := slice.MarshalBinary()
				if err != nil {
					fmt.Printf("could not marshal binary for slice session id %d", slice.Meta.ID)
					continue
				}

				portalPublishers[publisherIndex].Publish(ctx, pubsub.TopicPortalCruncherSessionData, sessionBytes)
				publisherIndex = (publisherIndex + 1) % len(portalPublishers)
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
			if sliceBegin >= SecondsInDay {
				i = 0
				sliceBegin = currentSecs() // account for any inaccuracy by calling sleep()
				dateOffset = getLastMidnight()
			}

			var interval int64 = 10

			// only useful at 11:50:5x pm - midnight
			// forces the last batch to be sent within the above interval
			if sliceBegin+interval > SecondsInDay {
				interval = SecondsInDay - sliceBegin
			}

			// seek to the next position slices should be from
			// mainly useful when starting the program
			for i < len(slices) && slices[i].Slice.Timestamp.Unix() < sliceBegin {
				i++
			}

			before := i

			// only read for the interval, usually 10 seconds
			for i < len(slices) && slices[i].Slice.Timestamp.Unix() < sliceBegin+interval {
				slice := slices[i]

				// slice timestamp will be in the range of 0 - SecondsInDay * 3,
				// so adjust the timestamp by the time the loop was started
				slice.Slice.Timestamp = dateOffset.Add(time.Second * time.Duration(slice.Slice.Timestamp.Unix()))

				publishChan <- slice
				i++
			}

			diff := i - before

			if diff > estimatedPeakSessionCount {
				fmt.Printf("sent more than %d slices this interval, num sent = %d, current interval in secs = %d - %d\n", estimatedPeakSessionCount, diff, sliceBegin, sliceBegin+interval)
			}

			// increment by the interval
			sliceBegin += interval

			time.Sleep((time.Second * time.Duration(interval)) - time.Since(begin))
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
