package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"time"
	// "strings"

	"github.com/networknext/next/modules/admin"
	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	db "github.com/networknext/next/modules/database"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/portal"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	// "github.com/golang-jwt/jwt/v5"
)

var redisPortalClient redis.Cmdable
var redisRelayBackendClient *redis.Client

var controller *admin.Controller

var service *common.Service

var privateKey string
var pgsqlConfig string
var databaseURL string
var sessionCruncherURL string
var serverCruncherURL string

var topSessionsWatcher *portal.TopSessionsWatcher
var topServersWatcher *portal.TopServersWatcher
var mapDataWatcher *portal.MapDataWatcher
var adminTimeSeriesWatcher *common.RedisTimeSeriesWatcher
var buyerTimeSeriesWatcher *common.RedisTimeSeriesWatcher
var relayTimeSeriesWatcher *common.RedisTimeSeriesWatcher
var adminCountersWatcher *common.RedisCountersWatcher
var buyerCountersWatcher *common.RedisCountersWatcher

var enableRedisTimeSeries bool

func main() {

	service = common.CreateService("api")

	enableRedisTimeSeries = envvar.GetBool("ENABLE_REDIS_TIME_SERIES", false)
	redisTimeSeriesCluster := envvar.GetStringArray("REDIS_TIME_SERIES_CLUSTER", []string{})
	redisTimeSeriesHostname := envvar.GetString("REDIS_TIME_SERIES_HOSTNAME", "127.0.0.1:6379")

	if enableRedisTimeSeries {
		core.Debug("redis time series cluster: %s", redisTimeSeriesCluster)
		core.Debug("redis time series hostname: %s", redisTimeSeriesHostname)
	}

	privateKey = envvar.GetString("API_PRIVATE_KEY", "")
	pgsqlConfig = envvar.GetString("PGSQL_CONFIG", "host=127.0.0.1 port=5432 user=developer password=developer dbname=postgres sslmode=disable")
	databaseURL = envvar.GetString("DATABASE_URL", "")
	sessionCruncherURL = envvar.GetString("SESSION_CRUNCHER_URL", "http://127.0.0.1:40200")
	serverCruncherURL = envvar.GetString("SERVER_CRUNCHER_URL", "http://127.0.0.1:40300")
	redisPortalCluster := envvar.GetStringArray("REDIS_PORTAL_CLUSTER", []string{})
	redisPortalHostname := envvar.GetString("REDIS_PORTAL_HOSTNAME", "127.0.0.1:6379")
	redisRelayBackendHostname := envvar.GetString("REDIS_RELAY_BACKEND_HOSTNAME", "127.0.0.1:6379")
	enableAdmin := envvar.GetBool("ENABLE_ADMIN", true)
	enablePortal := envvar.GetBool("ENABLE_PORTAL", true)
	enableDatabase := envvar.GetBool("ENABLE_DATABASE", true)

	if privateKey == "" {
		core.Error("You must specify API_PRIVATE_KEY!")
		os.Exit(1)
	}

	core.Debug("pgsql config: %s", pgsqlConfig)
	if databaseURL != "" {
		core.Debug("database url: %s", databaseURL)
	}
	if sessionCruncherURL != "" {
		core.Debug("session cruncher url: %s", sessionCruncherURL)
	}
	core.Debug("redis portal cluster: %s", redisPortalCluster)
	core.Debug("redis portal hostname: %s", redisPortalHostname)
	core.Debug("redis relay backend hostname: %s", redisRelayBackendHostname)
	core.Debug("enable admin: %v", enableAdmin)
	core.Debug("enable portal: %v", enablePortal)
	core.Debug("enable database: %v", enableDatabase)

	service.Router.HandleFunc("/ping", isAuthorized(pingHandler))

	if enableAdmin {

		controller = admin.CreateController(pgsqlConfig)

		service.Router.HandleFunc("/admin/database", isAuthorized(adminDatabaseHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/commit", isAuthorized(adminCommitHandler)).Methods("PUT")

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

		service.Router.HandleFunc("/admin/create_relay_keypair", isAuthorized(adminCreateRelayKeypairHandler)).Methods("POST")
		service.Router.HandleFunc("/admin/relay_keypairs", isAuthorized(adminReadRelayKeypairsHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/relay_keypair/{relayKeypairId}", isAuthorized(adminReadRelayKeypairHandler)).Methods("GET")
		service.Router.HandleFunc("/admin/update_relay_keypair", isAuthorized(adminUpdateRelayKeypairHandler)).Methods("PUT")
		service.Router.HandleFunc("/admin/delete_relay_keypair/{relayKeypairId}", isAuthorized(adminDeleteRelayKeypairHandler)).Methods("DELETE")
	}

	if enablePortal {

		topSessionsWatcher = portal.CreateTopSessionsWatcher(sessionCruncherURL)

		topServersWatcher = portal.CreateTopServersWatcher(serverCruncherURL)

		mapDataWatcher = portal.CreateMapDataWatcher(sessionCruncherURL)

		if enableRedisTimeSeries {

			// create admin time series watcher

			timeSeriesConfig := common.RedisTimeSeriesConfig{
				RedisHostname: redisTimeSeriesHostname,
				RedisCluster:  redisTimeSeriesCluster,
			}

			var err error
			adminTimeSeriesWatcher, err = common.CreateRedisTimeSeriesWatcher(service.Context, timeSeriesConfig)
			if err != nil {
				core.Error("could not create admin time series watcher: %v", err)
				os.Exit(1)
			}

			keys := []string{"accelerated_percent", "route_matrix_total_routes", "route_matrix_bytes", "route_matrix_optimize_ms"}

			adminTimeSeriesWatcher.SetKeys(keys)

			// create buyer time series watcher

			buyerTimeSeriesWatcher, err = common.CreateRedisTimeSeriesWatcher(service.Context, timeSeriesConfig)
			if err != nil {
				core.Error("could not create buyer time series watcher: %v", err)
				os.Exit(1)
			}

			go func(ctx context.Context) {
				ticker := time.NewTicker(time.Second)
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						database := service.Database()
						if database == nil {
							break
						}
						keys := []string{}
						buyerIds := database.GetBuyerIds()
						for i := range buyerIds {
							keys = append(keys, fmt.Sprintf("accelerated_percent_%016x", buyerIds[i]))
						}
						buyerTimeSeriesWatcher.SetKeys(keys)
					}
				}
			}(service.Context)

			// create relay time series watcher

			relayTimeSeriesWatcher, err = common.CreateRedisTimeSeriesWatcher(service.Context, timeSeriesConfig)
			if err != nil {
				core.Error("could not create relay time series watcher: %v", err)
				os.Exit(1)
			}

			go func(ctx context.Context) {
				ticker := time.NewTicker(time.Second)
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						database := service.Database()
						if database == nil {
							break
						}
						keys := []string{}
						relayIds := database.GetRelayIds()
						for i := range relayIds {
							keys = append(keys, fmt.Sprintf("relay_%016x_session_count", relayIds[i]))
							keys = append(keys, fmt.Sprintf("relay_%016x_packets_sent_per_second", relayIds[i]))
							keys = append(keys, fmt.Sprintf("relay_%016x_packets_received_per_second", relayIds[i]))
							keys = append(keys, fmt.Sprintf("relay_%016x_bandwidth_sent_kbps", relayIds[i]))
							keys = append(keys, fmt.Sprintf("relay_%016x_bandwidth_received_kbps", relayIds[i]))
						}
						relayTimeSeriesWatcher.SetKeys(keys)
					}
				}
			}(service.Context)

			// create the admin counters watcher

			countersConfig := common.RedisCountersConfig{
				RedisHostname: redisTimeSeriesHostname,
				RedisCluster:  redisTimeSeriesCluster,
			}

			adminCountersWatcher, err = common.CreateRedisCountersWatcher(service.Context, countersConfig)
			if err != nil {
				core.Error("could not create admin counters watcher: %v", err)
				os.Exit(1)
			}

			keys = []string{
				"session_update",
				"next_session_update",
				"server_update",
				"relay_update",
				"retry",
				"fallback_to_direct",
			}

			adminCountersWatcher.SetKeys(keys)

			// create the buyer counters watcher

			buyerCountersWatcher, err = common.CreateRedisCountersWatcher(service.Context, countersConfig)
			if err != nil {
				core.Error("could not create buyer counters watcher: %v", err)
				os.Exit(1)
			}

			go func(ctx context.Context) {
				ticker := time.NewTicker(time.Second)
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						database := service.Database()
						if database == nil {
							break
						}
						keys := []string{}
						buyerIds := database.GetBuyerIds()
						for i := range buyerIds {
							keys = append(keys, fmt.Sprintf("session_update_%016x", buyerIds[i]))
							keys = append(keys, fmt.Sprintf("next_session_update_%016x", buyerIds[i]))
							keys = append(keys, fmt.Sprintf("server_update_%016x", buyerIds[i]))
						}
						buyerCountersWatcher.SetKeys(keys)
					}
				}
			}(service.Context)
		}

		if len(redisPortalCluster) > 0 {
			redisPortalClient = common.CreateRedisClusterClient(redisPortalCluster)
		} else {
			redisPortalClient = common.CreateRedisClient(redisPortalHostname)
		}

		redisRelayBackendClient = common.CreateRedisClient(redisRelayBackendHostname)

		service.Router.HandleFunc("/portal/session_counts", isAuthorized(portalSessionCountsHandler))
		service.Router.HandleFunc("/portal/sessions", isAuthorized(portalSessionsHandler))
		service.Router.HandleFunc("/portal/sessions/{page}", isAuthorized(portalSessionsHandler))
		service.Router.HandleFunc("/portal/user_sessions/{user_hash}", isAuthorized(portalUserSessionsHandler))
		service.Router.HandleFunc("/portal/user_sessions/{user_hash}/{page}", isAuthorized(portalUserSessionsHandler))
		service.Router.HandleFunc("/portal/session/{session_id}", isAuthorized(portalSessionDataHandler))

		service.Router.HandleFunc("/portal/server_count", isAuthorized(portalServerCountHandler))
		service.Router.HandleFunc("/portal/servers/{page}", isAuthorized(portalServersHandler))
		service.Router.HandleFunc("/portal/server/{server_address}", isAuthorized(portalServerDataHandler))
		service.Router.HandleFunc("/portal/server/{server_address}/{page}", isAuthorized(portalServerDataHandler))

		service.Router.HandleFunc("/portal/relay_count", isAuthorized(portalRelayCountHandler))
		service.Router.HandleFunc("/portal/relays", isAuthorized(portalRelaysHandler))
		service.Router.HandleFunc("/portal/relays/{page}", isAuthorized(portalRelaysHandler))
		service.Router.HandleFunc("/portal/all_relays", isAuthorized(portalAllRelaysHandler))
		service.Router.HandleFunc("/portal/relay/{relay_name}", isAuthorized(portalRelayDataHandler))

		service.Router.HandleFunc("/portal/buyers", isAuthorized(portalBuyersHandler))
		service.Router.HandleFunc("/portal/buyers/{page}", isAuthorized(portalBuyersHandler))
		service.Router.HandleFunc("/portal/buyer/{buyer_code}", isAuthorized(portalBuyerDataHandler))

		service.Router.HandleFunc("/portal/sellers", isAuthorized(portalSellersHandler))
		service.Router.HandleFunc("/portal/sellers/{page}", isAuthorized(portalSellersHandler))
		service.Router.HandleFunc("/portal/seller/{seller_code}/{page}", isAuthorized(portalSellerDataHandler))

		service.Router.HandleFunc("/portal/datacenters", isAuthorized(portalDatacentersHandler))
		service.Router.HandleFunc("/portal/datacenters/{page}", isAuthorized(portalDatacentersHandler))
		service.Router.HandleFunc("/portal/datacenter/{datacenter_name}", isAuthorized(portalDatacenterDataHandler))

		service.Router.HandleFunc("/portal/map_data", isAuthorized(portalMapDataHandler))

		service.Router.HandleFunc("/portal/cost_matrix", isAuthorized(portalCostMatrixHandler))

		service.Router.HandleFunc("/portal/admin_data", isAuthorized(portalAdminDataHandler))
	}

	if enableDatabase {

		service.Router.HandleFunc("/database/json", isAuthorized(databaseJSONHandler)).Methods("GET")
		service.Router.HandleFunc("/database/binary", isAuthorized(databaseBinaryHandler)).Methods("GET")
		service.Router.HandleFunc("/database/header", isAuthorized(databaseHeaderHandler)).Methods("GET")
		service.Router.HandleFunc("/database/buyers", isAuthorized(databaseBuyersHandler)).Methods("GET")
		service.Router.HandleFunc("/database/sellers", isAuthorized(databaseSellersHandler)).Methods("GET")
		service.Router.HandleFunc("/database/datacenters", isAuthorized(databaseDatacentersHandler)).Methods("GET")
		service.Router.HandleFunc("/database/relays", isAuthorized(databaseRelaysHandler)).Methods("GET")
		service.Router.HandleFunc("/database/buyer_datacenter_settings", isAuthorized(databaseBuyerDatacenterSettingsHandler)).Methods("GET")
	}

	if enablePortal || enableDatabase {
		service.LoadDatabase() // needed by both portal and database REST APIs
	}

	service.StartWebServer()

	service.WaitForShutdown()
}

// ---------------------------------------------------------------------------------------------------------------------

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		endpoint(w, r)

		// todo: sort out auth
		/*
		auth := r.Header.Get("Authorization")

		split := strings.Split(auth, "Bearer ")

		if len(split) == 2 {

			apiKey := split[1]

			// todo: ParseWithClaims and check if "portal" or "admin" authorized. split into two function versions

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
		*/
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	if service.Tag != "" {
		w.Write([]byte(fmt.Sprintf("pong [%s]", service.Tag)))
	} else {
		w.Write([]byte(fmt.Sprintf("pong [%s]", service.Env)))
	}
}

// ---------------------------------------------------------------------------------------------------------------------

type PortalSessionCountsResponse struct {
	NextSessionCount  int `json:"next_session_count"`
	TotalSessionCount int `json:"total_session_count"`
}

func portalSessionCountsHandler(w http.ResponseWriter, r *http.Request) {
	response := PortalSessionCountsResponse{}
	adminCountersWatcher.Lock()
	sessionUpdate := adminCountersWatcher.GetFloatValue("session_update")
	nextSessionUpdate := adminCountersWatcher.GetFloatValue("next_session_update")
	adminCountersWatcher.Unlock()
	response.TotalSessionCount = int(math.Ceil(sessionUpdate * 10.0 / 60.0))
	response.NextSessionCount = int(math.Ceil(nextSessionUpdate * 10.0 / 60.0))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalSessionData struct {
	SessionId           uint64   `json:"session_id,string"`
	Score               uint32   `json:"score"`
	UserHash            uint64   `json:"user_hash,string"`
	StartTime           uint64   `json:"start_time,string"`
	ISP                 string   `json:"isp"`
	ConnectionType      uint8    `json:"connection_type"`
	PlatformType        uint8    `json:"platform_type"`
	Latitude            float32  `json:"latitude"`
	Longitude           float32  `json:"longitude"`
	DirectRTT           uint32   `json:"direct_rtt"`
	NextRTT             uint32   `json:"next_rtt"`
	BuyerId             uint64   `json:"buyer_id,string"`
	BuyerName           string   `json:"buyer_name"`
	BuyerCode           string   `json:"buyer_code"`
	DatacenterId        uint64   `json:"datacenter_id,string"`
	DatacenterName      string   `json:"datacenter_name"`
	ServerAddress       string   `json:"server_address"`
	NumRouteRelays      int      `json:"num_route_relays"`
	RouteRelayIds       []uint64 `json:"route_relay_ids,string"`
	RouteRelayNames     []string `json:"route_relay_names"`
	RouteRelayAddresses []string `json:"route_relay_addresses"`
}

func upgradePortalSessionData(database *db.Database, input *portal.SessionData, output *PortalSessionData) {
	output.SessionId = input.SessionId
	output.UserHash = input.UserHash
	output.StartTime = input.StartTime
	output.ISP = input.ISP
	output.ConnectionType = input.ConnectionType
	output.PlatformType = input.PlatformType
	output.Latitude = input.Latitude
	output.Longitude = input.Longitude
	output.DirectRTT = input.DirectRTT
	output.NextRTT = input.NextRTT
	output.BuyerId = input.BuyerId
	output.DatacenterId = input.DatacenterId
	output.ServerAddress = input.ServerAddress
	if database != nil {
		buyer := database.GetBuyer(input.BuyerId)
		if buyer != nil {
			output.BuyerName = buyer.Name
			output.BuyerCode = buyer.Code
		}
		datacenter := database.GetDatacenter(input.DatacenterId)
		if datacenter != nil {
			output.DatacenterName = datacenter.Name
		}
	}
	output.NumRouteRelays = input.NumRouteRelays
	output.RouteRelayIds = make([]uint64, output.NumRouteRelays)
	output.RouteRelayNames = make([]string, output.NumRouteRelays)
	output.RouteRelayAddresses = make([]string, output.NumRouteRelays)
	for i := 0; i < output.NumRouteRelays; i++ {
		output.RouteRelayIds[i] = input.RouteRelays[i]
		relay := database.GetRelay(input.RouteRelays[i])
		if relay != nil {
			output.RouteRelayNames[i] = relay.Name
			output.RouteRelayAddresses[i] = relay.PublicAddress.String()
		}
	}
	output.Score = core.GetSessionScore(input.NextRTT > 0, int32(input.DirectRTT), int32(input.NextRTT))
}

type PortalSessionsResponse struct {
	Sessions   []PortalSessionData `json:"sessions"`
	OutputPage int                 `json:"output_page"`
	NumPages   int                 `json:"num_pages"`
}

func portalSessionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, err := strconv.ParseInt(vars["page"], 10, 64)
	if err != nil {
		page = 0
	}
	response := PortalSessionsResponse{}
	sessionIds := topSessionsWatcher.GetTopSessions()
	begin, end, outputPage, numPages := core.DoPagination_Simple(int(page), len(sessionIds))
	sessionIds = sessionIds[begin:end]
	sessions := portal.GetSessionList(service.Context, redisPortalClient, sessionIds)
	response.Sessions = make([]PortalSessionData, len(sessions))
	response.OutputPage = outputPage
	response.NumPages = numPages
	database := service.Database()
	for i := range response.Sessions {
		upgradePortalSessionData(database, sessions[i], &response.Sessions[i])
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalUserSessionsResponse struct {
	Sessions   []PortalSessionData `json:"sessions"`
	OutputPage int                 `json:"output_page"`
	NumPages   int                 `json:"num_pages"`
}

func portalUserSessionsHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	userHash, err := strconv.ParseUint(vars["user_hash"], 16, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	page, err := strconv.ParseInt(vars["page"], 10, 64)
	if err != nil {
		page = 0
	}

	response := PortalUserSessionsResponse{}

	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sessions := portal.GetUserSessionList(service.Context, redisPortalClient, userHash, time.Now().Unix()/60, 1000)

	if sessions == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	begin, end, outputPage, numPages := core.DoPagination_Simple(int(page), len(sessions))
	sessions = sessions[begin:end]
	response.Sessions = make([]PortalSessionData, len(sessions))
	response.OutputPage = outputPage
	response.NumPages = numPages

	for i := range response.Sessions {
		upgradePortalSessionData(database, sessions[i], &response.Sessions[i])
	}

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

type NearRelayData struct {
	Timestamp           uint64                           `json:"timestamp,string"`
	NumNearRelays       uint32                           `json:"num_near_relays"`
	NearRelayName       [constants.MaxNearRelays]string  `json:"near_relay_name"`
	NearRelayId         [constants.MaxNearRelays]uint64  `json:"near_relay_id"`
	NearRelayRTT        [constants.MaxNearRelays]uint8   `json:"near_relay_rtt"`
	NearRelayJitter     [constants.MaxNearRelays]uint8   `json:"near_relay_jitter"`
	NearRelayPacketLoss [constants.MaxNearRelays]float32 `json:"near_relay_packet_loss"`
}

func upgradeNearRelayData(database *db.Database, input []portal.NearRelayData, output *[]NearRelayData) {
	*output = make([]NearRelayData, len(input))
	for i := range input {
		(*output)[i].Timestamp = input[i].Timestamp
		(*output)[i].NumNearRelays = input[i].NumNearRelays
		for j := 0; j < int(input[i].NumNearRelays); j++ {
			(*output)[i].NearRelayId[j] = input[i].NearRelayId[j]
			(*output)[i].NearRelayRTT[j] = input[i].NearRelayRTT[j]
			(*output)[i].NearRelayJitter[j] = input[i].NearRelayJitter[j]
			(*output)[i].NearRelayPacketLoss[j] = input[i].NearRelayPacketLoss[j]
			if database != nil {
				relay := database.GetRelay(input[i].NearRelayId[j])
				if relay != nil {
					(*output)[i].NearRelayName[j] = relay.Name
				}
			}
		}
	}
}

type PortalSessionDataResponse struct {
	SessionData   PortalSessionData  `json:"session_data"`
	SliceData     []portal.SliceData `json:"slice_data"`
	NearRelayData []NearRelayData    `json:"near_relay_data"`
}

func portalSessionDataHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	sessionId, err := strconv.ParseUint(vars["session_id"], 16, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := PortalSessionDataResponse{}

	sessionData, sliceData, nearRelayData := portal.GetSessionData(service.Context, redisPortalClient, sessionId)
	if sessionData == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	upgradePortalSessionData(database, sessionData, &response.SessionData)

	response.SliceData = sliceData

	upgradeNearRelayData(database, nearRelayData, &response.NearRelayData)

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
	adminCountersWatcher.Lock()
	serverUpdate := adminCountersWatcher.GetFloatValue("server_update")
	adminCountersWatcher.Unlock()
	response.ServerCount = int(math.Ceil(serverUpdate * 10.0 / 60.0))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalServerData struct {
	ServerAddress    string `json:"server_address"`
	SDKVersion_Major uint8  `json:"sdk_version_major"`
	SDKVersion_Minor uint8  `json:"sdk_version_minor"`
	SDKVersion_Patch uint8  `json:"sdk_version_patch"`
	BuyerId          uint64 `json:"buyer_id,string"`
	DatacenterId     uint64 `json:"datacenter_id,string"`
	NumSessions      uint32 `json:"num_sessions"`
	Uptime           uint64 `json:"uptime,string"`
	BuyerName        string `json:"buyer_name"`
	BuyerCode        string `json:"buyer_code"`
	DatacenterName   string `json:"datacenter_name"`
}

type PortalServersResponse struct {
	Servers    []PortalServerData `json:"servers"`
	OutputPage int                `json:"output_page"`
	NumPages   int                `json:"num_pages"`
}

func upgradePortalServer(database *db.Database, input *portal.ServerData, output *PortalServerData) {
	output.ServerAddress = input.ServerAddress
	output.SDKVersion_Major = input.SDKVersion_Major
	output.SDKVersion_Minor = input.SDKVersion_Minor
	output.SDKVersion_Patch = input.SDKVersion_Patch
	output.BuyerId = input.BuyerId
	output.DatacenterId = input.DatacenterId
	output.NumSessions = input.NumSessions
	output.Uptime = input.Uptime
	if database != nil {
		buyer := database.GetBuyer(output.BuyerId)
		if buyer != nil {
			output.BuyerName = buyer.Name
			output.BuyerCode = buyer.Code
		}
		datacenter := database.GetDatacenter(output.DatacenterId)
		if datacenter != nil {
			output.DatacenterName = datacenter.Name
		}
	}
}

func portalServersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, err := strconv.ParseInt(vars["page"], 10, 64)
	if err != nil {
		page = 0
	}
	serverAddresses := topServersWatcher.GetTopServers()
	begin, end, outputPage, numPages := core.DoPagination_Simple(int(page), len(serverAddresses))
	serverAddresses = serverAddresses[begin:end]
	servers := portal.GetServerList(service.Context, redisPortalClient, serverAddresses)
	response := PortalServersResponse{}
	response.Servers = make([]PortalServerData, len(servers))
	response.OutputPage = outputPage
	response.NumPages = numPages
	database := service.Database()
	for i := range servers {
		upgradePortalServer(database, servers[i], &response.Servers[i])
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalServerDataResponse struct {
	ServerData     PortalServerData    `json:"server_data"`
	ServerSessions []PortalSessionData `json:"server_sessions"`
	OutputPage     int                 `json:"output_page"`
	NumPages       int                 `json:"num_pages"`
}

func portalServerDataHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	serverAddress := vars["server_address"]
	page, err := strconv.ParseInt(vars["page"], 10, 64)
	if err != nil {
		page = 0
	}

	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	serverData, serverSessions := portal.GetServerData(service.Context, redisPortalClient, serverAddress, time.Now().Unix()/60)
	if serverData == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	begin, end, outputPage, numPages := core.DoPagination_Simple(int(page), len(serverSessions))

	response := PortalServerDataResponse{}

	response.ServerSessions = make([]PortalSessionData, len(serverSessions))

	for i := range response.ServerSessions {
		upgradePortalSessionData(database, serverSessions[i], &response.ServerSessions[i])
	}

	sort.Slice(response.ServerSessions, func(i, j int) bool {
		return response.ServerSessions[i].SessionId < response.ServerSessions[j].SessionId
	})

	sort.SliceStable(response.ServerSessions, func(i, j int) bool { return response.ServerSessions[i].Score < response.ServerSessions[j].Score })

	response.ServerSessions = response.ServerSessions[begin:end]

	upgradePortalServer(database, serverData, &response.ServerData)

	response.OutputPage = outputPage
	response.NumPages = numPages

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
	response.RelayCount = portal.GetRelayCount(service.Context, redisPortalClient, time.Now().Unix()/60)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalRelayData struct {
	RelayName                           string   `json:"relay_name"`
	RelayId                             uint64   `json:"relay_id,string"`
	RelayAddress                        string   `json:"relay_address"`
	NumSessions                         uint32   `json:"num_sessions"`
	MaxSessions                         uint32   `json:"max_sessions"`
	StartTime                           uint64   `json:"start_time,string"`
	RelayFlags                          uint64   `json:"relay_flags,string"`
	RelayVersion                        string   `json:"relay_version"`
	SellerId                            uint64   `json:"seller_id,string"`
	SellerName                          string   `json:"seller_name"`
	SellerCode                          string   `json:"seller_code"`
	DatacenterId                        uint64   `json:"datacenter_id,string"`
	DatacenterName                      string   `json:"datacenter_name"`
	Uptime                              uint64   `json:"uptime,string"`
	Latitude                            float32  `json:"latitude"`
	Longitude                           float32  `json:"longitude"`
	SessionCount_Timestamps             []uint64 `json:"session_count_timestamps,string"`
	SessionCount_Values                 []int    `json:"session_count_values"`
	BandwidthSentKbps_Timestamps        []uint64 `json:"bandwidth_sent_kbps_timestamps,string"`
	BandwidthSentKbps_Values            []int    `json:"bandwidth_sent_kbps_values"`
	BandwidthReceivedKbps_Timestamps    []uint64 `json:"bandwidth_received_kbps_timestamps,string"`
	BandwidthReceivedKbps_Values        []int    `json:"bandwidth_received_kbps_values"`
	PacketsSentPerSecond_Timestamps     []uint64 `json:"packets_sent_per_second_timestamps,string"`
	PacketsSentPerSecond_Values         []int    `json:"packets_sent_per_second_values"`
	PacketsReceivedPerSecond_Timestamps []uint64 `json:"packets_received_per_second_timestamps,string"`
	PacketsReceivedPerSecond_Values     []int    `json:"packets_received_per_second_values"`
}

func upgradePortalRelayData(database *db.Database, input *portal.RelayData, output *PortalRelayData, withTimeSeries bool) {
	output.RelayName = input.RelayName
	output.RelayId = input.RelayId
	output.RelayAddress = input.RelayAddress
	output.NumSessions = input.NumSessions
	output.MaxSessions = input.MaxSessions
	output.StartTime = input.StartTime
	output.RelayFlags = input.RelayFlags
	output.RelayVersion = input.RelayVersion
	currentTime := uint64(time.Now().Unix())
	if database != nil {
		relay := database.GetRelay(input.RelayId)
		if relay != nil {
			output.SellerId = relay.Seller.Id
			output.SellerName = relay.Seller.Name
			output.SellerCode = relay.Seller.Code
			output.DatacenterId = relay.Datacenter.Id
			output.DatacenterName = relay.Datacenter.Name
			output.Uptime = currentTime - output.StartTime
			output.Latitude = relay.Datacenter.Latitude
			output.Longitude = relay.Datacenter.Longitude
		}
	}
	if withTimeSeries {
		relayTimeSeriesWatcher.Lock()
		relayTimeSeriesWatcher.GetIntValues(&output.SessionCount_Timestamps, &output.SessionCount_Values, fmt.Sprintf("relay_%016x_session_count", input.RelayId))
		relayTimeSeriesWatcher.GetIntValues(&output.BandwidthSentKbps_Timestamps, &output.BandwidthSentKbps_Values, fmt.Sprintf("relay_%016x_bandwidth_sent_kbps", input.RelayId))
		relayTimeSeriesWatcher.GetIntValues(&output.BandwidthReceivedKbps_Timestamps, &output.BandwidthReceivedKbps_Values, fmt.Sprintf("relay_%016x_bandwidth_received_kbps", input.RelayId))
		relayTimeSeriesWatcher.GetIntValues(&output.PacketsSentPerSecond_Timestamps, &output.PacketsSentPerSecond_Values, fmt.Sprintf("relay_%016x_packets_sent_per_second", input.RelayId))
		relayTimeSeriesWatcher.GetIntValues(&output.PacketsReceivedPerSecond_Timestamps, &output.PacketsReceivedPerSecond_Values, fmt.Sprintf("relay_%016x_packets_received_per_second", input.RelayId))
		relayTimeSeriesWatcher.Unlock()
	}
}

type PortalRelaysResponse struct {
	Relays     []PortalRelayData `json:"relays"`
	OutputPage int               `json:"output_page"`
	NumPages   int               `json:"num_pages"`
}

func portalRelaysHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, err := strconv.ParseInt(vars["page"], 10, 64)
	if err != nil {
		page = 0
	}
	relayAddresses := portal.GetRelayAddresses(service.Context, redisPortalClient, time.Now().Unix()/60, 0, constants.MaxRelays)
	relays := portal.GetRelayList(service.Context, redisPortalClient, relayAddresses)
	begin, end, outputPage, numPages := core.DoPagination_Simple(int(page), len(relays))
	sort.Slice(relays, func(i, j int) bool { return relays[i].RelayName < relays[j].RelayName })
	sort.SliceStable(relays, func(i, j int) bool { return relays[i].NumSessions > relays[j].NumSessions })
	relays = relays[begin:end]
	response := PortalRelaysResponse{}
	database := service.Database()
	response.Relays = make([]PortalRelayData, len(relays))
	response.OutputPage = outputPage
	response.NumPages = numPages
	for i := range response.Relays {
		upgradePortalRelayData(database, relays[i], &response.Relays[i], false)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func portalAllRelaysHandler(w http.ResponseWriter, r *http.Request) {
	relayAddresses := portal.GetRelayAddresses(service.Context, redisPortalClient, time.Now().Unix()/60, 0, constants.MaxRelays)
	relays := portal.GetRelayList(service.Context, redisPortalClient, relayAddresses)
	sort.Slice(relays, func(i, j int) bool { return relays[i].RelayName < relays[j].RelayName })
	sort.SliceStable(relays, func(i, j int) bool { return relays[i].NumSessions > relays[j].NumSessions })
	response := PortalRelaysResponse{}
	database := service.Database()
	response.Relays = make([]PortalRelayData, len(relays))
	for i := range response.Relays {
		upgradePortalRelayData(database, relays[i], &response.Relays[i], false)
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalRelayDataResponse struct {
	RelayData PortalRelayData `json:"relay_data"`
}

func portalRelayDataHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	relayName := vars["relay_name"]

	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	relay := database.GetRelayByName(relayName)
	if relay == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := PortalRelayDataResponse{}

	relayAddress := relay.PublicAddress.String()

	relayData := portal.GetRelayData(service.Context, redisPortalClient, relayAddress)
	if relayData == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	upgradePortalRelayData(database, relayData, &response.RelayData, true)

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type PortalBuyer struct {
	Id                            uint64            `json:"id,string"`
	Name                          string            `json:"name"`
	Code                          string            `json:"code"`
	Live                          bool              `json:"live"`
	Debug                         bool              `json:"debug"`
	PublicKey                     []byte            `json:"public_key"`
	RouteShader                   *core.RouteShader `json:"route_shader"`
	TotalSessions                 int               `json:"total_sessions"`
	NextSessions                  int               `json:"next_sessions"`
	ServerCount                   int               `json:"server_count"`
	TotalSessions_Timestamps      []uint64          `json:"total_sessions_timestamps,string"`
	TotalSessions_Values          []float32         `json:"total_sessions_values"`
	NextSessions_Timestamps       []uint64          `json:"next_sessions_timestamps,string"`
	NextSessions_Values           []float32         `json:"next_sessions_values"`
	AcceleratedPercent_Timestamps []uint64          `json:"accelerated_percent_timestamps,string"`
	AcceleratedPercent_Values     []float32         `json:"accelerated_percent_values"`
	ServerCount_Timestamps        []uint64          `json:"server_count_timestamps,string"`
	ServerCount_Values            []float32         `json:"server_count_values"`
}

type PortalBuyersResponse struct {
	Buyers     []PortalBuyer `json:"buyers"`
	OutputPage int           `json:"output_page"`
	NumPages   int           `json:"num_pages"`
}

func upgradePortalBuyer(input *db.Buyer, output *PortalBuyer, withRouteShader bool, withTimeSeries bool) {

	output.Id = input.Id
	output.Name = input.Name
	output.Code = input.Code
	output.Live = input.Live
	output.Debug = input.Debug
	output.PublicKey = input.PublicKey

	if enableRedisTimeSeries {
		buyerCountersWatcher.Lock()
		sessionUpdates := buyerCountersWatcher.GetFloatValue(fmt.Sprintf("session_update_%016x", input.Id))
		nextSessionUpdates := buyerCountersWatcher.GetFloatValue(fmt.Sprintf("next_session_update_%016x", input.Id))
		serverUpdates := buyerCountersWatcher.GetFloatValue(fmt.Sprintf("server_update_%016x", input.Id))
		buyerCountersWatcher.Unlock()

		output.TotalSessions = int(math.Ceil(sessionUpdates * 10.0 / 60.0))
		output.NextSessions = int(math.Ceil(nextSessionUpdates * 10.0 / 60.0))
		output.ServerCount = int(math.Ceil(serverUpdates * 10.0 / 60.0))
	}

	if withRouteShader {
		output.RouteShader = &input.RouteShader
	}

	if enableRedisTimeSeries && withTimeSeries {

		buyerCountersWatcher.Lock()
		buyerCountersWatcher.GetFloat32Values(&output.TotalSessions_Timestamps, &output.TotalSessions_Values, fmt.Sprintf("session_update_%016x", input.Id))
		buyerCountersWatcher.GetFloat32Values(&output.NextSessions_Timestamps, &output.NextSessions_Values, fmt.Sprintf("next_session_update_%016x", input.Id))
		buyerCountersWatcher.GetFloat32Values(&output.ServerCount_Timestamps, &output.ServerCount_Values, fmt.Sprintf("server_update_%016x", input.Id))
		buyerCountersWatcher.Unlock()

		buyerTimeSeriesWatcher.Lock()
		buyerTimeSeriesWatcher.GetFloat32Values(&output.AcceleratedPercent_Timestamps, &output.AcceleratedPercent_Values, fmt.Sprintf("accelerated_percent_%016x", input.Id))
		buyerTimeSeriesWatcher.Unlock()

		for i := range output.TotalSessions_Values {
			output.TotalSessions_Values[i] *= 10.0 / 60.0
		}

		for i := range output.NextSessions_Values {
			output.NextSessions_Values[i] *= 10.0 / 60.0
		}

		for i := range output.ServerCount_Values {
			output.ServerCount_Values[i] *= 10.0 / 60.0
		}
	}
}

func portalBuyersHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	page, err := strconv.ParseInt(vars["page"], 10, 64)
	if err != nil {
		page = 0
	}

	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	database_response := database.GetBuyers()

	buyers := database_response.Buyers

	response := PortalBuyersResponse{}
	response.Buyers = make([]PortalBuyer, len(buyers))
	for i := range buyers {
		upgradePortalBuyer(&buyers[i], &response.Buyers[i], false, false)
	}

	sort.Slice(response.Buyers, func(i, j int) bool { return response.Buyers[i].Name < response.Buyers[j].Name })
	sort.SliceStable(response.Buyers, func(i, j int) bool { return response.Buyers[i].TotalSessions > response.Buyers[j].TotalSessions })

	begin, end, outputPage, numPages := core.DoPagination_Simple(int(page), len(buyers))

	response.Buyers = response.Buyers[begin:end]

	response.OutputPage = outputPage

	response.NumPages = numPages

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

type PortalBuyerDataResponse struct {
	BuyerData PortalBuyer `json:"buyer_data"`
}

func portalBuyerDataHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	buyerCode := vars["buyer_code"]

	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := PortalBuyerDataResponse{}

	buyer := database.GetBuyerByCode(buyerCode)
	if buyer == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	upgradePortalBuyer(buyer, &response.BuyerData, true, true)

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type PortalSellersResponse struct {
	Sellers    []db.Seller `json:"sellers"`
	OutputPage int         `json:"output_page"`
	NumPages   int         `json:"num_pages"`
}

func portalSellersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, err := strconv.ParseInt(vars["page"], 10, 64)
	if err != nil {
		page = 0
	}
	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	database_response := database.GetSellers()
	sellers := database_response.Sellers
	sort.Slice(sellers, func(i, j int) bool { return sellers[i].Name < sellers[j].Name })
	begin, end, outputPage, numPages := core.DoPagination_Simple(int(page), len(sellers))
	sellers = sellers[begin:end]
	response := PortalSellersResponse{}
	response.Sellers = sellers
	response.OutputPage = outputPage
	response.NumPages = numPages
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalSellerDataResponse struct {
	SellerData *db.Seller        `json:"seller_data"`
	Relays     []PortalRelayData `json:"relays"`
	OutputPage int               `json:"output_page"`
	NumPages   int               `json:"num_pages"`
}

func portalSellerDataHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	sellerCode := vars["seller_code"]
	page, err := strconv.ParseInt(vars["page"], 10, 64)
	if err != nil {
		page = 0
	}

	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := PortalSellerDataResponse{}
	response.SellerData = database.GetSellerByCode(sellerCode)
	if response.SellerData == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	relayAddresses := portal.GetRelayAddresses(service.Context, redisPortalClient, time.Now().Unix()/60, 0, constants.MaxRelays)
	rawRelays := portal.GetRelayList(service.Context, redisPortalClient, relayAddresses)
	relays := make([]portal.RelayData, 0, len(rawRelays))
	for i := range rawRelays {
		relay := database.GetRelay(rawRelays[i].RelayId)
		if relay != nil && relay.Seller.Code == sellerCode {
			relays = append(relays, *rawRelays[i])
		}
	}
	begin, end, outputPage, numPages := core.DoPagination_Simple(int(page), len(relays))
	sort.Slice(relays, func(i, j int) bool { return relays[i].RelayName < relays[j].RelayName })
	sort.SliceStable(relays, func(i, j int) bool { return relays[i].NumSessions > relays[j].NumSessions })
	relays = relays[begin:end]

	response.Relays = make([]PortalRelayData, len(relays))
	for i := range response.Relays {
		upgradePortalRelayData(database, &relays[i], &response.Relays[i], false)
	}

	response.OutputPage = outputPage
	response.NumPages = numPages

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type PortalDatacenterData struct {
	Id         uint64  `json:"id,string"`
	Name       string  `json:"name"`
	Native     string  `json:"native"`
	Latitude   float32 `json:"latitude"`
	Longitude  float32 `json:"longitude"`
	SellerId   uint64  `json:"seller_id,string"`
	SellerCode string  `json:"seller_code"`
	SellerName string  `json:"seller_name"`
}

type PortalDatacentersResponse struct {
	Datacenters []PortalDatacenterData `json:"datacenters"`
	OutputPage  int                    `json:"output_page"`
	NumPages    int                    `json:"num_pages"`
}

func portalDatacentersHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	page, err := strconv.ParseInt(vars["page"], 10, 64)
	if err != nil {
		page = 0
	}
	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	database_response := database.GetDatacenters()
	datacenters := database_response.Datacenters
	sort.Slice(datacenters, func(i, j int) bool { return datacenters[i].Name < datacenters[j].Name })
	begin, end, outputPage, numPages := core.DoPagination_Simple(int(page), len(datacenters))
	datacenters = datacenters[begin:end]
	response := PortalDatacentersResponse{}
	response.Datacenters = make([]PortalDatacenterData, len(datacenters))
	for i := range response.Datacenters {
		response.Datacenters[i].Id = datacenters[i].Id
		response.Datacenters[i].Name = datacenters[i].Name
		response.Datacenters[i].Native = datacenters[i].Native
		response.Datacenters[i].Latitude = datacenters[i].Latitude
		response.Datacenters[i].Longitude = datacenters[i].Longitude
		response.Datacenters[i].SellerId = datacenters[i].SellerId
		seller := database.GetSeller(datacenters[i].SellerId)
		if seller != nil {
			response.Datacenters[i].SellerName = seller.Name
			response.Datacenters[i].SellerCode = seller.Code
		}
	}
	response.OutputPage = outputPage
	response.NumPages = numPages
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type PortalDatacenterDataResponse struct {
	DatacenterData PortalDatacenterData `json:"datacenter_data"`
	Relays         []PortalRelayData    `json:"relays"`
	OutputPage     int                  `json:"output_page"`
	NumPages       int                  `json:"num_pages"`
}

func portalDatacenterDataHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	datacenterName := vars["datacenter_name"]

	database := service.Database()
	if database == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	datacenter := database.GetDatacenterByName(datacenterName)
	if datacenter == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := PortalDatacenterDataResponse{}
	response.OutputPage = 0
	response.NumPages = 1

	response.DatacenterData.Id = datacenter.Id
	response.DatacenterData.Name = datacenter.Name
	response.DatacenterData.Native = datacenter.Native
	response.DatacenterData.Latitude = datacenter.Latitude
	response.DatacenterData.Longitude = datacenter.Longitude
	response.DatacenterData.SellerId = datacenter.SellerId

	seller := database.GetSeller(datacenter.SellerId)
	if seller != nil {
		response.DatacenterData.SellerName = seller.Name
		response.DatacenterData.SellerCode = seller.Code
	}

	datacenterRelayIds := database.GetDatacenterRelays(datacenter.Id)
	datacenterRelays := make([]*db.Relay, len(datacenterRelayIds))
	for i := range datacenterRelayIds {
		datacenterRelays[i] = database.GetRelay(datacenterRelayIds[i])
	}

	datacenterRelayAddresses := make([]string, len(datacenterRelays))
	for i := range datacenterRelays {
		datacenterRelayAddresses[i] = datacenterRelays[i].PublicAddress.String()
	}

	relays := portal.GetRelayList(service.Context, redisPortalClient, datacenterRelayAddresses)

	sort.Slice(relays, func(i, j int) bool { return relays[i].RelayName < relays[j].RelayName })
	sort.SliceStable(relays, func(i, j int) bool { return relays[i].NumSessions > relays[j].NumSessions })

	response.Relays = make([]PortalRelayData, len(datacenterRelays))
	for i := range datacenterRelays {
		upgradePortalRelayData(database, relays[i], &response.Relays[i], false)
	}

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

func portalMapDataHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	data := mapDataWatcher.GetMapData()
	w.Write(data)
}

// ---------------------------------------------------------------------------------------------------------------------

func portalCostMatrixHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	data := common.LoadMasterServiceData(service.Context, redisRelayBackendClient, "relay_backend", "cost_matrix")
	w.Write(data)
}

// ---------------------------------------------------------------------------------------------------------------------

type PortalAdminDataResponse struct {
	AcceleratedPercent_Timestamps []uint64  `json:"accelerated_percent_timestamps,string"`
	AcceleratedPercent_Values     []float32 `json:"accelerated_percent_values"`
	TotalRoutes_Timestamps        []uint64  `json:"total_routes_timestamps,string"`
	TotalRoutes_Values            []int     `json:"total_routes_values"`
	RouteMatrixBytes_Timestamps   []uint64  `json:"route_matrix_bytes_timestamps,string"`
	RouteMatrixBytes_Values       []int     `json:"route_matrix_bytes_values"`
	OptimizeMs_Timestamps         []uint64  `json:"optimize_ms_timestamps,string"`
	OptimizeMs_Values             []float32 `json:"optimize_ms_values"`

	TotalSessions_Timestamps    []uint64  `json:"total_sessions_timestamps,string"`
	TotalSessions_Values        []float32 `json:"total_sessions_values"`
	NextSessions_Timestamps     []uint64  `json:"next_sessions_timestamps,string"`
	NextSessions_Values         []float32 `json:"next_sessions_values"`
	ServerCount_Timestamps      []uint64  `json:"server_count_timestamps,string"`
	ServerCount_Values          []float32 `json:"server_count_values"`
	ActiveRelays_Timestamps     []uint64  `json:"active_relays_timestamps,string"`
	ActiveRelays_Values         []float32 `json:"active_relays_values"`
	Retry_Timestamps            []uint64  `json:"retry_timestamps,string"`
	Retry_Values                []int     `json:"retry_values"`
	FallbackToDirect_Timestamps []uint64  `json:"fallback_to_direct_timestamps,string"`
	FallbackToDirect_Values     []int     `json:"fallback_to_direct_values"`
}

func portalAdminDataHandler(w http.ResponseWriter, r *http.Request) {

	response := PortalAdminDataResponse{}

	if enableRedisTimeSeries {

		adminTimeSeriesWatcher.Lock()
		adminTimeSeriesWatcher.GetFloat32Values(&response.AcceleratedPercent_Timestamps, &response.AcceleratedPercent_Values, "accelerated_percent")
		adminTimeSeriesWatcher.GetIntValues(&response.TotalRoutes_Timestamps, &response.TotalRoutes_Values, "route_matrix_total_routes")
		adminTimeSeriesWatcher.GetIntValues(&response.RouteMatrixBytes_Timestamps, &response.RouteMatrixBytes_Values, "route_matrix_bytes")
		adminTimeSeriesWatcher.GetFloat32Values(&response.OptimizeMs_Timestamps, &response.OptimizeMs_Values, "route_matrix_optimize_ms")
		adminTimeSeriesWatcher.Unlock()

		adminCountersWatcher.Lock()
		adminCountersWatcher.GetFloat32Values(&response.TotalSessions_Timestamps, &response.TotalSessions_Values, "session_update")
		adminCountersWatcher.GetFloat32Values(&response.NextSessions_Timestamps, &response.NextSessions_Values, "next_session_update")
		adminCountersWatcher.GetFloat32Values(&response.ServerCount_Timestamps, &response.ServerCount_Values, "server_update")
		adminCountersWatcher.GetFloat32Values(&response.ActiveRelays_Timestamps, &response.ActiveRelays_Values, "relay_update")
		adminCountersWatcher.GetIntValues(&response.Retry_Timestamps, &response.Retry_Values, "retry")
		adminCountersWatcher.GetIntValues(&response.FallbackToDirect_Timestamps, &response.FallbackToDirect_Values, "fallback_to_direct")
		adminCountersWatcher.Unlock()

		for i := range response.TotalSessions_Values {
			response.TotalSessions_Values[i] *= 10.0 / 60.0
		}

		for i := range response.NextSessions_Values {
			response.NextSessions_Values[i] *= 10.0 / 60.0
		}

		for i := range response.ServerCount_Values {
			response.ServerCount_Values[i] *= 10.0 / 60.0
		}

		for i := range response.ActiveRelays_Values {
			response.ActiveRelays_Values[i] /= 60.0
		}
	}

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminDatabaseResponse struct {
	Database string `json:"database_base64"`
	Error    string `json:"error"`
}

func adminDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	var response AdminDatabaseResponse
	database, err := db.ExtractDatabase(pgsqlConfig)
	if err != nil {
		fmt.Printf("error: failed to extract database: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = database.Validate()
	if err != nil {
		fmt.Printf("error: database did not validate: %v\n", err)
		response.Error = fmt.Sprintf("error: database did not validate: %v\n", err)
	} else {
		response.Database = base64.StdEncoding.EncodeToString(database.GetBinary())
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------------------------------------------------------------------

type AdminCommitRequest struct {
	User     string `json:"user"`
	Database string `json:"database_base64"`
}

type AdminCommitResponse struct {
	Error string `json:"error"`
}

func bash(command string) bool {
	cmd := exec.Command("bash", "-c", command)
	if cmd == nil {
		fmt.Printf("error: could not run bash!\n")
		return false
	}
	if err := cmd.Run(); err != nil {
		fmt.Printf("error: failed to run command: %v\n", err)
		return false
	}
	cmd.Wait()
	return true
}

func adminCommitHandler(w http.ResponseWriter, r *http.Request) {
	var request AdminCommitRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		core.Error("failed to read commit request data in commit handler: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	database_binary, err := base64.StdEncoding.DecodeString(request.Database)
	if err != nil {
		core.Error("failed to decode database base64 string: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var response AdminCommitResponse
	if databaseURL == "" {
		core.Error("DATABASE_URL env var is not set. We have nowhere to write the database.bin to")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tempFileA := fmt.Sprintf("/tmp/database-%s.bin", common.RandomString(64))
	err = os.WriteFile(tempFileA, database_binary, 0666)
	if err != nil {
		core.Error("could not write database binary data to temp file '%s'", tempFileA)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	database, err := db.LoadDatabase(tempFileA)
	if err != nil {
		core.Error("could not load database from binary data in temp file '%s'", tempFileA)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = database.Validate()
	if err != nil {
		response.Error = fmt.Sprintf("error: database did not validate: %v\n", err)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
	database.Creator = request.User
	tempFileB := fmt.Sprintf("/tmp/database-%s.bin", common.RandomString(64))
	err = database.Save(tempFileB)
	if err != nil {
		core.Error("could not save database to temp file '%s'", tempFileB)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !bash(fmt.Sprintf("gsutil cp %s %s", tempFileB, databaseURL)) {
		core.Error("could not upload database.bin to database bucket")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	core.Log("committed database to %s for %s at time %s", databaseURL, request.User, database.CreationTime)
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
		core.Debug("create seller %d -> %+v", sellerId, sellerData)
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
	sellers, err := controller.ReadSellers()
	response := AdminReadSellersResponse{Sellers: sellers}
	if err != nil {
		core.Error("failed to read sellers: %v", err)
		response.Error = err.Error()
	}
	core.Debug("read sellers -> %+v", sellers)
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
	seller, err := controller.ReadSeller(sellerId)
	response := AdminReadSellerResponse{Seller: seller}
	if err != nil {
		response.Error = err.Error()
	}
	core.Debug("read seller %d -> %+v", sellerId, seller)
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
	response := AdminUpdateSellerResponse{Seller: seller}
	err = controller.UpdateSeller(&seller)
	if err != nil {
		core.Error("failed to update seller: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("update seller %d -> %+v", seller.SellerId, seller)
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
	core.Debug("delete seller %d", sellerId)
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
		core.Debug("create buyer %d -> %+v", buyerId, buyerData)
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
		core.Error("failed to read buyers: %v", err)
		response.Error = err.Error()
	}
	core.Debug("read buyers -> %+v", buyers)
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
	buyer, err := controller.ReadBuyer(buyerId)
	response := AdminReadBuyerResponse{Buyer: buyer}
	if err != nil {
		core.Error("failed to read buyer: %v", err)
		response.Error = err.Error()
	}
	core.Debug("read buyer %d -> %+v", buyerId, buyer)
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
	response := AdminUpdateBuyerResponse{Buyer: buyer}
	err = controller.UpdateBuyer(&buyer)
	if err != nil {
		core.Error("failed to update buyer: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("update buyer %d -> %+v", buyer.BuyerId, buyer)
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
	core.Debug("delete buyer %d", buyerId)
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
		core.Debug("create datacenter %d -> %+v", datacenterId, datacenterData)
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
	datacenters, err := controller.ReadDatacenters()
	response := AdminReadDatacentersResponse{Datacenters: datacenters}
	if err != nil {
		core.Error("failed to read datacenters: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("read datacenters -> %+v", datacenters)
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
	datacenter, err := controller.ReadDatacenter(datacenterId)
	response := AdminReadDatacenterResponse{Datacenter: datacenter}
	if err != nil {
		core.Error("failed to read datacenter: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("read datacenter %d -> %+v", datacenterId, datacenter)
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
	response := AdminUpdateDatacenterResponse{Datacenter: datacenter}
	err = controller.UpdateDatacenter(&datacenter)
	if err != nil {
		core.Error("failed to update datacenter: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("update datacenter %d -> %+v", datacenter.DatacenterId, datacenter)
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
	core.Debug("delete datacenter %d", datacenterId)
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
		core.Debug("create relay %d -> %+v", relayId, relayData)
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
	relays, err := controller.ReadRelays()
	response := AdminReadRelaysResponse{Relays: relays}
	if err != nil {
		core.Error("failed to read relays: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("read relays -> %+v", relays)
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
	relay, err := controller.ReadRelay(relayId)
	response := AdminReadRelayResponse{Relay: relay}
	if err != nil {
		core.Error("failed to read relay: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("read relay %d -> %+v", relayId, relay)
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
	response := AdminUpdateRelayResponse{Relay: relay}
	err = controller.UpdateRelay(&relay)
	if err != nil {
		core.Error("failed to update relay: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("update relay %d -> %+v", relay.RelayId, relay)
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
	core.Debug("delete relay %d", relayId)
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
		core.Debug("create route shader %d -> %+v", routeShaderId, routeShaderData)
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
	routeShaders, err := controller.ReadRouteShaders()
	response := AdminReadRouteShadersResponse{RouteShaders: routeShaders}
	if err != nil {
		core.Error("failed to read route shaders: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("read route shaders -> %+v", routeShaders)
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
	routeShader, err := controller.ReadRouteShader(routeShaderId)
	response := AdminReadRouteShaderResponse{RouteShader: routeShader}
	if err != nil {
		core.Error("failed to read route shader: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("read route shader %d -> %+v", routeShaderId, routeShader)
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
	response := AdminUpdateRouteShaderResponse{RouteShader: routeShader}
	err = controller.UpdateRouteShader(&routeShader)
	if err != nil {
		core.Error("failed to update route shader: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("update route shader %d -> %+v", routeShader.RouteShaderId, routeShader)
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
	core.Debug("delete route shader %d", routeShaderId)
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
		core.Debug("create buyer datacenter settings %d.%d -> %+v", buyerId, datacenterId, settings)
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
	settings, err := controller.ReadBuyerDatacenterSettings(buyerId, datacenterId)
	response := AdminReadBuyerDatacenterSettingsResponse{Settings: settings}
	if err != nil {
		core.Error("failed to read buyer datacenter settings: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("read buyer datacenter settings %d.%d -> %+v", buyerId, datacenterId, settings)
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
	response := AdminUpdateBuyerDatacenterSettingsResponse{Settings: settings}
	err = controller.UpdateBuyerDatacenterSettings(&settings)
	if err != nil {
		core.Error("failed to update buyer datacenter settings: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("update buyer datacenter settings %d.%d -> %+v", settings.BuyerId, settings.DatacenterId, settings)
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
	core.Debug("delete buyer datacenter settings %d.%d", buyerId, datacenterId)
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

type AdminCreateRelayKeypairResponse struct {
	RelayKeypair admin.RelayKeypairData `json:"relay_keypair"`
	Error        string                 `json:"error"`
}

func adminCreateRelayKeypairHandler(w http.ResponseWriter, r *http.Request) {
	var response AdminCreateRelayKeypairResponse
	relayKeypairData, err := controller.CreateRelayKeypair()
	if err != nil {
		core.Error("failed to create relay keypair: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("create relay keypair -> %+v", relayKeypairData)
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
	relayKeypairs, err := controller.ReadRelayKeypairs()
	response := AdminReadRelayKeypairsResponse{RelayKeypairs: relayKeypairs}
	if err != nil {
		core.Error("failed to read relay keypairs: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("get relay keypairs -> %+v", relayKeypairs)
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
	relayKeypair, err := controller.ReadRelayKeypair(relayKeypairId)
	response := AdminReadRelayKeypairResponse{RelayKeypair: relayKeypair}
	if err != nil {
		core.Error("failed to read relay keypair: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("read relay keypair %d -> %+v", relayKeypairId, relayKeypair)
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
	response := AdminUpdateRelayKeypairResponse{RelayKeypair: relayKeypair}
	err = controller.UpdateRelayKeypair(&relayKeypair)
	if err != nil {
		core.Error("failed to update relay keypair: %v", err)
		response.Error = err.Error()
	} else {
		core.Debug("update relay keypair %d -> %+v", relayKeypair.RelayKeypairId, relayKeypair)
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
	core.Debug("delete relay keypair %d", relayKeypairId)
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
