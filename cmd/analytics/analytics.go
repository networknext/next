package main

import (
	"context"
	"time"
	"io/ioutil"
	"net/http"
	
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/envvar"
)

var costMatrixURI string
var routeMatrixURI string
var costMatrixInterval time.Duration
var routeMatrixInterval time.Duration

func main() {

	service := common.CreateService("analytics")

	costMatrixURI = envvar.GetString("COST_MATRIX_URI", "http://127.0.0.1:30001/cost_matrix")
	routeMatrixURI = envvar.GetString("ROUTE_MATRIX_URI", "http://127.0.0.1:30001/route_matrix")
	costMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", 10 * time.Second)
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", 10 * time.Second)

	core.Log("cost matrix uri: %s", costMatrixURI)
	core.Log("route matrix uri: %s", routeMatrixURI)
	core.Log("cost matrix interval: %s", costMatrixInterval)
	core.Log("route matrix interval: %s", routeMatrixInterval)

	ProcessCostMatrix(service.Context)

	ProcessRouteMatrix(service.Context)

	ProcessBilling(service.Context)

	ProcessMatchData(service.Context)

	service.StartWebServer()

	service.WaitForShutdown()
}

func ProcessCostMatrix(ctx context.Context) {

	httpClient := &http.Client{
		Timeout: routeMatrixInterval,
	}

	ticker := time.NewTicker(routeMatrixInterval)

	go func() {	
		for {
			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				core.Debug("get cost matrix")

				response, err := httpClient.Get(routeMatrixURI)
				if err != nil {
					core.Error("failed to http get cost matrix: %v", err)
					continue
				}

				buffer, err := ioutil.ReadAll(response.Body)
				if err != nil {
					core.Error("failed to read cost matrix: %v", err)
					continue
				}

				response.Body.Close()

				// todo: read in cost matrix
				_ = buffer
			}
		}
	}()
}

func ProcessRouteMatrix(ctx context.Context) {

	httpClient := &http.Client{
		Timeout: routeMatrixInterval,
	}

	ticker := time.NewTicker(routeMatrixInterval)

	go func() {	
		for {
			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				core.Debug("get route matrix")

				response, err := httpClient.Get(routeMatrixURI)
				if err != nil {
					core.Error("failed to http get route matrix: %v", err)
					continue
				}

				buffer, err := ioutil.ReadAll(response.Body)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					continue
				}

				response.Body.Close()

				// todo: read in route matrix
				_ = buffer
			}
		}
	}()
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
