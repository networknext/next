package main

import (
	"context"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/envvar"
)

var routeMatrixURI string
var routeMatrixInterval time.Duration

func main() {

	service := common.CreateService("analytics")

	routeMatrixURI = envvar.GetString("ROUTE_MATRIX_URI", "http://127.0.0.1:30001")
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", 10 * time.Second)

	core.Log("route matrix uri: %s", routeMatrixURI)
	core.Log("route matrix interval: %s", routeMatrixInterval)

	ProcessRelayStats(service.Context)

	ProcessBilling(service.Context)

	ProcessMatchData(service.Context)

	service.StartWebServer()

	service.WaitForShutdown()
}

func ProcessRelayStats(ctx context.Context) {

	ticker := time.NewTicker(routeMatrixInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			core.Debug("get route matrix")
		}
	}
}

func ProcessBilling(ctx context.Context) {

	// todo: create google pubsub consumer

	/*
	for {
		select {

		case <-service.Context.Done():
			return

		case message := <-consumer.MessageChannel:
			// todo: process message
			core.Debug("received billing message")
			_ = message
		}
	}
	*/
}

func ProcessMatchData(ctx context.Context) {

	// todo: create google pubsub consumer

	/*
	for {
		select {

		case <-service.Context.Done():
			return

		case message := <-consumer.MessageChannel:
			// todo: process message
			core.Debug("received match data message")
			_ = message
		}
	}
	*/
}
