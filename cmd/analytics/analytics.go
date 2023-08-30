package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/messages"
)

var costMatrixURL string
var routeMatrixURL string
var costMatrixInterval time.Duration
var routeMatrixInterval time.Duration
var googleProjectId string
var bigqueryDataset string
var enableGooglePubsub bool
var enableGoogleBigquery bool

var logMutex sync.Mutex

func main() {

	service := common.CreateService("analytics")

	costMatrixURL = envvar.GetString("COST_MATRIX_URL", "http://127.0.0.1:30001/cost_matrix")
	routeMatrixURL = envvar.GetString("ROUTE_MATRIX_URL", "http://127.0.0.1:30001/route_matrix")
	costMatrixInterval = envvar.GetDuration("COST_MATRIX_INTERVAL", time.Second)
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)
	googleProjectId = envvar.GetString("GOOGLE_PROJECT_ID", "")
	bigqueryDataset = envvar.GetString("BIGQUERY_DATASET", service.Env)
	enableGooglePubsub = envvar.GetBool("ENABLE_GOOGLE_PUBSUB", false)
	enableGoogleBigquery = envvar.GetBool("ENABLE_GOOGLE_BIGQUERY", false)

	core.Debug("cost matrix url: %s", costMatrixURL)
	core.Debug("route matrix url: %s", routeMatrixURL)
	core.Debug("cost matrix interval: %s", costMatrixInterval)
	core.Debug("route matrix interval: %s", routeMatrixInterval)
	core.Debug("google project id: %s", googleProjectId)
	core.Debug("bigquery dataset: %s", bigqueryDataset)
	core.Debug("enable google pubsub: %v", enableGooglePubsub)
	core.Debug("enable google bigquery: %v", enableGoogleBigquery)

	ProcessCostMatrix(service)

	ProcessRouteMatrix(service)

	if enableGooglePubsub && enableGoogleBigquery {

		important := envvar.GetBool("GOOGLE_PUBSUB_IMPORTANT", true)

		Process[*messages.AnalyticsCostMatrixUpdateMessage](service, "cost_matrix_update", &messages.AnalyticsCostMatrixUpdateMessage{}, important)
		Process[*messages.AnalyticsRouteMatrixUpdateMessage](service, "route_matrix_update", &messages.AnalyticsRouteMatrixUpdateMessage{}, important)
		Process[*messages.AnalyticsRelayToRelayPingMessage](service, "relay_to_relay_ping", &messages.AnalyticsRelayToRelayPingMessage{}, important)
		Process[*messages.AnalyticsRelayUpdateMessage](service, "relay_update", &messages.AnalyticsRelayUpdateMessage{}, important)
		Process[*messages.AnalyticsServerInitMessage](service, "server_init", &messages.AnalyticsServerInitMessage{}, important)
		Process[*messages.AnalyticsServerUpdateMessage](service, "server_update", &messages.AnalyticsServerUpdateMessage{}, important)
		Process[*messages.AnalyticsSessionUpdateMessage](service, "session_update", &messages.AnalyticsSessionUpdateMessage{}, important)
		Process[*messages.AnalyticsSessionSummaryMessage](service, "session_summary", &messages.AnalyticsSessionSummaryMessage{}, important)
		Process[*messages.AnalyticsNearRelayUpdateMessage](service, "near_relay_update", &messages.AnalyticsNearRelayUpdateMessage{}, important)
		Process[*messages.AnalyticsMatchDataMessage](service, "match_data", &messages.AnalyticsMatchDataMessage{}, important)
	}

	service.StartWebServer()

	service.LeaderElection()

	service.WaitForShutdown()
}

// --------------------------------------------------------------------

func Process[T messages.BigQueryMessage](service *common.Service, name string, message T, important bool) {

	namePrefix := strings.ToUpper(name) + "_"

	pubsubTopic := envvar.GetString(namePrefix+"PUBSUB_TOPIC", name)
	pubsubSubscription := envvar.GetString(namePrefix+"PUBSUB_SUBSCRIPTION", name)
	bigqueryTable := envvar.GetString(namePrefix+"BIGQUERY_TABLE", name)

	core.Debug("processing %s messages: topic = '%s', subscription = '%s', bigquery table = '%s'", name, pubsubTopic, pubsubSubscription, bigqueryTable)

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
						core.Error("could not read %s message. not important, so dropping", name)
						pubsubMessage.Ack()
						break
					}

					core.Error("could not read %s message, important, so not acking it.", name)
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

	httpClient := &http.Client{
		Timeout: costMatrixInterval,
	}

	var googlePubsubProducer *common.GooglePubsubProducer
	if enableGooglePubsub {
		pubsubTopic := envvar.GetString("COST_MATRIX_STATS_PUBSUB_TOPIC", "cost_matrix_stats")
		core.Debug("cost matrix stats google pubsub topic: %s", pubsubTopic)
		config := common.GooglePubsubConfig{
			ProjectId:          googleProjectId,
			Topic:              pubsubTopic,
			MessageChannelSize: 10 * 1024,
		}
		var err error
		googlePubsubProducer, err = common.CreateGooglePubsubProducer(service.Context, config)
		if err != nil {
			core.Error("could not create google pubsub producer for cost matrix update: %v", err)
			os.Exit(1)
		}
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

				response, err := httpClient.Get(costMatrixURL)
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

				costMatrixSize := len(buffer)
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
				core.Debug("cost matrix size: %d", costMatrixSize)
				core.Debug("cost matrix num relays: %d", costMatrixNumRelays)
				core.Debug("cost matrix num dest relays: %d", costMatrixNumDestRelays)
				core.Debug("cost matrix num datacenters: %d", costMatrixNumDatacenters)
				core.Debug("---------------------------------------------")

				logMutex.Unlock()

				// send cost matrix update message via google pubsub

				message := messages.AnalyticsCostMatrixUpdateMessage{}

				message.Version = messages.AnalyticsCostMatrixUpdateMessageVersion_Write
				message.Timestamp = uint64(time.Now().Unix())
				message.CostMatrixSize = costMatrixSize
				message.NumRelays = costMatrixNumRelays
				message.NumDestRelays = costMatrixNumDestRelays
				message.NumDatacenters = costMatrixNumDatacenters

				messageData := message.Write(make([]byte, message.GetMaxSize()))

				if enableGooglePubsub {
					googlePubsubProducer.MessageChannel <- messageData
				}
			}
		}
	}()
}

func ProcessRouteMatrix(service *common.Service) {

	var googlePubsubProducer *common.GooglePubsubProducer
	if enableGooglePubsub {
		pubsubTopic := envvar.GetString("ROUTE_MATRIX_UPDATE_PUBSUB_TOPIC", "route_matrix_update")
		core.Debug("route matrix stats google pubsub topic: %s", pubsubTopic)
		config := common.GooglePubsubConfig{
			ProjectId:          googleProjectId,
			Topic:              pubsubTopic,
			MessageChannelSize: 10 * 1024,
		}
		var err error
		googlePubsubProducer, err = common.CreateGooglePubsubProducer(service.Context, config)
		if err != nil {
			core.Error("could not create google pubsub producer for route matrix update: %v", err)
			os.Exit(1)
		}
	}

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

				response, err := httpClient.Get(routeMatrixURL)
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

				routeMatrixSize := len(buffer)
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

				analysis := routeMatrix.Analyze()

				routeMatrixNumFullRelays := 0

				core.Debug("---------------------------------------------")

				core.Debug("route matrix size: %d", routeMatrixSize)

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

				// send route matrix stats via google pubsub pubsub

				message := messages.AnalyticsRouteMatrixUpdateMessage{}

				message.Version = messages.AnalyticsRouteMatrixUpdateMessageVersion_Write
				message.Timestamp = uint64(time.Now().Unix())
				message.RouteMatrixSize = routeMatrixSize
				message.NumRelays = routeMatrixNumRelays
				message.NumDestRelays = routeMatrixNumDestRelays
				message.NumDatacenters = routeMatrixNumDatacenters
				message.TotalRoutes = analysis.TotalRoutes
				message.AverageNumRoutes = analysis.AverageNumRoutes
				message.AverageRouteLength = analysis.AverageRouteLength
				message.NoRoutePercent = analysis.NoRoutePercent
				message.OneRoutePercent = analysis.OneRoutePercent
				message.NoDirectRoutePercent = analysis.NoDirectRoutePercent
				message.RTTBucket_NoImprovement = analysis.RTTBucket_NoImprovement
				message.RTTBucket_0_5ms = analysis.RTTBucket_0_5ms
				message.RTTBucket_5_10ms = analysis.RTTBucket_5_10ms
				message.RTTBucket_10_15ms = analysis.RTTBucket_10_15ms
				message.RTTBucket_15_20ms = analysis.RTTBucket_15_20ms
				message.RTTBucket_20_25ms = analysis.RTTBucket_20_25ms
				message.RTTBucket_25_30ms = analysis.RTTBucket_25_30ms
				message.RTTBucket_30_35ms = analysis.RTTBucket_30_35ms
				message.RTTBucket_35_40ms = analysis.RTTBucket_35_40ms
				message.RTTBucket_40_45ms = analysis.RTTBucket_40_45ms
				message.RTTBucket_45_50ms = analysis.RTTBucket_45_50ms
				message.RTTBucket_50ms_Plus = analysis.RTTBucket_50ms_Plus

				messageData := message.Write(make([]byte, message.GetMaxSize()))

				if enableGooglePubsub {
					googlePubsubProducer.MessageChannel <- messageData
				}
			}
		}
	}()
}

// --------------------------------------------------------------------
