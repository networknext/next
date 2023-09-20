package main

import (
	"time"
	"net/http"
	"io/ioutil"
	"encoding/binary"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"

	_ "github.com/wangjia184/sortedset"
)

type Session struct {
	sessionId	uint64
	score       int32
}

var sessionChannel chan Session

func main() {

	service := common.CreateService("session_cruncher")

	service.Router.HandleFunc("/session_batch", sessionBatchHandler).Methods("POST")
	service.Router.HandleFunc("/session_counts", sessionCountsHandler).Methods("GET")
	service.Router.HandleFunc("/sessions/{begin}/{end}", sessionsHandler).Methods("GET")

	go ProcessThread()

	go SortThread()

	go TestThread()

	service.StartWebServer()

	service.WaitForShutdown()
}

func TestThread() {
	for {
		session := Session{}
		session.timestamp = uint64(time.Now().Unix())
		session.sessionId = binary.LittleEndian.Uint64(body[index:index+8])
		session.score = int32(binary.LittleEndian.Uint32(body[index+8:index+12]))
		index += 12
		sessionChannel <- session
		time.Sleep(time.Second)		
	}
}

func ProcessThread() {
	for {
		select {
		case session := <-sessionChannel:
			core.Log("session %016x (%d)", session.sessionId, session.score)
			// todo
		}
	}
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
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		core.Error("could not read body")
		return
	}
	defer r.Body.Close()
	if len(body) % 12 != 0 {
		core.Error("session batch should be multiple of 12 bytes")
		return
	}
	numSessions := len(body) / 12
	index := 0
	currentTime := uint64(time.Now().Unix())
	session := Session{}
	for i := 0; i < numSessions; i++ {
		session.timestamp = currentTime
		session.sessionId = binary.LittleEndian.Uint64(body[index:index+8])
		session.score = int32(binary.LittleEndian.Uint32(body[index+8:index+12]))
		index += 12
		sessionChannel <- session
    }
}

func sessionCountsHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("session counts handler")
}

func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("sessions handler")
}
