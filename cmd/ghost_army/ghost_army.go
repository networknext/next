package main

import (
	"context"
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

	slices := make([]transport.SessionPortalData, count)
	for i := uint64(0); i < count; i++ {
		var entry ghostarmy.Entry
		if !entry.ReadFrom(bin, &index) {
			fmt.Printf("can't read entry at index %d\n", i)
		}

		entry.Into(&slices[i])
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
		index := 0
		var startTime int64 = 0

		for {
			begin := time.Now()
			endIndex := (startTime + 10) / 10

			if endIndex > SlicesInDay {
				startTime = 10
			}

			for slices[index].Slice.Timestamp.Unix() < startTime {
				fmt.Printf("skipping entry at index %d\n", index)
				index++
				index = index % len(slices)
			}

			for slices[index].Slice.Timestamp.Unix() < startTime+10 {
				fmt.Printf("publishing entry at index %d\n", index)
				publishChan <- slices[index]
				index++
				index = index % len(slices)
			}

			startTime += 10

			end := time.Now()

			diff := end.Sub(begin)

			time.Sleep((time.Second * 10) - diff)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
