package looker

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/looker-open-source/sdk-codegen/go/rtl"
	v3 "github.com/looker-open-source/sdk-codegen/go/sdk/v3"
	v4 "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

const (
	LOOKER_SESSION_TIMEOUT          = 86400
	EMBEDDED_USER_GROUP_ID          = 3
	USAGE_DASH_URL                  = "/embed/dashboards-next/11"
	BASE_URL                        = "https://networknextexternal.cloud.looker.com"
	API_VERSION                     = "4.0"
	API_TIMEOUT                     = 300
	LOOKER_SAVES_ROW_LIMIT          = "10"
	LOOKER_AUTH_URI                 = "%s/api/3.1/login?client_id=%s&client_secret=%s"
	LOOKER_QUERY_RUNNER_URI         = "%s/api/3.1/queries/run/json?force_production=true&cache=true"
	LOOKER_PROD_MODEL               = "network_next_prod"
	LOOKER_SAVES_VIEW               = "daily_big_saves"
	LOOKER_BILLING2_VIEW            = "billing2"
	LOOKER_SESSION_SUMMARY_VIEW     = "billing2_session_summary"
	LOOKER_SESSION_LOOK_UP_VIEW     = "billing2_session_summary_lookup_only"
	LOOKER_DATACENTER_INFO_VIEW     = "datacenter_info_v3"
	LOOKER_RELAY_INFO_VIEW          = "relay_info_v3"
	LOOKER_NEAR_RELAY_OFFSET_FILTER = "if(is_null(${relay_info_v3.relay_name}) OR length(${relay_info_v3.relay_name}) = 0,yes,${billing2_session_summary__near_relay_ids.offset}=${billing2_session_summary__near_relay_jitters.offset} AND ${billing2_session_summary__near_relay_ids.offset} = ${billing2_session_summary__near_relay_rtts.offset} AND ${billing2_session_summary__near_relay_ids.offset} = ${billing2_session_summary__near_relay_packet_losses.offset})"
)

type LookerWebhookAttachment struct {
	Data []LookerSave `json:"data"` // TODO: Potentially break this out into a bunch of different webhook attachment types
}

type LookerWebhookScheduledPlan struct {
	Title          string `json:"title"`
	URL            string `json:"url"`
	SchedulePlanID int    `json:"schedule_plan_id"`
	Type           string `json:"type"`
}

type LookerWebhookPayload struct {
	Attachment    LookerWebhookAttachment    `json:"attachment"`
	ScheduledPlan LookerWebhookScheduledPlan `json:"scheduled_plan"`
}

type LookerClient struct {
	HostURL     string
	Secret      string
	APISettings rtl.ApiSettings
}

func NewLookerClient(hostURL string, secret string, clientID string, apiSecret string) (*LookerClient, error) {
	if hostURL == "" || secret == "" {
		return nil, fmt.Errorf("Host and Secret are required")
	}

	// TODO: Pull these in from environment in portal.go and swap creds for a dedicated API user
	settings := rtl.ApiSettings{
		ClientId:     clientID,
		ClientSecret: apiSecret,
		ApiVersion:   API_VERSION,
		VerifySsl:    true,
		Timeout:      API_TIMEOUT, // TODO: 5 minute timeout is excesive but is good for now
		BaseUrl:      BASE_URL,
	}

	return &LookerClient{
		HostURL:     hostURL,
		Secret:      secret,
		APISettings: settings,
	}, nil
}

type LookerAuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int32  `json:"expires_in"`
}

func (l *LookerClient) FetchAuthToken() (string, error) {
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

type LookerSessionTimestampLookup struct {
	TimestampDate string `json:"billing2_session_summary.start_timestamp_date"`
	TimestampTime string `json:"billing2_session_summary.start_timestamp_time"`
	SessionID     int64  `json:"billing2_session_summary.session_id"`
	BuyerID       int64  `json:"billing2_session_summary.buyer_id"`
}

type LookerUserSessionMeta struct {
	Timestamp       string `json:"billing2_session_summary.start_timestamp_time"`
	SessionID       int64  `json:"billing2_session_summary.session_id"`
	BuyerID         int64  `json:"billing2_session_summary.buyer_id"`
	Platform        int8   `json:"billing2_session_summary.platform_type"`
	Connection      int8   `json:"billing2_session_summary.connection_type"`
	ISP             string `json:"billing2_session_summary.isp"`
	ServerAddress   string `json:"billing2_session_summary.server_address"`
	DatacenterName  string `json:"datacenter_info_v3.datacenter_name"`
	DatacenterAlias string `json:"datacenter_info_v3.alias"`
}
type LookerSessionMeta struct {
	Timestamp             string      `json:"billing2_session_summary.start_timestamp_date"`
	SessionID             int64       `json:"billing2_session_summary.session_id"`
	BuyerID               int64       `json:"billing2_session_summary.buyer_id"`
	UserHash              int64       `json:"billing2_session_summary.user_hash"`
	EverOnNext            string      `json:"billing2_session_summary.ever_on_next"`
	SDKVersion            string      `json:"billing2_session_summary.sdk_version"`
	ClientAddress         string      `json:"billing2_session_summary.client_address"`
	ServerAddress         string      `json:"billing2_session_summary.server_address"`
	Longitude             float64     `json:"billing2_session_summary.longitude"`
	Latitude              float64     `json:"billing2_session_summary.latitude"`
	ISP                   string      `json:"billing2_session_summary.isp"`
	Platform              int8        `json:"billing2_session_summary.platform_type"`
	Connection            int8        `json:"billing2_session_summary.connection_type"`
	DatacenterName        string      `json:"datacenter_info_v3.datacenter_name"`
	DatacenterAlias       string      `json:"datacenter_info_v3.alias"`
	NearRelayNames        string      `json:"relay_info_v3.relay_name"`
	NearRelayIDs          json.Number `json:"billing2_session_summary__near_relay_ids.billing2_session_summary__near_relay_ids"`
	NearRelayRTTs         json.Number `json:"billing2_session_summary__near_relay_rtts.billing2_session_summary__near_relay_rtts"`
	NearRelayJitters      json.Number `json:"billing2_session_summary__near_relay_jitters.billing2_session_summary__near_relay_jitters"`
	NearRelayPacketLosses json.Number `json:"billing2_session_summary__near_relay_packet_losses.billing2_session_summary__near_relay_packet_losses"`
}

type LookerSessionSlice struct {
	Timestamp         string            `json:"billing2.timestamp_time"`
	SessionID         int64             `json:"billing2.session_id"`
	SliceNumber       int               `json:"billing2.slice_number"`
	NextRTT           float64           `json:"billing2.next_rtt"`
	NextJitter        float64           `json:"billing2.next_jitter"`
	NextPacketLoss    float64           `json:"billing2.next_packet_loss"`
	DirectRTT         float64           `json:"billing2.direct_rtt"`
	DirectJitter      float64           `json:"billing2.direct_jitter"`
	DirectPacketLoss  float64           `json:"billing2.direct_packet_loss"`
	PredictedRTT      float64           `json:"billing2.predicted_next_rtt"`
	RouteDiversity    int32             `json:"billing2.route_diversity"`
	EnvelopeUp        int64             `json:"billing2.next_bytes_up"`
	EnvelopeDown      int64             `json:"billing2.next_bytes_down"`
	OnNetworkNext     string            `json:"billing2.next"`
	IsMultiPath       string            `json:"billing2.multipath"`
	IsTryBeforeYouBuy string            `json:"billing2.is_try_before_you_buy"`
	NextRelays        []LookerNextRelay `json:"next_relays"`
	NextRelayName     string            `json:"relay_info_v3.relay_name,omitempty"`
	NextRelayOffset   int               `json:"billing2__next_relays.offset"`
}

type LookerNearRelay struct {
	ID     int64
	Name   string
	RTT    float64
	Jitter float64
	PL     float64
}

type LookerNextRelay struct {
	Offset int
	Name   string
}

type LookerSession struct {
	Meta       LookerSessionMeta
	NearRelays []LookerNearRelay
	Slices     []LookerSessionSlice
}

func (l *LookerClient) RunSessionTimestampLookupQuery(sessionID string, timeFrame string) (LookerSessionTimestampLookup, error) {
	// Looker API always passes back an array - "this is the rows for that query - # rows >= 0"
	querySessionLookup := make([]LookerSessionTimestampLookup, 0)

	uintID64, err := strconv.ParseUint(sessionID, 16, 64)
	if err != nil {
		return LookerSessionTimestampLookup{}, err
	}

	queryTimeFrame := timeFrame

	if queryTimeFrame == "" {
		queryTimeFrame = "7 days"
	}

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

func (l *LookerClient) RunSessionMetaDataQuery(sessionID int64, date string, customerCode string, analysisOnly bool) ([]LookerSessionMeta, error) {
	queryMeta := make([]LookerSessionMeta, 0)

	if date == "" {
		return queryMeta, fmt.Errorf("a specific date is required for session slice lookup")
	}

	token, err := l.FetchAuthToken()
	if err != nil {
		return queryMeta, err
	}

	requiredFields := []string{
		LOOKER_SESSION_SUMMARY_VIEW + ".start_timestamp_time",
		LOOKER_SESSION_SUMMARY_VIEW + ".session_id",
		LOOKER_SESSION_SUMMARY_VIEW + ".user_hash",
		LOOKER_SESSION_SUMMARY_VIEW + ".platform_type",
		LOOKER_SESSION_SUMMARY_VIEW + ".connection_type",
		LOOKER_SESSION_SUMMARY_VIEW + ".isp",
		LOOKER_SESSION_SUMMARY_VIEW + ".latitude",
		LOOKER_SESSION_SUMMARY_VIEW + ".longitude",
		LOOKER_SESSION_SUMMARY_VIEW + ".sdk_version",
		LOOKER_SESSION_SUMMARY_VIEW + ".client_address",
		LOOKER_SESSION_SUMMARY_VIEW + ".server_address",
		LOOKER_SESSION_SUMMARY_VIEW + ".ever_on_next",
		LOOKER_DATACENTER_INFO_VIEW + ".datacenter_name",
		LOOKER_DATACENTER_INFO_VIEW + ".alias",
		LOOKER_RELAY_INFO_VIEW + ".relay_name",
		"billing2_session_summary__near_relay_ids.billing2_session_summary__near_relay_ids",
		"billing2_session_summary__near_relay_rtts.billing2_session_summary__near_relay_rtts",
		"billing2_session_summary__near_relay_jitters.billing2_session_summary__near_relay_jitters",
		"billing2_session_summary__near_relay_packet_losses.billing2_session_summary__near_relay_packet_losses",
	}

	sorts := []string{}
	requiredFilters := make(map[string]interface{})

	requiredFilters[LOOKER_SESSION_SUMMARY_VIEW+".session_id"] = fmt.Sprintf("%d", sessionID)
	requiredFilters[LOOKER_SESSION_SUMMARY_VIEW+".start_timestamp_date"] = date

	filterExpression := ""
	if !analysisOnly {
		filterExpression = fmt.Sprintf(LOOKER_NEAR_RELAY_OFFSET_FILTER)
	}

	if customerCode != "" {
		if filterExpression != "" {
			filterExpression = "(" + filterExpression + ") AND "
		}

		filterExpression = fmt.Sprintf(`%s${buyer_info_v2.customer_code} = "%s"`, filterExpression, customerCode)
	}

	query := v4.WriteQuery{
		Model:            LOOKER_PROD_MODEL,
		View:             LOOKER_SESSION_SUMMARY_VIEW,
		Fields:           &requiredFields,
		Filters:          &requiredFilters,
		FilterExpression: &filterExpression,
		Sorts:            &sorts,
	}

	lookerBody, err := json.Marshal(query)
	if err != nil {
		return queryMeta, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(LOOKER_QUERY_RUNNER_URI, l.APISettings.BaseUrl), bytes.NewBuffer(lookerBody))
	if err != nil {
		return queryMeta, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return queryMeta, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return queryMeta, err
	}

	if err = json.Unmarshal(buf.Bytes(), &queryMeta); err != nil {
		return queryMeta, err
	}

	if len(queryMeta) == 0 {
		return queryMeta, fmt.Errorf("failed to fetch near relays and meta data")
	}

	return queryMeta, nil
}

func (l *LookerClient) RunSessionSliceLookupQuery(sessionID int64, date string, customerCode string) ([]LookerSessionSlice, error) {
	querySessionSlices := make([]LookerSessionSlice, 0)

	if date == "" {
		return querySessionSlices, fmt.Errorf("a specific date is required for session slice lookup")
	}

	token, err := l.FetchAuthToken()
	if err != nil {
		return querySessionSlices, err
	}

	requiredFields := []string{
		LOOKER_BILLING2_VIEW + ".timestamp_time",
		LOOKER_BILLING2_VIEW + ".slice_number",
		LOOKER_BILLING2_VIEW + ".next_rtt",
		LOOKER_BILLING2_VIEW + ".next_jitter",
		LOOKER_BILLING2_VIEW + ".next_packet_loss",
		LOOKER_BILLING2_VIEW + ".predicted_next_rtt",
		LOOKER_BILLING2_VIEW + ".direct_rtt",
		LOOKER_BILLING2_VIEW + ".direct_jitter",
		LOOKER_BILLING2_VIEW + ".direct_packet_loss",
		LOOKER_BILLING2_VIEW + ".next_bytes_up",
		LOOKER_BILLING2_VIEW + ".next_bytes_down",
		LOOKER_BILLING2_VIEW + ".route_diversity",
		LOOKER_BILLING2_VIEW + ".next",
		LOOKER_BILLING2_VIEW + ".multipath",
		LOOKER_RELAY_INFO_VIEW + ".relay_name",
		"billing2__next_relays.offset",
	}

	sorts := []string{LOOKER_BILLING2_VIEW + ".timestamp_time"}
	requiredFilters := make(map[string]interface{})

	requiredFilters[LOOKER_BILLING2_VIEW+".session_id"] = fmt.Sprintf("%d", sessionID)
	requiredFilters[LOOKER_BILLING2_VIEW+".timestamp_time"] = date

	query := v4.WriteQuery{
		Model:   LOOKER_PROD_MODEL,
		View:    LOOKER_BILLING2_VIEW,
		Fields:  &requiredFields,
		Filters: &requiredFilters,
		Sorts:   &sorts,
	}

	lookerBody, err := json.Marshal(query)
	if err != nil {
		return querySessionSlices, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(LOOKER_QUERY_RUNNER_URI, l.APISettings.BaseUrl), bytes.NewBuffer(lookerBody))
	if err != nil {
		return querySessionSlices, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return querySessionSlices, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return querySessionSlices, err
	}

	if err = json.Unmarshal(buf.Bytes(), &querySessionSlices); err != nil {
		return querySessionSlices, err
	}

	return querySessionSlices, nil
}

func (l *LookerClient) RunUserSessionsLookupQuery(userID string, userIDHex string, userIDHash string, timeFrame string, customerCode string) ([]LookerUserSessionMeta, error) { // Timeframes 7, 10, 30, 60, 90
	querySessions := make([]LookerUserSessionMeta, 0)

	// Auth Looker API connection
	token, err := l.FetchAuthToken()
	if err != nil {
		return querySessions, err
	}

	// Set up required fields (columns from Looker table)
	requiredFields := []string{
		LOOKER_SESSION_SUMMARY_VIEW + ".start_timestamp_time",
		LOOKER_SESSION_SUMMARY_VIEW + ".session_id",
		LOOKER_SESSION_SUMMARY_VIEW + ".buyer_id",
		LOOKER_SESSION_SUMMARY_VIEW + ".user_hash",
		LOOKER_SESSION_SUMMARY_VIEW + ".platform_type",
		LOOKER_SESSION_SUMMARY_VIEW + ".connection_type",
		LOOKER_SESSION_SUMMARY_VIEW + ".isp",
		LOOKER_DATACENTER_INFO_VIEW + ".datacenter_name",
		LOOKER_SESSION_SUMMARY_VIEW + ".server_address",
		LOOKER_DATACENTER_INFO_VIEW + ".alias",
	}
	sorts := []string{}
	requiredFilters := make(map[string]interface{})

	// Add the timeframe to optimize query
	requiredFilters[LOOKER_SESSION_SUMMARY_VIEW+".start_timestamp_date"] = timeFrame

	filterExpression := ""

	// Build User ID filter - User could pass in something that hasn't been hashed yet or isn't in hex so we have to handle those cases
	if userID != "" {
		uintID64, err := strconv.ParseUint(userID, 16, 64)
		if err == nil {
			filterExpression = fmt.Sprintf("${%s.user_hash} = %d", LOOKER_SESSION_SUMMARY_VIEW, int64(uintID64))
		}
	}

	if userIDHex != "" {
		uintID64, err := strconv.ParseUint(userIDHex, 16, 64)
		if err == nil {
			if filterExpression != "" {
				filterExpression = filterExpression + " OR "
			}
			filterExpression = filterExpression + fmt.Sprintf("${%s.user_hash} = %d", LOOKER_SESSION_SUMMARY_VIEW, int64(uintID64))
		}
	}

	if userIDHash != "" {
		uintID64, err := strconv.ParseUint(userIDHash, 16, 64)
		if err == nil {
			if filterExpression != "" {
				filterExpression = filterExpression + " OR "
			}
			filterExpression = filterExpression + fmt.Sprintf("${%s.user_hash} = %d", LOOKER_SESSION_SUMMARY_VIEW, int64(uintID64))
		}
	}

	if customerCode != "" {
		filterExpression = "(" + filterExpression + ")"
		filterExpression = filterExpression + " AND " + fmt.Sprintf(`${buyer_info_v2.customer_code} = "%s"`, customerCode)
	}

	// If none of the passed in user IDs work, return no sessions
	if filterExpression == "" {
		return querySessions, nil
	}

	query := v4.WriteQuery{
		Model:            LOOKER_PROD_MODEL,
		View:             LOOKER_SESSION_SUMMARY_VIEW,
		Fields:           &requiredFields,
		Filters:          &requiredFilters,
		FilterExpression: &filterExpression,
		Sorts:            &sorts,
	}

	lookerBody, err := json.Marshal(query)
	if err != nil {
		return querySessions, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(LOOKER_QUERY_RUNNER_URI, l.APISettings.BaseUrl), bytes.NewBuffer(lookerBody))
	if err != nil {
		return querySessions, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return querySessions, err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return querySessions, err
	}

	if err = json.Unmarshal(buf.Bytes(), &querySessions); err != nil {
		return querySessions, err
	}

	return querySessions, nil
}

// Custom Queries ======================================================================================

func (l *LookerClient) RemoveLookerUserByAuth0ID(auth0ID string) error {
	authSession := rtl.NewAuthSession(l.APISettings)

	sdk := v3.NewLookerSDK(authSession)

	users, err := sdk.AllUsers(v3.RequestAllUsers{}, nil)
	if err != nil {
		return err
	}

	foundUserID := int64(-1)
	for _, user := range users {
		embedCreds := *user.CredentialsEmbed

		if len(embedCreds) > 0 {
			if *embedCreds[0].ExternalUserId == auth0ID && user.Email == nil {
				foundUserID = *user.Id
			}
		}
	}

	if foundUserID >= 0 {
		_, err := sdk.DeleteUser(foundUserID, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

type LookerSave struct {
	SessionID               int64   `json:"daily_big_saves.session_id"`
	SaveScore               float64 `json:"daily_big_saves.save_score"`
	AverageDirectRTT        float64 `json:"daily_big_saves.avg_directrtt"`
	AverageNextRTT          float64 `json:"daily_big_saves.avg_nextrtt"`
	AverageDirectPacketLoss float64 `json:"daily_big_saves.avg_directpacketloss"`
	AverageNextPacketLoss   float64 `json:"daily_big_saves.avg_nextpacketloss"`
	Duration                float64 `json:"daily_big_saves.duration"`
}

func (l *LookerClient) RunSavesQuery(customerCode string) ([]LookerSave, error) {
	saves := []LookerSave{}

	token, err := l.FetchAuthToken()
	if err != nil {
		return saves, err
	}

	requiredFields := []string{
		LOOKER_SAVES_VIEW + ".date_date",
		LOOKER_SAVES_VIEW + ".session_id",
		LOOKER_SAVES_VIEW + ".save_score",
		LOOKER_SAVES_VIEW + ".avg_directrtt",
		LOOKER_SAVES_VIEW + ".avg_nextrtt",
		LOOKER_SAVES_VIEW + ".avg_directpacketloss",
		LOOKER_SAVES_VIEW + ".avg_nextpacketloss",
		LOOKER_SAVES_VIEW + ".duration",
	}
	sorts := []string{LOOKER_SAVES_VIEW + ".save_score desc 0"}
	requiredFilters := make(map[string]interface{})
	rowLimit := LOOKER_SAVES_ROW_LIMIT

	requiredFilters["buyer_info_v2.customer_code"] = customerCode
	requiredFilters[LOOKER_SAVES_VIEW+".date_date"] = "7 days"

	query := v4.WriteQuery{
		Model:   LOOKER_PROD_MODEL,
		View:    LOOKER_SAVES_VIEW,
		Fields:  &requiredFields,
		Filters: &requiredFilters,
		Limit:   &rowLimit,
		Sorts:   &sorts,
	}

	lookerBody, _ := json.Marshal(query)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(LOOKER_QUERY_RUNNER_URI, l.APISettings.BaseUrl), bytes.NewBuffer(lookerBody))
	if err != nil {
		return saves, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return saves, err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return saves, err
	}

	if err = json.Unmarshal(buf.Bytes(), &saves); err != nil {
		return saves, err
	}

	return saves, nil
}

// Looker Dashboards ======================================================================================

type AnalyticsDashboardCategory struct {
	ID               int64  `json:"id"`
	Order            int32  `json:"order"`
	Label            string `json:"label"`
	ParentCategoryID int64  `json:"parent_category_id"`
}

type AnalyticsDashboard struct {
	ID           int64                      `json:"id"`
	Order        int32                      `json:"order"`
	Name         string                     `json:"name"`
	Premium      bool                       `json:"premium"`
	Admin        bool                       `json:"admin"`
	LookerID     int64                      `json:"looker_id"`
	CustomerCode string                     `json:"customer_code"`
	Category     AnalyticsDashboardCategory `json:"category"`
}

type LookerURLOptions struct {
	Secret          string                            //required
	Host            string                            //required
	EmbedURL        string                            //required
	Nonce           string                            //required
	Time            int64                             //required
	SessionLength   int                               //required
	ExternalUserId  string                            //required
	Permissions     []string                          //required
	Models          []string                          //required
	ForceLogout     bool                              //required
	GroupsIds       []int                             //optional
	ExternalGroupId string                            //optional
	UserAttributes  map[string]interface{}            //optional
	AccessFilters   map[string]map[string]interface{} //required
	FirstName       string                            //optional
	LastName        string                            //optional
}

// TODO: this into a cache to be speed up frontend calls -> ID to name mapping is helpful
type LookerDashboard struct {
	ID    int32  `json:"id"`
	Title string `json:"title"`
}

func (l *LookerClient) FetchCurrentLookerDashboards() ([]LookerDashboard, error) {
	dashboardList := make([]LookerDashboard, 0)

	lookerSession := rtl.NewAuthSession(l.APISettings)

	// New instance of LookerSDK
	sdk := v4.NewLookerSDK(lookerSession)

	// List all Dashboards in Looker
	dashboards, err := sdk.SearchDashboards(v4.RequestSearchDashboards{}, nil)
	if err != nil {
		return dashboardList, err
	}
	for _, d := range dashboards {
		intID, err := strconv.Atoi(*d.Id)
		if err != nil {
			continue
		}
		dashboardList = append(dashboardList, LookerDashboard{
			ID:    int32(intID),
			Title: *d.Title,
		})
	}

	return dashboardList, nil
}

func (l *LookerClient) BuildGeneralPortalLookerURLWithDashID(id string, customerCode string, requestID, origin string) (string, error) {
	nonce, err := GenerateRandomString(16)
	if err != nil {
		return "", err
	}

	embedURL := ""
	if origin != "" {
		embedURL = fmt.Sprintf("/embed/dashboards-next/%s?embed_domain=%s", id, origin)
	} else {
		embedURL = fmt.Sprintf("/embed/dashboards-next/%s", id)
	}

	urlOptions := LookerURLOptions{
		Host:            l.HostURL,
		Secret:          l.Secret,
		ExternalUserId:  fmt.Sprintf("\"%s\"", requestID),
		GroupsIds:       []int{EMBEDDED_USER_GROUP_ID},
		ExternalGroupId: "",
		Permissions:     []string{"access_data", "see_looks", "see_user_dashboards", "download_without_limit", "clear_cache_refresh"}, // TODO: This may or may not need to change
		Models:          []string{"networknext_prod"},                                                                                 // TODO: This may or may not need to change
		AccessFilters:   make(map[string]map[string]interface{}),
		UserAttributes:  make(map[string]interface{}),
		SessionLength:   LOOKER_SESSION_TIMEOUT,
		EmbedURL:        "/login/embed/" + url.QueryEscape(embedURL),
		ForceLogout:     true,
		Nonce:           fmt.Sprintf("\"%s\"", nonce),
		Time:            time.Now().Unix(),
	}

	urlOptions.UserAttributes["customer_code"] = customerCode

	return BuildLookerURL(urlOptions), nil
}

func BuildLookerURL(urlOptions LookerURLOptions) string {
	// TODO: Verify logic below, this came from here: https://github.com/looker/looker_embed_sso_examples/pull/36 and is NOT an official implementation. That being said, be careful changing it because it works :P
	jsonPerms, _ := json.Marshal(urlOptions.Permissions)
	jsonModels, _ := json.Marshal(urlOptions.Models)
	jsonUserAttrs, _ := json.Marshal(urlOptions.UserAttributes)
	jsonFilters, _ := json.Marshal(urlOptions.AccessFilters)
	jsonGroupIds, _ := json.Marshal(urlOptions.GroupsIds)
	strTime := strconv.Itoa(int(urlOptions.Time))
	strSessionLen := strconv.Itoa(urlOptions.SessionLength)
	strForceLogin := strconv.FormatBool(urlOptions.ForceLogout)

	strToSign := strings.Join([]string{urlOptions.Host,
		urlOptions.EmbedURL,
		urlOptions.Nonce,
		strTime,
		strSessionLen,
		urlOptions.ExternalUserId,
		string(jsonPerms),
		string(jsonModels)}, "\n")

	strToSign = strToSign + "\n"

	if len(urlOptions.GroupsIds) > 0 {
		strToSign = strToSign + string(jsonGroupIds) + "\n"
	}

	if urlOptions.ExternalGroupId != "" {
		strToSign = strToSign + urlOptions.ExternalGroupId + "\n"
	}

	if len(urlOptions.UserAttributes) > 0 {
		strToSign = strToSign + string(jsonUserAttrs) + "\n"
	}

	strToSign = strToSign + string(jsonFilters)

	h := hmac.New(sha1.New, []byte(urlOptions.Secret))
	h.Write([]byte(strToSign))
	encoded := base64.StdEncoding.EncodeToString(h.Sum(nil))

	query := url.Values{}
	query.Add("nonce", urlOptions.Nonce)
	query.Add("time", strTime)
	query.Add("session_length", strSessionLen)
	query.Add("external_user_id", urlOptions.ExternalUserId)
	query.Add("permissions", string(jsonPerms))
	query.Add("models", string(jsonModels))
	query.Add("access_filters", string(jsonFilters))
	query.Add("first_name", urlOptions.FirstName)
	query.Add("last_name", urlOptions.LastName)
	query.Add("force_logout_login", strForceLogin)
	query.Add("signature", encoded)

	if len(urlOptions.GroupsIds) > 0 {
		query.Add("group_ids", string(jsonGroupIds))
	}

	if urlOptions.ExternalGroupId != "" {
		query.Add("external_group_id", urlOptions.ExternalGroupId)
	}

	if len(urlOptions.UserAttributes) > 0 {
		query.Add("user_attributes", string(jsonUserAttrs))
	}

	finalUrl := fmt.Sprintf("https://%s%s?%s", urlOptions.Host, urlOptions.EmbedURL, query.Encode())

	return finalUrl
}

func (l *LookerClient) GenerateUsageDashboardURL(customerCode string, requestID string, origin string, dateString string) (string, error) {
	nonce, err := GenerateRandomString(16)
	if err != nil {
		return "", err
	}

	dashURL := fmt.Sprintf("%s?embed_domain=%s", USAGE_DASH_URL, origin)
	if dateString != "" {
		dashURL = fmt.Sprintf("%s&Billing+Period=%s", dashURL, dateString)
	}

	urlOptions := LookerURLOptions{
		Host:            l.HostURL,
		Secret:          l.Secret,
		ExternalUserId:  fmt.Sprintf("\"%s\"", requestID),
		GroupsIds:       []int{EMBEDDED_USER_GROUP_ID},
		ExternalGroupId: "",
		Permissions:     []string{"access_data", "see_looks", "see_user_dashboards", "download_without_limit", "clear_cache_refresh"}, // TODO: This may or may not need to change
		Models:          []string{"networknext_prod"},                                                                                 // TODO: This may or may not need to change
		AccessFilters:   make(map[string]map[string]interface{}),
		UserAttributes:  make(map[string]interface{}),
		SessionLength:   LOOKER_SESSION_TIMEOUT,
		EmbedURL:        "/login/embed/" + url.QueryEscape(dashURL),
		ForceLogout:     true,
		Nonce:           fmt.Sprintf("\"%s\"", nonce),
		Time:            time.Now().Unix(),
	}

	urlOptions.UserAttributes["customer_code"] = customerCode

	return BuildLookerURL(urlOptions), nil
}

func (l *LookerClient) GenerateLookerTrialURL(requestID string) string {
	nonce, err := GenerateRandomString(16)
	if err != nil {
		return ""
	}

	options := LookerURLOptions{
		Host:            l.HostURL,
		Secret:          l.Secret,
		ExternalUserId:  fmt.Sprintf("\"%s\"", requestID),
		FirstName:       "",
		LastName:        "",
		GroupsIds:       make([]int, 0),
		ExternalGroupId: "",
		Permissions:     []string{"access_data", "see_looks", "see_user_dashboards", "download_without_limit", "clear_cache_refresh"}, // TODO: Verify these are final
		Models:          []string{"networknext_pbl"},                                                                                  // TODO: Verify these are final
		AccessFilters:   make(map[string]map[string]interface{}),
		UserAttributes:  make(map[string]interface{}),
		SessionLength:   3600,
		EmbedURL:        "/login/embed/" + url.QueryEscape("/embed/dashboards-next/?"), // TODO: Replace the ? with the correct dash ID
		ForceLogout:     true,
		Nonce:           fmt.Sprintf("\"%s\"", nonce),
		Time:            time.Now().Unix(),
	}

	return BuildLookerURL(options)
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		// not method, no service logger
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		// not method, no service logger
		return nil, err
	}

	return b, nil
}
