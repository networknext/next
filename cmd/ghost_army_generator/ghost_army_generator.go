package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/networknext/backend/cmd/ghost_army"
	"github.com/networknext/backend/encoding"
)

type sortable []ghost_army.Entry

func (self sortable) Len() int {
	return len(self)
}

func (self sortable) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self sortable) Less(i, j int) bool {
	return self[i].Timestamp.Unix() < self[j].Timestamp.Unix()
}

func main() {
	// read in exported data
	inputfile, err := os.Open("ghost_army.csv")
	if err != nil {
		fmt.Printf("could not open ghost_army.csv: %v\n", err)
		os.Exit(1)
	}
	defer inputfile.Close()

	lines, err := csv.NewReader(inputfile).ReadAll()
	if err != nil {
		fmt.Printf("could not read csv data: %v\n", err)
		os.Exit(1)
	}

	// convert to binary

	var list sortable
	list = make([]ghost_army.Entry, 0)
	for _, line := range lines {
		var err error
		var entry ghost_army.Entry

		i := 0

		// line into Entry
		entry.SessionID, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err)
		i++

		entry.Timestamp, err = time.Parse("2006-1-2 15:04:05", line[i])
		checkErr(err)
		i++

		entry.BuyerId, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err)
		i++

		entry.SliceNumber, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err)
		i++

		entry.Next, err = strconv.ParseBool(line[i])
		checkErr(err)
		i++

		entry.DirectRTT, err = strconv.ParseFloat(line[i], 64)
		checkErr(err)
		i++

		entry.DirectJitter, err = strconv.ParseFloat(line[i], 64)
		checkErr(err)
		i++

		entry.DirectPacketLoss, err = strconv.ParseFloat(line[i], 64)
		checkErr(err)
		i++

		entry.NextRTT, err = strconv.ParseFloat(line[i], 64)
		checkErr(err)
		i++

		entry.NextJitter, err = strconv.ParseFloat(line[i], 64)
		checkErr(err)
		i++

		entry.NextPacketLoss, err = strconv.ParseFloat(line[i], 64)
		checkErr(err)
		i++

		relays, err := csv.NewReader(strings.NewReader(line[i])).ReadAll()
		checkErr(err)
		i++

		for _, relay := range relays {
			err = json.Unmarshal([]byte(relay[0]), &entry.NextRelays)
			checkErr(err)
		}

		entry.TotalPrice, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err)
		i++

		entry.ClientToServerPacketsLost, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err)
		i++

		entry.ServerToClientPacketsLost, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err)
		i++

		entry.Committed, err = strconv.ParseBool(line[i])
		checkErr(err)
		i++

		entry.Flagged, err = strconv.ParseBool(line[i])
		checkErr(err)
		i++

		entry.Multipath, err = strconv.ParseBool(line[i])
		checkErr(err)
		i++

		entry.NextBytesUp, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err)
		i++

		entry.NextBytesDown, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err)
		i++

		entry.Initial, err = strconv.ParseBool(line[i])
		checkErr(err)
		i++

		entry.DatacenterID, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err)
		i++

		entry.RttReduction, err = strconv.ParseBool(line[i])
		checkErr(err)
		i++

		entry.PacketLossReduction, err = strconv.ParseBool(line[i])
		checkErr(err)
		i++

		prices, err := csv.NewReader(strings.NewReader(line[i])).ReadAll()
		checkErr(err)
		i++

		for _, price := range prices {
			err = json.Unmarshal([]byte(price[0]), &entry.NextRelaysPrice)
			checkErr(err)
		}

		entry.Userhash, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err)
		i++

		// push back entry
		list = append(list, entry)
	}

	bin := make([]byte, 8)

	// sort on timestamp
	sort.Sort(list)

	// encode to binary format
	index := 0
	encoding.WriteUint64(bin, &index, uint64(len(list)))
	for _, item := range list {
		dat, err := item.MarshalBinary()
		if err != nil {
			fmt.Printf("unable to marshal binary for session %d: %v\n", item.SessionID, err)
			continue
		}

		bin = append(bin, dat...)
	}

	// export

	err = ioutil.WriteFile("ghost_army.dat", bin, 644)
	if err != nil {
		fmt.Printf("could not create output file: %v\n", err)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
