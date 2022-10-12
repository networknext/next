package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
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
var redisHostname string
var redisPassword string
var redisSelector *common.RedisSelector
var redisSelectorTimeout time.Duration

func main() {
	service := common.CreateService("website cruncher")

	statsRefreshInterval = envvar.GetDuration("STATS_REFRESH_INTERVAL", time.Hour*24)
	redisHostname = envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")
	redisPassword = envvar.GetString("REDIS_PASSWORD", "")
	redisSelectorTimeout = envvar.GetDuration("REDIS_SELECTOR_TIMEOUT", time.Second*10)

	core.Log("stats refresh interval: %s", statsRefreshInterval)
	core.Log("redis hostname: %s", redisHostname)
	core.Log("redis password: %s", redisPassword)
	core.Log("redis selector timeout: %s", redisSelectorTimeout)

	StartStatCollection(service)

	service.Router.HandleFunc("/stats", getAllStats).Methods(http.MethodGet)

	service.StartWebServer()

	service.WaitForShutdown()
}

func getAllStats(w http.ResponseWriter, r *http.Request) {

	websiteStatsMutex.RLock()
	stats := websiteStats
	websiteStatsMutex.RUnlock()

	numMinutesPerInterval := (int64(statsRefreshInterval.Seconds()) / 60)

	oldUniquePlayers := stats.UniquePlayers
	oldBandwidth := stats.AcceleratedBandwidth
	oldPlaytime := stats.AcceleratedPlayTime

	deltaUniquePerMinute := stats.UniquePlayersDelta / numMinutesPerInterval
	deltaBanwidthPerMinute := stats.AcceleratedBandwidthDelta / numMinutesPerInterval
	deltaPlaytimePerMinute := stats.AcceleratedPlayTimeDelta / numMinutesPerInterval

	currentMinute := time.Now().UTC().Minute()

	newUniquePlayers := oldUniquePlayers + (deltaUniquePerMinute * int64(currentMinute))
	newBanwidth := oldBandwidth + (deltaBanwidthPerMinute * int64(currentMinute))
	newPlaytime := oldPlaytime + (deltaPlaytimePerMinute * int64(currentMinute))

	newStats := LiveStats{
		UniquePlayers:             newUniquePlayers,
		AcceleratedBandwidth:      newBanwidth,
		AcceleratedPlayTime:       newPlaytime,
		UniquePlayersDelta:        stats.UniquePlayersDelta,
		AcceleratedBandwidthDelta: stats.AcceleratedBandwidthDelta,
		AcceleratedPlayTimeDelta:  stats.UniquePlayersDelta,
	}

	json.NewEncoder(w).Encode(newStats)
}

func currentStats() LiveStats {

	websiteStatsMutex.RLock()
	stats := websiteStats
	websiteStatsMutex.RUnlock()

	return stats
}

// -----------------------------------------------------------------------------------------

// todo - setup in service
type LiveStats struct {
	UniquePlayers             int64 `json:"unique_players"`
	AcceleratedPlayTime       int64 `json:"accelerated_play_time"`
	AcceleratedBandwidth      int64 `json:"accelerated_bandwidth"`
	UniquePlayersDelta        int64 `json:"unique_players_delta"`
	AcceleratedPlayTimeDelta  int64 `json:"accelerated_play_time_delta"`
	AcceleratedBandwidthDelta int64 `json:"accelerated_bandwidth_delta"`
}

func StartStatCollection(service *common.Service) {

	var err error

	ticker := time.NewTicker(statsRefreshInterval)

	config := common.RedisSelectorConfig{}

	config.RedisHostname = redisHostname
	config.RedisPassword = redisPassword
	config.Timeout = redisSelectorTimeout

	redisSelector, err = common.CreateRedisSelector(service.Context, config)
	if err != nil {
		core.Error("failed to create redis selector: %v", err)
		os.Exit(1)
	}

	go func() {

		for {

			select {
			case <-service.Context.Done():
				return
			case <-ticker.C:

				newStats := LiveStats{}

				currentStats := currentStats()

				// todo - grab stats from somewhere (looker, redis, etc)
				newStats.UniquePlayers = int64(common.RandomInt(int(currentStats.UniquePlayers), int(currentStats.UniquePlayers)+1000))
				newStats.UniquePlayersDelta = newStats.UniquePlayers - currentStats.UniquePlayers

				newStats.AcceleratedBandwidth = int64(common.RandomInt(int(currentStats.AcceleratedBandwidth), int(currentStats.AcceleratedBandwidth)+1000))
				newStats.AcceleratedBandwidthDelta = newStats.AcceleratedBandwidth - currentStats.AcceleratedBandwidth

				newStats.AcceleratedPlayTime = int64(common.RandomInt(int(currentStats.AcceleratedPlayTime), int(currentStats.AcceleratedPlayTime)+1000))
				newStats.AcceleratedPlayTimeDelta = newStats.AcceleratedPlayTime - currentStats.AcceleratedPlayTime

				var buffer bytes.Buffer
				encoder := gob.NewEncoder(&buffer)
				if err := encoder.Encode(newStats); err != nil {
					core.Error("failed to encode new stats")
					continue
				}

				newStatsData := buffer.Bytes()

				dataStores := []common.DataStoreConfig{
					{
						Name: "live_stats",
						Data: newStatsData,
					},
				}

				redisSelector.Store(service.Context, dataStores)

				dataStores = redisSelector.Load(service.Context)

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
