package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/networknext/backend/modules-old/transport/middleware"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/rs/cors"
)

/* todo
const (
	BASE_URL                = "https://networknextexternal.cloud.looker.com"
	LOOKER_AUTH_URI         = "%s/api/3.1/login?client_id=%s&client_secret=%s"
	LOOKER_QUERY_RUNNER_URI = "%s/api/3.1/queries/run/json?force_production=true&cache=true"
	LOOKER_PROD_MODEL       = "network_next_prod"
)
*/

var websiteStatsMutex sync.RWMutex
var websiteStats LiveStats
var statsRefreshInterval time.Duration

var PLATFORM_TYPES = []string{
	"PS4",
	"PS5",
	"XBOX",
	"Switch",
	"Linux",
	"Mac",
	"PC",
}

var CONNECTION_TYPES = []string{
	"Wired",
	"WiFi",
	"Mobile",
}

func main() {
	service := common.CreateService("website_cruncher")

	statsRefreshInterval = envvar.GetDuration("STATS_REFRESH_INTERVAL", time.Minute*5)

	core.Log("stats refresh interval: %s", statsRefreshInterval)

	service.LeaderElection(false)

	StartDataCollection(service)

	service.Router.HandleFunc("/stats", getAllStats())
	service.Router.HandleFunc("/sessions/counts", getLiveSessionCounts())
	service.Router.HandleFunc("/sessions/list", getTopSessionsList())

	service.StartWebServer()

	service.WaitForShutdown()
}

func getTopSessionsList() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		topSessionsList := make([]TopSession, 10)

		// TODO: call out to redis for 10 sessions

		for i := 0; i < 10; i++ {
			topSessionsList[i] = TopSession{}

			numSlices := common.RandomInt(18, 60)

			slices := make([]SessionSlice, numSlices)

			for j := 0; j < numSlices; j++ {
				slices[j] = SessionSlice{
					Timestamp: time.Now().Add(time.Second * time.Duration(-10*(numSlices-i))),
					Next: Stats{
						RTT:        float32(common.RandomInt(0, 30)),
						Jitter:     float32(common.RandomInt(0, 30)),
						PacketLoss: 0,
					},
					Direct: Stats{
						RTT:        float32(common.RandomInt(100, 2000)),
						Jitter:     float32(common.RandomInt(100, 2000)),
						PacketLoss: float32(common.RandomInt(0, 100)),
					},
				}
			}

			topSessionsList[i].Slices = slices

			directRTT := int32(common.RandomInt(100, 2000))
			nextRTT := int32(common.RandomInt(0, 60))

			topSessionsList[i].Meta = SessionMeta{
				ISP:            fmt.Sprintf("%s Communications", common.RandomString(6)),
				Datacenter:     fmt.Sprintf("provider.%s", common.RandomString(6)),
				Platform:       PLATFORM_TYPES[common.RandomInt(0, len(PLATFORM_TYPES)-1)],
				ConnectionType: CONNECTION_TYPES[common.RandomInt(0, len(CONNECTION_TYPES)-1)],
				DirectRTT:      directRTT,
				NextRTT:        nextRTT,
				Improvement:    directRTT - nextRTT,
			}
		}

		sort.Slice(topSessionsList, func(i int, j int) bool {
			return topSessionsList[i].Meta.Improvement > topSessionsList[j].Meta.Improvement
		})

		middleware.CORSControlHandlerFunc(envvar.GetList("ALLOWED_ORIGINS", []string{}), w, r)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(topSessionsList); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func getLiveSessionCounts() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		liveSessionCounts := LiveSessionCounts{
			TotalOnNext:   int32(common.RandomInt(0, 1000)),
			TotalSessions: int32(common.RandomInt(2000, 10000)),
		}

		middleware.CORSControlHandlerFunc(envvar.GetList("ALLOWED_ORIGINS", []string{}), w, r)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(liveSessionCounts); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func getAllStats() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		websiteStatsMutex.RLock()
		stats := websiteStats
		websiteStatsMutex.RUnlock()

		numSecondsPerInterval := statsRefreshInterval.Seconds()

		oldUniquePlayers := float64(stats.UniquePlayers)
		oldBandwidth := float64(stats.AcceleratedBandwidth)
		oldPlaytime := float64(stats.AcceleratedPlayTime)

		deltaUniquePerSecond := float64(stats.UniquePlayersDelta) / numSecondsPerInterval
		deltaBanwidthPerSecond := float64(stats.AcceleratedBandwidthDelta) / numSecondsPerInterval
		deltaPlaytimePerSecond := float64(stats.AcceleratedPlayTimeDelta) / numSecondsPerInterval

		currentSecond := float64(time.Now().UTC().Second())

		newUniquePlayers := oldUniquePlayers + (deltaUniquePerSecond * currentSecond)
		newBanwidth := oldBandwidth + (deltaBanwidthPerSecond * currentSecond)
		newPlaytime := oldPlaytime + (deltaPlaytimePerSecond * currentSecond)

		newStats := LiveStats{
			UniquePlayers:             int32(newUniquePlayers),
			AcceleratedBandwidth:      int32(newBanwidth),
			AcceleratedPlayTime:       int32(newPlaytime),
			UniquePlayersDelta:        stats.UniquePlayersDelta,
			AcceleratedBandwidthDelta: stats.AcceleratedBandwidthDelta,
			AcceleratedPlayTimeDelta:  stats.UniquePlayersDelta,
		}

		middleware.CORSControlHandlerFunc(envvar.GetList("ALLOWED_ORIGINS", []string{}), w, r)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(newStats); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func currentStats() LiveStats {

	websiteStatsMutex.RLock()
	stats := websiteStats
	websiteStatsMutex.RUnlock()

	return stats

}

func CORSControlHandlerFunc(allowedOrigins []string, w http.ResponseWriter, r *http.Request) {
	cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods: []string{
			http.MethodPost,
			http.MethodGet,
			http.MethodOptions,
		},
	}).HandlerFunc(w, r)
}

// -----------------------------------------------------------------------------------------

// todo - setup in service
type LiveStats struct {
	UniquePlayers             int32 `json:"unique_players"`
	AcceleratedPlayTime       int32 `json:"accelerated_play_time"`
	AcceleratedBandwidth      int32 `json:"accelerated_bandwidth"`
	UniquePlayersDelta        int32 `json:"unique_players_delta"`
	AcceleratedPlayTimeDelta  int32 `json:"accelerated_play_time_delta"`
	AcceleratedBandwidthDelta int32 `json:"accelerated_bandwidth_delta"`
}

type LiveSessionCounts struct {
	TotalSessions int32 `json:"total_sessions"`
	TotalOnNext   int32 `json:"total_on_next"`
}

// TODO: Move these somewhere else - don't want to use old routing structs

type Stats struct {
	RTT        float32 `json:"rtt"`
	Jitter     float32 `json:"jitter"`
	PacketLoss float32 `json:"packet_loss"`
}

type Envelope struct {
	Up   int32 `json:"up"`
	Down int32 `json:"down"`
}
type SessionSlice struct {
	Timestamp time.Time `json:"timestamp"`
	Next      Stats     `json:"next"`
	Direct    Stats     `json:"direct"`
	Envelope  Envelope  `json:"envelope"`
}
type SessionMeta struct {
	ISP            string `json:"isp"`
	Datacenter     string `json:"datacenter"`
	Platform       string `json:"platform"`
	ConnectionType string `json:"connection"`
	DirectRTT      int32  `json:"direct_rtt"`
	NextRTT        int32  `json:"next_rtt"`
	Improvement    int32  `json:"improvement"`
}
type SessionPoint struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

type TopSession struct {
	Slices []SessionSlice `json:"slices"`
	Meta   SessionMeta    `json:"meta"`
}

func StartDataCollection(service *common.Service) {

	ticker := time.NewTicker(statsRefreshInterval)

	go func() {

		for {

			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:

				newStats := LiveStats{}

				currentStats := currentStats()

				// todo - grab stats from somewhere (looker, redis, etc)
				newStats.UniquePlayers = int32(common.RandomInt(int(currentStats.UniquePlayers), int(currentStats.UniquePlayers)+1000))
				newStats.UniquePlayersDelta = newStats.UniquePlayers - currentStats.UniquePlayers

				newStats.AcceleratedBandwidth = int32(common.RandomInt(int(currentStats.AcceleratedBandwidth), int(currentStats.AcceleratedBandwidth)+1000))
				newStats.AcceleratedBandwidthDelta = newStats.AcceleratedBandwidth - currentStats.AcceleratedBandwidth

				newStats.AcceleratedPlayTime = int32(common.RandomInt(int(currentStats.AcceleratedPlayTime), int(currentStats.AcceleratedPlayTime)+1000))
				newStats.AcceleratedPlayTimeDelta = newStats.AcceleratedPlayTime - currentStats.AcceleratedPlayTime

				var statsBuffer bytes.Buffer
				encoder := gob.NewEncoder(&statsBuffer)
				if err := encoder.Encode(newStats); err != nil {
					core.Error("failed to encode new stats")
					continue
				}

				newStatsData := statsBuffer.Bytes()

				dataStores := []common.DataStoreConfig{
					{
						Name: "live_stats",
						Data: newStatsData,
					},
				}

				service.UpdateLeaderStore(dataStores)

				dataStores = service.LoadLeaderStore()

				newLiveStats := LiveStats{}

				decoder := gob.NewDecoder(bytes.NewBuffer(dataStores[0].Data))
				err := decoder.Decode(&newLiveStats)
				if err != nil {
					core.Debug("could not decode live stats data: %v", err)
					continue
				}

				websiteStatsMutex.Lock()
				websiteStats = newLiveStats
				websiteStatsMutex.Unlock()

			}
		}
	}()
}

// -----------------------------------------------------------------------------------------

// todo - move into a common module
/*
type LookerClient struct {
	APISettings rtl.ApiSettings
}

type LookerAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int32  `json:"expires_in"`
}

func setupLookerClient() {}

func (l *LookerClient) fetchLookerAuthToken() (string, error) {
	authURL := fmt.Sprintf(LOOKER_AUTH_URI, l.APISettings.BaseUrl, l.APISettings.ClientId, l.APISettings.ClientSecret)
	req, err := http.NewRequest(http.MethodPost, authURL, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	authResponse := LookerAuthResponse{}
	if err = json.Unmarshal(buf.Bytes(), &authResponse); err != nil {
		return "", err
	}

	return authResponse.AccessToken, nil
}

func (l *LookerClient) getWebsiteStats() error {
	// Looker API always passes back an array - "this is the rows for that query - # rows >= 0"
	queryWebsiteStats := make([]LiveStats, 0)

	token, err := l.fetchLookerAuthToken()
	if err != nil {
		return err
	}

	// Fetch Meta data for session

	requiredFields := []string{
		// todo: work with alex for table and field names
	}
	sorts := []string{}
	requiredFilters := make(map[string]interface{})

	query := v4.WriteQuery{
		Model: LOOKER_PROD_MODEL,
		// View:    , todo - work with alex for view name
		Fields:  &requiredFields,
		Filters: &requiredFilters,
		Sorts:   &sorts,
	}

	lookerBody, err := json.Marshal(query)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(LOOKER_QUERY_RUNNER_URI, l.APISettings.BaseUrl), bytes.NewBuffer(lookerBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(buf.Bytes(), &queryWebsiteStats); err != nil {
		return err
	}

	resp.Body.Close()

	if len(queryWebsiteStats) == 0 {
		return fmt.Errorf("failed to look up site data")
	}

	return nil
}
*/
