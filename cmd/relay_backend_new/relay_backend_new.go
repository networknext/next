package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
	"os"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
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

var costMatrixMutex sync.RWMutex
var costMatrixData []byte

var routeMatrixMutex sync.RWMutex
var routeMatrixData []byte

var costMatrixInternalMutex sync.RWMutex
var costMatrixInternalData []byte

var routeMatrixInternalMutex sync.RWMutex
var routeMatrixInternalData []byte

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

	core.Debug("max rtt: %.1f", maxRTT)
	core.Debug("max jitter: %.1f", maxJitter)
	core.Debug("max packet loss: %.1f", maxPacketLoss)
	core.Debug("cost matrix buffer size: %d bytes", costMatrixBufferSize)
	core.Debug("route matrix buffer size: %d bytes", routeMatrixBufferSize)
	core.Debug("route matrix interval: %s", routeMatrixInterval)
	core.Debug("redis hostname: %s", redisHostname)
	core.Debug("redis password: %s", redisPassword)
	core.Debug("redis pubsub channel name: %s", redisPubsubChannelName)
	core.Debug("relay update channel size: %d", relayUpdateChannelSize)

	service.LoadDatabase()

	relayStats := common.CreateRelayStats()

	// todo: override health function so we only become healthy once we have
	// our own internal cost/route matrix to serve up, plus cached data from
	// redis for the leader cost/route matrix

	service.Router.HandleFunc("/relay_data", relayDataHandler(service))
	service.Router.HandleFunc("/relay_stats", relayStatsHandler(service))
	service.Router.HandleFunc("/cost_matrix", costMatrixHandler)
	service.Router.HandleFunc("/route_matrix", routeMatrixHandler)
	service.Router.HandleFunc("/cost_matrix_internal", costMatrixInternalHandler)
	service.Router.HandleFunc("/route_matrix_internal", routeMatrixInternalHandler)

	service.StartWebServer()

	ProcessRelayUpdates(service.Context, relayStats)

	UpdateRouteMatrix(service, relayStats)

	service.WaitForShutdown()
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

func relayStatsHandler(service *common.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo: relay stats handler
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

func ProcessRelayUpdates(ctx context.Context, relayStats *common.RelayStats) {

	config := common.RedisPubsubConfig{}

	config.RedisHostname = redisHostname
	config.RedisPassword = redisPassword
	config.PubsubChannelName = redisPubsubChannelName
	config.MessageChannelSize = relayUpdateChannelSize

	consumer, err := common.CreateRedisPubsubConsumer(ctx, config)

	if err != nil {
		core.Error("could not create redis pubsub consumer")
		os.Exit(1)
	}

	go func() {

		for {
			select {

			case <-ctx.Done():
				return
			
			case message := <-consumer.MessageChannel:

				core.Debug("received relay update")

				// todo: parse relay update from message
				_ = message

				sourceRelayId := uint64(0)
				numSamples := 0
				var sampleRelayIds []uint64
				var sampleRTT []float32
				var sampleJitter []float32
				var samplePacketLoss []float32

				relayStats.ProcessRelayUpdate(sourceRelayId, numSamples, sampleRelayIds, sampleRTT, sampleJitter, samplePacketLoss)
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

				costMatrixDataNew, err := costMatrixNew.Write(costMatrixBufferSize)
				if err != nil {
					core.Error("could not write cost matrix: %v", err)
					continue
				}

				// serve up as internal cost matrix

				costMatrixInternalMutex.Lock()
				costMatrixInternalData = costMatrixDataNew
				costMatrixInternalMutex.Unlock()

				// todo: full relays

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

				routeMatrixDataNew, err := routeMatrixNew.Write(routeMatrixBufferSize)
				if err != nil {
					core.Error("could not write route matrix: %v", err)
					continue
				}

				// serve up as internal cost matrix

				routeMatrixInternalMutex.Lock()
				routeMatrixInternalData = routeMatrixDataNew
				routeMatrixInternalMutex.Unlock()

				// store our most recent cost and route matrix in redis

				redisSelector.Store(service.Context, costMatrixDataNew, routeMatrixDataNew)

				// load the master cost and route matrix from redis (leader election)

				costMatrixDataNew, routeMatrixDataNew = redisSelector.Load(service.Context)

				if costMatrixDataNew == nil {
					core.Error("failed to get cost matrix from redis selector")
					continue
				}

				if routeMatrixDataNew == nil {
					core.Error("failed to get route matrix from redis selector")
					continue
				}

				// serve up as official cost matrix and route matrix

				costMatrixMutex.Lock()
				costMatrixData = costMatrixDataNew
				costMatrixMutex.Unlock()

				routeMatrixMutex.Lock()
				routeMatrixData = routeMatrixDataNew
				routeMatrixMutex.Unlock()

				// we are done

				timeFinish := time.Now()

				optimizeDuration := timeFinish.Sub(timeStart)

				fmt.Printf("route optimization: %d relays in %s\n", relayData.NumRelays, optimizeDuration)
			}
		}
	}()
}
