package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/messages"
)

var costMatrixURI string
var routeMatrixURI string
var costMatrixInterval time.Duration
var routeMatrixInterval time.Duration
var googleProjectId string

var logMutex sync.Mutex

func main() {

	service := common.CreateService("analytics")

	costMatrixURI = envvar.GetString("COST_MATRIX_URI", "http://127.0.0.1:30001/cost_matrix")
	routeMatrixURI = envvar.GetString("ROUTE_MATRIX_URI", "http://127.0.0.1:30001/route_matrix")
	costMatrixInterval = envvar.GetDuration("COST_MATRIX_INTERVAL", 1*time.Second)
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", 1*time.Second)
	googleProjectId = envvar.GetString("GOOGLE_PROJECT_ID", "local")

	core.Log("cost matrix uri: %s", costMatrixURI)
	core.Log("route matrix uri: %s", routeMatrixURI)
	core.Log("cost matrix interval: %s", costMatrixInterval)
	core.Log("route matrix interval: %s", routeMatrixInterval)
	core.Log("google project id: %s", googleProjectId)

	ProcessCostMatrix(service)

	ProcessRouteMatrix(service)

	// todo
	/*
	Process[messages.BillingEntry](service, "billing")
	Process[messages.SummaryEntry](service, "summary")
	Process[messages.MatchDataEntry](service, "match_data")
	Process[messages.PingStatsEntry](service, "ping_stats")
	Process[messages.RelayStatsEntry](service, "relay_stats")
	*/

	Process[*messages.CostMatrixStatsEntry](service, "cost_matrix_stats")
	Process[*messages.RouteMatrixStatsEntry](service, "route_matrix_stats")

	service.StartWebServer()

	service.LeaderElection()

	service.WaitForShutdown()
}

// --------------------------------------------------------------------

func Process[T messages.Message](service *common.Service, name string) {

	envPrefix := strings.ToUpper(name) + "_"

	pubsubTopic := envvar.GetString(envPrefix+"PUBSUB_TOPIC", name)
	bigqueryTable := envvar.GetString(envPrefix+"BIGQUERY_TABLE", name)

	core.Debug("%s pubsub topic: %s", name, pubsubTopic)
	core.Debug("%s bigquery table: %s", name, bigqueryTable)

	config := common.GooglePubsubConfig{Topic: pubsubTopic, BatchDuration: 10 * time.Second}

	consumer, err := common.CreateGooglePubsubConsumer(service.Context, config)
	if err != nil {
		core.Error("could not create google pubsub consumer for %s: %v", name, err)
		os.Exit(1)
	}

	core.Debug("processing %s messages", name)

	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case pubsubMessage := <-consumer.MessageChannel:

				core.Debug("received %s message", name)
				
				messageData := pubsubMessage.Data
				var message T
				err := message.Read(messageData)
				if err != nil {
					core.Error("could not read %s message", name)
					break
				}
				
				// todo: insert into bigquery
				
				insert_ok := true
				if insert_ok {
					pubsubMessage.Ack()
				} else {
					pubsubMessage.Nack()
				}
			}
		}
	}()
}

// --------------------------------------------------------------------

func ProcessCostMatrix(service *common.Service) {

	maxBytes := envvar.GetInt("COST_MATRIX_STATS_ENTRY_MAX_BYTES", 1024)
	pubsubTopic := envvar.GetString("PUBSUB_TOPIC", "cost_matrix_stats")
	pubsubSubscription := envvar.GetString("PUBSUB_SUBSCRIPTION", "cost_matrix_stats")

	core.Log("cost matrix stats entry max bytes: %d", maxBytes)
	core.Log("cost matrix stats entry pubsub topic: %s", pubsubTopic)
	core.Log("cost matrix stats entry pubsub subscription: %s", pubsubSubscription)

	httpClient := &http.Client{
		Timeout: costMatrixInterval,
	}

	config := common.GooglePubsubConfig{
		ProjectId:          googleProjectId,
		Topic:              pubsubTopic,
		Subscription:       pubsubSubscription,
		MessageChannelSize: 10 * 1024,
	}

	statsPubsubProducer, err := common.CreateGooglePubsubProducer(service.Context, config)
	if err != nil {
		core.Error("could not create google pubsub producer for processing cost matrix: %v", err)
		os.Exit(1)
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

				// send cost matrix entry via pubsub

				costMatrixStatsEntry := messages.CostMatrixStatsEntry{}

				costMatrixStatsEntry.Version = messages.CostMatrixStatsVersion
				costMatrixStatsEntry.Bytes = costMatrixBytes
				costMatrixStatsEntry.NumRelays = costMatrixNumRelays
				costMatrixStatsEntry.NumDestRelays = costMatrixNumDestRelays
				costMatrixStatsEntry.NumDatacenters = costMatrixNumDatacenters

				message := costMatrixStatsEntry.Write(make([]byte, maxBytes))

				statsPubsubProducer.MessageChannel <- message

				core.Debug("cost matrix stats message is %d bytes", len(message))
			}
		}
	}()
}

func ProcessRouteMatrix(service *common.Service) {

	pubsubTopic := envvar.GetString("PUBSUB_TOPIC", "route_matrix_stats")
	pubsubSubscription := envvar.GetString("PUBSUB_SUBSCRIPTION", "route_matrix_stats")

	core.Log("route matrix stats entry google project id: %s", googleProjectId)
	core.Log("route matrix stats entry pubsub topic: %s", pubsubTopic)
	core.Log("route matrix stats entry pubsub subscription: %s", pubsubSubscription)

	config := common.GooglePubsubConfig{
		ProjectId:          googleProjectId,
		Topic:              pubsubTopic,
		Subscription:       pubsubSubscription,
		MessageChannelSize: 10 * 1024,
	}

	statsPubsubProducer, err := common.CreateGooglePubsubProducer(service.Context, config)
	if err != nil {
		core.Error("could not create google pubsub producer for processing route matrix: %v", err)
		os.Exit(1)
	}

	maxBytes := envvar.GetInt("COST_MATRIX_STATS_ENTRY_MAX_BYTES", 1024)

	core.Log("cost matrix stats entry max bytes: %d", maxBytes)

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
				core.Debug("route matrix average num routes: %.1f", analysis.AverageNumRoutes)
				core.Debug("route matrix average route length: %.1f", analysis.AverageRouteLength)
				core.Debug("no route percent: %.1f%%", analysis.NoRoutePercent)
				core.Debug("one route percent: %.1f%%", analysis.OneRoutePercent)
				core.Debug("no direct route percent: %.1f%%", analysis.NoDirectRoutePercent)

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

				// send route matrix stats via pubsub
				routeMatrixStatsEntry := messages.RouteMatrixStatsEntry{}

				routeMatrixStatsEntry.Version = messages.RouteMatrixStatsVersion
				routeMatrixStatsEntry.Bytes = routeMatrixBytes
				routeMatrixStatsEntry.NumRelays = routeMatrixNumRelays
				routeMatrixStatsEntry.NumDestRelays = routeMatrixNumDestRelays
				routeMatrixStatsEntry.NumDatacenters = routeMatrixNumDatacenters

				message := routeMatrixStatsEntry.Write(make([]byte, maxBytes))

				statsPubsubProducer.MessageChannel <- message
				_ = message

				core.Debug("route matrix stats message is %d bytes", len(message))
			}
		}
	}()
}

// --------------------------------------------------------------------
