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
	"sync"
	"sync/atomic"
	"time"

	"github.com/networknext/backend/modules/encoding"
	ghostarmy "github.com/networknext/backend/modules/ghost_army"
)

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
	if len(os.Args) < 6 {
		fmt.Println("you must supply at least 5 arguments: <processing threads> <datacenter csv> <ISPs file> <output file> <input file name(s)>")
		os.Exit(1)
	}

	var processingThreads uint64
	if v, err := strconv.ParseUint(os.Args[1], 10, 64); err == nil {
		processingThreads = v
	} else {
		fmt.Printf("could not parse processing thread count '%s': %v\n", os.Args[1], err)
	}
	fmt.Printf("%d\n", processingThreads)
	dcmap := parseDatacenterCSV(os.Args[2])
	isps := parseISPCSV(os.Args[3])
	outfile := os.Args[4]
	infiles := os.Args[5:]

	// convert to binary

	var entryMutex sync.Mutex
	var entries = make(sortableEntries, 0)

	var wg sync.WaitGroup
	wg.Add(len(infiles))

	processedFiles := uint64(0)
	for i := uint64(0); i < processingThreads; i++ {
		go func(processedFiles *uint64) {
			index := atomic.AddUint64(processedFiles, 1) - 1
			for index < uint64(len(infiles)) {
				localentries := make(sortableEntries, 0)
				infile := infiles[index]

				fmt.Printf("reading %s\n", infile)
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
					var entry ghostarmy.Entry
					parseLine(line, lineNum, &entry, dcmap, isps)
					localentries = append(localentries, entry)
				}

				entryMutex.Lock()
				entries = append(entries, localentries...)
				entryMutex.Unlock()
				index = atomic.AddUint64(processedFiles, 1) - 1
				wg.Done()
			}
		}(&processedFiles)
	}

	wg.Wait()

	// sort on timestamp
	fmt.Printf("sorting %d entries...\n", len(entries))
	sort.Sort(entries)

	// encode to binary format
	fmt.Println("encoding to buffer...")
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

	entries = nil // free the data buffer now that it's unused

	// export

	fmt.Println("writing to file...")
	err := ioutil.WriteFile(outfile, bin, 0644)
	if err != nil {
		fmt.Printf("could not create output file '%s': %v\n", outfile, err)
	}

	fmt.Println("done!")
}

func checkErr(err error, lineNum int) {
	if err != nil {
		fmt.Printf("paniced on line %d of csv\n", lineNum)
		panic(err)
	}
}

func parseDatacenterCSV(filename string) map[uint64]ghostarmy.StrippedDatacenter {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("could not open '%s': %v\n", filename, err)
		os.Exit(1)
	}

	lines, err := csv.NewReader(file).ReadAll()
	if err != nil {
		fmt.Printf("could not read csv data: %v\n", err)
		os.Exit(1)
	}
	file.Close()

	dcmap := make(map[uint64]ghostarmy.StrippedDatacenter)

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

	return dcmap
}

func parseISPCSV(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("could not open %s: %v", filename, err)
		os.Exit(1)
	}

	var isps []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		isps = append(isps, scanner.Text())
	}
	file.Close()
	return isps
}

func parseLine(line []string, lineNum int, entry *ghostarmy.Entry, dcmap ghostarmy.DatacenterMap, isps []string) {
	var err error

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

	if line[i] != "" {
		entry.Userhash, err = strconv.ParseInt(line[i], 10, 64)
		checkErr(err, lineNum)
	} else {
		entry.Userhash = entry.SessionID ^ entry.DatacenterID
	}
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
}
