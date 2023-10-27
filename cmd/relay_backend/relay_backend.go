package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
	"io/ioutil"

	"github.com/gorilla/mux"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/packets"
	"github.com/networknext/next/modules/messages"
	"github.com/networknext/next/modules/portal"

	"github.com/redis/go-redis/v9"
)

var maxJitter int32
var maxPacketLoss float32
var routeMatrixInterval time.Duration

var redisHostName string

var analyticsRelayToRelayPingGooglePubsubTopic string
var analyticsRelayToRelayPingGooglePubsubChannelSize int

var analyticsRelayUpdateGooglePubsubTopic string
var analyticsRelayUpdateGooglePubsubChannelSize int

var enableGooglePubsub bool

var initialDelay int

var relaysMutex sync.RWMutex
var relaysCSVData []byte

var costMatrixMutex sync.RWMutex
var costMatrixData []byte

var routeMatrixMutex sync.RWMutex
var routeMatrixData []byte

var delayMutex sync.RWMutex
var delayCompleted bool

var startTime time.Time

var counterNames [constants.NumRelayCounters]string

var enableRedisTimeSeries bool
var redisTimeSeriesCluster []string
var redisTimeSeriesHostname string

var timeSeriesPublisher *common.RedisTimeSeriesPublisher

var redisPortalHostname string
var redisPortalCluster []string

var relayInserterBatchSize int
var relayInserter *portal.RelayInserter
var countersPublisher *common.RedisCountersPublisher
var analyticsRelayUpdateProducer *common.GooglePubsubProducer
var analyticsRelayToRelayPingProducer *common.GooglePubsubProducer

var postRelayUpdateRequestChannel chan *packets.RelayUpdateRequestPacket

func main() {

	service := common.CreateService("relay_backend")

	maxJitter = int32(envvar.GetInt("MAX_JITTER", 1000))
	maxPacketLoss = float32(envvar.GetFloat("MAX_PACKET_LOSS", 100.0))
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)

	redisHostName = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	analyticsRelayToRelayPingGooglePubsubTopic = envvar.GetString("ANALYTICS_RELAY_TO_RELAY_PING_GOOGLE_PUBSUB_TOPIC", "relay_to_relay_ping")
	analyticsRelayToRelayPingGooglePubsubChannelSize = envvar.GetInt("ANALYTICS_RELAY_TO_RELAY_PING_GOOGLE_PUBSUB_CHANNEL_SIZE", 10*1024)

	analyticsRelayUpdateGooglePubsubTopic = envvar.GetString("ANALYTICS_RELAY_UPDATE_GOOGLE_PUBSUB_TOPIC", "relay_update")
	analyticsRelayUpdateGooglePubsubChannelSize = envvar.GetInt("ANALYTICS_RELAY_UPDATE_GOOGLE_PUBSUB_CHANNEL_SIZE", 10*1024)

	enableGooglePubsub = envvar.GetBool("ENABLE_GOOGLE_PUBSUB", false)

	initialDelay = envvar.GetInt("INITIAL_DELAY", 15)

	relayInserterBatchSize = envvar.GetInt("RELAY_INSERTER_BATCH_SIZE", 1024)

	startTime = time.Now()

	enableRedisTimeSeries = envvar.GetBool("ENABLE_REDIS_TIME_SERIES", false)
	redisTimeSeriesCluster = envvar.GetStringArray("REDIS_TIME_SERIES_CLUSTER", []string{})
	redisTimeSeriesHostname = envvar.GetString("REDIS_TIME_SERIES_HOSTNAME", "127.0.0.1:6379")

	if enableRedisTimeSeries {
		core.Debug("redis time series cluster: %s", redisTimeSeriesCluster)
		core.Debug("redis time series hostname: %s", redisTimeSeriesHostname)
	}

	redisPortalCluster = envvar.GetStringArray("REDIS_PORTAL_CLUSTER", []string{})
	redisPortalHostname = envvar.GetString("REDIS_PORTAL_HOSTNAME", "127.0.0.1:6379")

	core.Debug("max jitter: %d", maxJitter)
	core.Debug("max packet loss: %.1f", maxPacketLoss)
	core.Debug("route matrix interval: %s", routeMatrixInterval)
	core.Debug("redis host name: %s", redisHostName)

	core.Debug("analytics relay to relay ping google pubsub topic: %s", analyticsRelayToRelayPingGooglePubsubTopic)
	core.Debug("analytics relay to relay ping google pubsub channel size: %d", analyticsRelayToRelayPingGooglePubsubChannelSize)

	core.Debug("analytics relay update google pubsub topic: %s", analyticsRelayUpdateGooglePubsubTopic)
	core.Debug("analytics relay update google pubsub channel size: %d", analyticsRelayUpdateGooglePubsubChannelSize)

	core.Debug("enable google pubsub: %v", enableGooglePubsub)

	core.Debug("initial delay: %d", initialDelay)

	core.Debug("start time: %s", startTime.String())

	var redisClient redis.Cmdable
	if len(redisPortalCluster) > 0 {
		redisClient = common.CreateRedisClusterClient(redisPortalCluster)
	} else {
		redisClient = common.CreateRedisClient(redisPortalHostname)
	}

	relayInserter = portal.CreateRelayInserter(redisClient, relayInserterBatchSize)

	if enableRedisTimeSeries {

		timeSeriesConfig := common.RedisTimeSeriesConfig{
			RedisHostname: redisTimeSeriesHostname,
			RedisCluster:  redisTimeSeriesCluster,
		}
		var err error
		timeSeriesPublisher, err = common.CreateRedisTimeSeriesPublisher(service.Context, timeSeriesConfig)
		if err != nil {
			core.Error("could not create redis time series publisher: %v", err)
			os.Exit(1)
		}

		countersConfig := common.RedisCountersConfig{
			RedisHostname: redisTimeSeriesHostname,
			RedisCluster:  redisTimeSeriesCluster,
		}
		countersPublisher, err = common.CreateRedisCountersPublisher(service.Context, countersConfig)
		if err != nil {
			core.Error("could not create redis counters publisher: %v", err)
			os.Exit(1)
		}
	}

	if enableGooglePubsub {

		var err error
		analyticsRelayUpdateProducer, err = common.CreateGooglePubsubProducer(service.Context, common.GooglePubsubConfig{
			ProjectId:          service.GoogleProjectId,
			Topic:              analyticsRelayUpdateGooglePubsubTopic,
			MessageChannelSize: analyticsRelayUpdateGooglePubsubChannelSize,
		})
		if err != nil {
			core.Error("could not create analytics relay update google pubsub producer")
			os.Exit(1)
		}

		analyticsRelayToRelayPingProducer, err = common.CreateGooglePubsubProducer(service.Context, common.GooglePubsubConfig{
			ProjectId:          service.GoogleProjectId,
			Topic:              analyticsRelayToRelayPingGooglePubsubTopic,
			MessageChannelSize: analyticsRelayToRelayPingGooglePubsubChannelSize,
		})
		if err != nil {
			core.Error("could not create analytics relay to relay ping google pubsub producer")
			os.Exit(1)
		}
	}

	postRelayUpdateRequestChannel = make(chan *packets.RelayUpdateRequestPacket, 1024*1024) // todo: make configurable

	service.LoadDatabase()

	initCounterNames()

	relayManager := common.CreateRelayManager(service.Local)

	service.Router.HandleFunc("/relay_update", relayUpdateHandler(service, relayManager)).Methods("POST")
	service.Router.HandleFunc("/health_fanout", healthFanoutHandler)
	service.Router.HandleFunc("/relays", relaysHandler)
	service.Router.HandleFunc("/relay_data", relayDataHandler(service))
	service.Router.HandleFunc("/cost_matrix", costMatrixHandler)
	service.Router.HandleFunc("/route_matrix", routeMatrixHandler)
	service.Router.HandleFunc("/relay_counters/{relay_name}", relayCountersHandler(service, relayManager))
	service.Router.HandleFunc("/cost_matrix_html", costMatrixHtmlHandler(service, relayManager))
	service.Router.HandleFunc("/routes/{src}/{dest}", routesHandler(service, relayManager))
	service.Router.HandleFunc("/relay_manager", relayManagerHandler(service, relayManager))
	service.Router.HandleFunc("/costs", costsHandler(service, relayManager))
	service.Router.HandleFunc("/active_relays", activeRelaysHandler(service, relayManager))

	service.SetHealthFunctions(sendTrafficToMe(service), machineIsHealthy, ready(service))

	service.StartWebServer()

	service.LeaderElection(initialDelay)

	UpdateRouteMatrix(service, relayManager)

	UpdateInitialDelayState(service)

	PostRelayUpdateRequest(service)

	service.WaitForShutdown()
}

func relayUpdateHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		startTime := time.Now()

		defer func() {
			duration := time.Since(startTime)
			if duration.Milliseconds() > 1000 {
				core.Warn("long relay update: %s", duration.String())
			}
		}()

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			core.Error("could not read request body: %v", err)
			return
		}
		defer r.Body.Close()

		// discard if the body is too small to possibly be valid

		if len(body) < 64 {
			core.Error("relay update is too small to be valid")
			return
		}

		// read the relay update request packet

		var relayUpdateRequest packets.RelayUpdateRequestPacket
		err = relayUpdateRequest.Read(body)
		if err != nil {
			core.Error("could not read relay update: %v", err)
			return
		}

		go func() {

			// check if we are overloaded

			currentTime := uint64(time.Now().Unix())

			if relayUpdateRequest.CurrentTime < currentTime - 5 {
				core.Error("relay update is old. relay gateway -> relay backend is overloaded!")
			}

			// look up the relay in the database

			relayData := service.RelayData()

			relayId := common.RelayId(relayUpdateRequest.Address.String())
			relayIndex, ok := relayData.RelayIdToIndex[relayId]
			if !ok {
				core.Error("unknown relay id %016x", relayId)
				return
			}

			relayName := relayData.RelayNames[relayIndex]
			relayAddress := relayData.RelayAddresses[relayIndex]

			// process samples in the relay update (this drives the cost matrix...)

			core.Debug("[%s] received update for %s [%016x]", relayAddress.String(), relayName, relayId)

			numSamples := int(relayUpdateRequest.NumSamples)

			relayManager.ProcessRelayUpdate(int64(currentTime),
				relayId,
				relayName,
				relayUpdateRequest.Address,
				int(relayUpdateRequest.SessionCount),
				relayUpdateRequest.RelayVersion,
				relayUpdateRequest.RelayFlags,
				numSamples,
				relayUpdateRequest.SampleRelayId[:numSamples],
				relayUpdateRequest.SampleRTT[:numSamples],
				relayUpdateRequest.SampleJitter[:numSamples],
				relayUpdateRequest.SamplePacketLoss[:numSamples],
				relayUpdateRequest.RelayCounters[:],
			)

			postRelayUpdateRequestChannel <- &relayUpdateRequest
		}()
	}
}

func PostRelayUpdateRequest(service *common.Service) {

	for {
		relayUpdateRequest := <- postRelayUpdateRequestChannel

		// build relay to relay ping messages for analytics

		numRoutable := 0

		// look up the relay in the database

		relayData := service.RelayData()

		relayId := common.RelayId(relayUpdateRequest.Address.String())
		relayIndex, ok := relayData.RelayIdToIndex[relayId]
		if !ok {
			continue
		}

		relayName := relayData.RelayNames[relayIndex]
		relayAddress := relayData.RelayAddresses[relayIndex]

		numSamples := int(relayUpdateRequest.NumSamples)

		pingMessages := make([]messages.AnalyticsRelayToRelayPingMessage, numSamples)

		for i := 0; i < numSamples; i++ {

			rtt := relayUpdateRequest.SampleRTT[i]
			jitter := relayUpdateRequest.SampleJitter[i]
			pl := float32(relayUpdateRequest.SamplePacketLoss[i] / 65535.0 * 100.0)

			if rtt < 255 && int32(jitter) <= maxJitter && pl <= maxPacketLoss {
				numRoutable++
			}

			sampleRelayId := relayUpdateRequest.SampleRelayId[i]

			pingMessages[i] = messages.AnalyticsRelayToRelayPingMessage{
				Version:            messages.AnalyticsRelayToRelayPingMessageVersion_Write,
				Timestamp:          uint64(time.Now().Unix()),
				SourceRelayId:      relayId,
				DestinationRelayId: sampleRelayId,
				RTT:                rtt,
				Jitter:             jitter,
				PacketLoss:         pl,
			}
		}

		numUnroutable := numSamples - numRoutable

		// send relay update message to portal
		{
			message := messages.PortalRelayUpdateMessage{
				Version:                   messages.PortalRelayUpdateMessageVersion_Write,
				Timestamp:                 uint64(time.Now().Unix()),
				RelayName:                 relayName,
				RelayId:                   relayId,
				SessionCount:              relayUpdateRequest.SessionCount,
				MaxSessions:               uint32(relayData.RelayArray[relayIndex].MaxSessions),
				EnvelopeBandwidthUpKbps:   relayUpdateRequest.EnvelopeBandwidthUpKbps,
				EnvelopeBandwidthDownKbps: relayUpdateRequest.EnvelopeBandwidthDownKbps,
				PacketsSentPerSecond:      relayUpdateRequest.PacketsSentPerSecond,
				PacketsReceivedPerSecond:  relayUpdateRequest.PacketsReceivedPerSecond,
				BandwidthSentKbps:         relayUpdateRequest.BandwidthSentKbps,
				BandwidthReceivedKbps:     relayUpdateRequest.BandwidthReceivedKbps,
				NearPingsPerSecond:        relayUpdateRequest.NearPingsPerSecond,
				RelayPingsPerSecond:       relayUpdateRequest.RelayPingsPerSecond,
				RelayFlags:                relayUpdateRequest.RelayFlags,
				RelayVersion:              relayUpdateRequest.RelayVersion,
				NumRoutable:               uint32(numRoutable),
				NumUnroutable:             uint32(numUnroutable),
				StartTime:                 relayUpdateRequest.StartTime,
				CurrentTime:               relayUpdateRequest.CurrentTime,
				RelayAddress:              relayAddress,
			}

			if service.IsLeader() {

				relayData := portal.RelayData{
					RelayId:      message.RelayId,
					RelayName:    message.RelayName,
					RelayAddress: message.RelayAddress.String(),
					NumSessions:  message.SessionCount,
					MaxSessions:  message.MaxSessions,
					StartTime:    message.StartTime,
					RelayFlags:   message.RelayFlags,
					RelayVersion: message.RelayVersion,
				}

				relayInserter.Insert(service.Context, &relayData)

				if enableRedisTimeSeries {

					// send time series to redis

					timeSeriesMessage := common.RedisTimeSeriesMessage{}

					timeSeriesMessage.Timestamp = uint64(time.Now().UnixNano() / 1000000)

					timeSeriesMessage.Keys = []string{
						fmt.Sprintf("relay_%016x_session_count", message.RelayId),
						fmt.Sprintf("relay_%016x_envelope_bandwidth_up_kbps", message.RelayId),
						fmt.Sprintf("relay_%016x_envelope_bandwidth_down_kbps", message.RelayId),
						fmt.Sprintf("relay_%016x_packets_sent_per_second", message.RelayId),
						fmt.Sprintf("relay_%016x_packets_received_per_second", message.RelayId),
						fmt.Sprintf("relay_%016x_bandwidth_sent_kbps", message.RelayId),
						fmt.Sprintf("relay_%016x_bandwidth_received_kbps", message.RelayId),
						fmt.Sprintf("relay_%016x_near_pings_per_second", message.RelayId),
						fmt.Sprintf("relay_%016x_relay_pings_per_second", message.RelayId),
						fmt.Sprintf("relay_%016x_num_routable", message.RelayId),
						fmt.Sprintf("relay_%016x_num_unroutable", message.RelayId),
					}

					timeSeriesMessage.Values = []float64{
						float64(message.SessionCount),
						float64(message.EnvelopeBandwidthUpKbps),
						float64(message.EnvelopeBandwidthDownKbps),
						float64(message.PacketsSentPerSecond),
						float64(message.PacketsReceivedPerSecond),
						float64(message.BandwidthSentKbps),
						float64(message.BandwidthReceivedKbps),
						float64(message.NearPingsPerSecond),
						float64(message.RelayPingsPerSecond),
						float64(message.NumRoutable),
						float64(message.NumUnroutable),
					}

					timeSeriesPublisher.MessageChannel <- &timeSeriesMessage

					// send counters to redis

					countersPublisher.MessageChannel <- "relay_update"
				}
			}
		}

		// send relay update message to analytics
		{
			message := messages.AnalyticsRelayUpdateMessage{
				Version:                   messages.AnalyticsRelayUpdateMessageVersion_Write,
				Timestamp:                 uint64(time.Now().Unix()),
				RelayId:                   relayId,
				SessionCount:              relayUpdateRequest.SessionCount,
				MaxSessions:               uint32(relayData.RelayArray[relayIndex].MaxSessions),
				EnvelopeBandwidthUpKbps:   relayUpdateRequest.EnvelopeBandwidthUpKbps,
				EnvelopeBandwidthDownKbps: relayUpdateRequest.EnvelopeBandwidthDownKbps,
				PacketsSentPerSecond:      relayUpdateRequest.PacketsSentPerSecond,
				PacketsReceivedPerSecond:  relayUpdateRequest.PacketsReceivedPerSecond,
				BandwidthSentKbps:         relayUpdateRequest.BandwidthSentKbps,
				BandwidthReceivedKbps:     relayUpdateRequest.BandwidthReceivedKbps,
				NearPingsPerSecond:        relayUpdateRequest.NearPingsPerSecond,
				RelayPingsPerSecond:       relayUpdateRequest.RelayPingsPerSecond,
				RelayFlags:                relayUpdateRequest.RelayFlags,
				NumRelayCounters:          relayUpdateRequest.NumRelayCounters,
				RelayCounters:             relayUpdateRequest.RelayCounters,
				NumRoutable:               uint32(numRoutable),
				NumUnroutable:             uint32(numUnroutable),
				StartTime:                 relayUpdateRequest.StartTime,
				CurrentTime:               relayUpdateRequest.CurrentTime,
			}

			if service.IsLeader() {
				messageBuffer := make([]byte, message.GetMaxSize())
				messageData := message.Write(messageBuffer[:])
				if enableGooglePubsub {
					analyticsRelayUpdateProducer.MessageChannel <- messageData
				}
			}
		}

		// send relay to relay ping messages to analytics

		if service.IsLeader() {
			for i := 0; i < len(pingMessages); i++ {
				messageBuffer := make([]byte, pingMessages[i].GetMaxSize())
				messageData := pingMessages[i].Write(messageBuffer[:])
				if enableGooglePubsub {
					analyticsRelayToRelayPingProducer.MessageChannel <- messageData
				}
			}
		}
	}
}

func healthFanoutHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func activeRelaysHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		activeRelays := relayManager.GetActiveRelays(time.Now().Unix())
		for i := range activeRelays {
			fmt.Fprintf(w, "%s, ", activeRelays[i].Name)
		}
		fmt.Fprintf(w, "\n")
	}
}

func costsHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		costMatrixMutex.RLock()
		data := costMatrixData
		costMatrixMutex.RUnlock()
		costMatrix := common.CostMatrix{}
		err := costMatrix.Read(data)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "no cost matrix: %v\n", err)
			return
		}
		activeRelayMap := relayManager.GetActiveRelayMap(time.Now().Unix())
		for i := range costMatrix.RelayNames {
			if _, exists := activeRelayMap[costMatrix.RelayIds[i]]; !exists {
				continue
			}
			fmt.Fprintf(w, "%s: ", costMatrix.RelayNames[i])
			for j := range costMatrix.RelayNames {
				if _, exists := activeRelayMap[costMatrix.RelayIds[j]]; !exists {
					continue
				}
				if i == j {
					continue
				}
				index := core.TriMatrixIndex(i, j)
				cost := costMatrix.Costs[index]
				fmt.Fprintf(w, "%d,", cost)
			}
			fmt.Fprintf(w, "\n")
		}
	}
}

func relayManagerHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		copy := relayManager.Copy()
		var buffer bytes.Buffer
		err := gob.NewEncoder(&buffer).Encode(copy)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "no relay manager: %v\n", err)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, err = buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func routesHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		src := vars["src"]
		dest := vars["dest"]
		routeMatrixMutex.RLock()
		data := routeMatrixData
		routeMatrixMutex.RUnlock()
		routeMatrix := common.RouteMatrix{}
		err := routeMatrix.Read(data)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "no route matrix\n")
			return
		}
		src_index := -1
		for i := range routeMatrix.RelayNames {
			if routeMatrix.RelayNames[i] == src {
				src_index = i
				break
			}
		}
		dest_index := -1
		for i := range routeMatrix.RelayNames {
			if routeMatrix.RelayNames[i] == dest {
				dest_index = i
				break
			}
		}
		if src_index == -1 || dest_index == -1 || src_index == dest_index {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		const htmlHeader = `<!DOCTYPE html>
		<html lang="en">
		<head>
		  <meta charset="utf-8">
		  <title>Routes</title>
		  <style>
			table, th, td {
		      border: 1px solid black;
		      border-collapse: collapse;
		      text-align: center;
		      padding: 10px;
		    }
			*{
		    font-family:Courier;
		  }	  
		  </style>
		</head>
		<body>`
		fmt.Fprintf(w, "%s\n", htmlHeader)
		fmt.Fprintf(w, "route matrix: %s - %s<br><br>\n", src, dest)
		fmt.Fprintf(w, "<table>\n")
		fmt.Fprintf(w, "<tr><td><b>Route Cost</b></td><td><b>Route Hash</b></td><td><b>Route Relays</b></td></tr>\n")
		index := core.TriMatrixIndex(src_index, dest_index)
		entry := routeMatrix.RouteEntries[index]
		for i := 0; i < int(entry.NumRoutes); i++ {
			routeRelays := ""
			numRouteRelays := int(entry.RouteNumRelays[i])
			for j := 0; j < numRouteRelays; j++ {
				routeRelayIndex := entry.RouteRelays[i][j]
				routeRelayName := routeMatrix.RelayNames[routeRelayIndex]
				routeRelays += routeRelayName
				if j != numRouteRelays-1 {
					routeRelays += " - "
				}
			}
			fmt.Fprintf(w, "<tr><td>%d</td><td>%0x</td><td>%s</td></tr>", entry.RouteCost[i], entry.RouteHash[i], routeRelays)
		}
		fmt.Fprintf(w, "<tr><td>%d</td><td></td><td>%s</td></tr>", entry.DirectCost, "direct")
		fmt.Fprintf(w, "</table>\n")
		const htmlFooter = `</body></html>`
		fmt.Fprintf(w, "%s\n", htmlFooter)
	}
}

func costMatrixHtmlHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		costMatrixMutex.RLock()
		data := costMatrixData
		costMatrixMutex.RUnlock()
		costMatrix := common.CostMatrix{}
		err := costMatrix.Read(data)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "no cost matrix: %v\n", err)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		const htmlHeader = `<!DOCTYPE html>
		<html lang="en">
		<head>
		  <meta charset="utf-8">
		  <title>Cost Matrix</title>
		  <style>
			table, th, td {
		      border: 1px solid black;
		      border-collapse: collapse;
		      text-align: center;
		      padding: 10px;
		    }
			cost{
         	  color: white;
      		}
			*{
		    font-family:Courier;
		    }	  
		  </style>
		</head>
		<body>`
		fmt.Fprintf(w, "%s\n", htmlHeader)
		fmt.Fprintf(w, "cost matrix:<br><br><table>\n")
		fmt.Fprintf(w, "<tr><td></td>")
		for i := range costMatrix.RelayNames {
			fmt.Fprintf(w, "<td><b>%s</b></td>", costMatrix.RelayNames[i])
		}
		fmt.Fprintf(w, "</tr>\n")
		for i := range costMatrix.RelayNames {
			fmt.Fprintf(w, "<tr><td><b>%s</b></td>", costMatrix.RelayNames[i])
			for j := range costMatrix.RelayNames {
				if i == j {
					fmt.Fprint(w, "<td bgcolor=\"lightgrey\"></td>")
					continue
				}
				nope := false
				costString := ""
				index := core.TriMatrixIndex(i, j)
				cost := costMatrix.Costs[index]
				if cost >= 0 && cost < 255 {
					costString = fmt.Sprintf("%d", cost)
				} else {
					nope = true
				}
				clickable := fmt.Sprintf("class=\"clickable\" onclick=\"window.location='/routes/%s/%s'\"", costMatrix.RelayNames[i], costMatrix.RelayNames[j])

				if nope {
					fmt.Fprintf(w, "<td %s bgcolor=\"red\"></td>", clickable)
				} else {
					fmt.Fprintf(w, "<td %s bgcolor=\"green\"><cost>%s</cost></td>", clickable, costString)
				}
			}
			fmt.Fprintf(w, "</tr>\n")
		}
		fmt.Fprintf(w, "</table>\n")
		const htmlFooter = `</body></html>`
		fmt.Fprintf(w, "%s\n", htmlFooter)
	}
}

func initCounterNames() {
	// awk '/^#define RELAY_COUNTER_/ {print "    counterNames["$3"] = \""$2"\""}' ./relay/relay.cpp
	counterNames[0] = "RELAY_COUNTER_PACKETS_SENT"
	counterNames[1] = "RELAY_COUNTER_PACKETS_RECEIVED"
	counterNames[2] = "RELAY_COUNTER_BYTES_SENT"
	counterNames[3] = "RELAY_COUNTER_BYTES_RECEIVED"
	counterNames[4] = "RELAY_COUNTER_BASIC_PACKET_FILTER_DROPPED_PACKET"
	counterNames[5] = "RELAY_COUNTER_ADVANCED_PACKET_FILTER_DROPPED_PACKET"
	counterNames[6] = "RELAY_COUNTER_SESSION_CREATED"
	counterNames[7] = "RELAY_COUNTER_SESSION_CONTINUED"
	counterNames[8] = "RELAY_COUNTER_SESSION_DESTROYED"
	counterNames[9] = "RELAY_COUNTER_SESSION_EXPIRED"
	counterNames[10] = "RELAY_COUNTER_RELAY_PING_PACKET_SENT"
	counterNames[11] = "RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED"
	counterNames[12] = "RELAY_COUNTER_RELAY_PING_PACKET_DID_NOT_VERIFY"
	counterNames[13] = "RELAY_COUNTER_RELAY_PING_PACKET_EXPIRED"
	counterNames[14] = "RELAY_COUNTER_RELAY_PING_PACKET_WRONG_SIZE"
	counterNames[15] = "RELAY_COUNTER_RELAY_PONG_PACKET_SENT"
	counterNames[16] = "RELAY_COUNTER_RELAY_PONG_PACKET_RECEIVED"
	counterNames[17] = "RELAY_COUNTER_RELAY_PONG_PACKET_WRONG_SIZE"
	counterNames[20] = "RELAY_COUNTER_NEAR_PING_PACKET_RECEIVED"
	counterNames[21] = "RELAY_COUNTER_NEAR_PING_PACKET_WRONG_SIZE"
	counterNames[22] = "RELAY_COUNTER_NEAR_PING_PACKET_RESPONDED_WITH_PONG"
	counterNames[23] = "RELAY_COUNTER_NEAR_PING_PACKET_DID_NOT_VERIFY"
	counterNames[24] = "RELAY_COUNTER_NEAR_PING_PACKET_EXPIRED"
	counterNames[30] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED"
	counterNames[31] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_WRONG_SIZE"
	counterNames[32] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_COULD_NOT_READ_TOKEN"
	counterNames[33] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_TOKEN_EXPIRED"
	counterNames[34] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[35] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[40] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED"
	counterNames[41] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_WRONG_SIZE"
	counterNames[43] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[45] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED"
	counterNames[46] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[47] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[48] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[50] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED"
	counterNames[51] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_WRONG_SIZE"
	counterNames[52] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_READ_TOKEN"
	counterNames[53] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED"
	counterNames[55] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[56] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[60] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED"
	counterNames[61] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_WRONG_SIZE"
	counterNames[63] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED"
	counterNames[64] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[66] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[67] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[68] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[70] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED"
	counterNames[71] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_SMALL"
	counterNames[72] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_BIG"
	counterNames[74] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[76] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED"
	counterNames[77] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[78] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[79] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[80] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED"
	counterNames[81] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_SMALL"
	counterNames[82] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_BIG"
	counterNames[84] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[86] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED"
	counterNames[87] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[88] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[89] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[90] = "RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED"
	counterNames[91] = "RELAY_COUNTER_SESSION_PING_PACKET_WRONG_SIZE"
	counterNames[93] = "RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[95] = "RELAY_COUNTER_SESSION_PING_PACKET_ALREADY_RECEIVED"
	counterNames[96] = "RELAY_COUNTER_SESSION_PING_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[97] = "RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[98] = "RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[100] = "RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED"
	counterNames[101] = "RELAY_COUNTER_SESSION_PONG_PACKET_WRONG_SIZE"
	counterNames[103] = "RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[105] = "RELAY_COUNTER_SESSION_PONG_PACKET_ALREADY_RECEIVED"
	counterNames[106] = "RELAY_COUNTER_SESSION_PONG_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[107] = "RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[108] = "RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[110] = "RELAY_COUNTER_PACKETS_RECEIVED_BEFORE_INITIALIZE"
	counterNames[111] = "RELAY_COUNTER_UNKNOWN_PACKETS"
	counterNames[112] = "RELAY_COUNTER_PONGS_PROCESSED"
}

func relayCountersHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		relayName := vars["relay_name"]
		relayData := service.RelayData()
		relayIndex := -1
		for i := range relayData.RelayNames {
			if relayData.RelayNames[i] == relayName {
				relayIndex = i
				break
			}
		}
		if relayIndex == -1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		relayId := relayData.RelayIds[relayIndex]
		counters := relayManager.GetRelayCounters(relayId)
		w.Header().Set("Content-Type", "text/html")
		const htmlHeader = `<!DOCTYPE html>
		<html lang="en">
		<head>
		  <meta charset="utf-8">
		  <meta http-equiv="refresh" content="1">
		  <title>Relay Counters</title>
		  <style>
			table, th, td {
		      border: 1px solid black;
		      border-collapse: collapse;
		      text-align: center;
		      padding: 10px;
		    }
			*{
		    font-family:Courier;
		  }	  
		  </style>
		</head>
		<body>`
		fmt.Fprintf(w, "%s>\n", htmlHeader)
		fmt.Fprintf(w, "%s<br><br><table>\n", relayName)
		for i := range counterNames {
			if counterNames[i] == "" {
				continue
			}
			fmt.Fprintf(w, "<tr><td>%s</td><td>%d</td></tr>\n", counterNames[i], counters[i])
		}
		fmt.Fprintf(w, "</table>\n")
		const htmlFooter = `</body></html>`
		fmt.Fprintf(w, "%s\n", htmlFooter)
	}
}

func sendTrafficToMe(service *common.Service) func() bool {
	return func() bool {
		routeMatrixMutex.RLock()
		hasRouteMatrix := routeMatrixData != nil
		routeMatrixMutex.RUnlock()
		return hasRouteMatrix
	}
}

func machineIsHealthy() bool {
	return true
}

func ready(service *common.Service) func() bool {
	return func() bool {
		return initialDelayCompleted() && service.IsReady()
	}
}

func initialDelayCompleted() bool {
	delayMutex.RLock()
	result := delayCompleted
	delayMutex.RUnlock()
	return result
}

func UpdateInitialDelayState(service *common.Service) {
	go func() {
		for {
			if int(time.Since(startTime).Seconds()) >= initialDelay {
				core.Debug("initial delay completed")
				delayMutex.Lock()
				delayCompleted = true
				delayMutex.Unlock()
				return
			}
			time.Sleep(time.Second)
		}
	}()
}

func relaysHandler(w http.ResponseWriter, r *http.Request) {
	relaysMutex.RLock()
	responseData := relaysCSVData
	relaysMutex.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	buffer := bytes.NewBuffer(responseData)
	_, err := buffer.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type RelayJSON struct {
	RelayIds           []string  `json:"relay_ids"`
	RelayNames         []string  `json:"relay_names"`
	RelayAddresses     []string  `json:"relay_addresses"`
	RelayLatitudes     []float32 `json:"relay_latitudes"`
	RelayLongitudes    []float32 `json:"relay_longitudes"`
	RelayDatacenterIds []string  `json:"relay_datacenter_ids"`
	RelayIdToIndex     []string  `json:"relay_id_to_index"`
	DestRelays         []string  `json:"dest_relays"`
	DestRelayNames     []string  `json:"dest_relay_names"`
}

func relayDataHandler(service *common.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		relayData := service.RelayData()
		relayJSON := RelayJSON{}
		relayJSON.RelayIds = make([]string, relayData.NumRelays)
		relayJSON.RelayNames = relayData.RelayNames
		relayJSON.RelayAddresses = make([]string, relayData.NumRelays)
		relayJSON.RelayDatacenterIds = make([]string, relayData.NumRelays)
		relayJSON.RelayLatitudes = relayData.RelayLatitudes
		relayJSON.RelayLongitudes = relayData.RelayLongitudes
		relayJSON.RelayIdToIndex = make([]string, relayData.NumRelays)
		for i := 0; i < relayData.NumRelays; i++ {
			relayJSON.RelayIdToIndex[i] = fmt.Sprintf("%016x - %d", relayData.RelayIds[i], i)
		}
		relayJSON.DestRelays = make([]string, relayData.NumRelays)
		for i := 0; i < relayData.NumRelays; i++ {
			relayJSON.RelayIds[i] = fmt.Sprintf("%016x", relayData.RelayIds[i])
			relayJSON.RelayAddresses[i] = relayData.RelayAddresses[i].String()
			relayJSON.RelayDatacenterIds[i] = fmt.Sprintf("%016x", relayData.RelayDatacenterIds[i])
			if relayData.DestRelays[i] {
				relayJSON.DestRelays[i] = "1"
				relayJSON.DestRelayNames = append(relayJSON.DestRelayNames, relayData.RelayNames[i])
			} else {
				relayJSON.DestRelays[i] = "0"
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(relayJSON); err != nil {
			core.Error("could not write relay data json: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func costMatrixHandler(w http.ResponseWriter, r *http.Request) {
	costMatrixMutex.RLock()
	core.Debug("cost matrix handler (%d bytes)", len(costMatrixData))
	responseData := costMatrixData
	costMatrixMutex.RUnlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	buffer := bytes.NewBuffer(responseData)
	_, err := buffer.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func routeMatrixHandler(w http.ResponseWriter, r *http.Request) {
	routeMatrixMutex.RLock()
	core.Debug("route matrix handler (%d bytes)", len(routeMatrixData))
	responseData := routeMatrixData
	routeMatrixMutex.RUnlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	buffer := bytes.NewBuffer(responseData)
	_, err := buffer.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func UpdateRouteMatrix(service *common.Service, relayManager *common.RelayManager) {

	ticker := time.NewTicker(routeMatrixInterval)

	go func() {

		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				timeStart := time.Now()

				// build relays csv

				currentTime := time.Now().Unix()

				relayData := service.RelayData()

				relaysCSVDataNew := relayManager.GetRelaysCSV(currentTime, relayData.RelayIds, relayData.RelayNames, relayData.RelayAddresses)

				// build the cost matrix

				activeRelays := relayManager.GetActiveRelays(currentTime)

				costs := relayManager.GetCosts(currentTime, relayData.RelayIds, float32(maxJitter), maxPacketLoss)

				costMatrixNew := &common.CostMatrix{
					Version:            common.CostMatrixVersion_Write,
					RelayIds:           relayData.RelayIds,
					RelayAddresses:     relayData.RelayAddresses,
					RelayNames:         relayData.RelayNames,
					RelayLatitudes:     relayData.RelayLatitudes,
					RelayLongitudes:    relayData.RelayLongitudes,
					RelayDatacenterIds: relayData.RelayDatacenterIds,
					DestRelays:         relayData.DestRelays,
					Costs:              costs,
				}

				// optimize cost matrix -> route matrix

				numCPUs := runtime.NumCPU()

				numSegments := relayData.NumRelays
				if numCPUs < relayData.NumRelays {
					numSegments = relayData.NumRelays / 5
					if numSegments == 0 {
						numSegments = 1
					}
				}

				// write cost matrix data

				costMatrixDataNew, err := costMatrixNew.Write()
				if err != nil {
					core.Error("could not write cost matrix: %v", err)
					continue
				}

				// optimize

				routeEntries := core.Optimize2(relayData.NumRelays, numSegments, costs, relayData.RelayDatacenterIds, relayData.DestRelays)

				timeFinish := time.Now()

				optimizeDuration := timeFinish.Sub(timeStart)

				core.Debug("updated route matrix: %d relays in %dms", relayData.NumRelays, optimizeDuration.Milliseconds())

				if optimizeDuration.Milliseconds() > routeMatrixInterval.Milliseconds() {
					core.Warn("optimize can't keep up! increase the number of cores or increase ROUTE_MATRIX_INTERVAL to provide more time to complete the optimization!")
				}

				// create new route matrix

				routeMatrixNew := &common.RouteMatrix{
					CreatedAt:          uint64(time.Now().Unix()),
					Version:            common.RouteMatrixVersion_Write,
					RelayIds:           costMatrixNew.RelayIds,
					RelayAddresses:     costMatrixNew.RelayAddresses,
					RelayNames:         costMatrixNew.RelayNames,
					RelayLatitudes:     costMatrixNew.RelayLatitudes,
					RelayLongitudes:    costMatrixNew.RelayLongitudes,
					RelayDatacenterIds: costMatrixNew.RelayDatacenterIds,
					DestRelays:         costMatrixNew.DestRelays,
					RouteEntries:       routeEntries,
					BinFileBytes:       int32(len(relayData.DatabaseBinFile)),
					BinFileData:        relayData.DatabaseBinFile,
					CostMatrixSize:     uint32(len(costMatrixDataNew)),
					OptimizeTime:       uint32(optimizeDuration.Milliseconds()),
				}

				// write route matrix data

				routeMatrixDataNew, err := routeMatrixNew.Write()
				if err != nil {
					core.Error("could not write route matrix: %v", err)
					continue
				}

				// store our data in redis

				service.Store("relays", relaysCSVDataNew)
				service.Store("cost_matrix", costMatrixDataNew)
				service.Store("route_matrix", routeMatrixDataNew)

				// load the leader data from redis

				relaysCSVDataNew = service.Load("relays")
				if relaysCSVDataNew == nil {
					continue
				}

				costMatrixDataNew = service.Load("cost_matrix")
				if costMatrixDataNew == nil {
					continue
				}

				routeMatrixDataNew = service.Load("route_matrix")
				if routeMatrixDataNew == nil {
					continue
				}

				// serve up as official data

				relaysMutex.Lock()
				relaysCSVData = relaysCSVDataNew
				relaysMutex.Unlock()

				costMatrixMutex.Lock()
				costMatrixData = costMatrixDataNew
				costMatrixMutex.Unlock()

				routeMatrixMutex.Lock()
				routeMatrixData = routeMatrixDataNew
				routeMatrixMutex.Unlock()

				// analyze route matrix and send time series data to redis if leader

				if enableRedisTimeSeries {

					analysis := routeMatrixNew.Analyze()

					keys := []string{
						"active_relays",
						"total_relays",
						"route_matrix_total_routes",
						"route_matrix_average_num_routes",
						"route_matrix_average_route_length",
						"route_matrix_no_route_percent",
						"route_matrix_one_route_percent",
						"route_matrix_no_direct_route_percent",
						"route_matrix_database_bytes",
						"route_matrix_cost_matrix_bytes",
						"route_matrix_bytes",
						"route_matrix_optimize_ms",
					}

					values := []float64{
						float64(len(activeRelays)),
						float64(len(routeMatrixNew.RelayIds)),
						float64(analysis.TotalRoutes),
						float64(analysis.AverageNumRoutes),
						float64(analysis.AverageRouteLength),
						float64(analysis.NoRoutePercent),
						float64(analysis.OneRoutePercent),
						float64(analysis.NoDirectRoutePercent),
						float64(len(relayData.DatabaseBinFile)),
						float64(len(costMatrixDataNew)),
						float64(len(routeMatrixDataNew)),
						float64(optimizeDuration.Milliseconds()),
					}

					message := common.RedisTimeSeriesMessage{}
					message.Timestamp = uint64(time.Now().UnixNano() / 1000000)
					message.Keys = keys
					message.Values = values

					if service.IsLeader() {
						timeSeriesPublisher.MessageChannel <- &message
					}
				}
			}
		}
	}()
}
