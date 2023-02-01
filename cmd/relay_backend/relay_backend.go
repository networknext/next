package main

import (
	"bytes"
	"encoding/json"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/messages"
	"github.com/networknext/backend/modules/packets"
)

var maxRTT float32
var maxJitter float32
var maxPacketLoss float32
var costMatrixBufferSize int
var routeMatrixBufferSize int
var routeMatrixInterval time.Duration
var redisHostName string
var redisPassword string
var redisPubsubChannelName string
var relayUpdateChannelSize int
var pingStatsPubsubTopic string
var maxPingStatsChannelSize int
var maxPingStatsMessageBytes int
var relayStatsPubsubTopic string
var maxRelayStatsChannelSize int
var maxRelayStatsMessageBytes int
var disableGooglePubsub bool
var readyDelay time.Duration

var relaysMutex sync.RWMutex
var relaysCSVData []byte

var costMatrixMutex sync.RWMutex
var costMatrixData []byte

var routeMatrixMutex sync.RWMutex
var routeMatrixData []byte

var costMatrixInternalMutex sync.RWMutex
var costMatrixInternalData []byte

var routeMatrixInternalMutex sync.RWMutex
var routeMatrixInternalData []byte

var readyMutex sync.RWMutex
var ready bool

var startTime time.Time

var counterNames [common.NumRelayCounters]string

func main() {

	service := common.CreateService("relay_backend")

	maxRTT = float32(envvar.GetFloat("MAX_RTT", 1000.0))
	maxJitter = float32(envvar.GetFloat("MAX_JITTER", 1000.0))
	maxPacketLoss = float32(envvar.GetFloat("MAX_PACKET_LOSS", 100.0))
	costMatrixBufferSize = envvar.GetInt("COST_MATRIX_BUFFER_SIZE", 10*1024*1024)
	routeMatrixBufferSize = envvar.GetInt("ROUTE_MATRIX_BUFFER_SIZE", 100*1024*1024)
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)
	redisPubsubChannelName = envvar.GetString("REDIS_PUBSUB_CHANNEL_NAME", "relay_updates")
	relayUpdateChannelSize = envvar.GetInt("RELAY_UPDATE_CHANNEL_SIZE", 10*1024)
	pingStatsPubsubTopic = envvar.GetString("PING_STATS_PUBSUB_TOPIC", "ping_stats")
	maxPingStatsChannelSize = envvar.GetInt("MAX_PING_STATS_CHANNEL_SIZE", 10*1024)
	maxPingStatsMessageBytes = envvar.GetInt("MAX_PING_STATS_MESSAGE_BYTES", 1024)
	relayStatsPubsubTopic = envvar.GetString("RELAY_STATS_PUBSUB_TOPIC", "relay_stats")
	maxRelayStatsChannelSize = envvar.GetInt("MAX_RELAY_STATS_CHANNEL_SIZE", 10*1024)
	maxRelayStatsMessageBytes = envvar.GetInt("MAX_RELAY_STATS_MESSAGE_BYTES", 1024)
	redisHostName = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")
	disableGooglePubsub = envvar.GetBool("DISABLE_GOOGLE_PUBSUB", false)
	readyDelay = envvar.GetDuration("READY_DELAY", 1*time.Second)
	startTime = time.Now()

	core.Log("max rtt: %.1f", maxRTT)
	core.Log("max jitter: %.1f", maxJitter)
	core.Log("max packet loss: %.1f", maxPacketLoss)
	core.Log("cost matrix buffer size: %d bytes", costMatrixBufferSize)
	core.Log("route matrix buffer size: %d bytes", routeMatrixBufferSize)
	core.Log("route matrix interval: %s", routeMatrixInterval)
	core.Log("redis host name: %s", redisHostName)
	core.Log("redis password: %s", redisPassword)
	core.Log("redis pubsub channel name: %s", redisPubsubChannelName)
	core.Log("relay update channel size: %d", relayUpdateChannelSize)
	core.Log("ping stats pubsub channel: %s", pingStatsPubsubTopic)
	core.Log("max ping stats channel size: %d", maxPingStatsChannelSize)
	core.Log("max ping stats message size: %d", maxPingStatsMessageBytes)
	core.Log("relay stats pubsub channel: %s", relayStatsPubsubTopic)
	core.Log("max relay stats channel: %d", maxRelayStatsChannelSize)
	core.Log("max relay stats message size: %d", maxRelayStatsMessageBytes)
	core.Log("disable google pubsub: %v", disableGooglePubsub)
	core.Log("ready delay: %s", readyDelay.String())
	core.Log("start time: %s", startTime.String())

	service.LoadDatabase()

	initCounterNames()

	relayManager := common.CreateRelayManager()

	service.Router.HandleFunc("/relays", relaysHandler)
	service.Router.HandleFunc("/relay_data", relayDataHandler(service))
	service.Router.HandleFunc("/cost_matrix", costMatrixHandler)
	service.Router.HandleFunc("/route_matrix", routeMatrixHandler)
	service.Router.HandleFunc("/cost_matrix_internal", costMatrixInternalHandler)
	service.Router.HandleFunc("/route_matrix_internal", routeMatrixInternalHandler)
    service.Router.HandleFunc("/relay_counters/{relay_name}", relayCountersHandler(service, relayManager))
    service.Router.HandleFunc("/cost_matrix_html", costMatrixHtmlHandler(service, relayManager))
    service.Router.HandleFunc("/routes/{src}/{dest}", routesHandler(service, relayManager))
    service.Router.HandleFunc("/relay_manager", relayManagerHandler(service, relayManager))

	service.SetHealthFunctions(sendTrafficToMe(service), machineIsHealthy)

	service.StartWebServer()

	service.LeaderElection(false)

	ProcessRelayUpdates(service, relayManager)

	UpdateRouteMatrix(service, relayManager)

	UpdateReadyState(service)

	service.WaitForShutdown()
}

func relayManagerHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		copy := relayManager.Copy()
		var buffer bytes.Buffer
		err := gob.NewEncoder(&buffer).Encode(copy)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "no relay manager: %v\n", err)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		_, err = buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func routesHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		src := vars["src"]
		dest := vars["dest"]
		routeMatrixMutex.RLock()
		data := routeMatrixData
		routeMatrixMutex.RUnlock()
		routeMatrix := common.RouteMatrix{}
		err := routeMatrix.Read(data)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "no route matrix\n")
			return
		}
		fmt.Printf("relay names: %+v\n", routeMatrix.RelayNames)
		src_index := -1
		for i := range routeMatrix.RelayNames {
			if routeMatrix.RelayNames[i] == src {
				src_index = i
				break
			}
		}
		dest_index := -1
		for i := range routeMatrix.RelayNames {
			if routeMatrix.RelayNames[i] == dest {
				dest_index = i
				break
			}
		}
		fmt.Printf("%s %s %d %d\n", src, dest, src_index, dest_index)
		if src_index == -1 || dest_index == -1 || src_index == dest_index {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		const htmlHeader = `<!DOCTYPE html>
		<html lang="en">
		<head>
		  <meta charset="utf-8">
		  <title>Routes</title>
		  <style>
			table, th, td {
		      border: 1px solid black;
		      border-collapse: collapse;
		      text-align: center;
		      padding: 10px;
		    }
			*{
		    font-family:Courier;
		  }	  
		  </style>
		</head>
		<body>`
		fmt.Fprintf(w, "%s\n", htmlHeader)
		fmt.Fprintf(w, "route matrix: %s - %s<br><br>\n", src, dest)
		fmt.Fprintf(w, "<table>\n")
		fmt.Fprintf(w, "<tr><td><b>Route Cost</b></td><td><b>Route Hash</b></td><td><b>Route Relays</b></td></tr>\n")
		index := core.TriMatrixIndex(src_index, dest_index)
		entry := routeMatrix.RouteEntries[index]
		for i := 0; i < int(entry.NumRoutes); i++ {
			routeRelays := ""
			numRouteRelays := int(entry.RouteNumRelays[i])
			for j := 0; j < numRouteRelays; j++ {
				routeRelayIndex := entry.RouteRelays[i][j]
				routeRelayName := routeMatrix.RelayNames[routeRelayIndex]
				routeRelays += routeRelayName
				if j != numRouteRelays-1 {
					routeRelays += " - "
				}
			}
			fmt.Fprintf(w, "<tr><td>%d</td><td>%0x</td><td>%s</td></tr>", entry.RouteCost[i], entry.RouteHash[i], routeRelays)
		}
		fmt.Fprintf(w, "<tr><td>%d</td><td></td><td>%s</td></tr>", entry.DirectCost, "direct")
		fmt.Fprintf(w, "</table>\n")
	  const htmlFooter = `</body></html>`
		fmt.Fprintf(w, "%s\n", htmlFooter)
	}
}

func costMatrixHtmlHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		costMatrixMutex.RLock()
		data := costMatrixData
		costMatrixMutex.RUnlock()
		costMatrix := common.CostMatrix{}
		err := costMatrix.Read(data)
		if err != nil {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "no cost matrix\n")
			return
		}
		w.Header().Set("Content-Type", "text/html")
		const htmlHeader = `<!DOCTYPE html>
		<html lang="en">
		<head>
		  <meta charset="utf-8">
		  <title>Cost Matrix</title>
		  <style>
			table, th, td {
		      border: 1px solid black;
		      border-collapse: collapse;
		      text-align: center;
		      padding: 10px;
		    }
			cost{
         	  color: white;
      		}
			*{
		    font-family:Courier;
		    }	  
		  </style>
		</head>
		<body>`
		fmt.Fprintf(w, "%s\n", htmlHeader)
		fmt.Fprintf(w, "cost matrix:<br><br><table>\n")
		fmt.Fprintf(w, "<tr><td></td>")
		for i := range costMatrix.RelayNames {
			fmt.Fprintf(w, "<td><b>%s</b></td>", costMatrix.RelayNames[i])
		}
		fmt.Fprintf(w, "</tr>\n")
		for i := range costMatrix.RelayNames {
			fmt.Fprintf(w, "<tr><td><b>%s</b></td>", costMatrix.RelayNames[i])
			for j := range costMatrix.RelayNames {
				if i == j {
					fmt.Fprint(w, "<td bgcolor=\"lightgrey\"></td>")
					continue
				}
				nope := false
				costString := ""
				index := core.TriMatrixIndex(i,j)
				cost := costMatrix.Costs[index]
				if cost >= 0 {
					costString = fmt.Sprintf("%d", cost)
				} else {
					nope = true
				}
				clickable := fmt.Sprintf("class=\"clickable\" onclick=\"window.location='/routes/%s/%s'\"", costMatrix.RelayNames[i], costMatrix.RelayNames[j])

				if nope {
					fmt.Fprintf(w, "<td %s bgcolor=\"red\"></td>", clickable)
				} else {
					fmt.Fprintf(w, "<td %s bgcolor=\"green\"><cost>%s</cost></td>", clickable, costString)
				}
			}
			fmt.Fprintf(w, "</tr>\n")
		}
		fmt.Fprintf(w, "</table>\n")
	  const htmlFooter = `</body></html>`
		fmt.Fprintf(w, "%s\n", htmlFooter)
	}
}

func initCounterNames() {
	counterNames[0] = "RELAY_PING_PACKET_RECEIVED"
	counterNames[1] = "RELAY_PONG_PACKET_RECEIVED"
	counterNames[2] = "BASIC_PACKET_FILTER_DROPPED_PACKET"
	counterNames[3] = "ADVANCED_PACKET_FILTER_DROPPED_PACKET"
	counterNames[4] = "ROUTE_REQUEST_PACKET_RECEIVED"
	counterNames[5] = "ROUTE_REQUEST_PACKET_BAD_SIZE"
	counterNames[6] = "ROUTE_REQUEST_PACKET_COULD_NOT_READ_TOKEN"
	counterNames[7] = "ROUTE_REQUEST_PACKET_TOKEN_EXPIRED"
	counterNames[8] = "SESSION_CREATED"
	counterNames[9] = "ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[10] = "ROUTE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[11] = "ROUTE_RESPONSE_PACKET_RECEIVED"
	counterNames[12] = "ROUTE_RESPONSE_PACKET_BAD_SIZE"
	counterNames[13] = "ROUTE_RESPONSE_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[14] = "ROUTE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[15] = "ROUTE_RESPONSE_PACKET_SESSION_EXPIRED"
	counterNames[16] = "ROUTE_RESPONSE_PACKET_ALREADY_RECEIVED"
	counterNames[17] = "ROUTE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[18] = "ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[19] = "ROUTE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[20] = "CONTINUE_REQUEST_PACKET_RECEIVED"
	counterNames[21] = "CONTINUE_REQUEST_PACKET_BAD_SIZE"
	counterNames[22] = "CONTINUE_REQUEST_PACKET_COULD_NOT_READ_TOKEN"
	counterNames[23] = "CONTINUE_REQUEST_PACKET_TOKEN_EXPIRED"
	counterNames[24] = "CONTINUE_REQUEST_PACKET_SESSION_EXPIRED"
	counterNames[25] = "SESSION_CONTINUED"
	counterNames[26] = "CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[27] = "CONTINUE_REQUEST_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[28] = "CONTINUE_RESPONSE_PACKET_RECEIVED"
	counterNames[29] = "CONTINUE_RESPONSE_PACKET_BAD_SIZE"
	counterNames[30] = "CONTINUE_RESPONSE_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[31] = "CONTINUE_RESPONSE_PACKET_ALREADY_RECEIVED"
	counterNames[32] = "CONTINUE_RESPONSE_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[33] = "CONTINUE_RESPONSE_PACKET_SESSION_EXPIRED"
	counterNames[34] = "CONTINUE_RESPONSE_PACKET_HEADER_DID_NOT_VERIFY"
	counterNames[35] = "CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[36] = "CONTINUE_RESPONSE_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[37] = "CLIENT_TO_SERVER_PACKET_RECEIVED"
	counterNames[38] = "CLIENT_TO_SERVER_PACKET_TOO_SMALL"
	counterNames[39] = "CLIENT_TO_SERVER_PACKET_TOO_BIG"
	counterNames[40] = "CLIENT_TO_SERVER_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[41] = "CLIENT_TO_SERVER_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[42] = "CLIENT_TO_SERVER_PACKET_SESSION_EXPIRED"
	counterNames[43] = "CLIENT_TO_SERVER_PACKET_ALREADY_RECEIVED"
	counterNames[44] = "CLIENT_TO_SERVER_PACKET_COULD_NOT_VERIFY_HEADER"
	counterNames[45] = "CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[46] = "CLIENT_TO_SERVER_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[47] = "SERVER_TO_CLIENT_PACKET_RECEIVED"
	counterNames[48] = "SERVER_TO_CLIENT_PACKET_TOO_SMALL"
	counterNames[49] = "SERVER_TO_CLIENT_PACKET_TOO_BIG"
	counterNames[50] = "SERVER_TO_CLIENT_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[51] = "SERVER_TO_CLIENT_PACKET_COULD_NOT_FIND_SESSION"
	counterNames[52] = "SERVER_TO_CLIENT_PACKET_SESSION_EXPIRED"
	counterNames[53] = "SERVER_TO_CLIENT_PACKET_ALREADY_RECEIVED"
	counterNames[54] = "SERVER_TO_CLIENT_PACKET_COULD_NOT_VERIFY_HEADER"
	counterNames[55] = "SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[56] = "SERVER_TO_CLIENT_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[57] = "SESSION_PING_PACKET_RECEIVED"
	counterNames[58] = "SESSION_PING_PACKET_BAD_PACKET_SIZE"
	counterNames[59] = "SESSION_PING_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[60] = "SESSION_PING_PACKET_SESSION_DOES_NOT_EXIST"
	counterNames[61] = "SESSION_PING_PACKET_SESSION_EXPIRED"
	counterNames[62] = "SESSION_PING_PACKET_ALREADY_RECEIVED"
	counterNames[63] = "SESSION_PING_PACKET_COULD_NOT_VERIFY_HEADER"
	counterNames[64] = "SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP_PUBLIC_ADDRESS"
	counterNames[65] = "SESSION_PING_PACKET_FORWARD_TO_NEXT_HOP_INTERNAL_ADDRESS"
	counterNames[66] = "SESSION_PONG_PACKET_RECEIVED"
	counterNames[67] = "SESSION_PONG_PACKET_BAD_SIZE"
	counterNames[68] = "SESSION_PONG_PACKET_COULD_NOT_PEEK_HEADER"
	counterNames[69] = "SESSION_PONG_PACKET_SESSION_DOES_NOT_EXIST"
	counterNames[70] = "SESSION_PONG_PACKET_SESSION_EXPIRED"
	counterNames[71] = "SESSION_PONG_PACKET_ALREADY_RECEIVED"
	counterNames[72] = "SESSION_PONG_PACKET_COULD_NOT_VERIFY_HEADER"
	counterNames[73] = "SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP_PUBLIC_ADDRESS"
	counterNames[74] = "SESSION_PONG_PACKET_FORWARD_TO_PREVIOUS_HOP_INTERNAL_ADDRESS"
	counterNames[75] = "NEAR_PING_PACKET_RECEIVED"
	counterNames[76] = "NEAR_PING_PACKET_BAD_SIZE"
	counterNames[77] = "NEAR_PING_PACKET_RESPONDED_WITH_PONG"
	counterNames[78] = "RELAY_PING_PACKET_SENT"
}

func relayCountersHandler(service *common.Service, relayManager *common.RelayManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		relayName := vars["relay_name"]
		relayData := service.RelayData()
		relayIndex := -1
		for i := range relayData.RelayNames {
			if relayData.RelayNames[i] == relayName {
				relayIndex = i
				break
			}
		}
		if relayIndex == -1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		relayId := relayData.RelayIds[relayIndex]
		counters := relayManager.GetRelayCounters(relayId)
		w.Header().Set("Content-Type", "text/html")
		const htmlHeader = `<!DOCTYPE html>
		<html lang="en">
		<head>
		  <meta charset="utf-8">
		  <meta http-equiv="refresh" content="1">
		  <title>Relay Counters</title>
		  <style>
			table, th, td {
		      border: 1px solid black;
		      border-collapse: collapse;
		      text-align: center;
		      padding: 10px;
		    }
			*{
		    font-family:Courier;
		  }	  
		  </style>
		</head>
		<body>`
		fmt.Fprintf(w, "%s<br><br>\n", htmlHeader)
		fmt.Fprintf(w, "%s\n\n<table>\n", relayName)
		for i := range counterNames {
			if counterNames[i] == "" {
				continue
			}
			fmt.Fprintf(w, "<tr><td>%s</td><td>%d</td></tr>\n", counterNames[i], counters[i])
		}
		fmt.Fprintf(w, "</table>\n")
	  const htmlFooter = `</body></html>`
		fmt.Fprintf(w, "%s\n", htmlFooter)
	}
}

func sendTrafficToMe(service *common.Service) func() bool {
	return func() bool {
		routeMatrixMutex.RLock()
		hasRouteMatrix := routeMatrixData != nil
		routeMatrixMutex.RUnlock()
		return isReady() && hasRouteMatrix
	}
}

func machineIsHealthy() bool {
	return true
}

func isReady() bool {
	readyMutex.RLock()
	result := ready
	readyMutex.RUnlock()
	return result
}

func UpdateReadyState(service *common.Service) {
	go func() {
		for {
			time.Sleep(time.Second)
			delayReady := time.Since(startTime) >= readyDelay
			if delayReady {
				core.Log("relay backend is ready")
				readyMutex.Lock()
				ready = true
				readyMutex.Unlock()
				break
			}
		}
	}()
}

func relaysHandler(w http.ResponseWriter, r *http.Request) {
	relaysMutex.RLock()
	responseData := relaysCSVData
	relaysMutex.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	buffer := bytes.NewBuffer(responseData)
	_, err := buffer.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type RelayJSON struct {
	RelayIds           []string  `json:"relay_ids"`
	RelayNames         []string  `json:"relay_names"`
	RelayAddresses     []string  `json:"relay_addresses"`
	RelayLatitudes     []float32 `json:"relay_latitudes"`
	RelayLongitudes    []float32 `json:"relay_longitudes"`
	RelayDatacenterIds []string  `json:"relay_datacenter_ids"`
	RelayIdToIndex     []string  `json:"relay_id_to_index"`
	DestRelays         []string  `json:"dest_relays"`
	DestRelayNames     []string  `json:"dest_relay_names"`
}

func relayDataHandler(service *common.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		relayData := service.RelayData()
		relayJSON := RelayJSON{}
		relayJSON.RelayIds = make([]string, relayData.NumRelays)
		relayJSON.RelayNames = relayData.RelayNames
		relayJSON.RelayAddresses = make([]string, relayData.NumRelays)
		relayJSON.RelayDatacenterIds = make([]string, relayData.NumRelays)
		relayJSON.RelayLatitudes = relayData.RelayLatitudes
		relayJSON.RelayLongitudes = relayData.RelayLongitudes
		relayJSON.RelayIdToIndex = make([]string, relayData.NumRelays)
		for i := 0; i < relayData.NumRelays; i++ {
			relayJSON.RelayIdToIndex[i] = fmt.Sprintf("%016x - %d", relayData.RelayIds[i], i)
		}
		relayJSON.DestRelays = make([]string, relayData.NumRelays)
		for i := 0; i < relayData.NumRelays; i++ {
			relayJSON.RelayIds[i] = fmt.Sprintf("%016x", relayData.RelayIds[i])
			relayJSON.RelayAddresses[i] = relayData.RelayAddresses[i].String()
			relayJSON.RelayDatacenterIds[i] = fmt.Sprintf("%016x", relayData.RelayDatacenterIds[i])
			if relayData.DestRelays[i] {
				relayJSON.DestRelays[i] = "1"
			} else {
				relayJSON.DestRelays[i] = "0"
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(relayJSON); err != nil {
			core.Error("could not write relay data json: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func costMatrixHandler(w http.ResponseWriter, r *http.Request) {
	costMatrixMutex.RLock()
	responseData := costMatrixData
	costMatrixMutex.RUnlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	buffer := bytes.NewBuffer(responseData)
	_, err := buffer.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func routeMatrixHandler(w http.ResponseWriter, r *http.Request) {
	routeMatrixMutex.RLock()
	responseData := routeMatrixData
	routeMatrixMutex.RUnlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	buffer := bytes.NewBuffer(responseData)
	_, err := buffer.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func costMatrixInternalHandler(w http.ResponseWriter, r *http.Request) {
	costMatrixInternalMutex.RLock()
	responseData := costMatrixInternalData
	costMatrixInternalMutex.RUnlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	buffer := bytes.NewBuffer(responseData)
	_, err := buffer.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func routeMatrixInternalHandler(w http.ResponseWriter, r *http.Request) {
	routeMatrixInternalMutex.RLock()
	responseData := routeMatrixInternalData
	routeMatrixInternalMutex.RUnlock()
	w.Header().Set("Content-Type", "application/octet-stream")
	buffer := bytes.NewBuffer(responseData)
	_, err := buffer.WriteTo(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func ProcessRelayUpdates(service *common.Service, relayManager *common.RelayManager) {

	config := common.RedisPubsubConfig{}

	config.RedisHostname = redisHostName
	config.RedisPassword = redisPassword
	config.PubsubChannelName = redisPubsubChannelName
	config.MessageChannelSize = relayUpdateChannelSize

	consumer, err := common.CreateRedisPubsubConsumer(service.Context, config)

	if err != nil {
		core.Error("could not create redis pubsub consumer")
		os.Exit(1)
	}

	var pingStatsProducer *common.GooglePubsubProducer
	var relayStatsProducer *common.GooglePubsubProducer

	if !disableGooglePubsub {

		pingStatsProducer, err = common.CreateGooglePubsubProducer(service.Context, common.GooglePubsubConfig{
			ProjectId:          service.GoogleProjectId,
			Topic:              pingStatsPubsubTopic,
			MessageChannelSize: maxPingStatsChannelSize,
		})
		if err != nil {
			core.Error("could not create ping stats producer")
			os.Exit(1)
		}

		relayStatsProducer, err = common.CreateGooglePubsubProducer(service.Context, common.GooglePubsubConfig{
			ProjectId:          service.GoogleProjectId,
			Topic:              relayStatsPubsubTopic,
			MessageChannelSize: maxRelayStatsChannelSize,
		})
		if err != nil {
			core.Error("could not create relay stats producer")
			os.Exit(1)
		}
	}

	go func() {

		for {
			select {

			case <-service.Context.Done():
				return

			case message := <-consumer.MessageChannel:

				// read the relay update request packet

				var relayUpdateRequest packets.RelayUpdateRequestPacket

				err = relayUpdateRequest.Read(message)
				if err != nil {
					core.Error("could not read relay update: %v", err)
					return
				}

				if relayUpdateRequest.Version != packets.VersionNumberRelayUpdateRequest {
					core.Error("relay update version mismatch")
					return
				}

				// look up the relay in the database

				relayData := service.RelayData()

				relayId := common.RelayId(relayUpdateRequest.Address.String())
				relayIndex, ok := relayData.RelayIdToIndex[relayId]
				if !ok {
					core.Error("unknown relay id %016x", relayId)
					return
				}

				relayName := relayData.RelayNames[relayIndex]
				relayAddress := relayData.RelayAddresses[relayIndex].String()

				// debug

				core.Debug("[%s] received update for %s [%x]", relayAddress, relayName, relayId)

				numSamples := int(relayUpdateRequest.NumSamples)

				for i := 0; i < numSamples; i++ {
					rtt := relayUpdateRequest.SampleRTT[i]
					jitter := relayUpdateRequest.SampleJitter[i]
					pl := relayUpdateRequest.SamplePacketLoss[i]
					id := relayUpdateRequest.SampleRelayId[i]
					index, ok := relayData.RelayIdToIndex[id]
					if !ok {
						continue
					}
					name := relayData.RelayNames[index]
					if rtt < 1000.0 && pl < 100.0 {
						core.Debug("[%s] %s -> %s: rtt = %.1f, jitter = %.1f, pl = %.2f%%", relayAddress, relayName, name, rtt, jitter, pl)
					}
				}

				// process samples in the relay update

				currentTime := time.Now().Unix()

				relayManager.ProcessRelayUpdate(currentTime,
					relayId,
					relayName,
					relayUpdateRequest.Address,
					int(relayUpdateRequest.SessionCount),
					relayUpdateRequest.RelayVersion,
					relayUpdateRequest.ShuttingDown,
					numSamples,
					relayUpdateRequest.SampleRelayId[:numSamples],
					relayUpdateRequest.SampleRTT[:numSamples],
					relayUpdateRequest.SampleJitter[:numSamples],
					relayUpdateRequest.SamplePacketLoss[:numSamples],
					relayUpdateRequest.Counters[:],
				)

				if disableGooglePubsub {
					break
				}

				// build ping stats message

				numRoutable := 0

				pingStatsMessages := make([]messages.PingStatsMessage, 0)
				sampleRelayIds := make([]uint64, numSamples)
				sampleRTT := make([]float32, numSamples)
				sampleJitter := make([]float32, numSamples)
				samplePacketLoss := make([]float32, numSamples)
				sampleRoutable := make([]bool, numSamples)

				for i := 0; i < numSamples; i++ {

					rtt := relayUpdateRequest.SampleRTT[i]
					jitter := relayUpdateRequest.SampleJitter[i]
					pl := relayUpdateRequest.SamplePacketLoss[i]

					sampleRelayId := relayUpdateRequest.SampleRelayId[i]

					sampleRelayIds[i] = sampleRelayId
					sampleRTT[i] = rtt
					sampleJitter[i] = jitter
					samplePacketLoss[i] = pl

					if rtt <= maxRTT && jitter <= maxJitter && pl <= maxPacketLoss {
						numRoutable++
						sampleRoutable[i] = true
					}

					pingStatsMessages = append(pingStatsMessages, messages.PingStatsMessage{
						Version:    messages.PingStatsMessageVersion_Write,
						Timestamp:  uint64(time.Now().Unix()),
						RelayA:     relayId,
						RelayB:     sampleRelayId,
						RTT:        rtt,
						Jitter:     jitter,
						PacketLoss: pl,
						Routable:   sampleRoutable[i],
					})
				}

				// build relay stats message

				numUnroutable := numSamples - numRoutable
				maxSessions := relayData.RelayArray[relayIndex].MaxSessions
				numSessions := relayUpdateRequest.SessionCount
				full := maxSessions != 0 && numSessions >= uint64(maxSessions)

				relayStatsMessage := messages.RelayStatsMessage{
					Version:       messages.RelayStatsMessageVersion_Write,
					Timestamp:     uint64(time.Now().Unix()),
					NumSessions:   uint32(numSessions),
					MaxSessions:   uint32(maxSessions),
					NumRoutable:   uint32(numRoutable),
					NumUnroutable: uint32(numUnroutable),
					Full:          full,
				}

				// update relay stats

				if service.IsLeader() {

					messageBuffer := make([]byte, maxRelayStatsMessageBytes)

					message := relayStatsMessage.Write(messageBuffer[:])

					relayStatsProducer.MessageChannel <- message
				}

				// update ping stats

				if service.IsLeader() {

					messageBuffer := make([]byte, maxPingStatsMessageBytes)

					for i := 0; i < len(pingStatsMessages); i++ {
						message := pingStatsMessages[i].Write(messageBuffer[:])
						pingStatsProducer.MessageChannel <- message
					}
				}
			}
		}
	}()
}

func UpdateRouteMatrix(service *common.Service, relayManager *common.RelayManager) {

	ticker := time.NewTicker(routeMatrixInterval)

	hasSeenDataStores := false

	go func() {

		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				timeStart := time.Now()

				// build relays csv

				currentTime := time.Now().Unix()

				relayData := service.RelayData()

				relaysCSVDataNew := relayManager.GetRelaysCSV(currentTime, relayData.RelayIds, relayData.RelayNames, relayData.RelayAddresses)

				// build the cost matrix

				costs := relayManager.GetCosts(currentTime, relayData.RelayIds, maxRTT, maxJitter, maxPacketLoss)

				costMatrixNew := &common.CostMatrix{
					Version:            common.CostMatrixVersion_Write,
					RelayIds:           relayData.RelayIds,
					RelayAddresses:     relayData.RelayAddresses,
					RelayNames:         relayData.RelayNames,
					RelayLatitudes:     relayData.RelayLatitudes,
					RelayLongitudes:    relayData.RelayLongitudes,
					RelayDatacenterIds: relayData.RelayDatacenterIds,
					DestRelays:         relayData.DestRelays,
					Costs:              costs,
				}

				// serve up as internal cost matrix

				costMatrixDataNew, err := costMatrixNew.Write(costMatrixBufferSize)
				if err != nil {
					core.Error("could not write cost matrix: %v", err)
					continue
				}

				costMatrixInternalMutex.Lock()
				costMatrixInternalData = costMatrixDataNew
				costMatrixInternalMutex.Unlock()

				// optimize cost matrix -> route matrix

				numCPUs := runtime.NumCPU()

				numSegments := relayData.NumRelays
				if numCPUs < relayData.NumRelays {
					numSegments = relayData.NumRelays / 5
					if numSegments == 0 {
						numSegments = 1
					}
				}

				costThreshold := int32(1)

				routeMatrixNew := &common.RouteMatrix{
					CreatedAt:          uint64(time.Now().Unix()),
					Version:            common.RouteMatrixVersion_Write,
					RelayIds:           costMatrixNew.RelayIds,
					RelayAddresses:     costMatrixNew.RelayAddresses,
					RelayNames:         costMatrixNew.RelayNames,
					RelayLatitudes:     costMatrixNew.RelayLatitudes,
					RelayLongitudes:    costMatrixNew.RelayLongitudes,
					RelayDatacenterIds: costMatrixNew.RelayDatacenterIds,
					DestRelays:         costMatrixNew.DestRelays,
					RouteEntries:       core.Optimize2(relayData.NumRelays, numSegments, costs, costThreshold, relayData.RelayDatacenterIds, relayData.DestRelays),
					BinFileBytes:       int32(len(relayData.DatabaseBinFile)),
					BinFileData:        relayData.DatabaseBinFile,
				}

				// serve up as internal route matrix

				routeMatrixDataNew, err := routeMatrixNew.Write(routeMatrixBufferSize)
				if err != nil {
					core.Error("could not write route matrix: %v", err)
					continue
				}

				routeMatrixInternalMutex.Lock()
				routeMatrixInternalData = routeMatrixDataNew
				routeMatrixInternalMutex.Unlock()

				// store our most recent cost and route matrix in redis

				dataStores := []common.DataStoreConfig{
					{
						Name: "relays",
						Data: relaysCSVDataNew,
					},
					{
						Name: "cost_matrix",
						Data: costMatrixDataNew,
					},
					{
						Name: "route_matrix",
						Data: routeMatrixDataNew,
					},
				}

				service.UpdateLeaderStore(dataStores)

				// load the master cost and route matrix from redis (leader election)

				dataStores = service.LoadLeaderStore()

				if len(dataStores) == 0 {
					// IMPORTANT: don't error unless we've already seen data stores succeed once, and it is failing now
					// otherwise, we get error spam on startup while we are waiting for the first leader to be elected
					if hasSeenDataStores {
						core.Error("could not get data stores from redis selector")
					}
					continue
				}

				hasSeenDataStores = true

				relaysCSVDataNew = dataStores[0].Data
				if relaysCSVDataNew == nil {
					core.Error("failed to get relays from redis selector")
					continue
				}

				costMatrixDataNew = dataStores[1].Data
				if costMatrixDataNew == nil {
					core.Error("failed to get cost matrix from redis selector")
					continue
				}

				routeMatrixDataNew = dataStores[2].Data
				if routeMatrixDataNew == nil {
					core.Error("failed to get route matrix from redis selector")
					continue
				}

				// serve up as official data

				relaysMutex.Lock()
				relaysCSVData = relaysCSVDataNew
				relaysMutex.Unlock()

				costMatrixMutex.Lock()
				costMatrixData = costMatrixDataNew
				costMatrixMutex.Unlock()

				routeMatrixMutex.Lock()
				routeMatrixData = routeMatrixDataNew
				routeMatrixMutex.Unlock()

				// we are done!

				timeFinish := time.Now()

				optimizeDuration := timeFinish.Sub(timeStart)

				core.Debug("route optimization: %d relays in %s", relayData.NumRelays, optimizeDuration)
			}
		}
	}()
}
