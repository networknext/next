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
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/transport"
)

var maxRTT float32
var maxJitter float32
var maxPacketLoss float32
var costMatrixBufferSize int
var routeMatrixBufferSize int
var routeMatrixInterval time.Duration
var redisHostname string
var redisPassword string
var redisPubsubChannelName string
var relayUpdateChannelSize int
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

	service := common.CreateService("relay_backend_new")

	maxRTT = float32(envvar.GetFloat("MAX_RTT", 1000.0))
	maxJitter = float32(envvar.GetFloat("MAX_JITTER", 1000.0))
	maxPacketLoss = float32(envvar.GetFloat("MAX_JITTER", 100.0))
	costMatrixBufferSize = envvar.GetInt("COST_MATRIX_BUFFER_SIZE", 1*1024*1024)
	routeMatrixBufferSize = envvar.GetInt("ROUTE_MATRIX_BUFFER_SIZE", 10*1024*1024)
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)
	redisHostname = envvar.Get("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.Get("REDIS_PASSWORD", "")
	redisPubsubChannelName = envvar.Get("REDIS_PUBSUB_CHANNEL_NAME", "relay_updates")
	relayUpdateChannelSize = envvar.GetInt("RELAY_UPDATE_CHANNEL_SIZE", 10*1024)
	readyDelay = envvar.GetDuration("READY_DELAY", 6*time.Minute)
	startTime = time.Now()

	core.Log("max rtt: %.1f", maxRTT)
	core.Log("max jitter: %.1f", maxJitter)
	core.Log("max packet loss: %.1f", maxPacketLoss)
	core.Log("cost matrix buffer size: %d bytes", costMatrixBufferSize)
	core.Log("route matrix buffer size: %d bytes", routeMatrixBufferSize)
	core.Log("route matrix interval: %s", routeMatrixInterval)
	core.Log("redis hostname: %s", redisHostname)
	core.Log("redis password: %s", redisPassword)
	core.Log("redis pubsub channel name: %s", redisPubsubChannelName)
	core.Log("relay update channel size: %d", relayUpdateChannelSize)
	core.Log("ready delay: %s", readyDelay.String())
	core.Log("start time: %s", startTime.String())

	service.LoadDatabase()

	relayStats := common.CreateRelayStats()

	service.Router.HandleFunc("/health", healthHandler)
	service.Router.HandleFunc("/relays", relaysHandler)
	service.Router.HandleFunc("/relay_data", relayDataHandler(service))
	service.Router.HandleFunc("/cost_matrix", costMatrixHandler)
	service.Router.HandleFunc("/route_matrix", routeMatrixHandler)
	service.Router.HandleFunc("/cost_matrix_internal", costMatrixInternalHandler)
	service.Router.HandleFunc("/route_matrix_internal", routeMatrixInternalHandler)

	service.StartWebServer()

	ProcessRelayUpdates(service, relayStats)

	UpdateRouteMatrix(service, relayStats)

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
		fmt.Fprintf(w, "not ready")
	} else {
		fmt.Fprintf(w, "OK")
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

func ProcessRelayUpdates(service *common.Service, relayStats *common.RelayStats) {

	config := common.RedisPubsubConfig{}

	config.RedisHostname = redisHostname
	config.RedisPassword = redisPassword
	config.PubsubChannelName = redisPubsubChannelName
	config.MessageChannelSize = relayUpdateChannelSize

	consumer, err := common.CreateRedisPubsubConsumer(service.Context, config)

	if err != nil {
		core.Error("could not create redis pubsub consumer")
		os.Exit(1)
	}

	go func() {

		for {
			select {

			case <-service.Context.Done():
				return

			case message := <-consumer.MessageChannel:

				var relayUpdate transport.RelayUpdateRequest
				if err = relayUpdate.UnmarshalBinary(message); err != nil {
					core.Error("could not read relay update")
					return
				}

				relayData := service.RelayData()

				relayId := crypto.HashID(relayUpdate.Address.String())
				relayIndex, ok := relayData.RelayIdToIndex[relayId]
				if !ok {
					core.Error("unknown relay id %016x", relayId)
					return
				}

				relayName := relayData.RelayNames[relayIndex]

				core.Debug("received relay update for '%s'", relayName)

				/*
					type RelayUpdateRequest struct {
						Version           uint32
						Address           net.UDPAddr
						Token             []byte
						PingStats         []routing.RelayStatsPing
						SessionCount      uint64
						ShuttingDown      bool
						RelayVersion      string
						CPU               uint8
						EnvelopeUpKbps    uint64
						EnvelopeDownKbps  uint64
						BandwidthSentKbps uint64
						BandwidthRecvKbps uint64
					}
				*/

				numSamples := len(relayUpdate.PingStats)

				sampleRelayIds := make([]uint64, numSamples)
				sampleRTT := make([]float32, numSamples)
				sampleJitter := make([]float32, numSamples)
				samplePacketLoss := make([]float32, numSamples)

				for i := 0; i < numSamples; i++ {
					sampleRelayIds[i] = relayUpdate.PingStats[i].RelayID
					sampleRTT[i] = relayUpdate.PingStats[i].RTT
					sampleJitter[i] = relayUpdate.PingStats[i].Jitter
					samplePacketLoss[i] = relayUpdate.PingStats[i].PacketLoss
				}

				relayStats.ProcessRelayUpdate(relayId, 
					int(relayUpdate.SessionCount),
					relayUpdate.RelayVersion, 
					relayUpdate.ShuttingDown,
					numSamples,
					sampleRelayIds, 
					sampleRTT, 
					sampleJitter, 
					samplePacketLoss)
			}
		}
	}()
}

func UpdateRouteMatrix(service *common.Service, relayStats *common.RelayStats) {

	ticker := time.NewTicker(routeMatrixInterval)

	config := common.RedisSelectorConfig{}

	config.RedisHostname = redisHostname
	config.RedisPassword = redisPassword

	redisSelector, err := common.CreateRedisSelector(service.Context, config)
	if err != nil {
		core.Error("failed to create redis selector: %v", err)
		os.Exit(1)
	}

	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				timeStart := time.Now()

				// build relays data

				relaysDataNew := relayStats.GetRelaysCSV()

				// build the cost matrix

				relayData := service.RelayData()

				costs := relayStats.GetCosts(relayData.RelayIds, maxRTT, maxJitter, maxPacketLoss, service.Local)

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

				redisSelector.Store(service.Context, relaysDataNew, costMatrixDataNew, routeMatrixDataNew)

				// load the master cost and route matrix from redis (leader election)

				relaysDataNew, costMatrixDataNew, routeMatrixDataNew = redisSelector.Load(service.Context)

				if relaysDataNew == nil {
					core.Error("failed to get relays from redis selector")
					continue
				}

				if costMatrixDataNew == nil {
					core.Error("failed to get cost matrix from redis selector")
					continue
				}

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
