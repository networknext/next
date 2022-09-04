package main

import (
	"context"
	// "net"
	"net/http"
	"time"
	"sync"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/envvar"
)

var maxRTT float32
var maxJitter float32
var maxPacketLoss float32
var costMatrixBufferSize int
var routeMatrixBufferSize int
var routeMatrixInterval time.Duration

var costMatrixMutex sync.RWMutex
var costMatrix *common.CostMatrix
var costMatrixData []byte

func main() {

	service := common.CreateService("relay_backend_new")

	maxRTT = float32(envvar.GetFloat("MAX_RTT", 1000.0))
	maxJitter = float32(envvar.GetFloat("MAX_JITTER", 1000.0))
	maxPacketLoss = float32(envvar.GetFloat("MAX_JITTER", 100.0))
	costMatrixBufferSize = envvar.GetInt("COST_MATRIX_BUFFER_SIZE", 1*1024*1024)
	routeMatrixBufferSize = envvar.GetInt("ROUTE_MATRIX_BUFFER_SIZE", 10*1024*1024)
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)

	core.Debug("max rtt: %.1f", maxRTT)
	core.Debug("max jitter: %.1f", maxJitter)
	core.Debug("max packet loss: %.1f", maxPacketLoss)
	core.Debug("cost matrix buffer size: %d bytes", costMatrixBufferSize)
	core.Debug("route matrix buffer size: %d bytes", routeMatrixBufferSize)
	core.Debug("route matrix interval: %s", routeMatrixInterval)

	service.LoadDatabase()

	relayStats := common.CreateRelayStats()

	// todo: override health function so we only become healthy once we have our own route matrix to serve up

	service.StartWebServer()

	ProcessRelayUpdates(service.Context, relayStats)

	UpdateRouteMatrix(service, relayStats)

	service.WaitForShutdown()
}

func ProcessRelayUpdates(ctx context.Context, relayStats *common.RelayStats) {

	// todo: setup redis pubsub consumer
	
	go func() {
		for {
			// todo: not sure this is the best way to exit on the context...
			select {
			case <-ctx.Done():
				return
			default:
			}

			// todo: get relay update message from redis pubsub consumer
			message := make([]byte, 100)

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
	}()
}

func UpdateRouteMatrix(service *common.Service, relayStats *common.RelayStats) {

	ticker := time.NewTicker(routeMatrixInterval)
	
	go func() {
		for {
			select {
			
			case <-service.Context.Done():
				return

			case <-ticker.C:

				relayData := service.RelayData()

				core.Debug("%d relays", relayData.NumRelays)

				costs := relayStats.GetCosts(relayData.RelayIds, maxRTT, maxJitter, maxPacketLoss, service.Local)

				costMatrixNew := common.CostMatrix{
					Version:            common.CostMatrixSerializeVersion,
					RelayIDs:           relayData.RelayIds,
					RelayAddresses:     relayData.RelayAddresses,
					RelayNames:         relayData.RelayNames,
					RelayLatitudes:     relayData.RelayLatitudes,
					RelayLongitudes:    relayData.RelayLongitudes,
					RelayDatacenterIDs: relayData.RelayDatacenterIds,
					DestRelays:         relayData.DestRelays,
					Costs:              costs,
				}

				costMatrixDataNew, err := costMatrixNew.Write(costMatrixBufferSize); 
				if err != nil {
					core.Error("could not write cost matrix: %v", err)
					continue
				}

				core.Debug("updated cost matrix: %d bytes (%d relays)", len(costMatrixDataNew), relayData.NumRelays)
			}
		}
	}()
}

func routeMatrixHandler(w http.ResponseWriter, r *http.Request) {

	// todo: serve up cached route matrix data
}
