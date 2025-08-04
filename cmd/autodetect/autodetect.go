package main

import (
	"net/http"
	
	"github.com/networknext/next/modules/common"
)

var magicUpdateSeconds int

func main() {

	service := common.CreateService("autodetect")

	service.Router.HandleFunc("/", autodetectHandler).Methods("GET")

	service.StartWebServer()

	service.WaitForShutdown()
}

func autodetectHandler(w http.ResponseWriter, r *http.Request) {

	// todo: grab mutex, look for cache

	// todo: run whois

	// todo: update cache

	w.Write([]byte("latitude.saopaulo"))
}
