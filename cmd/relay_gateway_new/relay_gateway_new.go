package main

import (
	"net/http"
	// "fmt"
	// "time"
	// "context"

	"github.com/networknext/backend/modules/common"
	// "github.com/networknext/backend/modules/envvar"
)

func relayUpdateHandler(w http.ResponseWriter, r *http.Request) {

	// ...
}

func main() {

	service := common.CreateService("relay_gateway_new")

	service.Router.HandleFunc("/relay_update", relayUpdateHandler).Methods("POST")

	service.LoadDatabase()

	service.StartWebServer()

	service.WaitForShutdown()
}
