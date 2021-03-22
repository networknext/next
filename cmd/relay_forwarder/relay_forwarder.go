package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/networknext/backend/modules/common/helpers"

	"github.com/networknext/backend/modules/transport"

	"github.com/gorilla/mux"
	"github.com/networknext/backend/modules/envvar"
)

func main() {
	lbAddr := envvar.Get("GATEWAY_LB", "")
	ip := net.ParseIP(lbAddr)
	if ip == nil {
		log.Fatal("msg", "bad gateway address")
	}

	initURI := fmt.Sprintf("http://%s/relay_init", lbAddr)
	updateURI := fmt.Sprintf("http://%s/relay_update", lbAddr)

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/relay_init", forwardPost(initURI)).Methods("POST")
	router.HandleFunc("/relay_update", forwardPost(updateURI)).Methods("POST")

	go func() {
		port := envvar.Get("PORT", "30000")

		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			log.Fatal(err)
			os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
		}
	}()

	go func() {
		syncTimer := helpers.NewSyncTimer(10 * time.Second)
		for {
			syncTimer.Run()
			fmt.Printf("number of go routines %v", runtime.NumGoroutine())
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}

func forwardGet(address string, octet bool) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(address)
		if err != nil {
			log.Printf("error forwarding get: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error reading response body: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		if octet {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		w.Write(body)
	}
}

func forwardPost(address string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("post started")

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("error reading response body: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body.Close()

		reqBuf := bytes.NewBuffer(reqBody)
		resp, err := http.Post(address, r.Header.Get("Content-Type"), reqBuf)
		if err != nil {
			fmt.Printf("error forwarding get: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error reading response body: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp.Body.Close()

		w.WriteHeader(resp.StatusCode)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Write(respBody)

		fmt.Println("post finished")
	}
}
