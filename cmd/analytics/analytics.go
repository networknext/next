package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
)

var costMatrixURI string
var routeMatrixURI string
var costMatrixInterval time.Duration
var routeMatrixInterval time.Duration

var logMutex sync.Mutex

func main() {

	service := common.CreateService("analytics")

	costMatrixURI = envvar.GetString("COST_MATRIX_URI", "http://127.0.0.1:30001/cost_matrix")
	routeMatrixURI = envvar.GetString("ROUTE_MATRIX_URI", "http://127.0.0.1:30001/route_matrix")
	costMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", 1*time.Second)
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", 1*time.Second)

	core.Log("cost matrix uri: %s", costMatrixURI)
	core.Log("route matrix uri: %s", routeMatrixURI)
	core.Log("cost matrix interval: %s", costMatrixInterval)
	core.Log("route matrix interval: %s", routeMatrixInterval)

	ProcessCostMatrix(service)

	ProcessRouteMatrix(service)

	/*
	ProcessBilling(service)

	ProcessMatchData(service)
	*/

	service.LeaderElection()

	service.StartWebServer()

	service.WaitForShutdown()
}

func ProcessCostMatrix(service *common.Service) {

	httpClient := &http.Client{
		Timeout: costMatrixInterval,
	}

	ticker := time.NewTicker(costMatrixInterval)

	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				if !service.IsLeader() {
					break
				}

				response, err := httpClient.Get(costMatrixURI)
				if err != nil {
					core.Error("failed to http get cost matrix: %v", err)
					continue
				}

				buffer, err := ioutil.ReadAll(response.Body)
				if err != nil {
					core.Error("failed to read cost matrix data: %v", err)
					continue
				}

				response.Body.Close()

				costMatrix := common.CostMatrix{}

				err = costMatrix.Read(buffer)
				if err != nil {
					core.Error("failed to read cost matrix: %v", err)
					continue
				}

				costMatrixBytes := len(buffer)
				costMatrixNumRelays := len(costMatrix.RelayIds)

				costMatrixNumDestRelays := 0
				for i := range costMatrix.DestRelays {
					if costMatrix.DestRelays[i] {
						costMatrixNumDestRelays++
					}
				}

				datacenterMap := make(map[uint64]bool)
				for i := range costMatrix.RelayDatacenterIds {
					datacenterMap[costMatrix.RelayDatacenterIds[i]] = true
				}
				costMatrixNumDatacenters := len(datacenterMap)

				logMutex.Lock()

				core.Debug("---------------------------------------------")
				core.Debug("cost matrix bytes: %d", costMatrixBytes)
				core.Debug("cost matrix num relays: %d", costMatrixNumRelays)
				core.Debug("cost matrix num dest relays: %d", costMatrixNumDestRelays)
				core.Debug("cost matrix num datacenters: %d", costMatrixNumDatacenters)
				core.Debug("---------------------------------------------")

				logMutex.Unlock()
			}
		}
	}()
}

func ProcessRouteMatrix(service *common.Service) {

	httpClient := &http.Client{
		Timeout: routeMatrixInterval,
	}

	ticker := time.NewTicker(routeMatrixInterval)

	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				if !service.IsLeader() {
					break
				}

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

				routeMatrix := common.RouteMatrix{}

				err = routeMatrix.Read(buffer)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					continue
				}

				logMutex.Lock()

				routeMatrixBytes := len(buffer)
				routeMatrixNumRelays := len(routeMatrix.RelayIds)

				routeMatrixNumDestRelays := 0
				for i := range routeMatrix.DestRelays {
					if routeMatrix.DestRelays[i] {
						routeMatrixNumDestRelays++
					}
				}

				datacenterMap := make(map[uint64]bool)
				for i := range routeMatrix.RelayDatacenterIds {
					datacenterMap[routeMatrix.RelayDatacenterIds[i]] = true
				}
				routeMatrixNumDatacenters := len(datacenterMap)

				routeMatrixNumFullRelays := len(routeMatrix.FullRelayIds)

				analysis := routeMatrix.Analyze()

				core.Debug("---------------------------------------------")

				core.Debug("route matrix bytes: %d", routeMatrixBytes)

				core.Debug("route matrix num relays: %d", routeMatrixNumRelays)
				core.Debug("route matrix num dest relays: %d", routeMatrixNumDestRelays)
				core.Debug("route matrix num full relays: %d", routeMatrixNumFullRelays)
				core.Debug("route matrix num datacenters: %d", routeMatrixNumDatacenters)

				core.Debug("route matrix total routes: %d", analysis.TotalRoutes)
				core.Debug("route matrix num relay pairs: %d", analysis.NumRelayPairs)
				core.Debug("route matrix num valid relay pairs: %d", analysis.NumValidRelayPairs)
				core.Debug("route matrix num valid relay pairs without improvement: %d", analysis.NumValidRelayPairsWithoutImprovement)
				core.Debug("route matrix num relay pairs with no routes: %d", analysis.NumRelayPairsWithNoRoutes)
				core.Debug("route matrix num relay pairs with one route: %d", analysis.NumRelayPairsWithOneRoute)
				core.Debug("route matrix average num routes: %.1f", analysis.AverageNumRoutes)
				core.Debug("route matrix average route length: %.1f", analysis.AverageRouteLength)

				core.Debug("route matrix rtt bucket no improvement: %.1f%%", analysis.RTTBucket_NoImprovement)
				core.Debug("route matrix rtt bucket 0-5ms: %.1f%%", analysis.RTTBucket_0_5ms)
				core.Debug("route matrix rtt bucket 5-10ms: %.1f%%", analysis.RTTBucket_5_10ms)
				core.Debug("route matrix rtt bucket 10-15ms: %.1f%%", analysis.RTTBucket_10_15ms)
				core.Debug("route matrix rtt bucket 15-20ms: %.1f%%", analysis.RTTBucket_15_20ms)
				core.Debug("route matrix rtt bucket 20-25ms: %.1f%%", analysis.RTTBucket_20_25ms)
				core.Debug("route matrix rtt bucket 25-30ms: %.1f%%", analysis.RTTBucket_25_30ms)
				core.Debug("route matrix rtt bucket 30-35ms: %.1f%%", analysis.RTTBucket_30_35ms)
				core.Debug("route matrix rtt bucket 35-40ms: %.1f%%", analysis.RTTBucket_35_40ms)
				core.Debug("route matrix rtt bucket 40-45ms: %.1f%%", analysis.RTTBucket_40_45ms)
				core.Debug("route matrix rtt bucket 45-50ms: %.1f%%", analysis.RTTBucket_45_50ms)
				core.Debug("route matrix rtt bucket 50ms+: %.1f%%", analysis.RTTBucket_50ms_Plus)

				totalPercent := analysis.RTTBucket_NoImprovement +
					analysis.RTTBucket_0_5ms +
					analysis.RTTBucket_5_10ms +
					analysis.RTTBucket_10_15ms +
					analysis.RTTBucket_15_20ms +
					analysis.RTTBucket_20_25ms +
					analysis.RTTBucket_25_30ms +
					analysis.RTTBucket_30_35ms +
					analysis.RTTBucket_35_40ms +
					analysis.RTTBucket_40_45ms +
					analysis.RTTBucket_45_50ms +
					analysis.RTTBucket_50ms_Plus

				core.Debug("route matrix rtt bucket total percent: %.1f%%", totalPercent)

				core.Debug("---------------------------------------------")

				logMutex.Unlock()
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
