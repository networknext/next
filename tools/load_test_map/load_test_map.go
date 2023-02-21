package main

import (
	"fmt"
	"time"
	"math/rand"
	"context"
	"encoding/binary"
	"hash/fnv"
)

const MapWidth = 360
const MapHeight = 180
const CellSize = 10
const NumCells = (MapWidth/CellSize) * (MapHeight/CellSize)
const UpdateChannelSize = 10 * 1024
const OutputChannelSize = 1024

type CellEntry struct {
	SessionId      uint64
	Latitude       float32
	Longitude      float32
	LastUpdateTime uint64
	Next           bool
}

type CellUpdate struct {
	SessionId      uint64
	Latitude       float32
	Longitude      float32
	Next           bool
}

type CellOutput struct {
	Entries []CellEntry
}

type MapCell struct {
	UpdateChan chan *CellUpdate
	OutputChan chan *CellOutput
	Entries map[uint64]CellEntry
}

func (cell *MapCell) RunCellThread(ctx context.Context) {

	go func() {

		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		ticker := time.NewTicker(time.Second)

		for {
			select {

			case <-ctx.Done():
				return

			case update := <- cell.UpdateChan:
				entry := CellEntry{}
				entry.SessionId = update.SessionId
				entry.Latitude = update.Latitude
				entry.Longitude = update.Latitude
				entry.LastUpdateTime = uint64(time.Now().Unix())
				entry.Next = update.Next
				cell.Entries[update.SessionId] = entry
				break

			case <-ticker.C:
				output := CellOutput{}
				output.Entries = make([]CellEntry, 0, len(cell.Entries))
				currentTime := uint64(time.Now().Unix())
				for k,v := range cell.Entries {
					if currentTime - v.LastUpdateTime >= 30 {
						delete(cell.Entries, k)
						continue
					}
					output.Entries = append(output.Entries, v)
				}
				cell.OutputChan <- &output
			}
		}
	}()
}

type Map struct {
	Cells []MapCell
}

func CreateMap() *Map {
	mapInstance := Map{}
	mapInstance.Cells = make([]MapCell, NumCells)
	for i := range mapInstance.Cells {
		mapInstance.Cells[i].UpdateChan = make(chan *CellUpdate, UpdateChannelSize)
		mapInstance.Cells[i].OutputChan = make(chan *CellOutput, OutputChannelSize)
		mapInstance.Cells[i].Entries = make(map[uint64]CellEntry)
	}
	return &mapInstance
}

func getSessionLatLong(sessionId uint64) (float32, float32) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, sessionId)
	hash := fnv.New64a()
	hash.Write(data)
	value := hash.Sum64()
	latitude := float64(value&0xFFFFFFFF) / float64(0xFFFFFFFF) * 180.0 - 90.0
	longitude := float64(value>>32) / float64(0xFFFFFFFF) * 360.0 - 180.0
	return float32(latitude), float32(longitude)
}

func getCellIndex(latitude float32, longitude float32) int {
	if latitude < -90.0 || latitude > +90.0 || longitude < -180.0 || longitude > +180.0 {
		return -1
	}
	x := int( ( longitude + 180.0 ) / CellSize )
	y := int( ( latitude + 90.0 ) / CellSize )
	index := x + (MapWidth/CellSize) * y
	return index
}

func RunInsertThreads(mapInstance *Map) {

	threadCount := 100

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 10000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration

					latitude, longitude := getSessionLatLong(sessionId)

					cellIndex := getCellIndex(latitude, longitude)
					if cellIndex == -1 {
						continue
					}

					next := (sessionId % 10) == 0

					update := CellUpdate{}
					update.SessionId = sessionId
					update.Latitude = latitude
					update.Longitude = longitude
					update.Next = next

					mapInstance.Cells[cellIndex].UpdateChan <- &update
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func RunPollThread(mapInstance *Map) {
	iteration := uint64(0)
	previousSize := 0
	for {
		time.Sleep(time.Second)
		start := time.Now()
		entries := make([]CellEntry, 0, previousSize)
		for i := 0; i < NumCells; i++ {
			for {
				var output *CellOutput
				select {
                    case output = <- mapInstance.Cells[i].OutputChan:
					default:
				}
				if output == nil {
					break
				}
				entries = append(entries, output.Entries...)
			}
		}
		fmt.Printf("iteration %d: %d entries (%dms)\n", iteration, len(entries), time.Since(start).Milliseconds())
		previousSize = len(entries)
		iteration++
	}
}

func main() {

	fmt.Printf("\nmap cruncher\n")

	mapInstance := CreateMap()

	ctx := context.Background()

	for i := 0; i < NumCells; i++ {
		mapInstance.Cells[i].RunCellThread(ctx)
	}

	RunInsertThreads(mapInstance)
	RunPollThread(mapInstance)

	time.Sleep(time.Minute)
}

/*
	mapData := MapData{}
	mapData.Latitude = sessionData.Latitude
	mapData.Longitude = sessionData.Longitude
	mapData.Next = next
	mapData.LastUpdateTime = uint64(currentTime.Unix())
	inserter.redisClient.Send("HSET", fmt.Sprintf("m-%d", minutes), fmt.Sprintf("%016x", sessionId), mapData.Value())
	inserter.redisClient.Send("EXPIRE", fmt.Sprintf("m-%d", minutes), 30)
*/
