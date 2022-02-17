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
	v4 "github.com/looker-open-source/sdk-codegen/go/sdk/v4"
)

const (
	LOOKER_SESSION_TIMEOUT      = 86400
	EMBEDDED_USER_GROUP_ID      = 3
	USAGE_DASH_URL              = "/embed/dashboards-next/11"
	BASE_URL                    = "https://networknextexternal.cloud.looker.com"
	API_VERSION                 = "4.0"
	API_TIMEOUT                 = 300
	LOOKER_SAVES_ROW_LIMIT      = "10"
	LOOKER_AUTH_URI             = "%s/api/3.1/login?client_id=%s&client_secret=%s"
	LOOKER_QUERY_RUNNER_URI     = "%s/api/3.1/queries/run/json?force_production=true&cache=true"
	LOOKER_PROD_MODEL           = "network_next_prod"
	LOOKER_SAVES_VIEW           = "daily_big_saves"
	LOOKER_BILLING2_VIEW        = "billing2"
	LOOKER_SESSION_SUMMARY_VIEW = "billing2_session_summary"
	LOOKER_DATACENTER_INFO_VIEW = "datacenter_info_v3"
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

type LookerSessionMeta struct {
	Timestamp       string  `json:"billing2_session_summary.start_timestamp_time"`
	SessionID       int64   `json:"billing2_session_summary.session_id"`
	UserHash        int64   `json:"billing2_session_summary.user_hash"`
	Platform        string  `json:"billing2_session_summary.platform_type_label"`
	Connection      string  `json:"billing2_session_summary.connection_type_label"`
	ISP             string  `json:"billing2_session_summary.isp"`
	Longitude       float64 `json:"billing2_session_summary.longitude,omitempty"`
	Latitude        float64 `json:"billing2_session_summary.latitude,omitempty"`
	BuyerID         int64   `json:"billing2_session_summary.buyer_id,omitempty"`
	SDK             int64   `json:"billing2_session_summary.sdk_version,omitempty"`
	CustomerAddress float64 `json:"billing2_session_summary.customer_address,omitempty"`
	// ServerAddress   float64 `json:"billing2_session_summary.server_address"`
	DatacenterName  string `json:"datacenter_info_v3.datacenter_name"`
	DatacenterAlias string `json:"datacenter_info_v3.alias"`
}

type LookerSessionSlice struct {
	Timestamp                 string  `json:"REPLACE_ME.timestamp"`
	NextRTT                   float64 `json:"REPLACE_ME.next_rtt"`
	NextJitter                float64 `json:"REPLACE_ME.next_jitter"`
	NextPacketLoss            float64 `json:"REPLACE_ME.next_pl"`
	DirectRTT                 float64 `json:"REPLACE_ME.direct_rtt"`
	DirectJitter              float64 `json:"REPLACE_ME.direct_jitter"`
	DirectPacketLoss          float64 `json:"REPLACE_ME.direct_pl"`
	PredictedRTT              float64 `json:"REPLACE_ME.predicted_rtt"`
	PredictedJitter           float64 `json:"REPLACE_ME.predicted_jitter"`
	PredictedPacketLoss       float64 `json:"REPLACE_ME.predicted_pl"`
	ClientToServerRTT         float64 `json:"REPLACE_ME.client_to_server_rtt"`
	ClientToServerJitter      float64 `json:"REPLACE_ME.client_to_server_jitter"`
	ClientToServerPacketLoss  float64 `json:"REPLACE_ME.client_to_server_pl"`
	ServerToClientsRTT        float64 `json:"REPLACE_ME.server_to_client_rtt"`
	ServerToClientsJitter     float64 `json:"REPLACE_ME.server_to_client_jitter"`
	ServerToClientsPacketLoss float64 `json:"REPLACE_ME.server_to_client_pl"`
	RouteDiversity            uint32  `json:"REPLACE_ME.route_diversity"`
	EnvelopeUp                int64   `json:"REPLACE_ME.envelope_up"`
	EnvelopeDown              int64   `json:"REPLACE_ME.envelope_down"`
	OnNetworkNext             bool    `json:"REPLACE_ME.on_network_next"`
	IsMultiPath               bool    `json:"REPLACE_ME.is_multipath"`
	IsTryBeforeYouBuy         bool    `json:"REPLACE_ME.is_try_before_you_buy"`
}

type LookerSession struct {
	Meta   LookerSessionMeta
	Slices []LookerSessionSlice
}

func (l *LookerClient) RunSessionLookupQuery(sessionID string, timeFrame string) (LookerSession, error) {
	querySessionMeta := LookerSessionMeta{}
	querySessionSlices := make([]LookerSessionSlice, 0)

	uintID64, err := strconv.ParseUint(sessionID, 16, 64)
	if err != nil {
		return LookerSession{}, err
	}

	queryTimeFrame := timeFrame

	if queryTimeFrame == "" {
		queryTimeFrame = "7 days"
	}

	token, err := l.FetchAuthToken()
	if err != nil {
		return LookerSession{}, err
	}

	requiredFields := []string{
		LOOKER_SESSION_SUMMARY_VIEW + ".start_timestamp_time",
		LOOKER_SESSION_SUMMARY_VIEW + ".session_id",
		LOOKER_SESSION_SUMMARY_VIEW + ".buyer_id",
		LOOKER_SESSION_SUMMARY_VIEW + ".user_hash",
		LOOKER_SESSION_SUMMARY_VIEW + ".platform_type_label",
		LOOKER_SESSION_SUMMARY_VIEW + ".connection_type_label",
		LOOKER_SESSION_SUMMARY_VIEW + ".isp",
		LOOKER_SESSION_SUMMARY_VIEW + ".latitude",
		LOOKER_SESSION_SUMMARY_VIEW + ".longitude",
		LOOKER_SESSION_SUMMARY_VIEW + ".sdk_version",
		LOOKER_DATACENTER_INFO_VIEW + ".datacenter_name",
		LOOKER_DATACENTER_INFO_VIEW + ".customer_address",
		// LOOKER_DATACENTER_INFO_VIEW + ".server_address",
		LOOKER_DATACENTER_INFO_VIEW + ".alias",
	}
	sorts := []string{}
	requiredFilters := make(map[string]interface{})

	requiredFilters[LOOKER_SESSION_SUMMARY_VIEW+".session_id"] = fmt.Sprintf("%d", int64(uintID64))
	requiredFilters[LOOKER_SESSION_SUMMARY_VIEW+".start_timestamp_date"] = queryTimeFrame

	query := v4.WriteQuery{
		Model:   LOOKER_PROD_MODEL,
		View:    LOOKER_SESSION_SUMMARY_VIEW,
		Fields:  &requiredFields,
		Filters: &requiredFilters,
		Sorts:   &sorts,
	}

	lookerBody, _ := json.Marshal(query)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(LOOKER_QUERY_RUNNER_URI, l.APISettings.BaseUrl), bytes.NewBuffer(lookerBody))
	if err != nil {
		return LookerSession{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return LookerSession{}, err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return LookerSession{}, err
	}

	if err = json.Unmarshal(buf.Bytes(), &querySessionMeta); err != nil {
		return LookerSession{}, err
	}

	// TODO: Add slice call here if needed

	return LookerSession{
		Meta:   querySessionMeta,
		Slices: querySessionSlices,
	}, nil
}

func (l *LookerClient) RunUserSessionsLookupQuery(userID string, timeFrame string) ([]LookerSessionMeta, error) { // Timeframes 7, 10, 30, 60, 90
	querySessions := make([]LookerSessionMeta, 0)

	uintID64, err := strconv.ParseUint(userID, 16, 64)
	if err != nil {
		return querySessions, err
	}

	// TODO: If nothing comes back from looker for this ID, try hashing it and try again - similar to bigtable implementation

	queryTimeFrame := timeFrame

	if queryTimeFrame == "" {
		queryTimeFrame = "7 days"
	}

	token, err := l.FetchAuthToken()
	if err != nil {
		return querySessions, err
	}

	requiredFields := []string{
		LOOKER_SESSION_SUMMARY_VIEW + ".start_timestamp_time",
		LOOKER_SESSION_SUMMARY_VIEW + ".session_id",
		LOOKER_SESSION_SUMMARY_VIEW + ".user_hash",
		LOOKER_SESSION_SUMMARY_VIEW + ".platform_type_label",
		LOOKER_SESSION_SUMMARY_VIEW + ".connection_type_label",
		LOOKER_SESSION_SUMMARY_VIEW + ".isp",
		LOOKER_DATACENTER_INFO_VIEW + ".datacenter_name",
		// LOOKER_DATACENTER_INFO_VIEW + ".server_address",
		LOOKER_DATACENTER_INFO_VIEW + ".alias",
	}
	sorts := []string{}
	requiredFilters := make(map[string]interface{})

	requiredFilters[LOOKER_SESSION_SUMMARY_VIEW+".user_hash"] = fmt.Sprintf("%d", int64(uintID64))
	requiredFilters[LOOKER_SESSION_SUMMARY_VIEW+".start_timestamp_date"] = queryTimeFrame

	query := v4.WriteQuery{
		Model:   LOOKER_PROD_MODEL,
		View:    LOOKER_SESSION_SUMMARY_VIEW,
		Fields:  &requiredFields,
		Filters: &requiredFilters,
		Sorts:   &sorts,
	}

	lookerBody, _ := json.Marshal(query)
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

type AnalyticsDashboardCategory struct {
	ID      int64  `json:"id"`
	Label   string `json:"label"`
	Premium bool   `json:"premium"`
	Admin   bool   `json:"admin"`
	Seller  bool   `json:"seller"`
}

type AnalyticsDashboard struct {
	ID           int64                      `json:"id"`
	Name         string                     `json:"name"`
	Discovery    bool                       `json:"discovery"`
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
		Permissions:     []string{"access_data", "see_looks", "see_user_dashboards"}, // TODO: This may or may not need to change
		Models:          []string{"networknext_prod"},                                // TODO: This may or may not need to change
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
		Permissions:     []string{"access_data", "see_looks", "see_user_dashboards"}, // TODO: This may or may not need to change
		Models:          []string{"networknext_prod"},                                // TODO: This may or may not need to change
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
		Permissions:     []string{"access_data", "see_looks", "see_user_dashboards"}, // TODO: Verify these are final
		Models:          []string{"networknext_pbl"},                                 // TODO: Verify these are final
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
