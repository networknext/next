/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"

	"github.com/networknext/backend/transport"
)

func main() {
	port := os.Getenv("NN_BACKEND_PORT")

	// is there a better way to test if a string is empty?
	if len(port) == 0 {
		port = "30000"
	}

	router := makeRouter()

	go optimizeRoutine()

	go timeoutRoutine()

	go listen(port, router)

	for {
	}
}

// TODO
func optimizeRoutine() {
	fmt.Println("TODO optimizeRoutine()")
}

// TODO
func timeoutRoutine() {
	fmt.Println("TODO timeoutRoutine()")
}

func makeRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/relay_init", transport.RelayInitHandlerFunc()).Methods("POST")
	router.HandleFunc("/relay_update", transport.RelayUpdateHandlerFunc()).Methods("POST")
	router.HandleFunc("/cost_matrix", transport.CostMatrixHandlerFunc()).Methods("GET")
	router.HandleFunc("/route_matrix", transport.RouteMatrixHandlerFunc()).Methods("GET")
	router.HandleFunc("/near", transport.NearHandlerFunc()).Methods("GET")
	return router
}

func listen(port string, router *mux.Router) {
	log.Printf("Starting server with port %s\n", port) // log
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), router)
	if err != nil {
		fmt.Println(err)
	}
}
