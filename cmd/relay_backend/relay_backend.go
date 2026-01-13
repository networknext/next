package main

import (
	"bytes"
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
	"context"

	"github.com/gorilla/mux"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/messages"
	"github.com/networknext/next/modules/packets"
	"github.com/networknext/next/modules/portal"

	"github.com/hamba/avro"
	"github.com/redis/go-redis/v9"
)

var maxJitter int32
var maxPacketLoss float32
var routeMatrixInterval time.Duration

var redisHostname string
var redisCluster []string

var analyticsRelayUpdateGooglePubsubTopic string
var analyticsRelayUpdateGooglePubsubChannelSize int

var analyticsRouteMatrixUpdateGooglePubsubTopic string
var analyticsRouteMatrixUpdateGooglePubsubChannelSize int

var analyticsRelayToRelayPingGooglePubsubTopic string
var analyticsRelayToRelayPingGooglePubsubChannelSize int

var enableRelayToRelayPingAnalytics bool

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

var startTime int64

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
var analyticsRouteMatrixUpdateProducer *common.GooglePubsubProducer
var analyticsRelayToRelayPingProducer []*common.GooglePubsubProducer
var analyticsRelayToRelayPingReps int

var postRelayUpdateRequestChannel chan *packets.RelayUpdateRequestPacket

var relayBackendPublicKey []byte
var relayBackendPrivateKey []byte

//go:embed relay_update.json
var relayUpdateSchemaData string

//go:embed route_matrix_update.json
var routeMatrixUpdateSchemaData string

//go:embed relay_to_relay_ping.json
var relayToRelayPingSchemaData string

var relayUpdateSchema avro.Schema
var routeMatrixUpdateSchema avro.Schema
var relayToRelayPingSchema avro.Schema

var lastTimeSeriesUpdateTime map[uint64]int64 // IMPORTANT: per-relay, we only send time series stats once per-minute, otherwise we overload the time series redis @ 1000 relays

var enableRelayHistory bool

func main() {

	service := common.CreateService("relay_backend")

	maxJitter = int32(envvar.GetInt("MAX_JITTER", 1000))
	maxPacketLoss = float32(envvar.GetFloat("MAX_PACKET_LOSS", 100.0))
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)

	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisCluster := envvar.GetStringArray("REDIS_CLUSTER", []string{})

	analyticsRelayUpdateGooglePubsubTopic = envvar.GetString("ANALYTICS_RELAY_UPDATE_GOOGLE_PUBSUB_TOPIC", "relay_update")
	analyticsRelayUpdateGooglePubsubChannelSize = envvar.GetInt("ANALYTICS_RELAY_UPDATE_GOOGLE_PUBSUB_CHANNEL_SIZE", 10*1024*1024)

	analyticsRouteMatrixUpdateGooglePubsubTopic = envvar.GetString("ANALYTICS_ROUTE_MATRIX_UPDATE_GOOGLE_PUBSUB_TOPIC", "route_matrix_update")
	analyticsRouteMatrixUpdateGooglePubsubChannelSize = envvar.GetInt("ANALYTICS_ROUTE_MATRIX_UPDATE_GOOGLE_PUBSUB_CHANNEL_SIZE", 10*1024*1024)

	analyticsRelayToRelayPingGooglePubsubTopic = envvar.GetString("ANALYTICS_RELAY_TO_RELAY_PING_GOOGLE_PUBSUB_TOPIC", "relay_to_relay_ping")
	analyticsRelayToRelayPingGooglePubsubChannelSize = envvar.GetInt("ANALYTICS_RELAY_TO_RELAY_PING_GOOGLE_PUBSUB_CHANNEL_SIZE", 10*1024*1024)
	analyticsRelayToRelayPingReps = envvar.GetInt("ANALYTICS_RELAY_TO_RELAY_PING_REPS", 16)

	enableRelayHistory = envvar.GetBool("ENABLE_RELAY_HISTORY", false)

	enableGooglePubsub = envvar.GetBool("ENABLE_GOOGLE_PUBSUB", false)

	enableRelayToRelayPingAnalytics = envvar.GetBool("ENABLE_RELAY_TO_RELAY_PING_ANALYTICS", false)

	initialDelay = envvar.GetInt("INITIAL_DELAY", 15)

	relayInserterBatchSize = envvar.GetInt("RELAY_INSERTER_BATCH_SIZE", 1024)

	startTime = time.Now().Unix()

	lastTimeSeriesUpdateTime = make(map[uint64]int64, constants.MaxRelays)

	enableRedisTimeSeries = envvar.GetBool("ENABLE_REDIS_TIME_SERIES", false)
	redisTimeSeriesCluster = envvar.GetStringArray("REDIS_TIME_SERIES_CLUSTER", []string{})
	redisTimeSeriesHostname = envvar.GetString("REDIS_TIME_SERIES_HOSTNAME", "127.0.0.1:6379")

	if enableRedisTimeSeries {
		core.Debug("redis time series cluster: %s", redisTimeSeriesCluster)
		core.Debug("redis time series hostname: %s", redisTimeSeriesHostname)
	}

	redisPortalCluster = envvar.GetStringArray("REDIS_PORTAL_CLUSTER", []string{})
	redisPortalHostname = envvar.GetString("REDIS_PORTAL_HOSTNAME", "127.0.0.1:6379")

	relayBackendPublicKey = envvar.GetBase64("RELAY_BACKEND_PUBLIC_KEY", []byte{})
	relayBackendPrivateKey = envvar.GetBase64("RELAY_BACKEND_PRIVATE_KEY", []byte{})

	if len(relayBackendPublicKey) == 0 {
		core.Error("You must supply RELAY_BACKEND_PUBLIC_KEY")
		os.Exit(1)
	}

	if len(relayBackendPrivateKey) == 0 {
		core.Error("You must supply RELAY_BACKEND_PRIVATE_KEY")
		os.Exit(1)
	}

	if len(redisCluster) > 0 {
		core.Debug("redis cluster: %v", redisCluster)
	} else {
		core.Debug("redis hostname: %s", redisHostname)
	}

	core.Debug("max jitter: %d", maxJitter)
	core.Debug("max packet loss: %.1f", maxPacketLoss)
	core.Debug("route matrix interval: %s", routeMatrixInterval)
	core.Debug("redis host name: %s", redisHostname)

	core.Debug("analytics relay update google pubsub topic: %s", analyticsRelayUpdateGooglePubsubTopic)
	core.Debug("analytics relay update google pubsub channel size: %d", analyticsRelayUpdateGooglePubsubChannelSize)

	core.Debug("analytics route matrix update google pubsub topic: %s", analyticsRouteMatrixUpdateGooglePubsubTopic)
	core.Debug("analytics route matrix update google pubsub channel size: %d", analyticsRouteMatrixUpdateGooglePubsubChannelSize)

	core.Debug("analytics relay to relay ping google pubsub topic: %s", analyticsRelayToRelayPingGooglePubsubTopic)
	core.Debug("analytics relay to relay ping google pubsub channel size: %d", analyticsRelayToRelayPingGooglePubsubChannelSize)

	core.Debug("enable google pubsub: %v", enableGooglePubsub)

	core.Debug("enable relay to relay ping analytics: %v", enableRelayToRelayPingAnalytics)

	core.Debug("initial delay: %d", initialDelay)

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

		analyticsRouteMatrixUpdateProducer, err = common.CreateGooglePubsubProducer(service.Context, common.GooglePubsubConfig{
			ProjectId:          service.GoogleProjectId,
			Topic:              analyticsRouteMatrixUpdateGooglePubsubTopic,
			MessageChannelSize: analyticsRouteMatrixUpdateGooglePubsubChannelSize,
		})
		if err != nil {
			core.Error("could not create analytics route matrix update google pubsub producer")
			os.Exit(1)
		}

		if enableRelayToRelayPingAnalytics {
			// IMPORTANT: this is too expensive to have on by default. use for debugging only.
			analyticsRelayToRelayPingProducer = make([]*common.GooglePubsubProducer, analyticsRelayToRelayPingReps)
			for i := 0; i < analyticsRelayToRelayPingReps; i++ {
				analyticsRelayToRelayPingProducer[i], err = common.CreateGooglePubsubProducer(service.Context, common.GooglePubsubConfig{
					ProjectId:          service.GoogleProjectId,
					Topic:              analyticsRelayToRelayPingGooglePubsubTopic,
					MessageChannelSize: analyticsRelayToRelayPingGooglePubsubChannelSize,
				})
				if err != nil {
					core.Error("could not create analytics relay to relay ping google pubsub producer")
					os.Exit(1)
				}
			}
		}

		relayUpdateSchema, err = avro.Parse(relayUpdateSchemaData)
		if err != nil {
			panic(fmt.Sprintf("invalid relay update schema: %v", err))
		}

		routeMatrixUpdateSchema, err = avro.Parse(routeMatrixUpdateSchemaData)
		if err != nil {
			panic(fmt.Sprintf("invalid route matrix update schema: %v", err))
		}

		relayToRelayPingSchema, err = avro.Parse(relayToRelayPingSchemaData)
		if err != nil {
			panic(fmt.Sprintf("invalid relay to relay ping schema: %v", err))
		}
	}

	postRelayUpdateRequestChannel = make(chan *packets.RelayUpdateRequestPacket, 1024*1024)

	service.LoadDatabase(relayBackendPublicKey, relayBackendPrivateKey)

	initCounterNames()

	relayManager := common.CreateRelayManager(enableRelayHistory)

	service.Router.HandleFunc("/relay_update", relayUpdateHandler(service, relayManager)).Methods("POST")
	service.Router.HandleFunc("/relays", relaysHandler)
	service.Router.HandleFunc("/relay_data", relayDataHandler(service))
	service.Router.HandleFunc("/cost_matrix", costMatrixHandler)
	service.Router.HandleFunc("/route_matrix", routeMatrixHandler)
	service.Router.HandleFunc("/relay_counters/{relay_name}", relayCountersHandler(service, relayManager))
	service.Router.HandleFunc("/relay_history/{src}/{dest}", relayHistoryHandler(service, relayManager))
	service.Router.HandleFunc("/relay_manager", relayManagerHandler(service, relayManager))
	service.Router.HandleFunc("/costs", costsHandler(service, relayManager))
	service.Router.HandleFunc("/active_relays", activeRelaysHandler(service, relayManager))

	service.SetHealthFunctions(sendTrafficToMe(service), machineIsHealthy, ready(service))

	service.StartWebServer()

	service.LeaderElection(initialDelay)

	UpdateRelayBackendInstance(service)

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

		body, err := io.ReadAll(r.Body)
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

			if relayUpdateRequest.CurrentTime < currentTime-5 {
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

func UpdateRelayBackendInstance(service *common.Service) {

	var redisClient redis.Cmdable
	if len(redisCluster) > 0 {
		redisClient = common.CreateRedisClusterClient(redisCluster)
	} else {
		redisClient = common.CreateRedisClient(redisHostname)
	}

	go func() {

		ctx := context.Background()

		ticker := time.NewTicker(time.Second)

		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				core.Debug("updated relay backend instance")

				// todo: we need to be able to look up our local IP address and port here
				address := "127.0.0.1:30001"

				err := redisClient.HSet(ctx, "relay-backends", address, "1").Err()
				if err != nil {
					core.Warn("failed to update relay backend field in redis: %v", err)
				}

				err = redisClient.HExpire(ctx, "relay-backends", 30 * time.Second, address).Err()
				if err != nil {
					core.Warn("failed to set hexpire on relay backend field in redis: %v", err)
				}
			}
		}
	}()
}

func PostRelayUpdateRequest(service *common.Service) {

	for {
		relayUpdateRequest := <-postRelayUpdateRequestChannel

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

		var pingMessages []messages.AnalyticsRelayToRelayPingMessage
		if enableRelayToRelayPingAnalytics {
			pingMessages = make([]messages.AnalyticsRelayToRelayPingMessage, numSamples)
		}

		timestamp := time.Now().UnixNano() / 1000 // nano -> microseconds

		for i := 0; i < numSamples; i++ {

			rtt := relayUpdateRequest.SampleRTT[i]
			jitter := relayUpdateRequest.SampleJitter[i]
			pl := float32(relayUpdateRequest.SamplePacketLoss[i] / 65535.0 * 100.0)

			if rtt < 255 && int32(jitter) <= maxJitter && pl <= maxPacketLoss {
				numRoutable++
			}

			sampleRelayId := relayUpdateRequest.SampleRelayId[i]

			if enableRelayToRelayPingAnalytics {
				pingMessages[i] = messages.AnalyticsRelayToRelayPingMessage{
					Timestamp:          timestamp,
					SourceRelayId:      int64(relayId),
					DestinationRelayId: int64(sampleRelayId),
					RTT:                int32(rtt),
					Jitter:             int32(jitter),
					PacketLoss:         pl,
				}
			}
		}

		numUnroutable := numSamples - numRoutable

		// send relay update message to portal
		{
			message := messages.PortalRelayUpdateMessage{
				Timestamp:                 uint64(timestamp / 1000000), // microseconds -> seconds
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
				ClientPingsPerSecond:      relayUpdateRequest.ClientPingsPerSecond,
				ServerPingsPerSecond:      relayUpdateRequest.ServerPingsPerSecond,
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
					NumSessions:  message.SessionCount,
					MaxSessions:  message.MaxSessions,
					StartTime:    message.StartTime,
					RelayFlags:   message.RelayFlags,
					RelayVersion: message.RelayVersion,
				}

				relayInserter.Insert(service.Context, &relayData)

				if enableRedisTimeSeries {

					// send counters to redis

					countersPublisher.MessageChannel <- "relay_update"

					// send relay time series data to redis (once per-minute only!)

					sendTimeSeries := false
					currentTime := time.Now().Unix()
					lastUpdateTime, exists := lastTimeSeriesUpdateTime[message.RelayId]
					if !exists || lastUpdateTime+60 < currentTime {
						lastTimeSeriesUpdateTime[message.RelayId] = currentTime
						sendTimeSeries = true
					}

					if sendTimeSeries {

						timeSeriesMessage := common.RedisTimeSeriesMessage{}

						timeSeriesMessage.Timestamp = uint64(timestamp / 1000) // microseconds -> milliseconds

						timeSeriesMessage.Keys = []string{
							fmt.Sprintf("relay_%016x_session_count", message.RelayId),
							fmt.Sprintf("relay_%016x_packets_sent_per_second", message.RelayId),
							fmt.Sprintf("relay_%016x_packets_received_per_second", message.RelayId),
							fmt.Sprintf("relay_%016x_bandwidth_sent_kbps", message.RelayId),
							fmt.Sprintf("relay_%016x_bandwidth_received_kbps", message.RelayId),
						}

						timeSeriesMessage.Values = []float64{
							float64(message.SessionCount),
							float64(message.PacketsSentPerSecond),
							float64(message.PacketsReceivedPerSecond),
							float64(message.BandwidthSentKbps),
							float64(message.BandwidthReceivedKbps),
						}

						timeSeriesPublisher.MessageChannel <- &timeSeriesMessage
					}
				}
			}
		}

		// send relay update message to analytics
		{
			relayCounters := make([]int64, relayUpdateRequest.NumRelayCounters)

			for i := range relayCounters {
				relayCounters[i] = int64(relayUpdateRequest.RelayCounters[i])
			}

			message := messages.AnalyticsRelayUpdateMessage{
				Timestamp:                 timestamp,
				RelayId:                   int64(relayId),
				SessionCount:              int32(relayUpdateRequest.SessionCount),
				MaxSessions:               int32(relayData.RelayArray[relayIndex].MaxSessions),
				EnvelopeBandwidthUpKbps:   int64(relayUpdateRequest.EnvelopeBandwidthUpKbps),
				EnvelopeBandwidthDownKbps: int64(relayUpdateRequest.EnvelopeBandwidthDownKbps),
				PacketsSentPerSecond:      relayUpdateRequest.PacketsSentPerSecond,
				PacketsReceivedPerSecond:  relayUpdateRequest.PacketsReceivedPerSecond,
				BandwidthSentKbps:         relayUpdateRequest.BandwidthSentKbps,
				BandwidthReceivedKbps:     relayUpdateRequest.BandwidthReceivedKbps,
				ClientPingsPerSecond:      relayUpdateRequest.ClientPingsPerSecond,
				ServerPingsPerSecond:      relayUpdateRequest.ServerPingsPerSecond,
				RelayPingsPerSecond:       relayUpdateRequest.RelayPingsPerSecond,
				RelayFlags:                int64(relayUpdateRequest.RelayFlags),
				RelayCounters:             relayCounters,
				NumRoutable:               int32(numRoutable),
				NumUnroutable:             int32(numUnroutable),
				StartTime:                 int64(relayUpdateRequest.StartTime),
				CurrentTime:               int64(relayUpdateRequest.CurrentTime),
			}

			if service.IsLeader() && relayUpdateSchema != nil {
				data, err := avro.Marshal(relayUpdateSchema, &message)
				if err == nil {
					analyticsRelayUpdateProducer.MessageChannel <- data
				} else {
					core.Warn("failed to encode relay update message: %v", err)
				}
			}
		}

		// send relay to relay ping messages to analytics

		if service.IsLeader() && enableRelayToRelayPingAnalytics {
			for i := 0; i < len(pingMessages); i++ {
				data, err := avro.Marshal(relayToRelayPingSchema, &pingMessages[i])
				if err == nil {
					analyticsRelayToRelayPingProducer[i%analyticsRelayToRelayPingReps].MessageChannel <- data
				} else {
					core.Warn("failed to encode relay to relay ping message: %v", err)
				}
			}
		}
	}
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

func relayHistoryHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		routeMatrixMutex.RLock()
		data := routeMatrixData
		routeMatrixMutex.RUnlock()

		var routeMatrix common.RouteMatrix
		err := routeMatrix.Read(data)
		if err != nil {
			fmt.Fprintf(w, "error: could not read route matrix: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		vars := mux.Vars(r)

		src := vars["src"]
		dest := vars["dest"]

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

		if src_index == -1 {
			fmt.Printf("error: could not find source relay '%s'\n", src)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if dest_index == -1 {
			fmt.Printf("error: could not find source relay '%s'\n", dest)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if src_index == dest_index {
			fmt.Printf("error: no history between same relays\n")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		sourceRelayId := routeMatrix.RelayIds[src_index]
		destRelayId := routeMatrix.RelayIds[dest_index]

		rtt, jitter, packetLoss := relayManager.GetHistory(sourceRelayId, destRelayId)

		fmt.Fprintf(w, "history: %s -> %s\n", src, dest)
		fmt.Fprintf(w, "%v\n", rtt)
		fmt.Fprintf(w, "%v\n", jitter)
		fmt.Fprintf(w, "%v\n", packetLoss)
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

func initCounterNames() {
	// awk '/^#define RELAY_COUNTER_/ {print "    counterNames["$3"] = \""$2"\""}' ./relay/reference/reference_relay.cpp
	counterNames[0] = "RELAY_COUNTER_PACKETS_SENT"
	counterNames[1] = "RELAY_COUNTER_PACKETS_RECEIVED"
	counterNames[2] = "RELAY_COUNTER_BYTES_SENT"
	counterNames[3] = "RELAY_COUNTER_BYTES_RECEIVED"
	counterNames[4] = "RELAY_COUNTER_BASIC_PACKET_FILTER_DROPPED_PACKET"
	counterNames[5] = "RELAY_COUNTER_ADVANCED_PACKET_FILTER_DROPPED_PACKET"
	counterNames[6] = "RELAY_COUNTER_SESSION_CREATED"
	counterNames[7] = "RELAY_COUNTER_SESSION_CONTINUED"
	counterNames[8] = "RELAY_COUNTER_SESSION_DESTROYED"
	counterNames[10] = "RELAY_COUNTER_RELAY_PING_PACKET_SENT"
	counterNames[11] = "RELAY_COUNTER_RELAY_PING_PACKET_RECEIVED"
	counterNames[12] = "RELAY_COUNTER_RELAY_PING_PACKET_DID_NOT_VERIFY"
	counterNames[13] = "RELAY_COUNTER_RELAY_PING_PACKET_EXPIRED"
	counterNames[14] = "RELAY_COUNTER_RELAY_PING_PACKET_WRONG_SIZE"
	counterNames[15] = "RELAY_COUNTER_RELAY_PING_PACKET_UNKNOWN_RELAY"
	counterNames[15] = "RELAY_COUNTER_RELAY_PONG_PACKET_SENT"
	counterNames[16] = "RELAY_COUNTER_RELAY_PONG_PACKET_RECEIVED"
	counterNames[17] = "RELAY_COUNTER_RELAY_PONG_PACKET_WRONG_SIZE"
	counterNames[18] = "RELAY_COUNTER_RELAY_PONG_PACKET_UNKNOWN_RELAY"
	counterNames[20] = "RELAY_COUNTER_CLIENT_PING_PACKET_RECEIVED"
	counterNames[21] = "RELAY_COUNTER_CLIENT_PING_PACKET_WRONG_SIZE"
	counterNames[22] = "RELAY_COUNTER_CLIENT_PING_PACKET_RESPONDED_WITH_PONG"
	counterNames[23] = "RELAY_COUNTER_CLIENT_PING_PACKET_DID_NOT_VERIFY"
	counterNames[24] = "RELAY_COUNTER_CLIENT_PING_PACKET_EXPIRED"
	counterNames[30] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_RECEIVED"
	counterNames[31] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_WRONG_SIZE"
	counterNames[32] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_COULD_NOT_DECRYPT_ROUTE_TOKEN"
	counterNames[33] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_TOKEN_EXPIRED"
	counterNames[34] = "RELAY_COUNTER_ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP"
	counterNames[40] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_RECEIVED"
	counterNames[41] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_WRONG_SIZE"
	counterNames[42] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[43] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_SESSION_EXPIRED"
	counterNames[44] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED"
	counterNames[45] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[46] = "RELAY_COUNTER_ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP"
	counterNames[50] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_RECEIVED"
	counterNames[51] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_WRONG_SIZE"
	counterNames[52] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_DECRYPT_CONTINUE_TOKEN"
	counterNames[53] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED"
	counterNames[54] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[55] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_SESSION_EXPIRED"
	counterNames[56] = "RELAY_COUNTER_CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP"
	counterNames[60] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_RECEIVED"
	counterNames[61] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_WRONG_SIZE"
	counterNames[62] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED"
	counterNames[63] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[64] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_SESSION_EXPIRED"
	counterNames[65] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[66] = "RELAY_COUNTER_CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP"
	counterNames[70] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_RECEIVED"
	counterNames[71] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_SMALL"
	counterNames[72] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_TOO_BIG"
	counterNames[73] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[74] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_SESSION_EXPIRED"
	counterNames[75] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED"
	counterNames[76] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[77] = "RELAY_COUNTER_CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP"
	counterNames[80] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_RECEIVED"
	counterNames[81] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_SMALL"
	counterNames[82] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_TOO_BIG"
	counterNames[83] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[84] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_SESSION_EXPIRED"
	counterNames[85] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED"
	counterNames[86] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[87] = "RELAY_COUNTER_SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP"
	counterNames[90] = "RELAY_COUNTER_SESSION_PING_PACKET_RECEIVED"
	counterNames[91] = "RELAY_COUNTER_SESSION_PING_PACKET_WRONG_SIZE"
	counterNames[92] = "RELAY_COUNTER_SESSION_PING_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[93] = "RELAY_COUNTER_SESSION_PING_PACKET_SESSION_EXPIRED"
	counterNames[94] = "RELAY_COUNTER_SESSION_PING_PACKET_ALREADY_RECEIVED"
	counterNames[95] = "RELAY_COUNTER_SESSION_PING_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[96] = "RELAY_COUNTER_SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP"
	counterNames[100] = "RELAY_COUNTER_SESSION_PONG_PACKET_RECEIVED"
	counterNames[101] = "RELAY_COUNTER_SESSION_PONG_PACKET_WRONG_SIZE"
	counterNames[102] = "RELAY_COUNTER_SESSION_PONG_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[103] = "RELAY_COUNTER_SESSION_PONG_PACKET_SESSION_EXPIRED"
	counterNames[104] = "RELAY_COUNTER_SESSION_PONG_PACKET_ALREADY_RECEIVED"
	counterNames[105] = "RELAY_COUNTER_SESSION_PONG_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[106] = "RELAY_COUNTER_SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP"
	counterNames[110] = "RELAY_COUNTER_SERVER_PING_PACKET_RECEIVED"
	counterNames[111] = "RELAY_COUNTER_SERVER_PING_PACKET_WRONG_SIZE"
	counterNames[112] = "RELAY_COUNTER_SERVER_PING_PACKET_RESPONDED_WITH_PONG"
	counterNames[113] = "RELAY_COUNTER_SERVER_PING_PACKET_DID_NOT_VERIFY"
	counterNames[114] = "RELAY_COUNTER_SERVER_PING_PACKET_EXPIRED"
	counterNames[120] = "RELAY_COUNTER_PACKET_TOO_LARGE"
	counterNames[121] = "RELAY_COUNTER_PACKET_TOO_SMALL"
	counterNames[122] = "RELAY_COUNTER_DROP_FRAGMENT"
	counterNames[123] = "RELAY_COUNTER_DROP_LARGE_IP_HEADER"
	counterNames[124] = "RELAY_COUNTER_REDIRECT_NOT_IN_WHITELIST"
	counterNames[125] = "RELAY_COUNTER_DROPPED_PACKETS"
	counterNames[126] = "RELAY_COUNTER_DROPPED_BYTES"
	counterNames[127] = "RELAY_COUNTER_NOT_IN_WHITELIST"
	counterNames[128] = "RELAY_COUNTER_WHITELIST_ENTRY_EXPIRED"
	counterNames[130] = "RELAY_COUNTER_SESSIONS"
	counterNames[131] = "RELAY_COUNTER_ENVELOPE_KBPS_UP"
	counterNames[132] = "RELAY_COUNTER_ENVELOPE_KBPS_DOWN"
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
		return hasRouteMatrix && time.Now().Unix() > startTime+int64(initialDelay)
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
			currentTime := int64(time.Now().Unix())
			if currentTime-startTime >= int64(initialDelay) {
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
	w.Header().Set("Content-Type", "text/plain")
	buffer := bytes.NewBuffer(responseData)
	buffer.WriteTo(w)
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
		json.NewEncoder(w).Encode(relayJSON)
	}
}

func costMatrixHandler(w http.ResponseWriter, r *http.Request) {
	costMatrixMutex.RLock()
	core.Debug("cost matrix handler (%d bytes)", len(costMatrixData))
	responseData := costMatrixData
	costMatrixMutex.RUnlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	buffer := bytes.NewBuffer(responseData)
	buffer.WriteTo(w)
}

func routeMatrixHandler(w http.ResponseWriter, r *http.Request) {
	routeMatrixMutex.RLock()
	core.Debug("route matrix handler (%d bytes)", len(routeMatrixData))
	responseData := routeMatrixData
	routeMatrixMutex.RUnlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	buffer := bytes.NewBuffer(responseData)
	buffer.WriteTo(w)
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

				numRelays := len(relayData.RelayIds)

				costs := relayManager.GetCosts(currentTime, relayData.RelayIds, float32(maxJitter), maxPacketLoss)

				relayPrice := make([]byte, numRelays)

				copy(relayPrice, relayData.RelayPrice)

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
					RelayPrice:         relayPrice,
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

				routeEntries := core.Optimize2(relayData.NumRelays, numSegments, costs, relayPrice, relayData.RelayDatacenterIds, relayData.DestRelays)

				timeFinish := time.Now()

				optimizeDuration := timeFinish.Sub(timeStart)

				core.Debug("updated route matrix: %d relays in %dms", relayData.NumRelays, optimizeDuration.Milliseconds())

				if optimizeDuration.Milliseconds() > routeMatrixInterval.Milliseconds() {
					core.Warn("optimize can't keep up! increase the number of cores or increase ROUTE_MATRIX_INTERVAL to provide more time to complete the optimization!")
				}

				// create new route matrix

				routeMatrixNew := &common.RouteMatrix{
					CreatedAt:          uint64(currentTime),
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
					Costs:              costs,
					RelayPrice:         relayPrice,
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

				// analyze route matrix

				analysis := routeMatrixNew.Analyze()

				// send route matrix time series data to redis if leader

				if enableRedisTimeSeries {

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
					message.Timestamp = uint64(time.Now().UnixNano() / 1000000) // nano -> milliseconds
					message.Keys = keys
					message.Values = values

					if service.IsLeader() {
						timeSeriesPublisher.MessageChannel <- &message
					}
				}

				// send route matrix update to pubsub/bigquery if leader

				if enableGooglePubsub {

					numDestRelays := 0
					for i := range routeMatrixNew.DestRelays {
						if routeMatrixNew.DestRelays[i] {
							numDestRelays++
						}
					}

					message := messages.AnalyticsRouteMatrixUpdateMessage{}

					message.Timestamp = int64(time.Now().UnixNano() / 1000) // nano -> microseconds
					message.NumRelays = int32(len(routeMatrixNew.RelayIds))
					message.NumActiveRelays = int32(len(activeRelays))
					message.NumDestRelays = int32(numDestRelays)
					message.NumDatacenters = int32(len(relayData.RelayDatacenterIds))
					message.TotalRoutes = int32(analysis.TotalRoutes)
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
					message.CostMatrixSize = int32(len(costMatrixDataNew))
					message.RouteMatrixSize = int32(len(routeMatrixDataNew))
					message.DatabaseSize = int32(len(relayData.DatabaseBinFile))
					message.OptimizeTime = int32(optimizeDuration.Milliseconds())

					if service.IsLeader() {
						data, err := avro.Marshal(routeMatrixUpdateSchema, &message)
						if err == nil {
							analyticsRouteMatrixUpdateProducer.MessageChannel <- data
						} else {
							core.Warn("failed to encode route matrix update message: %v", err)
						}
					}
				}
			}
		}
	}()
}
