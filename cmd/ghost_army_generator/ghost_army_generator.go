package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/networknext/backend/modules/encoding"

	ghostarmy "github.com/networknext/backend/modules-old/ghost_army"
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

type muxSessionMap struct {
	mux      *sync.RWMutex
	sessions map[int64]bool
}

func NewMuxSessionMap() muxSessionMap {
	return muxSessionMap{
		mux:      new(sync.RWMutex),
		sessions: make(map[int64]bool),
	}
}

func (m *muxSessionMap) Add(id int64) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.sessions[id] = true
}

func (m *muxSessionMap) Exists(id int64) bool {
	m.mux.RLock()
	defer m.mux.RUnlock()
	_, exists := m.sessions[id]
	return exists
}

func (m *muxSessionMap) Length() int {
	m.mux.Lock()
	defer m.mux.Unlock()
	return len(m.sessions)
}

// Input files can be in any order

func main() {
	if len(os.Args) < 8 {
		fmt.Println("you must supply at least 5 arguments: <processing threads> <max rtt> <max packet loss> <datacenter csv> <ISPs file> <output file> <input file name(s)>")
		os.Exit(1)
	}

	var processingThreads uint64
	if v, err := strconv.ParseUint(os.Args[1], 10, 64); err == nil {
		processingThreads = v
	} else {
		fmt.Printf("could not parse processing thread count '%s': %v\n", os.Args[1], err)
	}

	var maxRTT float64
	if v, err := strconv.ParseFloat(os.Args[2], 64); err == nil {
		maxRTT = v
	} else {
		fmt.Printf("could not parse max RTT '%s': %v\n", os.Args[2], err)
	}

	var maxPL float64
	if v, err := strconv.ParseFloat(os.Args[3], 64); err == nil {
		maxPL = v
	} else {
		fmt.Printf("could not parse max PL '%s': %v\n", os.Args[3], err)
	}
	fmt.Printf("%d\n", processingThreads)
	dcmap := parseDatacenterCSV(os.Args[4])
	isps := parseISPCSV(os.Args[5])
	outfile := os.Args[6]
	infiles := os.Args[7:]

	//fileMap stores the files by int hour
	fileMap := make(map[int][]string)
	for i := 0; i < 24; i++ {
		fileMap[i] = make([]string, 0)
	}

	for i := 0; i < len(infiles); i++ {
		parts := strings.Split(infiles[i], "_")
		hrIndex, err := strconv.Atoi(parts[5])
		if err != nil {
			fmt.Printf("unable to parse filename %s \n", infiles[i])
			continue
		}
		fileMap[hrIndex] = append(fileMap[hrIndex], infiles[i])
	}

	//open output file
	output, err := os.OpenFile(outfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {

		log.Fatalf("could not create output file '%s': %v\n", outfile, err)
	}
	defer func() {
		if err := output.Close(); err != nil {
			log.Fatal("error", err)
		}
	}()

	//metrics
	var totalNumEntries uint64
	var remainingEntries uint64
	var removedPL uint64
	var removedRTT uint64

	// convert to binary
	var entryMutex sync.Mutex
	var entries = make(sortableEntries, 0)

	thisDeleteSessionMap := NewMuxSessionMap()
	lastDeleteSessionMap := NewMuxSessionMap()
	sessionList := NewMuxSessionMap()

	//set segments in first output to file
	bin := make([]byte, 8)
	binIndex := 0
	encoding.WriteUint64(bin, &binIndex, 24)

	//process by hr to save ram usage
	for hr := 0; hr < 24; hr++ {
		filesInHrArr := fileMap[hr]
		numFiles := uint64(len(filesInHrArr))

		segments := numFiles / processingThreads
		if numFiles%processingThreads != 0 {
			segments++
		}

		processedFiles := uint64(0)
		for s := uint64(0); s < segments; s++ {
			filesToProcess := processingThreads
			if processedFiles+processingThreads > numFiles {
				filesToProcess = numFiles % processingThreads
			}
			var wg sync.WaitGroup
			for i := uint64(0); i < filesToProcess; i++ {
				wg.Add(1)
				var index uint64
				index = atomic.AddUint64(&processedFiles, 1) - 1

				go func() {
					localEntries := make(sortableEntries, 0)
					infile := filesInHrArr[index]

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
						atomic.AddUint64(&totalNumEntries, 1)

						parseLine(line, lineNum, &entry, dcmap, isps)

						//skip known bad sessions
						if exists := thisDeleteSessionMap.Exists(entry.SessionID); exists {
							continue
						}
						if exists := lastDeleteSessionMap.Exists(entry.SessionID); exists {
							continue
						}

						//skip bad session entry and add a delete to the
						if entry.NextRTT >= maxRTT {
							atomic.AddUint64(&removedRTT, 1)
							thisDeleteSessionMap.Add(entry.SessionID)
							continue
						}

						if entry.NextPacketLoss >= maxPL {
							atomic.AddUint64(&removedPL, 1)
							thisDeleteSessionMap.Add(entry.SessionID)
							continue
						}

						if exists := sessionList.Exists(entry.SessionID); !exists {
							sessionList.Add(entry.SessionID)
						}

						localEntries = append(localEntries, entry)
					}

					entryMutex.Lock()
					entries = append(entries, localEntries...)
					entryMutex.Unlock()
					wg.Done()

				}()
			}
			wg.Wait()
		}

		//check entries for sessions that should be deleted
		for i := 0; i < len(entries); i++ {
			//only check thisDeleteSessionMap as lastDeleteSessionMap is already checked
			if exists := thisDeleteSessionMap.Exists(entries[i].SessionID); exists {
				entries[i] = entries[len(entries)-1]
				entries[len(entries)-1] = ghostarmy.Entry{}
				entries = entries[:len(entries)-1]
				i--
			}
		}
		atomic.AddUint64(&remainingEntries, uint64(len(entries)))

		// sort on timestamp
		fmt.Printf("sorting %d entries...\n", len(entries))
		sort.Sort(entries)

		// encode to binary format
		fmt.Println("encoding to buffer...")
		index := 0

		entriesBin := make([]byte, 8)
		encoding.WriteUint64(entriesBin, &index, uint64(len(entries)))
		for _, entry := range entries {
			dat, err := entry.MarshalBinary()
			if err != nil {
				fmt.Printf("unable to marshal binary for session %d: %v\n", entry.SessionID, err)
				continue
			}

			entriesBin = append(entriesBin, dat...)
		}
		bin = append(bin, entriesBin...)

		// export
		fmt.Println("writing to file...")
		if _, err := output.Write(bin); err != nil {
			log.Fatal(err)
		}

		//clean up memory between runs
		entries = nil // free the data buffer now that it's unused
		bin = make([]byte, 0)
		lastDeleteSessionMap = thisDeleteSessionMap
		thisDeleteSessionMap = NewMuxSessionMap()
	}

	fmt.Printf("Total entries: %v \n", totalNumEntries)
	fmt.Printf("Remaining entries: %v \n", remainingEntries)
	fmt.Printf("Total sessions: %v \n", sessionList.Length())
	fmt.Printf("Sessions removed RTT: %v \n", removedRTT)
	fmt.Printf("Sessions removed PL: %v \n", removedPL)
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
