package main

import (
	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
)


func main() {

	service := common.CreateService("load_test_relays")

	numRelays := envvar.GetInt("NUM_RELAYS", 1000)

	core.Log("simulating %d relays", numRelays)

	go SimulateRelays(numRelays)

	service.WaitForShutdown()
}

func SimulateRelays(numRelays int) {
	for i := 0; i < numRelays; i++ {
		// ...
	}
}

func RunRelay() {
	for {
		// ...
	}
}