package main

import (
	"time"
	"net/http"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"

	_ "github.com/wangjia184/sortedset"
)

func main() {

	service := common.CreateService("session_cruncher")

	service.Router.HandleFunc("/session_batch", sessionBatchHandler).Methods("POST")
	service.Router.HandleFunc("/session_counts", sessionCountsHandler).Methods("GET")
	service.Router.HandleFunc("/sessions/{begin}/{end}", sessionsHandler).Methods("GET")

	service.StartWebServer()

	service.WaitForShutdown()
}

func SortThread() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			core.Log("sort")
		}
	}
}

func sessionBatchHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("session batch handler")
}

func sessionCountsHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("session counts handler")
}

func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("sessions handler")
}
