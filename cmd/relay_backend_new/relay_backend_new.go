package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/envvar"
)

func ProcessRelayUpdate(relayStats *common.RelayStats, message []byte) {

	// todo: parse relay update from message

	// todo: process relay update

	sourceRelayId := uint64(0)
	numSamples := 0
	var sampleRelayIds []uint64
	var sampleRTT []float32
	var sampleJitter []float32
	var samplePacketLoss []float32

	relayStats.ProcessRelayUpdate(sourceRelayId, numSamples, sampleRelayIds, sampleRTT, sampleJitter, samplePacketLoss)
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
			ProcessRelayUpdate(relayStats, message)
		}
	}()
}

func UpdateRouteMatrix(ctx context.Context) {

	syncInterval := envvar.GetDuration("COST_MATRIX_INTERVAL", time.Second)

	matrixBufferSize := envvar.GetInt("MATRIX_BUFFER_SIZE", 100000)

	ticker := time.NewTicker(syncInterval)

	go func() {
		for {
			select {

			case <-ctx.Done():
				return

			case <-ticker.C:
				fmt.Printf("update route matrix\n")
				_ = matrixBufferSize
			}
		}
	}()
}

func routeMatrixHandler(w http.ResponseWriter, r *http.Request) {

	// todo: serve up cached route matrix data
}

func main() {

	service := common.CreateService("relay_backend_new")

	service.LoadDatabase()

	relayStats := common.CreateRelayStats()

	service.StartWebServer()

	ProcessRelayUpdates(service.Context, relayStats)

	UpdateRouteMatrix(service.Context)

	service.WaitForShutdown()
}
