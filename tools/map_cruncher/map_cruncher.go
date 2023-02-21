package main

import (
	"fmt"
	"time"
	"math/rand"
	"context"
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
				fmt.Printf("cell processing update: %v\n", update)
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

func RunInsertThreads(mapInstance *Map) {
	// todo
	mapInstance.Cells[0].UpdateChan <- &CellUpdate{}
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
