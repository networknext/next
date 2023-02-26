package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"fmt"
	"io/ioutil"

	"github.com/networknext/backend/modules/admin"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/portal"

	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

var pool *redis.Pool

var controller *admin.Controller

func main() {

	service := common.CreateService("api")

	pgsqlConfig := envvar.GetString("PGSQL_CONFIG", "host=127.0.0.1 port=5432 dbname=postgres sslmode=disable")
	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPoolActive := envvar.GetInt("REDIS_POOL_ACTIVE", 1000)
	redisPoolIdle := envvar.GetInt("REDIS_POOL_IDLE", 10000)

	core.Log("pgsql config: %s", pgsqlConfig)
	core.Log("redis hostname: %s", redisHostname)
	core.Log("redis pool active: %d", redisPoolActive)
	core.Log("redis pool idle: %d", redisPoolIdle)

	pool = common.CreateRedisPool(redisHostname, redisPoolActive, redisPoolIdle)

	controller = admin.CreateController(pgsqlConfig)

	service.Router.HandleFunc("/ping", pingHandler)

	service.Router.HandleFunc("/portal/session_counts", portalSessionCountsHandler)
	service.Router.HandleFunc("/portal/sessions/{begin}/{end}", portalSessionsHandler)
	service.Router.HandleFunc("/portal/session/{session_id}", portalSessionDataHandler)

	service.Router.HandleFunc("/portal/server_count", portalServerCountHandler)
	service.Router.HandleFunc("/portal/servers/{begin}/{end}", portalServersHandler)
	service.Router.HandleFunc("/portal/server/{server_address}", portalServerDataHandler)

	service.Router.HandleFunc("/portal/relay_count", portalRelayCountHandler)
	service.Router.HandleFunc("/portal/relays/{begin}/{end}", portalRelaysHandler)
	service.Router.HandleFunc("/portal/relay/{relay_address}", portalRelayDataHandler)

	service.Router.HandleFunc("/portal/map_data", portalMapDataHandler)

	service.Router.HandleFunc("/admin/create_customer", adminCreateCustomerHandler).Methods("POST")
	service.Router.HandleFunc("/admin/customers", adminReadCustomersHandler).Methods("GET")
	service.Router.HandleFunc("/admin/update_customer", adminUpdateCustomerHandler).Methods("PUT")
	service.Router.HandleFunc("/admin/delete_customer", adminDeleteCustomerHandler).Methods("DELETE")

	service.Router.HandleFunc("/admin/buyers", adminReadBuyersHandler)

	service.Router.HandleFunc("/admin/sellers", adminReadSellersHandler)

	service.Router.HandleFunc("/admin/datacenters", adminReadDatacentersHandler)

	service.Router.HandleFunc("/admin/relays", adminReadRelaysHandler)

	service.Router.HandleFunc("/admin/route_shaders", adminReadRouteShadersHandler)

	service.Router.HandleFunc("/admin/buyer_datacenter_settings", adminReadBuyerDatacenterSettingsHandler)

	service.StartWebServer()

	service.WaitForShutdown()
}

// ---------------------------------------------------------------------------------------------------------------------

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
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
	SessionData   *portal.SessionData    `json:"session_data"`
	SliceData     []portal.SliceData     `json:"slice_data"`
	NearRelayData []portal.NearRelayData `json:"near_relay_data"`
}

func portalSessionDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionId, err := strconv.ParseUint(vars["session_id"], 16, 64)
	if err != nil {
		sessionId, err = strconv.ParseUint(vars["session_id"], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
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

type PortalServerDataResponse struct {
	ServerData       *portal.ServerData `json:"server_data"`
	ServerSessionIds []uint64           `json:"server_session_ids"`
}

func portalServerDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serverAddress := vars["server_address"]
	response := PortalServerDataResponse{}
	response.ServerData, response.ServerSessionIds = portal.GetServerData(pool, serverAddress, time.Now().Unix()/60)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// --------------------------------------------------------------------------------------------------------------------

type PortalRelayCountResponse struct {
	RelayCount int `json:"relay_count"`
}

func portalRelayCountHandler(w http.ResponseWriter, r *http.Request) {
	response := PortalRelayCountResponse{}
	response.RelayCount = portal.GetRelayCount(pool, time.Now().Unix()/60)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalRelaysResponse struct {
	Relays []portal.RelayEntry `json:"relays"`
}

func portalRelaysHandler(w http.ResponseWriter, r *http.Request) {
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
	response := PortalRelaysResponse{}
	response.Relays = portal.GetRelays(pool, time.Now().Unix()/60, int(begin), int(end))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalRelayDataResponse struct {
	RelayData    *portal.RelayData    `json:"relay_data"`
	RelaySamples []portal.RelaySample `json:"relay_samples"`
}

func portalRelayDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	relayAddress := vars["relay_address"]
	response := PortalRelayDataResponse{}
	response.RelayData, response.RelaySamples = portal.GetRelayData(pool, relayAddress)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

func portalMapDataHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	data := common.LoadMasterServiceData(pool, "map_cruncher", "map_data")
	w.Write(data)
}

// ---------------------------------------------------------------------------------------------------------------------

func adminCreateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	var customer admin.CustomerData
    err := json.NewDecoder(r.Body).Decode(&customer)
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
    }
    customerId, err := controller.CreateCustomer(&customer)
    if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
    }
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprintf(w, "%d", customerId)
}

type AdminReadCustomersResponse struct {
	Customers []admin.CustomerData `json:"customers"`
	Error     string               `json:"error"`
}

func adminReadCustomersHandler(w http.ResponseWriter, r *http.Request) {
	customers, err := controller.ReadCustomers()
	response := AdminReadCustomersResponse{Customers: customers}
	if err != nil {
		response.Error = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func adminUpdateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	var customer admin.CustomerData
    err := json.NewDecoder(r.Body).Decode(&customer)
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
    }
    err = controller.UpdateCustomer(&customer)
    if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
    }
	w.WriteHeader(http.StatusOK)
}

func adminDeleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
    }
    r.Body.Close()
    customerId, err := strconv.ParseUint(string(body), 10, 64)
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
    }
    err = controller.DeleteCustomer(customerId)
    if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
    }
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminReadBuyersResponse struct {
	Buyers []admin.BuyerData `json:"buyers"`
	Error  string            `json:"error"`
}

func adminReadBuyersHandler(w http.ResponseWriter, r *http.Request) {
	buyers, err := controller.ReadBuyers()
	response := AdminReadBuyersResponse{Buyers: buyers}
	if err != nil {
		response.Error = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminReadSellersResponse struct {
	Sellers []admin.SellerData `json:"sellers"`
	Error   string             `json:"error"`
}

func adminReadSellersHandler(w http.ResponseWriter, r *http.Request) {
	sellers, err := controller.ReadSellers()
	response := AdminReadSellersResponse{Sellers: sellers}
	if err != nil {
		response.Error = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminReadDatacentersResponse struct {
	Datacenters []admin.DatacenterData `json:"datacenters"`
	Error       string                 `json:"error"`
}

func adminReadDatacentersHandler(w http.ResponseWriter, r *http.Request) {
	datacenters, err := controller.ReadDatacenters()
	response := AdminReadDatacentersResponse{Datacenters: datacenters}
	if err != nil {
		response.Error = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminReadRelaysResponse struct {
	Relays []admin.RelayData `json:"relays"`
	Error  string            `json:"error"`
}

func adminReadRelaysHandler(w http.ResponseWriter, r *http.Request) {
	relays, err := controller.ReadRelays()
	response := AdminReadRelaysResponse{Relays: relays}
	if err != nil {
		response.Error = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminReadRouteShadersResponse struct {
	RouteShaders []admin.RouteShaderData `json:"route_shaders"`
	Error        string                  `json:"error"`
}

func adminReadRouteShadersHandler(w http.ResponseWriter, r *http.Request) {
	routeShaders, err := controller.ReadRouteShaders()
	response := AdminReadRouteShadersResponse{RouteShaders: routeShaders}
	if err != nil {
		response.Error = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminReadBuyerDatacenterSettingsResponse struct {
	BuyerDatacenterSettings []admin.BuyerDatacenterSettings `json:"buyer_datacenter_settings"`
	Error                   string                          `json:"error"`
}

func adminReadBuyerDatacenterSettingsHandler(w http.ResponseWriter, r *http.Request) {
	buyerDatacenterSettings, err := controller.ReadBuyerDatacenterSettings()
	response := AdminReadBuyerDatacenterSettingsResponse{BuyerDatacenterSettings: buyerDatacenterSettings}
	if err != nil {
		response.Error = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------
