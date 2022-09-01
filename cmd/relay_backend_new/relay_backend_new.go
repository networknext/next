package main

import (
	"net/http"
	"fmt"
	"time"
	"context"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/envvar"
)

func Update() {
	fmt.Printf("update\n")
}

// todo: make this an update loop request from the service, with a function to call to say done

func UpdateLoop(ctx context.Context) {

	syncInterval := envvar.GetDuration("COST_MATRIX_INTERVAL", time.Second)

	matrixBufferSize := envvar.GetInt("MATRIX_BUFFER_SIZE", 100000)

	ticker := time.NewTicker(syncInterval)

	// todo: do stuff

	for {
		select {

			case <-ctx.Done():
				return

			case <-ticker.C:
				Update()
		}
	}

	_ = matrixBufferSize
}

func routeMatrixHandler(w http.ResponseWriter, r *http.Request) {

	// ...
}

func main() {

	service := common.CreateService("relay_backend_new")

	go UpdateLoop(service.Context)

	service.LoadDatabase()

	service.StartWebServer()

	service.WaitForShutdown()
}
