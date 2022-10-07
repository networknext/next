package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/looker-open-source/sdk-codegen/go/rtl"
	v4 "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/networknext/backend/modules/common"
)

// todo - setup in service
var (
	WebsiteStatsMutex sync.RWMutex
	WebsiteStats      LiveStats
)

const (
	BASE_URL                = "https://networknextexternal.cloud.looker.com"
	LOOKER_AUTH_URI         = "%s/api/3.1/login?client_id=%s&client_secret=%s"
	LOOKER_QUERY_RUNNER_URI = "%s/api/3.1/queries/run/json?force_production=true&cache=true"
	LOOKER_PROD_MODEL       = "network_next_prod"
)

func main() {
	service := common.CreateService("collector")

	/*
		allowedOrigins := envvar.GetList("ALLOWED_ORIGINS", []string{"127.0.0.1:3000", "127.0.0.1:8080", "127.0.0.1:80"})
		timeout := envvar.GetDuration("HTTP_TIMEOUT", time.Second*30)

		core.Log("allowed origins: ")
		for _, origin := range allowedOrigins {
			core.Log("%s", origin)
		}
		core.Log("http timeout: %s", timeout)
	*/

	StartStatCollection()

	service.Router.HandleFunc("/stats", getAllStats).Methods(http.MethodGet)

	service.LeaderElection()

	service.StartWebServer()

	service.WaitForShutdown()
}

func getAllStats(w http.ResponseWriter, r *http.Request) {
	WebsiteStatsMutex.RLock()
	stats := WebsiteStats
	WebsiteStatsMutex.RUnlock()

	json.NewEncoder(w).Encode(stats)
}

// -----------------------------------------------------------------------------------------

// todo - setup in service
type LiveStats struct{}

func StartStatCollection() {
	go func() {

		newStats := LiveStats{}

		// todo - grab stats from somewhere (looker, redis, etc)

		WebsiteStatsMutex.Lock()
		WebsiteStats = newStats
		WebsiteStatsMutex.Unlock()
	}()
}

// -----------------------------------------------------------------------------------------

// todo - move into a common module
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
