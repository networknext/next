package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	v4 "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/networknext/backend/modules/common"
)

// todo - setup in service
var (
	WebsiteStatsMutex sync.RWMutex
	WebsiteStats      LiveStats
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
type LookerClient struct{}

type LookerAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int32  `json:"expires_in"`
}

func setupLookerClient() {}

func fetchLookerAuthToken() (string, error) {
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

func getWebsiteStats() {
	// Looker API always passes back an array - "this is the rows for that query - # rows >= 0"
	queryWebsiteStats := make([]LiveStats, 0)

	token, err := l.FetchAuthToken()
	if err != nil {
		return LookerSessionTimestampLookup{}, err
	}

	// Fetch Meta data for session

	requiredFields := []string{
		LOOKER_SESSION_SUMMARY_VIEW + ".start_timestamp_time",
		LOOKER_SESSION_SUMMARY_VIEW + ".start_timestamp_date",
		LOOKER_SESSION_SUMMARY_VIEW + ".session_id",
		LOOKER_SESSION_SUMMARY_VIEW + ".buyer_id",
	}
	sorts := []string{}
	requiredFilters := make(map[string]interface{})

	requiredFilters[LOOKER_SESSION_SUMMARY_VIEW+".session_id"] = fmt.Sprintf("%d", int64(uintID64))
	requiredFilters[LOOKER_SESSION_SUMMARY_VIEW+".start_timestamp_date"] = queryTimeFrame

	query := v4.WriteQuery{
		Model:   LOOKER_PROD_MODEL,
		View:    LOOKER_SESSION_LOOK_UP_VIEW,
		Fields:  &requiredFields,
		Filters: &requiredFilters,
		Sorts:   &sorts,
	}

	lookerBody, err := json.Marshal(query)
	if err != nil {
		return LookerSessionTimestampLookup{}, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(LOOKER_QUERY_RUNNER_URI, l.APISettings.BaseUrl), bytes.NewBuffer(lookerBody))
	if err != nil {
		return LookerSessionTimestampLookup{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return LookerSessionTimestampLookup{}, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return LookerSessionTimestampLookup{}, err
	}

	if err = json.Unmarshal(buf.Bytes(), &querySessionLookup); err != nil {
		return LookerSessionTimestampLookup{}, err
	}

	resp.Body.Close()

	if len(querySessionLookup) == 0 {
		return LookerSessionTimestampLookup{}, fmt.Errorf("failed to look up session meta data")
	}

	return querySessionLookup[0], nil
}
