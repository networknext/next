package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

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

var service *common.Service

func main() {

	service = common.CreateService("api")

	pgsqlConfig := envvar.GetString("PGSQL_CONFIG", "host=127.0.0.1 port=5432 user=developer dbname=postgres sslmode=disable")
	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPoolActive := envvar.GetInt("REDIS_POOL_ACTIVE", 1000)
	redisPoolIdle := envvar.GetInt("REDIS_POOL_IDLE", 10000)
	enableAdmin := envvar.GetBool("ENABLE_ADMIN", true)
	enablePortal := envvar.GetBool("ENABLE_PORTAL", true)
	enableDatabase := envvar.GetBool("ENABLE_DATABASE", true)

	core.Log("pgsql config: %s", pgsqlConfig)
	core.Log("redis hostname: %s", redisHostname)
	core.Log("redis pool active: %d", redisPoolActive)
	core.Log("redis pool idle: %d", redisPoolIdle)
	core.Log("enable admin: %v", enableAdmin)
	core.Log("enable portal: %v", enablePortal)
	core.Log("enable database: %v", enableDatabase)

	service.Router.HandleFunc("/ping", pingHandler)

	if enablePortal {

		pool = common.CreateRedisPool(redisHostname, redisPoolActive, redisPoolIdle)

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
	}

	if enableAdmin {

		controller = admin.CreateController(pgsqlConfig)

		service.Router.HandleFunc("/admin/create_customer", adminCreateCustomerHandler).Methods("POST")
		service.Router.HandleFunc("/admin/customers", adminReadCustomersHandler).Methods("GET")
		service.Router.HandleFunc("/admin/update_customer", adminUpdateCustomerHandler).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_customer", adminDeleteCustomerHandler).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_buyer", adminCreateBuyerHandler).Methods("POST")
		service.Router.HandleFunc("/admin/buyers", adminReadBuyersHandler).Methods("GET")
		service.Router.HandleFunc("/admin/update_buyer", adminUpdateBuyerHandler).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_buyer", adminDeleteBuyerHandler).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_seller", adminCreateSellerHandler).Methods("POST")
		service.Router.HandleFunc("/admin/sellers", adminReadSellersHandler).Methods("GET")
		service.Router.HandleFunc("/admin/update_seller", adminUpdateSellerHandler).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_seller", adminDeleteSellerHandler).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_datacenter", adminCreateDatacenterHandler).Methods("POST")
		service.Router.HandleFunc("/admin/datacenters", adminReadDatacentersHandler).Methods("GET")
		service.Router.HandleFunc("/admin/update_datacenter", adminUpdateDatacenterHandler).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_datacenter", adminDeleteDatacenterHandler).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_relay", adminCreateRelayHandler).Methods("POST")
		service.Router.HandleFunc("/admin/relays", adminReadRelaysHandler).Methods("GET")
		service.Router.HandleFunc("/admin/update_relay", adminUpdateRelayHandler).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_relay", adminDeleteRelayHandler).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_route_shader", adminCreateRouteShaderHandler).Methods("POST")
		service.Router.HandleFunc("/admin/route_shaders", adminReadRouteShadersHandler).Methods("GET")
		service.Router.HandleFunc("/admin/update_route_shader", adminUpdateRouteShaderHandler).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_route_shader", adminDeleteRouteShaderHandler).Methods("DELETE")
		service.Router.HandleFunc("/admin/route_shader_defaults", adminRouteShaderDefaultsHandler).Methods("GET")

		service.Router.HandleFunc("/admin/create_buyer_datacenter_settings", adminCreateBuyerDatacenterSettingsHandler).Methods("POST")
		service.Router.HandleFunc("/admin/buyer_datacenter_settings", adminReadBuyerDatacenterSettingsHandler).Methods("GET")
		service.Router.HandleFunc("/admin/update_buyer_datacenter_settings", adminUpdateBuyerDatacenterSettingsHandler).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_buyer_datacenter_settings/{buyerId}/{datacenterId}", adminDeleteBuyerDatacenterSettingsHandler).Methods("DELETE")
	}

	if enableDatabase {

		service.LoadDatabase()

		service.Router.HandleFunc("/database/json", databaseJSONHandler).Methods("GET")
		service.Router.HandleFunc("/database/binary", databaseBinaryHandler).Methods("GET")
		service.Router.HandleFunc("/database/header", databaseHeaderHandler).Methods("GET")
		service.Router.HandleFunc("/database/buyers", databaseBuyersHandler).Methods("GET")
		service.Router.HandleFunc("/database/sellers", databaseSellersHandler).Methods("GET")
		service.Router.HandleFunc("/database/datacenters", databaseDatacentersHandler).Methods("GET")
		service.Router.HandleFunc("/database/relays", databaseRelaysHandler).Methods("GET")
		service.Router.HandleFunc("/database/buyer_datacenter_settings", databaseBuyerDatacenterSettingsHandler).Methods("GET")
	}

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
	Sessions []portal.SessionData `json:"sessions"`
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
	sessionId, err := strconv.ParseUint(vars["session_id"], 10, 64)
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
	Servers []portal.ServerData `json:"servers"`
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
	Relays []portal.RelayData `json:"relays"`
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

func adminCreateBuyerHandler(w http.ResponseWriter, r *http.Request) {
	var buyer admin.BuyerData
	err := json.NewDecoder(r.Body).Decode(&buyer)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	buyerId, err := controller.CreateBuyer(&buyer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprintf(w, "%d", buyerId)
}

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

func adminUpdateBuyerHandler(w http.ResponseWriter, r *http.Request) {
	var buyer admin.BuyerData
	err := json.NewDecoder(r.Body).Decode(&buyer)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.UpdateBuyer(&buyer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func adminDeleteBuyerHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	buyerId, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.DeleteBuyer(buyerId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------------------------------------------------

func adminCreateSellerHandler(w http.ResponseWriter, r *http.Request) {
	var seller admin.SellerData
	err := json.NewDecoder(r.Body).Decode(&seller)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sellerId, err := controller.CreateSeller(&seller)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprintf(w, "%d", sellerId)
}

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

func adminUpdateSellerHandler(w http.ResponseWriter, r *http.Request) {
	var seller admin.SellerData
	err := json.NewDecoder(r.Body).Decode(&seller)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.UpdateSeller(&seller)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func adminDeleteSellerHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	sellerId, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.DeleteSeller(sellerId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------------------------------------------------

func adminCreateDatacenterHandler(w http.ResponseWriter, r *http.Request) {
	var datacenter admin.DatacenterData
	err := json.NewDecoder(r.Body).Decode(&datacenter)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	datacenterId, err := controller.CreateDatacenter(&datacenter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprintf(w, "%d", datacenterId)
}

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

func adminUpdateDatacenterHandler(w http.ResponseWriter, r *http.Request) {
	var datacenter admin.DatacenterData
	err := json.NewDecoder(r.Body).Decode(&datacenter)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.UpdateDatacenter(&datacenter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func adminDeleteDatacenterHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	datacenterId, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.DeleteDatacenter(datacenterId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------------------------------------------------

func adminCreateRelayHandler(w http.ResponseWriter, r *http.Request) {
	var relay admin.RelayData
	err := json.NewDecoder(r.Body).Decode(&relay)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	relayId, err := controller.CreateRelay(&relay)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprintf(w, "%d", relayId)
}

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

func adminUpdateRelayHandler(w http.ResponseWriter, r *http.Request) {
	var relay admin.RelayData
	err := json.NewDecoder(r.Body).Decode(&relay)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.UpdateRelay(&relay)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func adminDeleteRelayHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	relayId, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.DeleteRelay(relayId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------------------------------------------------

func adminRouteShaderDefaultsHandler(w http.ResponseWriter, r *http.Request) {
	routeShader := controller.RouteShaderDefaults()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routeShader)
}

func adminCreateRouteShaderHandler(w http.ResponseWriter, r *http.Request) {
	var routeShader admin.RouteShaderData
	err := json.NewDecoder(r.Body).Decode(&routeShader)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	routeShaderId, err := controller.CreateRouteShader(&routeShader)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprintf(w, "%d", routeShaderId)
}

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

func adminUpdateRouteShaderHandler(w http.ResponseWriter, r *http.Request) {
	var routeShader admin.RouteShaderData
	err := json.NewDecoder(r.Body).Decode(&routeShader)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.UpdateRouteShader(&routeShader)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func adminDeleteRouteShaderHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	routeShaderId, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.DeleteRouteShader(routeShaderId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------------------------------------------------

func adminCreateBuyerDatacenterSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var settings admin.BuyerDatacenterSettings
	err := json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.CreateBuyerDatacenterSettings(&settings)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprintf(w, "1")
}

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

func adminUpdateBuyerDatacenterSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var settings admin.BuyerDatacenterSettings
	err := json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.UpdateBuyerDatacenterSettings(&settings)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func adminDeleteBuyerDatacenterSettingsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buyerId, err := strconv.ParseUint(vars["buyerId"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	datacenterId, err := strconv.ParseUint(vars["datacenterId"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	_ = body
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.DeleteBuyerDatacenterSettings(buyerId, datacenterId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------------------------------------------------

func databaseJSONHandler(w http.ResponseWriter, r *http.Request) {
	database := service.Database()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(database)
}

func databaseBinaryHandler(w http.ResponseWriter, r *http.Request) {
	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := database.GetBinary()
	w.Header().Set("Content-Type", "application/octet-stream")
	json.NewEncoder(w).Encode(response)
}

func databaseHeaderHandler(w http.ResponseWriter, r *http.Request) {
	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := database.GetHeader()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func databaseBuyersHandler(w http.ResponseWriter, r *http.Request) {
	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := database.GetBuyers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func databaseSellersHandler(w http.ResponseWriter, r *http.Request) {
	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := database.GetSellers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func databaseDatacentersHandler(w http.ResponseWriter, r *http.Request) {
	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := database.GetDatacenters()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func databaseRelaysHandler(w http.ResponseWriter, r *http.Request) {
	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := database.GetRelays()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func databaseBuyerDatacenterSettingsHandler(w http.ResponseWriter, r *http.Request) {
	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := database.GetBuyerDatacenterSettings()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------
