package main

import (
	"time"
	"strconv"
	"net/http"
	"encoding/json"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/portal"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)


var pool *redis.Pool

func main() {

	service := common.CreateService("api")

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPoolActive := envvar.GetInt("REDIS_POOL_ACTIVE", 1000)
	redisPoolIdle := envvar.GetInt("REDIS_POOL_IDLE", 10000)

	core.Log("redis hostname: %s", redisHostname)
	core.Log("redis pool active: %s", redisPoolActive)
	core.Log("redis pool idle: %s", redisPoolIdle)

	pool = common.CreateRedisPool(redisHostname, redisPoolActive, redisPoolIdle)

	service.Router.HandleFunc("/portal/session_counts", portalSessionCountsHandler)
	service.Router.HandleFunc("/portal/sessions/{begin}/{end}", portalSessionsHandler)
	service.Router.HandleFunc("/portal/session_data/{session_id}", portalSessionDataHandler)

	/*
	service.Router.HandleFunc("/portal/server_count", portalServerCountHandler)
	service.Router.HandleFunc("/portal/servers/{begin}/{end}", portalServersHandler)
	service.Router.HandleFunc("/portal/server_data/{server_address}", portalServerDataHandler)

	service.Router.HandleFunc("/portal/relay_count", portalRelayCountHandler)
	service.Router.HandleFunc("/portal/relays/{begin}/{end}", portalRelaysHandler)
	service.Router.HandleFunc("/portal/relay_data/{relay_address}", portalRelayDataHandler)
	*/

	service.Router.HandleFunc("/admin/relays", adminRelaysHandler)

	service.StartWebServer()

	service.WaitForShutdown()
}

// ---------------------------------------------------------------------------------------------------------------------

type PortalSessionCountsResponse struct {
	TotalSessionCount int `json:"total_session_count"`
	NextSessionCount  int `json:"next_session_count"`
}

func portalSessionCountsHandler(w http.ResponseWriter, r *http.Request) {
	response := PortalSessionCountsResponse{}
	response.TotalSessionCount, response.NextSessionCount = portal.GetSessionCounts(pool, time.Now().Unix()/60)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalSessionsResponse struct {
	Sessions []portal.SessionEntry `json:"sessions"`
}

func portalSessionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	begin, err := strconv.ParseUint(vars["begin"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	end, err := strconv.ParseUint(vars["end"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response := PortalSessionsResponse{}
	response.Sessions = portal.GetSessions(pool, time.Now().Unix()/60, int(begin), int(end))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalSessionDataResponse struct {
	SessionData *portal.SessionData `json:"session_data"`
	SliceData   []portal.SliceData `json:"slice_data"`
	NearRelayData []portal.NearRelayData `json:"near_relay_data"`
}

func portalSessionDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionId, err := strconv.ParseUint(vars["session_id"], 16, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response := PortalSessionDataResponse{}
	response.SessionData, response.SliceData, response.NearRelayData = portal.GetSessionData(pool, sessionId)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type PortalServerCountResponse struct {
	ServerCount int `json:"server_count"`
}

func portalServerCountHandler(w http.ResponseWriter, r *http.Request) {
	response := PortalServerCountResponse{}
	response.ServerCount = portal.GetServerCount(pool, time.Now().Unix()/60)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalServersResponse struct {
	Servers []portal.ServerEntry `json:"servers"`
}

func portalServersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	begin, err := strconv.ParseUint(vars["begin"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	end, err := strconv.ParseUint(vars["end"], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	response := PortalServersResponse{}
	response.Servers = portal.GetServers(pool, time.Now().Unix()/60, int(begin), int(end))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminRelaysResponse struct {
	// todo: array of *admin* relay data, including ssh address, user, etc.
}

func adminRelaysHandler(w http.ResponseWriter, r *http.Request) {
	response := AdminRelaysResponse{}
	// todo: get admin relays from somewhere (postgres?)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------
