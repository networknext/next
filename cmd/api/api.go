package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/networknext/accelerate/modules/admin"
	"github.com/networknext/accelerate/modules/common"
	"github.com/networknext/accelerate/modules/core"
	"github.com/networknext/accelerate/modules/envvar"
	"github.com/networknext/accelerate/modules/portal"

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
		service.Router.HandleFunc("/admin/delete_customer/{customerId}", isAuthorized(adminDeleteCustomerHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_seller", isAuthorized(adminCreateSellerHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/sellers", isAuthorized(adminReadSellersHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/seller/{sellerId}", isAuthorized(adminReadSellerHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_seller", isAuthorized(adminUpdateSellerHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_seller/{sellerId}", isAuthorized(adminDeleteSellerHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_buyer", isAuthorized(adminCreateBuyerHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/buyers", isAuthorized(adminReadBuyersHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/buyer/{buyerId}", isAuthorized(adminReadBuyerHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_buyer", isAuthorized(adminUpdateBuyerHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_buyer/{buyerId}", isAuthorized(adminDeleteBuyerHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_datacenter", isAuthorized(adminCreateDatacenterHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/datacenters", isAuthorized(adminReadDatacentersHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/datacenter/{datacenterId}", isAuthorized(adminReadDatacenterHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_datacenter", isAuthorized(adminUpdateDatacenterHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_datacenter/{datacenterId}", isAuthorized(adminDeleteDatacenterHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_relay", isAuthorized(adminCreateRelayHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/relays", isAuthorized(adminReadRelaysHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/relay/{relayId}", isAuthorized(adminReadRelayHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_relay", isAuthorized(adminUpdateRelayHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_relay/{relayId}", isAuthorized(adminDeleteRelayHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_route_shader", isAuthorized(adminCreateRouteShaderHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/route_shaders", isAuthorized(adminReadRouteShadersHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/route_shader/{routeShaderId}", isAuthorized(adminReadRouteShaderHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_route_shader", isAuthorized(adminUpdateRouteShaderHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_route_shader/{routeShaderId}", isAuthorized(adminDeleteRouteShaderHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_buyer_datacenter_settings", isAuthorized(adminCreateBuyerDatacenterSettingsHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/buyer_datacenter_settings", isAuthorized(adminReadBuyerDatacenterSettingsListHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/buyer_datacenter_settings/{buyerId}/{datacenterId}", isAuthorized(adminReadBuyerDatacenterSettingsHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_buyer_datacenter_settings", isAuthorized(adminUpdateBuyerDatacenterSettingsHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_buyer_datacenter_settings/{buyerId}/{datacenterId}", isAuthorized(adminDeleteBuyerDatacenterSettingsHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_buyer_keypair", isAuthorized(adminCreateBuyerKeypairHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/buyer_keypairs", isAuthorized(adminReadBuyerKeypairHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/buyer_keypair/{buyerKeypairId}", isAuthorized(adminReadBuyerKeypairHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_buyer_keypair", isAuthorized(adminUpdateBuyerKeypairHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_buyer_keypair/{buyerKeypairId}", isAuthorized(adminDeleteBuyerKeypairHandler)).Methods("DELETE")

		service.Router.HandleFunc("/admin/create_relay_keypair", isAuthorized(adminCreateRelayKeypairHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/relay_keypairs", isAuthorized(adminReadRelayKeypairHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/relay_keypair/{relayKeypairId}", isAuthorized(adminReadRelayKeypairHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_relay_keypair", isAuthorized(adminUpdateRelayKeypairHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_relay_keypair/{relayKeypairId}", isAuthorized(adminDeleteRelayKeypairHandler)).Methods("DELETE")
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

			if token == nil || err != nil {
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

type AdminCreateCustomerResponse struct {
	Customer admin.CustomerData `json:"customer"`
	Error    string             `json:"error"`
}

func adminCreateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	var response AdminCreateCustomerResponse
	var customerData admin.CustomerData
	err := json.NewDecoder(r.Body).Decode(&customerData)
	if err != nil {
		core.Error("failed to read customer data in create customer request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	customerId, err := controller.CreateCustomer(&customerData)
	if err != nil {
		core.Error("failed to create customer: %v", err)
		response.Error = err.Error()
	} else {
		customerData.CustomerId = customerId
		core.Log("create customer %x", customerId)
		core.Debug("%+v", customerData)
		response.Customer = customerData
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	} else {
		core.Debug("customers = %+v", customers)
	}
	w.WriteHeader(http.StatusOK)
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
		core.Error("read customer could not parse customer id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read customer %x", customerId)
	customer, err := controller.ReadCustomer(customerId)
	response := AdminReadCustomerResponse{Customer: customer}
	if err != nil {
		core.Error("failed to read customer: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", customer)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminUpdateCustomerResponse struct {
	Customer admin.CustomerData `json:"customer"`
	Error    string             `json:"error"`
}

func adminUpdateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	var customer admin.CustomerData
	err := json.NewDecoder(r.Body).Decode(&customer)
	if err != nil {
		core.Error("failed to decode update customer request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("update customer %x", customer.CustomerId)
	response := AdminUpdateCustomerResponse{Customer: customer}
	err = controller.UpdateCustomer(&customer)
	if err != nil {
		core.Error("failed to update customer: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", customer)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminDeleteCustomerResponse struct {
	Error string `json:"error"`
}

func adminDeleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerId, err := strconv.ParseUint(vars["customerId"], 10, 64)
	if err != nil {
		core.Error("delete customer could not parse customer id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("delete customer %x", customerId)
	response := AdminDeleteCustomerResponse{}
	err = controller.DeleteCustomer(customerId)
	if err != nil {
		core.Error("failed to delete customer: %v", err)
		response.Error = err.Error()
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminCreateSellerResponse struct {
	Seller admin.SellerData `json:"seller"`
	Error  string           `json:"error"`
}

func adminCreateSellerHandler(w http.ResponseWriter, r *http.Request) {
	var response AdminCreateSellerResponse
	var sellerData admin.SellerData
	err := json.NewDecoder(r.Body).Decode(&sellerData)
	if err != nil {
		core.Error("failed to read seller data in create seller request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sellerId, err := controller.CreateSeller(&sellerData)
	if err != nil {
		core.Error("failed to create seller: %v", err)
		response.Error = err.Error()
	} else {
		sellerData.SellerId = sellerId
		core.Log("create seller %x", sellerId)
		core.Debug("%+v", sellerData)
		response.Seller = sellerData
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
		core.Error("read seller could not parse seller id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read seller %x", sellerId)
	seller, err := controller.ReadSeller(sellerId)
	response := AdminReadSellerResponse{Seller: seller}
	if err != nil {
		response.Error = err.Error()
	}
	core.Debug("%+v", seller)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminUpdateSellerResponse struct {
	Seller admin.SellerData `json:"seller"`
	Error  string           `json:"error"`
}

func adminUpdateSellerHandler(w http.ResponseWriter, r *http.Request) {
	var seller admin.SellerData
	err := json.NewDecoder(r.Body).Decode(&seller)
	if err != nil {
		core.Error("failed to decode update seller request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("update seller %x", seller.SellerId)
	response := AdminUpdateSellerResponse{Seller: seller}
	err = controller.UpdateSeller(&seller)
	if err != nil {
		core.Error("failed to update seller: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", seller)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminDeleteSellerResponse struct {
	Error string `json:"error"`
}

func adminDeleteSellerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sellerId, err := strconv.ParseUint(vars["sellerId"], 10, 64)
	if err != nil {
		core.Error("delete seller could not parse seller id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("delete seller %x", sellerId)
	response := AdminDeleteSellerResponse{}
	err = controller.DeleteSeller(sellerId)
	if err != nil {
		core.Error("failed to delete seller: %v", err)
		response.Error = err.Error()
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminCreateBuyerResponse struct {
	Buyer admin.BuyerData `json:"buyer"`
	Error string          `json:"error"`
}

func adminCreateBuyerHandler(w http.ResponseWriter, r *http.Request) {
	var response AdminCreateBuyerResponse
	var buyerData admin.BuyerData
	err := json.NewDecoder(r.Body).Decode(&buyerData)
	if err != nil {
		core.Error("failed to read buyer data in create buyer request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	buyerId, err := controller.CreateBuyer(&buyerData)
	if err != nil {
		core.Error("failed to create buyer: %v", err)
		response.Error = err.Error()
	} else {
		buyerData.BuyerId = buyerId
		core.Log("create buyer %x", buyerId)
		core.Debug("%+v", buyerData)
		response.Buyer = buyerData
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	Buyer admin.BuyerData `json:"buyer"`
	Error string          `json:"error"`
}

func adminReadBuyerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buyerId, err := strconv.ParseUint(vars["buyerId"], 10, 64)
	if err != nil {
		core.Error("read buyer could not parse buyer id")
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
	core.Debug("%+v", buyer)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminUpdateBuyerResponse struct {
	Buyer admin.BuyerData `json:"buyer"`
	Error string          `json:"error"`
}

func adminUpdateBuyerHandler(w http.ResponseWriter, r *http.Request) {
	var buyer admin.BuyerData
	err := json.NewDecoder(r.Body).Decode(&buyer)
	if err != nil {
		core.Error("failed to decode update buyer request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("update buyer %x", buyer.BuyerId)
	response := AdminUpdateBuyerResponse{Buyer: buyer}
	err = controller.UpdateBuyer(&buyer)
	if err != nil {
		core.Error("failed to update buyer: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", buyer)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminDeleteBuyerResponse struct {
	Error string `json:"error"`
}

func adminDeleteBuyerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buyerId, err := strconv.ParseUint(vars["buyerId"], 10, 64)
	if err != nil {
		core.Error("delete buyer could not parse buyer id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("delete buyer %x", buyerId)
	response := AdminDeleteBuyerResponse{}
	err = controller.DeleteBuyer(buyerId)
	if err != nil {
		core.Error("failed to delete buyer: %v", err)
		response.Error = err.Error()
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminCreateDatacenterResponse struct {
	Datacenter admin.DatacenterData `json:"datacenter"`
	Error      string               `json:"error"`
}

func adminCreateDatacenterHandler(w http.ResponseWriter, r *http.Request) {
	var response AdminCreateDatacenterResponse
	var datacenterData admin.DatacenterData
	err := json.NewDecoder(r.Body).Decode(&datacenterData)
	if err != nil {
		core.Error("failed to read datacenter data in create datacenter request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	datacenterId, err := controller.CreateDatacenter(&datacenterData)
	if err != nil {
		core.Error("failed to create datacenter: %v", err)
		response.Error = err.Error()
	} else {
		datacenterData.DatacenterId = datacenterId
		core.Log("create datacenter %x", datacenterId)
		core.Debug("%+v", datacenterData)
		response.Datacenter = datacenterData
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	} else {
		core.Debug("datacenters = %+v", datacenters)
	}
	w.WriteHeader(http.StatusOK)
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
		core.Error("read datacenter could not parse datacenter id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read datacenter %x", datacenterId)
	datacenter, err := controller.ReadDatacenter(datacenterId)
	response := AdminReadDatacenterResponse{Datacenter: datacenter}
	if err != nil {
		core.Error("failed to read datacenter: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", datacenter)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminUpdateDatacenterResponse struct {
	Datacenter admin.DatacenterData `json:"datacenter"`
	Error      string               `json:"error"`
}

func adminUpdateDatacenterHandler(w http.ResponseWriter, r *http.Request) {
	var datacenter admin.DatacenterData
	err := json.NewDecoder(r.Body).Decode(&datacenter)
	if err != nil {
		core.Error("failed to decode update datacenter request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("update datacenter %x", datacenter.DatacenterId)
	response := AdminUpdateDatacenterResponse{Datacenter: datacenter}
	err = controller.UpdateDatacenter(&datacenter)
	if err != nil {
		core.Error("failed to update datacenter: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", datacenter)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminDeleteDatacenterResponse struct {
	Error string `json:"error"`
}

func adminDeleteDatacenterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	datacenterId, err := strconv.ParseUint(vars["datacenterId"], 10, 64)
	if err != nil {
		core.Error("delete datacenter could not parse datacenter id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("delete datacenter %x", datacenterId)
	response := AdminDeleteDatacenterResponse{}
	err = controller.DeleteDatacenter(datacenterId)
	if err != nil {
		core.Error("failed to delete datacenter: %v", err)
		response.Error = err.Error()
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminCreateRelayResponse struct {
	Relay admin.RelayData `json:"relay"`
	Error string          `json:"error"`
}

func adminCreateRelayHandler(w http.ResponseWriter, r *http.Request) {
	var response AdminCreateRelayResponse
	var relayData admin.RelayData
	err := json.NewDecoder(r.Body).Decode(&relayData)
	if err != nil {
		core.Error("failed to read relay data in create relay request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	relayId, err := controller.CreateRelay(&relayData)
	if err != nil {
		core.Error("failed to create relay: %v", err)
		response.Error = err.Error()
	} else {
		relayData.RelayId = relayId
		core.Log("create relay %x", relayId)
		core.Debug("%+v", relayData)
		response.Relay = relayData
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadRelaysResponse struct {
	Relays []admin.RelayData `json:"relays"`
	Error  string            `json:"error"`
}

func adminReadRelaysHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("read relays")
	relays, err := controller.ReadRelays()
	response := AdminReadRelaysResponse{Relays: relays}
	if err != nil {
		core.Error("failed to read relays: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("relays = %+v", relays)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadRelayResponse struct {
	Relay admin.RelayData `json:"relay"`
	Error string          `json:"error"`
}

func adminReadRelayHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	relayId, err := strconv.ParseUint(vars["relayId"], 10, 64)
	if err != nil {
		core.Error("read relay could not parse relay id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read relay %x", relayId)
	relay, err := controller.ReadRelay(relayId)
	response := AdminReadRelayResponse{Relay: relay}
	if err != nil {
		core.Error("failed to read relay: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", relay)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminUpdateRelayResponse struct {
	Relay admin.RelayData `json:"relay"`
	Error string          `json:"error"`
}

func adminUpdateRelayHandler(w http.ResponseWriter, r *http.Request) {
	var relay admin.RelayData
	err := json.NewDecoder(r.Body).Decode(&relay)
	if err != nil {
		core.Error("failed to decode update relay request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("update relay %x", relay.RelayId)
	response := AdminUpdateRelayResponse{Relay: relay}
	err = controller.UpdateRelay(&relay)
	if err != nil {
		core.Error("failed to update relay: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", relay)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminDeleteRelayResponse struct {
	Error string `json:"error"`
}

func adminDeleteRelayHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	relayId, err := strconv.ParseUint(vars["relayId"], 10, 64)
	if err != nil {
		core.Error("delete relay could not parse relay id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("delete relay %x", relayId)
	response := AdminDeleteRelayResponse{}
	err = controller.DeleteRelay(relayId)
	if err != nil {
		core.Error("failed to delete relay: %v", err)
		response.Error = err.Error()
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminCreateRouteShaderResponse struct {
	RouteShader admin.RouteShaderData `json:"route_shader"`
	Error       string                `json:"error"`
}

func adminCreateRouteShaderHandler(w http.ResponseWriter, r *http.Request) {
	var response AdminCreateRouteShaderResponse
	var routeShaderData admin.RouteShaderData
	err := json.NewDecoder(r.Body).Decode(&routeShaderData)
	if err != nil {
		core.Error("failed to read route shader data in create route shader request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	routeShaderId, err := controller.CreateRouteShader(&routeShaderData)
	if err != nil {
		core.Error("failed to create route shader: %v", err)
		response.Error = err.Error()
	} else {
		routeShaderData.RouteShaderId = routeShaderId
		core.Log("create route shader %x", routeShaderId)
		core.Debug("%+v", routeShaderData)
		response.RouteShader = routeShaderData
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadRouteShadersResponse struct {
	RouteShaders []admin.RouteShaderData `json:"route_shaders"`
	Error        string                  `json:"error"`
}

func adminReadRouteShadersHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("read route shaders")
	routeShaders, err := controller.ReadRouteShaders()
	response := AdminReadRouteShadersResponse{RouteShaders: routeShaders}
	if err != nil {
		core.Error("failed to read route shaders: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("route shaders = %+v", routeShaders)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadRouteShaderResponse struct {
	RouteShader admin.RouteShaderData `json:"route_shader"`
	Error       string                `json:"error"`
}

func adminReadRouteShaderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	routeShaderId, err := strconv.ParseUint(vars["routeShaderId"], 10, 64)
	if err != nil {
		core.Error("read route shader could not parse route shader id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read route shader %x", routeShaderId)
	routeShader, err := controller.ReadRouteShader(routeShaderId)
	response := AdminReadRouteShaderResponse{RouteShader: routeShader}
	if err != nil {
		core.Error("failed to read route shader: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", routeShader)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminUpdateRouteShaderResponse struct {
	RouteShader admin.RouteShaderData `json:"route_shader"`
	Error       string                `json:"error"`
}

func adminUpdateRouteShaderHandler(w http.ResponseWriter, r *http.Request) {
	var routeShader admin.RouteShaderData
	err := json.NewDecoder(r.Body).Decode(&routeShader)
	if err != nil {
		core.Error("failed to decode update route shader request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("update route shader %x", routeShader.RouteShaderId)
	response := AdminUpdateRouteShaderResponse{RouteShader: routeShader}
	err = controller.UpdateRouteShader(&routeShader)
	if err != nil {
		core.Error("failed to update route shader: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", routeShader)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminDeleteRouteShaderResponse struct {
	Error string `json:"error"`
}

func adminDeleteRouteShaderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	routeShaderId, err := strconv.ParseUint(vars["routeShaderId"], 10, 64)
	if err != nil {
		core.Error("delete route shader could not parse route shader id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("delete route shader %x", routeShaderId)
	response := AdminDeleteRouteShaderResponse{}
	err = controller.DeleteRouteShader(routeShaderId)
	if err != nil {
		core.Error("failed to delete route shader: %v", err)
		response.Error = err.Error()
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminCreateBuyerDatacenterSettingsResponse struct {
	Settings admin.BuyerDatacenterSettings `json:"settings"`
	Error    string                        `json:"error"`
}

func adminCreateBuyerDatacenterSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var response AdminCreateBuyerDatacenterSettingsResponse
	var settings admin.BuyerDatacenterSettings
	err := json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
		core.Error("failed to read route shader data in create route shader request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = controller.CreateBuyerDatacenterSettings(&settings)
	if err != nil {
		core.Error("failed to create buyer datacenter settings: %v", err)
		response.Error = err.Error()
	} else {
		buyerId := settings.BuyerId
		datacenterId := settings.DatacenterId
		core.Log("create buyer datacenter settings %x.%x", buyerId, datacenterId)
		core.Debug("%+v", settings)
		response.Settings = settings
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadBuyerDatacenterSettingsListResponse struct {
	Settings []admin.BuyerDatacenterSettings `json:"settings"`
	Error    string                          `json:"error"`
}

func adminReadBuyerDatacenterSettingsListHandler(w http.ResponseWriter, r *http.Request) {
	buyerDatacenterSettings, err := controller.ReadBuyerDatacenterSettingsList()
	response := AdminReadBuyerDatacenterSettingsListResponse{Settings: buyerDatacenterSettings}
	if err != nil {
		response.Error = err.Error()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadBuyerDatacenterSettingsResponse struct {
	Settings admin.BuyerDatacenterSettings `json:"settings"`
	Error    string                        `json:"error"`
}

func adminReadBuyerDatacenterSettingsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buyerId, err := strconv.ParseUint(vars["buyerId"], 10, 64)
	if err != nil {
		core.Error("read buyer datacenter settings could not parse buyer id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	datacenterId, err := strconv.ParseUint(vars["datacenterId"], 10, 64)
	if err != nil {
		core.Error("read buyer datacenter settings could not parse datacenter id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read buyer datacenter settings %x.%x", buyerId, datacenterId)
	settings, err := controller.ReadBuyerDatacenterSettings(buyerId, datacenterId)
	response := AdminReadBuyerDatacenterSettingsResponse{Settings: settings}
	if err != nil {
		core.Error("failed to read buyer datacenter settings: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", settings)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminUpdateBuyerDatacenterSettingsResponse struct {
	Settings admin.BuyerDatacenterSettings `json:"settings"`
	Error    string                        `json:"error"`
}

func adminUpdateBuyerDatacenterSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var settings admin.BuyerDatacenterSettings
	err := json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
		core.Error("failed to decode update buyer datacenter settings request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("update buyer datacenter settings %x.%x", settings.BuyerId, settings.DatacenterId)
	response := AdminUpdateBuyerDatacenterSettingsResponse{Settings: settings}
	err = controller.UpdateBuyerDatacenterSettings(&settings)
	if err != nil {
		core.Error("failed to update buyer datacenter settings: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", settings)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminDeleteBuyerDatacenterSettingsResponse struct {
	Error string `json:"error"`
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
	core.Log("delete buyer datacenter settings %x.%x", buyerId, datacenterId)
	response := AdminDeleteBuyerDatacenterSettingsResponse{}
	err = controller.DeleteBuyerDatacenterSettings(buyerId, datacenterId)
	if err != nil {
		core.Error("failed to delete buyer datacenter settings: %v", err)
		response.Error = err.Error()
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminCreateBuyerKeypairResponse struct {
	BuyerKeypair admin.BuyerKeypairData `json:"buyer_keypair"`
	Error        string                 `json:"error"`
}

func adminCreateBuyerKeypairHandler(w http.ResponseWriter, r *http.Request) {
	var response AdminCreateBuyerKeypairResponse
	var buyerKeypairData admin.BuyerKeypairData
	err := json.NewDecoder(r.Body).Decode(&buyerKeypairData)
	if err != nil {
		core.Error("failed to read buyer keypair data in create buyer keypair request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	buyerKeypairId, err := controller.CreateBuyerKeypair(&buyerKeypairData)
	if err != nil {
		core.Error("failed to create buyer keypair: %v", err)
		response.Error = err.Error()
	} else {
		buyerKeypairData.BuyerKeypairId = buyerKeypairId
		core.Log("create buyer keypair %x", buyerKeypairId)
		core.Debug("%+v", buyerKeypairData)
		response.BuyerKeypair = buyerKeypairData
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadBuyerKeypairsResponse struct {
	BuyerKeypairs []admin.BuyerKeypairData `json:"buyer_keypairs"`
	Error         string                   `json:"error"`
}

func adminReadBuyerKeypairsHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("read buyer keypairs")
	buyerKeypairs, err := controller.ReadBuyerKeypairs()
	response := AdminReadBuyerKeypairsResponse{BuyerKeypairs: buyerKeypairs}
	if err != nil {
		core.Error("failed to read buyer keypairs: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("buyer keypairs = %+v", buyerKeypairs)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadBuyerKeypairResponse struct {
	BuyerKeypair admin.BuyerKeypairData `json:"buyer_keypair"`
	Error        string                 `json:"error"`
}

func adminReadBuyerKeypairHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buyerKeypairId, err := strconv.ParseUint(vars["buyerKeypairId"], 10, 64)
	if err != nil {
		core.Error("read buyer keypair could not parse buyer keypair id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read buyer keypair %x", buyerKeypairId)
	buyerKeypair, err := controller.ReadBuyerKeypair(buyerKeypairId)
	response := AdminReadBuyerKeypairResponse{BuyerKeypair: buyerKeypair}
	if err != nil {
		core.Error("failed to read buyer keypair: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", buyerKeypair)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminUpdateBuyerKeypairResponse struct {
	BuyerKeypair admin.BuyerKeypairData `json:"buyer_keypair"`
	Error        string                 `json:"error"`
}

func adminUpdateBuyerKeypairHandler(w http.ResponseWriter, r *http.Request) {
	var buyerKeypair admin.BuyerKeypairData
	err := json.NewDecoder(r.Body).Decode(&buyerKeypair)
	if err != nil {
		core.Error("failed to decode update buyer keypair request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("update buyer keypair %x", buyerKeypair.BuyerKeypairId)
	response := AdminUpdateBuyerKeypairResponse{BuyerKeypair: buyerKeypair}
	err = controller.UpdateBuyerKeypair(&buyerKeypair)
	if err != nil {
		core.Error("failed to update buyer keypair: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", buyerKeypair)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminDeleteBuyerKeypairResponse struct {
	Error string `json:"error"`
}

func adminDeleteBuyerKeypairHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buyerKeypairId, err := strconv.ParseUint(vars["buyerKeypairId"], 10, 64)
	if err != nil {
		core.Error("delete buyer keypair could not parse buyer keypair id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("delete buyer keypair %x", buyerKeypairId)
	response := AdminDeleteBuyerKeypairResponse{}
	err = controller.DeleteBuyerKeypair(buyerKeypairId)
	if err != nil {
		core.Error("failed to delete buyer keypair: %v", err)
		response.Error = err.Error()
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminCreateRelayKeypairResponse struct {
	RelayKeypair admin.RelayKeypairData `json:"relay_keypair"`
	Error        string                 `json:"error"`
}

func adminCreateRelayKeypairHandler(w http.ResponseWriter, r *http.Request) {
	var response AdminCreateRelayKeypairResponse
	var relayKeypairData admin.RelayKeypairData
	err := json.NewDecoder(r.Body).Decode(&relayKeypairData)
	if err != nil {
		core.Error("failed to read relay keypair data in create relay keypair request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	relayKeypairId, err := controller.CreateRelayKeypair(&relayKeypairData)
	if err != nil {
		core.Error("failed to create relay keypair: %v", err)
		response.Error = err.Error()
	} else {
		relayKeypairData.RelayKeypairId = relayKeypairId
		core.Log("create relay keypair %x", relayKeypairId)
		core.Debug("%+v", relayKeypairData)
		response.RelayKeypair = relayKeypairData
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadRelayKeypairsResponse struct {
	RelayKeypairs []admin.RelayKeypairData `json:"relay_keypairs"`
	Error         string                   `json:"error"`
}

func adminReadRelayKeypairsHandler(w http.ResponseWriter, r *http.Request) {
	core.Log("read relay keypairs")
	relayKeypairs, err := controller.ReadRelayKeypairs()
	response := AdminReadRelayKeypairsResponse{RelayKeypairs: relayKeypairs}
	if err != nil {
		core.Error("failed to read relay keypairs: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("relay keypairs = %+v", relayKeypairs)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminReadRelayKeypairResponse struct {
	RelayKeypair admin.RelayKeypairData `json:"relay_keypair"`
	Error        string                 `json:"error"`
}

func adminReadRelayKeypairHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	relayKeypairId, err := strconv.ParseUint(vars["relayKeypairId"], 10, 64)
	if err != nil {
		core.Error("read relay keypair could not parse relay keypair id")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("read relay keypair %x", relayKeypairId)
	relayKeypair, err := controller.ReadRelayKeypair(relayKeypairId)
	response := AdminReadRelayKeypairResponse{RelayKeypair: relayKeypair}
	if err != nil {
		core.Error("failed to read relay keypair: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", relayKeypair)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminUpdateRelayKeypairResponse struct {
	RelayKeypair admin.RelayKeypairData `json:"relay_keypair"`
	Error        string                 `json:"error"`
}

func adminUpdateRelayKeypairHandler(w http.ResponseWriter, r *http.Request) {
	var relayKeypair admin.RelayKeypairData
	err := json.NewDecoder(r.Body).Decode(&relayKeypair)
	if err != nil {
		core.Error("failed to decode update relay keypair request json: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("update relay keypair %x", relayKeypair.RelayKeypairId)
	response := AdminUpdateRelayKeypairResponse{RelayKeypair: relayKeypair}
	err = controller.UpdateRelayKeypair(&relayKeypair)
	if err != nil {
		core.Error("failed to update relay keypair: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("%+v", relayKeypair)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type AdminDeleteRelayKeypairResponse struct {
	Error string `json:"error"`
}

func adminDeleteRelayKeypairHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	relayKeypairId, err := strconv.ParseUint(vars["relayKeypairId"], 10, 64)
	if err != nil {
		core.Error("delete relay keypair could not parse relay keypair id: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	core.Log("delete relay keypair %x", relayKeypairId)
	response := AdminDeleteRelayKeypairResponse{}
	err = controller.DeleteRelayKeypair(relayKeypairId)
	if err != nil {
		core.Error("failed to delete relay keypair: %v", err)
		response.Error = err.Error()
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
