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
var bigqueryDataset string

var logMutex sync.Mutex

func main() {

	service := common.CreateService("analytics")

	costMatrixURI = envvar.GetString("COST_MATRIX_URI", "http://127.0.0.1:30001/cost_matrix")
	routeMatrixURI = envvar.GetString("ROUTE_MATRIX_URI", "http://127.0.0.1:30001/route_matrix")
	costMatrixInterval = envvar.GetDuration("COST_MATRIX_INTERVAL", time.Second)
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)
	googleProjectId = envvar.GetString("GOOGLE_PROJECT_ID", "local")
	bigqueryDataset = envvar.GetString("BIGQUERY_DATASET", service.Env)

	core.Log("cost matrix uri: %s", costMatrixURI)
	core.Log("route matrix uri: %s", routeMatrixURI)
	core.Log("cost matrix interval: %s", costMatrixInterval)
	core.Log("route matrix interval: %s", routeMatrixInterval)
	core.Log("google project id: %s", googleProjectId)
	core.Log("bigquery dataset: %s", bigqueryDataset)

	ProcessCostMatrix(service)

	ProcessRouteMatrix(service)

	Process[*messages.CostMatrixStatsMessage](service, "cost_matrix_stats", &messages.CostMatrixStatsMessage{}, false)
	Process[*messages.RouteMatrixStatsMessage](service, "route_matrix_stats", &messages.RouteMatrixStatsMessage{}, false)
	Process[*messages.PingStatsMessage](service, "ping_stats", &messages.PingStatsMessage{}, false)
	Process[*messages.RelayStatsMessage](service, "relay_stats", &messages.RelayStatsMessage{}, false)
	Process[*messages.MatchDataMessage](service, "match_data", &messages.MatchDataMessage{}, false)

	service.StartWebServer()

	service.LeaderElection()

	service.WaitForShutdown()
}

// --------------------------------------------------------------------

func Process[T messages.Message](service *common.Service, name string, message messages.Message, important bool) {

	namePrefix := strings.ToUpper(name) + "_"

	pubsubTopic := envvar.GetString(namePrefix+"PUBSUB_TOPIC", name)
	pubsubSubscription := envvar.GetString(namePrefix+"PUBSUB_SUBSCRIPTION", name)
	bigqueryTable := envvar.GetString(namePrefix+"BIGQUERY_TABLE", name)

	core.Debug("%s pubsub topic: %s", name, pubsubTopic)
	core.Debug("%s pubsub subscription: %s", name, pubsubSubscription)
	core.Debug("%s bigquery table: %s", name, bigqueryTable)

	consumerConfig := common.GooglePubsubConfig{
		ProjectId:     googleProjectId,
		Subscription:  pubsubSubscription,
		Topic:         pubsubTopic,
		BatchDuration: 10 * time.Second,
	}

	publisherConfig := common.GoogleBigQueryConfig{
		ProjectId:     googleProjectId,
		Dataset:       bigqueryDataset,
		TableName:     bigqueryTable,
		BatchSize:     100,
		BatchDuration: 10 * time.Second,
	}

	consumer, err := common.CreateGooglePubsubConsumer(service.Context, consumerConfig)
	if err != nil {
		core.Error("could not create google pubsub consumer for %s: %v", name, err)
		os.Exit(1)
	}

	publisher, err := common.CreateGoogleBigQueryPublisher(service.Context, publisherConfig)
	if err != nil {
		core.Error("could not create google bigquery publisher for %s: %v", name, err)
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

				err := message.Read(messageData)
				if err != nil {
					if !important {
						core.Error("could not read %s message - dropping", name)
						pubsubMessage.Ack()
						break
					}

					core.Error("could not read %s message", name)
					pubsubMessage.Nack()
					break
				}

				publisher.PublishChannel <- message

				pubsubMessage.Ack()
			}
		}
	}()
}

// --------------------------------------------------------------------

func ProcessCostMatrix(service *common.Service) {

	maxBytes := envvar.GetInt("COST_MATRIX_STATS_MESSAGE_MAX_BYTES", 1024)
	pubsubTopic := envvar.GetString("COST_MATRIX_STATS_PUBSUB_TOPIC", "cost_matrix_stats")

	core.Log("cost matrix stats message max bytes: %d", maxBytes)
	core.Log("cost matrix stats message pubsub topic: %s", pubsubTopic)

	httpClient := &http.Client{
		Timeout: costMatrixInterval,
	}

	config := common.GooglePubsubConfig{
		ProjectId:          googleProjectId,
		Topic:              pubsubTopic,
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

				// send cost matrix message via pubsub

				costMatrixStatsMessage := messages.CostMatrixStatsMessage{}

				costMatrixStatsMessage.Version = messages.CostMatrixStatsMessageVersion
				costMatrixStatsMessage.Bytes = costMatrixBytes
				costMatrixStatsMessage.NumRelays = costMatrixNumRelays
				costMatrixStatsMessage.NumDestRelays = costMatrixNumDestRelays
				costMatrixStatsMessage.NumDatacenters = costMatrixNumDatacenters

				message := costMatrixStatsMessage.Write(make([]byte, maxBytes))

				statsPubsubProducer.MessageChannel <- message
			}
		}
	}()
}

func ProcessRouteMatrix(service *common.Service) {

	pubsubTopic := envvar.GetString("ROUTE_MATRIX_STATS_PUBSUB_TOPIC", "route_matrix_stats")

	core.Log("route matrix stats entry pubsub topic: %s", pubsubTopic)

	config := common.GooglePubsubConfig{
		ProjectId:          googleProjectId,
		Topic:              pubsubTopic,
		MessageChannelSize: 10 * 1024,
	}

	statsPubsubProducer, err := common.CreateGooglePubsubProducer(service.Context, config)
	if err != nil {
		core.Error("could not create google pubsub producer for processing route matrix: %v", err)
		os.Exit(1)
	}

	maxBytes := envvar.GetInt("ROUTE_MATRIX_STATS_MESSAGE_MAX_BYTES", 1024)

	core.Log("route matrix stats message max bytes: %d", maxBytes)

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

				routeMatrixStatsEntry := messages.RouteMatrixStatsMessage{}

				routeMatrixStatsEntry.Version = messages.RouteMatrixStatsMessageVersion
				routeMatrixStatsEntry.Bytes = routeMatrixBytes
				routeMatrixStatsEntry.NumRelays = routeMatrixNumRelays
				routeMatrixStatsEntry.NumDestRelays = routeMatrixNumDestRelays
				routeMatrixStatsEntry.NumDatacenters = routeMatrixNumDatacenters
				routeMatrixStatsEntry.TotalRoutes = analysis.TotalRoutes
				routeMatrixStatsEntry.AverageNumRoutes = analysis.AverageNumRoutes
				routeMatrixStatsEntry.AverageRouteLength = analysis.AverageRouteLength
				routeMatrixStatsEntry.NoRoutePercent = analysis.NoRoutePercent
				routeMatrixStatsEntry.OneRoutePercent = analysis.OneRoutePercent
				routeMatrixStatsEntry.NoDirectRoutePercent = analysis.NoDirectRoutePercent
				routeMatrixStatsEntry.RTTBucket_NoImprovement = analysis.RTTBucket_NoImprovement
				routeMatrixStatsEntry.RTTBucket_0_5ms = analysis.RTTBucket_0_5ms
				routeMatrixStatsEntry.RTTBucket_5_10ms = analysis.RTTBucket_5_10ms
				routeMatrixStatsEntry.RTTBucket_10_15ms = analysis.RTTBucket_10_15ms
				routeMatrixStatsEntry.RTTBucket_15_20ms = analysis.RTTBucket_15_20ms
				routeMatrixStatsEntry.RTTBucket_20_25ms = analysis.RTTBucket_20_25ms
				routeMatrixStatsEntry.RTTBucket_25_30ms = analysis.RTTBucket_25_30ms
				routeMatrixStatsEntry.RTTBucket_30_35ms = analysis.RTTBucket_30_35ms
				routeMatrixStatsEntry.RTTBucket_35_40ms = analysis.RTTBucket_35_40ms
				routeMatrixStatsEntry.RTTBucket_40_45ms = analysis.RTTBucket_40_45ms
				routeMatrixStatsEntry.RTTBucket_45_50ms = analysis.RTTBucket_45_50ms
				routeMatrixStatsEntry.RTTBucket_50ms_Plus = analysis.RTTBucket_50ms_Plus

				message := routeMatrixStatsEntry.Write(make([]byte, routeMatrixBytes))

				statsPubsubProducer.MessageChannel <- message
			}
		}
	}()
}

// --------------------------------------------------------------------
