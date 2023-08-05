package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"

	"github.com/networknext/next/modules/portal"
)

func getSessionLatLong(sessionId uint64) (float32, float32) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, sessionId)
	hash := fnv.New64a()
	hash.Write(data)
	value := hash.Sum64()
	latitude := float64(value&0xFFFFFFFF)/float64(0xFFFFFFFF)*180.0 - 90.0
	longitude := float64(value>>32)/float64(0xFFFFFFFF)*360.0 - 180.0
	return float32(latitude), float32(longitude)
}

func RunInsertThreads(mapInstance *portal.Map) {

	threadCount := 100

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 10000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration

					latitude, longitude := getSessionLatLong(sessionId)

					cellIndex := portal.GetCellIndex(latitude, longitude)
					if cellIndex == -1 {
						continue
					}

					next := (sessionId % 10) == 0

					update := portal.CellUpdate{}
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

func RunPollThread(mapInstance *portal.Map) {
	go func() {
		iteration := uint64(0)
		previousSize := 0
		for {
			time.Sleep(time.Second)
			start := time.Now()
			entries := make([]portal.CellEntry, 0, previousSize)
			for i := 0; i < portal.NumCells; i++ {
				for {
					var output *portal.CellOutput
					select {
					case output = <-mapInstance.Cells[i].OutputChan:
					default:
					}
					if output == nil {
						break
					}
					entries = append(entries, output.Entries...)
				}
			}
			data := portal.WriteMapData(entries)
			fmt.Printf("iteration %d: %d entries, %d data bytes (%dms)\n", iteration, len(entries), len(data), time.Since(start).Milliseconds())
			previousSize = len(entries)
			iteration++
		}
	}()
}

func main() {

	ctx := context.Background()

	mapInstance := portal.CreateMap(ctx)

	RunInsertThreads(mapInstance)

	RunPollThread(mapInstance)

	time.Sleep(time.Minute)
}
