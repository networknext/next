package main

import (
	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
)

func main() {

	service := common.CreateService("load_test_sessions")

	numSessions := envvar.GetInt("NUM_SESSIONS", 100000)

	core.Log("simulating %d sessions", numSessions)

	go SimulateSessions(numSessions)

	service.WaitForShutdown()
}

func SimulateSessions(numSessions int) {
	for i := 0; i < numSessions; i++ {
		go RunSession()
	}
}

func RunSession() {
	for {
		// ...
	}
}
