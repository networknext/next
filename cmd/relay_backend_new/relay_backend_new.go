package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/envvar"
)

func ProcessRelayUpdate(body []byte) {
	// todo: process relay update	
}

func ProcessRelayUpdates(ctx context.Context) {

	// todo: setup redis pubsub consumer
	
	go func() {
		for {
			// todo: not sure this is the best way to exit on the context...
			select {
			case <-ctx.Done():
				return
			default:
			}

			// todo: process redis pubsub messages, insert into relay database
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

	service.StartWebServer()

	ProcessRelayUpdates(service.Context)

	UpdateRouteMatrix(service.Context)

	service.WaitForShutdown()
}
