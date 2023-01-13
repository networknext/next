package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/looker-open-source/sdk-codegen/go/rtl"
	v4 "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
	"github.com/networknext/backend/modules/core"
)

const (
	BASE_URL    = "https://networknextexternal.cloud.looker.com"
	API_VERSION = "4.0"
	API_TIMEOUT = 300

	LOOKER_AUTH_URI         = "%s/api/%s/login?client_id=%s&client_secret=%s"
	LOOKER_QUERY_RUNNER_URI = "%s/api/%s/queries/run/json?force_production=true&cache=true"
	LOOKER_PROD_MODEL       = "network_next_prod"

	DEFAULT_LOOKER_HOST   = "networknextexternal.cloud.looker.com"
	DEFAULT_CLIENT_ID     = "QXG3cfyWd8xqsVnT7QbT"
	DEFAULT_CLIENT_SECRET = "JT2BpTYNc7fybyHNGs3S24g7"
	DEFAULT_API_SECRET    = "d61764ff20f99e672af3ec7fde75531a790acdb6d58bf46dbe55dac06a6019c0" // TODO - this is tied to andrew@networknext.com user in looker

	WEBSITE_STATS_VIEW = ""
)

type LookerHandlerConfig struct {
	HostURL   string
	ClientID  string
	Secret    string
	APISecret string
}

type LookerHandler struct {
	hostURL     string
	secret      string
	apiSettings rtl.ApiSettings
}

func NewLookerHandler(config LookerHandlerConfig) (*LookerHandler, error) {

	if config.HostURL == "" {
		core.Log("using default looker host")
		config.HostURL = DEFAULT_LOOKER_HOST
	}

	if config.Secret == "" {
		core.Log("using default looker client secret")
		config.Secret = DEFAULT_CLIENT_SECRET
	}

	if config.APISecret == "" {
		core.Log("using default looker api secret")
		config.APISecret = DEFAULT_API_SECRET
	}

	settings := rtl.ApiSettings{
		ClientId:     config.ClientID,
		ClientSecret: config.APISecret,
		ApiVersion:   API_VERSION,
		VerifySsl:    true,
		Timeout:      API_TIMEOUT, // TODO: 5 minute timeout is excesive but is good for now
		BaseUrl:      BASE_URL,
	}

	return &LookerHandler{
		hostURL:     config.HostURL,
		secret:      config.Secret,
		apiSettings: settings,
	}, nil
}

type LookerAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int32  `json:"expires_in"`
}

func (l *LookerHandler) fetchAuthToken() (string, error) {
	authURL := fmt.Sprintf(LOOKER_AUTH_URI, l.apiSettings.BaseUrl, API_VERSION, l.apiSettings.ClientId, l.apiSettings.ClientSecret)
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

type LookerWebsiteStatsQueryResults struct {
	TimestampDate        string `json:"VIEW.start_timestamp_date"`
	AcceleratedBandwidth string `json:"VIEW.accelerated_bandwidth"`
	AcceleratedPlaytime  string `json:"VIEW.accelerated_playtime"`
}

func (l *LookerHandler) RunWebsiteStatsQuery() (LookerWebsiteStatsQueryResults, error) {

	queryResults := LookerWebsiteStatsQueryResults{}

	token, err := l.fetchAuthToken()
	if err != nil {
		return queryResults, err
	}

	// SELECT requiredFields FROM
	requiredFields := []string{}

	sorts := []string{}

	// WHERE requiredFilters == true
	requiredFilters := make(map[string]interface{})

	query := v4.WriteQuery{
		Model:   LOOKER_PROD_MODEL,
		View:    WEBSITE_STATS_VIEW,
		Fields:  &requiredFields,
		Filters: &requiredFilters,
		Sorts:   &sorts,
	}

	respBytes, err := l.runQuery(token, query)
	if err != nil {
		return queryResults, err
	}

	if err = json.Unmarshal(respBytes, &queryResults); err != nil {
		return queryResults, err
	}

	return queryResults, nil
}

func (l *LookerHandler) runQuery(authToken string, query v4.WriteQuery) ([]byte, error) {

	emptyReturn := make([]byte, 0)

	lookerBody, err := json.Marshal(query)
	if err != nil {
		return emptyReturn, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(LOOKER_QUERY_RUNNER_URI, l.apiSettings.BaseUrl, API_VERSION), bytes.NewBuffer(lookerBody))
	if err != nil {
		return emptyReturn, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return emptyReturn, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return emptyReturn, err
	}

	return buf.Bytes(), nil

}
