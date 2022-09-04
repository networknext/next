package main

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/routing"
)

func main() {

	service := common.CreateService("relay_backend_new")

	maxRTT := float32(envvar.GetFloat("MAX_RTT", 1000.0))
	maxJitter := float32(envvar.GetFloat("MAX_JITTER", 1000.0))
	maxPacketLoss := float32(envvar.GetFloat("MAX_JITTER", 100.0))
	matrixBufferSize := envvar.GetInt("MATRIX_BUFFER_SIZE", 10*1024*1024)
	costMatrixInterval := envvar.GetDuration("COST_MATRIX_INTERVAL", time.Second)

	core.Debug("max rtt: %.1f", maxRTT)
	core.Debug("max jitter: %.1f", maxJitter)
	core.Debug("max packet loss: %.1f", maxPacketLoss)
	core.Debug("matrix buffer size: %d bytes", matrixBufferSize)
	core.Debug("cost matrix interval: %s", costMatrixInterval)

	service.LoadDatabase()

	relayStats := common.CreateRelayStats()

	service.StartWebServer()

	ProcessRelayUpdates(service.Context, relayStats)

	UpdateRouteMatrix(service, relayStats, maxRTT, maxJitter, maxPacketLoss, matrixBufferSize, costMatrixInterval)

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

func UpdateRouteMatrix(service *common.Service, relayStats *common.RelayStats, maxRTT float32, maxJitter float32, maxPacketLoss float32, matrixBufferSize int, costMatrixInterval time.Duration) {

	ticker := time.NewTicker(costMatrixInterval)
	
	go func() {
		for {
			select {
			
			case <-service.Context.Done():
				return

			case <-ticker.C:

				relayIds := service.RelayIds()

				numRelays := len(relayIds)

				core.Debug("%d relays", numRelays)

				costs := relayStats.GetCosts(relayIds, maxRTT, maxJitter, maxPacketLoss, service.Local)

				// todo: get relay addresses

				// todo: get relay names

				// todo: get relay latitides

				// todo: get relay longitudes

				// todo: get relay datacenter ids

				// todo: get dest relays

				relayAddresses := []net.UDPAddr{}
				relayNames := []string{}
				relayLatitudes := []float32{}
				relayLongitudes := []float32{}
				relayDatacenterIds := []uint64{}
				destRelays := make([]bool, numRelays)

				costMatrixNew := routing.CostMatrix{
					RelayIDs:           relayIds,
					RelayAddresses:     relayAddresses,
					RelayNames:         relayNames,
					RelayLatitudes:     relayLatitudes,
					RelayLongitudes:    relayLongitudes,
					RelayDatacenterIDs: relayDatacenterIds,
					Costs:              costs,
					Version:            routing.CostMatrixSerializeVersion,
					DestRelays:         destRelays,
				}

				_ = costMatrixNew

				core.Debug("updated route matrix: %d relays", len(relayIds))		// todo: print size in MB
			}
		}
	}()
}

func routeMatrixHandler(w http.ResponseWriter, r *http.Request) {

	// todo: serve up cached route matrix data
}
