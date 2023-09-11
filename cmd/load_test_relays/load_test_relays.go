package main

import (
	"fmt"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
)


func main() {

	service := common.CreateService("load_test_relays")

	numRelays := envvar.GetInt("NUM_RELAYS", 1000)

	core.Log("simulating %d relays", numRelays)

	go SimulateRelays(service, numRelays)

	service.WaitForShutdown()
}

func SimulateRelays(service *common.Service, numRelays int) {
	for i := 0; i < numRelays; i++ {
		go RunRelay(service, i)
	}
}

func RunRelay(service *common.Service, index int) {
	time.Sleep(time.Duration(common.RandomInt(0,1000))*time.Millisecond)
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				// todo: send relay update to relay backend
				fmt.Printf("update relay %d\n", index)
			}
		}
	}()
}
