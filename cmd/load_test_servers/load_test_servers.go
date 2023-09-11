package main

import (
	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
)

func main() {

	service := common.CreateService("load_test_servers")

	numServers := envvar.GetInt("NUM_SERVERS", 10000)

	core.Log("simulating %d servers", numServers)

	go SimulateServers(numServers)

	service.WaitForShutdown()
}

func SimulateServers(numServers int) {
	for i := 0; i < numServers; i++ {
		go RunServer()
	}
}

func RunServer() {
	for {
		// ...
	}
}
