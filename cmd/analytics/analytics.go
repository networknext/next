package main

import (
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	"os"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
)

var costMatrixURI string
var routeMatrixURI string
var costMatrixInterval time.Duration
var routeMatrixInterval time.Duration

var billingPubsubTopic string
var matchDataPubsubTopic string
var pingStatsPubsubTopic string
var relayStatsPubsubTopic string
var costMatrixStatsPubsubTopic string
var routeMatrixStatsPubsubTopic string

var billingBigQueryTable string
var matchDataBigQueryTable string
var pingStatsBigQueryTable string
var relayStatsBigQueryTable string
var costMatrixStatsBigQueryTable string
var routeMatrixStatsBigQueryTable string

var logMutex sync.Mutex

func main() {

	service := common.CreateService("analytics")

	costMatrixURI = envvar.GetString("COST_MATRIX_URI", "http://127.0.0.1:30001/cost_matrix")
	routeMatrixURI = envvar.GetString("ROUTE_MATRIX_URI", "http://127.0.0.1:30001/route_matrix")
	costMatrixInterval = envvar.GetDuration("COST_MATRIX_INTERVAL", 1*time.Second)
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", 1*time.Second)

	billingPubsubTopic = envvar.GetString("BILLING_PUBSUB_TOPIC", "billing")
	matchDataPubsubTopic = envvar.GetString("MATCH_DATA_PUBSUB_TOPIC", "match_data")
	pingStatsPubsubTopic = envvar.GetString("PING_STATS_PUBSUB_TOPIC", "ping_stats")
	relayStatsPubsubTopic = envvar.GetString("RELAY_STATS_PUBSUB_TOPIC", "relay_stats")
	costMatrixStatsPubsubTopic = envvar.GetString("COST_MATRIX_STATS_PUBSUB_TOPIC", "cost_matrix_stats")
	routeMatrixStatsPubsubTopic = envvar.GetString("ROUTE_MATRIX_STATS_PUBSUB_TOPIC", "route_matrix_stats")

	billingBigQueryTable = envvar.GetString("BILLING_BIGQUERY_TABLE", "billing2")
	matchDataBigQueryTable = envvar.GetString("MATCH_DATA_BIGQUERY_TABLE", "match_data")
	pingStatsBigQueryTable = envvar.GetString("PING_STATS_BIGQUERY_TABLE", "ping_stats")
	relayStatsBigQueryTable = envvar.GetString("RELAY_STATS_BIGQUERY_TABLE", "relay_stats")
	costMatrixStatsBigQueryTable = envvar.GetString("COST_MATRIX_STATS_BIGQUERY_TABLE", "cost_matrix_stats")
	routeMatrixStatsBigQueryTable = envvar.GetString("ROUTE_MATRIX_STATS_BIGQUERY_TABLE", "route_matrix_stats")

	core.Log("cost matrix uri: %s", costMatrixURI)
	core.Log("route matrix uri: %s", routeMatrixURI)
	core.Log("cost matrix interval: %s", costMatrixInterval)
	core.Log("route matrix interval: %s", routeMatrixInterval)

	core.Log("billing pubsub topic: %s", billingPubsubTopic)
	core.Log("match data pubsub topic: %s", matchDataPubsubTopic)
	core.Log("ping stats pubsub topic: %s", pingStatsPubsubTopic)
	core.Log("relay stats pubsub topic: %s", relayStatsPubsubTopic)
	core.Log("cost matrix stats pubsub topic: %s", costMatrixStatsPubsubTopic)
	core.Log("route matrix stats pubsub topic: %s", routeMatrixStatsPubsubTopic)

	core.Log("billing bigquery table: %s", billingBigQueryTable)
	core.Log("match data bigquery table: %s", matchDataBigQueryTable)
	core.Log("ping stats bigquery table: %s", pingStatsBigQueryTable)
	core.Log("relay stats bigquery table: %s", relayStatsBigQueryTable)
	core.Log("cost matrix stats bigquery table: %s", costMatrixStatsBigQueryTable)
	core.Log("route matrix stats bigquery table: %s", routeMatrixStatsBigQueryTable)

	ProcessCostMatrix(service)

	ProcessRouteMatrix(service)

	ProcessBilling(service)

	ProcessMatchData(service)

	ProcessPingStats(service)

	ProcessRelayStats(service)

	ProcessCostMatrixStats(service)

	ProcessRouteMatrixStats(service)

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

				// todo: send cost matrix stats via pubsub
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

				// todo: send route matrix stats via pubsub
			}
		}
	}()
}

func ProcessBilling(service *common.Service) {

	config := common.GooglePubsubConfig{Topic: billingPubsubTopic}

	consumer, err := common.CreateGooglePubsubConsumer(service.Context, config)
	if err != nil {
		core.Error("could not create google pubsub consumer: %v", err)
		os.Exit(1)
	}

	core.Debug("processing billing messages")

	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case message := <-consumer.MessageChannel:
				core.Debug("received billing message")
				_ = message
				// todo: parse billing message
				// todo: publish billing message to bigquery
			}
		}
	}()
}

func ProcessMatchData(service *common.Service) {

	config := common.GooglePubsubConfig{Topic: matchDataPubsubTopic}

	consumer, err := common.CreateGooglePubsubConsumer(service.Context, config)
	if err != nil {
		core.Error("could not create google pubsub consumer: %v", err)
		os.Exit(1)
	}

	core.Debug("processing match data messages")

	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case message := <-consumer.MessageChannel:
				core.Debug("received match data message")
				_ = message
				// todo: process match data message
			}
		}
	}()
}

func ProcessPingStats(service *common.Service) {

	config := common.GooglePubsubConfig{Topic: pingStatsPubsubTopic}

	consumer, err := common.CreateGooglePubsubConsumer(service.Context, config)
	if err != nil {
		core.Error("could not create google pubsub consumer: %v", err)
		os.Exit(1)
	}

	core.Debug("processing ping stats messages")

	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case message := <-consumer.MessageChannel:
				core.Debug("received ping stats message")
				_ = message
				// todo: process ping stats message
			}
		}
	}()
}

func ProcessRelayStats(service *common.Service) {

	config := common.GooglePubsubConfig{Topic: relayStatsPubsubTopic}

	consumer, err := common.CreateGooglePubsubConsumer(service.Context, config)
	if err != nil {
		core.Error("could not create google pubsub consumer: %v", err)
		os.Exit(1)
	}

	core.Debug("processing relay stats messages")

	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case message := <-consumer.MessageChannel:
				core.Debug("received relay stats message")
				_ = message
				// todo: process relay stats message
			}
		}
	}()
}


func ProcessCostMatrixStats(service *common.Service) {

	config := common.GooglePubsubConfig{Topic: costMatrixStatsPubsubTopic}

	consumer, err := common.CreateGooglePubsubConsumer(service.Context, config)
	if err != nil {
		core.Error("could not create google pubsub consumer: %v", err)
		os.Exit(1)
	}

	core.Debug("processing cost matrix stats messages")

	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case message := <-consumer.MessageChannel:
				core.Debug("received cost matrix stats message")
				_ = message
				// todo: process cost matrix stats message
			}
		}
	}()
}

func ProcessRouteMatrixStats(service *common.Service) {

	config := common.GooglePubsubConfig{Topic: routeMatrixStatsPubsubTopic}

	consumer, err := common.CreateGooglePubsubConsumer(service.Context, config)
	if err != nil {
		core.Error("could not create google pubsub consumer: %v", err)
		os.Exit(1)
	}

	core.Debug("processing route matrix stats messages")

	go func() {
		for {
			select {
			case <-service.Context.Done():
				return
			case message := <-consumer.MessageChannel:
				core.Debug("received route matrix stats message")
				_ = message
				// todo: process route matrix stats message
			}
		}
	}()
}
