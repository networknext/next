package main

import (
	"time"
	"fmt"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
)

var service *common.Service
var numServers int
var serverAddress string

func main() {

	service = common.CreateService("load_test_servers")

	numServers = envvar.GetInt("NUM_SERVERS", 10000)

	serverAddress = envvar.GetString("SERVER_ADDRESS", "127.0.0.1")

	core.Log("simulating %d servers", numServers)

	go SimulateServers()

	service.WaitForShutdown()
}

func SimulateServers() {
	for i := 0; i < numServers; i++ {
		go RunServer(i)
	}
}

func RunServer(index int) {

	time.Sleep(time.Duration(common.RandomInt(0, 1000)) * time.Millisecond)

	address := core.ParseAddress(fmt.Sprintf("%s:%d", serverAddress, 10000+index))

	ticker := time.NewTicker(time.Second)

	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				fmt.Printf("update server %03d\n", index)

				// ...

				_ = address
			}
		}
	}()

}
