package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/networknext/backend/encoding"
	ghostarmy "github.com/networknext/backend/ghost_army"
)

type sortable []ghostarmy.Entry

func (self sortable) Len() int {
	return len(self)
}

func (self sortable) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self sortable) Less(i, j int) bool {
	return self[i].Timestamp < self[j].Timestamp
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("you must supply arguments (input file name, output file name)")
		os.Exit(1)
	}

	infile := os.Args[1]
	outfile := os.Args[2]

	// read in exported data
	inputfile, err := os.Open(infile)
	if err != nil {
		fmt.Printf("could not open '%s': %v\n", infile, err)
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
	list = make([]ghostarmy.Entry, 0)
	for lineNum, line := range lines {
		if lineNum == 0 {
			continue
		}
		var err error
		var entry ghostarmy.Entry

		i := 0

		// line into Entry
		entry.SessionID, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err, lineNum)
		i++

		ts, err := time.Parse("2006-1-2 15:04:05", line[i])
		checkErr(err, lineNum)
		i++

		entry.Timestamp = ts.Unix()

		entry.BuyerID, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err, lineNum)
		i++

		entry.SliceNumber, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err, lineNum)
		i++

		entry.Next, err = strconv.ParseBool(line[i])
		checkErr(err, lineNum)
		i++

		entry.DirectRTT, err = strconv.ParseFloat(line[i], 64)
		checkErr(err, lineNum)
		i++

		entry.DirectJitter, err = strconv.ParseFloat(line[i], 64)
		checkErr(err, lineNum)
		i++

		entry.DirectPacketLoss, err = strconv.ParseFloat(line[i], 64)
		checkErr(err, lineNum)
		i++

		entry.NextRTT, err = strconv.ParseFloat(line[i], 64)
		checkErr(err, lineNum)
		i++

		entry.NextJitter, err = strconv.ParseFloat(line[i], 64)
		checkErr(err, lineNum)
		i++

		entry.NextPacketLoss, err = strconv.ParseFloat(line[i], 64)
		checkErr(err, lineNum)
		i++

		err = json.Unmarshal([]byte(line[i]), &entry.NextRelays)
		checkErr(err, lineNum)
		i++

		entry.TotalPrice, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err, lineNum)
		i++

		if line[i] != "" {
			ctspl, err := strconv.ParseFloat(line[i], 64)
			checkErr(err, lineNum)
			entry.ClientToServerPacketsLost = int64(ctspl)
		}
		i++

		if line[i] != "" {
			stcpl, err := strconv.ParseFloat(line[i], 64)
			checkErr(err, lineNum)
			entry.ServerToClientPacketsLost = int64(stcpl)
		}
		i++

		entry.Committed, err = strconv.ParseBool(line[i])
		checkErr(err, lineNum)
		i++

		entry.Flagged, err = strconv.ParseBool(line[i])
		checkErr(err, lineNum)
		i++

		entry.Multipath, err = strconv.ParseBool(line[i])
		checkErr(err, lineNum)
		i++

		entry.NextBytesUp, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err, lineNum)
		i++

		entry.NextBytesDown, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err, lineNum)
		i++

		entry.Initial, err = strconv.ParseBool(line[i])
		checkErr(err, lineNum)
		i++

		entry.DatacenterID, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err, lineNum)
		i++

		entry.RttReduction, err = strconv.ParseBool(line[i])
		checkErr(err, lineNum)
		i++

		entry.PacketLossReduction, err = strconv.ParseBool(line[i])
		checkErr(err, lineNum)
		i++

		err = json.Unmarshal([]byte(line[i]), &entry.NextRelaysPrice)
		checkErr(err, lineNum)
		i++

		entry.Userhash, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err, lineNum)
		i++

		// push back entry
		list = append(list, entry)
	}

	// sort on timestamp
	sort.Sort(list)

	// encode to binary format
	index := 0

	bin := make([]byte, 8)
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

	err = ioutil.WriteFile(outfile, bin, 0644)
	if err != nil {
		fmt.Printf("could not create output file '%s': %v\n", outfile, err)
	}
}

func checkErr(err error, lineNum int) {
	if err != nil {
		fmt.Printf("paniced on line %d of csv\n", lineNum)
		panic(err)
	}
}
