package main

import (
	"fmt"
	"os"
	"sort"
	"bytes"
	"encoding/gob"

	"github.com/networknext/next/modules/common"
)

func main() {

	fmt.Printf("loading relay manager\n")

	data, err := os.ReadFile("data/relay_manager.bin")
	if err != nil {
		panic(err)
	}

	buffer := bytes.NewReader(data)

	relayManager := common.RelayManager{}

	err = gob.NewDecoder(buffer).Decode(&relayManager)
	if err != nil {
		panic(err)
	}

	fmt.Printf("loaded relay manager\n")

	type RelayData struct {
		relayId uint64
		relayName string
		lastUpdateTime uint64
	}

	relayData := []RelayData{}

	for _,v := range relayManager.SourceEntries {
		relayData = append(relayData, RelayData{relayId: v.RelayId, relayName: v.RelayName, lastUpdateTime: uint64(v.LastUpdateTime)})
	}

	sort.Slice(relayData, func(i, j int) bool {
		return relayData[i].relayName < relayData[j].relayName
	})

	relayIds := make([]uint64, len(relayData))
	for i := range relayData {
		relayIds[i] = relayData[i].relayId
	}

	oldestLastUpdateTime := uint64(0xFFFFFFFFFFFFFFFF)
	for i := range relayData {
		if relayData[i].lastUpdateTime < oldestLastUpdateTime {
			oldestLastUpdateTime = relayData[i].lastUpdateTime
		}
	}

	fmt.Printf("oldest update time is %d\n", oldestLastUpdateTime)

	for i := range relayData {
		deltaTime := relayData[i].lastUpdateTime - oldestLastUpdateTime
		fmt.Printf("%s [+%d]\n", relayData[i].relayName, deltaTime)
	}

	fmt.Printf("adjusting timestamps\n")

	for _,u := range relayManager.SourceEntries {
		u.LastUpdateTime = int64(oldestLastUpdateTime)
		for _,v := range u.DestEntries {
			v.LastUpdateTime = int64(oldestLastUpdateTime)
		}
	}

	fmt.Printf("getting costs\n")

	costs := relayManager.GetCosts(int64(oldestLastUpdateTime), relayIds, 100.0, 100.0)

	fmt.Printf("costs: %v\n", costs[:100])
}
