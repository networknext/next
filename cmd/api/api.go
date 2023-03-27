package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
	"strings"

	"github.com/networknext/backend/modules/admin"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/portal"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
)

var pool *redis.Pool

var controller *admin.Controller

var service *common.Service

var privateKey string

func main() {

	service = common.CreateService("api")

	privateKey = envvar.GetString("API_PRIVATE_KEY", "")
	pgsqlConfig := envvar.GetString("PGSQL_CONFIG", "host=127.0.0.1 port=5432 user=developer password=developer dbname=postgres sslmode=disable")
	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPoolActive := envvar.GetInt("REDIS_POOL_ACTIVE", 1000)
	redisPoolIdle := envvar.GetInt("REDIS_POOL_IDLE", 10000)
	enableAdmin := envvar.GetBool("ENABLE_ADMIN", true)
	enablePortal := envvar.GetBool("ENABLE_PORTAL", true)
	enableDatabase := envvar.GetBool("ENABLE_DATABASE", true)

	if privateKey == "" {
		core.Error("You must specify API_PRIVATE_KEY!")
		os.Exit(1)
	}

	core.Debug("pgsql config: %s", pgsqlConfig)
	core.Debug("redis hostname: %s", redisHostname)
	core.Debug("redis pool active: %d", redisPoolActive)
	core.Debug("redis pool idle: %d", redisPoolIdle)
	core.Debug("enable admin: %v", enableAdmin)
	core.Debug("enable portal: %v", enablePortal)
	core.Debug("enable database: %v", enableDatabase)

	service.Router.HandleFunc("/ping", isAuthorized(pingHandler))

	if enablePortal {

		pool = common.CreateRedisPool(redisHostname, redisPoolActive, redisPoolIdle)

		service.Router.HandleFunc("/portal/session_counts", isAuthorized(portalSessionCountsHandler))
		service.Router.HandleFunc("/portal/sessions/{begin}/{end}", isAuthorized(portalSessionsHandler))
		service.Router.HandleFunc("/portal/session/{session_id}", isAuthorized(portalSessionDataHandler))

		service.Router.HandleFunc("/portal/server_count", isAuthorized(portalServerCountHandler))
		service.Router.HandleFunc("/portal/servers/{begin}/{end}", isAuthorized(portalServersHandler))
		service.Router.HandleFunc("/portal/server/{server_address}", isAuthorized(portalServerDataHandler))

		service.Router.HandleFunc("/portal/relay_count", isAuthorized(portalRelayCountHandler))
		service.Router.HandleFunc("/portal/relays/{begin}/{end}", isAuthorized(portalRelaysHandler))
		service.Router.HandleFunc("/portal/relay/{relay_address}", isAuthorized(portalRelayDataHandler))

		service.Router.HandleFunc("/portal/map_data", isAuthorized(portalMapDataHandler))

		service.Router.HandleFunc("/portal/cost_matrix", isAuthorized(portalCostMatrixHandler))
	}

	if enableAdmin {

		controller = admin.CreateController(pgsqlConfig)

		service.Router.HandleFunc("/admin/create_customer", isAuthorized(adminCreateCustomerHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/customers", isAuthorized(adminReadCustomersHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/customer/{customerId}", isAuthorized(adminReadCustomerHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_customer", isAuthorized(adminUpdateCustomerHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_customer", isAuthorized(adminDeleteCustomerHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_buyer", isAuthorized(adminCreateBuyerHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/buyers", isAuthorized(adminReadBuyersHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/buyer/{buyerId}", isAuthorized(adminReadBuyerHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_buyer", isAuthorized(adminUpdateBuyerHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_buyer", isAuthorized(adminDeleteBuyerHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_seller", isAuthorized(adminCreateSellerHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/sellers", isAuthorized(adminReadSellersHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/seller/{sellerId}", isAuthorized(adminReadSellerHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_seller", isAuthorized(adminUpdateSellerHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_seller", isAuthorized(adminDeleteSellerHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_datacenter", isAuthorized(adminCreateDatacenterHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/datacenters", isAuthorized(adminReadDatacentersHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/datacenter/{datacenterId}", isAuthorized(adminReadDatacenterHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_datacenter", isAuthorized(adminUpdateDatacenterHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_datacenter", isAuthorized(adminDeleteDatacenterHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_relay", isAuthorized(adminCreateRelayHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/relays", isAuthorized(adminReadRelaysHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_relay", isAuthorized(adminUpdateRelayHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_relay", isAuthorized(adminDeleteRelayHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_route_shader", isAuthorized(adminCreateRouteShaderHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/route_shaders", isAuthorized(adminReadRouteShadersHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_route_shader", isAuthorized(adminUpdateRouteShaderHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_route_shader", isAuthorized(adminDeleteRouteShaderHandler)).Methods("DELETE")
		service.Router.HandleFunc("/admin/route_shader_defaults", isAuthorized(adminRouteShaderDefaultsHandler)).Methods("GET")

		service.Router.HandleFunc("/admin/create_buyer_datacenter_settings", isAuthorized(adminCreateBuyerDatacenterSettingsHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/buyer_datacenter_settings", isAuthorized(adminReadBuyerDatacenterSettingsHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_buyer_datacenter_settings", isAuthorized(adminUpdateBuyerDatacenterSettingsHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_buyer_datacenter_settings/{buyerId}/{datacenterId}", isAuthorized(adminDeleteBuyerDatacenterSettingsHandler)).Methods("DELETE")
	}

	if enableDatabase {

		service.LoadDatabase()

		service.Router.HandleFunc("/database/json", isAuthorized(databaseJSONHandler)).Methods("GET")
		service.Router.HandleFunc("/database/binary", isAuthorized(databaseBinaryHandler)).Methods("GET")
		service.Router.HandleFunc("/database/header", isAuthorized(databaseHeaderHandler)).Methods("GET")
		service.Router.HandleFunc("/database/buyers", isAuthorized(databaseBuyersHandler)).Methods("GET")
		service.Router.HandleFunc("/database/sellers", isAuthorized(databaseSellersHandler)).Methods("GET")
		service.Router.HandleFunc("/database/datacenters", isAuthorized(databaseDatacentersHandler)).Methods("GET")
		service.Router.HandleFunc("/database/relays", isAuthorized(databaseRelaysHandler)).Methods("GET")
		service.Router.HandleFunc("/database/buyer_datacenter_settings", isAuthorized(databaseBuyerDatacenterSettingsHandler)).Methods("GET")
	}

	service.StartWebServer()

	service.WaitForShutdown()
}

// ---------------------------------------------------------------------------------------------------------------------

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")

		split := strings.Split(auth, "Bearer ")

		if len(split) == 2 {

			apiKey := split[1]

			token, err := jwt.Parse(apiKey, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return []byte(privateKey), nil
			})

			if token == nil ||err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, err.Error())
			}

			endpoint(w, r)

		} else {

			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Not Authorized")
			
		}
	}
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

func portalCostMatrixHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	data := common.LoadMasterServiceData(pool, "relay_backend", "cost_matrix")
	w.Write(data)
}

// ---------------------------------------------------------------------------------------------------------------------

func adminCreateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	var customerData admin.CustomerData
	err := json.NewDecoder(r.Body).Decode(&customerData)
	if err != nil {
		core.Error("failed to read customer data in request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	customerId, err := controller.CreateCustomer(&customerData)
	if err != nil {
		core.Error("failed to create customer: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	core.Log("create customer: %x = %+v", customerId, customerData)
	fmt.Fprintf(w, "%d", customerId)
}

type AdminReadCustomersResponse struct {
	Customers []admin.CustomerData `json:"customers"`
	Error     string               `json:"error"`
}

func adminReadCustomersHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("read customers")
	customers, err := controller.ReadCustomers()
	response := AdminReadCustomersResponse{Customers: customers}
	if err != nil {
		core.Error("failed to read customers: %v", err)
		response.Error = err.Error()
	}
	core.Debug("customers = %+v", customers)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadCustomerResponse struct {
	Customer admin.CustomerData `json:"customer"`
	Error    string             `json:"error"`
}

func adminReadCustomerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerId, err := strconv.ParseUint(vars["customerId"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read customer %x", customerId)
	customer, err := controller.ReadCustomer(customerId)
	response := AdminReadCustomerResponse{Customer: customer}
	if err != nil {
		core.Error("failed to read customer: %v", err)
		response.Error = err.Error()
	}
	core.Debug("customer %x = %+v", customerId, customer)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func adminUpdateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	var customer admin.CustomerData
	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		core.Error("failed to decode update customer request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("update customer: %+v", customer)
	err = controller.UpdateCustomer(&customer)
	if err != nil {
		core.Error("failed to update customer: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	core.Debug("update customer %x = %+v", customer.CustomerId, customer)
	w.WriteHeader(http.StatusOK)
}

func adminDeleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		core.Error("failed to read delete customer request body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	customerId, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		core.Error("failed to parse customer id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("delete customer: %x", customerId)
	err = controller.DeleteCustomer(customerId)
	if err != nil {
		core.Error("failed to delete customer: %v", err)
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
		// todo
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	buyerId, err := controller.CreateBuyer(&buyer)
	if err != nil {
		// todo
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// todo
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
		// todo
		response.Error = err.Error()
	}
	// todo
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadBuyerResponse struct {
	Buyer    admin.BuyerData 	`json:"buyer"`
	Error    string             `json:"error"`
}

func adminReadBuyerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buyerId, err := strconv.ParseUint(vars["buyerId"], 10, 64)
	if err != nil {
		// todo
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read buyer %x", buyerId)
	buyer, err := controller.ReadBuyer(buyerId)
	response := AdminReadBuyerResponse{Buyer: buyer}
	if err != nil {
		core.Error("failed to read buyer: %v", err)
		response.Error = err.Error()
	}
	core.Debug("buyer %x = %+v", buyerId, buyer)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func adminUpdateBuyerHandler(w http.ResponseWriter, r *http.Request) {
	var buyer admin.BuyerData
	err := json.NewDecoder(r.Body).Decode(&buyer)
	if err != nil {
		// todo
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.UpdateBuyer(&buyer)
	if err != nil {
		// todo
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// todo
	w.WriteHeader(http.StatusOK)
}

func adminDeleteBuyerHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// todo
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	buyerId, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		// todo
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.DeleteBuyer(buyerId)
	if err != nil {
		// todo
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------------------------------------------------

func adminCreateSellerHandler(w http.ResponseWriter, r *http.Request) {
	var sellerData admin.SellerData
	err := json.NewDecoder(r.Body).Decode(&sellerData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		core.Error("failed to read seller data in request: %v", err)
		return
	}
	sellerId, err := controller.CreateSeller(&sellerData)
	if err != nil {
		core.Error("failed to create seller: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	core.Log("create seller: %x = %+v", sellerId, sellerData)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprintf(w, "%d", sellerId)
}

type AdminReadSellersResponse struct {
	Sellers []admin.SellerData `json:"sellers"`
	Error   string             `json:"error"`
}

func adminReadSellersHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("read sellers")
	sellers, err := controller.ReadSellers()
	response := AdminReadSellersResponse{Sellers: sellers}
	if err != nil {
		core.Error("failed to read sellers: %v", err)
		response.Error = err.Error()
	}
	core.Debug("sellers = %+v", sellers)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadSellerResponse struct {
	Seller admin.SellerData `json:"seller"`
	Error  string           `json:"error"`
}

func adminReadSellerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sellerId, err := strconv.ParseUint(vars["sellerId"], 10, 64)
	if err != nil {
		// todo
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read seller %x", sellerId)
	seller, err := controller.ReadSeller(sellerId)
	response := AdminReadSellerResponse{Seller: seller}
	if err != nil {
		response.Error = err.Error()
	}
	core.Debug("seller %x = %+v", sellerId, seller)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func adminUpdateSellerHandler(w http.ResponseWriter, r *http.Request) {
	var seller admin.SellerData
	err := json.NewDecoder(r.Body).Decode(&seller)
	if err != nil {
		core.Error("failed to decode update seller request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("update seller: %+v", seller)
	err = controller.UpdateSeller(&seller)
	if err != nil {
		core.Error("failed to update seller: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	core.Debug("update seller %x = %+v", seller.SellerId, seller)
	w.WriteHeader(http.StatusOK)
}

func adminDeleteSellerHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		core.Error("failed to read delete seller request body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	sellerId, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		core.Error("failed to parse seller id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("delete seller: %x", sellerId)
	err = controller.DeleteSeller(sellerId)
	if err != nil {
		core.Error("failed to delete seller: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------------------------------------------------

func adminCreateDatacenterHandler(w http.ResponseWriter, r *http.Request) {
	var datacenterData admin.DatacenterData
	err := json.NewDecoder(r.Body).Decode(&datacenterData)
	if err != nil {
		core.Error("failed to read datacenter data in request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	datacenterId, err := controller.CreateDatacenter(&datacenterData)
	if err != nil {
		core.Error("failed to create datacenter: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	core.Log("create datacenter: %x = %+v", datacenterId, datacenterData)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	fmt.Fprintf(w, "%d", datacenterId)
}

type AdminReadDatacentersResponse struct {
	Datacenters []admin.DatacenterData `json:"datacenters"`
	Error       string                 `json:"error"`
}

func adminReadDatacentersHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("read datacenters")
	datacenters, err := controller.ReadDatacenters()
	response := AdminReadDatacentersResponse{Datacenters: datacenters}
	if err != nil {
		core.Error("failed to read datacenters: %v", err)
		response.Error = err.Error()
	}
	core.Debug("datacenters = %+v", datacenters)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadDatacenterResponse struct {
	Datacenter admin.DatacenterData `json:"datacenter"`
	Error      string               `json:"error"`
}

func adminReadDatacenterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	datacenterId, err := strconv.ParseUint(vars["datacenterId"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read datacenter %x", datacenterId)
	datacenter, err := controller.ReadDatacenter(datacenterId)
	response := AdminReadDatacenterResponse{Datacenter: datacenter}
	if err != nil {
		response.Error = err.Error()
	}
	core.Debug("datacenter %x = %+v", datacenterId, datacenter)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func adminUpdateDatacenterHandler(w http.ResponseWriter, r *http.Request) {
	var datacenter admin.DatacenterData
	err := json.NewDecoder(r.Body).Decode(&datacenter)
	if err != nil {
		core.Error("failed to decode update datacenter request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.UpdateDatacenter(&datacenter)
	if err != nil {
		core.Error("failed to update datacenter: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	core.Debug("update datacenter %x = %+v", datacenter.DatacenterId, datacenter)
	w.WriteHeader(http.StatusOK)
}

func adminDeleteDatacenterHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		core.Error("failed to read delete datacenter request body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	datacenterId, err := strconv.ParseUint(string(body), 10, 64)
	if err != nil {
		core.Error("failed to parse datacenter id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("delete datacenter: %x", datacenterId)
	err = controller.DeleteDatacenter(datacenterId)
	if err != nil {
		core.Error("failed to delete datacenter: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------------------------------------------------

func adminCreateRelayHandler(w http.ResponseWriter, r *http.Request) {
	var relayData admin.RelayData
	err := json.NewDecoder(r.Body).Decode(&relayData)
	if err != nil {
		core.Error("create relay failed to read relay data in request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	relayId, err := controller.CreateRelay(&relayData)
	if err != nil {
		core.Error("failed to create relay: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		// todo: we definitely need a way to return this error to the caller here
		return
	}
	core.Log("create relay: %x = %+v", relayId, relayData)
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
	data := database.GetBinary()
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
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
