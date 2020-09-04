package main

import (
	"bufio"
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

/*
 * August 18th: Needs to be pulled again
 * August 19th: Needs to be pulled again
 * August 20th: Needs to be pulled again
 * August 27th: https://drive.google.com/file/d/1y4zOJXXx_9KQD8pz_l5b46lNKrK-pkGN/view
 * August 28th: https://drive.google.com/file/d/1C8GvCLUGA5v0IIMY4ywqleO61iZN_H1p/view
 * August 29th: https://drive.google.com/file/d/1kutKjR7KmmBmJDmJ4BNj0GuC9I-7zukm/view?usp=sharing
 * August 30th: https://drive.google.com/file/d/1OYshu0pfoVDEsH_4pBBbwmfEYnpUjHC0/view?usp=sharing
 * August 31th: https://drive.google.com/file/d/1aXXfkA5s4nF7eWXHr8G6VdMvShmrk6re/view?usp=sharing
 * September 1st: https://drive.google.com/file/d/1dXAMg-aKwMEZt22GlzYKRKc49yvE1KhK/view?usp=sharing
 * September 2nd: https://drive.google.com/file/d/1HgJMZd88gjXmeSx79ueah1oXxGXiOYjp/view?usp=sharing
 */

type sortableEntries []ghostarmy.Entry

func (self sortableEntries) Len() int {
	return len(self)
}

func (self sortableEntries) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

func (self sortableEntries) Less(i, j int) bool {
	return self[i].Timestamp < self[j].Timestamp
}

// Input files can be in any order

func main() {
	if len(os.Args) < 3 {
		fmt.Println("you must supply at least 2 arguments: <input file name(s)> <output file name>")
		os.Exit(1)
	}

	var datacenterCSV string
	if v, ok := os.LookupEnv("DATACENTERS_CSV"); ok {
		datacenterCSV = v
	} else {
		fmt.Println("you must set DATACENTERS_CSV to a file")
		os.Exit(1)
	}

	// parse datacenter csv
	inputfile, err := os.Open(datacenterCSV)
	if err != nil {
		fmt.Printf("could not open '%s': %v\n", datacenterCSV, err)
		os.Exit(1)
	}

	lines, err := csv.NewReader(inputfile).ReadAll()
	if err != nil {
		fmt.Printf("could not read csv data: %v\n", err)
		os.Exit(1)
	}
	inputfile.Close()

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

	var ispsFilename string
	if v, ok := os.LookupEnv("ISPS_FILE"); ok {
		ispsFilename = v
	} else {
		fmt.Println("you must set ISPS_FILE to a file")
		os.Exit(1)
	}

	file, err := os.Open(ispsFilename)
	if err != nil {
		fmt.Printf("could not open %s: %v", ispsFilename, err)
		os.Exit(1)
	}

	var isps []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		isps = append(isps, scanner.Text())
	}
	file.Close()

	infiles := make([]string, len(os.Args)-2)
	for i := 1; i < len(os.Args)-1; i++ {
		infiles[i-1] = os.Args[i]
	}

	outfile := os.Args[len(os.Args)-1]

	// convert to binary

	var entries sortableEntries
	entries = make([]ghostarmy.Entry, 0)

	for _, infile := range infiles {
		// read in exported data
		inputfile, err := os.Open(infile)
		if err != nil {
			fmt.Printf("could not open '%s': %v\n", infile, err)
			os.Exit(1)
		}

		lines, err := csv.NewReader(inputfile).ReadAll()
		if err != nil {
			fmt.Printf("could not read csv data: %v\n", err)
			os.Exit(1)
		}

		inputfile.Close()

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

			t, err := time.Parse("2006-1-2 15:04:05", line[i])
			checkErr(err, lineNum)
			i++

			year, month, day := t.Date()
			t2 := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
			secsIntoDay := int64(t.Sub(t2).Seconds())

			entry.Timestamp = secsIntoDay

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

			if line[i] == "" {
				// get latitude from datacenter
				entry.Latitude = dcmap[uint64(entry.DatacenterID)].Lat
			} else {
				// use actual latitude
				lat, err := strconv.ParseFloat(line[i], 64)
				checkErr(err, lineNum)
				entry.Latitude = lat
			}
			i++

			if line[i] == "" {
				// get longitude from datacenter
				entry.Longitude = dcmap[uint64(entry.DatacenterID)].Long
			} else {
				// use actual longitude
				long, err := strconv.ParseFloat(line[i], 64)
				checkErr(err, lineNum)
				entry.Longitude = long
			}
			i++

			if line[i] == "" {
				// get random isp
				entry.ISP = isps[uint64(entry.Userhash)%uint64(len(isps))]
			} else {
				// use actual isp
				entry.ISP = line[i]
			}

			// push back entry
			entries = append(entries, entry)
		}
	}

	// sort on timestamp
	sort.Sort(entries)

	// encode to binary format
	index := 0

	bin := make([]byte, 8)
	encoding.WriteUint64(bin, &index, uint64(len(entries)))
	for _, entry := range entries {
		dat, err := entry.MarshalBinary()
		if err != nil {
			fmt.Printf("unable to marshal binary for session %d: %v\n", entry.SessionID, err)
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
