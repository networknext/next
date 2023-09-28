package portal

import (
	// "context"
	// "math/rand"
	// "time"

	// "github.com/networknext/next/modules/encoding"
)

// todo: remove
/*
const MapWidth = 360
const MapHeight = 180
const CellSize = 10
const NumCells = (MapWidth / CellSize) * (MapHeight / CellSize)
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
	SessionId uint64
	Latitude  float32
	Longitude float32
	Next      bool
}

type CellOutput struct {
	Entries []CellEntry
}

type MapCell struct {
	UpdateChan chan *CellUpdate
	OutputChan chan *CellOutput
	Entries    map[uint64]CellEntry
}

func (cell *MapCell) RunCellThread(ctx context.Context) {

	go func() {

		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		ticker := time.NewTicker(time.Second)

		for {
			select {

			case <-ctx.Done():
				return

			case update := <-cell.UpdateChan:
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
				for k, v := range cell.Entries {
					if currentTime-v.LastUpdateTime >= 30 {
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

func CreateMap(ctx context.Context) *Map {
	mapInstance := Map{}
	mapInstance.Cells = make([]MapCell, NumCells)
	for i := range mapInstance.Cells {
		mapInstance.Cells[i].UpdateChan = make(chan *CellUpdate, UpdateChannelSize)
		mapInstance.Cells[i].OutputChan = make(chan *CellOutput, OutputChannelSize)
		mapInstance.Cells[i].Entries = make(map[uint64]CellEntry)
	}
	for i := 0; i < NumCells; i++ {
		mapInstance.Cells[i].RunCellThread(ctx)
	}
	return &mapInstance
}

func GetCellIndex(latitude float32, longitude float32) int {
	if latitude < -90.0 || latitude > +90.0 || longitude < -180.0 || longitude > +180.0 {
		return -1
	}
	x := int((longitude + 180.0) / CellSize)
	y := int((latitude + 90.0) / CellSize)
	index := x + (MapWidth/CellSize)*y
	return index
}

const MapDataVersion = 1

func WriteMapData(entries []CellEntry) []byte {
	size := 1 + 4 + (8+4+4+1)*len(entries)
	data := make([]byte, size)
	index := 0
	encoding.WriteUint8(data, &index, MapDataVersion)
	encoding.WriteUint32(data, &index, uint32(len(entries)))
	for i := range entries {
		encoding.WriteUint64(data, &index, entries[i].SessionId)
		encoding.WriteFloat32(data, &index, entries[i].Latitude)
		encoding.WriteFloat32(data, &index, entries[i].Longitude)
		encoding.WriteBool(data, &index, entries[i].Next)
	}
	return data
}
*/