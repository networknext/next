package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	_ "github.com/networknext/backend/modules/messages"
	"github.com/networknext/backend/modules/packets"
)

var googleProjectId string
var maxRTT float32
var maxJitter float32
var maxPacketLoss float32
var costMatrixBufferSize int
var routeMatrixBufferSize int
var routeMatrixInterval time.Duration
var redisPubsubChannelName string
var relayUpdateChannelSize int
var pingStatsPubsubTopic string
var pingStatsChannelSize int
var relayStatsPubsubTopic string
var relayStatsChannelSize int
var readyDelay time.Duration

var relaysMutex sync.RWMutex
var relaysData []byte

var costMatrixMutex sync.RWMutex
var costMatrixData []byte

var routeMatrixMutex sync.RWMutex
var routeMatrixData []byte

var costMatrixInternalMutex sync.RWMutex
var costMatrixInternalData []byte

var routeMatrixInternalMutex sync.RWMutex
var routeMatrixInternalData []byte

var readyMutex sync.RWMutex
var ready bool

var startTime time.Time

func main() {

	service := common.CreateService("relay_backend")

	googleProjectId = envvar.GetString("GOOGLE_PROJECT_ID", "local")
	maxRTT = float32(envvar.GetFloat("MAX_RTT", 1000.0))
	maxJitter = float32(envvar.GetFloat("MAX_JITTER", 1000.0))
	maxPacketLoss = float32(envvar.GetFloat("MAX_JITTER", 100.0))
	costMatrixBufferSize = envvar.GetInt("COST_MATRIX_BUFFER_SIZE", 1*1024*1024)
	routeMatrixBufferSize = envvar.GetInt("ROUTE_MATRIX_BUFFER_SIZE", 10*1024*1024)
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)
	redisPubsubChannelName = envvar.GetString("REDIS_PUBSUB_CHANNEL_NAME", "relay_updates")
	relayUpdateChannelSize = envvar.GetInt("RELAY_UPDATE_CHANNEL_SIZE", 10*1024)
	pingStatsPubsubTopic = envvar.GetString("PING_STATS_PUBSUB_TOPIC", "ping_stats")
	pingStatsChannelSize = envvar.GetInt("PING_STATS_CHANNEL_SIZE", 10*1024)
	relayStatsPubsubTopic = envvar.GetString("RELAY_STATS_PUBSUB_TOPIC", "relay_stats")
	relayStatsChannelSize = envvar.GetInt("RELAY_STATS_CHANNEL_SIZE", 10*1024)
	readyDelay = envvar.GetDuration("READY_DELAY", 6*time.Minute)
	startTime = time.Now()

	core.Log("google project id: %s", googleProjectId)
	core.Log("max rtt: %.1f", maxRTT)
	core.Log("max jitter: %.1f", maxJitter)
	core.Log("max packet loss: %.1f", maxPacketLoss)
	core.Log("cost matrix buffer size: %d bytes", costMatrixBufferSize)
	core.Log("route matrix buffer size: %d bytes", routeMatrixBufferSize)
	core.Log("route matrix interval: %s", routeMatrixInterval)
	core.Log("redis pubsub channel name: %s", redisPubsubChannelName)
	core.Log("relay update channel size: %d", relayUpdateChannelSize)
	core.Log("ping stats pubsub channel: %s", pingStatsPubsubTopic)
	core.Log("ping stats channel size: %d", pingStatsChannelSize)
	core.Log("relay stats pubsub channel: %s", relayStatsPubsubTopic)
	core.Log("relay stats channel size: %d", relayStatsChannelSize)
	core.Log("ready delay: %s", readyDelay.String())
	core.Log("start time: %s", startTime.String())

	service.LoadDatabase()

	relayManager := common.CreateRelayManager()

	service.Router.HandleFunc("/relays", relaysHandler)
	service.Router.HandleFunc("/relay_data", relayDataHandler(service))
	service.Router.HandleFunc("/cost_matrix", costMatrixHandler)
	service.Router.HandleFunc("/route_matrix", routeMatrixHandler)
	service.Router.HandleFunc("/cost_matrix_internal", costMatrixInternalHandler)
	service.Router.HandleFunc("/route_matrix_internal", routeMatrixInternalHandler)

	service.Selector()

	service.OverrideHealthHandler(healthHandler)

	service.StartWebServer()

	ProcessRelayUpdates(service, relayManager)

	UpdateRouteMatrix(service, relayManager)

	UpdateReadyState()

	service.WaitForShutdown()
}

func UpdateReadyState() {
	go func() {
		for {
			time.Sleep(time.Second)

			routeMatrixMutex.RLock()
			routeMatrixReady := len(routeMatrixData) > 0
			routeMatrixMutex.RUnlock()

			routeMatrixInternalMutex.RLock()
			routeMatrixInternalReady := len(routeMatrixInternalData) > 0
			routeMatrixInternalMutex.RUnlock()

			delayReady := time.Since(startTime) >= readyDelay

			if routeMatrixReady && routeMatrixInternalReady && delayReady {
				core.Log("relay backend is ready")
				readyMutex.Lock()
				ready = true
				readyMutex.Unlock()
				break
			}
		}
	}()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {

	readyMutex.RLock()
	not_ready := !ready
	readyMutex.RUnlock()

	if not_ready {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func relaysHandler(w http.ResponseWriter, r *http.Request) {
	relaysMutex.RLock()
	responseData := relaysData
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
			} else {
				relayJSON.DestRelays[i] = "0"
			}
		}
		relayJSON.DestRelayNames = relayData.DestRelayNames
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(relayJSON); err != nil {
			core.Error("could not write relay data json: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func costMatrixHandler(w http.ResponseWriter, r *http.Request) {
	costMatrixMutex.RLock()
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
	responseData := routeMatrixData
	routeMatrixMutex.RUnlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	buffer := bytes.NewBuffer(responseData)
	_, err := buffer.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func costMatrixInternalHandler(w http.ResponseWriter, r *http.Request) {
	costMatrixInternalMutex.RLock()
	responseData := costMatrixInternalData
	costMatrixInternalMutex.RUnlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	buffer := bytes.NewBuffer(responseData)
	_, err := buffer.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func routeMatrixInternalHandler(w http.ResponseWriter, r *http.Request) {
	routeMatrixInternalMutex.RLock()
	responseData := routeMatrixInternalData
	routeMatrixInternalMutex.RUnlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	buffer := bytes.NewBuffer(responseData)
	_, err := buffer.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func ProcessRelayUpdates(service *common.Service, relayManager *common.RelayManager) {

	config := common.RedisPubsubConfig{}

	config.PubsubChannelName = redisPubsubChannelName
	config.MessageChannelSize = relayUpdateChannelSize

	consumer, err := common.CreateRedisPubsubConsumer(service.Context, config)

	if err != nil {
		core.Error("could not create redis pubsub consumer")
		os.Exit(1)
	}

	// todo
	/*
		pingStatsProducer, err := common.CreateGooglePubsubProducer(service.Context, common.GooglePubsubConfig{
			ProjectId:          googleProjectId,
			Topic:              pingStatsPubsubTopic,
			MessageChannelSize: pingStatsChannelSize,
		})
		if err != nil {
			core.Error("could not create ping stats producer")
			os.Exit(1)
		}

		relayStatsProducer, err := common.CreateGooglePubsubProducer(service.Context, common.GooglePubsubConfig{
			ProjectId:          googleProjectId,
			Topic:              relayStatsPubsubTopic,
			MessageChannelSize: relayStatsChannelSize,
		})
		if err != nil {
			core.Error("could not create relay stats producer")
			os.Exit(1)
		}
	*/

	go func() {

		for {
			select {

			case <-service.Context.Done():
				return

			case message := <-consumer.MessageChannel:

				// read the relay update request packet

				var relayUpdateRequest packets.RelayUpdateRequestPacket

				err = relayUpdateRequest.Read(message)
				if err != nil {
					core.Error("could not read relay update: %v", err)
					return
				}

				if relayUpdateRequest.Version != packets.VersionNumberRelayUpdateRequest {
					core.Error("relay update version mismatch")
					return
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

				core.Debug("received relay update for '%s'", relayName)

				// process samples in the relay update

				numSamples := int(relayUpdateRequest.NumSamples)

				relayManager.ProcessRelayUpdate(relayId,
					relayName,
					relayUpdateRequest.Address,
					int(relayUpdateRequest.SessionCount),
					relayUpdateRequest.RelayVersion,
					relayUpdateRequest.ShuttingDown,
					numSamples,
					relayUpdateRequest.SampleRelayId[:numSamples],
					relayUpdateRequest.SampleRTT[:numSamples],
					relayUpdateRequest.SampleJitter[:numSamples],
					relayUpdateRequest.SamplePacketLoss[:numSamples],
				)

				// Build relay stats

				// todo: below needs to be fixed up (Andrew)

				/*
					numRoutable := 0

					pingStatsMessages := make([]messages.PingStatsMessage, numSamples)

					for i := 0; i < numSamples; i++ {

						rtt := relayUpdate.PingStats[i].RTT
						jitter := relayUpdate.PingStats[i].Jitter
						pl := relayUpdate.PingStats[i].PacketLoss

						sampleRelayId := relayUpdate.PingStats[i].RelayID

						sampleRelayIds[i] = sampleRelayId
						sampleRTT[i] = rtt
						sampleJitter[i] = jitter
						samplePacketLoss[i] = pl

						// todo: remove dependency on routing
						if rtt != routing.InvalidRouteValue && jitter != routing.InvalidRouteValue && pl != routing.InvalidRouteValue {
							if jitter <= maxJitter && pl <= maxPacketLoss {
								numRoutable++
								sampleRoutable[i] = true
							}
						}

						pingStatsMessages[i] = messages.PingStatsMessage{
							Version:    messages.PingStatsMessageVersion,
							Timestamp:  uint64(time.Now().Unix()),
							RelayA:     relayId,
							RelayB:     sampleRelayId,
							RTT:        rtt,
							Jitter:     jitter,
							PacketLoss: pl,
							Routable:   sampleRoutable[i],
						}
					}
				*/

				/*
					numUnroutable := numSamples - numRoutable

					bwSentPercent := float32(0)
					bwRecvPercent := float32(0)

					if relayData.RelayArray[relayIndex].NICSpeedMbps > 0 {
						bwSentPercent = float32(relayUpdate.BandwidthSentKbps/uint64(relayData.RelayArray[relayIndex].NICSpeedMbps)) * 100.0
						bwRecvPercent = float32(relayUpdate.BandwidthRecvKbps/uint64(relayData.RelayArray[relayIndex].NICSpeedMbps)) * 100.0
					}

					envSentPercent := float32(0)
					envRecvPercent := float32(0)

					if relayUpdate.EnvelopeUpKbps > 0 {
						envSentPercent = float32(relayUpdate.BandwidthSentKbps/relayUpdate.EnvelopeUpKbps) * 100.0
					}

					if relayUpdate.EnvelopeUpKbps > 0 {
						envRecvPercent = float32(relayUpdate.BandwidthRecvKbps/relayUpdate.EnvelopeDownKbps) * 100.0
					}

					cpuUsage := relayUpdate.CPU

					maxSessions := relayData.RelayArray[relayIndex].MaxSessions
					numSessions := relayUpdate.SessionCount

					full := maxSessions != 0 && numSessions >= uint64(maxSessions)

					relayStatsMessage := messages.RelayStatsMessage{
						Version:                  messages.RelayStatsMessageVersion,
						Timestamp:                uint64(time.Now().Unix()),
						NumSessions:              uint32(numSessions),
						MaxSessions:              maxSessions,
						NumRoutable:              uint32(numRoutable),
						NumUnroutable:            uint32(numUnroutable),
						Full:                     full,
						CPUUsage:                 float32(cpuUsage),
						BandwidthSentPercent:     bwSentPercent,
						BandwidthReceivedPercent: bwRecvPercent,
						EnvelopeSentPercent:      envSentPercent,
						EnvelopeReceivedPercent:  envRecvPercent,
						BandwidthSentMbps:        float32(relayUpdate.BandwidthSentKbps),
						BandwidthReceivedMbps:    float32(relayUpdate.BandwidthRecvKbps),
						EnvelopeSentMbps:         float32(relayUpdate.EnvelopeUpKbps),
						EnvelopeReceivedMbps:     float32(relayUpdate.EnvelopeDownKbps),
					}

					// update relay stats

					if service.IsLeader() {

						messageBuffer := make([]byte, relayStatsChannelSize) // todo: WUT

						message := relayStatsMessage.Write(messageBuffer[:])

						relayStatsProducer.MessageChannel <- message
					}

					// update ping stats

					if service.IsLeader() {

						messageBuffer := make([]byte, pingStatsChannelSize) // todo: WUT!!!

						for i := 0; i < len(pingStatsMessages); i++ {
							message := pingStatsMessages[i].Write(messageBuffer[:])
							pingStatsProducer.MessageChannel <- message
						}
					}
				*/
			}
		}
	}()
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

				// build relays data

				relaysDataNew := relayManager.GetRelaysCSV()

				// build the cost matrix

				relayData := service.RelayData()

				costs := relayManager.GetCosts(relayData.RelayIds, maxRTT, maxJitter, maxPacketLoss, service.Local)

				costMatrixNew := &common.CostMatrix{
					Version:            common.CostMatrixSerializeVersion,
					RelayIds:           relayData.RelayIds,
					RelayAddresses:     relayData.RelayAddresses,
					RelayNames:         relayData.RelayNames,
					RelayLatitudes:     relayData.RelayLatitudes,
					RelayLongitudes:    relayData.RelayLongitudes,
					RelayDatacenterIds: relayData.RelayDatacenterIds,
					DestRelays:         relayData.DestRelays,
					Costs:              costs,
				}

				// serve up as internal cost matrix

				costMatrixDataNew, err := costMatrixNew.Write(costMatrixBufferSize)
				if err != nil {
					core.Error("could not write cost matrix: %v", err)
					continue
				}

				costMatrixInternalMutex.Lock()
				costMatrixInternalData = costMatrixDataNew
				costMatrixInternalMutex.Unlock()

				// optimize cost matrix -> route matrix

				numCPUs := runtime.NumCPU()

				numSegments := relayData.NumRelays
				if numCPUs < relayData.NumRelays {
					numSegments = relayData.NumRelays / 5
					if numSegments == 0 {
						numSegments = 1
					}
				}

				costThreshold := int32(1)

				routeMatrixNew := &common.RouteMatrix{
					CreatedAt:          uint64(time.Now().Unix()),
					Version:            common.RouteMatrixSerializeVersion,
					RelayIds:           costMatrixNew.RelayIds,
					RelayAddresses:     costMatrixNew.RelayAddresses,
					RelayNames:         costMatrixNew.RelayNames,
					RelayLatitudes:     costMatrixNew.RelayLatitudes,
					RelayLongitudes:    costMatrixNew.RelayLongitudes,
					RelayDatacenterIds: costMatrixNew.RelayDatacenterIds,
					DestRelays:         costMatrixNew.DestRelays,
					RouteEntries:       core.Optimize2(relayData.NumRelays, numSegments, costs, costThreshold, relayData.RelayDatacenterIds, relayData.DestRelays),
					BinFileBytes:       int32(len(relayData.DatabaseBinFile)),
					BinFileData:        relayData.DatabaseBinFile,
				}

				// serve up as internal route matrix

				routeMatrixDataNew, err := routeMatrixNew.Write(routeMatrixBufferSize)
				if err != nil {
					core.Error("could not write route matrix: %v", err)
					continue
				}

				routeMatrixInternalMutex.Lock()
				routeMatrixInternalData = routeMatrixDataNew
				routeMatrixInternalMutex.Unlock()

				// store our most recent cost and route matrix in redis

				dataStores := []common.DataStoreConfig{
					{
						Name: "relays",
						Data: relaysDataNew,
					},
					{
						Name: "cost_matrix",
						Data: costMatrixDataNew,
					},
					{
						Name: "route_matrix",
						Data: routeMatrixDataNew,
					},
				}

				service.UpdateSelectorStore(dataStores)

				// load the master cost and route matrix from redis (leader election)

				dataStores = service.LoadSelectorStore()

				if len(dataStores) == 0 {
					core.Error("failed to get data stores from redis selector")
					continue
				}

				relaysDataNew = dataStores[0].Data
				if relaysDataNew == nil {
					core.Error("failed to get relays from redis selector")
					continue
				}

				costMatrixDataNew = dataStores[1].Data
				if costMatrixDataNew == nil {
					core.Error("failed to get cost matrix from redis selector")
					continue
				}

				routeMatrixDataNew = dataStores[2].Data
				if routeMatrixDataNew == nil {
					core.Error("failed to get route matrix from redis selector")
					continue
				}

				// serve up as official data

				relaysMutex.Lock()
				relaysData = relaysDataNew
				relaysMutex.Unlock()

				costMatrixMutex.Lock()
				costMatrixData = costMatrixDataNew
				costMatrixMutex.Unlock()

				routeMatrixMutex.Lock()
				routeMatrixData = routeMatrixDataNew
				routeMatrixMutex.Unlock()

				// we are done!

				timeFinish := time.Now()

				optimizeDuration := timeFinish.Sub(timeStart)

				core.Debug("route optimization: %d relays in %s", relayData.NumRelays, optimizeDuration)
			}
		}
	}()
}
